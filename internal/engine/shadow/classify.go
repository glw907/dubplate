package shadow

import (
	"errors"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
	"github.com/glw907/dubplate/internal/engine/tag"
)

// classify returns the divergence class for one reconciled observation,
// or the reason it defers. Exactly one of the two is ever set.
//
// The oracle is binary (contract S3), so the comparable quantity is a
// binary projection of Proposal.Action and nothing finer: RecStrong is
// exactly attested, and RecMedium, RecLow and RecNone are
// indistinguishable to the oracle.
//
// NO ERROR OTHER THAN tag.ErrNoCandidates MAY EVER BE RECORDED AS A GO
// SKIP. A truncated album written as go-skips-oracle-applied is a
// fabricated divergence in the one artifact whose whole value is that
// it is honest, and it consumes exactly the owner adjudication ruling
// B's stamp exists to save. Every other error defers, and an error this
// function does not recognize defers with DeferUnclassified rather than
// falling into a verdict arm: a matcher revision that adds a sentinel
// gets a counted deferral, and the counter is what makes the omission
// visible.
//
// Every branch tests with errors.Is, never equality, because tag.Album
// wraps every client error in a message (tag.go:128, :170).
func classify(oracle OracleSignal, p *tag.Proposal, err error) (class, deferred string) {
	switch {
	case errors.Is(err, musicbrainz.ErrUnavailable):
		return "", DeferUnavailable
	case errors.Is(err, musicbrainz.ErrBudgetExhausted):
		return "", DeferBudgetExhausted
	case errors.Is(err, musicbrainz.ErrLimiterClosed):
		return "", DeferLimiterClosed
	case errors.Is(err, tag.ErrNoCandidates):
		return compareVerdicts(oracle, false, ""), ""
	case err != nil:
		return "", DeferUnclassified
	case p == nil || len(p.Candidates) == 0:
		// tag.Album returns a proposal with Candidates[0] well defined on
		// both flow paths whenever it returns no error (contract S1), so
		// this shape is unreachable through the flow the shadow drives.
		// It defers rather than reading the empty slice as a skip,
		// because a skip here would be a divergence nothing observed.
		return "", DeferUnclassified
	case p.Action != matcher.ActionApply:
		return compareVerdicts(oracle, false, ""), ""
	default:
		return compareVerdicts(oracle, true, p.Candidates[0].Info.AlbumID), ""
	}
}

// classifyVerdict is the same table over a stored GoVerdict rather than
// a live proposal, which is what the DefaultCompat re-score's annotation
// reads: a pending row states whether the second verdict agrees with the
// oracle where the first does not.
func classifyVerdict(oracle OracleSignal, v *GoVerdict) string {
	applies := v.Action == string(matcher.ActionApply)
	return compareVerdicts(oracle, applies, v.AlbumID)
}

// compareVerdicts is contract S4's table itself: two binary signals and,
// when both applied, the release each side named.
func compareVerdicts(oracle OracleSignal, goApplies bool, goRelease string) string {
	oracleApplies := oracle.Signal == "apply"
	switch {
	case !oracleApplies && !goApplies:
		return ClassAgreement
	case !oracleApplies:
		return ClassGoAppliesOracleSkipped
	case !goApplies:
		return ClassGoSkipsOracleApplied
	case goRelease == oracle.ReleaseMBID:
		return ClassAgreement
	default:
		return ClassGoAppliesDifferentRelease
	}
}

// isDivergence reports whether class is one of contract S4's three
// divergences, which is what "news" means: an agreement stages no row,
// and neither does a deferral, which carries no class at all.
func isDivergence(class string) bool {
	switch class {
	case ClassGoAppliesOracleSkipped, ClassGoSkipsOracleApplied, ClassGoAppliesDifferentRelease:
		return true
	default:
		return false
	}
}
