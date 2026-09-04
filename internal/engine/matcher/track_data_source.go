package matcher

// trackDataSourceComponent is track_distance's data-source component
// (autotag/distance.py:424): it always calls AddDataSource, which
// carries its own fire condition (contract 6's three-way test) inside
// the accumulator rather than here.
var trackDataSourceComponent = trackComponent{
	Name: "data_source",
	Key:  "data_source",
	Add:  addTrackDataSource,
}

func addTrackDataSource(dist *Distance, _ Tunables, item ItemFields, info *TrackInfo, _ bool) {
	dist.AddDataSource(item.DataSource, info.DataSource)
}
