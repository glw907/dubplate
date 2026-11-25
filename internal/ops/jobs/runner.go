package jobs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
)

// writeWait bounds each store write the runner issues: how long it
// waits for the write pool's one connection in this process. The wait
// on another process's write lock is the store's busy timeout.
const writeWait = 30 * time.Second

// backupDisabled is the reason a start gives while backup integration
// is off.
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
	Retention settings.Reader[time.Duration]

	// Now returns the current time. It stamps every job record and
	// the backup run, and it must not be nil.
	Now func() time.Time
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
	GuardFraction    settings.Reader[float64]
	SnapshotKeep     settings.Reader[int]
	ArchiveRetention settings.Reader[time.Duration]
}

// A Runner runs jobs in this process and answers the start, status,
// and cancel operations. It is safe for concurrent use by multiple
// goroutines.
type Runner struct {
	store     *store.Handle
	lease     *store.InstanceLease
	logger    *slog.Logger
	backup    BackupConfig
	retention settings.Reader[time.Duration]
	now       func() time.Time

	// mu guards jobs and closed. Every wg.Add happens under mu while
	// closed is false, so none can race Close's Wait at zero.
	mu     sync.Mutex
	jobs   map[id.ID]*job
	closed bool
	wg     sync.WaitGroup
}

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
		now:       cfg.Now,
		jobs:      make(map[id.ID]*job),
	}
	if err := r.recoverJobs(ctx); err != nil {
		return nil, errors.Join(fmt.Errorf("open job runner: %w", err), lease.Release())
	}
	return r, nil
}

// Close refuses every later start, waits for every job this runner
// started to end, its terminal write included, then releases the
// instance lease. It cancels no job: a caller that wants its jobs to
// stop cancels them first.
func (r *Runner) Close() error {
	r.mu.Lock()
	r.closed = true
	r.mu.Unlock()
	r.wg.Wait()
	if err := r.lease.Release(); err != nil {
		return fmt.Errorf("close job runner: %w", err)
	}
	return nil
}

// StartBackup starts a backup job owned by this runner, recording
// actor as the one who started it, and returns its record id. It
// answers unavailable while backup integration is off, and conflict
// while another backup holds the kind's lease: naming that job's id
// when its record is live, or a transfer that outlived its process
// when none is.
func (r *Runner) StartBackup(ctx context.Context, actor identity.Actor) (id.ID, error) {
	return r.startBackup(ctx, actor, r.backupBind)
}

// startBackup is [Runner.StartBackup] with the way the job's binder is
// made left to bindFor, which runs once the integration is known to be
// on and before any record is queued. A test passes a bindFor that
// returns a body of its choosing, so a start runs without the three
// backup reads.
func (r *Runner) startBackup(ctx context.Context, actor identity.Actor, bindFor func(context.Context) (binder, error)) (jobID id.ID, err error) {
	defer func() { err = wrapCanceled(ctx, err) }()

	if !r.backup.Enabled {
		return id.ID{}, ops.NewUnavailableError(backupDisabled)
	}
	bind, err := bindFor(ctx)
	if err != nil {
		return id.ID{}, fmt.Errorf("start backup job: %w", err)
	}
	return r.start(ctx, actor, ops.KindBackup, bind)
}

// backupBind returns the binder StartBackup passes to start, from the
// three backup values read now, before start queues a record, so a
// reader error fails the start and leaves no job record rather than
// surfacing only once the bound body runs.
func (r *Runner) backupBind(ctx context.Context) (binder, error) {
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
// it. The job is tracked before its record exists, so recovery never
// finds the record live here without a job running it, and a start
// after [Runner.Close] has begun answers unavailable and writes
// nothing.
func (r *Runner) start(ctx context.Context, actor identity.Actor, kind string, bind binder) (id.ID, error) {
	retention, err := r.retention(ctx)
	if err != nil {
		return id.ID{}, fmt.Errorf("start %s job: %w", kind, err)
	}
	kl, err := r.store.TakeKindLease(kind)
	if errors.Is(err, store.ErrLeaseBusy) {
		return id.ID{}, r.conflict(ctx, kind)
	}
	if err != nil {
		return id.ID{}, fmt.Errorf("start %s job: %w", kind, err)
	}
	if err := r.recoverJobs(ctx); err != nil {
		return id.ID{}, errors.Join(fmt.Errorf("start %s job: %w", kind, err), kl.Release())
	}
	j := newJob(ctx, nil)
	j.body = bind(kl, j.id)
	if err := r.track(j); err != nil {
		j.cancel()
		return id.ID{}, errors.Join(err, kl.Release())
	}
	if err := r.insert(ctx, actor, kind, retention, j); err != nil {
		j.cancel()
		r.untrack(j)
		r.wg.Done()
		return id.ID{}, errors.Join(fmt.Errorf("start %s job: %w", kind, err), kl.Release())
	}

	go func() {
		defer r.wg.Done()
		defer func() {
			r.untrack(j)
			if err := kl.Release(); err != nil {
				r.logger.Error("release kind lease", "job", j.id, "err", err)
			}
		}()
		r.run(j)
	}()
	return j.id, nil
}

// track records j as running under this runner and counts it in
// Close's wait, which the caller ends with r.wg.Done once j's kind
// lease is released. It answers unavailable once Close has begun.
func (r *Runner) track(j *job) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ops.NewUnavailableError("job runner closed")
	}
	r.jobs[j.id] = j
	r.wg.Add(1)
	return nil
}

// untrack forgets j once it has made its last write.
func (r *Runner) untrack(j *job) {
	r.mu.Lock()
	delete(r.jobs, j.id)
	r.mu.Unlock()
}

// tracked reports whether a job of this runner's still runs jobID's
// record.
func (r *Runner) tracked(jobID id.ID) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.jobs[jobID]
	return ok
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
		live  id.ID
		state string
	)
	err := r.store.QueryRow(ctx, `SELECT id, state FROM jobs
WHERE kind = ? AND state IN ('queued', 'running')
ORDER BY at DESC, id DESC LIMIT 1`, kind).Scan(&live, &state)
	if errors.Is(err, sql.ErrNoRows) {
		return ops.NewConflictError("a previous run's transfer is still running")
	}
	if err != nil {
		return fmt.Errorf("start %s job: read live job: %w", kind, err)
	}
	return ops.NewConflictError(fmt.Sprintf("%s job %s is %s", kind, live, state))
}

// JobStatus reads jobID's record. It answers not found for an id no
// record holds.
func (r *Runner) JobStatus(ctx context.Context, _ identity.Actor, jobID id.ID) (st ops.JobStatus, err error) {
	defer func() { err = wrapCanceled(ctx, err) }()
	return r.status(ctx, jobID)
}

// CancelJob asks for the cancellation of jobID's job. A queued job
// ends canceled at once. A running job, whichever process owns it, has
// its cancel_requested set, and its owner ends it canceled unless it
// finishes first; the status returned reports it running. A finished
// job is left unchanged and its state reported.
func (r *Runner) CancelJob(ctx context.Context, _ identity.Actor, jobID id.ID) (st ops.JobStatus, err error) {
	defer func() { err = wrapCanceled(ctx, err) }()

	if err := r.requestCancel(ctx, jobID); err != nil {
		return ops.JobStatus{}, err
	}
	return r.status(ctx, jobID)
}

// wrapCanceled reports err as an error matching [ops.ErrCanceled] when
// ctx's own error is context.Canceled at an operation's return, so a
// caller whose context died sees the canceled sentinel instead
// of whatever error the aborted store call produced. The returned
// error's [*ops.Error] carries an empty Reason: the aborted call's own
// text may name a path, and the wire projection never writes it to the
// answer body. That text stays in the returned error's chain (formatted
// with %v, not wrapped as a second sentinel) so a log line can still
// report it. wrapCanceled leaves err unchanged when ctx is not
// canceled, including when ctx.Err() is context.DeadlineExceeded: the
// operation boundary wraps only a caller's own cancellation, never its
// deadline running out.
func wrapCanceled(ctx context.Context, err error) error {
	if err == nil || ctx.Err() != context.Canceled {
		return err
	}
	return fmt.Errorf("%w: %v", ops.NewCanceledError(""), err)
}

// CancelOwn cancels the context of jobID's job when this runner owns
// it and it has not yet ended, and reports whether it did. It writes
// nothing: it is the path for a caller whose cancel write failed, and
// the job then ends canceled through its own terminal write.
func (r *Runner) CancelOwn(jobID id.ID) bool {
	r.mu.Lock()
	j, ok := r.jobs[jobID]
	r.mu.Unlock()
	if ok {
		j.cancel()
	}
	return ok
}
