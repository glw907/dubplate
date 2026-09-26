package store

import (
	"context"
	"database/sql"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// State is the condition [Inspect] finds a store in. The set is closed:
// every leaf and database file an inspection can meet reports exactly
// one of these.
type State int

const (
	// StateAbsent is a state directory with no store yet: no leaf, or a
	// leaf holding no database file and no copy in snapshots/.
	StateAbsent State = iota + 1

	// StateRestoreInProgress is a leaf holding a restore-intent record.
	// The report's Reason names the record's state and what finishes
	// the restore.
	StateRestoreInProgress

	// StateBusy is a store whose lock another process holds exclusively.
	StateBusy

	// StateNotAStore is a database path holding something other than a
	// dubplate store, an unfinished first create among them.
	StateNotAStore

	// StateRefused is a store the inspection cannot vouch for: a wrong
	// owner, a multi-host filesystem, an unreadable file, a ledger that
	// disagrees with the binary, or a lost database beside a copy.
	StateRefused

	// StateOlder is a store below the binary's version, which serve
	// migrates.
	StateOlder

	// StateNewer is a store above the binary's version, which this
	// binary never opens for writing.
	StateNewer

	// StateCurrent is a store at the binary's version.
	StateCurrent
)

// String returns the state's name as a status line prints it.
func (s State) String() string {
	switch s {
	case StateAbsent:
		return "absent"
	case StateRestoreInProgress:
		return "restore in progress"
	case StateBusy:
		return "busy"
	case StateNotAStore:
		return "not a store"
	case StateRefused:
		return "refused"
	case StateOlder:
		return "older"
	case StateNewer:
		return "newer"
	case StateCurrent:
		return "current"
	default:
		return "unknown"
	}
}

// InspectOptions carries what [Inspect] needs from its caller: the
// store reads no environment.
type InspectOptions struct {
	// StateDir is the directory whose "store" subdirectory holds the
	// store.
	StateDir string

	// BusyTimeout bounds how long a read waits on another process's
	// lock inside SQLite. It must be at least one millisecond.
	BusyTimeout time.Duration
}

// Report is what [Inspect] found.
type Report struct {
	State State

	// BinaryVersion is the highest migration this binary carries.
	BinaryVersion int

	// StoreVersion is the store's user_version, set for StateOlder,
	// StateNewer, and StateCurrent.
	StoreVersion int

	// Pending names each migration file an older store has not applied,
	// in the order serve applies them.
	Pending []string

	// Holders names the pid of each process holding a busy store's
	// lock, where /proc/locks names them; it is empty when it cannot.
	Holders []string

	// Reason says why the store is busy, not a store, refused, or
	// mid-restore, and names the fix.
	Reason string

	// Warning is set when the store is on a FUSE filesystem, which may
	// span hosts. It does not change State.
	Warning string

	// Facts is set for an older or current store whose version holds
	// the jobs table, and nil otherwise: a newer store's schema is not
	// this binary's to read.
	Facts *Facts
}

// jobsVersion is the lowest store version holding the jobs table.
const jobsVersion = 2

// Facts are the upkeep facts [Inspect] reads beside a store's state.
// Every size comes from stat(2) or statfs(2), never from opening the
// database or its -wal.
type Facts struct {
	// LastSuccess maps each job kind that has succeeded to its newest
	// succeeded record.
	LastSuccess map[string]JobSuccess

	// WALSize is the -wal file's size in bytes, 0 when there is none.
	WALSize int64

	// FreeBytes is the space available to an unprivileged user on the
	// leaf's filesystem, valid only when FreeKnown is true: a statfs
	// failure leaves it unknown rather than failing the inspection.
	FreeBytes uint64
	FreeKnown bool

	// Copies counts the pre-migration and pre-restore copies in
	// snapshots/, and CopiesSize totals their bytes. A backup
	// reference is not a copy.
	Copies     int
	CopiesSize int64
}

// A JobSuccess is a job kind's newest succeeded record.
type JobSuccess struct {
	// Finished is when the job finished, in UTC.
	Finished time.Time

	// Progress is the record's final progress text, "" when it has
	// none.
	Progress string
}

// Inspect reports the state of the store under opts.StateDir without
// changing it. It creates nothing, the lock file included, never
// migrates, and reads the database read-only, so the database file's
// bytes and mode are unchanged afterward; SQLite may create the -wal
// and -shm files beside a database that has none. It takes the store's
// lock shared and does not wait for it: a lock held exclusively
// reports StateBusy. It checks the leaf's owner against the package's
// geteuid seam, which a test swaps to meet a leaf owned by another
// user without running as one.
func Inspect(ctx context.Context, opts InspectOptions) Report {
	l := newLeaf(opts.StateDir)
	files, err := readMigrations(migrationSource())
	if err != nil {
		return reportFrom(Report{}, err)
	}
	r := Report{BinaryVersion: len(files)}

	ls, err := l.stat()
	if err != nil {
		return reportFrom(r, err)
	}
	if ls.intent != nil {
		return reportFrom(r, ls.intent.refusal(l))
	}
	if !ls.exists {
		r.State = StateAbsent
		return r
	}
	r.Warning, err = l.probe()
	if err != nil {
		return reportFrom(r, err)
	}

	if !ls.hasLock {
		c, err := l.classifyLockless()
		if err != nil {
			return reportFrom(r, err)
		}
		return reportClass(r, l, c, files)
	}
	lock, err := acquireSharedNonblocking(l.lockPath)
	if errors.Is(err, errLockBusy) {
		r.State = StateBusy
		r.Holders = lockHolders(l.lockPath)
		r.Reason = err.Error()
		return r
	}
	if err != nil {
		return reportFrom(r, err)
	}
	defer func() { _ = lock.release() }()

	c, err := classify(ctx, l, opts.BusyTimeout, files)
	if err != nil {
		return reportFrom(r, err)
	}
	r = reportClass(r, l, c, files)
	if c.kind != dbStore || c.version < jobsVersion || c.version > len(files) {
		return r
	}
	if r.Facts, err = readFacts(ctx, l, opts.BusyTimeout); err != nil {
		return reportFrom(r, err)
	}
	return r
}

// readFacts reads the upkeep facts of the store in l, whose version
// holds the jobs table. It queries through the inspect DSN, which
// opens the database read-only.
func readFacts(ctx context.Context, l leaf, busyTimeout time.Duration) (*Facts, error) {
	f := &Facts{}
	info, err := os.Stat(l.dbPath + "-wal")
	switch {
	case err == nil:
		f.WALSize = info.Size()
	case !errors.Is(err, fs.ErrNotExist):
		return nil, refusal(StateRefused, "stat %s-wal: %v", l.dbPath, err)
	}

	if f.LastSuccess, err = lastSuccesses(ctx, l.dbPath, busyTimeout); err != nil {
		return nil, err
	}

	var st syscall.Statfs_t
	if statfsFunc(l.root, &st) == nil {
		f.FreeBytes, f.FreeKnown = st.Bavail*uint64(st.Frsize), true
	}

	entries, err := os.ReadDir(l.snapshotsDir)
	if err != nil {
		return nil, refusal(StateRefused, "read %s: %v", l.snapshotsDir, err)
	}
	for _, e := range entries {
		if !e.Type().IsRegular() || !isCopy(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, refusal(StateRefused, "stat %s: %v", filepath.Join(l.snapshotsDir, e.Name()), err)
		}
		f.Copies++
		f.CopiesSize += info.Size()
	}
	return f, nil
}

// lastSuccesses returns each job kind's newest succeeded record in the
// store at dbPath, breaking a tie in finish time by id.
func lastSuccesses(ctx context.Context, dbPath string, busyTimeout time.Duration) (_ map[string]JobSuccess, err error) {
	db, err := sql.Open(sqliteDriverName, inspectDSN(dbPath, busyTimeout))
	if err != nil {
		return nil, refusal(StateRefused, "open %s: %v", dbPath, err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil && err == nil {
			err = refusal(StateRefused, "close %s: %v", dbPath, cerr)
		}
	}()

	rows, err := db.QueryContext(ctx, `SELECT kind, finished_at, progress FROM (
	SELECT kind, finished_at, progress,
		row_number() OVER (PARTITION BY kind ORDER BY finished_at DESC, id DESC) AS rank
	FROM jobs WHERE state = 'succeeded' AND finished_at IS NOT NULL
) WHERE rank = 1`)
	if err != nil {
		return nil, refusal(StateRefused, "read jobs in %s: %v", dbPath, err)
	}
	defer func() { _ = rows.Close() }()

	last := make(map[string]JobSuccess)
	for rows.Next() {
		var (
			kind     string
			finished int64
			progress sql.NullString
		)
		if err := rows.Scan(&kind, &finished, &progress); err != nil {
			return nil, refusal(StateRefused, "read jobs in %s: %v", dbPath, err)
		}
		last[kind] = JobSuccess{Finished: time.UnixMicro(finished).UTC(), Progress: progress.String}
	}
	if err := rows.Err(); err != nil {
		return nil, refusal(StateRefused, "read jobs in %s: %v", dbPath, err)
	}
	return last, nil
}

// reportFrom returns r carrying err's state and text: a refusal
// reports the state it names, and any other error reports refused.
func reportFrom(r Report, err error) Report {
	r.State = StateRefused
	if se, ok := errors.AsType[*stateError](err); ok {
		r.State = se.state
	}
	r.Reason = err.Error()
	return r
}

// reportClass returns r carrying what classify found.
func reportClass(r Report, l leaf, c dbClass, files []migration) Report {
	switch c.kind {
	case dbAbsent:
		r.State = StateAbsent
		return r
	case dbInterrupted:
		return reportFrom(r, interruptedCreate(l))
	}
	r.StoreVersion = c.version
	switch head := len(files); {
	case c.version > head:
		r.State = StateNewer
	case c.version < head:
		r.State = StateOlder
		for _, m := range files[c.version:] {
			r.Pending = append(r.Pending, m.name)
		}
	default:
		r.State = StateCurrent
	}
	return r
}
