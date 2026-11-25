package auth

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/glw907/dubplate/internal/ops"
)

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
// refused with an error wrapping [ops.CodeUnauthenticated], and logged at
// Info with the fixed reason [gateRefusalReason] names. Any other
// failure, a roster lookup or a setting read, is the server failing: it
// is returned without that code, so no caller reports it as a
// signed-out visitor, and logged at Error. Neither log
// record carries an error's own text, which may quote an address.
func VerifyRequest(cfg VerifyConfig, w http.ResponseWriter, r *http.Request) (VerifiedIdentity, error) {
	now := cfg.Now()
	sameOrigin := r.Header.Get("Sec-Fetch-Site") == "same-origin"
	v, err := cfg.Verifier.VerifyIdentity(r, now, sameOrigin)
	var reason string
	level := slog.LevelInfo
	switch {
	case err != nil:
		reason, level = gateRefusalReason(err)
	case !v.Identity.Valid():
		reason = "identity is not a verified login"
	}
	if reason != "" {
		message := "session check refused"
		if level >= slog.LevelError {
			message = "session check failed"
		}
		cfg.Logger.Log(r.Context(), level, message, "reason", reason, "client_ip", cfg.ClientIP(r))
		if level >= slog.LevelError {
			return VerifiedIdentity{}, fmt.Errorf("verify request: %w", err)
		}
		return VerifiedIdentity{}, fmt.Errorf("%w: %s", ops.CodeUnauthenticated, reason)
	}
	if v.RefreshedValue != "" {
		http.SetCookie(w, SessionCookie(v.RefreshedValue, now, v.Session.ExpiresAt))
	}
	return v, nil
}
