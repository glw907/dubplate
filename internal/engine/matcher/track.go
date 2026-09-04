package matcher

// trackComponent is one line of track_distance's fixed order: a name,
// the distance key it writes penalties under, and the function that
// fires it (contract 8: which components fire, and in what order,
// determines the denominator). score.go declares the album-level
// composition the same way, so both sides compose through a declared
// slice rather than inline calls.
type trackComponent struct {
	Name string
	Key  string
	Add  func(dist *Distance, tun Tunables, item ItemFields, info *TrackInfo, inclArtist bool)
}

// trackComponents is track_distance's fixed order (contract 8:
// track_length, track_title, track_artist, track_index, track_id,
// medium, data_source). The order is load-bearing, never resorted: a
// component that declines to add anything must decline in exactly the
// place beets declines.
var trackComponents = []trackComponent{
	trackLengthComponent,
	trackTitleComponent,
	trackArtistComponent,
	trackIndexComponent,
	trackIDComponent,
	trackMediumComponent,
	trackDataSourceComponent,
}

// DistanceForTrack is the Go port of beets' track_distance
// (autotag/distance.py:383-426): the significance of one candidate
// track's metadata against the item proposed to match it. It is named
// DistanceForTrack rather than TrackDistance because that identifier is
// already the pairing type Distance.Tracks holds (penalty.go).
//
// inclArtist is a parameter, not a config read, because beets' two call
// sites for this function disagree on it: the assignment solve (3d)
// calls track_distance with inclArtist=false for every track, while the
// final per-track distances a user sees pass album_info.va. The two
// objectives diverge exactly on VA albums, where the solve wants every
// track scored the same way regardless of artist credit but the final
// distance should still catch a mismatched one. 3a does not build the
// solve; this parameter exists so that divergence is explicit here
// rather than rediscovered by 3d's tie-mirroring work.
func DistanceForTrack(tun Tunables, item ItemFields, info *TrackInfo, inclArtist bool) *Distance {
	dist := NewDistance(tun)
	for _, c := range trackComponents {
		c.Add(dist, tun, item, info, inclArtist)
	}
	return dist
}
