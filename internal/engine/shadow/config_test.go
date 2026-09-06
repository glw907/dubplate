package shadow

import (
	"testing"

	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// TestDefaultIntervalIsTwiceMusicbrainzDefault pins the shadow's wire
// pace at twice musicbrainz.DefaultInterval: the box's own beet import
// shares the per-IP MusicBrainz budget, so the shadow paces itself
// slower rather than contending for the same budget at the same rate.
func TestDefaultIntervalIsTwiceMusicbrainzDefault(t *testing.T) {
	want := 2 * musicbrainz.DefaultInterval
	if DefaultInterval != want {
		t.Errorf("DefaultInterval = %v, want %v", DefaultInterval, want)
	}
}
