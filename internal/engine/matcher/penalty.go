package matcher

import (
	"cmp"
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// A Distance accumulates the weighted penalties one candidate earns and
// normalizes them into a single score, reproducing beets' Distance
// (autotag/distance.py:122-361).
//
// Penalty keys are held in insertion order beside the map, because
// RawDistance and MaxDistance sum over that order and Go's map iteration
// is randomized: beets' dict preserves insertion order for free and the
// port must not lose it (contract 5, and the pass's determinism rule).
// Float addition is not associative, so a different summation order is a
// different number.
//
// A Distance is not safe for concurrent use, and every scoring path in
// this package is sequential.
type Distance struct {
	tun       Tunables
	penalties map[string][]float64
	order     []string

	// Tracks holds the per-track sub-distances the album-level tracks
	// component fills in, in pairing order.
	//
	// Python keys the same structure by object identity
	// (dist.tracks: dict[TrackInfo, Distance], distance.py:131), which
	// works only because TrackInfo defines no __hash__ of its own. Go
	// has no value equivalent: TrackInfo is a struct, so a
	// map[TrackInfo]*Distance would key by value and silently merge two
	// candidate tracks that happen to carry identical metadata. A
	// value-keyed map is wrong here. This slice keys by pairing
	// position instead, which is both identity-preserving and
	// deterministic to iterate.
	Tracks []TrackDistance
}

// A TrackDistance pairs one candidate track with the distance
// accumulated for it.
type TrackDistance struct {
	Info *TrackInfo
	Dist *Distance
}

// A PenaltyItem is one penalty key and its weighted, normalized
// distance, the Go form of one (key, dist) pair from beets'
// Distance.items (distance.py:190-205).
type PenaltyItem struct {
	Key      string
	Distance float64
}

// NewDistance returns an empty accumulator that scores against tun's
// distance weights and, for AddDataSource, tun's loaded metadata source
// set.
func NewDistance(tun Tunables) *Distance {
	return &Distance{tun: tun, penalties: make(map[string][]float64)}
}

// RawDistance returns the denormalized distance: for every added key,
// the sum of its penalties times the key's weight (contract 5).
func (d *Distance) RawDistance() float64 {
	var raw float64
	for _, key := range d.order {
		raw += sumPenalties(d.penalties[key]) * d.weight(key)
	}
	return raw
}

// MaxDistance returns the normalization factor: for every added key,
// the number of penalties added under it times the key's weight
// (contract 5). The count is of penalties actually added, not of
// configured weights, so a key added at 0.0 still enlarges the result
// and a key never added contributes nothing. AddNumber appends one
// penalty per unit of difference, so one key can enlarge this by its
// weight several times over.
func (d *Distance) MaxDistance() float64 {
	var maxDist float64
	for _, key := range d.order {
		maxDist += float64(len(d.penalties[key])) * d.weight(key)
	}
	return maxDist
}

// Distance returns the weighted, normalized distance across every added
// penalty, and 0.0 when MaxDistance is zero (contract 5). The zero-max
// case is beets' own guard (distance.py:157-160) and is reachable: an
// accumulator with no penalties has one, and so does one whose every
// added key carries weight 0.0.
func (d *Distance) Distance() float64 {
	maxDist := d.MaxDistance()
	if maxDist == 0 {
		return 0
	}
	return d.RawDistance() / maxDist
}

// Penalties returns key's individual added penalties, in append order,
// unweighted and unnormalized: beets' dist._penalties[key]
// (distance.py:131). It is the one raw view Distance exposes; every
// other accessor already weights, normalizes, or aggregates. Tier 0's
// decoder is the caller that needs it: an adder-shape components.json
// row pins the exact append-order slice a fresh accumulator produces,
// and a track_distance or add_data_source row pins the same slice per
// key rather than a weighted value. A key never added returns nil.
func (d *Distance) Penalties(key string) []float64 {
	return slices.Clone(d.penalties[key])
}

// AddedKeys returns every penalty key in insertion order, including
// keys added at 0.0.
//
// This is the _penalties view, the one derive_oracle.py emits and
// cassette.Candidate.Penalties grades, so key presence at 0.0 is part
// of what it reports. NonzeroKeys is the other view and drops those
// keys; the two are different objects and a consumer that picks the
// wrong one is port hazard 4.
func (d *Distance) AddedKeys() []string {
	return slices.Clone(d.order)
}

// PerKey returns the weighted, normalized distance of every added key,
// including keys added at 0.0, which map to 0.0 rather than being
// dropped. Use AddedKeys for the insertion order this map cannot carry.
//
// This is the _penalties view; see AddedKeys.
func (d *Distance) PerKey() map[string]float64 {
	perKey := make(map[string]float64, len(d.order))
	for _, key := range d.order {
		perKey[key] = d.keyDistance(key)
	}
	return perKey
}

// NonzeroItems returns every key whose weighted, normalized distance is
// nonzero, sorted by (-distance, key): biggest distance first, ties
// broken by key ascending (distance.py:190-205).
//
// This is the items view, the one beets' max_rec downgrade reads
// through keys(). It omits keys added at 0.0 and keys whose weight is
// 0.0, so a zero penalty never triggers a downgrade. AddedKeys and
// PerKey are the other view and keep those keys; grading against one
// and downgrading against the other is port hazard 4.
func (d *Distance) NonzeroItems() []PenaltyItem {
	var items []PenaltyItem
	for _, key := range d.order {
		if dist := d.keyDistance(key); dist != 0 {
			items = append(items, PenaltyItem{Key: key, Distance: dist})
		}
	}
	slices.SortFunc(items, func(a, b PenaltyItem) int {
		return cmp.Or(cmp.Compare(-a.Distance, -b.Distance), cmp.Compare(a.Key, b.Key))
	})
	return items
}

// NonzeroKeys returns the keys of NonzeroItems, in the same order.
//
// This is the items view; see NonzeroItems.
func (d *Distance) NonzeroKeys() []string {
	items := d.NonzeroItems()
	keys := make([]string, len(items))
	for i, item := range items {
		keys[i] = item.Key
	}
	return keys
}

// GenericPenaltyKeys returns the human-facing name of every key whose
// weighted distance is nonzero, in insertion order rather than sorted
// (distance.py:142-148).
//
// A name is the key with "album_" removed, then "track_" removed, then
// every remaining underscore turned into a space, in that order: run
// the underscore step first and "album_id" reads "album id" instead of
// "id". The nonzero test is the same one NonzeroItems applies, so a key
// added at 0.0 never appears.
func (d *Distance) GenericPenaltyKeys() []string {
	var names []string
	for _, key := range d.order {
		if d.keyDistance(key) == 0 {
			continue
		}
		name := strings.ReplaceAll(key, "album_", "")
		name = strings.ReplaceAll(name, "track_", "")
		names = append(names, strings.ReplaceAll(name, "_", " "))
	}
	return names
}

// Update appends every penalty accumulated in other under other's own
// keys, in other's insertion order, which is how the album-level
// components fold a track distance into the album distance
// (distance.py:249-256). Keys already present keep their position;
// new ones are appended.
func (d *Distance) Update(other *Distance) {
	for _, key := range other.order {
		for _, penalty := range other.penalties[key] {
			d.Add(key, penalty)
		}
	}
}

// Add appends one penalty under key.
//
// dist must lie in [0, 1]; anything else, NaN included, is a
// programming error in a component and panics naming the key and the
// value, which is beets' own ValueError (distance.py:275-276) in the
// form Go reserves for a broken invariant. A clamped penalty would
// score a wrong candidate quietly, so Add never clamps. The range test
// is written as the negation of "inside the range" so that NaN fails
// it, which "below zero or above one" would not.
//
// key needs a configured weight only when an aggregate or a view is
// read; the adder tables of components.json use keys beets does not
// weight, and adding under one of those is fine on its own.
func (d *Distance) Add(key string, dist float64) {
	if !(dist >= 0 && dist <= 1) {
		panic(fmt.Sprintf("matcher: penalty %v for key %q is outside [0, 1]", dist, key))
	}
	if _, seen := d.penalties[key]; !seen {
		d.order = append(d.order, key)
	}
	d.penalties[key] = append(d.penalties[key], dist)
}

// AddEquality adds 0.0 when value equals any of options and 1.0 when
// none of them matches (distance.py:279-295).
//
// beets wraps a bare scalar option into a one-element list
// (distance.py:287-288), and distance() relies on that wrap, passing a
// scalar at :465 and :522-524. The variadic form is that wrap, so one
// option and a spread slice both work, and no options at all scores
// 1.0.
//
// An option of type PatternOption matches value from position 0 instead
// of comparing equal to it; every other option is compared with Go's
// own equality.
func (d *Distance) AddEquality(key string, value any, options ...any) {
	for _, option := range options {
		if optionMatches(option, value) {
			d.Add(key, 0)
			return
		}
	}
	d.Add(key, 1)
}

// AddExpr adds 1.0 when expr holds and 0.0 when it does not
// (distance.py:297-304).
func (d *Distance) AddExpr(key string, expr bool) {
	dist := 0.0
	if expr {
		dist = 1
	}
	d.Add(key, dist)
}

// AddNumber adds one 1.0 penalty for each unit of difference between
// n1 and n2, or a single 0.0 penalty when they are equal
// (distance.py:306-317).
//
// The several penalties are the point: each one enlarges MaxDistance by
// the key's weight, so a difference of three costs three times the
// denominator a difference of one does. Use this only where the
// difference has no upper bound.
func (d *Distance) AddNumber(key string, n1, n2 int) {
	diff := max(n1-n2, n2-n1)
	if diff == 0 {
		d.Add(key, 0)
		return
	}
	for range diff {
		d.Add(key, 1)
	}
}

// AddPriority adds a penalty for the position at which value appears in
// options: 0.0 for the first, and i/len(options) for the one at index i
// (distance.py:319-337). Nothing matching scores 1.0.
//
// As with AddEquality the variadic form is beets' scalar wrap
// (distance.py:328-329), and the unit is computed after that wrap, so a
// single option makes the unit 1.0. The guard against an empty option
// list is beets' own "len(options) or 1" and is unobservable in either
// language, because the unit is read only inside the loop that an empty
// list never enters. It is kept for parity with the source.
//
// An option of type PatternOption matches value from position 0; see
// AddEquality.
func (d *Distance) AddPriority(key string, value any, options ...any) {
	unit := 1.0 / float64(max(len(options), 1))
	for i, option := range options {
		if optionMatches(option, value) {
			d.Add(key, float64(i)*unit)
			return
		}
	}
	d.Add(key, 1)
}

// AddRatio adds n1 as a ratio of n2, with n1 bound below at 0 and above
// at n2, and adds 0.0 when n2 is zero (distance.py:339-348).
//
// The clamp is what keeps the ratio inside Add's range, and the
// zero-denominator rule is beets' falsy test on n2 rather than a guard
// against dividing by zero: a zero n2 scores no penalty, not an
// infinite one.
func (d *Distance) AddRatio(key string, n1, n2 float64) {
	if n2 == 0 {
		d.Add(key, 0)
		return
	}
	d.Add(key, max(min(n1, n2), 0)/n2)
}

// AddDataSource adds the data_source penalty for a candidate that came
// from a different metadata source than the item already carries
// (distance.py:357-361).
//
// It adds nothing at all, not even 0.0, in two cases: when before and
// after agree, and when the item carries no source and only one
// metadata source is loaded, which is what a fresh single-source import
// looks like. Adding nothing keeps data_source out of the denominator
// entirely, so the distinction matters to every other key's normalized
// value.
//
// An unset source is the empty string on either side. The penalty is
// the loaded source's own, taken from Tunables.DataSourcePenalties, and
// Tunables.DataSourceMismatchPenalty when no loaded source is named
// after, which is beets' get_penalty (metadata_plugins.py:160-169).
func (d *Distance) AddDataSource(before, after string) {
	if before == after {
		return
	}
	if before == "" && len(d.tun.DataSourcePenalties) <= 1 {
		return
	}
	d.Add("data_source", d.dataSourcePenalty(after))
}

// AddString, the eighth adder, lands beside StringDist in strdist.go.
// Its whole body is one call into that function, Go has no forward
// declaration, and a placeholder standing in for string_dist would be
// the silent-success fake the assertion-discipline rule forbids.

// A PatternOption is an AddEquality or AddPriority option that matches
// its candidate value rather than comparing equal to it, the Go form of
// beets' compiled-pattern option (distance.py:260-267).
//
// The match is anchored at position 0, because Python's
// re.Pattern.match is. Options anchor and SD_PATTERNS never do
// (contract 7), so this type exists rather than a bare *regexp.Regexp:
// an unanchored MatchString on a pattern meant for .match accepts
// values beets rejects. A value that is not a string never matches.
type PatternOption struct {
	re *regexp.Regexp
}

// CompilePatternOption compiles pattern into an option anchored at
// position 0. ignoreCase is Python's re.I and compiles as Go's (?i).
func CompilePatternOption(pattern string, ignoreCase bool) (PatternOption, error) {
	anchored := "^(?:" + pattern + ")"
	if ignoreCase {
		anchored = "(?i)" + anchored
	}
	re, err := regexp.Compile(anchored)
	if err != nil {
		return PatternOption{}, fmt.Errorf("compile option pattern %q: %w", pattern, err)
	}
	return PatternOption{re: re}, nil
}

func (p PatternOption) matches(value any) bool {
	s, ok := value.(string)
	return ok && p.re.MatchString(s)
}

func optionMatches(option, value any) bool {
	if pattern, ok := option.(PatternOption); ok {
		return pattern.matches(value)
	}
	return option == value
}

// keyDistance is beets' Distance.__getitem__ (distance.py:232-238): the
// key's summed penalties times its weight, normalized by MaxDistance,
// and 0.0 when that factor is zero.
func (d *Distance) keyDistance(key string) float64 {
	dist := sumPenalties(d.penalties[key]) * d.weight(key)
	maxDist := d.MaxDistance()
	if maxDist == 0 {
		return 0
	}
	return dist / maxDist
}

func (d *Distance) dataSourcePenalty(source string) float64 {
	if penalty, loaded := d.tun.DataSourcePenalties[source]; loaded {
		return penalty
	}
	return d.tun.DataSourceMismatchPenalty
}

// weight returns key's configured weight and panics when the tunables
// carry none, which is the KeyError beets raises out of the same lookup
// (distance.py:167). Treating an unweighted key as weight 0.0 instead
// would drop it from both the numerator and the denominator without
// saying so.
func (d *Distance) weight(key string) float64 {
	w, configured := d.tun.DistanceWeights[key]
	if !configured {
		panic(fmt.Sprintf("matcher: no distance weight configured for key %q", key))
	}
	return w
}

func sumPenalties(penalties []float64) float64 {
	var total float64
	for _, penalty := range penalties {
		total += penalty
	}
	return total
}
