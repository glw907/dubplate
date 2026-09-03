package analysis

import (
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// ClaimsFileName is the per-case claims file, a sibling of case.json.
// The claims deliberately sit outside cassette.Expected: the record
// schema is frozen, and a claim is an author's reading of a derived
// record rather than part of the record itself.
const ClaimsFileName = "claims.json"

// Claims is what a case's author asserts about its derived record,
// written off that record and never invented. It is the check that
// keeps a degraded case from freezing green: a case whose candidate
// list collapsed, whose best release changed, or whose family penalty
// stopped firing still re-derives to a self-consistent record, and only
// a claim written against the behavior the case exists to exercise can
// tell the difference.
//
// MinCandidates defaults to 1 when absent. Set to 0 with BestAlbumID
// left empty, it declares the zero-candidate mode: the case's record
// carries no candidate at all (a zero-track release, dropped by beets'
// own _add_candidate), so there is no best candidate to name and
// [CheckClaims] asserts the empty list directly instead of treating
// zero as a trivially-satisfied lower bound. The remaining optional
// fields are per family: Assignment, ExtraItemsLen, and ExtraTracksLen
// for the assignment family, AppliedFields for the per_disc_numbering
// cases. PenaltyKeysAbsent names keys the best candidate's penalty map
// must not carry at all, distinct from PenaltyKeysZero: a component
// gated on a falsy `likelies` value never runs `add_*`, so the key is
// never added, where a component that runs and finds no distance adds
// the key at 0.0 and still counts it in the normalizer's denominator.
type Claims struct {
	MinCandidates      *int                      `json:"min_candidates,omitempty"`
	BestAlbumID        string                    `json:"best_album_id"`
	Recommendation     string                    `json:"recommendation"`
	PenaltyKeysNonzero []string                  `json:"penalty_keys_nonzero,omitempty"`
	PenaltyKeysZero    []string                  `json:"penalty_keys_zero,omitempty"`
	PenaltyKeysAbsent  []string                  `json:"penalty_keys_absent,omitempty"`
	Assignment         map[string]int            `json:"assignment,omitempty"`
	ExtraItemsLen      *int                      `json:"extra_items_len,omitempty"`
	ExtraTracksLen     *int                      `json:"extra_tracks_len,omitempty"`
	AppliedFields      map[string]map[string]any `json:"applied_fields,omitempty"`
}

// LoadClaims reads the claims file in a case directory. A missing file
// is an error: every committed case carries one. best_album_id is
// required unless the claims declare [Claims.declaresZeroCandidates],
// the zero-candidate mode a case whose only candidate was dropped by
// beets' own _add_candidate has to state instead of naming one.
func LoadClaims(dir string) (*Claims, error) {
	var c Claims
	if err := readJSON(filepath.Join(dir, ClaimsFileName), &c); err != nil {
		return nil, fmt.Errorf("claims: %w", err)
	}
	if c.BestAlbumID == "" && !c.declaresZeroCandidates() {
		return nil, errors.New("claims: best_album_id: missing")
	}
	if c.Recommendation == "" {
		return nil, errors.New("claims: recommendation: missing")
	}
	return &c, nil
}

// declaresZeroCandidates reports whether c states the zero-candidate
// claims mode: min_candidates explicitly 0 and no best_album_id, since
// a record with no candidate has no best one to name.
func (c *Claims) declaresZeroCandidates() bool {
	return c.MinCandidates != nil && *c.MinCandidates == 0 && c.BestAlbumID == ""
}

// CheckClaims reports every claim exp fails to satisfy. A penalty key
// claimed zero must be present in the record at 0.0 rather than absent:
// beets adds a key at 0.0 and still counts it in the normalizer's
// denominator, so "added at zero" and "never added" are different facts
// about a case (the mechanics study, Q2).
func CheckClaims(exp *cassette.Expected, c *Claims) error {
	var errs []error

	if c.declaresZeroCandidates() {
		// The mode names the record's shape directly rather than folding
		// into the ordinary "at least minCandidates" check below: min 0 is
		// trivially satisfied by any candidate count, so without this
		// explicit assertion the mode would wave through a degraded case
		// whose candidate list is nonempty but the author never noticed.
		if len(exp.Candidates) != 0 {
			errs = append(errs, fmt.Errorf("claims: min_candidates: 0 declares the record carries no candidate, but it carries %d", len(exp.Candidates)))
		}
		if exp.Recommendation != c.Recommendation {
			errs = append(errs, fmt.Errorf("claims: recommendation = %q, want %q", exp.Recommendation, c.Recommendation))
		}
		return errors.Join(errs...)
	}

	minCandidates := 1
	if c.MinCandidates != nil {
		minCandidates = *c.MinCandidates
	}
	if len(exp.Candidates) < minCandidates {
		errs = append(errs, fmt.Errorf("claims: candidates = %d, want at least %d", len(exp.Candidates), minCandidates))
	}
	if exp.Recommendation != c.Recommendation {
		errs = append(errs, fmt.Errorf("claims: recommendation = %q, want %q", exp.Recommendation, c.Recommendation))
	}
	if len(exp.Candidates) == 0 {
		// Every remaining claim reads the best candidate; without one there
		// is nothing further to check, and indexing would panic.
		errs = append(errs, errors.New("claims: record carries no candidate, so best_album_id and the penalty claims are unverifiable"))
		return errors.Join(errs...)
	}

	best := exp.Candidates[0]
	if best.AlbumID != c.BestAlbumID {
		errs = append(errs, fmt.Errorf("claims: best_album_id = %q, want %q", best.AlbumID, c.BestAlbumID))
	}
	for _, key := range c.PenaltyKeysNonzero {
		switch v, ok := best.Penalties[key]; {
		case !ok:
			errs = append(errs, fmt.Errorf("claims: penalty %q claimed nonzero but the best candidate never added it", key))
		case v == 0:
			errs = append(errs, fmt.Errorf("claims: penalty %q claimed nonzero, recorded 0", key))
		}
	}
	for _, key := range c.PenaltyKeysZero {
		switch v, ok := best.Penalties[key]; {
		case !ok:
			errs = append(errs, fmt.Errorf("claims: penalty %q claimed zero but the best candidate never added it", key))
		case v != 0:
			errs = append(errs, fmt.Errorf("claims: penalty %q claimed zero, recorded %v", key, v))
		}
	}
	for _, key := range c.PenaltyKeysAbsent {
		if _, ok := best.Penalties[key]; ok {
			errs = append(errs, fmt.Errorf("claims: penalty %q claimed absent, but the best candidate carries it", key))
		}
	}
	if c.ExtraItemsLen != nil && len(exp.ExtraItems) != *c.ExtraItemsLen {
		errs = append(errs, fmt.Errorf("claims: extra_items = %d, want %d", len(exp.ExtraItems), *c.ExtraItemsLen))
	}
	if c.ExtraTracksLen != nil && len(exp.ExtraTracks) != *c.ExtraTracksLen {
		errs = append(errs, fmt.Errorf("claims: extra_tracks = %d, want %d", len(exp.ExtraTracks), *c.ExtraTracksLen))
	}
	errs = append(errs, checkAssignment(exp.Assignment, c.Assignment)...)
	errs = append(errs, checkAppliedFields(exp.AppliedFields, c.AppliedFields)...)

	return errors.Join(errs...)
}

// checkAssignment compares a claimed item-index to track-index mapping
// against the record's own. The claim's keys are JSON object keys, so
// they arrive as strings where the record's are already integers.
func checkAssignment(got map[int]int, want map[string]int) []error {
	if want == nil {
		return nil
	}
	var errs []error
	if len(got) != len(want) {
		errs = append(errs, fmt.Errorf("claims: assignment covers %d items, want %d", len(got), len(want)))
	}
	for _, key := range slices.Sorted(maps.Keys(want)) {
		item, err := strconv.Atoi(key)
		if err != nil {
			errs = append(errs, fmt.Errorf("claims: assignment key %q is not an item index", key))
			continue
		}
		track, ok := got[item]
		if !ok {
			errs = append(errs, fmt.Errorf("claims: assignment claims item %d, which the record leaves unassigned", item))
			continue
		}
		if track != want[key] {
			errs = append(errs, fmt.Errorf("claims: assignment item %d -> track %d, want %d", item, track, want[key]))
		}
	}
	return errs
}

// checkAppliedFields compares claimed spot values against the applied
// record, which is keyed by the item's index in the case's pinned input
// order. Both sides decode from JSON, so a number is a float64 on each
// and reflect.DeepEqual compares like with like.
func checkAppliedFields(got map[string]any, want map[string]map[string]any) []error {
	var errs []error
	for _, item := range slices.Sorted(maps.Keys(want)) {
		fields, ok := got[item].(map[string]any)
		if !ok {
			errs = append(errs, fmt.Errorf("claims: applied_fields claims item %s, which the record does not carry", item))
			continue
		}
		for _, name := range slices.Sorted(maps.Keys(want[item])) {
			value, ok := fields[name]
			if !ok {
				errs = append(errs, fmt.Errorf("claims: applied_fields item %s: field %q was never applied", item, name))
				continue
			}
			if !reflect.DeepEqual(value, want[item][name]) {
				errs = append(errs, fmt.Errorf("claims: applied_fields item %s: %s = %v, want %v", item, name, value, want[item][name]))
			}
		}
	}
	return errs
}
