//go:build oracle

// The tests in this file are the only ones that run the real shim, so
// they are the only ones needing the uv venv. `make derive` runs them;
// the repository gate does not build this file at all.
package oracle

import (
	"bytes"
	"context"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// venvOptions resolves the pinned venv, honoring DUBPLATE_ORACLE_VENV
// for a venv built somewhere other than tools/venv/.venv.
func venvOptions(t *testing.T) Options {
	t.Helper()
	venv := os.Getenv("DUBPLATE_ORACLE_VENV")
	if venv == "" {
		root, err := repoRoot()
		if err != nil {
			t.Fatalf("locate repo root: %v", err)
		}
		venv = filepath.Join(root, "tools", "venv", ".venv")
	}
	if _, err := os.Stat(filepath.Join(venv, "bin", "python3")); err != nil {
		t.Fatalf("no venv at %s: build it per tools/venv/README.md: %v", venv, err)
	}
	return Options{Venv: venv}
}

// smokeCase loads the checked-in fixture case and points caseDir at
// it. The fixture is shaped like a smoke corpus case (one album, one
// release lookup, a live year penalty) but is synthetic: real captures
// arrive with the corpus itself.
func smokeCase(t *testing.T, evalDate string) *cassette.Case {
	t.Helper()
	dir, err := filepath.Abs(filepath.Join("testdata", "cases", "smoke"))
	if err != nil {
		t.Fatalf("resolve fixture path: %v", err)
	}
	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("load fixture case: %v", err)
	}
	c.EvalDate = evalDate

	orig := caseDir
	caseDir = func(*cassette.Case) (string, error) { return dir, nil }
	t.Cleanup(func() { caseDir = orig })
	return c
}

func TestDeriveRunsPinnedBeets(t *testing.T) {
	c := smokeCase(t, "2024-01-15")
	exp, err := Derive(context.Background(), c, venvOptions(t))
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}

	if exp.Recommendation != "strong" {
		t.Errorf("recommendation = %q, want strong", exp.Recommendation)
	}
	if exp.Action != "apply" {
		t.Errorf("action = %q, want apply", exp.Action)
	}
	if len(exp.Candidates) != 1 {
		t.Fatalf("candidates = %d, want 1", len(exp.Candidates))
	}

	cand := exp.Candidates[0]
	if cand.Source != "MusicBrainz" {
		t.Errorf("candidate source = %q, want MusicBrainz", cand.Source)
	}
	if cand.AlbumID != "3b6b0a4e-5e1a-4a52-9c31-3c0c1c1f0001" {
		t.Errorf("candidate album id = %q, want the fixture release id", cand.AlbumID)
	}
	// The fixture's items are tagged 1990 against a 1997 release, so
	// the year penalty is the one nonzero component; every other
	// component matches exactly.
	if got, ok := cand.Penalties["year"]; !ok || got <= 0 {
		t.Errorf("year penalty = %v (present: %t), want a positive weighted distance", got, ok)
	}
	if len(cand.Tracks) != 2 {
		t.Errorf("per-track distances = %d, want 2", len(cand.Tracks))
	}
	if want := (map[int]int{0: 0, 1: 1}); !maps.Equal(exp.Assignment, want) {
		t.Errorf("assignment = %v, want %v", exp.Assignment, want)
	}
	if len(exp.ExtraItems) != 0 || len(exp.ExtraTracks) != 0 {
		t.Errorf("extras = (%v, %v), want both empty", exp.ExtraItems, exp.ExtraTracks)
	}

	// AppliedFields is keyed by the item's index in the case's pinned
	// input order; apply is per item, so a merged view could not
	// express two items disagreeing on a field.
	first, ok := exp.AppliedFields["0"].(map[string]any)
	if !ok {
		t.Fatalf("applied fields for item 0 = %#v, want an object", exp.AppliedFields["0"])
	}
	if first["mb_trackid"] != "3b6b0a4e-5e1a-4a52-9c31-3c0c1c1f0021" {
		t.Errorf("applied mb_trackid = %v, want the fixture's first recording id", first["mb_trackid"])
	}
	if first["year"] != float64(1997) {
		t.Errorf("applied year = %v, want 1997 from the release", first["year"])
	}
}

// TestShimStdoutIsByteIdentical is the determinism proof: two shim
// processes deriving the same case must agree on the exact bytes, not
// merely on a Go-decoded record, so Python-side key ordering or float
// formatting drift can actually fail.
func TestShimStdoutIsByteIdentical(t *testing.T) {
	c := smokeCase(t, "2024-01-15")
	opts := venvOptions(t)

	first, err := deriveRaw(context.Background(), c, opts)
	if err != nil {
		t.Fatalf("first derivation: %v", err)
	}
	second, err := deriveRaw(context.Background(), c, opts)
	if err != nil {
		t.Fatalf("second derivation: %v", err)
	}

	if !bytes.Equal(first, second) {
		t.Errorf("shim stdout differs between processes:\nfirst:  %s\nsecond: %s", first, second)
	}
	if len(bytes.TrimSpace(first)) == 0 {
		t.Fatal("shim stdout is empty; the comparison above proves nothing")
	}
}

// TestPinnedDateMovesTheYearPenalty proves the shim's date pinning is
// load-bearing rather than merely echoed: beets divides the year
// penalty by the span ending at today's year, so a case derived under
// a later date scores differently on identical input.
func TestPinnedDateMovesTheYearPenalty(t *testing.T) {
	opts := venvOptions(t)

	near := smokeCase(t, "2024-01-15")
	nearExp, err := Derive(context.Background(), near, opts)
	if err != nil {
		t.Fatalf("derive at 2024-01-15: %v", err)
	}

	far := smokeCase(t, "2034-01-15")
	farExp, err := Derive(context.Background(), far, opts)
	if err != nil {
		t.Fatalf("derive at 2034-01-15: %v", err)
	}

	if nearExp.Candidates[0].Distance <= farExp.Candidates[0].Distance {
		t.Errorf("distance at 2024 (%v) should exceed distance at 2034 (%v): the year penalty divides by the span to today",
			nearExp.Candidates[0].Distance, farExp.Candidates[0].Distance)
	}
}

func TestReconcileAcceptsTheCheckedInManifest(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	m, err := cassette.LoadManifest(filepath.Join(root, "corpus", "manifest.json"))
	if err != nil {
		t.Fatalf("load manifest: %v", err)
	}
	if err := Reconcile(context.Background(), m, venvOptions(t)); err != nil {
		t.Errorf("Reconcile: %v", err)
	}
}

func TestReconcileNamesTheDisagreeingComponent(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	opts := venvOptions(t)

	tests := []struct {
		name    string
		mutate  func(*cassette.Manifest)
		wantHas string
	}{
		{"wrong version", func(m *cassette.Manifest) { m.Jellyfish = "0.0.1" }, "jellyfish"},
		{"unpinned field", func(m *cassette.Manifest) { m.Lap = "" }, "lap"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := cassette.LoadManifest(filepath.Join(root, "corpus", "manifest.json"))
			if err != nil {
				t.Fatalf("load manifest: %v", err)
			}
			tt.mutate(m)

			err = Reconcile(context.Background(), m, opts)
			if err == nil {
				t.Fatalf("Reconcile accepted a manifest with a %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not name %q", err, tt.wantHas)
			}
		})
	}
}
