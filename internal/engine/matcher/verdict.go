package matcher

import (
	"fmt"
	"maps"
	"slices"
)

// A Recommendation is beets' Recommendation (autotag/match.py:39-47).
// The values are ordered so that Go's min reproduces the max_rec
// downgrade's own min(rec, max_rec) (match.py:321).
type Recommendation int

const (
	RecNone Recommendation = iota
	RecLow
	RecMedium
	RecStrong
)

// String returns the name beets uses, which is also the name
// cassette.Expected.Recommendation carries.
func (r Recommendation) String() string {
	switch r {
	case RecNone:
		return "none"
	case RecLow:
		return "low"
	case RecMedium:
		return "medium"
	case RecStrong:
		return "strong"
	default:
		return fmt.Sprintf("Recommendation(%d)", int(r))
	}
}

// ParseRecommendation resolves one of the four names beets' as_choice
// accepts (match.py:313-320), returning an error wrapping
// ErrInvalidTunable for anything else.
func ParseRecommendation(s string) (Recommendation, error) {
	switch s {
	case "none":
		return RecNone, nil
	case "low":
		return RecLow, nil
	case "medium":
		return RecMedium, nil
	case "strong":
		return RecStrong, nil
	default:
		return 0, fmt.Errorf("recommendation %q: %w", s, ErrInvalidTunable)
	}
}

// An Action is the decision beets takes unprompted
// (beets/importer/actions.py:8-19), lowercased as the shim records it
// and with "ask" standing for beets' None return
// (derive_oracle.py:206).
type Action string

const (
	ActionApply Action = "apply"
	ActionSkip  Action = "skip"
	ActionAsIs  Action = "asis"
	ActionAsk   Action = "ask"
)

// ParseAction resolves an action name to an Action, returning an error
// wrapping ErrInvalidTunable for anything else. It is the shared lexer
// and never the membership check: beets accepts only skip and asis for
// import.quiet_fallback (session.py:341) and skip, asis and ask for
// import.none_rec_action (session.py:348-350), so Validate and
// ActionFor test membership per key against those two sets. A shared
// four-name set would bless quiet_fallback: apply, which confuse
// raises on (fold M-d).
func ParseAction(s string) (Action, error) {
	switch Action(s) {
	case ActionApply, ActionSkip, ActionAsIs, ActionAsk:
		return Action(s), nil
	default:
		return "", fmt.Errorf("action %q: %w", s, ErrInvalidTunable)
	}
}

// LegalQuietFallbacks and LegalNoneRecActions are those two sets, each
// citing the as_choice call it mirrors. They are exported so
// Tunables.Validate and ActionFor test the same membership and a
// reader can see that they differ on purpose.
var (
	// LegalQuietFallbacks mirrors as_choice({"skip": SKIP, "asis": ASIS})
	// (session.py:340-342).
	LegalQuietFallbacks = []Action{ActionSkip, ActionAsIs}
	// LegalNoneRecActions mirrors
	// as_choice({"skip": SKIP, "asis": ASIS, "ask": None})
	// (session.py:346-352).
	LegalNoneRecActions = []Action{ActionSkip, ActionAsIs, ActionAsk}
)

// Recommend is _recommendation (autotag/match.py:269-323). cands must be
// Candidates.Sorted's output; an empty slice returns RecNone.
//
// The three thresholds are asymmetric on purpose: strong compares with a
// strict < and medium with an inclusive <= (match.py:285, 288). That
// asymmetry is the whole reason corpus/analysis/proximity.json exists,
// so a case sitting exactly on a threshold admits no epsilon.
func Recommend(tun Tunables, cands []Candidate) (Recommendation, error) {
	if len(cands) == 0 {
		return RecNone, nil
	}

	minDist := cands[0].Dist.Distance()
	var rec Recommendation
	switch {
	case minDist < tun.StrongRecThresh:
		rec = RecStrong
	case minDist <= tun.MediumRecThresh:
		rec = RecMedium
	case len(cands) == 1:
		rec = RecLow
	case cands[1].Dist.Distance()-minDist >= tun.RecGapThresh:
		rec = RecLow
	default:
		// Return immediately. Can't be downgraded any further
		// (match.py:300-302, contract V13).
		return RecNone, nil
	}

	return downgrade(tun, cands[0].Dist, rec)
}

// downgrade applies match.py:304-322's min(rec, max_rec) ladder. The
// key set against which max_rec is APPLIED is the NONZERO view, unioned
// with every track sub-distance's own nonzero keys (contract V14): a
// key added at 0.0 never downgrades.
//
// Every entry of tun.MaxRec is resolved, not only the ones that match a
// key this candidate carries: 3d owns the name validation beets gets
// from confuse (contract V15), and confuse resolves a config subtree's
// choices once regardless of which entries a given call site consults.
// tun.MaxRec is walked in sorted key order so a failure is reproducible
// even though map iteration order is not, and so this loop never runs
// at all when the caller returned before the downgrade (contract V13):
// an unreachable bad entry is never reported.
func downgrade(tun Tunables, dist *Distance, rec Recommendation) (Recommendation, error) {
	keys := make(map[string]struct{})
	for _, key := range dist.NonzeroKeys() {
		keys[key] = struct{}{}
	}
	for _, track := range dist.Tracks {
		for _, key := range track.Dist.NonzeroKeys() {
			keys[key] = struct{}{}
		}
	}

	for _, key := range slices.Sorted(maps.Keys(tun.MaxRec)) {
		capped, err := ParseRecommendation(tun.MaxRec[key])
		if err != nil {
			return 0, fmt.Errorf("max_rec[%q]: %w", key, err)
		}
		if _, present := keys[key]; present {
			rec = min(rec, capped)
		}
	}
	return rec, nil
}

// ActionFor is _summary_judgment
// (beets/ui/commands/import_/session.py:328-360).
//
// The quiet branch is tested FIRST, so Timid and NoneRecAction are
// unreachable whenever Quiet is set. Every committed corpus record was
// derived with quiet forced true (derive_oracle.py:204), so the corpus
// grades two of this function's five outcomes and the rest are graded
// by the table below.
func ActionFor(tun Tunables, rec Recommendation) (Action, error) {
	if tun.Quiet {
		if rec == RecStrong {
			return ActionApply, nil
		}
		fallback, err := ParseAction(tun.QuietFallback)
		if err != nil {
			return "", err
		}
		if !slices.Contains(LegalQuietFallbacks, fallback) {
			return "", fmt.Errorf("import.quiet_fallback: %q: %w (session.py:340-342)", tun.QuietFallback, ErrInvalidTunable)
		}
		return fallback, nil
	}

	if tun.Timid {
		return ActionAsk, nil
	}

	if rec == RecNone {
		action, err := ParseAction(tun.NoneRecAction)
		if err != nil {
			return "", err
		}
		if !slices.Contains(LegalNoneRecActions, action) {
			return "", fmt.Errorf("import.none_rec_action: %q: %w (session.py:346-350)", tun.NoneRecAction, ErrInvalidTunable)
		}
		return action, nil
	}

	return ActionAsk, nil
}
