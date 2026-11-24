package main

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/store"
)

// The two archives under testdata are store snapshots the backup verb
// wrote: archive-d1e93e9.db at the box's commit (schema 5, with
// migration 0003's text from before users.streamer_username became
// username), and archive-2954031.db at schema 6. Each store holds the
// kebab-case spellings migration 0007 rewrites. The pull request that
// added them gives the commands that built them.

// archiveRows are the queries every archive answers, with the spellings
// a restore and a serve start leave. The jobs query names the seeded
// rows by id, since the backup that wrote the archive left its own.
var archiveRows = []struct {
	query string
	want  []string
}{
	{"SELECT key || ' ' || source FROM settings ORDER BY key", []string{
		"DUBPLATE_BACKUP_SNAPSHOT_KEEP admin",
		"DUBPLATE_JOB_RETENTION env_seed",
	}},
	{"SELECT channel FROM settings_audit WHERE id IN (x'e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1e1', x'e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2e2') ORDER BY id", []string{
		"env_seed",
		"command_line",
	}},
	{"SELECT failure_code FROM jobs WHERE id IN (x'f1f1f1f1f1f1f1f1f1f1f1f1f1f1f1f1', x'f2f2f2f2f2f2f2f2f2f2f2f2f2f2f2f2', x'f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3f3') ORDER BY id", []string{
		"rclone_exit_3",
		"library_short",
		"job_error",
	}},
	{"SELECT u.username FROM collections c JOIN users u ON u.id = c.user_id", []string{"owner"}},
}

// intakeRows are the queries only the schema-6 archive answers, its
// intake tables being absent at schema 5. The expiry's reason is free
// text and keeps its spelling.
var intakeRows = []struct {
	query string
	want  []string
}{
	{"SELECT state || ' ' || source FROM submissions ORDER BY id", []string{
		"check_unfinished free_download",
		"rejected own_recording",
	}},
	{"SELECT coalesce(verdict, '-') FROM submission_files ORDER BY id", []string{"-", "below_cd_quality"}},
	{"SELECT kind || ' ' || coalesce(reason, '-') FROM submission_events ORDER BY id", []string{
		"uploading -",
		"check_unfinished could_not_check",
		"rejected already_in_library",
		"expired staged file went-missing",
	}},
	{"SELECT code FROM submission_warnings ORDER BY id", []string{"no_cover_art", "extras_dropped"}},
	{"SELECT u.username FROM submissions s JOIN users u ON u.id = s.user_id ORDER BY s.id", []string{"owner", "owner"}},
}

// TestRestoreCmd_BoxArchive restores the archive the box's commit
// wrote, at schema 5, and starts serve on it: the store is current, and
// every kebab-case spelling it held reads back in snake_case.
func TestRestoreCmd_BoxArchive(t *testing.T) {
	db := restoreArchive(t, "archive-d1e93e9.db")
	assertArchiveRows(t, db, archiveRows)
}

// TestRestoreCmd_ArchiveAtSix restores an archive written at schema 6
// and starts serve on it: the store is current, and every kebab-case
// spelling it held, the intake tables' included, reads back in
// snake_case.
func TestRestoreCmd_ArchiveAtSix(t *testing.T) {
	db := restoreArchive(t, "archive-2954031.db")
	assertArchiveRows(t, db, archiveRows)
	assertArchiveRows(t, db, intakeRows)
}

// restoreArchive restores a copy of testdata/name with the restore
// verb, starts serve's preparation on the result as a serve start does,
// and returns a connection to the restored store, which it requires to
// be current. The environment sets DUBPLATE_JOB_RETENTION to the value
// the archive's env-seed row holds, so the start leaves that row alone.
func restoreArchive(t *testing.T, name string) *sql.DB {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(src, raw, 0o600); err != nil {
		t.Fatal(err)
	}

	environ := append(replaceKey(serveAuthEnviron(t), "DUBPLATE_LIBRARY_ROOT", t.TempDir()), "DUBPLATE_JOB_RETENTION=2160h")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	var stdout bytes.Buffer
	code, err := restoreRunFor(context.Background(), environ, restoreFlags{fromFile: src}, &stdout, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("restore exit = %d, want 0; error: %v\nstdout:\n%s", code, err, stdout.String())
	}
	if shell, err := buildServeShellWithLogger(t, environ, discardTestLogger()); err != nil || shell == nil {
		t.Fatalf("prepare = %v, %v, want a shell", shell, err)
	}
	if r := storeReport(t, stateDir); r.State != store.StateCurrent {
		t.Fatalf("store after prepare = %v (%s), want current", r.State, r.Reason)
	}

	db, err := sql.Open("sqlite", "file:"+filepath.Join(stateDir, "store", "db", "dubplate.db")+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func assertArchiveRows(t *testing.T, db *sql.DB, rows []struct {
	query string
	want  []string
}) {
	t.Helper()
	for _, r := range rows {
		got, err := queryStrings(db, r.query)
		if err != nil {
			t.Errorf("%s: %v", r.query, err)
			continue
		}
		if !slices.Equal(got, r.want) {
			t.Errorf("%s = %q, want %q", r.query, got, r.want)
		}
	}
}

func queryStrings(db *sql.DB, query string) ([]string, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}
