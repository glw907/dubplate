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
		name    string
		caseIDs []string
		adj     *AdjudicationLedger
		entries []OracleBugEntry
		wantHas string // empty means the ledger is accepted
	}{
		{"a well-formed not-attempted row", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{base}, ""},
		{"a human ruling oracle-correct", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{humanRulesOracleCorrect}, ""},
		{"reproduced yes with a resolvable ref", []string{"clean-match"}, &AdjudicationLedger{Entries: []AdjudicationEntry{{ID: "adj-1"}}}, []OracleBugEntry{reproducedYesWithRef}, ""},
		{"a probe case no committed case carries", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{unknownProbe}, "no such committed case"},
		{"no bug class", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{noBugClass}, "bug_class: missing"},
		{"no citation", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{noCitation}, "citation: missing"},
		{"not-attempted with no reason", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{notAttemptedNoReason}, "reproduced \"not-attempted\" with no reason"},
		{"an unrecognized reproduced value", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{badReproduced}, "reproduced: missing or unrecognized"},
		{"reproduced yes with no adjudication row referenced", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{reproducedYesNoRef}, "no adjudication row referenced"},
		{"a ref no adjudication row carries", []string{"clean-match"}, &AdjudicationLedger{Entries: []AdjudicationEntry{{ID: "adj-1"}}}, []OracleBugEntry{unknownRef}, "no such adjudication row"},
		{"no reason", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{noReason}, "reason: missing"},
		{"no adjudicator", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{noAdjudicator}, "adjudicator: missing"},
		{"no date", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{noDate}, "date: missing"},
		{"an unrecognized disposition", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{badDisposition}, "unrecognized"},
		{"an agent ruling oracle-correct", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{agentRulesOracleCorrect}, "only a human name may set anything but"},
		{"one id listed twice", []string{"clean-match"}, &AdjudicationLedger{}, []OracleBugEntry{base, base}, "listed twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckOracleBugs(tt.caseIDs, tt.adj, &OracleBugLedger{Entries: tt.entries})
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

// loadProbeRuleFixture loads testdata/oracle_bugs_probe_rule's single
// oracle_bugs.json and adjudications.json, and returns the one
// oracle-bug entry named name (paired with the full adjudication
// ledger, since the probe rule joins across the whole ledger) plus the
// case ids that entry's own probes name.
func loadProbeRuleFixture(t *testing.T, name string) (*OracleBugLedger, *AdjudicationLedger, []string) {
	t.Helper()
	bugs, err := LoadOracleBugs("testdata/oracle_bugs_probe_rule/oracle_bugs.json")
	if err != nil {
		t.Fatalf("LoadOracleBugs: %v", err)
	}
	adj, err := LoadAdjudications("testdata/oracle_bugs_probe_rule/adjudications.json")
	if err != nil {
		t.Fatalf("LoadAdjudications: %v", err)
	}
	for _, e := range bugs.Entries {
		if e.ID == name {
			return &OracleBugLedger{Entries: []OracleBugEntry{e}}, adj, e.ProbeCaseIDs
		}
	}
	t.Fatalf("testdata/oracle_bugs_probe_rule/oracle_bugs.json carries no entry named %q", name)
	return nil, nil, nil
}

func TestCheckOracleBugsRequiresEveryOracleWrongProbeToResolve(t *testing.T) {
	// The rule fires on NOTHING in the committed corpus: every probe
	// case of every oracle-wrong hypothesis resolves today, and the two
	// probe cases with no adjudication row sit under oracle-correct
	// hypotheses. So its proof is this hand-authored pair, per the
	// spec's assertion discipline.
	tests := []struct {
		name    string
		fixture string // testdata/oracle_bugs_probe_rule/<name>
		wantHas string
	}{
		{"every oracle-wrong probe resolves", "resolved", ""},
		{"an oracle-wrong probe with no adjudication row", "unresolved", `probe case "probe-b": no adjudication row`},
		{"an oracle-correct probe with no row is fine", "correct-unresolved", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bugs, adj, caseIDs := loadProbeRuleFixture(t, tt.fixture)
			err := CheckOracleBugs(caseIDs, adj, bugs)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckOracleBugs: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantHas) {
				t.Fatalf("CheckOracleBugs = %v, want a message containing %q", err, tt.wantHas)
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
