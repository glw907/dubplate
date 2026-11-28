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

// VerifyConfig is what [RequireAdmin] and [VerifyRequest] need. Every
// field is required.
type VerifyConfig struct {
	Verifier Verifier

	// ClientIP returns the client address under the shell's
	// trusted-proxy policy.
	ClientIP func(r *http.Request) string

	Now func() time.Time

	Logger *slog.Logger
}

// RequireAdmin admits a request to next only when its session verifies
// as an identity holding the admin flag. Every other request gets
// exactly what [http.NotFound] answers, with no cookie, so a gated path
// tells a non-admin nothing an unregistered path would not.
//
// The session is refreshed only when Sec-Fetch-Site is exactly
// same-origin: the shell's cross-origin check admits every GET, so
// being reached says nothing about where a GET came from. The refreshed
// cookie is set only after the admin check, so a non-admin's 404
// carries no Set-Cookie.
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

// VerifyRequest verifies the session r carries, for a caller that admits
// every signed-in member, and returns the verified identity. It
// refreshes the session and sets its cookie on w under the same
// same-origin rule as [RequireAdmin].
//
// A request with no valid session is refused with an error wrapping
// [ops.CodeUnauthenticated] and logged at Info. Any other failure, a
// roster lookup or a setting read, is the server failing: it is returned
// without that code, so no caller reports it as a signed-out visitor,
// and logged at Error. No log record carries an error's text, which may
// quote an address.
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
		return VerifiedIdentity{}, fmt.Errorf("%w: %s", ops.CodeUnauthenticated, c.reason)
	}
	c.setRefreshed(w)
	return c.v, nil
}

// A gateCheck is one request's session verified at now. reason is empty
// for a verified login and otherwise names the refusal to log at level.
type gateCheck struct {
	v      VerifiedIdentity
	now    time.Time
	reason string
	level  slog.Level
	err    error
}

// check sets no cookie, so each gate decides whether the refreshed one
// reaches the response.
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
		http.SetCookie(w, sessionCookie(c.v.RefreshedValue, c.now, c.v.Session.ExpiresAt))
	}
}

// gateRefusals are the causes the gate's log names. A verifier error's
// text is never logged, since a failed read's text may carry anything.
var gateRefusals = []error{
	http.ErrNoCookie,
	errSessionMalformed,
	errSessionMAC,
	errSessionCapReached,
	errSessionExpired,
	errSessionRevoked,
	member.ErrMemberNotFound,
}

// gateRefusalReason names err by the first of gateRefusals it wraps, at
// Info. Any other error is the server failing, named at Error.
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
