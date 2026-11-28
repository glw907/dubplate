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
	own := os.Geteuid()

	tests := []struct {
		name string
		// setup shapes the auth dir under stateDir and returns the status
		// text the row must carry, and the euid it runs with.
		setup      func(t *testing.T, stateDir string) (want string, euid int)
		wantFailed bool
	}{
		{
			name: "an absent auth dir under a writable state directory",
			setup: func(*testing.T, string) (string, int) {
				return "absent; serve creates it", own
			},
		},
		{
			name: "an auth dirOpen built",
			setup: func(t *testing.T, stateDir string) (string, int) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				return "ok", own
			},
		},
		{
			name: "an auth dir at mode 0750",
			setup: func(t *testing.T, stateDir string) (string, int) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				authDir := filepath.Join(stateDir, "auth")
				chmodRestoring(t, authDir, 0o750)
				return authDir + " has mode 750, want at most 700", own
			},
			wantFailed: true,
		},
		{
			name: "an auth dir another uid owns",
			setup: func(t *testing.T, stateDir string) (string, int) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				authDir := filepath.Join(stateDir, "auth")
				return authDir + " is owned by uid " + strconv.Itoa(own) + ", want uid " + strconv.Itoa(own+1), own + 1
			},
			wantFailed: true,
		},
		{
			name: "a nonce subdirectory the process may not write",
			setup: func(t *testing.T, stateDir string) (string, int) {
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
			name: "an absent epoch subdirectory under an auth dir the process may not write",
			setup: func(t *testing.T, stateDir string) (string, int) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				authDir := filepath.Join(stateDir, "auth")
				if err := os.Remove(filepath.Join(authDir, "epoch")); err != nil {
					t.Fatalf("fixture state: remove epoch: %v", err)
				}
				chmodRestoring(t, authDir, 0o500)
				return filepath.Join(authDir, "epoch") + " is absent, and " + authDir + " is not writable: permission denied", own
			},
			wantFailed: true,
		},
		{
			name: "an absent auth dir under a state directory the process may not write",
			setup: func(t *testing.T, stateDir string) (string, int) {
				chmodRestoring(t, stateDir, 0o500)
				return "absent, and serve cannot create it: " + stateDir + " is not writable: permission denied", own
			},
			wantFailed: true,
		},
		{
			name: "an auth dir that is a file",
			setup: func(t *testing.T, stateDir string) (string, int) {
				authDir := filepath.Join(stateDir, "auth")
				if err := os.WriteFile(authDir, nil, 0o600); err != nil {
					t.Fatalf("fixture state: %v", err)
				}
				return authDir + " is not a directory", own
			},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			want, euid := tt.setup(t, stateDir)
			got, failed := dirStatus(&config.Config{StateDir: stateDir}, euid)
			if got != want || failed != tt.wantFailed {
				t.Errorf("dirStatus() = %q, %v, want %q, %v", got, failed, want, tt.wantFailed)
			}
		})
	}
}

func TestDirStatus_NilConfigIsNotChecked(t *testing.T) {
	got, failed := dirStatus(nil, os.Geteuid())
	if got != "not checked" || failed {
		t.Errorf("dirStatus(nil) = %q, %v, want %q, false", got, failed, "not checked")
	}
}
