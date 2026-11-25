// Package identity holds the verified identity a login proves and the
// channel it arrived on.
package identity

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/glw907/dubplate/internal/id"
)

// maxAddressLength is the byte length [NormalizeAddress] refuses to
// exceed: 254, the longest a reverse-path address may be end to end
// under RFC 5321 section 4.5.3.1.3, and far beyond anything a login
// page is meant to collect.
const maxAddressLength = 254

// errInvalidAddress reports that a value is not a normalizable login
// address: [net/mail] cannot parse it as a single RFC 5322 address, it
// exceeds [maxAddressLength], or its normalized form is not itself an
// address [NormalizeAddress] would return unchanged.
var errInvalidAddress = errors.New("invalid address")

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
// result also carries no rune [IsInvisibleRune] reports: a quoted
// local part can carry a byte-order mark, a bidi override, or an
// invisible filler that renders differently from how it compares.
//
// A returned error names the rule value broke, a refused rune by its
// code point, and never quotes any part of value, so a caller may log
// it.
func NormalizeAddress(value string) (string, error) {
	if len(value) > maxAddressLength {
		return "", fmt.Errorf("%w: exceeds %d bytes", errInvalidAddress, maxAddressLength)
	}
	normalized, err := parseLower(value)
	if err != nil {
		return "", err
	}
	if i := strings.IndexFunc(normalized, IsInvisibleRune); i >= 0 {
		r, _ := utf8.DecodeRuneInString(normalized[i:])
		return "", fmt.Errorf("%w: carries a control character or invisible character (%U)", errInvalidAddress, r)
	}
	if again, err := parseLower(normalized); err != nil || again != normalized {
		return "", fmt.Errorf("%w: normalized form does not parse to itself", errInvalidAddress)
	}
	return normalized, nil
}

// IsInvisibleRune reports whether r renders as nothing, or as something
// other than how it compares: every control character, every rune in
// the Unicode Cf (format), Zs (space separator), Zl (line separator),
// or Zp (paragraph separator) categories, every rune in the derived
// Default_Ignorable_Code_Point set, and U+2800 (braille pattern blank).
// The ordinary space (U+0020, in Zs) is among them.
func IsInvisibleRune(r rune) bool {
	return unicode.IsControl(r) || unicode.In(r, unicode.Cf, unicode.Zs, unicode.Zl, unicode.Zp) ||
		isDefaultIgnorableCodePoint(r) || r == brailleBlankRune
}

// brailleBlankRune is U+2800, braille pattern blank: category So (a
// visible symbol), so it is not in the derived Default_Ignorable_Code_Point
// set, but a cell rendered as no dots is indistinguishable from
// whitespace, so an address refuses it by name.
const brailleBlankRune = '⠀'

// isDefaultIgnorableCodePoint reports whether r is in Unicode's derived
// Default_Ignorable_Code_Point set, computed from the formula
// DerivedCoreProperties.txt (Unicode 17.0.0) states: "Generated from
// Other_Default_Ignorable_Code_Point + Cf (Format characters) +
// Variation_Selector - White_Space - FFF9..FFFB (Interlinear annotation
// format characters) - 13430..13440 (Egyptian hieroglyph format
// characters) - Prepended_Concatenation_Mark".
func isDefaultIgnorableCodePoint(r rune) bool {
	if !unicode.Is(unicode.Other_Default_Ignorable_Code_Point, r) &&
		!unicode.Is(unicode.Cf, r) &&
		!unicode.Is(unicode.Variation_Selector, r) {
		return false
	}
	if unicode.Is(unicode.White_Space, r) {
		return false
	}
	if r >= 0xFFF9 && r <= 0xFFFB {
		return false
	}
	if r >= 0x13430 && r <= 0x13440 {
		return false
	}
	if unicode.Is(unicode.Prepended_Concatenation_Mark, r) {
		return false
	}
	return true
}

// parseLower returns the address value parses to, lowercased. The
// parse error is dropped rather than wrapped: [net/mail] quotes its
// input, or the unparsed remainder of it, in several of its errors.
func parseLower(value string) (string, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return "", fmt.Errorf("%w: does not parse as a single address", errInvalidAddress)
	}
	return strings.ToLower(addr.Address), nil
}

// Channel names how the caller acting on a change reached the server,
// for an audit record.
type Channel int

const (
	// ChannelCookie identifies an identity asserted by a signed
	// session cookie on a browser request.
	ChannelCookie Channel = iota + 1
	// ChannelCommandLine identifies a change made by an
	// operating-system user through the server's own command line.
	ChannelCommandLine
)

// channelNames holds every recognized channel and the name an audit
// record or a log line carries for it.
var channelNames = map[Channel]string{
	ChannelCookie:      "cookie",
	ChannelCommandLine: "command_line",
}

// String returns the name an audit record or a log line carries for
// c. An unrecognized value reports "unknown" rather than a number.
func (c Channel) String() string {
	if name, ok := channelNames[c]; ok {
		return name
	}
	return "unknown"
}

// Valid reports whether c is a recognized channel. The zero value is
// not one.
func (c Channel) Valid() bool {
	_, ok := channelNames[c]
	return ok
}

// A UserID is a member's durable surrogate id: the 16-byte key the
// store mints once when the member is first added, which outlives a
// change of address or username. Per-member records key on it.
type UserID id.ID

// String returns u in [id.ID.String]'s form.
func (u UserID) String() string { return id.ID(u).String() }

// IsZero reports whether u is the zero UserID, which names no member.
func (u UserID) IsZero() bool { return id.ID(u).IsZero() }

// Identity is a verified login. The zero value is not a valid
// Identity; call [Identity.Valid] before trusting one.
type Identity struct {
	// Address is the login address, normalized by [NormalizeAddress].
	Address string
	// UserID is the member's durable id as the store holds it. The
	// zero value means the identity was never resolved against the
	// store.
	UserID UserID
	// Username is the readable name a log line or a music app knows the
	// member by. It is never the durable key; UserID is, so an audit
	// actor carries UserID and a renamed member keeps one actor.
	Username string
	// Channel is how this identity was asserted.
	Channel Channel
}

// Valid reports whether i is a complete verified login. An identity on
// [ChannelCommandLine] never is.
func (i Identity) Valid() bool {
	return i.Address != "" && !i.UserID.IsZero() && i.Username != "" &&
		i.Channel.Valid() && i.Channel != ChannelCommandLine
}
