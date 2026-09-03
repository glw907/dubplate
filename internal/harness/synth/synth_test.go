package synth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// minimalSpec returns a two-track, single-medium spec with no
// perturbation: the shape every override test starts from.
func minimalSpec() *ReleaseSpec {
	return &ReleaseSpec{
		Album:     "Test Album",
		Artist:    "Test Artist",
		Date:      "1999-03-02",
		Country:   "US",
		MBAlbumID: "aaaaaaaa-1111-2222-3333-444444444444",
		Profile:   "estate",
		EvalDate:  "2024-01-15",
		Media: []Medium{{
			Format: "CD",
			Tracks: []Track{
				{Title: "First Track", Length: 200.5, Position: 1},
				{Title: "Second Track", Length: 180.0, Position: 2},
			},
		}},
	}
}

func TestBuildDerivesItemsFromReleaseTracks(t *testing.T) {
	c, doc, err := Build("test-case", minimalSpec())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(c.Input.Items) != 2 {
		t.Fatalf("items = %d, want 2", len(c.Input.Items))
	}

	first := c.Input.Items[0]
	if first["title"] != "First Track" {
		t.Errorf("item 0 title = %v, want First Track", first["title"])
	}
	if first["album"] != "Test Album" || first["albumartist"] != "Test Artist" || first["artist"] != "Test Artist" {
		t.Errorf("item 0 album fields = %+v, want Test Album/Test Artist", first)
	}
	if first["mb_albumid"] != "aaaaaaaa-1111-2222-3333-444444444444" {
		t.Errorf("item 0 mb_albumid = %v, want the spec's release id", first["mb_albumid"])
	}
	if first["year"] != 1999 {
		t.Errorf("item 0 year = %v, want 1999", first["year"])
	}
	if _, hasDisc := first["disc"]; hasDisc {
		t.Errorf("single-medium item carries a disc field: %+v", first)
	}
	if _, hasComp := first["comp"]; hasComp {
		t.Errorf("non-VA item carries a comp field: %+v", first)
	}

	if len(doc) == 0 {
		t.Fatal("Build returned an empty release document")
	}
}

func TestBuildMultiDiscSetsDiscField(t *testing.T) {
	spec := minimalSpec()
	spec.Media = append(spec.Media, Medium{
		Format: "CD",
		Tracks: []Track{{Title: "Disc Two Track", Length: 150, Position: 1}},
	})

	c, _, err := Build("multi-disc-test", spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(c.Input.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(c.Input.Items))
	}
	if c.Input.Items[0]["disc"] != 1 {
		t.Errorf("disc-1 item disc field = %v, want 1", c.Input.Items[0]["disc"])
	}
	if c.Input.Items[2]["disc"] != 2 {
		t.Errorf("disc-2 item disc field = %v, want 2", c.Input.Items[2]["disc"])
	}
	if c.Input.Items[2]["track"] != 1 {
		t.Errorf("disc-2 track field = %v, want 1 (per-disc numbering)", c.Input.Items[2]["track"])
	}
}

func TestBuildVariousArtistsSetsCompAndPerTrackArtist(t *testing.T) {
	spec := minimalSpec()
	spec.VariousArtists = true
	spec.Artist = "Various Artists"
	spec.Media[0].Tracks[0].ArtistCredit = "Artist One"
	spec.Media[0].Tracks[1].ArtistCredit = "Artist Two"

	c, doc, err := Build("va-test", spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if c.Input.Items[0]["comp"] != true {
		t.Errorf("VA item comp = %v, want true", c.Input.Items[0]["comp"])
	}
	if c.Input.Items[0]["artist"] != "Artist One" || c.Input.Items[1]["artist"] != "Artist Two" {
		t.Errorf("VA per-track artists = %v / %v, want Artist One / Artist Two", c.Input.Items[0]["artist"], c.Input.Items[1]["artist"])
	}
	if c.Input.Items[0]["albumartist"] != "Various Artists" {
		t.Errorf("VA albumartist = %v, want Various Artists", c.Input.Items[0]["albumartist"])
	}

	if !strings.Contains(string(doc), variousArtistsMBID) {
		t.Errorf("VA release document does not carry MusicBrainz's well-known Various Artists id:\n%s", doc)
	}
}

func TestBuildAppliesFieldOverride(t *testing.T) {
	spec := minimalSpec()
	spec.Overrides = []Override{
		{Op: OpField, Index: 0, Fields: cassette.ItemFields{"title": "Wrong Title", "year": 1970}},
	}

	c, _, err := Build("field-override-test", spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if c.Input.Items[0]["title"] != "Wrong Title" {
		t.Errorf("overridden title = %v, want Wrong Title", c.Input.Items[0]["title"])
	}
	if c.Input.Items[0]["year"] != 1970 {
		t.Errorf("overridden year = %v, want 1970", c.Input.Items[0]["year"])
	}
	if c.Input.Items[1]["title"] != "Second Track" {
		t.Errorf("item 1 title changed by an override targeting item 0: %v", c.Input.Items[1]["title"])
	}
	if c.Input.Items[0]["album"] != "Test Album" {
		t.Errorf("overridden item lost an untouched field: album = %v, want Test Album", c.Input.Items[0]["album"])
	}
}

func TestBuildAppliesMissingOverride(t *testing.T) {
	spec := minimalSpec()
	spec.Overrides = []Override{{Op: OpMissing, Index: 0}}

	c, _, err := Build("missing-override-test", spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(c.Input.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(c.Input.Items))
	}
	if c.Input.Items[0]["title"] != "Second Track" {
		t.Errorf("remaining item = %v, want Second Track", c.Input.Items[0]["title"])
	}
}

func TestBuildAppliesExtraOverrideCarryingTheAlbumMBID(t *testing.T) {
	spec := minimalSpec()
	spec.Overrides = []Override{
		{Op: OpExtra, Fields: cassette.ItemFields{"title": "Bonus Track", "album": "Test Album", "artist": "Test Artist", "track": 99}},
	}

	c, _, err := Build("extra-override-test", spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(c.Input.Items) != 3 {
		t.Fatalf("items = %d, want 3", len(c.Input.Items))
	}
	extra := c.Input.Items[2]
	if extra["title"] != "Bonus Track" {
		t.Errorf("extra item title = %v, want Bonus Track", extra["title"])
	}
	if extra["mb_albumid"] != spec.MBAlbumID {
		t.Errorf("extra item mb_albumid = %v, want %v (the album's own id)", extra["mb_albumid"], spec.MBAlbumID)
	}
}

func TestBuildFencesMBAlbumIDOffFieldOverrides(t *testing.T) {
	spec := minimalSpec()
	spec.Overrides = []Override{
		{Op: OpField, Index: 0, Fields: cassette.ItemFields{"mb_albumid": "bbbbbbbb-0000-0000-0000-000000000000"}},
	}

	_, _, err := Build("fenced-test", spec)
	if err == nil {
		t.Fatal("Build with an override perturbing mb_albumid outside family 10: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "fenced") {
		t.Errorf("error %q does not name the fencing rule", err)
	}
}

func TestBuildFencesMBAlbumIDOffExtraOverrides(t *testing.T) {
	spec := minimalSpec()
	spec.Overrides = []Override{
		{Op: OpExtra, Fields: cassette.ItemFields{"title": "Sneaky", "mb_albumid": "bbbbbbbb-0000-0000-0000-000000000000"}},
	}

	_, _, err := Build("fenced-extra-test", spec)
	if err == nil {
		t.Fatal("Build with an extra override setting its own mb_albumid outside family 10: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "fenced") {
		t.Errorf("error %q does not name the fencing rule", err)
	}
}

func TestBuildAllowsMBAlbumIDOverrideWhenFlagSet(t *testing.T) {
	spec := minimalSpec()
	spec.AllowAlbumIDOverride = true
	spec.Overrides = []Override{
		{Op: OpField, Index: 0, Fields: cassette.ItemFields{"mb_albumid": "bbbbbbbb-0000-0000-0000-000000000000"}},
	}

	c, _, err := Build("family-10-test", spec)
	if err != nil {
		t.Fatalf("Build with AllowAlbumIDOverride set: %v", err)
	}
	if c.Input.Items[0]["mb_albumid"] != "bbbbbbbb-0000-0000-0000-000000000000" {
		t.Errorf("overridden mb_albumid = %v, want the family-10 override value", c.Input.Items[0]["mb_albumid"])
	}
}

func TestBuildRejectsOverrideIndexOutOfRange(t *testing.T) {
	spec := minimalSpec()
	spec.Overrides = []Override{{Op: OpField, Index: 99, Fields: cassette.ItemFields{"title": "x"}}}

	_, _, err := Build("bad-index-test", spec)
	if err == nil {
		t.Fatal("Build with an out-of-range override index: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("error %q does not name the out-of-range index", err)
	}
}

func TestBuildRejectsUnknownOverrideOp(t *testing.T) {
	spec := minimalSpec()
	spec.Overrides = []Override{{Op: "bogus", Index: 0}}

	_, _, err := Build("bad-op-test", spec)
	if err == nil {
		t.Fatal("Build with an unknown override op: got nil error, want one")
	}
}

func TestBuildWithIndependentItemsIgnoresMedia(t *testing.T) {
	spec := minimalSpec()
	spec.Items = []cassette.ItemFields{
		{"title": "Freestanding", "album": "Test Album", "artist": "Test Artist", "mb_albumid": spec.MBAlbumID, "track": 1, "year": 1999},
	}

	c, _, err := Build("independent-items-test", spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(c.Input.Items) != 1 || c.Input.Items[0]["title"] != "Freestanding" {
		t.Errorf("items = %+v, want the spec's own independent item set verbatim", c.Input.Items)
	}
}

func TestSpecValidateRejectsMissingFields(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*ReleaseSpec)
		wantHas string
	}{
		{"no album", func(s *ReleaseSpec) { s.Album = "" }, "album"},
		{"no artist", func(s *ReleaseSpec) { s.Artist = "" }, "artist"},
		{"no date", func(s *ReleaseSpec) { s.Date = "" }, "date"},
		{"no country", func(s *ReleaseSpec) { s.Country = "" }, "country"},
		{"no mb_albumid", func(s *ReleaseSpec) { s.MBAlbumID = "" }, "mb_albumid"},
		{"no profile", func(s *ReleaseSpec) { s.Profile = "" }, "profile"},
		{"no eval_date", func(s *ReleaseSpec) { s.EvalDate = "" }, "eval_date"},
		{"no media or items", func(s *ReleaseSpec) { s.Media = nil }, "media"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := minimalSpec()
			tt.mutate(spec)
			_, _, err := Build("validate-test", spec)
			if err == nil {
				t.Fatalf("Build with %s: got nil error, want one", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not name %q", err, tt.wantHas)
			}
		})
	}
}

// TestBuildProducesLoadCaseCanonicalRequestLog is the round-trip proof
// that canonical form comes from cassette's own rule, never a literal
// re-typed by hand: cassette.LoadCase rejects a case whose path_query
// is not already canonical (Task 1's frozen validation), and this test
// writes Build's own output straight through it with no massaging.
func TestBuildProducesLoadCaseCanonicalRequestLog(t *testing.T) {
	c, doc, err := Build("canonical-test", minimalSpec())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), doc, 0o644); err != nil {
		t.Fatalf("write response: %v", err)
	}
	if err := c.Write(dir); err != nil {
		t.Fatalf("write case: %v", err)
	}

	loaded, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase rejected Build's own output: %v", err)
	}
	if len(loaded.Requests) != 1 {
		t.Fatalf("loaded requests = %d, want 1", len(loaded.Requests))
	}

	// The inc= list joins with a literal "+" logically but must carry
	// the wire form "%2B" (corpus/README.md's canonicalization rule);
	// a decode-and-re-encode step would corrupt this silently.
	if !strings.Contains(loaded.Requests[0].PathQuery, "artists%2Bmedia") {
		t.Errorf("path_query = %q, want a %%2B-escaped inc= join", loaded.Requests[0].PathQuery)
	}
	if strings.Contains(loaded.Requests[0].PathQuery, "artists+media") {
		t.Errorf("path_query = %q carries a literal '+' join, never valid wire form", loaded.Requests[0].PathQuery)
	}
}

// TestBuildReleaseDocumentIsValidJSON is a cheap sanity check ahead of
// the oracle-tagged round trip: the document must at least parse
// before it is worth handing to a real shim.
func TestBuildReleaseDocumentIsValidJSON(t *testing.T) {
	_, doc, err := Build("json-sanity-test", minimalSpec())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	var probe map[string]any
	if err := json.Unmarshal(doc, &probe); err != nil {
		t.Fatalf("release document is not valid JSON: %v", err)
	}
	if probe["id"] != "aaaaaaaa-1111-2222-3333-444444444444" {
		t.Errorf("document id = %v, want the spec's mb_albumid", probe["id"])
	}
}

func TestPseudoUUIDIsDeterministic(t *testing.T) {
	a := pseudoUUID("seed-one")
	b := pseudoUUID("seed-one")
	c := pseudoUUID("seed-two")
	if a != b {
		t.Errorf("pseudoUUID(%q) = %q and %q, want equal for the same seed", "seed-one", a, b)
	}
	if a == c {
		t.Errorf("pseudoUUID for two different seeds collided: %q", a)
	}
}
