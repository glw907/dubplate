package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/intake/staging"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/ops/admin"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
	"github.com/glw907/dubplate/internal/webapp"
	"github.com/spf13/cobra"
)

// The three subtree names the auth handlers mount under the owned
// prefix.
const (
	requestLinkRouteName = "request-link"
	consumeLinkRouteName = "consume-link"
	signOutRouteName     = "sign-out"
)

// membersRouteName is the add-member document's subtree name, served
// to verified admins only.
const membersRouteName = "members"

// serverSettingsRouteName is the admin settings document's subtree name,
// served to verified admins only. It is not "settings", which the visual
// identity spec reserves for the member preferences page.
const serverSettingsRouteName = "server-settings"

// The web subtree names [newWebRoutes] builds routers for: the login
// documents, the embedded bundle's hashed assets, and the two tier
// stylesheets that follow the bundle's own.
const (
	loginRouteName       = "login"
	assetsRouteName      = "_app"
	brandingRouteName    = "branding.css"
	operatorCSSRouteName = "operator.css"
)

// newServeCmd builds the serve verb: it loads the wall through
// [loadConfig], builds the shell with newServeShell, and runs it until
// cmd's context is done, holding the store open until then. Every
// resolution and store open happens before [serve.Run] opens a
// listener, so an unsatisfied wall refuses at startup in the words
// doctor reports.
func newServeCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "serve",
		Short:         "Run the HTTP server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			cfg, err := loadConfig(environ, reg)
			if err != nil {
				return err
			}
			prefix, addrs, port, err := serve.ParseConfig(environ)
			if err != nil {
				return err
			}
			srv, err := newServeShell(cmd.Context(), environ, reg, prefix, addrs, cfg, slog.Default())
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, srv.Close()) }()
			return serve.Run(cmd.Context(), srv.shell.Logger(), srv.shell.Handler(), addrs, port)
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd
}

// newServeShell builds the shell the serve verb runs, owning prefix, and
// returns it as a [server] the caller closes when the shell stops
// serving. Every refusal returns no server, and none opens a listener.
//
// The order matters in three places. The intake startup check runs
// before the store opens, and the settings seed applies before anything
// else is built. Intake starts last, once nothing else can refuse: its
// lease is held until Close, and a lease another process holds refuses
// the start naming the holder.
//
// The four rate limits the seed left in force are read once, here, and
// apply at the next start. Every other auth setting is a reader over
// the store, called at each use.
//
// No mount authenticates by bearer credential, so none registers a
// bearer bypass and every one stays behind the origin check. A route
// that bypasses it needs a shell-level test that it never reads the
// session cookie.
//
// The members document mounts behind [auth.RequireAdmin], outside the
// method and path checks, so a non-admin meets the gate's 404 whatever
// the method or path and only a verified admin meets a 405.
//
// The trusted proxy class governs both the strip's forwarded-address
// carve-out and every handler's ClientIP, so the peer whose forwarded
// headers survive the strip is the peer whose header names the client.
// Every identity header is stripped from every peer, a trusted proxy
// included. The boundary probe mounts like every other route, so the
// request it counts is one the strip has already cleared.
func newServeShell(ctx context.Context, environ []string, reg *config.Registry, prefix serve.Prefix, bindAddrs []string, cfg *config.Config, logger *slog.Logger) (*server, error) {
	buildFS, err := embeddedBuild()
	if err != nil {
		return nil, err
	}
	return newServeShellVia(ctx, environ, reg, prefix, bindAddrs, cfg, logger, func(s *serve.Shell) registrar { return s }, buildFS)
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

// A registrar mounts a route under a subtree name. *serve.Shell is
// one; the route inventory's test wraps it in a second that records
// each name before forwarding.
type registrar interface {
	Register(name string, h http.Handler)
}

// newServeShellVia is [newServeShell] mounting every route through
// the registrar mount makes of the new shell, so a test sees each name
// the build registers. Every registration must go through that
// registrar and never through the shell directly: a route registered
// around it escapes the route inventory test unseen. The web routes
// come from buildFS, so a test can serve a fixture tree where the
// embedded one holds only the placeholder.
func newServeShellVia(ctx context.Context, environ []string, reg *config.Registry, prefix serve.Prefix, bindAddrs []string, cfg *config.Config, logger *slog.Logger, mount func(*serve.Shell) registrar, buildFS fs.FS) (_ *server, err error) {
	proxyConfig, err := serve.ParseTrustedProxy(cfg)
	if err != nil {
		return nil, err
	}
	allowlist, err := intake.Allowlist(environ)
	if err != nil {
		return nil, err
	}
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
	if err := applySettingsSeed(ctx, h, reg, environ, logger); err != nil {
		return nil, err
	}
	acc := settings.NewAccessor(h, reg)
	authConfig, err := auth.ConfigFrom(cfg, acc)
	if err != nil {
		return nil, err
	}
	limits, err := auth.ReadLimits(ctx, acc)
	if err != nil {
		return nil, err
	}
	strip := serve.NewStrip(proxyConfig)
	shell := serve.NewShell(prefix, strip, logger)
	routes := mount(shell)
	if sendsHSTS(authConfig.PublicOrigin) {
		shell.EnableHSTS()
	}

	authDir, err := auth.Open(cfg.StateDir)
	if err != nil {
		return nil, err
	}
	nonces, err := openNonceStore(ctx, authDir, authConfig.NonceRetention, time.Now())
	if err != nil {
		return nil, err
	}
	epoch := auth.OpenEpochStore(authDir)

	roster := member.NewRoster(h, epoch, time.Now)

	limiter := auth.NewLimiter(limits)

	mailConfig, err := notify.ConfigFrom(cfg)
	if err != nil {
		return nil, err
	}
	mailer := notify.New(mailConfig)

	verifier := auth.NewMagicLinkVerifier(roster, epoch, authConfig.SessionSecret, authConfig.PreviousSessionSecret,
		authConfig.SessionLifetime, authConfig.SessionAbsoluteCap)

	clientIP := func(r *http.Request) string { return serve.ClientIP(r, proxyConfig) }
	handlersConfig := authHandlersConfig(prefix, authConfig, roster, nonces, epoch, limiter, mailer, clientIP, shell.Logger())
	handlers := auth.NewHandlers(handlersConfig)

	// The auth handlers apply no method check of their own; mounting
	// each behind methodOnly is what fixes the one method it accepts.
	routes.Register(requestLinkRouteName, methodOnly(http.MethodPost, handlers.RequestLink))
	routes.Register(consumeLinkRouteName, methodOnly(http.MethodPost, handlers.ConsumeLink))
	routes.Register(signOutRouteName, methodOnly(http.MethodPost, handlers.SignOut))

	adminConfig := auth.VerifyConfig{
		Verifier: verifier,
		ClientIP: clientIP,
		Now:      time.Now,
		Logger:   shell.Logger(),
	}
	// The shell mounts a name and everything under it to one handler, so
	// the probe answers the shell's 404 to any deeper path itself.
	probePath := prefix.String() + "/" + serve.ProbeRouteName
	routes.Register(serve.ProbeRouteName, webapp.ExactPaths(map[string]http.Handler{probePath: serve.NewProbeRoute(bindAddrs, proxyConfig)}))

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
	// tusd's own warnings go to stderr as JSON, where serve's default
	// logger writes: its logger is golang.org/x/exp/slog's, a type the
	// shell's log/slog logger cannot fill.
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

	webRoutes, err := newWebRoutesFromFS(environ, prefix, cfg.SiteName, privateFiles{stateDir: cfg.StateDir}, shell.Logger(), buildFS)
	if err != nil {
		return nil, err
	}
	registerWebRoutes(routes, webRoutes)
	routes.Register(membersRouteName, auth.RequireAdmin(adminConfig, webRoutes.Members))
	routes.Register(serverSettingsRouteName, auth.RequireAdmin(adminConfig, webRoutes.ServerSettings))

	if err := in.Start(ctx); err != nil {
		return nil, err
	}
	routes.Register(tusRouteName, tusMount)

	go sweepNonces(ctx, nonces, shell.Logger())
	return &server{shell: shell, store: h, intake: in}, nil
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

// openNonceStore opens the consumed-nonce store under authDir, with
// its retention read through retention, and sweeps it once at now. A
// test opens the store the way newServeShell does and reads the
// retention it was wired with.
func openNonceStore(ctx context.Context, authDir *auth.Dir, retention settings.Reader[time.Duration], now time.Time) (*auth.NonceStore, error) {
	nonces, _, err := auth.OpenNonceStore(ctx, authDir, retention, auth.MagicLinkExpiryCeiling, now)
	return nonces, err
}

// authHandlersConfig builds the [auth.HandlersConfig] the auth handlers
// mount from, its secrets, origin, and readers taken from ac. A test
// reads the config it returns, the only proof that no caller loses a
// wired value.
func authHandlersConfig(prefix serve.Prefix, ac auth.Config, roster *member.Roster, nonces *auth.NonceStore, epoch *auth.EpochStore, limiter *auth.Limiter, mailer *notify.Mailer, clientIP func(r *http.Request) string, logger *slog.Logger) auth.HandlersConfig {
	return auth.HandlersConfig{
		Roster:             roster,
		Nonces:             nonces,
		Epoch:              epoch,
		Limiter:            limiter,
		Send:               mailer.Send,
		ClientIP:           clientIP,
		Now:                time.Now,
		Logger:             logger,
		LinkSecret:         ac.LinkSecret,
		PreviousLinkSecret: ac.PreviousLinkSecret,
		SessionSecret:      ac.SessionSecret,
		LinkExpiry:         ac.LinkExpiry,
		SessionLifetime:    ac.SessionLifetime,
		SessionAbsoluteCap: ac.SessionAbsoluteCap,
		PublicOrigin:       ac.PublicOrigin,
		Prefix:             prefix.String(),
		BodyCap:            ac.BodyCap,
		BodyCapCeiling:     auth.RequestBodyCapCeiling,
		GenericAnswer:      ac.GenericAnswer,
	}
}

// methodOnly wraps h, answering 405 with an Allow header naming method
// to any request whose method is not method.
func methodOnly(method string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			w.Header().Set("Allow", method)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// sweepNonces runs the nonce sweep every [auth.SweepInterval] and
// returns when ctx is canceled.
func sweepNonces(ctx context.Context, nonces *auth.NonceStore, logger *slog.Logger) {
	ticker := time.NewTicker(auth.SweepInterval)
	defer ticker.Stop()
	auth.RunSweep(ctx, nonces.Sweep, time.Now, ticker.C, logger)
}
