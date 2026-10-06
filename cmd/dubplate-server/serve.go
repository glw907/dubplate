package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

// The magic-link identity source's key names this file reads through
// [config.Config.IdentitySourceValue]; auth.go declares them as the
// source's Required and Optional rows. None of these constants carries
// a secret value; each names the environment variable a real deployment
// carries one under.
const (
	rosterFileKey              = "DUBPLATE_ROSTER_FILE"
	magicLinkSecretKey         = "DUBPLATE_MAGIC_LINK_SECRET"          // secret-guard-allow
	magicLinkSecretPreviousKey = "DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS" // secret-guard-allow
	sessionSecretKey           = "DUBPLATE_SESSION_SECRET"             // secret-guard-allow
	sessionSecretPreviousKey   = "DUBPLATE_SESSION_SECRET_PREVIOUS"    // secret-guard-allow
	publicURLKey               = "DUBPLATE_PUBLIC_URL"
	smtpHostKey                = "DUBPLATE_SMTP_HOST"
	smtpPortKey                = "DUBPLATE_SMTP_PORT"
	smtpUserKey                = "DUBPLATE_SMTP_USER"
	smtpPasswordKey            = "DUBPLATE_SMTP_PASSWORD" // secret-guard-allow
	smtpFromKey                = "DUBPLATE_SMTP_FROM"
)

// The three subtree names the auth handlers mount under the owned
// prefix.
const (
	requestLinkRouteName = "request-link"
	consumeLinkRouteName = "consume-link"
	signOutRouteName     = "sign-out"
)

// rosterRouteName is the roster API's subtree name: the admin gate and
// its POST/GET/HEAD surface.
const rosterRouteName = "roster"

// membersRouteName is the add-member document's subtree name, served
// to verified admins only.
const membersRouteName = "members"

// The web subtree names [newWebRoutes] builds routers for: the login
// documents, the embedded bundle's hashed assets, and the two tier
// stylesheets that follow the bundle's own.
const (
	loginRouteName       = "login"
	assetsRouteName      = "_app"
	brandingRouteName    = "branding.css"
	operatorCSSRouteName = "operator.css"
)

// serveBindAddrsThreshold, servePortThreshold, and servePrefixThreshold
// are the serve shell's own three keys, declared in this file rather
// than internal/config: the alternative, declaring inside
// internal/serve, would make the shell import internal/config, which
// it has no reason to do.
//
// The three keys replace internal/config/keys.go's own "serve" reserved
// class, deleted in the same commit as this declaration (the removal
// half of the rung 1b registry contract). That class's own comment
// carried one argument forward: a bind address is not an identity
// concern, which is why the wall named it its own class apart from
// "auth and identity" rather than folding it in.
//
// ruled, conductor 2026-09-08, R-A, R-B, R-C

// serveBindAddrsThreshold is the comma-separated list of addresses the
// serve verb binds. Its default, loopback only, is the salvage
// configuration's own default. The registry has no list kind (R-B), so
// this registers as KindString and [serve.SplitBindAddrs] splits and
// validates it in the package that binds, keeping the address grammar
// with the code that uses it.
var serveBindAddrsThreshold = config.Threshold{
	Key:     "DUBPLATE_SERVE_BIND_ADDRS",
	Kind:    config.KindString,
	Default: "127.0.0.1",
	Doc:     "a comma-separated list of bind addresses, for example 127.0.0.1,::1",
}

// servePortThreshold is the TCP port the serve verb binds on every
// configured address. Its default, 8080, is the salvage configuration's
// own default (_salvage/internal/config/config_test.go:44-45).
var servePortThreshold = config.Threshold{
	Key:     "DUBPLATE_SERVE_PORT",
	Kind:    config.KindInt,
	Default: "8080",
	Doc:     "an integer TCP port from 1024 through 65535",
}

// servePortFloor and servePortCeiling bound DUBPLATE_SERVE_PORT. The
// floor is Linux's default net.ipv4.ip_unprivileged_port_start: the
// serve unit's empty CapabilityBoundingSet= withholds
// CAP_NET_BIND_SERVICE, so a lower port fails at bind, and port 0 would
// listen on whatever ephemeral port the kernel picks.
const (
	servePortFloor   = 1024
	servePortCeiling = 65535
)

// validateServePort checks DUBPLATE_SERVE_PORT against
// [servePortFloor] and [servePortCeiling], naming the capability a
// privileged port would need. [validateMagicLinkSource] calls it, so
// [config.Load] refuses an out-of-range port for doctor and every verb
// before serve ever binds.
func validateServePort(environ []string) error {
	port, err := config.ParseInt(environ, servePortThreshold)
	if err != nil {
		return &sourceKeyError{key: servePortThreshold.Key, reason: splitKeyReason(err)}
	}
	if port < servePortFloor {
		return &sourceKeyError{key: servePortThreshold.Key, reason: fmt.Sprintf(
			"%d is below %d: the serve unit withholds CAP_NET_BIND_SERVICE, so it cannot bind a privileged port", port, servePortFloor)}
	}
	if port > servePortCeiling {
		return &sourceKeyError{key: servePortThreshold.Key, reason: fmt.Sprintf("%d is above %d", port, servePortCeiling)}
	}
	return nil
}

// errDirNotWritable reports that this process may not list, create,
// and remove entries in a directory a later write needs.
var errDirNotWritable = errors.New("not writable")

// checkDirWritable asks the kernel whether this process's effective
// uid may read, write, and search dir: faccessat(2) with AT_EACCESS,
// whose man page says it "Perform[s] access checks using the effective
// user and group IDs" and fails with EROFS when "Write permission was
// requested for a file on a read-only filesystem", the shape a
// directory outside the unit's ReadWritePaths takes under
// ProtectSystem=strict. It writes nothing, so it can neither race the
// roster's directory flock nor leave a probe file behind. Read and
// search are asked for beside write because every writer here opens or
// lists the directory too.
func checkDirWritable(dir string) error {
	if err := unix.Faccessat(unix.AT_FDCWD, dir, unix.R_OK|unix.W_OK|unix.X_OK, unix.AT_EACCESS); err != nil {
		return fmt.Errorf("%s is %w: %w", dir, errDirNotWritable, err)
	}
	return nil
}

// checkRosterDirWritable refuses when the roster file's directory is
// not writable: every roster write replaces the file by rename under a
// lock on that directory, so a directory the process cannot write
// fails the first member add rather than the start.
func checkRosterDirWritable(rosterPath string) error {
	if err := checkDirWritable(filepath.Dir(rosterPath)); err != nil {
		return fmt.Errorf("roster directory %w", err)
	}
	return nil
}

// servePrefixThreshold is the one owned path prefix: every route the
// shell registers lives under it, and every path outside it answers
// 404. Its default, /dubplate, is one segment so the segment-aware
// match has a boundary to find, and carries no estate value.
var servePrefixThreshold = config.Threshold{
	Key:     "DUBPLATE_SERVE_PREFIX",
	Kind:    config.KindString,
	Default: "/dubplate",
	Doc:     "a single path segment beginning with a separator, for example /dubplate",
}

// registerServeConfig adds the serve shell's three thresholds to reg,
// a plain function: newRootCmd builds a fresh, independent registry on
// every construction, so registering onto it can never collide with a
// prior process's own registration.
func registerServeConfig(reg *config.Registry) {
	reg.AddThreshold(serveBindAddrsThreshold)
	reg.AddThreshold(servePortThreshold)
	reg.AddThreshold(servePrefixThreshold)
}

// resolveServeConfig reads the three thresholds against environ and
// returns the values [serve.Run] and its construction values take: the
// validated owned prefix, the split and validated bind addresses, and
// the port. It returns before any listener opens, so a caller can
// refuse an invalid prefix or bind list without ever touching the
// network.
func resolveServeConfig(environ []string) (serve.Prefix, []string, int, error) {
	prefixValue, err := config.ParseString(environ, servePrefixThreshold)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}
	prefix, err := serve.NewPrefix(prefixValue)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}

	bindRaw, err := config.ParseString(environ, serveBindAddrsThreshold)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}
	addrs, err := serve.SplitBindAddrs(bindRaw)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}

	port, err := config.ParseInt(environ, servePortThreshold)
	if err != nil {
		return serve.Prefix{}, nil, 0, err
	}

	return prefix, addrs, port, nil
}

// newServeCmd builds the serve verb: it validates the owned
// configuration wall through [config.Load] against reg, resolves the
// three serve thresholds and rung 5's ten auth thresholds, builds the
// shell and mounts the auth handlers through newServeShell, and runs it
// until cmd's context is done. Every resolution and every store or
// roster open in newServeShell happens before serve.Run opens a
// listener, so a server whose wall or whose auth dependencies are
// unsatisfied refuses at startup, in the same words doctor reports,
// rather than accepting connections and failing at the first request
// that needs a missing value.
func newServeCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "serve",
		Short:         "Run the HTTP server",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(environ, reg)
			if err != nil {
				return err
			}
			prefix, addrs, port, err := resolveServeConfig(environ)
			if err != nil {
				return err
			}
			authCfg, err := resolveAuthConfig(environ)
			if err != nil {
				return err
			}
			shell, err := newServeShell(cmd.Context(), environ, prefix, addrs, cfg, authCfg, slog.Default())
			if err != nil {
				return err
			}
			return serve.Run(cmd.Context(), shell.Logger(), shell.Handler(), addrs, port)
		},
	}
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd
}

// newServeShell builds the shell the serve verb runs, owning prefix,
// opens the auth store, the nonce and epoch stores, and the roster,
// builds the rate limiter, the mailer, and the magic-link identity
// seam, constructs the auth handlers, and mounts them under the owned
// prefix, alongside the web routes [newWebRoutes] builds from environ.
// It returns before any listener opens, so a caller refuses a trusted
// proxy value [resolveTrustedProxy] refuses, a missing roster file, a
// roster directory it cannot write, an unopenable store, an unparseable mail port, a bad theme or branding
// value, or a refused operator CSS file without ever accepting a
// connection.
//
// It registers no bearer bypass: no mount authenticates by bearer
// credential, so every mount stays behind the origin check's
// cookie-carrying protection. The first bypassed route must add a
// shell-level test that it never reads the session cookie.
//
// Every record the shell and its handlers write goes to logger.
//
// The shell sends Strict-Transport-Security only when the public URL
// is https on a host that is not loopback; see [sendsHSTS].
//
// The members document is mounted behind [auth.RequireAdmin], outside
// its own method and path checks, so a non-admin meets the gate's 404
// whatever the method or the path, and only a verified admin ever
// meets a 405.
//
// It starts the nonce sweep bound to ctx; see [sweepNonces].
//
// The trusted proxy class, resolved through [resolveTrustedProxy],
// governs both the strip's forwarded-address carve-out and every
// handler's ClientIP, so the peer whose forwarded headers survive the
// strip is exactly the peer whose header names the client. With the
// class off, PeerAddress is the client address. Every identity header
// is stripped from every peer, the trusted proxy included, since a
// proxy is trusted for the client address alone.
//
// The boundary probe route mounts under the owned prefix like every
// other route, so the request it counts is the one the strip has
// already cleared. Its peer rule reads bindAddrs, the resolved list
// the caller listens on, and the same trusted proxy configuration the
// strip carries.
func newServeShell(ctx context.Context, environ []string, prefix serve.Prefix, bindAddrs []string, cfg *config.Config, authCfg authConfig, logger *slog.Logger) (*serve.Shell, error) {
	proxyConfig, err := resolveTrustedProxy(cfg)
	if err != nil {
		return nil, err
	}
	strip := serve.NewStrip(proxyConfig)
	shell := serve.NewShell(prefix, strip, logger)
	if sendsHSTS(cfg.IdentitySourceValue(publicURLKey)) {
		shell.EnableHSTS()
	}

	store, err := auth.Open(cfg.StateDir)
	if err != nil {
		return nil, err
	}
	nonces, _, err := auth.OpenNonceStore(store, authCfg.nonceRetention, magicLinkExpiryCeiling, time.Now())
	if err != nil {
		return nil, err
	}
	epoch := auth.OpenEpochStore(store)

	roster, err := auth.OpenRoster(cfg.IdentitySourceValue(rosterFileKey))
	if err != nil {
		return nil, err
	}
	if err := checkRosterDirWritable(cfg.IdentitySourceValue(rosterFileKey)); err != nil {
		return nil, err
	}

	limiter := auth.NewLimiter(auth.Limits{
		LinkRequestsPerAddress:       authCfg.linkRequestsPerAddress,
		LinkRequestsPerClient:        authCfg.linkRequestsPerClient,
		ConsumptionAttemptsPerClient: authCfg.consumptionAttemptsPerClient,
		LinkMailCeiling:              authCfg.linkMailCeiling,
	})

	mailer, err := newMailer(cfg)
	if err != nil {
		return nil, err
	}

	seam := auth.NewMagicLinkSeam(roster, epoch,
		[]byte(cfg.IdentitySourceValue(sessionSecretKey)),
		[]byte(cfg.IdentitySourceValue(sessionSecretPreviousKey)),
		authCfg.sessionLifetime, authCfg.sessionAbsoluteCap)

	clientIP := func(r *http.Request) string { return serve.ClientIP(r, proxyConfig) }
	handlersConfig := authHandlersConfig(prefix, cfg, authCfg, roster, nonces, epoch, limiter, mailer, clientIP, shell.Logger())
	handlers := auth.NewHandlers(handlersConfig)

	// The auth handlers apply no method check of their own; mounting
	// each behind methodOnly is what fixes the one method it accepts.
	shell.Register(requestLinkRouteName, methodOnly(http.MethodPost, handlers.RequestLink))
	shell.Register(consumeLinkRouteName, methodOnly(http.MethodPost, handlers.ConsumeLink))
	shell.Register(signOutRouteName, methodOnly(http.MethodPost, handlers.SignOut))

	adminConfig := auth.AdminConfig{
		Seam:     seam,
		Roster:   roster,
		ClientIP: clientIP,
		Now:      time.Now,
		Logger:   shell.Logger(),
		BodyCap:  int64(authCfg.authRequestBodyCap),
	}
	probePath := prefix.String() + "/" + serve.ProbeRouteName
	shell.Register(serve.ProbeRouteName, exactPath(probePath, serve.NewProbeRoute(bindAddrs, proxyConfig)))

	shell.Register(rosterRouteName, exactPath(prefix.String()+"/"+rosterRouteName, auth.NewRosterAPI(adminConfig)))

	webRoutes, err := newWebRoutes(environ, prefix, cfg.SiteName, privateFiles{
		roster:   cfg.IdentitySourceValue(rosterFileKey),
		stateDir: cfg.StateDir,
	}, shell.Logger())
	if err != nil {
		return nil, err
	}
	webRoutes.register(shell)
	shell.Register(membersRouteName, auth.RequireAdmin(adminConfig, serve.LogRequests(shell.Logger(), webRoutes.members)))

	go sweepNonces(ctx, nonces, shell.Logger())
	return shell, nil
}

// sendsHSTS reports whether the shell sends Strict-Transport-Security
// for publicURL, the URL browsers reach the server at: only when it is
// https on a host [isLoopbackHost] does not match. The request's own
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
	return u.Scheme == "https" && !isLoopbackHost(u.Hostname())
}

// exactPath wraps next, answering [http.NotFoundHandler]'s own 404 to any
// request whose r.URL.Path is not exactly want. [serve.Shell.Register]
// mounts both the subtree name and everything under name+"/" to the
// same handler (internal/serve/serve.go:94-96), and the shell
// dispatches through its mux's own Handler method, which leaves
// r.Pattern unset, so a mounted handler cannot otherwise tell its own
// exact path from a deeper one. The roster API needs the distinction:
// a request to <prefix>/roster/roster must answer the shell's 404,
// never the member list, and the shell's 404 this wrap answers is
// byte-equal to the admin gate's own refusal, so a non-admin sees the
// same answer on either side of it. It is [newExactPathRoutes] (web.go)
// narrowed to the one-route case.
func exactPath(want string, next http.Handler) http.Handler {
	return newExactPathRoutes(map[string]http.Handler{want: next})
}

// authHandlersConfig builds the [auth.HandlersConfig] newServeShell
// mounts the auth handlers from. It exists as its own function so a
// test can read the config it built rather than recomputing the same
// values by hand: PublicOrigin, Prefix, and SessionAbsoluteCap are each
// wiring rather than logic, so the only proof a caller loses each one
// correctly is a test that reads this function's own return value.
func authHandlersConfig(prefix serve.Prefix, cfg *config.Config, authCfg authConfig, roster *auth.Roster, nonces *auth.NonceStore, epoch *auth.EpochStore, limiter *auth.Limiter, mailer *notify.Mailer, clientIP func(r *http.Request) string, logger *slog.Logger) auth.HandlersConfig {
	return auth.HandlersConfig{
		Roster:             roster,
		Nonces:             nonces,
		Epoch:              epoch,
		Limiter:            limiter,
		Send:               mailer.Send,
		ClientIP:           clientIP,
		Now:                time.Now,
		Logger:             logger,
		LinkSecret:         []byte(cfg.IdentitySourceValue(magicLinkSecretKey)),
		PreviousLinkSecret: []byte(cfg.IdentitySourceValue(magicLinkSecretPreviousKey)),
		SessionSecret:      []byte(cfg.IdentitySourceValue(sessionSecretKey)),
		LinkExpiry:         authCfg.magicLinkExpiry,
		SessionLifetime:    authCfg.sessionLifetime,
		SessionAbsoluteCap: authCfg.sessionAbsoluteCap,
		PublicOrigin:       cfg.IdentitySourceValue(publicURLKey),
		Prefix:             prefix.String(),
		BodyCap:            int64(authCfg.authRequestBodyCap),
		GenericAnswer:      authCfg.unknownAddressAnswer == "generic",
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

// newMailer resolves the magic-link source's five SMTP keys into a
// [notify.Mailer]. The port is the one value among them this file
// converts rather than passes through. [validateMagicLinkSource]'s
// Validate hook already bounds the port to 1 through 65535 at load
// time, so doctor and every verb see a bad port before the serve verb
// ever listens; this parse is the defensive second read that turns the
// bounded string back into the int [notify.Config] takes.
func newMailer(cfg *config.Config) (*notify.Mailer, error) {
	port, err := strconv.Atoi(cfg.IdentitySourceValue(smtpPortKey))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", smtpPortKey, err)
	}
	return notify.New(notify.Config{
		Host:     cfg.IdentitySourceValue(smtpHostKey),
		Port:     port,
		User:     cfg.IdentitySourceValue(smtpUserKey),
		Password: cfg.IdentitySourceValue(smtpPasswordKey),
		From:     cfg.IdentitySourceValue(smtpFromKey),
	}), nil
}

// sweepNonces runs the nonce sweep every [auth.SweepInterval] and
// returns when ctx is canceled.
func sweepNonces(ctx context.Context, nonces *auth.NonceStore, logger *slog.Logger) {
	ticker := time.NewTicker(auth.SweepInterval)
	defer ticker.Stop()
	auth.RunSweep(ctx, nonces.Sweep, time.Now, ticker.C, logger)
}
