package webapp

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"
)

func TestStylesheetHandler(t *testing.T) {
	body := []byte(":root {\n\t--depth: 1;\n}\n")
	handler := stylesheetHandler("/branding.css", body)

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

func assetsRoot() fstest.MapFS {
	return fstest.MapFS{
		"immutable/app.js": &fstest.MapFile{Data: []byte("console.log(1)")},
		"version.json":     &fstest.MapFile{Data: []byte(`{"version":"1"}`)},
	}
}

func TestAssetsHandler(t *testing.T) {
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
			handler := assetsHandler(assetsRoot(), "/dubplate/_app")
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

func TestAssetsHandler_NeverListsOrRedirectsADirectory(t *testing.T) {
	handler := assetsHandler(assetsRoot(), "/dubplate/_app")
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

func TestAssetsHandler_MethodNotAllowed(t *testing.T) {
	handler := assetsHandler(assetsRoot(), "/dubplate/_app")
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
