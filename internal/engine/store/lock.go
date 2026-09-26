package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"
)

// lockPollInterval is how often a bounded lock wait retries a held
// flock, since flock(2) has no cancellable blocking form.
const lockPollInterval = 50 * time.Millisecond

// ErrLockBusy is the sentinel every refusal of a held store lock
// wraps, so a caller can tell a busy lock from a real I/O error
// without matching error text. The refusal's message names the lock's
// holders where /proc/locks names them.
var ErrLockBusy = errors.New("lock busy")

// storeLock is a held flock(2) lock on store.lock. The zero value is
// not usable; obtain one from [acquireSharedNonblocking],
// [acquireShared], or [acquireExclusive].
type storeLock struct {
	file *os.File
}

// release releases l and closes its underlying file. flock releases
// the lock when the last open file description on it closes, so no
// explicit LOCK_UN call is needed.
func (l *storeLock) release() error {
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("release lock %s: %w", l.file.Name(), err)
	}
	return nil
}

// acquireSharedNonblocking makes one attempt at a shared flock on path
// and returns [ErrLockBusy] immediately if another holder has it
// exclusively. It never creates path: a missing lock file is a state
// inspect and write report, not a file they make.
func acquireSharedNonblocking(path string) (*storeLock, error) {
	return tryLock(path, syscall.LOCK_SH)
}

// acquireShared polls for a shared flock on path at lockPollInterval
// until it succeeds or ctx ends, whichever comes first.
func acquireShared(ctx context.Context, path string) (*storeLock, error) {
	return acquireBounded(ctx, path, syscall.LOCK_SH)
}

// acquireExclusive polls for an exclusive flock on path at
// lockPollInterval until it succeeds or ctx ends, whichever comes
// first.
func acquireExclusive(ctx context.Context, path string) (*storeLock, error) {
	return acquireBounded(ctx, path, syscall.LOCK_EX)
}

// acquireBounded is acquireShared and acquireExclusive's shared poll
// loop. It returns the last attempt's busy error, naming the lock's
// holders, when ctx ends before a lock is won.
func acquireBounded(ctx context.Context, path string, how int) (*storeLock, error) {
	for {
		l, err := tryLock(path, how)
		if err == nil {
			return l, nil
		}
		if !errors.Is(err, ErrLockBusy) {
			return nil, err
		}
		select {
		case <-ctx.Done():
			return nil, err
		case <-time.After(lockPollInterval):
		}
	}
}

// tryLock makes one non-blocking flock attempt on path in mode how
// (syscall.LOCK_SH or syscall.LOCK_EX). path must already exist: the
// lock file is created only by the leaf's create, never by an open
// that merely wants to lock it.
func tryLock(path string, how int) (*storeLock, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("open lock file %s: %w", path, err)
	}
	if err := syscall.Flock(int(f.Fd()), how|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, busyError(path)
		}
		return nil, fmt.Errorf("flock %s: %w", path, err)
	}
	return &storeLock{file: f}, nil
}

// busyError builds the refusal for a held lock at path, naming its
// holders when /proc/locks names them and falling back to "held by
// another process" when it does not: an unreadable /proc/locks, no
// matching device and inode (the mechanics review measured this always
// happens on btrfs, whose superblock device differs from stat's
// per-subvolume one), or a queued waiter with no holder line at all.
func busyError(path string) error {
	pids := lockHolders(path)
	if len(pids) == 0 {
		return fmt.Errorf("lock %s: held by another process: %w", path, ErrLockBusy)
	}
	return fmt.Errorf("lock %s: held by pid %s: %w", path, strings.Join(pids, ", "), ErrLockBusy)
}

// lockHolders returns the pid of every process /proc/locks (proc_locks(5))
// names as a current holder of path's lock, or nil if path cannot be
// stat'd, has no unix stat info, or /proc/locks cannot be read.
func lockHolders(path string) []string {
	info, err := os.Stat(path)
	if err != nil {
		return nil
	}
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return nil
	}
	data, err := os.ReadFile("/proc/locks")
	if err != nil {
		return nil
	}
	major, minor := devMajorMinor(st.Dev)
	return procLockHolders(string(data), major, minor, st.Ino)
}

// devMajorMinor splits a raw dev_t the way glibc's sysmacros.h does,
// matching /proc/locks's own hex-major:hex-minor:inode device field
// (bits/sysmacros.h, read 2026-09-23).
func devMajorMinor(dev uint64) (major, minor uint64) {
	major = ((dev & 0x00000000000fff00) >> 8) | ((dev & 0xfffff00000000000) >> 32)
	minor = (dev & 0xff) | ((dev & 0x00000ffffff00000) >> 12)
	return major, minor
}

// procLockHolders parses /proc/locks text and returns the pid of every
// line whose device and inode match major, minor, and inode. Fields
// are read from the end of the line (pid, then device:inode, then the
// byte range) rather than a fixed front index, because a queued
// waiter's line carries an extra "->" token before its lock type
// (proc_locks(5)); reading from the front would land on that request's
// own pid and device:inode instead, so a waiter is skipped outright
// rather than misread as a holder.
func procLockHolders(text string, major, minor, inode uint64) []string {
	var pids []string
	for line := range strings.SplitSeq(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 || fields[1] == "->" {
			continue
		}
		if len(fields) < 4 {
			continue
		}
		pid := fields[len(fields)-4]
		devInode := fields[len(fields)-3]
		var m, n, ino uint64
		if _, err := fmt.Sscanf(devInode, "%x:%x:%d", &m, &n, &ino); err != nil {
			continue
		}
		if m == major && n == minor && ino == inode {
			pids = append(pids, pid)
		}
	}
	return pids
}

// dupCloexec duplicates f's descriptor with fcntl(F_DUPFD_CLOEXEC). The
// duplicate shares f's open file description, and with it any flock
// on f, and is close-on-exec, which a plain dup(2) would clear.
func dupCloexec(f *os.File) (*os.File, error) {
	rc, err := f.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("duplicate %s: %w", f.Name(), err)
	}
	var (
		fd    uintptr
		errno syscall.Errno
	)
	if err := rc.Control(func(src uintptr) {
		fd, _, errno = syscall.Syscall(syscall.SYS_FCNTL, src, syscall.F_DUPFD_CLOEXEC, 0)
	}); err != nil {
		return nil, fmt.Errorf("duplicate %s: %w", f.Name(), err)
	}
	if errno != 0 {
		return nil, fmt.Errorf("fcntl F_DUPFD_CLOEXEC on %s: %w", f.Name(), errno)
	}
	return os.NewFile(fd, f.Name()), nil
}
