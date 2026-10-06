package serve

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode"
)

// ErrInvalidPrefix is the sentinel [NewPrefix] wraps when a candidate
// prefix value fails construction validation.
var ErrInvalidPrefix = errors.New("invalid serve prefix")

// Prefix is the one path prefix dubplate owns end to end: every route
// the shell registers lives under it, and every path outside it
// answers 404. A Prefix is built once, from [NewPrefix], and carries
// the owned value, [Prefix.Owns]'s segment-aware match, and
// [Prefix.Route]'s composition of that match with the terminator.
type Prefix struct {
	value string
}

// NewPrefix validates value and returns the [Prefix] it names. value
// must begin with a separator, carry no trailing separator, and carry
// no character Normalize's rules would refuse in a request path: no
// empty, ".", or ".." segment, no backslash, and no percent-encoded
// byte. A failure wraps [ErrInvalidPrefix] rather than surfacing as a
// runtime 500, since a misconfigured prefix is a construction-time
// mistake rather than a per-request one.
func NewPrefix(value string) (Prefix, error) {
	if err := validatePrefix(value); err != nil {
		return Prefix{}, err
	}
	return Prefix{value: value}, nil
}

func validatePrefix(value string) error {
	if !strings.HasPrefix(value, "/") {
		return fmt.Errorf("%q: must begin with a separator: %w", value, ErrInvalidPrefix)
	}
	// A bare "/" also satisfies HasSuffix(value, "/"), so this same
	// check refuses the root path as an owned prefix along with every
	// other trailing separator: the root owns nothing distinguishable
	// from the paths outside the prefix.
	if strings.HasSuffix(value, "/") {
		return fmt.Errorf("%q: must carry no trailing separator: %w", value, ErrInvalidPrefix)
	}
	// The three shape checks run before the escaping check below, and
	// their order is the whole point of their existing separately. A
	// bare URL{Path: value}.EscapedPath escapes a backslash, a space, a
	// tab, and a "!" alike, so the escaping check refuses every one of
	// them; put first, it would report "must carry no percent-encoded
	// byte" for a prefix carrying no percent sign at all, which sends an
	// adopter looking for an encoding they never wrote. Each shape
	// reports itself here, and the escaping check keeps only what is
	// left.
	if strings.Contains(value, `\`) {
		return fmt.Errorf("%q: must carry no backslash: %w", value, ErrInvalidPrefix)
	}
	if strings.ContainsFunc(value, unicode.IsSpace) {
		return fmt.Errorf("%q: must carry no whitespace: %w", value, ErrInvalidPrefix)
	}
	if strings.Contains(value, "%") {
		return fmt.Errorf("%q: must carry no percent-encoded byte: %w", value, ErrInvalidPrefix)
	}
	// A bare URL{Path: value} never decodes value, so its EscapedPath
	// is the default escaping of value's literal bytes; a mismatch that
	// survives the three checks above means value carries some other
	// byte a request path may not carry unescaped, and [Normalize]'s
	// raw-versus-decoded rule would refuse it there.
	if escaped := (&url.URL{Path: value}).EscapedPath(); escaped != value {
		return fmt.Errorf("%q: must carry only bytes a request path admits unescaped: %w", value, ErrInvalidPrefix)
	}
	// The leading separator opens the path rather than closing a
	// segment, so it is trimmed before the split. An empty segment
	// anywhere but last is a doubled separator; an empty segment last
	// is the trailing separator the check above already refused on
	// its own, so a mutation that drops that check surfaces here as a
	// double separator at match time (p.value+"/" against a path that
	// already ends p.value+"/") rather than a second refusal of the
	// same shape.
	segments := strings.Split(strings.TrimPrefix(value, "/"), "/")
	for i, segment := range segments {
		switch segment {
		case ".", "..":
			return fmt.Errorf("%q: must carry no %q or %q segment: %w", value, ".", "..", ErrInvalidPrefix)
		case "":
			if i != len(segments)-1 {
				return fmt.Errorf("%q: must carry no doubled separator: %w", value, ErrInvalidPrefix)
			}
		}
	}
	return nil
}

// String returns p's owned value, the form the shell uses to build
// patterns for the subtrees it registers under p.
func (p Prefix) String() string {
	return p.value
}

// Owns reports whether path enters the surface p owns: p's value with
// nothing following, or p's value followed by a separator. The
// comparison is case-sensitive bytes on segment boundaries, never a
// plain string-prefix and never [http.StripPrefix]. The salvage
// tree's own mountBasePath already proved the defect this guards
// against: http.StripPrefix does a plain trim, so an unguarded
// neighbor path such as basePath+"foo" trims to a value that reaches
// the wrapped handler instead of a 404
// (_salvage/internal/serve/serve.go:82-108, the comment documenting
// the gap it then guards by hand). This task refuses the gap at the
// match itself rather than documenting it a second time.
func (p Prefix) Owns(path string) bool {
	if path == p.value {
		return true
	}
	return strings.HasPrefix(path, p.value+"/")
}

// Terminator answers with 404 and nothing else. It is what a request
// under p's owned surface reaches when no registered subtree claims
// its path: the request stops here, so an unknown path under the owned
// prefix never reaches a subtree.
func Terminator() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
}

// Route returns the prefix-routing step: a request whose path p owns
// is handed to owned, and every other request is handed to next. The
// decision is p's alone: a path outside the prefix is never offered to
// owned, and a path inside it is never offered to next, so a later
// root route arrives as a named route the shell owns rather than a
// change to the owned surface. The shell passes the mux it registers
// subtrees on as owned, and [Terminator] answers a path p owns that no
// subtree claims.
//
// The bare prefix and the prefix followed by a separator are both
// handled here, by p.Owns, rather than left to [http.ServeMux]'s own
// redirect for a pattern ending in "/": Route never registers p's
// value as a mux pattern, so no such redirect exists to fire, and the
// decision belongs to the shell instead of the mux.
func (p Prefix) Route(owned, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if p.Owns(r.URL.Path) {
			owned.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}
