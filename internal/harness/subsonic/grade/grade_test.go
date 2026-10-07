package grade

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"flag"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/atomicfile"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

var update = flag.Bool("update", false, "rewrite the seed grade's golden report")

const (
	seedLedger = "testdata/seed-ledger.json"
	seedGolden = "testdata/seed-report.golden"
	sessionsIn = "../data/sessions"
)

func seedConfig(ledger string) Config {
	return Config{Ledger: ledger, Sessions: sessionsIn, OpenAPI: openapiDir, Manifest: committedManifest}
}

func hasFinding(r *Report, request string, kind Kind, path, row string) bool {
	for i, f := range r.Findings {
		if f.Request == request && f.Kind == kind && f.Path == path && r.ledger.Matches[i] == row {
			return true
		}
	}
	return false
}

// TestSeedGrade grades the committed seed session, Navidrome against
// gonic, under the test-only ledger and pins the whole report.
func TestSeedGrade(t *testing.T) {
	g, err := New(seedConfig(seedLedger))
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(context.Background(), seedSession, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	got := r.Format()
	if *update {
		if err := atomicfile.WriteFile(seedGolden, got, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(seedGolden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("report differs from %s; rerun with -update and review the diff", seedGolden)
	}
	// The totals sit here as well, so an -update that moves them needs a
	// reviewed edit to this test.
	if len(r.Findings) != 1686 || r.Unledgered() != 893 {
		t.Errorf("findings %d, unledgered %d; want 1686 and 893", len(r.Findings), r.Unledgered())
	}
	byKind := map[Kind]int{}
	for _, f := range r.Findings {
		byKind[f.Kind]++
	}
	wantKinds := map[Kind]int{KindContract: 51, KindIdentity: 13, KindPresence: 1218, KindType: 35, KindValue: 369}
	if !maps.Equal(byKind, wantKinds) {
		t.Errorf("findings by kind %v, want %v", byKind, wantKinds)
	}
	again, err := g.Grade(context.Background(), seedSession, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(again.Format(), got) {
		t.Error("a second grade of the same session formats differently")
	}

	// Navidrome serves the compilation's folder from a path it builds
	// from tags, where gonic serves the file's own path.
	compilation := "getMusicDirectory.json&id=folder%3AVarious+Artists%2FMixtape+Volume+One"
	if !hasFinding(r, compilation, KindValue, "subsonic-response/directory/child/0/path", "folder-view-getmusicdirectory") {
		t.Error("no by-construction path finding in the compilation's folder")
	}
	contract := []struct{ request, path, row string }{
		{"getAlbum.json&id=album%3AAlpha+Artist%2FSingle+Disc", "subsonic-response/album/userRating", "navidrome-album-userrating"},
		{"getAlbum.json&id=album%3AAlpha+Artist%2FSingle+Disc", "subsonic-response/album/song/2/replayGain", "gonic-song-replaygain"},
	}
	for _, c := range contract {
		if !hasFinding(r, c.request, KindContract, c.path, c.row) {
			t.Errorf("no contract finding at %s in %s matched by %s", c.path, c.request, c.row)
		}
	}
	l, err := LoadLedger(seedLedger, sessionsIn)
	if err != nil {
		t.Fatal(err)
	}
	for _, row := range l.Rows {
		if r.ledger.Counts[row.ID] < 1 {
			t.Errorf("row %s matched nothing", row.ID)
		}
	}
}

// handSession writes a session of one getAlbum request whose two sources
// answered with the given bodies, under the given library digests, and
// returns its directory.
func handSession(t *testing.T, name string, bodies [2][]byte, digests [2]string) string {
	t.Helper()
	return handSessionAt(t, name, "getAlbum", 200, bodies, digests)
}

// handSessionAt is handSession for a request to endpoint that both
// sources answered with status.
func handSessionAt(t *testing.T, name, endpoint string, status int, bodies [2][]byte, digests [2]string) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), name)
	const version = "0.64.2 (10114574)"
	s := &session.Session{
		Sources: map[string]session.Source{
			"navidrome": {LibraryDigest: digests[0], ServerVersion: version},
			"copy":      {LibraryDigest: digests[1], ServerVersion: version},
		},
		Requests: []session.Request{{
			Endpoint: endpoint, Format: client.JSON,
			Params: []session.Param{{Name: "id", Value: "album:Alpha Artist/Single Disc"}},
			Responses: []session.Response{
				{Source: "navidrome", Status: status, ContentType: "application/json", Body: bodies[0]},
				{Source: "copy", Status: status, ContentType: "application/json", Body: bodies[1]},
			},
		}},
	}
	if err := session.Write(dir, s); err != nil {
		t.Fatal(err)
	}
	return dir
}

func emptyLedger(t *testing.T) Config {
	t.Helper()
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := os.WriteFile(path, []byte("[]"), 0o644); err != nil {
		t.Fatal(err)
	}
	return seedConfig(path)
}

// TestGradeIdenticalBodiesReportOnlyContract grades two sources that gave
// the same bytes, so the oracle tier is clean and the contract tier must
// still run on each.
func TestGradeIdenticalBodiesReportOnlyContract(t *testing.T) {
	body := readBody(t, albumJSON)
	digest := "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	dir := handSession(t, "same", [2][]byte{body, body}, [2]string{digest, digest})
	g, err := New(emptyLedger(t))
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(context.Background(), dir, [2]string{"navidrome", "copy"})
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, f := range r.Findings {
		got = append(got, string(f.Kind)+" "+strings.Join(f.Sources, ",")+" "+f.Path+" "+strings.Join(f.Values, ","))
	}
	want := []string{
		"contract copy subsonic-response/album/userRating 0: number must be at least 1",
		"contract navidrome subsonic-response/album/userRating 0: number must be at least 1",
	}
	if !slices.Equal(got, want) {
		t.Errorf("findings:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

// TestGradeReportsContractOutcomesWithoutViolations grades identical
// bodies that the contract tier cannot validate: each source's answer
// is one finding at the envelope, whose value is the outcome.
func TestGradeReportsContractOutcomesWithoutViolations(t *testing.T) {
	body := readBody(t, albumJSON)
	digest := "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	tests := []struct {
		name     string
		endpoint string
		status   int
		want     string
	}{
		{"no route", "getNothing", 200, "no-route"},
		{"undeclared status", "getAlbum", 500, "undeclared-status"},
		{"no schema", "getAlbum", 304, "no-schema"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := handSessionAt(t, "outcome", tt.endpoint, tt.status, [2][]byte{body, body}, [2]string{digest, digest})
			g, err := New(emptyLedger(t))
			if err != nil {
				t.Fatal(err)
			}
			r, err := g.Grade(context.Background(), dir, [2]string{"navidrome", "copy"})
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, f := range r.Findings {
				got = append(got, string(f.Kind)+" "+strings.Join(f.Sources, ",")+" "+f.Path+" "+strings.Join(f.Values, ","))
			}
			want := []string{
				"contract copy subsonic-response " + tt.want,
				"contract navidrome subsonic-response " + tt.want,
			}
			if !slices.Equal(got, want) {
				t.Errorf("findings:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
			}
		})
	}
}

// TestGradeCountsOwedRowsBySlice grades a session under a ledger holding
// one owed row, which the report lists under its slice.
func TestGradeCountsOwedRowsBySlice(t *testing.T) {
	body := readBody(t, albumJSON)
	digest := "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	dir := handSession(t, "same", [2][]byte{body, body}, [2]string{digest, digest})
	ledger, _ := writeLedger(t, []Row{owedRow("owed", "slice 3")})
	g, err := New(seedConfig(ledger))
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(context.Background(), dir, [2]string{"navidrome", "copy"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(r.Format()); !strings.Contains(got, "owed rows by slice:\n  slice 3: 1\n") {
		t.Errorf("report does not list the owed row under its slice:\n%s", got)
	}
}

// TestGradeRefusesAVersionMismatch grades the seed under a row adjudicated
// against another gonic version than the one the session recorded.
func TestGradeRefusesAVersionMismatch(t *testing.T) {
	row := activeRow("old-gonic", ClassVolatile)
	row.Session = "seed"
	row.Pattern = Pattern{Endpoint: "ping", Format: "any", Path: "subsonic-response/type"}
	row.Versions["gonic"] = "0.21.0"
	ledger, _ := writeLedger(t, []Row{row})
	g, err := New(seedConfig(ledger))
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(context.Background(), seedSession, pairNG)
	if !errors.Is(err, ErrVersionMismatch) || r != nil {
		t.Fatalf("Grade = %v, %v; want a version refusal and no report", r, err)
	}
}

func TestGradeRefusesMismatchedLibraries(t *testing.T) {
	body := readBody(t, albumJSON)
	dir := handSession(t, "split", [2][]byte{body, body}, [2]string{"digest-one", "digest-two"})
	g, err := New(emptyLedger(t))
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(context.Background(), dir, [2]string{"navidrome", "copy"})
	if !errors.Is(err, errLibraryMismatch) || r != nil {
		t.Fatalf("Grade = %v, %v; want a refusal", r, err)
	}
	for _, d := range []string{"digest-one", "digest-two"} {
		if !strings.Contains(err.Error(), d) {
			t.Errorf("refusal %q does not name %s", err, d)
		}
	}
}

// TestOracleTierStates grades in-memory sessions the session format
// itself would refuse to write, since the grader must still report them.
func TestOracleTierStates(t *testing.T) {
	body := readBody(t, albumJSON)
	broken := []byte(`<subsonic-response><a></b></subsonic-response>`)
	tests := []struct {
		name   string
		format client.Format
		bodies [][]byte // nil drops the second response
		values []string
		kind   Kind
	}{
		{"one source answered", client.JSON, [][]byte{body}, []string{"status 200", "no response"}, KindMissingSource},
		{"empty body", client.JSON, [][]byte{body, nil}, []string{"parsed", "json: empty body"}, KindParse},
		{"body does not parse", client.XML, [][]byte{broken, broken}, nil, KindParse},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := session.Request{Endpoint: "getAlbum", Format: tt.format}
			for i, b := range tt.bodies {
				req.Responses = append(req.Responses, session.Response{
					Source: []string{"left", "right"}[i], Status: 200, ContentType: "application/json", Body: b,
				})
			}
			g := &Grader{lib: testLibrary(t)}
			found := g.oracleTier(&session.Session{Requests: []session.Request{req}}, [2]string{"left", "right"}, nil)
			if len(found) != 1 || found[0].Kind != tt.kind || found[0].Path != envelopePath {
				t.Fatalf("findings %+v, want one %s", found, tt.kind)
			}
			if tt.values != nil && !slices.Equal(found[0].Values, tt.values) {
				t.Errorf("values %q, want %q", found[0].Values, tt.values)
			}
			for _, v := range found[0].Values {
				if tt.kind == KindParse && tt.values == nil && !strings.HasPrefix(v, "xml: ") {
					t.Errorf("parse value %q does not name the format", v)
				}
			}
		})
	}
}

func TestGradeRefusesAnUnknownSource(t *testing.T) {
	body := readBody(t, albumJSON)
	dir := handSession(t, "one", [2][]byte{body, body}, [2]string{"d", "d"})
	g, err := New(emptyLedger(t))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Grade(context.Background(), dir, pairNG); err == nil || !strings.Contains(err.Error(), `no source "gonic"`) {
		t.Errorf("err = %v", err)
	}
}

// scopeSession writes a session holding the seed's JSON answers to
// getArtist for Alpha Artist and getMusicDirectory for the Alpha Artist
// folder, in that order. Navidrome gives that artist and that folder one
// id; gonic gives them two.
func scopeSession(t *testing.T) string {
	t.Helper()
	seed, err := session.Load(seedSession)
	if err != nil {
		t.Fatal(err)
	}
	keep := []string{"getArtist.json&id=artist%3AAlpha+Artist", "getMusicDirectory.json&id=folder%3AAlpha+Artist"}
	s := &session.Session{Sources: seed.Sources, Resolutions: map[string]map[string]string{}}
	for _, k := range keep {
		for _, req := range seed.Requests {
			if req.Key() != k {
				continue
			}
			s.Requests = append(s.Requests, req)
			for _, p := range req.Params {
				s.Resolutions[p.Value] = seed.Resolutions[p.Value]
			}
		}
	}
	if len(s.Requests) != len(keep) {
		t.Fatalf("found %d of the seed's requests, want %d", len(s.Requests), len(keep))
	}
	dir := filepath.Join(t.TempDir(), "scope")
	if err := session.Write(dir, s); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestGradeKeepsScopedPairsInScope grades the scope session with and
// without a by-construction row over the folder view. Without the row,
// the folder's id pairing collides with the artist's and the artist
// answer reports it; with the row, the folder's pairs stay inside the
// row's scope and the artist answer reports no identity finding.
func TestGradeKeepsScopedPairsInScope(t *testing.T) {
	dir := scopeSession(t)
	artist := "getArtist.json&id=artist%3AAlpha+Artist"
	identities := func(r *Report, request string) []Finding {
		var out []Finding
		for _, f := range r.Findings {
			if f.Request == request && f.Kind == KindIdentity {
				out = append(out, f)
			}
		}
		return out
	}

	g, err := New(emptyLedger(t))
	if err != nil {
		t.Fatal(err)
	}
	r, err := g.Grade(context.Background(), dir, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	if len(identities(r, artist)) == 0 {
		t.Fatal("without a scope the folder and artist pairs do not collide; the fixture proves nothing")
	}

	row := Row{
		ID: "folder-view", Area: "getMusicDirectory", Class: ClassByConstruction, Sources: []string{"navidrome", "gonic"},
		Oracle: "o", Dubplate: "d", Evidence: "e", State: StateActive,
		Pattern: Pattern{Endpoint: "getMusicDirectory", Format: "any", Path: "subsonic-response/directory"},
		Session: "scope", Versions: map[string]string{"navidrome": navidromeVersion, "gonic": gonicVersion},
	}
	ledger, _ := writeLedger(t, []Row{row})
	cfg := seedConfig(ledger)
	cfg.Sessions = filepath.Dir(dir)
	g, err = New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	r, err = g.Grade(context.Background(), dir, pairNG)
	if err != nil {
		t.Fatal(err)
	}
	if got := identities(r, artist); len(got) != 0 {
		t.Errorf("a folder pair leaked into the artist answer: %+v", got)
	}
	inScope := identities(r, "getMusicDirectory.json&id=folder%3AAlpha+Artist")
	if len(inScope) == 0 {
		t.Error("the collision is not reported inside the folder view's scope")
	}
}

// gradeNamedSessions grades, under the ledger cfg names, every session an
// active row names, on that row's pair, and returns every refusal and
// stale row found.
func gradeNamedSessions(t *testing.T, cfg Config) (graded int, err error) {
	t.Helper()
	g, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	var errs []error
	for _, r := range g.ledger.Rows {
		if r.State != StateActive {
			continue
		}
		pair := [2]string{r.Sources[0], r.Sources[1]}
		key := r.Session + "\x00" + strings.Join(slices.Sorted(slices.Values(pair[:])), "\x00")
		if seen[key] {
			continue
		}
		seen[key] = true
		graded++
		_, err := g.Grade(context.Background(), filepath.Join(cfg.Sessions, r.Session), pair)
		errs = append(errs, err)
	}
	return graded, errors.Join(errs...)
}

// TestCommittedLedgerRowsAreLive grades every session the committed
// ledger's active rows name, so a session joins the staleness check by
// being named in a row, with no edit here.
func TestCommittedLedgerRowsAreLive(t *testing.T) {
	n, err := gradeNamedSessions(t, seedConfig("../data/ledger.json"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("graded %d session pairs named by the committed ledger", n)
}

// TestGradeNamedSessionsReportsAStaleRow drives the same walk over a
// fixture ledger naming the seed session: one row matches a real finding
// and one names a field no answer carries.
func TestGradeNamedSessionsReportsAStaleRow(t *testing.T) {
	live := Row{
		ID: "live", Area: "ping", Class: ClassVolatile, Sources: []string{"navidrome", "gonic"},
		Oracle: "o", Dubplate: "d", Evidence: "e", State: StateActive,
		Pattern: Pattern{Endpoint: "ping", Format: "any", Path: "subsonic-response/type"},
		Session: "seed", Versions: map[string]string{"navidrome": navidromeVersion, "gonic": gonicVersion},
	}
	dead := live
	dead.ID, dead.Pattern.Path = "dead", "subsonic-response/nothing"
	raw, err := json.Marshal([]Row{live, dead})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "ledger.json")
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	n, err := gradeNamedSessions(t, seedConfig(path))
	if n != 1 {
		t.Errorf("graded %d session pairs, want 1", n)
	}
	if !errors.Is(err, ErrStaleRow) || !strings.Contains(err.Error(), `"dead"`) || strings.Contains(err.Error(), `"live"`) {
		t.Errorf("err = %v, want a stale row naming dead alone", err)
	}
}

func TestReportUnledgered(t *testing.T) {
	r := &Report{ledger: Result{Matches: []string{"a", "", "b", ""}}}
	if got := r.Unledgered(); got != 2 {
		t.Errorf("Unledgered = %d, want 2", got)
	}
}
