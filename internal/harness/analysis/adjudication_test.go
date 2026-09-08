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

	shadowSubject := base
	shadowSubject.Subject = AdjudicationSubject{Album: "Some Album", ReleaseMBID: "aaaa-bbbb"}
	shadowSubject.MatcherCommit = "abc1234"

	shadowSubjectNoCommit := base
	shadowSubjectNoCommit.Subject = AdjudicationSubject{Album: "Some Album", ReleaseMBID: "aaaa-bbbb"}

	caseSubjectWithMatcherFields := base
	caseSubjectWithMatcherFields.MatcherCommit = "abc1234"

	caseSubjectWithConfigHashOnly := base
	caseSubjectWithConfigHashOnly.MatcherConfigHash = "deadbeef"

	bothSubjects := base
	bothSubjects.Subject = AdjudicationSubject{CaseID: "clean-match", Album: "Some Album"}

	halfShadowSubject := base
	halfShadowSubject.Subject = AdjudicationSubject{Album: "Some Album"}

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
		{"a well-formed shadow-subject row", nil, []AdjudicationEntry{shadowSubject}, ""},
		{"a shadow-subject row with no matcher_commit", nil, []AdjudicationEntry{shadowSubjectNoCommit}, "subject is a shadow row, but matcher_commit is missing"},
		{"a case-subject row carrying matcher_commit", []string{"clean-match"}, []AdjudicationEntry{caseSubjectWithMatcherFields}, "subject is a case row, but carries matcher_commit or matcher_config_hash"},
		{"a case-subject row carrying matcher_config_hash only", []string{"clean-match"}, []AdjudicationEntry{caseSubjectWithConfigHashOnly}, "subject is a case row, but carries matcher_commit or matcher_config_hash"},
		{"a human ruling oracle-wrong with a corrected verdict", []string{"clean-match"}, []AdjudicationEntry{humanRulesOracleWrong}, ""},
		{"a subject naming both a case id and an album", []string{"clean-match"}, []AdjudicationEntry{bothSubjects}, "case_id and album/release_mbid both set"},
		{"a shadow subject missing release_mbid", nil, []AdjudicationEntry{halfShadowSubject}, "album and release_mbid must both be set"},
		{"a subject naming nothing", []string{"clean-match"}, []AdjudicationEntry{noSubject}, "names neither a case id"},
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
	goodBlock := DefectFree{Distance: 0.101851852, Recommendation: "medium", Action: "skip", PatternTable: "SDPatternsWordBoundary"}

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

// loadDefectFreeFixture loads the hand-authored oracle-bug and
// adjudication ledgers under testdata/defect_free: three hypotheses (two
// StringHeuristicHypotheses members and bug-assignment-rejection-gap,
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
					if adj.Entries[i].ID == "adj-ft-fixture-a" {
						adj.Entries[i].DefectFree = nil
					}
				}
			},
			wantHas: `probe case "ft-fixture-a" (bug-ft-pattern): adjudication row adj-ft-fixture-a carries no defect_free block`,
		},
		{
			name: "an illegal pattern_table value",
			mutate: func(_ *OracleBugLedger, adj *AdjudicationLedger) {
				for i := range adj.Entries {
					if adj.Entries[i].ID == "adj-ft-fixture-a" {
						adj.Entries[i].DefectFree.PatternTable = "BeetsSDPatterns"
					}
				}
			},
			wantHas: `pattern_table "BeetsSDPatterns" is not one of the two legal values`,
		},
		{
			name: "a recommendation that does not appear in corrected_verdict",
			mutate: func(_ *OracleBugLedger, adj *AdjudicationLedger) {
				for i := range adj.Entries {
					if adj.Entries[i].ID == "adj-ft-fixture-a" {
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
						adj.Entries[i].DefectFree = &DefectFree{Distance: 0.05, Recommendation: "medium", Action: "skip", PatternTable: "SDPatternsWordBoundary"}
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
// and compares those, order-independent. It also asserts the raw
// re-encoded bytes carry neither "matcher_commit" nor
// "matcher_config_hash" anywhere, which is the real detector for
// omitempty: every committed row is a case-subject row, so both
// fields are always empty, and a missing omitempty tag would still
// pass a field-set comparison but would print the key with an empty
// value.
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

	if bytes.Contains(reencoded, []byte("matcher_commit")) {
		t.Error("re-encoded ledger carries matcher_commit, but every committed row is a case-subject row")
	}
	if bytes.Contains(reencoded, []byte("matcher_config_hash")) {
		t.Error("re-encoded ledger carries matcher_config_hash, but every committed row is a case-subject row")
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
