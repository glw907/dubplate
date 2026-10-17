package local

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// The environment a re-executed test binary reads to run as a second
// process: the mode names its job's body, and the directory the state
// directory whose store it opens.
const (
	childModeEnv = "DUBPLATE_OPS_LOCAL_CHILD_MODE"
	childDirEnv  = "DUBPLATE_OPS_LOCAL_CHILD_DIR"
)

// The child's two modes. A holding child's body reports one phase and
// then waits for its context until the process is killed or its stdin
// closes. A canceling child's body reports one phase, then waits 5
// seconds, longer than the cancel poll interval, or for its context,
// and the child exits once its runner closes.
const (
	modeHold   = "hold"
	modeCancel = "cancel"
)

func TestMain(m *testing.M) {
	if mode := os.Getenv(childModeEnv); mode != "" {
		os.Exit(runChild(mode, os.Getenv(childDirEnv)))
	}
	os.Exit(m.Run())
}

// runChild is a second process's whole life: it opens a runner over
// dir's store, starts a backup, and once the body has reported its
// first phase writes the job's handle and its runner's instance to
// descriptor 3.
func runChild(mode, dir string) int {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	fail := func(err error) int {
		logger.Error("child", "err", err)
		return 2
	}
	h, err := store.OpenWrite(context.Background(), dir, testBusy)
	if err != nil {
		return fail(err)
	}
	defer func() { _ = h.Close() }()

	started := make(chan struct{})
	b := func(ctx context.Context, progress func(string) error) (string, error) {
		if err := progress("child working"); err != nil {
			return "", err
		}
		close(started)
		var wait <-chan time.Time
		if mode == modeCancel {
			wait = time.After(5 * time.Second)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-wait:
			return "", nil
		}
	}
	r, err := newRunner(context.Background(), Config{Store: h, Backup: BackupConfig{Enabled: true}, Logger: logger}, b)
	if err != nil {
		return fail(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resp, err := r.StartBackup(ctx, ops.LocalActor("child"), ops.StartBackupRequest{})
	if err != nil {
		return fail(err)
	}
	<-started
	ready := os.NewFile(3, "ready")
	if _, err := fmt.Fprintf(ready, "%s %s\n", resp.JobID, r.lease.Instance()); err != nil {
		return fail(err)
	}
	if err := ready.Close(); err != nil {
		return fail(err)
	}

	if mode == modeHold {
		_, _ = io.Copy(io.Discard, os.Stdin)
		r.CancelOwn(resp.JobID)
	}
	if err := r.Close(); err != nil {
		return fail(err)
	}
	return 0
}

// A child is a re-executed test binary running one job.
type child struct {
	cmd      *exec.Cmd
	handle   ops.Handle
	instance string
	stderr   *syncBuffer
	exited   chan error
}

// startChild re-executes the test binary in mode over dir's store, in
// its own process group, and returns once its job is running. The
// child is killed at cleanup if it is still alive.
func startChild(t *testing.T, dir, mode string) *child {
	t.Helper()
	readyR, readyW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	defer func() { _ = readyR.Close() }()
	stdinR, stdinW, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}

	c := &child{stderr: &syncBuffer{}, exited: make(chan error, 1)}
	c.cmd = exec.Command(os.Args[0], "-test.run=^$")
	c.cmd.Env = append(os.Environ(), childModeEnv+"="+mode, childDirEnv+"="+dir)
	c.cmd.ExtraFiles = []*os.File{readyW}
	c.cmd.Stdin = stdinR
	c.cmd.Stderr = c.stderr
	c.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	if err := c.cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	_ = readyW.Close()
	_ = stdinR.Close()
	go func() { c.exited <- c.cmd.Wait() }()
	t.Cleanup(func() {
		_ = stdinW.Close()
		_ = syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL)
		// SIGKILL makes a nonzero exit expected; log it, never fail on it.
		if err := <-c.exited; err != nil {
			t.Logf("child exit: %v", err)
		}
	})

	line := make(chan string, 1)
	go func() {
		s, _ := bufio.NewReader(readyR).ReadString('\n')
		line <- s
	}()
	select {
	case s := <-line:
		handle, instance, ok := strings.Cut(strings.TrimSpace(s), " ")
		if !ok {
			t.Fatalf("child's ready line %q is not a handle and an instance; stderr:\n%s", s, c.stderr)
		}
		c.handle, c.instance = ops.Handle(handle), instance
	case <-time.After(30 * time.Second):
		t.Fatalf("child never became ready; stderr:\n%s", c.stderr)
	}
	return c
}

// wait waits for the child to exit and fails unless it exits 0.
func (c *child) wait(t *testing.T, limit time.Duration) {
	t.Helper()
	select {
	case err := <-c.exited:
		c.exited <- err
		if err != nil {
			t.Fatalf("child exited %v; stderr:\n%s", err, c.stderr)
		}
	case <-time.After(limit):
		t.Fatalf("child still running after %v; stderr:\n%s", limit, c.stderr)
	}
}

// kill sends SIGKILL to the child's process group and waits for it.
func (c *child) kill(t *testing.T) {
	t.Helper()
	if err := syscall.Kill(-c.cmd.Process.Pid, syscall.SIGKILL); err != nil {
		t.Fatalf("kill child: %v", err)
	}
	err := <-c.exited
	c.exited <- err
	if _, ok := errors.AsType[*exec.ExitError](err); !ok {
		t.Fatalf("child's wait = %v, want it killed", err)
	}
}

// TestCrossProcess_OpenLeavesLiveOwner pins spec criterion 3: a
// runner opened beside another process's running job leaves that job's
// record running, while the same open interrupts a planted record
// whose owner's lease is free, the positive control that proves the
// open recovered at all.
func TestCrossProcess_OpenLeavesLiveOwner(t *testing.T) {
	dir := newStateDir(t)
	c := startChild(t, dir, modeHold)
	h := openWrite(t, dir)
	deadOwner := newInstance()
	touch(t, filepath.Join(leasesDir(dir), deadOwner), 0o600)
	dead := plant(t, h, record{state: "running", owner: deadOwner, started: time.Now().UnixMicro()})

	newRunnerFor(t, h, nil, succeed)

	if rec := mustRecord(t, h, idOf(t, c.handle)); rec.state != "running" {
		t.Errorf("the child's record reads %s, want running", rec.state)
	}
	if rec := mustRecord(t, h, dead); rec.state != "interrupted" || rec.finished == 0 {
		t.Errorf("the dead owner's record reads %s, finished_at %d; want interrupted and finished", rec.state, rec.finished)
	}
}

// TestCrossProcess_KilledOwner pins spec criterion 5: a child killed
// mid-job leaves a record the next open interrupts, and a terminal
// update issued afterward through the runner's own transition is
// refused and logged, leaving the record interrupted.
func TestCrossProcess_KilledOwner(t *testing.T) {
	dir := newStateDir(t)
	c := startChild(t, dir, modeHold)
	c.kill(t)
	h := openWrite(t, dir)
	logs := &syncBuffer{}
	r := newRunnerFor(t, h, logs, succeed)
	id := idOf(t, c.handle)
	if rec := mustRecord(t, h, id); rec.state != "interrupted" {
		t.Fatalf("the killed child's record reads %s after recovery, want interrupted", rec.state)
	}
	if _, err := os.Stat(filepath.Join(leasesDir(dir), c.instance)); !os.IsNotExist(err) {
		t.Errorf("the killed child's lease file survived recovery: stat = %v", err)
	}

	late := &job{id: id, handle: c.handle, ctx: context.Background(), cancel: func() {}}
	if err := r.finish(late, outcome{state: ops.JobSucceeded}); err != nil {
		t.Fatalf("finish: %v", err)
	}
	if rec := mustRecord(t, h, id); rec.state != "interrupted" {
		t.Errorf("record reads %s after a late terminal update, want interrupted", rec.state)
	}
	got := logs.String()
	if !strings.Contains(got, "level=WARN") || !strings.Contains(got, "job end refused") || !strings.Contains(got, string(c.handle)) {
		t.Errorf("logs %q do not warn of the refused update to %s", got, c.handle)
	}
}

// TestCrossProcess_Cancel pins spec criterion 6. A cancel from a
// second runner on the child's running job sets cancel_requested and
// leaves the record running at that moment; the child's runner, polling
// while its body waits 5 seconds, ends it canceled. A queued record
// owned by the live child ends canceled at once, and a succeeded record
// reports succeeded with no write.
func TestCrossProcess_Cancel(t *testing.T) {
	dir := newStateDir(t)
	c := startChild(t, dir, modeCancel)
	h := openWrite(t, dir)
	past := time.Now().Add(-time.Hour).UnixMicro()
	queued := plant(t, h, record{state: "queued", owner: c.instance, at: past})
	done := plant(t, h, record{state: "succeeded", owner: c.instance, at: past, started: past, finished: past})
	r := newRunnerFor(t, h, nil, succeed)

	resp, err := r.CancelJob(deadline(t), testActor, ops.CancelJobRequest{JobID: handleOf(queued)})
	if err != nil || resp.Job.State != ops.JobCanceled {
		t.Errorf("cancel of a queued record = %v, %v; want canceled", resp.Job.State, err)
	}
	resp, err = r.CancelJob(deadline(t), testActor, ops.CancelJobRequest{JobID: handleOf(done)})
	if err != nil || resp.Job.State != ops.JobSucceeded {
		t.Errorf("cancel of a succeeded record = %v, %v; want succeeded", resp.Job.State, err)
	}
	if rec := mustRecord(t, h, done); rec.updated != past || rec.cancelRequested {
		t.Errorf("succeeded record written: updated_at %d, cancel_requested %v", rec.updated, rec.cancelRequested)
	}

	id := idOf(t, c.handle)
	resp, err = r.CancelJob(deadline(t), testActor, ops.CancelJobRequest{JobID: c.handle})
	if err != nil {
		t.Fatalf("cancel of the child's job: %v", err)
	}
	if resp.Job.State != ops.JobRunning {
		t.Errorf("cancel of the child's job reports %s, want running", resp.Job.State)
	}
	if rec := mustRecord(t, h, id); rec.state != "running" || !rec.cancelRequested {
		t.Errorf("child's record right after the cancel = %s, cancel_requested %v; want running, requested", rec.state, rec.cancelRequested)
	}

	c.wait(t, 30*time.Second)
	if rec := mustRecord(t, h, id); rec.state != "canceled" || rec.finished == 0 {
		t.Errorf("child's record = %s, finished_at %d; want canceled and finished", rec.state, rec.finished)
	}
}

// TestCrossProcess_Conflict pins spec criterion 7: a second start
// while the child's backup runs answers conflict naming the child's
// handle.
func TestCrossProcess_Conflict(t *testing.T) {
	dir := newStateDir(t)
	c := startChild(t, dir, modeHold)
	h := openWrite(t, dir)
	r := newRunnerFor(t, h, nil, succeed)

	_, err := r.StartBackup(deadline(t), testActor, ops.StartBackupRequest{})
	if !errors.Is(err, ops.ErrConflict) {
		t.Fatalf("StartBackup() = %v, want a conflict", err)
	}
	if !strings.Contains(err.Error(), string(c.handle)) {
		t.Errorf("conflict %q does not name the child's job %s", err, c.handle)
	}
}
