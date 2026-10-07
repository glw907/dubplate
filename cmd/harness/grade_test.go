package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/grade"
)

const subsonicData = "../../internal/harness/subsonic"

func seedGrader(t *testing.T, ledger string) *grade.Grader {
	t.Helper()
	g, err := grade.New(grade.Config{
		Ledger:   ledger,
		Sessions: subsonicData + "/data/sessions",
		OpenAPI:  subsonicData + "/data/openapi",
		Manifest: subsonicData + "/data/fixture/manifest.json",
	})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

// TestRunGradeWritesTheReportAndFailsOnUnledgered grades the seed session
// under the grade package's test ledger, which leaves the two oracles'
// unadjudicated differences unledgered.
func TestRunGradeWritesTheReportAndFailsOnUnledgered(t *testing.T) {
	out := filepath.Join(t.TempDir(), "report.txt")
	var status bytes.Buffer
	g := seedGrader(t, subsonicData+"/grade/testdata/seed-ledger.json")
	err := runGrade(t.Context(), g, subsonicData+"/data/sessions/seed", [2]string{"navidrome", "gonic"}, out, &status)
	if !errors.Is(err, errUnledgered) {
		t.Fatalf("err = %v, want errUnledgered", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	want, err := os.ReadFile(subsonicData + "/grade/testdata/seed-report.golden")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Error("the written report differs from the grade package's golden")
	}
	if !strings.Contains(status.String(), out) {
		t.Errorf("status %q does not name the report", status.String())
	}
}

func TestRunGradeRefusalWritesNoReport(t *testing.T) {
	out := filepath.Join(t.TempDir(), "report.txt")
	g := seedGrader(t, subsonicData+"/data/ledger.json")
	err := runGrade(t.Context(), g, subsonicData+"/data/sessions/seed", [2]string{"navidrome", "dubplate"}, out, &bytes.Buffer{})
	if err == nil || !strings.Contains(err.Error(), `no source "dubplate"`) {
		t.Fatalf("err = %v", err)
	}
	if _, err := os.Stat(out); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("a refused grade wrote %s", out)
	}
}

func TestGradeNeedsItsFlags(t *testing.T) {
	cmd := newGradeCmd()
	cmd.SetArgs([]string{"--session", "x", "--pair", "navidrome"})
	if err := cmd.ExecuteContext(t.Context()); err == nil || !strings.Contains(err.Error(), "--pair naming two sources") {
		t.Errorf("err = %v", err)
	}
}

func TestGradeRegistered(t *testing.T) {
	cmd, _, err := newRootCmd().Find([]string{"grade"})
	if err != nil || cmd.Name() != "grade" {
		t.Fatalf("Find(grade) = %v, %v", cmd, err)
	}
}
