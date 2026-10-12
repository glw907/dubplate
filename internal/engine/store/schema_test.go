package store

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"
)

// sqliteUniqueCode through sqlitePrimaryKeyCode are SQLite's extended
// result codes this file pins beyond write_test.go's sqliteFKCode
// (sqlite.org/rescode.html): SQLITE_CONSTRAINT_UNIQUE; SQLITE_MISMATCH's
// type-affinity form, which a STRICT table answers when a bound value's
// storage class does not match a column's declared type;
// SQLITE_CONSTRAINT_NOTNULL; SQLITE_CONSTRAINT_CHECK; and
// SQLITE_CONSTRAINT_PRIMARYKEY, a second row's own extended code for a
// duplicate rowid-less primary key, distinct from SQLITE_CONSTRAINT_UNIQUE.
const (
	sqliteUniqueCode     = 2067
	sqliteDatatypeCode   = 3091
	sqliteNotNullCode    = 1299
	sqliteCheckCode      = 275
	sqlitePrimaryKeyCode = 1555
)

// fullyMigratedStore returns a store migrated to head and populated
// with every migration's fixture, so a test has one user
// (a1a1...a1), one collection (b1b1...b1) it owns, one album
// (c1c1...c1), and one track (d1d1...d1) to build on, plus each later
// migration's own fixture rows (0004's settings and audit rows, 0005's
// per-member and routed rows).
func fullyMigratedStore(t *testing.T) (leaf, *Handle) {
	t.Helper()
	l, h := newMigrateStore(t)
	ctx := context.Background()
	files := readMigrationsT(t, migrationSource())
	if err := migrate(ctx, h, l, files, defaultOpts()); err != nil {
		t.Fatalf("migrate to head: %v", err)
	}
	for _, m := range files {
		body, err := os.ReadFile(filepath.Join("testdata", "fixtures", m.name))
		if err != nil {
			t.Fatalf("read fixture %s: %v", m.name, err)
		}
		if _, err := h.write.ExecContext(ctx, string(body)); err != nil {
			t.Fatalf("apply fixture %s: %v", m.name, err)
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
// username is refused the same way.
func TestSchemaDuplicateAddressRefused(t *testing.T) {
	ctx := context.Background()

	t.Run("against a live row", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
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
			`INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
			 VALUES (?, 'gone@example.invalid', 'Gone', 'gone', 0, 999, 2, 2)`,
			NewID(),
		); err != nil {
			t.Fatalf("insert the tombstoned baseline: %v", err)
		}
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
			 VALUES (?, 'gone@example.invalid', 'Someone Else', 'someoneelse2', 0, NULL, 3, 3)`,
			NewID(),
		)
		assertErrNames(t, err, "UNIQUE constraint failed: users.address")
		if code := sqliteCode(err); code != sqliteUniqueCode {
			t.Errorf("code = %d, want %d", code, sqliteUniqueCode)
		}
	})

	t.Run("a case-folded username", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
			 VALUES (?, 'other@example.invalid', 'Other', 'Member1', 0, NULL, 2, 2)`,
			NewID(),
		)
		assertErrNames(t, err, "UNIQUE constraint failed: users.username")
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
	"0001_ledger.sql":   "9c0accd382e16f0a546ade4e76231c2235977e27bfc30106f69fa4516566291a",
	"0002_jobs.sql":     "d91729ec63c3fa4132f4af9fc19325d2077a135f4116d7bdd13178a6bfa2e1d5",
	"0003_catalog.sql":  "42834b8f62df9727bea686bbc02030f1b9af0de2a6c61a59050be72399c7b7fd",
	"0004_settings.sql": "9d312b17c391dc7e4a4059ac2c302d4affa381b5955fc1038dedb099e74bfc1f",
	"0005_routed.sql":   "e3eeb662f78b5142f23729d69fc59db332fc04305f1618269347e06a5d2a8a4c",
	"0006_intake.sql":   "e688d90185ae70c7ced94f775124d7e599961f50fa050a7f1b83c3378f728b38",
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

// TestSchemaMoveIntentPathRefusesGoString is spec criterion 11's
// checklist line repeated for the routed schema: move_intents.old_path
// and .new_path are BLOB columns on a STRICT table, so a Go string
// bound into either is refused under the datatype-mismatch code, the
// same one TestSchemaBlobColumnRefusesText proves for tracks.rel_path.
func TestSchemaMoveIntentPathRefusesGoString(t *testing.T) {
	for _, tt := range []struct {
		name    string
		oldPath any
		newPath any
	}{
		{"old_path", "a Go string, not a []byte", []byte("new/path.flac")},
		{"new_path", []byte("old/path.flac"), "a Go string, not a []byte"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, h := fullyMigratedStore(t)
			_, err := h.write.ExecContext(context.Background(),
				`INSERT INTO move_intents (id, at, kind, old_path, new_path, granularity, state)
				 VALUES (?, 1, 'rename', ?, ?, 'file', 'pending')`,
				NewID(), tt.oldPath, tt.newPath,
			)
			assertErrNames(t, err, "cannot store TEXT value in BLOB column")
			if code := sqliteCode(err); code != sqliteDatatypeCode {
				t.Errorf("code = %d, want %d", code, sqliteDatatypeCode)
			}
		})
	}
}

// TestSchemaMemberOverridesForeignKeyRefusals is criterion 2's proof
// over member_overrides: a user_id naming no users row is refused under
// the foreign-key code, and a NULL user_id is refused under the
// NOT NULL code (member_overrides has no other primary key to fall
// back to). Fixture state: the fully migrated store already carries a
// member_overrides row for the fixture's own user, so each subtest's
// user_id differs from it to isolate the failure under test.
func TestSchemaMemberOverridesForeignKeyRefusals(t *testing.T) {
	t.Run("a user_id with no users row", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(context.Background(),
			`INSERT INTO member_overrides (user_id, quota_bytes, may_upload, may_create_collections, upload_concurrency_cap, updated_at)
			 VALUES (?, NULL, NULL, NULL, NULL, 1)`,
			[]byte("\x99\x99\x99\x99\x99\x99\x99\x99\x99\x99\x99\x99\x99\x99\x99\x99"),
		)
		assertErrNames(t, err, "FOREIGN KEY constraint failed")
		if code := sqliteCode(err); code != sqliteFKCode {
			t.Errorf("code = %d, want %d", code, sqliteFKCode)
		}
	})

	t.Run("a NULL user_id", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(context.Background(),
			`INSERT INTO member_overrides (user_id, quota_bytes, may_upload, may_create_collections, upload_concurrency_cap, updated_at)
			 VALUES (NULL, NULL, NULL, NULL, NULL, 1)`,
		)
		assertErrNames(t, err, "NOT NULL constraint failed: member_overrides.user_id")
		if code := sqliteCode(err); code != sqliteNotNullCode {
			t.Errorf("code = %d, want %d", code, sqliteNotNullCode)
		}
		assertColumnNotNull(t, h, "member_overrides", "user_id")
	})
}

// assertColumnNotNull fails the test unless pragma_table_info reports
// table.column as declared NOT NULL, the schema-level half of a NOT
// NULL refusal proof: an insert failing is not proof the column is
// declared NOT NULL rather than refused some other way.
func assertColumnNotNull(t *testing.T, h *Handle, table, column string) {
	t.Helper()
	cols, err := tableColumns(context.Background(), h.write, table)
	if err != nil {
		t.Fatalf("read columns of %s: %v", table, err)
	}
	for _, c := range cols {
		if c.name == column {
			if !c.notNull {
				t.Errorf("%s.%s is not declared NOT NULL", table, column)
			}
			return
		}
	}
	t.Fatalf("%s has no column %s", table, column)
}

// TestSchemaCollectionVisibilityCheck is criterion 3's proof over
// collections.visibility (declared at 6a): 'public' and the empty
// string are both outside the CHECK's closed set and refused, while
// 'group' and 'private' are both accepted. Each subtest's root_path
// differs to keep the UNIQUE constraint from masking the row under
// test.
func TestSchemaCollectionVisibilityCheck(t *testing.T) {
	insert := func(t *testing.T, h *Handle, rootPath []byte, visibility string) error {
		t.Helper()
		_, err := h.write.ExecContext(context.Background(),
			`INSERT INTO collections (id, name, root_path, owner, visibility, writer, writer_epoch, mirror_generation, last_mirror_at, duplicate_collapse, layout_template, tagging_policy, admission_format_allowlist, admission_match_threshold)
			 VALUES (?, 'Second', ?, x'a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1a1', ?, 'server', 0, 0, NULL, 0, '', '', NULL, NULL)`,
			NewID(), rootPath, visibility,
		)
		return err
	}

	for _, tt := range []struct {
		visibility string
		wantErr    bool
	}{
		{"public", true},
		{"", true},
		{"group", false},
		{"private", false},
	} {
		t.Run("visibility "+tt.visibility, func(t *testing.T) {
			_, h := fullyMigratedStore(t)
			err := insert(t, h, []byte("second-"+tt.visibility), tt.visibility)
			if tt.wantErr {
				assertErrNames(t, err, "CHECK constraint failed")
				if code := sqliteCode(err); code != sqliteCheckCode {
					t.Errorf("code = %d, want %d", code, sqliteCheckCode)
				}
				return
			}
			if err != nil {
				t.Errorf("visibility %q: %v", tt.visibility, err)
			}
		})
	}
}

// TestSchemaAtKindNotNull is invariant 3's proof over the two routed
// event-shaped tables criterion 1 names: settings_audit (named _audit)
// and move_intents (carries a kind column), each refusing a NULL at or
// a NULL kind under the NOT NULL code, four subtests over the two
// tables and two columns.
func TestSchemaAtKindNotNull(t *testing.T) {
	tests := []struct {
		name   string
		table  string
		column string
		query  string
		args   []any
		want   string
	}{
		{
			name:   "settings_audit.at NULL",
			table:  "settings_audit",
			column: "at",
			query:  `INSERT INTO settings_audit (id, at, kind, actor, channel, key, old_value, new_value) VALUES (?, NULL, 'seed', 'system', 'env-seed', 'K', NULL, 'v')`,
			args:   []any{NewID()},
			want:   "NOT NULL constraint failed: settings_audit.at",
		},
		{
			name:   "settings_audit.kind NULL",
			table:  "settings_audit",
			column: "kind",
			query:  `INSERT INTO settings_audit (id, at, kind, actor, channel, key, old_value, new_value) VALUES (?, 1, NULL, 'system', 'env-seed', 'K', NULL, 'v')`,
			args:   []any{NewID()},
			want:   "NOT NULL constraint failed: settings_audit.kind",
		},
		{
			name:   "move_intents.at NULL",
			table:  "move_intents",
			column: "at",
			query:  `INSERT INTO move_intents (id, at, kind, old_path, new_path, granularity, state) VALUES (?, NULL, 'rename', ?, ?, 'file', 'pending')`,
			args:   []any{NewID(), []byte("a/b"), []byte("a/c")},
			want:   "NOT NULL constraint failed: move_intents.at",
		},
		{
			name:   "move_intents.kind NULL",
			table:  "move_intents",
			column: "kind",
			query:  `INSERT INTO move_intents (id, at, kind, old_path, new_path, granularity, state) VALUES (?, 1, NULL, ?, ?, 'file', 'pending')`,
			args:   []any{NewID(), []byte("a/b"), []byte("a/c")},
			want:   "NOT NULL constraint failed: move_intents.kind",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := fullyMigratedStore(t)
			_, err := h.write.ExecContext(context.Background(), tt.query, tt.args...)
			assertErrNames(t, err, tt.want)
			if code := sqliteCode(err); code != sqliteNotNullCode {
				t.Errorf("code = %d, want %d", code, sqliteNotNullCode)
			}
			assertColumnNotNull(t, h, tt.table, tt.column)
		})
	}
}

// TestSchemaSettingsSourceCheck is R-6c6's proof that settings.source
// admits only 'env-seed' and 'admin': a row stored with no source means
// the value is the registry's default (R-6c7), so 'default' as a
// literal source value is outside the CHECK's closed set.
func TestSchemaSettingsSourceCheck(t *testing.T) {
	_, h := fullyMigratedStore(t)
	_, err := h.write.ExecContext(context.Background(),
		`INSERT INTO settings (key, value, source, seed_value, actor, updated_at) VALUES ('K', 'v', 'default', 'v', 'system', 1)`,
	)
	assertErrNames(t, err, "CHECK constraint failed")
	if code := sqliteCode(err); code != sqliteCheckCode {
		t.Errorf("code = %d, want %d", code, sqliteCheckCode)
	}
}

// TestSchemaSettingsGenerationSingleton is R-6c6's proof that
// settings_generation never holds more than the one row the migration
// inserts: a second row under a different key is outside the CHECK
// pinning singleton to 1, and a second row repeating the pinned key
// collides on the primary key itself, a different extended code from
// the CHECK refusal.
func TestSchemaSettingsGenerationSingleton(t *testing.T) {
	t.Run("a second row, a different key", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(context.Background(),
			`INSERT INTO settings_generation (singleton, generation) VALUES (2, 0)`,
		)
		assertErrNames(t, err, "CHECK constraint failed")
		if code := sqliteCode(err); code != sqliteCheckCode {
			t.Errorf("code = %d, want %d", code, sqliteCheckCode)
		}
	})

	t.Run("a second row repeating the pinned key", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		_, err := h.write.ExecContext(context.Background(),
			`INSERT INTO settings_generation (singleton, generation) VALUES (1, 99)`,
		)
		assertErrNames(t, err, "UNIQUE constraint failed: settings_generation.singleton")
		if code := sqliteCode(err); code != sqlitePrimaryKeyCode {
			t.Errorf("code = %d, want %d", code, sqlitePrimaryKeyCode)
		}
	})
}
