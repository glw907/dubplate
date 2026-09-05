package matcher

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"
)

// evalDate is a fixed evaluation date the tests in this package share,
// so a component whose only clock-shaped input is the year component's
// ratio branches gets a deterministic reference year rather than each
// test rolling its own.
var evalDate = time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)

func TestAddDropsAZeroTrackCandidateBeforeTheDuplicateCheck(t *testing.T) {
	// Gate 1 before gate 2 (match.py:345-352). A zero-track candidate
	// sharing an identity with an admitted one must be dropped by the
	// zero-track rule, so the admitted one survives untouched. A port
	// that checked duplicates first would drop it for the wrong reason
	// and report the same bool, which is why the assertion is on the
	// SURVIVING candidate rather than on the return value alone.
	c := NewCandidates()
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	good := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz",
		Tracks: []TrackInfo{{Title: "one", Index: 1, Length: 100}}}
	empty := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz"}
	if ok, err := c.Add(DefaultTunables(), items, good, evalDate); err != nil || !ok {
		t.Fatalf("Add(good) = %t, %v; want true, nil", ok, err)
	}
	if ok, err := c.Add(DefaultTunables(), items, empty, evalDate); err != nil || ok {
		t.Fatalf("Add(zero-track duplicate) = %t, %v; want false, nil", ok, err)
	}
	if c.Len() != 1 {
		t.Fatalf("Len = %d, want 1", c.Len())
	}
	if got := c.All()[0].Info; got != good {
		t.Error("the admitted candidate was replaced; the zero-track gate must run before the duplicate gate")
	}
}

func TestAddDropsADuplicateIdentity(t *testing.T) {
	c := NewCandidates()
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	tracks := []TrackInfo{{Title: "one", Index: 1, Length: 100}}
	first := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz", Tracks: tracks}
	same := &AlbumInfo{Album: "b", AlbumID: "id-1", DataSource: "MusicBrainz", Tracks: tracks}
	other := &AlbumInfo{Album: "c", AlbumID: "id-2", DataSource: "MusicBrainz", Tracks: tracks}
	noID := &AlbumInfo{Album: "d", AlbumID: "", DataSource: "MusicBrainz", Tracks: tracks}
	// A second empty-album-id candidate is admitted too (gate 2 guards
	// on info.album_id), and it OVERWRITES the first at the same
	// identity while keeping its position, exactly as Python's dict
	// assignment does at match.py:378-380. An accumulator that appends
	// to its order slice on every Add reads Len 4 here.
	noID2 := &AlbumInfo{Album: "e", AlbumID: "", DataSource: "MusicBrainz", Tracks: tracks}
	for _, tt := range []struct {
		info *AlbumInfo
		want bool
	}{{first, true}, {same, false}, {other, true}, {noID, true}, {noID2, true}} {
		if ok, err := c.Add(DefaultTunables(), items, tt.info, evalDate); err != nil || ok != tt.want {
			t.Errorf("Add(%s) = %t, %v; want %t, nil", tt.info.Album, ok, err, tt.want)
		}
	}
	if c.Len() != 3 {
		t.Errorf("Len = %d, want 3: both empty-album-id candidates are admitted (match.py:350 guards on info.album_id) and the second overwrites the first at the same identity", c.Len())
	}
	if got := c.All()[2].Info; got != noID2 {
		t.Errorf("the third slot holds %q, want the second empty-album-id candidate: a repeated identity overwrites in place", got.Album)
	}
}

func TestAddAppliesRequiredAgainstTheZeroValue(t *testing.T) {
	tun := DefaultTunables()
	tun.Required = []string{"label"}
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	tracks := []TrackInfo{{Title: "one", Index: 1, Length: 100}}
	c := NewCandidates()
	missing := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz", Tracks: tracks}
	present := &AlbumInfo{Album: "b", AlbumID: "id-2", DataSource: "MusicBrainz", Label: "Harbour", Tracks: tracks}
	if ok, _ := c.Add(tun, items, missing, evalDate); ok {
		t.Error("a candidate with no label was admitted under required = [label]")
	}
	if ok, err := c.Add(tun, items, present, evalDate); err != nil || !ok {
		t.Errorf("Add(present) = %t, %v; want true, nil", ok, err)
	}
}

func TestAddRejectsAnUnknownRequiredFieldName(t *testing.T) {
	tun := DefaultTunables()
	tun.Required = []string{"nosuchfield"}
	c := NewCandidates()
	items := []ItemFields{{Title: "one", Track: 1, Length: 100}}
	info := &AlbumInfo{Album: "a", AlbumID: "id-1", DataSource: "MusicBrainz",
		Tracks: []TrackInfo{{Title: "one", Index: 1, Length: 100}}}
	_, err := c.Add(tun, items, info, evalDate)
	if !errors.Is(err, ErrInvalidTunable) {
		t.Errorf("Add with required = [nosuchfield] returned %v, want an error wrapping ErrInvalidTunable; beets raises AttributeError here and crashes the import (match.py:357)", err)
	}
}

func TestAddAppliesIgnoredAgainstTheNonzeroView(t *testing.T) {
	// The m5 hazard's second site. The fixture is built so the candidate
	// carries "media" in the ADDED-key view at exactly 0.0 and carries
	// "album" nonzero. ignored = ["media"] must NOT drop it, because
	// beets iterates Distance.items(), which omits zero-weighted
	// penalties (distance.py:190-205, 240-241). The mutation that proves
	// this test fires: read AddedKeys() instead of NonzeroKeys().
	tun := DefaultTunables()
	tun.Ignored = []string{"media"}
	items, info := ignoredFixture(t) // testdata/candidate/ignored-zero-media.json
	c := NewCandidates()
	ok, err := c.Add(tun, items, info, evalDate)
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if !ok {
		t.Fatal("candidate dropped on a media penalty of exactly 0.0; ignored reads the nonzero view")
	}
	if !slices.Contains(c.All()[0].Dist.AddedKeys(), "media") {
		t.Fatal("the fixture no longer adds media at 0.0, so this test can no longer discriminate the two views")
	}
	if slices.Contains(c.All()[0].Dist.NonzeroKeys(), "media") {
		t.Fatal("the fixture's media penalty is nonzero, so this test can no longer discriminate the two views")
	}
}

func TestSortedIsStableOnAnExactDistanceTie(t *testing.T) {
	// Two candidates at identical distances, admitted in a known order.
	// slices.SortFunc in place of slices.SortStableFunc is the mutation,
	// and it fires here and on exactly two committed corpus cases,
	// should-ask and should-reject.
	c := candidatesWithEqualDistances(t, "first", "second", "third")
	got := make([]string, 0, 3)
	for _, cand := range c.Sorted() {
		got = append(got, cand.Info.AlbumID)
	}
	if !slices.Equal(got, []string{"first", "second", "third"}) {
		t.Errorf("Sorted = %v, want insertion order preserved across an exact tie (match.py:326-328)", got)
	}
}

// candidateFixture is the decoded shape of a testdata/candidate/*.json
// fixture: a hand-authored items and info pair, never a corpus case
// (the assertion-discipline rule for the ignored-nonzero-view gap,
// which no committed case exercises).
type candidateFixture struct {
	Items []ItemFields
	Info  AlbumInfo
}

// ignoredFixture reads testdata/candidate/ignored-zero-media.json: one
// item and one candidate built so the candidate's media penalty is
// exactly 0.0 while its album penalty is nonzero, discriminating
// NonzeroKeys from AddedKeys.
func ignoredFixture(t *testing.T) ([]ItemFields, *AlbumInfo) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", "candidate", "ignored-zero-media.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fx candidateFixture
	if err := json.Unmarshal(raw, &fx); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	if len(fx.Items) == 0 || len(fx.Info.Tracks) == 0 {
		t.Fatal("fixture is empty; want at least one item and one track")
	}
	return fx.Items, &fx.Info
}

// candidatesWithEqualDistances returns an accumulator carrying one
// candidate per id, in the given order, each scored at the same
// distance under DefaultTunables' weights: a single "album" penalty of
// 0.5 on every candidate, so Sorted's stability is the only thing that
// can distinguish their order.
func candidatesWithEqualDistances(t *testing.T, ids ...string) *Candidates {
	t.Helper()
	c := NewCandidates()
	tun := DefaultTunables()
	for _, id := range ids {
		dist := NewDistance(tun)
		dist.Add("album", 0.5)
		info := &AlbumInfo{Album: id, AlbumID: id, DataSource: "MusicBrainz"}
		key := identity{Source: info.DataSource, AlbumID: info.AlbumID}
		c.order = append(c.order, key)
		c.byID[key] = Candidate{Info: info, Dist: dist}
	}
	return c
}
