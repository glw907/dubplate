package store

import (
	"context"
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

// amend applies, in one transaction, every amendment whose former hash
// the ledger of h, a store at version current, records, and records
// each file's current hash from files in its place. checkLedger has
// already accepted the ledger. A store whose ledger records no former
// hash takes no write.
func amend(ctx context.Context, h *Handle, files []migration, current int) error {
	var due []amendment
	for _, a := range amendments {
		if a.version > current {
			continue
		}
		var n int
		if err := h.write.QueryRowContext(ctx,
			"SELECT count(*) FROM "+ledgerTable+" WHERE version = ? AND sha256 = ?", a.version, a.formerSum,
		).Scan(&n); err != nil {
			return fmt.Errorf("amend version %d: read ledger: %w", a.version, classifyErr(err))
		}
		if n > 0 {
			due = append(due, a)
		}
	}
	if len(due) == 0 {
		return nil
	}

	tx, err := h.write.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("amend: begin: %w", classifyErr(err))
	}
	for _, a := range due {
		if _, err := tx.ExecContext(ctx,
			"UPDATE "+ledgerTable+" SET sha256 = ? WHERE version = ?", files[a.version-1].sum, a.version,
		); err != nil {
			return rollbackMigration(tx, fmt.Errorf("amend version %d: ledger: %w", a.version, classifyWriteErr(h.path, err)))
		}
		if _, err := tx.ExecContext(ctx, a.fix); err != nil {
			return rollbackMigration(tx, fmt.Errorf("amend version %d: %w", a.version, classifyWriteErr(h.path, err)))
		}
	}
	if err := tx.Commit(); err != nil {
		return rollbackMigration(tx, fmt.Errorf("amend: commit: %w", classifyWriteErr(h.path, err)))
	}
	return nil
}
