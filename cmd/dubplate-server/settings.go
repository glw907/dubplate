package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/glw907/dubplate/internal/store"
	"github.com/spf13/cobra"
)

func newSettingsCmd(environ []string, reg *config.Registry) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Change the settings the store holds",
	}
	cmd.AddCommand(newSettingsResetCmd(environ, reg))
	return cmd
}

// newSettingsResetCmd builds "settings reset <key>", with serve stopped
// or running. It is the way out of a refusal the seed rule raises over
// an administrator's row, which no start clears by itself. It sends no
// notice, since the operator is at the box.
func newSettingsResetCmd(environ []string, reg *config.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "reset <key>",
		Short: "Delete an administrator's value so the setting returns to its default",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return withStore(cmd.Context(), environ, reg, func(ctx context.Context, h *store.Handle, _ string) error {
				by := identity.LocalActor(resolveActor())
				key := args[0]
				res, err := settings.Reset(ctx, h, reg, environ, by, nil, key, time.Now())
				if err != nil {
					return err
				}
				_, envSets := config.Lookup(environ, key)
				return printResetResult(cmd.OutOrStdout(), key, res, envSets)
			})
		},
	}
}

// printResetResult writes the one line a reset answers with: what the
// reset did, or why it did nothing, and the value now in force. When
// the environment sets the key, the default is not what the next start
// leaves in force, so a deleted row says that instead.
func printResetResult(w io.Writer, key string, res settings.Result, envSets bool) error {
	var err error
	if res.Written && envSets {
		_, err = fmt.Fprintf(w, "%s reset: the administrator's value %q is deleted; the environment sets %s, so serve seeds its value at the next start\n", key, res.Previous, key)
	} else if res.Written {
		_, err = fmt.Fprintf(w, "%s reset: %q (source %s, %s) is in force, replacing %q\n", key, res.Value, res.Source, res.Apply, res.Previous)
	} else {
		_, err = fmt.Fprintf(w, "%s: no administrator's value to reset; %q (source %s, %s) is in force\n", key, res.Value, res.Source, res.Apply)
	}
	return err
}
