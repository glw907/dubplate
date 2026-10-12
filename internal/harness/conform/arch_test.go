package conform

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os/exec"
	"path/filepath"
	"slices"
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
// syntactically valid. A testdata subtree is skipped entirely: a
// fixture's own .go file (an arch-check counter-example, say) is not
// production code and must never make this walk false-positive.
func checkNoHarnessImports(root string) error {
	var errs []error
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
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
// never a clock read. A testdata subtree is skipped entirely, for the
// same reason checkNoHarnessImports skips one. The walk descends into
// every subdirectory under root, not only root itself, which is what
// lets root widen from internal/engine/matcher to internal/engine
// without missing a package the widening added (R21).
func checkNoTimeNow(root string) error {
	var errs []error
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return fs.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
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
				errs = append(errs, fmt.Errorf("%s:%s: calls time.Now, which internal/engine must never do", path, fset.Position(call.Pos())))
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

// TestEngineNeverReadsTheClock checks the real tree: nothing under
// internal/engine, matcher or musicbrainz alike, calls time.Now (R21:
// widened from internal/engine/matcher now that musicbrainz sits
// beside it under the same no-clock-read invariant).
func TestEngineNeverReadsTheClock(t *testing.T) {
	root := filepath.Join(repoRoot(t), "internal", "engine")
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

// TestTimeNowCheckDescendsIntoNestedPackages proves the walk descends
// into a subdirectory rather than only scanning the root (R21): the
// violation in testdata/arch/time-now-nested/musicbrainz/events.go
// fires when the root is the tree above it, and the walk narrowed back
// to the root alone must miss it.
func TestTimeNowCheckDescendsIntoNestedPackages(t *testing.T) {
	root := filepath.Join("testdata", "arch", "time-now-nested")
	err := checkNoTimeNow(root)
	if err == nil {
		t.Fatal("expected checkNoTimeNow to descend into the nested fixture package and fire")
	}
	if !strings.Contains(err.Error(), "time.Now") {
		t.Errorf("error does not name the violation: %v", err)
	}
	if !strings.Contains(err.Error(), filepath.Join("musicbrainz", "events.go")) {
		t.Errorf("error does not name the nested file: %v", err)
	}
}

// serverUnlinkable lists the import-path prefixes the production server
// binary must never link: the harness, the contract checker, the OpenAPI
// library the checker wraps, and tusd's file locker with its lock library,
// whose flock is Unix-only (AC 33's server clause).
var serverUnlinkable = []string{
	"github.com/glw907/dubplate/internal/harness",
	"github.com/glw907/dubplate/internal/opensubsonic/conformance",
	"github.com/getkin",
	"github.com/tus/tusd/v2/pkg/filelocker",
	"github.com/tus/lockfile",
}

// serverLinkViolations returns the paths in a `go list -deps` listing that
// lie under a prefix in serverUnlinkable. Each line is cut at its first
// space, which drops a test-variant annotation.
func serverLinkViolations(listing string) []string {
	var bad []string
	for line := range strings.Lines(listing) {
		path, _, _ := strings.Cut(strings.TrimSpace(line), " ")
		for _, prefix := range serverUnlinkable {
			if path == prefix || strings.HasPrefix(path, prefix+"/") {
				bad = append(bad, path)
				break
			}
		}
	}
	return bad
}

// TestServerNeverLinksTheHarnessOrChecker checks the real build: the
// dependency closure of cmd/dubplate-server names no harness package, no
// contract checker, and no kin-openapi path.
func TestServerNeverLinksTheHarnessOrChecker(t *testing.T) {
	cmd := exec.CommandContext(t.Context(), "go", "list", "-deps", "./cmd/dubplate-server")
	cmd.Dir = repoRoot(t)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), "github.com/glw907/dubplate/cmd/dubplate-server") {
		t.Fatalf("listing never names the server package; the check proves nothing:\n%s", out)
	}
	if bad := serverLinkViolations(string(out)); len(bad) > 0 {
		t.Errorf("cmd/dubplate-server links packages it must not:\n%s", strings.Join(bad, "\n"))
	}
}

// TestServerLinkCheckFiresOnItsOwnListing proves serverLinkViolations
// names a harness path and a kin-openapi path in a canned listing, and
// passes a clean one.
func TestServerLinkCheckFiresOnItsOwnListing(t *testing.T) {
	const prefix = "github.com/glw907/dubplate/"
	tests := []struct {
		name    string
		listing string
		want    []string
	}{
		{
			"harness and kin-openapi",
			"net/http\n" + prefix + "internal/harness/subsonic/grade\ngithub.com/getkin/kin-openapi/openapi3\n",
			[]string{prefix + "internal/harness/subsonic/grade", "github.com/getkin/kin-openapi/openapi3"},
		},
		{"the checker", prefix + "internal/opensubsonic/conformance\n", []string{prefix + "internal/opensubsonic/conformance"}},
		{"tusd's file locker and its lock library", "github.com/tus/tusd/v2/pkg/filelocker\ngithub.com/tus/lockfile\n", []string{"github.com/tus/tusd/v2/pkg/filelocker", "github.com/tus/lockfile"}},
		{"tusd's memory locker is allowed", "github.com/tus/tusd/v2/pkg/memorylocker\n", nil},
		{"the harness root", prefix + "internal/harness\n", []string{prefix + "internal/harness"}},
		{"a test-variant line", prefix + "internal/harness/conform [" + prefix + "internal/harness/conform.test]\n", []string{prefix + "internal/harness/conform"}},
		{"the client is allowed", prefix + "internal/opensubsonic/client\n", nil},
		{"a sibling with the harness name as a prefix", prefix + "internal/harnessx\n", nil},
		{"clean", "net/http\n" + prefix + "internal/config\n", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := serverLinkViolations(tt.listing); !slices.Equal(got, tt.want) {
				t.Errorf("serverLinkViolations = %v, want %v", got, tt.want)
			}
		})
	}
}
