package matcher

import "testing"

// track_title carries no fire condition (contract 6: "AddString
// always"), unlike every other track component, so this test's decline
// proof is the edge case a fire-conditioned component would use to
// decline: both titles empty. track_title still adds here, which is
// the point.
func TestAddTrackTitle(t *testing.T) {
	tun := DefaultTunables()

	tests := []struct {
		name      string
		itemTitle string
		infoTitle string
		want      float64
	}{
		{"always adds, even when both titles are empty", "", "", 0},
		{"adds zero for identical titles", "Ember Wharf", "Ember Wharf", 0},
		{"adds a nonzero string distance for different titles", "Ember Wharf", "Northern Lines", 0.8461538461538461},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := NewDistance(tun)
			addTrackTitle(dist, tun, ItemFields{Title: tt.itemTitle}, &TrackInfo{Title: tt.infoTitle}, true)

			got, ok := dist.penalties["track_title"]
			if !ok {
				t.Fatalf("addTrackTitle(%q, %q) added nothing, want a track_title penalty", tt.itemTitle, tt.infoTitle)
			}
			if len(got) != 1 || got[0] != tt.want {
				t.Errorf("addTrackTitle(%q, %q) added %v, want [%v]", tt.itemTitle, tt.infoTitle, got, tt.want)
			}
		})
	}
}
