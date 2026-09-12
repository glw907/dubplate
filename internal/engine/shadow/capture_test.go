package shadow

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/tagio"
)

// passNow is the pass clock every capture test shares, never a live
// clock read (the engine reads no clock, Global Constraints).
var passNow = time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)

// fixedModTime is the modification time every fixture writer in this
// package pins its files to, so a capture digest is a function of the
// fixture only (Resolution R5, fold B3).
var fixedModTime = time.Date(2026, 8, 15, 9, 0, 0, 0, time.UTC)

// writeAlbum writes one inbox album at dir/uploader/basename, one FLAC
// track per entry in tracks, each written through writeFixtureTrack.
// Every written file's modification time is pinned to fixedModTime, so
// two temp trees written at two wall-clock instants still yield the
// same capture digest.
func writeAlbum(t *testing.T, dir, uploader, basename string, tracks []tagio.TagSet) {
	t.Helper()

	albumDir := filepath.Join(dir, uploader, basename)
	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("mkdir album dir: %v", err)
	}

	for i, ts := range tracks {
		path := filepath.Join(albumDir, fmt.Sprintf("%02d.flac", i+1))
		writeFixtureTrack(t, path, ts)
	}
}

// corrupt overwrites path with a truncated FLAC container: the
// signature alone, with no metadata block. It is the shape a file
// renamed out from under the walk presents to readFLAC: the entry was
// listed and cannot be read whole.
func corrupt(t *testing.T, path string) {
	t.Helper()

	if err := os.WriteFile(path, []byte("fLaC"), 0o644); err != nil {
		t.Fatalf("corrupt fixture: %v", err)
	}
	if err := os.Chtimes(path, fixedModTime, fixedModTime); err != nil {
		t.Fatalf("chtimes corrupted fixture: %v", err)
	}
}

func windroseTracks() []tagio.TagSet {
	return []tagio.TagSet{
		{Albumartist: new("Cabinet"), Album: new("Windrose"), Title: new("One"), Track: new(1)},
		{Albumartist: new("Cabinet"), Album: new("Windrose"), Title: new("Two"), Track: new(2)},
	}
}

func TestCaptureReadsOneAlbumPerDirectory(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "windrose", windroseTracks())

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 1 {
		t.Fatalf("captures = %d, want 1", len(got.Captures))
	}
	c := got.Captures[0]
	if c.Uploader != "ada" || c.Basename != "windrose" {
		t.Errorf("capture = %s/%s, want ada/windrose", c.Uploader, c.Basename)
	}
	if len(c.Items) != 2 {
		t.Errorf("items = %d, want 2", len(c.Items))
	}
	if !c.CapturedAt.Equal(passNow) {
		t.Errorf("CapturedAt = %v, want the pass clock %v", c.CapturedAt, passNow)
	}
}

func TestCaptureExcludesAMultiAlbumDirectory(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "contributor-zip", []tagio.TagSet{
		{Albumartist: new("Cabinet"), Album: new("Windrose"), Title: new("One"), Track: new(1)},
		{Albumartist: new("Cabinet"), Album: new("Windrose"), Title: new("Two"), Track: new(2)},
		{Albumartist: new("Thicket"), Album: new("Understory"), Title: new("One"), Track: new(1)},
		{Albumartist: new("Thicket"), Album: new("Understory"), Title: new("Two"), Track: new(2)},
	})

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 0 {
		t.Errorf("captures = %d, want 0; a multi-album directory is excluded, never scored", len(got.Captures))
	}
	if len(got.Excluded) != 1 || got.Excluded[0].Reason != ExcludedMultiAlbum || got.Excluded[0].Groups != 2 {
		t.Errorf("excluded = %+v, want one multi-album row with 2 groups", got.Excluded)
	}
}

// TestCaptureExcludesAFlatArtistOnlyMultiAlbumDirectory proves the
// artist fallback: two ARTIST values under one ALBUM, with no
// ALBUMARTIST anywhere, is the shape a flat contributor ZIP produces
// (config/beets-config.yaml:19-22) and the shape beets splits at
// stages.py:102-103. An albumartist-only key would collapse this to
// one group (conductor ruling C1, spec:444).
func TestCaptureExcludesAFlatArtistOnlyMultiAlbumDirectory(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "flat-extract", []tagio.TagSet{
		{Artist: new("Cabinet"), Album: new("Various Artists"), Title: new("One"), Track: new(1)},
		{Artist: new("Cabinet"), Album: new("Various Artists"), Title: new("Two"), Track: new(2)},
		{Artist: new("Thicket"), Album: new("Various Artists"), Title: new("One"), Track: new(1)},
		{Artist: new("Thicket"), Album: new("Various Artists"), Title: new("Two"), Track: new(2)},
	})

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 0 {
		t.Errorf("captures = %d, want 0; a flat artist-only directory splits by beets' own fallback key", len(got.Captures))
	}
	if len(got.Excluded) != 1 || got.Excluded[0].Reason != ExcludedMultiAlbum || got.Excluded[0].Groups != 2 {
		t.Errorf("excluded = %+v, want one multi-album row with 2 groups", got.Excluded)
	}
}

func TestCaptureExcludesADirectoryWithNoTracks(t *testing.T) {
	dir := t.TempDir()
	albumDir := filepath.Join(dir, "ada", "cover-only")
	if err := os.MkdirAll(albumDir, 0o755); err != nil {
		t.Fatalf("mkdir album dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(albumDir, "cover.jpg"), []byte("not a track"), 0o644); err != nil {
		t.Fatalf("write cover: %v", err)
	}

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 0 {
		t.Errorf("captures = %d, want 0", len(got.Captures))
	}
	if len(got.Excluded) != 1 || got.Excluded[0].Reason != ExcludedNoTracks {
		t.Errorf("excluded = %+v, want one no-tracks row", got.Excluded)
	}
}

// TestCaptureDoesNotDescendIntoANestedDirectory proves the one-level
// rule: a subdirectory under an album directory carries a track for a
// different album entirely, and a capture that descended into it would
// join that track's fields to the parent's item list and change the
// group count. The capture observes one level only.
func TestCaptureDoesNotDescendIntoANestedDirectory(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "windrose", windroseTracks())

	nested := filepath.Join(dir, "ada", "windrose", "disc2")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir nested dir: %v", err)
	}
	nestedTrack := filepath.Join(nested, "01.flac")
	writeFixtureTrack(t, nestedTrack, tagio.TagSet{
		Albumartist: new("Thicket"), Album: new("Understory"), Title: new("One"), Track: new(1),
	})

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 1 {
		t.Fatalf("captures = %d, want 1; the nested directory is not descended into", len(got.Captures))
	}
	if len(got.Captures[0].Items) != 2 {
		t.Errorf("items = %d, want 2; a nested track must not join the parent's item list", len(got.Captures[0].Items))
	}
}

func TestCaptureDiscardsAPartialRead(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "half-gone", []tagio.TagSet{
		{Albumartist: new("Cabinet"), Album: new("Windrose"), Title: new("One"), Track: new(1)},
		{Albumartist: new("Cabinet"), Album: new("Windrose"), Title: new("Two"), Track: new(2)},
		{Albumartist: new("Cabinet"), Album: new("Windrose"), Title: new("Three"), Track: new(3)},
	})
	// Replace one track's bytes with a truncated container, which is
	// what a file renamed out from under the walk looks like to
	// readFLAC: the entry was listed and cannot be read whole.
	corrupt(t, filepath.Join(dir, "ada", "half-gone", "02.flac"))

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 0 {
		t.Fatalf("captures = %d, want 0; a partial read is never stored, never scored, never counted", len(got.Captures))
	}
	if len(got.Excluded) != 0 {
		t.Errorf("excluded = %+v, want none; a partial read is discarded, not excluded", got.Excluded)
	}
	if len(got.PartialReads) != 1 || got.PartialReads[0] != "ada/half-gone" {
		t.Errorf("partial reads = %v, want [ada/half-gone]", got.PartialReads)
	}
}

// TestCaptureTreatsAnUnreadableAlbumDirectoryAsAPartialRead proves the
// race the doc comment names: review.Move renames the album directory
// itself, so the walk can list a basename and then take ENOENT (or, as
// simulated here, EACCES) reading that directory's own entries. That
// directory is recorded as a partial read and the pass still returns
// its other captures, rather than failing the whole pass.
func TestCaptureTreatsAnUnreadableAlbumDirectoryAsAPartialRead(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "windrose", windroseTracks())
	writeAlbum(t, dir, "ada", "vanished", windroseTracks())

	vanishedDir := filepath.Join(dir, "ada", "vanished")
	if err := os.Chmod(vanishedDir, 0o000); err != nil {
		t.Fatalf("chmod vanished dir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(vanishedDir, 0o755)
	})

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 1 || got.Captures[0].Basename != "windrose" {
		t.Errorf("captures = %+v, want the one readable album", got.Captures)
	}
	if len(got.Excluded) != 0 {
		t.Errorf("excluded = %+v, want none; an unreadable directory is a partial read, not exclusion", got.Excluded)
	}
	if len(got.PartialReads) != 1 || got.PartialReads[0] != "ada/vanished" {
		t.Errorf("partial reads = %v, want [ada/vanished]", got.PartialReads)
	}
}

func TestCaptureIsIdempotentUnderAnUnmovedDigest(t *testing.T) {
	dir := t.TempDir()
	writeAlbum(t, dir, "ada", "windrose", windroseTracks())

	first, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture (first): %v", err)
	}
	if len(first.Captures) != 1 {
		t.Fatalf("captures (first) = %d, want 1", len(first.Captures))
	}
	digest := first.Captures[0].Digest
	id := observationID(digest)

	second, err := capture(dir, func(candidate string) bool { return candidate == id }, passNow)
	if err != nil {
		t.Fatalf("capture (second): %v", err)
	}
	if len(second.Captures) != 0 {
		t.Errorf("captures (second) = %d, want 0; a stored digest is not re-captured", len(second.Captures))
	}
}

func TestCaptureDigestMovesWithEveryFileProperty(t *testing.T) {
	baseline := func(t *testing.T) (dir string, digest string) {
		t.Helper()
		dir = t.TempDir()
		writeAlbum(t, dir, "ada", "windrose", windroseTracks())
		got, err := capture(dir, func(string) bool { return false }, passNow)
		if err != nil {
			t.Fatalf("capture (baseline): %v", err)
		}
		if len(got.Captures) != 1 {
			t.Fatalf("captures (baseline) = %d, want 1", len(got.Captures))
		}
		return dir, got.Captures[0].Digest
	}

	tests := []struct {
		name   string
		mutate func(t *testing.T, dir string)
	}{
		{
			name: "a renamed file",
			mutate: func(t *testing.T, dir string) {
				t.Helper()
				albumDir := filepath.Join(dir, "ada", "windrose")
				if err := os.Rename(filepath.Join(albumDir, "01.flac"), filepath.Join(albumDir, "00.flac")); err != nil {
					t.Fatalf("rename: %v", err)
				}
				if err := os.Chtimes(filepath.Join(albumDir, "00.flac"), fixedModTime, fixedModTime); err != nil {
					t.Fatalf("chtimes: %v", err)
				}
			},
		},
		{
			name: "a resized file",
			mutate: func(t *testing.T, dir string) {
				t.Helper()
				path := filepath.Join(dir, "ada", "windrose", "01.flac")
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read: %v", err)
				}
				data = append(data, 0)
				if err := os.WriteFile(path, data, 0o644); err != nil {
					t.Fatalf("write: %v", err)
				}
				if err := os.Chtimes(path, fixedModTime, fixedModTime); err != nil {
					t.Fatalf("chtimes: %v", err)
				}
			},
		},
		{
			name: "a touched file",
			mutate: func(t *testing.T, dir string) {
				t.Helper()
				path := filepath.Join(dir, "ada", "windrose", "01.flac")
				second := fixedModTime.Add(time.Hour)
				if err := os.Chtimes(path, second, second); err != nil {
					t.Fatalf("chtimes: %v", err)
				}
			},
		},
		{
			name: "an added cover image",
			mutate: func(t *testing.T, dir string) {
				t.Helper()
				path := filepath.Join(dir, "ada", "windrose", "cover.jpg")
				if err := os.WriteFile(path, []byte("cover art"), 0o644); err != nil {
					t.Fatalf("write cover: %v", err)
				}
				if err := os.Chtimes(path, fixedModTime, fixedModTime); err != nil {
					t.Fatalf("chtimes: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir, base := baseline(t)
			tt.mutate(t, dir)

			got, err := capture(dir, func(string) bool { return false }, passNow)
			if err != nil {
				t.Fatalf("capture (mutated): %v", err)
			}
			if len(got.Captures) != 1 {
				t.Fatalf("captures (mutated) = %d, want 1", len(got.Captures))
			}
			if got.Captures[0].Digest == base {
				t.Errorf("digest unchanged after %s, want it to move", tt.name)
			}
		})
	}
}

func TestCaptureWalksInLexicalOrder(t *testing.T) {
	dir := t.TempDir()
	for _, uploader := range []string{"carla", "bea", "ada"} {
		for _, basename := range []string{"zephyr", "midway", "arbor"} {
			writeAlbum(t, dir, uploader, basename, windroseTracks())
		}
	}

	got, err := capture(dir, func(string) bool { return false }, passNow)
	if err != nil {
		t.Fatalf("capture: %v", err)
	}
	if len(got.Captures) != 9 {
		t.Fatalf("captures = %d, want 9", len(got.Captures))
	}

	wantUploaders := []string{"ada", "ada", "ada", "bea", "bea", "bea", "carla", "carla", "carla"}
	wantBasenames := []string{"arbor", "midway", "zephyr", "arbor", "midway", "zephyr", "arbor", "midway", "zephyr"}
	for i, c := range got.Captures {
		if c.Uploader != wantUploaders[i] || c.Basename != wantBasenames[i] {
			t.Errorf("capture[%d] = %s/%s, want %s/%s", i, c.Uploader, c.Basename, wantUploaders[i], wantBasenames[i])
		}
	}
}
