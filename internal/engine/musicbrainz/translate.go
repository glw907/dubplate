package musicbrainz

import (
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
// layer's (3a contract 9), and internal/harness/conform's widened
// no-time.Now walk is what enforces that.
//
// The composition runs in album_info's own order, and the order is
// load-bearing in exactly one place: the release-event assignment
// overwrites Country and Year AFTER the release-group parse has already
// supplied the original-date fallback, and that fallback is a
// truthiness test on the resolved date, not a nil check (T11). Moving
// the release-event step earlier would let an empty date wipe out a
// fallback that has not been computed yet.
//
// The browse-fallback guard T16 names is not this function's
// obligation: internal/harness/conform's
// TestNoCommittedReleaseReachesTheBrowseThreshold reads the corpus
// directly, because this package has no channel to a request the
// client (3c) has not issued.
func TranslateAlbum(rel *Release, tun matcher.Tunables) (*matcher.AlbumInfo, error) {
	if err := rel.Validate(); err != nil {
		return nil, err
	}

	built := extractTracks(rel, tun)
	media := deriveMedia(built)
	mediums := len(rel.Media)

	album := keyWithPreferredAlias(rel.Aliases, rel.Title, tun)

	credit, va := applyVariousArtists(parseArtistCredits(rel.ArtistCredit, tun), tun)

	originalYear, _, _ := parseDate(rel.ReleaseGroup.FirstReleaseDate)

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
		year, _, _ = parseDate(date)
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
