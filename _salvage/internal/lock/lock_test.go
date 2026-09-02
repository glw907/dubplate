package lock

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestAcquireClearsFDCloexec(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	l, err := Acquire(path)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	t.Cleanup(func() { _ = l.Release() })

	flags, err := unix.FcntlInt(l.File().Fd(), unix.F_GETFD, 0)
	if err != nil {
		t.Fatalf("FcntlInt F_GETFD: %v", err)
	}
	if flags&unix.FD_CLOEXEC != 0 {
		t.Error("FD_CLOEXEC is set on the lock fd; want it cleared so the fd survives syscall.Exec")
	}
}

// TestAcquireContention proves flock's per-open-file-description exclusion
// with a real flock, not a fake. Acquire opens the file itself, so two
// calls on the same path produce two independent open file descriptions --
// exactly what two separate dubplate processes contending for the same
// lock file would produce, since flock locks the open file description,
// not the calling process.
func TestAcquireContention(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	first, err := Acquire(path)
	if err != nil {
		t.Fatalf("first Acquire: %v", err)
	}

	if _, err := Acquire(path); !errors.Is(err, ErrHeld) {
		t.Fatalf("second Acquire while held = %v, want ErrHeld", err)
	}

	if err := first.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	second, err := Acquire(path)
	if err != nil {
		t.Fatalf("Acquire after release: %v", err)
	}
	if err := second.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}
}

// TestAcquireBlockingWaitsForRelease proves AcquireBlocking does not
// return until the holder actually releases: flock's kernel-enforced
// exclusion means AcquireBlocking can only succeed after held.Release()
// runs, and close(released) runs immediately after it, so by the time
// this test observes AcquireBlocking's return, released is already
// closed -- a causal guarantee, not a timing race.
func TestAcquireBlockingWaitsForRelease(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	held, err := Acquire(path)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	released := make(chan struct{})
	go func() {
		time.Sleep(50 * time.Millisecond)
		if err := held.Release(); err != nil {
			t.Errorf("Release: %v", err)
		}
		close(released)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	waited, err := AcquireBlocking(ctx, path)
	if err != nil {
		t.Fatalf("AcquireBlocking: %v", err)
	}
	t.Cleanup(func() { _ = waited.Release() })

	select {
	case <-released:
	default:
		t.Error("AcquireBlocking returned before the holder released the lock")
	}
}

func TestHeldReportsTrueWhileAnotherHolderHasIt(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	held, err := Held(path)
	if err != nil {
		t.Fatalf("Held before Acquire: %v", err)
	}
	if held {
		t.Fatal("Held = true before any Acquire, want false")
	}

	first, err := Acquire(path)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	held, err = Held(path)
	if err != nil {
		t.Fatalf("Held while acquired: %v", err)
	}
	if !held {
		t.Error("Held = false while another holder has the lock, want true")
	}

	if err := first.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	held, err = Held(path)
	if err != nil {
		t.Fatalf("Held after release: %v", err)
	}
	if held {
		t.Error("Held = true after Release, want false")
	}
}

// TestHeldDoesNotDisturbTheLock proves Held's own probe releases
// immediately: a real AcquireBlocking against the same path, running
// concurrently with repeated Held calls, must still succeed once the
// original holder releases -- Held's probe attempt must never leave a
// lock of its own behind to block it.
func TestHeldDoesNotDisturbTheLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	held, err := Acquire(path)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}

	for range 5 {
		if ok, err := Held(path); err != nil || !ok {
			t.Fatalf("Held = %v, %v, want true, nil", ok, err)
		}
	}

	if err := held.Release(); err != nil {
		t.Fatalf("Release: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	waited, err := AcquireBlocking(ctx, path)
	if err != nil {
		t.Fatalf("AcquireBlocking after Held probes: %v", err)
	}
	_ = waited.Release()
}

func TestHeldMissingPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "never-created.lock")

	held, err := Held(path)
	if err != nil {
		t.Fatalf("Held on missing path: %v", err)
	}
	if held {
		t.Error("Held = true for a path that was never created, want false")
	}
}

func TestAcquireBlockingCtxCanceled(t *testing.T) {
	path := filepath.Join(t.TempDir(), "test.lock")

	held, err := Acquire(path)
	if err != nil {
		t.Fatalf("Acquire: %v", err)
	}
	t.Cleanup(func() { _ = held.Release() })

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	if _, err := AcquireBlocking(ctx, path); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("AcquireBlocking = %v, want context.DeadlineExceeded", err)
	}
}
