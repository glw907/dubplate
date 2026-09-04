package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumMediums(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name      string
		discTotal int
		mediums   int
		want      []float64
	}{
		{"declines when the item set carries no disc total", 0, 2, nil},
		{"declines when the candidate carries no medium count", 2, 0, nil},
		{"matching disc counts add zero", 2, 2, []float64{0}},
		{"a one-disc difference adds one unit", 1, 2, []float64{1}},
		{"a two-disc difference adds two units", 1, 3, []float64{1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			likelies := Likelies{DiscTotal: tt.discTotal}
			info := &AlbumInfo{Mediums: tt.mediums}
			addAlbumMediums(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["mediums"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumMediums(discTotal=%d, mediums=%d) added %v, want %v", tt.discTotal, tt.mediums, got, tt.want)
			}
		})
	}
}
