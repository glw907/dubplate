package serve

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"unicode"
)

// errInvalidPrefix is the sentinel [NewPrefix] wraps when a candidate
// prefix value fails construction validation.
var errInvalidPrefix = errors.New("invalid serve prefix")

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
// no character normalize would refuse in a request path: no
// empty, ".", or ".." segment, no backslash, and no percent-encoded
// byte. A failure wraps [errInvalidPrefix] rather than surfacing as a
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
		return fmt.Errorf("%q: must begin with a separator: %w", value, errInvalidPrefix)
	}
	// The check also refuses a bare "/", which owns nothing.
	if strings.HasSuffix(value, "/") {
		return fmt.Errorf("%q: must carry no trailing separator: %w", value, errInvalidPrefix)
	}
	// The shape checks run before the escaping check, which would
	// otherwise report each of them as a percent-encoded byte.
	if strings.Contains(value, `\`) {
		return fmt.Errorf("%q: must carry no backslash: %w", value, errInvalidPrefix)
	}
	if strings.ContainsFunc(value, unicode.IsSpace) {
		return fmt.Errorf("%q: must carry no whitespace: %w", value, errInvalidPrefix)
	}
	if strings.Contains(value, "%") {
		return fmt.Errorf("%q: must carry no percent-encoded byte: %w", value, errInvalidPrefix)
	}
	if escaped := (&url.URL{Path: value}).EscapedPath(); escaped != value {
		return fmt.Errorf("%q: must carry only bytes a request path admits unescaped: %w", value, errInvalidPrefix)
	}
	switch dot, doubled := segmentFault(value); {
	case dot:
		return fmt.Errorf("%q: must carry no %q or %q segment: %w", value, ".", "..", errInvalidPrefix)
	case doubled:
		return fmt.Errorf("%q: must carry no doubled separator: %w", value, errInvalidPrefix)
	}
	return nil
}

// String returns p's owned value, the form the shell uses to build
// patterns for the subtrees it registers under p.
func (p Prefix) String() string {
	return p.value
}
