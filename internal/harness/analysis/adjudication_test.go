package analysis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
