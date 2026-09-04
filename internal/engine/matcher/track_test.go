package matcher

import (
	"slices"
	"testing"
)

// TestTrackDistance hand-authors the equivalents of the corpus's 16
// track_distance rows: the data-source parameter space (contract 6's
// three-way condition in all four states) plus identical, tolerate-va,
// different-artist, and different-title. Every want was measured by
// calling beets.autotag.distance.track_distance in the pinned venv
// (tools/venv/.venv, beets 2.13.1) on exactly these inputs, with
// MusicBrainz alone loaded except where a scenario's own name says
// otherwise, and the comparison is exact rather than tolerant (contract
// 12).
func TestTrackDistance(t *testing.T) {
	base := func() (ItemFields, TrackInfo) {
		item := ItemFields{
			Title: "Ember Wharf", Artist: "Northern Lines", Length: 200,
			Track: 3, Disc: 1, MBTrackID: "mbid-1", DataSource: "MusicBrainz",
		}
		info := TrackInfo{
			Title: "Ember Wharf", Artist: "Northern Lines", Length: 200,
			Index: 3, MediumIndex: 3, Medium: 1, TrackID: "mbid-1",
			DataSource: "MusicBrainz",
		}
		return item, info
	}

	t.Run("identical", func(t *testing.T) {
		item, info := base()
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("tolerate-va: a VA solve ignores an artist mismatch", func(t *testing.T) {
		item, info := base()
		item.Artist = "Various Artists"
		dist := DistanceForTrack(DefaultTunables(), item, &info, false)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_index", "track_id", "medium")
	})

	t.Run("different-artist", func(t *testing.T) {
		item, info := base()
		item.Artist = "Quiet Season"
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0.12087912087912088)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("different-title", func(t *testing.T) {
		item, info := base()
		item.Title = "A Quiet Season"
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0.17857142857142858)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("data source: sources match", func(t *testing.T) {
		item, info := base()
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("data source: a mismatch against a set before", func(t *testing.T) {
		item, info := base()
		item.DataSource = "Spotify"
		dist := DistanceForTrack(DefaultTunables(), item, &info, true)
		wantDistance(t, dist, 0.0625)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium", "data_source")
	})

	t.Run("data source: no before and a single loaded source", func(t *testing.T) {
		item, info := base()
		item.DataSource = ""
		info.DataSource = "Spotify"
		tun := DefaultTunables()
		tun.DataSourcePenalties = map[string]float64{"MusicBrainz": 0.5}
		dist := DistanceForTrack(tun, item, &info, true)
		wantDistance(t, dist, 0)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium")
	})

	t.Run("data source: no before and several loaded sources", func(t *testing.T) {
		item, info := base()
		item.DataSource = ""
		info.DataSource = "Spotify"
		tun := DefaultTunables()
		tun.DataSourcePenalties = map[string]float64{"MusicBrainz": 0.5, "Spotify": 0.5}
		dist := DistanceForTrack(tun, item, &info, true)
		wantDistance(t, dist, 0.0625)
		wantAddedKeys(t, dist, "track_length", "track_title", "track_artist", "track_index", "track_id", "medium", "data_source")
	})
}

// TestTrackDistanceComposesInFixedOrder pins contract 8's order
// directly, so a reordering of trackComponents fails here even on
// inputs where every component adds the same value.
func TestTrackDistanceComposesInFixedOrder(t *testing.T) {
	want := []string{"track_length", "track_title", "track_artist", "track_index", "track_id", "medium", "data_source"}
	got := make([]string, len(trackComponents))
	for i, c := range trackComponents {
		got[i] = c.Name
	}
	if !slices.Equal(got, want) {
		t.Fatalf("trackComponents order = %v, want %v", got, want)
	}
}

func wantDistance(t *testing.T, dist *Distance, want float64) {
	t.Helper()
	if got := dist.Distance(); got != want {
		t.Errorf("Distance() = %v, want %v", got, want)
	}
}

func wantAddedKeys(t *testing.T, dist *Distance, want ...string) {
	t.Helper()
	if got := dist.AddedKeys(); !slices.Equal(got, want) {
		t.Fatalf("AddedKeys() = %v, want %v", got, want)
	}
}
