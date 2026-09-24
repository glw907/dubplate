package matcher

import (
	"slices"
	"testing"
)

func TestApplySelectsATableForEveryCombination(t *testing.T) {
	tests := []struct {
		name string
		c    Compat
		want []SDPattern
	}{
		{"on reproduces beets", Compat{RemixParenthetical: true}, BeetsSDPatterns},
		{"remix fixed", Compat{RemixParenthetical: false}, SDPatternsNoBracketDiscount},
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

func TestDefaultCompatDefaultsTheSwitchOff(t *testing.T) {
	// A hypothesis at oracle-wrong defaults its switch OFF, and
	// bug-remix-parenthetical reads oracle-wrong in
	// corpus/analysis/oracle_bugs.json. internal/harness/conform
	// cross-checks this against the ledger; the engine never reads
	// corpus/.
	if got := DefaultCompat(); got != (Compat{}) {
		t.Errorf("DefaultCompat = %+v, want the switch off", got)
	}
}

func TestFullCompatSwitchOn(t *testing.T) {
	if got := FullCompat(); got != (Compat{RemixParenthetical: true}) {
		t.Errorf("FullCompat = %+v, want the switch on", got)
	}
}
