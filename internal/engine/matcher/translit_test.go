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
			if got := transliterate(tt.input); got != tt.want {
				t.Errorf("transliterate(%q) = %q, want %q", tt.input, got, tt.want)
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
			t.Fatal("expected transliterate to panic with nothing installed")
		}
		msg, ok := r.(string)
		if !ok || !strings.Contains(msg, "SetTransliterator") {
			t.Errorf("panic value %v does not name SetTransliterator", r)
		}
	}()
	transliterate("abc")
}
