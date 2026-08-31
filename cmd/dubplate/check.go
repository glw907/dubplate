package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/checks"
	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/ping"
)

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check <disk|navidrome>",
		Short: "Run one health check and ping its result",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt, syscall.SIGTERM)
			defer cancel()
			return runCheck(ctx, args[0])
		},
	}
}

// runCheck loads config and dispatches to the named check. An unknown
// name errors rather than a silent no-op, so a typoed systemd ExecStart
// cannot look healthy.
func runCheck(ctx context.Context, name string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	pc := ping.NewClient()

	switch name {
	case "disk":
		return (&checks.DiskChecker{Cfg: cfg, Ping: pc}).Run(ctx)
	case "navidrome":
		return (&checks.NavidromeChecker{Cfg: cfg, Ping: pc}).Run(ctx)
	default:
		return fmt.Errorf("check %q: want disk or navidrome", name)
	}
}
