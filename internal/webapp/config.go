package webapp

import (
	"fmt"
	"regexp"

	"github.com/glw907/dubplate/internal/config"
)

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

// themeMatchThreshold is DUBPLATE_THEME_MATCH: the theme axis naming
// which group of themes the server presents. The server ships one
// group, so the only accepted value is acceptedThemeMatch.
var themeMatchThreshold = config.Threshold{
	Key:         "DUBPLATE_THEME_MATCH",
	Kind:        config.KindString,
	Default:     acceptedThemeMatch,
	Doc:         `exactly "dubplate"`,
	Description: "the theme group the server presents",
	Sensitivity: config.NotSensitive,
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
	Sensitivity: config.NotSensitive,
}

// brandingThreshold is DUBPLATE_BRANDING: the operator's declaration
// list, rendered once at startup into the branding stylesheet.
var brandingThreshold = config.Threshold{
	Key:         "DUBPLATE_BRANDING",
	Kind:        config.KindString,
	Default:     notConfigured,
	Doc:         `"none", or ";"-separated name: value declarations`,
	Description: "the operator's style declarations, applied to every page",
	Sensitivity: config.NotSensitive,
}

// operatorCSSFileThreshold is DUBPLATE_OPERATOR_CSS_FILE: the
// operator's own stylesheet, read once at startup by [readOperatorCSS].
var operatorCSSFileThreshold = config.Threshold{
	Key:         "DUBPLATE_OPERATOR_CSS_FILE",
	Kind:        config.KindString,
	Default:     notConfigured,
	Doc:         `"none", or a path to a .css file`,
	Description: "the path to the operator's own stylesheet",
	Sensitivity: config.Sensitive,
}

// Register adds the web surface's four presentation thresholds to
// reg. The command calls it after the store's.
func Register(reg *config.Registry) {
	reg.AddThreshold(themeMatchThreshold)
	reg.AddThreshold(themePaletteThreshold)
	reg.AddThreshold(brandingThreshold)
	reg.AddThreshold(operatorCSSFileThreshold)
}

// FromEnviron resolves the four web thresholds from environ into the
// settings half of a [Config]: Match, Palette, Branding,
// BrandingNeutral, and OperatorCSS. The caller fills in the rest. It
// checks the theme match, the palette's grammar, the branding, and the
// operator stylesheet in that order and returns the first refusal.
// stateDir is the directory the operator stylesheet must not sit
// under: the operator route is public, so a path beneath it would
// publish the store. An empty stateDir guards nothing.
func FromEnviron(environ []string, stateDir string) (Config, error) {
	match, err := config.ParseString(environ, themeMatchThreshold)
	if err != nil {
		return Config{}, err
	}
	if match != acceptedThemeMatch {
		return Config{}, fmt.Errorf("%s: %q: must be exactly %q", themeMatchThreshold.Key, match, acceptedThemeMatch)
	}
	palette, err := config.ParseString(environ, themePaletteThreshold)
	if err != nil {
		return Config{}, err
	}
	if !themePaletteGrammarRE.MatchString(palette) {
		return Config{}, fmt.Errorf("%s: %q: must match %s", themePaletteThreshold.Key, palette, themePaletteGrammarRE)
	}
	branding, neutral, err := resolveBranding(environ)
	if err != nil {
		return Config{}, err
	}
	operatorCSS, err := resolveOperatorCSS(environ, stateDir)
	if err != nil {
		return Config{}, err
	}
	return Config{
		Match:           match,
		Palette:         palette,
		Branding:        branding,
		BrandingNeutral: neutral,
		OperatorCSS:     operatorCSS,
	}, nil
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
	css, neutral, err := renderBranding(decls)
	if err != nil {
		return nil, "", fmt.Errorf("%s: %w", brandingThreshold.Key, err)
	}
	return css, neutral, nil
}

// resolveOperatorCSS reads the file DUBPLATE_OPERATOR_CSS_FILE names,
// returning nil when the key is unset, empty, or "none".
func resolveOperatorCSS(environ []string, stateDir string) ([]byte, error) {
	path, err := config.ParseString(environ, operatorCSSFileThreshold)
	if err != nil {
		return nil, err
	}
	if path == notConfigured {
		return nil, nil
	}
	css, err := readOperatorCSS(path, stateDir)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operatorCSSFileThreshold.Key, err)
	}
	return css, nil
}
