package conform

import (
	"context"
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
	Ungraded           []string // probe cases of an oracle-wrong class with no expressible rung-3 fix
	NearMisses         []NearMiss
	VerdictMoves       []string // ModeDefault only
}

// GradeMode selects which of ruling A's two runs a sweep is.
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
	if mode != ModeCompat {
		// The default-compat run is a different sweep rather than a flag
		// on this one: it re-scores each case's committed candidate set
		// and issues no request at all, because every defect_free probe
		// case logs one lookup and no search, so replaying the flow
		// under a corrected table would issue a request the case never
		// logged.
		return CaseGradeSummary{}, fmt.Errorf("grade mode %d: only the compat sweep is implemented", mode)
	}

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
		res, caseErr := gradeCase(ctx, filepath.Join(casesRoot, e.Name()), mode, led)
		if caseErr != nil {
			errs = append(errs, caseErr)
		}
		if !res.graded {
			// A case that could not be loaded, stubbed or replayed
			// contributes its errors and nothing else: the counters
			// report what was compared, and this case was not.
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
		if res.pending {
			summary.Pending = append(summary.Pending, res.caseID)
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
}

// gradeCase replays one case and compares every tier-1 field of its
// record. It keeps comparing after a disagreement rather than returning
// the first, so one systemic defect reports its full count in a single
// run.
func gradeCase(ctx context.Context, dir string, mode GradeMode, led ledgers) (res caseGradeResult, err error) {
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

	evalDate, dateErr := time.Parse(time.DateOnly, c.EvalDate)
	if dateErr != nil {
		errs = append(errs, fmt.Errorf("case %s: eval_date %q: %w", c.ID, c.EvalDate, dateErr))
		return res, nil
	}
	tun, tunErr := gradeTunables(c.Profile)
	if tunErr != nil {
		errs = append(errs, fmt.Errorf("case %s: %w", c.ID, tunErr))
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

	client, clientErr := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   "http://" + srv.Addr(),
		UserAgent: gradeUserAgent,
		Budget:    musicbrainz.Unmetered(),
	})
	if clientErr != nil {
		errs = append(errs, fmt.Errorf("case %s: build client: %w", c.ID, clientErr))
		return res, nil
	}

	items := make([]matcher.ItemFields, len(c.Input.Items))
	for i, raw := range c.Input.Items {
		items[i] = matcher.ProjectItemFields(raw)
	}

	prop, tagErr := tag.Album(ctx, client, items, tun, compatFor(mode, policy), evalDate, tag.Options{})
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

	if mode == ModeCompat {
		errs = append(errs, checkRequestSet(c.ID, srv)...)
	}

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

	return res, nil
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
// oracle-wrong hypotheses have no fix a rung-3 port can express.
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
