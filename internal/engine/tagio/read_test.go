package tagio

import (
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// commentsFrom turns "NAME=VALUE" lines into the comment sequence a
// parse hands the projection, so a case table reads the way
// golden.txt records a block.
func commentsFrom(lines ...string) []Comment {
	cs := make([]Comment, len(lines))
	for i, line := range lines {
		name, value, _ := strings.Cut(line, "=")
		cs[i] = Comment{Name: name, Value: value}
	}
	return cs
}

// projectCase is one comment sequence and the whole TagSet it
// projects to. Every case compares all 69 members at once rather
// than the one the case is named for, so a projection that puts a
// value on the wrong member fails the case that value belongs to.
type projectCase struct {
	name string
	in   []string
	want TagSet
}

func runProjectCases(t *testing.T, cases []projectCase) {
	t.Helper()
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := project(Container{VorbisComment: VorbisComment{Comments: commentsFrom(c.in...)}})
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("projection differs: %s", diffFields(got, c.want))
			}
		})
	}
}

// diffFields names the TagSet members on which got and want differ,
// so a whole-set comparison still says which field drifted.
func diffFields(got, want TagSet) string {
	vg, vw := reflect.ValueOf(got), reflect.ValueOf(want)
	var out []string
	for i := range vg.NumField() {
		if reflect.DeepEqual(vg.Field(i).Interface(), vw.Field(i).Interface()) {
			continue
		}
		out = append(out, fmt.Sprintf("%s: got %s, want %s", vg.Type().Field(i).Name, show(vg.Field(i)), show(vw.Field(i))))
	}
	if out == nil {
		return "no member differs"
	}
	return strings.Join(out, "; ")
}

// show renders a TagSet member's value, dereferencing a pointer so a
// failure reports the value rather than an address.
func show(v reflect.Value) string {
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return "absent"
		}
		return fmt.Sprintf("%#v", v.Elem().Interface())
	}
	if v.IsNil() {
		return "absent"
	}
	return fmt.Sprintf("%#v", v.Interface())
}

// goldenProjection is the read-back of testdata/golden.flac, field by
// field, transcribed from golden.txt's provenance record: its decoded
// comment sequence for the 32 members a value reaches, and its
// recorded read-back for the four corpus-unreachable cases (year 0
// with month and day absent, arrangers absent, catalognum absent,
// barcode absent). No value here is computed by Serialize, by project
// or by the generator that wrote the golden.
func goldenProjection() TagSet {
	return TagSet{
		AcoustIDFingerprint: new("invented-fingerprint-value"),
		AcoustIDID:          new("11111111-2222-3333-4444-555555555555"),
		Album:               new("Al"),
		Albumartists:        []string{"X", "Y"},
		Albumtype:           new("album"),
		Albumtypes:          []string{"album", "live"},
		Artist:              new("A"),
		Bpm:                 new(120),
		Comments:            new("Invented comment text."),
		Comp:                new(true),
		Year:                new(0),
		Disc:                new(1),
		Disctotal:           new(2),
		Encoder:             new("Invented Encoder 1.0"),
		Genres:              []string{"Ambient", "Drone"},
		Grouping:            new("Invented Grouping"),
		InitialKey:          new("Am"),
		Lyrics:              new("la la la"),
		OriginalYear:        new(1999),
		OriginalMonth:       new(3),
		OriginalDay:         new(4),
		R128AlbumGain:       new(-2.25),
		R128TrackGain:       new(-1.5),
		Remixers:            []string{"Invented Remixer"},
		RgAlbumGain:         new(-6.5),
		RgAlbumPeak:         new(0.654321),
		RgTrackGain:         new(-7.25),
		RgTrackPeak:         new(0.123456),
		Subtitle:            new("Invented Subtitle"),
		Title:               new("T"),
		Track:               new(3),
		Tracktotal:          new(12),
	}
}

// TestReadGolden grades all 69 projected values of the committed
// golden against golden.txt in one equality, which is what catches a
// projection that reads the right value onto the wrong member: a
// per-field test asserting disctotal is 2 and tracktotal is 12 passes
// unchanged when the two members are swapped in the projection, and
// this comparison does not.
func TestReadGolden(t *testing.T) {
	got, err := Read(filepath.Join("testdata", "golden.flac"))
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if want := goldenProjection(); !reflect.DeepEqual(got, want) {
		t.Errorf("golden projection differs: %s", diffFields(got, want))
	}
}

// TestProjectStylePrecedence grades C4's read half on every field
// carrying more than one style: the first truthy style wins, an empty
// style falls through to the next, and a "0" does not fall through,
// because mediafile tests truthiness on the fetched string before
// safe_cast sees it and "0" is a truthy Python string
// (mediafile/fields.py:48-54).
func TestProjectStylePrecedence(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"albumartist first style truthy", []string{"ALBUM ARTIST=First", "ALBUM_ARTIST=Second", "ALBUMARTIST=Third"}, TagSet{Albumartist: new("First")}},
		{"albumartist first style empty", []string{"ALBUM ARTIST=", "ALBUM_ARTIST=Second", "ALBUMARTIST=Third"}, TagSet{Albumartist: new("Second")}},
		{"albumartist zero in first style", []string{"ALBUM ARTIST=0", "ALBUM_ARTIST=Second"}, TagSet{Albumartist: new("0")}},

		{"albumartists first style truthy", []string{"ALBUMARTISTS=X", "ALBUMARTISTS=Y", "ALBUM_ARTISTS=Z"}, TagSet{Albumartists: []string{"X", "Y"}}},
		{"albumartists first style has no values", []string{"ALBUM_ARTISTS=Z", "ALBUM ARTISTS=RO"}, TagSet{Albumartists: []string{"Z"}}},
		{"albumartists zero in first style", []string{"ALBUMARTISTS=0", "ALBUM_ARTISTS=Z"}, TagSet{Albumartists: []string{"0"}}},

		{"albumstatus first style truthy", []string{"RELEASESTATUS=official", "MUSICBRAINZ_ALBUMSTATUS=bootleg"}, TagSet{Albumstatus: new("official")}},
		{"albumstatus first style empty", []string{"RELEASESTATUS=", "MUSICBRAINZ_ALBUMSTATUS=bootleg"}, TagSet{Albumstatus: new("bootleg")}},
		{"albumstatus zero in first style", []string{"RELEASESTATUS=0", "MUSICBRAINZ_ALBUMSTATUS=bootleg"}, TagSet{Albumstatus: new("0")}},

		{"albumtype first style truthy", []string{"RELEASETYPE=album", "MUSICBRAINZ_ALBUMTYPE=live"}, TagSet{Albumtype: new("album"), Albumtypes: []string{"album"}}},
		{"albumtype first style empty", []string{"RELEASETYPE=", "MUSICBRAINZ_ALBUMTYPE=live"}, TagSet{Albumtype: new("live"), Albumtypes: []string{""}}},
		{"albumtype zero in first style", []string{"RELEASETYPE=0", "MUSICBRAINZ_ALBUMTYPE=live"}, TagSet{Albumtype: new("0"), Albumtypes: []string{"0"}}},

		{"comments first style truthy", []string{"DESCRIPTION=First", "COMMENT=Second"}, TagSet{Comments: new("First")}},
		{"comments first style empty", []string{"DESCRIPTION=", "COMMENT=Second"}, TagSet{Comments: new("Second")}},
		{"comments zero in first style", []string{"DESCRIPTION=0", "COMMENT=Second"}, TagSet{Comments: new("0")}},

		{"disc first style truthy", []string{"DISC=1", "DISCNUMBER=9"}, TagSet{Disc: new(1)}},
		{"disc first style empty", []string{"DISC=", "DISCNUMBER=9"}, TagSet{Disc: new(9)}},
		{"disc zero in first style", []string{"DISC=0", "DISCNUMBER=9"}, TagSet{Disc: new(0)}},

		{"disctotal first style truthy", []string{"DISCTOTAL=2", "DISCC=7", "TOTALDISCS=8"}, TagSet{Disctotal: new(2)}},
		{"disctotal first style empty", []string{"DISCTOTAL=", "DISCC=7", "TOTALDISCS=8"}, TagSet{Disctotal: new(7)}},
		{"disctotal zero in first style", []string{"DISCTOTAL=0", "DISCC=7"}, TagSet{Disctotal: new(0)}},

		{"encoder first style truthy", []string{"ENCODEDBY=First", "ENCODER=Second"}, TagSet{Encoder: new("First")}},
		{"encoder first style empty", []string{"ENCODEDBY=", "ENCODER=Second"}, TagSet{Encoder: new("Second")}},
		{"encoder zero in first style", []string{"ENCODEDBY=0", "ENCODER=Second"}, TagSet{Encoder: new("0")}},

		{"label first style truthy", []string{"LABEL=First", "PUBLISHER=Second"}, TagSet{Label: new("First")}},
		{"label first style empty", []string{"LABEL=", "PUBLISHER=Second"}, TagSet{Label: new("Second")}},
		{"label zero in first style", []string{"LABEL=0", "PUBLISHER=Second"}, TagSet{Label: new("0")}},

		{"track first style truthy", []string{"TRACK=3", "TRACKNUMBER=5"}, TagSet{Track: new(3)}},
		{"track first style empty", []string{"TRACK=", "TRACKNUMBER=5"}, TagSet{Track: new(5)}},
		{"track zero in first style", []string{"TRACK=0", "TRACKNUMBER=5"}, TagSet{Track: new(0)}},

		{"tracktotal first style truthy", []string{"TRACKTOTAL=12", "TRACKC=7", "TOTALTRACKS=8"}, TagSet{Tracktotal: new(12)}},
		{"tracktotal first style empty", []string{"TRACKTOTAL=", "TRACKC=7", "TOTALTRACKS=8"}, TagSet{Tracktotal: new(7)}},
		{"tracktotal zero in first style", []string{"TRACKTOTAL=0", "TRACKC=7"}, TagSet{Tracktotal: new(0)}},
	})
}

// TestProjectLastStyleRule grades C5. MediaField.__get__ overwrites
// its result on every style and breaks only on a truthy one, so what
// survives a read with nothing truthy is the LAST style's value,
// absent included. catalognum and barcode are the two fields of the
// 69 whose trailing styles are read-only, so an empty written value
// reads back absent; a non-empty one reads back, which is what proves
// the rule fires only when nothing is truthy.
func TestProjectLastStyleRule(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"catalognum empty reads absent", []string{"CATALOGNUMBER="}, TagSet{}},
		{"catalognum value reads back", []string{"CATALOGNUMBER=CAT-1"}, TagSet{Catalognum: new("CAT-1")}},
		{"catalognum empty falls through to a read-only style", []string{"CATALOGNUMBER=", "CATALOGID=CAT-2"}, TagSet{Catalognum: new("CAT-2")}},

		{"barcode empty reads absent", []string{"BARCODE="}, TagSet{}},
		{"barcode value reads back", []string{"BARCODE=0123456789012"}, TagSet{Barcode: new("0123456789012")}},
		{"barcode empty falls through to its last read-only style", []string{"BARCODE=", "UPN=9"}, TagSet{Barcode: new("9")}},
		{"barcode walks past two absent middle styles to a truthy one", []string{"EAN=E-1", "UPN=U-1"}, TagSet{Barcode: new("E-1")}},

		{"comments empty in every style keeps the last style's empty value", []string{"DESCRIPTION=", "COMMENT="}, TagSet{Comments: new("")}},
		{"comments empty in the first style alone reads absent", []string{"DESCRIPTION="}, TagSet{}},
		{"album empty in its only style reads present and empty", []string{"ALBUM="}, TagSet{Album: new("")}},
	})
}

// TestProjectDates grades the date recomposition
// (mediafile/fields.py:156-172): the value splits on a hyphen or a
// slash after anything from a T or a space onward is dropped, each
// component parses strictly rather than through safe_cast's leading
// prefix, and a missing component is absent rather than zero. The
// bare-year YEAR fallback is read only when DATE's own year
// component is falsy.
func TestProjectDates(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"year alone", []string{"DATE=2001"}, TagSet{Year: new(2001)}},
		{"zero year", []string{"DATE=0000"}, TagSet{Year: new(0)}},
		{"zero year with its YEAR fallback", []string{"DATE=0000", "YEAR=0"}, TagSet{Year: new(0)}},
		{"full date", []string{"DATE=1999-03-04"}, TagSet{Year: new(1999), Month: new(3), Day: new(4)}},
		{"year and month", []string{"DATE=2001-05"}, TagSet{Year: new(2001), Month: new(5)}},
		{"time suffix dropped", []string{"DATE=2001-03-04T10:11:12"}, TagSet{Year: new(2001), Month: new(3), Day: new(4)}},
		{"slash separators", []string{"DATE=2001/03/04"}, TagSet{Year: new(2001), Month: new(3), Day: new(4)}},
		{"YEAR fallback when DATE is absent", []string{"YEAR=1998"}, TagSet{Year: new(1998)}},
		{"YEAR fallback when DATE is empty", []string{"DATE=", "YEAR=1998"}, TagSet{Year: new(1998)}},
		{"unparsable year is absent, not zero", []string{"DATE=19x9-03"}, TagSet{Month: new(3)}},
		{"no date key at all", nil, TagSet{}},
		{"original date recomposes with no YEAR fallback", []string{"ORIGINALDATE=1999-03-04", "YEAR=1998"}, TagSet{Year: new(1998), OriginalYear: new(1999), OriginalMonth: new(3), OriginalDay: new(4)}},
		{"original zero year", []string{"ORIGINALDATE=0000"}, TagSet{OriginalYear: new(0)}},
	})
}

// TestProjectNumericCasts grades the two numeric rules a Vorbis
// string does not carry on its face: the R128 pair is a Q8 fixed
// point integer on disk and divides by 2^8 on read
// (mediafile:874, :881), and the two ReplayGain gain fields carry the
// " dB" suffix their styles declare, which deserialize strips before
// safe_cast runs (mediafile:836, :846). The peaks carry no suffix
// (:856, :864). safe_cast itself takes a leading numeric prefix and
// yields the type's zero when there is none, which is a value rather
// than an absence.
func TestProjectNumericCasts(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"r128 track gain divides by 256", []string{"R128_TRACK_GAIN=-384"}, TagSet{R128TrackGain: new(-1.5)}},
		{"r128 album gain divides by 256", []string{"R128_ALBUM_GAIN=-576"}, TagSet{R128AlbumGain: new(-2.25)}},
		{"r128 positive value", []string{"R128_TRACK_GAIN=256"}, TagSet{R128TrackGain: new(1.0)}},
		{"r128 unparsable is zero", []string{"R128_TRACK_GAIN=none"}, TagSet{R128TrackGain: new(0.0)}},

		{"replaygain track gain strips dB", []string{"REPLAYGAIN_TRACK_GAIN=-7.50 dB"}, TagSet{RgTrackGain: new(-7.5)}},
		{"replaygain album gain strips dB", []string{"REPLAYGAIN_ALBUM_GAIN=3.00 dB"}, TagSet{RgAlbumGain: new(3.0)}},
		{"replaygain gain without the suffix", []string{"REPLAYGAIN_TRACK_GAIN=-7.50"}, TagSet{RgTrackGain: new(-7.5)}},
		{"replaygain peaks carry no suffix", []string{"REPLAYGAIN_TRACK_PEAK=0.987654", "REPLAYGAIN_ALBUM_PEAK=1.000000"}, TagSet{RgTrackPeak: new(0.987654), RgAlbumPeak: new(1.0)}},

		{"int takes a leading numeric prefix", []string{"TRACK=3/12"}, TagSet{Track: new(3)}},
		{"int without a prefix is zero", []string{"BPM=fast"}, TagSet{Bpm: new(0)}},
		{"bool one is true", []string{"COMPILATION=1"}, TagSet{Comp: new(true)}},
		{"bool zero is false", []string{"COMPILATION=0"}, TagSet{Comp: new(false)}},
		{"bool unparsable is false", []string{"COMPILATION=yes"}, TagSet{Comp: new(false)}},
	})
}

// TestProjectLists grades the list shape: a list-valued field reads
// every value under its first style holding any, in on-disk order,
// while a single-valued field over the same repeated key reads the
// first value only.
func TestProjectLists(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"every value under the key", []string{"GENRE=Ambient", "GENRE=Drone", "GENRE=Field Recording"}, TagSet{Genres: []string{"Ambient", "Drone", "Field Recording"}}},
		{"one value is a one-element list", []string{"REMIXER=Invented Remixer"}, TagSet{Remixers: []string{"Invented Remixer"}}},
		{"a single-valued field over a repeated key reads the first value", []string{"MUSICBRAINZ_ARTISTID=first-id", "MUSICBRAINZ_ARTISTID=second-id"}, TagSet{MBArtistid: new("first-id"), MBArtistids: []string{"first-id", "second-id"}}},
		{"values keep their on-disk order across styles", []string{"ALBUMARTISTS=Y", "ALBUMARTISTS=X"}, TagSet{Albumartists: []string{"Y", "X"}}},
	})
}

// TestProjectAbsentKeys grades the read half of C2's four-state rule
// in one equality: safe_cast of an absent value is absent for every
// type (INV:253-260), so a container with no comments at all projects
// to the zero TagSet rather than to 69 typed zeros. The empty list
// has no separate case here because it cannot have one: a list field
// with no values writes no comment, so an empty list and an absent
// key leave identical bytes.
func TestProjectAbsentKeys(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"no comments at all", nil, TagSet{}},
		{"only foreign keys", []string{"WAVEFORMATEXTENSIBLE_CHANNEL_MASK=0x3", "ENCODER_SETTINGS=-8"}, TagSet{}},
	})
}

// TestProjectComparesNamesCaseInsensitively grades the RFC 9639
// section 8.6 rule C8 states for the writer, on the read side: a
// container tagged in lowercase by another tagger projects the same
// values as one tagged in the conventional uppercase.
func TestProjectComparesNamesCaseInsensitively(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"lowercase names", []string{"title=T", "genre=Ambient", "track=3"}, TagSet{Title: new("T"), Genres: []string{"Ambient"}, Track: new(3)}},
		{"mixed case names", []string{"Album=Al"}, TagSet{Album: new("Al")}},
	})
}

// TestProjectReplacesInvalidUTF8 grades ruling 10 on the read path: a
// comment value carrying bytes that are not valid UTF-8 projects with
// each ill-formed sequence replaced by U+FFFD, the way mutagen's
// errors='replace' decode hands mediafile its value. The want strings
// are literals, never computed from replaceInvalidUTF8.
func TestProjectReplacesInvalidUTF8(t *testing.T) {
	runProjectCases(t, []projectCase{
		{"lone continuation byte", []string{"TITLE=a\x80b"}, TagSet{Title: new("a�b")}},
		{"truncated three-byte sequence", []string{"GENRE=\xe2\x82"}, TagSet{Genres: []string{"�"}}},
	})
}

// TestReadReturnsParseSentinels grades C14 through the exported entry
// point: a caller branching on either sentinel branches on the same
// error whether it parsed the container itself or read its tags.
func TestReadReturnsParseSentinels(t *testing.T) {
	blocks := wellFormedBlocks()
	blocks[1].Payload = vorbisCommentPayload("tagio test", []string{"TITLE=A"})
	second := blocks[1]
	malformedChain := append([]testBlock{blocks[0], blocks[1], second}, blocks[2:]...)

	tests := []struct {
		name string
		data []byte
		want error
	}{
		{"not a FLAC stream", []byte("ID3\x04rest of a tagged MP3"), ErrUnsupportedContainer},
		{"two VORBIS_COMMENT blocks", buildContainer(malformedChain, nil), ErrMalformedContainer},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts, err := Read(writeFixture(t, tt.data))
			if !errors.Is(err, tt.want) {
				t.Fatalf("Read error = %v, want %v", err, tt.want)
			}
			if !reflect.DeepEqual(ts, TagSet{}) {
				t.Errorf("refused read returned %s, want the zero TagSet", diffFields(ts, TagSet{}))
			}
		})
	}
}
