package tagio

import (
	"reflect"
	"slices"
	"testing"
)

// baseFixtureBlocks is the minimal well-formed block chain every case
// in this file writes onto: a STREAMINFO block followed by a tiny
// PADDING block that Write drops and replaces with its own 8 KiB
// reservation (rulings 5 and 6). Nothing in the fixture carries a
// VORBIS_COMMENT block, so the comment set a case reads back is
// exactly what its own write produced, with no leftover comment from
// the fixture to filter out.
func baseFixtureBlocks() []testBlock {
	return []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 0)},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
}

// writeAndRead writes ts onto a fresh copy of baseFixtureBlocks and
// returns both the raw on-disk comment sequence Parse decodes and the
// TagSet Read projects back from the same file, so one write grades a
// serialization rule on disk and its read-back counterpart together.
func writeAndRead(t *testing.T, ts TagSet) ([]Comment, TagSet) {
	t.Helper()
	path := writeFixture(t, buildContainer(baseFixtureBlocks(), nil))
	if err := Write(path, ts, nil, 0o644); err != nil {
		t.Fatalf("Write: %v", err)
	}
	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	return c.VorbisComment.Comments, got
}

// TestRulesLedgerSilentKeys is ruling 1's own test: one row per
// serialization rule for the sixteen keys the ledger cannot grade
// (spec:577-596), plus bpm, which the ledger does grade but which
// keeps a row because it is the integer 0 on all twelve corpus rows
// and a zero that must survive a write is worth pinning. Every
// wantComments value is hard-coded and cites the mediafile line it
// reproduces, never derived from Serialize, rgGainField, r128Field or
// any other code under test. Each case writes onto an otherwise empty
// TagSet, so the comment set it reads back on disk carries only the
// comments its own field produced.
func TestRulesLedgerSilentKeys(t *testing.T) {
	tests := []struct {
		name         string
		ts           TagSet
		wantComments []Comment
		checkRead    func(t *testing.T, got TagSet)
		citation     string
	}{
		{
			name:         "acoustid_fingerprint",
			ts:           TagSet{AcoustIDFingerprint: new("invented-fingerprint-value")},
			wantComments: []Comment{{Name: "ACOUSTID_FINGERPRINT", Value: "invented-fingerprint-value"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.AcoustIDFingerprint == nil || *got.AcoustIDFingerprint != "invented-fingerprint-value" {
					t.Errorf("AcoustIDFingerprint = %v, want %q", got.AcoustIDFingerprint, "invented-fingerprint-value")
				}
			},
			citation: "mediafile:817",
		},
		{
			name:         "acoustid_id",
			ts:           TagSet{AcoustIDID: new("11111111-2222-3333-4444-555555555555")},
			wantComments: []Comment{{Name: "ACOUSTID_ID", Value: "11111111-2222-3333-4444-555555555555"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.AcoustIDID == nil || *got.AcoustIDID != "11111111-2222-3333-4444-555555555555" {
					t.Errorf("AcoustIDID = %v, want %q", got.AcoustIDID, "11111111-2222-3333-4444-555555555555")
				}
			},
			citation: "mediafile:823",
		},
		{
			name: "comments",
			ts:   TagSet{Comments: new("Invented comment text.")},
			wantComments: []Comment{
				{Name: "DESCRIPTION", Value: "Invented comment text."},
				{Name: "COMMENT", Value: "Invented comment text."},
			},
			checkRead: func(t *testing.T, got TagSet) {
				if got.Comments == nil || *got.Comments != "Invented comment text." {
					t.Errorf("Comments = %v, want %q", got.Comments, "Invented comment text.")
				}
			},
			citation: "mediafile:514,515",
		},
		{
			name: "encoder",
			ts:   TagSet{Encoder: new("Invented Encoder 1.0")},
			wantComments: []Comment{
				{Name: "ENCODEDBY", Value: "Invented Encoder 1.0"},
				{Name: "ENCODER", Value: "Invented Encoder 1.0"},
			},
			checkRead: func(t *testing.T, got TagSet) {
				if got.Encoder == nil || *got.Encoder != "Invented Encoder 1.0" {
					t.Errorf("Encoder = %v, want %q", got.Encoder, "Invented Encoder 1.0")
				}
			},
			citation: "mediafile:638,639",
		},
		{
			name: "genres",
			ts:   TagSet{Genres: []string{"Ambient", "Drone"}},
			wantComments: []Comment{
				{Name: "GENRE", Value: "Ambient"},
				{Name: "GENRE", Value: "Drone"},
			},
			checkRead: func(t *testing.T, got TagSet) {
				if !slices.Equal(got.Genres, []string{"Ambient", "Drone"}) {
					t.Errorf("Genres = %v, want [Ambient Drone]", got.Genres)
				}
			},
			citation: "mediafile:416",
		},
		{
			name:         "grouping",
			ts:           TagSet{Grouping: new("Invented Grouping")},
			wantComments: []Comment{{Name: "GROUPING", Value: "Invented Grouping"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.Grouping == nil || *got.Grouping != "Invented Grouping" {
					t.Errorf("Grouping = %v, want %q", got.Grouping, "Invented Grouping")
				}
			},
			citation: "mediafile:452",
		},
		{
			name:         "initial_key",
			ts:           TagSet{InitialKey: new("Am")},
			wantComments: []Comment{{Name: "INITIALKEY", Value: "Am"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.InitialKey == nil || *got.InitialKey != "Am" {
					t.Errorf("InitialKey = %v, want %q", got.InitialKey, "Am")
				}
			},
			citation: "mediafile:888",
		},
		{
			name:         "lyrics",
			ts:           TagSet{Lyrics: new("la la la")},
			wantComments: []Comment{{Name: "LYRICS", Value: "la la la"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.Lyrics == nil || *got.Lyrics != "la la la" {
					t.Errorf("Lyrics = %v, want %q", got.Lyrics, "la la la")
				}
			},
			citation: "mediafile:504",
		},
		{
			// The Q8 shift: round(value * 2**8). -2.25 * 256 = -576.0
			// exactly, hard-coded rather than computed here
			// (mediafile:881, __init__.py:870-884).
			name:         "r128_album_gain",
			ts:           TagSet{R128AlbumGain: new(-2.25)},
			wantComments: []Comment{{Name: "R128_ALBUM_GAIN", Value: "-576"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.R128AlbumGain == nil || *got.R128AlbumGain != -2.25 {
					t.Errorf("R128AlbumGain = %v, want -2.25", got.R128AlbumGain)
				}
			},
			citation: "mediafile:881",
		},
		{
			// -1.5 * 256 = -384.0 exactly, INV's own measurement
			// (mediafile:874, __init__.py:870-884).
			name:         "r128_track_gain",
			ts:           TagSet{R128TrackGain: new(-1.5)},
			wantComments: []Comment{{Name: "R128_TRACK_GAIN", Value: "-384"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.R128TrackGain == nil || *got.R128TrackGain != -1.5 {
					t.Errorf("R128TrackGain = %v, want -1.5", got.R128TrackGain)
				}
			},
			citation: "mediafile:874",
		},
		{
			name:         "remixers",
			ts:           TagSet{Remixers: []string{"Invented Remixer"}},
			wantComments: []Comment{{Name: "REMIXER", Value: "Invented Remixer"}},
			checkRead: func(t *testing.T, got TagSet) {
				if !slices.Equal(got.Remixers, []string{"Invented Remixer"}) {
					t.Errorf("Remixers = %v, want [Invented Remixer]", got.Remixers)
				}
			},
			citation: "mediafile:750",
		},
		{
			// Two decimal places, the default float_places, with the
			// " dB" suffix that is the only suffix in the FLAC table
			// (mediafile:846).
			name:         "rg_album_gain",
			ts:           TagSet{RgAlbumGain: new(-6.5)},
			wantComments: []Comment{{Name: "REPLAYGAIN_ALBUM_GAIN", Value: "-6.50 dB"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.RgAlbumGain == nil || *got.RgAlbumGain != -6.5 {
					t.Errorf("RgAlbumGain = %v, want -6.5", got.RgAlbumGain)
				}
			},
			citation: "mediafile:846",
		},
		{
			// Six decimal places, no suffix (mediafile:864).
			name:         "rg_album_peak",
			ts:           TagSet{RgAlbumPeak: new(0.654321)},
			wantComments: []Comment{{Name: "REPLAYGAIN_ALBUM_PEAK", Value: "0.654321"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.RgAlbumPeak == nil || *got.RgAlbumPeak != 0.654321 {
					t.Errorf("RgAlbumPeak = %v, want 0.654321", got.RgAlbumPeak)
				}
			},
			citation: "mediafile:864",
		},
		{
			name:         "rg_track_gain",
			ts:           TagSet{RgTrackGain: new(-7.25)},
			wantComments: []Comment{{Name: "REPLAYGAIN_TRACK_GAIN", Value: "-7.25 dB"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.RgTrackGain == nil || *got.RgTrackGain != -7.25 {
					t.Errorf("RgTrackGain = %v, want -7.25", got.RgTrackGain)
				}
			},
			citation: "mediafile:836",
		},
		{
			name:         "rg_track_peak",
			ts:           TagSet{RgTrackPeak: new(0.123456)},
			wantComments: []Comment{{Name: "REPLAYGAIN_TRACK_PEAK", Value: "0.123456"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.RgTrackPeak == nil || *got.RgTrackPeak != 0.123456 {
					t.Errorf("RgTrackPeak = %v, want 0.123456", got.RgTrackPeak)
				}
			},
			citation: "mediafile:856",
		},
		{
			name:         "subtitle",
			ts:           TagSet{Subtitle: new("Invented Subtitle")},
			wantComments: []Comment{{Name: "SUBTITLE", Value: "Invented Subtitle"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.Subtitle == nil || *got.Subtitle != "Invented Subtitle" {
					t.Errorf("Subtitle = %v, want %q", got.Subtitle, "Invented Subtitle")
				}
			},
			citation: "mediafile:457",
		},
		{
			// bpm is graded by the ledger (it is BPM=0 on all twelve
			// rows), but a zero that must survive a write, rather than
			// being treated as absent, is worth pinning here too
			// (mediafile:528).
			name:         "bpm",
			ts:           TagSet{Bpm: new(0)},
			wantComments: []Comment{{Name: "BPM", Value: "0"}},
			checkRead: func(t *testing.T, got TagSet) {
				if got.Bpm == nil || *got.Bpm != 0 {
					t.Errorf("Bpm = %v, want pointer to 0", got.Bpm)
				}
			},
			citation: "mediafile:528",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotComments, gotRead := writeAndRead(t, tt.ts)
			if !reflect.DeepEqual(gotComments, tt.wantComments) {
				t.Errorf("comments = %+v, want %+v (%s)", gotComments, tt.wantComments, tt.citation)
			}
			tt.checkRead(t, gotRead)
		})
	}
}

// TestRulesZeroYearRoundTrip covers the zero year, a case the corpus
// cannot reach because no MusicBrainz release carries one (ruling 1,
// spec:201-207). A write of year 0 on both the year and original_year
// triples emits DATE=0000, YEAR=0 and ORIGINALDATE=0000 (the
// four-digit zero-padded year with no month or day component, since
// both are absent), and a read-back finds year 0 with month and day
// absent, because a nil month or day parses to nil rather than 0
// (mediafile/fields.py:156-172, serializeDate's compose-as-a-unit
// rule).
func TestRulesZeroYearRoundTrip(t *testing.T) {
	ts := TagSet{Year: new(0), OriginalYear: new(0)}
	gotComments, gotRead := writeAndRead(t, ts)

	want := []Comment{
		{Name: "DATE", Value: "0000"},
		{Name: "YEAR", Value: "0"},
		{Name: "ORIGINALDATE", Value: "0000"},
	}
	if !reflect.DeepEqual(gotComments, want) {
		t.Errorf("comments = %+v, want %+v", gotComments, want)
	}

	if gotRead.Year == nil || *gotRead.Year != 0 {
		t.Errorf("Year = %v, want a pointer to 0", gotRead.Year)
	}
	if gotRead.Month != nil {
		t.Errorf("Month = %v, want nil (absent)", gotRead.Month)
	}
	if gotRead.Day != nil {
		t.Errorf("Day = %v, want nil (absent)", gotRead.Day)
	}
	if gotRead.OriginalYear == nil || *gotRead.OriginalYear != 0 {
		t.Errorf("OriginalYear = %v, want a pointer to 0", gotRead.OriginalYear)
	}
	if gotRead.OriginalMonth != nil {
		t.Errorf("OriginalMonth = %v, want nil (absent)", gotRead.OriginalMonth)
	}
	if gotRead.OriginalDay != nil {
		t.Errorf("OriginalDay = %v, want nil (absent)", gotRead.OriginalDay)
	}
}

// TestRulesEmptyListRoundTrip covers C2's third state on a round trip
// rather than at Serialize alone: an empty list writes no comment at
// all (ListStorageStyle.set_list, mediafile/storage/base.py:189-199),
// and the empty state is indistinguishable from absent once it is on
// disk, so a read-back finds a nil slice, the same value an absent
// field projects to (deviation 3).
func TestRulesEmptyListRoundTrip(t *testing.T) {
	gotComments, gotRead := writeAndRead(t, TagSet{Genres: EmptyList()})

	if len(gotComments) != 0 {
		t.Errorf("comments = %+v, want none: an empty list writes no comment", gotComments)
	}
	if gotRead.Genres != nil {
		t.Errorf("Genres = %+v, want nil: an empty list is indistinguishable from absent on read-back", gotRead.Genres)
	}
}

// TestRulesCatalognumEmptyRoundTrip covers the pair C5 exists for: an
// empty string writes one comment carrying an empty value under
// CATALOGNUMBER, the only writable style, but that empty value is
// falsy, so the read-back last-style rule falls through to
// DISCOGS_CATALOG, a read-only style this write never touches and the
// fixture never carried, leaving catalognum absent on read-back
// (mediafile/fields.py:48-54).
func TestRulesCatalognumEmptyRoundTrip(t *testing.T) {
	gotComments, gotRead := writeAndRead(t, TagSet{Catalognum: EmptyString()})

	want := []Comment{{Name: "CATALOGNUMBER", Value: ""}}
	if !reflect.DeepEqual(gotComments, want) {
		t.Errorf("comments = %+v, want %+v", gotComments, want)
	}
	if gotRead.Catalognum != nil {
		t.Errorf("Catalognum = %v, want nil (absent)", gotRead.Catalognum)
	}
}

// TestRulesBarcodeEmptyRoundTrip is catalognum's own pair: an empty
// string writes one comment carrying an empty value under BARCODE, the
// only writable style, and the read-back last-style rule falls through
// to UPN, a read-only style this write never touches, leaving barcode
// absent on read-back (mediafile/fields.py:48-54).
func TestRulesBarcodeEmptyRoundTrip(t *testing.T) {
	gotComments, gotRead := writeAndRead(t, TagSet{Barcode: EmptyString()})

	want := []Comment{{Name: "BARCODE", Value: ""}}
	if !reflect.DeepEqual(gotComments, want) {
		t.Errorf("comments = %+v, want %+v", gotComments, want)
	}
	if gotRead.Barcode != nil {
		t.Errorf("Barcode = %v, want nil (absent)", gotRead.Barcode)
	}
}

// TestRulesInvalidUTF8ReadBack covers an invalid UTF-8 sequence, a
// case the corpus cannot reach because every value the oracle wrote
// there is valid UTF-8 by construction. The fixture is built directly
// with a raw invalid byte in the comment payload, never through
// Write, which only ever handles values that are already valid UTF-8
// (ruling 10). The want string is a hard-coded literal, not a value
// computed from replaceInvalidUTF8 or the stdlib UTF-8 decoder.
func TestRulesInvalidUTF8ReadBack(t *testing.T) {
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 0)},
		{Type: blockVorbisComment, Payload: vorbisCommentPayload("tagio test", []string{"TITLE=before\x80after"})},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	want := "before�after"
	if got.Title == nil || *got.Title != want {
		t.Errorf("Title = %v, want %q (mutagen's errors='replace', mutagen/_vorbis.py:109)", got.Title, want)
	}
}

// TestRulesU007EFieldNamePreservedThroughRead covers the U+007E field
// name at the Read level rather than at Parse alone
// (TestParseCommentNameU007EPreserved, container_test.go): a comment
// whose name is exactly at RFC 9639's bound, one code point past
// mutagen's own is_valid_key bound, is a foreign key no field in
// Mapping owns, and it must not disturb projection of a known field
// that sits beside it (ruling 1, spec:201-207).
func TestRulesU007EFieldNamePreservedThroughRead(t *testing.T) {
	blocks := []testBlock{
		{Type: blockStreamInfo, Payload: streamInfoPayload(44100, 2, 16, 0)},
		{Type: blockVorbisComment, Payload: vorbisCommentPayload("tagio test", []string{"TITLE=A", "X~=tilde value"})},
		{Type: blockPadding, Last: true, Payload: make([]byte, 16)},
	}
	path := writeFixture(t, buildContainer(blocks, nil))

	got, err := Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if got.Title == nil || *got.Title != "A" {
		t.Errorf("Title = %v, want %q: a foreign key at the U+007E bound must not break projection of a known field", got.Title, "A")
	}

	c, err := Parse(path)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	found := false
	for _, com := range c.VorbisComment.Comments {
		if com.Name == "X~" {
			found = true
			if com.Value != "tilde value" {
				t.Errorf("X~ value = %q, want %q", com.Value, "tilde value")
			}
		}
	}
	if !found {
		t.Error("X~ comment did not survive the parse")
	}
}
