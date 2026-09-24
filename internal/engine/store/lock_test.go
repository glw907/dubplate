package store

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// blockedSignalingCtx wraps a context.Context and closes blocked the
// first time Done is called. acquireBounded's loop reaches ctx.Done()
// only after a busy tryLock attempt, never before, so closing blocked
// proves the waiter made at least one attempt and found the lock held.
// A caller that waits on blocked before releasing the holder therefore
// guarantees the waiter's first attempt happened while the lock was
// still busy, rather than racing the release.
type blockedSignalingCtx struct {
	context.Context
	once    sync.Once
	blocked chan struct{}
}

func (c *blockedSignalingCtx) Done() <-chan struct{} {
	c.once.Do(func() { close(c.blocked) })
	return c.Context.Done()
}

// shmLockFile makes a lock file under /dev/shm, tmpfs on this
// workstation and every dubplate host (the fold probe P9), so its
// device and inode always appear in /proc/locks. t.TempDir() may sit
// on btrfs, where the mechanics review measured stat and /proc/locks
// disagreeing about the device, so it skips this test rather than
// flaking there. The test skips outright when /dev/shm is unavailable.
func shmLockFile(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/dev/shm", "dubplate-store-lock-*")
	if err != nil {
		t.Skipf("no /dev/shm available: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })

	path := filepath.Join(dir, "store.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatalf("create lock file: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("close lock file: %v", err)
	}
	return path
}

// takeRawExclusive holds path's flock exclusively on a fresh
// descriptor, independent of this package's own acquire functions, so
// the test proves those functions see a real external holder rather
// than agreeing with themselves.
func takeRawExclusive(t *testing.T, path string) {
	t.Helper()
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("open lock file: %v", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatalf("take raw exclusive flock: %v", err)
	}
	t.Cleanup(func() { _ = f.Close() })
}

func TestAcquireSharedNonblocking_HeldExclusiveNamesOwnPid(t *testing.T) {
	path := shmLockFile(t)
	takeRawExclusive(t, path)

	_, err := acquireSharedNonblocking(path)
	if err == nil {
		t.Fatal("acquireSharedNonblocking: want busy error, got nil")
	}
	if !errors.Is(err, errLockBusy) {
		t.Errorf("acquireSharedNonblocking error %v does not wrap errLockBusy", err)
	}
	wantPid := strconv.Itoa(os.Getpid())
	if !strings.Contains(err.Error(), wantPid) {
		t.Errorf("busy error %q does not name this process's pid %s", err, wantPid)
	}
}

func TestAcquireShared_BoundedByCtx_HeldExclusive(t *testing.T) {
	path := shmLockFile(t)
	takeRawExclusive(t, path)

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	_, err := acquireShared(ctx, path)
	if err == nil {
		t.Fatal("acquireShared: want busy error after ctx ends, got nil")
	}
	if !errors.Is(err, errLockBusy) {
		t.Errorf("acquireShared error %v does not wrap errLockBusy", err)
	}
	if ctx.Err() == nil {
		t.Error("acquireShared returned before its ctx ended")
	}
}

func TestAcquireExclusive_WaitsForSharedRelease(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "store.lock")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	holder, err := acquireSharedNonblocking(path)
	if err != nil {
		t.Fatalf("acquire shared holder: %v", err)
	}

	// While the shared holder is up, a bounded exclusive request must
	// stay busy. This proves the wait below actually contends rather
	// than trivially succeeding regardless of the holder.
	shortCtx, shortCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer shortCancel()
	if _, err := acquireExclusive(shortCtx, path); err == nil {
		t.Fatal("acquireExclusive: want busy while the shared holder is up, got nil")
	}

	waiterBlocked := make(chan struct{})
	acquired := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		l, err := acquireExclusive(&blockedSignalingCtx{Context: ctx, blocked: waiterBlocked}, path)
		if err == nil {
			err = l.release()
		}
		acquired <- err
	}()

	<-waiterBlocked
	if err := holder.release(); err != nil {
		t.Fatalf("release shared holder: %v", err)
	}

	select {
	case err := <-acquired:
		if err != nil {
			t.Fatalf("acquireExclusive after the shared holder released: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("acquireExclusive did not complete after the shared holder released")
	}
}

func TestProcLockHolders(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		major uint64
		minor uint64
		inode uint64
		want  []string
	}{
		{
			name:  "matching holder line",
			text:  "1: FLOCK  ADVISORY  READ  54300 00:3c:12157465 0 EOF\n",
			major: 0x00,
			minor: 0x3c,
			inode: 12157465,
			want:  []string{"54300"},
		},
		{
			name:  "queued waiter is never a holder",
			text:  "96: -> FLOCK ADVISORY WRITE 63012 00:3c:12157465 0 EOF\n",
			major: 0x00,
			minor: 0x3c,
			inode: 12157465,
			want:  nil,
		},
		{
			name: "holder and waiter on the same file: only the holder is named",
			text: "1: FLOCK  ADVISORY  READ  54300 00:3c:12157465 0 EOF\n" +
				"96: -> FLOCK ADVISORY WRITE 63012 00:3c:12157465 0 EOF\n",
			major: 0x00,
			minor: 0x3c,
			inode: 12157465,
			want:  []string{"54300"},
		},
		{
			name:  "device mismatch names nobody",
			text:  "1: FLOCK  ADVISORY  READ  54300 00:3c:12157465 0 EOF\n",
			major: 0x00,
			minor: 0x99,
			inode: 12157465,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := procLockHolders(tt.text, tt.major, tt.minor, tt.inode)
			if !slices.Equal(got, tt.want) {
				t.Errorf("procLockHolders(%q) = %v, want %v", tt.text, got, tt.want)
			}
		})
	}
}

func TestDevMajorMinor(t *testing.T) {
	tests := []struct {
		name      string
		dev       uint64
		wantMajor uint64
		wantMinor uint64
	}{
		{
			// The fold probe P9 measured /dev/shm's device split as
			// 00:1b, entirely within the low byte-aligned field.
			name:      "low bits only, matching the fold probe's /dev/shm device",
			dev:       0x1b,
			wantMajor: 0x00,
			wantMinor: 0x1b,
		},
		{
			// A dev_t with bits set in every field the glibc macros
			// read, verified independently against a Python
			// implementation of the same two expressions.
			name:      "bits scattered across the high and low fields",
			dev:       0x123400056789,
			wantMajor: 0x1567,
			wantMinor: 0x23400089,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMajor, gotMinor := devMajorMinor(tt.dev)
			if gotMajor != tt.wantMajor || gotMinor != tt.wantMinor {
				t.Errorf("devMajorMinor(%#x) = %#x:%#x, want %#x:%#x", tt.dev, gotMajor, gotMinor, tt.wantMajor, tt.wantMinor)
			}
		})
	}
}
