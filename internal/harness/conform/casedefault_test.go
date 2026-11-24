package conform

import (
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/analysis"
)

// defaultBranchCase is a committed case no adjudication row or oracle
// bug names, so default-mode grading compares it against its own
// committed verdict.
const defaultBranchCase = "admit-duplicate-candidate"

// copyCorpusCase copies the committed case id into a fresh directory
// named for it, applies edit to its case.json as a generic map, and
// returns the copy's path. A nil edit leaves case.json as committed.
func copyCorpusCase(t *testing.T, id string, edit func(map[string]any)) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), id)
	if err := os.CopyFS(dir, os.DirFS(filepath.Join(repoRoot(t), "corpus", "cases", id))); err != nil {
		t.Fatalf("copy case %s: %v", id, err)
	}
	if edit == nil {
		return dir
	}
	path := filepath.Join(dir, "case.json")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	edit(doc)
	if data, err = json.Marshal(doc); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
	return dir
}

// expectedBlock returns case.json's expected object from doc.
func expectedBlock(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	exp, ok := doc["expected"].(map[string]any)
	if !ok {
		t.Fatalf("case.json carries no expected object")
	}
	return exp
}

func committedLedgers(t *testing.T) ledgers {
	t.Helper()
	led, err := loadLedgers(repoRoot(t))
	if err != nil {
		t.Fatalf("loadLedgers: %v", err)
	}
	return led
}

// TestDefaultGradingGradesTheCommittedCase is the baseline the other
// tests in this file perturb: the committed case grades with no error
// and its verdict stays where its record put it.
func TestDefaultGradingGradesTheCommittedCase(t *testing.T) {
	dir := copyCorpusCase(t, defaultBranchCase, nil)
	res, err := gradeCaseDefault(t.Context(), dir, committedLedgers(t))
	if err != nil {
		t.Fatalf("gradeCaseDefault = %v, want nil", err)
	}
	if !res.graded || res.pending || res.ungraded || res.verdictMoved {
		t.Errorf("result graded=%v pending=%v ungraded=%v verdictMoved=%v, want graded alone",
			res.graded, res.pending, res.ungraded, res.verdictMoved)
	}
}

// TestDefaultGradingReportsEachBrokenInput breaks one input of a
// committed case at a time and expects the error each break names.
func TestDefaultGradingReportsEachBrokenInput(t *testing.T) {
	const lookupMBID = "64d79b35-57f2-409d-a5d3-cdd7dc21d5a0"
	const unloggedMBID = "00000000-0000-4000-8000-000000000000"

	tests := []struct {
		name    string
		dir     func(t *testing.T) string
		wantErr string
	}{
		{
			name:    "a directory with no case",
			dir:     func(t *testing.T) string { return t.TempDir() },
			wantErr: ": load: ",
		},
		{
			name: "a case with no derived record",
			dir: func(t *testing.T) string {
				return copyCorpusCase(t, defaultBranchCase, func(doc map[string]any) { delete(doc, "expected") })
			},
			wantErr: "case admit-duplicate-candidate: carries no derived record",
		},
		{
			name: "an eval date that does not parse",
			dir: func(t *testing.T) string {
				return copyCorpusCase(t, defaultBranchCase, func(doc map[string]any) { doc["eval_date"] = "03/09/2026" })
			},
			wantErr: `case admit-duplicate-candidate: eval_date "03/09/2026"`,
		},
		{
			name: "a committed candidate the log never fetched",
			dir: func(t *testing.T) string {
				return copyCorpusCase(t, defaultBranchCase, func(doc map[string]any) {
					exp := expectedBlock(t, doc)
					cands, _ := exp["candidates"].([]any)
					first, _ := cands[0].(map[string]any)
					extra := maps.Clone(first)
					extra["album_id"] = unloggedMBID
					exp["candidates"] = append(cands, extra)
				})
			},
			wantErr: "case admit-duplicate-candidate: re-score: committed candidate " + unloggedMBID + " not found in the case's own log",
		},
		{
			name: "a logged lookup whose body does not parse",
			dir: func(t *testing.T) string {
				dir := copyCorpusCase(t, defaultBranchCase, nil)
				if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), []byte("{"), 0o644); err != nil {
					t.Fatal(err)
				}
				return dir
			},
			wantErr: "case admit-duplicate-candidate: re-score: look up " + lookupMBID + ": ",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := gradeCaseDefault(t.Context(), tc.dir(t), committedLedgers(t))
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("gradeCaseDefault = %v, want an error containing %q", err, tc.wantErr)
			}
		})
	}
}

// TestDefaultGradingReportsAMovedVerdict changes the committed
// recommendation, then the committed action, and expects each change to
// read as a moved verdict that fails nothing.
func TestDefaultGradingReportsAMovedVerdict(t *testing.T) {
	for _, field := range []string{"recommendation", "action"} {
		t.Run(field, func(t *testing.T) {
			dir := copyCorpusCase(t, defaultBranchCase, func(doc map[string]any) {
				expectedBlock(t, doc)[field] = "moved-" + field
			})
			res, err := gradeCaseDefault(t.Context(), dir, committedLedgers(t))
			if err != nil {
				t.Fatalf("gradeCaseDefault = %v, want nil: a moved verdict fails nothing", err)
			}
			if !res.verdictMoved {
				t.Errorf("verdictMoved = false, want true once the committed %s differs", field)
			}
		})
	}
}

// TestDefaultGradingLeavesAPendingCaseUngraded adds a
// pending-adjudication row for the committed case and expects it
// reported pending before any replay.
func TestDefaultGradingLeavesAPendingCaseUngraded(t *testing.T) {
	led := committedLedgers(t)
	led.adj.Entries = append(led.adj.Entries, analysis.AdjudicationEntry{
		ID:          "adj-" + defaultBranchCase,
		Subject:     analysis.AdjudicationSubject{CaseID: defaultBranchCase},
		Disposition: analysis.Pending,
	})
	res, err := gradeCaseDefault(t.Context(), copyCorpusCase(t, defaultBranchCase, nil), led)
	if err != nil {
		t.Fatalf("gradeCaseDefault = %v, want nil", err)
	}
	if !res.pending || res.graded {
		t.Errorf("pending=%v graded=%v, want pending and not graded", res.pending, res.graded)
	}
}

// TestDefaultGradingRefusesEachDefectFreeMismatch changes one figure of
// adv-remix-bracket's defect_free block at a time and expects the error
// that names it.
func TestDefaultGradingRefusesEachDefectFreeMismatch(t *testing.T) {
	const id = "adv-remix-bracket"
	tests := []struct {
		name    string
		edit    func(*analysis.DefectFree)
		wantErr string
	}{
		{"the distance", func(d *analysis.DefectFree) { d.Distance += 0.5 }, "defect_free.distance"},
		{"the recommendation", func(d *analysis.DefectFree) { d.Recommendation = "moved" }, "case adv-remix-bracket: default-mode recommendation "},
		{"the action", func(d *analysis.DefectFree) { d.Action = "moved" }, "case adv-remix-bracket: default-mode action "},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			led := committedLedgers(t)
			var row *analysis.DefectFree
			for i := range led.adj.Entries {
				if led.adj.Entries[i].Subject.CaseID == id {
					row = led.adj.Entries[i].DefectFree
				}
			}
			if row == nil {
				t.Fatalf("the ledger row for %s carries no defect_free block", id)
			}
			tc.edit(row)
			_, err := gradeCaseDefault(t.Context(), copyCorpusCase(t, id, nil), led)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("gradeCaseDefault = %v, want an error containing %q", err, tc.wantErr)
			}
		})
	}
}
