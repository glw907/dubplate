package analysis

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// fullFileTagCases builds a committed-corpus stand-in carrying every
// FileTagSubjects id with a well-formed Expected.AppliedFields entry at
// item index 0, so fullFileTagLedger's rows all resolve cleanly.
func fullFileTagCases() map[string]*cassette.Case {
	cases := make(map[string]*cassette.Case, len(FileTagSubjects))
	for _, s := range FileTagSubjects {
		cases[s.CaseID] = &cassette.Case{
			ID: s.CaseID,
			Expected: &cassette.Expected{
				AppliedFields: map[string]any{
					"0": map[string]any{"title": "x", "artist": "y"},
				},
			},
		}
	}
	return cases
}

// fullFileTagLedger builds a ledger that satisfies every check, so each
// subtest below can break exactly one thing and attribute the failure
// to it, matching fullTable's role for CheckComponentTable.
func fullFileTagLedger() *FileTagLedger {
	l := &FileTagLedger{
		ShimVersion:    "1",
		BeetsVersion:   "2.14.1",
		WritableFields: []string{"title", "artist"},
		ReadonlyFields: []string{"length", "bitrate"},
		Template:       map[string]any{"length": 0.0, "bitrate": 0},
	}
	for _, s := range FileTagSubjects {
		l.Rows = append(l.Rows, FileTagLedgerRow{
			CaseID:    s.CaseID,
			ItemIndex: s.ItemIndex,
			Values:    map[string]any{"title": "x", "artist": "y"},
		})
	}
	return l
}

func TestCheckFileTagLedger(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(*FileTagLedger)
		wantHas string // empty means the ledger is accepted
	}{
		{name: "a complete ledger", mutate: func(*FileTagLedger) {}},
		{
			name:    "a shrunken ledger",
			mutate:  func(l *FileTagLedger) { l.Rows = l.Rows[:4] },
			wantHas: "want at least",
		},
		{
			name: "a row pointing at a case that is not committed",
			mutate: func(l *FileTagLedger) {
				l.Rows[0].CaseID = "no-such-case"
			},
			wantHas: `"no-such-case" is not a committed corpus case`,
		},
		{
			name: "a row carrying a value key outside the writable set",
			mutate: func(l *FileTagLedger) {
				l.Rows[0].Values["genre"] = "Rock"
			},
			wantHas: `outside the ledger's own writable_fields header`,
		},
		{
			name: "a row carrying a read-only key",
			mutate: func(l *FileTagLedger) {
				l.Rows[0].Values["length"] = 180.0
			},
			wantHas: `carries read-only field "length"`,
		},
		{
			name: "a row whose item index does not resolve",
			mutate: func(l *FileTagLedger) {
				l.Rows[0].ItemIndex = 7
			},
			wantHas: "does not resolve into the case's own applied_fields",
		},
		{
			name:    "a subject missing from the ledger",
			mutate:  func(l *FileTagLedger) { l.Rows = l.Rows[1:] },
			wantHas: `missing row for subject "clean-match"`,
		},
		{
			name:    "no shim version",
			mutate:  func(l *FileTagLedger) { l.ShimVersion = "" },
			wantHas: "shim_version: missing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := fullFileTagLedger()
			tt.mutate(l)
			err := CheckFileTagLedger(fullFileTagCases(), l)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckFileTagLedger: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckFileTagLedger accepted a ledger that is %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

// TestLoadFileTagLedgerParsesTheTestdataFixture proves LoadFileTagLedger
// reads a well-formed, on-disk document covering all twelve named
// subjects and that the loaded ledger then satisfies CheckFileTagLedger
// against a matching case set, not only the in-code fullFileTagLedger.
func TestLoadFileTagLedgerParsesTheTestdataFixture(t *testing.T) {
	l, err := LoadFileTagLedger(filepath.Join("testdata", "file_tags", "good", "file_tags.json"))
	if err != nil {
		t.Fatalf("LoadFileTagLedger: %v", err)
	}
	if len(l.Rows) != len(FileTagSubjects) {
		t.Fatalf("rows = %d, want %d", len(l.Rows), len(FileTagSubjects))
	}
	if err := CheckFileTagLedger(fullFileTagCases(), l); err != nil {
		t.Errorf("CheckFileTagLedger: %v", err)
	}
}

func TestLoadFileTagLedgerRejectsAnInventedKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file_tags.json")
	body := `{"shim_version":"1","beets_version":"2.14.1","writable_fields":[],"readonly_fields":[],"template":{},"rows":[],"extra_key":true}`
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if _, err := LoadFileTagLedger(path); err == nil {
		t.Fatal("LoadFileTagLedger accepted a key nothing parses")
	}
}
