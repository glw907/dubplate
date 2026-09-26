package local

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// writeWait bounds each store write the runner issues: how long it
// waits for the write pool's one connection in this process. The wait
// on another process's write lock is the store's busy timeout.
const writeWait = 30 * time.Second

// backupDisabled is the reason a start and the catalog give while
// backup integration is off.
const backupDisabled = "backup integration disabled"

// Config is what [New] needs from its caller. The runner opens nothing
// itself.
type Config struct {
	// Store is a handle opened in write mode. The runner never closes
	// it; its caller closes it after [Runner.Close].
	Store *store.Handle

	// Backup is the backup integration's configuration.
	Backup BackupConfig

	// Logger receives every transition the runner cannot record in the
	// store. It must not be nil.
	Logger *slog.Logger

	// Retention returns DUBPLATE_JOB_RETENTION's value in force. A
	// start reads it fresh before it queues a record, since a stored
	// value applies at the next operation that reads it; it must not
	// be nil.
	Retention func(ctx context.Context) (time.Duration, error)
}

// BackupConfig configures the backup job.
type BackupConfig struct {
	// Enabled reports whether backup integration is on. While it is
	// off, a backup start answers unavailable and writes nothing.
	Enabled bool

	// Remote is the rclone destination, name:path. A job refuses a
	// malformed one before it reads the library.
	Remote string

	// LibraryRoot is the library tree a backup syncs.
	LibraryRoot string

	// GuardFraction, SnapshotKeep, and ArchiveRetention return the
	// three backup keys' values in force. A backup start reads all
	// three before it queues a record, so a reader error fails the
	// start naming the key and leaves no job record; they must not be
	// nil while Enabled is true.
	GuardFraction    func(ctx context.Context) (float64, error)
	SnapshotKeep     func(ctx context.Context) (int, error)
	ArchiveRetention func(ctx context.Context) (time.Duration, error)
}

// A Runner runs jobs in this process and answers the four operations
// of [ops.Seam]. It is safe for concurrent use by multiple goroutines.
type Runner struct {
	store     *store.Handle
	lease     *store.InstanceLease
	logger    *slog.Logger
	backup    BackupConfig
	retention func(ctx context.Context) (time.Duration, error)

	// bindBackup overrides the backup kind's job body: nil in
	// production, where StartBackup reads the three backup values
	// itself and binds the real [backup.Job]. A test sets it to run a
	// body of its own choosing without those reads.
	bindBackup binder

	mu   sync.Mutex
	jobs map[ops.Handle]*job
	wg   sync.WaitGroup
}

var _ ops.Seam = (*Runner)(nil)

// New takes an instance lease through cfg.Store, then recovers every
// queued or running record whose owner is dead, marking it
// interrupted. It fails, releasing the lease, when a record's owner
// lease cannot be read. ctx bounds the recovery's write.
func New(ctx context.Context, cfg Config) (*Runner, error) {
	if cfg.Logger == nil {
		return nil, errors.New("open job runner: nil logger")
	}
	lease, err := cfg.Store.TakeInstanceLease()
	if err != nil {
		return nil, fmt.Errorf("open job runner: %w", err)
	}
	r := &Runner{
		store:     cfg.Store,
		lease:     lease,
		logger:    cfg.Logger,
		backup:    cfg.Backup,
		retention: cfg.Retention,
		jobs:      make(map[ops.Handle]*job),
	}
	if err := r.recoverJobs(ctx); err != nil {
		return nil, errors.Join(fmt.Errorf("open job runner: %w", err), lease.Release())
	}
	return r, nil
}

// Close waits for every job this runner started to end, its terminal
// write included, then releases the instance lease. It cancels no job:
// a caller that wants its jobs to stop cancels them first.
func (r *Runner) Close() error {
	r.wg.Wait()
	if err := r.lease.Release(); err != nil {
		return fmt.Errorf("close job runner: %w", err)
	}
	return nil
}

// Catalog lists the four operations. A backup start is unavailable
// while backup integration is off; the other three always answer.
// Each entry's privilege comes from [ops.PrivilegeOf], never a literal
// repeated here, so the catalog cannot drift from what an authorizing
// layer enforces.
func (r *Runner) Catalog(ctx context.Context, _ ops.Actor, _ ops.CatalogRequest) (resp ops.CatalogResponse, err error) {
	defer func() { err = wrapCanceled(ctx, err) }()

	entries := []ops.CatalogEntry{
		{Name: ops.OpStartBackup, Available: r.backup.Enabled},
		{Name: ops.OpJobStatus, Available: true},
		{Name: ops.OpCancelJob, Available: true},
		{Name: ops.OpCatalog, Available: true},
	}
	if !r.backup.Enabled {
		entries[0].Reason = backupDisabled
	}
	for i, e := range entries {
		priv, ok := ops.PrivilegeOf(e.Name)
		if !ok {
			return ops.CatalogResponse{}, fmt.Errorf("catalog: operation %q has no privilege", e.Name)
		}
		entries[i].Privilege = priv
	}
	cat, err := ops.BuildCatalog(entries...)
	if err != nil {
		return ops.CatalogResponse{}, err
	}
	return ops.CatalogResponse{Operations: cat.Entries()}, nil
}

// StartBackup starts a backup job owned by this runner, recording
// actor as the one who started it, and returns its handle. It answers
// unavailable while backup integration is off, and conflict while
// another backup holds the kind's lease: naming that job's handle when
// its record is live, or a transfer that outlived its process when
// none is.
func (r *Runner) StartBackup(ctx context.Context, actor ops.Actor, _ ops.StartBackupRequest) (resp ops.StartBackupResponse, err error) {
	defer func() { err = wrapCanceled(ctx, err) }()

	if !r.backup.Enabled {
		return ops.StartBackupResponse{}, ops.NewUnavailableError(backupDisabled)
	}
	bind, err := r.backupBind(ctx)
	if err != nil {
		return ops.StartBackupResponse{}, fmt.Errorf("start backup job: %w", err)
	}
	h, err := r.start(ctx, actor, ops.KindBackup, bind)
	if err != nil {
		return ops.StartBackupResponse{}, err
	}
	return ops.StartBackupResponse{JobID: h}, nil
}

// backupBind returns the binder StartBackup passes to start: a test's
// own override when one is set, or, in production, the three backup
// values read now, before start queues a record, so a reader error
// fails the start and leaves no job record rather than surfacing only
// once the bound body runs.
func (r *Runner) backupBind(ctx context.Context) (binder, error) {
	if r.bindBackup != nil {
		return r.bindBackup, nil
	}
	fraction, err := r.backup.GuardFraction(ctx)
	if err != nil {
		return nil, err
	}
	keep, err := r.backup.SnapshotKeep(ctx)
	if err != nil {
		return nil, err
	}
	archiveRetention, err := r.backup.ArchiveRetention(ctx)
	if err != nil {
		return nil, err
	}
	return r.backupJob(fraction, keep, archiveRetention), nil
}

// start reads the retention value in force, takes kind's lease,
// recovers, inserts a queued record, and runs the body bind makes on a
// goroutine that releases the lease when it ends. The retention read
// happens before insert, fresh on every start rather than once at
// [New], since a stored value applies at the next operation that reads
// it.
func (r *Runner) start(ctx context.Context, actor ops.Actor, kind string, bind binder) (ops.Handle, error) {
	retention, err := r.retention(ctx)
	if err != nil {
		return "", fmt.Errorf("start %s job: %w", kind, err)
	}
	kl, err := r.store.TakeKindLease(kind)
	if errors.Is(err, store.ErrLeaseBusy) {
		return "", r.conflict(ctx, kind)
	}
	if err != nil {
		return "", fmt.Errorf("start %s job: %w", kind, err)
	}
	if err := r.recoverJobs(ctx); err != nil {
		return "", errors.Join(fmt.Errorf("start %s job: %w", kind, err), kl.Release())
	}
	j, err := r.insert(ctx, actor, kind, retention, nil)
	if err != nil {
		return "", errors.Join(fmt.Errorf("start %s job: %w", kind, err), kl.Release())
	}
	j.body = bind(kl, j.handle)

	r.mu.Lock()
	r.jobs[j.handle] = j
	r.mu.Unlock()
	r.wg.Go(func() {
		defer func() {
			r.mu.Lock()
			delete(r.jobs, j.handle)
			r.mu.Unlock()
			if err := kl.Release(); err != nil {
				r.logger.Error("release kind lease", "job", j.handle, "err", err)
			}
		}()
		r.run(j)
	})
	return j.handle, nil
}

// conflict answers a start whose kind lease is busy. It recovers first,
// so a dead owner's record is interrupted rather than named.
func (r *Runner) conflict(ctx context.Context, kind string) error {
	ctx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	if err := r.recoverJobs(ctx); err != nil {
		return fmt.Errorf("start %s job: %w", kind, err)
	}
	var (
		id    []byte
		state string
	)
	err := r.store.QueryRow(ctx, `SELECT id, state FROM jobs
WHERE kind = ? AND state IN ('queued', 'running')
ORDER BY at DESC, id DESC LIMIT 1`, kind).Scan(&id, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return ops.NewConflictError("a previous run's transfer is still running")
	}
	if err != nil {
		return fmt.Errorf("start %s job: read live job: %w", kind, err)
	}
	return ops.NewConflictError(fmt.Sprintf("%s job %s is %s", kind, handleOf(id), state))
}

// JobStatus reads the record req names.
func (r *Runner) JobStatus(ctx context.Context, _ ops.Actor, req ops.JobStatusRequest) (resp ops.JobStatusResponse, err error) {
	defer func() { err = wrapCanceled(ctx, err) }()

	id, err := parseID(req.JobID)
	if err != nil {
		return ops.JobStatusResponse{}, err
	}
	st, err := r.status(ctx, id)
	if err != nil {
		return ops.JobStatusResponse{}, err
	}
	return ops.JobStatusResponse{Job: st}, nil
}

// CancelJob asks for the cancellation of the job req names. A queued
// job ends canceled at once. A running job, whichever process owns it,
// has its cancel_requested set, and its owner ends it canceled unless
// it finishes first; the response reports it running. A finished job
// is left unchanged and its state reported.
func (r *Runner) CancelJob(ctx context.Context, _ ops.Actor, req ops.CancelJobRequest) (resp ops.CancelJobResponse, err error) {
	defer func() { err = wrapCanceled(ctx, err) }()

	id, err := parseID(req.JobID)
	if err != nil {
		return ops.CancelJobResponse{}, err
	}
	if err := r.requestCancel(ctx, id); err != nil {
		return ops.CancelJobResponse{}, err
	}
	st, err := r.status(ctx, id)
	if err != nil {
		return ops.CancelJobResponse{}, err
	}
	return ops.CancelJobResponse{Job: st}, nil
}

// wrapCanceled reports err as an error matching [ops.ErrCanceled] when
// ctx's own error is context.Canceled at a seam method's return, so a
// caller whose context died sees the seam's canceled sentinel instead
// of whatever error the aborted store call produced. The returned
// error's [*ops.Error] carries an empty Reason: the aborted call's own
// text may name a path, and the wire projection never writes it to the
// answer body. That text stays in the returned error's chain (formatted
// with %v, not wrapped as a second sentinel) so a log line can still
// report it. wrapCanceled leaves err unchanged when ctx is not
// canceled, including when ctx.Err() is context.DeadlineExceeded: the
// seam boundary wraps only a caller's own cancellation, never its
// deadline running out.
func wrapCanceled(ctx context.Context, err error) error {
	if err == nil || ctx.Err() != context.Canceled {
		return err
	}
	return fmt.Errorf("%w: %v", ops.NewCanceledError(""), err)
}

// CancelOwn cancels the context of the job h names when this runner
// owns it and it has not yet ended, and reports whether it did. It
// writes nothing: it is the path for a caller whose cancel write
// failed, and the job then ends canceled through its own terminal
// write.
func (r *Runner) CancelOwn(h ops.Handle) bool {
	r.mu.Lock()
	j, ok := r.jobs[h]
	r.mu.Unlock()
	if ok {
		j.cancel()
	}
	return ok
}
