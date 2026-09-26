package store

import (
	"cmp"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	sqlite "modernc.org/sqlite"
)

// storeMarker is the PRAGMA application_id every dubplate store
// carries, the ASCII bytes "dubp". The first migration sets it; a copy
// that does not read it back is not a copy of a dubplate store.
const storeMarker = 0x64756270

// ledgerTable is the applied-migration ledger the first migration
// creates: one row per applied file, keyed by version, holding the
// file's name and the lowercase hex SHA-256 of its bytes.
const ledgerTable = "migrations"

// rebuildMarker is the first line of a migration file that must run
// with foreign-key enforcement off.
const rebuildMarker = "-- dubplate:foreign-keys-off"

// sqliteErrorCode is SQLite's primary result code SQLITE_ERROR, the
// code a COMMIT with no open transaction answers.
const sqliteErrorCode = 1

// runnerSavepoint is the name applyInTx opens around each file's body
// and expects to still exist when the body returns. No shipped
// migration may declare a savepoint of this name: a body that runs
// COMMIT (ending the runner's own transaction) drops the savepoint
// with it, so RELEASE fails and applyInTx refuses the file before its
// version write or ledger row (R-6c12). A body that reopens a
// transaction and declares its own savepoint of this same name defeats
// the check; commitMigration's "no transaction is active" branch below
// is the only detection left for that shape.
const runnerSavepoint = "dubplate_runner"

// migrationName matches a migration file's name: a decimal version,
// "_", a label, and ".sql".
var migrationName = regexp.MustCompile(`^([0-9]+)_[^/]+\.sql$`)

// migration is one numbered SQL file of a migration source.
type migration struct {
	version int
	name    string
	sum     string
	body    string
	rebuild bool
}

// migrateOptions carries the runner's thresholds, each the caller's:
// the store reads no environment.
type migrateOptions struct {
	// busyTimeout is the busy timeout of every connection the runner
	// opens beyond h's own pools.
	busyTimeout time.Duration

	// backup reports whether remote backup is on. With it off, the
	// runner prunes pre-migration copies after a migration.
	backup bool

	// keep is how many pre-migration copies survive a prune. It must
	// be at least 1, so a prune never deletes the copy just taken.
	keep int
}

// migrate brings the store h opened, whose leaf is l, to the highest
// version files holds. files is the caller's already-parsed migration
// source, each named "<version>_<label>.sql", numbered from 1 with no
// gap: migrate parses nothing itself, so a caller that already read
// its source (OpenServe's settle, above all) never reads it twice.
//
// It refuses, changing nothing, a store whose user_version is above
// files' highest, and a store whose ledger disagrees with files: a row
// whose name or hash differs from the file at its version, a row above
// user_version or with no file, or a version from 1 through
// user_version with no row. Before applying anything to a store at a
// version above 0, it takes a verified copy of the store as
// snapshots/pre-migrate-v<N>.db, replacing an existing copy of the same
// version.
//
// Each file applies in its own IMMEDIATE transaction that also sets
// user_version and inserts the file's ledger row, so a file that fails
// leaves the store at the prior version, schema, and ledger. The write
// pool runs every file except one whose first line is the rebuild
// marker, which runs on a maintenance connection with foreign keys off
// and closes that connection afterward. The first statement migrate
// issues reads user_version through the write pool, whose DSN sets WAL,
// so a store is in WAL before its first file applies on either class.
func migrate(ctx context.Context, h *Handle, l leaf, files []migration, opts migrateOptions) error {
	if opts.keep < 1 {
		return fmt.Errorf("migrate %s: keep %d copies, want at least 1", h.path, opts.keep)
	}
	head := len(files)

	var current int
	if err := h.write.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("migrate %s: read user_version: %w", h.path, classifyErr(err))
	}
	if current < 0 {
		return fmt.Errorf("migrate %s: store at version %d, below any version a migration sets", h.path, current)
	}
	if current > head {
		return fmt.Errorf("migrate %s: store at version %d is newer than this binary's version %d", h.path, current, head)
	}
	if err := checkLedger(ctx, h.write, files, current); err != nil {
		return fmt.Errorf("migrate %s: %w", h.path, err)
	}
	if current == head {
		return nil
	}

	if current > 0 {
		if err := takeSnapshot(ctx, l, opts.busyTimeout, preMigrateName(current), current); err != nil {
			return fmt.Errorf("migrate %s: %w", h.path, err)
		}
	}
	for _, m := range files[current:] {
		if err := applyMigration(ctx, h, m, opts.busyTimeout); err != nil {
			return fmt.Errorf("migrate %s: %w", h.path, err)
		}
	}
	if !opts.backup {
		if err := pruneSnapshots(l.snapshotsDir, opts.keep); err != nil {
			return fmt.Errorf("migrate %s: %w", h.path, err)
		}
	}
	return nil
}

// readMigrations reads every file in src's root, in version order,
// refusing a name that is not a migration file's, a file whose first
// line misses the rebuild marker only by trailing whitespace, and a set
// whose versions do not run from 1 without a gap or a repeat.
func readMigrations(src fs.FS) ([]migration, error) {
	entries, err := fs.ReadDir(src, ".")
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}
	files := make([]migration, 0, len(entries))
	for _, e := range entries {
		match := migrationName.FindStringSubmatch(e.Name())
		if match == nil || !e.Type().IsRegular() {
			return nil, fmt.Errorf("read migrations: %s is not a migration file named <version>_<label>.sql", e.Name())
		}
		v, err := strconv.Atoi(match[1])
		if err != nil {
			return nil, fmt.Errorf("read migrations: %s: version: %w", e.Name(), err)
		}
		body, err := fs.ReadFile(src, e.Name())
		if err != nil {
			return nil, fmt.Errorf("read migrations: %w", err)
		}
		first, _, _ := strings.Cut(string(body), "\n")
		if first != rebuildMarker && nearMarker(first) {
			return nil, fmt.Errorf("read migrations: %s: first line %q is not exactly the rebuild marker %q", e.Name(), first, rebuildMarker)
		}
		sum := sha256.Sum256(body)
		files = append(files, migration{
			version: v,
			name:    e.Name(),
			sum:     hex.EncodeToString(sum[:]),
			body:    string(body),
			rebuild: first == rebuildMarker,
		})
	}
	// fs.ReadDir sorts by name, and a zero-padded version sorts the
	// same by name as by number, but an unpadded one does not.
	slices.SortFunc(files, func(a, b migration) int { return cmp.Compare(a.version, b.version) })
	for i, m := range files {
		if m.version != i+1 {
			return nil, fmt.Errorf("read migrations: %s has version %d, want %d", m.name, m.version, i+1)
		}
	}
	return files, nil
}

// nearMarker reports whether line reads as the rebuild marker once one
// trailing carriage return and any trailing spaces and tabs are
// removed. Such a file was meant as a rebuild, but would run as an
// ordinary one with foreign keys on.
func nearMarker(line string) bool {
	return strings.TrimRight(strings.TrimSuffix(line, "\r"), " \t") == rebuildMarker
}

// checkLedger compares the ledger db holds with files, for a store at
// version current, no higher than len(files). A store at version 0
// that has no ledger table yet agrees.
func checkLedger(ctx context.Context, db *sql.DB, files []migration, current int) error {
	var tables int
	if err := db.QueryRowContext(ctx,
		"SELECT count(*) FROM sqlite_schema WHERE type = 'table' AND name = ?", ledgerTable,
	).Scan(&tables); err != nil {
		return fmt.Errorf("read ledger: %w", classifyErr(err))
	}
	if tables == 0 {
		if current == 0 {
			return nil
		}
		return fmt.Errorf("ledger table %s is missing at user_version %d", ledgerTable, current)
	}

	rows, err := db.QueryContext(ctx, "SELECT version, name, sha256 FROM "+ledgerTable+" ORDER BY version")
	if err != nil {
		return fmt.Errorf("read ledger: %w", classifyErr(err))
	}
	defer func() { _ = rows.Close() }()

	recorded := make(map[int]bool)
	for rows.Next() {
		var (
			v         int
			name, sum string
		)
		if err := rows.Scan(&v, &name, &sum); err != nil {
			return fmt.Errorf("read ledger: %w", err)
		}
		if v > current {
			return fmt.Errorf("ledger row for version %d is above user_version %d", v, current)
		}
		if v < 1 {
			return fmt.Errorf("ledger row for version %d has no migration file in this binary", v)
		}
		m := files[v-1]
		if name != m.name || sum != m.sum {
			return fmt.Errorf("ledger version %d records %s (sha256 %s), this binary has %s (sha256 %s)", v, name, sum, m.name, m.sum)
		}
		recorded[v] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("read ledger: %w", classifyErr(err))
	}
	for v := 1; v <= current; v++ {
		if !recorded[v] {
			return fmt.Errorf("ledger has no row for version %d at user_version %d", v, current)
		}
	}
	return nil
}

// applyMigration applies m in its own transaction, on the write pool
// or, for a rebuild, on a maintenance connection with foreign keys off.
func applyMigration(ctx context.Context, h *Handle, m migration, busyTimeout time.Duration) error {
	if m.rebuild {
		return applyRebuild(ctx, h.path, m, busyTimeout)
	}
	tx, err := h.write.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("apply %s: begin: %w", m.name, classifyErr(err))
	}
	if err := applyInTx(ctx, tx, h.path, m); err != nil {
		return rollbackMigration(tx, err)
	}
	return commitMigration(tx, h.path, m)
}

// applyRebuild applies m on a maintenance connection of its own with
// foreign-key enforcement off, which SQLite permits only outside a
// transaction (sqlite.org/pragma.html: enforcement "may only be enabled
// or disabled when there is no pending BEGIN or SAVEPOINT"). It refuses
// to commit when PRAGMA foreign_key_check reports any row, and closes
// the connection either way, so enforcement off never reaches a pool.
func applyRebuild(ctx context.Context, dbPath string, m migration, busyTimeout time.Duration) (err error) {
	db, err := openMaintenanceConn(dbPath, busyTimeout)
	if err != nil {
		return fmt.Errorf("apply %s: %w", m.name, err)
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("apply %s: connect: %w", m.name, classifyErr(err))
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("apply %s: foreign keys off: %w", m.name, classifyErr(err))
	}
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("apply %s: begin: %w", m.name, classifyErr(err))
	}
	if err := applyInTx(ctx, tx, dbPath, m); err != nil {
		return rollbackMigration(tx, err)
	}
	if err := checkForeignKeys(ctx, tx, m); err != nil {
		return rollbackMigration(tx, err)
	}
	return commitMigration(tx, dbPath, m)
}

// applyInTx runs m's statements inside tx, between SAVEPOINT and
// RELEASE (R-6c12), then sets user_version to m's version and records m
// in the ledger. The version is formatted into the statement because a
// PRAGMA takes no bound parameter. The first failing statement ends the
// file: a multi-statement exec stops there, and nothing after it runs.
//
// The SAVEPOINT catches a body that ends the runner's own transaction
// (a COMMIT, a ROLLBACK, or a COMMIT followed by a BEGIN): RELEASE then
// fails naming a savepoint that no longer exists, before user_version
// or the ledger row changes, so the caller's rollback finds nothing of
// this file committed beyond what the body's own COMMIT already
// persisted.
func applyInTx(ctx context.Context, tx *sql.Tx, dbPath string, m migration) error {
	var current int
	if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&current); err != nil {
		return fmt.Errorf("apply %s: read user_version: %w", m.name, classifyWriteErr(dbPath, err))
	}
	if current != m.version-1 {
		return fmt.Errorf("apply %s: store at version %d, want %d", m.name, current, m.version-1)
	}
	if _, err := tx.ExecContext(ctx, "SAVEPOINT "+runnerSavepoint); err != nil {
		return fmt.Errorf("apply %s: open savepoint: %w", m.name, classifyWriteErr(dbPath, err))
	}
	if _, err := tx.ExecContext(ctx, m.body); err != nil {
		return fmt.Errorf("apply %s: %w", m.name, classifyWriteErr(dbPath, err))
	}
	if _, err := tx.ExecContext(ctx, "RELEASE "+runnerSavepoint); err != nil {
		return fmt.Errorf("apply %s: release savepoint: the file's own transaction control ended the runner's transaction: %w", m.name, classifyWriteErr(dbPath, err))
	}
	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", m.version)); err != nil {
		return fmt.Errorf("apply %s: set user_version: %w", m.name, classifyWriteErr(dbPath, err))
	}
	if _, err := tx.ExecContext(ctx,
		"INSERT INTO "+ledgerTable+" (version, name, sha256) VALUES (?, ?, ?)", m.version, m.name, m.sum,
	); err != nil {
		return fmt.Errorf("apply %s: record in ledger: %w", m.name, classifyWriteErr(dbPath, err))
	}
	return nil
}

// checkForeignKeys refuses when PRAGMA foreign_key_check reports any
// row inside tx, naming the first row's table.
func checkForeignKeys(ctx context.Context, tx *sql.Tx, m migration) error {
	var table string
	err := tx.QueryRowContext(ctx, `SELECT "table" FROM pragma_foreign_key_check`).Scan(&table)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("apply %s: foreign key check: %w", m.name, classifyErr(err))
	}
	return fmt.Errorf("apply %s: foreign key check: table %s holds a row whose parent is missing", m.name, table)
}

// rollbackMigration rolls tx back and returns err, joined with the
// rollback's own error unless SQLite had already ended the transaction
// after one of the failures that end it.
func rollbackMigration(tx *sql.Tx, err error) error {
	rbErr := tx.Rollback()
	if rbErr == nil || errors.Is(rbErr, sql.ErrTxDone) || endsTransaction(err) {
		return err
	}
	return errors.Join(err, fmt.Errorf("roll back: %w", rbErr))
}

// commitMigration commits tx. A COMMIT that finds no open transaction
// after a clean apply means m ended the transaction itself, so every
// statement after its own COMMIT, the version write and the ledger row
// included, committed on its own: a hard error naming m. This is
// applyInTx's RELEASE check with the runner's savepoint's own name
// redeclared inside the body's ended-and-reopened transaction, which
// lets RELEASE succeed on the wrong savepoint (R-6c12); this branch is
// the only detection left for that shape.
func commitMigration(tx *sql.Tx, dbPath string, m migration) error {
	err := tx.Commit()
	if err == nil {
		return nil
	}
	var se *sqlite.Error
	if errors.As(err, &se) && se.Code()&0xff == sqliteErrorCode && strings.Contains(err.Error(), "no transaction is active") {
		return fmt.Errorf("apply %s: the file ends its own transaction, so its statements committed outside the runner's: %w", m.name, err)
	}
	return fmt.Errorf("apply %s: commit: %w", m.name, classifyWriteErr(dbPath, err))
}
