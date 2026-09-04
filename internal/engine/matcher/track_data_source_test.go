package matcher

import (
	"slices"
	"testing"
)

// TestAddTrackDataSource walks contract 6's three-way condition in all
// four states, the same states TestAddDataSource proves at the
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
			addTrackDataSource(dist, tun, ItemFields{DataSource: tt.before}, &TrackInfo{DataSource: tt.after}, true)

			got := dist.penalties["data_source"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackDataSource(%q, %q) added %v, want %v", tt.before, tt.after, got, tt.want)
			}
		})
	}
}
