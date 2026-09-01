package importer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/beets"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/lock"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/ping"
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

// testUploaderEmail is testUploader's verified email in every
// session.json sidecar fixture below.
const testUploaderEmail = "alice@example.com"

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

// --- U8: session mode, terminal ledger events, per-run reports ---------

// testImporterForRun extends testImporter with the paths and
// collaborators [Importer.Run] and [Importer.ImportSessionLocked]
// exercise beyond gateAndQuarantine/fileSkips alone: a state directory
// for the lock and report files, an inbox directory, and a live (but
// harmless -- an empty PingURLs map makes every ping a no-op) Ping
// client the healthy exit paths call unconditionally.
func testImporterForRun(root string) *Importer {
	im := testImporter(root)
	im.Cfg.Paths.InboxDir = filepath.Join(root, "inbox")
	im.Cfg.Paths.StateDir = filepath.Join(root, "state")
	im.Ping = ping.NewClient()
	return im
}

// sidecarJSON returns session.json's exact on-disk bytes, matching
// internal/serve's sessionSidecar shape field for field -- the importer
// package cannot import internal/serve (an import cycle: serve already
// imports nothing from importer, but importer's tests pin the contract
// from this side rather than share a type across the package
// boundary), so this is the shared contract's other half.
func sidecarJSON(t *testing.T, sessionID, email, dirName, album, collection string) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{
		"version":    1,
		"session_id": sessionID,
		"email":      email,
		"dir_name":   dirName,
		"album":      album,
		"collection": collection,
		"created_at": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		t.Fatalf("marshal sidecar: %v", err)
	}
	return data
}

// stagePromotedSession creates
// <inboxDir>/<testUploader>/<sessionID>__<album>/<track> plus its
// session.json sidecar, matching exactly what [Engine.promote]
// (internal/serve) leaves behind after a real upload's atomic
// promotion, and returns the session directory's path. Every call site
// in this file wants the same testUploader dirName and
// testUploaderEmail, so neither is a parameter.
func stagePromotedSession(t *testing.T, inboxDir, sessionID, album, track, collection string) string {
	t.Helper()
	dir := filepath.Join(inboxDir, testUploader, sessionID+"__"+album)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir session dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, track), []byte("x"), 0o644); err != nil {
		t.Fatalf("write track: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, sessionSidecarName), sidecarJSON(t, sessionID, testUploaderEmail, testUploader, album, collection), 0o644); err != nil {
		t.Fatalf("write sidecar: %v", err)
	}
	return dir
}

func TestFindSessionDir(t *testing.T) {
	root := t.TempDir()
	want := stagePromotedSession(t, root, "sess123", "Album A", "track.flac", "")

	got, err := findSessionDir(root, "sess123")
	if err != nil {
		t.Fatalf("findSessionDir: %v", err)
	}
	if got != want {
		t.Errorf("findSessionDir = %q, want %q", got, want)
	}

	if _, err := findSessionDir(root, "no-such-session"); !errors.Is(err, ErrSessionNotFound) {
		t.Errorf("findSessionDir(unknown) error = %v, want ErrSessionNotFound", err)
	}
}

func TestReadAndConsumeSidecar(t *testing.T) {
	t.Run("present", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, sessionSidecarName), sidecarJSON(t, "sess1", "a@example.com", "alice", "Album", "vinyl"), 0o644); err != nil {
			t.Fatalf("write sidecar: %v", err)
		}
		sc, ok, err := readAndConsumeSidecar(dir)
		if err != nil {
			t.Fatalf("readAndConsumeSidecar: %v", err)
		}
		if !ok {
			t.Fatal("ok = false, want true")
		}
		if sc.SessionID != "sess1" || sc.Email != "a@example.com" || sc.DirName != "alice" || sc.Collection != "vinyl" {
			t.Errorf("sidecar = %+v, want the written fields", sc)
		}
		if _, err := os.Stat(filepath.Join(dir, sessionSidecarName)); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("sidecar file survived readAndConsumeSidecar: err = %v", err)
		}
	})

	t.Run("absent is not an error", func(t *testing.T) {
		_, ok, err := readAndConsumeSidecar(t.TempDir())
		if err != nil {
			t.Fatalf("readAndConsumeSidecar: %v", err)
		}
		if ok {
			t.Error("ok = true for a directory with no sidecar, want false")
		}
	})

	t.Run("malformed json errors", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, sessionSidecarName), []byte("not json"), 0o644); err != nil {
			t.Fatalf("write sidecar: %v", err)
		}
		if _, _, err := readAndConsumeSidecar(dir); err == nil {
			t.Error("readAndConsumeSidecar(malformed) error = nil, want an error")
		}
	})
}

func TestStageSession(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)
	srcDir := stagePromotedSession(t, im.Cfg.Paths.InboxDir, "sess1", "Album A", "track.flac", "vinyl")

	u, err := im.stageSession(srcDir, "sess1")
	if err != nil {
		t.Fatalf("stageSession: %v", err)
	}
	if u.uploader != "alice" || u.sessionID != "sess1" || u.email != "alice@example.com" || u.collection != "vinyl" {
		t.Errorf("unit = %+v, want the sidecar's fields", u)
	}
	wantStagedDir := filepath.Join(im.Cfg.Paths.StagingDir, "alice", "sess1__Album A")
	if u.stagedDir != wantStagedDir {
		t.Errorf("stagedDir = %q, want %q", u.stagedDir, wantStagedDir)
	}
	if _, err := os.Stat(filepath.Join(u.stagedDir, "track.flac")); err != nil {
		t.Errorf("track not found under staged dir: %v", err)
	}
	if _, err := os.Stat(srcDir); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("source dir survived staging: err = %v", err)
	}

	if _, err := im.stageSession(t.TempDir(), "sess2"); !errors.Is(err, ErrMissingSidecar) {
		t.Errorf("stageSession(no sidecar) error = %v, want ErrMissingSidecar", err)
	}
}

// TestStageSessionFailedMoveLeavesSessionRetryable proves that a failed
// review.Move never destroys the sidecar first: srcDir's session.json
// must still be there, and readable, so a later retry of the same
// session ID can succeed once whatever blocked the move is cleared.
// Consuming the sidecar before the move (the pre-fix behavior) would
// leave srcDir stranded in the inbox with ErrMissingSidecar on every
// retry, forever.
func TestStageSessionFailedMoveLeavesSessionRetryable(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)
	srcDir := stagePromotedSession(t, im.Cfg.Paths.InboxDir, "sess1", "Album A", "track.flac", "vinyl")

	// Pre-create the destination stageSession will try to rename srcDir
	// onto, as a plain file rather than a directory, so os.Rename fails
	// for a reason other than EXDEV and review.Move returns the bare
	// error instead of falling back to copy-and-remove.
	destDir := filepath.Join(im.Cfg.Paths.StagingDir, "alice")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		t.Fatalf("mkdir staging dir: %v", err)
	}
	blocker := filepath.Join(destDir, filepath.Base(srcDir))
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatalf("write blocking file: %v", err)
	}

	if _, err := im.stageSession(srcDir, "sess1"); err == nil {
		t.Fatal("stageSession(blocked move) error = nil, want the move failure")
	}

	sc, ok, err := readSidecar(srcDir)
	if err != nil {
		t.Fatalf("readSidecar after failed move: %v", err)
	}
	if !ok {
		t.Fatal("session.json missing after failed move, want it intact for retry")
	}
	if sc.SessionID != "sess1" || sc.DirName != "alice" {
		t.Errorf("sidecar after failed move = %+v, want the original fields", sc)
	}
	if _, err := os.Stat(filepath.Join(srcDir, "track.flac")); err != nil {
		t.Errorf("track missing from srcDir after failed move: %v", err)
	}

	// Clear the blocker and retry: the same call must now succeed,
	// proving srcDir was left in a retryable state.
	if err := os.Remove(blocker); err != nil {
		t.Fatalf("remove blocker: %v", err)
	}
	u, err := im.stageSession(srcDir, "sess1")
	if err != nil {
		t.Fatalf("stageSession(retry after clearing blocker): %v", err)
	}
	if u.sessionID != "sess1" {
		t.Errorf("retried unit sessionID = %q, want sess1", u.sessionID)
	}
	if _, err := os.Stat(filepath.Join(u.stagedDir, sessionSidecarName)); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("sidecar survived in staged dir after a successful move: err = %v", err)
	}
}

func TestAlbumFromStagedDir(t *testing.T) {
	tests := []struct {
		stagedDir string
		want      string
	}{
		{"/staging/alice/sess123__Album A", "Album A"},
		{"/staging/alice/sess123__Album__With__Dashes", "Album__With__Dashes"},
		{"/staging/alice/no-separator", "no-separator"},
	}
	for _, tt := range tests {
		if got := albumFromStagedDir(tt.stagedDir); got != tt.want {
			t.Errorf("albumFromStagedDir(%q) = %q, want %q", tt.stagedDir, got, tt.want)
		}
	}
}

func TestSweepReports(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)
	if err := os.MkdirAll(im.Cfg.Paths.StateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	// Five report files, each with a distinct, increasing mtime so
	// sweepReports' oldest-first deletion order is unambiguous.
	var paths []string
	base := time.Now().Add(-time.Hour)
	for i := range 5 {
		path := im.reportPath(fmt.Sprintf("run%d", i))
		if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
			t.Fatalf("write report %d: %v", i, err)
		}
		mtime := base.Add(time.Duration(i) * time.Minute)
		if err := os.Chtimes(path, mtime, mtime); err != nil {
			t.Fatalf("chtimes report %d: %v", i, err)
		}
		paths = append(paths, path)
	}

	if err := im.sweepReports(2); err != nil {
		t.Fatalf("sweepReports: %v", err)
	}

	for i, path := range paths {
		_, err := os.Stat(path)
		wantGone := i < 3 // the three oldest are swept; the two newest survive
		if wantGone && !errors.Is(err, os.ErrNotExist) {
			t.Errorf("report %d survived sweep: err = %v", i, err)
		}
		if !wantGone && err != nil {
			t.Errorf("report %d removed by sweep: %v", i, err)
		}
	}
}

func TestSweepReportsNonPositiveRetentionKeepsEverything(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)
	if err := os.MkdirAll(im.Cfg.Paths.StateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}
	path := im.reportPath("only")
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatalf("write report: %v", err)
	}
	if err := im.sweepReports(0); err != nil {
		t.Fatalf("sweepReports: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("sweepReports(0) removed a report: %v", err)
	}
}

// TestImportSessionLockedMissingSession proves a --session (or
// nudge-triggered) target that matches no promoted directory fails with
// [ErrSessionNotFound], still writing a per-run report -- the absence
// of a report is the monitoring event, not the caller's returned error.
func TestImportSessionLockedMissingSession(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)

	err := im.ImportSessionLocked(context.Background(), "does-not-exist")
	if !errors.Is(err, ErrSessionNotFound) {
		t.Fatalf("ImportSessionLocked error = %v, want ErrSessionNotFound", err)
	}

	reports, globErr := filepath.Glob(filepath.Join(im.Cfg.Paths.StateDir, reportFilePrefix+"*"+reportFileSuffix))
	if globErr != nil {
		t.Fatalf("glob reports: %v", globErr)
	}
	if len(reports) != 1 {
		t.Fatalf("report files = %d, want exactly 1", len(reports))
	}
	data, err := os.ReadFile(reports[0])
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report RunReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if report.Error == "" {
		t.Error("report.Error is empty, want the session-not-found failure recorded")
	}
}

// TestAcquireOrRecordHeldReportsHealthyOverlap proves the held-lock
// path RunSession shares with Run: another holder's lock is a healthy
// overlap, not a failure -- ok is false, err is nil, and a minimal
// LockHeld report is still written.
func TestAcquireOrRecordHeldReportsHealthyOverlap(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)
	if err := os.MkdirAll(im.Cfg.Paths.StateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	held, err := lock.Acquire(lock.Path(im.Cfg.Paths.StateDir))
	if err != nil {
		t.Fatalf("acquire lock for the test's own holder: %v", err)
	}
	defer func() { _ = held.Release() }()

	l, ok, err := im.acquireOrRecordHeld(context.Background())
	if err != nil {
		t.Fatalf("acquireOrRecordHeld error = %v, want nil", err)
	}
	if ok {
		t.Fatal("acquireOrRecordHeld ok = true, want false: the lock is already held")
	}
	if l != nil {
		t.Errorf("acquireOrRecordHeld lock = %v, want nil", l)
	}

	reports, globErr := filepath.Glob(filepath.Join(im.Cfg.Paths.StateDir, reportFilePrefix+"*"+reportFileSuffix))
	if globErr != nil {
		t.Fatalf("glob reports: %v", globErr)
	}
	if len(reports) != 1 {
		t.Fatalf("report files = %d, want exactly 1", len(reports))
	}
	data, err := os.ReadFile(reports[0])
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	var report RunReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("unmarshal report: %v", err)
	}
	if !report.LockHeld {
		t.Error("report.LockHeld = false, want true")
	}
}

// TestTenConsecutiveRunsPreserveLedgerHistory proves the move from one
// shared, overwritten report file to per-run report-<runid>.json files
// (U8) does not break the ledger's own accumulation: ten consecutive
// poll-mode runs, sharing one frozen clock so their report IDs cannot
// derive uniqueness from time, each still append their own terminal
// ledger event, and none of the earlier nine are lost or overwritten.
func TestTenConsecutiveRunsPreserveLedgerHistory(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)

	led, err := ledger.Open(filepath.Join(root, "contributions.jsonl"))
	if err != nil {
		t.Fatalf("open ledger: %v", err)
	}
	im.Ledger = led
	// lock.Acquire requires StateDir to already exist (production relies
	// on deploy.sh's pre-create, same as the persistent-lock-file
	// doctrine internal/lock documents); Run's own prepareRunDirs only
	// runs after the lock is already held.
	if err := os.MkdirAll(im.Cfg.Paths.StateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	// Frozen well ahead of real wall-clock time so every iteration's
	// freshly written file -- real mtime, taken during the test's brief
	// real run -- reads as quiescent against this one unmoving cutoff,
	// no matter how many iterations run before it.
	frozen := time.Now().Add(time.Hour)
	im.Now = func() time.Time { return frozen }

	for i := range 10 {
		sessionID := fmt.Sprintf("sess%02d", i)
		album := fmt.Sprintf("Album %02d", i)
		// track.mp3 fails the quality gate's extension check before beets
		// is ever invoked (pendingUnits is empty), so this run needs no
		// beet/flac/metaflac binaries -- it exercises exactly the
		// gate-reject-then-append-ledger path a timer run takes.
		stagePromotedSession(t, im.Cfg.Paths.InboxDir, sessionID, album, "track.mp3", "")

		if err := im.Run(context.Background()); err != nil {
			t.Fatalf("Run() iteration %d: %v", i, err)
		}
	}

	events, err := led.History("alice@example.com")
	if err != nil {
		t.Fatalf("ledger.History: %v", err)
	}
	if len(events) != 10 {
		t.Fatalf("ledger history = %d events, want 10 (one per run): %+v", len(events), events)
	}
	seen := make(map[string]bool, 10)
	for _, e := range events {
		if e.State != ledger.StateRejected {
			t.Errorf("event state = %q, want %q", e.State, ledger.StateRejected)
		}
		if seen[e.SessionID] {
			t.Errorf("session id %q appended more than once", e.SessionID)
		}
		seen[e.SessionID] = true
	}
	if len(seen) != 10 {
		t.Errorf("distinct session ids in history = %d, want 10", len(seen))
	}

	reports, err := filepath.Glob(filepath.Join(im.Cfg.Paths.StateDir, reportFilePrefix+"*"+reportFileSuffix))
	if err != nil {
		t.Fatalf("glob reports: %v", err)
	}
	if len(reports) != 10 {
		t.Errorf("report files = %d, want 10 distinct per-run files despite a shared frozen clock", len(reports))
	}
}

// writeStubExec writes an executable shell script at dir/name and
// returns its path -- the same BeetPath/FlacPath/MetaflacPath/
// SqlitePath test-seam pattern internal/beets and internal/gate's own
// test suites use.
func writeStubExec(t *testing.T, dir, name, script string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+script), 0o755); err != nil {
		t.Fatalf("write stub %s: %v", name, err)
	}
	return path
}

// TestImportSessionLockedAppliesSidecarCollectionAndTagsLedger drives
// ImportSessionLocked through a full accept: the quality gate passes
// (stubbed flac/metaflac), the beet import moves the file out of
// staging (stubbed beet, proving OutcomeImported), and the sidecar's
// OWN collection -- not [config.Config.Collections]' default -- is what
// reaches beets' --set flag, and what the appended StateTagging ledger
// event carries alongside the beets-resolved LibraryPaths (stubbed
// sqlite3).
func TestImportSessionLockedAppliesSidecarCollectionAndTagsLedger(t *testing.T) {
	root := t.TempDir()
	im := testImporterForRun(root)
	im.Cfg.Collections = []string{"default"}

	led, err := ledger.Open(filepath.Join(root, "contributions.jsonl"))
	if err != nil {
		t.Fatalf("open ledger: %v", err)
	}
	im.Ledger = led

	binDir := t.TempDir()
	im.FlacPath = writeStubExec(t, binDir, "flac", "exit 0\n")
	im.MetaflacPath = writeStubExec(t, binDir, "metaflac", "echo 44100; echo 16\n")
	im.SqlitePath = writeStubExec(t, binDir, "sqlite3", "echo '1\t/srv/music/library/Alice/Album/01.flac'\n")

	argsLog := filepath.Join(binDir, "beet-args.log")
	beetScript := fmt.Sprintf(`for a in "$@"; do printf '%%s\n' "$a"; done > %q
for a in "$@"; do lastarg="$a"; done
find "$lastarg" -type f -delete
`, argsLog)
	im.Beets = &beets.Runner{BeetPath: writeStubExec(t, binDir, "beet", beetScript), StateDir: im.Cfg.Paths.StateDir}

	stagePromotedSession(t, im.Cfg.Paths.InboxDir, "sessABC", "Album A", "track.flac", "vinyl")

	if err := im.ImportSessionLocked(context.Background(), "sessABC"); err != nil {
		t.Fatalf("ImportSessionLocked: %v", err)
	}

	argsData, err := os.ReadFile(argsLog)
	if err != nil {
		t.Fatalf("read beet args log: %v", err)
	}
	if !strings.Contains(string(argsData), "--set\ncollection=vinyl\n") {
		t.Errorf("beet args = %q, want the sidecar's own collection (vinyl), not the config default", argsData)
	}

	events, err := led.Session("sessABC")
	if err != nil {
		t.Fatalf("ledger.Session: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("ledger events for sessABC = %d, want 1", len(events))
	}
	e := events[0]
	if e.State != ledger.StateTagging {
		t.Errorf("state = %q, want %q", e.State, ledger.StateTagging)
	}
	if e.Email != "alice@example.com" {
		t.Errorf("email = %q, want alice@example.com", e.Email)
	}
	if len(e.LibraryPaths) != 1 || e.LibraryPaths[0] != "/srv/music/library/Alice/Album/01.flac" {
		t.Errorf("library paths = %v, want the beets-database match", e.LibraryPaths)
	}
}
