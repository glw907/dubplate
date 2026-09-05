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
// (musicbrainz.py:673-681). artist_sort is deliberately NOT produced:
// beets computes it at musicbrainz.py:337, but R6 puts it in rung 8's
// applied-field territory and out of the ruling-E ledger's graded set,
// so carrying it here would be tested dead data of exactly the shape
// 3a's close had to remove (fold, M-i).
//
// artist_credit is absent for the same reason and one more: it is the
// only field whose value depends on musicbrainz.aliases_as_credits,
// which R8 keeps out of Tunables, so producing the string would need a
// knob nothing else reads.
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
// Callers pass a credit list Validate has accepted, at the release level
// and at every surviving track's resolved level both, so the empty-list
// case that raises IndexError at musicbrainz.py:350 cannot reach here.
// Revision 1 claimed that guarantee against a Validate that checked the
// release level only, which left this function indexing element zero of
// an empty slice on track and recording credits (fold, M-d).
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
// AlbumInfo.VA, which the album artist component reads to skip itself
// entirely (matcher/album_artist.go:13-15).
func applyVariousArtists(c credits, tun matcher.Tunables) (credits, bool) {
	if c.ArtistID != variousArtistsID {
		return c, false
	}
	c.Artist = tun.VAName
	return c, true
}
