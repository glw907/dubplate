package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumArtist(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		va     bool
		likely string
		info   string
		want   []float64
	}{
		{"declines entirely on a various-artists release", true, "Northern Lines", "Various Artists", nil},
		{"adds zero for identical artists", false, "Northern Lines", "Northern Lines", []float64{0}},
		{"adds a nonzero string distance for different artists", false, "Northern Lines", "Ember Wharf", []float64{0.8461538461538461}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Artist: tt.likely}
			info := &AlbumInfo{VA: tt.va, Artist: tt.info}
			addAlbumArtist(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["artist"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumArtist(va=%v, %q, %q) added %v, want %v", tt.va, tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
