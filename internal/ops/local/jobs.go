package local

import (
	"context"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
)

// cancelPoll is how often a running job's owner re-reads
// cancel_requested between the body's progress reports.
const cancelPoll = 2 * time.Second

// retention is how long a finished record is kept, the newest
// succeeded record of each kind excepted.
const retention = 90 * 24 * time.Hour

// jobErrorCode is the failure code of a body error that carries none.
const jobErrorCode = "job-error"

// A body is a job's work. It reports each phase through progress,
// which records the phase and returns the context's error once a
// cancel has been requested. A body that succeeds may return a final
// progress text for its record; one that fails returns an error, whose
// failure code the runner reads through a [coded] in its chain.
type body func(ctx context.Context, progress func(phase string) error) (final string, err error)

// A binder makes a job's body once the runner holds its kind lease
// and has named its handle.
type binder func(kl *store.KindLease, h ops.Handle) body

// backupJob binds the backup kind's body to the job h names, run under
// kl.
func (r *Runner) backupJob(kl *store.KindLease, h ops.Handle) body {
	return backup.Job{
		Store:       r.store,
		Lease:       kl,
		Instance:    r.lease.Instance(),
		RunID:       string(h),
		Remote:      r.backup.Remote,
		LibraryRoot: r.backup.LibraryRoot,
		Logger:      r.logger,
	}.Run
}

// coded is implemented by a body error that names its own failure
// code. Any other error is recorded as [jobErrorCode].
type coded interface {
	error
	FailureCode() string
}

// A job is one record this runner owns while its body runs.
type job struct {
	id     []byte
	handle ops.Handle
	body   body
	ctx    context.Context
	cancel context.CancelFunc
}

// now returns the current time as the store records one: integer UTC
// microseconds.
func now() int64 { return time.Now().UnixMicro() }

// storeContext returns a context for one of a job's own store calls,
// which must work after the job's context is done: its terminal write
// above all.
func storeContext(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), writeWait)
}

// insert records a queued job of kind owned by this runner's instance
// lease, then deletes every finished record past retention but the
// newest succeeded record of each kind, in one write.
func (r *Runner) insert(ctx context.Context, actor ops.Actor, kind string, b body) (*job, error) {
	ctx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	id := store.NewID()
	at := now()
	err := r.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if _, err := tx.Exec(ctx, `INSERT INTO jobs (id, kind, at, state, owner_instance, actor, updated_at)
VALUES (?, ?, ?, 'queued', ?, ?, ?)`, id, kind, at, r.lease.Instance(), actor.String(), at); err != nil {
			return fmt.Errorf("insert job: %w", err)
		}
		cutoff := at - retention.Microseconds()
		if _, err := tx.Exec(ctx, `DELETE FROM jobs
WHERE state IN ('succeeded', 'failed', 'canceled', 'interrupted')
	AND finished_at < ?
	AND id NOT IN (SELECT id FROM (
		SELECT id, row_number() OVER (PARTITION BY kind ORDER BY finished_at DESC, id DESC) AS rank
		FROM jobs WHERE state = 'succeeded' AND finished_at IS NOT NULL
	) WHERE rank = 1)`, cutoff); err != nil {
			return fmt.Errorf("prune jobs: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	jctx, jcancel := context.WithCancel(context.WithoutCancel(ctx))
	return &job{id: id, handle: handleOf(id), body: b, ctx: jctx, cancel: jcancel}, nil
}

// run moves j from queued to running, runs its body, and records how
// it ended. A record no longer queued, canceled before it ran, ends j
// without running the body.
func (r *Runner) run(j *job) {
	defer j.cancel()
	ran, err := r.markRunning(j)
	if err != nil {
		r.logger.Error("job could not start", "job", j.handle, "err", err)
		return
	}
	if !ran {
		return
	}

	var poller sync.WaitGroup
	stop := make(chan struct{})
	poller.Go(func() { r.pollCancel(j, stop) })
	final, bodyErr := j.body(j.ctx, func(phase string) error { return r.progress(j, phase) })
	close(stop)
	poller.Wait()

	if err := r.finish(j, outcomeOf(j.ctx, final, bodyErr)); err != nil {
		r.logger.Error("job end not recorded", "job", j.handle, "err", err)
	}
}

// markRunning moves j's record from queued to running and reports
// whether it did. A record in any other state is left alone and
// logged: a cancel landed before the job ran.
func (r *Runner) markRunning(j *job) (bool, error) {
	ctx, cancel := storeContext(j.ctx)
	defer cancel()
	var (
		ran   bool
		state string
	)
	err := r.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		t := now()
		res, err := tx.Exec(ctx, `UPDATE jobs SET state = 'running', started_at = ?, updated_at = ?
WHERE id = ? AND state = 'queued'`, t, t, j.id)
		if err != nil {
			return fmt.Errorf("mark job running: %w", err)
		}
		if ran, err = affected(res); err != nil || ran {
			return err
		}
		return tx.QueryRow(ctx, "SELECT state FROM jobs WHERE id = ?", j.id).Scan(&state)
	})
	if err != nil {
		return false, err
	}
	if !ran {
		r.logger.Info("job ended before it ran", "job", j.handle, "state", state)
	}
	return ran, nil
}

// An outcome is how a job's body ended, as its terminal write records
// it.
type outcome struct {
	state    ops.JobState
	progress string
	code     string
	message  string
}

// outcomeOf decides a body's outcome. An error after the job's
// context is done is a cancel, whatever the error; a success is a
// success, whatever the context.
func outcomeOf(ctx context.Context, final string, err error) outcome {
	switch {
	case err == nil:
		return outcome{state: ops.JobSucceeded, progress: final}
	case ctx.Err() != nil:
		return outcome{state: ops.JobCanceled}
	}
	code := jobErrorCode
	if c, ok := errors.AsType[coded](err); ok {
		code = c.FailureCode()
	}
	return outcome{state: ops.JobFailed, code: code, message: err.Error()}
}

// finish moves j's record from running to o's terminal state. A record
// no longer running, finished by recovery or a queued cancel, is left
// alone and the refusal logged.
func (r *Runner) finish(j *job, o outcome) error {
	ctx, cancel := storeContext(j.ctx)
	defer cancel()
	var (
		done  bool
		state string
	)
	err := r.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		t := now()
		res, err := tx.Exec(ctx, `UPDATE jobs SET state = ?, finished_at = ?, updated_at = ?,
	progress = coalesce(nullif(?, ''), progress), failure_code = nullif(?, ''), failure_message = nullif(?, '')
WHERE id = ? AND state = 'running'`, string(o.state), t, t, o.progress, o.code, o.message, j.id)
		if err != nil {
			return fmt.Errorf("record job %s: %w", o.state, err)
		}
		if done, err = affected(res); err != nil || done {
			return err
		}
		return tx.QueryRow(ctx, "SELECT state FROM jobs WHERE id = ?", j.id).Scan(&state)
	})
	if err != nil {
		return err
	}
	if !done {
		r.logger.Warn("job end refused: record already finished", "job", j.handle, "want", string(o.state), "state", state)
	}
	return nil
}

// progress records phase as j's progress and reports whether a cancel
// has been requested, canceling j's context and returning its error
// when one has.
func (r *Runner) progress(j *job, phase string) error {
	ctx, cancel := storeContext(j.ctx)
	defer cancel()
	var requested bool
	err := r.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		if _, err := tx.Exec(ctx, `UPDATE jobs SET progress = ?, updated_at = ? WHERE id = ? AND state = 'running'`,
			phase, now(), j.id); err != nil {
			return fmt.Errorf("record job progress: %w", err)
		}
		return tx.QueryRow(ctx, "SELECT cancel_requested FROM jobs WHERE id = ?", j.id).Scan(&requested)
	})
	if err != nil {
		return err
	}
	if requested {
		j.cancel()
		return j.ctx.Err()
	}
	return nil
}

// pollCancel re-reads j's cancel_requested every [cancelPoll] until
// stop closes, canceling j's context once a cancel is requested.
func (r *Runner) pollCancel(j *job, stop <-chan struct{}) {
	t := time.NewTicker(cancelPoll)
	defer t.Stop()
	for {
		select {
		case <-stop:
			return
		case <-j.ctx.Done():
			return
		case <-t.C:
		}
		ctx, cancel := storeContext(j.ctx)
		var requested bool
		err := r.store.QueryRow(ctx, "SELECT cancel_requested FROM jobs WHERE id = ?", j.id).Scan(&requested)
		cancel()
		if err != nil {
			r.logger.Warn("read job cancel request", "job", j.handle, "err", err)
			continue
		}
		if requested {
			j.cancel()
			return
		}
	}
}

// requestCancel records a cancel of id's job: canceled outright when
// queued, cancel_requested when running, and nothing when finished.
func (r *Runner) requestCancel(ctx context.Context, id []byte) error {
	h := handleOf(id)
	ctx, cancel := context.WithTimeout(ctx, writeWait)
	defer cancel()
	return r.store.Update(ctx, func(ctx context.Context, tx *store.Tx) error {
		var state string
		err := tx.QueryRow(ctx, "SELECT state FROM jobs WHERE id = ?", id).Scan(&state)
		if errors.Is(err, sql.ErrNoRows) {
			return ops.NewNotFoundError(fmt.Sprintf("job %s", h))
		}
		if err != nil {
			return fmt.Errorf("read job %s: %w", h, err)
		}
		t := now()
		switch ops.JobState(state) {
		case ops.JobQueued:
			_, err = tx.Exec(ctx, `UPDATE jobs SET state = 'canceled', cancel_requested = 1, finished_at = ?, updated_at = ?
WHERE id = ? AND state = 'queued'`, t, t, id)
		case ops.JobRunning:
			_, err = tx.Exec(ctx, `UPDATE jobs SET cancel_requested = 1, updated_at = ?
WHERE id = ? AND state = 'running'`, t, id)
		}
		if err != nil {
			return fmt.Errorf("cancel job %s: %w", h, err)
		}
		return nil
	})
}

// status reads id's record.
func (r *Runner) status(ctx context.Context, id []byte) (ops.JobStatus, error) {
	h := handleOf(id)
	var (
		st                   ops.JobStatus
		state                string
		phase, code, message sql.NullString
		queued, updated      int64
		started, finished    sql.NullInt64
	)
	err := r.store.QueryRow(ctx, `SELECT kind, state, progress, failure_code, failure_message, at, started_at, updated_at, finished_at
FROM jobs WHERE id = ?`, id).Scan(&st.Kind, &state, &phase, &code, &message, &queued, &started, &updated, &finished)
	if errors.Is(err, sql.ErrNoRows) {
		return ops.JobStatus{}, ops.NewNotFoundError(fmt.Sprintf("job %s", h))
	}
	if err != nil {
		return ops.JobStatus{}, fmt.Errorf("read job %s: %w", h, err)
	}
	st.ID = h
	st.State = ops.JobState(state)
	st.Phase, st.FailureCode, st.FailureMessage = phase.String, code.String, message.String
	st.QueuedAt, st.UpdatedAt = micros(queued), micros(updated)
	st.StartedAt, st.FinishedAt = optionalMicros(started), optionalMicros(finished)
	return st, nil
}

func micros(us int64) time.Time { return time.UnixMicro(us).UTC() }

func optionalMicros(us sql.NullInt64) *time.Time {
	if !us.Valid {
		return nil
	}
	t := micros(us.Int64)
	return &t
}

// affected reports whether res changed a row.
func affected(res sql.Result) (bool, error) {
	n, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("count changed rows: %w", err)
	}
	return n > 0, nil
}

// parseID parses h as a job handle and returns the record id it names.
func parseID(h ops.Handle) ([]byte, error) {
	if _, err := ops.ParseHandle(string(h)); err != nil {
		return nil, err
	}
	return hex.DecodeString(string(h))
}

func handleOf(id []byte) ops.Handle { return ops.Handle(hex.EncodeToString(id)) }
