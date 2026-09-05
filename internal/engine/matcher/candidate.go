package matcher

import (
	"cmp"
	"fmt"
	"slices"
	"time"
)

// A Candidate is one admitted release with its solved pairing and its
// accumulated distance: beets' AlbumMatch (autotag/match.py:142-201).
type Candidate struct {
	Info       *AlbumInfo
	Dist       *Distance
	Assignment Assignment
}

// Candidates accumulates admitted candidates in insertion order under
// beets' own (data_source, album_id) identity, reproducing tag_album's
// results dict (autotag/match.py:414, hooks.py:166-168).
//
// It is a mutable accumulator rather than a pure function, and it earns
// that: the ID stage and the search stage add into the SAME set, so the
// duplicate rule spans both, tag_album reads the set's length between
// them (match.py:432), and the insertion order the stable sort falls
// back on is exactly what a rebuilt slice would lose. Add performs no
// I/O, so this type stays in matcher and matcher stays importable by
// internal/engine/musicbrainz.
type Candidates struct {
	order []identity
	byID  map[identity]Candidate
}

// identity is beets' Info.identifier: the cross-provider key
// (data_source, id) (hooks.py:166-168), matching cassette.Candidate's
// own Source and AlbumID pair (cassette.go:103-109).
type identity struct {
	Source  string
	AlbumID string
}

// NewCandidates returns an empty accumulator.
func NewCandidates() *Candidates {
	return &Candidates{byID: make(map[identity]Candidate)}
}

// requiredFields resolves match.required's accepted names to a check
// of the corresponding [AlbumInfo] field's Go zero value, resolving
// Resolution R6. It carries exactly the thirteen scalar AlbumInfo
// field names; Tracks is excluded on purpose, because gate 1 already
// drops a zero-track candidate before this gate runs, so naming it
// here could never change an outcome.
//
// beets tests getattr(info, req_tag) is None, which distinguishes an
// absent or None field from an empty one, and beets accepts any
// attribute name, raising AttributeError on an unknown one, crashing
// the import mid-run (match.py:355-359). Go has no None to test: every
// field here is checked against its own zero value instead. For va,
// this is a real behavioral divergence rather than a modelling choice:
// beets declares va: bool = False and never None (hooks.py:351,381),
// so required: [va] filters nothing in beets and would drop every
// non-VA release here. The three int fields (year, original_year,
// mediums) carry a weaker form of the same, since beets can hold None
// there; 3b's guard 2 settled that no rung 3 source produces
// an empty-but-present field, because internal/engine/musicbrainz is
// the only translation source and every field it writes comes from a
// Go decode that yields "" rather than a null. That condition reopens
// the moment a second metadata source arrives, and the translation
// tier's three nullStrict fields would catch it.
var requiredFields = map[string]func(*AlbumInfo) bool{
	"artist":        func(i *AlbumInfo) bool { return i.Artist == "" },
	"album":         func(i *AlbumInfo) bool { return i.Album == "" },
	"va":            func(i *AlbumInfo) bool { return !i.VA },
	"media":         func(i *AlbumInfo) bool { return i.Media == "" },
	"mediums":       func(i *AlbumInfo) bool { return i.Mediums == 0 },
	"year":          func(i *AlbumInfo) bool { return i.Year == 0 },
	"original_year": func(i *AlbumInfo) bool { return i.OriginalYear == 0 },
	"country":       func(i *AlbumInfo) bool { return i.Country == "" },
	"label":         func(i *AlbumInfo) bool { return i.Label == "" },
	"catalognum":    func(i *AlbumInfo) bool { return i.CatalogNum == "" },
	"albumdisambig": func(i *AlbumInfo) bool { return i.AlbumDisambig == "" },
	"album_id":      func(i *AlbumInfo) bool { return i.AlbumID == "" },
	"data_source":   func(i *AlbumInfo) bool { return i.DataSource == "" },
}

// Add scores info against items and admits it, reporting whether it was
// admitted. A drop is not an error: beets logs at debug and returns
// (match.py:345-375). It returns an error only for a malformed Tunables
// value Validate would also have caught.
//
// The five gates run in beets' fixed order (contract V7). Gate 1 (zero
// tracks) runs before gate 2 (duplicate) because match.py:345-352 runs
// them in that order; the swap is NOT observable through Add on a
// zero-track duplicate, since both gates return (false, nil) with no
// side effect, so this ordering is held by the beets citation rather
// than by a test. Reordering gate 3 (required) and gate 4 (the solve)
// does change something observable: cost, since a candidate rejected by
// required has not yet paid for a solve while one rejected by ignored
// (gate 5) has.
//
// An admitted candidate whose identity is already present OVERWRITES
// the stored value and keeps the earlier insertion position, matching
// Python's own dict assignment at match.py:378-380. It is reachable
// only through an empty album id, because gate 2 drops a duplicate with
// a non-empty one (match.py:350), and it is why order is appended to
// only when the identity is new: appending on every Add would
// double-count the position and desynchronize order from byID. No
// committed corpus case reaches it.
func (c *Candidates) Add(tun Tunables, items []ItemFields, info *AlbumInfo, evalDate time.Time) (bool, error) {
	if len(info.Tracks) == 0 { // gate 1: zero-track drop (match.py:345-347)
		return false, nil
	}

	id := identity{Source: info.DataSource, AlbumID: info.AlbumID}
	if info.AlbumID != "" { // gate 2: duplicate drop (match.py:350-352, contract V10)
		if _, seen := c.byID[id]; seen {
			return false, nil
		}
	}

	for _, name := range tun.Required { // gate 3: match.required (match.py:355-359)
		isZero, ok := requiredFields[name]
		if !ok {
			return false, fmt.Errorf("required: %q: %w", name, ErrInvalidTunable)
		}
		if isZero(info) {
			return false, nil
		}
	}

	assignment := AssignItems(tun, items, info.Tracks)
	pairs := AlbumPairsFor(assignment, items, info)
	dist := DistanceForAlbum(items, info, pairs, tun, evalDate) // gate 4: score

	// gate 5: match.ignored reads the NONZERO view (contract V8). beets
	// iterates Distance.items(), which omits zero-weighted penalties
	// (distance.py:190-205, 240-241); AddedKeys and PerKey are the
	// other view, and tier 1 cannot catch a wrong choice here because
	// match.ignored is corpus-unreachable.
	nonzero := dist.NonzeroKeys()
	for _, name := range tun.Ignored {
		if slices.Contains(nonzero, name) {
			return false, nil
		}
	}

	if _, seen := c.byID[id]; !seen {
		c.order = append(c.order, id)
	}
	c.byID[id] = Candidate{Info: info, Dist: dist, Assignment: assignment}
	return true, nil
}

// Len reports how many candidates have been admitted.
func (c *Candidates) Len() int {
	return len(c.order)
}

// All returns the admitted candidates in insertion order.
func (c *Candidates) All() []Candidate {
	out := make([]Candidate, len(c.order))
	for i, id := range c.order {
		out[i] = c.byID[id]
	}
	return out
}

// Sorted returns the admitted candidates by ascending distance, stably,
// so an exact distance tie keeps insertion order (match.py:326-328,
// Python's stable Timsort). Corpus reach: exactly two cases carry an
// exact candidate-distance tie, should-ask and should-reject, and in
// both the tie group's recorded order equals the order the case logged
// its release lookups (contract V11).
func (c *Candidates) Sorted() []Candidate {
	out := c.All()
	slices.SortStableFunc(out, func(a, b Candidate) int {
		return cmp.Compare(a.Dist.Distance(), b.Dist.Distance())
	})
	return out
}
