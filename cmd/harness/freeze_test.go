package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
	"github.com/glw907/dubplate/internal/harness/oracle"
)

// fakeShimPath resolves internal/harness/oracle's own fakeshim
// fixture relative to this package, the same shim its own tests run
// against: freeze is a thin wrapper over oracle.Derive, so its own
// tests need nothing more than the same fake shim proves the adapter
// against.
func fakeShimPath(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "..", "internal", "harness", "oracle", "testdata", "fakeshim"))
	if err != nil {
		t.Fatalf("resolve fakeshim path: %v", err)
	}
	return path
}

// seedFreezeCase writes a minimal, LoadCase-valid case at dir (one
// item, one logged request and response) and returns it.
func seedFreezeCase(t *testing.T, dir string) *cassette.Case {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), []byte(`{"release":{}}`), 0o644); err != nil {
		t.Fatalf("write response fixture: %v", err)
	}

	c := &cassette.Case{
		ID:       filepath.Base(dir),
		Profile:  "estate",
		EvalDate: "2024-01-15",
		Input: cassette.Input{
			Items: []cassette.ItemFields{{"title": "Track One"}},
		},
		Requests: []cassette.Request{
			{Method: "GET", PathQuery: "/release/00000000-0000-0000-0000-000000000000?fmt=json", Response: "responses/001.json"},
		},
	}
	if err := c.Write(dir); err != nil {
		t.Fatalf("write case: %v", err)
	}
	return c
}

// stripExpected removes case.json's "expected" key so a test can
// compare the rest of the document without coupling to the shim's own
// determinism, already proven in internal/harness/oracle.
func stripExpected(t *testing.T, data []byte) string {
	t.Helper()
	var m map[string]json.RawMessage
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("parse case.json: %v", err)
	}
	delete(m, "expected")
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("re-marshal case.json without expected: %v", err)
	}
	return string(out)
}

func TestFreezeWritesRecordThroughCaseWrite(t *testing.T) {
	dir := t.TempDir()
	seedFreezeCase(t, dir)

	if err := runFreeze(context.Background(), dir, oracle.Options{Shim: fakeShimPath(t)}); err != nil {
		t.Fatalf("runFreeze: %v", err)
	}

	c, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("LoadCase after freeze: %v", err)
	}
	if c.Expected == nil {
		t.Fatal("Expected is nil after freeze")
	}
	if c.Expected.Recommendation != "strong" {
		t.Errorf("Recommendation = %q, want strong", c.Expected.Recommendation)
	}
	if c.Expected.Thresholds.StrongRecThresh != 0.04 {
		t.Errorf("Thresholds.StrongRecThresh = %v, want 0.04 (the echo should have decoded)", c.Expected.Thresholds.StrongRecThresh)
	}
}

func TestFreezeRefusesToOverwriteOnMisses(t *testing.T) {
	dir := t.TempDir()
	c := seedFreezeCase(t, dir)
	c.Expected = &cassette.Expected{Recommendation: "strong"} // stands in for an already-committed record
	if err := c.Write(dir); err != nil {
		t.Fatalf("seed committed record: %v", err)
	}

	t.Setenv("FAKESHIM_TRIGGER_MISS", "1")
	err := runFreeze(context.Background(), dir, oracle.Options{Shim: fakeShimPath(t)})
	if err == nil {
		t.Fatal("runFreeze with an unanswered request: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "half-derived") {
		t.Errorf("error %q does not explain the miss refusal", err)
	}

	if _, statErr := os.Stat(filepath.Join(dir, missFileName)); statErr != nil {
		t.Errorf("%s missing after a derivation with an unanswered request: %v", missFileName, statErr)
	}

	reloaded, err := cassette.LoadCase(dir)
	if err != nil {
		t.Fatalf("reload case: %v", err)
	}
	if reloaded.Expected == nil || reloaded.Expected.Recommendation != "strong" {
		t.Errorf("committed record was overwritten despite the miss refusal: %+v", reloaded.Expected)
	}
}

// TestFreezeIsIdempotentExceptExpected covers Case.Write's own
// re-marshal hazard: an ItemFields float loses its .0 spelling on
// re-encode, so freezing an already-normalized case a second time
// must change nothing about it. "expected" is excluded from the
// comparison deliberately: the shim's own determinism is
// internal/harness/oracle's proof (TestShimStdoutIsByteIdentical),
// not this test's.
func TestFreezeIsIdempotentExceptExpected(t *testing.T) {
	dir := t.TempDir()
	seedFreezeCase(t, dir)
	opts := oracle.Options{Shim: fakeShimPath(t)}

	if err := runFreeze(context.Background(), dir, opts); err != nil {
		t.Fatalf("first freeze: %v", err)
	}
	first, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatalf("read case.json after first freeze: %v", err)
	}

	if err := runFreeze(context.Background(), dir, opts); err != nil {
		t.Fatalf("second freeze: %v", err)
	}
	second, err := os.ReadFile(filepath.Join(dir, "case.json"))
	if err != nil {
		t.Fatalf("read case.json after second freeze: %v", err)
	}

	if got, want := stripExpected(t, second), stripExpected(t, first); got != want {
		t.Errorf("freeze is not idempotent outside 'expected':\nfirst:  %s\nsecond: %s", want, got)
	}
}

func TestFreezeCmdRequiresCaseFlag(t *testing.T) {
	cmd := newFreezeCmd()
	cmd.SetArgs(nil)
	if err := cmd.Execute(); err == nil {
		t.Fatal("freeze without --case: got nil error, want one")
	}
}
