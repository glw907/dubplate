// Package matcher is the Go port of beets' scoring core: the tunable
// surface, the vocabulary a candidate release and a user's item are
// expressed in, and (in later files) the [Distance] accumulator and the
// penalty components that compose it. It never imports
// internal/harness and never reads a clock; every citation below is
// checked against pinned beets 2.13.1 (tools/venv/.venv).
package matcher

// Tunables is the verified-complete surface of beets' shipped
// match-time configuration: every field the distance and
// recommendation path reads, each carrying pinned beets' shipped
// default and the config_default.yaml or plugin-source line that
// proves it (Ruling D: every default equals pinned beets' shipped
// value at launch).
//
// Two config keys beets ships under match: are deliberately absent:
// original_date and per_disc_numbering are read only in
// [beets/autotag/hooks.py]'s AlbumInfo.item_data and TrackInfo.raw_data
// respectively, never on the distance path (autotag/distance.py,
// autotag/match.py), so they move applied_fields after a verdict and
// never the verdict itself (corpus/README.md's config-profile transform
// table; the two "verdict-moving" cells there are corrected by this
// task's draft, not by an edit to Tunables).
//
// Two more are deliberately absent for a different reason:
// album_disambig_fields and singleton_disambig_fields
// (config_default.yaml:205-206) are display-only, read through
// Match.disambig_fields (match.py:113-142) and never on the distance
// path either.
//
// Four fields are the operator-facing subset a later rung 3 pass
// registers on internal/config's wall once rung 1b's registry lands:
// StrongRecThresh, MediumRecThresh, RecGapThresh, and SearchLimit. This
// pass names them and defers registration; it imports nothing from
// internal/config.
//
// Three surfaces below have no corpus fixture: Required, Ignored, and
// Preferred.OriginalYear. Their doc comments say so; coverage is a
// declared gap here, closed only by the table-driven unit tests of the
// component that reads each (Preferred.OriginalYear in the year
// component; Required and Ignored in 3d).
type Tunables struct {
	// StrongRecThresh is the distance below which a match earns a
	// "strong" recommendation. The comparison is strict "<", the
	// asymmetry that makes the proximity flag possible (contract 3).
	// Default 0.04 (config_default.yaml:167). Operator-facing.
	StrongRecThresh float64

	// MediumRecThresh is the distance at or below which a match earns
	// at least a "medium" recommendation. The comparison is "<=",
	// unlike StrongRecThresh's strict form (contract 3). Default 0.25
	// (config_default.yaml:168). Operator-facing.
	MediumRecThresh float64

	// RecGapThresh is the minimum distance gap between the best and
	// second-best candidate below which the recommendation degrades,
	// the other half of the proximity flag (contract 3). Default 0.25
	// (config_default.yaml:169). Operator-facing.
	RecGapThresh float64

	// MaxRec caps the recommendation a penalty key may reach even at
	// zero distance: a penalty-key name to the recommendation level it
	// ceilings at, so the value is a level name and not a number
	// (matching cassette.Thresholds.MaxRec, cassette.go:88-93). Default
	// {"missing_tracks": "medium", "unmatched_tracks": "medium"}
	// (config_default.yaml:170-172).
	MaxRec map[string]string

	// DistanceWeights is match.distance_weights: twenty keys
	// (config_default.yaml:174-193; :173 is the distance_weights:
	// header). The mechanics study's Q4 says nineteen; the venv and
	// every committed corpus case's echoed weight set agree on twenty
	// (m2), so twenty is what this field ships.
	DistanceWeights map[string]float64

	// Preferred holds match.preferred's three sub-keys
	// (config_default.yaml:194-197).
	Preferred Preferred

	// Ignored is match.ignored: source identifiers a candidate is
	// dropped for regardless of distance. Default [] (empty)
	// (config_default.yaml:198). No committed corpus case exercises a
	// nonempty value; coverage is 3d's table-driven unit tests, not a
	// corpus fixture.
	Ignored []string

	// Required is match.required: metadata fields a candidate must
	// carry to be considered. Default [] (empty)
	// (config_default.yaml:199). No committed corpus case exercises a
	// nonempty value; coverage is 3d's table-driven unit tests, not a
	// corpus fixture.
	Required []string

	// IgnoredMedia is match.ignored_media: media type strings excluded
	// before scoring. Default [] (empty) (config_default.yaml:200).
	IgnoredMedia []string

	// IgnoreDataTracks is match.ignore_data_tracks. Default true
	// (config_default.yaml:201).
	IgnoreDataTracks bool

	// IgnoreVideoTracks is match.ignore_video_tracks. Default true
	// (config_default.yaml:202).
	IgnoreVideoTracks bool

	// TrackLengthGrace is match.track_length_grace, in seconds:
	// subtracted from the absolute track-length difference before it
	// is ratioed against TrackLengthMax
	// (autotag/distance.py:365-367,398). Default 10
	// (config_default.yaml:203).
	TrackLengthGrace float64

	// TrackLengthMax is match.track_length_max, in seconds: the
	// denominator AddRatio divides the graced length difference by
	// (autotag/distance.py:371-373,399). Default 30
	// (config_default.yaml:204).
	TrackLengthMax float64

	// SearchLimit is the metadata-source plugin's own search_limit
	// subkey, not a top-level match: key
	// (metadata_plugins.py:201,415). The legacy spelling beets still
	// reads and deprecates is musicbrainz.searchlimit
	// (beetsplug/musicbrainz.py:291-298). Default 5. Operator-facing.
	SearchLimit int

	// DataSourceMismatchPenalty is a loaded metadata-source plugin's
	// own class-level fallback penalty,
	// MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_PENALTY
	// (metadata_plugins.py:180), wired into the plugin's own config at
	// init time (:201-202) and read back through the source_weight-
	// first resolution property data_source_mismatch_penalty
	// (:190-195, whose own deprecated legacy spelling is
	// musicbrainz.source_weight). It is the value get_penalty falls
	// back to when DataSourcePenalties names no matching source.
	// Default 0.5.
	DataSourceMismatchPenalty float64

	// DataSourcePenalties is the loaded metadata source set
	// AddDataSource's get_penalty(after) scans before falling back to
	// DataSourceMismatchPenalty: each loaded plugin's own data_source
	// name (metadata_plugins.py:182-188, cls.__name__ with "Plugin"
	// stripped) mapped to that plugin's own mismatch penalty
	// (metadata_plugins.py:159-169). At launch beets loads one plugin,
	// plugins: [musicbrainz] (config_default.yaml:10), whose class is
	// MusicBrainzPlugin (beetsplug/musicbrainz.py:243) and so whose
	// data_source name is "MusicBrainz", carrying the plugin base
	// class's own shipped default (metadata_plugins.py:180). Default
	// {"MusicBrainz": 0.5}.
	DataSourcePenalties map[string]float64

	// ExtraTags is the musicbrainz plugin's own extra_tags subkey, not
	// a config_default.yaml entry (beetsplug/musicbrainz.py:286).
	// Default [] (empty).
	ExtraTags []string

	// Timid is import.timid: whether every match, not only a low-
	// confidence one, stops for confirmation. Default false
	// (config_default.yaml:27).
	Timid bool

	// Languages is import.languages: locale preference order for a
	// candidate's preferred-alias selection. Default [] (empty)
	// (config_default.yaml:32), which makes _preferred_alias
	// structurally unreachable at launch. Shipping a nonempty default
	// would be an improvement riding the port, forbidden by Ruling D;
	// closing this reachability gap is a post-launch candidate.
	Languages []string

	// IgnoredAliasTypes is import.ignored_alias_types: alias type
	// names excluded from preferred-alias selection. Default [] (empty)
	// (config_default.yaml:58).
	IgnoredAliasTypes []string

	// Quiet is import.quiet: whether a "none" recommendation runs
	// NoneRecAction without asking. Default false
	// (config_default.yaml:28).
	Quiet bool

	// QuietFallback is import.quiet_fallback: the action a quiet,
	// timid-incompatible import takes on an ambiguous match. Default
	// "skip" (config_default.yaml:33).
	QuietFallback string

	// NoneRecAction is import.none_rec_action: the action taken on a
	// "none" recommendation when not Quiet. Default "ask"
	// (config_default.yaml:34).
	NoneRecAction string

	// StringDistPatterns is the SD_PATTERNS table string_dist's
	// pattern loop runs, defaulting to [BeetsSDPatterns]. It exists as
	// a field, not a package-level constant a component reads
	// directly, so 3d's compat switch for the adjudicated
	// bug-ft-pattern/bug-remix-parenthetical defects (adj-adv-ft-draft-
	// colon) swaps in [SDPatternsWordBoundary] by assignment rather
	// than editing string_dist itself.
	StringDistPatterns []SDPattern
}

// Preferred is match.preferred (config_default.yaml:194-197).
type Preferred struct {
	// Countries is preferred.countries: country codes ranked above
	// others at equal distance. Default [] (empty).
	Countries []string

	// Media is preferred.media: media type strings ranked above others
	// at equal distance. Default [] (empty).
	Media []string

	// OriginalYear is preferred.original_year: whether the year
	// component prefers a candidate's original release year over its
	// reissue year. Default false. No committed corpus case exercises
	// this key; coverage is the year component's own table-driven unit
	// test (Task 7), not a corpus fixture.
	OriginalYear bool
}

// DefaultTunables returns beets' shipped defaults, pinned at 2.13.1
// and cited field by field on [Tunables].
func DefaultTunables() Tunables {
	return Tunables{
		StrongRecThresh: 0.04,
		MediumRecThresh: 0.25,
		RecGapThresh:    0.25,
		MaxRec: map[string]string{
			"missing_tracks":   "medium",
			"unmatched_tracks": "medium",
		},
		DistanceWeights: map[string]float64{
			"data_source":      2.0,
			"artist":           3.0,
			"album":            3.0,
			"media":            1.0,
			"mediums":          1.0,
			"year":             1.0,
			"country":          0.5,
			"label":            0.5,
			"catalognum":       0.5,
			"albumdisambig":    0.5,
			"album_id":         5.0,
			"tracks":           2.0,
			"missing_tracks":   0.9,
			"unmatched_tracks": 0.6,
			"track_title":      3.0,
			"track_artist":     2.0,
			"track_index":      1.0,
			"track_length":     2.0,
			"track_id":         5.0,
			"medium":           1.0,
		},
		Preferred: Preferred{
			Countries:    nil,
			Media:        nil,
			OriginalYear: false,
		},
		Ignored:                   nil,
		Required:                  nil,
		IgnoredMedia:              nil,
		IgnoreDataTracks:          true,
		IgnoreVideoTracks:         true,
		TrackLengthGrace:          10,
		TrackLengthMax:            30,
		SearchLimit:               5,
		DataSourceMismatchPenalty: 0.5,
		DataSourcePenalties: map[string]float64{
			"MusicBrainz": 0.5,
		},
		ExtraTags:          nil,
		Timid:              false,
		Languages:          nil,
		IgnoredAliasTypes:  nil,
		Quiet:              false,
		QuietFallback:      "skip",
		NoneRecAction:      "ask",
		StringDistPatterns: BeetsSDPatterns,
	}
}

// SDPattern is one entry in string_dist's pattern-discount table: pat
// is a regular expression deleted from both inputs wherever it matches,
// unanchored, over the whole string (contract 7: SD_PATTERNS never
// anchor), and weight is the penalty added to the base distance for
// each side the deletion changed.
type SDPattern struct {
	Pattern string
	Weight  float64
}

// SDEndWords is SD_END_WORDS (autotag/distance.py:33): words moved
// from a trailing ", <word>" to the front of the string before the
// pattern loop runs, so "something, the" scores equal to "the
// something".
var SDEndWords = []string{"the", "a", "an"}

// SDReplacement is one substitution string_dist performs, in order,
// before the pattern loop.
type SDReplacement struct {
	Pattern     string
	Replacement string
}

// SDReplace is SD_REPLACE (autotag/distance.py:44): applied once to
// each input, in order, before the pattern loop.
var SDReplace = []SDReplacement{
	{Pattern: "&", Replacement: "and"},
}

// BeetsSDPatterns is SD_PATTERNS (autotag/distance.py:35-42), beets'
// six shipped discounts in their fixed order. It is the default of
// [Tunables.StringDistPatterns] and must stay byte-identical to beets'
// literals (Task 4's byte-identity test): the featuring pattern here
// carries the adjudicated bug-ft-pattern defect (matching "ft" inside
// an ordinary word like "Draft") on purpose, because 3a ships beets'
// behavior exactly and a "corrected" default here would ship a green
// tier 0 and a wrong matcher (M-a: the defect is provably invisible to
// all 72 committed rows).
var BeetsSDPatterns = []SDPattern{
	{Pattern: `^the `, Weight: 0.1},
	{Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
	{Pattern: `[\[\(]?(featuring|feat|ft)[\. :].+`, Weight: 0.1},
	{Pattern: `\(.*?\)`, Weight: 0.3},
	{Pattern: `\[.*?\]`, Weight: 0.3},
	{Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
}

// SDPatternsWordBoundary carries the same six discounts as
// [BeetsSDPatterns] with one change: the featuring pattern requires a
// word boundary before its trigger, matching upstream's symmetric fix
// (19d091440, unreleased at pinned 2.13.1) for the defect adjudication
// row adj-adv-ft-draft-colon names ("Draft:" wrongly read as a
// featuring credit). Nothing in 3a uses this table; it exists so 3b's
// defect_free derivation and 3d's compat switch can select it by
// assignment rather than editing string_dist.
var SDPatternsWordBoundary = []SDPattern{
	{Pattern: `^the `, Weight: 0.1},
	{Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
	{Pattern: `\b[\[\(]?(featuring|feat|ft)[\. :].+`, Weight: 0.1},
	{Pattern: `\(.*?\)`, Weight: 0.3},
	{Pattern: `\[.*?\]`, Weight: 0.3},
	{Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
}
