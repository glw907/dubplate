package shadow

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// libraryReader builds a reconcileInput.LibraryReadFn over a fixed map
// of library paths to their decoded fields, so a reconcile test needs
// no library tree on disk.
func libraryReader(files map[string]matcher.ItemFields) func(string) (matcher.ItemFields, error) {
	return func(path string) (matcher.ItemFields, error) {
		fields, ok := files[path]
		if !ok {
			return matcher.ItemFields{}, fmt.Errorf("no fixture library file at %s", path)
		}
		return fields, nil
	}
}

// TestReconcileJoinsAnImportedRowByContentNotBySourceDir proves fold
// B2's finding mechanically: two captures under one uploader, both
// imported in one report that gives them the SAME source_dir, which is
// exactly the shape stageAllContributors produces
// (importer.go:782-800). A source_dir join would map every fixture to
// every outcome; the content join must not.
func TestReconcileJoinsAnImportedRowByContentNotBySourceDir(t *testing.T) {
	alpha := Capture{
		Uploader: "ada", Basename: "alpha", CapturedAt: passNow,
		Items: []matcher.ItemFields{{Title: "Alpha One", Length: 200}},
	}
	beta := Capture{
		Uploader: "ada", Basename: "beta", CapturedAt: passNow,
		Items: []matcher.ItemFields{{Title: "Beta One", Length: 180}},
	}

	libraryFiles := map[string]matcher.ItemFields{
		"/library/alpha/01.flac": {Title: "Alpha One", Length: 200.4, MBAlbumID: "mbid-alpha"},
		"/library/beta/01.flac":  {Title: "Beta One", Length: 179.6, MBAlbumID: "mbid-beta"},
	}

	report := oracleRunReport{
		Version: 2, RunID: "run1", StartedAt: passNow, EndedAt: passNow.Add(time.Minute),
		Albums: []oracleAlbumOutcome{
			{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeImported, LibraryPaths: []string{"/library/alpha/01.flac"}},
			{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeImported, LibraryPaths: []string{"/library/beta/01.flac"}},
		},
	}

	got, err := reconcile(reconcileInput{
		Captures:      []Capture{alpha, beta},
		Reports:       []oracleRunReport{report},
		LibraryReadFn: libraryReader(libraryFiles),
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(got.Reconciliations) != 2 {
		t.Fatalf("reconciliations = %+v, want 2", got.Reconciliations)
	}
	for _, r := range got.Reconciliations {
		if r.Unreconciled != "" || r.Excluded != "" {
			t.Errorf("reconciliation = %+v, want neither unreconciled nor excluded", r)
		}
		switch r.Capture.Basename {
		case "alpha":
			if r.Oracle == nil || r.Oracle.ReleaseMBID != "mbid-alpha" {
				t.Errorf("alpha reconciliation = %+v, want release mbid-alpha", r)
			}
		case "beta":
			if r.Oracle == nil || r.Oracle.ReleaseMBID != "mbid-beta" {
				t.Errorf("beta reconciliation = %+v, want release mbid-beta", r)
			}
		default:
			t.Errorf("unexpected capture in reconciliation: %+v", r.Capture)
		}
	}
}

func TestReconcileResolvesASkipFromEitherBackstop(t *testing.T) {
	tests := []struct {
		name   string
		tagLog []string
		review []string
		wantOK bool
	}{
		{name: "tag log alone", tagLog: []string{"/srv/staging/ada/windrose"}, wantOK: true},
		{name: "review tree alone", review: []string{"windrose"}, wantOK: true},
		{name: "review tree with an HHMMSS collision suffix", review: []string{"windrose-142530"}, wantOK: true},
		{name: "review tree with a numeric collision suffix", review: []string{"windrose-142530.2"}, wantOK: true},
		{name: "a different album's directory does not resolve it", review: []string{"windrose-live"}, wantOK: false},
		{name: "neither backstop", wantOK: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			capture := Capture{Uploader: "ada", Basename: "windrose", CapturedAt: passNow}

			tagLogSkips := make(map[string]bool)
			for _, p := range tt.tagLog {
				tagLogSkips[filepath.Base(p)] = true
			}
			var reviewed map[string][]string
			if len(tt.review) > 0 {
				reviewed = map[string][]string{"ada": tt.review}
			}

			got, err := reconcile(reconcileInput{
				Captures:      []Capture{capture},
				TagLogSkips:   tagLogSkips,
				Reviewed:      reviewed,
				LibraryReadFn: libraryReader(nil),
			})
			if err != nil {
				t.Fatalf("reconcile: %v", err)
			}

			if tt.wantOK {
				if len(got.Reconciliations) != 1 || got.Reconciliations[0].Oracle == nil || got.Reconciliations[0].Oracle.Signal != "skip" {
					t.Fatalf("reconciliations = %+v, want one resolved skip", got.Reconciliations)
				}
				return
			}
			if len(got.Reconciliations) != 0 {
				t.Fatalf("reconciliations = %+v, want none", got.Reconciliations)
			}
		})
	}
}

func TestReconcilePartialSkipIsUnreconciledNotASkip(t *testing.T) {
	capture := Capture{Uploader: "ada", Basename: "windrose", CapturedAt: passNow}
	report := oracleRunReport{
		Version: 2, RunID: "run1", StartedAt: passNow, EndedAt: passNow.Add(time.Minute),
		Albums: []oracleAlbumOutcome{
			{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeReview, Error: "left in staging after import"},
		},
	}

	got, err := reconcile(reconcileInput{
		Captures:      []Capture{capture},
		Reports:       []oracleRunReport{report},
		TagLogSkips:   map[string]bool{"windrose": true},
		LibraryReadFn: libraryReader(nil),
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(got.Reconciliations) != 1 || got.Reconciliations[0].Unreconciled != UnreconciledPartialSkip {
		t.Fatalf("reconciliations = %+v, want one UnreconciledPartialSkip", got.Reconciliations)
	}
	if got.Reconciliations[0].Oracle != nil {
		t.Errorf("oracle = %+v, want nil; a partial skip must never fabricate a divergence", got.Reconciliations[0].Oracle)
	}
}

func TestReconcileExcludesGateOutcomes(t *testing.T) {
	report := oracleRunReport{
		Version: 2, RunID: "run1", StartedAt: passNow, EndedAt: passNow.Add(time.Minute),
		Albums: []oracleAlbumOutcome{
			{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeQuarantined},
			{Uploader: "bea", SourceDir: "/inbox/bea", Outcome: outcomeRejected},
		},
	}

	got, err := reconcile(reconcileInput{
		Reports:       []oracleRunReport{report},
		LibraryReadFn: libraryReader(nil),
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(got.Reconciliations) != 2 {
		t.Fatalf("reconciliations = %+v, want 2", got.Reconciliations)
	}
	for _, r := range got.Reconciliations {
		if r.Excluded != ExcludedGateRejected {
			t.Errorf("reconciliation = %+v, want ExcludedGateRejected", r)
		}
		if r.Oracle != nil {
			t.Errorf("reconciliation = %+v, want no oracle signal; a gate outcome never reached beets", r)
		}
	}
}

func TestReconcileUnreconciledReasons(t *testing.T) {
	t.Run("an imported row with no library paths", func(t *testing.T) {
		capture := Capture{Uploader: "ada", Basename: "windrose", CapturedAt: passNow}
		report := oracleRunReport{
			Version: 2, RunID: "run1", StartedAt: passNow, EndedAt: passNow.Add(time.Minute),
			Albums: []oracleAlbumOutcome{
				{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeImported, Error: "a beets-database match that never turned up"},
			},
		}

		got, err := reconcile(reconcileInput{
			Captures:      []Capture{capture},
			Reports:       []oracleRunReport{report},
			LibraryReadFn: libraryReader(nil),
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(got.Reconciliations) != 1 || got.Reconciliations[0].Unreconciled != UnreconciledNoReleaseID {
			t.Fatalf("reconciliations = %+v, want one UnreconciledNoReleaseID", got.Reconciliations)
		}
	})

	t.Run("an imported row whose library file carries no album id", func(t *testing.T) {
		capture := Capture{
			Uploader: "ada", Basename: "windrose", CapturedAt: passNow,
			Items: []matcher.ItemFields{{Title: "One", Length: 200}},
		}
		report := oracleRunReport{
			Version: 2, RunID: "run1", StartedAt: passNow, EndedAt: passNow.Add(time.Minute),
			Albums: []oracleAlbumOutcome{
				{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeImported, LibraryPaths: []string{"/library/windrose/01.flac"}},
			},
		}

		got, err := reconcile(reconcileInput{
			Captures: []Capture{capture},
			Reports:  []oracleRunReport{report},
			LibraryReadFn: libraryReader(map[string]matcher.ItemFields{
				"/library/windrose/01.flac": {Title: "One", Length: 200.2},
			}),
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(got.Reconciliations) != 1 || got.Reconciliations[0].Unreconciled != UnreconciledNoReleaseID {
			t.Fatalf("reconciliations = %+v, want one UnreconciledNoReleaseID", got.Reconciliations)
		}
		if got.Reconciliations[0].Oracle != nil {
			t.Errorf("oracle = %+v, want nil; an empty release must never classify as go-applies-different-release", got.Reconciliations[0].Oracle)
		}
	})

	t.Run("two uploaders with an identically named album in one report", func(t *testing.T) {
		adaCapture := Capture{Uploader: "ada", Basename: "windrose", CapturedAt: passNow}
		beaCapture := Capture{Uploader: "bea", Basename: "windrose", CapturedAt: passNow}

		got, err := reconcile(reconcileInput{
			Captures:      []Capture{adaCapture, beaCapture},
			TagLogSkips:   map[string]bool{"windrose": true},
			LibraryReadFn: libraryReader(nil),
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(got.Reconciliations) != 2 {
			t.Fatalf("reconciliations = %+v, want 2", got.Reconciliations)
		}
		for _, r := range got.Reconciliations {
			if r.Unreconciled != UnreconciledAmbiguousBasename {
				t.Errorf("reconciliation = %+v, want UnreconciledAmbiguousBasename", r)
			}
		}
	})

	t.Run("a capture whose report was swept", func(t *testing.T) {
		old := Capture{Uploader: "ada", Basename: "windrose", CapturedAt: passNow.Add(-48 * time.Hour)}
		consumed := []ConsumedReport{{
			File:      "report-old.json",
			StartedAt: passNow.Add(-72 * time.Hour),
			EndedAt:   passNow.Add(-71 * time.Hour),
		}}
		report := oracleRunReport{Version: 2, RunID: "new", StartedAt: passNow.Add(-time.Hour), EndedAt: passNow}

		got, err := reconcile(reconcileInput{
			Captures:      []Capture{old},
			Reports:       []oracleRunReport{report},
			Consumed:      consumed,
			LibraryReadFn: libraryReader(nil),
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if got.Gaps != 1 {
			t.Fatalf("gaps = %d, want 1", got.Gaps)
		}
		if len(got.Reconciliations) != 1 || got.Reconciliations[0].Unreconciled != UnreconciledReportSwept {
			t.Fatalf("reconciliations = %+v, want one UnreconciledReportSwept", got.Reconciliations)
		}
	})

	t.Run("a review row carrying an error", func(t *testing.T) {
		capture := Capture{Uploader: "ada", Basename: "windrose", CapturedAt: passNow}
		report := oracleRunReport{
			Version: 2, RunID: "run1", StartedAt: passNow, EndedAt: passNow.Add(time.Minute),
			Albums: []oracleAlbumOutcome{
				{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeReview, Error: "left in staging after import"},
			},
		}

		got, err := reconcile(reconcileInput{
			Captures:      []Capture{capture},
			Reports:       []oracleRunReport{report},
			TagLogSkips:   map[string]bool{"windrose": true},
			LibraryReadFn: libraryReader(nil),
		})
		if err != nil {
			t.Fatalf("reconcile: %v", err)
		}
		if len(got.Reconciliations) != 1 || got.Reconciliations[0].Unreconciled != UnreconciledPartialSkip {
			t.Fatalf("reconciliations = %+v, want one UnreconciledPartialSkip", got.Reconciliations)
		}
	})
}

// TestReconcileSupersedesAnOlderCaptureOfOneAlbum proves Resolution R5's
// supersede rule: two unreconciled captures of ada/windrose, at
// different digests and CapturedAt values, with one imported row. The
// newer reconciles to the apply and the older is marked Superseded
// rather than either being reported as ambiguous.
func TestReconcileSupersedesAnOlderCaptureOfOneAlbum(t *testing.T) {
	older := Capture{
		Uploader: "ada", Basename: "windrose", Digest: "digest-old", CapturedAt: passNow.Add(-time.Hour),
		Items: []matcher.ItemFields{{Title: "One", Length: 200}},
	}
	newer := Capture{
		Uploader: "ada", Basename: "windrose", Digest: "digest-new", CapturedAt: passNow,
		Items: []matcher.ItemFields{{Title: "One", Length: 200}},
	}
	report := oracleRunReport{
		Version: 2, RunID: "run1", StartedAt: passNow, EndedAt: passNow.Add(time.Minute),
		Albums: []oracleAlbumOutcome{
			{Uploader: "ada", SourceDir: "/inbox/ada", Outcome: outcomeImported, LibraryPaths: []string{"/library/windrose/01.flac"}},
		},
	}

	got, err := reconcile(reconcileInput{
		Captures: []Capture{older, newer},
		Reports:  []oracleRunReport{report},
		LibraryReadFn: libraryReader(map[string]matcher.ItemFields{
			"/library/windrose/01.flac": {Title: "One", Length: 200.1, MBAlbumID: "mbid-1"},
		}),
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if len(got.Reconciliations) != 2 {
		t.Fatalf("reconciliations = %+v, want 2", got.Reconciliations)
	}

	var sawSuperseded, sawApplied bool
	for _, r := range got.Reconciliations {
		switch r.Capture.Digest {
		case "digest-old":
			if !r.Superseded {
				t.Errorf("older capture reconciliation = %+v, want Superseded", r)
			}
			sawSuperseded = true
		case "digest-new":
			if r.Oracle == nil || r.Oracle.ReleaseMBID != "mbid-1" {
				t.Errorf("newer capture reconciliation = %+v, want the apply", r)
			}
			sawApplied = true
		}
	}
	if !sawSuperseded || !sawApplied {
		t.Fatalf("reconciliations = %+v, want one superseded and one applied", got.Reconciliations)
	}
}

// TestReconcileDetectsAReportGap proves Resolution R6: the ledger's
// newest consumed EndedAt precedes the oldest present report's
// StartedAt, so Gaps == 1 and a still-unreconciled capture older than
// the boundary comes back UnreconciledReportSwept.
func TestReconcileDetectsAReportGap(t *testing.T) {
	consumed := []ConsumedReport{{
		File:      "report-old.json",
		StartedAt: passNow.Add(-3 * time.Hour),
		EndedAt:   passNow.Add(-2*time.Hour - 45*time.Minute),
	}}
	report := oracleRunReport{Version: 2, RunID: "new", StartedAt: passNow.Add(-time.Hour), EndedAt: passNow}
	stillCapture := Capture{Uploader: "ada", Basename: "windrose", CapturedAt: passNow.Add(-4 * time.Hour)}

	got, err := reconcile(reconcileInput{
		Captures:      []Capture{stillCapture},
		Reports:       []oracleRunReport{report},
		Consumed:      consumed,
		LibraryReadFn: libraryReader(nil),
	})
	if err != nil {
		t.Fatalf("reconcile: %v", err)
	}
	if got.Gaps != 1 {
		t.Fatalf("gaps = %d, want 1", got.Gaps)
	}
	if len(got.Reconciliations) != 1 || got.Reconciliations[0].Unreconciled != UnreconciledReportSwept {
		t.Fatalf("reconciliations = %+v, want one UnreconciledReportSwept", got.Reconciliations)
	}
	if len(got.Consumed) != 2 {
		t.Fatalf("consumed = %+v, want the old ledger entry plus the new report", got.Consumed)
	}
}
