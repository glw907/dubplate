package store

import (
	"context"
	"testing"
)

// The ids below are migration 0006's fixture (testdata/fixtures/0006_intake.sql).
const (
	intakeFixtureSubmission = "x'c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8c8'"
	intakeFixtureAdmin      = "x'a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8a8'"
)

// insertCollection inserts a server-written group collection owned by
// the fixture's admin, with root and route as given (route NULL when
// nil), and returns the insert's error.
func insertCollection(h *Handle, root string, route any) error {
	_, err := h.write.ExecContext(context.Background(),
		`INSERT INTO collections (id, name, root_path, owner, visibility, writer, intake_route)
		 VALUES (?, 'Probe', ?, `+intakeFixtureAdmin+`, 'group', 'server', ?)`,
		NewID(), []byte(root), route)
	return err
}

func TestIntakeSchemaRouteIndex(t *testing.T) {
	t.Run("two collections with no route coexist", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		for _, root := range []string{"null-one", "null-two"} {
			if err := insertCollection(h, root, nil); err != nil {
				t.Fatalf("insert %s with a NULL route: %v", root, err)
			}
		}
	})

	t.Run("a second default is refused", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		// The fixture already holds the default route's collection.
		err := insertCollection(h, "second-default", "default")
		assertErrNames(t, err, "UNIQUE constraint failed: collections.intake_route")
		if code := sqliteCode(err); code != sqliteUniqueCode {
			t.Errorf("code = %d, want %d", code, sqliteUniqueCode)
		}
	})

	t.Run("one unlisted collection lands and a second is refused", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		if err := insertCollection(h, "unlisted-one", "unlisted"); err != nil {
			t.Fatalf("first unlisted: %v", err)
		}
		assertErrNames(t, insertCollection(h, "unlisted-two", "unlisted"),
			"UNIQUE constraint failed: collections.intake_route")
	})

	t.Run("a route outside the two is refused", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		assertErrNames(t, insertCollection(h, "odd-route", "private"), "CHECK constraint failed")
	})
}

func TestIntakeSchemaSubmissionFiles(t *testing.T) {
	ctx := context.Background()
	insertFile := func(h *Handle, relPath string) error {
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO submission_files (id, submission_id, rel_path, size, role)
			 VALUES (?, `+intakeFixtureSubmission+`, ?, 10, 'audio')`,
			NewID(), []byte(relPath))
		return err
	}

	t.Run("a second rel_path in one submission is refused", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		// The fixture's submission already holds "album/01 Track.flac".
		err := insertFile(h, "album/01 Track.flac")
		assertErrNames(t, err, "UNIQUE constraint failed: submission_files.submission_id, submission_files.rel_path")
		if code := sqliteCode(err); code != sqliteUniqueCode {
			t.Errorf("code = %d, want %d", code, sqliteUniqueCode)
		}
	})

	t.Run("a distinct rel_path lands", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		if err := insertFile(h, "album/02 Track.flac"); err != nil {
			t.Fatalf("insert a distinct path: %v", err)
		}
	})

	t.Run("a digest of the wrong length is refused", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		for _, column := range []string{"audio_key", "header_md5", "file_hash"} {
			_, err := h.write.ExecContext(ctx,
				`UPDATE submission_files SET `+column+` = x'0102' WHERE submission_id = `+intakeFixtureSubmission)
			assertErrNames(t, err, "CHECK constraint failed")
		}
		_, err := h.write.ExecContext(ctx,
			`UPDATE submissions SET album_key = x'0102' WHERE id = `+intakeFixtureSubmission)
		assertErrNames(t, err, "CHECK constraint failed")
	})

	t.Run("a digest of the right length lands", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		if _, err := h.write.ExecContext(ctx,
			`UPDATE submission_files SET audio_key = randomblob(16), header_md5 = randomblob(16), file_hash = randomblob(32)
			 WHERE submission_id = `+intakeFixtureSubmission); err != nil {
			t.Fatalf("file digests: %v", err)
		}
		if _, err := h.write.ExecContext(ctx,
			`UPDATE submissions SET album_key = randomblob(32) WHERE id = `+intakeFixtureSubmission); err != nil {
			t.Fatalf("album key: %v", err)
		}
	})
}

func TestIntakeSchemaEventSeq(t *testing.T) {
	ctx := context.Background()
	insertEvent := func(h *Handle, seq int, kind string) error {
		_, err := h.write.ExecContext(ctx,
			`INSERT INTO submission_events (id, submission_id, seq, at, kind)
			 VALUES (?, `+intakeFixtureSubmission+`, ?, 1700000000000000, ?)`,
			NewID(), seq, kind)
		return err
	}

	t.Run("a second event at one seq is refused", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		// The fixture's submission already holds events at seq 1 and 2.
		err := insertEvent(h, 2, "stalled")
		assertErrNames(t, err, "UNIQUE constraint failed: submission_events.submission_id, submission_events.seq")
		if code := sqliteCode(err); code != sqliteUniqueCode {
			t.Errorf("code = %d, want %d", code, sqliteUniqueCode)
		}
	})

	t.Run("the next seq lands, and kind carries no CHECK", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		if err := insertEvent(h, 3, "a-state-rung-9-adds"); err != nil {
			t.Fatalf("insert at seq 3: %v", err)
		}
	})

	t.Run("seq as the largest plus one, in one statement", func(t *testing.T) {
		_, h := fullyMigratedStore(t)
		for range 2 {
			if _, err := h.write.ExecContext(ctx,
				`INSERT INTO submission_events (id, submission_id, seq, at, kind)
				 SELECT ?, id, COALESCE((SELECT MAX(seq) FROM submission_events WHERE submission_id = submissions.id), 0) + 1,
				        1700000000000000, 'stalled'
				 FROM submissions WHERE id = `+intakeFixtureSubmission,
				NewID()); err != nil {
				t.Fatalf("insert at max+1: %v", err)
			}
		}
		var largest int
		if err := h.write.QueryRowContext(ctx,
			`SELECT MAX(seq) FROM submission_events WHERE submission_id = `+intakeFixtureSubmission).Scan(&largest); err != nil {
			t.Fatal(err)
		}
		if largest != 4 {
			t.Errorf("largest seq = %d, want 4 (the fixture's 1 and 2, then two inserts)", largest)
		}
	})
}

// TestIntakeSchemaRelPathChecks pins that submission_files.rel_path
// carries 0003_catalog.sql's relative-path CHECK: each unsafe path is
// refused by the database. The backslash is the exception, which SQL
// admits and the intake package's path validator refuses.
func TestIntakeSchemaRelPathChecks(t *testing.T) {
	_, h := fullyMigratedStore(t)
	for _, tc := range []struct {
		name string
		path []byte
	}{
		{"absolute", []byte("/a")},
		{"empty component", []byte("a//b")},
		{"dot component", []byte("a/./b")},
		{"dotdot component", []byte("a/../b")},
		{"dotdot", []byte("..")},
		{"NUL", []byte("a\x00b")},
		{"empty", []byte{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.write.ExecContext(context.Background(),
				`INSERT INTO submission_files (id, submission_id, rel_path, size, role)
				 VALUES (?, `+intakeFixtureSubmission+`, ?, 1, 'extra')`,
				NewID(), tc.path)
			assertErrNames(t, err, "CHECK constraint failed")
		})
	}
}
