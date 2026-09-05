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

// parseDate splits a MusicBrainz date on "-" and returns the parts it
// can parse, mirroring _get_date (beetsplug/musicbrainz.py:180-200): an
// unparsable part becomes zero without failing the whole parse, and
// extra parts beyond the third are discarded silently. A "stop at the
// first failure" implementation gets the "1999-XX-06" row wrong: the
// unparsable month must not swallow the real day that follows it.
func parseDate(s string) (year, month, day int) {
	parts := strings.Split(s, "-")
	if len(parts) > 0 {
		year = parseDatePart(parts[0])
	}
	if len(parts) > 1 {
		month = parseDatePart(parts[1])
	}
	if len(parts) > 2 {
		day = parseDatePart(parts[2])
	}
	return year, month, day
}

// parseDatePart returns s parsed as an int, or zero when s is not a
// valid integer, mirroring _get_date's own per-part int() call that
// catches ValueError and substitutes None.
func parseDatePart(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}
