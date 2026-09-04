package matcher

import "time"

// addAlbumDisambig is distance()'s albumdisambig component
// (autotag/distance.py:514-518): fires only when both the item set's
// likely disambiguation string and the candidate's own are set.
func addAlbumDisambig(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if likelies.AlbumDisambig == "" || info.AlbumDisambig == "" {
		return
	}
	dist.AddString("albumdisambig", &likelies.AlbumDisambig, &info.AlbumDisambig)
}
