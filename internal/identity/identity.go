// Package identity holds the verified identity a login proves and the
// channel it arrived on. It imports no other internal package and no
// third-party package: [internal/ops] carries a verified identity at
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
)

// MaxAddressLength is the byte length [NormalizeAddress] refuses to
// exceed: 254, the longest a reverse-path address may be end to end
// under RFC 5321 section 4.5.3.1.3, and far beyond anything a login
// page is meant to collect.
const MaxAddressLength = 254

// ErrInvalidAddress reports that a value is not a normalizable login
// address, either because [net/mail] cannot parse it as a single RFC
// 5322 address or because it exceeds [MaxAddressLength].
var ErrInvalidAddress = errors.New("identity: invalid address")

// NormalizeAddress parses value as a single RFC 5322 address and
// returns the address portion lowercased. Every caller that keys,
// logs, or rate-limits on a login address normalizes it here first,
// so the address inside a token's MAC and the address in a lookup
// never diverge.
func NormalizeAddress(value string) (string, error) {
	if len(value) > MaxAddressLength {
		return "", fmt.Errorf("%w: exceeds %d bytes", ErrInvalidAddress, MaxAddressLength)
	}
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

// Identity is a verified login. Address is the normalized address a
// member authenticated with; Principal is what names that member to a
// downstream consumer and is kept separate from Address so rung 6 can
// substitute a durable surrogate id there without touching a call
// site that reads either field. Source names the identity source that
// asserted the identity, and Channel is how it arrived. The zero
// value is not a valid Identity; call [Identity.Valid] before
// trusting one.
type Identity struct {
	Address     string
	Principal   string
	DisplayName string
	Source      string
	Channel     Channel
}

// Valid reports whether id carries a normalized address, a
// principal, a source, and a recognized channel. The zero value
// fails on every one of these.
func (id Identity) Valid() bool {
	return id.Address != "" && id.Principal != "" && id.Source != "" && id.Channel.Valid()
}
