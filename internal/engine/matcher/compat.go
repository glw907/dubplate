package matcher

// Compat selects, per ledgered defect class, whether the matcher
// reproduces pinned beets' defect or the adjudicated corrected
// behavior. A true field means "reproduce beets".
//
// Two classes and no more. corpus/analysis/oracle_bugs.json carries
// three oracle-wrong hypotheses and one disposed no-correct-answer
// (bug-assignment-rejection-gap, owner ruling 2026-09-07), and only
// these two have a fix a port can express: bug-nonlatin-alias-selection's
// corrected verdict is an applied artist string rung 3 computes none of,
// and bug-assignment-rejection-gap needs a per-pair rejection mechanism
// beets does not have at all, which is why it carries no correct answer
// for the oracle to have produced rather than a corrected verdict rung 3
// merely fails to express. Both are recorded as improvements-register
// items rather than shipped as switches nobody can flip.
type Compat struct {
	// FtPattern reproduces bug-ft-pattern: SD_PATTERNS' featuring
	// pattern matching "ft" inside an ordinary word such as "Draft".
	FtPattern bool
	// RemixParenthetical reproduces bug-remix-parenthetical: the two
	// 0.3-weight bracket discounts.
	RemixParenthetical bool
}

// FullCompat returns the selection that reproduces every ledgered
// defect, both switches on. It is a named engine surface rather than a
// bare struct literal in the harness because it IS a selection with a
// meaning: it is what internal/harness/conform's ModeCompat runs, the
// bit-identity claim against every committed record rests on it, and
// DefaultCompat is not its complement in any sense a reader should have
// to infer (fold m6).
func FullCompat() Compat {
	return Compat{FtPattern: true, RemixParenthetical: true}
}

// DefaultCompat returns the selection corpus/analysis/oracle_bugs.json's
// own dispositions imply, as literals citing the hypothesis id per
// field. The engine never reads corpus/;
// internal/harness/conform cross-checks these two values against the
// ledger, and that check is what fires the day a disposition moves and
// nobody moves this function.
func DefaultCompat() Compat {
	return Compat{
		FtPattern:          false, // bug-ft-pattern: oracle-wrong
		RemixParenthetical: false, // bug-remix-parenthetical: oracle-wrong
	}
}

// Apply returns tun with c's string-heuristic selection installed in
// StringDistPatterns. All four combinations name a real table, which is
// why SDPatternsGuardedNoBracketDiscount exists.
func (c Compat) Apply(tun Tunables) Tunables {
	switch {
	case c.FtPattern && c.RemixParenthetical:
		tun.StringDistPatterns = BeetsSDPatterns
	case c.RemixParenthetical: // FtPattern fixed alone
		tun.StringDistPatterns = SDPatternsWordBoundary
	case c.FtPattern: // RemixParenthetical fixed alone
		tun.StringDistPatterns = SDPatternsNoBracketDiscount
	default: // both fixed
		tun.StringDistPatterns = SDPatternsGuardedNoBracketDiscount
	}
	return tun
}
