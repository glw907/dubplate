package matcher

import (
	"strings"
	"sync/atomic"
)

// transliterator holds the process-wide function Transliterate
// consults for every rune. It starts nil: a program installs the
// pinned Unidecode table's lookup through SetTransliterator before
// any scoring call, and a call before that panics, the way
// crypto.Hash.New panics for a hash no program registered.
var transliterator atomic.Pointer[func(rune) (string, bool)]

// SetTransliterator installs fn as the transliterator every
// Transliterate call consults from this point on. A program calls it
// once, before scoring. It panics on a nil fn: there is no channel to
// uninstall a transliterator once one is installed.
func SetTransliterator(fn func(rune) (string, bool)) {
	if fn == nil {
		panic("matcher: SetTransliterator called with a nil function")
	}
	transliterator.Store(&fn)
}

// Transliterate reproduces `unidecode(s)`'s default errors="ignore"
// behavior (autotag/distance.py:9,55-56): every rune maps to its
// ASCII approximation, and a rune with no mapping contributes nothing.
// It panics, naming SetTransliterator, if nothing has been installed:
// an uninstalled seam is a wiring defect in a program, never a
// runtime condition to degrade gracefully from.
func Transliterate(s string) string {
	fn := transliterator.Load()
	if fn == nil {
		panic("matcher: Transliterate called before SetTransliterator installed a transliterator")
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		repl, _ := (*fn)(r)
		b.WriteString(repl)
	}
	return b.String()
}

// Levenshtein is jellyfish.levenshtein_distance's plain-unit-cost edit
// distance over Unicode code points: the minimum number of single-rune
// insertions, deletions, and substitutions turning a into b.
func Levenshtein(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}

	prev := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	curr := make([]int, len(rb)+1)
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}
