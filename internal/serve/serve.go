package serve

import (
	"bufio"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
)

// Shell is the dubplate HTTP shell's construction value (C1, C5): the
// point a consumer registers a subtree under the owned prefix and sets
// the one root-level fallthrough slot, before [Shell.Handler] builds
// the composed root http.Handler. The shell owns the mux every
// registration lands in (R-L); it exposes no interface of its own,
// since http.Handler is the stdlib's and the extension point's whole
// value is that a consumer brings one.
type Shell struct {
	prefix   Prefix
	strip    Strip
	logger   *slog.Logger
	mux      *http.ServeMux
	subtrees map[string]struct{}

	fallthroughHandler http.Handler
	fallthroughSet     bool
}

// NewShell builds a [Shell] that owns prefix and strips ingress traffic
// through strip. logger receives the shell's one structured log record
// (the refusal record doc.go describes); a nil logger falls back to
// [slog.Default]. NewShell reads no environment and no [internal/config]
// value: every input is a construction-time argument, per the plan's
// shell-reads-values-not-environment rule.
func NewShell(prefix Prefix, strip Strip, logger *slog.Logger) *Shell {
	if logger == nil {
		logger = slog.Default()
	}
	return &Shell{
		prefix:   prefix,
		strip:    strip,
		logger:   logger,
		mux:      http.NewServeMux(),
		subtrees: make(map[string]struct{}),
	}
}

// Register mounts h under the owned prefix as the subtree name, one
// method taking a subtree name and an http.Handler (ruled, conductor
// 2026-09-08, R-L, as applied): both prefix+"/"+name and everything
// under prefix+"/"+name+"/" reach h. A path under the owned prefix that
// no registered subtree claims falls through to the mux's own default
// 404 (C4's terminator), never to the fallthrough slot.
//
// Two registrations of the same name panic, naming the subtree, on the
// same collision-gate precedent internal/config/registry.go cites
// (flag.(*FlagSet).Var, database/sql.Register, expvar.Publish): a
// colliding name is a programmer error in the binary's own wiring,
// never an adopter's misconfiguration, so it fails loudly at
// construction rather than silently at request time.
func (s *Shell) Register(name string, h http.Handler) {
	if _, exists := s.subtrees[name]; exists {
		panic(fmt.Sprintf("serve: subtree %q already registered", name))
	}
	s.subtrees[name] = struct{}{}

	pattern := s.prefix.String() + "/" + name
	s.mux.Handle(pattern, h)
	s.mux.Handle(pattern+"/", h)
}

// SetFallthrough sets the one root-level slot rung 12's reverse proxy
// occupies (C5, 2.0 hook 6's fallthrough half). A second call panics:
// like [Shell.Register], both calls are the binary's own wiring rather
// than adopter input, so a second call is a programmer error caught at
// construction instead of deferred to a request. A shell with no
// SetFallthrough call answers every request outside the owned prefix
// 404, the same status the prefix terminator returns (ruled, conductor
// 2026-09-08, R-N, as applied): a deployment with no streamer leaves
// the slot empty, which is what makes rung 5 safe to ship without a
// driver.
func (s *Shell) SetFallthrough(h http.Handler) {
	if s.fallthroughSet {
		panic("serve: fallthrough already set")
	}
	s.fallthroughHandler = h
	s.fallthroughSet = true
}

// Handler composes C1's four steps, in this order, each wrapping the
// next, and returns the root http.Handler: the ingress strip, path
// normalization, prefix routing, and the fallthrough. doc.go states the
// order and the extension point's contract; this method builds it. The
// order is a property of this composition, never a mux feature: no
// step here is registered as a route.
func (s *Shell) Handler() http.Handler {
	fallthroughHandler := s.fallthroughHandler
	if fallthroughHandler == nil {
		fallthroughHandler = http.HandlerFunc(http.NotFound)
	}

	routed := s.prefix.Route(s.mux, s.stripSessionCookie(fallthroughHandler))
	normalized := Normalize(routed)
	stripped := s.strip.Wrap(normalized)
	return s.logRefusal(stripped)
}

// stripSessionCookie wraps next, the fallthrough handler, removing the
// shell's own session cookie before next sees the request (ruled,
// conductor 2026-09-08, R-E, as applied: the removal lands at 5a, at
// the fallthrough boundary, parameterized by the cookie name). The
// removal lands here, on the way out to the fallthrough, rather than as
// a fourth strip group: the cookie's construction is 5b's, and the
// prefix-owned surface that reads it must still see it. A shell built
// with no cookie name removes nothing, which is 5a's own wiring.
func (s *Shell) stripSessionCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		removeCookie(r, s.strip.sessionCookie)
		next.ServeHTTP(w, r)
	})
}

// removeCookie deletes every cookie named name from r's raw Cookie
// header, leaving every other cookie's text byte-for-byte as the
// client sent it. It operates on the header string directly rather
// than [http.Request.Cookies] and a rebuilt Name+"="+Value: the latter
// loses RFC 6265 double quotes net/http strips on parse and drops any
// pair net/http considers invalid, either of which would hand rung
// 12's proxy a Cookie header the client never sent. A request carrying
// name twice loses both instances, so a front that appends rather than
// replaces cannot smuggle a second value past the removal. An empty
// name is a no-op, the shape a shell built before 5b's cookie exists
// takes.
func removeCookie(r *http.Request, name string) {
	if name == "" {
		return
	}
	header := r.Header.Get("Cookie")
	if header == "" {
		return
	}

	segments := strings.Split(header, ";")
	kept := make([]string, 0, len(segments))
	for _, seg := range segments {
		cookieName, _, _ := strings.Cut(strings.TrimSpace(seg), "=")
		if cookieName == name {
			continue
		}
		kept = append(kept, seg)
	}

	if len(kept) == 0 {
		r.Header.Del("Cookie")
		return
	}
	r.Header.Set("Cookie", strings.Join(kept, ";"))
}

// statusCapture is a http.ResponseWriter wrapper that records the
// status written, so logRefusal can decide whether the wrapped chain
// refused a request by observing the response it actually sent, never
// by reading the chain's construction. logRefusal wraps the whole
// chain, so statusCapture reaches every handler the shell dispatches
// to, registered subtrees and the fallthrough alike; doc.go reserves
// the fallthrough for rung 12's reverse proxy, and strip.go keeps the
// websocket Upgrade handshake intact, so statusCapture stays
// transparent to both: Unwrap lets [http.ResponseController] reach the
// real writer, and Flush and Hijack pass through directly for a caller
// that asserts [http.Flusher] or [http.Hijacker] itself.
type statusCapture struct {
	http.ResponseWriter
	status int
}

func (c *statusCapture) WriteHeader(status int) {
	c.status = status
	c.ResponseWriter.WriteHeader(status)
}

// Unwrap exposes the wrapped writer to [http.ResponseController], so a
// handler that reaches it via the controller sees the real writer's
// capabilities rather than statusCapture's own limited surface.
func (c *statusCapture) Unwrap() http.ResponseWriter {
	return c.ResponseWriter
}

// Flush passes through to the wrapped writer's [http.Flusher] when it
// has one, for a handler that type-asserts http.Flusher directly
// instead of going through http.ResponseController.
func (c *statusCapture) Flush() {
	if f, ok := c.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack passes through to the wrapped writer's [http.Hijacker] when
// it has one, returning [http.ErrNotSupported] otherwise, for a
// handler that type-asserts http.Hijacker directly.
func (c *statusCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := c.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// logRefusal wraps the whole composed chain (next) and, once it
// returns, writes one structured log record naming whether any strip-set
// identity spelling was present on the request, for every request C3
// refused (400 or 414, its only two refusal statuses; a 404 from the
// prefix terminator or the empty fallthrough emits no record). Positioned
// outside strip, normalization, and prefix routing alike, it relies on
// http.Request's shared pointer rather than its own nesting position:
// by the time next.ServeHTTP returns, r.Header carries whatever the
// chain actually did to it, strip's removal included if strip ran
// before the refusal and not otherwise. This is the mechanism the
// fourth order assertion reads (spec:938-944).
func (s *Shell) logRefusal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture := &statusCapture{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(capture, r)

		if capture.status != http.StatusBadRequest && capture.status != http.StatusRequestURITooLong {
			return
		}
		s.logger.Info("serve: refused request",
			"path", r.URL.Path,
			"status", capture.status,
			"identity_spelling_present", identitySpellingPresent(r, s.strip.header),
		)
	})
}

// identitySpellingPresent reports whether r's header carries any of the
// strip's identity spellings for driverHeader, group 1's own list
// (identityExtras plus the driver's configured name), so logRefusal
// asserts against the same set the strip removes rather than a second
// one written out here.
func identitySpellingPresent(r *http.Request, driverHeader string) bool {
	for _, name := range identitySpellings(driverHeader) {
		if r.Header.Get(name) != "" {
			return true
		}
	}
	return false
}
