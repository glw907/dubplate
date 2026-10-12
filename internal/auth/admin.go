package auth

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// AdminConfig is everything the admin gate and [CheckRequest] need,
// resolved by the caller. Every field is required.
type AdminConfig struct {
	// Seam verifies the session a request carries. The admin flag the
	// gate checks is the one the seam read from the roster record.
	Seam Seam

	// ClientIP returns the client address a request is held
	// accountable for, under the shell's trusted-proxy policy.
	ClientIP func(r *http.Request) string

	// Now returns the instant a request is evaluated at.
	Now func() time.Time

	Logger *slog.Logger
}

// adminKey is the context key under which the gate passes the
// verified admin to the handler it wraps.
type adminKey struct{}

// RequireAdmin admits a request to next only when its session verifies
// through cfg.Seam as a valid identity holding the admin flag. Every
// other request answers exactly what [http.NotFound] answers and
// carries no cookie, so a gated path tells a non-admin nothing an
// unregistered path would not. next decides every method and path
// check itself, and those answers reach verified admins only.
//
// The session is refreshed only when Sec-Fetch-Site is exactly
// same-origin: the shell's cross-origin check always admits a GET, so
// having been reached says nothing about where a GET came from. The
// seam refreshes a verified non-admin's session as well, so the
// refreshed cookie is set only once the admin check has passed; setting
// it earlier would put a Set-Cookie on a non-admin's 404 that an
// unregistered path never carries.
func RequireAdmin(cfg AdminConfig, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
		case !v.Admin:
			reason = "not an admin"
		}
		if reason != "" {
			cfg.Logger.Log(r.Context(), level, "admin gate refused", "reason", reason, "client_ip", cfg.ClientIP(r))
			http.NotFound(w, r)
			return
		}
		if v.RefreshedValue != "" {
			http.SetCookie(w, SessionCookie(v.RefreshedValue, now, v.Session.ExpiresAt))
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), adminKey{}, v)))
	})
}

// gateRefusals are the causes the gate's log names. A seam error's own
// text is never logged: the epoch check wraps the normalizer's error,
// which quotes the address a signed session carries, and a failed
// setting read's text may carry whatever the read met.
var gateRefusals = []error{
	http.ErrNoCookie,
	ErrSessionMalformed,
	ErrSessionMAC,
	ErrSessionCapReached,
	ErrSessionExpired,
	ErrSessionRevoked,
	ErrMemberNotFound,
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
