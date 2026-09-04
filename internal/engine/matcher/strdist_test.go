package matcher

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// Every want below was measured by calling
// beets.autotag.distance.string_dist in the pinned venv
// (tools/venv/.venv, beets 2.13.1) on exactly these inputs, and the
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

		// The bug-ft-pattern defect, shipped on purpose: "ft" inside
		// "Draft" is read as a featuring credit, so the colon and
		// everything after it is discounted away. See
		// TestStringDistSeparatesPatternTables.
		{"ft inside a word", new("Draft: Ember Wharf"), new("Draft: Northern Lines"), 0.061111111111111116},

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

// TestBeetsSDPatternsMatchBeetsLiterals holds BeetsSDPatterns to the
// six literals at beets/autotag/distance.py:35-42, pinned 2.13.1, byte
// for byte.
//
// The literals are written out here rather than compared against
// anything the package exports, so correcting the shipped table takes
// two edits and cannot happen quietly. Tier 0 cannot stand in for this
// check: the word-boundary guard on the featuring pattern changes none
// of the 72 committed rows, so the corpus is blind to exactly the edit
// most likely to be made.
func TestBeetsSDPatternsMatchBeetsLiterals(t *testing.T) {
	shipped := []SDPattern{
		{Pattern: `^the `, Weight: 0.1},
		{Pattern: `[\[\(]?(ep|single)[\]\)]?`, Weight: 0.0},
		{Pattern: `[\[\(]?(featuring|feat|ft)[\. :].+`, Weight: 0.1},
		{Pattern: `\(.*?\)`, Weight: 0.3},
		{Pattern: `\[.*?\]`, Weight: 0.3},
		{Pattern: `(, )?(pt\.|part) .+`, Weight: 0.2},
	}
	if len(BeetsSDPatterns) != len(shipped) {
		t.Fatalf("BeetsSDPatterns has %d entries, beets ships %d", len(BeetsSDPatterns), len(shipped))
	}
	for i, want := range shipped {
		if got := BeetsSDPatterns[i]; got != want {
			t.Errorf("BeetsSDPatterns[%d] = {%q, %v}, beets ships {%q, %v}", i, got.Pattern, got.Weight, want.Pattern, want.Weight)
		}
	}
	if DefaultTunables().StringDistPatterns[2].Pattern != shipped[2].Pattern {
		t.Error("DefaultTunables does not select the shipped table")
	}
}

// TestStringDistSeparatesPatternTables is the fixture that tells the
// two shipped tables apart, so SDPatternsWordBoundary is a measured
// value rather than dead weight.
//
// Both figures were measured against pinned beets, the second by
// substituting the word-boundary table into the module before calling
// string_dist. They differ by an order of magnitude because
// BeetsSDPatterns reads the "ft" in "Draft:" as a featuring credit and
// discounts everything after it, which is the defect
// adj-adv-ft-draft-colon names.
func TestStringDistSeparatesPatternTables(t *testing.T) {
	tests := []struct {
		name     string
		patterns []SDPattern
		want     float64
	}{
		{"beets shipped table", BeetsSDPatterns, 0.061111111111111116},
		{"word boundary table", SDPatternsWordBoundary, 0.6111111111111112},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tun := DefaultTunables()
			tun.StringDistPatterns = tt.patterns
			got := StringDist(tun, new("Draft: Ember Wharf"), new("Draft: Northern Lines"))
			if got != tt.want {
				t.Errorf("StringDist = %v, want %v", got, tt.want)
			}
		})
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
// pipeline. Only one of the 56, U+0130, is the full-versus-simple case
// mapping difference the port was checked for: Python expands it to "i"
// plus a combining dot and Go maps it straight to "i", and the strip
// removes the dot, so the two agree.
//
// The other 55 are a Unicode version skew. Go 1.27 carries Unicode
// 17.0.0 case tables and CPython 3.13 carries 15.1.0, so Go lowercases
// code points Python leaves alone. Fifty-three of those transliterate
// to nothing either way. Two do not, and they are the finding this test
// pins: U+A7CB and U+A7DC are unassigned in Unicode 15.1, and unidecode
// 1.4.0 maps their Unicode 16.0 lowercase forms (U+0264 and U+019B) to
// letters while mapping the uppercase forms to nothing. Go's pipeline
// therefore yields a letter where beets yields nothing.
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
