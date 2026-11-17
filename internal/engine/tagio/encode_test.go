package tagio

import (
	"slices"
	"testing"
)

// TestSerializeFloatPlaces grades float_places (2 by default, 6 on
// both ReplayGain peaks) and the " dB" suffix that appears on the two
// gain fields and nowhere else (mediafile:836, :846, :856, :864).
// Every want string is hard-coded, not computed from formatGain or
// formatPeak.
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
			kw := mustKeyWrite(t, serialize(c.ts), fieldFor(t, c.wantKey))
			if len(kw.writes) != 1 || kw.writes[0].Name != c.wantKey || kw.writes[0].Value != c.wantVal {
				t.Errorf("%s: got %+v, want one comment %s=%q (%s)", c.name, kw.writes, c.wantKey, c.wantVal, c.citation)
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
			kw := mustKeyWrite(t, serialize(TagSet{Comp: new(c.v)}), "comp")
			if len(kw.writes) != 1 || kw.writes[0].Name != "COMPILATION" || kw.writes[0].Value != c.want {
				t.Errorf("got %+v, want one comment COMPILATION=%q", kw.writes, c.want)
			}
		})
	}
}

// TestSerializeR128Shift grades the Q8 shift: r128_track_gain = -1.5
// writes R128_TRACK_GAIN=-384 (mediafile:874, __init__.py:870-884).
func TestSerializeR128Shift(t *testing.T) {
	kw := mustKeyWrite(t, serialize(TagSet{R128TrackGain: new(-1.5)}), "r128_track_gain")
	if len(kw.writes) != 1 || kw.writes[0].Name != "R128_TRACK_GAIN" || kw.writes[0].Value != "-384" {
		t.Errorf("got %+v, want one comment R128_TRACK_GAIN=-384", kw.writes)
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
			kw := mustKeyWrite(t, serialize(TagSet{Year: c.year, Month: c.month, Day: c.day}), "year")
			if c.wantDateAbsent {
				if len(kw.writes) != 0 {
					t.Fatalf("got writes %+v, want none", kw.writes)
				}
				if !slices.Contains(kw.deletes, "DATE") || !slices.Contains(kw.deletes, "YEAR") {
					t.Errorf("deletes %v, want DATE and YEAR both removed", kw.deletes)
				}
				return
			}
			gotDate, ok := commentValue(kw.writes, "DATE")
			if !ok || gotDate != c.wantDate {
				t.Errorf("DATE = %q, ok=%v, want %q", gotDate, ok, c.wantDate)
			}
			gotYear, ok := commentValue(kw.writes, "YEAR")
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
	kw := mustKeyWrite(t, serialize(TagSet{OriginalYear: new(0), OriginalMonth: new(0), OriginalDay: new(0)}), "original_year")
	if len(kw.writes) != 1 || kw.writes[0].Name != "ORIGINALDATE" || kw.writes[0].Value != "0000" {
		t.Errorf("got %+v, want one comment ORIGINALDATE=0000", kw.writes)
	}
}

// TestSerializeListRepetition grades a two-element list writing its
// key twice, in list order, under every writable style of the field
// (mediafile:554-556): albumartists has two writable styles
// (ALBUMARTISTS, ALBUM_ARTISTS) and one read-only style (ALBUM
// ARTISTS) that a write never emits.
func TestSerializeListRepetition(t *testing.T) {
	kw := mustKeyWrite(t, serialize(TagSet{Albumartists: []string{"X", "Y"}}), "albumartists")
	want := []Comment{
		{Name: "ALBUMARTISTS", Value: "X"},
		{Name: "ALBUMARTISTS", Value: "Y"},
		{Name: "ALBUM_ARTISTS", Value: "X"},
		{Name: "ALBUM_ARTISTS", Value: "Y"},
	}
	if !slices.Equal(kw.writes, want) {
		t.Errorf("got %+v, want %+v", kw.writes, want)
	}
	for _, c := range kw.writes {
		if c.Name == "ALBUM ARTISTS" {
			t.Errorf("wrote read-only style ALBUM ARTISTS: %+v", c)
		}
	}
}

// TestSerializeEmptyList grades the empty-list state: an empty list
// writes no comment at all, asserted against an empty comment
// sequence rather than against an empty string
// (ListStorageStyle.set_list, mediafile/storage/base.py:189-199).
func TestSerializeEmptyList(t *testing.T) {
	kw := mustKeyWrite(t, serialize(TagSet{Genres: []string{}}), "genres")
	if len(kw.writes) != 0 {
		t.Errorf("got %d comments, want an empty comment sequence: %+v", len(kw.writes), kw.writes)
	}
}

// TestSerializeEmptyStringWritesComment grades the empty-string state:
// an empty string writes a comment carrying an empty value, distinct
// from the empty-list state above that writes nothing (proves the two
// states are handled separately).
func TestSerializeEmptyStringWritesComment(t *testing.T) {
	kw := mustKeyWrite(t, serialize(TagSet{Title: new("")}), "title")
	want := []Comment{{Name: "TITLE", Value: ""}}
	if !slices.Equal(kw.writes, want) {
		t.Errorf("got %+v, want %+v", kw.writes, want)
	}
}

// TestSerializeDeleteBarcode grades the delete rule: read_only blocks a
// set, not a delete, so deleting barcode removes all five keys
// (BARCODE, UPC, EAN/UPN, EAN, UPN), not just the writable one.
func TestSerializeDeleteBarcode(t *testing.T) {
	kw := mustKeyWrite(t, serialize(TagSet{Barcode: nil}), "barcode")
	if len(kw.writes) != 0 {
		t.Errorf("got writes %+v, want none", kw.writes)
	}
	want := []string{"BARCODE", "UPC", "EAN/UPN", "EAN", "UPN"}
	if !slices.Equal(kw.deletes, want) {
		t.Errorf("deletes = %v, want %v", kw.deletes, want)
	}
}

// mustKeyWrite returns serialize's keyWrite for field, failing the
// test if none exists.
func mustKeyWrite(t *testing.T, kws []keyWrite, field string) keyWrite {
	t.Helper()
	for _, kw := range kws {
		if kw.field == field {
			return kw
		}
	}
	t.Fatalf("no keyWrite for field %q", field)
	return keyWrite{}
}

// fieldFor maps a Vorbis key back to the first field table row that
// owns it, for tests that key their case table on the Vorbis key
// rather than the field name.
func fieldFor(t *testing.T, vorbisKey string) string {
	t.Helper()
	for _, f := range fields {
		for _, s := range f.styles {
			if s.key == vorbisKey {
				return f.name
			}
		}
	}
	t.Fatalf("no field table row carries Vorbis key %q", vorbisKey)
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
