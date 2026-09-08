package tagio

import (
	"slices"
	"testing"
)

// TestSerializeFloatPlaces grades float_places (2 by default, 6 on
// both ReplayGain peaks) and the " dB" suffix that appears on the two
// gain fields and nowhere else (mediafile:836, :846, :856, :864).
// Every want string is hard-coded, not computed from rgGainField or
// rgPeakField.
func TestSerializeFloatPlaces(t *testing.T) {
	cases := []struct {
		name     string
		ts       TagSet
		wantKey  string
		wantVal  string
		citation string
	}{
		{"rg_track_gain two places with dB", TagSet{RgTrackGain: new(-7.5)}, "REPLAYGAIN_TRACK_GAIN", "-7.50 dB", "mediafile:836"},
		{"rg_album_gain two places with dB", TagSet{RgAlbumGain: new(3.0)}, "REPLAYGAIN_ALBUM_GAIN", "3.00 dB", "mediafile:846"},
		{"rg_track_peak six places no suffix", TagSet{RgTrackPeak: new(0.987654321)}, "REPLAYGAIN_TRACK_PEAK", "0.987654", "mediafile:856"},
		{"rg_album_peak six places no suffix", TagSet{RgAlbumPeak: new(1.0)}, "REPLAYGAIN_ALBUM_PEAK", "1.000000", "mediafile:864"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kw := mustKeyWrite(t, Serialize(c.ts), fieldFor(t, c.wantKey))
			if len(kw.Writes) != 1 || kw.Writes[0].Name != c.wantKey || kw.Writes[0].Value != c.wantVal {
				t.Errorf("%s: got %+v, want one comment %s=%q (%s)", c.name, kw.Writes, c.wantKey, c.wantVal, c.citation)
			}
		})
	}
}

// TestSerializeBool grades comp's "1"/"0" encoding, never "true" or
// "false" (mediafile:535).
func TestSerializeBool(t *testing.T) {
	cases := []struct {
		name string
		v    bool
		want string
	}{
		{"true writes 1", true, "1"},
		{"false writes 0", false, "0"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kw := mustKeyWrite(t, Serialize(TagSet{Comp: new(c.v)}), FieldComp)
			if len(kw.Writes) != 1 || kw.Writes[0].Name != "COMPILATION" || kw.Writes[0].Value != c.want {
				t.Errorf("got %+v, want one comment COMPILATION=%q", kw.Writes, c.want)
			}
		})
	}
}

// TestSerializeR128Shift grades the Q8 shift: r128_track_gain = -1.5
// writes R128_TRACK_GAIN=-384, INV's own measurement
// (mediafile:874, __init__.py:870-884).
func TestSerializeR128Shift(t *testing.T) {
	kw := mustKeyWrite(t, Serialize(TagSet{R128TrackGain: new(-1.5)}), FieldR128TrackGain)
	if len(kw.Writes) != 1 || kw.Writes[0].Name != "R128_TRACK_GAIN" || kw.Writes[0].Value != "-384" {
		t.Errorf("got %+v, want one comment R128_TRACK_GAIN=-384", kw.Writes)
	}
}

// TestSerializeDateTable grades date decomposition, the four-digit
// zero-padded year, the unpadded YEAR fallback, and the
// compose-as-a-unit rule for a partial triple (mediafile/fields.py:182-201).
// Every want is hard-coded against mediafile:693-695 (year/month/day)
// and mediafile:706-708 (the original_ trio, no YEAR fallback).
func TestSerializeDateTable(t *testing.T) {
	cases := []struct {
		name           string
		year           *int
		month          *int
		day            *int
		wantDate       string
		wantYear       string
		wantDateAbsent bool
	}{
		{"all components present", new(2001), new(3), new(4), "2001-03-04", "2001", false},
		{"truthy month, zero day", new(2001), new(3), new(0), "2001-03", "2001", false},
		{"zero month, truthy day", new(2001), new(0), new(15), "2001", "2001", false},
		{"every component zero", new(0), new(0), new(0), "0000", "0", false},
		{"absent year", nil, new(3), new(4), "", "", true},
		{"present year, absent month", new(2001), nil, nil, "2001", "2001", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kw := mustKeyWrite(t, Serialize(TagSet{Year: c.year, Month: c.month, Day: c.day}), FieldYear)
			if c.wantDateAbsent {
				if len(kw.Writes) != 0 {
					t.Fatalf("got writes %+v, want none", kw.Writes)
				}
				if !slices.Contains(kw.Deletes, "DATE") || !slices.Contains(kw.Deletes, "YEAR") {
					t.Errorf("deletes %v, want DATE and YEAR both removed", kw.Deletes)
				}
				return
			}
			gotDate, ok := commentValue(kw.Writes, "DATE")
			if !ok || gotDate != c.wantDate {
				t.Errorf("DATE = %q, ok=%v, want %q", gotDate, ok, c.wantDate)
			}
			gotYear, ok := commentValue(kw.Writes, "YEAR")
			if !ok || gotYear != c.wantYear {
				t.Errorf("YEAR = %q, ok=%v, want %q", gotYear, ok, c.wantYear)
			}
		})
	}
}

// TestSerializeOriginalDateEveryComponentZero grades
// ORIGINALDATE=0000 on the same all-zero row as DATE and YEAR, and
// that original_year carries no YEAR-style fallback (mediafile:706-708).
func TestSerializeOriginalDateEveryComponentZero(t *testing.T) {
	kw := mustKeyWrite(t, Serialize(TagSet{OriginalYear: new(0), OriginalMonth: new(0), OriginalDay: new(0)}), FieldOriginalYear)
	if len(kw.Writes) != 1 || kw.Writes[0].Name != "ORIGINALDATE" || kw.Writes[0].Value != "0000" {
		t.Errorf("got %+v, want one comment ORIGINALDATE=0000", kw.Writes)
	}
}

// TestSerializeListRepetition grades a two-element list writing its
// key twice, in list order, under every writable style of the field
// (mediafile:554-556): albumartists has two writable styles
// (ALBUMARTISTS, ALBUM_ARTISTS) and one read-only style (ALBUM
// ARTISTS) that a write never emits.
func TestSerializeListRepetition(t *testing.T) {
	kw := mustKeyWrite(t, Serialize(TagSet{Albumartists: []string{"X", "Y"}}), FieldAlbumartists)
	want := []Comment{
		{Name: "ALBUMARTISTS", Value: "X"},
		{Name: "ALBUMARTISTS", Value: "Y"},
		{Name: "ALBUM_ARTISTS", Value: "X"},
		{Name: "ALBUM_ARTISTS", Value: "Y"},
	}
	if !slices.Equal(kw.Writes, want) {
		t.Errorf("got %+v, want %+v", kw.Writes, want)
	}
	for _, c := range kw.Writes {
		if c.Name == "ALBUM ARTISTS" {
			t.Errorf("wrote read-only style ALBUM ARTISTS: %+v", c)
		}
	}
}

// TestSerializeEmptyList grades C2's third state: an empty list
// writes no comment at all, asserted against an empty comment
// sequence rather than against an empty string
// (ListStorageStyle.set_list, mediafile/storage/base.py:189-199).
func TestSerializeEmptyList(t *testing.T) {
	kw := mustKeyWrite(t, Serialize(TagSet{Genres: EmptyList()}), FieldGenres)
	if len(kw.Writes) != 0 {
		t.Errorf("got %d comments, want an empty comment sequence: %+v", len(kw.Writes), kw.Writes)
	}
}

// TestSerializeEmptyStringWritesComment grades C2's second state: an
// empty string writes a comment carrying an empty value, distinct
// from the empty-list state above that writes nothing (proves the two
// states are handled separately).
func TestSerializeEmptyStringWritesComment(t *testing.T) {
	kw := mustKeyWrite(t, Serialize(TagSet{Title: EmptyString()}), FieldTitle)
	want := []Comment{{Name: "TITLE", Value: ""}}
	if !slices.Equal(kw.Writes, want) {
		t.Errorf("got %+v, want %+v", kw.Writes, want)
	}
}

// TestSerializeDeleteBarcode grades INV:238-240: read_only blocks a
// set, not a delete, so deleting barcode removes all five keys
// (BARCODE, UPC, EAN/UPN, EAN, UPN), not just the writable one.
func TestSerializeDeleteBarcode(t *testing.T) {
	kw := mustKeyWrite(t, Serialize(TagSet{Barcode: nil}), FieldBarcode)
	if len(kw.Writes) != 0 {
		t.Errorf("got writes %+v, want none", kw.Writes)
	}
	want := []string{"BARCODE", "UPC", "EAN/UPN", "EAN", "UPN"}
	if !slices.Equal(kw.Deletes, want) {
		t.Errorf("deletes = %v, want %v", kw.Deletes, want)
	}
}

// TestSafeCastInt grades safe_cast's leading-numeric-prefix rule and
// its zero fallback (mediafile/utils/type_conversion.py:4-60).
func TestSafeCastInt(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"7", 7},
		{"-3", -3},
		{"+12", 12},
		{"5 dB", 5},
		{"abc", 0},
		{"", 0},
	}
	for _, c := range cases {
		if got := safeCastInt(c.in); got != c.want {
			t.Errorf("safeCastInt(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// TestSafeCastFloat grades the float path of the same rule, including
// the suffix strip: "-7.50 dB" casts to -7.5 because the prefix match
// stops at the space either way (mediafile/utils/type_conversion.py:52).
func TestSafeCastFloat(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"0.987654", 0.987654},
		{"-7.50 dB", -7.5},
		{"3", 3},
		{"nope", 0},
		{"", 0},
	}
	for _, c := range cases {
		if got := safeCastFloat(c.in); got != c.want {
			t.Errorf("safeCastFloat(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestSafeCastBool grades bool(int(val)) with a parse failure falling
// to false, the whole-string int() rather than a leading prefix.
func TestSafeCastBool(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"1", true},
		{"0", false},
		{"2", true},
		{"1x", false},
		{"", false},
	}
	for _, c := range cases {
		if got := safeCastBool(c.in); got != c.want {
			t.Errorf("safeCastBool(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

// TestReplaceInvalidUTF8 grades ruling 10: invalid UTF-8 is replaced
// with U+FFFD per code point on read, matching mutagen's
// errors='replace'. The input carries a lone continuation byte (0x80)
// between two valid runs; the want string is a hard-coded literal, not
// computed from replaceInvalidUTF8 or utf8.ToValidString.
func TestReplaceInvalidUTF8(t *testing.T) {
	in := "before\x80after"
	want := "before�after"
	if got := replaceInvalidUTF8(in); got != want {
		t.Errorf("replaceInvalidUTF8(%q) = %q, want %q", in, got, want)
	}
}

// mustKeyWrite returns Serialize's KeyWrite for field, failing the
// test if none exists.
func mustKeyWrite(t *testing.T, kws []KeyWrite, field Field) KeyWrite {
	t.Helper()
	for _, kw := range kws {
		if kw.Field == field {
			return kw
		}
	}
	t.Fatalf("no KeyWrite for field %q", field)
	return KeyWrite{}
}

// fieldFor maps a Vorbis key back to the one field in Mapping that
// owns it, for tests that key their case table on the Vorbis key
// rather than the field name.
func fieldFor(t *testing.T, vorbisKey string) Field {
	t.Helper()
	for _, row := range Mapping {
		for _, s := range row.Styles {
			if s.Key == vorbisKey {
				return row.Field
			}
		}
	}
	t.Fatalf("no Mapping row carries Vorbis key %q", vorbisKey)
	return ""
}

// commentValue returns the value of the first comment in cs named
// key, and whether one was found.
func commentValue(cs []Comment, key string) (string, bool) {
	for _, c := range cs {
		if c.Name == key {
			return c.Value, true
		}
	}
	return "", false
}
