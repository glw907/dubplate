// Package synth turns a compact release spec into a synthetic
// conformance-corpus case: a single ID-lookup request against an
// invented MusicBrainz release, and the pinned item set beets matches
// against it. It is the tool the case-authoring tasks (rung 2, pass
// 2b, Tasks 4 through 13) call to generate every structural-stratum
// case, per corpus/README.md's "The synth tool" section.
//
// A ReleaseSpec names one release: album, artist, date, country, and a
// media list of per-medium track lists. [Build] derives the case's
// item set from that release by default (one item per track, in
// release order) and then applies the spec's Overrides, so a case
// expressing "the release's own tracks, with one deliberate
// divergence" needs only the divergence spelled out, not the whole
// item set restated. A family that needs a wholly independent item set
// can supply Items directly instead.
//
// Every value Build invents (release, artist, recording, and track
// ids) is deterministic: the same spec produces byte-identical output
// on every run, which is what makes a re-run of `harness synth`
// against an already-authored spec safe.
package synth

import (
	"errors"
	"fmt"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// Override operations. See [Override].
const (
	OpField   = "field"   // merge Fields into the item at Index
	OpMissing = "missing" // drop the item at Index entirely
	OpExtra   = "extra"   // append a brand-new item carrying Fields
)

// mbAlbumIDField is the item field the fencing rule protects: outside
// family 10, an override may not perturb it, because beets picks the
// looked-up release id from current_metadata's most-common value
// across items, and perturbing it on enough items changes which
// release gets fetched, which is unanswerable under the no-failing-
// request constraint (corpus/README.md, "The synth tool").
const mbAlbumIDField = "mb_albumid"

// ReleaseSpec is the compact input `harness synth --spec` reads: one
// invented release, its track lists, and the item-set perturbations a
// case wants on top of it.
type ReleaseSpec struct {
	// Album is the release title, both the case's item-level "album"
	// tag and the WS/2 document's "title".
	Album string `json:"album"`
	// Artist is the release-level artist display name. Ignored for
	// display purposes when VariousArtists is set (beets overwrites it
	// from its own va_name config once it detects the well-known
	// Various Artists MBID), but still required: it names the artist
	// credit's display string in the invented document.
	Artist string `json:"artist"`
	// VariousArtists marks the release's artist-credit with
	// MusicBrainz's own well-known Various Artists id, the sole signal
	// beets' musicbrainz plugin uses to set AlbumInfo.va.
	VariousArtists bool `json:"various_artists,omitempty"`
	// Date is the release date, "YYYY-MM-DD"; it also seeds every
	// derived item's "year" tag and, absent an OriginalDate override,
	// the release-group's first-release-date.
	Date string `json:"date"`
	// OriginalDate overrides the release-group's own first-release-date,
	// the value a profile with original_date: yes (estate.yaml) makes
	// pinned beets read as AlbumInfo.original_year. Absent (nil) keeps
	// today's behavior: the release-group's first-release-date equals
	// Date, so there is no divergence between a release's own date and
	// its group's original date. A non-empty value names an earlier (or
	// later) original date directly, the shape a reissue case needs to
	// drive a year-penalty branch (corpus/README.md's "The synth tool").
	// The empty string is a distinct, explicit value: it emits an empty
	// release-group first-release-date ("first-release-date": ""), the
	// "no original date on record" shape beets reads as absent. A
	// pointer, rather than a bare string, is what lets that explicit
	// empty string be told apart from the field being left out of the
	// spec altogether, which a plain string cannot express.
	OriginalDate *string `json:"original_date,omitempty"`
	// Country is the release's WS/2 country code (for example "US").
	Country string `json:"country"`
	// Label and CatalogNumber populate a single label-info entry when
	// either is set; leaving both empty emits an empty label-info list.
	Label         string `json:"label,omitempty"`
	CatalogNumber string `json:"catalog_number,omitempty"`
	// Disambiguation is the release's own disambiguation comment.
	Disambiguation string `json:"disambiguation,omitempty"`
	// MBAlbumID is the invented release id: every derived item's
	// mb_albumid, the WS/2 document's own "id", and the request log's
	// release lookup path.
	MBAlbumID string `json:"mb_albumid"`
	// Profile names the corpus/profiles config the case derives under
	// ("estate" or "stock").
	Profile string `json:"profile"`
	// EvalDate is the case's pinned evaluation date.
	EvalDate string `json:"eval_date"`
	// Media lists the release's discs, in order.
	Media []Medium `json:"media"`
	// Items, when non-empty, replaces the release-derived item set
	// wholesale: Overrides is not applied in this mode, and the caller
	// is responsible for every field, including mb_albumid.
	Items []cassette.ItemFields `json:"items,omitempty"`
	// Overrides perturbs the release-derived item set. Ignored when
	// Items is set.
	Overrides []Override `json:"overrides,omitempty"`
	// AllowAlbumIDOverride lifts the mb_albumid fencing rule. Family
	// 10 is the only shape that needs this: every other family's cases
	// terminate under the no-failing-request constraint only because
	// every item resolves to the same looked-up release.
	AllowAlbumIDOverride bool `json:"allow_album_id_override,omitempty"`
}

// Medium is one disc's track list.
type Medium struct {
	// Format is the WS/2 medium format ("CD", "Vinyl", ...).
	Format string  `json:"format"`
	Tracks []Track `json:"tracks"`
}

// Track is one release track, addressed by its position within its
// medium (1-based, matching beets' own per-disc track numbering).
type Track struct {
	Title    string  `json:"title"`
	Length   float64 `json:"length"` // seconds
	Position int     `json:"position"`
	// ArtistCredit overrides the track's own artist credit, for a
	// various-artists release where each track carries a different
	// performer. Empty means the release-level Artist applies.
	ArtistCredit string `json:"artist_credit,omitempty"`
}

// Override perturbs one derived item, or appends a new one that
// matches no candidate track.
type Override struct {
	// Op is one of OpField, OpMissing, or OpExtra.
	Op string `json:"op"`
	// Index selects the derived item OpField and OpMissing act on
	// (0-based, in release track order). Unused for OpExtra.
	Index int `json:"index,omitempty"`
	// Fields is merged onto the item at Index (OpField), or becomes
	// the whole of a brand-new item (OpExtra). mb_albumid is rejected
	// here unless the spec sets AllowAlbumIDOverride; OpExtra always
	// carries the release's own mb_albumid regardless.
	Fields cassette.ItemFields `json:"fields,omitempty"`
}

// validate checks the spec carries every field [Build] needs before
// it starts inventing ids and rendering a document.
func (s *ReleaseSpec) validate() error {
	switch {
	case s.Album == "":
		return errors.New("synth: spec: album: missing")
	case s.Artist == "":
		return errors.New("synth: spec: artist: missing")
	case s.Date == "":
		return errors.New("synth: spec: date: missing")
	case s.Country == "":
		return errors.New("synth: spec: country: missing")
	case s.MBAlbumID == "":
		return errors.New("synth: spec: mb_albumid: missing")
	case s.Profile == "":
		return errors.New("synth: spec: profile: missing")
	case s.EvalDate == "":
		return errors.New("synth: spec: eval_date: missing")
	}
	if len(s.Items) == 0 && len(s.Media) == 0 {
		return errors.New("synth: spec: media: empty, and items was not supplied as an independent item set")
	}
	for i, m := range s.Media {
		if m.Format == "" {
			return fmt.Errorf("synth: spec: media[%d]: format: missing", i)
		}
		if len(m.Tracks) == 0 {
			return fmt.Errorf("synth: spec: media[%d]: tracks: empty", i)
		}
	}
	return nil
}
