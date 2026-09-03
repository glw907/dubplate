//go:build oracle

package oracle

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/glw907/dubplate/internal/harness/analysis"
)

// TestComponentTableReDerivesToItsCommittedTable is the assertion the
// gate cannot make: it re-runs derive_component.py and diffs the result
// against corpus/analysis/components.json. The gate-resident test
// checks that table's schema and coverage only, since the Go matcher a
// row's output will eventually be compared against does not exist until
// rung 3.
func TestComponentTableReDerivesToItsCommittedTable(t *testing.T) {
	opts := venvOptions(t)
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	path := filepath.Join(root, "corpus", "analysis", "components.json")

	raw, err := DeriveComponentTable(context.Background(), opts)
	if err != nil {
		t.Fatalf("DeriveComponentTable: %v", err)
	}
	var derived analysis.ComponentTable
	if err := json.Unmarshal(raw, &derived); err != nil {
		t.Fatalf("decode derived table: %v", err)
	}
	if err := analysis.CheckComponentTable(&derived); err != nil {
		t.Fatalf("the freshly derived table does not satisfy its own schema:\n%v", err)
	}

	if *updateLedgers {
		data, err := json.MarshalIndent(derived, "", "  ")
		if err != nil {
			t.Fatalf("encode components.json: %v", err)
		}
		if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
			t.Fatalf("write components.json: %v", err)
		}
		t.Logf("rewrote %s with %d rows", path, len(derived.Rows))
		return
	}

	committed, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read committed table: %v", err)
	}
	// Compared as decoded values, not as bytes: the committed file is
	// indented for review while the shim prints one line, and both sides
	// decode to the same doubles either way.
	if got, want := decodeAny(t, raw), decodeAny(t, committed); !reflect.DeepEqual(got, want) {
		t.Errorf("re-derivation diverges from corpus/analysis/components.json; re-run with -update-ledgers to see the new table")
	}
}

func decodeAny(t *testing.T, data []byte) any {
	t.Helper()
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Fatalf("decode table: %v", err)
	}
	return v
}
