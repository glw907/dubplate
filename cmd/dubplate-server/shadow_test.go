package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/engine/shadow"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// TestCaseJSONRoundTripsThroughLoadCase proves caseJSON's bytes are a
// legal cassette.Case: Expected nil, path_query accepted as already
// canonical, and Profile/EvalDate/Requests carried through untouched.
func TestCaseJSONRoundTripsThroughLoadCase(t *testing.T) {
	f := shadow.CaseFixture{
		ObservationID: "obs-1",
		Profile:       "estate",
		EvalDate:      "2026-09-10",
		Items: []map[string]any{
			{"title": "Track One", "artist": "Some Artist"},
		},
		Requests: []shadow.RecordedRequest{
			{Method: "GET", PathQuery: "/ws/2/release-group/?fmt=json&query=x", Response: "responses/001.json"},
		},
	}

	data, err := caseJSON(f)
	if err != nil {
		t.Fatalf("caseJSON: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "case.json"), data, 0o644); err != nil {
		t.Fatalf("write case.json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write response: %v", err)
	}

	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase: %v", err)
	}
	if c.Expected != nil {
		t.Errorf("Expected = %+v, want nil", c.Expected)
	}
	if c.Profile != "estate" {
		t.Errorf("Profile = %q, want estate", c.Profile)
	}
	if c.EvalDate != "2026-09-10" {
		t.Errorf("EvalDate = %q, want 2026-09-10", c.EvalDate)
	}
	if len(c.Requests) != 1 || c.Requests[0].Method != "GET" ||
		c.Requests[0].PathQuery != "/ws/2/release-group/?fmt=json&query=x" ||
		c.Requests[0].Response != "responses/001.json" {
		t.Errorf("Requests = %+v, want one round-tripped request", c.Requests)
	}
}

// TestCaseJSONNonCanonicalPathQueryFailsToLoad proves caseJSON does not
// itself canonicalize a path_query: a fixture carrying an
// already-non-canonical query round-trips into bytes cassette.LoadCase
// refuses, which is what makes mutation (4) (writing a non-canonical
// path_query) detectable.
func TestCaseJSONNonCanonicalPathQueryFailsToLoad(t *testing.T) {
	f := shadow.CaseFixture{
		ObservationID: "obs-2",
		Profile:       "estate",
		EvalDate:      "2026-09-10",
		Requests: []shadow.RecordedRequest{
			{Method: "GET", PathQuery: "/ws/2/release-group/?query=x&fmt=json", Response: "responses/001.json"},
		},
	}
	data, err := caseJSON(f)
	if err != nil {
		t.Fatalf("caseJSON: %v", err)
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "case.json"), data, 0o644); err != nil {
		t.Fatalf("write case.json: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write response: %v", err)
	}

	if _, err := cassette.LoadCase(dir); err == nil {
		t.Error("LoadCase: got nil error, want a non-canonical path_query rejection")
	}
}

func completePendingRow() shadow.PendingRow {
	return shadow.PendingRow{
		ID:                "obs-1",
		Album:             "uploader/basename",
		ReleaseMBID:       "11111111-1111-1111-1111-111111111111",
		MatcherCommit:     "cafebabe",
		MatcherConfigHash: "deadbeef",
		Suspected:         "go-applies-oracle-skipped: beets skipped the album and the Go matcher applied release 111...",
		Citation:          "shadow observation obs-1, class go-applies-oracle-skipped",
		ReleaseTruth:      "the release under dispute is 111..., proposed by the Go matcher; no truth is ruled for it",
		Reason:            "observed against the incumbent's own live decision at matcher commit cafebabe, pass pass-1",
		Date:              "2026-09-10",
	}
}

// TestPendingJSONValidatesBeforeItReturns proves pendingJSON runs the
// row through analysis.CheckAdjudications before returning: a complete
// row loads and decodes with Disposition and Adjudicator set by the
// marshaller rather than read off the row, and a row missing any of
// three shadow-only-required fields is refused, naming the field.
func TestPendingJSONValidatesBeforeItReturns(t *testing.T) {
	tests := []struct {
		name       string
		mutate     func(shadow.PendingRow) shadow.PendingRow
		wantErr    bool
		wantErrSub string
	}{
		{
			name:    "complete row",
			mutate:  func(r shadow.PendingRow) shadow.PendingRow { return r },
			wantErr: false,
		},
		{
			name: "missing ReleaseMBID",
			mutate: func(r shadow.PendingRow) shadow.PendingRow {
				r.ReleaseMBID = ""
				return r
			},
			wantErr:    true,
			wantErrSub: "release_mbid",
		},
		{
			name: "missing MatcherCommit",
			mutate: func(r shadow.PendingRow) shadow.PendingRow {
				r.MatcherCommit = ""
				return r
			},
			wantErr:    true,
			wantErrSub: "matcher_commit",
		},
		{
			name: "missing Reason",
			mutate: func(r shadow.PendingRow) shadow.PendingRow {
				r.Reason = ""
				return r
			},
			wantErr:    true,
			wantErrSub: "reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := tt.mutate(completePendingRow())
			data, err := pendingJSON(row)
			if tt.wantErr {
				if err == nil {
					t.Fatal("pendingJSON: got nil error, want a validation failure")
				}
				if !strings.Contains(err.Error(), tt.wantErrSub) {
					t.Errorf("pendingJSON error = %q, want it to name %q", err.Error(), tt.wantErrSub)
				}
				return
			}
			if err != nil {
				t.Fatalf("pendingJSON: %v", err)
			}

			var entry analysis.AdjudicationEntry
			if err := json.Unmarshal(data, &entry); err != nil {
				t.Fatalf("decode entry: %v", err)
			}
			if entry.Disposition != analysis.Pending {
				t.Errorf("Disposition = %q, want %q", entry.Disposition, analysis.Pending)
			}
			if entry.Adjudicator != analysis.AgentAdjudicator {
				t.Errorf("Adjudicator = %q, want %q", entry.Adjudicator, analysis.AgentAdjudicator)
			}
			if entry.Subject.Album != row.Album || entry.Subject.ReleaseMBID != row.ReleaseMBID {
				t.Errorf("Subject = %+v, want album/release_mbid from the row", entry.Subject)
			}
		})
	}
}

// TestPendingJSONNeverEmitsATodo proves the four prose fields a pending
// row carries are always non-empty and never the literal word this
// package's own doc forbids: a row that says nothing still consumes an
// owner adjudication (run.go:363-367).
func TestPendingJSONNeverEmitsATodo(t *testing.T) {
	row := completePendingRow()
	data, err := pendingJSON(row)
	if err != nil {
		t.Fatalf("pendingJSON: %v", err)
	}

	var entry analysis.AdjudicationEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("decode entry: %v", err)
	}

	prose := []struct {
		name string
		val  string
	}{
		{"Suspected", entry.Suspected},
		{"Citation", entry.Citation},
		{"ReleaseTruth", entry.ReleaseTruth},
		{"Reason", entry.Reason},
	}
	for _, p := range prose {
		if p.val == "" {
			t.Errorf("%s: empty, want non-empty prose", p.name)
		}
		if strings.Contains(p.val, "TODO") {
			t.Errorf("%s: contains the literal TODO: %q", p.name, p.val)
		}
	}
}

// TestStoreArtifactsWritesOneFilePerObservationAndOnePerNewsRow proves
// storeArtifacts writes one case.json per fixture and one pending row
// per divergence, with no pending row for the agreement class: three
// observations, two divergences among them, gives three case.json
// files and exactly two pending rows.
func TestStoreArtifactsWritesOneFilePerObservationAndOnePerNewsRow(t *testing.T) {
	root := filepath.Join(t.TempDir(), "shadow")
	s, err := shadow.OpenStore(root)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}

	res := &shadow.PassResult{
		Fixtures: []shadow.CaseFixture{
			{ObservationID: "obs-agree", Profile: "estate", EvalDate: "2026-09-10"},
			{ObservationID: "obs-div-1", Profile: "estate", EvalDate: "2026-09-10"},
			{ObservationID: "obs-div-2", Profile: "estate", EvalDate: "2026-09-10"},
		},
		Pending: []shadow.PendingRow{
			func() shadow.PendingRow {
				r := completePendingRow()
				r.ID = "obs-div-1"
				return r
			}(),
			func() shadow.PendingRow {
				r := completePendingRow()
				r.ID = "obs-div-2"
				return r
			}(),
		},
	}

	if err := storeArtifacts(s, res); err != nil {
		t.Fatalf("storeArtifacts: %v", err)
	}

	for _, id := range []string{"obs-agree", "obs-div-1", "obs-div-2"} {
		path := filepath.Join(root, "observations", id, "case.json")
		if _, err := os.Stat(path); err != nil {
			t.Errorf("case.json for %s: %v", id, err)
		}
	}

	pendingIDs, err := s.PendingIDs()
	if err != nil {
		t.Fatalf("PendingIDs: %v", err)
	}
	if len(pendingIDs) != 2 {
		t.Fatalf("PendingIDs = %v, want exactly 2 (one per divergence, none for the agreement)", pendingIDs)
	}
	for _, want := range []string{"obs-div-1", "obs-div-2"} {
		found := false
		for _, id := range pendingIDs {
			if id == want {
				found = true
			}
		}
		if !found {
			t.Errorf("PendingIDs = %v, want it to contain %q", pendingIDs, want)
		}
	}
}
