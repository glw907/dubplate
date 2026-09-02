package main

import (
	"fmt"
	"io"

	"github.com/glw907/dubplate/internal/config"
	"github.com/spf13/cobra"
)

// newCheckCmd builds the check subcommand: it validates configuration
// loaded from environ, printing "configuration: ok" to stdout on
// success or the load error alone to stderr on failure. It never
// prints usage text, so a failed check in a systemd journal is the
// error alone.
func newCheckCmd(environ []string, stdout, stderr io.Writer) *cobra.Command {
	return &cobra.Command{
		Use:           "check",
		Short:         "Validate configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(*cobra.Command, []string) error {
			if _, err := config.Load(environ); err != nil {
				_, _ = fmt.Fprintln(stderr, err)
				return err
			}
			_, err := fmt.Fprintln(stdout, "configuration: ok")
			return err
		},
	}
}
