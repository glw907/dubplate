package cassette

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCanonicalizePathQuery(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"no query", "release", "release"},
		{"trailing slash, no query", "release/", "release/"},
		{"already sorted", "release?fmt=json&inc=media", "release?fmt=json&inc=media"},
		{"unsorted pairs get sorted", "release?inc=media&fmt=json", "release?fmt=json&inc=media"},
		{"literal plus in inc value preserved", "release?fmt=json&inc=artists+media", "release?fmt=json&inc=artists+media"},
		{"lucene escapes preserved verbatim", `release/?query=artist%3A%22Sigur%5C+R%C3%B3s%22&fmt=json`, `release/?fmt=json&query=artist%3A%22Sigur%5C+R%C3%B3s%22`},
		{"empty query", "release?", "release"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canonicalizePathQuery(tt.in); got != tt.want {
				t.Errorf("canonicalizePathQuery(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

// TestCanonicalizePathQueryNearMissesDoNotCollapse proves the
// canonicalization rule only reorders pairs; it does not treat a
// trailing-slash variant as the same request just because its query
// pairs, once sorted, look the same. A lookup keyed on one string never
// matches a case that stored the other.
func TestCanonicalizePathQueryNearMissesDoNotCollapse(t *testing.T) {
	withSlash := canonicalizePathQuery("release/?b=2&a=1")
	withoutSlash := canonicalizePathQuery("release?a=1&b=2")
	if withSlash == withoutSlash {
		t.Errorf("canonicalizePathQuery collapsed a trailing-slash near miss: both produced %q", withSlash)
	}
}

// baseCase returns a fully populated Case, including a derived
// Expected record, so the round-trip test exercises the whole schema
// rather than only its required fields.
func baseCase() *Case {
	return &Case{
		ID:       "smoke-clean-match",
		Profile:  "estate",
		EvalDate: "2020-06-15",
		Input: Input{
			// Numeric fields use float64: that is what json.Unmarshal
			// produces for a map[string]any, so a literal int here would
			// never compare equal to the value LoadCase decodes back.
			Items: []ItemFields{
				{"title": "Track One", "artist": "Example Artist", "track": float64(1)},
				{"title": "Track Two", "artist": "Example Artist", "track": float64(2)},
			},
		},
		Requests: []Request{
			{Method: "GET", PathQuery: "release/00000000-0000-0000-0000-000000000000?fmt=json&inc=artists+media", Response: "responses/001.json"},
		},
		Extra: []Request{
			{Method: "GET", PathQuery: "release/11111111-1111-1111-1111-111111111111?fmt=json", Response: "responses/002.json"},
		},
		Expected: &Expected{
			Recommendation: "strong",
			Action:         "apply",
			EchoedDate:     "2020-06-15",
			ConfigHash:     "deadbeef",
			Weights:        map[string]float64{"album": 3, "artist": 3, "year": 1},
			Thresholds: Thresholds{
				StrongRecThresh: 0.04,
				MediumRecThresh: 0.25,
				RecGapThresh:    0.25,
				MaxRec:          map[string]string{"missing_tracks": "medium", "unmatched_tracks": "medium"},
			},
			Candidates: []Candidate{
				{
					Source:    "MusicBrainz",
					AlbumID:   "00000000-0000-0000-0000-000000000000",
					Distance:  0.01,
					Penalties: map[string]float64{"source": 0.0, "album": 0.01},
					Tracks:    []float64{0.0, 0.0},
				},
			},
			Assignment:    map[int]int{0: 0, 1: 1},
			ExtraItems:    []int{},
			ExtraTracks:   []int{},
			AppliedFields: map[string]any{"album": "Example Album", "year": float64(2020)},
		},
	}
}

// writeResponseFiles drops the case's referenced response bodies onto
// disk directly: Case.Write only persists case.json, so a fixture's
// bodies are placed independently, the same way the capture tool will
// place them.
func writeResponseFiles(t *testing.T, dir string, bodies map[string][]byte) {
	t.Helper()
	for name, body := range bodies {
		full := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", name, err)
		}
		if err := os.WriteFile(full, body, 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
}

func TestWriteLoadCaseRoundTrip(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	bodies := map[string][]byte{
		"responses/001.json": []byte(`{"release":{"id":"00000000-0000-0000-0000-000000000000"}}`),
		"responses/002.json": []byte(`{"release":{"id":"11111111-1111-1111-1111-111111111111"}}`),
	}
	writeResponseFiles(t, dir, bodies)

	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}

	got, err := LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase: %v", err)
	}
	if !reflect.DeepEqual(got, c) {
		t.Errorf("LoadCase after Write = %#v, want %#v", got, c)
	}

	// The response bodies were never touched by Write or LoadCase: prove
	// they still read back byte-identical to what was written.
	for name, want := range bodies {
		gotBody, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read back %s: %v", name, err)
		}
		if string(gotBody) != string(want) {
			t.Errorf("response %s = %q, want %q", name, gotBody, want)
		}
	}
}

func TestWriteIsDeterministic(t *testing.T) {
	dir1, dir2 := t.TempDir(), t.TempDir()
	c := baseCase()
	if err := c.Write(dir1); err != nil {
		t.Fatalf("Write dir1: %v", err)
	}
	if err := c.Write(dir2); err != nil {
		t.Fatalf("Write dir2: %v", err)
	}
	b1, err := os.ReadFile(filepath.Join(dir1, caseFileName))
	if err != nil {
		t.Fatalf("read dir1 case.json: %v", err)
	}
	b2, err := os.ReadFile(filepath.Join(dir2, caseFileName))
	if err != nil {
		t.Fatalf("read dir2 case.json: %v", err)
	}
	if string(b1) != string(b2) {
		t.Errorf("Write produced different bytes for an identical Case:\n%s\nvs\n%s", b1, b2)
	}
}

func TestLoadCaseRejectsMissingCaseFile(t *testing.T) {
	dir := t.TempDir() // empty: no case.json
	if _, err := LoadCase(dir); err == nil {
		t.Fatal("LoadCase on a directory with no case.json: got nil error, want one")
	}
}

func TestLoadCaseRejectsAbsentEvalDate(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.EvalDate = ""
	c.Extra = nil
	writeResponseFiles(t, dir, map[string][]byte{"responses/001.json": []byte(`{}`)})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with absent eval_date: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "eval_date") {
		t.Errorf("error %q does not name eval_date", err)
	}
}

func TestLoadCaseRejectsNonexistentResponseFile(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = nil
	// Deliberately do not write responses/001.json.
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with a Request.Response naming a nonexistent file: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "responses/001.json") {
		t.Errorf("error %q does not name the missing response file", err)
	}
}

func TestLoadCaseRejectsUnreadableResponseFile(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root: file permissions do not block reads")
	}
	dir := t.TempDir()
	c := baseCase()
	c.Extra = nil
	writeResponseFiles(t, dir, map[string][]byte{"responses/001.json": []byte(`{}`)})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	unreadable := filepath.Join(dir, "responses", "001.json")
	if err := os.Chmod(unreadable, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}
	t.Cleanup(func() { _ = os.Chmod(unreadable, 0o644) })

	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with an unreadable response file: got nil error, want one")
	}
}

func TestLoadCaseRejectsExtraNamingConsumedResponse(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	// A different (method, path_query) from the sole Request, but the
	// same response filename it already consumes: this isolates the
	// filename-exclusivity rule from the key-exclusivity rule below.
	c.Extra = []Request{
		{Method: "GET", PathQuery: "release/22222222-2222-2222-2222-222222222222?fmt=json", Response: "responses/001.json"},
	}
	writeResponseFiles(t, dir, map[string][]byte{"responses/001.json": []byte(`{}`)})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with an Extra entry naming a consumed response: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "already consumed") {
		t.Errorf("error %q does not explain the consumed-response conflict", err)
	}
}

// TestLoadCaseRejectsExtraWithMissingMethod proves the Extra loop
// applies the same non-empty-Method rule the Requests loop applies.
func TestLoadCaseRejectsExtraWithMissingMethod(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = []Request{
		{Method: "", PathQuery: "release/11111111-1111-1111-1111-111111111111?fmt=json", Response: "responses/002.json"},
	}
	writeResponseFiles(t, dir, map[string][]byte{
		"responses/001.json": []byte(`{}`),
		"responses/002.json": []byte(`{}`),
	})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with an Extra entry missing Method: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "extra[0]") || !strings.Contains(err.Error(), "method") {
		t.Errorf("error %q does not name extra[0]'s missing method", err)
	}
}

// TestLoadCaseRejectsExtraWithMissingResponse proves the Extra loop
// applies the same non-empty-Response rule the Requests loop applies.
func TestLoadCaseRejectsExtraWithMissingResponse(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = []Request{
		{Method: "GET", PathQuery: "release/11111111-1111-1111-1111-111111111111?fmt=json", Response: ""},
	}
	writeResponseFiles(t, dir, map[string][]byte{"responses/001.json": []byte(`{}`)})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with an Extra entry missing Response: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "extra[0]") || !strings.Contains(err.Error(), "response: missing") {
		t.Errorf("error %q does not name extra[0]'s missing response", err)
	}
}

// TestLoadCaseRejectsExtraWithNonCanonicalPathQuery proves the Extra
// loop applies the same canonical-form rule the Requests loop applies,
// naming both the stored and the canonical form.
func TestLoadCaseRejectsExtraWithNonCanonicalPathQuery(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = []Request{
		{Method: "GET", PathQuery: "release/11111111-1111-1111-1111-111111111111?inc=x&fmt=json", Response: "responses/002.json"}, // unsorted
	}
	writeResponseFiles(t, dir, map[string][]byte{
		"responses/001.json": []byte(`{}`),
		"responses/002.json": []byte(`{}`),
	})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with a non-canonical Extra path_query: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "not canonical") {
		t.Errorf("error %q does not explain the canonicalization failure", err)
	}
}

// TestLoadCaseRejectsExtraNamingNonexistentResponseFile proves the
// Extra loop applies the same readability rule the Requests loop
// applies.
func TestLoadCaseRejectsExtraNamingNonexistentResponseFile(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = []Request{
		{Method: "GET", PathQuery: "release/11111111-1111-1111-1111-111111111111?fmt=json", Response: "responses/999.json"},
	}
	// Deliberately do not write responses/999.json.
	writeResponseFiles(t, dir, map[string][]byte{"responses/001.json": []byte(`{}`)})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with an Extra entry naming a nonexistent response file: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "responses/999.json") {
		t.Errorf("error %q does not name the missing response file", err)
	}
}

// TestLoadCaseRejectsExtraRepeatingRequestKey proves an Extra entry may
// not repeat a Request's own (Method, PathQuery) pair even when it
// names a different response file: the stub keys its replay map on
// that pair, and a Request entry would silently shadow the Extra one.
func TestLoadCaseRejectsExtraRepeatingRequestKey(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = []Request{
		{Method: c.Requests[0].Method, PathQuery: c.Requests[0].PathQuery, Response: "responses/002.json"},
	}
	writeResponseFiles(t, dir, map[string][]byte{
		"responses/001.json": []byte(`{}`),
		"responses/002.json": []byte(`{}`),
	})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with an Extra entry repeating a Request's key: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "is already a logged request") {
		t.Errorf("error %q does not explain the request-key conflict", err)
	}
}

// TestLoadCaseRejectsDuplicateExtraKeys proves two Extra entries may
// not share a (Method, PathQuery) pair even when they name different
// response files, and that the error names both indices.
func TestLoadCaseRejectsDuplicateExtraKeys(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = []Request{
		{Method: "GET", PathQuery: "release/11111111-1111-1111-1111-111111111111?fmt=json", Response: "responses/002.json"},
		{Method: "GET", PathQuery: "release/11111111-1111-1111-1111-111111111111?fmt=json", Response: "responses/003.json"},
	}
	writeResponseFiles(t, dir, map[string][]byte{
		"responses/001.json": []byte(`{}`),
		"responses/002.json": []byte(`{}`),
		"responses/003.json": []byte(`{}`),
	})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with two Extra entries sharing a key: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "extra[1]") || !strings.Contains(err.Error(), "duplicates extra[0]") {
		t.Errorf("error %q does not name both colliding indices", err)
	}
}

// TestLoadCaseLoadsExtraRequestShapeFixture loads the committed
// hand-authored fixture whose single Extra entry uses the
// Request-shaped schema, and asserts the decoded method, canonical
// path_query, and response name survive unchanged.
func TestLoadCaseLoadsExtraRequestShapeFixture(t *testing.T) {
	c, err := LoadCase(filepath.Join("testdata", "extra-request-shape"))
	if err != nil {
		t.Fatalf("LoadCase(testdata/extra-request-shape): %v", err)
	}
	if len(c.Extra) != 1 {
		t.Fatalf("Extra has %d entries, want 1", len(c.Extra))
	}
	extra := c.Extra[0]
	if extra.Method != "GET" {
		t.Errorf("Extra[0].Method = %q, want GET", extra.Method)
	}
	wantPathQuery := "release/00000000-0000-0000-0000-000000000002?fmt=json"
	if extra.PathQuery != canonicalizePathQuery(extra.PathQuery) {
		t.Errorf("Extra[0].PathQuery %q is not canonical", extra.PathQuery)
	}
	if extra.PathQuery != wantPathQuery {
		t.Errorf("Extra[0].PathQuery = %q, want %q", extra.PathQuery, wantPathQuery)
	}
	if extra.Response != "responses/002.json" {
		t.Errorf("Extra[0].Response = %q, want responses/002.json", extra.Response)
	}
}

func TestLoadCaseRejectsNonCanonicalPathQuery(t *testing.T) {
	dir := t.TempDir()
	c := baseCase()
	c.Extra = nil
	c.Requests[0].PathQuery = "release/00000000-0000-0000-0000-000000000000?inc=artists+media&fmt=json" // unsorted
	writeResponseFiles(t, dir, map[string][]byte{"responses/001.json": []byte(`{}`)})
	if err := c.Write(dir); err != nil {
		t.Fatalf("Write: %v", err)
	}
	_, err := LoadCase(dir)
	if err == nil {
		t.Fatal("LoadCase with a non-canonical (unsorted) path_query: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "not canonical") {
		t.Errorf("error %q does not explain the canonicalization failure", err)
	}
}

func TestLoadCaseCorruptedFixture(t *testing.T) {
	_, err := LoadCase(filepath.Join("testdata", "corrupted-case"))
	if err == nil {
		t.Fatal("LoadCase on testdata/corrupted-case: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "eval_date") {
		t.Errorf("error %q does not name the fixture's corruption (missing eval_date)", err)
	}
}

func TestLoadManifest(t *testing.T) {
	// The committed corpus manifest itself must load clean.
	m, err := LoadManifest(filepath.Join("..", "..", "..", "corpus", "manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest(corpus/manifest.json): %v", err)
	}
	if m.SchemaVersion == "" {
		t.Error("corpus/manifest.json: SchemaVersion is empty")
	}
	if m.Beets == "" {
		t.Error("corpus/manifest.json: Beets is empty")
	}
	if m.MediaFile == "" {
		t.Error("corpus/manifest.json: MediaFile is empty")
	}
}

func TestLoadManifestRejectsMissingSchemaVersion(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(`{"beets":"2.13.1"}`), 0o644); err != nil {
		t.Fatalf("write manifest fixture: %v", err)
	}
	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("LoadManifest with schema_version missing: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "schema_version") {
		t.Errorf("error %q does not name schema_version", err)
	}
}

func TestLoadManifestRejectsMissingBeets(t *testing.T) {
	path := filepath.Join(t.TempDir(), "manifest.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"1"}`), 0o644); err != nil {
		t.Fatalf("write manifest fixture: %v", err)
	}
	_, err := LoadManifest(path)
	if err == nil {
		t.Fatal("LoadManifest with beets missing: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "beets") {
		t.Errorf("error %q does not name beets", err)
	}
}

func TestLoadManifestRejectsUnreadablePath(t *testing.T) {
	_, err := LoadManifest(filepath.Join(t.TempDir(), "does-not-exist.json"))
	if err == nil {
		t.Fatal("LoadManifest on a nonexistent path: got nil error, want one")
	}
}
