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
// caller presenting a cookie.
//
// GET, HEAD, and OPTIONS are always allowed, so a handler must never
// change state on one; a native client that does is guarded by its
// handler's own authentication alone.
//
// There is no double-submit token: on a surface only browsers reach
// with a cookie, it would repeat this check's job.
type originCheck struct {
	protection *http.CrossOriginProtection
	deny       http.Handler
}

func newOriginCheck() *originCheck {
	return &originCheck{
		protection: http.NewCrossOriginProtection(),
		deny: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			refuse(w, r, http.StatusForbidden)
		}),
	}
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
	return r.Header.Get("Sec-Fetch-Site") != "" || r.Header.Get("Origin") != ""
}
