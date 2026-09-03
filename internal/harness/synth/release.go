package synth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

// variousArtistsMBID is MusicBrainz's own well-known Various Artists
// artist id: not an invented value, and not a real person or estate
// identifier (CLAUDE.md's universality rule), but a fixed protocol
// constant beets' musicbrainz plugin checks release["artist_credit"]'s
// artist id against to set AlbumInfo.va.
const variousArtistsMBID = "89ad4ac3-39f7-470e-963a-56509c546377"

// pseudoUUID derives a deterministic, invented id in UUID's 8-4-4-4-12
// hex grouping from seed, so the same spec always renders the same
// document: beets treats every id it reads as an opaque string, so
// nothing requires RFC 4122's version bits, only stability across runs
// of the same spec.
func pseudoUUID(seed string) string {
	sum := sha256.Sum256([]byte(seed))
	h := hex.EncodeToString(sum[:16])
	return fmt.Sprintf("%s-%s-%s-%s-%s", h[0:8], h[8:12], h[12:16], h[16:20], h[20:32])
}

// wsArtist is one WS/2 artist object, nested inside an artist-credit
// entry.
type wsArtist struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	SortName       string   `json:"sort-name"`
	Disambiguation string   `json:"disambiguation"`
	Aliases        []string `json:"aliases"`
}

// wsArtistCredit is one entry in a WS/2 artist-credit array: a display
// name, a joinphrase to the next entry, and the artist it names.
type wsArtistCredit struct {
	Name       string   `json:"name"`
	JoinPhrase string   `json:"joinphrase"`
	Artist     wsArtist `json:"artist"`
}

// singleArtistCredit builds a one-entry artist-credit array for name,
// with its artist id resolved deterministically by [pseudoUUID], or
// MusicBrainz's own Various Artists id when va is set.
func singleArtistCredit(name string, va bool) []wsArtistCredit {
	id := variousArtistsMBID
	if !va {
		id = pseudoUUID("artist:" + name)
	}
	return []wsArtistCredit{{
		Name:       name,
		JoinPhrase: "",
		Artist: wsArtist{
			ID:       id,
			Name:     name,
			SortName: name,
			Aliases:  []string{},
		},
	}}
}

// wsTextRepresentation is a release's script/language pair.
type wsTextRepresentation struct {
	Script   string `json:"script"`
	Language string `json:"language"`
}

// wsReleaseGroup is a release's parent release-group.
type wsReleaseGroup struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	PrimaryType      string   `json:"primary-type"`
	SecondaryTypes   []string `json:"secondary-types"`
	FirstReleaseDate string   `json:"first-release-date"`
	Disambiguation   string   `json:"disambiguation"`
	Aliases          []string `json:"aliases"`
}

// wsLabel is the label named by a wsLabelInfo entry.
type wsLabel struct {
	Name string `json:"name"`
}

// wsLabelInfo is one release label-info entry: a catalog number and
// the label that issued it.
type wsLabelInfo struct {
	CatalogNumber *string  `json:"catalog-number"`
	Label         *wsLabel `json:"label"`
}

// wsArea is a release-event's area (country) reference.
type wsArea struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	ISO31661Codes []string `json:"iso-3166-1-codes"`
}

// wsReleaseEvent is one release-events entry.
type wsReleaseEvent struct {
	Date string `json:"date"`
	Area wsArea `json:"area"`
}

// wsRecording is a WS/2 recording, nested inside a track.
type wsRecording struct {
	ID             string           `json:"id"`
	Title          string           `json:"title"`
	Length         int              `json:"length"` // milliseconds
	Disambiguation string           `json:"disambiguation"`
	Video          bool             `json:"video"`
	Aliases        []string         `json:"aliases"`
	ISRCs          []string         `json:"isrcs"`
	ArtistCredit   []wsArtistCredit `json:"artist-credit"`
}

// wsTrack is one WS/2 track within a medium. ArtistCredit is always
// nil here: get_tracks_from_medium falls back to the nested
// recording's own artist-credit when the track-level one is falsy, so
// a per-track credit is carried on wsRecording instead, matching the
// smoke fixture's shape (corpus/README.md, "The synth tool").
type wsTrack struct {
	ID           string           `json:"id"`
	Position     string           `json:"position"`
	Number       string           `json:"number"`
	Title        string           `json:"title"`
	Length       int              `json:"length"` // milliseconds
	ArtistCredit []wsArtistCredit `json:"artist-credit"`
	Recording    wsRecording      `json:"recording"`
}

// wsMedium is one WS/2 medium (disc).
type wsMedium struct {
	Position   int       `json:"position"`
	Format     string    `json:"format"`
	Title      string    `json:"title"`
	TrackCount int       `json:"track-count"`
	Tracks     []wsTrack `json:"tracks"`
}

// wsRelease is the WS/2 release document [buildRelease] emits: modeled
// on internal/harness/oracle/testdata/cases/smoke/responses/release.json,
// the proven minimal shape a real derivation accepts, and cross-checked
// against beetsplug/musicbrainz.py's own album_info/track_info parsing
// (pinned beets 2.13.1).
type wsRelease struct {
	ID                 string               `json:"id"`
	Title              string               `json:"title"`
	Status             string               `json:"status"`
	Disambiguation     string               `json:"disambiguation"`
	ASIN               *string              `json:"asin"`
	Barcode            *string              `json:"barcode"`
	Country            string               `json:"country"`
	Date               string               `json:"date"`
	Quality            string               `json:"quality"`
	Packaging          *string              `json:"packaging"`
	TextRepresentation wsTextRepresentation `json:"text-representation"`
	ArtistCredit       []wsArtistCredit     `json:"artist-credit"`
	ReleaseGroup       wsReleaseGroup       `json:"release-group"`
	LabelInfo          []wsLabelInfo        `json:"label-info"`
	ReleaseEvents      []wsReleaseEvent     `json:"release-events"`
	Media              []wsMedium           `json:"media"`
}

// buildRelease renders spec's release into the WS/2 document pinned
// beets fetches at its ID-lookup request.
func buildRelease(spec *ReleaseSpec) *wsRelease {
	labelInfo := []wsLabelInfo{}
	if spec.Label != "" || spec.CatalogNumber != "" {
		entry := wsLabelInfo{}
		if spec.CatalogNumber != "" {
			catno := spec.CatalogNumber
			entry.CatalogNumber = &catno
		}
		if spec.Label != "" {
			entry.Label = &wsLabel{Name: spec.Label}
		}
		labelInfo = append(labelInfo, entry)
	}

	media := make([]wsMedium, len(spec.Media))
	for mi, m := range spec.Media {
		tracks := make([]wsTrack, len(m.Tracks))
		for ti, tr := range m.Tracks {
			artist := spec.Artist
			if tr.ArtistCredit != "" {
				artist = tr.ArtistCredit
			}
			recordingID := pseudoUUID(fmt.Sprintf("%s|recording|%d|%d", spec.MBAlbumID, mi, tr.Position))
			trackID := pseudoUUID(fmt.Sprintf("%s|track|%d|%d", spec.MBAlbumID, mi, tr.Position))
			pos := fmt.Sprintf("%d", tr.Position)
			tracks[ti] = wsTrack{
				ID:           trackID,
				Position:     pos,
				Number:       pos,
				Title:        "",
				Length:       int(tr.Length * 1000),
				ArtistCredit: nil,
				Recording: wsRecording{
					ID:             recordingID,
					Title:          tr.Title,
					Length:         int(tr.Length * 1000),
					Disambiguation: "",
					Video:          false,
					Aliases:        []string{},
					ISRCs:          []string{},
					ArtistCredit:   singleArtistCredit(artist, spec.VariousArtists && tr.ArtistCredit == ""),
				},
			}
		}
		media[mi] = wsMedium{
			Position:   mi + 1,
			Format:     m.Format,
			Title:      "",
			TrackCount: len(tracks),
			Tracks:     tracks,
		}
	}

	return &wsRelease{
		ID:             spec.MBAlbumID,
		Title:          spec.Album,
		Status:         "Official",
		Disambiguation: spec.Disambiguation,
		ASIN:           nil,
		Barcode:        nil,
		Country:        spec.Country,
		Date:           spec.Date,
		Quality:        "normal",
		Packaging:      nil,
		TextRepresentation: wsTextRepresentation{
			Script:   "Latn",
			Language: "eng",
		},
		ArtistCredit: singleArtistCredit(spec.Artist, spec.VariousArtists),
		ReleaseGroup: wsReleaseGroup{
			ID:               pseudoUUID(spec.MBAlbumID + "|release-group"),
			Title:            spec.Album,
			PrimaryType:      "Album",
			SecondaryTypes:   []string{},
			FirstReleaseDate: spec.Date,
			Disambiguation:   "",
			Aliases:          []string{},
		},
		LabelInfo: labelInfo,
		ReleaseEvents: []wsReleaseEvent{{
			Date: spec.Date,
			Area: wsArea{
				ID:            pseudoUUID(spec.MBAlbumID + "|area|" + spec.Country),
				Name:          spec.Country,
				ISO31661Codes: []string{spec.Country},
			},
		}},
		Media: media,
	}
}

// marshalRelease encodes rel as the verbatim response body a case's
// responses/001.json carries.
func marshalRelease(rel *wsRelease) ([]byte, error) {
	data, err := json.MarshalIndent(rel, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("synth: encode release document: %w", err)
	}
	return append(data, '\n'), nil
}
