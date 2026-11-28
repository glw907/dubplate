package serve

import "net/http"

// originCheck refuses a state-changing browser request that did not
// come from this server's own pages: the one cross-site request forgery
// defense for every surface the shell serves. [Shell.Handler] places it
// before subtree dispatch, so every subtree sits behind it without each
// mount installing it.
//
// It is [http.CrossOriginProtection], which reads Sec-Fetch-Site and
// falls back to comparing Origin with Host, tightened in one respect:
// a request carrying neither header, which the stdlib allows, is
// refused. A current browser sends one of the two on every
// state-changing request, so the tightening refuses a non-browser
// caller presenting a cookie, which only a bearer bypass may admit.
//
// GET, HEAD, and OPTIONS are always allowed, so a handler must never
// change state on one; a native client that does is guarded by its
// handler's own authentication alone.
//
// There is no double-submit token: on a surface only browsers reach
// with a cookie, it would repeat this check's job.
//
// # The bearer bypass
//
// A third-party page can make a browser attach cookies but never a
// bearer credential, so a bearer-authenticated path is exempted through
// [originCheck.addBearerBypass]. Every handler behind a bypassed
// pattern must not authenticate by the session cookie: one that fell
// back to it would turn the exemption into the forgery this check
// refuses.
type originCheck struct {
	protection *http.CrossOriginProtection

	// bypass holds protection's bypass patterns again, since protection
	// allows a request carrying neither header before it consults its
	// own set and cannot be asked whether a request matched.
	bypass *http.ServeMux

	deny http.Handler
}

// bypassSentinel is the handler every bypass pattern registers, never
// served: comparing the resolved handler with bypassMatch tells a
// matched pattern from the mux's not-found or redirect handler.
type bypassSentinel struct{}

func (bypassSentinel) ServeHTTP(http.ResponseWriter, *http.Request) {}

var bypassMatch http.Handler = bypassSentinel{}

func newOriginCheck() *originCheck {
	return &originCheck{
		protection: http.NewCrossOriginProtection(),
		bypass:     http.NewServeMux(),
		deny: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			refuse(w, r, http.StatusForbidden)
		}),
	}
}

// addBearerBypass exempts every request matching pattern, in
// [http.ServeMux] syntax, from the check, in both the stdlib value and
// the check's own set. It panics on a pattern the mux rejects or one
// already registered.
func (o *originCheck) addBearerBypass(pattern string) {
	o.bypass.Handle(pattern, bypassMatch)
	o.protection.AddInsecureBypassPattern(pattern)
}

func (o *originCheck) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !o.allowed(r) {
			o.deny.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (o *originCheck) allowed(r *http.Request) bool {
	if o.protection.Check(r) != nil {
		return false
	}
	switch r.Method {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	if r.Header.Get("Sec-Fetch-Site") != "" || r.Header.Get("Origin") != "" {
		return true
	}
	h, _ := o.bypass.Handler(r)
	return h == bypassMatch
}
