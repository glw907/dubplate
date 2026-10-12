package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/glw907/dubplate/internal/auth"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/settings"
	"github.com/spf13/cobra"
)

// newSettingsCmd builds the settings verb, whose one subcommand is
// reset. Like member, it resolves configuration through config.Load
// against reg and reads the environment only through environ.
func newSettingsCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Change the settings the store holds",
	}
	cmd.AddCommand(newSettingsResetCmd(environ, reg, stdout, stderr))
	return cmd
}

// newSettingsResetCmd builds "settings reset <key>": it deletes an
// administrator's value so the setting reads its default again, from
// the command line, with serve stopped or running. It is the way out
// of a refusal the seed rule raises over an administrator's row, which
// no start clears by itself. It opens the store in write mode, so it
// refuses an older store, a newer store, and a restore in progress as
// member does; it runs under the validator the API's writes run
// under, and it sends no notice, since the operator is at the box.
func newSettingsResetCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "reset <key>",
		Short:         "Delete an administrator's value so the setting returns to its default",
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			ctx, cancel := context.WithTimeout(cmd.Context(), operatorVerbWriteWait)
			defer cancel()
			h, _, err := openOperatorStore(ctx, environ, reg)
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, h.Close()) }()

			by := settings.Author{
				Actor:   auth.OperatorActor(resolveActor()).String(),
				Channel: identity.ChannelCommandLine.String(),
			}
			key := args[0]
			res, err := settings.Reset(ctx, h, reg, environ, by, nil, key, time.Now())
			if err != nil {
				return err
			}
			return printResetResult(stdout, key, res)
		},
	}
	cmd.SetErr(stderr)
	return cmd
}

// printResetResult writes the one line a reset answers with: what the
// reset did, or why it did nothing, and the value now in force.
func printResetResult(w io.Writer, key string, res settings.Result) error {
	var err error
	if res.Written {
		_, err = fmt.Fprintf(w, "%s reset: %q (source %s, %s) is in force, replacing %q\n", key, res.Value, res.Source, res.Apply, res.Previous)
	} else {
		_, err = fmt.Fprintf(w, "%s: no administrator's value to reset; %q (source %s, %s) is in force\n", key, res.Value, res.Source, res.Apply)
	}
	return err
}
