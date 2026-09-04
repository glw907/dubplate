package matcher

import (
	"slices"
	"testing"
)

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
			addTrackMedium(dist, tun, ItemFields{Disc: tt.itemDisc}, &TrackInfo{Medium: tt.infoMedium}, true)

			got := dist.penalties["medium"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackMedium(disc=%v, medium=%v) added %v, want %v", tt.itemDisc, tt.infoMedium, got, tt.want)
			}
		})
	}
}
