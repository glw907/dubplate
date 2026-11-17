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
	"maps"
	"net/http"
	"path"
	"regexp"
	"slices"
	"strings"
)

// ErrNoThemeSelector reports that the palette [Config] names has no
// [data-theme=<match>-<palette>] selector in the built bundle.
var ErrNoThemeSelector = errors.New("no theme selector")

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

// Config is what [New] builds the routes from. Every value is resolved
// and checked by the caller; New checks only what it reads against the
// build.
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

	// Branding is the branding stylesheet, as [BrandingStylesheet]
	// renders it, and BrandingNeutral the value it declares for
	// --color-neutral, "" when it declares none.
	Branding        []byte
	BrandingNeutral string

	// OperatorCSS is the operator's stylesheet, as [ReadOperatorCSS]
	// reads it, or nil when none is configured.
	OperatorCSS []byte

	// Logger takes the documents' request records and the one record
	// written when the branding neutral leaves the documents without a
	// theme-color meta.
	Logger *slog.Logger
}

// Routes are the http.Handlers [New] builds, each answering under its
// own subtree name. Operator is nil when no operator stylesheet is
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
	Operator       http.Handler
}

// themePlaceholder is the token app.html carries on <html data-theme>,
// substituted once per stampable palette at startup. It must match
// web/src/app.html's own literal exactly.
const themePlaceholder = "__DUBPLATE_THEME__"

// authRobots is the robots meta the two sign-in documents carry, so a
// search engine neither lists a sign-in page nor follows a link out of
// one.
const authRobots = "noindex, nofollow"

// themeColorRE is the shape a value must have before it enters a
// theme-color meta. A bundle's own --color-neutral outside it is a
// build defect to refuse, since the bundle is this server's own build;
// a branding neutral outside it drops the meta instead.
var themeColorRE = regexp.MustCompile(`^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`)

// document is one document the routes serve: its path inside the build
// tree, the path under the owned prefix it answers at, its
// Cache-Control value, and whether it carries the robots meta.
type document struct {
	name         string
	route        string
	cacheControl string
	noIndex      bool
}

// documents returns every document the build emits and the server
// serves. A document is no-cache by default: the stamp and the tier
// links are fixed at startup, so a cache may keep it and revalidate.
// The consumption document is no-store because its URL carried a
// sign-in token. The members and server settings documents are
// no-store because they answer only behind the admin gate, as the API
// they read from does.
func documents(names RouteNames) []document {
	return []document{
		{name: "login.html", route: "/" + names.Login, cacheControl: "no-cache", noIndex: true},
		{name: "login/link.html", route: "/" + names.Login + "/link", cacheControl: "no-store", noIndex: true},
		{name: "members.html", route: "/" + names.Members, cacheControl: "no-store"},
		{name: "server-settings.html", route: "/" + names.ServerSettings, cacheControl: "no-store"},
	}
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
		return Routes{}, fmt.Errorf("web: %w", err)
	}
	if placeholderOnly {
		return Routes{}, errors.New("web: the embedded build tree holds only the placeholder; run `npm run build` in web/")
	}

	webDocuments := documents(cfg.Names)
	docs := make([]namedDocument, len(webDocuments))
	csps := make([]string, len(webDocuments))
	for i, d := range webDocuments {
		raw, err := fs.ReadFile(build, d.name)
		if err != nil {
			return Routes{}, fmt.Errorf("web: document %q: %w", d.name, err)
		}
		csp, err := DocumentCSP(d.name, raw)
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
		return Routes{}, fmt.Errorf("%w named %q in the built bundle", ErrNoThemeSelector, composed)
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
		return Routes{}, fmt.Errorf("web: %w", err)
	}
	fontsFS, err := fs.Sub(build, "fonts")
	if err != nil {
		return Routes{}, fmt.Errorf("web: %w", err)
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
		routes.Operator = stylesheetRoute(cfg.Prefix+"/"+cfg.Names.OperatorCSS, cfg.OperatorCSS)
	}
	return routes, nil
}

// checkVariantCSP refuses when any served variant of the document
// name holds an inline script the build's own CSP meta does not hash,
// or yields a policy other than served, the header computed over the
// build's document. The rewrite is meant to touch no script and no
// meta, but a placeholder or a stylesheet-link anchor the build left
// inside a script body would change that script's hash, and the
// header would then block it in a browser; checking each served body
// turns that into a startup refusal.
func checkVariantCSP(name string, variants map[string][]byte, served string) error {
	for _, palette := range slices.Sorted(maps.Keys(variants)) {
		csp, err := DocumentCSP(name, variants[palette])
		if err != nil {
			return fmt.Errorf("web: the %q variant: %w", palette, err)
		}
		if csp != served {
			return fmt.Errorf("web: the %q variant of document %q: its CSP differs from the served header", palette, name)
		}
	}
	return nil
}

// stylesheetRoute serves css at exactly route and answers the shell's
// 404 everywhere under it.
func stylesheetRoute(route string, css []byte) http.Handler {
	return ExactPaths(map[string]http.Handler{route: newStylesheet(css)})
}

// themeColors returns each palette's theme-color metas. Each palette
// takes its composed theme's own --color-neutral from neutrals, as
// [paletteThemeColors] reads it, so a bundle theme with no usable
// neutral refuses startup whatever the branding says.
//
// A branding neutral flattens every palette, since the branding
// stylesheet overrides the variable in every theme, so it replaces
// them all with one meta at its value. The meta takes the value only
// when it matches themeColorRE, the same #rgb or #rrggbb grammar a
// bundle neutral meets. The branding grammar also admits a color
// function and a four- or eight-digit hex literal; either leaves every
// document with no theme-color meta and writes one record to logger,
// so the meta never carries an operator value outside that grammar and
// a configuration the branding grammar admits still starts.
func themeColors(neutrals map[string][]themeColor, match string, palettes []string, brandNeutral string, logger *slog.Logger) (map[string][]themeColor, error) {
	colors := make(map[string][]themeColor, len(palettes))
	for _, palette := range palettes {
		theme := match + "-" + palette
		tcs, err := paletteThemeColors(theme, neutrals[theme])
		if err != nil {
			return nil, err
		}
		colors[palette] = tcs
	}
	switch {
	case brandNeutral == "":
		return colors, nil
	case themeColorRE.MatchString(brandNeutral):
		for _, palette := range palettes {
			colors[palette] = []themeColor{{Color: brandNeutral}}
		}
		return colors, nil
	default:
		logger.Warn("web: DUBPLATE_BRANDING sets --color-neutral to a value other than a #rgb or #rrggbb color, so no document carries a theme-color meta; declare it as a three- or six-digit hex color to keep one")
		return nil, nil
	}
}

// paletteThemeColors returns the theme-color metas for theme, from the
// --color-neutral values its rules declare. Every value must match
// themeColorRE. One distinct value is one unconditional meta; one
// value under the light scheme and a different one under the dark is
// the media pair. Anything else refuses, naming the theme but never
// the value: no value at all, or values the two schemes cannot tell
// apart.
func paletteThemeColors(theme string, declared []themeColor) ([]themeColor, error) {
	if len(declared) == 0 {
		return nil, fmt.Errorf("web: theme %q: the built bundle declares no --color-neutral for it", theme)
	}
	byScheme := make(map[string]string, len(declared))
	for _, tc := range declared {
		if !themeColorRE.MatchString(tc.Color) {
			return nil, fmt.Errorf("web: theme %q: its --color-neutral is not a #rgb or #rrggbb color", theme)
		}
		if prior, ok := byScheme[tc.Scheme]; ok && prior != tc.Color {
			return nil, fmt.Errorf("web: theme %q: it declares two --color-neutral values under one color scheme", theme)
		}
		byScheme[tc.Scheme] = tc.Color
	}
	distinct := slices.Compact(slices.Sorted(maps.Values(byScheme)))
	if len(distinct) == 1 {
		return []themeColor{{Color: distinct[0]}}, nil
	}
	light, hasLight := byScheme["light"]
	dark, hasDark := byScheme["dark"]
	if len(byScheme) != 2 || !hasLight || !hasDark {
		return nil, fmt.Errorf("web: theme %q: its --color-neutral values differ under color schemes other than light and dark", theme)
	}
	return []themeColor{{Scheme: "light", Color: light}, {Scheme: "dark", Color: dark}}, nil
}

// tierHrefs returns the tier links for the document at docPath inside
// the build tree: the branding stylesheet always, and the operator
// stylesheet when one is configured. Each href is relative to the
// document's own URL, as the build's own links are, so the same
// document resolves them under any prefix.
func tierHrefs(docPath string, names RouteNames, hasOperatorCSS bool) []string {
	root := "./"
	if depth := strings.Count(docPath, "/"); depth > 0 {
		root = strings.Repeat("../", depth)
	}
	hrefs := []string{root + names.Branding}
	if hasOperatorCSS {
		hrefs = append(hrefs, root+names.OperatorCSS)
	}
	return hrefs
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

// namedDocument pairs a document's path inside the embedded build tree
// with its raw bytes, so [extractThemes] can resolve each stylesheet
// href it finds relative to the document that carries it.
type namedDocument struct {
	path string
	raw  []byte
}

// bundleThemes is what the built bundle's stylesheets declare about
// its themes: every [data-theme=<name>] selector name, and each named
// theme's --color-neutral declarations as [themeNeutrals] reads them.
type bundleThemes struct {
	selectors map[string]struct{}
	neutrals  map[string][]themeColor
}

// extractThemes reads the stylesheets docs link to, each href resolved
// relative to its own document's path inside build, and returns the
// union of what they declare about the bundle's themes.
func extractThemes(build fs.FS, docs ...namedDocument) (bundleThemes, error) {
	cssPaths := make(map[string]struct{})
	for _, doc := range docs {
		for _, href := range StylesheetHrefs(doc.raw) {
			cssPaths[resolveStylesheetPath(doc.path, href)] = struct{}{}
		}
	}
	if len(cssPaths) == 0 {
		return bundleThemes{}, fmt.Errorf("web: document %q: no stylesheet link found", docs[0].path)
	}

	themes := bundleThemes{
		selectors: make(map[string]struct{}),
		neutrals:  make(map[string][]themeColor),
	}
	for _, cssPath := range slices.Sorted(maps.Keys(cssPaths)) {
		data, err := fs.ReadFile(build, cssPath)
		if err != nil {
			return bundleThemes{}, fmt.Errorf("web: stylesheet %q: %w", cssPath, err)
		}
		for name := range themeSelectors(data) {
			themes.selectors[name] = struct{}{}
		}
		for name, tcs := range themeNeutrals(data) {
			themes.neutrals[name] = append(themes.neutrals[name], tcs...)
		}
	}
	return themes, nil
}

// resolveStylesheetPath resolves href, a document-relative path a
// built document's own <link> carries, against docPath, that
// document's own path inside the embedded build tree, and returns the
// result as a path relative to the tree's root.
func resolveStylesheetPath(docPath, href string) string {
	return path.Clean(path.Join(path.Dir(docPath), href))
}

// stampablePalettes returns every name in selectors that begins with
// match+"-", with that prefix removed, sorted: the set
// [buildDocumentVariants] builds a body for and a theme cookie's value
// is checked against for exact membership.
func stampablePalettes(selectors map[string]struct{}, match string) []string {
	prefix := match + "-"
	palettes := make([]string, 0, len(selectors))
	for name := range selectors {
		if after, ok := strings.CutPrefix(name, prefix); ok {
			palettes = append(palettes, after)
		}
	}
	slices.Sort(palettes)
	return palettes
}

// ExactPaths returns an http.Handler dispatching by exact
// http.Request.URL.Path to routes, answering [http.NotFoundHandler]'s
// 404 to any other path, the trailing-slash form of a route included.
// The shell mounts a subtree's every deeper path to the same handler,
// so the exact-match check here is what keeps a deeper or
// trailing-slash path from reaching a route.
func ExactPaths(routes map[string]http.Handler) http.Handler {
	return exactPaths(routes, http.NotFoundHandler())
}

// exactPaths is [ExactPaths] answering notFound to a path routes does
// not hold.
func exactPaths(routes map[string]http.Handler, notFound http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := routes[r.URL.Path]; ok {
			h.ServeHTTP(w, r)
			return
		}
		notFound.ServeHTTP(w, r)
	})
}
