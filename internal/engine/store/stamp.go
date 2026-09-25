package store

import (
	"fmt"
	"time"
)

// stampLayout is the one layout every backup stamp takes: UTC to the
// microsecond at a fixed width, so stamps sort lexicographically in
// time order.
const stampLayout = "20060102T150405.000000Z"

// FormatStamp returns t as a backup stamp. It converts t to UTC first,
// since the layout's "Z" is a literal: formatting a local time would
// write local digits beside it.
func FormatStamp(t time.Time) string {
	return t.UTC().Format(stampLayout)
}

// ParseStamp returns the time s names when s is exactly what
// [FormatStamp] writes: fixed width, six fractional digits, and the
// zone "Z". Every other form is refused.
func ParseStamp(s string) (time.Time, error) {
	t, err := time.Parse(stampLayout, s)
	// time.Parse also accepts a comma before the fraction, so only the
	// round trip pins the one spelling.
	if err != nil || FormatStamp(t) != s {
		return time.Time{}, fmt.Errorf("stamp %q: want the layout %s", s, stampLayout)
	}
	return t, nil
}
