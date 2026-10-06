package serve

import "net/http"

// originCheck refuses a state-changing browser request that did not
// come from this server's own pages. It is the one cross-site request
// forgery defense for every surface the shell serves: [Shell.Handler]
// places it after the ingress strip and path normalization and before
// prefix routing, so every registered subtree and every path outside
// the prefix sit behind it by construction rather than by each mount
// remembering to install it.
//
// The mechanism is [http.CrossOriginProtection], which honors
// Sec-Fetch-Site on its own and falls back to comparing Origin's host
// with Host only when Sec-Fetch-Site is absent. That order matters: a
// check reading Origin alone refuses every development request that
// reaches the server through a proxy rewriting Host. The stdlib value
// is too permissive in one respect, and originCheck tightens exactly
// that one: a request carrying neither Sec-Fetch-Site nor Origin is
// allowed by the stdlib, on the assumption that it is same-origin or
// not from a browser, and is refused here. A current browser sends at
// least one of the two on every state-changing request, so what the
// tightening refuses is a non-browser caller presenting a cookie, which
// only the bearer bypass below may admit.
//
// GET, HEAD and OPTIONS are always allowed, because the stdlib treats
// them as safe. A handler must therefore never change state on one of
// them, and a native client that does change state on a GET, as some
// streaming clients do, is never checked here: the handler serving it
// stands behind its own authentication alone.
//
// There is no double-submit token. The session cookie's __Host- prefix
// and SameSite=Lax attribute and this check each do a distinct job,
// and a token would repeat the check's job on a surface only browsers
// reach with a cookie.
//
// # The bearer bypass
//
// A request authenticated by a bearer credential rather than by the
// session cookie cannot be forged by a third-party page, which can make
// a browser attach cookies but never a bearer credential. Such a path
// is exempted through [originCheck.addBearerBypass], keyed on its
// path. The admin API's bearer mount and the OpenSubsonic API's
// non-GET requests are its consumers. The exemption carries a contract
// every handler behind a bypassed pattern keeps: it must not read the
// session cookie to authenticate a request. A bypassed handler that
// fell back to the cookie would turn the exemption into exactly the
// forgery this check refuses.
type originCheck struct {
	protection *http.CrossOriginProtection

	// bypass holds the same patterns as protection's own bypass set.
	// It is kept here because the stdlib cannot be asked whether a
	// request matched: protection allows a request carrying neither
	// header before it consults its bypass set, so the tightening must
	// decide the exemption for itself.
	bypass *http.ServeMux

	// deny answers every refused request.
	deny http.Handler
}

// bypassSentinel is the handler every bypass pattern registers. It is
// never served: the check compares the handler the mux resolves against
// bypassMatch, which tells a matched pattern from the mux's own
// not-found or redirect handler.
type bypassSentinel struct{}

func (bypassSentinel) ServeHTTP(http.ResponseWriter, *http.Request) {}

var bypassMatch http.Handler = bypassSentinel{}

func newOriginCheck() *originCheck {
	return &originCheck{
		protection: http.NewCrossOriginProtection(),
		bypass:     http.NewServeMux(),
		deny: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			markRefusal(r, http.StatusForbidden)
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
		}),
	}
}

// addBearerBypass exempts every request matching pattern, in
// [http.ServeMux] syntax, from the check. It registers pattern with the
// stdlib value, which governs a cross-origin request, and with the
// check's own set, which governs a request carrying neither header;
// either alone would leave one of the two refusing. It panics on a
// pattern the mux rejects or one already registered.
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
