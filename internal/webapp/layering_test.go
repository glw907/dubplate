package webapp

import (
	"os/exec"
	"strings"
	"testing"
)

// TestLayering proves each listed package's transitive imports reach
// none of the module packages it must sit below. The web app sits below
// the admin gate in internal/auth, which wraps its handlers from
// cmd/dubplate-server; it declares its own settings, so internal/config
// is allowed. A leaf reaches no module package at all.
func TestLayering(t *testing.T) {
	const module = "github.com/glw907/dubplate/internal/"
	tests := []struct {
		name string
		dir  string
		// leaf forbids every module package but the one under test.
		leaf      bool
		forbidden []string
		// forbiddenTrees forbids a package and everything under it.
		forbiddenTrees []string
	}{
		{
			name:           "webapp",
			dir:            ".",
			forbidden:      []string{"auth", "ops", "settings", "store"},
			forbiddenTrees: []string{"engine"},
		},
		{name: "fsys", dir: "../fsys", leaf: true},
		{name: "doctor", dir: "../doctor", leaf: true},
		{name: "auth", dir: "../auth", forbidden: []string{"serve"}},
		{name: "member", dir: "../member", forbidden: []string{"auth", "serve"}},
		{
			name:           "settings",
			dir:            "../settings",
			forbidden:      []string{"auth", "serve", "member", "intake", "backup", "notify"},
			forbiddenTrees: []string{"ops"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := exec.CommandContext(t.Context(), "go", "list", "-deps", tt.dir).Output()
			if err != nil {
				t.Fatalf("go list -deps %s: %v", tt.dir, err)
			}
			deps := strings.Fields(string(out))
			if len(deps) == 0 {
				t.Fatalf("go list -deps %s listed nothing", tt.dir)
			}
			self := module + tt.name
			for _, dep := range deps {
				if tt.leaf && dep != self && strings.HasPrefix(dep, module) {
					t.Errorf("%s depends on %s", tt.name, dep)
				}
				for _, f := range tt.forbidden {
					if dep == module+f {
						t.Errorf("%s depends on %s", tt.name, dep)
					}
				}
				for _, f := range tt.forbiddenTrees {
					if dep == module+f || strings.HasPrefix(dep, module+f+"/") {
						t.Errorf("%s depends on %s", tt.name, dep)
					}
				}
			}
		})
	}
}
