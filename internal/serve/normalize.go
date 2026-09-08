package serve

import (
	"net/http"
	"net/url"
	"strings"
)

// maxPathBytes bounds the request path rule 5 admits, measured on the
// bytes the client sent. It is a constant rather than a threshold with
// a default (ruled, conductor 2026-09-08, R-H, as applied): the rung 5
// tunable list does not name it, and a threshold nobody tunes is a
// configuration surface for nothing. The value leaves room for a deep
// catalog path under the owned prefix while keeping an oversized path
// away from the prefix match.
const maxPathBytes = 2048

// Normalize returns a handler that answers every request breaking one
// of six path rules and hands every other request to next untouched.
// It is a gate, never a rewriter: an admitted request reaches next
// with the path it arrived on, byte for byte.
//
// A refusal is answered here and nowhere else. No route, no prefix
// decision, and no fallthrough sees a refused request, which is what
// makes the rules a boundary rather than a filter.
//
// Normalize wraps the mux and is never registered as a route on one.
// http.ServeMux cleans a path and redirects before any route runs, so
// a check registered as a route would run after the decision it exists
// to guard. TestServeMuxCleansBeforeARouteRuns measures that redirect.
//
// The six rules:
//
//  1. No percent-encoding in the path. The predicate is
//     r.URL.EscapedPath() != r.URL.Path (ruled, conductor 2026-09-08,
//     R-D, as applied), and RawPath is never the comparison basis:
//     net/url populates RawPath only when the escaped form differs
//     from the default encoding of Path, so a path needing no escaping
//     carries an empty RawPath and a RawPath-based test would admit
//     the very shapes this rule exists to refuse. EscapedPath is
//     defined for every request, returning RawPath when it is a valid
//     encoding of Path and re-deriving the escaped form otherwise.
//     The rule closes %2F, %5C, and %2E%2E, and it refuses a benign
//     encoded byte such as /%E2%98%83 too. That blanket refusal is the
//     stated policy rather than a side effect: the shell owns no path
//     needing an encoded byte, a driver declares its paths as
//     literals, and a rule admitting some encodings has to decide
//     which at the one place in the shell where a wrong decision is
//     invisible. Rung 12 revisits the strictness with measured
//     evidence from a real driver.
//  2. No "." or ".." segment, refused rather than cleaned. Cleaning is
//     what http.ServeMux does, and a streamer behind the fallthrough
//     may normalize differently again.
//  3. No doubled separator anywhere. The leading separator is not a
//     segment and a single trailing separator is admitted, so "/",
//     "/dubplate/", and "/dubplate/x/" pass while "//dubplate" and
//     "/dubplate//x" are refused.
//  4. No backslash, which a streamer's own router may read as a
//     separator on the fallthrough.
//  5. No path past maxPathBytes.
//  6. No request target whose path fails to begin with a separator
//     (ruled, conductor 2026-09-08, R-U, as applied). A CONNECT
//     request arrives with an empty URL.Path and a populated URL.Host,
//     so every other rule passes over it vacuously. An absolute-form
//     target is a different shape and is admitted: it arrives with its
//     path intact and routes like any origin-form request. OPTIONS *
//     reaches no rule here, since net/http answers it from
//     globalOptionsHandler unless DisableGeneralOptionsHandler is set.
//
// Every rule but the length bound refuses with 400 (ruled, conductor
// 2026-09-08, R-T, as applied), so a client cannot tell which
// smuggling rule refused it. The length bound answers 414, a shape
// already distinguishable from outside by its own size.
func Normalize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status, refused := pathRefusal(r.URL); refused {
			http.Error(w, http.StatusText(status), status)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// pathRefusal reports the status u is refused with, and whether it is
// refused at all. Rule 6 runs first, since an empty path passes every
// later rule vacuously, and the length bound runs before the content
// rules so an oversized path is answered by its own status.
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

	// The leading separator opens the path rather than closing a
	// segment, so it is trimmed before the split. What remains is one
	// element per segment, and an empty element in any position but
	// the last is a doubled separator; an empty element in the last
	// position is the trailing separator rule 3 admits.
	segments := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	for i, segment := range segments {
		switch segment {
		case ".", "..":
			return http.StatusBadRequest, true
		case "":
			if i != len(segments)-1 {
				return http.StatusBadRequest, true
			}
		}
	}

	return 0, false
}
