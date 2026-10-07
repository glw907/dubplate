package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/glw907/dubplate/internal/harness/subsonic/fixture"
	"github.com/glw907/dubplate/internal/harness/subsonic/stack"
	"github.com/spf13/cobra"
)

// newOracleUpCmd builds the oracle-up subcommand: it builds the streamer
// fixture library, starts the pinned Navidrome and gonic oracles on it,
// prints their loopback base URLs and the run's credentials to standard
// output, and holds until interrupted, then removes the containers. The
// credentials go nowhere else, since a roster walk types them into a
// client.
func newOracleUpCmd() *cobra.Command {
	var manifest, digests string

	cmd := &cobra.Command{
		Use:           "oracle-up",
		Short:         "Run the pinned Navidrome and gonic oracles on the fixture library until interrupted",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runOracleUp(ctx, cmd, manifest, digests)
		},
	}
	cmd.Flags().StringVar(&manifest, "manifest", "", "fixture manifest (default: the committed one)")
	cmd.Flags().StringVar(&digests, "digests", "", "oracle image digest file (default: the committed one)")
	return cmd
}

func runOracleUp(ctx context.Context, cmd *cobra.Command, manifest, digests string) error {
	if manifest == "" || digests == "" {
		p, err := defaultSubsonicPaths()
		if err != nil {
			return err
		}
		manifest = cmp.Or(manifest, p.manifest)
		digests = cmp.Or(digests, p.digests)
	}
	lib, err := fixture.Load(manifest)
	if err != nil {
		return err
	}
	music, err := os.MkdirTemp("", "dubplate-oracle-music-")
	if err != nil {
		return fmt.Errorf("create library directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(music) }()
	if _, err := lib.Build(music); err != nil {
		return err
	}

	s, err := stack.Start(ctx, stack.Config{MusicDir: music, Tracks: len(lib.Tracks()), DigestFile: digests})
	if err != nil {
		return err
	}
	defer s.Close(ctx)

	out := cmd.OutOrStdout()
	for _, o := range s.Oracles {
		if _, err := fmt.Fprintf(out, "%-10s %s  user %s  password %s\n", o.Name, o.BaseURL, o.User, o.Password); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(cmd.ErrOrStderr(), "oracles ready; interrupt to stop"); err != nil {
		return err
	}
	<-ctx.Done()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}
	return ctx.Err()
}
