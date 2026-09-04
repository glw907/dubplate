package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumCatalogNum(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely catalog number", "", "CAT-001", nil},
		{"declines when the candidate carries no catalog number", "CAT-001", "", nil},
		{"adds zero for identical catalog numbers", "CAT-001", "CAT-001", []float64{0}},
		{"adds a nonzero string distance for different catalog numbers", "CAT-001", "CAT-002", []float64{0.16666666666666666}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{CatalogNum: tt.likely}
			info := &AlbumInfo{CatalogNum: tt.info}
			addAlbumCatalogNum(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["catalognum"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumCatalogNum(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
