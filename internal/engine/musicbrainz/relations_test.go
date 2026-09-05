package musicbrainz

import (
	"fmt"
	"testing"
)

// TestGroupRelations_TypesAndIntraGroupOrder proves both structural
// rules at once: an interleaved url/artist/url/artist array regroups
// into exactly the keys "url" and "artist", and each group keeps the
// relations' original response order within it (T2's stable-sort
// sub-rule). The fixture carries ten relations per group rather than
// two: mutating slices.SortStableFunc to the unstable slices.SortFunc
// left a two-per-group fixture green (Go's small-slice insertion sort
// happens to preserve order at that size), so the fixture was enlarged
// until the mutation reliably failed the test.
func TestGroupRelations_TypesAndIntraGroupOrder(t *testing.T) {
	const perGroup = 10
	var rels []Relation
	var wantURLOrder, wantArtistOrder []string
	for i := range perGroup {
		urlID := fmt.Sprintf("url-%d", i)
		artistID := fmt.Sprintf("artist-%d", i)
		rels = append(rels,
			Relation{TargetType: "url", Type: urlID, Direction: "forward"},
			Relation{TargetType: "artist", Type: artistID, Direction: "forward"},
		)
		wantURLOrder = append(wantURLOrder, urlID)
		wantArtistOrder = append(wantArtistOrder, artistID)
	}

	groups := GroupRelations(rels)

	if len(groups) != 2 {
		t.Fatalf("GroupRelations returned %d keys, want 2 (url, artist)", len(groups))
	}
	assertRelationOrder(t, "url", groups["url"], wantURLOrder)
	assertRelationOrder(t, "artist", groups["artist"], wantArtistOrder)
}

func assertRelationOrder(t *testing.T, key string, got []Relation, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("group %q has %d entries, want %d", key, len(got), len(want))
	}
	for i, r := range got {
		if r.Type != want[i] {
			t.Errorf("group %q entry %d has Type %q, want %q (order not preserved)", key, i, r.Type, want[i])
		}
	}
}

// TestGroupRelations_EmptyInputYieldsEmptyMap proves GroupRelations does
// not synthesize a group for an empty relations list.
func TestGroupRelations_EmptyInputYieldsEmptyMap(t *testing.T) {
	got := GroupRelations(nil)
	if len(got) != 0 {
		t.Errorf("GroupRelations(nil) = %v, want an empty map", got)
	}
}
