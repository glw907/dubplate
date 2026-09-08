package conform

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
	"github.com/glw907/dubplate/internal/engine/tag"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/stub"
)

// Floors for the whole-corpus sweep, measured off the tree at
// 2026-09-05. They are floors rather than exact counts because lane E1
// adds cases: a shrink means a case, a candidate or a comparison went
// missing, which is the failure worth catching, while growth is the
// corpus doing its job.
//
// AssignmentFloor is the 214 candidate-bearing cases, and it is never
// compared bare: ruling F's escalation moves the two pinned cases into
// TiesAwaitingRuling and an owner ruling can move a case into ties
// excluded, both of which legally shrink the population. The sweep is
// compared against AssignmentFloor - len(TiesAwaitingRuling) -
// TiesExcluded. Comparing bare would red the gate on the one path
// ruling F pre-authorizes.
//
// CaseFloor is RequestCaseFloor (request.go:28), the same count of the
// same directories, so it is defined from it rather than re-typed.
const (
	CaseFloor        = RequestCaseFloor
	CandidateFloor   = 294
	DistanceFloor    = 294
	AssignmentFloor  = 214
	VerdictFloor     = 215
	TrackVectorFloor = 294
)

// DefectFreeFloor is the number of adjudication rows carrying a
// defect_free block today, which is ModeDefault's whole graded
// population.
const DefectFreeFloor = 10

// ProbeCaseFloor is the number of probe cases across the four
// hypotheses in corpus/analysis/oracle_bugs.json not ruled
// oracle-correct: bug-ft-pattern (6), bug-remix-parenthetical (4),
// bug-nonlatin-alias-selection (4), and bug-assignment-rejection-gap
// (6, no-correct-answer, owner ruling 2026-09-07). The other two
// hypotheses carry probe cases too but are excluded here because they
// are ruled oracle-correct: bug-threshold-boundary's
// rec-strong-medium-boundary, and bug-locale-alias-ignored's
// real-nonlatin-utada-first-love and real-nonlatin-utada-first-love-locale.
const ProbeCaseFloor = 20

// ProbeCaseOracleCorrectRows is the number of those probe cases whose
// OWN adjudication row reads oracle-correct, so they carry no corrected
// verdict to grade and belong to neither the graded nor the ungraded
// population: real-nonlatin-alla-pugacheva-vol1 and
// real-nonlatin-x-japan-art-of-life, both probes of
// bug-nonlatin-alias-selection. Naming it keeps the accounting
// assertion readable and makes it obvious what goes stale the day Geoff
// rules another row; the assertion recomputes the figure from the
// ledgers and compares against this constant, so a movement fails
// loudly rather than being absorbed (fold m17).
const ProbeCaseOracleCorrectRows = 2

// TiesExcludedCeiling is the number of ties.json rows dispositioned
// excluded today. It is a ceiling rather than a floor: excluding a case
// stops grading it, disposition is owner-only, and "the excluded count
// is reported at every run and may not grow without an owner ruling"
// (rung 3 spec, :292). Growth is a hard failure, never a pending row.
const TiesExcludedCeiling = 0

// nearMissTolerance is the relative agreement a distance comparison
// admits before it is a disagreement. A match inside it that is not
// bit-identical is a reported NearMiss, so the port's remaining
// arithmetic drift stays visible (rung 3 spec, :141).
const nearMissTolerance = 1e-12

// tieMirroringFinding is the escalation document ruling F requires
// before a tie row may go ungraded, relative to the repo root.
const tieMirroringFinding = "docs/superpowers/research/2026-09-05-rung-3d-tie-mirroring-finding.md"

// tiesAwaitingRuling names the pinned ties.json cases whose assignment
// this engine does not reproduce, and which therefore grade as pending
// rather than as passing until Geoff rules their disposition (rung 3
// spec, ruling F).
//
// It is EMPTY, and that is a measurement rather than a placeholder: the
// tie-mirroring attempt reproduced lap 0.5.13's choice among equal
// optima on both pinned cases, so no ruling is owed and no finding
// document exists. The mechanism stays because the ruling can be
// reopened by any future lap or corpus movement, and
// checkTiesAwaitingRuling is what keeps a non-empty list from being a
// silent escape.
var tiesAwaitingRuling []string

// A NearMiss is a comparison that matched within nearMissTolerance
// relative without being bit-identical.
type NearMiss struct {
	CaseID string
	Field  string
	Want   float64
	Got    float64
}

// CaseGradeSummary reports what one sweep actually compared, never what
// it visited.
type CaseGradeSummary struct {
	Mode               GradeMode
	Cases              int
	Candidates         int
	Distances          int
	TrackVectors       int
	Assignments        int
	Verdicts           int
	TiesExcluded       int
	TiesAwaitingRuling []string // ruling F is owed on these; graded as pending, never as passing
	Pending            []string // case ids whose adjudication row leaves the verdict ungraded
	Ungraded           []string // cases with no gradable corrected verdict: an oracle-wrong class with no expressible rung-3 fix, or a no-correct-answer disposition
	NearMisses         []NearMiss
	VerdictMoves       []string // ModeDefault only
}

// GradeMode selects which of ruling A's two runs a sweep is.
//
// Two runs rather than one, and the reason is arithmetic. Both
// string-heuristic hypotheses are ruled oracle-wrong, so ruling A
// defaults both compat switches OFF, and a default-mode run scores every
// case against a non-beets pattern table and diverges from every
// committed distance. One run cannot be both bit-identical against the
// committed records and correct against the adjudicated verdicts.
//
// ModeDefault takes the ledger's own dispositions. Each probe case of an
// oracle-wrong hypothesis with a machine-readable defect_free block is
// graded against that block; NO other case's verdict is graded, because
// no other case has a corrected expectation.
//
// ModeDefault does not drive tag.Album. It re-scores each case's own
// committed candidate set under the corrected table, and the reason is
// that a corrected verdict changes the FLOW: all ten defect_free probe
// cases log one lookup and zero searches and score strong under compat
// and medium under the corrected table, so re-running the flow would
// lose the strong-match early return (match.py:436), issue a search the
// case never logged, and fail the request-set rule before any
// defect_free figure was compared. The request-set rule is therefore
// ModeCompat's alone. Re-scoring reproduces the ten committed figures
// exactly, because the ladder's medium rung is decided by the minimum
// distance alone (match.py:288-290) and a candidate set a search would
// have widened cannot move a medium verdict off medium. Revision 2,
// fold B2; the ruling is recorded at the rung 4 spec's conductor fold
// and at 3d's HISTORY entry, and rung 4's second scoring mode tracks
// this design.
type GradeMode int

const (
	// ModeCompat reproduces every ledgered defect: bit-identity against
	// every committed record, including the adversarial stratum. This is
	// the run tier 1 means.
	ModeCompat GradeMode = iota
	// ModeDefault takes the ledger's own dispositions.
	ModeDefault
)

// GradeCases replays every committed case through the Go matcher's own
// public entry point and grades the result against the case's record,
// honoring each case's dispositions across all four analysis ledgers.
//
// It grades the request SET as well as the record: each case runs
// against a stub built from its own log, and a run leaving a non-empty
// Misses or a non-empty Unused fails naming both. 3c graded request
// construction and handed the set here, and the set is a function of
// the ID-match early return, _recommendation and _add_candidate, which
// this pass owns. The rule does not reach ModeDefault, where a
// corrected verdict legally changes the flow.
func GradeCases(root string, mode GradeMode) (CaseGradeSummary, error) {
	led, err := loadLedgers(root)
	if err != nil {
		return CaseGradeSummary{}, err
	}

	casesRoot := filepath.Join(root, "corpus", "cases")
	entries, err := os.ReadDir(casesRoot)
	if err != nil {
		return CaseGradeSummary{}, fmt.Errorf("read corpus/cases: %w", err)
	}

	ctx := context.Background()
	summary := CaseGradeSummary{Mode: mode}
	var errs []error
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(casesRoot, e.Name())

		var res caseGradeResult
		var caseErr error
		if mode == ModeCompat {
			res, caseErr = gradeCase(ctx, dir, led)
		} else {
			res, caseErr = gradeCaseDefault(ctx, dir, led)
		}
		if caseErr != nil {
			errs = append(errs, caseErr)
		}

		// The fields both modes report: a case's verdict is graded,
		// pending, ungraded or moved regardless of which sweep produced
		// it.
		summary.Verdicts += res.verdicts
		if res.pending {
			summary.Pending = append(summary.Pending, res.caseID)
		}
		if res.ungraded {
			summary.Ungraded = append(summary.Ungraded, res.ungradedReason)
		}
		if res.verdictMoved {
			summary.VerdictMoves = append(summary.VerdictMoves, res.caseID)
		}

		if mode != ModeCompat || !res.graded {
			// ModeDefault re-scores rather than replaying the whole
			// flow, so the candidate/distance/assignment floors are
			// ModeCompat's own claim; a case that could not be loaded,
			// stubbed or replayed contributes its errors and nothing
			// else, and the counters report what was compared.
			continue
		}
		summary.Cases++
		summary.Candidates += res.candidates
		summary.Distances += res.distances
		summary.TrackVectors += res.trackVectors
		summary.Assignments += res.assignments
		summary.NearMisses = append(summary.NearMisses, res.nearMisses...)
		if res.excluded {
			summary.TiesExcluded++
		}
		if res.awaitingRuling {
			summary.TiesAwaitingRuling = append(summary.TiesAwaitingRuling, res.caseID)
		}
	}

	return summary, errors.Join(errs...)
}

// ledgers is the four analysis ledgers a sweep resolves every case's
// grading policy from.
type ledgers struct {
	ties *analysis.TieLedger
	prox *analysis.ProximityLedger
	adj  *analysis.AdjudicationLedger
	bugs *analysis.OracleBugLedger
}

func loadLedgers(root string) (ledgers, error) {
	dir := filepath.Join(root, "corpus", "analysis")
	var (
		led ledgers
		err error
	)
	if led.ties, err = analysis.LoadTieLedger(filepath.Join(dir, "ties.json")); err != nil {
		return ledgers{}, err
	}
	if led.prox, err = analysis.LoadProximityLedger(filepath.Join(dir, "proximity.json")); err != nil {
		return ledgers{}, err
	}
	if led.adj, err = analysis.LoadAdjudications(filepath.Join(dir, "adjudications.json")); err != nil {
		return ledgers{}, err
	}
	if led.bugs, err = analysis.LoadOracleBugs(filepath.Join(dir, "oracle_bugs.json")); err != nil {
		return ledgers{}, err
	}
	return led, nil
}

// caseGradeResult is one case's contribution to a CaseGradeSummary.
// graded says the case reached its record at all: a case that failed to
// load, stub or replay contributes no count, which is what makes a
// silently skipped case show up as a floor shortfall rather than as a
// clean run over fewer cases.
type caseGradeResult struct {
	caseID         string
	graded         bool
	candidates     int
	distances      int
	trackVectors   int
	assignments    int
	excluded       bool
	pending        bool
	awaitingRuling bool
	nearMisses     []NearMiss

	// verdicts, ungraded, ungradedReason and verdictMoved are the tier-2
	// contribution both modes report (Task 8): 1 when this case's
	// recommendation and action were graded and agreed, ungraded when
	// it is a probe case with no expressible rung-3 corrected
	// expectation (Resolution R1), and verdictMoved when ModeDefault's
	// re-scored verdict differs from the case's own committed (that is,
	// ModeCompat) verdict.
	verdicts       int
	ungraded       bool
	ungradedReason string
	verdictMoved   bool
}

// caseInputs resolves what both sweeps need from a loaded record before
// they can replay it: the evaluation date, the tunables the case's
// profile selects, and the case's items projected into the engine's own
// field set.
func caseInputs(c *cassette.Case) (time.Time, matcher.Tunables, []matcher.ItemFields, error) {
	evalDate, err := time.Parse(time.DateOnly, c.EvalDate)
	if err != nil {
		return time.Time{}, matcher.Tunables{}, nil, fmt.Errorf("case %s: eval_date %q: %w", c.ID, c.EvalDate, err)
	}
	tun, err := gradeTunables(c.Profile)
	if err != nil {
		return time.Time{}, matcher.Tunables{}, nil, fmt.Errorf("case %s: %w", c.ID, err)
	}
	items := make([]matcher.ItemFields, len(c.Input.Items))
	for i, raw := range c.Input.Items {
		items[i] = matcher.ProjectItemFields(raw)
	}
	return evalDate, tun, items, nil
}

// gradeClient builds the client a sweep replays one case through: the
// case's own stub as the base URL, and an unmetered budget, because a
// replay spends no real MusicBrainz quota.
func gradeClient(addr string) (*musicbrainz.Client, error) {
	return musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   "http://" + addr,
		UserAgent: gradeUserAgent,
		Budget:    musicbrainz.Unmetered(),
	})
}

// gradeCase is ModeCompat's own per-case grading: it replays one case
// and compares every tier-1 field of its record. It keeps comparing
// after a disagreement rather than returning the first, so one systemic
// defect reports its full count in a single run. ModeDefault never
// reaches here; gradeCaseDefault is its whole path.
func gradeCase(ctx context.Context, dir string, led ledgers) (res caseGradeResult, err error) {
	var errs []error
	defer func() {
		err = errors.Join(errs...)
	}()

	c, loadErr := cassette.LoadCase(dir)
	if loadErr != nil {
		errs = append(errs, fmt.Errorf("case at %s: load: %w", dir, loadErr))
		return res, nil
	}
	res.caseID = c.ID
	if c.Expected == nil {
		errs = append(errs, fmt.Errorf("case %s: carries no derived record", c.ID))
		return res, nil
	}
	exp := c.Expected

	policy := policyFor(c.ID, led.ties, led.prox, led.adj, led.bugs)
	res.excluded = policy.Excluded
	res.pending = policy.Pending
	res.awaitingRuling = slices.Contains(tiesAwaitingRuling, c.ID)

	evalDate, tun, items, inputErr := caseInputs(c)
	if inputErr != nil {
		errs = append(errs, inputErr)
		return res, nil
	}

	srv, startErr := stub.New(c, dir)
	if startErr != nil {
		errs = append(errs, fmt.Errorf("case %s: start stub: %w", c.ID, startErr))
		return res, nil
	}
	defer func() {
		// One stub per case, closed before the next one starts, so peak
		// memory is one case's response bodies rather than the corpus's.
		if closeErr := srv.Close(); closeErr != nil {
			errs = append(errs, fmt.Errorf("case %s: close stub: %w", c.ID, closeErr))
		}
	}()

	client, clientErr := gradeClient(srv.Addr())
	if clientErr != nil {
		errs = append(errs, fmt.Errorf("case %s: build client: %w", c.ID, clientErr))
		return res, nil
	}

	compat := compatFor(ModeCompat, policy)
	prop, tagErr := tag.Album(ctx, client, items, tun, compat, evalDate, tag.Options{})
	if tagErr != nil && !errors.Is(tagErr, tag.ErrNoCandidates) {
		errs = append(errs, fmt.Errorf("case %s: %w", c.ID, tagErr))
		return res, nil
	}
	// ErrNoCandidates is not a failure by itself: one committed case
	// produces no candidate by design, and the record's own candidate
	// count is what says whether this run should have.
	var got []matcher.Candidate
	if prop != nil {
		got = prop.Candidates
	}
	res.graded = true

	errs = append(errs, checkRequestSet(c.ID, srv)...)

	if exp.EchoedDate != c.EvalDate {
		errs = append(errs, fmt.Errorf("case %s: echoed_date %q does not equal the case's own eval_date %q", c.ID, exp.EchoedDate, c.EvalDate))
	}
	if weightsErr := CheckCaseWeights(c.ID, tun.DistanceWeights, exp.Weights); weightsErr != nil {
		errs = append(errs, weightsErr)
	}
	if threshErr := CheckCaseThresholds(c.ID, tun, exp.Thresholds); threshErr != nil {
		errs = append(errs, threshErr)
	}

	errs = append(errs, gradeCandidates(c.ID, policy, exp.Candidates, got, &res)...)
	errs = append(errs, gradeAssignment(c.ID, policy, exp, got, &res)...)

	// The verdict: Recommendation and Action, against the case's own
	// record. tag.Album already computed both when it produced a
	// proposal; the one committed case that produces no candidate at
	// all took the ErrNoCandidates branch above and never reached
	// Recommend, so its verdict is computed here the same way tag.Album
	// would have (Recommend on an empty slice returns RecNone directly,
	// match.py:279-281).
	gotRec, gotAction := matcher.RecNone, matcher.Action("")
	if prop != nil {
		gotRec, gotAction = prop.Recommendation, prop.Action
	} else {
		var actionErr error
		gotAction, actionErr = matcher.ActionFor(compat.Apply(tun), gotRec)
		if actionErr != nil {
			errs = append(errs, fmt.Errorf("case %s: %w", c.ID, actionErr))
		}
	}
	if verdictErr := gradeCaseVerdict(c.ID, policy, exp.Recommendation, exp.Action, gotRec, gotAction, &res); verdictErr != nil {
		errs = append(errs, verdictErr)
	}

	return res, nil
}

// gradeCaseVerdict is ModeCompat's own verdict grading: it compares one
// case's computed recommendation and action against its own committed
// record, honoring the pending escape (a policy marked pending is never
// compared at all) and refusing to absorb an unadjudicated divergence
// silently (Resolution R15, rung 3 spec :311). It consults
// policy.Pending rather than policy.GradeVerdict: a no-correct-answer
// hypothesis still has a committed record to compare against, and
// compat measures fidelity to that recorded output regardless of
// whether a corrected verdict exists (owner ruling 2026-09-07).
// GradeVerdict itself is ModeDefault's own escape, consulted by
// gradeCaseDefault, since a no-correct-answer hypothesis has no
// corrected verdict for ModeDefault to grade.
func gradeCaseVerdict(caseID string, policy casePolicy, wantRec, wantAction string, gotRec matcher.Recommendation, gotAction matcher.Action, res *caseGradeResult) error {
	if policy.Pending {
		return nil
	}
	if gotRec.String() == wantRec && string(gotAction) == wantAction {
		res.verdicts++
		return nil
	}

	// An unadjudicated divergence. The gate refuses it rather than
	// absorbing it silently: a mechanical disagreement is a Go defect
	// until an owner says otherwise, and the one escape, a case whose
	// adjudication row already reads pending-adjudication, was already
	// excluded above via policy.GradeVerdict. The failure prints the
	// exact row a human pastes into adjudications.json.
	row := analysis.AdjudicationEntry{
		ID:           "adj-" + caseID,
		Subject:      analysis.AdjudicationSubject{CaseID: caseID},
		Suspected:    "TODO",
		Citation:     "TODO",
		ReleaseTruth: "TODO",
		Disposition:  analysis.Pending,
		Reason:       "TODO",
		Adjudicator:  analysis.AgentAdjudicator,
		Date:         "TODO",
	}
	data, marshalErr := json.MarshalIndent(row, "", "  ")
	if marshalErr != nil {
		return fmt.Errorf("case %s: marshal pending row: %w", caseID, marshalErr)
	}
	return fmt.Errorf("case %s: verdict %s/%s does not match record %s/%s, and no adjudication row makes that legal; paste this row into adjudications.json:\n%s",
		caseID, gotRec, gotAction, wantRec, wantAction, data)
}

// checkRequestSet applies the two halves of the request-set rule.
// Misses names a request the flow issued that the case never logged;
// Unused names a logged request the flow never issued, which no miss
// count can see. Both are failures: a run that took a different set
// than pinned beets took took a different flow.
func checkRequestSet(caseID string, srv *stub.Server) []error {
	var errs []error
	for _, m := range srv.Misses() {
		errs = append(errs, fmt.Errorf("case %s: unresolved request %s %s", caseID, m.Method, m.PathQuery))
	}
	for _, req := range srv.Unused() {
		errs = append(errs, fmt.Errorf("case %s: logged request never issued: %s %s", caseID, req.Method, req.PathQuery))
	}
	return errs
}

// compatFor returns the compat selection one case is graded under: the
// mode's own baseline, widened by every defect class the case's policy
// turns on. Widening is what a pending hypothesis asks for, and it is
// inert under ModeCompat, where both switches are already on.
func compatFor(mode GradeMode, policy casePolicy) matcher.Compat {
	base := matcher.DefaultCompat()
	if mode == ModeCompat {
		base = matcher.FullCompat()
	}
	return matcher.Compat{
		FtPattern:          base.FtPattern || policy.Compat.FtPattern,
		RemixParenthetical: base.RemixParenthetical || policy.Compat.RemixParenthetical,
	}
}

// gradeCaseDefault is ModeDefault's own per-case grading (revision 2,
// fold B2, Resolution R2/R9). It never drives tag.Album: instead it
// replays only the release lookups the case's own committed candidates
// name against the case's own stub, translates each, re-solves the
// assignment, re-scores the album distance and re-sorts, all under
// matcher.DefaultCompat()'s corrected pattern table, and issues no
// search. That reproduces the ten committed defect_free figures exactly
// (GradeMode's own doc comment states why) without ever taking a flow
// the case's request log does not cover, so the request-set rule stays
// ModeCompat's alone.
//
// A case whose adjudication row reads pending-adjudication is ungraded
// and reported pending, exactly as ModeCompat's own escape works. Every
// other case falls into one of three buckets: a case ruled
// no-correct-answer (policy.GradeVerdict false) has no corrected verdict
// to grade against and is marked ungraded; a probe of an oracle-wrong
// hypothesis whose OWN row agrees (policy.VerdictFrom) is graded against
// its defect_free block when the row carries one (the ten defect_free
// rows), or otherwise marked ungraded, since it has no expressible
// rung-3 fix (Resolution R1) either; every other case is compared
// against its own committed (that is, ModeCompat) verdict, and a
// difference is a VerdictMoves entry rather than a failure (Resolution
// R9).
func gradeCaseDefault(ctx context.Context, dir string, led ledgers) (res caseGradeResult, err error) {
	var errs []error
	defer func() {
		err = errors.Join(errs...)
	}()

	c, loadErr := cassette.LoadCase(dir)
	if loadErr != nil {
		errs = append(errs, fmt.Errorf("case at %s: load: %w", dir, loadErr))
		return res, nil
	}
	res.caseID = c.ID
	if c.Expected == nil {
		errs = append(errs, fmt.Errorf("case %s: carries no derived record", c.ID))
		return res, nil
	}
	exp := c.Expected

	policy := policyFor(c.ID, led.ties, led.prox, led.adj, led.bugs)
	if policy.Pending {
		res.pending = true
		return res, nil
	}

	evalDate, tun, items, inputErr := caseInputs(c)
	if inputErr != nil {
		errs = append(errs, inputErr)
		return res, nil
	}
	tunDefault := compatFor(ModeDefault, policy).Apply(tun)
	if valErr := tunDefault.Validate(); valErr != nil {
		errs = append(errs, fmt.Errorf("case %s: %w", c.ID, valErr))
		return res, nil
	}

	srv, startErr := stub.New(c, dir)
	if startErr != nil {
		errs = append(errs, fmt.Errorf("case %s: start stub: %w", c.ID, startErr))
		return res, nil
	}
	defer func() {
		if closeErr := srv.Close(); closeErr != nil {
			errs = append(errs, fmt.Errorf("case %s: close stub: %w", c.ID, closeErr))
		}
	}()

	client, clientErr := gradeClient(srv.Addr())
	if clientErr != nil {
		errs = append(errs, fmt.Errorf("case %s: build client: %w", c.ID, clientErr))
		return res, nil
	}

	cands := matcher.NewCandidates()
	for _, w := range exp.Candidates {
		info, lookupErr := client.AlbumForID(ctx, w.AlbumID, tunDefault)
		if lookupErr != nil {
			if errors.Is(lookupErr, musicbrainz.ErrNotFound) {
				// A committed candidate implies the lookup it was
				// derived from succeeded; a 404 here means the case's
				// own log no longer covers its own record.
				errs = append(errs, fmt.Errorf("case %s: re-score: committed candidate %s not found in the case's own log", c.ID, w.AlbumID))
				continue
			}
			errs = append(errs, fmt.Errorf("case %s: re-score: look up %s: %w", c.ID, w.AlbumID, lookupErr))
			return res, nil
		}
		if _, addErr := cands.Add(tunDefault, items, info, evalDate); addErr != nil {
			errs = append(errs, fmt.Errorf("case %s: re-score: %w", c.ID, addErr))
			return res, nil
		}
	}
	res.graded = true

	sorted := cands.Sorted()
	rec, recErr := matcher.Recommend(tunDefault, sorted)
	if recErr != nil {
		errs = append(errs, fmt.Errorf("case %s: %w", c.ID, recErr))
		return res, nil
	}
	action, actionErr := matcher.ActionFor(tunDefault, rec)
	if actionErr != nil {
		errs = append(errs, fmt.Errorf("case %s: %w", c.ID, actionErr))
		return res, nil
	}

	row, hasRow := led.adj.BySubjectCase()[c.ID]
	switch {
	case !policy.GradeVerdict:
		// A hypothesis (or its case's own row) ruled no-correct-answer:
		// the defect stays reproduced, since no compat switch exists to
		// fix it, but there is no corrected verdict to grade against
		// (Resolution R1). Honors the same GradeVerdict escape
		// gradeCaseVerdict does for ModeCompat, so the two modes agree
		// on which cases carry a gradable verdict at all.
		res.ungraded = true
		res.ungradedReason = fmt.Sprintf("%s: adjudicated no-correct-answer, so it has no corrected verdict to grade against (Resolution R1)", c.ID)
	case policy.VerdictFrom == verdictFromAdjudication && hasRow && row.DefectFree != nil:
		// One of the ten defect_free rows: graded against the block
		// rather than the case's own (defective) committed record.
		var dist float64
		if len(sorted) > 0 {
			dist = sorted[0].Dist.Distance()
		}
		if distErr := compareFloat(c.ID, "defect_free.distance", compareEpsilon, row.DefectFree.Distance, dist, &res); distErr != nil {
			errs = append(errs, distErr)
		}
		if rec.String() != row.DefectFree.Recommendation {
			errs = append(errs, fmt.Errorf("case %s: default-mode recommendation %s, defect_free record carries %s", c.ID, rec, row.DefectFree.Recommendation))
		}
		if string(action) != row.DefectFree.Action {
			errs = append(errs, fmt.Errorf("case %s: default-mode action %s, defect_free record carries %s", c.ID, action, row.DefectFree.Action))
		}
		res.verdicts++
	case policy.VerdictFrom == verdictFromAdjudication:
		// One of bug-nonlatin-alias-selection's two oracle-wrong probe
		// cases with no defect_free block (Resolution R1): no rung-3
		// port can express its corrected verdict, so it is neither
		// graded nor compared for movement.
		res.ungraded = true
		res.ungradedReason = fmt.Sprintf("%s: probe of an oracle-wrong hypothesis with no rung-3 fix, so its corrected verdict is not machine-readable (Resolution R1)", c.ID)
	default:
		// Every other case: compared against its own committed (that
		// is, ModeCompat) verdict. A difference is reported as a
		// VerdictMoves entry and fails nothing (Resolution R9).
		if rec.String() != exp.Recommendation || string(action) != exp.Action {
			res.verdictMoved = true
		}
	}

	return res, nil
}

// CheckDefaultCompatMatchesTheLedger proves matcher.DefaultCompat's two
// literals still agree with corpus/analysis/oracle_bugs.json's own
// dispositions. The engine never reads corpus/, so this is the only
// place the two can be compared, and it is a HARD failure rather than a
// pending row: a disposition is an owner decision, and an agent
// silently absorbing one is exactly what the sentinel rules exist to
// prevent.
func CheckDefaultCompatMatchesTheLedger(root string) error {
	bugs, err := analysis.LoadOracleBugs(filepath.Join(root, "corpus", "analysis", "oracle_bugs.json"))
	if err != nil {
		return err
	}
	byID := make(map[string]analysis.OracleBugEntry, len(bugs.Entries))
	for _, e := range bugs.Entries {
		byID[e.ID] = e
	}

	def := matcher.DefaultCompat()
	checks := []struct {
		hypothesisID string
		switchOn     bool
	}{
		{"bug-ft-pattern", def.FtPattern},
		{"bug-remix-parenthetical", def.RemixParenthetical},
	}

	var errs []error
	for _, chk := range checks {
		e, ok := byID[chk.hypothesisID]
		if !ok {
			errs = append(errs, fmt.Errorf("oracle_bugs.json: no entry %q", chk.hypothesisID))
			continue
		}
		reproducesDefect := e.Disposition != analysis.OracleWrong
		if chk.switchOn != reproducesDefect {
			errs = append(errs, fmt.Errorf("matcher.DefaultCompat: %s: switch is %v, but oracle_bugs.json disposition %q implies %v",
				chk.hypothesisID, chk.switchOn, e.Disposition, reproducesDefect))
		}
	}
	return errors.Join(errs...)
}

// gradeCandidates compares the run's candidates against the record's,
// matched in order by (Source, AlbumID) so a reordering reports itself
// rather than cascading into a value mismatch per field.
func gradeCandidates(caseID string, policy casePolicy, want []cassette.Candidate, got []matcher.Candidate, res *caseGradeResult) []error {
	var errs []error
	if policy.ExpectNoCandidate && len(got) != 0 {
		errs = append(errs, fmt.Errorf("case %s: produced %d candidates; its ledger row records that the record carries none", caseID, len(got)))
	}
	if len(want) != len(got) {
		return append(errs, fmt.Errorf("case %s: produced %d candidates, record carries %d", caseID, len(got), len(want)))
	}

	for k := range want {
		w, g := want[k], got[k]
		if g.Info.DataSource != w.Source || g.Info.AlbumID != w.AlbumID {
			errs = append(errs, fmt.Errorf("case %s: candidate %d is (%s, %s), record carries (%s, %s)",
				caseID, k, g.Info.DataSource, g.Info.AlbumID, w.Source, w.AlbumID))
			continue
		}
		res.candidates++

		if policy.GradeAlbumDistance {
			field := fmt.Sprintf("candidates[%d].distance", k)
			if err := compareFloat(caseID, field, policy.DistanceCompare, w.Distance, g.Dist.Distance(), res); err != nil {
				errs = append(errs, err)
			}
			res.distances++
			errs = append(errs, gradePenalties(caseID, k, policy, w.Penalties, g.Dist, res)...)
		}

		if len(w.Tracks) != len(g.Dist.Tracks) {
			errs = append(errs, fmt.Errorf("case %s: candidate %d: %d per-track distances, record carries %d",
				caseID, k, len(g.Dist.Tracks), len(w.Tracks)))
			continue
		}
		if !policy.GradeAlbumDistance {
			continue
		}
		for i, track := range g.Dist.Tracks {
			field := fmt.Sprintf("candidates[%d].tracks[%d]", k, i)
			if err := compareFloat(caseID, field, policy.DistanceCompare, w.Tracks[i], track.Dist.Distance(), res); err != nil {
				errs = append(errs, err)
			}
		}
		res.trackVectors++
	}
	return errs
}

// gradePenalties compares one candidate's penalty breakdown against the
// record's, which is the ADDED-KEY view: derive_oracle.py emits
// {key: dist[key] for key in dist._penalties} (:176), so a key added at
// 0.0 is part of what the record reports and grading against the
// nonzero view would drop it.
func gradePenalties(caseID string, k int, policy casePolicy, want map[string]float64, got *matcher.Distance, res *caseGradeResult) []error {
	var errs []error
	perKey := got.PerKey()
	for _, key := range got.AddedKeys() {
		if _, ok := want[key]; !ok {
			errs = append(errs, fmt.Errorf("case %s: candidate %d: penalty key %q is not in the record", caseID, k, key))
		}
	}
	for _, key := range slices.Sorted(maps.Keys(want)) {
		g, ok := perKey[key]
		if !ok {
			errs = append(errs, fmt.Errorf("case %s: candidate %d: penalty key %q missing (record carries %v)", caseID, k, key, want[key]))
			continue
		}
		field := fmt.Sprintf("candidates[%d].penalties[%s]", k, key)
		if err := compareFloat(caseID, field, policy.DistanceCompare, want[key], g, res); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// gradeAssignment compares the best candidate's solved pairing and its
// two extra lists. The best candidate is Sorted()[0], matching beets
// grading its own best candidate: derive_oracle.py records the top
// element of proposal.candidates (:242-243, :261-263).
func gradeAssignment(caseID string, policy casePolicy, exp *cassette.Expected, got []matcher.Candidate, res *caseGradeResult) []error {
	var errs []error
	if len(got) == 0 {
		if len(exp.ExtraItems) != 0 || len(exp.ExtraTracks) != 0 {
			errs = append(errs, fmt.Errorf("case %s: produced no candidate, record carries %d extra items and %d extra tracks",
				caseID, len(exp.ExtraItems), len(exp.ExtraTracks)))
		}
		return errs
	}
	if !policy.GradeAssignment || res.awaitingRuling {
		return errs
	}

	best := got[0].Assignment
	pairs := make(map[int]int, len(best.Pairs))
	for _, p := range best.Pairs {
		pairs[p.Item] = p.Track
	}
	if !maps.Equal(pairs, exp.Assignment) {
		errs = append(errs, fmt.Errorf("case %s: assignment %v, record carries %v", caseID, pairs, exp.Assignment))
	}
	res.assignments++

	if !slices.Equal(best.ExtraItems, exp.ExtraItems) {
		errs = append(errs, fmt.Errorf("case %s: extra items %v, record carries %v", caseID, best.ExtraItems, exp.ExtraItems))
	}
	if !slices.Equal(best.ExtraTracks, exp.ExtraTracks) {
		errs = append(errs, fmt.Errorf("case %s: extra tracks %v, record carries %v", caseID, best.ExtraTracks, exp.ExtraTracks))
	}
	return errs
}

// distanceCompare names how one case's distances are compared, which is
// what corpus/analysis/proximity.json's own row decides.
type distanceCompare int

const (
	// compareEpsilon compares at nearMissTolerance relative and reports
	// a match that is not bit-identical as a NearMiss.
	compareEpsilon distanceCompare = iota
	// compareBitIdentical admits no tolerance at all, because a flagged
	// case sits on a comparison boundary where an epsilon can change the
	// verdict (rung 3 spec, :288).
	compareBitIdentical
)

// compareFloat compares one recorded float against the run's own, under
// mode, appending a NearMiss to res for a match that needed the
// tolerance.
func compareFloat(caseID, field string, mode distanceCompare, want, got float64, res *caseGradeResult) error {
	if want == got {
		return nil
	}
	if mode == compareEpsilon && math.Abs(want-got) <= nearMissTolerance*max(math.Abs(want), math.Abs(got)) {
		res.nearMisses = append(res.nearMisses, NearMiss{CaseID: caseID, Field: field, Want: want, Got: got})
		return nil
	}
	return fmt.Errorf("case %s: %s: got %.17g, record carries %.17g", caseID, field, got, want)
}

// verdictSource names where a case's expected verdict comes from.
type verdictSource int

const (
	// verdictFromRecord is the case's own Expected.Recommendation and
	// Expected.Action.
	verdictFromRecord verdictSource = iota
	// verdictFromAdjudication is the corrected verdict on the case's own
	// adjudications.json row, resolved by subject.case_id, used once its
	// hypothesis's compat switch is off (rung 3 spec, :295).
	verdictFromAdjudication
)

// A casePolicy is how one case is graded, joined from all four analysis
// ledgers. It is the Go form of the rung 3 spec's disposition table
// (:285-299), and every field exists because one row of that table asks
// for it.
//
// Compat is the whole per-class selection rather than one bool, because
// "a switch is per defect class and cannot be half on" (spec :302) is a
// statement about EACH class: a case probing a pending bug-ft-pattern
// while bug-remix-parenthetical stays oracle-wrong needs {true, false},
// which a single bool would render as {true, true} and bake into rung
// 15's reference behavior. It is dormant today only because
// oracle_bugs.json carries zero pending rows.
type casePolicy struct {
	DistanceCompare    distanceCompare
	GradeAssignment    bool
	GradeAlbumDistance bool
	GradeVerdict       bool
	ExpectNoCandidate  bool
	Compat             matcher.Compat
	Excluded           bool
	Pending            bool
	VerdictFrom        verdictSource
}

// policyFor joins the four ledgers by case id and returns how the case
// is graded. It is the single place the spec's disposition table lives,
// so a row of that table has exactly one implementation and one fixture.
//
// The precedence rule is the spec's own (:302): the hypothesis row
// decides behavior and the case row supplies the expectation. So a
// hypothesis at pending-adjudication turns its class's compat switch on
// whatever the case row says, and a corrected verdict is taken only
// where the hypothesis is ruled oracle-wrong AND the case's own row
// carries that ruling.
func policyFor(caseID string, ties *analysis.TieLedger, prox *analysis.ProximityLedger,
	adj *analysis.AdjudicationLedger, bugs *analysis.OracleBugLedger) casePolicy {
	p := casePolicy{
		DistanceCompare:    compareEpsilon,
		GradeAssignment:    true,
		GradeAlbumDistance: true,
		GradeVerdict:       true,
		VerdictFrom:        verdictFromRecord,
	}

	for _, e := range prox.Entries {
		if e.CaseID != caseID {
			continue
		}
		switch e.Status {
		case analysis.ProximityFlagged:
			p.DistanceCompare = compareBitIdentical
		case analysis.ProximitySkipped:
			p.GradeAssignment = false
			p.GradeAlbumDistance = false
			p.ExpectNoCandidate = true
		}
	}

	for _, e := range ties.Entries {
		if e.CaseID != caseID {
			continue
		}
		if e.Status == analysis.TieSkipped {
			p.GradeAssignment = false
			p.GradeAlbumDistance = false
			p.ExpectNoCandidate = true
		}
		if e.Disposition == analysis.TieExcluded {
			p.GradeAssignment = false
			p.GradeAlbumDistance = false
			p.Excluded = true
		}
	}

	row, hasRow := adj.BySubjectCase()[caseID]
	for _, e := range bugs.Entries {
		if !slices.Contains(e.ProbeCaseIDs, caseID) {
			continue
		}
		switch e.Disposition {
		case analysis.Pending:
			p.Compat = widenCompat(p.Compat, e.ID)
		case analysis.NoCorrectAnswer:
			// No correct answer means no corrected verdict to grade
			// against, so the defect stays reproduced and the verdict
			// goes ungraded. The distances are still graded, at the tier
			// the case's own proximity row set above.
			p.Compat = widenCompat(p.Compat, e.ID)
			p.GradeVerdict = false
		case analysis.OracleWrong:
			if hasRow && row.Disposition == analysis.OracleWrong {
				p.VerdictFrom = verdictFromAdjudication
			}
		}
	}

	if hasRow {
		switch row.Disposition {
		case analysis.NoCorrectAnswer:
			p.GradeVerdict = false
		case analysis.Pending:
			p.GradeVerdict = false
			p.Pending = true
		}
	}

	return p
}

// widenCompat turns on the defect class hypothesisID names, leaving
// every other class as it was. A hypothesis whose defect no compat
// switch can express widens nothing: matcher.Compat carries the two
// string-heuristic classes and no more, because the other two
// hypotheses (bug-nonlatin-alias-selection, oracle-wrong;
// bug-assignment-rejection-gap, no-correct-answer) have no fix a rung-3
// port can express.
func widenCompat(c matcher.Compat, hypothesisID string) matcher.Compat {
	switch hypothesisID {
	case "bug-ft-pattern":
		c.FtPattern = true
	case "bug-remix-parenthetical":
		c.RemixParenthetical = true
	}
	return c
}

// checkTiesAwaitingRuling enforces ruling F's escape: a tie row may go
// ungraded only while the escalation finding it rests on exists. root
// is the repo root the finding is resolved against.
//
// It is the whole reason an empty awaiting list is safe to ship. A list
// that grows without the finding fails the gate rather than quietly
// shrinking the graded population.
func checkTiesAwaitingRuling(root string, awaiting []string) error {
	if len(awaiting) == 0 {
		return nil
	}
	if _, err := os.Stat(filepath.Join(root, tieMirroringFinding)); err != nil {
		return fmt.Errorf("%d tie rows await a ruling (%v) but the ruling F finding is missing: %w", len(awaiting), awaiting, err)
	}
	return nil
}
