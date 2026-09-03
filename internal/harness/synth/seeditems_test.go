package synth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// seedItemsRepoRoot walks up from the working directory to the nearest
// ancestor holding go.mod. A distinct name from synth_derive_test.go's
// own repoRoot(t): that helper lives behind the "oracle" build tag, and
// `go test -tags oracle` compiles both files together, so a shared name
// here would collide.
func seedItemsRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("locate repo root: no go.mod above %s", dir)
		}
		dir = parent
	}
}

// committedCase is the case.json subset readCommittedCase needs: the
// response a case's own primary request names, and its committed
// baseline Input.Items, decoded as plain maps so its values compare
// directly against ProjectItems' own JSON-round-tripped output.
type committedCase struct {
	Input struct {
		Items []map[string]any `json:"items"`
	} `json:"input"`
	Requests []struct {
		Response string `json:"response"`
	} `json:"requests"`
}

// readCommittedCase reads the already-committed smoke case id's
// primary response document and its committed Input.Items, the two
// halves [ProjectItems]' round-trip proof diffs against each other.
func readCommittedCase(t *testing.T, root, id string) (doc []byte, items []map[string]any) {
	t.Helper()
	dir := filepath.Join(root, "corpus", "cases", id)

	data, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatalf("read %s case.json: %v", id, err)
	}
	var c committedCase
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("parse %s case.json: %v", id, err)
	}
	if len(c.Requests) == 0 {
		t.Fatalf("%s: case carries no requests to project from", id)
	}

	doc, err = os.ReadFile(filepath.Join(dir, c.Requests[0].Response))
	if err != nil {
		t.Fatalf("read %s response: %v", id, err)
	}
	return doc, c.Input.Items
}

// normalizeItems JSON-round-trips items so every value (int, float64,
// bool, string) decodes the same way encoding/json would decode a
// committed case.json's own items, making the two sides of a
// round-trip proof directly comparable with reflect.DeepEqual.
func normalizeItems(t *testing.T, items []map[string]any) []map[string]any {
	t.Helper()
	data, err := json.Marshal(items)
	if err != nil {
		t.Fatalf("marshal projected items: %v", err)
	}
	var out []map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("unmarshal projected items: %v", err)
	}
	return out
}

// diffItems reports got against want, key by key, per item index: the
// fire-proving evidence a round-trip proof owes, not merely a pass/fail
// bit. A projection that matches produces no diff lines at all.
func diffItems(got, want []map[string]any) []string {
	var diffs []string
	n := max(len(got), len(want))
	for i := range n {
		if i >= len(got) {
			diffs = append(diffs, fmt.Sprintf("item %d: missing from projection, want %+v", i, want[i]))
			continue
		}
		if i >= len(want) {
			diffs = append(diffs, fmt.Sprintf("item %d: projected but not committed: %+v", i, got[i]))
			continue
		}
		diffs = append(diffs, diffFields(i, got[i], want[i])...)
	}
	return diffs
}

// diffFields reports every key in got or want[i] whose value diverges,
// in sorted key order for a stable, readable report.
func diffFields(i int, got, want map[string]any) []string {
	keys := make(map[string]bool, len(got)+len(want))
	for k := range got {
		keys[k] = true
	}
	for k := range want {
		keys[k] = true
	}
	sorted := make([]string, 0, len(keys))
	for k := range keys {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	var diffs []string
	for _, k := range sorted {
		gv, gok := got[k]
		wv, wok := want[k]
		if !reflect.DeepEqual(gv, wv) || gok != wok {
			diffs = append(diffs, fmt.Sprintf("item %d[%s]: got %v (present=%v), want %v (present=%v)", i, k, gv, gok, wv, wok))
		}
	}
	return diffs
}

func TestProjectItemsMatchesCleanMatch(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, want := readCommittedCase(t, root, "clean-match")

	got, err := ProjectItems(doc, nil, false)
	if err != nil {
		t.Fatalf("ProjectItems: %v", err)
	}

	gotGeneric := make([]map[string]any, len(got))
	for i, item := range got {
		gotGeneric[i] = item
	}
	diffs := diffItems(normalizeItems(t, gotGeneric), normalizeItems(t, want))
	if len(diffs) != 0 {
		t.Errorf("clean-match: projection diverges from the committed baseline:\n%s", strings.Join(diffs, "\n"))
	}
}

func TestProjectItemsMatchesMultiDisc(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, want := readCommittedCase(t, root, "multi-disc")

	got, err := ProjectItems(doc, nil, false)
	if err != nil {
		t.Fatalf("ProjectItems: %v", err)
	}

	gotGeneric := make([]map[string]any, len(got))
	for i, item := range got {
		gotGeneric[i] = item
	}
	diffs := diffItems(normalizeItems(t, gotGeneric), normalizeItems(t, want))
	if len(diffs) != 0 {
		t.Errorf("multi-disc: projection diverges from the committed baseline:\n%s", strings.Join(diffs, "\n"))
	}
}

func TestProjectItemsMatchesVariousArtists(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, want := readCommittedCase(t, root, "various-artists")

	got, err := ProjectItems(doc, nil, false)
	if err != nil {
		t.Fatalf("ProjectItems: %v", err)
	}

	gotGeneric := make([]map[string]any, len(got))
	for i, item := range got {
		gotGeneric[i] = item
	}
	diffs := diffItems(normalizeItems(t, gotGeneric), normalizeItems(t, want))
	if len(diffs) != 0 {
		t.Errorf("various-artists: projection diverges from the committed baseline:\n%s", strings.Join(diffs, "\n"))
	}
}

// TestProjectItemsAppliesOverridesLikeYearPenalty proves the overrides
// reuse (B1's second half): year-penalty's committed items all carry
// year: 1978 while the release document itself dates to 1991, a
// divergence only a per-item field override can produce, and this
// test drives that exact override list through ProjectItems' overrides
// parameter rather than assuming the baseline alone would match.
func TestProjectItemsAppliesOverridesLikeYearPenalty(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, want := readCommittedCase(t, root, "year-penalty")

	overrides := make([]Override, len(want))
	for i := range want {
		overrides[i] = Override{Op: OpField, Index: i, Fields: map[string]any{"year": 1978}}
	}

	got, err := ProjectItems(doc, overrides, false)
	if err != nil {
		t.Fatalf("ProjectItems: %v", err)
	}

	gotGeneric := make([]map[string]any, len(got))
	for i, item := range got {
		gotGeneric[i] = item
	}
	diffs := diffItems(normalizeItems(t, gotGeneric), normalizeItems(t, want))
	if len(diffs) != 0 {
		t.Errorf("year-penalty: projection diverges from the committed baseline:\n%s", strings.Join(diffs, "\n"))
	}
}

func TestProjectItemsRejectsMalformedJSON(t *testing.T) {
	_, err := ProjectItems([]byte("not json"), nil, false)
	if err == nil {
		t.Fatal("ProjectItems against malformed JSON: got nil error, want one")
	}
}

func TestProjectItemsRejectsAMissingID(t *testing.T) {
	_, err := ProjectItems([]byte(`{"title": "No Id", "media": [{"tracks": [{"position": 1}]}]}`), nil, false)
	if err == nil {
		t.Fatal("ProjectItems against a document with no release id: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "id") {
		t.Errorf("error %q does not name the missing id", err)
	}
}

func TestProjectItemsRejectsEmptyMedia(t *testing.T) {
	_, err := ProjectItems([]byte(`{"id": "aaaaaaaa-1111-2222-3333-444444444444", "title": "No Media"}`), nil, false)
	if err == nil {
		t.Fatal("ProjectItems against a document with no media: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "media") {
		t.Errorf("error %q does not name the missing media", err)
	}
}

func TestProjectItemsPropagatesAnOutOfRangeOverride(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, _ := readCommittedCase(t, root, "clean-match")

	_, err := ProjectItems(doc, []Override{{Op: OpField, Index: 999, Fields: map[string]any{"title": "x"}}}, false)
	if err == nil {
		t.Fatal("ProjectItems with an out-of-range override index: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("error %q does not explain the out-of-range index", err)
	}
}

func TestProjectItemsFencesMBAlbumIDOverride(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, _ := readCommittedCase(t, root, "clean-match")

	_, err := ProjectItems(doc, []Override{{Op: OpField, Index: 0, Fields: map[string]any{mbAlbumIDField: "forged-id"}}}, false)
	if err == nil {
		t.Fatal("ProjectItems overriding mb_albumid without allowAlbumIDOverride: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "fenced off") {
		t.Errorf("error %q does not explain the mb_albumid fence", err)
	}
}

// wantYearPenaltyFacts is year-penalty's response document's own
// structural facts, transcribed by hand from responses/001.json (its
// release-group's first-release-date agrees with the release's own
// date, 1991-09-24; the override that diverges Input.Items' year to
// 1978 lives in case.json, not the captured document): a diff against
// this literal, not against FormatFacts' own output, is what proves
// the format matches the document rather than merely itself.
const wantYearPenaltyFacts = `release date: 1991-09-24
release-group first-release-date: 1991-09-24
release status: Official
discs: 1
  disc 1: 12 tracks
aliases:
  release:
    none
  artist-credit "Nirvana":
    Nirvana (type: Artist name, locale: en)
    Nirvana US
    ニルヴァーナ (type: Artist name, locale: ja)
`

func TestFormatFactsMatchesYearPenalty(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, _ := readCommittedCase(t, root, "year-penalty")

	facts, err := ExtractFacts(doc)
	if err != nil {
		t.Fatalf("ExtractFacts: %v", err)
	}
	got := FormatFacts(facts)
	if got != wantYearPenaltyFacts {
		t.Errorf("FormatFacts(year-penalty) =\n%s\nwant\n%s", got, wantYearPenaltyFacts)
	}
}

// wantMultiDiscFacts is multi-disc's response document's own
// structural facts (The Wall, two 13-track CDs; a release-group
// first-release-date of 1979-11-30 against the release's own
// 1984-09), transcribed by hand for the same reason
// wantYearPenaltyFacts is.
const wantMultiDiscFacts = `release date: 1984-09
release-group first-release-date: 1979-11-30
release status: Official
discs: 2
  disc 1: 13 tracks
  disc 2: 13 tracks
aliases:
  release:
    none
  artist-credit "Pink Floyd":
    Floyd (type: Search hint)
    Pink Floid (type: Search hint)
    Pink Floyd (type: Artist name, locale: en)
    The Pink Floyd (type: Search hint)
    ピンク・フロイド (type: Artist name, locale: ja)
    핑크 플로이드 (type: Artist name, locale: ko)
`

func TestFormatFactsMatchesMultiDisc(t *testing.T) {
	root := seedItemsRepoRoot(t)
	doc, _ := readCommittedCase(t, root, "multi-disc")

	facts, err := ExtractFacts(doc)
	if err != nil {
		t.Fatalf("ExtractFacts: %v", err)
	}
	got := FormatFacts(facts)
	if got != wantMultiDiscFacts {
		t.Errorf("FormatFacts(multi-disc) =\n%s\nwant\n%s", got, wantMultiDiscFacts)
	}
}

// noAliasesDoc is a hand-written release document that documents no
// alias at all: no top-level `aliases` entries and no artist-credit
// entry carrying an `aliases` array, the shape a `harness synth`-built
// document always has (marshalRelease never writes an aliases field).
const noAliasesDoc = `{
  "id": "aaaaaaaa-1111-2222-3333-444444444444",
  "title": "No Aliases",
  "date": "2000-01-01",
  "status": "Official",
  "release-group": {"first-release-date": "2000-01-01"},
  "artist-credit": [{"name": "Solo Artist", "artist": {"id": "bbbbbbbb-1111-2222-3333-444444444444"}}],
  "media": [{"tracks": [{"position": 1, "title": "Only Track", "length": 100000, "recording": {"id": "cccccccc-1111-2222-3333-444444444444"}}]}]
}`

// wantNoAliasesFacts is noAliasesDoc's structural facts: an explicit
// "aliases: none" line, never an omitted section, so a document that
// documents no alias is never indistinguishable from a formatting bug
// that silently dropped a populated aliases section.
const wantNoAliasesFacts = `release date: 2000-01-01
release-group first-release-date: 2000-01-01
release status: Official
discs: 1
  disc 1: 1 tracks
aliases: none
`

func TestFormatFactsPrintsExplicitNoneForADocumentWithNoAliases(t *testing.T) {
	facts, err := ExtractFacts([]byte(noAliasesDoc))
	if err != nil {
		t.Fatalf("ExtractFacts: %v", err)
	}
	got := FormatFacts(facts)
	if got != wantNoAliasesFacts {
		t.Errorf("FormatFacts(no-aliases) =\n%s\nwant\n%s", got, wantNoAliasesFacts)
	}
}

func TestExtractFactsRejectsMalformedJSON(t *testing.T) {
	_, err := ExtractFacts([]byte("not json"))
	if err == nil {
		t.Fatal("ExtractFacts against malformed JSON: got nil error, want one")
	}
}
