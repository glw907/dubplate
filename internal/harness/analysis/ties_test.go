package analysis

import (
	"strings"
	"testing"
)

func TestCheckTieLedger(t *testing.T) {
	unique := TieEntry{CaseID: "clean", AlbumID: "aaaa", OptimalCost: 0.5, BestForbiddenCost: 1.5, Difference: 1.0, Unique: true}
	tied := TieEntry{CaseID: "clean", AlbumID: "aaaa", OptimalCost: 0.5, BestForbiddenCost: 0.5, Difference: 0.0}

	dispositioned := tied
	dispositioned.Disposition = TiePinned
	dispositioned.Reason = "two identical track titles; the recorded mapping is the expectation"
	dispositioned.Adjudicator = AgentAdjudicator

	noReason := tied
	noReason.Disposition = TieExcluded

	claimsUniqueness := tied
	claimsUniqueness.Unique = true

	inconsistentCosts := unique
	inconsistentCosts.Difference = 0.25

	skipped := TieEntry{CaseID: "no-candidate", Status: TieSkipped, Reason: "record carries no candidate to measure"}

	skippedNoReason := skipped
	skippedNoReason.Reason = ""

	skippedWithAlbumID := skipped
	skippedWithAlbumID.AlbumID = "aaaa"

	skippedWithDisposition := skipped
	skippedWithDisposition.Disposition = TiePinned

	excludedByAgent := tied
	excludedByAgent.Disposition, excludedByAgent.Reason, excludedByAgent.Adjudicator = TieExcluded, "why", AgentAdjudicator

	excludedByHuman := excludedByAgent
	excludedByHuman.Adjudicator = "Geoff"

	pinnedByAgent := tied
	pinnedByAgent.Disposition, pinnedByAgent.Reason, pinnedByAgent.Adjudicator = TiePinned, "why", AgentAdjudicator

	pinnedNoAdjudicator := tied
	pinnedNoAdjudicator.Disposition, pinnedNoAdjudicator.Reason = TiePinned, "why"

	skippedWithAdjudicator := skipped
	skippedWithAdjudicator.Adjudicator = "Geoff"

	tests := []struct {
		name    string
		caseIDs []string
		ledger  TieLedger
		wantHas string // empty means the ledger is accepted
	}{
		{
			name:    "unique optimum needs no disposition",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{unique}},
		},
		{
			name:    "tie dispositioned with a reason",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{dispositioned}},
		},
		{
			name:    "tie with no disposition",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{tied}},
			wantHas: "disposition must be",
		},
		{
			name:    "tie dispositioned with no reason",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{noReason}},
			wantHas: "no reason",
		},
		{
			name:    "a row claiming a uniqueness its own costs deny",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{claimsUniqueness}},
			wantHas: "unique = true",
		},
		{
			name:    "a difference that is not the recorded costs' gap",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{inconsistentCosts}},
			wantHas: "recorded costs differ by",
		},
		{
			name:    "a corpus case with no row",
			caseIDs: []string{"clean", "assign-scrambled"},
			ledger:  TieLedger{Entries: []TieEntry{unique}},
			wantHas: "assign-scrambled: missing",
		},
		{
			name:    "a row naming no corpus case",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{unique, {CaseID: "retired", AlbumID: "bbbb", Unique: true}}},
			wantHas: "no such corpus case",
		},
		{
			name:    "one case listed twice",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{unique, unique}},
			wantHas: "listed twice",
		},
		{
			name:    "a skipped row with a written reason",
			caseIDs: []string{"no-candidate"},
			ledger:  TieLedger{Entries: []TieEntry{skipped}},
		},
		{
			name:    "a skipped row with no reason",
			caseIDs: []string{"no-candidate"},
			ledger:  TieLedger{Entries: []TieEntry{skippedNoReason}},
			wantHas: "skipped with no reason",
		},
		{
			name:    "a skipped row still carrying an album_id",
			caseIDs: []string{"no-candidate"},
			ledger:  TieLedger{Entries: []TieEntry{skippedWithAlbumID}},
			wantHas: "skipped, but carries album_id",
		},
		{
			name:    "a skipped row still carrying a disposition",
			caseIDs: []string{"no-candidate"},
			ledger:  TieLedger{Entries: []TieEntry{skippedWithDisposition}},
			wantHas: "skipped, but carries disposition",
		},
		{
			name:    "a skipped case with no row at all is still missing",
			caseIDs: []string{"no-candidate"},
			ledger:  TieLedger{},
			wantHas: "no-candidate: missing",
		},
		{
			name:    "an excluded row adjudicated by the agent sentinel",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{excludedByAgent}},
			wantHas: "only a human name may exclude",
		},
		{
			name:    "an excluded row adjudicated by a human",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{excludedByHuman}},
		},
		{
			// The asymmetry's own proof: pinned keeps the case graded, so an
			// agent may set it, and the two committed pinned rows do.
			name:    "a pinned row adjudicated by the agent sentinel",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{pinnedByAgent}},
		},
		{
			name:    "a dispositioned row with no adjudicator",
			caseIDs: []string{"clean"},
			ledger:  TieLedger{Entries: []TieEntry{pinnedNoAdjudicator}},
			wantHas: "adjudicator: missing",
		},
		{
			name:    "a skipped row carrying an adjudicator",
			caseIDs: []string{"no-candidate"},
			ledger:  TieLedger{Entries: []TieEntry{skippedWithAdjudicator}},
			wantHas: "skipped, but carries adjudicator",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckTieLedger(tt.caseIDs, &tt.ledger)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckTieLedger: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckTieLedger accepted a ledger with %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

func TestMergeTieDispositionsCarriesEveryHandWrittenField(t *testing.T) {
	// derive_ties.py emits six measured fields and never disposition,
	// reason or adjudicator, so a fresh derivation's rows carry all
	// three empty. This is the whole reason make derive does not erase
	// the ledger, and both sites that would otherwise prove it are
	// //go:build oracle.
	derived := []TieEntry{
		{CaseID: "kept", AlbumID: "a", OptimalCost: 1, BestForbiddenCost: 1, Unique: false},
		{CaseID: "fresh", AlbumID: "b", OptimalCost: 1, BestForbiddenCost: 2, Difference: 1, Unique: true},
	}
	prior := &TieLedger{Entries: []TieEntry{
		{CaseID: "kept", Disposition: TiePinned, Reason: "both optima agree", Adjudicator: AgentAdjudicator},
		{CaseID: "gone", Disposition: TieExcluded, Reason: "retired", Adjudicator: "Geoff"},
	}}
	got := MergeTieDispositions(derived, prior)
	if len(got) != 2 {
		t.Fatalf("MergeTieDispositions returned %d rows, want 2: it carries fields, never rows", len(got))
	}
	if got[0].Disposition != TiePinned || got[0].Reason == "" || got[0].Adjudicator != AgentAdjudicator {
		t.Errorf("row %q = %+v, want disposition, reason and adjudicator all carried", got[0].CaseID, got[0])
	}
	if got[1].Disposition != "" || got[1].Adjudicator != "" {
		t.Errorf("row %q = %+v, want nothing carried: no prior row names it", got[1].CaseID, got[1])
	}
	if got[0].OptimalCost != 1 || got[0].BestForbiddenCost != 1 {
		t.Errorf("row %q measured fields were overwritten: %+v", got[0].CaseID, got[0])
	}
}
