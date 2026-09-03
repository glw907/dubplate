package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// loadRecord reads a hand-authored Expected document from testdata.
// These fixtures are never corpus cases: they exist to park a record
// exactly where a derived one is not expected to sit, which is the only
// way to prove the flag fires at all while the corpus stays clear of
// every boundary.
func loadRecord(t *testing.T, path string) *cassette.Expected {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", path))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var exp cassette.Expected
	if err := json.Unmarshal(data, &exp); err != nil {
		t.Fatalf("parse fixture %s: %v", path, err)
	}
	return &exp
}

// TestFlagMeasuresAgainstTheRecordsOwnThresholds runs the flag over the
// fixture set. records/clear.json and records/on-strong-threshold.json
// are the same record with one mutation between them, the best
// distance moved from 0.1 to the echoed strong_rec_thresh, so the pair
// proves the flag distinguishes them rather than answering one way for
// every input.
func TestFlagMeasuresAgainstTheRecordsOwnThresholds(t *testing.T) {
	tests := []struct {
		name           string
		fixture        string
		wantStatus     string
		wantBoundaries []string
	}{
		{"clear of every boundary", "records/clear.json", ProximityClear, nil},
		{"parked on the strong threshold", "records/on-strong-threshold.json", ProximityFlagged, []string{"strong_rec_thresh"}},
		{"candidate gap on the gap threshold", "records/on-gap-threshold.json", ProximityFlagged, []string{"rec_gap_thresh"}},
		{"no candidate to measure", "records/zero-candidates.json", ProximitySkipped, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Flag(loadRecord(t, tt.fixture))
			if err != nil {
				t.Fatalf("Flag: %v", err)
			}
			if got.Status != tt.wantStatus {
				t.Errorf("status = %q, want %q", got.Status, tt.wantStatus)
			}
			if !slices.Equal(got.Boundaries, tt.wantBoundaries) {
				t.Errorf("boundaries = %v, want %v", got.Boundaries, tt.wantBoundaries)
			}
		})
	}
}

// TestFlagRefusesARecordCarryingNoThresholds proves the flag never
// falls back to a hardcoded boundary: a record predating the schema-2
// echo has nothing to measure against, and measuring it against zeros
// would flag every exact match as sitting on a threshold.
func TestFlagRefusesARecordCarryingNoThresholds(t *testing.T) {
	_, err := Flag(loadRecord(t, "records/no-thresholds.json"))
	if err == nil {
		t.Fatal("Flag accepted a record with no thresholds")
	}
	if !strings.Contains(err.Error(), "thresholds") {
		t.Errorf("error %q does not name the missing thresholds", err)
	}
}

func TestCheckProximityLedger(t *testing.T) {
	flagged := Proximity{Status: ProximityFlagged, Boundaries: []string{"strong_rec_thresh"}}
	skipped := Proximity{Status: ProximitySkipped}
	clear := Proximity{Status: ProximityClear}

	tests := []struct {
		name    string
		got     map[string]Proximity
		ledger  ProximityLedger
		wantHas string // empty means the ledger is accepted
	}{
		{
			name:   "acknowledged flag",
			got:    map[string]Proximity{"rec-boundary": flagged, "clean": clear},
			ledger: ProximityLedger{Entries: []ProximityEntry{{CaseID: "rec-boundary", Status: ProximityFlagged, Boundaries: []string{"strong_rec_thresh"}, Reason: "parks on the strong boundary by design"}}},
		},
		{
			name:    "flagged but unlisted",
			got:     map[string]Proximity{"rec-boundary": flagged},
			ledger:  ProximityLedger{},
			wantHas: "not listed",
		},
		{
			name:    "listed but no longer flagged",
			got:     map[string]Proximity{"rec-boundary": clear},
			ledger:  ProximityLedger{Entries: []ProximityEntry{{CaseID: "rec-boundary", Status: ProximityFlagged, Boundaries: []string{"strong_rec_thresh"}, Reason: "was on the boundary"}}},
			wantHas: "no longer near any boundary",
		},
		{
			name:    "listed with no reason",
			got:     map[string]Proximity{"rec-boundary": flagged},
			ledger:  ProximityLedger{Entries: []ProximityEntry{{CaseID: "rec-boundary", Status: ProximityFlagged, Boundaries: []string{"strong_rec_thresh"}}}},
			wantHas: "no reason",
		},
		{
			name:    "boundaries moved",
			got:     map[string]Proximity{"rec-boundary": flagged},
			ledger:  ProximityLedger{Entries: []ProximityEntry{{CaseID: "rec-boundary", Status: ProximityFlagged, Boundaries: []string{"rec_gap_thresh"}, Reason: "the gap sits on its threshold"}}},
			wantHas: "listed boundaries",
		},
		{
			name:    "zero-candidate case unlisted",
			got:     map[string]Proximity{"admit-zero-tracks": skipped},
			ledger:  ProximityLedger{},
			wantHas: "skipped",
		},
		{
			name:    "entry names no corpus case",
			got:     map[string]Proximity{"clean": clear},
			ledger:  ProximityLedger{Entries: []ProximityEntry{{CaseID: "retired", Status: ProximityFlagged, Boundaries: []string{"strong_rec_thresh"}, Reason: "left over from a deleted case"}}},
			wantHas: "no such corpus case",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckProximityLedger(tt.got, &tt.ledger)
			if tt.wantHas == "" {
				if err != nil {
					t.Fatalf("CheckProximityLedger: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckProximityLedger accepted a ledger that is %s", tt.name)
			}
			if !strings.Contains(err.Error(), tt.wantHas) {
				t.Errorf("error %q does not mention %q", err, tt.wantHas)
			}
		})
	}
}

func TestLoadProximityLedgerRejectsAnInventedKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "proximity.json")
	if err := os.WriteFile(path, []byte(`{"entries":[],"epsilon":1e-6}`), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	if _, err := LoadProximityLedger(path); err == nil {
		t.Fatal("LoadProximityLedger accepted a key nothing parses")
	}
}
