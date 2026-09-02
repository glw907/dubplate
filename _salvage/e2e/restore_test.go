package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeRestoreFixture stages a bucket stand-in for `dubplate restore`:
// a library file, both database snapshots, and one dated tree under
// `_archive` for each of library and state, matching backup's own
// bucket layout.
func (e *testEnv) writeRestoreFixture() {
	e.t.Helper()
	bucketDir := filepath.Join(e.root, "bucket")

	write := func(rel, content string) {
		path := filepath.Join(bucketDir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			e.t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			e.t.Fatalf("write %s: %v", rel, err)
		}
	}

	write("library/alice/Good Album/track1.flac", "audio")
	write("state/library.db", "beets-db")
	write("state/navidrome.db", "navidrome-db")
	write("_archive/library/2026-08-30/replaced.flac", "old-audio")
	write("_archive/state/2026-08-30/library.db", "old-beets-db")

	e.env["STUB_RCLONE_BUCKET_DIR"] = bucketDir
}

func TestRestoreHappyPath(t *testing.T) {
	e := newTestEnv(t)
	e.writeRestoreFixture()
	to := filepath.Join(e.root, "restored")

	output, code := e.run("restore", "--to", to)
	if code != 0 {
		t.Fatalf("restore exited %d, want 0: %s", code, output)
	}

	if _, err := os.Stat(filepath.Join(to, "library", "alice", "Good Album", "track1.flac")); err != nil {
		t.Errorf("restored library file missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(to, "state", "library.db")); err != nil {
		t.Errorf("restored library.db missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(to, "state", "navidrome.db")); err != nil {
		t.Errorf("restored navidrome.db missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(to, "_archive")); !os.IsNotExist(err) {
		t.Error("_archive was pulled into the target, want it only listed")
	}

	if !strings.Contains(output, "library: 1 file(s) restored") {
		t.Errorf("output missing library restored count: %s", output)
	}
	if !strings.Contains(output, "state: 2 file(s) restored") {
		t.Errorf("output missing state restored count: %s", output)
	}
	if !strings.Contains(output, "library/2026-08-30") || !strings.Contains(output, "state/2026-08-30") {
		t.Errorf("output missing archive tree listing: %s", output)
	}
	if !strings.Contains(output, "library.db: integrity check ok") || !strings.Contains(output, "navidrome.db: integrity check ok") {
		t.Errorf("output missing database integrity results: %s", output)
	}
}

func TestRestoreRefusesNonEmptyTargetWithoutForce(t *testing.T) {
	e := newTestEnv(t)
	e.writeRestoreFixture()
	to := filepath.Join(e.root, "restored")
	if err := os.MkdirAll(to, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(to, "preexisting.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write preexisting file: %v", err)
	}

	output, code := e.run("restore", "--to", to)
	if code == 0 {
		t.Fatalf("restore exited 0 against a non-empty target without --force: %s", output)
	}
	if !strings.Contains(output, "not empty") {
		t.Errorf("output = %q, want it to name the non-empty target refusal", output)
	}

	if _, err := os.Stat(filepath.Join(to, "state")); !os.IsNotExist(err) {
		t.Error("restore wrote into the refused target")
	}
}

func TestRestoreForceAllowsNonEmptyTarget(t *testing.T) {
	e := newTestEnv(t)
	e.writeRestoreFixture()
	to := filepath.Join(e.root, "restored")
	if err := os.MkdirAll(to, 0o755); err != nil {
		t.Fatalf("mkdir target: %v", err)
	}
	if err := os.WriteFile(filepath.Join(to, "preexisting.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write preexisting file: %v", err)
	}

	output, code := e.run("restore", "--to", to, "--force")
	if code != 0 {
		t.Fatalf("restore exited %d, want 0 with --force: %s", code, output)
	}
	if _, err := os.Stat(filepath.Join(to, "state", "library.db")); err != nil {
		t.Errorf("restored library.db missing: %v", err)
	}
}

func TestRestoreIntegrityCheckFailureAborts(t *testing.T) {
	e := newTestEnv(t)
	e.writeRestoreFixture()
	to := filepath.Join(e.root, "restored")
	e.env["STUB_SQLITE_CORRUPT_DB"] = filepath.Join(to, "state", "library.db")

	output, code := e.run("restore", "--to", to)
	if code == 0 {
		t.Fatalf("restore exited 0 with a corrupt restored database: %s", output)
	}
	if !strings.Contains(output, "library.db: integrity check FAILED") {
		t.Errorf("output = %q, want the failed integrity check reported", output)
	}
	// The tree pulls still ran and are reported even though the run as a
	// whole failed: a hard stop on the database check must not hide what
	// was already restored.
	if !strings.Contains(output, "library: 1 file(s) restored") {
		t.Errorf("output missing the library tree report despite the later failure: %s", output)
	}
}
