package tag

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/musicbrainz"
)

// evalDate is the fixed evaluation date every test in this package
// shares, never a clock read (contract 9).
var evalDate = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

// taggingServer serves the release and search bodies committed under
// testdata/name, and records every request path (including its query
// string) in the order the flow issues them.
func taggingServer(t *testing.T, name string) (*httptest.Server, *[]string) {
	t.Helper()
	dir := filepath.Join("testdata", name)
	var paths []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.RequestURI())

		var file string
		if r.URL.Path == "/ws/2/release" {
			file = "search.json"
		} else {
			id := strings.TrimPrefix(r.URL.Path, "/ws/2/release/")
			file = "release-" + id + ".json"
		}

		body, err := os.ReadFile(filepath.Join(dir, file))
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(body) //nolint:errcheck // best-effort write to a test stub; the client sees a truncated body either way
	}))
	t.Cleanup(srv.Close)
	return srv, &paths
}

// statusServer answers every request with status, alongside
// Retry-After: 0 so a retryable status collapses its backoff schedule
// to no real wall time (fold M-h): without it maxAttempts (7) with
// backoff waits 0+1+2+4+8+16 spends 31 real seconds per gate run.
func statusServer(t *testing.T, status int) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "0")
		w.WriteHeader(status)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// clientFor builds a *musicbrainz.Client against srv with an unmetered
// budget, the shape every conformance run and this package's own tests
// share.
func clientFor(t *testing.T, srv *httptest.Server) *musicbrainz.Client {
	t.Helper()
	c, err := musicbrainz.NewClient(musicbrainz.ClientConfig{
		BaseURL:   srv.URL,
		UserAgent: "dubplate-test/0 ( test )",
		Budget:    musicbrainz.Unmetered(),
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// earlyReturnItems is the 167-case shape: a unanimous mb_albumid whose
// lookup scores strong (every field matches testdata/early-return's
// candidate exactly, so DistanceForAlbum's raw distance is 0.0).
func earlyReturnItems() []matcher.ItemFields {
	return []matcher.ItemFields{{
		Artist:    "Early Return Artist",
		Album:     "Early Return Album",
		Title:     "Track One",
		MBAlbumID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
	}}
}

// mediumIDItems is earlyReturnItems' fixture with one deliberate,
// deterministic mismatch: DiscTotal 2 against the candidate's single
// medium. AddNumber's mediums component then adds one 1.0 penalty at
// weight 1.0, and with album_id (weight 5), artist (3), album (3) and
// tracks (2) all matching exactly, the resulting distance is
// 1/14 = 0.0714286: above StrongRecThresh (0.04) and at or below
// MediumRecThresh (0.25), so this ID candidate scores medium rather
// than strong (revision 2, fold M-e).
func mediumIDItems() []matcher.ItemFields {
	items := earlyReturnItems()
	items[0].DiscTotal = 2
	return items
}

// searchFallbackItems carries the same deliberate mediums mismatch as
// mediumIDItems, against testdata/search-fallback's ID candidate, so
// its lookup scores medium and the flow falls through to a search.
func searchFallbackItems() []matcher.ItemFields {
	return []matcher.ItemFields{{
		Artist:    "Fallback Artist",
		Album:     "Fallback Album",
		Title:     "Track One",
		MBAlbumID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
		DiscTotal: 2,
	}}
}

// itemsWithoutAlbumID carries no mb_albumid at all, so match_by_id's
// first early branch fires (match.py:252-266) and no lookup is issued.
func itemsWithoutAlbumID() []matcher.ItemFields {
	return []matcher.ItemFields{{
		Artist: "No ID Artist",
		Album:  "No ID Album",
		Title:  "Track One",
	}}
}

// itemsWithSplitAlbumID carries two disagreeing mb_albumid values, so
// consensus on the field is false and match_by_id's second early
// branch fires (match.py:252-266) even though a likely value exists.
func itemsWithSplitAlbumID() []matcher.ItemFields {
	return []matcher.ItemFields{
		{Artist: "Split Artist", Album: "Split Album", Title: "Track One", MBAlbumID: "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"},
		{Artist: "Split Artist", Album: "Split Album", Title: "Track Two", MBAlbumID: "dddddddd-dddd-dddd-dddd-dddddddddddd"},
	}
}

func TestAlbumTakesTheIDMatchEarlyReturnAndIssuesNoSearch(t *testing.T) {
	// The 167-case shape: a unanimous mb_albumid whose lookup scores
	// strong. The assertion that matters is the REQUEST SET, because
	// that is what conform's tier 1 grades: exactly one lookup, no
	// search. The mutation that proves it fires: relax the rec ==
	// strong guard, which turns this into two requests.
	srv, paths := taggingServer(t, "early-return")
	p, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if p.Recommendation != matcher.RecStrong {
		t.Fatalf("Recommendation = %v, want strong", p.Recommendation)
	}
	if got := *paths; len(got) != 1 || !strings.HasPrefix(got[0], "/ws/2/release/") {
		t.Errorf("issued %v, want exactly one release lookup and no search", got)
	}
}

func TestAlbumSearchesWhenTheIDMatchIsExactlyMedium(t *testing.T) {
	// The detector for "relax rec == strong to rec >= medium", which
	// the early-return test above cannot see: its fixture scores strong
	// and a strong match stays on the same branch under either guard
	// (revision 2, fold M-e). This fixture's single ID candidate scores
	// medium, so the flow must fall through: one lookup, then one
	// search. Relaxing the guard turns that into one lookup and no
	// search.
	srv, paths := taggingServer(t, "medium-id")
	p, err := Album(context.Background(), clientFor(t, srv), mediumIDItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if p.Recommendation == matcher.RecStrong {
		t.Fatal("the fixture's ID candidate scores strong, so it can no longer discriminate the guard")
	}
	var searches int
	for _, path := range *paths {
		if strings.HasPrefix(path, "/ws/2/release?") {
			searches++
		}
	}
	if searches != 1 {
		t.Errorf("issued %v, want one lookup and one search: the early return needs rec == strong (match.py:436)", *paths)
	}
}

func TestAlbumFallsThroughToSearchWhenTheIDMatchIsNotStrong(t *testing.T) {
	srv, paths := taggingServer(t, "search-fallback")
	p, err := Album(context.Background(), clientFor(t, srv), searchFallbackItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if len(p.Candidates) < 2 {
		t.Fatalf("Candidates = %d, want the ID candidate plus at least one search candidate", len(p.Candidates))
	}
	var searches int
	for _, path := range *paths {
		if strings.HasPrefix(path, "/ws/2/release?") {
			searches++
		}
	}
	if searches != 1 {
		t.Errorf("issued %d searches, want exactly 1", searches)
	}
}

func TestAlbumTimidNeverTakesTheEarlyReturn(t *testing.T) {
	// import.timid is false in all three committed profiles, so this
	// branch is corpus-unreachable (declared gap 4). The hand fixture is
	// the early-return case with Timid set: the same strong ID match
	// must now also search.
	tun := matcher.DefaultTunables()
	tun.Timid = true
	srv, paths := taggingServer(t, "early-return")
	if _, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), tun, matcher.DefaultCompat(), evalDate, Options{}); err != nil {
		t.Fatalf("Album: %v", err)
	}
	if len(*paths) < 2 {
		t.Errorf("issued %v, want a search as well: timid forbids the early return (match.py:432)", *paths)
	}
}

func TestAlbumExplicitSearchIDsSkipTheSearchAndTakeNoEarlyReturn(t *testing.T) {
	// import.search_ids is set by no committed case (declared gap 1).
	// Two ids in, two lookups out, no search, and no early return even
	// when the first scores strong.
	srv, paths := taggingServer(t, "explicit-ids")
	p, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate,
		Options{SearchIDs: []string{"11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"}})
	if err != nil {
		t.Fatalf("Album: %v", err)
	}
	if len(p.Candidates) != 2 {
		t.Errorf("Candidates = %d, want 2", len(p.Candidates))
	}
	for _, path := range *paths {
		if strings.HasPrefix(path, "/ws/2/release?") {
			t.Errorf("issued a search under search_ids: match.py:417-420 skips it entirely")
		}
	}
}

func TestAlbumMatchByIDEarlyBranches(t *testing.T) {
	// Both branches are corpus-unreachable: all 216 committed cases
	// carry a full mb_albumid consensus (declared gaps 2 and 3). Each
	// must skip the lookup and go straight to the search.
	tests := []struct {
		name  string
		items []matcher.ItemFields
	}{
		{"no album id at all", itemsWithoutAlbumID()},
		{"no consensus on the album id", itemsWithSplitAlbumID()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, paths := taggingServer(t, "search-fallback")
			if _, err := Album(context.Background(), clientFor(t, srv), tt.items, matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{}); err != nil {
				t.Fatalf("Album: %v", err)
			}
			if strings.HasPrefix((*paths)[0], "/ws/2/release/") {
				t.Errorf("first request was a lookup (%s); match_by_id returns () here (match.py:252-266)", (*paths)[0])
			}
		})
	}
}

func TestAlbumDropsA404LookupAndPropagatesEverythingElse(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr error
	}{
		{"404 drops the candidate", http.StatusNotFound, ErrNoCandidates},
		{"503 propagates", http.StatusServiceUnavailable, musicbrainz.ErrUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// statusServer sends Retry-After: 0 alongside the status.
			// Without it the 503 row spends 31 seconds of real wall
			// time on every gate run: maxAttempts is 7
			// (client.go:42), 503 is retryable (:50-56), and the six
			// backoff waits are 0+1+2+4+8+16 (:418-423). The sleep
			// hook is unexported and ClientConfig exposes no way to
			// inject one (:93-99, :118), so the header is the only
			// lever a caller outside musicbrainz has, and retryWait
			// parses delta-seconds ahead of the backoff (:403-411).
			// Revision 2, fold M-h.
			//
			// The one-second budget is proved with a time.After race
			// rather than time.Now/time.Since: internal/engine must
			// never read the clock (TestEngineNeverReadsTheClock walks
			// this file too), and time.After makes no such call.
			srv := statusServer(t, tt.status)
			done := make(chan error, 1)
			go func() {
				_, err := Album(context.Background(), clientFor(t, srv), earlyReturnItems(), matcher.DefaultTunables(), matcher.DefaultCompat(), evalDate, Options{SearchIDs: []string{"11111111-1111-1111-1111-111111111111"}})
				done <- err
			}()
			select {
			case err := <-done:
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("Album = %v, want an error wrapping %v", err, tt.wantErr)
				}
			case <-time.After(time.Second):
				t.Fatal("took over a second; statusServer must send Retry-After: 0 so the retry schedule collapses (fold M-h)")
			}
		})
	}
}
