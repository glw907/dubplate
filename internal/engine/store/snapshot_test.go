package store

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"syscall"
	"testing"
	"testing/fstest"
	"time"
)

// populatedStore returns a store migrated to version 1 of the
// synthetic set and holding populate's rows, so an empty copy reads
// differently from it.
func populatedStore(t *testing.T) (leaf, *handle) {
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
