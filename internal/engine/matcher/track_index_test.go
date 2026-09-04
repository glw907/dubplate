package matcher

import (
	"slices"
	"testing"
)

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
			addTrackIndex(dist, tun, item, info, true)

			got := dist.penalties["track_index"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackIndex(track=%v, index=%v, mediumIndex=%v) added %v, want %v",
					tt.itemTrack, tt.infoIndex, tt.infoMedium, got, tt.want)
			}
		})
	}
}
