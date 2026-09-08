package tagio

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"slices"
	"strconv"
	"strings"
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
// at which item index, the row grades, and the 69 values pinned
// beets read back out of the container it wrote. Values is the
// conformance target and the only expectation the grading uses; the
// ledger's shared template block carries the nine read-only keys and
// describes the fixture rather than any subject, so this decode does
// not name it at all (corpus/README.md:2401-2412).
type ledgerRow struct {
	CaseID    string         `json:"case_id"`
	ItemIndex int            `json:"item_index"`
	Values    map[string]any `json:"values"`
}

// The nine read-only keys, hand-transcribed from
// corpus/README.md:2401-2412. They are audio properties a reader
// takes from the container's own STREAMINFO rather than anything
// Item.write sets, they sit once in the ledger's template block and
// never in a subject row, and the grading asserts by name that none
// of them reaches a comparison, so a later reader cannot mistake
// their absence for an oversight.
var readOnlyFields = []string{
	"bitdepth",
	"bitrate",
	"bitrate_mode",
	"channels",
	"encoder_info",
	"encoder_settings",
	"format",
	"length",
	"samplerate",
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

// divergence is one key on which one subject's read-back disagreed
// with its ledger row. It carries both sides rendered as JSON text
// rather than as decoded values, because the rendering is what a
// failure message and Task 6's staged row both need and because it
// spells null, "" and [] distinctly.
type divergence struct {
	subject  string
	key      string
	ledger   string
	readBack string
}

func (d divergence) String() string {
	return fmt.Sprintf("subject %s: key %s: ledger %s, read-back %s", d.subject, d.key, d.ledger, d.readBack)
}

// absentValue renders a key that is not present on its side of a
// comparison at all. No JSON value spells it, so it can never
// collide with a rendered null, empty string or empty list.
const absentValue = "<absent>"

// renderValue renders one side of a comparison for a failure
// message.
func renderValue(v any, present bool) string {
	if !present {
		return absentValue
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<unrenderable %T: %v>", v, err)
	}
	return string(b)
}

// compareSubject grades subject's read-back projection against its
// ledger row under R10's exact typed comparison. Both sides arrive
// already decoded through encoding/json, so a number on either side
// is the same float64 and equality is exact: no tolerance on any
// float, no case for any key, and a list compared in its own order.
//
// The walk covers the union of both sides' keys rather than the
// ledger's alone, so a key present in one side and absent from the
// other is a divergence rather than a comparison that silently did
// not happen. Every disagreement is reported, never the first, so
// one run produces the whole divergence set; the returned count is
// how many keys were compared, which the caller grades against 69.
func compareSubject(subject string, ledger, readBack map[string]any) ([]divergence, int) {
	keys := make([]string, 0, len(ledger)+len(readBack))
	for k := range ledger {
		keys = append(keys, k)
	}
	for k := range readBack {
		if _, ok := ledger[k]; !ok {
			keys = append(keys, k)
		}
	}
	slices.Sort(keys)

	var found []divergence
	for _, k := range keys {
		want, wantOK := ledger[k]
		got, gotOK := readBack[k]
		if wantOK && gotOK && reflect.DeepEqual(want, got) {
			continue
		}
		found = append(found, divergence{
			subject:  subject,
			key:      k,
			ledger:   renderValue(want, wantOK),
			readBack: renderValue(got, gotOK),
		})
	}
	return found, len(keys)
}

// projectToLedgerShape marshals ts through R11's JSON tags and
// decodes the result back through encoding/json, so the read-back
// side of a comparison carries the same number type the ledger's own
// decode produced. There is no second projection: tagio.Read
// produced ts, and the tags do the rest (P11).
func projectToLedgerShape(t *testing.T, ts TagSet) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("marshal read-back TagSet: %v", err)
	}
	var projected map[string]any
	if err := json.Unmarshal(encoded, &projected); err != nil {
		t.Fatalf("decode read-back TagSet projection: %v", err)
	}
	return projected
}

// decodeSide decodes one synthetic side of a comparator row. The
// rows are JSON text so both sides of the comparator's own test
// carry encoding/json's number type exactly as the instrument's do.
func decodeSide(t *testing.T, text string) map[string]any {
	t.Helper()
	var side map[string]any
	if err := json.Unmarshal([]byte(text), &side); err != nil {
		t.Fatalf("decode synthetic side %s: %v", text, err)
	}
	return side
}

// TestCompareSubjectNamesEveryDisagreement is the exit gate's proof
// that the instrument can fail (spec:1134), discharged as a standing
// test over synthetic rows rather than as a one-off demonstration
// that leaves no trace. Every value below is hard-coded here; none
// is computed by Read, Serialize or the projection, so a comparator
// that stopped comparing would fail these rows rather than agree
// with itself.
//
// Its name deliberately does not match the -run Conformance pattern
// rung 15's launch gate selects with.
func TestCompareSubjectNamesEveryDisagreement(t *testing.T) {
	tests := []struct {
		name         string
		ledger       string
		readBack     string
		wantCompared int
		want         []string
	}{
		{
			name:         "agreeing row",
			ledger:       `{"album":"Rumours","albumartists":["Fleetwood Mac"],"bpm":0,"initial_key":null,"rg_track_peak":0.9999999}`,
			readBack:     `{"album":"Rumours","albumartists":["Fleetwood Mac"],"bpm":0,"initial_key":null,"rg_track_peak":0.9999999}`,
			wantCompared: 5,
			want:         nil,
		},
		{
			name:         "a string differs",
			ledger:       `{"album":"Rumours","title":"The Chain"}`,
			readBack:     `{"album":"Tusk","title":"The Chain"}`,
			wantCompared: 2,
			want:         []string{`subject synthetic: key album: ledger "Rumours", read-back "Tusk"`},
		},
		{
			name:         "a list differs only in its order",
			ledger:       `{"artists":["Lindsey Buckingham","Stevie Nicks"]}`,
			readBack:     `{"artists":["Stevie Nicks","Lindsey Buckingham"]}`,
			wantCompared: 1,
			want:         []string{`subject synthetic: key artists: ledger ["Lindsey Buckingham","Stevie Nicks"], read-back ["Stevie Nicks","Lindsey Buckingham"]`},
		},
		{
			name:         "null on one side against an empty string on the other",
			ledger:       `{"catalognum":null}`,
			readBack:     `{"catalognum":""}`,
			wantCompared: 1,
			want:         []string{`subject synthetic: key catalognum: ledger null, read-back ""`},
		},
		{
			name:         "a float differs in its last digit",
			ledger:       `{"rg_track_peak":0.9999999}`,
			readBack:     `{"rg_track_peak":0.9999998}`,
			wantCompared: 1,
			want:         []string{`subject synthetic: key rg_track_peak: ledger 0.9999999, read-back 0.9999998`},
		},
		{
			name:         "a key the ledger carries is missing from the read-back",
			ledger:       `{"album":"Rumours","script":"Latn"}`,
			readBack:     `{"album":"Rumours"}`,
			wantCompared: 2,
			want:         []string{`subject synthetic: key script: ledger "Latn", read-back <absent>`},
		},
		{
			name:         "a key the read-back carries is missing from the ledger",
			ledger:       `{"album":"Rumours"}`,
			readBack:     `{"album":"Rumours","script":"Latn"}`,
			wantCompared: 2,
			want:         []string{`subject synthetic: key script: ledger <absent>, read-back "Latn"`},
		},
		{
			name:         "every disagreement is named, not the first",
			ledger:       `{"album":"Rumours","genres":["Rock"],"year":1977}`,
			readBack:     `{"album":"Tusk","genres":[],"year":1977}`,
			wantCompared: 3,
			want: []string{
				`subject synthetic: key album: ledger "Rumours", read-back "Tusk"`,
				`subject synthetic: key genres: ledger ["Rock"], read-back []`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			found, compared := compareSubject("synthetic", decodeSide(t, tt.ledger), decodeSide(t, tt.readBack))
			if compared != tt.wantCompared {
				t.Errorf("compared %d keys, want %d", compared, tt.wantCompared)
			}

			var got []string
			for _, d := range found {
				got = append(got, d.String())
				t.Logf("observed failure text: %s", d)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("compareSubject reported\n\t%q\nwant\n\t%q", got, tt.want)
			}
		})
	}
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
// no picture behavior, INV:432-441).
//
// Step 4: the scratch copy is read back through tagio.Read, the one
// read entry point, and the resulting TagSet is projected to the
// ledger's shape through R11's tags, so no second projection exists
// (P11). Step 5: all 69 writable keys are compared against the
// subject's ledger row under R10's exact typed comparison. The nine
// read-only keys are never compared and the ledger's template block
// is never a per-subject expectation.
//
// A disagreement here is a finding for the conductor, never a
// comparison to relax: the run names every subject, key and pair of
// values it found, and the gate stays red until the divergence is
// ruled on.
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
	var divergences []divergence

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

			readBack, err := Read(scratch)
			if err != nil {
				t.Fatalf("Read %q back from %s: %v", row.CaseID, scratch, err)
			}
			projected := projectToLedgerShape(t, readBack)

			found, compared := compareSubject(row.CaseID, row.Values, projected)
			if compared != 69 {
				t.Errorf("case %q: compared %d keys, want 69", row.CaseID, compared)
			}
			for _, name := range readOnlyFields {
				if _, ok := row.Values[name]; ok {
					t.Errorf("case %q: ledger row carries read-only key %q, which belongs to the template block alone", row.CaseID, name)
				}
				if _, ok := projected[name]; ok {
					t.Errorf("case %q: read-back projection carries read-only key %q, which no subject grades", row.CaseID, name)
				}
			}

			divergences = append(divergences, found...)
			for _, d := range found {
				t.Errorf("%s", d)
			}
		})
	}

	if len(divergences) > 0 {
		lines := make([]string, 0, len(divergences))
		for _, d := range divergences {
			lines = append(lines, d.String())
		}
		t.Errorf("%d divergences between the ledger and the read-back, the full list:\n\t%s",
			len(divergences), strings.Join(lines, "\n\t"))
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
