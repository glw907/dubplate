package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// discardTestLogger returns a logger this file's tests can pass
// without asserting on its output.
func discardTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// serveAuthEnviron returns authEnviron with DUBPLATE_STATE_DIR pointed
// at a fresh t.TempDir(): authEnviron's own fixed path is not one a
// test process may create the auth leaf's directories under.
func serveAuthEnviron(t *testing.T) []string {
	t.Helper()
	return replaceKey(authEnviron(t), "DUBPLATE_STATE_DIR", t.TempDir())
}

// replaceKey returns a copy of environ with name's entry replaced by
// value, appending it when environ carries no entry for name.
func replaceKey(environ []string, name, value string) []string {
	return append(dropKey(environ, name), name+"="+value)
}

// buildServeShell resolves environ through the production registry and
// [newServeShell]'s own three inputs, so a test exercises the verb's
// real construction path rather than a restatement of it.
func buildServeShell(t *testing.T, environ []string) (*serve.Shell, <-chan struct{}, error) {
	t.Helper()
	return buildServeShellWithContext(t, t.Context(), environ)
}

// buildServeShellWithContext is [buildServeShell] with the caller's own
// context, for a test that must control when the sweep driver's context
// is canceled rather than relying on t.Context()'s own lifetime.
func buildServeShellWithContext(t *testing.T, ctx context.Context, environ []string) (*serve.Shell, <-chan struct{}, error) {
	t.Helper()
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig: %v", err)
	}
	authCfg, err := resolveAuthConfig(environ)
	if err != nil {
		t.Fatalf("resolveAuthConfig: %v", err)
	}
	return newServeShell(ctx, prefix, cfg, authCfg)
}

// productionRegistry builds a registry exactly as newRootCmd does, so
// this file's own tests exercise the same collision and derivation
// surface the binary does, on registry.go:24-31's own precedent. It
// calls newRootCmd's own shared constructor (main.go) rather than
// restating its register calls, so the two can never drift apart.
func productionRegistry() *config.Registry {
	return newProductionRegistry()
}

func TestRegisterServeConfig_ClaimsThreeKeysAndRetiresReservedClass(t *testing.T) {
	reg := productionRegistry()

	want := map[string]struct {
		kind  config.ThresholdKind
		deflt string
	}{
		"DUBPLATE_SERVE_BIND_ADDRS": {config.KindString, "127.0.0.1"},
		"DUBPLATE_SERVE_PORT":       {config.KindInt, "8080"},
		"DUBPLATE_SERVE_PREFIX":     {config.KindString, "/dubplate"},
	}

	got := make(map[string]config.Threshold)
	for _, th := range reg.Thresholds() {
		got[th.Key] = th
	}
	for key, w := range want {
		th, ok := got[key]
		if !ok {
			t.Errorf("Thresholds() missing %s", key)
			continue
		}
		if th.Kind != w.kind {
			t.Errorf("%s.Kind = %v, want %v", key, th.Kind, w.kind)
		}
		if th.Default != w.deflt {
			t.Errorf("%s.Default = %q, want %q", key, th.Default, w.deflt)
		}
	}

	for _, class := range reg.ReservedClasses() {
		if class.Name == "serve" {
			t.Error(`ReservedClasses() still carries the "serve" class, want it retired`)
		}
	}
}

func TestResolveServeConfig(t *testing.T) {
	tests := []struct {
		name       string
		environ    []string
		wantErr    error
		wantAddrs  []string
		wantPort   int
		wantPrefix string
	}{
		{
			name:       "defaults",
			environ:    nil,
			wantAddrs:  []string{"127.0.0.1"},
			wantPort:   8080,
			wantPrefix: "/dubplate",
		},
		{
			name: "configured",
			environ: []string{
				"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,::1",
				"DUBPLATE_SERVE_PORT=9090",
				"DUBPLATE_SERVE_PREFIX=/library",
			},
			wantAddrs:  []string{"127.0.0.1", "::1"},
			wantPort:   9090,
			wantPrefix: "/library",
		},
		{
			name:    "invalid prefix",
			environ: []string{"DUBPLATE_SERVE_PREFIX=no-leading-separator"},
			wantErr: serve.ErrInvalidPrefix,
		},
		{
			name:    "unparseable bind list",
			environ: []string{"DUBPLATE_SERVE_BIND_ADDRS=not-an-address"},
			wantErr: serve.ErrInvalidBindAddr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prefix, addrs, port, err := resolveServeConfig(tt.environ)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("resolveServeConfig() error = %v, want it to wrap %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveServeConfig() error = %v, want nil", err)
			}
			if !slices.Equal(addrs, tt.wantAddrs) {
				t.Errorf("addrs = %v, want %v", addrs, tt.wantAddrs)
			}
			if port != tt.wantPort {
				t.Errorf("port = %d, want %d", port, tt.wantPort)
			}
			if prefix.String() != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", prefix.String(), tt.wantPrefix)
			}
		})
	}
}

// TestNewServeCmd_RefusesInvalidConfigWithoutListening proves the verb
// validates the wall and resolves configuration, refusing before ever
// calling serve.Run: an unsatisfied owned wall, an invalid prefix, or
// an unparseable bind list each surfaces through RunE with the
// underlying sentinel intact, and no case opens a listener, since all
// three refusals precede resolution's own return.
//
// Each case runs under an already-canceled context so a regression
// reports as a failure rather than a hang: should a refusal stop
// firing, serve.Run returns at once instead of serving until the
// gate's own timeout.
func TestNewServeCmd_RefusesInvalidConfigWithoutListening(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		wantErr error
	}{
		{
			name:    "unsatisfied owned wall",
			environ: dropKey(baseEnviron(), "DUBPLATE_ADMIN_EMAIL"),
			wantErr: config.ErrMissing,
		},
		{
			name:    "invalid prefix",
			environ: append(authEnviron(t), "DUBPLATE_SERVE_PREFIX=no-leading-separator"),
			wantErr: serve.ErrInvalidPrefix,
		},
		{
			name:    "unparseable bind list",
			environ: append(authEnviron(t), "DUBPLATE_SERVE_BIND_ADDRS=not-an-address"),
			wantErr: serve.ErrInvalidBindAddr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := newServeCmd(tt.environ, productionRegistry(), &stdout, &stderr)

			ctx, cancel := context.WithCancel(t.Context())
			cancel()

			err := cmd.ExecuteContext(ctx)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Execute() error = %v, want it to wrap %v", err, tt.wantErr)
			}
		})
	}
}

// TestNewServeCmd_WiredIntoRootCmd proves newServeCmd's line sits
// inside newRootCmd's AddCommand block, beside newShadowCmd's: a verb
// constructed but never added is unreachable through the command tree
// cobra dispatches, and root.Find alone would not tell the two apart
// from a verb reachable only because some other command re-added it.
//
// The second half proves the wiring carries the production registry
// and the process environment rather than a construction that only
// looks right: dispatching "serve" through the root command with an
// invalid prefix in the environment refuses with the resolution
// sentinel, which can only happen if the added command resolved
// configuration. No listener opens, since resolution fails first.
func TestNewServeCmd_WiredIntoRootCmd(t *testing.T) {
	var serveCmd, shadowCmd bool
	for _, child := range newRootCmd().Commands() {
		switch child.Use {
		case "serve":
			serveCmd = true
		case "shadow":
			shadowCmd = true
		}
	}
	if !serveCmd {
		t.Error("newRootCmd().Commands() carries no serve verb, want it added beside shadow")
	}
	if !shadowCmd {
		t.Fatal("newRootCmd().Commands() carries no shadow verb; the AddCommand block this test reads against has moved")
	}

	// newRootCmd reads os.Environ() at construction, so the process
	// environment is this half's input. A developer's shell carries
	// DUBPLATE_* names of its own, and the wall would refuse them
	// before resolution ever ran; clearing each to the empty string is
	// what the wall reads as unset (config.wall.get's own convention).
	for _, kv := range os.Environ() {
		if name, _, ok := strings.Cut(kv, "="); ok && strings.HasPrefix(name, "DUBPLATE_") {
			t.Setenv(name, "")
		}
	}
	for _, kv := range authEnviron(t) {
		name, value, _ := strings.Cut(kv, "=")
		t.Setenv(name, value)
	}
	t.Setenv("DUBPLATE_SERVE_PREFIX", "no-leading-separator")

	root := newRootCmd()
	root.SetArgs([]string{"serve"})
	var stdout, stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)

	if err := root.Execute(); !errors.Is(err, serve.ErrInvalidPrefix) {
		t.Fatalf("Execute([]string{\"serve\"}) error = %v, want it to wrap %v", err, serve.ErrInvalidPrefix)
	}
}

// TestDoctorCmd_CarriesServeThresholdsAndNoServeReservedRow proves the
// net row change registerServeConfig plus the keys.go deletion makes:
// minus one reserved class, plus three thresholds.
func TestDoctorCmd_CarriesServeThresholdsAndNoServeReservedRow(t *testing.T) {
	reg := productionRegistry()
	lines := doctorLines(t, reg)

	for _, key := range []string{"DUBPLATE_SERVE_BIND_ADDRS", "DUBPLATE_SERVE_PORT", "DUBPLATE_SERVE_PREFIX"} {
		found := false
		for _, line := range lines {
			if strings.HasPrefix(line, key) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("doctor output missing a row for %s", key)
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "serve ") || strings.HasPrefix(line, "serve\t") {
			t.Errorf("doctor output still carries a serve reserved-class row: %q", line)
		}
	}
}

// TestNewServeShell_RemovesTheSessionCookieAtTheFallthrough closes 5a's
// declared coverage gap 4. The shell rung 5a shipped was constructed
// with an empty cookie name, which removes nothing, so the removal was
// proved for an arbitrary name and never for the one the server
// actually sets. This row runs the shell the serve verb constructs,
// with the cookie name written out rather than read from the constant
// under test, and asserts the property behaviourally: the session
// cookie is gone at the fallthrough and every other cookie arrives
// intact, so a request leaving the owned surface carries no credential
// that authenticates against this server.
func TestNewServeShell_RemovesTheSessionCookieAtTheFallthrough(t *testing.T) {
	shell, _, err := buildServeShell(t, serveAuthEnviron(t))
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}

	var got http.Header
	shell.SetFallthrough(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
	}))

	r := httptest.NewRequest(http.MethodGet, "/outside", nil)
	r.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: "a-session-value"})
	r.AddCookie(&http.Cookie{Name: "navidrome_session", Value: "another-value"})
	shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

	if got == nil {
		t.Fatal("the fallthrough handler was never invoked")
	}
	arrived := &http.Request{Header: got}
	if c, err := arrived.Cookie("__Host-dubplate_session"); err == nil {
		t.Errorf("fallthrough request carries %q = %q, want it removed", c.Name, c.Value)
	}
	c, err := arrived.Cookie("navidrome_session")
	if err != nil {
		t.Fatalf("fallthrough request carries no %q cookie: %v", "navidrome_session", err)
	}
	if c.Value != "another-value" {
		t.Errorf("cookie %q = %q, want %q", c.Name, c.Value, "another-value")
	}
}

// authRoutePOST issues a same-origin POST against target with body,
// the one shape [methodOnly] and the origin check both admit.
func authRoutePOST(handler http.Handler, target, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Sec-Fetch-Site", "same-origin")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// TestNewServeShell_MountsTheThreeAuthRoutes proves the verb's own
// construction registers request-link, consume-link, and sign-out
// under the owned prefix, and that an unregistered path under the same
// prefix still terminates at 404 rather than falling through.
func TestNewServeShell_MountsTheThreeAuthRoutes(t *testing.T) {
	environ := serveAuthEnviron(t)
	shell, _, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	handler := shell.Handler()

	for _, tt := range []struct {
		name    string
		path    string
		body    string
		notWant int
	}{
		{"request-link", requestLinkRouteName, "address=member@example.com", http.StatusNotFound},
		{"consume-link", consumeLinkRouteName, "token=not-a-real-token", http.StatusNotFound},
		{"sign-out", signOutRouteName, "", http.StatusNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			rec := authRoutePOST(handler, prefix.String()+"/"+tt.path, tt.body)
			if rec.Code == tt.notWant {
				t.Errorf("status = %d, want anything but %d: the route is not mounted", rec.Code, tt.notWant)
			}
		})
	}

	rec := authRoutePOST(handler, prefix.String()+"/no-such-route", "")
	if rec.Code != http.StatusNotFound {
		t.Errorf("unregistered owned path status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

// TestNewServeShell_MissingRosterRefusesBeforeConstruction proves that
// a missing roster file refuses the shell's own construction, naming
// the path, rather than building a shell that would answer requests
// with no roster behind it.
func TestNewServeShell_MissingRosterRefusesBeforeConstruction(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "no-such-roster.json")
	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_ROSTER_FILE", missingPath)

	_, _, err := buildServeShell(t, environ)
	if !errors.Is(err, auth.ErrRosterMissing) {
		t.Fatalf("buildServeShell() error = %v, want it to wrap auth.ErrRosterMissing", err)
	}
	if !strings.Contains(err.Error(), missingPath) {
		t.Errorf("error %q does not name the missing roster path %q", err, missingPath)
	}
}

// TestNewServeShell_StateChangingFallthroughRequestIsDenied proves no
// bypass pattern is registered: a cookie-carrying POST outside the
// owned prefix, carrying neither Sec-Fetch-Site nor Origin, is refused
// by the origin check rather than reaching the fallthrough.
func TestNewServeShell_StateChangingFallthroughRequestIsDenied(t *testing.T) {
	shell, _, err := buildServeShell(t, serveAuthEnviron(t))
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	reached := false
	shell.SetFallthrough(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { reached = true }))

	req := httptest.NewRequest(http.MethodPost, "/outside", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: "a-session-value"})
	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
	if reached {
		t.Error("the fallthrough handler was reached; a bypass pattern must not exempt it")
	}
}

// TestStartNonceSweep_RunsAndStopsWithContext is the sweep driver's
// leak assertion at the verb's own wiring: canceling ctx stops the
// goroutine [startNonceSweep] started, proved by the channel that
// goroutine closes on return rather than by inspecting its source.
func TestStartNonceSweep_RunsAndStopsWithContext(t *testing.T) {
	stateDir := t.TempDir()
	store, err := auth.Open(stateDir)
	if err != nil {
		t.Fatalf("auth.Open: %v", err)
	}
	nonces, _, err := auth.OpenNonceStore(store, time.Hour, magicLinkExpiryCeiling, time.Now())
	if err != nil {
		t.Fatalf("auth.OpenNonceStore: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := startNonceSweep(ctx, nonces, discardTestLogger())
	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("startNonceSweep left its goroutine running after ctx was canceled")
	}
}

// TestNewServeShell_SweepDriverStopsWithTheGivenContext is the leak
// assertion against newServeShell's own wiring, not against
// [startNonceSweep] called directly: it proves the driver goroutine
// newServeShell starts is bound to the ctx argument it was passed,
// rather than to a context of its own, by canceling that argument and
// waiting on the channel newServeShell returns.
func TestNewServeShell_SweepDriverStopsWithTheGivenContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	_, done, err := buildServeShellWithContext(t, ctx, serveAuthEnviron(t))
	if err != nil {
		t.Fatalf("buildServeShellWithContext() error = %v, want nil", err)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the sweep driver did not stop after the passed context was canceled")
	}
}

// TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree holds
// security review finding 2, R-5c9's routing test: a link minted
// through the real handler config, the prefix and the public origin
// [newServeShell] itself resolves, lands on a path the owned prefix
// owns and a registered "login" subtree answers, never the
// fallthrough. It runs for both public-URL forms, a bare origin and one
// carrying a trailing separator, and asserts the same link path for
// each, so a hand concatenation that dropped url.JoinPath's cleaning
// (which would double the separator for the trailing-slash form alone)
// fails it.
//
// Fixture state: the fallthrough's recorded query starts at a sentinel
// a real request's RawQuery can never equal, so a pass cannot be
// explained by a 404 that terminates the request before the
// fallthrough is even reached.
func TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree(t *testing.T) {
	origins := []string{"https://h.example.test", "https://h.example.test/"}
	gotPaths := make([]string, len(origins))

	for i, origin := range origins {
		t.Run(origin, func(t *testing.T) {
			environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_PUBLIC_URL", origin)

			cfg, err := config.Load(environ, productionRegistry())
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			prefix, _, _, err := resolveServeConfig(environ)
			if err != nil {
				t.Fatalf("resolveServeConfig: %v", err)
			}
			authCfg, err := resolveAuthConfig(environ)
			if err != nil {
				t.Fatalf("resolveAuthConfig: %v", err)
			}
			shell, _, err := newServeShell(t.Context(), prefix, cfg, authCfg)
			if err != nil {
				t.Fatalf("newServeShell: %v", err)
			}

			handlersConfig := authHandlersConfig(prefix, cfg, authCfg, nil, nil, nil, nil, nil, nil, nil)
			if handlersConfig.SessionAbsoluteCap != authCfg.sessionAbsoluteCap {
				t.Errorf("authHandlersConfig SessionAbsoluteCap = %v, want %v", handlersConfig.SessionAbsoluteCap, authCfg.sessionAbsoluteCap)
			}
			if handlersConfig.Prefix != prefix.String() {
				t.Errorf("authHandlersConfig Prefix = %q, want %q", handlersConfig.Prefix, prefix.String())
			}

			var loginReached bool
			shell.Register("login", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				loginReached = true
				w.WriteHeader(http.StatusOK)
			}))
			const sentinel = "fallthrough not invoked"
			fallthroughQuery := sentinel
			shell.SetFallthrough(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				fallthroughQuery = r.URL.RawQuery
			}))

			link, err := auth.LinkURL(handlersConfig.PublicOrigin, handlersConfig.Prefix, "fixture-token")
			if err != nil {
				t.Fatalf("auth.LinkURL: %v", err)
			}
			u, err := url.Parse(link)
			if err != nil {
				t.Fatalf("parse minted link: %v", err)
			}
			if !prefix.Owns(u.Path) {
				t.Fatalf("minted link path %q is not owned by the prefix %q", u.Path, prefix.String())
			}
			gotPaths[i] = u.Path

			r := httptest.NewRequest(http.MethodGet, u.Path+"?"+u.RawQuery, nil)
			shell.Handler().ServeHTTP(httptest.NewRecorder(), r)

			if !loginReached {
				t.Error("the registered login subtree was never reached")
			}
			if fallthroughQuery != sentinel {
				t.Errorf("fallthrough saw RawQuery=%q, want it never reached", fallthroughQuery)
			}
		})
	}

	if gotPaths[0] != gotPaths[1] {
		t.Errorf("minted link path = %q for %q and %q for %q, want the same path for both origin forms", gotPaths[0], origins[0], gotPaths[1], origins[1])
	}
}
