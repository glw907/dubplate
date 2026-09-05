package matcher

import "testing"

// TestProjectItemFields_TypedDefaults proves contract 10's typed-default
// projection: a field the raw bag omits reads beets' own type default
// ("" for a string field, 0 for an int or float field), not some
// "absent" sentinel. The fixture is hand-authored here, not read from
// any corpus case (M-d), and deliberately does not use data_source as
// its omitted-string example (fold, m6): year is used instead.
//
// A mutation that returns a sentinel for a missing field, for example
// intField returning -1 instead of 0, or stringField returning "unset"
// instead of "", flips this test from pass to fail, because want's
// omitted members are Go's own zero values and got must match them
// exactly.
func TestProjectItemFields_TypedDefaults(t *testing.T) {
	raw := map[string]any{
		"album":     "Ember Wharf",
		"disctotal": float64(2), // encoding/json decodes a JSON int as float64
		"length":    183.5,
		"track":     "not-a-number", // present but wrong-typed
		// artist, year, mb_trackid, data_source, and every other
		// ItemFields member are deliberately omitted.
	}

	got := ProjectItemFields(raw)

	want := ItemFields{
		Album:     "Ember Wharf",
		DiscTotal: 2,
		Length:    183.5,
		// Artist, Year, MBTrackID, DataSource, Track (wrong-typed), and
		// every other member stay at beets' typed default: "" or 0.
	}

	if got != want {
		t.Errorf("ProjectItemFields(%v) = %+v, want %+v", raw, got, want)
	}
}

func TestProjectItemFields_AllNineteenFields(t *testing.T) {
	raw := map[string]any{
		"artist":        "Alpha",
		"album":         "Beta",
		"albumartist":   "Gamma",
		"year":          float64(1998),
		"disctotal":     float64(2),
		"mb_albumid":    "album-id",
		"label":         "Delta Records",
		"barcode":       "012345",
		"catalognum":    "DR-001",
		"country":       "GB",
		"media":         "CD",
		"albumdisambig": "remaster",
		"data_source":   "MusicBrainz",
		"length":        321.75,
		"title":         "Epsilon",
		"track":         float64(4),
		"mb_trackid":    "track-id",
		"disc":          float64(1),
		"comp":          true,
	}

	want := ItemFields{
		Artist:        "Alpha",
		Album:         "Beta",
		AlbumArtist:   "Gamma",
		Year:          1998,
		DiscTotal:     2,
		MBAlbumID:     "album-id",
		Label:         "Delta Records",
		Barcode:       "012345",
		CatalogNum:    "DR-001",
		Country:       "GB",
		Media:         "CD",
		AlbumDisambig: "remaster",
		DataSource:    "MusicBrainz",
		Length:        321.75,
		Title:         "Epsilon",
		Track:         4,
		MBTrackID:     "track-id",
		Disc:          1,
		Comp:          true,
	}

	if got := ProjectItemFields(raw); got != want {
		t.Errorf("ProjectItemFields(%v) = %+v, want %+v", raw, got, want)
	}
}

// TestProjectItemFields_CompRejectsWrongType proves boolField's
// typed-default narrowing (mutation: boolField returning true for a
// non-bool value, for example treating a truthy string as true, must
// fail this test because want.Comp stays false).
func TestProjectItemFields_CompRejectsWrongType(t *testing.T) {
	got := ProjectItemFields(map[string]any{"comp": "yes"})
	if got.Comp != false {
		t.Errorf("ProjectItemFields({comp: %q}).Comp = %v, want false", "yes", got.Comp)
	}
}

func TestProjectItemFields_EmptyBagIsAllZeroValues(t *testing.T) {
	got := ProjectItemFields(map[string]any{})
	if got != (ItemFields{}) {
		t.Errorf("ProjectItemFields(empty) = %+v, want the zero ItemFields", got)
	}
}
