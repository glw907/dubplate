package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/jobs"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
	"github.com/spf13/cobra"
)

// backupPollInterval is how often the verb re-reads a running job's
// status while it waits for a terminal state.
const backupPollInterval = 200 * time.Millisecond

// newBackupCmd builds the backup verb: it loads the wall, opens the
// store in write mode, starts a backup as a local operator, polls its
// status, and exits with the status its outcome maps to. Its Long text
// lists every exit status this file assigns, so an operator reading
// --help sees the same table the verb's tests assert against.
func newBackupCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backup",
		Short: "Snapshot the library database and sync it to remote storage",
		Long: `Snapshot the library database and sync it to remote storage.

Exit statuses:
  0    the backup succeeded
  1    the backup failed, or a load or store refusal stopped it before it started
  69   backup integration is disabled (DUBPLATE_BACKUP_ENABLED)
  75   a backup is already running
  130  canceled by any means: a signal (SIGINT, SIGTERM, SIGHUP) or another process's cancel`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBackup(cmd.Context(), environ, reg, stdout, stderr)
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// runBackup loads the wall, opens the store in write mode, and runs
// the backup job to a terminal state. Every runner call after
// construction runs on opsCtx, a context.WithoutCancel copy of ctx:
// ctx is the verb's own signal-canceled context
// (executeWithSignalContext, main.go), and a job or a status read that
// ran under it would fail the instant a signal arrived rather than
// letting the verb answer the signal itself.
func runBackup(ctx context.Context, environ []string, reg *config.Registry, stdout, stderr io.Writer) error {
	cfg, err := loadConfig(environ, reg)
	if err != nil {
		return err
	}
	busyTimeout, err := store.BusyTimeout(environ)
	if err != nil {
		return err
	}

	logger := slog.New(slog.NewTextHandler(stderr, nil))
	opsCtx := context.WithoutCancel(ctx)

	openCtx, cancelOpen := context.WithTimeout(opsCtx, storeWait)
	defer cancelOpen()
	h, err := store.OpenWrite(openCtx, cfg.StateDir, busyTimeout)
	if err != nil {
		return err
	}
	defer func() {
		if cerr := h.Close(); cerr != nil {
			logger.Error("close store", "err", cerr)
		}
	}()

	evalCtx, cancelEval := context.WithTimeout(opsCtx, storeWait)
	ev, evalErr := settings.Evaluate(evalCtx, h, reg, environ)
	cancelEval()
	if evalErr != nil {
		logger.Error("evaluate settings", "err", evalErr)
	} else {
		logSettingsEvaluation(logger, ev)
	}

	acc := settings.NewAccessor(h, reg)
	runnerCtx, cancelRunner := context.WithTimeout(opsCtx, storeWait)
	defer cancelRunner()
	runner, err := jobs.New(runnerCtx, jobs.Config{
		Store:     h,
		Retention: jobs.JobRetention(acc),
		Backup:    jobs.BackupConfigFrom(cfg, acc),
		Logger:    logger,
		Now:       time.Now,
	})
	if err != nil {
		return err
	}
	defer func() {
		if cerr := runner.Close(); cerr != nil {
			logger.Error("close job runner", "err", cerr)
		}
	}()

	actor := identity.LocalActor(resolveActor(user.LookupId))
	startCtx, cancelStart := context.WithTimeout(opsCtx, storeWait)
	jobID, err := runner.StartBackup(startCtx, actor)
	cancelStart()
	if err != nil {
		return backupExitCode(err)
	}
	if _, err := fmt.Fprintf(stdout, "backup %s: started\n", jobID); err != nil {
		return err
	}

	return awaitBackup(ctx, opsCtx, runner, actor, jobID, stdout, logger)
}

// logSettingsEvaluation logs one warning line per disagreement or
// refusal the seed rule finds against the store: a refusal names the
// key and its reason, an environment disagreement names the key and
// what serve applies at its next start. backup never applies the
// rule and never fails on its outcome: serve's own seed at its next
// start is what carries a change, this run included.
func logSettingsEvaluation(logger *slog.Logger, ev settings.Evaluation) {
	for _, o := range ev.Outcomes {
		switch {
		case o.Err != nil:
			logger.Warn("setting refused; serve will not start until it is resolved", "key", o.Key, "reason", o.Err)
		case o.Change != settings.Unchanged:
			logger.Warn("setting disagrees with the environment; serve applies it at its next start", "key", o.Key, "outcome", o.Report)
		}
	}
}

// awaitBackup polls jobID's status until it turns terminal, calling the
// cancel operation on the first signal that arrives: SIGINT and
// SIGTERM through signalCtx's own Done channel (executeWithSignalContext
// already watches for them), and SIGHUP through this function's own
// handler, since main's signal context does not (a closed
// terminal sends SIGHUP, whose default action would kill this process
// while rclone, in its own process group, survives). Every runner call
// here runs on opsCtx, never signalCtx, so a poll or a cancel write
// keeps working after signalCtx is done.
func awaitBackup(signalCtx, opsCtx context.Context, runner *jobs.Runner, actor identity.Actor, jobID id.ID, stdout io.Writer, logger *slog.Logger) error {
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	defer signal.Stop(hup)

	ticker := time.NewTicker(backupPollInterval)
	defer ticker.Stop()

	// done and hupCh are nil'd out once their cancel fires, so a
	// signal already answered never wins the select again: a closed
	// channel is always ready, and without this the loop would spin
	// on JobStatus at full CPU for the rest of the wait instead of
	// blocking on ticker.C.
	done := signalCtx.Done()
	hupCh := hup
	for {
		select {
		case <-done:
			done = nil
			requestCancel(opsCtx, runner, actor, jobID, logger)
		case <-hupCh:
			hupCh = nil
			requestCancel(opsCtx, runner, actor, jobID, logger)
		case <-ticker.C:
		}

		statusCtx, cancel := context.WithTimeout(opsCtx, storeWait)
		job, err := runner.JobStatus(statusCtx, actor, jobID)
		cancel()
		if err != nil {
			return err
		}
		if job.State.Terminal() {
			return reportOutcome(stdout, job)
		}
	}
}

// requestCancel asks the runner to cancel jobID through the cancel
// operation's own write. When the write fails (the store busy past its
// busy timeout, or any other error), the job would otherwise run to
// completion unanswered, so this cancels it in-process through the
// runner this verb owns instead, and logs the write error rather than
// returning it: the job still ends canceled, through its own terminal
// write, and awaitBackup's next poll reports that.
func requestCancel(opsCtx context.Context, runner *jobs.Runner, actor identity.Actor, jobID id.ID, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(opsCtx, storeWait)
	defer cancel()
	if _, err := runner.CancelJob(ctx, actor, jobID); err != nil {
		logger.Error("cancel job", "job", jobID, "err", err)
		runner.CancelOwn(jobID)
	}
}

// reportOutcome prints job's terminal state and reports run's exit
// status for it. A succeeded job prints to stdout and returns nil;
// every other terminal state returns an error, whose message run
// prints to stderr, wrapped with its exit status where that status is
// not the default of 1.
func reportOutcome(stdout io.Writer, job ops.JobStatus) error {
	switch job.State {
	case ops.JobSucceeded:
		msg := fmt.Sprintf("backup %s: succeeded", job.JobID)
		if job.Phase != "" {
			msg += ": " + job.Phase
		}
		_, err := fmt.Fprintln(stdout, msg)
		return err
	case ops.JobCanceled:
		return &exitCodeErr{code: exitInterrupted, err: fmt.Errorf("backup %s: canceled", job.JobID)}
	case ops.JobFailed:
		return fmt.Errorf("backup %s: failed, code %s: %s", job.JobID, job.FailureCode, job.FailureMessage)
	default:
		// Interrupted reaches this arm in-process when the job's running
		// or terminal write failed and the next status read recovered
		// the stranded record. It exits 1 like any other non-success.
		return fmt.Errorf("backup %s: %s", job.JobID, job.State)
	}
}

// backupExitCode maps a StartBackup error to its exit status: unavailable
// while backup integration is off, conflict while another backup holds
// the kind lease, and every other error unwrapped, run's own default
// of 1.
func backupExitCode(err error) error {
	switch {
	case errors.Is(err, ops.ErrUnavailable):
		return &exitCodeErr{code: exitUnavailable, err: err}
	case errors.Is(err, ops.ErrConflict):
		return &exitCodeErr{code: exitTempFail, err: err}
	default:
		return err
	}
}
