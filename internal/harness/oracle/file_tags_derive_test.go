//go:build oracle

package oracle

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// loadFileTagSubjects reads analysis.FileTagSubjects' twelve named
// cases off the committed corpus and projects each one's own item, at
// the subject's recorded index, out of its frozen
// Expected.AppliedFields. The subject set and which case resolves to
// which committed row is [analysis.FileTagSubjects]' own concern, not
// re-decided here; this only turns that list into DeriveFileTags'
// input shape. The cases the subjects were projected from come back
// alongside them, keyed by id, so a caller checking a derived ledger
// against the corpus does not read the same twelve cases twice.
func loadFileTagSubjects(t *testing.T, root string) ([]FileTagSubject, map[string]*cassette.Case) {
	t.Helper()
	subjects := make([]FileTagSubject, 0, len(analysis.FileTagSubjects))
	cases := make(map[string]*cassette.Case, len(analysis.FileTagSubjects))
	for _, s := range analysis.FileTagSubjects {
		dir := filepath.Join(root, "corpus", "cases", s.CaseID)
		c, err := cassette.LoadCase(dir)
		if err != nil {
			t.Fatalf("LoadCase(%s): %v", dir, err)
		}
		if c.Expected == nil {
			t.Fatalf("case %s carries no expected record to project a file-tags subject from", s.CaseID)
		}
		raw, ok := c.Expected.AppliedFields[strconv.Itoa(s.ItemIndex)]
		if !ok {
			t.Fatalf("case %s: item index %d does not resolve into its own applied_fields", s.CaseID, s.ItemIndex)
		}
		fields, ok := raw.(map[string]any)
		if !ok {
			t.Fatalf("case %s: item index %d's applied_fields entry is %T, want an object", s.CaseID, s.ItemIndex, raw)
		}
		subjects = append(subjects, FileTagSubject{
			CaseID:        s.CaseID,
			ItemIndex:     s.ItemIndex,
			AppliedFields: fields,
		})
		cases[s.CaseID] = c
	}
	return subjects, cases
}

// TestFileTagsLedgerMatchesTheCorpus re-derives the file-level
// applied-tag ledger through derive_file_tags.py and diffs the result
// against corpus/analysis/file_tags.json, matching
// TestComponentTableReDerivesToItsCommittedTable's own shape: the gate
// checks the ledger's shape and coverage only, since measuring the
// actual write/read-back values needs beets.
func TestFileTagsLedgerMatchesTheCorpus(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	path := filepath.Join(root, "corpus", "analysis", "file_tags.json")

	subjects, cases := loadFileTagSubjects(t, root)
	raw, err := DeriveFileTags(context.Background(), opts, subjects)
	if err != nil {
		t.Fatalf("DeriveFileTags: %v", err)
	}
	var derived analysis.FileTagLedger
	if err := json.Unmarshal(raw, &derived); err != nil {
		t.Fatalf("decode derived ledger: %v", err)
	}

	if err := analysis.CheckFileTagLedger(cases, &derived); err != nil {
		t.Fatalf("the freshly derived ledger does not satisfy its own schema:\n%v", err)
	}

	if *updateLedgers {
		data, err := json.MarshalIndent(derived, "", "  ")
		if err != nil {
			t.Fatalf("encode file_tags.json: %v", err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			t.Fatalf("write file_tags.json: %v", err)
		}
		t.Logf("rewrote %s with %d rows", path, len(derived.Rows))
		return
	}

	committed, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed ledger: %v", err)
	}
	// Compared as decoded values, not as bytes: the committed file is
	// indented for review while the shim prints one line, and both sides
	// decode to the same doubles either way.
	if got, want := decodeAny(t, raw), decodeAny(t, committed); !reflect.DeepEqual(got, want) {
		t.Errorf("re-derivation diverges from corpus/analysis/file_tags.json; re-run with -update-ledgers to see the new ledger")
	}
}

// TestFileTagsLedgerReDerivationCatchesAPerturbation is the fixture
// with real teeth for the oracle-tagged half. It replays
// TestFileTagsLedgerMatchesTheCorpus's own comparison (a fresh
// derivation, decoded, against the committed file, decoded) but with
// one value in a copy of the committed bytes perturbed first, proving
// that comparison would actually catch a committed ledger that
// diverged from the corpus by a single value, rather than vacuously
// accepting whatever the shim happens to produce.
func TestFileTagsLedgerReDerivationCatchesAPerturbation(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	path := filepath.Join(root, "corpus", "analysis", "file_tags.json")

	subjects, _ := loadFileTagSubjects(t, root)
	raw, err := DeriveFileTags(context.Background(), opts, subjects)
	if err != nil {
		t.Fatalf("DeriveFileTags: %v", err)
	}

	committed, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed ledger: %v", err)
	}
	var perturbed map[string]any
	if err := json.Unmarshal(committed, &perturbed); err != nil {
		t.Fatalf("decode committed ledger: %v", err)
	}
	rows, ok := perturbed["rows"].([]any)
	if !ok || len(rows) == 0 {
		t.Fatalf("committed ledger carries no rows to perturb")
	}
	row, ok := rows[0].(map[string]any)
	if !ok {
		t.Fatalf("rows[0] is %T, want an object", rows[0])
	}
	values, ok := row["values"].(map[string]any)
	if !ok {
		t.Fatalf("rows[0].values is %T, want an object", row["values"])
	}
	values["title"] = fmt.Sprintf("perturbed-%v", values["title"])

	perturbedBytes, err := json.Marshal(perturbed)
	if err != nil {
		t.Fatalf("encode perturbed ledger: %v", err)
	}

	if got, want := decodeAny(t, raw), decodeAny(t, perturbedBytes); reflect.DeepEqual(got, want) {
		t.Fatal("a fresh derivation compared equal to a committed ledger perturbed by one value; the re-derive-and-diff assertion has no teeth")
	}
}
