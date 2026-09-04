package matcher

import "time"

// addAlbumID is distance()'s album_id component
// (autotag/distance.py:520-524): fires only when the item set agrees
// on a MusicBrainz album ID, and compares it against the candidate's
// own. Note the argument order: value is the likely ID, and the
// candidate's ID is the (scalar, auto-wrapped) option, matching
// beets' own `add_equality("album_id", likelies["mb_albumid"],
// album_info.album_id)`.
func addAlbumID(dist *Distance, _ Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if likelies.MBAlbumID == "" {
		return
	}
	dist.AddEquality("album_id", likelies.MBAlbumID, info.AlbumID)
}
