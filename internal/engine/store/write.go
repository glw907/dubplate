package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sqlite "modernc.org/sqlite"
)

// SQLite's primary result codes for the four failures after which
// SQLite may roll back the whole transaction on its own
// (sqlite.org/lang_transaction.html section 3): SQLITE_NOMEM,
// SQLITE_INTERRUPT, SQLITE_IOERR, and SQLITE_FULL.
const (
	sqliteNoMemCode     = 7
	sqliteInterruptCode = 9
	sqliteIOErrCode     = 10
	sqliteFullCode      = 13
)

// errNoDeadline is returned by a write whose context carries no
// deadline, before it waits on anything.
var errNoDeadline = errors.New("store write needs a context deadline")

// errNestedWrite is returned by a write issued from inside another
// write's callback on the callback's context.
var errNestedWrite = errors.New("store write nested inside another write")

// errVersionMismatch is returned by a write that finds the store's
// user_version differing from the binary's highest migration.
var errVersionMismatch = errors.New("store version differs from the binary's")

// errDiskFull is returned by a write that SQLite refused with
// SQLITE_FULL.
var errDiskFull = errors.New("store disk full")

// writeMark is the context key a write sets on the context it hands
// its callback, so a nested write on that context fails at once rather
// than waiting for the one connection its own caller holds.
type writeMark struct{}

// Update runs fn inside one IMMEDIATE transaction on the write pool
// and commits only when fn returns nil and no statement fn issued
// through tx failed. Otherwise it rolls back and returns fn's error,
// or, when fn returned nil, the first statement error tx recorded.
// SQLite lets a statement fail while its transaction continues, so a
// callback that swallows a statement's error still never commits.
//
// ctx must carry a deadline, which bounds the wait for the write
// pool's one connection inside this process. The wait on another
// process's write lock is bounded by the busy timeout alone: a
// deadline shorter than the busy timeout does not shorten it.
//
// fn receives a context derived from ctx and marked, so an Update
// called on that context from inside fn fails at once. An Update
// called from inside fn on ctx itself, or on any
// context not derived from fn's, carries no mark and waits for the
// connection fn's own transaction holds until that context ends.
//
// The first statement reads user_version, and Update refuses without
// calling fn when it differs from the binary's version.
func (h *Handle) Update(ctx context.Context, fn func(ctx context.Context, tx *Tx) error) error {
	if ctx.Value(writeMark{}) != nil {
		return errNestedWrite
	}
	if _, ok := ctx.Deadline(); !ok {
		return errNoDeadline
	}

	sqlTx, err := h.write.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin write: %w", classifyErr(err))
	}
	tx := &Tx{tx: sqlTx, path: h.path}

	var got int
	if err := tx.QueryRow(ctx, "PRAGMA user_version").Scan(&got); err != nil {
		return tx.rollback(fmt.Errorf("read store version: %w", err))
	}
	if got != h.version {
		return tx.rollback(fmt.Errorf("%w: store at %d, binary at %d", errVersionMismatch, got, h.version))
	}

	if err := fn(context.WithValue(ctx, writeMark{}, true), tx); err != nil {
		return tx.rollback(err)
	}
	if tx.err != nil {
		return tx.rollback(tx.err)
	}
	if err := sqlTx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", classifyWriteErr(h.path, err))
	}
	return nil
}

// A Tx is the transaction an [Handle.Update] callback writes through.
// It is valid only until the callback returns. It records the first
// statement error, and once one is recorded every
// later statement returns it without reaching SQLite: a transaction
// SQLite rolled back on its own leaves the connection in autocommit
// mode, where each later statement would commit by itself.
type Tx struct {
	tx   *sql.Tx
	path string
	err  error
}

// record remembers err as tx's first statement error and returns the
// error the statement reports. sql.ErrNoRows is returned unrecorded:
// a query that matched no row leaves the transaction intact.
func (tx *Tx) record(err error) error {
	if err == nil || errors.Is(err, sql.ErrNoRows) {
		return err
	}
	if tx.err == nil {
		tx.err = classifyWriteErr(tx.path, err)
	}
	return tx.err
}

// rollback rolls tx back and returns err. The rollback's own error is
// joined to err unless the transaction had already ended: the sql
// package rolls back when the context ends, and SQLite may roll back on
// its own after one of the four failures lang_transaction.html section 3
// names, in which case its "no transaction is active" answer is not a
// second failure.
func (tx *Tx) rollback(err error) error {
	rbErr := tx.tx.Rollback()
	if rbErr == nil || errors.Is(rbErr, sql.ErrTxDone) || endsTransaction(tx.err) {
		return err
	}
	return errors.Join(err, fmt.Errorf("roll back: %w", rbErr))
}

// Exec runs a statement that returns no rows.
func (tx *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if tx.err != nil {
		return nil, tx.err
	}
	res, err := tx.tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, tx.record(err)
	}
	return res, nil
}

// QueryRow runs a query expected to return at most one row. Its error,
// the query's or the scan's, surfaces from the returned row's
// [Row.Scan].
func (tx *Tx) QueryRow(ctx context.Context, query string, args ...any) *Row {
	if tx.err != nil {
		return &Row{tx: tx}
	}
	return &Row{tx: tx, row: tx.tx.QueryRowContext(ctx, query, args...)}
}

// Query runs a query that returns rows. The caller closes the returned
// rows.
func (tx *Tx) Query(ctx context.Context, query string, args ...any) (*Rows, error) {
	if tx.err != nil {
		return nil, tx.err
	}
	rows, err := tx.tx.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, tx.record(err)
	}
	return &Rows{tx: tx, rows: rows}, nil
}

// A Row is the result of [Tx.QueryRow].
type Row struct {
	tx  *Tx
	row *sql.Row
}

// Scan copies the row's columns into dest, as [sql.Row.Scan] does, and
// records any error but [sql.ErrNoRows], which it returns.
func (r *Row) Scan(dest ...any) error {
	if r.tx.err != nil {
		return r.tx.err
	}
	return r.tx.record(r.row.Scan(dest...))
}

// A Rows is the result of [Tx.Query].
type Rows struct {
	tx   *Tx
	rows *sql.Rows
}

// Next advances to the next row, as [sql.Rows.Next] does. When it
// returns false because iteration failed, the failure is recorded, so a
// caller that never checks an error still cannot commit.
func (r *Rows) Next() bool {
	if r.tx.err != nil {
		return false
	}
	if r.rows.Next() {
		return true
	}
	_ = r.tx.record(r.rows.Err())
	return false
}

// Scan copies the current row's columns into dest and records any
// error.
func (r *Rows) Scan(dest ...any) error {
	if r.tx.err != nil {
		return r.tx.err
	}
	return r.tx.record(r.rows.Scan(dest...))
}

// Close closes the rows and returns the first error the transaction
// recorded, which includes one that ended iteration, so it serves as
// both [sql.Rows.Close] and [sql.Rows.Err].
func (r *Rows) Close() error {
	_ = r.tx.record(r.rows.Close())
	return r.tx.err
}

// classifyWriteErr wraps a SQLITE_FULL failure in errDiskFull, naming
// the store at path and keeping SQLite's error in the chain, and
// otherwise defers to classifyErr.
func classifyWriteErr(path string, err error) error {
	var se *sqlite.Error
	if errors.As(err, &se) && se.Code()&0xff == sqliteFullCode {
		return fmt.Errorf("%w: %s: %w", errDiskFull, path, err)
	}
	return classifyErr(err)
}

// endsTransaction reports whether err carries one of the four SQLite
// failures after which SQLite may already have rolled the transaction
// back.
func endsTransaction(err error) bool {
	var se *sqlite.Error
	if !errors.As(err, &se) {
		return false
	}
	switch se.Code() & 0xff {
	case sqliteNoMemCode, sqliteInterruptCode, sqliteIOErrCode, sqliteFullCode:
		return true
	default:
		return false
	}
}
