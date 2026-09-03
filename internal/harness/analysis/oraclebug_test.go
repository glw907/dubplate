package analysis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckOracleBugs(t *testing.T) {
	base := OracleBugEntry{
		ID:               "bug-ft-pattern",
		BugClass:         "featuring-pattern false positive",
		Citation:         "742ebdf69",
		ProbeCaseIDs:     []string{"clean-match"},
		Reproduced:       ReproducedNotAttempted,
		ReproducedReason: "no probe case authored yet; Task 16 authors it",
		Disposition:      Pending,
		Reason:           "seeded this pass, ahead of the adversarial stratum",
		Adjudicator:      AgentAdjudicator,
		Date:             "2026-09-03",
	}

	unknownProbe := base
	unknownProbe.ProbeCaseIDs = []string{"retired-case"}

	noBugClass := base
	noBugClass.BugClass = ""

	noCitation := base
	noCitation.Citation = ""

	notAttemptedNoReason := base
	notAttemptedNoReason.ReproducedReason = ""

	badReproduced := base
	badReproduced.Reproduced = "maybe"

	reproducedYesNoRef := base
	reproducedYesNoRef.Reproduced = ReproducedYes
	reproducedYesNoRef.ReproducedReason = ""

	reproducedYesWithRef := reproducedYesNoRef
	reproducedYesWithRef.AdjudicationRefs = []string{"adj-1"}

	unknownRef := base
	unknownRef.Reproduced = ReproducedNo
	unknownRef.ReproducedReason = ""
	unknownRef.AdjudicationRefs = []string{"retired-adj"}

	noReason := base
	noReason.Reason = ""

	noAdjudicator := base
	noAdjudicator.Adjudicator = ""

	noDate := base
	noDate.Date = ""

	badDisposition := base
	badDisposition.Disposition = "oracle-shrug"

	agentRulesOracleCorrect := base
	agentRulesOracleCorrect.Reproduced = ReproducedNo
	agentRulesOracleCorrect.ReproducedReason = ""
	agentRulesOracleCorrect.Disposition = OracleCorrect

	humanRulesOracleCorrect := agentRulesOracleCorrect
	humanRulesOracleCorrect.Adjudicator = "Geoff"

	tests := []struct {
		name            string
		caseIDs         []string
		adjudicationIDs map[string]bool
		entries         []OracleBugEntry
		wantHas         string // empty means the ledger is accepted
	}{
		{"a well-formed not-attempted row", []string{"clean-match"}, nil, []OracleBugEntry{base}, ""},
		{"a human ruling oracle-correct", []string{"clean-match"}, nil, []OracleBugEntry{humanRulesOracleCorrect}, ""},
		{"reproduced yes with a resolvable ref", []string{"clean-match"}, map[string]bool{"adj-1": true}, []OracleBugEntry{reproducedYesWithRef}, ""},
		{"a probe case no committed case carries", []string{"clean-match"}, nil, []OracleBugEntry{unknownProbe}, "no such committed case"},
		{"no bug class", []string{"clean-match"}, nil, []OracleBugEntry{noBugClass}, "bug_class: missing"},
		{"no citation", []string{"clean-match"}, nil, []OracleBugEntry{noCitation}, "citation: missing"},
		{"not-attempted with no reason", []string{"clean-match"}, nil, []OracleBugEntry{notAttemptedNoReason}, "reproduced \"not-attempted\" with no reason"},
		{"an unrecognized reproduced value", []string{"clean-match"}, nil, []OracleBugEntry{badReproduced}, "reproduced: missing or unrecognized"},
		{"reproduced yes with no adjudication row referenced", []string{"clean-match"}, nil, []OracleBugEntry{reproducedYesNoRef}, "no adjudication row referenced"},
		{"a ref no adjudication row carries", []string{"clean-match"}, map[string]bool{"adj-1": true}, []OracleBugEntry{unknownRef}, "no such adjudication row"},
		{"no reason", []string{"clean-match"}, nil, []OracleBugEntry{noReason}, "reason: missing"},
		{"no adjudicator", []string{"clean-match"}, nil, []OracleBugEntry{noAdjudicator}, "adjudicator: missing"},
		{"no date", []string{"clean-match"}, nil, []OracleBugEntry{noDate}, "date: missing"},
		{"an unrecognized disposition", []string{"clean-match"}, nil, []OracleBugEntry{badDisposition}, "unrecognized"},
		{"an agent ruling oracle-correct", []string{"clean-match"}, nil, []OracleBugEntry{agentRulesOracleCorrect}, "only a human name may set anything but"},
		{"one id listed twice", []string{"clean-match"}, nil, []OracleBugEntry{base, base}, "listed twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckOracleBugs(tt.caseIDs, tt.adjudicationIDs, &OracleBugLedger{Entries: tt.entries})
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckOracleBugs: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckOracleBugs accepted a ledger that is %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

func TestLoadOracleBugsRejectsAnInventedKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "oracle_bugs.json")
	if err := os.WriteFile(path, []byte(`{"entries":[],"schema_version":"1"}`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if _, err := LoadOracleBugs(path); err == nil {
		t.Fatal("LoadOracleBugs accepted a key nothing parses")
	}
}
