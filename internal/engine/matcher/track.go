package matcher

import (
	"math"
	"slices"
	"strings"
)

// DistanceForTrack is the Go port of beets' track_distance
// (autotag/distance.py:383-426): the significance of one candidate
// track's metadata against the item proposed to match it. It is named
// DistanceForTrack rather than TrackDistance because that identifier is
// already the pairing type Distance.Tracks holds. The calls run in
// beets' own order, load-bearing for the reason DistanceForAlbum gives.
//
// inclArtist is a parameter, not a config read, because beets' two call
// sites disagree on it: the assignment solve calls track_distance with
// inclArtist false for every track, while the final per-track
// distances a user sees pass album_info.va. The two objectives diverge
// exactly on VA albums, where the solve wants every track scored the
// same way regardless of artist credit but the final distance should
// still catch a mismatched one.
func DistanceForTrack(tun Tunables, item ItemFields, info *TrackInfo, inclArtist bool) *Distance {
	dist := NewDistance(tun)
	addTrackLength(dist, tun.TrackLengthGrace, tun.TrackLengthMax, item, info)
	addTrackTitle(dist, item, info)
	addTrackArtist(dist, item, info, inclArtist)
	addTrackIndex(dist, item, info)
	addTrackID(dist, item, info)
	addTrackMedium(dist, item, info)
	addTrackDataSource(dist, item, info)
	return dist
}

// addTrackLength is track_distance's length component
// (autotag/distance.py:396-399): it grades the part of the gap beyond
// grace, ratioed against maxDiff. A candidate with an unknown length
// contributes nothing, not even a 0.0 penalty, rather than being scored
// against a made-up zero.
func addTrackLength(dist *Distance, grace, maxDiff float64, item ItemFields, info *TrackInfo) {
	if info.Length == 0 {
		return
	}
	diff := math.Abs(item.Length-info.Length) - grace
	dist.AddRatio("track_length", diff, maxDiff)
}

// addTrackTitle is track_distance's title component
// (autotag/distance.py:401-402): it always adds, because a track's
// title is never optional the way its length or artist credit are.
func addTrackTitle(dist *Distance, item ItemFields, info *TrackInfo) {
	dist.AddString("track_title", &item.Title, &info.Title)
}

// addTrackArtist is track_distance's artist component
// (autotag/distance.py:404-410): it fires only when the caller asked
// for it (inclArtist), the candidate carries an artist credit, and the
// item's own credit is not one of VA_ARTISTS.
func addTrackArtist(dist *Distance, item ItemFields, info *TrackInfo, inclArtist bool) {
	if !inclArtist || info.Artist == "" || IsVariousArtistsName(item.Artist) {
		return
	}
	dist.AddString("track_artist", &item.Artist, &info.Artist)
}

// vaArtists is VA_ARTISTS (autotag/distance.py:29): item artist
// credits that mark a various-artists track, compared
// case-insensitively. [IsVariousArtistsName] is its one exported
// reader.
var vaArtists = []string{"", "various artists", "various", "va", "unknown"}

// IsVariousArtistsName reports whether credit is one of VA_ARTISTS
// (autotag/distance.py:29), compared case-insensitively. It has two
// consumers: the track artist component, which declines when an item's
// own credit is one of these, and internal/engine/musicbrainz's
// va_likely, which treats the album's likely artist being one of these
// as evidence of a various-artists release (autotag/match.py:453).
func IsVariousArtistsName(credit string) bool {
	return slices.Contains(vaArtists, strings.ToLower(credit))
}

// addTrackIndex is track_distance's index component
// (autotag/distance.py:412-414): it fires only when both the
// candidate's own index and the item's track number are known, and
// track_index_changed (distance.py:376-380) accepts the item's track
// number against either the per-disc or the absolute index.
func addTrackIndex(dist *Distance, item ItemFields, info *TrackInfo) {
	if info.Index == 0 || item.Track == 0 {
		return
	}
	changed := item.Track != info.MediumIndex && item.Track != info.Index
	dist.AddExpr("track_index", changed)
}

// addTrackID is track_distance's ID component
// (autotag/distance.py:416-418): it fires only when the item carries
// its own MusicBrainz track ID, since a track never tagged with one has
// nothing to compare against the candidate's.
func addTrackID(dist *Distance, item ItemFields, info *TrackInfo) {
	if item.MBTrackID == "" {
		return
	}
	dist.AddExpr("track_id", item.MBTrackID != info.TrackID)
}

// addTrackMedium is track_distance's disc component
// (autotag/distance.py:420-422): it fires only when both the
// candidate's medium and the item's disc number are known, penalizing a
// mismatch between the two.
func addTrackMedium(dist *Distance, item ItemFields, info *TrackInfo) {
	if info.Medium == 0 || item.Disc == 0 {
		return
	}
	dist.AddExpr("medium", item.Disc != info.Medium)
}

// addTrackDataSource is track_distance's data-source component
// (autotag/distance.py:424). AddDataSource carries its own fire
// condition inside the accumulator rather than here.
func addTrackDataSource(dist *Distance, item ItemFields, info *TrackInfo) {
	dist.AddDataSource(item.DataSource, info.DataSource)
}
