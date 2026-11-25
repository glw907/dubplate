package intake

import (
	"context"
	"database/sql"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/store"
)

// TestFactsReader: an inspection running the reader lists every
// collection root in byte order. Fixture state: two roots written out
// of order, so an empty or unordered read shows.
func TestFactsReader(t *testing.T) {
	dir := t.TempDir()
	h, err := store.OpenServe(t.Context(), dir, store.ServeOptions{
		BusyTimeout:   5 * time.Second,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("open store: %v", err)
	}
	admin := insertUser(t, h, userSpec{admin: true, createdAt: 1})
	err = h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		for _, root := range []string{"zeta", "alpha"} {
			if _, err := tx.Exec(ctx, `INSERT INTO collections (id, name, root_path, user_id, visibility, writer)
				VALUES (?, ?, ?, ?, 'group', 'server')`, id.New(), root, []byte(root), admin); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("seed roots: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close store: %v", err)
	}

	var f Facts
	r := store.Inspect(context.Background(), store.InspectOptions{
		StateDir:    dir,
		BusyTimeout: 5 * time.Second,
		Readers:     []store.Reader{f.Reader()},
	})
	if r.State != store.StateCurrent || r.Facts == nil {
		t.Fatalf("inspect = %v (%s), want current with facts", r.State, r.Reason)
	}
	want := [][]byte{[]byte("alpha"), []byte("zeta")}
	if !slices.EqualFunc(f.roots, want, slices.Equal) {
		t.Errorf("roots = %q, want %q", f.roots, want)
	}
}

// TestFactsReader_PredatesTable: a store older than the collections
// table reads no root and runs no query.
func TestFactsReader_PredatesTable(t *testing.T) {
	var f Facts
	if err := f.Reader()(context.Background(), refusingQuerier{t}, collectionsVersion-1); err != nil {
		t.Fatalf("reader: %v", err)
	}
	if f.roots != nil {
		t.Errorf("roots = %q, want none", f.roots)
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
