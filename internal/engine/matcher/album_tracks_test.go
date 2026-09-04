package matcher

import (
	"testing"
	"time"
)

func TestAddAlbumTracks(t *testing.T) {
	tun := DefaultTunables()

	t.Run("declines (adds nothing) with zero pairs", func(t *testing.T) {
		dist := NewDistance(tun)
		info := &AlbumInfo{}
		addAlbumTracks(dist, tun, Likelies{}, info, nil, nil, time.Time{})

		if got := dist.penalties["tracks"]; len(got) != 0 {
			t.Errorf("addAlbumTracks(no pairs) added %v, want nothing", got)
		}
		if len(dist.Tracks) != 0 {
			t.Errorf("addAlbumTracks(no pairs) recorded %d sub-distances, want 0", len(dist.Tracks))
		}
	})

	t.Run("adds one normalized track distance per pair and records the sub-distance", func(t *testing.T) {
		dist := NewDistance(tun)
		info := &AlbumInfo{VA: false}
		pairs := []AlbumPair{
			{
				Item: ItemFields{Title: "Ember Wharf"},
				Info: &TrackInfo{Title: "Ember Wharf"},
			},
			{
				Item: ItemFields{Title: "A Quiet Season"},
				Info: &TrackInfo{Title: "Ember Wharf"},
			},
		}
		addAlbumTracks(dist, tun, Likelies{}, info, nil, pairs, time.Time{})

		got := dist.penalties["tracks"]
		if len(got) != 2 {
			t.Fatalf("addAlbumTracks(2 pairs) added %v, want 2 entries", got)
		}
		if got[0] != 0 {
			t.Errorf("first pair (identical) added %v, want 0", got[0])
		}
		if got[1] == 0 {
			t.Errorf("second pair (mismatched title) added 0, want a nonzero distance")
		}
		if len(dist.Tracks) != 2 {
			t.Fatalf("addAlbumTracks(2 pairs) recorded %d sub-distances, want 2", len(dist.Tracks))
		}
		if dist.Tracks[0].Info != pairs[0].Info || dist.Tracks[1].Info != pairs[1].Info {
			t.Error("addAlbumTracks recorded sub-distances against the wrong TrackInfo")
		}
	})

	// The album-level tracks component threads album_info.VA straight
	// through as track_distance's inclArtist, unnegated
	// (distance.py:529: `track_distance(item, track, album_info.va)`).
	// On a non-VA release the album-level artist component already
	// grades the artist once, so the per-track fold skips it; on a VA
	// release there is no single album artist to grade, so the
	// per-track fold is where a mismatched track artist gets caught.
	t.Run("inclArtist follows album_info.VA unnegated", func(t *testing.T) {
		pairs := []AlbumPair{
			{
				Item: ItemFields{Title: "Ember Wharf", Artist: "Quiet Season"},
				Info: &TrackInfo{Title: "Ember Wharf", Artist: "Northern Lines"},
			},
		}

		nonVA := NewDistance(tun)
		addAlbumTracks(nonVA, tun, Likelies{}, &AlbumInfo{VA: false}, nil, pairs, time.Time{})
		if got := nonVA.penalties["tracks"]; len(got) != 1 || got[0] != 0 {
			t.Errorf("non-VA release added %v, want [0] (artist not re-graded per track)", got)
		}

		va := NewDistance(tun)
		addAlbumTracks(va, tun, Likelies{}, &AlbumInfo{VA: true}, nil, pairs, time.Time{})
		if got := va.penalties["tracks"]; len(got) != 1 || got[0] == 0 {
			t.Errorf("VA release added %v, want a nonzero distance (artist graded per track)", got)
		}
	})
}
