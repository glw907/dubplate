package grade

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
)

const (
	emptyStatesLedger  = "testdata/empty-states-ledger.json"
	emptyStatesGolden  = "testdata/empty-states-report.golden"
	emptyStatesSession = sessionsIn + "/empty-states"
)

// TestEmptyStatesSessionLoads loads the committed session through the
// loader: every request is answered by both oracles.
func TestEmptyStatesSessionLoads(t *testing.T) {
	s, err := session.Load(emptyStatesSession)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Requests) == 0 {
		t.Fatal("session holds no requests")
	}
	for _, r := range s.Requests {
		for _, src := range []string{"navidrome", "gonic"} {
			found := false
			for _, resp := range r.Responses {
				found = found || resp.Source == src
			}
			if !found {
				t.Errorf("%s has no %s response", r.Endpoint, src)
			}
		}
	}
}

// TestEmptyStatesGrade grades the committed empty-states session,
// Navidrome against gonic, under the test-only ledger and pins the report.
func TestEmptyStatesGrade(t *testing.T) {
	g, err := New(seedConfig(emptyStatesLedger))
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(context.Background(), emptyStatesSession, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	got := r.Format()
	if *update {
		if err := atomicfile.WriteFile(emptyStatesGolden, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(emptyStatesGolden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("report differs from %s; rerun with -update and review the diff", emptyStatesGolden)
	}
	again, err := g.Grade(context.Background(), emptyStatesSession, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again.Format(), got) {
		t.Error("a second grade of the same session formats differently")
	}

	// The study's predictions, each pinned as the recording shows it. An
	// absent prediction is a fact about the oracles, so it stays here as
	// a false case rather than a failure.
	predictions := []struct {
		name    string
		request string
		kind    Kind
		path    string
		row     string
		present bool
	}{
		{"gonic structuredLyrics null", "getLyricsBySongId.json&id=song%3AAlpha+Artist%2FSingle+Disc%2FTogether", KindContract, "subsonic-response/lyricsList/structuredLyrics", "gonic-structuredlyrics-null", true},
		{"navidrome idle nowPlaying without entry", "getNowPlaying.json", KindContract, "subsonic-response/nowPlaying/entry", "navidrome-nowplaying-entry", true},
		{"gonic error 70 for getNowPlaying (JSON)", "getNowPlaying.json", KindPresence, "subsonic-response/error", "gonic-nowplaying-error", true},
		{"gonic error 70 for getNowPlaying (XML)", "getNowPlaying.xml", KindPresence, "subsonic-response/error", "gonic-nowplaying-error", true},
		{"mixed empty containers: getPlaylists", "getPlaylists.json", KindPresence, "subsonic-response/playlists/playlist", "playlists-empty-container", true},
		{"mixed empty containers: getInternetRadioStations", "getInternetRadioStations.json", KindPresence, "subsonic-response/internetRadioStations", "", false},
		{"mixed empty containers: getStarred2", "getStarred2.json", KindPresence, "subsonic-response/starred2", "", false},
	}
	for _, p := range predictions {
		if got := hasFinding(r, p.request, p.kind, p.path, p.row); got != p.present {
			t.Errorf("%s: finding at %s present = %v, want %v", p.name, p.path, got, p.present)
		}
	}

	l, err := LoadLedger(emptyStatesLedger, sessionsIn)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range l.Rows {
		if r.ledger.Counts[row.ID] < 1 {
			t.Errorf("row %s matched nothing", row.ID)
		}
	}
}
