package matcher

import "time"

// album_extras.go is the plan's one named exception to one-component-
// one-file (fold, M-i): missing_tracks and unmatched_tracks are one
// counting rule read from both ends of the same pairing, and
// splitting them across two files would put half a rule in each.
// Both still get their own declared component entry (so Task 8's
// order test still sees fourteen names) and their own fire and
// decline test below.
//
// Both components count only, and never contribute contents to the
// pairing or the matched-pair distances: missing_tracks is the
// candidate's own tracks beyond what got paired, and unmatched_tracks
// is the item set's own entries beyond what got paired. Neither
// component orders anything; sorting the extras for display is 3d's
// job (contract 8's own text: "extras count, never contribute
// contents, sort orders are 3d's").

// addAlbumMissingTracks is distance()'s missing_tracks component
// (autotag/distance.py:532-534): one 1.0 penalty for every candidate
// track beyond the matched pairs.
func addAlbumMissingTracks(dist *Distance, _ Tunables, _ Likelies, info *AlbumInfo, _ []ItemFields, pairs []AlbumPair, _ time.Time) {
	for range max(0, len(info.Tracks)-len(pairs)) {
		dist.Add("missing_tracks", 1)
	}
}

// addAlbumUnmatchedTracks is distance()'s unmatched_tracks component
// (autotag/distance.py:536-538): one 1.0 penalty for every item
// beyond the matched pairs.
func addAlbumUnmatchedTracks(dist *Distance, _ Tunables, _ Likelies, _ *AlbumInfo, items []ItemFields, pairs []AlbumPair, _ time.Time) {
	for range max(0, len(items)-len(pairs)) {
		dist.Add("unmatched_tracks", 1)
	}
}
