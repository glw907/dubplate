package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/atomicfile"
)

// PreMigratePrefix and PreMigrateSuffix frame a pre-migration copy's
// name around its version.
const (
	PreMigratePrefix = "pre-migrate-v"
	PreMigrateSuffix = ".db"
)

// preMigrateName is the name of the copy taken before migrating a
// store at version v.
func preMigrateName(v int) string {
	return PreMigratePrefix + strconv.Itoa(v) + PreMigrateSuffix
}

// takeSnapshot writes a verified copy of the store in l, at version
// want, to l's snapshots directory as name. It stages the copy in l's
// staging directory, which must not already hold name: VACUUM INTO
// writes only into an absent or empty file
// (sqlite.org/lang_vacuum.html), and the stage is pre-created at 0600
// so the copy is never wider. A copy of name already in snapshots is
// replaced by the rename.
//
// The copy runs on a maintenance connection: VACUUM fails on a
// connection with an open transaction, and on the write pool it would
// hold the one connection every in-process writer waits for. VACUUM
// INTO gives the copy the source connection's pager flags, so its
// synchronous=FULL makes the copy durable without a sync of its own;
// the snapshots directory is synced by hand after the rename. The
// stage is removed on any failure through the rename. The copy ends
// with a truncating WAL checkpoint on the store.
func takeSnapshot(ctx context.Context, l leaf, busyTimeout time.Duration, name string, want int) (err error) {
	final := filepath.Join(l.snapshotsDir, name)
	staged := filepath.Join(l.stagingDir, name)
	defer func() {
		if err != nil {
			err = fmt.Errorf("snapshot %s: %w", final, err)
		}
	}()

	db, err := openMaintenanceConn(l.dbPath, busyTimeout)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	if err := checkFreeSpace(ctx, db, l.stagingDir); err != nil {
		return err
	}
	if err := preCreate(staged); err != nil {
		return err
	}
	if err := copyInto(ctx, db, staged, busyTimeout, want); err != nil {
		if rmErr := os.Remove(staged); rmErr != nil {
			err = errors.Join(err, fmt.Errorf("remove %s: %w", staged, rmErr))
		}
		return err
	}
	if err := os.Rename(staged, final); err != nil {
		err = fmt.Errorf("rename %s into %s: %w", staged, l.snapshotsDir, err)
		if rmErr := os.Remove(staged); rmErr != nil {
			err = errors.Join(err, fmt.Errorf("remove %s: %w", staged, rmErr))
		}
		return err
	}
	if err := syncDir(l.snapshotsDir); err != nil {
		return err
	}

	// A checkpoint that a reader blocks reports busy rather than
	// failing; it only bounds the WAL, so busy is not a refusal.
	var busy, walFrames, checkpointed int
	if err := db.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &walFrames, &checkpointed); err != nil {
		return fmt.Errorf("checkpoint: %w", classifyErr(err))
	}
	return nil
}

// copyInto writes db's store into the empty file staged and verifies
// the copy.
func copyInto(ctx context.Context, db *sql.DB, staged string, busyTimeout time.Duration, want int) error {
	if _, err := db.ExecContext(ctx, "VACUUM INTO ?", staged); err != nil {
		return fmt.Errorf("copy into %s: %w", staged, classifyWriteErr(staged, err))
	}
	return verifyCopy(ctx, staged, busyTimeout, want)
}

// versionRule is how a copy's user_version must compare with the
// version a verifier wants.
type versionRule int

const (
	// versionExact accepts only the wanted version: a copy the store
	// takes of itself.
	versionExact versionRule = iota + 1

	// versionAtMost accepts the wanted version or any below it: a
	// snapshot a restore brings in, which serve migrates.
	versionAtMost
)

// verifyCopy refuses the copy at path unless PRAGMA integrity_check
// answers the single row "ok", user_version reads want, and
// application_id reads the store marker. It opens the copy through the
// inspect DSN: the copy is a rollback-journal database, and a writing
// DSN would convert it to WAL and leave -wal and -shm files beside it.
func verifyCopy(ctx context.Context, path string, busyTimeout time.Duration, want int) (err error) {
	db, err := sql.Open(sqliteDriverName, inspectDSN(path, busyTimeout))
	if err != nil {
		return fmt.Errorf("verify %s: %w", path, err)
	}
	defer func() { err = errors.Join(err, db.Close()) }()
	_, err = verifyOpen(ctx, db, path, want, versionExact)
	return err
}

// verifyOpen is the check a verifier runs on db, open on the copy at
// path, and returns the copy's user_version. It refuses unless PRAGMA
// integrity_check answers the single row "ok", user_version meets want
// under rule, and application_id reads the store marker. An empty file
// passes the integrity check and reads version 0, so the marker is
// what refuses it.
func verifyOpen(ctx context.Context, db *sql.DB, path string, want int, rule versionRule) (int, error) {
	rows, err := db.QueryContext(ctx, "PRAGMA integrity_check")
	if err != nil {
		return 0, fmt.Errorf("verify %s: integrity check: %w", path, err)
	}
	var report []string
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			_ = rows.Close()
			return 0, fmt.Errorf("verify %s: integrity check: %w", path, err)
		}
		report = append(report, line)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return 0, fmt.Errorf("verify %s: integrity check: %w", path, err)
	}
	if !slices.Equal(report, []string{"ok"}) {
		return 0, fmt.Errorf("verify %s: integrity check: %s", path, strings.Join(report, "; "))
	}

	var version, marker int
	if err := db.QueryRowContext(ctx, "SELECT user_version, application_id FROM pragma_user_version, pragma_application_id").Scan(&version, &marker); err != nil {
		return 0, fmt.Errorf("verify %s: read version and marker: %w", path, err)
	}
	switch {
	case rule == versionExact && version != want:
		return 0, fmt.Errorf("verify %s: copy at version %d, want %d", path, version, want)
	case rule == versionAtMost && version > want:
		return 0, fmt.Errorf("verify %s: copy at version %d is newer than this binary's version %d", path, version, want)
	case version < 0:
		return 0, fmt.Errorf("verify %s: user_version %d is below any version a migration sets", path, version)
	}
	if marker != storeMarker {
		return 0, fmt.Errorf("verify %s: application_id %#x, want the store marker %#x", path, marker, storeMarker)
	}
	return version, nil
}

// checkFreeSpace refuses unless the filesystem holding dir has room for
// one full copy of db's store plus a WAL at its post-checkpoint bound,
// counting the space available to an unprivileged user.
func checkFreeSpace(ctx context.Context, db *sql.DB, dir string) error {
	var pages, pageSize uint64
	if err := db.QueryRowContext(ctx, "SELECT page_count, page_size FROM pragma_page_count, pragma_page_size").Scan(&pages, &pageSize); err != nil {
		return fmt.Errorf("read store size: %w", classifyErr(err))
	}
	return requireFree(dir, pages*pageSize+journalSizeLimit)
}

// requireFree refuses unless the filesystem holding dir has need bytes
// available to an unprivileged user.
func requireFree(dir string, need uint64) error {
	var st syscall.Statfs_t
	if err := statfsFunc(dir, &st); err != nil {
		return fmt.Errorf("statfs %s: %w", dir, err)
	}
	free := st.Bavail * uint64(st.Frsize)
	if free < need {
		return fmt.Errorf("free space on %s: %d bytes free, need %d", dir, free, need)
	}
	return nil
}

// syncDir syncs the directory dir, making a rename into it durable.
func syncDir(dir string) error {
	f, err := os.Open(dir)
	if err != nil {
		return fmt.Errorf("sync %s: %w", dir, err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("sync %s: %w", dir, err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("sync %s: close: %w", dir, err)
	}
	return nil
}

// pruneSnapshots deletes every pre-migration copy in dir but the keep
// with the highest versions. It considers only names of the form
// pre-migrate-v<N>.db, exactly as preMigrateName writes them, and
// leaves every other file alone.
func pruneSnapshots(dir string, keep int) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("prune %s: %w", dir, err)
	}
	var versions []int
	for _, e := range entries {
		if v, ok := parsePreMigrateName(e.Name()); ok && e.Type().IsRegular() {
			versions = append(versions, v)
		}
	}
	if len(versions) <= keep {
		return nil
	}
	slices.Sort(versions)
	for _, v := range versions[:len(versions)-keep] {
		p := filepath.Join(dir, preMigrateName(v))
		if err := os.Remove(p); err != nil {
			return fmt.Errorf("prune %s: %w", p, err)
		}
	}
	return nil
}

// parsePreMigrateName reports the version name carries when name is
// exactly what preMigrateName writes for it.
func parsePreMigrateName(name string) (int, bool) {
	digits, ok := strings.CutPrefix(name, PreMigratePrefix)
	if !ok {
		return 0, false
	}
	digits, ok = strings.CutSuffix(digits, PreMigrateSuffix)
	if !ok {
		return 0, false
	}
	v, err := strconv.Atoi(digits)
	if err != nil || v < 1 || preMigrateName(v) != name {
		return 0, false
	}
	return v, true
}

// backupPrefix begins the name of a backup copy's stage in staging/
// and of the backup reference in snapshots/; backupRefSuffix ends the
// reference's.
const (
	backupPrefix    = "backup-"
	backupRefSuffix = ".ref"
)

// BackupCopy writes a verified copy of the store to
// staging/backup-<instance>.db and returns its path. The caller sends
// the copy and then removes it; the store never keeps one. instance is
// the caller's instance lease's, which names no other live stage.
//
// It refuses before it creates anything when the filesystem lacks room
// for one copy plus a WAL at its post-checkpoint bound, or when the
// stage already exists. The copy runs on a maintenance connection by
// VACUUM INTO, which reads the last committed state without taking the
// write lock, so a writer in this or any other process keeps
// committing across it. A copy that fails verification is removed.
//
// The copy ends with a truncating WAL checkpoint that does not wait: a
// checkpoint that waits for a reader blocks every new writer until it
// gives up. A checkpoint a reader blocks is logged to logger and does
// not fail the copy.
func (h *Handle) BackupCopy(ctx context.Context, logger *slog.Logger, instance string) (string, error) {
	if !validInstance(instance) {
		return "", fmt.Errorf("backup copy: instance %q is not 32 lowercase hex characters", instance)
	}
	staged := filepath.Join(h.leaf.stagingDir, backupPrefix+instance+PreMigrateSuffix)
	if err := h.backupCopy(ctx, logger, staged); err != nil {
		return "", fmt.Errorf("backup copy %s: %w", staged, err)
	}
	return staged, nil
}

// backupCopy is BackupCopy once staged is named. It removes staged on
// any failure after it created it, the maintenance connection's close
// included.
func (h *Handle) backupCopy(ctx context.Context, logger *slog.Logger, staged string) (err error) {
	created := false
	defer func() {
		if err == nil || !created {
			return
		}
		if rmErr := os.Remove(staged); rmErr != nil {
			err = errors.Join(err, fmt.Errorf("remove %s: %w", staged, rmErr))
		}
	}()

	db, err := openMaintenanceConn(h.path, h.busyTimeout)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, db.Close()) }()

	if err := checkFreeSpace(ctx, db, h.leaf.stagingDir); err != nil {
		return err
	}
	if err := preCreate(staged); err != nil {
		return err
	}
	created = true
	if err := copyInto(ctx, db, staged, h.busyTimeout, h.version); err != nil {
		return err
	}
	return checkpointNoWait(ctx, db, logger, h.path)
}

// checkpointNoWait runs a truncating WAL checkpoint on one of db's
// connections with that connection's busy timeout at 0.
// sqlite.org/c3ref/wal_checkpoint_v2.html: a FULL or stronger
// checkpoint "blocks new database writers while it is pending" while
// it waits on readers through the busy handler, so with the store's own timeout every writer would
// stall for up to that long behind a single held reader. With 0, a
// blocked checkpoint answers busy at once, and that answer is logged.
func checkpointNoWait(ctx context.Context, db *sql.DB, logger *slog.Logger, path string) (err error) {
	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("checkpoint: connect: %w", classifyErr(err))
	}
	defer func() { err = errors.Join(err, conn.Close()) }()

	if _, err := conn.ExecContext(ctx, "PRAGMA busy_timeout = 0"); err != nil {
		return fmt.Errorf("checkpoint: busy timeout off: %w", classifyErr(err))
	}
	var busy, walFrames, checkpointed int
	if err := conn.QueryRowContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)").Scan(&busy, &walFrames, &checkpointed); err != nil {
		return fmt.Errorf("checkpoint: %w", classifyErr(err))
	}
	if busy != 0 {
		logger.Warn("store checkpoint blocked by a reader; the WAL stays until a later checkpoint",
			"store", path, "wal_frames", walFrames, "checkpointed", checkpointed)
	}
	return nil
}

// RemoveBackupStages removes every staging/backup-*.db and leaves every
// other name in staging/ alone. Only a caller holding the backup kind
// lease may call it: the lease proves no backup, a transfer that
// inherited it included, still reads a stage.
func (h *Handle) RemoveBackupStages() error {
	entries, err := os.ReadDir(h.leaf.stagingDir)
	if err != nil {
		return fmt.Errorf("remove backup stages: %w", err)
	}
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, backupPrefix) || !strings.HasSuffix(name, PreMigrateSuffix) {
			continue
		}
		if err := os.Remove(filepath.Join(h.leaf.stagingDir, name)); err != nil {
			return fmt.Errorf("remove backup stage: %w", err)
		}
	}
	return nil
}

// WriteBackupRef records that the backup stamped stamp reached remote,
// as snapshots/backup-<stamp>.ref, and then removes every earlier
// reference, so snapshots/ holds one. It refuses a stamp [ParseStamp]
// refuses, writing nothing. The reference is written durably, its
// directory synced after the rename.
//
// The reference is what tells a lost database from a store never
// made once a backup has deleted every local copy: serve refuses to
// create a store beside one, naming the stamp to restore.
//
// It refuses a remote containing a newline, a carriage return, or a
// NUL, since the reference's body is line-oriented and any of those
// would corrupt it or a later parse of it.
func (h *Handle) WriteBackupRef(stamp, remote string) error {
	return writeBackupRef(h.leaf, stamp, remote)
}

// writeBackupRef is [Handle.WriteBackupRef] on the leaf l, which a
// restore session shares so the two writers cannot drift.
func writeBackupRef(l leaf, stamp, remote string) error {
	if _, err := ParseStamp(stamp); err != nil {
		return fmt.Errorf("write backup reference: %w", err)
	}
	if strings.ContainsAny(remote, "\n\r\x00") {
		return fmt.Errorf("write backup reference: remote %q contains a newline, a carriage return, or a NUL", remote)
	}
	name := backupPrefix + stamp + backupRefSuffix
	path := filepath.Join(l.snapshotsDir, name)
	body := "stamp " + stamp + "\nremote " + remote + "\n"
	err := atomicfile.WriteStream(path, 0o600, func(w io.Writer) error {
		_, err := io.WriteString(w, body)
		return err
	})
	if err != nil {
		return fmt.Errorf("write backup reference: %w", err)
	}

	entries, err := os.ReadDir(l.snapshotsDir)
	if err != nil {
		return fmt.Errorf("remove earlier backup references: %w", err)
	}
	for _, e := range entries {
		if _, ok := parseBackupRefName(e.Name()); !ok || e.Name() == name {
			continue
		}
		if err := os.Remove(filepath.Join(l.snapshotsDir, e.Name())); err != nil {
			return fmt.Errorf("remove earlier backup reference: %w", err)
		}
	}
	return nil
}

// parseBackupRefName reports the stamp name carries when name is a
// backup reference's, backup-<stamp>.ref with a stamp [ParseStamp]
// accepts.
func parseBackupRefName(name string) (string, bool) {
	stamp, ok := strings.CutPrefix(name, backupPrefix)
	if !ok {
		return "", false
	}
	stamp, ok = strings.CutSuffix(stamp, backupRefSuffix)
	if !ok {
		return "", false
	}
	if _, err := ParseStamp(stamp); err != nil {
		return "", false
	}
	return stamp, true
}
