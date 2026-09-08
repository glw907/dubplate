package serve

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/textproto"
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

	// handlerBuilt records that Handler has composed the root chain,
	// after which the mux is live and a further registration would
	// never be reached by an already-serving binary.
	handlerBuilt bool
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
//
// A registration after [Shell.Handler] has composed the root chain
// panics for the same reason, naming the subtree: registration is the
// binary's own wiring and belongs to construction, and a late
// registration in a binary already serving would land on a mux the
// composed chain is reading concurrently.
func (s *Shell) Register(name string, h http.Handler) {
	if s.handlerBuilt {
		panic(fmt.Sprintf("serve: subtree %q registered after Handler; registration belongs to construction", name))
	}
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
	s.handlerBuilt = true

	fallthroughHandler := s.fallthroughHandler
	if fallthroughHandler == nil {
		fallthroughHandler = http.HandlerFunc(http.NotFound)
	}

	routed := s.prefix.Route(s.terminated(), s.stripSessionCookie(fallthroughHandler))
	normalized := Normalize(routed)
	stripped := s.strip.Wrap(normalized)
	return s.logRefusal(stripped)
}

// terminated is the owned half of C4: a path under the owned prefix
// that some registered subtree claims goes to that subtree, and every
// other owned path goes to [Terminator]. The terminator is wired here
// explicitly rather than left to http.ServeMux's own default 404,
// which answers the same status by coincidence of the mux's behavior
// rather than by the shell's decision. Rung 14 depends on the
// decision, not the status: an unknown owned path must stop here
// rather than reach the proxy at the fallthrough.
func (s *Shell) terminated() http.Handler {
	terminator := Terminator()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := s.mux.Handler(r)
		if pattern == "" {
			terminator.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
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

// removeCookie deletes every cookie named name from every Cookie
// header line r carries, leaving every other cookie's text
// byte-for-byte as the client sent it. It operates on the header
// strings directly rather than [http.Request.Cookies] and a rebuilt
// Name+"="+Value: the latter loses RFC 6265 double quotes net/http
// strips on parse and drops any pair net/http considers invalid,
// either of which would hand rung 12's proxy a Cookie header the
// client never sent. A request carrying name twice loses both
// instances, so a front that appends rather than replaces cannot
// smuggle a second value past the removal. An empty name is a no-op,
// the shape a shell built before 5b's cookie exists takes.
//
// Every line is read, not only the first. A client may send Cookie
// more than once, and net/http keeps each as its own value; rewriting
// only r.Header.Get("Cookie")'s line would collapse the rest of them
// into nothing and leave a session cookie standing on any line but the
// first.
func removeCookie(r *http.Request, name string) {
	if name == "" {
		return
	}
	values := r.Header["Cookie"]
	if len(values) == 0 {
		return
	}

	kept := make([]string, 0, len(values))
	for _, value := range values {
		if rewritten := withoutCookie(value, name); rewritten != "" {
			kept = append(kept, rewritten)
		}
	}

	if len(kept) == 0 {
		r.Header.Del("Cookie")
		return
	}
	r.Header["Cookie"] = kept
}

// withoutCookie returns one Cookie header value with every segment
// naming name removed, or "" when nothing survives. A value naming
// nothing is returned unchanged, so the common case never rewrites the
// client's bytes at all. Surviving segments keep their exact text,
// quotes and spacing included; the one edit is trimming the leading
// space that separated a dropped first segment from the one now at the
// head of the value.
func withoutCookie(value, name string) string {
	segments := strings.Split(value, ";")
	kept := make([]string, 0, len(segments))
	for _, segment := range segments {
		cookieName, _, _ := strings.Cut(strings.TrimSpace(segment), "=")
		if cookieName == name {
			continue
		}
		kept = append(kept, segment)
	}

	if len(kept) == len(segments) {
		return value
	}
	if len(kept) == 0 {
		return ""
	}
	kept[0] = strings.TrimLeft(kept[0], " \t")
	if len(kept) == 1 && kept[0] == "" {
		return ""
	}
	return strings.Join(kept, ";")
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

// ReadFrom forwards to the wrapped writer's [io.ReaderFrom] when it has
// one, and copies otherwise. net/http's own response writer implements
// io.ReaderFrom to reach sendfile(2), and [io.Copy] consults the
// destination for that interface; a wrapper that did not forward it
// would silently downgrade every proxied stream rung 12 serves to a
// userspace copy.
func (c *statusCapture) ReadFrom(src io.Reader) (int64, error) {
	if rf, ok := c.ResponseWriter.(io.ReaderFrom); ok {
		return rf.ReadFrom(src)
	}
	return io.Copy(c.ResponseWriter, src)
}

// refusalContextKey is the context key logRefusal files its per-request
// [refusalRecord] under.
type refusalContextKey struct{}

// refusalRecord is the per-request slot the shell's own refusing
// wrappers write to. A zero Status means the shell refused nothing,
// whatever status a handler beyond it went on to write.
type refusalRecord struct {
	Status int
}

// markRefusal records that the shell itself refused r with status. It
// is a no-op outside the composed chain, so [Normalize] stays usable
// on its own.
func markRefusal(r *http.Request, status int) {
	if rec, ok := r.Context().Value(refusalContextKey{}).(*refusalRecord); ok {
		rec.Status = status
	}
}

// logRefusal wraps the whole composed chain (next) and, once it
// returns, writes one structured log record naming whether any
// strip-set identity spelling was present on the request, for every
// request the shell itself refused.
//
// The record is keyed on the shell's own refusal, filed by the
// wrapper that made it, never on the response status alone: a
// registered subtree answering 400 to its own bad input is not an
// ingress refusal, and reading the status would report one every time
// a handler rejected a malformed form. A 404 from the prefix
// terminator or the empty fallthrough is likewise no refusal, and
// emits no record.
//
// Positioned outside strip, normalization, and prefix routing alike,
// it relies on http.Request's shared header map rather than its own
// nesting position: by the time next.ServeHTTP returns, r.Header
// carries whatever the chain actually did to it, strip's removal
// included if strip ran before the refusal and not otherwise. This is
// the mechanism the fourth order assertion reads (spec:938-944).
func (s *Shell) logRefusal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &refusalRecord{}
		r = r.WithContext(context.WithValue(r.Context(), refusalContextKey{}, rec))

		capture := &statusCapture{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(capture, r)

		if rec.Status == 0 {
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
//
// It tests the key's presence in the map, not the value it holds. A
// header sent with an empty value is still a header the client sent,
// and the record exists to say whether one arrived at all.
func identitySpellingPresent(r *http.Request, driverHeader string) bool {
	for _, name := range identitySpellings(driverHeader) {
		if _, ok := r.Header[textproto.CanonicalMIMEHeaderKey(name)]; ok {
			return true
		}
	}
	return false
}
