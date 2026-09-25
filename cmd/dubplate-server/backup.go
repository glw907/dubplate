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
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/local"
	"github.com/spf13/cobra"
)

// backupIntegrationClass is rung 6's backup integration: its own
// enable flag and the one key required while it is on, the rclone
// destination a backup syncs to. Declared in this rung's own file per
// the rung 1b registry contract.
var backupIntegrationClass = config.IntegrationClass{
	Name: "backup",
	Rung: 6,
	Flag: "DUBPLATE_BACKUP_ENABLED",
	Required: []string{
		"DUBPLATE_BACKUP_REMOTE",
	},
}

// registerBackupConfig adds backupIntegrationClass to reg.
// newProductionRegistry calls it after registerStoreConfig. The six
// names backup replaced (five DUBPLATE_R2_* names and
// DUBPLATE_BACKUP_STAGING) are superseded entries declared in
// internal/config/keys.go, on rung 5's own precedent
// (ee79b5b26d9d8614357c26e7410d9db7d01673a7): a supersession message
// is base-wall record-keeping, not a rung's own owned surface, so it
// stays with the package that seeds every registry rather than moving
// to this rung's file. The "backup targets" reserved class those six
// names occupied left keys.go in the same commit that added their
// supersession messages, since a name claimed by both families panics
// at registration (internal/config/registry.go, claim).
func registerBackupConfig(reg *config.Registry) {
	reg.AddIntegrationClass(backupIntegrationClass)
}

// backupConfigFrom builds ops/local's backup configuration from cfg:
// the flag and the remote through the integration class, and the
// library root through the owned wall row rung 1 declared (R-6bi18).
func backupConfigFrom(cfg *config.Config) local.BackupConfig {
	return local.BackupConfig{
		Enabled:     cfg.IntegrationEnabled(backupIntegrationClass.Name),
		Remote:      cfg.IntegrationValue(backupIntegrationClass.Name, "DUBPLATE_BACKUP_REMOTE"),
		LibraryRoot: cfg.LibraryRoot,
	}
}

// The exit statuses this verb's outcomes carry beyond run's own
// default of 1 (Ruling 3). 0 (succeeded) and 1 (failed, or any load or
// store refusal) need no wrapping: run already returns them for a nil
// or a plain error.
const (
	exitBackupUnavailable = 69
	exitBackupConflict    = 75
	exitBackupCanceled    = 130
)

// backupVerbWait bounds each of the verb's own store and runner calls:
// how long it waits for the store's lock, another process's write, or
// one status read.
const backupVerbWait = 30 * time.Second

// backupPollInterval is how often the verb re-reads a running job's
// status while it waits for a terminal state.
const backupPollInterval = 200 * time.Millisecond

// newBackupCmd builds the backup verb (R-6bi10, R-6bi11): it loads the
// wall, opens the store in write mode, starts a backup as a local
// operator, polls its status, and exits by Ruling 3. Its Long text
// lists every exit status this file assigns, so an operator reading
// --help sees the same table criterion 20's own test asserts against.
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
// construction runs on opsCtx, a context.WithoutCancel copy of ctx
// (R-6bi11): ctx is the verb's own signal-canceled context
// (executeWithSignalContext, main.go), and a job or a status read that
// ran under it would fail the instant a signal arrived rather than
// letting the verb answer the signal itself.
func runBackup(ctx context.Context, environ []string, reg *config.Registry, stdout, stderr io.Writer) error {
	cfg, err := config.Load(environ, reg)
	if err != nil {
		return err
	}
	busyTimeout, parseErr, rangeErr := storeBusyTimeout(environ)
	if parseErr != nil {
		return parseErr
	}
	if rangeErr != nil {
		return rangeErr
	}

	logger := slog.New(slog.NewTextHandler(stderr, nil))
	opsCtx := context.WithoutCancel(ctx)

	openCtx, cancelOpen := context.WithTimeout(opsCtx, backupVerbWait)
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

	runnerCtx, cancelRunner := context.WithTimeout(opsCtx, backupVerbWait)
	defer cancelRunner()
	runner, err := local.New(runnerCtx, local.Config{
		Store:  h,
		Backup: backupConfigFrom(cfg),
		Logger: logger,
	})
	if err != nil {
		return err
	}
	defer func() {
		if cerr := runner.Close(); cerr != nil {
			logger.Error("close job runner", "err", cerr)
		}
	}()

	actor := ops.LocalActor(resolveActor())
	startCtx, cancelStart := context.WithTimeout(opsCtx, backupVerbWait)
	resp, err := runner.StartBackup(startCtx, actor, ops.StartBackupRequest{})
	cancelStart()
	if err != nil {
		return wrapExitCode(err)
	}
	if _, err := fmt.Fprintf(stdout, "backup %s: started\n", resp.JobID); err != nil {
		return err
	}

	return awaitBackup(ctx, opsCtx, runner, actor, resp.JobID, stdout, logger)
}

// awaitBackup polls h's status until it turns terminal, calling the
// cancel operation on the first signal that arrives: SIGINT and
// SIGTERM through signalCtx's own Done channel (executeWithSignalContext
// already watches for them), and SIGHUP through this function's own
// handler, since main's signal context does not (R-6bi11: a closed
// terminal sends SIGHUP, whose default action would kill this process
// while rclone, in its own process group, survives). Every runner call
// here runs on opsCtx, never signalCtx, so a poll or a cancel write
// keeps working after signalCtx is done.
func awaitBackup(signalCtx, opsCtx context.Context, runner *local.Runner, actor ops.Actor, h ops.Handle, stdout io.Writer, logger *slog.Logger) error {
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
			requestCancel(opsCtx, runner, actor, h, logger)
		case <-hupCh:
			hupCh = nil
			requestCancel(opsCtx, runner, actor, h, logger)
		case <-ticker.C:
		}

		statusCtx, cancel := context.WithTimeout(opsCtx, backupVerbWait)
		resp, err := runner.JobStatus(statusCtx, actor, ops.JobStatusRequest{JobID: h})
		cancel()
		if err != nil {
			return err
		}
		if resp.Job.State.Terminal() {
			return reportOutcome(stdout, resp.Job)
		}
	}
}

// requestCancel asks the runner to cancel h through the cancel
// operation's own write. When the write fails (the store busy past its
// busy timeout, or any other error), the job would otherwise run to
// completion unanswered, so this cancels it in-process through the
// runner this verb owns instead, and logs the write error rather than
// returning it: the job still ends canceled, through its own terminal
// write, and awaitBackup's next poll reports that.
func requestCancel(opsCtx context.Context, runner *local.Runner, actor ops.Actor, h ops.Handle, logger *slog.Logger) {
	ctx, cancel := context.WithTimeout(opsCtx, backupVerbWait)
	defer cancel()
	if _, err := runner.CancelJob(ctx, actor, ops.CancelJobRequest{JobID: h}); err != nil {
		logger.Error("cancel job", "job", h, "err", err)
		runner.CancelOwn(h)
	}
}

// reportOutcome prints h's terminal state and reports run's exit
// status for it (Ruling 3). A succeeded job prints to stdout and
// returns nil, run's own 0; every other terminal state returns an
// error, whose message run prints to stderr, wrapped with its exit
// status where that status is not run's own default of 1.
func reportOutcome(stdout io.Writer, job ops.JobStatus) error {
	switch job.State {
	case ops.JobSucceeded:
		msg := fmt.Sprintf("backup %s: succeeded", job.ID)
		if job.Phase != "" {
			msg += ": " + job.Phase
		}
		_, err := fmt.Fprintln(stdout, msg)
		return err
	case ops.JobCanceled:
		return &exitCodeErr{code: exitBackupCanceled, err: fmt.Errorf("backup %s: canceled", job.ID)}
	case ops.JobFailed:
		return fmt.Errorf("backup %s: failed, code %s: %s", job.ID, job.FailureCode, job.FailureMessage)
	default:
		// Interrupted is unreachable in-process: this verb's own runner
		// holds the job's lease for its whole life, so recovery never
		// touches this verb's own record. Kept so the state set this
		// switch handles stays closed over every [ops.JobState].
		return fmt.Errorf("backup %s: %s", job.ID, job.State)
	}
}

// wrapExitCode maps a StartBackup error to its exit status: unavailable
// while backup integration is off, conflict while another backup holds
// the kind lease, and every other error unwrapped, run's own default
// of 1.
func wrapExitCode(err error) error {
	switch {
	case errors.Is(err, ops.ErrUnavailable):
		return &exitCodeErr{code: exitBackupUnavailable, err: err}
	case errors.Is(err, ops.ErrConflict):
		return &exitCodeErr{code: exitBackupConflict, err: err}
	default:
		return err
	}
}
