package main

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/xmlrule"
)

// realCachedXSD reads the workstation's cached schema before a test
// redirects the cache. It fails rather than skips when the cache is
// empty.
func realCachedXSD(t *testing.T) []byte {
	t.Helper()
	path, err := xmlrule.CachePath()
	if err != nil {
		t.Fatalf("locate schema cache: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached schema at %s: %v; run make subsonic-xsd", path, err)
	}
	return data
}

// xsdClient answers every request, asserted to target xmlrule.XSDURL,
// with the given status and body, and counts the requests.
func xsdClient(t *testing.T, status int, body []byte) (*http.Client, *int) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)
	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	requests := new(int)
	return &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != xmlrule.XSDURL {
			t.Errorf("request to %s, want %s", req.URL, xmlrule.XSDURL)
		}
		*requests++
		clone := req.Clone(req.Context())
		clone.URL.Scheme, clone.URL.Host, clone.Host = target.Scheme, target.Host, target.Host
		return http.DefaultTransport.RoundTrip(clone)
	})}, requests
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRunFetchXSD(t *testing.T) {
	good := realCachedXSD(t)
	corrupt := append([]byte(nil), good...)
	corrupt[10] ^= 0x01

	tests := []struct {
		name         string
		cached       []byte
		status       int
		body         []byte
		wantRequests int
		wantErr      bool
		wantFile     []byte // nil: no file may exist
	}{
		{name: "empty cache downloads", status: 200, body: good, wantRequests: 1, wantFile: good},
		{name: "verifying cache is a no-op", cached: good, status: 200, body: good, wantRequests: 0, wantFile: good},
		{name: "failing cache is replaced", cached: corrupt, status: 200, body: good, wantRequests: 1, wantFile: good},
		{name: "non-200 writes nothing", status: 500, body: good, wantRequests: 1, wantErr: true},
		{name: "bad download writes nothing", status: 200, body: corrupt, wantRequests: 1, wantErr: true},
		{name: "bad download leaves the failing cache", cached: corrupt, status: 200, body: corrupt, wantRequests: 1, wantErr: true, wantFile: corrupt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			t.Setenv("XDG_CACHE_HOME", dir)
			path := filepath.Join(dir, "dubplate", "subsonic-rest-api-1.16.1.xsd")
			if tt.cached != nil {
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, tt.cached, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			client, requests := xsdClient(t, tt.status, tt.body)

			var stderr bytes.Buffer
			err := runFetchXSD(context.Background(), client, &stderr)
			if (err != nil) != tt.wantErr {
				t.Fatalf("runFetchXSD err = %v, wantErr %v", err, tt.wantErr)
			}
			if *requests != tt.wantRequests {
				t.Errorf("requests = %d, want %d", *requests, tt.wantRequests)
			}
			got, readErr := os.ReadFile(path)
			if tt.wantFile == nil {
				if !errors.Is(readErr, os.ErrNotExist) {
					t.Fatalf("a file exists at %s after a refused fetch (read err %v)", path, readErr)
				}
				return
			}
			if readErr != nil || !bytes.Equal(got, tt.wantFile) {
				t.Fatalf("cache file mismatch (read err %v)", readErr)
			}
		})
	}
}

func TestFetchXSDVerbRegistered(t *testing.T) {
	cmd, _, err := newRootCmd().Find([]string{"fetch-subsonic-xsd"})
	if err != nil || cmd.Name() != "fetch-subsonic-xsd" {
		t.Fatalf("Find = %v, %v; want the fetch-subsonic-xsd verb", cmd, err)
	}
}
