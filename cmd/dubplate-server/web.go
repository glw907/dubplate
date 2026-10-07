package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"maps"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/web"
)

// fontsRouteName is the embedded faces' subtree name, beside the four
// [newWebRoutes] names in serve.go.
const fontsRouteName = "fonts"

// The mark's subtree names: the favicon in its two forms, and the
// directory holding the mark's three detail levels.
const (
	faviconSVGRouteName = "favicon.svg"
	faviconPNGRouteName = "favicon.png"
	markRouteName       = "mark"
)

// authRobots is the robots meta the two sign-in documents carry, so a
// search engine neither lists a sign-in page nor follows a link out of
// one (W7.15).
const authRobots = "noindex, nofollow"

// themeColorRE is the shape a value must have before it enters a
// theme-color meta. A bundle's own --color-neutral outside it is a
// build defect to refuse, since the bundle is this server's own build;
// a branding neutral outside it drops the meta instead.
var themeColorRE = regexp.MustCompile(`^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`)

// themePlaceholder is the token app.html carries on <html data-theme>,
// substituted once per stampable palette at serve startup. It must
// match web/src/app.html's own literal exactly.
const themePlaceholder = "__DUBPLATE_THEME__"

// acceptedThemeMatch is the one value themeMatchThreshold accepts (DL
// R3.2): the match axis names a group of themes rather than a single
// one, and dubplate ships one group.
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

// operatorCSSMaxBytes bounds the operator CSS file, 256 KiB: well above
// any hand-written override, and small enough that a wrong path naming
// a large file refuses startup rather than loading it into memory.
const operatorCSSMaxBytes = 256 << 10

// The build tree's three document paths, relative to the embedded
// build root.
const (
	loginDocumentName     = "login.html"
	loginLinkDocumentName = "login/link.html"
	membersDocumentName   = "members.html"
)

// webDocument is one document the web routes serve: its path inside
// the build tree, the path under the owned prefix it answers at, its
// Cache-Control value, and whether it carries the robots meta.
type webDocument struct {
	name         string
	route        string
	cacheControl string
	noIndex      bool
}

// webDocuments is every document the build emits and the server
// serves. A document is no-cache by default: the stamp and the tier
// links are fixed at startup, so a cache may keep it and revalidate.
// The consumption document is no-store because its URL carried a
// sign-in token. The members document is no-store because it answers
// only behind the admin gate, as the roster API it reads from does.
var webDocuments = []webDocument{
	{name: loginDocumentName, route: "/" + loginRouteName, cacheControl: "no-cache", noIndex: true},
	{name: loginLinkDocumentName, route: "/" + loginRouteName + "/link", cacheControl: "no-store", noIndex: true},
	{name: membersDocumentName, route: "/" + membersRouteName, cacheControl: "no-store"},
}

// privateFiles names the state the operator CSS file must never be:
// anything under the state directory. The operator route is public, so
// a path under it would publish the store. An empty field guards
// nothing.
type privateFiles struct {
	stateDir string
}

// themeMatchThreshold is DUBPLATE_THEME_MATCH: the theme axis naming
// which group of themes the server presents. Rung 5 ships one group,
// so the only accepted value is acceptedThemeMatch.
var themeMatchThreshold = config.Threshold{
	Key:     "DUBPLATE_THEME_MATCH",
	Kind:    config.KindString,
	Default: acceptedThemeMatch,
	Doc:     `exactly "dubplate"`,
}

// themePaletteThreshold is DUBPLATE_THEME_PALETTE: the theme a visitor
// without a theme cookie sees. Its Doc string is the operator-facing
// text this key's own doctor row and the config wall report.
var themePaletteThreshold = config.Threshold{
	Key:     "DUBPLATE_THEME_PALETTE",
	Kind:    config.KindString,
	Default: "system",
	Doc:     `The theme a visitor without a theme cookie sees. "system" follows the device's light or dark preference between "label" and "label-night", and falls back to "label"; a palette name pins that palette.`,
}

// brandingThreshold is DUBPLATE_BRANDING: the operator's declaration
// list, rendered once at startup into the branding stylesheet.
var brandingThreshold = config.Threshold{
	Key:     "DUBPLATE_BRANDING",
	Kind:    config.KindString,
	Default: notConfigured,
	Doc:     `"none", or ";"-separated name: value declarations`,
}

// operatorCSSFileThreshold is DUBPLATE_OPERATOR_CSS_FILE: the
// operator's own stylesheet, read once at startup by [readOperatorCSS].
var operatorCSSFileThreshold = config.Threshold{
	Key:     "DUBPLATE_OPERATOR_CSS_FILE",
	Kind:    config.KindString,
	Default: notConfigured,
	Doc:     `"none", or a path to a .css file`,
}

// registerWebConfig adds the web surface's four presentation
// thresholds to reg (R-5cii12). newProductionRegistry calls it after
// registerStoreConfig.
func registerWebConfig(reg *config.Registry) {
	reg.AddThreshold(themeMatchThreshold)
	reg.AddThreshold(themePaletteThreshold)
	reg.AddThreshold(brandingThreshold)
	reg.AddThreshold(operatorCSSFileThreshold)
}

// webRoutes are the http.Handlers newServeShell mounts under the
// shell's owned prefix. operator is nil when no operator CSS file is
// configured, which leaves its path to the shell's 404. members
// is the members document with no gate of its own, so it is never
// mounted except behind [auth.RequireAdmin].
type webRoutes struct {
	login    http.Handler
	members  http.Handler
	assets   http.Handler
	fonts    http.Handler
	icons    http.Handler
	branding http.Handler
	operator http.Handler
}

// register mounts wr's public routes on shell under their subtree
// names, the login documents logged through the shell's own logger.
// It leaves members to the caller, which alone holds the admin gate.
func (wr webRoutes) register(shell *serve.Shell) {
	shell.Register(loginRouteName, serve.LogRequests(shell.Logger(), wr.login))
	shell.Register(assetsRouteName, wr.assets)
	shell.Register(fontsRouteName, wr.fonts)
	shell.Register(faviconSVGRouteName, wr.icons)
	shell.Register(faviconPNGRouteName, wr.icons)
	shell.Register(markRouteName, wr.icons)
	shell.Register(brandingRouteName, wr.branding)
	if wr.operator != nil {
		shell.Register(operatorCSSRouteName, wr.operator)
	}
}

// newWebRoutes reads web.Assets, refuses on a placeholder-only
// embedded tree naming `npm run build`, builds each document's CSP
// header from its own CSP meta, resolves and validates
// DUBPLATE_THEME_MATCH and DUBPLATE_THEME_PALETTE against the built
// bundle's own extracted [data-theme=<name>] selectors, renders
// DUBPLATE_BRANDING into the branding stylesheet, reads the file
// DUBPLATE_OPERATOR_CSS_FILE names, rewrites the documents into one
// variant per stampable palette with the tier links and the head
// metas in place, and returns the routes. It returns before any
// listener opens, so a document with an inline script its CSP does not
// hash, in the build or in any rewritten variant, a bad theme or
// branding value, a bundle theme with no usable neutral, a refused
// operator file, or a missing build refuses startup rather than the
// first request.
//
// siteName enters every document as the dubplate-site-name meta.
// private names the files the operator CSS path must not resolve to.
// logger takes the one record written when a branding neutral leaves
// the documents without a theme-color meta.
func newWebRoutes(environ []string, prefix serve.Prefix, siteName string, private privateFiles, logger *slog.Logger) (webRoutes, error) {
	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		return webRoutes{}, fmt.Errorf("web: %w", err)
	}
	return newWebRoutesFromFS(environ, prefix, siteName, private, logger, buildFS)
}

// newWebRoutesFromFS is [newWebRoutes]'s logic, parameterized on
// buildFS rather than reading the package-level [web.Assets]: every
// state a bad or missing build can be in is a fixture this file's own
// tests build as an fstest.MapFS, since web.Assets is fixed to
// whatever bundle this binary was compiled against.
func newWebRoutesFromFS(environ []string, prefix serve.Prefix, siteName string, private privateFiles, logger *slog.Logger, buildFS fs.FS) (webRoutes, error) {
	placeholderOnly, err := isPlaceholderOnly(buildFS)
	if err != nil {
		return webRoutes{}, fmt.Errorf("web: %w", err)
	}
	if placeholderOnly {
		return webRoutes{}, errors.New("web: the embedded build tree holds only the placeholder; run `npm run build` in web/")
	}

	docs := make([]namedDocument, len(webDocuments))
	csps := make([]string, len(webDocuments))
	for i, d := range webDocuments {
		raw, err := fs.ReadFile(buildFS, d.name)
		if err != nil {
			return webRoutes{}, fmt.Errorf("web: document %q: %w", d.name, err)
		}
		csp, err := serve.DocumentCSP(d.name, raw)
		if err != nil {
			return webRoutes{}, err
		}
		docs[i] = namedDocument{path: d.name, raw: raw}
		csps[i] = csp
	}

	match, err := config.ParseString(environ, themeMatchThreshold)
	if err != nil {
		return webRoutes{}, err
	}
	if match != acceptedThemeMatch {
		return webRoutes{}, fmt.Errorf("%s: %q: must be exactly %q", themeMatchThreshold.Key, match, acceptedThemeMatch)
	}

	palette, err := config.ParseString(environ, themePaletteThreshold)
	if err != nil {
		return webRoutes{}, err
	}
	if !themePaletteGrammarRE.MatchString(palette) {
		return webRoutes{}, fmt.Errorf("%s: %q: must match %s", themePaletteThreshold.Key, palette, themePaletteGrammarRE)
	}

	themes, err := extractThemes(buildFS, docs...)
	if err != nil {
		return webRoutes{}, err
	}

	composed := match + "-" + palette
	if _, ok := themes.selectors[composed]; !ok {
		return webRoutes{}, fmt.Errorf("%s: %q: no theme selector named %q in the built bundle", themePaletteThreshold.Key, palette, composed)
	}
	palettes := stampablePalettes(themes.selectors, match)

	branding, brandNeutral, err := resolveBranding(environ)
	if err != nil {
		return webRoutes{}, err
	}
	colors, err := themeColors(themes.neutrals, match, palettes, brandNeutral, logger)
	if err != nil {
		return webRoutes{}, err
	}
	operatorCSS, hasOperatorCSS, err := resolveOperatorCSS(environ, private)
	if err != nil {
		return webRoutes{}, err
	}

	icons := []serve.Icon{
		{Href: prefix.String() + "/" + faviconSVGRouteName, Type: "image/svg+xml"},
		{Href: prefix.String() + "/" + faviconPNGRouteName, Type: "image/png", Sizes: "32x32"},
	}
	documents := make(map[string]http.Handler, len(webDocuments))
	for i, d := range webDocuments {
		cfg := serve.VariantConfig{
			Placeholder: themePlaceholder,
			Match:       match,
			Palettes:    palettes,
			TierHrefs:   tierHrefs(d.name, hasOperatorCSS),
			SiteName:    siteName,
			ThemeColors: colors,
			Icons:       icons,
		}
		if d.noIndex {
			cfg.Robots = authRobots
		}
		variants, err := serve.BuildDocumentVariants(d.name, docs[i].raw, cfg)
		if err != nil {
			return webRoutes{}, err
		}
		if err := checkVariantCSP(d.name, variants, csps[i]); err != nil {
			return webRoutes{}, err
		}
		documents[prefix.String()+d.route] = serve.NewDocument(serve.DocumentConfig{
			Variants:       variants,
			DefaultPalette: palette,
			CacheControl:   d.cacheControl,
			CSP:            csps[i],
		})
	}
	membersPath := prefix.String() + "/" + membersRouteName
	members := newExactPathRoutes(map[string]http.Handler{membersPath: documents[membersPath]})
	delete(documents, membersPath)
	login := newExactPathRoutes(documents)

	assetsFS, err := fs.Sub(buildFS, "_app")
	if err != nil {
		return webRoutes{}, fmt.Errorf("web: %w", err)
	}
	fontsFS, err := fs.Sub(buildFS, "fonts")
	if err != nil {
		return webRoutes{}, fmt.Errorf("web: %w", err)
	}
	fontsRoute, err := newFontsRoute(fontsFS, prefix.String()+"/"+fontsRouteName)
	if err != nil {
		return webRoutes{}, err
	}
	iconsRoute, err := newIconsRoute(buildFS, prefix)
	if err != nil {
		return webRoutes{}, err
	}
	routes := webRoutes{
		login:    login,
		members:  members,
		assets:   serve.NewAssets(assetsFS, prefix.String()+"/"+assetsRouteName),
		fonts:    fontsRoute,
		icons:    iconsRoute,
		branding: stylesheetRoute(prefix, brandingRouteName, branding),
	}
	if hasOperatorCSS {
		routes.operator = stylesheetRoute(prefix, operatorCSSRouteName, operatorCSS)
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
		csp, err := serve.DocumentCSP(name, variants[palette])
		if err != nil {
			return fmt.Errorf("web: the %q variant: %w", palette, err)
		}
		if csp != served {
			return fmt.Errorf("web: the %q variant of document %q: its CSP differs from the served header", palette, name)
		}
	}
	return nil
}

// stylesheetRoute serves css at exactly <prefix>/<name> and answers
// the shell's 404 everywhere under it.
func stylesheetRoute(prefix serve.Prefix, name string, css []byte) http.Handler {
	return newExactPathRoutes(map[string]http.Handler{
		prefix.String() + "/" + name: serve.NewStylesheet(css),
	})
}

// svgCSP is the policy every SVG response carries. A browser opening an
// SVG directly runs it as a document, script and all; the policy
// refuses every fetch and script and sandboxes it, keeping only the
// inline styles an SVG's own presentation may use.
const svgCSP = "default-src 'none'; style-src 'unsafe-inline'; sandbox"

// staticFile is one embedded file served at an exact path: its bytes,
// its Content-Type, and its strong ETag, all fixed at startup so no
// request byte reaches them.
type staticFile struct {
	data        []byte
	contentType string
	etag        string
}

// newStaticFile returns data as a staticFile of contentType, its ETag
// a strong SHA-256 hash of data.
func newStaticFile(data []byte, contentType string) staticFile {
	sum := sha256.Sum256(data)
	return staticFile{data: data, contentType: contentType, etag: `"` + hex.EncodeToString(sum[:]) + `"`}
}

// newStaticFiles returns an http.Handler serving GET and HEAD only, at
// exactly the request paths files holds, and answering
// [http.NotFoundHandler]'s 404 on every other path, with no directory
// listing. Every file is served no-cache, since its name stays put
// across builds while its bytes may not.
//
// Content-Type is set on w before [http.ServeContent] runs, so no host
// mime table can override it (Go's built-in extension table carries no
// ".woff2" entry, and its fallback reads host files that differ
// between the workstation and the box). The ETag lets a revalidation
// answer 304: "If the caller has set w's ETag header formatted per RFC
// 7232, section 2.3, ServeContent uses it to handle requests using
// If-Match, If-None-Match, or If-Range" (go doc
// net/http.ServeContent), which holds even though every response
// carries a zero modtime.
func newStaticFiles(files map[string]staticFile) http.Handler {
	notFound := http.NotFoundHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		file, ok := files[r.URL.Path]
		if !ok {
			notFound.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Content-Type", file.contentType)
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("ETag", file.etag)
		if file.contentType == "image/svg+xml" {
			w.Header().Set("Content-Security-Policy", svgCSP)
		}
		http.ServeContent(w, r, path.Base(r.URL.Path), time.Time{}, bytes.NewReader(file.data))
	})
}

// newFontsRoute returns a [newStaticFiles] handler serving, as
// font/woff2, every .woff2 file fontsFS holds at its root, each at
// exactly routePrefix followed by a separator and its name. Every
// other path under it -- routePrefix itself, a directory (fontsFS's
// own "licenses/"), a top-level file that is not a .woff2 (FONTS.md),
// a trailing slash, and a name fontsFS does not hold -- answers
// [http.NotFoundHandler]'s 404. It refuses when fontsFS holds no .woff2
// file, the same shape [isPlaceholderOnly] refuses on for the whole
// bundle.
func newFontsRoute(fontsFS fs.FS, routePrefix string) (http.Handler, error) {
	faces, err := readStaticDir(fontsFS, ".", ".woff2", "font/woff2", routePrefix)
	if err != nil {
		return nil, fmt.Errorf("web: fonts: %w", err)
	}
	if len(faces) == 0 {
		return nil, errors.New("web: fonts: no .woff2 file embedded under web/build/fonts/")
	}
	return newStaticFiles(faces), nil
}

// newIconsRoute returns a [newStaticFiles] handler serving the mark:
// the favicon as <prefix>/favicon.svg and <prefix>/favicon.png, the
// two files the documents' icon links name, and every .svg file under
// the build's mark/ directory as <prefix>/mark/<name>. It refuses when
// either favicon or every mark file is missing, since a document would
// then link a file the server cannot answer.
func newIconsRoute(buildFS fs.FS, prefix serve.Prefix) (http.Handler, error) {
	files, err := readStaticDir(buildFS, markRouteName, ".svg", "image/svg+xml", prefix.String()+"/"+markRouteName)
	if err != nil {
		return nil, fmt.Errorf("web: mark: %w", err)
	}
	if len(files) == 0 {
		return nil, errors.New("web: mark: no .svg file embedded under web/build/mark/")
	}
	for _, favicon := range []struct{ name, contentType string }{
		{faviconSVGRouteName, "image/svg+xml"},
		{faviconPNGRouteName, "image/png"},
	} {
		data, err := fs.ReadFile(buildFS, favicon.name)
		if err != nil {
			return nil, fmt.Errorf("web: favicon: %w", err)
		}
		files[prefix.String()+"/"+favicon.name] = newStaticFile(data, favicon.contentType)
	}
	return newStaticFiles(files), nil
}

// readStaticDir returns every regular file directly under dir in fsys
// whose name ends in suffix, as a staticFile of contentType keyed by
// routePrefix followed by a separator and its name. Subdirectories are
// skipped.
func readStaticDir(fsys fs.FS, dir, suffix, contentType, routePrefix string) (map[string]staticFile, error) {
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return nil, err
	}
	files := make(map[string]staticFile, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), suffix) {
			continue
		}
		data, err := fs.ReadFile(fsys, path.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("%q: %w", entry.Name(), err)
		}
		files[routePrefix+"/"+entry.Name()] = newStaticFile(data, contentType)
	}
	return files, nil
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
	css, err := serve.BrandingStylesheet(decls)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", brandingThreshold.Key, err)
	}
	return css, brandingNeutral(decls), nil
}

// brandingNeutral returns the value decls declares for
// --color-neutral, or "" when it declares none. decls must already
// have passed [serve.BrandingStylesheet], which fixes every
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
func themeColors(neutrals map[string][]serve.ThemeColor, match string, palettes []string, brandNeutral string, logger *slog.Logger) (map[string][]serve.ThemeColor, error) {
	colors := make(map[string][]serve.ThemeColor, len(palettes))
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
			colors[palette] = []serve.ThemeColor{{Color: brandNeutral}}
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
func paletteThemeColors(theme string, declared []serve.ThemeColor) ([]serve.ThemeColor, error) {
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
		return []serve.ThemeColor{{Color: distinct[0]}}, nil
	}
	light, hasLight := byScheme["light"]
	dark, hasDark := byScheme["dark"]
	if len(byScheme) != 2 || !hasLight || !hasDark {
		return nil, fmt.Errorf("web: theme %q: its --color-neutral values differ under color schemes other than light and dark", theme)
	}
	return []serve.ThemeColor{{Scheme: "light", Color: light}, {Scheme: "dark", Color: dark}}, nil
}

// resolveOperatorCSS reads the file DUBPLATE_OPERATOR_CSS_FILE names,
// reporting false when the key is unset, empty, or "none".
func resolveOperatorCSS(environ []string, private privateFiles) ([]byte, bool, error) {
	path, err := config.ParseString(environ, operatorCSSFileThreshold)
	if err != nil {
		return nil, false, err
	}
	if path == notConfigured {
		return nil, false, nil
	}
	css, err := readOperatorCSS(path, private)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", operatorCSSFileThreshold.Key, err)
	}
	return css, true, nil
}

// readOperatorCSS reads the operator's stylesheet at path. The route
// that serves it is public, so the file's class is checked before its
// bytes are kept: a wrong path would otherwise publish whatever the
// service user can read, an environment file included.
//
// The path must end in ".css", and so must the path it resolves to
// once every symbolic link in it is followed, which is the path
// opened: a link named x.css cannot stand in for another file. The
// resolved path must not lie under private.stateDir. The file is opened once, with
// O_NOFOLLOW so a link swapped in for the file itself after resolution
// is refused, and with O_NONBLOCK so a FIFO cannot hang startup; the
// class check reads the open descriptor, so the file checked is the
// file read. O_NOFOLLOW reaches the last path element alone: a
// directory on the resolved path swapped for a link after resolution
// is still followed, so the checks above hold only while no one else
// can write a directory on the path.
// The content must be UTF-8 text holding no NUL byte and no more than
// operatorCSSMaxBytes. A returned error names the rule or the reason
// and never carries the path or any byte of the file.
func readOperatorCSS(path string, private privateFiles) ([]byte, error) {
	if !strings.HasSuffix(path, ".css") {
		return nil, errors.New(`the path must end in ".css"`)
	}
	resolved, err := resolvePath(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	if !strings.HasSuffix(resolved, ".css") {
		return nil, errors.New(`the path must resolve to a file ending in ".css"`)
	}
	if private.stateDir != "" {
		stateDir, err := resolvePath(private.stateDir)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve the state directory: %w", withoutPath(err))
		}
		if rel, err := filepath.Rel(stateDir, resolved); err == nil && filepath.IsLocal(rel) {
			return nil, errors.New("the file is under the state directory")
		}
	}

	f, err := os.OpenFile(resolved, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	defer func() { _ = f.Close() }()
	info, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("the path is not a regular file")
	}
	css, err := io.ReadAll(io.LimitReader(f, operatorCSSMaxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	switch {
	case len(css) > operatorCSSMaxBytes:
		return nil, fmt.Errorf("the file is over %d bytes", operatorCSSMaxBytes)
	case !utf8.Valid(css):
		return nil, errors.New("the file is not valid UTF-8")
	case bytes.IndexByte(css, 0) >= 0:
		return nil, errors.New("the file holds a NUL byte")
	}
	return css, nil
}

// resolvePath returns path made absolute with every symbolic link in
// it followed.
func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
}

// withoutPath returns the error inside err's *fs.PathError, or err
// itself when it carries none, so a reported reason ("no such file or
// directory") keeps its errors.Is identity and drops the path.
func withoutPath(err error) error {
	if pathErr, ok := errors.AsType[*fs.PathError](err); ok {
		return pathErr.Err
	}
	return err
}

// tierHrefs returns the tier links for the document at docPath inside
// the build tree: the branding stylesheet always, and the operator
// stylesheet when one is configured. Each href is relative to the
// document's own URL, as the build's own links are, so the same
// document resolves them under any prefix.
func tierHrefs(docPath string, hasOperatorCSS bool) []string {
	root := "./"
	if depth := strings.Count(docPath, "/"); depth > 0 {
		root = strings.Repeat("../", depth)
	}
	hrefs := []string{root + brandingRouteName}
	if hasOperatorCSS {
		hrefs = append(hrefs, root+operatorCSSRouteName)
	}
	return hrefs
}

// isPlaceholderOnly reports whether buildFS's root directory holds
// nothing but placeholder.txt, the shape a checkout that has never run
// `npm run build` leaves (web/embed.go, R-5cii9).
func isPlaceholderOnly(buildFS fs.FS) (bool, error) {
	entries, err := fs.ReadDir(buildFS, ".")
	if err != nil {
		return false, err
	}
	return len(entries) == 1 && entries[0].Name() == "placeholder.txt", nil
}

// namedDocument pairs a document's path inside the embedded build tree
// with its raw bytes, so [extractThemeSelectors] can resolve each
// stylesheet href it finds relative to the document that carries it.
type namedDocument struct {
	path string
	raw  []byte
}

// bundleThemes is what the built bundle's stylesheets declare about
// its themes: every [data-theme=<name>] selector name, and each named
// theme's --color-neutral declarations as [serve.ThemeNeutrals] reads
// them.
type bundleThemes struct {
	selectors map[string]struct{}
	neutrals  map[string][]serve.ThemeColor
}

// extractThemes reads the stylesheets docs link to, each href resolved
// relative to its own document's path inside buildFS (R-5cii12), and
// returns the union of what they declare about the bundle's themes.
func extractThemes(buildFS fs.FS, docs ...namedDocument) (bundleThemes, error) {
	cssPaths := make(map[string]struct{})
	for _, doc := range docs {
		for _, href := range serve.StylesheetHrefs(doc.raw) {
			cssPaths[resolveStylesheetPath(doc.path, href)] = struct{}{}
		}
	}
	if len(cssPaths) == 0 {
		return bundleThemes{}, fmt.Errorf("web: document %q: no stylesheet link found", docs[0].path)
	}

	themes := bundleThemes{
		selectors: make(map[string]struct{}),
		neutrals:  make(map[string][]serve.ThemeColor),
	}
	for _, cssPath := range slices.Sorted(maps.Keys(cssPaths)) {
		data, err := fs.ReadFile(buildFS, cssPath)
		if err != nil {
			return bundleThemes{}, fmt.Errorf("web: stylesheet %q: %w", cssPath, err)
		}
		for name := range serve.ThemeSelectors(data) {
			themes.selectors[name] = struct{}{}
		}
		for name, tcs := range serve.ThemeNeutrals(data) {
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
// [serve.BuildDocumentVariants] builds a body for and a theme cookie's
// value is checked against for exact membership (R-5cii12, R-5cii22).
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

// newExactPathRoutes returns an http.Handler dispatching by exact
// http.Request.URL.Path to routes, answering [http.NotFoundHandler]'s 404
// to any other path, the trailing-slash form of a route included:
// [serve.Shell.Register] mounts a subtree's every deeper path to the
// same handler, so the exact-match check here is what keeps a deeper
// or trailing-slash path from reaching a document (R-5cii21).
func newExactPathRoutes(routes map[string]http.Handler) http.Handler {
	notFound := http.NotFoundHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := routes[r.URL.Path]; ok {
			h.ServeHTTP(w, r)
			return
		}
		notFound.ServeHTTP(w, r)
	})
}
