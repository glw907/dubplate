package musicbrainz

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func loadTestdataRelease(t *testing.T, name string) Release {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	var rel Release
	if err := json.Unmarshal(data, &rel); err != nil {
		t.Fatalf("decoding testdata/%s: %v", name, err)
	}
	return rel
}

// TestRelease_DecodesTrackPositionAsWireInt proves the census claim
// against the real corpus rather than asserting it: media-baseline's
// committed release-lookup response carries every track position as a
// JSON string, and a plain int Position field fails the whole decode
// (fold, B3). Mutating Track.Position from wireInt to int reproduces
// exactly that failure.
func TestRelease_DecodesTrackPositionAsWireInt(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "corpus", "cases", "media-baseline", "responses", "001.json"))
	if err != nil {
		t.Fatalf("reading corpus fixture: %v", err)
	}

	var rel Release
	if err := json.Unmarshal(data, &rel); err != nil {
		t.Fatalf("decoding media-baseline's committed response: %v", err)
	}
	if len(rel.Media) == 0 {
		t.Fatal("media-baseline's response decoded with zero media, want at least one")
	}

	var sawTrack bool
	for _, medium := range rel.Media {
		for _, track := range medium.Tracks {
			sawTrack = true
			if track.Position <= 0 {
				t.Errorf("track %s decoded position %d, want a positive position", track.ID, track.Position)
			}
		}
	}
	if !sawTrack {
		t.Fatal("media-baseline's response decoded with zero tracks, want at least one")
	}
}

func TestRelease_ValidateAcceptsWellFormedRelease(t *testing.T) {
	rel := loadTestdataRelease(t, "well_formed.json")
	if err := rel.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil", err)
	}
}

// TestRelease_ValidateRejectsEmptyTargetType proves Validate's first
// rule fires: a relation with no target-type panics beets'
// _normalize_data (AttributeError, T2) and must instead be a boundary
// error here.
func TestRelease_ValidateRejectsEmptyTargetType(t *testing.T) {
	rel := loadTestdataRelease(t, "empty_target_type.json")
	err := rel.Validate()
	if !errors.Is(err, ErrMalformedRelease) {
		t.Fatalf("Validate() = %v, want an error wrapping ErrMalformedRelease", err)
	}
}

func TestRelease_ValidateRejectsEmptyReleaseCredit(t *testing.T) {
	rel := loadTestdataRelease(t, "empty_release_credit.json")
	err := rel.Validate()
	if !errors.Is(err, ErrMalformedRelease) {
		t.Fatalf("Validate() = %v, want an error wrapping ErrMalformedRelease", err)
	}
}

// TestRelease_ValidateRejectsEmptyTrackCredit is hand-authored on the
// proximity precedent: no committed case carries a track whose own
// credit and whose recording's credit are both empty, but
// parseArtistCredits (Task 4) indexes element zero of the resolved
// credit list and would panic on one (fold, M-d).
func TestRelease_ValidateRejectsEmptyTrackCredit(t *testing.T) {
	rel := loadTestdataRelease(t, "empty_track_credit.json")
	err := rel.Validate()
	if !errors.Is(err, ErrMalformedRelease) {
		t.Fatalf("Validate() = %v, want an error wrapping ErrMalformedRelease", err)
	}
}

// TestRelease_ValidateRejectsPseudoTargetMissing is hand-authored on the
// proximity precedent: no committed backward transl-tracklisting
// relation lacks its release object, but PseudoReleaseTarget
// dereferences Relation.Release and would panic on one (fold, M-e).
func TestRelease_ValidateRejectsPseudoTargetMissing(t *testing.T) {
	rel := loadTestdataRelease(t, "pseudo_target_missing.json")
	err := rel.Validate()
	if !errors.Is(err, ErrMalformedRelease) {
		t.Fatalf("Validate() = %v, want an error wrapping ErrMalformedRelease", err)
	}
}

// TestRelease_ValidateAllowsMediumWithNoTracks names the one deliberate
// non-check: a medium whose tracks, pregap, and data-tracks are all
// absent is legal, because beets reads all three through .get and such
// a medium simply contributes nothing to track extraction (T5).
func TestRelease_ValidateAllowsMediumWithNoTracks(t *testing.T) {
	rel := loadTestdataRelease(t, "medium_no_tracks.json")
	if err := rel.Validate(); err != nil {
		t.Errorf("Validate() = %v, want nil (a trackless medium is legal)", err)
	}
}
