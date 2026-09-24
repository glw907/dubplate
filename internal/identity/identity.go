// Package identity holds the verified identity a login proves and the
// channel it arrived on. It imports no other internal package and no
// third-party package: internal/ops carries a verified identity at
// rung 6 without pulling in the server package that verified it, and
// the address normalizer here is the one place every other package
// keys, logs, or rate-limits a login address the same way. Nothing in
// this package may gain a dependency on config, auth, or any other
// internal package; a rung that needs one belongs elsewhere.
package identity

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode"
)

// MaxAddressLength is the byte length [NormalizeAddress] refuses to
// exceed: 254, the longest a reverse-path address may be end to end
// under RFC 5321 section 4.5.3.1.3, and far beyond anything a login
// page is meant to collect.
const MaxAddressLength = 254

// ErrInvalidAddress reports that a value is not a normalizable login
// address: [net/mail] cannot parse it as a single RFC 5322 address, it
// exceeds [MaxAddressLength], or its normalized form is not itself an
// address [NormalizeAddress] would return unchanged.
var ErrInvalidAddress = errors.New("identity: invalid address")

// NormalizeAddress parses value as a single RFC 5322 address and
// returns the address portion lowercased. Every caller that keys,
// logs, or rate-limits on a login address normalizes it here first,
// so the address inside a token's MAC and the address in a lookup
// never diverge.
//
// The result is a fixed point: normalizing it again returns it
// unchanged. [net/mail] unquotes a quoted local part, so
// "a,b"@example.com parses to a,b@example.com, which it then refuses;
// such an address is refused here, before any credential can name it,
// rather than accepted once and refused by every later step. The
// result also carries no control character and no rune in the Unicode
// format, space-separator, line-separator, or paragraph-separator
// categories: a quoted local part can carry a byte-order mark or a
// bidi override that renders differently from how it compares, and
// every such rune is refused rather than enumerated one at a time.
func NormalizeAddress(value string) (string, error) {
	if len(value) > MaxAddressLength {
		return "", fmt.Errorf("%w: exceeds %d bytes", ErrInvalidAddress, MaxAddressLength)
	}
	normalized, err := parseLower(value)
	if err != nil {
		return "", err
	}
	if strings.ContainsFunc(normalized, addressRefusesRune) {
		return "", fmt.Errorf("%w: carries a control character or invisible character", ErrInvalidAddress)
	}
	if again, err := parseLower(normalized); err != nil || again != normalized {
		return "", fmt.Errorf("%w: normalized form does not parse to itself", ErrInvalidAddress)
	}
	return normalized, nil
}

// addressRefusesRune reports whether an address refuses r: every
// control character plus every rune in the Unicode Cf (format), Zs
// (space separator), Zl (line separator), or Zp (paragraph separator)
// categories. The ordinary space (U+0020, in Zs) is already refused by
// this rule, so it changes nothing that the byte-range check it
// replaces did not already refuse.
func addressRefusesRune(r rune) bool {
	return unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zs, unicode.Zl, unicode.Zp)
}

func parseLower(value string) (string, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidAddress, err)
	}
	return strings.ToLower(addr.Address), nil
}

// Channel identifies how a verified identity was asserted. Admin-only
// verbs reject a cookie-channel identity, so the channel must travel
// with the identity rather than be inferred later from the request.
type Channel int

const (
	// ChannelCookie identifies an identity asserted by a signed
	// session cookie on a browser request.
	ChannelCookie Channel = iota + 1
	// ChannelHeader identifies an identity asserted by a header a
	// reverse-proxy sets on a forward-auth front.
	ChannelHeader
	// ChannelServiceToken identifies an identity asserted by a
	// service token presented directly, without a browser session.
	ChannelServiceToken
)

// String returns the name an audit record or a log line carries for
// c. An unrecognized value reports "unknown" rather than a number.
func (c Channel) String() string {
	switch c {
	case ChannelCookie:
		return "cookie"
	case ChannelHeader:
		return "header"
	case ChannelServiceToken:
		return "service-token"
	default:
		return "unknown"
	}
}

// Valid reports whether c names one of the three assertion channels.
// The zero value is not among them.
func (c Channel) Valid() bool {
	switch c {
	case ChannelCookie, ChannelHeader, ChannelServiceToken:
		return true
	default:
		return false
	}
}

// Identity is a verified login. The zero value is not a valid
// Identity; call [Identity.Valid] before trusting one.
type Identity struct {
	// Address is the login address, normalized by [NormalizeAddress].
	Address string
	// Principal is what names this member to a downstream consumer,
	// kept separate from Address so rung 6 can substitute a durable
	// surrogate id here without touching a call site that reads
	// either field.
	Principal string
	// DisplayName is the member's human-readable name.
	DisplayName string
	// Source names the identity source that asserted this identity,
	// for example "magic-link".
	Source string
	// Channel is how this identity was asserted.
	Channel Channel
}

// Valid reports whether id carries a normalized address, a
// principal, a source, and a recognized channel. The zero value
// fails on every one of these.
func (id Identity) Valid() bool {
	return id.Address != "" && id.Principal != "" && id.Source != "" && id.Channel.Valid()
}
