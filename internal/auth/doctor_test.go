package auth

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// chmodRestoring sets path's mode and restores 0700 at cleanup, so
// t.TempDir's own removal can still descend into it.
func chmodRestoring(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("fixture state: Chmod %s: %v", path, err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o700) })
}

func TestDirStatus(t *testing.T) {
	euid := os.Geteuid()
	own := dirOptions{geteuid: os.Geteuid}

	tests := []struct {
		name string
		// setup shapes the leaf under stateDir and returns the status
		// text the row must carry, and the options it runs with.
		setup      func(t *testing.T, stateDir string) (want string, opts dirOptions)
		wantFailed bool
	}{
		{
			name: "an absent leaf under a writable state directory",
			setup: func(*testing.T, string) (string, dirOptions) {
				return "absent; serve creates it", own
			},
		},
		{
			name: "a leaf Open built",
			setup: func(t *testing.T, stateDir string) (string, dirOptions) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				return "ok", own
			},
		},
		{
			name: "a leaf at mode 0750",
			setup: func(t *testing.T, stateDir string) (string, dirOptions) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				leaf := filepath.Join(stateDir, "auth")
				chmodRestoring(t, leaf, 0o750)
				return "leaf " + leaf + " has mode 750, want at most 700", own
			},
			wantFailed: true,
		},
		{
			name: "a leaf another uid owns",
			setup: func(t *testing.T, stateDir string) (string, dirOptions) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				leaf := filepath.Join(stateDir, "auth")
				return "leaf " + leaf + " is owned by uid " + strconv.Itoa(euid) + ", want uid " + strconv.Itoa(euid+1),
					dirOptions{geteuid: func() int { return euid + 1 }}
			},
			wantFailed: true,
		},
		{
			name: "a nonce subdirectory the process may not write",
			setup: func(t *testing.T, stateDir string) (string, dirOptions) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				nonce := filepath.Join(stateDir, "auth", "nonce")
				chmodRestoring(t, nonce, 0o500)
				return nonce + " is not writable: permission denied", own
			},
			wantFailed: true,
		},
		{
			name: "an absent epoch subdirectory under a leaf the process may not write",
			setup: func(t *testing.T, stateDir string) (string, dirOptions) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				leaf := filepath.Join(stateDir, "auth")
				if err := os.Remove(filepath.Join(leaf, "epoch")); err != nil {
					t.Fatalf("fixture state: remove epoch: %v", err)
				}
				chmodRestoring(t, leaf, 0o500)
				return filepath.Join(leaf, "epoch") + " is absent, and " + leaf + " is not writable: permission denied", own
			},
			wantFailed: true,
		},
		{
			name: "an absent leaf under a state directory the process may not write",
			setup: func(t *testing.T, stateDir string) (string, dirOptions) {
				chmodRestoring(t, stateDir, 0o500)
				return "absent, and serve cannot create it: " + stateDir + " is not writable: permission denied", own
			},
			wantFailed: true,
		},
		{
			name: "a leaf that is a file",
			setup: func(t *testing.T, stateDir string) (string, dirOptions) {
				leaf := filepath.Join(stateDir, "auth")
				if err := os.WriteFile(leaf, nil, 0o600); err != nil {
					t.Fatalf("fixture state: %v", err)
				}
				return "leaf " + leaf + " is not a directory", own
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			want, opts := tt.setup(t, stateDir)
			got, failed := dirStatus(&config.Config{StateDir: stateDir}, opts)
			if got != want || failed != tt.wantFailed {
				t.Errorf("dirStatus() = %q, %v, want %q, %v", got, failed, want, tt.wantFailed)
			}
		})
	}
}

func TestDirStatus_NilConfigIsNotChecked(t *testing.T) {
	got, failed := dirStatus(nil, dirOptions{geteuid: os.Geteuid})
	if got != "not checked" || failed {
		t.Errorf("dirStatus(nil) = %q, %v, want %q, false", got, failed, "not checked")
	}
}
