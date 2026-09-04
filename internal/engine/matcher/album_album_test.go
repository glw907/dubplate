package matcher

import (
	"slices"
	"testing"
	"time"
)

// album carries no fire condition (contract 8: "AddString always"),
// unlike most album components, so this test's decline proof is the
// edge case a fire-conditioned component would use to decline: both
// titles empty. addAlbumAlbum still adds here, which is the point.
func TestAddAlbumAlbum(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"always adds, even when both titles are empty", "", "", []float64{0}},
		{"adds zero for identical titles", "Harbour Sound", "Harbour Sound", []float64{0}},
		{"adds a nonzero string distance for different titles", "Ember Wharf", "Northern Lines", []float64{0.8461538461538461}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Album: tt.likely}
			info := &AlbumInfo{Album: tt.info}
			addAlbumAlbum(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["album"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumAlbum(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
