package musicbrainz

import (
	"slices"
	"strconv"
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// preferredReleaseEvent returns the country and date the translation
// writes, mirroring _preferred_release_event
// (beetsplug/musicbrainz.py:165-177): an outer loop over
// tun.Preferred.Countries in configured order and an inner loop over
// rel.ReleaseEvents, returning the first configured country any event's
// area lists, together with that event's date. Otherwise it returns the
// release's own top-level country and date.
//
// The membership test is a literal, case-sensitive slices.Contains
// against the area's ISO codes, and the country returned is the
// CONFIGURED string rather than the area's own code. The same
// Tunables.Preferred.Countries field is compiled as an anchored,
// case-insensitive regex list by the country distance component
// (matcher/album_country.go:31-40), so one config list carries two
// matching regimes in two packages. Reusing that compiler here would be
// silently wrong on any pattern-shaped entry.
//
// The preferred-country branch is a declared corpus-coverage gap: no
// committed profile sets a match.preferred.countries key, even though
// 290 of the 303 release-lookup responses' 298 top-level release events
// do carry a non-null area with iso-3166-1-codes, across 208 case
// directories. The branch would be reachable on all of those the moment
// one profile set the key; its coverage here is this file's own
// hand-authored fixtures rather than a corpus case.
func preferredReleaseEvent(rel *Release, tun matcher.Tunables) (country string, date string) {
	for _, configured := range tun.Preferred.Countries {
		for _, event := range rel.ReleaseEvents {
			if event.Area == nil {
				continue
			}
			if slices.Contains(event.Area.ISO31661Codes, configured) {
				return configured, event.Date
			}
		}
	}

	if rel.Country != nil {
		country = *rel.Country
	}
	return country, rel.Date
}

// parseYear returns the year of a MusicBrainz date, mirroring _get_date
// (beetsplug/musicbrainz.py:180-200) for the one part the matcher reads.
// It is zero when the year is absent or unparsable, however the month and
// day that follow read.
func parseYear(s string) int {
	year, _, _ := strings.Cut(s, "-")
	n, err := strconv.Atoi(year)
	if err != nil {
		return 0
	}
	return n
}
