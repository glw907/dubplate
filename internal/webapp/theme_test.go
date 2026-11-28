package webapp

import (
	"io/fs"
	"slices"
	"testing"

	"github.com/glw907/dubplate/web"
)

func TestStampablePalettes(t *testing.T) {
	selectors := map[string]struct{}{
		"dubplate-label":       {},
		"dubplate-label-night": {},
		"dubplate-system":      {},
		"light":                {}, // no "dubplate-" prefix: DaisyUI's own bundled default, must be excluded
		"dark":                 {},
	}
	got := stampablePalettes(selectors, "dubplate")
	want := []string{"label", "label-night", "system"}
	if len(got) != len(want) {
		t.Fatalf("stampablePalettes = %v, want %v", got, want)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("stampablePalettes[%d] = %q, want %q", i, got[i], w)
		}
	}
}

// TestStampablePalettes_RealBundle proves the stampable set the real
// built bundle's login documents yield is exactly {label, label-night,
// system}, the set cmd/dubplate-server's real-bundle stamp test reads
// its cookie rows against.
func TestStampablePalettes_RealBundle(t *testing.T) {
	const loginDocumentName, loginLinkDocumentName = "login.html", "login/link.html"
	buildFS := web.Build
	loginRaw, err := fs.ReadFile(buildFS, loginDocumentName)
	if err != nil {
		t.Fatalf("fs.ReadFile(%q): %v", loginDocumentName, err)
	}
	linkRaw, err := fs.ReadFile(buildFS, loginLinkDocumentName)
	if err != nil {
		t.Fatalf("fs.ReadFile(%q): %v", loginLinkDocumentName, err)
	}
	themes, err := extractThemes(buildFS,
		namedDocument{path: loginDocumentName, raw: loginRaw},
		namedDocument{path: loginLinkDocumentName, raw: linkRaw},
	)
	if err != nil {
		t.Fatalf("extractThemes: %v", err)
	}
	gotPalettes := stampablePalettes(themes.selectors, "dubplate")
	wantPalettes := []string{"label", "label-night", "system"}
	if !slices.Equal(gotPalettes, wantPalettes) {
		t.Fatalf("stampablePalettes(real bundle) = %v, want %v", gotPalettes, wantPalettes)
	}
}

func TestThemeNeutrals(t *testing.T) {
	tests := []struct {
		name string
		css  string
		want map[string][]themeColor
	}{
		{
			name: "one rule with its scheme",
			css:  `[data-theme=dubplate-label]{color-scheme:light;--color-neutral:#1f1b16;--color-neutral-content:#fff}`,
			want: map[string][]themeColor{"dubplate-label": {{scheme: "light", color: "#1f1b16"}}},
		},
		{
			name: "a media branch is read on its own selector",
			css: `[data-theme="dubplate-system"]{color-scheme: light; --color-neutral: #1f1b16}` +
				`@media (prefers-color-scheme:dark){[data-theme="dubplate-system"]{color-scheme:dark;--color-neutral:#110f0d}}`,
			want: map[string][]themeColor{"dubplate-system": {
				{scheme: "light", color: "#1f1b16"},
				{scheme: "dark", color: "#110f0d"},
			}},
		},
		{
			name: "a rule with no color-scheme",
			css:  `[data-theme=x]{--color-neutral:oklch(20% 0 0)}`,
			want: map[string][]themeColor{"x": {{color: "oklch(20% 0 0)"}}},
		},
		{
			name: "neutral-content alone is not the neutral",
			css:  `[data-theme=x]{color-scheme:dark;--color-neutral-content:#fff}`,
			want: map[string][]themeColor{},
		},
		{
			name: "a rule naming no theme",
			css:  `:root{--color-neutral:#000}`,
			want: map[string][]themeColor{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := themeNeutrals([]byte(tt.css))
			if len(got) != len(tt.want) {
				t.Fatalf("themeNeutrals = %v, want %v", got, tt.want)
			}
			for name, want := range tt.want {
				if !slices.Equal(got[name], want) {
					t.Errorf("themeNeutrals[%q] = %v, want %v", name, got[name], want)
				}
			}
		})
	}
}

func TestThemeSelectors(t *testing.T) {
	tests := []struct {
		name string
		css  string
		want map[string]struct{}
	}{
		{
			name: "unquoted",
			css:  `[data-theme=dubplate-label]{color:red}`,
			want: map[string]struct{}{"dubplate-label": {}},
		},
		{
			name: "quoted",
			css:  `[data-theme="dubplate-label-night"]{color:red}`,
			want: map[string]struct{}{"dubplate-label-night": {}},
		},
		{
			name: "inside a compound selector list",
			css:  `:is(:root:has(input.theme-controller[value=dubplate-system]:checked),[data-theme=dubplate-system]){}`,
			want: map[string]struct{}{"dubplate-system": {}},
		},
		{
			name: "a prefix is never captured as the longer name",
			css:  `[data-theme=dubplate-label]{}[data-theme=dubplate-label-night]{}`,
			want: map[string]struct{}{"dubplate-label": {}, "dubplate-label-night": {}},
		},
		{
			name: "no selector at all",
			css:  `body{color:red}`,
			want: map[string]struct{}{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := themeSelectors([]byte(tt.css))
			if len(got) != len(tt.want) {
				t.Fatalf("themeSelectors = %v, want %v", got, tt.want)
			}
			for name := range tt.want {
				if _, ok := got[name]; !ok {
					t.Errorf("themeSelectors: missing %q", name)
				}
			}
		})
	}
}
