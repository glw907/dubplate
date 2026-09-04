package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumLabel(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely label", "", "Cedarwood Records", nil},
		{"declines when the candidate carries no label", "Cedarwood Records", "", nil},
		{"adds zero for identical labels", "Cedarwood Records", "Cedarwood Records", []float64{0}},
		{"adds a nonzero string distance for different labels", "Cedarwood Records", "Cedarwood Recordings", []float64{0.15789473684210525}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{Label: tt.likely}
			info := &AlbumInfo{Label: tt.info}
			addAlbumLabel(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["label"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumLabel(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
