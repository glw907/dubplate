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

	noReason := tied
	noReason.Disposition = TieExcluded

	claimsUniqueness := tied
	claimsUniqueness.Unique = true

	inconsistentCosts := unique
	inconsistentCosts.Difference = 0.25

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
