package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// NonceLength is the byte length of a magic link's nonce: 16 bytes of
// crypto/rand output, the 128 bits the single-use record is named by.
const NonceLength = 16

const (
	// linkVersion labels the token's field layout, on the wire and
	// inside the MAC. A later field, rung 12's destination among
	// them, lands under a new label rather than changing what an
	// existing field means, so a token minted before the change can
	// never be read with the new layout's meanings.
	linkVersion = "dubplate-link-1"

	// linkPurpose is the magic link's domain-separation label. The
	// session cookie signs under its own, so a session value handed
	// to the link verifier fails the MAC instead of being read as the
	// wrong kind of credential under a shared secret.
	linkPurpose = "dubplate/magic-link"

	// fieldSeparator joins the encoded fields. Every field is
	// base64url-encoded before the join, so no field value can carry
	// the separator: a raw join would split a dotted local part,
	// which is the majority of real addresses, and the failure would
	// be invisible to any test whose fixture address has no dot.
	//
	// ruled, conductor 2026-09-12, A4 fold 3
	fieldSeparator = "."

	// linkFieldCount is how many fields precede the MAC on the wire:
	// version, address, expiry, mint time, nonce.
	linkFieldCount = 5
)

const (
	// linkPath is the path a magic link points at, below whatever
	// path the configured public origin carries.
	linkPath = "/login/link"

	// linkTokenParam carries the token in the link's query string
	// rather than in a path segment: R32's mitigation is a
	// consumption route that logs the path and never RawQuery, and a
	// path-borne token would be written by every access log the
	// request passes through.
	//
	// ruled, conductor 2026-09-08, R-5b13
	linkTokenParam = "t"
)

// Sentinels a caller branches on. The three refusals are distinct here
// so the consumption handler can log which check failed while showing
// one generic outcome, which is the handler's choice rather than this
// package's.
var (
	// ErrLinkMalformed reports a value that is not a link token at
	// all: the wrong field count, a field that is not base64url, an
	// unknown version label, or a field whose contents do not parse.
	ErrLinkMalformed = errors.New("auth: malformed magic link")

	// ErrLinkMAC reports a link token whose MAC verifies under
	// neither the current nor the previous secret.
	ErrLinkMAC = errors.New("auth: magic link signature")

	// ErrLinkExpired reports a link token whose MAC verified and
	// whose expiry has passed.
	ErrLinkExpired = errors.New("auth: magic link expired")
)

// LinkToken is a magic link's fields. VerifyLink returns one only for a
// value that passed both the MAC and the expiry, so a zero LinkToken
// beside a non-nil error is the whole of what a refused value yields.
type LinkToken struct {
	// Address is the login address, normalized by
	// identity.NormalizeAddress before the MAC covered it.
	Address string

	// MintedAt is when the link was minted, at second granularity.
	// The consumption handler compares it against the address's
	// revocation epoch, so an outstanding link cannot outlive an
	// offboarding.
	MintedAt time.Time

	// ExpiresAt is when the link stops verifying.
	ExpiresAt time.Time

	// Nonce is the link's single-use value, which the caller inserts
	// into the consumed-nonce store after verification returns it.
	Nonce [NonceLength]byte
}

// MintLink signs a magic link token for address, valid for expiry after
// now, and returns the encoded value alongside the fields it carries.
// The address is normalized here rather than assumed normalized, so the
// value a MAC covers is the value every later lookup keys on.
//
// secret is the current magic-link secret. There is no minting path for
// the previous secret, which is what makes rotation routine rather than
// an estate-wide sign-out, and no destination parameter: a later
// destination rides inside the MAC-covered token under a new
// linkVersion, never as a free query parameter a recipient can rewrite.
func MintLink(secret []byte, address string, now time.Time, expiry time.Duration) (string, LinkToken, error) {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return "", LinkToken{}, fmt.Errorf("mint magic link: %w", err)
	}
	tok := LinkToken{Address: normalized, MintedAt: now.UTC().Truncate(time.Second)}
	tok.ExpiresAt = tok.MintedAt.Add(expiry)
	// crypto/rand.Read never returns an error, which its own contract
	// states; the branch exists because a silently short nonce would
	// be a single-use record nobody could trust.
	if _, err := rand.Read(tok.Nonce[:]); err != nil {
		return "", LinkToken{}, fmt.Errorf("mint magic link nonce: %w", err)
	}
	return encodeLink(secret, tok), tok, nil
}

// VerifyLink verifies value against the current secret and, when
// previous is non-empty, against the previous one, and returns the
// token's fields. A token that verifies under previous alone is
// accepted, which is the whole of what rotation overlap buys.
//
// The order is fixed: the MAC first, then the expiry, and the nonce
// only after both. Reading the expiry first would act on an
// unauthenticated field, and returning the nonce earlier would let an
// unauthenticated caller fill the consumed-nonce store.
func VerifyLink(current, previous []byte, value string, now time.Time) (LinkToken, error) {
	tok, fields, mac, err := parseLink(value)
	if err != nil {
		return LinkToken{}, err
	}
	if !macMatches(current, linkPurpose, fields, mac) && !macMatches(previous, linkPurpose, fields, mac) {
		return LinkToken{}, ErrLinkMAC
	}
	if !now.Before(tok.ExpiresAt) {
		return LinkToken{}, ErrLinkExpired
	}
	return tok, nil
}

// LinkURL builds the link a member receives, from publicOrigin, the
// owned prefix, and the encoded token. publicOrigin is the configured
// public URL and never a request header: a link built from a Host a
// request supplied would mail an attacker's origin a token minted for
// someone else. prefix is the owned prefix's string form (serve.Prefix's
// String), so the link lands on a path the shell routes to its own
// handlers rather than to the fallthrough, whatever that slot holds.
func LinkURL(publicOrigin, prefix, token string) (string, error) {
	u, err := url.Parse(publicOrigin)
	if err != nil {
		return "", fmt.Errorf("magic link origin: %w", err)
	}
	u = u.JoinPath(prefix, linkPath)
	u.RawQuery = url.Values{linkTokenParam: {token}}.Encode()
	return u.String(), nil
}

func encodeLink(secret []byte, tok LinkToken) string {
	return encodeSigned(secret, linkPurpose, linkFields(tok))
}

// linkFields is the token's field order, shared by the wire encoding
// and the MAC input so the two can never disagree about what was
// signed.
func linkFields(tok LinkToken) [][]byte {
	return [][]byte{
		[]byte(linkVersion),
		[]byte(tok.Address),
		[]byte(strconv.FormatInt(tok.ExpiresAt.Unix(), 10)),
		[]byte(strconv.FormatInt(tok.MintedAt.Unix(), 10)),
		tok.Nonce[:],
	}
}

// parseLink decodes value into its fields without authenticating any of
// them. It returns the fields as well as the token so the caller can
// verify the MAC over exactly the bytes that arrived, rather than over
// a re-serialization that could differ from them.
func parseLink(value string) (LinkToken, [][]byte, []byte, error) {
	fields, mac, err := splitSigned(value, linkFieldCount, ErrLinkMalformed)
	if err != nil {
		return LinkToken{}, nil, nil, err
	}

	if string(fields[0]) != linkVersion {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: version label", ErrLinkMalformed)
	}
	expires, err := strconv.ParseInt(string(fields[2]), 10, 64)
	if err != nil {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: expiry", ErrLinkMalformed)
	}
	minted, err := strconv.ParseInt(string(fields[3]), 10, 64)
	if err != nil {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: mint time", ErrLinkMalformed)
	}
	if len(fields[4]) != NonceLength {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: nonce length", ErrLinkMalformed)
	}

	tok := LinkToken{
		Address:   string(fields[1]),
		MintedAt:  time.Unix(minted, 0).UTC(),
		ExpiresAt: time.Unix(expires, 0).UTC(),
	}
	copy(tok.Nonce[:], fields[4])
	return tok, fields, mac, nil
}

// Both credentials are the same shape on the wire, a run of
// base64url-encoded fields joined by fieldSeparator with the MAC last,
// and differ only in their field list and their domain-separation
// label. The three functions below carry that shape for both, so the
// encoder, the decoder, and the comparison can never drift apart
// between the two.

// fieldEncoding is unpadded base64url that refuses non-zero trailing
// bits in a field's final character. A lax decoder ignores those bits,
// so a field whose byte length is not a multiple of three has several
// encodings that decode alike, and every one of them would verify.
// Refusing them is the canonical encoding RFC 4648 section 3.5
// describes (read 2026-09-23). A strict decoder still skips CR and LF,
// so splitSigned refuses those itself, and between the two each
// credential has exactly one string.
var fieldEncoding = base64.RawURLEncoding.Strict()

// encodeSigned renders fields and the MAC computed over them under
// purpose.
func encodeSigned(secret []byte, purpose string, fields [][]byte) string {
	parts := make([]string, 0, len(fields)+1)
	for _, field := range fields {
		parts = append(parts, fieldEncoding.EncodeToString(field))
	}
	parts = append(parts, fieldEncoding.EncodeToString(sign(secret, purpose, fields...)))
	return strings.Join(parts, fieldSeparator)
}

// splitSigned decodes value into count MAC-covered fields and the MAC
// that follows them, authenticating none of them. malformed is the
// caller's sentinel, so a refusal names the credential that was
// malformed rather than a shared one.
//
// The split stops after count+2 parts, one more than a well-formed
// value holds, so the field-count check stays exact while a value of
// nothing but separators costs a fixed handful of substrings rather
// than one per byte.
func splitSigned(value string, count int, malformed error) ([][]byte, []byte, error) {
	if strings.ContainsAny(value, "\r\n") {
		return nil, nil, fmt.Errorf("%w: field encoding", malformed)
	}
	parts := strings.SplitN(value, fieldSeparator, count+2)
	if len(parts) != count+1 {
		return nil, nil, fmt.Errorf("%w: field count", malformed)
	}
	decoded := make([][]byte, 0, len(parts))
	for _, part := range parts {
		raw, err := fieldEncoding.DecodeString(part)
		if err != nil {
			return nil, nil, fmt.Errorf("%w: field encoding", malformed)
		}
		decoded = append(decoded, raw)
	}
	return decoded[:count], decoded[count], nil
}

// macMatches reports whether mac is the MAC of fields under secret and
// purpose. An empty secret matches nothing, which is how an
// unconfigured rotation overlap is absent rather than permissive.
//
// The comparison is crypto/subtle.ConstantTimeCompare rather than
// bytes.Equal because bytes.Equal returns at the first differing byte:
// its timing tells an attacker who can present attempts how long a
// forged prefix matched, which turns forgery into a byte-at-a-time
// search. No test in this suite can observe the difference; the call
// itself is the guarantee.
func macMatches(secret []byte, purpose string, fields [][]byte, mac []byte) bool {
	if len(secret) == 0 {
		return false
	}
	return subtle.ConstantTimeCompare(sign(secret, purpose, fields...), mac) == 1
}

// sign returns the HMAC-SHA256 of purpose and fields under secret.
func sign(secret []byte, purpose string, fields ...[]byte) []byte {
	mac := hmac.New(sha256.New, secret)
	// hash.Hash's Write never returns an error, which its contract
	// states, so there is nothing here for a caller to handle.
	_, _ = mac.Write(macInput(purpose, fields))
	return mac.Sum(nil)
}

// macInput length-prefixes the domain-separation label and every field.
// Without the prefixes, two different field sets whose values
// concatenate to the same bytes share a MAC input, so one token's
// signature authenticates the other's fields: a character moved from
// the end of an address onto the front of an expiry would carry its
// signature with it.
func macInput(purpose string, fields [][]byte) []byte {
	buf := appendField(nil, []byte(purpose))
	for _, field := range fields {
		buf = appendField(buf, field)
	}
	return buf
}

func appendField(dst, field []byte) []byte {
	dst = binary.BigEndian.AppendUint64(dst, uint64(len(field)))
	return append(dst, field...)
}
