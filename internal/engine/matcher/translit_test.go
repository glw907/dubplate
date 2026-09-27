package matcher

import (
	"strings"
	"testing"
)

func TestTransliterate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "empty string", input: "", want: ""},
		{name: "already ascii", input: "hello", want: "hello"},
		{name: "cafe with e-acute (U+00E9)", input: "caf" + string(rune(0xe9)), want: "cafe"},
		{
			name:  "an unmapped codepoint contributes nothing",
			input: "a" + string(rune(0x241)) + "b",
			want:  "ab",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Transliterate(tt.input); got != tt.want {
				t.Errorf("Transliterate(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// TestTransliterateUninstalledPanics proves that a program which
// never calls SetTransliterator gets a panic naming it, on every
// input, rather than a silent ASCII-only fallback. It clears and
// restores the package-internal holder directly, since this test file
// is package matcher, never through SetTransliterator's own exported
// path (which refuses a nil install).
func TestTransliterateUninstalledPanics(t *testing.T) {
	saved := transliterator.Load()
	transliterator.Store(nil)
	defer transliterator.Store(saved)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected Transliterate to panic with nothing installed")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "SetTransliterator") {
			t.Errorf("panic value %v does not name SetTransliterator", r)
		}
	}()
	Transliterate("abc")
}

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{name: "both empty", a: "", b: "", want: 0},
		{name: "empty vs non-empty", a: "", b: "abc", want: 3},
		{name: "non-empty vs empty", a: "abc", b: "", want: 3},
		{name: "equal-length identical", a: "kitten", b: "kitten", want: 0},
		{name: "equal-length one substitution", a: "abc", b: "abd", want: 1},
		{name: "equal-length every position differs", a: "abc", b: "xyz", want: 3},
		{name: "classic kitten/sitting", a: "kitten", b: "sitting", want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Levenshtein(tt.a, tt.b); got != tt.want {
				t.Errorf("Levenshtein(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}
