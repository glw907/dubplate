package auth

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/store"
)

// facts builds the report facts the two roster rows read.
func facts(members, admins int) *store.Facts {
	return &store.Facts{Members: members, Admins: admins}
}

func TestMembersRowStatus(t *testing.T) {
	tests := []struct {
		name       string
		report     store.Report
		checked    bool
		want       string
		wantFailed bool
	}{
		{name: "the store was not inspected", want: "not read"},
		{
			name:    "no store yet",
			report:  store.Report{State: store.StateAbsent},
			checked: true,
			want:    "no store yet; start serve, then member add --admin",
		},
		{
			name:    "a state the store row refuses",
			report:  store.Report{State: store.StateBusy},
			checked: true,
			want:    "not read",
		},
		{
			name:    "a store older than the users table",
			report:  store.Report{State: store.StateOlder, StoreVersion: 2},
			checked: true,
			want:    "members not recorded before store version 3",
		},
		{
			name:    "a current store that carries no facts",
			report:  store.Report{State: store.StateCurrent, StoreVersion: 3},
			checked: true,
			want:    "not read",
		},
		{
			name:       "no members",
			report:     store.Report{State: store.StateCurrent, StoreVersion: 3, Facts: facts(0, 0)},
			checked:    true,
			want:       "no members; run dubplate-server member add --admin",
			wantFailed: true,
		},
		{
			name:       "members and no admin",
			report:     store.Report{State: store.StateCurrent, StoreVersion: 3, Facts: facts(1, 0)},
			checked:    true,
			want:       "members: 1, admins: 0; run dubplate-server member add --admin",
			wantFailed: true,
		},
		{
			name:    "members and an admin",
			report:  store.Report{State: store.StateCurrent, StoreVersion: 3, Facts: facts(2, 1)},
			checked: true,
			want:    "members: 2, admins: 1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, failed := MembersRowStatus(tt.report, tt.checked)
			if got != tt.want || failed != tt.wantFailed {
				t.Errorf("MembersRowStatus() = %q, %v, want %q, %v", got, failed, tt.want, tt.wantFailed)
			}
		})
	}
}

func TestAdminAddressRowStatus(t *testing.T) {
	const (
		notMember = "warning: this address is not a live member; run dubplate-server member add <address> --admin, or set DUBPLATE_ADMIN_EMAIL to a live admin"
		notAdmin  = "warning: this address is a live member without the admin flag; set DUBPLATE_ADMIN_EMAIL to a live admin"
	)
	current := func(standing store.Standing, f *store.Facts) store.Report {
		return store.Report{State: store.StateCurrent, StoreVersion: 3, Facts: f, Standing: standing}
	}
	tests := []struct {
		name    string
		wall    string
		report  store.Report
		checked bool
		want    string
	}{
		{name: "a wall failure stands", wall: "missing", report: current(store.StandingNone, facts(1, 1)), checked: true, want: "missing"},
		{name: "the store was not inspected", wall: "ok", report: current(store.StandingNone, facts(1, 1)), want: "ok"},
		{name: "a state the store row refuses", wall: "ok", report: store.Report{State: store.StateBusy, Standing: store.StandingNone, Facts: facts(1, 1)}, checked: true, want: "ok"},
		{name: "no facts", wall: "ok", report: current(store.StandingNone, nil), checked: true, want: "ok"},
		{name: "a store older than the users table", wall: "ok", report: store.Report{State: store.StateOlder, StoreVersion: 2, Facts: facts(1, 1), Standing: store.StandingNone}, checked: true, want: "ok"},
		{name: "no live member, which the members row fails", wall: "ok", report: current(store.StandingNone, facts(0, 0)), checked: true, want: "ok"},
		{name: "an address no member holds", wall: "ok", report: current(store.StandingNone, facts(1, 1)), checked: true, want: notMember},
		{name: "a removed member", wall: "ok", report: current(store.StandingRemoved, facts(1, 1)), checked: true, want: notMember},
		{name: "a live member without the admin flag", wall: "ok", report: current(store.StandingMember, facts(2, 1)), checked: true, want: notAdmin},
		{name: "a live admin", wall: "ok", report: current(store.StandingAdmin, facts(2, 1)), checked: true, want: "ok"},
		{name: "a standing the inspection did not read", wall: "ok", report: current(store.StandingNotRead, facts(2, 1)), checked: true, want: "ok"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AdminAddressRowStatus(tt.wall, tt.report, tt.checked); got != tt.want {
				t.Errorf("AdminAddressRowStatus() = %q, want %q", got, tt.want)
			}
		})
	}
}

// chmodRestoring sets path's mode and restores 0700 at cleanup, so
// t.TempDir's own removal can still descend into it.
func chmodRestoring(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.Chmod(path, mode); err != nil {
		t.Fatalf("fixture state: Chmod %s: %v", path, err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o700) })
}

func TestDirRowStatus(t *testing.T) {
	euid := os.Geteuid()
	own := DirRowOptions{Geteuid: os.Geteuid}

	tests := []struct {
		name string
		// setup shapes the leaf under stateDir and returns the status
		// text the row must carry, and the options it runs with.
		setup      func(t *testing.T, stateDir string) (want string, opts DirRowOptions)
		wantFailed bool
	}{
		{
			name: "an absent leaf under a writable state directory",
			setup: func(*testing.T, string) (string, DirRowOptions) {
				return "absent; serve creates it", own
			},
		},
		{
			name: "a leaf Open built",
			setup: func(t *testing.T, stateDir string) (string, DirRowOptions) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				return "ok", own
			},
		},
		{
			name: "a leaf at mode 0750",
			setup: func(t *testing.T, stateDir string) (string, DirRowOptions) {
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
			setup: func(t *testing.T, stateDir string) (string, DirRowOptions) {
				if _, err := Open(stateDir); err != nil {
					t.Fatalf("fixture state: Open: %v", err)
				}
				leaf := filepath.Join(stateDir, "auth")
				return "leaf " + leaf + " is owned by uid " + strconv.Itoa(euid) + ", want uid " + strconv.Itoa(euid+1),
					DirRowOptions{Geteuid: func() int { return euid + 1 }}
			},
			wantFailed: true,
		},
		{
			name: "a nonce subdirectory the process may not write",
			setup: func(t *testing.T, stateDir string) (string, DirRowOptions) {
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
			setup: func(t *testing.T, stateDir string) (string, DirRowOptions) {
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
			setup: func(t *testing.T, stateDir string) (string, DirRowOptions) {
				chmodRestoring(t, stateDir, 0o500)
				return "absent, and serve cannot create it: " + stateDir + " is not writable: permission denied", own
			},
			wantFailed: true,
		},
		{
			name: "a leaf that is a file",
			setup: func(t *testing.T, stateDir string) (string, DirRowOptions) {
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
			got, failed := DirRowStatus(&config.Config{StateDir: stateDir}, opts)
			if got != want || failed != tt.wantFailed {
				t.Errorf("DirRowStatus() = %q, %v, want %q, %v", got, failed, want, tt.wantFailed)
			}
		})
	}
}

func TestDirRowStatus_NilConfigIsNotChecked(t *testing.T) {
	got, failed := DirRowStatus(nil, DirRowOptions{Geteuid: os.Geteuid})
	if got != "not checked" || failed {
		t.Errorf("DirRowStatus(nil) = %q, %v, want %q, false", got, failed, "not checked")
	}
}
