package main

import (
	"fmt"
	"io"

	"github.com/glw907/dubplate/internal/config"
	"github.com/spf13/cobra"
)

// newCheckCmd builds the check subcommand: it validates configuration
// loaded from environ, printing "configuration: ok" to stdout on
// success. On failure it returns the load error unprinted: main is
// the sole writer of command errors to stderr, so a failed check in a
// systemd journal is the error alone, never doubled.
func newCheckCmd(environ []string, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "check",
		Short:         "Validate configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
			if _, err := config.Load(environ); err != nil {
				return err
			}
			_, err := fmt.Fprintln(stdout, "configuration: ok")
			return err
		},
	}
	cmd.SetErr(stderr)
	return cmd
}
