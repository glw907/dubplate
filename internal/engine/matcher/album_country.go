package matcher

import "time"

// addAlbumCountry is distance()'s country component
// (autotag/distance.py:495-502): with a nonempty Preferred.Countries
// list and a candidate country, ranks the candidate by pattern
// position (AddPriority); otherwise, when both the item set's likely
// country and the candidate's own are set, grades a plain string
// distance between the two codes.
//
// Each preferred pattern is compiled case-insensitive with no
// `(\d+x)?` wrapper (unlike album_media.go): beets compiles country
// patterns as a bare `re.compile(pat, re.I)` (distance.py:497).
// Distance._eq matches a compiled option with re.Pattern.match, which
// anchors at position 0 (distance.py:265-267); CompilePatternOption
// reproduces that anchor explicitly (contract 7), so this is the
// second of the two places (with album_media.go) that anchored
// case-insensitive pattern matching is proved.
func addAlbumCountry(dist *Distance, tun Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	options := compileCountryPatterns(tun.Preferred.Countries)
	if info.Country != "" && len(options) > 0 {
		dist.AddPriority("country", info.Country, options...)
		return
	}
	if likelies.Country != "" && info.Country != "" {
		dist.AddString("country", &likelies.Country, &info.Country)
	}
}

// compileCountryPatterns panics on an uncompilable pattern because this
// call site has no error path of its own; Tunables.Validate is the
// check that makes that panic unreachable in a configured system, by
// calling this same function at config-load time and turning its panic
// into a reportable error.
func compileCountryPatterns(patterns []string) []any {
	options := make([]any, len(patterns))
	for i, pattern := range patterns {
		option, err := CompilePatternOption(pattern, true)
		if err != nil {
			panic("matcher: preferred.countries pattern " + pattern + " does not compile: " + err.Error())
		}
		options[i] = option
	}
	return options
}
