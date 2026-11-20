package webapp

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

func TestFromEnviron_Defaults(t *testing.T) {
	got, err := FromEnviron(nil, t.TempDir())
	if err != nil {
		t.Fatalf("FromEnviron(nil): %v", err)
	}
	if got.Match != "dubplate" || got.Palette != "system" {
		t.Errorf("Match, Palette = %q, %q; want %q, %q", got.Match, got.Palette, "dubplate", "system")
	}
	if len(got.Branding) != 0 || got.BrandingNeutral != "" || got.OperatorCSS != nil {
		t.Errorf("unset branding and operator CSS resolved to %q, %q, %q; want all empty", got.Branding, got.BrandingNeutral, got.OperatorCSS)
	}
}

func TestFromEnviron_BrandingNeutral(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{"declared", "--color-neutral: #123456", "#123456"},
		{"among others", "--color-primary: #abcdef; --color-neutral: #123456", "#123456"},
		{"absent", "--color-primary: #abcdef", ""},
		{"none", "none", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromEnviron([]string{"DUBPLATE_BRANDING=" + tt.value}, "")
			if err != nil {
				t.Fatalf("FromEnviron: %v", err)
			}
			if got.BrandingNeutral != tt.want {
				t.Errorf("BrandingNeutral = %q, want %q", got.BrandingNeutral, tt.want)
			}
		})
	}
}

func TestFromEnviron_BrandingStylesheet(t *testing.T) {
	got, err := FromEnviron([]string{"DUBPLATE_BRANDING=--color-primary: #abcdef"}, "")
	if err != nil {
		t.Fatalf("FromEnviron: %v", err)
	}
	if !strings.Contains(string(got.Branding), "--color-primary: #abcdef") {
		t.Errorf("Branding = %q, want it to carry the declaration", got.Branding)
	}
}

func TestFromEnviron_Refusals(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    string
	}{
		{"theme match", []string{"DUBPLATE_THEME_MATCH=other"}, `DUBPLATE_THEME_MATCH: "other": must be exactly "dubplate"`},
		{"palette grammar", []string{"DUBPLATE_THEME_PALETTE=Bad Palette"}, `DUBPLATE_THEME_PALETTE: "Bad Palette": must match ^[a-z0-9-]+$`},
		{"branding", []string{"DUBPLATE_BRANDING=not a declaration"}, "DUBPLATE_BRANDING: "},
		{"operator css", []string{"DUBPLATE_OPERATOR_CSS_FILE=/nonexistent/operator.css"}, "DUBPLATE_OPERATOR_CSS_FILE: "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := FromEnviron(tt.environ, t.TempDir())
			if err == nil || !strings.HasPrefix(err.Error(), tt.want) {
				t.Errorf("error = %v, want one starting %q", err, tt.want)
			}
		})
	}
}

// A bad theme match and a bad palette in one environment report the
// match: the reader checks it first.
func TestFromEnviron_MatchIsCheckedBeforePalette(t *testing.T) {
	environ := []string{"DUBPLATE_THEME_PALETTE=Bad Palette", "DUBPLATE_THEME_MATCH=other"}
	_, err := FromEnviron(environ, "")
	const want = `DUBPLATE_THEME_MATCH: "other": must be exactly "dubplate"`
	if err == nil || err.Error() != want {
		t.Errorf("error = %v, want %q", err, want)
	}
}

func TestFromEnviron_OperatorCSS(t *testing.T) {
	path := filepath.Join(t.TempDir(), "operator.css")
	if err := os.WriteFile(path, []byte("body{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	got, err := FromEnviron([]string{"DUBPLATE_OPERATOR_CSS_FILE=" + path}, t.TempDir())
	if err != nil {
		t.Fatalf("FromEnviron: %v", err)
	}
	if string(got.OperatorCSS) != "body{}" {
		t.Errorf("OperatorCSS = %q, want %q", got.OperatorCSS, "body{}")
	}
}

func TestFromEnviron_OperatorCSSUnderStateDir(t *testing.T) {
	state := t.TempDir()
	path := filepath.Join(state, "operator.css")
	if err := os.WriteFile(path, []byte("body{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := FromEnviron([]string{"DUBPLATE_OPERATOR_CSS_FILE=" + path}, state)
	if err == nil || !strings.HasPrefix(err.Error(), "DUBPLATE_OPERATOR_CSS_FILE: ") {
		t.Errorf("error = %v, want a DUBPLATE_OPERATOR_CSS_FILE refusal", err)
	}
}

func TestRegister(t *testing.T) {
	reg := config.NewRegistry()
	Register(reg)
	var got []string
	for _, th := range reg.Thresholds() {
		got = append(got, th.Key)
	}
	want := []string{"DUBPLATE_THEME_MATCH", "DUBPLATE_THEME_PALETTE", "DUBPLATE_BRANDING", "DUBPLATE_OPERATOR_CSS_FILE"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("registered thresholds = %v, want %v", got, want)
	}
}
