package conform

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// forbiddenImportPrefix is the harness import path internal/engine may
// never carry: the first architectural invariant (internal/harness
// reads the corpus, drives beets, and grades; the engine is the
// product, and tier 0's decoder lives here rather than there exactly
// because of this rule).
const forbiddenImportPrefix = "github.com/glw907/dubplate/internal/harness"

// checkNoHarnessImports walks every .go file under root and fails
// (naming the file and the import) if any file imports a path under
// forbiddenImportPrefix. Parsing runs in go/parser's ImportsOnly mode:
// cheap, and it needs nothing beyond the import block to be
// syntactically valid.
func checkNoHarnessImports(root string) error {
	var errs []error
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		for _, imp := range f.Imports {
			importPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				continue
			}
			if importPath == forbiddenImportPrefix || strings.HasPrefix(importPath, forbiddenImportPrefix+"/") {
				errs = append(errs, fmt.Errorf("%s: imports %s, which internal/engine must never import", path, importPath))
			}
		}
		return nil
	})
	if walkErr != nil {
		errs = append(errs, walkErr)
	}
	return errors.Join(errs...)
}

// checkNoTimeNow walks every .go file under root and fails (naming the
// file and position) if any file calls time.Now: the evaluation date
// is an explicit parameter on the album-distance call (contract 9),
// never a clock read.
func checkNoTimeNow(root string) error {
	var errs []error
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
		if err != nil {
			return fmt.Errorf("parse %s: %w", path, err)
		}
		timeName := importedName(f, "time")
		if timeName == "" {
			return nil
		}
		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := sel.X.(*ast.Ident)
			if ok && ident.Name == timeName && sel.Sel.Name == "Now" {
				errs = append(errs, fmt.Errorf("%s:%s: calls time.Now, which internal/engine/matcher must never do", path, fset.Position(call.Pos())))
			}
			return true
		})
		return nil
	})
	if walkErr != nil {
		errs = append(errs, walkErr)
	}
	return errors.Join(errs...)
}

// importedName returns the local identifier f uses for the package at
// path, or "" if f does not import it.
func importedName(f *ast.File, path string) string {
	for _, imp := range f.Imports {
		importPath, err := strconv.Unquote(imp.Path.Value)
		if err != nil || importPath != path {
			continue
		}
		if imp.Name != nil {
			return imp.Name.Name
		}
		return path
	}
	return ""
}

// TestEngineNeverImportsHarness is the first architectural invariant,
// checked against the real tree: internal/engine imports nothing under
// internal/harness.
func TestEngineNeverImportsHarness(t *testing.T) {
	root := filepath.Join(repoRoot(t), "internal", "engine")
	if err := checkNoHarnessImports(root); err != nil {
		t.Error(err)
	}
}

// TestMatcherNeverReadsTheClock checks the real tree: nothing under
// internal/engine/matcher calls time.Now.
func TestMatcherNeverReadsTheClock(t *testing.T) {
	root := filepath.Join(repoRoot(t), "internal", "engine", "matcher")
	if err := checkNoTimeNow(root); err != nil {
		t.Error(err)
	}
}

// TestHarnessImportCheckFiresOnItsOwnFixture proves
// checkNoHarnessImports fires, against a hand-authored fixture tree
// (testdata/arch/harness-import-violation) carrying one file that
// imports internal/harness/cassette.
func TestHarnessImportCheckFiresOnItsOwnFixture(t *testing.T) {
	root := filepath.Join("testdata", "arch", "harness-import-violation")
	err := checkNoHarnessImports(root)
	if err == nil {
		t.Fatal("expected checkNoHarnessImports to fire against its own violating fixture")
	}
	if !strings.Contains(err.Error(), "internal/harness/cassette") {
		t.Errorf("error does not name the violating import: %v", err)
	}
}

// TestTimeNowCheckFiresOnItsOwnFixture proves checkNoTimeNow fires,
// against a hand-authored fixture tree
// (testdata/arch/time-now-violation) carrying one file that calls
// time.Now.
func TestTimeNowCheckFiresOnItsOwnFixture(t *testing.T) {
	root := filepath.Join("testdata", "arch", "time-now-violation")
	err := checkNoTimeNow(root)
	if err == nil {
		t.Fatal("expected checkNoTimeNow to fire against its own violating fixture")
	}
	if !strings.Contains(err.Error(), "time.Now") {
		t.Errorf("error does not name the violation: %v", err)
	}
}
