package matcher

// trackTitleComponent is track_distance's title component
// (autotag/distance.py:401-402): it always adds, because a track's
// title is never optional the way its length or artist credit are.
var trackTitleComponent = trackComponent{
	Name: "track_title",
	Key:  "track_title",
	Add:  addTrackTitle,
}

// addTrackTitle has no fire condition: it runs on every pair. item.Title
// and info.Title are already typed-default strings, never nil, so the
// pointers AddString takes are always non-nil here.
func addTrackTitle(dist *Distance, _ Tunables, item ItemFields, info *TrackInfo, _ bool) {
	dist.AddString("track_title", &item.Title, &info.Title)
}
