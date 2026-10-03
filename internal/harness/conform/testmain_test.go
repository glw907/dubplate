package conform

import (
	"fmt"
	"os"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
	"github.com/glw907/dubplate/internal/engine/matcher/unidecode/unidecodetest"
)

// TestMain installs the pinned Unidecode transliterator before any
// test in this package runs, the same load a scoring verb performs
// before it starts. Run with -count=1: the workstation's cached wheel
// file sits outside the module, GOPATH, and GOROOT roots, so Go's test
// cache does not track it, and a cached "ok" would survive a missing
// or replaced wheel.
func TestMain(m *testing.M) {
	table := unidecodetest.LoadCachedTable(mainTB{})
	matcher.SetTransliterator(table.Lookup)
	os.Exit(m.Run())
}

// mainTB lets TestMain, which has no *testing.T, call
// unidecodetest.LoadCachedTable: a failure prints and exits.
type mainTB struct{ testing.TB }

func (mainTB) Helper() {}

func (mainTB) Fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
