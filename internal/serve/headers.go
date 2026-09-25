package serve

import "net/http"

// hstsValue is the Strict-Transport-Security value the shell sends
// when HSTS is enabled: one year, and no other directive. RFC 6797's
// includeSubDomains would reach sibling hosts an operator runs under
// the same domain, and preload is a decision only the domain's owner
// can make.
const hstsValue = "max-age=31536000"

// EnableHSTS makes every response the shell answers carry
// Strict-Transport-Security. The caller decides from the public URL
// browsers reach the server at, never from the request, because the
// shell itself speaks plain HTTP to the proxy or tunnel in front of it.
// RFC 6797 matches a known HSTS host by name alone, so a caller must
// not enable it for a loopback host: one response would pin every port
// on that host to https in the browser for the full max-age.
//
// Like [Shell.Register], a call after [Shell.Handler] has composed the
// chain panics, since the composed chain has already read the setting.
func (s *Shell) EnableHSTS() {
	if s.handlerBuilt {
		panic("serve: EnableHSTS called after Handler; it belongs to construction")
	}
	s.hsts = true
}

// securityHeaders sets the headers every response the shell answers
// carries, whichever step answers it: X-Content-Type-Options: nosniff,
// Referrer-Policy: no-referrer, and, when hsts is set,
// Strict-Transport-Security. It sets them before next runs, so a
// handler that answers by [http.Error] or [http.NotFound] keeps them.
func securityHeaders(hsts bool, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("Referrer-Policy", "no-referrer")
		if hsts {
			h.Set("Strict-Transport-Security", hstsValue)
		}
		next.ServeHTTP(w, r)
	})
}
