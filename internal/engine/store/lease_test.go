package store

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"
)

// leaseChildEnv tells a re-executed test binary to act as a transfer
// child that holds its inherited descriptors until its stdin closes.
const leaseChildEnv = "DUBPLATE_STORE_LEASE_CHILD"

// currentStore returns a state directory holding a store at the
// binary's version, with no handle left open, so the only locks a test
// meets are the ones it takes.
func currentStore(t *testing.T) string {
	t.Helper()
	stateDir := t.TempDir()
	h, err := OpenServe(context.Background(), stateDir, serveOpts(slog.New(slog.DiscardHandler)))
	if err != nil {
		t.Fatalf("serve create: %v", err)
	}
	if err := h.Close(); err != nil {
		t.Fatalf("close serve handle: %v", err)
	}
	return stateDir
}

// writeHandle opens stateDir in write mode and closes the handle at
// cleanup unless the test closed it first.
func writeHandle(t *testing.T, stateDir string) *Handle {
	t.Helper()
	h, err := OpenWrite(deadline(t, 5*time.Second), stateDir, openBusy)
	if err != nil {
		t.Fatalf("open write handle: %v", err)
	}
	t.Cleanup(func() { _ = h.Close() })
	return h
}

// takeInstance takes an instance lease through h, released at cleanup
// unless the test released it first.
func takeInstance(t *testing.T, h *Handle) *InstanceLease {
	t.Helper()
	l, err := h.TakeInstanceLease()
	if err != nil {
		t.Fatalf("take instance lease: %v", err)
	}
	t.Cleanup(func() { _ = l.Release() })
	return l
}

// assertBusyMessage fails unless err reads as a busy lease in either
// form: the gate may run on btrfs, where /proc/locks names no pid.
func assertBusyMessage(t *testing.T, err error) {
	t.Helper()
	if msg := err.Error(); !strings.Contains(msg, "acquired by pid ") && !strings.Contains(msg, "held by another process") {
		t.Errorf("busy error %q names neither an acquirer pid nor another process", msg)
	}
}

// TestKindLease_TwoHandles: flock(2) treats two open file descriptions
// in one process independently, so a second handle meets the first's
// kind lease as busy and its instance lease as held.
func TestKindLease_TwoHandles(t *testing.T) {
	stateDir := currentStore(t)
	h1, h2 := writeHandle(t, stateDir), writeHandle(t, stateDir)
	i1 := takeInstance(t, h1)
	takeInstance(t, h2)

	k1, err := h1.TakeKindLease("backup")
	if err != nil {
		t.Fatalf("first kind lease: %v", err)
	}
	_, err = h2.TakeKindLease("backup")
	if !errors.Is(err, ErrLeaseBusy) {
		t.Fatalf("second kind lease err = %v, want ErrLeaseBusy", err)
	}
	assertBusyMessage(t, err)

	got, err := h2.TryLease(i1.Instance())
	if err != nil {
		t.Fatalf("try first instance: %v", err)
	}
	if got != LeaseHeld {
		t.Errorf("try first instance = %v, want held", got)
	}
	if _, err := os.Stat(filepath.Join(newLeaf(stateDir).leasesDir, i1.Instance())); err != nil {
		t.Errorf("a held lease's file is gone after a try: %v", err)
	}

	if err := k1.Release(); err != nil {
		t.Fatalf("release first kind lease: %v", err)
	}
	k2, err := h2.TakeKindLease("backup")
	if err != nil {
		t.Fatalf("kind lease after release: %v", err)
	}
	if err := k2.Release(); err != nil {
		t.Fatalf("release second kind lease: %v", err)
	}
}

// TestTakeKindLease_Refusals: a kind that is not lowercase ASCII
// letters, digits, and hyphens is refused before any path is built
// from it, so it can never join outside leases/. A shape the rule
// admits still takes the lease.
func TestTakeKindLease_Refusals(t *testing.T) {
	stateDir := currentStore(t)
	h := writeHandle(t, stateDir)
	for _, kind := range []string{
		"",
		"../store",
		"back/up",
		"back\x00up",
		"Backup",
		"back up",
		strings.Repeat("a", maxKindLen+1),
	} {
		if _, err := h.TakeKindLease(kind); err == nil {
			t.Errorf("take kind %q: got no error", kind)
		}
	}

	l, err := h.TakeKindLease("backup-copy")
	if err != nil {
		t.Fatalf("take kind %q: %v", "backup-copy", err)
	}
	if err := l.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
}

// TestTryLease_DeadOwner: an owner that closed its lease without
// removing it (a process that died) reads free once, which removes the
// file, and absent after that.
func TestTryLease_DeadOwner(t *testing.T) {
	stateDir := currentStore(t)
	h := writeHandle(t, stateDir)
	l, err := h.TakeInstanceLease()
	if err != nil {
		t.Fatalf("take instance lease: %v", err)
	}
	if err := l.file.Close(); err != nil {
		t.Fatalf("close lease without removal: %v", err)
	}
	path := filepath.Join(newLeaf(stateDir).leasesDir, l.Instance())

	for _, want := range []LeaseState{LeaseFree, LeaseAbsent} {
		got, err := h.TryLease(l.Instance())
		if err != nil {
			t.Fatalf("try: %v", err)
		}
		if got != want {
			t.Fatalf("try = %v, want %v", got, want)
		}
		if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
			t.Errorf("after a %v try, stat %s = %v, want not exist", want, path, err)
		}
	}
}

// TestTryLease_RemovalRace: two recoveries open the same dead lease
// file; the first locks and removes it, and the second then wins the
// lock on the unlinked inode, whose removal answers not-exist. That
// answer still proves the owner dead, so both read free.
func TestTryLease_RemovalRace(t *testing.T) {
	stateDir := currentStore(t)
	h := writeHandle(t, stateDir)
	l, err := h.TakeInstanceLease()
	if err != nil {
		t.Fatalf("take instance lease: %v", err)
	}
	if err := l.file.Close(); err != nil {
		t.Fatalf("close lease without removal: %v", err)
	}
	path := filepath.Join(newLeaf(stateDir).leasesDir, l.Instance())
	first, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	second, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}

	for i, f := range []*os.File{first, second} {
		got, err := tryOpenLease(f)
		if err != nil {
			t.Fatalf("try %d: %v", i+1, err)
		}
		if got != LeaseFree {
			t.Errorf("try %d = %v, want free", i+1, got)
		}
	}
}

// TestTryLease_Refusals: an instance that is not the shape a lease is
// named with is refused before any path is built from it, and a lease
// file that cannot be opened for a reason other than absence is
// refused naming its path.
func TestTryLease_Refusals(t *testing.T) {
	stateDir := currentStore(t)
	h := writeHandle(t, stateDir)
	for _, instance := range []string{"", "../store.lock", strings.Repeat("A", 32), strings.Repeat("a", 31)} {
		if _, err := h.TryLease(instance); err == nil {
			t.Errorf("try %q: got no error", instance)
		}
	}

	unreadable := strings.Repeat("0", 32)
	path := filepath.Join(newLeaf(stateDir).leasesDir, unreadable)
	if err := os.WriteFile(path, nil, 0); err != nil {
		t.Fatal(err)
	}
	if os.Geteuid() == 0 {
		t.Skip("root opens a mode-0 file")
	}
	_, err := h.TryLease(unreadable)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Errorf("try an unreadable lease: err = %v, want a refusal naming %s", err, path)
	}
	if _, serr := os.Stat(path); serr != nil {
		t.Errorf("a refused try removed %s: %v", path, serr)
	}
}

// TestInstanceLease_Lifecycle: opening and closing a handle creates no
// lease, a taken lease is one file named by a 32-character lowercase
// hex instance, and its release leaves leases/ empty again.
func TestInstanceLease_Lifecycle(t *testing.T) {
	stateDir := currentStore(t)
	leases := newLeaf(stateDir).leasesDir
	h := writeHandle(t, stateDir)
	if got := listDir(t, leases); len(got) != 0 {
		t.Fatalf("leases/ after an open = %q, want empty", got)
	}

	l, err := h.TakeInstanceLease()
	if err != nil {
		t.Fatalf("take instance lease: %v", err)
	}
	if !regexp.MustCompile(`^[0-9a-f]{32}$`).MatchString(l.Instance()) {
		t.Errorf("instance %q is not 32 lowercase hex characters", l.Instance())
	}
	if got := listDir(t, leases); len(got) != 1 || got[0] != l.Instance() {
		t.Errorf("leases/ while held = %q, want [%s]", got, l.Instance())
	}
	if err := l.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if got := listDir(t, leases); len(got) != 0 {
		t.Errorf("leases/ after release = %q, want empty", got)
	}

	if err := h.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	if got := listDir(t, leases); len(got) != 0 {
		t.Errorf("leases/ after close = %q, want empty", got)
	}
}

// TestInstanceLease_Collision: an instance whose file already exists is
// refused naming the path, and the existing file is left as it was.
func TestInstanceLease_Collision(t *testing.T) {
	dir := t.TempDir()
	instance := strings.Repeat("ab", 16)
	path := filepath.Join(dir, instance)
	if err := os.WriteFile(path, []byte("another owner"), lockFilePerm); err != nil {
		t.Fatal(err)
	}
	_, err := takeInstanceLease(dir, instance)
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("take a colliding instance: err = %v, want a refusal naming %s", err, path)
	}
	got, rerr := os.ReadFile(path)
	if rerr != nil || string(got) != "another owner" {
		t.Errorf("the colliding file reads %q (%v), want it unchanged", got, rerr)
	}
}

// TestLeaseModes: under a 0022 umask, serve creates leases/ at 0700 and
// both lease kinds create their files at 0600.
func TestLeaseModes(t *testing.T) {
	old := syscall.Umask(0o022)
	t.Cleanup(func() { syscall.Umask(old) })

	stateDir := currentStore(t)
	leases := newLeaf(stateDir).leasesDir
	h := writeHandle(t, stateDir)
	i := takeInstance(t, h)
	k, err := h.TakeKindLease("backup")
	if err != nil {
		t.Fatalf("take kind lease: %v", err)
	}
	t.Cleanup(func() { _ = k.Release() })

	tests := []struct {
		path string
		want os.FileMode
	}{
		{leases, leafDirPerm},
		{filepath.Join(leases, i.Instance()), 0o600},
		{filepath.Join(leases, "kind-backup"), 0o600},
	}
	for _, tt := range tests {
		info, err := os.Stat(tt.path)
		if err != nil {
			t.Fatalf("stat %s: %v", tt.path, err)
		}
		if got := info.Mode().Perm(); got != tt.want {
			t.Errorf("%s: mode %04o, want %04o", tt.path, got, tt.want)
		}
	}
}

// TestOpenWrite_LeafWithoutLeases: a store whose leaf lacks leases/ is
// refused by write mode, naming the directory and the fix, and serve
// then finishes the leaf.
func TestOpenWrite_LeafWithoutLeases(t *testing.T) {
	stateDir := currentStore(t)
	leases := newLeaf(stateDir).leasesDir
	if err := os.Remove(leases); err != nil {
		t.Fatal(err)
	}

	_, err := quietWrite(t, stateDir)
	if err == nil {
		t.Fatal("write mode opened a leaf without leases/")
	}
	assertErrNames(t, err, leases, "start serve to finish creating it")
	if got := errState(t, err); got != StateNotAStore {
		t.Errorf("state = %v, want not a store", got)
	}
	if _, serr := os.Stat(leases); !errors.Is(serr, os.ErrNotExist) {
		t.Errorf("write mode created %s: %v", leases, serr)
	}

	if _, err := quietServe(t, stateDir); err != nil {
		t.Fatalf("serve: %v", err)
	}
	info, err := os.Stat(leases)
	if err != nil {
		t.Fatalf("serve left no %s: %v", leases, err)
	}
	if got := info.Mode().Perm(); got != leafDirPerm {
		t.Errorf("%s: mode %04o, want %04o", leases, got, leafDirPerm)
	}
}

// TestOpenServe_NilLogger: serve refuses options without a logger
// before it touches the state directory.
func TestOpenServe_NilLogger(t *testing.T) {
	stateDir := t.TempDir()
	opts := serveOpts(nil)
	_, err := OpenServe(context.Background(), stateDir, opts)
	if err == nil || !strings.Contains(err.Error(), "nil logger") {
		t.Fatalf("serve with a nil logger: err = %v, want a nil-logger refusal", err)
	}
	if got := listDir(t, stateDir); len(got) != 0 {
		t.Errorf("serve with a nil logger created %q", got)
	}
}

// TestLeaseChildHold is not a test on its own: it does nothing unless a
// parent test re-executes the binary naming it. The child holds the
// descriptors it inherited until its stdin closes.
func TestLeaseChildHold(t *testing.T) {
	if os.Getenv(leaseChildEnv) == "" {
		return
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	os.Exit(0)
}

// fdCloexec reports whether f's descriptor has FD_CLOEXEC set.
func fdCloexec(t *testing.T, f *os.File) bool {
	t.Helper()
	flags, _, errno := syscall.Syscall(syscall.SYS_FCNTL, f.Fd(), syscall.F_GETFD, 0)
	if errno != 0 {
		t.Fatalf("F_GETFD %s: %v", f.Name(), errno)
	}
	return flags&syscall.FD_CLOEXEC != 0
}

// storeLockBusy reports whether an exclusive store lock is refused,
// which it is while any process holds the lock shared.
func storeLockBusy(t *testing.T, stateDir string) bool {
	t.Helper()
	l, err := tryLock(newLeaf(stateDir).lockPath, syscall.LOCK_EX)
	if errors.Is(err, errLockBusy) {
		return true
	}
	if err != nil {
		t.Fatalf("try the store lock: %v", err)
	}
	if err := l.release(); err != nil {
		t.Fatal(err)
	}
	return false
}

// TestKindLease_DupFiles: the duplicates are close-on-exec, closing them
// in the parent leaves the store lock held, and a child that inherits
// them keeps both the kind lease and the store lock held after the
// parent closes everything of its own, until the child exits.
func TestKindLease_DupFiles(t *testing.T) {
	stateDir := currentStore(t)
	h1 := writeHandle(t, stateDir)
	k, err := h1.TakeKindLease("backup")
	if err != nil {
		t.Fatalf("take kind lease: %v", err)
	}

	dups, err := k.DupFiles()
	if err != nil {
		t.Fatalf("duplicate: %v", err)
	}
	if len(dups) != 2 {
		t.Fatalf("got %d duplicates, want 2", len(dups))
	}
	for _, f := range dups {
		if !fdCloexec(t, f) {
			t.Errorf("duplicate of %s: FD_CLOEXEC clear", f.Name())
		}
		if err := f.Close(); err != nil {
			t.Fatalf("close duplicate: %v", err)
		}
	}
	if !storeLockBusy(t, stateDir) {
		t.Fatal("closing the duplicates released the handle's store lock")
	}

	dups, err = k.DupFiles()
	if err != nil {
		t.Fatalf("duplicate again: %v", err)
	}
	cmd := exec.CommandContext(context.Background(), os.Args[0], "-test.run=^TestLeaseChildHold$", "-test.count=1")
	cmd.Env = append(os.Environ(), leaseChildEnv+"=1")
	cmd.ExtraFiles = dups
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatalf("start child: %v", err)
	}
	t.Cleanup(func() { _ = stdin.Close(); _ = cmd.Wait() })

	for _, f := range dups {
		if err := f.Close(); err != nil {
			t.Fatalf("close duplicate: %v", err)
		}
	}
	if err := k.Release(); err != nil {
		t.Fatalf("release kind lease: %v", err)
	}
	if err := h1.Close(); err != nil {
		t.Fatalf("close first handle: %v", err)
	}

	if !storeLockBusy(t, stateDir) {
		t.Error("the store lock is free with the child alive")
	}
	h2 := writeHandle(t, stateDir)
	if _, err := h2.TakeKindLease("backup"); !errors.Is(err, ErrLeaseBusy) {
		t.Fatalf("kind lease with the child alive: err = %v, want ErrLeaseBusy", err)
	}

	if err := stdin.Close(); err != nil {
		t.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		t.Fatalf("child: %v", err)
	}
	k2, err := h2.TakeKindLease("backup")
	if err != nil {
		t.Fatalf("kind lease after the child exited: %v", err)
	}
	if err := k2.Release(); err != nil {
		t.Fatal(err)
	}
}
