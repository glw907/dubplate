package conform

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// repoRootPath is repoRoot's *testing.T-free sibling: it walks up from
// the working directory to the nearest ancestor holding go.mod. The
// shared sweep below runs inside a sync.OnceValues, outside any one
// test's t, so the walk needs a form that returns an error rather than
// failing a test that may not be the one waiting on it.
func repoRootPath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("locate repo root: no go.mod above the working directory")
		}
		dir = parent
	}
}

// loadEveryCase loads every committed corpus case under root.
func loadEveryCase(t *testing.T, root string) []*cassette.Case {
	t.Helper()
	dirs := corpusCaseDirs(t, root)
	cases := make([]*cassette.Case, 0, len(dirs))
	for _, dir := range dirs {
		c, err := cassette.LoadCase(dir)
		if err != nil {
			t.Fatalf("load %s: %v", dir, err)
		}
		cases = append(cases, c)
	}
	return cases
}

// fullCompat and noCompat are the two matcher.Compat literals the
// policy table names, spelled here so a row reads as a selection rather
// than as two bools.
var (
	fullCompat = matcher.Compat{RemixParenthetical: true}
	noCompat   = matcher.Compat{}
)

// loadPolicyFixture reads one hand-authored policy fixture: the four
// analysis ledgers and the case id the row is about, so a policy is
// resolved from exactly the inputs the real sweep resolves it from. A
// ledger a row has nothing to say about is omitted from the fixture
// directory and loads empty.
func loadPolicyFixture(t *testing.T, name string) (*analysis.TieLedger, *analysis.ProximityLedger,
	*analysis.AdjudicationLedger, *analysis.OracleBugLedger, string) {
	t.Helper()
	dir := filepath.Join("testdata", "case", "policy", strings.ReplaceAll(name, " ", "-"))

	ties := &analysis.TieLedger{}
	prox := &analysis.ProximityLedger{}
	adj := &analysis.AdjudicationLedger{}
	bugs := &analysis.OracleBugLedger{}
	loadIfPresent(t, dir, "ties.json", ties)
	loadIfPresent(t, dir, "proximity.json", prox)
	loadIfPresent(t, dir, "adjudications.json", adj)
	loadIfPresent(t, dir, "oracle_bugs.json", bugs)

	var subject struct {
		CaseID string `json:"case_id"`
	}
	loadIfPresent(t, dir, "subject.json", &subject)
	if subject.CaseID == "" {
		t.Fatalf("%s: subject.json names no case id", dir)
	}
	return ties, prox, adj, bugs, subject.CaseID
}

func loadIfPresent(t *testing.T, dir, name string, into any) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, name))
	if errors.Is(err, os.ErrNotExist) {
		return
	}
	if err != nil {
		t.Fatalf("read %s/%s: %v", dir, name, err)
	}
	if err := json.Unmarshal(data, into); err != nil {
		t.Fatalf("parse %s/%s: %v", dir, name, err)
	}
}

// TestPolicyForHonorsEveryDispositionRow walks the rung 3 spec's
// disposition table (:285-299) row by row, plus the adjudications
// pending path (:304-314), the half-on Compat shape, and both
// precedence shapes (:302), against hand-authored ledgers in
// testdata/case/policy. Eight of the sixteen rows have ZERO committed
// corpus instances, which is exactly why the fixtures are hand-authored
// rather than borrowed from a case.
//
// Every field of every want literal is written out, including the ones
// that happen to equal the zero value, because the comparison is struct
// equality and an omitted field is a silent assertion that it is false.
func TestPolicyForHonorsEveryDispositionRow(t *testing.T) {
	tests := []struct {
		name string
		want casePolicy
	}{
		{"proximity clear", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"proximity flagged", casePolicy{DistanceCompare: compareBitIdentical, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"proximity skipped", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: false, GradeAlbumDistance: false, GradeVerdict: true, ExpectNoCandidate: true, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties unique", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties pinned", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties excluded", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: false, GradeAlbumDistance: false, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: true, Pending: false, VerdictFrom: verdictFromRecord}},
		{"ties skipped", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: false, GradeAlbumDistance: false, GradeVerdict: true, ExpectNoCandidate: true, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"oracle bug pending", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: fullCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"oracle bug oracle wrong", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromAdjudication}},
		{"oracle bug no correct answer", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: false, ExpectNoCandidate: false, Compat: fullCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"oracle bug oracle correct", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"adjudication no correct answer", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: false, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"adjudication pending", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: false, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: true, VerdictFrom: verdictFromRecord}},
		{"case row oracle wrong under a pending hypothesis", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: fullCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
		{"probe row oracle correct under an oracle wrong hypothesis", casePolicy{DistanceCompare: compareEpsilon, GradeAssignment: true, GradeAlbumDistance: true, GradeVerdict: true, ExpectNoCandidate: false, Compat: noCompat, Excluded: false, Pending: false, VerdictFrom: verdictFromRecord}},
	}
	if len(tests) != 15 {
		t.Fatalf("the table carries %d rows; it must carry one per spec disposition row (12), the adjudications-pending path, and both precedence shapes", len(tests))
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ties, prox, adj, bugs, caseID := loadPolicyFixture(t, tt.name)
			got := policyFor(caseID, ties, prox, adj, bugs)
			if got != tt.want {
				t.Errorf("policyFor(%s) = %+v, want %+v", caseID, got, tt.want)
			}
		})
	}
}

// TestCheckTiesAwaitingRulingRefusesASilentEscape proves ruling F's
// escape cannot be taken without its finding document: a non-empty
// awaiting list fails while the document is missing and passes once it
// exists. Both roots are hand-built here, so the assertion does not
// turn on whether the real repo carries the document today (it does
// not, because the mirroring attempt succeeded).
func TestCheckTiesAwaitingRulingRefusesASilentEscape(t *testing.T) {
	bare := t.TempDir()
	withFinding := t.TempDir()
	path := filepath.Join(withFinding, tieMirroringFinding)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("# fixture finding\n"), 0o644); err != nil {
		t.Fatalf("write finding: %v", err)
	}

	tests := []struct {
		name     string
		root     string
		awaiting []string
		wantErr  bool
	}{
		{"nothing awaiting, no finding", bare, nil, false},
		{"awaiting with no finding", bare, []string{"should-ask", "should-reject"}, true},
		{"awaiting with the finding", withFinding, []string{"should-ask", "should-reject"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkTiesAwaitingRuling(tt.root, tt.awaiting)
			if (err != nil) != tt.wantErr {
				t.Fatalf("checkTiesAwaitingRuling = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "should-ask") {
				t.Errorf("error does not name the awaiting rows: %v", err)
			}
		})
	}
}

// gradeFixtureCase loads a single hand-authored case directory under
// testdata/case/verdict/name (case.json, its responses/, and any of
// the four analysis ledgers a row has something to say about) and
// grades it exactly as GradeCases' own ModeCompat sweep would grade
// one case, without needing a whole fixture corpus tree.
func gradeFixtureCase(t *testing.T, name string) (CaseGradeSummary, error) {
	t.Helper()
	dir := filepath.Join("testdata", "case", "verdict", strings.ReplaceAll(name, " ", "-"))

	ties := &analysis.TieLedger{}
	prox := &analysis.ProximityLedger{}
	adj := &analysis.AdjudicationLedger{}
	bugs := &analysis.OracleBugLedger{}
	loadIfPresent(t, dir, "ties.json", ties)
	loadIfPresent(t, dir, "proximity.json", prox)
	loadIfPresent(t, dir, "adjudications.json", adj)
	loadIfPresent(t, dir, "oracle_bugs.json", bugs)
	led := ledgers{ties: ties, prox: prox, adj: adj, bugs: bugs}

	res, err := gradeCase(context.Background(), dir, led)
	if err != nil {
		return CaseGradeSummary{}, err
	}

	var summary CaseGradeSummary
	summary.Verdicts = res.verdicts
	if res.pending {
		summary.Pending = append(summary.Pending, res.caseID)
	}
	return summary, nil
}

// compatSummary runs the ModeCompat sweep once for the whole package,
// because more than one test reads it and the sweep starts a stub and
// replays a flow per case.
var compatSummary = sync.OnceValues(func() (CaseGradeSummary, error) {
	root, err := repoRootPath()
	if err != nil {
		return CaseGradeSummary{}, err
	}
	return GradeCases(root, ModeCompat)
})

func TestWholeCorpusGradesCleanlyInCompatMode(t *testing.T) {
	root := repoRoot(t)
	summary, err := compatSummary()
	if err != nil {
		t.Fatalf("GradeCases: %v", err)
	}
	// The assignment floor is the only one that moves, and it moves for
	// exactly two owner-owned reasons: ruling F's escalation parks the
	// pinned cases in TiesAwaitingRuling, and an owner ruling can
	// exclude a case. Subtracting both keeps the gate green on ruling
	// F's pre-authorized path and still fires on a shrink from any other
	// cause.
	assignmentFloor := AssignmentFloor - len(summary.TiesAwaitingRuling) - summary.TiesExcluded
	for _, f := range []struct {
		name  string
		got   int
		floor int
	}{
		{"cases", summary.Cases, CaseFloor},
		{"candidates", summary.Candidates, CandidateFloor},
		{"distances", summary.Distances, DistanceFloor},
		{"track vectors", summary.TrackVectors, TrackVectorFloor},
		{"assignments", summary.Assignments, assignmentFloor},
	} {
		if f.got < f.floor {
			t.Errorf("compared %d %s, floor is %d (assignments: %d candidate-bearing cases less %d awaiting ruling F and %d excluded): a comparison went missing",
				f.got, f.name, f.floor, AssignmentFloor, len(summary.TiesAwaitingRuling), summary.TiesExcluded)
		}
	}
	if summary.TiesExcluded > TiesExcludedCeiling {
		t.Errorf("ties.json carries %d excluded rows, ceiling is %d; excluding a case stops grading it and disposition is owner-only (spec :292)", summary.TiesExcluded, TiesExcludedCeiling)
	}
	if err := checkTiesAwaitingRuling(root, summary.TiesAwaitingRuling); err != nil {
		t.Error(err)
	}
	if len(summary.TiesAwaitingRuling) > 0 {
		t.Logf("awaiting Geoff's ruling F disposition, graded as pending rather than passing: %v", summary.TiesAwaitingRuling)
	}
	for _, nm := range summary.NearMisses {
		t.Logf("near miss: %s %s want %.17g got %.17g", nm.CaseID, nm.Field, nm.Want, nm.Got)
	}
	t.Logf("compared %d cases, %d candidates, %d album distances, %d track vectors, %d assignments",
		summary.Cases, summary.Candidates, summary.Distances, summary.TrackVectors, summary.Assignments)
}

// TestNoCommittedCaseHasTwoExtraItemsSharingASortKey guards beets'
// extra_items sort key. Its pre-sort order is a Python set order over
// identity-hashed objects, outside PYTHONHASHSEED's reach, so every
// committed record is reproducible only while (disc, track, title) is
// total over that case's extras. It is today, on all 215. The day lane
// E1 lands a case where it is not, beets' own record stops being
// reproducible run to run, and this is where that is learned rather
// than at a flaky gate.
//
// extra_tracks needs no such guard: album_info renumbers every track
// 1..n across the whole release (beetsplug/musicbrainz.py:645-646), so
// t.index is unique and (index, title) can never tie.
func TestNoCommittedCaseHasTwoExtraItemsSharingASortKey(t *testing.T) {
	root := repoRoot(t)
	checked := 0
	for _, c := range loadEveryCase(t, root) {
		if c.Expected == nil || len(c.Expected.ExtraItems) < 2 {
			continue
		}
		checked++
		seen := make(map[[3]string]bool, len(c.Expected.ExtraItems))
		for _, idx := range c.Expected.ExtraItems {
			it := matcher.ProjectItemFields(c.Input.Items[idx])
			k := [3]string{strconv.Itoa(it.Disc), strconv.Itoa(it.Track), it.Title}
			if seen[k] {
				t.Errorf("%s: two extra items share the sort key %v; beets' own record is no longer reproducible run to run", c.ID, k)
			}
			seen[k] = true
		}
	}
	if checked == 0 {
		t.Error("no case carried two or more extra items, so this guard asserted nothing; three cases do today")
	}
	t.Logf("checked the extra-items sort key over %d cases carrying two or more extras", checked)
}

// defaultSummary runs the ModeDefault sweep once for the whole package,
// the same one-run-shared-by-many-tests shape as compatSummary: it
// re-scores every case's committed candidate set, which is cheap
// (fold B2), but there is no reason to pay for it twice.
var defaultSummary = sync.OnceValues(func() (CaseGradeSummary, error) {
	root, err := repoRootPath()
	if err != nil {
		return CaseGradeSummary{}, err
	}
	return GradeCases(root, ModeDefault)
})

func TestWholeCorpusGradesVerdictsInCompatMode(t *testing.T) {
	// compatSummary is a package-level sync.OnceValues over
	// GradeCases(repoRoot, ModeCompat), so Task 7's sweep test and this
	// one share ONE run rather than paying for the corpus twice in the
	// same package (fold m8). Task 7's test switches to it in the same
	// change. ModeDefault stays its own run and is cheap, because it
	// re-scores committed candidates and issues no requests (fold B2).
	summary, err := compatSummary()
	if err != nil {
		t.Fatalf("GradeCases: %v", err)
	}
	if summary.Verdicts < VerdictFloor-len(summary.Pending)-len(summary.Ungraded) {
		t.Errorf("graded %d verdicts, want at least %d (floor %d less %d pending and %d ungraded)",
			summary.Verdicts, VerdictFloor-len(summary.Pending)-len(summary.Ungraded), VerdictFloor, len(summary.Pending), len(summary.Ungraded))
	}
	for _, id := range summary.Pending {
		t.Logf("pending divergence, verdict ungraded, launch gate refuses: %s", id)
	}
}

func TestWholeCorpusGradesTheCorrectedVerdictsInDefaultMode(t *testing.T) {
	summary, err := defaultSummary()
	if err != nil {
		t.Fatalf("GradeCases: %v", err)
	}
	if summary.Verdicts != DefectFreeFloor {
		t.Errorf("default mode graded %d verdicts, want exactly the %d defect_free rows: no other case has a corrected expectation", summary.Verdicts, DefectFreeFloor)
	}
	want := ProbeCaseFloor - ProbeCaseOracleCorrectRows
	if got := len(summary.Ungraded) + summary.Verdicts; got < want {
		t.Errorf("graded plus ungraded probe cases = %d, want at least %d (%d probe cases less the %d whose own row reads oracle-correct); a probe case fell out of both populations",
			got, want, ProbeCaseFloor, ProbeCaseOracleCorrectRows)
	}
	t.Logf("default mode moved %d non-probe verdicts: %v", len(summary.VerdictMoves), summary.VerdictMoves)
}

func TestCheckDefaultCompatMatchesTheLedger(t *testing.T) {
	if err := CheckDefaultCompatMatchesTheLedger(repoRoot(t)); err != nil {
		t.Errorf("matcher.DefaultCompat no longer agrees with corpus/analysis/oracle_bugs.json: %v", err)
	}
}

// TestPendingDivergenceLeavesTheGateGreenAndTheVerdictUngraded proves
// the pending escape: a case whose own adjudications.json row already
// reads pending-adjudication leaves the verdict ungraded rather than
// failed, even though its Go verdict deliberately differs from its
// record. Zero committed rows read pending-adjudication, so this is a
// hand-authored fixture in testdata/case/verdict/pending.
func TestPendingDivergenceLeavesTheGateGreenAndTheVerdictUngraded(t *testing.T) {
	summary, err := gradeFixtureCase(t, "pending")
	if err != nil {
		t.Fatalf("grade: %v", err)
	}
	if !slices.Contains(summary.Pending, "fixture-pending") {
		t.Errorf("Pending = %v, want the case id", summary.Pending)
	}
	if summary.Verdicts != 0 {
		t.Errorf("graded %d verdicts, want 0: a pending row leaves the verdict ungraded", summary.Verdicts)
	}
}

// TestAnUnadjudicatedDivergenceFailsAndPrintsTheRowAHumanMustWrite
// proves the gate refuses to absorb a mechanical verdict disagreement
// that carries no adjudication row at all (rung 3 spec, :311).
func TestAnUnadjudicatedDivergenceFailsAndPrintsTheRowAHumanMustWrite(t *testing.T) {
	_, err := gradeFixtureCase(t, "unadjudicated")
	if err == nil {
		t.Fatal("an unadjudicated verdict divergence passed; a mechanical disagreement is a Go defect until an owner says otherwise (spec :311)")
	}
	for _, want := range []string{"fixture-unadjudicated", "pending-adjudication", `"adjudicator": "agent"`} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %v does not carry %q; the message is what a human pastes into adjudications.json", err, want)
		}
	}
}
