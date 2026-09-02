// Package lock provides a filesystem-based mutual-exclusion lock built on
// flock(2), used to keep overlapping dubplate import, backup, and beet
// runs from touching the library at the same time.
package lock

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/sys/unix"
)

// ErrHeld indicates that another holder already has the lock.
var ErrHeld = errors.New("lock held")

// importLockName is the shared lock file's name within stateDir. It is
// unexported: every caller goes through [Path], the one lock-path
// definition import, beet, and backup all consume, so a fix applied to
// one of them can never drift onto a different file than the others.
const importLockName = "import.lock"

// Path returns the import lock's path within stateDir: the persistent
// state directory, not a systemd RuntimeDirectory, since the lock itself
// -- not its content -- must outlive a unit restart and a reboot.
func Path(stateDir string) string {
	return filepath.Join(stateDir, importLockName)
}

// pollInterval is how often AcquireBlocking retries a held lock. flock has
// no cancellable blocking wait, so AcquireBlocking polls Acquire instead.
const pollInterval = 100 * time.Millisecond

// Lock is a held flock(2) lock on a file. The zero value is not usable;
// obtain one from [Acquire] or [AcquireBlocking].
type Lock struct {
	file *os.File
}

// Acquire takes an exclusive, non-blocking flock on path, creating the
// file if it does not exist. It returns [ErrHeld] if another holder
// already has the lock.
func Acquire(path string) (*Lock, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open lock file %s: %w", path, err)
	}

	fd := f.Fd()
	if err := unix.Flock(int(fd), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, ErrHeld
		}
		return nil, fmt.Errorf("flock %s: %w", path, err)
	}

	// Go opens every file O_CLOEXEC, so this fd would close across
	// syscall.Exec and silently drop the lock. dubplate beet holds the
	// lock through an exec into the real beet, so clear the flag here
	// rather than leave it to that one call site.
	if _, err := unix.FcntlInt(fd, unix.F_SETFD, 0); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("clear FD_CLOEXEC on %s: %w", path, err)
	}

	return &Lock{file: f}, nil
}

// AcquireBlocking is [Acquire], but waits for a held lock to free instead
// of returning [ErrHeld] immediately. It returns ctx's error if ctx is
// canceled before the lock frees.
func AcquireBlocking(ctx context.Context, path string) (*Lock, error) {
	for {
		l, err := Acquire(path)
		if err == nil {
			return l, nil
		}
		if !errors.Is(err, ErrHeld) {
			return nil, err
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(pollInterval):
		}
	}
}

// Release releases the lock and closes its underlying file. Closing the
// descriptor is sufficient: flock releases the lock when the last open
// file description referencing it closes, so no explicit LOCK_UN is
// needed.
func (l *Lock) Release() error {
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("close lock file %s: %w", l.file.Name(), err)
	}
	return nil
}

// File returns the lock's underlying file, whose descriptor has
// FD_CLOEXEC cleared so it survives a syscall.Exec while the lock is
// held (dubplate beet's guarantee that beet never runs unguarded).
func (l *Lock) File() *os.File {
	return l.file
}

// Held reports whether another holder currently has the exclusive lock
// at path, without ever taking or blocking on it itself: the
// non-destructive probe internal/journey's queued "the library is busy"
// variant needs, since serve must never take the import lock just to
// test it.
//
// It probes with its own non-blocking flock(2) attempt -- a fresh
// LOCK_EX|LOCK_NB on a fresh file descriptor, released immediately on
// success -- rather than fcntl(2) F_GETLK. Since Linux 2.6, flock()
// locks and fcntl() record locks are maintained in entirely independent
// tables with no interaction between them: an F_GETLK probe can never
// see a lock [Acquire] took with flock(), so it would always report
// "not held" even while another process holds it. Probing with the same
// mechanism [Acquire] itself uses is the only way this function's answer
// can ever be true. A path that does not exist yet reports false, not
// an error: nobody can hold a lock on a file that was never created.
func Held(path string) (bool, error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, fmt.Errorf("open lock file %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	fd := int(f.Fd())
	if err := unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB); err != nil {
		if errors.Is(err, unix.EWOULDBLOCK) {
			return true, nil
		}
		return false, fmt.Errorf("flock %s: %w", path, err)
	}
	// Nobody held it: release the probe's own lock immediately so it
	// never lingers as one of its own.
	if err := unix.Flock(fd, unix.LOCK_UN); err != nil {
		return false, fmt.Errorf("release probe flock %s: %w", path, err)
	}
	return false, nil
}
