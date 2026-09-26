package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	sqlite "modernc.org/sqlite"
)

// holdersLogInterval is how often a serve-mode lock wait logs the
// lock's holders after its first record, which it logs on the first
// failed poll.
const holdersLogInterval = 5 * time.Second

// holdersLogMsg is the message of every record a serve-mode lock wait
// logs while another process holds the lock.
const holdersLogMsg = "store lock held by another process; waiting"

// sqliteNotADBCode is SQLite's primary result code SQLITE_NOTADB, which
// a file that is not a SQLite database answers on its first read.
const sqliteNotADBCode = 26

// PreRestorePrefix frames the name of a copy taken before a restore.
const PreRestorePrefix = "pre-restore-"

// stateError is a refusal carrying the inspect state it reports as.
type stateError struct {
	state State
	msg   string
}

func (e *stateError) Error() string { return e.msg }

// refusal returns a stateError reporting as s.
func refusal(s State, format string, args ...any) error {
	return &stateError{state: s, msg: fmt.Sprintf(format, args...)}
}

// dbKind is what classify found at the database path.
type dbKind int

const (
	// dbAbsent is no database file, and no copy in snapshots/.
	dbAbsent dbKind = iota + 1

	// dbInterrupted is a first create a crash stopped before its first
	// migration committed: a zero-byte file with no -wal, or a file
	// holding only SQLite's header, and no copy in snapshots/.
	dbInterrupted

	// dbStore is a marked dubplate store whose ledger agrees with the
	// binary at every version the binary carries.
	dbStore
)

// dbClass is classify's verdict. version is the store's user_version
// when kind is dbStore.
type dbClass struct {
	kind    dbKind
	version int
}

// leafState is what l.stat found.
type leafState struct {
	exists    bool
	complete  bool
	hasLock   bool
	hasLeases bool

	// intent is the leaf's restore-intent record, nil when it has none.
	intent *intentRecord
}

// ServeOptions carries serve mode's thresholds and its reconciler, each
// the caller's: the store reads no environment.
type ServeOptions struct {
	// BusyTimeout bounds how long a statement waits on another
	// connection's lock inside SQLite. It must be at least one
	// millisecond.
	BusyTimeout time.Duration

	// MigrationWait bounds each wait for the store lock, shared or
	// exclusive.
	MigrationWait time.Duration

	// Keep is how many pre-migration copies a migration keeps when
	// Backup is off. It must be at least 1.
	Keep int

	// Backup reports whether remote backup is on. The zero value is
	// off, and a migration then prunes pre-migration copies to Keep;
	// with it on, a migration prunes none.
	Backup bool

	// Logger receives the lock holders during a wait, and a warning for
	// a FUSE filesystem. It must not be nil.
	Logger *slog.Logger

	// Reconcile finishes a restore that left its store below the
	// binary's version. Serve calls it once the store is migrated, on
	// a handle at the binary's version, and writes the report it
	// returns into the leaf before removing the restore-intent record.
	// A nil Reconcile refuses such a store; an error refuses it too,
	// leaving the record for the next start. h is valid only for the
	// call.
	Reconcile func(ctx context.Context, h *Handle) (report []byte, err error)
}

// OpenWrite opens the store under stateDir in write mode, holding its
// lock shared until the handle closes. It never creates a store, the
// leaf, lock file, and leases/ included, never migrates one, and never
// removes a staging file: a live backup may be staging under its own
// shared lock. It refuses an absent store naming "start serve first",
// a leaf serve has not finished creating, a store at any version but
// the binary's, and a leaf holding a restore-intent record, which it
// reads again once it holds the lock, since a restore may write one
// while this open waits. ctx bounds the wait for the lock.
//
// The handle takes no lease of its own; a caller that runs jobs takes
// one with [Handle.TakeInstanceLease].
func OpenWrite(ctx context.Context, stateDir string, busyTimeout time.Duration) (*Handle, error) {
	l := newLeaf(stateDir)
	files, err := readMigrations(migrationSource())
	if err != nil {
		return nil, err
	}
	ls, err := l.stat()
	if err != nil {
		return nil, err
	}
	if ls.intent != nil {
		return nil, ls.intent.refusal(l)
	}
	if !ls.exists {
		return nil, noStore(l)
	}
	if _, err := l.probe(); err != nil {
		return nil, err
	}
	if !ls.hasLock {
		c, err := l.classifyLockless()
		if err != nil {
			return nil, err
		}
		return nil, writeRefusal(l, c, len(files))
	}

	lock, err := acquireShared(ctx, l.lockPath)
	if err != nil {
		return nil, fmt.Errorf("open store for writing: %w", err)
	}
	// A restore may have written its record while this open waited.
	if err := l.checkIntent(); err != nil {
		return nil, errors.Join(err, lock.release())
	}
	c, err := classify(ctx, l, busyTimeout, files)
	if err == nil {
		err = writeRefusal(l, c, len(files))
	}
	if err == nil && !ls.hasLeases {
		err = refusal(StateNotAStore, "leaf %s has no %s: start serve to finish creating it", l.root, l.leasesDir)
	}
	if err != nil {
		return nil, errors.Join(err, lock.release())
	}
	h, err := openHandle(l.dbPath, busyTimeout)
	if err != nil {
		return nil, errors.Join(err, lock.release())
	}
	h.version = len(files)
	h.lock = lock
	h.leaf = l
	return h, nil
}

// writeRefusal returns write mode's refusal of c, or nil for a store at
// the binary's version head.
func writeRefusal(l leaf, c dbClass, head int) error {
	switch {
	case c.kind == dbAbsent:
		return noStore(l)
	case c.kind == dbInterrupted:
		return interruptedCreate(l)
	case c.version > head:
		return newerStore(l, c.version, head)
	case c.version < head:
		return refusal(StateOlder, "store %s at version %d, binary at version %d: start serve to migrate", l.dbPath, c.version, head)
	}
	return nil
}

// OpenServe opens the store under stateDir in serve mode, the one mode
// that creates or migrates a store, and returns a handle at the
// binary's version holding the lock shared until it closes. It refuses
// opts without a Logger before it touches any file.
//
// Serve accepts one restore-intent record, reconcile-pending, and
// refuses every other. Beside that record, under the exclusive lock,
// it migrates the store, calls opts.Reconcile, writes the report into
// the leaf as restore-report-<stamp>.txt, and removes the record
// before it opens the store.
//
// flock(2) does not convert a lock atomically: "the existing lock is
// first removed, and then a new lock is established", and another
// process's pending request may be granted between the two. So serve
// holds no connection across a conversion, and after taking the lock
// exclusively it reads the store afresh, since a restore may have
// replaced it in the gap. Each wait for the lock, shared or exclusive,
// is bounded by opts.MigrationWait and logs the holders.
func OpenServe(ctx context.Context, stateDir string, opts ServeOptions) (*Handle, error) {
	if opts.Logger == nil {
		return nil, errors.New("open store for serving: nil logger")
	}
	l := newLeaf(stateDir)
	files, err := readMigrations(migrationSource())
	if err != nil {
		return nil, err
	}
	ls, err := l.stat()
	if err != nil {
		return nil, err
	}
	if ls.intent != nil && ls.intent.State != intentReconcilePending {
		return nil, ls.intent.refusal(l)
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

	for {
		lock, err := waitLock(ctx, l.lockPath, syscall.LOCK_SH, opts)
		if err != nil {
			return nil, err
		}
		h, err := openShared(ctx, l, files, opts)
		if err == nil && h != nil {
			h.lock = lock
			return h, nil
		}
		if err := errors.Join(err, lock.release()); err != nil {
			return nil, err
		}

		lock, err = waitLock(ctx, l.lockPath, syscall.LOCK_EX, opts)
		if err != nil {
			return nil, err
		}
		if err := errors.Join(settle(ctx, l, files, opts), lock.release()); err != nil {
			return nil, err
		}
	}
}

// openShared is serve mode's read under the shared lock. It returns a
// handle, holding no lock of its own, on a store at the binary's
// version, and nil for every store that needs settle's work under the
// exclusive lock. A restore-intent record always needs it, one a
// restore wrote while serve waited for this lock included.
func openShared(ctx context.Context, l leaf, files []migration, opts ServeOptions) (*Handle, error) {
	recorded, err := pathExists(l.restoreIntentPath)
	if err != nil || recorded {
		return nil, err
	}
	c, err := classify(ctx, l, opts.BusyTimeout, files)
	if err != nil {
		return nil, err
	}
	head := len(files)
	switch {
	case c.kind != dbStore || c.version < head:
		return nil, nil
	case c.version > head:
		return nil, newerStore(l, c.version, head)
	}
	h, err := openHandle(l.dbPath, opts.BusyTimeout)
	if err != nil {
		return nil, err
	}
	h.version = head
	h.leaf = l
	return h, nil
}

// settle is serve mode's work under the exclusive lock. It re-reads
// the leaf, refusing every restore-intent record but a
// reconcile-pending one, removes staging leftovers, and then creates,
// completes an interrupted create of, or migrates the store. Beside a
// reconcile-pending record it then finishes the restore. It closes
// every connection it opens.
func settle(ctx context.Context, l leaf, files []migration, opts ServeOptions) (err error) {
	rec, err := l.readIntent()
	if err != nil {
		return err
	}
	if rec != nil && rec.State != intentReconcilePending {
		return rec.refusal(l)
	}
	if rec != nil && opts.Reconcile == nil {
		return refusal(StateRestoreInProgress, "restore-intent record %s (state %s): serve has no reconciler to finish the restore", l.restoreIntentPath, rec.State)
	}
	c, err := classify(ctx, l, opts.BusyTimeout, files)
	if err != nil {
		return err
	}
	head := len(files)
	if c.kind == dbStore && c.version > head {
		return newerStore(l, c.version, head)
	}
	if rec != nil && c.kind != dbStore {
		return refusal(StateRefused, "restore-intent record %s (state %s) names a restored store, but %s holds none", l.restoreIntentPath, rec.State, l.dbPath)
	}
	if err := clearStaging(l.stagingDir); err != nil {
		return err
	}
	switch {
	case c.kind == dbStore && c.version == head && rec == nil:
		return nil
	case c.kind == dbAbsent:
		if err := preCreate(l.dbPath); err != nil {
			return err
		}
	}

	h, err := openHandle(l.dbPath, opts.BusyTimeout)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, h.Close()) }()
	if err := migrate(ctx, h, l, files, migrateOptions{busyTimeout: opts.BusyTimeout, keep: opts.Keep, backup: opts.Backup}); err != nil {
		return err
	}
	if rec == nil {
		return nil
	}
	h.version = head
	h.leaf = l
	return finishRestore(ctx, l, h, rec.Stamp, opts.Reconcile)
}

// finishRestore runs reconcile on h, the migrated store a
// reconcile-pending record left, writes the report it returns into the
// leaf, and only then removes the record, so a failure at any step
// leaves the record for the next start to retry.
func finishRestore(ctx context.Context, l leaf, h *Handle, stamp string, reconcile func(context.Context, *Handle) ([]byte, error)) error {
	report, err := reconcile(ctx, h)
	if err != nil {
		return fmt.Errorf("reconcile the restored store %s: %w", l.dbPath, err)
	}
	if err := writeReport(l, stamp, report); err != nil {
		return fmt.Errorf("write the reconciliation report: %w", err)
	}
	return removeIntent(l)
}

// waitLock polls for path's lock in mode how (syscall.LOCK_SH or
// syscall.LOCK_EX) until it wins or opts.MigrationWait passes. While
// another process holds the lock, it logs the holders on the first
// failed poll and every holdersLogInterval after. On expiry its error
// names the holders.
func waitLock(ctx context.Context, path string, how int, opts ServeOptions) (*storeLock, error) {
	ctx, cancel := context.WithTimeout(ctx, opts.MigrationWait)
	defer cancel()
	want := "shared"
	if how == syscall.LOCK_EX {
		want = "exclusive"
	}
	logTick := time.NewTicker(holdersLogInterval)
	defer logTick.Stop()

	logged := false
	for {
		lock, err := tryLock(path, how)
		if err == nil {
			return lock, nil
		}
		if !errors.Is(err, ErrLockBusy) {
			return nil, err
		}
		if !logged {
			opts.Logger.Warn(holdersLogMsg, "lock", path, "want", want, "holders", lockHolders(path))
			logTick.Reset(holdersLogInterval)
			logged = true
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("wait %s for the store lock %s: %w", opts.MigrationWait, want, err)
		case <-logTick.C:
			opts.Logger.Warn(holdersLogMsg, "lock", path, "want", want, "holders", lockHolders(path))
		case <-time.After(lockPollInterval):
		}
	}
}

// stat runs the steps every mode takes before the filesystem
// probe: statLeaf's, and then a read of the restore-intent record,
// refusing one it cannot read. Each mode decides what a record it read
// means.
func (l leaf) stat() (leafState, error) {
	ls, err := l.statLeaf()
	if err != nil || !ls.exists {
		return ls, err
	}
	ls.intent, err = l.readIntent()
	return ls, err
}

// statLeaf stats l's leaf and its five names, and checks a present
// leaf's shape and its owner against geteuid(). An absent leaf reports
// exists false and meets no owner check, since there is nothing yet to
// own.
func (l leaf) statLeaf() (leafState, error) {
	info, err := os.Stat(l.root)
	if errors.Is(err, fs.ErrNotExist) {
		return leafState{}, nil
	}
	if err != nil {
		return leafState{}, refusal(StateRefused, "stat %s: %v", l.root, err)
	}
	if err := checkLeafShape(info, l.root); err != nil {
		return leafState{}, refusal(StateRefused, "%v", err)
	}
	if err := checkOwner(info); err != nil {
		return leafState{}, refusal(StateRefused, "leaf %s: %v", l.root, err)
	}

	ls := leafState{exists: true, complete: true}
	for _, p := range []string{l.lockPath, l.dbDir, l.snapshotsDir, l.stagingDir, l.leasesDir} {
		ok, err := pathExists(p)
		if err != nil {
			return leafState{}, err
		}
		ls.complete = ls.complete && ok
		switch p {
		case l.lockPath:
			ls.hasLock = ok
		case l.leasesDir:
			ls.hasLeases = ok
		}
	}
	return ls, nil
}

// probe refuses when l sits on a filesystem that spans hosts, and
// returns a warning for FUSE, which may.
func (l leaf) probe() (warning string, err error) {
	target := l.probeTarget()
	verdict, magic, err := probeFilesystem(target)
	if err != nil {
		return "", refusal(StateRefused, "probe filesystem: %v", err)
	}
	switch verdict {
	case fsRefuseMultiHost:
		return "", refusal(StateRefused, "%s is on a filesystem of type %#x that spans hosts; WAL needs every process on one host", target, magic)
	case fsWarnFUSE:
		return fmt.Sprintf("%s is on FUSE (type %#x), which may span hosts; WAL needs every process on one host", target, magic), nil
	}
	return "", nil
}

// classify decides what l's database path holds. It decides absence
// and a zero-byte file by os.Stat alone, and reads everything else
// through the inspect DSN, which opens read-only: a writing DSN sets
// WAL on open, which would rewrite a foreign rollback-journal file's
// header before it could be refused. A database that is absent or an
// interrupted create beside a copy in snapshots/ is refused, since the
// store it replaces may have been lost rather than never made.
func classify(ctx context.Context, l leaf, busyTimeout time.Duration, files []migration) (dbClass, error) {
	info, err := os.Stat(l.dbPath)
	if errors.Is(err, fs.ErrNotExist) {
		return l.besideCopy(dbAbsent)
	}
	if err != nil {
		return dbClass{}, refusal(StateRefused, "stat %s: %v", l.dbPath, err)
	}
	if info.Size() > 0 {
		return classifyOpen(ctx, l, busyTimeout, files)
	}

	wal := l.dbPath + "-wal"
	ok, err := pathExists(wal)
	if err != nil {
		return dbClass{}, err
	}
	if ok {
		return dbClass{}, refusal(StateNotAStore, "%s is zero bytes beside %s: not a store; no mode changes either file", l.dbPath, wal)
	}
	return l.besideCopy(dbInterrupted)
}

// classifyLockless is classify for a leaf with no store.lock, which
// inspect and write mode never create. It opens nothing it cannot
// lock, so a database file there is not yet a store.
func (l leaf) classifyLockless() (dbClass, error) {
	ok, err := pathExists(l.dbPath)
	if err != nil {
		return dbClass{}, err
	}
	if ok {
		return dbClass{}, refusal(StateNotAStore, "%s has no lock file %s: start serve to finish creating it", l.dbPath, l.lockPath)
	}
	return l.besideCopy(dbAbsent)
}

// classifyOpen is classify for a database file that is not empty.
func classifyOpen(ctx context.Context, l leaf, busyTimeout time.Duration, files []migration) (c dbClass, err error) {
	db, err := sql.Open(sqliteDriverName, inspectDSN(l.dbPath, busyTimeout))
	if err != nil {
		return dbClass{}, refusal(StateRefused, "open %s: %v", l.dbPath, err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil && err == nil {
			err = refusal(StateRefused, "close %s: %v", l.dbPath, cerr)
		}
	}()

	var marker, version, objects int64
	err = db.QueryRowContext(ctx,
		"SELECT application_id, user_version, (SELECT count(*) FROM sqlite_schema) FROM pragma_application_id, pragma_user_version",
	).Scan(&marker, &version, &objects)
	if err != nil {
		var se *sqlite.Error
		if errors.As(err, &se) && se.Code()&0xff == sqliteNotADBCode {
			return dbClass{}, refusal(StateNotAStore, "%s is not a database", l.dbPath)
		}
		return dbClass{}, refusal(StateRefused, "read %s: %v", l.dbPath, err)
	}

	switch {
	case marker == 0 && version == 0 && objects == 0:
		return l.besideCopy(dbInterrupted)
	case marker == 0:
		return dbClass{}, refusal(StateNotAStore, "%s is a SQLite database never marked as a dubplate store", l.dbPath)
	case marker != storeMarker:
		return dbClass{}, refusal(StateNotAStore, "%s is a SQLite database with application_id %#x, not a dubplate store", l.dbPath, marker)
	case version < 0:
		return dbClass{}, refusal(StateRefused, "%s: user_version %d is below any version a migration sets", l.dbPath, version)
	}

	// A binary cannot judge ledger rows above its own highest version,
	// which a newer binary leaves behind as a matter of course.
	if version > int64(len(files)) {
		return dbClass{kind: dbStore, version: int(version)}, nil
	}
	if err := checkLedger(ctx, db, files, int(version)); err != nil {
		return dbClass{}, refusal(StateRefused, "%s: %v", l.dbPath, err)
	}
	return dbClass{kind: dbStore, version: int(version)}, nil
}

// besideCopy returns kind for a database that is absent or an
// interrupted create, unless snapshots/ holds a copy or a backup
// reference, in which case it refuses naming the newest and the
// restore that brings its store back.
func (l leaf) besideCopy(kind dbKind) (dbClass, error) {
	name, err := newestCopy(l.snapshotsDir)
	if err != nil {
		return dbClass{}, err
	}
	if name == "" {
		return dbClass{kind: kind}, nil
	}
	p := filepath.Join(l.snapshotsDir, name)
	if stamp, ok := parseBackupRefName(name); ok {
		return dbClass{}, refusal(StateRefused, "%s holds no store, but the backup reference %s exists: restore the backup with restore --snapshot %s", l.dbPath, p, stamp)
	}
	return dbClass{}, refusal(StateRefused, "%s holds no store, but a copy %s exists: restore it with restore --from-file %s", l.dbPath, p, p)
}

// isCopy reports whether name is a pre-migration or pre-restore copy's.
func isCopy(name string) bool {
	_, migrated := parsePreMigrateName(name)
	return migrated || strings.HasPrefix(name, PreRestorePrefix) && strings.HasSuffix(name, PreMigrateSuffix)
}

// newestCopy returns the name of the most recently modified
// pre-migration copy, pre-restore copy, or backup reference in dir, or
// "" when dir holds none or does not exist.
func newestCopy(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	}
	if err != nil {
		return "", refusal(StateRefused, "read %s: %v", dir, err)
	}
	var (
		newest string
		at     time.Time
	)
	for _, e := range entries {
		name := e.Name()
		_, ref := parseBackupRefName(name)
		if !e.Type().IsRegular() || !isCopy(name) && !ref {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return "", refusal(StateRefused, "stat %s: %v", filepath.Join(dir, name), err)
		}
		if newest == "" || info.ModTime().After(at) {
			newest, at = name, info.ModTime()
		}
	}
	return newest, nil
}

// clearStaging removes every entry in dir. Only a process holding the
// lock stages a file, so under serve's exclusive lock any entry there
// is a leftover of a copy a crash interrupted.
func clearStaging(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("clear staging: %w", err)
	}
	for _, e := range entries {
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return fmt.Errorf("clear staging: %w", err)
		}
	}
	return nil
}

// pathExists reports whether p exists, refusing on any stat error but
// not-exist.
func pathExists(p string) (bool, error) {
	_, err := os.Stat(p)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, refusal(StateRefused, "stat %s: %v", p, err)
}

// noStore is the refusal of a mode that never creates a store.
func noStore(l leaf) error {
	return refusal(StateAbsent, "no store at %s: start serve first", l.dbPath)
}

// interruptedCreate is how inspect and write mode report a first
// create a crash stopped, which only serve completes.
func interruptedCreate(l leaf) error {
	return refusal(StateNotAStore, "%s is an unfinished create, not a store: start serve to finish creating it", l.dbPath)
}

// newerStore is the refusal of a store above the binary's version.
func newerStore(l leaf, version, head int) error {
	return refusal(StateNewer, "store %s at version %d is newer than this binary's version %d", l.dbPath, version, head)
}
