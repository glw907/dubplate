package matcher

// trackIndexComponent is track_distance's index component
// (autotag/distance.py:412-414): it fires only when both the
// candidate's own index and the item's track number are known, and
// tolerates either per-disc or absolute numbering.
var trackIndexComponent = trackComponent{
	Name: "track_index",
	Key:  "track_index",
	Add:  addTrackIndex,
}

// addTrackIndex declines when info carries no index or the item
// carries no track number. track_index_changed (distance.py:376-380)
// accepts item.track against either info.MediumIndex (per-disc) or
// info.Index (absolute).
func addTrackIndex(dist *Distance, _ Tunables, item ItemFields, info *TrackInfo, _ bool) {
	if info.Index == 0 || item.Track == 0 {
		return
	}
	changed := item.Track != info.MediumIndex && item.Track != info.Index
	dist.AddExpr("track_index", changed)
}
