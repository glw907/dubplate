package matcher

// Compat selects, per ledgered defect class, whether the matcher
// reproduces pinned beets' defect or the adjudicated corrected
// behavior. A true field means "reproduce beets".
//
// Only bug-remix-parenthetical, of corpus/analysis/oracle_bugs.json's
// hypotheses, has a fix the matcher can express. The other two need an
// applied artist string or a per-pair rejection the matcher does not
// compute, so they are improvements-register items rather than switches
// nobody can flip.
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
