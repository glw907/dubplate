package shadow

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// The replay client's own identity and base URL. musicbrainz.NewClient
// refuses an empty User-Agent (client.go:124-130), so the re-score needs
// one to be constructible at all, and it is a constant rather than a
// site value because this client never reaches a network: replayTransport
// answers every request from the observation's own stored bodies and
// refuses a miss. The .invalid base states the same thing at the address:
// the name is reserved and resolves nowhere, so a transport that ever
// went missing would fail to dial rather than reach MusicBrainz.
const (
	rescoreUserAgent = "dubplate-shadow-rescore/1 ( replay )"
	rescoreBaseURL   = "http://replay.invalid"
)

// rescoreScoring returns the tunables and compat pair the re-score
// scores under: DefaultCompat applied to the estate's own tunables
// (ruling 3, provisional). It mirrors shadowScoring, so a test can hash
// exactly what the re-score used rather than inferring it from a verdict
// that names no table.
func rescoreScoring() (matcher.Tunables, matcher.Compat) {
	compat := matcher.DefaultCompat()
	return compat.Apply(matcher.QuietImportTunables()), compat
}

// rescoreDefault re-scores one observation's own recorded candidate set
// under DefaultCompat and returns the second verdict.
//
// It drives NO second tag.Album and issues NO search. That is what
// conform's own ModeDefault does, and the reason is that a corrected
// verdict changes the FLOW: the ID-consensus early return fires only at
// matcher.RecStrong (internal/engine/tag/tag.go:96-118, condition at
// :106), Compat.Apply swaps StringDistPatterns
// (internal/engine/matcher/compat.go:51-63), which moves the distance
// and therefore the recommendation, so an album that took the one-lookup
// early return under FullCompat falls through to src.SearchReleases
// (tag.go:125-126) under DefaultCompat, a request the recording never
// captured.
//
// The five steps mirror gradeCaseDefault
// (internal/harness/conform/case.go:568-622): apply and validate, look
// each recorded candidate up through a replay client, cands.Add per
// candidate, cands.Sorted then matcher.Recommend, then matcher.ActionFor.
//
// The replay client takes musicbrainz.Unmetered()
// (musicbrainz/limiter.go:354-359), because the budget sits ABOVE the
// transport: doWithRetry calls budget.Wait before every attempt
// (client.go:346) whatever ClientConfig.HTTP holds, so a replay through
// the metered client would double the ClassBackground spend and double
// the figure the report prints as honest.
//
// A candidate the recording does not cover is an error rather than a
// dropped candidate: a verdict computed over a candidate set the run
// never had would retire a divergence that never actually re-ran.
//
// The verdict it returns is marked Hypothesis. A re-scored DefaultCompat
// verdict is not the product's true DefaultCompat verdict wherever the
// real flow would have widened the candidate set, and conform's own
// exactness argument (case.go:149-152) covers ten pinned corpus cases
// and does not generalize to an arbitrary live album.
func rescoreDefault(ctx context.Context, recorded []RecordedRequest, bodies [][]byte,
	albumIDs []string, items []matcher.ItemFields, evalDate time.Time) (*GoVerdict, error) {
	if len(recorded) != len(bodies) {
		return nil, fmt.Errorf("re-score: %d recorded requests against %d bodies", len(recorded), len(bodies))
	}

	tun, _ := rescoreScoring()
	if err := tun.Validate(); err != nil {
		return nil, fmt.Errorf("re-score: %w", err)
	}

	replay := &replayTransport{bodies: make(map[string][]byte, len(recorded))}
	for i, r := range recorded {
		replay.bodies[r.PathQuery] = bodies[i]
	}
	client, err := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   rescoreBaseURL,
		UserAgent: rescoreUserAgent,
		Budget:    musicbrainz.Unmetered(),
		HTTP:      &http.Client{Transport: replay},
	})
	if err != nil {
		return nil, fmt.Errorf("re-score: building replay client: %w", err)
	}

	cands := matcher.NewCandidates()
	for _, id := range albumIDs {
		info, err := client.AlbumForID(ctx, id, tun)
		if err != nil {
			// The transport's own miss list is the signal, not the
			// error's shape: a 501 leaves doWithRetry as an unwrapped
			// fmt.Errorf with no sentinel to test (contract S18).
			if len(replay.misses) > 0 {
				return nil, fmt.Errorf("re-score: candidate %s is not covered by the recording (missed %v)", id, replay.misses)
			}
			return nil, fmt.Errorf("re-score: look up %s: %w", id, err)
		}
		if _, err := cands.Add(tun, items, info, evalDate); err != nil {
			return nil, fmt.Errorf("re-score: admitting %s: %w", id, err)
		}
	}

	sorted := cands.Sorted()
	rec, err := matcher.Recommend(tun, sorted)
	if err != nil {
		return nil, fmt.Errorf("re-score: recommending: %w", err)
	}
	action, err := matcher.ActionFor(tun, rec)
	if err != nil {
		return nil, fmt.Errorf("re-score: resolving the action: %w", err)
	}

	verdict := goVerdict(rec, action, sorted)
	verdict.Hypothesis = true
	return verdict, nil
}

// goVerdict projects one scoring mode's answer into the stored record's
// shape. cands is in ascending distance order, so cands[0] is the
// winning candidate and its release is the one the mode would apply.
func goVerdict(rec matcher.Recommendation, action matcher.Action, cands []matcher.Candidate) *GoVerdict {
	v := &GoVerdict{Recommendation: rec.String(), Action: string(action)}
	if len(cands) > 0 {
		v.AlbumID = cands[0].Info.AlbumID
		v.Distance = cands[0].Dist.Distance()
		v.NonzeroKeys = cands[0].Dist.NonzeroKeys()
	}
	return v
}
