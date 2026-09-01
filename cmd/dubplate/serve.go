package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/web"
)

func newServeCmd() *cobra.Command {
	var devIdentity string

	cmd := &cobra.Command{
		Use:   "serve",
		Short: "Serve the uploader: the embedded web app, the tus upload API, and the journey API",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := signalContext(cmd)
			defer cancel()
			return runServe(ctx, devIdentity)
		},
	}
	cmd.Flags().StringVar(&devIdentity, "dev-identity", "",
		"skip Access verification and treat every request as this email; refuses to start unless every bind address is loopback")
	return cmd
}

// runServe loads config, wires every collaborator package into a
// [serve.Server], and runs it until ctx is done.
func runServe(ctx context.Context, devIdentity string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	roster, err := config.LoadContributors(cfg.Contributors)
	if err != nil {
		return fmt.Errorf("load contributors: %w", err)
	}

	srv, err := serve.NewServer(cfg, roster, serve.NewServerDeps{
		Assets:      web.Assets,
		DevIdentity: devIdentity,
	})
	if err != nil {
		return fmt.Errorf("build server: %w", err)
	}
	return srv.Run(ctx)
}
