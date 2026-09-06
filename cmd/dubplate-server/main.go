// Command dubplate-server organizes and serves one shared music library.
//
// This doc comment carries the subcommand surface, re-derived from the
// organizer spec's two-binary split minus beet passthrough, which does
// not exist on the greenfield server. Each verb names the rung that
// lands it; a rung claiming a new verb amends this list.
//
//   - version: print the build version (rung 1)
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
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// version is set via -ldflags "-X main.version=..." at release build
// time; it stays "dev" for workstation builds.
var version = "dev"

func main() {
	cmd := newRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	registerShadowThresholds()

	cmd := &cobra.Command{
		Use:           "dubplate-server",
		Short:         "organizes and serves one shared music library",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.AddCommand(
		newVersionCmd(),
		newCheckCmd(os.Environ(), os.Stdout, os.Stderr),
		newDoctorCmd(os.Environ(), os.Stdout, os.Stderr),
	)
	return cmd
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the dubplate-server version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version)
			return err
		},
	}
}
