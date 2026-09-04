// Package fixture is a hand-authored violation of the first
// architectural invariant, used only by
// TestHarnessImportCheckFiresOnItsOwnFixture: it imports
// internal/harness/cassette, which internal/engine must never do.
// This file is never built (testdata/ is ignored by the go tool).
package fixture

import "github.com/glw907/dubplate/internal/harness/cassette"

var _ = cassette.Case{}
