// Package backup is a hand-authored violation of the first
// architectural invariant, used only by
// TestHarnessImportCheckFiresOnItsOwnFixture: a package outside
// internal/harness that imports the harness.
// This file is never built (testdata/ is ignored by the go tool).
package backup

import "github.com/glw907/dubplate/internal/harness/oracle"

var _ = oracle.Reconcile
