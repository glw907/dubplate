package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

// ErrNotSignedIn is what [CheckRequest] wraps when the request carries
// no valid session: a missing, forged, expired, revoked, or
// unrostered one, or a verified identity that is no login. Any other
// error from it is the server failing, which a caller must not report
// as a signed-out visitor.
var ErrNotSignedIn = errors.New("auth: no valid session")

// CheckRequest verifies the session r carries through cfg.Seam, for a
// caller that admits every signed-in member rather than only an admin.
// It returns the verified identity, whose Admin flag the seam read from
// the roster and never from anything the request carries.
//
// The session is refreshed, and its new cookie set on w, only when
// Sec-Fetch-Site is exactly same-origin, as [RequireAdmin] does and for
// the same reason. A refresh moves the issue time and leaves the
// anchor.
//
// A refusal wraps [ErrNotSignedIn] and logs at Info the fixed reason
// [gateRefusalReason] names. Any other failure, a roster lookup or a
// setting read, is returned without [ErrNotSignedIn] and logged at Error. Neither log
// record carries an error's own text, which may quote an address.
func CheckRequest(cfg AdminConfig, w http.ResponseWriter, r *http.Request) (VerifiedIdentity, error) {
	now := cfg.Now()
	sameOrigin := r.Header.Get("Sec-Fetch-Site") == "same-origin"
	v, err := cfg.Seam.VerifyIdentity(r, now, sameOrigin)
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
			return VerifiedIdentity{}, fmt.Errorf("check request: %w", err)
		}
		return VerifiedIdentity{}, fmt.Errorf("%w: %s", ErrNotSignedIn, reason)
	}
	if v.RefreshedValue != "" {
		http.SetCookie(w, SessionCookie(v.RefreshedValue, now, v.Session.ExpiresAt))
	}
	return v, nil
}
