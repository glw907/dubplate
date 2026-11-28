// Command dubplate-server organizes and serves one shared music library.
//
// Its subcommands:
//
//   - version: print the build version and matcher commit
//   - check: validate configuration, exit nonzero on failure
//   - doctor: report the full config wall's effective values
//   - serve: run the HTTP server
//   - member: manage the group's members
//   - settings reset: return a setting the store holds to its default
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

// version and matcherCommit are set with -ldflags -X; only
// scripts/deploy.sh sets matcherCommit.
var (
	version       = "dev"
	matcherCommit = "dev"
)

func main() {
	os.Exit(run(newRootCmd(os.Environ(), newProductionRegistry())))
}

// run executes cmd under the signal-derived context and returns the
// process's exit status: 0 for a nil error, the code an *exitError
// names, and 1 for any other error, which it prints.
func run(cmd *cobra.Command) int {
	err := executeWithSignalContext(cmd)
	if err == nil {
		return 0
	}
	fmt.Fprintln(os.Stderr, err)
	if ee, ok := errors.AsType[*exitError](err); ok {
		return ee.code
	}
	return 1
}

// storeWait bounds each store call a verb or the serve seed makes, since
// [store.Handle.Update] refuses a context with no deadline. The busy
// timeout, not this wait, bounds another process's lock.
const storeWait = 30 * time.Second

const (
	exitUnavailable = 69  // EX_UNAVAILABLE: backup integration is disabled
	exitTempFail    = 75  // EX_TEMPFAIL: another backup or a restore holds the lock
	exitInterrupted = 130 // 128+SIGINT: canceled by a signal or another process
)

// An exitError wraps an error with the process exit status its
// outcome carries, so run never matches on error text.
type exitError struct {
	code int
	err  error
}

func (e *exitError) Error() string { return e.err.Error() }
func (e *exitError) Unwrap() error { return e.err }

// executeWithSignalContext runs cmd under a context canceled by SIGINT
// or SIGTERM, without which the signal's default action would kill
// serve before its graceful drain.
func executeWithSignalContext(cmd *cobra.Command) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return cmd.ExecuteContext(ctx)
}

// newProductionRegistry registers every package in the order the
// registry dump holds, which places serve's thresholds and its trusted
// proxy class apart.
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

// loadConfig is [loadWall] returning a config only when every check
// holds, so every verb but doctor refuses what serve refuses.
func loadConfig(environ []string, reg *config.Registry) (*config.Config, error) {
	cfg, err := loadWall(environ, reg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// loadWall is config.Load plus the staging and port checks no single
// registration can express. The staging check needs Load's config; the
// port check runs either way, so doctor reports a bad port beside every
// other refusal. cfg is Load's result even when a later check refuses.
func loadWall(environ []string, reg *config.Registry) (*config.Config, error) {
	cfg, err := config.Load(environ, reg)
	if err == nil {
		err = intake.ValidateStaging(cfg)
	}
	return cfg, errors.Join(err, serve.ValidatePort(environ))
}

// newRootCmd builds the command tree over environ and reg. Its silence
// covers every subcommand, so run is the only printer of an error.
func newRootCmd(environ []string, reg *config.Registry) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "dubplate-server",
		Short:         "organizes and serves one shared music library",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(
		newVersionCmd(),
		newCheckCmd(environ, reg),
		newDoctorCmd(environ, reg),
		newServeCmd(environ, reg),
		newMemberCmd(environ, reg),
		newSettingsCmd(environ, reg),
		newBackupCmd(environ, reg),
		newRestoreCmd(environ, reg),
		newUnidecodeCmd(environ),
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
