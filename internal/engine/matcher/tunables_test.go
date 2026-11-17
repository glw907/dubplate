package matcher

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

// TestSDPatternsNoBracketDiscount_OnlyBracketPatternsRemoved proves the
// table carries the other four discounts byte-identical to beets'
// literals, through the same sdTranslations map
// TestBeetsSDPatternsMatchBeetsLiterals uses (so no test outside that
// map ever writes the featuring row's Go form), and drops exactly the
// two bracket patterns.
func TestSDPatternsNoBracketDiscount_OnlyBracketPatternsRemoved(t *testing.T) {
	var literalsLessBrackets []sdLiteral
	for _, literal := range beetsSDPatternLiterals {
		if literal.weight == 0.3 {
			continue
		}
		literalsLessBrackets = append(literalsLessBrackets, literal)
	}
	assertMatchesTranslatedLiterals(t, "SDPatternsNoBracketDiscount", SDPatternsNoBracketDiscount, literalsLessBrackets)
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

func TestValidateRejectsAnUnknownRequiredFieldAndABadPattern(t *testing.T) {
	tests := []struct {
		name string
		mut  func(*Tunables)
		want string
	}{
		{"unknown required field", func(t *Tunables) { t.Required = []string{"nosuchfield"} },
			`required: "nosuchfield": invalid tunable`},
		{"unparseable preferred media pattern", func(t *Tunables) { t.Preferred.Media = []string{"("} },
			"preferred.media: invalid tunable: matcher: preferred.media pattern ( does not compile: compile option pattern \"(\\\\d+x)?(()\": error parsing regexp: missing closing ): `(?i)^(?:(\\d+x)?(())`"},
		{"unparseable preferred country pattern", func(t *Tunables) { t.Preferred.Countries = []string{"("} },
			"preferred.countries: invalid tunable: matcher: preferred.countries pattern ( does not compile: compile option pattern \"(\": error parsing regexp: missing closing ): `(?i)^(?:()`"},
		{"unresolvable max_rec level", func(t *Tunables) { t.MaxRec = map[string]string{"album": "not-a-level"} },
			`max_rec["album"]: recommendation "not-a-level": invalid tunable`},
		{"quiet_fallback outside its legal set: apply", func(t *Tunables) { t.QuietFallback = "apply" },
			`import.quiet_fallback: "apply": invalid tunable`},
		{"quiet_fallback outside its legal set: ask", func(t *Tunables) { t.QuietFallback = "ask" },
			`import.quiet_fallback: "ask": invalid tunable`},
		{"none_rec_action outside its legal set: apply", func(t *Tunables) { t.NoneRecAction = "apply" },
			`import.none_rec_action: "apply": invalid tunable`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tt.mut(&tun)
			err := tun.Validate()
			if !errors.Is(err, ErrInvalidTunable) {
				t.Fatalf("Validate = %v, want an error wrapping ErrInvalidTunable", err)
			}
			if got := err.Error(); got != tt.want {
				t.Errorf("Validate = %q, want %q", got, tt.want)
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

func TestQuietImportTunablesAppliesTheFourImportAssignments(t *testing.T) {
	def := DefaultTunables()
	got := QuietImportTunables()

	if !got.Quiet {
		t.Error("Quiet = false, want true")
	}
	if got.QuietFallback != "skip" {
		t.Errorf("QuietFallback = %q, want %q", got.QuietFallback, "skip")
	}
	if got.NoneRecAction != "skip" {
		t.Errorf("NoneRecAction = %q, want %q", got.NoneRecAction, "skip")
	}
	if got.Timid {
		t.Error("Timid = true, want false")
	}

	// Nothing else moves. The whole distance and threshold surface runs
	// at beets' shipped defaults, because the estate's beets config sets
	// no match: block and no musicbrainz: block
	// (internal/harness/conform/defaults.go:16-18).
	got.Quiet, got.QuietFallback, got.NoneRecAction, got.Timid =
		def.Quiet, def.QuietFallback, def.NoneRecAction, def.Timid
	if !reflect.DeepEqual(got, def) {
		t.Error("QuietImportTunables moved a field other than the four import assignments")
	}
}
