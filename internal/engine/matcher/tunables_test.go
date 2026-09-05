package matcher

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestDefaultTunables_MatchScalars proves DefaultTunables ships beets'
// pinned 2.13.1 shipped values for the scalar and small-structure match:
// keys (config_default.yaml:167-204), each hard-coded here rather than
// derived from DefaultTunables itself.
func TestDefaultTunables_MatchScalars(t *testing.T) {
	got := DefaultTunables()

	want := Tunables{
		StrongRecThresh:   0.04,
		MediumRecThresh:   0.25,
		RecGapThresh:      0.25,
		IgnoreDataTracks:  true,
		IgnoreVideoTracks: true,
		TrackLengthGrace:  10,
		TrackLengthMax:    30,
	}

	if got.StrongRecThresh != want.StrongRecThresh {
		t.Errorf("StrongRecThresh = %v, want %v", got.StrongRecThresh, want.StrongRecThresh)
	}
	if got.MediumRecThresh != want.MediumRecThresh {
		t.Errorf("MediumRecThresh = %v, want %v", got.MediumRecThresh, want.MediumRecThresh)
	}
	if got.RecGapThresh != want.RecGapThresh {
		t.Errorf("RecGapThresh = %v, want %v", got.RecGapThresh, want.RecGapThresh)
	}
	if got.IgnoreDataTracks != want.IgnoreDataTracks {
		t.Errorf("IgnoreDataTracks = %v, want %v", got.IgnoreDataTracks, want.IgnoreDataTracks)
	}
	if got.IgnoreVideoTracks != want.IgnoreVideoTracks {
		t.Errorf("IgnoreVideoTracks = %v, want %v", got.IgnoreVideoTracks, want.IgnoreVideoTracks)
	}
	if got.TrackLengthGrace != want.TrackLengthGrace {
		t.Errorf("TrackLengthGrace = %v, want %v", got.TrackLengthGrace, want.TrackLengthGrace)
	}
	if got.TrackLengthMax != want.TrackLengthMax {
		t.Errorf("TrackLengthMax = %v, want %v", got.TrackLengthMax, want.TrackLengthMax)
	}
}

func TestDefaultTunables_MaxRec(t *testing.T) {
	got := DefaultTunables().MaxRec
	want := map[string]string{
		"missing_tracks":   "medium",
		"unmatched_tracks": "medium",
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("MaxRec = %v, want %v", got, want)
	}
}

// TestDefaultTunables_DistanceWeights pins all twenty
// match.distance_weights keys (config_default.yaml:174-193) against
// pinned beets, resolving m2: the mechanics study's Q4 says nineteen
// keys, but the venv and every committed corpus case's echoed weight
// set carry twenty.
func TestDefaultTunables_DistanceWeights(t *testing.T) {
	got := DefaultTunables().DistanceWeights

	want := map[string]float64{
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
	}

	if len(got) != 20 {
		t.Fatalf("DistanceWeights has %d keys, want 20 (m2: the study's nineteen is superseded)", len(got))
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("DistanceWeights = %v, want %v", got, want)
	}
}

func TestDefaultTunables_Preferred(t *testing.T) {
	got := DefaultTunables().Preferred
	want := Preferred{Countries: nil, Media: nil, OriginalYear: false}
	if len(got.Countries) != 0 || len(got.Media) != 0 || got.OriginalYear != want.OriginalYear {
		t.Errorf("Preferred = %+v, want %+v", got, want)
	}
}

// TestDefaultTunables_OutsideMatch pins every tunable Q4 lists outside
// match: (SearchLimit, DataSourceMismatchPenalty, DataSourcePenalties,
// ExtraTags, Timid, Languages, IgnoredAliasTypes) and the
// verdict-to-action trio (Quiet, QuietFallback, NoneRecAction).
func TestDefaultTunables_OutsideMatch(t *testing.T) {
	got := DefaultTunables()

	if got.SearchLimit != 5 {
		t.Errorf("SearchLimit = %v, want 5", got.SearchLimit)
	}
	if got.DataSourceMismatchPenalty != 0.5 {
		t.Errorf("DataSourceMismatchPenalty = %v, want 0.5", got.DataSourceMismatchPenalty)
	}
	wantSources := map[string]float64{"MusicBrainz": 0.5}
	if !reflect.DeepEqual(got.DataSourcePenalties, wantSources) {
		t.Errorf("DataSourcePenalties = %v, want %v", got.DataSourcePenalties, wantSources)
	}
	if len(got.ExtraTags) != 0 {
		t.Errorf("ExtraTags = %v, want empty", got.ExtraTags)
	}
	if got.Timid != false {
		t.Errorf("Timid = %v, want false", got.Timid)
	}
	if len(got.Languages) != 0 {
		t.Errorf("Languages = %v, want empty", got.Languages)
	}
	if len(got.IgnoredAliasTypes) != 0 {
		t.Errorf("IgnoredAliasTypes = %v, want empty", got.IgnoredAliasTypes)
	}
	if got.Quiet != false {
		t.Errorf("Quiet = %v, want false", got.Quiet)
	}
	if got.QuietFallback != "skip" {
		t.Errorf("QuietFallback = %q, want %q", got.QuietFallback, "skip")
	}
	if got.NoneRecAction != "ask" {
		t.Errorf("NoneRecAction = %q, want %q", got.NoneRecAction, "ask")
	}
}

// TestDefaultTunables_VAName pins va_name's shipped default
// (config_default.yaml:106), the string every artist field is
// overwritten with on a various-artists release.
func TestDefaultTunables_VAName(t *testing.T) {
	got := DefaultTunables().VAName
	if got != "Various Artists" {
		t.Errorf("VAName = %q, want %q", got, "Various Artists")
	}
}

func TestDefaultTunables_ExcludedKeysStayAbsent(t *testing.T) {
	got := DefaultTunables()
	if len(got.Ignored) != 0 {
		t.Errorf("Ignored = %v, want empty", got.Ignored)
	}
	if len(got.Required) != 0 {
		t.Errorf("Required = %v, want empty", got.Required)
	}
	if len(got.IgnoredMedia) != 0 {
		t.Errorf("IgnoredMedia = %v, want empty", got.IgnoredMedia)
	}
}

// TestDefaultTunables_StringDistPatternsDefaultsToBeets proves the
// SD_PATTERNS-as-a-value seam: Tunables.StringDistPatterns defaults to
// BeetsSDPatterns, not SDPatternsWordBoundary, so 3a's default replays
// beets' behavior including the adjudicated bug-ft-pattern defect.
func TestDefaultTunables_StringDistPatternsDefaultsToBeets(t *testing.T) {
	got := DefaultTunables().StringDistPatterns
	if !reflect.DeepEqual(got, BeetsSDPatterns) {
		t.Errorf("StringDistPatterns = %v, want BeetsSDPatterns (%v)", got, BeetsSDPatterns)
	}
	if reflect.DeepEqual(got, SDPatternsWordBoundary) {
		t.Error("StringDistPatterns defaults to SDPatternsWordBoundary; it must default to BeetsSDPatterns")
	}
}

// TestBeetsSDPatterns_SixEntriesInBeetsOrder pins SD_PATTERNS
// (autotag/distance.py:35-42) byte-for-byte in beets' own order.
func TestBeetsSDPatterns_SixEntriesInBeetsOrder(t *testing.T) {
	want := []SDPattern{
		{Pattern: `^the `, Weight: 0.1},
		{Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
		{Pattern: `[\[\(]?(featuring|feat|ft)[\. :].+`, Weight: 0.1},
		{Pattern: `\(.*?\)`, Weight: 0.3},
		{Pattern: `\[.*?\]`, Weight: 0.3},
		{Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
	}
	if !reflect.DeepEqual(BeetsSDPatterns, want) {
		t.Errorf("BeetsSDPatterns = %v, want %v", BeetsSDPatterns, want)
	}
}

// TestSDPatternsWordBoundary_OnlyFeaturingPatternDiffers proves the
// guarded table (adj-adv-ft-draft-colon, upstream 19d091440) changes
// exactly one entry, at the same index, by prepending a word boundary,
// and leaves the other five discounts byte-identical to
// [BeetsSDPatterns].
func TestSDPatternsWordBoundary_OnlyFeaturingPatternDiffers(t *testing.T) {
	if len(SDPatternsWordBoundary) != len(BeetsSDPatterns) {
		t.Fatalf("SDPatternsWordBoundary has %d entries, want %d", len(SDPatternsWordBoundary), len(BeetsSDPatterns))
	}

	const featuringIndex = 2
	for i, beets := range BeetsSDPatterns {
		guarded := SDPatternsWordBoundary[i]
		if i == featuringIndex {
			wantGuarded := `\b` + beets.Pattern
			if guarded.Pattern != wantGuarded {
				t.Errorf("SDPatternsWordBoundary[%d].Pattern = %q, want %q", i, guarded.Pattern, wantGuarded)
			}
			if guarded.Weight != beets.Weight {
				t.Errorf("SDPatternsWordBoundary[%d].Weight = %v, want %v (unchanged)", i, guarded.Weight, beets.Weight)
			}
			continue
		}
		if guarded != beets {
			t.Errorf("SDPatternsWordBoundary[%d] = %+v, want byte-identical %+v", i, guarded, beets)
		}
	}
}

// TestSDPatternsNoBracketDiscount_OnlyBracketPatternsRemoved proves the
// table carries the other four discounts byte-identical to
// [BeetsSDPatterns] and drops exactly the two bracket patterns.
func TestSDPatternsNoBracketDiscount_OnlyBracketPatternsRemoved(t *testing.T) {
	want := []SDPattern{
		{Pattern: `^the `, Weight: 0.1},
		{Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
		{Pattern: `[\[\(]?(featuring|feat|ft)[\. :].+`, Weight: 0.1},
		{Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
	}
	if !reflect.DeepEqual(SDPatternsNoBracketDiscount, want) {
		t.Errorf("SDPatternsNoBracketDiscount = %v, want %v", SDPatternsNoBracketDiscount, want)
	}
}

// TestStringDistSeparatesNoBracketDiscount is the fixture that tells
// [SDPatternsNoBracketDiscount] apart from [BeetsSDPatterns], on
// adv-remix-both-sides's own pair (corpus/README.md:1186-1189): the
// bracket discount collapses "Glass Harbor (Nocturne Rework)" and
// "Glass Harbor (Original Mix)" to the same residual string and charges
// the whole differentiator at 0.3, reading 0.144000; removing the
// discount leaves the true, undiscounted 0.480000.
func TestStringDistSeparatesNoBracketDiscount(t *testing.T) {
	item := "Glass Harbor (Nocturne Rework)"
	candidate := "Glass Harbor (Original Mix)"

	tests := []struct {
		name     string
		patterns []SDPattern
		want     float64
	}{
		{"beets shipped table (bracket discount)", BeetsSDPatterns, 0.144},
		{"no-bracket-discount table", SDPatternsNoBracketDiscount, 0.48},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.StringDistPatterns = tt.patterns
			got := StringDist(tun, &item, &candidate)
			if got != tt.want {
				t.Errorf("StringDist = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSDEndWordsAndReplace(t *testing.T) {
	wantEndWords := []string{"the", "a", "an"}
	if !reflect.DeepEqual(SDEndWords, wantEndWords) {
		t.Errorf("SDEndWords = %v, want %v", SDEndWords, wantEndWords)
	}

	wantReplace := []SDReplacement{{Pattern: "&", Replacement: "and"}}
	if !reflect.DeepEqual(SDReplace, wantReplace) {
		t.Errorf("SDReplace = %v, want %v", SDReplace, wantReplace)
	}
}
func TestValidateRejectsAnUnknownRequiredFieldAndABadPattern(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*Tunables)
		want string
	}{
		{"unknown required field", func(t *Tunables) { t.Required = []string{"nosuchfield"} }, "nosuchfield"},
		{"unparseable preferred media pattern", func(t *Tunables) { t.Preferred.Media = []string{"("} }, "preferred.media"},
		{"unparseable preferred country pattern", func(t *Tunables) { t.Preferred.Countries = []string{"[a-"} }, "preferred.countries"},
		{"unresolvable max_rec level", func(t *Tunables) { t.MaxRec = map[string]string{"album": "not-a-level"} }, "max_rec"},
		{"quiet_fallback outside its legal set: apply", func(t *Tunables) { t.QuietFallback = "apply" }, "quiet_fallback"},
		{"quiet_fallback outside its legal set: ask", func(t *Tunables) { t.QuietFallback = "ask" }, "quiet_fallback"},
		{"none_rec_action outside its legal set: apply", func(t *Tunables) { t.NoneRecAction = "apply" }, "none_rec_action"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tt.mut(&tun)
			err := tun.Validate()
			if !errors.Is(err, ErrInvalidTunable) {
				t.Fatalf("Validate = %v, want an error wrapping ErrInvalidTunable", err)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("Validate = %v, want the message to name %q", err, tt.want)
			}
		})
	}
	if err := DefaultTunables().Validate(); err != nil {
		t.Errorf("DefaultTunables().Validate() = %v, want nil", err)
	}
}

func TestValidateReportsEveryProblemRatherThanTheFirst(t *testing.T) {
	tun := DefaultTunables()
	tun.Required = []string{"nosuchfield"}
	tun.Preferred.Media = []string{"("}
	err := tun.Validate()
	for _, want := range []string{"nosuchfield", "preferred.media"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("Validate = %v, want the joined message to name %q", err, want)
		}
	}
}
