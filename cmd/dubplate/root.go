package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// signalContext derives cmd's context into one canceled on SIGINT or
// SIGTERM, so a systemd stop or an operator's Ctrl-C reaches every
// exec.CommandContext subprocess a run has open. Every subcommand that
// does real work starts this way; the caller defers the cancel.
func signalContext(cmd *cobra.Command) (context.Context, context.CancelFunc) {
	return signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "dubplate",
		Short:         "The dubplate estate binary: import, backup, checks, and doctor",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(newVersionCmd())
	cmd.AddCommand(newImportCmd())
	cmd.AddCommand(newBackupCmd())
	cmd.AddCommand(newBeetCmd())
	cmd.AddCommand(newPingFailCmd())
	cmd.AddCommand(newRestoreCmd())
	cmd.AddCommand(newCheckCmd())
	cmd.AddCommand(newDoctorCmd())

	return cmd
}
