package matcher

import "math"

// trackLengthComponent is track_distance's length component
// (autotag/distance.py:396-399): it fires only when the candidate
// carries a known length, and grades the part of the gap beyond
// TrackLengthGrace, ratioed against TrackLengthMax.
var trackLengthComponent = trackComponent{
	Name: "track_length",
	Key:  "track_length",
	Add:  addTrackLength,
}

// addTrackLength declines when info carries no length: a candidate
// with an unknown duration contributes nothing, not even a 0.0
// penalty, rather than being scored against a made-up zero.
func addTrackLength(dist *Distance, tun Tunables, item ItemFields, info *TrackInfo, _ bool) {
	if info.Length == 0 {
		return
	}
	diff := math.Abs(item.Length-info.Length) - tun.TrackLengthGrace
	dist.AddRatio("track_length", diff, tun.TrackLengthMax)
}
