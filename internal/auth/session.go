package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// The browser, not this server, enforces the __Host- prefix: it stores
// the cookie only when it is Secure, has Path=/, and carries no Domain,
// which confines it to the exact origin that set it.
const sessionCookieName = "__Host-dubplate_session"

const (
	// sessionVersion labels the field layout, on the wire and inside the
	// MAC. A new layout takes a new label, so an older cookie is never
	// read with the new layout's meanings.
	sessionVersion = "dubplate-session-1"

	// sessionPurpose separates the session's MAC from the magic link's.
	// Both may be signed under one secret and share a wire shape, so this
	// label alone stops a link token being presented as a session.
	sessionPurpose = "dubplate/session"

	sessionFieldCount = 5
)

// The session refusals are distinct so the gate can log which check
// failed while showing one generic outcome.
var (
	errSessionMalformed = errors.New("malformed session")

	errSessionMAC = errors.New("bad session signature")

	// errSessionCapReached is reported even while the value's own expiry
	// is in the future, the case a rolling refresh produces.
	errSessionCapReached = errors.New("session absolute cap reached")

	errSessionExpired = errors.New("session expired")
)

// Session is a session cookie's fields. [VerifySession] returns a
// non-zero Session only for a value that passed every check.
type Session struct {
	// Address is normalized by [identity.NormalizeAddress].
	Address string

	// IssuedAt is the last issue time, in whole seconds. A refresh moves it.
	IssuedAt time.Time

	// AnchoredAt is the original issue time, which a refresh keeps. The
	// absolute cap and the revocation-epoch comparison are measured from
	// it, never from IssuedAt.
	AnchoredAt time.Time

	// ExpiresAt is when the session stops verifying absent a refresh.
	ExpiresAt time.Time
}

// MintSession signs a session for address, valid for lifetime after now,
// and returns the encoded value with the fields it carries. The address
// is normalized here, so the MAC covers the value later lookups key on.
// The expiry is clamped to absoluteCap past the anchor, so the cookie's
// Max-Age never promises longer than [VerifySession] honors.
//
// secret is the current session secret; the previous one only verifies.
func MintSession(secret []byte, address string, now time.Time, lifetime, absoluteCap time.Duration) (string, Session, error) {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return "", Session{}, fmt.Errorf("mint session: %w", err)
	}
	s := Session{Address: normalized, IssuedAt: now.UTC().Truncate(time.Second)}
	s.AnchoredAt = s.IssuedAt
	s.clampExpiry(lifetime, absoluteCap)
	return encodeSession(secret, s), s, nil
}

// refreshSession re-signs s with its issue time moved to now and its
// expiry clamped as in [MintSession]. The anchor carries across, so no
// number of refreshes outlives the absolute cap.
//
// Callers refresh on same-origin requests only: a cross-origin
// subresource load is the browser's traffic, not the person's, and
// would hold an abandoned session open to its cap.
func refreshSession(secret []byte, s Session, now time.Time, lifetime, absoluteCap time.Duration) (string, Session) {
	s.IssuedAt = now.UTC().Truncate(time.Second)
	s.clampExpiry(lifetime, absoluteCap)
	return encodeSession(secret, s), s
}

func (s *Session) clampExpiry(lifetime, absoluteCap time.Duration) {
	expiresAt := s.IssuedAt.Add(lifetime)
	if capAt := s.AnchoredAt.Add(absoluteCap); expiresAt.After(capAt) {
		expiresAt = capAt
	}
	s.ExpiresAt = expiresAt
}

// sessionCookie builds the cookie carrying value. Secure, Path=/, and no
// Domain are what the __Host- prefix requires. SameSite is Lax so a
// magic link followed from a mail client arrives with the session.
// MaxAge is the primary control because it does not depend on the
// client's clock; Expires serves a client that honors only the older
// attribute. The server enforces the value's signed expiry either way.
func sessionCookie(value string, now, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(expiresAt.Sub(now).Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// clearSessionCookie returns a cookie that deletes the session cookie,
// repeating the __Host- attributes so the browser matches it.
//
// Sessions are stateless: a value captured before sign-out keeps
// verifying until its expiry or an epoch bump, and sign-out copy must
// not claim more.
func clearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// VerifySession verifies value under current and, when previous is
// non-empty, under previous, and returns the session's fields.
//
// The MAC is checked before either time, so no unauthenticated field is
// acted on. The cap precedes the expiry so a session past both reports
// the refusal a refresh cannot cure.
func VerifySession(current, previous []byte, value string, now time.Time, absoluteCap time.Duration) (Session, error) {
	s, err := authenticateSession(current, previous, value)
	if err != nil {
		return Session{}, err
	}
	if err := s.checkLife(now, absoluteCap); err != nil {
		return Session{}, err
	}
	return s, nil
}

// authenticateSession checks no time, so a caller that reads the
// absolute cap from somewhere costly does so only for an authentic value.
func authenticateSession(current, previous []byte, value string) (Session, error) {
	s, fields, mac, err := parseSession(value)
	if err != nil {
		return Session{}, err
	}
	if !macMatches(current, sessionPurpose, fields, mac) && !macMatches(previous, sessionPurpose, fields, mac) {
		return Session{}, errSessionMAC
	}
	return s, nil
}

func (s Session) checkLife(now time.Time, absoluteCap time.Duration) error {
	if !now.Before(s.AnchoredAt.Add(absoluteCap)) {
		return errSessionCapReached
	}
	if !now.Before(s.ExpiresAt) {
		return errSessionExpired
	}
	return nil
}

func encodeSession(secret []byte, s Session) string {
	return encodeSigned(secret, sessionPurpose, s.fields())
}

// fields is the field order shared by the wire encoding and the MAC
// input, so the two cannot disagree about what was signed.
func (s Session) fields() [][]byte {
	return [][]byte{
		[]byte(sessionVersion),
		[]byte(s.Address),
		[]byte(strconv.FormatInt(s.IssuedAt.Unix(), 10)),
		[]byte(strconv.FormatInt(s.AnchoredAt.Unix(), 10)),
		[]byte(strconv.FormatInt(s.ExpiresAt.Unix(), 10)),
	}
}

// parseSession decodes value without authenticating it. It returns the
// raw fields so the MAC is verified over the bytes that arrived, not a
// re-serialization of them.
func parseSession(value string) (Session, [][]byte, []byte, error) {
	fields, mac, err := splitSigned(value, sessionFieldCount, errSessionMalformed)
	if err != nil {
		return Session{}, nil, nil, err
	}

	if string(fields[0]) != sessionVersion {
		return Session{}, nil, nil, fmt.Errorf("%w: version label", errSessionMalformed)
	}
	issued, err := strconv.ParseInt(string(fields[2]), 10, 64)
	if err != nil {
		return Session{}, nil, nil, fmt.Errorf("%w: issue time", errSessionMalformed)
	}
	anchored, err := strconv.ParseInt(string(fields[3]), 10, 64)
	if err != nil {
		return Session{}, nil, nil, fmt.Errorf("%w: anchor", errSessionMalformed)
	}
	expires, err := strconv.ParseInt(string(fields[4]), 10, 64)
	if err != nil {
		return Session{}, nil, nil, fmt.Errorf("%w: expiry", errSessionMalformed)
	}

	return Session{
		Address:    string(fields[1]),
		IssuedAt:   time.Unix(issued, 0).UTC(),
		AnchoredAt: time.Unix(anchored, 0).UTC(),
		ExpiresAt:  time.Unix(expires, 0).UTC(),
	}, fields, mac, nil
}
