package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/glw907/dubplate/internal/member"
	"github.com/glw907/dubplate/internal/ops"
)

// VerifyConfig is everything a session check needs, for the admin gate
// [RequireAdmin] and for [VerifyRequest], which admits every member.
// The caller builds it, and every field is required.
type VerifyConfig struct {
	// Verifier verifies the session a request carries. The admin flag the
	// gate checks is the one the verifier read from the roster record.
	Verifier Verifier

	// ClientIP returns the client address a request is held
	// accountable for, under the shell's trusted-proxy policy.
	ClientIP func(r *http.Request) string

	// Now returns the instant a request is evaluated at.
	Now func() time.Time

	Logger *slog.Logger
}

// RequireAdmin admits a request to next only when its session verifies
// through cfg.Verifier as a valid identity holding the admin flag. Every
// other request answers exactly what [http.NotFound] answers and
// carries no cookie, so a gated path tells a non-admin nothing an
// unregistered path would not. next decides every method and path
// check itself, and those answers reach verified admins only.
//
// The session is refreshed only when Sec-Fetch-Site is exactly
// same-origin: the shell's cross-origin check always admits a GET, so
// having been reached says nothing about where a GET came from. The
// verifier refreshes a verified non-admin's session as well, so the
// refreshed cookie is set only once the admin check has passed; setting
// it earlier would put a Set-Cookie on a non-admin's 404 that an
// unregistered path never carries.
func RequireAdmin(cfg VerifyConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c := cfg.check(r)
		if c.reason == "" && !c.v.Admin {
			c.reason, c.level = "not an admin", slog.LevelInfo
		}
		if c.reason != "" {
			cfg.Logger.Log(r.Context(), c.level, "admin gate refused", "reason", c.reason, "client_ip", cfg.ClientIP(r))
			http.NotFound(w, r)
			return
		}
		c.setRefreshed(w)
		next.ServeHTTP(w, r)
	})
}

// VerifyRequest verifies the session r carries through cfg.Verifier, for a
// caller that admits every signed-in member rather than only an admin.
// It returns the verified identity, whose Admin flag the verifier read from
// the roster and never from anything the request carries.
//
// The session is refreshed, and its new cookie set on w, only when
// Sec-Fetch-Site is exactly same-origin, as [RequireAdmin] does and for
// the same reason. A refresh moves the issue time and leaves the
// anchor.
//
// A request with no valid session (a missing, forged, expired, revoked,
// or unrostered one, or a verified identity that is no login) is
// refused with an error wrapping [ops.ErrUnauthenticated], and logged at
// Info with the fixed reason [gateRefusalReason] names. Any other
// failure, a roster lookup or a setting read, is the server failing: it
// is returned without that sentinel, so no caller reports it as a
// signed-out visitor, and logged at Error. Neither log
// record carries an error's own text, which may quote an address.
func VerifyRequest(cfg VerifyConfig, w http.ResponseWriter, r *http.Request) (VerifiedIdentity, error) {
	c := cfg.check(r)
	if c.reason != "" {
		message := "session check refused"
		if c.level >= slog.LevelError {
			message = "session check failed"
		}
		cfg.Logger.Log(r.Context(), c.level, message, "reason", c.reason, "client_ip", cfg.ClientIP(r))
		if c.level >= slog.LevelError {
			return VerifiedIdentity{}, fmt.Errorf("verify request: %w", c.err)
		}
		return VerifiedIdentity{}, fmt.Errorf("%w: %s", ops.ErrUnauthenticated, c.reason)
	}
	c.setRefreshed(w)
	return c.v, nil
}

// A gateCheck is one request's session verified at now. reason is empty
// when the session verifies as a login, and otherwise names the refusal
// to log at level; err is the verifier's error, if it returned one.
type gateCheck struct {
	v      VerifiedIdentity
	now    time.Time
	reason string
	level  slog.Level
	err    error
}

// check verifies r's session, asking for a refresh only when
// Sec-Fetch-Site is exactly same-origin. It sets no cookie, so each gate
// decides whether the refreshed one reaches the response.
func (cfg VerifyConfig) check(r *http.Request) gateCheck {
	c := gateCheck{now: cfg.Now(), level: slog.LevelInfo}
	sameOrigin := r.Header.Get("Sec-Fetch-Site") == "same-origin"
	c.v, c.err = cfg.Verifier.VerifyIdentity(r, c.now, sameOrigin)
	switch {
	case c.err != nil:
		c.reason, c.level = gateRefusalReason(c.err)
	case !c.v.Identity.Valid():
		c.reason = "identity is not a verified login"
	}
	return c
}

func (c gateCheck) setRefreshed(w http.ResponseWriter) {
	if c.v.RefreshedValue != "" {
		http.SetCookie(w, SessionCookie(c.v.RefreshedValue, c.now, c.v.Session.ExpiresAt))
	}
}

// gateRefusals are the causes the gate's log names. A verifier error's
// text is never logged, because a failed setting read's text may carry
// whatever the read met.
var gateRefusals = []error{
	http.ErrNoCookie,
	ErrSessionMalformed,
	ErrSessionMAC,
	ErrSessionCapReached,
	ErrSessionExpired,
	ErrSessionRevoked,
	member.ErrMemberNotFound,
}

// gateRefusalReason names err by the first of gateRefusals it wraps,
// at Info. Any other error is the server failing, not a visitor failing
// to verify, so it is named at Error: a failed setting read by its key,
// and anything else by a fixed reason.
func gateRefusalReason(err error) (string, slog.Level) {
	if errors.Is(err, errSettingUnreadable) {
		return settingFailureReason(err), slog.LevelError
	}
	for _, known := range gateRefusals {
		if errors.Is(err, known) {
			return known.Error(), slog.LevelInfo
		}
	}
	return "roster lookup failed", slog.LevelError
}
