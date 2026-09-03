package analysis

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// ProximityEpsilon is the half-width of the band around a
// recommendation boundary that a case has to fall inside to be flagged.
// beets compares the strong threshold with a strict "<" and the medium
// one with "<=" (beets/autotag/match.py, _recommendation), so a case
// this close to a boundary can change verdict under a
// comparison-direction disagreement that no float tolerance absorbs.
// The mechanics study's epsilon-regime finding is the source.
const ProximityEpsilon = 1e-6

// Proximity statuses, one per case.
const (
	// ProximityClear means the record's distances sit clear of every boundary.
	ProximityClear = "clear"
	// ProximityFlagged means at least one boundary is within ProximityEpsilon.
	ProximityFlagged = "flagged"
	// ProximitySkipped means the record has no candidate to measure, so no
	// boundary applies. A zero-track candidate release is dropped by beets'
	// own _add_candidate, which is how a committed record reaches this state.
	ProximitySkipped = "skipped"
)

// Proximity is one case's recomputed standing against the verdict
// boundaries its own record echoed.
type Proximity struct {
	Status     string
	Boundaries []string // threshold keys within ProximityEpsilon, sorted; empty unless Status is ProximityFlagged
}

// ProximityLedger is corpus/analysis/proximity.json: the acknowledgment
// list for every case the flag fires on. A flag is not a failure; an
// unacknowledged flag is.
type ProximityLedger struct {
	Entries []ProximityEntry `json:"entries"`
}

// ProximityEntry acknowledges one flagged or skipped case. Reason is
// the author's written justification and is never optional: an entry
// without one records that a case was noticed, not why it is allowed to
// sit on a boundary.
type ProximityEntry struct {
	CaseID     string   `json:"case_id"`
	Status     string   `json:"status"`               // flagged | skipped
	Boundaries []string `json:"boundaries,omitempty"` // required for a flagged entry, forbidden otherwise
	Reason     string   `json:"reason"`
}

// Flag recomputes exp's proximity to the recommendation boundaries the
// record itself echoed. The thresholds come from the record rather than
// from a constant here, so a case derived under a profile that moves
// strong_rec_thresh is measured against the boundary it actually scored
// under. A record carrying no thresholds at all is an error, not a case
// measured against zeros.
func Flag(exp *cassette.Expected) (Proximity, error) {
	t := exp.Thresholds
	if t.StrongRecThresh == 0 && t.MediumRecThresh == 0 && t.RecGapThresh == 0 {
		return Proximity{}, errors.New("record echoes no thresholds; it predates the schema-2 echo and cannot be flagged")
	}
	if len(exp.Candidates) == 0 {
		return Proximity{Status: ProximitySkipped}, nil
	}

	best := exp.Candidates[0].Distance
	var boundaries []string
	if withinEpsilon(best, t.StrongRecThresh) {
		boundaries = append(boundaries, "strong_rec_thresh")
	}
	if withinEpsilon(best, t.MediumRecThresh) {
		boundaries = append(boundaries, "medium_rec_thresh")
	}
	if len(exp.Candidates) > 1 && withinEpsilon(exp.Candidates[1].Distance-best, t.RecGapThresh) {
		boundaries = append(boundaries, "rec_gap_thresh")
	}
	if boundaries == nil {
		return Proximity{Status: ProximityClear}, nil
	}
	return Proximity{Status: ProximityFlagged, Boundaries: boundaries}, nil
}

func withinEpsilon(got, want float64) bool {
	return math.Abs(got-want) <= ProximityEpsilon
}

// LoadProximityLedger reads corpus/analysis/proximity.json from path.
func LoadProximityLedger(path string) (*ProximityLedger, error) {
	var l ProximityLedger
	if err := readJSON(path, &l); err != nil {
		return nil, fmt.Errorf("proximity ledger: %w", err)
	}
	return &l, nil
}

// CheckProximityLedger diffs the ledger against got, the flag
// recomputed from every committed record, keyed by case id. It reports
// a case flagged or skipped but not listed, a listed case that is no
// longer either, a listed case whose boundaries have moved, a listed
// case with no written reason, and an entry naming a case the corpus
// does not carry. Every disagreement is reported, not just the first,
// so one gate run names the whole repair.
func CheckProximityLedger(got map[string]Proximity, l *ProximityLedger) error {
	var errs []error

	listed := make(map[string]ProximityEntry, len(l.Entries))
	for _, e := range l.Entries {
		if _, dup := listed[e.CaseID]; dup {
			errs = append(errs, fmt.Errorf("proximity ledger: %s: listed twice", e.CaseID))
			continue
		}
		listed[e.CaseID] = e
		if _, known := got[e.CaseID]; !known {
			errs = append(errs, fmt.Errorf("proximity ledger: %s: no such corpus case", e.CaseID))
		}
	}

	for _, id := range slices.Sorted(maps.Keys(got)) {
		p := got[id]
		e, ok := listed[id]
		if p.Status == ProximityClear {
			if ok {
				errs = append(errs, fmt.Errorf("proximity ledger: %s: listed as %q but its record is no longer near any boundary", id, e.Status))
			}
			continue
		}
		if !ok {
			errs = append(errs, fmt.Errorf("proximity ledger: %s: %s (%v) but not listed", id, p.Status, p.Boundaries))
			continue
		}
		if e.Status != p.Status {
			errs = append(errs, fmt.Errorf("proximity ledger: %s: listed as %q, recomputes as %q", id, e.Status, p.Status))
		}
		if !slices.Equal(e.Boundaries, p.Boundaries) {
			errs = append(errs, fmt.Errorf("proximity ledger: %s: listed boundaries %v, recomputed %v", id, e.Boundaries, p.Boundaries))
		}
		if e.Reason == "" {
			errs = append(errs, fmt.Errorf("proximity ledger: %s: listed with no reason", id))
		}
	}

	return errors.Join(errs...)
}
