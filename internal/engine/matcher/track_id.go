package matcher

// trackIDComponent is track_distance's ID component
// (autotag/distance.py:416-418): it fires only when the item carries
// its own MusicBrainz track ID.
var trackIDComponent = trackComponent{
	Name: "track_id",
	Key:  "track_id",
	Add:  addTrackID,
}

// addTrackID declines when the item carries no MusicBrainz track ID:
// a track never tagged with one has nothing to compare against the
// candidate's.
func addTrackID(dist *Distance, _ Tunables, item ItemFields, info *TrackInfo, _ bool) {
	if item.MBTrackID == "" {
		return
	}
	dist.AddExpr("track_id", item.MBTrackID != info.TrackID)
}
