package matcher

import (
	"encoding/json"
	"maps"
	"math"
	"os"
	"slices"
	"strings"
	"testing"
)

// An aggregateFixture is one hand-authored accumulator replay: a weight
// override layered over beets' shipped defaults, a list of Add calls,
// and every aggregate and view they produce. The expected numbers are
// worked out by hand from contract 5 in testdata/aggregates.json, never
// computed from this package, so a wrong accumulator cannot agree with
// them. components.json's aggregate rows express the same arithmetic a
// second time and are Task 6's to replay.
type aggregateFixture struct {
	Name    string             `json:"name"`
	Why     string             `json:"why"`
	Weights map[string]float64 `json:"weights"`
	Calls   []penaltyCall      `json:"calls"`
	Want    aggregateWant      `json:"want"`
}

type penaltyCall struct {
	Key  string  `json:"key"`
	Dist float64 `json:"dist"`
}

type aggregateWant struct {
	RawDistance        float64            `json:"raw_distance"`
	MaxDistance        float64            `json:"max_distance"`
	Distance           float64            `json:"distance"`
	AddedKeys          []string           `json:"added_keys"`
	PerKey             map[string]float64 `json:"per_key"`
	NonzeroItems       []PenaltyItem      `json:"nonzero_items"`
	NonzeroKeys        []string           `json:"nonzero_keys"`
	GenericPenaltyKeys []string           `json:"generic_penalty_keys"`
}

// TestDistanceAggregatesAndViews replays every fixture and compares
// with exact equality rather than a tolerance, because the target is
// bit identity and every figure in the fixture is exactly
// representable (contract 12).
//
// Four mutations this test catches, each named by the fixture that
// catches it: a denominator summing configured weights instead of added
// penalties fails zero-added-key-enlarges-the-denominator; a nonzero
// view that keeps a key of weight 0.0 fails
// only-key-carries-weight-zero; a sort keyed on (distance, key) instead
// of (-distance, key) fails nonzero-view-sorts-by-descending-distance
// and equal-distances-break-on-key-ascending; and a zero-max path that
// divides anyway returns NaN and fails both zero-denominator fixtures.
func TestDistanceAggregatesAndViews(t *testing.T) {
	for _, fixture := range loadAggregateFixtures(t) {
		t.Run(fixture.Name, func(t *testing.T) {
			d := replay(fixture)

			if got := d.RawDistance(); got != fixture.Want.RawDistance {
				t.Errorf("RawDistance() = %v, want %v", got, fixture.Want.RawDistance)
			}
			if got := d.MaxDistance(); got != fixture.Want.MaxDistance {
				t.Errorf("MaxDistance() = %v, want %v", got, fixture.Want.MaxDistance)
			}
			if got := d.Distance(); got != fixture.Want.Distance {
				t.Errorf("Distance() = %v, want %v", got, fixture.Want.Distance)
			}
			if got := d.AddedKeys(); !slices.Equal(got, fixture.Want.AddedKeys) {
				t.Errorf("AddedKeys() = %v, want %v", got, fixture.Want.AddedKeys)
			}
			if got := d.PerKey(); !maps.Equal(got, fixture.Want.PerKey) {
				t.Errorf("PerKey() = %v, want %v", got, fixture.Want.PerKey)
			}
			if got := d.NonzeroItems(); !slices.Equal(got, fixture.Want.NonzeroItems) {
				t.Errorf("NonzeroItems() = %v, want %v", got, fixture.Want.NonzeroItems)
			}
			if got := d.NonzeroKeys(); !slices.Equal(got, fixture.Want.NonzeroKeys) {
				t.Errorf("NonzeroKeys() = %v, want %v", got, fixture.Want.NonzeroKeys)
			}
			if got := d.GenericPenaltyKeys(); !slices.Equal(got, fixture.Want.GenericPenaltyKeys) {
				t.Errorf("GenericPenaltyKeys() = %v, want %v", got, fixture.Want.GenericPenaltyKeys)
			}
		})
	}
}

// TestAddRejectsPenaltyOutsideRange pins beets' ValueError
// (distance.py:275-276) in the form this port gives it. A silently
// clamping Add passes every other test in this file and scores a wrong
// candidate in production, so the range test is its own case, NaN
// included: NaN compares false against both bounds in Python and must
// do the same here.
func TestAddRejectsPenaltyOutsideRange(t *testing.T) {
	tests := []struct {
		name      string
		dist      float64
		wantPanic bool
	}{
		{"lower bound is allowed", 0, false},
		{"upper bound is allowed", 1, false},
		{"just below the lower bound", -0.0001, true},
		{"just above the upper bound", 1.0001, true},
		{"not a number", math.NaN(), true},
		{"infinity", math.Inf(1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				recovered := recover()
				switch {
				case tt.wantPanic && recovered == nil:
					t.Fatalf("Add(\"album\", %v) accepted an out-of-range penalty", tt.dist)
				case !tt.wantPanic && recovered != nil:
					t.Fatalf("Add(\"album\", %v) panicked: %v", tt.dist, recovered)
				case !tt.wantPanic:
					return
				}
				message, ok := recovered.(string)
				if !ok || !strings.Contains(message, "album") || !strings.Contains(message, "[0, 1]") {
					t.Errorf("panic value %v names neither the key nor the range", recovered)
				}
			}()

			NewDistance(DefaultTunables()).Add("album", tt.dist)
		})
	}
}

// TestAddedKeysPreserveInsertionOrder is the determinism rule stated as
// a test: a port that iterates the penalty map instead of the key slice
// fails this at random, and one that sorts the keys fails it every
// time, because the keys below are added in neither alphabetical nor
// reverse order.
func TestAddedKeysPreserveInsertionOrder(t *testing.T) {
	d := NewDistance(DefaultTunables())
	want := []string{"year", "album", "medium", "country", "label", "artist"}
	for _, key := range want {
		d.Add(key, 0.5)
	}
	d.Add("medium", 0.5) // a repeat keeps its original position

	if got := d.AddedKeys(); !slices.Equal(got, want) {
		t.Errorf("AddedKeys() = %v, want %v", got, want)
	}
	if got := d.penalties["medium"]; !slices.Equal(got, []float64{0.5, 0.5}) {
		t.Errorf("penalties[medium] = %v, want two penalties", got)
	}
}

func TestAddEquality(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		options []any
		want    []float64
	}{
		{"a single scalar option that matches", "CD", []any{"CD"}, []float64{0}},
		{"a single scalar option that differs", "Vinyl", []any{"CD"}, []float64{1}},
		{"one of several options matches", "Vinyl", []any{"CD", "Vinyl"}, []float64{0}},
		{"none of several options matches", "Cassette", []any{"CD", "Vinyl"}, []float64{1}},
		{"no options at all", "CD", nil, []float64{1}},
		{"a non-string value compares by equality", 1998, []any{1998}, []float64{0}},
		{"a pattern option matches from position 0", "2xCD", []any{mustPattern(t, `(\d+x)?(CD)`, true)}, []float64{0}},
		{"a pattern option does not match later in the value", "Deluxe CD", []any{mustPattern(t, `(\d+x)?(CD)`, true)}, []float64{1}},
		{"a case-insensitive pattern option", "2xcd", []any{mustPattern(t, `(\d+x)?(CD)`, true)}, []float64{0}},
		{"a case-sensitive pattern option", "2xcd", []any{mustPattern(t, `(\d+x)?(CD)`, false)}, []float64{1}},
		{"a pattern option never matches a non-string value", 7, []any{mustPattern(t, `7`, false)}, []float64{1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDistance(DefaultTunables())
			d.AddEquality("media", tt.value, tt.options...)
			if got := d.penalties["media"]; !slices.Equal(got, tt.want) {
				t.Errorf("AddEquality(media, %v, %v) added %v, want %v", tt.value, tt.options, got, tt.want)
			}
		})
	}
}

func TestAddExpr(t *testing.T) {
	tests := []struct {
		name string
		expr bool
		want []float64
	}{
		{"a true expression is the full penalty", true, []float64{1}},
		{"a false expression still adds a penalty", false, []float64{0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDistance(DefaultTunables())
			d.AddExpr("track_index", tt.expr)
			if got := d.penalties["track_index"]; !slices.Equal(got, tt.want) {
				t.Errorf("AddExpr(track_index, %v) added %v, want %v", tt.expr, got, tt.want)
			}
		})
	}
}

func TestAddNumber(t *testing.T) {
	tests := []struct {
		name   string
		n1, n2 int
		want   []float64
	}{
		{"equal numbers add a single zero", 3, 3, []float64{0}},
		{"a difference of one", 4, 3, []float64{1}},
		{"a difference of three", 1, 4, []float64{1, 1, 1}},
		{"the argument order does not matter", 4, 1, []float64{1, 1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDistance(DefaultTunables())
			d.AddNumber("mediums", tt.n1, tt.n2)
			if got := d.penalties["mediums"]; !slices.Equal(got, tt.want) {
				t.Errorf("AddNumber(mediums, %d, %d) added %v, want %v", tt.n1, tt.n2, got, tt.want)
			}
		})
	}
}

// TestAddNumberInflatesTheDenominatorPerUnit is the consequence that
// makes AddNumber different from every other adder, and the reason
// contract 5 says the denominator counts penalties rather than keys.
// Both accumulators below score a full mediums penalty; only the one
// built from a difference of three pays for it three times over.
func TestAddNumberInflatesTheDenominatorPerUnit(t *testing.T) {
	perUnit := NewDistance(DefaultTunables())
	perUnit.AddNumber("mediums", 1, 4)

	single := NewDistance(DefaultTunables())
	single.Add("mediums", 1)

	// mediums weighs 1.0 (config_default.yaml:179).
	if got := perUnit.MaxDistance(); got != 3 {
		t.Errorf("MaxDistance() after a difference of three = %v, want 3", got)
	}
	if got := single.MaxDistance(); got != 1 {
		t.Errorf("MaxDistance() after one penalty = %v, want 1", got)
	}
	if got := perUnit.RawDistance(); got != 3 {
		t.Errorf("RawDistance() after a difference of three = %v, want 3", got)
	}
}

func TestAddPriority(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		options []any
		want    []float64
	}{
		{"the first of two options", "CD", []any{"CD", "Vinyl"}, []float64{0}},
		{"the second of two options", "Vinyl", []any{"CD", "Vinyl"}, []float64{0.5}},
		{"the second of four options", "B", []any{"A", "B", "C", "D"}, []float64{0.25}},
		{"the last of four options", "D", []any{"A", "B", "C", "D"}, []float64{0.75}},
		{"no option matches", "X", []any{"A", "B"}, []float64{1}},
		{"a single scalar option makes the unit 1.0", "A", []any{"A"}, []float64{0}},
		{"a single scalar option that differs", "B", []any{"A"}, []float64{1}},
		{"an empty option list divides by one", "A", nil, []float64{1}},
		{"a pattern option holds its position", "GB", []any{"US", mustPattern(t, `G.`, false)}, []float64{0.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDistance(DefaultTunables())
			d.AddPriority("country", tt.value, tt.options...)
			if got := d.penalties["country"]; !slices.Equal(got, tt.want) {
				t.Errorf("AddPriority(country, %v, %v) added %v, want %v", tt.value, tt.options, got, tt.want)
			}
		})
	}
}

func TestAddRatio(t *testing.T) {
	tests := []struct {
		name   string
		n1, n2 float64
		want   []float64
	}{
		{"a plain ratio", 15, 30, []float64{0.5}},
		{"a numerator at the denominator", 30, 30, []float64{1}},
		{"a numerator above the denominator clamps to one", 45, 30, []float64{1}},
		{"a negative numerator clamps to zero", -5, 30, []float64{0}},
		{"a zero denominator adds zero rather than dividing", 20, 0, []float64{0}},
		{"a zero numerator", 0, 30, []float64{0}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDistance(DefaultTunables())
			d.AddRatio("track_length", tt.n1, tt.n2)
			if got := d.penalties["track_length"]; !slices.Equal(got, tt.want) {
				t.Errorf("AddRatio(track_length, %v, %v) added %v, want %v", tt.n1, tt.n2, got, tt.want)
			}
		})
	}
}

// TestAddDataSource walks the three-way condition in all four states.
// The two states that add nothing are the load-bearing ones: adding a
// 0.0 instead would leave data_source in the denominator and shift
// every other key's normalized value, which is a divergence no test of
// the data_source key alone would show.
func TestAddDataSource(t *testing.T) {
	twoSources := map[string]float64{"Original": 0.5, "Other": 0.75}
	oneSource := map[string]float64{"Other": 0.75}

	tests := []struct {
		name    string
		sources map[string]float64
		before  string
		after   string
		want    []float64
	}{
		{"the sources match", twoSources, "Other", "Other", nil},
		{"both sources are unset", twoSources, "", "", nil},
		{"a mismatch against a set before", twoSources, "Original", "Other", []float64{0.75}},
		{"no before and a single loaded source", oneSource, "", "Other", nil},
		{"no before and several loaded sources", twoSources, "", "Other", []float64{0.75}},
		{"an after no loaded source claims", twoSources, "Original", "Unknown", []float64{0.5}},
		{"an unset after", twoSources, "Original", "", []float64{0.5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.DataSourcePenalties = tt.sources
			d := NewDistance(tun)
			d.AddDataSource(tt.before, tt.after)

			if got := d.penalties["data_source"]; !slices.Equal(got, tt.want) {
				t.Fatalf("AddDataSource(%q, %q) added %v, want %v", tt.before, tt.after, got, tt.want)
			}
			if tt.want != nil {
				return
			}
			if got := d.AddedKeys(); len(got) != 0 {
				t.Errorf("AddedKeys() = %v, want none: an unadded key must stay out of the denominator", got)
			}
			if got := d.MaxDistance(); got != 0 {
				t.Errorf("MaxDistance() = %v, want 0", got)
			}
		})
	}
}

// TestGenericPenaltyKeys pins the three replacements and their order.
// Turning underscores into spaces before stripping the prefixes reads
// "album_id" as "album id"; dropping the nonzero test lets
// data_source, added here at 0.0, into the list.
func TestGenericPenaltyKeys(t *testing.T) {
	d := NewDistance(DefaultTunables())
	d.Add("album_id", 1)
	d.Add("track_title", 1)
	d.Add("missing_tracks", 1)
	d.Add("data_source", 0)
	d.Add("album", 1)

	want := []string{"id", "title", "missing tracks", "album"}
	if got := d.GenericPenaltyKeys(); !slices.Equal(got, want) {
		t.Errorf("GenericPenaltyKeys() = %v, want %v", got, want)
	}
}

// TestUpdate folds one accumulator into another, which is how the
// album-level tracks component consumes a track distance. A shared key
// keeps the position it already had and collects both penalties, so the
// denominator counts three penalties rather than two keys.
func TestUpdate(t *testing.T) {
	album := NewDistance(DefaultTunables())
	album.Add("album", 1)

	track := NewDistance(DefaultTunables())
	track.Add("track_title", 0.5)
	track.Add("album", 0.25)

	album.Update(track)

	wantKeys := []string{"album", "track_title"}
	if got := album.AddedKeys(); !slices.Equal(got, wantKeys) {
		t.Errorf("AddedKeys() = %v, want %v", got, wantKeys)
	}
	if got := album.penalties["album"]; !slices.Equal(got, []float64{1, 0.25}) {
		t.Errorf("penalties[album] = %v, want [1 0.25]", got)
	}
	if got := album.penalties["track_title"]; !slices.Equal(got, []float64{0.5}) {
		t.Errorf("penalties[track_title] = %v, want [0.5]", got)
	}
	// album and track_title both weigh 3.0, over three penalties.
	if got := album.MaxDistance(); got != 9 {
		t.Errorf("MaxDistance() = %v, want 9", got)
	}
}

// TestTracksKeepIdenticalCandidatesDistinct is contract m13's proof.
// The two candidate tracks below are equal as values and are different
// candidates, so a map[TrackInfo]*Distance would keep one entry and
// silently discard the other's score.
func TestTracksKeepIdenticalCandidatesDistinct(t *testing.T) {
	first := &TrackInfo{Title: "Ember Wharf", Index: 1}
	second := &TrackInfo{Title: "Ember Wharf", Index: 1}
	if *first != *second {
		t.Fatal("the two candidates must be equal as values for this test to mean anything")
	}

	scored := NewDistance(DefaultTunables())
	scored.Add("track_title", 1)
	unscored := NewDistance(DefaultTunables())
	unscored.Add("track_title", 0)

	d := NewDistance(DefaultTunables())
	d.Tracks = append(d.Tracks, TrackDistance{Info: first, Dist: unscored}, TrackDistance{Info: second, Dist: scored})

	if len(d.Tracks) != 2 {
		t.Fatalf("len(Tracks) = %d, want 2", len(d.Tracks))
	}
	if d.Tracks[0].Info != first || d.Tracks[1].Info != second {
		t.Error("Tracks lost the pairing order or the candidate identities")
	}
	if got := d.Tracks[0].Dist.Distance(); got != 0 {
		t.Errorf("Tracks[0] distance = %v, want 0", got)
	}
	if got := d.Tracks[1].Dist.Distance(); got != 1 {
		t.Errorf("Tracks[1] distance = %v, want 1", got)
	}
}

// TestUnweightedKeyIsAddableAndUnaggregatable pins both halves of the
// weight lookup. The adder tables of components.json add under keys
// beets does not weight, so Add must accept one; reading an aggregate
// over such a key is a port bug, and beets raises a KeyError out of the
// same lookup rather than treating the key as weightless.
func TestUnweightedKeyIsAddableAndUnaggregatable(t *testing.T) {
	d := NewDistance(DefaultTunables())
	d.Add("equality", 0.5)

	if got := d.penalties["equality"]; !slices.Equal(got, []float64{0.5}) {
		t.Fatalf("penalties[equality] = %v, want [0.5]", got)
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("MaxDistance() over an unweighted key did not panic")
		}
		if message, ok := recovered.(string); !ok || !strings.Contains(message, "equality") {
			t.Errorf("panic value %v does not name the key", recovered)
		}
	}()

	d.MaxDistance()
}

func TestCompilePatternOptionRejectsABadPattern(t *testing.T) {
	if _, err := CompilePatternOption(`(unclosed`, false); err == nil {
		t.Error("CompilePatternOption accepted an unparsable pattern")
	}
}

func loadAggregateFixtures(t *testing.T) []aggregateFixture {
	t.Helper()
	raw, err := os.ReadFile("testdata/aggregates.json")
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var fixtures []aggregateFixture
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatalf("decode fixtures: %v", err)
	}
	if len(fixtures) == 0 {
		t.Fatal("testdata/aggregates.json carries no fixtures")
	}
	return fixtures
}

// replay layers the fixture's weight overrides onto beets' shipped
// defaults, which is what components.json's rows do
// (derive_component.py's Distance._weights.update), so a fixture names
// only the weights it cares about.
func replay(fixture aggregateFixture) *Distance {
	tun := DefaultTunables()
	maps.Copy(tun.DistanceWeights, fixture.Weights)

	d := NewDistance(tun)
	for _, call := range fixture.Calls {
		d.Add(call.Key, call.Dist)
	}
	return d
}

func mustPattern(t *testing.T, pattern string, ignoreCase bool) PatternOption {
	t.Helper()
	option, err := CompilePatternOption(pattern, ignoreCase)
	if err != nil {
		t.Fatalf("CompilePatternOption(%q, %v): %v", pattern, ignoreCase, err)
	}
	return option
}
