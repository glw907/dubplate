package store

import (
	"testing"
	"time"
)

// TestFormatStamp: a time in a zone other than UTC formats to its UTC
// digits, truncated to the microsecond, and parses back to the same
// instant.
func TestFormatStamp(t *testing.T) {
	// 05:06:07 at UTC-8 is 13:06:07 UTC.
	in := time.Date(2026, 9, 24, 5, 6, 7, 123456789, time.FixedZone("UTC-8", -8*60*60))
	const want = "20260924T130607.123456Z"

	got := FormatStamp(in)
	if got != want {
		t.Fatalf("FormatStamp = %q, want %q", got, want)
	}
	back, err := ParseStamp(got)
	if err != nil {
		t.Fatalf("ParseStamp(%q): %v", got, err)
	}
	if wantAt := time.Date(2026, 9, 24, 13, 6, 7, 123456000, time.UTC); !back.Equal(wantAt) || back.Location() != time.UTC {
		t.Errorf("ParseStamp(%q) = %v, want %v in UTC", got, back, wantAt)
	}
}

// TestMicros: a literal microsecond count is the UTC instant it names,
// in the UTC location, whatever zone the runner holds.
func TestMicros(t *testing.T) {
	// 2026-09-24T13:06:07.123456Z.
	const us = 1790255167123456
	got := Micros(us)
	want := time.Date(2026, 9, 24, 13, 6, 7, 123456000, time.UTC)
	if !got.Equal(want) {
		t.Errorf("Micros(%d) = %v, want %v", us, got, want)
	}
	if got.Location() != time.UTC {
		t.Errorf("Micros(%d) is not in the time.UTC location (got %v)", us, got.Location())
	}
}

// TestParseStamp_Refuses: every form but the fixed-width layout with
// six fractional digits and a literal Z is refused.
func TestParseStamp_Refuses(t *testing.T) {
	tests := []struct {
		name string
		in   string
	}{
		{name: "empty", in: ""},
		{name: "no fraction", in: "20260924T130607Z"},
		{name: "a short fraction", in: "20260924T130607.12345Z"},
		{name: "a long fraction", in: "20260924T130607.1234567Z"},
		{name: "a comma before the fraction", in: "20260924T130607,123456Z"},
		{name: "a numeric zone", in: "20260924T130607.123456+0000"},
		{name: "a lowercase zone", in: "20260924T130607.123456z"},
		{name: "no zone", in: "20260924T130607.123456"},
		{name: "a trailing character", in: "20260924T130607.123456Z.db"},
		{name: "a day out of range", in: "20260931T130607.123456Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ParseStamp(tt.in); err == nil {
				t.Errorf("ParseStamp(%q) = %v, want an error", tt.in, got)
			}
		})
	}
}
