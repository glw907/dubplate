package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
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
	built, _, err := buildCollections(root, cols)
	if err != nil {
		return err
	}

	s, err := stack.Start(ctx, stack.Config{Collections: built, DigestFile: digests})
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

// collection is one library a run serves: a name and the manifest it builds
// from.
type collection struct {
	name, manifest string
}

// defaultCollectionName names the one collection a --manifest run serves.
const defaultCollectionName = "music"

// parseCollections reads repeated name=manifest flag values. Whether a name
// is acceptable is the stack's to decide.
func parseCollections(values []string) ([]collection, error) {
	var cols []collection
	for _, v := range values {
		name, manifest, ok := strings.Cut(v, "=")
		if !ok || manifest == "" {
			return nil, fmt.Errorf("collection %q: want name=manifest", v)
		}
		cols = append(cols, collection{name: name, manifest: manifest})
	}
	return cols, nil
}

// resolveCollections returns the collection list and the digest file a run
// uses: the --collection values when given, otherwise one collection from
// manifest, with the committed manifest and digest file filling any gap.
func resolveCollections(manifest string, flags []string, digests string) ([]collection, string, error) {
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
		cols = []collection{{name: defaultCollectionName, manifest: manifest}}
	}
	return cols, digests, nil
}

// buildCollections builds each collection's library into its own directory
// under root and returns them with the digest of the whole list.
func buildCollections(root string, cols []collection) ([]stack.Collection, string, error) {
	built := make([]stack.Collection, 0, len(cols))
	digests := make([]fixture.NamedDigest, 0, len(cols))
	for i, c := range cols {
		lib, err := fixture.Load(c.manifest)
		if err != nil {
			return nil, "", fmt.Errorf("collection %q: %w", c.name, err)
		}
		dir := filepath.Join(root, strconv.Itoa(i))
		digest, err := lib.Build(dir)
		if err != nil {
			return nil, "", fmt.Errorf("collection %q: %w", c.name, err)
		}
		built = append(built, stack.Collection{Name: c.name, Dir: dir, Tracks: len(lib.Tracks())})
		digests = append(digests, fixture.NamedDigest{Name: c.name, Digest: digest})
	}
	return built, fixture.CombineDigests(digests), nil
}
