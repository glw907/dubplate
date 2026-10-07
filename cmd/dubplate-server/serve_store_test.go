package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// storeReport inspects the store under stateDir.
func storeReport(t *testing.T, stateDir string) store.Report {
	t.Helper()
	return store.Inspect(context.Background(), store.InspectOptions{StateDir: stateDir, BusyTimeout: testBackupBusy})
}

// preMigrateCopies lists the pre-migration copies under stateDir's
// snapshots directory, sorted.
func preMigrateCopies(t *testing.T, stateDir string) []string {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(stateDir, "store", "snapshots"))
	if err != nil {
		t.Fatalf("read snapshots: %v", err)
	}
	var names []string
	for _, e := range entries {
		if store.IsPreMigrateCopy(e.Name()) {
			names = append(names, e.Name())
		}
	}
	slices.Sort(names)
	return names
}

// plantCopies writes a pre-migration copy for each of versions.
func plantCopies(t *testing.T, stateDir string, versions ...int) {
	t.Helper()
	for _, v := range versions {
		p := filepath.Join(stateDir, "store", "snapshots", fmt.Sprintf("pre-migrate-v%d.db", v))
		if err := os.WriteFile(p, []byte("an older copy"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

// plantJobRow inserts one succeeded job into the closed store file at
// path, the row a silent re-creation of the store would lose.
func plantJobRow(t *testing.T, path string) {
	t.Helper()
	execStoreSQL(t, path, `INSERT INTO jobs (id, kind, at, state, owner_instance, actor, updated_at)
VALUES (X'0123456789ABCDEF0123456789ABCDEF', 'backup', 1, 'succeeded', 'inst', 'local:test', 1)`)
}

// plantKeepRow inserts a stored value of the local snapshot keep
// setting, with the given source, into the closed store file at path.
func plantKeepRow(t *testing.T, path, value, source string) {
	t.Helper()
	execStoreSQL(t, path, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
VALUES ('DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP', ?, ?, ?, 'serve', 1)`, value, source, value)
}

// jobCount reads the store's job count through a write-mode handle,
// which a shared holder's own lock does not exclude.
func jobCount(t *testing.T, stateDir string) int {
	t.Helper()
	h, err := store.OpenWrite(context.Background(), stateDir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()
	var n int
	if err := h.QueryRow(context.Background(), "SELECT count(*) FROM jobs").Scan(&n); err != nil {
		t.Fatalf("count jobs: %v", err)
	}
	return n
}

// fileSum returns the SHA-256 of the file at path.
func fileSum(t *testing.T, path string) [sha256.Size]byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return sha256.Sum256(b)
}

// TestPrepareServe_StoreAtStart is the start's states, each through the
// preparation function the verb calls: the store it returns from is
// at the binary's version, the first moment a listener could open.
// Fixture: each older store holds a populated job row, so a start that
// silently re-created the store would lose it.
func TestPrepareServe_StoreAtStart(t *testing.T) {
	head := storeReport(t, t.TempDir()).BinaryVersion

	tests := []struct {
		name       string
		setup      func(t *testing.T, stateDir string)
		wantJobs   int
		wantCopies []string
	}{
		{
			name:  "absent",
			setup: func(*testing.T, string) {},
		},
		{
			name: "leaf present and empty",
			setup: func(t *testing.T, stateDir string) {
				if err := os.Mkdir(filepath.Join(stateDir, "store"), 0o700); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "older, an env-seed keep row of 1, two copies, backup off",
			setup: func(t *testing.T, stateDir string) {
				path := olderStore(t, stateDir, 4)
				plantJobRow(t, path)
				plantKeepRow(t, path, "1", "env-seed")
				plantCopies(t, stateDir, 2, 3)
			},
			wantJobs:   1,
			wantCopies: []string{"pre-migrate-v4.db"},
		},
		{
			name: "older, no keep row, copies v1 to v3, backup off",
			setup: func(t *testing.T, stateDir string) {
				path := olderStore(t, stateDir, 4)
				plantJobRow(t, path)
				plantCopies(t, stateDir, 1, 2, 3)
			},
			wantJobs:   1,
			wantCopies: []string{"pre-migrate-v2.db", "pre-migrate-v3.db", "pre-migrate-v4.db"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := serveAuthEnviron(t)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			tt.setup(t, stateDir)

			var logs bytes.Buffer
			shell, err := buildServeShellWithLogger(t, environ, slog.New(slog.NewTextHandler(&logs, nil)))
			if err != nil {
				t.Fatalf("prepare: %v", err)
			}
			if shell == nil {
				t.Fatal("prepare returned no shell")
			}

			r := storeReport(t, stateDir)
			if r.State != store.StateCurrent || r.StoreVersion != head {
				t.Errorf("store after prepare = %v at v%d (%s), want current at v%d", r.State, r.StoreVersion, r.Reason, head)
			}
			if tt.wantJobs > 0 {
				if got := jobCount(t, stateDir); got != tt.wantJobs {
					t.Errorf("jobs after the migration = %d, want %d", got, tt.wantJobs)
				}
			}
			if tt.wantCopies != nil {
				if got := preMigrateCopies(t, stateDir); !slices.Equal(got, tt.wantCopies) {
					t.Errorf("pre-migration copies = %q, want %q", got, tt.wantCopies)
				}
			}
			if strings.Contains(logs.String(), "settling again") {
				t.Errorf("a plain start logged the settle warning: %s", logs.String())
			}
		})
	}
}

// TestPrepareServe_StoreRefusals is the start's refusing states: each
// returns an error and no shell, leaves the store's database file as it
// was, and writes no pre-migration copy.
func TestPrepareServe_StoreRefusals(t *testing.T) {
	head := storeReport(t, t.TempDir()).BinaryVersion

	tests := []struct {
		name        string
		setup       func(t *testing.T, stateDir string)
		extra       []string
		wantErr     []string
		wantState   store.State
		wantVersion int
	}{
		{
			name: "older, an admin keep row outside its bounds",
			setup: func(t *testing.T, stateDir string) {
				path := olderStore(t, stateDir, 4)
				plantKeepRow(t, path, "500", "admin")
			},
			wantErr:     []string{"DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP", "above the ceiling 100"},
			wantState:   store.StateOlder,
			wantVersion: 4,
		},
		{
			name: "older, the environment's keep value does not parse",
			setup: func(t *testing.T, stateDir string) {
				olderStore(t, stateDir, 4)
			},
			extra:       []string{"DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP=many"},
			wantErr:     []string{"DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP", "invalid value"},
			wantState:   store.StateOlder,
			wantVersion: 4,
		},
		{
			name: "older than the settings table, an auth value out of bounds",
			setup: func(t *testing.T, stateDir string) {
				olderStore(t, stateDir, 3)
			},
			extra:       []string{"DUBPLATE_MAGIC_LINK_EXPIRY=2h"},
			wantErr:     []string{"DUBPLATE_MAGIC_LINK_EXPIRY", "above the ceiling"},
			wantState:   store.StateOlder,
			wantVersion: 3,
		},
		{
			name: "older than the jobs table, an auth value out of bounds",
			setup: func(t *testing.T, stateDir string) {
				olderStore(t, stateDir, 1)
			},
			extra:       []string{"DUBPLATE_MAGIC_LINK_EXPIRY=2h"},
			wantErr:     []string{"DUBPLATE_MAGIC_LINK_EXPIRY", "above the ceiling"},
			wantState:   store.StateOlder,
			wantVersion: 1,
		},
		{
			name: "newer than the binary",
			setup: func(t *testing.T, stateDir string) {
				path := olderStore(t, stateDir, head)
				execStoreSQL(t, path, fmt.Sprintf("PRAGMA user_version = %d", head+1))
			},
			wantErr:     []string{fmt.Sprintf("version %d", head+1), fmt.Sprintf("version %d", head)},
			wantState:   store.StateNewer,
			wantVersion: head + 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := append(serveAuthEnviron(t), tt.extra...)
			stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
			tt.setup(t, stateDir)
			db := storeDBPath(stateDir)
			before := fileSum(t, db)

			shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
			if err == nil {
				t.Fatal("prepare error = nil, want a refusal")
			}
			if shell != nil {
				t.Error("prepare returned a shell beside its refusal")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error = %q, want it to contain %q", err, want)
				}
			}
			if fileSum(t, db) != before {
				t.Error("a refused start changed the database file")
			}
			if got := preMigrateCopies(t, stateDir); len(got) != 0 {
				t.Errorf("a refused start wrote copies %q, want none", got)
			}
			r := storeReport(t, stateDir)
			if r.State != tt.wantState || r.StoreVersion != tt.wantVersion {
				t.Errorf("store after refusal = %v at v%d, want %v at v%d", r.State, r.StoreVersion, tt.wantState, tt.wantVersion)
			}
		})
	}
}

// TestPrepareServe_MigrationWaitReachesTheOpen holds the store's lock
// shared on a second descriptor in this process, which a migrating
// start's exclusive lock must wait out. The wall's one-second wait must
// bound that wait: the start fails naming the holder and the wait it
// gave up on, well inside the context's own deadline, which an
// unbounded or default five-minute wait would run into. Fixture: an
// older store, since only a migration needs the exclusive lock.
func TestPrepareServe_MigrationWaitReachesTheOpen(t *testing.T) {
	environ := append(serveAuthEnviron(t), "DUBPLATE_STORE_MIGRATION_WAIT=1s")
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	olderStore(t, stateDir, 4)
	db := storeDBPath(stateDir)
	before := fileSum(t, db)

	f, err := os.OpenFile(filepath.Join(stateDir, "store", "store.lock"), os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = f.Close() })
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_SH); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	start := time.Now()
	shell, err := buildServeShellContext(t, ctx, environ, discardTestLogger())
	if err == nil {
		t.Fatal("prepare error = nil, want it to give up on the held lock")
	}
	if shell != nil {
		t.Error("prepare returned a shell beside its refusal")
	}
	if !strings.Contains(err.Error(), "wait 1s for the store lock exclusive") {
		t.Errorf("error = %q, want it to name the 1s wait", err)
	}
	if !strings.Contains(err.Error(), "held by pid") && !strings.Contains(err.Error(), "held by another process") {
		t.Errorf("error = %q, want it to name the holder", err)
	}
	if elapsed := time.Since(start); elapsed > 15*time.Second {
		t.Errorf("prepare took %v, want it bounded by the 1s wait", elapsed)
	}
	if fileSum(t, db) != before {
		t.Error("a refused start changed the database file")
	}
	if got := preMigrateCopies(t, stateDir); len(got) != 0 {
		t.Errorf("a refused start wrote copies %q, want none", got)
	}
}

// TestPrepareServe_ReconcilePending finishes a restore the way serve
// does: a real restore from a file one version below the binary leaves a
// reconcile-pending record, which this test then gives a library half
// naming a root other than the configured one. The start migrates the
// store, writes the report into the leaf with restore's head over the
// record's root, and removes the record. Fixture: the restored store
// holds one track row, main/a.flac, which the configured root's tree
// holds and the record's root does not, so the report lists it as
// without a file only when the walk read the record's root.
func TestPrepareServe_ReconcilePending(t *testing.T) {
	configured := t.TempDir()
	recorded := t.TempDir()
	for _, p := range []string{filepath.Join(configured, "main", "a.flac"), filepath.Join(recorded, "main", "other.flac")} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("audio"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	srcDir := t.TempDir()
	src := olderStore(t, srcDir, 4)
	execStoreSQL(t, src, `
INSERT INTO users (id, address, display_name, username, created_at, updated_at)
VALUES (X'01010101010101010101010101010101', 'owner@example.org', 'Owner', 'owner', 1, 1);
INSERT INTO collections (id, name, root_path, owner, visibility, writer)
VALUES (X'02020202020202020202020202020202', 'main', CAST('main' AS BLOB), X'01010101010101010101010101010101', 'group', 'server');
INSERT INTO albums (id, collection_id, created_at, updated_at)
VALUES (X'03030303030303030303030303030303', X'02020202020202020202020202020202', 1, 1);
INSERT INTO tracks (id, album_id, collection_id, rel_path, created_at, updated_at)
VALUES (X'04040404040404040404040404040404', X'03030303030303030303030303030303', X'02020202020202020202020202020202', CAST('a.flac' AS BLOB), 1, 1)`)

	environ := replaceKey(serveAuthEnviron(t), "DUBPLATE_LIBRARY_ROOT", configured)
	stateDir := environValue(environ, "DUBPLATE_STATE_DIR")
	code, err := restoreRunFor(context.Background(), environ, restoreFlags{fromFile: src}, &bytes.Buffer{}, &bytes.Buffer{})
	if code != 0 {
		t.Fatalf("restore exit = %d, want 0; error: %v", code, err)
	}
	addLibraryHalf(t, filepath.Join(stateDir, "store", "restore-intent"), recorded)

	shell, err := buildServeShellWithLogger(t, environ, discardTestLogger())
	if err != nil || shell == nil {
		t.Fatalf("prepare = %v, %v, want a shell", shell, err)
	}
	r := storeReport(t, stateDir)
	if r.State != store.StateCurrent {
		t.Fatalf("store after prepare = %v (%s), want current", r.State, r.Reason)
	}
	if _, err := os.Stat(filepath.Join(stateDir, "store", "restore-intent")); !os.IsNotExist(err) {
		t.Errorf("restore-intent record after prepare: %v, want it removed", err)
	}

	reports, err := filepath.Glob(filepath.Join(stateDir, "store", "restore-report-*.txt"))
	if err != nil || len(reports) != 1 {
		t.Fatalf("reports = %q, %v, want exactly one", reports, err)
	}
	stamp := strings.TrimSuffix(strings.TrimPrefix(filepath.Base(reports[0]), "restore-report-"), ".txt")
	body, err := os.ReadFile(reports[0])
	if err != nil {
		t.Fatal(err)
	}
	text := string(body)
	if want := "restore " + stamp + "\nsource: file " + src + "\n"; !strings.HasPrefix(text, want) {
		t.Errorf("report head = %q, want it to start with %q", text, want)
	}
	resolved, err := filepath.EvalSymlinks(recorded)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"library root: " + resolved + "\n", "tracks without a file:\n  main/a.flac\n"} {
		if !strings.Contains(text, want) {
			t.Errorf("report %q, want it to hold %q", text, want)
		}
	}
	if strings.Contains(text, "the users table is empty: run member import") {
		t.Errorf("report %q names a member import verb that does not exist", text)
	}
}

// addLibraryHalf rewrites the restore-intent record at path to owe the
// library half under root, leaving every other field as restore wrote
// it.
func addLibraryHalf(t *testing.T, path, root string) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var rec map[string]any
	if err := json.Unmarshal(raw, &rec); err != nil {
		t.Fatalf("parse the record: %v", err)
	}
	rec["library"] = map[string]any{"root": root, "into_nonempty": false}
	out, err := json.Marshal(rec)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		t.Fatal(err)
	}
}
