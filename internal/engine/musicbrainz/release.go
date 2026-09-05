package musicbrainz

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

// ErrMalformedRelease is the one sentinel every boundary failure in
// this package wraps, so a caller (3c's client) branches once on "this
// response is unusable" rather than on a hierarchy of error types.
var ErrMalformedRelease = errors.New("malformed musicbrainz release")

// Release is the subset of a WS/2 release beets' own album_info reads.
// Several of its fields feed applied metadata rather than a distance
// (medium title and track count, track id and number, alias sort names,
// barcode, the release-group type fields), so nothing in 3b reads them;
// they are decoded here because the wire shape is decoded once and
// rung 8 should not re-derive it. That is the deliberate counterpart to
// dropping credits.SortName (fold, M-i): the wire vocabulary names the
// payload, and the computed vocabulary produces only what the matcher
// reads. The struct tags name the wire keys directly, which discharges the
// hyphen-to-underscore half of the client's own _normalize_data
// (beetsplug/_utils/musicbrainz.py:774) with no runtime pass: that
// rewrite exists so a Python dict can be read with underscore keys, and
// a Go struct does not need it. The relations regrouping half is
// structural and lives in GroupRelations.
//
// Rejected alternative: decoding to map[string]any, running a faithful
// recursive normalizer over every key, and re-decoding into typed
// fields. That reproduces a normalizer for hundreds of keys nothing
// reads, loses every type along the way, and makes the decoder testable
// only through goldens, which go-conventions calls last-resort.
type Release struct {
	ID             string         `json:"id"`
	Title          string         `json:"title"`
	Status         string         `json:"status"`
	Disambiguation string         `json:"disambiguation"`
	Country        *string        `json:"country"`
	Date           string         `json:"date"`
	Barcode        *string        `json:"barcode"`
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
	Name     string  `json:"name"`
	SortName string  `json:"sort-name"`
	Locale   *string `json:"locale"`
	Primary  *bool   `json:"primary"`
	Type     *string `json:"type"`
}

// CreditPart is one element of an artist-credit array.
type CreditPart struct {
	Name       string `json:"name"`
	JoinPhrase string `json:"joinphrase"`
	Artist     Artist `json:"artist"`
}

// Artist is a credit element's artist entity.
type Artist struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	SortName string  `json:"sort-name"`
	Aliases  []Alias `json:"aliases"`
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

// ReleaseGroup carries the original-release date and the type fields.
type ReleaseGroup struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Disambiguation   string   `json:"disambiguation"`
	FirstReleaseDate string   `json:"first-release-date"`
	PrimaryType      *string  `json:"primary-type"`
	SecondaryTypes   []string `json:"secondary-types"`
	Aliases          []Alias  `json:"aliases"`
}

// Medium is one disc.
type Medium struct {
	Position   int     `json:"position"`
	Title      string  `json:"title"`
	Format     *string `json:"format"`
	TrackCount int     `json:"track-count"`
	Pregap     *Track  `json:"pregap"`
	Tracks     []Track `json:"tracks"`
	DataTracks []Track `json:"data-tracks"`
}

// Track is one track entry on a medium. Position is a wireInt because
// the wire value is an integer in real captures and a string in every
// synthetic case, and beets casts it for the same reason
// (int(track["position"]), beetsplug/musicbrainz.py:592;
// internal/harness/synth/release.go:128 writes the string form).
type Track struct {
	ID           string       `json:"id"`
	Position     wireInt      `json:"position"`
	Number       string       `json:"number"`
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
// JSON number or as a JSON string. Only track positions need it: a
// census over all 399 committed release documents found 6,763 integer
// and 297 string track positions and no other mismatched field, so this
// type is deliberately narrow rather than applied to every number.
type wireInt int

// UnmarshalJSON accepts both spellings and rejects anything else,
// wrapping ErrMalformedRelease on both failure paths: this is the
// decode boundary the sentinel's own comment names as "the one
// sentinel every boundary failure in this package wraps"
// (ErrMalformedRelease's doc comment above), and 3c's client is the
// pass whose Release call makes that boundary real. encoding/json
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
		return fmt.Errorf("musicbrainz: position %s is neither a JSON number nor a JSON string: %w", data, ErrMalformedRelease)
	}
	parsed, err := strconv.Atoi(asString)
	if err != nil {
		return fmt.Errorf("musicbrainz: position %q is not an integer: %w: %w", asString, err, ErrMalformedRelease)
	}
	*n = wireInt(parsed)
	return nil
}

// Relation is one entry of a release's generic relations array. Release
// is nil on any relation whose target is not a release, which is most of
// them: the corpus carries 351 url, 285 artist, and 36 release
// relations.
type Relation struct {
	TargetType string          `json:"target-type"`
	Type       string          `json:"type"`
	Direction  string          `json:"direction"`
	Release    *RelationTarget `json:"release"`
}

// RelationTarget is a relation's target entity, named rather than
// anonymous so `go doc` shows it and so the caller that must test it can
// name the type (fold, m13).
type RelationTarget struct {
	ID string `json:"id"`
}

// Validate reports whether rel can be translated at all, wrapping
// ErrMalformedRelease and naming the offending element. It is the
// boundary check TranslateAlbum runs first, so the rest of this package
// may assume well-formed input: a WS response is external input, and
// go-conventions puts validation at the boundary rather than a nil check
// at every read.
//
// It checks exactly four things: a relation with an empty target-type;
// an empty release-level artist-credit array; an empty resolved track
// credit on any track of any medium (the track's own when non-empty,
// the recording's otherwise); and a backward transl-tracklisting
// relation whose release target is absent. The last two exist because
// the rest of this package panics rather than errors on them:
// parseArtistCredits indexes element zero of the resolved credit list,
// and PseudoReleaseTarget dereferences Relation.Release. Neither has a
// corpus instance, so both fixtures proving them are hand-authored on
// the proximity precedent.
//
// A medium whose Tracks, Pregap, and DataTracks are all absent is
// legal, not an error: beets reads all three through .get, and a medium
// missing all three simply contributes nothing to track extraction
// (T5). This is a named non-check, not an oversight.
func (rel *Release) Validate() error {
	for i, r := range rel.Relations {
		if r.TargetType == "" {
			return fmt.Errorf("relation %d has empty target-type: %w", i, ErrMalformedRelease)
		}
	}

	if len(rel.ArtistCredit) == 0 {
		return fmt.Errorf("release %s has empty artist-credit: %w", rel.ID, ErrMalformedRelease)
	}

	for _, medium := range rel.Media {
		for _, track := range assembledTracks(medium, true) {
			if len(resolvedTrackCredit(track)) == 0 {
				return fmt.Errorf(
					"medium %d track %d has empty artist-credit on both the track and the recording: %w",
					medium.Position, track.Position, ErrMalformedRelease,
				)
			}
		}
	}

	for i, r := range rel.Relations {
		if r.TargetType == "release" && r.Type == "transl-tracklisting" && r.Direction == "backward" && r.Release == nil {
			return fmt.Errorf("relation %d is a backward transl-tracklisting relation with no release target: %w", i, ErrMalformedRelease)
		}
	}

	return nil
}

// assembledTracks returns a medium's pregap, tracks, and data-tracks in
// that fixed order (T5), the same assembly get_tracks_from_medium reads
// through before either filter runs. It is the one copy of that order:
// tracksFromMedium passes withDataTracks as !tun.IgnoreDataTracks, and
// Validate passes true, because it has no Tunables to apply a filter
// with and a track later dropped by a filter is still a track this
// boundary check must be able to read without panicking.
func assembledTracks(medium Medium, withDataTracks bool) []Track {
	var tracks []Track
	if medium.Pregap != nil {
		tracks = append(tracks, *medium.Pregap)
	}
	tracks = append(tracks, medium.Tracks...)
	if withDataTracks {
		tracks = append(tracks, medium.DataTracks...)
	}
	return tracks
}

// resolvedTrackCredit is the track-level override T6 describes: the
// track's own artist-credit when non-empty, the recording's otherwise
// (recording["artist_credit"] = track["artist_credit"] or
// recording["artist_credit"], beetsplug/musicbrainz.py:586-588).
func resolvedTrackCredit(track Track) []CreditPart {
	if len(track.ArtistCredit) > 0 {
		return track.ArtistCredit
	}
	return track.Recording.ArtistCredit
}
