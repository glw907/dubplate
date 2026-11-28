package main

import (
	"io"
	"io/fs"
	"log/slog"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/internal/webapp"
	"github.com/glw907/dubplate/web"
)

// webBuild is the build tree serve serves the web routes from and
// doctor checks them against.
var webBuild fs.FS = web.Build

// webDoctorRows renders the web row: "ok" when [newWebRoutes] builds
// the routes serve would, and a failing "invalid: " plus the refusal
// otherwise. cfg is nil when [config.Load] refused, and the row reads
// "not checked" without failing, since Load's report already fails
// doctor. Any record the build logs belongs to serve, so it is
// discarded.
func webDoctorRows(environ []string, cfg *config.Config) []doctor.Row {
	row := doctor.Row{Name: "web", Status: "not checked"}
	if cfg != nil {
		if err := checkWebRoutes(environ, cfg); err != nil {
			row.Status, row.Failed = "invalid: "+strings.ReplaceAll(err.Error(), "\n", "; "), true
		} else {
			row.Status = "ok"
		}
	}
	return []doctor.Row{row}
}

func checkWebRoutes(environ []string, cfg *config.Config) error {
	prefix, err := serve.ParsePrefix(environ)
	if err != nil {
		return err
	}
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	_, err = newWebRoutes(environ, prefix, cfg.SiteName, cfg.StateDir, discard, webBuild)
	return err
}

// newWebRoutes builds the web app's routes from build before any
// listener opens, so a bad value or a build defect stops startup rather
// than the first request.
func newWebRoutes(environ []string, prefix serve.Prefix, siteName, stateDir string, logger *slog.Logger, build fs.FS) (webapp.Routes, error) {
	cfg, err := webapp.FromEnviron(environ, stateDir)
	if err != nil {
		return webapp.Routes{}, err
	}
	cfg.Prefix = prefix.String()
	cfg.SiteName = siteName
	cfg.Logger = logger

	return webapp.New(build, cfg)
}
