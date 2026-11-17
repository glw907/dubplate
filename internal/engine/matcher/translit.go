package matcher

import (
	"strings"
	"sync/atomic"
)

// transliterator holds the process-wide function transliterate
// consults for every rune. It starts nil: any program that scores must
// install the pinned Unidecode table's lookup through SetTransliterator
// first, and a call before that panics, the way crypto.Hash.New panics
// for a hash no program registered.
var transliterator atomic.Pointer[func(rune) (string, bool)]

// SetTransliterator installs fn as the transliterator every
// transliterate call consults from this point on. Any program that
// scores must call it once, before scoring; no present command does. It panics on a nil fn: there is no channel to
// uninstall a transliterator once one is installed.
func SetTransliterator(fn func(rune) (string, bool)) {
	if fn == nil {
		panic("matcher: SetTransliterator called with a nil function")
	}
	transliterator.Store(&fn)
}

// transliterate reproduces `unidecode(s)`'s default errors="ignore"
// behavior (autotag/distance.py:9,55-56): every rune maps to its
// ASCII approximation, and a rune with no mapping contributes nothing.
// It panics, naming SetTransliterator, if nothing has been installed:
// an uninstalled seam is a wiring defect in a program, never a
// runtime condition to degrade gracefully from.
func transliterate(s string) string {
	fn := transliterator.Load()
	if fn == nil {
		panic("matcher: transliterate called before SetTransliterator installed a transliterator")
	}
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		repl, _ := (*fn)(r)
		b.WriteString(repl)
	}
	return b.String()
}
