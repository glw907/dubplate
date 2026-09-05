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

	"github.com/glw907/dubplate/internal/harness/analysis"
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
const wantManifestSchemaVersion = "3"

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

// TestEveryCorpusCaseSatisfiesItsClaims reads each case's sibling
// claims.json and checks it against the committed record. A record that
// still re-derives byte-identically can nevertheless have stopped
// exercising the behavior its case exists for; the claims are what
// notice.
func TestEveryCorpusCaseSatisfiesItsClaims(t *testing.T) {
	root := repoRoot(t)
	forEachCase(t, root, func(t *testing.T, id, dir string, c *cassette.Case) {
		if c.Expected == nil {
			t.Fatalf("case %s carries no expected record to check claims against", id)
		}
		claims, err := analysis.LoadClaims(dir)
		if err != nil {
			t.Fatalf("case %s: %v", id, err)
		}
		if err := analysis.CheckClaims(c.Expected, claims); err != nil {
			t.Errorf("case %s:\n%v", id, err)
		}
	})
}

// TestProximityLedgerAcknowledgesEveryFlaggedCase recomputes the
// threshold-proximity flag from every committed record and diffs it
// against corpus/analysis/proximity.json. A flag is not a failure; an
// unacknowledged flag, and a stale acknowledgment, both are.
func TestProximityLedgerAcknowledgesEveryFlaggedCase(t *testing.T) {
	root := repoRoot(t)
	got := make(map[string]analysis.Proximity)
	forEachCase(t, root, func(t *testing.T, id, _ string, c *cassette.Case) {
		if c.Expected == nil {
			t.Fatalf("case %s carries no expected record to flag", id)
		}
		p, err := analysis.Flag(c.Expected)
		if err != nil {
			t.Fatalf("case %s: %v", id, err)
		}
		got[id] = p
	})

	ledger, err := analysis.LoadProximityLedger(filepath.Join(root, "corpus", "analysis", "proximity.json"))
	if err != nil {
		t.Fatalf("load proximity ledger: %v", err)
	}
	if err := analysis.CheckProximityLedger(got, ledger); err != nil {
		t.Error(err)
	}
}

// TestTieLedgerCoversEveryCorpusCase checks the assignment-tie ledger's
// coverage and its dispositions. The measurements themselves are
// derived work (internal/harness/oracle's derive_ties.py, under `make
// derive`); what the gate can prove is that no case is missing from the
// ledger and no tied case sits there undispositioned.
func TestTieLedgerCoversEveryCorpusCase(t *testing.T) {
	root := repoRoot(t)
	var ids []string
	for _, dir := range corpusCaseDirs(t, root) {
		ids = append(ids, filepath.Base(dir))
	}

	ledger, err := analysis.LoadTieLedger(filepath.Join(root, "corpus", "analysis", "ties.json"))
	if err != nil {
		t.Fatalf("load tie ledger: %v", err)
	}
	if err := analysis.CheckTieLedger(ids, ledger); err != nil {
		t.Error(err)
	}
}

// TestComponentTableIsWellFormed checks corpus/analysis/components.json
// for schema validity, row-id uniqueness, its row-count floor, and its
// per-method coverage. It is deliberately not a conformance check: the
// Go matcher the table's outputs will be compared against arrives at
// rung 3, and re-deriving the table is oracle-tagged work under `make
// derive`.
func TestComponentTableIsWellFormed(t *testing.T) {
	root := repoRoot(t)
	table, err := analysis.LoadComponentTable(filepath.Join(root, "corpus", "analysis", "components.json"))
	if err != nil {
		t.Fatalf("load component table: %v", err)
	}
	if err := analysis.CheckComponentTable(table); err != nil {
		t.Error(err)
	}
}

// TestFileTagLedgerIsWellFormed checks corpus/analysis/file_tags.json:
// every row resolves into a committed case's own applied_fields at the
// row's item index, every row's value keys sit inside the ledger's own
// writable_fields header and never carry a read-only key, and the row
// set is exactly the twelve named analysis.FileTagSubjects. It is
// deliberately not a conformance check: the numbers themselves are
// oracle-tagged work under `make derive`, matching
// TestComponentTableIsWellFormed's own discipline.
func TestFileTagLedgerIsWellFormed(t *testing.T) {
	root := repoRoot(t)
	cases := make(map[string]*cassette.Case)
	forEachCase(t, root, func(t *testing.T, id, _ string, c *cassette.Case) {
		cases[id] = c
	})

	ledger, err := analysis.LoadFileTagLedger(filepath.Join(root, "corpus", "analysis", "file_tags.json"))
	if err != nil {
		t.Fatalf("load file tag ledger: %v", err)
	}
	if err := analysis.CheckFileTagLedger(cases, ledger); err != nil {
		t.Error(err)
	}
}

// TestAlbumInfoLedgerIsWellFormed checks
// corpus/analysis/album_info.json: every row resolves to a committed
// case, the row set is exactly the nineteen named
// analysis.AlbumInfoSubjects, every release lookup in a subject's own
// request log has exactly one row, and a merged block sits on exactly
// the cases named in analysis.AlbumInfoPseudoSubjects. It is
// deliberately not a conformance check: the translated values are
// oracle-tagged work under `make derive`, and grading the Go
// translation against them is internal/harness/conform's, matching
// TestFileTagLedgerIsWellFormed's own discipline.
func TestAlbumInfoLedgerIsWellFormed(t *testing.T) {
	root := repoRoot(t)
	cases := make(map[string]*cassette.Case)
	forEachCase(t, root, func(t *testing.T, id, _ string, c *cassette.Case) {
		cases[id] = c
	})

	ledger, err := analysis.LoadAlbumInfoLedger(filepath.Join(root, "corpus", "analysis", "album_info.json"))
	if err != nil {
		t.Fatalf("load album info ledger: %v", err)
	}
	if err := analysis.CheckAlbumInfoLedger(cases, ledger); err != nil {
		t.Error(err)
	}
}

// TestAdjudicationsLedgerIsWellFormed checks corpus/analysis/adjudications.json:
// every subject case id resolves to a committed case, every row carries
// a recognized disposition with its required evidence, and only a
// human adjudicator may set anything but pending-adjudication. It is
// deliberately not a coverage check: unlike ties.json and
// proximity.json, adjudications.json is an exception log, carrying a
// row only for a case someone suspects the oracle got wrong.
func TestAdjudicationsLedgerIsWellFormed(t *testing.T) {
	root := repoRoot(t)
	var ids []string
	for _, dir := range corpusCaseDirs(t, root) {
		ids = append(ids, filepath.Base(dir))
	}

	ledger, err := analysis.LoadAdjudications(filepath.Join(root, "corpus", "analysis", "adjudications.json"))
	if err != nil {
		t.Fatalf("load adjudications ledger: %v", err)
	}
	if err := analysis.CheckAdjudications(ids, ledger); err != nil {
		t.Error(err)
	}
}

// TestOracleBugsLedgerIsWellFormed checks corpus/analysis/oracle_bugs.json
// against both the committed corpus (every probe case id must resolve)
// and the adjudications ledger (every adjudication_refs entry must
// resolve, and a row reproduced "yes" must reference at least one).
func TestOracleBugsLedgerIsWellFormed(t *testing.T) {
	root := repoRoot(t)
	var ids []string
	for _, dir := range corpusCaseDirs(t, root) {
		ids = append(ids, filepath.Base(dir))
	}

	adjLedger, err := analysis.LoadAdjudications(filepath.Join(root, "corpus", "analysis", "adjudications.json"))
	if err != nil {
		t.Fatalf("load adjudications ledger: %v", err)
	}
	adjIDs := make(map[string]bool, len(adjLedger.Entries))
	for _, e := range adjLedger.Entries {
		adjIDs[e.ID] = true
	}

	bugs, err := analysis.LoadOracleBugs(filepath.Join(root, "corpus", "analysis", "oracle_bugs.json"))
	if err != nil {
		t.Fatalf("load oracle bugs ledger: %v", err)
	}
	if err := analysis.CheckOracleBugs(ids, adjIDs, bugs); err != nil {
		t.Error(err)
	}
}

// TestDefectFreeBlocksAreWellFormed checks the defect_free block
// ruling A / M6 adds to corpus/analysis/adjudications.json: every
// string-heuristic probe case named in corpus/analysis/oracle_bugs.json
// resolves to an adjudication row carrying a well-formed block, and no
// row outside that set carries one. It is deliberately not a
// conformance check: the derived figures themselves are oracle-tagged
// work under `make derive`, matching TestOracleBugsLedgerIsWellFormed's
// own discipline.
func TestDefectFreeBlocksAreWellFormed(t *testing.T) {
	root := repoRoot(t)

	adjLedger, err := analysis.LoadAdjudications(filepath.Join(root, "corpus", "analysis", "adjudications.json"))
	if err != nil {
		t.Fatalf("load adjudications ledger: %v", err)
	}
	bugs, err := analysis.LoadOracleBugs(filepath.Join(root, "corpus", "analysis", "oracle_bugs.json"))
	if err != nil {
		t.Fatalf("load oracle bugs ledger: %v", err)
	}
	if err := analysis.CheckDefectFree(bugs, adjLedger); err != nil {
		t.Error(err)
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
