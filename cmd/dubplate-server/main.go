// Command dubplate-server organizes and serves one shared music library.
//
// This doc comment carries the subcommand surface, re-derived from the
// organizer spec's two-binary split minus beet passthrough, which does
// not exist on the greenfield server. Each verb names the rung that
// lands it; a rung claiming a new verb amends this list.
//
//   - version: print the build version and matcher commit (rung 1)
//   - check: validate configuration, exit nonzero on failure (rung 1)
//   - doctor: report the full config wall's effective values (rung 1)
//   - shadow: run the matcher against an existing library without writing (rung 4)
//   - serve: run the HTTP server (rung 5)
//   - backup: snapshot the library database and sync it to remote storage (rung 6)
//   - restore: restore a library snapshot (rung 6)
//   - import: admit a reviewed batch into the library (rung 9)
//   - ping-fail: report a failed run to a monitoring endpoint (rung 9)
//
// nudge-consume is not carried forward: its systemd units were retired
// in the 2026-09-02 wind-down, and no greenfield rung claims it.
package main

import (
	"context"
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
// for workstation builds, and shadow observe REFUSES to run on "dev"
// or on an empty value, because an unstamped divergence cannot be
// mechanically re-evaluated and is therefore worse than no
// observation.
//
// scripts/deploy.sh is the only build path that sets it, passing
// -X main.matcherCommit= through make build-linux's LDFLAGS variable;
// every other build (bare `go build`, `make build`, a bare `make
// build-linux`) leaves it "dev", so the refusal keeps firing for
// workstation builds.
var matcherCommit = "dev"

func main() {
	if err := executeWithSignalContext(newRootCmd()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

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
	registerShadowIntegration(reg)
	registerShadowThresholds(reg)
	registerServeConfig(reg)
	registerAuthConfig(reg)
	registerStoreConfig(reg)
	registerWebConfig(reg)
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
		newShadowCmd(os.Environ(), reg, matcherCommit, os.Stdout, os.Stderr),
		newServeCmd(os.Environ(), reg, os.Stdout, os.Stderr),
		newMemberCmd(os.Environ(), reg, os.Stdout, os.Stderr),
	)
	return cmd
}

// newVersionCmd prints two labeled rows, "version" then
// "matcher-commit", in doctor.go's own printRow shape. The
// matcher-commit row is what a real deploy's smoke step asserts on
// (Resolution R4): shadow observe refuses to run on a "dev" or empty
// matcherCommit, so an operator reading this output can see the same
// value the refusal checks before it ever reaches the box.
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
