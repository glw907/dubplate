package importer

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
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

// testUploader is the inbox directory name every fixture below stages
// under; the tests assert their outcomes carry it back.
const testUploader = "alice"

// stagedTrack creates <root>/staging/alice/<album>/<track> and returns
// the album directory and the track's own path.
func stagedTrack(t *testing.T, root, album, track string) (albumDir, trackPath string) {
	t.Helper()
	albumDir = filepath.Join(root, "staging", testUploader, album)
	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("mkdir album: %v", err)
	}
	trackPath = filepath.Join(albumDir, track)
	if err := os.WriteFile(trackPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write track: %v", err)
	}
	return albumDir, trackPath
}

// blockPath puts a plain file where a directory is expected, so MkdirAll
// there (and therefore any move into it) fails with ENOTDIR regardless of
// the test runner's privilege level -- a permission-based fixture would
// pass trivially for a root test runner.
func blockPath(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("block %s: %v", path, err)
	}
}

// TestGateAndQuarantineFallsBackToQuarantine proves quarantineOne's
// outcome classification -- not just review.Filer's own fallback, which
// internal/review already tests -- by driving it through a real,
// unwritable ReviewDir.
func TestGateAndQuarantineFallsBackToQuarantine(t *testing.T) {
	root := t.TempDir()
	albumDir, _ := stagedTrack(t, root, "Bad Album", "track1.mp3")

	im := testImporter(root)
	blockPath(t, im.Cfg.Paths.ReviewDir)

	units := []*albumUnit{{uploader: "alice", stagedDir: albumDir}}
	if err := im.gateAndQuarantine(context.Background(), &units, "2026-08-31"); err != nil {
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
	albumDir, _ := stagedTrack(t, root, "Bad Album", "track1.mp3")

	im := testImporter(root)
	blockPath(t, im.Cfg.Paths.ReviewDir)
	blockPath(t, im.Cfg.Paths.QuarantineDir)

	units := []*albumUnit{{uploader: "alice", stagedDir: albumDir}}
	if err := im.gateAndQuarantine(context.Background(), &units, "2026-08-31"); err != nil {
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

// TestQuarantineOneEmitsShellCouldNotMoveWording pins the admin summary's
// two could-not-move lines against the retired shell layer's
// scripts/music-import: a quarantine-fallback success states where the
// album actually landed (Geoff's promised "see each item for where it
// landed"), and a complete failure states the shell's own manual-removal
// alarm rather than the misleading "excluded from import" -- the album is
// still in staging and the beet import that follows may ingest it anyway.
func TestQuarantineOneEmitsShellCouldNotMoveWording(t *testing.T) {
	tests := []struct {
		name            string
		blockReview     bool
		blockQuarantine bool
		wantSubstr      string
	}{
		{
			name:        "quarantine fallback succeeds",
			blockReview: true,
			wantSubstr:  "alice/Bad Album: could not move; excluded from import (see ",
		},
		{
			name:            "neither move succeeds",
			blockReview:     true,
			blockQuarantine: true,
			wantSubstr:      "alice/Bad Album: COULD NOT MOVE BY ANY MEANS -- still in staging and beet import may ingest it; manual removal required:",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			albumDir, _ := stagedTrack(t, root, "Bad Album", "track1.mp3")

			im := testImporter(root)
			if tt.blockReview {
				blockPath(t, im.Cfg.Paths.ReviewDir)
			}
			if tt.blockQuarantine {
				blockPath(t, im.Cfg.Paths.QuarantineDir)
			}

			u := &albumUnit{uploader: "alice", stagedDir: albumDir}
			var adminSummary strings.Builder
			im.quarantineOne(context.Background(), u, nil, "2026-08-31", &adminSummary)

			got := adminSummary.String()
			if !strings.Contains(got, tt.wantSubstr) {
				t.Errorf("admin summary = %q, want substring %q", got, tt.wantSubstr)
			}
			if n := strings.Count(got, "\n"); n != 1 {
				t.Errorf("admin summary has %d lines, want exactly 1 (no per-file detail block on this path): %q", n, got)
			}
		})
	}
}

// TestGateAndQuarantineFilesOrphanVerdict proves a quality-gate verdict
// matching no albumUnit staged this run -- a leftover album from an
// earlier or aborted run, still under StagingDir -- is filed to review
// exactly like a matched unit's failure and appended to the report,
// rather than silently skipped: staging can hold leftovers, and beet
// import scans the whole staging tree, so an unfiled leftover risks
// getting ingested by whatever run comes next.
func TestGateAndQuarantineFilesOrphanVerdict(t *testing.T) {
	root := t.TempDir()
	albumDir, _ := stagedTrack(t, root, "Leftover Album", "track1.mp3")

	im := testImporter(root)
	im.Cfg.Paths.InboxDir = filepath.Join(root, "inbox")

	var units []*albumUnit // this run staged nothing: the album is a leftover
	if err := im.gateAndQuarantine(context.Background(), &units, "2026-08-31"); err != nil {
		t.Fatalf("gateAndQuarantine() error = %v", err)
	}

	if len(units) != 1 {
		t.Fatalf("units = %d entries, want 1 appended for the orphan album", len(units))
	}
	if units[0].uploader != "alice" {
		t.Errorf("uploader = %q, want %q", units[0].uploader, "alice")
	}
	if units[0].result != OutcomeRejected {
		t.Errorf("result = %q, want %q", units[0].result, OutcomeRejected)
	}
	if _, err := os.Stat(albumDir); !os.IsNotExist(err) {
		t.Errorf("orphan album still in staging after gateAndQuarantine: err = %v", err)
	}
}

// TestFileSkipsFilesOrphanSkip proves a beets skip matching no pending
// albumUnit -- a leftover album this run never staged -- is filed to
// review and appended to the report the same way, mirroring
// TestGateAndQuarantineFilesOrphanVerdict for the beets-skip call site.
func TestFileSkipsFilesOrphanSkip(t *testing.T) {
	root := t.TempDir()
	_, trackPath := stagedTrack(t, root, "Leftover Album", "track1.flac")

	im := testImporter(root)
	im.Cfg.Paths.InboxDir = filepath.Join(root, "inbox")

	var units []*albumUnit // nothing pending this run: the album is a leftover
	skipUsers := im.fileSkips(&units, nil, []string{trackPath}, "2026-08-31")

	if !skipUsers["alice"] {
		t.Errorf("skipUsers = %v, want alice present", skipUsers)
	}
	if len(units) != 1 {
		t.Fatalf("units = %d entries, want 1 appended for the orphan skip", len(units))
	}
	if units[0].uploader != "alice" {
		t.Errorf("uploader = %q, want %q", units[0].uploader, "alice")
	}
	if units[0].result != OutcomeReview {
		t.Errorf("result = %q, want %q", units[0].result, OutcomeReview)
	}
	if _, err := os.Stat(trackPath); !os.IsNotExist(err) {
		t.Errorf("orphan skip still in staging after fileSkips: err = %v", err)
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
