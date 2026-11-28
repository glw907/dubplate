package serve

import (
	"net/http"
	"net/url"
	"strings"
)

// maxPathBytes bounds the request path rule 5 admits, inclusive and
// measured on the bytes the client sent: room for a deep catalog path
// under the prefix, while an oversized path never reaches the prefix
// match.
const maxPathBytes = 2048

// normalize returns a handler that refuses every request breaking one
// of six path rules and hands every other request to next untouched. It
// is a gate, never a rewriter: an admitted request reaches next with
// the path it arrived on, byte for byte, and no route or prefix
// decision sees a refused one.
//
// normalize wraps the mux and is never registered on it, since
// [http.ServeMux] cleans a path and redirects before any route runs.
//
// The six rules:
//
//  1. No percent-encoding: r.URL.EscapedPath() must equal r.URL.Path.
//     RawPath is never the basis, since net/url leaves it empty for a
//     path needing no escaping. The rule refuses %2F, %5C, and %2E%2E,
//     and a benign encoded byte such as /%E2%98%83 too, by policy: every
//     route is a literal path, and admitting some encodings means
//     choosing which at the one place a wrong choice is invisible.
//  2. No "." or ".." segment, refused rather than cleaned, since a
//     handler behind the shell may clean differently.
//  3. No doubled separator. A single trailing separator is admitted, so
//     "/dubplate/x/" passes and "/dubplate//x" is refused.
//  4. No backslash, which a handler's router may read as a separator.
//     net/url escapes one today, so rule 1 refuses it first; this rule
//     holds should that change.
//  5. No path past maxPathBytes.
//  6. No path that fails to begin with a separator. A CONNECT request
//     has an empty path, which every other rule passes vacuously. An
//     absolute-form target arrives with its path and routes normally.
//
// Rule 5 answers 414 and every other rule 400, so a client cannot tell
// which smuggling rule refused it.
func normalize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status, refused := pathRefusal(r.URL); refused {
			refuse(w, r, status)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// pathRefusal reports the status u is refused with, and whether it is
// refused at all. Rule 6 runs first, since an empty path passes every
// later rule, and the length bound before the content rules, so an
// oversized path gets its own status.
func pathRefusal(u *url.URL) (int, bool) {
	escaped := u.EscapedPath()

	if !strings.HasPrefix(u.Path, "/") {
		return http.StatusBadRequest, true
	}
	if len(escaped) > maxPathBytes {
		return http.StatusRequestURITooLong, true
	}
	if escaped != u.Path {
		return http.StatusBadRequest, true
	}
	if strings.Contains(u.Path, `\`) {
		return http.StatusBadRequest, true
	}

	if dot, doubled := segmentFault(u.Path); dot || doubled {
		return http.StatusBadRequest, true
	}
	return 0, false
}

// segmentFault reports the first segment fault in path, which begins
// with a separator: a "." or ".." segment, or a doubled separator.
func segmentFault(path string) (dot, doubled bool) {
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	for i, segment := range segments {
		switch segment {
		case ".", "..":
			return true, false
		case "":
			if i != len(segments)-1 {
				return false, true
			}
		}
	}
	return false, false
}
