package analysis

import (
	"errors"
	"fmt"
	"strings"
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

	// DefectFree is the distance a case scores once the named defect is
	// removed, derived through pinned beets with one SD_PATTERNS
	// substitution and nothing else. It is derived data beside the
	// owner's own corrected_verdict prose, never a disposition and
	// never a verdict about who is right: it is the expectation an
	// oracle-wrong case is graded against once its hypothesis's compat
	// switch defaults off (ruling A / M6). A pointer because absence is
	// meaningful and omitempty does not omit a struct value.
	DefectFree *DefectFree `json:"defect_free,omitempty"`
}

// DefectFree is the derived-distance block ruling A / M6 attaches to an
// oracle-wrong adjudication row. See [AdjudicationEntry.DefectFree].
type DefectFree struct {
	Distance       float64 `json:"distance"`
	Recommendation string  `json:"recommendation"`
	Action         string  `json:"action"`
	PatternTable   string  `json:"pattern_table"`
}

// StringHeuristicHypotheses names the oracle-bug hypotheses whose
// defect is a defect OF SD_PATTERNS, and therefore the only ones whose
// corrected expectation a pattern-table swap can express. The other two
// oracle-wrong hypotheses are excluded on purpose:
// bug-nonlatin-alias-selection is a translation defect with no
// SD_PATTERNS counterfactual, and bug-assignment-rejection-gap is the
// class ruling A names as having no beets counterfactual at all, so its
// six adv-assign-* probe cases owe no block.
var StringHeuristicHypotheses = []string{"bug-ft-pattern", "bug-remix-parenthetical"}

// legalPatternTables is the two legal values [DefectFree.PatternTable]
// carries: the guarded table for bug-ft-pattern's six adv-ft-* rows and
// the bracket-discount-free table for bug-remix-parenthetical's four
// adv-remix-* rows. A hand-edit to a third string fails loudly rather
// than becoming an unreproducible figure.
var legalPatternTables = map[string]bool{
	"SDPatternsWordBoundary":      true,
	"SDPatternsNoBracketDiscount": true,
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

	if e.DefectFree != nil {
		if e.Disposition != OracleWrong {
			errs = append(errs, fmt.Errorf("adjudications: %s: defect_free block on a row whose disposition is %q, not %q", e.ID, e.Disposition, OracleWrong))
		}
		if e.DefectFree.Recommendation == "" {
			errs = append(errs, fmt.Errorf("adjudications: %s: defect_free: recommendation: missing", e.ID))
		}
		if e.DefectFree.Action == "" {
			errs = append(errs, fmt.Errorf("adjudications: %s: defect_free: action: missing", e.ID))
		}
		if e.DefectFree.PatternTable == "" {
			errs = append(errs, fmt.Errorf("adjudications: %s: defect_free: pattern_table: missing", e.ID))
		}
		if e.DefectFree.Distance < 0 || e.DefectFree.Distance > 1 {
			errs = append(errs, fmt.Errorf("adjudications: %s: defect_free: distance %v outside [0, 1]", e.ID, e.DefectFree.Distance))
		}
	}
	return errs
}

// CheckDefectFree proves the cross-ledger rule the per-row checker
// cannot see: every probe case of a [StringHeuristicHypotheses]
// hypothesis at [OracleWrong] resolves to an adjudication row carrying
// a defect_free block, and every block's recommendation and action
// appear verbatim in its row's corrected_verdict. It also forbids the
// converse: a defect_free block on any row whose case is not a probe
// of one of those hypotheses (fold, m12), which is what stops a
// fabricated block on an unrelated oracle-wrong row (the six
// adv-assign-* rows, say) from passing every check. It reports every
// failure rather than the first.
func CheckDefectFree(bugs *OracleBugLedger, adj *AdjudicationLedger) error {
	byID := make(map[string]OracleBugEntry, len(bugs.Entries))
	for _, e := range bugs.Entries {
		byID[e.ID] = e
	}

	rows := make(map[string]AdjudicationEntry, len(adj.Entries))
	for _, e := range adj.Entries {
		if e.Subject.CaseID != "" {
			rows[e.Subject.CaseID] = e
		}
	}

	var errs []error
	probeCases := make(map[string]bool)
	for _, hyp := range StringHeuristicHypotheses {
		entry, ok := byID[hyp]
		if !ok {
			errs = append(errs, fmt.Errorf("defect_free: hypothesis %q: no such oracle bug entry", hyp))
			continue
		}
		if entry.Disposition != OracleWrong {
			continue
		}
		for _, caseID := range entry.ProbeCaseIDs {
			probeCases[caseID] = true
			row, ok := rows[caseID]
			if !ok {
				errs = append(errs, fmt.Errorf("defect_free: probe case %q (%s): no adjudication row", caseID, hyp))
				continue
			}
			if row.DefectFree == nil {
				errs = append(errs, fmt.Errorf("defect_free: probe case %q (%s): adjudication row %s carries no defect_free block", caseID, hyp, row.ID))
				continue
			}
			block := row.DefectFree
			if !legalPatternTables[block.PatternTable] {
				errs = append(errs, fmt.Errorf("defect_free: %s: pattern_table %q is not one of the two legal values (SDPatternsWordBoundary, SDPatternsNoBracketDiscount)", row.ID, block.PatternTable))
			}
			if !strings.Contains(row.CorrectedVerdict, block.Recommendation) {
				errs = append(errs, fmt.Errorf("defect_free: %s: recommendation %q does not appear verbatim in corrected_verdict %q", row.ID, block.Recommendation, row.CorrectedVerdict))
			}
			if !strings.Contains(row.CorrectedVerdict, block.Action) {
				errs = append(errs, fmt.Errorf("defect_free: %s: action %q does not appear verbatim in corrected_verdict %q", row.ID, block.Action, row.CorrectedVerdict))
			}
		}
	}

	for _, e := range adj.Entries {
		if e.DefectFree == nil {
			continue
		}
		if !probeCases[e.Subject.CaseID] {
			errs = append(errs, fmt.Errorf("defect_free: %s: carries a defect_free block, but its case %q is not a probe of a StringHeuristicHypotheses hypothesis", e.ID, e.Subject.CaseID))
		}
	}

	return errors.Join(errs...)
}
