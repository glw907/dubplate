package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"html"
	"io/fs"
	"log/slog"
	"mime"
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

	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/web"
)

// fixtureThemeCSS declares the three composed selectors a fixture
// build carries, in both quoted and unquoted form, mirroring the real
// bundle's own shape (internal/serve/document_test.go's ThemeSelectors
// cases already prove the extraction handles both).
// Each also declares a --color-neutral under its color scheme, as
// every real composed theme does, the system theme's dark value inside
// its media branch.
const fixtureThemeCSS = `[data-theme=dubplate-label]{color-scheme:light;--color-base-100:#fff;--color-neutral:#111}` +
	`[data-theme="dubplate-label-night"]{color-scheme:dark;--color-base-100:#000;--color-neutral:#222}` +
	`[data-theme=dubplate-system]{color-scheme:light;--color-base-100:#fff;--color-neutral:#111}` +
	`@media (prefers-color-scheme:dark){[data-theme=dubplate-system]{color-scheme:dark;--color-neutral:#222}}`

// fixtureSiteName is the site name every web route test builds with.
const fixtureSiteName = "Fixture Group"

// fixtureCSPMeta is the CSP meta every fixture document carries: a
// script-src with no hash, which a document holding no inline script
// satisfies.
const fixtureCSPMeta = `<meta http-equiv="content-security-policy" content="script-src 'self'">`

// newFixtureBuildFS returns a build tree fstest.MapFS shaped like the
// real SvelteKit output this task rewrites: a login document, a
// login/link document, and a members document, each carrying a CSP
// meta, the placeholder on its data-theme attribute, and a stylesheet
// link resolved relative to its own path, and the linked CSS itself.
// The two counts set how often the login and login/link documents
// carry the placeholder; the members document carries it once.
func newFixtureBuildFS(loginPlaceholderCount, linkPlaceholderCount int) fstest.MapFS {
	token := func(n int) string {
		return strings.Repeat(themePlaceholder, n)
	}
	loginDoc := `<html data-theme="` + token(loginPlaceholderCount) + `"><head>` + fixtureCSPMeta + `<link href="./_app/assets/0.css" rel="stylesheet"></head></html>`
	linkDoc := `<html data-theme="` + token(linkPlaceholderCount) + `"><head>` + fixtureCSPMeta + `<link href="../_app/assets/0.css" rel="stylesheet"></head></html>`
	membersDoc := `<html data-theme="` + themePlaceholder + `"><head>` + fixtureCSPMeta + `<link href="./_app/assets/0.css" rel="stylesheet"></head></html>`
	return fstest.MapFS{
		"login.html":             &fstest.MapFile{Data: []byte(loginDoc)},
		"login/link.html":        &fstest.MapFile{Data: []byte(linkDoc)},
		"members.html":           &fstest.MapFile{Data: []byte(membersDoc)},
		"_app/assets/0.css":      &fstest.MapFile{Data: []byte(fixtureThemeCSS)},
		"_app/entry/start.js":    &fstest.MapFile{Data: []byte("// entry")},
		"login/nested/other.txt": &fstest.MapFile{Data: []byte("unrelated")},
		"fonts/fixture.woff2":    &fstest.MapFile{Data: []byte("fixture woff2 bytes")},
		"favicon.svg":            &fstest.MapFile{Data: []byte("<svg/>")},
		"favicon.png":            &fstest.MapFile{Data: []byte("fixture png bytes")},
		"mark/mark-small.svg":    &fstest.MapFile{Data: []byte("<svg/>")},
	}
}

func webTestPrefix(t *testing.T) serve.Prefix {
	t.Helper()
	prefix, err := serve.NewPrefix("/dubplate")
	if err != nil {
		t.Fatalf("serve.NewPrefix: %v", err)
	}
	return prefix
}

func TestNewWebRoutesFromFS_States(t *testing.T) {
	prefix := webTestPrefix(t)

	tests := []struct {
		name       string
		buildFS    fstest.MapFS
		environ    []string
		wantErrHas string
	}{
		{
			name:       "placeholder-only tree",
			buildFS:    fstest.MapFS{"placeholder.txt": &fstest.MapFile{Data: []byte("placeholder")}},
			wantErrHas: "npm run build",
		},
		{
			name:       "login.html missing",
			buildFS:    fstest.MapFS{"login/link.html": &fstest.MapFile{Data: []byte("x")}},
			wantErrHas: `"login.html"`,
		},
		{
			name: "login/link.html missing",
			buildFS: func() fstest.MapFS {
				m := newFixtureBuildFS(1, 1)
				delete(m, "login/link.html")
				return m
			}(),
			wantErrHas: `"login/link.html"`,
		},
		{
			name: "members.html missing",
			buildFS: func() fstest.MapFS {
				m := newFixtureBuildFS(1, 1)
				delete(m, "members.html")
				return m
			}(),
			wantErrHas: `"members.html"`,
		},
		{
			name: "members.html holding an inline script its CSP meta does not hash",
			buildFS: func() fstest.MapFS {
				m := newFixtureBuildFS(1, 1)
				m["members.html"] = &fstest.MapFile{Data: []byte(`<html data-theme="` + themePlaceholder + `"><head>` + fixtureCSPMeta + `<link href="./_app/assets/0.css" rel="stylesheet"></head><body><script>boot();</script></body></html>`)}
				return m
			}(),
			wantErrHas: `document "members.html": an inline script's hash is not in the CSP meta's script-src`,
		},
		{
			name:       "placeholder absent from login.html",
			buildFS:    newFixtureBuildFS(0, 1),
			wantErrHas: `"login.html"`,
		},
		{
			name:       "placeholder present twice in login/link.html",
			buildFS:    newFixtureBuildFS(1, 2),
			wantErrHas: `"login/link.html"`,
		},
		{
			name:       "match value outside the accepted set",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{"DUBPLATE_THEME_MATCH=other"},
			wantErrHas: `must be exactly "dubplate"`,
		},
		{
			name:       "palette value carrying an uppercase letter",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{"DUBPLATE_THEME_PALETTE=Label"},
			wantErrHas: "must match",
		},
		{
			name:       "palette value carrying a quote",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{`DUBPLATE_THEME_PALETTE=label"`},
			wantErrHas: "must match",
		},
		{
			name:       "palette value carrying an angle bracket",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{"DUBPLATE_THEME_PALETTE=label>"},
			wantErrHas: "must match",
		},
		{
			name:       "palette value carrying whitespace",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{"DUBPLATE_THEME_PALETTE=la bel"},
			wantErrHas: "must match",
		},
		{
			name:       "palette composed name absent from the bundle",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{"DUBPLATE_THEME_PALETTE=gruvbox"},
			wantErrHas: `no theme selector named "dubplate-gruvbox"`,
		},
		{
			name: "favicon.png missing",
			buildFS: func() fstest.MapFS {
				m := newFixtureBuildFS(1, 1)
				delete(m, "favicon.png")
				return m
			}(),
			wantErrHas: `web: favicon: open favicon.png`,
		},
		{
			name: "no mark file",
			buildFS: func() fstest.MapFS {
				m := newFixtureBuildFS(1, 1)
				delete(m, "mark/mark-small.svg")
				m["mark/README"] = &fstest.MapFile{Data: []byte("not a mark")}
				return m
			}(),
			wantErrHas: "web: mark: no .svg file",
		},
		{
			name:       "palette value a prefix of a real name",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{"DUBPLATE_THEME_PALETTE=label-n"},
			wantErrHas: `no theme selector named "dubplate-label-n"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newWebRoutesFromFS(tt.environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), tt.buildFS)
			if err == nil {
				t.Fatalf("newWebRoutesFromFS: got nil error, want one containing %q", tt.wantErrHas)
			}
			if !strings.Contains(err.Error(), tt.wantErrHas) {
				t.Fatalf("newWebRoutesFromFS error = %q, want it to contain %q", err.Error(), tt.wantErrHas)
			}
		})
	}
}

func TestNewWebRoutesFromFS_EmptyKeysResolveToDefaults(t *testing.T) {
	prefix := webTestPrefix(t)
	buildFS := newFixtureBuildFS(1, 1)

	routes, err := newWebRoutesFromFS(
		[]string{"DUBPLATE_THEME_MATCH=", "DUBPLATE_THEME_PALETTE="},
		prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), buildFS,
	)
	if err != nil {
		t.Fatalf("newWebRoutesFromFS: %v", err)
	}

	r := httptest.NewRequest(http.MethodGet, "/dubplate/login", nil)
	rec := httptest.NewRecorder()
	routes.login.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /dubplate/login: status %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `data-theme="dubplate-system"`) {
		t.Errorf("body = %q, want it to carry data-theme=%q", rec.Body.String(), "dubplate-system")
	}
}

// TestNewWebRoutesFromFS_AcceptedPaletteValues proves each of the
// bundle's three stampable names (R-5cii12's accepted set) proceeds
// and stamps its own composed name on a request with no cookie.
func TestNewWebRoutesFromFS_AcceptedPaletteValues(t *testing.T) {
	prefix := webTestPrefix(t)
	buildFS := newFixtureBuildFS(1, 1)

	for _, palette := range []string{"label", "label-night", "system"} {
		t.Run(palette, func(t *testing.T) {
			routes, err := newWebRoutesFromFS([]string{"DUBPLATE_THEME_PALETTE=" + palette}, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), buildFS)
			if err != nil {
				t.Fatalf("newWebRoutesFromFS: %v", err)
			}
			r := httptest.NewRequest(http.MethodGet, "/dubplate/login", nil)
			rec := httptest.NewRecorder()
			routes.login.ServeHTTP(rec, r)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			want := `data-theme="dubplate-` + palette + `"`
			if !strings.Contains(rec.Body.String(), want) {
				t.Errorf("body = %q, want it to carry %s", rec.Body.String(), want)
			}
		})
	}
}

func TestNewWebRoutesFromFS_LoginRouting(t *testing.T) {
	prefix := webTestPrefix(t)
	buildFS := newFixtureBuildFS(1, 1)
	routes, err := newWebRoutesFromFS([]string{"DUBPLATE_THEME_PALETTE=label"}, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), buildFS)
	if err != nil {
		t.Fatalf("newWebRoutesFromFS: %v", err)
	}

	tests := []struct {
		name       string
		path       string
		method     string
		wantStatus int
	}{
		{name: "login exact path", path: "/dubplate/login", method: http.MethodGet, wantStatus: http.StatusOK},
		{name: "login/link exact path", path: "/dubplate/login/link", method: http.MethodGet, wantStatus: http.StatusOK},
		{name: "login trailing slash", path: "/dubplate/login/", method: http.MethodGet, wantStatus: http.StatusNotFound},
		{name: "an unregistered deeper path", path: "/dubplate/login/other", method: http.MethodGet, wantStatus: http.StatusNotFound},
		{name: "login POST", path: "/dubplate/login", method: http.MethodPost, wantStatus: http.StatusMethodNotAllowed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()
			routes.login.ServeHTTP(rec, r)
			if rec.Code != tt.wantStatus {
				t.Errorf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.wantStatus)
			}
		})
	}
}

func TestStampablePalettes(t *testing.T) {
	selectors := map[string]struct{}{
		"dubplate-label":       {},
		"dubplate-label-night": {},
		"dubplate-system":      {},
		"light":                {}, // no "dubplate-" prefix: DaisyUI's own bundled default, must be excluded
		"dark":                 {},
	}
	got := stampablePalettes(selectors, "dubplate")
	want := []string{"label", "label-night", "system"}
	if len(got) != len(want) {
		t.Fatalf("stampablePalettes = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("stampablePalettes[%d] = %q, want %q", i, got[i], w)
		}
	}
}

// scriptBodyRE extracts every inline <script>...</script> body, in
// document order, independent of internal/serve's own rewrite logic.
var scriptBodyRE = regexp.MustCompile(`(?is)<script(?:\s[^>]*)?>(.*?)</script>`)

func extractScriptBodies(doc []byte) []string {
	var bodies []string
	for _, m := range scriptBodyRE.FindAllSubmatch(doc, -1) {
		bodies = append(bodies, string(m[1]))
	}
	return bodies
}

// TestNewWebRoutes_RewriteChangesOnlyTheToken proves the rewrite's
// whole edit against the real embedded bundle, both documents and
// every stampable palette, with an operator CSS file configured so
// both tier links are inserted: the served body, with the head block
// (the two tier links, the site-name and robots metas, the palette's
// theme-color metas, and the icon links, written out below) removed
// and the composed name swapped back for [themePlaceholder], is
// byte-equal to the embedded original, and every inline <script> body
// served is byte-identical to the original's, in the same order. The
// theme-color values are the visual identity spec's neutrals: #1f1b16
// for light and #110f0d for dark.
func TestNewWebRoutes_RewriteChangesOnlyTheToken(t *testing.T) {
	prefix := webTestPrefix(t)

	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	environ := []string{"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, []byte("body{}"))}
	routes, err := newWebRoutes(environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger())
	if err != nil {
		t.Fatalf("newWebRoutes: %v", err)
	}

	documents := []struct {
		name  string
		path  string
		tiers string // the two inserted links, written out
	}{
		{
			name:  loginDocumentName,
			path:  prefix.String() + "/login",
			tiers: `<link href="./branding.css" rel="stylesheet"><link href="./operator.css" rel="stylesheet">`,
		},
		{
			name:  loginLinkDocumentName,
			path:  prefix.String() + "/login/link",
			tiers: `<link href="../branding.css" rel="stylesheet"><link href="../operator.css" rel="stylesheet">`,
		},
	}

	themeColorMetas := map[string]string{
		"label":       `<meta name="theme-color" content="#1f1b16">`,
		"label-night": `<meta name="theme-color" content="#110f0d">`,
		"system": `<meta name="theme-color" media="(prefers-color-scheme: light)" content="#1f1b16">` +
			`<meta name="theme-color" media="(prefers-color-scheme: dark)" content="#110f0d">`,
	}
	for _, doc := range documents {
		original, err := fs.ReadFile(buildFS, doc.name)
		if err != nil {
			t.Fatalf("fs.ReadFile(%q): %v", doc.name, err)
		}
		wantScripts := extractScriptBodies(original)
		if len(wantScripts) == 0 {
			t.Fatalf("%s carries no inline <script>; the fixture proves nothing", doc.name)
		}

		for _, palette := range []string{"label", "label-night", "system"} {
			t.Run(doc.name+"/"+palette, func(t *testing.T) {
				r := httptest.NewRequest(http.MethodGet, doc.path, nil)
				r.Header.Set("Cookie", serve.ThemeCookieName+"="+palette)
				rec := httptest.NewRecorder()
				routes.login.ServeHTTP(rec, r)
				if rec.Code != http.StatusOK {
					t.Fatalf("status = %d, want 200", rec.Code)
				}
				served := rec.Body.Bytes()

				head := doc.tiers + `<meta name="dubplate-site-name" content="Fixture Group">` +
					`<meta name="robots" content="noindex, nofollow">` + themeColorMetas[palette] +
					`<link rel="icon" href="/dubplate/favicon.svg" type="image/svg+xml">` +
					`<link rel="icon" href="/dubplate/favicon.png" type="image/png" sizes="32x32">`
				if n := bytes.Count(served, []byte(head)); n != 1 {
					t.Fatalf("%s served for %q carries the head block %d times, want 1:\n%s", doc.name, palette, n, served)
				}
				composed := acceptedThemeMatch + "-" + palette
				restored := bytes.Replace(served, []byte(head), nil, 1)
				restored = bytes.Replace(restored, []byte(composed), []byte(themePlaceholder), 1)
				if !bytes.Equal(restored, original) {
					t.Errorf("%s served for %q, with the head block removed and %q swapped back for the placeholder, is not byte-equal to the embedded original", doc.name, palette, composed)
				}

				gotScripts := extractScriptBodies(served)
				if len(gotScripts) != len(wantScripts) {
					t.Fatalf("%s served for %q carries %d inline scripts, want %d", doc.name, palette, len(gotScripts), len(wantScripts))
				}
				for i := range wantScripts {
					if gotScripts[i] != wantScripts[i] {
						t.Errorf("%s served for %q script[%d] = %q, want %q", doc.name, palette, i, gotScripts[i], wantScripts[i])
					}
				}
			})
		}
	}
}

// newCapturedWebHandler builds a real [serve.Shell] mounting
// [newWebRoutes]'s login and _app routes over the real embedded
// bundle, with logger writing to a fresh, returned buffer, so a test
// can assert on what the shell logs about a request as well as on the
// response it answers.
func newCapturedWebHandler(t *testing.T, environ []string, prefix serve.Prefix) (http.Handler, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	routes, err := newWebRoutes(environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger())
	if err != nil {
		t.Fatalf("newWebRoutes(%v): %v", environ, err)
	}
	strip := serve.NewStrip(serve.TrustedProxyConfig{})
	shell := serve.NewShell(prefix, strip, logger)
	routes.register(shell)
	return shell.Handler(), &buf
}

// TestNewWebRoutes_RealBundleThemeStamp proves the stamp deliverable
// end-to-end against the real built bundle: the extracted stampable
// set is exactly {label, label-night, system} before any cookie row
// runs, a no-cookie request (and the empty-match-key resolution path)
// stamps the operator's default, and every cookie row from the plan's
// States table stamps its stated composed name, with a body byte-equal
// to a startup variant, no Set-Cookie, and no log record holding the
// cookie's value.
func TestNewWebRoutes_RealBundleThemeStamp(t *testing.T) {
	prefix := webTestPrefix(t)

	// The mutation this test guards against can write the cookie's
	// value through either sink a handler in this chain could reach:
	// the shell's own logger (captured per request below) or the
	// package-wide slog.Default. captureDefaultLogger swaps the latter
	// for the run and restores it on cleanup, so this test must not run
	// in parallel with another that reads or writes slog.Default.
	defaultLogs := captureDefaultLogger(t)

	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	loginRaw, err := fs.ReadFile(buildFS, loginDocumentName)
	if err != nil {
		t.Fatalf("fs.ReadFile(%q): %v", loginDocumentName, err)
	}
	linkRaw, err := fs.ReadFile(buildFS, loginLinkDocumentName)
	if err != nil {
		t.Fatalf("fs.ReadFile(%q): %v", loginLinkDocumentName, err)
	}
	themes, err := extractThemes(buildFS,
		namedDocument{path: loginDocumentName, raw: loginRaw},
		namedDocument{path: loginLinkDocumentName, raw: linkRaw},
	)
	if err != nil {
		t.Fatalf("extractThemes: %v", err)
	}
	gotPalettes := stampablePalettes(themes.selectors, acceptedThemeMatch)
	wantPalettes := []string{"label", "label-night", "system"}
	if !slices.Equal(gotPalettes, wantPalettes) {
		t.Fatalf("stampablePalettes(real bundle) = %v, want %v", gotPalettes, wantPalettes)
	}

	// The no-cookie row, proved under both the unset key and the
	// empty-string key: R-5cii12's wall convention resolves both to the
	// same accepted default.
	for _, environ := range [][]string{nil, {"DUBPLATE_THEME_MATCH="}} {
		handler, _ := newCapturedWebHandler(t, environ, prefix)
		r := httptest.NewRequest(http.MethodGet, "/dubplate/login", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, r)
		if rec.Code != http.StatusOK {
			t.Fatalf("environ %v: status = %d, want 200", environ, rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `data-theme="dubplate-system"`) {
			t.Errorf("environ %v: body = %q, want it to carry data-theme=%q", environ, rec.Body.String(), "dubplate-system")
		}
	}

	tests := []struct {
		name         string
		paletteEnv   string // DUBPLATE_THEME_PALETTE; "" leaves the key unset (default "system")
		cookieHeader string
		wantComposed string
	}{
		{name: "absent cookie, default keys", wantComposed: "dubplate-system"},
		{name: "cookie label", cookieHeader: "__Host-dubplate_theme=label", wantComposed: "dubplate-label"},
		{name: "cookie label-night overrides a palette key pinned to label", paletteEnv: "label", cookieHeader: "__Host-dubplate_theme=label-night", wantComposed: "dubplate-label-night"},
		{name: "cookie system overrides a palette key pinned to label-night", paletteEnv: "label-night", cookieHeader: "__Host-dubplate_theme=system", wantComposed: "dubplate-system"},
		{name: "cookie value enclosed in double quotes", cookieHeader: `__Host-dubplate_theme="label"`, wantComposed: "dubplate-label"},
		{name: "unknown palette falls back to default", cookieHeader: "__Host-dubplate_theme=gruvbox", wantComposed: "dubplate-system"},
		{name: "a prefix of a real palette never matches", cookieHeader: "__Host-dubplate_theme=label-n", wantComposed: "dubplate-system"},
		{name: "another case never matches", cookieHeader: "__Host-dubplate_theme=Label", wantComposed: "dubplate-system"},
		{name: "an interior space never matches", cookieHeader: "__Host-dubplate_theme=label night", wantComposed: "dubplate-system"},
		{name: "empty value falls back to default", cookieHeader: "__Host-dubplate_theme=", wantComposed: "dubplate-system"},
		{name: "a backslash byte the parser refuses", cookieHeader: `__Host-dubplate_theme=a\b`, wantComposed: "dubplate-system"},
		{name: "an interior double quote", cookieHeader: `__Host-dubplate_theme=la"bel`, wantComposed: "dubplate-system"},
		{name: "a control byte", cookieHeader: "__Host-dubplate_theme=la\x01bel", wantComposed: "dubplate-system"},
		{name: "a non-ASCII byte", cookieHeader: "__Host-dubplate_theme=la\x80bel", wantComposed: "dubplate-system"},
		{name: "two cookies of the name: only the first parsed one is read", cookieHeader: "__Host-dubplate_theme=gruvbox; __Host-dubplate_theme=label-night", wantComposed: "dubplate-system"},
		{name: "the same value under another cookie name is ignored", cookieHeader: "dubplate_theme=label-night", wantComposed: "dubplate-system"},
		{name: "two cookies, the first unparseable: the second is read", cookieHeader: `__Host-dubplate_theme=a\b; __Host-dubplate_theme=label-night`, wantComposed: "dubplate-label-night"},
		{name: "a padded value keeps the space inside, so it never matches", cookieHeader: "__Host-dubplate_theme= label", wantComposed: "dubplate-system"},
		{name: "an unknown value carrying a planted marker never matches, and the marker reaches no log", cookieHeader: "__Host-dubplate_theme=qx7-plantedmarker-4kd9", wantComposed: "dubplate-system"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultLogs.Reset()
			var environ []string
			if tt.paletteEnv != "" {
				environ = []string{"DUBPLATE_THEME_PALETTE=" + tt.paletteEnv}
			}
			handler, logBuf := newCapturedWebHandler(t, environ, prefix)

			r := httptest.NewRequest(http.MethodGet, "/dubplate/login", nil)
			if tt.cookieHeader != "" {
				r.Header.Set("Cookie", tt.cookieHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r)
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}

			// The startup variant for the row's own composed name: a
			// request whose cookie is exactly that palette name always
			// selects the same variant map entry NewDocument answers
			// this row from.
			wantPalette := strings.TrimPrefix(tt.wantComposed, "dubplate-")
			variantReq := httptest.NewRequest(http.MethodGet, "/dubplate/login", nil)
			variantReq.Header.Set("Cookie", serve.ThemeCookieName+"="+wantPalette)
			variantRec := httptest.NewRecorder()
			handler.ServeHTTP(variantRec, variantReq)
			if rec.Body.String() != variantRec.Body.String() {
				t.Errorf("body is not byte-equal to the %q startup variant", tt.wantComposed)
			}
			if !strings.Contains(rec.Body.String(), `data-theme="`+tt.wantComposed+`"`) {
				t.Errorf("body = %q, want it to carry data-theme=%q", rec.Body.String(), tt.wantComposed)
			}
			if set := rec.Header().Values("Set-Cookie"); len(set) != 0 {
				t.Errorf("Set-Cookie = %v, want none", set)
			}
			// The shell's logger holds one document-request record
			// per request this row sent, and each must be exactly
			// that record: any other content, not merely the cookie
			// header as a substring, is the row logging something it
			// must not.
			const wantRecord = `level=INFO msg="serve: document request" method=GET path=/dubplate/login status=200`
			lines := strings.Split(strings.TrimSuffix(logBuf.String(), "\n"), "\n")
			if len(lines) != 2 {
				t.Errorf("shell log holds %d lines, want 2 (one per request): %s", len(lines), logBuf.String())
			}
			for _, line := range lines {
				_, record, _ := strings.Cut(line, " ")
				if record != wantRecord {
					t.Errorf("shell log line = %q, want its record to be exactly %q", line, wantRecord)
				}
			}
			// NewDocument reads no logger, so nothing on this path
			// should reach slog.Default either; a mutant that logs
			// cookie.Value through it makes this buffer non-empty.
			if defaultLogs.Len() != 0 {
				t.Errorf("slog.Default buffer is not empty: %s", defaultLogs.String())
			}
		})
	}
}

// hrefOrSrcRE finds every href or src attribute value on any tag, so
// the prefix proof checks every reference a document carries rather
// than one tag shape.
var hrefOrSrcRE = regexp.MustCompile(`(?i)\b(?:href|src)\s*=\s*"([^"]*)"`)

func extractHrefsAndSrcs(doc []byte) []string {
	var refs []string
	for _, m := range hrefOrSrcRE.FindAllSubmatch(doc, -1) {
		refs = append(refs, string(m[1]))
	}
	return refs
}

// resolveAgainstDocument resolves ref, a reference a document at
// docPath carries, the way a browser would: relative to docPath's own
// URL, per RFC 3986.
func resolveAgainstDocument(docPath, ref string) (string, error) {
	base := &url.URL{Path: docPath}
	rel, err := url.Parse(ref)
	if err != nil {
		return "", err
	}
	return base.ResolveReference(rel).Path, nil
}

// TestPrefixProof_EveryHrefAndSrcResolves is the gate-resident half of
// R-5cii2's own proof: with the prefix set to a non-default,
// two-segment value the build never saw, every href and src the served
// login documents carry, the two tier stylesheet links included,
// resolves, relative to that document's own URL, to a 200 through the
// real [serve.Shell] composition -- never a plain http.ServeMux, so the origin check and path normalization sit in the
// path too. It runs against the real embedded bundle
// ([newWebRoutes]/[web.Assets]), the fixture state the acceptance
// criterion names.
func TestPrefixProof_EveryHrefAndSrcResolves(t *testing.T) {
	const prefixValue = "/p/q"
	prefix, err := serve.NewPrefix(prefixValue)
	if err != nil {
		t.Fatalf("serve.NewPrefix: %v", err)
	}

	environ := []string{"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, []byte("body{}"))}
	routes, err := newWebRoutes(environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger())
	if err != nil {
		t.Fatalf("newWebRoutes: %v", err)
	}

	strip := serve.NewStrip(serve.TrustedProxyConfig{})
	shell := serve.NewShell(prefix, strip, discardTestLogger())
	routes.register(shell)
	handler := shell.Handler()

	documents := []string{prefixValue + "/login", prefixValue + "/login/link"}
	for _, docPath := range documents {
		t.Run(docPath, func(t *testing.T) {
			rec := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, docPath, nil)
			handler.ServeHTTP(rec, r)
			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s: status %d", docPath, rec.Code)
			}

			refs := extractHrefsAndSrcs(rec.Body.Bytes())
			if len(refs) == 0 {
				t.Fatalf("no href or src found in %s; the fixture proves nothing", docPath)
			}
			var resolvedTiers []string
			for _, ref := range refs {
				if resolved, err := resolveAgainstDocument(docPath, ref); err == nil && (strings.HasSuffix(resolved, "/branding.css") || strings.HasSuffix(resolved, "/operator.css")) {
					resolvedTiers = append(resolvedTiers, resolved)
				}
			}
			if want := []string{prefixValue + "/branding.css", prefixValue + "/operator.css"}; !slices.Equal(resolvedTiers, want) {
				t.Fatalf("tier references in %s resolve to %v, want %v", docPath, resolvedTiers, want)
			}
			for _, ref := range refs {
				resolved, err := resolveAgainstDocument(docPath, ref)
				if err != nil {
					t.Errorf("resolve %q against %s: %v", ref, docPath, err)
					continue
				}
				refRec := httptest.NewRecorder()
				refReq := httptest.NewRequest(http.MethodGet, resolved, nil)
				handler.ServeHTTP(refRec, refReq)
				if refRec.Code != http.StatusOK {
					t.Errorf("GET %s (from %s's reference %q): status %d, want 200", resolved, docPath, ref, refRec.Code)
				}
			}
		})
	}
}

// TestNewServeShell_RegistersWebRoutes proves newServeShell mounts the
// real web routes rather than leaving them to a caller: they answer
// through the shell built by the production wiring, using the default
// prefix and the real embedded bundle.
func TestNewServeShell_RegistersWebRoutes(t *testing.T) {
	environ := serveAuthEnviron(t)
	shell, _, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell: %v", err)
	}

	tests := []struct {
		path string
		want int
	}{
		{path: "/dubplate/login", want: http.StatusOK},
		{path: "/dubplate/login/link", want: http.StatusOK},
		{path: "/dubplate/branding.css", want: http.StatusOK},
		{path: "/dubplate/fonts/AlfaSlabOne-Regular.woff2", want: http.StatusOK},
		{path: "/dubplate/fonts/WorkSans-Variable.woff2", want: http.StatusOK},
		{path: "/dubplate/fonts/IBMPlexMono-Regular.woff2", want: http.StatusOK},
		{path: "/dubplate/fonts/FONTS.md", want: http.StatusNotFound},
		{path: "/dubplate/favicon.svg", want: http.StatusOK},
		{path: "/dubplate/favicon.png", want: http.StatusOK},
		{path: "/dubplate/mark/mark-large.svg", want: http.StatusOK},
		// The operator key defaults to "none", so nothing is mounted.
		{path: "/dubplate/operator.css", want: http.StatusNotFound},
	}
	for _, tt := range tests {
		r := httptest.NewRequest(http.MethodGet, tt.path, nil)
		rec := httptest.NewRecorder()
		shell.Handler().ServeHTTP(rec, r)
		if rec.Code != tt.want {
			t.Errorf("GET %s: status %d, want %d", tt.path, rec.Code, tt.want)
		}
	}

	// The shell passes the configuration's own site name, baseEnviron's
	// DUBPLATE_SITE_NAME, into the documents.
	rec := httptest.NewRecorder()
	shell.Handler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/dubplate/login", nil))
	if want := `<meta name="dubplate-site-name" content="Test Family">`; strings.Count(rec.Body.String(), want) != 1 {
		t.Errorf("GET /dubplate/login: body carries %q %d times, want 1", want, strings.Count(rec.Body.String(), want))
	}
}

// writeOperatorCSS writes content to a fresh operator.css under
// t.TempDir() and returns its path.
func writeOperatorCSS(t *testing.T, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "operator.css")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatalf("write operator CSS fixture: %v", err)
	}
	return path
}

// writeRosterFixture writes content to a fresh roster.json under
// t.TempDir() and returns its path.
func writeRosterFixture(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "roster.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write roster fixture: %v", err)
	}
	return path
}

// symlinkFixture creates link pointing at target and returns link.
func symlinkFixture(t *testing.T, target, link string) string {
	t.Helper()
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("symlink fixture: %v", err)
	}
	return link
}

// newWebShellHandler builds a real [serve.Shell] mounting every route
// [newWebRoutesFromFS] builds from buildFS under the default prefix.
func newWebShellHandler(t *testing.T, environ []string, buildFS fs.FS) http.Handler {
	t.Helper()
	prefix := webTestPrefix(t)
	routes, err := newWebRoutesFromFS(environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), buildFS)
	if err != nil {
		t.Fatalf("newWebRoutesFromFS(%v): %v", environ, err)
	}
	strip := serve.NewStrip(serve.TrustedProxyConfig{})
	shell := serve.NewShell(prefix, strip, discardTestLogger())
	routes.register(shell)
	return shell.Handler()
}

// serveRequest answers method on path through handler and returns
// the recorded response. The request carries Sec-Fetch-Site:
// same-origin, as a browser's own does, so the shell's origin check
// admits an unsafe method and the method rule behind it is what
// answers.
func serveRequest(handler http.Handler, method, path string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, nil)
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	return rec
}

// TestNewWebRoutesFromFS_BrandingRefusals proves every refused branding
// value stops startup with an error naming DUBPLATE_BRANDING and never
// echoing the value. Each bad declaration sits beside a valid one, so
// the refusal is the bad value's.
func TestNewWebRoutesFromFS_BrandingRefusals(t *testing.T) {
	prefix := webTestPrefix(t)
	tests := []struct {
		name    string
		bad     string
		refused string
	}{
		{name: "closing brace", bad: "--color-primary: #fff}", refused: "#fff}"},
		{name: "opening brace", bad: "--color-primary: #fff{", refused: "#fff{"},
		{name: "semicolon inside a value", bad: "--color-primary: #fff;color:red", refused: "#fff;color:red"},
		{name: "url()", bad: "--color-primary: url(https://example.invalid/x)", refused: "url(https://example.invalid/x)"},
		{name: "var()", bad: "--color-primary: var(--x)", refused: "var(--x)"},
		{name: "expression()", bad: "--color-primary: expression(alert(1))", refused: "expression(alert(1))"},
		{name: "closing style tag", bad: "--color-primary: </style>", refused: "</style>"},
		{name: "backslash escape", bad: `--color-primary: #f\66f`, refused: `#f\66f`},
		{name: "control character", bad: "--color-primary: #ff\x01f", refused: "#ff\x01f"},
		{name: "65-byte value", bad: "--color-primary: oklch(45." + strings.Repeat("0", 47) + "% 0.24 1)", refused: "oklch(45." + strings.Repeat("0", 47)},
		{name: "name outside the set", bad: "--tab-radius: 1rem", refused: "1rem"},
		{name: "repeated name", bad: "--color-base-100: #000000", refused: "#000000"},
		{name: "doubled semicolon", bad: "; --depth: 1", refused: "--depth: 1"},
		{name: "no colon", bad: "--depth 1", refused: "--depth 1"},
		{name: "value outside its grammar", bad: "--depth: yes", refused: "yes"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := []string{"DUBPLATE_BRANDING=--color-base-100: #ffffff; " + tt.bad}
			_, err := newWebRoutesFromFS(environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), newFixtureBuildFS(1, 1))
			if err == nil {
				t.Fatal("newWebRoutesFromFS: got nil error, want the branding refusal")
			}
			if !strings.HasPrefix(err.Error(), "DUBPLATE_BRANDING: ") {
				t.Errorf("error = %q, want it to name DUBPLATE_BRANDING", err.Error())
			}
			if strings.Contains(err.Error(), tt.refused) {
				t.Errorf("error = %q echoes the refused value %q", err.Error(), tt.refused)
			}
		})
	}
}

// TestNewWebRoutesFromFS_BrandingStylesheet proves the route serves
// the rendered stylesheet byte for byte, and an empty one when the key
// is unset, empty, or "none".
func TestNewWebRoutesFromFS_BrandingStylesheet(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    string
	}{
		{name: "key unset", want: ""},
		{name: "key empty", environ: []string{"DUBPLATE_BRANDING="}, want: ""},
		{name: "key none", environ: []string{"DUBPLATE_BRANDING=none"}, want: ""},
		{
			name:    "several pairs",
			environ: []string{"DUBPLATE_BRANDING=--color-primary: #3b5bdb; --radius-box: 0.75rem;--depth:0"},
			want:    ":root {\n\t--color-primary: #3b5bdb;\n\t--radius-box: 0.75rem;\n\t--depth: 0;\n}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newWebShellHandler(t, tt.environ, newFixtureBuildFS(1, 1))
			rec := serveRequest(handler, http.MethodGet, "/dubplate/branding.css")
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
				t.Errorf("Content-Type = %q, want %q", got, "text/css; charset=utf-8")
			}
			if got := rec.Body.String(); got != tt.want {
				t.Errorf("body = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestNewWebRoutesFromFS_OperatorCSSRefusals proves each file state
// the operator CSS rule refuses stops startup, naming
// DUBPLATE_OPERATOR_CSS_FILE and the rule, never the path or a byte of
// the file. The file content carries a marker so an error echoing it
// is caught.
func TestNewWebRoutesFromFS_OperatorCSSRefusals(t *testing.T) {
	prefix := webTestPrefix(t)
	const marker = "qx7operatormarker"
	tests := []struct {
		name       string
		path       func(t *testing.T, private *privateFiles) string
		wantErrHas string
	}{
		{
			name: "path not ending in .css",
			path: func(t *testing.T, _ *privateFiles) string {
				p := filepath.Join(t.TempDir(), "operator.json")
				if err := os.WriteFile(p, []byte(marker), 0o600); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErrHas: `the path must end in ".css"`,
		},
		{
			name:       "path ending in .CSS",
			path:       func(t *testing.T, _ *privateFiles) string { return filepath.Join(t.TempDir(), "operator.CSS") },
			wantErrHas: `the path must end in ".css"`,
		},
		{
			name:       "path absent",
			path:       func(t *testing.T, _ *privateFiles) string { return filepath.Join(t.TempDir(), "missing.css") },
			wantErrHas: "cannot read the file: no such file or directory",
		},
		{
			name: "not a regular file",
			path: func(t *testing.T, _ *privateFiles) string {
				p := filepath.Join(t.TempDir(), "dir.css")
				if err := os.Mkdir(p, 0o700); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErrHas: "the path is not a regular file",
		},
		{
			name: "unreadable",
			path: func(t *testing.T, _ *privateFiles) string {
				if os.Geteuid() == 0 {
					t.Skip("root reads a mode 0000 file")
				}
				p := writeOperatorCSS(t, []byte(marker))
				if err := os.Chmod(p, 0); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErrHas: "cannot read the file: permission denied",
		},
		{
			name: "one byte over the bound",
			path: func(t *testing.T, _ *privateFiles) string {
				return writeOperatorCSS(t, append([]byte(marker), bytes.Repeat([]byte("a"), 262144+1-len(marker))...))
			},
			wantErrHas: "the file is over 262144 bytes",
		},
		{
			name:       "invalid UTF-8",
			path:       func(t *testing.T, _ *privateFiles) string { return writeOperatorCSS(t, []byte(marker+"\xff\xfe")) },
			wantErrHas: "the file is not valid UTF-8",
		},
		{
			name:       "a NUL byte",
			path:       func(t *testing.T, _ *privateFiles) string { return writeOperatorCSS(t, []byte(marker+"\x00body{}")) },
			wantErrHas: "the file holds a NUL byte",
		},
		{
			name: "a .css symlink to the roster",
			path: func(t *testing.T, private *privateFiles) string {
				private.roster = writeRosterFixture(t, marker)
				return symlinkFixture(t, private.roster, filepath.Join(t.TempDir(), "operator.css"))
			},
			wantErrHas: `the path must resolve to a file ending in ".css"`,
		},
		{
			name: "a .css hard link to the roster",
			path: func(t *testing.T, private *privateFiles) string {
				private.roster = writeRosterFixture(t, marker)
				p := filepath.Join(t.TempDir(), "operator.css")
				if err := os.Link(private.roster, p); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErrHas: "the file is the roster",
		},
		{
			name: "a parent-directory symlink into the state directory",
			path: func(t *testing.T, private *privateFiles) string {
				private.stateDir = t.TempDir()
				if err := os.WriteFile(filepath.Join(private.stateDir, "operator.css"), []byte(marker), 0o600); err != nil {
					t.Fatal(err)
				}
				dir := symlinkFixture(t, private.stateDir, filepath.Join(t.TempDir(), "styles"))
				return filepath.Join(dir, "operator.css")
			},
			wantErrHas: "the file is under the state directory",
		},
		{
			name: "a FIFO",
			path: func(t *testing.T, _ *privateFiles) string {
				p := filepath.Join(t.TempDir(), "operator.css")
				if err := syscall.Mkfifo(p, 0o600); err != nil {
					t.Fatal(err)
				}
				return p
			},
			wantErrHas: "the path is not a regular file",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var private privateFiles
			path := tt.path(t, &private)
			_, err := newWebRoutesFromFS([]string{"DUBPLATE_OPERATOR_CSS_FILE=" + path}, prefix, fixtureSiteName, private, discardTestLogger(), newFixtureBuildFS(1, 1))
			if err == nil {
				t.Fatalf("newWebRoutesFromFS: got nil error, want one containing %q", tt.wantErrHas)
			}
			if want := "DUBPLATE_OPERATOR_CSS_FILE: " + tt.wantErrHas; !strings.Contains(err.Error(), want) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), want)
			}
			if strings.Contains(err.Error(), path) || strings.Contains(err.Error(), marker) {
				t.Errorf("error = %q echoes the path or the file's bytes", err.Error())
			}
		})
	}
}

// TestNewWebRoutesFromFS_OperatorCSSServed proves the admitted states:
// a file at the bound and an empty file are each served as read, with
// the tier 3 link in place, and the key unset, empty, or "none" links
// nothing and leaves the route to the terminator's 404.
func TestNewWebRoutesFromFS_OperatorCSSServed(t *testing.T) {
	atBound := bytes.Repeat([]byte("a"), 262144)
	tests := []struct {
		name     string
		environ  func(t *testing.T) []string
		wantCode int
		wantBody []byte
		wantLink bool
	}{
		{name: "key unset", environ: func(*testing.T) []string { return nil }, wantCode: http.StatusNotFound},
		{name: "key empty", environ: func(*testing.T) []string { return []string{"DUBPLATE_OPERATOR_CSS_FILE="} }, wantCode: http.StatusNotFound},
		{name: "key none", environ: func(*testing.T) []string { return []string{"DUBPLATE_OPERATOR_CSS_FILE=none"} }, wantCode: http.StatusNotFound},
		{
			name: "file empty",
			environ: func(t *testing.T) []string {
				return []string{"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, nil)}
			},
			wantCode: http.StatusOK, wantBody: []byte{}, wantLink: true,
		},
		{
			name: "file at the bound",
			environ: func(t *testing.T) []string {
				return []string{"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, atBound)}
			},
			wantCode: http.StatusOK, wantBody: atBound, wantLink: true,
		},
		{
			name: "a .css symlink to a .css file",
			environ: func(t *testing.T) []string {
				target := writeOperatorCSS(t, []byte("body { margin: 0 }\n"))
				return []string{"DUBPLATE_OPERATOR_CSS_FILE=" + symlinkFixture(t, target, filepath.Join(t.TempDir(), "linked.css"))}
			},
			wantCode: http.StatusOK, wantBody: []byte("body { margin: 0 }\n"), wantLink: true,
		},
		{
			name: "file with multibyte text",
			environ: func(t *testing.T) []string {
				return []string{"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, []byte("/* café */ body { margin: 0 }\n"))}
			},
			wantCode: http.StatusOK, wantBody: []byte("/* café */ body { margin: 0 }\n"), wantLink: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newWebShellHandler(t, tt.environ(t), newFixtureBuildFS(1, 1))
			rec := serveRequest(handler, http.MethodGet, "/dubplate/operator.css")
			if rec.Code != tt.wantCode {
				t.Fatalf("GET operator.css: status = %d, want %d", rec.Code, tt.wantCode)
			}
			if tt.wantCode == http.StatusOK {
				if !bytes.Equal(rec.Body.Bytes(), tt.wantBody) {
					t.Errorf("GET operator.css: body is %d bytes, want the file's %d bytes", rec.Body.Len(), len(tt.wantBody))
				}
				if got := rec.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
					t.Errorf("Content-Type = %q, want %q", got, "text/css; charset=utf-8")
				}
			}
			doc := serveRequest(handler, http.MethodGet, "/dubplate/login").Body.Bytes()
			if got := slices.Contains(serve.StylesheetHrefs(doc), "./operator.css"); got != tt.wantLink {
				t.Errorf("login document links ./operator.css = %v, want %v", got, tt.wantLink)
			}
		})
	}
}

// TestNewWebRoutesFromFS_OperatorCSSReadOnce proves the file is read
// at startup only: bytes written after the routes are built never
// reach a response.
func TestNewWebRoutesFromFS_OperatorCSSReadOnce(t *testing.T) {
	path := writeOperatorCSS(t, []byte("body { color: black }\n"))
	handler := newWebShellHandler(t, []string{"DUBPLATE_OPERATOR_CSS_FILE=" + path}, newFixtureBuildFS(1, 1))
	if err := os.WriteFile(path, []byte("body { color: red }\n"), 0o600); err != nil {
		t.Fatalf("rewrite operator CSS: %v", err)
	}
	rec := serveRequest(handler, http.MethodGet, "/dubplate/operator.css")
	if got, want := rec.Body.String(), "body { color: black }\n"; got != want {
		t.Errorf("body = %q, want the startup bytes %q", got, want)
	}
}

// TestStylesheetRoutes_ExactPaths proves both stylesheet routes answer
// their one path and read nothing else from the request, through the
// real shell so path normalization runs first.
func TestStylesheetRoutes_ExactPaths(t *testing.T) {
	const operatorBody = "body { margin: 0 }\n"
	environ := []string{
		"DUBPLATE_BRANDING=--depth: 1",
		"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, []byte(operatorBody)),
	}
	handler := newWebShellHandler(t, environ, newFixtureBuildFS(1, 1))

	tests := []struct {
		name      string
		method    string
		path      string
		wantCode  int
		wantBody  string
		wantAllow string
	}{
		{name: "operator deeper path", method: http.MethodGet, path: "/dubplate/operator.css/x", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "branding deeper path", method: http.MethodGet, path: "/dubplate/branding.css/x", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "operator trailing slash", method: http.MethodGet, path: "/dubplate/operator.css/", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "a query names another file", method: http.MethodGet, path: "/dubplate/operator.css?path=/etc/passwd", wantCode: http.StatusOK, wantBody: operatorBody},
		{name: "a dot-dot segment is the normalizer's", method: http.MethodGet, path: "/dubplate/operator.css/../x", wantCode: http.StatusBadRequest},
		{name: "operator HEAD", method: http.MethodHead, path: "/dubplate/operator.css", wantCode: http.StatusOK},
		{name: "branding HEAD", method: http.MethodHead, path: "/dubplate/branding.css", wantCode: http.StatusOK},
		{name: "operator POST", method: http.MethodPost, path: "/dubplate/operator.css", wantCode: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
		{name: "branding POST", method: http.MethodPost, path: "/dubplate/branding.css", wantCode: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveRequest(handler, tt.method, tt.path)
			if rec.Code != tt.wantCode {
				t.Fatalf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.wantCode)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("%s %s: body = %q, want %q", tt.method, tt.path, rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("%s %s: Allow = %q, want %q", tt.method, tt.path, got, tt.wantAllow)
			}
		})
	}
}

// fontsFixtureBuildFS extends newFixtureBuildFS's own build tree with
// two named faces, a manifest, and a license text, so the fonts
// route's every state has something to answer: a served face, a
// sibling face, a directory, a top-level non-.woff2 file, and a name
// it holds none of.
func fontsFixtureBuildFS() fstest.MapFS {
	buildFS := newFixtureBuildFS(1, 1)
	delete(buildFS, "fonts/fixture.woff2")
	buildFS["fonts/a.woff2"] = &fstest.MapFile{Data: []byte("face-a-bytes")}
	buildFS["fonts/b.woff2"] = &fstest.MapFile{Data: []byte("face-b-bytes")}
	buildFS["fonts/FONTS.md"] = &fstest.MapFile{Data: []byte("# manifest")}
	buildFS["fonts/licenses/a-OFL.txt"] = &fstest.MapFile{Data: []byte("license text")}
	return buildFS
}

// fontsFaceETag returns the strong ETag [newFontsRoute] computes for
// data: a quoted hex SHA-256, independent of that function's own
// hashing so a change to one cannot go unnoticed by the other.
func fontsFaceETag(data []byte) string {
	sum := sha256.Sum256(data)
	return `"` + hex.EncodeToString(sum[:]) + `"`
}

// TestFontsRoute_States is the fonts route's own state table: each
// served face, every 404 shape (a directory, a license text,
// FONTS.md, a missing name, a trailing slash, and the bare subtree
// with no name), and the 405.
func TestFontsRoute_States(t *testing.T) {
	handler := newWebShellHandler(t, nil, fontsFixtureBuildFS())

	tests := []struct {
		name      string
		method    string
		path      string
		wantCode  int
		wantBody  string
		wantAllow string
	}{
		{name: "GET a face", method: http.MethodGet, path: "/dubplate/fonts/a.woff2", wantCode: http.StatusOK, wantBody: "face-a-bytes"},
		{name: "GET the sibling face", method: http.MethodGet, path: "/dubplate/fonts/b.woff2", wantCode: http.StatusOK, wantBody: "face-b-bytes"},
		{name: "HEAD a face", method: http.MethodHead, path: "/dubplate/fonts/a.woff2", wantCode: http.StatusOK},
		{name: "a missing name", method: http.MethodGet, path: "/dubplate/fonts/c.woff2", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "FONTS.md", method: http.MethodGet, path: "/dubplate/fonts/FONTS.md", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "a license text", method: http.MethodGet, path: "/dubplate/fonts/licenses/a-OFL.txt", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "the licenses directory itself", method: http.MethodGet, path: "/dubplate/fonts/licenses", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "a trailing slash", method: http.MethodGet, path: "/dubplate/fonts/a.woff2/", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "the bare subtree, no name", method: http.MethodGet, path: "/dubplate/fonts", wantCode: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "POST", method: http.MethodPost, path: "/dubplate/fonts/a.woff2", wantCode: http.StatusMethodNotAllowed, wantAllow: "GET, HEAD"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serveRequest(handler, tt.method, tt.path)
			if rec.Code != tt.wantCode {
				t.Fatalf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.wantCode)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("%s %s: body = %q, want %q", tt.method, tt.path, rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Allow"); got != tt.wantAllow {
				t.Errorf("%s %s: Allow = %q, want %q", tt.method, tt.path, got, tt.wantAllow)
			}
		})
	}
}

// TestFontsRoute_Headers proves a served face carries three headers:
// an explicit Content-Type, Cache-Control: no-cache, and a strong
// ETag matching the file's own SHA-256, and that a matching
// If-None-Match answers 304 (go doc net/http.ServeContent).
func TestFontsRoute_Headers(t *testing.T) {
	buildFS := fontsFixtureBuildFS()
	handler := newWebShellHandler(t, nil, buildFS)
	wantETag := fontsFaceETag(buildFS["fonts/a.woff2"].Data)

	rec := serveRequest(handler, http.MethodGet, "/dubplate/fonts/a.woff2")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET a.woff2: status = %d, want 200", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "font/woff2" {
		t.Errorf("Content-Type = %q, want %q", got, "font/woff2")
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
		t.Errorf("Cache-Control = %q, want %q", got, "no-cache")
	}
	if got := rec.Header().Get("ETag"); got != wantETag {
		t.Errorf("ETag = %q, want %q", got, wantETag)
	}

	r := httptest.NewRequest(http.MethodGet, "/dubplate/fonts/a.woff2", nil)
	r.Header.Set("Sec-Fetch-Site", "same-origin")
	r.Header.Set("If-None-Match", wantETag)
	revalidate := httptest.NewRecorder()
	handler.ServeHTTP(revalidate, r)
	if revalidate.Code != http.StatusNotModified {
		t.Errorf("GET a.woff2 with a matching If-None-Match: status = %d, want %d", revalidate.Code, http.StatusNotModified)
	}
}

// TestFontsRoute_ContentTypeIgnoresHostMimeTable proves the served
// Content-Type comes from the route itself, never from the host's own
// extension table: [mime.AddExtensionType] seeds ".woff2" to a wrong
// value first, and the seed is asserted in force before the request
// runs.
func TestFontsRoute_ContentTypeIgnoresHostMimeTable(t *testing.T) {
	if err := mime.AddExtensionType(".woff2", "application/x-seeded"); err != nil {
		t.Fatalf("mime.AddExtensionType: %v", err)
	}
	if got := mime.TypeByExtension(".woff2"); got != "application/x-seeded" {
		t.Fatalf("mime.TypeByExtension(%q) = %q, want the seed in force before the request", ".woff2", got)
	}

	handler := newWebShellHandler(t, nil, fontsFixtureBuildFS())
	rec := serveRequest(handler, http.MethodGet, "/dubplate/fonts/a.woff2")
	if got := rec.Header().Get("Content-Type"); got != "font/woff2" {
		t.Errorf("Content-Type = %q, want %q even with a seeded host mapping in force", got, "font/woff2")
	}
}

// TestNewWebRoutesFromFS_NoFontFilesRefusesStartup proves a build
// tree carrying no .woff2 file under "fonts/" refuses startup, the
// same shape a placeholder-only build refuses on, rather than
// registering a fonts route that answers every path with 404.
func TestNewWebRoutesFromFS_NoFontFilesRefusesStartup(t *testing.T) {
	buildFS := newFixtureBuildFS(1, 1)
	delete(buildFS, "fonts/fixture.woff2")
	buildFS["fonts/FONTS.md"] = &fstest.MapFile{Data: []byte("# manifest")}

	_, err := newWebRoutesFromFS(nil, webTestPrefix(t), fixtureSiteName, privateFiles{}, discardTestLogger(), buildFS)
	if err == nil {
		t.Fatal("newWebRoutesFromFS: got nil error, want the no-.woff2-file refusal")
	}
	if !strings.Contains(err.Error(), "no .woff2 file") {
		t.Errorf("newWebRoutesFromFS error = %q, want it to mention the missing .woff2 file", err.Error())
	}
}

// TestNewWebRoutes_TierLinkOrder proves, on the real built bundle,
// that each served document links the bundle's stylesheets, then
// branding.css, then operator.css when a file is configured, and that
// the branding link is present when the branding key is "none".
func TestNewWebRoutes_TierLinkOrder(t *testing.T) {
	prefix := webTestPrefix(t)
	operatorPath := writeOperatorCSS(t, []byte("body{}"))
	tests := []struct {
		name     string
		environ  []string
		wantTier map[string][]string // document path -> expected trailing hrefs
	}{
		{
			name:    "operator file configured",
			environ: []string{"DUBPLATE_BRANDING=none", "DUBPLATE_OPERATOR_CSS_FILE=" + operatorPath},
			wantTier: map[string][]string{
				"/dubplate/login":      {"./branding.css", "./operator.css"},
				"/dubplate/login/link": {"../branding.css", "../operator.css"},
			},
		},
		{
			name:    "operator key none",
			environ: []string{"DUBPLATE_BRANDING=none", "DUBPLATE_OPERATOR_CSS_FILE=none"},
			wantTier: map[string][]string{
				"/dubplate/login":      {"./branding.css"},
				"/dubplate/login/link": {"../branding.css"},
			},
		},
	}
	for _, tt := range tests {
		routes, err := newWebRoutes(tt.environ, prefix, fixtureSiteName, privateFiles{}, discardTestLogger())
		if err != nil {
			t.Fatalf("%s: newWebRoutes: %v", tt.name, err)
		}
		for docPath, wantTier := range tt.wantTier {
			t.Run(tt.name+docPath, func(t *testing.T) {
				rec := serveRequest(routes.login, http.MethodGet, docPath)
				hrefs := serve.StylesheetHrefs(rec.Body.Bytes())
				if len(hrefs) <= len(wantTier) {
					t.Fatalf("stylesheet hrefs = %v, want at least one bundle stylesheet before %v", hrefs, wantTier)
				}
				bundle, tier := hrefs[:len(hrefs)-len(wantTier)], hrefs[len(hrefs)-len(wantTier):]
				for _, href := range bundle {
					if !strings.Contains(href, "/_app/immutable/") {
						t.Errorf("stylesheet %q precedes the tier links but is not a bundle stylesheet (hrefs %v)", href, hrefs)
					}
				}
				if !slices.Equal(tier, wantTier) {
					t.Errorf("trailing stylesheet hrefs = %v, want %v (hrefs %v)", tier, wantTier, hrefs)
				}
			})
		}
	}
}

// TestNewWebRoutesFromFS_TierLinkAnchor proves the anchor table's
// states through the command's own wiring: a document with no
// stylesheet link refuses startup naming it, and a document carrying
// two gets the tier links after the second.
func TestNewWebRoutesFromFS_TierLinkAnchor(t *testing.T) {
	prefix := webTestPrefix(t)

	noLink := newFixtureBuildFS(1, 1)
	noLink["login.html"] = &fstest.MapFile{Data: []byte(`<html data-theme="` + themePlaceholder + `"><head>` + fixtureCSPMeta + `<link href="./_app/entry/start.js" rel="modulepreload"></head></html>`)}
	if _, err := newWebRoutesFromFS(nil, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), noLink); err == nil || !strings.Contains(err.Error(), `document "login.html": no stylesheet link`) {
		t.Errorf("no stylesheet link: error = %v, want one naming login.html", err)
	}

	twoLinks := newFixtureBuildFS(1, 1)
	twoLinks["login.html"] = &fstest.MapFile{Data: []byte(`<html data-theme="` + themePlaceholder + `"><head>` + fixtureCSPMeta + `<link href="./_app/assets/0.css" rel="stylesheet"><link href="./_app/assets/2.css" rel="stylesheet"></head></html>`)}
	twoLinks["_app/assets/2.css"] = &fstest.MapFile{Data: []byte("p{}")}
	handler := newWebShellHandler(t, nil, twoLinks)
	hrefs := serve.StylesheetHrefs(serveRequest(handler, http.MethodGet, "/dubplate/login").Body.Bytes())
	if want := []string{"./_app/assets/0.css", "./_app/assets/2.css", "./branding.css"}; !slices.Equal(hrefs, want) {
		t.Errorf("stylesheet hrefs = %v, want %v", hrefs, want)
	}
}

// TestNewWebRoutesFromFS_VariantCSPRefusals proves startup refuses a
// build whose rewrite would change a hashed inline script: the
// placeholder inside a script body, and the last stylesheet link (the
// tier links' anchor) inside a script body. Each fixture's CSP meta
// hashes the script as built, so the build's own document passes
// [serve.DocumentCSP] and only the served variant can fail it.
func TestNewWebRoutesFromFS_VariantCSPRefusals(t *testing.T) {
	prefix := webTestPrefix(t)
	tests := []struct {
		name   string
		script string
	}{
		{name: "placeholder inside a script", script: `const theme = "` + themePlaceholder + `";`},
		{name: "anchor inside a script", script: `document.head.insertAdjacentHTML("beforeend", '<link href="./_app/assets/0.css" rel="stylesheet">');`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sum := sha256.Sum256([]byte(tt.script))
			meta := `<meta http-equiv="content-security-policy" content="script-src 'self' 'sha256-` + base64.StdEncoding.EncodeToString(sum[:]) + `'">`
			theme := themePlaceholder
			if strings.Contains(tt.script, themePlaceholder) {
				theme = "stock"
			}
			buildFS := newFixtureBuildFS(1, 1)
			buildFS["login.html"] = &fstest.MapFile{Data: []byte(`<html data-theme="` + theme + `"><head>` + meta +
				`<link href="./_app/assets/0.css" rel="stylesheet"><script>` + tt.script + `</script></head></html>`)}
			if _, err := serve.DocumentCSP("login.html", buildFS["login.html"].Data); err != nil {
				t.Fatalf("fixture: the build's own document fails DocumentCSP: %v", err)
			}

			_, err := newWebRoutesFromFS(nil, prefix, fixtureSiteName, privateFiles{}, discardTestLogger(), buildFS)
			if err == nil {
				t.Fatal("newWebRoutesFromFS: got nil error, want the variant's CSP refusal")
			}
			if want := "an inline script's hash is not in the CSP meta's script-src"; !strings.Contains(err.Error(), want) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), want)
			}
		})
	}
}

// TestCheckVariantCSP_RefusesAPolicyUnlikeTheHeader proves a served
// variant must yield exactly the policy the header carries, computed
// from the build's own document: a variant whose meta lists one more
// hash still covers its every script, so only the equality check can
// refuse it.
func TestCheckVariantCSP_RefusesAPolicyUnlikeTheHeader(t *testing.T) {
	source := func(script string) string {
		sum := sha256.Sum256([]byte(script))
		return "'sha256-" + base64.StdEncoding.EncodeToString(sum[:]) + "'"
	}
	doc := func(sources string) []byte {
		return []byte(`<meta http-equiv="content-security-policy" content="script-src 'self' ` + sources + `"><script>boot();</script>`)
	}
	boot := source("boot();")
	served, err := serve.DocumentCSP("login.html", doc(boot))
	if err != nil {
		t.Fatalf("fixture: DocumentCSP: %v", err)
	}
	if err := checkVariantCSP("login.html", map[string][]byte{"label": doc(boot)}, served); err != nil {
		t.Fatalf("checkVariantCSP on a matching variant: %v", err)
	}
	variants := map[string][]byte{"label": doc(boot), "label-night": doc(boot + " " + source("next();"))}
	err = checkVariantCSP("login.html", variants, served)
	if err == nil {
		t.Fatal("checkVariantCSP: got nil error, want the policy mismatch refusal")
	}
	if want := `the "label-night" variant of document "login.html": its CSP differs from the served header`; !strings.Contains(err.Error(), want) {
		t.Errorf("error = %q, want it to contain %q", err.Error(), want)
	}
}

// svelteHeadTitleRE captures the text of a <title> inside a route's
// <svelte:head> block, markup expressions included.
var svelteHeadTitleRE = regexp.MustCompile(`(?s)<svelte:head>.*?<title>(.+?)</title>.*?</svelte:head>`)

// TestRoutes_EveryPageHasAUniqueTitle proves every page and the error
// page in web/src/routes set a document title through <svelte:head>,
// and that no two pages share one, so a screen reader announces which
// page loaded. The build prerenders with ssr off, so the title is set
// only once the client runs and no built document carries it; the
// route source is what this reads.
func TestRoutes_EveryPageHasAUniqueTitle(t *testing.T) {
	root := filepath.Join("..", "..", "web", "src", "routes")
	titles := make(map[string]string)
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if name := d.Name(); name != "+page.svelte" && name != "+error.svelte" {
			return nil
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		m := svelteHeadTitleRE.FindSubmatch(src)
		if m == nil {
			t.Errorf("%s: no <title> inside <svelte:head>", p)
			return nil
		}
		title := strings.TrimSpace(string(m[1]))
		if other, ok := titles[title]; ok {
			t.Errorf("%s: title %q is also %s's", p, title, other)
		}
		titles[title] = p
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
	if len(titles) < 4 {
		t.Errorf("found %d titled routes, want the three pages and the error page", len(titles))
	}
}

// realBuildMapFS copies the real embedded build tree into a fresh
// fstest.MapFS, so a test can edit one file of the real bundle and
// build the routes over the result.
func realBuildMapFS(t *testing.T) fstest.MapFS {
	t.Helper()
	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	out := fstest.MapFS{}
	err = fs.WalkDir(buildFS, ".", func(name string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := fs.ReadFile(buildFS, name)
		if err != nil {
			return err
		}
		out[name] = &fstest.MapFile{Data: data}
		return nil
	})
	if err != nil {
		t.Fatalf("copy the embedded build: %v", err)
	}
	return out
}

// editBundleCSS rewrites the one stylesheet the real bundle's login
// document links with edit, failing t unless the edit changed it.
func editBundleCSS(t *testing.T, buildFS fstest.MapFS, edit func(string) string) {
	t.Helper()
	hrefs := serve.StylesheetHrefs(buildFS[loginDocumentName].Data)
	if len(hrefs) != 1 {
		t.Fatalf("the built login document links %d stylesheets, want 1: %v", len(hrefs), hrefs)
	}
	name := resolveStylesheetPath(loginDocumentName, hrefs[0])
	file, ok := buildFS[name]
	if !ok {
		t.Fatalf("the built bundle holds no %q", name)
	}
	edited := edit(string(file.Data))
	if edited == string(file.Data) {
		t.Fatalf("the edit left %q unchanged; the row proves nothing", name)
	}
	buildFS[name] = &fstest.MapFile{Data: []byte(edited)}
}

// themeColorMetaRE finds every theme-color meta in a served document.
var themeColorMetaRE = regexp.MustCompile(`<meta name="theme-color"[^>]*>`)

// servedVariant answers GET path through handler with the theme cookie
// set to palette and returns the body, failing t on any status but 200.
func servedVariant(t *testing.T, handler http.Handler, path, palette string) string {
	t.Helper()
	r := httptest.NewRequest(http.MethodGet, path, nil)
	r.Header.Set("Cookie", serve.ThemeCookieName+"="+palette)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s as %q: status = %d, want 200", path, palette, rec.Code)
	}
	return rec.Body.String()
}

// TestNewWebRoutes_ThemeColorStates is the theme-color state table
// over the real built bundle. The expected neutrals are the visual
// identity spec's: #1f1b16 for light and #110f0d for dark, as a media
// pair under the system stamp and one value under a pinned palette.
// Each bundle-defect row edits the real stylesheet; each branding row
// sets DUBPLATE_BRANDING. A refusal must name the theme and never the
// value.
func TestNewWebRoutes_ThemeColorStates(t *testing.T) {
	light := `<meta name="theme-color" content="#1f1b16">`
	dark := `<meta name="theme-color" content="#110f0d">`
	pair := []string{
		`<meta name="theme-color" media="(prefers-color-scheme: light)" content="#1f1b16">`,
		`<meta name="theme-color" media="(prefers-color-scheme: dark)" content="#110f0d">`,
	}
	every := func(metas ...string) map[string][]string {
		return map[string][]string{"label": metas, "label-night": metas, "system": metas}
	}

	tests := []struct {
		name        string
		editCSS     func(string) string
		branding    string
		want        map[string][]string
		wantErrHas  string
		wantErrNot  string
		wantLogLine string
	}{
		{
			name: "bundle neutral present",
			want: map[string][]string{"label": {light}, "label-night": {dark}, "system": pair},
		},
		{
			name: "bundle neutral absent for a stampable palette",
			editCSS: func(css string) string {
				return strings.ReplaceAll(css, "--color-neutral:#110f0d", "--color-unrelated:#110f0d")
			},
			wantErrHas: `theme "dubplate-label-night": the built bundle declares no --color-neutral`,
		},
		{
			name: "bundle neutral outside the hex grammar",
			editCSS: func(css string) string {
				return strings.ReplaceAll(css, "--color-neutral:#1f1b16", "--color-neutral:oklch(20% 0 0)")
			},
			wantErrHas: `theme "dubplate-label": its --color-neutral is not a #rgb or #rrggbb color`,
			wantErrNot: "oklch",
		},
		{
			name: "bundle declares two neutrals under one scheme",
			editCSS: func(css string) string {
				return css + `[data-theme=dubplate-label]{color-scheme:light;--color-neutral:#123456}`
			},
			wantErrHas: `theme "dubplate-label": it declares two --color-neutral values under one color scheme`,
		},
		{
			name: "bundle neutral under no scheme beside a light one",
			editCSS: func(css string) string {
				return css + `[data-theme=dubplate-label]{--color-neutral:#123456}`
			},
			wantErrHas: `theme "dubplate-label": its --color-neutral values differ under color schemes other than light and dark`,
		},
		{
			name:     "branding neutral, six-digit hex",
			branding: "--color-neutral: #1f1b16",
			want:     every(`<meta name="theme-color" content="#1f1b16">`),
		},
		{
			name:     "branding neutral, three-digit hex",
			branding: "--color-neutral: #123",
			want:     every(`<meta name="theme-color" content="#123">`),
		},
		{
			name:        "branding neutral, eight-digit hex",
			branding:    "--color-neutral: #1f1b16ff",
			want:        every(),
			wantLogLine: "DUBPLATE_BRANDING sets --color-neutral to a value other than a #rgb or #rrggbb color, so no document carries a theme-color meta",
		},
		{
			name:        "branding neutral, a color function",
			branding:    "--color-neutral: oklch(20% 0 0)",
			want:        every(),
			wantLogLine: "DUBPLATE_BRANDING sets --color-neutral to a value other than a #rgb or #rrggbb color, so no document carries a theme-color meta",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buildFS := realBuildMapFS(t)
			if tt.editCSS != nil {
				editBundleCSS(t, buildFS, tt.editCSS)
			}
			var environ []string
			if tt.branding != "" {
				environ = []string{"DUBPLATE_BRANDING=" + tt.branding}
			}
			var logs bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logs, nil))
			routes, err := newWebRoutesFromFS(environ, webTestPrefix(t), fixtureSiteName, privateFiles{}, logger, buildFS)
			if tt.wantErrHas != "" {
				if err == nil {
					t.Fatalf("newWebRoutesFromFS: got nil error, want one containing %q", tt.wantErrHas)
				}
				if !strings.Contains(err.Error(), tt.wantErrHas) {
					t.Fatalf("newWebRoutesFromFS error = %q, want it to contain %q", err, tt.wantErrHas)
				}
				if tt.wantErrNot != "" && strings.Contains(err.Error(), tt.wantErrNot) {
					t.Fatalf("newWebRoutesFromFS error = %q carries the value %q", err, tt.wantErrNot)
				}
				return
			}
			if err != nil {
				t.Fatalf("newWebRoutesFromFS: %v", err)
			}

			for _, doc := range []struct {
				path    string
				handler http.Handler
			}{
				{"/dubplate/login", routes.login},
				{"/dubplate/login/link", routes.login},
				{"/dubplate/members", routes.members},
			} {
				for palette, want := range tt.want {
					body := servedVariant(t, doc.handler, doc.path, palette)
					if got := themeColorMetaRE.FindAllString(body, -1); !slices.Equal(got, want) {
						t.Errorf("%s as %q: theme-color metas = %q, want %q", doc.path, palette, got, want)
					}
				}
			}

			lines := strings.Count(logs.String(), "theme-color")
			switch {
			case tt.wantLogLine == "" && lines != 0:
				t.Errorf("startup logged a theme-color record, want none:\n%s", logs.String())
			case tt.wantLogLine != "" && (lines != 1 || !strings.Contains(logs.String(), tt.wantLogLine)):
				t.Errorf("startup log = %q, want exactly one record containing %q", logs.String(), tt.wantLogLine)
			}
		})
	}
}

// siteNameMetaRE captures the dubplate-site-name meta's content as
// served, escaped.
var siteNameMetaRE = regexp.MustCompile(`<meta name="dubplate-site-name" content="([^"]*)"`)

// TestNewWebRoutes_SiteNameMeta proves the site name enters every
// variant of every document as exactly one meta whose content, read
// back through html.UnescapeString, is the configured name: go doc
// html.UnescapeString states "UnescapeString(EscapeString(s)) == s
// always holds". The markup row carries a quote, an angle bracket, and
// an ampersand; unescaped, its quote would end the attribute early and
// the read-back would stop short of the name.
func TestNewWebRoutes_SiteNameMeta(t *testing.T) {
	tests := []struct {
		name     string
		siteName string
		wantRaw  string
	}{
		{name: "plain text", siteName: "Fixture Group", wantRaw: "Fixture Group"},
		{name: "markup characters", siteName: `The "Shed" <Crew> & Co`, wantRaw: "The &#34;Shed&#34; &lt;Crew&gt; &amp; Co"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			routes, err := newWebRoutesFromFS(nil, webTestPrefix(t), tt.siteName, privateFiles{}, discardTestLogger(), realBuildMapFS(t))
			if err != nil {
				t.Fatalf("newWebRoutesFromFS: %v", err)
			}
			for _, doc := range []struct {
				path    string
				handler http.Handler
			}{
				{"/dubplate/login", routes.login},
				{"/dubplate/login/link", routes.login},
				{"/dubplate/members", routes.members},
			} {
				for _, palette := range []string{"label", "label-night", "system"} {
					body := servedVariant(t, doc.handler, doc.path, palette)
					if n := strings.Count(body, `<meta name="dubplate-site-name"`); n != 1 {
						t.Fatalf("%s as %q carries %d site-name metas, want 1", doc.path, palette, n)
					}
					m := siteNameMetaRE.FindStringSubmatch(body)
					if m == nil {
						t.Fatalf("%s as %q: no well-formed site-name meta in %q", doc.path, palette, body)
					}
					if m[1] != tt.wantRaw {
						t.Errorf("%s as %q: content attribute = %q, want %q", doc.path, palette, m[1], tt.wantRaw)
					}
					if got := html.UnescapeString(m[1]); got != tt.siteName {
						t.Errorf("%s as %q: content read back = %q, want %q", doc.path, palette, got, tt.siteName)
					}
				}
			}
		})
	}
}

// TestNewWebRoutes_RobotsMeta proves the two sign-in documents carry
// the robots meta exactly once in every variant (W7.15: "Both auth
// pages carry noindex, nofollow") and the members document carries
// none.
func TestNewWebRoutes_RobotsMeta(t *testing.T) {
	routes, err := newWebRoutesFromFS(nil, webTestPrefix(t), fixtureSiteName, privateFiles{}, discardTestLogger(), realBuildMapFS(t))
	if err != nil {
		t.Fatalf("newWebRoutesFromFS: %v", err)
	}
	tests := []struct {
		path    string
		handler http.Handler
		want    int
	}{
		{"/dubplate/login", routes.login, 1},
		{"/dubplate/login/link", routes.login, 1},
		{"/dubplate/members", routes.members, 0},
	}
	for _, tt := range tests {
		for _, palette := range []string{"label", "label-night", "system"} {
			body := servedVariant(t, tt.handler, tt.path, palette)
			if n := strings.Count(body, `<meta name="robots" content="noindex, nofollow">`); n != tt.want {
				t.Errorf("%s as %q carries the robots meta %d times, want %d", tt.path, palette, n, tt.want)
			}
			if tt.want == 0 && strings.Contains(body, `name="robots"`) {
				t.Errorf("%s as %q carries a robots meta, want none", tt.path, palette)
			}
		}
	}
}

// TestIconsRoute proves the mark's five files answer at their exact
// paths over the real bundle, each with the bytes of the checked-in
// file under web/static/, its own Content-Type, Cache-Control:
// no-cache, svgCSP on an SVG and no policy on the PNG, and a strong ETag that answers a matching If-None-Match
// with 304, and that every other shape under the three subtrees
// answers 404.
func TestIconsRoute(t *testing.T) {
	handler := newWebShellHandler(t, nil, realBuildMapFS(t))

	files := []struct {
		path        string
		static      string
		contentType string
	}{
		{"/dubplate/favicon.svg", "favicon.svg", "image/svg+xml"},
		{"/dubplate/favicon.png", "favicon.png", "image/png"},
		{"/dubplate/mark/mark-large.svg", "mark/mark-large.svg", "image/svg+xml"},
		{"/dubplate/mark/mark-medium.svg", "mark/mark-medium.svg", "image/svg+xml"},
		{"/dubplate/mark/mark-small.svg", "mark/mark-small.svg", "image/svg+xml"},
	}
	for _, f := range files {
		t.Run(f.path, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("..", "..", "web", "static", f.static))
			if err != nil {
				t.Fatalf("read the checked-in file: %v", err)
			}
			rec := serveRequest(handler, http.MethodGet, f.path)
			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s: status = %d, want 200", f.path, rec.Code)
			}
			if !bytes.Equal(rec.Body.Bytes(), want) {
				t.Errorf("GET %s: body differs from web/static/%s", f.path, f.static)
			}
			if got := rec.Header().Get("Content-Type"); got != f.contentType {
				t.Errorf("Content-Type = %q, want %q", got, f.contentType)
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
				t.Errorf("Cache-Control = %q, want %q", got, "no-cache")
			}
			wantCSP := ""
			if f.contentType == "image/svg+xml" {
				wantCSP = svgCSP
			}
			if got := rec.Header().Get("Content-Security-Policy"); got != wantCSP {
				t.Errorf("Content-Security-Policy = %q, want %q", got, wantCSP)
			}
			wantETag := fontsFaceETag(want)
			if got := rec.Header().Get("ETag"); got != wantETag {
				t.Errorf("ETag = %q, want %q", got, wantETag)
			}
			r := httptest.NewRequest(http.MethodGet, f.path, nil)
			r.Header.Set("If-None-Match", wantETag)
			revalidate := httptest.NewRecorder()
			handler.ServeHTTP(revalidate, r)
			if revalidate.Code != http.StatusNotModified {
				t.Errorf("GET %s with a matching If-None-Match: status = %d, want 304", f.path, revalidate.Code)
			}
		})
	}

	for _, path := range []string{
		"/dubplate/mark",
		"/dubplate/mark/",
		"/dubplate/mark/mark-tiny.svg",
		"/dubplate/mark/mark-small.svg/",
		"/dubplate/favicon.svg/",
		"/dubplate/favicon.png/x",
		"/favicon.ico",
	} {
		if rec := serveRequest(handler, http.MethodGet, path); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: status = %d, want 404", path, rec.Code)
		}
	}
	if rec := serveRequest(handler, http.MethodPost, "/dubplate/favicon.svg"); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST /dubplate/favicon.svg: status = %d, want 405", rec.Code)
	}
}
