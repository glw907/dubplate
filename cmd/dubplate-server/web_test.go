package main

import (
	"bytes"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/web"
)

// fixtureThemeCSS declares the three composed selectors a fixture
// build carries, in both quoted and unquoted form, mirroring the real
// bundle's own shape (internal/serve/document_test.go's ThemeSelectors
// cases already prove the extraction handles both).
const fixtureThemeCSS = `[data-theme=dubplate-label]{--color-base-100:#fff}[data-theme="dubplate-label-night"]{--color-base-100:#000}[data-theme=dubplate-system]{--color-base-100:#fff}`

// newFixtureBuildFS returns a build tree fstest.MapFS shaped like the
// real SvelteKit output this task rewrites: a login document, a
// login/link document, each carrying placeholder once on its
// data-theme attribute and a stylesheet link resolved relative to its
// own path, and the linked CSS itself.
func newFixtureBuildFS(loginPlaceholderCount, linkPlaceholderCount int) fstest.MapFS {
	token := func(n int) string {
		return strings.Repeat(themePlaceholder, n)
	}
	loginDoc := `<html data-theme="` + token(loginPlaceholderCount) + `"><head><link href="./_app/assets/0.css" rel="stylesheet"></head></html>`
	linkDoc := `<html data-theme="` + token(linkPlaceholderCount) + `"><head><link href="../_app/assets/0.css" rel="stylesheet"></head></html>`
	return fstest.MapFS{
		"login.html":             &fstest.MapFile{Data: []byte(loginDoc)},
		"login/link.html":        &fstest.MapFile{Data: []byte(linkDoc)},
		"_app/assets/0.css":      &fstest.MapFile{Data: []byte(fixtureThemeCSS)},
		"_app/entry/start.js":    &fstest.MapFile{Data: []byte("// entry")},
		"login/nested/other.txt": &fstest.MapFile{Data: []byte("unrelated")},
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
			name:       "login/link.html missing",
			buildFS:    fstest.MapFS{"login.html": &fstest.MapFile{Data: []byte("x")}},
			wantErrHas: `"login/link.html"`,
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
			name:       "palette value a prefix of a real name",
			buildFS:    newFixtureBuildFS(1, 1),
			environ:    []string{"DUBPLATE_THEME_PALETTE=label-n"},
			wantErrHas: `no theme selector named "dubplate-label-n"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newWebRoutesFromFS(tt.environ, prefix, tt.buildFS)
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
		prefix, buildFS,
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
			routes, err := newWebRoutesFromFS([]string{"DUBPLATE_THEME_PALETTE=" + palette}, prefix, buildFS)
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
	routes, err := newWebRoutesFromFS([]string{"DUBPLATE_THEME_PALETTE=label"}, prefix, buildFS)
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
// every stampable palette: the served body, with the composed name
// swapped back for [themePlaceholder], is byte-equal to the embedded
// original, and every inline <script> body served is byte-identical to
// the original's, in the same order.
func TestNewWebRoutes_RewriteChangesOnlyTheToken(t *testing.T) {
	prefix := webTestPrefix(t)

	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	routes, err := newWebRoutes(nil, prefix)
	if err != nil {
		t.Fatalf("newWebRoutes: %v", err)
	}

	documents := []struct {
		name string
		path string
	}{
		{name: loginDocumentName, path: prefix.String() + "/login"},
		{name: loginLinkDocumentName, path: prefix.String() + "/login/link"},
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

				composed := acceptedThemeMatch + "-" + palette
				restored := bytes.Replace(served, []byte(composed), []byte(themePlaceholder), 1)
				if !bytes.Equal(restored, original) {
					t.Errorf("%s served for %q, with %q swapped back for the placeholder, is not byte-equal to the embedded original", doc.name, palette, composed)
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
	routes, err := newWebRoutes(environ, prefix)
	if err != nil {
		t.Fatalf("newWebRoutes(%v): %v", environ, err)
	}
	strip := serve.NewStrip("", "", serve.TrustedProxyConfig{}, serve.FrontExemption{})
	shell := serve.NewShell(prefix, strip, logger)
	shell.Register(loginRouteName, routes.login)
	shell.Register(assetsRouteName, routes.assets)
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
	selectors, err := extractThemeSelectors(buildFS,
		namedDocument{path: loginDocumentName, raw: loginRaw},
		namedDocument{path: loginLinkDocumentName, raw: linkRaw},
	)
	if err != nil {
		t.Fatalf("extractThemeSelectors: %v", err)
	}
	gotPalettes := stampablePalettes(selectors, acceptedThemeMatch)
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
			// The shell's own logger writes only on refusal
			// (internal/serve/serve.go's logRefusal), so a 200 row
			// leaves it untouched; any content at all, not merely
			// the cookie header as a substring, is the row logging
			// something it must not.
			if logBuf.Len() != 0 {
				t.Errorf("shell log buffer is not empty: %s", logBuf.String())
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
// login documents carry resolves, relative to that document's own URL,
// to a 200 through the real [serve.Shell] composition -- never a plain
// http.ServeMux, so the origin check and path normalization sit in the
// path too. It runs against the real embedded bundle
// ([newWebRoutes]/[web.Assets]), the fixture state the acceptance
// criterion names.
func TestPrefixProof_EveryHrefAndSrcResolves(t *testing.T) {
	const prefixValue = "/p/q"
	prefix, err := serve.NewPrefix(prefixValue)
	if err != nil {
		t.Fatalf("serve.NewPrefix: %v", err)
	}

	routes, err := newWebRoutes(nil, prefix)
	if err != nil {
		t.Fatalf("newWebRoutes: %v", err)
	}

	strip := serve.NewStrip("", "", serve.TrustedProxyConfig{}, serve.FrontExemption{})
	shell := serve.NewShell(prefix, strip, discardTestLogger())
	shell.Register("login", routes.login)
	shell.Register("_app", routes.assets)
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
// real login and _app routes rather than leaving them to a caller: the
// two subtrees answer through the shell built by the production
// wiring, using the default prefix and the real embedded bundle.
func TestNewServeShell_RegistersWebRoutes(t *testing.T) {
	environ := serveAuthEnviron(t)
	shell, _, err := buildServeShell(t, environ)
	if err != nil {
		t.Fatalf("buildServeShell: %v", err)
	}

	for _, path := range []string{"/dubplate/login", "/dubplate/login/link"} {
		r := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		shell.Handler().ServeHTTP(rec, r)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s: status %d, want 200", path, rec.Code)
		}
	}
}
