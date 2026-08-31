package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/beets"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/importer"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/ping"
)

func newImportCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "import",
		Short: "Stage, gate, and beet-import each contributor's quiescent uploads",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			return runImport(ctx)
		},
	}
}

// runImport loads config, wires every collaborator package into an
// [importer.Importer], and runs it. It is the only place the import
// command's dependencies are assembled, so cmd/ stays wiring, not logic.
func runImport(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	roster, err := config.LoadContributors(cfg.Contributors)
	if err != nil {
		return fmt.Errorf("load contributors: %w", err)
	}

	im := &importer.Importer{
		Cfg:    cfg,
		Roster: roster,
		Ping:   ping.NewClient(),
		Mailer: notify.New(cfg.SMTP),
		Beets: &beets.Runner{
			BeetPath:   "beet",
			ConfigPath: cfg.BeetsConfig,
			StateDir:   cfg.Paths.StateDir,
		},
		FlacPath:     "flac",
		MetaflacPath: "metaflac",
		SqlitePath:   "sqlite3",
		Now:          time.Now,
	}
	return im.Run(ctx)
}
