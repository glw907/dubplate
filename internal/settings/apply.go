package settings

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// errNoGeneration reports a store whose settings_generation table has
// lost its one row, so no cache could ever see a change.
var errNoGeneration = errors.New("settings_generation holds no row")

// apply runs the seed rule for reg's settings against h's store and
// environ, and writes the result in one transaction: each changed row,
// an audit row per change, and one generation bump when anything
// changed. When the rule refuses, it writes nothing and returns the
// joined refusal. now stamps every row it writes, and log receives a
// line per change once the transaction commits.
//
// The rows are read inside the write's own transaction, so a write
// another process commits between the read and the write cannot be
// overwritten. Every row it writes carries the source 'env-seed'.
func apply(ctx context.Context, h *store.Handle, reg *config.Registry, environ []string, now time.Time, log *slog.Logger) error {
	var ev Evaluation
	err := h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		rs, err := tx.Query(ctx, selectRows)
		if err != nil {
			return fmt.Errorf("read settings: %w", err)
		}
		rows, err := scanRows(rs)
		if closeErr := rs.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			return fmt.Errorf("read settings: %w", err)
		}

		ev = evaluate(reg, environ, rows)
		if ev.Err != nil {
			return ev.Err
		}
		return writeChanges(ctx, tx, ev.Outcomes, now)
	})
	if err != nil {
		return err
	}

	for _, o := range ev.Outcomes {
		if o.Change != Unchanged {
			log.Info("setting "+o.Change.String(), "key", o.Key, "outcome", o.Report)
		}
	}
	return nil
}

// writeChanges writes every changed outcome's row and audit row
// through tx, then bumps the settings generation once if anything
// changed.
func writeChanges(ctx context.Context, tx *store.Tx, outcomes []Outcome, now time.Time) error {
	at := now.UnixMicro()
	changed := false
	for _, o := range outcomes {
		// oldValue and newValue stay nil, stored as NULL, for a seed
		// from no row and a removal to none.
		var oldValue, newValue any
		var err error
		switch o.Change {
		case Unchanged:
			continue
		case Seed:
			newValue = o.Env
			_, err = tx.Exec(ctx, `INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
				VALUES (?, ?, 'env-seed', ?, 'system', ?)`, o.Key, o.Env, o.Env, at)
		case Reseed:
			oldValue, newValue = o.Stored.Value, o.Env
			_, err = tx.Exec(ctx, `UPDATE settings SET value = ?, seed_value = ?, actor = 'system', updated_at = ?
				WHERE key = ? AND source = 'env-seed'`, o.Env, o.Env, at, o.Key)
		case Remove:
			oldValue = o.Stored.Value
			_, err = tx.Exec(ctx, `DELETE FROM settings WHERE key = ? AND source = 'env-seed'`, o.Key)
		}
		if err != nil {
			return fmt.Errorf("%s: write %s: %w", o.Key, o.Change, err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO settings_audit (id, at, kind, actor, channel, key, old_value, new_value)
			VALUES (?, ?, ?, 'system', 'env-seed', ?, ?, ?)`,
			store.NewID(), at, o.Change.String(), o.Key, oldValue, newValue); err != nil {
			return fmt.Errorf("%s: audit %s: %w", o.Key, o.Change, err)
		}
		changed = true
	}
	if !changed {
		return nil
	}

	res, err := tx.Exec(ctx, `UPDATE settings_generation SET generation = generation + 1 WHERE singleton = 1`)
	if err != nil {
		return fmt.Errorf("bump settings_generation: %w", err)
	}
	if n, err := res.RowsAffected(); err != nil || n != 1 {
		return fmt.Errorf("bump settings_generation: %w", cmp.Or(err, errNoGeneration))
	}
	return nil
}
