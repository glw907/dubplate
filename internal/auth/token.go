package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// nonceLength is the byte length of a magic link's nonce: 16 bytes of
// crypto/rand output, the 128 bits the single-use record is named by.
const nonceLength = 16

const (
	// linkVersion labels the token's field layout, on the wire and
	// inside the MAC. A later field lands under a new label rather than changing what an
	// existing field means, so a token minted before the change can
	// never be read with the new layout's meanings.
	linkVersion = "dubplate-link-1"

	// linkPurpose is the magic link's domain-separation label. The
	// session cookie signs under its own, so a session value handed
	// to the link verifier fails the MAC instead of being read as the
	// wrong kind of credential under a shared secret.
	linkPurpose = "dubplate/magic-link"

	// linkFieldCount is how many fields precede the MAC on the wire:
	// version, address, expiry, mint time, nonce.
	linkFieldCount = 5
)

const (
	// linkPath is the path a magic link points at, below whatever
	// path the configured public origin carries.
	linkPath = "/login/link"

	// linkTokenParam carries the token in the link's query string
	// rather than in a path segment: the consumption route logs the
	// path and never RawQuery, and a path-borne token would be
	// written by every access log the request passes through.
	linkTokenParam = "t"
)

// The link refusals are distinct so the consumption handler can log
// which check failed while showing one generic outcome.
var (
	// errLinkMalformed reports a value that is not a link token at
	// all: the wrong field count, a field that is not base64url, an
	// unknown version label, or a field whose contents do not parse.
	errLinkMalformed = errors.New("malformed magic link")

	// errLinkMAC reports a link token whose MAC verifies under
	// neither the current nor the previous secret.
	errLinkMAC = errors.New("bad magic link signature")

	// errLinkExpired reports a link token whose MAC verified and
	// whose expiry has passed.
	errLinkExpired = errors.New("magic link expired")
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
	Nonce [nonceLength]byte
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
		return LinkToken{}, errLinkMAC
	}
	if !now.Before(tok.ExpiresAt) {
		return LinkToken{}, errLinkExpired
	}
	return tok, nil
}

// LinkURL builds the link a member receives, from publicOrigin, the
// owned prefix, and the encoded token. publicOrigin is the configured
// public URL and never a request header: a link built from a Host a
// request supplied would mail an attacker's origin a token minted for
// someone else. prefix is the owned prefix's string form (serve.Prefix's
// String), so the link lands on a path the shell routes to its own
// handlers. An empty or "/" prefix is refused, since it would build a
// link at the bare origin, outside that surface.
func LinkURL(publicOrigin, prefix, token string) (string, error) {
	if prefix == "" || prefix == "/" {
		return "", fmt.Errorf("magic link prefix: empty or root, must name a subtree")
	}
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
	fields, mac, err := splitSigned(value, linkFieldCount, errLinkMalformed)
	if err != nil {
		return LinkToken{}, nil, nil, err
	}

	if string(fields[0]) != linkVersion {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: version label", errLinkMalformed)
	}
	expires, err := strconv.ParseInt(string(fields[2]), 10, 64)
	if err != nil {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: expiry", errLinkMalformed)
	}
	minted, err := strconv.ParseInt(string(fields[3]), 10, 64)
	if err != nil {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: mint time", errLinkMalformed)
	}
	if len(fields[4]) != nonceLength {
		return LinkToken{}, nil, nil, fmt.Errorf("%w: nonce length", errLinkMalformed)
	}

	tok := LinkToken{
		Address:   string(fields[1]),
		MintedAt:  time.Unix(minted, 0).UTC(),
		ExpiresAt: time.Unix(expires, 0).UTC(),
	}
	copy(tok.Nonce[:], fields[4])
	return tok, fields, mac, nil
}
