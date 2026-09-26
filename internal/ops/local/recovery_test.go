package local

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// TestNew_RecoversDeadOwners pins spec criterion 4. Owner A's lease
// file exists unlocked and names a queued and a running record; owner
// B has no lease file and names one running record. A terminal record
// naming A stays unchanged, which only a WHERE naming the two live
// states guarantees, since recovery marks A's records by owner.
func TestNew_RecoversDeadOwners(t *testing.T) {
	dir, h := newStore(t)
	a, b := newInstance(), newInstance()
	aLease := filepath.Join(leasesDir(dir), a)
	touch(t, aLease, 0o600)
	past := time.Now().Add(-time.Hour).UnixMicro()
	live := [][]byte{
		plant(t, h, record{state: "queued", owner: a, at: past}),
		plant(t, h, record{state: "running", owner: a, at: past, started: past}),
		plant(t, h, record{state: "running", owner: b, at: past, started: past}),
	}
	done := plant(t, h, record{state: "succeeded", owner: a, at: past, started: past, finished: past})
	doneBefore := mustRecord(t, h, done)

	newRunnerFor(t, h, nil, succeed)

	for _, id := range live {
		rec := mustRecord(t, h, id)
		if rec.state != "interrupted" || rec.finished == 0 || rec.updated == past {
			t.Errorf("record %x = %s, finished_at %d, updated_at %d: want interrupted with both times set now",
				id, rec.state, rec.finished, rec.updated)
		}
	}
	if doneAfter := mustRecord(t, h, done); !reflect.DeepEqual(doneAfter, doneBefore) {
		t.Errorf("terminal record changed: %+v, want %+v", doneAfter, doneBefore)
	}
	if _, err := os.Stat(aLease); !os.IsNotExist(err) {
		t.Errorf("the unlocked lease file survived recovery: stat = %v", err)
	}
}

// TestNew_LeavesHeldOwners pins the held row: a record whose owner's
// lease is held stays as it is, whatever its kind.
func TestNew_LeavesHeldOwners(t *testing.T) {
	_, h := newStore(t)
	held, err := h.TakeInstanceLease()
	if err != nil {
		t.Fatalf("TakeInstanceLease: %v", err)
	}
	t.Cleanup(func() { _ = held.Release() })
	ids := [][]byte{
		plant(t, h, record{state: "running", owner: held.Instance(), started: time.Now().UnixMicro()}),
		plant(t, h, record{kind: "other-kind", state: "queued", owner: held.Instance()}),
	}
	want := []string{"running", "queued"}

	newRunnerFor(t, h, nil, succeed)

	for i, id := range ids {
		if rec := mustRecord(t, h, id); rec.state != want[i] || rec.finished != 0 {
			t.Errorf("record %x = %s, finished_at %d; want %s, unfinished", id, rec.state, rec.finished, want[i])
		}
	}
}

// TestNew_UnreadableLease pins the unreadable row: a lease file the
// runner cannot open fails construction naming its path, and the
// failed runner leaves no lease of its own behind.
func TestNew_UnreadableLease(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Fatal("this test needs a non-root user: root opens a mode-0000 file")
	}
	dir, h := newStore(t)
	owner := newInstance()
	path := filepath.Join(leasesDir(dir), owner)
	touch(t, path, 0)
	id := plant(t, h, record{state: "running", owner: owner, started: time.Now().UnixMicro()})

	_, err := New(context.Background(), Config{Store: h, Logger: slog.New(slog.DiscardHandler)})
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("New() = %v, want a refusal naming %s", err, path)
	}
	if names := leaseNames(t, dir); !slices.Equal(names, []string{owner}) {
		t.Errorf("leases/ holds %v, want only the unreadable %s", names, owner)
	}
	if rec := mustRecord(t, h, id); rec.state != "running" {
		t.Errorf("record reads %s after a refused recovery, want running", rec.state)
	}
}

// touch creates an empty file at path with mode perm, as a dead
// owner's lease file is left.
func touch(t *testing.T, path string, perm os.FileMode) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, perm)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close %s: %v", path, err)
	}
}

// abortOn installs a trigger that fails every write moving a job record
// to state, standing in for a store that refuses the one write a job's
// end depends on (a full disk, a lock held past the busy timeout).
func abortOn(t *testing.T, h *store.Handle, state string) {
	t.Helper()
	err := h.Update(deadline(t), func(ctx context.Context, tx *store.Tx) error {
		_, err := tx.Exec(ctx, fmt.Sprintf(`CREATE TRIGGER abort_%[1]s BEFORE UPDATE OF state ON jobs
WHEN NEW.state = '%[1]s' BEGIN SELECT RAISE(ABORT, 'refused for the test'); END`, state))
		return err
	})
	if err != nil {
		t.Fatalf("install %s trigger: %v", state, err)
	}
}

// TestJobStatus_RecoversOwnStrandedRecord pins that a job whose own
// store write failed does not leave its record live forever under the
// runner's held lease: the next status read finds the record live,
// owned here, and tracked by no job, and interrupts it, so a caller
// polling for a terminal state gets one.
func TestJobStatus_RecoversOwnStrandedRecord(t *testing.T) {
	tests := []struct {
		name     string
		refused  string // the state whose write fails
		stranded string // the state the failure leaves behind
	}{
		{name: "terminal write fails", refused: "succeeded", stranded: "running"},
		{name: "running write fails", refused: "running", stranded: "queued"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, h := newStore(t)
			r := newRunnerFor(t, h, nil, succeed)
			abortOn(t, h, tt.refused)

			resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
			if err != nil {
				t.Fatalf("StartBackup: %v", err)
			}
			r.wg.Wait()
			if rec := mustRecord(t, h, idOf(t, resp.JobID)); rec.state != tt.stranded {
				t.Fatalf("record reads %s after the failed write, want %s", rec.state, tt.stranded)
			}

			jh := resp.JobID
			st, err := r.JobStatus(deadline(t), testActor, ops.JobStatusRequest{JobID: jh})
			if err != nil {
				t.Fatalf("JobStatus: %v", err)
			}
			if st.Job.State != ops.JobInterrupted {
				t.Errorf("JobStatus() state = %s, want interrupted", st.Job.State)
			}
			if rec := mustRecord(t, h, idOf(t, resp.JobID)); rec.state != "interrupted" || rec.finished == 0 {
				t.Errorf("record reads %s, finished_at %d; want interrupted with finished_at set", rec.state, rec.finished)
			}
		})
	}
}

// TestRecovery_LeavesOwnTrackedJob pins that recovering this runner's
// own records spares a job it still runs: a status read and a recovery
// pass during the body both leave the record running.
func TestRecovery_LeavesOwnTrackedJob(t *testing.T) {
	_, h := newStore(t)
	release := make(chan struct{})
	b, started := blockUntil(release)
	r := newRunnerFor(t, h, nil, b)
	resp, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
	if err != nil {
		t.Fatalf("StartBackup: %v", err)
	}
	<-started

	if err := r.recoverJobs(deadline(t)); err != nil {
		t.Fatalf("recoverJobs: %v", err)
	}
	jh := resp.JobID
	st, err := r.JobStatus(deadline(t), testActor, ops.JobStatusRequest{JobID: jh})
	if err != nil {
		t.Fatalf("JobStatus: %v", err)
	}
	if st.Job.State != ops.JobRunning {
		t.Errorf("JobStatus() state = %s while the body runs, want running", st.Job.State)
	}
	close(release)
	waitState(t, h, idOf(t, resp.JobID), "succeeded", 10*time.Second)
}
