package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/id"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/jobs"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/spf13/cobra"
)

// backupPollInterval is how often the verb re-reads a running job's
// status while it waits for a terminal state.
const backupPollInterval = 200 * time.Millisecond

// newBackupCmd builds the backup verb, whose Long text lists every exit
// status this file assigns.
func newBackupCmd(environ []string, reg *config.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "backup",
		Short: "Snapshot the library database and sync it to remote storage",
		Long: `Snapshot the library database and sync it to remote storage.

Exit statuses:
  0    the backup succeeded
  1    the backup failed, or a load or store refusal stopped it before it started
  69   backup integration is disabled (DUBPLATE_BACKUP_ENABLED)
  75   a backup is already running
  130  canceled by any means: a signal (SIGINT, SIGTERM, SIGHUP) or another process's cancel`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runBackup(cmd.Context(), environ, reg, cmd.OutOrStdout(), cmd.ErrOrStderr())
		},
	}
}

// runBackup runs one backup job as a local operator to a terminal
// state. ctx is canceled by a signal, so every runner call runs on a
// context.WithoutCancel copy of it: a status read under ctx would fail
// the instant a signal arrived, before the verb could answer it.
func runBackup(ctx context.Context, environ []string, reg *config.Registry, stdout, stderr io.Writer) error {
	logger := verbLogger(stderr)
	opsCtx := context.WithoutCancel(ctx)

	openCtx, cancelOpen := context.WithTimeout(opsCtx, storeWait)
	defer cancelOpen()
	h, cfg, err := openOperatorStore(openCtx, environ, reg)
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

	actor := identity.LocalActor(resolveActor())
	startCtx, cancelStart := context.WithTimeout(opsCtx, storeWait)
	jobID, err := runner.StartBackup(startCtx, actor)
	cancelStart()
	if err != nil {
		return withExitStatus(err)
	}
	if _, err := fmt.Fprintf(stdout, "backup %s: started\n", jobID); err != nil {
		return err
	}

	return awaitBackup(ctx, opsCtx, runner, actor, jobID, stdout, logger)
}

// logSettingsEvaluation logs one warning per refusal or disagreement
// the seed rule finds against the store. backup never applies the rule
// and never fails on it: serve's seed at its next start carries the
// change.
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

// awaitBackup polls jobID's status until it turns terminal, requesting
// a cancel on the first signal: SIGINT or SIGTERM through signalCtx,
// and SIGHUP through its own handler. A closed terminal sends SIGHUP,
// whose default action would kill this process while rclone, in its
// own process group, survives. Runner calls run on opsCtx, so they
// keep working after signalCtx is done.
func awaitBackup(signalCtx, opsCtx context.Context, runner *jobs.Runner, actor identity.Actor, jobID id.ID, stdout io.Writer, logger *slog.Logger) error {
	hup := make(chan os.Signal, 1)
	signal.Notify(hup, syscall.SIGHUP)
	defer signal.Stop(hup)

	ticker := time.NewTicker(backupPollInterval)
	defer ticker.Stop()

	// A nil channel never fires: each signal cancels once.
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
// operation's write. When that write fails, the job would run on
// unanswered, so it cancels the job in-process instead and logs the
// error: the job still ends canceled, and the next poll reports it.
func requestCancel(opsCtx context.Context, runner *jobs.Runner, actor identity.Actor, jobID id.ID, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(opsCtx, storeWait)
	defer cancel()
	if _, err := runner.CancelJob(ctx, actor, jobID); err != nil {
		logger.Error("cancel job", "job", jobID, "err", err)
		runner.CancelOwn(jobID)
	}
}

// reportOutcome prints a succeeded job to stdout and returns nil, and
// returns every other terminal state of job as an error, carrying its
// exit status where that is not 1.
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
		return &exitError{code: exitInterrupted, err: fmt.Errorf("backup %s: canceled", job.JobID)}
	case ops.JobFailed:
		return fmt.Errorf("backup %s: failed, code %s: %s", job.JobID, job.FailureCode, job.FailureMessage)
	default:
		// Interrupted lands here when a failed write stranded the job's
		// record and the next status read recovered it.
		return fmt.Errorf("backup %s: %s", job.JobID, job.State)
	}
}

// withExitStatus wraps a StartBackup error in the exit status it maps
// to: unavailable while backup integration is off, conflict while
// another backup holds the kind lease. Every other error returns
// unwrapped, for run's own default of 1.
func withExitStatus(err error) error {
	switch {
	case errors.Is(err, ops.CodeUnavailable):
		return &exitError{code: exitUnavailable, err: err}
	case errors.Is(err, ops.CodeConflict):
		return &exitError{code: exitTempFail, err: err}
	default:
		return err
	}
}
