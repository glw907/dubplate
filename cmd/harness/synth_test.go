package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// writeSpecFile writes spec (a JSON object) to a scratch file and
// returns its path.
func writeSpecFile(t *testing.T, spec string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "spec.json")
	if err := os.WriteFile(path, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec file: %v", err)
	}
	return path
}

const minimalSpecJSON = `{
  "album": "Test Album",
  "artist": "Test Artist",
  "date": "1999-03-02",
  "country": "US",
  "mb_albumid": "aaaaaaaa-1111-2222-3333-444444444444",
  "profile": "estate",
  "eval_date": "2024-01-15",
  "media": [
    {
      "format": "CD",
      "tracks": [
        {"title": "First Track", "length": 200.5, "position": 1},
        {"title": "Second Track", "length": 180.0, "position": 2}
      ]
    }
  ]
}`

func TestRunSynthWritesLoadableCase(t *testing.T) {
	specPath := writeSpecFile(t, minimalSpecJSON)
	out := filepath.Join(t.TempDir(), "test-case")

	if err := runSynth(specPath, out); err != nil {
		t.Fatalf("runSynth: %v", err)
	}

	c, err := cassette.LoadCase(out)
	if err != nil {
		t.Fatalf("LoadCase on synth's own output: %v", err)
	}
	if c.ID != "test-case" {
		t.Errorf("case id = %q, want the --out directory's base name", c.ID)
	}
	if len(c.Input.Items) != 2 {
		t.Errorf("items = %d, want 2", len(c.Input.Items))
	}
	if len(c.Requests) != 1 || c.Requests[0].Response != "responses/001.json" {
		t.Errorf("requests = %+v, want one entry naming responses/001.json", c.Requests)
	}

	if _, err := os.Stat(filepath.Join(out, "responses", "001.json")); err != nil {
		t.Errorf("responses/001.json missing: %v", err)
	}
}

func TestRunSynthRefusesToOverwriteAnExistingCase(t *testing.T) {
	specPath := writeSpecFile(t, minimalSpecJSON)
	out := filepath.Join(t.TempDir(), "test-case")

	if err := runSynth(specPath, out); err != nil {
		t.Fatalf("first runSynth: %v", err)
	}
	err := runSynth(specPath, out)
	if err == nil {
		t.Fatal("second runSynth against the same --out: got nil error, want a refusal")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("error %q does not explain the overwrite refusal", err)
	}
}

func TestRunSynthRejectsAMissingSpecFile(t *testing.T) {
	out := filepath.Join(t.TempDir(), "test-case")
	err := runSynth(filepath.Join(t.TempDir(), "does-not-exist.json"), out)
	if err == nil {
		t.Fatal("runSynth against a missing spec file: got nil error, want one")
	}
}

func TestRunSynthRejectsAnInvalidSpec(t *testing.T) {
	specPath := writeSpecFile(t, `{"album": "Only An Album"}`)
	out := filepath.Join(t.TempDir(), "test-case")
	err := runSynth(specPath, out)
	if err == nil {
		t.Fatal("runSynth against an incomplete spec: got nil error, want one")
	}
}

func TestSynthCmdRequiresSpecAndOutFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{"missing both", nil},
		{"missing out", []string{"--spec", "spec.json"}},
		{"missing spec", []string{"--out", "corpus/staging/x"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newSynthCmd()
			cmd.SetArgs(tt.args)
			if err := cmd.Execute(); err == nil {
				t.Fatalf("synth with %s: got nil error, want one", tt.name)
			}
		})
	}
}
