package matcher

// Compat selects, per ledgered defect class, whether the matcher
// reproduces pinned beets' defect or the adjudicated corrected
// behavior. A true field means "reproduce beets".
//
// Of corpus/analysis/oracle_bugs.json's oracle-wrong and
// no-correct-answer hypotheses, only bug-remix-parenthetical has a fix
// the matcher can express. bug-nonlatin-alias-selection needs an applied
// artist string and bug-assignment-rejection-gap a per-pair rejection,
// neither of which the matcher computes, so both are
// improvements-register items rather than switches nobody can flip.
type Compat struct {
	// RemixParenthetical reproduces bug-remix-parenthetical: the two
	// 0.3-weight bracket discounts.
	RemixParenthetical bool
}

// FullCompat returns the selection that reproduces every ledgered
// defect. It is a named engine surface rather than a bare struct
// literal in the harness because it IS a selection with a meaning: it
// is what internal/harness/conform's ModeCompat runs, the bit-identity
// claim against every committed record rests on it, and DefaultCompat
// is not its complement in any sense a reader should have to infer.
func FullCompat() Compat {
	return Compat{RemixParenthetical: true}
}

// DefaultCompat returns the selection corpus/analysis/oracle_bugs.json's
// own dispositions imply, as literals citing the hypothesis id per
// field. The engine never reads corpus/;
// internal/harness/conform cross-checks this value against the ledger,
// and that check is what fires the day a disposition moves and nobody
// moves this function.
func DefaultCompat() Compat {
	return Compat{
		RemixParenthetical: false, // bug-remix-parenthetical: oracle-wrong
	}
}

// Apply returns tun with c's string-heuristic selection installed in
// StringDistPatterns.
func (c Compat) Apply(tun Tunables) Tunables {
	if c.RemixParenthetical {
		tun.StringDistPatterns = beetsSDPatterns
		return tun
	}
	tun.StringDistPatterns = SDPatternsNoBracketDiscount
	return tun
}
