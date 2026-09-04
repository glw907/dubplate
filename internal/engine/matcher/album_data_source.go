package matcher

import "time"

// addAlbumDataSource is distance()'s data_source component
// (autotag/distance.py:540): always calls AddDataSource, which
// carries its own fire condition (contract 6's three-way test) inside
// the accumulator rather than here, exactly like
// addTrackDataSource.
func addAlbumDataSource(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	dist.AddDataSource(likelies.DataSource, info.DataSource)
}
