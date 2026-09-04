package matcher

import "time"

// AlbumPair is one item matched to one candidate track: the input the
// album-level tracks, missing_tracks, and unmatched_tracks components
// read, never produce. Beets' own distance() takes this pairing as
// item_info_pairs, a caller-supplied list (distance.py:429-432,
// 528-530); computing it is 3d's assignment work, and a component
// here that reached for a solver would be out of scope.
type AlbumPair struct {
	Item ItemFields
	Info *TrackInfo
}

// addAlbumTracks is distance()'s tracks component
// (autotag/distance.py:526-530): per matched pair, the pair's own
// track distance (scored with inclArtist = info.VA) is folded in two
// ways: its normalized Distance() is added as one penalty under
// "tracks", and the full sub-distance is recorded in dist.Tracks so a
// caller can inspect why a given pair scored what it did.
//
// This always runs, even with zero pairs, matching beets' own
// unconditional for loop; zero pairs simply add nothing.
func addAlbumTracks(dist *Distance, tun Tunables, _ Likelies, info *AlbumInfo, _ []ItemFields, pairs []AlbumPair, _ time.Time) {
	for _, pair := range pairs {
		trackDist := DistanceForTrack(tun, pair.Item, pair.Info, info.VA)
		dist.Add("tracks", trackDist.Distance())
		dist.Tracks = append(dist.Tracks, TrackDistance{Info: pair.Info, Dist: trackDist})
	}
}
