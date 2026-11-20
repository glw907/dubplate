package ops

import "time"

// A test file may read the clock; only the strict form reads it.
var testClock = time.Now()
