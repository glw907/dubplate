package local

import (
	"context"
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/engine/store"
)

// recoverJobs marks interrupted every queued or running record whose owner
// no longer holds its instance lease, in one write. It tries only the
// owners the records name: a record is inserted only after its owner
// holds its lease, so no lease it tries is one still being taken. It
// fails, naming the lease, when a lease can be neither locked nor
// found. A record this runner owns is judged by the jobs it tracks
// instead of by its lease, which it holds itself.
func (r *Runner) recoverJobs(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	return r.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		owners, err := liveOwners(ctx, tx)
		if err != nil {
			return err
		}
		for _, owner := range owners {
			if owner == r.lease.Instance() {
				if err := r.recoverOwn(ctx, tx); err != nil {
					return err
				}
				continue
			}
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

// recoverOwn marks interrupted every queued or running record this
// runner owns that no job of its runs: one whose job ended without the
// write that would have finished it. start tracks a job before its
// record is inserted and untracks it only after its last write, so a
// tracked job's record is never taken for a stranded one.
func (r *Runner) recoverOwn(ctx context.Context, tx *store.Tx) error {
	rows, err := tx.Query(ctx, "SELECT id FROM jobs WHERE owner_instance = ? AND state IN ('queued', 'running')", r.lease.Instance())
	if err != nil {
		return fmt.Errorf("read own jobs: %w", err)
	}
	var ids [][]byte
	for rows.Next() {
		var id []byte
		if err := rows.Scan(&id); err != nil {
			return errors.Join(fmt.Errorf("read own job: %w", err), rows.Close())
		}
		ids = append(ids, id)
	}
	if err := rows.Close(); err != nil {
		return fmt.Errorf("read own jobs: %w", err)
	}
	for _, id := range ids {
		if r.tracked(handleOf(id)) {
			continue
		}
		t := now()
		if _, err := tx.Exec(ctx, `UPDATE jobs SET state = 'interrupted', finished_at = ?, updated_at = ?
WHERE id = ? AND state IN ('queued', 'running')`, t, t, id); err != nil {
			return fmt.Errorf("mark job %s interrupted: %w", handleOf(id), err)
		}
	}
	return nil
}
