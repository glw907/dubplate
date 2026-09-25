package store

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"testing/fstest"
	"time"
)

// populatedStore returns a store migrated to version 1 of the
// synthetic set and holding populate's rows, so an empty copy reads
// differently from it.
func populatedStore(t *testing.T) (leaf, *Handle) {
	t.Helper()
	l, h := newMigrateStore(t)
	mustMigrate(t, h, l, synthSet(1), defaultOpts())
	populate(t, l.dbPath)
	return l, h
}

// readFile reads a snapshot's bytes. A snapshot is a separate file
// from the store, and no connection in the test process holds it.
func readFile(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return b
}

// TestMigrate_Snapshot is spec criterion 9: migrating a populated store
// at N leaves a verified 0600 copy at N holding the rows, the staging
// directory empty, and the copy still a rollback-journal database; a
// second migration leaves the lower copy byte-identical.
func TestMigrate_Snapshot(t *testing.T) {
	// Under umask 0 a copy wider than 0600 would show as wider.
	old := syscall.Umask(0)
	t.Cleanup(func() { syscall.Umask(old) })

	l, h := populatedStore(t)
	mustMigrate(t, h, l, synthSet(2), defaultOpts())

	v1 := filepath.Join(l.snapshotsDir, "pre-migrate-v1.db")
	info, err := os.Stat(v1)
	if err != nil {
		t.Fatalf("stat the copy: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("copy mode = %o, want 0600", mode)
	}
	if staged := listDir(t, l.stagingDir); len(staged) != 0 {
		t.Errorf("staging = %q, want empty", staged)
	}
	if got := queryInt(t, v1, "PRAGMA user_version"); got != 1 {
		t.Errorf("copy user_version = %d, want 1", got)
	}
	if got := queryInt(t, v1, "PRAGMA application_id"); got != 0x64756270 {
		t.Errorf("copy application_id = %#x, want 0x64756270", got)
	}
	if got := queryInt(t, v1, "SELECT count(*) FROM child"); got != 3 {
		t.Errorf("copy child rows = %d, want the store's 3", got)
	}
	if header := readFile(t, v1); header[18] != 1 || header[19] != 1 {
		t.Errorf("copy header bytes 18 and 19 = %v, want [1 1] (rollback journal)", header[18:20])
	}

	before := readFile(t, v1)
	mustMigrate(t, h, l, synthSet(3), defaultOpts())
	if !bytes.Equal(readFile(t, v1), before) {
		t.Error("pre-migrate-v1.db changed across the next migration, want it byte-identical")
	}
	if got := queryInt(t, filepath.Join(l.snapshotsDir, "pre-migrate-v2.db"), "PRAGMA user_version"); got != 2 {
		t.Errorf("pre-migrate-v2.db user_version = %d, want 2", got)
	}
}

func TestMigrate_SnapshotReplacesSameVersion(t *testing.T) {
	l, h := populatedStore(t)
	stale := filepath.Join(l.snapshotsDir, "pre-migrate-v1.db")
	if err := os.WriteFile(stale, []byte("an earlier run's copy"), 0o600); err != nil {
		t.Fatalf("plant: %v", err)
	}

	mustMigrate(t, h, l, synthSet(2), defaultOpts())
	if got := queryInt(t, stale, "SELECT count(*) FROM child"); got != 3 {
		t.Errorf("replaced copy child rows = %d, want 3", got)
	}
}

// TestMigrate_SnapshotRefusals: every failed copy refuses the
// migration with the store at N, names the path, and leaves no copy in
// snapshots.
func TestMigrate_SnapshotRefusals(t *testing.T) {
	tests := []struct {
		name string
		// setup returns the migration source and the error's names.
		setup func(t *testing.T, l leaf) (fstest.MapFS, []string)
		// wantStaged is staging's listing afterward.
		wantStaged []string
	}{
		{
			name: "staging is unwritable",
			setup: func(t *testing.T, l leaf) (fstest.MapFS, []string) {
				if os.Geteuid() == 0 {
					t.Skip("root writes through a 0500 directory")
				}
				if err := os.Chmod(l.stagingDir, 0o500); err != nil {
					t.Fatalf("chmod staging: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(l.stagingDir, leafDirPerm) })
				return synthSet(2), []string{filepath.Join(l.stagingDir, "pre-migrate-v1.db"), filepath.Join(l.snapshotsDir, "pre-migrate-v1.db")}
			},
		},
		{
			// leaf.create chmods every leaf directory back to 0700, so
			// snapshots/ is narrowed after the leaf exists: the copy
			// stages and verifies, and only its rename fails.
			name: "snapshots is unwritable, so the rename fails",
			setup: func(t *testing.T, l leaf) (fstest.MapFS, []string) {
				if os.Geteuid() == 0 {
					t.Skip("root renames into a 0500 directory")
				}
				if err := os.Chmod(l.snapshotsDir, 0o500); err != nil {
					t.Fatalf("chmod snapshots: %v", err)
				}
				t.Cleanup(func() { _ = os.Chmod(l.snapshotsDir, leafDirPerm) })
				return synthSet(2), []string{"rename " + filepath.Join(l.stagingDir, "pre-migrate-v1.db"), l.snapshotsDir}
			},
		},
		{
			name: "the staging target already exists",
			setup: func(t *testing.T, l leaf) (fstest.MapFS, []string) {
				p := filepath.Join(l.stagingDir, "pre-migrate-v1.db")
				if err := os.WriteFile(p, []byte("leftover"), 0o600); err != nil {
					t.Fatalf("plant: %v", err)
				}
				return synthSet(2), []string{p}
			},
			wantStaged: []string{"pre-migrate-v1.db"},
		},
		{
			name: "free space is one byte short",
			setup: func(t *testing.T, l leaf) (fstest.MapFS, []string) {
				need := uint64(queryInt(t, l.dbPath, "SELECT page_count * page_size FROM pragma_page_count, pragma_page_size")) + 67108864
				stubFree(t, need-1)
				return synthSet(2), []string{
					l.stagingDir,
					strconv.FormatUint(need-1, 10) + " bytes free",
					"need " + strconv.FormatUint(need, 10),
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := populatedStore(t)
			src, names := tt.setup(t, l)
			before := readState(t, l.dbPath)

			err := migrate(context.Background(), h, l, readMigrationsT(t, src), defaultOpts())
			assertErrNames(t, err, names...)
			assertUnchanged(t, l.dbPath, before)
			if snaps := listDir(t, l.snapshotsDir); len(snaps) != 0 {
				t.Errorf("snapshots = %q, want none", snaps)
			}
			if staged := listDir(t, l.stagingDir); !slices.Equal(staged, tt.wantStaged) {
				t.Errorf("staging = %q, want %q", staged, tt.wantStaged)
			}
		})
	}
}

// stubFree makes statfs report free bytes available, one byte per
// fragment, until the test ends.
func stubFree(t *testing.T, free uint64) {
	t.Helper()
	orig := statfsFunc
	t.Cleanup(func() { statfsFunc = orig })
	statfsFunc = func(_ string, st *syscall.Statfs_t) error {
		st.Bavail = free
		st.Frsize = 1
		return nil
	}
}

func TestMigrate_FreeSpaceExactlyEnough(t *testing.T) {
	l, h := populatedStore(t)
	need := uint64(queryInt(t, l.dbPath, "SELECT page_count * page_size FROM pragma_page_count, pragma_page_size")) + 67108864
	stubFree(t, need)

	mustMigrate(t, h, l, synthSet(2), defaultOpts())
	if got := readState(t, l.dbPath).version; got != 2 {
		t.Errorf("user_version = %d, want 2", got)
	}
}

// TestTakeSnapshot_VerifyRefuses: a copy that does not read the
// version asked for, or does not carry the store marker, is removed
// from staging and never reaches snapshots.
func TestTakeSnapshot_VerifyRefuses(t *testing.T) {
	tests := []struct {
		name    string
		first   string
		want    int
		wantErr string
	}{
		{name: "a version other than the one asked for", first: baseMigration, want: 2, wantErr: "copy at version 1, want 2"},
		{
			name:    "no store marker",
			first:   "CREATE TABLE migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL, sha256 TEXT NOT NULL) STRICT;\n",
			want:    1,
			wantErr: "application_id 0x0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := newMigrateStore(t)
			mustMigrate(t, h, l, fstest.MapFS{synthName(1): {Data: []byte(tt.first)}}, defaultOpts())

			err := takeSnapshot(context.Background(), l, time.Second, "pre-migrate-v1.db", tt.want)
			assertErrNames(t, err, tt.wantErr, filepath.Join(l.snapshotsDir, "pre-migrate-v1.db"))
			if staged := listDir(t, l.stagingDir); len(staged) != 0 {
				t.Errorf("staging = %q, want empty", staged)
			}
			if snaps := listDir(t, l.snapshotsDir); len(snaps) != 0 {
				t.Errorf("snapshots = %q, want none", snaps)
			}
		})
	}
}

// TestMigrate_Retention: with backup off, a migration keeps the keep
// highest-versioned pre-migration copies and deletes the rest, touching
// no other name; with backup on it deletes nothing. The copies' times
// run against their versions, so a prune by time deletes the wrong one.
func TestMigrate_Retention(t *testing.T) {
	others := []string{"pre-restore-x.db", "pre-migrate-v1.db.bak", "pre-migrate-v01.db", "notes.txt"}
	tests := []struct {
		name   string
		backup bool
		want   []string
	}{
		{
			name:   "backup on keeps every copy",
			backup: true,
			want:   []string{"notes.txt", "pre-migrate-v01.db", "pre-migrate-v1.db", "pre-migrate-v1.db.bak", "pre-migrate-v2.db", "pre-migrate-v3.db", "pre-restore-x.db"},
		},
		{
			name:   "backup off keeps the newest two by version",
			backup: false,
			want:   []string{"notes.txt", "pre-migrate-v01.db", "pre-migrate-v1.db.bak", "pre-migrate-v2.db", "pre-migrate-v3.db", "pre-restore-x.db"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := populatedStore(t)
			for _, name := range others {
				if err := os.WriteFile(filepath.Join(l.snapshotsDir, name), []byte("not a copy"), 0o600); err != nil {
					t.Fatalf("plant %s: %v", name, err)
				}
			}
			opts := defaultOpts()
			opts.backup = tt.backup
			opts.keep = 2

			mustMigrate(t, h, l, synthSet(2), opts)
			mustMigrate(t, h, l, synthSet(3), opts)

			// Version 1 newest by time, version 2 oldest, and every
			// planted name older still.
			setTime(t, filepath.Join(l.snapshotsDir, "pre-migrate-v1.db"), time.Date(2100, 1, 1, 0, 0, 0, 0, time.UTC))
			setTime(t, filepath.Join(l.snapshotsDir, "pre-migrate-v2.db"), time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC))
			for _, name := range others {
				setTime(t, filepath.Join(l.snapshotsDir, name), time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))
			}
			mustMigrate(t, h, l, synthSet(4), opts)

			if got := listDir(t, l.snapshotsDir); !slices.Equal(got, tt.want) {
				t.Errorf("snapshots = %q, want %q", got, tt.want)
			}
		})
	}
}

// setTime sets path's access and modification times to at.
func setTime(t *testing.T, path string, at time.Time) {
	t.Helper()
	if err := os.Chtimes(path, at, at); err != nil {
		t.Fatalf("chtimes %s: %v", path, err)
	}
}

// testInstance is a well-formed instance for the backup copy tests.
const testInstance = "0123456789abcdef0123456789abcdef"

// insertJob inserts one queued job of kind through h's write method.
func insertJob(ctx context.Context, h *Handle, kind string) error {
	return h.Update(ctx, func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx,
			"INSERT INTO jobs (id, kind, at, state, owner_instance, actor, updated_at) VALUES (?, ?, 1, 'queued', ?, 'local:test', 1)",
			NewID(), kind, testInstance)
		return err
	})
}

// populateJobs inserts n queued jobs of kind "backup" through h.
func populateJobs(t *testing.T, h *Handle, n int) {
	t.Helper()
	for range n {
		if err := insertJob(deadline(t, 5*time.Second), h, "backup"); err != nil {
			t.Fatalf("insert job: %v", err)
		}
	}
}

// stallBound is the longest a looping writer's single commit may take
// before it counts as stalled.
const stallBound = time.Second

// loopWriter commits one job at a time through h until stop closes. It
// counts every commit in committed, and sets stalled when any one
// commit outlasts stallBound, timing each against a [time.Timer]:
// internal/engine never reads the clock, its tests included.
type loopWriter struct {
	committed atomic.Int64
	stalled   atomic.Bool
	stop      chan struct{}
	done      chan error
}

func startLoopWriter(h *Handle) *loopWriter {
	w := &loopWriter{stop: make(chan struct{}), done: make(chan error, 1)}
	go func() {
		for {
			select {
			case <-w.stop:
				w.done <- nil
				return
			default:
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			watch := time.AfterFunc(stallBound, func() { w.stalled.Store(true) })
			err := insertJob(ctx, h, "backup")
			watch.Stop()
			cancel()
			if err != nil {
				w.done <- err
				return
			}
			w.committed.Add(1)
		}
	}()
	return w
}

// finish stops w and fails the test on any error it met.
func (w *loopWriter) finish(t *testing.T) {
	t.Helper()
	close(w.stop)
	if err := <-w.done; err != nil {
		t.Errorf("looping writer: %v", err)
	}
}

// awaitCommits blocks until w has committed at least n rows.
func (w *loopWriter) awaitCommits(t *testing.T, n int64) {
	t.Helper()
	limit := time.NewTimer(10 * time.Second)
	defer limit.Stop()
	for w.committed.Load() < n {
		select {
		case <-limit.C:
			t.Fatalf("looping writer committed %d rows in 10s, want %d", w.committed.Load(), n)
		case <-time.After(time.Millisecond):
		}
	}
}

// assertVerifiedCopy fails unless the copy at path passes the integrity
// check and reads the binary's version and the store marker.
func assertVerifiedCopy(t *testing.T, path string) {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, inspectDSN(path, migrateBusy))
	if err != nil {
		t.Fatalf("open copy: %v", err)
	}
	defer func() { _ = db.Close() }()
	if got := queryStrings(t, db, "PRAGMA integrity_check"); !slices.Equal(got, []string{"ok"}) {
		t.Errorf("copy integrity_check = %q, want [ok]", got)
	}
	if got := queryInt(t, path, "PRAGMA user_version"); got != binaryHead {
		t.Errorf("copy user_version = %d, want %d", got, binaryHead)
	}
	if got := queryInt(t, path, "PRAGMA application_id"); got != 0x64756270 {
		t.Errorf("copy application_id = %#x, want 0x64756270", got)
	}
}

// TestBackupCopy_LoopingWriter is spec criterion 9's store half,
// fixture A: a writer commits through the write method in a loop across
// the copy, and the copy holds at least every row committed before it
// started. The stage is 0600 under a 0022 umask.
func TestBackupCopy_LoopingWriter(t *testing.T) {
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	h := writeHandle(t, currentStore(t))
	const populated = 20
	populateJobs(t, h, populated)

	w := startLoopWriter(h)
	w.awaitCommits(t, 10)
	before := w.committed.Load()
	path, err := h.BackupCopy(deadline(t, 30*time.Second), slog.New(slog.DiscardHandler), testInstance)
	w.finish(t)
	if err != nil {
		t.Fatalf("BackupCopy: %v", err)
	}

	if want := filepath.Join(h.leaf.stagingDir, "backup-"+testInstance+".db"); path != want {
		t.Errorf("copy path = %q, want %q", path, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat the copy: %v", err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("copy mode = %o, want 0600", mode)
	}
	assertVerifiedCopy(t, path)
	if got := queryInt(t, path, "SELECT count(*) FROM jobs"); int64(got) < populated+before {
		t.Errorf("copy holds %d jobs, want at least the %d committed before it started", got, populated+before)
	}
}

// TestBackupCopy_UncommittedWrite is spec criterion 9's store half,
// fixture B: a second connection holds an uncommitted insert across the
// copy, which finishes without it. That connection holds SQLite's one
// write lock, so the copy's checkpoint meets it and reports busy.
func TestBackupCopy_UncommittedWrite(t *testing.T) {
	h := writeHandle(t, currentStore(t))
	populateJobs(t, h, 5)

	other, err := sql.Open(sqliteDriverName, maintenanceDSN(h.path, migrateBusy))
	if err != nil {
		t.Fatalf("open second connection: %v", err)
	}
	t.Cleanup(func() { _ = other.Close() })
	tx, err := other.Begin()
	if err != nil {
		t.Fatalf("begin on the second connection: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	if _, err := tx.Exec("INSERT INTO jobs (id, kind, at, state, owner_instance, actor, updated_at) VALUES (?, 'marked', 1, 'queued', ?, 'local:test', 1)", NewID(), testInstance); err != nil {
		t.Fatalf("uncommitted insert: %v", err)
	}

	var logs bytes.Buffer
	var path string
	err = assertWithin(t, 5*time.Second, func() error {
		var err error
		path, err = h.BackupCopy(deadline(t, 30*time.Second), slog.New(slog.NewTextHandler(&logs, nil)), testInstance)
		return err
	})
	if err != nil {
		t.Fatalf("BackupCopy: %v", err)
	}
	assertVerifiedCopy(t, path)
	if got := queryInt(t, path, "SELECT count(*) FROM jobs WHERE kind = 'marked'"); got != 0 {
		t.Errorf("copy holds %d uncommitted rows, want 0", got)
	}
	if got := queryInt(t, path, "SELECT count(*) FROM jobs WHERE kind = 'backup'"); got != 5 {
		t.Errorf("copy holds %d committed rows, want 5", got)
	}
	if !strings.Contains(logs.String(), "checkpoint blocked") {
		t.Errorf("logs %q hold no busy checkpoint record", logs.String())
	}
}

// TestBackupCopy_CheckpointDoesNotStallWriters: with a read
// transaction held open on a second connection, the copy's truncating
// checkpoint cannot finish. A checkpoint that waited for that reader
// through the handle's 15 s busy timeout would hold every writer for as
// long, so no commit of the looping writer outlasts 1 s, and the busy
// checkpoint is logged.
func TestBackupCopy_CheckpointDoesNotStallWriters(t *testing.T) {
	h, err := OpenWrite(deadline(t, 5*time.Second), currentStore(t), 15*time.Second)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	populateJobs(t, h, 5)

	reader, err := sql.Open(sqliteDriverName, readDSN(h.path, migrateBusy))
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	t.Cleanup(func() { _ = reader.Close() })
	rtx, err := reader.BeginTx(context.Background(), &sql.TxOptions{ReadOnly: true})
	if err != nil {
		t.Fatalf("begin read: %v", err)
	}
	t.Cleanup(func() { _ = rtx.Rollback() })
	var n int
	if err := rtx.QueryRow("SELECT count(*) FROM jobs").Scan(&n); err != nil {
		t.Fatalf("read inside the held transaction: %v", err)
	}

	var logs bytes.Buffer
	w := startLoopWriter(h)
	w.awaitCommits(t, 1)
	before := w.committed.Load()
	_, err = h.BackupCopy(deadline(t, 60*time.Second), slog.New(slog.NewTextHandler(&logs, nil)), testInstance)
	during := w.committed.Load() - before
	w.finish(t)
	if err != nil {
		t.Fatalf("BackupCopy: %v", err)
	}

	if w.stalled.Load() {
		t.Errorf("a commit outlasted %s while the copy ran, want every one under it", stallBound)
	}
	if during == 0 {
		t.Error("the looping writer committed nothing while the copy ran")
	}
	if !strings.Contains(logs.String(), "checkpoint blocked") {
		t.Errorf("logs %q hold no busy checkpoint record", logs.String())
	}
}

// TestBackupCopy_Refusals: each refusal names what stopped it, and
// leaves staging as it found it.
func TestBackupCopy_Refusals(t *testing.T) {
	stage := "backup-" + testInstance + ".db"
	tests := []struct {
		name     string
		instance string
		// setup returns the error's names.
		setup      func(t *testing.T, h *Handle) []string
		wantStaged []string
	}{
		{
			name:     "free space one byte short",
			instance: testInstance,
			setup: func(t *testing.T, h *Handle) []string {
				need := uint64(queryInt(t, h.path, "SELECT page_count * page_size FROM pragma_page_count, pragma_page_size")) + 67108864
				stubFree(t, need-1)
				return []string{
					filepath.Join(h.leaf.stagingDir, stage),
					strconv.FormatUint(need-1, 10) + " bytes free",
					"need " + strconv.FormatUint(need, 10),
				}
			},
		},
		{
			name:     "the stage already exists",
			instance: testInstance,
			setup: func(t *testing.T, h *Handle) []string {
				p := filepath.Join(h.leaf.stagingDir, stage)
				if err := os.WriteFile(p, []byte("leftover"), 0o600); err != nil {
					t.Fatalf("plant: %v", err)
				}
				return []string{p, "file exists"}
			},
			wantStaged: []string{stage},
		},
		{
			name:     "an instance that could name another path",
			instance: "../../db/dubplate",
			setup: func(*testing.T, *Handle) []string {
				return []string{"../../db/dubplate", "32 lowercase hex characters"}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := writeHandle(t, currentStore(t))
			names := tt.setup(t, h)

			path, err := h.BackupCopy(deadline(t, 30*time.Second), slog.New(slog.DiscardHandler), tt.instance)
			assertErrNames(t, err, names...)
			if path != "" {
				t.Errorf("path = %q on a refusal, want empty", path)
			}
			if got := listDir(t, h.leaf.stagingDir); !slices.Equal(got, tt.wantStaged) {
				t.Errorf("staging = %q, want %q", got, tt.wantStaged)
			}
			if tt.wantStaged != nil {
				if got := readFile(t, filepath.Join(h.leaf.stagingDir, stage)); string(got) != "leftover" {
					t.Errorf("planted stage now reads %q, want it untouched", got)
				}
			}
		})
	}
}

// TestRemoveBackupStages removes every backup stage and no other name.
func TestRemoveBackupStages(t *testing.T) {
	h := writeHandle(t, currentStore(t))
	for _, name := range []string{"backup-" + testInstance + ".db", "backup-fedcba9876543210fedcba9876543210.db", "pre-migrate-v1.db", "backup-notes.txt"} {
		if err := os.WriteFile(filepath.Join(h.leaf.stagingDir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("plant %s: %v", name, err)
		}
	}

	if err := h.RemoveBackupStages(); err != nil {
		t.Fatalf("RemoveBackupStages: %v", err)
	}
	if got, want := listDir(t, h.leaf.stagingDir), []string{"backup-notes.txt", "pre-migrate-v1.db"}; !slices.Equal(got, want) {
		t.Errorf("staging = %q, want %q", got, want)
	}
}

// TestWriteBackupRef: a second reference replaces the first, touching
// no copy, and a stamp outside the layout is refused before anything
// is written.
func TestWriteBackupRef(t *testing.T) {
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	h := writeHandle(t, currentStore(t))
	for _, name := range []string{"pre-migrate-v1.db", "notes.txt"} {
		if err := os.WriteFile(filepath.Join(h.leaf.snapshotsDir, name), []byte("x"), 0o600); err != nil {
			t.Fatalf("plant %s: %v", name, err)
		}
	}

	if err := h.WriteBackupRef("20260924T120000.000000Z", "first:a"); err != nil {
		t.Fatalf("first reference: %v", err)
	}
	if err := h.WriteBackupRef("20260925T120000.000001Z", "second:b"); err != nil {
		t.Fatalf("second reference: %v", err)
	}
	want := []string{"backup-20260925T120000.000001Z.ref", "notes.txt", "pre-migrate-v1.db"}
	if got := listDir(t, h.leaf.snapshotsDir); !slices.Equal(got, want) {
		t.Errorf("snapshots = %q, want %q", got, want)
	}
	ref := filepath.Join(h.leaf.snapshotsDir, "backup-20260925T120000.000001Z.ref")
	if got := string(readFile(t, ref)); got != "stamp 20260925T120000.000001Z\nremote second:b\n" {
		t.Errorf("reference reads %q", got)
	}
	info, err := os.Stat(ref)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("reference mode = %o, want 0600", mode)
	}

	err = h.WriteBackupRef("20260926T120000Z", "third:c")
	assertErrNames(t, err, "20260926T120000Z")
	if got := listDir(t, h.leaf.snapshotsDir); !slices.Equal(got, want) {
		t.Errorf("snapshots after a refused stamp = %q, want %q", got, want)
	}
}

// TestWriteBackupRef_RemoteRefusals: a remote that would corrupt the
// reference's line-oriented body is refused before anything is
// written, and a remote free of those bytes is still accepted.
func TestWriteBackupRef_RemoteRefusals(t *testing.T) {
	h := writeHandle(t, currentStore(t))
	for _, remote := range []string{
		"crypt:a\nremote bogus",
		"crypt:a\rb",
		"crypt:a\x00b",
	} {
		if err := h.WriteBackupRef("20260924T120000.000000Z", remote); err == nil {
			t.Errorf("write reference with remote %q: got no error", remote)
		}
		if got := listDir(t, h.leaf.snapshotsDir); len(got) != 0 {
			t.Errorf("snapshots after a refused remote %q = %q, want empty", remote, got)
		}
	}

	if err := h.WriteBackupRef("20260924T120000.000000Z", "crypt:remote-a"); err != nil {
		t.Fatalf("write reference with a plain remote: %v", err)
	}
}
