package timereads

import "time"

// clock holds time.Now as a value, never calling it here; the check
// must refuse the reference all the same.
var clock = time.Now
