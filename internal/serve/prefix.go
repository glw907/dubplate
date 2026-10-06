package serve

import (
	"errors"
	"fmt"
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
// the owned value. [Shell.Register] builds each subtree's patterns
// from it on segment boundaries, so a neighbor sharing the prefix's
// bytes without the boundary is not owned.
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
