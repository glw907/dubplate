package harness

import "time"

// The harness may read the clock; the wide form skips it.
var harnessClock = time.Now()
