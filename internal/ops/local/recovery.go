package local

import (
	"context"
	"fmt"

	"github.com/glw907/dubplate/internal/engine/store"
)

// recoverJobs marks interrupted every queued or running record whose owner
// no longer holds its instance lease, in one write. It tries only the
// owners the records name: a record is inserted only after its owner
// holds its lease, so no lease it tries is one still being taken. It
// fails, naming the lease, when a lease can be neither locked nor
// found.
func (r *Runner) recoverJobs(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	return r.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		owners, err := liveOwners(ctx, tx)
		if err != nil {
			return err
		}
		for _, owner := range owners {
			state, err := r.store.TryLease(owner)
			if err != nil {
				return fmt.Errorf("recover jobs: %w", err)
			}
			if state == store.LeaseHeld {
				continue
			}
			t := now()
			if _, err := tx.Exec(ctx, `UPDATE jobs SET state = 'interrupted', finished_at = ?, updated_at = ?
WHERE owner_instance = ? AND state IN ('queued', 'running')`, t, t, owner); err != nil {
				return fmt.Errorf("mark jobs of %s interrupted: %w", owner, err)
			}
		}
		return nil
	})
}

// liveOwners returns each distinct owner of a queued or running record.
func liveOwners(ctx context.Context, tx *store.Tx) ([]string, error) {
	rows, err := tx.Query(ctx, "SELECT DISTINCT owner_instance FROM jobs WHERE state IN ('queued', 'running')")
	if err != nil {
		return nil, fmt.Errorf("read job owners: %w", err)
	}
	var owners []string
	for rows.Next() {
		var o string
		if err := rows.Scan(&o); err != nil {
			return nil, fmt.Errorf("read job owner: %w", err)
		}
		owners = append(owners, o)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("read job owners: %w", err)
	}
	return owners, nil
}
