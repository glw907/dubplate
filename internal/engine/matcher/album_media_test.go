package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumMedia(t *testing.T) {
	tests := []struct {
		name      string
		preferred []string
		likely    string
		info      string
		want      []float64
	}{
		{"declines when the candidate carries no media", nil, "CD", "", nil},
		{"declines when neither preferred patterns nor a likely media are set", nil, "", "CD", nil},
		{"equality: current media matches", nil, "CD", "CD", []float64{0}},
		{"equality: current media mismatches", nil, "CD", "Vinyl", []float64{1}},
		{"priority: the first pattern matches", []string{"CD", "Vinyl"}, "", "CD", []float64{0}},
		{"priority: the second pattern matches", []string{"CD", "Vinyl"}, "", "Vinyl", []float64{0.5}},
		{"priority: a disc-count prefix is tolerated", []string{"CD", "Vinyl"}, "", "2xCD", []float64{0}},
		{"priority: case-insensitive", []string{"cd"}, "", "CD", []float64{0}},
		{"priority: nothing matches", []string{"CD", "Vinyl"}, "", "Cassette", []float64{1}},
		{"priority: a pattern later in the value does not match", []string{"CD"}, "", "Deluxe CD", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.Preferred.Media = tt.preferred
			dist := NewDistance(tun)
			likelies := Likelies{Media: tt.likely}
			info := &AlbumInfo{Media: tt.info}
			addAlbumMedia(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["media"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumMedia(preferred=%v, likely=%q, info=%q) added %v, want %v", tt.preferred, tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
