package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set via -ldflags "-X main.version=..." at release build time;
// it stays "dev" for workstation builds.
var version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the dubplate version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), version)
			return err
		},
	}
}
