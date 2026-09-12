package shadow

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/tagio"
)

// TestReadFLACMapsEveryVorbisKey grades the eighteen-row projection to
// matcher.ItemFields over a real container, written through the
// repo's own writer rather than a hand-rolled builder (ruling 4, R3).
// Every one of the seventeen tag-derived rows carries a value
// distinguishable from every other row's, so a projection that swaps
// two rows, or drops one, fails on the swapped or dropped field by
// name rather than by coincidence. Length is the eighteenth row,
// graded here by the fixture's one-second seed
// (committedTemplateFLACOneSecond) on top of A4's own table below.
func TestReadFLACMapsEveryVorbisKey(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.flac")
	writeFixtureTrack(t, path, tagio.TagSet{
		Title:         new("Sub Rosa"),
		Artist:        new("Cabinet"),
		Album:         new("Windrose"),
		Albumartist:   new("Cabinet Ensemble"),
		Track:         new(3),
		Disc:          new(2),
		Disctotal:     new(4),
		Comp:          new(true),
		Label:         new("Nightwood"),
		Catalognum:    new("NW-004"),
		Barcode:       new("5901234123457"),
		Country:       new("GB"),
		Media:         new("CD"),
		Albumdisambig: new("deluxe"),
		MBTrackid:     new("11111111-1111-1111-1111-111111111111"),
		MBAlbumid:     new("22222222-2222-2222-2222-222222222222"),
		Year:          new(1994),
	})

	got, err := readFLAC(path)
	if err != nil {
		t.Fatalf("readFLAC: %v", err)
	}
	want := matcher.ItemFields{
		Title:         "Sub Rosa",
		Artist:        "Cabinet",
		Album:         "Windrose",
		AlbumArtist:   "Cabinet Ensemble",
		Track:         3,
		Disc:          2,
		DiscTotal:     4,
		Comp:          true,
		Label:         "Nightwood",
		CatalogNum:    "NW-004",
		Barcode:       "5901234123457",
		Country:       "GB",
		Media:         "CD",
		AlbumDisambig: "deluxe",
		MBTrackID:     "11111111-1111-1111-1111-111111111111",
		MBAlbumID:     "22222222-2222-2222-2222-222222222222",
		Year:          1994,
		Length:        1,
		// DataSource stays "": beets keeps it as a flexible attribute
		// rather than in the file, so no Vorbis key exists
		// (libraryread.go's projectItemFields doc comment).
	}
	if got != want {
		t.Errorf("readFLAC =\n%+v\nwant\n%+v", got, want)
	}
}

// TestReadFLACProjectsAbsentFieldsToTypedZeroValues grades
// ProjectItemFields' own typed-default rule (info.go:178-188) on the
// read side: a tagio.TagSet member no Vorbis key set is nil, and every
// nil member here projects to its Go zero value rather than to a
// dereference panic or a literal "nil" string.
func TestReadFLACProjectsAbsentFieldsToTypedZeroValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.flac")
	writeFixtureTrack(t, path, tagio.TagSet{})

	got, err := readFLAC(path)
	if err != nil {
		t.Fatalf("readFLAC: %v", err)
	}
	want := matcher.ItemFields{Length: 1} // the fixture's one-second seed
	if got != want {
		t.Errorf("readFLAC =\n%+v\nwant\n%+v", got, want)
	}
}

// TestReadFLACLengthFromStreaminfo grades A4's seconds division and
// its zero-rate guard over a tagio.StreamInfo value directly, needing
// no file at all.
func TestReadFLACLengthFromStreaminfo(t *testing.T) {
	tests := []struct {
		name string
		info tagio.StreamInfo
		want float64
	}{
		{name: "ten seconds at 44100 Hz", info: tagio.StreamInfo{SampleRate: 44100, TotalSamples: 441000}, want: 10},
		{name: "a fractional length", info: tagio.StreamInfo{SampleRate: 44100, TotalSamples: 22050}, want: 0.5},
		{name: "a zero rate is zero, never a division by zero", info: tagio.StreamInfo{SampleRate: 0, TotalSamples: 441000}, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := trackLength(tt.info); got != tt.want {
				t.Errorf("trackLength(%+v) = %v, want %v", tt.info, got, tt.want)
			}
		})
	}
}

// TestReadFLACProjectsInvalidUTF8AsReplacementChar grades R6's carried
// value change on the shadow side: a VORBIS_COMMENT value with an
// ill-formed UTF-8 byte projects U+FFFD, not the raw byte. The
// fixture is a length-preserving byte splice of a writer-built track
// (spliceInvalidUTF8TitleByte, fixture_test.go), because tagio.Write
// only ever accepts values already valid UTF-8 (ruling 10) and this
// package has no other way to put an invalid sequence in front of a
// reader. readFLAC reads through the same exported tagio.Tags
// projection tagio's own TestRulesInvalidUTF8ReadBack
// (internal/engine/tagio/rules_test.go) covers directly; this row
// proves only that shadow's readFLAC carries that same value change
// through, not a second copy of tagio's rule.
func TestReadFLACProjectsInvalidUTF8AsReplacementChar(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.flac")
	writeFixtureTrack(t, path, tagio.TagSet{Title: new("before_after")})

	spliced := spliceInvalidUTF8TitleByte(t, path)
	if err := os.WriteFile(path, spliced, 0o644); err != nil {
		t.Fatalf("write spliced fixture: %v", err)
	}
	if err := os.Chtimes(path, fixedModTime, fixedModTime); err != nil {
		t.Fatalf("chtimes spliced fixture: %v", err)
	}

	got, err := readFLAC(path)
	if err != nil {
		t.Fatalf("readFLAC: %v", err)
	}
	want := "before�after"
	if got.Title != want {
		t.Errorf("Title = %q, want %q", got.Title, want)
	}
}

// TestReadFLACRefusesAMalformedContainer grades the capture half of
// the landed reader's refusal behavior: readFLAC returns an error and
// the zero ItemFields for a container tagio.Parse refuses. The
// container half of this behavior, which byte shape refuses and why,
// is tagio's own (container_test.go's TestParseRefusals); this test
// only proves readFLAC never hands back a partly filled bag.
func TestReadFLACRefusesAMalformedContainer(t *testing.T) {
	path := filepath.Join(t.TempDir(), "track.flac")
	if err := os.WriteFile(path, []byte("not a FLAC stream"), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	got, err := readFLAC(path)
	if err == nil {
		t.Fatalf("readFLAC = %+v, want an error", got)
	}
	if got != (matcher.ItemFields{}) {
		t.Errorf("readFLAC = %+v, want the zero fields on an error", got)
	}
}
