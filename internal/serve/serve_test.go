package serve

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/glw907/dubplate/internal/config"
)

// buildTestRoster returns a roster with exactly [testEmailAlice] on file
// under [testDirAlice], matching upload_test.go's own fixture identity so
// tests in this file can share those constants.
func buildTestRoster(t *testing.T) config.Roster {
	t.Helper()
	path := filepath.Join(t.TempDir(), "contributors.yaml")
	body := "contributors:\n  - name: " + testDirAlice + "\n    email: " + testEmailAlice + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write contributors: %v", err)
	}
	roster, err := config.LoadContributors(path)
	if err != nil {
		t.Fatalf("LoadContributors: %v", err)
	}
	return roster
}

// fakeVerify stands in for [access.Verifier.Verify]: it returns email and
// err regardless of the token it is handed, which is all enforceAccess's
// own tests need -- the real JWT machinery is internal/access's own.
func fakeVerify(email string, err error) func(context.Context, string) (string, error) {
	return func(context.Context, string) (string, error) { return email, err }
}

// --- mountBasePath: the /upload prefix ----------------------------------

func TestMountBasePath(t *testing.T) {
	echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Echo-Path", r.URL.Path)
		w.WriteHeader(http.StatusTeapot) // distinctive: proves next actually ran
	})
	h := mountBasePath(echo)

	t.Run("bare /upload redirects to /upload/", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/upload", nil))

		if w.Code != http.StatusMovedPermanently {
			t.Fatalf("status = %d, want 301", w.Code)
		}
		if loc := w.Header().Get("Location"); loc != "/upload/" {
			t.Fatalf("Location = %q, want /upload/", loc)
		}
	})

	t.Run("prefixed API path reaches next with the prefix stripped", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/upload/api/config", nil))

		if w.Code != http.StatusTeapot {
			t.Fatalf("status = %d, want 418 (next ran)", w.Code)
		}
		if got := w.Header().Get("X-Echo-Path"); got != "/api/config" {
			t.Fatalf("next saw path %q, want /api/config (prefix stripped)", got)
		}
	})

	t.Run("prefixed asset root reaches next as /", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/upload/", nil))

		if got := w.Header().Get("X-Echo-Path"); got != "/" {
			t.Fatalf("next saw path %q, want /", got)
		}
	})

	t.Run("unprefixed API path 404s, never reaches next", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/config", nil))

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
		if w.Header().Get("X-Echo-Path") != "" {
			t.Fatal("next ran on an unprefixed /api/ path; it must not")
		}
	})

	t.Run("unprefixed root 404s, never reaches next", func(t *testing.T) {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
	})

	t.Run("a path merely starting with the prefix string 404s, never reaches next", func(t *testing.T) {
		// "/uploadfoo" is not "/upload" plus a path segment -- a plain
		// [http.StripPrefix] would trim the literal string "/upload" and
		// pass "foo" straight through to next, which must not happen.
		w := httptest.NewRecorder()
		h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/uploadfoo", nil))

		if w.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", w.Code)
		}
		if w.Header().Get("X-Echo-Path") != "" {
			t.Fatal("next ran on /uploadfoo; it must not")
		}
	})
}

// --- securityHeaders: baseline response headers (F2.7) ------------------

func TestSecurityHeaders(t *testing.T) {
	echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := securityHeaders(echo)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	if w.Code != http.StatusTeapot {
		t.Fatalf("status = %d, want the wrapped handler to run", w.Code)
	}
	tests := map[string]string{
		"X-Content-Type-Options":  "nosniff",
		"Referrer-Policy":         "no-referrer",
		"X-Frame-Options":         "DENY",
		"Content-Security-Policy": "default-src 'self'; frame-ancestors 'none'; base-uri 'none'; object-src 'none'; form-action 'self'",
	}
	for header, want := range tests {
		if got := w.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}
}

// --- requireSameOrigin: CSRF defense on state-changing requests (F2.6) --

func TestRequireSameOrigin(t *testing.T) {
	echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })
	h := requireSameOrigin(echo)

	newReq := func(method, host string, setHeaders func(*http.Request)) *http.Request {
		r := httptest.NewRequest(method, "/api/session", nil)
		r.Host = host
		if setHeaders != nil {
			setHeaders(r)
		}
		return r
	}

	tests := []struct {
		name string
		req  *http.Request
		want int
	}{
		{
			"GET passes through with no headers at all",
			newReq(http.MethodGet, "music.example.com", nil),
			http.StatusTeapot,
		},
		{
			"POST with a matching Origin passes",
			newReq(http.MethodPost, "music.example.com", func(r *http.Request) {
				r.Header.Set("Origin", "https://music.example.com")
			}),
			http.StatusTeapot,
		},
		{
			"POST with a cross-origin Origin is refused",
			newReq(http.MethodPost, "music.example.com", func(r *http.Request) {
				r.Header.Set("Origin", "https://evil.example")
			}),
			http.StatusForbidden,
		},
		{
			"POST with no Origin but Sec-Fetch-Site: same-origin passes",
			newReq(http.MethodPost, "music.example.com", func(r *http.Request) {
				r.Header.Set("Sec-Fetch-Site", "same-origin")
			}),
			http.StatusTeapot,
		},
		{
			"POST with Sec-Fetch-Site: cross-site is refused",
			newReq(http.MethodPost, "music.example.com", func(r *http.Request) {
				r.Header.Set("Sec-Fetch-Site", "cross-site")
			}),
			http.StatusForbidden,
		},
		{
			// web/vite.config.ts's dev proxy runs with changeOrigin: true,
			// which rewrites Host to the proxied backend but leaves Origin
			// naming the Vite dev server and Sec-Fetch-Site untouched --
			// exactly this shape. F2.6's fix checks Sec-Fetch-Site first so
			// this still passes instead of 403ing every dev POST/PATCH.
			"POST through the dev proxy: Sec-Fetch-Site same-origin passes despite a mismatched Origin",
			newReq(http.MethodPost, "localhost:8080", func(r *http.Request) {
				r.Header.Set("Origin", "http://localhost:5173")
				r.Header.Set("Sec-Fetch-Site", "same-origin")
			}),
			http.StatusTeapot,
		},
		{
			"POST with neither header is refused",
			newReq(http.MethodPost, "music.example.com", nil),
			http.StatusForbidden,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			h.ServeHTTP(w, tt.req)
			if w.Code != tt.want {
				t.Errorf("status = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

// --- enforceAccess: per-surface failure shapes -------------------------

func TestEnforceAccess(t *testing.T) {
	roster := buildTestRoster(t)
	const adminEmail = "admin@example.com"

	echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email, _ := EmailFromContext(r.Context())
		w.Header().Set("X-Echo-Email", email)
		w.WriteHeader(http.StatusTeapot) // distinctive: proves next actually ran
	})

	newReq := func(method, path, token string) *http.Request {
		r := httptest.NewRequest(method, path, nil)
		r.RemoteAddr = "198.51.100.7:5555" // TEST-NET-2: deliberately non-loopback
		if token != "" {
			r.Header.Set(accessAssertionHeader, token)
		}
		return r
	}

	t.Run("api missing assertion gets 401 JSON", func(t *testing.T) {
		ac := accessConfig{verify: fakeVerify("", errors.New("no token")), roster: roster}
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, newReq(http.MethodGet, "/api/config", ""))

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
		var body accessFailure
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal body %q: %v", w.Body.String(), err)
		}
		if body.Error != "unauthenticated" {
			t.Fatalf("error = %q, want unauthenticated", body.Error)
		}
	})

	t.Run("files expired assertion gets 401 JSON", func(t *testing.T) {
		ac := accessConfig{verify: fakeVerify("", errors.New("token is expired")), roster: roster}
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, newReq(http.MethodPost, "/files/abc", "expired-token"))

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); !strings.Contains(ct, "json") {
			t.Fatalf("Content-Type = %q, want JSON", ct)
		}
	})

	t.Run("api roster-absent gets 403 JSON with warm copy", func(t *testing.T) {
		ac := accessConfig{verify: fakeVerify("stranger@example.com", nil), roster: roster, adminEmail: adminEmail}
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, newReq(http.MethodGet, "/api/history", "tok"))

		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403", w.Code)
		}
		var body accessFailure
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("unmarshal body %q: %v", w.Body.String(), err)
		}
		if body.Error != "not_on_roster" || !strings.Contains(body.Message, adminEmail) {
			t.Fatalf("body = %+v, want not_on_roster naming %s", body, adminEmail)
		}
	})

	t.Run("document roster-absent gets warm 200 HTML", func(t *testing.T) {
		ac := accessConfig{
			verify:     fakeVerify("stranger@example.com", nil),
			roster:     roster,
			adminEmail: adminEmail,
			warmPage:   warmRosterAbsentPage(adminEmail),
		}
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, newReq(http.MethodGet, "/", "tok"))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
			t.Fatalf("Content-Type = %q, want text/html", ct)
		}
		if !strings.Contains(w.Body.String(), adminEmail) {
			t.Fatalf("warm page body = %q, want it to name %s", w.Body.String(), adminEmail)
		}
	})

	t.Run("document missing assertion gets plain 401, not JSON", func(t *testing.T) {
		ac := accessConfig{verify: fakeVerify("", errors.New("no token")), roster: roster}
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, newReq(http.MethodGet, "/", ""))

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401", w.Code)
		}
		if ct := w.Header().Get("Content-Type"); strings.Contains(ct, "json") {
			t.Fatalf("Content-Type = %q, document surface must not render JSON", ct)
		}
	})

	t.Run("verified and on roster passes through with email in context", func(t *testing.T) {
		ac := accessConfig{verify: fakeVerify(testEmailAlice, nil), roster: roster}
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, newReq(http.MethodGet, "/api/config", "tok"))

		if w.Code != http.StatusTeapot {
			t.Fatalf("status = %d, want the wrapped handler to run", w.Code)
		}
		if got := w.Header().Get("X-Echo-Email"); got != testEmailAlice {
			t.Fatalf("context email = %q, want %q", got, testEmailAlice)
		}
	})
}

// --- dev-identity: startup refusal, off-by-default, per-request inertness --

func TestValidateDevIdentity(t *testing.T) {
	tests := []struct {
		name        string
		devIdentity string
		bindAddrs   []string
		wantErr     bool
	}{
		{"off by default, any bind addrs", "", []string{"0.0.0.0", "10.89.0.1"}, false},
		{"loopback v4 only, accepted", "geoff@example.com", []string{"127.0.0.1"}, false},
		{"loopback v4 and v6, accepted", "geoff@example.com", []string{"127.0.0.1", "::1"}, false},
		{"one non-loopback address, refused", "geoff@example.com", []string{"127.0.0.1", "10.89.0.1"}, true},
		{"unparseable address, refused", "geoff@example.com", []string{"not-an-ip"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDevIdentity(tt.devIdentity, tt.bindAddrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDevIdentity(%q, %v) = %v, wantErr %v", tt.devIdentity, tt.bindAddrs, err, tt.wantErr)
			}
		})
	}
}

func TestDevIdentityLoopbackOnly(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		want       bool
	}{
		{"loopback v4", "127.0.0.1:54321", true},
		{"loopback v6", "[::1]:54321", true},
		{"non-loopback", "203.0.113.5:54321", false},
		{"no port, unparseable by SplitHostPort", "203.0.113.5", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/api/config", nil)
			r.RemoteAddr = tt.remoteAddr
			if got := devIdentityLoopbackOnly(r); got != tt.want {
				t.Errorf("devIdentityLoopbackOnly(%q) = %v, want %v", tt.remoteAddr, got, tt.want)
			}
		})
	}
}

func TestEnforceAccess_DevIdentity(t *testing.T) {
	roster := buildTestRoster(t)
	echo := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusTeapot) })

	t.Run("bypasses verification on loopback", func(t *testing.T) {
		ac := accessConfig{
			verify: func(context.Context, string) (string, error) {
				t.Fatal("real verification must not run when the dev-identity bypass applies")
				return "", nil
			},
			roster:      roster,
			devIdentity: testEmailAlice,
		}
		r := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		r.RemoteAddr = "127.0.0.1:9999"
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, r)

		if w.Code != http.StatusTeapot {
			t.Fatalf("status = %d, want the dev identity to admit the request", w.Code)
		}
	})

	t.Run("is inert on a non-loopback request", func(t *testing.T) {
		verifyCalled := false
		ac := accessConfig{
			verify: func(context.Context, string) (string, error) {
				verifyCalled = true
				return "", errors.New("not a real token")
			},
			roster:      roster,
			devIdentity: testEmailAlice,
		}
		r := httptest.NewRequest(http.MethodGet, "/api/config", nil)
		r.RemoteAddr = "203.0.113.9:9999"
		r.Header.Set(accessAssertionHeader, "some-token")
		w := httptest.NewRecorder()
		enforceAccess(echo, ac).ServeHTTP(w, r)

		if !verifyCalled {
			t.Fatal("dev identity bypassed real verification on a non-loopback request")
		}
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want 401 from the real (failing) verification path", w.Code)
		}
	})
}

// --- startup preconditions ----------------------------------------------

func TestSameDevice(t *testing.T) {
	a := unix.Stat_t{Dev: 5}
	b := unix.Stat_t{Dev: 5}
	c := unix.Stat_t{Dev: 6}
	if !sameDevice(a, b) {
		t.Error("sameDevice with equal Dev = false, want true")
	}
	if sameDevice(a, c) {
		t.Error("sameDevice with different Dev = true, want false")
	}
}

func TestSameFilesystem(t *testing.T) {
	t.Run("two directories under the same temp root", func(t *testing.T) {
		root := t.TempDir()
		a, b := filepath.Join(root, "a"), filepath.Join(root, "b")
		if err := os.MkdirAll(a, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(b, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := sameFilesystem(a, b); err != nil {
			t.Fatalf("sameFilesystem(%s, %s) = %v, want nil", a, b, err)
		}
	})

	t.Run("a nonexistent path errors", func(t *testing.T) {
		if err := sameFilesystem(filepath.Join(t.TempDir(), "nope"), t.TempDir()); err == nil {
			t.Fatal("want an error for a nonexistent path")
		}
	})
}

// minimalTestConfig returns a [config.Config] with every field [NewServer]
// dereferences filled in against real temp directories, uploadsDir and
// inboxDir sharing one root so the same-filesystem precondition passes by
// default; callers needing the negative case redirect inboxDir themselves.
func minimalTestConfig(t *testing.T) *config.Config {
	t.Helper()
	root := t.TempDir()
	dirs := map[string]string{}
	for _, name := range []string{"uploads", "inbox", "run", "state"} {
		p := filepath.Join(root, name)
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", name, err)
		}
		dirs[name] = p
	}

	return &config.Config{
		AdminEmail: "admin@example.com",
		SMTP:       config.SMTPConfig{Host: "127.0.0.1", Port: 25, From: "dubplate@example.com"},
		Navidrome:  config.NavidromeConfig{URL: "http://127.0.0.1:4533", AdminUser: "admin", AdminPassword: "secret"},
		Serve: config.ServeConfig{
			BindAddrs:          []string{"127.0.0.1"},
			Port:               0,
			AccessTeamDomain:   "family.cloudflareaccess.com",
			AccessAUD:          "aud",
			PublicNavidromeURL: "https://music.example.com",
		},
		Upload: config.UploadConfig{ChunkMB: 8, MaxAlbumBytes: 1 << 30, PartialTTL: time.Hour, MaxSessionsPerUser: 2},
		Paths: config.Paths{
			UploadsDir: dirs["uploads"],
			InboxDir:   dirs["inbox"],
			TriggerDir: dirs["run"],
			StateDir:   dirs["state"],
		},
		LedgerPath:  filepath.Join(dirs["state"], "contributions.jsonl"),
		Thresholds:  config.Thresholds{StallMinutes: 30, RetentionDays: 60},
		Collections: []string{""},
	}
}

func TestNewServer_StartupPreconditions(t *testing.T) {
	probe := stubTagProbe(TagProbeResult{Confident: true}, nil, new(string))

	t.Run("succeeds when uploads and inbox share a filesystem", func(t *testing.T) {
		cfg := minimalTestConfig(t)
		roster := buildTestRoster(t)
		if _, err := NewServer(cfg, roster, NewServerDeps{Assets: emptyDistFS(t), TagProbe: probe}); err != nil {
			t.Fatalf("NewServer: %v", err)
		}
	})

	t.Run("refuses to start when InboxDir is unreachable", func(t *testing.T) {
		cfg := minimalTestConfig(t)
		cfg.Paths.InboxDir = filepath.Join(cfg.Paths.InboxDir, "does-not-exist")
		roster := buildTestRoster(t)
		if _, err := NewServer(cfg, roster, NewServerDeps{Assets: emptyDistFS(t), TagProbe: probe}); err == nil {
			t.Fatal("want an error when the same-filesystem precondition cannot even be checked")
		}
	})

	t.Run("refuses to start when --dev-identity targets a non-loopback bind address", func(t *testing.T) {
		cfg := minimalTestConfig(t)
		cfg.Serve.BindAddrs = []string{"127.0.0.1", "10.89.0.1"}
		roster := buildTestRoster(t)
		if _, err := NewServer(cfg, roster, NewServerDeps{Assets: emptyDistFS(t), DevIdentity: testEmailAlice, TagProbe: probe}); err == nil {
			t.Fatal("want an error for a non-loopback bind address under --dev-identity")
		}
	})

	// A Server built on NewEngine's fail-closed default would hold every
	// main-lane upload unpromoted forever, staged directory and all, with
	// no retry that could ever succeed. Refusing to start is the only
	// honest answer.
	t.Run("refuses to start without a tag-check probe", func(t *testing.T) {
		cfg := minimalTestConfig(t)
		roster := buildTestRoster(t)
		if _, err := NewServer(cfg, roster, NewServerDeps{Assets: emptyDistFS(t)}); err == nil {
			t.Fatal("want an error when no tag-check probe was supplied")
		}
	})

	t.Run("hands the supplied probe to the upload engine", func(t *testing.T) {
		cfg := minimalTestConfig(t)
		roster := buildTestRoster(t)
		const wiredAlbum = "the wired probe ran"
		var probedDir string
		wired := stubTagProbe(TagProbeResult{Confident: true, Album: wiredAlbum}, nil, &probedDir)

		srv, err := NewServer(cfg, roster, NewServerDeps{Assets: emptyDistFS(t), TagProbe: wired})
		if err != nil {
			t.Fatalf("NewServer: %v", err)
		}
		got, err := srv.engine.tagProbe(t.Context(), cfg.Paths.UploadsDir)
		if err != nil {
			t.Fatalf("engine tag probe: %v", err)
		}
		if !got.Confident || got.Album != wiredAlbum {
			t.Errorf("engine tag probe = %+v, want the supplied probe's confident match for %q", got, wiredAlbum)
		}
		if probedDir != cfg.Paths.UploadsDir {
			t.Errorf("engine tag probe ran against %q, want %q", probedDir, cfg.Paths.UploadsDir)
		}
	})
}

// TestNewServer_FullStackAppliesSecurityAndCSRF proves securityHeaders
// and requireSameOrigin are actually wired into the handler NewServer
// builds, not only unit-tested in isolation: a real request through
// s.handler carries the baseline headers, and a forged cross-origin POST
// to /upload/api/session is refused before it ever reaches the session
// store.
func TestNewServer_FullStackAppliesSecurityAndCSRF(t *testing.T) {
	cfg := minimalTestConfig(t)
	roster := buildTestRoster(t)
	probe := stubTagProbe(TagProbeResult{Confident: true}, nil, new(string))

	srv, err := NewServer(cfg, roster, NewServerDeps{Assets: emptyDistFS(t), DevIdentity: testEmailAlice, TagProbe: probe})
	if err != nil {
		t.Fatalf("NewServer: %v", err)
	}

	loopbackReq := func(method, path string, setHeaders func(*http.Request)) *http.Request {
		r := httptest.NewRequest(method, path, nil)
		r.RemoteAddr = "127.0.0.1:9999" // required for the --dev-identity bypass
		if setHeaders != nil {
			setHeaders(r)
		}
		return r
	}

	t.Run("a plain GET carries the baseline security headers", func(t *testing.T) {
		w := httptest.NewRecorder()
		srv.handler.ServeHTTP(w, loopbackReq(http.MethodGet, "/upload/api/config", nil))

		if w.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", w.Code)
		}
		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Errorf("X-Content-Type-Options = %q, want nosniff", got)
		}
		if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
			t.Errorf("X-Frame-Options = %q, want DENY", got)
		}
	})

	t.Run("a forged cross-origin POST to /api/session is refused", func(t *testing.T) {
		w := httptest.NewRecorder()
		srv.handler.ServeHTTP(w, loopbackReq(http.MethodPost, "/upload/api/session", func(r *http.Request) {
			r.Header.Set("Origin", "https://evil.example")
			r.Header.Set("Content-Type", "application/json")
		}))

		if w.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want 403 (refused before it ever reached the session store)", w.Code)
		}
	})
}

// emptyDistFS returns a minimal fs.FS holding just "dist/index.html", the
// shape [NewServer] expects from web.Assets (still rooted at "dist").
func emptyDistFS(t *testing.T) fs.FS {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "dist"), 0o755); err != nil {
		t.Fatalf("mkdir dist: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "dist", "index.html"), []byte("<!doctype html>"), 0o644); err != nil {
		t.Fatalf("write index.html: %v", err)
	}
	return os.DirFS(root)
}

// --- nudge trigger -------------------------------------------------------

func TestWriteNudgeTrigger(t *testing.T) {
	t.Run("writes an empty file named for the session", func(t *testing.T) {
		dir := t.TempDir()
		id, err := newSessionID()
		if err != nil {
			t.Fatalf("newSessionID: %v", err)
		}
		writeNudgeTrigger(dir, id)

		info, err := os.Stat(filepath.Join(dir, id))
		if err != nil {
			t.Fatalf("trigger file missing: %v", err)
		}
		if info.Size() != 0 {
			t.Fatalf("trigger file size = %d, want 0 (presence is the signal)", info.Size())
		}
	})

	t.Run("refuses a malformed session id", func(t *testing.T) {
		dir := t.TempDir()
		writeNudgeTrigger(dir, "../escape")

		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read dir: %v", err)
		}
		if len(entries) != 0 {
			t.Fatalf("trigger dir has %d entries, want 0 for a malformed id", len(entries))
		}
	})

	t.Run("a missing trigger directory logs and does not panic", func(t *testing.T) {
		id, err := newSessionID()
		if err != nil {
			t.Fatalf("newSessionID: %v", err)
		}
		writeNudgeTrigger(filepath.Join(t.TempDir(), "does-not-exist"), id)
	})
}

// --- GET /api/config -----------------------------------------------------

func TestFormatSpecText(t *testing.T) {
	got := formatSpecText(44100, 16, 2<<30)
	want := "FLAC only, 44100 Hz / 16-bit or better, one album per upload, up to 2.0 GiB total."
	if got != want {
		t.Errorf("formatSpecText(44100, 16, 2GiB) = %q, want %q", got, want)
	}
}

func TestHumanGiB(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{2 << 30, "2.0 GiB"},
		{1 << 30, "1.0 GiB"},
		{3*(1<<30) + (1 << 29), "3.5 GiB"},
	}
	for _, tt := range tests {
		if got := humanGiB(tt.bytes); got != tt.want {
			t.Errorf("humanGiB(%d) = %q, want %q", tt.bytes, got, tt.want)
		}
	}
}

// getConfig serves GET /api/config from a Server wrapping collections and
// an Engine built with opts, then decodes the wire response. It is the one
// place the fixture config below is spelled out.
func getConfig(t *testing.T, collections []string, opts Options) configResponse {
	t.Helper()
	cfg := &config.Config{
		Serve:       config.ServeConfig{SiteName: "dubplate"},
		Collections: collections,
		Upload:      config.UploadConfig{ChunkMB: 8, MaxAlbumBytes: 2 << 30, MaxSessionsPerUser: 3},
		Thresholds:  config.Thresholds{RetentionDays: 60},
		Paths:       config.Paths{UploadsDir: t.TempDir()},
	}
	engine, err := NewEngine(cfg, buildTestRoster(t), opts)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}
	s := &Server{cfg: cfg, engine: engine}
	w := httptest.NewRecorder()
	s.handleConfig(w, httptest.NewRequest(http.MethodGet, "/api/config", nil))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var got configResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal %q: %v", w.Body.String(), err)
	}
	return got
}

func TestHandleConfig(t *testing.T) {
	got := getConfig(t, []string{"", "family"}, Options{})

	if got.SiteName != "dubplate" {
		t.Errorf("SiteName = %q, want dubplate", got.SiteName)
	}
	if len(got.Collections) != 2 || got.Collections[0] != "" || got.Collections[1] != "family" {
		t.Errorf("Collections = %v, want [\"\" \"family\"]", got.Collections)
	}
	if got.ChunkMB != 8 {
		t.Errorf("ChunkMB = %d, want 8", got.ChunkMB)
	}
	if got.MaxAlbumBytes != 2<<30 {
		t.Errorf("MaxAlbumBytes = %d, want %d", got.MaxAlbumBytes, 2<<30)
	}
	if got.MaxSessionsPerUser != 3 {
		t.Errorf("MaxSessionsPerUser = %d, want 3", got.MaxSessionsPerUser)
	}
	if got.ReviewClockDays != 60 {
		t.Errorf("ReviewClockDays = %d, want 60", got.ReviewClockDays)
	}
	wantSpec := "FLAC only, 44100 Hz / 16-bit or better, one album per upload, up to 2.0 GiB total."
	if got.FormatSpec != wantSpec {
		t.Errorf("FormatSpec = %q, want %q", got.FormatSpec, wantSpec)
	}
}

// TestHandleConfig_PrivateCollection proves the wire field tracks an
// operator's [Options.PrivateCollection] rather than always answering the
// documented default: the client tells the private lane apart by this
// field, never by matching the default name itself.
func TestHandleConfig_PrivateCollection(t *testing.T) {
	tests := []struct {
		name        string
		collections []string
		opts        Options
		want        string
	}{
		{"unset Options falls back to the default", []string{"", "family"}, Options{}, defaultPrivateCollection},
		{
			"operator override is served verbatim",
			[]string{"", "Home Recordings"},
			Options{PrivateCollection: "Home Recordings"},
			"Home Recordings",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getConfig(t, tt.collections, tt.opts).PrivateCollection; got != tt.want {
				t.Errorf("PrivateCollection = %q, want %q", got, tt.want)
			}
		})
	}
}

// --- listener: exact BindAddrs, IP_FREEBIND -------------------------------

func TestFreebindListeners_BindsAnAddressNoInterfaceHas(t *testing.T) {
	// 203.0.113.0/24 is TEST-NET-3 (RFC 5737): reserved for documentation,
	// never assigned to a real interface, so a plain net.Listen (no
	// IP_FREEBIND) is expected to fail here -- exactly the boot-time
	// bridge-gateway-not-up-yet case N7 exists for.
	const absentAddr = "203.0.113.77"

	if l, err := net.Listen("tcp", net.JoinHostPort(absentAddr, "0")); err == nil {
		_ = l.Close()
		t.Skip("this host has 203.0.113.77 configured; the freebind contrast is meaningless here")
	}

	listeners, err := freebindListeners(context.Background(), []string{absentAddr}, 0)
	if err != nil {
		t.Fatalf("freebindListeners(%s): %v", absentAddr, err)
	}
	defer func() {
		for _, l := range listeners {
			_ = l.Close()
		}
	}()

	if len(listeners) != 1 {
		t.Fatalf("got %d listeners, want 1", len(listeners))
	}
	if !strings.HasPrefix(listeners[0].Addr().String(), absentAddr) {
		t.Fatalf("listener addr = %q, want prefix %q", listeners[0].Addr().String(), absentAddr)
	}
}

func TestFreebindListeners_OneListenerPerAddr(t *testing.T) {
	listeners, err := freebindListeners(context.Background(), []string{"127.0.0.1", "127.0.0.1"}, 0)
	if err != nil {
		t.Fatalf("freebindListeners: %v", err)
	}
	defer func() {
		for _, l := range listeners {
			_ = l.Close()
		}
	}()
	if len(listeners) != 2 {
		t.Fatalf("got %d listeners, want 2 (one per BindAddrs entry)", len(listeners))
	}
}

func TestFreebindListeners_RollsBackOnFailure(t *testing.T) {
	listeners, err := freebindListeners(context.Background(), []string{"127.0.0.1", "not-an-ip"}, 0)
	if err == nil {
		t.Fatal("want an error for an unparseable bind address")
	}
	if listeners != nil {
		t.Fatalf("want nil listeners on failure, got %v", listeners)
	}
}

// --- graceful shutdown: drains an in-flight request -----------------------

func TestServeUntilDone_DrainsInFlightRequest(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("done"))
	})

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	srv := &http.Server{Handler: handler}
	ctx, cancel := context.WithCancel(context.Background())

	runDone := make(chan error, 1)
	go func() { runDone <- serveUntilDone(ctx, srv, []net.Listener{l}, 5*time.Second) }()

	client := &http.Client{Timeout: 3 * time.Second}
	reqDone := make(chan struct{})
	var reqStatus int
	var reqBody string
	var reqErr error
	go func() {
		defer close(reqDone)
		resp, err := client.Get("http://" + l.Addr().String() + "/")
		if err != nil {
			reqErr = err
			return
		}
		defer func() { _ = resp.Body.Close() }()
		body, _ := io.ReadAll(resp.Body)
		reqStatus, reqBody = resp.StatusCode, string(body)
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("in-flight request never reached the handler")
	}

	// Simulate SIGTERM while the request is still blocked inside the
	// handler: Shutdown must wait for it, never sever it.
	cancel()
	time.Sleep(50 * time.Millisecond) // let Shutdown actually start draining
	close(release)

	select {
	case <-reqDone:
		if reqErr != nil {
			t.Fatalf("in-flight request failed instead of draining: %v", reqErr)
		}
		if reqStatus != http.StatusOK || reqBody != "done" {
			t.Fatalf("in-flight response = %d %q, want 200 \"done\"", reqStatus, reqBody)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("in-flight request never completed")
	}

	select {
	case err := <-runDone:
		if err != nil {
			t.Fatalf("serveUntilDone returned an error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("serveUntilDone never returned after the drain")
	}
}
