//go:build oracle

// The tests in this file are the authoring loop's own end-to-end
// proof: generate a case with [Build], load it through
// [cassette.LoadCase], and derive it through the real, pinned oracle
// shim ([oracle.Derive], which starts its own [stub.Server] over the
// case directory internally). They are the only tests in this package
// that launch the real shim, so they are the only ones needing the uv
// venv (tools/venv/README.md). `make derive` runs them; the repository
// gate does not build this file at all.
package synth

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/oracle"
)

// venvOptions resolves the pinned oracle venv, honoring
// DUBPLATE_ORACLE_VENV per corpus/README.md's freeze section, the same
// resolution internal/harness/oracle's own derivation tests use.
func venvOptions(t *testing.T) oracle.Options {
	t.Helper()
	venv := os.Getenv("DUBPLATE_ORACLE_VENV")
	if venv == "" {
		venv = filepath.Join(repoRoot(t), "tools", "venv", ".venv")
	}
	if _, err := os.Stat(filepath.Join(venv, "bin", "python3")); err != nil {
		t.Fatalf("no venv at %s: build it per tools/venv/README.md: %v", venv, err)
	}
	return oracle.Options{Venv: venv}
}

// repoRoot walks up from the working directory to the nearest
// ancestor holding go.mod, matching internal/harness/corpus_test.go's
// own helper.
func repoRoot(t *testing.T) string {
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

// writeCase writes c's case.json and its release response into dir,
// the same layout `harness synth` itself produces.
func writeCase(t *testing.T, dir string, c *cassette.Case, doc []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), doc, 0o644); err != nil {
		t.Fatalf("write release response: %v", err)
	}
	if err := c.Write(dir); err != nil {
		t.Fatalf("write case: %v", err)
	}
}

// missFileExists reports whether dir carries a misses.json,
// oracle.Derive's own on-disk record of every request its internal
// stub could not answer (corpus/README.md, "The miss-file mechanism").
func missFileExists(t *testing.T, dir string) bool {
	t.Helper()
	_, err := os.Stat(filepath.Join(dir, "misses.json"))
	if err == nil {
		return true
	}
	if !os.IsNotExist(err) {
		t.Fatalf("stat misses.json: %v", err)
	}
	return false
}

// TestGeneratedCaseRoundTripsThroughRealDerivation is the authoring
// loop's end-to-end proof: a case Build generated, with every field
// exact, derives to a strong match against pinned beets and re-derives
// to byte-for-byte the same decoded record, proving the loop generate
// -> load -> serve -> derive -> re-derive holds for output this
// package actually produces, not a hand-typed fixture standing in for
// it.
func TestGeneratedCaseRoundTripsThroughRealDerivation(t *testing.T) {
	c, doc, err := Build("synth-roundtrip", minimalSpec())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	dir := t.TempDir()
	writeCase(t, dir, c, doc)

	loaded, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase: %v", err)
	}

	opts := venvOptions(t)
	opts.Dir = dir

	first, err := oracle.Derive(context.Background(), loaded, opts)
	if err != nil {
		t.Fatalf("first derivation: %v", err)
	}
	if first.Recommendation != "strong" {
		t.Errorf("recommendation = %q, want strong (every field matches the release exactly)", first.Recommendation)
	}
	if len(first.Candidates) != 1 {
		t.Fatalf("candidates = %d, want 1", len(first.Candidates))
	}
	if missFileExists(t, dir) {
		t.Error("misses.json present after a strong ID-lookup match; the search fallback should never fire")
	}

	second, err := oracle.Derive(context.Background(), loaded, opts)
	if err != nil {
		t.Fatalf("second derivation: %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Errorf("re-derivation diverges from the first:\nfirst:  %+v\nsecond: %+v", first, second)
	}
}

// TestGeneratedCaseWithEarlierOriginalDateDerivesASmallerYearPenalty
// proves OriginalDate reaches a real derivation, not only the
// generated document, by discriminating on its value rather than
// merely its presence. beets/autotag/distance.py's year branch (with
// preferred.original_year unset, the default both corpus profiles
// leave it at) scores zero whenever the items' own year matches either
// the release's year or the release-group's original year, so items
// carrying the release's own year alone are not enough: this spec also
// overrides every item's "year" tag away from both the release date's
// year (1999, from Date) and the release-group's original year (1975,
// when OriginalDate is set), landing on the diff/diff_max ratio branch,
// which is where a nonzero "year" component comes from. Deriving the
// identical spec twice, once with OriginalDate set and once left nil,
// shows the derived output actually changes with the field: an earlier
// original year widens distance.py's diff_max denominator, so the
// OriginalDate run's penalty must land strictly below the nil run's. A
// test that only checked "nonzero" would still pass with OriginalDate
// deleted from release.go entirely, since the nil run alone is already
// nonzero (both years differ from the item's overridden year).
func TestGeneratedCaseWithEarlierOriginalDateDerivesASmallerYearPenalty(t *testing.T) {
	yearOverrides := []Override{
		{Op: OpField, Index: 0, Fields: cassette.ItemFields{"year": 1980}},
		{Op: OpField, Index: 1, Fields: cassette.ItemFields{"year": 1980}},
	}

	withOriginal := minimalSpec()
	original := "1975-06-15"
	withOriginal.OriginalDate = &original
	withOriginal.Overrides = yearOverrides
	withPenalty := derivedYearPenalty(t, "synth-original-date-set", withOriginal)

	withoutOriginal := minimalSpec()
	withoutOriginal.Overrides = yearOverrides
	withoutPenalty := derivedYearPenalty(t, "synth-original-date-nil", withoutOriginal)

	if withPenalty == 0 {
		t.Errorf("year penalty (OriginalDate set) = 0, want nonzero")
	}
	if withoutPenalty == 0 {
		t.Errorf("year penalty (OriginalDate nil) = 0, want nonzero")
	}
	if withPenalty >= withoutPenalty {
		t.Errorf("year penalty (OriginalDate set) = %v, want strictly less than the OriginalDate-nil penalty %v: an earlier original year should widen distance.py's diff_max denominator", withPenalty, withoutPenalty)
	}
}

// derivedYearPenalty builds and derives spec under id, returning the
// single candidate's "year" penalty component.
func derivedYearPenalty(t *testing.T, id string, spec *ReleaseSpec) float64 {
	t.Helper()

	c, doc, err := Build(id, spec)
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	dir := t.TempDir()
	writeCase(t, dir, c, doc)

	loaded, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase: %v", err)
	}

	opts := venvOptions(t)
	opts.Dir = dir

	exp, err := oracle.Derive(context.Background(), loaded, opts)
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}
	if len(exp.Candidates) == 0 {
		t.Fatal("candidates = 0, want the single looked-up release")
	}
	return exp.Candidates[0].Penalties["year"]
}

// TestMalformedReleaseDocumentFailsLoudly proves a generated case
// whose response body beets cannot parse never sneaks a plausible-
// looking record past a derivation: it drops the "text-representation"
// key album_info reads directly (release["text_representation"]
// ["script"]), which pinned beets can only get from the response body
// itself.
//
// beets' own plugin-call boundary (metadata_plugins.
// maybe_handle_plugin_error, config raise_on_error: no by default)
// catches that KeyError and logs it rather than raising: Derive itself
// returns no error, the shim process exits 0, and the record it
// prints is well-formed JSON, exactly the shape a naive round-trip
// test would wave through. The candidate that KeyError killed never
// gets scored, though, so recommendation collapses to "none" with an
// empty candidate list and a search-fallback miss behind it; asserting
// that shape explicitly is what makes a malformed document "fail
// loudly" in this test, rather than the corrupted document quietly
// producing what looks like an ordinary derived record.
func TestMalformedReleaseDocumentFailsLoudly(t *testing.T) {
	c, doc, err := Build("synth-malformed", minimalSpec())
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	var raw map[string]any
	if err := json.Unmarshal(doc, &raw); err != nil {
		t.Fatalf("decode generated document: %v", err)
	}
	delete(raw, "text-representation")
	corrupted, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("re-encode corrupted document: %v", err)
	}

	dir := t.TempDir()
	writeCase(t, dir, c, corrupted)

	loaded, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase: %v", err)
	}

	opts := venvOptions(t)
	opts.Dir = dir

	exp, err := oracle.Derive(context.Background(), loaded, opts)
	if err != nil {
		t.Fatalf("Derive against a malformed release document returned an error instead of a none-recommendation record: %v", err)
	}
	if exp.Recommendation != "none" {
		t.Errorf("recommendation = %q, want none: a malformed document must never score a plausible candidate", exp.Recommendation)
	}
	if len(exp.Candidates) != 0 {
		t.Errorf("candidates = %d, want 0: the malformed release must never survive as a scored candidate", len(exp.Candidates))
	}
	if !missFileExists(t, dir) {
		t.Error("misses.json absent; expected the search fallback beets falls through to once the ID lookup's candidate is dropped")
	}
}
