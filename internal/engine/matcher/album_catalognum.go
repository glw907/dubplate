package matcher

import "time"

// addAlbumCatalogNum is distance()'s catalognum component
// (autotag/distance.py:508-512): fires only when both the item set's
// likely catalog number and the candidate's own are set.
func addAlbumCatalogNum(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if likelies.CatalogNum == "" || info.CatalogNum == "" {
		return
	}
	dist.AddString("catalognum", &likelies.CatalogNum, &info.CatalogNum)
}
