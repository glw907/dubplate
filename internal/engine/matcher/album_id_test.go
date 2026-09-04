package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumID(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		likely string
		info   string
		want   []float64
	}{
		{"declines when the item set carries no likely album ID", "", "mb-1", nil},
		{"matching IDs add zero", "mb-1", "mb-1", []float64{0}},
		{"mismatched IDs add one", "mb-1", "mb-2", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{MBAlbumID: tt.likely}
			info := &AlbumInfo{AlbumID: tt.info}
			addAlbumID(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["album_id"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumID(%q, %q) added %v, want %v", tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
