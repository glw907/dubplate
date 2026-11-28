package main

import (
	"errors"
	"fmt"
	"io"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
	"github.com/spf13/cobra"
)

// newCheckCmd builds the check verb: it validates the configuration,
// prints the store row doctor prints, and evaluates the seed rule
// read-only against the store, printing one line per non-quiet outcome.
func newCheckCmd(environ []string, reg *config.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Validate configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			stdout := cmd.OutOrStdout()
			cfg, err := loadConfig(environ, reg)
			if err != nil {
				return err
			}
			if _, err := fmt.Fprintln(stdout, "configuration: ok"); err != nil {
				return err
			}

			// Load leaves thresholds to their point of use, so the busy
			// timeout's range is checked here.
			d, err := store.BusyTimeout(environ)
			if err != nil {
				return err
			}

			var stored settings.Facts
			r := store.Inspect(cmd.Context(), store.InspectOptions{
				StateDir:    cfg.StateDir,
				BusyTimeout: d,
				Readers:     []store.Reader{stored.Reader()},
			})
			row := r.Row()
			if _, err := fmt.Fprintf(stdout, "store: %s\n", row.Status); err != nil {
				return err
			}

			// An absent or older store evaluates the empty table serve's
			// first seed meets. A store the row above cannot vouch for
			// gets no settings verdict, since no row of it was read.
			settingsFailed := false
			if facts, ok := settings.SeedFacts(r, stored); ok {
				settingsFailed, err = printSettingsRows(stdout, environ, reg, facts)
				if err != nil {
					return err
				}
			}

			if row.Failed {
				return fmt.Errorf("store: %s", row.Status)
			}
			if settingsFailed {
				return errors.New("the settings seed rule refuses to start serve")
			}
			return nil
		},
	}
}

// printSettingsRows prints one "settings: <key>: <outcome>" line per
// outcome against facts that refuses or changes the value in force,
// and reports whether any refuses.
func printSettingsRows(w io.Writer, environ []string, reg *config.Registry, facts settings.Facts) (failed bool, err error) {
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
