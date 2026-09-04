package matcher

import (
	"slices"
	"testing"
)

// The three arithmetic shapes AddRatio's own contract already pins
// (contract 6): a negative diff clamps to 0, a diff beyond
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
			addTrackLength(dist, tun, ItemFields{Length: tt.itemLength}, &TrackInfo{Length: tt.infoLength}, true)

			got := dist.penalties["track_length"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackLength(%v, %v) added %v, want %v", tt.itemLength, tt.infoLength, got, tt.want)
			}
		})
	}
}
