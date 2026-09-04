package matcher

import (
	"slices"
	"testing"
	"time"
)

func TestAddAlbumMissingTracks(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name       string
		infoTracks int
		pairs      int
		want       []float64
	}{
		{"declines when every candidate track is paired", 2, 2, nil},
		{"declines when the pairing covers every candidate track and more", 1, 2, nil},
		{"one 1.0 penalty per unpaired candidate track", 3, 1, []float64{1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			info := &AlbumInfo{Tracks: make([]TrackInfo, tt.infoTracks)}
			pairs := make([]AlbumPair, tt.pairs)
			addAlbumMissingTracks(dist, tun, Likelies{}, info, nil, pairs, time.Time{})

			got := dist.penalties["missing_tracks"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumMissingTracks(tracks=%d, pairs=%d) added %v, want %v", tt.infoTracks, tt.pairs, got, tt.want)
			}
		})
	}
}

func TestAddAlbumUnmatchedTracks(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name  string
		items int
		pairs int
		want  []float64
	}{
		{"declines when every item is paired", 2, 2, nil},
		{"declines when the pairing covers every item and more", 1, 2, nil},
		{"one 1.0 penalty per unmatched item", 3, 1, []float64{1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			items := make([]ItemFields, tt.items)
			pairs := make([]AlbumPair, tt.pairs)
			addAlbumUnmatchedTracks(dist, tun, Likelies{}, &AlbumInfo{}, items, pairs, time.Time{})

			got := dist.penalties["unmatched_tracks"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addAlbumUnmatchedTracks(items=%d, pairs=%d) added %v, want %v", tt.items, tt.pairs, got, tt.want)
			}
		})
	}
}
