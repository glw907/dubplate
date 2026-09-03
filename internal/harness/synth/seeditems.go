package synth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// ProjectItems reads doc, a WS/2 release document's raw bytes (a
// captured case's responses/001.json, verbatim), and derives the
// baseline item set a real rip of that release would carry: one item
// per track, release order, mirroring [buildItems]'s own default shape
// but sourced from a parsed captured document instead of a
// [ReleaseSpec]. overrides then perturbs that baseline through the
// exact same [applyOverrides] vocabulary (field, missing, extra) a
// synthetic spec uses, so the mistagged and partial-rip axes reach
// their divergence the same way every other case does.
//
// This is the sanctioned reader of a captured response
// (corpus/README.md, "The seed-items tool"): a captured document
// carries per-track fields (title, length, recording id) no MBID or
// axis label can invent, and this is the only code path in the corpus
// tooling that opens one.
func ProjectItems(doc []byte, overrides []Override, allowAlbumIDOverride bool) ([]cassette.ItemFields, error) {
	var rel capturedRelease
	if err := json.Unmarshal(doc, &rel); err != nil {
		return nil, fmt.Errorf("synth: seed-items: parse release document: %w", err)
	}
	if rel.ID == "" {
		return nil, errors.New("synth: seed-items: release document: id: missing")
	}
	if len(rel.Media) == 0 {
		return nil, errors.New("synth: seed-items: release document: media: empty")
	}

	albumArtist := creditString(rel.ArtistCredit)
	year := releaseYear(rel.Date)
	multiDisc := len(rel.Media) > 1
	comp := releaseIsVariousArtists(rel.ArtistCredit)

	var items []cassette.ItemFields
	for mi, m := range rel.Media {
		for _, tr := range m.Tracks {
			title := tr.Title
			if title == "" {
				title = tr.Recording.Title
			}
			artist := trackArtist(tr, albumArtist)
			item := cassette.ItemFields{
				"album":       rel.Title,
				"albumartist": albumArtist,
				"artist":      artist,
				"length":      float64(tr.Length) / 1000,
				"mb_albumid":  rel.ID,
				"mb_trackid":  tr.Recording.ID,
				"title":       title,
				"track":       int(tr.Position),
				"year":        year,
			}
			if comp {
				item["comp"] = true
			}
			if multiDisc {
				item["disc"] = mi + 1
			}
			items = append(items, item)
		}
	}

	return applyOverrides(items, overrides, rel.ID, allowAlbumIDOverride)
}

// flexPosition decodes a WS/2 track or medium "position" field: the
// live MusicBrainz API renders it as a JSON number, but
// internal/harness/synth's own built documents (marshalRelease, via
// wsTrack.Position) render it as a JSON string, and both shapes appear
// in the committed corpus. The decoder accepts either rather than
// picking a side.
type flexPosition int

// UnmarshalJSON implements [json.Unmarshaler].
func (p *flexPosition) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*p = flexPosition(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("position: neither a number nor a string: %w", err)
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("position: %q is not an integer: %w", s, err)
	}
	*p = flexPosition(parsed)
	return nil
}

// capturedRelease is the WS/2 release document decode shape
// [ProjectItems] reads: only the fields a baseline item set needs, so
// unknown keys (aliases, genres, cover-art-archive, and every other
// field a maximal `inc=` lookup returns) are simply ignored by
// encoding/json rather than modeled here.
type capturedRelease struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	Date         string           `json:"date"`
	ArtistCredit []capturedCredit `json:"artist-credit"`
	Media        []capturedMedium `json:"media"`
}

// capturedCredit is one artist-credit entry: a display name, and the
// joinphrase WS/2 inserts before the next entry ("Marvin Gaye" + " & "
// + "Tammi Terrell"), which [creditString] reassembles into the same
// single display string beets itself builds from a multi-entry credit.
type capturedCredit struct {
	Name       string         `json:"name"`
	JoinPhrase string         `json:"joinphrase"`
	Artist     capturedArtist `json:"artist"`
}

// capturedArtist is one artist-credit entry's artist object.
type capturedArtist struct {
	ID string `json:"id"`
}

// capturedMedium is one disc's track list.
type capturedMedium struct {
	Tracks []capturedTrack `json:"tracks"`
}

// capturedTrack is one WS/2 track. Title is the track-level title the
// live API returns; a document built by [marshalRelease] always leaves
// it empty, so [ProjectItems] falls back to Recording.Title, matching
// beets' own get_tracks_from_medium behavior.
type capturedTrack struct {
	Position     flexPosition      `json:"position"`
	Title        string            `json:"title"`
	Length       int               `json:"length"` // milliseconds
	ArtistCredit []capturedCredit  `json:"artist-credit"`
	Recording    capturedRecording `json:"recording"`
}

// capturedRecording is a track's nested recording.
type capturedRecording struct {
	ID           string           `json:"id"`
	Title        string           `json:"title"`
	ArtistCredit []capturedCredit `json:"artist-credit"`
}

// creditString reassembles credits into the one display string beets
// builds from a WS/2 artist-credit array: each entry's name followed
// by its own joinphrase, concatenated in order. "" for an empty credit
// list.
func creditString(credits []capturedCredit) string {
	var b strings.Builder
	for _, c := range credits {
		b.WriteString(c.Name)
		b.WriteString(c.JoinPhrase)
	}
	return b.String()
}

// releaseIsVariousArtists reports whether credits' lead artist carries
// MusicBrainz's well-known Various Artists id, the same signal
// [singleArtistCredit] writes for a synthetic VA release and the sole
// one beets' musicbrainz plugin reads to set AlbumInfo.va.
func releaseIsVariousArtists(credits []capturedCredit) bool {
	return len(credits) > 0 && credits[0].Artist.ID == variousArtistsMBID
}

// trackArtist returns tr's own performer credit string: its
// track-level credit when the live API populates one, else its
// recording's credit (the shape a document built by [marshalRelease]
// always uses), else fallback (the release's own albumartist).
func trackArtist(tr capturedTrack, fallback string) string {
	if name := creditString(tr.ArtistCredit); name != "" {
		return name
	}
	if name := creditString(tr.Recording.ArtistCredit); name != "" {
		return name
	}
	return fallback
}
