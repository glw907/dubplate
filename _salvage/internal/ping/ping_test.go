package ping

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewClientDefaults(t *testing.T) {
	c := NewClient()
	if c.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s (the shell's curl --max-time budget)", c.Timeout)
	}
	if c.Retries != 3 {
		t.Errorf("Retries = %d, want 3 (the shell's curl --retry budget)", c.Retries)
	}
}

func TestClientEmptyURLIsNoop(t *testing.T) {
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
	}))
	defer srv.Close()

	c := NewClient()
	if err := c.Success(context.Background(), ""); err != nil {
		t.Errorf(`Success("") = %v, want nil`, err)
	}
	if err := c.Fail(context.Background(), ""); err != nil {
		t.Errorf(`Fail("") = %v, want nil`, err)
	}
	if n := hits.Load(); n != 0 {
		t.Errorf("server saw %d requests for an empty URL, want 0", n)
	}
}

func TestClientSuccessGETs(t *testing.T) {
	var gotMethod, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotMethod, gotPath = r.Method, r.URL.Path
	}))
	defer srv.Close()

	c := NewClient()
	if err := c.Success(context.Background(), srv.URL+"/ping/abc"); err != nil {
		t.Fatalf("Success: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Errorf("Success sent method %q, want GET", gotMethod)
	}
	if gotPath != "/ping/abc" {
		t.Errorf("Success sent path %q, want /ping/abc", gotPath)
	}
}

// TestClientFailPOSTsWithFailSuffix is the regression guard named in the
// task criteria: the pings Worker routes /fail on POST only and 404s a
// bare GET, so a Fail call that regresses to GET would be silently
// dropped by the real provider even though this test stub would accept
// it. Asserting the recorded method, not just a nil error, is what makes
// that regression visible.
func TestClientFailPOSTsWithFailSuffix(t *testing.T) {
	tests := []struct {
		name     string
		suffix   string
		wantPath string
	}{
		{"no trailing slash", "/ping/abc", "/ping/abc/fail"},
		{"trailing slash stripped before appending fail", "/ping/abc/", "/ping/abc/fail"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotMethod, gotPath string
			srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
				gotMethod, gotPath = r.Method, r.URL.Path
			}))
			defer srv.Close()

			c := NewClient()
			if err := c.Fail(context.Background(), srv.URL+tt.suffix); err != nil {
				t.Fatalf("Fail: %v", err)
			}
			if gotMethod != http.MethodPost {
				t.Errorf("Fail sent method %q, want POST", gotMethod)
			}
			if gotPath != tt.wantPath {
				t.Errorf("Fail sent path %q, want %q", gotPath, tt.wantPath)
			}
		})
	}
}

func TestClientRetriesThenSucceeds(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if attempts.Add(1) < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient()
	if err := c.Success(context.Background(), srv.URL); err != nil {
		t.Fatalf("Success: %v", err)
	}
	if n := attempts.Load(); n != 3 {
		t.Errorf("server saw %d attempts, want 3 (2 failures then a success, within the 3-retry budget)", n)
	}
}

func TestClientExhaustsRetries(t *testing.T) {
	var attempts atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClient()
	if err := c.Success(context.Background(), srv.URL); err == nil {
		t.Fatal("Success = nil, want an error after exhausting retries")
	}
	if want := int32(c.Retries + 1); attempts.Load() != want {
		t.Errorf("server saw %d attempts, want %d (1 initial + %d retries)", attempts.Load(), want, c.Retries)
	}
}
