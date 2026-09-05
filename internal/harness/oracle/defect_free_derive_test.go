//go:build oracle

package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"math"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// hypothesisPatternTable names the SD_PATTERNS substitution each
// analysis.StringHeuristicHypotheses member's probe cases derive their
// defect-free figure under ([analysis.DefectFree]'s own doc comment):
// the word-boundary-guarded table for bug-ft-pattern's six adv-ft-*
// rows, and the bracket-discount-free table for
// bug-remix-parenthetical's four adv-remix-* rows.
var hypothesisPatternTable = map[string]string{
	"bug-ft-pattern":          "SDPatternsWordBoundary",
	"bug-remix-parenthetical": "SDPatternsNoBracketDiscount",
}

// defectFreeSubjects resolves analysis.StringHeuristicHypotheses' probe
// cases from the real, committed oracle-bug and adjudication ledgers
// into the ten DefectFreeSubjects this pass populates. pattern names
// the substitution: an empty string asks for the both-ways check's
// baseline half, and the named table asks for the defect-free figure.
func defectFreeSubjects(t *testing.T, root string, pattern func(hypothesis string) string) []DefectFreeSubject {
	t.Helper()
	bugs, err := analysis.LoadOracleBugs(filepath.Join(root, "corpus", "analysis", "oracle_bugs.json"))
	if err != nil {
		t.Fatalf("LoadOracleBugs: %v", err)
	}
	adj, err := analysis.LoadAdjudications(filepath.Join(root, "corpus", "analysis", "adjudications.json"))
	if err != nil {
		t.Fatalf("LoadAdjudications: %v", err)
	}
	adjByCase := make(map[string]analysis.AdjudicationEntry, len(adj.Entries))
	for _, e := range adj.Entries {
		if e.Subject.CaseID != "" {
			adjByCase[e.Subject.CaseID] = e
		}
	}
	bugByID := make(map[string]analysis.OracleBugEntry, len(bugs.Entries))
	for _, e := range bugs.Entries {
		bugByID[e.ID] = e
	}

	var subjects []DefectFreeSubject
	for _, hyp := range analysis.StringHeuristicHypotheses {
		entry, ok := bugByID[hyp]
		if !ok {
			t.Fatalf("oracle_bugs.json carries no entry for %s", hyp)
		}
		for _, caseID := range entry.ProbeCaseIDs {
			c, err := cassette.LoadCase(filepath.Join(root, "corpus", "cases", caseID))
			if err != nil {
				t.Fatalf("LoadCase(%s): %v", caseID, err)
			}
			if c.Expected == nil || len(c.Expected.Candidates) == 0 {
				t.Fatalf("%s carries no committed candidate to resolve a release id from", caseID)
			}
			if len(c.Requests) == 0 {
				t.Fatalf("%s logs no request to resolve a response file from", caseID)
			}
			row, ok := adjByCase[caseID]
			if !ok {
				t.Fatalf("adjudications.json carries no row for probe case %s", caseID)
			}
			subjects = append(subjects, DefectFreeSubject{
				CaseID:         caseID,
				AdjudicationID: row.ID,
				ReleaseMBID:    c.Expected.Candidates[0].AlbumID,
				ResponsePath:   c.Requests[0].Response,
				EvalDate:       c.EvalDate,
				PatternTable:   pattern(hyp),
			})
		}
	}
	return subjects
}

// deriveDefectFreeRows runs DeriveDefectFree over subjects and decodes
// its document, rejecting any key the row type has no field for.
func deriveDefectFreeRows(t *testing.T, opts Options, subjects []DefectFreeSubject) []DefectFreeRow {
	t.Helper()
	raw, err := DeriveDefectFree(context.Background(), opts, subjects)
	if err != nil {
		t.Fatalf("DeriveDefectFree: %v", err)
	}
	var doc struct {
		ShimVersion  string          `json:"shim_version"`
		BeetsVersion string          `json:"beets_version"`
		Rows         []DefectFreeRow `json:"rows"`
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&doc); err != nil {
		t.Fatalf("decode DeriveDefectFree output: %v", err)
	}
	if doc.ShimVersion == "" || doc.BeetsVersion == "" {
		t.Fatal("DeriveDefectFree document carries no shim_version or beets_version")
	}
	return doc.Rows
}

// TestDefectFreeReproducesCommittedDistanceBitIdentically is the
// both-ways check's baseline half (R20, M-a): run with SD_PATTERNS
// untouched over all ten string-heuristic subjects, and every row's
// distance must equal the case's own committed candidate distance,
// bit-identically. A script that cannot reproduce the committed number
// has not proved its counterfactual half, so the defect-free half below
// is only asserted once this passes.
func TestDefectFreeReproducesCommittedDistanceBitIdentically(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	subjects := defectFreeSubjects(t, root, func(string) string { return "" })
	rows := deriveDefectFreeRows(t, opts, subjects)
	if len(rows) != 10 {
		t.Fatalf("derived %d baseline rows, want 10", len(rows))
	}

	for _, row := range rows {
		c, err := cassette.LoadCase(filepath.Join(root, "corpus", "cases", row.CaseID))
		if err != nil {
			t.Fatalf("LoadCase(%s): %v", row.CaseID, err)
		}
		want := c.Expected.Candidates[0].Distance
		if row.Distance != want {
			t.Errorf("%s: baseline distance %v, want case.json's committed %v (bit-identical)", row.CaseID, row.Distance, want)
		}
		t.Logf("%s: baseline %v vs committed %v", row.CaseID, row.Distance, want)
	}
}

// TestDefectFreeMatchesTheFirstPopulation is this task's acceptance
// line for the ten rows: every derived defect-free distance,
// recommendation, action, and pattern_table must equal the block
// corpus/analysis/adjudications.json commits for that row, and the test
// reports each pair (R13's one-time numeric half).
func TestDefectFreeMatchesTheFirstPopulation(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	subjects := defectFreeSubjects(t, root, func(hyp string) string { return hypothesisPatternTable[hyp] })
	rows := deriveDefectFreeRows(t, opts, subjects)
	if len(rows) != 10 {
		t.Fatalf("derived %d defect-free rows, want 10", len(rows))
	}

	adj, err := analysis.LoadAdjudications(filepath.Join(root, "corpus", "analysis", "adjudications.json"))
	if err != nil {
		t.Fatalf("LoadAdjudications: %v", err)
	}
	byCase := make(map[string]analysis.AdjudicationEntry, len(adj.Entries))
	for _, e := range adj.Entries {
		if e.Subject.CaseID != "" {
			byCase[e.Subject.CaseID] = e
		}
	}

	for _, row := range rows {
		entry, ok := byCase[row.CaseID]
		if !ok || entry.DefectFree == nil {
			t.Fatalf("adjudications.json carries no defect_free block for probe case %s", row.CaseID)
		}
		want := entry.DefectFree
		t.Logf("%s: derived distance %v vs committed %v", row.CaseID, row.Distance, want.Distance)
		if row.Distance != want.Distance {
			t.Errorf("%s: defect-free distance %v, want committed %v (bit-identical)", row.CaseID, row.Distance, want.Distance)
		}
		if row.Recommendation != want.Recommendation {
			t.Errorf("%s: recommendation %q, want committed %q", row.CaseID, row.Recommendation, want.Recommendation)
		}
		if row.Action != want.Action {
			t.Errorf("%s: action %q, want committed %q", row.CaseID, row.Action, want.Action)
		}
		if row.PatternTable != want.PatternTable {
			t.Errorf("%s: pattern_table %q, want committed %q", row.CaseID, row.PatternTable, want.PatternTable)
		}

		// The substitution changes track_title distances, not the
		// solved assignment: verify that rather than assume it, since
		// a mapping change under the substituted table would be a
		// finding, not a formality.
		c, err := cassette.LoadCase(filepath.Join(root, "corpus", "cases", row.CaseID))
		if err != nil {
			t.Fatalf("LoadCase(%s): %v", row.CaseID, err)
		}
		if c.Expected == nil {
			t.Fatalf("%s carries no committed oracle record to compare the re-solved mapping against", row.CaseID)
		}
		gotMapping := make(map[int]int, len(row.Mapping))
		for itemStr, track := range row.Mapping {
			item, err := strconv.Atoi(itemStr)
			if err != nil {
				t.Fatalf("%s: mapping key %q is not an item index: %v", row.CaseID, itemStr, err)
			}
			gotMapping[item] = track
		}
		if len(gotMapping) != len(c.Expected.Assignment) {
			t.Errorf("%s: re-solved mapping %v has %d pairs, want %d matching case.json's committed assignment %v", row.CaseID, gotMapping, len(gotMapping), len(c.Expected.Assignment), c.Expected.Assignment)
		}
		for item, wantTrack := range c.Expected.Assignment {
			gotTrack, ok := gotMapping[item]
			if !ok {
				t.Errorf("%s: re-solved mapping carries no item %d, want track %d (case.json's committed assignment)", row.CaseID, item, wantTrack)
				continue
			}
			if gotTrack != wantTrack {
				t.Errorf("%s: re-solved mapping maps item %d to track %d, want case.json's committed track %d", row.CaseID, item, gotTrack, wantTrack)
			}
		}
	}
}

// TestDefectFreeReDeriveCatchesAPerturbation is named mutation 7: one
// committed defect_free.distance perturbed by one ULP must make the
// re-derive diff fire, proving the comparison above has teeth rather
// than vacuously passing whatever the shim happens to produce.
func TestDefectFreeReDeriveCatchesAPerturbation(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	subjects := defectFreeSubjects(t, root, func(hyp string) string { return hypothesisPatternTable[hyp] })
	rows := deriveDefectFreeRows(t, opts, subjects)
	if len(rows) == 0 {
		t.Fatal("derived no rows to compare")
	}
	fresh := rows[0]

	adj, err := analysis.LoadAdjudications(filepath.Join(root, "corpus", "analysis", "adjudications.json"))
	if err != nil {
		t.Fatalf("LoadAdjudications: %v", err)
	}
	var committed float64
	found := false
	for _, e := range adj.Entries {
		if e.Subject.CaseID == fresh.CaseID && e.DefectFree != nil {
			committed = e.DefectFree.Distance
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("adjudications.json carries no defect_free block for %s", fresh.CaseID)
	}
	if fresh.Distance != committed {
		t.Fatalf("fresh derivation (%v) already diverges from the committed distance (%v) before any perturbation", fresh.Distance, committed)
	}

	perturbed := math.Nextafter(committed, committed+1)
	if perturbed == committed {
		t.Fatal("math.Nextafter produced no perturbation; the fixture proves nothing")
	}
	if fresh.Distance == perturbed {
		t.Fatalf("a perturbed committed distance (%v) compared equal to the fresh derivation (%v); the re-derive-and-diff assertion has no teeth", perturbed, fresh.Distance)
	}
}

// TestDefectFreeReDeriveCatchesTheMonkeypatchRemoved is named mutation
// 8: with the SD_PATTERNS substitution removed (every subject asks for
// the untouched table, as the both-ways check's baseline half does),
// every derived distance equals its own baseline rather than its
// committed defect-free figure, so the re-derive diff fires on all ten
// rows.
func TestDefectFreeReDeriveCatchesTheMonkeypatchRemoved(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	subjects := defectFreeSubjects(t, root, func(string) string { return "" })
	rows := deriveDefectFreeRows(t, opts, subjects)
	if len(rows) != 10 {
		t.Fatalf("derived %d rows, want 10", len(rows))
	}

	adj, err := analysis.LoadAdjudications(filepath.Join(root, "corpus", "analysis", "adjudications.json"))
	if err != nil {
		t.Fatalf("LoadAdjudications: %v", err)
	}
	byCase := make(map[string]analysis.AdjudicationEntry, len(adj.Entries))
	for _, e := range adj.Entries {
		if e.Subject.CaseID != "" {
			byCase[e.Subject.CaseID] = e
		}
	}

	diverged := 0
	for _, row := range rows {
		entry, ok := byCase[row.CaseID]
		if !ok || entry.DefectFree == nil {
			t.Fatalf("adjudications.json carries no defect_free block for probe case %s", row.CaseID)
		}
		if row.Distance != entry.DefectFree.Distance {
			diverged++
		}
	}
	if diverged != 10 {
		t.Errorf("removing the monkeypatch diverged from the committed defect_free figures on %d of 10 rows, want 10", diverged)
	}
}
