package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strings"
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
	var collections []string

	cmd := &cobra.Command{
		Use:           "oracle-up",
		Short:         "Run the pinned Navidrome and gonic oracles on the fixture library until interrupted",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer stop()
			return runOracleUp(ctx, cmd, manifest, collections, digests)
		},
	}
	cmd.Flags().StringVar(&manifest, "manifest", "", "fixture manifest, the only collection (default: the committed one)")
	cmd.Flags().StringArrayVar(&collections, "collection", nil, "collection as name=manifest; repeat for several")
	cmd.MarkFlagsMutuallyExclusive("manifest", "collection")
	cmd.Flags().StringVar(&digests, "digests", "", "oracle image digest file (default: the committed one)")
	return cmd
}

func runOracleUp(ctx context.Context, cmd *cobra.Command, manifest string, collectionFlags []string, digests string) error {
	cols, digests, err := resolveCollections(manifest, collectionFlags, digests)
	if err != nil {
		return err
	}
	root, err := os.MkdirTemp("", "dubplate-oracle-music-")
	if err != nil {
		return fmt.Errorf("create library directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(root) }()
	built, _, err := fixture.BuildCollections(root, cols)
	if err != nil {
		return err
	}

	s, err := stack.Start(ctx, stack.Config{Collections: stackCollections(built), DigestFile: digests})
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

// parseCollections reads repeated name=manifest flag values. Whether a name
// is acceptable is the stack's to decide.
func parseCollections(values []string) ([]fixture.Collection, error) {
	var cols []fixture.Collection
	for _, v := range values {
		name, manifest, ok := strings.Cut(v, "=")
		if !ok || manifest == "" {
			return nil, fmt.Errorf("collection %q: want name=manifest", v)
		}
		cols = append(cols, fixture.Collection{Name: name, Manifest: manifest})
	}
	return cols, nil
}

// resolveCollections returns the collection list and the digest file a run
// uses: the --collection values when given, otherwise one collection from
// manifest, with the committed manifest and digest file filling any gap.
func resolveCollections(manifest string, flags []string, digests string) ([]fixture.Collection, string, error) {
	cols, err := parseCollections(flags)
	if err != nil {
		return nil, "", err
	}
	if digests == "" || (len(cols) == 0 && manifest == "") {
		p, err := defaultSubsonicPaths()
		if err != nil {
			return nil, "", err
		}
		digests = cmp.Or(digests, p.digests)
		manifest = cmp.Or(manifest, p.manifest)
	}
	if len(cols) == 0 {
		cols = []fixture.Collection{{Name: fixture.DefaultCollectionName, Manifest: manifest}}
	}
	return cols, digests, nil
}

// stackCollections returns the built collections as the stack mounts them.
func stackCollections(built []fixture.BuiltCollection) []stack.Collection {
	cols := make([]stack.Collection, len(built))
	for i, b := range built {
		cols[i] = stack.Collection{Name: b.Name, Dir: b.Dir, Tracks: len(b.Library.Tracks())}
	}
	return cols
}
