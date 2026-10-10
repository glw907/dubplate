package main

import (
	"context"
	"encoding/json/v2"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/harness/subsonic/stack"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

const (
	cannedNavidromeVersion = "0.99.1 (nd-only-7f3a)"
	cannedGonicVersion     = "0.88.2-gonic-only-91c4"
	pinnedNavidrome        = "sha256:38dc2727bfcfd5ede290f8ada114fc90368146f265ae4701ddddbcbe2a44ee52"
	pinnedGonic            = "sha256:516fd9645614ba3a596d86174216c3e944808b9ec970c581678713be4c8b1d49"
)

// cannedOptions shape one canned oracle.
type cannedOptions struct {
	prefix       string // every id the oracle mints starts with it
	version      string
	twoHits      bool   // a second album titled "Greatest Hits" under Alpha Artist
	failEndpoint string // answers this endpoint with HTTP 500
	// payload, when set, rewrites the members each JSON answer carries
	// beside the envelope's own.
	payload func(endpoint, payload string) string
}

// cannedOracle serves the answers of a small library: Alpha Artist and
// Delta Duo each own an album titled "Greatest Hits". It records the
// (endpoint, id) pairs it was asked.
type cannedOracle struct {
	URL string

	mu   sync.Mutex
	asks []string
}

func (c *cannedOracle) asked() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return slices.Clone(c.asks)
}

func newCannedOracle(t *testing.T, o cannedOptions) *cannedOracle {
	t.Helper()
	p := o.prefix
	extra := ""
	if o.twoHits {
		extra = fmt.Sprintf(`,{"id":%q,"name":"Greatest Hits"}`, p+"al9")
	}
	answers := map[string]string{
		"ping|": ``,
		"getArtists|": fmt.Sprintf(`"artists":{"index":[{"name":"A","artist":[{"id":%q,"name":"Alpha Artist"},{"id":%q,"name":"Delta Duo"}]}]}`,
			p+"ar1", p+"ar2"),
		"getIndexes|": fmt.Sprintf(`"indexes":{"index":[{"name":"A","artist":[{"id":%q,"name":"Alpha Artist"},{"id":%q,"name":"Delta Duo"}]}]}`,
			p+"ar1", p+"ar2"),
		"getArtist|" + p + "ar1": fmt.Sprintf(`"artist":{"id":%q,"name":"Alpha Artist","album":[{"id":%q,"name":"Greatest Hits"},{"id":%q,"name":"Single Disc"}%s]}`,
			p+"ar1", p+"al1", p+"al2", extra),
		"getArtist|" + p + "ar2": fmt.Sprintf(`"artist":{"id":%q,"name":"Delta Duo","album":[{"id":%q,"name":"Greatest Hits"}]}`,
			p+"ar2", p+"al3"),
		"getAlbum|" + p + "al1": fmt.Sprintf(`"album":{"id":%q,"name":"Greatest Hits","song":[{"id":%q,"title":"Opening"},{"id":%q,"title":"Together"}]}`,
			p+"al1", p+"s1", p+"s2"),
		"getAlbum|" + p + "al3": fmt.Sprintf(`"album":{"id":%q,"name":"Greatest Hits","song":[{"id":%q,"title":"First Light"}]}`,
			p+"al3", p+"s3"),
		"getSong|" + p + "s3": fmt.Sprintf(`"song":{"id":%q,"title":"First Light"}`, p+"s3"),
		"getMusicDirectory|" + p + "ar1": fmt.Sprintf(`"directory":{"child":[{"id":%q,"title":"Greatest Hits","isDir":true},{"id":%q,"title":"Single Disc","isDir":true},{"id":%q,"title":"cover.jpg","isDir":false}]}`,
			p+"al1", p+"al2", p+"cv"),
		"getMusicDirectory|" + p + "ar2": fmt.Sprintf(`"directory":{"child":[{"id":%q,"title":"Greatest Hits","isDir":true}]}`, p+"al3"),
		"getMusicDirectory|" + p + "al1": fmt.Sprintf(`"directory":{"child":[{"id":%q,"title":"Opening","isDir":false}]}`, p+"s1"),
	}
	c := &cannedOracle{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		endpoint := strings.TrimPrefix(r.URL.Path, "/rest/")
		key := endpoint + "|" + r.URL.Query().Get("id")
		c.mu.Lock()
		c.asks = append(c.asks, key)
		c.mu.Unlock()
		payload, ok := answers[key]
		switch {
		case endpoint == o.failEndpoint:
			http.Error(w, "boom", http.StatusInternalServerError)
		case !ok:
			http.NotFound(w, r)
		case r.URL.Query().Get("f") == "xml":
			w.Header().Set("Content-Type", "application/xml; charset=utf-8")
			_, _ = fmt.Fprintf(w, `<subsonic-response status="ok" version="1.16.1" serverVersion=%q></subsonic-response>`, o.version)
		default:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			if o.payload != nil {
				payload = o.payload(endpoint, payload)
			}
			if payload != "" {
				payload = "," + payload
			}
			_, _ = fmt.Fprintf(w, `{"subsonic-response":{"status":"ok","version":"1.16.1","serverVersion":%q%s}}`, o.version, payload)
		}
	}))
	t.Cleanup(srv.Close)
	c.URL = srv.URL
	return c
}

const happySpec = `{"requests":[
 {"endpoint":"ping","format":"json"},
 {"endpoint":"getArtists","format":"json"},
 {"endpoint":"getIndexes","format":"json"},
 {"endpoint":"getArtist","format":"json","params":{"id":"artist:Alpha Artist"}},
 {"endpoint":"getArtist","format":"json","params":{"id":"artist:Delta Duo"}},
 {"endpoint":"getAlbum","format":"json","params":{"id":"album:Alpha Artist/Greatest Hits"}},
 {"endpoint":"getAlbum","format":"json","params":{"id":"album:Delta Duo/Greatest Hits"}},
 {"endpoint":"getSong","format":"json","params":{"id":"song:Delta Duo/Greatest Hits/First Light"}},
 {"endpoint":"getMusicDirectory","format":"json","params":{"id":"folder:Alpha Artist"}},
 {"endpoint":"getMusicDirectory","format":"json","params":{"id":"folder:Alpha Artist/Greatest Hits"}},
 {"endpoint":"getMusicDirectory","format":"json","params":{"id":"folder:Delta Duo"}},
 {"endpoint":"getAlbum","format":"xml","params":{"id":"album:Delta Duo/Greatest Hits"}},
 {"endpoint":"ping","format":"xml"}
]}`

// recordRun is one record-oracle run against canned oracles.
type recordRun struct {
	out     string
	nd, gon *cannedOracle
	stopped *bool
	started []stack.Collection // what the last startOracles call was given
	cfg     recordConfig
}

// fixtureManifest returns the path of a committed fixture manifest.
func fixtureManifest(name string) string {
	return filepath.Join("..", "..", "internal", "harness", "subsonic", "data", "fixture", name)
}

func newRecordRun(t *testing.T, spec string, nd, gon cannedOptions) *recordRun {
	t.Helper()
	dir := t.TempDir()
	specPath := filepath.Join(dir, "spec.json")
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	r := &recordRun{
		out:     filepath.Join(dir, "session"),
		nd:      newCannedOracle(t, nd),
		gon:     newCannedOracle(t, gon),
		stopped: new(bool),
	}
	r.cfg = recordConfig{
		spec:        specPath,
		collections: []collection{{name: "music", manifest: fixtureManifest("manifest.json")}},
		out:         r.out,
	}
	old := startOracles
	t.Cleanup(func() { startOracles = old })
	startOracles = func(_ context.Context, _ string, cols []stack.Collection) ([]runningOracle, func(), error) {
		r.started = cols
		mk := func(name, url, digest string) runningOracle {
			return runningOracle{
				name:        name,
				client:      &client.Client{BaseURL: url, User: "admin", Password: "pw", Name: "t"},
				imageDigest: digest,
			}
		}
		return []runningOracle{mk("navidrome", r.nd.URL, pinnedNavidrome), mk("gonic", r.gon.URL, pinnedGonic)},
			func() { *r.stopped = true }, nil
	}
	return r
}

func nd() cannedOptions    { return cannedOptions{prefix: "nd-", version: cannedNavidromeVersion} }
func gonic() cannedOptions { return cannedOptions{prefix: "g-", version: cannedGonicVersion} }

func TestRecordOracleResolvesByIdentity(t *testing.T) {
	r := newRecordRun(t, happySpec, nd(), gonic())
	if err := runRecordOracle(t.Context(), r.cfg); err != nil {
		t.Fatal(err)
	}
	if !*r.stopped {
		t.Error("oracles not stopped")
	}
	s, err := session.Load(r.out)
	if err != nil {
		t.Fatalf("load recorded session: %v", err)
	}

	wantRes := map[string]map[string]string{
		"artist:Alpha Artist":                      {"navidrome": "nd-ar1", "gonic": "g-ar1"},
		"artist:Delta Duo":                         {"navidrome": "nd-ar2", "gonic": "g-ar2"},
		"album:Alpha Artist/Greatest Hits":         {"navidrome": "nd-al1", "gonic": "g-al1"},
		"album:Delta Duo/Greatest Hits":            {"navidrome": "nd-al3", "gonic": "g-al3"},
		"song:Delta Duo/Greatest Hits/First Light": {"navidrome": "nd-s3", "gonic": "g-s3"},
		"folder:Alpha Artist":                      {"navidrome": "nd-ar1", "gonic": "g-ar1"},
		"folder:Alpha Artist/Greatest Hits":        {"navidrome": "nd-al1", "gonic": "g-al1"},
		"folder:Delta Duo":                         {"navidrome": "nd-ar2", "gonic": "g-ar2"},
	}
	for key, want := range wantRes {
		for src, id := range want {
			if got := s.Resolutions[key][src]; got != id {
				t.Errorf("resolution %q on %s = %q, want %q", key, src, got, id)
			}
		}
	}
	if len(s.Resolutions) != len(wantRes) {
		t.Errorf("resolution keys = %d, want %d", len(s.Resolutions), len(wantRes))
	}

	// The oracle was asked for the resolved id, never the symbol.
	for _, c := range []struct {
		o      *cannedOracle
		prefix string
	}{{r.nd, "nd-"}, {r.gon, "g-"}} {
		asks := c.o.asked()
		for _, want := range []string{"getAlbum|" + c.prefix + "al3", "getAlbum|" + c.prefix + "al1", "getSong|" + c.prefix + "s3", "getMusicDirectory|" + c.prefix + "al1"} {
			if !slices.Contains(asks, want) {
				t.Errorf("%s oracle never asked %q; asked %q", c.prefix, want, asks)
			}
		}
	}

	hex := regexp.MustCompile(`^[0-9a-f]{64}$`)
	nds, gos := s.Sources["navidrome"], s.Sources["gonic"]
	if nds.ServerVersion != cannedNavidromeVersion || gos.ServerVersion != cannedGonicVersion {
		t.Errorf("serverVersion = %q / %q, want each oracle's own envelope version", nds.ServerVersion, gos.ServerVersion)
	}
	if nds.ImageDigest != pinnedNavidrome || gos.ImageDigest != pinnedGonic {
		t.Errorf("image digests = %q / %q", nds.ImageDigest, gos.ImageDigest)
	}
	if !hex.MatchString(nds.LibraryDigest) || nds.LibraryDigest != gos.LibraryDigest {
		t.Errorf("library digests = %q / %q, want one 64-hex digest", nds.LibraryDigest, gos.LibraryDigest)
	}
	if len(s.Requests) != 13 {
		t.Fatalf("requests = %d, want 13", len(s.Requests))
	}
	for i, req := range s.Requests {
		if len(req.Responses) != 2 {
			t.Errorf("request %d answered by %d sources, want 2", i+1, len(req.Responses))
		}
	}
	// Symbols stay symbolic in the stored request.
	if got := s.Requests[6].Params[0].Value; got != "album:Delta Duo/Greatest Hits" {
		t.Errorf("stored param = %q, want the symbolic key", got)
	}
}

func TestRecordOracleRefusals(t *testing.T) {
	withReq := func(reqs ...string) string {
		return `{"requests":[` + strings.Join(reqs, ",") + `]}`
	}
	artists := `{"endpoint":"getArtists","format":"json"}`
	alpha := `{"endpoint":"getArtist","format":"json","params":{"id":"artist:Alpha Artist"}}`
	delta := `{"endpoint":"getArtist","format":"json","params":{"id":"artist:Delta Duo"}}`
	album := func(key string) string {
		return `{"endpoint":"getAlbum","format":"json","params":{"id":"` + key + `"}}`
	}
	tests := []struct {
		name    string
		spec    string
		nd, gon cannedOptions
		want    error
		status  int // HTTP status the failure must carry, when want is nil
		mention []string
	}{
		{
			name: "two albums with one title under one artist",
			spec: withReq(artists, alpha, album("album:Alpha Artist/Greatest Hits")),
			nd:   cannedOptions{prefix: "nd-", version: "v", twoHits: true}, gon: gonic(),
			want: errAmbiguous, mention: []string{"album:Alpha Artist/Greatest Hits", "navidrome"},
		},
		{
			name: "album no earlier response reaches",
			spec: withReq(artists, alpha, delta, album("album:Alpha Artist/Nonesuch")),
			nd:   nd(), gon: gonic(),
			want: errUnreachable, mention: []string{"album:Alpha Artist/Nonesuch", "navidrome"},
		},
		{
			name: "album whose artist was never listed",
			spec: withReq(alpha, album("album:Alpha Artist/Greatest Hits")),
			nd:   nd(), gon: gonic(),
			want: errUnreachable, mention: []string{"artist:Alpha Artist"},
		},
		{
			name: "album before its artist's answer",
			spec: withReq(artists, album("album:Alpha Artist/Greatest Hits"), alpha),
			nd:   nd(), gon: gonic(),
			want: errUnreachable, mention: []string{"album:Alpha Artist/Greatest Hits"},
		},
		{
			name: "malformed album reference",
			spec: withReq(artists, album("album:no-slash")),
			nd:   nd(), gon: gonic(),
			want: errBadRef, mention: []string{"album:no-slash"},
		},
		{
			name: "oracle failing mid-run",
			spec: withReq(artists, alpha, album("album:Alpha Artist/Greatest Hits")),
			nd:   nd(), gon: cannedOptions{prefix: "g-", version: "v", failEndpoint: "getAlbum"},
			status:  500,
			mention: []string{"gonic", "getAlbum", "request 3"},
		},
		{
			name: "empty serverVersion",
			spec: withReq(artists),
			nd:   cannedOptions{prefix: "nd-", version: ""}, gon: gonic(),
			want: errNoVersion, mention: []string{"navidrome"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			r := newRecordRun(t, tc.spec, tc.nd, tc.gon)
			err := runRecordOracle(t.Context(), r.cfg)
			if err == nil {
				t.Fatal("recording succeeded")
			}
			if tc.status != 0 {
				if want := fmt.Sprintf("http status %d", tc.status); !strings.Contains(err.Error(), want) {
					t.Fatalf("error = %v, want one naming %q", err, want)
				}
			} else if !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
			for _, m := range tc.mention {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("error %q does not mention %q", err, m)
				}
			}
			if _, statErr := os.Stat(r.out); !errors.Is(statErr, os.ErrNotExist) {
				t.Errorf("session directory exists after a refused run (stat: %v)", statErr)
			}
			if !*r.stopped {
				t.Error("oracles not stopped after a refused run")
			}
		})
	}
}

// TestRecordOracleRefusesBodyCredential has one oracle echo a credential
// in a URL its answer carries, and expects a refusal that writes nothing.
func TestRecordOracleRefusesBodyCredential(t *testing.T) {
	leaking := nd()
	leaking.payload = func(endpoint, payload string) string {
		if endpoint == "getArtists" {
			return payload + `,"coverUrl":"http://oracle/cover?size=300&t=tokenvalue&s=saltvalue"`
		}
		return payload
	}
	r := newRecordRun(t, `{"requests":[{"endpoint":"getArtists","format":"json"}]}`, leaking, gonic())
	err := runRecordOracle(t.Context(), r.cfg)
	if err == nil {
		t.Fatal("recording succeeded")
	}
	for _, want := range []string{"credential", "navidrome", "getArtists", "subsonic-response/coverUrl", "parameter t, s"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err, want)
		}
	}
	for _, secret := range []string{"tokenvalue", "saltvalue"} {
		if strings.Contains(err.Error(), secret) {
			t.Errorf("error repeats the credential %q: %v", secret, err)
		}
	}
	// Write refuses before it creates the directory, so the output is
	// absent; an empty directory would hold nothing either.
	if entries, readErr := os.ReadDir(r.out); readErr == nil && len(entries) > 0 {
		t.Errorf("refused recording left %d files in %s", len(entries), r.out)
	}
	if !*r.stopped {
		t.Error("oracles not stopped after a refused run")
	}
}

func TestRecordOracleSpecStates(t *testing.T) {
	tests := []struct {
		name    string
		spec    *string // nil: no file
		mention []string
	}{
		{"absent", nil, []string{"read session spec"}},
		{"empty file", new(""), []string{"bad session spec"}},
		{"no requests", new(`{"requests":[]}`), []string{"lists no requests"}},
		{"malformed", new(`{"requests":`), []string{"bad session spec"}},
		{"unknown member", new(`{"requests":[{"endpoint":"ping","format":"json","extra":1}]}`), []string{"bad session spec"}},
		{"auth parameter", new(`{"requests":[{"endpoint":"ping","format":"json","params":{"t":"abc"}}]}`), []string{`parameter "t": auth parameter`}},
		{"many bad", new(`{"requests":[
			{"endpoint":"ping.view","format":"json"},
			{"endpoint":"ping","format":"yaml"},
			{"endpoint":"getSong","format":"json","params":{"apiKey":"k","p":"x"}},
			{"endpoint":"ping","format":"json"},
			{"endpoint":"ping","format":"json"}]}`), []string{
			"request 1 (ping.view): endpoint is not in canonical form", `request 2 (ping): unknown format "yaml"`,
			`request 3 (getSong): parameter "apiKey"`, `request 3 (getSong): parameter "p"`, "request 5 (ping): duplicate",
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "spec.json")
			if tc.spec != nil {
				if err := os.WriteFile(path, []byte(*tc.spec), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			_, err := loadRecordSpec(path)
			if err == nil {
				t.Fatal("loaded")
			}
			for _, m := range tc.mention {
				if !strings.Contains(err.Error(), m) {
					t.Errorf("error %q does not mention %q", err, m)
				}
			}
		})
	}
}

func TestRecordOracleRefusesNonEmptyOut(t *testing.T) {
	r := newRecordRun(t, `{"requests":[{"endpoint":"ping","format":"json"}]}`, nd(), gonic())
	if err := os.MkdirAll(r.out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.out, "keep"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := runRecordOracle(t.Context(), r.cfg)
	if !errors.Is(err, errOutExists) {
		t.Fatalf("error = %v, want errOutExists", err)
	}
	if got := len(r.nd.asked()) + len(r.gon.asked()); got != 0 {
		t.Errorf("oracles asked %d requests before the refusal", got)
	}
}

func TestRecordOracleRegistered(t *testing.T) {
	cmd, _, err := newRootCmd().Find([]string{"record-oracle"})
	if err != nil || cmd.Name() != "record-oracle" {
		t.Fatalf("Find(record-oracle) = %v, %v", cmd, err)
	}
}

const seedDir = "../../internal/harness/subsonic/data/sessions/seed"

// seedAnswer holds the members of a recorded answer that name the entity it
// describes.
type seedAnswer struct {
	Response struct {
		Artist struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"artist"`
		Album struct {
			ID     string     `json:"id"`
			Name   string     `json:"name"`
			Artist string     `json:"artist"`
			Song   []seedSong `json:"song"`
		} `json:"album"`
		Song      seedSong `json:"song"`
		Directory struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"directory"`
	} `json:"subsonic-response"`
}

// xmlEntity is the id attribute of an XML answer's entity element.
type xmlEntity struct {
	ID string `xml:"id,attr"`
}

type seedSong struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Album    string `json:"album"`
	Duration int    `json:"duration"`
}

// TestSeedSession proves the committed seed answers every request from
// both oracles and that each answer describes the entity its symbolic key
// names, on each oracle's own id.
func TestSeedSession(t *testing.T) {
	s, err := session.Load(seedDir)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"navidrome", "gonic"} {
		if _, ok := s.Sources[name]; !ok {
			t.Errorf("seed has no source %q", name)
		}
	}
	spec, err := loadRecordSpec(filepath.Join(seedDir, "..", "seed.spec.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Requests) != len(spec) {
		t.Errorf("seed holds %d requests, spec lists %d", len(s.Requests), len(spec))
	}

	type cell struct {
		endpoint string
		format   client.Format
	}
	covered := map[cell]bool{}
	xmlChecked := map[string]int{}
	songs := 0
	for i, req := range s.Requests {
		covered[cell{req.Endpoint, req.Format}] = true
		if len(req.Responses) != 2 {
			t.Errorf("request %d (%s) answered by %d sources, want 2", i+1, req.Key(), len(req.Responses))
		}
		var key string
		if len(req.Params) == 1 && req.Params[0].Name == "id" {
			key = req.Params[0].Value
		}
		for _, resp := range req.Responses {
			if resp.Status != http.StatusOK {
				t.Errorf("%s: %s answered HTTP %d", req.Key(), resp.Source, resp.Status)
			}
			if req.Format == client.XML && key != "" {
				var x struct {
					Artist    xmlEntity `xml:"artist"`
					Album     xmlEntity `xml:"album"`
					Song      xmlEntity `xml:"song"`
					Directory xmlEntity `xml:"directory"`
				}
				if err := xml.Unmarshal(resp.Body, &x); err != nil {
					t.Fatalf("%s: %s: %v", req.Key(), resp.Source, err)
				}
				got := map[string]string{
					"getArtist": x.Artist.ID, "getAlbum": x.Album.ID, "getSong": x.Song.ID, "getMusicDirectory": x.Directory.ID,
				}
				id, ok := got[req.Endpoint]
				if !ok {
					t.Fatalf("%s: no xml identity check for %s", req.Key(), req.Endpoint)
				}
				xmlChecked[req.Endpoint]++
				if want := s.Resolutions[key][resp.Source]; id != want {
					t.Errorf("%s: %s xml %s id = %q, want %q", req.Key(), resp.Source, req.Endpoint, id, want)
				}
			}
			if req.Format != client.JSON {
				continue
			}
			var a seedAnswer
			if err := json.Unmarshal(resp.Body, &a); err != nil {
				t.Fatalf("%s: %s: %v", req.Key(), resp.Source, err)
			}
			id := s.Resolutions[key][resp.Source]
			r := a.Response
			who := req.Key() + " on " + resp.Source
			switch req.Endpoint {
			case "getArtist":
				name := strings.TrimPrefix(key, "artist:")
				if r.Artist.ID != id || r.Artist.Name != name {
					t.Errorf("%s: answered artist %q (%s), want %q (%s)", who, r.Artist.Name, r.Artist.ID, name, id)
				}
			case "getAlbum":
				artist, title, _ := strings.Cut(strings.TrimPrefix(key, "album:"), "/")
				if r.Album.ID != id || r.Album.Name != title || r.Album.Artist != artist {
					t.Errorf("%s: answered album %q by %q (%s), want %q by %q (%s)", who, r.Album.Name, r.Album.Artist, r.Album.ID, title, artist, id)
				}
				for _, sg := range r.Album.Song {
					songs++
					if sg.Duration <= 0 {
						t.Errorf("%s: song %q has duration %d, so the oracle did not decode the audio", who, sg.Title, sg.Duration)
					}
				}
			case "getSong":
				parts := strings.SplitN(strings.TrimPrefix(key, "song:"), "/", 3)
				if r.Song.ID != id || r.Song.Title != parts[2] || r.Song.Album != parts[1] {
					t.Errorf("%s: answered song %q on %q (%s), want %q on %q (%s)", who, r.Song.Title, r.Song.Album, r.Song.ID, parts[2], parts[1], id)
				}
				songs++
				if r.Song.Duration <= 0 {
					t.Errorf("%s: song %q has duration %d, so the oracle did not decode the audio", who, r.Song.Title, r.Song.Duration)
				}
			case "getMusicDirectory":
				if r.Directory.ID != id || r.Directory.Name != path.Base(strings.TrimPrefix(key, "folder:")) {
					t.Errorf("%s: answered directory %q (%s), want name %q (%s)", who, r.Directory.Name, r.Directory.ID, path.Base(strings.TrimPrefix(key, "folder:")), id)
				}
			}
		}
	}
	if songs == 0 {
		t.Error("seed holds no song answers to check")
	}
	for _, ep := range []string{"getArtist", "getAlbum", "getSong", "getMusicDirectory"} {
		if xmlChecked[ep] == 0 {
			t.Errorf("seed holds no xml %s answer to check", ep)
		}
	}
	endpoints := []string{
		"ping", "getLicense", "getOpenSubsonicExtensions", "getMusicFolders", "getIndexes",
		"getMusicDirectory", "getArtists", "getArtist", "getAlbum", "getSong",
	}
	for _, ep := range endpoints {
		for _, f := range []client.Format{client.JSON, client.XML} {
			if !covered[cell{ep, f}] {
				t.Errorf("seed lacks %s in %s", ep, f)
			}
		}
	}
	for _, key := range []string{
		"album:Alpha Artist/Greatest Hits", "album:Delta Duo/Greatest Hits", // the same-title pair
		"album:Various Artists/Mixtape Volume One", // the compilation
		"album:Carmine Ensemble/Two Sides",         // a multi-disc album
	} {
		for _, src := range []string{"navidrome", "gonic"} {
			if s.Resolutions[key][src] == "" {
				t.Errorf("seed does not resolve %q on %s", key, src)
			}
		}
	}
	if a, b := s.Resolutions["album:Alpha Artist/Greatest Hits"], s.Resolutions["album:Delta Duo/Greatest Hits"]; a["navidrome"] == b["navidrome"] || a["gonic"] == b["gonic"] {
		t.Errorf("the same-title pair shares an id: %v / %v", a, b)
	}
}

func TestRecordOracleSeedDigestSurvivesTheCollectionList(t *testing.T) {
	r := newRecordRun(t, `{"requests":[{"endpoint":"ping","format":"json"}]}`, nd(), gonic())
	if err := runRecordOracle(t.Context(), r.cfg); err != nil {
		t.Fatal(err)
	}
	s, err := session.Load(r.out)
	if err != nil {
		t.Fatal(err)
	}
	// The digest the committed seed session records.
	const seedDigest = "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	for name, src := range s.Sources {
		if src.LibraryDigest != seedDigest {
			t.Errorf("%s library digest = %s, want the seed's %s", name, src.LibraryDigest, seedDigest)
		}
	}
	if len(r.started) != 1 || r.started[0].Name != "music" || r.started[0].Tracks != 29 {
		t.Errorf("collections started = %+v, want one named music with 29 tracks", r.started)
	}
}

func TestRecordOracleTwoCollections(t *testing.T) {
	r := newRecordRun(t, `{"requests":[{"endpoint":"ping","format":"json"}]}`, nd(), gonic())
	r.cfg.collections = []collection{
		{name: "first", manifest: fixtureManifest("collection-a.json")},
		{name: "second", manifest: fixtureManifest("collection-b.json")},
	}
	if err := runRecordOracle(t.Context(), r.cfg); err != nil {
		t.Fatal(err)
	}
	if len(r.started) != 2 || r.started[0].Name != "first" || r.started[1].Name != "second" ||
		r.started[0].Tracks != 11 || r.started[1].Tracks != 11 || r.started[0].Dir == r.started[1].Dir {
		t.Errorf("collections started = %+v, want first and second, 11 tracks each, in distinct directories", r.started)
	}
	s, err := session.Load(r.out)
	if err != nil {
		t.Fatal(err)
	}
	const seedDigest = "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	for name, src := range s.Sources {
		if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(src.LibraryDigest) || src.LibraryDigest == seedDigest {
			t.Errorf("%s library digest = %q, want a combined 64-hex digest", name, src.LibraryDigest)
		}
	}
}
