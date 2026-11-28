package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
	"github.com/spf13/cobra"
)

// unidecodeInstallTimeout bounds the install verb's wheel download.
const unidecodeInstallTimeout = 30 * time.Second

// newUnidecodeCmd builds the unidecode subcommand's parent: today it
// carries the install verb alone.
func newUnidecodeCmd(environ []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unidecode",
		Short: "Manage the pinned Unidecode transliteration table",
	}
	cmd.AddCommand(newUnidecodeInstallCmd(environ))
	return cmd
}

// newUnidecodeInstallCmd builds the install verb: it reads
// DUBPLATE_STATE_DIR alone, never config.Load, because setup runs it
// before the rest of the wall may be rendered.
func newUnidecodeInstallCmd(environ []string) *cobra.Command {
	var from string
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install the pinned Unidecode transliteration table",
		Long: fmt.Sprintf(
			"Downloads Unidecode %s's pinned wheel from %s (sha256 %s) and\n"+
				"builds its transliteration table into the state directory. Offline,\n"+
				"download the wheel yourself and pass --from its path.",
			unidecode.Version, unidecode.WheelURL, unidecode.WheelSHA256),
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := runUnidecodeInstall(cmd.Context(), environ, from, cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("unidecode install: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "install from a local wheel file instead of downloading")
	return cmd
}

// runUnidecodeInstall reads DUBPLATE_STATE_DIR and installs the table
// into it, downloading over a client with unidecodeInstallTimeout when
// from names no local wheel.
func runUnidecodeInstall(ctx context.Context, environ []string, from string, stdout io.Writer) error {
	stateDir, ok := config.Lookup(environ, "DUBPLATE_STATE_DIR")
	if !ok {
		return errors.New("DUBPLATE_STATE_DIR is not set")
	}
	return unidecode.Install(ctx, stateDir, from, stdout, unidecode.InstallOptions{
		Client:  &http.Client{Timeout: unidecodeInstallTimeout},
		Geteuid: os.Geteuid,
	})
}
