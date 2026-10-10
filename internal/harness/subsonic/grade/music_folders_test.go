package grade

import (
	"bytes"
	"encoding/json/v2"
	"os"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
)

const (
	musicFoldersLedger  = "testdata/music-folders-ledger.json"
	musicFoldersGolden  = "testdata/music-folders-report.golden"
	musicFoldersSession = sessionsIn + "/music-folders"
)

// musicFoldersConfig is the grader the music-folders session was recorded
// for: its two collections, under the names the recording gave them.
func musicFoldersConfig() Config {
	cfg := seedConfig(musicFoldersLedger)
	cfg.Collections = []Collection{{"first", collectionA}, {"second", collectionB}}
	return cfg
}

// folderRequests are the endpoints that take a musicFolderId, each with the
// list members its answer carries.
var folderRequests = map[string][]string{
	"getIndexes":      {"indexes", "index"},
	"getArtists":      {"artists", "index"},
	"getAlbumList":    {"albumList", "album"},
	"getAlbumList2":   {"albumList2", "album"},
	"getRandomSongs":  {"randomSongs", "song"},
	"getSongsByGenre": {"songsByGenre", "song"},
	"getStarred":      {"starred", "artist", "album", "song"},
	"getStarred2":     {"starred2", "artist", "album", "song"},
	"search2":         {"searchResult2", "artist", "album", "song"},
	"search3":         {"searchResult3", "artist", "album", "song"},
}

// folderBody returns the response source gave to the request for endpoint whose
// musicFolderId parameter is folder ("" for absent), decoded as JSON.
func folderBody(t *testing.T, s *session.Session, endpoint, folder, source string) map[string]map[string]any {
	t.Helper()
	for _, req := range s.Requests {
		if req.Endpoint != endpoint || req.Format != "json" {
			continue
		}
		got := ""
		for _, p := range req.Params {
			if p.Name == "musicFolderId" {
				got = p.Value
			}
		}
		if got != folder {
			continue
		}
		for _, resp := range req.Responses {
			if resp.Source != source {
				continue
			}
			var body map[string]map[string]any
			if err := json.Unmarshal(resp.Body, &body); err != nil {
				t.Fatalf("%s %q on %s: %v", endpoint, folder, source, err)
			}
			return body
		}
	}
	t.Fatalf("session has no %s answer to %s with musicFolderId %q", source, endpoint, folder)
	return nil
}

// TestMusicFoldersSessionLoads loads the committed session: every request is
// answered by both oracles, and the two collection keys resolve on both.
func TestMusicFoldersSessionLoads(t *testing.T) {
	s, err := session.Load(musicFoldersSession)
	if err != nil {
		t.Fatal(err)
	}
	if want := 2 + 4*len(folderRequests); len(s.Requests) != want {
		t.Errorf("requests = %d, want %d", len(s.Requests), want)
	}
	for _, r := range s.Requests {
		if len(r.Responses) != 2 {
			t.Errorf("%s answered by %d sources, want 2", r.Endpoint, len(r.Responses))
		}
	}
	for _, key := range []string{"collection:first", "collection:second"} {
		for _, src := range []string{"navidrome", "gonic"} {
			if s.Resolutions[key][src] == "" {
				t.Errorf("no %s resolution for %s", src, key)
			}
		}
	}
}

// TestMusicFoldersRecordedBehavior asserts the two behaviors the session
// records, on the bodies themselves: an unknown music folder fails on
// Navidrome and answers empty on gonic, and a filtered getArtists lists only
// the filtered collection's artists on both.
func TestMusicFoldersRecordedBehavior(t *testing.T) {
	s, err := session.Load(musicFoldersSession)
	if err != nil {
		t.Fatal(err)
	}

	for endpoint, members := range folderRequests {
		t.Run("unknown folder, "+endpoint, func(t *testing.T) {
			nav := folderBody(t, s, endpoint, "999", "navidrome")["subsonic-response"]
			if nav["status"] != "failed" {
				t.Errorf("navidrome status = %v, want failed", nav["status"])
			}
			if e, _ := nav["error"].(map[string]any); e["code"] != float64(70) {
				t.Errorf("navidrome error = %v, want code 70", nav["error"])
			}
			gon := folderBody(t, s, endpoint, "999", "gonic")["subsonic-response"]
			if gon["status"] != "ok" {
				t.Errorf("gonic status = %v, want ok", gon["status"])
			}
			result, _ := gon[members[0]].(map[string]any)
			for _, list := range members[1:] {
				if items, _ := result[list].([]any); len(items) != 0 {
					t.Errorf("gonic %s.%s holds %d entries, want none", members[0], list, len(items))
				}
			}
		})
	}

	// Each collection's album artists, as the manifests declare them.
	collections := map[string][]string{
		"collection:first":  {"Alpha Artist", "Borealis Trio", "Delta Duo", "Zoë Ångström"},
		"collection:second": {"Alpha Artist", "Carmine Ensemble", "The Meridians", "Various Artists"},
	}
	for key, want := range collections {
		for _, src := range []string{"navidrome", "gonic"} {
			t.Run("filtered getArtists, "+key+" on "+src, func(t *testing.T) {
				resp := folderBody(t, s, "getArtists", key, src)["subsonic-response"]
				artists, _ := resp["artists"].(map[string]any)
				indexes, _ := artists["index"].([]any)
				var got []string
				for _, ix := range indexes {
					list, _ := ix.(map[string]any)["artist"].([]any)
					for _, a := range list {
						name, _ := a.(map[string]any)["name"].(string)
						got = append(got, name)
					}
				}
				slices.Sort(got)
				if !slices.Equal(got, want) {
					t.Errorf("artists = %q, want %q", got, want)
				}
			})
		}
	}
}

// TestMusicFoldersGrade grades the committed music-folders session,
// Navidrome against gonic, under the test-only ledger and pins the report.
func TestMusicFoldersGrade(t *testing.T) {
	g, err := New(musicFoldersConfig())
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(t.Context(), musicFoldersSession, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	got := r.Format()
	if *update {
		if err := atomicfile.WriteFile(musicFoldersGolden, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(musicFoldersGolden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("report differs from %s; rerun with -update and review the diff", musicFoldersGolden)
	}
	again, err := g.Grade(t.Context(), musicFoldersSession, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again.Format(), got) {
		t.Error("a second grade of the same session formats differently")
	}

	l, err := LoadLedger(musicFoldersLedger, sessionsIn)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range l.Rows {
		if r.ledger.Counts[row.ID] < 1 {
			t.Errorf("row %s matched nothing", row.ID)
		}
	}
}
