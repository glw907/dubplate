package conform

import (
	"cmp"
	"encoding/json"
	"maps"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/harness/analysis"
)

// repoRoot walks up from the working directory to the nearest ancestor
// holding go.mod, so corpus/ resolves relative to the module root
// rather than to `go test`'s per-package working directory. This is
// its own copy of internal/harness/corpus_test.go's helper: repoRoot
// there is unexported and test-only, so this package reimplements the
// walk rather than importing anything (plan's primary-inputs note 5).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("locate repo root: no go.mod above %s", dir)
		}
		dir = parent
	}
}

func loadComponentTable(t *testing.T) *analysis.ComponentTable {
	t.Helper()
	path := filepath.Join(repoRoot(t), "corpus", "analysis", "components.json")
	table, err := analysis.LoadComponentTable(path)
	if err != nil {
		t.Fatalf("load %s: %v", path, err)
	}
	return table
}

func loadFixtureTable(t *testing.T, name string) *analysis.ComponentTable {
	t.Helper()
	table, err := analysis.LoadComponentTable(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("load testdata/%s: %v", name, err)
	}
	return table
}

func findRow(t *testing.T, table *analysis.ComponentTable, id string) analysis.ComponentRow {
	t.Helper()
	for _, r := range table.Rows {
		if r.ID == id {
			return r
		}
	}
	t.Fatalf("components.json: no row named %q", id)
	return analysis.ComponentRow{}
}

func mustUnmarshal(t *testing.T, raw json.RawMessage, v any) {
	t.Helper()
	if err := json.Unmarshal(raw, v); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
}

// TestComponentTableReplaysAllRows is tier 0: every committed row in
// corpus/analysis/components.json replays against the matcher and
// agrees with the committed output. The row floor and per-method
// coverage are enforced through analysis.CheckComponentTable, which
// already carries analysis.ComponentRowFloor rather than a second copy
// of 72 (fold, M-g on the coverage half; conform's own mirror check
// runs below with its own method list).
func TestComponentTableReplaysAllRows(t *testing.T) {
	table := loadComponentTable(t)
	if err := analysis.CheckComponentTable(table); err != nil {
		t.Fatalf("corpus/analysis/components.json: %v", err)
	}

	summary, err := ReplayTable(table)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}

	total := 0
	for _, method := range slices.Sorted(maps.Keys(summary.CountByMethod)) {
		count := summary.CountByMethod[method]
		t.Logf("replayed %d rows for method %q", count, method)
		total += count
	}
	if total != summary.TotalRows {
		t.Fatalf("replayed %d of %d committed rows; a row was silently skipped rather than failed", total, summary.TotalRows)
	}

	for _, method := range componentMethods {
		if summary.CountByMethod[method] == 0 {
			t.Errorf("decoder method %q has no replayed row", method)
		}
	}

	if len(summary.NearMisses) == 0 {
		t.Log("near-miss findings: none; every row is bit-identical")
	}
	for _, nm := range summary.NearMisses {
		t.Logf("near-miss finding: %s (%s): %s", nm.RowID, nm.Method, nm.Detail)
	}
}

// TestComponentDecoderRejectsUnknownMethod proves the "unknown method
// is a failure naming the row id, never a skip" rule fires, against a
// hand-authored fixture (never a corpus case): testdata/unknown_method.json
// names a method the decoder does not know.
func TestComponentDecoderRejectsUnknownMethod(t *testing.T) {
	table := loadFixtureTable(t, "unknown_method.json")
	summary, err := ReplayTable(table)
	if err == nil {
		t.Fatal("expected ReplayTable to report the unknown-method row as a failure")
	}
	msg := err.Error()
	if !strings.Contains(msg, "bogus/unknown-method") || !strings.Contains(msg, "add_frobnicate") {
		t.Errorf("error does not name the row id and the unknown method: %v", err)
	}
	if summary.CountByMethod["add_frobnicate"] != 0 {
		t.Error("an unknown method must never be counted as replayed")
	}
}

// TestComponentReplayRecordsULPNearMiss proves contract 12's own
// assertion discipline: a row whose output is perturbed by a single
// ULP fails bit-identity but still agrees within the 1e-12 relative
// fallback, and the replay must record that as a finding rather than
// silently treating it as an exact pass. testdata/ulp_mismatch.json is
// hand-authored: 1.0000000000000002 is math.Nextafter(1, 2), one ULP
// above the true add_expr(true) result of 1.0.
func TestComponentReplayRecordsULPNearMiss(t *testing.T) {
	table := loadFixtureTable(t, "ulp_mismatch.json")
	if len(table.Rows) != 1 {
		t.Fatalf("testdata/ulp_mismatch.json: got %d rows, want 1", len(table.Rows))
	}
	outcome := ReplayRow(table.Rows[0])
	if outcome.Err != nil {
		t.Fatalf("a one-ULP difference is within the 1e-12 fallback and must not fail outright: %v", outcome.Err)
	}
	if !outcome.NearMiss {
		t.Fatal("a one-ULP difference must be recorded as a near-miss finding, not a silent pass")
	}
	t.Logf("near-miss finding: %s: %s", outcome.RowID, outcome.Detail)
}

// TestComponentDecoderCatchesMethodWithNoRows proves "every decoder
// method has at least one replayed row" fires: testdata/method_missing.json
// exercises only two of the eleven known methods, so every other
// method (add_ratio included) has zero replayed rows.
func TestComponentDecoderCatchesMethodWithNoRows(t *testing.T) {
	table := loadFixtureTable(t, "method_missing.json")
	summary, err := ReplayTable(table)
	if err != nil {
		t.Fatalf("replay: %v", err)
	}
	if summary.CountByMethod["add_ratio"] != 0 {
		t.Fatal("fixture unexpectedly exercises add_ratio; the coverage check needs a method with zero rows to prove it fires")
	}
	missing := 0
	for _, method := range componentMethods {
		if summary.CountByMethod[method] == 0 {
			missing++
		}
	}
	if missing == 0 {
		t.Fatal("expected the per-method coverage check to flag at least one method with no replayed rows")
	}
}

// TestComponentTableBelowFloorFails proves the row-floor check fires,
// reusing analysis.ComponentRowFloor rather than a second copy of 72
// (plan's own instruction). The fixture is built in code rather than
// as a static testdata file: floor-1 rows of trivially valid filler
// content is a lot of repetition to hand-author as JSON, and the
// property under test (a row count comparison) needs no realistic row
// shape to exercise.
func TestComponentTableBelowFloorFails(t *testing.T) {
	rows := make([]analysis.ComponentRow, analysis.ComponentRowFloor-1)
	for i := range rows {
		rows[i] = analysis.ComponentRow{
			ID:     "filler/" + strconv.Itoa(i),
			Method: "add_expr",
			Args:   json.RawMessage(`{"key":"expr","calls":[[true]]}`),
			Output: json.RawMessage(`{"penalties":[1]}`),
		}
	}
	table := &analysis.ComponentTable{Beets: "2.13.1", ShimVersion: "1", Rows: rows}
	if err := analysis.CheckComponentTable(table); err == nil {
		t.Fatal("CheckComponentTable accepted a table one row below the floor")
	}
}

// The four arithmetic mutations restated against the rows that pin
// them (fold, M-b). Task 2 proved each against its own hand-authored
// fixtures; these tests prove the same four against the committed
// rows, by computing the wrong arithmetic directly (never by editing
// matcher's production code) and showing it diverges from the row's
// committed output.

// TestArithmeticMutationFixedWeightDenominatorFailsZeroAddedKey proves
// a denominator counting every configured weight instead of only added
// penalties diverges from aggregate/zero-added-key's committed
// max_distance, which is exactly the row the mechanics study's
// denominator rule (contract 5) pins.
func TestArithmeticMutationFixedWeightDenominatorFailsZeroAddedKey(t *testing.T) {
	table := loadComponentTable(t)
	row := findRow(t, table, "aggregate/zero-added-key")
	var args aggregateArgs
	mustUnmarshal(t, row.Args, &args)
	var want aggregateOutput
	mustUnmarshal(t, row.Output, &want)

	tun := tunablesWithOverrides(args.Weights)
	var mutantMax float64
	for _, w := range tun.DistanceWeights {
		mutantMax += w
	}
	if mutantMax == want.MaxDistance {
		t.Fatalf("mutant fixed-weight denominator (%v) coincidentally matches the committed max_distance (%v); %s no longer discriminates this bug", mutantMax, want.MaxDistance, row.ID)
	}
}

// TestArithmeticMutationRawFilterIncludesZeroWeightedKey proves a
// nonzero view filtering on the raw added value, rather than the
// weighted and normalized one, keeps a key whose weight is 0.0 -
// diverging from aggregate/all-zero-weights' committed empty items
// view (port hazard 4).
func TestArithmeticMutationRawFilterIncludesZeroWeightedKey(t *testing.T) {
	table := loadComponentTable(t)
	row := findRow(t, table, "aggregate/all-zero-weights")
	var args aggregateArgs
	mustUnmarshal(t, row.Args, &args)
	var want aggregateOutput
	mustUnmarshal(t, row.Output, &want)
	wantItems, err := decodeItems(want.Items)
	if err != nil {
		t.Fatalf("decode items: %v", err)
	}

	tun := tunablesWithOverrides(args.Weights)
	dist := matcher.NewDistance(tun)
	if err := applyAggregateCalls(dist, args.Calls); err != nil {
		t.Fatalf("replay %s: %v", row.ID, err)
	}

	var mutantItems []keyDist
	for _, key := range dist.AddedKeys() {
		var raw float64
		for _, p := range dist.Penalties(key) {
			raw += p
		}
		if raw != 0 {
			mutantItems = append(mutantItems, keyDist{Key: key, Dist: raw})
		}
	}
	if len(mutantItems) == len(wantItems) {
		t.Fatalf("mutant raw-filtered items %v has the same length as the committed items %v; %s no longer discriminates this bug", mutantItems, wantItems, row.ID)
	}
}

// TestArithmeticMutationWrongSortFailsBothSortedRows proves a sort
// keyed on (distance, key) instead of beets' (-distance, key)
// diverges from both aggregate/sorted-penalties, whose two keys carry
// different weighted distances (pinning the primary direction), and
// aggregate/sorted-penalties-tied, whose two keys tie on distance
// (pinning the key tiebreak direction, the only thing that can
// distinguish them once distance no longer can).
func TestArithmeticMutationWrongSortFailsBothSortedRows(t *testing.T) {
	table := loadComponentTable(t)
	for _, id := range []string{"aggregate/sorted-penalties", "aggregate/sorted-penalties-tied"} {
		row := findRow(t, table, id)
		var args aggregateArgs
		mustUnmarshal(t, row.Args, &args)
		var want aggregateOutput
		mustUnmarshal(t, row.Output, &want)
		wantItems, err := decodeItems(want.Items)
		if err != nil {
			t.Fatalf("decode items: %v", err)
		}

		tun := tunablesWithOverrides(args.Weights)
		dist := matcher.NewDistance(tun)
		if err := applyAggregateCalls(dist, args.Calls); err != nil {
			t.Fatalf("replay %s: %v", row.ID, err)
		}

		mutant := slices.Clone(dist.NonzeroItems())
		slices.SortFunc(mutant, func(a, b matcher.PenaltyItem) int {
			return cmp.Or(cmp.Compare(a.Distance, b.Distance), cmp.Compare(b.Key, a.Key))
		})
		var mutantItems []keyDist
		for _, item := range mutant {
			mutantItems = append(mutantItems, keyDist{Key: item.Key, Dist: item.Distance})
		}

		if itemsEqual(mutantItems, wantItems) {
			t.Fatalf("mutant sort order coincidentally matches the committed items for %s; this fixture no longer discriminates the bug", row.ID)
		}
	}
}

// TestArithmeticMutationUnguardedDivisionProducesNaN proves an
// unguarded raw/max division returns NaN, not 0.0, on both rows whose
// committed max_distance is zero: aggregate/empty (nothing added) and
// aggregate/all-zero-weights (a key added under weight 0.0).
func TestArithmeticMutationUnguardedDivisionProducesNaN(t *testing.T) {
	table := loadComponentTable(t)
	for _, id := range []string{"aggregate/empty", "aggregate/all-zero-weights"} {
		row := findRow(t, table, id)
		var args aggregateArgs
		mustUnmarshal(t, row.Args, &args)
		var want aggregateOutput
		mustUnmarshal(t, row.Output, &want)

		tun := tunablesWithOverrides(args.Weights)
		dist := matcher.NewDistance(tun)
		if err := applyAggregateCalls(dist, args.Calls); err != nil {
			t.Fatalf("replay %s: %v", row.ID, err)
		}

		mutant := dist.RawDistance() / dist.MaxDistance()
		if mutant == want.Distance {
			t.Fatalf("mutant unguarded division (%v) coincidentally matches the committed distance (%v) for %s", mutant, want.Distance, row.ID)
		}
		if !math.IsNaN(mutant) {
			t.Fatalf("expected the unguarded division to produce NaN for %s, got %v", row.ID, mutant)
		}
	}
}
