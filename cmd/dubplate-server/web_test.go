package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/glw907/dubplate/internal/serve"
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
