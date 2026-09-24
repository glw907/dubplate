package store

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	sqlite "modernc.org/sqlite"
)

// precreateFile makes an empty file at path, the shape serve's
// pre-create leaves before the first migration runs. No DSN this
// package builds can create a database, so every test that opens a
// writing class precreates the file itself first.
func precreateFile(t *testing.T, path string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("precreate %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close precreated %s: %v", path, err)
	}
}

// assertPragma fails the test unless PRAGMA name on db's first (and,
// for a capped-at-one pool, only) connection reads back want.
func assertPragma(t *testing.T, db *sql.DB, name, want string) {
	t.Helper()
	var got string
	if err := db.QueryRow("PRAGMA " + name).Scan(&got); err != nil {
		t.Fatalf("PRAGMA %s: %v", name, err)
	}
	if got != want {
		t.Errorf("PRAGMA %s = %q, want %q", name, got, want)
	}
}

// assertFreshPragma is assertPragma forced onto a connection db's pool
// has not yet opened: it holds one connection with DB.Conn so the
// PRAGMA query below must open a second, new one. A pool whose settings
// come from a one-time statement after open, rather than from every
// connection's DSN, passes assertPragma but fails this.
func assertFreshPragma(t *testing.T, db *sql.DB, name, want string) {
	t.Helper()
	held, err := db.Conn(context.Background())
	if err != nil {
		t.Fatalf("hold a connection: %v", err)
	}
	defer func() { _ = held.Close() }()

	assertPragma(t, db, name, want)
}

func TestConnectionClassSettings(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)
	const busy = 300 * time.Millisecond

	t.Run("write pool", func(t *testing.T) {
		h, err := openHandle(path, busy)
		if err != nil {
			t.Fatalf("openHandle: %v", err)
		}
		t.Cleanup(func() { _ = h.close() })

		assertPragma(t, h.write, "journal_mode", "wal")
		assertPragma(t, h.write, "foreign_keys", "1")
		assertPragma(t, h.write, "synchronous", "2")
		assertPragma(t, h.write, "busy_timeout", "300")
		assertPragma(t, h.write, "journal_size_limit", "67108864")
	})

	t.Run("read pool", func(t *testing.T) {
		h, err := openHandle(path, busy)
		if err != nil {
			t.Fatalf("openHandle: %v", err)
		}
		t.Cleanup(func() { _ = h.close() })

		assertFreshPragma(t, h.read, "journal_mode", "wal")
		assertFreshPragma(t, h.read, "foreign_keys", "1")
		assertFreshPragma(t, h.read, "synchronous", "2")
		assertFreshPragma(t, h.read, "busy_timeout", "300")
		assertFreshPragma(t, h.read, "journal_size_limit", "67108864")
		assertFreshPragma(t, h.read, "query_only", "1")
	})

	t.Run("maintenance", func(t *testing.T) {
		db, err := openMaintenanceConn(path, busy)
		if err != nil {
			t.Fatalf("openMaintenanceConn: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		assertPragma(t, db, "foreign_keys", "1")
		assertPragma(t, db, "synchronous", "2")
		assertPragma(t, db, "busy_timeout", "300")
		assertPragma(t, db, "journal_size_limit", "67108864")
	})

	t.Run("inspect", func(t *testing.T) {
		db, err := sql.Open(sqliteDriverName, inspectDSN(path, busy))
		if err != nil {
			t.Fatalf("open: %v", err)
		}
		t.Cleanup(func() { _ = db.Close() })

		assertFreshPragma(t, db, "busy_timeout", "300")
		assertFreshPragma(t, db, "query_only", "1")
	})
}

// TestDSNRejectsUnknownShorthandValue proves a shorthand value outside
// the driver's accepted set fails at the driver, not silently: the
// PingContext call is what actually dials the connection, since
// sql.Open never does.
func TestDSNRejectsUnknownShorthandValue(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)

	dsn := buildDSN(path, "mode=rw", "_synchronous=NOTAVALUE")
	db, err := sql.Open(sqliteDriverName, dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if err := db.PingContext(context.Background()); err == nil {
		t.Error("ping with an invalid _synchronous value: got nil error, want the driver's own rejection")
	}
}

// TestDSNEscapesSpecialPathChars proves the escaping DSN builder, not
// string concatenation: a directory whose name carries '#', '%', and
// '?' opens its database at the exact path, and creates nothing beside
// the directory itself (the fold probe's P2 shape: a concatenated DSN
// would truncate the path and create a stray entry such as "a" next to
// the directory).
func TestDSNEscapesSpecialPathChars(t *testing.T) {
	const dirName = "a#b%41?c"
	base := t.TempDir()
	dir := filepath.Join(base, dirName)
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", dir, err)
	}
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)
	const busy = 300 * time.Millisecond

	// The write class runs first and sets WAL mode via its
	// _journal_mode DSN param, since mode persists in the database
	// file: the other three classes open the same file afterward and
	// must see it, not set it themselves.
	classes := []struct {
		name string
		dsn  string
	}{
		{"write", writeDSN(path, busy)},
		{"read", readDSN(path, busy)},
		{"maintenance", maintenanceDSN(path, busy)},
		{"inspect", inspectDSN(path, busy)},
	}

	for _, class := range classes {
		t.Run(class.name, func(t *testing.T) {
			db, err := sql.Open(sqliteDriverName, class.dsn)
			if err != nil {
				t.Fatalf("open: %v", err)
			}
			t.Cleanup(func() { _ = db.Close() })

			if class.name == "write" {
				if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS t (id INTEGER)`); err != nil {
					t.Fatalf("create table: %v", err)
				}
			} else {
				var got string
				if err := db.QueryRow("PRAGMA busy_timeout").Scan(&got); err != nil {
					t.Fatalf("PRAGMA busy_timeout: %v", err)
				}
			}

			entries, err := os.ReadDir(base)
			if err != nil {
				t.Fatalf("read dir: %v", err)
			}
			if len(entries) != 1 || entries[0].Name() != dirName {
				t.Errorf("base directory listing = %v, want exactly [%q]", entries, dirName)
			}
		})
	}

	// A writing class against an absent file at the same escaped
	// directory returns SQLITE_CANTOPEN and creates nothing.
	absentPath := filepath.Join(dir, "absent.db")
	absentDB, err := sql.Open(sqliteDriverName, writeDSN(absentPath, 300*time.Millisecond))
	if err != nil {
		t.Fatalf("open absent: %v", err)
	}
	t.Cleanup(func() { _ = absentDB.Close() })

	before, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir before: %v", err)
	}
	_, err = absentDB.Exec(`CREATE TABLE t (id INTEGER)`)
	if err == nil {
		t.Fatal("write against an absent file: got nil error, want SQLITE_CANTOPEN (14)")
	}
	var se *sqlite.Error
	if !errors.As(err, &se) || se.Code() != 14 {
		t.Errorf("write against an absent file: err = %v, want a *sqlite.Error with code 14", err)
	}
	after, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir after: %v", err)
	}
	if len(after) != len(before) {
		t.Errorf("directory listing changed after the absent-file write attempt: before %v, after %v", before, after)
	}
}

// TestReadPoolRefusesWrite proves criterion 5's first half: a write
// issued through the read pool's DSN fails, and classifyErr wraps that
// failure in errReadOnly so a caller can match it with errors.Is.
func TestReadPoolRefusesWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)
	const busy = 300 * time.Millisecond

	h, err := openHandle(path, busy)
	if err != nil {
		t.Fatalf("openHandle: %v", err)
	}
	t.Cleanup(func() { _ = h.close() })

	_, err = h.read.Exec(`CREATE TABLE t (id INTEGER)`)
	if err == nil {
		t.Fatal("write on the read pool: got nil error, want the read-only refusal")
	}
	if !errors.Is(classifyErr(err), errReadOnly) {
		t.Errorf("classifyErr(%v) does not match errReadOnly", err)
	}
}

// assertNotBefore runs fn in a goroutine and fails the test if it
// returns before d elapses, ordering the check against a [time.Timer]
// rather than a start/elapsed time.Now measurement: internal/engine
// (this package's parent tree) never reads the clock.
func assertNotBefore(t *testing.T, d time.Duration, fn func() error) error {
	t.Helper()
	timer := time.NewTimer(d)
	defer timer.Stop()
	done := make(chan error, 1)
	go func() { done <- fn() }()

	select {
	case err := <-done:
		t.Errorf("operation returned before the %s deadline", d)
		return err
	case <-timer.C:
		return <-done
	}
}

// beginAndWrite begins a transaction on db and writes through it,
// returning whichever step errors first. An IMMEDIATE begin mode takes
// the write lock at BeginTx, so a contending handle's error surfaces
// there; a deferred begin defers the same contention to the write
// statement. Testing "the second writer" against this helper, rather
// than against BeginTx alone, holds for either begin mode.
func beginAndWrite(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.Exec(`INSERT INTO t DEFAULT VALUES`); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// TestBusyTimeout proves criterion 3 at the connection level: a second
// writer waits behind the first's IMMEDIATE transaction and commits
// once it releases; held past the busy timeout, the second returns the
// busy error.
func TestBusyTimeout(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)
	const busy = 300 * time.Millisecond
	ctx := context.Background()

	setup, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
	if err != nil {
		t.Fatalf("open setup: %v", err)
	}
	if _, err := setup.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if err := setup.Close(); err != nil {
		t.Fatalf("close setup: %v", err)
	}

	t.Run("waits then commits", func(t *testing.T) {
		a, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
		if err != nil {
			t.Fatalf("open a: %v", err)
		}
		t.Cleanup(func() { _ = a.Close() })
		b, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
		if err != nil {
			t.Fatalf("open b: %v", err)
		}
		t.Cleanup(func() { _ = b.Close() })

		txA, err := a.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin a: %v", err)
		}
		if _, err := txA.Exec(`INSERT INTO t DEFAULT VALUES`); err != nil {
			t.Fatalf("a insert: %v", err)
		}

		// Release a's hold at half the busy timeout, ordered by a
		// channel rather than a race between the release and b's
		// blocked write.
		commitErr := make(chan error, 1)
		go func() {
			time.Sleep(busy / 2)
			commitErr <- txA.Commit()
		}()

		txB, err := b.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin b: %v", err)
		}
		if _, err := txB.Exec(`INSERT INTO t DEFAULT VALUES`); err != nil {
			t.Fatalf("b waited for a and got: %v", err)
		}
		if err := <-commitErr; err != nil {
			t.Fatalf("a commit: %v", err)
		}
		if err := txB.Commit(); err != nil {
			t.Fatalf("b commit: %v", err)
		}
	})

	t.Run("past the timeout returns the busy error", func(t *testing.T) {
		a, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
		if err != nil {
			t.Fatalf("open a: %v", err)
		}
		t.Cleanup(func() { _ = a.Close() })
		txA, err := a.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin a: %v", err)
		}
		if _, err := txA.Exec(`INSERT INTO t DEFAULT VALUES`); err != nil {
			t.Fatalf("a insert: %v", err)
		}
		t.Cleanup(func() { _ = txA.Rollback() })

		b, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
		if err != nil {
			t.Fatalf("open b: %v", err)
		}
		t.Cleanup(func() { _ = b.Close() })

		err = assertNotBefore(t, busy, func() error { return beginAndWrite(ctx, b) })
		if err == nil {
			t.Fatal("b writes while a holds: got nil error, want the busy error")
		}
		if !errors.Is(classifyErr(err), errBusy) {
			t.Errorf("classifyErr(%v) does not match errBusy", err)
		}
	})
}

// TestBusyErrorClassifiesExtendedCode is the mechanics review's probe
// E shape: a deferred reader that later tries to upgrade to a write
// after missing a commit gets SQLITE_BUSY_SNAPSHOT (517), and
// classifyErr still matches it to errBusy by masking the extended code.
// A maintenance-class transaction under the same sequence takes the
// write lock at its BEGIN, so it never meets that race.
func TestBusyErrorClassifiesExtendedCode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)
	const busy = 300 * time.Millisecond
	ctx := context.Background()

	setup, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
	if err != nil {
		t.Fatalf("open setup: %v", err)
	}
	if _, err := setup.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if err := setup.Close(); err != nil {
		t.Fatalf("close setup: %v", err)
	}

	t.Run("deferred upgrade returns 517", func(t *testing.T) {
		// A test-opened handle whose DSN carries no _txlock, so its
		// BEGIN is deferred rather than immediate.
		deferredDSN := buildDSN(path, "mode=rw", busyTimeoutParam(busy), "_journal_mode=WAL")
		a, err := sql.Open(sqliteDriverName, deferredDSN)
		if err != nil {
			t.Fatalf("open a: %v", err)
		}
		t.Cleanup(func() { _ = a.Close() })

		txA, err := a.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin a: %v", err)
		}
		var cnt int
		if err := txA.QueryRow(`SELECT count(*) FROM t`).Scan(&cnt); err != nil {
			t.Fatalf("a read: %v", err)
		}

		b, err := sql.Open(sqliteDriverName, maintenanceDSN(path, busy))
		if err != nil {
			t.Fatalf("open b: %v", err)
		}
		t.Cleanup(func() { _ = b.Close() })
		txB, err := b.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin b: %v", err)
		}
		if _, err := txB.Exec(`INSERT INTO t DEFAULT VALUES`); err != nil {
			t.Fatalf("b insert: %v", err)
		}
		if err := txB.Commit(); err != nil {
			t.Fatalf("b commit: %v", err)
		}

		_, err = txA.Exec(`INSERT INTO t DEFAULT VALUES`)
		t.Cleanup(func() { _ = txA.Rollback() })
		if err == nil {
			t.Fatal("a's deferred upgrade after b's commit: got nil error, want SQLITE_BUSY_SNAPSHOT")
		}
		var se *sqlite.Error
		if !errors.As(err, &se) || se.Code() != 517 {
			t.Fatalf("a's deferred upgrade: err = %v, want a *sqlite.Error with code 517", err)
		}
		if !errors.Is(classifyErr(err), errBusy) {
			t.Errorf("classifyErr(%v) does not match errBusy", err)
		}
	})

	t.Run("maintenance transaction never returns 517", func(t *testing.T) {
		m, err := sql.Open(sqliteDriverName, maintenanceDSN(path, busy))
		if err != nil {
			t.Fatalf("open m: %v", err)
		}
		t.Cleanup(func() { _ = m.Close() })

		// BeginTx runs "begin immediate" synchronously, so m already
		// holds the write lock before any write statement of its own.
		txM, err := m.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin m: %v", err)
		}
		t.Cleanup(func() { _ = txM.Rollback() })

		other, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
		if err != nil {
			t.Fatalf("open other: %v", err)
		}
		t.Cleanup(func() { _ = other.Close() })

		err = assertNotBefore(t, busy, func() error {
			_, err := other.Exec(`INSERT INTO t DEFAULT VALUES`)
			return err
		})
		if err == nil {
			t.Fatal("other's write while m holds the immediate lock: got nil error, want the busy error")
		}
		if !errors.Is(classifyErr(err), errBusy) {
			t.Errorf("classifyErr(%v) does not match errBusy", err)
		}

		// m's own write never needed to upgrade a lock it already
		// held, so it never meets 517.
		if _, err := txM.Exec(`INSERT INTO t DEFAULT VALUES`); err != nil {
			t.Errorf("m insert after already holding the lock: %v, want nil", err)
		}
		if err := txM.Commit(); err != nil {
			t.Fatalf("m commit: %v", err)
		}
	})
}
