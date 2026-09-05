//go:build oracle

package oracle

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// updateLedgers rewrites the derived ledgers under corpus/analysis from
// the current run instead of diffing against them. It is how a case
// author adds a row after promoting a case:
//
//	go test -tags oracle ./internal/harness/oracle -run Ledger -update-ledgers
//
// A tie row's disposition and reason are human adjudication, so an
// update carries them forward rather than overwriting them.
var updateLedgers = flag.Bool("update-ledgers", false, "rewrite corpus/analysis's derived ledgers from this run")

// corpusCaseDirs lists the committed case directories, in the order
// os.ReadDir returns them (sorted by name).
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
		t.Fatal("corpus/cases has no case directories")
	}
	return dirs
}

// TestTieLedgerMatchesTheCorpus re-measures every committed case's
// assignment uniqueness through derive_ties.py and diffs the result
// against corpus/analysis/ties.json. The gate checks the ledger's shape
// and its coverage; only this test can check its numbers, since
// measuring them needs beets.
func TestTieLedgerMatchesTheCorpus(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	var entries []analysis.TieEntry
	for _, dir := range corpusCaseDirs(t, root) {
		c, err := cassette.LoadCase(dir)
		if err != nil {
			t.Fatalf("LoadCase(%s): %v", dir, err)
		}
		entry, err := DeriveTies(context.Background(), c, opts)
		if err != nil {
			t.Fatalf("DeriveTies(%s): %v", c.ID, err)
		}
		entries = append(entries, entry)
	}

	path := filepath.Join(root, "corpus", "analysis", "ties.json")
	if *updateLedgers {
		writeTieLedger(t, path, entries)
		return
	}

	committed, err := analysis.LoadTieLedger(path)
	if err != nil {
		t.Fatalf("load committed ledger: %v", err)
	}
	for _, got := range entries {
		i := slices.IndexFunc(committed.Entries, func(e analysis.TieEntry) bool { return e.CaseID == got.CaseID })
		if i < 0 {
			t.Errorf("ties.json carries no row for %s", got.CaseID)
			continue
		}
		want := committed.Entries[i]
		want.Disposition, want.Reason = got.Disposition, got.Reason
		want.Adjudicator = got.Adjudicator
		if want != got {
			t.Errorf("re-measurement diverges from the committed row for %s:\nwant: %+v\ngot:  %+v", got.CaseID, want, got)
		}
	}
}

// TestDeriveTiesSkipsARecordWithNoCandidate proves the skip path
// directly, ahead of any family 14 case landing: a case whose derived
// record carries no candidate (the shape a zero-track release leaves
// behind once beets' own _add_candidate drops it) returns a skipped
// row instead of erroring, and never starts the replay stub or
// launches derive_ties.py to get there. It carries no oracle build tag
// dependency of its own, but lives in this file because it is the
// tie driver's own skip-path proof.
func TestDeriveTiesSkipsARecordWithNoCandidate(t *testing.T) {
	c := &cassette.Case{
		ID:       "no-candidate-test",
		Expected: &cassette.Expected{Candidates: nil},
	}

	entry, err := DeriveTies(context.Background(), c, Options{})
	if err != nil {
		t.Fatalf("DeriveTies: %v", err)
	}
	if entry.CaseID != c.ID {
		t.Errorf("case_id = %q, want %q", entry.CaseID, c.ID)
	}
	if entry.Status != analysis.TieSkipped {
		t.Errorf("status = %q, want %q", entry.Status, analysis.TieSkipped)
	}
	if entry.AlbumID != "" {
		t.Errorf("album_id = %q, want empty on a skipped row", entry.AlbumID)
	}
	if entry.Reason == "" {
		t.Error("skipped row carries no reason")
	}
}

// writeTieLedger rewrites path from entries, carrying every committed
// disposition and reason forward: those are the human half of the
// ledger and a re-measurement has nothing to say about them.
func writeTieLedger(t *testing.T, path string, entries []analysis.TieEntry) {
	t.Helper()
	if prior, err := analysis.LoadTieLedger(path); err == nil {
		entries = analysis.MergeTieDispositions(entries, prior)
	}

	data, err := json.MarshalIndent(analysis.TieLedger{Entries: entries}, "", "  ")
	if err != nil {
		t.Fatalf("encode ties.json: %v", err)
	}
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write ties.json: %v", err)
	}
	t.Logf("rewrote %s with %d rows", path, len(entries))
}
