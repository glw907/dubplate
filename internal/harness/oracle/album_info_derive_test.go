//go:build oracle

package oracle

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// deriveAlbumInfoLedger re-derives every one of analysis.AlbumInfoSubjects'
// nineteen cases, one process each, and concatenates their rows into a
// single ledger in subject order. The two header fields are the shim's,
// so they must agree across every process: a run whose subjects
// disagreed would be a ledger half-derived under two shims.
//
// The cases the rows were derived from come back alongside the ledger,
// keyed by id, so a caller checking the derivation against the corpus
// does not load the same nineteen cases twice.
func deriveAlbumInfoLedger(t *testing.T, opts Options, root string) (*analysis.AlbumInfoLedger, map[string]*cassette.Case) {
	t.Helper()
	ledger := &analysis.AlbumInfoLedger{}
	cases := make(map[string]*cassette.Case, len(analysis.AlbumInfoSubjects))
	for _, s := range analysis.AlbumInfoSubjects {
		dir := filepath.Join(root, "corpus", "cases", s.CaseID)
		c, err := cassette.LoadCase(dir)
		if err != nil {
			t.Fatalf("LoadCase(%s): %v", dir, err)
		}
		cases[s.CaseID] = c

		raw, err := DeriveAlbumInfo(context.Background(), c, opts)
		if err != nil {
			t.Fatalf("DeriveAlbumInfo(%s): %v", s.CaseID, err)
		}
		// DisallowUnknownFields, not a plain Unmarshal: a field the shim
		// renames would otherwise vanish into a silently dropped key and
		// the ledger would ship without it.
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.DisallowUnknownFields()
		var derived analysis.AlbumInfoLedger
		if err := dec.Decode(&derived); err != nil {
			t.Fatalf("decode %s's derivation: %v", s.CaseID, err)
		}

		if ledger.ShimVersion == "" {
			ledger.ShimVersion, ledger.BeetsVersion = derived.ShimVersion, derived.BeetsVersion
		}
		if derived.ShimVersion != ledger.ShimVersion || derived.BeetsVersion != ledger.BeetsVersion {
			t.Fatalf("%s derived under shim %q / beets %q, the run's earlier subjects under %q / %q",
				s.CaseID, derived.ShimVersion, derived.BeetsVersion, ledger.ShimVersion, ledger.BeetsVersion)
		}
		ledger.Rows = append(ledger.Rows, derived.Rows...)
	}
	return ledger, cases
}

// TestAlbumInfoLedgerMatchesTheCorpus re-derives the AlbumInfo ledger
// through derive_album_info.py and diffs the result against
// corpus/analysis/album_info.json, matching
// TestFileTagsLedgerMatchesTheCorpus's own shape: the gate checks the
// ledger's schema and coverage, and only this test can check the
// translated values, since building them needs beets.
func TestAlbumInfoLedgerMatchesTheCorpus(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	path := filepath.Join(root, "corpus", "analysis", "album_info.json")

	derived, cases := deriveAlbumInfoLedger(t, opts, root)
	if err := analysis.CheckAlbumInfoLedger(cases, derived); err != nil {
		t.Fatalf("the freshly derived ledger does not satisfy its own schema:\n%v", err)
	}

	data, err := json.MarshalIndent(derived, "", "  ")
	if err != nil {
		t.Fatalf("encode album_info.json: %v", err)
	}
	data = append(data, '\n')

	if *updateLedgers {
		if err := os.WriteFile(path, data, 0o644); err != nil {
			t.Fatalf("write album_info.json: %v", err)
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
	if got, want := decodeAny(t, data), decodeAny(t, committed); !reflect.DeepEqual(got, want) {
		t.Errorf("re-derivation diverges from corpus/analysis/album_info.json; re-run with -update-ledgers to see the new ledger")
	}
}

// TestAlbumInfoLedgerReDerivationCatchesAPerturbation is the fixture
// with real teeth for the oracle-tagged half. It replays
// TestAlbumInfoLedgerMatchesTheCorpus's own comparison but with one
// album title in a copy of the committed bytes perturbed first, proving
// that comparison would actually catch a committed ledger that diverged
// from beets by a single value, rather than vacuously accepting whatever
// the shim happens to produce.
func TestAlbumInfoLedgerReDerivationCatchesAPerturbation(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	derived, _ := deriveAlbumInfoLedger(t, opts, root)
	data, err := json.MarshalIndent(derived, "", "  ")
	if err != nil {
		t.Fatalf("encode derived ledger: %v", err)
	}

	committed, err := os.ReadFile(filepath.Join(root, "corpus", "analysis", "album_info.json"))
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
	album, ok := row["album"].(map[string]any)
	if !ok {
		t.Fatalf("rows[0].album is %T, want an object", row["album"])
	}
	album["album"] = fmt.Sprintf("perturbed-%v", album["album"])

	perturbedBytes, err := json.Marshal(perturbed)
	if err != nil {
		t.Fatalf("encode perturbed ledger: %v", err)
	}

	if got, want := decodeAny(t, data), decodeAny(t, perturbedBytes); reflect.DeepEqual(got, want) {
		t.Fatal("a fresh derivation compared equal to a committed ledger perturbed by one value; the re-derive-and-diff assertion has no teeth")
	}
}
