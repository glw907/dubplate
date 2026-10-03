package unidecode

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

// loadCachedWheel reads the workstation's cached copy of the pinned
// wheel, filled by `make unidecode-wheel`. It fails rather than skips
// when the cache is empty, since Go's test cache does not track the
// wheel file and a stale "ok" would hide the gap.
func loadCachedWheel(t *testing.T) []byte {
	t.Helper()
	path, err := WheelCachePath()
	if err != nil {
		t.Fatalf("locate wheel cache: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cached wheel at %s: %v; run make unidecode-wheel", path, err)
	}
	return data
}

// rewritingTransport asserts that every request targets WheelURL,
// counts the requests it sees, and forwards a clone to a test server
// standing in for PyPI. Any other host fails the test instead of
// reaching the network.
type rewritingTransport struct {
	t        *testing.T
	server   *httptest.Server
	requests int
}

func (rt *rewritingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.t.Helper()
	if req.URL.String() != WheelURL {
		rt.t.Fatalf("request to %s, want %s", req.URL, WheelURL)
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

func newTestClient(t *testing.T, server *httptest.Server) (*http.Client, *rewritingTransport) {
	t.Helper()
	rt := &rewritingTransport{t: t, server: server}
	return &http.Client{Transport: rt}, rt
}

func TestFetchReturnsVerifiedWheel(t *testing.T) {
	wheel := loadCachedWheel(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(wheel)
	}))
	defer server.Close()
	client, rt := newTestClient(t, server)

	got, err := Fetch(context.Background(), client)
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if !bytes.Equal(got, wheel) {
		t.Fatal("Fetch returned bytes that differ from the cached wheel")
	}
	if rt.requests != 1 {
		t.Fatalf("requests = %d, want 1", rt.requests)
	}
}

func TestFetchRefusesNon200Status(t *testing.T) {
	for _, status := range []int{http.StatusNotFound, http.StatusInternalServerError} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(status)
		}))
		client, _ := newTestClient(t, server)

		_, err := Fetch(context.Background(), client)
		server.Close()
		if err == nil {
			t.Fatalf("Fetch against status %d = nil error, want a refusal naming it", status)
		}
	}
}

func TestFetchRefusesOversizedBody(t *testing.T) {
	body := bytes.Repeat([]byte{'x'}, maxWheelSize+1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer server.Close()
	client, _ := newTestClient(t, server)

	_, err := Fetch(context.Background(), client)
	if !errors.Is(err, errWheelTooLarge) {
		t.Fatalf("Fetch(oversized body) error = %v, want errWheelTooLarge", err)
	}
	if errors.Is(err, errWheelChecksum) {
		t.Fatal("an oversized body must refuse on size, not checksum")
	}
}

func TestFetchDoesNotRefuseExactMaxSizeForSize(t *testing.T) {
	body := bytes.Repeat([]byte{'x'}, maxWheelSize)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer server.Close()
	client, _ := newTestClient(t, server)

	_, err := Fetch(context.Background(), client)
	if errors.Is(err, errWheelTooLarge) {
		t.Fatal("a body of exactly maxWheelSize must not be refused for size")
	}
	if !errors.Is(err, errWheelChecksum) {
		t.Fatalf("Fetch(exact-size wrong body) error = %v, want errWheelChecksum", err)
	}
}

func TestFetchRefusesWrongChecksum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not the pinned wheel"))
	}))
	defer server.Close()
	client, _ := newTestClient(t, server)

	_, err := Fetch(context.Background(), client)
	if !errors.Is(err, errWheelChecksum) {
		t.Fatalf("Fetch(wrong body) error = %v, want errWheelChecksum", err)
	}
}

func TestFetchPropagatesContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("unreached"))
	}))
	defer server.Close()
	client, _ := newTestClient(t, server)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Fetch(ctx, client)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Fetch(canceled context) error = %v, want context.Canceled", err)
	}
}

func TestWheelCachePathRequiresACacheDirectory(t *testing.T) {
	t.Setenv("HOME", "")
	t.Setenv("XDG_CACHE_HOME", "")
	if _, err := WheelCachePath(); err == nil {
		t.Fatal("WheelCachePath with no HOME or XDG_CACHE_HOME = nil error, want one")
	}
}

type failingReader struct{ err error }

func (r failingReader) Read([]byte) (int, error) { return 0, r.err }

func TestReadWheel(t *testing.T) {
	wheel := loadCachedWheel(t)
	readErr := errors.New("disk on fire")
	tests := []struct {
		name    string
		r       io.Reader
		wantErr error
	}{
		{"pinned wheel", bytes.NewReader(wheel), nil},
		{"oversized", bytes.NewReader(bytes.Repeat([]byte{'x'}, maxWheelSize+1)), errWheelTooLarge},
		{"exactly the maximum, wrong bytes", bytes.NewReader(bytes.Repeat([]byte{'x'}, maxWheelSize)), errWheelChecksum},
		{"wrong bytes", strings.NewReader("not the pinned wheel"), errWheelChecksum},
		{"read failure", failingReader{readErr}, readErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadWheel(tt.r)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ReadWheel error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && !bytes.Equal(got, wheel) {
				t.Fatal("ReadWheel returned bytes that differ from the wheel")
			}
			if tt.wantErr != nil && got != nil {
				t.Fatalf("ReadWheel returned %d bytes alongside an error", len(got))
			}
		})
	}
}
