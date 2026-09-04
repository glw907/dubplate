package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumCountry(t *testing.T) {
	tests := []struct {
		name      string
		preferred []string
		likely    string
		info      string
		want      []float64
	}{
		{"declines when neither preferred patterns nor a likely country produce a match", nil, "", "GB", nil},
		{"declines when the candidate carries no country and nothing else fires", nil, "US", "", nil},
		{"string: matching codes add zero", nil, "GB", "GB", []float64{0}},
		{"string: mismatched codes add a nonzero distance", nil, "US", "GB", []float64{1}},
		{"priority: the first pattern matches", []string{"GB", "US"}, "", "GB", []float64{0}},
		{"priority: the second pattern matches", []string{"GB", "US"}, "", "US", []float64{0.5}},
		{"priority: case-insensitive", []string{"gb"}, "", "GB", []float64{0}},
		{"priority: nothing matches", []string{"GB", "US"}, "", "FR", []float64{1}},
		{"priority takes precedence over the string fallback", []string{"GB"}, "US", "GB", []float64{0}},
		{"priority: a pattern later in the value does not match", []string{"GB"}, "", "XGB", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.Preferred.Countries = tt.preferred
			dist := NewDistance(tun)
			likelies := Likelies{Country: tt.likely}
			info := &AlbumInfo{Country: tt.info}
			addAlbumCountry(dist, tun, likelies, info, nil, nil, time.Time{})

			got := dist.penalties["country"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumCountry(preferred=%v, likely=%q, info=%q) added %v, want %v", tt.preferred, tt.likely, tt.info, got, tt.want)
			}
		})
	}
}
