// Package matcher is the Go port of beets' scoring core: the tunable
// surface, the vocabulary a candidate release and a user's item are
// expressed in, the [Distance] accumulator and the penalty components
// that compose it, and the candidate set and recommendation built on
// them. It never imports
// internal/harness and never reads a clock; every citation below is
// checked against pinned beets 2.14.1 (tools/venv/.venv).
package matcher

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
)

// ErrInvalidTunable reports a Tunables value that names something the
// matcher cannot resolve: a recommendation level, an action, a
// required field, or a preferred-media or preferred-country pattern.
//
// It exists because beets gets this validation from confuse and from
// re.compile, both of which raise mid-import: an unknown max_rec level
// raises inside _recommendation (match.py:313-320) and an unknown
// required tag raises AttributeError inside _add_candidate
// (match.py:357). A refusal at config load beats a crash mid-import,
// and that difference is a deliberate divergence recorded here.
var ErrInvalidTunable = errors.New("invalid tunable")

// Tunables is the verified-complete surface of beets' shipped
// match-time configuration: every field the distance and
// recommendation path reads, each carrying pinned beets' shipped
// default and the config_default.yaml or plugin-source line that
// proves it. Every default equals pinned beets' shipped value.
//
// Two config keys beets ships under match: are deliberately absent:
// original_date and per_disc_numbering are read only in
// [beets/autotag/hooks.py]'s AlbumInfo.item_data and TrackInfo.raw_data
// respectively, never on the distance path (autotag/distance.py,
// autotag/match.py), so they move applied_fields after a verdict and
// never the verdict itself.
//
// Two more are deliberately absent for a different reason:
// album_disambig_fields and singleton_disambig_fields
// (config_default.yaml:206-207) are display-only, read through
// Match.disambig_fields (match.py:113-139) and never on the distance
// path either.
//
// Four fields are operator-facing: StrongRecThresh, MediumRecThresh,
// RecGapThresh, and SearchLimit. None is registered as a config key
// yet, and this package imports nothing from internal/config.
//
// Three surfaces below have no corpus fixture: Required, Ignored, and
// Preferred.OriginalYear. Their coverage is the unit tests of the code
// that reads each: the year component for Preferred.OriginalYear, and
// [Candidates.Add] for Required and Ignored.
type Tunables struct {
	// StrongRecThresh is the distance below which a match earns a
	// "strong" recommendation. The comparison is strict "<"
	// (match.py:285), the asymmetry that makes the proximity flag
	// possible. Default 0.04 (config_default.yaml:168). Operator-facing.
	StrongRecThresh float64

	// MediumRecThresh is the distance at or below which a match earns
	// at least a "medium" recommendation. The comparison is "<="
	// (match.py:288), unlike StrongRecThresh's strict form. Default 0.25
	// (config_default.yaml:169). Operator-facing.
	MediumRecThresh float64

	// RecGapThresh is the minimum distance gap between the best and
	// second-best candidate below which the recommendation degrades,
	// the other half of the proximity flag (match.py:294-299). Default 0.25
	// (config_default.yaml:170). Operator-facing.
	RecGapThresh float64

	// MaxRec caps the recommendation a penalty key may reach even at
	// zero distance: a penalty-key name to the recommendation level it
	// ceilings at, so the value is a level name and not a number
	// (matching cassette.Thresholds.MaxRec, cassette.go:88-93). Default
	// {"missing_tracks": "medium", "unmatched_tracks": "medium"}
	// (config_default.yaml:171-173).
	MaxRec map[string]string

	// DistanceWeights is match.distance_weights: twenty keys
	// (config_default.yaml:175-194; :174 is the distance_weights:
	// header), the count every committed corpus case's echoed weight
	// set agrees on.
	DistanceWeights map[string]float64

	// Preferred holds match.preferred's three sub-keys
	// (config_default.yaml:195-198).
	Preferred Preferred

	// Ignored is match.ignored: source identifiers a candidate is
	// dropped for regardless of distance. Default [] (empty)
	// (config_default.yaml:199). No committed corpus case exercises a
	// nonempty value.
	Ignored []string

	// Required is match.required: metadata fields a candidate must
	// carry to be considered. Default [] (empty)
	// (config_default.yaml:200). No committed corpus case exercises a
	// nonempty value.
	Required []string

	// IgnoredMedia is match.ignored_media: media type strings excluded
	// before scoring. Default [] (empty) (config_default.yaml:201).
	IgnoredMedia []string

	// IgnoreDataTracks is match.ignore_data_tracks. Default true
	// (config_default.yaml:202).
	IgnoreDataTracks bool

	// IgnoreVideoTracks is match.ignore_video_tracks. Default true
	// (config_default.yaml:203).
	IgnoreVideoTracks bool

	// TrackLengthGrace is match.track_length_grace, in seconds:
	// subtracted from the absolute track-length difference before it
	// is ratioed against TrackLengthMax. 2.14.1 reads it through the
	// cached get_track_length_grace() (autotag/distance.py:376-379),
	// called inline in track_distance's length component
	// (autotag/distance.py:410-411), replacing the direct config read
	// 2.13.1 had at the same site. Default 10 (config_default.yaml:204).
	TrackLengthGrace float64

	// TrackLengthMax is match.track_length_max, in seconds: the
	// denominator AddRatio divides the graced length difference by.
	// 2.14.1 reads it through the cached get_track_length_max()
	// (autotag/distance.py:382-385), called inline alongside
	// get_track_length_grace() (autotag/distance.py:411). Default 30
	// (config_default.yaml:205).
	TrackLengthMax float64

	// SearchLimit is the metadata-source plugin's own search_limit
	// subkey, not a top-level match: key
	// (metadata_plugins.py:203,421). The legacy spelling beets still
	// reads and deprecates is musicbrainz.match.searchlimit
	// (beetsplug/musicbrainz.py:292-294), read relative to the plugin's
	// own config namespace through deprecate_for_user
	// (beets/util/deprecation.py:37). Beets' own deprecation message
	// spells it musicbrainz.searchlimit (beetsplug/musicbrainz.py:297),
	// which is where every wrong copy of this fact came from. Default 5.
	// Operator-facing.
	SearchLimit int

	// DataSourceMismatchPenalty is the loaded MusicBrainz plugin's own
	// mismatch penalty, resolved source_weight-first through the
	// property data_source_mismatch_penalty (metadata_plugins.py:192-
	// 197, whose own deprecated legacy spelling is
	// musicbrainz.source_weight) and wired into the plugin's own config
	// at init time (:203-204). It is the value that populates
	// DataSourcePenalties["MusicBrainz"] in [DefaultTunables], not the
	// class-level fallback get_penalty falls back to when
	// DataSourcePenalties names no matching source at all: that fallback
	// is the unexported defaultDataSourceMismatchPenalty constant
	// (penalty.go), MetadataSourcePlugin.DEFAULT_DATA_SOURCE_MISMATCH_
	// PENALTY (metadata_plugins.py:182). The two are the same value at
	// launch (0.5) and diverge the moment source_weight is configured.
	// Default 0.5.
	DataSourceMismatchPenalty float64

	// DataSourcePenalties is the loaded metadata source set
	// AddDataSource's get_penalty(after) scans before falling back to
	// the class-level defaultDataSourceMismatchPenalty constant
	// (penalty.go): each loaded plugin's own data_source name
	// (metadata_plugins.py:184-190, cls.__name__ with "Plugin"
	// stripped) mapped to that plugin's own mismatch penalty
	// (metadata_plugins.py:162-171). At launch beets loads one plugin,
	// plugins: [musicbrainz] (config_default.yaml:11), whose class is
	// MusicBrainzPlugin (beetsplug/musicbrainz.py:243) and so whose
	// data_source name is "MusicBrainz", carrying the plugin base
	// class's own shipped default (metadata_plugins.py:182). Default
	// {"MusicBrainz": 0.5}.
	DataSourcePenalties map[string]float64

	// ExtraTags is the musicbrainz plugin's own extra_tags subkey, not
	// a config_default.yaml entry (beetsplug/musicbrainz.py:286).
	// Default [] (empty).
	ExtraTags []string

	// Timid is import.timid: whether every match, not only a low-
	// confidence one, stops for confirmation. Default false
	// (config_default.yaml:28).
	Timid bool

	// Languages is import.languages: locale preference order for a
	// candidate's preferred-alias selection. Default [] (empty)
	// (config_default.yaml:33), which makes _preferred_alias
	// structurally unreachable at launch. A nonempty default would be an
	// improvement riding the port, and every default here is beets'
	// shipped value.
	Languages []string

	// IgnoredAliasTypes is import.ignored_alias_types: alias type
	// names excluded from preferred-alias selection. Default [] (empty)
	// (config_default.yaml:59).
	IgnoredAliasTypes []string

	// Quiet is import.quiet: whether a "none" recommendation runs
	// NoneRecAction without asking. Default false
	// (config_default.yaml:29).
	Quiet bool

	// QuietFallback is import.quiet_fallback: the action a quiet,
	// timid-incompatible import takes on an ambiguous match. Default
	// "skip" (config_default.yaml:34).
	QuietFallback string

	// NoneRecAction is import.none_rec_action: the action taken on a
	// "none" recommendation when not Quiet. Default "ask"
	// (config_default.yaml:35).
	NoneRecAction string

	// VAName is the va_name config key: the artist string every
	// artist field is overwritten with on a various-artists release
	// (beetsplug/musicbrainz.py:673-681). Default "Various Artists"
	// (config_default.yaml:107). It moves no distance, because the
	// album artist component is skipped entirely when AlbumInfo.VA
	// (addAlbumArtist). It does move AlbumInfo.Artist, which the
	// adjudication ledger grades field by field, so it is a tunable
	// here rather than a literal in the translation layer. No
	// committed case echoes it, so the conform defaults cross-check
	// cannot cover it; the translation layer's VA test does.
	//
	// musicbrainz.aliases_as_credits is deliberately NOT here: it
	// selects only which object supplies artist_credit
	// (musicbrainz.py:331-333), and artist_credit is neither in
	// AlbumInfo nor in the ledger's graded set, so the translation
	// never reads it.
	VAName string

	// StringDistPatterns is the SD_PATTERNS table string_dist's
	// pattern loop runs, defaulting to beets' own six. It exists as
	// a field, not a package-level constant a component reads
	// directly, so the compat switch for the adjudicated
	// bug-remix-parenthetical defect swaps in
	// [SDPatternsNoBracketDiscount] by assignment rather than editing
	// string_dist itself.
	StringDistPatterns []SDPattern
}

// Preferred is match.preferred (config_default.yaml:195-198).
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
	// this key.
	OriginalYear bool
}

// DefaultTunables returns beets' shipped defaults, pinned at 2.14.1
// and cited field by field on [Tunables].
func DefaultTunables() Tunables {
	tun := Tunables{
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
		ExtraTags:                 nil,
		Timid:                     false,
		Languages:                 nil,
		IgnoredAliasTypes:         nil,
		Quiet:                     false,
		QuietFallback:             "skip",
		NoneRecAction:             "ask",
		VAName:                    "Various Artists",
		StringDistPatterns:        beetsSDPatterns,
	}
	// The MusicBrainz entry is the plugin's resolved penalty
	// (metadata_plugins.py:192-197), read from the field above rather
	// than written as a second literal, so the two cannot disagree.
	tun.DataSourcePenalties = map[string]float64{"MusicBrainz": tun.DataSourceMismatchPenalty}
	return tun
}

// QuietImportTunables returns DefaultTunables with the four assignments
// an unattended quiet import makes, the settings
// corpus/profiles/estate.yaml records. No import pipeline calls it yet.
//
// It lives here rather than in the harness because internal/engine may
// import no harness package (TestEngineNeverImportsHarness in
// internal/harness/conform) and conform's gradeTunables is on the wrong
// side of that wall.
//
// Timid and QuietFallback already ship at these values and are assigned
// anyway, so the set reads as one configuration rather than as a diff a
// reader has to compute.
func QuietImportTunables() Tunables {
	tun := DefaultTunables()
	tun.Quiet = true           // beet import's -q flag
	tun.QuietFallback = "skip" // corpus/profiles/estate.yaml import.quiet_fallback
	tun.NoneRecAction = "skip" // corpus/profiles/estate.yaml import.none_rec_action
	tun.Timid = false          // corpus/profiles/estate.yaml import.timid
	return tun
}

// SDPattern is one entry in string_dist's pattern-discount table:
// Pattern is deleted from both inputs wherever it matches, unanchored,
// over the whole string, because SD_PATTERNS are re.sub arguments and
// re.sub never anchors; the only "^" in beets' table is its own in
// "^the ". The anchoring [CompilePatternOption] applies belongs to
// Distance._eq's options and to nothing here. Weight is the penalty
// added to the base distance for each side the deletion changed.
//
// Replacement is a [regexp.Regexp.Expand] template substituted for the
// match instead of Python's empty string. Every row carries the zero
// value except the featuring row, whose rewrite must keep the
// character its Unicode-aware boundary consumed; see beetsSDPatterns'
// doc comment.
type SDPattern struct {
	Pattern     *regexp.Regexp
	Weight      float64
	Replacement string
}

// sdEndWords is SD_END_WORDS (autotag/distance.py:34): words moved
// from a trailing ", <word>" to the front of the string before the
// pattern loop runs, so "something, the" scores equal to "the
// something".
var sdEndWords = []string{"the", "a", "an"}

// The rows beetsSDPatterns and SDPatternsNoBracketDiscount share are
// compiled once and held here, so the two tables cannot drift apart on
// them.
var (
	sdArticle   = SDPattern{Pattern: regexp.MustCompile(`^the `), Weight: 0.1}
	sdEPSingle  = SDPattern{Pattern: regexp.MustCompile(`[\[\(]?(ep|single)[\]\)]?`), Weight: 0.0}
	sdFeaturing = SDPattern{Pattern: regexp.MustCompile(sdFeaturingPattern), Weight: 0.1, Replacement: `${1}`}
	sdPart      = SDPattern{Pattern: regexp.MustCompile(`(, )?(pt\.|part) .+`), Weight: 0.2}
)

// sdFeaturingPattern is the featuring row's Go form.
//
// Beets' own featuring pattern is
// [\[\(]?\b(featuring|feat|ft)\b[\. :].+ (autotag/distance.py:39).
// Python re's \b is Unicode-aware and Go RE2's \b is ASCII-only, so a
// straight port sees a word boundary inside "Geschäft" that Python does
// not. sdFeaturingPattern translates the two \b checks into two
// branches instead: a leading bracket needs no boundary check of its
// own, because the trailing \b always holds (the separator class
// [\. :] that must follow it carries no word character) and the
// leading one is automatic after a bracket (a bracket is never a word
// character, so the boundary there is already satisfied); every other
// case reads the leading boundary as the start of the string or a rune
// outside letters, digits, and underscore. That second branch keeps
// the boundary rune in its own group, and the row's `${1}` Replacement
// reinserts it, because beets' Unicode-aware \b consumes no character
// while the translated branch's [^\pL\pN_] does; without the
// replacement the rewrite would delete the non-word rune before the
// trigger, which matters because unidecode maps some of those runes to
// ASCII letters ("©" to "(c)", "™" to "(tm)").
const sdFeaturingPattern = `[\[\(](?:featuring|feat|ft)[\. :].+|(^|[^\pL\pN_])(?:featuring|feat|ft)[\. :].+`

// beetsSDPatterns is SD_PATTERNS (autotag/distance.py:36-42), beets'
// six shipped discounts in their fixed order. It is the default of
// [Tunables.StringDistPatterns], held to beets' literals one for one
// (TestBeetsSDPatternsMatchBeetsLiterals), with one named departure:
// the featuring row carries sdFeaturingPattern and a `${1}`
// Replacement, two ASCII-boundary branches sharing a
// kept-character group, in place of beets' own literal and its empty
// substitution, because Go RE2's \b is ASCII-only where Python re's is
// Unicode-aware; see sdFeaturingPattern's doc comment above for the
// full translation. The featuring pattern carries the word boundaries
// upstream shipped in beets 2.14.0 (19d091440), so "ft" inside an
// ordinary word such as "Draft" no longer reads as a featuring credit.
var beetsSDPatterns = []SDPattern{
	sdArticle,
	sdEPSingle,
	sdFeaturing,
	{Pattern: regexp.MustCompile(`\(.*?\)`), Weight: 0.3},
	{Pattern: regexp.MustCompile(`\[.*?\]`), Weight: 0.3},
	sdPart,
}

// SDPatternsNoBracketDiscount carries beetsSDPatterns' six discounts
// with the two 0.3-weight bracket patterns removed. The corpus README
// calls that "arithmetically the same as charging a parenthetical at
// full weight", which holds only while no later pattern fires: a
// full-weight discount still shifts str1, str2 and base_dist for the
// rest of the loop where a removed pattern does not
// (beets/autotag/distance.py:98-120). The two agree on all four
// adv-remix-* pairs, so the committed figures stand; the table is
// defined by what it removes, not by that equivalence. It is the table
// the four adv-remix-* cases' defect_free figures are derived under,
// and the corrected behavior for the adjudicated
// bug-remix-parenthetical defect (#5868). Nothing on the distance path
// selects it; [Compat.Apply] assigns it to
// Tunables.StringDistPatterns.
var SDPatternsNoBracketDiscount = []SDPattern{
	sdArticle,
	sdEPSingle,
	sdFeaturing,
	sdPart,
}

// Validate resolves every name-typed and pattern-typed field once, so a
// malformed configuration fails where an operator can see it rather
// than inside a scoring run. It reports every problem rather than the
// first.
//
// It is a config-load check and never a precondition the rest of this
// package assumes: Recommend and ActionFor re-resolve their own names
// and return their own errors, because a Tunables value can reach them
// without passing through here.
func (t Tunables) Validate() error {
	var errs []error

	for _, name := range t.Required {
		if _, ok := requiredFields[name]; !ok {
			errs = append(errs, fmt.Errorf("required: %q: %w", name, ErrInvalidTunable))
		}
	}

	if _, err := compileMediaPatterns(t.Preferred.Media); err != nil {
		errs = append(errs, fmt.Errorf("preferred.media: %w: %w", ErrInvalidTunable, err))
	}
	if _, err := compileCountryPatterns(t.Preferred.Countries); err != nil {
		errs = append(errs, fmt.Errorf("preferred.countries: %w: %w", ErrInvalidTunable, err))
	}

	// MaxRec's levels are walked in sorted key order so a joined error
	// message is reproducible, even though map iteration order would
	// otherwise decide it (the determinism constraint applies to
	// Validate's output too).
	for _, key := range slices.Sorted(maps.Keys(t.MaxRec)) {
		if _, err := parseRecommendation(t.MaxRec[key]); err != nil {
			errs = append(errs, fmt.Errorf("max_rec[%q]: %w", key, err))
		}
	}

	if _, err := resolveAction("import.quiet_fallback", t.QuietFallback, quietFallbacks); err != nil {
		errs = append(errs, err)
	}
	if _, err := resolveAction("import.none_rec_action", t.NoneRecAction, noneRecActions); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}
