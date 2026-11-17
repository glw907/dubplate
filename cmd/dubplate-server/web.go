package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"regexp"
	"strings"

	"github.com/glw907/dubplate/internal/config"
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

// acceptedThemeMatch is the one value themeMatchThreshold accepts: the
// match axis names a group of themes rather than a single one, and
// dubplate ships one group.
const acceptedThemeMatch = "dubplate"

// themePaletteGrammarRE is the shape [themePaletteThreshold] admits
// before its value is looked up in the bundle's own extracted
// selectors: refusing everything outside it first means a value
// carrying a quote, an angle bracket, or whitespace never reaches a
// selector-set lookup or an attribute.
var themePaletteGrammarRE = regexp.MustCompile(`^[a-z0-9-]+$`)

// notConfigured is the value DUBPLATE_BRANDING and
// DUBPLATE_OPERATOR_CSS_FILE default to, meaning the operator set
// nothing. The registry refuses an empty string default, so the two
// optional keys spell "unset" as a word. It cannot shadow a real
// operator file, because an operator file's path must end in ".css".
const notConfigured = "none"

// privateFiles names the state the operator CSS file must never be:
// anything under the state directory. The operator route is public, so
// a path under it would publish the store. An empty field guards
// nothing.
type privateFiles struct {
	stateDir string
}

// themeMatchThreshold is DUBPLATE_THEME_MATCH: the theme axis naming
// which group of themes the server presents. The server ships one
// group, so the only accepted value is acceptedThemeMatch.
var themeMatchThreshold = config.Threshold{
	Key:         "DUBPLATE_THEME_MATCH",
	Kind:        config.KindString,
	Default:     acceptedThemeMatch,
	Doc:         `exactly "dubplate"`,
	Description: "the theme group the server presents",
	Mark:        config.NotSensitive,
}

// themePaletteThreshold is DUBPLATE_THEME_PALETTE: the theme a visitor
// without a theme cookie sees. Its Doc string is the operator-facing
// text this key's own doctor row and the config wall report.
var themePaletteThreshold = config.Threshold{
	Key:         "DUBPLATE_THEME_PALETTE",
	Kind:        config.KindString,
	Default:     "system",
	Doc:         `The theme a visitor without a theme cookie sees. "system" follows the device's light or dark preference between "label" and "label-night", and falls back to "label"; a palette name pins that palette.`,
	Description: "the theme a visitor with no saved choice sees",
	Mark:        config.NotSensitive,
}

// brandingThreshold is DUBPLATE_BRANDING: the operator's declaration
// list, rendered once at startup into the branding stylesheet.
var brandingThreshold = config.Threshold{
	Key:         "DUBPLATE_BRANDING",
	Kind:        config.KindString,
	Default:     notConfigured,
	Doc:         `"none", or ";"-separated name: value declarations`,
	Description: "the operator's style declarations, applied to every page",
	Mark:        config.NotSensitive,
}

// operatorCSSFileThreshold is DUBPLATE_OPERATOR_CSS_FILE: the
// operator's own stylesheet, read once at startup by
// [webapp.ReadOperatorCSS].
var operatorCSSFileThreshold = config.Threshold{
	Key:         "DUBPLATE_OPERATOR_CSS_FILE",
	Kind:        config.KindString,
	Default:     notConfigured,
	Doc:         `"none", or a path to a .css file`,
	Description: "the path to the operator's own stylesheet",
	Mark:        config.Sensitive,
}

// registerWebConfig adds the web surface's four presentation
// thresholds to reg. newProductionRegistry calls it after
// registerStoreConfig.
func registerWebConfig(reg *config.Registry) {
	reg.AddThreshold(themeMatchThreshold)
	reg.AddThreshold(themePaletteThreshold)
	reg.AddThreshold(brandingThreshold)
	reg.AddThreshold(operatorCSSFileThreshold)
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
	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		return webapp.Routes{}, fmt.Errorf("web: %w", err)
	}
	return newWebRoutesFromFS(environ, prefix, siteName, private, logger, buildFS)
}

// newWebRoutesFromFS is [newWebRoutes] over buildFS, so a test can
// build any state a bad build can be in, since [web.Assets] is fixed
// at compile time. It resolves the four web thresholds and hands the
// values to [webapp.New].
func newWebRoutesFromFS(environ []string, prefix serve.Prefix, siteName string, private privateFiles, logger *slog.Logger, buildFS fs.FS) (webapp.Routes, error) {
	match, err := config.ParseString(environ, themeMatchThreshold)
	if err != nil {
		return webapp.Routes{}, err
	}
	if match != acceptedThemeMatch {
		return webapp.Routes{}, fmt.Errorf("%s: %q: must be exactly %q", themeMatchThreshold.Key, match, acceptedThemeMatch)
	}
	palette, err := config.ParseString(environ, themePaletteThreshold)
	if err != nil {
		return webapp.Routes{}, err
	}
	if !themePaletteGrammarRE.MatchString(palette) {
		return webapp.Routes{}, fmt.Errorf("%s: %q: must match %s", themePaletteThreshold.Key, palette, themePaletteGrammarRE)
	}
	branding, brandNeutral, err := resolveBranding(environ)
	if err != nil {
		return webapp.Routes{}, err
	}
	operatorCSS, err := resolveOperatorCSS(environ, private)
	if err != nil {
		return webapp.Routes{}, err
	}

	routes, err := webapp.New(buildFS, webapp.Config{
		Prefix:          prefix.String(),
		Names:           webRouteNames,
		SiteName:        siteName,
		Match:           match,
		Palette:         palette,
		Branding:        branding,
		BrandingNeutral: brandNeutral,
		OperatorCSS:     operatorCSS,
		Logger:          logger,
	})
	if errors.Is(err, webapp.ErrNoThemeSelector) {
		return webapp.Routes{}, fmt.Errorf("%s: %q: %w", themePaletteThreshold.Key, palette, err)
	}
	return routes, err
}

// resolveBranding renders DUBPLATE_BRANDING into the branding
// stylesheet, empty when the key is unset, empty, or "none", and
// returns the value it declares for --color-neutral, "" when it
// declares none.
func resolveBranding(environ []string) ([]byte, string, error) {
	decls, err := config.ParseString(environ, brandingThreshold)
	if err != nil {
		return nil, "", err
	}
	if decls == notConfigured {
		decls = ""
	}
	css, err := webapp.BrandingStylesheet(decls)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", brandingThreshold.Key, err)
	}
	return css, brandingNeutral(decls), nil
}

// brandingNeutral returns the value decls declares for
// --color-neutral, or "" when it declares none. decls must already
// have passed [webapp.BrandingStylesheet], which fixes every
// declaration's shape and admits each name once.
func brandingNeutral(decls string) string {
	for decl := range strings.SplitSeq(decls, ";") {
		name, value, _ := strings.Cut(decl, ":")
		if strings.TrimSpace(name) == "--color-neutral" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// resolveOperatorCSS reads the file DUBPLATE_OPERATOR_CSS_FILE names,
// returning nil when the key is unset, empty, or "none".
func resolveOperatorCSS(environ []string, private privateFiles) ([]byte, error) {
	path, err := config.ParseString(environ, operatorCSSFileThreshold)
	if err != nil {
		return nil, err
	}
	if path == notConfigured {
		return nil, nil
	}
	css, err := webapp.ReadOperatorCSS(path, private.stateDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operatorCSSFileThreshold.Key, err)
	}
	return css, nil
}
