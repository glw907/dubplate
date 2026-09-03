//go:build oracle

// corpus_derive_test.go proves the frozen corpus is re-derivable from
// the tree as checked in, not merely internally consistent: it re-runs
// the real oracle against every committed case and against the real
// year-penalty case under a second eval date. `make derive` runs it;
// the repository gate does not build this file at all.
package oracle

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// TestCorpusReDerivesToItsCommittedRecords runs Derive against every
// case under corpus/cases whose Expected is committed and fails on any
// divergence from that record. This is a deep comparison of the
// decoded cassette.Expected record, not the shim's raw stdout: a case
// directory stores only the decoded record in case.json, never the raw
// bytes the shim originally printed, so there is nothing at the
// raw-stdout level to diff a re-derivation against here.
// TestShimStdoutIsByteIdentical and corpus/README.md's second full
// derivation of the smoke corpus already prove the shim itself is
// byte-deterministic process to process; this test proves the
// committed decode still matches what today's pinned venv produces.
//
// Each case's own EvalDate drives its derivation (every smoke-corpus
// case is pinned to its 2026-09-02 capture date, per corpus/README.md),
// so a divergence here means either the venv drifted from
// corpus/manifest.json's pins or a case's committed record is stale,
// not that "today" moved.
func TestCorpusReDerivesToItsCommittedRecords(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	casesRoot := filepath.Join(root, "corpus", "cases")
	entries, err := os.ReadDir(casesRoot)
	if err != nil {
		t.Fatalf("read corpus/cases: %v", err)
	}

	dirs := 0
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dirs++
		id := e.Name()
		t.Run(id, func(t *testing.T) {
			dir := filepath.Join(casesRoot, id)
			c, err := cassette.LoadCase(dir)
			if err != nil {
				t.Fatalf("LoadCase(%s): %v", dir, err)
			}
			if c.Expected == nil {
				t.Skip("case carries no committed expected record; nothing to re-derive against")
			}

			got, err := Derive(context.Background(), c, opts)
			if err != nil {
				t.Fatalf("re-derive %s: %v", id, err)
			}
			if !reflect.DeepEqual(got, c.Expected) {
				t.Errorf("re-derivation diverges from the committed record for %s:\nwant: %s\ngot:  %s",
					id, mustJSON(t, c.Expected), mustJSON(t, got))
			}
		})
	}
	if dirs == 0 {
		t.Fatal("corpus/cases has no case directories; nothing was re-derived")
	}
}

// TestPinnedDateMovesTheRealYearPenalty is the plan's required proof
// riding the actual smoke corpus: it derives corpus/cases/year-penalty
// under its committed eval date and again under a different one, and
// checks the case's own year-penalty component (not just the aggregate
// distance) moves between them, so the pinned date is causally in the
// verdict on live-captured data. TestPinnedDateMovesTheYearPenalty
// (derive_test.go) proves the same causality against the synthetic
// smoke fixture; this test is its real-corpus counterpart, not a
// replacement, since the synthetic fixture stays useful as a
// venv-independent-of-corpus-content smoke check.
func TestPinnedDateMovesTheRealYearPenalty(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	dir := filepath.Join(root, "corpus", "cases", "year-penalty")

	near, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("load year-penalty case: %v", err)
	}
	nearDate := near.EvalDate
	nearExp, err := Derive(context.Background(), near, opts)
	if err != nil {
		t.Fatalf("derive at %s: %v", nearDate, err)
	}

	far, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("load year-penalty case: %v", err)
	}
	far.EvalDate = "2036-09-02"
	farExp, err := Derive(context.Background(), far, opts)
	if err != nil {
		t.Fatalf("derive at %s: %v", far.EvalDate, err)
	}

	nearYear, ok := nearExp.Candidates[0].Penalties["year"]
	if !ok {
		t.Fatalf("derivation at %s carries no year penalty", nearDate)
	}
	farYear, ok := farExp.Candidates[0].Penalties["year"]
	if !ok {
		t.Fatalf("derivation at %s carries no year penalty", far.EvalDate)
	}
	if nearYear == farYear {
		t.Errorf("year penalty at %s (%v) equals year penalty at %s (%v): the pinned date should be causally in the verdict",
			nearDate, nearYear, far.EvalDate, farYear)
	}
	// beets divides the year penalty's diff by the span from the
	// original year to today: a later pinned date widens that span, so
	// a fixed tag/release mismatch scores a smaller penalty later on.
	if nearYear <= farYear {
		t.Errorf("year penalty at %s (%v) should exceed year penalty at %s (%v): a later eval date widens the year penalty's span denominator",
			nearDate, nearYear, far.EvalDate, farYear)
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatalf("marshal for diff: %v", err)
	}
	return string(data)
}
