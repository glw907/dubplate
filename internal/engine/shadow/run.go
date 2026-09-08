package shadow

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

const (
	// budgetName names this caller's spend in the limiter's own records
	// and in every pass record the report sums.
	budgetName = "shadow"

	// fixtureProfile is the corpus profile whose reduction equals
	// matcher.QuietImportTunables value for value
	// (corpus/profiles/estate.yaml:16-19).
	fixtureProfile = "estate"
)

// PassResult is one pass's record plus the two artifacts
// cmd/dubplate-server marshals. This package may import no harness
// package (contract S19), so the cassette-shaped fixture and the
// AdjudicationEntry-shaped row are described here in engine types and
// written there.
type PassResult struct {
	Pass PassRecord

	// Observations are in CAPTURE order, (uploader, basename), which is
	// the order the match step ran them in and therefore the order the
	// budget was spent in. It is a value here rather than a property of
	// the store, because an observation id is a digest prefix
	// (Resolution R5) and id order says nothing about capture order
	// (Resolution R18).
	Observations []*Observation

	Fixtures     []CaseFixture
	Pending      []PendingRow
	ClockStarted bool
}

// A CaseFixture is what cmd/dubplate-server needs to write
// observations/<obs-id>/case.json in cassette.Case's shape, with
// Expected nil.
//
// Profile is "estate". EvalDate is YYYY-MM-DD, the form every committed
// case carries.
type CaseFixture struct {
	ObservationID string
	Profile       string
	EvalDate      string
	Items         []map[string]any
	Requests      []RecordedRequest
}

// A PendingRow is what cmd/dubplate-server needs to write
// pending/<obs-id>.json in analysis.AdjudicationEntry's shape.
// Disposition and Adjudicator are not fields here: they are constants
// the marshaller supplies, because disposition is owner-only and every
// row the shadow writes is pending-adjudication by adjudicator agent
// (internal/harness/analysis/adjudication.go:198-200).
type PendingRow struct {
	ID                string
	Album             string
	ReleaseMBID       string
	MatcherCommit     string
	MatcherConfigHash string
	Suspected         string
	Citation          string
	ReleaseTruth      string
	Reason            string
	Date              string
}

// Run executes one observe pass: capture, reconcile, match.
//
// now is the pass clock and every timestamp the pass writes derives from
// it, which is what TestEngineNeverReadsTheClock requires
// (internal/harness/conform/arch_test.go:152-157) and what makes a whole
// pass deterministic under test.
//
// A pass is SUCCESSFUL when all three steps complete without a fatal
// error, a fatal error being a store write failure, an unusable state
// directory, or an invalid configuration. A pass that captures nothing,
// reconciles nothing and observes nothing is successful: the clock must
// start at the deploy rather than at the first upload, because the floor
// measures the observation window and not the arrival rate. A deferred
// observation is an ordinary branch and never makes a pass unsuccessful.
func Run(ctx context.Context, cfg Config, now time.Time) (*PassResult, error) {
	store, err := OpenStore(StoreDir(cfg.StateDir))
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}

	tun, compat := shadowScoring()
	res := &PassResult{Pass: PassRecord{
		ID:            passID(now),
		StartedAt:     now,
		EndedAt:       now,
		MatcherCommit: cfg.MatcherCommit,
		WireInterval:  wireInterval(cfg).String(),
		Divergences:   map[string]int{},
		Deferred:      map[string]int{},
		Unreconciled:  map[string]int{},
		Excluded:      map[string]int{},
	}}

	if err := captureStep(store, cfg, configHash(tun, compat), now, res); err != nil {
		return nil, err
	}
	matched, err := reconcileStep(store, cfg, res)
	if err != nil {
		return nil, err
	}
	if err := matchStep(ctx, store, cfg, matched, now, res); err != nil {
		return nil, err
	}

	if err := store.PutPass(&res.Pass); err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}
	started, err := store.StartClock(now)
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}
	res.ClockStarted = started
	return res, nil
}

// captureStep walks the observe directory and stores one observation per
// new album, carrying no oracle signal and no verdict yet.
func captureStep(store *Store, cfg Config, configHash string, now time.Time, res *PassResult) error {
	result, err := capture(cfg.ObserveDir, store.HasObservation, now)
	if err != nil {
		return fmt.Errorf("shadow: %w", err)
	}
	for _, c := range result.Captures {
		obs := &Observation{
			ID:                observationID(c.Digest),
			Album:             c.Uploader + "/" + c.Basename,
			MatcherCommit:     cfg.MatcherCommit,
			MatcherConfigHash: configHash,
			EvalDate:          now,
			Capture:           c,
		}
		if err := store.PutObservation(obs); err != nil {
			return fmt.Errorf("shadow: %w", err)
		}
	}
	res.Pass.Captured = len(result.Captures)
	res.Pass.PartialReads = len(result.PartialReads)
	for _, e := range result.Excluded {
		res.Pass.Excluded[e.Reason]++
	}
	return nil
}

// reconcileStep resolves every stored observation still awaiting an
// oracle signal against the incumbent's newly arrived run reports, tag
// logs and review tree, and returns the ones that gained one, in capture
// order.
//
// An observation that gained a signal is stored here as well as after
// the match step: the reports it was resolved from are consumed by the
// time this returns, so a later failure must not cost the signal.
func reconcileStep(store *Store, cfg Config, res *PassResult) ([]*Observation, error) {
	stored, err := store.Observations()
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}
	byDigest := make(map[string]*Observation, len(stored))
	var awaiting []Capture
	for _, obs := range stored {
		if obs.Oracle != nil || obs.Unreconciled != "" || obs.Excluded != "" || obs.Superseded {
			continue
		}
		byDigest[obs.Capture.Digest] = obs
		awaiting = append(awaiting, obs.Capture)
	}

	consumed, err := store.ConsumedReports()
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}
	reports, err := readRunReports(cfg.OracleStateDir)
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}
	skips, err := readTagLogSkips(cfg.OracleStateDir)
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}
	reviewed, err := reviewedBasenames(cfg.ReviewDir)
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}

	result, err := reconcile(reconcileInput{
		Captures:      awaiting,
		Reports:       reports,
		Consumed:      consumed,
		TagLogSkips:   skips,
		Reviewed:      reviewed,
		LibraryReadFn: readFLAC,
	})
	if err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}

	var matched []*Observation
	for _, r := range result.Reconciliations {
		obs := byDigest[r.Capture.Digest]
		if obs == nil {
			// A gate outcome that named no capture: quarantined and
			// rejected albums never reached beets and carry no digest,
			// so this one is counted and has no record to write onto.
			if r.Excluded != "" {
				res.Pass.Excluded[r.Excluded]++
			}
			continue
		}
		switch {
		case r.Superseded:
			obs.Superseded = true
			res.Pass.Superseded++
		case r.Excluded != "":
			obs.Excluded = r.Excluded
			res.Pass.Excluded[r.Excluded]++
		case r.Unreconciled != "":
			obs.Unreconciled = r.Unreconciled
			res.Pass.Unreconciled[r.Unreconciled]++
		case r.Oracle != nil:
			obs.Oracle = r.Oracle
			matched = append(matched, obs)
		}
		if err := store.PutObservation(obs); err != nil {
			return nil, fmt.Errorf("shadow: %w", err)
		}
	}
	if err := store.PutConsumedReports(result.Consumed); err != nil {
		return nil, fmt.Errorf("shadow: %w", err)
	}
	res.Pass.ReportGaps = result.Gaps

	slices.SortStableFunc(matched, func(a, b *Observation) int {
		return cmp.Or(
			cmp.Compare(a.Capture.Uploader, b.Capture.Uploader),
			cmp.Compare(a.Capture.Basename, b.Capture.Basename),
		)
	})
	return matched, nil
}

// matchStep scores every newly reconciled album, in capture order, which
// is the order the budget is spent in and therefore the order a cap
// bites in (Resolution R18).
func matchStep(ctx context.Context, store *Store, cfg Config, matched []*Observation, now time.Time, res *PassResult) error {
	limiter := cfg.Limiter
	if limiter == nil {
		limiter = musicbrainz.NewLimiter(wireInterval(cfg))
		// Close reports only that the limiter was already closed
		// (limiter.go:149-160), which cannot happen for one this call
		// built and owns.
		defer func() { _ = limiter.Close() }()
	}
	budget := limiter.Budget(budgetName, musicbrainz.ClassBackground, cfg.BudgetLimit)

	for _, obs := range matched {
		if err := observe(ctx, store, cfg, budget, obs, now); err != nil {
			return err
		}
		res.Pass.Observed++
		switch {
		case obs.Deferred != "":
			res.Pass.Deferred[obs.Deferred]++
		case obs.Class == ClassAgreement:
			res.Pass.Agreements++
		default:
			res.Pass.Divergences[obs.Class]++
		}

		res.Observations = append(res.Observations, obs)
		res.Fixtures = append(res.Fixtures, caseFixtureFor(obs, now))
		if row, ok := PendingRowFor(obs, now); ok {
			res.Pending = append(res.Pending, row)
		}
	}
	res.Pass.Budget = BudgetRecord{Name: budget.Name(), Spent: budget.Spent()}
	return nil
}

// observe runs one album through both scoring modes and the classifier,
// and stores what came back.
//
// A re-score that cannot run leaves the second verdict nil rather than a
// zero verdict a reader would compare against (the GoVerdicts contract),
// and it never fails the pass: mode one's recorded verdict is the
// measurement, and the re-score is the annotation beside it.
func observe(ctx context.Context, store *Store, cfg Config, budget *musicbrainz.Budget, obs *Observation, now time.Time) error {
	prop, requests, bodies, matchErr := matchAlbum(ctx, cfg, budget, obs.Capture.Items, now)
	for i, body := range bodies {
		if err := store.PutResponse(obs.ID, i+1, body); err != nil {
			return fmt.Errorf("shadow: %w", err)
		}
		requests[i].Response = responsePath(i + 1)
	}
	obs.Requests = requests
	obs.Budget = BudgetRecord{Name: budget.Name(), Spent: len(requests)}

	if matchErr == nil && prop != nil {
		obs.Go = &GoVerdicts{Compat: goVerdict(prop.Recommendation, prop.Action, prop.Candidates)}
		ids := make([]string, len(prop.Candidates))
		for i, c := range prop.Candidates {
			ids[i] = c.Info.AlbumID
		}
		if verdict, err := rescoreDefault(ctx, requests, bodies, ids, obs.Capture.Items, now); err == nil {
			obs.Go.Default = verdict
		}
	}
	obs.Class, obs.Deferred = classify(*obs.Oracle, prop, matchErr)

	if err := store.PutObservation(obs); err != nil {
		return fmt.Errorf("shadow: %w", err)
	}
	return nil
}

// caseFixtureFor projects one matched observation into the fixture
// cmd/dubplate-server writes as case.json.
func caseFixtureFor(obs *Observation, now time.Time) CaseFixture {
	items := make([]map[string]any, len(obs.Capture.Items))
	for i, fields := range obs.Capture.Items {
		items[i] = itemBag(fields)
	}
	return CaseFixture{
		ObservationID: obs.ID,
		Profile:       fixtureProfile,
		EvalDate:      now.Format(time.DateOnly),
		Items:         items,
		Requests:      obs.Requests,
	}
}

// PendingRowFor stages one divergence for owner adjudication, and
// reports whether it staged anything: an agreement and a deferral stage
// nothing. It is exported for cmd/dubplate-server's reevaluate wiring: a
// newly diverged observation Reevaluate promotes from an agreement was
// never staged by an observe pass, so the verb stages one itself, off
// the same construction an observe pass uses.
//
// News is every one of the three divergence classes, with one annotation
// rather than one exclusion. A divergence the DefaultCompat re-score
// does not reproduce is a hypothesised compat-class divergence, and
// Suspected says so; it still gets a row, because the annotation is a
// hypothesis rather than a ruling, and suppressing a row on a hypothesis
// is exactly the silent loss ruling 3's own limit paragraph warns about.
//
// The four prose fields are rendered from the observation rather than
// staged as TODO: a row that says nothing still consumes an owner
// adjudication.
func PendingRowFor(obs *Observation, now time.Time) (PendingRow, bool) {
	if !isDivergence(obs.Class) {
		return PendingRow{}, false
	}
	release, proposer := disputedRelease(obs)
	return PendingRow{
		ID:                obs.ID,
		Album:             obs.Album,
		ReleaseMBID:       release,
		MatcherCommit:     obs.MatcherCommit,
		MatcherConfigHash: obs.MatcherConfigHash,
		Suspected:         suspectedProse(obs),
		Citation:          "shadow observation " + obs.ID + ", class " + obs.Class,
		ReleaseTruth: "the release under dispute is " + release + ", proposed by " + proposer +
			"; no truth is ruled for it, which is what this row asks for",
		Reason: "observed against the incumbent's own live decision at matcher commit " +
			obs.MatcherCommit + ", pass " + passID(now),
		Date: now.Format(time.DateOnly),
	}, true
}

// disputedRelease returns the release under dispute and the side that
// proposed it (ruling 2): the oracle's release wherever the
// oracle applied, and the Go matcher's own wherever it did not. Every
// divergence has exactly one release under dispute, and the observation
// record says which side proposed it, so nothing is lost and the
// AdjudicationSubject pair rule stays a check that can fire.
func disputedRelease(obs *Observation) (release, proposer string) {
	if obs.Class == ClassGoAppliesOracleSkipped {
		if obs.Go != nil && obs.Go.Compat != nil {
			return obs.Go.Compat.AlbumID, "the Go matcher"
		}
		return "", "the Go matcher"
	}
	return obs.Oracle.ReleaseMBID, "beets"
}

// suspectedProse names the class, both verdicts, and the compat-class
// hypothesis where the re-score does not reproduce the divergence.
func suspectedProse(obs *Observation) string {
	var b strings.Builder
	b.WriteString(obs.Class + ": beets " + oracleProse(obs.Oracle) + " and the Go matcher " + verdictProse(compatVerdict(obs)))
	if compatClassHypothesis(obs) {
		b.WriteString(". The DefaultCompat re-score of this run's own candidate set agrees with beets, " +
			"so the divergence is hypothesised to be compat-class rather than a matcher defect. " +
			"The re-score is a hypothesis and not a ruling: it scores the candidate set the recorded " +
			"FullCompat flow gathered, not the one a DefaultCompat flow would have")
	}
	return b.String()
}

// compatClassHypothesis reports whether the second verdict agrees with
// the oracle where the first does not.
func compatClassHypothesis(obs *Observation) bool {
	if obs.Go == nil || obs.Go.Default == nil {
		return false
	}
	return classifyVerdict(*obs.Oracle, obs.Go.Default) == ClassAgreement
}

// compatVerdict is mode one's verdict, or nil where the flow returned no
// candidate at all.
func compatVerdict(obs *Observation) *GoVerdict {
	if obs.Go == nil {
		return nil
	}
	return obs.Go.Compat
}

func oracleProse(o *OracleSignal) string {
	if o.Signal == "apply" {
		return "applied release " + o.ReleaseMBID
	}
	return "skipped the album"
}

func verdictProse(v *GoVerdict) string {
	switch {
	case v == nil:
		return "found no candidate at all"
	case v.Action == string(matcher.ActionApply):
		return "applied release " + v.AlbumID
	default:
		return "declined to apply, at recommendation " + v.Recommendation
	}
}

// responsePath names one stored body the way Store.PutResponse writes
// it, relative to the observation directory, which is the form
// cassette.Case's own Request.Response carries.
func responsePath(n int) string {
	return fmt.Sprintf("%s/%03d.json", responsesDir, n)
}

// passID names one pass by its own clock. The observe timer is a systemd
// singleton firing at minutes' pace, so second resolution cannot
// collide; a re-run at one frozen clock overwrites its own record rather
// than accumulating a duplicate, which is what makes a pass reproducible
// under test.
func passID(now time.Time) string {
	return "pass-" + now.UTC().Format("20060102T150405Z")
}

// wireInterval is the pace a pass builds its own limiter at: cfg.Interval
// when set, DefaultInterval otherwise.
func wireInterval(cfg Config) time.Duration {
	return cmp.Or(cfg.Interval, DefaultInterval)
}
