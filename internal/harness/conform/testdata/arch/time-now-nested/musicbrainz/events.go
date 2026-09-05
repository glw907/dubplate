// Package fixture is a hand-authored violation of the "no clock read"
// rule, nested one directory below its tree's root, used only by
// TestTimeNowCheckDescendsIntoNestedPackages: it proves checkNoTimeNow
// walks into a subdirectory rather than only scanning the root it is
// given. This file is never built (testdata/ is ignored by the go
// tool).
package fixture

import "time"

func now() time.Time {
	return time.Now()
}
