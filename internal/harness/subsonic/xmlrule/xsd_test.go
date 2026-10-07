package xmlrule

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// realXSD reads the workstation's cached schema before a test
// redirects the cache. It fails rather than skips when the cache is
// empty.
func realXSD(t *testing.T) []byte {
	t.Helper()
	dir, err := os.UserCacheDir()
	if err != nil {
		t.Fatalf("locate cache: %v", err)
	}
	path := filepath.Join(dir, "dubplate", xsdFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached schema at %s: %v; run make subsonic-xsd", path, err)
	}
	return data
}

// seedCache points the cache at a temporary directory and, when data
// is non-nil, writes it as the cached schema.
func seedCache(t *testing.T, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	path := filepath.Join(dir, "dubplate", xsdFileName)
	if data != nil {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestCachePath(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	got, err := CachePath()
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "dubplate", "subsonic-rest-api-1.16.1.xsd")
	if got != want {
		t.Fatalf("CachePath = %s, want %s", got, want)
	}
}

func TestCachedUnresolvableDirectory(t *testing.T) {
	t.Setenv("XDG_CACHE_HOME", "")
	t.Setenv("HOME", "")
	if _, err := Cached(); err == nil || !strings.Contains(err.Error(), "locate cache directory") {
		t.Fatalf("Cached err = %v, want a wrapped cache-directory error", err)
	}
}

func TestCached(t *testing.T) {
	good := realXSD(t)
	corrupt := append([]byte(nil), good...)
	corrupt[len(corrupt)/2] ^= 0x01

	tests := []struct {
		name    string
		seed    []byte
		wantErr error
		wantMsg []string
	}{
		{name: "verifies", seed: good},
		{name: "absent", seed: nil, wantErr: ErrNotCached, wantMsg: []string{"make subsonic-xsd", xsdFileName}},
		{name: "one byte changed", seed: corrupt, wantErr: ErrChecksum, wantMsg: []string{xsdSHA256, xsdFileName}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seedCache(t, tt.seed)
			got, err := Cached()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Cached err = %v, want %v", err, tt.wantErr)
			}
			for _, m := range tt.wantMsg {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("error %q lacks %q", err, m)
				}
			}
			if tt.wantErr == nil && string(got) != string(good) {
				t.Fatal("Cached returned bytes that differ from the schema")
			}
			if tt.wantErr != nil && got != nil {
				t.Fatalf("Cached returned %d bytes alongside an error", len(got))
			}
		})
	}
}

// serve returns a client whose every request, asserted to target
// XSDURL, is answered by h.
func serve(t *testing.T, h http.HandlerFunc) *http.Client {
	t.Helper()
	server := httptest.NewServer(h)
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != XSDURL {
			t.Errorf("request to %s, want %s", req.URL, XSDURL)
		}
		clone := req.Clone(req.Context())
		clone.URL.Scheme, clone.URL.Host, clone.Host = target.Scheme, target.Host, target.Host
		return http.DefaultTransport.RoundTrip(clone)
	})}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetch(t *testing.T) {
	good := realXSD(t)
	wrongSum := append([]byte(nil), good...)
	wrongSum[0] ^= 0x01
	tests := []struct {
		name    string
		status  int
		body    []byte
		wantErr error
		wantMsg string
	}{
		{name: "verified bytes", status: http.StatusOK, body: good},
		{name: "non-200", status: http.StatusNotFound, body: good, wantMsg: "404"},
		{name: "wrong sum", status: http.StatusOK, body: wrongSum, wantErr: ErrChecksum},
		{name: "too long", status: http.StatusOK, body: append(append([]byte(nil), good...), '\n'), wantErr: ErrSize},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := serve(t, func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write(tt.body)
			})
			got, err := Fetch(context.Background(), client)
			if tt.wantErr == nil && tt.wantMsg == "" {
				if err != nil || string(got) != string(good) {
					t.Fatalf("Fetch = %d bytes, %v; want the schema", len(got), err)
				}
				return
			}
			if err == nil || got != nil {
				t.Fatalf("Fetch = %d bytes, %v; want an error and no bytes", len(got), err)
			}
			if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
				t.Errorf("err = %v, want %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantMsg) {
				t.Errorf("err = %v, want it to contain %q", err, tt.wantMsg)
			}
		})
	}
}
