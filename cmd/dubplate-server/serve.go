package main

import (
	"errors"
	"log/slog"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/spf13/cobra"
)

// newServeCmd builds the serve verb: it loads the wall through
// [loadConfig], builds the shell with newServeShell, and runs it until
// cmd's context is done, holding the store open until then. Every
// resolution and store open happens before [serve.Run] opens a
// listener, so an unsatisfied wall refuses at startup in the words
// doctor reports.
func newServeCmd(environ []string, reg *config.Registry) *cobra.Command {
	return &cobra.Command{
		Use:   "serve",
		Short: "Run the HTTP server",
		RunE: func(cmd *cobra.Command, _ []string) (err error) {
			cfg, err := loadConfig(environ, reg)
			if err != nil {
				return err
			}
			listen, err := serve.ParseConfig(environ)
			if err != nil {
				return err
			}
			srv, err := newServeShell(cmd.Context(), environ, reg, listen.Prefix, listen.BindAddrs, cfg, slog.Default())
			if err != nil {
				return err
			}
			defer func() { err = errors.Join(err, srv.Close()) }()
			return serve.Run(cmd.Context(), srv.shell.Logger(), srv.shell.Handler(), listen.BindAddrs, listen.Port)
		},
	}
}
