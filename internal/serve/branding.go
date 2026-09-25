package serve

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// brandingValueMaxBytes bounds one branding value. The longest value
// DaisyUI's own theme files write, an oklch() color with three decimal
// arguments, is under 30 bytes, so the bound admits any real value
// with room to spare and refuses anything shaped like a payload.
const brandingValueMaxBytes = 64

// brandingKind names the value grammar one branding name accepts.
type brandingKind int

const (
	kindColor brandingKind = iota
	kindColorScheme
	kindLength
	kindFlag
	kindFontStack
)

// brandingKinds is the design language's 32-name theme-variable set,
// each name mapped to its value grammar. The kinds follow DaisyUI 5's
// own theme files: every --color-* value there is an oklch() color,
// every radius, size, and border value a number with a rem or px unit
// ("--radius-box: 0rem", "--border: 2px"), --depth and --noise are 0
// or 1, and color-scheme is light or dark. The three font names are
// Tailwind 4's font stacks.
var brandingKinds = map[string]brandingKind{
	"color-scheme":              kindColorScheme,
	"--color-base-100":          kindColor,
	"--color-base-200":          kindColor,
	"--color-base-300":          kindColor,
	"--color-base-content":      kindColor,
	"--color-primary":           kindColor,
	"--color-primary-content":   kindColor,
	"--color-secondary":         kindColor,
	"--color-secondary-content": kindColor,
	"--color-accent":            kindColor,
	"--color-accent-content":    kindColor,
	"--color-neutral":           kindColor,
	"--color-neutral-content":   kindColor,
	"--color-info":              kindColor,
	"--color-info-content":      kindColor,
	"--color-success":           kindColor,
	"--color-success-content":   kindColor,
	"--color-warning":           kindColor,
	"--color-warning-content":   kindColor,
	"--color-error":             kindColor,
	"--color-error-content":     kindColor,
	"--radius-selector":         kindLength,
	"--radius-field":            kindLength,
	"--radius-box":              kindLength,
	"--size-selector":           kindLength,
	"--size-field":              kindLength,
	"--border":                  kindLength,
	"--depth":                   kindFlag,
	"--noise":                   kindFlag,
	"--font-sans":               kindFontStack,
	"--font-serif":              kindFontStack,
	"--font-mono":               kindFontStack,
}

// The value grammar, one expression per kind. Each is anchored at both
// ends, so a value is admitted whole or not at all.
var (
	hexColorRE  = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3,4}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)
	funcColorRE = regexp.MustCompile(
		`^(?:oklch|rgb|hsl)\(\s*` + colorArg + `(?:(?:\s+|\s*/\s*)` + colorArg + `)*\s*\)$`)
	lengthRE    = regexp.MustCompile(`^` + number + `(?:rem|px|em)$`)
	fontStackRE = regexp.MustCompile(`^` + fontFamily + `(?:,\s*` + fontFamily + `)*$`)
)

// The pieces the value grammar is built from: an unsigned decimal, a
// color function argument (a possibly negative decimal with an
// optional % or deg), and an unquoted font family of one or more
// identifiers.
const (
	number     = `(?:[0-9]+(?:\.[0-9]+)?|\.[0-9]+)`
	colorArg   = `-?` + number + `(?:%|deg)?`
	fontFamily = `[A-Za-z][A-Za-z0-9-]*(?: [A-Za-z][A-Za-z0-9-]*)*`
)

// brandingRefusedSequences may appear nowhere in a value, whatever its
// grammar. The grammar alone already refuses every one; the list stands
// in front of it so a later widening of one grammar cannot admit a
// sequence that closes the block, opens a string, or fetches a
// resource, and so a refusal names the sequence rather than the
// grammar. ";" is absent because it separates declarations and so
// never reaches a value.
var brandingRefusedSequences = []string{
	"{", "}", "<", ">", `"`, "'", `\`, "@", "url(", "var(", "expression(",
}

// BrandingStylesheet renders decls, the operator's branding value, as
// one :root block. decls is a ";"-separated list of "name: value"
// declarations; an empty decls renders an empty stylesheet.
//
// Every name must be one of the theme variables and appear once, and
// every value must match its name's grammar and hold none of the
// refused sequences. Anything else is refused whole: a value is never
// escaped into place, because an escaped value that still parses as
// CSS carries the operator's text into the stylesheet all the same. A
// returned error names the declaration's position, its name when the
// name is the fault, and the rule broken, but never the refused value.
func BrandingStylesheet(decls string) ([]byte, error) {
	if decls == "" {
		return []byte{}, nil
	}

	var buf bytes.Buffer
	buf.WriteString(":root {\n")
	seen := make(map[string]bool)
	for i, decl := range strings.Split(decls, ";") {
		pos := i + 1
		rawName, rawValue, ok := strings.Cut(decl, ":")
		name := strings.TrimSpace(rawName)
		value := strings.TrimSpace(rawValue)
		switch {
		case strings.TrimSpace(decl) == "":
			return nil, fmt.Errorf(`declaration %d: empty; remove the doubled or trailing ";"`, pos)
		case !ok:
			return nil, fmt.Errorf(`declaration %d: no ":" between name and value`, pos)
		}
		kind, known := brandingKinds[name]
		if !known {
			return nil, fmt.Errorf("declaration %d: %q is not a theme variable", pos, name)
		}
		if seen[name] {
			return nil, fmt.Errorf("declaration %d: %q is declared more than once", pos, name)
		}
		seen[name] = true
		if err := checkBrandingValue(kind, value); err != nil {
			return nil, fmt.Errorf("declaration %d: %s: %w", pos, name, err)
		}
		fmt.Fprintf(&buf, "\t%s: %s;\n", name, value)
	}
	buf.WriteString("}\n")
	return buf.Bytes(), nil
}

// checkBrandingValue reports why value is refused for a name of kind,
// or nil when it is admitted. The error never carries value itself.
func checkBrandingValue(kind brandingKind, value string) error {
	if len(value) > brandingValueMaxBytes {
		return fmt.Errorf("value is over %d bytes", brandingValueMaxBytes)
	}
	if strings.ContainsFunc(value, isControl) {
		return errors.New("value holds a control character")
	}
	for _, seq := range brandingRefusedSequences {
		if strings.Contains(value, seq) {
			return fmt.Errorf("value holds %q, which is refused", seq)
		}
	}

	switch kind {
	case kindColor:
		if !hexColorRE.MatchString(value) && !funcColorRE.MatchString(value) {
			return errors.New("value is not a hex color or an oklch(), rgb(), or hsl() color")
		}
	case kindColorScheme:
		if value != "light" && value != "dark" {
			return errors.New(`value is not "light" or "dark"`)
		}
	case kindLength:
		if !lengthRE.MatchString(value) {
			return errors.New("value is not a number with a rem, px, or em unit")
		}
	case kindFlag:
		if value != "0" && value != "1" {
			return errors.New(`value is not "0" or "1"`)
		}
	case kindFontStack:
		if !fontStackRE.MatchString(value) {
			return errors.New("value is not a comma-separated list of unquoted font families")
		}
	}
	return nil
}

// isControl reports whether r is an ASCII control character, DEL
// included.
func isControl(r rune) bool {
	return r < 0x20 || r == 0x7f
}
