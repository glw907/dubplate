package webapp

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

// captureDefaultLogger swaps slog.Default() for a logger writing to
// the returned buffer, restoring the prior default when t ends.
// documentHandler writes its own records to the logger it is handed, so a
// mutant logging the theme cookie's value could reach that logger or
// slog.Default; a test watching for the value reads both. The caller
// must not run this test in parallel with another that reads or writes
// slog.Default.
func captureDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prior := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prior) })
	return &buf
}

// discardLogger returns a logger that writes nowhere, for a test that
// reads no record.
func discardLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

func TestRewriteDocument(t *testing.T) {
	tests := []struct {
		name        string
		raw         string
		placeholder string
		composed    string
		want        string
		wantErrHas  string
	}{
		{
			name:        "placeholder present once",
			raw:         `<html data-theme="TOKEN">`,
			placeholder: "TOKEN",
			composed:    "dubplate-label",
			want:        `<html data-theme="dubplate-label">`,
		},
		{
			name:        "placeholder absent",
			raw:         `<html data-theme="other">`,
			placeholder: "TOKEN",
			composed:    "dubplate-label",
			wantErrHas:  "not found",
		},
		{
			name:        "placeholder present twice",
			raw:         `TOKEN and TOKEN again`,
			placeholder: "TOKEN",
			composed:    "dubplate-label",
			wantErrHas:  "appears 2 times",
		},
		{
			name:        "no other byte changes",
			raw:         `<script>1+1</script><html data-theme="TOKEN">`,
			placeholder: "TOKEN",
			composed:    "dubplate-system",
			want:        `<script>1+1</script><html data-theme="dubplate-system">`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := rewriteDocument("doc.html", []byte(tt.raw), tt.placeholder, tt.composed)
			if tt.wantErrHas != "" {
				if err == nil {
					t.Fatalf("rewriteDocument: got nil error, want one containing %q", tt.wantErrHas)
				}
				if !strings.Contains(err.Error(), tt.wantErrHas) {
					t.Fatalf("rewriteDocument error = %q, want it to contain %q", err.Error(), tt.wantErrHas)
				}
				return
			}
			if err != nil {
				t.Fatalf("rewriteDocument: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("rewriteDocument = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildDocumentVariants(t *testing.T) {
	raw := []byte(`<html data-theme="TOKEN"><link href="./a.css" rel="stylesheet"></head>`)
	variants, err := buildDocumentVariants("doc.html", raw, variantConfig{
		placeholder: "TOKEN",
		match:       "dubplate",
		palettes:    []string{"label", "label-night", "system"},
		siteName:    "Group",
	})
	if err != nil {
		t.Fatalf("buildDocumentVariants: %v", err)
	}
	want := map[string]string{
		"label":       `<html data-theme="dubplate-label"><link href="./a.css" rel="stylesheet"><meta name="dubplate-site-name" content="Group"></head>`,
		"label-night": `<html data-theme="dubplate-label-night"><link href="./a.css" rel="stylesheet"><meta name="dubplate-site-name" content="Group"></head>`,
		"system":      `<html data-theme="dubplate-system"><link href="./a.css" rel="stylesheet"><meta name="dubplate-site-name" content="Group"></head>`,
	}
	if len(variants) != len(want) {
		t.Fatalf("buildDocumentVariants: got %d variants, want %d", len(variants), len(want))
	}
	for palette, wantBody := range want {
		if got := string(variants[palette]); got != wantBody {
			t.Errorf("variants[%q] = %q, want %q", palette, got, wantBody)
		}
	}
}

func TestBuildDocumentVariants_PropagatesRewriteError(t *testing.T) {
	raw := []byte(`<link href="./a.css" rel="stylesheet"></head>no token here`)
	if _, err := buildDocumentVariants("doc.html", raw, variantConfig{placeholder: "TOKEN", match: "dubplate", palettes: []string{"label"}}); err == nil {
		t.Fatal("buildDocumentVariants: got nil error, want the missing-placeholder refusal")
	}
}

// TestBuildDocumentVariants_PlaceholderInValue covers an operator value
// carrying the theme placeholder. The rewrite runs on the build's own
// bytes before any insertion, so a document missing its placeholder is
// refused as missing even when the site name would supply one, and an
// inserted value carrying it is refused by name rather than counted.
func TestBuildDocumentVariants_PlaceholderInValue(t *testing.T) {
	const token = "__DUBPLATE_THEME__"
	tests := []struct {
		name       string
		raw        string
		cfg        variantConfig
		wantErrHas string
	}{
		{
			name:       "the build lost its placeholder, the site name carries one",
			raw:        `<html data-theme="x"><head><link href="./0.css" rel="stylesheet"></head></html>`,
			cfg:        variantConfig{siteName: "Group " + token},
			wantErrHas: `placeholder "__DUBPLATE_THEME__" not found`,
		},
		{
			name:       "the site name carries the placeholder",
			raw:        `<html data-theme="` + token + `"><head><link href="./0.css" rel="stylesheet"></head></html>`,
			cfg:        variantConfig{siteName: "Group " + token},
			wantErrHas: `an inserted head value carries the theme placeholder`,
		},
		{
			name:       "a tier href carries the placeholder",
			raw:        `<html data-theme="` + token + `"><head><link href="./0.css" rel="stylesheet"></head></html>`,
			cfg:        variantConfig{siteName: "Group", tierHrefs: []string{"/" + token + "/branding.css"}},
			wantErrHas: `an inserted head value carries the theme placeholder`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			cfg.placeholder, cfg.match, cfg.palettes = token, "dubplate", []string{"label"}
			_, err := buildDocumentVariants("login.html", []byte(tt.raw), cfg)
			if err == nil || !strings.Contains(err.Error(), tt.wantErrHas) {
				t.Fatalf("buildDocumentVariants error = %v, want one containing %q", err, tt.wantErrHas)
			}
		})
	}
}

// TestBuildDocumentVariants_TierLinkAnchor covers the three anchor
// states: no stylesheet link refuses naming the document, and one or
// several put the tier links directly after the last. The expected
// bodies are written out whole, so a link landing anywhere else, or any
// other byte moving, fails the comparison.
func TestBuildDocumentVariants_TierLinkAnchor(t *testing.T) {
	tiers := []string{"./branding.css", "./operator.css"}
	tests := []struct {
		name       string
		raw        string
		tiers      []string
		want       string
		wantErrHas string
	}{
		{
			name:       "no stylesheet link",
			raw:        `<html data-theme="T"><head><link href="./a.js" rel="modulepreload"></head><body></body></html>`,
			tiers:      tiers,
			wantErrHas: `document "members.html": no stylesheet link`,
		},
		{
			name:  "one stylesheet link",
			raw:   `<html data-theme="T"><head><link href="./_app/0.css" rel="stylesheet">` + "\n" + `</head><body><script>x</script></body></html>`,
			tiers: tiers,
			want:  `<html data-theme="dubplate-label"><head><link href="./_app/0.css" rel="stylesheet"><link href="./branding.css" rel="stylesheet"><link href="./operator.css" rel="stylesheet"><meta name="dubplate-site-name" content="S">` + "\n" + `</head><body><script>x</script></body></html>`,
		},
		{
			name:  "two stylesheet links, a modulepreload between them",
			raw:   `<html data-theme="T"><head><link href="./_app/0.css" rel="stylesheet"><link href="./_app/n.js" rel="modulepreload"><link href="./_app/2.css" rel="stylesheet"></head></html>`,
			tiers: tiers,
			want:  `<html data-theme="dubplate-label"><head><link href="./_app/0.css" rel="stylesheet"><link href="./_app/n.js" rel="modulepreload"><link href="./_app/2.css" rel="stylesheet"><link href="./branding.css" rel="stylesheet"><link href="./operator.css" rel="stylesheet"><meta name="dubplate-site-name" content="S"></head></html>`,
		},
		{
			name:  "rel before href, one tier",
			raw:   `<html data-theme="T"><link rel="stylesheet" href="./_app/0.css"></head></html>`,
			tiers: []string{"../branding.css"},
			want:  `<html data-theme="dubplate-label"><link rel="stylesheet" href="./_app/0.css"><link href="../branding.css" rel="stylesheet"><meta name="dubplate-site-name" content="S"></head></html>`,
		},
		{
			name:  "an href is escaped as an attribute value",
			raw:   `<html data-theme="T"><link href="./_app/0.css" rel="stylesheet"></head></html>`,
			tiers: []string{`./a"b&c.css`},
			want:  `<html data-theme="dubplate-label"><link href="./_app/0.css" rel="stylesheet"><link href="./a&#34;b&amp;c.css" rel="stylesheet"><meta name="dubplate-site-name" content="S"></head></html>`,
		},
		{
			name:       "the last stylesheet link sits past </head>",
			raw:        `<html data-theme="T"><head><link href="./0.css" rel="stylesheet"></head><body><link href="./1.css" rel="stylesheet"></body></html>`,
			tiers:      tiers,
			wantErrHas: `document "members.html": the last stylesheet link sits outside <head>`,
		},
		{
			name:       "no </head>",
			raw:        `<html data-theme="T"><link href="./0.css" rel="stylesheet"><body></body></html>`,
			tiers:      tiers,
			wantErrHas: `document "members.html": the last stylesheet link sits outside <head>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			variants, err := buildDocumentVariants("members.html", []byte(tt.raw), variantConfig{
				placeholder: "T",
				match:       "dubplate",
				palettes:    []string{"label"},
				tierHrefs:   tt.tiers,
				siteName:    "S",
			})
			if tt.wantErrHas != "" {
				if err == nil {
					t.Fatalf("buildDocumentVariants: got nil error, want one containing %q", tt.wantErrHas)
				}
				if !strings.Contains(err.Error(), tt.wantErrHas) {
					t.Fatalf("buildDocumentVariants error = %q, want it to contain %q", err.Error(), tt.wantErrHas)
				}
				return
			}
			if err != nil {
				t.Fatalf("buildDocumentVariants: %v", err)
			}
			if got := string(variants["label"]); got != tt.want {
				t.Errorf("variant = %q\nwant      %q", got, tt.want)
			}
		})
	}
}

// TestBuildDocumentVariants_Head writes every variant's whole expected
// body out, so a head element missing, doubled, reordered, left
// unescaped, or carried into the wrong palette fails the comparison.
// The system palette carries the media pair, label one unconditional
// meta, and label-night none, the three shapes a caller can hand in.
func TestBuildDocumentVariants_Head(t *testing.T) {
	raw := `<html data-theme="T"><head><link href="./0.css" rel="stylesheet"></head><body><script>s</script></body></html>`
	cfg := variantConfig{
		placeholder: "T",
		match:       "dubplate",
		palettes:    []string{"label", "label-night", "system"},
		tierHrefs:   []string{"./branding.css"},
		siteName:    `A "B" <C> & D`,
		robots:      "noindex, nofollow",
		themeColors: map[string][]themeColor{
			"label":  {{color: "#1f1b16"}},
			"system": {{scheme: "light", color: "#1f1b16"}, {scheme: "dark", color: "#110f0d"}},
		},
		icons: []icon{
			{href: "/p/favicon.svg", contentType: "image/svg+xml"},
			{href: "/p/favicon.png", contentType: "image/png", sizes: "32x32"},
		},
	}
	shared := `<link href="./branding.css" rel="stylesheet">` +
		`<meta name="dubplate-site-name" content="A &#34;B&#34; &lt;C&gt; &amp; D">` +
		`<meta name="robots" content="noindex, nofollow">`
	icons := `<link rel="icon" href="/p/favicon.svg" type="image/svg+xml">` +
		`<link rel="icon" href="/p/favicon.png" type="image/png" sizes="32x32">`
	body := func(theme, colors string) string {
		return `<html data-theme="` + theme + `"><head><link href="./0.css" rel="stylesheet">` + shared + colors + icons +
			`</head><body><script>s</script></body></html>`
	}
	want := map[string]string{
		"label":       body("dubplate-label", `<meta name="theme-color" content="#1f1b16">`),
		"label-night": body("dubplate-label-night", ""),
		"system": body("dubplate-system",
			`<meta name="theme-color" media="(prefers-color-scheme: light)" content="#1f1b16">`+
				`<meta name="theme-color" media="(prefers-color-scheme: dark)" content="#110f0d">`),
	}

	variants, err := buildDocumentVariants("login.html", []byte(raw), cfg)
	if err != nil {
		t.Fatalf("buildDocumentVariants: %v", err)
	}
	if len(variants) != len(want) {
		t.Fatalf("buildDocumentVariants: got %d variants, want %d", len(variants), len(want))
	}
	for palette, wantBody := range want {
		if got := string(variants[palette]); got != wantBody {
			t.Errorf("variants[%q] =\n%s\nwant\n%s", palette, got, wantBody)
		}
	}
}

// documentVariantFixture is the three-palette variant set every
// TestDocumentHandler_* case selects from: each body is the palette's own
// name, so a case asserts on the palette selected rather than on
// literal markup.
func documentVariantFixture() map[string][]byte {
	return map[string][]byte{
		"label":       []byte("label"),
		"label-night": []byte("label-night"),
		"system":      []byte("system"),
	}
}

func TestDocumentHandler_CookieSelectsVariant(t *testing.T) {
	tests := []struct {
		name           string
		defaultPalette string
		cookieHeader   string // raw Cookie header line(s), joined with "; " when more than one entry is wanted on one line
		want           string
	}{
		{
			name:           "absent cookie, default system",
			defaultPalette: "system",
			cookieHeader:   "",
			want:           "system",
		},
		{
			name:           "cookie label",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=label",
			want:           "label",
		},
		{
			name:           "cookie label-night overrides default label",
			defaultPalette: "label",
			cookieHeader:   "__Host-dubplate_theme=label-night",
			want:           "label-night",
		},
		{
			name:           "cookie system overrides default label-night",
			defaultPalette: "label-night",
			cookieHeader:   "__Host-dubplate_theme=system",
			want:           "system",
		},
		{
			name:           "cookie value enclosed in double quotes",
			defaultPalette: "system",
			cookieHeader:   `__Host-dubplate_theme="label"`,
			want:           "label",
		},
		{
			name:           "unknown palette falls back to default",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=gruvbox",
			want:           "system",
		},
		{
			name:           "a prefix of a real palette never matches",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=label-n",
			want:           "system",
		},
		{
			name:           "another case never matches",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=Label",
			want:           "system",
		},
		{
			name:           "an interior space never matches a variant key",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=label night",
			want:           "system",
		},
		{
			name:           "empty value falls back to default",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=",
			want:           "system",
		},
		{
			name:           "a byte the parser refuses is skipped, so the cookie reads as absent",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=a\\b",
			want:           "system",
		},
		{
			name:           "two cookies of the name: only the first parsed one is read",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=gruvbox; __Host-dubplate_theme=label-night",
			want:           "system",
		},
		{
			name:           "two cookies, the first unparseable: the second is read",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=a\\b; __Host-dubplate_theme=label-night",
			want:           "label-night",
		},
		{
			name:           "the same value under another cookie name is ignored",
			defaultPalette: "system",
			cookieHeader:   "dubplate_theme=label-night",
			want:           "system",
		},
		{
			name:           "a padded value keeps the space inside the value, so it never matches",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme= label",
			want:           "system",
		},
		{
			name:           "an unknown value carrying a planted marker never matches, and the marker reaches no log",
			defaultPalette: "system",
			cookieHeader:   "__Host-dubplate_theme=qx7-plantedmarker-4kd9",
			want:           "system",
		},
	}
	defaultLogs := captureDefaultLogger(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultLogs.Reset()
			var handlerLogs bytes.Buffer
			handler := documentHandler(documentConfig{
				variants:       documentVariantFixture(),
				defaultPalette: tt.defaultPalette,
				logger:         slog.New(slog.NewTextHandler(&handlerLogs, nil)),
			})
			r := httptest.NewRequest(http.MethodGet, "/login", nil)
			if tt.cookieHeader != "" {
				r.Header.Set("Cookie", tt.cookieHeader)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := rec.Body.String(); got != tt.want {
				t.Errorf("body = %q, want %q", got, tt.want)
			}
			if set := rec.Header().Values("Set-Cookie"); len(set) != 0 {
				t.Errorf("Set-Cookie = %v, want none", set)
			}
			// documentHandler's own doc comment promises it "logs
			// nothing about the cookie's value"; a mutant that
			// logs cookie.Value through slog.Default makes this
			// buffer non-empty, and one that logs it through the
			// handler's own logger puts the planted marker there.
			if defaultLogs.Len() != 0 {
				t.Errorf("slog.Default buffer is not empty: %s", defaultLogs.String())
			}
			if strings.Contains(handlerLogs.String(), "plantedmarker") {
				t.Errorf("the handler's log carries the cookie's value: %s", handlerLogs.String())
			}
		})
	}
}

func TestDocumentHandler_ControlByteAndNonASCIICookiesAreSkipped(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "control byte", value: "__Host-dubplate_theme=la\x01bel"},
		{name: "non-ASCII byte", value: "__Host-dubplate_theme=la\x80bel"},
		{name: "interior double quote", value: `__Host-dubplate_theme=la"bel`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := documentHandler(documentConfig{variants: documentVariantFixture(), defaultPalette: "system", logger: discardLogger()})
			r := httptest.NewRequest(http.MethodGet, "/login", nil)
			r.Header.Set("Cookie", tt.value)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r)
			if got := rec.Body.String(); got != "system" {
				t.Errorf("body = %q, want the default %q", got, "system")
			}
		})
	}
}

func TestDocumentHandler_MethodNotAllowed(t *testing.T) {
	tests := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range tests {
		t.Run(method, func(t *testing.T) {
			handler := documentHandler(documentConfig{variants: documentVariantFixture(), defaultPalette: "system", logger: discardLogger()})
			r := httptest.NewRequest(method, "/login", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r)
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want 405", rec.Code)
			}
			if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
				t.Errorf("Allow = %q, want %q", got, "GET, HEAD")
			}
		})
	}
}

// TestDocumentHandler_HEADWritesNoBody proves the real behavior over an
// actual round trip rather than httptest.NewRecorder, which records
// whatever bytes a handler writes regardless of method: it is
// net/http's own server, not this handler, that discards a written
// body for a HEAD request while still reporting Content-Length.
func TestDocumentHandler_HEADWritesNoBody(t *testing.T) {
	handler := documentHandler(documentConfig{variants: documentVariantFixture(), defaultPalette: "system", logger: discardLogger()})
	srv := httptest.NewServer(handler)
	defer srv.Close()

	resp, err := http.Head(srv.URL + "/login")
	if err != nil {
		t.Fatalf("HEAD: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("HEAD body = %q, want empty", body)
	}
	if got := resp.Header.Get("Content-Length"); got != "6" {
		t.Errorf("Content-Length = %q, want %q (len(%q))", got, "6", "system")
	}
}

// TestDocumentHandler_ResponseHeaders proves a served document carries the
// configured Cache-Control and CSP beside Vary: Cookie, on GET and
// HEAD, and that a 405 carries none of the three: a refusal is no
// document, so it caches and varies like any other refusal.
func TestDocumentHandler_ResponseHeaders(t *testing.T) {
	const csp = "default-src 'self'; script-src 'self'"
	handler := documentHandler(documentConfig{
		variants:       documentVariantFixture(),
		defaultPalette: "system",
		cacheControl:   "private, no-cache",
		csp:            csp,
		logger:         discardLogger(),
	})
	tests := []struct {
		method    string
		wantCache string
		wantCSP   string
		wantVary  string
	}{
		{method: http.MethodGet, wantCache: "private, no-cache", wantCSP: csp, wantVary: "Cookie"},
		{method: http.MethodHead, wantCache: "private, no-cache", wantCSP: csp, wantVary: "Cookie"},
		{method: http.MethodPost},
	}
	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(tt.method, "/members", nil))
			h := rec.Header()
			if got := h.Get("Cache-Control"); got != tt.wantCache {
				t.Errorf("Cache-Control = %q, want %q", got, tt.wantCache)
			}
			if got := h.Get("Content-Security-Policy"); got != tt.wantCSP {
				t.Errorf("Content-Security-Policy = %q, want %q", got, tt.wantCSP)
			}
			if got := h.Values("Vary"); !slices.Equal(got, nonEmpty(tt.wantVary)) {
				t.Errorf("Vary = %q, want %q", got, nonEmpty(tt.wantVary))
			}
		})
	}
}

// nonEmpty returns []string{v}, or nil when v is empty, so a header's
// Values compare against a single expected value or its absence.
func nonEmpty(v string) []string {
	if v == "" {
		return nil
	}
	return []string{v}
}

// requestLogBuildFS returns the smallest build tree [New] accepts:
// the four documents, each with a CSP meta, the theme placeholder, and
// a stylesheet link relative to its own path, the one stylesheet they
// share, an asset, a face, the two favicons, and a mark.
func requestLogBuildFS() fstest.MapFS {
	doc := func(css string) *fstest.MapFile {
		return &fstest.MapFile{Data: []byte(`<html data-theme="__DUBPLATE_THEME__"><head>` +
			`<meta http-equiv="content-security-policy" content="script-src 'self'">` +
			`<link href="` + css + `" rel="stylesheet"></head></html>`)}
	}
	return fstest.MapFS{
		"login.html":           doc("./_app/0.css"),
		"login/link.html":      doc("../_app/0.css"),
		"members.html":         doc("./_app/0.css"),
		"server-settings.html": doc("./_app/0.css"),
		"_app/0.css":           &fstest.MapFile{Data: []byte(`[data-theme=dubplate-system]{color-scheme:light;--color-neutral:#111}`)},
		"_app/version.json":    &fstest.MapFile{Data: []byte(`{"version":"1"}`)},
		"fonts/face.woff2":     &fstest.MapFile{Data: []byte("woff2 bytes")},
		"favicon.svg":          &fstest.MapFile{Data: []byte("<svg/>")},
		"favicon.png":          &fstest.MapFile{Data: []byte("png bytes")},
		"mark/mark-small.svg":  &fstest.MapFile{Data: []byte("<svg/>")},
	}
}

// requestLogFixture is the composed handlers the document request log
// covers and the ones it must not: the routes [New] builds, and a plain
// exact-path router standing in for the probe's.
type requestLogFixture struct {
	routes Routes
	probe  http.Handler
}

// newRequestLogFixture builds a requestLogFixture whose document
// records go to logger.
func newRequestLogFixture(t *testing.T, logger *slog.Logger) requestLogFixture {
	t.Helper()
	routes, err := New(requestLogBuildFS(), Config{
		Prefix: "/dubplate",
		Names: RouteNames{
			Login:          "login",
			Members:        "members",
			ServerSettings: "server-settings",
			Assets:         "_app",
			Fonts:          "fonts",
			FaviconSVG:     "favicon.svg",
			FaviconPNG:     "favicon.png",
			Mark:           "mark",
			Branding:       "branding.css",
			OperatorCSS:    "operator.css",
		},
		SiteName: "S",
		Match:    "dubplate",
		Palette:  "system",
		Branding: []byte{},
		Logger:   logger,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	ok := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	return requestLogFixture{
		routes: routes,
		probe:  exactPaths(map[string]http.Handler{"/dubplate/healthz": ok}, http.NotFoundHandler()),
	}
}

// decodeRecords returns every JSON record buf holds, one per line.
func decodeRecords(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var records []map[string]any
	for line := range strings.Lines(buf.String()) {
		var rec map[string]any
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			t.Fatalf("log line %q is not a JSON record: %v", line, err)
		}
		records = append(records, rec)
	}
	return records
}

// TestDocumentRequestLog proves every answer a document route writes
// is logged once with its method, path, and status, and that a route
// beside the documents, refusing or not found, writes no record.
func TestDocumentRequestLog(t *testing.T) {
	tests := []struct {
		name    string
		handler func(requestLogFixture) http.Handler
		method  string
		path    string
		want    map[string]any // nil when no record is wanted
	}{
		{
			name:    "a document GET",
			handler: func(f requestLogFixture) http.Handler { return f.routes.Login },
			method:  http.MethodGet,
			path:    "/dubplate/login",
			want:    map[string]any{"msg": "webapp: document request", "method": "GET", "path": "/dubplate/login", "status": float64(200)},
		},
		{
			name:    "a document POST",
			handler: func(f requestLogFixture) http.Handler { return f.routes.Login },
			method:  http.MethodPost,
			path:    "/dubplate/login",
			want:    map[string]any{"msg": "webapp: document request", "method": "POST", "path": "/dubplate/login", "status": float64(405)},
		},
		{
			name:    "an unknown path under the document router",
			handler: func(f requestLogFixture) http.Handler { return f.routes.Login },
			method:  http.MethodGet,
			path:    "/dubplate/login/other",
			want:    map[string]any{"msg": "webapp: document request", "method": "GET", "path": "/dubplate/login/other", "status": float64(404)},
		},
		{
			name:    "the members document GET",
			handler: func(f requestLogFixture) http.Handler { return f.routes.Members },
			method:  http.MethodGet,
			path:    "/dubplate/members",
			want:    map[string]any{"msg": "webapp: document request", "method": "GET", "path": "/dubplate/members", "status": float64(200)},
		},
		{
			name:    "a stylesheet POST",
			handler: func(f requestLogFixture) http.Handler { return f.routes.Branding },
			method:  http.MethodPost,
			path:    "/dubplate/branding.css",
		},
		{
			name:    "a path under the stylesheet",
			handler: func(f requestLogFixture) http.Handler { return f.routes.Branding },
			method:  http.MethodGet,
			path:    "/dubplate/branding.css/deeper",
		},
		{
			name:    "an asset POST",
			handler: func(f requestLogFixture) http.Handler { return f.routes.Assets },
			method:  http.MethodPost,
			path:    "/dubplate/_app/version.json",
		},
		{
			name:    "a probe 404",
			handler: func(f requestLogFixture) http.Handler { return f.probe },
			method:  http.MethodGet,
			path:    "/dubplate/healthz/deeper",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			fixture := newRequestLogFixture(t, slog.New(slog.NewJSONHandler(&buf, nil)))
			rec := httptest.NewRecorder()
			tt.handler(fixture).ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))

			records := decodeRecords(t, &buf)
			if tt.want == nil {
				if len(records) != 0 {
					t.Fatalf("status %d: records = %v, want none", rec.Code, records)
				}
				return
			}
			if len(records) != 1 {
				t.Fatalf("records = %v, want exactly one", records)
			}
			for key, want := range tt.want {
				if got := records[0][key]; got != want {
					t.Errorf("record[%q] = %#v, want %#v", key, got, want)
				}
			}
			if got := records[0]["status"]; got != float64(rec.Code) {
				t.Errorf("record status = %#v, but the response's is %d", got, rec.Code)
			}
		})
	}
}

// TestDocumentRequestLog_OmitsQuery proves the record names the
// method, the path, and the status the document route wrote, and that
// a query string never reaches it, whatever it carries.
func TestDocumentRequestLog_OmitsQuery(t *testing.T) {
	var buf bytes.Buffer
	const token = "qz81-fixture-token"
	handler := newRequestLogFixture(t, slog.New(slog.NewJSONHandler(&buf, nil))).routes.Login

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/dubplate/login/link?t="+token, nil))

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log = %q, want one JSON record: %v", buf.String(), err)
	}
	if rec["method"] != http.MethodPost || rec["path"] != "/dubplate/login/link" || rec["status"] != float64(http.StatusMethodNotAllowed) {
		t.Errorf("record = %v, want method POST, path /dubplate/login/link, status 405", rec)
	}
	if strings.Contains(buf.String(), token) {
		t.Errorf("log carries the query's token: %s", buf.String())
	}
}

// TestDocumentRequestLog_BoundsMethod proves a method outside the ones
// net/http names is logged as "other", so a client-chosen method of any
// length never reaches a record, while a named method is logged as sent.
func TestDocumentRequestLog_BoundsMethod(t *testing.T) {
	tests := []struct {
		name   string
		method string
		want   string
	}{
		{name: "GET", method: http.MethodGet, want: "GET"},
		{name: "TRACE", method: http.MethodTrace, want: "TRACE"},
		{name: "lowercase get", method: "get", want: "other"},
		{name: "a long token", method: strings.Repeat("X", 4096), want: "other"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			handler := newRequestLogFixture(t, slog.New(slog.NewJSONHandler(&buf, nil))).routes.Login
			r := httptest.NewRequest(http.MethodGet, "/dubplate/login", nil)
			r.Method = tt.method
			handler.ServeHTTP(httptest.NewRecorder(), r)

			var rec map[string]any
			if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
				t.Fatalf("log = %q, want one JSON record: %v", buf.String(), err)
			}
			if rec["method"] != tt.want {
				t.Errorf("method = %.40q, want %q", rec["method"], tt.want)
			}
		})
	}
}

func TestDocumentHandler_PanicsOnMissingDefaultVariant(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("documentHandler: want a panic when defaultPalette has no variant")
		}
	}()
	documentHandler(documentConfig{variants: documentVariantFixture(), defaultPalette: "gruvbox", logger: discardLogger()})
}

func TestStylesheetHrefs(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want []string
	}{
		{
			name: "href before rel (the built document's own shape)",
			doc:  `<link href="./_app/immutable/assets/0.css" rel="stylesheet">`,
			want: []string{"./_app/immutable/assets/0.css"},
		},
		{
			name: "rel before href",
			doc:  `<link rel="stylesheet" href="./x.css">`,
			want: []string{"./x.css"},
		},
		{
			name: "a non-stylesheet rel is ignored",
			doc:  `<link href="./x.js" rel="modulepreload">`,
			want: nil,
		},
		{
			name: "several links, in document order",
			doc:  `<link href="./a.css" rel="stylesheet"><link href="./x.js" rel="modulepreload"><link href="./b.css" rel="stylesheet">`,
			want: []string{"./a.css", "./b.css"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := stylesheetHrefs([]byte(tt.doc))
			if !slices.Equal(got, tt.want) {
				t.Errorf("stylesheetHrefs =%v, want %v", got, tt.want)
			}
		})
	}
}
