package tag

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
	table, err := unidecode.LoadCachedTable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	matcher.SetTransliterator(table.Lookup)
	os.Exit(m.Run())
}
