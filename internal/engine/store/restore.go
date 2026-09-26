package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

// restoreStagePrefix begins the name of a restore's incoming store in
// staging/, which its stamp and PreMigrateSuffix complete.
const restoreStagePrefix = "restore-"

// RestoreOptions carries what [OpenRestore] needs from its caller: the
// store reads no environment and no clock.
type RestoreOptions struct {
	// BusyTimeout bounds how long a statement waits on another
	// connection's lock inside SQLite. It must be at least one
	// millisecond.
	BusyTimeout time.Duration

	// Now stamps a new restore: its stage, its report, and every copy
	// it moves aside are named by [FormatStamp] of it. A resume takes
	// the stamp its record carries and ignores Now.
	Now time.Time

	// Resume continues the restore the leaf's restore-intent record
	// names. Without it, the open refuses a leaf holding a record;
	// with it, the open refuses a leaf holding none.
	Resume bool

	// FromFile is the local file a restore from a file stages, through
	// [RestoreSession.StageFile]. A new restore refuses it, before
	// clearing staging/, when it lies inside staging/. Empty for a
	// restore that fetches its stage.
	FromFile string

	// Logger receives a warning for a FUSE filesystem. It must not be
	// nil.
	Logger *slog.Logger
}

// A RestoreSession is a restore's hold on a store leaf: the store lock,
// taken exclusively, from [OpenRestore] until [RestoreSession.Close].
// Every operation a restore makes on the leaf's files is a method of
// the session, so each runs under that lock. A session is not safe for
// concurrent use.
type RestoreSession struct {
	leaf        leaf
	lock        *storeLock
	files       []migration
	busyTimeout time.Duration

	// fromFile is the local file [RestoreSession.StageFile] stages.
	fromFile string

	// stamp names the restore's stage, report, and moved-aside copies.
	stamp string

	// stage is the incoming store's path in staging/.
	stage string

	// intent is the record the open found, nil for a new restore.
	intent *intentRecord
}

// OpenRestore opens the store leaf under stateDir for a restore. It
// creates the leaf when it is absent or incomplete, under the rule
// that the effective uid must own stateDir, and refuses a present leaf
// another uid owns. It takes the store lock exclusively without
// waiting, refusing a held lock with an error naming its holders; a
// lost database beside a copy or a backup reference does not refuse,
// since a restore is how that store comes back.
//
// A new restore refuses a leaf holding a restore-intent record, naming
// restore --resume, refuses an opts.FromFile inside staging/, naming
// it, and then removes every entry in staging/: the lock proves no
// other process is staging. A resume refuses a leaf holding
// no record, or one it cannot read, and removes nothing. Both refuse
// before anything is staged when db/, staging/, and snapshots/ do not
// share one device, since the swap renames between them.
func OpenRestore(stateDir string, opts RestoreOptions) (_ *RestoreSession, err error) {
	if opts.Logger == nil {
		return nil, errors.New("open store for restore: nil logger")
	}
	l := newLeaf(stateDir)
	files, err := readMigrations(migrationSource())
	if err != nil {
		return nil, err
	}
	ls, err := l.statLeaf()
	if err != nil {
		return nil, err
	}
	if !ls.complete {
		if err := l.create(stateDir); err != nil {
			return nil, err
		}
	}
	warning, err := l.probe()
	if err != nil {
		return nil, err
	}
	if warning != "" {
		opts.Logger.Warn("store filesystem may span hosts", "detail", warning)
	}

	lock, err := tryLock(l.lockPath, syscall.LOCK_EX)
	if err != nil {
		return nil, fmt.Errorf("open store for restore: %w", err)
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, lock.release())
		}
	}()

	s := &RestoreSession{leaf: l, lock: lock, files: files, busyTimeout: opts.BusyTimeout, fromFile: opts.FromFile}
	if err := s.readRecord(opts); err != nil {
		return nil, err
	}
	if err := l.checkDevices(); err != nil {
		return nil, err
	}
	if s.intent == nil {
		if err := l.checkOutsideStaging(opts.FromFile); err != nil {
			return nil, err
		}
		if err := clearStaging(l.stagingDir); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// readRecord reads the leaf's restore-intent record under the session's
// lock and decides, by opts.Resume, whether the restore may go on. It
// sets the session's stamp and stage from the record on a resume, and
// from opts.Now otherwise.
func (s *RestoreSession) readRecord(opts RestoreOptions) error {
	l := s.leaf
	if !opts.Resume {
		if err := l.checkIntent(); err != nil {
			return err
		}
		s.stamp = FormatStamp(opts.Now)
		s.stage = filepath.Join(l.stagingDir, restoreStagePrefix+s.stamp+PreMigrateSuffix)
		return nil
	}

	b, err := os.ReadFile(l.restoreIntentPath)
	if errors.Is(err, fs.ErrNotExist) {
		return refusal(StateRefused, "no restore-intent record at %s: there is no restore to resume", l.restoreIntentPath)
	}
	var r intentRecord
	if err == nil {
		r, err = decodeIntent(b)
	}
	if err != nil {
		return refusal(StateRestoreInProgress, "restore-intent record %s is unreadable: %v: restore --resume cannot finish it; recover the leaf by hand as the runbook's restore-intent procedure describes", l.restoreIntentPath, err)
	}
	s.intent = &r
	s.stamp = r.Stamp
	s.stage = filepath.Join(l.stagingDir, r.Incoming)
	return nil
}

// checkOutsideStaging refuses src, naming it, when the entry it names
// or the file it resolves to lies inside l's staging/, since a new
// restore clears staging/ before it stages. An empty src, or one whose
// directory is absent, passes: there is nothing in staging/ to lose,
// and [RestoreSession.StageFile] refuses the absent source.
func (l leaf) checkOutsideStaging(src string) error {
	if src == "" {
		return nil
	}
	abs, err := filepath.Abs(src)
	if err != nil {
		return refusal(StateRefused, "restore from %s: %v", src, err)
	}
	dir, err := filepath.EvalSymlinks(filepath.Dir(abs))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return refusal(StateRefused, "restore from %s: %v", src, err)
	}
	staging, err := filepath.EvalSymlinks(l.stagingDir)
	if err != nil {
		return refusal(StateRefused, "resolve %s: %v", l.stagingDir, err)
	}
	paths := []string{filepath.Join(dir, filepath.Base(abs))}
	switch resolved, err := filepath.EvalSymlinks(abs); {
	case err == nil:
		paths = append(paths, resolved)
	case !errors.Is(err, fs.ErrNotExist):
		return refusal(StateRefused, "restore from %s: %v", src, err)
	}
	for _, p := range paths {
		if rel, err := filepath.Rel(staging, p); err == nil && filepath.IsLocal(rel) {
			return refusal(StateRefused, "restore from %s: it is in %s, which a restore clears", src, l.stagingDir)
		}
	}
	return nil
}

// dirDevice is a directory, resolved through any symlink, and the
// device holding it.
type dirDevice struct {
	path string
	dev  uint64
}

// checkDevices refuses unless l's db/, staging/, and snapshots/ sit on
// one device, each resolved through any symlink, since os.Rename fails
// across devices.
func (l leaf) checkDevices() error {
	var dirs []dirDevice
	for _, dir := range []string{l.dbDir, l.stagingDir, l.snapshotsDir} {
		resolved, err := filepath.EvalSymlinks(dir)
		if err != nil {
			return refusal(StateRefused, "resolve %s: %v", dir, err)
		}
		info, err := os.Stat(resolved)
		if err != nil {
			return refusal(StateRefused, "stat %s: %v", resolved, err)
		}
		st, ok := info.Sys().(*syscall.Stat_t)
		if !ok {
			return refusal(StateRefused, "stat %s: no unix stat info", resolved)
		}
		dirs = append(dirs, dirDevice{path: resolved, dev: st.Dev})
	}
	return oneDevice(dirs)
}

// oneDevice refuses, naming every directory and its device, unless
// every entry of dirs has the same device.
func oneDevice(dirs []dirDevice) error {
	for _, d := range dirs[1:] {
		if d.dev == dirs[0].dev {
			continue
		}
		names := make([]string, len(dirs))
		for i, d := range dirs {
			names[i] = fmt.Sprintf("%s (device %d)", d.path, d.dev)
		}
		return refusal(StateRefused, "restore renames between %s, which must share one device", strings.Join(names, ", "))
	}
	return nil
}

// Close releases the session's store lock.
func (s *RestoreSession) Close() error {
	return s.lock.release()
}

// StagePath returns the path the incoming store is staged at, in
// staging/. A fetch writes it there itself; [RestoreSession.StageFile]
// copies the local file [RestoreOptions.FromFile] names there.
func (s *RestoreSession) StagePath() string {
	return s.stage
}

// DupFiles returns a duplicate of the session's store-lock descriptor
// for a child process's [os/exec.Cmd.ExtraFiles], as
// [KindLease.DupFiles] does for a backup: a child that inherits it
// holds the store lock exclusively until it exits, even after the
// session closes. The duplicate is close-on-exec, and the caller owns
// and closes it.
func (s *RestoreSession) DupFiles() ([]*os.File, error) {
	f, err := dupCloexec(s.lock.file)
	if err != nil {
		return nil, err
	}
	return []*os.File{f}, nil
}

// CheckSpace refuses unless staging/'s filesystem has room for a stage
// of size bytes plus a WAL at the store's post-checkpoint bound. The
// swap renames within one device, so the copy of the current store it
// moves aside costs no space.
func (s *RestoreSession) CheckSpace(size int64) error {
	if size < 0 {
		return fmt.Errorf("check free space: size %d is unknown", size)
	}
	return requireFree(s.leaf.stagingDir, uint64(size)+journalSizeLimit)
}

// StageFile copies the local file [RestoreOptions.FromFile] names,
// src here, into the session's stage at mode 0600 and syncs it,
// leaving src untouched. [OpenRestore] has already refused a src
// inside staging/. StageFile refuses, staging nothing, when the session
// has no src; when src is absent, not a regular file, or unreadable;
// when a -wal or -journal file sits beside it, since a byte copy of
// such a database is torn; when src is the live database; and when
// staging/ lacks the space [RestoreSession.CheckSpace] checks.
func (s *RestoreSession) StageFile() error {
	src := s.fromFile
	if src == "" {
		return errors.New("stage: the restore session names no file to restore from")
	}
	resolved, err := s.checkSource(src)
	if err != nil {
		return err
	}
	in, err := os.Open(resolved)
	if err != nil {
		return fmt.Errorf("stage %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()
	info, err := in.Stat()
	if err != nil {
		return fmt.Errorf("stage %s: %w", src, err)
	}
	if err := s.CheckSpace(info.Size()); err != nil {
		return fmt.Errorf("stage %s: %w", src, err)
	}
	if err := preCreate(s.stage); err != nil {
		return fmt.Errorf("stage %s: %w", src, err)
	}
	if err := copyFile(s.stage, in); err != nil {
		return errors.Join(fmt.Errorf("stage %s: %w", src, err), os.Remove(s.stage))
	}
	return nil
}

// checkSource refuses src as StageFile does and returns it resolved
// through any symlink.
func (s *RestoreSession) checkSource(src string) (string, error) {
	resolved, err := filepath.EvalSymlinks(src)
	if err != nil {
		return "", fmt.Errorf("restore from %s: %w", src, err)
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return "", fmt.Errorf("restore from %s: %w", src, err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("restore from %s: not a regular file", src)
	}
	live, err := os.Stat(s.leaf.dbPath)
	switch {
	case err == nil && os.SameFile(info, live):
		return "", fmt.Errorf("restore from %s: it is the live database %s", src, s.leaf.dbPath)
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return "", fmt.Errorf("stat %s: %w", s.leaf.dbPath, err)
	}
	for _, p := range []string{src, resolved} {
		for _, suffix := range []string{"-wal", "-journal"} {
			ok, err := pathExists(p + suffix)
			if err != nil {
				return "", err
			}
			if ok {
				return "", fmt.Errorf("restore from %s: %s exists beside it, so a copy of the file alone may be torn", src, p+suffix)
			}
		}
	}
	return resolved, nil
}

// copyFile copies src into the existing empty file at path and syncs
// it.
func copyFile(path string, src io.Reader) error {
	out, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, src); err != nil {
		return errors.Join(err, out.Close())
	}
	if err := out.Sync(); err != nil {
		return errors.Join(err, out.Close())
	}
	return out.Close()
}

// VerifyStage checks the staged incoming store and returns its
// user_version. It first sets the stage to mode 0600, since a fetch
// renames in a file at the mode its umask gives, and then refuses a
// stage that fails the integrity check, is not a marked dubplate
// store, is newer than this binary, or carries a migration ledger that
// disagrees with this binary's. An older store passes: serve migrates
// it.
//
// Reading the stage leaves -wal and -shm files beside it, which it
// removes; a -wal that is not empty is refused, since the stage alone
// would then lack what it holds. A refused stage is removed with them,
// unless a restore-intent record names it.
func (s *RestoreSession) VerifyStage(ctx context.Context) (int, error) {
	version, err := s.verifyStage(ctx)
	if err == nil || s.intent != nil {
		return version, err
	}
	for _, p := range []string{s.stage, s.stage + "-wal", s.stage + "-shm"} {
		if rmErr := os.Remove(p); rmErr != nil && !errors.Is(rmErr, fs.ErrNotExist) {
			err = errors.Join(err, fmt.Errorf("remove %s: %w", p, rmErr))
		}
	}
	return 0, err
}

// verifyStage is VerifyStage before its removal of a refused stage.
func (s *RestoreSession) verifyStage(ctx context.Context) (int, error) {
	info, err := os.Lstat(s.stage)
	if err != nil {
		return 0, fmt.Errorf("verify the stage: %w", err)
	}
	if !info.Mode().IsRegular() {
		return 0, fmt.Errorf("verify the stage: %s is not a regular file", s.stage)
	}
	if err := os.Chmod(s.stage, 0o600); err != nil {
		return 0, fmt.Errorf("verify the stage: %w", err)
	}
	version, err := s.checkStage(ctx)
	if err != nil {
		return 0, err
	}

	wal := s.stage + "-wal"
	switch info, err := os.Stat(wal); {
	case err == nil && info.Size() > 0:
		return 0, fmt.Errorf("verify %s: %s holds %d bytes the stage alone lacks", s.stage, wal, info.Size())
	case err != nil && !errors.Is(err, fs.ErrNotExist):
		return 0, fmt.Errorf("verify %s: %w", s.stage, err)
	}
	for _, p := range []string{wal, s.stage + "-shm"} {
		if err := os.Remove(p); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return 0, fmt.Errorf("verify %s: %w", s.stage, err)
		}
	}
	return version, nil
}

// checkStage runs the verifier and the ledger check on the stage over
// one inspect-DSN connection, closed before it returns.
func (s *RestoreSession) checkStage(ctx context.Context) (_ int, err error) {
	db, err := sql.Open(sqliteDriverName, inspectDSN(s.stage, s.busyTimeout))
	if err != nil {
		return 0, fmt.Errorf("verify %s: %w", s.stage, err)
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	version, err := verifyOpen(ctx, db, s.stage, len(s.files), versionAtMost)
	if err != nil {
		return 0, err
	}
	if err := checkLedger(ctx, db, s.files, version); err != nil {
		return 0, fmt.Errorf("verify %s: %w", s.stage, err)
	}
	return version, nil
}

// OpenRead opens a handle on the store in db/, which must be at the
// binary's version, for the restore's reconciliation. The handle takes
// no lock of its own: the session's exclusive lock covers it, and a
// second lock from this process would conflict with that one. The
// caller closes the handle before the session.
func (s *RestoreSession) OpenRead(ctx context.Context) (*Handle, error) {
	c, err := classify(ctx, s.leaf, s.busyTimeout, s.files)
	if err != nil {
		return nil, err
	}
	head := len(s.files)
	switch {
	case c.kind != dbStore:
		return nil, refusal(StateRefused, "%s holds no restored store", s.leaf.dbPath)
	case c.version != head:
		return nil, refusal(StateRefused, "restored store %s at version %d, binary at version %d: only serve opens it", s.leaf.dbPath, c.version, head)
	}
	h, err := openHandle(s.leaf.dbPath, s.busyTimeout)
	if err != nil {
		return nil, err
	}
	h.version = head
	h.leaf = s.leaf
	return h, nil
}

// WriteBackupRef records the restored snapshot as the leaf's backup
// reference, exactly as [Handle.WriteBackupRef] records a backup.
func (s *RestoreSession) WriteBackupRef(stamp, remote string) error {
	return writeBackupRef(s.leaf, stamp, remote)
}
