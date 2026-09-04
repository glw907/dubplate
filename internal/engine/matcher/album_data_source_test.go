package matcher

import (
	"slices"
	"testing"
	"time"
)

// TestAddAlbumDataSource walks contract 6's three-way condition in
// all four states, mirroring TestAddTrackDataSource: this component
// carries no fire condition of its own and wires likelies.DataSource
// and info.DataSource straight into AddDataSource.
func TestAddAlbumDataSource(t *testing.T) {
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
			likelies := Likelies{DataSource: tt.before}
			info := &AlbumInfo{DataSource: tt.after}
			addAlbumDataSource(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["data_source"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumDataSource(%q, %q) added %v, want %v", tt.before, tt.after, got, tt.want)
			}
		})
	}
}
