package store

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// testStamp is the stamp of every restore these tests record.
const testStamp = "20260925T120000.000000Z"

// testIncoming is the staged store's basename every test record names.
const testIncoming = "restore-20260925T120000.000000Z.db"

// testIntent returns a well-formed record in state, owing a pre-restore
// copy, as a restore of a remote snapshot writes it.
func testIntent(state intentState) intentRecord {
	return intentRecord{
		Format:     intentFormat,
		State:      state,
		Stamp:      testStamp,
		Source:     intentSource{Remote: "remote:dubplate", Snapshot: "20260924T030000.000000Z"},
		Incoming:   testIncoming,
		PreRestore: "pre-restore-20260925T120000.000000Z.db",
		Version:    2,
	}
}

// plantIntent writes r as l's record through the record's own writer.
func plantIntent(t *testing.T, l leaf, r intentRecord) {
	t.Helper()
	if err := l.writeIntent(r); err != nil {
		t.Fatalf("write intent: %v", err)
	}
}

// plantIntentFile copies the fixture name under testdata/intent to l's
// record path.
func plantIntentFile(t *testing.T, l leaf, name string) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", "intent", name))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.restoreIntentPath, b, 0o600); err != nil {
		t.Fatal(err)
	}
}

// TestDecodeIntent_Fixtures decodes each fixture under testdata/intent:
// the four well-formed states read back field for field, and each
// malformed record is refused naming what is wrong with it.
func TestDecodeIntent_Fixtures(t *testing.T) {
	remote := intentSource{Remote: "remote:dubplate", Snapshot: "20260924T030000.000000Z"}
	valid := func(state intentState) intentRecord {
		return intentRecord{
			Format: 1, State: state, Stamp: "20260925T120000.000000Z", Source: remote,
			Incoming: "restore-20260925T120000.000000Z.db", PreRestore: "pre-restore-20260925T120000.000000Z.db", Version: 2,
		}
	}
	tests := []struct {
		file    string
		want    intentRecord
		wantErr string
	}{
		{file: "valid-swap.json", want: valid("swap")},
		{file: "valid-reconcile.json", want: valid("reconcile")},
		{file: "valid-reconcile-pending.json", want: valid("reconcile-pending")},
		{file: "valid-library.json", want: intentRecord{
			Format: 1, State: "library", Stamp: "20260925T120000.000000Z",
			Source:   intentSource{File: "/var/lib/dubplate/store/snapshots/pre-migrate-v2.db"},
			Incoming: "restore-20260925T120000.000000Z.db", DiscardCurrent: true,
			Library: &libraryOwed{Root: "/srv/music", IntoNonempty: true}, Version: 3,
		}},
		{file: "empty.json", wantErr: "the record is empty"},
		{file: "malformed.json", wantErr: "parse: unexpected EOF"},
		{file: "unknown-key.json", wantErr: `unknown field "owner"`},
		{file: "unknown-format.json", wantErr: "format 2, want 1"},
		{file: "unknown-state.json", wantErr: `unknown state "copying"`},
		{file: "separator-name.json", wantErr: `field incoming "../db/dubplate.db" is not a file name`},
		{file: "dotdot-name.json", wantErr: `field pre_restore ".." is not a file name`},
		{file: "empty-name.json", wantErr: "field incoming is empty"},
		{file: "bad-stamp.json", wantErr: `field stamp: stamp "2026-09-25"`},
		{file: "library-without-root.json", wantErr: "field library is empty in state library"},
		{file: "library-relative-root.json", wantErr: `field library.root "music" is not an absolute path`},
		{file: "discard-and-copy.json", wantErr: "field pre_restore names pre-restore-20260925T120000.000000Z.db beside discard_current"},
		{file: "trailing-data.json", wantErr: "data after the record's object"},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			b, err := os.ReadFile(filepath.Join("testdata", "intent", tt.file))
			if err != nil {
				t.Fatal(err)
			}
			got, err := decodeIntent(b)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("decode = %v, want an error naming %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("decode = %+v (library %+v), want %+v (library %+v)", got, got.Library, tt.want, tt.want.Library)
			}
		})
	}
}

// TestWriteIntent: the writer lands a record the reader reads back, at
// 0600 with no temp file left, and refuses a malformed record, leaving
// the record already there byte for byte.
func TestWriteIntent(t *testing.T) {
	_, l := newLeafDir(t)
	want := testIntent(intentReconcile)
	plantIntent(t, l, want)

	info, err := os.Stat(l.restoreIntentPath)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("record mode = %04o, want 0600", mode)
	}
	got, err := l.readIntent()
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if got == nil || !reflect.DeepEqual(*got, want) {
		t.Errorf("read back %+v, want %+v", got, want)
	}
	if names := listDir(t, l.root); slices.ContainsFunc(names, func(n string) bool { return strings.Contains(n, ".tmp-") }) {
		t.Errorf("leaf after the write = %q, want no temp file", names)
	}

	before := hashFile(t, l.restoreIntentPath)
	bad := testIntent(intentSwap)
	bad.Incoming = "../db/dubplate.db"
	if err := l.writeIntent(bad); err == nil || !strings.Contains(err.Error(), "field incoming") {
		t.Errorf("write of a record naming a path = %v, want a refusal naming the field", err)
	}
	if hashFile(t, l.restoreIntentPath) != before {
		t.Error("a refused write changed the record already there")
	}
}

// recordingReconciler stands in for serve's reconciler. It records
// each call and what the handle it was given reads, and returns its
// report and err.
type recordingReconciler struct {
	report []byte
	err    error

	mu          sync.Mutex
	calls       int
	version     int
	userVersion int
	jobIDs      []string
	stamp       string
	intent      RestoreIntent
}

func (r *recordingReconciler) reconcile(ctx context.Context, h *Handle, stamp string, in RestoreIntent) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	r.stamp, r.intent = stamp, in
	r.version = h.version
	if err := h.QueryRow(ctx, "PRAGMA user_version").Scan(&r.userVersion); err != nil {
		return nil, err
	}
	rows, err := h.Query(ctx, "SELECT hex(id) FROM jobs ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		r.jobIDs = append(r.jobIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.report, r.err
}

func (r *recordingReconciler) callCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.calls
}

// reconcilingServe opens stateDir in serve mode with rec as its
// reconciler, discarding every log record, and closes the handle at
// cleanup.
func reconcilingServe(t *testing.T, stateDir string, rec *recordingReconciler) error {
	t.Helper()
	opts := serveOpts(slog.New(slog.DiscardHandler))
	opts.Reconcile = rec.reconcile
	h, err := OpenServe(context.Background(), stateDir, opts)
	if err == nil {
		t.Cleanup(func() { _ = h.Close() })
	}
	return err
}

// assertNoIntent fails the test when l still holds a record.
func assertNoIntent(t *testing.T, l leaf) {
	t.Helper()
	if _, err := os.Stat(l.restoreIntentPath); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("restore-intent record after serve: %v, want it removed", err)
	}
}

// assertReport fails the test unless l holds the report for testStamp
// at 0600 carrying want.
func assertReport(t *testing.T, l leaf, want string) {
	t.Helper()
	p := filepath.Join(l.root, "restore-report-"+testStamp+".txt")
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("read report: %v", err)
	}
	if string(got) != want {
		t.Errorf("report = %q, want %q", got, want)
	}
	info, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode != 0o600 {
		t.Errorf("report mode = %04o, want 0600", mode)
	}
}

// TestOpenModes_IntentRecordFiles is R-6bii7's every-mode table, one
// fixture file per row: a record no mode can use is refused by inspect,
// write, and serve alike, naming the path, what is wrong, and restore
// --resume, and serve reconciles nothing. Fixture: a current store.
func TestOpenModes_IntentRecordFiles(t *testing.T) {
	tests := []struct {
		file   string
		detail string
	}{
		{file: "empty.json", detail: "the record is empty"},
		{file: "malformed.json", detail: "unexpected EOF"},
		{file: "unknown-key.json", detail: `unknown field "owner"`},
		{file: "unknown-format.json", detail: "format 2"},
		{file: "unknown-state.json", detail: `unknown state "copying"`},
		{file: "separator-name.json", detail: "field incoming"},
		{file: "dotdot-name.json", detail: "field pre_restore"},
		{file: "empty-name.json", detail: "field incoming is empty"},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead)
			plantIntentFile(t, l, tt.file)
			names := []string{l.restoreIntentPath, "unreadable", tt.detail, "restore --resume"}

			r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if r.State != StateRestoreInProgress {
				t.Errorf("inspect state = %v, want restore in progress", r.State)
			}
			assertErrNames(t, errors.New(r.Reason), names...)
			_, werr := quietWrite(t, stateDir)
			assertErrNames(t, werr, names...)
			rec := &recordingReconciler{}
			serr := reconcilingServe(t, stateDir, rec)
			assertErrNames(t, serr, names...)
			if rec.callCount() != 0 {
				t.Errorf("serve called the reconciler %d times beside an unreadable record", rec.callCount())
			}
		})
	}
}

// TestOpenModes_IntentTempFileIgnored: a temp file the record's writer
// leaves when a crash stops it before its rename is no record, so every
// mode proceeds as it would with none.
func TestOpenModes_IntentTempFileIgnored(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	b, err := os.ReadFile(filepath.Join("testdata", "intent", "valid-swap.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(l.restoreIntentPath+".tmp-4242", b, 0o600); err != nil {
		t.Fatal(err)
	}

	if r := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy}); r.State != StateCurrent {
		t.Errorf("inspect state = %v (%s), want current", r.State, r.Reason)
	}
	w, err := quietWrite(t, stateDir)
	if err != nil {
		t.Fatalf("OpenWrite beside a temp file: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := quietServe(t, stateDir); err != nil {
		t.Fatalf("OpenServe beside a temp file: %v", err)
	}
}

// TestOpenModes_IntentStates is R-6bii7's state table for inspect,
// write, and serve's refusals: each well-formed state is reported and
// refused naming what finishes it, serve refuses every state but
// reconcile-pending without calling its reconciler, and neither open
// touches the staged incoming file a swap record names. Fixture: a
// current store and, for every row, the staged incoming file.
func TestOpenModes_IntentStates(t *testing.T) {
	const (
		resume = "finish the restore with restore --resume"
		serve  = "start serve to finish the restore"
	)
	tests := []struct {
		state      intentState
		wantAction string
	}{
		{state: intentSwap, wantAction: resume},
		{state: intentLibrary, wantAction: resume},
		{state: intentReconcile, wantAction: resume},
		{state: intentReconcilePending, wantAction: serve},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead)
			r := testIntent(tt.state)
			if tt.state == intentLibrary {
				r.Library = &libraryOwed{Root: "/srv/music"}
			}
			plantIntent(t, l, r)
			staged := filepath.Join(l.stagingDir, testIncoming)
			if err := os.WriteFile(staged, []byte("incoming"), 0o600); err != nil {
				t.Fatal(err)
			}
			wantReason := fmt.Sprintf("restore-intent record %s (state %s): %s", l.restoreIntentPath, tt.state, tt.wantAction)

			rep := Inspect(context.Background(), InspectOptions{StateDir: stateDir, BusyTimeout: openBusy})
			if rep.State != StateRestoreInProgress || rep.Reason != wantReason {
				t.Errorf("inspect = %v %q, want restore in progress %q", rep.State, rep.Reason, wantReason)
			}
			if tt.state != intentReconcilePending {
				rec := &recordingReconciler{report: []byte("report")}
				serr := reconcilingServe(t, stateDir, rec)
				if serr == nil || serr.Error() != wantReason {
					t.Errorf("OpenServe = %v, want %q", serr, wantReason)
				}
				if rec.callCount() != 0 {
					t.Errorf("serve called the reconciler %d times beside a %s record", rec.callCount(), tt.state)
				}
			}
			_, werr := quietWrite(t, stateDir)
			if werr == nil || werr.Error() != wantReason {
				t.Errorf("OpenWrite = %v, want %q", werr, wantReason)
			}

			if got := listDir(t, l.stagingDir); !slices.Equal(got, []string{testIncoming}) {
				t.Errorf("staging after the opens = %q, want %q", got, testIncoming)
			}
			if _, err := os.Stat(l.restoreIntentPath); err != nil {
				t.Errorf("record after the opens: %v, want it kept", err)
			}
		})
	}
}

// plantJob inserts one job row with a fixed id into the store at path
// through a connection of its own, and returns the id as hex(id)
// reads it.
func plantJob(t *testing.T, path string) string {
	t.Helper()
	rawExec(t, path, `INSERT INTO jobs (id, kind, at, state, owner_instance, actor, updated_at)
VALUES (X'0123456789ABCDEF0123456789ABCDEF', 'backup', 1, 'succeeded', 'inst', 'local:test', 1)`)
	return "0123456789ABCDEF0123456789ABCDEF"
}

// TestOpenServe_ReconcilePendingOlder is spec criterion 11: a restored
// store below the binary's version with a reconcile-pending record
// refuses a write-mode open naming serve; serve takes a pre-migration
// copy, migrates, calls the reconciler once on a handle at the
// binary's version that reads the store's surviving job, writes the
// report, and removes the record; and a write-mode open then succeeds.
// Fixture: a store at version 2, made by the migration runner from the
// real first two files, holding one job row; a staging leftover named
// as the migration's own stage, which a migration run before the
// staging clear would collide with.
func TestOpenServe_ReconcilePendingOlder(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, 2)
	jobID := plantJob(t, l.dbPath)
	plantIntent(t, l, testIntent(intentReconcilePending))
	leftover := filepath.Join(l.stagingDir, "pre-migrate-v2.db")
	if err := os.WriteFile(leftover, []byte("half a copy"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, werr := quietWrite(t, stateDir)
	assertErrNames(t, werr, l.restoreIntentPath, "reconcile-pending", "start serve")

	rec := &recordingReconciler{report: []byte("rows without files: 0\n")}
	if err := reconcilingServe(t, stateDir, rec); err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if rec.calls != 1 {
		t.Errorf("reconciler calls = %d, want 1", rec.calls)
	}
	if rec.version != binaryHead || rec.userVersion != binaryHead {
		t.Errorf("reconciler's handle: version %d, user_version %d, want %d for both", rec.version, rec.userVersion, binaryHead)
	}
	if !slices.Equal(rec.jobIDs, []string{jobID}) {
		t.Errorf("jobs the reconciler read = %q, want %q", rec.jobIDs, jobID)
	}
	if _, err := os.Stat(filepath.Join(l.snapshotsDir, "pre-migrate-v2.db")); err != nil {
		t.Errorf("the pre-migration copy is missing: %v", err)
	}
	if got := listDir(t, l.stagingDir); len(got) != 0 {
		t.Errorf("staging after serve = %q, want empty", got)
	}
	assertReport(t, l, "rows without files: 0\n")
	assertNoIntent(t, l)

	w, err := quietWrite(t, stateDir)
	if err != nil {
		t.Fatalf("OpenWrite after serve finished the restore: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
}

// TestOpenServe_ReconcilePendingCurrent: a crash after serve migrated
// but before it removed the record leaves a current store beside a
// reconcile-pending record; the next serve reconciles and removes the
// record without migrating, so it takes no pre-migration copy.
func TestOpenServe_ReconcilePendingCurrent(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	plantIntent(t, l, testIntent(intentReconcilePending))

	rec := &recordingReconciler{report: []byte("clean\n")}
	if err := reconcilingServe(t, stateDir, rec); err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	if rec.calls != 1 {
		t.Errorf("reconciler calls = %d, want 1", rec.calls)
	}
	if got := listDir(t, l.snapshotsDir); len(got) != 0 {
		t.Errorf("snapshots after serve = %q, want no copy: nothing migrated", got)
	}
	assertReport(t, l, "clean\n")
	assertNoIntent(t, l)
}

// TestOpenServe_ReconcilerReceivesRecord: the reconciler is handed the
// record's stamp and its source, and the library half exactly when the
// record owes one. Fixture: a current store beside a reconcile-pending
// record, one row per record shape.
func TestOpenServe_ReconcilerReceivesRecord(t *testing.T) {
	library := testIntent(intentReconcilePending)
	library.Library = &libraryOwed{Root: "/srv/music/library", IntoNonempty: true}
	fromFile := testIntent(intentReconcilePending)
	fromFile.Source = intentSource{File: "/var/tmp/dubplate.db"}

	tests := []struct {
		name   string
		record intentRecord
		want   RestoreIntent
	}{
		{
			name:   "snapshot without a library half",
			record: testIntent(intentReconcilePending),
			want:   RestoreIntent{Remote: "remote:dubplate", Snapshot: "20260924T030000.000000Z"},
		},
		{
			name:   "snapshot with a library half",
			record: library,
			want:   RestoreIntent{Remote: "remote:dubplate", Snapshot: "20260924T030000.000000Z", LibraryRoot: "/srv/music/library", IntoNonempty: true},
		},
		{
			name:   "file source",
			record: fromFile,
			want:   RestoreIntent{File: "/var/tmp/dubplate.db"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead)
			plantIntent(t, l, tt.record)

			rec := &recordingReconciler{report: []byte("clean\n")}
			if err := reconcilingServe(t, stateDir, rec); err != nil {
				t.Fatalf("OpenServe: %v", err)
			}
			if rec.stamp != testStamp {
				t.Errorf("stamp = %q, want %q", rec.stamp, testStamp)
			}
			if rec.intent != tt.want {
				t.Errorf("intent = %+v, want %+v", rec.intent, tt.want)
			}
		})
	}
}

// TestOpenServe_ReconcilePendingRefusals: with no reconciler, or one
// that fails, serve refuses naming why, keeps the record for the next
// start, and writes no report. With no reconciler it also leaves the
// store unmigrated. Fixture: a store at version 2.
func TestOpenServe_ReconcilePendingRefusals(t *testing.T) {
	tests := []struct {
		name        string
		rec         *recordingReconciler
		want        string
		wantVersion int
	}{
		{name: "nil reconciler", want: "serve has no reconciler to finish the restore", wantVersion: 2},
		{name: "reconciler fails", rec: &recordingReconciler{err: errors.New("walk /srv/music: permission denied")}, want: "walk /srv/music: permission denied", wantVersion: binaryHead},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, 2)
			plantIntent(t, l, testIntent(intentReconcilePending))
			before := hashFile(t, l.restoreIntentPath)

			opts := serveOpts(slog.New(slog.DiscardHandler))
			if tt.rec != nil {
				opts.Reconcile = tt.rec.reconcile
			}
			h, err := OpenServe(context.Background(), stateDir, opts)
			if err == nil {
				_ = h.Close()
			}
			assertErrNames(t, err, tt.want)

			if hashFile(t, l.restoreIntentPath) != before {
				t.Error("the record changed; want it kept for the next start")
			}
			if _, err := os.Stat(filepath.Join(l.root, "restore-report-"+testStamp+".txt")); !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("report after a refusal: %v, want none", err)
			}
			if got := queryInt(t, l.dbPath, "PRAGMA user_version"); got != tt.wantVersion {
				t.Errorf("user_version = %d, want %d", got, tt.wantVersion)
			}
		})
	}
}

// TestOpenServe_ReconcileReplacesReport: a report an earlier attempt
// wrote at the same stamp, before a crash kept the record, is replaced
// whole, at 0600 whatever mode the earlier file had.
func TestOpenServe_ReconcileReplacesReport(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	plantIntent(t, l, testIntent(intentReconcilePending))
	earlier := filepath.Join(l.root, "restore-report-"+testStamp+".txt")
	if err := os.WriteFile(earlier, []byte("an earlier, longer report\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(earlier, 0o644); err != nil {
		t.Fatal(err)
	}

	rec := &recordingReconciler{report: []byte("second\n")}
	if err := reconcilingServe(t, stateDir, rec); err != nil {
		t.Fatalf("OpenServe: %v", err)
	}
	assertReport(t, l, "second\n")
	assertNoIntent(t, l)
}

// intentChildEnv tells a re-executed test binary which record state
// TestIntentChildRace writes.
const intentChildEnv = "DUBPLATE_STORE_INTENT_CHILD"

// TestIntentChildRace is not a test on its own: it does nothing unless
// a parent test re-executes the binary naming it. The child stands in
// for a restore: it takes the store lock exclusively, prints "locked",
// waits for its stdin to close, writes a record in the named state
// through the record's own writer, and exits, which releases the lock.
func TestIntentChildRace(t *testing.T) {
	state := os.Getenv(intentChildEnv)
	if state == "" {
		return
	}
	l := newLeaf(os.Getenv(childStateDirEnv))
	lock, err := tryLock(l.lockPath, syscall.LOCK_EX)
	if err != nil {
		t.Fatalf("child lock: %v", err)
	}
	fmt.Println("locked")
	_, _ = io.Copy(io.Discard, os.Stdin)
	if err := l.writeIntent(testIntent(intentState(state))); err != nil {
		t.Fatalf("child write: %v", err)
	}
	runtime.KeepAlive(lock)
	os.Exit(0)
}

// startIntentChild re-executes the test binary as TestIntentChildRace
// on stateDir and waits until it holds the store lock exclusively. The
// returned func lets the child write its record and exit, and waits
// for it.
func startIntentChild(t *testing.T, stateDir string, state intentState) (finish func()) {
	t.Helper()
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestIntentChildRace$", "-test.count=1")
	cmd.Env = append(os.Environ(), intentChildEnv+"="+string(state), childStateDirEnv+"="+stateDir)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	out := bufio.NewReader(stdout)
	line, err := out.ReadString('\n')
	if strings.TrimSpace(line) != "locked" {
		_ = stdin.Close()
		rest, _ := io.ReadAll(out)
		_ = cmd.Wait()
		t.Fatalf("child did not take the lock: %q %v\n%s%s", line, err, rest, stderr.Bytes())
	}
	var once sync.Once
	finish = func() {
		once.Do(func() {
			_ = stdin.Close()
			rest, _ := io.ReadAll(out)
			if err := cmd.Wait(); err != nil {
				t.Errorf("child: %v\n%s%s", err, rest, stderr.Bytes())
			}
		})
	}
	t.Cleanup(finish)
	return finish
}

// writeSettle is how long the write-mode race waits for OpenWrite to
// pass its first read of the record. Write mode's bounded wait logs
// nothing, so no signal marks the point; the open reaches it within
// microseconds of starting.
const writeSettle = 300 * time.Millisecond

// TestOpenWrite_IntentWrittenWhileWaiting: a write-mode open waiting
// on the lock a restore holds exclusively refuses the record the
// restore writes before it dies, naming restore --resume, rather than
// opening the store once the lock frees.
func TestOpenWrite_IntentWrittenWhileWaiting(t *testing.T) {
	stateDir := t.TempDir()
	l := buildStore(t, stateDir, binaryHead)
	finish := startIntentChild(t, stateDir, intentSwap)

	ctx := deadline(t, 20*time.Second)
	done := make(chan error, 1)
	go func() {
		h, err := OpenWrite(ctx, stateDir, openBusy)
		if err == nil {
			err = errors.Join(errors.New("opened"), h.Close())
		}
		done <- err
	}()
	time.Sleep(writeSettle)
	finish()

	err := <-done
	want := fmt.Sprintf("restore-intent record %s (state swap): finish the restore with restore --resume", l.restoreIntentPath)
	if err == nil || err.Error() != want {
		t.Errorf("OpenWrite = %v, want %q", err, want)
	}
}

// TestOpenServe_IntentWrittenWhileWaiting: a serve open waiting on its
// shared lock while a restore holds the lock exclusively re-reads the
// record once it wins, and settles it under the exclusive lock: a swap
// record is refused with the staged incoming file untouched and the
// reconciler never called, and a reconcile-pending record beside a
// current store is reconciled, reported, and removed before the store
// opens. The serve log's first holders record, wanting the shared
// lock, marks that the open has passed its first read of the record.
func TestOpenServe_IntentWrittenWhileWaiting(t *testing.T) {
	tests := []struct {
		state   intentState
		wantErr string
	}{
		{state: intentSwap, wantErr: "(state swap): finish the restore with restore --resume"},
		{state: intentReconcilePending},
	}
	for _, tt := range tests {
		t.Run(string(tt.state), func(t *testing.T) {
			stateDir := t.TempDir()
			l := buildStore(t, stateDir, binaryHead)
			staged := filepath.Join(l.stagingDir, testIncoming)
			if tt.state == intentSwap {
				if err := os.WriteFile(staged, []byte("incoming"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			finish := startIntentChild(t, stateDir, tt.state)

			logs := newHoldersHandler("shared")
			rec := &recordingReconciler{report: []byte("raced\n")}
			opts := serveOpts(slog.New(logs))
			opts.Reconcile = rec.reconcile
			done := make(chan serveResult, 1)
			go func() {
				h, err := OpenServe(context.Background(), stateDir, opts)
				done <- serveResult{h, err}
			}()
			select {
			case <-logs.signal:
			case <-time.After(20 * time.Second):
				t.Error("serve logged no shared-lock holders record")
			}
			finish()

			res := <-done
			if res.err == nil {
				if err := res.h.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if tt.wantErr != "" {
				if res.err == nil || !strings.Contains(res.err.Error(), tt.wantErr) {
					t.Fatalf("OpenServe = %v, want an error naming %q", res.err, tt.wantErr)
				}
				if rec.callCount() != 0 {
					t.Errorf("reconciler calls = %d, want 0", rec.callCount())
				}
				if got := listDir(t, l.stagingDir); !slices.Equal(got, []string{testIncoming}) {
					t.Errorf("staging after serve = %q, want %q", got, testIncoming)
				}
				return
			}
			if res.err != nil {
				t.Fatalf("OpenServe: %v", res.err)
			}
			if rec.callCount() != 1 {
				t.Errorf("reconciler calls = %d, want 1", rec.callCount())
			}
			assertReport(t, l, "raced\n")
			assertNoIntent(t, l)
		})
	}
}
