package grade

import (
	"bytes"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

func manifestTracks(t *testing.T) []fixture.Track {
	t.Helper()
	lib, err := fixture.Load(committedManifest)
	if err != nil {
		t.Fatal(err)
	}
	return lib.Tracks()
}

func committedAssign(t *testing.T) *assignment {
	t.Helper()
	a, err := loadAssignment(committedAssignment, manifestTracks(t))
	if err != nil {
		t.Fatal(err)
	}
	return a
}

// pinnedPaths returns the pinned OpenAPI document's paths with the
// "/rest/" prefix cut.
func pinnedPaths(t *testing.T) []string {
	t.Helper()
	var doc struct {
		Paths map[string]jsontext.Value `json:"paths"`
	}
	if err := json.Unmarshal(mustRead(t, filepath.Join(openapiDir, openapiFile)), &doc); err != nil {
		t.Fatal(err)
	}
	var out []string
	for p := range doc.Paths {
		out = append(out, strings.TrimPrefix(p, "/rest/"))
	}
	slices.Sort(out)
	return out
}

func TestAssignmentCoversThePinnedDocument(t *testing.T) {
	a := committedAssign(t)
	var got []string
	for _, e := range a.Endpoints {
		got = append(got, e.Endpoint)
	}
	slices.Sort(got)
	want := pinnedPaths(t)
	if !slices.Equal(got, want) {
		t.Errorf("assignment lacks %v and lists %v beyond the document",
			slices.DeleteFunc(slices.Clone(want), func(p string) bool { return slices.Contains(got, p) }),
			slices.DeleteFunc(slices.Clone(got), func(p string) bool { return slices.Contains(want, p) }))
	}
	if len(want) != 87 {
		t.Errorf("the pinned document has %d paths, want 87", len(want))
	}
}

// verdict is one transcribed assignment: the endpoint, the oracle that
// grades it, and the slice that first serves it.
type verdict struct {
	endpoint string
	oracle   oracle
	slice    string
}

// servedVerdicts holds the 49 served endpoints the study record's oracle
// assignment names, followed by the three empty-success additions, and
// then the two the record names that the launch does not serve. A split
// verdict takes its narrow value.
var servedVerdicts = []verdict{
	// System.
	{"ping", oracleBoth, "slice 3"},
	{"getLicense", oracleBoth, "slice 3"},
	{"getOpenSubsonicExtensions", oracleContractOnly, "slice 3"},
	{"getScanStatus", oracleContractOnly, "slice 3"},
	{"startScan", oracleContractOnly, "slice 3"},
	// Browsing, ID3.
	{"getArtists", oracleBoth, "slice 3"},
	{"getArtist", oracleBoth, "slice 3"},
	{"getAlbum", oracleBoth, "slice 3"},
	{"getSong", oracleBoth, "slice 3"},
	{"getGenres", oracleBoth, "slice 3"},
	{"getAlbumInfo2", oracleBoth, "slice 3"},
	{"getAlbumInfo", oracleBoth, "slice 3"},
	{"getMusicFolders", oracleGonic, "slice 3"},
	// Browsing, folder.
	{"getIndexes", oracleContractOnly, "slice 3"},
	{"getMusicDirectory", oracleContractOnly, "slice 3"},
	// Info family.
	{"getArtistInfo", oracleBoth, "slice 3"},
	{"getArtistInfo2", oracleBoth, "slice 3"},
	{"getTopSongs", oracleBoth, "slice 3"},
	{"getSimilarSongs", oracleBoth, "slice 3"},
	{"getSimilarSongs2", oracleBoth, "slice 3"},
	// Album lists.
	{"getAlbumList2", oracleBoth, "slice 3"},
	{"getAlbumList", oracleBoth, "slice 3"},
	{"getRandomSongs", oracleBoth, "slice 3"},
	{"getSongsByGenre", oracleBoth, "slice 3"},
	{"getStarred2", oracleBoth, "slice 7"},
	{"getStarred", oracleBoth, "slice 7"},
	{"getNowPlaying", oracleNavidrome, "slice 7"},
	// Search.
	{"search3", oracleBoth, "slice 4"},
	{"search2", oracleBoth, "slice 4"},
	// Media.
	{"stream", oracleNavidrome, "slice 5"},
	{"download", oracleContractOnly, "slice 5"},
	{"getCoverArt", oracleBoth, "slice 4"},
	{"getLyrics", oracleContractOnly, "slice 3"},
	// Annotation.
	{"star", oracleBoth, "slice 7"},
	{"unstar", oracleBoth, "slice 7"},
	{"scrobble", oracleNavidrome, "slice 7"},
	// Playlists.
	{"getPlaylists", oracleBoth, "slice 7"},
	{"getPlaylist", oracleBoth, "slice 7"},
	{"createPlaylist", oracleBoth, "slice 7"},
	{"updatePlaylist", oracleBoth, "slice 7"},
	{"deletePlaylist", oracleBoth, "slice 7"},
	// Play queue.
	{"getPlayQueue", oracleContractOnly, "slice 7"},
	{"savePlayQueue", oracleContractOnly, "slice 7"},
	// Shares.
	{"getShares", oracleNavidrome, "slice 6"},
	{"createShare", oracleNavidrome, "slice 6"},
	{"updateShare", oracleNavidrome, "slice 6"},
	{"deleteShare", oracleNavidrome, "slice 6"},
	// User and extension endpoints.
	{"getUser", oracleNavidrome, "slice 6"},
	{"tokenInfo", oracleContractOnly, "slice 6"},
	// The empty-success additions.
	{"getInternetRadioStations", oracleContractOnly, "slice 3"},
	{"getPodcasts", oracleContractOnly, "slice 3"},
	{"getNewestPodcasts", oracleContractOnly, "slice 3"},
}

// notServedVerdicts are the two endpoints the record names that answer
// error 0 at launch.
var notServedVerdicts = []verdict{
	{"getTranscodeDecision", oracleContractOnly, notServed},
	{"getTranscodeStream", oracleContractOnly, notServed},
}

// contractOnlyNamed lists every served endpoint allowed to be contract
// only, by the reason it is: a new contract-only value needs an edit here.
var contractOnlyNamed = map[string][]string{
	"the record's contract-only rows": {"getScanStatus", "startScan", "getLyrics", "download", "tokenInfo"},
	"the folder-browsing ruling":      {"getIndexes", "getMusicDirectory"},
	"a split verdict narrowed":        {"getOpenSubsonicExtensions", "getPlayQueue", "savePlayQueue"},
	"the empty-success additions":     {"getInternetRadioStations", "getPodcasts", "getNewestPodcasts"},
}

func TestAssignmentValues(t *testing.T) {
	a := committedAssign(t)
	if n := len(servedVerdicts); n != 52 {
		t.Fatalf("the served table has %d rows, want 52", n)
	}
	want := map[string]verdict{}
	for _, v := range slices.Concat(servedVerdicts, notServedVerdicts) {
		if _, dup := want[v.endpoint]; dup {
			t.Fatalf("%s is transcribed twice", v.endpoint)
		}
		want[v.endpoint] = v
	}
	for _, v := range want {
		e, ok := a.lookup(v.endpoint)
		switch {
		case !ok:
			t.Errorf("%s: no entry", v.endpoint)
		case e.Oracle != v.oracle || e.Slice != v.slice:
			t.Errorf("%s: oracle %q slice %q, want %q and %q", v.endpoint, e.Oracle, e.Slice, v.oracle, v.slice)
		}
	}

	allowed := map[string]bool{}
	for _, names := range contractOnlyNamed {
		for _, n := range names {
			allowed[n] = true
		}
	}
	served := map[string]bool{}
	for _, v := range servedVerdicts {
		served[v.endpoint] = true
	}
	var notServedGot []string
	for _, e := range a.Endpoints {
		if e.Slice == notServed {
			notServedGot = append(notServedGot, e.Endpoint)
		}
		if e.Oracle == oracleContractOnly && e.Slice != notServed && !allowed[e.Endpoint] {
			t.Errorf("%s is contract only and on no named list", e.Endpoint)
		}
	}
	slices.Sort(notServedGot)
	wantNotServed := slices.DeleteFunc(pinnedPaths(t), func(p string) bool { return served[p] })
	if len(wantNotServed) != 35 || !slices.Equal(notServedGot, wantNotServed) {
		t.Errorf("not served = %d endpoints %v; want the %d paths outside the served set", len(notServedGot), notServedGot, len(wantNotServed))
	}
}

func TestAssignmentAreaList(t *testing.T) {
	a := committedAssign(t)
	areas := map[string]string{}
	for _, ar := range a.Areas {
		areas[ar.Name] = ar.Slice
	}
	wantAreas := map[string]string{
		"multi-user isolation":          "slice 6",
		"per-endpoint folder filtering": "slice 3",
		"permission refusals":           "slice 6",
	}
	if !maps.Equal(areas, wantAreas) {
		t.Errorf("areas = %v, want %v", areas, wantAreas)
	}
	shapesGot := map[string]string{}
	for _, s := range a.AlbumShapes {
		shapesGot[s.Shape] = s.Album
	}
	wantShapes := map[string]string{
		"compilation":          "Various Artists/Mixtape Volume One",
		"multi-disc":           "Borealis Trio/Double Feature",
		"collaboration credit": "Alpha Artist/Single Disc",
		"unicode":              "Zoë Ångström/Café Années",
	}
	if !maps.Equal(shapesGot, wantShapes) {
		t.Errorf("album shapes = %v, want %v", shapesGot, wantShapes)
	}
	if len(a.NotApplicable) != 1 || a.NotApplicable[0].Shape != "empty album" || a.NotApplicable[0].Reason == "" {
		t.Errorf("not applicable = %+v, want the empty album with a reason", a.NotApplicable)
	}
}

// validAssignment is a minimal assignment the loader accepts.
func validAssignment() assignment {
	return assignment{
		Endpoints: []endpointEntry{
			{Endpoint: "ping", Oracle: oracleBoth, Slice: "slice 3", Reason: "r"},
			{Endpoint: "jukeboxControl", Oracle: oracleContractOnly, Slice: notServed, Reason: "r"},
		},
		Areas:       []areaEntry{{Name: "a", Slice: "slice 3", Reason: "r"}},
		Agreements:  []agreementEntry{{ID: "A-01", Area: "a", Behavior: "b", Evidence: "e", Test: "t", Slice: "slice 3"}},
		AlbumShapes: []shapeEntry{{Shape: "compilation", Album: "Various Artists/Mixtape Volume One", Reason: "r"}},
	}
}

func TestLoadAssignmentStates(t *testing.T) {
	tests := []struct {
		name string
		// file is the assignment file's bytes, or the result of edit on
		// the valid assignment; nil file and nil edit leave it absent.
		file    []byte
		edit    func(*assignment)
		wantErr []string // substrings of the error; none means success
		wantIs  error
	}{
		{name: "valid", edit: func(*assignment) {}},
		{name: "absent", wantErr: []string{"read assignment"}, wantIs: fs.ErrNotExist},
		{name: "empty file", file: []byte(`{}`), wantErr: []string{"no endpoints"}, wantIs: errInvalidAssignment},
		{name: "malformed", file: []byte(`{"endpoints": [`), wantErr: []string{"decode assignment", "assignment.json"}},
		{name: "unknown member", file: []byte(`{"endpoints":[],"extra":1}`), wantErr: []string{"decode assignment", "extra"}},
		{
			name:    "endpoint twice",
			edit:    func(a *assignment) { a.Endpoints = append(a.Endpoints, a.Endpoints[0]) },
			wantErr: []string{`endpoint "ping": listed twice`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "unknown oracle",
			edit:    func(a *assignment) { a.Endpoints[0].Oracle = "subsonic" },
			wantErr: []string{`unknown oracle "subsonic"`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "empty slice",
			edit:    func(a *assignment) { a.Endpoints[0].Slice = "" },
			wantErr: []string{`endpoint "ping": empty slice`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "empty reason",
			edit:    func(a *assignment) { a.Endpoints[0].Reason = "" },
			wantErr: []string{`endpoint "ping": empty reason`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "not served with an oracle",
			edit:    func(a *assignment) { a.Endpoints[1].Oracle = oracleNavidrome },
			wantErr: []string{`endpoint "jukeboxControl": not served at launch`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "shape names no manifest album",
			edit:    func(a *assignment) { a.AlbumShapes[0].Album = "Nobody/Nothing" },
			wantErr: []string{`album "Nobody/Nothing" is not in the manifest`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "shape the album lacks",
			edit:    func(a *assignment) { a.AlbumShapes[0].Album = "Alpha Artist/Second Light" },
			wantErr: []string{`album "Alpha Artist/Second Light" lacks the shape`}, wantIs: errInvalidAssignment,
		},
		{
			name: "multi-disc shape the album lacks",
			edit: func(a *assignment) {
				a.AlbumShapes[0].Shape = "multi-disc"
				a.AlbumShapes[0].Album = "Alpha Artist/Single Disc"
			},
			wantErr: []string{`album "Alpha Artist/Single Disc" lacks the shape`}, wantIs: errInvalidAssignment,
		},
		{
			name: "collaboration credit shape the album lacks",
			edit: func(a *assignment) {
				a.AlbumShapes[0].Shape = "collaboration credit"
				a.AlbumShapes[0].Album = "Alpha Artist/Second Light"
			},
			wantErr: []string{`album "Alpha Artist/Second Light" lacks the shape`}, wantIs: errInvalidAssignment,
		},
		{
			name: "unicode shape the album lacks",
			edit: func(a *assignment) {
				a.AlbumShapes[0].Shape = "unicode"
				a.AlbumShapes[0].Album = "Alpha Artist/Second Light"
			},
			wantErr: []string{`album "Alpha Artist/Second Light" lacks the shape`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "unknown shape",
			edit:    func(a *assignment) { a.AlbumShapes[0].Shape = "concept album" },
			wantErr: []string{`unknown shape`}, wantIs: errInvalidAssignment,
		},
		{
			name:    "agreement twice",
			edit:    func(a *assignment) { a.Agreements = append(a.Agreements, a.Agreements[0]) },
			wantErr: []string{`agreement "A-01": listed twice`}, wantIs: errInvalidAssignment,
		},
		{
			name: "agreement with every field empty",
			edit: func(a *assignment) {
				a.Agreements = append(a.Agreements, agreementEntry{ID: "A-02"})
			},
			wantErr: []string{
				`agreement "A-02": empty area`, `agreement "A-02": empty behavior`, `agreement "A-02": empty evidence`,
				`agreement "A-02": empty test`, `agreement "A-02": empty slice`,
			},
			wantIs: errInvalidAssignment,
		},
		{
			name:    "agreement with no id",
			edit:    func(a *assignment) { a.Agreements[0].ID = "" },
			wantErr: []string{`agreement 1: empty id`}, wantIs: errInvalidAssignment,
		},
		{
			name: "many bad",
			edit: func(a *assignment) {
				a.Endpoints = append(a.Endpoints, a.Endpoints[0])
				a.Endpoints[0].Oracle = "subsonic"
				a.Endpoints[1].Oracle = oracleGonic
				a.Endpoints[1].Reason = ""
				a.Areas = append(a.Areas, a.Areas[0])
				a.AlbumShapes[0].Album = "Nobody/Nothing"
			},
			wantErr: []string{
				`unknown oracle "subsonic"`, `endpoint "ping": listed twice`, `endpoint "jukeboxControl": empty reason`,
				`endpoint "jukeboxControl": not served at launch`, `area "a": listed twice`, `is not in the manifest`,
			},
			wantIs: errInvalidAssignment,
		},
	}
	tracks := manifestTracks(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "assignment.json")
			file := tt.file
			if file == nil && tt.edit != nil {
				a := validAssignment()
				tt.edit(&a)
				var err error
				if file, err = json.Marshal(a); err != nil {
					t.Fatal(err)
				}
			}
			if file != nil {
				if err := os.WriteFile(path, file, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			a, err := loadAssignment(path, tracks)
			if len(tt.wantErr) == 0 {
				if err != nil || a == nil {
					t.Fatalf("loadAssignment = %v, %v; want an assignment", a, err)
				}
				return
			}
			if err == nil {
				t.Fatal("loadAssignment succeeded")
			}
			if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
				t.Errorf("err = %v, want %v", err, tt.wantIs)
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %v, want it to contain %q", err, want)
				}
			}
		})
	}
}

// versions each source of the grading proof records. The three servers
// share one library digest.
const (
	proofDigest     = "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	dubplateVersion = "dubplate-0.8.0"
)

var serverVersionMember = regexp.MustCompile(`"serverVersion":"[^"]*"`)

// proofSession writes a session of four requests to three sources and
// returns its directory. Dubplate's getAlbum and getIndexes bodies are
// gonic's, with getIndexes missing a required member; its getMusicFolders
// body is Navidrome's. The last request goes to an endpoint the
// assignment lacks, and the sources answer it with three different
// messages. Every dubplate body names the dubplate source's own
// serverVersion, as the session loader requires.
func proofSession(t *testing.T) string {
	t.Helper()
	body := func(n, source string) []byte { return mustRead(t, filepath.Join(seedSession, n+"-"+source+".json")) }
	asDubplate := func(b []byte) []byte {
		out := serverVersionMember.ReplaceAll(b, []byte(`"serverVersion":"`+dubplateVersion+`"`))
		if string(out) == string(b) {
			t.Fatal("no serverVersion member to rewrite")
		}
		return out
	}
	indexes := body("005", "gonic")
	perturbed := regexp.MustCompile(`"ignoredArticles":"",`).ReplaceAll(indexes, nil)
	if len(perturbed) == len(indexes) {
		t.Fatal("getIndexes body has no ignoredArticles member to remove")
	}
	renamedFolder := bytes.Replace(body("004", "navidrome"), []byte(`"Music Library"`), []byte(`"Dubplate Library"`), 1)
	if bytes.Equal(renamedFolder, body("004", "navidrome")) {
		t.Fatal("getMusicFolders body has no folder name to rewrite")
	}
	probe := func(version, code, msg string) []byte {
		return []byte(`{"subsonic-response":{"status":"failed","version":"1.16.1","serverVersion":"` + version +
			`","error":{"code":` + code + `,"message":"` + msg + `"}}}`)
	}
	versions := map[string]string{"navidrome": navidromeVersion, "gonic": gonicVersion, dubplate: dubplateVersion}
	answers := []struct {
		endpoint string
		params   []session.Param
		by       map[string][]byte
	}{
		{"getAlbum", []session.Param{{Name: "id", Value: "album:Alpha Artist/Single Disc"}}, map[string][]byte{
			"navidrome": body("012", "navidrome"), "gonic": body("012", "gonic"), dubplate: asDubplate(body("012", "gonic")),
		}},
		{"getIndexes", nil, map[string][]byte{
			"navidrome": body("005", "navidrome"), "gonic": indexes, dubplate: asDubplate(perturbed),
		}},
		{"getMusicFolders", nil, map[string][]byte{
			"navidrome": body("004", "navidrome"), "gonic": body("004", "gonic"), dubplate: asDubplate(renamedFolder),
		}},
		{"noSuchEndpoint", nil, map[string][]byte{
			"navidrome": probe(navidromeVersion, "0", "n"), "gonic": probe(gonicVersion, "10", "g"), dubplate: probe(dubplateVersion, "70", "d"),
		}},
	}
	s := &session.Session{Sources: map[string]session.Source{}}
	for name, v := range versions {
		s.Sources[name] = session.Source{LibraryDigest: proofDigest, ServerVersion: v}
	}
	for _, a := range answers {
		req := session.Request{Endpoint: a.endpoint, Format: client.JSON, Params: a.params}
		for _, name := range slices.Sorted(maps.Keys(a.by)) {
			req.Responses = append(req.Responses, session.Response{Source: name, Status: 200, ContentType: "application/json", Body: a.by[name]})
		}
		s.Requests = append(s.Requests, req)
	}
	dir := filepath.Join(t.TempDir(), "proof")
	if err := session.Write(dir, s); err != nil {
		t.Fatal(err)
	}
	return dir
}

// oracleFindings counts the findings of request-endpoint kinds the oracle
// tier reports, and below counts those under the payload member, not at
// the envelope.
func oracleFindings(r *Report, endpoint string) (all, below int) {
	for _, f := range r.findings {
		if f.Endpoint != endpoint || f.Kind == KindContract {
			continue
		}
		all++
		if strings.Count(f.Path, "/") >= 2 {
			below++
		}
	}
	return all, below
}

func contractFindings(r *Report, endpoint, source string) []Finding {
	var out []Finding
	for _, f := range r.findings {
		if f.Endpoint == endpoint && f.Kind == KindContract && slices.Equal(f.Sources, []string{source}) {
			out = append(out, f)
		}
	}
	return out
}

func TestGradeOfDubplateFollowsTheAssignment(t *testing.T) {
	cfg := emptyLedger(t)
	g, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	dir := proofSession(t)

	t.Run("against navidrome", func(t *testing.T) {
		r, err := g.Grade(t.Context(), dir, [2]string{dubplate, "navidrome"})
		if err != nil {
			t.Fatal(err)
		}
		if all, below := oracleFindings(r, "getAlbum"); all == 0 || below == 0 {
			t.Errorf("getAlbum oracle findings %d, below the envelope %d; want both non-zero", all, below)
		}
		for _, endpoint := range []string{"getIndexes", "getMusicFolders", "noSuchEndpoint"} {
			if all, _ := oracleFindings(r, endpoint); all != 0 {
				t.Errorf("%s has %d oracle findings, want none", endpoint, all)
			}
		}
		got := contractFindings(r, "getIndexes", dubplate)
		if len(got) != 1 || !strings.Contains(got[0].Path, "ignoredArticles") && !strings.Contains(got[0].Values[0], "ignoredArticles") {
			t.Errorf("dubplate's getIndexes contract findings = %+v, want the one missing ignoredArticles", got)
		}
		wantSection := "not assigned to navidrome:\n  getIndexes: 1\n  getMusicFolders: 1\nunmapped:\n  noSuchEndpoint: 1\n"
		if !strings.Contains(string(r.Format()), wantSection) {
			t.Errorf("report lacks the skip sections %q:\n%s", wantSection, r.Format())
		}
	})

	t.Run("against gonic", func(t *testing.T) {
		r, err := g.Grade(t.Context(), dir, [2]string{dubplate, "gonic"})
		if err != nil {
			t.Fatal(err)
		}
		if all, below := oracleFindings(r, "getMusicFolders"); all == 0 || below == 0 {
			t.Errorf("getMusicFolders oracle findings %d, below the envelope %d; want both non-zero", all, below)
		}
		for _, endpoint := range []string{"getIndexes", "noSuchEndpoint"} {
			if all, _ := oracleFindings(r, endpoint); all != 0 {
				t.Errorf("%s has %d oracle findings, want none", endpoint, all)
			}
		}
		wantSection := "not assigned to gonic:\n  getIndexes: 1\nunmapped:\n  noSuchEndpoint: 1\n"
		if !strings.Contains(string(r.Format()), wantSection) {
			t.Errorf("report lacks the skip sections %q:\n%s", wantSection, r.Format())
		}
	})

	t.Run("a pair of oracles reports no skips", func(t *testing.T) {
		r, err := g.Grade(t.Context(), dir, pairNG)
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(r.Format()), "not assigned") || strings.Contains(string(r.Format()), "unmapped") {
			t.Errorf("an oracle pair's report names the assignment:\n%s", r.Format())
		}
		if all, _ := oracleFindings(r, "getIndexes"); all == 0 {
			t.Error("the oracle pair's getIndexes was not compared")
		}
	})
}

// TestSplitVerdictsQuoteTheStudyRecord pins each split verdict's reason
// to the wider grant the study record's per-endpoint table words. Every
// quote must appear in the record and in the reason, so a paraphrase on
// either side fails.
func TestSplitVerdictsQuoteTheStudyRecord(t *testing.T) {
	const record = "../../../../docs/research/2026-10-06-slice-1-study-record.md"
	table := string(mustRead(t, record))
	section3 := table[strings.Index(table, "## 3. The per-endpoint oracle assignment"):]
	section3 = section3[:strings.Index(section3, "## 4.")]

	const shares = "Navidrome for shape; the spec only for the six bounds"
	const playQueue = "Navidrome for the empty answer; contract and the spec's edges for the rest"
	const folderBrowsing = "gonic if fork C is yes; contract only if no"
	tests := []struct {
		endpoint string
		quotes   []string
	}{
		{"stream", []string{"`stream` (direct)", "`stream` (transcode)"}},
		{"getOpenSubsonicExtensions", []string{"Navidrome for the carve-out; contract only for the list"}},
		{"getPlayQueue", []string{playQueue}},
		{"savePlayQueue", []string{playQueue}},
		{"getShares", []string{shares}},
		{"createShare", []string{shares}},
		{"updateShare", []string{shares}},
		{"deleteShare", []string{shares}},
		{"getMusicFolders", []string{"gonic (spec `:278`), shape only"}},
		{"getIndexes", []string{folderBrowsing}},
		{"getMusicDirectory", []string{folderBrowsing}},
	}
	a := committedAssign(t)
	reasons := make(map[string]string)
	for _, e := range a.Endpoints {
		reasons[e.Endpoint] = e.Reason
	}
	for _, tc := range tests {
		t.Run(tc.endpoint, func(t *testing.T) {
			reason, ok := reasons[tc.endpoint]
			if !ok {
				t.Fatalf("%s is not in the assignment", tc.endpoint)
			}
			for _, q := range tc.quotes {
				if !strings.Contains(section3, q) {
					t.Errorf("section 3 lacks %q", q)
				}
				if !strings.Contains(reason, q) {
					t.Errorf("%s's reason lacks section 3's words %q:\n%s", tc.endpoint, q, reason)
				}
			}
		})
	}
}
