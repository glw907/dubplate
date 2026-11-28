package auth

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/store"
)

// newUsersStore creates a current store under t.TempDir() and returns
// its serve-mode handle, closed at cleanup.
func newUsersStore(t *testing.T) *store.Handle {
	t.Helper()
	h, err := store.OpenServe(context.Background(), t.TempDir(), store.ServeOptions{
		BusyTimeout:   5 * time.Second,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// writeCtx returns a context carrying the deadline every store write
// needs.
func writeCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// userRow is a users row a test writes directly, before the code under
// test runs, so a fixture can hold a tombstoned row or a fixed id.
type userRow struct {
	id                             identity.UserID
	address, displayName, username string
	admin, tombstoned              bool
	createdAt                      int64
}

// insertUsers writes rows into h's users table in one transaction.
func insertUsers(t *testing.T, h *store.Handle, rows ...userRow) {
	t.Helper()
	err := h.Update(writeCtx(t), func(ctx context.Context, tx *store.Tx) error {
		for _, r := range rows {
			var tombstonedAt any
			if r.tombstoned {
				tombstonedAt = r.createdAt + 1
			}
			if _, err := tx.Exec(ctx, `INSERT INTO users (id, address, display_name, username, admin, tombstoned_at, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
				r.id[:], r.address, r.displayName, r.username, r.admin, tombstonedAt, r.createdAt, r.createdAt); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("insert fixture users: %v", err)
	}
}

// dumpUsers renders every users row, so a refused write can be shown to
// have changed none.
func dumpUsers(t *testing.T, h *store.Handle) string {
	t.Helper()
	rows, err := h.Query(t.Context(), `SELECT hex(id), address, display_name, username, admin, coalesce(tombstoned_at, -1), created_at, updated_at FROM users ORDER BY address`)
	if err != nil {
		t.Fatalf("dump users: %v", err)
	}
	defer func() { _ = rows.Close() }()
	var b strings.Builder
	for rows.Next() {
		var id, address, displayName, username string
		var admin, tombstonedAt, createdAt, updatedAt int64
		if err := rows.Scan(&id, &address, &displayName, &username, &admin, &tombstonedAt, &createdAt, &updatedAt); err != nil {
			t.Fatalf("scan users: %v", err)
		}
		fmt.Fprintf(&b, "%s %s %q %s %d %d %d %d\n", id, address, displayName, username, admin, tombstonedAt, createdAt, updatedAt)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("dump users: %v", err)
	}
	return b.String()
}

// Fixed ids for the users-table fixtures, so a test can assert an id
// byte for byte.
var (
	adminID = identity.UserID{0xad, 0x01}
	aliceID = identity.UserID{0xa1, 0x1c, 0xe0}
)

// rosterNow is the instant a fixture roster's clock reads unless the
// test moves it.
var rosterNow = time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

// fixedNow returns a clock that always reads at.
func fixedNow(at time.Time) func() time.Time {
	return func() time.Time { return at }
}

// newRosterOver returns a roster over a fresh store holding rows, with
// a fresh epoch store as its revoker. A row with no id, username, or
// creation stamp gets one, so a fixture names only what the test reads.
func newRosterOver(t *testing.T, rows ...userRow) (*member.Roster, *store.Handle) {
	t.Helper()
	h := newUsersStore(t)
	for i := range rows {
		if rows[i].id == (identity.UserID{}) {
			rows[i].id = identity.UserID{0xf1, byte(i + 1)}
		}
		if rows[i].username == "" {
			rows[i].username, _, _ = strings.Cut(rows[i].address, "@")
		}
		if rows[i].createdAt == 0 {
			rows[i].createdAt = int64(100 + i)
		}
	}
	insertUsers(t, h, rows...)
	return member.NewRoster(h, newEpochStore(t, t.TempDir()), fixedNow(rosterNow), discardLogger()), h
}
