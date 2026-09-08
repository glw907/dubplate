package tagio

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
)

// The three tracked inputs the conformance instrument reads, each
// path relative to this package's directory, which is `go test`'s own
// working directory. Each is owned by rung 2's corpus derivation and
// is read-only here: the instrument copies templatePath out before
// any write, and the copy under t.TempDir() is what Write ever
// touches (R5). A missing file fails loudly, naming the expected
// path, rather than skipping: a check either gates or does not exist.
const (
	// ledgerPath is rung 2's file-level applied-tag ledger, the
	// conformance target corpus/README.md's handoff section names.
	ledgerPath = "../../../corpus/analysis/file_tags.json"
	// caseFileFormat is rung 2's per-case corpus directory, %s the
	// case id, joined to the case's own case.json.
	caseFileFormat = "../../../corpus/cases/%s/case.json"
	// templatePath is rung 2's vendored template FLAC, the fixture
	// derive_file_tags.py's own per-subject scratch copies came from.
	templatePath = "../../harness/oracle/testdata/template.flac"
)

// ledgerDocument is this test's own local decode of
// corpus/analysis/file_tags.json's shape. It is declared here rather
// than imported from internal/harness/analysis because
// TestEngineNeverImportsHarness (internal/harness/conform/arch_test.go)
// walks this package's test files too (spec:692-697).
type ledgerDocument struct {
	WritableFields []string    `json:"writable_fields"`
	Rows           []ledgerRow `json:"rows"`
}

// ledgerRow is one row of ledgerDocument.Rows: which committed case,
// at which item index, the row grades.
type ledgerRow struct {
	CaseID    string `json:"case_id"`
	ItemIndex int    `json:"item_index"`
}

// caseDocument is this test's own local decode of a corpus
// case.json's shape, narrowed to the one section the instrument
// reads: Expected.AppliedFields, keyed by item index as a decimal
// string. Each field's raw JSON is kept undecoded so the writable
// filter can select a subset and hand it straight to
// json.Unmarshal(..., *TagSet) without a second round trip.
type caseDocument struct {
	Expected struct {
		AppliedFields map[string]map[string]json.RawMessage `json:"applied_fields"`
	} `json:"expected"`
}

// loadLedger reads and decodes ledgerPath, failing loudly with the
// path it expected rather than skipping.
func loadLedger(t *testing.T) ledgerDocument {
	t.Helper()
	data, err := os.ReadFile(ledgerPath)
	if err != nil {
		t.Fatalf("read ledger at %s: %v", ledgerPath, err)
	}
	var doc ledgerDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode ledger at %s: %v", ledgerPath, err)
	}
	return doc
}

// loadCase reads and decodes caseID's committed case.json, failing
// loudly with the path it expected rather than skipping. A successful
// read is what R6 means by a case_id resolving to a committed case
// directory.
func loadCase(t *testing.T, caseID string) caseDocument {
	t.Helper()
	path := fmt.Sprintf(caseFileFormat, caseID)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("resolve case %q: read %s: %v", caseID, path, err)
	}
	var doc caseDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("decode case %q at %s: %v", caseID, path, err)
	}
	return doc
}

// filterToWritable selects applied's keys that name one of writable's
// 69 fields, per the ledger's own writable_fields header rather than
// a second hard-coded list (spec:843-847). It asserts the result
// carries all 69, which is also the check that the case is not
// missing a writable key.
func filterToWritable(t *testing.T, applied map[string]json.RawMessage, writable []string) map[string]json.RawMessage {
	t.Helper()
	filtered := make(map[string]json.RawMessage, len(writable))
	for _, name := range writable {
		v, ok := applied[name]
		if !ok {
			t.Fatalf("applied fields carry no writable key %q", name)
		}
		filtered[name] = v
	}
	if len(filtered) != 69 {
		t.Fatalf("filtered applied fields carry %d keys, want 69", len(filtered))
	}
	return filtered
}

// assertCleanMatchFourStates grades C2's four states against values
// hand-transcribed from corpus/cases/clean-match/case.json's own
// expected.applied_fields["0"], never computed: a string value
// (album), an empty string (acoustid_fingerprint), a list value
// (albumartists), an empty list (arrangers), a JSON null mapping to
// absent (initial_key), a numeric zero mapping to a present zero
// (bpm), and a false boolean mapping to a present false (comp).
func assertCleanMatchFourStates(t *testing.T, ts TagSet) {
	t.Helper()

	if ts.Album == nil || *ts.Album != "Rumours" {
		t.Errorf("clean-match: Album = %v, want a value \"Rumours\"", derefString(ts.Album))
	}
	if ts.AcoustIDFingerprint == nil || *ts.AcoustIDFingerprint != "" {
		t.Errorf("clean-match: AcoustIDFingerprint = %v, want present-but-empty \"\"", derefString(ts.AcoustIDFingerprint))
	}
	if len(ts.Albumartists) != 1 || ts.Albumartists[0] != "Fleetwood Mac" {
		t.Errorf("clean-match: Albumartists = %v, want [\"Fleetwood Mac\"]", ts.Albumartists)
	}
	if ts.Arrangers == nil || len(ts.Arrangers) != 0 {
		t.Errorf("clean-match: Arrangers = %v, want present-but-empty []string{}", ts.Arrangers)
	}
	if ts.InitialKey != nil {
		t.Errorf("clean-match: InitialKey = %v, want absent (nil)", *ts.InitialKey)
	}
	if ts.Bpm == nil || *ts.Bpm != 0 {
		t.Errorf("clean-match: Bpm = %v, want a present value 0", derefInt(ts.Bpm))
	}
	if ts.Comp == nil || *ts.Comp != false {
		t.Errorf("clean-match: Comp = %v, want a present value false", derefBool(ts.Comp))
	}
}

func derefString(p *string) string {
	if p == nil {
		return "<nil>"
	}
	return *p
}

func derefInt(p *int) string {
	if p == nil {
		return "<nil>"
	}
	return strconv.Itoa(*p)
}

func derefBool(p *bool) string {
	if p == nil {
		return "<nil>"
	}
	return strconv.FormatBool(*p)
}

// TestConformanceLedgerSubjects lands P10 steps 1 through 3 and proves
// them on their own: rung 15's launch-gate re-run selects this
// function by name with `go test ./internal/engine/tagio/... -run
// Conformance -count=1`, so no other test in this file carries a name
// that pattern matches.
//
// Step 1: the twelve subjects come off ledgerDocument's own rows
// (R6), never a second hard-coded list. Step 2: each subject's
// applied fields, read from its committed case at the row's item
// index, are filtered to the ledger's own writable_fields header.
// Step 3: the filtered object fills a TagSet under the four-state
// rule (unmarshalled straight through R11's tags) and the filled set
// is written through Write into a per-subject scratch copy of the
// committed template, with a nil picture argument (the corpus grades
// no picture behavior, INV:432-441). Steps 4 and 5, the read-back and
// the grading, are Task 5's.
func TestConformanceLedgerSubjects(t *testing.T) {
	ledger := loadLedger(t)
	if len(ledger.Rows) != 12 {
		t.Fatalf("ledger carries %d subject rows, want 12", len(ledger.Rows))
	}
	if len(ledger.WritableFields) != 69 {
		t.Fatalf("ledger carries %d writable_fields, want 69", len(ledger.WritableFields))
	}

	templateBefore, err := os.Stat(templatePath)
	if err != nil {
		t.Fatalf("stat template at %s: %v", templatePath, err)
	}

	ids := make([]string, 0, len(ledger.Rows))
	scratchPaths := make([]string, 0, len(ledger.Rows))

	for _, row := range ledger.Rows {
		ids = append(ids, row.CaseID)

		t.Run(row.CaseID, func(t *testing.T) {
			caseDoc := loadCase(t, row.CaseID)
			applied, ok := caseDoc.Expected.AppliedFields[strconv.Itoa(row.ItemIndex)]
			if !ok {
				t.Fatalf("case %q carries no expected.applied_fields entry for item %d", row.CaseID, row.ItemIndex)
			}

			filtered := filterToWritable(t, applied, ledger.WritableFields)

			if row.CaseID == "clean-match" && row.ItemIndex == 0 {
				for _, leftover := range []string{"data_source", "album_id", "id", "bitrate"} {
					if _, ok := applied[leftover]; !ok {
						t.Errorf("clean-match applied fields do not carry %q, expected them to for this assertion to mean anything", leftover)
					}
					if _, ok := filtered[leftover]; ok {
						t.Errorf("filtered applied fields still carry %q, the writable filter should have dropped it", leftover)
					}
				}
			}

			if len(filtered) != 69 {
				t.Fatalf("case %q: filtered applied fields carry %d keys, want 69", row.CaseID, len(filtered))
			}

			filteredJSON, err := json.Marshal(filtered)
			if err != nil {
				t.Fatalf("marshal filtered fields for %q: %v", row.CaseID, err)
			}
			var ts TagSet
			if err := json.Unmarshal(filteredJSON, &ts); err != nil {
				t.Fatalf("fill TagSet for %q: %v", row.CaseID, err)
			}

			if row.CaseID == "clean-match" && row.ItemIndex == 0 {
				assertCleanMatchFourStates(t, ts)
			}

			scratch := copyFixture(t, templatePath)
			scratchPaths = append(scratchPaths, scratch)
			if err := Write(scratch, ts, nil, 0o644); err != nil {
				t.Fatalf("Write %q's filled TagSet to %s: %v", row.CaseID, scratch, err)
			}
		})
	}

	if len(ids) != 12 {
		t.Errorf("derived %d subject ids from the ledger, want 12: %v", len(ids), ids)
	}
	t.Logf("conformance subjects (from the ledger's own rows): %v", ids)

	seenPaths := make(map[string]bool, len(scratchPaths))
	for _, p := range scratchPaths {
		if seenPaths[p] {
			t.Errorf("scratch path %q was reused across subjects, want one distinct path per subject", p)
		}
		seenPaths[p] = true
	}
	if len(scratchPaths) != 12 {
		t.Errorf("wrote %d scratch copies, want 12", len(scratchPaths))
	}

	templateAfter, err := os.Stat(templatePath)
	if err != nil {
		t.Fatalf("re-stat template at %s: %v", templatePath, err)
	}
	if templateAfter.Size() != templateBefore.Size() {
		t.Errorf("committed template changed size: was %d bytes, now %d", templateBefore.Size(), templateAfter.Size())
	}
}
