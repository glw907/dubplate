// Package fixture is a hand-authored violation of the "no clock read"
// rule, used only by TestTimeNowCheckFiresOnItsOwnFixture: it calls
// time.Now, which internal/engine/matcher must never do (the
// evaluation date is a parameter, contract 9). This file is never
// built (testdata/ is ignored by the go tool).
package fixture

import "time"

func now() time.Time {
	return time.Now()
}
