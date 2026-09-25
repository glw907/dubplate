package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
)

// TestStoreRowStatus is the golden test for the store row's render: one
// exact expected line per report state, built directly rather than
// through a real store, since cmd/dubplate-server has no unexported
// runner to build an older or newer store with.
func TestStoreRowStatus(t *testing.T) {
	tests := []struct {
		name       string
		report     store.Report
		wantStatus string
		wantFailed bool
	}{
		{
			name:       "absent",
			report:     store.Report{State: store.StateAbsent},
			wantStatus: "absent",
			wantFailed: false,
		},
		{
			name:       "restore in progress",
			report:     store.Report{State: store.StateRestoreInProgress, Reason: "a restore is in progress (/state/store/restore-intent): finish it with restore --resume"},
			wantStatus: "restore in progress: finish it with restore --resume",
			wantFailed: true,
		},
		{
			name:       "busy naming holder pids",
			report:     store.Report{State: store.StateBusy, Holders: []string{"123", "456"}},
			wantStatus: "busy: held by pid 123, 456",
			wantFailed: true,
		},
		{
			name:       "busy with no holders named",
			report:     store.Report{State: store.StateBusy},
			wantStatus: "busy: held by another process",
			wantFailed: true,
		},
		{
			name:       "not a store",
			report:     store.Report{State: store.StateNotAStore, Reason: "/state/store/db/dubplate.db is not a database"},
			wantStatus: "not a store: /state/store/db/dubplate.db is not a database",
			wantFailed: true,
		},
		{
			name:       "refused",
			report:     store.Report{State: store.StateRefused, Reason: "leaf /state/store: owned by uid 0, not the caller's uid 1000"},
			wantStatus: "refused: leaf /state/store: owned by uid 0, not the caller's uid 1000",
			wantFailed: true,
		},
		{
			name:       "older names the pending migration files",
			report:     store.Report{State: store.StateOlder, StoreVersion: 1, BinaryVersion: 3, Pending: []string{"0002_x.sql", "0003_y.sql"}},
			wantStatus: "older (v1, binary v3): pending 0002_x.sql, 0003_y.sql",
			wantFailed: false,
		},
		{
			name:       "newer names both versions",
			report:     store.Report{State: store.StateNewer, StoreVersion: 4, BinaryVersion: 3},
			wantStatus: "newer (v4, binary v3)",
			wantFailed: true,
		},
		{
			name:       "current names the version",
			report:     store.Report{State: store.StateCurrent, StoreVersion: 3},
			wantStatus: "current (v3)",
			wantFailed: false,
		},
		{
			name:       "a FUSE warning appends to the status without changing failed",
			report:     store.Report{State: store.StateCurrent, StoreVersion: 3, Warning: "/state is on FUSE (type 0x65735546), which may span hosts"},
			wantStatus: "current (v3); /state is on FUSE (type 0x65735546), which may span hosts",
			wantFailed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, failed := storeRowStatus(tt.report)
			if status != tt.wantStatus {
				t.Errorf("storeRowStatus() status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("storeRowStatus() failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// TestStoreRow_EndToEnd is spec criterion 15's other two golden cases:
// the real inspect mode against a state directory with no leaf, and
// one whose db/ holds a zero-byte database.
func TestStoreRow_EndToEnd(t *testing.T) {
	t.Run("no leaf reports absent", func(t *testing.T) {
		environ := []string{"DUBPLATE_STATE_DIR=" + t.TempDir()}
		status, failed := storeRow(context.Background(), environ)
		if status != "absent" {
			t.Errorf("storeRow() status = %q, want %q", status, "absent")
		}
		if failed {
			t.Errorf("storeRow() failed = true, want false")
		}
	})

	t.Run("a zero-byte db/ file reports not a store", func(t *testing.T) {
		stateDir := t.TempDir()
		buildZeroByteStoreLeaf(t, stateDir)
		environ := []string{"DUBPLATE_STATE_DIR=" + stateDir}

		status, failed := storeRow(context.Background(), environ)
		if !strings.HasPrefix(status, "not a store: ") {
			t.Errorf("storeRow() status = %q, want prefix %q", status, "not a store: ")
		}
		if !failed {
			t.Errorf("storeRow() failed = false, want true")
		}
	})
}

// TestStoreRow_BusyTimeoutUnparseable proves storeRow's parseErr
// branch: an unparseable DUBPLATE_STORE_BUSY_TIMEOUT reports "not
// checked" and fails the row without ever opening the state directory,
// distinct from TestStoreBusyTimeout_ParseFailure's proof that
// storeBusyTimeout itself sets parseErr rather than rangeErr.
func TestStoreRow_BusyTimeoutUnparseable(t *testing.T) {
	environ := []string{
		"DUBPLATE_STATE_DIR=" + t.TempDir(),
		"DUBPLATE_STORE_BUSY_TIMEOUT=not-a-duration",
	}
	status, failed := storeRow(context.Background(), environ)
	if status != "not checked" {
		t.Errorf("storeRow() status = %q, want %q", status, "not checked")
	}
	if !failed {
		t.Errorf("storeRow() failed = false, want true")
	}
}

// TestStoreRow_StateDirUnset proves the "not checked" row a missing
// DUBPLATE_STATE_DIR reports, with no failure of its own (the owned
// wall row already reports the missing key).
func TestStoreRow_StateDirUnset(t *testing.T) {
	status, failed := storeRow(context.Background(), nil)
	if status != "not checked" {
		t.Errorf("storeRow() status = %q, want %q", status, "not checked")
	}
	if failed {
		t.Errorf("storeRow() failed = true, want false")
	}
}

// TestStoreBusyTimeout_Bounds proves the busy timeout's floor (1ms)
// and ceiling (2147483647ms), each a named unexported constant
// 0s, -1s, and 500us all fail the floor since a positive
// duration under one millisecond truncates to zero; 1000h
// (3600000000ms, the risk review's measured over-range value) fails
// the ceiling; 2147483647ms itself passes.
func TestStoreBusyTimeout_Bounds(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "zero", value: "0s", wantErr: true},
		{name: "negative", value: "-1s", wantErr: true},
		{name: "below one millisecond", value: "500us", wantErr: true},
		{name: "far above the ceiling", value: "1000h", wantErr: true},
		{name: "at the ceiling", value: "2147483647ms", wantErr: false},
		{name: "the default", value: "", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var environ []string
			if tt.value != "" {
				environ = []string{"DUBPLATE_STORE_BUSY_TIMEOUT=" + tt.value}
			}
			_, parseErr, rangeErr := storeBusyTimeout(environ)
			if parseErr != nil {
				t.Fatalf("storeBusyTimeout() parseErr = %v, want nil", parseErr)
			}
			if tt.wantErr {
				if rangeErr == nil {
					t.Fatal("storeBusyTimeout() rangeErr = nil, want a refusal")
				}
				if !strings.Contains(rangeErr.Error(), "DUBPLATE_STORE_BUSY_TIMEOUT") {
					t.Errorf("rangeErr = %q, want it to name DUBPLATE_STORE_BUSY_TIMEOUT", rangeErr.Error())
				}
				return
			}
			if rangeErr != nil {
				t.Errorf("storeBusyTimeout() rangeErr = %v, want nil", rangeErr)
			}
		})
	}
}

// TestStoreBusyTimeout_ParseFailure proves storeBusyTimeout reports a
// parse failure through parseErr, distinct from a range failure: a
// caller (storeRow, check.go) tells the two apart without matching
// error text.
func TestStoreBusyTimeout_ParseFailure(t *testing.T) {
	environ := []string{"DUBPLATE_STORE_BUSY_TIMEOUT=not-a-duration"}
	_, parseErr, rangeErr := storeBusyTimeout(environ)
	if parseErr == nil {
		t.Fatal("storeBusyTimeout() parseErr = nil, want a parse failure")
	}
	if rangeErr != nil {
		t.Errorf("storeBusyTimeout() rangeErr = %v, want nil alongside a parse failure", rangeErr)
	}
}

// buildZeroByteStoreLeaf creates a leaf under stateDir shaped as
// internal/engine/store's own leaf (store.lock, db/dubplate.db at
// zero bytes, snapshots/, staging/), so the inspect mode classifies it
// as an interrupted create rather than a real store.
// cmd/dubplate-server has no caller into store's own unexported leaf
// builder, so this mirrors its four names directly; a change to that
// shape is store's own package, not this one, to notice.
func buildZeroByteStoreLeaf(t *testing.T, stateDir string) {
	t.Helper()
	root := filepath.Join(stateDir, "store")
	for _, dir := range []string{root, filepath.Join(root, "db"), filepath.Join(root, "snapshots"), filepath.Join(root, "staging")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "store.lock"), nil, 0o600); err != nil {
		t.Fatalf("write store.lock: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "db", "dubplate.db"), nil, 0o600); err != nil {
		t.Fatalf("write dubplate.db: %v", err)
	}
}

// TestRegisterStoreConfig proves the threshold registers with its
// default: doctor's own threshold row derives from
// [config.Registry.Thresholds], not from storeBusyTimeoutThreshold
// directly, so a registration that never landed would leave the row
// silently absent rather than failing loudly.
func TestRegisterStoreConfig(t *testing.T) {
	reg := config.NewRegistry()
	registerStoreConfig(reg)

	found := false
	for _, th := range reg.Thresholds() {
		if th.Key == "DUBPLATE_STORE_BUSY_TIMEOUT" {
			found = true
			if th.Default != "15s" {
				t.Errorf("threshold default = %q, want %q", th.Default, "15s")
			}
		}
	}
	if !found {
		t.Fatal("registerStoreConfig did not register DUBPLATE_STORE_BUSY_TIMEOUT")
	}
}
