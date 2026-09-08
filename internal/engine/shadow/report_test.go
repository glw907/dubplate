package shadow

import (
	"path/filepath"
	"testing"
	"time"
)

// seedReportStore builds a store carrying, across three pass records:
// one divergence of each class, one deferral of each reason, one
// excluded row of each reason, one unreconciled row of each of the
// four reasons, one report gap, and spends 4, 7 and 0 (summing to 11,
// not 3 and not the largest, which is what proves BuildReport sums
// rather than takes the last or the max). It also stores five
// observations: two agreements (one applied, one skipped) and three
// divergences, one per class, stamped under two different matcher
// commits.
func seedReportStore(t *testing.T) *Store {
	t.Helper()
	s, err := OpenStore(filepath.Join(t.TempDir(), "shadow"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	passes := []*PassRecord{
		{
			ID: "pass-1",
			Divergences: map[string]int{
				ClassGoAppliesOracleSkipped:    1,
				ClassGoSkipsOracleApplied:      1,
				ClassGoAppliesDifferentRelease: 1,
			},
			Deferred: map[string]int{
				DeferUnavailable:     1,
				DeferBudgetExhausted: 1,
				DeferLimiterClosed:   1,
				DeferUnclassified:    1,
			},
			Excluded: map[string]int{
				ExcludedMultiAlbum:   1,
				ExcludedGateRejected: 1,
				ExcludedNoTracks:     1,
			},
			Unreconciled: map[string]int{
				UnreconciledAmbiguousBasename: 1,
				UnreconciledNoReleaseID:       1,
				UnreconciledReportSwept:       1,
				UnreconciledPartialSkip:       1,
			},
			ReportGaps: 1,
			Budget:     BudgetRecord{Name: "shadow", Spent: 4},
		},
		{ID: "pass-2", Budget: BudgetRecord{Name: "shadow", Spent: 7}},
		{ID: "pass-3", Budget: BudgetRecord{Name: "shadow", Spent: 0}},
	}
	for _, p := range passes {
		if err := s.PutPass(p); err != nil {
			t.Fatalf("PutPass %s: %v", p.ID, err)
		}
	}

	observations := []*Observation{
		{ID: "obsagree1", Class: ClassAgreement, MatcherCommit: "aaa", Oracle: &OracleSignal{Signal: "apply"}},
		{ID: "obsagree2", Class: ClassAgreement, MatcherCommit: "aaa", Oracle: &OracleSignal{Signal: "skip"}},
		{ID: "obsdiv1", Class: ClassGoAppliesOracleSkipped, MatcherCommit: "aaa", Oracle: &OracleSignal{Signal: "skip"}},
		{ID: "obsdiv2", Class: ClassGoSkipsOracleApplied, MatcherCommit: "aaa", Oracle: &OracleSignal{Signal: "apply"}},
		{ID: "obsdiv3", Class: ClassGoAppliesDifferentRelease, MatcherCommit: "bbb", Oracle: &OracleSignal{Signal: "apply"}},
	}
	for _, o := range observations {
		if err := s.PutObservation(o); err != nil {
			t.Fatalf("PutObservation %s: %v", o.ID, err)
		}
	}

	return s
}

// TestBuildReportSumsTheStore proves every row is a sum across the
// store's pass records and observations, by value.
func TestBuildReportSumsTheStore(t *testing.T) {
	s := seedReportStore(t)
	now := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)

	r, err := BuildReport(s, now, 720*time.Hour)
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}

	if r.ObservedApplied != 3 {
		t.Errorf("ObservedApplied = %d, want 3", r.ObservedApplied)
	}
	if r.ObservedSkipped != 2 {
		t.Errorf("ObservedSkipped = %d, want 2", r.ObservedSkipped)
	}

	wantExcluded := map[string]int{ExcludedMultiAlbum: 1, ExcludedGateRejected: 1, ExcludedNoTracks: 1}
	if !mapsEqual(r.Excluded, wantExcluded) {
		t.Errorf("Excluded = %v, want %v", r.Excluded, wantExcluded)
	}

	wantUnreconciled := map[string]int{
		UnreconciledAmbiguousBasename: 1,
		UnreconciledNoReleaseID:       1,
		UnreconciledReportSwept:       1,
		UnreconciledPartialSkip:       1,
	}
	if !mapsEqual(r.Unreconciled, wantUnreconciled) {
		t.Errorf("Unreconciled = %v, want %v", r.Unreconciled, wantUnreconciled)
	}

	if r.ReportGaps != 1 {
		t.Errorf("ReportGaps = %d, want 1", r.ReportGaps)
	}

	wantDeferred := map[string]int{
		DeferUnavailable:     1,
		DeferBudgetExhausted: 1,
		DeferLimiterClosed:   1,
		DeferUnclassified:    1,
	}
	if !mapsEqual(r.Deferred, wantDeferred) {
		t.Errorf("Deferred = %v, want %v", r.Deferred, wantDeferred)
	}

	wantDivergences := map[string]int{
		ClassGoAppliesOracleSkipped:    1,
		ClassGoSkipsOracleApplied:      1,
		ClassGoAppliesDifferentRelease: 1,
	}
	if !mapsEqual(r.Divergences, wantDivergences) {
		t.Errorf("Divergences = %v, want %v", r.Divergences, wantDivergences)
	}
	if r.GoAppliesOracleSkipped != 1 {
		t.Errorf("GoAppliesOracleSkipped = %d, want 1", r.GoAppliesOracleSkipped)
	}

	wantCommits := map[string]int{"aaa": 2, "bbb": 1}
	if !mapsEqual(r.MatcherCommits, wantCommits) {
		t.Errorf("MatcherCommits = %v, want %v", r.MatcherCommits, wantCommits)
	}

	if len(r.Passes) != 3 {
		t.Fatalf("Passes = %v, want 3 entries", r.Passes)
	}
	if r.BudgetTotal != 11 {
		t.Errorf("BudgetTotal = %d, want 11 (4+7+0, not 3 and not the largest)", r.BudgetTotal)
	}

	if r.StagedPending != 0 {
		t.Errorf("StagedPending = %d, want 0", r.StagedPending)
	}
	if r.Resolved != 0 {
		t.Errorf("Resolved = %d, want 0", r.Resolved)
	}
	if r.StoreBytes <= 0 {
		t.Errorf("StoreBytes = %d, want > 0 for a non-empty store", r.StoreBytes)
	}
}

// TestBuildReportElapsedIsMeasuredAgainstTheParameter proves elapsed is
// now minus the recorded clock start, and never derived from a clock
// this file reads: report.go calls time.Now nowhere.
func TestBuildReportElapsedIsMeasuredAgainstTheParameter(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "shadow"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	start := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	if _, err := s.StartClock(start); err != nil {
		t.Fatalf("StartClock: %v", err)
	}

	first := start.Add(24 * time.Hour)
	r1, err := BuildReport(s, first, 720*time.Hour)
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}
	if r1.Elapsed != 24*time.Hour {
		t.Errorf("Elapsed = %v, want 24h", r1.Elapsed)
	}

	second := start.Add(48 * time.Hour)
	r2, err := BuildReport(s, second, 720*time.Hour)
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}
	if r2.Elapsed != 48*time.Hour {
		t.Errorf("Elapsed = %v, want 48h", r2.Elapsed)
	}
	if r1.Elapsed == r2.Elapsed {
		t.Error("Elapsed did not change between two different now values")
	}
	if !r1.ClockStarted || !r2.ClockStarted {
		t.Error("ClockStarted = false, want true once StartClock has run")
	}
}

// TestBuildReportOnAnEmptyStore proves a fresh box's first report is
// readable rather than an error: every count zero, and the clock start
// reported as not started.
func TestBuildReportOnAnEmptyStore(t *testing.T) {
	s, err := OpenStore(filepath.Join(t.TempDir(), "shadow"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	r, err := BuildReport(s, time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC), 720*time.Hour)
	if err != nil {
		t.Fatalf("BuildReport: %v", err)
	}

	if r.ClockStarted {
		t.Error("ClockStarted = true, want false on an empty store")
	}
	if r.Elapsed != 0 {
		t.Errorf("Elapsed = %v, want 0", r.Elapsed)
	}
	if r.ObservedApplied != 0 || r.ObservedSkipped != 0 {
		t.Errorf("ObservedApplied/Skipped = %d/%d, want 0/0", r.ObservedApplied, r.ObservedSkipped)
	}
	if len(r.Excluded) != 0 || len(r.Unreconciled) != 0 || len(r.Deferred) != 0 || len(r.Divergences) != 0 || len(r.MatcherCommits) != 0 {
		t.Errorf("expected every map row empty, got Excluded=%v Unreconciled=%v Deferred=%v Divergences=%v MatcherCommits=%v",
			r.Excluded, r.Unreconciled, r.Deferred, r.Divergences, r.MatcherCommits)
	}
	if r.ReportGaps != 0 || r.GoAppliesOracleSkipped != 0 || r.StagedPending != 0 || r.Resolved != 0 || r.BudgetTotal != 0 {
		t.Error("expected every scalar row zero on an empty store")
	}
	if len(r.Passes) != 0 {
		t.Errorf("Passes = %v, want empty", r.Passes)
	}
}

func mapsEqual(a, b map[string]int) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}
