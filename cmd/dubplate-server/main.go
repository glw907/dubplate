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
	"syscall"

	"github.com/glw907/dubplate/internal/config"
	"github.com/spf13/cobra"
)

// version is set via -ldflags "-X main.version=..." at release build
// time; it stays "dev" for workstation builds.
var version = "dev"

// matcherCommit is the Go matcher revision this build carries, set via
// -ldflags "-X main.matcherCommit=..." at build time. It stays "dev"
// for workstation builds. The version verb prints it, so an operator
// can read which matcher revision a deployed binary carries.
//
// scripts/deploy.sh is the only build path that sets it, passing
// -X main.matcherCommit= through make build-linux's LDFLAGS variable;
// every other build (bare `go build`, `make build`, a bare `make
// build-linux`) leaves it "dev".
var matcherCommit = "dev"

func main() {
	os.Exit(run(newRootCmd()))
}

// run executes cmd under the signal-derived context and returns the
// process's exit status. A nil error is 0; an error wrapping an
// *exitCodeErr exits with the code it names (Ruling 3); every other
// error prints and exits 1. It is split out of main so a test can
// assert the status without calling os.Exit itself.
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

// An exitCodeErr wraps an error with the process exit status Ruling 3
// assigns its outcome: a verb whose failure is not a plain refusal
// (backup's unavailable, conflict, and canceled outcomes) names its
// own status here rather than run matching on error text.
type exitCodeErr struct {
	code int
	err  error
}

func (e *exitCodeErr) Error() string { return e.err.Error() }
func (e *exitCodeErr) Unwrap() error { return e.err }

// executeWithSignalContext runs cmd under a context canceled by SIGINT
// or SIGTERM. It is what makes the serve verb's graceful drain reachable
// at all: cobra's own Execute passes context.Background, whose Done
// channel is nil, so serve.Run would block on a select that can never
// fire and the process would die on the signal's default disposition
// with every in-flight request cut. Every other verb returns before a
// signal can arrive, so this costs them nothing.
//
// It is split out of main so a test can assert the wiring. main itself
// holds only the error print and the exit status, which a test cannot
// call.
func executeWithSignalContext(cmd *cobra.Command) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	return cmd.ExecuteContext(ctx)
}

// newProductionRegistry builds the registry the binary runs against:
// rung 1's own seed, plus every rung's own registration call, in
// landing order. newRootCmd and the test suite's own productionRegistry
// helper (serve_test.go) both call it, so the production registration
// order is written once rather than restated by hand in two places
// (merge carry 2).
func newProductionRegistry() *config.Registry {
	reg := config.NewRegistry()
	registerServeConfig(reg)
	registerAuthConfig(reg)
	registerStoreConfig(reg)
	registerWebConfig(reg)
	registerBackupConfig(reg)
	registerProxyConfig(reg)
	registerAPIConfig(reg)
	return reg
}

// newRootCmd builds the production registry once, visibly, and passes
// it to every subcommand that enforces against it: registration order
// stays readable here rather than hidden behind an init function, the
// explicit-registration rule stated in the Registry doc comment
// (internal/config/registry.go).
//
// ruled, conductor 2026-09-07, R1
func newRootCmd() *cobra.Command {
	reg := newProductionRegistry()

	cmd := &cobra.Command{
		Use:           "dubplate-server",
		Short:         "organizes and serves one shared music library",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(
		newVersionCmd(),
		newCheckCmd(os.Environ(), reg, os.Stdout, os.Stderr),
		newDoctorCmd(os.Environ(), reg, os.Stdout, os.Stderr),
		newServeCmd(os.Environ(), reg, os.Stdout, os.Stderr),
		newMemberCmd(os.Environ(), reg, os.Stdout, os.Stderr),
		newBackupCmd(os.Environ(), reg, os.Stdout, os.Stderr),
		newRestoreCmd(os.Environ(), reg, os.Stdout, os.Stderr),
		newUnidecodeCmd(os.Environ(), os.Stdout, os.Stderr),
	)
	return cmd
}

// newVersionCmd prints two labeled rows, "version" then
// "matcher-commit", in doctor.go's own printRow shape. The
// matcher-commit row is what a real deploy's smoke step asserts on.
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
