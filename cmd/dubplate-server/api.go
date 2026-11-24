package main

import (
	"net/http"
	"path"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/opshttp"
	"github.com/glw907/dubplate/internal/serve"
)

// apiRouteName is the cookie mount's subtree name. The shell hands the
// mount the full path, so it cuts <prefix>/api/v1/ itself and answers
// every other path under api with a JSON 404.
const apiRouteName = "api"

// newAPIMount builds the /api/v1 handler over the admin and submission
// operations for the sessions cfg verifies.
func newAPIMount(prefix serve.Prefix, adminOps ops.Admin, submissionOps ops.Submissions, cfg auth.VerifyConfig) (http.Handler, error) {
	return opshttp.NewMount(adminOps, submissionOps, prefix.String(), apiVerifier(cfg), cfg.Logger)
}

// apiVerifier admits a request whose session verifies as any member,
// building the actor from the session's anchor, never its issue time: a
// refresh moves the issue time, and staying signed in must never renew
// the freshness the re-authentication gate reads. It logs one record
// per admitted request, naming the principal under its own key because
// the actor carries the durable id.
func apiVerifier(cfg auth.VerifyConfig) opshttp.Verifier {
	return func(w http.ResponseWriter, r *http.Request) (identity.Actor, ops.Privilege, error) {
		v, err := auth.VerifyRequest(cfg, w, r)
		if err != nil {
			return identity.Actor{}, 0, err
		}
		clientIP := cfg.ClientIP(r)
		actor := identity.NetworkActor(v.Identity, v.Session.AnchoredAt, clientIP)
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
