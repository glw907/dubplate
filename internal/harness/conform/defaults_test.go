package conform

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// corpusCaseDirs lists every case directory under corpus/cases,
// mirroring internal/harness/corpus_test.go's own helper (unexported
// and test-only there, so this package reimplements it rather than
// importing anything).
func corpusCaseDirs(t *testing.T, root string) []string {
	t.Helper()
	casesRoot := filepath.Join(root, "corpus", "cases")
	entries, err := os.ReadDir(casesRoot)
	if err != nil {
		t.Fatalf("read corpus/cases: %v", err)
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, filepath.Join(casesRoot, e.Name()))
		}
	}
	if len(dirs) == 0 {
		t.Fatal("corpus/cases has no case directories; the smoke corpus is missing")
	}
	return dirs
}

// TestCommittedCasesEchoTunablesDefaults is the M1 defaults
// cross-check's own half that can fire today: every committed case's
// echoed Weights and Thresholds must equal the tunables the sweep
// grades that case under, which is gradeTunables' merged result.
func TestCommittedCasesEchoTunablesDefaults(t *testing.T) {
	root := repoRoot(t)
	checked := 0
	for _, dir := range corpusCaseDirs(t, root) {
		c, err := cassette.LoadCase(dir)
		if err != nil {
			t.Fatalf("load %s: %v", dir, err)
		}
		if c.Expected == nil {
			continue // not yet derived; nothing to cross-check
		}
		checked++
		// The echo is compared against the case's OWN merged
		// configuration rather than against DefaultTunables directly,
		// which is what the M1 fold specifies (rung 3 spec, :279).
		// The two agree today because no committed profile sets a
		// match: or musicbrainz: key, so the move is
		// behavior-preserving and closes the drift before a fourth
		// profile makes it real.
		want, err := gradeTunables(c.Profile)
		if err != nil {
			t.Fatalf("case %s: %v", c.ID, err)
		}
		if err := CheckCaseWeights(c.ID, want.DistanceWeights, c.Expected.Weights); err != nil {
			t.Error(err)
		}
		if err := CheckCaseThresholds(c.ID, want, c.Expected.Thresholds); err != nil {
			t.Error(err)
		}
	}
	if checked == 0 {
		t.Fatal("no committed case carried an Expected record; the cross-check graded nothing")
	}
	t.Logf("cross-checked weights and thresholds against %d committed cases", checked)
}

// TestDefaultsCrossCheckCatchesPerturbedWeight proves the weights
// cross-check fires and names the key and both values, against a
// hand-authored perturbation (0.9 -> 0.09 on missing_tracks) rather
// than anything read from a committed case.
func TestDefaultsCrossCheckCatchesPerturbedWeight(t *testing.T) {
	want := matcher.DefaultTunables().DistanceWeights
	mutant := maps.Clone(want)
	mutant["missing_tracks"] = 0.09

	err := CheckCaseWeights("fixture/perturbed-weight", want, mutant)
	if err == nil {
		t.Fatal("expected the weights cross-check to fire on a perturbed default")
	}
	msg := err.Error()
	if !strings.Contains(msg, "missing_tracks") || !strings.Contains(msg, "0.09") || !strings.Contains(msg, "0.9") {
		t.Errorf("error does not name missing_tracks and both values: %v", err)
	}
}

// TestDefaultsCrossCheckCatchesPerturbedThreshold proves the
// thresholds half of the M1 cross-check fires and names the key and
// both values, against a hand-authored perturbation
// (0.25 -> 0.09 on medium_rec_thresh) rather than anything read from a
// committed case.
func TestDefaultsCrossCheckCatchesPerturbedThreshold(t *testing.T) {
	want := matcher.DefaultTunables()
	got := cassette.Thresholds{
		StrongRecThresh: want.StrongRecThresh,
		MediumRecThresh: 0.09,
		RecGapThresh:    want.RecGapThresh,
		MaxRec:          maps.Clone(want.MaxRec),
	}

	err := CheckCaseThresholds("fixture/perturbed-threshold", want, got)
	if err == nil {
		t.Fatal("expected the thresholds cross-check to fire on a perturbed default")
	}
	msg := err.Error()
	if !strings.Contains(msg, "medium_rec_thresh") || !strings.Contains(msg, "0.09") || !strings.Contains(msg, "0.25") {
		t.Errorf("error does not name medium_rec_thresh and both values: %v", err)
	}
}

// TestUnidecodeCheckAgreesWithCommittedManifest reads
// corpus/manifest.json (read-only) and confirms its pinned unidecode
// field agrees with the engine's own embedded table version.
func TestUnidecodeCheckAgreesWithCommittedManifest(t *testing.T) {
	root := repoRoot(t)
	manifest, err := cassette.LoadManifest(filepath.Join(root, "corpus", "manifest.json"))
	if err != nil {
		t.Fatalf("load corpus/manifest.json: %v", err)
	}
	if err := CheckUnidecodeVersion(*manifest); err != nil {
		t.Error(err)
	}
}

// TestUnidecodeCheckCatchesStampMismatch proves the pin check fires
// and names both versions, against a hand-authored stale stamp.
func TestUnidecodeCheckCatchesStampMismatch(t *testing.T) {
	err := CheckUnidecodeVersion(cassette.Manifest{Unidecode: "1.3.9"})
	if err == nil {
		t.Fatal("expected the unidecode pin check to fire on a mismatched stamp")
	}
	msg := err.Error()
	if !strings.Contains(msg, "1.3.9") || !strings.Contains(msg, matcher.UnidecodeVersion) {
		t.Errorf("error does not name both versions: %v", err)
	}
}

// TestItemFieldsProjectionAgreesWithCorpus proves matcher.ProjectItemFields'
// typed-default projection (contract 10) agrees with real corpus data,
// not only with Task 1's own hand-authored fixture: it reads a
// committed case's first item, picks a field that item omits, and
// asserts the projection reads that field as beets' typed default
// rather than as an absent value.
func TestItemFieldsProjectionAgreesWithCorpus(t *testing.T) {
	root := repoRoot(t)
	dirs := corpusCaseDirs(t, root)
	c, err := cassette.LoadCase(dirs[0])
	if err != nil {
		t.Fatalf("load %s: %v", dirs[0], err)
	}
	if len(c.Input.Items) == 0 {
		t.Fatalf("case %s: no items", c.ID)
	}
	item := c.Input.Items[0]
	if _, present := item["barcode"]; present {
		t.Fatalf("case %s: item[0] unexpectedly carries barcode; the fixture case or field needs to change", c.ID)
	}

	got := matcher.ProjectItemFields(item)
	if got.Barcode != "" {
		t.Errorf("case %s: item[0]: barcode is omitted and projected to %q, want the beets typed default \"\"", c.ID, got.Barcode)
	}
}

// TestPatternTableNamesResolve is the m11 fold: analysis.DefectFree's
// PatternTable field carries one of the string literals
// analysis.LegalPatternTables names (that package must not import
// internal/engine), and nothing checks that they still resolve to real
// tables in package matcher. This test does, in CheckUnidecodeVersion's
// shape: the harness reads both sides, so neither a rename in matcher
// nor an edit to the ledger's own legal set can orphan the other
// silently.
func TestPatternTableNamesResolve(t *testing.T) {
	tables := map[string][]matcher.SDPattern{
		"SDPatternsNoBracketDiscount": matcher.SDPatternsNoBracketDiscount,
	}
	for _, name := range analysis.LegalPatternTables {
		table, ok := tables[name]
		if !ok {
			t.Errorf("legal pattern table %q names no table in package matcher", name)
			continue
		}
		if len(table) == 0 {
			t.Errorf("pattern table %q resolves to an empty []matcher.SDPattern", name)
		}
	}
}

// TestQuietImportTunablesMatchesTheEstateProfile is the drift guard
// between two live definitions of one configuration: gradeTunables must
// stay profile-driven, because the stock profile overlays nothing
// (translate.go:205-206), while matcher.QuietImportTunables is the
// engine-side constructor rung 4's shadow needs and may not reach
// across the arch wall for it. Two definitions of one configuration
// is the drift internal/engine/matcher/info.go:148-150 warns about;
// this one-line assertion is what fires the day either moves.
func TestQuietImportTunablesMatchesTheEstateProfile(t *testing.T) {
	want, err := gradeTunables("estate")
	if err != nil {
		t.Fatalf("gradeTunables(\"estate\") = %v", err)
	}
	if got := matcher.QuietImportTunables(); !reflect.DeepEqual(got, want) {
		t.Error("matcher.QuietImportTunables and gradeTunables(\"estate\") have drifted")
	}
}
