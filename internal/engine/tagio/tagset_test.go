package tagio

import (
	"reflect"
	"strings"
	"testing"
)

// handoffNames is corpus/README.md's handoff section (:2354 onward),
// the 69 writable_fields names reproduced verbatim from :2381 in the
// order the ledger's own header lists them. C1 names this section the
// single source of record for the 69 names; this list is copied by
// hand from the doc rather than derived from the field table, so a
// dropped or renamed row has something independent to fail against.
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
// of record. It compares the field table's names, as a set, against
// handoffNames, and reports the count and any name only one side
// carries, so a dropped field names itself in the failure rather than
// surfacing as a bare count mismatch.
func TestFieldCountAndNames(t *testing.T) {
	if len(handoffNames) != 69 {
		t.Fatalf("handoffNames has %d entries, want 69", len(handoffNames))
	}
	if len(fields) != 69 {
		t.Errorf("fields has %d rows, want 69", len(fields))
	}

	got := make(map[string]bool, len(fields))
	for _, f := range fields {
		got[string(f.name)] = true
	}
	want := make(map[string]bool, len(handoffNames))
	for _, n := range handoffNames {
		want[n] = true
	}

	for n := range want {
		if !got[n] {
			t.Errorf("handoff names %q, the field table does not", n)
		}
	}
	for n := range got {
		if !want[n] {
			t.Errorf("the field table names %q, the handoff does not", n)
		}
	}
}

// TestFieldTableCitations grades every row of the field table: each
// carries at least one Vorbis key and a non-empty mediafile citation.
func TestFieldTableCitations(t *testing.T) {
	for _, f := range fields {
		if len(f.styles) == 0 {
			t.Errorf("field %q has no Vorbis keys", f.name)
		}
		for _, s := range f.styles {
			if s.key == "" {
				t.Errorf("field %q has a style with an empty key", f.name)
			}
		}
		if f.cite == "" {
			t.Errorf("field %q has no mediafile citation", f.name)
		}
	}
}

// TestFieldAccessorsTieToMembers grades the one fact the field table's
// literal cannot show by eye: each row's codec reads and writes the
// TagSet member whose JSON tag is the row's name. For every member in
// turn it sets that member alone, by reflection, and requires that
// Serialize writes comments for exactly the row of the same name, then
// that projecting those comments sets that member alone again. A
// month or day member writes nothing by itself, because the date
// comment is keyed on the year, so those four rows expect no write.
func TestFieldAccessorsTieToMembers(t *testing.T) {
	typ := reflect.TypeFor[TagSet]()
	for i := range typ.NumField() {
		member := typ.Field(i)
		name := member.Tag.Get("json")
		t.Run(name, func(t *testing.T) {
			var ts TagSet
			v := reflect.ValueOf(&ts).Elem().Field(i)
			switch member.Type.Kind() {
			case reflect.Slice:
				v.Set(reflect.ValueOf([]string{"7"}))
			case reflect.Pointer:
				p := reflect.New(member.Type.Elem())
				switch member.Type.Elem().Kind() {
				case reflect.String:
					p.Elem().SetString("7")
				case reflect.Int:
					p.Elem().SetInt(7)
				case reflect.Float64:
					p.Elem().SetFloat(0.5)
				case reflect.Bool:
					p.Elem().SetBool(true)
				}
				v.Set(p)
			default:
				t.Fatalf("TagSet.%s has kind %v, want a slice or a pointer", member.Name, member.Type.Kind())
			}

			var written []Comment
			var writers []Field
			for _, kw := range Serialize(ts) {
				if len(kw.Writes) > 0 {
					writers = append(writers, kw.Field)
					written = append(written, kw.Writes...)
				}
			}
			dateOnly := map[string]bool{"month": true, "day": true, "original_month": true, "original_day": true}
			if dateOnly[name] {
				if len(writers) != 0 {
					t.Fatalf("setting %s alone wrote fields %v, want none", name, writers)
				}
				return
			}
			if len(writers) != 1 || string(writers[0]) != name {
				t.Fatalf("setting %s alone wrote fields %v, want [%s]", name, writers, name)
			}

			// The three colliding pairs share their keys, so reading
			// either member's comments back sets its partner too.
			partner := map[string]string{
				"albumtype": "albumtypes", "albumtypes": "albumtype",
				"mb_artistid": "mb_artistids", "mb_artistids": "mb_artistid",
				"mb_albumartistid": "mb_albumartistids", "mb_albumartistids": "mb_albumartistid",
			}
			got := reflect.ValueOf(project(Container{VorbisComment: VorbisComment{Comments: written}}))
			for j := range typ.NumField() {
				other := typ.Field(j).Tag.Get("json")
				want := j == i || other == partner[name]
				if set := !got.Field(j).IsNil(); set != want {
					t.Errorf("reading %s's comments back: TagSet.%s set = %v, want %v", name, typ.Field(j).Name, set, want)
				}
			}
		})
	}
}

// fieldStructName pairs each Field constant with the TagSet struct
// field it names, hand-authored independent of both tagset.go's
// struct literal and the field table: naming a removed Field constant
// here fails to compile, which is the mechanical tie a dropped
// constant needs. TestTagSetMembersTieToFieldTable compares this map's
// keys against the table's names and its values against TagSet's own
// field names by reflection, so a dropped struct member or a dropped
// row each fail a named assertion below, not just a bare count.
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

// TestTagSetMembersTieToFieldTable grades the three lists that must
// agree: removing a TagSet member, a table row or a Field constant
// each fail a named assertion here. A removed Field constant fails
// fieldStructName's own compilation, before the test ever runs; the
// two loops below catch a removed struct member (against
// reflect.TypeOf(TagSet{})) and a removed row (against the table's
// own names).
func TestTagSetMembersTieToFieldTable(t *testing.T) {
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

	gotTable := make(map[Field]bool, len(fields))
	for _, f := range fields {
		gotTable[f.name] = true
	}
	for f := range fieldStructName {
		if !gotTable[f] {
			t.Errorf("fieldStructName names %q, the field table does not", f)
		}
	}
	for f := range gotTable {
		if _, ok := fieldStructName[f]; !ok {
			t.Errorf("the field table names %q, fieldStructName does not", f)
		}
	}
}

// TestTagSetJSONTagsTieToFieldConstants grades R11's mandatory safety
// net in two halves, because a wrong tag is otherwise invisible: the
// conformance instrument's fill and its projection both go through
// the same tag, so a tag that maps album to the artist member writes
// ARTIST and reads it back under album, and the comparison agrees
// while the file on disk is wrong.
//
// By name: for each entry in fieldStructName, the named TagSet
// member's JSON tag, read through reflection, equals the Field
// constant that names it. This ties the tag, the member and the
// table row together by fieldStructName's one literal rather than by
// struct declaration order.
//
// By position: for each index, the struct field at that index's JSON
// tag equals the name of the table's row at the same index, and the
// member count equals 69. This is the cheaper, position-based check
// that fails loudly if TagSet's declaration order ever drifts from
// the table's.
func TestTagSetJSONTagsTieToFieldConstants(t *testing.T) {
	typ := reflect.TypeFor[TagSet]()

	t.Run("by name", func(t *testing.T) {
		for field, structName := range fieldStructName {
			structField, ok := typ.FieldByName(structName)
			if !ok {
				t.Errorf("TagSet has no field named %q (fieldStructName[%q])", structName, field)
				continue
			}
			tag, ok := structField.Tag.Lookup("json")
			if !ok {
				t.Errorf("TagSet.%s carries no json tag, want %q", structName, field)
				continue
			}
			if tag != string(field) {
				t.Errorf("TagSet.%s json tag = %q, want %q (the Field constant naming it)", structName, tag, field)
			}
		}
	})

	t.Run("by position", func(t *testing.T) {
		if typ.NumField() != 69 {
			t.Fatalf("TagSet has %d fields, want 69", typ.NumField())
		}
		if len(fields) != 69 {
			t.Fatalf("fields has %d rows, want 69", len(fields))
		}
		for i, f := range fields {
			structField := typ.Field(i)
			tag, ok := structField.Tag.Lookup("json")
			if !ok {
				t.Errorf("TagSet field %d (%s) carries no json tag, want %q", i, structField.Name, f.name)
				continue
			}
			if tag != string(f.name) {
				t.Errorf("TagSet field %d (%s) json tag = %q, want %q (fields[%d].name)", i, structField.Name, tag, f.name, i)
			}
		}
	})
}

// TestTagSetJSONTagsCarryNoOmitempty grades R11: every ledger row
// carries all 69 keys, absent values included as explicit nulls, so
// an omitempty tag would drop an absent member from the projection
// and fail the conformance instrument's 69-key count on every
// subject carrying that field as null.
func TestTagSetJSONTagsCarryNoOmitempty(t *testing.T) {
	typ := reflect.TypeFor[TagSet]()
	if typ.NumField() != 69 {
		t.Fatalf("TagSet has %d fields, want 69", typ.NumField())
	}
	for f := range typ.Fields() {
		tag, ok := f.Tag.Lookup("json")
		if !ok {
			t.Errorf("TagSet.%s carries no json tag", f.Name)
			continue
		}
		if strings.Contains(tag, "omitempty") {
			t.Errorf("TagSet.%s json tag %q carries omitempty, want none", f.Name, tag)
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
		for _, s := range row.styles {
			want := !readOnly[s.key]
			if s.writable != want {
				t.Errorf("field %q key %q: writable = %v, want %v", c.field, s.key, s.writable, want)
			}
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

// mustRow returns the field table's row for name, failing the test if
// none exists.
func mustRow(t *testing.T, name Field) field {
	t.Helper()
	for _, f := range fields {
		if f.name == name {
			return f
		}
	}
	t.Fatalf("no field table row for %q", name)
	return field{}
}
