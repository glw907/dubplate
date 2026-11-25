package main

import (
	"errors"
	"fmt"
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

// The subtree names beside the ones in serve.go: the embedded faces,
// the favicon in its two forms, and the directory holding the mark's
// three detail levels.
const (
	fontsRouteName      = "fonts"
	faviconSVGRouteName = "favicon.svg"
	faviconPNGRouteName = "favicon.png"
	markRouteName       = "mark"
)

// webRouteNames are the subtree names [webapp.New] builds its routes
// under, the same names [registerWebRoutes] mounts them at.
var webRouteNames = webapp.RouteNames{
	Login:          loginRouteName,
	Members:        membersRouteName,
	ServerSettings: serverSettingsRouteName,
	Assets:         assetsRouteName,
	Fonts:          fontsRouteName,
	FaviconSVG:     faviconSVGRouteName,
	FaviconPNG:     faviconPNGRouteName,
	Mark:           markRouteName,
	Branding:       brandingRouteName,
	OperatorCSS:    operatorCSSRouteName,
}

// privateFiles names the state the operator CSS file must never be:
// anything under the state directory. The operator route is public, so
// a path under it would publish the store. An empty field guards
// nothing.
type privateFiles struct {
	stateDir string
}

// registerWebRoutes mounts wr's public routes on routes under their
// subtree names. It leaves members and server settings to the caller,
// which alone holds the admin gate.
func registerWebRoutes(routes registrar, wr webapp.Routes) {
	routes.Register(loginRouteName, wr.Login)
	routes.Register(assetsRouteName, wr.Assets)
	routes.Register(fontsRouteName, wr.Fonts)
	routes.Register(faviconSVGRouteName, wr.Icons)
	routes.Register(faviconPNGRouteName, wr.Icons)
	routes.Register(markRouteName, wr.Icons)
	routes.Register(brandingRouteName, wr.Branding)
	if wr.Operator != nil {
		routes.Register(operatorCSSRouteName, wr.Operator)
	}
}

// newWebRoutes builds the web app's routes from web.Assets, as
// [newWebRoutesFromFS] does from a given tree. It runs before any
// listener opens, so a bad config value or build defect stops startup
// rather than the first request. siteName enters every document as
// the dubplate-site-name meta; logger takes the documents' request
// records and the record a theme-color-less branding neutral writes.
func newWebRoutes(environ []string, prefix serve.Prefix, siteName string, private privateFiles, logger *slog.Logger) (webapp.Routes, error) {
	buildFS, err := embeddedBuild()
	if err != nil {
		return webapp.Routes{}, err
	}
	return newWebRoutesFromFS(environ, prefix, siteName, private, logger, buildFS)
}

// webDoctorRows renders the web row: "ok" when [newWebRoutes] builds
// the routes from the inputs serve startup passes, and "invalid: " plus
// the refusal otherwise, which fails the row. cfg is nil when
// [config.Load] refused the environment, and the row then reads "not
// checked" and adds no failure, since Load's report already fails
// doctor. The logger discards: the record newWebRoutes writes for a
// branding neutral that drops the theme-color meta belongs to serve,
// and doctor's stdout is the table alone.
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
	prefixValue, err := config.ParseString(environ, serve.PrefixThreshold)
	if err != nil {
		return err
	}
	prefix, err := serve.NewPrefix(prefixValue)
	if err != nil {
		return err
	}
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	_, err = newWebRoutes(environ, prefix, cfg.SiteName, privateFiles{stateDir: cfg.StateDir}, discard)
	return err
}

// embeddedBuild is [web.Assets] rooted at its build directory.
func embeddedBuild() (fs.FS, error) {
	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		return nil, fmt.Errorf("web: %w", err)
	}
	return buildFS, nil
}

// newWebRoutesFromFS is [newWebRoutes] over buildFS, so a test can
// build any state a bad build can be in, since [web.Assets] is fixed
// at compile time. It resolves the web settings with [webapp.FromEnviron]
// and hands them to [webapp.New].
func newWebRoutesFromFS(environ []string, prefix serve.Prefix, siteName string, private privateFiles, logger *slog.Logger, buildFS fs.FS) (webapp.Routes, error) {
	cfg, err := webapp.FromEnviron(environ, private.stateDir)
	if err != nil {
		return webapp.Routes{}, err
	}
	cfg.Prefix = prefix.String()
	cfg.Names = webRouteNames
	cfg.SiteName = siteName
	cfg.Logger = logger

	routes, err := webapp.New(buildFS, cfg)
	if errors.Is(err, webapp.ErrNoThemeSelector) {
		return webapp.Routes{}, fmt.Errorf("%s: %q: %w", webapp.ThemePaletteKey, cfg.Palette, err)
	}
	return routes, err
}
