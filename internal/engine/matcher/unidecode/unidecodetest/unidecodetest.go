// Package unidecodetest loads the pinned Unidecode table for tests.
package unidecodetest

import (
	"os"
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher/unidecode"
)

// LoadCachedTable builds and parses a table from the workstation's
// cached wheel (see [unidecode.WheelCachePath]) and fails tb when it
// cannot. A missing cache names `make unidecode-wheel` rather than a
// raw read error.
func LoadCachedTable(tb testing.TB) *unidecode.Table {
	tb.Helper()
	path, err := unidecode.WheelCachePath()
	if err != nil {
		tb.Fatalf("%v", err)
	}
	wheel, err := os.ReadFile(path)
	if err != nil {
		tb.Fatalf("unidecode: read cached wheel at %s: %v; run make unidecode-wheel", path, err)
	}
	data, err := unidecode.Build(wheel)
	if err != nil {
		tb.Fatalf("unidecode: build cached wheel: %v", err)
	}
	table, err := unidecode.Parse(data)
	if err != nil {
		tb.Fatalf("unidecode: parse built table: %v", err)
	}
	return table
}
