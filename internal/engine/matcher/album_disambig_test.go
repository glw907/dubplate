package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumDisambig(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely disambiguation", "", "Deluxe Edition", nil},
		{"declines when the candidate carries no disambiguation", "Deluxe Edition", "", nil},
		{"adds zero for identical disambiguation strings", "Deluxe Edition", "Deluxe Edition", []float64{0}},
		{"adds a nonzero string distance for different disambiguation strings", "Deluxe Edition", "Deluxe Reissue", []float64{0.46153846153846156}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{AlbumDisambig: tt.likely}
			info := &AlbumInfo{AlbumDisambig: tt.info}
			addAlbumDisambig(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["albumdisambig"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumDisambig(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
