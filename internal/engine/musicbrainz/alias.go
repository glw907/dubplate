package musicbrainz

import (
	"strings"

	"github.com/glw907/dubplate/internal/engine/matcher"
)

// preferredAlias returns the first alias matching the configured
// locales, mirroring _preferred_alias (beetsplug/musicbrainz.py:128-155):
// an outer loop over tun.Languages in configured order and an inner loop
// over aliases in response order, accepting the first alias whose locale
// matches exactly (case-sensitively), whose primary flag is truthy, and
// whose lowercased type is not in tun.IgnoredAliasTypes. It returns nil
// when nothing matches, which is every case under beets' shipped empty
// languages default (Ruling D).
//
// An ignored type falls through to the NEXT matching alias rather than
// to nil; the ignored set is lowercased while the locale comparison is
// not.
//
// beets' explicit languages parameter is not ported, because no call
// site in the plugin ever passes it: both callers reach the function
// through _key_with_preferred_alias or _parse_artist_credits, and both
// let it read config. A parameter nothing supplies would be an untested
// branch, so it is named here rather than added later.
//
// A nil alias list is legal and selects nothing, which is a deliberate
// divergence from beets' own recording["aliases"] at
// musicbrainz.py:589: an absent aliases key raises KeyError there and
// decodes to a nil slice here. RELEASE_INCLUDES always requests aliases,
// so the raising branch is unreachable through a well-formed capture,
// and _key_with_preferred_alias itself reads the same key through
// .get(..., []) at :161.
func preferredAlias(aliases []Alias, tun matcher.Tunables) *Alias {
	ignored := make(map[string]struct{}, len(tun.IgnoredAliasTypes))
	for _, aliasType := range tun.IgnoredAliasTypes {
		ignored[strings.ToLower(aliasType)] = struct{}{}
	}

	for _, locale := range tun.Languages {
		for i, alias := range aliases {
			if alias.Locale == nil || *alias.Locale != locale {
				continue
			}
			if alias.Primary == nil || !*alias.Primary {
				continue
			}
			var aliasType string
			if alias.Type != nil {
				aliasType = *alias.Type
			}
			if _, skip := ignored[strings.ToLower(aliasType)]; skip {
				continue
			}
			return &aliases[i]
		}
	}
	return nil
}

// keyWithPreferredAlias returns the alias name when preferredAlias
// finds one and fallback otherwise, mirroring
// _key_with_preferred_alias (musicbrainz.py:158-162). beets calls it at
// three sites: the release title (:648), the release-group title
// (:487-489), and the recording title (:443). This port has two
// callers, because release_group_title is not in matcher.AlbumInfo and
// is rung 8's applied-field territory.
func keyWithPreferredAlias(aliases []Alias, fallback string, tun matcher.Tunables) string {
	if alias := preferredAlias(aliases, tun); alias != nil {
		return alias.Name
	}
	return fallback
}
