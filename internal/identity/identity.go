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
// result also carries no control character; no rune in the Unicode
// format, space-separator, line-separator, or paragraph-separator
// categories; no rune in the derived Default_Ignorable_Code_Point set
// ([isDefaultIgnorableCodePoint]); and no U+2800 (braille pattern
// blank): a quoted local part can carry a byte-order mark, a bidi
// override, or an invisible filler that renders differently from how
// it compares, and every such rune is refused rather than enumerated
// one at a time.
func NormalizeAddress(value string) (string, error) {
	if len(value) > MaxAddressLength {
		return "", fmt.Errorf("%w: exceeds %d bytes", ErrInvalidAddress, MaxAddressLength)
	}
	normalized, err := parseLower(value)
	if err != nil {
		return "", err
	}
	if strings.ContainsFunc(normalized, addressRefusesRune) {
		return "", fmt.Errorf("%w: carries a control character or invisible character (%s)", ErrInvalidAddress, invisibleRune(normalized))
	}
	if again, err := parseLower(normalized); err != nil || again != normalized {
		return "", fmt.Errorf("%w: normalized form does not parse to itself", ErrInvalidAddress)
	}
	return normalized, nil
}

// addressRefusesRune reports whether an address refuses r: every
// control character, every rune in the Unicode Cf (format), Zs (space
// separator), Zl (line separator), or Zp (paragraph separator)
// categories, every rune in the derived Default_Ignorable_Code_Point
// set ([isDefaultIgnorableCodePoint]), and U+2800 (braille pattern
// blank). The ordinary space (U+0020, in Zs) is already refused by
// this rule, so it changes nothing that the byte-range check it
// replaces did not already refuse.
func addressRefusesRune(r rune) bool {
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
// characters) - Prepended_Concatenation_Mark". The standard library
// exposes every input the formula names (identity_test.go pins the
// derived set's size against the file's own "Total code points: 4174"
// and unicode.Version, so a toolchain move to a later Unicode version
// fails loudly rather than silently widening or narrowing the set).
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

// invisibleRune returns the first rune in value that addressRefusesRune
// flags, formatted as its code point ("U+XXXX"), so a refusal can name
// what it refused without quoting value itself. It returns "" when no
// rune in value is refused, which never happens on the caller's own
// path: NormalizeAddress calls it only after confirming
// strings.ContainsFunc(value, addressRefusesRune).
func invisibleRune(value string) string {
	for _, r := range value {
		if addressRefusesRune(r) {
			return fmt.Sprintf("U+%04X", r)
		}
	}
	return ""
}

func parseLower(value string) (string, error) {
	addr, err := mail.ParseAddress(value)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidAddress, err)
	}
	return strings.ToLower(addr.Address), nil
}

// Channel identifies how an identity reached the code acting on it,
// and travels with the identity rather than being inferred later from
// the request.
//
// [ChannelCookie] is an assertion: a verifier emits it for a login a
// signed session cookie proves. [ChannelCommandLine] is not. It names
// an operating-system user, established by owning the process that
// invokes the command, and it exists so an audit record can say where
// a change came from. No verifier emits it, and [Identity.Valid]
// refuses an identity that carries it.
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
	ChannelCommandLine: "command-line",
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

// Valid reports whether id carries a non-empty address, a non-empty
// principal, a non-empty source, and a recognized channel other than
// [ChannelCommandLine], which names an operating-system user and never
// a verified login. The zero value fails on every one of these.
func (id Identity) Valid() bool {
	return id.Address != "" && id.Principal != "" && id.Source != "" &&
		id.Channel.Valid() && id.Channel != ChannelCommandLine
}
