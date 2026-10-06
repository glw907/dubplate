package analysis

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// repoRoot walks up from the working directory to the nearest ancestor
// holding go.mod, matching corpus_test.go's own helper: `go test` runs
// from this package's directory, so corpus/ has to be resolved relative
// to the module root, not the CWD.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("locate repo root: no go.mod above %s", dir)
		}
		dir = parent
	}
}

func TestCheckAdjudications(t *testing.T) {
	base := AdjudicationEntry{
		ID:           "adj-1",
		Subject:      AdjudicationSubject{CaseID: "clean-match"},
		Suspected:    "the featuring discount tips the distance under threshold",
		Citation:     "U1/ft-pattern",
		ReleaseTruth: "the correct release is a different, unrelated title",
		Disposition:  Pending,
		Reason:       "authored this pass; awaiting Geoff's ruling",
		Adjudicator:  AgentAdjudicator,
		Date:         "2026-09-03",
	}

	noSubject := base
	noSubject.Subject = AdjudicationSubject{}

	unknownCase := base
	unknownCase.Subject = AdjudicationSubject{CaseID: "retired-case"}

	noSuspected := base
	noSuspected.Suspected = ""

	noReleaseTruth := base
	noReleaseTruth.ReleaseTruth = ""

	noReason := base
	noReason.Reason = ""

	noAdjudicator := base
	noAdjudicator.Adjudicator = ""

	noDate := base
	noDate.Date = ""

	badDisposition := base
	badDisposition.Disposition = "oracle-shrug"

	agentRulesOracleWrong := base
	agentRulesOracleWrong.Disposition = OracleWrong
	agentRulesOracleWrong.CorrectedVerdict = "medium"

	humanRulesOracleWrong := agentRulesOracleWrong
	humanRulesOracleWrong.Adjudicator = "Geoff"

	oracleWrongNoVerdict := base
	oracleWrongNoVerdict.Disposition = OracleWrong
	oracleWrongNoVerdict.Adjudicator = "Geoff"

	tests := []struct {
		name    string
		caseIDs []string
		entries []AdjudicationEntry
		wantHas string // empty means the ledger is accepted
	}{
		{"a well-formed pending row", []string{"clean-match"}, []AdjudicationEntry{base}, ""},
		{"a human ruling oracle-wrong with a corrected verdict", []string{"clean-match"}, []AdjudicationEntry{humanRulesOracleWrong}, ""},
		{"a subject naming nothing", []string{"clean-match"}, []AdjudicationEntry{noSubject}, "case_id missing"},
		{"a case id no committed case carries", []string{"clean-match"}, []AdjudicationEntry{unknownCase}, "no such committed case"},
		{"no written suspicion", []string{"clean-match"}, []AdjudicationEntry{noSuspected}, "suspected: missing"},
		{"no written release truth", []string{"clean-match"}, []AdjudicationEntry{noReleaseTruth}, "release_truth: missing"},
		{"no reason", []string{"clean-match"}, []AdjudicationEntry{noReason}, "reason: missing"},
		{"no adjudicator", []string{"clean-match"}, []AdjudicationEntry{noAdjudicator}, "adjudicator: missing"},
		{"no date", []string{"clean-match"}, []AdjudicationEntry{noDate}, "date: missing"},
		{"an unrecognized disposition", []string{"clean-match"}, []AdjudicationEntry{badDisposition}, "unrecognized"},
		{"an agent ruling oracle-wrong", []string{"clean-match"}, []AdjudicationEntry{agentRulesOracleWrong}, "only a human name may set anything but"},
		{"oracle-wrong with no corrected verdict", []string{"clean-match"}, []AdjudicationEntry{oracleWrongNoVerdict}, "needs a corrected_verdict"},
		{"an entry with no id", []string{"clean-match"}, []AdjudicationEntry{{Subject: base.Subject, Suspected: base.Suspected, Citation: base.Citation, ReleaseTruth: base.ReleaseTruth, Disposition: Pending, Reason: base.Reason, Adjudicator: base.Adjudicator, Date: base.Date}}, "entry with no id"},
		{"one id listed twice", []string{"clean-match"}, []AdjudicationEntry{base, base}, "listed twice"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAdjudications(tt.caseIDs, &AdjudicationLedger{Entries: tt.entries})
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckAdjudications: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckAdjudications accepted a ledger that is %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

// TestCheckAdjudications_DefectFreeWithinRowRules extends
// TestCheckAdjudications with the three rules checkAdjudicationEntry
// gains for the defect_free block (ruling A / M6): a block on a row
// whose disposition is not oracle-wrong, a block with an empty field,
// and a block whose distance sits outside [0, 1].
func TestCheckAdjudications_DefectFreeWithinRowRules(t *testing.T) {
	oracleWrong := AdjudicationEntry{
		ID:               "adj-1",
		Subject:          AdjudicationSubject{CaseID: "clean-match"},
		Suspected:        "the featuring discount tips the distance under threshold",
		Citation:         "U1/ft-pattern",
		ReleaseTruth:     "the correct release is a different, unrelated title",
		Disposition:      OracleWrong,
		CorrectedVerdict: "medium, skip; defect-free distance 0.101852",
		Reason:           "adjudicated",
		Adjudicator:      "Geoff",
		Date:             "2026-09-04",
	}
	goodBlock := DefectFree{Distance: 0.101851852, Recommendation: "medium", Action: "skip", PatternTable: "SDPatternsNoBracketDiscount"}

	withBlock := oracleWrong
	block := goodBlock
	withBlock.DefectFree = &block

	blockOnOracleCorrect := oracleWrong
	blockOnOracleCorrect.Disposition = OracleCorrect
	blockOnOracleCorrect.CorrectedVerdict = ""
	oracleCorrectBlock := goodBlock
	blockOnOracleCorrect.DefectFree = &oracleCorrectBlock

	emptyPatternTable := oracleWrong
	b1 := goodBlock
	b1.PatternTable = ""
	emptyPatternTable.DefectFree = &b1

	emptyRecommendation := oracleWrong
	b2 := goodBlock
	b2.Recommendation = ""
	emptyRecommendation.DefectFree = &b2

	emptyAction := oracleWrong
	b3 := goodBlock
	b3.Action = ""
	emptyAction.DefectFree = &b3

	distanceTooHigh := oracleWrong
	b4 := goodBlock
	b4.Distance = 1.5
	distanceTooHigh.DefectFree = &b4

	distanceNegative := oracleWrong
	b5 := goodBlock
	b5.Distance = -0.1
	distanceNegative.DefectFree = &b5

	tests := []struct {
		name    string
		entry   AdjudicationEntry
		wantHas string
	}{
		{"a well-formed defect_free block", withBlock, ""},
		{"a block on an oracle-correct row", blockOnOracleCorrect, `defect_free block on a row whose disposition is "oracle-correct"`},
		{"a block with pattern_table empty", emptyPatternTable, "defect_free: pattern_table: missing"},
		{"a block with recommendation empty", emptyRecommendation, "defect_free: recommendation: missing"},
		{"a block with action empty", emptyAction, "defect_free: action: missing"},
		{"a block with distance above 1", distanceTooHigh, "defect_free: distance 1.5 outside [0, 1]"},
		{"a block with distance below 0", distanceNegative, "defect_free: distance -0.1 outside [0, 1]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAdjudications([]string{"clean-match"}, &AdjudicationLedger{Entries: []AdjudicationEntry{tt.entry}})
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckAdjudications: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckAdjudications accepted a ledger that is %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

// TestCheckAdjudications_AppliedFieldWithinRowRules extends
// TestCheckAdjudications with R7's three rules over the applied_field
// block: legal only on a case-subject row, field and commit both
// required when the block is present, and the pre-existing
// agent-may-only-set-pending rule still fires when a block rides along
// with it (the malformed-row table's fourth case), proving the new
// block does not carve out an exception to that rule.
func TestCheckAdjudications_AppliedFieldWithinRowRules(t *testing.T) {
	base := AdjudicationEntry{
		ID:           "adj-1",
		Subject:      AdjudicationSubject{CaseID: "clean-match"},
		Suspected:    "tagio's read-back disagreed with the ledger on one field",
		Citation:     "rung 8 conformance instrument",
		ReleaseTruth: "the ledger's own value is the truth: rg_track_peak reads 0.9999999",
		Disposition:  Pending,
		Reason:       "staged automatically on a divergence; awaiting the owner's ruling",
		Adjudicator:  AgentAdjudicator,
		Date:         "2026-09-08",
	}
	goodBlock := AppliedField{ItemIndex: 0, Field: "rg_track_peak", LedgerValue: "0.9999999", ReadBackValue: "0.9999998", Commit: "abc1234"}

	withBlock := base
	block := goodBlock
	withBlock.AppliedField = &block

	emptyField := base
	b1 := goodBlock
	b1.Field = ""
	emptyField.AppliedField = &b1

	emptyCommit := base
	b2 := goodBlock
	b2.Commit = ""
	emptyCommit.AppliedField = &b2

	blockOnNonPendingAgentRow := base
	blockOnNonPendingAgentRow.Disposition = OracleWrong
	blockOnNonPendingAgentRow.CorrectedVerdict = "medium"
	b3 := goodBlock
	blockOnNonPendingAgentRow.AppliedField = &b3

	tests := []struct {
		name    string
		entry   AdjudicationEntry
		wantHas string
	}{
		{"a well-formed applied_field block", withBlock, ""},
		{"a block with field empty", emptyField, "applied_field: field: missing"},
		{"a block with commit empty", emptyCommit, "applied_field: commit: missing"},
		{"a block on a row whose disposition is not pending with the agent adjudicator", blockOnNonPendingAgentRow, "only a human name may set anything but"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckAdjudications([]string{"clean-match"}, &AdjudicationLedger{Entries: []AdjudicationEntry{tt.entry}})
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckAdjudications: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckAdjudications accepted a ledger that is %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

// TestAppliedFieldExampleDecodesAndPasses proves R7's committed
// example, internal/engine/tagio/testdata/divergence-example.json,
// decodes into AdjudicationEntry and passes CheckAdjudications on its
// own, with a caller-supplied caseIDs slice standing in for the corpus
// (CheckAdjudications resolves a case-subject row against the caller's
// list, adjudication.go:156-159 and :183-185, rather than against the
// corpus). This is the byte-identical literal
// internal/engine/tagio grades its own emitter against.
func TestAppliedFieldExampleDecodesAndPasses(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "internal", "engine", "tagio", "testdata", "divergence-example.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read example at %s: %v", path, err)
	}

	var entry AdjudicationEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("decode example at %s: %v", path, err)
	}
	if entry.AppliedField == nil {
		t.Fatalf("example at %s decoded with no applied_field block", path)
	}

	if err := CheckAdjudications([]string{entry.Subject.CaseID}, &AdjudicationLedger{Entries: []AdjudicationEntry{entry}}); err != nil {
		t.Errorf("CheckAdjudications rejected the committed example: %v", err)
	}
}

// loadDefectFreeFixture loads the hand-authored oracle-bug and
// adjudication ledgers under testdata/defect_free: two hypotheses (the
// StringHeuristicHypotheses member and bug-assignment-rejection-gap,
// which is oracle-wrong but not a member) resolving to five
// adjudication rows, invented rather than drawn from the committed
// corpus per the fixture discipline.
func loadDefectFreeFixture(t *testing.T) (*OracleBugLedger, *AdjudicationLedger) {
	t.Helper()
	bugs, err := LoadOracleBugs("testdata/defect_free/oracle_bugs.json")
	if err != nil {
		t.Fatalf("LoadOracleBugs: %v", err)
	}
	adj, err := LoadAdjudications("testdata/defect_free/adjudications.json")
	if err != nil {
		t.Fatalf("LoadAdjudications: %v", err)
	}
	return bugs, adj
}

// TestCheckDefectFree proves the cross-ledger half: a probe case
// missing its block, an illegal pattern_table value, a categorical
// mismatch between the block and its row's corrected_verdict, and the
// forbidden direction (fold, m12), a block on a row whose case is not a
// probe of a StringHeuristicHypotheses hypothesis.
func TestCheckDefectFree(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*OracleBugLedger, *AdjudicationLedger)
		wantHas string
	}{
		{"a well-formed pair of ledgers", func(*OracleBugLedger, *AdjudicationLedger) {}, ""},
		{
			name: "a probe case's defect_free block deleted",
			mutate: func(_ *OracleBugLedger, adj *AdjudicationLedger) {
				for i := range adj.Entries {
					if adj.Entries[i].ID == "adj-remix-fixture-a" {
						adj.Entries[i].DefectFree = nil
					}
				}
			},
			wantHas: `probe case "remix-fixture-a" (bug-remix-parenthetical): adjudication row adj-remix-fixture-a carries no defect_free block`,
		},
		{
			name: "an illegal pattern_table value",
			mutate: func(_ *OracleBugLedger, adj *AdjudicationLedger) {
				for i := range adj.Entries {
					if adj.Entries[i].ID == "adj-remix-fixture-a" {
						adj.Entries[i].DefectFree.PatternTable = "BeetsSDPatterns"
					}
				}
			},
			wantHas: `pattern_table "BeetsSDPatterns" is not a legal value`,
		},
		{
			name: "a recommendation that does not appear in corrected_verdict",
			mutate: func(_ *OracleBugLedger, adj *AdjudicationLedger) {
				for i := range adj.Entries {
					if adj.Entries[i].ID == "adj-remix-fixture-a" {
						adj.Entries[i].DefectFree.Recommendation = "strong"
					}
				}
			},
			wantHas: `recommendation "strong" does not appear verbatim in corrected_verdict`,
		},
		{
			name: "a defect_free block on a case that is not a string-heuristic probe",
			mutate: func(_ *OracleBugLedger, adj *AdjudicationLedger) {
				for i := range adj.Entries {
					if adj.Entries[i].ID == "adj-assign-fixture-d" {
						adj.Entries[i].DefectFree = &DefectFree{Distance: 0.05, Recommendation: "medium", Action: "skip", PatternTable: "SDPatternsNoBracketDiscount"}
					}
				}
			},
			wantHas: `adj-assign-fixture-d: carries a defect_free block, but its case "assign-fixture-d" is not a probe`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bugs, adj := loadDefectFreeFixture(t)
			tt.mutate(bugs, adj)
			err := CheckDefectFree(bugs, adj)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckDefectFree: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckDefectFree accepted a pair of ledgers where %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not contain %q", err, tt.wantHas)
			}
		})
	}
}

// TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost proves
// AdjudicationEntry's field set matches corpus/analysis/adjudications.json
// exactly: decoding the file into AdjudicationLedger and re-encoding
// must not gain or lose a key. Byte-for-byte identity is impossible
// (fold B2): the committed key order puts corrected_verdict after
// date while the struct declares it before Reason, and the file's
// literal angle brackets get escaped by Go's encoder unless
// SetEscapeHTML(false) is used. So the comparison decodes both the
// original bytes and the re-encoded bytes into map[string]any trees
// and compares those, order-independent.
func TestTheCommittedLedgerRoundTripsWithNoKeyGainedOrLost(t *testing.T) {
	root := repoRoot(t)
	path := filepath.Join(root, "corpus", "analysis", "adjudications.json")
	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed ledger: %v", err)
	}

	var ledger AdjudicationLedger
	if err := json.Unmarshal(original, &ledger); err != nil {
		t.Fatalf("decode committed ledger: %v", err)
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(&ledger); err != nil {
		t.Fatalf("re-encode ledger: %v", err)
	}
	reencoded := buf.Bytes()

	var wantTree, gotTree map[string]any
	if err := json.Unmarshal(original, &wantTree); err != nil {
		t.Fatalf("decode original as tree: %v", err)
	}
	if err := json.Unmarshal(reencoded, &gotTree); err != nil {
		t.Fatalf("decode re-encoded bytes as tree: %v", err)
	}
	if !reflect.DeepEqual(wantTree, gotTree) {
		t.Fatalf("re-encoded ledger tree does not match the original: gained or lost a key")
	}
}

func TestLoadAdjudicationsRejectsAnInventedKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "adjudications.json")
	if err := os.WriteFile(path, []byte(`{"entries":[],"schema_version":"1"}`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if _, err := LoadAdjudications(path); err == nil {
		t.Fatal("LoadAdjudications accepted a key nothing parses")
	}
}

// TestLoadAdjudicationsRefusesAFieldTheLedgerNoLongerCarries proves a
// row carrying a removed field fails at decode, naming the field, so a
// stale row cannot load as if the field still meant something.
func TestLoadAdjudicationsRefusesAFieldTheLedgerNoLongerCarries(t *testing.T) {
	tests := []struct {
		name string
		row  string
		want string
	}{
		{"album in the subject", `{"id":"a","subject":{"case_id":"c","album":"Some Album"}}`, `"album"`},
		{"release_mbid in the subject", `{"id":"a","subject":{"case_id":"c","release_mbid":"aaaa-bbbb"}}`, `"release_mbid"`},
		{"matcher_commit on the row", `{"id":"a","subject":{"case_id":"c"},"matcher_commit":"abc1234"}`, `"matcher_commit"`},
		{"matcher_config_hash on the row", `{"id":"a","subject":{"case_id":"c"},"matcher_config_hash":"deadbeef"}`, `"matcher_config_hash"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "adjudications.json")
			if err := os.WriteFile(path, []byte(`{"entries":[`+tt.row+`]}`), 0o644); err != nil {
				t.Fatalf("write fixture: %v", err)
			}
			_, err := LoadAdjudications(path)
			if err == nil {
				t.Fatalf("LoadAdjudications accepted a row carrying %s", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q does not name %s", err, tt.want)
			}
		})
	}
}
