package main

import (
	"bytes"
	"io/fs"
	"log/slog"
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
// every stampable palette, with an operator CSS file configured so
// both tier links are inserted: the served body, with the two tier
// links removed and the composed name swapped back for
// [themePlaceholder], is byte-equal to the embedded original, and every
// inline <script> body served is byte-identical to the original's, in
// the same order.
func TestNewWebRoutes_RewriteChangesOnlyTheToken(t *testing.T) {
	prefix := webTestPrefix(t)

	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		t.Fatalf("fs.Sub: %v", err)
	}
	environ := []string{"DUBPLATE_OPERATOR_CSS_FILE=" + writeOperatorCSS(t, []byte("body{}"))}
	routes, err := newWebRoutes(environ, prefix)
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

				if n := bytes.Count(served, []byte(doc.tiers)); n != 1 {
					t.Fatalf("%s served for %q carries the tier links %d times, want 1", doc.name, palette, n)
				}
				composed := acceptedThemeMatch + "-" + palette
				restored := bytes.Replace(served, []byte(doc.tiers), nil, 1)
				restored = bytes.Replace(restored, []byte(composed), []byte(themePlaceholder), 1)
				if !bytes.Equal(restored, original) {
					t.Errorf("%s served for %q, with the tier links removed and %q swapped back for the placeholder, is not byte-equal to the embedded original", doc.name, palette, composed)
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
	routes, err := newWebRoutes(environ, prefix)
	if err != nil {
		t.Fatalf("newWebRoutes: %v", err)
	}

	strip := serve.NewStrip("", "", serve.TrustedProxyConfig{}, serve.FrontExemption{})
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

// newWebShellHandler builds a real [serve.Shell] mounting every route
// [newWebRoutesFromFS] builds from buildFS under the default prefix.
func newWebShellHandler(t *testing.T, environ []string, buildFS fs.FS) http.Handler {
	t.Helper()
	prefix := webTestPrefix(t)
	routes, err := newWebRoutesFromFS(environ, prefix, buildFS)
	if err != nil {
		t.Fatalf("newWebRoutesFromFS(%v): %v", environ, err)
	}
	strip := serve.NewStrip("", "", serve.TrustedProxyConfig{}, serve.FrontExemption{})
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
			_, err := newWebRoutesFromFS(environ, prefix, newFixtureBuildFS(1, 1))
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
		path       func(t *testing.T) string
		wantErrHas string
	}{
		{
			name: "path not ending in .css",
			path: func(t *testing.T) string {
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
			path:       func(t *testing.T) string { return filepath.Join(t.TempDir(), "operator.CSS") },
			wantErrHas: `the path must end in ".css"`,
		},
		{
			name:       "path absent",
			path:       func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing.css") },
			wantErrHas: "cannot read the file: no such file or directory",
		},
		{
			name: "not a regular file",
			path: func(t *testing.T) string {
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
			path: func(t *testing.T) string {
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
			path: func(t *testing.T) string {
				return writeOperatorCSS(t, append([]byte(marker), bytes.Repeat([]byte("a"), 262144+1-len(marker))...))
			},
			wantErrHas: "the file is over 262144 bytes",
		},
		{
			name:       "invalid UTF-8",
			path:       func(t *testing.T) string { return writeOperatorCSS(t, []byte(marker+"\xff\xfe")) },
			wantErrHas: "the file is not valid UTF-8",
		},
		{
			name:       "a NUL byte",
			path:       func(t *testing.T) string { return writeOperatorCSS(t, []byte(marker+"\x00body{}")) },
			wantErrHas: "the file holds a NUL byte",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.path(t)
			_, err := newWebRoutesFromFS([]string{"DUBPLATE_OPERATOR_CSS_FILE=" + path}, prefix, newFixtureBuildFS(1, 1))
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
		routes, err := newWebRoutes(tt.environ, prefix)
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
	noLink["login.html"] = &fstest.MapFile{Data: []byte(`<html data-theme="` + themePlaceholder + `"><head><link href="./_app/entry/start.js" rel="modulepreload"></head></html>`)}
	if _, err := newWebRoutesFromFS(nil, prefix, noLink); err == nil || !strings.Contains(err.Error(), `document "login.html": no stylesheet link`) {
		t.Errorf("no stylesheet link: error = %v, want one naming login.html", err)
	}

	twoLinks := newFixtureBuildFS(1, 1)
	twoLinks["login.html"] = &fstest.MapFile{Data: []byte(`<html data-theme="` + themePlaceholder + `"><head><link href="./_app/assets/0.css" rel="stylesheet"><link href="./_app/assets/2.css" rel="stylesheet"></head></html>`)}
	twoLinks["_app/assets/2.css"] = &fstest.MapFile{Data: []byte("p{}")}
	handler := newWebShellHandler(t, nil, twoLinks)
	hrefs := serve.StylesheetHrefs(serveRequest(handler, http.MethodGet, "/dubplate/login").Body.Bytes())
	if want := []string{"./_app/assets/0.css", "./_app/assets/2.css", "./branding.css"}; !slices.Equal(hrefs, want) {
		t.Errorf("stylesheet hrefs = %v, want %v", hrefs, want)
	}
}
