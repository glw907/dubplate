package matcher

import "testing"

func TestMostCommonTags(t *testing.T) {
	items := []ItemFields{
		{
			Artist: "Northern Lines", Album: "Harbour Sound", AlbumArtist: "Northern Lines",
			Year: 2011, DiscTotal: 1, MBAlbumID: "mb-1", Label: "Cedarwood Records",
			Barcode: "111", CatalogNum: "CAT-001", Country: "GB", Media: "CD",
			AlbumDisambig: "Deluxe Edition", DataSource: "MusicBrainz",
		},
		{
			Artist: "Northern Lines", Album: "Harbour Sound", AlbumArtist: "Northern Lines",
			Year: 2011, DiscTotal: 1, MBAlbumID: "mb-1", Label: "Cedarwood Records",
			Barcode: "112", CatalogNum: "CAT-001", Country: "GB", Media: "CD",
			AlbumDisambig: "Deluxe Edition", DataSource: "MusicBrainz",
		},
		{
			Artist: "Northern Lines", Album: "Harbour Sound", AlbumArtist: "Northern Lines",
			Year: 2011, DiscTotal: 1, MBAlbumID: "mb-1", Label: "Cedarwood Records",
			Barcode: "113", CatalogNum: "CAT-001", Country: "GB", Media: "CD",
			AlbumDisambig: "Deluxe Edition", DataSource: "MusicBrainz",
		},
	}

	likelies, consensus := MostCommonTags(items)

	want := Likelies{
		Artist: "Northern Lines", Album: "Harbour Sound", AlbumArtist: "Northern Lines",
		Year: 2011, DiscTotal: 1, MBAlbumID: "mb-1", Label: "Cedarwood Records",
		Barcode: "111", CatalogNum: "CAT-001", Country: "GB", Media: "CD",
		AlbumDisambig: "Deluxe Edition", DataSource: "MusicBrainz",
	}
	if likelies != want {
		t.Errorf("MostCommonTags() likelies = %+v, want %+v", likelies, want)
	}

	for _, field := range []string{
		fieldArtist, fieldAlbum, fieldAlbumArtist, fieldYear, fieldDiscTotal,
		fieldMBAlbumID, fieldLabel, fieldCatalogNum, fieldCountry, fieldMedia,
		fieldAlbumDisambig, fieldDataSource,
	} {
		if !consensus[field] {
			t.Errorf("consensus[%q] = false, want true (every item agrees)", field)
		}
	}
	if consensus[fieldBarcode] {
		t.Error(`consensus["barcode"] = true, want false (every item carries a distinct barcode)`)
	}
}

// TestMostCommonTagsOverridesArtistFromAlbumartistConsensus proves
// util/__init__.py:843-844: a non-empty albumartist consensus wins
// over whatever the plain artist field's own plurality picked, even
// when that plurality is unambiguous on its own.
func TestMostCommonTagsOverridesArtistFromAlbumartistConsensus(t *testing.T) {
	items := []ItemFields{
		{Artist: "Feat X", AlbumArtist: "Northern Lines"},
		{Artist: "Feat Y", AlbumArtist: "Northern Lines"},
	}

	likelies, consensus := MostCommonTags(items)

	if !consensus[fieldAlbumArtist] {
		t.Fatal(`consensus["albumartist"] = false, want true`)
	}
	if likelies.Artist != "Northern Lines" {
		t.Errorf("likelies.Artist = %q, want the albumartist consensus %q", likelies.Artist, "Northern Lines")
	}
}

// TestMostCommonTagsNoOverrideWithoutAlbumartistConsensus proves the
// override's two guards: a disagreeing albumartist never overrides,
// and neither does an agreed-but-empty one (a VA release, where every
// item's albumartist is "").
func TestMostCommonTagsNoOverrideWithoutAlbumartistConsensus(t *testing.T) {
	t.Run("albumartist disagrees", func(t *testing.T) {
		items := []ItemFields{
			{Artist: "Feat X", AlbumArtist: "Northern Lines"},
			{Artist: "Feat X", AlbumArtist: "Southern Lines"},
		}
		likelies, consensus := MostCommonTags(items)
		if consensus[fieldAlbumArtist] {
			t.Fatal(`consensus["albumartist"] = true, want false`)
		}
		if likelies.Artist != "Feat X" {
			t.Errorf("likelies.Artist = %q, want the artist plurality %q", likelies.Artist, "Feat X")
		}
	})

	t.Run("albumartist consensus is empty", func(t *testing.T) {
		items := []ItemFields{
			{Artist: "Feat X", AlbumArtist: ""},
			{Artist: "Feat X", AlbumArtist: ""},
		}
		likelies, consensus := MostCommonTags(items)
		if !consensus[fieldAlbumArtist] {
			t.Fatal(`consensus["albumartist"] = false, want true`)
		}
		if likelies.Artist != "Feat X" {
			t.Errorf("likelies.Artist = %q, want the artist plurality %q", likelies.Artist, "Feat X")
		}
	})
}

// TestPluralityBreaksTiesByFirstSeen is contract 11's own test: a
// deliberately tied item set, where two values each occur exactly
// once and the winner must be the one item order saw first, not the
// one a randomized map iteration would happen to produce.
func TestPluralityBreaksTiesByFirstSeen(t *testing.T) {
	items := []ItemFields{
		{Label: "Second Sight Records"},
		{Label: "Cedarwood Records"},
	}

	likelies, _ := MostCommonTags(items)

	if likelies.Label != "Second Sight Records" {
		t.Errorf("likelies.Label = %q, want the first-seen tied value %q", likelies.Label, "Second Sight Records")
	}
}

func TestPluralityOfEmptyReturnsZeroAndZeroCount(t *testing.T) {
	value, count := plurality[string](nil)
	if value != "" || count != 0 {
		t.Errorf("plurality(nil) = (%q, %d), want (%q, 0)", value, count, "")
	}
}
