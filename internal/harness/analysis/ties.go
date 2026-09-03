package analysis

import (
	"errors"
	"fmt"
	"math"
	"slices"
)

// TieMargin is the cost margin by which a re-solve that forbids one
// assigned pair must exceed the optimum for that optimum to count as
// unique. It sits three orders of magnitude clear of ProximityEpsilon
// so a tie verdict can never turn on the same float noise the
// verdict-boundary flag is watching.
const TieMargin = 1e-9

// Tie dispositions, the two ladder entry 2 allows for a tied case.
const (
	// TieExcluded means the case is kept out of conformance grading.
	TieExcluded = "excluded"
	// TiePinned means the case keeps its recorded mapping as the expectation.
	TiePinned = "pinned"
)

// TieLedger is corpus/analysis/ties.json: one row per corpus case,
// written by internal/harness/oracle's derive_ties.py driver, checked
// here. lapjv picks one optimal assignment; when several assignments
// share the minimal cost, a Go solver may pick a different one and
// still score the same assignment cost, because beets recomputes the
// final per-track distances with incl_artist=va while the assignment
// objective ignores the track artist. That divergence is why a tied
// case has to be dispositioned rather than merely noticed.
type TieLedger struct {
	Entries []TieEntry `json:"entries"`
}

// TieEntry is one case's uniqueness measurement. The costs are
// recorded, not only the boolean: a case that becomes tied later is
// then a visible movement in a number rather than a silent flip.
// OptimalCost is the assignment problem's optimum, BestForbiddenCost
// the cheapest re-solve that forbids one of the optimum's own pairs,
// and Difference their gap.
type TieEntry struct {
	CaseID            string  `json:"case_id"`
	AlbumID           string  `json:"album_id"` // the candidate the tie question was asked of, Candidates[0]
	OptimalCost       float64 `json:"optimal_cost"`
	BestForbiddenCost float64 `json:"best_forbidden_cost"`
	Difference        float64 `json:"difference"`
	Unique            bool    `json:"unique"`
	Disposition       string  `json:"disposition,omitempty"` // excluded | pinned; required when Unique is false
	Reason            string  `json:"reason,omitempty"`
}

// LoadTieLedger reads corpus/analysis/ties.json from path.
func LoadTieLedger(path string) (*TieLedger, error) {
	var l TieLedger
	if err := readJSON(path, &l); err != nil {
		return nil, fmt.Errorf("tie ledger: %w", err)
	}
	return &l, nil
}

// CheckTieLedger proves the ledger covers caseIDs exactly and that
// every tied case carries a disposition with a reason. It also checks
// each row against itself: the recorded difference must be the two
// recorded costs' gap, and the uniqueness verdict must be that
// difference measured against TieMargin, so a hand-edited row cannot
// claim a uniqueness its own numbers deny.
func CheckTieLedger(caseIDs []string, l *TieLedger) error {
	var errs []error

	listed := make(map[string]bool, len(l.Entries))
	known := make(map[string]bool, len(caseIDs))
	for _, id := range caseIDs {
		known[id] = true
	}

	for _, e := range l.Entries {
		if listed[e.CaseID] {
			errs = append(errs, fmt.Errorf("tie ledger: %s: listed twice", e.CaseID))
			continue
		}
		listed[e.CaseID] = true
		if !known[e.CaseID] {
			errs = append(errs, fmt.Errorf("tie ledger: %s: no such corpus case", e.CaseID))
			continue
		}
		errs = append(errs, checkTieEntry(e)...)
	}

	for _, id := range slices.Sorted(slices.Values(caseIDs)) {
		if !listed[id] {
			errs = append(errs, fmt.Errorf("tie ledger: %s: missing; every corpus case is measured, not only the tied ones", id))
		}
	}

	return errors.Join(errs...)
}

func checkTieEntry(e TieEntry) []error {
	var errs []error
	if e.AlbumID == "" {
		errs = append(errs, fmt.Errorf("tie ledger: %s: album_id: missing", e.CaseID))
	}
	if gap := e.BestForbiddenCost - e.OptimalCost; math.Abs(gap-e.Difference) > 1e-12 {
		errs = append(errs, fmt.Errorf("tie ledger: %s: difference = %v, but the recorded costs differ by %v", e.CaseID, e.Difference, gap))
	}
	if want := e.Difference > TieMargin; e.Unique != want {
		errs = append(errs, fmt.Errorf("tie ledger: %s: unique = %t, but a difference of %v against a %v margin says %t", e.CaseID, e.Unique, e.Difference, TieMargin, want))
	}
	if e.Unique {
		if e.Disposition != "" {
			errs = append(errs, fmt.Errorf("tie ledger: %s: dispositioned %q, but its optimum is unique", e.CaseID, e.Disposition))
		}
		return errs
	}
	if e.Disposition != TieExcluded && e.Disposition != TiePinned {
		errs = append(errs, fmt.Errorf("tie ledger: %s: tied, so disposition must be %q or %q, got %q", e.CaseID, TieExcluded, TiePinned, e.Disposition))
	}
	if e.Reason == "" {
		errs = append(errs, fmt.Errorf("tie ledger: %s: tied and dispositioned with no reason", e.CaseID))
	}
	return errs
}
