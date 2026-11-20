package store

import (
	"context"
	"database/sql"
	"errors"
	"slices"
	"testing"
	"time"
)

// openSmall returns a handle on a store whose table small holds the
// rows 1, 2, and 3, and a connection that can write to it.
func openSmall(t *testing.T) (*Handle, *sql.DB) {
	t.Helper()
	path := newWriteStore(t)
	db, err := sql.Open(sqliteDriverName, writeDSN(path, time.Second))
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(`INSERT INTO small (v) VALUES (2), (3)`); err != nil {
		t.Fatalf("plant rows: %v", err)
	}
	return openWriteHandle(t, path, time.Second), db
}

func scanV(rows *sql.Rows) (int, error) {
	var v int
	err := rows.Scan(&v)
	return v, err
}

// TestQueryAll: three planted rows come back in query order, and a
// query that matches nothing returns an empty, non-nil slice.
func TestQueryAll(t *testing.T) {
	h, _ := openSmall(t)
	ctx := context.Background()

	got, err := QueryAll(ctx, h, scanV, `SELECT v FROM small ORDER BY v DESC`)
	if err != nil {
		t.Fatalf("QueryAll: %v", err)
	}
	if want := []int{3, 2, 1}; !slices.Equal(got, want) {
		t.Errorf("QueryAll = %v, want %v", got, want)
	}

	none, err := QueryAll(ctx, h, scanV, `SELECT v FROM small WHERE v > ?`, 99)
	if err != nil {
		t.Fatalf("QueryAll, no rows: %v", err)
	}
	if none == nil || len(none) != 0 {
		t.Errorf("QueryAll, no rows = %#v, want an empty non-nil slice", none)
	}
}

// TestQueryAll_ScanError: a scan that fails ends the read with that
// error and no rows.
func TestQueryAll_ScanError(t *testing.T) {
	h, _ := openSmall(t)
	boom := errors.New("scan failed")
	got, err := QueryAll(context.Background(), h, func(*sql.Rows) (int, error) { return 0, boom },
		`SELECT v FROM small`)
	if !errors.Is(err, boom) || got != nil {
		t.Errorf("QueryAll = %v, %v; want nil and %v", got, err, boom)
	}
}

type countErrResult struct{ sql.Result }

func (countErrResult) RowsAffected() (int64, error) { return 0, errors.New("driver gave no count") }

// TestAffected: an update that matches one row reports true, one that
// matches none reports false, and a result with no count is an error.
func TestAffected(t *testing.T) {
	_, db := openSmall(t)

	hit, err := db.Exec(`UPDATE small SET v = v + 10 WHERE v = 1`)
	if err != nil {
		t.Fatalf("update one: %v", err)
	}
	if ok, err := Affected(hit); err != nil || !ok {
		t.Errorf("Affected(one row) = %v, %v; want true, nil", ok, err)
	}

	miss, err := db.Exec(`UPDATE small SET v = v + 10 WHERE v = 99`)
	if err != nil {
		t.Fatalf("update none: %v", err)
	}
	if ok, err := Affected(miss); err != nil || ok {
		t.Errorf("Affected(no rows) = %v, %v; want false, nil", ok, err)
	}

	if ok, err := Affected(countErrResult{}); err == nil || ok {
		t.Errorf("Affected(no count) = %v, %v; want false and an error", ok, err)
	}
}
