package shadow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// The release ids the thirteen-row fixture's WS/2 traffic touches. Each
// one that is actually looked up carries a committed testdata/observe
// body; obsRel13 is never requested at all (row 13's reserve fails
// before any request is issued) and obsRel03Oracle/obsRel05Oracle name
// releases the oracle applied that the Go matcher never looks up.
const (
	obsRel01       = "a0000000-0000-0000-0000-000000000001"
	obsRel03Oracle = "a0000000-0000-0000-0000-000000000030"
	obsRel04       = "a0000000-0000-0000-0000-000000000004"
	obsRel05Go     = "a0000000-0000-0000-0000-000000000050"
	obsRel05Oracle = "a0000000-0000-0000-0000-000000000051"
	obsRel12       = "a0000000-0000-0000-0000-000000000012"
	obsRel13       = "a0000000-0000-0000-0000-000000000013"
)

// observeServer answers WS/2 requests for the thirteen-row fixture from
// the committed testdata/observe bodies (rows 1, 4 and 5's own
// RecStrong-scoring documents, plus one shared empty search.json), with
// one status override: obsRel12 always answers 503 with Retry-After: 0,
// the collapse fold M-h's own servers use so a maxAttempts exhaustion
// costs no real wall time (mutation 4).
func observeServer(t *testing.T) *httptest.Server {
	t.Helper()

	dir := filepath.Join("testdata", "observe")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")

		var id string
		if r.URL.Path == "/ws/2/release" {
			id = "search"
		} else {
			id = strings.TrimPrefix(r.URL.Path, "/ws/2/release/")
		}
		if id == obsRel12 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		file := "release-" + id + ".json"
		if id == "search" {
			file = "search.json"
		}
		body, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) //nolint:errcheck // best-effort write to a test stub
	}))
	t.Cleanup(srv.Close)
	return srv
}

// buildInbox writes the thirteen-row fixture table (plus the one
// structurally necessary fourteenth directory the ambiguous-basename
// row needs; see the package doc comment on TestObservePassEndToEnd)
// and every artifact reconcile reads to resolve it: the incumbent's
// library files, tag log, review tree and one run report, plus a
// pre-existing consumed-report entry old enough to open the gap row 11
// needs.
//
// Every file writeAlbum and writeLibraryFile write is pinned to
// fixedModTime through os.Chtimes (mutation 6): the capture digest
// ingests mtime (digestFor, capture.go), so an unpinned file would make
// the observation ids, and therefore determinism and idempotence both,
// a function of when the test ran rather than of the fixture alone.
func buildInbox(t *testing.T, tr shadowTree, library string) {
	t.Helper()

	// 01: unanimous ID-consensus early return, agreement (apply/apply,
	// same release both sides).
	writeAlbum(t, tr.observe, "ada", "01-anvil", albumTracks(obsRel01, "Anvil", "One"))
	lib01 := writeLibraryFile(t, library, "ada-01-anvil.flac", obsRel01, "One")

	// 02: empty search returning tag.ErrNoCandidates, agreement
	// (skip/skip): no MUSICBRAINZ_ALBUMID tag, so match_by_id never
	// fires and the search comes back empty.
	writeAlbum(t, tr.observe, "ada", "02-basalt", albumTracks("", "Basalt", "Two"))

	// 03: go-skips-oracle-applied. The same empty-search shape as 02,
	// but the oracle applied a release the inbox tags never named.
	writeAlbum(t, tr.observe, "ada", "03-cobalt", albumTracks("", "Cobalt", "Three"))
	lib03 := writeLibraryFile(t, library, "ada-03-cobalt.flac", obsRel03Oracle, "Three")

	// 04: go-applies-oracle-skipped, the oracle's skip signal read
	// through the review tree's -HHMMSS collision-suffix form
	// (reviewSuffixPattern). It also carries a plain tag-log skip line
	// naming its own literal basename, so !tagLogSkips[c.Basename] is
	// false for it and it can never be mistaken for row 07's
	// gate-rejected exclusion regardless of the two rows' relative
	// order in the store's own id-sorted read (Store.Observations
	// sorts by digest, not by basename).
	writeAlbum(t, tr.observe, "ada", "04-diorite", albumTracks(obsRel04, "Diorite", "Four"))

	// 05: go-applies-different-release. The inbox tags carry one
	// unanimous release id and the incumbent's own library file names
	// another.
	writeAlbum(t, tr.observe, "ada", "05-ecolite", albumTracks(obsRel05Go, "Ecolite", "Five"))
	lib05 := writeLibraryFile(t, library, "ada-05-ecolite.flac", obsRel05Oracle, "Five")

	// 06: excluded multi-album directory. Two tracks group to two
	// distinct (artist, album) keys, so capture excludes the whole
	// directory before it ever becomes an observation.
	writeAlbum(t, tr.observe, "ada", "06-feldspar", [][]string{
		{"ALBUMARTIST=Granite", "ARTIST=Granite", "ALBUM=Feldspar A", "TITLE=Six A", "TRACKNUMBER=1"},
		{"ALBUMARTIST=Marble", "ARTIST=Marble", "ALBUM=Feldspar B", "TITLE=Six B", "TRACKNUMBER=1"},
	})

	// 07: excluded gate-rejected. Filed under the review tree at its
	// own plain basename and never tag-logged, so it is the only
	// capture that can ever satisfy resolveSkips' gate-rejected branch
	// (viaReview && !tagLogSkips && gateRejectedRemaining > 0).
	writeAlbum(t, tr.observe, "ada", "07-gneiss", albumTracks("", "Gneiss", "Seven"))
	writeReviewDir(t, tr.review, "2026-09-01", "ada", "07-gneiss")

	// 08: imported-with-no-release-id. The library file the incumbent
	// named carries no MUSICBRAINZ_ALBUMID, so the content join
	// succeeds but the release itself is empty (Resolution R22).
	writeAlbum(t, tr.observe, "ada", "08-hematite", albumTracks("", "Hematite", "Eight"))
	lib08 := writeLibraryFile(t, library, "ada-08-hematite.flac", "", "Eight")

	// 09: partial-skip, on its own uploader. partialSkipUploader is
	// scoped by uploader for the whole pass (Resolution R4), so a
	// partial-skip row sharing an uploader with an ordinary skip row
	// (02, 04) would silently relabel that row's clean skip too;
	// isolating it on "pk" is what keeps 02 and 04 honest agreements
	// and divergences instead.
	writeAlbum(t, tr.observe, "pk", "09-iolite", albumTracks("", "Iolite", "Nine"))

	// 10: ambiguous-basename, plus its structurally required partner.
	// The tag log's basenames carry no uploader (reconcile.go's own
	// resolveSkips doc comment), so ambiguity is two captures under
	// DIFFERENT uploaders sharing one basename; it cannot be produced
	// by two captures under a single uploader, since a filesystem
	// cannot hold two directories of the same name in one parent. The
	// partner lives on its own throwaway uploader "xx" purely to
	// manufacture the collision.
	writeAlbum(t, tr.observe, "ada", "10-jasper", albumTracks("", "Jasper", "Ten"))
	writeAlbum(t, tr.observe, "xx", "10-jasper", albumTracks("", "Jasper Partner", "Ten Partner"))

	// 11: report-swept. Named nowhere: not tag-logged, not filed under
	// review, not named in the run report below. Reconciled purely by
	// the gap: buildInbox seeds an old consumed-report entry whose
	// EndedAt precedes this pass's own report's StartedAt, and this
	// capture's CapturedAt (the pass clock) precedes that StartedAt in
	// turn.
	writeAlbum(t, tr.observe, "ada", "11-kyanite", albumTracks("", "Kyanite", "Eleven"))

	// 12: deferred, ErrUnavailable. A unanimous id consensus whose
	// lookup 503s for all seven attempts.
	writeAlbum(t, tr.observe, "ada", "12-lepidolite", albumTracks(obsRel12, "Lepidolite", "Twelve"))
	lib12 := writeLibraryFile(t, library, "ada-12-lepidolite.flac", obsRel12, "Twelve")

	// 13: deferred, ErrBudgetExhausted. Its own release id is never
	// requested: the pass's budget cap is sized to the exact sum spent
	// through album 12 (below), so this album's first reserve already
	// exceeds it.
	writeAlbum(t, tr.observe, "ada", "13-muscovite", albumTracks(obsRel13, "Muscovite", "Thirteen"))
	lib13 := writeLibraryFile(t, library, "ada-13-muscovite.flac", obsRel13, "Thirteen")

	writeTagLog(t, tr.oracle, []string{"02-basalt", "04-diorite", "09-iolite", "10-jasper"})
	writeReviewDir(t, tr.review, "2026-09-01", "ada", "04-diorite-153000")

	reportStartedAt := passNow.Add(time.Minute)
	writeRunReport(t, tr.oracle, oracleRunReport{
		RunID:     "run1",
		StartedAt: reportStartedAt,
		EndedAt:   reportStartedAt.Add(time.Minute),
		Albums: []oracleAlbumOutcome{
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib01}},
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib03}},
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib05}},
			{Uploader: "ada", Outcome: outcomeQuarantined},
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib08}},
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib12}},
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib13}},
			{Uploader: "pk", Outcome: outcomeReview, Error: "left in staging after import"},
		},
	})
}

// seedReportGap writes a consumed-report ledger entry old enough that
// buildInbox's own run report opens a gap: the ledger's newest EndedAt
// must precede the oldest report currently on disk's StartedAt
// (Resolution R6, reconcile.go).
func seedReportGap(t *testing.T, store *Store) {
	t.Helper()

	if err := store.PutConsumedReports([]ConsumedReport{{
		File:      "report-priorrun.json",
		StartedAt: passNow.Add(-3 * time.Hour),
		EndedAt:   passNow.Add(-90 * time.Minute),
	}}); err != nil {
		t.Fatalf("seed consumed reports: %v", err)
	}
}

// TestObservePassEndToEnd is the exit gate: one hand-authored,
// thirteen-row inbox under uploader "ada" exercising one of every shape
// contract S1 through S22 admits, run through the whole Run pipeline at
// once.
//
// The fixture carries a fourteenth directory beyond the thirteen the
// table names, under a throwaway uploader "xx": the ambiguous-basename
// row (10) is, by reconcile.go's own resolveSkips doc comment,
// necessarily a collision between two DIFFERENT uploaders' captures
// sharing one basename, which a single uploader's directory tree cannot
// produce (two directories cannot share one name under one parent).
// Stored observations therefore total thirteen, not twelve: the twelve
// the thirteen-row table itself accounts for (06 is excluded at capture
// and never becomes one) plus the ambiguous partner.
func TestObservePassEndToEnd(t *testing.T) {
	tr := newShadowTree(t)
	library := t.TempDir()
	store := tr.openStore(t)
	seedReportGap(t, store)
	buildInbox(t, tr, library)
	srv := observeServer(t)

	cfg := tr.config(srv.URL)
	cfg.BudgetLimit = 12 // the exact sum spent through album 12: 1+1+1+1+1+7 (mutation 3, 5)

	res, err := Run(context.Background(), cfg, passNow)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if !res.Pass.StartedAt.Equal(passNow) || !res.ClockStarted {
		t.Errorf("StartedAt = %s, ClockStarted = %v, want the pass clock and a started clock", res.Pass.StartedAt, res.ClockStarted)
	}

	wantAgreements := 2 // 01 (id-consensus), 02 (empty search)
	if res.Pass.Agreements != wantAgreements {
		t.Errorf("Agreements = %d, want %d", res.Pass.Agreements, wantAgreements)
	}

	wantDivergences := map[string]int{
		ClassGoSkipsOracleApplied:      1, // 03
		ClassGoAppliesOracleSkipped:    1, // 04
		ClassGoAppliesDifferentRelease: 1, // 05
	}
	for class, want := range wantDivergences {
		if res.Pass.Divergences[class] != want {
			t.Errorf("Divergences[%s] = %d, want %d (all: %v)", class, res.Pass.Divergences[class], want, res.Pass.Divergences)
		}
	}

	wantDeferred := map[string]int{
		DeferUnavailable:     1, // 12
		DeferBudgetExhausted: 1, // 13
	}
	for reason, want := range wantDeferred {
		if res.Pass.Deferred[reason] != want {
			t.Errorf("Deferred[%s] = %d, want %d (all: %v)", reason, res.Pass.Deferred[reason], want, res.Pass.Deferred)
		}
	}

	wantExcluded := map[string]int{
		ExcludedMultiAlbum:   1, // 06
		ExcludedGateRejected: 1, // 07
	}
	for reason, want := range wantExcluded {
		if res.Pass.Excluded[reason] != want {
			t.Errorf("Excluded[%s] = %d, want %d (all: %v)", reason, res.Pass.Excluded[reason], want, res.Pass.Excluded)
		}
	}

	wantUnreconciled := map[string]int{
		UnreconciledNoReleaseID:       1, // 08
		UnreconciledPartialSkip:       1, // 09
		UnreconciledAmbiguousBasename: 2, // 10 and its structural partner
		UnreconciledReportSwept:       1, // 11
	}
	for reason, want := range wantUnreconciled {
		if res.Pass.Unreconciled[reason] != want {
			t.Errorf("Unreconciled[%s] = %d, want %d (all: %v)", reason, res.Pass.Unreconciled[reason], want, res.Pass.Unreconciled)
		}
	}

	// mutation 3: a budget cap of zero means no cap at all
	// (musicbrainz/limiter.go:130-133), so every album would succeed
	// and 12 and 13 would never defer. This positive cap, sized to the
	// exact sum spent through album 12, is what makes the assertion
	// above discriminating.
	if res.Pass.Budget.Spent != 12 {
		t.Errorf("Budget.Spent = %d, want 12 (1 each for 01-05, 7 for 12's exhausted retries, 0 for 13's blocked reserve)", res.Pass.Budget.Spent)
	}

	if len(res.Pending) != 3 {
		t.Fatalf("staged %d pending rows, want 3 (one per divergence: 03, 04, 05)", len(res.Pending))
	}
	for _, row := range res.Pending {
		if row.ReleaseMBID == "" {
			t.Errorf("pending row %+v carries no release_mbid", row)
		}
	}

	stored, err := store.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	// Thirteen: the twelve captured, non-excluded-at-capture rows the
	// table names (13 rows minus 06, which capture excludes before it
	// ever becomes an observation) plus the ambiguous-basename partner
	// on uploader "xx" (see the function doc comment above).
	if len(stored) != 13 {
		t.Fatalf("stored %d observations, want 13", len(stored))
	}
}

// TestObservePassDefersEverythingOnAClosedLimiter proves Resolution R7
// separately from the exit gate above: ErrLimiterClosed and
// ErrBudgetExhausted are mutually exclusive within one pass, because
// Budget.Wait reserves before it waits (limiter.go:372-381), so a
// budget this test never caps still defers every album through the
// closed limiter alone.
func TestObservePassDefersEverythingOnAClosedLimiter(t *testing.T) {
	tr := newShadowTree(t)
	writeAlbum(t, tr.observe, "ada", "01-anvil", albumTracks(obsRel01, "Anvil", "One"))
	writeAlbum(t, tr.observe, "ada", "02-basalt", albumTracks(obsRel04, "Basalt", "Two"))
	lib := t.TempDir()
	lib1 := writeLibraryFile(t, lib, "ada-01.flac", obsRel01, "One")
	lib2 := writeLibraryFile(t, lib, "ada-02.flac", obsRel04, "Two")
	writeRunReport(t, tr.oracle, oracleRunReport{
		RunID: "run1", StartedAt: passNow.Add(-time.Hour), EndedAt: passNow.Add(-time.Hour).Add(time.Minute),
		Albums: []oracleAlbumOutcome{
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib1}},
			{Uploader: "ada", Outcome: outcomeImported, LibraryPaths: []string{lib2}},
		},
	})
	srv := observeServer(t)

	closed := musicbrainz.NewLimiter(time.Millisecond)
	if err := closed.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	cfg := tr.config(srv.URL)
	cfg.Limiter = closed

	res, err := Run(context.Background(), cfg, passNow)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Observations) != 2 {
		t.Fatalf("observed %d albums, want 2", len(res.Observations))
	}
	for _, obs := range res.Observations {
		if obs.Deferred != DeferLimiterClosed {
			t.Errorf("observation %s deferred %q, want %q", obs.ID, obs.Deferred, DeferLimiterClosed)
		}
	}
	if res.Pass.Deferred[DeferLimiterClosed] != 2 {
		t.Errorf("Deferred[%s] = %d, want 2", DeferLimiterClosed, res.Pass.Deferred[DeferLimiterClosed])
	}
	if res.Pass.Agreements != 0 || len(res.Pass.Divergences) != 0 {
		t.Errorf("Agreements = %d, Divergences = %v, want none: a closed limiter records no verdict at all", res.Pass.Agreements, res.Pass.Divergences)
	}
}

// TestObservePassIsIdempotent runs the exit-gate fixture twice under
// one frozen clock and proves contract S7 end to end: the second pass
// captures, matches and spends nothing, and every stored record is
// byte-identical to the first pass's own.
func TestObservePassIsIdempotent(t *testing.T) {
	tr := newShadowTree(t)
	library := t.TempDir()
	store := tr.openStore(t)
	seedReportGap(t, store)
	buildInbox(t, tr, library)
	srv := observeServer(t)

	cfg := tr.config(srv.URL)
	cfg.BudgetLimit = 12

	first, err := Run(context.Background(), cfg, passNow)
	if err != nil {
		t.Fatalf("first Run: %v", err)
	}
	firstStored, err := store.Observations()
	if err != nil {
		t.Fatalf("first Observations: %v", err)
	}
	firstFiles := make(map[string]string, len(firstStored))
	for _, obs := range firstStored {
		firstFiles[obs.ID] = readStoreFile(t, tr, filepath.Join("observations", obs.ID, "observation.json"))
	}

	second, err := Run(context.Background(), cfg, passNow)
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}

	if second.Pass.Captured != 0 {
		t.Errorf("second pass Captured = %d, want 0", second.Pass.Captured)
	}
	if len(second.Observations) != 0 {
		t.Errorf("second pass matched %d albums, want 0", len(second.Observations))
	}
	if second.Pass.Budget.Spent != 0 {
		t.Errorf("second pass Budget.Spent = %d, want 0", second.Pass.Budget.Spent)
	}
	if second.ClockStarted {
		t.Error("second pass restarted the clock")
	}

	secondStored, err := store.Observations()
	if err != nil {
		t.Fatalf("second Observations: %v", err)
	}
	if len(secondStored) != len(firstStored) {
		t.Fatalf("second pass stored %d observations, want %d (unchanged)", len(secondStored), len(firstStored))
	}
	for _, obs := range secondStored {
		want, ok := firstFiles[obs.ID]
		if !ok {
			t.Errorf("second pass stored a new observation %s", obs.ID)
			continue
		}
		got := readStoreFile(t, tr, filepath.Join("observations", obs.ID, "observation.json"))
		if got != want {
			t.Errorf("observation %s changed between passes:\nfirst:  %s\nsecond: %s", obs.ID, want, got)
		}
	}

	if first.Pass.ID != second.Pass.ID {
		t.Errorf("pass ids %q and %q differ under one frozen clock", first.Pass.ID, second.Pass.ID)
	}
}
