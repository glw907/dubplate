package local

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// succeed is a body that returns at once with no final text.
func succeed(context.Context, func(string) error) (string, error) { return "", nil }

// blockUntil returns a body that reports one phase and then waits for
// release or its context, and a channel closed once it has reported.
func blockUntil(release <-chan struct{}) (body, <-chan struct{}) {
	started := make(chan struct{})
	return func(ctx context.Context, progress func(string) error) (string, error) {
		if err := progress("blocking"); err != nil {
			return "", err
		}
		close(started)
		select {
		case <-release:
			return "", nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}, started
}

func TestNew_RefusesNilLogger(t *testing.T) {
	dir, h := newStore(t)
	_, err := New(context.Background(), Config{Store: h})
	if err == nil || !strings.Contains(err.Error(), "nil logger") {
		t.Fatalf("New() = %v, want a nil-logger refusal", err)
	}
	if names := leaseNames(t, dir); len(names) != 0 {
		t.Errorf("leases/ holds %v after a refused New, want nothing", names)
	}
}

// TestClose_ReleasesInstanceLease pins that a runner's instance lease
// file exists while it is open and is gone after its close.
func TestClose_ReleasesInstanceLease(t *testing.T) {
	dir, h := newStore(t)
	r, err := New(context.Background(), Config{Store: h, Logger: slog.New(slog.DiscardHandler)})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	instance := r.lease.Instance()
	if !onlyLease(t, dir, instance) {
		t.Fatalf("leases/ holds %v while open, want only %s", leaseNames(t, dir), instance)
	}
	if err := r.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if names := leaseNames(t, dir); slices.Contains(names, instance) {
		t.Errorf("leases/ holds %v after Close, want no %s", names, instance)
	}
}

// TestCatalog_BackupAvailability pins spec criterion 8's runner half:
// the start operation's availability follows the flag, with the exact
// reason while it is off, and the other three always answer.
func TestCatalog_BackupAvailability(t *testing.T) {
	tests := []struct {
		name       string
		enabled    bool
		wantAvail  bool
		wantReason string
	}{
		{"flag on", true, true, ""},
		{"flag off", false, false, "backup integration disabled"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := newStore(t)
			r, err := New(context.Background(), Config{Store: h, Backup: BackupConfig{Enabled: tt.enabled}, Logger: slog.New(slog.DiscardHandler)})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			t.Cleanup(func() { _ = r.Close() })
			resp, err := r.Catalog(context.Background(), testActor, ops.CatalogRequest{})
			if err != nil {
				t.Fatalf("Catalog: %v", err)
			}
			want := []ops.CatalogEntry{
				{Name: "start_backup", Privilege: ops.PrivilegeAdmin, Available: tt.wantAvail, Reason: tt.wantReason},
				{Name: "job_status", Privilege: ops.PrivilegeAdmin, Available: true},
				{Name: "cancel_job", Privilege: ops.PrivilegeAdmin, Available: true},
				{Name: "catalog", Privilege: ops.PrivilegeMember, Available: true},
			}
			if !slices.Equal(resp.Operations, want) {
				t.Errorf("Catalog() = %+v, want %+v", resp.Operations, want)
			}
		})
	}
}

// TestStartBackup_FlagOff pins spec criterion 8's start half: the
// start answers unavailable with the exact reason, writes no record,
// and takes no kind lease.
func TestStartBackup_FlagOff(t *testing.T) {
	dir, h := newStore(t)
	r, err := New(context.Background(), Config{Store: h, Logger: slog.New(slog.DiscardHandler)})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })

	_, err = r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
	var opsErr *ops.Error
	if !errors.As(err, &opsErr) || !errors.Is(err, ops.ErrUnavailable) {
		t.Fatalf("StartBackup() = %v, want an unavailable *ops.Error", err)
	}
	if opsErr.Reason != "backup integration disabled" {
		t.Errorf("reason = %q, want %q", opsErr.Reason, "backup integration disabled")
	}
	if n := countRecords(t, h); n != 0 {
		t.Errorf("%d records after a refused start, want 0", n)
	}
	if !onlyLease(t, dir, r.lease.Instance()) {
		t.Errorf("leases/ holds %v, want only the instance lease %s", leaseNames(t, dir), r.lease.Instance())
	}
}

// TestStartBackup_RecordsActor pins spec criterion 2's record half:
// the start's actor parameter is the record's actor, and the runner's
// own instance its owner.
func TestStartBackup_RecordsActor(t *testing.T) {
	_, h := newStore(t)
	r := newRunnerFor(t, h, nil, succeed)
	resp, err := r.StartBackup(deadline(t), ops.LocalActor("alice"), ops.StartBackupRequest{})
	if err != nil {
		t.Fatalf("StartBackup: %v", err)
	}
	rec := waitState(t, h, idOf(t, resp.JobID), "succeeded", 10*time.Second)
	if rec.actor != "local:alice" {
		t.Errorf("actor = %q, want %q", rec.actor, "local:alice")
	}
	if rec.owner != r.lease.Instance() {
		t.Errorf("owner = %q, want the runner's instance %q", rec.owner, r.lease.Instance())
	}
	if rec.started == 0 || rec.finished == 0 {
		t.Errorf("started_at %d, finished_at %d: want both set", rec.started, rec.finished)
	}
}

// TestStartBackup_KindLeaseBusy pins the two conflict rows: a live
// record is named by its handle, and with none left after recovery the
// answer names a transfer that outlived its process. The second row's
// store holds a dead owner's running record, which recovery must
// interrupt before the answer, so an answer given before recovery
// names that record's handle instead.
func TestStartBackup_KindLeaseBusy(t *testing.T) {
	t.Run("a live record", func(t *testing.T) {
		_, h := newStore(t)
		release := make(chan struct{})
		defer close(release)
		b, started := blockUntil(release)
		first := newRunnerFor(t, h, nil, b)
		resp, err := first.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
		if err != nil {
			t.Fatalf("first StartBackup: %v", err)
		}
		<-started

		second := newRunnerFor(t, h, nil, succeed)
		_, err = second.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
		if !errors.Is(err, ops.ErrConflict) {
			t.Fatalf("second StartBackup() = %v, want a conflict", err)
		}
		if !strings.Contains(err.Error(), string(resp.JobID)) {
			t.Errorf("conflict %q does not name the running job %s", err, resp.JobID)
		}
	})

	t.Run("no live record after recovery", func(t *testing.T) {
		_, h := newStore(t)
		held, err := h.TakeKindLease(ops.KindBackup)
		if err != nil {
			t.Fatalf("TakeKindLease: %v", err)
		}
		t.Cleanup(func() { _ = held.Release() })
		r := newRunnerFor(t, h, nil, succeed)
		dead := plant(t, h, record{state: "running", owner: newInstance(), started: time.Now().UnixMicro()})

		_, err = r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
		if !errors.Is(err, ops.ErrConflict) {
			t.Fatalf("StartBackup() = %v, want a conflict", err)
		}
		if want := "a previous run's transfer is still running"; !strings.Contains(err.Error(), want) {
			t.Errorf("conflict %q, want it to say %q", err, want)
		}
		if strings.Contains(err.Error(), fmt.Sprintf("%x", dead)) {
			t.Errorf("conflict %q names the dead owner's record", err)
		}
		if rec := mustRecord(t, h, dead); rec.state != "interrupted" {
			t.Errorf("dead owner's record reads %s, want interrupted", rec.state)
		}
		if n := countRecords(t, h); n != 1 {
			t.Errorf("%d records, want only the planted one", n)
		}
	})
}

// codedErr is a body error that names its own failure code.
type codedErr struct{ code, msg string }

func (e *codedErr) Error() string       { return e.msg }
func (e *codedErr) FailureCode() string { return e.code }

// TestRun_BodyOutcomes pins how a body's return reaches its record: a
// coded error's code and message, any other error as job-error, and a
// success's final progress text.
func TestRun_BodyOutcomes(t *testing.T) {
	tests := []struct {
		name         string
		final        string
		err          error
		wantState    string
		wantCode     string
		wantMessage  string
		wantProgress string
	}{
		{"coded error", "", fmt.Errorf("send: %w", &codedErr{code: "rclone-exit-7", msg: "sync failed"}), "failed", "rclone-exit-7", "send: sync failed", "started"},
		{"other error", "", errors.New("disk on fire"), "failed", "job-error", "disk on fire", "started"},
		{"success with final text", "prune skipped: clock behind", nil, "succeeded", "", "", "prune skipped: clock behind"},
		{"success without final text", "", nil, "succeeded", "", "", "started"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := newStore(t)
			r := newRunnerFor(t, h, nil, func(_ context.Context, progress func(string) error) (string, error) {
				if err := progress("started"); err != nil {
					return "", err
				}
				return tt.final, tt.err
			})
			resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
			if err != nil {
				t.Fatalf("StartBackup: %v", err)
			}
			if err := r.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
			rec := mustRecord(t, h, idOf(t, resp.JobID))
			if rec.state != tt.wantState || rec.code != tt.wantCode || rec.message != tt.wantMessage || rec.progress != tt.wantProgress {
				t.Errorf("record = %s code %q message %q progress %q, want %s code %q message %q progress %q",
					rec.state, rec.code, rec.message, rec.progress, tt.wantState, tt.wantCode, tt.wantMessage, tt.wantProgress)
			}
			if rec.finished == 0 {
				t.Error("finished_at is NULL on a terminal record")
			}
		})
	}
}

// TestClose_WaitsForRunningJob pins that Close returns only after a
// running job's goroutine has written its terminal state.
func TestClose_WaitsForRunningJob(t *testing.T) {
	_, h := newStore(t)
	release := make(chan struct{})
	b, started := blockUntil(release)
	r, err := newRunner(context.Background(), Config{Store: h, Backup: BackupConfig{Enabled: true}, Logger: slog.New(slog.DiscardHandler)}, b)
	if err != nil {
		t.Fatalf("newRunner: %v", err)
	}
	resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
	if err != nil {
		t.Fatalf("StartBackup: %v", err)
	}
	<-started

	closed := make(chan error, 1)
	go func() { closed <- r.Close() }()
	select {
	case err := <-closed:
		t.Fatalf("Close returned %v while the job still ran", err)
	case <-time.After(200 * time.Millisecond):
	}
	close(release)
	if err := <-closed; err != nil {
		t.Fatalf("Close: %v", err)
	}
	if rec := mustRecord(t, h, idOf(t, resp.JobID)); rec.state != "succeeded" {
		t.Errorf("record reads %s when Close returned, want succeeded", rec.state)
	}
}

// TestStartBackup_ReleasesKindLease pins that a finished job frees its
// kind for the next start.
func TestStartBackup_ReleasesKindLease(t *testing.T) {
	_, h := newStore(t)
	r := newRunnerFor(t, h, nil, succeed)
	for i := range 2 {
		resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
		if err != nil {
			t.Fatalf("start %d: %v", i, err)
		}
		waitState(t, h, idOf(t, resp.JobID), "succeeded", 10*time.Second)
		r.wg.Wait()
	}
	kl, err := h.TakeKindLease(ops.KindBackup)
	if errors.Is(err, store.ErrLeaseBusy) {
		t.Fatal("kind lease still held after both jobs ended")
	}
	if err != nil {
		t.Fatalf("TakeKindLease: %v", err)
	}
	_ = kl.Release()
}

// TestSeamBoundary_ContextCanceled pins R-6bii5's seam boundary: a
// caller whose context is already canceled at the method's return
// sees an error matching [ops.ErrCanceled], never the store's own
// classification of the aborted call.
func TestSeamBoundary_ContextCanceled(t *testing.T) {
	_, h := newStore(t)
	r := newRunnerFor(t, h, nil, succeed)
	id := plant(t, h, record{state: "queued"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.JobStatus(ctx, testActor, ops.JobStatusRequest{JobID: handleOf(id)})
	if !errors.Is(err, ops.ErrCanceled) {
		t.Fatalf("JobStatus() with an already-canceled context = %v, want it to match ops.ErrCanceled", err)
	}
	opsErr, ok := errors.AsType[*ops.Error](err)
	if !ok {
		t.Fatalf("JobStatus() with an already-canceled context = %v, want an *ops.Error in its chain", err)
	}
	if opsErr.Reason != "" {
		t.Fatalf("JobStatus() with an already-canceled context has Reason %q, want it empty: the aborted store call's text must not reach the wire body", opsErr.Reason)
	}
}

// TestSeamBoundary_ContextAliveClassifiesItsOwnError pins the seam
// boundary's other row: a caller whose context is still alive gets
// the store's own classification, never canceled, even for a job
// handle that does not exist.
func TestSeamBoundary_ContextAliveClassifiesItsOwnError(t *testing.T) {
	_, h := newStore(t)
	r := newRunnerFor(t, h, nil, succeed)
	missing := ops.Handle("0123456789abcdef0123456789abcdef")

	_, err := r.JobStatus(deadline(t), testActor, ops.JobStatusRequest{JobID: missing})
	if !errors.Is(err, ops.ErrNotFound) {
		t.Fatalf("JobStatus() for a missing handle = %v, want it to match ops.ErrNotFound", err)
	}
	if errors.Is(err, ops.ErrCanceled) {
		t.Fatalf("JobStatus() for a missing handle also matches ops.ErrCanceled, want it not to")
	}
}

// TestCancelJob_ContextCanceledDuringWrite pins R-6bii5's write-pool
// row: a context canceled while CancelJob waits for the write pool's
// one connection, held here by a concurrent Update on the same
// handle, answers ops.ErrCanceled well before the store's busy
// timeout, because database/sql's connection wait honors ctx directly
// rather than falling through to SQLite's busy handler.
func TestCancelJob_ContextCanceledDuringWrite(t *testing.T) {
	_, h := newStore(t)
	r := newRunnerFor(t, h, nil, succeed)
	id := plant(t, h, record{state: "running", owner: r.lease.Instance()})

	holding := make(chan struct{})
	release := make(chan struct{})
	done := make(chan error, 1)
	go func() {
		done <- h.Update(deadline(t), func(context.Context, *store.Tx) error {
			close(holding)
			<-release
			return nil
		})
	}()
	<-holding

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	start := time.Now()
	_, err := r.CancelJob(ctx, testActor, ops.CancelJobRequest{JobID: handleOf(id)})
	elapsed := time.Since(start)

	close(release)
	if updateErr := <-done; updateErr != nil {
		t.Fatalf("the blocking Update: %v", updateErr)
	}

	if !errors.Is(err, ops.ErrCanceled) {
		t.Fatalf("CancelJob() while the write pool was held = %v, want it to match ops.ErrCanceled", err)
	}
	if elapsed >= testBusy {
		t.Errorf("CancelJob() took %v to answer, want well under the %v busy timeout", elapsed, testBusy)
	}
}
