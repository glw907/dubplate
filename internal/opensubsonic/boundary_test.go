package opensubsonic

import (
	"os/exec"
	"slices"
	"strings"
	"testing"
)

const (
	modulePrefix = "github.com/glw907/dubplate/"
	layerRoot    = modulePrefix + "internal/opensubsonic"
)

// violations returns the module-internal paths in a `go list -deps -test`
// listing that lie outside the protocol layer. Each line is cut at its first
// space, which drops the test-variant annotation.
func violations(listing string) []string {
	var bad []string
	for line := range strings.Lines(listing) {
		path, _, _ := strings.Cut(strings.TrimSpace(line), " ")
		path = strings.TrimSuffix(path, ".test")
		if !strings.HasPrefix(path, modulePrefix) {
			continue
		}
		if path == layerRoot || strings.HasPrefix(path, layerRoot+"/") {
			continue
		}
		bad = append(bad, path)
	}
	return bad
}

func TestViolations(t *testing.T) {
	tests := []struct {
		name    string
		listing string
		want    []string
	}{
		{"stdlib and layer only", "net/http\n" + layerRoot + "/client\n" + layerRoot + "/client [" + layerRoot + "/client.test]\n" + layerRoot + "/client.test\n", nil},
		{"third party is fine", "github.com/getkin/kin-openapi/openapi3\n", nil},
		{"one module-internal path", "net/http\n" + modulePrefix + "internal/atomicfile\n", []string{modulePrefix + "internal/atomicfile"}},
		{"sibling with the layer's name as a prefix", modulePrefix + "internal/opensubsonicx\n", []string{modulePrefix + "internal/opensubsonicx"}},
		{"violation in a test variant line", modulePrefix + "internal/config [" + layerRoot + "/client.test]\n", []string{modulePrefix + "internal/config"}},
		{"several", modulePrefix + "internal/a\n" + modulePrefix + "internal/b\n", []string{modulePrefix + "internal/a", modulePrefix + "internal/b"}},
		{"empty", "", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := violations(tt.listing); !slices.Equal(got, tt.want) {
				t.Errorf("violations = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLayerImportsNothingInternal(t *testing.T) {
	out, err := exec.CommandContext(t.Context(), "go", "list", "-deps", "-test", "./...").CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	if !strings.Contains(string(out), layerRoot+"/client") {
		t.Fatalf("listing never names the client package; the leg proves nothing:\n%s", out)
	}
	if bad := violations(string(out)); len(bad) > 0 {
		t.Errorf("the protocol layer depends on module-internal paths outside internal/opensubsonic:\n%s", strings.Join(bad, "\n"))
	}
}
