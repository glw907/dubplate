package main

import (
	"bytes"
	"cmp"
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
	"testing"
	"testing/fstest"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/settings"
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
func buildServeShell(t *testing.T, environ []string) (*serve.Shell, error) {
	t.Helper()
	return buildServeShellWithLogger(t, environ, slog.Default())
}

// buildServeShellWithLogger is [buildServeShell] with the shell's
// logger supplied, so a test reads every record the shell and its
// handlers write without swapping slog.Default.
func buildServeShellWithLogger(t *testing.T, environ []string, logger *slog.Logger) (*serve.Shell, error) {
	t.Helper()
	return buildServeShellContext(t, t.Context(), environ, logger)
}

// buildServeShellContext is [buildServeShellWithLogger] with the
// context [newServeShell] waits on supplied. The store handle a
// successful build returns is closed at cleanup.
func buildServeShellContext(t *testing.T, ctx context.Context, environ []string, logger *slog.Logger) (*serve.Shell, error) {
	t.Helper()
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	prefix, addrs, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig: %v", err)
	}
	shell, h, err := newServeShell(ctx, environ, productionRegistry(), prefix, addrs, cfg, logger)
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() {
		if err := h.Close(); err != nil {
			t.Errorf("close store handle: %v", err)
		}
	})
	return shell, nil
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
// inside newRootCmd's AddCommand block, beside newDoctorCmd's: a verb
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
	var serveCmd, doctorCmd bool
	for _, child := range newRootCmd().Commands() {
		switch child.Use {
		case "serve":
			serveCmd = true
		case "doctor":
			doctorCmd = true
		}
	}
	if !serveCmd {
		t.Error("newRootCmd().Commands() carries no serve verb, want it added beside doctor")
	}
	if !doctorCmd {
		t.Fatal("newRootCmd().Commands() carries no doctor verb; the AddCommand block this test reads against has moved")
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
	shell, err := buildServeShell(t, environ)
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

// TestNewServeShell_OperatorCSSRefusesPrivateState proves the operator
// CSS rule on the serve verb's own construction path: the key pointed
// at the store's database file, or at a ".css" file in the state
// directory, refuses startup, naming the key and the rule, and neither
// the path nor a byte of the file reaches the error. The second row
// fails unless the verb hands the web routes its own state directory. A
// shell that admitted either would build and publish the file at
// <prefix>/operator.css.
func TestNewServeShell_OperatorCSSRefusesPrivateState(t *testing.T) {
	tests := []struct {
		name    string
		path    func(t *testing.T, stateDir string) string
		wantErr string
	}{
		{
			name:    "the store's database file",
			path:    func(_ *testing.T, stateDir string) string { return storeDBPath(stateDir) },
			wantErr: `DUBPLATE_OPERATOR_CSS_FILE: the path must end in ".css"`,
		},
		{
			name: "a .css file in the state directory",
			path: func(t *testing.T, stateDir string) string {
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
			path := tt.path(t, valueOf(t, environ, "DUBPLATE_STATE_DIR"))
			environ = replaceKey(environ, "DUBPLATE_OPERATOR_CSS_FILE", path)

			shell, err := buildServeShell(t, environ)
			if err == nil {
				rec := httptest.NewRecorder()
				shell.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dubplate/operator.css", nil))
				t.Fatalf("buildServeShell: got nil error, want the operator CSS refusal; GET operator.css answers %d", rec.Code)
			}
			if err.Error() != tt.wantErr {
				t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
			}
			if strings.Contains(err.Error(), path) {
				t.Errorf("error = %q echoes the path", err.Error())
			}
		})
	}
}

// TestNewServeShell_BrandingRefusesStartup proves a refused branding
// value stops the serve verb's own construction, naming the key.
func TestNewServeShell_BrandingRefusesStartup(t *testing.T) {
	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_BRANDING", "--color-base-100: #ffffff; --color-primary: #fff}")
	_, err := buildServeShell(t, environ)
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

// TestNewServeShell_StateChangingRequestIsDenied proves no bypass
// pattern is registered: a cookie-carrying POST carrying neither
// Sec-Fetch-Site nor Origin is refused with 403 by the origin check, a
// status neither the root 404 nor any subtree's own answer produces for
// these paths. The rows are a path outside the owned prefix and one
// path under every subtree [newServeShell] registers, so a bypass
// registered for any of them fails its row.
func TestNewServeShell_StateChangingRequestIsDenied(t *testing.T) {
	under := func(name string) string { return "/dubplate/" + name }
	tests := []struct {
		name string
		path string
	}{
		{"outside the owned prefix", "/outside"},
		{"request-link", under(requestLinkRouteName)},
		{"consume-link", under(consumeLinkRouteName)},
		{"sign-out", under(signOutRouteName)},
		{"boundary probe", under(serve.ProbeRouteName)},
		{"roster API", under(rosterRouteName)},
		{"roster API member", under(rosterRouteName) + "/a-member"},
		{"members", under(membersRouteName)},
		{"login", under(loginRouteName)},
		{"assets", under(assetsRouteName) + "/app.js"},
		{"fonts", under(fontsRouteName) + "/f.woff2"},
		{"favicon.svg", under(faviconSVGRouteName)},
		{"favicon.png", under(faviconPNGRouteName)},
		{"mark", under(markRouteName)},
		{"branding.css", under(brandingRouteName)},
		{"operator.css", under(operatorCSSRouteName)},
	}
	shell, err := buildServeShell(t, serveAuthEnviron(t))
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			req.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: "a-session-value"})
			rec := httptest.NewRecorder()
			shell.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusForbidden {
				t.Errorf("POST %s: status = %d, want %d", tt.path, rec.Code, http.StatusForbidden)
			}
		})
	}
}

// TestSweepNonces_StopsWithContext is the sweep's leak assertion:
// canceling ctx makes [sweepNonces] return, proved by the channel its
// caller closes rather than by inspecting its source.
func TestSweepNonces_StopsWithContext(t *testing.T) {
	stateDir := t.TempDir()
	store, err := auth.Open(stateDir)
	if err != nil {
		t.Fatalf("auth.Open: %v", err)
	}
	nonces, _, err := auth.OpenNonceStore(t.Context(), store, func(context.Context) (time.Duration, error) { return time.Hour, nil }, magicLinkExpiryCeiling, time.Now())
	if err != nil {
		t.Fatalf("auth.OpenNonceStore: %v", err)
	}

	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan struct{})
	go func() {
		defer close(done)
		sweepNonces(ctx, nonces, discardTestLogger())
	}()

	select {
	case <-done:
		t.Fatal("sweepNonces returned before ctx was canceled")
	case <-time.After(50 * time.Millisecond):
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("sweepNonces kept running after ctx was canceled")
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
// Fixture state: the assertion is a 200, a status the shell's 404
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
			shell, h, err := newServeShell(t.Context(), environ, productionRegistry(), prefix, addrs, cfg, discardTestLogger())
			if err != nil {
				t.Fatalf("newServeShell: %v", err)
			}
			t.Cleanup(func() { _ = h.Close() })
			handlersConfig := authHandlersConfig(prefix, cfg, settings.NewAccessor(h, productionRegistry()), nil, nil, nil, nil, nil, nil, nil)
			if handlersConfig.Prefix != prefix.String() {
				t.Errorf("authHandlersConfig Prefix = %q, want %q", handlersConfig.Prefix, prefix.String())
			}
			if handlersConfig.BodyCapCeiling != 65536 {
				t.Errorf("authHandlersConfig BodyCapCeiling = %d, want the registered ceiling 65536", handlersConfig.BodyCapCeiling)
			}

			link, err := auth.LinkURL(handlersConfig.PublicOrigin, handlersConfig.Prefix, "fixture-token")
			if err != nil {
				t.Fatalf("auth.LinkURL: %v", err)
			}
			u, err := url.Parse(link)
			if err != nil {
				t.Fatalf("parse minted link: %v", err)
			}
			if !strings.HasPrefix(u.Path, prefix.String()+"/") {
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

// seedMember is one users row [seedMembers] writes.
type seedMember struct {
	address, displayName, username string
	admin, tombstoned              bool
}

// seedMembers writes members into the store under stateDir through its
// own write-mode handle, in order, with created_at values one apart
// from a fixed base so List's order is the same on every run. It needs
// a current store, and takes the store's lock shared as a running serve
// does, so it may run beside one. The ids are fixed, the first byte
// 0xf0 and the second the row's position.
func seedMembers(t *testing.T, stateDir string, members ...seedMember) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	h, err := store.OpenWrite(ctx, stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() {
		if err := h.Close(); err != nil {
			t.Errorf("close store handle: %v", err)
		}
	}()
	err = h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		for i, m := range members {
			at := int64(1_000_000 + i)
			var tombstonedAt any
			if m.tombstoned {
				tombstonedAt = at + 1
			}
			if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				[]byte{0xf0, byte(i + 1), 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, m.address, m.displayName, m.username, m.admin, tombstonedAt, at, at); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed members: %v", err)
	}
}

// rosterAPIEnviron returns serveAuthEnviron over a current store
// already holding one admin and one non-admin, so a shell built from it
// opens a store with a roster behind it.
func rosterAPIEnviron(t *testing.T) []string {
	t.Helper()
	environ := serveAuthEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	seedBackupStore(t, stateDir)
	seedMembers(t, stateDir,
		seedMember{address: rosterAdminAddress, displayName: "Group Admin", username: "groupadmin", admin: true},
		seedMember{address: rosterMemberAddress, displayName: "Member", username: "member"},
	)
	return environ
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
// under environ's own DUBPLATE_SESSION_SECRET and the lifetime and
// absolute cap the store under environ's state directory holds in
// force, issued at issued: the same values [newServeShell]'s own seam
// verifies against, so a test session is one a real login would have
// minted. The shell must already have been built, so its seed has
// committed.
func mintServeSession(t *testing.T, environ []string, address string, issued time.Time) string {
	t.Helper()
	var lifetime, absoluteCap time.Duration
	err := withStoreAccessor(t.Context(), environ, func(acc *settings.Accessor) error {
		var err error
		if lifetime, err = acc.Duration(t.Context(), sessionLifetimeSetting); err != nil {
			return err
		}
		absoluteCap, err = acc.Duration(t.Context(), sessionAbsoluteCapSetting)
		return err
	})
	if err != nil {
		t.Fatalf("read the session settings: %v", err)
	}
	secret := environValue(environ, "DUBPLATE_SESSION_SECRET")
	value, _, err := auth.MintSession([]byte(secret), address, issued, lifetime, absoluteCap)
	if err != nil {
		t.Fatalf("auth.MintSession(%s): %v", address, err)
	}
	return value
}

// withStoreAccessor calls fn with an accessor over the store under
// environ's DUBPLATE_STATE_DIR, on a shared write-mode handle beside
// any serve handle a test holds open, and closes the handle after. A
// test calls it after building the shell, so the values fn reads are
// the ones that shell's seed left in force.
func withStoreAccessor(ctx context.Context, environ []string, fn func(*settings.Accessor) error) (err error) {
	h, err := store.OpenWrite(ctx, environValue(environ, "DUBPLATE_STATE_DIR"), testBackupBusy)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, h.Close()) }()
	return fn(settings.NewAccessor(h, productionRegistry()))
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
		r.AddCookie(&http.Cookie{Name: "__Host-dubplate_session", Value: session})
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

// wantNotFoundBody is the body [http.NotFound] writes, [http.NotFoundHandler]'s
// own answer and the admin gate's refusal alike.
const wantNotFoundBody = "404 page not found\n"

// TestNewServeShell_RosterPOSTWithoutOriginSignalIsDenied proves the
// roster API sits behind the shell's own origin check like every other
// state-changing route: a POST carrying neither Origin nor
// Sec-Fetch-Site is refused before it ever reaches the admin gate, and
// the roster is unchanged.
func TestNewServeShell_RosterPOSTWithoutOriginSignalIsDenied(t *testing.T) {
	environ := rosterAPIEnviron(t)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	shell, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	session := mintServeSession(t, environ, rosterAdminAddress, time.Now())

	before := userAddresses(t, stateDir)

	req := sessionRequest(http.MethodPost, prefix.String()+"/roster", session, "", validPost)
	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, body = %s, want %d", rec.Code, rec.Body.String(), http.StatusForbidden)
	}

	if after := userAddresses(t, stateDir); !slices.Equal(before, after) {
		t.Errorf("roster changed: before %q, after %q", before, after)
	}
}

// validPost is a form every field check accepts for a new member,
// [internal/auth]'s own fixture body (admin_test.go), restated here so
// this file's tests need no cross-package export.
const validPost = "address=new%40example.com&display_name=New+Member&username=newmember"

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
	environ := rosterAPIEnviron(t)
	// [serve.NewShell] reads slog.Default() once, at construction, into
	// its own stored field, so the capture must be in place before the
	// shell (and the AdminConfig.Logger it hands the roster API) is
	// built, not merely before the request.
	logs := captureDefaultLogger(t)
	shell, err := buildServeShell(t, environ)
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

	if after := userAddresses(t, environValue(environ, "DUBPLATE_STATE_DIR")); !slices.Contains(after, "new@example.com") {
		t.Errorf("roster after the POST = %q, want it to name the new member", after)
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
// same-origin so both reach the gate and the 404 rather than the
// origin check. It then proves the same route answers an admin's GET
// with the roster list, so the 404s above are the non-admin's and not
// the path's.
func TestNewServeShell_RosterNonAdminIndistinguishableFromUnknownPath(t *testing.T) {
	environ := rosterAPIEnviron(t)
	shell, err := buildServeShell(t, environ)
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
// trailing-slash forms included, answers the shell's 404.
// Fixture state: the same admin's GET to the exact path answers the
// list in this same test, so the 404s below are the path's and not the
// session's.
func TestNewServeShell_RosterExactPathOnly(t *testing.T) {
	environ := rosterAPIEnviron(t)
	shell, err := buildServeShell(t, environ)
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

// TestNewServeShell_RosterAndMemberVerbShareStore proves "member list
// reads what the API wrote, and the API reads what the verb wrote".
// Fixture state: the shell is built, and so holds the store open with
// its roster, before either write; the verb then opens its own
// write-mode handle beside the running shell's, as an operator does on
// a live box, so each side's answer can only be explained by the shared
// store rather than by both having read one snapshot.
func TestNewServeShell_RosterAndMemberVerbShareStore(t *testing.T) {
	environ := rosterAPIEnviron(t)
	shell, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell() error = %v, want nil", err)
	}
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig() error = %v, want nil", err)
	}
	adminSession := mintServeSession(t, environ, rosterAdminAddress, time.Now())

	apiBody := "address=api-added%40example.com&display_name=API+Added&username=apiadded"
	apiReq := sessionRequest(http.MethodPost, prefix.String()+"/roster", adminSession, "same-origin", apiBody)
	apiRec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(apiRec, apiReq)
	if apiRec.Code != http.StatusCreated {
		t.Fatalf("API POST status = %d, body = %s, want %d", apiRec.Code, apiRec.Body.String(), http.StatusCreated)
	}

	if _, stderr, err := runMember(t, environ, "add", "--name", "Verb Added", "--username", "verbadded", "verb-added@example.com"); err != nil {
		t.Fatalf("member add error = %v; stderr = %q", err, stderr)
	}

	// "member list" reads what the API wrote: a fresh open, after both
	// writes, must see the API's own addition.
	listOut, _, err := runMember(t, environ, "list")
	if err != nil {
		t.Fatalf("member list error = %v", err)
	}
	if !strings.Contains(listOut, "api-added@example.com") {
		t.Errorf("member list output does not name the API's addition: %s", listOut)
	}

	// The API reads what the verb wrote.
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
// internal/auth's TestAddMember_DefaultIgnorableAddressRefused pin,
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
	shell, err := buildServeShell(t, environ)
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
// is what keeps the gate's 404 equal to the shell's 404.
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
		{name: "unclaimed subtree 404", method: http.MethodGet, target: "/dubplate/no-such-subtree", status: http.StatusNotFound},
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
		environ := rosterAPIEnviron(t)
		environ = replaceKey(environ, "DUBPLATE_PUBLIC_URL", c.publicURL)
		shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
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
// deeper path answers the shell's 404.
//
// Fixture state: the POSTs carry Sec-Fetch-Site: same-origin, so each
// passes the origin check and reaches the gate, the shell's 404, or the
// method check.
func TestNewServeShell_MembersRoute(t *testing.T) {
	environ := rosterAPIEnviron(t)
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
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
				t.Errorf("status %d body %q, want the shell's 404 %q", rec.Code, rec.Body.String(), unknown.Body.String())
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
	shell, err := buildServeShellWithLogger(t, environ, slog.New(slog.NewJSONHandler(&logs, nil)))
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
			shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
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
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
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
	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
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

// startServeShell builds a serve shell from environ as the verb does
// and returns it with its store handle, prefix, and config. The caller
// closes the handle; the shell's background work stops with ctx.
func startServeShell(t *testing.T, ctx context.Context, environ []string) (*serve.Shell, *store.Handle, serve.Prefix, *config.Config, error) {
	t.Helper()
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("config.Load: %v", err)
	}
	prefix, addrs, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig: %v", err)
	}
	shell, h, err := newServeShell(ctx, environ, productionRegistry(), prefix, addrs, cfg, discardTestLogger())
	return shell, h, prefix, cfg, err
}

// writeAdminRows writes each of values as an admin-sourced settings row
// through h, replacing any row the key holds, and bumps the settings
// generation in the same transaction: the shape of an admin write, as
// the settings home's writer will make it.
func writeAdminRows(t *testing.T, h *store.Handle, values map[string]string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	err := h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		for _, key := range slices.Sorted(maps.Keys(values)) {
			if _, err := tx.Exec(ctx, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
				VALUES (?, ?, 'admin', NULL, 'fixture', 2)
				ON CONFLICT (key) DO UPDATE SET value = excluded.value, source = 'admin', seed_value = NULL, actor = 'fixture', updated_at = 2`,
				key, values[key]); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx, `UPDATE settings_generation SET generation = generation + 1 WHERE singleton = 1`)
		return err
	})
	if err != nil {
		t.Fatalf("write admin rows: %v", err)
	}
}

// strangerLinkRequests posts n link requests for an address no member
// holds through handler and returns each status. A stranger is told so
// under the kind answer and sent no mail, so the requests spend the
// client's link-request limit and nothing else.
func strangerLinkRequests(handler http.Handler, prefix serve.Prefix, n int) []int {
	var codes []int
	for range n {
		rec := authRoutePOST(handler, prefix.String()+"/"+requestLinkRouteName, "address=stranger%40example.com")
		codes = append(codes, rec.Code)
	}
	return codes
}

// refreshedSession requests the roster as an admin over session with a
// same-origin GET, so the seam refreshes it, and returns the status,
// the refreshed cookie value the response sets, and its session,
// verified against secret at at with absoluteCap.
func refreshedSession(t *testing.T, handler http.Handler, prefix serve.Prefix, session, secret string, at time.Time, absoluteCap time.Duration) (int, string, auth.Session) {
	t.Helper()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, sessionRequest(http.MethodGet, prefix.String()+"/"+rosterRouteName, session, "same-origin", ""))
	if rec.Code != http.StatusOK {
		return rec.Code, "", auth.Session{}
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == "__Host-dubplate_session" {
			sess, err := auth.VerifySession([]byte(secret), nil, c.Value, at, absoluteCap)
			if err != nil {
				t.Fatalf("verify the refreshed session: %v", err)
			}
			return rec.Code, c.Value, sess
		}
	}
	t.Fatalf("a same-origin admin GET set no session cookie")
	return 0, "", auth.Session{}
}

// TestNewServeShell_AdminRowAppliesByMode is spec criterion 9: serve
// starts under environment values A, an admin row B is then written
// for each of five keys with the generation bumped, and each value
// applies on its own schedule. The link expiry and the session
// lifetime apply at the next operation, the absolute cap at the next
// request, and the client's link-request limit only at the next start.
//
// Fixture state: a current store holding an admin, and A and B chosen
// so each step tells them apart. The admin's capped session is five
// hours old, inside A's ten-hour cap and past B's three, and it is
// admitted before the write, so its refusal after the write is the
// cap's. Three stranger requests trip A's limit of two and not B's
// five.
func TestNewServeShell_AdminRowAppliesByMode(t *testing.T) {
	const (
		expiryKey   = "DUBPLATE_MAGIC_LINK_EXPIRY"
		lifetimeKey = "DUBPLATE_SESSION_LIFETIME"
		capKey      = "DUBPLATE_SESSION_ABSOLUTE_CAP"
		limitKey    = "DUBPLATE_LINK_REQUESTS_PER_CLIENT_PER_HOUR"
	)
	envA := map[string]string{expiryKey: "15m", lifetimeKey: "6h", capKey: "10h", limitKey: "2"}
	rowB := map[string]string{expiryKey: "40m", lifetimeKey: "1h", capKey: "3h", limitKey: "5"}

	environ := rosterAPIEnviron(t)
	for _, key := range slices.Sorted(maps.Keys(envA)) {
		environ = replaceKey(environ, key, envA[key])
	}
	secret := environValue(environ, "DUBPLATE_SESSION_SECRET")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	shell, h, prefix, cfg, err := startServeShell(t, ctx, environ)
	if err != nil {
		t.Fatalf("start under A: %v", err)
	}
	handler := shell.Handler()

	agedAt := time.Now().Add(-5 * time.Hour)
	aged, _, err := auth.MintSession([]byte(secret), rosterAdminAddress, agedAt, 6*time.Hour, 10*time.Hour)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	if code, _, _ := refreshedSession(t, handler, prefix, aged, secret, time.Now(), 10*time.Hour); code != http.StatusOK {
		t.Fatalf("the five-hour-old session under A's cap answered %d, want 200", code)
	}

	writeAdminRows(t, h, rowB)

	t.Run("a link minted after the write carries B's expiry", func(t *testing.T) {
		var mailed []string
		hc := authHandlersConfig(prefix, cfg, settings.NewAccessor(h, productionRegistry()), auth.NewRoster(h), nil, nil,
			auth.NewLimiter(auth.Limits{LinkRequestsPerAddress: 10, LinkRequestsPerClient: 10, ConsumptionAttemptsPerClient: 10, LinkMailCeiling: 10}),
			nil, func(*http.Request) string { return "192.0.2.50" }, discardTestLogger())
		hc.Send = func(_ context.Context, _, _, body string) error {
			mailed = append(mailed, body)
			return nil
		}
		rec := authRoutePOST(auth.NewHandlers(hc).RequestLink, "/request-link", "address="+url.QueryEscape(rosterMemberAddress))
		if rec.Code != http.StatusOK || len(mailed) != 1 {
			t.Fatalf("request = %d %s with %d mails, want 200 and one mail", rec.Code, rec.Body.String(), len(mailed))
		}
		if !strings.Contains(mailed[0], "expires in 40 minutes") {
			t.Errorf("mail body %q, want B's 40 minutes", mailed[0])
		}
		var token string
		for field := range strings.FieldsSeq(mailed[0]) {
			if u, err := url.Parse(field); err == nil && u.Scheme == "https" {
				token = u.Query().Get("t")
			}
		}
		tok, err := auth.VerifyLink([]byte(environValue(environ, "DUBPLATE_MAGIC_LINK_SECRET")), nil, token, time.Now())
		if err != nil {
			t.Fatalf("VerifyLink: %v", err)
		}
		if got := tok.ExpiresAt.Sub(tok.MintedAt); got != 40*time.Minute {
			t.Errorf("link life = %v, want B's 40m", got)
		}
	})

	t.Run("the next refresh uses B's lifetime", func(t *testing.T) {
		fresh, _, err := auth.MintSession([]byte(secret), rosterAdminAddress, time.Now(), 6*time.Hour, 10*time.Hour)
		if err != nil {
			t.Fatalf("MintSession: %v", err)
		}
		code, _, sess := refreshedSession(t, handler, prefix, fresh, secret, time.Now(), 3*time.Hour)
		if code != http.StatusOK {
			t.Fatalf("refresh answered %d, want 200", code)
		}
		if got := sess.ExpiresAt.Sub(sess.IssuedAt); got != time.Hour {
			t.Errorf("refreshed lifetime = %v, want B's 1h", got)
		}
	})

	t.Run("the next request enforces B's lowered cap", func(t *testing.T) {
		if code, _, _ := refreshedSession(t, handler, prefix, aged, secret, time.Now(), 10*time.Hour); code != http.StatusNotFound {
			t.Errorf("the five-hour-old session answered %d after the cap fell to 3h, want the gate's 404", code)
		}
	})

	t.Run("the limiter keeps A until a new start", func(t *testing.T) {
		if got, want := strangerLinkRequests(handler, prefix, 3), []int{200, 200, 429}; !slices.Equal(got, want) {
			t.Errorf("statuses = %v, want %v under A's limit of 2", got, want)
		}
	})

	cancel()
	if err := h.Close(); err != nil {
		t.Fatalf("close the first start's store: %v", err)
	}

	t.Run("a start with env A beside admin row B refuses naming both", func(t *testing.T) {
		envOnlyLimit := environ
		for _, key := range []string{expiryKey, lifetimeKey, capKey} {
			envOnlyLimit = dropKey(envOnlyLimit, key)
		}
		_, h2, _, _, err := startServeShell(t, t.Context(), envOnlyLimit)
		if err == nil {
			_ = h2.Close()
			t.Fatal("start succeeded, want the seed's refusal")
		}
		for _, want := range []string{limitKey, `"2"`, `"5"`} {
			if !strings.Contains(err.Error(), want) {
				t.Errorf("refusal %q, want it to name %s", err, want)
			}
		}
	})

	t.Run("the next start with the env unset uses B's limit", func(t *testing.T) {
		unset := environ
		for key := range envA {
			unset = dropKey(unset, key)
		}
		ctx2, cancel2 := context.WithCancel(t.Context())
		defer cancel2()
		shell2, h2, prefix2, _, err := startServeShell(t, ctx2, unset)
		if err != nil {
			t.Fatalf("restart: %v", err)
		}
		defer func() { _ = h2.Close() }()
		if got, want := strangerLinkRequests(shell2.Handler(), prefix2, 3), []int{200, 200, 200}; !slices.Equal(got, want) {
			t.Errorf("statuses = %v, want %v under B's limit of 5", got, want)
		}
	})
}

// TestNewServeShell_AdminRowRaisesTheCapThroughARefresh is spec
// criterion 9's raised-cap case: a cap raised by an admin row admits a
// session past its old cap only through a refresh, because the cookie's
// expiry was clamped to the old cap when it was minted. A session
// refreshed after the write carries an expiry past its anchor plus A,
// and it verifies at an instant between the two.
func TestNewServeShell_AdminRowRaisesTheCapThroughARefresh(t *testing.T) {
	const capKey = "DUBPLATE_SESSION_ABSOLUTE_CAP"
	environ := replaceKey(replaceKey(rosterAPIEnviron(t), "DUBPLATE_SESSION_LIFETIME", "2h"), capKey, "3h")
	secret := environValue(environ, "DUBPLATE_SESSION_SECRET")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	shell, h, prefix, _, err := startServeShell(t, ctx, environ)
	if err != nil {
		t.Fatalf("start under A: %v", err)
	}
	defer func() { _ = h.Close() }()

	anchor := time.Now().Add(-90 * time.Minute).Truncate(time.Second)
	session, minted, err := auth.MintSession([]byte(secret), rosterAdminAddress, anchor, 2*time.Hour, 3*time.Hour)
	if err != nil {
		t.Fatalf("MintSession: %v", err)
	}
	capA := anchor.Add(3 * time.Hour)
	if minted.ExpiresAt.After(capA) {
		t.Fatalf("minted expiry %v past the anchor plus A, %v", minted.ExpiresAt, capA)
	}

	writeAdminRows(t, h, map[string]string{capKey: "10h"})

	code, value, refreshed := refreshedSession(t, shell.Handler(), prefix, session, secret, time.Now(), 10*time.Hour)
	if code != http.StatusOK {
		t.Fatalf("refresh answered %d, want 200", code)
	}
	if !refreshed.ExpiresAt.After(capA) {
		t.Fatalf("refreshed expiry %v, want past the anchor plus A, %v", refreshed.ExpiresAt, capA)
	}
	between := capA.Add(refreshed.ExpiresAt.Sub(capA) / 2)
	var absoluteCap time.Duration
	err = withStoreAccessor(t.Context(), environ, func(acc *settings.Accessor) error {
		var err error
		absoluteCap, err = acc.Duration(t.Context(), sessionAbsoluteCapSetting)
		return err
	})
	if err != nil {
		t.Fatalf("read the cap in force: %v", err)
	}
	if _, err := auth.VerifySession([]byte(secret), nil, value, between, absoluteCap); err != nil {
		t.Errorf("verification at %v, between the anchor plus A and the refreshed expiry: %v, want admitted", between, err)
	}
}

// TestNewServeShell_RequestReadsTheSeededAnswer is spec criterion 12's
// request half: a request handled after the start reads the value the
// seed put in force. With the unknown-address answer seeded generic, a
// stranger's link request answers requested, where the default kind
// answer names it unknown.
func TestNewServeShell_RequestReadsTheSeededAnswer(t *testing.T) {
	for _, tt := range []struct {
		answer, want string
	}{
		{"", "unknown"},
		{"generic", "requested"},
	} {
		t.Run(cmp.Or(tt.answer, "the default"), func(t *testing.T) {
			environ := rosterAPIEnviron(t)
			if tt.answer != "" {
				environ = replaceKey(environ, "DUBPLATE_UNKNOWN_ADDRESS_ANSWER", tt.answer)
			}
			shell, err := buildServeShell(t, environ)
			if err != nil {
				t.Fatalf("buildServeShell: %v", err)
			}
			prefix, _, _, err := resolveServeConfig(environ)
			if err != nil {
				t.Fatalf("resolveServeConfig: %v", err)
			}
			rec := authRoutePOST(shell.Handler(), prefix.String()+"/"+requestLinkRouteName, "address=stranger%40example.com")
			var got struct {
				Outcome string `json:"outcome"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode %q: %v", rec.Body.String(), err)
			}
			if rec.Code != http.StatusOK || got.Outcome != tt.want {
				t.Errorf("stranger's request = %d %q, want 200 %q", rec.Code, got.Outcome, tt.want)
			}
		})
	}
}

// TestNewServeShell_NonceFallbackOutlivesTheExpiryCeiling proves serve
// hands the nonce store the link expiry's registered ceiling, never the
// expiry in force: a record the sweep cannot parse ages out only past
// max(the retention, the ceiling), so with the expiry and the retention
// both below the ceiling, a crash-truncated record older than both but
// younger than the ceiling survives serve's opening sweep.
func TestNewServeShell_NonceFallbackOutlivesTheExpiryCeiling(t *testing.T) {
	environ := replaceKey(replaceKey(serveAuthEnviron(t), "DUBPLATE_MAGIC_LINK_EXPIRY", "5m"), "DUBPLATE_NONCE_RETENTION", "30m")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	if _, err := auth.Open(stateDir); err != nil {
		t.Fatalf("auth.Open: %v", err)
	}
	record := filepath.Join(auth.StoreDir(stateDir), "nonce", "truncated")
	if err := os.WriteFile(record, nil, 0o600); err != nil {
		t.Fatalf("plant the truncated record: %v", err)
	}
	age := time.Now().Add(-45 * time.Minute)
	if err := os.Chtimes(record, age, age); err != nil {
		t.Fatalf("age the record: %v", err)
	}

	if _, err := buildServeShell(t, environ); err != nil {
		t.Fatalf("buildServeShell: %v", err)
	}
	if _, err := os.Stat(record); err != nil {
		t.Errorf("the opening sweep removed a record younger than the expiry ceiling %v: %v", magicLinkExpiryCeiling, err)
	}
}

// TestNewServeShell_UnreadableSettingFailsTheRequest proves the readers
// serve wires never fall back to a default: a row the accessor refuses,
// written after the start so the seed never saw it, fails the request
// that reads it through each surface's server-error path, and the log
// names the key. A link request answers 500 failed; the admin gate
// answers its 404 and logs at Error.
func TestNewServeShell_UnreadableSettingFailsTheRequest(t *testing.T) {
	tests := []struct {
		name, key, value string
		request          func(t *testing.T, environ []string, prefix serve.Prefix) *http.Request
		wantCode         int
	}{
		{
			name: "the body cap below its floor", key: "DUBPLATE_AUTH_REQUEST_BODY_CAP", value: "1",
			request: func(t *testing.T, _ []string, prefix serve.Prefix) *http.Request {
				return sessionRequest(http.MethodPost, prefix.String()+"/"+requestLinkRouteName, "", "same-origin", "address="+url.QueryEscape(rosterMemberAddress))
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "the answer outside its set", key: "DUBPLATE_UNKNOWN_ADDRESS_ANSWER", value: "open",
			request: func(t *testing.T, _ []string, prefix serve.Prefix) *http.Request {
				return sessionRequest(http.MethodPost, prefix.String()+"/"+requestLinkRouteName, "", "same-origin", "address=stranger%40example.com")
			},
			wantCode: http.StatusInternalServerError,
		},
		{
			name: "the absolute cap below its floor", key: "DUBPLATE_SESSION_ABSOLUTE_CAP", value: "1ms",
			request: func(t *testing.T, environ []string, prefix serve.Prefix) *http.Request {
				session := mintServeSession(t, environ, rosterAdminAddress, time.Now())
				return sessionRequest(http.MethodGet, prefix.String()+"/"+rosterRouteName, session, "", "")
			},
			wantCode: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := rosterAPIEnviron(t)
			var logs bytes.Buffer
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			cfg, err := config.Load(environ, productionRegistry())
			if err != nil {
				t.Fatalf("config.Load: %v", err)
			}
			prefix, addrs, _, err := resolveServeConfig(environ)
			if err != nil {
				t.Fatalf("resolveServeConfig: %v", err)
			}
			shell, h, err := newServeShell(ctx, environ, productionRegistry(), prefix, addrs, cfg, slog.New(slog.NewJSONHandler(&logs, nil)))
			if err != nil {
				t.Fatalf("newServeShell: %v", err)
			}
			defer func() { _ = h.Close() }()
			r := tt.request(t, environ, prefix)

			writeAdminRows(t, h, map[string]string{tt.key: tt.value})
			rec := httptest.NewRecorder()
			shell.Handler().ServeHTTP(rec, r)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d %s, want %d", rec.Code, rec.Body.String(), tt.wantCode)
			}
			var found bool
			for line := range strings.Lines(logs.String()) {
				var record map[string]any
				if err := json.Unmarshal([]byte(line), &record); err != nil {
					t.Fatalf("log line %q: %v", line, err)
				}
				reason, _ := record["reason"].(string)
				if record["level"] == "ERROR" && strings.Contains(reason, tt.key) {
					found = true
				}
			}
			if !found {
				t.Errorf("no Error record naming %s in:\n%s", tt.key, logs.String())
			}
		})
	}
}
