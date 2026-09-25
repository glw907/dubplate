package local

import (
	"context"
	"errors"
	"log/slog"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// TestRun_CancelFirst pins the cancel-first row: a queued record
// canceled before its run step moves it to running never runs its
// body, and stays canceled.
func TestRun_CancelFirst(t *testing.T) {
	_, h := newStore(t)
	logs := &syncBuffer{}
	r := newRunnerFor(t, h, logs, nil)
	ran := false
	j, err := r.insert(deadline(t), testActor, ops.KindBackup, func(context.Context, func(string) error) (string, error) {
		ran = true
		return "", nil
	})
	if err != nil {
		t.Fatalf("insert: %v", err)
	}
	resp, err := r.CancelJob(deadline(t), testActor, ops.CancelJobRequest{JobID: j.handle})
	if err != nil {
		t.Fatalf("CancelJob: %v", err)
	}
	if resp.Job.State != ops.JobCanceled {
		t.Fatalf("CancelJob() state = %s, want canceled", resp.Job.State)
	}

	r.run(j)
	if ran {
		t.Error("the body ran after its record was canceled")
	}
	if rec := mustRecord(t, h, j.id); rec.state != "canceled" {
		t.Errorf("record reads %s, want canceled", rec.state)
	}
	if got := logs.String(); !strings.Contains(got, "job ended before it ran") || !strings.Contains(got, string(j.handle)) {
		t.Errorf("logs %q do not record the refused start of %s", got, j.handle)
	}
}

// TestCancelJob_States pins the in-process cancel rows. The queued
// record's owner is the runner itself, whose lease is held, so its own
// recovery never interrupts it first.
func TestCancelJob_States(t *testing.T) {
	past := time.Now().Add(-time.Hour).UnixMicro()
	tests := []struct {
		name      string
		handle    func(t *testing.T, h *store.Handle, r *Runner) ops.Handle
		wantErr   error
		wantState ops.JobState
		wantWrite bool
	}{
		{
			name:    "unknown handle",
			handle:  func(*testing.T, *store.Handle, *Runner) ops.Handle { return ops.Handle(newInstance()) },
			wantErr: ops.ErrNotFound,
		},
		{
			name:    "malformed handle",
			handle:  func(*testing.T, *store.Handle, *Runner) ops.Handle { return "not-a-handle" },
			wantErr: ops.ErrInvalidRequest,
		},
		{
			name: "queued",
			handle: func(t *testing.T, h *store.Handle, r *Runner) ops.Handle {
				return handleOf(plant(t, h, record{state: "queued", owner: r.lease.Instance(), at: past}))
			},
			wantState: ops.JobCanceled,
			wantWrite: true,
		},
		{
			name: "terminal",
			handle: func(t *testing.T, h *store.Handle, r *Runner) ops.Handle {
				return handleOf(plant(t, h, record{state: "succeeded", owner: r.lease.Instance(), at: past, started: past, finished: past}))
			},
			wantState: ops.JobSucceeded,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := newStore(t)
			r := newRunnerFor(t, h, nil, succeed)
			jh := tt.handle(t, h, r)
			var before record
			if tt.wantErr == nil {
				before = mustRecord(t, h, idOf(t, jh))
			}

			resp, err := r.CancelJob(deadline(t), testActor, ops.CancelJobRequest{JobID: jh})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("CancelJob() = %v, want %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("CancelJob: %v", err)
			}
			if resp.Job.State != tt.wantState {
				t.Errorf("CancelJob() state = %s, want %s", resp.Job.State, tt.wantState)
			}
			after := mustRecord(t, h, idOf(t, jh))
			if wrote := after.updated != before.updated; wrote != tt.wantWrite {
				t.Errorf("record written = %v, want %v", wrote, tt.wantWrite)
			}
			if tt.wantState == ops.JobCanceled && (after.finished == 0 || resp.Job.FinishedAt == nil) {
				t.Errorf("finished_at %d, response %v: want both set", after.finished, resp.Job.FinishedAt)
			}
			if tt.wantState == ops.JobCanceled && !after.cancelRequested {
				t.Error("cancel_requested is unset on a canceled record")
			}
		})
	}
}

// TestJobStatus pins the status rows: an unknown handle is not found,
// a malformed one invalid, and a known one reads its record.
func TestJobStatus(t *testing.T) {
	_, h := newStore(t)
	r := newRunnerFor(t, h, nil, succeed)
	at := time.Date(2026, 9, 1, 12, 0, 0, 0, time.UTC)
	id := plant(t, h, record{state: "succeeded", owner: newInstance(), at: at.UnixMicro(), started: at.UnixMicro(), finished: at.Add(time.Minute).UnixMicro()})

	if _, err := r.JobStatus(deadline(t), testActor, ops.JobStatusRequest{JobID: ops.Handle(newInstance())}); !errors.Is(err, ops.ErrNotFound) {
		t.Errorf("unknown handle: JobStatus() = %v, want not found", err)
	}
	if _, err := r.JobStatus(deadline(t), testActor, ops.JobStatusRequest{JobID: "ZZ"}); !errors.Is(err, ops.ErrInvalidRequest) {
		t.Errorf("malformed handle: JobStatus() = %v, want invalid request", err)
	}
	resp, err := r.JobStatus(deadline(t), testActor, ops.JobStatusRequest{JobID: handleOf(id)})
	if err != nil {
		t.Fatalf("JobStatus: %v", err)
	}
	got := resp.Job
	if got.ID != handleOf(id) || got.Kind != "backup" || got.State != ops.JobSucceeded || !got.QueuedAt.Equal(at) {
		t.Errorf("JobStatus() = %+v, want id %s, kind backup, succeeded, queued at %v", got, handleOf(id), at)
	}
	if got.FinishedAt == nil || !got.FinishedAt.Equal(at.Add(time.Minute)) {
		t.Errorf("FinishedAt = %v, want %v", got.FinishedAt, at.Add(time.Minute))
	}
}

// TestRetention pins spec criterion 17 and the retention rows. Each
// fixture's records are planted through the write method, the next
// start prunes, and only the records named in keep survive beside the
// start's own. The interrupted fixture is written by recovery and then
// aged by shifting its times back, so a recovery that left finished_at
// NULL leaves nothing to age.
func TestRetention(t *testing.T) {
	day := 24 * time.Hour
	ago := func(d time.Duration) int64 { return time.Now().Add(-d).UnixMicro() }
	tests := []struct {
		name  string
		plant func(t *testing.T, h *store.Handle) (keep, drop [][]byte)
	}{
		{"no record past 90 days", func(t *testing.T, h *store.Handle) ([][]byte, [][]byte) {
			ok := plant(t, h, record{state: "failed", owner: newInstance(), at: ago(10 * day), finished: ago(10 * day)})
			return [][]byte{ok}, nil
		}},
		{"only the newest succeeded record is protected", func(t *testing.T, h *store.Handle) ([][]byte, [][]byte) {
			oldOK := plant(t, h, record{state: "succeeded", owner: newInstance(), at: ago(100 * day), finished: ago(100 * day)})
			oldFail := plant(t, h, record{state: "failed", owner: newInstance(), at: ago(100 * day), finished: ago(100 * day)})
			newOK := plant(t, h, record{state: "succeeded", owner: newInstance(), at: ago(10 * day), finished: ago(10 * day)})
			return [][]byte{newOK}, [][]byte{oldOK, oldFail}
		}},
		{"the only succeeded record of a kind", func(t *testing.T, h *store.Handle) ([][]byte, [][]byte) {
			only := plant(t, h, record{state: "succeeded", owner: newInstance(), at: ago(100 * day), finished: ago(100 * day)})
			return [][]byte{only}, nil
		}},
		{"an interrupted record written by recovery", func(t *testing.T, h *store.Handle) ([][]byte, [][]byte) {
			id := plant(t, h, record{state: "running", owner: newInstance()})
			r, err := newRunner(context.Background(), Config{Store: h, Logger: slog.New(slog.DiscardHandler)}, nil)
			if err != nil {
				t.Fatalf("recovering runner: %v", err)
			}
			if err := r.Close(); err != nil {
				t.Fatalf("close recovering runner: %v", err)
			}
			shift := (100 * day).Microseconds()
			err = h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
				_, err := tx.Exec(ctx, `UPDATE jobs SET at = at - ?, updated_at = updated_at - ?, finished_at = finished_at - ?
WHERE id = ? AND state = 'interrupted'`, shift, shift, shift, id)
				return err
			})
			if err != nil {
				t.Fatalf("age the interrupted record: %v", err)
			}
			return nil, [][]byte{id}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := newStore(t)
			keep, drop := tt.plant(t, h)
			r := newRunnerFor(t, h, nil, succeed)
			resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
			if err != nil {
				t.Fatalf("StartBackup: %v", err)
			}
			if err := r.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
			for _, id := range keep {
				if _, ok := readRecord(t, h, id); !ok {
					t.Errorf("record %x deleted, want it kept", id)
				}
			}
			for _, id := range drop {
				if rec, ok := readRecord(t, h, id); ok {
					t.Errorf("%s record %x kept, want it deleted", rec.state, id)
				}
			}
			if n, want := countRecords(t, h), len(keep)+1; n != want {
				t.Errorf("%d records after the start, want %d (the kept ones and job %s)", n, want, resp.JobID)
			}
		})
	}
}

// TestNew_BindsBackupBody pins the backup kind's binding: a start
// through [New] runs the backup body with the configuration's remote
// and library root, each refused in turn before rclone runs.
func TestNew_BindsBackupBody(t *testing.T) {
	if _, err := exec.LookPath("rclone"); err != nil {
		t.Fatalf("rclone is a prerequisite of this test: %v", err)
	}
	absent := filepath.Join(t.TempDir(), "absent")
	tests := []struct {
		name     string
		remote   string
		root     string
		wantCode string
		wantMsg  string
	}{
		{name: "remote", remote: "no-colon", root: absent, wantCode: "remote-invalid", wantMsg: `"no-colon"`},
		{name: "library root", remote: "name:dir", root: absent, wantCode: "library-missing", wantMsg: absent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := newStore(t)
			r, err := New(context.Background(), Config{
				Store:  h,
				Backup: BackupConfig{Enabled: true, Remote: tt.remote, LibraryRoot: tt.root},
				Logger: slog.New(slog.DiscardHandler),
			})
			if err != nil {
				t.Fatalf("New: %v", err)
			}
			resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
			if err != nil {
				t.Fatalf("StartBackup: %v", err)
			}
			if err := r.Close(); err != nil {
				t.Fatalf("Close: %v", err)
			}
			rec := mustRecord(t, h, idOf(t, resp.JobID))
			if rec.state != "failed" || rec.code != tt.wantCode || !strings.Contains(rec.message, tt.wantMsg) {
				t.Errorf("record = %s, %q, %q; want failed, %q, naming %q", rec.state, rec.code, rec.message, tt.wantCode, tt.wantMsg)
			}
		})
	}
}
