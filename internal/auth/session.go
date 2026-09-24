package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// SessionCookieName is the name of the session cookie (ruled,
// conductor 2026-09-08, R-5b9). The __Host- prefix is enforced by the
// browser rather than by this server: a cookie so named is stored only
// when it is Secure, carries no Domain, and has Path=/, which is why
// [SessionCookie] sets exactly those attributes and why same-origin is
// a stated precondition of the SSO capability rather than something a
// deployment can relax.
const SessionCookieName = "__Host-dubplate_session"

const (
	// sessionVersion labels the cookie's field layout, on the wire and
	// inside the MAC. A later field lands under a new label rather
	// than changing what an existing field means, so a cookie minted
	// before the change can never be read with the new layout's
	// meanings.
	sessionVersion = "dubplate-session-1"

	// sessionPurpose is the session cookie's domain-separation label,
	// distinct from the magic link's. The two credentials may be
	// signed under one secret and are the same shape on the wire, so
	// the label is the whole of what stops a link token from being
	// presented as a session.
	sessionPurpose = "dubplate/session"

	// sessionFieldCount is how many fields precede the MAC on the
	// wire: version, address, issue time, anchor, expiry.
	sessionFieldCount = 5
)

// Sentinels a caller branches on. The refusals are distinct so the
// session middleware can log which check failed while showing one
// generic outcome, which is the middleware's choice rather than this
// package's.
var (
	// ErrSessionMalformed reports a value that is not a session cookie
	// at all: the wrong field count, a field that is not base64url, an
	// unknown version label, or a field whose contents do not parse.
	ErrSessionMalformed = errors.New("auth: malformed session")

	// ErrSessionMAC reports a session value whose MAC verifies under
	// neither the current nor the previous secret.
	ErrSessionMAC = errors.New("auth: session signature")

	// ErrSessionCapReached reports a session whose anchor plus the
	// absolute cap has passed. It is reported even when the value's
	// own expiry is still in the future, which is the case a rolling
	// refresh produces.
	ErrSessionCapReached = errors.New("auth: session absolute cap reached")

	// ErrSessionExpired reports a session inside its absolute cap
	// whose own expiry has passed.
	ErrSessionExpired = errors.New("auth: session expired")
)

// Session is a session cookie's fields. [VerifySession] returns one
// only for a value that passed the MAC, the absolute cap, and the
// expiry, so a zero Session beside a non-nil error is the whole of what
// a refused value yields.
type Session struct {
	// Address is the logged-in address, normalized by
	// [identity.NormalizeAddress] before the MAC covered it.
	Address string

	// IssuedAt is when the session was last issued, at second
	// granularity. A refresh moves it.
	IssuedAt time.Time

	// AnchoredAt is the session's original issue time. Refresh moves
	// the issue time, so the anchor is the session's original issue
	// time: every check that must reason about when the session was
	// born reads the anchor rather than IssuedAt, which a refresh
	// advances. The absolute cap is measured from here, and so is
	// rung 9's revocation-epoch comparison.
	AnchoredAt time.Time

	// ExpiresAt is when the session stops verifying absent a refresh.
	ExpiresAt time.Time
}

// MintSession signs a session for address, valid for lifetime after
// now, and returns the encoded value alongside the fields it carries.
// The address is normalized here rather than assumed normalized, so the
// value a MAC covers is the value every later lookup keys on. The
// anchor is the issue time at a first mint; only [RefreshSession] can
// separate the two.
//
// The expiry is clamped to the anchor plus absoluteCap, the same clamp
// [RefreshSession] applies: a lifetime configured above the cap would
// otherwise mint a cookie whose Max-Age promises longer than
// [VerifySession] will ever honor, misleading whatever reads the
// cookie's own attributes rather than the value it carries.
//
// secret is the current session secret. There is no minting path for
// the previous secret, which is what makes rotation routine rather than
// an estate-wide sign-out.
func MintSession(secret []byte, address string, now time.Time, lifetime, absoluteCap time.Duration) (string, Session, error) {
	normalized, err := identity.NormalizeAddress(address)
	if err != nil {
		return "", Session{}, fmt.Errorf("mint session: %w", err)
	}
	s := Session{Address: normalized, IssuedAt: now.UTC().Truncate(time.Second)}
	s.AnchoredAt = s.IssuedAt
	expiresAt := s.IssuedAt.Add(lifetime)
	if capAt := s.AnchoredAt.Add(absoluteCap); expiresAt.After(capAt) {
		expiresAt = capAt
	}
	s.ExpiresAt = expiresAt
	return encodeSession(secret, s), s, nil
}

// RefreshSession re-signs s with its issue time moved to now and its
// expiry moved to the earlier of now plus lifetime and the session's
// own anchor plus absoluteCap, so the cookie a browser is handed never
// carries an expiry [VerifySession] will refuse anyway (review m2). The
// anchor is carried across untouched: a session refreshed any number
// of times is still refused once its anchor plus the absolute cap has
// passed, and moving the anchor here would remove that cap without any
// other change being visible.
//
// Callers refresh on same-origin requests only. A cross-origin
// subresource load counts as traffic to the browser but not to the
// person: honouring one would hold a session open to its absolute cap
// on a machine its owner walked away from, which is the shared-device
// adversary the threat model names. The session middleware is the
// caller that enforces the precondition; this function cannot see the
// request.
//
// s is a verified session, so there is nothing left to refuse and no
// error to return.
func RefreshSession(secret []byte, s Session, now time.Time, lifetime, absoluteCap time.Duration) (string, Session) {
	s.IssuedAt = now.UTC().Truncate(time.Second)
	expiresAt := s.IssuedAt.Add(lifetime)
	if capAt := s.AnchoredAt.Add(absoluteCap); expiresAt.After(capAt) {
		expiresAt = capAt
	}
	s.ExpiresAt = expiresAt
	return encodeSession(secret, s), s
}

// SessionCookie builds the cookie carrying value, expiring at
// expiresAt as measured from now. Every attribute follows from the
// __Host- prefix [SessionCookieName] carries: Secure and Path=/ are
// required for the browser to store it at all, and Domain must stay
// empty, which is also what confines the cookie to the exact origin
// that set it. SameSite is Lax so a magic link followed from a mail
// client arrives with the session attached.
//
// Both Expires and MaxAge carry the same instant (review m1): Expires
// is evaluated against the browser's own clock, so a client clock that
// drifts from the server's would otherwise keep or drop the cookie at
// the wrong moment, and MaxAge is the primary control OWASP's Session
// Management cheat sheet recommends
// (https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html#cookie-duration,
// read 2026-09-12) precisely because it does not depend on the
// client's clock at all. Either way, the value's own signed expiry is
// what the server enforces.
func SessionCookie(value string, now, expiresAt time.Time) *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    value,
		Path:     "/",
		Expires:  expiresAt,
		MaxAge:   int(expiresAt.Sub(now).Seconds()),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// ClearSessionCookie returns a cookie that instructs the browser to
// stop storing the session cookie, carrying the same Path, Secure,
// HttpOnly, SameSite, and absent Domain the __Host- prefix requires, so
// the browser recognizes it as the same cookie to delete (review m9).
// MaxAge is negative, which every browser treats as "delete
// immediately"; Expires carries a moment far in the past for a client
// that honors only the older attribute.
//
// These sessions are stateless: clearing the cookie does not
// invalidate the signed value it carried, so a session captured before
// this runs keeps verifying until its own expiry or an epoch bump
// revokes it. A caller's sign-out copy must not claim more than that.
func ClearSessionCookie() *http.Cookie {
	return &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0).UTC(),
		MaxAge:   -1,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

// VerifySession verifies value against the current secret and, when
// previous is non-empty, against the previous one, and returns the
// session's fields. A value that verifies under previous alone is
// accepted, which is the whole of what rotation overlap buys.
//
// The order is fixed: the MAC first, then the absolute cap, then the
// expiry. Reading either time first would act on an unauthenticated
// field. The cap precedes the expiry so a session that has outlived
// both reports the refusal a refresh cannot cure.
func VerifySession(current, previous []byte, value string, now time.Time, absoluteCap time.Duration) (Session, error) {
	s, fields, mac, err := parseSession(value)
	if err != nil {
		return Session{}, err
	}
	if !macMatches(current, sessionPurpose, fields, mac) && !macMatches(previous, sessionPurpose, fields, mac) {
		return Session{}, ErrSessionMAC
	}
	if !now.Before(s.AnchoredAt.Add(absoluteCap)) {
		return Session{}, ErrSessionCapReached
	}
	if !now.Before(s.ExpiresAt) {
		return Session{}, ErrSessionExpired
	}
	return s, nil
}

func encodeSession(secret []byte, s Session) string {
	return encodeSigned(secret, sessionPurpose, sessionFields(s))
}

// sessionFields is the cookie's field order, shared by the wire
// encoding and the MAC input so the two can never disagree about what
// was signed.
func sessionFields(s Session) [][]byte {
	return [][]byte{
		[]byte(sessionVersion),
		[]byte(s.Address),
		[]byte(strconv.FormatInt(s.IssuedAt.Unix(), 10)),
		[]byte(strconv.FormatInt(s.AnchoredAt.Unix(), 10)),
		[]byte(strconv.FormatInt(s.ExpiresAt.Unix(), 10)),
	}
}

// parseSession decodes value into its fields without authenticating any
// of them. It returns the fields as well as the session so the caller
// can verify the MAC over exactly the bytes that arrived, rather than
// over a re-serialization that could differ from them.
func parseSession(value string) (Session, [][]byte, []byte, error) {
	fields, mac, err := splitSigned(value, sessionFieldCount, ErrSessionMalformed)
	if err != nil {
		return Session{}, nil, nil, err
	}

	if string(fields[0]) != sessionVersion {
		return Session{}, nil, nil, fmt.Errorf("%w: version label", ErrSessionMalformed)
	}
	issued, err := strconv.ParseInt(string(fields[2]), 10, 64)
	if err != nil {
		return Session{}, nil, nil, fmt.Errorf("%w: issue time", ErrSessionMalformed)
	}
	anchored, err := strconv.ParseInt(string(fields[3]), 10, 64)
	if err != nil {
		return Session{}, nil, nil, fmt.Errorf("%w: anchor", ErrSessionMalformed)
	}
	expires, err := strconv.ParseInt(string(fields[4]), 10, 64)
	if err != nil {
		return Session{}, nil, nil, fmt.Errorf("%w: expiry", ErrSessionMalformed)
	}

	return Session{
		Address:    string(fields[1]),
		IssuedAt:   time.Unix(issued, 0).UTC(),
		AnchoredAt: time.Unix(anchored, 0).UTC(),
		ExpiresAt:  time.Unix(expires, 0).UTC(),
	}, fields, mac, nil
}
