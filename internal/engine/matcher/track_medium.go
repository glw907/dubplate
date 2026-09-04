package matcher

// trackMediumComponent is track_distance's disc component
// (autotag/distance.py:420-422): it fires only when both the
// candidate's medium and the item's disc number are known, penalizing
// a mismatch between the two.
var trackMediumComponent = trackComponent{
	Name: "medium",
	Key:  "medium",
	Add:  addTrackMedium,
}

// addTrackMedium declines when info carries no medium or the item
// carries no disc number.
func addTrackMedium(dist *Distance, _ Tunables, item ItemFields, info *TrackInfo, _ bool) {
	if info.Medium == 0 || item.Disc == 0 {
		return
	}
	dist.AddExpr("medium", item.Disc != info.Medium)
}
