package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	sqlite "modernc.org/sqlite"
)

// sqliteDriverName is the name modernc.org/sqlite registers with
// database/sql (sqlite.go's sql.Register call).
const sqliteDriverName = "sqlite"

// journalSizeLimit bounds the WAL file after a checkpoint. SQLite's
// wal.html section 6: a checkpoint "does not normally truncate the WAL
// file (unless the journal_size_limit pragma is set)." 64 MiB.
const journalSizeLimit = 64 * 1024 * 1024

// sqliteBusyCode and sqliteReadOnlyCode are SQLite's primary result
// codes for SQLITE_BUSY and SQLITE_READONLY. classifyErr masks a
// *sqlite.Error's extended code against these, since the driver enables
// extended result codes on every connection (newConn's
// extendedResultCodes(true) call).
const (
	sqliteBusyCode     = 5
	sqliteReadOnlyCode = 8
)

// errBusy wraps a failure after which SQLite's write lock stayed held
// past the busy timeout.
var errBusy = errors.New("store busy")

// errReadOnly wraps a write that reached a connection SQLite itself
// refuses to write to.
var errReadOnly = errors.New("store read-only")

// A Handle is an open store: the write pool's one connection, the read
// pool, and the store lock its mode's open holds until [Handle.Close].
// It is safe for concurrent use by multiple goroutines.
type Handle struct {
	write *sql.DB
	read  *sql.DB

	// path is the database file, named in errors a caller reports.
	path string

	// version is the binary's highest migration. The write method
	// refuses a store whose user_version differs from it.
	version int

	// lock is the store lock a mode's open holds for the handle's
	// life, or nil for a handle opened on a bare file.
	lock *storeLock

	// leasesDir is the leaf's leases/ directory, where every lease
	// taken through the handle lives.
	leasesDir string
}

// openHandle opens dbPath's write pool (capped at one connection) and
// read pool. busyTimeout is the caller's threshold: the store reads no
// environment. Every per-connection setting rides the DSN, because
// database/sql opens connections at will; nothing here issues a PRAGMA
// statement after open.
func openHandle(dbPath string, busyTimeout time.Duration) (*Handle, error) {
	write, err := sql.Open(sqliteDriverName, writeDSN(dbPath, busyTimeout))
	if err != nil {
		return nil, fmt.Errorf("open write pool: %w", err)
	}
	write.SetMaxOpenConns(1)

	read, err := sql.Open(sqliteDriverName, readDSN(dbPath, busyTimeout))
	if err != nil {
		_ = write.Close()
		return nil, fmt.Errorf("open read pool: %w", err)
	}

	return &Handle{write: write, read: read, path: dbPath}, nil
}

// Close closes h's write and read pools, then releases the store lock,
// joining every error rather than stopping at the first. A lease taken
// through h is not released: its taker releases it.
func (h *Handle) Close() error {
	err := errors.Join(h.write.Close(), h.read.Close())
	if h.lock != nil {
		err = errors.Join(err, h.lock.release())
	}
	return err
}

// QueryRow runs query on h's read pool, which SQLite itself refuses to
// write through, and returns at most one row. A read on the read pool
// sees the last committed state but holds no lock against a write that
// follows it, so it serves reports; a read that a write decides on
// runs inside that write's [Handle.Update] callback instead.
func (h *Handle) QueryRow(ctx context.Context, query string, args ...any) *sql.Row {
	return h.read.QueryRowContext(ctx, query, args...)
}

// openMaintenanceConn opens a maintenance connection at dbPath for one
// job (a migration run, a snapshot, a retention sweep). The caller
// closes it when the job ends; the pool is capped at one connection,
// same as the write pool.
func openMaintenanceConn(dbPath string, busyTimeout time.Duration) (*sql.DB, error) {
	db, err := sql.Open(sqliteDriverName, maintenanceDSN(dbPath, busyTimeout))
	if err != nil {
		return nil, fmt.Errorf("open maintenance connection: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

// writeDSN builds the write pool's DSN: read-write, WAL, foreign keys
// on, full synchronous, an immediate begin mode so newTx's "begin
// immediate" takes the write lock up front, and the bounded WAL size.
func writeDSN(dbPath string, busyTimeout time.Duration) string {
	return buildDSN(dbPath,
		"mode=rw",
		busyTimeoutParam(busyTimeout),
		"_foreign_keys=1",
		"_journal_mode=WAL",
		"_synchronous=FULL",
		"_txlock=immediate",
		journalSizeLimitParam(),
	)
}

// readDSN builds the read pool's DSN: writeDSN's keys without
// _txlock, so a ReadOnly transaction takes no write lock, plus
// _query_only=1 so SQLite itself refuses a write.
func readDSN(dbPath string, busyTimeout time.Duration) string {
	return buildDSN(dbPath,
		"mode=rw",
		busyTimeoutParam(busyTimeout),
		"_foreign_keys=1",
		"_journal_mode=WAL",
		"_synchronous=FULL",
		journalSizeLimitParam(),
		"_query_only=1",
	)
}

// maintenanceDSN builds the maintenance class's DSN: an immediate
// begin mode, like the write pool, so a maintenance transaction takes
// the write lock at BEGIN rather than deferring it to the first write.
// It omits _journal_mode: journal mode persists in the database file,
// so this class neither sets nor converts it.
func maintenanceDSN(dbPath string, busyTimeout time.Duration) string {
	return buildDSN(dbPath,
		"mode=rw",
		busyTimeoutParam(busyTimeout),
		"_foreign_keys=1",
		"_synchronous=FULL",
		"_txlock=immediate",
		journalSizeLimitParam(),
	)
}

// inspectDSN builds the inspect class's DSN: read-only at the SQLite
// open-mode level, plus _query_only=1 so a caller that runs a
// migration or a write against it fails at the driver rather than
// through a silent no-op.
func inspectDSN(dbPath string, busyTimeout time.Duration) string {
	return buildDSN(dbPath, "mode=ro", busyTimeoutParam(busyTimeout), "_query_only=1")
}

// buildDSN forms a file: URI DSN from path, followed by params joined
// with '&'. It is the one builder every DSN goes through: url.URL's
// Path field escapes '#', '?', and '%' so a path carrying any of them
// opens the exact file it names, rather than SQLite parsing a
// concatenated string as a URI with a truncated path. The driver
// applies busy_timeout, the _pragma values, and query_only in a fixed
// order regardless of the DSN's own key order, so params need not
// appear in the driver's application order.
func buildDSN(path string, params ...string) string {
	base := (&url.URL{Scheme: "file", Path: path}).String()
	return base + "?" + strings.Join(params, "&")
}

// busyTimeoutParam formats the _busy_timeout shorthand key, which the
// driver documents as taking an integer count of milliseconds.
func busyTimeoutParam(busyTimeout time.Duration) string {
	return fmt.Sprintf("_busy_timeout=%d", busyTimeout.Milliseconds())
}

// journalSizeLimitParam formats journalSizeLimit as a _pragma value:
// journal_size_limit has no shorthand key, so it rides the one
// unvalidated escape hatch the driver documents.
func journalSizeLimitParam() string {
	return fmt.Sprintf("_pragma=journal_size_limit(%d)", journalSizeLimit)
}

// classifyErr wraps err in errBusy or errReadOnly when it carries a
// *sqlite.Error whose primary code, masking the extended code every
// connection reports (newConn's extendedResultCodes(true)), is
// SQLITE_BUSY or SQLITE_READONLY. A caller then matches with errors.Is
// regardless of which extended variant SQLite returned (a deferred
// upgrade's SQLITE_BUSY_SNAPSHOT included). Any other error, including
// one that carries no *sqlite.Error at all, passes through unchanged.
func classifyErr(err error) error {
	if err == nil {
		return nil
	}
	var se *sqlite.Error
	if !errors.As(err, &se) {
		return err
	}
	switch se.Code() & 0xff {
	case sqliteBusyCode:
		return fmt.Errorf("%w: %v", errBusy, err)
	case sqliteReadOnlyCode:
		return fmt.Errorf("%w: %v", errReadOnly, err)
	default:
		return err
	}
}
