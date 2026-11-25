package store

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestStoreRowStatus is the golden test for the store row's render: one
// exact expected line per report state, built directly rather than
// through a real store, which cannot be made to report an older or
// newer one here.
func TestReportRow(t *testing.T) {
	tests := []struct {
		name       string
		report     Report
		wantStatus string
		wantFailed bool
	}{
		{
			name:       "absent",
			report:     Report{State: StateAbsent},
			wantStatus: "absent",
			wantFailed: false,
		},
		{
			name:       "restore in progress",
			report:     Report{State: StateRestoreInProgress, Reason: "restore-intent record /state/store/restore-intent (state swap): finish the restore with restore --resume"},
			wantStatus: "restore in progress: restore-intent record /state/store/restore-intent (state swap): finish the restore with restore --resume",
			wantFailed: true,
		},
		{
			name:       "busy naming holder pids",
			report:     Report{State: StateBusy, Holders: []string{"123", "456"}},
			wantStatus: "busy: held by pid 123, 456",
			wantFailed: true,
		},
		{
			name:       "busy with no holders named",
			report:     Report{State: StateBusy},
			wantStatus: "busy: held by another process",
			wantFailed: true,
		},
		{
			name:       "not a store",
			report:     Report{State: StateNotAStore, Reason: "/state/store/db/dubplate.db is not a database"},
			wantStatus: "not a store: /state/store/db/dubplate.db is not a database",
			wantFailed: true,
		},
		{
			name:       "refused",
			report:     Report{State: StateRefused, Reason: "leaf /state/store: owned by uid 0, not the caller's uid 1000"},
			wantStatus: "refused: leaf /state/store: owned by uid 0, not the caller's uid 1000",
			wantFailed: true,
		},
		{
			name:       "older names the pending migration files",
			report:     Report{State: StateOlder, StoreVersion: 1, BinaryVersion: 3, Pending: []string{"0002_x.sql", "0003_y.sql"}},
			wantStatus: "older (v1, binary v3): pending 0002_x.sql, 0003_y.sql",
			wantFailed: false,
		},
		{
			name:       "newer names both versions",
			report:     Report{State: StateNewer, StoreVersion: 4, BinaryVersion: 3},
			wantStatus: "newer (v4, binary v3)",
			wantFailed: true,
		},
		{
			name:       "current names the version",
			report:     Report{State: StateCurrent, StoreVersion: 3},
			wantStatus: "current (v3)",
			wantFailed: false,
		},
		{
			name:       "an unrecognized state is refused",
			report:     Report{State: State(99)},
			wantStatus: "refused: unrecognized store state 99",
			wantFailed: true,
		},
		{
			name:       "a FUSE warning appends to the status without changing failed",
			report:     Report{State: StateCurrent, StoreVersion: 3, Warning: "/state is on FUSE (type 0x65735546), which may span hosts"},
			wantStatus: "current (v3); /state is on FUSE (type 0x65735546), which may span hosts",
			wantFailed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			row := tt.report.Row()
			if row.Name != "store" {
				t.Errorf("Row() name = %q, want %q", row.Name, "store")
			}
			if row.Status != tt.wantStatus {
				t.Errorf("Row() status = %q, want %q", row.Status, tt.wantStatus)
			}
			if row.Failed != tt.wantFailed {
				t.Errorf("Row() failed = %v, want %v", row.Failed, tt.wantFailed)
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
		row, report := Doctor(context.Background(), environ)
		status, failed := row.Status, row.Failed
		if status != "absent" {
			t.Errorf("Doctor() status = %q, want %q", status, "absent")
		}
		if failed {
			t.Errorf("Doctor() failed = true, want false")
		}
		if report.State != StateAbsent {
			t.Errorf("Doctor() report.State = %v, want %v", report.State, StateAbsent)
		}
		if report.Facts != nil {
			t.Errorf("Doctor() report.Facts = %+v, want nil for an absent store", report.Facts)
		}
	})

	t.Run("a zero-byte db/ file reports not a store", func(t *testing.T) {
		stateDir := t.TempDir()
		buildZeroByteStoreLeaf(t, stateDir)
		environ := []string{"DUBPLATE_STATE_DIR=" + stateDir}

		row, report := Doctor(context.Background(), environ)
		status, failed := row.Status, row.Failed
		if !strings.HasPrefix(status, "not a store: ") {
			t.Errorf("Doctor() status = %q, want prefix %q", status, "not a store: ")
		}
		if !failed {
			t.Errorf("Doctor() failed = false, want true")
		}
		if report.State != StateNotAStore {
			t.Errorf("Doctor() report.State = %v, want %v", report.State, StateNotAStore)
		}
	})
}

// TestStoreRow_RestoreIntent is the store row's render of each
// restore-intent record state through the real inspect mode: every
// state fails the row, a reconcile_pending record names serve, every
// other state names restore --resume, and a record inspect cannot read
// names why. Each record is written by hand in the record's JSON form.
func TestStoreRow_RestoreIntent(t *testing.T) {
	record := func(state, library string) string {
		return `{"format": 1, "state": "` + state + `", "stamp": "20260925T120000.000000Z",
"remote": "remote:dubplate", "snapshot": "20260924T030000.000000Z",
"incoming": "restore-20260925T120000.000000Z.db", "pre_restore": "pre-restore-20260925T120000.000000Z.db",
"discard_current": false, ` + library + `"version": 2}`
	}
	tests := []struct {
		name   string
		body   string
		detail string
	}{
		{name: "swap", body: record("swap", ""), detail: "(state swap): finish the restore with restore --resume"},
		{name: "library", body: record("library", `"library_root": "/srv/music", "into_nonempty": false, `), detail: "(state library): finish the restore with restore --resume"},
		{name: "reconcile", body: record("reconcile", ""), detail: "(state reconcile): finish the restore with restore --resume"},
		{name: "reconcile_pending", body: record("reconcile_pending", ""), detail: "(state reconcile_pending): start serve to finish the restore"},
		{name: "retired reconcile-pending", body: record("reconcile-pending", ""), detail: "(state reconcile_pending): start serve to finish the restore"},
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

			row, report := Doctor(context.Background(), []string{"DUBPLATE_STATE_DIR=" + stateDir})
			status, failed := row.Status, row.Failed
			want := "restore in progress: restore-intent record " + path + " " + tt.detail
			if status != want {
				t.Errorf("Doctor() status = %q, want %q", status, want)
			}
			if !failed || report.State != StateRestoreInProgress {
				t.Errorf("Doctor() failed, report.State = %v, %v, want true, %v", failed, report.State, StateRestoreInProgress)
			}
		})
	}
}

// TestStoreRow_BusyTimeoutUnparseable proves Doctor's parse-failure
// branch: an unparseable DUBPLATE_STORE_BUSY_TIMEOUT reports "not
// checked" and fails the row without ever opening the state directory,
// distinct from TestStoreBusyTimeout_ParseFailure's proof that
// the parse failure does not wrap store.ErrOutOfRange.
func TestStoreRow_BusyTimeoutUnparseable(t *testing.T) {
	var ran readerRan
	environ := []string{
		"DUBPLATE_STATE_DIR=" + t.TempDir(),
		"DUBPLATE_STORE_BUSY_TIMEOUT=not-a-duration",
	}
	row, report := Doctor(context.Background(), environ, ran.reader)
	status, failed := row.Status, row.Failed
	if status != "not checked" {
		t.Errorf("Doctor() status = %q, want %q", status, "not checked")
	}
	if !failed {
		t.Errorf("Doctor() failed = false, want true")
	}
	ran.requireNotInspected(t, report, "an unparseable busy timeout")
}

// TestStoreRow_BusyTimeoutOutOfRange proves the range refusal reads
// differently from a parse failure: the row names the key and the
// bound, fails, and never opens the state directory.
func TestStoreRow_BusyTimeoutOutOfRange(t *testing.T) {
	var ran readerRan
	environ := []string{
		"DUBPLATE_STATE_DIR=" + t.TempDir(),
		"DUBPLATE_STORE_BUSY_TIMEOUT=0s",
	}
	row, report := Doctor(context.Background(), environ, ran.reader)
	status, failed := row.Status, row.Failed
	if want := "not checked: DUBPLATE_STORE_BUSY_TIMEOUT must be at least 1ms"; status != want {
		t.Errorf("Doctor() status = %q, want %q", status, want)
	}
	if !failed {
		t.Error("Doctor() failed = false, want true")
	}
	ran.requireNotInspected(t, report, "an out-of-range busy timeout")
}

// TestStoreRow_StateDirUnset proves the "not checked" row a missing
// DUBPLATE_STATE_DIR reports, with no failure of its own (the owned
// wall row already reports the missing key).
func TestStoreRow_StateDirUnset(t *testing.T) {
	var ran readerRan
	row, report := Doctor(context.Background(), nil, ran.reader)
	status, failed := row.Status, row.Failed
	if status != "not checked" {
		t.Errorf("Doctor() status = %q, want %q", status, "not checked")
	}
	if failed {
		t.Errorf("Doctor() failed = true, want false")
	}
	ran.requireNotInspected(t, report, "a missing state directory")
}

// readerRan records whether its reader ran, for a test of a row that
// must never reach [Inspect].
type readerRan bool

func (r *readerRan) reader(context.Context, Querier, int) error {
	*r = true
	return nil
}

// requireNotInspected fails t unless report is the zero Report and the
// reader never ran, as Doctor promises when why stops it before
// [Inspect].
func (r *readerRan) requireNotInspected(t *testing.T, report Report, why string) {
	t.Helper()
	if *r || report.State != 0 || report.Facts != nil {
		t.Errorf("Doctor() reader ran = %v, report = %+v; want no reader run and the zero Report: %s never runs Inspect", bool(*r), report, why)
	}
}

// buildZeroByteStoreLeaf creates a leaf under stateDir shaped as
// the store's own leaf (store.lock, db/dubplate.db at
// zero bytes, snapshots/, staging/), so the inspect mode classifies it
// as an interrupted create rather than a real store.
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
