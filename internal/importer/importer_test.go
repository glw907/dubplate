package importer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/notify"
)

func TestIsArchiveName(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"Album.zip", true},
		{"Album.ZIP", true},
		{"Album.tar", true},
		{"Album.tar.gz", true},
		{"Album.tgz", true},
		{"Album.tar.bz2", true},
		{"Album.tbz2", true},
		{"track.flac", false},
		{"cover.jpg", false},
		{"Album", false},
	}
	for _, tt := range tests {
		if got := isArchiveName(tt.name); got != tt.want {
			t.Errorf("isArchiveName(%q) = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestStagingDestPath(t *testing.T) {
	dir := t.TempDir()

	first := stagingDestPath(dir, "Album")
	if first != filepath.Join(dir, "Album") {
		t.Fatalf("first stagingDestPath = %q, want plain join", first)
	}
	if err := os.Mkdir(first, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	second := stagingDestPath(dir, "Album")
	if second != filepath.Join(dir, "Album.2") {
		t.Errorf("second stagingDestPath = %q, want %q", second, filepath.Join(dir, "Album.2"))
	}
	if err := os.Mkdir(second, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	third := stagingDestPath(dir, "Album")
	if third != filepath.Join(dir, "Album.3") {
		t.Errorf("third stagingDestPath = %q, want %q", third, filepath.Join(dir, "Album.3"))
	}

	// An extension is preserved on the disambiguated name, not swallowed
	// into it (a loose file entry, not a directory).
	if err := os.WriteFile(filepath.Join(dir, "track.flac"), nil, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	got := stagingDestPath(dir, "track.flac")
	if want := filepath.Join(dir, "track.2.flac"); got != want {
		t.Errorf("stagingDestPath with extension = %q, want %q", got, want)
	}
}

func TestDirHasEntries(t *testing.T) {
	empty := t.TempDir()
	has, err := dirHasEntries(empty)
	if err != nil {
		t.Fatalf("dirHasEntries(empty): %v", err)
	}
	if has {
		t.Error("dirHasEntries(empty) = true, want false")
	}

	nonEmpty := t.TempDir()
	if err := os.MkdirAll(filepath.Join(nonEmpty, "sub", "deep"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	has, err = dirHasEntries(nonEmpty)
	if err != nil {
		t.Fatalf("dirHasEntries(nonEmpty): %v", err)
	}
	if !has {
		t.Error("dirHasEntries(nonEmpty) = false, want true")
	}
}

func TestQuiescentSince(t *testing.T) {
	dir := t.TempDir()
	old := filepath.Join(dir, "old.flac")
	if err := os.WriteFile(old, []byte("x"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	past := time.Now().Add(-time.Hour)
	if err := os.Chtimes(old, past, past); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	cutoff := time.Now().Add(-10 * time.Minute)
	quiescent, err := quiescentSince(dir, cutoff)
	if err != nil {
		t.Fatalf("quiescentSince: %v", err)
	}
	if !quiescent {
		t.Error("quiescentSince with only an old file = false, want true")
	}

	fresh := filepath.Join(dir, "fresh.flac")
	if err := os.WriteFile(fresh, []byte("y"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	quiescent, err = quiescentSince(dir, cutoff)
	if err != nil {
		t.Fatalf("quiescentSince: %v", err)
	}
	if quiescent {
		t.Error("quiescentSince with a fresh file = true, want false")
	}
}

func TestDirIsEmptyOrGone(t *testing.T) {
	tests := []struct {
		name  string
		setup func(t *testing.T) string
		want  bool
	}{
		{
			name: "does not exist",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "gone")
			},
			want: true,
		},
		{
			name: "empty directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			want: true,
		},
		{
			name: "directory with content",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.WriteFile(filepath.Join(dir, "leftover.flac"), nil, 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
				return dir
			},
			want: false,
		},
		{
			name: "loose file still present",
			setup: func(t *testing.T) string {
				path := filepath.Join(t.TempDir(), "track.flac")
				if err := os.WriteFile(path, nil, 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
				return path
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.setup(t)
			got, err := dirIsEmptyOrGone(path)
			if err != nil {
				t.Fatalf("dirIsEmptyOrGone(%s): %v", path, err)
			}
			if got != tt.want {
				t.Errorf("dirIsEmptyOrGone(%s) = %v, want %v", path, got, tt.want)
			}
		})
	}
}

func TestCleanEmptyStagingDirsCascades(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "alice", "Empty Album")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	keep := filepath.Join(root, "bob", "Still Has A File")
	if err := os.MkdirAll(keep, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(keep, "track.flac"), nil, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}

	if err := cleanEmptyStagingDirs(root); err != nil {
		t.Fatalf("cleanEmptyStagingDirs: %v", err)
	}

	// alice's directory tree was empty all the way up, so both levels --
	// and alice's own now-empty top-level directory -- are gone.
	if _, err := os.Stat(filepath.Join(root, "alice")); !os.IsNotExist(err) {
		t.Errorf("alice's empty tree survived cleanup: err = %v", err)
	}
	// bob's non-empty leaf keeps its whole ancestor chain, root included.
	if _, err := os.Stat(keep); err != nil {
		t.Errorf("bob's non-empty directory was removed: %v", err)
	}
	if _, err := os.Stat(root); err != nil {
		t.Errorf("root itself was removed: %v", err)
	}
}

func TestFindUnit(t *testing.T) {
	units := []*albumUnit{
		{uploader: "alice", stagedDir: "/staging/alice/Album A"},
		{uploader: "bob", stagedDir: "/staging/bob/track.flac"},
	}

	tests := []struct {
		path string
		want *albumUnit
	}{
		{"/staging/alice/Album A", units[0]},
		{"/staging/alice/Album A/01.flac", units[0]},
		{"/staging/bob/track.flac", units[1]},
		{"/staging/carol/Album C", nil},
		// A path that merely shares a prefix string, without the
		// separator, must not false-match.
		{"/staging/alice/Album A Extended/01.flac", nil},
	}
	for _, tt := range tests {
		if got := findUnit(units, tt.path); got != tt.want {
			t.Errorf("findUnit(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestPendingUnits(t *testing.T) {
	resolved := &albumUnit{result: OutcomeRejected}
	pendingA := &albumUnit{}
	pendingB := &albumUnit{}
	units := []*albumUnit{resolved, pendingA, pendingB}

	got := pendingUnits(units)
	if len(got) != 2 || got[0] != pendingA || got[1] != pendingB {
		t.Errorf("pendingUnits() = %v, want [pendingA, pendingB]", got)
	}
}

// testImporter builds an [Importer] whose Mailer points at a definitely
// refused local address, so a Send attempt fails fast and is logged and
// swallowed rather than hanging the test -- gateAndQuarantine's own
// best-effort contract is what these tests exercise, not delivery.
func testImporter(root string) *Importer {
	return &Importer{
		Cfg: &config.Config{
			AdminEmail: "admin@example.com",
			Paths: config.Paths{
				StagingDir:    filepath.Join(root, "staging"),
				ReviewDir:     filepath.Join(root, "review"),
				QuarantineDir: filepath.Join(root, "quarantine"),
			},
		},
		Roster: config.Roster{},
		Mailer: notify.New(config.SMTPConfig{Host: "127.0.0.1", Port: 1}),
		Now:    time.Now,
	}
}

// TestGateAndQuarantineFallsBackToQuarantine proves quarantineOne's
// outcome classification -- not just review.Filer's own fallback, which
// internal/review already tests -- by driving it through a real,
// unwritable ReviewDir.
func TestGateAndQuarantineFallsBackToQuarantine(t *testing.T) {
	root := t.TempDir()
	albumDir := filepath.Join(root, "staging", "alice", "Bad Album")
	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("mkdir album: %v", err)
	}
	if err := os.WriteFile(filepath.Join(albumDir, "track1.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write track: %v", err)
	}

	im := testImporter(root)
	// ReviewDir exists as a plain file, so MkdirAll (and therefore the
	// primary move) can never succeed there.
	if err := os.WriteFile(im.Cfg.Paths.ReviewDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("block review dir: %v", err)
	}

	units := []*albumUnit{{uploader: "alice", stagedDir: albumDir}}
	if err := im.gateAndQuarantine(context.Background(), units, "2026-08-31"); err != nil {
		t.Fatalf("gateAndQuarantine() error = %v", err)
	}

	if units[0].result != OutcomeQuarantined {
		t.Errorf("result = %q, want %q", units[0].result, OutcomeQuarantined)
	}
	if units[0].errMsg != "" {
		t.Errorf("errMsg = %q, want empty: the quarantine fallback succeeded", units[0].errMsg)
	}
	if _, err := os.Stat(albumDir); !os.IsNotExist(err) {
		t.Errorf("album still in staging after a successful quarantine fallback: err = %v", err)
	}
}

// TestGateAndQuarantineCouldNotMoveByAnyMeans proves the shell layer's
// once-untested "COULD NOT MOVE BY ANY MEANS" branch: when both the
// review and quarantine destinations are unusable, the album is left in
// place, and the outcome still records it as quarantined with the
// failure noted, rather than silently vanishing from the report.
func TestGateAndQuarantineCouldNotMoveByAnyMeans(t *testing.T) {
	root := t.TempDir()
	albumDir := filepath.Join(root, "staging", "alice", "Bad Album")
	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("mkdir album: %v", err)
	}
	if err := os.WriteFile(filepath.Join(albumDir, "track1.mp3"), []byte("x"), 0o644); err != nil {
		t.Fatalf("write track: %v", err)
	}

	im := testImporter(root)
	if err := os.WriteFile(im.Cfg.Paths.ReviewDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("block review dir: %v", err)
	}
	if err := os.WriteFile(im.Cfg.Paths.QuarantineDir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("block quarantine dir: %v", err)
	}

	units := []*albumUnit{{uploader: "alice", stagedDir: albumDir}}
	if err := im.gateAndQuarantine(context.Background(), units, "2026-08-31"); err != nil {
		t.Fatalf("gateAndQuarantine() error = %v", err)
	}

	if units[0].result != OutcomeQuarantined {
		t.Errorf("result = %q, want %q", units[0].result, OutcomeQuarantined)
	}
	if units[0].errMsg == "" {
		t.Error("errMsg is empty, want the could-not-move failure recorded")
	}
	if _, err := os.Stat(filepath.Join(albumDir, "track1.mp3")); err != nil {
		t.Errorf("album no longer in staging despite both moves failing: %v", err)
	}
}

// TestRunReportJSONShape pins the run report's on-disk contract: a
// consumer (the uploader pass) reads these exact field names and the
// version number, so a rename or a dropped omitempty is a breaking
// change this test would catch.
func TestRunReportJSONShape(t *testing.T) {
	report := RunReport{
		Version:   ReportVersion,
		StartedAt: time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC),
		EndedAt:   time.Date(2026, 8, 31, 12, 5, 0, 0, time.UTC),
		Albums: []AlbumOutcome{
			{
				Uploader:     "alice",
				SourceDir:    "/srv/music/inbox/alice",
				Outcome:      OutcomeImported,
				LibraryPaths: []string{"/srv/music/library/Alice/Album/01.flac"},
			},
		},
	}

	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, field := range []string{"version", "started_at", "ended_at", "albums"} {
		if _, ok := decoded[field]; !ok {
			t.Errorf("run report JSON missing field %q: %s", field, data)
		}
	}
	if held, ok := decoded["lock_held"].(bool); !ok || held {
		t.Errorf("lock_held = %v, want the field present and false on a clean run: %s", decoded["lock_held"], data)
	}
	if _, ok := decoded["error"]; ok {
		t.Errorf("error present with its zero value; want omitempty to drop it: %s", data)
	}

	var roundTrip RunReport
	if err := json.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("round-trip unmarshal: %v", err)
	}
	if roundTrip.Albums[0].Outcome != OutcomeImported {
		t.Errorf("round-trip Outcome = %q, want %q", roundTrip.Albums[0].Outcome, OutcomeImported)
	}
}

// TestRunReportHeldLockOmitsAlbums proves a held-lock report -- the
// report written on the exit path that touches nothing else -- still
// carries the fields a monitor needs to tell it apart from a crash.
func TestRunReportHeldLockOmitsAlbums(t *testing.T) {
	report := RunReport{Version: ReportVersion, LockHeld: true}
	data, err := json.Marshal(report)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if held, ok := decoded["lock_held"].(bool); !ok || !held {
		t.Errorf("lock_held = %v, want true: %s", decoded["lock_held"], data)
	}
	if _, ok := decoded["albums"]; !ok {
		t.Errorf("albums field missing on a held-lock report: %s", data)
	}
}
