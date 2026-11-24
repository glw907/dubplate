package store

import (
	"context"
	"database/sql"
	"fmt"
)

// An amendment reconciles a migration file that was edited after a
// store applied it. A ledger row at version whose hash is formerSum
// records the file's earlier text; fix brings that store's schema to
// what the current text creates, and the ledger row then records the
// current hash.
type amendment struct {
	version   int
	formerSum string
	fix       string
}

// amendments are the edits a store's ledger may still record. The box's
// stores were created before 0003_catalog.sql renamed
// users.streamer_username to username in place.
var amendments = []amendment{{
	version:   3,
	formerSum: "f962133f7366b5e941e2346fc1d2e3cd4defa743ec02902241408e9173a089c5",
	fix:       "ALTER TABLE users RENAME COLUMN streamer_username TO username",
}}

// amendable reports whether a ledger row at version recording sum is
// one an amendment fixes.
func amendable(version int, sum string) bool {
	for _, a := range amendments {
		if a.version == version && a.formerSum == sum {
			return true
		}
	}
	return false
}

// A repair is an amendment due on one store: fix runs, and the ledger
// row at version then records sum, the current file's hash.
type repair struct {
	version int
	sum     string
	fix     string
}

// dueRepairs returns a repair for every amendment whose former hash the
// ledger of a store at version current records, each taking its
// current hash from files. checkLedger has already accepted the ledger.
// It reads and never writes.
func dueRepairs(ctx context.Context, q *sql.DB, files []migration, current int) ([]repair, error) {
	var due []repair
	for _, a := range amendments {
		if a.version > current {
			continue
		}
		var n int
		if err := q.QueryRowContext(ctx,
			"SELECT count(*) FROM "+ledgerTable+" WHERE version = ? AND sha256 = ?", a.version, a.formerSum,
		).Scan(&n); err != nil {
			return nil, fmt.Errorf("amend version %d: read ledger: %w", a.version, classifyErr(err))
		}
		if n > 0 {
			due = append(due, repair{version: a.version, sum: files[a.version-1].sum, fix: a.fix})
		}
	}
	return due, nil
}

// applyRepairs runs each repair inside tx, the transaction of the first
// migration the store takes. A store never commits an amended ledger at
// its old version, so a retry after a failed migration copies the store
// as it was, and the pre-migrate copy stays one the earlier binary
// opens.
func applyRepairs(ctx context.Context, tx *sql.Tx, dbPath string, due []repair) error {
	for _, r := range due {
		if _, err := tx.ExecContext(ctx,
			"UPDATE "+ledgerTable+" SET sha256 = ? WHERE version = ?", r.sum, r.version,
		); err != nil {
			return fmt.Errorf("amend version %d: ledger: %w", r.version, classifyWriteErr(dbPath, err))
		}
		if _, err := tx.ExecContext(ctx, r.fix); err != nil {
			return fmt.Errorf("amend version %d: %w", r.version, classifyWriteErr(dbPath, err))
		}
	}
	return nil
}
