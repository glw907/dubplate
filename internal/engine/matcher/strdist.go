package matcher

import (
	"regexp"
	"strings"
	"sync"
)

// StringDist is beets' string_dist (autotag/distance.py:63-120): an edit
// distance normalized by string length, discounted for the parts of a
// title that carry little identifying weight.
//
// A nil pointer is Python's None. Two nils score 0.0 and one nil scores
// 1.0, which is beets' own asymmetry: an absent field on both sides is
// agreement, an absent field on one side is total disagreement. The
// vocabulary types in info.go never produce a nil, so a nil here comes
// from a caller modeling an unset source field.
//
// The pipeline runs lowercase, article rotation, SD_REPLACE, then the
// pattern loop, in that order, and the order is load-bearing at every
// step (contract 1). The lowercase happens here and again inside
// sdStrip after transliteration, and both are needed: see
// TestStringDistLowercasesBeforeTransliterating.
//
// The pattern table comes from tun rather than from [BeetsSDPatterns]
// directly, so 3d's compat switch selects a table by assignment instead
// of editing this function.
func StringDist(tun Tunables, str1, str2 *string) float64 {
	if str1 == nil && str2 == nil {
		return 0
	}
	if str1 == nil || str2 == nil {
		return 1
	}

	a := applySDReplace(rotateEndWords(strings.ToLower(*str1)))
	b := applySDReplace(rotateEndWords(strings.ToLower(*str2)))

	baseDist := stringDistBasic(a, b)
	var penalty float64
	for _, pattern := range tun.StringDistPatterns {
		re := sdRegexps.compile(pattern.Pattern)
		caseA := re.ReplaceAllString(a, "")
		caseB := re.ReplaceAllString(b, "")
		if caseA == a && caseB == b {
			continue
		}
		// A pattern that deletes text the two strings disagree about
		// earns its discount and shifts the baseline. One that deletes
		// text they already agreed on earns nothing and must leave the
		// baseline alone, because the strings it would install decide
		// what the remaining patterns match. That is why the six
		// patterns do not commute, and adopting on a zero delta is the
		// easiest way to get this loop wrong.
		caseDist := stringDistBasic(caseA, caseB)
		caseDelta := max(0, baseDist-caseDist)
		if caseDelta == 0 {
			continue
		}
		a, b = caseA, caseB
		baseDist = caseDist
		penalty += pattern.Weight * caseDelta
	}
	return baseDist + penalty
}

// AddString adds the string distance between str1 and str2 under key
// (distance.py:350-355). A nil is Python's None; see [StringDist].
func (d *Distance) AddString(key string, str1, str2 *string) {
	d.Add(key, StringDist(d.tun, str1, str2))
}

// stringDistBasic is _string_dist_basic (distance.py:47-60): the edit
// distance between two strings reduced to their ASCII alphanumerics,
// normalized by the longer of the two.
//
// The normalizing length is the byte length of the stripped string,
// which is the same as its rune count because the strip leaves only
// [a-z0-9]. Do not "fix" it to utf8.RuneCountInString: the two agree
// here, and the byte form says the strip already happened.
//
// Two strings that strip to nothing score 0.0 rather than dividing by
// zero, so a pair that is all punctuation reads as agreement.
func stringDistBasic(str1, str2 string) float64 {
	a, b := sdStrip(str1), sdStrip(str2)
	if a == "" && b == "" {
		return 0
	}
	return float64(Levenshtein(a, b)) / float64(max(len(a), len(b)))
}

// sdStrip transliterates s, lowercases it, and drops every character
// outside [a-z0-9], which is the normalization both halves of
// _string_dist_basic perform before measuring.
//
// The lowercase after transliteration is not redundant with the one
// [StringDist] applies before it: unidecode maps an uppercase source
// rune to uppercase ASCII, and it maps some runes to different letters
// depending on their case.
func sdStrip(s string) string {
	var stripped strings.Builder
	stripped.Grow(len(s))
	for _, r := range strings.ToLower(Transliterate(s)) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			stripped.WriteByte(byte(r))
		}
	}
	return stripped.String()
}

// rotateEndWords moves a trailing ", the", ", a", or ", an" to the
// front of s, so that "something, the" scores equal to "the something"
// (distance.py:75-83).
//
// The three words are tried in [SDEndWords] order and each rotation
// feeds the next, so "harbour, a, the" rotates twice.
func rotateEndWords(s string) string {
	for _, word := range SDEndWords {
		if prefix, found := strings.CutSuffix(s, ", "+word); found {
			s = word + " " + prefix
		}
	}
	return s
}

// applySDReplace applies [SDReplace] to s (distance.py:85-87).
//
// The substitution is literal on the replacement side. Python's re.sub
// expands backslash group references there and Go's ReplaceAllString
// expands dollar ones, so neither language's default is a faithful port
// of the other; beets' shipped table carries no group reference, and
// the literal form is the one that cannot silently reinterpret a Go
// dollar sign in a later entry.
func applySDReplace(s string) string {
	for _, replacement := range SDReplace {
		s = sdRegexps.compile(replacement.Pattern).ReplaceAllLiteralString(s, replacement.Replacement)
	}
	return s
}

// sdRegexpCache compiles each distinct SD_PATTERNS and SD_REPLACE
// expression once and hands out the compiled form.
//
// Both shipped tables and SD_REPLACE are compiled during package
// initialization, so a malformed literal in either is a startup panic
// rather than a per-call error. A table assigned into
// [Tunables.StringDistPatterns] from outside this package compiles on
// first use and panics the same way, because a pattern table is code
// and a broken one is a programming error rather than an input.
//
// The lock guards package state that outlives any one scoring path. No
// scoring path in this package is concurrent; a test binary running two
// of them is.
type sdRegexpCache struct {
	mu       sync.Mutex
	compiled map[string]*regexp.Regexp
}

var sdRegexps = newSDRegexpCache(BeetsSDPatterns, SDPatternsWordBoundary)

func newSDRegexpCache(tables ...[]SDPattern) *sdRegexpCache {
	cache := &sdRegexpCache{compiled: make(map[string]*regexp.Regexp)}
	for _, table := range tables {
		for _, pattern := range table {
			cache.compile(pattern.Pattern)
		}
	}
	for _, replacement := range SDReplace {
		cache.compile(replacement.Pattern)
	}
	return cache
}

// compile returns pattern compiled unanchored, exactly as beets writes
// it. SD_PATTERNS are re.sub arguments and re.sub never anchors, so the
// only "^" in the table is beets' own in "^the " (contract 7). The
// anchoring [CompilePatternOption] applies belongs to Distance._eq's
// options and to nothing here.
func (c *sdRegexpCache) compile(pattern string) *regexp.Regexp {
	c.mu.Lock()
	defer c.mu.Unlock()
	if re, cached := c.compiled[pattern]; cached {
		return re
	}
	re := regexp.MustCompile(pattern)
	c.compiled[pattern] = re
	return re
}
