package store

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	sqlite "modernc.org/sqlite"
)

// testVersion is the synthetic binary version the write tests open
// their handles against; the fixture sets the store's user_version to
// match, so a zero-valued field or a fresh file cannot pass by default.
const testVersion = 3

// writeFixture is the schema every write test starts from: a parent
// and a child whose foreign key names it, a table of two integers the
// second of which overflows abs(), and a one-row table for the
// disk-full rows.
const writeFixture = `
CREATE TABLE parent (id INTEGER PRIMARY KEY);
CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER NOT NULL REFERENCES parent(id));
CREATE TABLE nums (v INTEGER NOT NULL);
INSERT INTO nums (v) VALUES (1), (-9223372036854775808);
CREATE TABLE small (v INTEGER NOT NULL);
INSERT INTO small (v) VALUES (1);
PRAGMA user_version = 3;
`

// newWriteStore creates a store file holding writeFixture and returns
// its path.
func newWriteStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "s.db")
	precreateFile(t, path)
	setup, err := sql.Open(sqliteDriverName, writeDSN(path, time.Second))
	if err != nil {
		t.Fatalf("open setup: %v", err)
	}
	defer func() { _ = setup.Close() }()
	if _, err := setup.Exec(writeFixture); err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	return path
}

// openWriteHandle opens a handle on path against testVersion, closed
// at cleanup.
func openWriteHandle(t *testing.T, path string, busy time.Duration) *Handle {
	t.Helper()
	h, err := openHandle(path, busy)
	if err != nil {
		t.Fatalf("openHandle: %v", err)
	}
	h.version = testVersion
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// countRows counts table's rows through a fresh connection, so the
// count sees only what committed.
func countRows(t *testing.T, path, table string) int {
	t.Helper()
	db, err := sql.Open(sqliteDriverName, readDSN(path, time.Second))
	if err != nil {
		t.Fatalf("open count: %v", err)
	}
	defer func() { _ = db.Close() }()
	var n int
	if err := db.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

// sqliteCode returns the extended result code of the *sqlite.Error in
// err's chain, or -1 when there is none.
func sqliteCode(err error) int {
	var se *sqlite.Error
	if !errors.As(err, &se) {
		return -1
	}
	return se.Code()
}

// assertWithin runs fn in a goroutine and fails the test if it has not
// returned once d elapses, then waits for it either way.
func assertWithin(t *testing.T, d time.Duration, fn func() error) error {
	t.Helper()
	timer := time.NewTimer(d)
	defer timer.Stop()
	done := make(chan error, 1)
	go func() { done <- fn() }()

	select {
	case err := <-done:
		return err
	case <-timer.C:
		t.Errorf("operation still waiting after %s", d)
		return <-done
	}
}

// deadline returns a context that ends after d, cancelled at cleanup.
func deadline(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

const (
	insertParent = `INSERT INTO parent (id) VALUES (?)`
	insertOrphan = `INSERT INTO child (id, parent_id) VALUES (1, 99)`
)

// sqliteFKCode is SQLITE_CONSTRAINT_FOREIGNKEY, SQLite's extended code
// for a foreign-key failure (sqlite.org/rescode.html, 787).
const sqliteFKCode = 787

func TestUpdateCommitRule(t *testing.T) {
	errStop := errors.New("callback stopped")

	tests := []struct {
		name string
		fn   func(ctx context.Context, t *testing.T, tx *Tx) error
		// check inspects Update's error.
		check       func(t *testing.T, err error)
		wantParents int
	}{
		{
			name: "clean callback commits",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				res, err := tx.Exec(ctx, insertParent, 1)
				if err != nil {
					return err
				}
				if n, err := res.RowsAffected(); err != nil || n != 1 {
					t.Errorf("RowsAffected = %d, %v; want 1, nil", n, err)
				}
				return nil
			},
			check: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("update: %v, want nil", err)
				}
			},
			wantParents: 1,
		},
		{
			name: "callback error rolls back and is returned",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				if _, err := tx.Exec(ctx, insertParent, 1); err != nil {
					return err
				}
				return errStop
			},
			check: func(t *testing.T, err error) {
				if !errors.Is(err, errStop) {
					t.Errorf("update: %v, want errStop", err)
				}
			},
		},
		{
			name: "orphan insert returned fails with the foreign-key code",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				_, err := tx.Exec(ctx, insertOrphan)
				return err
			},
			check: func(t *testing.T, err error) {
				if got := sqliteCode(err); got != sqliteFKCode {
					t.Errorf("update: %v (code %d), want code %d", err, got, sqliteFKCode)
				}
			},
		},
		{
			// A plain *sql.Tx commits the first insert here: SQLite
			// undoes only the failed statement and the transaction
			// continues.
			name: "swallowed foreign-key failure does not commit an earlier insert",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				if _, err := tx.Exec(ctx, insertParent, 1); err != nil {
					return err
				}
				_, _ = tx.Exec(ctx, insertOrphan)
				return nil
			},
			check: func(t *testing.T, err error) {
				if got := sqliteCode(err); got != sqliteFKCode {
					t.Errorf("update: %v (code %d), want the swallowed code %d", err, got, sqliteFKCode)
				}
			},
		},
		{
			name: "swallowed single-row scan error does not commit",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				if _, err := tx.Exec(ctx, insertParent, 1); err != nil {
					return err
				}
				var n int
				_ = tx.QueryRow(ctx, `SELECT 'not a number'`).Scan(&n)
				return nil
			},
			check: func(t *testing.T, err error) {
				if err == nil || !strings.Contains(err.Error(), "converting") {
					t.Errorf("update: %v, want the swallowed scan conversion error", err)
				}
			},
		},
		{
			name: "swallowed single-row query error does not commit",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				if _, err := tx.Exec(ctx, insertParent, 1); err != nil {
					return err
				}
				var n int
				_ = tx.QueryRow(ctx, `SELECT v FROM missing`).Scan(&n)
				return nil
			},
			check: func(t *testing.T, err error) {
				if err == nil || !strings.Contains(err.Error(), "no such table") {
					t.Errorf("update: %v, want the swallowed query error", err)
				}
			},
		},
		{
			// abs() of the smallest integer overflows on the second
			// row, after the first scanned cleanly.
			name: "swallowed iteration error does not commit",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				if _, err := tx.Exec(ctx, insertParent, 1); err != nil {
					return err
				}
				rows, err := tx.Query(ctx, `SELECT abs(v) FROM nums ORDER BY rowid`)
				if err != nil {
					return err
				}
				var seen int
				for rows.Next() {
					var v int64
					if err := rows.Scan(&v); err != nil {
						return err
					}
					seen++
				}
				if err := rows.Close(); err == nil || !strings.Contains(err.Error(), "integer overflow") {
					t.Errorf("close after the overflow: %v, want the iteration error", err)
				}
				if seen != 1 {
					t.Errorf("scanned %d rows before the overflow, want 1", seen)
				}
				return nil
			},
			check: func(t *testing.T, err error) {
				if err == nil || !strings.Contains(err.Error(), "integer overflow") {
					t.Errorf("update: %v, want the swallowed iteration error", err)
				}
			},
		},
		{
			name: "swallowed multi-row scan error does not commit",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				if _, err := tx.Exec(ctx, insertParent, 1); err != nil {
					return err
				}
				rows, err := tx.Query(ctx, `SELECT 'not a number'`)
				if err != nil {
					return err
				}
				for rows.Next() {
					var n int
					_ = rows.Scan(&n)
				}
				_ = rows.Close()
				return nil
			},
			check: func(t *testing.T, err error) {
				if err == nil || !strings.Contains(err.Error(), "converting") {
					t.Errorf("update: %v, want the swallowed scan conversion error", err)
				}
			},
		},
		{
			// A lookup that matches nothing is an answer, not a failure,
			// so it leaves the transaction free to commit.
			name: "no matching row is not recorded",
			fn: func(ctx context.Context, t *testing.T, tx *Tx) error {
				var id int
				if err := tx.QueryRow(ctx, `SELECT id FROM parent WHERE id = ?`, 1).Scan(&id); !errors.Is(err, sql.ErrNoRows) {
					t.Errorf("scan of an absent row: %v, want sql.ErrNoRows", err)
				}
				_, err := tx.Exec(ctx, insertParent, 1)
				return err
			},
			check: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("update: %v, want nil", err)
				}
			},
			wantParents: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := newWriteStore(t)
			h := openWriteHandle(t, path, time.Second)

			err := h.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
				return tt.fn(ctx, t, tx)
			})
			tt.check(t, err)
			if got := countRows(t, path, "parent"); got != tt.wantParents {
				t.Errorf("parent rows = %d, want %d", got, tt.wantParents)
			}
		})
	}
}

// TestUpdateStatementsAfterRecordedError proves that once a statement
// fails, every later statement method answers with that failure and
// never runs: the parent insert below would succeed on its own.
func TestUpdateStatementsAfterRecordedError(t *testing.T) {
	path := newWriteStore(t)
	h := openWriteHandle(t, path, time.Second)

	var first error
	err := h.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
		_, first = tx.Exec(ctx, insertOrphan)
		if first == nil {
			t.Fatal("orphan insert succeeded; the fixture cannot record an error")
		}
		if _, err := tx.Exec(ctx, insertParent, 1); err != first {
			t.Errorf("exec after a recorded error: %v, want the recorded %v", err, first)
		}
		var n int
		if err := tx.QueryRow(ctx, `SELECT 1`).Scan(&n); err != first {
			t.Errorf("queryRow after a recorded error: %v, want the recorded %v", err, first)
		}
		if _, err := tx.Query(ctx, `SELECT 1`); err != first {
			t.Errorf("query after a recorded error: %v, want the recorded %v", err, first)
		}
		return nil
	})
	if err != first {
		t.Errorf("update: %v, want the recorded %v", err, first)
	}
	if got := countRows(t, path, "parent"); got != 0 {
		t.Errorf("parent rows = %d, want 0", got)
	}
}

// pinPageCount caps the connection behind tx at the pages the store
// already has, so a small row still fits in an existing page and an
// oversized one fails with SQLITE_FULL. The cap is per-connection, so
// the caller closes the handle rather than trusting a reset.
func pinPageCount(ctx context.Context, t *testing.T, tx *Tx) {
	t.Helper()
	var pages int
	if err := tx.QueryRow(ctx, `PRAGMA page_count`).Scan(&pages); err != nil {
		t.Fatalf("read page_count: %v", err)
	}
	if _, err := tx.Exec(ctx, "PRAGMA max_page_count = "+strconv.Itoa(pages)); err != nil {
		t.Fatalf("pin max_page_count: %v", err)
	}
}

const insertOversized = `INSERT INTO small (v) VALUES (randomblob(65536))`

// TestUpdateDiskFull proves a callback returning SQLITE_FULL gets the
// named disk-full error, naming the store, with the rollback's "no
// transaction is active" answer suppressed and no row committed.
func TestUpdateDiskFull(t *testing.T) {
	path := newWriteStore(t)
	h := openWriteHandle(t, path, time.Second)

	err := h.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
		pinPageCount(ctx, t, tx)
		if _, err := tx.Exec(ctx, `INSERT INTO small (v) VALUES (2)`); err != nil {
			t.Fatalf("small insert under the pinned cap: %v", err)
		}
		_, err := tx.Exec(ctx, insertOversized)
		return err
	})
	if !errors.Is(err, errDiskFull) {
		t.Fatalf("update: %v, want errDiskFull", err)
	}
	if got := sqliteCode(err) & 0xff; got != sqliteFullCode {
		t.Errorf("update: SQLite code %d in chain, want %d", got, sqliteFullCode)
	}
	if !strings.Contains(err.Error(), path) {
		t.Errorf("update: %q does not name the store %s", err, path)
	}
	if strings.Contains(err.Error(), "no transaction is active") {
		t.Errorf("update: %q carries the rollback's answer", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close handle: %v", err)
	}
	if got := countRows(t, path, "small"); got != 1 {
		t.Errorf("small rows = %d, want the fixture's 1", got)
	}
}

// TestUpdateNoAutocommitAfterRollback is the shape SQLite's own
// rollback leaves: after SQLITE_FULL ends the transaction, the
// connection is in autocommit mode, so a late statement that reached
// SQLite would commit by itself even though the callback returns nil.
func TestUpdateNoAutocommitAfterRollback(t *testing.T) {
	path := newWriteStore(t)
	h := openWriteHandle(t, path, time.Second)

	err := h.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
		pinPageCount(ctx, t, tx)
		if _, err := tx.Exec(ctx, `INSERT INTO small (v) VALUES (2)`); err != nil {
			t.Fatalf("small insert under the pinned cap: %v", err)
		}
		if _, err := tx.Exec(ctx, insertOversized); sqliteCode(err)&0xff != sqliteFullCode {
			t.Fatalf("oversized insert: %v, want SQLITE_FULL", err)
		}
		_, _ = tx.Exec(ctx, `INSERT INTO small (v) VALUES (3)`)
		return nil
	})
	if !errors.Is(err, errDiskFull) {
		t.Errorf("update: %v, want errDiskFull", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close handle: %v", err)
	}
	if got := countRows(t, path, "small"); got != 1 {
		t.Errorf("small rows = %d, want the fixture's 1 (a late row committed in autocommit mode)", got)
	}
}

func TestUpdateVersionCheck(t *testing.T) {
	tests := []struct {
		name string
		// storeVersion is written to the store after the handle opens.
		storeVersion int
		wantErr      error
		wantMsg      string
		wantCalled   bool
		wantParents  int
	}{
		{name: "equal runs the callback", storeVersion: testVersion, wantCalled: true, wantParents: 1},
		{
			name:         "changed under the handle fails closed",
			storeVersion: testVersion + 1,
			wantErr:      errVersionMismatch,
			wantMsg:      "store at 4, binary at 3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := newWriteStore(t)
			h := openWriteHandle(t, path, time.Second)

			other, err := sql.Open(sqliteDriverName, writeDSN(path, time.Second))
			if err != nil {
				t.Fatalf("open second handle: %v", err)
			}
			defer func() { _ = other.Close() }()
			if _, err := other.Exec("PRAGMA user_version = " + strconv.Itoa(tt.storeVersion)); err != nil {
				t.Fatalf("set user_version: %v", err)
			}

			var called bool
			err = h.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
				called = true
				_, err := tx.Exec(ctx, insertParent, 1)
				return err
			})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("update: %v, want %v", err, tt.wantErr)
			}
			if tt.wantMsg != "" && (err == nil || !strings.Contains(err.Error(), tt.wantMsg)) {
				t.Errorf("update: %v, want a message containing %q", err, tt.wantMsg)
			}
			if called != tt.wantCalled {
				t.Errorf("callback called = %v, want %v", called, tt.wantCalled)
			}
			if got := countRows(t, path, "parent"); got != tt.wantParents {
				t.Errorf("parent rows = %d, want %d", got, tt.wantParents)
			}
		})
	}
}

// TestUpdateVersionReadInsideTransaction proves the version is read
// after the write lock is taken: another writer bumps user_version
// inside a transaction the update must wait behind, and commits at
// half the busy timeout. A version read before the lock sees the old
// value and lets the update commit on a store it no longer matches.
func TestUpdateVersionReadInsideTransaction(t *testing.T) {
	const busy = 3 * time.Second
	path := newWriteStore(t)
	h := openWriteHandle(t, path, busy)

	other, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
	if err != nil {
		t.Fatalf("open second handle: %v", err)
	}
	defer func() { _ = other.Close() }()
	otx, err := other.Begin()
	if err != nil {
		t.Fatalf("begin second handle: %v", err)
	}
	if _, err := otx.Exec("PRAGMA user_version = 4"); err != nil {
		t.Fatalf("bump user_version: %v", err)
	}
	commitErr := make(chan error, 1)
	go func() {
		time.Sleep(busy / 6)
		commitErr <- otx.Commit()
	}()

	err = h.Update(deadline(t, 2*busy), func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx, insertParent, 1)
		return err
	})
	if err := <-commitErr; err != nil {
		t.Fatalf("second handle commit: %v", err)
	}
	if !errors.Is(err, errVersionMismatch) {
		t.Errorf("update: %v, want errVersionMismatch", err)
	}
	if got := countRows(t, path, "parent"); got != 0 {
		t.Errorf("parent rows = %d, want 0", got)
	}
}

func TestUpdateDeadline(t *testing.T) {
	const busy = 400 * time.Millisecond

	t.Run("none is refused without waiting", func(t *testing.T) {
		path := newWriteStore(t)
		h := openWriteHandle(t, path, busy)
		// Holding the one write connection makes any wait for it block
		// until ctx is cancelled.
		held, err := h.write.Conn(context.Background())
		if err != nil {
			t.Fatalf("hold the write connection: %v", err)
		}
		defer func() { _ = held.Close() }()

		// Cancellable but deadline-free, so a wait that should not
		// happen can still be ended once the timer has failed the test.
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan error, 1)
		go func() {
			done <- h.Update(ctx, func(context.Context, *Tx) error { return nil })
		}()
		timer := time.NewTimer(2 * time.Second)
		defer timer.Stop()
		select {
		case err = <-done:
		case <-timer.C:
			t.Error("update without a deadline is still waiting after 2s")
			cancel()
			err = <-done
		}
		if !errors.Is(err, errNoDeadline) {
			t.Errorf("update: %v, want errNoDeadline", err)
		}
	})

	t.Run("expires waiting for the in-process connection", func(t *testing.T) {
		path := newWriteStore(t)
		h := openWriteHandle(t, path, busy)
		held, err := h.write.Conn(context.Background())
		if err != nil {
			t.Fatalf("hold the write connection: %v", err)
		}
		defer func() { _ = held.Close() }()

		err = h.Update(deadline(t, 100*time.Millisecond), func(context.Context, *Tx) error { return nil })
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("update: %v, want context.DeadlineExceeded", err)
		}
	})

	// Measured: SQLite's busy handler does not stop for the driver's
	// interrupt, so a deadline shorter than the busy timeout still waits
	// the full busy timeout on another process's write lock.
	t.Run("shorter than the busy timeout waits the busy timeout", func(t *testing.T) {
		path := newWriteStore(t)
		h := openWriteHandle(t, path, busy)
		other := openWriteHandle(t, path, busy)
		otx, err := other.write.BeginTx(context.Background(), nil)
		if err != nil {
			t.Fatalf("begin second handle: %v", err)
		}
		defer func() { _ = otx.Rollback() }()

		err = assertNotBefore(t, busy, func() error {
			return h.Update(deadline(t, 100*time.Millisecond), func(context.Context, *Tx) error { return nil })
		})
		if !errors.Is(err, ErrBusy) {
			t.Errorf("update: %v, want ErrBusy", err)
		}
	})
}

func TestUpdateNesting(t *testing.T) {
	noop := func(context.Context, *Tx) error { return nil }

	t.Run("on the callback's context fails at once", func(t *testing.T) {
		path := newWriteStore(t)
		h := openWriteHandle(t, path, time.Second)

		var nested error
		err := h.Update(deadline(t, 3*time.Second), func(ctx context.Context, tx *Tx) error {
			nested = assertWithin(t, 500*time.Millisecond, func() error { return h.Update(ctx, noop) })
			return nil
		})
		if !errors.Is(nested, errNestedWrite) {
			t.Errorf("nested update: %v, want errNestedWrite", nested)
		}
		if err != nil {
			t.Errorf("outer update: %v, want nil", err)
		}
	})

	// The documented limit: a nested call on a context the callback
	// closed over carries no mark, so it waits for the connection its
	// own caller holds until that context ends.
	t.Run("on a closed-over outer context waits for it to end", func(t *testing.T) {
		path := newWriteStore(t)
		h := openWriteHandle(t, path, time.Second)

		outer := deadline(t, 300*time.Millisecond)
		var nested error
		_ = h.Update(outer, func(context.Context, *Tx) error {
			nested = h.Update(outer, noop)
			return nil
		})
		if !errors.Is(nested, context.DeadlineExceeded) {
			t.Errorf("nested update: %v, want context.DeadlineExceeded", nested)
		}
	})
}

// TestUpdateTwoHandlesWait holds one handle's callback for half the
// busy timeout and proves a second handle's update on the same file
// waits for it and commits.
func TestUpdateTwoHandlesWait(t *testing.T) {
	const busy = 600 * time.Millisecond
	path := newWriteStore(t)
	first := openWriteHandle(t, path, busy)
	second := openWriteHandle(t, path, busy)

	entered := make(chan struct{})
	firstErr := make(chan error, 1)
	go func() {
		firstErr <- first.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
			if _, err := tx.Exec(ctx, insertParent, 1); err != nil {
				return err
			}
			close(entered)
			time.Sleep(busy / 2)
			return nil
		})
	}()
	<-entered

	err := second.Update(deadline(t, 5*time.Second), func(ctx context.Context, tx *Tx) error {
		_, err := tx.Exec(ctx, insertParent, 2)
		return err
	})
	if err != nil {
		t.Errorf("second update: %v, want it to wait and commit", err)
	}
	if err := <-firstErr; err != nil {
		t.Errorf("first update: %v", err)
	}
	if got := countRows(t, path, "parent"); got != 2 {
		t.Errorf("parent rows = %d, want 2", got)
	}
}
