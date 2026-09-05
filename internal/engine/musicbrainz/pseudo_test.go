package musicbrainz

import (
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// TestPseudoReleaseTarget_ExactStatusAndRelationPredicate proves the
// exact status and relation predicate T12 and the interface comment
// describe: both must hold before PseudoReleaseTarget reports ok. Two
// negative fixtures each isolate one half of the predicate, and one
// more shows the two negatives compose (status right, no relation) as
// well as the reverse (relation right, status wrong).
//
// Mutation 1 (accepting direction == "forward") fails the
// forward-relation row, because the fixture's only relation runs
// forward. Mutation 2 (comparing Status case-insensitively) fails the
// status-case-mismatch row, because the fixture's status is
// "pseudo-release" lowercase.
func TestPseudoReleaseTarget_ExactStatusAndRelationPredicate(t *testing.T) {
	tests := []struct {
		name    string
		fixture string
		wantID  string
		wantOK  bool
	}{
		{"status and relation both match", "pseudo_target_valid.json", "actual-release-1", true},
		{"forward direction is not backward", "pseudo_target_forward_relation.json", "", false},
		{"status compared case-sensitively", "pseudo_target_status_case_mismatch.json", "", false},
		{"status matches but no relation exists", "pseudo_target_no_relation.json", "", false},
		{"relation matches but status is not pseudo", "pseudo_target_relation_wrong_status.json", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rel := loadTestdata[Release](t, tt.fixture)
			gotID, gotOK := PseudoReleaseTarget(&rel)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("PseudoReleaseTarget() = (%q, %v), want (%q, %v)", gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

// allSevenDifferPseudo and allSevenDifferActual are hand-authored
// AlbumInfo values, not corpus-derived: T12 records that Year,
// OriginalYear, and Mediums were already equal on all four committed
// pseudo cases, so no committed case can catch a transcription slip in
// the merged-field list. Every field differs between the two sides so a
// dropped or wrongly-added field shows up unambiguously.
func allSevenDifferPseudo() matcher.AlbumInfo {
	return matcher.AlbumInfo{
		Artist:        "Pseudo Artist",
		Album:         "Pseudo Album",
		VA:            false,
		Media:         "Vinyl",
		Mediums:       1,
		Year:          1999,
		OriginalYear:  1998,
		Country:       "JP",
		Label:         "Pseudo Label",
		CatalogNum:    "PSEUDO-001",
		AlbumDisambig: "pseudo disambig",
		AlbumID:       "pseudo-release-id",
		DataSource:    "MusicBrainz",
		Tracks:        []matcher.TrackInfo{{Title: "Pseudo Track"}},
	}
}

func allSevenDifferActual() matcher.AlbumInfo {
	return matcher.AlbumInfo{
		Artist:        "Actual Artist",
		Album:         "Actual Album",
		VA:            true,
		Media:         "CD",
		Mediums:       2,
		Year:          2001,
		OriginalYear:  2000,
		Country:       "GB",
		Label:         "Actual Label",
		CatalogNum:    "ACTUAL-002",
		AlbumDisambig: "actual disambig",
		AlbumID:       "actual-release-id",
		DataSource:    "Discogs",
		Tracks:        []matcher.TrackInfo{{Title: "Actual Track"}, {Title: "Actual Track 2"}},
	}
}

// TestMergePseudoRelease_MovesExactlySevenFields proves the merge moves
// exactly the seven verdict-moving AlbumInfo fields the doc comment
// names and leaves the other seven, plus Tracks, at the pseudo's own
// value.
//
// Mutation 3 (dropping any one of the seven merged fields, for example
// Label) fails this test on that field, because allSevenDifferPseudo
// and allSevenDifferActual disagree on every merged field. Mutation 4
// (adding AlbumID to the merged set) fails this test's AlbumID
// assertion, proving the negative half: AlbumID must stay the pseudo's.
func TestMergePseudoRelease_MovesExactlySevenFields(t *testing.T) {
	pseudo := allSevenDifferPseudo()
	actual := allSevenDifferActual()

	got := MergePseudoRelease(&pseudo, &actual)

	moved := []struct {
		name      string
		got, want any
	}{
		{"Media", got.Media, actual.Media},
		{"Mediums", got.Mediums, actual.Mediums},
		{"Country", got.Country, actual.Country},
		{"CatalogNum", got.CatalogNum, actual.CatalogNum},
		{"Year", got.Year, actual.Year},
		{"OriginalYear", got.OriginalYear, actual.OriginalYear},
		{"Label", got.Label, actual.Label},
	}
	for _, m := range moved {
		if m.got != m.want {
			t.Errorf("merged.%s = %v, want the actual's %v", m.name, m.got, m.want)
		}
	}

	kept := []struct {
		name      string
		got, want any
	}{
		{"Artist", got.Artist, pseudo.Artist},
		{"Album", got.Album, pseudo.Album},
		{"VA", got.VA, pseudo.VA},
		{"AlbumDisambig", got.AlbumDisambig, pseudo.AlbumDisambig},
		{"AlbumID", got.AlbumID, pseudo.AlbumID},
		{"DataSource", got.DataSource, pseudo.DataSource},
	}
	for _, k := range kept {
		if k.got != k.want {
			t.Errorf("merged.%s = %v, want the pseudo's %v", k.name, k.got, k.want)
		}
	}
}

// TestMergePseudoRelease_TracksSurviveFromPseudo proves Tracks stays
// the pseudo's own list even when the actual carries a different track
// count, another field T12 records as having no corpus signal on any
// of the four committed pseudo cases.
//
// Mutation 5 (taking Tracks from the actual) fails this test, because
// the fixture's pseudo and actual carry different track counts and
// different titles.
func TestMergePseudoRelease_TracksSurviveFromPseudo(t *testing.T) {
	pseudo := allSevenDifferPseudo()
	actual := allSevenDifferActual()

	got := MergePseudoRelease(&pseudo, &actual)

	if len(got.Tracks) != len(pseudo.Tracks) {
		t.Fatalf("merged.Tracks has %d tracks, want the pseudo's %d", len(got.Tracks), len(pseudo.Tracks))
	}
	for i, tr := range got.Tracks {
		if tr.Title != pseudo.Tracks[i].Title {
			t.Errorf("merged.Tracks[%d].Title = %q, want the pseudo's %q", i, tr.Title, pseudo.Tracks[i].Title)
		}
	}
}

// TestMergePseudoRelease_DoesNotMutateItsArguments proves
// MergePseudoRelease returns a new value rather than mutating pseudo or
// actual, matching the doc comment's claim.
func TestMergePseudoRelease_DoesNotMutateItsArguments(t *testing.T) {
	pseudo := allSevenDifferPseudo()
	actual := allSevenDifferActual()
	pseudoBefore := pseudo
	actualBefore := actual

	_ = MergePseudoRelease(&pseudo, &actual)

	if pseudo.Media != pseudoBefore.Media || pseudo.Country != pseudoBefore.Country {
		t.Error("MergePseudoRelease mutated pseudo, want it left untouched")
	}
	if actual.Album != actualBefore.Album || actual.Artist != actualBefore.Artist {
		t.Error("MergePseudoRelease mutated actual, want it left untouched")
	}
}
