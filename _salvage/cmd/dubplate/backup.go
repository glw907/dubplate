package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/backup"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ping"
)

func newBackupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "backup",
		Short: "Snapshot the beets and Navidrome databases and sync the library to R2",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext(cmd)
			defer cancel()
			return runBackup(ctx)
		},
	}
}

// runBackup loads config and wires it into a [backup.Runner]. It is the
// only place the backup command's dependencies are assembled, so cmd/
// stays wiring, not logic.
func runBackup(ctx context.Context) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	r := &backup.Runner{
		Cfg:        cfg,
		Ping:       ping.NewClient(),
		SqlitePath: "sqlite3",
		RclonePath: "rclone",
		Now:        time.Now,
	}
	return r.Run(ctx)
}
