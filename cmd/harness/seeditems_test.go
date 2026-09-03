package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/synth"
)

// writeScratchCase builds a two-track synthetic case via synth.Build,
// clears its Input.Items (a freshly captured case carries a request
// log and nothing else; corpus/README.md's "The capture tool"), and
// writes case.json plus responses/001.json under a fresh temp
// directory. It returns the directory, ready for runSeedItems.
func writeScratchCase(t *testing.T) string {
	t.Helper()
	spec := &synth.ReleaseSpec{
		Album:     "Test Album",
		Artist:    "Test Artist",
		Date:      "1999-03-02",
		Country:   "US",
		MBAlbumID: "aaaaaaaa-1111-2222-3333-444444444444",
		Profile:   "estate",
		EvalDate:  "2024-01-15",
		Media: []synth.Medium{{
			Format: "CD",
			Tracks: []synth.Track{
				{Title: "First Track", Length: 200.5, Position: 1},
				{Title: "Second Track", Length: 180.0, Position: 2},
			},
		}},
	}
	c, doc, err := synth.Build("scratch-case", spec)
	if err != nil {
		t.Fatalf("synth.Build: %v", err)
	}
	c.Input.Items = nil

	dir := filepath.Join(t.TempDir(), "scratch-case")
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("create case dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), doc, 0o644); err != nil {
		t.Fatalf("write response: %v", err)
	}
	if err := c.Write(dir); err != nil {
		t.Fatalf("write case.json: %v", err)
	}
	return dir
}

func TestRunSeedItemsProjectsItemsIntoCaseJSON(t *testing.T) {
	dir := writeScratchCase(t)

	n, err := runSeedItems(dir, "")
	if err != nil {
		t.Fatalf("runSeedItems: %v", err)
	}
	if n != 2 {
		t.Fatalf("projected %d items, want 2", n)
	}

	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase after seed-items: %v", err)
	}
	if len(c.Input.Items) != 2 {
		t.Fatalf("case.json items = %d, want 2", len(c.Input.Items))
	}
	if c.Input.Items[0]["title"] != "First Track" {
		t.Errorf("item 0 title = %v, want First Track", c.Input.Items[0]["title"])
	}
	if c.Input.Items[1]["title"] != "Second Track" {
		t.Errorf("item 1 title = %v, want Second Track", c.Input.Items[1]["title"])
	}
	if c.Profile != "estate" || c.EvalDate != "2024-01-15" {
		t.Errorf("seed-items touched profile/eval_date: profile=%q eval_date=%q, want estate/2024-01-15 untouched", c.Profile, c.EvalDate)
	}
}

func TestRunSeedItemsAppliesOverridesFile(t *testing.T) {
	dir := writeScratchCase(t)
	overridesPath := filepath.Join(t.TempDir(), "overrides.json")
	overrides := []synth.Override{
		{Op: synth.OpField, Index: 0, Fields: map[string]any{"title": "Mistagged Title"}},
	}
	data, err := json.Marshal(overrides)
	if err != nil {
		t.Fatalf("marshal overrides: %v", err)
	}
	if err := os.WriteFile(overridesPath, data, 0o644); err != nil {
		t.Fatalf("write overrides file: %v", err)
	}

	if _, err := runSeedItems(dir, overridesPath); err != nil {
		t.Fatalf("runSeedItems: %v", err)
	}

	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase after seed-items: %v", err)
	}
	if c.Input.Items[0]["title"] != "Mistagged Title" {
		t.Errorf("item 0 title = %v, want the overridden Mistagged Title", c.Input.Items[0]["title"])
	}
	if c.Input.Items[1]["title"] != "Second Track" {
		t.Errorf("item 1 title = %v, want the unperturbed Second Track", c.Input.Items[1]["title"])
	}
}

func TestRunSeedItemsFailsOnAMissingCaseDir(t *testing.T) {
	_, err := runSeedItems(filepath.Join(t.TempDir(), "does-not-exist"), "")
	if err == nil {
		t.Fatal("runSeedItems against a missing case dir: got nil error, want one")
	}
}

func TestRunSeedItemsFailsWhenCaseCarriesNoRequests(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "no-requests-case")
	c := &cassette.Case{ID: "no-requests-case", Profile: "estate", EvalDate: "2024-01-15"}
	if err := c.Write(dir); err != nil {
		t.Fatalf("write case.json: %v", err)
	}

	_, err := runSeedItems(dir, "")
	if err == nil {
		t.Fatal("runSeedItems against a case with no requests: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "no requests") {
		t.Errorf("error %q does not explain the missing request log", err)
	}
}

func TestRunSeedItemsFailsOnAMalformedOverridesFile(t *testing.T) {
	dir := writeScratchCase(t)
	overridesPath := filepath.Join(t.TempDir(), "overrides.json")
	if err := os.WriteFile(overridesPath, []byte("not json"), 0o644); err != nil {
		t.Fatalf("write overrides file: %v", err)
	}

	_, err := runSeedItems(dir, overridesPath)
	if err == nil {
		t.Fatal("runSeedItems with a malformed overrides file: got nil error, want one")
	}
}

func TestSeedItemsCmdRequiresCaseFlag(t *testing.T) {
	cmd := newSeedItemsCmd()
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err == nil {
		t.Fatal("seed-items with no --case: got nil error, want one")
	}
}

// caseDirSnapshot captures dir's own files' mtimes and content, so a
// test can prove a run left the directory byte-for-byte and
// stat-for-stat untouched: `--facts` is read-only by contract, and a
// fixed-in-place mtime is stronger evidence of that than a content
// diff alone (a rewrite that reproduces identical bytes would still
// touch mtime).
type caseDirSnapshot struct {
	modTime map[string]int64
	content map[string][]byte
}

func snapshotCaseDir(t *testing.T, dir string) caseDirSnapshot {
	t.Helper()
	snap := caseDirSnapshot{modTime: map[string]int64{}, content: map[string][]byte{}}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		snap.modTime[path] = info.ModTime().UnixNano()
		snap.content[path] = data
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot case dir: %v", err)
	}
	return snap
}

func TestRunSeedItemsFactsWritesNothing(t *testing.T) {
	dir := writeScratchCase(t)
	if _, err := runSeedItems(dir, ""); err != nil {
		t.Fatalf("runSeedItems (seeding items before the facts run): %v", err)
	}
	before := snapshotCaseDir(t, dir)

	if _, err := runSeedItemsFacts(dir); err != nil {
		t.Fatalf("runSeedItemsFacts: %v", err)
	}

	after := snapshotCaseDir(t, dir)
	if !reflect.DeepEqual(before.content, after.content) {
		t.Errorf("runSeedItemsFacts changed case dir content: before %v, after %v", before.content, after.content)
	}
	if !reflect.DeepEqual(before.modTime, after.modTime) {
		t.Errorf("runSeedItemsFacts touched a case dir file's mtime: before %v, after %v", before.modTime, after.modTime)
	}
}

func TestRunSeedItemsFactsReturnsTheFormattedFacts(t *testing.T) {
	dir := writeScratchCase(t)

	out, err := runSeedItemsFacts(dir)
	if err != nil {
		t.Fatalf("runSeedItemsFacts: %v", err)
	}
	if !strings.Contains(out, "release date: 1999-03-02") {
		t.Errorf("facts output %q does not carry the release date", out)
	}
	if !strings.Contains(out, "aliases: none") {
		t.Errorf("facts output %q does not print an explicit aliases: none line", out)
	}
}

func TestSeedItemsCmdFactsRejectsOverrides(t *testing.T) {
	dir := writeScratchCase(t)

	cmd := newSeedItemsCmd()
	cmd.SetArgs([]string{"--case", dir, "--facts", "--overrides", filepath.Join(t.TempDir(), "overrides.json")})
	if err := cmd.Execute(); err == nil {
		t.Fatal("seed-items --facts with --overrides: got nil error, want one")
	}
}

func TestSeedItemsCmdFactsPrintsToStdout(t *testing.T) {
	dir := writeScratchCase(t)

	cmd := newSeedItemsCmd()
	var stdout bytes.Buffer
	cmd.SetOut(&stdout)
	cmd.SetArgs([]string{"--case", dir, "--facts"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("seed-items --facts: %v", err)
	}
	if !strings.Contains(stdout.String(), "release date: 1999-03-02") {
		t.Errorf("stdout %q does not carry the release date", stdout.String())
	}
}
