package grade

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json/v2"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/opensubsonic/conformance"
)

const (
	openapiDir  = "../data/openapi"
	seedSession = "../data/sessions/seed"
)

func TestLoadContract(t *testing.T) {
	sum := func(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
	garbage := []byte(`[1,2]`)
	tests := []struct {
		name    string
		file    []byte // written as openapi.json; nil leaves the directory empty
		sum     string
		wantErr []string // substrings of the error; none means success
	}{
		{name: "pinned document", file: mustRead(t, filepath.Join(openapiDir, openapiFile)), sum: pinnedDocSum},
		{name: "absent", sum: pinnedDocSum, wantErr: []string{openapiFile}},
		{
			name: "checksum mismatch", file: []byte(`{}`), sum: pinnedDocSum,
			wantErr: []string{openapiFile, sum([]byte(`{}`)), pinnedDocSum},
		},
		{name: "does not load", file: garbage, sum: sum(garbage), wantErr: []string{openapiFile}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.file != nil {
				if err := os.WriteFile(filepath.Join(dir, openapiFile), tt.file, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			c, err := loadContract(dir, tt.sum)
			if len(tt.wantErr) == 0 {
				if err != nil || c == nil {
					t.Fatalf("loadContract = %v, %v; want a checker", c, err)
				}
				return
			}
			if err == nil {
				t.Fatal("loadContract succeeded")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error %q lacks %q", err, want)
				}
			}
		})
	}
}

func TestPinnedOpenAPISumMatchesTheCommittedFile(t *testing.T) {
	if _, err := loadContract(openapiDir, pinnedDocSum); err != nil {
		t.Fatal(err)
	}
}

// TestSeedGetAlbumContract validates the recorded getAlbum bodies of both
// oracles against the pinned document and pins each oracle's exact
// violations: Navidrome reports an album userRating of 0 (the document
// says at least 1), and gonic reports a null replayGain on every song.
func TestSeedGetAlbumContract(t *testing.T) {
	c, err := loadContract(openapiDir, pinnedDocSum)
	if err != nil {
		t.Fatal(err)
	}
	s, err := session.Load(seedSession)
	if err != nil {
		t.Fatal(err)
	}
	const (
		navidromeReason = "number must be at least 1"
		gonicReason     = "Value is not nullable"
	)
	songs := map[string]int{
		"album:Alpha Artist/Single Disc":           3,
		"album:Alpha Artist/Greatest Hits":         2,
		"album:Delta Duo/Greatest Hits":            2,
		"album:Various Artists/Mixtape Volume One": 3,
		"album:Carmine Ensemble/Two Sides":         4,
	}
	seen := map[string]int{}
	for _, req := range s.Requests {
		if req.Endpoint != "getAlbum" || req.Format != "json" {
			continue
		}
		album := paramValue(req, "id")
		n, ok := songs[album]
		if !ok {
			t.Fatalf("request for unexpected album %q", album)
		}
		for _, resp := range req.Responses {
			var want []conformance.Violation
			var drop []string // pointers whose removal repairs the body
			switch resp.Source {
			case "navidrome":
				want = []conformance.Violation{{Pointer: "/subsonic-response/album/userRating", Reason: navidromeReason}}
			case "gonic":
				for i := range n {
					p := "/subsonic-response/album/song/" + strconv.Itoa(i) + "/replayGain"
					want = append(want, conformance.Violation{Pointer: p, Reason: gonicReason})
				}
			default:
				t.Fatalf("unexpected source %q", resp.Source)
			}
			for _, v := range want {
				drop = append(drop, v.Pointer)
			}
			seen[resp.Source]++
			t.Run(resp.Source+"/"+album, func(t *testing.T) {
				got := c.json.Check(context.Background(), http.MethodGet, "/rest/getAlbum", resp.Status, resp.ContentType, resp.Body)
				if got.Outcome != conformance.Invalid || !slices.Equal(got.Violations, want) {
					t.Fatalf("got %s %+v\nwant %+v", got.Outcome, got.Violations, want)
				}
				repaired := removePointers(t, resp.Body, drop)
				fixed := c.json.Check(context.Background(), http.MethodGet, "/rest/getAlbum", resp.Status, resp.ContentType, repaired)
				if fixed.Outcome != conformance.Valid {
					t.Errorf("repaired body: %s %+v", fixed.Outcome, fixed.Violations)
				}
			})
		}
	}
	for _, src := range []string{"navidrome", "gonic"} {
		if seen[src] != len(songs) {
			t.Errorf("%s: checked %d getAlbum bodies, want %d", src, seen[src], len(songs))
		}
	}
}

func paramValue(r session.Request, name string) string {
	for _, p := range r.Params {
		if p.Name == name {
			return p.Value
		}
	}
	return ""
}

// removePointers deletes each pointer's member from the JSON body and
// returns the re-encoded document.
func removePointers(t *testing.T, body []byte, pointers []string) []byte {
	t.Helper()
	var doc any
	if err := json.Unmarshal(body, &doc); err != nil {
		t.Fatal(err)
	}
	for _, p := range pointers {
		segs := strings.Split(strings.TrimPrefix(p, "/"), "/")
		node := doc
		for _, seg := range segs[:len(segs)-1] {
			switch n := node.(type) {
			case map[string]any:
				node = n[seg]
			case []any:
				i, err := strconv.Atoi(seg)
				if err != nil {
					t.Fatalf("pointer %s: %v", p, err)
				}
				node = n[i]
			}
		}
		m, ok := node.(map[string]any)
		if !ok {
			t.Fatalf("pointer %s: parent is not an object", p)
		}
		delete(m, segs[len(segs)-1])
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
