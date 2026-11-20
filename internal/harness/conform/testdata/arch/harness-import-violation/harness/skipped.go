// Package harness stands for internal/harness itself, which may import
// its own subpackages; TestHarnessImportCheckFiresOnItsOwnFixture
// requires the walk to skip it.
// This file is never built (testdata/ is ignored by the go tool).
package harness

import "github.com/glw907/dubplate/internal/harness/cassette"

var _ = cassette.Case{}
