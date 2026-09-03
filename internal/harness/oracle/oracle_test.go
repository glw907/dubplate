package oracle

import (
	"context"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/cassette"
)

// fakeShim resolves testdata/fakeshim relative to this package.
func fakeShim(t *testing.T) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", "fakeshim"))
	if err != nil {
		t.Fatalf("resolve fakeshim path: %v", err)
	}
	return path
}

// testCase builds a minimal case (one item, one logged request) with
// its response file written under a fresh case directory, and points
// caseDir at that directory for the duration of the test.
func testCase(t *testing.T, evalDate string) *cassette.Case {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "responses"), 0o755); err != nil {
		t.Fatalf("mkdir responses: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "responses", "001.json"), []byte(`{"release":{}}`), 0o644); err != nil {
		t.Fatalf("write response fixture: %v", err)
	}

	c := &cassette.Case{
		ID:       "oracle-test",
		Profile:  "estate",
		EvalDate: evalDate,
		Input: cassette.Input{
			Items: []cassette.ItemFields{{"title": "Track One"}},
		},
		Requests: []cassette.Request{
			{Method: "GET", PathQuery: "/release/00000000-0000-0000-0000-000000000000?fmt=json", Response: "responses/001.json"},
		},
	}

	orig := caseDir
	caseDir = func(*cassette.Case) (string, error) { return dir, nil }
	t.Cleanup(func() { caseDir = orig })
	return c
}

func TestDeriveHitDecodesExpected(t *testing.T) {
	c := testCase(t, "2024-01-15")
	exp, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}
	if exp.Recommendation != "strong" {
		t.Errorf("Recommendation = %q, want strong", exp.Recommendation)
	}
	if exp.Action != "apply" {
		t.Errorf("Action = %q, want apply", exp.Action)
	}
	if exp.EchoedDate != "2024-01-15" {
		t.Errorf("EchoedDate = %q, want 2024-01-15", exp.EchoedDate)
	}
	if exp.ConfigHash == "" {
		t.Error("ConfigHash is empty on a valid shim document")
	}
	if len(exp.Candidates) != 1 || exp.Candidates[0].Source != "MusicBrainz" {
		t.Errorf("Candidates = %+v, want one MusicBrainz candidate", exp.Candidates)
	}
}

// TestDeriveDecodesWeightsAndThresholds proves the echo fields decode
// against the fake shim's actual output, including the max_rec hazard
// (a mapping of penalty-key name to recommendation name beside three
// plain floats) that a flat map[string]float64 cannot represent.
func TestDeriveDecodesWeightsAndThresholds(t *testing.T) {
	c := testCase(t, "2024-01-15")
	exp, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err != nil {
		t.Fatalf("Derive: %v", err)
	}

	if got := exp.Weights["source"]; got != 2.0 {
		t.Errorf("Weights[source] = %v, want 2.0", got)
	}
	if got := exp.Weights["album"]; got != 3.0 {
		t.Errorf("Weights[album] = %v, want 3.0", got)
	}
	if exp.Thresholds.StrongRecThresh != 0.04 {
		t.Errorf("Thresholds.StrongRecThresh = %v, want 0.04", exp.Thresholds.StrongRecThresh)
	}
	if exp.Thresholds.MediumRecThresh != 0.25 {
		t.Errorf("Thresholds.MediumRecThresh = %v, want 0.25", exp.Thresholds.MediumRecThresh)
	}
	if exp.Thresholds.RecGapThresh != 0.25 {
		t.Errorf("Thresholds.RecGapThresh = %v, want 0.25", exp.Thresholds.RecGapThresh)
	}
	wantMaxRec := map[string]string{"missing_tracks": "medium", "unmatched_tracks": "medium"}
	if !maps.Equal(exp.Thresholds.MaxRec, wantMaxRec) {
		t.Errorf("Thresholds.MaxRec = %v, want %v", exp.Thresholds.MaxRec, wantMaxRec)
	}
}

func TestDeriveRejectsMissingThresholds(t *testing.T) {
	c := testCase(t, "2024-01-15")
	t.Setenv("FAKESHIM_OMIT_THRESHOLDS", "1")
	_, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err == nil {
		t.Fatal("Derive with a shim that omits thresholds: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "thresholds") {
		t.Errorf("error %q does not name the missing thresholds", err)
	}
}

func TestDeriveRejectsEchoedDateMismatch(t *testing.T) {
	c := testCase(t, "2024-06-01") // fakeshim always echoes 2024-01-15
	_, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err == nil {
		t.Fatal("Derive with a date the shim never echoed: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "echoed date") {
		t.Errorf("error %q does not name the echoed-date mismatch", err)
	}
}

func TestDeriveRejectsMissingWeights(t *testing.T) {
	c := testCase(t, "2024-01-15")
	t.Setenv("FAKESHIM_OMIT_WEIGHTS", "1")
	_, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err == nil {
		t.Fatal("Derive with a shim that omits weights: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "weights") {
		t.Errorf("error %q does not name the missing weights", err)
	}
}

func TestDeriveRejectsMissingConfigHash(t *testing.T) {
	c := testCase(t, "2024-01-15")
	t.Setenv("FAKESHIM_OMIT_HASH", "1")
	_, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err == nil {
		t.Fatal("Derive with a shim that omits config_hash: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "config hash") {
		t.Errorf("error %q does not name the missing config hash", err)
	}
}

func TestDeriveRejectsMultiSourceProfile(t *testing.T) {
	origProfile := loadProfile
	loadProfile = func(string) ([]byte, error) {
		return []byte("plugins: musicbrainz discogs\n"), nil
	}
	t.Cleanup(func() { loadProfile = origProfile })

	c := testCase(t, "2024-01-15")
	_, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err == nil {
		t.Fatal("Derive against a two-source profile: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "metadata sources") {
		t.Errorf("error %q does not name the metadata-source count", err)
	}
}

func TestDeriveRejectsZeroSourceProfile(t *testing.T) {
	origProfile := loadProfile
	loadProfile = func(string) ([]byte, error) {
		return []byte("import:\n  timid: no\n"), nil // no plugins: line at all
	}
	t.Cleanup(func() { loadProfile = origProfile })

	c := testCase(t, "2024-01-15")
	_, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)})
	if err == nil {
		t.Fatal("Derive against a profile with no plugins: line: got nil error, want one")
	}
	if !strings.Contains(err.Error(), "metadata sources") {
		t.Errorf("error %q does not name the metadata-source count", err)
	}
}

func TestDeriveLaunchesExactlyOneShimProcess(t *testing.T) {
	c := testCase(t, "2024-01-15")
	callLog := filepath.Join(t.TempDir(), "calls.log")
	t.Setenv("FAKESHIM_CALL_LOG", callLog)

	if _, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)}); err != nil {
		t.Fatalf("Derive: %v", err)
	}

	data, err := os.ReadFile(callLog)
	if err != nil {
		t.Fatalf("read call log: %v", err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("shim ran %d times, want exactly 1: %v", len(lines), lines)
	}
}

func TestDeriveSetsPythonHashSeedZero(t *testing.T) {
	c := testCase(t, "2024-01-15")
	envLog := filepath.Join(t.TempDir(), "env.log")
	t.Setenv("FAKESHIM_ENV_LOG", envLog)

	if _, err := Derive(context.Background(), c, Options{Shim: fakeShim(t)}); err != nil {
		t.Fatalf("Derive: %v", err)
	}

	data, err := os.ReadFile(envLog)
	if err != nil {
		t.Fatalf("read env log: %v", err)
	}
	if got := strings.TrimSpace(string(data)); got != "0" {
		t.Errorf("PYTHONHASHSEED in shim environment = %q, want 0", got)
	}
}

func TestRenderProfileSubstitutesHostAndBoxPaths(t *testing.T) {
	rendered, err := renderProfile("estate", "/scratch/lib", "/scratch/lib/library.db", "127.0.0.1:54321")
	if err != nil {
		t.Fatalf("renderProfile: %v", err)
	}
	text := string(rendered)
	if !strings.Contains(text, `directory: "/scratch/lib"`) {
		t.Errorf("rendered profile did not substitute the directory placeholder:\n%s", text)
	}
	if !strings.Contains(text, `library: "/scratch/lib/library.db"`) {
		t.Errorf("rendered profile did not substitute the library placeholder:\n%s", text)
	}
	if !strings.Contains(text, `host: "127.0.0.1:54321"`) ||
		!strings.Contains(text, "https: no") ||
		!strings.Contains(text, "ratelimit: 100") {
		t.Errorf("rendered profile is missing the stub musicbrainz block:\n%s", text)
	}
}

func TestRenderProfileStockIsUntouchedByBoxPathSubstitution(t *testing.T) {
	rendered, err := renderProfile("stock", "/scratch/lib", "/scratch/lib/library.db", "127.0.0.1:54321")
	if err != nil {
		t.Fatalf("renderProfile: %v", err)
	}
	text := string(rendered)
	if strings.Contains(text, "directory:") || strings.Contains(text, "library:") {
		t.Errorf("stock.yaml should carry no directory/library keys at all:\n%s", text)
	}
	if !strings.Contains(text, "plugins: musicbrainz") {
		t.Errorf("rendered stock profile lost its plugins line:\n%s", text)
	}
}
