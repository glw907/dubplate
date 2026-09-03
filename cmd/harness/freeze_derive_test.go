//go:build oracle

// The test in this file is the only one in this package that launches
// the real oracle shim, so it is the only one needing the pinned uv
// venv (tools/venv/README.md). `make derive` runs it; the repository
// gate does not build this file at all.
package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/glw907/dubplate/internal/harness/oracle"
)

// venvOptions resolves the pinned venv the same way freeze's own
// resolveVenv does, failing loudly when tools/venv/.venv was never
// built.
func venvOptions(t *testing.T) oracle.Options {
	t.Helper()
	venv, err := resolveVenv()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	if _, err := os.Stat(filepath.Join(venv, "bin", "python3")); err != nil {
		t.Fatalf("no venv at %s: build it per tools/venv/README.md: %v", venv, err)
	}
	return oracle.Options{Venv: venv}
}

// copyCaseDir copies src's case.json and responses/ tree into dst
// verbatim, so a test can freeze a scratch copy of a committed corpus
// case without touching the checked-in tree.
func copyCaseDir(t *testing.T, src, dst string) {
	t.Helper()
	entries, err := os.ReadDir(src)
	if err != nil {
		t.Fatalf("read %s: %v", src, err)
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			if err := os.MkdirAll(dstPath, 0o755); err != nil {
				t.Fatalf("mkdir %s: %v", dstPath, err)
			}
			copyCaseDir(t, srcPath, dstPath)
			continue
		}
		data, err := os.ReadFile(srcPath)
		if err != nil {
			t.Fatalf("read %s: %v", srcPath, err)
		}
		if err := os.WriteFile(dstPath, data, 0o644); err != nil {
			t.Fatalf("write %s: %v", dstPath, err)
		}
	}
}

// TestFreezeReDerivesACommittedCaseIdempotently is the real-venv
// counterpart to TestFreezeIsIdempotentExceptExpected: it freezes a
// scratch copy of the committed year-penalty case twice against
// pinned beets and proves everything but "expected" stays byte-
// identical, the property Case.Write's Input re-marshal has to hold
// for `harness freeze` to be safe to run against an already-frozen
// corpus case.
func TestFreezeReDerivesACommittedCaseIdempotently(t *testing.T) {
	opts := venvOptions(t)

	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	src := filepath.Join(root, "corpus", "cases", "year-penalty")

	dir := t.TempDir()
	copyCaseDir(t, src, dir)

	if err := runFreeze(context.Background(), dir, opts); err != nil {
		t.Fatalf("first freeze: %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatalf("read case.json after first freeze: %v", err)
	}

	if err := runFreeze(context.Background(), dir, opts); err != nil {
		t.Fatalf("second freeze: %v", err)
	}
	second, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatalf("read case.json after second freeze: %v", err)
	}

	if got, want := stripExpected(t, second), stripExpected(t, first); got != want {
		t.Errorf("freeze against the real shim is not idempotent outside 'expected':\nfirst:  %s\nsecond: %s", want, got)
	}
}
