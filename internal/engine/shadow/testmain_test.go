package shadow

import (
	"fmt"
	"os"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
)

// TestMain installs the pinned Unidecode transliterator before any
// test in this package runs, the same load a scoring verb performs
// before it starts. Run with -count=1: the workstation's cached wheel
// file sits outside the module, GOPATH, and GOROOT roots, so Go's test
// cache does not track it, and a cached "ok" would survive a missing
// or replaced wheel.
func TestMain(m *testing.M) {
	path, err := unidecode.WheelCachePath()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	wheel, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read cached unidecode wheel at %s: %v; run make unidecode-wheel\n", path, err)
		os.Exit(1)
	}
	table, err := unidecode.Build(wheel)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	parsed, err := unidecode.Parse(table)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	matcher.SetTransliterator(parsed.Lookup)
	os.Exit(m.Run())
}
