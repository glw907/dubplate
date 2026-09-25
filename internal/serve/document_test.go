package serve

import (
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
)

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
	raw := []byte(`<html data-theme="TOKEN">`)
	variants, err := BuildDocumentVariants("doc.html", raw, "TOKEN", "dubplate", []string{"label", "label-night", "system"})
	if err != nil {
		t.Fatalf("BuildDocumentVariants: %v", err)
	}
	want := map[string]string{
		"label":       `<html data-theme="dubplate-label">`,
		"label-night": `<html data-theme="dubplate-label-night">`,
		"system":      `<html data-theme="dubplate-system">`,
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
	raw := []byte(`no token here`)
	if _, err := BuildDocumentVariants("doc.html", raw, "TOKEN", "dubplate", []string{"label"}); err == nil {
		t.Fatal("BuildDocumentVariants: got nil error, want the missing-placeholder refusal")
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDocument(documentVariantFixture(), tt.defaultPalette)
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewDocument(documentVariantFixture(), "system")
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
			handler := NewDocument(documentVariantFixture(), "system")
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
	handler := NewDocument(documentVariantFixture(), "system")
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

func TestNewDocument_PanicsOnMissingDefaultVariant(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("NewDocument: want a panic when defaultPalette has no variant")
		}
	}()
	NewDocument(documentVariantFixture(), "gruvbox")
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
	}{
		{name: "an embedded file", path: "/dubplate/_app/immutable/app.js", wantStatus: http.StatusOK, wantBody: "console.log(1)"},
		{name: "another embedded file", path: "/dubplate/_app/version.json", wantStatus: http.StatusOK, wantBody: `{"version":"1"}`},
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
