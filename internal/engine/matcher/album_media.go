package matcher

import "time"

// addAlbumMedia is distance()'s media component
// (autotag/distance.py:454-465): fires only when the candidate names
// a media type. With a nonempty Preferred.Media list it ranks the
// candidate's media by pattern position (AddPriority); otherwise, when
// the item set itself agrees on a media type, it grades a plain
// equality against that.
//
// Each preferred pattern is compiled as `(\d+x)?(<pat>)` case-
// insensitive, matching beets' own `re.compile(rf"(\d+x)?({pat})",
// re.I)` (distance.py:459); the `(\d+x)?` prefix tolerates a leading
// disc-count marker like "2xCD". Distance._eq matches a compiled
// option with re.Pattern.match, which anchors at position 0
// (distance.py:265-267); CompilePatternOption reproduces that anchor
// explicitly (contract 7), so this is one of the two places (with
// album_country.go) that anchored case-insensitive pattern matching
// is proved.
func addAlbumMedia(dist *Distance, tun Tunables, likelies Likelies, info *AlbumInfo, _ []ItemFields, _ []AlbumPair, _ time.Time) {
	if info.Media == "" {
		return
	}

	options := compileMediaPatterns(tun.Preferred.Media)
	if len(options) > 0 {
		dist.AddPriority("media", info.Media, options...)
		return
	}
	if likelies.Media != "" {
		dist.AddEquality("media", info.Media, likelies.Media)
	}
}

func compileMediaPatterns(patterns []string) []any {
	options := make([]any, len(patterns))
	for i, pattern := range patterns {
		option, err := CompilePatternOption(`(\d+x)?(`+pattern+`)`, true)
		if err != nil {
			panic("matcher: preferred.media pattern " + pattern + " does not compile: " + err.Error())
		}
		options[i] = option
	}
	return options
}
