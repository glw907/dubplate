package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
	"github.com/glw907/dubplate/web"
)

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
// the build tree, the path under the owned prefix it answers at, and
// its Cache-Control value.
type webDocument struct {
	name         string
	route        string
	cacheControl string
}

// webDocuments is every document the build emits and the server
// serves. A document is no-cache by default: the stamp and the tier
// links are fixed at startup, so a cache may keep it and revalidate.
// The consumption document is no-store because its URL carried a
// sign-in token. The members document is private because it answers
// per requester, behind the admin gate.
var webDocuments = []webDocument{
	{name: loginDocumentName, route: "/" + loginRouteName, cacheControl: "no-cache"},
	{name: loginLinkDocumentName, route: "/" + loginRouteName + "/link", cacheControl: "no-store"},
	{name: membersDocumentName, route: "/" + membersRouteName, cacheControl: "private, no-cache"},
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
// configured, which leaves its path to the terminator's 404. members
// is the members document with no gate of its own, so it is never
// mounted except behind [auth.RequireAdmin].
type webRoutes struct {
	login    http.Handler
	members  http.Handler
	assets   http.Handler
	branding http.Handler
	operator http.Handler
}

// register mounts wr's public routes on shell under their subtree
// names, the login documents logged through the shell's own logger.
// It leaves members to the caller, which alone holds the admin gate.
func (wr webRoutes) register(shell *serve.Shell) {
	shell.Register(loginRouteName, serve.LogRequests(shell.Logger(), wr.login))
	shell.Register(assetsRouteName, wr.assets)
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
// variant per stampable palette with the tier links in place, and
// returns the routes. It returns before any listener opens, so a
// document with an inline script its CSP does not hash, a bad theme or
// branding value, a refused operator file, or a missing build refuses
// startup rather than the first request.
func newWebRoutes(environ []string, prefix serve.Prefix) (webRoutes, error) {
	buildFS, err := fs.Sub(web.Assets, "build")
	if err != nil {
		return webRoutes{}, fmt.Errorf("web: %w", err)
	}
	return newWebRoutesFromFS(environ, prefix, buildFS)
}

// newWebRoutesFromFS is [newWebRoutes]'s logic, parameterized on
// buildFS rather than reading the package-level [web.Assets]: every
// state a bad or missing build can be in is a fixture this file's own
// tests build as an fstest.MapFS, since web.Assets is fixed to
// whatever bundle this binary was compiled against.
func newWebRoutesFromFS(environ []string, prefix serve.Prefix, buildFS fs.FS) (webRoutes, error) {
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

	selectors, err := extractThemeSelectors(buildFS, docs...)
	if err != nil {
		return webRoutes{}, err
	}

	composed := match + "-" + palette
	if _, ok := selectors[composed]; !ok {
		return webRoutes{}, fmt.Errorf("%s: %q: no theme selector named %q in the built bundle", themePaletteThreshold.Key, palette, composed)
	}
	palettes := stampablePalettes(selectors, match)

	branding, err := resolveBranding(environ)
	if err != nil {
		return webRoutes{}, err
	}
	operatorCSS, hasOperatorCSS, err := resolveOperatorCSS(environ)
	if err != nil {
		return webRoutes{}, err
	}

	documents := make(map[string]http.Handler, len(webDocuments))
	for i, d := range webDocuments {
		variants, err := serve.BuildDocumentVariants(d.name, docs[i].raw, themePlaceholder, match, palettes,
			tierHrefs(d.name, hasOperatorCSS))
		if err != nil {
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
	routes := webRoutes{
		login:    login,
		members:  members,
		assets:   serve.NewAssets(assetsFS, prefix.String()+"/"+assetsRouteName),
		branding: stylesheetRoute(prefix, brandingRouteName, branding),
	}
	if hasOperatorCSS {
		routes.operator = stylesheetRoute(prefix, operatorCSSRouteName, operatorCSS)
	}
	return routes, nil
}

// stylesheetRoute serves css at exactly <prefix>/<name> and answers
// the terminator's 404 everywhere under it.
func stylesheetRoute(prefix serve.Prefix, name string, css []byte) http.Handler {
	return newExactPathRoutes(map[string]http.Handler{
		prefix.String() + "/" + name: serve.NewStylesheet(css),
	})
}

// resolveBranding renders DUBPLATE_BRANDING into the branding
// stylesheet, empty when the key is unset, empty, or "none".
func resolveBranding(environ []string) ([]byte, error) {
	decls, err := config.ParseString(environ, brandingThreshold)
	if err != nil {
		return nil, err
	}
	if decls == notConfigured {
		decls = ""
	}
	css, err := serve.BrandingStylesheet(decls)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", brandingThreshold.Key, err)
	}
	return css, nil
}

// resolveOperatorCSS reads the file DUBPLATE_OPERATOR_CSS_FILE names,
// reporting false when the key is unset, empty, or "none".
func resolveOperatorCSS(environ []string) ([]byte, bool, error) {
	path, err := config.ParseString(environ, operatorCSSFileThreshold)
	if err != nil {
		return nil, false, err
	}
	if path == notConfigured {
		return nil, false, nil
	}
	css, err := readOperatorCSS(path)
	if err != nil {
		return nil, false, fmt.Errorf("%s: %w", operatorCSSFileThreshold.Key, err)
	}
	return css, true, nil
}

// readOperatorCSS reads the operator's stylesheet at path. The route
// that serves it is public, so the file's class is checked before its
// bytes are kept: a wrong path would otherwise publish whatever the
// service user can read, the roster or an environment file included.
// The path must end in ".css", which none of those do, and the content
// must be UTF-8 text holding no NUL byte and no more than
// operatorCSSMaxBytes. A returned error names the rule or the reason
// and never carries the path or any byte of the file.
func readOperatorCSS(path string) ([]byte, error) {
	if !strings.HasSuffix(path, ".css") {
		return nil, errors.New(`the path must end in ".css"`)
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("the path is not a regular file")
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read the file: %w", withoutPath(err))
	}
	defer func() { _ = f.Close() }()
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

// extractThemeSelectors returns the union of every [data-theme=<name>]
// selector name declared in the stylesheets docs link to, each href
// resolved relative to its own document's path inside buildFS
// (R-5cii12).
func extractThemeSelectors(buildFS fs.FS, docs ...namedDocument) (map[string]struct{}, error) {
	cssPaths := make(map[string]struct{})
	for _, doc := range docs {
		for _, href := range serve.StylesheetHrefs(doc.raw) {
			cssPaths[resolveStylesheetPath(doc.path, href)] = struct{}{}
		}
	}
	if len(cssPaths) == 0 {
		return nil, fmt.Errorf("web: document %q: no stylesheet link found", docs[0].path)
	}

	selectors := make(map[string]struct{})
	for cssPath := range cssPaths {
		data, err := fs.ReadFile(buildFS, cssPath)
		if err != nil {
			return nil, fmt.Errorf("web: stylesheet %q: %w", cssPath, err)
		}
		for name := range serve.ThemeSelectors(data) {
			selectors[name] = struct{}{}
		}
	}
	return selectors, nil
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
// http.Request.URL.Path to routes, answering [serve.Terminator]'s 404
// to any other path, the trailing-slash form of a route included:
// [serve.Shell.Register] mounts a subtree's every deeper path to the
// same handler, so the exact-match check here is what keeps a deeper
// or trailing-slash path from reaching a document (R-5cii21).
func newExactPathRoutes(routes map[string]http.Handler) http.Handler {
	terminator := serve.Terminator()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h, ok := routes[r.URL.Path]; ok {
			h.ServeHTTP(w, r)
			return
		}
		terminator.ServeHTTP(w, r)
	})
}
