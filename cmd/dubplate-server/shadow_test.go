package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/shadow"
	"github.com/glw907/dubplate/internal/harness/analysis"
	"github.com/glw907/dubplate/internal/harness/cassette"
)

// shadowTestRegistry returns a registry carrying the shadow
// integration class and its two thresholds, mirroring the production
// wiring newRootCmd builds, so a test that runs the shadow subcommand
// directly (rather than through newRootCmd) sees the same registry
// shape.
func shadowTestRegistry() *config.Registry {
	reg := config.NewRegistry()
	registerShadowIntegration(reg)
	registerShadowThresholds(reg)
	return reg
}

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

	firstRow, secondRow := completePendingRow(), completePendingRow()
	firstRow.ID, secondRow.ID = "obs-div-1", "obs-div-2"
	res := &shadow.PassResult{
		Fixtures: []shadow.CaseFixture{
			{ObservationID: "obs-agree", Profile: "estate", EvalDate: "2026-09-10"},
			{ObservationID: "obs-div-1", Profile: "estate", EvalDate: "2026-09-10"},
			{ObservationID: "obs-div-2", Profile: "estate", EvalDate: "2026-09-10"},
		},
		Pending: []shadow.PendingRow{firstRow, secondRow},
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
		if !slices.Contains(pendingIDs, want) {
			t.Errorf("PendingIDs = %v, want it to contain %q", pendingIDs, want)
		}
	}
}

// baseServerEnviron returns a minimal environ with every owned-now key
// set to a valid value under root, so a test can override or add
// exactly the entries its case needs.
func baseServerEnviron(root string) []string {
	return []string{
		"DUBPLATE_LIBRARY_ROOT=" + filepath.Join(root, "library"),
		"DUBPLATE_STATE_DIR=" + filepath.Join(root, "state"),
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
	}
}

// shadowEnabledEnviron adds the shadow integration's enable flag and
// its three required-when-enabled paths under root.
func shadowEnabledEnviron(root string) []string {
	return []string{
		"DUBPLATE_SHADOW_ENABLED=true",
		"DUBPLATE_SHADOW_OBSERVE_DIR=" + filepath.Join(root, "inbox"),
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR=" + filepath.Join(root, "incumbent"),
		"DUBPLATE_SHADOW_REVIEW_DIR=" + filepath.Join(root, "review"),
	}
}

// TestObserveRefusesAnUnstampedBuild proves observe refuses on "dev"
// and on an empty matcher commit, naming it, and creates no store: a
// refused pass must leave no half-started clock.
func TestObserveRefusesAnUnstampedBuild(t *testing.T) {
	tests := []struct{ name, stamp string }{
		{"dev refuses", "dev"},
		{"empty refuses", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			environ := append(baseServerEnviron(root), shadowEnabledEnviron(root)...)
			var stdout, stderr bytes.Buffer

			cmd := newShadowCmd(environ, shadowTestRegistry(), tt.stamp, &stdout, &stderr)
			cmd.SetArgs([]string{"observe"})
			err := cmd.Execute()
			if err == nil {
				t.Fatal("Execute: got nil error, want a refusal naming the matcher commit")
			}
			if !strings.Contains(err.Error(), "matcher") {
				t.Errorf("error = %q, want it to name the matcher commit", err.Error())
			}

			if _, statErr := os.Stat(filepath.Join(root, "state", "shadow")); !os.IsNotExist(statErr) {
				t.Errorf("store dir exists after a refused observe: %v", statErr)
			}
		})
	}
}

// TestObserveRefusesWhenTheIntegrationIsDisabled proves observe reports
// the missing flag rather than dereferencing a nil shadowFromConfig
// result.
func TestObserveRefusesWhenTheIntegrationIsDisabled(t *testing.T) {
	root := t.TempDir()
	environ := baseServerEnviron(root)
	var stdout, stderr bytes.Buffer

	cmd := newShadowCmd(environ, shadowTestRegistry(), "cafebabe", &stdout, &stderr)
	cmd.SetArgs([]string{"observe"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("Execute: got nil error, want a refusal naming DUBPLATE_SHADOW_ENABLED")
	}
	if !strings.Contains(err.Error(), "DUBPLATE_SHADOW_ENABLED") {
		t.Errorf("error = %q, want it to name DUBPLATE_SHADOW_ENABLED", err.Error())
	}
}

// TestShadowVerbsAreRegisteredAndPrintToStdout proves newShadowCmd
// carries exactly three verbs, both silence flags are set on the
// parent and every child, and report writes to the injected stdout
// writer rather than to os.Stdout.
func TestShadowVerbsAreRegisteredAndPrintToStdout(t *testing.T) {
	root := t.TempDir()
	environ := baseServerEnviron(root)
	var stdout, stderr bytes.Buffer

	cmd := newShadowCmd(environ, shadowTestRegistry(), "cafebabe", &stdout, &stderr)
	if !cmd.SilenceUsage || !cmd.SilenceErrors {
		t.Errorf("shadow: SilenceUsage=%v SilenceErrors=%v, want both true", cmd.SilenceUsage, cmd.SilenceErrors)
	}

	var names []string
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
		if !sub.SilenceUsage || !sub.SilenceErrors {
			t.Errorf("%s: SilenceUsage=%v SilenceErrors=%v, want both true", sub.Name(), sub.SilenceUsage, sub.SilenceErrors)
		}
	}
	want := []string{"observe", "report", "reevaluate"}
	if len(names) != len(want) {
		t.Fatalf("subcommands = %v, want %v", names, want)
	}
	for _, w := range want {
		if !slices.Contains(names, w) {
			t.Errorf("subcommands = %v, want it to contain %q", names, w)
		}
	}

	cmd.SetArgs([]string{"report"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute report: %v", err)
	}
	if stdout.Len() == 0 {
		t.Error("report wrote nothing to the injected stdout writer; mutation 5 (os.Stdout instead) would leave it empty")
	}
}

// TestReportPrintsThePinnedRowForm proves every printed line matches
// doctor's own "%-32s %s" row form.
func TestReportPrintsThePinnedRowForm(t *testing.T) {
	root := t.TempDir()
	environ := baseServerEnviron(root)
	var stdout, stderr bytes.Buffer

	cmd := newShadowCmd(environ, shadowTestRegistry(), "cafebabe", &stdout, &stderr)
	cmd.SetArgs([]string{"report"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute report: %v", err)
	}

	wantLines := []string{
		fmt.Sprintf("%-32s %s", "clock start", "not started"),
		fmt.Sprintf("%-32s %s", "floor", (720 * time.Hour).String()),
		fmt.Sprintf("%-32s %s", "observed albums", "0 applied, 0 skipped"),
		fmt.Sprintf("%-32s %s", "report gaps", "0"),
		fmt.Sprintf("%-32s %s", "go-applies-oracle-skipped", "0 false confident matches (owner disposition joined off-box)"),
		fmt.Sprintf("%-32s %s", "divergences by disposition", "joined off-box against corpus/analysis/adjudications.json"),
		fmt.Sprintf("%-32s %s", "staged pending", "0"),
		fmt.Sprintf("%-32s %s", "resolved", "0"),
		fmt.Sprintf("%-32s %s", "budget total", "0"),
	}
	output := stdout.String()
	for _, want := range wantLines {
		if !strings.Contains(output, want+"\n") {
			t.Errorf("report output missing pinned-form line %q; got:\n%s", want, output)
		}
	}
}

// TestReportPrintsMapRowsInSortedKeyOrder proves the excluded,
// unreconciled, deferred, divergences and matcher-commits rows print
// in sorted key order rather than map-iteration order, which is what
// keeps the table reproducible across runs (mutation 4).
func TestReportPrintsMapRowsInSortedKeyOrder(t *testing.T) {
	root := t.TempDir()
	environ := baseServerEnviron(root)

	storeRoot := filepath.Join(root, "state", "shadow")
	s, err := shadow.OpenStore(storeRoot)
	if err != nil {
		t.Fatalf("OpenStore: %v", err)
	}
	pass := &shadow.PassRecord{
		ID: "pass-1",
		Excluded: map[string]int{
			shadow.ExcludedNoTracks:     1,
			shadow.ExcludedGateRejected: 1,
			shadow.ExcludedMultiAlbum:   1,
		},
	}
	if err := s.PutPass(pass); err != nil {
		t.Fatalf("PutPass: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd := newShadowCmd(environ, shadowTestRegistry(), "cafebabe", &stdout, &stderr)
	cmd.SetArgs([]string{"report"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute report: %v", err)
	}

	output := stdout.String()
	wantOrder := []string{
		"excluded " + shadow.ExcludedGateRejected,
		"excluded " + shadow.ExcludedMultiAlbum,
		"excluded " + shadow.ExcludedNoTracks,
	}
	var positions []int
	for _, label := range wantOrder {
		idx := strings.Index(output, fmt.Sprintf("%-32s", label))
		if idx < 0 {
			t.Fatalf("output missing label %q; got:\n%s", label, output)
		}
		positions = append(positions, idx)
	}
	for i := 1; i < len(positions); i++ {
		if positions[i] <= positions[i-1] {
			t.Errorf("excluded rows out of sorted order: positions = %v for labels %v", positions, wantOrder)
		}
	}
}

// TestLoadShadowIntegration proves ruling 1's required-when-enabled
// semantics: a shadow key set with the flag unset is a load error
// naming the flag, every required path is reported at once when the
// flag is set, a threshold row is accepted and never required, and an
// unparseable flag is an error rather than a silent off.
//
// Moved verbatim from internal/config/config_test.go (Resolution R5):
// the eight rows and every wantErr assertion are unchanged. Only the
// two assertions that used to reach into cfg.Shadow are reworded to
// reach the resolved values by class name, and reg is built through
// shadowTestRegistry rather than config.NewRegistry, since the shadow
// class now registers from this file rather than from
// internal/config's own default registry.
func TestLoadShadowIntegration(t *testing.T) {
	base := []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/library",
		"DUBPLATE_STATE_DIR=/var/lib/x",
		"DUBPLATE_SITE_NAME=Example",
		"DUBPLATE_ADMIN_EMAIL=admin@example.test",
	}
	shadowPaths := []string{
		"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox",
		"DUBPLATE_SHADOW_ORACLE_STATE_DIR=/var/lib/incumbent",
		"DUBPLATE_SHADOW_REVIEW_DIR=/srv/review",
	}

	tests := []struct {
		name       string
		extra      []string
		wantErr    []string // substrings every one of which must appear
		wantShadow bool
	}{
		{
			name:  "flag unset, no shadow keys: loads with a nil Shadow",
			extra: nil,
		},
		{
			name:    "a shadow key set with the flag unset names the flag",
			extra:   []string{"DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ENABLED"},
		},
		{
			name:    "flag set with no paths reports every missing row at once",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=true"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ORACLE_STATE_DIR", "DUBPLATE_SHADOW_REVIEW_DIR"},
		},
		{
			name:       "flag set with every path resolves Shadow",
			extra:      append([]string{"DUBPLATE_SHADOW_ENABLED=true"}, shadowPaths...),
			wantShadow: true,
		},
		{
			name:       "a threshold row is accepted and never required",
			extra:      append([]string{"DUBPLATE_SHADOW_ENABLED=true", "DUBPLATE_SHADOW_FLOOR=48h"}, shadowPaths...),
			wantShadow: true,
		},
		{
			name:    "an unparseable flag is an error rather than off",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=maybe"},
			wantErr: []string{"DUBPLATE_SHADOW_ENABLED"},
		},
		// The two rows fold M-d adds. An adopter who turns the
		// integration off explicitly must be able to start the
		// server, and a shadow key set beside an off flag must
		// still name the flag.
		{
			name:  "an explicit false loads clean with a nil Shadow",
			extra: []string{"DUBPLATE_SHADOW_ENABLED=false"},
		},
		{
			name:    "an explicit false with a shadow key still names the flag",
			extra:   []string{"DUBPLATE_SHADOW_ENABLED=false", "DUBPLATE_SHADOW_OBSERVE_DIR=/srv/inbox"},
			wantErr: []string{"DUBPLATE_SHADOW_OBSERVE_DIR", "DUBPLATE_SHADOW_ENABLED"},
		},
	}

	reg := shadowTestRegistry()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.Load(append(slices.Clone(base), tt.extra...), reg)
			if len(tt.wantErr) == 0 {
				if err != nil {
					t.Fatalf("Load: %v", err)
				}
				if cfg.IntegrationEnabled(shadowIntegrationClass.Name) != tt.wantShadow {
					t.Fatalf("Shadow non-nil = %v, want %v", cfg.IntegrationEnabled(shadowIntegrationClass.Name), tt.wantShadow)
				}
				if tt.wantShadow && cfg.IntegrationValue(shadowIntegrationClass.Name, "DUBPLATE_SHADOW_OBSERVE_DIR") != "/srv/inbox" {
					t.Errorf("ObserveDir = %q, want %q", cfg.IntegrationValue(shadowIntegrationClass.Name, "DUBPLATE_SHADOW_OBSERVE_DIR"), "/srv/inbox")
				}
				return
			}
			if err == nil {
				t.Fatal("Load returned nil; want an error")
			}
			for _, want := range tt.wantErr {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("error does not name %s: %v", want, err)
				}
			}
			if tt.name == "an unparseable flag is an error rather than off" {
				if n := strings.Count(err.Error(), "DUBPLATE_SHADOW_ENABLED"); n != 1 {
					t.Errorf("error names DUBPLATE_SHADOW_ENABLED %d times, want exactly 1: %v", n, err)
				}
			}
		})
	}
}
