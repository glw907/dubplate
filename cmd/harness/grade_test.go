package main

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/harness/subsonic/grade"
	"github.com/glw907/dubplate/internal/harness/subsonic/session"
	"github.com/glw907/dubplate/internal/opensubsonic/client"
)

const subsonicData = "../../internal/harness/subsonic"

func seedGrader(t *testing.T, ledger string) *grade.Grader {
	t.Helper()
	g, err := grade.New(grade.Config{
		Ledger:     ledger,
		Sessions:   subsonicData + "/data/sessions",
		OpenAPI:    subsonicData + "/data/openapi",
		Manifest:   subsonicData + "/data/fixture/manifest.json",
		Assignment: subsonicData + "/data/assignment.json",
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

// TestRunGradeExitRule grades a session whose two sources gave the same
// bytes, so the only findings are the contract tier's two userRating
// reports, which two ledger rows cover. A grade with every finding
// ledgered succeeds; a row matching nothing alone makes it fail.
func TestRunGradeExitRule(t *testing.T) {
	const version = "0.64.2 (10114574)"
	const digest = "5c82e274df9df7b57e91df2caa48ff103b6036893cf1e8e6bc187c88d668883c"
	body, err := os.ReadFile(subsonicData + "/data/sessions/seed/012-navidrome.json")
	if err != nil {
		t.Fatal(err)
	}
	sessions := t.TempDir()
	s := &session.Session{
		Sources: map[string]session.Source{
			"navidrome": {LibraryDigest: digest, ServerVersion: version},
			"copy":      {LibraryDigest: digest, ServerVersion: version},
		},
		Requests: []session.Request{{
			Endpoint: "getAlbum", Format: client.JSON,
			Params: []session.Param{{Name: "id", Value: "album:Alpha Artist/Single Disc"}},
			Responses: []session.Response{
				{Source: "navidrome", Status: 200, ContentType: "application/json", Body: body},
				{Source: "copy", Status: 200, ContentType: "application/json", Body: body},
			},
		}},
	}
	if err := session.Write(filepath.Join(sessions, "same"), s); err != nil {
		t.Fatal(err)
	}
	row := func(id string, class grade.Class, source, path string) grade.Row {
		return grade.Row{
			ID: id, Area: "getAlbum", Class: class, Sources: []string{"navidrome", "copy"},
			Oracle: "o", Dubplate: "d", Evidence: "e", State: grade.StateActive, Source: source,
			Pattern:  grade.Pattern{Endpoint: "getAlbum", Format: "json", Path: path},
			Session:  "same",
			Versions: map[string]string{"navidrome": version, "copy": version},
		}
	}
	userRating := "subsonic-response/album/userRating"
	covering := []grade.Row{
		row("navidrome-rating", grade.ClassContract, "navidrome", userRating),
		row("copy-rating", grade.ClassContract, "copy", userRating),
	}
	tests := []struct {
		name  string
		rows  []grade.Row
		stale bool
	}{
		{name: "every finding ledgered", rows: covering},
		{name: "a row matching nothing", rows: append(slices.Clone(covering), row("stale", grade.ClassVolatile, "", "subsonic-response/album/nothing")), stale: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			raw, err := json.Marshal(tt.rows)
			if err != nil {
				t.Fatal(err)
			}
			ledger := filepath.Join(t.TempDir(), "ledger.json")
			if err := os.WriteFile(ledger, raw, 0o644); err != nil {
				t.Fatal(err)
			}
			g, err := grade.New(grade.Config{
				Ledger:     ledger,
				Sessions:   sessions,
				OpenAPI:    subsonicData + "/data/openapi",
				Manifest:   subsonicData + "/data/fixture/manifest.json",
				Assignment: subsonicData + "/data/assignment.json",
			})
			if err != nil {
				t.Fatal(err)
			}
			out := filepath.Join(t.TempDir(), "report.txt")
			err = runGrade(t.Context(), g, filepath.Join(sessions, "same"), [2]string{"navidrome", "copy"}, out, &bytes.Buffer{})
			if tt.stale {
				if !errors.Is(err, grade.ErrStaleRow) || errors.Is(err, errUnledgered) {
					t.Fatalf("err = %v, want a stale row alone", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("err = %v, want success with every finding ledgered", err)
			}
		})
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

// TestGradeResolvesLedgerSessionsFromTheCommittedHome grades a copy of the
// seed session kept outside the committed sessions directory under a
// ledger whose active row names the committed seed session. The row's
// session must resolve against the committed home, never against the
// graded session's parent.
func TestGradeResolvesLedgerSessionsFromTheCommittedHome(t *testing.T) {
	seed := subsonicData + "/data/sessions/seed"
	s, err := session.Load(seed)
	if err != nil {
		t.Fatal(err)
	}
	elsewhere := filepath.Join(t.TempDir(), "elsewhere")
	if err := session.Write(elsewhere, s); err != nil {
		t.Fatal(err)
	}
	row := grade.Row{
		ID: "seed-name", Area: "getAlbum", Class: grade.ClassVolatile, Sources: []string{"navidrome", "gonic"},
		Oracle: "o", Dubplate: "d", Evidence: "e", State: grade.StateActive,
		Pattern: grade.Pattern{Endpoint: "getAlbum", Format: "json", Path: "subsonic-response/album/name"},
		Session: "seed",
		Versions: map[string]string{
			"navidrome": s.Sources["navidrome"].ServerVersion,
			"gonic":     s.Sources["gonic"].ServerVersion,
		},
	}
	raw, err := json.Marshal([]grade.Row{row})
	if err != nil {
		t.Fatal(err)
	}
	ledger := filepath.Join(t.TempDir(), "ledger.json")
	if err := os.WriteFile(ledger, raw, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := newGradeCmd()
	cmd.SetErr(&bytes.Buffer{})
	cmd.SetArgs([]string{
		"--session", elsewhere, "--pair", "navidrome,gonic",
		"--out", filepath.Join(t.TempDir(), "report.txt"), "--ledger", ledger,
	})
	err = cmd.ExecuteContext(t.Context())
	if !errors.Is(err, errUnledgered) || errors.Is(err, grade.ErrInvalidLedger) {
		t.Fatalf("err = %v, want a graded session with unledgered findings", err)
	}
}

// TestDefaultSubsonicPathsNameCommittedFiles builds a grader from the
// default paths alone, so a path that points nowhere fails here.
func TestDefaultSubsonicPathsNameCommittedFiles(t *testing.T) {
	p, err := defaultSubsonicPaths()
	if err != nil {
		t.Fatal(err)
	}
	for name, path := range map[string]string{"assignment": p.assignment, "ledger": p.ledger, "manifest": p.manifest} {
		if _, err := os.Stat(path); err != nil {
			t.Errorf("default %s path: %v", name, err)
		}
	}
	_, err = grade.New(grade.Config{
		Ledger: p.ledger, Sessions: p.sessions, OpenAPI: p.openAPI, Manifest: p.manifest, Assignment: p.assignment,
	})
	if err != nil {
		t.Errorf("grade.New over the default paths: %v", err)
	}
}
