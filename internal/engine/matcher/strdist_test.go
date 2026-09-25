package matcher

import (
	"encoding/json"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// Every want below was measured by calling
// beets.autotag.distance.string_dist in the pinned venv
// (tools/venv/.venv, beets 2.14.1) on exactly these inputs, and the
// comparison is exact rather than tolerant because bit identity is the
// port's target (contract 12). A row that needs a tolerance is a
// finding about operation order, not a test to loosen.
func TestStringDist(t *testing.T) {
	tun := DefaultTunables()
	tests := []struct {
		name string
		str1 *string
		str2 *string
		want float64
	}{
		{"both none", nil, nil, 0.0},
		{"left none", nil, new("Ember Wharf"), 1.0},
		{"right none", new("Ember Wharf"), nil, 1.0},
		{"identical", new("Ember Wharf"), new("Ember Wharf"), 0.0},
		{"case only", new("EMBER WHARF"), new("ember wharf"), 0.0},
		{"both empty", new(""), new(""), 0.0},
		{"punctuation only", new("!!!"), new("???"), 0.0},
		{"transliterated", new("Björk"), new("Bjork"), 0.0},

		// The article rotation, once per SD_END_WORDS entry. The
		// cascade row rotates twice: "the" first, which leaves a
		// string ending in ", a" for "a" to rotate in turn, so a port
		// that stops after the first hit fails it.
		{"rotate the", new("Northern Lines, The"), new("The Northern Lines"), 0.0},
		{"rotate a", new("Quiet Season, A"), new("A Quiet Season"), 0.0},
		{"rotate an", new("Ember Wharf, An"), new("An Ember Wharf"), 0.0},
		{"rotate cascade", new("Harbour, A, The"), new("The Harbour, A"), 0.0},

		{"ampersand", new("Ember & Wharf"), new("Ember and Wharf"), 0.0},
		{"ampersand mid word", new("R&B"), new("RandB"), 0.0},

		// The six discounts, each fired on its own.
		{"leading the", new("The Ember Wharf"), new("Ember Wharf"), 0.023076923076923078},
		{"ep is free", new("Ember Wharf (EP)"), new("Ember Wharf"), 0.0},
		{"single is free", new("Ember Wharf [Single]"), new("Ember Wharf"), 0.0},
		{"featuring", new("Ember Wharf featuring Northern Lines"), new("Ember Wharf"), 0.06875},
		{"feat", new("Ember Wharf feat Northern"), new("Ember Wharf"), 0.05454545454545454},
		{"parenthetical", new("Ember Wharf (Remix)"), new("Ember Wharf"), 0.09999999999999999},
		{"bracketed", new("Ember Wharf [Remix]"), new("Ember Wharf"), 0.09999999999999999},
		{"part", new("Ember Wharf part 2"), new("Ember Wharf"), 0.06666666666666667},
		{"pt", new("Ember Wharf, pt. 2"), new("Ember Wharf"), 0.046153846153846156},

		// The featuring trigger needs a word boundary on both sides, so
		// the "ft" inside "Draft" is not a featuring credit and the two
		// titles are charged their whole difference. See
		// TestStringDistFeaturingNeedsAWordBoundary.
		{"ft inside a word", new("Draft: Ember Wharf"), new("Draft: Northern Lines"), 0.6111111111111112},

		// The kept character: the featuring row's rewrite keeps the
		// rune its boundary group consumed instead of deleting it, so
		// "©", which unidecode maps to "(c)", survives into the
		// residual string both sides are compared on. This row already
		// passed before the rewrite, since Go's ASCII \b holds between
		// "©" and "f" too; it exists to fail the mutation that drops
		// the kept group.
		{"kept character before ft", new("Ember Wharf ©ft. Northern"), new("Ember Wharf ©"), 0.047619047619047616},

		// The three solo rows, where one side empties. "the " deletes
		// to nothing and earns the full 0.1; "(ep)" deletes to nothing
		// on both sides for a zero delta and earns nothing; ", an"
		// rotates to "an " and matches no pattern at all, so it stays
		// at the bare edit distance of 1.0.
		{"solo the", new("The "), new(""), 0.1},
		{"solo ep", new("(EP)"), new("(EP)"), 0.0},
		{"solo an", new(", An"), new(""), 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringDist(tun, tt.str1, tt.str2); got != tt.want {
				t.Errorf("StringDist(%s, %s) = %v, want %v", quote(tt.str1), quote(tt.str2), got, tt.want)
			}
		})
	}
}

// TestStringDistPatternLoopAdoptsOnlyOnPositiveDelta pins the pattern
// loop's baseline-shift arithmetic with one pair, measured at
// 0.09999999999999999 against pinned beets.
//
// The pair is the loop's discriminating fixture: each of these six
// plausible mutations of the loop changes its result, so the row fails
// if any one of them is made.
//
//	adopt on a zero delta                        0.11041666666666666
//	adopt but leave baseDist alone               0.5944444444444444
//	require both strings to change, not either   0.2833333333333333
//	drop the max(0, ...) clamp                   0.11041666666666666
//	anchor the patterns at ^                     0.29444444444444445
//	add the flat weight instead of weight*delta  0.4
func TestStringDistPatternLoopAdoptsOnlyOnPositiveDelta(t *testing.T) {
	const want = 0.09999999999999999
	got := StringDist(DefaultTunables(), new("The Quiet Season (EP)"), new("Quiet Season (EP) (Remix)"))
	if got != want {
		t.Errorf("StringDist = %v, want %v", got, want)
	}
}

// beetsSDPatternLiterals is SD_PATTERNS (autotag/distance.py:36-42),
// beets' six shipped discounts, pinned 2.14.1 byte for byte.
//
// The literals are written out here rather than compared against
// anything the package exports, so editing the shipped table takes two
// edits and cannot happen quietly. Tier 0 cannot stand in for this
// check: the featuring pattern's word boundaries change none of the 72
// committed rows, so tier 0 is blind to exactly the edit most likely to
// be made.
var beetsSDPatternLiterals = []SDPattern{
	{Pattern: `^the `, Weight: 0.1},
	{Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
	{Pattern: `[\[\(]?\b(featuring|feat|ft)\b[\. :].+`, Weight: 0.1},
	{Pattern: `\(.*?\)`, Weight: 0.3},
	{Pattern: `\[.*?\]`, Weight: 0.3},
	{Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
}

// sdGoForm is the Go pattern and Expand template a shipped row carries
// in place of one beets literal and its empty substitution.
type sdGoForm struct {
	pattern     string
	replacement string
}

// sdTranslations maps a beets literal to the Go form this package
// departs from, each written out independently of tunables.go's
// sdFeaturingPattern and sdFeaturingReplacement so this test does not
// merely compare production against itself. It has one entry today,
// the featuring row; a row whose literal is absent from this
// map must equal the literal exactly, with an empty Replacement, in
// every shipped table.
var sdTranslations = map[string]sdGoForm{
	`[\[\(]?\b(featuring|feat|ft)\b[\. :].+`: {
		pattern:     `[\[\(](?:featuring|feat|ft)[\. :].+|(^|[^\pL\pN_])(?:featuring|feat|ft)[\. :].+`,
		replacement: `${1}`,
	},
}

// beetsFeaturingRowIndex is the featuring row's position in
// BeetsSDPatterns and beetsSDPatternLiterals: its weight (0.1) is not
// unique in the table, so index is how this file, like the production
// loop's callers, locates it.
const beetsFeaturingRowIndex = 2

// wantSDPattern applies sdTranslations to literal, marking the
// translation used, so a shipped table is checked against what it
// should actually carry: the literal untouched, or the Go form on the
// one row this package departs from.
func wantSDPattern(literal SDPattern, used map[string]bool) SDPattern {
	translation, translated := sdTranslations[literal.Pattern]
	if !translated {
		return literal
	}
	used[literal.Pattern] = true
	return SDPattern{Pattern: translation.pattern, Weight: literal.Weight, Replacement: translation.replacement}
}

// assertMatchesTranslatedLiterals checks got against literals through
// sdTranslations and fails on any map entry no row in literals uses.
func assertMatchesTranslatedLiterals(t *testing.T, name string, got, literals []SDPattern) {
	t.Helper()
	if len(got) != len(literals) {
		t.Fatalf("%s has %d entries, beets ships %d", name, len(got), len(literals))
	}
	used := make(map[string]bool, len(sdTranslations))
	for i, literal := range literals {
		want := wantSDPattern(literal, used)
		if got[i] != want {
			t.Errorf("%s[%d] = {%q, %v, %q}, want {%q, %v, %q}", name, i, got[i].Pattern, got[i].Weight, got[i].Replacement, want.Pattern, want.Weight, want.Replacement)
		}
	}
	for literal := range sdTranslations {
		if !used[literal] {
			t.Errorf("%s: translation for %q unused", name, literal)
		}
	}
}

// TestBeetsSDPatternsMatchBeetsLiterals holds BeetsSDPatterns to
// beetsSDPatternLiterals through sdTranslations: byte-identical except
// for the featuring row, which carries its Go translation.
func TestBeetsSDPatternsMatchBeetsLiterals(t *testing.T) {
	assertMatchesTranslatedLiterals(t, "BeetsSDPatterns", BeetsSDPatterns, beetsSDPatternLiterals)

	wantFeaturing := wantSDPattern(beetsSDPatternLiterals[2], map[string]bool{})
	if DefaultTunables().StringDistPatterns[2].Pattern != wantFeaturing.Pattern {
		t.Error("DefaultTunables does not select the shipped table")
	}
}

// TestStringDistFeaturingNeedsAWordBoundary pins both sides of the
// featuring pattern's word boundary check, each figure measured against
// pinned beets. A trigger inside a word fires nothing; a trigger after
// a space or an opening bracket discounts the credit at 0.1. The
// pre-2.14 defect never checked the boundary before the trigger, and
// scored the "Draft:" and "Gift " rows at 0.061111 and 0.064706; beets
// 2.14.0 added the check and this row's rewrite carries it in two
// branches instead of \b (BeetsSDPatterns' doc comment). The trailing
// boundary is always satisfied, since the separator class that must
// follow the trigger holds no word character, so no row here can tell
// it apart.
func TestStringDistFeaturingNeedsAWordBoundary(t *testing.T) {
	tests := []struct {
		name       string
		str1, str2 string
		want       float64
	}{
		{"ft inside a word, colon", "Draft: Ember Wharf", "Draft: Northern Lines", 0.6111111111111112},
		{"ft inside a word, space", "Gift Wrapped Nights", "Gift Horse Blues", 0.6470588235294118},
		{"ft after a space", "Ember Wharf ft. Northern", "Ember Wharf", 0.05},
		{"feat after a parenthesis", "Ember Wharf (feat. Northern)", "Ember Wharf", 0.05454545454545454},
		{"ft after a bracket", "Ember Wharf [ft. Northern]", "Ember Wharf", 0.05},

		// A non-ASCII letter or number directly before the trigger,
		// where Go's ASCII-only \b saw a boundary and Python's
		// Unicode-aware one did not. Figures measured in the pinned
		// venv; the same for both pattern tables, since neither
		// table's other rows fire on these pairs.
		{"non-ASCII letter before ft, umlaut", "Geschäft: Ember Wharf", "Geschäft: Northern Lines", 0.5238095238095238},
		{"non-ASCII letter before ft, u-diaeresis", "Rüft: Ember Wharf", "Rüft: Northern Lines", 0.6470588235294118},
		{"non-ASCII letter before ft, omega", "Ωft Ember Wharf", "Ωft Northern Lines", 0.6875},
	}
	tables := []struct {
		name     string
		patterns []SDPattern
	}{
		{"beets shipped table", BeetsSDPatterns},
		{"no-bracket-discount table", SDPatternsNoBracketDiscount},
	}
	for _, table := range tables {
		for _, tt := range tests {
			t.Run(table.name+"/"+tt.name, func(t *testing.T) {
				tun := DefaultTunables()
				tun.StringDistPatterns = table.patterns
				if got := StringDist(tun, &tt.str1, &tt.str2); got != tt.want {
					t.Errorf("StringDist(%q, %q) = %v, want %v", tt.str1, tt.str2, got, tt.want)
				}
			})
		}
	}
}

// TestStringDistFeaturingUppercaseDottedICaseMapping declares, rather
// than fixes, one divergence upstream of the featuring row: Python's
// str.lower maps U+0130 (İ) to "i" plus a combining dot, a mark \w
// rejects, so beets' Unicode-aware boundary check still sees a
// boundary before "ft". Go's strings.ToLower maps U+0130 to a bare
// "i", so the character directly before the trigger is a plain ASCII
// word character on both sides and neither pattern table's boundary
// check fires. The two pipelines therefore charge the pair
// differently: beets discounts the credit at 0.06875 (pinned venv,
// 2026-09-25) and Go, seeing no boundary, charges the whole
// difference. A sweep of every code point found U+0130 is the only
// one whose Python lowercase changes whether it ends in a word
// character; a toolchain bump that moves Go's figure re-runs that
// sweep.
func TestStringDistFeaturingUppercaseDottedICaseMapping(t *testing.T) {
	const wantGo = 0.6875
	str1, str2 := "İft: Ember Wharf", "İft: Northern Lines"
	tables := []struct {
		name     string
		patterns []SDPattern
	}{
		{"beets shipped table", BeetsSDPatterns},
		{"no-bracket-discount table", SDPatternsNoBracketDiscount},
	}
	for _, table := range tables {
		t.Run(table.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.StringDistPatterns = table.patterns
			if got := StringDist(tun, &str1, &str2); got != wantGo {
				t.Errorf("StringDist(%q, %q) = %v, want %v (declared divergence from beets' 0.06875; a figure equal to 0.06875 means the divergence was fixed, so update this row and its HISTORY note)", str1, str2, got, wantGo)
			}
		})
	}
}

// wantCuratedFeaturingLabels lists every curated label
// featuring-golden.json must carry, so a curated row going
// missing from the file is a named failure rather than a silently
// weaker golden.
var wantCuratedFeaturingLabels = []string{
	"geschaft-mixed", "geschaft-lower",
	"ruft-mixed", "ruft-lower",
	"omegaft-mixed", "omegaft-lower",
	"draft-mixed", "draft-lower",
	"arft-mixed", "arft-lower",
	"twoft-mixed", "twoft-lower",
	"cafeft-mixed", "cafeft-lower",
	"emberwharfft-mixed", "emberwharfft-lower",
	"a-umlaut", "sharp-s", "omega", "titlecase-dz-caron",
	"superscript-two", "roman-numeral-twelve", "combining-acute",
	"underscore", "no-break-space", "paren-open", "bracket-open",
	"copyright", "trademark", "euro",
	"start", "two-triggers", "newline",
}

// featuringGoldenFile is featuring-golden.json's shape: a
// provenance header, curated [label, input, beets_output] rows, and
// drawn [input, beets_output] pairs.
type featuringGoldenFile struct {
	Header struct {
		BeetsVersion       string `json:"beets_version"`
		UnicodedataVersion string `json:"unicodedata_version"`
		Literal            string `json:"literal"`
	} `json:"header"`
	Curated [][3]string `json:"curated"`
	Drawn   [][2]string `json:"drawn"`
}

// The golden's two floors: below either, it is too weak to tell the
// Go translation from a defect that merely deletes less than beets
// does.
const (
	minFiredDrawnFeaturingRows = 150
	minASCIILiteralDivergences = 40
)

// TestFeaturingRowMatchesBeetsGolden runs the shipped featuring row
// over every curated and drawn input in featuring-golden.json and
// requires beets' own substitution on each. The ASCII-literal control
// compiles the golden's own recorded beets literal with Go's regexp
// (Go's \b, ASCII-only) and counts how often that undiscounted port
// disagrees with beets; that count, pinned at gen_featuring_golden.py's
// MIN_FIRED-checked draw, is the test's own non-vacuity proof: a golden
// too weak to separate the translation from the defect it replaces
// would pass here for the wrong reason.
func TestFeaturingRowMatchesBeetsGolden(t *testing.T) {
	raw, err := os.ReadFile("testdata/featuring-golden.json")
	if err != nil {
		t.Fatalf("read featuring-golden.json: %v", err)
	}
	var golden featuringGoldenFile
	if err := json.Unmarshal(raw, &golden); err != nil {
		t.Fatalf("decode featuring-golden.json: %v", err)
	}

	if golden.Header.BeetsVersion != "2.14.1" {
		t.Fatalf("header.beets_version = %q, want %q", golden.Header.BeetsVersion, "2.14.1")
	}
	if golden.Header.UnicodedataVersion != "15.1.0" {
		t.Fatalf("header.unicodedata_version = %q, want %q", golden.Header.UnicodedataVersion, "15.1.0")
	}
	if golden.Header.Literal != beetsSDPatternLiterals[2].Pattern {
		t.Fatalf("header.literal = %q, want %q", golden.Header.Literal, beetsSDPatternLiterals[2].Pattern)
	}

	const wantCuratedRows = 33
	if len(golden.Curated) != wantCuratedRows {
		t.Fatalf("golden carries %d curated rows, want %d", len(golden.Curated), wantCuratedRows)
	}
	const wantDrawnRows = 7955
	if len(golden.Drawn) != wantDrawnRows {
		t.Fatalf("golden carries %d drawn rows, want %d", len(golden.Drawn), wantDrawnRows)
	}

	gotLabels := make(map[string]bool, len(golden.Curated))
	for _, row := range golden.Curated {
		gotLabels[row[0]] = true
	}
	for _, label := range wantCuratedFeaturingLabels {
		if !gotLabels[label] {
			t.Errorf("golden is missing curated label %q", label)
		}
	}

	fired := 0
	for _, row := range golden.Drawn {
		if row[0] != row[1] {
			fired++
		}
	}
	if fired < minFiredDrawnFeaturingRows {
		t.Fatalf("golden's drawn rows fired %d times, want at least %d: too weak to prove the row", fired, minFiredDrawnFeaturingRows)
	}
	const wantFired = 251
	if fired != wantFired {
		t.Errorf("golden's drawn rows fired %d times, want %d", fired, wantFired)
	}

	asciiLiteral := sdRegexp(golden.Header.Literal)
	featuring := BeetsSDPatterns[beetsFeaturingRowIndex]
	shipped := sdRegexp(featuring.Pattern)

	type divergence struct {
		label, input, want, got string
	}
	var diverged []divergence
	asciiDivergences := 0
	for _, row := range golden.Curated {
		label, input, want := row[0], row[1], row[2]
		if got := asciiLiteral.ReplaceAllString(input, ""); got != want {
			asciiDivergences++
		}
		if got := shipped.ReplaceAllString(input, featuring.Replacement); got != want {
			diverged = append(diverged, divergence{label, input, want, got})
		}
	}
	for _, row := range golden.Drawn {
		input, want := row[0], row[1]
		if got := asciiLiteral.ReplaceAllString(input, ""); got != want {
			asciiDivergences++
		}
		if got := shipped.ReplaceAllString(input, featuring.Replacement); got != want {
			diverged = append(diverged, divergence{"", input, want, got})
		}
	}

	if len(diverged) > 0 {
		for i, d := range diverged {
			if i >= 20 {
				break
			}
			name := d.label
			if name == "" {
				name = strconv.Quote(d.input)
			}
			t.Errorf("row %s: got %q, want %q", name, d.got, d.want)
		}
		t.Errorf("%d of %d golden rows diverge from beets", len(diverged), len(golden.Curated)+len(golden.Drawn))
	}

	const wantASCIILiteralDivergences = 98
	if asciiDivergences < minASCIILiteralDivergences {
		t.Fatalf("ASCII-literal control diverged on %d rows, want at least %d: the golden cannot tell the translation from the defect", asciiDivergences, minASCIILiteralDivergences)
	}
	if asciiDivergences != wantASCIILiteralDivergences {
		t.Errorf("ASCII-literal control diverged on %d rows, want %d", asciiDivergences, wantASCIILiteralDivergences)
	}
}

// wantGoOnlyWordCodePoints is the measured gap between Go
// 1.27.1's unicode.Version 17.0.0 and Python's word class at CPython
// 3.13.15's unicodedata.unidata_version 15.1.0: 9,039 code points
// Unicode 16.0 and 17.0 assigned that 15.1 leaves unassigned, and that
// the shipped featuring row therefore reads as word characters where
// beets' own Python does not.
const wantGoOnlyWordCodePoints = 9039

// pythonWordRangesFile is python-word-ranges.json's shape:
// the same provenance header as the featuring golden, and Python's
// whole \w class as sorted, non-overlapping inclusive hex ranges.
type pythonWordRangesFile struct {
	Header struct {
		BeetsVersion       string `json:"beets_version"`
		UnicodedataVersion string `json:"unicodedata_version"`
		Literal            string `json:"literal"`
	} `json:"header"`
	Ranges [][2]string `json:"ranges"`
}

// TestFeaturingRowWordClassMatchesPythonRanges classifies every code
// point outside the surrogates through the shipped featuring row
// itself, never a copy of its class, by substituting on
// "a"+string(r)+"ft. x": an unchanged output means r is a Go word
// character, and an output of "a"+string(r) means it is not (the
// bracket branch removes "(" and "[" outright, so their non-word
// output is "a" alone). It compares that classification, code point by
// code point, with python-word-ranges.json.
func TestFeaturingRowWordClassMatchesPythonRanges(t *testing.T) {
	raw, err := os.ReadFile("testdata/python-word-ranges.json")
	if err != nil {
		t.Fatalf("read python-word-ranges.json: %v", err)
	}
	var file pythonWordRangesFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("decode python-word-ranges.json: %v", err)
	}
	if file.Header.UnicodedataVersion != "15.1.0" {
		t.Fatalf("header.unicodedata_version = %q, want %q", file.Header.UnicodedataVersion, "15.1.0")
	}

	const wantRangeCount = 749
	if len(file.Ranges) != wantRangeCount {
		t.Fatalf("range file carries %d ranges, want %d", len(file.Ranges), wantRangeCount)
	}

	ranges := make([][2]rune, len(file.Ranges))
	for i, hexRange := range file.Ranges {
		lo, err := strconv.ParseUint(hexRange[0], 16, 32)
		if err != nil {
			t.Fatalf("range %d: parse low bound %q: %v", i, hexRange[0], err)
		}
		hi, err := strconv.ParseUint(hexRange[1], 16, 32)
		if err != nil {
			t.Fatalf("range %d: parse high bound %q: %v", i, hexRange[1], err)
		}
		if lo > hi {
			t.Fatalf("range %d (%s-%s) is empty: low bound exceeds high bound", i, hexRange[0], hexRange[1])
		}
		ranges[i] = [2]rune{rune(lo), rune(hi)}
		if i > 0 && ranges[i][0] <= ranges[i-1][1] {
			t.Fatalf("range %d (%s-%s) overlaps or is unsorted against range %d", i, hexRange[0], hexRange[1], i-1)
		}
	}

	isPythonWord := func(r rune) bool {
		_, found := slices.BinarySearchFunc(ranges, r, func(rg [2]rune, target rune) int {
			switch {
			case target < rg[0]:
				return 1
			case target > rg[1]:
				return -1
			default:
				return 0
			}
		})
		return found
	}

	featuring := BeetsSDPatterns[beetsFeaturingRowIndex]
	shipped := sdRegexp(featuring.Pattern)
	var pythonOnly []rune
	goOnly := 0
	for r := rune(0); r <= 0x10FFFF; r++ {
		if r >= 0xD800 && r <= 0xDFFF {
			continue
		}
		input := "a" + string(r) + "ft. x"
		got := shipped.ReplaceAllString(input, featuring.Replacement)

		goWord := got == input
		switch {
		case goWord:
		case r == '(' || r == '[':
			if got != "a" {
				t.Fatalf("U+%04X: got %q, want %q (the bracket branch removes it)", r, got, "a")
			}
		case got == "a"+string(r):
		default:
			t.Fatalf("U+%04X: got %q, neither unchanged nor the non-word form", r, got)
		}

		switch pyWord := isPythonWord(r); {
		case pyWord && !goWord:
			pythonOnly = append(pythonOnly, r)
		case goWord && !pyWord:
			goOnly++
		}
	}
	if len(pythonOnly) > 0 {
		for _, r := range pythonOnly[:min(5, len(pythonOnly))] {
			t.Errorf("U+%04X is a Python word character and not a Go one", r)
		}
		t.Fatalf("%d code points are Python word characters and not Go ones", len(pythonOnly))
	}
	if goOnly != wantGoOnlyWordCodePoints {
		t.Errorf("Go-only word-character count = %d, want %d (unicode.Version %s): confirm the Python-only count is still zero, check whether any newly assigned code point can sit before a trigger in a real title, then update the count and record it in HISTORY", goOnly, wantGoOnlyWordCodePoints, unicode.Version)
	}
}

// TestStringDistLowercasesBeforeTransliterating proves port hazard 1:
// string_dist lowercases before _string_dist_basic transliterates, and
// the pre-lowercase is not redundant with the one that follows the
// transliteration.
//
// unidecode maps the Greek capital xi to "Ks" and the small xi to "x",
// so the two orders give different letters and not merely different
// case. Deleting the strings.ToLower calls at the head of StringDist
// turns this 0.0 into 1.0, which is the mutation that proves the
// ordering is real.
func TestStringDistLowercasesBeforeTransliterating(t *testing.T) {
	if got := StringDist(DefaultTunables(), new("Ξ"), new("x")); got != 0.0 {
		t.Errorf("StringDist(Ξ, x) = %v, want 0", got)
	}
	if got := sdStrip("Ξ"); got != "ks" {
		t.Errorf("sdStrip(Ξ) = %q, want %q: the pre-lowercase is what makes the pipeline read it as x", got, "ks")
	}
}

// TestStringDistDoesNotNormalizeUnicode proves port hazard 2: nothing
// on beets' distance path applies NFC, NFD, NFKC, or NFKD, so the Go
// port must not either.
//
// Evidence from the pinned venv: grepping beets/autotag/ and
// beetsplug/musicbrainz.py for unicodedata, normalize, NFC, and NFD
// returns nothing but prose in docstrings, and the loaded
// levenshtein_distance is jellyfish's Rust implementation, which does
// not normalize. The deprecated pure-Python jellyfish module defines a
// _normalize helper but never calls it from levenshtein_distance.
//
// The pair below is the observable consequence. Composed U+304C
// transliterates to "ga"; the same character decomposed to U+304B
// U+3099 transliterates to "ka" plus a mark that strips away, so beets
// scores the two forms 0.5 apart. A port that normalized either way
// would score 0.0.
func TestStringDistDoesNotNormalizeUnicode(t *testing.T) {
	const (
		composed   = "が"
		decomposed = "が"
	)
	if got := StringDist(DefaultTunables(), new(composed), new(decomposed)); got != 0.5 {
		t.Errorf("StringDist(composed, decomposed) = %v, want 0.5", got)
	}
	if got := sdStrip(decomposed); got != "ka" {
		t.Errorf("sdStrip(decomposed) = %q, want %q", got, "ka")
	}
}

// TestStringDistCaseMappingSweep closes the two places Python's
// str.lower and Go's strings.ToLower can part company, and pins the one
// that turns out to matter.
//
// Sweeping all 1112064 valid code points found 56 where the two
// disagree. The fixture lists them with the post-strip form pinned
// beets produces for each, and this test replays them through the Go
// pipeline. The 56 group three ways: 1 full-versus-simple case mapping,
// 53 a Unicode version skew that strips to nothing on both sides, and 2
// the same skew surviving the strip.
//
// The mapping difference is U+0130, the case the port was checked for.
// Python expands it to "i" plus a combining dot and Go maps it straight
// to "i". The strip this test sweeps removes the dot either way, so
// the two agree here; the featuring row's boundary check runs before
// any strip and sees the two lowercase forms differently, and that
// divergence is declared, not fixed
// (TestStringDistFeaturingUppercaseDottedICaseMapping).
//
// The skew is Go 1.27's Unicode 17.0.0 case tables against CPython
// 3.13's 15.1.0, so Go lowercases code points Python leaves alone. The
// two survivors are the finding this test pins: U+A7CB and U+A7DC are
// unassigned in Unicode 15.1, and unidecode 1.4.0 maps their Unicode
// 16.0 lowercase forms (U+0264 and U+019B) to letters while mapping the
// uppercase forms to nothing. Go's pipeline therefore yields a letter
// where beets yields nothing.
//
// The consequence is narrow today and it is real: the port's fidelity
// to beets depends on the toolchain's Unicode version and not only on
// the pinned unidecode table, and a Go release that adds another such
// case pair widens it. The divergence is written out below rather than
// tolerated, so a third one reds the gate instead of passing quietly.
func TestStringDistCaseMappingSweep(t *testing.T) {
	const wantRows = 56

	// The post-strip form Go produces for the two code points where it
	// and pinned beets disagree, against beets' own empty result.
	knownDivergences := map[rune]string{0xA7CB: "u", 0xA7DC: "l"}

	fixture, err := os.ReadFile("testdata/case-mapping-sweep.tsv")
	if err != nil {
		t.Fatalf("read sweep fixture: %v", err)
	}

	rows, nonEmpty, diverged := 0, 0, 0
	for line := range strings.Lines(string(fixture)) {
		line = strings.TrimSuffix(line, "\n")
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 4 {
			t.Fatalf("row %q: want 4 fields, got %d", line, len(fields))
		}
		cp, err := strconv.ParseUint(fields[0], 16, 32)
		if err != nil {
			t.Fatalf("row %q: parse code point: %v", line, err)
		}
		rows++
		want := fields[3]
		if want != "" {
			nonEmpty++
		}
		if divergent, known := knownDivergences[rune(cp)]; known {
			diverged++
			want = divergent
		}
		// The outer lowercase is string_dist's own, and it is the step
		// whose case mapping differs between the languages.
		if got := sdStrip(strings.ToLower(string(rune(cp)))); got != want {
			t.Errorf("U+%04X: Go strips to %q, want %q", cp, got, want)
		}
	}
	if rows != wantRows {
		t.Errorf("swept %d code points, fixture should carry %d", rows, wantRows)
	}
	if nonEmpty == 0 {
		t.Error("every fixture row expects an empty strip, so the sweep proves nothing")
	}
	if diverged != len(knownDivergences) {
		t.Errorf("matched %d known divergences, declared %d", diverged, len(knownDivergences))
	}
}

func TestStringDistBasic(t *testing.T) {
	tests := []struct {
		name string
		str1 string
		str2 string
		want float64
	}{
		{"both strip to nothing", "(((", "]]]", 0.0},
		{"one strips to nothing", "abc", "!!!", 1.0},
		{"identical", "Ember", "ember", 0.0},
		{"one substitution", "abc", "abd", 1.0 / 3.0},
		{"longer string normalizes", "abcdef", "abc", 3.0 / 6.0},
		// The normalizing length counts the stripped string, not the
		// input: four multi-byte runes strip to four ASCII bytes, so
		// the denominator is 4 and not 8.
		{"multi byte input", "Ünnö", "unno", 0.0},
		{"multi byte denominator", "Ünnö", "unn", 1.0 / 4.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringDistBasic(tt.str1, tt.str2); got != tt.want {
				t.Errorf("stringDistBasic(%q, %q) = %v, want %v", tt.str1, tt.str2, got, tt.want)
			}
		})
	}
}

// TestAddString proves the eighth adder writes string_dist's own number
// into the accumulator under its key, rather than a rounded or clamped
// stand-in.
func TestAddString(t *testing.T) {
	dist := NewDistance(DefaultTunables())
	dist.AddString("album", new("The Ember Wharf"), new("Ember Wharf"))
	dist.AddString("artist", new("Northern Lines"), new("Northern Lines"))

	const wantAlbum = 0.023076923076923078
	if got := dist.penalties["album"]; len(got) != 1 || got[0] != wantAlbum {
		t.Errorf("album penalties = %v, want [%v]", got, wantAlbum)
	}
	if got := dist.AddedKeys(); len(got) != 2 || got[0] != "album" || got[1] != "artist" {
		t.Errorf("AddedKeys() = %v, want [album artist]", got)
	}
	if got := dist.NonzeroKeys(); len(got) != 1 || got[0] != "album" {
		t.Errorf("NonzeroKeys() = %v, want [album]: an identical pair adds 0.0 and stays out", got)
	}
}

func quote(s *string) string {
	if s == nil {
		return "nil"
	}
	return strconv.Quote(*s)
}
