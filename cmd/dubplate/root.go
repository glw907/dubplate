package main

import "github.com/spf13/cobra"

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
