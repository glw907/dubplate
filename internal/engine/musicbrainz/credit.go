package musicbrainz

import (
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// variousArtistsID is MusicBrainz's own special-purpose artist entity
// for various-artists releases (beetsplug/musicbrainz.py:43).
const variousArtistsID = "89ad4ac3-39f7-470e-963a-56509c546377"

// credits carries the artist strings _parse_artist_credits produces
// that this port needs, and no others. Artist reaches
// matcher.AlbumInfo and matcher.TrackInfo; ArtistID is here because the
// various-artists override is keyed on the first artist's id
// (musicbrainz.py:673-681). The matcher reads neither artist_sort
// (musicbrainz.py:337) nor artist_credit, so neither is produced.
type credits struct {
	Artist   string
	ArtistID string
}

// parseArtistCredits flattens an artist-credit array, mirroring
// _parse_artist_credits (beetsplug/musicbrainz.py:301-357). The
// joinphrase is appended after the LAST element too, so a malformed
// credit whose final joinphrase is non-empty leaks a trailing separator
// straight into the weight-3.0 album artist penalty; that is beets'
// behavior and the port reproduces it.
//
// A preferred alias supplies the artist name where one is selected,
// because beets takes the name from `alias or el["artist"]` (:328-330)
// rather than from the credit element. The credit element's own display
// name feeds artist_credit alone, which credits does not carry.
//
// Callers pass a credit list validate has accepted, at the release level
// and at every surviving track's resolved level both, so the empty-list
// case that raises IndexError at musicbrainz.py:350 cannot reach here.
func parseArtistCredits(parts []CreditPart, tun matcher.Tunables) credits {
	var artist strings.Builder
	for _, part := range parts {
		name := part.Artist.Name
		if alias := preferredAlias(part.Artist.Aliases, tun); alias != nil {
			name = alias.Name
		}
		artist.WriteString(name)
		artist.WriteString(part.JoinPhrase)
	}

	return credits{Artist: artist.String(), ArtistID: parts[0].Artist.ID}
}

// applyVariousArtists overwrites c's artist strings with tun.VAName
// when c names the various-artists entity, mirroring
// musicbrainz.py:673-681, and reports whether it fired. The report is
// AlbumInfo.VA, which the matcher's addAlbumArtist reads to skip the
// album artist component entirely.
func applyVariousArtists(c credits, tun matcher.Tunables) (credits, bool) {
	if c.ArtistID != variousArtistsID {
		return c, false
	}
	c.Artist = tun.VAName
	return c, true
}
