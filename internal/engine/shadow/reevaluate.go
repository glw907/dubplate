package shadow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/glw907/dubplate/internal/engine/musicbrainz"
	"github.com/glw907/dubplate/internal/engine/tag"
)

// A ReevaluateResult names what moved, by observation id, so the verb
// prints a record rather than a count.
type ReevaluateResult struct {
	Retired       []string
	Restamped     []string
	NewlyDiverged []string
	Unevaluable   []string
	Skipped       []string // already at the current commit
}

// A ResolvedRecord is resolved/<obs-id>.json: a divergence Reevaluate
// retired, naming both commits, both classes and both verdict pairs, so
// a later reader sees exactly what changed and under which build
// without re-running anything. It is the engine's own type and needs no
// harness import.
type ResolvedRecord struct {
	ID             string      `json:"id"`
	OriginalCommit string      `json:"original_commit"`
	RetiringCommit string      `json:"retiring_commit"`
	OriginalClass  string      `json:"original_class"`
	NewClass       string      `json:"new_class"`
	OriginalGo     *GoVerdicts `json:"original_go"`
	NewGo          *GoVerdicts `json:"new_go"`
	Date           string      `json:"date"`
}

// Reevaluate re-runs every observation stamped with a commit other than
// cfg.MatcherCommit against that observation's own stored responses.
//
// It spends no MusicBrainz budget, reaches no network, and reads no
// clock: the replay client takes musicbrainz.Unmetered()
// (limiter.go:354-359), the transport answers only from disk, and each
// observation is re-run at its OWN stored eval date, so a re-evaluation
// reproduces the original run exactly rather than re-scoring it against
// today. now is used only to date the resolved/ record it writes.
//
// Only a classified observation (an agreement or one of the three
// divergences) is in scope: a deferred, unreconciled, excluded or
// superseded row carries no class and no verdict to re-run against, so
// it is left exactly as it is.
//
// Four outcomes, and the fourth is why the replay transport refuses a
// miss rather than answering 404:
//
//   - RETIRED: the divergence no longer reproduces. resolved/<obs-id>.json
//     records the retiring commit and both verdicts, and the staged
//     pending row is removed.
//   - RESTAMPED: the divergence still reproduces, so the row is
//     restamped with the current commit and stays staged.
//   - NEWLY DIVERGED: an observation that previously agreed now
//     diverges. That is evidence the launch bar wants, and a
//     retire-only design would silently lose it. The new pending row
//     itself is staged by the caller, which is the one artifact this
//     package may not marshal (contract S19); the observation carries
//     everything PendingRowFor needs.
//   - UNEVALUABLE: the replay asked for a request the recording does not
//     carry. The row stays stamped and staged. A newer matcher revision
//     can take a different branch through tag.Album for the same input,
//     exactly as a different Compat can, and a silently retired
//     divergence is the one outcome worse than an un-retired one.
//
// An agreement that re-runs as an agreement is neither news nor a
// staged row, so it is left untouched rather than restamped.
func Reevaluate(ctx context.Context, cfg Config, s *Store, now time.Time) (*ReevaluateResult, error) {
	obsList, err := s.Observations()
	if err != nil {
		return nil, fmt.Errorf("reevaluate: %w", err)
	}

	result := &ReevaluateResult{}
	for _, obs := range obsList {
		if obs.MatcherCommit == cfg.MatcherCommit {
			result.Skipped = append(result.Skipped, obs.ID)
			continue
		}
		if obs.Class == "" || obs.Oracle == nil {
			continue
		}

		newGo, unevaluable, err := reevaluateAlbum(ctx, s, obs)
		if err != nil {
			return nil, fmt.Errorf("reevaluate %s: %w", obs.ID, err)
		}
		if unevaluable {
			result.Unevaluable = append(result.Unevaluable, obs.ID)
			continue
		}

		originalClass := obs.Class
		originalGo := obs.Go
		originalCommit := obs.MatcherCommit
		newClass := reclassify(*obs.Oracle, newGo.Compat)

		switch {
		case isDivergence(originalClass) && newClass == ClassAgreement:
			record := &ResolvedRecord{
				ID:             obs.ID,
				OriginalCommit: originalCommit,
				RetiringCommit: cfg.MatcherCommit,
				OriginalClass:  originalClass,
				NewClass:       newClass,
				OriginalGo:     originalGo,
				NewGo:          newGo,
				Date:           now.Format(time.DateOnly),
			}
			data, err := json.MarshalIndent(record, "", "  ")
			if err != nil {
				return nil, fmt.Errorf("reevaluate %s: %w", obs.ID, err)
			}
			if err := s.PutResolved(obs.ID, data); err != nil {
				return nil, fmt.Errorf("reevaluate %s: %w", obs.ID, err)
			}
			if err := s.RemovePending(obs.ID); err != nil {
				return nil, fmt.Errorf("reevaluate %s: %w", obs.ID, err)
			}
			// The observation itself is restamped to the retiring commit
			// and reclassified as an agreement: otherwise it keeps its
			// superseded commit and its divergence class forever, so a
			// later Reevaluate re-processes it (never lands in Skipped)
			// and BuildReport's per-commit row keeps counting a
			// divergence that resolved/ already says is retired.
			obs.MatcherCommit = cfg.MatcherCommit
			obs.Class = newClass
			obs.Go = newGo
			if err := s.PutObservation(obs); err != nil {
				return nil, fmt.Errorf("reevaluate %s: %w", obs.ID, err)
			}
			result.Retired = append(result.Retired, obs.ID)
		case isDivergence(originalClass):
			obs.MatcherCommit = cfg.MatcherCommit
			obs.Class = newClass
			obs.Go = newGo
			if err := s.PutObservation(obs); err != nil {
				return nil, fmt.Errorf("reevaluate %s: %w", obs.ID, err)
			}
			result.Restamped = append(result.Restamped, obs.ID)
		case isDivergence(newClass):
			obs.MatcherCommit = cfg.MatcherCommit
			obs.Class = newClass
			obs.Go = newGo
			if err := s.PutObservation(obs); err != nil {
				return nil, fmt.Errorf("reevaluate %s: %w", obs.ID, err)
			}
			result.NewlyDiverged = append(result.NewlyDiverged, obs.ID)
		}
		// An agreement that re-ran as an agreement falls through every
		// case above and is left exactly as it was.
	}
	return result, nil
}

// reclassify is classifyVerdict extended to a nil verdict, which is
// what reevaluateAlbum returns when the replay admitted no candidate at
// all (tag.ErrNoCandidates): the oracle comparison is then "no apply,
// no release", exactly classify's own ErrNoCandidates branch.
func reclassify(oracle OracleSignal, v *GoVerdict) string {
	if v == nil {
		return compareVerdicts(oracle, false, "")
	}
	return classifyVerdict(oracle, v)
}

// reevaluateAlbum re-runs the whole tag.Album flow for one observation
// against its own stored responses, and reports whether the replay
// asked for a request the recording does not carry.
//
// It reuses replayTransport and matchAlbum rather than declaring a
// second replay path: matchAlbum builds shadowScoring's own tunables
// and compat internally, wraps the transport it is given in a
// recordingTransport, and drives musicbrainz.NewClient exactly as the
// original run did. The one difference from rescoreDefault is that this
// re-runs the WHOLE flow through tag.Album, because a newer matcher
// revision may take a different branch (a search where the original run
// took the id-consensus early return, or the reverse), while
// rescoreDefault re-scores a recorded candidate set under the other
// compat mode.
//
// A miss is checked only once matchAlbum has returned an error, on the
// same grounds rescoreDefault checks it: a 404 would be silently
// absorbed by tag.Album's lookupAndAdd (tag.go:166-170) rather than
// erroring at all, which is exactly why the transport refuses a miss
// with 501 instead.
func reevaluateAlbum(ctx context.Context, s *Store, obs *Observation) (*GoVerdicts, bool, error) {
	replay := &replayTransport{bodies: make(map[string][]byte, len(obs.Requests))}
	for _, r := range obs.Requests {
		body, err := s.Response(obs.ID, r.Response)
		if err != nil {
			return nil, false, fmt.Errorf("reading stored response %s: %w", r.Response, err)
		}
		replay.bodies[r.PathQuery] = body
	}

	prop, requests, bodies, err := matchAlbum(ctx, Config{
		BaseURL:   rescoreBaseURL,
		UserAgent: rescoreUserAgent,
		Transport: replay,
	}, musicbrainz.Unmetered(), obs.Capture.Items, obs.EvalDate)
	switch {
	case errors.Is(err, tag.ErrNoCandidates):
		// A legitimate outcome, exactly as classify treats it: nothing
		// was admitted, so there is no verdict to score, only a skip by
		// absence.
		return &GoVerdicts{}, false, nil
	case err != nil:
		if len(replay.misses) > 0 {
			return nil, true, nil
		}
		return nil, false, fmt.Errorf("re-running tag.Album: %w", err)
	}

	verdicts := &GoVerdicts{Compat: goVerdict(prop.Recommendation, prop.Action, prop.Candidates)}
	ids := make([]string, len(prop.Candidates))
	for i, c := range prop.Candidates {
		ids[i] = c.Info.AlbumID
	}
	if v, err := rescoreDefault(ctx, requests, bodies, ids, obs.Capture.Items, obs.EvalDate); err == nil {
		verdicts.Default = v
	}
	return verdicts, false, nil
}
