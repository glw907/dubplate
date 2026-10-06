package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/identity"
)

// magicLinkSource is the identity source name a session verified through
// MagicLinkSeam reports. A member's session was always minted by
// consuming a magic link, so the source names that mechanism rather
// than the cookie carrying the session now.
const magicLinkSource = "magic-link"

// A Seam verifies the credential a request carries and yields the
// identity it asserts, so a caller never inspects a credential's own
// encoding to decide who is logged in or whether they are an operator.
// MagicLinkSeam is its implementation.
type Seam interface {
	// VerifyIdentity verifies the credential r carries and returns the
	// identity it asserts. now is the instant every time comparison
	// runs against. sameOrigin is the same-origin condition as the
	// caller's own header inspection determined it, since deriving it
	// here would require this package to know which headers a shell
	// trusts.
	//
	// A nil error means the request is permitted; any non-nil error
	// means it is refused. The specific sentinel names the check that
	// failed for a caller's own log line; the refusal itself never
	// depends on which one it is.
	VerifyIdentity(r *http.Request, now time.Time, sameOrigin bool) (VerifiedIdentity, error)

	// Read returns address's current revocation epoch, exposed as the
	// authority a credential store can consult directly rather than
	// relying only on a Bump notification it may have missed.
	Read(address string) (epoch time.Time, ok bool, err error)
}

// VerifiedIdentity is what a Seam yields for a credential that
// verified.
type VerifiedIdentity struct {
	// Identity is the verified login, carrying the login address, the
	// resolved principal, the display name, the source, and the
	// channel it was asserted on.
	Identity identity.Identity

	// Admin is read from the roster record at the seam, never from
	// anything the request carries.
	Admin bool

	// Session is the session backing Identity, refreshed when the
	// request was same-origin.
	Session Session

	// RefreshedValue is the newly signed session cookie value when
	// Session was refreshed, empty when the request was not
	// same-origin and nothing was refreshed.
	RefreshedValue string
}

// MagicLinkSeam is the Seam's reference implementation: it verifies the
// session cookie, re-checks the address against the roster and the
// revocation epoch on every call, and refreshes the session only for a
// same-origin request.
type MagicLinkSeam struct {
	roster *Roster
	epoch  *EpochStore

	currentSecret, previousSecret []byte
	lifetime, absoluteCap         time.Duration
}

var _ Seam = (*MagicLinkSeam)(nil)

// NewMagicLinkSeam returns a MagicLinkSeam verifying against roster and
// epoch, signing and clamping refreshed sessions the same way
// currentSecret, previousSecret, lifetime, and absoluteCap already
// govern minting and verification.
func NewMagicLinkSeam(roster *Roster, epoch *EpochStore, currentSecret, previousSecret []byte, lifetime, absoluteCap time.Duration) *MagicLinkSeam {
	return &MagicLinkSeam{
		roster:         roster,
		epoch:          epoch,
		currentSecret:  currentSecret,
		previousSecret: previousSecret,
		lifetime:       lifetime,
		absoluteCap:    absoluteCap,
	}
}

// VerifyIdentity implements Seam.
//
// The order is fixed: the session cookie's MAC and expiry, then the
// revocation epoch, then the roster. The roster lookup runs on every
// call rather than once at login, so a member removed after logging in
// stops working at their next request rather than at their session's
// expiry. A refresh, when the request is same-origin, runs last, after
// every refusal has already had its chance to deny.
func (m *MagicLinkSeam) VerifyIdentity(r *http.Request, now time.Time, sameOrigin bool) (VerifiedIdentity, error) {
	cookie, err := r.Cookie(SessionCookieName)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	sess, err := VerifySession(m.currentSecret, m.previousSecret, cookie.Value, now, m.absoluteCap)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	if err := m.epoch.CheckSession(sess); err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	member, err := m.roster.Lookup(sess.Address)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}

	principal := member.StreamerUsername
	if principal == "" {
		principal = member.Address
	}

	result := VerifiedIdentity{
		Identity: identity.Identity{
			Address:     member.Address,
			Principal:   principal,
			DisplayName: member.DisplayName,
			Source:      magicLinkSource,
			Channel:     identity.ChannelCookie,
		},
		Admin:   member.Admin,
		Session: sess,
	}
	if sameOrigin {
		value, refreshed := RefreshSession(m.currentSecret, sess, now, m.lifetime, m.absoluteCap)
		result.Session = refreshed
		result.RefreshedValue = value
	}
	return result, nil
}

// Read implements Seam.
func (m *MagicLinkSeam) Read(address string) (time.Time, bool, error) {
	return m.epoch.Read(address)
}
