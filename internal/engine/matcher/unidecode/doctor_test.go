package unidecode

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRowStatus(t *testing.T) {
	const fix = "run dubplate-server unidecode install"

	// writeTable puts content at the table's path, leaf included.
	writeTable := func(t *testing.T, stateDir string, content []byte) string {
		t.Helper()
		leaf := filepath.Join(stateDir, "unidecode")
		if err := os.MkdirAll(leaf, 0o755); err != nil {
			t.Fatalf("fixture state: %v", err)
		}
		path := filepath.Join(leaf, TableFileName)
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatalf("fixture state: %v", err)
		}
		return path
	}

	tests := []struct {
		name       string
		stateDir   func(t *testing.T) string
		want       func(stateDir string) (prefix, suffix string)
		wantExact  string
		wantFailed bool
	}{
		{
			name:      "not checked when the state directory is unset",
			stateDir:  func(*testing.T) string { return "" },
			wantExact: "not checked",
		},
		{
			name: "ok when a valid table is installed",
			stateDir: func(t *testing.T) string {
				dir := t.TempDir()
				installTestTable(t, dir)
				return dir
			},
			wantExact: "ok (1.4.0)",
		},
		{
			name:       "missing when the leaf was never created",
			stateDir:   func(t *testing.T) string { return t.TempDir() },
			wantExact:  "missing: " + fix,
			wantFailed: true,
		},
		{
			name:       "missing when the state directory is absent",
			stateDir:   func(t *testing.T) string { return filepath.Join(t.TempDir(), "absent") },
			wantExact:  "missing: " + fix,
			wantFailed: true,
		},
		{
			name: "missing when the leaf holds no table",
			stateDir: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.Mkdir(filepath.Join(dir, "unidecode"), 0o755); err != nil {
					t.Fatalf("fixture state: %v", err)
				}
				return dir
			},
			wantExact:  "missing: " + fix,
			wantFailed: true,
		},
		{
			name: "invalid when the table fails to parse",
			stateDir: func(t *testing.T) string {
				dir := t.TempDir()
				writeTable(t, dir, []byte("not a table"))
				return dir
			},
			want: func(dir string) (string, string) {
				return "invalid: " + filepath.Join(dir, "unidecode", TableFileName) + ": ", ": " + fix
			},
			wantFailed: true,
		},
		{
			name: "invalid when a directory stands where the table belongs",
			stateDir: func(t *testing.T) string {
				dir := t.TempDir()
				if err := os.MkdirAll(filepath.Join(dir, "unidecode", TableFileName), 0o755); err != nil {
					t.Fatalf("fixture state: %v", err)
				}
				return dir
			},
			want: func(dir string) (string, string) {
				return "invalid: read " + filepath.Join(dir, "unidecode", TableFileName) + ": is a directory", ": " + fix
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := tt.stateDir(t)
			status, failed := RowStatus(stateDir)
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
			if tt.want == nil {
				if status != tt.wantExact {
					t.Errorf("status = %q, want %q", status, tt.wantExact)
				}
				return
			}
			prefix, suffix := tt.want(stateDir)
			if !strings.HasPrefix(status, prefix) || !strings.HasSuffix(status, suffix) {
				t.Errorf("status = %q, want it to begin %q and end %q", status, prefix, suffix)
			}
		})
	}
}

// TestRowStatus_CreatesNothing proves the row reads and never writes:
// an empty state directory is still empty after the row reports the
// table missing.
func TestRowStatus_CreatesNothing(t *testing.T) {
	stateDir := t.TempDir()
	if status, failed := RowStatus(stateDir); !failed {
		t.Fatalf("RowStatus = %q, %v, want a failure on an empty state directory", status, failed)
	}
	after, err := os.ReadDir(stateDir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(after) != 0 {
		t.Fatalf("state directory listing = %v, want empty", after)
	}
}
