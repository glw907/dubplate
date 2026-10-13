package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/store"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/spf13/cobra"
)

// newCheckCmd builds the check subcommand: it validates configuration
// loaded from environ, printing "configuration: ok" to stdout on
// success, then runs the store's inspect mode and prints its "store"
// row the same way doctor does, then evaluates the seed rule
// read-only against the store's facts, an absent store's included, and
// prints one "settings:" line per non-quiet outcome. On failure
// it returns the failing error unprinted: main is the sole writer of
// command errors to stderr, so a failed check in a systemd journal is
// the error alone, never doubled.
func newCheckCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "check",
		Short:         "Validate configuration",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg, err := loadConfig(environ, reg)
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
			d, err := storeBusyTimeout(environ)
			if err != nil {
				return err
			}

			r := store.Inspect(cmd.Context(), store.InspectOptions{StateDir: cfg.StateDir, BusyTimeout: d})
			status, failed := storeRowStatus(r)
			if _, err := fmt.Fprintf(stdout, "store: %s\n", status); err != nil {
				return err
			}

			// The settings evaluation runs against the facts
			// seedFacts picks: an absent store, or one older than
			// the settings table, evaluates the empty table serve's
			// first seed meets, so a value that start would refuse
			// fails check first. A store this check cannot vouch for
			// already fails on the store row above, and printing a
			// settings verdict for one it never read would report on
			// rows it never saw.
			settingsFailed := false
			if facts, ok := seedFacts(r); ok {
				settingsFailed, err = printSettingsRows(stdout, environ, reg, facts)
				if err != nil {
					return err
				}
			}

			if failed {
				return fmt.Errorf("store: %s", status)
			}
			if settingsFailed {
				return errors.New("settings: the seed rule refuses to start serve")
			}
			return nil
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// printSettingsRows prints one "settings: <key>: <outcome>" line per
// non-quiet outcome the seed rule finds against facts: a refusal
// always prints, and an outcome that leaves the value in force
// unchanged never does, since check surfaces only what an operator
// need act on. It reports whether any outcome refuses.
func printSettingsRows(w io.Writer, environ []string, reg *config.Registry, facts store.Facts) (failed bool, err error) {
	ev := settings.EvaluateFacts(reg, environ, facts)
	for _, o := range ev.Outcomes {
		if o.Err == nil && o.Change == settings.Unchanged {
			continue
		}
		if o.Err != nil {
			failed = true
		}
		if _, err := fmt.Fprintf(w, "settings: %s: %s\n", o.Key, o.Report); err != nil {
			return failed, err
		}
	}
	return failed, nil
}
