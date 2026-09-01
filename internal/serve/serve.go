// serve.go is dubplate serve's composition root: it builds the Engine
// (upload.go), the JourneyAPI (api.go), and the config/stream-probe
// endpoints this file owns, wires them behind one Cloudflare Access
// enforcement layer, and runs the whole thing over exactly
// Serve.BindAddrs until the caller's context is done.
package serve

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"

	"github.com/glw907/dubplate/internal/access"
	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/journey"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/notify"
)

// --- server assembly ---------------------------------------------------

// Server is dubplate serve's full HTTP surface, built once and run until
// its caller's context is done. The zero Server is not usable; construct
// with [NewServer].
type Server struct {
	cfg     *config.Config
	handler http.Handler
	engine  *Engine
}

// NewServerDeps holds [NewServer]'s inputs beyond cfg and roster: the
// embedded static assets ([github.com/glw907/dubplate/web].Assets, still
// rooted at "dist" -- NewServer strips that prefix itself, the one place
// able to return the fs.Sub error instead of panicking at program init),
// and the loopback-only dev-identity override cmd/dubplate/serve.go
// parses from --dev-identity.
type NewServerDeps struct {
	Assets      fs.FS
	DevIdentity string
	// Now stands in for time.Now; the zero value uses the real clock.
	Now func() time.Time
}

// nudgeTriggerSubdir is the directory under [config.Paths.TriggerDir]
// serve writes nudge trigger files into; U8's dubplate-import-nudge.path
// watches it with DirectoryNotEmpty= and U8's tmpfiles fragment creates
// it, both keyed to this same name.
const nudgeTriggerSubdir = "nudge-triggers"

// basePath is the path prefix serve owns end to end: every API route,
// the SSE probe, the tus endpoints, and the embedded web app all live
// under it (U12a amendment: the uploader has no hostname of its own --
// it lives at https://music.907.life/upload, and cloudflared's
// path-scoped ingress rule forwards the request with this prefix
// intact, since cloudflared does not strip path prefixes). U9 must set
// the web build's Vite `base` to "/upload/" so every emitted asset href
// resolves under this same prefix; this file owns only the server-side
// mount.
const basePath = "/upload"

// mountBasePath wraps next so every request must fall under [basePath]:
// a bare "/upload" request (no trailing slash) redirects to "/upload/"
// first, so relative asset hrefs and history.pushState reads all get
// consistent directory semantics; [http.StripPrefix] then removes the
// prefix for anything under "/upload/", so next (and everything it
// wraps -- [enforceAccess], [Server.buildMux]) never has to know
// basePath exists. The explicit boundary check below is required
// because [http.StripPrefix] does a plain string trim, not a
// path-segment-aware one: unguarded, a request to "/uploadfoo" would
// trim to "foo" and pass straight through to next rather than 404 --
// the same segment-boundary gap the sibling ingress-rule regex
// (config/cloudflared-tunnel-config.json) is separately anchored
// against.
func mountBasePath(next http.Handler) http.Handler {
	stripped := http.StripPrefix(basePath, next)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == basePath {
			http.Redirect(w, r, basePath+"/", http.StatusMovedPermanently)
			return
		}
		if !strings.HasPrefix(r.URL.Path, basePath+"/") {
			http.NotFound(w, r)
			return
		}
		stripped.ServeHTTP(w, r)
	})
}

// NewServer builds a [Server] over cfg and roster: the upload engine, the
// journey API, the Access verifier, and every route this file itself
// owns, all behind [enforceAccess]. It refuses to start -- returning an
// error rather than a partially built Server -- when --dev-identity is
// set against a non-loopback bind address (N/A in production, real in a
// misconfigured dev run) or when UploadsDir and InboxDir do not share a
// filesystem (the atomic-promotion precondition promote's doc comment
// depends on).
func NewServer(cfg *config.Config, roster config.Roster, deps NewServerDeps) (*Server, error) {
	if err := validateDevIdentity(deps.DevIdentity, cfg.Serve.BindAddrs); err != nil {
		return nil, err
	}
	if err := sameFilesystem(cfg.Paths.UploadsDir, cfg.Paths.InboxDir); err != nil {
		return nil, fmt.Errorf("serve startup: uploads and inbox must share a filesystem: %w", err)
	}
	if deps.Now == nil {
		deps.Now = time.Now
	}

	assets, err := fs.Sub(deps.Assets, "dist")
	if err != nil {
		return nil, fmt.Errorf("serve: open embedded dist: %w", err)
	}

	l, err := ledger.Open(cfg.LedgerPath)
	if err != nil {
		return nil, fmt.Errorf("serve: open ledger: %w", err)
	}

	triggerDir := filepath.Join(cfg.Paths.TriggerDir, nudgeTriggerSubdir)
	engine, err := NewEngine(cfg, roster, Options{
		Now:         deps.Now,
		OnPromoted:  func(sessionID, _ string) { writeNudgeTrigger(triggerDir, sessionID) },
		TusBasePath: basePath + filesBasePath,
	})
	if err != nil {
		return nil, fmt.Errorf("serve: build upload engine: %w", err)
	}

	japi := NewJourneyAPI(engine, JourneyDeps{
		Ledger: l,
		Navidrome: &journey.NavidromeScanClient{
			URL:      cfg.Navidrome.URL,
			User:     cfg.Navidrome.AdminUser,
			Password: cfg.Navidrome.AdminPassword,
		},
		Mailer:             notify.New(cfg.SMTP),
		LockPath:           lock.Path(cfg.Paths.StateDir),
		AdminEmail:         cfg.AdminEmail,
		StallMinutes:       cfg.Thresholds.StallMinutes,
		PublicNavidromeURL: cfg.Serve.PublicNavidromeURL,
		Now:                deps.Now,
	})

	verifier := access.NewVerifier(cfg.Serve.AccessTeamDomain, cfg.Serve.AccessAUD)
	s := &Server{
		cfg:    cfg,
		engine: engine,
	}
	s.handler = mountBasePath(enforceAccess(s.buildMux(engine, japi, assets), accessConfig{
		verify:      verifier.Verify,
		roster:      roster,
		devIdentity: deps.DevIdentity,
		adminEmail:  cfg.AdminEmail,
		warmPage:    warmRosterAbsentPage(cfg.AdminEmail),
	}))
	return s, nil
}

// buildMux assembles s's full route table: the tus and session API
// (engine's own mux, mounted at the same absolute paths it registers
// internally), the journey polling API (japi's own mux, same reasoning),
// this file's own /api/config and /api/stream/probe, and the embedded
// static app at "/" -- one index.html, one asset dir, no SPA-fallback
// routing (the spec's development-shape ruling).
func (s *Server) buildMux(engine *Engine, japi *JourneyAPI, assets fs.FS) http.Handler {
	eh := engine.Handler()
	jh := japi.Handler()

	mux := http.NewServeMux()
	mux.Handle("POST /api/session", eh)
	mux.Handle("POST /api/session/{id}/finish", eh)
	mux.Handle("/files/", eh)
	mux.Handle("GET /api/journey/{id}", jh)
	mux.Handle("GET /api/history", jh)
	mux.HandleFunc("GET /api/config", s.handleConfig)
	mux.HandleFunc("POST /api/stream/probe", s.handleStreamProbe)
	mux.Handle("/", staticHandler(assets))
	return mux
}

// staticHandler serves assets (the embedded dist/ tree) with the plan's
// two Cache-Control rules: a hashed /assets/* entry is immutable forever
// (embed.FS carries no mtimes to conditionally revalidate against, and a
// hash-named file never changes content under a stable name); every
// other path -- in practice only index.html -- is no-cache, so a
// redeployed app is never served stale from a shared cache. Every
// response here rides the authenticated transport enforceAccess already
// required, so "private" is correct rather than "public".
func staticHandler(assets fs.FS) http.Handler {
	fileServer := http.FileServerFS(assets)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/assets/") {
			w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
		} else {
			w.Header().Set("Cache-Control", "no-cache")
		}
		fileServer.ServeHTTP(w, r)
	})
}

// --- access enforcement --------------------------------------------

// accessAssertionHeader is the header Cloudflare Access stamps onto every
// request it forwards from behind the reverse-proxy layer.
const accessAssertionHeader = "Cf-Access-Jwt-Assertion"

// accessConfig bundles [enforceAccess]'s dependencies: the verification
// call (a func field, not an interface, so a test can inject a fake
// without a second real implementation ever existing), the roster
// membership is checked against, and the loopback-gated dev-identity
// bypass.
type accessConfig struct {
	verify      func(ctx context.Context, token string) (string, error)
	roster      config.Roster
	devIdentity string
	adminEmail  string
	warmPage    []byte
}

// accessFailure is the JSON body every /api/* and /files/* enforcement
// failure renders, so a client never has to parse prose to find out what
// happened.
type accessFailure struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// unauthenticatedFailure is the 401 body for a missing or invalid Access
// assertion on the API/files surface: the client's cue to reload and sign
// in again, never a bare error page.
var unauthenticatedFailure = accessFailure{
	Error:   "unauthenticated",
	Message: "Your sign-in has expired. Reload the page to sign in again.",
}

// rosterAbsentFailure is the 403 body for a verified identity absent from
// the roster, carrying the same "ask the admin" copy [warmRosterAbsentPage]
// renders for a document request.
func rosterAbsentFailure(adminEmail string) accessFailure {
	return accessFailure{
		Error:   "not_on_roster",
		Message: "This email address isn't on the family list yet. Ask " + adminEmail + " to add you, then try again.",
	}
}

// warmRosterAbsentPage renders the "ask the admin to add you" document
// response: the one non-JSON enforcement failure, since only a browser
// navigating the page directly (never an API client) ever reaches it.
func warmRosterAbsentPage(adminEmail string) []byte {
	return fmt.Appendf(nil, `<!doctype html>
<html lang="en">
<head><meta charset="utf-8"><title>dubplate</title></head>
<body>
<main>
<h1>You're not on the list yet</h1>
<p>This email address isn't on the family list yet. Ask the admin
(<a href="mailto:%s">%s</a>) to add you, then reload this page.</p>
</main>
</body>
</html>
`, adminEmail, adminEmail)
}

// enforceAccess wraps next with dubplate serve's whole identity layer:
// every request must resolve to a verified email (real Access
// verification, or the loopback-gated dev-identity override) and that
// email must be on the roster. The failure shape is per surface (the
// plan's per-surface ruling): a document request (anything outside
// /api/* and /files/*) from a roster-absent email gets the warm 200 HTML
// page; /api/* and /files/* get machine-readable JSON -- 401 for a
// missing or invalid assertion, 403 with the same "ask the admin" copy
// for roster-absent. A verified, on-roster request carries its email
// into ctx via [WithEmail], exactly as [withTestAuth] does in this
// package's own tests.
func enforceAccess(next http.Handler, ac accessConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		api := isAPIOrFilesPath(r.URL.Path)

		email, ok := verifiedIdentity(r, ac)
		if !ok {
			if api {
				writeJSON(w, http.StatusUnauthorized, unauthenticatedFailure)
				return
			}
			http.Error(w, unauthenticatedFailure.Message, http.StatusUnauthorized)
			return
		}

		if _, onRoster := ac.roster.DirNameFor(email); !onRoster {
			if api {
				writeJSON(w, http.StatusForbidden, rosterAbsentFailure(ac.adminEmail))
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(ac.warmPage)
			return
		}

		next.ServeHTTP(w, r.WithContext(WithEmail(r.Context(), email)))
	})
}

// isAPIOrFilesPath reports whether path is under the machine-readable
// surface (/api/* or /files/*) rather than a document request.
func isAPIOrFilesPath(path string) bool {
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/files/")
}

// verifiedIdentity resolves r's caller email: ac.devIdentity when set and
// r arrived over loopback ([devIdentityLoopbackOnly] -- defense in depth
// beyond [validateDevIdentity]'s startup refusal), otherwise a real,
// verified Access assertion. It reports false for a missing header or a
// verification failure alike; neither tells a prober which.
func verifiedIdentity(r *http.Request, ac accessConfig) (string, bool) {
	if ac.devIdentity != "" && devIdentityLoopbackOnly(r) {
		return ac.devIdentity, true
	}
	token := r.Header.Get(accessAssertionHeader)
	if token == "" {
		return "", false
	}
	email, err := ac.verify(r.Context(), token)
	if err != nil {
		return "", false
	}
	return email, true
}

// devIdentityLoopbackOnly reports whether r arrived over a loopback
// connection -- the per-request check that makes the --dev-identity
// bypass inert on a non-loopback listener even if [validateDevIdentity]'s
// startup refusal were ever bypassed.
func devIdentityLoopbackOnly(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

// validateDevIdentity refuses a non-empty devIdentity unless every entry
// in bindAddrs parses as a loopback address: the startup half of the
// dev-identity bypass's "loopback-only" contract, paired with
// [devIdentityLoopbackOnly]'s per-request check.
func validateDevIdentity(devIdentity string, bindAddrs []string) error {
	if devIdentity == "" {
		return nil
	}
	for _, addr := range bindAddrs {
		ip := net.ParseIP(addr)
		if ip == nil || !ip.IsLoopback() {
			return fmt.Errorf("serve: --dev-identity requires every bind address to be loopback, got %q", addr)
		}
	}
	return nil
}

// --- GET /api/config -----------------------------------------------

// configResponse is GET /api/config's body: everything the page needs to
// render its selection screen and validate a choice client-side before
// ever uploading a byte, none of it hand-duplicated from what the gate
// and admission control actually enforce.
type configResponse struct {
	SiteName           string   `json:"site_name"`
	Collections        []string `json:"collections"`
	ChunkMB            int      `json:"chunk_mb"`
	MaxAlbumBytes      int64    `json:"max_album_bytes"`
	MaxSessionsPerUser int      `json:"max_sessions_per_user"`
	ReviewClockDays    int      `json:"review_clock_days"`
	FormatSpec         string   `json:"format_spec"`
	// PrivateCollection names which entry of Collections is the private
	// lane, so the client can tell it apart from a same-named ordinary
	// collection instead of assuming the documented default
	// (passb-copy-final.md's site-values-through-config rule).
	PrivateCollection string `json:"private_collection"`
}

// handleConfig serves GET /api/config.
func (s *Server) handleConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, configResponse{
		SiteName:           s.cfg.Serve.SiteName,
		Collections:        s.cfg.Collections,
		ChunkMB:            s.cfg.Upload.ChunkMB,
		MaxAlbumBytes:      s.cfg.Upload.MaxAlbumBytes,
		MaxSessionsPerUser: s.cfg.Upload.MaxSessionsPerUser,
		ReviewClockDays:    s.cfg.Thresholds.RetentionDays,
		FormatSpec:         formatSpecText(cdMinSampleRate, cdMinBitsPerSample, s.cfg.Upload.MaxAlbumBytes),
		PrivateCollection:  s.engine.PrivateCollection(),
	})
}

// formatSpecText renders the page's one accepted-format sentence from
// the actual enforced constants ([cdMinSampleRate], [cdMinBitsPerSample],
// and the caller's maxAlbumBytes), never hand-written prose that could
// drift from what the gate enforces.
func formatSpecText(minSampleRate, minBitsPerSample int, maxAlbumBytes int64) string {
	return fmt.Sprintf("FLAC only, %d Hz / %d-bit or better, one album per upload, up to %s total.",
		minSampleRate, minBitsPerSample, humanGiB(maxAlbumBytes))
}

// humanGiB renders n bytes as a one-decimal GiB figure ("2.0 GiB").
func humanGiB(n int64) string {
	return strconv.FormatFloat(float64(n)/(1<<30), 'f', 1, 64) + " GiB"
}

// --- POST /api/stream/probe -----------------------------------------

// probeHeartbeatInterval is gate 1's mitigation-set heartbeat cadence
// (the plan's Gates section): well under cloudflared's documented ~100s
// idle cut, and under the spec's 20s ceiling.
const probeHeartbeatInterval = 15 * time.Second

// handleStreamProbe serves POST /api/stream/probe (gate 1's dedicated
// measurement target, deployed by U12a): a deliberately minimal
// text/event-stream endpoint carrying no journey semantics, existing
// solely to prove -- through the real tunnel and Access -- that
// cloudflared delivers bytes incrementally as they are flushed rather
// than buffering until the connection closes, and that it forwards
// Last-Event-ID on reconnect. It is POST-opened per the gate's mitigation
// set (a GET-opened stream is the buffering cloudflared#1449 hits).
// Every tick is both the required heartbeat and a numbered event a
// client can watch arrive in real time; there is no real backlog to
// replay, so honoring Last-Event-ID means resuming the count, not
// resending history.
func (s *Server) handleStreamProbe(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	seq := 0
	if last := r.Header.Get("Last-Event-ID"); last != "" {
		if n, err := strconv.Atoi(last); err == nil {
			seq = n + 1
		}
	}

	// A long-lived stream must not be cut by the server's ordinary
	// Read/WriteTimeout (set for the API's short JSON routes); tusd's own
	// clearBodyDeadlines makes the same call for /files/.
	rc := http.NewResponseController(w)
	_ = rc.SetWriteDeadline(time.Time{})

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	writeEvent := func(id int) bool {
		_, err := fmt.Fprintf(w, "id: %d\ndata: probe\n\n", id)
		if err != nil {
			return false
		}
		flusher.Flush()
		return true
	}
	if !writeEvent(seq) {
		return
	}

	ticker := time.NewTicker(probeHeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			seq++
			if !writeEvent(seq) {
				return
			}
		}
	}
}

// --- startup preconditions -------------------------------------------

// sameFilesystem checks that a and b resolve to the same device: the
// precondition atomic promotion depends on ([Engine.promote]'s doc
// comment), since a rename across filesystems silently falls back to a
// non-atomic copy-and-delete. Its own copy of
// internal/checks/uploader.go's probeSameFilesystem, not a shared
// import -- each package that depends on this precondition probes it
// itself, so serve can refuse to start with no dependency on internal/checks.
func sameFilesystem(a, b string) error {
	var statA, statB unix.Stat_t
	if err := unix.Stat(a, &statA); err != nil {
		return fmt.Errorf("stat %s: %w", a, err)
	}
	if err := unix.Stat(b, &statB); err != nil {
		return fmt.Errorf("stat %s: %w", b, err)
	}
	if !sameDevice(statA, statB) {
		return fmt.Errorf("%s and %s are on different filesystems (dev %d vs %d)", a, b, statA.Dev, statB.Dev)
	}
	return nil
}

// sameDevice reports whether a and b were stat'd from the same device.
// Split out from [sameFilesystem] so the comparison is testable without
// two real mount points.
func sameDevice(a, b unix.Stat_t) bool {
	return a.Dev == b.Dev
}

// --- nudge trigger ----------------------------------------------------

// writeNudgeTrigger writes an empty, atomically-renamed file named
// sessionID under triggerDir: the presence
// dubplate-import-nudge.path's DirectoryNotEmpty= watches, and the name
// U8's dubplate nudge-consume reads back to bypass quiescence for
// exactly this session. Called only from [Engine]'s OnPromoted hook,
// itself only invoked after promote's rename has already returned; any
// failure is logged and swallowed rather than returned, since the
// 15-minute import timer stays the backstop regardless -- a missed nudge
// costs latency, never correctness.
func writeNudgeTrigger(triggerDir, sessionID string) {
	if !validSessionID(sessionID) {
		slog.Error("nudge: refuse to write trigger for malformed session id", "session_id", sessionID)
		return
	}
	path := filepath.Join(triggerDir, sessionID)
	if err := atomicfile.WriteFile(path, nil, 0o644); err != nil {
		slog.Error("nudge: write trigger file", "session_id", sessionID, "err", err)
	}
}

// --- listener and run loop ---------------------------------------------

// freebindControl is a [net.ListenConfig.Control] hook that sets
// IP_FREEBIND on the listening socket before bind(2) runs, so serve can
// bind an address the host has not configured on any interface yet --
// the compose bridge's gateway IP at boot, before docker.service has
// brought the network up (N7). Unlike IP_TRANSPARENT, IP_FREEBIND needs
// no special capability.
func freebindControl(_, _ string, c syscall.RawConn) error {
	var sockErr error
	if err := c.Control(func(fd uintptr) {
		sockErr = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_FREEBIND, 1)
	}); err != nil {
		return err
	}
	return sockErr
}

// freebindListeners opens one TCP listener per addr in addrs, all on
// port, with [freebindControl] applied. Binding exactly this list --
// never "0.0.0.0" -- is what rejects a connection to any address not
// named in it, since nothing is listening there. Any failure closes
// every listener already opened before returning.
func freebindListeners(ctx context.Context, addrs []string, port int) ([]net.Listener, error) {
	lc := net.ListenConfig{Control: freebindControl}
	listeners := make([]net.Listener, 0, len(addrs))
	for _, addr := range addrs {
		l, err := lc.Listen(ctx, "tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
		if err != nil {
			for _, opened := range listeners {
				_ = opened.Close()
			}
			return nil, fmt.Errorf("listen %s: %w", addr, err)
		}
		listeners = append(listeners, l)
	}
	return listeners, nil
}

const (
	httpReadHeaderTimeout = 10 * time.Second
	httpReadTimeout       = 30 * time.Second
	httpWriteTimeout      = 30 * time.Second
	httpIdleTimeout       = 120 * time.Second
	// shutdownGrace bounds Run's own drain wait for an in-flight request;
	// systemd's own TimeoutStopSec (U12a) is the real backstop -- a
	// mid-PATCH SIGKILL is safe since tus resumes, so this only needs to
	// be generous, not exact.
	shutdownGrace = 25 * time.Second
	// sweepInterval is how often the partial-upload sweep runs; frequent
	// enough that an abandoned session's tus blobs and free-space
	// pressure never linger long against [config.UploadConfig.PartialTTL].
	sweepInterval = 15 * time.Minute
)

// Run starts every listener named by Serve.BindAddrs, starts the
// Engine's partial-upload sweep ticker, and serves until ctx is done.
// Shutdown then drains every in-flight request -- a mid-PATCH tus upload
// included -- before returning.
func (s *Server) Run(ctx context.Context) error {
	listeners, err := freebindListeners(ctx, s.cfg.Serve.BindAddrs, s.cfg.Serve.Port)
	if err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	sweepCtx, stopSweep := context.WithCancel(ctx)
	defer stopSweep()
	go s.engine.RunSweep(sweepCtx, sweepInterval)

	srv := &http.Server{
		Handler:           s.handler,
		ReadHeaderTimeout: httpReadHeaderTimeout,
		ReadTimeout:       httpReadTimeout,
		WriteTimeout:      httpWriteTimeout,
		IdleTimeout:       httpIdleTimeout,
	}
	return serveUntilDone(ctx, srv, listeners, shutdownGrace)
}

// serveUntilDone runs srv across every listener until ctx is done, then
// calls [http.Server.Shutdown] with a fresh shutdownTimeout-bounded
// context (derived via [context.WithoutCancel], since deriving directly
// from the already-canceled ctx would hand Shutdown a context that is
// already done). Split out from [Server.Run] so a test can exercise the
// drain behavior against a minimal handler without building a full
// Server.
func serveUntilDone(ctx context.Context, srv *http.Server, listeners []net.Listener, shutdownTimeout time.Duration) error {
	errCh := make(chan error, len(listeners))
	for _, l := range listeners {
		go func(l net.Listener) {
			err := srv.Serve(l)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
				return
			}
			errCh <- nil
		}(l)
	}

	select {
	case <-ctx.Done():
	case err := <-errCh:
		if err != nil {
			return fmt.Errorf("serve: %w", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}
