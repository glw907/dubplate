package unidecode

import (
	"strings"
	"testing"
)

func TestLexModule(t *testing.T) {
	tests := []struct {
		name      string
		src       string
		want      []*string
		errSubstr string
	}{
		{
			name: "empty tuple",
			src:  "data = (\n)\n",
			want: nil,
		},
		{
			name: "strings and none",
			src:  "data = (\n'a', None, '',\n)\n",
			want: []*string{new("a"), nil, new("")},
		},
		{
			name: "hex escape decodes to control bytes",
			src:  "data = (\n'" + `\x0a` + "', '" + `\x0a\x0a` + "',\n)\n",
			want: []*string{new("\n"), new("\n\n")},
		},
		{
			name: "double quoted string",
			src:  `data = (` + "\n\"''''\",\n)\n",
			want: []*string{new("''''")},
		},
		{
			name: "hash as a data entry, never a comment",
			src:  "data = (\n'#',\n)\n",
			want: []*string{new("#")},
		},
		{
			name: "comment lines and trailing per-entry comments",
			src:  "data = (\n# leading comment\n'a',    # 0x00\n)\n",
			want: []*string{new("a")},
		},
		{
			name: "escaped quote",
			src:  "data = (\n'it" + `\'` + "s',\n)\n",
			want: []*string{new("it's")},
		},
		{
			name: "escaped backslash",
			src:  "data = (\n'a" + `\\` + "b',\n)\n",
			want: []*string{new(`a\b`)},
		},
		{
			name:      "unknown escape",
			src:       "data = (\n'a" + `\n` + "b',\n)\n",
			errSubstr: "unknown escape",
		},
		{
			name:      "unterminated string",
			src:       "data = (\n'a\n)\n",
			errSubstr: "unterminated string",
		},
		{
			name:      "non-ASCII byte",
			src:       "data = (\n'café',\n)\n",
			errSubstr: "non-ASCII byte",
		},
		{
			name:      "carriage return",
			src:       "data = (\n'a\rb',\n)\n",
			errSubstr: "carriage return",
		},
		{
			name:      "missing comma between entries",
			src:       "data = (\n'a' 'b',\n)\n",
			errSubstr: `expected ","`,
		},
		{
			name:      "text after closing paren",
			src:       "data = (\n'a',\n)extra\n",
			errSubstr: "unexpected content after closing paren",
		},
		{
			name:      "wrong top-level identifier",
			src:       "table = (\n'a',\n)\n",
			errSubstr: `identifier "data"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := lexModule("test.py", []byte(tt.src))
			if tt.errSubstr != "" {
				if err == nil {
					t.Fatalf("lexModule(%q) = %v, nil, want error containing %q", tt.src, got, tt.errSubstr)
				}
				if !strings.Contains(err.Error(), tt.errSubstr) {
					t.Fatalf("lexModule(%q) error = %q, want it to contain %q", tt.src, err, tt.errSubstr)
				}
				return
			}
			if err != nil {
				t.Fatalf("lexModule(%q) unexpected error: %v", tt.src, err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("lexModule(%q) = %d entries, want %d", tt.src, len(got), len(tt.want))
			}
			for i := range got {
				g, w := got[i], tt.want[i]
				if (g == nil) != (w == nil) {
					t.Fatalf("entry %d = %v, want %v", i, g, w)
				}
				if g != nil && *g != *w {
					t.Fatalf("entry %d = %q, want %q", i, *g, *w)
				}
			}
		})
	}
}
