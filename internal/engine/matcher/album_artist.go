package matcher

import "time"

// addAlbumArtist is distance()'s artist component
// (autotag/distance.py:446-448): skipped entirely on a various-
// artists release, because a VA release has no single artist credit
// to compare a candidate against.
//
// score.go's albumComponents declaration, not this file, is what places
// the component in contract 8's fixed order.
func addAlbumArtist(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if info.VA {
		return
	}
	dist.AddString("artist", &likelies.Artist, &info.Artist)
}
