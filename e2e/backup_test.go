package e2e

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/lock"
)

// writeBackupFixture stages a populated library directory and both
// source databases for a clean backup run.
func (e *testEnv) writeBackupFixture() {
	e.t.Helper()
	libraryDir := e.env["DUBPLATE_LIBRARY_DIR"]
	if err := os.MkdirAll(libraryDir, 0o755); err != nil {
		e.t.Fatalf("mkdir library: %v", err)
	}
	if err := os.WriteFile(filepath.Join(libraryDir, "track.flac"), []byte("audio"), 0o644); err != nil {
		e.t.Fatalf("write library file: %v", err)
	}
	if err := os.WriteFile(e.beetsDB, []byte("beets-db"), 0o644); err != nil {
		e.t.Fatalf("write beets db: %v", err)
	}
	if err := os.WriteFile(e.env["DUBPLATE_NAVIDROME_DB"], []byte("navidrome-db"), 0o644); err != nil {
		e.t.Fatalf("write navidrome db: %v", err)
	}
}

func TestBackupHappyPath(t *testing.T) {
	e := newTestEnv(t)
	e.writeBackupFixture()
	rcloneLog := filepath.Join(e.root, "rclone-args.log")
	e.env["STUB_RCLONE_LOG"] = rcloneLog

	output, code := e.run("backup")
	if code != 0 {
		t.Fatalf("backup exited %d, want 0: %s", code, output)
	}

	data, err := os.ReadFile(rcloneLog)
	if err != nil {
		t.Fatalf("read rclone args log: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("rclone invocations = %d, want 2 (library, state): %q", len(lines), data)
	}
	if !strings.Contains(lines[0], "library") {
		t.Errorf("first rclone invocation = %q, want the library tree", lines[0])
	}
	if !strings.Contains(lines[1], "state") {
		t.Errorf("second rclone invocation = %q, want the state tree", lines[1])
	}

	if _, err := os.Stat(filepath.Join(e.env["DUBPLATE_BACKUP_STAGING"], "library.db")); err != nil {
		t.Errorf("beets db snapshot missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(e.env["DUBPLATE_BACKUP_STAGING"], "navidrome.db")); err != nil {
		t.Errorf("navidrome db snapshot missing: %v", err)
	}

	select {
	case hit := <-e.pingHits:
		if hit.method != http.MethodGet {
			t.Errorf("ping method = %s, want GET", hit.method)
		}
	default:
		t.Error("no success ping received")
	}
}

func TestBackupLibraryMissing(t *testing.T) {
	e := newTestEnv(t)
	// No library directory, no databases: writeBackupFixture is
	// deliberately skipped.

	output, code := e.run("backup")
	if code == 0 {
		t.Fatalf("backup exited 0, want nonzero with no library directory: %s", output)
	}

	select {
	case hit := <-e.pingHits:
		t.Errorf("in-process ping sent on a per-run failure: %+v (OnFailure= owns this, not the binary)", hit)
	default:
	}
}

func TestBackupLockHeld(t *testing.T) {
	e := newTestEnv(t)
	e.writeBackupFixture()

	stateDir := e.env["DUBPLATE_STATE_DIR"]
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	held, err := lock.Acquire(filepath.Join(stateDir, "backup.lock"))
	if err != nil {
		t.Fatalf("acquire backup lock: %v", err)
	}
	defer func() { _ = held.Release() }()

	output, code := e.run("backup")
	if code != 0 {
		t.Fatalf("backup exited %d, want 0 on a held lock: %s", code, output)
	}

	if _, err := os.Stat(e.env["DUBPLATE_BACKUP_STAGING"]); !os.IsNotExist(err) {
		t.Error("backup staging created despite a held lock")
	}

	select {
	case hit := <-e.pingHits:
		if hit.method != http.MethodGet {
			t.Errorf("ping method = %s, want GET", hit.method)
		}
	default:
		t.Error("no success ping sent on a held lock")
	}
}
