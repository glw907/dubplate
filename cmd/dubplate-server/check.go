package main

import (
	"fmt"
	"io"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/spf13/cobra"
)

// newCheckCmd builds the check subcommand: it validates configuration
// loaded from environ, printing "configuration: ok" to stdout on
// success, then runs the store's inspect mode and prints its "store"
// row the same way doctor does. On failure it returns the failing
// error unprinted: main is the sole writer of command errors to
// stderr, so a failed check in a systemd journal is the error alone,
// never doubled.
func newCheckCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "check",
		Short:         "Validate configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := config.Load(environ, reg)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(stdout, "configuration: ok"); err != nil {
				return err
			}

			// The busy timeout's range is checked before the inspect
			// mode runs, naming the key and the bound: a threshold's
			// own value resolves at its point of use rather than at
			// Load, so config.Load succeeding above says nothing
			// about it yet.
			d, parseErr, rangeErr := storeBusyTimeout(environ)
			if parseErr != nil {
				return parseErr
			}
			if rangeErr != nil {
				return rangeErr
			}

			r := store.Inspect(cmd.Context(), store.InspectOptions{StateDir: cfg.StateDir, BusyTimeout: d})
			status, failed := storeRowStatus(r)
			if _, err := fmt.Fprintf(stdout, "store: %s\n", status); err != nil {
				return err
			}
			if failed {
				return fmt.Errorf("store: %s", status)
			}
			return nil
		},
	}
	cmd.SetErr(stderr)
	return cmd
}
