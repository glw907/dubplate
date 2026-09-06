package shadow

import (
	"context"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

const (
	closeID = "33333333-3333-3333-3333-333333333333"
	farID   = "44444444-4444-4444-4444-444444444444"
)

// recordingOfTwoCandidates drives one real matchAlbum whose flow has no
// id consensus, so mode one issues a search and then one lookup per
// returned id, and returns that run's own recording. The re-score is
// then handed the two LOOKUPS ALONE: the search is deliberately left
// out, so a re-score that drove a second tag.Album would take a miss on
// a request the recording it was given never carried (mutation 4).
func recordingOfTwoCandidates(t *testing.T) (lookups []RecordedRequest, bodies [][]byte, items []matcher.ItemFields) {
	t.Helper()

	items = []matcher.ItemFields{
		{Artist: "Cabinet", AlbumArtist: "Cabinet", Album: "Windrose", Title: "One", Track: 1},
		{Artist: "Cabinet", AlbumArtist: "Cabinet", Album: "Windrose", Title: "Two", Track: 2},
	}
	srv := releaseServer(t, map[string][]byte{
		"search": searchBody(closeID, farID),
		closeID:  releaseBody(closeID, "Windrose", "Cabinet", []string{"One", "Two"}),
		farID:    releaseBody(farID, "Cardinal Points", "Sextant", []string{"Nine", "Ten"}),
	}, nil)

	_, requests, recorded, err := matchAlbum(context.Background(),
		Config{BaseURL: srv.URL, UserAgent: testUserAgent}, musicbrainz.Unmetered(), items, passNow)
	if err != nil {
		t.Fatalf("matchAlbum: %v", err)
	}
	if len(requests) != 3 {
		t.Fatalf("recorded %d requests, want 3 (one search and one lookup per returned id)", len(requests))
	}
	return requests[1:], recorded[1:], items
}

func TestRescoreDefaultIssuesNoSearchAndSpendsNoToken(t *testing.T) {
	lookups, bodies, items := recordingOfTwoCandidates(t)

	verdict, err := rescoreDefault(context.Background(), lookups, bodies, []string{closeID, farID}, items, passNow)
	if err != nil {
		t.Fatalf("rescoreDefault: %v (a search the recording never captured is the failure mutation 4 produces)", err)
	}
	// Both recorded candidates were looked up and scored: the near one
	// wins, which a re-score that looked up only the first id or scored
	// neither could not report.
	if verdict.AlbumID != closeID {
		t.Errorf("verdict.AlbumID = %q, want %q (the closer of the two recorded candidates)", verdict.AlbumID, closeID)
	}
	if verdict.Recommendation != matcher.RecStrong.String() {
		t.Errorf("verdict.Recommendation = %q, want %q", verdict.Recommendation, matcher.RecStrong)
	}
	if verdict.Action != string(matcher.ActionApply) {
		t.Errorf("verdict.Action = %q, want apply", verdict.Action)
	}
	// The token half of this claim is asserted at Run level, where a
	// pass budget exists to read: TestRunSpendsNoTokenOnTheRescore.
}

func TestRescoreDefaultMarksItsVerdictAsAHypothesis(t *testing.T) {
	lookups, bodies, items := recordingOfTwoCandidates(t)

	verdict, err := rescoreDefault(context.Background(), lookups, bodies, []string{closeID, farID}, items, passNow)
	if err != nil {
		t.Fatalf("rescoreDefault: %v", err)
	}
	if !verdict.Hypothesis {
		t.Error("verdict.Hypothesis is false; a re-scored DefaultCompat verdict is not the product's true one wherever the real flow would have widened the candidate set")
	}
}

func TestRescoreDefaultScoresUnderDefaultCompat(t *testing.T) {
	// The re-score's whole purpose is the OTHER compat selection, and
	// nothing in the verdict it returns says which table it used. The
	// observable is the tunables it builds: a re-score left on
	// FullCompat would hash equal to mode one's own pair.
	got := configHash(rescoreScoring())
	want := configHash(matcher.DefaultCompat().Apply(matcher.QuietImportTunables()), matcher.DefaultCompat())
	if got != want {
		t.Errorf("configHash(rescoreScoring()) = %s, want %s", got, want)
	}
	if got == configHash(shadowScoring()) {
		t.Error("the re-score scores under mode one's own tunables and compat; the second verdict would then be the first")
	}
}

func TestRescoreDefaultReportsAMissRatherThanDroppingACandidate(t *testing.T) {
	lookups, bodies, items := recordingOfTwoCandidates(t)

	// The recording covers the first candidate alone. A re-score that
	// silently dropped the second would return a verdict computed over
	// a candidate set the run never had, and reevaluate would retire a
	// divergence that never actually re-ran.
	verdict, err := rescoreDefault(context.Background(), lookups[:1], bodies[:1], []string{closeID, farID}, items, passNow)
	if err == nil {
		t.Fatalf("rescoreDefault returned verdict %+v and no error, want a refusal naming the uncovered candidate", verdict)
	}
	if verdict != nil {
		t.Errorf("verdict = %+v, want nil beside the error", verdict)
	}
	if !strings.Contains(err.Error(), farID) {
		t.Errorf("error = %q, want it to name the uncovered candidate %s", err, farID)
	}
}
