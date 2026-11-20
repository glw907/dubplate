package timereads

import "time"

// Remaining reads the clock through time.Until, which the check must refuse.
func Remaining(t time.Time) time.Duration { return time.Until(t) }
