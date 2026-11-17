package musicbrainz

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// errMalformedRelease reports a release body that does not decode or
// that the rest of this package cannot read without panicking.
var errMalformedRelease = errors.New("malformed musicbrainz release")

// Release is the subset of a WS/2 release the matcher's vocabulary is
// translated from. The struct tags name the wire keys directly, which
// does the key-renaming half of beets' _normalize_data
// (beetsplug/_utils/musicbrainz.py:774) with no runtime pass. The
// relations regrouping half is not needed either: PseudoReleaseTarget
// filters Relations by target type.
type Release struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Status         string         `json:"status"`
	Disambiguation string         `json:"disambiguation"`
	Country        *string        `json:"country"`
	Date           string         `json:"date"`
	Aliases        []Alias        `json:"aliases"`
	ArtistCredit   []CreditPart   `json:"artist-credit"`
	ReleaseEvents  []ReleaseEvent `json:"release-events"`
	LabelInfo      []LabelInfo    `json:"label-info"`
	Media          []Medium       `json:"media"`
	ReleaseGroup   ReleaseGroup   `json:"release-group"`
	Relations      []Relation     `json:"relations"`
}

// Alias is one MusicBrainz alias. Primary is a pointer because the wire
// value is bool | null and null is falsy in beets' own read
// (_utils/musicbrainz.py:138-155).
type Alias struct {
	Name    string  `json:"name"`
	Locale  *string `json:"locale"`
	Primary *bool   `json:"primary"`
	Type    *string `json:"type"`
}

// CreditPart is one element of an artist-credit array.
type CreditPart struct {
	JoinPhrase string `json:"joinphrase"`
	Artist     Artist `json:"artist"`
}

// Artist is a credit element's artist entity.
type Artist struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	Aliases []Alias `json:"aliases"`
}

// ReleaseEvent is one dated release event. Area is a pointer because
// the wire value is nullable and beets' walrus at
// beetsplug/musicbrainz.py:174 skips a null one.
type ReleaseEvent struct {
	Date string `json:"date"`
	Area *Area  `json:"area"`
}

// Area carries the ISO codes release-event selection tests membership
// against.
type Area struct {
	ISO31661Codes []string `json:"iso-3166-1-codes"`
}

// LabelInfo is one label/catalog-number pair.
type LabelInfo struct {
	CatalogNumber *string `json:"catalog-number"`
	Label         *Label  `json:"label"`
}

// Label is a label entity.
type Label struct {
	Name string `json:"name"`
}

// ReleaseGroup carries the original-release date.
type ReleaseGroup struct {
	FirstReleaseDate string `json:"first-release-date"`
}

// Medium is one disc.
type Medium struct {
	Position   int     `json:"position"`
	Format     *string `json:"format"`
	Pregap     *Track  `json:"pregap"`
	Tracks     []Track `json:"tracks"`
	DataTracks []Track `json:"data-tracks"`
}

// Track is one track entry on a medium. Position is a wireInt because
// the wire value is an integer in real captures and a string in every
// synthetic case, and beets casts it for the same reason
// (int(track["position"]), beetsplug/musicbrainz.py:592).
type Track struct {
	Position     wireInt      `json:"position"`
	Title        string       `json:"title"`
	Length       *float64     `json:"length"`
	ArtistCredit []CreditPart `json:"artist-credit"`
	Recording    Recording    `json:"recording"`
}

// Recording is a track's underlying recording.
type Recording struct {
	ID           string       `json:"id"`
	Title        string       `json:"title"`
	Length       *float64     `json:"length"`
	Video        bool         `json:"video"`
	Aliases      []Alias      `json:"aliases"`
	ArtistCredit []CreditPart `json:"artist-credit"`
}

// wireInt is an integer the MusicBrainz WS/2 payload spells either as a
// JSON number or as a JSON string. Only track positions are spelled
// both ways, so the type is applied to them alone.
type wireInt int

// UnmarshalJSON accepts both spellings and rejects anything else,
// wrapping errMalformedRelease on both failure paths. encoding/json
// returns a custom unmarshaler's error unchanged for a
// non-UnmarshalTypeError, so errors.Is survives json.Unmarshal.
func (n *wireInt) UnmarshalJSON(data []byte) error {
	var asInt int
	if err := json.Unmarshal(data, &asInt); err == nil {
		*n = wireInt(asInt)
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err != nil {
		return fmt.Errorf("musicbrainz: position %s is neither a JSON number nor a JSON string: %w", data, errMalformedRelease)
	}
	parsed, err := strconv.Atoi(asString)
	if err != nil {
		return fmt.Errorf("musicbrainz: position %q is not an integer: %w: %w", asString, err, errMalformedRelease)
	}
	*n = wireInt(parsed)
	return nil
}

// Relation is one entry of a release's generic relations array. Release
// is nil on any relation whose target is not a release.
type Relation struct {
	TargetType string          `json:"target-type"`
	Type       string          `json:"type"`
	Direction  string          `json:"direction"`
	Release    *RelationTarget `json:"release"`
}

// RelationTarget is a relation's target entity, named rather than
// anonymous so `go doc` shows it and so the caller that must test it can
// name the type.
type RelationTarget struct {
	ID string `json:"id"`
}

// validate reports whether rel can be translated at all, wrapping
// errMalformedRelease and naming the offending element. It is the
// boundary check TranslateAlbum runs first, so the rest of this package
// may assume well-formed input.
//
// It checks exactly four things: a relation with an empty target-type;
// an empty release-level artist-credit array; an empty resolved track
// credit on any track of any medium; and a backward transl-tracklisting
// relation whose release target is absent. The last two exist because
// the rest of this package panics rather than errors on them:
// parseArtistCredits indexes element zero of the resolved credit list,
// and PseudoReleaseTarget dereferences Relation.Release.
//
// A medium whose Tracks, Pregap, and DataTracks are all absent is
// legal: beets reads all three through .get, and such a medium
// contributes nothing to track extraction.
func (rel *Release) validate() error {
	for i, r := range rel.Relations {
		if r.TargetType == "" {
			return fmt.Errorf("relation %d has empty target-type: %w", i, errMalformedRelease)
		}
	}

	if len(rel.ArtistCredit) == 0 {
		return fmt.Errorf("release %s has empty artist-credit: %w", rel.ID, errMalformedRelease)
	}

	for _, medium := range rel.Media {
		for _, track := range assembledTracks(medium, true) {
			if len(resolvedTrackCredit(track)) == 0 {
				return fmt.Errorf(
					"medium %d track %d has empty artist-credit on both the track and the recording: %w",
					medium.Position, track.Position, errMalformedRelease,
				)
			}
		}
	}

	for i, r := range rel.Relations {
		if isTranslSource(r) && r.Release == nil {
			return fmt.Errorf("relation %d is a backward transl-tracklisting relation with no release target: %w", i, errMalformedRelease)
		}
	}

	return nil
}
