package matcher

import "time"

// addAlbumAlbum is distance()'s album component
// (autotag/distance.py:450-451): always adds, because an album title
// is never optional the way media, mediums, or country are.
func addAlbumAlbum(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	dist.AddString("album", &likelies.Album, &info.Album)
}
