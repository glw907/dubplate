package store

import (
	"context"
	"database/sql"
	"fmt"
)

// QueryAll runs query on h's read pool and returns what scan makes of
// each row, in the order the rows arrive. A query that matches nothing
// returns an empty, non-nil slice.
func QueryAll[T any](ctx context.Context, h *Handle, scan func(*sql.Rows) (T, error), query string, args ...any) ([]T, error) {
	rows, err := h.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []T{}
	for rows.Next() {
		v, err := scan(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, rows.Err()
}

// Affected reports whether res changed a row.
func Affected(res sql.Result) (bool, error) {
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("count changed rows: %w", err)
	}
	return n > 0, nil
}
