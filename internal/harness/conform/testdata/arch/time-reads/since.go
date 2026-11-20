package timereads

import "time"

// Elapsed reads the clock through time.Since, which the check must refuse.
func Elapsed(t time.Time) time.Duration { return time.Since(t) }
