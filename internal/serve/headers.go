package serve

import "net/http"

// hstsValue is one year and no other directive: includeSubDomains would
// reach sibling hosts under the same domain, and preload is the domain
// owner's decision.
const hstsValue = "max-age=31536000"

// EnableHSTS makes every response the shell answers carry
// Strict-Transport-Security. The caller decides from the public URL,
// never the request, since the shell speaks plain HTTP to whatever is
// in front of it. A caller must not enable it for a loopback host: RFC
// 6797 matches a host by name alone, so one response would pin every
// port on that host to https for the full max-age.
//
// Like [Shell.Register], a call after [Shell.Handler] panics.
func (s *Shell) EnableHSTS() {
	if s.handlerBuilt {
		panic("serve: EnableHSTS called after Handler; it belongs to construction")
	}
	s.hsts = true
}

// securityHeaders sets the package's security headers before next
// runs, so a handler that answers by [http.Error] or [http.NotFound]
// keeps them.
func (s *Shell) securityHeaders(next http.Handler) http.Handler {
	hsts := s.hsts
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
