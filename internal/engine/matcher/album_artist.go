package matcher

import "time"

// addAlbumArtist is distance()'s artist component
// (autotag/distance.py:446-448): skipped entirely on a various-
// artists release, because a VA release has no single artist credit
// to compare a candidate against.
//
// This file carries only the fire condition and the adder call, one
// component per file, mirroring track.go's shape. Task 8's score.go
// wires this function into the declared albumComponents slice in
// contract 8's fixed album order; that declaration, not this file, is
// what makes the order load-bearing.
func addAlbumArtist(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if info.VA {
		return
	}
	dist.AddString("artist", &likelies.Artist, &info.Artist)
}
