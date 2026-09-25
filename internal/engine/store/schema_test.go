package store

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// sqliteUniqueCode and sqliteDatatypeCode are SQLite's extended result
// codes this file pins beyond write_test.go's sqliteFKCode
// (sqlite.org/rescode.html): SQLITE_CONSTRAINT_UNIQUE, and
// SQLITE_MISMATCH's type-affinity form, which a STRICT table answers
// when a bound value's storage class does not match a column's
// declared type.
const (
	sqliteUniqueCode   = 2067
	sqliteDatatypeCode = 3091
)

// fullyMigratedStore returns a store migrated to head and populated
// with every migration's fixture, so a test has one user
// (a1a1...a1), one collection (b1b1...b1) it owns, one album
// (c1c1...c1), and one track (d1d1...d1) to build on.
func fullyMigratedStore(t *testing.T) (leaf, *Handle) {
	t.Helper()
	l, h := newMigrateStore(t)
	ctx := context.Background()
	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSource()), defaultOpts()); err != nil {
		t.Fatalf("migrate to head: %v", err)
	}
	for _, name := range []string{"0001_ledger.sql", "0002_jobs.sql", "0003_catalog.sql"} {
		body, err := os.ReadFile(filepath.Join("testdata", "fixtures", name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", name, err)
		}
		if _, err := h.write.ExecContext(ctx, string(body)); err != nil {
			t.Fatalf("apply fixture %s: %v", name, err)
		}
	}
	return l, h
}

// TestSchemaApplicationIDMarksEveryFixture asserts the store marker
// reads back from application_id over every migration's populated
// fixture, not only an empty database.
func TestSchemaApplicationIDMarksEveryFixture(t *testing.T) {
	_, h := fullyMigratedStore(t)
	var marker int
	if err := h.write.QueryRowContext(context.Background(), "PRAGMA application_id").Scan(&marker); err != nil {
		t.Fatalf("read application_id: %v", err)
	}
	if marker != storeMarker {
		t.Errorf("application_id = %#x, want the store marker %#x", marker, storeMarker)
	}
}

// TestSchemaJobFixtureSnapshotSurvivesUpgrade is spec criterion 9's
// embedded half: migrating the populated job-records fixture (version
// 2) forward through the skeleton leaves pre-migrate-v2.db at
// user_version 2, holding the fixture's three job rows.
func TestSchemaJobFixtureSnapshotSurvivesUpgrade(t *testing.T) {
	l, h := newMigrateStore(t)
	ctx := context.Background()

	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSourceThrough(t, 2)), defaultOpts()); err != nil {
		t.Fatalf("migrate to version 2: %v", err)
	}
	fixture, err := os.ReadFile(filepath.Join("testdata", "fixtures", "0002_jobs.sql"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if _, err := h.write.ExecContext(ctx, string(fixture)); err != nil {
		t.Fatalf("apply fixture: %v", err)
	}

	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSource()), defaultOpts()); err != nil {
		t.Fatalf("migrate to head: %v", err)
	}

	snapPath := filepath.Join(l.snapshotsDir, preMigrateName(2))
	db, err := sql.Open(sqliteDriverName, inspectDSN(snapPath, migrateBusy))
	if err != nil {
		t.Fatalf("open snapshot %s: %v", snapPath, err)
	}
	defer func() { _ = db.Close() }()

	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != 2 {
		t.Errorf("snapshot user_version = %d, want 2", version)
	}
	var jobs int
	if err := db.QueryRowContext(ctx, "SELECT count(*) FROM jobs").Scan(&jobs); err != nil {
		t.Fatalf("count jobs: %v", err)
	}
	if jobs != 3 {
		t.Errorf("snapshot holds %d job rows, want the fixture's 3", jobs)
	}
}

// TestSchemaPathChecks asserts collections.root_path and
// tracks.rel_path each refuse an absolute path, a path with an empty,
// ".", or ".." component, or a path carrying NUL, and accept a
// well-formed relative path.
func TestSchemaPathChecks(t *testing.T) {
	tests := []struct {
		name    string
		path    []byte
		wantErr bool
	}{
		{"a well-formed relative path", []byte("Artist/Album/01 Track.flac"), false},
		{"a component that merely contains dots", []byte("A/..B/.c/d..flac"), false},
		{"a non-UTF-8 byte", []byte("a\xffb"), false},
		{"absolute", []byte("/abs/path"), true},
		{"empty", []byte(""), true},
		{"NUL", []byte("a\x00b"), true},
		{"a single-dot component", []byte("a/./b"), true},
		{"a double-dot component", []byte("a/../b"), true},
		{"leading double-dot", []byte("../b"), true},
		{"trailing dot component", []byte("a/."), true},
		{"an empty component mid-path", []byte("a//b"), true},
		{"a trailing slash", []byte("a/b/"), true},
		{"exactly one dot", []byte("."), true},
	}

	for _, tt := range tests {
		t.Run("collections.root_path: "+tt.name, func(t *testing.T) {
			_, h := fullyMigratedStore(t)
			ctx := context.Background()
			_, err := h.write.ExecContext(ctx,
				`INSERT INTO collections (id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation, last_mirror_at, duplicate_collapse, layout_template, tagging_policy, admission_format_allowlist, admission_match_threshold)
				 VALUES (?, 'Second', ?, x'a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1', 'group', 'server', 0, 0, NULL, 0, '', '', NULL, NULL)`,
				NewID(), tt.path,
			)
			if tt.wantErr && err == nil {
				t.Errorf("root_path %q: got nil error, want a CHECK refusal", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("root_path %q: %v", tt.path, err)
			}
		})

		t.Run("tracks.rel_path: "+tt.name, func(t *testing.T) {
			_, h := fullyMigratedStore(t)
			ctx := context.Background()
			_, err := h.write.ExecContext(ctx,
				`INSERT INTO tracks (id, album_id, collection_id, rel_path, mb_recording_id, created_at, updated_at)
				 VALUES (?, x'c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1', x'b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1', ?, NULL, 1, 1)`,
				NewID(), tt.path,
			)
			if tt.wantErr && err == nil {
				t.Errorf("rel_path %q: got nil error, want a CHECK refusal", tt.path)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("rel_path %q: %v", tt.path, err)
			}
		})
	}
}

// TestSchemaBlobColumnRefusesText is spec criterion 11's checklist
// line: a Go string bound into a BLOB column is refused. The
// mechanics review measured "cannot store TEXT value in BLOB column
// c.path (3091)" on a STRICT table.
func TestSchemaBlobColumnRefusesText(t *testing.T) {
	_, h := fullyMigratedStore(t)
	_, err := h.write.ExecContext(context.Background(),
		`INSERT INTO tracks (id, album_id, collection_id, rel_path, mb_recording_id, created_at, updated_at)
		 VALUES (x'd2d2d2d2d2d2d2d2d2d2d2d2d2d2d2d2', x'c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1', x'b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1b1', ?, NULL, 1, 1)`,
		"a Go string, not a []byte",
	)
	assertErrNames(t, err, "cannot store TEXT value in BLOB column")
	if code := sqliteCode(err); code != sqliteDatatypeCode {
		t.Errorf("code = %d, want %d", code, sqliteDatatypeCode)
	}
}

// TestSchemaDuplicateAddressRefused is spec criterion 12: a duplicate
// address is refused whether the existing row is live or tombstoned,
// under extended code 2067, and a case-folded duplicate
// streamer_username is refused the same way.
func TestSchemaDuplicateAddressRefused(t *testing.T) {
	ctx := context.Background()

	t.Run("against a live row", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO users (id, address, display_name, streamer_username, admin, tombstoned_at, created_at, updated_at)
			 VALUES (?, 'member1@example.invalid', 'Someone Else', 'someoneelse', 0, NULL, 2, 2)`,
			NewID(),
		)
		assertErrNames(t, err, "UNIQUE constraint failed: users.address")
		if code := sqliteCode(err); code != sqliteUniqueCode {
			t.Errorf("code = %d, want %d", code, sqliteUniqueCode)
		}
	})

	t.Run("against a tombstoned row", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		if _, err := h.write.ExecContext(ctx,
			`INSERT INTO users (id, address, display_name, streamer_username, admin, tombstoned_at, created_at, updated_at)
			 VALUES (?, 'gone@example.invalid', 'Gone', 'gone', 0, 999, 2, 2)`,
			NewID(),
		); err != nil {
			t.Fatalf("insert the tombstoned baseline: %v", err)
		}
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO users (id, address, display_name, streamer_username, admin, tombstoned_at, created_at, updated_at)
			 VALUES (?, 'gone@example.invalid', 'Someone Else', 'someoneelse2', 0, NULL, 3, 3)`,
			NewID(),
		)
		assertErrNames(t, err, "UNIQUE constraint failed: users.address")
		if code := sqliteCode(err); code != sqliteUniqueCode {
			t.Errorf("code = %d, want %d", code, sqliteUniqueCode)
		}
	})

	t.Run("a case-folded streamer_username", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO users (id, address, display_name, streamer_username, admin, tombstoned_at, created_at, updated_at)
			 VALUES (?, 'other@example.invalid', 'Other', 'Member1', 0, NULL, 2, 2)`,
			NewID(),
		)
		assertErrNames(t, err, "UNIQUE constraint failed: users.streamer_username")
		if code := sqliteCode(err); code != sqliteUniqueCode {
			t.Errorf("code = %d, want %d", code, sqliteUniqueCode)
		}
	})
}

// TestSchemaForeignKeyRefusals is the acceptance criterion for the
// tracks-to-albums composite key: a track whose collection_id differs
// from its album's is refused with 787, and a second collection owned
// by a missing user is refused with 787. Each case's fixture state
// holds only the row that failure names: the album, both collections,
// and the owning user of the first collection all exist, so the row
// under test is the only failure either insert can meet.
func TestSchemaForeignKeyRefusals(t *testing.T) {
	ctx := context.Background()

	t.Run("a track's collection differs from its album's", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		if _, err := h.write.ExecContext(ctx,
			`INSERT INTO collections (id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation, last_mirror_at, duplicate_collapse, layout_template, tagging_policy, admission_format_allowlist, admission_match_threshold)
			 VALUES (x'b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2', 'Second', ?, x'a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1', 'group', 'server', 0, 0, NULL, 0, '', '', NULL, NULL)`,
			[]byte("second"),
		); err != nil {
			t.Fatalf("insert the second collection: %v", err)
		}
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO tracks (id, album_id, collection_id, rel_path, mb_recording_id, created_at, updated_at)
			 VALUES (x'd3d3d3d3d3d3d3d3d3d3d3d3d3d3d3d3', x'c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1c1', x'b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2b2', ?, NULL, 1, 1)`,
			[]byte("a/b.flac"),
		)
		assertErrNames(t, err, "FOREIGN KEY constraint failed")
		if code := sqliteCode(err); code != sqliteFKCode {
			t.Errorf("code = %d, want %d", code, sqliteFKCode)
		}
	})

	t.Run("a collection owned by a missing user", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO collections (id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation, last_mirror_at, duplicate_collapse, layout_template, tagging_policy, admission_format_allowlist, admission_match_threshold)
			 VALUES (x'b3b3b3b3b3b3b3b3b3b3b3b3b3b3b3b3', 'Second', ?, x'99999999999999999999999999999999', 'group', 'server', 0, 0, NULL, 0, '', '', NULL, NULL)`,
			[]byte("second"),
		)
		assertErrNames(t, err, "FOREIGN KEY constraint failed")
		if code := sqliteCode(err); code != sqliteFKCode {
			t.Errorf("code = %d, want %d", code, sqliteFKCode)
		}
	})
}

// migrationHashes pins every embedded migration's SHA-256, each
// computed once by an independent sha256sum run over the merged file,
// never by this package's own readMigrations: an edit to a merged
// file, a trailing newline included, then fails this table instead of
// every existing store's ledger check at its next start.
var migrationHashes = map[string]string{
	"0001_ledger.sql":  "9c0accd382e16f0a546ade4e76231c2235977e27bfc30106f69fa4516566291a",
	"0002_jobs.sql":    "d91729ec63c3fa4132f4af9fc19325d2077a135f4116d7bdd13178a6bfa2e1d5",
	"0003_catalog.sql": "f962133f7366b5e941e2346fc1d2e3cd4defa743ec02902241408e9173a089c5",
}

// TestSchemaMigrationHashPins asserts readMigrations, which every
// production ledger check trusts, computes the pinned hash for each
// embedded file, and that migrationHashes names no file readMigrations
// does not also find.
func TestSchemaMigrationHashPins(t *testing.T) {
	files, err := readMigrations(migrationSource())
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	if len(files) != len(migrationHashes) {
		t.Fatalf("readMigrations found %d files, migrationHashes pins %d", len(files), len(migrationHashes))
	}
	for _, m := range files {
		want, ok := migrationHashes[m.name]
		if !ok {
			t.Errorf("%s: no pinned hash", m.name)
			continue
		}
		if m.sum != want {
			t.Errorf("%s: sha256 %s, want the pinned %s", m.name, m.sum, want)
		}
	}
}

// TestID asserts two successive NewID() calls return 16-byte values that
// differ, the randomness evidence, and that a minted id inserted into
// an entity table leaves rowid unequal to the id's own bytes read as
// an integer, which proves only that the id is not the table's rowid.
func TestID(t *testing.T) {
	a, b := NewID(), NewID()
	if len(a) != idLen {
		t.Fatalf("len(NewID()) = %d, want %d", len(a), idLen)
	}
	if bytes.Equal(a, b) {
		t.Fatalf("two successive NewID() calls returned the same value: %x", a)
	}

	l, h := newMigrateStore(t)
	ctx := context.Background()
	if err := migrate(ctx, h, l, readMigrationsT(t, migrationSourceThrough(t, 2)), defaultOpts()); err != nil {
		t.Fatalf("migrate to version 2: %v", err)
	}
	minted := NewID()
	if _, err := h.write.ExecContext(ctx,
		`INSERT INTO jobs (id, kind, at, state, owner_instance, actor, cancel_requested, progress, failure_code, failure_message, started_at, updated_at, finished_at)
		 VALUES (?, 'backup', 1, 'queued', 'inst', 'cli', 0, NULL, NULL, NULL, NULL, 1, NULL)`,
		minted,
	); err != nil {
		t.Fatalf("insert minted id: %v", err)
	}
	var rowid int64
	if err := h.write.QueryRowContext(ctx, "SELECT rowid FROM jobs WHERE id = ?", minted).Scan(&rowid); err != nil {
		t.Fatalf("read rowid: %v", err)
	}
	mintedAsInt := int64(minted[8])<<56 | int64(minted[9])<<48 | int64(minted[10])<<40 | int64(minted[11])<<32 |
		int64(minted[12])<<24 | int64(minted[13])<<16 | int64(minted[14])<<8 | int64(minted[15])
	if mintedAsInt == rowid {
		t.Errorf("minted id's low 8 bytes read as int64 (%d) coincided with rowid %d", mintedAsInt, rowid)
	}
}
