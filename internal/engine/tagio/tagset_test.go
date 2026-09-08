package tagio

import (
	"slices"
	"testing"
)

// handoffNames is corpus/README.md's handoff section (:2354 onward),
// the 69 writable_fields names reproduced verbatim from :2381 in the
// order the ledger's own header lists them. C1 names this section the
// single source of record for the 69 names; this list is copied by
// hand from the doc rather than derived from Mapping, so a dropped or
// renamed row in Mapping has something independent to fail against.
var handoffNames = []string{
	"acoustid_fingerprint", "acoustid_id", "album", "albumartist",
	"albumartist_credit", "albumartist_sort", "albumartists",
	"albumartists_credit", "albumartists_sort", "albumdisambig",
	"albumstatus", "albumtype", "albumtypes", "arrangers", "artist",
	"artist_credit", "artist_sort", "artists", "artists_credit",
	"artists_sort", "asin", "barcode", "bpm", "catalognum", "comments",
	"comp", "composer_sort", "composers", "country", "day", "disc",
	"disctitle", "disctotal", "encoder", "genres", "grouping",
	"initial_key", "isrc", "label", "language", "lyricists", "lyrics",
	"mb_albumartistid", "mb_albumartistids", "mb_albumid",
	"mb_artistid", "mb_artistids", "mb_releasegroupid",
	"mb_releasetrackid", "mb_trackid", "mb_workid", "media", "month",
	"original_day", "original_month", "original_year",
	"r128_album_gain", "r128_track_gain", "remixers", "rg_album_gain",
	"rg_album_peak", "rg_track_gain", "rg_track_peak", "script",
	"subtitle", "title", "track", "tracktotal", "year",
}

// TestFieldCountAndNames grades C1's field count and its single source
// of record. It compares Mapping's field names, as a set, against
// handoffNames, and reports the count and any name only one side
// carries, so a dropped field (mutation 1) names itself in the
// failure rather than surfacing as a bare count mismatch.
func TestFieldCountAndNames(t *testing.T) {
	if len(handoffNames) != 69 {
		t.Fatalf("handoffNames has %d entries, want 69", len(handoffNames))
	}
	if len(Mapping) != 69 {
		t.Errorf("Mapping has %d rows, want 69", len(Mapping))
	}

	got := make(map[string]bool, len(Mapping))
	for _, row := range Mapping {
		got[string(row.Field)] = true
	}
	want := make(map[string]bool, len(handoffNames))
	for _, n := range handoffNames {
		want[n] = true
	}

	for n := range want {
		if !got[n] {
			t.Errorf("handoff names %q, Mapping does not", n)
		}
	}
	for n := range got {
		if !want[n] {
			t.Errorf("Mapping names %q, the handoff does not", n)
		}
	}
}

// TestMappingCitations walks Mapping and grades C16: every row carries
// at least one Vorbis key and a non-empty mediafile citation.
func TestMappingCitations(t *testing.T) {
	for _, row := range Mapping {
		if len(row.Styles) == 0 {
			t.Errorf("field %q has no Vorbis keys", row.Field)
		}
		for _, s := range row.Styles {
			if s.Key == "" {
				t.Errorf("field %q has a style with an empty key", row.Field)
			}
		}
		if row.Citation == "" {
			t.Errorf("field %q has no mediafile citation", row.Field)
		}
	}
}

// TestReadOnlyMarks grades the exceptional-rows table's read-only
// marks for albumartists, catalognum and barcode: the styles named
// read-only there, and no other style on those three rows, are marked
// Writable: false.
func TestReadOnlyMarks(t *testing.T) {
	cases := []struct {
		field        Field
		readOnlyKeys []string
	}{
		{FieldAlbumartists, []string{"ALBUM ARTISTS"}},
		{FieldCatalognum, []string{"CATALOGID", "DISCOGS_CATALOG"}},
		{FieldBarcode, []string{"UPC", "EAN/UPN", "EAN", "UPN"}},
	}

	for _, c := range cases {
		row := mustRow(t, c.field)
		readOnly := make(map[string]bool, len(c.readOnlyKeys))
		for _, k := range c.readOnlyKeys {
			readOnly[k] = true
		}
		for _, s := range row.Styles {
			want := !readOnly[s.Key]
			if s.Writable != want {
				t.Errorf("field %q key %q: Writable = %v, want %v", c.field, s.Key, s.Writable, want)
			}
		}
	}
}

// TestWriteOrderDeterministic grades C3's determinism property:
// WriteOrder called repeatedly always yields the same sequence. A
// map-backed implementation (mutation 5) ranges non-deterministically
// per call, so this loop is the check that would catch it; the loop
// count is the number of calls this task ran the check for.
func TestWriteOrderDeterministic(t *testing.T) {
	const runs = 200
	first := WriteOrder()
	if len(first) != 69 {
		t.Fatalf("WriteOrder returned %d fields, want 69", len(first))
	}
	for i := 1; i < runs; i++ {
		next := WriteOrder()
		if !slices.Equal(first, next) {
			t.Fatalf("WriteOrder is non-deterministic: run %d differs from run 0", i)
		}
	}
}

// TestCollidingPairsWriteOrder grades C3's pair-resolution rule: each
// of the three colliding pairs writes its plural member last, so the
// plural's position in WriteOrder follows the singular's.
func TestCollidingPairsWriteOrder(t *testing.T) {
	pairs := []struct {
		singular, plural Field
	}{
		{FieldAlbumtype, FieldAlbumtypes},
		{FieldMBArtistid, FieldMBArtistids},
		{FieldMBAlbumartistid, FieldMBAlbumartistids},
	}

	order := WriteOrder()
	index := make(map[Field]int, len(order))
	for i, f := range order {
		index[f] = i
	}

	for _, p := range pairs {
		si, ok := index[p.singular]
		if !ok {
			t.Fatalf("%q is not in WriteOrder", p.singular)
		}
		pi, ok := index[p.plural]
		if !ok {
			t.Fatalf("%q is not in WriteOrder", p.plural)
		}
		if si >= pi {
			t.Errorf("pair (%q, %q): singular is at %d, plural at %d, want singular before plural", p.singular, p.plural, si, pi)
		}
	}
}

// TestFourStates grades C2 for a representative string field, list
// field and numeric field: value, present-but-empty and absent are
// each distinguishable by reading the one TagSet member directly, no
// second lookup. The numeric row has no empty state, per C2.
func TestFourStates(t *testing.T) {
	value := "Wowbagger"
	t.Run("string", func(t *testing.T) {
		cases := map[string]*TagSet{
			"value":  {Title: &value},
			"empty":  {Title: EmptyString()},
			"absent": {Title: nil},
		}
		if cases["value"].Title == nil || *cases["value"].Title != value {
			t.Error("value state: Title does not hold the value")
		}
		if cases["empty"].Title == nil || *cases["empty"].Title != "" {
			t.Error("empty state: Title is not a non-nil pointer to \"\"")
		}
		if cases["absent"].Title != nil {
			t.Error("absent state: Title is not nil")
		}
	})

	t.Run("list", func(t *testing.T) {
		cases := map[string]*TagSet{
			"value":  {Genres: []string{"jazz", "trip-hop"}},
			"empty":  {Genres: EmptyList()},
			"absent": {},
		}
		if cases["value"].Genres == nil || len(cases["value"].Genres) != 2 {
			t.Error("value state: Genres does not hold the values")
		}
		if cases["empty"].Genres == nil || len(cases["empty"].Genres) != 0 {
			t.Error("empty state: Genres is not a non-nil zero-length slice")
		}
		if cases["absent"].Genres != nil {
			t.Error("absent state: Genres is not nil")
		}
	})

	t.Run("numeric", func(t *testing.T) {
		n := 7
		cases := map[string]*TagSet{
			"value":  {Track: &n},
			"absent": {},
		}
		if cases["value"].Track == nil || *cases["value"].Track != 7 {
			t.Error("value state: Track does not hold the value")
		}
		if cases["absent"].Track != nil {
			t.Error("absent state: Track is not nil")
		}
	})
}

// mustRow returns Mapping's row for f, failing the test if none
// exists.
func mustRow(t *testing.T, f Field) FieldMapping {
	t.Helper()
	for _, row := range Mapping {
		if row.Field == f {
			return row
		}
	}
	t.Fatalf("no Mapping row for field %q", f)
	return FieldMapping{}
}
