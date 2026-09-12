package shadow

import (
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// TestItemBagRoundTripsThroughProjectItemFields grades itemBag's
// nineteen beets field names against matcher.ProjectItemFields' own
// reads (matcher/info.go:189-211). Every one of the nineteen fields
// carries a distinguishable non-zero value, so a name this table
// spells wrong shows up as a dropped field rather than as a
// coincidence.
func TestItemBagRoundTripsThroughProjectItemFields(t *testing.T) {
	want := matcher.ItemFields{
		Artist:        "Cabinet",
		Album:         "Windrose",
		AlbumArtist:   "Cabinet Ensemble",
		Year:          1994,
		DiscTotal:     2,
		MBAlbumID:     "22222222-2222-2222-2222-222222222222",
		Label:         "Nightwood",
		Barcode:       "5901234123457",
		CatalogNum:    "NW-004",
		Country:       "GB",
		Media:         "CD",
		AlbumDisambig: "deluxe",
		DataSource:    "MusicBrainz",
		Length:        612.5,
		Title:         "Sub Rosa",
		Track:         3,
		MBTrackID:     "11111111-1111-1111-1111-111111111111",
		Disc:          2,
		Comp:          true,
	}
	if got := matcher.ProjectItemFields(itemBag(want)); got != want {
		t.Errorf("round trip =\n%+v\nwant\n%+v", got, want)
	}
}
