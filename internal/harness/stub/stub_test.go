package stub

import (
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// newTestServer writes body under dir/responses/001.json and starts a
// Server for a one-request case referencing it at pathQuery.
func newTestServer(t *testing.T, pathQuery string, body []byte) *Server {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), body, 0o644); err != nil {
		t.Fatalf("write response fixture: %v", err)
	}

	c := &cassette.Case{
		ID: "stub-test",
		Requests: []cassette.Request{
			{Method: "GET", PathQuery: pathQuery, Response: "responses/001.json"},
		},
	}
	s, err := New(c, dir)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func get(t *testing.T, addr, path string) *http.Response {
	t.Helper()
	resp, err := http.Get("http://" + addr + path) //nolint:gosec // fixed loopback test address
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	return resp
}

func TestServerServesHit(t *testing.T) {
	want := []byte(`{"release":{"id":"abc"}}`)
	s := newTestServer(t, "/release/abc?fmt=json&inc=media", want)

	resp := get(t, s.Addr(), "/release/abc?fmt=json&inc=media")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	got, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("body = %q, want %q", got, want)
	}
	if len(s.Misses()) != 0 {
		t.Errorf("Misses() = %v, want none for a hit", s.Misses())
	}
}

func TestServerCanonicalizesReorderedQuery(t *testing.T) {
	want := []byte(`{"ok":true}`)
	s := newTestServer(t, "/release/abc?fmt=json&inc=media", want)

	// The pairs arrive in the opposite order from how the case stored
	// them; canonicalization must still resolve this to the same key.
	resp := get(t, s.Addr(), "/release/abc?inc=media&fmt=json")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200 for a reordered-but-equivalent query", resp.StatusCode)
	}
	if len(s.Misses()) != 0 {
		t.Errorf("Misses() = %v, want none: reordering is not a miss", s.Misses())
	}
}

func TestServerRejectsTrailingSlashNearMiss(t *testing.T) {
	s := newTestServer(t, "/release/abc?fmt=json", []byte(`{"ok":true}`))

	// Same query, but the path carries a trailing slash the case never
	// logged: canonicalization must not collapse this into a hit.
	resp := get(t, s.Addr(), "/release/abc/?fmt=json")
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("status = %d, want 404 for a trailing-slash near miss", resp.StatusCode)
	}
	misses := s.Misses()
	if len(misses) != 1 {
		t.Fatalf("Misses() = %v, want exactly one", misses)
	}
	if misses[0].PathQuery != "/release/abc/?fmt=json" {
		t.Errorf("recorded miss path_query = %q, want %q", misses[0].PathQuery, "/release/abc/?fmt=json")
	}
}

func TestServerRecordsMissesInRequestOrder(t *testing.T) {
	s := newTestServer(t, "/release/abc?fmt=json", []byte(`{"ok":true}`))

	_ = get(t, s.Addr(), "/release/first?fmt=json").Body.Close()
	_ = get(t, s.Addr(), "/release/second?fmt=json").Body.Close()

	misses := s.Misses()
	if len(misses) != 2 {
		t.Fatalf("Misses() = %v, want 2", misses)
	}
	if misses[0].PathQuery != "/release/first?fmt=json" || misses[1].PathQuery != "/release/second?fmt=json" {
		t.Errorf("Misses() = %v, want first then second in request order", misses)
	}
	if misses[0].Method != "GET" {
		t.Errorf("recorded miss method = %q, want GET", misses[0].Method)
	}
}

func TestServerEphemeralPortLifecycle(t *testing.T) {
	s := newTestServer(t, "/release/abc?fmt=json", []byte(`{"ok":true}`))

	addr := s.Addr()
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		t.Fatalf("Addr() = %q is not host:port: %v", addr, err)
	}
	if host != "127.0.0.1" {
		t.Errorf("Addr() host = %q, want 127.0.0.1", host)
	}
	if portStr == "0" || portStr == "" {
		t.Errorf("Addr() port = %q, want a real OS-assigned port", portStr)
	}

	resp := get(t, addr, "/release/abc?fmt=json")
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status before Close = %d, want 200", resp.StatusCode)
	}

	if err := s.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	if _, err := (&http.Client{Timeout: 2 * time.Second}).Get("http://" + addr + "/release/abc?fmt=json"); err == nil {
		t.Error("request after Close succeeded, want a connection error")
	}
}
