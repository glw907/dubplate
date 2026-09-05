package matcher

import (
	"slices"
	"testing"
)

func TestRecommendWalksTheLadder(t *testing.T) {
	tun := DefaultTunables() // 0.04 / 0.25 / 0.25
	tests := []struct {
		name      string
		distances []float64
		want      Recommendation
	}{
		{"no candidates", nil, RecNone},
		{"below strong", []float64{0.039}, RecStrong},
		{"exactly strong is NOT strong", []float64{0.04}, RecMedium},
		{"exactly medium IS medium", []float64{0.25}, RecMedium},
		{"single candidate above medium", []float64{0.30}, RecLow},
		{"gap at or above rec_gap_thresh", []float64{0.30, 0.55}, RecLow},
		{"gap below rec_gap_thresh", []float64{0.30, 0.54}, RecNone},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Recommend(tun, candidatesAtDistances(t, tt.distances...))
			if err != nil {
				t.Fatalf("Recommend: %v", err)
			}
			if got != tt.want {
				t.Errorf("Recommend = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRecommendReturnsBeforeTheDowngradeOnNoConclusion(t *testing.T) {
	// match.py:300-302 returns immediately, "Can't be downgraded any
	// further". A port that fell through still returns none, so the
	// only thing that can detect the difference is a max_rec that would
	// RAISE the value, which min() never does. The assertion is
	// therefore on the downgrade never being consulted: max_rec names a
	// key the candidate carries, and a strong level for it, and the
	// result stays none.
	tun := DefaultTunables()
	tun.MaxRec = map[string]string{"album": "strong"}
	cands := candidatesAtDistances(t, 0.30, 0.54)
	got, err := Recommend(tun, cands)
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if got != RecNone {
		t.Errorf("Recommend = %v, want none: the no-conclusion branch returns before the downgrade", got)
	}
	// And the guard that keeps this test honest: an unresolvable max_rec
	// level on that same key must NOT be reported, because the branch
	// returned before reading max_rec at all.
	tun.MaxRec = map[string]string{"album": "not-a-level"}
	if _, err := Recommend(tun, cands); err != nil {
		t.Errorf("Recommend = %v, want nil: the no-conclusion branch never reads max_rec", err)
	}
}

func TestRecommendDowngradesAgainstTheNonzeroViewOnly(t *testing.T) {
	// The m5 hazard's first site, and the corpus is blind to it: zero of
	// 294 committed candidates carry missing_tracks or unmatched_tracks
	// at exactly 0.0. This fixture builds one. The mutation that proves
	// it fires: read AddedKeys() instead of NonzeroKeys().
	tun := DefaultTunables() // max_rec: missing_tracks -> medium
	cand := candidateWithZeroValuedKey(t, "missing_tracks", 0.02)
	got, err := Recommend(tun, []Candidate{cand})
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if got != RecStrong {
		t.Errorf("Recommend = %v, want strong: a key added at 0.0 never downgrades (distance.py:246-247)", got)
	}
	if !slices.Contains(cand.Dist.AddedKeys(), "missing_tracks") {
		t.Fatal("the fixture no longer adds missing_tracks at 0.0, so this test can no longer discriminate the two views")
	}
}

func TestRecommendDowngradesAgainstATrackLevelKey(t *testing.T) {
	// match.py:307-309 unions every track sub-distance's keys into the
	// downgrade set. Under the shipped max_rec that loop is inert,
	// because neither missing_tracks nor unmatched_tracks is ever a
	// track-level key, and a model omitting it reproduces all 214
	// recorded recommendations. So this needs a hand-authored max_rec
	// naming a track key. The mutation: delete the union loop.
	tun := DefaultTunables()
	tun.MaxRec = map[string]string{"track_length": "low"}
	cand := candidateWithNonzeroTrackKey(t, "track_length", 0.01)
	got, err := Recommend(tun, []Candidate{cand})
	if err != nil {
		t.Fatalf("Recommend: %v", err)
	}
	if got != RecLow {
		t.Errorf("Recommend = %v, want low: max_rec's key set unions every track sub-distance's nonzero keys", got)
	}
}

func TestActionForCoversEveryBranch(t *testing.T) {
	tests := []struct {
		name          string
		quiet         bool
		quietFallback string
		timid         bool
		noneRecAction string
		rec           Recommendation
		want          Action
	}{
		{"quiet and strong applies", true, "skip", false, "ask", RecStrong, ActionApply},
		{"quiet and medium takes the fallback", true, "skip", false, "ask", RecMedium, ActionSkip},
		{"quiet fallback asis", true, "asis", false, "ask", RecNone, ActionAsIs},
		{"quiet outranks timid", true, "skip", true, "ask", RecMedium, ActionSkip},
		{"timid always asks", false, "skip", true, "skip", RecStrong, ActionAsk},
		{"none_rec_action skip", false, "skip", false, "skip", RecNone, ActionSkip},
		{"none_rec_action asis", false, "skip", false, "asis", RecNone, ActionAsIs},
		{"none_rec_action ask", false, "skip", false, "ask", RecNone, ActionAsk},
		{"not quiet, not timid, not none, asks", false, "skip", false, "skip", RecMedium, ActionAsk},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.Quiet, tun.QuietFallback, tun.Timid, tun.NoneRecAction = tt.quiet, tt.quietFallback, tt.timid, tt.noneRecAction
			got, err := ActionFor(tun, tt.rec)
			if err != nil {
				t.Fatalf("ActionFor: %v", err)
			}
			if got != tt.want {
				t.Errorf("ActionFor = %q, want %q", got, tt.want)
			}
		})
	}
}

// candidatesAtDistances returns one Candidate per given distance, in
// the given order, each scored solely under an "album" penalty whose
// weight cancels out of Distance() (RawDistance/MaxDistance), so the
// candidate's Distance() equals the requested value exactly. distances
// is passed already in the ascending order Candidates.Sorted would
// produce; Recommend never re-sorts.
func candidatesAtDistances(t *testing.T, distances ...float64) []Candidate {
	t.Helper()
	tun := DefaultTunables()
	cands := make([]Candidate, len(distances))
	for i, d := range distances {
		dist := NewDistance(tun)
		dist.Add("album", d)
		cands[i] = Candidate{
			Info: &AlbumInfo{Album: "candidate", AlbumID: "id", DataSource: "MusicBrainz"},
			Dist: dist,
		}
	}
	return cands
}

// candidateWithZeroValuedKey returns one Candidate whose Dist carries
// key added at exactly 0.0 (present in AddedKeys, absent from
// NonzeroKeys) alongside an "album" penalty sized so the candidate's
// overall Distance() equals targetDistance, regardless of key's
// configured weight.
func candidateWithZeroValuedKey(t *testing.T, key string, targetDistance float64) Candidate {
	t.Helper()
	tun := DefaultTunables()
	dist := NewDistance(tun)
	dist.Add(key, 0)
	keyWeight := tun.DistanceWeights[key]
	albumWeight := tun.DistanceWeights["album"]
	// Solve for the "album" penalty that makes
	// (albumPenalty*albumWeight) / (keyWeight + albumWeight) equal
	// targetDistance, so key's own zero-valued weight does not perturb
	// the requested overall distance.
	albumPenalty := targetDistance * (keyWeight + albumWeight) / albumWeight
	dist.Add("album", albumPenalty)
	return Candidate{
		Info: &AlbumInfo{Album: "candidate", AlbumID: "id", DataSource: "MusicBrainz"},
		Dist: dist,
	}
}

// candidateWithNonzeroTrackKey returns one Candidate whose album-level
// Dist carries only a small "album" penalty (well below
// StrongRecThresh) and whose sole track sub-distance carries key at a
// nonzero value, so the key surfaces only through the track-level
// union the downgrade must apply.
func candidateWithNonzeroTrackKey(t *testing.T, key string, trackPenalty float64) Candidate {
	t.Helper()
	tun := DefaultTunables()
	albumDist := NewDistance(tun)
	albumDist.Add("album", 0.01)
	trackDist := NewDistance(tun)
	trackDist.Add(key, trackPenalty)
	albumDist.Tracks = []TrackDistance{{Info: &TrackInfo{Title: "track"}, Dist: trackDist}}
	return Candidate{
		Info: &AlbumInfo{Album: "candidate", AlbumID: "id", DataSource: "MusicBrainz"},
		Dist: albumDist,
	}
}
