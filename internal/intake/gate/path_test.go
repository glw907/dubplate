package gate

import "testing"

func TestValidPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want bool
	}{
		{"a file", "01 Track.flac", true},
		{"nested", "Artist/Album/01 Track.flac", true},
		{"a dot inside a name", "a.b/c..d/.hidden", true},
		{"bytes that are not UTF-8", "caf\xe9/01.flac", true},
		{"absolute", "/a/b", false},
		{"empty component", "a//b", false},
		{"dot component", "a/./b", false},
		{"dotdot component", "a/../b", false},
		{"leading dotdot", "../a", false},
		{"dotdot alone", "..", false},
		{"dot alone", ".", false},
		{"NUL", "a\x00b", false},
		{"backslash", `a\b`, false},
		{"trailing slash", "a/b/", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidPath([]byte(tt.path)); got != tt.want {
				t.Errorf("ValidPath(%q) = %t, want %t", tt.path, got, tt.want)
			}
		})
	}
}
