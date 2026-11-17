package musicbrainz

import (
	"testing"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// TestPreferredReleaseEvent_MeasuredTable reproduces T10's measured
// table exactly: configured-country membership is exact and
// case-sensitive, and configured order wins over response order.
// Mutation 1 (compiling Countries through matcher.CompilePatternOption)
// fails the "gb" row, because a case-insensitive regex would match
// "GB". Mutation 3 (the two loops transposed) fails the ["US","GB"]
// row, because the inner-outermost order would return the first
// event's country ("GB") rather than the first configured one ("US").
func TestPreferredReleaseEvent_MeasuredTable(t *testing.T) {
	rel := loadTestdata[Release](t, "preferred_country_events.json")

	tests := []struct {
		name        string
		countries   []string
		wantCountry string
		wantDate    string
	}{
		{"GB matches its own event", []string{"GB"}, "GB", "1999-01-02"},
		{"lowercase gb falls back", []string{"gb"}, "XW", "1980-05-06"},
		{"configured order wins", []string{"US", "GB"}, "US", "2001-03-04"},
	}
	for _, tt := range tests {
		tun := matcher.DefaultTunables()
		tun.Preferred.Countries = tt.countries
		gotCountry, gotDate := preferredReleaseEvent(&rel, tun)
		if gotCountry != tt.wantCountry || gotDate != tt.wantDate {
			t.Errorf("%s: preferredReleaseEvent(%v) = (%q, %q), want (%q, %q)",
				tt.name, tt.countries, gotCountry, gotDate, tt.wantCountry, tt.wantDate)
		}
	}
}

// TestPreferredReleaseEvent_NullAreaIsSkipped proves a release event
// whose area is null is skipped rather than panicking or matching, and
// that a later event in the same list is still reachable.
func TestPreferredReleaseEvent_NullAreaIsSkipped(t *testing.T) {
	rel := loadTestdata[Release](t, "preferred_country_null_area.json")
	tun := matcher.DefaultTunables()
	tun.Preferred.Countries = []string{"GB"}

	gotCountry, gotDate := preferredReleaseEvent(&rel, tun)
	if gotCountry != "GB" || gotDate != "1999-01-02" {
		t.Errorf("preferredReleaseEvent = (%q, %q), want (\"GB\", \"1999-01-02\")", gotCountry, gotDate)
	}
}

// TestPreferredReleaseEvent_ReturnsConfiguredStringNotAreaCode proves
// the returned country is the CONFIGURED string, not the area's own
// code, when the two differ in spelling. Mutation 2 (returning
// event.Area.ISO31661Codes[0] instead of the configured value) fails
// this test, because the fixture's matching area lists "GB" before the
// configured "XW".
func TestPreferredReleaseEvent_ReturnsConfiguredStringNotAreaCode(t *testing.T) {
	rel := loadTestdata[Release](t, "preferred_country_multi_area.json")
	tun := matcher.DefaultTunables()
	tun.Preferred.Countries = []string{"XW"}

	gotCountry, gotDate := preferredReleaseEvent(&rel, tun)
	if gotCountry != "XW" || gotDate != "1999-01-02" {
		t.Errorf("preferredReleaseEvent = (%q, %q), want (\"XW\", \"1999-01-02\")", gotCountry, gotDate)
	}
}

// TestPreferredReleaseEvent_NoEventsNoCountryYieldsEmpty proves a
// release with neither release-events nor a top-level country and date
// yields ("", ""), the R1 zero-value mapping of beets' (None, None).
func TestPreferredReleaseEvent_NoEventsNoCountryYieldsEmpty(t *testing.T) {
	rel := loadTestdata[Release](t, "no_release_events_no_country.json")
	gotCountry, gotDate := preferredReleaseEvent(&rel, matcher.DefaultTunables())
	if gotCountry != "" || gotDate != "" {
		t.Errorf("preferredReleaseEvent = (%q, %q), want (\"\", \"\")", gotCountry, gotDate)
	}
}

// TestParseYear_MeasuredTable reproduces T11's measured six-row table
// for the year alone: the month and day parts never change it.
func TestParseYear_MeasuredTable(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"1999", 1999},
		{"1999-05", 1999},
		{"1999-XX-06", 1999},
		{"199x", 0},
		{"1999-05-06-07", 1999},
	}
	for _, tt := range tests {
		if got := parseYear(tt.in); got != tt.want {
			t.Errorf("parseYear(%q) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
