// Package webapp builds the HTTP surface of the embedded web app: its
// documents with their Content-Security-Policy, its stylesheets with
// the theme and branding stamped in, its static files, fonts, and
// icons, and the exact-path router they sit behind.
//
// internal/serve is how a request reaches the process; this package is
// what the web app answers once it has. [New] takes resolved values and
// reads no configuration of its own, and no handler here gates a
// request: the caller wraps the members and server settings documents
// in the admin gate.
package webapp

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
)

// The subtree names, under the owned prefix, that the routes [New]
// builds answer at. The build fixes them: AssetsRouteName,
// FontsRouteName, and MarkRouteName also name build directories, and
// FaviconSVGRouteName and FaviconPNGRouteName build files, so the
// build's relative links resolve under these names alone. The server
// settings document is not "settings", a name kept for the member
// preferences page.
const (
	LoginRouteName          = "login"
	MembersRouteName        = "members"
	ServerSettingsRouteName = "server-settings"
	AssetsRouteName         = "_app"
	FontsRouteName          = "fonts"
	FaviconSVGRouteName     = "favicon.svg"
	FaviconPNGRouteName     = "favicon.png"
	MarkRouteName           = "mark"
	BrandingRouteName       = "branding.css"
	OperatorCSSRouteName    = "operator.css"
)

// Config is what [New] builds the routes from. [FromEnviron] resolves
// and checks Match, Palette, Branding, BrandingNeutral, and
// OperatorCSS; the caller sets Prefix, SiteName, and Logger. New
// checks only what it reads against the build.
type Config struct {
	// Prefix is the owned prefix every route answers under, as the
	// shell's prefix spells it.
	Prefix string

	// SiteName enters every document as the dubplate-site-name meta.
	SiteName string

	// Match names the theme group, and Palette the theme a visitor
	// without a theme cookie sees. Match+"-"+Palette must name a theme
	// the bundle declares.
	Match   string
	Palette string

	// Branding is the branding stylesheet rendered from
	// DUBPLATE_BRANDING, and BrandingNeutral the value it declares for
	// --color-neutral, "" when it declares none.
	Branding        []byte
	BrandingNeutral string

	// OperatorCSS is the stylesheet DUBPLATE_OPERATOR_CSS_FILE names, as
	// FromEnviron reads it, or nil when none is configured.
	OperatorCSS []byte

	// Logger takes the documents' request records and the one record
	// written when the branding neutral leaves the documents without a
	// theme-color meta.
	Logger *slog.Logger
}

// Routes are the http.Handlers [New] builds, each answering under its
// own subtree name. OperatorCSS is nil when no operator stylesheet is
// configured, which leaves its path to the shell's 404. Members and
// ServerSettings are documents with no gate of their own, so neither
// is ever mounted except behind the admin gate.
type Routes struct {
	Login          http.Handler
	Members        http.Handler
	ServerSettings http.Handler
	Assets         http.Handler
	Fonts          http.Handler
	Icons          http.Handler
	Branding       http.Handler
	OperatorCSS    http.Handler
}

// New reads the build tree build and returns the routes. It builds
// each document's CSP header from the document's own CSP meta. It
// checks cfg.Palette against the [data-theme=<name>] selectors the
// bundle declares. It rewrites each document into one variant per
// stampable palette, with the tier links and the head metas in place.
//
// New refuses a placeholder-only tree, naming `npm run build`. It also
// refuses a missing build file, a palette the bundle does not declare,
// a bundle theme with no usable neutral, and an inline script a
// document's CSP does not hash, in the build or in any variant. A
// caller that builds the routes before any listener opens therefore
// refuses startup rather than the first request.
func New(build fs.FS, cfg Config) (Routes, error) {
	placeholderOnly, err := isPlaceholderOnly(build)
	if err != nil {
		return Routes{}, fmt.Errorf("webapp: %w", err)
	}
	if placeholderOnly {
		return Routes{}, errors.New("the embedded web build tree holds only the placeholder; run `npm run build` in web/")
	}

	docs := documents()
	for i := range docs {
		d := &docs[i]
		d.raw, err = fs.ReadFile(build, d.name)
		if err != nil {
			return Routes{}, fmt.Errorf("webapp: document %q: %w", d.name, err)
		}
		d.csp, err = documentCSP(d.name, d.raw)
		if err != nil {
			return Routes{}, err
		}
	}

	themes, err := extractThemes(build, docs)
	if err != nil {
		return Routes{}, err
	}
	composed := cfg.Match + "-" + cfg.Palette
	if _, ok := themes.selectors[composed]; !ok {
		return Routes{}, fmt.Errorf("%s: %q: no theme selector named %q in the built bundle", themePaletteThreshold.Key, cfg.Palette, composed)
	}
	palettes := stampablePalettes(themes.selectors, cfg.Match)
	colors, err := themeColors(themes.neutrals, cfg.Match, palettes, cfg.BrandingNeutral, cfg.Logger)
	if err != nil {
		return Routes{}, err
	}

	hasOperatorCSS := cfg.OperatorCSS != nil
	var icons []icon
	for _, f := range favicons() {
		icons = append(icons, icon{href: cfg.Prefix + "/" + f.name, contentType: f.contentType, sizes: f.sizes})
	}
	var routes Routes
	groups := make(map[*http.Handler]map[string]http.Handler)
	for _, d := range docs {
		vc := variantConfig{
			match:       cfg.Match,
			palettes:    palettes,
			tierHrefs:   tierHrefs(d.name, hasOperatorCSS),
			siteName:    cfg.SiteName,
			themeColors: colors,
			icons:       icons,
		}
		if d.noIndex {
			vc.robots = authRobots
		}
		variants, err := buildDocumentVariants(d.name, d.raw, vc)
		if err != nil {
			return Routes{}, err
		}
		if err := checkVariantCSP(d.name, variants, d.csp); err != nil {
			return Routes{}, err
		}
		dst := d.dest(&routes)
		if groups[dst] == nil {
			groups[dst] = make(map[string]http.Handler)
		}
		groups[dst][cfg.Prefix+d.route] = documentHandler(documentConfig{
			variants:       variants,
			defaultPalette: cfg.Palette,
			cacheControl:   d.cacheControl,
			csp:            d.csp,
			logger:         cfg.Logger,
		})
	}
	notFound := documentNotFoundHandler(cfg.Logger)
	for dst, group := range groups {
		*dst = exactPaths(group, notFound)
	}

	assetsFS, err := fs.Sub(build, AssetsRouteName)
	if err != nil {
		return Routes{}, fmt.Errorf("webapp: %w", err)
	}
	fontsFS, err := fs.Sub(build, FontsRouteName)
	if err != nil {
		return Routes{}, fmt.Errorf("webapp: %w", err)
	}
	routes.Fonts, err = fontsHandler(fontsFS, cfg.Prefix+"/"+FontsRouteName)
	if err != nil {
		return Routes{}, err
	}
	routes.Icons, err = iconsHandler(build, cfg.Prefix)
	if err != nil {
		return Routes{}, err
	}
	routes.Assets = assetsHandler(assetsFS, cfg.Prefix+"/"+AssetsRouteName)
	routes.Branding = stylesheetHandler(cfg.Prefix+"/"+BrandingRouteName, cfg.Branding)
	if hasOperatorCSS {
		routes.OperatorCSS = stylesheetHandler(cfg.Prefix+"/"+OperatorCSSRouteName, cfg.OperatorCSS)
	}
	return routes, nil
}

// isPlaceholderOnly reports whether build's root directory holds
// nothing but placeholder.txt, the shape a checkout that has never run
// `npm run build` leaves (web/embed.go).
func isPlaceholderOnly(build fs.FS) (bool, error) {
	entries, err := fs.ReadDir(build, ".")
	if err != nil {
		return false, err
	}
	return len(entries) == 1 && entries[0].Name() == "placeholder.txt", nil
}

// exactPaths returns an http.Handler dispatching by exact
// http.Request.URL.Path to routes, answering notFound to any other
// path, the trailing-slash form of a route included. The shell mounts
// a subtree's every deeper path to the same handler, so this check is
// what keeps a deeper or trailing-slash path from reaching a route.
func exactPaths(routes map[string]http.Handler, notFound http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := routes[r.URL.Path]; ok {
			h.ServeHTTP(w, r)
			return
		}
		notFound.ServeHTTP(w, r)
	})
}
