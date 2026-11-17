package webapp

import (
	"os/exec"
	"strings"
	"testing"
)

// TestLayering proves the package's transitive imports reach none of
// the packages that sit above the web app: the admin gate in
// internal/auth wraps its handlers from cmd/dubplate-server, and the
// config keys it serves are resolved there into plain values.
func TestLayering(t *testing.T) {
	const module = "github.com/glw907/dubplate/internal/"
	out, err := exec.CommandContext(t.Context(), "go", "list", "-deps", ".").Output()
	if err != nil {
		t.Fatalf("go list -deps .: %v", err)
	}
	deps := strings.Fields(string(out))
	if len(deps) == 0 {
		t.Fatal("go list -deps . listed nothing")
	}
	for _, dep := range deps {
		for _, forbidden := range []string{module + "auth", module + "ops", module + "settings", module + "config"} {
			if dep == forbidden {
				t.Errorf("webapp depends on %s", dep)
			}
		}
		if dep == module+"engine" || strings.HasPrefix(dep, module+"engine/") {
			t.Errorf("webapp depends on %s", dep)
		}
	}
}
