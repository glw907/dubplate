package member

import (
	"context"
	"database/sql"
	"testing"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/store"
)

// standingFixture is a live admin, a live non-admin, a tombstoned admin,
// and a tombstoned non-admin, so a count over every row reads 4 and 2
// and a lookup that ignores tombstones reads the tombstoned admin as an
// admin.
var standingFixture = []userRow{
	{id: identity.UserID{1}, address: "boss@example.test", displayName: "Boss", username: "boss", admin: true, createdAt: 1},
	{id: identity.UserID{2}, address: "plain@example.test", displayName: "Plain", username: "plain", createdAt: 2},
	{id: identity.UserID{3}, address: "gone-admin@example.test", displayName: "Gone", username: "goneadmin", admin: true, tombstoned: true, createdAt: 3},
	{id: identity.UserID{4}, address: "gone-plain@example.test", displayName: "Gone", username: "goneplain", tombstoned: true, createdAt: 4},
}

// inspectRoster creates a current store holding rows, closes it, and
// returns the facts an inspection running read fills.
func inspectRoster(t *testing.T, rows []userRow, read func(*Facts) store.Reader) Facts {
	t.Helper()
	dir := t.TempDir()
	h := newUsersStoreAt(t, dir)
	if len(rows) > 0 {
		insertUsers(t, h, rows...)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close serve handle: %v", err)
	}
	var f Facts
	r := store.Inspect(context.Background(), store.InspectOptions{
		StateDir:    dir,
		BusyTimeout: usersBusy,
		Readers:     []store.Reader{read(&f)},
	})
	if r.State != store.StateCurrent || r.Facts == nil {
		t.Fatalf("inspect = %v (%s), want current with facts", r.State, r.Reason)
	}
	return f
}

// TestFactsReader: the reader counts the live members and the live
// admins among them, and reads the standing of the address the
// environment names as the admin email, folded to the roster's case.
func TestFactsReader(t *testing.T) {
	tests := []struct {
		name    string
		rows    []userRow
		address string
		want    Facts
	}{
		{name: "a live admin", rows: standingFixture, address: "boss@example.test", want: Facts{2, 1, standingAdmin}},
		{name: "a live admin, mixed case", rows: standingFixture, address: "Boss@Example.TEST", want: Facts{2, 1, standingAdmin}},
		{name: "a live non-admin", rows: standingFixture, address: "plain@example.test", want: Facts{2, 1, standingMember}},
		{name: "a tombstoned admin", rows: standingFixture, address: "gone-admin@example.test", want: Facts{2, 1, standingRemoved}},
		{name: "a tombstoned non-admin", rows: standingFixture, address: "gone-plain@example.test", want: Facts{2, 1, standingRemoved}},
		{name: "an address no row holds", rows: standingFixture, address: "nobody@example.test", want: Facts{2, 1, standingNone}},
		{name: "no address named", rows: standingFixture, want: Facts{2, 1, standingNotRead}},
		{name: "an unparseable address", rows: standingFixture, address: "not an address", want: Facts{2, 1, standingNotRead}},
		{name: "an empty table", address: "boss@example.test", want: Facts{0, 0, standingNone}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var environ []string
			if tt.address != "" {
				environ = []string{adminEmailKey + "=" + tt.address}
			}
			got := inspectRoster(t, tt.rows, func(f *Facts) store.Reader { return f.Reader(environ) })
			if got != tt.want {
				t.Errorf("facts = %+v, want %+v", got, tt.want)
			}
		})
	}
}

// TestFactsReader_Version: a store at the users table's own version
// reads the roster, and one below it reads nothing.
func TestFactsReader_Version(t *testing.T) {
	environ := []string{adminEmailKey + "=boss@example.test"}
	t.Run("at the users table's version", func(t *testing.T) {
		got := inspectRoster(t, standingFixture, func(f *Facts) store.Reader {
			read := f.Reader(environ)
			return func(ctx context.Context, q store.Querier, _ int) error { return read(ctx, q, usersVersion) }
		})
		if want := (Facts{2, 1, standingAdmin}); got != want {
			t.Errorf("facts = %+v, want %+v", got, want)
		}
	})
	t.Run("below the users table", func(t *testing.T) {
		var f Facts
		if err := f.Reader(environ)(context.Background(), refusingQuerier{t}, usersVersion-1); err != nil {
			t.Fatalf("reader: %v", err)
		}
		if f != (Facts{}) {
			t.Errorf("facts = %+v, want none read", f)
		}
	})
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
