package matcher

import (
	"slices"
	"strings"
)

// trackArtistComponent is track_distance's artist component
// (autotag/distance.py:404-410): it fires only when the caller asked
// for it (inclArtist), the candidate carries an artist credit, and the
// item's own credit is not one of VA_ARTISTS.
var trackArtistComponent = trackComponent{
	Name: "track_artist",
	Key:  "track_artist",
	Add:  addTrackArtist,
}

// vaArtists is VA_ARTISTS (autotag/distance.py:29): item artist
// credits that mark a various-artists track, compared
// case-insensitively. It stays unexported; [IsVariousArtistsName] is
// its one exported reader.
var vaArtists = []string{"", "various artists", "various", "va", "unknown"}

// IsVariousArtistsName reports whether credit is one of VA_ARTISTS
// (autotag/distance.py:29), compared case-insensitively. It has two
// consumers: the track artist component, which declines when an item's
// own credit is one of these, and internal/engine/musicbrainz's
// va_likely, which treats the album's likely artist being one of these
// as evidence of a various-artists release (autotag/match.py:453). One
// list, two readers.
func IsVariousArtistsName(credit string) bool {
	return slices.Contains(vaArtists, strings.ToLower(credit))
}

// addTrackArtist declines when inclArtist is false, when the candidate
// carries no artist credit, or when the item's credit names a
// various-artists placeholder. inclArtist is a parameter rather than a
// config read: the assignment solve (3d) calls track_distance with
// inclArtist=false while the final per-track distances pass
// album_info.va, and the two objectives diverge exactly on VA albums
// (match.py's Track assignment). 3a does not build the solve; this
// parameter exists so that divergence is explicit now rather than
// rediscovered later.
func addTrackArtist(dist *Distance, _ Tunables, item ItemFields, info *TrackInfo, inclArtist bool) {
	if !inclArtist || info.Artist == "" || IsVariousArtistsName(item.Artist) {
		return
	}
	dist.AddString("track_artist", &item.Artist, &info.Artist)
}
