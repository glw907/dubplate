package matcher

import "time"

// addAlbumMediums is distance()'s mediums component
// (autotag/distance.py:467-469): fires only when the item set agrees
// on a disc total and the candidate carries its own disc count.
// AddNumber makes each disc of difference a full weighted unit
// (contract 6: one 1.0 penalty per unit, not a ratio).
func addAlbumMediums(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if likelies.DiscTotal == 0 || info.Mediums == 0 {
		return
	}
	dist.AddNumber("mediums", likelies.DiscTotal, info.Mediums)
}
