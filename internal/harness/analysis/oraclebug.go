package analysis

import (
	"errors"
	"fmt"
)

// Reproduced values for an OracleBugEntry. Reproduced is agent-settable
// (it is a machine-observable fact about whether a probe case moved the
// verdict); Disposition is not.
const (
	ReproducedYes          = "yes"
	ReproducedNo           = "no"
	ReproducedNotAttempted = "not-attempted"
)

// OracleBugEntry is one row per known-defect hypothesis, not one per
// probe case: a hypothesis may be probed by several cases, or by none.
// A hypothesis that does not reproduce is still a written row, matching
// the corpus's "an unacknowledged flag is a failure" discipline.
type OracleBugEntry struct {
	ID               string   `json:"id"`
	BugClass         string   `json:"bug_class"`
	Citation         string   `json:"citation"` // an issue number, or a finding id such as U1/ft-pattern
	ProbeCaseIDs     []string `json:"probe_case_ids,omitempty"`
	Reproduced       string   `json:"reproduced"`
	ReproducedReason string   `json:"reproduced_reason,omitempty"` // required when Reproduced is ReproducedNotAttempted
	AdjudicationRefs []string `json:"adjudication_refs,omitempty"` // AdjudicationEntry ids the probes produced
	Disposition      string   `json:"disposition"`
	Reason           string   `json:"reason"`
	Adjudicator      string   `json:"adjudicator"`
	Date             string   `json:"date"`
}

// OracleBugLedger is corpus/analysis/oracle_bugs.json: the
// known-oracle-bug hypothesis ledger.
type OracleBugLedger struct {
	Entries []OracleBugEntry `json:"entries"`
}

// LoadOracleBugs reads corpus/analysis/oracle_bugs.json from path.
func LoadOracleBugs(path string) (*OracleBugLedger, error) {
	var l OracleBugLedger
	if err := readJSON(path, &l); err != nil {
		return nil, fmt.Errorf("oracle bugs: %w", err)
	}
	return &l, nil
}

// CheckOracleBugs proves every row's probe case ids resolve to a
// committed corpus case, every adjudication_refs entry resolves to a
// row in adj, a row reproduced "yes" names at least one adjudication
// row, the same owner-only disposition rule CheckAdjudications
// enforces, and that EVERY probe case of a hypothesis at OracleWrong
// resolves to an adjudication row by subject.case_id.
//
// That last rule is the B3 fold. Oracle-bug rows are per hypothesis and
// adjudication rows are per case, so a hypothesis ruled oracle-wrong
// whose probe case carries no row would have no corrected verdict to
// produce and would be silently ungraded at the conformance gate. It
// takes the whole adjudication ledger rather than an id set because it
// needs both indexes, the row ids for adjudication_refs and the subject
// index for the probe rule, and building them here beats threading two
// maps through every caller.
func CheckOracleBugs(caseIDs []string, adj *AdjudicationLedger, l *OracleBugLedger) error {
	known := make(map[string]bool, len(caseIDs))
	for _, id := range caseIDs {
		known[id] = true
	}
	adjudicationIDs := make(map[string]bool, len(adj.Entries))
	for _, e := range adj.Entries {
		adjudicationIDs[e.ID] = true
	}
	bySubject := adj.ByCase()

	var errs []error
	seen := make(map[string]bool, len(l.Entries))
	for _, e := range l.Entries {
		if e.ID == "" {
			errs = append(errs, errors.New("oracle bugs: entry with no id"))
		} else if seen[e.ID] {
			errs = append(errs, fmt.Errorf("oracle bugs: %s: listed twice", e.ID))
		}
		seen[e.ID] = true
		errs = append(errs, checkOracleBugEntry(known, adjudicationIDs, e)...)
		if e.Disposition == OracleWrong {
			for _, caseID := range e.ProbeCaseIDs {
				if _, ok := bySubject[caseID]; !ok {
					errs = append(errs, fmt.Errorf("oracle bugs: %s: probe case %q: no adjudication row", e.ID, caseID))
				}
			}
		}
	}
	return errors.Join(errs...)
}

func checkOracleBugEntry(known, adjudicationIDs map[string]bool, e OracleBugEntry) []error {
	var errs []error

	if e.BugClass == "" {
		errs = append(errs, fmt.Errorf("oracle bugs: %s: bug_class: missing", e.ID))
	}
	if e.Citation == "" {
		errs = append(errs, fmt.Errorf("oracle bugs: %s: citation: missing", e.ID))
	}
	for _, id := range e.ProbeCaseIDs {
		if !known[id] {
			errs = append(errs, fmt.Errorf("oracle bugs: %s: probe case %q: no such committed case", e.ID, id))
		}
	}
	for _, ref := range e.AdjudicationRefs {
		if !adjudicationIDs[ref] {
			errs = append(errs, fmt.Errorf("oracle bugs: %s: adjudication_refs: %q: no such adjudication row", e.ID, ref))
		}
	}

	switch e.Reproduced {
	case ReproducedYes:
		if len(e.AdjudicationRefs) == 0 {
			errs = append(errs, fmt.Errorf("oracle bugs: %s: reproduced %q with no adjudication row referenced", e.ID, ReproducedYes))
		}
	case ReproducedNo:
	case ReproducedNotAttempted:
		if e.ReproducedReason == "" {
			errs = append(errs, fmt.Errorf("oracle bugs: %s: reproduced %q with no reason", e.ID, ReproducedNotAttempted))
		}
	default:
		errs = append(errs, fmt.Errorf("oracle bugs: %s: reproduced: missing or unrecognized (%q)", e.ID, e.Reproduced))
	}

	if e.Reason == "" {
		errs = append(errs, fmt.Errorf("oracle bugs: %s: reason: missing", e.ID))
	}
	if e.Adjudicator == "" {
		errs = append(errs, fmt.Errorf("oracle bugs: %s: adjudicator: missing", e.ID))
	}
	if e.Date == "" {
		errs = append(errs, fmt.Errorf("oracle bugs: %s: date: missing", e.ID))
	}

	switch e.Disposition {
	case Pending, OracleWrong, OracleCorrect, NoCorrectAnswer:
	default:
		errs = append(errs, fmt.Errorf("oracle bugs: %s: disposition: missing or unrecognized (%q)", e.ID, e.Disposition))
		return errs
	}
	if e.Disposition != Pending && e.Adjudicator == AgentAdjudicator {
		errs = append(errs, fmt.Errorf("oracle bugs: %s: disposition %q set by adjudicator %q, but only a human name may set anything but %q", e.ID, e.Disposition, e.Adjudicator, Pending))
	}

	return errs
}
