package review

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// frozenNow matches the shell suite's stub_date, which always prints
// "120000" (12:00:00) so the HHMMSS collision suffix is a literal the
// test never has to recompute.
func frozenNow() time.Time {
	return time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
}

// stagedAlbum creates <root>/staging/<album>/<track> and returns the
// album directory, the fixture every MoveToReview test starts from.
func stagedAlbum(t *testing.T, root, album, track string) string {
	t.Helper()
	dir := filepath.Join(root, "staging", album)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, track), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// blockedDir writes a plain file at <root>/<name> and returns its path.
// A file where a directory is expected blocks MkdirAll deterministically
// (ENOTDIR), regardless of the test runner's privilege level -- a
// permission-based fixture would pass trivially for a root test runner.
func blockedDir(t *testing.T, root, name string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.WriteFile(path, []byte("not a directory"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestUniqueDestPath(t *testing.T) {
	tests := []struct {
		name  string
		setup func(dir string)
		want  string
	}{
		{
			name:  "nothing there yet",
			setup: func(dir string) {},
			want:  "Bad Album",
		},
		{
			name: "collision gets a time suffix",
			setup: func(dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "Bad Album"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			want: "Bad Album-120000",
		},
		{
			name: "time suffix also collides, falls back to numeric",
			setup: func(dir string) {
				if err := os.MkdirAll(filepath.Join(dir, "Bad Album"), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(dir, "Bad Album-120000"), 0o755); err != nil {
					t.Fatal(err)
				}
			},
			want: "Bad Album-120000.2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			tt.setup(dir)
			f := &Filer{Now: frozenNow}

			got := f.uniqueDestPath(dir, "Bad Album")
			want := filepath.Join(dir, tt.want)
			if got != want {
				t.Errorf("uniqueDestPath() = %q, want %q", got, want)
			}
		})
	}
}

func TestFilerMoveToReview_intoEmptyDestination(t *testing.T) {
	root := t.TempDir()
	staging := stagedAlbum(t, root, "Good Album", "one.flac")

	f := &Filer{ReviewDir: filepath.Join(root, "review"), QuarantineDir: filepath.Join(root, "quarantine"), Now: frozenNow}

	dest, err := f.MoveToReview(staging)
	if err != nil {
		t.Fatalf("MoveToReview() error = %v", err)
	}

	want := filepath.Join(root, "review", "Good Album")
	if dest != want {
		t.Errorf("dest = %q, want %q", dest, want)
	}
	if _, err := os.Stat(filepath.Join(dest, "one.flac")); err != nil {
		t.Errorf("moved file missing: %v", err)
	}
	if _, err := os.Stat(staging); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("source still exists at %s after a successful move", staging)
	}
}

func TestFilerMoveToReview_disambiguatesOnCollision(t *testing.T) {
	root := t.TempDir()
	staging := stagedAlbum(t, root, "Bad Album", "new.flac")

	existing := filepath.Join(root, "review", "Bad Album")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(existing, "keep.flac"), []byte("first"), 0o644); err != nil {
		t.Fatal(err)
	}

	f := &Filer{ReviewDir: filepath.Join(root, "review"), QuarantineDir: filepath.Join(root, "quarantine"), Now: frozenNow}

	dest, err := f.MoveToReview(staging)
	if err != nil {
		t.Fatalf("MoveToReview() error = %v", err)
	}
	if dest == existing {
		t.Fatalf("dest %q collided with the pre-existing directory instead of disambiguating", dest)
	}
	if _, err := os.Stat(filepath.Join(existing, "keep.flac")); err != nil {
		t.Errorf("pre-existing album was clobbered: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dest, "new.flac")); err != nil {
		t.Errorf("new album missing at disambiguated dest: %v", err)
	}
}

// The shell layer's fallback quarantine path (music-import's
// quarantine_quality_failures) was tested only for its primary
// safe_mv_to_review failure; nothing ever forced the second
// safe_mv_to_review call to fail too. Both cases get their first real
// coverage here.

func TestFilerMoveToReview_fallsBackToQuarantine(t *testing.T) {
	root := t.TempDir()
	staging := stagedAlbum(t, root, "Bad Album", "track1.flac")
	blockedReview := blockedDir(t, root, "review-blocked")

	f := &Filer{ReviewDir: blockedReview, QuarantineDir: filepath.Join(root, "quarantine"), Now: frozenNow}

	dest, err := f.MoveToReview(staging)
	if err != nil {
		t.Fatalf("MoveToReview() error = %v", err)
	}
	if filepath.Dir(dest) != f.QuarantineDir {
		t.Errorf("dest = %q, want a child of QuarantineDir %q", dest, f.QuarantineDir)
	}
	if _, err := os.Stat(filepath.Join(dest, "track1.flac")); err != nil {
		t.Errorf("quarantined file missing: %v", err)
	}
	if _, err := os.Stat(staging); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("source still exists at %s after landing in quarantine", staging)
	}
}

func TestFilerMoveToReview_couldNotMoveByAnyMeans(t *testing.T) {
	root := t.TempDir()
	staging := stagedAlbum(t, root, "Bad Album", "track1.flac")
	blockedReview := blockedDir(t, root, "review-blocked")
	blockedQuarantine := blockedDir(t, root, "quarantine-blocked")

	f := &Filer{ReviewDir: blockedReview, QuarantineDir: blockedQuarantine, Now: frozenNow}

	dest, err := f.MoveToReview(staging)
	if err == nil {
		t.Fatalf("MoveToReview() = %q, nil; want an error when both destinations are unusable", dest)
	}
	if !errors.Is(err, ErrCouldNotMove) {
		t.Errorf("error = %v, want it to wrap ErrCouldNotMove", err)
	}
	if dest != "" {
		t.Errorf("dest = %q, want empty on total failure", dest)
	}
	if _, err := os.Stat(staging); err != nil {
		t.Errorf("source no longer at %s after a total move failure: %v", staging, err)
	}
}

func TestFilerSweep(t *testing.T) {
	root := t.TempDir()
	reviewDir := filepath.Join(root, "review")
	quarantineDir := filepath.Join(root, "quarantine")

	oldReview := filepath.Join(reviewDir, "2020-01-01")
	freshReview := filepath.Join(reviewDir, "2026-08-31")
	oldQuarantine := filepath.Join(quarantineDir, "2020-01-01")
	freshQuarantine := filepath.Join(quarantineDir, "2026-08-31")
	for _, dir := range []string{oldReview, freshReview, oldQuarantine, freshQuarantine} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}

	old := frozenNow().AddDate(0, 0, -90)
	for _, dir := range []string{oldReview, oldQuarantine} {
		if err := os.Chtimes(dir, old, old); err != nil {
			t.Fatal(err)
		}
	}

	f := &Filer{ReviewDir: reviewDir, QuarantineDir: quarantineDir, Now: frozenNow}
	if err := f.Sweep(60 * 24 * time.Hour); err != nil {
		t.Fatalf("Sweep() error = %v", err)
	}

	for _, dir := range []string{oldReview, oldQuarantine} {
		if _, err := os.Stat(dir); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("%s survived the sweep past its retention window", dir)
		}
	}
	for _, dir := range []string{freshReview, freshQuarantine} {
		if _, err := os.Stat(dir); err != nil {
			t.Errorf("%s was swept despite being within retention: %v", dir, err)
		}
	}
}

func TestFilerSweep_missingQuarantineDirIsNotAnError(t *testing.T) {
	root := t.TempDir()
	f := &Filer{
		ReviewDir:     filepath.Join(root, "review"),
		QuarantineDir: filepath.Join(root, "quarantine-never-created"),
		Now:           frozenNow,
	}
	if err := f.Sweep(60 * 24 * time.Hour); err != nil {
		t.Fatalf("Sweep() error = %v, want nil for a QuarantineDir that was never created", err)
	}
}
