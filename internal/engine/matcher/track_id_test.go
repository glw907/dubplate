package matcher

import (
	"slices"
	"testing"
)

func TestAddTrackID(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name   string
		itemID string
		infoID string
		want   []float64
	}{
		{"declines when the item carries no MusicBrainz track ID", "", "mbid-1", nil},
		{"matching IDs add zero", "mbid-1", "mbid-1", []float64{0}},
		{"mismatched IDs add one", "mbid-1", "mbid-2", []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackID(dist, tun, ItemFields{MBTrackID: tt.itemID}, &TrackInfo{TrackID: tt.infoID}, true)

			got := dist.penalties["track_id"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackID(%q, %q) added %v, want %v", tt.itemID, tt.infoID, got, tt.want)
			}
		})
	}
}
