package matcher

import (
	"slices"
	"strings"
	"testing"
)

func TestApplySelectsATableForEveryCombination(t *testing.T) {
	tests := []struct {
		name string
		c    Compat
		want []SDPattern
	}{
		{"both on reproduces beets", Compat{true, true}, BeetsSDPatterns},
		{"ft fixed only", Compat{false, true}, SDPatternsWordBoundary},
		{"remix fixed only", Compat{true, false}, SDPatternsNoBracketDiscount},
		{"both fixed", Compat{false, false}, SDPatternsGuardedNoBracketDiscount},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.c.Apply(DefaultTunables()).StringDistPatterns
			if !slices.Equal(got, tt.want) {
				t.Errorf("Apply().StringDistPatterns = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSDPatternsGuardedNoBracketDiscountIsBothFixesAndNothingElse(t *testing.T) {
	// Byte-identity against its two parents, so a silent edit to any
	// pattern literal here fails rather than quietly changing ten
	// committed defect_free figures. The guarded featuring pattern comes
	// from SDPatternsWordBoundary; every other row comes from
	// SDPatternsNoBracketDiscount, in that table's own order.
	want := slices.Clone(SDPatternsNoBracketDiscount)
	for i := range want {
		if strings.Contains(want[i].Pattern, "featuring") {
			want[i] = SDPatternsWordBoundary[2]
		}
	}
	if !slices.Equal(SDPatternsGuardedNoBracketDiscount, want) {
		t.Errorf("SDPatternsGuardedNoBracketDiscount = %v, want %v", SDPatternsGuardedNoBracketDiscount, want)
	}
}

func TestDefaultCompatDefaultsBothSwitchesOff(t *testing.T) {
	// Ruling A's scheduling rule: a hypothesis at oracle-wrong defaults
	// its switch OFF. Both string-heuristic hypotheses read oracle-wrong
	// in corpus/analysis/oracle_bugs.json as of 2026-09-05.
	// internal/harness/conform cross-checks this against the ledger; the
	// engine never reads corpus/.
	if got := (DefaultCompat()); got != (Compat{}) {
		t.Errorf("DefaultCompat = %+v, want both switches off", got)
	}
}

func TestFullCompatBothSwitchesOn(t *testing.T) {
	if got := FullCompat(); got != (Compat{FtPattern: true, RemixParenthetical: true}) {
		t.Errorf("FullCompat = %+v, want both switches on", got)
	}
}
