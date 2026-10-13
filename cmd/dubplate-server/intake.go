package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/engine/tagio"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/intake/gate"
	"github.com/glw907/dubplate/internal/settings"
)

// stagingKey names the owned wall row holding the staging directory.
const stagingKey = "DUBPLATE_STAGING_DIR"

// stagingRow declares DUBPLATE_STAGING_DIR: the directory on the music
// volume where an upload stages until admission moves it. It is
// required and has no default, since a path a stranger's volume
// layout decides is no place for one.
var stagingRow = config.WallRow{
	Name:        stagingKey,
	Description: "the directory on the music volume where uploads stage before admission",
	Mark:        config.Sensitive,
}

// The retired names the staging row replaces. Rung 7's intake class
// reserved the first two; this rung stages in one directory and keeps
// no trigger file.
const (
	retiredUploadsKey = "DUBPLATE_UPLOADS_DIR"
	retiredTriggerKey = "DUBPLATE_TRIGGER_DIR"
	retiredIntakeNote = "superseded by " + stagingKey + ": uploads stage in one directory on the music volume"
)

// The intake limits. Each is store-homed (configuration-homes design,
// "the intake limits") and read at the next operation unless noted, so
// a raised cap reaches the next create without a restart. Each comment
// records the key's basis (build spec, "Defaults"). A setting always
// carries a default, so the group cap's "unset" is zero.
var (
	// A 79.8-minute 24-bit/192 kHz stereo album is about 5.5 GB raw
	// and about 3.3 GB at a typical 0.6 FLAC ratio, so the default
	// admits it where the salvage 2 GiB refused it. The floor is
	// 100 MiB, the ceiling 16 GiB.
	intakeMaxAlbumBytesSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_MAX_ALBUM_BYTES",
		Kind:        config.KindInt,
		Default:     "4294967296",
		Doc:         "an integer byte count from 104857600 to 17179869184, for example 4294967296",
		Apply:       config.ApplyNextOperation,
		Floor:       "104857600",
		Ceiling:     "17179869184",
		Description: "the largest total size of one album upload",
		Mark:        config.NotSensitive,
	}

	// A multi-disc box set plus scans. The ceiling is the salvage
	// ZIP entry limit, 2000.
	intakeMaxFilesSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_MAX_FILES",
		Kind:        config.KindInt,
		Default:     "500",
		Doc:         "an integer count from 1 to 2000, for example 500",
		Apply:       config.ApplyNextOperation,
		Floor:       "1",
		Ceiling:     "2000",
		Description: "the most files one upload may hold",
		Mark:        config.NotSensitive,
	}

	// Album, disc, and scan subfolders with margin.
	intakeMaxPathDepthSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_MAX_PATH_DEPTH",
		Kind:        config.KindInt,
		Default:     "8",
		Doc:         "an integer count from 1 to 32, for example 8",
		Apply:       config.ApplyNextOperation,
		Floor:       "1",
		Ceiling:     "32",
		Description: "the deepest folder nesting one upload may use",
		Mark:        config.NotSensitive,
	}

	// 1 MiB, below FLAC's 24-bit metadata block length (16777215
	// bytes), which is the ceiling.
	intakeMaxTagBytesSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_MAX_TAG_BYTES",
		Kind:        config.KindInt,
		Default:     "1048576",
		Doc:         "an integer byte count from 4096 to 16777215, for example 1048576",
		Apply:       config.ApplyNextOperation,
		Floor:       "4096",
		Ceiling:     "16777215",
		Description: "the largest tag block one file may carry",
		Mark:        config.NotSensitive,
	}

	// 8 MiB, below the same FLAC block length, so the cap bounds
	// something.
	intakeMaxPictureBytesSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_MAX_PICTURE_BYTES",
		Kind:        config.KindInt,
		Default:     "8388608",
		Doc:         "an integer byte count from 65536 to 16777215, for example 8388608",
		Apply:       config.ApplyNextOperation,
		Floor:       "65536",
		Ceiling:     "16777215",
		Description: "the largest embedded picture one file may carry",
		Mark:        config.NotSensitive,
	}

	// CD quality or better (the product copy's first section). The
	// floor is the default: no setting admits less than CD.
	intakeMinSampleRateSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_MIN_SAMPLE_RATE",
		Kind:        config.KindInt,
		Default:     "44100",
		Doc:         "an integer rate in hertz from 44100 to 192000, for example 44100",
		Apply:       config.ApplyNextOperation,
		Floor:       "44100",
		Ceiling:     "192000",
		Description: "the lowest sample rate an upload may have",
		Mark:        config.NotSensitive,
	}

	// CD quality or better. The ceiling is 24, the deepest sample the
	// FLAC decoder hashes.
	intakeMinBitsPerSampleSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_MIN_BITS_PER_SAMPLE",
		Kind:        config.KindInt,
		Default:     "16",
		Doc:         "an integer depth from 16 to 24, for example 16",
		Apply:       config.ApplyNextOperation,
		Floor:       "16",
		Ceiling:     "24",
		Description: "the lowest bit depth an upload may have",
		Mark:        config.NotSensitive,
	}

	// The salvage value.
	intakeSessionsPerMemberSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_SESSIONS_PER_MEMBER",
		Kind:        config.KindInt,
		Default:     "3",
		Doc:         "an integer count from 1 to 20, for example 3",
		Apply:       config.ApplyNextOperation,
		Floor:       "1",
		Ceiling:     "20",
		Description: "how many uploads one member may have open at once",
		Mark:        config.NotSensitive,
	}

	// The salvage floor, 5 GiB, which holds one maximum-size album's
	// tag-write copies. Sensitive on the reasoning the backup floors
	// carry: a stale session should not lower a protective floor
	// without re-authentication and a notice.
	intakeFreeSpaceFloorSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_FREE_SPACE_FLOOR",
		Kind:        config.KindInt,
		Default:     "5368709120",
		Doc:         "an integer byte count from 1073741824 to 1099511627776, for example 5368709120",
		Apply:       config.ApplyNextOperation,
		Floor:       "1073741824",
		Ceiling:     "1099511627776",
		Description: "the free space on the music volume uploads must leave untouched",
		Mark:        config.Sensitive,
	}

	// Zero means no group cap; the ceiling is 2^60.
	intakeGroupQuotaBytesSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_GROUP_QUOTA_BYTES",
		Kind:        config.KindInt,
		Default:     "0",
		Doc:         "an integer byte count from 0 to 1152921504606846976, where 0 means no group cap",
		Apply:       config.ApplyNextOperation,
		Floor:       "0",
		Ceiling:     "1152921504606846976",
		Description: "the most the group may have uploaded and not yet admitted",
		Mark:        config.NotSensitive,
	}

	// A weekend away from a slow residential upload, plus margin.
	intakeResumeWindowSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_RESUME_WINDOW",
		Kind:        config.KindDuration,
		Default:     "168h",
		Doc:         "a Go duration string from 1h to 720h, for example 168h",
		Apply:       config.ApplyNextOperation,
		Floor:       "1h",
		Ceiling:     "720h",
		Description: "how long an interrupted upload may wait to resume",
		Mark:        config.NotSensitive,
	}

	// Long enough to add a release to MusicBrainz and come back.
	intakeHoldSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_HOLD",
		Kind:        config.KindDuration,
		Default:     "168h",
		Doc:         "a Go duration string from 24h to 720h, for example 168h",
		Apply:       config.ApplyNextOperation,
		Floor:       "24h",
		Ceiling:     "720h",
		Description: "how long a held upload waits before it expires",
		Mark:        config.NotSensitive,
	}

	// Thirty minutes, to be confirmed against the gate's measured
	// decode throughput when the gate lands.
	intakeStallAfterSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_STALL_AFTER",
		Kind:        config.KindDuration,
		Default:     "30m",
		Doc:         "a Go duration string from 5m to 24h, for example 30m",
		Apply:       config.ApplyNextOperation,
		Floor:       "5m",
		Ceiling:     "24h",
		Description: "how long a check may run before the admins are told it has stalled",
		Mark:        config.NotSensitive,
	}

	// The interval is read once when the sweep starts, so a change
	// waits for a restart. The default leaves the stall rule's margin
	// a multiple of the interval.
	intakeSweepIntervalSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_SWEEP_INTERVAL",
		Kind:        config.KindDuration,
		Default:     "10m",
		Doc:         "a Go duration string from 1m to 1h, for example 10m",
		Apply:       config.ApplyRestart,
		Floor:       "1m",
		Ceiling:     "1h",
		Description: "how often the server sweeps uploads for stalls and expiries",
		Mark:        config.NotSensitive,
	}

	// One worker, so a decode never starves the requests that share
	// the process. The worker pool is sized once at start.
	intakeCheckWorkersSetting = config.Setting{
		Key:         "DUBPLATE_INTAKE_CHECK_WORKERS",
		Kind:        config.KindInt,
		Default:     "1",
		Doc:         "an integer count from 1 to 8, for example 1",
		Apply:       config.ApplyRestart,
		Floor:       "1",
		Ceiling:     "8",
		Description: "how many uploads the server checks at once",
		Mark:        config.NotSensitive,
	}
)

// intakeFormatAllowlistThreshold is DUBPLATE_INTAKE_FORMAT_ALLOWLIST:
// the formats an upload may carry, comma-separated, flac alone by
// default. The registry has no list kind and a string setting takes
// only a closed value set, so it registers as a KindString threshold
// on DUBPLATE_SERVE_BIND_ADDRS's precedent, and its consumer splits and
// validates it against the formats tagio supports.
var intakeFormatAllowlistThreshold = config.Threshold{
	Key:         "DUBPLATE_INTAKE_FORMAT_ALLOWLIST",
	Kind:        config.KindString,
	Default:     "flac",
	Doc:         "a comma-separated list of formats, for example flac",
	Description: "the audio formats an upload may carry, comma-separated",
	Mark:        config.NotSensitive,
}

// registerIntakeConfig adds the intake configuration to reg: the
// staging wall row, the two names it retires as superseded rows, the
// fifteen settings, and the format allowlist. newProductionRegistry
// calls it once. The intake and pipeline reserved classes lose their
// claim on those names in internal/config/keys.go in the same commit,
// since a name two families claim panics at registration.
func registerIntakeConfig(reg *config.Registry) {
	reg.AddWallRow(stagingRow)
	reg.AddSuperseded(retiredUploadsKey, retiredIntakeNote)
	reg.AddSuperseded(retiredTriggerKey, retiredIntakeNote)
	for _, s := range []config.Setting{
		intakeMaxAlbumBytesSetting,
		intakeMaxFilesSetting,
		intakeMaxPathDepthSetting,
		intakeMaxTagBytesSetting,
		intakeMaxPictureBytesSetting,
		intakeMinSampleRateSetting,
		intakeMinBitsPerSampleSetting,
		intakeSessionsPerMemberSetting,
		intakeFreeSpaceFloorSetting,
		intakeGroupQuotaBytesSetting,
		intakeResumeWindowSetting,
		intakeHoldSetting,
		intakeStallAfterSetting,
		intakeSweepIntervalSetting,
		intakeCheckWorkersSetting,
	} {
		reg.AddSetting(s)
	}
	reg.AddThreshold(intakeFormatAllowlistThreshold)
}

// intakeAllowlist reads DUBPLATE_INTAKE_FORMAT_ALLOWLIST from environ as
// its comma-separated format names, refusing an entry, an empty one
// among them, that names no row of the gate's format table or a format
// tagio does not read: serve would otherwise accept uploads it cannot
// check, or render bad-extension's wording with no format in it.
func intakeAllowlist(environ []string) ([]string, error) {
	value, err := config.ParseString(environ, intakeFormatAllowlistThreshold)
	if err != nil {
		return nil, err
	}
	names := strings.Split(value, ",")
	for i := range names {
		names[i] = strings.TrimSpace(names[i])
		row, ok := gate.LookupFormat(names[i])
		if !ok || !slices.Contains(tagio.SupportedFormats(), row.Format) {
			return nil, fmt.Errorf("%s: %q names no format this server can check", intakeFormatAllowlistThreshold.Key, names[i])
		}
	}
	return names, nil
}

// tusRouteName is the tus endpoint's subtree name. It cannot live under
// api, whose mount answers every path beneath it.
const tusRouteName = "tus"

// newIntake builds the server's intake over h for the staging directory
// cfg names: its limits and windows read through acc at each use, its
// sweep interval and worker count read once here, and its stall mail
// sent through send to the live admins roster lists.
func newIntake(ctx context.Context, h *store.Handle, cfg *config.Config, acc *settings.Accessor, allowlist []string,
	roster *auth.Roster, send func(ctx context.Context, to, subject, body string) error, logger *slog.Logger) (*intake.Intake, error) {
	interval, err := acc.Duration(ctx, intakeSweepIntervalSetting)
	if err != nil {
		return nil, settingReadError(intakeSweepIntervalSetting, err)
	}
	workers, err := acc.Int(ctx, intakeCheckWorkersSetting)
	if err != nil {
		return nil, settingReadError(intakeCheckWorkersSetting, err)
	}
	return intake.New(intake.Config{
		Store: h, Log: logger, Staging: cfg.WallValue(stagingKey), Now: time.Now,
		Limits:        intakeLimits(acc, allowlist),
		Windows:       intakeWindows(acc),
		SweepInterval: interval,
		CheckWorkers:  workers,
		Admins:        liveAdminAddresses(roster),
		Send:          send,
	})
}

// intakeWindows reads the sweep's windows in force through acc at each
// call. The first failed read fails the call naming its key.
func intakeWindows(acc *settings.Accessor) func(context.Context) (intake.Windows, error) {
	return func(ctx context.Context) (intake.Windows, error) {
		var failed error
		read := func(s config.Setting) time.Duration {
			d, err := acc.Duration(ctx, s)
			if err != nil && failed == nil {
				failed = settingReadError(s, err)
			}
			return d
		}
		w := intake.Windows{
			ResumeWindow: read(intakeResumeWindowSetting),
			Hold:         read(intakeHoldSetting),
			StallAfter:   read(intakeStallAfterSetting),
		}
		return w, failed
	}
}

// liveAdminAddresses reads the stall mail's recipients from roster at
// each call: every live member holding the admin flag, as the
// sensitive-change notice reads them, so a demoted or removed admin
// hears nothing.
func liveAdminAddresses(roster *auth.Roster) func(context.Context) ([]string, error) {
	return func(ctx context.Context) ([]string, error) {
		members, err := roster.List(ctx)
		if err != nil {
			return nil, err
		}
		var admins []string
		for _, m := range members {
			if m.Admin {
				admins = append(admins, m.Address)
			}
		}
		return admins, nil
	}
}

// tusVerifier admits a tus request whose session verifies as any
// member, through the same check the API's verifier makes. A request
// signed out wraps [intake.ErrUnauthenticated], which the mount answers
// 401; any other failure is logged here and answers 500.
func tusVerifier(cfg auth.AdminConfig) intake.Verifier {
	return func(w http.ResponseWriter, r *http.Request) (identity.UserID, error) {
		v, err := auth.CheckRequest(cfg, w, r)
		switch {
		case errors.Is(err, auth.ErrNotSignedIn):
			return identity.UserID{}, fmt.Errorf("%w: %w", intake.ErrUnauthenticated, err)
		case err != nil:
			cfg.Logger.Error("tus request could not verify its session", "error", err)
			return identity.UserID{}, err
		}
		return v.Identity.UserID, nil
	}
}

// freeSpaceFloor reads DUBPLATE_INTAKE_FREE_SPACE_FLOOR in force through
// acc at each call.
func freeSpaceFloor(acc *settings.Accessor) func(context.Context) (int64, error) {
	return func(ctx context.Context) (int64, error) {
		n, err := acc.Int(ctx, intakeFreeSpaceFloorSetting)
		return int64(n), settingReadError(intakeFreeSpaceFloorSetting, err)
	}
}

// resolveStaging refuses a staging directory and a library root that
// overlap, in either direction or as one path. The library tree is what
// backup syncs and restore refuses to fill, so a partial upload inside
// it would be backed up and skew the sync guard's count.
//
// Both paths must be absolute: filepath.Rel fails on a mixed pair, and
// a relative path is not one directory until the working directory
// says so. The comparison is lexical on cleaned paths and never resolves
// symlinks: EvalSymlinks fails on a path that does not exist, and a
// fresh box must still load its configuration so doctor can report the
// absence. The startup check compares the resolved paths once both
// exist. The error is keyed by the staging row, so doctor prints it
// there.
func resolveStaging(cfg *config.Config) error {
	staging := filepath.Clean(cfg.WallValue(stagingKey))
	library := filepath.Clean(cfg.LibraryRoot)
	if !filepath.IsAbs(staging) || !filepath.IsAbs(library) {
		return fmt.Errorf("%s: %s and DUBPLATE_LIBRARY_ROOT %s must both be absolute paths, so the overlap check can compare them",
			stagingKey, staging, library)
	}
	if contains(staging, library) || contains(library, staging) {
		return fmt.Errorf("%s: %s and DUBPLATE_LIBRARY_ROOT %s overlap; staging and the library root must be separate directories",
			stagingKey, staging, library)
	}
	return nil
}

// contains reports whether path is dir or lies inside it, lexically.
func contains(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && filepath.IsLocal(rel)
}
