package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/spf13/cobra"
)

// The exit statuses restore's outcomes carry beyond run's own default
// of 1. A held store lock carries backup's "try again later" status for
// restore's own lock, and an interrupt carries backup's canceled one.
const (
	exitRestoreLocked      = 75
	exitRestoreInterrupted = 130
)

// restoreRotation is the instruction a finished restore prints last: a
// restored store carries the sessions and sign-in links of the moment
// it was backed up, and rotating both secrets with no previous value
// fails every one of them.
const restoreRotation = `Before starting serve, mint new values for DUBPLATE_SESSION_SECRET and DUBPLATE_MAGIC_LINK_SECRET, and leave DUBPLATE_SESSION_SECRET_PREVIOUS and DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS unset, so every session and sign-in link issued before the restore fails.`

// restoreFlags holds the restore verb's flags.
type restoreFlags struct {
	snapshot       string
	fromFile       string
	resume         bool
	older          bool
	library        bool
	intoNonempty   bool
	discardCurrent bool
}

// restoreConflicts lists each pair of restore flags that never
// combine. The sources' own exclusion is checked apart, so each pair
// here names at most one source.
var restoreConflicts = [][2]string{
	{"--older", "--from-file"},
	{"--older", "--resume"},
	{"--older", "--library"},
	{"--library", "--from-file"},
	{"--library", "--resume"},
	{"--discard-current", "--resume"},
}

// given reports which of f's flags are set, by their names on the
// command line.
func (f restoreFlags) given() map[string]bool {
	return map[string]bool{
		"--snapshot":        f.snapshot != "",
		"--from-file":       f.fromFile != "",
		"--resume":          f.resume,
		"--older":           f.older,
		"--library":         f.library,
		"--into-nonempty":   f.intoNonempty,
		"--discard-current": f.discardCurrent,
	}
}

// check refuses a set of flags the verb cannot run, naming the flags:
// no source or more than one, and each pair of flags that never
// combine. It runs before the configuration loads, so a refusal
// precedes every store open and rclone run.
func (f restoreFlags) check() error {
	given := f.given()
	var sources []string
	for _, name := range []string{"--snapshot", "--from-file", "--resume"} {
		if given[name] {
			sources = append(sources, name)
		}
	}
	switch len(sources) {
	case 0:
		return errors.New("restore: name one source: --snapshot, --from-file, or --resume")
	case 1:
	default:
		return fmt.Errorf("restore: %s never combine: name one source", strings.Join(sources, " and "))
	}
	for _, c := range restoreConflicts {
		if given[c[0]] && given[c[1]] {
			return fmt.Errorf("restore: %s never combines with %s", c[0], c[1])
		}
	}
	if f.intoNonempty && !f.library {
		return errors.New("restore: --into-nonempty applies only with --library")
	}
	return nil
}

// newRestoreCmd builds the restore verb: it checks its flags, loads the
// wall, and runs [backup.Restore] on a context that SIGINT, SIGTERM, and
// SIGHUP cancel. Its Long text lists every exit status the verb
// assigns.
func newRestoreCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	var f restoreFlags
	cmd := &cobra.Command{
		Use:   "restore",
		Short: "Restore the library database, and optionally the library, from a backup",
		Long: `Restore the library database, and optionally the library, from a backup.

The server must be stopped: restore takes the store's lock exclusively
and refuses while any other process holds it. A restore that stops part
way leaves a restore-intent record, and restore --resume finishes it.

Exit statuses:
  0    the restore finished, or left a store older than this binary for serve to finish
  1    a flag, load, or store refusal, or a failed step
  75   another process holds the store's lock
  130  interrupted by a signal (SIGINT, SIGTERM, SIGHUP); restore --resume finishes a restore that wrote a restore-intent record`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runRestore(cmd.Context(), environ, reg, f, stdout, stderr)
		},
	}
	fl := cmd.Flags()
	fl.StringVar(&f.snapshot, "snapshot", "", "restore the remote snapshot with this stamp, or \"latest\"")
	fl.StringVar(&f.fromFile, "from-file", "", "restore from a local store file, such as a pre-migration copy")
	fl.BoolVar(&f.resume, "resume", false, "finish the restore the restore-intent record names")
	fl.BoolVar(&f.older, "older", false, "permit a --snapshot other than the latest; never with --library")
	fl.BoolVar(&f.library, "library", false, "also restore the library tree from the remote, with --snapshot")
	fl.BoolVar(&f.intoNonempty, "into-nonempty", false, "with --library, permit a library root that holds files; each file the copy replaces is kept in a sibling directory")
	fl.BoolVar(&f.discardCurrent, "discard-current", false, "move a current store that cannot be opened aside, unopened")
	cmd.SetErr(stderr)
	return cmd
}

// runRestore runs one restore under f and maps its outcome to an exit
// status. A signal's cancel wins over every other failure, since a
// step it interrupts fails for that reason. Output is best effort: a
// failed write to stdout never decides the outcome, so a restore that
// finished is never reported as failed, and one part way through never
// stops because the terminal went away.
func runRestore(ctx context.Context, environ []string, reg *config.Registry, f restoreFlags, stdout, stderr io.Writer) error {
	if err := f.check(); err != nil {
		return err
	}
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
	var remote string
	if f.snapshot != "" {
		if !cfg.IntegrationEnabled(backupIntegrationClass.Name) {
			return fmt.Errorf("restore --snapshot: backup integration is disabled (%s): a snapshot restore reads the backup remote", backupIntegrationClass.Flag)
		}
		remote = cfg.IntegrationValue(backupIntegrationClass.Name, "DUBPLATE_BACKUP_REMOTE")
	}

	// main's signal context cancels on SIGINT and SIGTERM. A closed
	// terminal sends SIGHUP, whose default action would kill this
	// process and leave rclone, in its own process group, running.
	ctx, stop := signal.NotifyContext(ctx, syscall.SIGHUP)
	defer stop()

	out := &bestEffortWriter{w: stdout}
	err = backup.Restore(ctx, backup.RestoreConfig{
		StateDir:       cfg.StateDir,
		BusyTimeout:    busyTimeout,
		Now:            time.Now().UTC(),
		Snapshot:       f.snapshot,
		Older:          f.older,
		FromFile:       f.fromFile,
		Resume:         f.resume,
		DiscardCurrent: f.discardCurrent,
		Remote:         remote,
		Library:        f.library,
		LibraryRoot:    cfg.LibraryRoot,
		IntoNonempty:   f.intoNonempty,
		Out:            out,
		Logger:         slog.New(slog.NewTextHandler(stderr, nil)),
	})
	switch {
	case err == nil:
	case ctx.Err() != nil:
		return &exitCodeErr{code: exitRestoreInterrupted, err: fmt.Errorf("restore interrupted: %w", err)}
	case errors.Is(err, store.ErrLockBusy):
		return &exitCodeErr{code: exitRestoreLocked, err: err}
	default:
		return err
	}
	_, _ = fmt.Fprintln(out, restoreRotation)
	if out.err != nil {
		_, _ = fmt.Fprintf(stderr, "restore: finished, but its output to stdout was lost: %v; the reconciliation report file in the store holds the report\n", out.err)
	}
	return nil
}

// A bestEffortWriter passes writes to w until one fails, and reports
// every write as done. It keeps the first failure in err.
type bestEffortWriter struct {
	w   io.Writer
	err error
}

func (b *bestEffortWriter) Write(p []byte) (int, error) {
	if b.err == nil {
		if _, err := b.w.Write(p); err != nil {
			b.err = err
		}
	}
	return len(p), nil
}
