package matcher

import (
	"slices"
	"testing"
)

func TestAddTrackArtist(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name       string
		inclArtist bool
		itemArtist string
		infoArtist string
		want       []float64
	}{
		{"declines when inclArtist is false", false, "Northern Lines", "Ember Wharf", nil},
		{"declines when the candidate carries no artist credit", true, "Northern Lines", "", nil},
		{"declines when the item carries no artist credit, the empty first member of VA_ARTISTS", true, "", "Ember Wharf", nil},
		{"declines when the item names a various-artists placeholder", true, "Various Artists", "Ember Wharf", nil},
		{"declines case-insensitively against the VA placeholder list", true, "VARIOUS", "Ember Wharf", nil},
		{"declines case-insensitively against the mixed-case unknown placeholder", true, "Unknown", "Ember Wharf", nil},
		{"adds zero for identical artists", true, "Ember Wharf", "Ember Wharf", []float64{0}},
		{"adds a nonzero string distance for different artists", true, "Northern Lines", "Ember Wharf", []float64{0.8461538461538461}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackArtist(dist, tun, ItemFields{Artist: tt.itemArtist}, &TrackInfo{Artist: tt.infoArtist}, tt.inclArtist)

			got := dist.penalties["track_artist"]
			if !slices.Equal(got, tt.want) {
				t.Errorf("addTrackArtist(inclArtist=%v, %q, %q) added %v, want %v", tt.inclArtist, tt.itemArtist, tt.infoArtist, got, tt.want)
			}
		})
	}
}

// TestIsVariousArtistsName pins the exported seam musicbrainz's
// va_likely shares with the track artist component: one list, compared
// case-insensitively. A mutation that compares case-sensitively fails
// the "Unknown" row below, the credit the corpus's own
// nonva-va-artists-misflag case carries.
func TestIsVariousArtistsName(t *testing.T) {
	tests := []struct {
		credit string
		want   bool
	}{
		{"", true},
		{"Various Artists", true},
		{"VARIOUS", true},
		{"va", true},
		{"Unknown", true},
		{"Northern Lines", false},
	}

	for _, tt := range tests {
		t.Run(tt.credit, func(t *testing.T) {
			if got := IsVariousArtistsName(tt.credit); got != tt.want {
				t.Errorf("IsVariousArtistsName(%q) = %v, want %v", tt.credit, got, tt.want)
			}
		})
	}
}
