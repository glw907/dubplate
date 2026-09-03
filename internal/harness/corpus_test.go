// Package harness has no production code of its own; corpus_test.go
// is the gate-resident static consistency check over corpus/cases and
// corpus/manifest.json (rung 2, pass 2a, Task 5). It carries no build
// tag and needs no venv, network, or live shim: everything here reads
// only what is already committed, through [cassette.LoadCase] and
// [cassette.LoadManifest].
package harness

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// repoRoot walks up from the working directory to the nearest
// ancestor holding go.mod, matching the oracle package's own helper:
// `go test` runs from this package's directory, so corpus/ has to be
// resolved relative to the module root, not the CWD.
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("locate repo root: no go.mod above %s", dir)
		}
		dir = parent
	}
}

// corpusCaseDirs lists every case directory under corpus/cases,
// skipping non-directory entries (there are none today, but a stray
// file should never silently drop out of the check).
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
		t.Fatal("corpus/cases has no case directories; the smoke corpus is missing")
	}
	return dirs
}

// forEachCase runs fn as a t.Run subtest per corpus case directory
// under root, after loading and validating the case through
// cassette.LoadCase, so every caller below can assume c loaded clean
// instead of repeating the load-and-fail boilerplate.
func forEachCase(t *testing.T, root string, fn func(t *testing.T, id, dir string, c *cassette.Case)) {
	t.Helper()
	for _, dir := range corpusCaseDirs(t, root) {
		id := filepath.Base(dir)
		t.Run(id, func(t *testing.T) {
			c, err := cassette.LoadCase(dir)
			if err != nil {
				t.Fatalf("LoadCase(%s): %v", dir, err)
			}
			fn(t, id, dir, c)
		})
	}
}

// TestEveryCorpusCaseLoadsClean proves cassette.LoadCase accepts
// every committed case as-is: a case that fails here would fail the
// same way for any consumer (harness capture, oracle.Derive, or a
// future dubplate-server importer).
func TestEveryCorpusCaseLoadsClean(t *testing.T) {
	root := repoRoot(t)
	forEachCase(t, root, func(t *testing.T, id, _ string, c *cassette.Case) {
		if c.ID != id {
			t.Errorf("case.json id = %q, want %q (directory name)", c.ID, id)
		}
		if c.Expected == nil {
			t.Errorf("case %s carries no expected record; it was never derived", id)
		}
	})
}

// TestEveryLoggedRequestResolvesToExactlyOneResponse proves the
// request log itself is unambiguous: no two requests in one case
// share a canonical (method, path_query) key, since [stub.New] keys
// its replay map on exactly that pair and a collision would let the
// second request silently shadow the first with no error from
// anywhere else in the pipeline. Response-file existence is already
// enforced by LoadCase itself; this test only checks the mapping
// this package's own docs promise is unambiguous.
func TestEveryLoggedRequestResolvesToExactlyOneResponse(t *testing.T) {
	root := repoRoot(t)
	forEachCase(t, root, func(t *testing.T, id, _ string, c *cassette.Case) {
		type key struct{ method, pathQuery string }
		seen := make(map[key]string, len(c.Requests))
		for _, req := range c.Requests {
			k := key{req.Method, req.PathQuery}
			if prior, dup := seen[k]; dup {
				t.Errorf("case %s: requests for %s %s resolve to two responses (%s and %s)",
					id, req.Method, req.PathQuery, prior, req.Response)
			}
			seen[k] = req.Response
		}
	})
}

// TestEveryResponseFileIsConsumed proves every file under a case's
// responses/ directory is either referenced by a logged request or
// listed in Extra. LoadCase only validates the reverse direction (a
// referenced file must exist); an orphaned response file would sit in
// the corpus unexplained, silently growing the fixture set without
// U8's recall-extension bookkeeping.
func TestEveryResponseFileIsConsumed(t *testing.T) {
	root := repoRoot(t)
	forEachCase(t, root, func(t *testing.T, id, dir string, c *cassette.Case) {
		consumed := make(map[string]bool, len(c.Requests)+len(c.Extra))
		for _, req := range c.Requests {
			consumed[req.Response] = true
		}
		for _, extra := range c.Extra {
			consumed[extra] = true
		}

		entries, err := os.ReadDir(filepath.Join(dir, "responses"))
		if err != nil {
			t.Fatalf("read %s/responses: %v", dir, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := filepath.Join("responses", e.Name())
			if !consumed[name] {
				t.Errorf("case %s: response file %s is neither a logged request's response nor listed in Extra", id, name)
			}
		}
	})
}

// wantManifestSchemaVersion pins the schema version this test
// expects corpus/manifest.json to carry. A schema bump is a
// corpus-revalidation event like any other pin change (corpus/README.md);
// this constant is what makes an un-updated manifest fail loudly here
// instead of silently loading under an assumption the corpus no
// longer meets.
const wantManifestSchemaVersion = "1"

// TestManifestParsesAndMatchesSchemaVersion proves corpus/manifest.json
// itself is a valid Manifest and carries the schema version this
// corpus was built against.
func TestManifestParsesAndMatchesSchemaVersion(t *testing.T) {
	root := repoRoot(t)
	m, err := cassette.LoadManifest(filepath.Join(root, "corpus", "manifest.json"))
	if err != nil {
		t.Fatalf("LoadManifest(corpus/manifest.json): %v", err)
	}
	if m.SchemaVersion != wantManifestSchemaVersion {
		t.Errorf("corpus/manifest.json schema_version = %q, want %q", m.SchemaVersion, wantManifestSchemaVersion)
	}
}

// TestLoadCaseRejectsCorruptedFixture is the negative case: a
// deliberately corrupted case directory (cassette's own testdata,
// missing eval_date) must fail LoadCase, proving this gate-resident
// check would actually catch a broken case rather than passing
// vacuously over an empty corpus/cases.
func TestLoadCaseRejectsCorruptedFixture(t *testing.T) {
	dir := filepath.Join("cassette", "testdata", "corrupted-case")
	_, err := cassette.LoadCase(dir)
	if err == nil {
		t.Fatalf("LoadCase(%s): got nil error, want one", dir)
	}
	if !strings.Contains(err.Error(), "eval_date") {
		t.Errorf("error %q does not name the fixture's corruption (missing eval_date)", err)
	}
}
