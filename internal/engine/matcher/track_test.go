package matcher

import (
	"slices"
	"testing"
)

// TestTrackDistance hand-authors the equivalents of the corpus's 16
// track_distance rows: the data-source parameter space (AddDataSource's
// three-way condition in all four states) plus identical, tolerate-va,
// different-artist, and different-title. Every want was measured by
// calling beets.autotag.distance.track_distance in the pinned venv
// (tools/venv/.venv, beets 2.13.1) on exactly these inputs, with
// MusicBrainz alone loaded except where a scenario's own name says
// otherwise, and the comparison is exact rather than tolerant.
func TestTrackDistance(t *testing.T) {
	base := func() (ItemFields, TrackInfo) {
		item := ItemFields{
			Title: "Ember Wharf", Artist: "Northern Lines", Length: 200,
			Track: 3, Disc: 1, MBTrackID: "mbid-1", DataSource: "MusicBrainz",
		}
		info := TrackInfo{
			Title: "Ember Wharf", Artist: "Northern Lines", Length: 200,
			Index: 3, MediumIndex: 3, Medium: 1, TrackID: "mbid-1",
			DataSource: "MusicBrainz",
		}
		return item, info
	}

	t.Run("identical", func(t *testing.T) {
		item, info := base()
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("tolerate-va: a VA solve ignores an artist mismatch", func(t *testing.T) {
		item, info := base()
		item.Artist = "Various Artists"
		dist := DistanceForTrack(DefaultTunables(), item, &info, false)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_index", "track_id", "medium")
	})

	t.Run("different-artist", func(t *testing.T) {
		item, info := base()
		item.Artist = "Quiet Season"
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0.12087912087912088)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("different-title", func(t *testing.T) {
		item, info := base()
		item.Title = "A Quiet Season"
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0.17857142857142858)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("data source: sources match", func(t *testing.T) {
		item, info := base()
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("data source: a mismatch against a set before", func(t *testing.T) {
		item, info := base()
		item.DataSource = "Spotify"
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0.0625)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium", "data_source")
	})

	t.Run("data source: no before and a single loaded source", func(t *testing.T) {
		item, info := base()
		item.DataSource = ""
		info.DataSource = "Spotify"
		tun := DefaultTunables()
		tun.DataSourcePenalties = map[string]float64{"MusicBrainz": 0.5}
		dist := DistanceForTrack(tun, item, &info, true)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("data source: no before and several loaded sources", func(t *testing.T) {
		item, info := base()
		item.DataSource = ""
		info.DataSource = "Spotify"
		tun := DefaultTunables()
		tun.DataSourcePenalties = map[string]float64{"MusicBrainz": 0.5, "Spotify": 0.5}
		dist := DistanceForTrack(tun, item, &info, true)
		wantDistance(t, dist, 0.0625)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium", "data_source")
	})
}

func wantDistance(t *testing.T, dist *Distance, want float64) {
	t.Helper()
	if got := dist.Distance(); got != want {
		t.Errorf("Distance() = %v, want %v", got, want)
	}
}

func wantAddedKeys(t *testing.T, dist *Distance, want ...string) {
	t.Helper()
	if got := dist.AddedKeys(); !slices.Equal(got, want) {
		t.Fatalf("AddedKeys() = %v, want %v", got, want)
	}
}

// The three arithmetic shapes AddRatio's own tests already pin: a negative diff clamps to 0, a diff beyond
// TrackLengthMax clamps to 1, and a diff in between ratios plainly.
// This test's job is the fire condition and the wiring into
// TrackLengthGrace and TrackLengthMax, not AddRatio's arithmetic.
func TestAddTrackLength(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name       string
		itemLength float64
		infoLength float64
		want       []float64
	}{
		{"declines when the candidate carries no length", 200, 0, nil},
		{"within grace clamps to zero", 100, 105, []float64{0}},
		{"beyond max clamps to one", 200, 100, []float64{1}},
		{"a middling gap ratios plainly", 130, 100, []float64{20.0 / 30.0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackLength(dist, tun.TrackLengthGrace, tun.TrackLengthMax, ItemFields{Length: tt.itemLength}, &TrackInfo{Length: tt.infoLength})

			got := dist.penalties["track_length"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackLength(%v, %v) added %v, want %v", tt.itemLength, tt.infoLength, got, tt.want)
			}
		})
	}
}

// track_title carries no fire condition, unlike every other track
// component, so this test's decline
// proof is the edge case a fire-conditioned component would use to
// decline: both titles empty. track_title still adds here, which is
// the point.
func TestAddTrackTitle(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name      string
		itemTitle string
		infoTitle string
		want      []float64
	}{
		{"always adds, even when both titles are empty", "", "", []float64{0}},
		{"adds zero for identical titles", "Ember Wharf", "Ember Wharf", []float64{0}},
		{"adds a nonzero string distance for different titles", "Ember Wharf", "Northern Lines", []float64{0.8461538461538461}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackTitle(dist, ItemFields{Title: tt.itemTitle}, &TrackInfo{Title: tt.infoTitle})

			got := dist.penalties["track_title"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackTitle(%q, %q) added %v, want %v", tt.itemTitle, tt.infoTitle, got, tt.want)
			}
		})
	}
}

func TestAddTrackArtist(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name       string
		inclArtist bool
		itemArtist string
		infoArtist string
		want       []float64
	}{
		{"declines when inclArtist is false", false, "Northern Lines", "Ember Wharf", nil},
		{"declines when the candidate carries no artist credit", true, "Northern Lines", "", nil},
		{"declines when the item carries no artist credit, the empty first member of VA_ARTISTS", true, "", "Ember Wharf", nil},
		{"declines when the item names a various-artists placeholder", true, "Various Artists", "Ember Wharf", nil},
		{"declines case-insensitively against the VA placeholder list", true, "VARIOUS", "Ember Wharf", nil},
		{"declines case-insensitively against the mixed-case unknown placeholder", true, "Unknown", "Ember Wharf", nil},
		{"adds zero for identical artists", true, "Ember Wharf", "Ember Wharf", []float64{0}},
		{"adds a nonzero string distance for different artists", true, "Northern Lines", "Ember Wharf", []float64{0.8461538461538461}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackArtist(dist, ItemFields{Artist: tt.itemArtist}, &TrackInfo{Artist: tt.infoArtist}, tt.inclArtist)

			got := dist.penalties["track_artist"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackArtist(inclArtist=%v, %q, %q) added %v, want %v", tt.inclArtist, tt.itemArtist, tt.infoArtist, got, tt.want)
			}
		})
	}
}

// TestIsVariousArtistsName pins the exported seam musicbrainz's
// va_likely shares with the track artist component: one list, compared
// case-insensitively. A mutation that compares case-sensitively fails
// the "Unknown" row below, the credit the corpus's own
// nonva-va-artists-misflag case carries.
func TestIsVariousArtistsName(t *testing.T) {
	tests := []struct {
		credit string
		want   bool
	}{
		{"", true},
		{"Various Artists", true},
		{"VARIOUS", true},
		{"va", true},
		{"Unknown", true},
		{"Northern Lines", false},
	}

	for _, tt := range tests {
		t.Run(tt.credit, func(t *testing.T) {
			if got := IsVariousArtistsName(tt.credit); got != tt.want {
				t.Errorf("IsVariousArtistsName(%q) = %v, want %v", tt.credit, got, tt.want)
			}
		})
	}
}

func TestAddTrackIndex(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name       string
		itemTrack  int
		infoIndex  int
		infoMedium int
		want       []float64
	}{
		{"declines when the candidate carries no index", 3, 0, 3, nil},
		{"declines when the item carries no track number", 0, 3, 3, nil},
		{"matches the absolute index", 3, 3, 9, []float64{0}},
		{"matches the per-disc index", 9, 3, 9, []float64{0}},
		{"neither index matches", 4, 3, 9, []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			item := ItemFields{Track: tt.itemTrack}
			info := &TrackInfo{Index: tt.infoIndex, MediumIndex: tt.infoMedium}
			addTrackIndex(dist, item, info)

			got := dist.penalties["track_index"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackIndex(track=%v, index=%v, mediumIndex=%v) added %v, want %v",
					tt.itemTrack, tt.infoIndex, tt.infoMedium, got, tt.want)
			}
		})
	}
}

func TestAddTrackID(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		itemID string
		infoID string
		want   []float64
	}{
		{"declines when the item carries no MusicBrainz track ID", "", "mbid-1", nil},
		{"matching IDs add zero", "mbid-1", "mbid-1", []float64{0}},
		{"mismatched IDs add one", "mbid-1", "mbid-2", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackID(dist, ItemFields{MBTrackID: tt.itemID}, &TrackInfo{TrackID: tt.infoID})

			got := dist.penalties["track_id"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackID(%q, %q) added %v, want %v", tt.itemID, tt.infoID, got, tt.want)
			}
		})
	}
}

func TestAddTrackMedium(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name       string
		itemDisc   int
		infoMedium int
		want       []float64
	}{
		{"declines when the candidate carries no medium", 1, 0, nil},
		{"declines when the item carries no disc number", 0, 1, nil},
		{"matching discs add zero", 1, 1, []float64{0}},
		{"mismatched discs add one", 1, 2, []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackMedium(dist, ItemFields{Disc: tt.itemDisc}, &TrackInfo{Medium: tt.infoMedium})

			got := dist.penalties["medium"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackMedium(disc=%v, medium=%v) added %v, want %v", tt.itemDisc, tt.infoMedium, got, tt.want)
			}
		})
	}
}

// TestAddTrackDataSource walks AddDataSource's three-way condition in
// all four states, the same states TestAddDataSource proves at the
// accumulator level; this test proves the component wires
// item.DataSource and info.DataSource into AddDataSource unconditionally,
// carrying no fire condition of its own.
func TestAddTrackDataSource(t *testing.T) {
	twoSources := map[string]float64{"MusicBrainz": 0.5, "Spotify": 0.75}
	oneSource := map[string]float64{"MusicBrainz": 0.5}

	tests := []struct {
		name    string
		sources map[string]float64
		before  string
		after   string
		want    []float64
	}{
		{"sources match", twoSources, "MusicBrainz", "MusicBrainz", nil},
		{"a mismatch against a set before", twoSources, "MusicBrainz", "Spotify", []float64{0.75}},
		{"no before and a single loaded source", oneSource, "", "Spotify", nil},
		{"no before and several loaded sources", twoSources, "", "Spotify", []float64{0.75}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.DataSourcePenalties = tt.sources
			dist := NewDistance(tun)
			addTrackDataSource(dist, ItemFields{DataSource: tt.before}, &TrackInfo{DataSource: tt.after})

			got := dist.penalties["data_source"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackDataSource(%q, %q) added %v, want %v", tt.before, tt.after, got, tt.want)
			}
		})
	}
}
