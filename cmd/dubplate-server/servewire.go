package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/netip"
	"net/url"
	"os"
	"path"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/intake/staging"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/admin"
	"github.com/glw907/dubplate/internal/ops/opshttp"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// newServeShell builds the shell the serve verb runs, owning prefix, and
// returns it as a [server] the caller closes when the shell stops
// serving. Every refusal returns no server, and none opens a listener.
func newServeShell(ctx context.Context, environ []string, reg *config.Registry, prefix serve.Prefix, bindAddrs []netip.Addr, cfg *config.Config, logger *slog.Logger) (_ *server, err error) {
	proxyConfig, err := serve.ParseTrustedProxy(cfg)
	if err != nil {
		return nil, err
	}
	allowlist, err := intake.Allowlist(environ)
	if err != nil {
		return nil, err
	}
	// Intake's startup check refuses before the store opens.
	if err := staging.Check(staging.Read(cfg.WallValue(intake.StagingKey)), staging.Read(cfg.LibraryRoot)); err != nil {
		return nil, fmt.Errorf("startup check: %w", err)
	}
	h, err := openServeStore(ctx, environ, reg, cfg, logger)
	if err != nil {
		return nil, err
	}
	var in *intake.Intake
	defer func() {
		if err == nil {
			return
		}
		if in != nil {
			err = errors.Join(err, in.Close())
		}
		err = errors.Join(err, h.Close())
	}()
	// The seed applies before anything reads a setting.
	if err := applySettingsSeed(ctx, h, reg, environ, logger); err != nil {
		return nil, err
	}
	acc := settings.NewAccessor(h, reg)
	authConfig, err := auth.ConfigFrom(cfg, acc)
	if err != nil {
		return nil, err
	}
	// The rate limits are read once and change at the next start; every
	// other auth setting is read from the store at each use.
	limits, err := auth.ReadLimits(ctx, acc)
	if err != nil {
		return nil, err
	}
	shell := serve.NewShell(prefix, proxyConfig, logger)
	// No mount takes a bearer credential, so every one stays behind the
	// origin check. A route that bypasses it needs a shell-level test
	// that it never reads the session cookie.
	routes := mountRoutes(shell)
	if sendsHSTS(authConfig.PublicOrigin) {
		shell.EnableHSTS()
	}

	authDir, err := auth.Open(cfg.StateDir)
	if err != nil {
		return nil, err
	}
	nonces, _, err := auth.OpenNonceStore(ctx, authDir, authConfig.NonceRetention, auth.MagicLinkExpiryCeiling, time.Now())
	if err != nil {
		return nil, err
	}
	epoch := auth.NewEpochStore(authDir, shell.Logger())

	roster := member.NewRoster(h, epoch, time.Now, shell.Logger())

	limiter := auth.NewLimiter(limits)

	mailer, err := notify.New(cfg)
	if err != nil {
		return nil, err
	}

	verifier := auth.NewMagicLinkVerifier(roster, epoch, authConfig.SessionSecret, authConfig.PreviousSessionSecret,
		authConfig.SessionLifetime, authConfig.SessionAbsoluteCap)

	// The shell's strip reads the same proxy class, so the peer whose
	// forwarded headers survive is the one whose header names the client.
	clientIP := func(r *http.Request) string { return serve.ClientIP(r, proxyConfig) }
	handlers := auth.NewHandlers(auth.HandlersConfig{
		Roster:             roster,
		Nonces:             nonces,
		Epoch:              epoch,
		Limiter:            limiter,
		Send:               mailer.Send,
		ClientIP:           clientIP,
		Now:                time.Now,
		Logger:             shell.Logger(),
		LinkSecret:         authConfig.LinkSecret,
		PreviousLinkSecret: authConfig.PreviousLinkSecret,
		SessionSecret:      authConfig.SessionSecret,
		LinkExpiry:         authConfig.LinkExpiry,
		SessionLifetime:    authConfig.SessionLifetime,
		SessionAbsoluteCap: authConfig.SessionAbsoluteCap,
		PublicOrigin:       authConfig.PublicOrigin,
		Prefix:             prefix.String(),
		BodyCap:            authConfig.BodyCap,
		BodyCapCeiling:     auth.RequestBodyCapCeiling,
		GenericAnswer:      authConfig.GenericAnswer,
	})

	registerAuthRoutes(routes, handlers)

	adminConfig := auth.VerifyConfig{
		Verifier: verifier,
		ClientIP: clientIP,
		Now:      time.Now,
		Logger:   shell.Logger(),
	}
	// The probe mounts like every route, so the request it counts is one
	// the strip has already cleared.
	routes.Register(serve.ProbeRouteName, serve.NewProbeRoute(prefix, bindAddrs, proxyConfig))

	adminService := admin.New(admin.Config{
		Store: h, Registry: reg, Environ: environ, Roster: roster,
		Settings: acc,
		Send:     mailer.Send, Now: time.Now, Logger: shell.Logger(),
	})
	in, err = newIntake(ctx, h, cfg, acc, allowlist, roster, mailer.Send, shell.Logger())
	if err != nil {
		return nil, err
	}
	apiMount, err := newAPIMount(prefix, adminService, in.Service(), adminConfig)
	if err != nil {
		return nil, err
	}
	routes.Register(apiRouteName, apiMount)
	// tusd logs through golang.org/x/exp/slog, which the shell's logger
	// cannot fill, so its warnings go to stderr as JSON.
	tusMount, err := in.TusMount(intake.TusConfig{
		Prefix:    prefix.String(),
		PublicURL: authConfig.PublicOrigin,
		Verify:    tusVerifier(adminConfig),
		Floor:     intake.FreeSpaceFloor(acc),
		TusLog:    os.Stderr,
	})
	if err != nil {
		return nil, err
	}

	webRoutes, err := newWebRoutes(environ, prefix, cfg.SiteName, cfg.StateDir, shell.Logger(), webBuild)
	if err != nil {
		return nil, err
	}
	registerWebRoutes(routes, webRoutes)
	registerAdminRoutes(routes, webRoutes, adminConfig)

	// Intake starts last, once nothing else can refuse, because it holds
	// its lease until Close.
	if err := in.Start(ctx); err != nil {
		return nil, err
	}
	routes.Register(tusRouteName, tusMount)

	go sweepNonces(ctx, nonces, shell.Logger())
	return &server{shell: shell, store: h, intake: in}, nil
}

// A server is what newServeShell builds: the shell, the store it holds
// open, and intake, whose work runs against that store.
type server struct {
	shell  *serve.Shell
	store  *store.Handle
	intake *intake.Intake
}

// Close tears the build down in the one safe order: intake's context is
// cancelled and its workers and sweep waited for, then its lease is
// released, and only then is the store closed, so no verdict or cleanup
// runs against a closing store.
func (s *server) Close() error {
	return errors.Join(s.intake.Close(), s.store.Close())
}

// openServeStore opens the dubplate store for serve: it reads the
// store's thresholds, decides the keep from the facts [store.Inspect]
// reads, and calls [store.OpenServe] with the backup flag and the
// reconciler a restore awaiting serve needs. A refusal at any step
// returns no handle.
func openServeStore(ctx context.Context, environ []string, reg *config.Registry, cfg *config.Config, logger *slog.Logger) (*store.Handle, error) {
	busyTimeout, err := store.BusyTimeout(environ)
	if err != nil {
		return nil, err
	}
	migrationWait, err := store.MigrationWait(environ)
	if err != nil {
		return nil, err
	}
	var stored settings.Facts
	report := store.Inspect(ctx, store.InspectOptions{
		StateDir:    cfg.StateDir,
		BusyTimeout: busyTimeout,
		Readers:     []store.Reader{stored.Reader()},
	})
	keep, err := settings.LocalSnapshotKeep(report, stored, environ, reg)
	if err != nil {
		return nil, err
	}
	h, err := store.OpenServe(ctx, cfg.StateDir, store.ServeOptions{
		BusyTimeout:   busyTimeout,
		MigrationWait: migrationWait,
		Keep:          keep,
		Backup:        cfg.IntegrationEnabled(backup.Class.Name),
		Logger:        logger,
		Reconcile:     backup.Reconciler(cfg.LibraryRoot),
	})
	if err != nil {
		return nil, fmt.Errorf("open store: %w", err)
	}
	return h, nil
}

// applySettingsSeed applies the environment seed rule to the store h
// holds, so every setting row serve reads from here on already agrees
// with the environment. A refusal returns the joined refusal and
// changes nothing; the caller then starts no listener.
func applySettingsSeed(ctx context.Context, h *store.Handle, reg *config.Registry, environ []string, logger *slog.Logger) error {
	ctx, cancel := context.WithTimeout(ctx, storeWait)
	defer cancel()
	if err := settings.Apply(ctx, h, reg, environ, time.Now(), logger); err != nil {
		return fmt.Errorf("apply settings seed: %w", err)
	}
	return nil
}

// sendsHSTS reports whether the shell sends Strict-Transport-Security
// for publicURL, the URL browsers reach the server at: only when it is
// https on a host [auth.IsLoopbackHost] does not match. The request's own
// scheme is no input, since the server speaks plain HTTP to whatever
// fronts it. A loopback host is excluded because RFC 6797 section 8.3
// matches a known HSTS host by name alone, so one response from
// https://localhost would force https on every localhost port in that
// browser for a year.
func sendsHSTS(publicURL string) bool {
	u, err := url.Parse(publicURL)
	if err != nil {
		return false
	}
	return u.Scheme == "https" && !auth.IsLoopbackHost(u.Hostname())
}

// sweepNonces runs the nonce sweep every [auth.SweepInterval] and
// returns when ctx is canceled.
func sweepNonces(ctx context.Context, nonces *auth.NonceStore, logger *slog.Logger) {
	ticker := time.NewTicker(auth.SweepInterval)
	defer ticker.Stop()
	nonces.RunSweep(ctx, ticker.C, time.Now, logger)
}

// newAPIMount builds the /api/v1 handler over the admin and submission
// operations for the sessions cfg verifies.
func newAPIMount(prefix serve.Prefix, adminOps ops.Admin, submissionOps ops.Submissions, cfg auth.VerifyConfig) (http.Handler, error) {
	return opshttp.NewMount(adminOps, submissionOps, prefix, apiVerifier(cfg), cfg.Logger)
}

// apiVerifier admits a request whose session verifies as any member,
// building the actor from the session's anchor, never its issue time: a
// refresh moves the issue time, and staying signed in must never renew
// the freshness the re-authentication gate reads. It logs one record
// per admitted request, naming the username under the "principal" key
// because the actor carries the durable id.
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
			"op", path.Base(r.URL.Path), "actor", actor.String(), "principal", v.Identity.Username,
			"channel", v.Identity.Channel.String(), "client_ip", clientIP)
		return actor, privilege, nil
	}
}

// newIntake builds the server's intake over h for the staging directory
// cfg names: its limits and windows read through acc at each use, its
// sweep interval and worker count read once here, and its stall mail
// sent through send to the live admins roster lists.
func newIntake(ctx context.Context, h *store.Handle, cfg *config.Config, acc *settings.Accessor, allowlist []string,
	roster *member.Roster, send func(ctx context.Context, to, subject, body string) error, logger *slog.Logger) (*intake.Intake, error) {
	interval, err := intake.SweepInterval(ctx, acc)
	if err != nil {
		return nil, err
	}
	workers, err := intake.CheckWorkerCount(ctx, acc)
	if err != nil {
		return nil, err
	}
	return intake.New(intake.Config{
		Store: h, Log: logger, Staging: cfg.WallValue(intake.StagingKey), Now: time.Now,
		EarliestAdmin: member.EarliestAdmin,
		Limits:        intake.LimitsFrom(acc, allowlist),
		Windows:       intake.WindowsFrom(acc),
		SweepInterval: interval,
		CheckWorkers:  workers,
		Admins:        liveAdminAddresses(roster),
		Send:          send,
	})
}

// liveAdminAddresses reads the stall mail's recipients from roster at
// each call: every live member holding the admin flag, as the
// sensitive-change notice reads them, so a demoted or removed admin
// hears nothing.
func liveAdminAddresses(roster *member.Roster) func(context.Context) ([]string, error) {
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
// signed out wraps [ops.CodeUnauthenticated], which the mount answers
// 401; any other failure is logged here and answers 500.
func tusVerifier(cfg auth.VerifyConfig) intake.Verifier {
	return func(w http.ResponseWriter, r *http.Request) (identity.UserID, error) {
		v, err := auth.VerifyRequest(cfg, w, r)
		switch {
		case errors.Is(err, ops.CodeUnauthenticated):
			return identity.UserID{}, err
		case err != nil:
			cfg.Logger.Error("tus request could not verify its session", "error", err)
			return identity.UserID{}, err
		}
		return v.Identity.UserID, nil
	}
}
