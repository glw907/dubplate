package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
)

// installTestUnidecodeTable installs a valid table into stateDir by
// running the install verb against the workstation's cached wheel,
// never by copying bytes: the leaf a doctor test needs already
// installed. It fails rather than skips when the cache is empty, since
// Go's test cache does not track the wheel file and a stale "ok" would
// hide the gap.
func installTestUnidecodeTable(t *testing.T, stateDir string) {
	t.Helper()
	path, err := unidecode.WheelCachePath()
	if err != nil {
		t.Fatalf("locate wheel cache: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("read cached wheel at %s: %v; run make unidecode-wheel", path, err)
	}
	var stdout bytes.Buffer
	environ := []string{"DUBPLATE_STATE_DIR=" + stateDir}
	if err := runUnidecodeInstall(context.Background(), environ, path, &stdout); err != nil {
		t.Fatalf("install test unidecode table into %s: %v", stateDir, err)
	}
}

func TestUnidecodeInstall_RefusesWithoutStateDir(t *testing.T) {
	var stdout bytes.Buffer
	err := runUnidecodeInstall(context.Background(), nil, "", &stdout)
	if err == nil || !strings.Contains(err.Error(), "DUBPLATE_STATE_DIR") {
		t.Fatalf("error = %v, want a refusal naming DUBPLATE_STATE_DIR", err)
	}
	if stdout.Len() != 0 {
		t.Errorf("stdout = %q, want empty", stdout.String())
	}
}

// TestUnidecodeInstall_InstallsUnderTheStateDir proves the command
// hands the installer the state directory from the environment and the
// running user's uid: the table lands at the literal path under it, and
// the installer's line names that path.
func TestUnidecodeInstall_InstallsUnderTheStateDir(t *testing.T) {
	stateDir := t.TempDir()
	wheel, err := unidecode.WheelCachePath()
	if err != nil {
		t.Fatalf("locate wheel cache: %v", err)
	}

	var stdout bytes.Buffer
	environ := []string{"DUBPLATE_STATE_DIR=" + stateDir}
	if err := runUnidecodeInstall(context.Background(), environ, wheel, &stdout); err != nil {
		t.Fatalf("runUnidecodeInstall() error = %v, want nil", err)
	}

	path := filepath.Join(stateDir, "unidecode", "unidecode-1.4.0.tbl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read installed table: %v", err)
	}
	if _, err := unidecode.Parse(data); err != nil {
		t.Errorf("Parse(installed table): %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, path) {
		t.Errorf("stdout = %q, want it to name %s", got, path)
	}
}

// TestUnidecodeInstallCmd_NamesTheVerbOnce proves the verb's prefix
// is added at the command boundary, once, to a refusal from below it.
func TestUnidecodeInstallCmd_NamesTheVerbOnce(t *testing.T) {
	var stdout, stderr bytes.Buffer
	cmd := newUnidecodeInstallCmd(nil, &stdout, &stderr)
	cmd.SetArgs(nil)
	err := cmd.Execute()
	if err == nil || err.Error() != "unidecode install: DUBPLATE_STATE_DIR is not set" {
		t.Fatalf("error = %v, want %q", err, "unidecode install: DUBPLATE_STATE_DIR is not set")
	}
}

// TestDoctorCmd_UnidecodeRow pins the row's four states and the
// command's exit for each, against baseEnviron's otherwise-satisfied
// wall so the row's contribution is what moves the exit.
func TestDoctorCmd_UnidecodeRow(t *testing.T) {
	tests := []struct {
		name         string
		setStateDir  bool
		setup        func(t *testing.T, stateDir string)
		wantContains []string
		wantFail     bool
	}{
		{
			// The row itself reports no failure, but the wall's
			// DUBPLATE_STATE_DIR row is unset too, and that failure is
			// what moves the exit here.
			name:         "not checked when DUBPLATE_STATE_DIR is unset",
			setStateDir:  false,
			wantContains: []string{"not checked"},
			wantFail:     true,
		},
		{
			name:         "ok when a valid table is installed",
			setStateDir:  true,
			setup:        func(t *testing.T, stateDir string) { installTestUnidecodeTable(t, stateDir) },
			wantContains: []string{fmt.Sprintf("ok (%s)", unidecode.Version)},
		},
		{
			name:         "missing when no table was ever installed",
			setStateDir:  true,
			wantContains: []string{"missing: run dubplate-server unidecode install"},
			wantFail:     true,
		},
		{
			name:        "invalid when the table fails Parse",
			setStateDir: true,
			setup: func(t *testing.T, stateDir string) {
				leafPath := filepath.Join(stateDir, "unidecode")
				if err := os.MkdirAll(leafPath, 0o755); err != nil {
					t.Fatalf("mkdir leaf: %v", err)
				}
				tablePath := filepath.Join(leafPath, unidecode.TableFileName)
				if err := os.WriteFile(tablePath, []byte("not a table"), 0o644); err != nil {
					t.Fatalf("write invalid table: %v", err)
				}
			},
			wantContains: []string{"invalid: ", "run dubplate-server unidecode install"},
			wantFail:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			if tt.setup != nil {
				tt.setup(t, stateDir)
			}
			environ := dropKey(baseEnviron(), "DUBPLATE_STATE_DIR")
			if tt.setStateDir {
				environ = append(environ, "DUBPLATE_STATE_DIR="+stateDir)
			}

			var stdout, stderr bytes.Buffer
			cmd := newDoctorCmd(environ, freshRegistry(), &stdout, &stderr)
			err := cmd.Execute()

			if tt.wantFail && err == nil {
				t.Fatal("Execute() error = nil, want a refusal")
			}
			if !tt.wantFail && err != nil {
				t.Fatalf("Execute() error = %v, want nil", err)
			}

			wantPrefix := fmt.Sprintf("%-32s", unidecodeDoctorRow)
			var gotLine string
			for line := range strings.SplitSeq(stdout.String(), "\n") {
				if strings.HasPrefix(line, wantPrefix) {
					gotLine = line
				}
			}
			if gotLine == "" {
				t.Fatalf("stdout has no unidecode row: %q", stdout.String())
			}
			for _, want := range tt.wantContains {
				if !strings.Contains(gotLine, want) {
					t.Errorf("unidecode row = %q, want it to contain %q", gotLine, want)
				}
			}
		})
	}
}

// TestDoctorCmd_UnidecodeRowCreatesNothing proves doctor's row calls
// the load half alone: an empty state directory is still empty after a
// doctor run reports its table missing.
func TestDoctorCmd_UnidecodeRowCreatesNothing(t *testing.T) {
	stateDir := t.TempDir()
	before, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("ReadDir before: %v", err)
	}

	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd([]string{"DUBPLATE_STATE_DIR=" + stateDir}, config.NewRegistry(), &stdout, &stderr)
	_ = cmd.Execute()

	after, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("ReadDir after: %v", err)
	}
	if len(before) != 0 || len(after) != 0 {
		t.Fatalf("state directory listing before = %v, after = %v, want both empty", before, after)
	}
}
