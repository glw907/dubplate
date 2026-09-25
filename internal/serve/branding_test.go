package serve

import (
	"regexp"
	"slices"
	"strings"
	"testing"
)

// themeVariableNames is the design language's complete theme-variable
// set, copied from the standard rather than from brandingKinds,
// so a name the code adds or drops shows up as a difference.
var themeVariableNames = []string{
	"color-scheme",
	"--color-base-100", "--color-base-200", "--color-base-300",
	"--color-base-content",
	"--color-primary", "--color-primary-content",
	"--color-secondary", "--color-secondary-content",
	"--color-accent", "--color-accent-content",
	"--color-neutral", "--color-neutral-content",
	"--color-info", "--color-info-content",
	"--color-success", "--color-success-content",
	"--color-warning", "--color-warning-content",
	"--color-error", "--color-error-content",
	"--radius-selector", "--radius-field", "--radius-box",
	"--size-selector", "--size-field",
	"--border", "--depth", "--noise",
	"--font-sans", "--font-serif", "--font-mono",
}

func TestBrandingStylesheet_Renders(t *testing.T) {
	tests := []struct {
		name  string
		decls string
		want  string
	}{
		{name: "nothing configured", decls: "", want: ""},
		{
			name:  "one pair",
			decls: "--color-primary: #3b5bdb",
			want:  ":root {\n\t--color-primary: #3b5bdb;\n}\n",
		},
		{
			name: "many pairs, one of every kind, spacing normalized",
			decls: "--color-primary:oklch(45% 0.24 277.023) ;  --color-base-100: rgb(255 250 240 / 90%);" +
				"--color-accent: hsl(-30deg 50% 40%);--color-info: #1c7ed6ff; color-scheme: dark;" +
				" --radius-box: 0rem; --border: 2px; --size-field: .25em; --depth: 0; --noise: 1;" +
				"--font-sans: Inter, ui-sans-serif, system-ui, sans-serif; --font-mono: Fira Code,monospace",
			want: ":root {\n" +
				"\t--color-primary: oklch(45% 0.24 277.023);\n" +
				"\t--color-base-100: rgb(255 250 240 / 90%);\n" +
				"\t--color-accent: hsl(-30deg 50% 40%);\n" +
				"\t--color-info: #1c7ed6ff;\n" +
				"\tcolor-scheme: dark;\n" +
				"\t--radius-box: 0rem;\n" +
				"\t--border: 2px;\n" +
				"\t--size-field: .25em;\n" +
				"\t--depth: 0;\n" +
				"\t--noise: 1;\n" +
				"\t--font-sans: Inter, ui-sans-serif, system-ui, sans-serif;\n" +
				"\t--font-mono: Fira Code,monospace;\n" +
				"}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BrandingStylesheet(tt.decls)
			if err != nil {
				t.Fatalf("BrandingStylesheet: %v", err)
			}
			if string(got) != tt.want {
				t.Errorf("BrandingStylesheet =\n%q\nwant\n%q", got, tt.want)
			}
		})
	}
}

// validPair sits beside every refused declaration below, so each row's
// refusal belongs to the bad declaration and never to an empty list.
const validPair = "--color-base-100: #ffffff"

// value64 is a valid color exactly 64 bytes long, the bound's edge.
var value64 = "oklch(45." + strings.Repeat("0", 46) + "% 0.24 1)"

func TestBrandingStylesheet_Refusals(t *testing.T) {
	if len(value64) != 64 {
		t.Fatalf("value64 is %d bytes, want 64", len(value64))
	}
	value65 := strings.Replace(value64, "45.", "45.0", 1)

	tests := []struct {
		name string
		// bad is the refused declaration; the input is validPair; bad.
		bad string
		// refused is the operator text the error must never echo;
		// empty when the refused value is itself empty.
		refused    string
		wantErrHas string
	}{
		// The injection table.
		{name: "closing brace", bad: "--color-primary: #fff}", refused: "#fff}", wantErrHas: `--color-primary: value holds "}"`},
		{name: "opening brace", bad: "--color-primary: #fff{", refused: "#fff{", wantErrHas: `--color-primary: value holds "{"`},
		{name: "semicolon inside a value", bad: "--color-primary: #fff;color:red", refused: "#fff;color:red", wantErrHas: `declaration 3: "color" is not a theme variable`},
		{name: "url()", bad: "--color-primary: url(https://example.invalid/x)", refused: "url(https://example.invalid/x)", wantErrHas: `--color-primary: value holds "url("`},
		{name: "var()", bad: "--color-primary: var(--x)", refused: "var(--x)", wantErrHas: `--color-primary: value holds "var("`},
		{name: "expression()", bad: "--color-primary: expression(alert(1))", refused: "expression(alert(1))", wantErrHas: `--color-primary: value holds "expression("`},
		{name: "closing style tag", bad: "--color-primary: </style>", refused: "</style>", wantErrHas: `--color-primary: value holds "<"`},
		{name: "backslash escape", bad: `--color-primary: #f\66f`, refused: `#f\66f`, wantErrHas: `--color-primary: value holds "\\"`},
		{name: "control character", bad: "--color-primary: #ff\x01f", refused: "#ff\x01f", wantErrHas: "--color-primary: value holds a control character"},
		{name: "DEL character", bad: "--color-primary: #ff\x7ff", refused: "#ff\x7ff", wantErrHas: "--color-primary: value holds a control character"},
		{name: "65-byte value", bad: "--color-primary: " + value65, refused: value65, wantErrHas: "--color-primary: value is over 64 bytes"},
		{name: "at sign", bad: "--font-sans: @import", refused: "@import", wantErrHas: `--font-sans: value holds "@"`},
		{name: "single quote", bad: "--font-sans: 'Inter'", refused: "'Inter'", wantErrHas: `--font-sans: value holds "'"`},
		{name: "double quote", bad: `--font-sans: "Inter"`, refused: `"Inter"`, wantErrHas: `--font-sans: value holds "\""`},

		// The state table.
		{name: "repeated name", bad: "--color-base-100: #000000", refused: "#000000", wantErrHas: `declaration 2: "--color-base-100" is declared more than once`},
		{name: "empty pair from a doubled semicolon", bad: "; --color-primary: #000", refused: "#000", wantErrHas: "declaration 2: empty"},
		{name: "empty pair from a trailing semicolon", bad: "--color-primary: #000;", refused: "#000", wantErrHas: "declaration 3: empty"},
		{name: "pair with no colon", bad: "--color-primary #000", refused: "#000", wantErrHas: `declaration 2: no ":"`},
		{name: "name outside the set", bad: "--color-custom: #000", refused: "#000", wantErrHas: `declaration 2: "--color-custom" is not a theme variable`},
		{name: "name differing in case", bad: "--Color-Primary: #000", refused: "#000", wantErrHas: `"--Color-Primary" is not a theme variable`},

		// A value outside its name's grammar, one row per kind.
		{name: "color by name", bad: "--color-primary: red", refused: "red", wantErrHas: "--color-primary: value is not a hex color"},
		{name: "hex of five digits", bad: "--color-primary: #12345", refused: "#12345", wantErrHas: "--color-primary: value is not a hex color"},
		{name: "color function outside the three", bad: "--color-primary: lab(50% 1 1)", refused: "lab(50% 1 1)", wantErrHas: "--color-primary: value is not a hex color"},
		{name: "color function with a comma", bad: "--color-primary: rgb(1, 2, 3)", refused: "rgb(1, 2, 3)", wantErrHas: "--color-primary: value is not a hex color"},
		{name: "color function with a keyword", bad: "--color-primary: rgb(1 2 calc)", refused: "rgb(1 2 calc)", wantErrHas: "--color-primary: value is not a hex color"},
		{name: "color scheme outside the two", bad: "color-scheme: light dark", refused: "light dark", wantErrHas: `color-scheme: value is not "light" or "dark"`},
		{name: "length with no unit", bad: "--radius-box: 4", refused: "4", wantErrHas: "--radius-box: value is not a number with a rem, px, or em unit"},
		{name: "length with a percent", bad: "--border: 10%", refused: "10%", wantErrHas: "--border: value is not a number"},
		{name: "negative length", bad: "--size-field: -1px", refused: "-1px", wantErrHas: "--size-field: value is not a number"},
		{name: "flag outside 0 and 1", bad: "--depth: 7", refused: "7", wantErrHas: `--depth: value is not "0" or "1"`},
		{name: "font stack with a paren", bad: "--font-mono: mono()", refused: "mono()", wantErrHas: "--font-mono: value is not a comma-separated list"},
		{name: "empty value", bad: "--noise:", refused: "", wantErrHas: `--noise: value is not "0" or "1"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := BrandingStylesheet(validPair + "; " + tt.bad)
			if err == nil {
				t.Fatalf("BrandingStylesheet: got nil error and %q, want a refusal containing %q", got, tt.wantErrHas)
			}
			if !strings.Contains(err.Error(), tt.wantErrHas) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErrHas)
			}
			if tt.refused != "" && strings.Contains(err.Error(), tt.refused) {
				t.Errorf("error = %q echoes the refused value %q", err.Error(), tt.refused)
			}
		})
	}
}

// TestBrandingStylesheet_ValueAtTheBound pairs with the 65-byte row
// above: a valid value of exactly the bound is admitted.
func TestBrandingStylesheet_ValueAtTheBound(t *testing.T) {
	if len(value64) != 64 {
		t.Fatalf("value64 is %d bytes, want 64", len(value64))
	}
	got, err := BrandingStylesheet(validPair + "; --color-primary: " + value64)
	if err != nil {
		t.Fatalf("BrandingStylesheet: %v", err)
	}
	if want := "\t--color-primary: " + value64 + ";\n"; !strings.Contains(string(got), want) {
		t.Errorf("stylesheet = %q, want it to carry %q", got, want)
	}
}

// declaredNameRE reads one declaration's name out of a rendered
// stylesheet line.
var declaredNameRE = regexp.MustCompile(`(?m)^\t([^:]+):`)

// TestBrandingStylesheet_G8 is the design language's branding-surface
// gate: the stylesheet declares only names from the theme-variable
// set, every name in that set is accepted, and nothing configured
// renders an empty file.
func TestBrandingStylesheet_G8(t *testing.T) {
	empty, err := BrandingStylesheet("")
	if err != nil {
		t.Fatalf("BrandingStylesheet(\"\"): %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("BrandingStylesheet(\"\") = %q, want empty", empty)
	}

	// Every name in the set, with a value its kind admits.
	valueFor := func(name string) string {
		switch {
		case name == "color-scheme":
			return "light"
		case strings.HasPrefix(name, "--color-"):
			return "#123456"
		case name == "--depth", name == "--noise":
			return "1"
		case strings.HasPrefix(name, "--font-"):
			return "system-ui, sans-serif"
		default:
			return "0.5rem"
		}
	}
	decls := make([]string, 0, len(themeVariableNames))
	for _, name := range themeVariableNames {
		decls = append(decls, name+": "+valueFor(name))
	}
	full, err := BrandingStylesheet(strings.Join(decls, "; "))
	if err != nil {
		t.Fatalf("BrandingStylesheet(every name): %v", err)
	}
	var declared []string
	for _, m := range declaredNameRE.FindAllStringSubmatch(string(full), -1) {
		declared = append(declared, m[1])
	}
	if !slices.Equal(declared, themeVariableNames) {
		t.Errorf("declared names = %v, want the 32-name set %v", declared, themeVariableNames)
	}

	// Names a stylesheet could plausibly carry and must not: DaisyUI's
	// internal variables, a plain property, and a near miss.
	for _, name := range []string{"--tab-radius", "--root-bg", "background", "color", "--color-base-400", "--font-display", "--radius"} {
		t.Run(name, func(t *testing.T) {
			got, err := BrandingStylesheet(validPair + "; " + name + ": " + valueFor(name))
			if err == nil {
				t.Fatalf("BrandingStylesheet admitted %q and rendered %q", name, got)
			}
			if want := `"` + name + `" is not a theme variable`; !strings.Contains(err.Error(), want) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), want)
			}
		})
	}
}
