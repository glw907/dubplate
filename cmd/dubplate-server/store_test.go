package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
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
		status, failed, report, checked := storeRow(context.Background(), environ)
		if status != "absent" {
			t.Errorf("storeRow() status = %q, want %q", status, "absent")
		}
		if failed {
			t.Errorf("storeRow() failed = true, want false")
		}
		if !checked {
			t.Error("storeRow() checked = false, want true")
		}
		if report.Facts != nil {
			t.Errorf("storeRow() report.Facts = %+v, want nil for an absent store", report.Facts)
		}
	})

	t.Run("a zero-byte db/ file reports not a store", func(t *testing.T) {
		stateDir := t.TempDir()
		buildZeroByteStoreLeaf(t, stateDir)
		environ := []string{"DUBPLATE_STATE_DIR=" + stateDir}

		status, failed, _, checked := storeRow(context.Background(), environ)
		if !strings.HasPrefix(status, "not a store: ") {
			t.Errorf("storeRow() status = %q, want prefix %q", status, "not a store: ")
		}
		if !failed {
			t.Errorf("storeRow() failed = false, want true")
		}
		if !checked {
			t.Error("storeRow() checked = false, want true")
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
	status, failed, _, checked := storeRow(context.Background(), environ)
	if status != "not checked" {
		t.Errorf("storeRow() status = %q, want %q", status, "not checked")
	}
	if !failed {
		t.Errorf("storeRow() failed = false, want true")
	}
	if checked {
		t.Error("storeRow() checked = true, want false: an unparseable busy timeout never runs Inspect")
	}
}

// TestStoreRow_StateDirUnset proves the "not checked" row a missing
// DUBPLATE_STATE_DIR reports, with no failure of its own (the owned
// wall row already reports the missing key).
func TestStoreRow_StateDirUnset(t *testing.T) {
	status, failed, _, checked := storeRow(context.Background(), nil)
	if status != "not checked" {
		t.Errorf("storeRow() status = %q, want %q", status, "not checked")
	}
	if failed {
		t.Errorf("storeRow() failed = true, want false")
	}
	if checked {
		t.Error("storeRow() checked = true, want false: a missing state directory never runs Inspect")
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

// testNow anchors every age computation in this file's goldens, so a
// wall-clock read can never make a test flaky.
var testNow = time.Date(2026, 9, 24, 12, 0, 0, 0, time.UTC)

// TestBackupFactsRows is the golden test for R-6bi13's four backup
// rows, each built from a constructed [store.Report] rather than a
// real store: the states table's "not checked" row (checked false, or
// a report whose Facts is nil), WAL size absent and present, free
// space known and unknown, local copies none and several, and last
// success never, an age, and an age with a prune's skipped or failed
// text appended.
func TestBackupFactsRows(t *testing.T) {
	tests := []struct {
		name    string
		now     time.Time
		report  store.Report
		checked bool
		want    []backupFactRow
	}{
		{
			name:    "not checked at all: storeRow never ran Inspect",
			now:     testNow,
			report:  store.Report{},
			checked: false,
			want: []backupFactRow{
				{backupLastSuccessRow, "not checked"},
				{backupWALSizeRow, "not checked"},
				{backupFreeSpaceRow, "not checked"},
				{backupLocalCopiesRow, "not checked"},
			},
		},
		{
			name:    "checked, but the store carries no Facts (absent, busy, older, newer, not a store, refused, or mid-restore)",
			now:     testNow,
			report:  store.Report{State: store.StateAbsent},
			checked: true,
			want: []backupFactRow{
				{backupLastSuccessRow, "not checked"},
				{backupWALSizeRow, "not checked"},
				{backupFreeSpaceRow, "not checked"},
				{backupLocalCopiesRow, "not checked"},
			},
		},
		{
			name:    "no succeeded backup: never, no wal, free known, no copies",
			now:     testNow,
			report:  store.Report{State: store.StateCurrent, Facts: &store.Facts{FreeBytes: 2048, FreeKnown: true}},
			checked: true,
			want: []backupFactRow{
				{backupLastSuccessRow, "never"},
				{backupWALSizeRow, "0 B"},
				{backupFreeSpaceRow, "2.0 KiB"},
				{backupLocalCopiesRow, "0 copies"},
			},
		},
		{
			name: "a clean succeeded backup: age alone, wal present, free unknown, several copies",
			now:  testNow,
			report: store.Report{State: store.StateCurrent, Facts: &store.Facts{
				LastSuccess: map[string]store.JobSuccess{
					ops.KindBackup: {Finished: testNow.Add(-95 * time.Minute), Progress: backup.FinishedMessage},
				},
				WALSize:    4096,
				FreeKnown:  false,
				Copies:     3,
				CopiesSize: 3*1024*1024 + 512*1024,
			}},
			checked: true,
			want: []backupFactRow{
				{backupLastSuccessRow, "1h35m ago"},
				{backupWALSizeRow, "4.0 KiB"},
				{backupFreeSpaceRow, "unknown"},
				{backupLocalCopiesRow, "3 copies, 3.5 MiB"},
			},
		},
		{
			name: "a succeeded backup whose prune was skipped names the reason",
			now:  testNow,
			report: store.Report{State: store.StateCurrent, Facts: &store.Facts{
				LastSuccess: map[string]store.JobSuccess{
					ops.KindBackup: {Finished: testNow.Add(-30 * time.Second), Progress: "prune skipped: the run's stamp is not the newest on the remote"},
				},
			}},
			checked: true,
			want: []backupFactRow{
				{backupLastSuccessRow, "1m ago: prune skipped: the run's stamp is not the newest on the remote"},
				{backupWALSizeRow, "0 B"},
				{backupFreeSpaceRow, "unknown"},
				{backupLocalCopiesRow, "0 copies"},
			},
		},
		{
			name: "a succeeded backup whose prune failed names the reason",
			now:  testNow,
			report: store.Report{State: store.StateCurrent, Facts: &store.Facts{
				LastSuccess: map[string]store.JobSuccess{
					ops.KindBackup: {Finished: testNow.Add(-25 * time.Hour), Progress: "prune failed: list remote snapshots: exit status 1"},
				},
			}},
			checked: true,
			want: []backupFactRow{
				{backupLastSuccessRow, "1d1h ago: prune failed: list remote snapshots: exit status 1"},
				{backupWALSizeRow, "0 B"},
				{backupFreeSpaceRow, "unknown"},
				{backupLocalCopiesRow, "0 copies"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := backupFactsRows(tt.now, tt.report, tt.checked)
			if len(got) != len(tt.want) {
				t.Fatalf("backupFactsRows() = %d rows, want %d: %+v", len(got), len(tt.want), got)
			}
			for i, row := range got {
				if row != tt.want[i] {
					t.Errorf("backupFactsRows()[%d] = %+v, want %+v", i, row, tt.want[i])
				}
			}
		})
	}
}

// TestFormatBytes proves the binary-unit scale directly, so
// TestBackupFactsRows's goldens are not the only proof a mutated
// divisor or exponent step would fail.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		n    int64
		want string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
		{1024 * 1024 * 1024, "1.0 GiB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.n); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

// TestFormatAge proves the minute-rounded age rendering across every
// scale it names: minutes alone, hours and minutes, and days and
// hours, plus a future finish time clamped to "0m" rather than a
// negative duration.
func TestFormatAge(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{0, "0m"},
		{-time.Minute, "0m"},
		{30 * time.Second, "1m"},
		{23 * time.Minute, "23m"},
		{95 * time.Minute, "1h35m"},
		{25 * time.Hour, "1d1h"},
	}
	for _, tt := range tests {
		if got := formatAge(tt.d); got != tt.want {
			t.Errorf("formatAge(%s) = %q, want %q", tt.d, got, tt.want)
		}
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
