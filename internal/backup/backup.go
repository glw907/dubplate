package backup

import (
	"context"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
)

// FinishedMessage is the [Job.Run] result for a run whose prune
// reported nothing of its own: the clean case, with no skip or
// failure to surface.
const FinishedMessage = "backup finished"

// A Job is one backup run's inputs. Every field is required.
type Job struct {
	// Store is a handle opened in write mode.
	Store *store.Handle

	// Lease is the backup kind lease the caller holds for the run.
	// rclone inherits it.
	Lease *store.KindLease

	// Instance is the name of the caller's instance lease, which names
	// the run's store copy in staging/.
	Instance string

	// RunID names the run in the snapshot's metadata.
	RunID string

	// Remote is the destination, as [ParseRemote] reads it.
	Remote string

	// LibraryRoot is the library tree to sync. Symlinks in it are
	// resolved once, at the run's start.
	LibraryRoot string

	// GuardFraction is the share of the catalog's tracks the library
	// tree must hold as regular files before the sync may run: at
	// least a value in (0, 1], never NaN. The caller's registry (a
	// store-homed key, R-6c11) owns the policy default and its
	// bounds; Run validates only the structural invariant this
	// package's own arithmetic depends on (R-6c13).
	GuardFraction float64

	// SnapshotKeep is how many remote snapshots a prune keeps in
	// total, this run's own included: at least 1.
	SnapshotKeep int

	// ArchiveRetention is how long an archive directory survives past
	// its stamp: above zero.
	ArchiveRetention time.Duration

	// Logger receives what the run cannot report through its result.
	Logger *slog.Logger
}

// validate refuses a structurally invalid GuardFraction, SnapshotKeep,
// or ArchiveRetention before any rclone launch, the LookPath included
// (R-6c13). A zero value left at the caller's own zero value would
// otherwise silently change what the run does rather than fail loudly:
// a guard fraction of 0 turns the short-tree guard off, an archive
// retention of 0 purges every archive directory but this run's own,
// and a keep of 0 makes othersKeep -1 and panics indexing the sorted
// snapshot list. This package never supplies a default of its own; the
// caller's registry owns the policy bounds, and a second copy of its
// default here is exactly what the registry exists to prevent.
func (j Job) validate() error {
	switch {
	case j.SnapshotKeep < 1:
		return fail("job-invalid", "snapshot keep %d must be at least 1", j.SnapshotKeep)
	case math.IsNaN(j.GuardFraction), j.GuardFraction <= 0, j.GuardFraction > 1:
		return fail("job-invalid", "guard fraction %g must be greater than 0 and at most 1", j.GuardFraction)
	case j.ArchiveRetention <= 0:
		return fail("job-invalid", "archive retention %s must be greater than 0", j.ArchiveRetention)
	}
	return nil
}

// Run runs the backup, reporting each phase through progress, and
// stops with progress's error when progress returns one. A failure
// before the sync never starts rclone. A failure's code, which its job
// record carries, is readable through a FailureCode method on an error
// in its chain.
//
// Run removes every staging/backup-*.db before it stages its own copy,
// which only the holder of the backup kind lease may do.
//
// Once the send succeeds, Run writes the backup reference and then
// prunes the remote's snapshots and archive and the local snapshots
// directory; a skipped or failed prune is folded into the returned
// text rather than the run's error, since the backup itself landed. A
// clean prune reports no text of its own, so Run's returned text
// names the run finished instead of repeating the last phase it
// reported.
func (j Job) Run(ctx context.Context, progress func(phase string) error) (string, error) {
	start := time.Now()
	stamp := store.FormatStamp(start)
	if err := progress("checking the library"); err != nil {
		return "", err
	}
	if err := j.validate(); err != nil {
		return "", err
	}
	bin, err := exec.LookPath("rclone")
	if err != nil {
		return "", fail("rclone-missing", "rclone is not on PATH: %v", err)
	}
	remote, err := ParseRemote(j.Remote)
	if err != nil {
		return "", fail("remote-invalid", "%v", err)
	}
	root, err := resolveRoot(j.LibraryRoot)
	if err != nil {
		return "", err
	}
	maxDelete, err := checkLibrary(ctx, j.Store, root, j.GuardFraction)
	if err != nil {
		return "", err
	}

	t := transport{bin: bin, inherit: j.Lease.DupFiles}
	if err := progress("syncing the library"); err != nil {
		return "", err
	}
	err = t.run(ctx, nil, "sync", root, remote.join("library"),
		"--backup-dir", remote.join("archive/"+stamp),
		"--max-delete", strconv.Itoa(maxDelete))
	if err != nil {
		return "", err
	}
	synced := time.Now().UTC().Format(time.RFC3339)

	if err := progress("copying the store"); err != nil {
		return "", err
	}
	if err := j.Store.RemoveBackupStages(); err != nil {
		return "", fail("snapshot-failed", "%v", err)
	}
	stage, err := j.Store.BackupCopy(ctx, j.Logger, j.Instance)
	if err != nil {
		return "", fail("snapshot-failed", "%v", err)
	}
	defer func() {
		if err := os.Remove(stage); err != nil {
			j.Logger.Warn("backup stage not removed; the next backup removes it", "stage", stage, "err", err)
		}
	}()

	if err := progress("sending the store"); err != nil {
		return "", err
	}
	err = t.run(ctx, nil, "copyto", stage, remote.join("snapshots/"+stamp+".db"),
		"-M",
		"--metadata-set", "run_id="+j.RunID,
		"--metadata-set", "sync_completed="+synced)
	if err != nil {
		return "", err
	}

	if err := j.Store.WriteBackupRef(stamp, j.Remote); err != nil {
		return "", fail("reference-failed", "%v", err)
	}

	if err := progress("pruning old snapshots"); err != nil {
		return "", err
	}
	localDir := filepath.Join(filepath.Dir(filepath.Dir(stage)), "snapshots")
	p := prune{t: t, remote: remote, stamp: stamp, localDir: localDir, start: start,
		snapshotKeep: j.SnapshotKeep, archiveRetention: j.ArchiveRetention}
	result := p.run(ctx, j.Logger)
	if result == "" {
		result = FinishedMessage
	}
	return result, nil
}
