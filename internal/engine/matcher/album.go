package matcher

import (
	"fmt"
	"time"
)

// AlbumPair is one item matched to one candidate track: the input the
// album-level tracks, missing_tracks, and unmatched_tracks components
// read, never produce. Beets' own distance() takes this pairing as
// item_info_pairs, a caller-supplied list (distance.py:429-432,
// 528-530); the assignment solve computes it, not the distance.
type AlbumPair struct {
	Item ItemFields
	Info *TrackInfo
}

// DistanceForAlbum is the Go port of beets' distance
// (autotag/distance.py:429-541): the significance of one candidate
// release's metadata, and its item-to-track pairing, against a user's
// existing items.
//
// The order of the calls below is beets' own and is load-bearing.
// Insertion order into Distance fixes the key sequence AddedKeys and
// GenericPenaltyKeys report and the summation order RawDistance and
// MaxDistance walk, and float addition is not associative, so a
// reordering both reports a different sequence and sums to a different
// number. A component that declines to add anything must decline in
// exactly the place beets declines: firing where beets does not changes
// the reported sequence even at a zero contribution.
//
// evalDate is a parameter, never a clock read: the year component's
// ratio branches divide by a span ending at evalDate's year, and a
// package that read time.Now would make every distance depend on when
// it happened to run.
//
// The only error is an uncompilable preferred pattern, wrapped exactly
// as [Tunables.Validate] wraps it at config load.
func DistanceForAlbum(items []ItemFields, info *AlbumInfo, pairs []AlbumPair, tun Tunables, evalDate time.Time) (*Distance, error) {
	media, err := compileMediaPatterns(tun.Preferred.Media)
	if err != nil {
		return nil, fmt.Errorf("preferred.media: %w: %w", ErrInvalidTunable, err)
	}
	countries, err := compileCountryPatterns(tun.Preferred.Countries)
	if err != nil {
		return nil, fmt.Errorf("preferred.countries: %w: %w", ErrInvalidTunable, err)
	}

	likelies, _ := MostCommonTags(items)
	dist := NewDistance(tun)
	addAlbumArtist(dist, likelies, info)
	addAlbumAlbum(dist, likelies, info)
	addAlbumMedia(dist, media, likelies, info)
	addAlbumMediums(dist, likelies, info)
	addAlbumYear(dist, tun.Preferred.OriginalYear, likelies, info, evalDate.Year())
	addAlbumCountry(dist, countries, likelies, info)
	addAlbumLabel(dist, likelies, info)
	addAlbumCatalogNum(dist, likelies, info)
	addAlbumDisambig(dist, likelies, info)
	addAlbumID(dist, likelies, info)
	addAlbumTracks(dist, tun, info.VA, pairs)
	addAlbumMissingTracks(dist, info, pairs)
	addAlbumUnmatchedTracks(dist, items, pairs)
	addAlbumDataSource(dist, likelies, info)
	return dist, nil
}

// addAlbumArtist is distance()'s artist component
// (autotag/distance.py:446-448): skipped entirely on a various-artists
// release, because a VA release has no single artist credit to compare
// a candidate against.
func addAlbumArtist(dist *Distance, likelies Likelies, info *AlbumInfo) {
	if info.VA {
		return
	}
	dist.AddString("artist", &likelies.Artist, &info.Artist)
}

// addAlbumAlbum is distance()'s album component
// (autotag/distance.py:450-451): always adds, because an album title is
// never optional the way media, mediums, or country are.
func addAlbumAlbum(dist *Distance, likelies Likelies, info *AlbumInfo) {
	dist.AddString("album", &likelies.Album, &info.Album)
}

// addAlbumMedia is distance()'s media component
// (autotag/distance.py:454-465): fires only when the candidate names a
// media type. With preferred patterns, compiled by compileMediaPatterns,
// it ranks the candidate's media by pattern position (AddPriority);
// otherwise, when the item set itself agrees on a media type, it grades
// a plain equality against that.
func addAlbumMedia(dist *Distance, preferred []any, likelies Likelies, info *AlbumInfo) {
	if info.Media == "" {
		return
	}

	if len(preferred) > 0 {
		dist.AddPriority("media", info.Media, preferred...)
		return
	}
	if likelies.Media != "" {
		dist.AddEquality("media", info.Media, likelies.Media)
	}
}

// compileMediaPatterns compiles each preferred pattern as
// `(\d+x)?(<pat>)`, case-insensitive, matching beets' own
// `re.compile(rf"(\d+x)?({pat})", re.I)` (distance.py:459); the
// `(\d+x)?` prefix tolerates a leading disc-count marker like "2xCD".
func compileMediaPatterns(patterns []string) ([]any, error) {
	options := make([]any, len(patterns))
	for i, pattern := range patterns {
		option, err := CompilePatternOption(`(\d+x)?(`+pattern+`)`, true)
		if err != nil {
			return nil, fmt.Errorf("matcher: preferred.media pattern %s does not compile: %w", pattern, err)
		}
		options[i] = option
	}
	return options, nil
}

// addAlbumMediums is distance()'s mediums component
// (autotag/distance.py:467-469): fires only when the item set agrees on
// a disc total and the candidate carries its own disc count. AddNumber
// makes each disc of difference a full weighted unit, not a ratio.
func addAlbumMediums(dist *Distance, likelies Likelies, info *AlbumInfo) {
	if likelies.DiscTotal == 0 || info.Mediums == 0 {
		return
	}
	dist.AddNumber("mediums", likelies.DiscTotal, info.Mediums)
}

// addAlbumYear is distance()'s year component
// (autotag/distance.py:471-493): three branches, tried in order.
//
//  1. With preferOriginal set and a candidate release year, the
//     candidate is graded against its own original year (or 1889, the
//     earliest first gramophone discs, when it carries none), ratioed
//     against the span up to evalYear.
//  2. Otherwise, with both a likely year and a candidate release year:
//     zero penalty when the likely year matches either the release or
//     the original year; a ratioed penalty when the candidate carries
//     an original year; a flat 1.0 otherwise.
//  3. Neither condition holds: decline (no penalty added at all).
//
// No committed corpus case sets preferred.original_year, so branch 1's
// coverage is the hand-authored
// testdata/album_year_preferred_original.json table.
func addAlbumYear(dist *Distance, preferOriginal bool, likelies Likelies, info *AlbumInfo, evalYear int) {
	if info.Year != 0 && preferOriginal {
		original := info.OriginalYear
		if original == 0 {
			original = 1889
		}
		diff := absInt(info.Year - original)
		diffMax := absInt(evalYear - original)
		dist.AddRatio("year", float64(diff), float64(diffMax))
		return
	}

	if likelies.Year == 0 || info.Year == 0 {
		return
	}

	if likelies.Year == info.Year || likelies.Year == info.OriginalYear {
		dist.Add("year", 0)
		return
	}
	if info.OriginalYear != 0 {
		diff := absInt(likelies.Year - info.Year)
		diffMax := absInt(evalYear - info.OriginalYear)
		dist.AddRatio("year", float64(diff), float64(diffMax))
		return
	}
	dist.Add("year", 1)
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// addAlbumCountry is distance()'s country component
// (autotag/distance.py:495-502): with preferred patterns, compiled by
// compileCountryPatterns, and a candidate country, ranks the candidate
// by pattern position (AddPriority); otherwise, when both the item
// set's likely country and the candidate's own are set, grades a plain
// string distance between the two codes.
func addAlbumCountry(dist *Distance, preferred []any, likelies Likelies, info *AlbumInfo) {
	if info.Country != "" && len(preferred) > 0 {
		dist.AddPriority("country", info.Country, preferred...)
		return
	}
	if likelies.Country != "" && info.Country != "" {
		dist.AddString("country", &likelies.Country, &info.Country)
	}
}

// compileCountryPatterns compiles each preferred pattern
// case-insensitive with no `(\d+x)?` wrapper: beets compiles country
// patterns as a bare `re.compile(pat, re.I)` (distance.py:497).
func compileCountryPatterns(patterns []string) ([]any, error) {
	options := make([]any, len(patterns))
	for i, pattern := range patterns {
		option, err := CompilePatternOption(pattern, true)
		if err != nil {
			return nil, fmt.Errorf("matcher: preferred.countries pattern %s does not compile: %w", pattern, err)
		}
		options[i] = option
	}
	return options, nil
}

// addAlbumLabel is distance()'s label component
// (autotag/distance.py:504-506): fires only when both the item set's
// likely label and the candidate's own label are set.
func addAlbumLabel(dist *Distance, likelies Likelies, info *AlbumInfo) {
	if likelies.Label == "" || info.Label == "" {
		return
	}
	dist.AddString("label", &likelies.Label, &info.Label)
}

// addAlbumCatalogNum is distance()'s catalognum component
// (autotag/distance.py:508-512): fires only when both the item set's
// likely catalog number and the candidate's own are set.
func addAlbumCatalogNum(dist *Distance, likelies Likelies, info *AlbumInfo) {
	if likelies.CatalogNum == "" || info.CatalogNum == "" {
		return
	}
	dist.AddString("catalognum", &likelies.CatalogNum, &info.CatalogNum)
}

// addAlbumDisambig is distance()'s albumdisambig component
// (autotag/distance.py:514-518): fires only when both the item set's
// likely disambiguation string and the candidate's own are set.
func addAlbumDisambig(dist *Distance, likelies Likelies, info *AlbumInfo) {
	if likelies.AlbumDisambig == "" || info.AlbumDisambig == "" {
		return
	}
	dist.AddString("albumdisambig", &likelies.AlbumDisambig, &info.AlbumDisambig)
}

// addAlbumID is distance()'s album_id component
// (autotag/distance.py:520-524): fires only when the item set agrees on
// a MusicBrainz album ID, and compares it against the candidate's own.
// The likely ID is the value and the candidate's ID the option,
// matching beets' own `add_equality("album_id", likelies["mb_albumid"],
// album_info.album_id)`.
func addAlbumID(dist *Distance, likelies Likelies, info *AlbumInfo) {
	if likelies.MBAlbumID == "" {
		return
	}
	dist.AddEquality("album_id", likelies.MBAlbumID, info.AlbumID)
}

// addAlbumTracks is distance()'s tracks component
// (autotag/distance.py:526-530): per matched pair, the pair's own track
// distance, scored with inclArtist set to va, is folded in two ways:
// its normalized Distance() is added as one penalty under "tracks", and
// the full sub-distance is recorded in dist.Tracks so a caller can
// inspect why a given pair scored what it did.
func addAlbumTracks(dist *Distance, tun Tunables, va bool, pairs []AlbumPair) {
	for _, pair := range pairs {
		trackDist := DistanceForTrack(tun, pair.Item, pair.Info, va)
		dist.Add("tracks", trackDist.Distance())
		dist.Tracks = append(dist.Tracks, TrackDistance{Info: pair.Info, Dist: trackDist})
	}
}

// addAlbumMissingTracks is distance()'s missing_tracks component
// (autotag/distance.py:532-534): one 1.0 penalty for every candidate
// track beyond the matched pairs. It and addAlbumUnmatchedTracks count
// only, and never contribute contents or an order.
func addAlbumMissingTracks(dist *Distance, info *AlbumInfo, pairs []AlbumPair) {
	for range max(0, len(info.Tracks)-len(pairs)) {
		dist.Add("missing_tracks", 1)
	}
}

// addAlbumUnmatchedTracks is distance()'s unmatched_tracks component
// (autotag/distance.py:536-538): one 1.0 penalty for every item beyond
// the matched pairs.
func addAlbumUnmatchedTracks(dist *Distance, items []ItemFields, pairs []AlbumPair) {
	for range max(0, len(items)-len(pairs)) {
		dist.Add("unmatched_tracks", 1)
	}
}

// addAlbumDataSource is distance()'s data_source component
// (autotag/distance.py:540). AddDataSource carries its own fire
// condition inside the accumulator rather than here.
func addAlbumDataSource(dist *Distance, likelies Likelies, info *AlbumInfo) {
	dist.AddDataSource(likelies.DataSource, info.DataSource)
}
