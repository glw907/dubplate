package musicbrainz

import (
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// dataSource is the literal every AlbumInfo and TrackInfo this package
// produces shares (album_info.data_source / track_info.data_source, set
// through the Info base class's data_source parameter, which every
// album_for_id call site fills from the plugin's own class name;
// beets/metadata_plugins.py:182-188, cls.__name__.replace("Plugin", "")
// applied to MusicBrainzPlugin).
const dataSource = "MusicBrainz"

// noLabelSentinel is the exact literal beets compares a label name
// against, case-sensitively and without stripping, to decide whether a
// label counts as present (beetsplug/musicbrainz.py:504).
const noLabelSentinel = "[no label]"

// TranslateAlbum builds the matcher's candidate vocabulary from one
// decoded WS/2 release, mirroring MusicBrainzPlugin.album_info
// (beetsplug/musicbrainz.py:631-695). It issues no request and reads no
// clock: the evaluation date is the distance path's input, not this
// layer's, and internal/harness/conform's no-time.Now walk enforces
// that.
//
// The composition runs in album_info's own order, and the order is
// load-bearing in exactly one place: the release-event assignment
// overwrites Country and Year AFTER the release-group parse has already
// supplied the original-date fallback, and that fallback is a
// truthiness test on the resolved date, not a nil check. Moving
// the release-event step earlier would let an empty date wipe out a
// fallback that has not been computed yet.
//
// The browse-fallback guard is not this function's obligation:
// internal/harness/conform's TestNoCommittedReleaseReachesTheBrowseThreshold
// reads the corpus directly, because this package has no channel to a
// request the client has not issued.
//
// It returns an error when rel cannot be translated at all; every other
// beets-raising site is handled as a zero value and the reasons are
// recorded on Release and on the functions concerned.
func TranslateAlbum(rel *Release, tun matcher.Tunables) (*matcher.AlbumInfo, error) {
	if err := rel.validate(); err != nil {
		return nil, err
	}

	built := extractTracks(rel, tun)
	media := deriveMedia(built)
	mediums := len(rel.Media)

	album := keyWithPreferredAlias(rel.Aliases, rel.Title, tun)

	credit, va := applyVariousArtists(parseArtistCredits(rel.ArtistCredit, tun), tun)

	originalYear := parseYear(rel.ReleaseGroup.FirstReleaseDate)

	var catalogNum string
	if len(rel.LabelInfo) > 0 && rel.LabelInfo[0].CatalogNumber != nil {
		catalogNum = *rel.LabelInfo[0].CatalogNumber
	}
	var label string
	if len(rel.LabelInfo) > 0 && rel.LabelInfo[0].Label != nil && rel.LabelInfo[0].Label.Name != noLabelSentinel {
		label = rel.LabelInfo[0].Label.Name
	}

	country, date := preferredReleaseEvent(rel, tun)
	year := originalYear
	if date != "" {
		year = parseYear(date)
	}

	tracks := make([]matcher.TrackInfo, len(built))
	for i, t := range built {
		t.Info.DataSource = dataSource
		tracks[i] = t.Info
	}

	return &matcher.AlbumInfo{
		Artist:        credit.Artist,
		Album:         album,
		VA:            va,
		Media:         media,
		Mediums:       mediums,
		Year:          year,
		OriginalYear:  originalYear,
		Country:       country,
		Label:         label,
		CatalogNum:    catalogNum,
		AlbumDisambig: rel.Disambiguation,
		AlbumID:       rel.ID,
		DataSource:    dataSource,
		Tracks:        tracks,
	}, nil
}

// preferredReleaseEvent returns the country and date the translation
// writes, mirroring _preferred_release_event
// (beetsplug/musicbrainz.py:165-177): an outer loop over
// tun.Preferred.Countries in configured order and an inner loop over
// rel.ReleaseEvents, returning the first configured country any event's
// area lists, together with that event's date. Otherwise it returns the
// release's own top-level country and date.
//
// The membership test is a literal, case-sensitive slices.Contains
// against the area's ISO codes, and the country returned is the
// CONFIGURED string rather than the area's own code. The same
// Tunables.Preferred.Countries field is compiled as an anchored,
// case-insensitive regex list by the matcher's compileCountryPatterns,
// so one config list carries two matching regimes in two packages.
func preferredReleaseEvent(rel *Release, tun matcher.Tunables) (country string, date string) {
	for _, configured := range tun.Preferred.Countries {
		for _, event := range rel.ReleaseEvents {
			if event.Area == nil {
				continue
			}
			if slices.Contains(event.Area.ISO31661Codes, configured) {
				return configured, event.Date
			}
		}
	}

	if rel.Country != nil {
		country = *rel.Country
	}
	return country, rel.Date
}

// parseYear returns the year of a MusicBrainz date, mirroring _get_date
// (beetsplug/musicbrainz.py:180-200) for the one part the matcher reads.
// It is zero when the year is absent or unparsable, however the month and
// day that follow read.
func parseYear(s string) int {
	year, _, _ := strings.Cut(s, "-")
	n, err := strconv.Atoi(year)
	if err != nil {
		return 0
	}
	return n
}
