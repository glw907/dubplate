package store

import (
	"context"
	"database/sql"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// compileOptions returns every value PRAGMA compile_options reports for
// db, the SQLite build's compile-time feature list.
func compileOptions(t *testing.T, db *sql.DB) []string {
	t.Helper()
	rows, err := db.Query("PRAGMA compile_options")
	if err != nil {
		t.Fatalf("PRAGMA compile_options: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var opts []string
	for rows.Next() {
		var opt string
		if err := rows.Scan(&opt); err != nil {
			t.Fatalf("scan compile_options row: %v", err)
		}
		opts = append(opts, opt)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("compile_options rows: %v", err)
	}
	return opts
}

// TestAdoptionFTS5Trigram is one of the two adoption conditions the
// spec's prior-art verdict sets for modernc.org/sqlite: FTS5's trigram
// tokenizer answers a substring query. The spec's path rule also relies
// on LIKE_DOESNT_MATCH_BLOBS, asserted here since both conditions read
// PRAGMA compile_options off the same connection.
func TestAdoptionFTS5Trigram(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)

	db, err := sql.Open(sqliteDriverName, writeDSN(path, 300*time.Millisecond))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if !slices.Contains(compileOptions(t, db), "LIKE_DOESNT_MATCH_BLOBS") {
		t.Error("compile_options: LIKE_DOESNT_MATCH_BLOBS not present")
	}

	if _, err := db.Exec(`CREATE TABLE tracks (id INTEGER PRIMARY KEY, title TEXT)`); err != nil {
		t.Fatalf("create tracks: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO tracks (title) VALUES (?), (?)`,
		"Bohemian Rhapsody", "A Different Track"); err != nil {
		t.Fatalf("insert tracks: %v", err)
	}
	if _, err := db.Exec(`CREATE VIRTUAL TABLE tracks_fts USING fts5(title, tokenize='trigram')`); err != nil {
		t.Fatalf("create fts5 table: %v", err)
	}
	if _, err := db.Exec(`INSERT INTO tracks_fts (rowid, title) SELECT id, title FROM tracks`); err != nil {
		t.Fatalf("populate fts5 table: %v", err)
	}

	rows, err := db.Query(`SELECT title FROM tracks_fts WHERE tracks_fts MATCH 'emian'`)
	if err != nil {
		t.Fatalf("fts5 match query: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var got []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			t.Fatalf("scan fts5 row: %v", err)
		}
		got = append(got, title)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("fts5 rows: %v", err)
	}
	if want := []string{"Bohemian Rhapsody"}; !slices.Equal(got, want) {
		t.Errorf("trigram match for %q = %v, want %v", "emian", got, want)
	}
}

// TestAdoptionSnapshotIsolation is the second adoption condition: a
// reader holding a transaction on the read pool's DSN keeps its
// snapshot across a second handle's commit, so a caller counting rows
// mid-transaction never observes a write that landed after the
// transaction began.
func TestAdoptionSnapshotIsolation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "s.db")
	precreateFile(t, path)
	const busy = 300 * time.Millisecond

	setup, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
	if err != nil {
		t.Fatalf("open setup: %v", err)
	}
	if _, err := setup.Exec(`CREATE TABLE rows_ (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := setup.Exec(`INSERT INTO rows_ DEFAULT VALUES; INSERT INTO rows_ DEFAULT VALUES`); err != nil {
		t.Fatalf("seed two rows: %v", err)
	}
	if err := setup.Close(); err != nil {
		t.Fatalf("close setup: %v", err)
	}

	reader, err := sql.Open(sqliteDriverName, readDSN(path, busy))
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	t.Cleanup(func() { _ = reader.Close() })

	ctx := context.Background()
	tx, err := reader.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin reader tx: %v", err)
	}

	var before int
	if err := tx.QueryRow(`SELECT count(*) FROM rows_`).Scan(&before); err != nil {
		t.Fatalf("count before: %v", err)
	}
	if before != 2 {
		t.Fatalf("count before = %d, want 2", before)
	}

	writer, err := sql.Open(sqliteDriverName, writeDSN(path, busy))
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	t.Cleanup(func() { _ = writer.Close() })
	if _, err := writer.Exec(`INSERT INTO rows_ DEFAULT VALUES`); err != nil {
		t.Fatalf("writer insert: %v", err)
	}

	var during int
	if err := tx.QueryRow(`SELECT count(*) FROM rows_`).Scan(&during); err != nil {
		t.Fatalf("count during: %v", err)
	}
	if during != 2 {
		t.Errorf("count during held transaction = %d, want 2 (snapshot should not see the commit)", during)
	}
	if err := tx.Rollback(); err != nil {
		t.Fatalf("rollback reader tx: %v", err)
	}

	var after int
	if err := reader.QueryRow(`SELECT count(*) FROM rows_`).Scan(&after); err != nil {
		t.Fatalf("count after: %v", err)
	}
	if after != 3 {
		t.Errorf("count after a fresh read = %d, want 3", after)
	}
}
