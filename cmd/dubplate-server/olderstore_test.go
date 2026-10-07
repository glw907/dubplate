package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// migrationTables names, per migration file's version, every table the
// file creates. olderStore drops the tables of each version above its
// target, so the store re-migrates cleanly: a migration whose tables
// this list does not name fails the re-migration with "table already
// exists" rather than passing vacuously.
//
// Version 1 needs no entry: the ledger table is the migration's own.
// Each list drops a table only after the tables that reference it, and
// an index goes with its table.
var migrationTables = map[int][]string{
	2: {"jobs"},
	3: {"tracks", "albums", "collections", "users"},
	4: {"settings", "settings_audit", "settings_generation"},
	5: {"move_intents", "device_credentials", "group_quota", "member_quota", "member_preferences", "member_overrides"},
}

// storeDBPath is the store database file under stateDir.
func storeDBPath(stateDir string) string {
	return filepath.Join(stateDir, "store", "db", "dubplate.db")
}

// olderStore creates a store under stateDir through [store.OpenServe]
// and then cuts it back to version target: the tables each migration
// above target created are dropped, the ledger loses their rows, and
// user_version is set to target. It returns the database path, closed
// with no -wal or -shm beside it. A migration above target that
// migrationTables does not list fails the later re-migration.
func olderStore(t *testing.T, stateDir string, target int) string {
	t.Helper()
	h, err := store.OpenServe(context.Background(), stateDir, store.ServeOptions{
		BusyTimeout:   testBackupBusy,
		MigrationWait: 30 * time.Second,
		Keep:          1,
		Logger:        slog.New(slog.DiscardHandler),
	})
	if err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close serve handle: %v", err)
	}

	head := store.Inspect(context.Background(), store.InspectOptions{StateDir: stateDir, BusyTimeout: testBackupBusy}).BinaryVersion
	if target < 1 || target > head {
		t.Fatalf("olderStore target %d: want a version from 1 through %d", target, head)
	}
	path := storeDBPath(stateDir)
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	for v := head; v > target; v-- {
		for _, table := range migrationTables[v] {
			if _, err := db.Exec("DROP TABLE " + table); err != nil {
				t.Fatalf("drop %s of migration %d: %v", table, v, err)
			}
		}
	}
	if _, err := db.Exec(fmt.Sprintf("DELETE FROM migrations WHERE version > %d; PRAGMA user_version = %d", target, target)); err != nil {
		t.Fatalf("cut %s to version %d: %v", path, target, err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, side := range []string{"-wal", "-shm"} {
		if _, err := os.Stat(path + side); !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("%s%s after the close: %v, want none", path, side, err)
		}
	}
	return path
}

// execStoreSQL runs query against the closed store file at path
// through a raw connection that is the last to close it, so no -wal
// remains.
func execStoreSQL(t *testing.T, path, query string, args ...any) {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(query, args...); err != nil {
		t.Fatalf("%s: %v", strings.Fields(query)[0], err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestOlderStore proves the helper before the start-state tests lean
// on it: the store it leaves reports older at the target version with
// the migrations above it pending, and migrates to the binary's
// version through the real runner, which a store with a surviving
// table above the target would refuse.
func TestOlderStore(t *testing.T) {
	head := store.Inspect(context.Background(), store.InspectOptions{StateDir: t.TempDir(), BusyTimeout: testBackupBusy}).BinaryVersion
	for target := 1; target < head; target++ {
		t.Run(fmt.Sprintf("v%d", target), func(t *testing.T) {
			stateDir := t.TempDir()
			olderStore(t, stateDir, target)

			r := store.Inspect(context.Background(), store.InspectOptions{StateDir: stateDir, BusyTimeout: testBackupBusy})
			if r.State != store.StateOlder || r.StoreVersion != target {
				t.Fatalf("report = %v at v%d (%s), want older at v%d", r.State, r.StoreVersion, r.Reason, target)
			}
			if got := len(r.Pending); got != head-target {
				t.Errorf("pending = %v, want %d files", r.Pending, head-target)
			}

			h, err := store.OpenServe(context.Background(), stateDir, store.ServeOptions{
				BusyTimeout: testBackupBusy, MigrationWait: 30 * time.Second, Keep: 3, Logger: slog.New(slog.DiscardHandler),
			})
			if err != nil {
				t.Fatalf("re-migrate: %v", err)
			}
			if err := h.Close(); err != nil {
				t.Fatal(err)
			}
			if r := store.Inspect(context.Background(), store.InspectOptions{StateDir: stateDir, BusyTimeout: testBackupBusy}); r.State != store.StateCurrent {
				t.Errorf("state after re-migration = %v (%s), want current", r.State, r.Reason)
			}
		})
	}
}
