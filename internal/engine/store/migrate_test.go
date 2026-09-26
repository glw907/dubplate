package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	sqlite "modernc.org/sqlite"
)

// migrateBusy is the busy timeout every migration test opens with.
const migrateBusy = 2 * time.Second

// baseMigration is the first file of every synthetic set: the ledger,
// the store marker (0x64756270), and a parent and child pair the
// populated fixtures and the rebuild rows use.
const baseMigration = `CREATE TABLE migrations (version INTEGER PRIMARY KEY, name TEXT NOT NULL, sha256 TEXT NOT NULL) STRICT;
PRAGMA application_id = 1685414512;
CREATE TABLE parent (id INTEGER PRIMARY KEY, label TEXT NOT NULL) STRICT;
CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id)) STRICT;
`

// tableMigration is the body of synthetic file n beyond the first.
func tableMigration(n int) string {
	return fmt.Sprintf("CREATE TABLE t%d (a INTEGER);\n", n)
}

// synthName is the file name the synthetic sets give version n.
func synthName(n int) string {
	return fmt.Sprintf("%04d_step.sql", n)
}

// synthSet returns a migration source holding baseMigration and then
// one tableMigration per version up to head.
func synthSet(head int) fstest.MapFS {
	src := fstest.MapFS{synthName(1): {Data: []byte(baseMigration)}}
	for n := 2; n <= head; n++ {
		src[synthName(n)] = &fstest.MapFile{Data: []byte(tableMigration(n))}
	}
	return src
}

// withSecond returns a copy of src whose version 2 is body under name.
func withSecond(src fstest.MapFS, name, body string) fstest.MapFS {
	out := fstest.MapFS{}
	for k, v := range src {
		if k != synthName(2) {
			out[k] = v
		}
	}
	out[name] = &fstest.MapFile{Data: []byte(body)}
	return out
}

// defaultOpts keeps production's default, backup off, so a migration
// prunes to keep. A row that needs no prune sets opts.backup = true by
// name.
func defaultOpts() migrateOptions {
	return migrateOptions{busyTimeout: migrateBusy, keep: 2}
}

// newMigrateStore makes a leaf under a fresh state directory with a
// pre-created empty database, the shape serve leaves before its first
// migration, and opens a handle on it, closed at cleanup.
func newMigrateStore(t *testing.T) (leaf, *Handle) {
	t.Helper()
	stateDir := t.TempDir()
	l := newLeaf(stateDir)
	if err := l.create(stateDir); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := preCreate(l.dbPath); err != nil {
		t.Fatalf("preCreate: %v", err)
	}
	h, err := openHandle(l.dbPath, migrateBusy)
	if err != nil {
		t.Fatalf("openHandle: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return l, h
}

// readMigrationsT reads src and fails the test on an error, so a
// migrate call site parses its own source, the way a production
// caller does, rather than sharing one cached parse across a table.
func readMigrationsT(t *testing.T, src fs.FS) []migration {
	t.Helper()
	files, err := readMigrations(src)
	if err != nil {
		t.Fatalf("readMigrations: %v", err)
	}
	return files
}

// mustMigrate reads src and runs migrate, failing the test on either
// error.
func mustMigrate(t *testing.T, h *Handle, l leaf, src fstest.MapFS, opts migrateOptions) {
	t.Helper()
	if err := migrate(context.Background(), h, l, readMigrationsT(t, src), opts); err != nil {
		t.Fatalf("migrate: %v", err)
	}
}

// rawExec runs stmts on a connection of its own, the way a test plants
// a state no runner would write.
func rawExec(t *testing.T, path, stmts string) {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, maintenanceDSN(path, migrateBusy))
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec(stmts); err != nil {
		t.Fatalf("raw exec %q: %v", stmts, err)
	}
}

// populate inserts three parents and a child of each, so a copy that
// lost its rows reads differently from the store.
func populate(t *testing.T, path string) {
	t.Helper()
	rawExec(t, path, `INSERT INTO parent (id, label) VALUES (1, 'a'), (2, 'b'), (3, 'c');
INSERT INTO child (id, parent_id) VALUES (1, 1), (2, 2), (3, 3);`)
}

// queryInt reads one integer from path through a read-only connection
// of its own.
func queryInt(t *testing.T, path, query string) int {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, inspectDSN(path, migrateBusy))
	if err != nil {
		t.Fatalf("open inspect: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.QueryRow(query).Scan(&n); err != nil {
		t.Fatalf("%s on %s: %v", query, path, err)
	}
	return n
}

// storeState is what a failed or refused migration must leave as it
// found: the version, every schema object, and every ledger row.
type storeState struct {
	version int
	schema  []string
	ledger  []string
}

// readState reads path's storeState through a read-only connection.
func readState(t *testing.T, path string) storeState {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, inspectDSN(path, migrateBusy))
	if err != nil {
		t.Fatalf("open inspect: %v", err)
	}
	defer func() { _ = db.Close() }()

	var s storeState
	if err := db.QueryRow("PRAGMA user_version").Scan(&s.version); err != nil {
		t.Fatalf("user_version: %v", err)
	}
	s.schema = queryStrings(t, db, "SELECT type || ' ' || name || ' ' || coalesce(sql, '') FROM sqlite_schema ORDER BY type, name")
	var ledgers int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_schema WHERE name = 'migrations'").Scan(&ledgers); err != nil {
		t.Fatalf("find ledger: %v", err)
	}
	if ledgers == 1 {
		s.ledger = queryStrings(t, db, "SELECT version || ' ' || name || ' ' || sha256 FROM migrations ORDER BY version")
	}
	return s
}

// queryStrings collects a one-column query's rows.
func queryStrings(t *testing.T, db *sql.DB, query string) []string {
	t.Helper()
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	defer func() { _ = rows.Close() }()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatalf("%s: scan: %v", query, err)
		}
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("%s: %v", query, err)
	}
	return out
}

// assertUnchanged fails the test unless path's state equals before.
func assertUnchanged(t *testing.T, path string, before storeState) {
	t.Helper()
	after := readState(t, path)
	if after.version != before.version {
		t.Errorf("user_version = %d, want %d unchanged", after.version, before.version)
	}
	if !slices.Equal(after.schema, before.schema) {
		t.Errorf("schema changed:\nbefore %q\nafter  %q", before.schema, after.schema)
	}
	if !slices.Equal(after.ledger, before.ledger) {
		t.Errorf("ledger changed:\nbefore %q\nafter  %q", before.ledger, after.ledger)
	}
}

// listDir names dir's entries, sorted.
func listDir(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read %s: %v", dir, err)
	}
	var names []string
	for _, e := range entries {
		names = append(names, e.Name())
	}
	return names
}

// assertErrNames fails the test unless err is non-nil and its text
// carries every one of names.
func assertErrNames(t *testing.T, err error, names ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got nil error, want one naming %q", names)
	}
	for _, n := range names {
		if !strings.Contains(err.Error(), n) {
			t.Errorf("error %q does not name %q", err, n)
		}
	}
}

func TestMigrate_Versions(t *testing.T) {
	tests := []struct {
		name string
		// setup brings a fresh store to the row's starting state.
		setup         func(t *testing.T, l leaf, h *Handle)
		wantErr       []string
		wantVersion   int
		wantSnapshots []string
	}{
		{
			name:        "empty store applies every file and takes no copy",
			setup:       func(*testing.T, leaf, *Handle) {},
			wantVersion: 3,
		},
		{
			name: "older store is copied at its version, then brought to the head",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(1), defaultOpts())
				populate(t, l.dbPath)
			},
			wantVersion:   3,
			wantSnapshots: []string{"pre-migrate-v1.db"},
		},
		{
			name: "current store applies nothing",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(3), defaultOpts())
			},
			wantVersion: 3,
		},
		{
			// A newer binary leaves ledger rows this binary has no file
			// for; the version decides before the ledger is read, so a
			// row that would fail every ledger check names nothing.
			name: "newer store is refused on its version alone",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(3), defaultOpts())
				rawExec(t, l.dbPath, `PRAGMA user_version = 4;
INSERT INTO migrations (version, name, sha256) VALUES (4, '0004_ahead.sql', 'x');
UPDATE migrations SET sha256 = 'tampered' WHERE version = 1;`)
			},
			wantErr:     []string{"version 4", "version 3", "newer"},
			wantVersion: 4,
		},
		{
			// A negative version indexes no file; refused rather than
			// reaching the apply loop.
			name: "negative version is refused",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(3), defaultOpts())
				rawExec(t, l.dbPath, `PRAGMA user_version = -1`)
			},
			wantErr:     []string{"version -1"},
			wantVersion: -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := newMigrateStore(t)
			tt.setup(t, l, h)
			before := readState(t, l.dbPath)

			err := migrate(context.Background(), h, l, readMigrationsT(t, synthSet(3)), defaultOpts())
			if tt.wantErr != nil {
				assertErrNames(t, err, tt.wantErr...)
				if strings.Contains(err.Error(), "ledger") {
					t.Errorf("error %q names the ledger, want the version refusal alone", err)
				}
				assertUnchanged(t, l.dbPath, before)
			} else if err != nil {
				t.Fatalf("migrate: %v", err)
			}

			got := readState(t, l.dbPath)
			if got.version != tt.wantVersion {
				t.Errorf("user_version = %d, want %d", got.version, tt.wantVersion)
			}
			if tt.wantErr == nil && len(got.ledger) != tt.wantVersion {
				t.Errorf("ledger rows = %q, want %d", got.ledger, tt.wantVersion)
			}
			if snaps := listDir(t, l.snapshotsDir); !slices.Equal(snaps, tt.wantSnapshots) {
				t.Errorf("snapshots = %q, want %q", snaps, tt.wantSnapshots)
			}
		})
	}
}

// TestMigrate_FailureLeavesPriorState is spec criterion 8's first half:
// a file that fails after a version write leaves the version, the
// schema, and the ledger as the run found them. Fixture: a store at
// the synthetic set's version 1.
func TestMigrate_FailureLeavesPriorState(t *testing.T) {
	tests := []struct {
		name    string
		file2   string
		wantErr string
	}{
		{
			// The runner's own PRAGMA user_version = 2 runs, then its
			// ledger insert fails on the file's trigger.
			name: "after the runner's version write",
			file2: `CREATE TABLE t2 (a INTEGER);
CREATE TRIGGER refuse_ledger BEFORE INSERT ON migrations BEGIN SELECT RAISE(ABORT, 'injected'); END;
`,
			wantErr: "injected",
		},
		{
			name: "inside the file, after the file's own version write",
			file2: `PRAGMA user_version = 2;
CREATE TABLE t2 (a INTEGER);
INSERT INTO no_such_table (a) VALUES (1);
`,
			wantErr: "no_such_table",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := newMigrateStore(t)
			mustMigrate(t, h, l, synthSet(1), defaultOpts())
			before := readState(t, l.dbPath)

			src := withSecond(synthSet(2), synthName(2), tt.file2)
			err := migrate(context.Background(), h, l, readMigrationsT(t, src), defaultOpts())
			assertErrNames(t, err, tt.wantErr, synthName(2))
			assertUnchanged(t, l.dbPath, before)
		})
	}
}

// TestMigrate_LedgerRefusals covers every way the ledger can disagree
// with the binary's files. Each row's store would migrate if the check
// were skipped: the set always holds a file above user_version.
func TestMigrate_LedgerRefusals(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, l leaf, h *Handle)
		src     fstest.MapFS
		wantErr []string
	}{
		{
			// The two hashes are sha256sum's output for file 2's bytes
			// with one trailing newline and with two.
			name: "a row's hash differs from the file's",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(2), defaultOpts())
			},
			src: withSecond(synthSet(3), synthName(2), tableMigration(2)+"\n"),
			wantErr: []string{
				"74ca2fcad4f28c82a4e46e098aabc20cb32e7494dc02879957a829a37f209ca1",
				"56664e60711e19c791503733270375c01003a7c1ec8b80452754912929ac2110",
			},
		},
		{
			name: "a row's name differs from the file's",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(2), defaultOpts())
			},
			src:     withSecond(synthSet(3), "0002_renamed.sql", tableMigration(2)),
			wantErr: []string{synthName(2), "0002_renamed.sql"},
		},
		{
			name: "a row above user_version",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(1), defaultOpts())
				rawExec(t, l.dbPath, `INSERT INTO migrations (version, name, sha256) VALUES (2, '0002_step.sql', 'x')`)
			},
			src:     synthSet(3),
			wantErr: []string{"version 2", "user_version 1"},
		},
		{
			name: "a version at or below user_version with no row",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(2), defaultOpts())
				rawExec(t, l.dbPath, `CREATE TABLE t3 (a INTEGER); PRAGMA user_version = 3;`)
			},
			src:     synthSet(4),
			wantErr: []string{"version 3", "user_version 3"},
		},
		{
			name: "a row for a version with no file",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(1), defaultOpts())
				rawExec(t, l.dbPath, `INSERT INTO migrations (version, name, sha256) VALUES (0, '0000_none.sql', 'x')`)
			},
			src:     synthSet(2),
			wantErr: []string{"version 0", "no migration file"},
		},
		{
			name: "the ledger table is missing",
			setup: func(t *testing.T, l leaf, h *Handle) {
				mustMigrate(t, h, l, synthSet(1), defaultOpts())
				rawExec(t, l.dbPath, `DROP TABLE migrations`)
			},
			src:     synthSet(2),
			wantErr: []string{"ledger table migrations is missing", "user_version 1"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := newMigrateStore(t)
			tt.setup(t, l, h)
			before := readState(t, l.dbPath)

			err := migrate(context.Background(), h, l, readMigrationsT(t, tt.src), defaultOpts())
			assertErrNames(t, err, tt.wantErr...)
			assertUnchanged(t, l.dbPath, before)
			if snaps := listDir(t, l.snapshotsDir); len(snaps) != 0 {
				t.Errorf("snapshots = %q, want none: the ledger check runs before the copy", snaps)
			}
		})
	}
}

// TestMigrate_FileEndingItsOwnTransaction is R-6c12's savepoint check:
// a file whose body ends the runner's transaction (COMMIT, optionally
// followed by BEGIN, or ROLLBACK) fails at RELEASE, before the version
// write or the ledger row, naming the file and the release failure
// rather than the whole error text, since the runner's own rollback may
// join a second error over the transaction the body already ended
// (probe rows 2 to 4). A body that reopens a transaction is caught the
// same way: RELEASE finds no savepoint of the runner's name to release.
func TestMigrate_FileEndingItsOwnTransaction(t *testing.T) {
	tests := []struct {
		name  string
		file2 string
	}{
		{name: "COMMIT then BEGIN and another table", file2: "CREATE TABLE e3 (a INTEGER);\nCOMMIT;\nBEGIN;\nCREATE TABLE e4 (a INTEGER);\n"},
		{name: "COMMIT with nothing reopened", file2: "CREATE TABLE e3 (a INTEGER);\nCOMMIT;\n"},
		{name: "ROLLBACK", file2: "CREATE TABLE e3 (a INTEGER);\nROLLBACK;\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := newMigrateStore(t)
			mustMigrate(t, h, l, synthSet(1), defaultOpts())

			// Not assertUnchanged: a body ending in a bare COMMIT (or
			// COMMIT then ROLLBACK) has already persisted whatever it
			// created before the runner's own RELEASE runs, so the
			// schema is not the pre-apply state; only the file name and
			// the release failure are asserted here (the plan's own
			// reading of the mechanics probe's rows 2 to 4).
			src := withSecond(synthSet(2), synthName(2), tt.file2)
			err := migrate(context.Background(), h, l, readMigrationsT(t, src), defaultOpts())
			assertErrNames(t, err, synthName(2), "release savepoint")
		})
	}
}

// TestMigrate_CommitThenBeginLeavesNoLedgerRow is the mechanics probe's
// clean case: a body that ends the runner's transaction with COMMIT and
// reopens one with BEGIN never reaches the version write or the ledger
// insert, so the runner's own rollback undoes only what BEGIN reopened,
// and the body's own COMMIT stands.
func TestMigrate_CommitThenBeginLeavesNoLedgerRow(t *testing.T) {
	l, h := newMigrateStore(t)
	mustMigrate(t, h, l, synthSet(1), defaultOpts())
	before := readState(t, l.dbPath)

	src := withSecond(synthSet(2), synthName(2), "CREATE TABLE e3 (a INTEGER);\nCOMMIT;\nBEGIN;\nCREATE TABLE e4 (a INTEGER);\n")
	err := migrate(context.Background(), h, l, readMigrationsT(t, src), defaultOpts())
	assertErrNames(t, err, synthName(2), "release savepoint")

	got := readState(t, l.dbPath)
	if got.version != before.version {
		t.Errorf("user_version = %d, want %d unchanged: the runner never reaches its own version write", got.version, before.version)
	}
	if !slices.Equal(got.ledger, before.ledger) {
		t.Errorf("ledger = %q, want %q unchanged: the runner never reaches its own ledger insert", got.ledger, before.ledger)
	}
	if n := queryInt(t, l.dbPath, "SELECT count(*) FROM sqlite_schema WHERE type = 'table' AND name = 'e3'"); n != 1 {
		t.Errorf("table e3 exists = %d, want 1: the body's own COMMIT persisted it before RELEASE ran", n)
	}
	if n := queryInt(t, l.dbPath, "SELECT count(*) FROM sqlite_schema WHERE type = 'table' AND name = 'e4'"); n != 0 {
		t.Errorf("table e4 exists = %d, want 0: the runner's rollback undid the reopened transaction", n)
	}
}

// TestMigrate_SavepointRedeclaredDefeatsRelease is R-6c12's named
// residual: a body that ends the runner's transaction and then
// redeclares a savepoint of the runner's own name passes RELEASE, so
// the version write and the ledger insert run and commit outside the
// runner's transaction before the runner's own Commit call finds
// nothing open, which commitMigration's "no transaction is active"
// branch still catches.
func TestMigrate_SavepointRedeclaredDefeatsRelease(t *testing.T) {
	l, h := newMigrateStore(t)
	mustMigrate(t, h, l, synthSet(1), defaultOpts())

	src := withSecond(synthSet(2), synthName(2), "CREATE TABLE e3 (a INTEGER);\nCOMMIT;\nSAVEPOINT "+runnerSavepoint+";\n")
	err := migrate(context.Background(), h, l, readMigrationsT(t, src), defaultOpts())
	assertErrNames(t, err, synthName(2), "the file ends its own transaction")

	got := readState(t, l.dbPath)
	if got.version != 2 {
		t.Errorf("user_version = %d, want 2: the version write committed outside the runner's own transaction", got.version)
	}
	if len(got.ledger) != 2 {
		t.Errorf("ledger rows = %q, want 2: the ledger insert committed outside the runner's own transaction", got.ledger)
	}
}

// TestMigrate_TriggerBodyApplies is R-6c12's non-refusal case: a
// trigger's own BEGIN ... END body runs inside the CREATE TRIGGER
// statement, never as a standalone transaction control statement, so it
// does not touch the runner's savepoint and the file applies normally.
func TestMigrate_TriggerBodyApplies(t *testing.T) {
	l, h := newMigrateStore(t)
	mustMigrate(t, h, l, synthSet(1), defaultOpts())

	body := "CREATE TABLE e3 (a INTEGER);\n" +
		"CREATE TRIGGER e3_guard AFTER INSERT ON e3 BEGIN SELECT 1; END;\n"
	mustMigrate(t, h, l, withSecond(synthSet(2), synthName(2), body), defaultOpts())

	got := readState(t, l.dbPath)
	if got.version != 2 {
		t.Errorf("user_version = %d, want 2", got.version)
	}
	if len(got.ledger) != 2 {
		t.Errorf("ledger rows = %q, want 2", got.ledger)
	}
}

func TestMigrate_KeepBelowOneRefusedBeforeApply(t *testing.T) {
	for _, keep := range []int{0, -1} {
		t.Run(fmt.Sprintf("keep %d", keep), func(t *testing.T) {
			l, h := newMigrateStore(t)
			before := readState(t, l.dbPath)

			opts := defaultOpts()
			opts.keep = keep
			err := migrate(context.Background(), h, l, readMigrationsT(t, synthSet(2)), opts)
			assertErrNames(t, err, fmt.Sprintf("keep %d", keep))
			assertUnchanged(t, l.dbPath, before)
		})
	}
}

// TestMigrate_Rebuild is spec criterion 10: a rebuild runs with foreign
// keys off on a connection that never reaches a pool, so the write
// method still refuses an orphan afterward; and a rebuild that leaves
// an orphan behind commits nothing. Fixture: a store at version 1 whose
// three children each name a parent.
func TestMigrate_Rebuild(t *testing.T) {
	// A rebuild that drops parent while children name it: with
	// enforcement on, DROP TABLE's implicit delete fails with 787.
	rebuild := func(where string) string {
		return rebuildMarker + `
CREATE TABLE parent_new (id INTEGER PRIMARY KEY, label TEXT NOT NULL, rank INTEGER NOT NULL DEFAULT 0) STRICT;
INSERT INTO parent_new (id, label) SELECT id, label FROM parent` + where + `;
DROP TABLE parent;
ALTER TABLE parent_new RENAME TO parent;
`
	}

	t.Run("rebuild applies and enforcement survives it", func(t *testing.T) {
		l, h := newMigrateStore(t)
		mustMigrate(t, h, l, synthSet(1), defaultOpts())
		populate(t, l.dbPath)

		mustMigrate(t, h, l, withSecond(synthSet(2), synthName(2), rebuild("")), defaultOpts())

		if got := queryInt(t, l.dbPath, "SELECT count(*) FROM pragma_table_info('parent') WHERE name = 'rank'"); got != 1 {
			t.Errorf("parent's rank column count = %d, want 1: the rebuild did not apply", got)
		}
		if got := queryInt(t, l.dbPath, "SELECT count(*) FROM child JOIN parent ON parent.id = child.parent_id"); got != 3 {
			t.Errorf("children joined to a parent = %d, want 3", got)
		}

		h.version = 2
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		err := h.Update(ctx, func(ctx context.Context, tx *Tx) error {
			_, err := tx.Exec(ctx, "INSERT INTO child (id, parent_id) VALUES (9, 99)")
			return err
		})
		var se *sqlite.Error
		if !errors.As(err, &se) || se.Code() != 787 {
			t.Fatalf("orphan insert after the rebuild: err = %v, want SQLITE_CONSTRAINT_FOREIGNKEY (787)", err)
		}
	})

	t.Run("rebuild that leaves an orphan commits nothing", func(t *testing.T) {
		l, h := newMigrateStore(t)
		mustMigrate(t, h, l, synthSet(1), defaultOpts())
		populate(t, l.dbPath)
		before := readState(t, l.dbPath)

		err := migrate(context.Background(), h, l, readMigrationsT(t, withSecond(synthSet(2), synthName(2), rebuild(" WHERE id = 1"))), defaultOpts())
		assertErrNames(t, err, synthName(2), "table child")
		assertUnchanged(t, l.dbPath, before)
		if got := queryInt(t, l.dbPath, "SELECT count(*) FROM parent"); got != 3 {
			t.Errorf("parent rows = %d, want 3", got)
		}
	})
}

// TestMigrate_FirstMigrationLeavesWAL: the runner's first statement
// reads user_version through the write pool, whose DSN sets WAL, so a
// pre-created empty file is in WAL after its first migration even when
// that file runs on the maintenance class, whose DSN sets no journal
// mode. SQLite's file format puts the read and write versions at header
// bytes 18 and 19: 2 for WAL, 1 for a rollback journal.
func TestMigrate_FirstMigrationLeavesWAL(t *testing.T) {
	tests := []struct {
		name  string
		first string
	}{
		{name: "first file on the write pool", first: baseMigration},
		{name: "first file on the maintenance class", first: rebuildMarker + "\n" + baseMigration},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l, h := newMigrateStore(t)
			mustMigrate(t, h, l, fstest.MapFS{synthName(1): {Data: []byte(tt.first)}}, defaultOpts())
			if err := h.Close(); err != nil {
				t.Fatalf("close handle: %v", err)
			}

			// Every connection in the test process is closed, so an os
			// read of the file cannot drop a lock another holds.
			header, err := os.ReadFile(l.dbPath)
			if err != nil {
				t.Fatalf("read %s: %v", l.dbPath, err)
			}
			if len(header) < 20 || header[18] != 2 || header[19] != 2 {
				t.Errorf("header bytes 18 and 19 = %v, want [2 2] (WAL)", header[18:20])
			}
		})
	}
}

func TestReadMigrations(t *testing.T) {
	file := &fstest.MapFile{Data: []byte("SELECT 1;\n")}
	// second returns a set whose file 2 begins with firstLine.
	second := func(firstLine string) fstest.MapFS {
		return fstest.MapFS{"0001_a.sql": file, "0002_b.sql": {Data: []byte(firstLine + "\nSELECT 1;\n")}}
	}
	tests := []struct {
		name         string
		src          fstest.MapFS
		wantErr      string
		wantVersions []int
		wantRebuild  []bool
	}{
		{
			name:         "unpadded versions order by number",
			src:          fstest.MapFS{"1_a.sql": file, "2_b.sql": file, "10_j.sql": file, "3_c.sql": file, "4_d.sql": file, "5_e.sql": file, "6_f.sql": file, "7_g.sql": file, "8_h.sql": file, "9_i.sql": file},
			wantVersions: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			wantRebuild:  make([]bool, 10),
		},
		{
			name:         "the exact rebuild marker",
			src:          second("-- dubplate:foreign-keys-off"),
			wantVersions: []int{1, 2},
			wantRebuild:  []bool{false, true},
		},
		{
			name:         "a first line that only begins like the marker",
			src:          second("-- dubplate:foreign-keys-offline"),
			wantVersions: []int{1, 2},
			wantRebuild:  []bool{false, false},
		},
		{name: "the marker ending in a carriage return", src: second("-- dubplate:foreign-keys-off\r"), wantErr: "0002_b.sql"},
		{name: "the marker ending in a space", src: second("-- dubplate:foreign-keys-off "), wantErr: "0002_b.sql"},
		{name: "the marker ending in a tab and a carriage return", src: second("-- dubplate:foreign-keys-off\t\r"), wantErr: "0002_b.sql"},
		{name: "a name that is not a migration file", src: fstest.MapFS{"0001_a.sql": file, "notes.txt": file}, wantErr: "notes.txt"},
		{name: "a gap", src: fstest.MapFS{"0001_a.sql": file, "0003_c.sql": file}, wantErr: "0003_c.sql has version 3, want 2"},
		{name: "a repeat", src: fstest.MapFS{"0001_a.sql": file, "1_b.sql": file}, wantErr: "version 1, want 2"},
		{name: "no version 1", src: fstest.MapFS{"0002_b.sql": file}, wantErr: "0002_b.sql has version 2, want 1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := readMigrations(tt.src)
			if tt.wantErr != "" {
				assertErrNames(t, err, tt.wantErr)
				return
			}
			if err != nil {
				t.Fatalf("readMigrations: %v", err)
			}
			var (
				got     []int
				rebuild []bool
			)
			for _, m := range files {
				got = append(got, m.version)
				rebuild = append(rebuild, m.rebuild)
			}
			if !slices.Equal(got, tt.wantVersions) {
				t.Errorf("versions = %v, want %v", got, tt.wantVersions)
			}
			if !slices.Equal(rebuild, tt.wantRebuild) {
				t.Errorf("rebuild = %v, want %v", rebuild, tt.wantRebuild)
			}
		})
	}
}
