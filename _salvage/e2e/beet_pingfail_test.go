package e2e

import (
	"net/http"
	"testing"
	"time"
)

// --- dubplate beet -----------------------------------------------------

// TestBeetHoldsLockAgainstOverlappingImport proves the shape T10 exists
// for: `dubplate beet` blocks on the import lock, execs into the real
// beet with the lock fd riding the exec, and a `dubplate import` that
// starts while beet is still running sees the lock held and skips rather
// than racing it.
func TestBeetHoldsLockAgainstOverlappingImport(t *testing.T) {
	e := newTestEnv(t)
	e.writeUploadedFlac("alice", "Good Album")
	e.env["STUB_BEET_SLEEP_MS"] = "1000"

	holdDir := t.TempDir() // beet stub walks its last arg; an empty real dir is enough
	beetCmd := e.startAsync("beet", "--", holdDir)

	// Give `dubplate beet` time to acquire the lock and exec into the
	// sleeping stub before the overlapping import starts; the lock
	// acquisition itself happens well before the stub's own sleep.
	time.Sleep(300 * time.Millisecond)

	output, code := e.run("import")
	if code != 0 {
		t.Fatalf("import exited %d, want 0 while beet holds the lock: %s", code, output)
	}

	report := e.readReport()
	if !report.LockHeld {
		t.Error("report.LockHeld = false, want true while dubplate beet holds the lock")
	}

	if err := beetCmd.Wait(); err != nil {
		t.Errorf("dubplate beet failed: %v", err)
	}
}

// --- dubplate ping-fail --------------------------------------------------

func TestPingFailPostsToFailEndpoint(t *testing.T) {
	e := newTestEnv(t)

	output, code := e.run("ping-fail", "import")
	if code != 0 {
		t.Fatalf("ping-fail exited %d, want 0: %s", code, output)
	}

	select {
	case hit := <-e.pingHits:
		if hit.method != http.MethodPost {
			t.Errorf("ping method = %s, want POST", hit.method)
		}
		if hit.path != "/fail" {
			t.Errorf("ping path = %q, want /fail", hit.path)
		}
	default:
		t.Error("no failure ping received")
	}
}

func TestPingFailUnsetURLIsNoOp(t *testing.T) {
	e := newTestEnv(t)

	output, code := e.run("ping-fail", "nonexistent-check")
	if code != 0 {
		t.Fatalf("ping-fail exited %d, want 0 on an unset ping URL: %s", code, output)
	}

	select {
	case hit := <-e.pingHits:
		t.Errorf("ping sent for a slug with no configured URL: %+v", hit)
	default:
	}
}

func TestPingFailUnreachableURLExitsZero(t *testing.T) {
	e := newTestEnv(t)
	// Well-formed but nothing listens on this port: connection refused,
	// exhausted after the client's own retries. Delivery failure must not
	// surface as a nonzero exit -- only a malformed URL does that.
	e.env["DUBPLATE_PING_UNREACHABLE_URL"] = "http://127.0.0.1:1/ping"

	output, code := e.run("ping-fail", "unreachable")
	if code != 0 {
		t.Fatalf("ping-fail exited %d, want 0 on an unreachable ping URL: %s", code, output)
	}
}

func TestPingFailMalformedURLExitsNonzero(t *testing.T) {
	e := newTestEnv(t)
	// A raw DEL (0x7f) control byte: net/url rejects it outright, before
	// any network attempt.
	e.env["DUBPLATE_PING_BADSLUG_URL"] = "http://example.invalid/\x7f"

	output, code := e.run("ping-fail", "badslug")
	if code == 0 {
		t.Fatalf("ping-fail exited 0, want nonzero on a malformed ping URL: %s", output)
	}
}

func TestPingFailIgnoresBrokenConfig(t *testing.T) {
	e := newTestEnv(t)
	// Break an unrelated required setting outright; ping-fail must still
	// work, since it never calls config.Load.
	delete(e.env, "DUBPLATE_SMTP_HOST")

	output, code := e.run("ping-fail", "import")
	if code != 0 {
		t.Fatalf("ping-fail exited %d, want 0 despite an unrelated missing config field: %s", code, output)
	}

	select {
	case <-e.pingHits:
	default:
		t.Error("no failure ping received despite a valid ping URL")
	}
}
