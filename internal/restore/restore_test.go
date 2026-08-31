package restore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// writeStubScript writes an executable shell script at dir/name and
// returns its path, the SqlitePath/RclonePath test seams -- mirroring
// [backup]'s own test helper.
func writeStubScript(t *testing.T, dir, name, script string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", name, err)
	}
	return path
}

const okSqliteScript = `
db="$1"
case "$2" in
  "PRAGMA integrity_check;") echo ok ;;
esac
`

// rcloneCopyStub copies a fixed fixture tree into whatever destination
// `copy` is given and logs every invocation to STUB_RCLONE_LOG, so a
// test can both assert on the args rclone received and see files land
// locally. `lsjson` answers with the canned bodies in bodies, keyed by
// the trailing path segment of its src argument ("library", "state",
// "_archive"); an unmatched src answers "[]".
func rcloneCopyStub(fixtureDir string, bodies map[string]string) string {
	var cases strings.Builder
	for suffix, body := range bodies {
		cases.WriteString("    *" + suffix + ") cat <<'JSON'\n" + body + "\nJSON\n    ;;\n")
	}
	return `
echo "$@" >> "$STUB_RCLONE_LOG"
cmd="$1"; shift
case "$cmd" in
  copy)
    src="$1"; dst="$2"
    mkdir -p "$dst"
    case "$src" in
      *library) cp "` + fixtureDir + `"/library/* "$dst"/ 2>/dev/null || true ;;
      *state) cp "` + fixtureDir + `"/state/* "$dst"/ 2>/dev/null || true ;;
    esac
    ;;
  lsjson)
    for a in "$@"; do last="$a"; done
    case "$last" in
` + cases.String() + `
      *) echo "[]" ;;
    esac
    ;;
esac
`
}

// newTestRunner builds a Runner and its Cfg over a fresh temp dir, with
// stub rclone/sqlite3 binaries. bodies keys the lsjson canned responses
// by tree suffix ("library", "state", "_archive").
func newTestRunner(t *testing.T, sqliteScript string, bodies map[string]string) (*Runner, string) {
	t.Helper()
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "fixture")
	for _, tree := range []string{"library", "state"} {
		if err := os.MkdirAll(filepath.Join(fixtureDir, tree), 0o755); err != nil {
			t.Fatalf("mkdir fixture %s: %v", tree, err)
		}
	}
	binDir := t.TempDir()
	rcloneLog := filepath.Join(root, "rclone-args.log")

	cfg := &config.Config{
		R2: config.R2Config{
			AccountID:       "acct",
			Bucket:          "bucket",
			Remote:          "r2",
			AccessKeyID:     "key",
			SecretAccessKey: "secret",
		},
	}

	r := &Runner{
		Cfg:        cfg,
		SqlitePath: writeStubScript(t, binDir, "sqlite3", sqliteScript),
		RclonePath: writeStubScript(t, binDir, "rclone", rcloneCopyStub(fixtureDir, bodies)),
		Now:        func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) },
	}
	t.Setenv("STUB_RCLONE_LOG", rcloneLog)
	return r, fixtureDir
}

func TestRunHappyPath(t *testing.T) {
	bodies := map[string]string{
		"library": `[{"Path":"track1.flac","IsDir":false}]`,
		"state":   `[{"Path":"library.db","IsDir":false},{"Path":"navidrome.db","IsDir":false}]`,
		"_archive": `[{"Path":"library","IsDir":true},{"Path":"library/2026-08-30","IsDir":true},` +
			`{"Path":"state","IsDir":true},{"Path":"state/2026-08-30","IsDir":true}]`,
	}
	r, fixtureDir := newTestRunner(t, okSqliteScript, bodies)
	if err := os.WriteFile(filepath.Join(fixtureDir, "library", "track1.flac"), []byte("audio"), 0o644); err != nil {
		t.Fatalf("write fixture track: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "state", "library.db"), []byte("beets-db"), 0o644); err != nil {
		t.Fatalf("write fixture library.db: %v", err)
	}
	if err := os.WriteFile(filepath.Join(fixtureDir, "state", "navidrome.db"), []byte("navidrome-db"), 0o644); err != nil {
		t.Fatalf("write fixture navidrome.db: %v", err)
	}

	to := filepath.Join(t.TempDir(), "restored")
	report, err := r.Run(context.Background(), Options{To: to})
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(report.Trees) != 2 {
		t.Fatalf("report.Trees = %d entries, want 2", len(report.Trees))
	}
	if report.Trees[0].Name != "library" || report.Trees[0].RestoredFiles != 1 {
		t.Errorf("library tree report = %+v, want name library, 1 restored file", report.Trees[0])
	}
	if len(report.Trees[0].BucketOnly) != 0 {
		t.Errorf("library BucketOnly = %v, want none (the only bucket file was restored)", report.Trees[0].BucketOnly)
	}
	if report.Trees[1].Name != "state" || report.Trees[1].RestoredFiles != 2 {
		t.Errorf("state tree report = %+v, want name state, 2 restored files", report.Trees[1])
	}

	if len(report.Archives) != 2 {
		t.Fatalf("report.Archives = %d entries, want 2 (tree/date pairs, depth-1 dirs excluded): %+v", len(report.Archives), report.Archives)
	}
	want := []ArchiveEntry{{Tree: "library", Date: "2026-08-30"}, {Tree: "state", Date: "2026-08-30"}}
	for i, a := range want {
		if report.Archives[i] != a {
			t.Errorf("report.Archives[%d] = %+v, want %+v", i, report.Archives[i], a)
		}
	}

	if len(report.Databases) != 2 {
		t.Fatalf("report.Databases = %d entries, want 2", len(report.Databases))
	}
	for _, db := range report.Databases {
		if !db.Present || !db.OK {
			t.Errorf("database check %+v, want present and ok", db)
		}
	}
}

func TestRunNeverInvokesSyncOrPullsArchive(t *testing.T) {
	bodies := map[string]string{
		"library":  `[]`,
		"state":    `[]`,
		"_archive": `[{"Path":"library/2026-08-30","IsDir":true}]`,
	}
	root := t.TempDir()
	fixtureDir := filepath.Join(root, "fixture")
	for _, tree := range []string{"library", "state"} {
		if err := os.MkdirAll(filepath.Join(fixtureDir, tree), 0o755); err != nil {
			t.Fatalf("mkdir fixture %s: %v", tree, err)
		}
	}
	rcloneLog := filepath.Join(root, "rclone-args.log")
	t.Setenv("STUB_RCLONE_LOG", rcloneLog)
	binDir := t.TempDir()
	r := &Runner{
		Cfg:        &config.Config{R2: config.R2Config{AccountID: "acct", Bucket: "bucket", Remote: "r2", AccessKeyID: "key", SecretAccessKey: "secret"}},
		SqlitePath: writeStubScript(t, binDir, "sqlite3", okSqliteScript),
		RclonePath: writeStubScript(t, binDir, "rclone", rcloneCopyStub(fixtureDir, bodies)),
		Now:        func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) },
	}

	to := filepath.Join(t.TempDir(), "restored")
	if _, err := r.Run(context.Background(), Options{To: to}); err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	data, err := os.ReadFile(rcloneLog)
	if err != nil {
		t.Fatalf("read rclone args log: %v", err)
	}
	for line := range strings.SplitSeq(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.HasPrefix(line, "sync") {
			t.Errorf("rclone invoked with sync, want copy/lsjson only: %q", line)
		}
		if strings.HasPrefix(line, "copy") && strings.Contains(line, "_archive") {
			t.Errorf("rclone copy invoked against _archive, want it only listed: %q", line)
		}
	}
}

func TestRunBucketOnlyDifference(t *testing.T) {
	// lsjson reports a file the copy stub never places locally (it only
	// copies files actually present under the fixture tree, and this
	// case's fixture tree is empty), proving BucketOnly surfaces a real
	// gap between what the bucket claims and what arrived.
	bodies := map[string]string{
		"library": `[{"Path":"missing.flac","IsDir":false}]`,
		"state":   `[]`,
	}
	r, _ := newTestRunner(t, okSqliteScript, bodies)

	to := filepath.Join(t.TempDir(), "restored")
	report, err := r.Run(context.Background(), Options{To: to})
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}

	if len(report.Trees[0].BucketOnly) != 1 || report.Trees[0].BucketOnly[0] != "missing.flac" {
		t.Errorf("library BucketOnly = %v, want [missing.flac]", report.Trees[0].BucketOnly)
	}
}

func TestRunTargetNotEmptyRefused(t *testing.T) {
	r, _ := newTestRunner(t, okSqliteScript, map[string]string{"library": "[]", "state": "[]"})

	to := t.TempDir()
	if err := os.WriteFile(filepath.Join(to, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	_, err := r.Run(context.Background(), Options{To: to})
	if !errors.Is(err, ErrTargetNotEmpty) {
		t.Fatalf("Run() error = %v, want ErrTargetNotEmpty", err)
	}
}

func TestRunTargetNotEmptyForced(t *testing.T) {
	bodies := map[string]string{"library": "[]", "state": "[]", "_archive": "[]"}
	r, _ := newTestRunner(t, okSqliteScript, bodies)

	to := t.TempDir()
	if err := os.WriteFile(filepath.Join(to, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write existing file: %v", err)
	}

	if _, err := r.Run(context.Background(), Options{To: to, Force: true}); err != nil {
		t.Fatalf("Run() unexpected error with --force: %v", err)
	}
}

func TestRunMissingDatabaseSkipsNotFails(t *testing.T) {
	// The state fixture tree is empty, so neither database arrives
	// locally: verifyDatabase must skip, not fail, exactly like backup's
	// snapshotDB treats a database that was never created.
	bodies := map[string]string{"library": "[]", "state": "[]"}
	r, _ := newTestRunner(t, okSqliteScript, bodies)

	to := filepath.Join(t.TempDir(), "restored")
	report, err := r.Run(context.Background(), Options{To: to})
	if err != nil {
		t.Fatalf("Run() unexpected error: %v", err)
	}
	for _, db := range report.Databases {
		if db.Present || db.OK {
			t.Errorf("database check %+v, want Present=false (never backed up)", db)
		}
	}
}

const corruptSqliteScript = `
case "$2" in
  "PRAGMA integrity_check;") echo "*** in database main ***
row 5 missing from index" ;;
esac
`

func TestRunIntegrityCheckFailureAborts(t *testing.T) {
	bodies := map[string]string{
		"library": `[]`,
		"state":   `[{"Path":"library.db","IsDir":false}]`,
	}
	r, fixtureDir := newTestRunner(t, corruptSqliteScript, bodies)
	if err := os.WriteFile(filepath.Join(fixtureDir, "state", "library.db"), []byte("beets-db"), 0o644); err != nil {
		t.Fatalf("write fixture library.db: %v", err)
	}

	to := filepath.Join(t.TempDir(), "restored")
	report, err := r.Run(context.Background(), Options{To: to})
	if !errors.Is(err, ErrIntegrityCheck) {
		t.Fatalf("Run() error = %v, want ErrIntegrityCheck", err)
	}
	if report == nil {
		t.Fatal("Run() report = nil, want the partial report even on a hard failure")
	}
	if len(report.Trees) != 2 {
		t.Errorf("report.Trees = %d entries, want both trees recorded before the database check ran", len(report.Trees))
	}
}

func TestRunRcloneCopyFailureAborts(t *testing.T) {
	binDir := t.TempDir()
	root := t.TempDir()
	failScript := `echo "rclone: forced failure" >&2; exit 1`
	r := &Runner{
		Cfg:        &config.Config{R2: config.R2Config{AccountID: "acct", Bucket: "bucket", Remote: "r2", AccessKeyID: "key", SecretAccessKey: "secret"}},
		SqlitePath: writeStubScript(t, binDir, "sqlite3", okSqliteScript),
		RclonePath: writeStubScript(t, binDir, "rclone", failScript),
		Now:        func() time.Time { return time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC) },
	}

	to := filepath.Join(root, "restored")
	report, err := r.Run(context.Background(), Options{To: to})
	if err == nil {
		t.Fatal("Run() error = nil, want the rclone copy failure propagated")
	}
	if len(report.Trees) != 0 {
		t.Errorf("report.Trees = %d entries, want 0 (aborted before the first tree completed)", len(report.Trees))
	}
}

func TestCheckTarget(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		force   bool
		wantErr bool
	}{
		{
			name:  "missing directory",
			setup: func(t *testing.T) string { return filepath.Join(t.TempDir(), "nope") },
		},
		{
			name:  "empty directory",
			setup: func(t *testing.T) string { return t.TempDir() },
		},
		{
			name: "non-empty directory refused",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "f"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantErr: true,
		},
		{
			name: "non-empty directory forced",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "f"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			force: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkTarget(tt.setup(t), tt.force)
			if tt.wantErr && !errors.Is(err, ErrTargetNotEmpty) {
				t.Errorf("checkTarget() error = %v, want ErrTargetNotEmpty", err)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("checkTarget() unexpected error: %v", err)
			}
		})
	}
}
