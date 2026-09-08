package shadow

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// yearReleaseBody is a releaseBody that also carries a release date and
// a release-group first-release-date, so the album-year component's
// ratio branch (album_year.go:38-44) has a nonzero OriginalYear to
// divide by, which is what makes the candidate's distance move with
// evalDate's own year.
func yearReleaseBody(albumID string, releaseYear, originalYear int) []byte {
	artistCredit := []map[string]any{{
		"name":       "Cabinet",
		"joinphrase": "",
		"artist":     map[string]any{"id": "artist-" + albumID, "name": "Cabinet", "sort-name": "Cabinet"},
	}}
	body := map[string]any{
		"id":            albumID,
		"title":         "Windrose",
		"status":        "Official",
		"date":          fmt.Sprintf("%d-01-01", releaseYear),
		"artist-credit": artistCredit,
		"media": []map[string]any{{
			"position": 1, "title": "", "track-count": 1,
			"tracks": []map[string]any{{
				"id": "track-1", "position": 1, "number": "1", "title": "One",
				"artist-credit": []any{},
				"recording":     map[string]any{"id": "rec-1", "title": "One", "video": false, "artist-credit": artistCredit},
			}},
		}},
		"release-group": map[string]any{"id": "rg-" + albumID, "title": "Windrose", "first-release-date": fmt.Sprintf("%d-01-01", originalYear)},
		"relations":     []any{},
	}
	data, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Sprintf("shadow: marshaling test year release body: %v", err))
	}
	return data
}

// recordAgainst drives one real matchAlbum against a stub server and
// returns what it recorded, so a fixture's PathQuery strings are the
// canonical ones the real client builds rather than hand-typed guesses.
func recordAgainst(t *testing.T, items []matcher.ItemFields, bodies map[string][]byte, evalDate time.Time) (requests []RecordedRequest, responseBodies [][]byte) {
	t.Helper()
	srv := releaseServer(t, bodies, nil)
	_, requests, responseBodies, err := matchAlbum(context.Background(),
		Config{BaseURL: srv.URL, UserAgent: testUserAgent}, musicbrainz.Unmetered(), items, evalDate)
	if err != nil {
		t.Fatalf("recordAgainst: matchAlbum: %v", err)
	}
	return requests, responseBodies
}

// seedObservation stores requests/bodies as one observation's own
// recorded evidence, stamped with a fabricated original verdict that
// stands in for whatever an earlier matcher commit actually recorded:
// the point of every test in this file is that the CURRENT matcher
// code, replaying the SAME stored evidence, may compute something
// different from that earlier commit's own answer.
func seedObservation(t *testing.T, s *Store, id string, items []matcher.ItemFields,
	requests []RecordedRequest, bodies [][]byte, evalDate time.Time,
	class string, goVerdicts *GoVerdicts, oracle OracleSignal) *Observation {
	t.Helper()
	obs := &Observation{
		ID:            id,
		Album:         "uploader/" + id,
		MatcherCommit: "old-commit", // every fixture in this file re-evaluates a superseded stamp
		EvalDate:      evalDate,
		Capture:       Capture{Uploader: "uploader", Basename: id, Digest: id, Items: items},
		Oracle:        &oracle,
		Go:            goVerdicts,
		Class:         class,
		Requests:      slices.Clone(requests),
	}
	for i, body := range bodies {
		if err := s.PutResponse(id, i+1, body); err != nil {
			t.Fatalf("PutResponse: %v", err)
		}
		obs.Requests[i].Response = responsePath(i + 1)
	}
	if err := s.PutObservation(obs); err != nil {
		t.Fatalf("PutObservation: %v", err)
	}
	return obs
}

// stageFakePending writes a minimal pending row, standing in for the
// AdjudicationEntry cmd/dubplate-server would actually stage: this
// package tests only that Reevaluate removes or leaves the staged row,
// never its shape.
func stageFakePending(t *testing.T, s *Store, id string) {
	t.Helper()
	if err := s.PutPending(id, []byte(`{"id":"`+id+`"}`)); err != nil {
		t.Fatalf("PutPending: %v", err)
	}
}

func stagedIDs(t *testing.T, s *Store) []string {
	t.Helper()
	ids, err := s.PendingIDs()
	if err != nil {
		t.Fatalf("PendingIDs: %v", err)
	}
	return ids
}

const (
	windroseID  = "66666666-6666-6666-6666-666666666666"
	mismatchID  = "77777777-7777-7777-7777-777777777777"
	unfetchedID = "88888888-8888-8888-8888-888888888888"
)

// TestReevaluateRetiresADivergenceThatNoLongerReproduces is also the
// eval-date fixture: obs.EvalDate (2026) scores the year-sensitive
// candidate at RecMedium (skip, agreeing with the oracle's skip), while
// the "now" this test hands Reevaluate (2200) would score the SAME
// candidate RecStrong (apply, still diverging) were it used instead
// (mutation 3). The real code retires either way, because it reads only
// obs.EvalDate.
func TestReevaluateRetiresADivergenceThatNoLongerReproduces(t *testing.T) {
	s := openTempStore(t)
	items := []matcher.ItemFields{{Artist: "Cabinet", AlbumArtist: "Cabinet", Album: "Windrose", Title: "One", Track: 1, Year: 1995}}
	evalDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	body := yearReleaseBody(windroseID, 2020, 1980)
	requests, bodies := recordAgainst(t, items, map[string][]byte{"search": searchBody(windroseID), windroseID: body}, evalDate)

	obs := seedObservation(t, s, "obs-retire", items, requests, bodies, evalDate,
		ClassGoAppliesOracleSkipped,
		&GoVerdicts{Compat: &GoVerdict{Recommendation: matcher.RecStrong.String(), Action: string(matcher.ActionApply), AlbumID: windroseID}},
		OracleSignal{Signal: "skip", Uploader: "uploader", AlbumBasename: "obs-retire"})
	stageFakePending(t, s, obs.ID)

	// A distant "now": a real observable divergence between reading
	// obs.EvalDate (correct) and this value (wrong) is exactly what
	// proves mutation 3.
	res, err := Reevaluate(context.Background(), Config{MatcherCommit: "new-commit"}, s, time.Date(2200, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("Reevaluate: %v", err)
	}
	if !slices.Contains(res.Retired, obs.ID) {
		t.Fatalf("Retired = %v, want it to contain %s", res.Retired, obs.ID)
	}

	resolvedIDs, err := s.ResolvedIDs()
	if err != nil {
		t.Fatalf("ResolvedIDs: %v", err)
	}
	if !slices.Contains(resolvedIDs, obs.ID) {
		t.Fatalf("ResolvedIDs = %v, want it to contain %s", resolvedIDs, obs.ID)
	}

	record, err := s.Resolved(obs.ID)
	if err != nil {
		t.Fatalf("Resolved: %v", err)
	}
	if record.OriginalCommit != "old-commit" {
		t.Errorf("OriginalCommit = %q, want %q", record.OriginalCommit, "old-commit")
	}
	if record.RetiringCommit != "new-commit" {
		t.Errorf("RetiringCommit = %q, want %q", record.RetiringCommit, "new-commit")
	}
	if record.OriginalClass != ClassGoAppliesOracleSkipped {
		t.Errorf("OriginalClass = %q, want %q", record.OriginalClass, ClassGoAppliesOracleSkipped)
	}
	if record.NewClass != ClassAgreement {
		t.Errorf("NewClass = %q, want %q", record.NewClass, ClassAgreement)
	}
	if record.OriginalGo == nil || record.OriginalGo.Compat == nil ||
		record.OriginalGo.Compat.Recommendation != matcher.RecStrong.String() ||
		record.OriginalGo.Compat.Action != string(matcher.ActionApply) {
		t.Errorf("OriginalGo = %+v, want the seeded RecStrong/apply verdict", record.OriginalGo)
	}
	if record.NewGo == nil || record.NewGo.Compat == nil ||
		record.NewGo.Compat.Recommendation == matcher.RecStrong.String() ||
		record.NewGo.Compat.Action == string(matcher.ActionApply) {
		t.Errorf("NewGo = %+v, want the replay's own (weaker, skip) verdict, distinct from the seeded original", record.NewGo)
	}
	if record.Date != "2200-01-01" {
		t.Errorf("Date = %q, want %q (time.DateOnly of now)", record.Date, "2200-01-01")
	}

	if slices.Contains(stagedIDs(t, s), obs.ID) {
		t.Error("the pending row is still staged after a retire; it should have been removed")
	}

	stored, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	idx := slices.IndexFunc(stored, func(o *Observation) bool { return o.ID == obs.ID })
	if idx < 0 {
		t.Fatalf("observation %s vanished", obs.ID)
	}
	if stored[idx].MatcherCommit != "new-commit" {
		t.Errorf("MatcherCommit = %q, want it restamped to %q", stored[idx].MatcherCommit, "new-commit")
	}
	if stored[idx].Class != ClassAgreement {
		t.Errorf("Class = %q, want it reclassified to %q", stored[idx].Class, ClassAgreement)
	}
	if stored[idx].Go == nil || stored[idx].Go.Compat == nil ||
		stored[idx].Go.Compat.Recommendation == matcher.RecStrong.String() ||
		stored[idx].Go.Compat.Action == string(matcher.ActionApply) {
		t.Errorf("Go = %+v, want it restamped to the replay's own (weaker, skip) verdict", stored[idx].Go)
	}
}

// TestReevaluateOfARetiredObservationIsIdempotent asserts that a second
// Reevaluate call at the same commit skips a retired observation (S7's
// idempotence shape) and leaves every stored record byte-identical: the
// restamp in the first call must be the observation's own, not a
// transient in-memory value the second call would redo.
func TestReevaluateOfARetiredObservationIsIdempotent(t *testing.T) {
	s := openTempStore(t)
	items := []matcher.ItemFields{{Artist: "Cabinet", AlbumArtist: "Cabinet", Album: "Windrose", Title: "One", Track: 1, Year: 1995}}
	evalDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	body := yearReleaseBody(windroseID, 2020, 1980)
	requests, bodies := recordAgainst(t, items, map[string][]byte{"search": searchBody(windroseID), windroseID: body}, evalDate)

	obs := seedObservation(t, s, "obs-retire-idempotent", items, requests, bodies, evalDate,
		ClassGoAppliesOracleSkipped,
		&GoVerdicts{Compat: &GoVerdict{Recommendation: matcher.RecStrong.String(), Action: string(matcher.ActionApply), AlbumID: windroseID}},
		OracleSignal{Signal: "skip", Uploader: "uploader", AlbumBasename: "obs-retire-idempotent"})
	stageFakePending(t, s, obs.ID)

	if _, err := Reevaluate(context.Background(), Config{MatcherCommit: "new-commit"}, s, passNow); err != nil {
		t.Fatalf("first Reevaluate: %v", err)
	}

	obsBefore, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	obsBeforeData, err := json.Marshal(obsBefore)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resolvedBefore, err := s.Resolved(obs.ID)
	if err != nil {
		t.Fatalf("Resolved: %v", err)
	}
	resolvedBeforeData, err := json.Marshal(resolvedBefore)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	res, err := Reevaluate(context.Background(), Config{MatcherCommit: "new-commit"}, s, passNow)
	if err != nil {
		t.Fatalf("second Reevaluate: %v", err)
	}
	if !slices.Contains(res.Skipped, obs.ID) {
		t.Fatalf("Skipped = %v, want it to contain the retired observation %s", res.Skipped, obs.ID)
	}

	obsAfter, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	obsAfterData, err := json.Marshal(obsAfter)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(obsBeforeData) != string(obsAfterData) {
		t.Errorf("observation moved on a second reevaluate:\nbefore: %s\nafter:  %s", obsBeforeData, obsAfterData)
	}

	resolvedAfter, err := s.Resolved(obs.ID)
	if err != nil {
		t.Fatalf("Resolved: %v", err)
	}
	resolvedAfterData, err := json.Marshal(resolvedAfter)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(resolvedBeforeData) != string(resolvedAfterData) {
		t.Errorf("resolved record moved on a second reevaluate:\nbefore: %s\nafter:  %s", resolvedBeforeData, resolvedAfterData)
	}

	if slices.Contains(stagedIDs(t, s), obs.ID) {
		t.Error("the pending row is staged again after a second reevaluate of a retired observation")
	}
}

// TestReevaluateRestampsADivergenceThatStillReproduces asserts a
// divergence that still reproduces is restamped with the current commit
// and stays staged.
func TestReevaluateRestampsADivergenceThatStillReproduces(t *testing.T) {
	s := openTempStore(t)
	items := []matcher.ItemFields{{Artist: "Cabinet", Album: "Windrose", Title: "Track One", MBAlbumID: windroseID}}
	body := releaseBody(windroseID, "Windrose", "Cabinet", []string{"Track One"})
	requests, bodies := recordAgainst(t, items, map[string][]byte{windroseID: body}, passNow)

	obs := seedObservation(t, s, "obs-restamp", items, requests, bodies, passNow,
		ClassGoAppliesOracleSkipped,
		&GoVerdicts{Compat: &GoVerdict{Recommendation: matcher.RecStrong.String(), Action: string(matcher.ActionApply), AlbumID: windroseID}},
		OracleSignal{Signal: "skip", Uploader: "uploader", AlbumBasename: "obs-restamp"})
	stageFakePending(t, s, obs.ID)

	res, err := Reevaluate(context.Background(), Config{MatcherCommit: "new-commit"}, s, passNow)
	if err != nil {
		t.Fatalf("Reevaluate: %v", err)
	}
	if !slices.Contains(res.Restamped, obs.ID) {
		t.Fatalf("Restamped = %v, want it to contain %s", res.Restamped, obs.ID)
	}

	stored, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	idx := slices.IndexFunc(stored, func(o *Observation) bool { return o.ID == obs.ID })
	if idx < 0 {
		t.Fatalf("observation %s vanished", obs.ID)
	}
	if stored[idx].MatcherCommit != "new-commit" {
		t.Errorf("MatcherCommit = %q, want %q", stored[idx].MatcherCommit, "new-commit")
	}
	if !slices.Contains(stagedIDs(t, s), obs.ID) {
		t.Error("the pending row was removed by a restamp; it should still be staged")
	}
}

// TestReevaluateRecordsANewDivergenceOnAnObservationThatAgreed asserts
// that an observation stamped as an agreement, whose replay now
// diverges, is reported as news rather than left silent. A retire-only
// design loses exactly this (mutation 2).
func TestReevaluateRecordsANewDivergenceOnAnObservationThatAgreed(t *testing.T) {
	s := openTempStore(t)
	items := []matcher.ItemFields{{Artist: "Cabinet", Album: "Windrose", Title: "Track One", MBAlbumID: windroseID}}
	body := releaseBody(windroseID, "Windrose", "Cabinet", []string{"Track One"})
	requests, bodies := recordAgainst(t, items, map[string][]byte{windroseID: body}, passNow)

	// The original commit's own (fabricated) story: it skipped, and the
	// oracle skipped too, so the pair agreed. No pending row was ever
	// staged for an agreement.
	obs := seedObservation(t, s, "obs-newlydiverged", items, requests, bodies, passNow,
		ClassAgreement,
		&GoVerdicts{Compat: &GoVerdict{Recommendation: matcher.RecNone.String(), Action: string(matcher.ActionSkip)}},
		OracleSignal{Signal: "skip", Uploader: "uploader", AlbumBasename: "obs-newlydiverged"})

	if slices.Contains(stagedIDs(t, s), obs.ID) {
		t.Fatal("no pending row should exist before an agreement is reevaluated")
	}

	res, err := Reevaluate(context.Background(), Config{MatcherCommit: "new-commit"}, s, passNow)
	if err != nil {
		t.Fatalf("Reevaluate: %v", err)
	}
	if !slices.Contains(res.NewlyDiverged, obs.ID) {
		t.Fatalf("NewlyDiverged = %v, want it to contain %s", res.NewlyDiverged, obs.ID)
	}

	stored, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	idx := slices.IndexFunc(stored, func(o *Observation) bool { return o.ID == obs.ID })
	if idx < 0 {
		t.Fatalf("observation %s vanished", obs.ID)
	}
	if !isDivergence(stored[idx].Class) {
		t.Errorf("Class = %q, want a divergence", stored[idx].Class)
	}

	// The wiring this package hands cmd/dubplate-server: PendingRowFor
	// builds a valid row off the now-updated observation, which is what
	// proves a pending row can be staged where none existed.
	row, ok := PendingRowFor(stored[idx], passNow)
	if !ok {
		t.Fatal("PendingRowFor reports no news for a newly diverged observation")
	}
	data, err := json.Marshal(row)
	if err != nil {
		t.Fatalf("marshal pending row: %v", err)
	}
	if err := s.PutPending(obs.ID, data); err != nil {
		t.Fatalf("PutPending: %v", err)
	}
	if !slices.Contains(stagedIDs(t, s), obs.ID) {
		t.Error("a pending row should now exist for the newly diverged observation")
	}
}

// TestReevaluateReportsUnevaluableRatherThanRetiring pins the fixture
// mutation 1 needs exactly: a go-applies-oracle-skipped observation
// with exactly two admitted candidates, stored responses covering only
// one lookup, and the remaining candidate's own data does not score an
// apply. Dropping the missing candidate silently (a 404 instead of a
// 501) would leave only the covered, badly-mismatched candidate,
// agreeing with the oracle's skip and wrongly retiring the row (fold,
// m16).
func TestReevaluateReportsUnevaluableRatherThanRetiring(t *testing.T) {
	s := openTempStore(t)
	items := []matcher.ItemFields{{Artist: "Cabinet", AlbumArtist: "Cabinet", Album: "Windrose", Title: "One", Track: 1}}
	mismatch := releaseBody(mismatchID, "Cardinal Points", "Sextant", []string{"Nine"})

	// The real recording: search returns two ids, the mismatched one is
	// covered, the other 404s at capture time (both are
	// admissible ways an id ends up uncovered). recordAgainst swallows
	// the 404 the same way tag.Album's own lookupAndAdd does, so the
	// real call succeeds; this test then keeps only the covered pair.
	requests, bodies := recordAgainst(t, items,
		map[string][]byte{"search": searchBody(mismatchID, unfetchedID), mismatchID: mismatch}, passNow)
	if len(requests) != 3 {
		t.Fatalf("recorded %d requests, want 3 (search, the covered lookup, the 404)", len(requests))
	}
	requests, bodies = requests[:2], bodies[:2] // drop the uncovered candidate's own request entirely

	obs := seedObservation(t, s, "obs-unevaluable", items, requests, bodies, passNow,
		ClassGoAppliesOracleSkipped,
		&GoVerdicts{Compat: &GoVerdict{Recommendation: matcher.RecStrong.String(), Action: string(matcher.ActionApply), AlbumID: unfetchedID}},
		OracleSignal{Signal: "skip", Uploader: "uploader", AlbumBasename: "obs-unevaluable"})
	stageFakePending(t, s, obs.ID)

	res, err := Reevaluate(context.Background(), Config{MatcherCommit: "new-commit"}, s, passNow)
	if err != nil {
		t.Fatalf("Reevaluate: %v", err)
	}
	if !slices.Contains(res.Unevaluable, obs.ID) {
		t.Fatalf("Unevaluable = %v, want it to contain %s", res.Unevaluable, obs.ID)
	}
	if slices.Contains(res.Retired, obs.ID) || slices.Contains(res.Restamped, obs.ID) || slices.Contains(res.NewlyDiverged, obs.ID) {
		t.Fatalf("result = %+v, want obs-unevaluable in Unevaluable alone", res)
	}

	stored, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	idx := slices.IndexFunc(stored, func(o *Observation) bool { return o.ID == obs.ID })
	if idx < 0 {
		t.Fatalf("observation %s vanished", obs.ID)
	}
	if stored[idx].MatcherCommit != "old-commit" {
		t.Errorf("MatcherCommit = %q, want it to survive at %q (mutation 5)", stored[idx].MatcherCommit, "old-commit")
	}
	if !slices.Contains(stagedIDs(t, s), obs.ID) {
		t.Error("the pending row was removed for an unevaluable observation; it should survive")
	}
	resolvedIDs, err := s.ResolvedIDs()
	if err != nil {
		t.Fatalf("ResolvedIDs: %v", err)
	}
	if slices.Contains(resolvedIDs, obs.ID) {
		t.Error("resolved/ was written for an unevaluable observation")
	}
}

// failingCountingTransport counts every call and fails it, so a test can
// assert Reevaluate never dials out.
type failingCountingTransport struct{ calls int }

func (c *failingCountingTransport) RoundTrip(*http.Request) (*http.Response, error) {
	c.calls++
	return nil, fmt.Errorf("failingCountingTransport: reevaluate must never reach the network")
}

// TestReevaluateSpendsNoBudgetAndReachesNoNetwork proves both halves of
// contract: cfg.Transport is never dialed (a network read is a test
// failure), and cfg.BudgetLimit is never honored (mutation 4): the
// fixture needs two real requests, so a metered client capped at one
// token would fail the second and report Unevaluable or an error
// instead of the expected Restamped.
func TestReevaluateSpendsNoBudgetAndReachesNoNetwork(t *testing.T) {
	s := openTempStore(t)
	items := []matcher.ItemFields{{Artist: "Cabinet", AlbumArtist: "Cabinet", Album: "Windrose", Title: "One", Track: 1}}
	body := releaseBody(windroseID, "Windrose", "Cabinet", []string{"One"})
	requests, bodies := recordAgainst(t, items, map[string][]byte{"search": searchBody(windroseID), windroseID: body}, passNow)
	if len(requests) != 2 {
		t.Fatalf("recorded %d requests, want 2 (search and one lookup)", len(requests))
	}

	obs := seedObservation(t, s, "obs-budget", items, requests, bodies, passNow,
		ClassGoAppliesOracleSkipped,
		&GoVerdicts{Compat: &GoVerdict{Recommendation: matcher.RecStrong.String(), Action: string(matcher.ActionApply), AlbumID: windroseID}},
		OracleSignal{Signal: "skip", Uploader: "uploader", AlbumBasename: "obs-budget"})

	transport := &failingCountingTransport{}
	res, err := Reevaluate(context.Background(), Config{
		MatcherCommit: "new-commit",
		Transport:     transport,
		BudgetLimit:   1,
	}, s, passNow)
	if err != nil {
		t.Fatalf("Reevaluate: %v", err)
	}
	if transport.calls != 0 {
		t.Errorf("failingCountingTransport.calls = %d, want 0: reevaluate reached the network", transport.calls)
	}
	if !slices.Contains(res.Restamped, obs.ID) {
		t.Fatalf("result = %+v, want %s in Restamped (a capped, metered budget would have starved the second request)", res, obs.ID)
	}
}

// TestReevaluateSkipsObservationsAtTheCurrentCommit asserts an
// observation already at the current commit is left byte-identical.
func TestReevaluateSkipsObservationsAtTheCurrentCommit(t *testing.T) {
	s := openTempStore(t)
	obs := &Observation{
		ID:            "obs-current",
		Album:         "uploader/obs-current",
		MatcherCommit: "current-commit",
		EvalDate:      passNow,
		Capture:       Capture{Uploader: "uploader", Basename: "obs-current", Digest: "obs-current"},
		Oracle:        &OracleSignal{Signal: "skip", Uploader: "uploader", AlbumBasename: "obs-current"},
		Class:         ClassAgreement,
	}
	if err := s.PutObservation(obs); err != nil {
		t.Fatalf("PutObservation: %v", err)
	}
	before, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	beforeData, err := json.Marshal(before)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	res, err := Reevaluate(context.Background(), Config{MatcherCommit: "current-commit"}, s, passNow)
	if err != nil {
		t.Fatalf("Reevaluate: %v", err)
	}
	if !slices.Contains(res.Skipped, obs.ID) {
		t.Fatalf("Skipped = %v, want it to contain %s", res.Skipped, obs.ID)
	}
	if len(res.Retired) != 0 || len(res.Restamped) != 0 || len(res.NewlyDiverged) != 0 || len(res.Unevaluable) != 0 {
		t.Fatalf("result = %+v, want everything else empty", res)
	}

	after, err := s.Observations()
	if err != nil {
		t.Fatalf("Observations: %v", err)
	}
	afterData, err := json.Marshal(after)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(beforeData) != string(afterData) {
		t.Errorf("observation moved:\nbefore: %s\nafter:  %s", beforeData, afterData)
	}
}
