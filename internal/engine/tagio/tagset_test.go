package tagio

import (
	"reflect"
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
	if len(mapping) != 69 {
		t.Errorf("mapping has %d rows, want 69", len(mapping))
	}

	got := make(map[string]bool, len(mapping))
	for _, row := range mapping {
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
	for _, row := range mapping {
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

// fieldStructName pairs each Field constant with the TagSet struct
// field it names, hand-authored independent of both tagset.go's
// struct literal and mapping: naming a removed Field constant here
// fails to compile, which is the mechanical tie a dropped constant
// needs. TestTagSetMembersTieToMapping compares this map's keys
// against mapping's Field set and its values against TagSet's own
// field names by reflection, so a dropped struct member or a dropped
// mapping row each fail a named assertion below, not just a bare
// count.
var fieldStructName = map[Field]string{
	FieldAcoustIDFingerprint: "AcoustIDFingerprint",
	FieldAcoustIDID:          "AcoustIDID",
	FieldAlbum:               "Album",
	FieldAlbumartist:         "Albumartist",
	FieldAlbumartistCredit:   "AlbumartistCredit",
	FieldAlbumartistSort:     "AlbumartistSort",
	FieldAlbumartists:        "Albumartists",
	FieldAlbumartistsCredit:  "AlbumartistsCredit",
	FieldAlbumartistsSort:    "AlbumartistsSort",
	FieldAlbumdisambig:       "Albumdisambig",
	FieldAlbumstatus:         "Albumstatus",
	FieldAlbumtype:           "Albumtype",
	FieldAlbumtypes:          "Albumtypes",
	FieldArrangers:           "Arrangers",
	FieldArtist:              "Artist",
	FieldArtistCredit:        "ArtistCredit",
	FieldArtistSort:          "ArtistSort",
	FieldArtists:             "Artists",
	FieldArtistsCredit:       "ArtistsCredit",
	FieldArtistsSort:         "ArtistsSort",
	FieldAsin:                "Asin",
	FieldBarcode:             "Barcode",
	FieldBpm:                 "Bpm",
	FieldCatalognum:          "Catalognum",
	FieldComments:            "Comments",
	FieldComp:                "Comp",
	FieldComposerSort:        "ComposerSort",
	FieldComposers:           "Composers",
	FieldCountry:             "Country",
	FieldYear:                "Year",
	FieldMonth:               "Month",
	FieldDay:                 "Day",
	FieldDisc:                "Disc",
	FieldDisctitle:           "Disctitle",
	FieldDisctotal:           "Disctotal",
	FieldEncoder:             "Encoder",
	FieldGenres:              "Genres",
	FieldGrouping:            "Grouping",
	FieldInitialKey:          "InitialKey",
	FieldIsrc:                "Isrc",
	FieldLabel:               "Label",
	FieldLanguage:            "Language",
	FieldLyricists:           "Lyricists",
	FieldLyrics:              "Lyrics",
	FieldMBAlbumartistid:     "MBAlbumartistid",
	FieldMBAlbumartistids:    "MBAlbumartistids",
	FieldMBAlbumid:           "MBAlbumid",
	FieldMBArtistid:          "MBArtistid",
	FieldMBArtistids:         "MBArtistids",
	FieldMBReleasegroupid:    "MBReleasegroupid",
	FieldMBReleasetrackid:    "MBReleasetrackid",
	FieldMBTrackid:           "MBTrackid",
	FieldMBWorkid:            "MBWorkid",
	FieldMedia:               "Media",
	FieldOriginalYear:        "OriginalYear",
	FieldOriginalMonth:       "OriginalMonth",
	FieldOriginalDay:         "OriginalDay",
	FieldR128AlbumGain:       "R128AlbumGain",
	FieldR128TrackGain:       "R128TrackGain",
	FieldRemixers:            "Remixers",
	FieldRgAlbumGain:         "RgAlbumGain",
	FieldRgAlbumPeak:         "RgAlbumPeak",
	FieldRgTrackGain:         "RgTrackGain",
	FieldRgTrackPeak:         "RgTrackPeak",
	FieldScript:              "Script",
	FieldSubtitle:            "Subtitle",
	FieldTitle:               "Title",
	FieldTrack:               "Track",
	FieldTracktotal:          "Tracktotal",
}

// TestTagSetMembersTieToMapping grades finding 1: removing a TagSet
// member, a mapping row or a Field constant each fail a named
// assertion here. A removed Field constant fails fieldStructName's
// own compilation, before the test ever runs; the two loops below
// catch a removed struct member (against reflect.TypeOf(TagSet{}))
// and a removed mapping row (against mapping's own Field set).
func TestTagSetMembersTieToMapping(t *testing.T) {
	if len(fieldStructName) != 69 {
		t.Fatalf("fieldStructName has %d entries, want 69", len(fieldStructName))
	}

	typ := reflect.TypeFor[TagSet]()
	gotStruct := make(map[string]bool, typ.NumField())
	for f := range typ.Fields() {
		gotStruct[f.Name] = true
	}
	if typ.NumField() != 69 {
		t.Errorf("TagSet has %d exported fields, want 69", typ.NumField())
	}
	wantStruct := make(map[string]bool, len(fieldStructName))
	for _, name := range fieldStructName {
		wantStruct[name] = true
	}
	for name := range wantStruct {
		if !gotStruct[name] {
			t.Errorf("fieldStructName names TagSet field %q, TagSet does not have it", name)
		}
	}
	for name := range gotStruct {
		if !wantStruct[name] {
			t.Errorf("TagSet has field %q, fieldStructName does not name it", name)
		}
	}

	gotMapping := make(map[Field]bool, len(mapping))
	for _, row := range mapping {
		gotMapping[row.Field] = true
	}
	for f := range fieldStructName {
		if !gotMapping[f] {
			t.Errorf("fieldStructName names %q, mapping does not", f)
		}
	}
	for f := range gotMapping {
		if _, ok := fieldStructName[f]; !ok {
			t.Errorf("mapping names %q, fieldStructName does not", f)
		}
	}
}

// TestMappingAccessorReturnsCopy grades finding 2: a caller cannot
// reorder WriteOrder's underlying table by mutating Mapping's
// returned slice.
func TestMappingAccessorReturnsCopy(t *testing.T) {
	before := WriteOrder()
	got := Mapping()
	if len(got) < 2 {
		t.Fatal("Mapping() returned fewer than 2 rows")
	}
	got[0], got[1] = got[1], got[0]

	after := WriteOrder()
	if !slices.Equal(before, after) {
		t.Error("mutating Mapping()'s returned slice changed WriteOrder")
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
// second lookup. The numeric row has no empty state, per C2. Each
// subtest's rows are a table indexed by position, and the absent row
// is a bare TagSet{} rather than a field set to nil by name, so the
// assertion reads a value the row's own construction produced instead
// of re-reading a nil the same statement just wrote.
func TestFourStates(t *testing.T) {
	value := "Wowbagger"
	t.Run("string", func(t *testing.T) {
		rows := []TagSet{
			{Title: &value},
			{Title: EmptyString()},
			{},
		}
		if rows[0].Title == nil || *rows[0].Title != value {
			t.Error("value state: Title does not hold the value")
		}
		if rows[1].Title == nil || *rows[1].Title != "" {
			t.Error("empty state: Title is not a non-nil pointer to \"\"")
		}
		if rows[2].Title != nil {
			t.Error("absent state: Title is not nil")
		}
	})

	t.Run("list", func(t *testing.T) {
		rows := []TagSet{
			{Genres: []string{"jazz", "trip-hop"}},
			{Genres: EmptyList()},
			{},
		}
		if rows[0].Genres == nil || len(rows[0].Genres) != 2 {
			t.Error("value state: Genres does not hold the values")
		}
		if rows[1].Genres == nil || len(rows[1].Genres) != 0 {
			t.Error("empty state: Genres is not a non-nil zero-length slice")
		}
		if rows[2].Genres != nil {
			t.Error("absent state: Genres is not nil")
		}
	})

	t.Run("numeric", func(t *testing.T) {
		n := 7
		rows := []TagSet{
			{Track: &n},
			{},
		}
		if rows[0].Track == nil || *rows[0].Track != 7 {
			t.Error("value state: Track does not hold the value")
		}
		if rows[1].Track != nil {
			t.Error("absent state: Track is not nil")
		}
	})
}

// mustRow returns Mapping's row for f, failing the test if none
// exists.
func mustRow(t *testing.T, f Field) FieldMapping {
	t.Helper()
	for _, row := range mapping {
		if row.Field == f {
			return row
		}
	}
	t.Fatalf("no Mapping row for field %q", f)
	return FieldMapping{}
}
