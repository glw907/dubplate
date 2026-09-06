package shadow

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	oracleReleaseID = "55555555-5555-5555-5555-555555555555"
	goReleaseID     = "66666666-6666-6666-6666-666666666666"
	testCommit      = "cafebabe"
)

// A shadowTree is one pass's four directories: the greenfield state
// directory the store roots in, and the incumbent's inbox, state
// directory and review tree, which Run only ever reads. library holds
// the files an imported run report names.
type shadowTree struct {
	state   string
	observe string
	oracle  string
	review  string
	library string
}

func newShadowTree(t *testing.T) shadowTree {
	t.Helper()

	root := t.TempDir()
	tr := shadowTree{
		state:   filepath.Join(root, "state"),
		observe: filepath.Join(root, "inbox"),
		oracle:  filepath.Join(root, "incumbent"),
		review:  filepath.Join(root, "review"),
		library: filepath.Join(root, "library"),
	}
	for _, dir := range []string{tr.state, tr.observe, tr.oracle, tr.review, tr.library} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	return tr
}

// config is the pass configuration every run test shares. Interval is a
// millisecond because the limiter paces every request and the estate's
// own two-second pace would buy nothing here but wall time.
func (tr shadowTree) config(baseURL string) Config {
	return Config{
		StateDir:       tr.state,
		ObserveDir:     tr.observe,
		OracleStateDir: tr.oracle,
		ReviewDir:      tr.review,
		UserAgent:      testUserAgent,
		MatcherCommit:  testCommit,
		BudgetLimit:    60,
		Interval:       time.Millisecond,
		BaseURL:        baseURL,
	}
}

func (tr shadowTree) openStore(t *testing.T) *Store {
	t.Helper()

	s, err := OpenStore(filepath.Join(tr.state, "shadow"))
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	return s
}

// albumTracks is one album's ordered Vorbis tag lists. Every track
// carries the same release id, which is the item consensus tag.Album's
// early return reads (tag.go:96-118), so a fixture album costs exactly
// one recorded request.
func albumTracks(albumID, album string, titles ...string) [][]string {
	tracks := make([][]string, len(titles))
	for i, title := range titles {
		tracks[i] = []string{
			"ALBUMARTIST=Cabinet",
			"ARTIST=Cabinet",
			"ALBUM=" + album,
			"TITLE=" + title,
			fmt.Sprintf("TRACKNUMBER=%d", i+1),
			"MUSICBRAINZ_ALBUMID=" + albumID,
		}
	}
	return tracks
}

// writeLibraryFile writes the one file an imported run report names: the
// shape beets leaves behind under write: yes, carrying the release it
// applied and the title and length the content join matches on.
func writeLibraryFile(t *testing.T, dir, name, albumID, title string) string {
	t.Helper()

	path := filepath.Join(dir, name)
	data := buildFLAC(t, []string{"ALBUM=Windrose", "TITLE=" + title, "MUSICBRAINZ_ALBUMID=" + albumID}, 44100, 44100)
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write library file: %v", err)
	}
	if err := os.Chtimes(path, fixedModTime, fixedModTime); err != nil {
		t.Fatalf("chtimes library file: %v", err)
	}
	return path
}

// importedReport writes one run report carrying a single imported
// outcome, naming the library file beets left behind for uploader.
func importedReport(t *testing.T, tr shadowTree, runID, uploader, libraryPath string) {
	t.Helper()

	writeRunReport(t, tr.oracle, oracleRunReport{
		RunID:     runID,
		StartedAt: passNow.Add(-time.Hour),
		EndedAt:   passNow.Add(-time.Hour).Add(time.Minute),
		Albums: []oracleAlbumOutcome{{
			Uploader:     uploader,
			Outcome:      outcomeImported,
			LibraryPaths: []string{libraryPath},
		}},
	})
}

// requestLogServer answers release lookups from bodies and records the
// order it was asked in. That order is the only place a pass's own
// match order is observable from outside the result it returns.
func requestLogServer(t *testing.T, bodies map[string][]byte) (*httptest.Server, func() []string) {
	t.Helper()

	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/ws/2/release/")
		mu.Lock()
		seen = append(seen, id)
		mu.Unlock()
		body, ok := bodies[id]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) //nolint:errcheck // best-effort write to a test stub
	}))
	t.Cleanup(srv.Close)
	return srv, func() []string {
		mu.Lock()
		defer mu.Unlock()
		return slices.Clone(seen)
	}
}

// agreeingPass writes the one-album fixture both sides agree on: the
// inbox album carries the release id in its own tags, so mode one takes
// the id-consensus early return and spends exactly one token, and the
// library file the run report names carries the same release.
func agreeingPass(t *testing.T, tr shadowTree, libraryDir string) {
	t.Helper()

	writeAlbum(t, tr.observe, "ada", "windrose", albumTracks(oracleReleaseID, "Windrose", "One", "Two"))
	lib := writeLibraryFile(t, libraryDir, "ada-one.flac", oracleReleaseID, "One")
	importedReport(t, tr, "runada1", "ada", lib)
}

func windroseServer(t *testing.T) *httptest.Server {
	t.Helper()

	return releaseServer(t, map[string][]byte{
		oracleReleaseID: releaseBody(oracleReleaseID, "Windrose", "Cabinet", []string{"One", "Two"}),
		goReleaseID:     releaseBody(goReleaseID, "Windrose", "Cabinet", []string{"One", "Two"}),
	}, nil)
}

func TestRunSpendsNoTokenOnTheRescore(t *testing.T) {
	tr := newShadowTree(t)
	agreeingPass(t, tr, tr.library)
	srv := windroseServer(t)

	res, err := Run(context.Background(), tr.config(srv.URL), passNow)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Observations) != 1 {
		t.Fatalf("observed %d albums, want 1", len(res.Observations))
	}
	obs := res.Observations[0]
	if obs.Class != ClassAgreement {
		t.Fatalf("class = %q (deferred %q), want %q", obs.Class, obs.Deferred, ClassAgreement)
	}
	// Without a second verdict there is no replay at all, and the
	// budget assertion below would assert nothing.
	if obs.Go == nil || obs.Go.Default == nil {
		t.Fatalf("Go = %+v, want a DefaultCompat re-score beside the recorded verdict", obs.Go)
	}
	if len(obs.Requests) != 1 {
		t.Fatalf("recorded %d requests, want 1 (the id-consensus early return)", len(obs.Requests))
	}
	if res.Pass.Budget.Spent != len(obs.Requests) {
		t.Errorf("Pass.Budget.Spent = %d, want %d: the re-score replays through musicbrainz.Unmetered, so it adds nothing to the figure the report prints as honest",
			res.Pass.Budget.Spent, len(obs.Requests))
	}
}

func TestRunIsDeterministicUnderAFrozenClock(t *testing.T) {
	// One server and one library directory for both trees, so the
	// recorded bodies are identical and the path the oracle signal
	// records the release from is too, leaving the pass itself as the
	// only variable. Every fixture writer pins its modification times to
	// fixedModTime, without which the capture digest, and therefore
	// every id below, would be a function of when the test ran
	// (Resolution R5, fold B3).
	srv := windroseServer(t)
	library := t.TempDir()

	runOnce := func() (shadowTree, *PassResult) {
		tr := newShadowTree(t)
		agreeingPass(t, tr, library)
		res, err := Run(context.Background(), tr.config(srv.URL), passNow)
		if err != nil {
			t.Fatalf("Run: %v", err)
		}
		return tr, res
	}

	firstTree, first := runOnce()
	secondTree, second := runOnce()

	if first.Pass.ID != second.Pass.ID {
		t.Errorf("pass ids %q and %q differ under one frozen clock", first.Pass.ID, second.Pass.ID)
	}
	if len(first.Observations) != 1 || len(second.Observations) != 1 {
		t.Fatalf("observed %d and %d albums, want 1 each", len(first.Observations), len(second.Observations))
	}
	if first.Observations[0].ID != second.Observations[0].ID {
		t.Errorf("observation ids %q and %q differ over identical fixtures", first.Observations[0].ID, second.Observations[0].ID)
	}

	obsID := first.Observations[0].ID
	for _, rel := range []string{
		filepath.Join("observations", obsID, "observation.json"),
		filepath.Join("passes", first.Pass.ID+".json"),
	} {
		a := readStoreFile(t, firstTree, rel)
		b := readStoreFile(t, secondTree, rel)
		if a != b {
			t.Errorf("%s differs between two passes over identical fixtures:\n%s\n%s", rel, a, b)
		}
	}
}

func readStoreFile(t *testing.T, tr shadowTree, rel string) string {
	t.Helper()

	data, err := os.ReadFile(filepath.Join(tr.state, "shadow", rel))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(data)
}

func TestRunMatchesInCaptureOrder(t *testing.T) {
	tr := newShadowTree(t)
	albums := []struct{ basename, albumID string }{
		{"01-anvil", "aaaaaaaa-0000-0000-0000-000000000001"},
		{"02-basalt", "aaaaaaaa-0000-0000-0000-000000000002"},
		{"03-cobalt", "aaaaaaaa-0000-0000-0000-000000000003"},
	}
	bodies := make(map[string][]byte, len(albums))
	var skipped []string
	for _, a := range albums {
		writeAlbum(t, tr.observe, "ada", a.basename, albumTracks(a.albumID, a.basename, "One", "Two"))
		bodies[a.albumID] = releaseBody(a.albumID, a.basename, "Cabinet", []string{"One", "Two"})
		skipped = append(skipped, "/var/lib/staging/ada/"+a.basename)
	}
	writeTagLog(t, tr.oracle, skipped)
	srv, asked := requestLogServer(t, bodies)

	res, err := Run(context.Background(), tr.config(srv.URL), passNow)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Observations) != len(albums) {
		t.Fatalf("observed %d albums, want %d", len(res.Observations), len(albums))
	}

	var gotBasenames, gotIDs []string
	for _, obs := range res.Observations {
		gotBasenames = append(gotBasenames, obs.Capture.Basename)
		gotIDs = append(gotIDs, obs.ID)
	}
	wantBasenames := []string{albums[0].basename, albums[1].basename, albums[2].basename}
	if !slices.Equal(gotBasenames, wantBasenames) {
		t.Errorf("match order = %v, want %v ((uploader, basename) capture order)", gotBasenames, wantBasenames)
	}
	// The fixture is only discriminating while its digest-derived ids
	// sort differently from its basenames: an id is a SHA-256 prefix and
	// says nothing about capture order (Resolution R18). Rename an album
	// if this ever fires; the ids are a function of the fixture alone.
	if slices.IsSorted(gotIDs) {
		t.Fatalf("observation ids %v are already in capture order, so this fixture cannot tell the two orders apart", gotIDs)
	}
	wantAsked := []string{albums[0].albumID, albums[1].albumID, albums[2].albumID}
	if !slices.Equal(asked(), wantAsked) {
		t.Errorf("the wire saw %v, want %v: the budget is spent in capture order", asked(), wantAsked)
	}
}

func TestRunStartsTheClockOnAPassThatObservesNothing(t *testing.T) {
	tr := newShadowTree(t)

	// An empty inbox, no run report and no review tree entry: the clock
	// starts at the deploy rather than at the first upload, because the
	// floor measures the observation window and not the arrival rate
	// (Resolution R11).
	res, err := Run(context.Background(), tr.config("http://unreachable.invalid"), passNow)
	if err != nil {
		t.Fatalf("Run over an empty inbox: %v", err)
	}
	if !res.ClockStarted {
		t.Error("ClockStarted is false; a pass that observes nothing is still a successful pass")
	}
	if res.Pass.Observed != 0 || len(res.Observations) != 0 {
		t.Errorf("Observed = %d over %d observations, want 0", res.Pass.Observed, len(res.Observations))
	}

	start, ok, err := tr.openStore(t).ClockStart()
	if err != nil || !ok {
		t.Fatalf("ClockStart = %v, %v, %v; want the pass clock", start, ok, err)
	}
	if !start.Equal(passNow) {
		t.Errorf("started-at holds %s, want the pass clock %s", start, passNow)
	}

	later, err := Run(context.Background(), tr.config("http://unreachable.invalid"), passNow.Add(time.Hour))
	if err != nil {
		t.Fatalf("second Run: %v", err)
	}
	if later.ClockStarted {
		t.Error("the second pass restarted the clock; launch date is the clock start plus the floor")
	}
}

func TestRunStoresExactlyOneOutcomePerObservation(t *testing.T) {
	tr := newShadowTree(t)

	// ada agrees, bo's import named no library file, cy was rejected at
	// the gate before beets ever saw it, and di is still waiting for the
	// incumbent's next run.
	agreeingPass(t, tr, tr.library)
	writeAlbum(t, tr.observe, "bo", "no-release", albumTracks("bbbbbbbb-0000-0000-0000-000000000001", "Ligature", "One"))
	writeAlbum(t, tr.observe, "cy", "gate-rejected", albumTracks("cccccccc-0000-0000-0000-000000000001", "Meridian", "One"))
	writeAlbum(t, tr.observe, "di", "waiting", albumTracks("dddddddd-0000-0000-0000-000000000001", "Nocturne", "One"))
	writeReviewDir(t, tr.review, "2026-09-01", "cy", "gate-rejected")
	writeRunReport(t, tr.oracle, oracleRunReport{
		RunID:     "runmixed",
		StartedAt: passNow.Add(-time.Hour),
		EndedAt:   passNow.Add(-time.Hour).Add(time.Minute),
		Albums: []oracleAlbumOutcome{
			{Uploader: "bo", Outcome: outcomeImported},
			{Uploader: "cy", Outcome: outcomeQuarantined},
		},
	})
	srv := windroseServer(t)

	if _, err := Run(context.Background(), tr.config(srv.URL), passNow); err != nil {
		t.Fatalf("Run: %v", err)
	}

	stored, err := tr.openStore(t).Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	if len(stored) != 4 {
		t.Fatalf("stored %d observations, want 4", len(stored))
	}

	counts := map[string]int{}
	for _, obs := range stored {
		set := 0
		for name, value := range map[string]string{
			"class": obs.Class, "deferred": obs.Deferred,
			"unreconciled": obs.Unreconciled, "excluded": obs.Excluded,
		} {
			if value != "" {
				set++
				counts[name]++
			}
		}
		if set > 1 {
			t.Errorf("observation %s (%s) carries %d outcomes, want exactly one: %+v", obs.ID, obs.Album, set, obs)
		}
		if set == 0 && obs.Oracle != nil {
			t.Errorf("observation %s (%s) carries an oracle signal and no outcome", obs.ID, obs.Album)
		}
	}
	want := map[string]int{"class": 1, "unreconciled": 1, "excluded": 1}
	for name, n := range want {
		if counts[name] != n {
			t.Errorf("%d observations carry %s, want %d (counts: %v)", counts[name], name, n, counts)
		}
	}
}

func TestRunStagesOnePendingRowPerDivergence(t *testing.T) {
	tr := newShadowTree(t)

	// The inbox tags name one release and the library file the incumbent
	// wrote names another, which is go-applies-different-release.
	writeAlbum(t, tr.observe, "ada", "windrose", albumTracks(goReleaseID, "Windrose", "One", "Two"))
	lib := writeLibraryFile(t, tr.library, "ada-one.flac", oracleReleaseID, "One")
	importedReport(t, tr, "rundiverge", "ada", lib)
	srv := windroseServer(t)

	res, err := Run(context.Background(), tr.config(srv.URL), passNow)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(res.Observations) != 1 || res.Observations[0].Class != ClassGoAppliesDifferentRelease {
		t.Fatalf("observations = %+v, want one %s", res.Observations, ClassGoAppliesDifferentRelease)
	}
	if res.Pass.Divergences[ClassGoAppliesDifferentRelease] != 1 {
		t.Errorf("Divergences = %v, want one %s", res.Pass.Divergences, ClassGoAppliesDifferentRelease)
	}
	if len(res.Pending) != 1 {
		t.Fatalf("staged %d pending rows, want 1", len(res.Pending))
	}
	row := res.Pending[0]
	if row.ID != res.Observations[0].ID || row.Album != "ada/windrose" {
		t.Errorf("row identifies %q / %q, want the observation's own id and album", row.ID, row.Album)
	}
	if row.ReleaseMBID != oracleReleaseID {
		t.Errorf("row.ReleaseMBID = %q, want the oracle's release %q: the release under dispute is the one the applying side named (ruling 2)", row.ReleaseMBID, oracleReleaseID)
	}
	if row.MatcherCommit != testCommit || row.MatcherConfigHash == "" {
		t.Errorf("row stamps = %q / %q, want the pass's own matcher commit and config hash", row.MatcherCommit, row.MatcherConfigHash)
	}
	if row.Date != "2026-09-01" {
		t.Errorf("row.Date = %q, want the pass clock's own date", row.Date)
	}

	if len(res.Fixtures) != 1 || res.Fixtures[0].ObservationID != row.ID {
		t.Fatalf("fixtures = %+v, want one for the matched observation", res.Fixtures)
	}
	fixture := res.Fixtures[0]
	if fixture.Profile != "estate" || fixture.EvalDate != "2026-09-01" {
		t.Errorf("fixture profile/eval_date = %q/%q, want estate/2026-09-01", fixture.Profile, fixture.EvalDate)
	}
	if len(fixture.Items) != 2 || fixture.Items[0]["title"] != "One" {
		t.Errorf("fixture items = %+v, want the captured item bag", fixture.Items)
	}
	if len(fixture.Requests) != 1 || fixture.Requests[0].Response != "responses/001.json" {
		t.Errorf("fixture requests = %+v, want the recorded request naming its stored body", fixture.Requests)
	}
	if _, err := tr.openStore(t).Response(row.ID, fixture.Requests[0].Response); err != nil {
		t.Errorf("stored response: %v", err)
	}
}

func TestPendingRowNamesTheReleaseUnderDispute(t *testing.T) {
	base := func(class string, oracle OracleSignal, verdicts *GoVerdicts) *Observation {
		return &Observation{
			ID: "shadow-0123456789abcdef", Album: "ada/windrose",
			MatcherCommit: testCommit, MatcherConfigHash: "0f0f",
			Oracle: &oracle, Go: verdicts, Class: class,
		}
	}
	applied := GoVerdicts{Compat: &GoVerdict{Action: "apply", AlbumID: goReleaseID}}
	skipped := GoVerdicts{Compat: &GoVerdict{Action: "skip"}}

	tests := []struct {
		name    string
		obs     *Observation
		wantRow bool
		wantMB  string
	}{
		{"an agreement stages nothing", base(ClassAgreement, OracleSignal{Signal: "skip"}, &skipped), false, ""},
		{"a deferral stages nothing", &Observation{ID: "shadow-1", Deferred: DeferBudgetExhausted}, false, ""},
		{"go applies where the oracle skipped: the Go release is under dispute",
			base(ClassGoAppliesOracleSkipped, OracleSignal{Signal: "skip"}, &applied), true, goReleaseID},
		{"go skips where the oracle applied: the oracle's release is under dispute",
			base(ClassGoSkipsOracleApplied, OracleSignal{Signal: "apply", ReleaseMBID: oracleReleaseID}, &skipped), true, oracleReleaseID},
		{"a different release: the oracle's release is under dispute",
			base(ClassGoAppliesDifferentRelease, OracleSignal{Signal: "apply", ReleaseMBID: oracleReleaseID}, &applied), true, oracleReleaseID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row, ok := pendingRowFor(tt.obs, passNow)
			if ok != tt.wantRow {
				t.Fatalf("pendingRowFor staged = %v, want %v", ok, tt.wantRow)
			}
			if !ok {
				return
			}
			if row.ReleaseMBID != tt.wantMB {
				t.Errorf("row.ReleaseMBID = %q, want %q", row.ReleaseMBID, tt.wantMB)
			}
			// The four prose fields are rendered from the observation
			// rather than staged as TODO, because a row that says
			// nothing still consumes an owner adjudication.
			for name, value := range map[string]string{
				"suspected": row.Suspected, "citation": row.Citation,
				"release_truth": row.ReleaseTruth, "reason": row.Reason,
			} {
				if value == "" || strings.Contains(value, "TODO") {
					t.Errorf("%s = %q, want prose rendered from the observation", name, value)
				}
			}
			if !strings.Contains(row.Suspected, tt.obs.Class) {
				t.Errorf("row.Suspected = %q, want it to name the class", row.Suspected)
			}
			if !strings.Contains(row.Reason, testCommit) {
				t.Errorf("row.Reason = %q, want it to name the matcher commit", row.Reason)
			}
			if !strings.Contains(row.Citation, tt.obs.ID) {
				t.Errorf("row.Citation = %q, want it to name the observation", row.Citation)
			}
		})
	}
}

func TestPendingRowSurvivesTheCompatClassAnnotation(t *testing.T) {
	// A divergence the DefaultCompat re-score does not reproduce is a
	// hypothesised compat-class divergence. It still gets a row: the
	// annotation is a hypothesis rather than a ruling, and suppressing a
	// row on a hypothesis is the silent loss ruling 3's own limit
	// paragraph warns about.
	obs := &Observation{
		ID: "shadow-0123456789abcdef", Album: "ada/windrose",
		MatcherCommit: testCommit, MatcherConfigHash: "0f0f",
		Oracle: &OracleSignal{Signal: "skip"},
		Go: &GoVerdicts{
			Compat:  &GoVerdict{Action: "apply", AlbumID: goReleaseID},
			Default: &GoVerdict{Action: "skip", Hypothesis: true},
		},
		Class: ClassGoAppliesOracleSkipped,
	}

	row, ok := pendingRowFor(obs, passNow)
	if !ok {
		t.Fatal("no row staged: a compat-class annotation is a hypothesis, never a suppression")
	}
	if !strings.Contains(row.Suspected, "DefaultCompat") {
		t.Errorf("row.Suspected = %q, want the compat-class hypothesis named", row.Suspected)
	}
}
