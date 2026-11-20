package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/opshttp"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
)

// apiRouteName is the cookie mount's subtree name. The shell hands the
// mount the full path, so it cuts <prefix>/api/v1/ itself and answers
// every other path under api with a JSON 404.
const apiRouteName = "api"

// newAPIMount builds the /api/v1 handler over the admin and submission
// seams for the sessions cfg verifies.
func newAPIMount(prefix serve.Prefix, adminSeam ops.AdminSeam, submissionSeam ops.SubmissionSeam, cfg auth.AdminConfig) (http.Handler, error) {
	return opshttp.NewMount(adminSeam, submissionSeam, prefix.String(), apiVerifier(cfg), cfg.Logger)
}

// intakeLimits reads the intake settings in force through acc at each
// call, so a raised cap reaches the next create. allowlist is the
// environment's format bound, which no setting changes. The first
// failed read fails the call naming its key, and no limit falls back to
// a default.
func intakeLimits(acc *settings.Accessor, allowlist []string) func(context.Context) (intake.Limits, error) {
	return func(ctx context.Context) (intake.Limits, error) {
		var failed error
		count := func(s config.Setting) int {
			n, err := acc.Int(ctx, s)
			if err != nil && failed == nil {
				failed = auth.SettingReadError(s, err)
			}
			return n
		}
		lim := intake.Limits{
			Policy: gate.Policy{
				MaxAlbumBytes:    int64(count(intakeMaxAlbumBytesSetting)),
				MaxFiles:         count(intakeMaxFilesSetting),
				MaxPathDepth:     count(intakeMaxPathDepthSetting),
				MaxTagBytes:      int64(count(intakeMaxTagBytesSetting)),
				MaxPictureBytes:  int64(count(intakeMaxPictureBytesSetting)),
				MinSampleRate:    count(intakeMinSampleRateSetting),
				MinBitsPerSample: count(intakeMinBitsPerSampleSetting),
				Allowlist:        allowlist,
			},
			SessionsPerMember: count(intakeSessionsPerMemberSetting),
			GroupQuotaBytes:   int64(count(intakeGroupQuotaBytesSetting)),
			FreeSpaceFloor:    int64(count(intakeFreeSpaceFloorSetting)),
		}
		return lim, failed
	}
}

// apiVerifier admits a request whose session verifies as any member,
// building the actor from the session's anchor, never its issue time: a
// refresh moves the issue time, and staying signed in must never renew
// the freshness the re-authentication gate reads. It logs one record
// per admitted request, naming the principal under its own key because
// the actor carries the durable id.
func apiVerifier(cfg auth.AdminConfig) opshttp.Verifier {
	return func(w http.ResponseWriter, r *http.Request) (ops.Actor, ops.Privilege, error) {
		v, err := auth.CheckRequest(cfg, w, r)
		switch {
		case errors.Is(err, auth.ErrNotSignedIn):
			return ops.Actor{}, 0, fmt.Errorf("%w: %w", opshttp.ErrUnauthenticated, err)
		case err != nil:
			return ops.Actor{}, 0, err
		}
		clientIP := cfg.ClientIP(r)
		actor := ops.NetworkActor(v.Identity, v.Session.AnchoredAt, clientIP)
		privilege := ops.PrivilegeMember
		if v.Admin {
			privilege = ops.PrivilegeAdmin
		}
		cfg.Logger.Info("api request admitted",
			"op", path.Base(r.URL.Path), "actor", actor.String(), "principal", v.Identity.Principal,
			"channel", v.Identity.Channel.String(), "client_ip", clientIP)
		return actor, privilege, nil
	}
}
