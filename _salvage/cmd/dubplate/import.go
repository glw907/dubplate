package main

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/beets"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/importer"
	"github.com/glw907/dubplate/internal/ledger"
	"github.com/glw907/dubplate/internal/notify"
	"github.com/glw907/dubplate/internal/ping"
)

func newImportCmd() *cobra.Command {
	var sessionID string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Stage, gate, and beet-import each contributor's quiescent uploads",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext(cmd)
			defer cancel()
			im, err := newImporter()
			if err != nil {
				return err
			}
			if sessionID != "" {
				return im.RunSession(ctx, sessionID)
			}
			return im.Run(ctx)
		},
	}
	cmd.Flags().StringVar(&sessionID, "session", "",
		"import exactly this promoted session's directory, bypassing quiescence for it")
	return cmd
}

// newImporter loads config and wires every collaborator package into an
// [importer.Importer]. It is the only place the import command's
// dependencies are assembled, so cmd/ stays wiring, not logic; both
// `dubplate import` and `dubplate nudge-consume` build their Importer
// through it.
func newImporter() (*importer.Importer, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	roster, err := config.LoadContributors(cfg.Contributors)
	if err != nil {
		return nil, fmt.Errorf("load contributors: %w", err)
	}
	led, err := ledger.Open(cfg.LedgerPath)
	if err != nil {
		return nil, fmt.Errorf("open ledger: %w", err)
	}

	return &importer.Importer{
		Cfg:    cfg,
		Roster: roster,
		Ping:   ping.NewClient(),
		Mailer: notify.New(cfg.SMTP),
		Beets: &beets.Runner{
			BeetPath:   "beet",
			ConfigPath: cfg.BeetsConfig,
			StateDir:   cfg.Paths.StateDir,
		},
		Ledger:       led,
		FlacPath:     "flac",
		MetaflacPath: "metaflac",
		SqlitePath:   "sqlite3",
		Now:          time.Now,
	}, nil
}
