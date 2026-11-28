package webapp

import (
	"bytes"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
)

// themeColorRE is the shape a value must have before it enters a
// theme-color meta. A bundle's own --color-neutral outside it is a
// build defect to refuse, since the bundle is this server's own build;
// a branding neutral outside it drops the meta instead.
var themeColorRE = regexp.MustCompile(`^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`)

// themeColor is one theme color: the color, and the color scheme it
// applies under, "light" or "dark", or "" when it applies whatever the
// device prefers. As a theme-color meta, a color with a scheme carries
// a prefers-color-scheme media query.
type themeColor struct {
	scheme string
	color  string
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
			colors[palette] = []themeColor{{color: brandNeutral}}
		}
		return colors, nil
	default:
		logger.Warn("webapp: DUBPLATE_BRANDING sets --color-neutral to a value other than a #rgb or #rrggbb color, so no document carries a theme-color meta; declare it as a three- or six-digit hex color to keep one")
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
		return nil, fmt.Errorf("webapp: theme %q: the built bundle declares no --color-neutral for it", theme)
	}
	byScheme := make(map[string]string, len(declared))
	for _, tc := range declared {
		if !themeColorRE.MatchString(tc.color) {
			return nil, fmt.Errorf("webapp: theme %q: its --color-neutral is not a #rgb or #rrggbb color", theme)
		}
		if prior, ok := byScheme[tc.scheme]; ok && prior != tc.color {
			return nil, fmt.Errorf("webapp: theme %q: it declares two --color-neutral values under one color scheme", theme)
		}
		byScheme[tc.scheme] = tc.color
	}
	distinct := slices.Compact(slices.Sorted(maps.Values(byScheme)))
	if len(distinct) == 1 {
		return []themeColor{{color: distinct[0]}}, nil
	}
	light, hasLight := byScheme["light"]
	dark, hasDark := byScheme["dark"]
	if len(byScheme) != 2 || !hasLight || !hasDark {
		return nil, fmt.Errorf("webapp: theme %q: its --color-neutral values differ under color schemes other than light and dark", theme)
	}
	return []themeColor{{scheme: "light", color: light}, {scheme: "dark", color: dark}}, nil
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
		for _, href := range stylesheetHrefs(doc.raw) {
			cssPaths[resolveStylesheetPath(doc.path, href)] = struct{}{}
		}
	}
	if len(cssPaths) == 0 {
		return bundleThemes{}, fmt.Errorf("webapp: document %q: no stylesheet link found", docs[0].path)
	}

	themes := bundleThemes{
		selectors: make(map[string]struct{}),
		neutrals:  make(map[string][]themeColor),
	}
	for _, cssPath := range slices.Sorted(maps.Keys(cssPaths)) {
		data, err := fs.ReadFile(build, cssPath)
		if err != nil {
			return bundleThemes{}, fmt.Errorf("webapp: stylesheet %q: %w", cssPath, err)
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

// dataThemeSelectorRE matches one [data-theme=<name>] compound
// selector, quoted or not, capturing the name alone: the closing "]"
// anchors the match, so a prefix of a real name (matched as a
// substring rather than the whole bracket expression) can never pass.
// It is applied to raw CSS text rather than a parsed selector tree, so
// a name inside an :is(...) list, which a minifier may produce, is
// found the same as a standalone selector.
var dataThemeSelectorRE = regexp.MustCompile(`\[data-theme=(?:"([a-zA-Z0-9-]+)"|([a-zA-Z0-9-]+))\]`)

// themeSelectors returns the set of every name css declares a
// [data-theme=<name>] compound selector for, quoted and unquoted forms
// both.
func themeSelectors(css []byte) map[string]struct{} {
	names := make(map[string]struct{})
	for _, m := range dataThemeSelectorRE.FindAllSubmatch(css, -1) {
		name := m[1]
		if len(name) == 0 {
			name = m[2]
		}
		names[string(name)] = struct{}{}
	}
	return names
}

// cssRuleRE matches one innermost CSS rule, capturing its selector
// text and its declaration block. Neither capture may hold a brace, so
// a rule nested inside an at-rule is matched on its own selector, never
// on the at-rule's prelude.
var cssRuleRE = regexp.MustCompile(`([^{}]*)\{([^{}]*)\}`)

// neutralDeclRE and colorSchemeDeclRE find one declaration inside a
// rule's block. The leading anchor keeps a longer name ending in the
// same letters from matching, and the name's own ":" keeps
// --color-neutral-content from matching as --color-neutral.
var (
	neutralDeclRE     = regexp.MustCompile(`(?:^|;)\s*--color-neutral\s*:\s*([^;]*)`)
	colorSchemeDeclRE = regexp.MustCompile(`(?:^|;)\s*color-scheme\s*:\s*([^;]*)`)
)

// themeNeutrals returns, for every theme name css declares a
// [data-theme=<name>] rule for, the --color-neutral value of each such
// rule that declares one, paired with the color-scheme the same rule
// declares ("" when it declares none). A rule whose selector names
// several themes counts for each. The values are returned as css
// writes them, trimmed of surrounding space, and are not validated.
func themeNeutrals(css []byte) map[string][]themeColor {
	neutrals := make(map[string][]themeColor)
	for _, rule := range cssRuleRE.FindAllSubmatch(css, -1) {
		neutral := neutralDeclRE.FindSubmatch(rule[2])
		if neutral == nil {
			continue
		}
		tc := themeColor{color: string(bytes.TrimSpace(neutral[1]))}
		if scheme := colorSchemeDeclRE.FindSubmatch(rule[2]); scheme != nil {
			tc.scheme = string(bytes.TrimSpace(scheme[1]))
		}
		for name := range themeSelectors(rule[1]) {
			neutrals[name] = append(neutrals[name], tc)
		}
	}
	return neutrals
}
