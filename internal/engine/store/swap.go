package store

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// discardedPrefix begins the name of a current store a restore moved
// aside unopened, in snapshots/. No prune matches it, and the check for
// a lost database never names it as a store to restore.
const discardedPrefix = "discarded-"

// sidecars are every file SQLite may keep beside a database, in either
// journal mode.
var sidecars = []string{"-wal", "-shm", "-journal"}

// inUseHint closes every refusal of a current store another process
// may still hold.
const inUseHint = "another process may have the database open"

// discardHint closes every refusal a restore with --discard-current
// would pass.
const discardHint = "restore --discard-current moves the current files aside"

// currentFound is what a restore found at db/ before it wrote its
// record.
type currentFound int

const (
	// currentUnsettled is a session that has not quiesced the current
	// store.
	currentUnsettled currentFound = iota

	// currentAbsent is a db/ holding no database and no sidecar.
	currentAbsent

	// currentQuiesced is a current store left in rollback-journal mode
	// with no sidecar, ready to be renamed whole.
	currentQuiesced

	// currentDiscard is a current store to be moved aside unopened.
	currentDiscard
)

// quiesce readies the store at path to be renamed as one file. On a
// maintenance connection, which leaves the journal mode as the file
// has it, it checkpoints the WAL into the database and truncates it,
// then leaves WAL mode, whose last close deletes the -wal and -shm. It
// refuses a file that is not a marked store, naming --discard-current,
// and refuses naming another process when the checkpoint reports a
// reader, the switch answers SQLITE_BUSY, or the switch answers any
// mode but delete. It closes the connection before it returns.
func quiesce(ctx context.Context, path string, busyTimeout time.Duration) (err error) {
	db, err := openMaintenanceConn(path, busyTimeout)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	var marker, version int64
	err = db.QueryRowContext(ctx, "SELECT application_id, user_version FROM pragma_application_id, pragma_user_version").Scan(&marker, &version)
	if err != nil {
		return fmt.Errorf("current store %s cannot be opened: %v: %s", path, err, discardHint)
	}
	if marker != storeMarker {
		return fmt.Errorf("current store %s is not a dubplate store (application_id %#x): %s", path, marker, discardHint)
	}

	// The busy column names a reader holding an old snapshot before
	// the switch below fails less specifically on the same reader.
	var busy, walFrames, checkpointed int
	if err := db.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &walFrames, &checkpointed); err != nil {
		return fmt.Errorf("quiesce %s: checkpoint: %w", path, classifyErr(err))
	}
	if busy != 0 {
		return fmt.Errorf("quiesce %s: a reader kept the checkpoint from finishing: %s", path, inUseHint)
	}
	var mode string
	if err := db.QueryRowContext(ctx, "PRAGMA journal_mode=DELETE").Scan(&mode); err != nil {
		err = classifyErr(err)
		if errors.Is(err, errBusy) {
			return fmt.Errorf("quiesce %s: leave WAL mode: %v: %s", path, err, inUseHint)
		}
		return fmt.Errorf("quiesce %s: leave WAL mode: %w", path, err)
	}
	if mode != "delete" {
		return fmt.Errorf("quiesce %s: journal mode %s after the switch to delete: %s", path, mode, inUseHint)
	}
	return nil
}

// presentSidecars names each of suffixes that exists beside path.
func presentSidecars(path string, suffixes []string) ([]string, error) {
	var present []string
	for _, suffix := range suffixes {
		ok, err := pathExists(path + suffix)
		if err != nil {
			return nil, err
		}
		if ok {
			present = append(present, path+suffix)
		}
	}
	return present, nil
}

// discardedPath is where a restore stamped stamp moves an unopenable
// current database, in l's snapshots/; its sidecars go beside it.
func (l leaf) discardedPath(stamp string) string {
	return filepath.Join(l.snapshotsDir, discardedPrefix+stamp+PreMigrateSuffix)
}

// A move renames one leaf file.
type move struct {
	from, to string
}

// manualRecovery closes a refusal restore --resume cannot get past.
const manualRecovery = "recover the leaf by hand as the runbook's restore-intent procedure describes"

// swapPlan is what remains of the swap r records, decided from which
// of the files it names exist: the moves that set the current store
// aside, and whether the incoming store in staging/ is still to be
// renamed into db/. A swap that has finished plans nothing. It refuses
// every combination a swap cannot have left, naming each file and
// whether it exists, and refuses before the incoming store would land
// beside a sidecar the plan does not move.
func (l leaf) swapPlan(r *intentRecord) (moves []move, in bool, err error) {
	incoming := filepath.Join(l.stagingDir, r.Incoming)
	var aside []move
	switch {
	case r.DiscardCurrent:
		discarded := l.discardedPath(r.Stamp)
		for _, suffix := range []string{"", "-wal", "-shm"} {
			aside = append(aside, move{from: l.dbPath + suffix, to: discarded + suffix})
		}
	case r.PreRestore != "":
		aside = []move{{from: l.dbPath, to: filepath.Join(l.snapshotsDir, r.PreRestore)}}
	}
	names := []string{incoming}
	if aside == nil {
		names = append(names, l.dbPath)
	}
	for _, m := range aside {
		names = append(names, m.from, m.to)
	}
	exists := make(map[string]bool, len(names))
	for _, p := range names {
		if exists[p], err = pathExists(p); err != nil {
			return nil, false, err
		}
	}

	staged, current := exists[incoming], exists[l.dbPath]
	ok := false
	switch {
	case r.DiscardCurrent:
		moves, in, ok = discardPlan(aside, exists, staged)
	case r.PreRestore != "":
		m := aside[0]
		switch copied := exists[m.to]; {
		case staged && current && !copied:
			moves, in, ok = aside, true, true
			// The quiesce left the current store without sidecars, so
			// one here was written after the record.
			written, err := presentSidecars(l.dbPath, sidecars)
			if err != nil {
				return nil, false, err
			}
			if len(written) > 0 {
				return nil, false, fmt.Errorf("restore-intent record %s (state swap): found %s beside the current store: a process wrote to it after the record: %s", l.restoreIntentPath, strings.Join(written, ", "), manualRecovery)
			}
		case staged && !current && copied:
			in, ok = true, true
		case !staged && current && copied:
			ok = true
		}
	default:
		in, ok = staged && !current, staged != current
	}
	if !ok {
		return nil, false, l.unknownSwap(names, exists)
	}
	if !in {
		return moves, false, nil
	}

	// SQLite would apply a sidecar of the store the incoming one
	// replaces to the incoming store.
	var stale []string
	for _, suffix := range sidecars {
		p := l.dbPath + suffix
		if slices.ContainsFunc(moves, func(m move) bool { return m.from == p }) {
			continue
		}
		ok, err := pathExists(p)
		if err != nil {
			return nil, false, err
		}
		if ok {
			stale = append(stale, p)
		}
	}
	if len(stale) > 0 {
		return nil, false, fmt.Errorf("restore-intent record %s (state swap): the incoming store would land beside %s: %s", l.restoreIntentPath, strings.Join(stale, ", "), manualRecovery)
	}
	return moves, true, nil
}

// discardPlan is swapPlan for a record owing a discard, whose aside
// moves are the database's, then its -wal's, then its -shm's. While the
// incoming store waits in staging/, each current file still in db/ is
// to be moved, and no file may be in both places; once the incoming
// store is in db/, the database must already have moved.
func discardPlan(aside []move, exists map[string]bool, staged bool) (moves []move, in, ok bool) {
	if !staged {
		return nil, false, exists[aside[0].from] && exists[aside[0].to]
	}
	for _, m := range aside {
		if exists[m.from] && exists[m.to] {
			return nil, false, false
		}
		if exists[m.from] {
			moves = append(moves, m)
		}
	}
	return moves, true, true
}

// unknownSwap refuses a swap whose files are in no state a swap
// leaves, naming each file and whether it exists.
func (l leaf) unknownSwap(names []string, exists map[string]bool) error {
	found := make([]string, len(names))
	for i, p := range names {
		found[i] = p + " absent"
		if exists[p] {
			found[i] = p + " present"
		}
	}
	return fmt.Errorf("restore-intent record %s (state swap) names files in no state a swap leaves (%s): %s", l.restoreIntentPath, strings.Join(found, ", "), manualRecovery)
}
