package matcher

import "time"

// addAlbumLabel is distance()'s label component
// (autotag/distance.py:504-506): fires only when both the item set's
// likely label and the candidate's own label are set.
func addAlbumLabel(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if likelies.Label == "" || info.Label == "" {
		return
	}
	dist.AddString("label", &likelies.Label, &info.Label)
}
