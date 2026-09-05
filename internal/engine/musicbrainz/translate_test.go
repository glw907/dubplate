package musicbrainz

import (
	"errors"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

func plainTrack(id string, position int, title string) Track {
	return Track{
		ID:       id,
		Position: wireInt(position),
		Number:   strconv.Itoa(position),
		Title:    title,
		Recording: Recording{
			ID:           "rec-" + id,
			Title:        title,
			ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Artist")},
		},
	}
}

// TestTranslateAlbum_PlainSingleMediumRelease exercises the composition
// end to end on a small, well-formed release: one medium, two tracks,
// a release-group original date, a label, and a catalog number.
func TestTranslateAlbum_PlainSingleMediumRelease(t *testing.T) {
	rel := &Release{
		ID:             "rel-plain",
		Title:          "Plain Album",
		Disambiguation: "remaster",
		ArtistCredit:   []CreditPart{simpleCreditPart("art-1", "Plain Artist")},
		Country:        new("XW"),
		Date:           "2001-03-04",
		LabelInfo: []LabelInfo{{
			CatalogNumber: new("CAT-1"),
			Label:         &Label{Name: "Plain Label"},
		}},
		ReleaseGroup: ReleaseGroup{FirstReleaseDate: "1999-01-01"},
		Media: []Medium{{
			Position: 1,
			Format:   new("CD"),
			Tracks: []Track{
				plainTrack("t1", 1, "Track One"),
				plainTrack("t2", 2, "Track Two"),
			},
		}},
	}

	got, err := TranslateAlbum(rel, matcher.DefaultTunables())
	if err != nil {
		t.Fatalf("TranslateAlbum: %v", err)
	}

	want := &matcher.AlbumInfo{
		Artist:        "Plain Artist",
		Album:         "Plain Album",
		VA:            false,
		Media:         "CD",
		Mediums:       1,
		Year:          2001,
		OriginalYear:  1999,
		Country:       "XW",
		Label:         "Plain Label",
		CatalogNum:    "CAT-1",
		AlbumDisambig: "remaster",
		AlbumID:       "rel-plain",
		DataSource:    "MusicBrainz",
		Tracks: []matcher.TrackInfo{
			{Title: "Track One", Artist: "Artist", Index: 1, MediumIndex: 1, Medium: 1, TrackID: "rec-t1", DataSource: "MusicBrainz"},
			{Title: "Track Two", Artist: "Artist", Index: 2, MediumIndex: 2, Medium: 1, TrackID: "rec-t2", DataSource: "MusicBrainz"},
		},
	}
	assertAlbumInfoEqual(t, got, want)
}

// TestTranslateAlbum_MediumsCountsUnfilteredWhileTracksAreFiltered
// proves T3's measured pair inside the full composition: a two-medium
// release reports Mediums=2 even when IgnoredMedia drops one medium's
// tracks entirely.
func TestTranslateAlbum_MediumsCountsUnfilteredWhileTracksAreFiltered(t *testing.T) {
	rel := &Release{
		ID:           "rel-two-media",
		Title:        "Two Media Album",
		ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Artist")},
		Media: []Medium{
			{Position: 1, Format: new("CD"), Tracks: []Track{plainTrack("t1", 1, "A")}},
			{Position: 2, Format: new("Vinyl"), Tracks: []Track{plainTrack("t2", 1, "B")}},
		},
	}

	tun := matcher.DefaultTunables()
	tun.IgnoredMedia = []string{"Vinyl"}

	got, err := TranslateAlbum(rel, tun)
	if err != nil {
		t.Fatalf("TranslateAlbum: %v", err)
	}
	if got.Mediums != 2 {
		t.Errorf("Mediums = %d, want 2 (the unfiltered count)", got.Mediums)
	}
	if len(got.Tracks) != 1 {
		t.Fatalf("Tracks has %d entries, want 1 (the Vinyl medium's track is dropped)", len(got.Tracks))
	}
	if got.Tracks[0].Title != "A" {
		t.Errorf("surviving track title = %q, want %q", got.Tracks[0].Title, "A")
	}
}

// TestTranslateAlbum_VariousArtistsOverwritesArtistFields proves the VA
// override reaches the full composition: AlbumInfo.Artist becomes
// tun.VAName and AlbumInfo.VA is true.
func TestTranslateAlbum_VariousArtistsOverwritesArtistFields(t *testing.T) {
	rel := &Release{
		ID:           "rel-va",
		Title:        "VA Album",
		ArtistCredit: []CreditPart{simpleCreditPart(variousArtistsID, "Ignored Credit Name")},
		Media: []Medium{{
			Position: 1,
			Format:   new("CD"),
			Tracks:   []Track{plainTrack("t1", 1, "A")},
		}},
	}

	got, err := TranslateAlbum(rel, matcher.DefaultTunables())
	if err != nil {
		t.Fatalf("TranslateAlbum: %v", err)
	}
	if !got.VA {
		t.Errorf("VA = false, want true")
	}
	if got.Artist != "Various Artists" {
		t.Errorf("Artist = %q, want %q", got.Artist, "Various Artists")
	}
}

// TestTranslateAlbum_NoEventsNoDateFallsBackToOriginalDate proves the
// original-date fallback: a release with no release-events and no
// top-level date takes its Year from the release-group's original
// date instead. Named mutation 6, moving the release-event assignment
// before the release-group parse, fails this test by leaving Year at
// zero rather than the original year.
func TestTranslateAlbum_NoEventsNoDateFallsBackToOriginalDate(t *testing.T) {
	rel := &Release{
		ID:           "rel-fallback",
		Title:        "Fallback Album",
		ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Artist")},
		ReleaseGroup: ReleaseGroup{FirstReleaseDate: "1990-06-15"},
		Media: []Medium{{
			Position: 1,
			Format:   new("CD"),
			Tracks:   []Track{plainTrack("t1", 1, "A")},
		}},
	}

	got, err := TranslateAlbum(rel, matcher.DefaultTunables())
	if err != nil {
		t.Fatalf("TranslateAlbum: %v", err)
	}
	if got.OriginalYear != 1990 {
		t.Fatalf("OriginalYear = %d, want 1990", got.OriginalYear)
	}
	if got.Year != 1990 {
		t.Errorf("Year = %d, want 1990 (the original-date fallback)", got.Year)
	}
	if got.Country != "" {
		t.Errorf("Country = %q, want \"\" (no top-level country)", got.Country)
	}
}

// TestTranslateAlbum_LabelAndCatalogNumAreGuardedDifferently proves the
// asymmetry fold m20 names: CatalogNum reads unconditionally even when
// the label object is the "[no label]" sentinel, while Label is empty
// exactly then.
func TestTranslateAlbum_LabelAndCatalogNumAreGuardedDifferently(t *testing.T) {
	rel := &Release{
		ID:           "rel-no-label",
		Title:        "No Label Album",
		ArtistCredit: []CreditPart{simpleCreditPart("art-1", "Artist")},
		LabelInfo: []LabelInfo{{
			CatalogNumber: new("CAT-9"),
			Label:         &Label{Name: "[no label]"},
		}},
		Media: []Medium{{
			Position: 1,
			Format:   new("CD"),
			Tracks:   []Track{plainTrack("t1", 1, "A")},
		}},
	}

	got, err := TranslateAlbum(rel, matcher.DefaultTunables())
	if err != nil {
		t.Fatalf("TranslateAlbum: %v", err)
	}
	if got.CatalogNum != "CAT-9" {
		t.Errorf("CatalogNum = %q, want %q (read unconditionally)", got.CatalogNum, "CAT-9")
	}
	if got.Label != "" {
		t.Errorf("Label = %q, want \"\" (the [no label] sentinel)", got.Label)
	}
}

// TestTranslateAlbum_MalformedReleaseWrapsSentinel proves TranslateAlbum
// runs Validate first and returns ErrMalformedRelease rather than
// panicking on a release whose artist-credit is empty.
func TestTranslateAlbum_MalformedReleaseWrapsSentinel(t *testing.T) {
	rel := &Release{ID: "rel-bad", Title: "Bad Album"}
	_, err := TranslateAlbum(rel, matcher.DefaultTunables())
	if err == nil {
		t.Fatal("TranslateAlbum returned no error for a release with empty artist-credit")
	}
	if !errors.Is(err, ErrMalformedRelease) {
		t.Errorf("error %v does not wrap ErrMalformedRelease", err)
	}
}

func assertAlbumInfoEqual(t *testing.T, got, want *matcher.AlbumInfo) {
	t.Helper()
	if got.Artist != want.Artist {
		t.Errorf("Artist = %q, want %q", got.Artist, want.Artist)
	}
	if got.Album != want.Album {
		t.Errorf("Album = %q, want %q", got.Album, want.Album)
	}
	if got.VA != want.VA {
		t.Errorf("VA = %v, want %v", got.VA, want.VA)
	}
	if got.Media != want.Media {
		t.Errorf("Media = %q, want %q", got.Media, want.Media)
	}
	if got.Mediums != want.Mediums {
		t.Errorf("Mediums = %d, want %d", got.Mediums, want.Mediums)
	}
	if got.Year != want.Year {
		t.Errorf("Year = %d, want %d", got.Year, want.Year)
	}
	if got.OriginalYear != want.OriginalYear {
		t.Errorf("OriginalYear = %d, want %d", got.OriginalYear, want.OriginalYear)
	}
	if got.Country != want.Country {
		t.Errorf("Country = %q, want %q", got.Country, want.Country)
	}
	if got.Label != want.Label {
		t.Errorf("Label = %q, want %q", got.Label, want.Label)
	}
	if got.CatalogNum != want.CatalogNum {
		t.Errorf("CatalogNum = %q, want %q", got.CatalogNum, want.CatalogNum)
	}
	if got.AlbumDisambig != want.AlbumDisambig {
		t.Errorf("AlbumDisambig = %q, want %q", got.AlbumDisambig, want.AlbumDisambig)
	}
	if got.AlbumID != want.AlbumID {
		t.Errorf("AlbumID = %q, want %q", got.AlbumID, want.AlbumID)
	}
	if got.DataSource != want.DataSource {
		t.Errorf("DataSource = %q, want %q", got.DataSource, want.DataSource)
	}
	if len(got.Tracks) != len(want.Tracks) {
		t.Fatalf("Tracks has %d entries, want %d", len(got.Tracks), len(want.Tracks))
	}
	for i := range got.Tracks {
		if got.Tracks[i] != want.Tracks[i] {
			t.Errorf("Tracks[%d] = %+v, want %+v", i, got.Tracks[i], want.Tracks[i])
		}
	}
}
