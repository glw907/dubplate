package ops

import "time"

// Stamp reads the clock in a package outside the engine, which only
// the wide form reaches.
func Stamp() time.Time { return time.Now() }
