package shadow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// releaseServer answers WS/2 requests from bodies, keyed "search" for a
// bare /ws/2/release and by release id otherwise, sending
// Retry-After: 0 beside any status override so a retryable row
// collapses its backoff instead of spending real wall time (fold M-h,
// tag_test.go:56-64's own precedent). status may be nil.
func releaseServer(t *testing.T, bodies map[string][]byte, status map[string]int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var key string
		if r.URL.Path == "/ws/2/release" {
			key = "search"
		} else {
			key = strings.TrimPrefix(r.URL.Path, "/ws/2/release/")
		}
		w.Header().Set("Retry-After", "0")
		if code, ok := status[key]; ok {
			w.WriteHeader(code)
			return
		}
		body, ok := bodies[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) //nolint:errcheck // best-effort write to a test stub
	}))
	t.Cleanup(srv.Close)
	return srv
}

// releaseBody is a minimal WS/2 release document, adapted from
// internal/engine/tag/testdata/early-return's committed body, which
// internal/engine/tag/tag_test.go:141-158 already proves scores
// matcher.RecStrong through the identical tag.Album path (fold M-f).
func releaseBody(albumID, albumTitle, artist string, tracks []string) []byte {
	artistCredit := []map[string]any{{
		"name":       artist,
		"joinphrase": "",
		"artist": map[string]any{
			"id":        "artist-" + albumID,
			"name":      artist,
			"sort-name": artist,
		},
	}}

	trackList := make([]map[string]any, len(tracks))
	for i, title := range tracks {
		n := i + 1
		trackList[i] = map[string]any{
			"id":            fmt.Sprintf("track-%s-%d", albumID, n),
			"position":      n,
			"number":        strconv.Itoa(n),
			"title":         title,
			"artist-credit": []any{},
			"recording": map[string]any{
				"id":            fmt.Sprintf("rec-%s-%d", albumID, n),
				"title":         title,
				"video":         false,
				"artist-credit": artistCredit,
			},
		}
	}

	body := map[string]any{
		"id":            albumID,
		"title":         albumTitle,
		"status":        "Official",
		"artist-credit": artistCredit,
		"media": []map[string]any{{
			"position":    1,
			"title":       "",
			"track-count": len(tracks),
			"tracks":      trackList,
		}},
		"release-group": map[string]any{"id": "rg-" + albumID, "title": albumTitle},
		"relations":     []any{},
	}

	data, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Sprintf("shadow: marshaling test release body: %v", err))
	}
	return data
}

// searchBody is a minimal WS/2 search document naming ids in order.
func searchBody(ids ...string) []byte {
	releases := make([]map[string]string, len(ids))
	for i, id := range ids {
		releases[i] = map[string]string{"id": id}
	}
	body := map[string]any{
		"count":    len(ids),
		"created":  "2026-01-01T00:00:00.000Z",
		"offset":   0,
		"releases": releases,
	}
	data, err := json.Marshal(body)
	if err != nil {
		panic(fmt.Sprintf("shadow: marshaling test search body: %v", err))
	}
	return data
}

func TestMatchAlbumTakesTheIDConsensusEarlyReturn(t *testing.T) {
	const albumID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	items := []matcher.ItemFields{{
		Artist:    "Cabinet",
		Album:     "Windrose",
		Title:     "Track One",
		MBAlbumID: albumID,
	}}
	srv := releaseServer(t, map[string][]byte{
		albumID: releaseBody(albumID, "Windrose", "Cabinet", []string{"Track One"}),
	}, nil)
	budget := musicbrainz.Unmetered()

	prop, requests, bodies, err := matchAlbum(context.Background(), Config{BaseURL: srv.URL, UserAgent: testUserAgent}, budget, items, passNow)
	if err != nil {
		t.Fatalf("matchAlbum: %v", err)
	}
	if len(requests) != 1 {
		t.Fatalf("requests = %d, want 1 (the early return issues exactly one lookup)", len(requests))
	}
	if len(bodies) != 1 {
		t.Fatalf("bodies = %d, want 1", len(bodies))
	}
	if prop.Action != matcher.ActionApply {
		t.Errorf("Action = %v, want apply", prop.Action)
	}
	if len(prop.Candidates) == 0 || prop.Candidates[0].Info.AlbumID != albumID {
		t.Errorf("Candidates[0].Info.AlbumID = %+v, want %s", prop.Candidates, albumID)
	}
	if budget.Spent() != 1 {
		t.Errorf("budget.Spent() = %d, want 1", budget.Spent())
	}
}

func TestMatchAlbumSearchesWhenThereIsNoConsensus(t *testing.T) {
	const id1, id2 = "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"
	items := []matcher.ItemFields{{
		Artist: "No Consensus Artist",
		Album:  "No Consensus Album",
		Title:  "Track One",
	}}
	srv := releaseServer(t, map[string][]byte{
		"search": searchBody(id1, id2),
		id1:      releaseBody(id1, "No Consensus Album", "No Consensus Artist", []string{"Track One"}),
		id2:      releaseBody(id2, "No Consensus Album", "No Consensus Artist", []string{"Track One"}),
	}, nil)
	budget := musicbrainz.Unmetered()

	_, requests, bodies, err := matchAlbum(context.Background(), Config{BaseURL: srv.URL, UserAgent: testUserAgent}, budget, items, passNow)
	if err != nil {
		t.Fatalf("matchAlbum: %v", err)
	}
	if len(requests) != 3 {
		t.Fatalf("requests = %d, want 3 (one search plus one lookup per returned id)", len(requests))
	}
	if !strings.HasPrefix(requests[0].PathQuery, "/ws/2/release?") {
		t.Errorf("first request = %q, want the search issued first", requests[0].PathQuery)
	}
	for i, id := range []string{id1, id2} {
		if !strings.HasPrefix(requests[i+1].PathQuery, "/ws/2/release/"+id) {
			t.Errorf("request %d = %q, want a lookup of %s in response order", i+1, requests[i+1].PathQuery, id)
		}
	}
	if len(bodies) != 3 {
		t.Fatalf("bodies = %d, want 3", len(bodies))
	}
	if budget.Spent() != len(requests) {
		t.Errorf("budget.Spent() = %d, want %d (one token per request)", budget.Spent(), len(requests))
	}
}

func TestMatchAlbumSpendsOneTokenPerAttemptIncludingRetries(t *testing.T) {
	// TWO collapses are needed (fold, M-b). Retry-After: 0 collapses the
	// retry backoff; cfg.Interval at time.Millisecond collapses the
	// LIMITER's own pacing, because the budget here is metered (mutation
	// 4 forbids Unmetered).
	srv := releaseServer(t, nil, map[string]int{"11111111-1111-1111-1111-111111111111": http.StatusServiceUnavailable})

	limiter := musicbrainz.NewLimiter(time.Millisecond)
	t.Cleanup(func() { _ = limiter.Close() })
	budget := limiter.Budget("shadow-test", musicbrainz.ClassBackground, 0)

	items := []matcher.ItemFields{{
		Artist:    "Retry Artist",
		Album:     "Retry Album",
		Title:     "Track One",
		MBAlbumID: "11111111-1111-1111-1111-111111111111",
	}}

	done := make(chan error, 1)
	go func() {
		_, _, _, err := matchAlbum(context.Background(), Config{BaseURL: srv.URL, UserAgent: testUserAgent}, budget, items, passNow)
		done <- err
	}()

	select {
	case err := <-done:
		if !errors.Is(err, musicbrainz.ErrUnavailable) {
			t.Errorf("matchAlbum error = %v, want a wrapped musicbrainz.ErrUnavailable", err)
		}
	case <-time.After(time.Second):
		t.Fatal("took over a second: Retry-After: 0 and a millisecond interval must both collapse (fold M-b)")
	}
	if budget.Spent() != 7 {
		t.Errorf("budget.Spent() = %d, want 7 (maxAttempts, every attempt spends a token)", budget.Spent())
	}
}

func TestMatchAlbumRunsUnderFullCompatAndQuietImportTunables(t *testing.T) {
	// Revision 1 proposed asserting Action is apply or skip and never
	// asis or ask. That reasoning is wrong (fold, m11): under Quiet,
	// ActionFor never reaches Timid or NoneRecAction (verdict.go:179-
	// 210), so a mutant that scores under DefaultTunables() with only
	// Quiet manually set produces an IDENTICAL Action and Proposal,
	// because the two configurations differ only in NoneRecAction,
	// which is dead code whenever Quiet is set. The hash is the one
	// observable that still moves: shadowScoring's hash must equal the
	// two production constructors' own hash, independent of whatever
	// shadowScoring's body actually does.
	got := configHash(shadowScoring())
	want := configHash(matcher.QuietImportTunables(), matcher.FullCompat())
	if got != want {
		t.Errorf("configHash(shadowScoring()) = %s, want %s (matchAlbum must score under exactly matcher.QuietImportTunables() and matcher.FullCompat())", got, want)
	}
}

func TestConfigHashMovesWithEitherInput(t *testing.T) {
	base := configHash(matcher.QuietImportTunables(), matcher.FullCompat())
	if configHash(matcher.QuietImportTunables(), matcher.DefaultCompat()) == base {
		t.Error("the hash does not move with Compat")
	}
	tun := matcher.QuietImportTunables()
	tun.DistanceWeights["album"] = 4.0
	if configHash(tun, matcher.FullCompat()) == base {
		t.Error("the hash does not move with Tunables")
	}
	if configHash(matcher.QuietImportTunables(), matcher.FullCompat()) != base {
		t.Error("the hash is not stable across two calls with equal inputs")
	}
}
