package store

import (
	"context"
	"os"
	"os/exec"
	"slices"
	"strings"
	"syscall"
	"testing"
)

// childActionEnv and childStateDirEnv tell a re-executed test binary
// which write TestInspectChildWrite makes, and on which store.
const (
	childActionEnv   = "DUBPLATE_STORE_TEST_CHILD"
	childStateDirEnv = "DUBPLATE_STORE_TEST_CHILD_STATE_DIR"
)

// TestInspectChildWrite is not a test on its own: it does nothing
// unless a parent test re-executes the binary naming it. The child
// makes the write that decides the store's state with automatic
// checkpoints off, and exits without closing, so the write lives only
// in the -wal the way a crashed or still-running process leaves it.
func TestInspectChildWrite(t *testing.T) {
	action := os.Getenv(childActionEnv)
	if action == "" {
		return
	}
	l := newLeaf(os.Getenv(childStateDirEnv))
	h, err := openHandle(l.dbPath, openBusy)
	if err != nil {
		t.Fatalf("child open: %v", err)
	}
	ctx := context.Background()
	if _, err := h.write.ExecContext(ctx, "PRAGMA wal_autocheckpoint = 0"); err != nil {
		t.Fatalf("child autocheckpoint off: %v", err)
	}
	switch action {
	case "older":
		if err := migrate(ctx, h, l, migrationSourceThrough(t, 2), defaultOpts()); err != nil {
			t.Fatalf("child migrate: %v", err)
		}
	case "newer":
		if _, err := h.write.ExecContext(ctx, futureLedgerRow); err != nil {
			t.Fatalf("child newer: %v", err)
		}
	default:
		t.Fatalf("child action %q", action)
	}
	os.Exit(0)
}

// runChildWrite re-executes the test binary to make action's write on
// stateDir's store, leaving it in a hot -wal.
func runChildWrite(t *testing.T, stateDir, action string) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestInspectChildWrite$", "-test.count=1")
	cmd.Env = append(os.Environ(), childActionEnv+"="+action, childStateDirEnv+"="+stateDir)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("child %s: %v\n%s", action, err, out)
	}
	info, err := os.Stat(newLeaf(stateDir).dbPath + "-wal")
	if err != nil || info.Size() == 0 {
		t.Fatalf("child %s left no hot -wal: %v", action, err)
	}
}

// TestInspect_HotWAL is spec criterion 15's in-package half: inspect
// reports an absent, an older, and a newer store by name and leaves the
// database file's bytes and mode unchanged. The older and newer stores
// carry the write that decides their state only in a hot -wal, so the
// reported version proves inspect read through it, and a read that
// checkpoints on close would change the main file.
func TestInspect_HotWAL(t *testing.T) {
	tests := []struct {
		name        string
		through     int
		child       string
		wantState   State
		wantVersion int
		wantPending []string
	}{
		{name: "absent", wantState: StateAbsent},
		{name: "older", through: 1, child: "older", wantState: StateOlder, wantVersion: 2, wantPending: []string{"0003_catalog.sql"}},
		{name: "newer", through: binaryHead, child: "newer", wantState: StateNewer, wantVersion: 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			if tt.through == 0 {
				r := inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy}, os.Geteuid())
				if r.State != tt.wantState {
					t.Errorf("state = %v (%s), want %v", r.State, r.Reason, tt.wantState)
				}
				if left := listDir(t, stateDir); len(left) != 0 {
					t.Errorf("inspect created %q", left)
				}
				return
			}
			l := buildStore(t, stateDir, tt.through)
			runChildWrite(t, stateDir, tt.child)
			hash := hashFile(t, l.dbPath)
			before, err := os.Stat(l.dbPath)
			if err != nil {
				t.Fatal(err)
			}

			r := inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy}, os.Geteuid())
			if r.State != tt.wantState {
				t.Errorf("state = %v (%s), want %v", r.State, r.Reason, tt.wantState)
			}
			if r.StoreVersion != tt.wantVersion || r.BinaryVersion != binaryHead {
				t.Errorf("versions = store %d, binary %d; want %d, %d", r.StoreVersion, r.BinaryVersion, tt.wantVersion, binaryHead)
			}
			if !slices.Equal(r.Pending, tt.wantPending) {
				t.Errorf("pending = %q, want %q", r.Pending, tt.wantPending)
			}
			if hashFile(t, l.dbPath) != hash {
				t.Error("the database file's bytes changed")
			}
			after, err := os.Stat(l.dbPath)
			if err != nil {
				t.Fatal(err)
			}
			if after.Mode() != before.Mode() {
				t.Errorf("mode = %v, want %v unchanged", after.Mode(), before.Mode())
			}
		})
	}
}

// TestInspect_EveryOtherState produces each remaining report state once
// from a real fixture.
func TestInspect_EveryOtherState(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T, stateDir string)
		euid       int
		wantState  State
		wantReason string
	}{
		{
			name: "restore in progress",
			setup: func(t *testing.T, stateDir string) {
				l := buildStore(t, stateDir, binaryHead)
				if err := os.WriteFile(l.restoreIntentPath, nil, 0o600); err != nil {
					t.Fatal(err)
				}
			},
			wantState:  StateRestoreInProgress,
			wantReason: "restore --resume",
		},
		{
			name: "busy",
			setup: func(t *testing.T, stateDir string) {
				takeRawExclusive(t, buildStore(t, stateDir, binaryHead).lockPath)
			},
			wantState:  StateBusy,
			wantReason: "held by",
		},
		{
			name: "not a store",
			setup: func(t *testing.T, stateDir string) {
				l := newLeaf(stateDir)
				if err := l.create(stateDir, os.Geteuid()); err != nil {
					t.Fatal(err)
				}
				precreateFile(t, l.dbPath)
			},
			wantState:  StateNotAStore,
			wantReason: "start serve to finish creating it",
		},
		{
			name:       "refused, owner",
			setup:      func(t *testing.T, stateDir string) { buildStore(t, stateDir, binaryHead) },
			euid:       os.Geteuid() + 1,
			wantState:  StateRefused,
			wantReason: "runuser -u",
		},
		{
			name:      "current",
			setup:     func(t *testing.T, stateDir string) { buildStore(t, stateDir, binaryHead) },
			wantState: StateCurrent,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			tt.setup(t, stateDir)
			euid := tt.euid
			if euid == 0 {
				euid = os.Geteuid()
			}

			r := inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy}, euid)
			if r.State != tt.wantState {
				t.Errorf("state = %v (%s), want %v", r.State, r.Reason, tt.wantState)
			}
			if !strings.Contains(r.Reason, tt.wantReason) {
				t.Errorf("reason %q does not name %q", r.Reason, tt.wantReason)
			}
			if tt.wantState == StateCurrent && r.StoreVersion != binaryHead {
				t.Errorf("store version = %d, want %d", r.StoreVersion, binaryHead)
			}
		})
	}
}

// TestInspect_AbsentLeafSkipsOwnerCheck: a state directory with no
// leaf is absent whatever uid asks, since there is nothing yet to own.
// The euid differs from the state directory's owner, so an owner check
// on the absent branch would refuse.
func TestInspect_AbsentLeafSkipsOwnerCheck(t *testing.T) {
	stateDir := t.TempDir()
	r := inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy}, os.Geteuid()+1)
	if r.State != StateAbsent {
		t.Errorf("state = %v (%s), want absent", r.State, r.Reason)
	}
}

// TestInspect_FUSEWarns: a store on FUSE opens and reports its state,
// carrying a warning that names FUSE.
func TestInspect_FUSEWarns(t *testing.T) {
	stateDir := t.TempDir()
	buildStore(t, stateDir, binaryHead)
	orig := statfsFunc
	t.Cleanup(func() { statfsFunc = orig })
	statfsFunc = func(_ string, st *syscall.Statfs_t) error {
		st.Type = fuseSuperMagic
		return nil
	}

	r := inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy}, os.Geteuid())
	if r.State != StateCurrent {
		t.Errorf("state = %v (%s), want current", r.State, r.Reason)
	}
	if !strings.Contains(r.Warning, "FUSE") {
		t.Errorf("warning %q does not name FUSE", r.Warning)
	}
}

// TestStateString pins each state's printed name, which status lines
// carry.
func TestStateString(t *testing.T) {
	want := map[State]string{
		StateAbsent:            "absent",
		StateRestoreInProgress: "restore in progress",
		StateBusy:              "busy",
		StateNotAStore:         "not a store",
		StateRefused:           "refused",
		StateOlder:             "older",
		StateNewer:             "newer",
		StateCurrent:           "current",
		0:                      "unknown",
	}
	for s, name := range want {
		if got := s.String(); got != name {
			t.Errorf("State(%d).String() = %q, want %q", int(s), got, name)
		}
	}
}
