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

	// Address, when set, names the one address whose roster standing
	// [Report.Standing] reports. It is matched exactly against the
	// stored address, which the roster keeps lowercase, so a caller
	// folds case first.
	Address string
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

	// Standing is the roster standing of the address the caller named
	// in [InspectOptions.Address]. It stays [StandingNotRead] when the
	// caller named none and for a store whose version holds no users
	// table.
	Standing Standing
}

// Standing is where one address stands in the roster. It is a report
// field, not a fact: [Facts] carries counts only, and no address
// enters it.
type Standing int

const (
	// StandingNotRead is the zero value: no address was named, or the
	// store holds no users table to look it up in.
	StandingNotRead Standing = iota

	// StandingNone is an address no users row holds.
	StandingNone

	// StandingRemoved is an address whose row is tombstoned, whatever
	// its admin flag.
	StandingRemoved

	// StandingMember is an address held by a live member without the
	// admin flag.
	StandingMember

	// StandingAdmin is an address held by a live admin.
	StandingAdmin
)

// jobsVersion is the lowest store version holding the jobs table.
const jobsVersion = 2

// UsersVersion is the lowest store version holding the users table
// (0003_catalog.sql), and so the lowest at which [Facts] counts members.
const UsersVersion = 3

// settingsVersion is the lowest store version holding the settings
// table (0004_settings.sql).
const settingsVersion = 4

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

	// Settings holds every settings row as stored, in key order, and
	// SettingsGeneration the store's settings generation, for a store
	// whose version holds the settings table. SettingsAbsent marks a
	// store whose version predates it (through 0003), so a table
	// absent is never confused with a table holding no rows; Settings
	// and SettingsGeneration are unset in that case.
	Settings           []SettingsRow
	SettingsGeneration int64
	SettingsAbsent     bool

	// Members counts the live users rows and Admins those holding the
	// admin flag, for a store whose version holds the users table. A
	// store that predates it (through 0002) leaves both zero, which a
	// caller tells from an empty table by the report's StoreVersion.
	Members int
	Admins  int

	// CollectionRoots holds every collection's root_path, relative to
	// the library root, in root order, for a store whose version holds
	// the collections table, which the users table's migration added.
	CollectionRoots [][]byte
}

// A JobSuccess is a job kind's newest succeeded record.
type JobSuccess struct {
	// Finished is when the job finished, in UTC.
	Finished time.Time

	// Progress is the record's final progress text, "" when it has
	// none.
	Progress string
}

// A SettingsRow is one settings row as [Facts] reads it, before
// internal/settings interprets it.
type SettingsRow struct {
	Key       string
	Value     string
	Source    string
	SeedValue string
	Actor     string
	UpdatedAt time.Time
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
		return Report{}.withError(err)
	}
	r := Report{BinaryVersion: len(files)}

	ls, err := l.stat()
	if err != nil {
		return r.withError(err)
	}
	if ls.intent != nil {
		return r.withError(ls.intent.refusal(l))
	}
	if !ls.exists {
		r.State = StateAbsent
		return r
	}
	r.Warning, err = l.probe()
	if err != nil {
		return r.withError(err)
	}

	if !ls.hasLock {
		c, err := l.classifyLockless()
		if err != nil {
			return r.withError(err)
		}
		return r.withClass(l, c, files)
	}
	lock, err := acquireSharedNonblocking(l.lockPath)
	if errors.Is(err, ErrLockBusy) {
		r.State = StateBusy
		r.Holders = lockHolders(l.lockPath)
		r.Reason = err.Error()
		return r
	}
	if err != nil {
		return r.withError(err)
	}
	defer func() { _ = lock.release() }()

	c, err := classify(ctx, l, opts.BusyTimeout, files)
	if err != nil {
		return r.withError(err)
	}
	r = r.withClass(l, c, files)
	if c.kind != dbStore || c.version < jobsVersion || c.version > len(files) {
		return r
	}
	if r.Facts, r.Standing, err = readFacts(ctx, l, opts, c.version); err != nil {
		return r.withError(err)
	}
	return r
}

// readFacts reads the upkeep facts of the store in l, whose version
// holds the jobs table, on one shared read-only connection through the
// inspect DSN: the jobs read and, when version holds the settings
// table, the settings read. When version holds the users table it also
// reads the standing of opts.Address, if one is named.
func readFacts(ctx context.Context, l leaf, opts InspectOptions, version int) (_ *Facts, standing Standing, err error) {
	f := &Facts{}
	info, err := os.Stat(l.dbPath + "-wal")
	switch {
	case err == nil:
		f.WALSize = info.Size()
	case !errors.Is(err, fs.ErrNotExist):
		return nil, StandingNotRead, refusal(StateRefused, "stat %s-wal: %v", l.dbPath, err)
	}

	db, err := sql.Open(sqliteDriverName, inspectDSN(l.dbPath, opts.BusyTimeout))
	if err != nil {
		return nil, StandingNotRead, refusal(StateRefused, "open %s: %v", l.dbPath, err)
	}
	defer func() {
		if cerr := db.Close(); cerr != nil && err == nil {
			err = refusal(StateRefused, "close %s: %v", l.dbPath, cerr)
		}
	}()

	if f.LastSuccess, err = lastSuccesses(ctx, db, l.dbPath); err != nil {
		return nil, StandingNotRead, err
	}

	if version >= settingsVersion {
		if f.Settings, f.SettingsGeneration, err = readSettingsFacts(ctx, db, l.dbPath); err != nil {
			return nil, StandingNotRead, err
		}
	} else {
		f.SettingsAbsent = true
	}

	if version >= UsersVersion {
		if f.Members, f.Admins, err = countMembers(ctx, db, l.dbPath); err != nil {
			return nil, StandingNotRead, err
		}
		if opts.Address != "" {
			if standing, err = readStanding(ctx, db, l.dbPath, opts.Address); err != nil {
				return nil, StandingNotRead, err
			}
		}
		if f.CollectionRoots, err = readCollectionRoots(ctx, db, l.dbPath); err != nil {
			return nil, StandingNotRead, err
		}
	}

	var st syscall.Statfs_t
	if statfsFunc(l.root, &st) == nil {
		f.FreeBytes, f.FreeKnown = st.Bavail*uint64(st.Frsize), true
	}

	entries, err := os.ReadDir(l.snapshotsDir)
	if err != nil {
		return nil, StandingNotRead, refusal(StateRefused, "read %s: %v", l.snapshotsDir, err)
	}
	for _, e := range entries {
		if !e.Type().IsRegular() || !isCopy(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			return nil, StandingNotRead, refusal(StateRefused, "stat %s: %v", filepath.Join(l.snapshotsDir, e.Name()), err)
		}
		f.Copies++
		f.CopiesSize += info.Size()
	}
	return f, standing, nil
}

// lastSuccesses returns each job kind's newest succeeded record in the
// store db reads, breaking a tie in finish time by id.
func lastSuccesses(ctx context.Context, db *sql.DB, dbPath string) (map[string]JobSuccess, error) {
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

// countMembers returns the live users rows db reads and how many of
// them hold the admin flag.
func countMembers(ctx context.Context, db *sql.DB, dbPath string) (members, admins int, err error) {
	err = db.QueryRowContext(ctx, `SELECT count(*), coalesce(sum(admin), 0) FROM users WHERE tombstoned_at IS NULL`).Scan(&members, &admins)
	if err != nil {
		return 0, 0, refusal(StateRefused, "read users in %s: %v", dbPath, err)
	}
	return members, admins, nil
}

// readCollectionRoots returns every collection's root_path db reads, in
// root order.
func readCollectionRoots(ctx context.Context, db *sql.DB, dbPath string) ([][]byte, error) {
	rows, err := db.QueryContext(ctx, `SELECT root_path FROM collections ORDER BY root_path`)
	if err != nil {
		return nil, refusal(StateRefused, "read collections in %s: %v", dbPath, err)
	}
	defer func() { _ = rows.Close() }()
	var roots [][]byte
	for rows.Next() {
		var root []byte
		if err := rows.Scan(&root); err != nil {
			return nil, refusal(StateRefused, "read collections in %s: %v", dbPath, err)
		}
		roots = append(roots, root)
	}
	if err := rows.Err(); err != nil {
		return nil, refusal(StateRefused, "read collections in %s: %v", dbPath, err)
	}
	return roots, nil
}

// readStanding returns where address stands in the users table db
// reads: a tombstoned row is removed whatever its admin flag, and an
// address no row holds is none.
func readStanding(ctx context.Context, db *sql.DB, dbPath, address string) (Standing, error) {
	var admin, tombstoned bool
	err := db.QueryRowContext(ctx, `SELECT admin, tombstoned_at IS NOT NULL FROM users WHERE address = ?`, address).Scan(&admin, &tombstoned)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return StandingNone, nil
	case err != nil:
		return StandingNotRead, refusal(StateRefused, "read users in %s: %v", dbPath, err)
	case tombstoned:
		return StandingRemoved, nil
	case admin:
		return StandingAdmin, nil
	}
	return StandingMember, nil
}

// readSettingsFacts returns every settings row db reads, in key order,
// and the settings generation.
func readSettingsFacts(ctx context.Context, db *sql.DB, dbPath string) ([]SettingsRow, int64, error) {
	rows, err := db.QueryContext(ctx, `SELECT key, value, source, seed_value, actor, updated_at FROM settings ORDER BY key`)
	if err != nil {
		return nil, 0, refusal(StateRefused, "read settings in %s: %v", dbPath, err)
	}
	defer func() { _ = rows.Close() }()

	var settings []SettingsRow
	for rows.Next() {
		var (
			r    SettingsRow
			seed sql.NullString
			at   int64
		)
		if err := rows.Scan(&r.Key, &r.Value, &r.Source, &seed, &r.Actor, &at); err != nil {
			return nil, 0, refusal(StateRefused, "read settings in %s: %v", dbPath, err)
		}
		r.SeedValue = seed.String
		r.UpdatedAt = time.UnixMicro(at).UTC()
		settings = append(settings, r)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, refusal(StateRefused, "read settings in %s: %v", dbPath, err)
	}

	var generation int64
	if err := db.QueryRowContext(ctx, `SELECT generation FROM settings_generation WHERE singleton = 1`).Scan(&generation); err != nil {
		return nil, 0, refusal(StateRefused, "read settings_generation in %s: %v", dbPath, err)
	}
	return settings, generation, nil
}

// withError returns r carrying err's state and text: a refusal
// reports the state it names, and any other error reports refused.
func (r Report) withError(err error) Report {
	r.State = StateRefused
	if se, ok := errors.AsType[*stateError](err); ok {
		r.State = se.state
	}
	r.Reason = err.Error()
	return r
}

// withClass returns r carrying what classify found.
func (r Report) withClass(l leaf, c dbClass, files []migration) Report {
	switch c.kind {
	case dbAbsent:
		r.State = StateAbsent
		return r
	case dbInterrupted:
		return r.withError(interruptedCreate(l))
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
