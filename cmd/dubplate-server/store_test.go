package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
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
			report:     store.Report{State: store.StateRestoreInProgress, Reason: "restore-intent record /state/store/restore-intent (state swap): finish the restore with restore --resume"},
			wantStatus: "restore in progress: restore-intent record /state/store/restore-intent (state swap): finish the restore with restore --resume",
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
			name:       "an unrecognized state is refused",
			report:     store.Report{State: store.State(99)},
			wantStatus: "refused: unrecognized store state 99",
			wantFailed: true,
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

// TestStoreRow_RestoreIntent is the store row's render of each
// restore-intent record state through the real inspect mode: every
// state fails the row, a reconcile-pending record names serve, every
// other state names restore --resume, and a record inspect cannot read
// names why. Each record is written by hand in the record's JSON form.
func TestStoreRow_RestoreIntent(t *testing.T) {
	record := func(state, library string) string {
		return `{"format": 1, "state": "` + state + `", "stamp": "20260925T120000.000000Z",
"source": {"remote": "remote:dubplate", "snapshot": "20260924T030000.000000Z"},
"incoming": "restore-20260925T120000.000000Z.db", "pre_restore": "pre-restore-20260925T120000.000000Z.db",
"discard_current": false, "library": ` + library + `, "version": 2}`
	}
	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "swap", body: record("swap", "null"), detail: "(state swap): finish the restore with restore --resume"},
		{name: "library", body: record("library", `{"root": "/srv/music", "into_nonempty": false}`), detail: "(state library): finish the restore with restore --resume"},
		{name: "reconcile", body: record("reconcile", "null"), detail: "(state reconcile): finish the restore with restore --resume"},
		{name: "reconcile-pending", body: record("reconcile-pending", "null"), detail: "(state reconcile-pending): start serve to finish the restore"},
		{name: "unreadable", body: "", detail: "is unreadable: the record is empty: finish the restore with restore --resume"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			buildZeroByteStoreLeaf(t, stateDir)
			path := filepath.Join(stateDir, "store", "restore-intent")
			if err := os.WriteFile(path, []byte(tt.body), 0o600); err != nil {
				t.Fatal(err)
			}

			status, failed, _, checked := storeRow(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir})
			want := "restore in progress: restore-intent record " + path + " " + tt.detail
			if status != want {
				t.Errorf("storeRow() status = %q, want %q", status, want)
			}
			if !failed || !checked {
				t.Errorf("storeRow() failed, checked = %v, %v, want true, true", failed, checked)
			}
		})
	}
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

// TestRegisterStoreConfig_Settings proves the two store-homed settings
// register with R-6c11's defaults.
func TestRegisterStoreConfig_Settings(t *testing.T) {
	reg := config.NewRegistry()
	registerStoreConfig(reg)

	want := map[string]string{
		"DUBPLATE_JOB_RETENTION":             "2160h",
		"DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP": "3",
	}
	got := make(map[string]string)
	for _, s := range reg.Settings() {
		got[s.Key] = s.Default
	}
	for key, def := range want {
		if got[key] != def {
			t.Errorf("registerStoreConfig registered %s with default %q, want %q", key, got[key], def)
		}
	}
}

// seedSettingRow writes key as a literal settings row of source
// ("env-seed" or "admin") with a matching seed value, and bumps the
// generation, in one transaction: production writes a settings row
// only through the seed apply or, for an admin row, rung 14's write
// path, so a test seeds one directly by literal SQL.
func seedSettingRow(t *testing.T, h *store.Handle, key, value, source string) {
	t.Helper()
	if source != "env-seed" && source != "admin" {
		t.Fatalf("seed setting %s: source %q, want env-seed or admin", key, source)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := h.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		q := fmt.Sprintf(`INSERT INTO settings (key, value, source, seed_value, actor, updated_at)
VALUES (?, ?, '%s', ?, 'fixture', 1)`, source)
		if _, err := tx.Exec(ctx, q, key, value, value); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `UPDATE settings_generation SET generation = generation + 1 WHERE singleton = 1`)
		return err
	})
	if err != nil {
		t.Fatalf("seed setting %s: %v", key, err)
	}
}

// TestSettingsReaders is spec criterion 13's reader half: each of the
// four keys' reader, built the same way runBackup builds it, returns
// the registry default while the store holds no row, and a stored
// env-seed row's non-default value once one exists. Each stored value
// sits at or above the floor under either answer to Ruling 3.
func TestSettingsReaders(t *testing.T) {
	dir := t.TempDir()
	seedBackupStore(t, dir)
	h, err := store.OpenWrite(context.Background(), dir, testBackupBusy)
	if err != nil {
		t.Fatalf("OpenWrite: %v", err)
	}
	defer func() { _ = h.Close() }()

	reg := config.NewRegistry()
	registerStoreConfig(reg)
	registerBackupConfig(reg)
	acc := settings.NewAccessor(h, reg)
	guardFraction, snapshotKeep, archiveRetention := backupSettingsReaders(acc)
	retention := jobRetentionReader(acc)

	tests := []struct {
		name        string
		key         string
		seed        string
		read        func() (string, error)
		wantDefault string
		wantSeeded  string
	}{
		{
			name: "job retention", key: "DUBPLATE_JOB_RETENTION", seed: "48h",
			read:        func() (string, error) { v, err := retention(context.Background()); return v.String(), err },
			wantDefault: "2160h0m0s", wantSeeded: "48h0m0s",
		},
		{
			name: "backup guard fraction", key: "DUBPLATE_BACKUP_GUARD_FRACTION", seed: "0.9",
			read:        func() (string, error) { v, err := guardFraction(context.Background()); return fmt.Sprint(v), err },
			wantDefault: "0.5", wantSeeded: "0.9",
		},
		{
			name: "backup snapshot keep", key: "DUBPLATE_BACKUP_SNAPSHOT_KEEP", seed: "3",
			read:        func() (string, error) { v, err := snapshotKeep(context.Background()); return fmt.Sprint(v), err },
			wantDefault: "14", wantSeeded: "3",
		},
		{
			name: "backup archive retention", key: "DUBPLATE_BACKUP_ARCHIVE_RETENTION", seed: "168h",
			read:        func() (string, error) { v, err := archiveRetention(context.Background()); return v.String(), err },
			wantDefault: "2160h0m0s", wantSeeded: "168h0m0s",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := tt.read(); err != nil || got != tt.wantDefault {
				t.Fatalf("%s with no row = %q, %v, want the default %q", tt.key, got, err, tt.wantDefault)
			}
			seedSettingRow(t, h, tt.key, tt.seed, "env-seed")
			if got, err := tt.read(); err != nil || got != tt.wantSeeded {
				t.Errorf("%s after an env-seed row = %q, %v, want %q", tt.key, got, err, tt.wantSeeded)
			}
		})
	}
}

// settingsDoctorRowsTestSetting and settingsDoctorRowsSensitive are the
// two settings TestSettingsDoctorRows registers: an ordinary integer
// (floor 3, ceiling 100) and a sensitive one, whose value prints with
// its mark.
var (
	settingsDoctorRowsTestSetting = config.Setting{
		Key: "DUBPLATE_TEST_SETTING", Kind: config.KindInt, Default: "10",
		Doc: "an integer", Apply: config.ApplyNextOperation, Floor: "3", Ceiling: "100",
	}
	settingsDoctorRowsSensitive = config.Setting{
		Key: "DUBPLATE_TEST_SECRET", Kind: config.KindInt, Default: "0",
		Doc: "an integer", Apply: config.ApplyImmediately, Floor: "0", Ceiling: "100", Sensitive: true,
	}
)

// settingsDoctorRowsRegistry declares the two settings above.
func settingsDoctorRowsRegistry() *config.Registry {
	reg := config.NewRegistry()
	reg.AddSetting(settingsDoctorRowsTestSetting)
	reg.AddSetting(settingsDoctorRowsSensitive)
	return reg
}

// TestSettingsDoctorRows is spec criterion 12's golden: one literal
// expected row per state in Task 5's doctor table, a store.Report built
// by hand ([TestStoreRowStatus]'s own precedent), since cmd/dubplate-server
// has no unexported runner to build an older or table-absent store
// with.
func TestSettingsDoctorRows(t *testing.T) {
	key, secretKey := settingsDoctorRowsTestSetting.Key, settingsDoctorRowsSensitive.Key
	current := func(facts store.Facts) store.Report {
		return store.Report{State: store.StateCurrent, Facts: &facts}
	}

	tests := []struct {
		name       string
		environ    []string
		report     store.Report
		checked    bool
		want       []backupFactRow
		wantFailed bool
	}{
		{
			name:    "no row, env absent: the default is in force",
			checked: true,
			report:  current(store.Facts{}),
			want: []backupFactRow{
				{key, `"10" (source default, next operation)`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "no row, env present: seeds at serve's next start",
			checked: true,
			environ: []string{key + "=20"},
			report:  current(store.Facts{}),
			want: []backupFactRow{
				{key, `"10" (source default, next operation); env "20" seeds at serve's next start`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "env-seed row, env equal: the seeded value is in force",
			checked: true,
			environ: []string{key + "=20"},
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "20", Source: "env-seed", SeedValue: "20"}}}),
			want: []backupFactRow{
				{key, `"20" (source env-seed, next operation)`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "env-seed row, env different: re-seeds at serve's next start",
			checked: true,
			environ: []string{key + "=30"},
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "20", Source: "env-seed", SeedValue: "20"}}}),
			want: []backupFactRow{
				{key, `"20" (source env-seed, next operation); env "30" re-seeds at serve's next start`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "env-seed row, env absent: returns to the default at serve's next start",
			checked: true,
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "20", Source: "env-seed", SeedValue: "20"}}}),
			want: []backupFactRow{
				{key, `"20" (source env-seed, next operation); returns to the default at serve's next start`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "admin row, env absent: the administrator's value is in force",
			checked: true,
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "40", Source: "admin", SeedValue: "20"}}}),
			want: []backupFactRow{
				{key, `"40" (source admin, next operation)`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "admin row, env equal: the administrator's value is in force",
			checked: true,
			environ: []string{key + "=40"},
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "40", Source: "admin", SeedValue: "20"}}}),
			want: []backupFactRow{
				{key, `"40" (source admin, next operation)`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "admin row, env different: refusing",
			checked: true,
			environ: []string{key + "=50"},
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "40", Source: "admin", SeedValue: "20"}}}),
			want: []backupFactRow{
				{key, `refused: DUBPLATE_TEST_SETTING: the environment seeds "50" but an administrator set "40"; unset DUBPLATE_TEST_SETTING, or set it to "40"`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
			wantFailed: true,
		},
		{
			name:    "out-of-bounds admin value stays in force: refusing",
			checked: true,
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "1", Source: "admin", SeedValue: "1"}}}),
			want: []backupFactRow{
				{key, `refused: stored admin value refused: DUBPLATE_TEST_SETTING: value "1" is below the floor 3`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
			wantFailed: true,
		},
		{
			name:    "out-of-bounds env-seed row healed by re-seed",
			checked: true,
			environ: []string{key + "=15"},
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "1", Source: "env-seed", SeedValue: "1"}}}),
			want: []backupFactRow{
				{key, `"1" (source env-seed, next operation); env "15" re-seeds at serve's next start`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "out-of-bounds env-seed row healed by removal",
			checked: true,
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "1", Source: "env-seed", SeedValue: "1"}}}),
			want: []backupFactRow{
				{key, `"1" (source env-seed, next operation); returns to the default at serve's next start`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
		},
		{
			name:    "a stored value that does not parse: refusing naming the reason",
			checked: true,
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: key, Value: "not-a-number", Source: "env-seed", SeedValue: "not-a-number"}}}),
			want: []backupFactRow{
				{key, `refused: stored env-seed value refused: DUBPLATE_TEST_SETTING: invalid value "not-a-number", want an integer`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
			wantFailed: true,
		},
		{
			name:    "an env value that does not parse: refusing naming the reason",
			checked: true,
			environ: []string{key + "=abc"},
			report:  current(store.Facts{}),
			want: []backupFactRow{
				{key, `refused: environment seed refused: DUBPLATE_TEST_SETTING: invalid value "abc", want an integer`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
			},
			wantFailed: true,
		},
		{
			name:    "an orphan row for an unregistered key",
			checked: true,
			report:  current(store.Facts{Settings: []store.SettingsRow{{Key: "DUBPLATE_TEST_RETIRED", Value: "1", Source: "admin", SeedValue: "1"}}}),
			want: []backupFactRow{
				{key, `"10" (source default, next operation)`},
				{secretKey, "\"0\" (source default, immediately, sensitive)"},
				{"DUBPLATE_TEST_RETIRED", "no longer registered; ignored"},
			},
		},
		{
			name:    "absent store: holds no settings yet",
			checked: true,
			report:  store.Report{State: store.StateAbsent},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store holds no settings yet`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:    "absent store, env present: seeds at serve's next start",
			checked: true,
			environ: []string{key + "=20"},
			report:  store.Report{State: store.StateAbsent},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store holds no settings yet; env "20" seeds at serve's next start`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:    "older store below the version that adds the settings table",
			checked: true,
			report:  store.Report{State: store.StateOlder, StoreVersion: 3, Facts: &store.Facts{SettingsAbsent: true}},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store holds no settings yet`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:    "older store below the version that adds the jobs table: nil facts",
			checked: true,
			report:  store.Report{State: store.StateOlder, StoreVersion: 1},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store holds no settings yet`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store holds no settings yet"},
			},
		},
		{
			name:    "busy: not checked",
			checked: true,
			report:  store.Report{State: store.StateBusy},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked (busy)`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked (busy)"},
			},
		},
		{
			name:    "busy, env present: not checked against the store",
			checked: true,
			environ: []string{key + "=20"},
			report:  store.Report{State: store.StateBusy},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked (busy); env "20" not checked against the store`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked (busy)"},
			},
		},
		{
			name:    "restore in progress: not checked",
			checked: true,
			report:  store.Report{State: store.StateRestoreInProgress},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked (restore in progress)`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked (restore in progress)"},
			},
		},
		{
			name:    "not a store: not checked",
			checked: true,
			report:  store.Report{State: store.StateNotAStore},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked (not a store)`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked (not a store)"},
			},
		},
		{
			name:    "refused: not checked",
			checked: true,
			report:  store.Report{State: store.StateRefused},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked (refused)`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked (refused)"},
			},
		},
		{
			name:    "newer: not checked",
			checked: true,
			report:  store.Report{State: store.StateNewer},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked (newer)`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked (newer)"},
			},
		},
		{
			name:    "an unrecognized state: not checked",
			checked: true,
			report:  store.Report{State: store.State(99), Facts: &store.Facts{}},
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked (unknown)`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked (unknown)"},
			},
		},
		{
			name:    "storeRow never ran Inspect: not checked",
			checked: false,
			want: []backupFactRow{
				{key, `"10" (source default, next operation); store not checked`},
				{secretKey, "\"0\" (source default, immediately, sensitive); store not checked"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, failed := settingsDoctorRows(tt.environ, settingsDoctorRowsRegistry(), tt.report, tt.checked)
			if !slices.EqualFunc(rows, tt.want, func(a, b backupFactRow) bool { return a == b }) {
				t.Errorf("rows = %+v, want %+v", rows, tt.want)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// TestStoreRowsAgreeOnVouch pins that the store row and the settings
// rows read one verdict per store state: the store row fails exactly
// the states whose settings rows read "not checked", an unrecognized
// state among them. Every report carries facts, so a state the
// settings rows wrongly vouched for would evaluate them instead.
func TestStoreRowsAgreeOnVouch(t *testing.T) {
	states := []store.State{
		store.StateAbsent, store.StateRestoreInProgress, store.StateBusy, store.StateNotAStore,
		store.StateRefused, store.StateOlder, store.StateNewer, store.StateCurrent, store.State(99),
	}
	for _, state := range states {
		t.Run(state.String(), func(t *testing.T) {
			report := store.Report{State: state, Facts: &store.Facts{}}
			_, storeFailed := storeRowStatus(report)
			rows, _ := settingsDoctorRows(nil, settingsDoctorRowsRegistry(), report, true)
			if len(rows) == 0 {
				t.Fatal("settingsDoctorRows rendered no rows")
			}
			for _, row := range rows {
				notChecked := strings.Contains(row.status, "store not checked")
				if notChecked != storeFailed {
					t.Errorf("state %d: store row failed = %v, but %s row = %q", state, storeFailed, row.key, row.status)
				}
			}
		})
	}
}

// TestSettingsDoctorRows_LocalKeep is criterion 13's local-keep half
// (D10): a stored env-seed value of "2" for
// DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP renders "2" under
// registerStoreConfig's own registry, matching the environment so the
// row carries no change note.
func TestSettingsDoctorRows_LocalKeep(t *testing.T) {
	reg := config.NewRegistry()
	registerStoreConfig(reg)
	key := storeLocalSnapshotKeepSetting.Key

	report := store.Report{State: store.StateCurrent, Facts: &store.Facts{
		Settings: []store.SettingsRow{{Key: key, Value: "2", Source: "env-seed", SeedValue: "2"}},
	}}
	rows, failed := settingsDoctorRows([]string{key + "=2"}, reg, report, true)
	if failed {
		t.Fatalf("failed = true, want false")
	}
	idx := slices.IndexFunc(rows, func(r backupFactRow) bool { return r.key == key })
	if idx < 0 {
		t.Fatalf("rows carry no entry for %s: %+v", key, rows)
	}
	if want := `"2" (source env-seed, next operation)`; rows[idx].status != want {
		t.Errorf("%s row = %q, want %q", key, rows[idx].status, want)
	}
}

// TestSettingsDoctorRows_CorrelatedRule is the doctor table's row 10
// over the real pair: a stored administrator's retention of 30m stays
// in force beside an expiry the environment raises to 45m, so the
// post-seed pair breaks the retention rule, and each key's own row
// refuses naming both, not only the joined evaluation error.
func TestSettingsDoctorRows_CorrelatedRule(t *testing.T) {
	const keyExpiry, keyRetention = "DUBPLATE_MAGIC_LINK_EXPIRY", "DUBPLATE_NONCE_RETENTION"
	report := store.Report{State: store.StateCurrent, Facts: &store.Facts{Settings: []store.SettingsRow{
		{Key: keyExpiry, Value: "15m", Source: "env-seed", SeedValue: "15m"},
		{Key: keyRetention, Value: "30m", Source: "admin", SeedValue: "24h"},
	}}}
	environ := []string{keyExpiry + "=45m"}

	rows, failed := settingsDoctorRows(environ, productionRegistry(), report, true)

	if !failed {
		t.Error("failed = false, want true: the correlated rule breaks")
	}
	for _, key := range []string{keyExpiry, keyRetention} {
		idx := slices.IndexFunc(rows, func(r backupFactRow) bool { return r.key == key })
		if idx < 0 {
			t.Fatalf("rows carry no entry for %s: %+v", key, rows)
		}
		status := rows[idx].status
		if !strings.Contains(status, "refused:") || !strings.Contains(status, keyExpiry) || !strings.Contains(status, keyRetention) {
			t.Errorf("%s: status = %q, want a refusal naming both %s and %s", key, status, keyExpiry, keyRetention)
		}
	}
}

// TestLocalSnapshotKeep is the keep decision's table: one subtest per
// state, over a store.Report built by hand ([TestStoreRowStatus]'s own
// precedent) and the real store registry. The first row's fixture
// leaves the keep key unset in the environment, since that is the
// state in which the outcome's post-seed value (the default, 3) and
// the stored row's value (1) differ; a row that set the key would pass
// by reading either.
func TestLocalSnapshotKeep(t *testing.T) {
	reg := config.NewRegistry()
	registerStoreConfig(reg)
	key := storeLocalSnapshotKeepSetting.Key
	retention := jobRetentionSetting.Key

	facts := func(rows ...store.SettingsRow) *store.Facts { return &store.Facts{Settings: rows} }
	row := func(key, value, source string) store.SettingsRow {
		return store.SettingsRow{Key: key, Value: value, Source: source, SeedValue: value}
	}
	current := func(f *store.Facts) store.Report { return store.Report{State: store.StateCurrent, Facts: f} }
	older := func(f *store.Facts) store.Report { return store.Report{State: store.StateOlder, Facts: f} }

	tests := []struct {
		name    string
		report  store.Report
		environ []string
		want    int
		wantErr []string
	}{
		{
			name:   "env-seed row of 1, key unset in the environment",
			report: older(facts(row(key, "1", "env-seed"))),
			want:   1,
		},
		{
			name:    "in-bounds row, the environment names another in-bounds value",
			report:  older(facts(row(key, "5", "env-seed"))),
			environ: []string{key + "=7"},
			want:    5,
		},
		{
			name:    "admin row outside its bounds",
			report:  older(facts(row(key, "500", "admin"))),
			wantErr: []string{key, "above the ceiling 100"},
		},
		{
			name:    "env-seed row outside its bounds, the environment re-seeds it in bounds",
			report:  older(facts(row(key, "500", "env-seed"))),
			environ: []string{key + "=4"},
			want:    4,
		},
		{
			name:   "env-seed row outside its bounds, the environment no longer sets it",
			report: older(facts(row(key, "500", "env-seed"))),
			want:   3,
		},
		{
			name:   "no keep row",
			report: older(facts(row(retention, "48h", "env-seed"))),
			want:   3,
		},
		{
			name:   "no facts: absent",
			report: store.Report{State: store.StateAbsent},
			want:   3,
		},
		{
			name:   "no facts: busy",
			report: store.Report{State: store.StateBusy},
			want:   3,
		},
		{
			name:   "no facts: restore in progress",
			report: store.Report{State: store.StateRestoreInProgress},
			want:   3,
		},
		{
			name:   "no facts: refused",
			report: store.Report{State: store.StateRefused},
			want:   3,
		},
		{
			name:   "no settings table",
			report: older(&store.Facts{SettingsAbsent: true}),
			want:   3,
		},
		{
			name:    "older store, another setting's stored row refuses",
			report:  older(facts(row(retention, "1h", "admin"))),
			wantErr: []string{retention, "below the floor 24h"},
		},
		{
			name:   "current store, another setting's stored row refuses: the seed decides",
			report: current(facts(row(retention, "1h", "admin"))),
			want:   3,
		},
		{
			name:    "older store, the environment's keep value does not parse",
			report:  older(facts(row(retention, "48h", "env-seed"))),
			environ: []string{key + "=many"},
			wantErr: []string{key, "invalid value"},
		},
		{
			name:    "current store, the environment's keep value does not parse: the seed decides",
			report:  current(facts(row(retention, "48h", "env-seed"))),
			environ: []string{key + "=many"},
			want:    3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := localSnapshotKeep(tt.report, tt.environ, reg)
			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatalf("localSnapshotKeep() = %d, nil, want an error naming %q", got, tt.wantErr)
				}
				for _, want := range tt.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("error = %q, want it to contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("localSnapshotKeep() error = %v, want nil", err)
			}
			if got != tt.want {
				t.Errorf("localSnapshotKeep() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestStoreMigrationWait pins the threshold's default and bounds, each
// bound exactly at and one step past.
func TestStoreMigrationWait(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    time.Duration
		wantErr string
	}{
		{name: "default", want: 5 * time.Minute},
		{name: "floor", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=1s"}, want: time.Second},
		{name: "below the floor", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=999ms"}, wantErr: "DUBPLATE_STORE_MIGRATION_WAIT must be at least 1s"},
		{name: "ceiling", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=24h"}, want: 24 * time.Hour},
		{name: "above the ceiling", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=24h1s"}, wantErr: "DUBPLATE_STORE_MIGRATION_WAIT must be at most 24h"},
		{name: "unparseable", environ: []string{"DUBPLATE_STORE_MIGRATION_WAIT=soon"}, wantErr: "DUBPLATE_STORE_MIGRATION_WAIT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := storeMigrationWait(tt.environ)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("storeMigrationWait() error = %v, want it to contain %q", err, tt.wantErr)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("storeMigrationWait() = %v, %v, want %v, nil", got, err, tt.want)
			}
		})
	}
}
