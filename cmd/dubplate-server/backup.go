package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"os/signal"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/ops"
	"github.com/glw907/dubplate/internal/ops/local"
	"github.com/glw907/dubplate/internal/settings"
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

// backupGuardFractionSetting is DUBPLATE_BACKUP_GUARD_FRACTION: the
// share of the catalog's tracks a library tree must hold as regular
// files before a sync may run. Its default, 0.5, carries unchanged the
// value internal/backup's retired guardFraction constant held
// (R-6bi8). The floor is the protective minimum against a value that
// would turn the short-tree guard off by accident (Ruling 3, R-6c11).
var backupGuardFractionSetting = config.Setting{
	Key:     "DUBPLATE_BACKUP_GUARD_FRACTION",
	Kind:    config.KindFloat,
	Default: "0.5",
	Doc:     "a fraction greater than 0 and at most 1, for example 0.5",
	Apply:   config.ApplyNextOperation,
	Floor:   "0.25",
	Ceiling: "1",
}

// backupSnapshotKeepSetting is DUBPLATE_BACKUP_SNAPSHOT_KEEP: how many
// remote snapshots a prune keeps in total, the run's own included. Its
// default, 14, carries unchanged the value internal/backup's retired
// snapshotKeep constant held (R-6bi8). The floor is the protective
// minimum against a keep that would let one run delete every other
// pre-upgrade remote snapshot (Ruling 3, R-6c11; spec ruling 10).
var backupSnapshotKeepSetting = config.Setting{
	Key:     "DUBPLATE_BACKUP_SNAPSHOT_KEEP",
	Kind:    config.KindInt,
	Default: "14",
	Doc:     "an integer count of remote snapshots to keep, for example 14",
	Apply:   config.ApplyNextOperation,
	Floor:   "3",
	Ceiling: "1000",
}

// backupArchiveRetentionSetting is DUBPLATE_BACKUP_ARCHIVE_RETENTION:
// how long an archive directory survives past its stamp. Its default,
// 2160h (90 days), carries unchanged the value internal/backup's
// retired archiveRetention constant held (R-6bi8). The floor is the
// protective minimum against a retention that would purge within a day
// the files a bad sync moved into archive/ (Ruling 3, R-6c11).
var backupArchiveRetentionSetting = config.Setting{
	Key:     "DUBPLATE_BACKUP_ARCHIVE_RETENTION",
	Kind:    config.KindDuration,
	Default: "2160h",
	Doc:     "a Go duration string, for example 2160h",
	Apply:   config.ApplyNextOperation,
	Floor:   "168h",
	Ceiling: "87600h",
}

// registerBackupConfig adds backupIntegrationClass and the three
// backup-homed settings to reg. newProductionRegistry calls it after
// registerStoreConfig. The six names backup replaced (five
// DUBPLATE_R2_* names and DUBPLATE_BACKUP_STAGING) are superseded
// entries declared in internal/config/keys.go, on rung 5's own
// precedent (ee79b5b26d9d8614357c26e7410d9db7d01673a7): a supersession
// message is base-wall record-keeping, not a rung's own owned surface,
// so it stays with the package that seeds every registry rather than
// moving to this rung's file. The "backup targets" reserved class
// those six names occupied left keys.go in the same commit that added
// their supersession messages, since a name claimed by both families
// panics at registration (internal/config/registry.go, claim).
func registerBackupConfig(reg *config.Registry) {
	reg.AddIntegrationClass(backupIntegrationClass)
	reg.AddSetting(backupGuardFractionSetting)
	reg.AddSetting(backupSnapshotKeepSetting)
	reg.AddSetting(backupArchiveRetentionSetting)
}

// backupSettingsReaders returns the reader functions [runBackup]
// injects into [local.BackupConfig]: the accessor's typed read of each
// of the three backup-homed settings, one per value ops/local needs
// (R-6c11).
func backupSettingsReaders(acc *settings.Accessor) (
	guardFraction func(context.Context) (float64, error),
	snapshotKeep func(context.Context) (int, error),
	archiveRetention func(context.Context) (time.Duration, error),
) {
	return func(ctx context.Context) (float64, error) { return acc.Float(ctx, backupGuardFractionSetting) },
		func(ctx context.Context) (int, error) { return acc.Int(ctx, backupSnapshotKeepSetting) },
		func(ctx context.Context) (time.Duration, error) {
			return acc.Duration(ctx, backupArchiveRetentionSetting)
		}
}

// backupConfigFrom builds ops/local's backup configuration from cfg
// and acc: the flag and the remote through the integration class, the
// library root through the owned wall row rung 1 declared (R-6bi18),
// and the three backup value readers through acc (R-6c11).
func backupConfigFrom(cfg *config.Config, acc *settings.Accessor) local.BackupConfig {
	guardFraction, snapshotKeep, archiveRetention := backupSettingsReaders(acc)
	return local.BackupConfig{
		Enabled:          cfg.IntegrationEnabled(backupIntegrationClass.Name),
		Remote:           cfg.IntegrationValue(backupIntegrationClass.Name, "DUBPLATE_BACKUP_REMOTE"),
		LibraryRoot:      cfg.LibraryRoot,
		GuardFraction:    guardFraction,
		SnapshotKeep:     snapshotKeep,
		ArchiveRetention: archiveRetention,
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

	acc := settings.NewAccessor(h, reg)
	runnerCtx, cancelRunner := context.WithTimeout(opsCtx, backupVerbWait)
	defer cancelRunner()
	runner, err := local.New(runnerCtx, local.Config{
		Store:     h,
		Retention: jobRetentionReader(acc),
		Backup:    backupConfigFrom(cfg, acc),
		Logger:    logger,
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

// doctor's own backup rows. backupRemoteRow prints unconditionally,
// the "backup remote" row's own key, distinct from the generic
// integration-class loop's "backup" row (enabled/disabled), which
// runs regardless.
const (
	backupRemoteRow      = "backup remote"
	backupLastSuccessRow = "backup last success"
	backupWALSizeRow     = "backup wal size"
	backupFreeSpaceRow   = "backup free space"
	backupLocalCopiesRow = "backup local copies"
)

// featuresProbeTimeout bounds the `rclone backend features` probe
// (Ruling 2's 10 seconds). A var, not a const, so a test can shorten
// it rather than wait out the real bound.
var featuresProbeTimeout = 10 * time.Second

// featuresProbeWaitDelay bounds how long the probe's child has to exit
// after its SIGINT before it is killed, [os/exec.Cmd.WaitDelay]'s own
// precedent in internal/backup/rclone.go.
const featuresProbeWaitDelay = 5 * time.Second

// backupCryptWarning is ruling 14's warning: a backup remote that is
// not wrapped in an rclone crypt remote carries every snapshot,
// addresses included after 6d, unencrypted at the provider.
const backupCryptWarning = "the backup remote is not wrapped in an rclone crypt remote; a snapshot carries member data unencrypted at the provider"

// backupRemoteRowStatus renders doctor's "backup remote" row (R-6bi14,
// R-6bi9, Ruling 2): unavailable while the flag is off, without
// running any subprocess; the parse, _TYPE, and crypt checks doctor
// alone runs, since [backup.Job.Run] reads no environment; and the
// features probe last, under featuresProbeTimeout. It calls
// [backup.ParseRemote], the same validator the job's first phase
// calls, so the two never disagree about a well-formed remote.
func backupRemoteRowStatus(ctx context.Context, environ []string) (status string, failed bool) {
	enabled, _ := backupIntegrationClass.Enabled(environ)
	if !enabled {
		return "unavailable: backup integration disabled", false
	}

	remoteValue, _ := lookupEnviron(environ, "DUBPLATE_BACKUP_REMOTE")
	if _, err := backup.ParseRemote(remoteValue); err != nil {
		return backupRemoteParseError(remoteValue), true
	}
	name, _, _ := strings.Cut(remoteValue, ":")
	upper := strings.ToUpper(name)

	typeKey := "RCLONE_CONFIG_" + upper + "_TYPE"
	remoteType, ok := lookupEnviron(environ, typeKey)
	if !ok {
		return "invalid remote: " + typeKey + " is not set", true
	}

	if remoteType == "crypt" {
		if status, failed, blocked := backupCryptCheck(environ, upper); blocked {
			return status, failed
		}
	}

	bin, err := exec.LookPath("rclone")
	if err != nil {
		return "rclone not found", true
	}

	status, failed = backupFeaturesStatus(ctx, bin, name)
	if remoteType != "crypt" {
		status += "; " + backupCryptWarning
	}
	return status, failed
}

// backupRemoteParseError renders [backup.ParseRemote]'s refusal of
// value as one of the two states table shapes: the generic "want
// name:path" for no colon or an empty name, or the offending character
// quoted against rclone's own rule for a name holding one.
func backupRemoteParseError(value string) string {
	name, _, hasColon := strings.Cut(value, ":")
	if !hasColon || name == "" {
		return "invalid remote: want <name>:<path>"
	}
	if r, ok := firstInvalidRemoteNameRune(name); ok {
		return fmt.Sprintf("invalid remote: name %q holds %q, and rclone remote names \"can only contain letters, digits, or the _ (underscore) character\"", name, string(r))
	}
	// Unreachable given ParseRemote's own rule (name and hasColon
	// already passed), kept so this function never returns an empty
	// status for a value ParseRemote refuses for a reason this switch
	// does not yet know.
	return "invalid remote: " + value
}

// firstInvalidRemoteNameRune returns the first rune in name outside
// rclone's remote-name alphabet (ASCII letters, digits, underscore).
func firstInvalidRemoteNameRune(name string) (rune, bool) {
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z', r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			continue
		}
		return r, true
	}
	return 0, false
}

// backupCryptCheck runs the crypt remote's own two checks against the
// remote named by upper: that RCLONE_CONFIG_<upper>_REMOTE is set, and,
// when its value names another remote rather than a plain path or a
// connection string, that the named remote's own _TYPE is set. blocked
// is true only when a check refuses; status and failed are meaningful
// only then. A plain path (no colon) or a connection string (a leading
// colon) skips the wrapped check, leaving the features probe to decide
// (R-6bi14).
func backupCryptCheck(environ []string, upper string) (status string, failed, blocked bool) {
	remoteKey := "RCLONE_CONFIG_" + upper + "_REMOTE"
	wrapped, ok := lookupEnviron(environ, remoteKey)
	if !ok {
		return "invalid remote: " + remoteKey + " is not set", true, true
	}
	wrappedName, _, hasColon := strings.Cut(wrapped, ":")
	if !hasColon || wrappedName == "" {
		return "", false, false
	}
	wrappedTypeKey := "RCLONE_CONFIG_" + strings.ToUpper(wrappedName) + "_TYPE"
	if _, ok := lookupEnviron(environ, wrappedTypeKey); !ok {
		return fmt.Sprintf("invalid remote: wrapped remote %s: %s is not set", wrappedName, wrappedTypeKey), true, true
	}
	return "", false, false
}

// backupFeaturesJSON is the shape of `rclone backend features`'s
// stdout this probe reads: everything else in the object is
// backend-specific and unread.
type backupFeaturesJSON struct {
	Features struct {
		Copy bool
		Move bool
	}
}

// backupFeaturesStatus runs `rclone backend features <name>:
// --low-level-retries 1` (R-6bi9) under featuresProbeTimeout and
// renders its outcome: "not checked: ..." on a deadline hit, whatever
// the child's own exit (R-6bi12's precedence, the same rule
// [classify] in internal/backup/rclone.go takes for the backup job's
// own child); the child's last stderr line on any other nonzero exit;
// and, on success, a refusal naming name when neither Copy nor Move is
// true, since rclone's sync then has no backend to satisfy
// --backup-dir.
func backupFeaturesStatus(ctx context.Context, bin, name string) (status string, failed bool) {
	probeCtx, cancel := context.WithTimeout(ctx, featuresProbeTimeout)
	defer cancel()

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(probeCtx, bin, "backend", "features", name+":", "--low-level-retries", "1")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Cancel = func() error { return cmd.Process.Signal(os.Interrupt) }
	cmd.WaitDelay = featuresProbeWaitDelay
	err := cmd.Run()

	if probeCtx.Err() != nil {
		return "not checked: rclone backend features timed out", false
	}
	if err != nil {
		if line := lastNonBlankLine(stderr.String()); line != "" {
			return line, true
		}
		return fmt.Sprintf("rclone backend features %s: %v", name, err), true
	}

	var parsed backupFeaturesJSON
	if err := json.Unmarshal(stdout.Bytes(), &parsed); err != nil {
		return fmt.Sprintf("rclone backend features %s: could not parse its output: %v", name, err), true
	}
	if !parsed.Features.Copy && !parsed.Features.Move {
		return fmt.Sprintf(`backend %s lacks server-side move and copy: rclone's manual requires one for --backup-dir ("the remote in use must support server-side move or copy")`, name), true
	}
	return "ok", false
}

// lastNonBlankLine returns the last non-blank line of s, trimmed, or
// "" when s holds none.
func lastNonBlankLine(s string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for _, i := range slices.Backward(lines) {
		if line := strings.TrimSpace(i); line != "" {
			return line
		}
	}
	return ""
}
