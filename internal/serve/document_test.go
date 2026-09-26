package serve

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
// NewDocument takes no logger of its own, so slog.Default is the only
// sink a mutant logging the theme cookie's value could reach; the
// caller must not run this test in parallel with another that reads
// or writes slog.Default.
func captureDefaultLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prior := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prior) })
	return &buf
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
			got, err := RewriteDocument("doc.html", []byte(tt.raw), tt.placeholder, tt.composed)
			if tt.wantErrHas != "" {
				if err == nil {
					t.Fatalf("RewriteDocument: got nil error, want one containing %q", tt.wantErrHas)
				}
				if !strings.Contains(err.Error(), tt.wantErrHas) {
					t.Fatalf("RewriteDocument error = %q, want it to contain %q", err.Error(), tt.wantErrHas)
				}
				return
			}
			if err != nil {
				t.Fatalf("RewriteDocument: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("RewriteDocument = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildDocumentVariants(t *testing.T) {
	raw := []byte(`<html data-theme="TOKEN"><link href="./a.css" rel="stylesheet"></head>`)
	variants, err := BuildDocumentVariants("doc.html", raw, VariantConfig{
		Placeholder: "TOKEN",
		Match:       "dubplate",
		Palettes:    []string{"label", "label-night", "system"},
		SiteName:    "Group",
	})
	if err != nil {
		t.Fatalf("BuildDocumentVariants: %v", err)
	}
	want := map[string]string{
		"label":       `<html data-theme="dubplate-label"><link href="./a.css" rel="stylesheet"><meta name="dubplate-site-name" content="Group"></head>`,
		"label-night": `<html data-theme="dubplate-label-night"><link href="./a.css" rel="stylesheet"><meta name="dubplate-site-name" content="Group"></head>`,
		"system":      `<html data-theme="dubplate-system"><link href="./a.css" rel="stylesheet"><meta name="dubplate-site-name" content="Group"></head>`,
	}
	if len(variants) != len(want) {
		t.Fatalf("BuildDocumentVariants: got %d variants, want %d", len(variants), len(want))
	}
	for palette, wantBody := range want {
		if got := string(variants[palette]); got != wantBody {
			t.Errorf("variants[%q] = %q, want %q", palette, got, wantBody)
		}
	}
}

func TestBuildDocumentVariants_PropagatesRewriteError(t *testing.T) {
	raw := []byte(`<link href="./a.css" rel="stylesheet"></head>no token here`)
	if _, err := BuildDocumentVariants("doc.html", raw, VariantConfig{Placeholder: "TOKEN", Match: "dubplate", Palettes: []string{"label"}}); err == nil {
		t.Fatal("BuildDocumentVariants: got nil error, want the missing-placeholder refusal")
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
		cfg        VariantConfig
		wantErrHas string
	}{
		{
			name:       "the build lost its placeholder, the site name carries one",
			raw:        `<html data-theme="x"><head><link href="./0.css" rel="stylesheet"></head></html>`,
			cfg:        VariantConfig{SiteName: "Group " + token},
			wantErrHas: `placeholder "__DUBPLATE_THEME__" not found`,
		},
		{
			name:       "the site name carries the placeholder",
			raw:        `<html data-theme="` + token + `"><head><link href="./0.css" rel="stylesheet"></head></html>`,
			cfg:        VariantConfig{SiteName: "Group " + token},
			wantErrHas: `an inserted head value carries the theme placeholder`,
		},
		{
			name:       "a tier href carries the placeholder",
			raw:        `<html data-theme="` + token + `"><head><link href="./0.css" rel="stylesheet"></head></html>`,
			cfg:        VariantConfig{SiteName: "Group", TierHrefs: []string{"/" + token + "/branding.css"}},
			wantErrHas: `an inserted head value carries the theme placeholder`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.cfg
			cfg.Placeholder, cfg.Match, cfg.Palettes = token, "dubplate", []string{"label"}
			_, err := BuildDocumentVariants("login.html", []byte(tt.raw), cfg)
			if err == nil || !strings.Contains(err.Error(), tt.wantErrHas) {
				t.Fatalf("BuildDocumentVariants error = %v, want one containing %q", err, tt.wantErrHas)
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
			variants, err := BuildDocumentVariants("members.html", []byte(tt.raw), VariantConfig{
				Placeholder: "T",
				Match:       "dubplate",
				Palettes:    []string{"label"},
				TierHrefs:   tt.tiers,
				SiteName:    "S",
			})
			if tt.wantErrHas != "" {
				if err == nil {
					t.Fatalf("BuildDocumentVariants: got nil error, want one containing %q", tt.wantErrHas)
				}
				if !strings.Contains(err.Error(), tt.wantErrHas) {
					t.Fatalf("BuildDocumentVariants error = %q, want it to contain %q", err.Error(), tt.wantErrHas)
				}
				return
			}
			if err != nil {
				t.Fatalf("BuildDocumentVariants: %v", err)
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
	cfg := VariantConfig{
		Placeholder: "T",
		Match:       "dubplate",
		Palettes:    []string{"label", "label-night", "system"},
		TierHrefs:   []string{"./branding.css"},
		SiteName:    `A "B" <C> & D`,
		Robots:      "noindex, nofollow",
		ThemeColors: map[string][]ThemeColor{
			"label":  {{Color: "#1f1b16"}},
			"system": {{Scheme: "light", Color: "#1f1b16"}, {Scheme: "dark", Color: "#110f0d"}},
		},
		Icons: []Icon{
			{Href: "/p/favicon.svg", Type: "image/svg+xml"},
			{Href: "/p/favicon.png", Type: "image/png", Sizes: "32x32"},
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

	variants, err := BuildDocumentVariants("login.html", []byte(raw), cfg)
	if err != nil {
		t.Fatalf("BuildDocumentVariants: %v", err)
	}
	if len(variants) != len(want) {
		t.Fatalf("BuildDocumentVariants: got %d variants, want %d", len(variants), len(want))
	}
	for palette, wantBody := range want {
		if got := string(variants[palette]); got != wantBody {
			t.Errorf("variants[%q] =\n%s\nwant\n%s", palette, got, wantBody)
		}
	}
}

func TestThemeNeutrals(t *testing.T) {
	tests := []struct {
		name string
		css  string
		want map[string][]ThemeColor
	}{
		{
			name: "one rule with its scheme",
			css:  `[data-theme=dubplate-label]{color-scheme:light;--color-neutral:#1f1b16;--color-neutral-content:#fff}`,
			want: map[string][]ThemeColor{"dubplate-label": {{Scheme: "light", Color: "#1f1b16"}}},
		},
		{
			name: "a media branch is read on its own selector",
			css: `[data-theme="dubplate-system"]{color-scheme: light; --color-neutral: #1f1b16}` +
				`@media (prefers-color-scheme:dark){[data-theme="dubplate-system"]{color-scheme:dark;--color-neutral:#110f0d}}`,
			want: map[string][]ThemeColor{"dubplate-system": {
				{Scheme: "light", Color: "#1f1b16"},
				{Scheme: "dark", Color: "#110f0d"},
			}},
		},
		{
			name: "a rule with no color-scheme",
			css:  `[data-theme=x]{--color-neutral:oklch(20% 0 0)}`,
			want: map[string][]ThemeColor{"x": {{Color: "oklch(20% 0 0)"}}},
		},
		{
			name: "neutral-content alone is not the neutral",
			css:  `[data-theme=x]{color-scheme:dark;--color-neutral-content:#fff}`,
			want: map[string][]ThemeColor{},
		},
		{
			name: "a rule naming no theme",
			css:  `:root{--color-neutral:#000}`,
			want: map[string][]ThemeColor{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ThemeNeutrals([]byte(tt.css))
			if len(got) != len(tt.want) {
				t.Fatalf("ThemeNeutrals = %v, want %v", got, tt.want)
			}
			for name, want := range tt.want {
				if !slices.Equal(got[name], want) {
					t.Errorf("ThemeNeutrals[%q] = %v, want %v", name, got[name], want)
				}
			}
		})
	}
}

func TestNewStylesheet(t *testing.T) {
	body := []byte(":root {\n\t--depth: 1;\n}\n")
	handler := NewStylesheet(body)

	for _, method := range []string{http.MethodGet, http.MethodHead} {
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(method, "/branding.css", nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200", rec.Code)
			}
			if got := rec.Header().Get("Content-Type"); got != "text/css; charset=utf-8" {
				t.Errorf("Content-Type = %q, want %q", got, "text/css; charset=utf-8")
			}
			if got := rec.Header().Get("Cache-Control"); got != "no-cache" {
				t.Errorf("Cache-Control = %q, want %q", got, "no-cache")
			}
			if !bytes.Equal(rec.Body.Bytes(), body) {
				t.Errorf("body = %q, want %q", rec.Body.Bytes(), body)
			}
		})
	}
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		t.Run(method, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(method, "/branding.css", nil))
			if rec.Code != http.StatusMethodNotAllowed {
				t.Fatalf("status = %d, want 405", rec.Code)
			}
			if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
				t.Errorf("Allow = %q, want %q", got, "GET, HEAD")
			}
			if got := rec.Header().Get("Cache-Control"); got != "" {
				t.Errorf("405 Cache-Control = %q, want none", got)
			}
		})
	}
}

// documentVariantFixture is the three-palette variant set every
// TestNewDocument_* case selects from: each body is the palette's own
// name, so a case asserts on the palette selected rather than on
// literal markup.
func documentVariantFixture() map[string][]byte {
	return map[string][]byte{
		"label":       []byte("label"),
		"label-night": []byte("label-night"),
		"system":      []byte("system"),
	}
}

func TestNewDocument_CookieSelectsVariant(t *testing.T) {
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
			handler := NewDocument(DocumentConfig{Variants: documentVariantFixture(), DefaultPalette: tt.defaultPalette})
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
			// NewDocument's own doc comment promises it "logs
			// nothing about the cookie's value"; a mutant that
			// logs cookie.Value through slog.Default makes this
			// buffer non-empty.
			if defaultLogs.Len() != 0 {
				t.Errorf("slog.Default buffer is not empty: %s", defaultLogs.String())
			}
		})
	}
}

func TestNewDocument_ControlByteAndNonASCIICookiesAreSkipped(t *testing.T) {
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
			handler := NewDocument(DocumentConfig{Variants: documentVariantFixture(), DefaultPalette: "system"})
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

func TestNewDocument_MethodNotAllowed(t *testing.T) {
	tests := []string{http.MethodPost, http.MethodPut, http.MethodDelete}
	for _, method := range tests {
		t.Run(method, func(t *testing.T) {
			handler := NewDocument(DocumentConfig{Variants: documentVariantFixture(), DefaultPalette: "system"})
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

// TestNewDocument_HEADWritesNoBody proves the real behavior over an
// actual round trip rather than httptest.NewRecorder, which records
// whatever bytes a handler writes regardless of method: it is
// net/http's own server, not this handler, that discards a written
// body for a HEAD request while still reporting Content-Length.
func TestNewDocument_HEADWritesNoBody(t *testing.T) {
	handler := NewDocument(DocumentConfig{Variants: documentVariantFixture(), DefaultPalette: "system"})
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

// TestNewDocument_ResponseHeaders proves a served document carries the
// configured Cache-Control and CSP beside Vary: Cookie, on GET and
// HEAD, and that a 405 carries none of the three: a refusal is no
// document, so it caches and varies like any other refusal.
func TestNewDocument_ResponseHeaders(t *testing.T) {
	const csp = "default-src 'self'; script-src 'self'"
	handler := NewDocument(DocumentConfig{
		Variants:       documentVariantFixture(),
		DefaultPalette: "system",
		CacheControl:   "private, no-cache",
		CSP:            csp,
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

// TestLogRequests proves the record names the method, the path, and
// the status the wrapped handler wrote, and that a query string never
// reaches it, whatever it carries.
func TestLogRequests(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	const token = "qz81-fixture-token"
	handler := LogRequests(logger, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	}))

	handler.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/dubplate/login/link?t="+token, nil))

	var rec map[string]any
	if err := json.Unmarshal(buf.Bytes(), &rec); err != nil {
		t.Fatalf("log = %q, want one JSON record: %v", buf.String(), err)
	}
	if rec["method"] != http.MethodPost || rec["path"] != "/dubplate/login/link" || rec["status"] != float64(http.StatusTeapot) {
		t.Errorf("record = %v, want method POST, path /dubplate/login/link, status 418", rec)
	}
	if strings.Contains(buf.String(), token) {
		t.Errorf("log carries the query's token: %s", buf.String())
	}
}

// TestLogRequests_MethodBounded proves a method outside the ones
// net/http names is logged as "other", so a client-chosen method of any
// length never reaches a record, while a named method is logged as sent.
func TestLogRequests_MethodBounded(t *testing.T) {
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
			logger := slog.New(slog.NewJSONHandler(&buf, nil))
			handler := LogRequests(logger, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
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

func TestNewDocument_PanicsOnMissingDefaultVariant(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewDocument: want a panic when defaultPalette has no variant")
		}
	}()
	NewDocument(DocumentConfig{Variants: documentVariantFixture(), DefaultPalette: "gruvbox"})
}

func newAssetsRoot() fstest.MapFS {
	return fstest.MapFS{
		"immutable/app.js": &fstest.MapFile{Data: []byte("console.log(1)")},
		"version.json":     &fstest.MapFile{Data: []byte(`{"version":"1"}`)},
	}
}

func TestNewAssets(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantBody   string
		wantCache  string
	}{
		{name: "an embedded file", path: "/dubplate/_app/immutable/app.js", wantStatus: http.StatusOK, wantBody: "console.log(1)", wantCache: "public, max-age=31536000, immutable"},
		{name: "another embedded file", path: "/dubplate/_app/version.json", wantStatus: http.StatusOK, wantBody: `{"version":"1"}`, wantCache: "no-cache"},
		{name: "the bare route prefix", path: "/dubplate/_app", wantStatus: http.StatusNotFound},
		{name: "a directory path", path: "/dubplate/_app/immutable/", wantStatus: http.StatusNotFound},
		{name: "a missing file", path: "/dubplate/_app/nope.js", wantStatus: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAssets(newAssetsRoot(), "/dubplate/_app")
			r := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, r)
			if rec.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rec.Code, tt.wantStatus)
			}
			if tt.wantBody != "" && rec.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rec.Body.String(), tt.wantBody)
			}
			if got := rec.Header().Get("Cache-Control"); got != tt.wantCache {
				t.Errorf("Cache-Control = %q, want %q", got, tt.wantCache)
			}
		})
	}
}

func TestNewAssets_NeverListsOrRedirectsADirectory(t *testing.T) {
	handler := NewAssets(newAssetsRoot(), "/dubplate/_app")
	r := httptest.NewRequest(http.MethodGet, "/dubplate/_app/immutable", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404 (never a listing or a redirect)", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Errorf("Location = %q, want none", loc)
	}
}

func TestNewAssets_MethodNotAllowed(t *testing.T) {
	handler := NewAssets(newAssetsRoot(), "/dubplate/_app")
	r := httptest.NewRequest(http.MethodPost, "/dubplate/_app/immutable/app.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, r)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want 405", rec.Code)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Errorf("Allow = %q, want %q", got, "GET, HEAD")
	}
}

func TestStylesheetHrefs(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want []string
	}{
		{
			name: "href before rel (the probe's own built shape, Task 0)",
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
			got := StylesheetHrefs([]byte(tt.doc))
			if !slices.Equal(got, tt.want) {
				t.Errorf("StylesheetHrefs = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestThemeSelectors(t *testing.T) {
	tests := []struct {
		name string
		css  string
		want map[string]struct{}
	}{
		{
			name: "unquoted",
			css:  `[data-theme=dubplate-label]{color:red}`,
			want: map[string]struct{}{"dubplate-label": {}},
		},
		{
			name: "quoted",
			css:  `[data-theme="dubplate-label-night"]{color:red}`,
			want: map[string]struct{}{"dubplate-label-night": {}},
		},
		{
			name: "inside a compound selector list",
			css:  `:is(:root:has(input.theme-controller[value=dubplate-system]:checked),[data-theme=dubplate-system]){}`,
			want: map[string]struct{}{"dubplate-system": {}},
		},
		{
			name: "a prefix is never captured as the longer name",
			css:  `[data-theme=dubplate-label]{}[data-theme=dubplate-label-night]{}`,
			want: map[string]struct{}{"dubplate-label": {}, "dubplate-label-night": {}},
		},
		{
			name: "no selector at all",
			css:  `body{color:red}`,
			want: map[string]struct{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ThemeSelectors([]byte(tt.css))
			if len(got) != len(tt.want) {
				t.Fatalf("ThemeSelectors = %v, want %v", got, tt.want)
			}
			for name := range tt.want {
				if _, ok := got[name]; !ok {
					t.Errorf("ThemeSelectors: missing %q", name)
				}
			}
		})
	}
}
