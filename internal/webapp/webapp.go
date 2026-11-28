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

// RouteNames are the subtree names, under the owned prefix, the routes
// [New] builds answer at. FaviconSVG, FaviconPNG, and Mark also name
// the files and the directory the icons are read from in the build.
type RouteNames struct {
	Login          string
	Members        string
	ServerSettings string
	Assets         string
	Fonts          string
	FaviconSVG     string
	FaviconPNG     string
	Mark           string
	Branding       string
	OperatorCSS    string
}

// Config is what [New] builds the routes from. [FromEnviron] resolves
// and checks Match, Palette, Branding, BrandingNeutral, and
// OperatorCSS; the caller sets Prefix, Names, SiteName, and Logger.
// New checks only what it reads against the build.
type Config struct {
	// Prefix is the owned prefix every route answers under, as the
	// shell's prefix spells it.
	Prefix string

	// Names are the routes' subtree names.
	Names RouteNames

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

// New reads the build tree build, refuses on a placeholder-only tree
// naming `npm run build`, builds each document's CSP header from its
// own CSP meta, checks cfg.Palette against the bundle's own extracted
// [data-theme=<name>] selectors, rewrites the documents into one
// variant per stampable palette with the tier links and the head metas
// in place, and returns the routes. A document with an inline script
// its CSP does not hash, in the build or in any rewritten variant, a
// palette the bundle does not declare, a bundle theme with no usable
// neutral, or a missing build file is refused here, so a caller that
// builds the routes before any listener opens refuses startup rather
// than the first request.
func New(build fs.FS, cfg Config) (Routes, error) {
	placeholderOnly, err := isPlaceholderOnly(build)
	if err != nil {
		return Routes{}, fmt.Errorf("webapp: %w", err)
	}
	if placeholderOnly {
		return Routes{}, errors.New("the embedded web build tree holds only the placeholder; run `npm run build` in web/")
	}

	webDocuments := documents(cfg.Names)
	docs := make([]namedDocument, len(webDocuments))
	csps := make([]string, len(webDocuments))
	for i, d := range webDocuments {
		raw, err := fs.ReadFile(build, d.name)
		if err != nil {
			return Routes{}, fmt.Errorf("webapp: document %q: %w", d.name, err)
		}
		csp, err := documentCSP(d.name, raw)
		if err != nil {
			return Routes{}, err
		}
		docs[i] = namedDocument{path: d.name, raw: raw}
		csps[i] = csp
	}

	themes, err := extractThemes(build, docs...)
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
	icons := []icon{
		{Href: cfg.Prefix + "/" + cfg.Names.FaviconSVG, Type: "image/svg+xml"},
		{Href: cfg.Prefix + "/" + cfg.Names.FaviconPNG, Type: "image/png", Sizes: "32x32"},
	}
	handlers := make(map[string]http.Handler, len(webDocuments))
	for i, d := range webDocuments {
		vc := variantConfig{
			Placeholder: themePlaceholder,
			Match:       cfg.Match,
			Palettes:    palettes,
			TierHrefs:   tierHrefs(d.name, cfg.Names, hasOperatorCSS),
			SiteName:    cfg.SiteName,
			ThemeColors: colors,
			Icons:       icons,
		}
		if d.noIndex {
			vc.Robots = authRobots
		}
		variants, err := buildDocumentVariants(d.name, docs[i].raw, vc)
		if err != nil {
			return Routes{}, err
		}
		if err := checkVariantCSP(d.name, variants, csps[i]); err != nil {
			return Routes{}, err
		}
		handlers[cfg.Prefix+d.route] = newDocument(documentConfig{
			Variants:       variants,
			DefaultPalette: cfg.Palette,
			CacheControl:   d.cacheControl,
			CSP:            csps[i],
			Logger:         cfg.Logger,
		})
	}
	notFound := documentNotFound(cfg.Logger)
	membersPath := cfg.Prefix + "/" + cfg.Names.Members
	members := exactPaths(map[string]http.Handler{membersPath: handlers[membersPath]}, notFound)
	delete(handlers, membersPath)
	serverSettingsPath := cfg.Prefix + "/" + cfg.Names.ServerSettings
	serverSettings := exactPaths(map[string]http.Handler{serverSettingsPath: handlers[serverSettingsPath]}, notFound)
	delete(handlers, serverSettingsPath)
	login := exactPaths(handlers, notFound)

	assetsFS, err := fs.Sub(build, "_app")
	if err != nil {
		return Routes{}, fmt.Errorf("webapp: %w", err)
	}
	fontsFS, err := fs.Sub(build, "fonts")
	if err != nil {
		return Routes{}, fmt.Errorf("webapp: %w", err)
	}
	fontsRoute, err := newFontsRoute(fontsFS, cfg.Prefix+"/"+cfg.Names.Fonts)
	if err != nil {
		return Routes{}, err
	}
	iconsRoute, err := newIconsRoute(build, cfg.Prefix, cfg.Names)
	if err != nil {
		return Routes{}, err
	}
	routes := Routes{
		Login:          login,
		Members:        members,
		ServerSettings: serverSettings,
		Assets:         newAssets(assetsFS, cfg.Prefix+"/"+cfg.Names.Assets),
		Fonts:          fontsRoute,
		Icons:          iconsRoute,
		Branding:       stylesheetRoute(cfg.Prefix+"/"+cfg.Names.Branding, cfg.Branding),
	}
	if hasOperatorCSS {
		routes.OperatorCSS = stylesheetRoute(cfg.Prefix+"/"+cfg.Names.OperatorCSS, cfg.OperatorCSS)
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
