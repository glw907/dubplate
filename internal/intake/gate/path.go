// Package gate judges a submission's files: the manifest a member declares,
// the archive it arrives in, and each audio file inside. It is pure Go over
// the standard library and tagio, and imports nothing from the store, so
// the intake cross-compile check can build it for every platform.
package gate

import (
	"path"
	"strings"
)

// ValidPath reports whether p is a relative path the library may store: no
// leading slash, no empty, ".", or ".." component, no NUL, no backslash,
// and equal to its own [path.Clean] form. It mirrors the relative-path
// CHECK on the store's path columns and adds the backslash, which SQL
// admits. p is bytes, not a string, because a stored path need not be valid
// UTF-8.
func ValidPath(p []byte) bool {
	s := string(p)
	if s == "" || s[0] == '/' || strings.ContainsAny(s, "\x00\\") || path.Clean(s) != s {
		return false
	}
	// Clean leaves a lone "." or a leading ".." in place.
	for part := range strings.SplitSeq(s, "/") {
		if part == "." || part == ".." {
			return false
		}
	}
	return true
}
