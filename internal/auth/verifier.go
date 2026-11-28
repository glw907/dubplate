package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/identity"
	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/settings"
)

// A Verifier verifies the credential a request carries and yields the
// identity it asserts, so no caller inspects a credential's encoding.
// MagicLinkVerifier is the production implementation.
type Verifier interface {
	// VerifyIdentity returns the identity r's credential asserts, with
	// every time comparison made against now. sameOrigin is the caller's
	// reading of the request headers, since only the shell knows which
	// headers it trusts. A nil error permits the request. The sentinel
	// in a non-nil one names the failed check for logging only.
	VerifyIdentity(r *http.Request, now time.Time, sameOrigin bool) (VerifiedIdentity, error)
}

// VerifiedIdentity is what a Verifier yields for a verified credential.
type VerifiedIdentity struct {
	Identity identity.Identity

	// Admin is read from the roster record, never from the request.
	Admin bool

	// Session is refreshed when the request was same-origin.
	Session Session

	// RefreshedValue is the new cookie value, empty when nothing was
	// refreshed.
	RefreshedValue string
}

// MagicLinkVerifier is the [Verifier] for a member's session: it checks
// the session cookie against the roster and the revocation epoch on
// every call, and refreshes it only for a same-origin request.
type MagicLinkVerifier struct {
	roster *member.Roster
	epoch  *EpochStore

	currentSecret, previousSecret []byte
	lifetime, absoluteCap         settings.Reader[time.Duration]
}

var _ Verifier = (*MagicLinkVerifier)(nil)

// NewMagicLinkVerifier returns a MagicLinkVerifier over roster and epoch
// that reads cfg's session secrets and session readers. The cap is read
// at every verification, so a changed cap applies to the next request in
// both directions; the lifetime is read at every refresh.
func NewMagicLinkVerifier(roster *member.Roster, epoch *EpochStore, cfg Config) *MagicLinkVerifier {
	return &MagicLinkVerifier{
		roster:         roster,
		epoch:          epoch,
		currentSecret:  cfg.SessionSecret,
		previousSecret: cfg.PreviousSessionSecret,
		lifetime:       cfg.SessionLifetime,
		absoluteCap:    cfg.SessionAbsoluteCap,
	}
}

// VerifyIdentity implements [Verifier].
//
// The cap is read only after the MAC verifies, so a forged cookie costs
// no setting read. The roster is checked on every call, so a removed
// member is refused at their next request, not at expiry. A failed
// setting read wraps errSettingUnreadable and never falls back to a
// default.
func (m *MagicLinkVerifier) VerifyIdentity(r *http.Request, now time.Time, sameOrigin bool) (VerifiedIdentity, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	sess, err := authenticateSession(m.currentSecret, m.previousSecret, cookie.Value)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	absoluteCap, err := readSetting(r.Context(), m.absoluteCap)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	if err := sess.checkLife(now, absoluteCap); err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	if err := m.epoch.checkSession(sess); err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}
	member, err := m.roster.Lookup(r.Context(), sess.Address)
	if err != nil {
		return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
	}

	result := VerifiedIdentity{
		Identity: identity.Identity{
			Address:  member.Address,
			UserID:   member.ID,
			Username: member.Username,
			Channel:  identity.ChannelCookie,
		},
		Admin:   member.Admin,
		Session: sess,
	}
	if sameOrigin {
		lifetime, err := readSetting(r.Context(), m.lifetime)
		if err != nil {
			return VerifiedIdentity{}, fmt.Errorf("verify identity: %w", err)
		}
		value, refreshed := refreshSession(m.currentSecret, sess, now, lifetime, absoluteCap)
		result.Session = refreshed
		result.RefreshedValue = value
	}
	return result, nil
}
