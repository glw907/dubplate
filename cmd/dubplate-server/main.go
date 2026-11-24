// Command dubplate-server organizes and serves one shared music library.
//
// Its subcommands:
//
//   - version: print the build version and matcher commit
//   - check: validate configuration, exit nonzero on failure
//   - doctor: report the full config wall's effective values
//   - serve: run the HTTP server
//   - member: manage the roster of group members
//   - backup: snapshot the library database and sync it to remote storage
//   - restore: restore a library snapshot
//   - unidecode install: build the pinned Unidecode transliteration
//     table into the state directory, online or from --from FILE
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"os/user"
	"syscall"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/intake"
	"github.com/glw907/dubplate/internal/ops/admin"
	"github.com/glw907/dubplate/internal/ops/jobs"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/store"
	"github.com/glw907/dubplate/internal/webapp"
	"github.com/spf13/cobra"
)

// version is set via -ldflags "-X main.version=..." at release build
// time; it stays "dev" for workstation builds.
var version = "dev"

// matcherCommit is the Go matcher revision this build carries, set via
// -ldflags "-X main.matcherCommit=..." by scripts/deploy.sh alone. Every
// other build leaves it "dev".
var matcherCommit = "dev"

func main() {
	os.Exit(run(newRootCmd()))
}

// run executes cmd under the signal-derived context and returns the
// process's exit status: 0 for a nil error, the code an *exitCodeErr
// names, and 1 for any other error, which it prints.
func run(cmd *cobra.Command) int {
	err := executeWithSignalContext(cmd)
	if err == nil {
		return 0
	}
	fmt.Fprintln(os.Stderr, err)
	if ece, ok := errors.AsType[*exitCodeErr](err); ok {
		return ece.code
	}
	return 1
}

// storeWait bounds each store open, write, and status read a verb or
// the serve seed makes. [store.Handle.Update] refuses a context with no
// deadline, and the busy timeout, not this wait, bounds another
// process's lock.
const storeWait = 30 * time.Second

// The exit statuses a verb's outcome can carry beyond run's default of
// 1, which any plain refusal gets.
const (
	exitUnavailable = 69  // EX_UNAVAILABLE: backup integration is disabled
	exitTempFail    = 75  // EX_TEMPFAIL: another backup or a restore holds the lock
	exitInterrupted = 130 // 128+SIGINT: canceled by a signal or another process
)

// An exitCodeErr wraps an error with the process exit status its
// outcome carries, so run never matches on error text.
type exitCodeErr struct {
	code int
	err  error
}

func (e *exitCodeErr) Error() string { return e.err.Error() }
func (e *exitCodeErr) Unwrap() error { return e.err }

// executeWithSignalContext runs cmd under a context canceled by SIGINT
// or SIGTERM. Without it serve's graceful drain is unreachable: cobra's
// Execute passes context.Background, whose Done channel is nil, so the
// process would die on the signal's default action with every in-flight
// request cut.
func executeWithSignalContext(cmd *cobra.Command) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return cmd.ExecuteContext(ctx)
}

// newProductionRegistry builds the registry the binary runs against:
// every package's registration, in the order the registry dump holds.
// The serve package registers twice because its thresholds and its
// trusted proxy class sit at two places in that order.
func newProductionRegistry() *config.Registry {
	reg := config.NewRegistry()
	auth.Register(reg)
	jobs.Register(reg)
	serve.Register(reg)
	store.Register(reg)
	webapp.Register(reg)
	backup.Register(reg)
	serve.RegisterTrustedProxy(reg)
	admin.Register(reg)
	intake.Register(reg)
	return reg
}

// loadConfig is config.Load plus the check no single row can express:
// staging and the library root must not overlap. Every verb that loads
// the wall loads it here, so backup and restore refuse an overlap as
// serve does.
func loadConfig(environ []string, reg *config.Registry) (*config.Config, error) {
	cfg, err := config.Load(environ, reg)
	if err != nil {
		return nil, err
	}
	if err := intake.ValidateStaging(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// newRootCmd builds the production registry once and passes it to every
// subcommand that enforces against it.
func newRootCmd() *cobra.Command {
	reg := newProductionRegistry()
	environ := os.Environ()
	operator := operatorVerbs{WriteWait: storeWait, LookupID: user.LookupId}

	cmd := &cobra.Command{
		Use:           "dubplate-server",
		Short:         "organizes and serves one shared music library",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(
		newVersionCmd(),
		newCheckCmd(environ, reg, os.Stdout, os.Stderr),
		newDoctorCmd(environ, reg, os.Stdout, os.Stderr),
		newServeCmd(environ, reg, os.Stdout, os.Stderr),
		operator.newMemberCmd(environ, reg, os.Stdout, os.Stderr),
		operator.newSettingsCmd(environ, reg, os.Stdout, os.Stderr),
		newBackupCmd(environ, reg, os.Stdout, os.Stderr),
		newRestoreCmd(environ, reg, os.Stdout, os.Stderr),
		newUnidecodeCmd(environ, os.Stdout, os.Stderr),
	)
	return cmd
}

// newVersionCmd prints the "version" row, then the "matcher-commit" row
// a deploy's smoke step asserts on.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the dubplate-server version",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			if err := printRow(out, "version", version); err != nil {
				return err
			}
			return printRow(out, "matcher-commit", matcherCommit)
		},
	}
}
