package main

import (
	"slices"
	"strings"
	"testing"
)

func TestParsePackageTimes(t *testing.T) {
	output := strings.Join([]string{
		"ok  \tgithub.com/glw907/dubplate/internal/config\t0.412s",
		"ok  \tgithub.com/glw907/dubplate/internal/auth\t(cached)",
		"?   \tgithub.com/glw907/dubplate/internal/atomicfile\t[no test files]",
		"--- FAIL: TestThing (0.00s)",
		"FAIL\tgithub.com/glw907/dubplate/internal/serve\t1.204s",
		"ok  \tgithub.com/glw907/dubplate/internal/identity\t12.003s",
	}, "\n")

	got := parsePackageTimes(output)
	want := []packageTime{
		{name: "github.com/glw907/dubplate/internal/config", elapsed: "0.412s"},
		{name: "github.com/glw907/dubplate/internal/identity", elapsed: "12.003s"},
	}
	if !slices.Equal(got, want) {
		t.Errorf("parsePackageTimes = %v, want %v", got, want)
	}
}

func TestFilterSubjects(t *testing.T) {
	tests := []struct {
		name    string
		subject string
		want    bool
	}{
		{"refactor verb", "abc1234 Refactor the registry into one declare method", true},
		{"simplify verb", "def5678 Simplify the tagio write switch", true},
		{"simplified adjective", "aaa1111 Note the simplified seam in HISTORY", true},
		{"plain feature", "bbb2222 Add the magic-link verify handler", false},
		{"word contains factor", "ccc3333 Record the calibration factor", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterSubjects([]string{tt.subject})
			if (len(got) == 1) != tt.want {
				t.Errorf("filterSubjects(%q) matched %d, want match=%v", tt.subject, len(got), tt.want)
			}
		})
	}
}

func TestSplitPatterns(t *testing.T) {
	got := splitPatterns(" ./internal/serve, ./internal/auth ,, ")
	want := []string{"./internal/serve", "./internal/auth"}
	if !slices.Equal(got, want) {
		t.Errorf("splitPatterns = %v, want %v", got, want)
	}
}

func TestTouched(t *testing.T) {
	if got := touched(nil); got != "unchanged" {
		t.Errorf("touched(nil) = %q, want unchanged", got)
	}
	got := touched([]string{"internal/serve/serve.go", "internal/serve/strip.go"})
	if !strings.HasPrefix(got, "changed (2 files): internal/serve/serve.go, ") {
		t.Errorf("touched(two) = %q", got)
	}
}
