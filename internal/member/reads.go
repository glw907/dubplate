package member

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/store"
)

// EarliestAdmin returns the id of the live admin created first, breaking
// a tie in creation time by the lower id. It reports false when no live
// admin exists. A new collection is owned by this member until the owner
// is changed. A read error comes back bare: both callers add the prefix
// their operators already read.
func EarliestAdmin(ctx context.Context, tx *store.Tx) (identity.UserID, bool, error) {
	var raw []byte
	err := tx.QueryRow(ctx, `SELECT id FROM users WHERE admin = 1 AND tombstoned_at IS NULL
		ORDER BY created_at, id LIMIT 1`).Scan(&raw)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return identity.UserID{}, false, nil
	case err != nil:
		return identity.UserID{}, false, err
	}
	var id identity.UserID
	if len(raw) != len(id) {
		return identity.UserID{}, false, fmt.Errorf("member: the earliest admin's id is %d bytes, want %d", len(raw), len(id))
	}
	copy(id[:], raw)
	return id, true, nil
}

// Live counts the members who can sign in: every row of the users table
// that is not tombstoned.
func Live(ctx context.Context, h *store.Handle) (int, error) {
	var n int
	if err := h.QueryRow(ctx, `SELECT count(*) FROM users WHERE tombstoned_at IS NULL`).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}
