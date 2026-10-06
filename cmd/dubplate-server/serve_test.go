package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"testing"
	"testing/fstest"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/web"
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
// real construction path rather than a restatement of it. The shell
// logs to slog.Default, as the serve verb's does.
func buildServeShell(t *testing.T, environ []string) (*serve.Shell, <-chan struct{}, error) {
	t.Helper()
	return buildServeShellWithContext(t, t.Context(), environ, slog.Default())
}

// buildServeShellWithLogger is [buildServeShell] with the shell's
// logger supplied, so a test reads every record the shell and its
// handlers write without swapping slog.Default.
func buildServeShellWithLogger(t *testing.T, environ []string, logger *slog.Logger) (*serve.Shell, <-chan struct{}, error) {
	t.Helper()
	return buildServeShellWithContext(t, t.Context(), environ, logger)
}

// buildServeShellWithContext is [buildServeShell] with the caller's own
// context, for a test that must control when the sweep's context
// is canceled rather than relying on t.Context()'s own lifetime.
func buildServeShellWithContext(t *testing.T, ctx context.Context, environ []string, logger *slog.Logger) (*serve.Shell, <-chan struct{}, error) {
	t.Helper()
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	prefix, addrs, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig: %v", err)
	}
	authCfg, err := resolveAuthConfig(environ)
	if err != nil {
		t.Fatalf("resolveAuthConfig: %v", err)
	}
	return newServeShell(ctx, environ, prefix, addrs, cfg, authCfg, logger)
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

// TestNewServeShell_OperatorCSSRefusesTheRosterFile proves the operator
// CSS rule on the serve verb's own construction path: the key pointed
// at the roster file this environment serves from, at a ".css" hard
// link to it, or at a ".css" file in the state directory refuses
// startup, naming the key and the rule, and neither the path nor a
// byte of the roster reaches the error. The last two rows fail unless
// the verb hands the web routes its own roster path and state
// directory. A shell that admitted any of them would build and publish
// the file at <prefix>/operator.css.
func TestNewServeShell_OperatorCSSRefusesTheRosterFile(t *testing.T) {
	tests := []struct {
		name    string
		path    func(t *testing.T, rosterPath, stateDir string) string
		wantErr string
	}{
		{
			name:    "the roster file",
			path:    func(_ *testing.T, rosterPath, _ string) string { return rosterPath },
			wantErr: `DUBPLATE_OPERATOR_CSS_FILE: the path must end in ".css"`,
		},
		{
			name: "a .css hard link to the roster",
			path: func(t *testing.T, rosterPath, _ string) string {
				p := filepath.Join(t.TempDir(), "operator.css")
				if err := os.Link(rosterPath, p); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: "DUBPLATE_OPERATOR_CSS_FILE: the file is the roster",
		},
		{
			name: "a .css file in the state directory",
			path: func(t *testing.T, _, stateDir string) string {
				p := filepath.Join(stateDir, "operator.css")
				if err := os.WriteFile(p, []byte("body{}"), 0o600); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErr: "DUBPLATE_OPERATOR_CSS_FILE: the file is under the state directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := serveAuthEnviron(t)
			rosterPath := valueOf(t, environ, "DUBPLATE_ROSTER_FILE")
			roster, err := os.ReadFile(rosterPath)
			if err != nil {
				t.Fatalf("read roster fixture: %v", err)
			}
			path := tt.path(t, rosterPath, valueOf(t, environ, "DUBPLATE_STATE_DIR"))
			environ = replaceKey(environ, "DUBPLATE_OPERATOR_CSS_FILE", path)

			shell, _, err := buildServeShell(t, environ)
			if err == nil {
				rec := httptest.NewRecorder()
				shell.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dubplate/operator.css", nil))
				t.Fatalf("buildServeShell: got nil error, want the operator CSS refusal; GET operator.css answers %d, serving the roster's bytes: %v",
					rec.Code, bytes.Equal(rec.Body.Bytes(), roster))
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
			if strings.Contains(err.Error(), path) || strings.Contains(err.Error(), string(roster)) {
				t.Errorf("error = %q echoes the path or the roster's bytes", err.Error())
			}
		})
	}
}

// TestNewServeShell_BrandingRefusesStartup proves a refused branding
// value stops the serve verb's own construction, naming the key.
func TestNewServeShell_BrandingRefusesStartup(t *testing.T) {
	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_BRANDING", "--color-base-100: #ffffff; --color-primary: #fff}")
	_, _, err := buildServeShell(t, environ)
	if err == nil {
		t.Fatal("buildServeShell: got nil error, want the branding refusal")
	}
	if !strings.HasPrefix(err.Error(), "DUBPLATE_BRANDING: ") || strings.Contains(err.Error(), "#fff}") {
		t.Errorf("error = %q, want it to name DUBPLATE_BRANDING and not echo the value", err.Error())
	}
}

// valueOf returns the value environ carries for name, failing t when
// it carries none.
func valueOf(t *testing.T, environ []string, name string) string {
	t.Helper()
	for _, kv := range environ {
		if n, v, _ := strings.Cut(kv, "="); n == name {
			return v
		}
	}
	t.Fatalf("environ carries no %s", name)
	return ""
}

// TestNewServeShell_StateChangingRequestOutsideThePrefixIsDenied proves
// no bypass pattern is registered: a cookie-carrying POST outside the
// owned prefix, carrying neither Sec-Fetch-Site nor Origin, is refused
// with 403 by the origin check, a status the root 404 cannot produce.
func TestNewServeShell_StateChangingRequestOutsideThePrefixIsDenied(t *testing.T) {
	shell, _, err := buildServeShell(t, serveAuthEnviron(t))
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/outside", nil)
	req.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: "a-session-value"})
	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusForbidden)
	}
}

// TestStartNonceSweep_RunsAndStopsWithContext is the sweep's leak
// assertion at the verb's own wiring: canceling ctx stops the
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

// TestNewServeShell_NonceSweepStopsWithTheGivenContext is the leak
// assertion against newServeShell's own wiring, not against
// [startNonceSweep] called directly: it proves the sweep goroutine
// newServeShell starts is bound to the ctx argument it was passed,
// rather than to a context of its own, by canceling that argument and
// waiting on the channel newServeShell returns.
func TestNewServeShell_NonceSweepStopsWithTheGivenContext(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	_, done, err := buildServeShellWithContext(t, ctx, serveAuthEnviron(t), discardTestLogger())
	if err != nil {
		t.Fatalf("buildServeShellWithContext() error = %v, want nil", err)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("the nonce sweep did not stop after the passed context was canceled")
	}
}

// TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree holds
// security review finding 2, R-5c9's routing test: a link minted
// through the real handler config, the prefix and the public origin
// [newServeShell] itself resolves, lands on a path the owned prefix
// owns and a registered "login" subtree answers. It runs for both
// public-URL forms, a bare origin and one carrying a trailing
// separator, and asserts the same link path for each, so a hand
// concatenation that dropped url.JoinPath's cleaning (which would
// double the separator for the trailing-slash form alone) fails it.
//
// Fixture state: the assertion is a 200, a status the terminator's 404
// and a path outside the prefix's 404 cannot produce, so a pass is
// never explained by a request that no subtree answered.
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
			prefix, addrs, _, err := resolveServeConfig(environ)
			if err != nil {
				t.Fatalf("resolveServeConfig: %v", err)
			}
			authCfg, err := resolveAuthConfig(environ)
			if err != nil {
				t.Fatalf("resolveAuthConfig: %v", err)
			}
			shell, _, err := newServeShell(t.Context(), environ, prefix, addrs, cfg, authCfg, discardTestLogger())
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
			rec := httptest.NewRecorder()
			shell.Handler().ServeHTTP(rec, r)

			// A minted link routes to the login subtree's "login/link"
			// document, so the registered-subtree proof is that it
			// answers 200.
			if rec.Code != http.StatusOK {
				t.Errorf("the registered login subtree answered %d, want 200", rec.Code)
			}
		})
	}

	if gotPaths[0] != gotPaths[1] {
		t.Errorf("minted link path = %q for %q and %q for %q, want the same path for both origin forms", gotPaths[0], origins[0], gotPaths[1], origins[1])
	}
}

// rosterAdminAddress and rosterMemberAddress are the two fixture
// members [rosterAPIEnviron] seeds: one admin, one not.
const (
	rosterAdminAddress  = "admin@example.com"
	rosterMemberAddress = "member@example.com"
)

// rosterAPIEnviron returns serveAuthEnviron with DUBPLATE_ROSTER_FILE
// replaced by a fixture roster holding one admin (with a
// streamer_username, so a GET has every field shape to answer) and one
// non-admin, and returns the roster's path alongside for a test that
// reads it directly.
func rosterAPIEnviron(t *testing.T) ([]string, string) {
	t.Helper()
	rosterPath := filepath.Join(t.TempDir(), "roster.json")
	const members = `{"schema_version":1,"members":[
		{"address":"` + rosterAdminAddress + `","display_name":"Group Admin","streamer_username":"groupadmin","admin":true},
		{"address":"` + rosterMemberAddress + `","display_name":"Member","admin":false}
	]}`
	if err := os.WriteFile(rosterPath, []byte(members), 0o600); err != nil {
		t.Fatalf("write fixture roster: %v", err)
	}
	return replaceKey(serveAuthEnviron(t), "DUBPLATE_ROSTER_FILE", rosterPath), rosterPath
}

// environValue returns the value environ names name under, or "" when
// no entry names it.
func environValue(environ []string, name string) string {
	for _, kv := range environ {
		n, v, _ := strings.Cut(kv, "=")
		if n == name {
			return v
		}
	}
	return ""
}

// mintServeSession mints a session cookie value for address, signed
// under environ's own DUBPLATE_SESSION_SECRET and the production
// lifetime and absolute cap [resolveAuthConfig] resolves from environ,
// issued at issued: the same construction path [newServeShell]'s own
// seam verifies against, so a test session is one a real login would
// have minted.
func mintServeSession(t *testing.T, environ []string, address string, issued time.Time) string {
	t.Helper()
	authCfg, err := resolveAuthConfig(environ)
	if err != nil {
		t.Fatalf("resolveAuthConfig: %v", err)
	}
	secret := environValue(environ, "DUBPLATE_SESSION_SECRET")
	value, _, err := auth.MintSession([]byte(secret), address, issued, authCfg.sessionLifetime, authCfg.sessionAbsoluteCap)
	if err != nil {
		t.Fatalf("auth.MintSession(%s): %v", address, err)
	}
	return value
}

// sessionRequest builds a request to target carrying session as its
// cookie, sfs as Sec-Fetch-Site when non-empty, and body as a
// urlencoded form when non-empty.
func sessionRequest(method, target, session, sfs, body string) *http.Request {
	r := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	if sfs != "" {
		r.Header.Set("Sec-Fetch-Site", sfs)
	}
	if session != "" {
		r.AddCookie(&http.Cookie{Name: auth.SessionCookieName, Value: session})
	}
	return r
}

// headersEqualButDate reports whether a and b carry the same header
// set once each header named "Date" is dropped from both: the wire
// server, never [httptest.ResponseRecorder], is what stamps a
// response's own Date, but the comparison drops it anyway so this
// helper reads the same under either.
func headersEqualButDate(a, b http.Header) bool {
	ac, bc := a.Clone(), b.Clone()
	ac.Del("Date")
	bc.Del("Date")
	if len(ac) != len(bc) {
		return false
	}
	for k, v := range ac {
		if !slices.Equal(v, bc[k]) {
			return false
		}
	}
	return true
}

// wantNotFoundBody is the body [http.NotFound] writes, [serve.Terminator]'s
// own answer and the admin gate's refusal alike.
const wantNotFoundBody = "404 page not found\n"

// TestNewServeShell_RosterPOSTWithoutOriginSignalIsDenied proves the
// roster API sits behind the shell's own origin check like every other
// state-changing route: a POST carrying neither Origin nor
// Sec-Fetch-Site is refused before it ever reaches the admin gate, and
// the roster file is unchanged.
func TestNewServeShell_RosterPOSTWithoutOriginSignalIsDenied(t *testing.T) {
	environ, rosterPath := rosterAPIEnviron(t)
	shell, _, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	session := mintServeSession(t, environ, rosterAdminAddress, time.Now())

	before, err := os.ReadFile(rosterPath)
	if err != nil {
		t.Fatalf("read fixture before request: %v", err)
	}

	req := sessionRequest(http.MethodPost, prefix.String()+"/roster", session, "", validPost)
	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, body = %s, want %d", rec.Code, rec.Body.String(), http.StatusForbidden)
	}

	after, err := os.ReadFile(rosterPath)
	if err != nil {
		t.Fatalf("read fixture after request: %v", err)
	}
	if !bytes.Equal(before, after) {
		t.Errorf("roster file changed: before %q, after %q", before, after)
	}
}

// validPost is a form every field check accepts for a new member,
// [internal/auth]'s own fixture body (admin_test.go), restated here so
// this file's tests need no cross-package export.
const validPost = "address=new%40example.com&display_name=New+Member&streamer_username=newmember"

// captureDefaultLogger swaps slog.Default() for a logger writing JSON
// lines to the returned buffer, restoring the prior default when t
// ends. [newServeShell] builds every shell through [serve.NewShell]
// with a nil logger, which falls back to slog.Default
// (internal/serve/serve.go:43-48), so this is the seam this package has
// to read a shell's own log line without widening newServeShell's own
// signature.
func captureDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prior := slog.Default()
	slog.SetDefault(slog.New(slog.NewJSONHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prior) })
	return &buf
}

// TestNewServeShell_RosterSameOriginPOSTAddsMemberAndAudits proves the
// same POST, carrying Sec-Fetch-Site: same-origin, reaches the admin
// gate, adds the member, and writes one audit record naming the actor
// and the cookie channel: the outcome mapping and the audit record's
// own shape are internal/auth's own proof (admin_test.go); this test
// proves the shell's composition reaches them at all.
func TestNewServeShell_RosterSameOriginPOSTAddsMemberAndAudits(t *testing.T) {
	environ, rosterPath := rosterAPIEnviron(t)
	// [serve.NewShell] reads slog.Default() once, at construction, into
	// its own stored field, so the capture must be in place before the
	// shell (and the AdminConfig.Logger it hands the roster API) is
	// built, not merely before the request.
	logs := captureDefaultLogger(t)
	shell, _, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	session := mintServeSession(t, environ, rosterAdminAddress, time.Now())

	req := sessionRequest(http.MethodPost, prefix.String()+"/roster", session, "same-origin", validPost)
	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s, want %d", rec.Code, rec.Body.String(), http.StatusCreated)
	}

	after, err := os.ReadFile(rosterPath)
	if err != nil {
		t.Fatalf("read fixture after request: %v", err)
	}
	if !strings.Contains(string(after), "new@example.com") {
		t.Errorf("roster file after the POST does not name the new member: %s", after)
	}

	auditCount := strings.Count(logs.String(), `"msg":"roster member added"`)
	if auditCount != 1 {
		t.Errorf("audit records for \"roster member added\" = %d, want exactly 1; log = %s", auditCount, logs.String())
	}
	if !strings.Contains(logs.String(), `"actor":"`+rosterAdminAddress+`"`) {
		t.Errorf("audit record does not name the actor %q: %s", rosterAdminAddress, logs.String())
	}
	if !strings.Contains(logs.String(), `"channel":"cookie"`) {
		t.Errorf("audit record does not name the cookie channel: %s", logs.String())
	}
}

// TestNewServeShell_RosterNonAdminIndistinguishableFromUnknownPath
// proves that a non-admin's GET and POST to the roster route answer
// exactly what the same requests to an unregistered owned path answer:
// same status, same body, every header equal but Date, and no
// Set-Cookie on either side. The POSTs carry Sec-Fetch-Site:
// same-origin so both reach the gate and the terminator rather than the
// origin check. It then proves the same route answers an admin's GET
// with the roster list, so the 404s above are the non-admin's and not
// the path's.
func TestNewServeShell_RosterNonAdminIndistinguishableFromUnknownPath(t *testing.T) {
	environ, _ := rosterAPIEnviron(t)
	shell, _, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	nonAdminSession := mintServeSession(t, environ, rosterMemberAddress, time.Now())
	adminSession := mintServeSession(t, environ, rosterAdminAddress, time.Now())
	handler := shell.Handler()

	rosterPath := prefix.String() + "/roster"
	unknownPath := prefix.String() + "/no-such-subtree"

	run := func(method, target, sfs, body string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, sessionRequest(method, target, nonAdminSession, sfs, body))
		return rec
	}

	getRoster := run(http.MethodGet, rosterPath, "", "")
	getUnknown := run(http.MethodGet, unknownPath, "", "")
	if getRoster.Code != getUnknown.Code {
		t.Errorf("GET status: roster = %d, unknown = %d, want equal", getRoster.Code, getUnknown.Code)
	}
	if getRoster.Body.String() != getUnknown.Body.String() {
		t.Errorf("GET body: roster = %q, unknown = %q, want equal", getRoster.Body.String(), getUnknown.Body.String())
	}
	if !headersEqualButDate(getRoster.Header(), getUnknown.Header()) {
		t.Errorf("GET headers differ (Date excepted): roster = %v, unknown = %v", getRoster.Header(), getUnknown.Header())
	}
	if getRoster.Header().Get("Set-Cookie") != "" || getUnknown.Header().Get("Set-Cookie") != "" {
		t.Errorf("GET carries Set-Cookie: roster = %q, unknown = %q, want neither", getRoster.Header().Get("Set-Cookie"), getUnknown.Header().Get("Set-Cookie"))
	}

	postRoster := run(http.MethodPost, rosterPath, "same-origin", validPost)
	postUnknown := run(http.MethodPost, unknownPath, "same-origin", validPost)
	if postRoster.Code != postUnknown.Code {
		t.Errorf("POST status: roster = %d, unknown = %d, want equal", postRoster.Code, postUnknown.Code)
	}
	if postRoster.Body.String() != postUnknown.Body.String() {
		t.Errorf("POST body: roster = %q, unknown = %q, want equal", postRoster.Body.String(), postUnknown.Body.String())
	}
	if !headersEqualButDate(postRoster.Header(), postUnknown.Header()) {
		t.Errorf("POST headers differ (Date excepted): roster = %v, unknown = %v", postRoster.Header(), postUnknown.Header())
	}
	if postRoster.Header().Get("Set-Cookie") != "" || postUnknown.Header().Get("Set-Cookie") != "" {
		t.Errorf("POST carries Set-Cookie: roster = %q, unknown = %q, want neither", postRoster.Header().Get("Set-Cookie"), postUnknown.Header().Get("Set-Cookie"))
	}

	adminGetRec := httptest.NewRecorder()
	handler.ServeHTTP(adminGetRec, sessionRequest(http.MethodGet, rosterPath, adminSession, "", ""))
	if adminGetRec.Code != http.StatusOK {
		t.Fatalf("admin GET status = %d, body = %s, want %d", adminGetRec.Code, adminGetRec.Body.String(), http.StatusOK)
	}
	if !strings.Contains(adminGetRec.Body.String(), rosterAdminAddress) {
		t.Errorf("admin GET body = %s, want it to name %q", adminGetRec.Body.String(), rosterAdminAddress)
	}
}

// TestNewServeShell_RosterExactPathOnly proves that only the exact
// path answers the roster API, and every deeper path,
// trailing-slash forms included, answers the terminator's own 404.
// Fixture state: the same admin's GET to the exact path answers the
// list in this same test, so the 404s below are the path's and not the
// session's.
func TestNewServeShell_RosterExactPathOnly(t *testing.T) {
	environ, _ := rosterAPIEnviron(t)
	shell, _, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	adminSession := mintServeSession(t, environ, rosterAdminAddress, time.Now())
	handler := shell.Handler()

	get := func(path string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, sessionRequest(http.MethodGet, path, adminSession, "", ""))
		return rec
	}

	list := get(prefix.String() + "/roster")
	if list.Code != http.StatusOK {
		t.Fatalf("admin GET %s status = %d, body = %s, want %d", "/roster", list.Code, list.Body.String(), http.StatusOK)
	}

	for _, deeper := range []string{"/roster/x", "/roster/", "/roster/roster"} {
		t.Run(deeper, func(t *testing.T) {
			rec := get(prefix.String() + deeper)
			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, body = %s, want %d", rec.Code, rec.Body.String(), http.StatusNotFound)
			}
			if rec.Body.String() != wantNotFoundBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), wantNotFoundBody)
			}
		})
	}
}

// TestNewServeShell_RosterAndMemberVerbShareReload proves "member list
// reads what the API wrote, and the API reads what the verb wrote".
// Fixture state: the shell's own roster (and the [*auth.Roster] the
// admin gate holds) is opened, by [buildServeShell], before the verb's
// first write, and the verb's own [addMember] call opens its roster
// after the API's write has already landed, so each side's answer can
// only be explained by a reload rather than by both having opened the
// file at the same instant.
func TestNewServeShell_RosterAndMemberVerbShareReload(t *testing.T) {
	environ, rosterPath := rosterAPIEnviron(t)
	shell, _, err := buildServeShell(t, environ) // the shell's roster, opened first
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	adminSession := mintServeSession(t, environ, rosterAdminAddress, time.Now())

	apiBody := "address=api-added%40example.com&display_name=API+Added&streamer_username=apiadded"
	apiReq := sessionRequest(http.MethodPost, prefix.String()+"/roster", adminSession, "same-origin", apiBody)
	apiRec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(apiRec, apiReq)
	if apiRec.Code != http.StatusCreated {
		t.Fatalf("API POST status = %d, body = %s, want %d", apiRec.Code, apiRec.Body.String(), http.StatusCreated)
	}

	// addMember opens its own *auth.Roster from rosterPath, strictly
	// after the API's write above.
	verbCandidate := auth.Member{Address: "verb-added@example.com", DisplayName: "Verb Added", StreamerUsername: "verbadded"}
	if err := addMember(t.Context(), rosterPath, verbCandidate, "operator", discardTestLogger()); err != nil {
		t.Fatalf("addMember: %v", err)
	}

	// "member list" reads what the API wrote: a fresh open, after both
	// writes, must see the API's own addition.
	listRoster, err := auth.OpenRoster(rosterPath)
	if err != nil {
		t.Fatalf("auth.OpenRoster: %v", err)
	}
	members, err := listRoster.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	var listOut bytes.Buffer
	if err := printMemberList(&listOut, members); err != nil {
		t.Fatalf("printMemberList: %v", err)
	}
	if !strings.Contains(listOut.String(), "api-added@example.com") {
		t.Errorf("member list output does not name the API's addition: %s", listOut.String())
	}

	// The API reads what the verb wrote: the shell's own roster, opened
	// before either write, must reload and see the verb's addition.
	getRec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(getRec, sessionRequest(http.MethodGet, prefix.String()+"/roster", adminSession, "", ""))
	if getRec.Code != http.StatusOK {
		t.Fatalf("API GET status = %d, body = %s, want %d", getRec.Code, getRec.Body.String(), http.StatusOK)
	}
	if !strings.Contains(getRec.Body.String(), "verb-added@example.com") {
		t.Errorf("API GET body does not name the verb's addition: %s", getRec.Body.String())
	}
}

// defaultIgnorableRoutingRows is the same ten code points
// internal/identity's TestNormalizeAddress_DefaultIgnorableRefused and
// internal/auth's TestOpenRoster_DefaultIgnorableAddressRefused pin,
// checked here at request-link, reached through the whole shell
// composition.
var defaultIgnorableRoutingRows = []struct {
	name string
	r    rune
}{
	{"Hangul filler U+115F", 'ᅟ'},
	{"Hangul jungseong filler U+1160", 'ᅠ'},
	{"Hangul filler U+3164", 'ㅤ'},
	{"halfwidth Hangul filler U+FFA0", 'ﾠ'},
	{"combining grapheme joiner U+034F", '͏'},
	{"variation selector 16 U+FE0F", '️'},
	{"zero width space U+200B", '​'},
	{"soft hyphen U+00AD", '­'},
	{"tag digit one U+E0001", '\U000E0001'},
	{"braille pattern blank U+2800", '⠀'},
}

// TestNewServeShell_RequestLinkRefusesDefaultIgnorableAddress proves
// the derived-set refusal at request-link: an address carrying one of
// the ten code points answers 400 invalid_address after
// allowLinkClient has already charged the client (handlers.go:194-208),
// reached through shell.Handler() rather than the handler called
// directly.
func TestNewServeShell_RequestLinkRefusesDefaultIgnorableAddress(t *testing.T) {
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

	plainBody := "address=" + url.QueryEscape(`"ab"@example.com`)
	plainRec := authRoutePOST(handler, prefix.String()+"/"+requestLinkRouteName, plainBody)
	if plainRec.Code != http.StatusOK {
		t.Fatalf("plain address status = %d, body = %s, want %d", plainRec.Code, plainRec.Body.String(), http.StatusOK)
	}
	var plainGot struct {
		Outcome string `json:"outcome"`
	}
	if err := json.Unmarshal(plainRec.Body.Bytes(), &plainGot); err != nil {
		t.Fatalf("json.Unmarshal(%s): %v", plainRec.Body.String(), err)
	}
	if plainGot.Outcome != "unknown" {
		t.Errorf("plain address outcome = %q, want %q", plainGot.Outcome, "unknown")
	}

	for _, tt := range defaultIgnorableRoutingRows {
		t.Run(tt.name, func(t *testing.T) {
			address := "\"a" + string(tt.r) + "b\"@example.com"
			body := "address=" + url.QueryEscape(address)
			rec := authRoutePOST(handler, prefix.String()+"/"+requestLinkRouteName, body)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, body = %s, want %d", rec.Code, rec.Body.String(), http.StatusBadRequest)
			}
			var got struct {
				Outcome string `json:"outcome"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("json.Unmarshal(%s): %v", rec.Body.String(), err)
			}
			if got.Outcome != "invalid_address" {
				t.Errorf("outcome = %q, want %q", got.Outcome, "invalid_address")
			}
		})
	}
}

func TestSendsHSTS(t *testing.T) {
	tests := []struct {
		publicURL string
		want      bool
	}{
		{publicURL: "https://dubplate.example.test", want: true},
		{publicURL: "https://dubplate.example.test/", want: true},
		{publicURL: "https://192.0.2.10", want: true},
		{publicURL: "http://localhost", want: false},
		{publicURL: "https://localhost", want: false},
		{publicURL: "https://localhost:8443", want: false},
		{publicURL: "https://LOCALHOST", want: false},
		{publicURL: "https://localhost.", want: false},
		{publicURL: "https://a.localhost", want: false},
		{publicURL: "https://localhost.example.test", want: true},
		{publicURL: "https://notlocalhost", want: true},
		{publicURL: "https://127.0.0.1", want: false},
		{publicURL: "https://[::1]", want: false},
		{publicURL: "http://127.0.0.1:8080", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.publicURL, func(t *testing.T) {
			if got := sendsHSTS(tt.publicURL); got != tt.want {
				t.Errorf("sendsHSTS(%q) = %v, want %v", tt.publicURL, got, tt.want)
			}
		})
	}
}

// firstImmutableAsset returns the path, relative to the build's _app
// directory, of one content-hashed file the real bundle carries.
func firstImmutableAsset(t *testing.T) string {
	t.Helper()
	var found string
	err := fs.WalkDir(web.Assets, "build/_app/immutable", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if found == "" && !d.IsDir() {
			found = strings.TrimPrefix(p, "build/_app/")
		}
		return nil
	})
	if err != nil || found == "" {
		t.Fatalf("no file under build/_app/immutable (err %v); run `npm run build` in web/", err)
	}
	return found
}

// TestNewServeShell_HeaderStepOnEveryResponse proves every response
// class the shell answers carries nosniff and Referrer-Policy:
// no-referrer, and carries HSTS exactly when the public URL is https on
// a host that is not loopback. It proves in the same rows that only a
// document carries a CSP or Vary: Cookie, and each class's
// Cache-Control: a 404 and a refusal that is not JSON set none, which
// is what keeps the gate's 404 equal to the terminator's.
//
// Fixture state: all three public URLs, so HSTS's presence and both
// absences are each proved, and an admin session, so the JSON row is
// the roster API's real answer.
func TestNewServeShell_HeaderStepOnEveryResponse(t *testing.T) {
	asset := firstImmutableAsset(t)
	type row struct {
		name     string
		method   string
		target   string
		admin    bool
		status   int
		cache    string
		document bool
	}
	rows := []row{
		{name: "document", method: http.MethodGet, target: "/dubplate/login", status: http.StatusOK, cache: "no-cache", document: true},
		{name: "asset", method: http.MethodGet, target: "/dubplate/_app/" + asset, status: http.StatusOK, cache: "public, max-age=31536000, immutable"},
		{name: "stylesheet", method: http.MethodGet, target: "/dubplate/branding.css", status: http.StatusOK, cache: "no-cache"},
		{name: "JSON answer", method: http.MethodGet, target: "/dubplate/roster", admin: true, status: http.StatusOK, cache: "no-store"},
		{name: "normalization 400", method: http.MethodGet, target: "/dubplate/login/../login", status: http.StatusBadRequest},
		{name: "origin 403", method: http.MethodPost, target: "/dubplate/request-link", status: http.StatusForbidden},
		{name: "echo method 405", method: http.MethodTrace, target: "/dubplate/login", status: http.StatusMethodNotAllowed},
		{name: "terminator 404", method: http.MethodGet, target: "/dubplate/no-such-subtree", status: http.StatusNotFound},
		{name: "gate 404", method: http.MethodGet, target: "/dubplate/members", status: http.StatusNotFound},
		{name: "outside the prefix 404", method: http.MethodGet, target: "/elsewhere", status: http.StatusNotFound},
	}
	configs := []struct {
		publicURL string
		wantHSTS  string
	}{
		{publicURL: "https://dubplate.example.test", wantHSTS: "max-age=31536000"},
		{publicURL: "http://localhost"},
		{publicURL: "https://localhost"},
	}
	for _, c := range configs {
		environ, _ := rosterAPIEnviron(t)
		environ = replaceKey(environ, "DUBPLATE_PUBLIC_URL", c.publicURL)
		shell, _, err := buildServeShellWithLogger(t, environ, discardTestLogger())
		if err != nil {
			t.Fatalf("%s: buildServeShell: %v", c.publicURL, err)
		}
		handler := shell.Handler()
		adminSession := mintServeSession(t, environ, rosterAdminAddress, time.Now())

		for _, rr := range rows {
			t.Run(c.publicURL+"/"+rr.name, func(t *testing.T) {
				session := ""
				if rr.admin {
					session = adminSession
				}
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, sessionRequest(rr.method, rr.target, session, "", ""))
				if rec.Code != rr.status {
					t.Fatalf("status = %d, want %d", rec.Code, rr.status)
				}
				h := rec.Header()
				if got := h.Get("X-Content-Type-Options"); got != "nosniff" {
					t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
				}
				if got := h.Get("Referrer-Policy"); got != "no-referrer" {
					t.Errorf("Referrer-Policy = %q, want no-referrer", got)
				}
				if got := h.Get("Strict-Transport-Security"); got != c.wantHSTS {
					t.Errorf("Strict-Transport-Security = %q, want %q", got, c.wantHSTS)
				}
				if got := h.Get("Cache-Control"); got != rr.cache {
					t.Errorf("Cache-Control = %q, want %q", got, rr.cache)
				}
				hasCSP := h.Get("Content-Security-Policy") != ""
				if hasCSP != rr.document {
					t.Errorf("Content-Security-Policy present = %v, want %v", hasCSP, rr.document)
				}
				wantVary := []string(nil)
				if rr.document {
					wantVary = []string{"Cookie"}
				}
				if got := h.Values("Vary"); !slices.Equal(got, wantVary) {
					t.Errorf("Vary = %q, want %q", got, wantVary)
				}
			})
		}
	}
}

// servedDocuments maps each document the build must emit to the path
// under the default prefix that serves it.
var servedDocuments = map[string]string{
	"login.html":      "/dubplate/login",
	"login/link.html": "/dubplate/login/link",
	"members.html":    "/dubplate/members",
}

// checkDocumentSet reports a failure unless buildFS's documents are
// exactly the three the routes serve. It runs first in each document
// check, so a check cannot pass over an empty set.
func checkDocumentSet(buildFS fs.FS) string {
	var got []string
	err := fs.WalkDir(buildFS, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".html") {
			got = append(got, p)
		}
		return nil
	})
	if err != nil {
		return "walk the build: " + err.Error()
	}
	slices.Sort(got)
	want := slices.Sorted(maps.Keys(servedDocuments))
	if !slices.Equal(got, want) {
		return fmt.Sprintf("the build's documents are %v, want %v; a tree with none holds only the placeholder, so run `npm run build` in web/", got, want)
	}
	return ""
}

// webFSHandler builds a shell over the routes newWebRoutesFromFS builds
// from buildFS, with the members document mounted with no gate, since
// these checks read the document rather than the gate.
func webFSHandler(environ []string, buildFS fs.FS) (http.Handler, error) {
	prefix, err := serve.NewPrefix("/dubplate")
	if err != nil {
		return nil, err
	}
	routes, err := newWebRoutesFromFS(environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), buildFS)
	if err != nil {
		return nil, err
	}
	shell := serve.NewShell(prefix, serve.NewStrip(serve.TrustedProxyConfig{}), discardTestLogger())
	routes.register(shell)
	shell.Register(membersRouteName, routes.members)
	return shell.Handler(), nil
}

// scanInlineScripts returns the body of every script element in doc,
// attributed or not. It shares nothing with the server's own
// extractor: it searches an ASCII-lowercased copy of doc for the open
// tag, the end of that tag, and the close tag by plain byte search, so
// one blind spot cannot hide a script from both.
func scanInlineScripts(doc []byte) [][]byte {
	lower := make([]byte, len(doc))
	for i, b := range doc {
		if 'A' <= b && b <= 'Z' {
			b += 'a' - 'A'
		}
		lower[i] = b
	}
	var bodies [][]byte
	pos := 0
	for {
		open := bytes.Index(lower[pos:], []byte("<script"))
		if open < 0 {
			return bodies
		}
		tagEnd := bytes.IndexByte(lower[pos+open:], '>')
		if tagEnd < 0 {
			return bodies
		}
		start := pos + open + tagEnd + 1
		end := bytes.Index(lower[start:], []byte("</script"))
		if end < 0 {
			return append(bodies, doc[start:])
		}
		bodies = append(bodies, doc[start:start+end])
		pos = start + end
	}
}

// cspDirectives parses a Content-Security-Policy value into its
// directives' sources, keyed by directive name.
func cspDirectives(policy string) map[string][]string {
	directives := make(map[string][]string)
	for part := range strings.SplitSeq(policy, ";") {
		fields := strings.Fields(part)
		if len(fields) > 0 {
			directives[fields[0]] = fields[1:]
		}
	}
	return directives
}

// cspMetaContentRE captures the content attribute of a served
// document's CSP meta, as the build writes it.
var cspMetaContentRE = regexp.MustCompile(`<meta http-equiv="content-security-policy" content="([^"]*)">`)

// checkServedCSP runs the design language's G6 over the documents
// served from buildFS and returns every failure. For each document the
// header must carry the directive set by name and value, and script-src
// must carry a sha256 source for every inline script body, hashed
// here, with no 'unsafe-inline'. The document's own CSP meta, which the
// kit config writes, must carry the header's directives other than
// script-src by name and value, frame-ancestors aside because a meta
// policy cannot carry it, so the two policies cannot drift apart.
func checkServedCSP(buildFS fs.FS) []string {
	if failure := checkDocumentSet(buildFS); failure != "" {
		return []string{failure}
	}
	handler, err := webFSHandler(nil, buildFS)
	if err != nil {
		return []string{"build the routes: " + err.Error()}
	}
	wantFixed := map[string][]string{
		"default-src":     {"'self'"},
		"style-src":       {"'self'", "'unsafe-inline'"},
		"img-src":         {"'self'", "data:"},
		"frame-ancestors": {"'none'"},
		"base-uri":        {"'none'"},
		"object-src":      {"'none'"},
		"form-action":     {"'self'"},
	}
	var failures []string
	for _, name := range slices.Sorted(maps.Keys(servedDocuments)) {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, servedDocuments[name], nil))
		if rec.Code != http.StatusOK {
			failures = append(failures, fmt.Sprintf("%s: status %d, want 200", name, rec.Code))
			continue
		}
		directives := cspDirectives(rec.Header().Get("Content-Security-Policy"))
		for directive, want := range wantFixed {
			if got := directives[directive]; !slices.Equal(got, want) {
				failures = append(failures, fmt.Sprintf("%s: %s = %q, want %q", name, directive, got, want))
			}
		}
		if _, ok := directives["font-src"]; ok {
			failures = append(failures, name+": carries font-src, want none")
		}
		if failure := checkMetaMatchesHeader(rec.Body.Bytes(), directives); failure != "" {
			failures = append(failures, name+": "+failure)
		}
		scriptSrc := directives["script-src"]
		if len(scriptSrc) == 0 || scriptSrc[0] != "'self'" {
			failures = append(failures, fmt.Sprintf("%s: script-src = %q, want 'self' first", name, scriptSrc))
		}
		if slices.Contains(scriptSrc, "'unsafe-inline'") {
			failures = append(failures, name+": script-src carries 'unsafe-inline'")
		}
		bodies := scanInlineScripts(rec.Body.Bytes())
		if len(bodies) == 0 {
			failures = append(failures, name+": no inline script, want the build's bootstrap")
		}
		for _, body := range bodies {
			sum := sha256.Sum256(body)
			source := "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
			if !slices.Contains(scriptSrc, source) {
				failures = append(failures, fmt.Sprintf("%s: inline script %s is not in script-src %q", name, source, scriptSrc))
			}
		}
	}
	return failures
}

// checkMetaMatchesHeader reports a failure unless the CSP meta in doc
// carries exactly header's directives, script-src and frame-ancestors
// excepted, with the same sources in the same order.
func checkMetaMatchesHeader(doc []byte, header map[string][]string) string {
	m := cspMetaContentRE.FindSubmatch(doc)
	if m == nil {
		return "no CSP meta"
	}
	meta := cspDirectives(html.UnescapeString(string(m[1])))
	delete(meta, "script-src")
	want := maps.Clone(header)
	delete(want, "script-src")
	delete(want, "frame-ancestors")
	if !maps.EqualFunc(meta, want, slices.Equal) {
		return fmt.Sprintf("CSP meta directives %q, want the header's %q", meta, want)
	}
	return ""
}

// placeholderOnlyBuild is the build tree a checkout that has never run
// the web build embeds.
func placeholderOnlyBuild() fstest.MapFS {
	return fstest.MapFS{"placeholder.txt": &fstest.MapFile{Data: []byte("placeholder")}}
}

func TestServedDocuments_CSPCoversEveryInlineScript(t *testing.T) {
	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	for _, failure := range checkServedCSP(buildFS) {
		t.Error(failure)
	}
}

func TestServedDocuments_ChecksFailOnThePlaceholderOnlyBuild(t *testing.T) {
	checks := map[string]func(fs.FS) []string{
		"CSP":             checkServedCSP,
		"tiers and cache": func(buildFS fs.FS) []string { return checkServedTiersAndCache(t, buildFS) },
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) {
			failures := check(placeholderOnlyBuild())
			if len(failures) == 0 || !strings.Contains(failures[0], "npm run build") {
				t.Errorf("failures = %q, want the first to name the missing build", failures)
			}
		})
	}
}

// stylesheetLinkRE matches a stylesheet link as the build and the tier
// rewrite both write it, href before rel.
var stylesheetLinkRE = regexp.MustCompile(`<link href="([^"]*)" rel="stylesheet">`)

// checkServedTiersAndCache runs the design language's G7, its link
// order and cache half, over the documents served from buildFS with an
// operator file configured, and returns every failure: each document
// links the bundle's stylesheets, then branding.css, then
// operator.css; each carries its Cache-Control and Vary: Cookie; and a
// hashed asset, the version file, both tier stylesheets, and a 404
// each carry their own Cache-Control and no Vary.
func checkServedTiersAndCache(t *testing.T, buildFS fs.FS) []string {
	t.Helper()
	if failure := checkDocumentSet(buildFS); failure != "" {
		return []string{failure}
	}
	environ := []string{"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, []byte("body{}"))}
	handler, err := webFSHandler(environ, buildFS)
	if err != nil {
		return []string{"build the routes: " + err.Error()}
	}
	get := func(target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))
		return rec
	}

	wantCache := map[string]string{
		"login.html":      "no-cache",
		"login/link.html": "no-store",
		"members.html":    "no-store",
	}
	var failures []string
	var bundleStylesheet string
	for _, name := range slices.Sorted(maps.Keys(servedDocuments)) {
		target := servedDocuments[name]
		rec := get(target)
		if rec.Code != http.StatusOK {
			failures = append(failures, fmt.Sprintf("%s: status %d, want 200", name, rec.Code))
			continue
		}
		var hrefs []string
		for _, m := range stylesheetLinkRE.FindAllSubmatch(rec.Body.Bytes(), -1) {
			hrefs = append(hrefs, string(m[1]))
		}
		root := "./"
		if strings.Contains(name, "/") {
			root = "../"
		}
		wantTier := []string{root + "branding.css", root + "operator.css"}
		if len(hrefs) <= len(wantTier) || !slices.Equal(hrefs[len(hrefs)-len(wantTier):], wantTier) {
			failures = append(failures, fmt.Sprintf("%s: stylesheet links %q, want the bundle's then %q", name, hrefs, wantTier))
		} else {
			for _, href := range hrefs[:len(hrefs)-len(wantTier)] {
				if !strings.HasPrefix(href, root+"_app/immutable/") {
					failures = append(failures, fmt.Sprintf("%s: stylesheet %q precedes the tier links but is not the bundle's", name, href))
				}
			}
			if name == "login.html" {
				bundleStylesheet = strings.TrimPrefix(hrefs[0], root)
			}
		}
		if got := rec.Header().Get("Cache-Control"); got != wantCache[name] {
			failures = append(failures, fmt.Sprintf("%s: Cache-Control %q, want %q", name, got, wantCache[name]))
		}
		if got := rec.Header().Values("Vary"); !slices.Equal(got, []string{"Cookie"}) {
			failures = append(failures, fmt.Sprintf("%s: Vary %q, want Cookie", name, got))
		}
	}
	if bundleStylesheet == "" {
		return append(failures, "login.html: no bundle stylesheet to read an immutable asset through")
	}

	others := []struct {
		target, cache string
		status        int
	}{
		{target: "/dubplate/" + bundleStylesheet, cache: "public, max-age=31536000, immutable", status: http.StatusOK},
		{target: "/dubplate/_app/version.json", cache: "no-cache", status: http.StatusOK},
		{target: "/dubplate/branding.css", cache: "no-cache", status: http.StatusOK},
		{target: "/dubplate/operator.css", cache: "no-cache", status: http.StatusOK},
		{target: "/dubplate/no-such-subtree", status: http.StatusNotFound},
	}
	for _, o := range others {
		rec := get(o.target)
		if rec.Code != o.status {
			failures = append(failures, fmt.Sprintf("%s: status %d, want %d", o.target, rec.Code, o.status))
		}
		if got := rec.Header().Get("Cache-Control"); got != o.cache {
			failures = append(failures, fmt.Sprintf("%s: Cache-Control %q, want %q", o.target, got, o.cache))
		}
		if got := rec.Header().Values("Vary"); len(got) != 0 {
			failures = append(failures, fmt.Sprintf("%s: Vary %q, want none", o.target, got))
		}
	}
	return failures
}

func TestServedDocuments_TierOrderAndCache(t *testing.T) {
	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	for _, failure := range checkServedTiersAndCache(t, buildFS) {
		t.Error(failure)
	}
}

// TestNewServeShell_MembersRoute proves the members document is gated
// ahead of its own method and path checks. A verified non-admin's GET
// and POST answer exactly what an unregistered owned path answers,
// with no Set-Cookie. A verified admin's GET answers the stamped
// document, an admin's POST answers 405, and an admin's request to a
// deeper path answers the terminator's 404.
//
// Fixture state: the POSTs carry Sec-Fetch-Site: same-origin, so each
// passes the origin check and reaches the gate, the terminator, or the
// method check.
func TestNewServeShell_MembersRoute(t *testing.T) {
	environ, _ := rosterAPIEnviron(t)
	shell, _, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err != nil {
		t.Fatalf("buildServeShell: %v", err)
	}
	handler := shell.Handler()
	nonAdmin := mintServeSession(t, environ, rosterMemberAddress, time.Now())
	admin := mintServeSession(t, environ, rosterAdminAddress, time.Now())
	run := func(method, target, session, sfs string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, sessionRequest(method, target, session, sfs, ""))
		return rec
	}

	for _, method := range []string{http.MethodGet, http.MethodPost} {
		t.Run("non-admin "+method, func(t *testing.T) {
			members := run(method, "/dubplate/members", nonAdmin, "same-origin")
			unknown := run(method, "/dubplate/no-such-subtree", nonAdmin, "same-origin")
			if members.Code != unknown.Code || members.Code != http.StatusNotFound {
				t.Errorf("status: members = %d, unknown = %d, want both 404", members.Code, unknown.Code)
			}
			if members.Body.String() != unknown.Body.String() {
				t.Errorf("body: members = %q, unknown = %q, want equal", members.Body.String(), unknown.Body.String())
			}
			if !headersEqualButDate(members.Header(), unknown.Header()) {
				t.Errorf("headers differ (Date excepted): members = %v, unknown = %v", members.Header(), unknown.Header())
			}
			if got := members.Header().Values("Set-Cookie"); len(got) != 0 {
				t.Errorf("Set-Cookie = %q, want none", got)
			}
		})
	}

	t.Run("admin GET", func(t *testing.T) {
		rec := run(http.MethodGet, "/dubplate/members", admin, "")
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `data-theme="dubplate-system"`) {
			t.Errorf("body does not carry the stamped theme: %s", rec.Body.String())
		}
		if got := rec.Header().Get("Cache-Control"); got != "no-store" {
			t.Errorf("Cache-Control = %q, want %q", got, "no-store")
		}
		if rec.Header().Get("Content-Security-Policy") == "" {
			t.Error("the members document carries no CSP")
		}
	})

	t.Run("admin POST", func(t *testing.T) {
		rec := run(http.MethodPost, "/dubplate/members", admin, "same-origin")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Fatalf("status = %d, want 405", rec.Code)
		}
		if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
			t.Errorf("Allow = %q, want %q", got, "GET, HEAD")
		}
	})

	for _, target := range []string{"/dubplate/members/", "/dubplate/members/x"} {
		t.Run("admin GET "+target, func(t *testing.T) {
			rec := run(http.MethodGet, target, admin, "")
			unknown := run(http.MethodGet, "/dubplate/no-such-subtree", admin, "")
			if rec.Code != http.StatusNotFound || rec.Body.String() != unknown.Body.String() {
				t.Errorf("status %d body %q, want the terminator's 404 %q", rec.Code, rec.Body.String(), unknown.Body.String())
			}
		})
	}
}

// TestNewServeShell_MintedLinkNeverLogsItsToken proves a link built
// through auth.LinkURL answers the login/link document with no-store,
// which binds the link's path to the document route, and that across
// that request, a trailing-slash GET, a POST, and a request
// normalization refuses, each carrying the token in its query, no
// record the shell's logger receives contains the token.
//
// Fixture state: every request leaves at least one record in the
// captured log, so "no line contains the token" reads real lines.
func TestNewServeShell_MintedLinkNeverLogsItsToken(t *testing.T) {
	const token = "qz81-minted-fixture-token"
	environ := serveAuthEnviron(t)
	var logs bytes.Buffer
	shell, _, err := buildServeShellWithLogger(t, environ, slog.New(slog.NewJSONHandler(&logs, nil)))
	if err != nil {
		t.Fatalf("buildServeShell: %v", err)
	}
	handler := shell.Handler()

	link, err := auth.LinkURL(environValue(environ, "DUBPLATE_PUBLIC_URL"), "/dubplate", token)
	if err != nil {
		t.Fatalf("auth.LinkURL: %v", err)
	}
	u, err := url.Parse(link)
	if err != nil {
		t.Fatalf("parse minted link: %v", err)
	}
	if u.Query().Get("t") != token {
		t.Fatalf("minted link %q does not carry the token as t", link)
	}

	plain := httptest.NewRecorder()
	handler.ServeHTTP(plain, httptest.NewRequest(http.MethodGet, "/dubplate/login/link", nil))
	login := httptest.NewRecorder()
	handler.ServeHTTP(login, httptest.NewRequest(http.MethodGet, "/dubplate/login", nil))

	query := "?" + u.RawQuery
	tests := []struct {
		name   string
		method string
		target string
		status int
	}{
		{name: "the minted link", method: http.MethodGet, target: u.Path + query, status: http.StatusOK},
		{name: "trailing slash", method: http.MethodGet, target: "/dubplate/login/link/" + query, status: http.StatusNotFound},
		{name: "POST", method: http.MethodPost, target: "/dubplate/login/link" + query, status: http.StatusMethodNotAllowed},
		{name: "refused by normalization", method: http.MethodGet, target: "/dubplate/login/./link" + query, status: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logs.Reset()
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, sessionRequest(tt.method, tt.target, "", "same-origin", ""))
			if rec.Code != tt.status {
				t.Fatalf("status = %d, want %d", rec.Code, tt.status)
			}
			if tt.status == http.StatusOK {
				if got := rec.Header().Get("Cache-Control"); got != "no-store" {
					t.Errorf("Cache-Control = %q, want no-store", got)
				}
				if rec.Body.String() != plain.Body.String() || rec.Body.String() == login.Body.String() {
					t.Error("the minted link did not answer the login/link document")
				}
			}
			lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
			if logs.Len() == 0 {
				t.Fatal("the request left no log record, so the token check would read nothing")
			}
			for _, line := range lines {
				if strings.Contains(line, token) {
					t.Errorf("log line carries the token: %s", line)
				}
			}
		})
	}
}

// trustedProxyEnviron returns serveAuthEnviron with the trusted proxy
// class enabled for addrs and Cf-Connecting-Ip, and the per-client
// link-request limit at 1, so a client's second request trips it.
func trustedProxyEnviron(t *testing.T, addrs string) []string {
	t.Helper()
	environ := serveAuthEnviron(t)
	environ = replaceKey(environ, "DUBPLATE_TRUSTED_PROXY_ENABLED", "true")
	environ = replaceKey(environ, "DUBPLATE_TRUSTED_PROXY_ADDRS", addrs)
	environ = replaceKey(environ, "DUBPLATE_TRUSTED_PROXY_HEADER", "Cf-Connecting-Ip")
	return replaceKey(environ, "DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR", "1")
}

// TestNewServeShell_TrustedProxyKeysTheLinkLimiter proves the shell's
// construction reads the trusted proxy class: a request-link POST from
// the configured proxy is charged to the client its header names, and
// the same POST from any other peer is charged to that peer.
//
// Each POST carries an address that does not parse. The handler
// charges the per-client bucket before it reads the form, then answers
// 400 for the address, so a request either spends its client's one
// request (400) or meets a spent bucket (429), and no mail is ever
// attempted. Origin and Sec-Fetch-Site match the fixture's public URL,
// so the origin check admits every POST.
func TestNewServeShell_TrustedProxyKeysTheLinkLimiter(t *testing.T) {
	type post struct {
		peer   string // RemoteAddr
		client string // the Cf-Connecting-Ip value
		want   int
	}
	tests := []struct {
		name  string
		addrs string
		posts []post
	}{
		{
			name:  "configured proxy: each client behind it has its own bucket",
			addrs: "192.0.2.10",
			posts: []post{
				{"192.0.2.10:40001", "203.0.113.9", http.StatusBadRequest},
				{"192.0.2.10:40002", "203.0.113.9", http.StatusTooManyRequests},
				{"192.0.2.10:40003", "203.0.113.10", http.StatusBadRequest},
			},
		},
		{
			// The proxy is configured in IPv4-mapped form and dials
			// from its dotted address: the set is compared unmapped.
			name:  "configured in mapped form: the dotted peer is still the proxy",
			addrs: "::ffff:198.51.100.7",
			posts: []post{
				{"198.51.100.7:40001", "203.0.113.9", http.StatusBadRequest},
				{"198.51.100.7:40002", "203.0.113.9", http.StatusTooManyRequests},
				{"198.51.100.7:40003", "203.0.113.10", http.StatusBadRequest},
			},
		},
		{
			// The header names two different clients, but the peer is
			// not the proxy, so both POSTs spend the peer's one bucket.
			name:  "unconfigured peer: every request keys on the peer",
			addrs: "192.0.2.10",
			posts: []post{
				{"198.51.100.20:40001", "203.0.113.9", http.StatusBadRequest},
				{"198.51.100.20:40002", "203.0.113.10", http.StatusTooManyRequests},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := trustedProxyEnviron(t, tt.addrs)
			shell, _, err := buildServeShellWithLogger(t, environ, discardTestLogger())
			if err != nil {
				t.Fatalf("buildServeShell() error = %v, want nil", err)
			}
			prefix, _, _, err := resolveServeConfig(environ)
			if err != nil {
				t.Fatalf("resolveServeConfig() error = %v, want nil", err)
			}
			publicURL := environValue(environ, "DUBPLATE_PUBLIC_URL")
			handler := shell.Handler()

			for i, p := range tt.posts {
				req := httptest.NewRequest(http.MethodPost, prefix.String()+"/"+requestLinkRouteName, strings.NewReader("address=not-an-address"))
				req.RemoteAddr = p.peer
				req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
				req.Header.Set("Origin", publicURL)
				req.Header.Set("Sec-Fetch-Site", "same-origin")
				req.Header.Set("Cf-Connecting-Ip", p.client)
				rec := httptest.NewRecorder()
				handler.ServeHTTP(rec, req)
				if rec.Code != p.want {
					t.Errorf("POST %d from %s naming %s: status = %d, body = %s, want %d", i+1, p.peer, p.client, rec.Code, rec.Body.String(), p.want)
				}
			}
		})
	}
}

// TestNewServeShell_StripsIdentityFromTheTrustedProxy proves the shell
// strips every identity header even from the configured trusted proxy:
// the proxy is trusted for the client address alone. The spy also
// records the configured address header, which the strip keeps from a
// trusted peer, so the request is known to have arrived as the
// proxy's.
func TestNewServeShell_StripsIdentityFromTheTrustedProxy(t *testing.T) {
	environ := trustedProxyEnviron(t, "192.0.2.10")
	shell, _, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}

	var reached bool
	var got http.Header
	shell.Register("spy", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reached = true
		got = r.Header.Clone()
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, prefix.String()+"/spy", nil)
	req.RemoteAddr = "192.0.2.10:40001"
	req.Header.Set("Remote-User", "admin@example.com")
	req.Header.Set("Cf-Connecting-Ip", "203.0.113.9")
	shell.Handler().ServeHTTP(httptest.NewRecorder(), req)

	if !reached {
		t.Fatal("the spy route was never reached, so the request proves nothing about the strip")
	}
	if got.Get("Cf-Connecting-Ip") != "203.0.113.9" {
		t.Errorf("Cf-Connecting-Ip = %q, want 203.0.113.9: the peer was not treated as the configured proxy", got.Get("Cf-Connecting-Ip"))
	}
	if v := got.Get("Remote-User"); v != "" {
		t.Errorf("Remote-User = %q reached the spy from the trusted proxy, want stripped", v)
	}
}

// TestNewServeShell_RefusesABadTrustedProxyValue proves a value the
// resolver refuses stops the shell's construction, which runs before
// serve opens any listener, rather than falling back to direct mode.
func TestNewServeShell_RefusesABadTrustedProxyValue(t *testing.T) {
	environ := trustedProxyEnviron(t, "192.0.2.0/24")
	shell, _, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err == nil {
		t.Fatal("buildServeShell() error = nil, want the network entry refused")
	}
	if !strings.Contains(err.Error(), `DUBPLATE_TRUSTED_PROXY_ADDRS "192.0.2.0/24": an address, not a network`) {
		t.Errorf("buildServeShell() error = %v, want it to name the network entry", err)
	}
	if shell != nil {
		t.Error("buildServeShell() returned a shell alongside its refusal, want nil")
	}
}

// TestNewServeShell_UnwritableRosterDirRefusesAtStart proves serve
// checks the roster file's directory at start: a roster that opens for
// reading in a directory the process may not write would otherwise
// start clean and fail the first web member add. The refusal names the
// directory and carries the kernel's own reason.
func TestNewServeShell_UnwritableRosterDirRefusesAtStart(t *testing.T) {
	environ := serveAuthEnviron(t)
	rosterDir := filepath.Dir(environValue(environ, "DUBPLATE_ROSTER_FILE"))
	if err := os.Chmod(rosterDir, 0o500); err != nil {
		t.Fatalf("Chmod roster directory: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(rosterDir, 0o700) })

	_, _, err := buildServeShell(t, environ)
	if !errors.Is(err, errDirNotWritable) {
		t.Fatalf("buildServeShell() error = %v, want it to wrap errDirNotWritable", err)
	}
	if !errors.Is(err, syscall.EACCES) {
		t.Errorf("buildServeShell() error = %v, want it to carry EACCES", err)
	}
	if want := "roster directory " + rosterDir + " is not writable"; !strings.Contains(err.Error(), want) {
		t.Errorf("buildServeShell() error = %q, want it to contain %q", err, want)
	}
}

// TestServePortBound proves the serve port's bound reaches every verb
// through [config.Load]: 0 would listen on an ephemeral port, and a
// port below 1024 cannot bind under the serve unit's empty capability
// set, so both refuse naming the capability, while 1024 and 65535, the
// bound's own ends, load.
func TestServePortBound(t *testing.T) {
	tests := []struct {
		port    string
		wantErr string // a substring of the refusal; empty means Load succeeds
	}{
		{port: "0", wantErr: "DUBPLATE_SERVE_PORT: 0 is below 1024"},
		{port: "80", wantErr: "DUBPLATE_SERVE_PORT: 80 is below 1024"},
		{port: "1023", wantErr: "DUBPLATE_SERVE_PORT: 1023 is below 1024"},
		{port: "1024"},
		{port: "65535"},
		{port: "65536", wantErr: "DUBPLATE_SERVE_PORT: 65536 is above 65535"},
	}
	for _, tt := range tests {
		t.Run(tt.port, func(t *testing.T) {
			environ := replaceKey(authEnviron(t), "DUBPLATE_SERVE_PORT", tt.port)
			_, err := config.Load(environ, productionRegistry())
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("config.Load() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("config.Load() error = nil, want a refusal containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("config.Load() error = %q, want it to contain %q", err, tt.wantErr)
			}
			if strings.Contains(tt.wantErr, "below") && !strings.Contains(err.Error(), "CAP_NET_BIND_SERVICE") {
				t.Errorf("config.Load() error = %q, want it to name CAP_NET_BIND_SERVICE", err)
			}
		})
	}
}
