package settings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/store"
)

// inspectSettings returns the facts an inspection of dir's store fills
// through the settings reader.
func inspectSettings(t *testing.T, dir string) Facts {
	t.Helper()
	var f Facts
	r := store.Inspect(context.Background(), store.InspectOptions{
		StateDir:    dir,
		BusyTimeout: testBusy,
		Readers:     []store.Reader{f.Reader()},
	})
	if r.State != store.StateCurrent || r.Facts == nil {
		t.Fatalf("inspect = %v (%s), want current with facts", r.State, r.Reason)
	}
	return f
}

// TestFactsReader: an inspection running the reader reads the stored
// rows in key order. Fixture state: the table holds two rows written
// out of key order, so an empty result passes no bug that never reads
// the table, and an unordered one shows.
func TestFactsReader(t *testing.T) {
	dir := newStateDir(t)
	writeFixture(t, openWrite(t, dir),
		fixtureRow{key: keyRetention, value: "720h", source: "admin", seed: "2160h"},
		fixtureRow{key: keyKeep, value: "21", source: "env_seed", seed: "21"},
	)

	got := inspectSettings(t, dir)
	at := time.UnixMicro(1).UTC()
	want := []Row{
		{Key: keyKeep, Value: "21", Source: "env_seed", SeedValue: "21", Actor: "fixture", UpdatedAt: at},
		{Key: keyRetention, Value: "720h", Source: "admin", SeedValue: "2160h", Actor: "fixture", UpdatedAt: at},
	}
	if got.Absent {
		t.Error("Absent = true, want false for a store holding the settings table")
	}
	if !slices.Equal(got.Rows, want) {
		t.Errorf("rows = %+v, want %+v", got.Rows, want)
	}
}

// TestFactsReader_EmptyTable: a store holding the settings table with
// no row reads no row and is not marked absent.
func TestFactsReader_EmptyTable(t *testing.T) {
	got := inspectSettings(t, newStateDir(t))
	if got.Absent || len(got.Rows) != 0 {
		t.Errorf("facts = %+v, want a present, empty table", got)
	}
}

// TestFactsReader_PredatesTable: a store older than the settings table
// is marked absent without a query, distinct from an empty table.
func TestFactsReader_PredatesTable(t *testing.T) {
	var f Facts
	if err := f.Reader()(context.Background(), refusingQuerier{t}, settingsVersion-1); err != nil {
		t.Fatalf("reader: %v", err)
	}
	if !f.Absent || f.Rows != nil {
		t.Errorf("facts = %+v, want absent with no rows", f)
	}
}

// TestScanRows_RespellsEnvSeed: a store below 0007 spells the seeded
// source env-seed, which reads as the env_seed its migration will
// write; the current CHECK refuses that spelling, so only a scanned
// row can carry it here.
func TestScanRows_RespellsEnvSeed(t *testing.T) {
	rs := &fakeRows{rows: [][]any{
		{keyKeep, "7", "env-seed", sql.NullString{String: "7", Valid: true}, "system", int64(1790251200000000)},
		{keyRetention, "720h", "admin", sql.NullString{}, "boss@example.test", int64(1)},
	}}
	got, err := scanRows(rs)
	if err != nil {
		t.Fatalf("scanRows: %v", err)
	}
	want := []Row{
		{Key: keyKeep, Value: "7", Source: "env_seed", SeedValue: "7", Actor: "system", UpdatedAt: time.UnixMicro(1790251200000000).UTC()},
		{Key: keyRetention, Value: "720h", Source: "admin", Actor: "boss@example.test", UpdatedAt: time.UnixMicro(1).UTC()},
	}
	if !slices.Equal(got, want) {
		t.Errorf("rows = %+v, want %+v", got, want)
	}
}

// fakeRows yields rows, each a key, value, source, seed, actor, and
// microsecond stamp, in the order scanRows scans them.
type fakeRows struct {
	rows [][]any
	next int
}

func (r *fakeRows) Next() bool {
	r.next++
	return r.next <= len(r.rows)
}

func (r *fakeRows) Scan(dest ...any) error {
	row := r.rows[r.next-1]
	if len(dest) != len(row) {
		return errors.New("scan: column count")
	}
	for i, d := range dest {
		ok := false
		switch d := d.(type) {
		case *string:
			*d, ok = row[i].(string)
		case *sql.NullString:
			*d, ok = row[i].(sql.NullString)
		case *int64:
			*d, ok = row[i].(int64)
		}
		if !ok {
			return fmt.Errorf("scan: column %d is %T, want %T", i, row[i], d)
		}
	}
	return nil
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
