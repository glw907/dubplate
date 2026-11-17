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

// quietFallbacks and noneRecActions are the as_choice sets beets
// resolves import.quiet_fallback and import.none_rec_action against.
// They differ on purpose: a shared set would bless quiet_fallback:
// apply, which confuse raises on.
var (
	// as_choice({"skip": SKIP, "asis": ASIS})
	// (ui/commands/import_/session.py:377-379).
	quietFallbacks = []Action{ActionSkip, ActionAsIs}
	// as_choice({"skip": SKIP, "asis": ASIS, "ask": None})
	// (ui/commands/import_/session.py:383-389).
	noneRecActions = []Action{ActionSkip, ActionAsIs, ActionAsk}
)

// resolveAction resolves name, the value of config key key, against
// legal, the as_choice set beets accepts for that key, returning an
// error wrapping ErrInvalidTunable for any other name.
func resolveAction(key, name string, legal []Action) (Action, error) {
	if !slices.Contains(legal, Action(name)) {
		return "", fmt.Errorf("%s: %q: %w", key, name, ErrInvalidTunable)
	}
	return Action(name), nil
}

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
// The candidate's own key union drives the walk, and only the max_rec
// entries whose key is present in it are resolved (match.py:309-316):
// keys is set(min_dist.keys()), and the loop tests key in max_rec_view
// before calling as_choice, so an unparseable level on a key no
// candidate touches is never reached and never reported. tun.MaxRec is
// NOT walked in full: 3d does not own eager validation of every
// configured max_rec entry, only beets' own lazy resolution of the
// ones a candidate's keys select. keys is walked in sorted order so a
// failure is reproducible even though map iteration order is not.
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

	for _, key := range slices.Sorted(maps.Keys(keys)) {
		level, present := tun.MaxRec[key]
		if !present {
			continue
		}
		capped, err := ParseRecommendation(level)
		if err != nil {
			return 0, fmt.Errorf("max_rec[%q]: %w", key, err)
		}
		rec = min(rec, capped)
	}
	return rec, nil
}

// ActionFor is _summary_judgment
// (beets/ui/commands/import_/session.py:365-396).
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
		return resolveAction("import.quiet_fallback", tun.QuietFallback, quietFallbacks)
	}

	if tun.Timid {
		return ActionAsk, nil
	}

	if rec == RecNone {
		return resolveAction("import.none_rec_action", tun.NoneRecAction, noneRecActions)
	}

	return ActionAsk, nil
}
