package analysis

import (
	"errors"
	"fmt"
)

// Adjudication dispositions. Pending is legal at every gate this
// package runs (bash scripts/check.sh); the launch gate, a different
// gate at a different time, refuses it. OracleWrong carries the
// adjudicated correct verdict in CorrectedVerdict. OracleCorrect means
// the crafted case failed to reproduce the suspected defect.
// NoCorrectAnswer means the input material has no correct MusicBrainz
// entry at all.
const (
	Pending         = "pending-adjudication"
	OracleWrong     = "oracle-wrong"
	OracleCorrect   = "oracle-correct"
	NoCorrectAnswer = "no-correct-answer"
)

// AgentAdjudicator is the reserved, non-human adjudicator value an
// agent-authored row must carry. disposition is owner-only (Geoff's
// alone to set to anything but Pending); a row's adjudicator being
// this literal sentinel is how the checker tells an agent-authored row
// from a human-ruled one, since Go has no other way to recognize "a
// human name."
const AgentAdjudicator = "agent"

// AdjudicationSubject is the join key a row's evidence is read against:
// a corpus case id today, or an album plus release MBID for a live
// shadow divergence rung 4 will write. Exactly one shape is populated.
type AdjudicationSubject struct {
	CaseID      string `json:"case_id,omitempty"`
	Album       string `json:"album,omitempty"`
	ReleaseMBID string `json:"release_mbid,omitempty"`
}

// AdjudicationEntry is one divergence row: a case (or, at rung 4, a
// live shadow run) whose derived verdict is suspected wrong against
// release truth. The oracle's own record (recommendation, action,
// distance, the best candidate's penalty breakdown) is never
// re-transcribed here; Subject is the reference back to it.
type AdjudicationEntry struct {
	ID               string              `json:"id"`
	Subject          AdjudicationSubject `json:"subject"`
	Suspected        string              `json:"suspected"`     // what is wrong and why, in prose
	Citation         string              `json:"citation"`      // an issue number, or a finding id such as U1/ft-pattern
	ReleaseTruth     string              `json:"release_truth"` // the correct verdict, or a statement that none exists
	Disposition      string              `json:"disposition"`
	CorrectedVerdict string              `json:"corrected_verdict,omitempty"` // required when Disposition is OracleWrong
	Reason           string              `json:"reason"`
	Adjudicator      string              `json:"adjudicator"`
	Date             string              `json:"date"`
}

// AdjudicationLedger is corpus/analysis/adjudications.json: the general
// divergence log, forward-compatible with rung 4's shadow divergences.
// Unlike ties.json and proximity.json it is not a coverage ledger over
// every corpus case; it is an exception log, carrying a row only for a
// case an author or a batch task suspects the oracle got wrong.
type AdjudicationLedger struct {
	Entries []AdjudicationEntry `json:"entries"`
}

// LoadAdjudications reads corpus/analysis/adjudications.json from path.
func LoadAdjudications(path string) (*AdjudicationLedger, error) {
	var l AdjudicationLedger
	if err := readJSON(path, &l); err != nil {
		return nil, fmt.Errorf("adjudications: %w", err)
	}
	return &l, nil
}

// CheckAdjudications proves every row is internally consistent and, for
// a case-subject row, resolves to a committed corpus case: a missing or
// unrecognized disposition, a row with no written suspicion or no
// written release truth, and a non-pending disposition set by an agent
// adjudicator are all reported, not just the first.
func CheckAdjudications(caseIDs []string, l *AdjudicationLedger) error {
	known := make(map[string]bool, len(caseIDs))
	for _, id := range caseIDs {
		known[id] = true
	}

	var errs []error
	seen := make(map[string]bool, len(l.Entries))
	for _, e := range l.Entries {
		if e.ID == "" {
			errs = append(errs, errors.New("adjudications: entry with no id"))
		} else if seen[e.ID] {
			errs = append(errs, fmt.Errorf("adjudications: %s: listed twice", e.ID))
		}
		seen[e.ID] = true
		errs = append(errs, checkAdjudicationEntry(known, e)...)
	}
	return errors.Join(errs...)
}

func checkAdjudicationEntry(known map[string]bool, e AdjudicationEntry) []error {
	var errs []error

	switch {
	case e.Subject.CaseID != "" && (e.Subject.Album != "" || e.Subject.ReleaseMBID != ""):
		errs = append(errs, fmt.Errorf("adjudications: %s: subject: case_id and album/release_mbid both set; a row is one or the other", e.ID))
	case e.Subject.CaseID != "":
		if !known[e.Subject.CaseID] {
			errs = append(errs, fmt.Errorf("adjudications: %s: subject case_id %q: no such committed case", e.ID, e.Subject.CaseID))
		}
	case e.Subject.Album != "" && e.Subject.ReleaseMBID != "":
		// A shadow-divergence subject; nothing to resolve against the corpus.
	case e.Subject.Album != "" || e.Subject.ReleaseMBID != "":
		errs = append(errs, fmt.Errorf("adjudications: %s: subject: album and release_mbid must both be set, or neither", e.ID))
	default:
		errs = append(errs, fmt.Errorf("adjudications: %s: subject: names neither a case id nor an album/release_mbid pair", e.ID))
	}

	if e.Suspected == "" {
		errs = append(errs, fmt.Errorf("adjudications: %s: suspected: missing", e.ID))
	}
	if e.ReleaseTruth == "" {
		errs = append(errs, fmt.Errorf("adjudications: %s: release_truth: missing", e.ID))
	}
	if e.Reason == "" {
		errs = append(errs, fmt.Errorf("adjudications: %s: reason: missing", e.ID))
	}
	if e.Adjudicator == "" {
		errs = append(errs, fmt.Errorf("adjudications: %s: adjudicator: missing", e.ID))
	}
	if e.Date == "" {
		errs = append(errs, fmt.Errorf("adjudications: %s: date: missing", e.ID))
	}

	switch e.Disposition {
	case Pending, OracleWrong, OracleCorrect, NoCorrectAnswer:
	default:
		errs = append(errs, fmt.Errorf("adjudications: %s: disposition: missing or unrecognized (%q)", e.ID, e.Disposition))
		return errs
	}
	if e.Disposition != Pending && e.Adjudicator == AgentAdjudicator {
		errs = append(errs, fmt.Errorf("adjudications: %s: disposition %q set by adjudicator %q, but only a human name may set anything but %q", e.ID, e.Disposition, e.Adjudicator, Pending))
	}
	if e.Disposition == OracleWrong && e.CorrectedVerdict == "" {
		errs = append(errs, fmt.Errorf("adjudications: %s: disposition %q needs a corrected_verdict", e.ID, OracleWrong))
	}
	return errs
}
