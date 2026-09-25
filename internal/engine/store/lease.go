package store

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// ErrLeaseBusy is wrapped by the error [Handle.TakeKindLease] returns
// when another open file description holds the kind's lease. A job
// runner matches it with [errors.Is] to answer a second start of the
// same kind with a conflict rather than a failure.
var ErrLeaseBusy = errors.New("lease busy")

// kindLeasePrefix frames a kind lease's file name in leases/.
const kindLeasePrefix = "kind-"

// maxKindLen bounds a kind lease's name so a caller cannot join an
// arbitrarily long string into leases/.
const maxKindLen = 64

// LeaseState is what [Handle.TryLease] found for another instance's
// lease.
type LeaseState int

const (
	// LeaseHeld is a lease file another open file description holds
	// locked: its owner is alive. TryLease removes nothing.
	LeaseHeld LeaseState = iota + 1

	// LeaseFree is a lease file nobody held: its owner is dead, and
	// TryLease has removed the file.
	LeaseFree

	// LeaseAbsent is no lease file at all.
	LeaseAbsent
)

// An InstanceLease is an exclusive flock(2) on a file in leases/ named
// by its instance, held from [Handle.TakeInstanceLease] until
// [InstanceLease.Release]. While it is held, every [Handle.TryLease] of
// its instance, from any process, reads [LeaseHeld]. A process that
// runs jobs takes one and names its instance as the owner of every job
// record it inserts; the store handle takes none.
type InstanceLease struct {
	file     *os.File
	instance string
}

// TakeInstanceLease creates a lease file in leases/ at mode 0600, named
// by a fresh instance of 32 lowercase hex characters from [NewID], and
// locks it exclusively. It never reuses an existing file: a name that
// already exists is refused.
func (h *Handle) TakeInstanceLease() (*InstanceLease, error) {
	return takeInstanceLease(h.leaf.leasesDir, hex.EncodeToString(NewID()))
}

func takeInstanceLease(dir, instance string) (*InstanceLease, error) {
	path := filepath.Join(dir, instance)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, lockFilePerm)
	if err != nil {
		return nil, fmt.Errorf("take instance lease: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		return nil, errors.Join(fmt.Errorf("lock instance lease %s: %w", path, err), f.Close(), os.Remove(path))
	}
	return &InstanceLease{file: f, instance: instance}, nil
}

// Instance returns the 32 lowercase hex characters that name l.
func (l *InstanceLease) Instance() string { return l.instance }

// Release removes l's file while still holding it, then closes it, so
// no clean exit leaves a lease file behind.
func (l *InstanceLease) Release() error {
	path := l.file.Name()
	var rmErr error
	if err := os.Remove(path); err != nil {
		rmErr = fmt.Errorf("remove instance lease: %w", err)
	}
	if err := l.file.Close(); err != nil {
		return errors.Join(rmErr, fmt.Errorf("close instance lease %s: %w", path, err))
	}
	return rmErr
}

// TryLease reports whether instance's lease is held, and removes the
// file of a lease nobody holds. It locks the file before it removes
// it, so it never removes a live owner's lease, and it reads a file
// that vanished after it locked it as [LeaseFree]: another process
// locked and removed it first, which already proved the owner dead.
//
// A caller tries only instances that a job record names as its owner.
// Such a record is inserted only after its owner holds its lease, so
// a lease file created but not yet locked is never tried.
func (h *Handle) TryLease(instance string) (LeaseState, error) {
	if !validInstance(instance) {
		return 0, fmt.Errorf("try lease: instance %q is not 32 lowercase hex characters", instance)
	}
	path := filepath.Join(h.leaf.leasesDir, instance)
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if errors.Is(err, fs.ErrNotExist) {
		return LeaseAbsent, nil
	}
	if err != nil {
		return 0, fmt.Errorf("try lease: %w", err)
	}
	return tryOpenLease(f)
}

// tryOpenLease is TryLease once the lease file is open as f. It closes
// f before it returns.
func tryOpenLease(f *os.File) (LeaseState, error) {
	path := f.Name()
	err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) {
		return LeaseHeld, f.Close()
	}
	if err != nil {
		return 0, errors.Join(fmt.Errorf("lock lease %s: %w", path, err), f.Close())
	}
	if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return 0, errors.Join(fmt.Errorf("remove dead lease: %w", err), f.Close())
	}
	if err := f.Close(); err != nil {
		return 0, fmt.Errorf("close dead lease %s: %w", path, err)
	}
	return LeaseFree, nil
}

// validInstance reports whether s has the shape TakeInstanceLease
// names a lease with, so no instance read from a record can name a
// path outside leases/.
func validInstance(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil && len(s) == 2*idLen && s == strings.ToLower(s)
}

// validKind reports whether s is a shape TakeKindLease names a lease
// with: one to [maxKindLen] lowercase ASCII letters, digits, and
// hyphens, so no kind read from a caller can name a path outside
// leases/.
func validKind(s string) bool {
	if s == "" || len(s) > maxKindLen {
		return false
	}
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '-' {
			continue
		}
		return false
	}
	return true
}

// A KindLease is an exclusive flock(2) on leases/kind-<kind>, which
// admits one job of a kind at a time across every process, held from
// [Handle.TakeKindLease] until [KindLease.Release].
type KindLease struct {
	file *os.File

	// store is the store lock of the handle the lease was taken
	// through, which DupFiles duplicates and never closes.
	store *storeLock
}

// TakeKindLease locks leases/kind-<kind> exclusively without waiting,
// creating the file at mode 0600 if it is absent. The file is never
// removed. When another open file description holds the lease, it
// returns an error wrapping [ErrLeaseBusy] that names the pids
// /proc/locks records as having acquired it, or "held by another
// process" when it names none. An acquirer's pid is not proof of a
// live holder: a child that inherited the lock holds it after the
// acquirer exits, and on some filesystems /proc/locks names no one.
func (h *Handle) TakeKindLease(kind string) (*KindLease, error) {
	if !validKind(kind) {
		return nil, fmt.Errorf("take kind lease: kind %q is not lowercase ASCII letters, digits, and hyphens", kind)
	}
	path := filepath.Join(h.leaf.leasesDir, kindLeasePrefix+kind)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, lockFilePerm)
	if err != nil {
		return nil, fmt.Errorf("take kind lease: %w", err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		_ = f.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil, leaseBusyError(path)
		}
		return nil, fmt.Errorf("lock kind lease %s: %w", path, err)
	}
	return &KindLease{file: f, store: h.lock}, nil
}

// Release closes l's file, which releases the lease unless a duplicate
// from [KindLease.DupFiles] is still open somewhere. The file stays in
// leases/ for the kind's next job.
func (l *KindLease) Release() error {
	if err := l.file.Close(); err != nil {
		return fmt.Errorf("release kind lease %s: %w", l.file.Name(), err)
	}
	return nil
}

// DupFiles returns duplicates of l's descriptor and of its handle's
// store-lock descriptor, in that order, for a child process's
// [os/exec.Cmd.ExtraFiles]. flock(2) ties a lock to an open file
// description, which a duplicate shares, so a child that inherits both
// keeps the kind lease and the store lock held until it exits, even
// after this process closes every descriptor of its own.
//
// Each duplicate is close-on-exec, so no exec but one that names it in
// ExtraFiles inherits it. The caller owns the duplicates and closes
// them; closing them releases neither lock while l and its handle stay
// open.
func (l *KindLease) DupFiles() ([]*os.File, error) {
	kind, err := dupCloexec(l.file)
	if err != nil {
		return nil, err
	}
	store, err := dupCloexec(l.store.file)
	if err != nil {
		return nil, errors.Join(err, kind.Close())
	}
	return []*os.File{kind, store}, nil
}

// leaseBusyError builds the refusal for a kind lease held at path. It
// says "acquired by" because /proc/locks records the process that took
// the lock, which may have exited while a child that inherited it
// holds it still.
func leaseBusyError(path string) error {
	pids := lockHolders(path)
	if len(pids) == 0 {
		return fmt.Errorf("lease %s: held by another process: %w", path, ErrLeaseBusy)
	}
	noun := "pid"
	if len(pids) > 1 {
		noun = "pids"
	}
	return fmt.Errorf("lease %s: acquired by %s %s: %w", path, noun, strings.Join(pids, ", "), ErrLeaseBusy)
}
