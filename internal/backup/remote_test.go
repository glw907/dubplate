package backup

import (
	"strings"
	"testing"
)

// TestParseRemote pins the validator's refusals and the join rule: the
// value splits at its first colon, and each destination is the name, a
// colon, and the path joined with its sub-path, so an empty path never
// reads as the backend's root.
func TestParseRemote(t *testing.T) {
	tests := []struct {
		in      string
		wantErr string
		want    string
	}{
		{in: "name:", want: "name:library"},
		{in: "name:dir", want: "name:dir/library"},
		{in: "name:dir/", want: "name:dir/library"},
		{in: "name:/abs", want: "name:/abs/library"},
		{in: "b2_1:bucket/sub", want: "b2_1:bucket/sub/library"},
		{in: "name:a:b", want: "name:a:b/library"},
		{in: "relative/dir", wantErr: "no colon"},
		{in: "", wantErr: "no colon"},
		{in: ":dir", wantErr: `name ""`},
		{in: "my-remote:dir", wantErr: `name "my-remote"`},
		{in: "a b:dir", wantErr: `name "a b"`},
		{in: "/abs/path:dir", wantErr: `name "/abs/path"`},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			r, err := ParseRemote(tt.in)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("ParseRemote(%q) = %v, want an error naming %q", tt.in, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseRemote(%q): %v", tt.in, err)
			}
			if got := r.join("library"); got != tt.want {
				t.Errorf("join(library) = %q, want %q", got, tt.want)
			}
		})
	}
}
