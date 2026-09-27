package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
)

// loadRealCachedWheel reads the workstation's real cached wheel,
// filled by `make unidecode-wheel` before this package's tests run.
// It fails rather than skips when the cache is empty.
func loadRealCachedWheel(t *testing.T) []byte {
	t.Helper()
	path, err := unidecode.WheelCachePath()
	if err != nil {
		t.Fatalf("locate wheel cache: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached wheel at %s: %v; run make unidecode-wheel", path, err)
	}
	return data
}

// rewritingTransport asserts every request targets
// unidecode.WheelURL, counts the requests it sees, and forwards a
// clone to a test server standing in for PyPI.
type rewritingTransport struct {
	t        *testing.T
	server   *httptest.Server
	requests int
}

func (rt *rewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.t.Helper()
	if req.URL.String() != unidecode.WheelURL {
		rt.t.Fatalf("request to %s, want %s", req.URL, unidecode.WheelURL)
	}
	rt.requests++

	target, err := url.Parse(rt.server.URL)
	if err != nil {
		return nil, err
	}
	clone := req.Clone(req.Context())
	clone.URL.Scheme = target.Scheme
	clone.URL.Host = target.Host
	clone.Host = target.Host
	return http.DefaultTransport.RoundTrip(clone)
}

// isolatedCacheDir points unidecode.WheelCachePath at a scratch
// directory for the duration of the test, so the real workstation
// cache is never written by these tests.
func isolatedCacheDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	return dir
}

func TestRunFetchUnidecodeDownloadsWhenCacheIsEmpty(t *testing.T) {
	wheel := loadRealCachedWheel(t)
	isolatedCacheDir(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(wheel)
	}))
	defer server.Close()
	rt := &rewritingTransport{t: t, server: server}
	client := &http.Client{Transport: rt}

	var stderr bytes.Buffer
	if err := runFetchUnidecode(context.Background(), client, &stderr); err != nil {
		t.Fatalf("runFetchUnidecode: %v", err)
	}
	if rt.requests != 1 {
		t.Fatalf("requests = %d, want 1", rt.requests)
	}

	path, err := unidecode.WheelCachePath()
	if err != nil {
		t.Fatalf("WheelCachePath: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached wheel: %v", err)
	}
	if !bytes.Equal(got, wheel) {
		t.Fatal("cached wheel differs from the fetched bytes")
	}
}

func TestRunFetchUnidecodeIsANoOpWhenCacheVerifies(t *testing.T) {
	wheel := loadRealCachedWheel(t)
	dir := isolatedCacheDir(t)

	path := filepath.Join(dir, "unidecode", "Unidecode-1.4.0-py3-none-any.whl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, wheel, 0o644); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("fetch-unidecode made a request although the cache already verifies")
	}))
	defer server.Close()
	rt := &rewritingTransport{t: t, server: server}
	client := &http.Client{Transport: rt}

	var stderr bytes.Buffer
	if err := runFetchUnidecode(context.Background(), client, &stderr); err != nil {
		t.Fatalf("runFetchUnidecode: %v", err)
	}
	if rt.requests != 0 {
		t.Fatalf("requests = %d, want 0", rt.requests)
	}
}

func TestRunFetchUnidecodeReplacesAFailingCachedFile(t *testing.T) {
	wheel := loadRealCachedWheel(t)
	dir := isolatedCacheDir(t)

	path := filepath.Join(dir, "unidecode", "Unidecode-1.4.0-py3-none-any.whl")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(path, []byte("not a valid wheel"), 0o644); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(wheel)
	}))
	defer server.Close()
	rt := &rewritingTransport{t: t, server: server}
	client := &http.Client{Transport: rt}

	var stderr bytes.Buffer
	if err := runFetchUnidecode(context.Background(), client, &stderr); err != nil {
		t.Fatalf("runFetchUnidecode: %v", err)
	}
	if rt.requests != 1 {
		t.Fatalf("requests = %d, want 1", rt.requests)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached wheel: %v", err)
	}
	if !bytes.Equal(got, wheel) {
		t.Fatal("the failing cached file was not replaced with the verified wheel")
	}
}
