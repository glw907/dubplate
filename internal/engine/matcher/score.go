package matcher

import "time"

// albumComponent is one line of distance()'s fixed album-level order: a
// name, the distance key it writes penalties under, and the function
// that fires it. It mirrors trackComponent (track.go), so both sides
// compose through a declared slice rather than inline calls.
//
// Key is read by TestComponentKeysAreWeighted, which asserts every
// component's key names an entry in Tunables.DistanceWeights.
type albumComponent struct {
	Name string
	Key  string
	Add  func(dist *Distance, tun Tunables, likelies Likelies, info *AlbumInfo, items []ItemFields, pairs []AlbumPair, evalDate time.Time)
}

// albumComponents is distance()'s fixed album-level order (contract 8),
// artist first and data_source last, never resorted.
//
// The order is load-bearing. Insertion order into Distance fixes the key
// sequence AddedKeys and GenericPenaltyKeys report, and the summation
// order RawDistance and MaxDistance walk, and float addition is not
// associative, so a reordered slice both reports a different sequence
// and sums to a different number. A component that declines to add
// anything must decline in exactly the place beets declines: firing
// where beets does not changes the reported sequence even at a zero
// contribution.
var albumComponents = []albumComponent{
	{Name: "artist", Key: "artist", Add: addAlbumArtist},
	{Name: "album", Key: "album", Add: addAlbumAlbum},
	{Name: "media", Key: "media", Add: addAlbumMedia},
	{Name: "mediums", Key: "mediums", Add: addAlbumMediums},
	{Name: "year", Key: "year", Add: addAlbumYear},
	{Name: "country", Key: "country", Add: addAlbumCountry},
	{Name: "label", Key: "label", Add: addAlbumLabel},
	{Name: "catalognum", Key: "catalognum", Add: addAlbumCatalogNum},
	{Name: "albumdisambig", Key: "albumdisambig", Add: addAlbumDisambig},
	{Name: "album_id", Key: "album_id", Add: addAlbumID},
	{Name: "tracks", Key: "tracks", Add: addAlbumTracks},
	{Name: "missing_tracks", Key: "missing_tracks", Add: addAlbumMissingTracks},
	{Name: "unmatched_tracks", Key: "unmatched_tracks", Add: addAlbumUnmatchedTracks},
	{Name: "data_source", Key: "data_source", Add: addAlbumDataSource},
}

// DistanceForAlbum is the Go port of beets' distance
// (autotag/distance.py:429-541): the significance of one candidate
// release's metadata, and its item-to-track pairing, against a user's
// existing items. It adds nothing of its own; every penalty comes from
// a component in albumComponents, walked in beets' own fixed order, and
// this function only computes the shared Likelies input, composes, and
// returns the result. A rule change belongs to one component file and
// its test, never to this function.
//
// evalDate is a parameter, never a clock read (contract 9): the year
// component's ratio branches divide by a span ending at evalDate's
// year, and a package that read time.Now would make every distance
// computation depend on when it happened to run. This function performs
// no I/O and imports nothing under internal/harness.
func DistanceForAlbum(items []ItemFields, info *AlbumInfo, pairs []AlbumPair, tun Tunables, evalDate time.Time) *Distance {
	likelies, _ := MostCommonTags(items)
	dist := NewDistance(tun)
	for _, c := range albumComponents {
		c.Add(dist, tun, likelies, info, items, pairs, evalDate)
	}
	return dist
}
