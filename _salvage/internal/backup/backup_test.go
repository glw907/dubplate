package backup

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/ping"
)

// newPingTestServer starts a local HTTP server that increments *hits on
// every request and returns its URL, standing in for a pings provider.
func newPingTestServer(t *testing.T, hits *int) string {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		*hits++
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(server.Close)
	return server.URL
}

// writeStubScript writes an executable shell script at dir/name and
// returns its path, the SqlitePath/RclonePath test seams.
func writeStubScript(t *testing.T, dir, name, script string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", name, err)
	}
	return path
}

// newTestRunner builds a Runner over a fresh temp estate: a populated
// library directory and both databases, ready for a clean run. The
// caller mutates the returned Runner and fields before calling Run.
func newTestRunner(t *testing.T, sqliteScript, rcloneScript string) (*Runner, *config.Config) {
	t.Helper()
	root := t.TempDir()
	libraryDir := filepath.Join(root, "library")
	stateDir := filepath.Join(root, "state")
	backupStaging := filepath.Join(root, "backup-staging")
	binDir := t.TempDir()

	if err := os.MkdirAll(libraryDir, 0o755); err != nil {
		t.Fatalf("mkdir library: %v", err)
	}
	if err := os.WriteFile(filepath.Join(libraryDir, "track.flac"), []byte("audio"), 0o644); err != nil {
		t.Fatalf("write library file: %v", err)
	}
	if err := os.MkdirAll(stateDir, 0o755); err != nil {
		t.Fatalf("mkdir state: %v", err)
	}
	beetsDB := filepath.Join(stateDir, "library.db")
	navidromeDB := filepath.Join(root, "navidrome.db")
	if err := os.WriteFile(beetsDB, []byte("beets-db"), 0o644); err != nil {
		t.Fatalf("write beets db: %v", err)
	}
	if err := os.WriteFile(navidromeDB, []byte("navidrome-db"), 0o644); err != nil {
		t.Fatalf("write navidrome db: %v", err)
	}

	cfg := &config.Config{
		BeetsDB:     beetsDB,
		NavidromeDB: navidromeDB,
		Paths: config.Paths{
			LibraryDir:    libraryDir,
			StateDir:      stateDir,
			BackupStaging: backupStaging,
		},
		R2: config.R2Config{
			AccountID:       "acct",
			Bucket:          "bucket",
			Remote:          "r2",
			AccessKeyID:     "key",
			SecretAccessKey: "secret",
		},
		PingURLs: map[string]string{},
	}

	r := &Runner{
		Cfg:        cfg,
		Ping:       ping.NewClient(),
		SqlitePath: writeStubScript(t, binDir, "sqlite3", sqliteScript),
		RclonePath: writeStubScript(t, binDir, "rclone", rcloneScript),
		Now:        func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) },
	}
	return r, cfg
}

const okScript = ": # ok"

func TestRunHappyPath(t *testing.T) {
	rcloneLog := filepath.Join(t.TempDir(), "rclone-args.log")
	rcloneScript := `echo "$@" >> "` + rcloneLog + `"`

	r, _ := newTestRunner(t, okScript, rcloneScript)

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	data, err := os.ReadFile(rcloneLog)
	if err != nil {
		t.Fatalf("read rclone args log: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("rclone invocations = %d, want 2 (library, state): %q", len(lines), data)
	}
	if !strings.Contains(lines[0], "r2:bucket/library") || !strings.Contains(lines[0], "--backup-dir r2:bucket/_archive/library/2026-08-31") {
		t.Errorf("library sync args = %q, want the library tree and its archive dir", lines[0])
	}
	if !strings.Contains(lines[1], "r2:bucket/state") || !strings.Contains(lines[1], "--backup-dir r2:bucket/_archive/state/2026-08-31") {
		t.Errorf("state sync args = %q, want the state tree and its archive dir", lines[1])
	}
}

func TestRunLibraryMissing(t *testing.T) {
	r, cfg := newTestRunner(t, okScript, okScript)
	if err := os.RemoveAll(cfg.Paths.LibraryDir); err != nil {
		t.Fatalf("remove library dir: %v", err)
	}

	err := r.Run(context.Background())
	if !errors.Is(err, ErrLibraryMissing) {
		t.Fatalf("Run() error = %v, want ErrLibraryMissing", err)
	}
}

func TestRunLibraryEmpty(t *testing.T) {
	r, cfg := newTestRunner(t, okScript, okScript)
	if err := os.Remove(filepath.Join(cfg.Paths.LibraryDir, "track.flac")); err != nil {
		t.Fatalf("empty library dir: %v", err)
	}

	err := r.Run(context.Background())
	if !errors.Is(err, ErrLibraryEmpty) {
		t.Fatalf("Run() error = %v, want ErrLibraryEmpty", err)
	}
}

func TestRunMissingDatabaseSkipsNotAborts(t *testing.T) {
	sqliteLog := filepath.Join(t.TempDir(), "sqlite-args.log")
	sqliteScript := `echo "$@" >> "` + sqliteLog + `"`

	r, cfg := newTestRunner(t, sqliteScript, okScript)
	if err := os.Remove(cfg.NavidromeDB); err != nil {
		t.Fatalf("remove navidrome db: %v", err)
	}

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() unexpected error with a missing (not yet created) database: %v", err)
	}

	data, err := os.ReadFile(sqliteLog)
	if err != nil {
		t.Fatalf("read sqlite3 args log: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Fatalf("sqlite3 invocations = %d, want 1 (beets db only, navidrome db skipped): %q", len(lines), data)
	}
	if !strings.Contains(lines[0], cfg.BeetsDB) {
		t.Errorf("sqlite3 invocation = %q, want the beets db named", lines[0])
	}
}

func TestRunSqliteFailureAborts(t *testing.T) {
	failScript := `echo "sqlite3: forced failure" >&2; exit 1`
	rcloneLog := filepath.Join(t.TempDir(), "rclone-args.log")
	rcloneScript := `echo "$@" >> "` + rcloneLog + `"`

	r, _ := newTestRunner(t, failScript, rcloneScript)

	if err := r.Run(context.Background()); err == nil {
		t.Fatal("Run() error = nil, want the sqlite3 failure propagated")
	}
	if _, err := os.Stat(rcloneLog); !os.IsNotExist(err) {
		t.Error("rclone was invoked after a sqlite3 snapshot failure")
	}
}

func TestRunRcloneFailureAborts(t *testing.T) {
	failScript := `echo "rclone: forced failure" >&2; exit 1`
	r, _ := newTestRunner(t, okScript, failScript)

	if err := r.Run(context.Background()); err == nil {
		t.Fatal("Run() error = nil, want the rclone failure propagated")
	}
}

func TestRunLockHeldPingsSuccessAndTouchesNothing(t *testing.T) {
	hits := 0
	server := newPingTestServer(t, &hits)

	r, cfg := newTestRunner(t, okScript, okScript)
	cfg.PingURLs[pingSlug] = server

	held, err := lock.Acquire(lockPath(cfg.Paths.StateDir))
	if err != nil {
		t.Fatalf("acquire backup lock: %v", err)
	}
	defer func() { _ = held.Release() }()

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() unexpected error on a held lock: %v", err)
	}
	if _, err := os.Stat(cfg.Paths.BackupStaging); !os.IsNotExist(err) {
		t.Error("backup staging created despite a held lock")
	}
	if hits != 1 {
		t.Errorf("success pings received = %d, want 1", hits)
	}
}

func TestRunClearsStaleStagingContentsButKeepsDir(t *testing.T) {
	r, cfg := newTestRunner(t, okScript, okScript)

	if err := os.MkdirAll(cfg.Paths.BackupStaging, 0o755); err != nil {
		t.Fatalf("mkdir backup staging: %v", err)
	}
	stalePath := filepath.Join(cfg.Paths.BackupStaging, "stale.db")
	if err := os.WriteFile(stalePath, []byte("stale"), 0o644); err != nil {
		t.Fatalf("write stale file: %v", err)
	}
	before, err := os.Stat(cfg.Paths.BackupStaging)
	if err != nil {
		t.Fatalf("stat backup staging before run: %v", err)
	}

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if _, err := os.Stat(stalePath); !os.IsNotExist(err) {
		t.Errorf("stat stale.db after Run() = %v, want IsNotExist", err)
	}
	after, err := os.Stat(cfg.Paths.BackupStaging)
	if err != nil {
		t.Fatalf("stat backup staging after run: %v", err)
	}
	if !os.SameFile(before, after) {
		t.Error("backup staging directory was recreated rather than left in place")
	}
}

func TestRunCreatesMissingStagingDir(t *testing.T) {
	r, cfg := newTestRunner(t, okScript, okScript)
	if _, err := os.Stat(cfg.Paths.BackupStaging); !os.IsNotExist(err) {
		t.Fatalf("backup staging exists before run: %v", err)
	}

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	info, err := os.Stat(cfg.Paths.BackupStaging)
	if err != nil {
		t.Fatalf("stat backup staging after run: %v", err)
	}
	if !info.IsDir() {
		t.Error("backup staging is not a directory")
	}
}

func TestRunPingsSuccessOnCleanRun(t *testing.T) {
	hits := 0
	server := newPingTestServer(t, &hits)

	r, cfg := newTestRunner(t, okScript, okScript)
	cfg.PingURLs[pingSlug] = server

	if err := r.Run(context.Background()); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	if hits != 1 {
		t.Errorf("success pings received = %d, want 1", hits)
	}
}
