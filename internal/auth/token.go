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

const nonceLength = 16

const (
	// linkVersion labels the field layout, on the wire and inside the
	// MAC. A new layout takes a new label, so an older token is never
	// read with the new layout's meanings.
	linkVersion = "dubplate-link-1"

	// linkPurpose separates the link's MAC from the session's, so a
	// session value handed to the link verifier fails the MAC under a
	// shared secret.
	linkPurpose = "dubplate/magic-link"

	linkFieldCount = 5
)

const (
	// linkPath is below whatever path the public origin carries.
	linkPath = "/login/link"

	// linkTokenParam puts the token in the query, not the path: the
	// consumption route logs the path and never RawQuery, and a path
	// token would land in every access log the request passes.
	linkTokenParam = "t"
)

// The link refusals are distinct so the consumption handler can log
// which check failed while showing one generic outcome.
var (
	errLinkMalformed = errors.New("malformed magic link")

	errLinkMAC = errors.New("bad magic link signature")

	errLinkExpired = errors.New("magic link expired")
)

// LinkToken is a magic link's fields. [VerifyLink] returns a non-zero
// LinkToken only for a value that passed the MAC and the expiry.
type LinkToken struct {
	// Address is normalized by [identity.NormalizeAddress].
	Address string

	// MintedAt is the mint time in whole seconds. The consumption
	// handler compares it with the revocation epoch, so an outstanding
	// link cannot outlive an offboarding.
	MintedAt time.Time

	ExpiresAt time.Time

	// Nonce is the single-use value the caller inserts into the
	// consumed-nonce store after verification.
	Nonce [nonceLength]byte
}

// MintLink signs a magic link for address, valid for expiry after now,
// and returns the encoded value with the fields it carries. The address
// is normalized here, so the MAC covers the value later lookups key on.
//
// secret is the current magic-link secret; the previous one only
// verifies. There is no destination parameter: one would ride inside
// the MAC under a new linkVersion, never as a free query parameter.
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

// VerifyLink verifies value under current and, when previous is
// non-empty, under previous, and returns the token's fields.
//
// The MAC is checked before the expiry, so no unauthenticated field is
// acted on, and the nonce is returned only after both, so an
// unauthenticated caller cannot fill the consumed-nonce store.
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

// LinkURL builds the link a member receives. publicOrigin is the
// configured public URL, never a request header: a link built from a
// request's Host would mail a token to an attacker's origin. prefix is
// the owned prefix's String form; an empty or "/" prefix is refused,
// since the link would land outside the shell's surface.
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

// linkFields is the field order shared by the wire encoding and the MAC
// input, so the two cannot disagree about what was signed.
func linkFields(tok LinkToken) [][]byte {
	return [][]byte{
		[]byte(linkVersion),
		[]byte(tok.Address),
		[]byte(strconv.FormatInt(tok.ExpiresAt.Unix(), 10)),
		[]byte(strconv.FormatInt(tok.MintedAt.Unix(), 10)),
		tok.Nonce[:],
	}
}

// parseLink decodes value without authenticating it. It returns the raw
// fields so the MAC is verified over the bytes that arrived.
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
