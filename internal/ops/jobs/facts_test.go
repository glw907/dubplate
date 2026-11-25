package jobs

import (
	"context"
	"database/sql"
	"maps"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/store"
)

// insertFinished inserts one terminal job record through h.
func insertFinished(t *testing.T, h *store.Handle, kind, state string, finished int64, progress any) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx,
			"INSERT INTO jobs (id, kind, at, state, owner_instance, actor, progress, updated_at, finished_at) VALUES (?, ?, 1, ?, ?, 'local:test', ?, ?, ?)",
			id.New(), kind, state, newInstance(), progress, finished, finished)
		return err
	})
	if err != nil {
		t.Fatalf("insert %s %s job: %v", kind, state, err)
	}
}

// TestFactsReader: an inspection running the reader reports the newest
// succeeded job per kind. Fixture state: two succeeded backups a day
// apart, a later failed backup, and one succeeded scan, so a reader
// that took the oldest success, or the newest job of any state, reads
// a different backup.
func TestFactsReader(t *testing.T) {
	dir, h := newStore(t)
	// 2026-09-24T12:00:00Z and a day later, in microseconds.
	const older, newest = 1790251200000000, 1790337600000000
	insertFinished(t, h, "backup", "succeeded", older, "an earlier run")
	insertFinished(t, h, "backup", "succeeded", newest, "prune skipped: a remote stamp is not older than this run's")
	insertFinished(t, h, "backup", "failed", newest+1, "a later failure")
	insertFinished(t, h, "scan", "succeeded", older, nil)

	var f Facts
	r := store.Inspect(context.Background(), store.InspectOptions{
		StateDir:    dir,
		BusyTimeout: testBusy,
		Readers:     []store.Reader{f.Reader()},
	})
	if r.State != store.StateCurrent || r.Facts == nil {
		t.Fatalf("inspect = %v (%s), want current with facts", r.State, r.Reason)
	}
	want := map[string]ops.JobSuccess{
		"backup": {Finished: time.Date(2026, 9, 25, 12, 0, 0, 0, time.UTC), Progress: "prune skipped: a remote stamp is not older than this run's"},
		"scan":   {Finished: time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)},
	}
	if !maps.EqualFunc(f.LastSuccess, want, func(a, b ops.JobSuccess) bool {
		return a.Finished.Equal(b.Finished) && a.Finished.Location() == time.UTC && a.Progress == b.Progress
	}) {
		t.Errorf("last success = %v, want %v", f.LastSuccess, want)
	}
}

// TestFactsReader_NoSuccess: a store whose jobs table holds no
// succeeded job reports an empty map.
func TestFactsReader_NoSuccess(t *testing.T) {
	dir, h := newStore(t)
	insertFinished(t, h, "backup", "failed", 1, "a failure")

	var f Facts
	store.Inspect(context.Background(), store.InspectOptions{StateDir: dir, BusyTimeout: testBusy, Readers: []store.Reader{f.Reader()}})
	if f.LastSuccess == nil || len(f.LastSuccess) != 0 {
		t.Errorf("last success = %#v, want an empty map", f.LastSuccess)
	}
}

// TestFactsReader_PredatesTable: a store older than the jobs table
// reads nothing and leaves the map nil.
func TestFactsReader_PredatesTable(t *testing.T) {
	var f Facts
	if err := f.Reader()(context.Background(), refusingQuerier{t}, jobsVersion-1); err != nil {
		t.Fatalf("reader: %v", err)
	}
	if f.LastSuccess != nil {
		t.Errorf("last success = %#v, want nil", f.LastSuccess)
	}
}

// refusingQuerier fails the test on any query, for a reader that must
// read nothing.
type refusingQuerier struct{ t *testing.T }

func (q refusingQuerier) QueryContext(_ context.Context, query string, _ ...any) (*sql.Rows, error) {
	q.t.Fatalf("queried %q", query)
	return nil, nil
}

func (q refusingQuerier) QueryRowContext(_ context.Context, query string, _ ...any) *sql.Row {
	q.t.Fatalf("queried %q", query)
	return nil
}
