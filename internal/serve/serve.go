package serve

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"unicode"
)

// Shell is the dubplate HTTP shell's construction value: the point a
// consumer registers a subtree under the owned prefix before
// [Shell.Handler] builds the composed root http.Handler. The shell owns
// the mux every registration lands in; it exposes no interface of its
// own, since http.Handler is the stdlib's and the extension point's
// whole value is that a consumer brings one.
type Shell struct {
	prefix   Prefix
	strip    Strip
	origin   *originCheck
	logger   *slog.Logger
	mux      *http.ServeMux
	subtrees map[string]struct{}

	hsts bool

	// forwardedWarning holds the direct-mode warning
	// warnForwardedFromLocalPeer writes to once per shell, and serve
	// builds one shell per process.
	forwardedWarning sync.Once

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
		origin:   newOriginCheck(),
		logger:   logger,
		mux:      http.NewServeMux(),
		subtrees: make(map[string]struct{}),
	}
}

// Register mounts h under the owned prefix as the subtree name: both
// prefix+"/"+name and everything under prefix+"/"+name+"/" reach h. A
// path under the owned prefix that no registered subtree claims is
// answers 404.
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
//
// A name that would widen the mount past one subtree panics too, on the
// same precedent and naming the reason: see subtreeNameRefusal.
func (s *Shell) Register(name string, h http.Handler) {
	if reason := subtreeNameRefusal(name); reason != "" {
		panic(fmt.Sprintf("serve: subtree %q: %s", name, reason))
	}
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

// subtreeNameRefusal reports why name may not be a subtree, or "" when
// it may. Every refused shape silently widens the mount rather than
// failing visibly, which is what makes a panic the right answer:
//
//   - An empty name registers the owned prefix itself and the whole
//     owned subtree under it, so one handler swallows every path
//     the shell otherwise answers 404.
//   - A name carrying "/" registers a deeper path than the caller wrote,
//     and one carrying "{" or "}" is a Go 1.22 mux wildcard: "{x}"
//     matches any single segment, and "{x...}" matches the rest of the
//     path.
//   - Whitespace cannot appear in a request path this shell admits
//     (Normalize refuses a space as an unescaped byte), so a name
//     carrying it registers a pattern nothing can ever reach: a silent
//     no-op rather than a widening, and still a wiring mistake.
func subtreeNameRefusal(name string) string {
	switch {
	case name == "":
		return "must not be empty; an empty name mounts the whole owned prefix"
	case strings.Contains(name, "/"):
		return "must carry no separator"
	case strings.ContainsAny(name, "{}"):
		return "must carry no brace; a brace is a mux wildcard"
	case strings.ContainsFunc(name, unicode.IsSpace):
		return "must carry no whitespace"
	}
	return ""
}

// Handler composes six steps, in this order, each wrapping the next,
// and returns the root http.Handler: the security headers, the ingress
// strip, path normalization, the echo-method refusal, the origin check,
// and subtree dispatch. doc.go states the order and the extension point's
// contract; this method builds it. The order is a property of this
// composition, never a mux feature: no step here is registered as a
// route.
//
// The security headers are outermost, so every response carries them,
// a refusal by any later step included.
//
// The origin check sits between normalization and subtree dispatch, so
// it judges only a path normalization has admitted and covers every
// registered subtree and every path outside the prefix alike. It runs
// after the echo method refusal, so TRACE and TRACK keep their 405
// rather than taking the check's 403.
//
// A path no registered subtree claims, outside the owned prefix or
// under it, answers 404 from [http.NotFoundHandler].
func (s *Shell) Handler() http.Handler {
	s.handlerBuilt = true

	protected := s.origin.wrap(s.terminated())
	guarded := refuseEchoMethods(protected)
	normalized := Normalize(guarded)
	stripped := s.strip.Wrap(normalized)
	return securityHeaders(s.hsts, s.logRefusal(s.warnForwardedFromLocalPeer(stripped)))
}

// warnForwardedFromLocalPeer wraps next, the strip, and in direct mode
// logs one warning for the shell the first time a loopback or
// private peer sends a forwarded-address header. That request is the
// signature of a same-host or same-network front left in direct mode:
// every client it relays arrives from the front's own address, so the
// whole group shares one rate-limit key and one client's flood locks
// every member out of sign-in. A browser sends no such header on its
// own, so the warning names the likely cause rather than a certainty.
// It runs before the strip, which
// removes these headers from every peer in direct mode. Outside direct
// mode it returns next unchanged, since a configured trusted proxy is
// the remedy.
//
// The count is the strip's own group 2 set by the strip's own
// comparison, so the warning fires on exactly the headers the strip
// discards. [netip.Addr.IsPrivate] documents itself as no security
// property and no access-control input; it decides only whether to
// log.
func (s *Shell) warnForwardedFromLocalPeer(next http.Handler) http.Handler {
	if s.strip.proxy.Mode() != ModeDirect {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.noteForwardedFromLocalPeer(r)
		next.ServeHTTP(w, r)
	})
}

func (s *Shell) noteForwardedFromLocalPeer(r *http.Request) {
	peer := PeerAddress(r)
	addr, err := netip.ParseAddr(peer)
	if err != nil {
		return
	}
	if addr = addr.Unmap(); !addr.IsLoopback() && !addr.IsPrivate() {
		return
	}
	n := countPresent(r.Header, s.strip.forwarded)
	if n == 0 {
		return
	}
	s.forwardedWarning.Do(func() {
		s.logger.Warn("serve: a loopback or private peer sent a forwarded-address header in direct mode; a front on this host or network must be configured as a trusted proxy, or every client it relays shares its address and its rate limits",
			"peer", peer,
			"forwarded_headers", n,
		)
	})
}

// Logger returns the logger the shell writes its refusal record to,
// [slog.Default] when NewShell was built with none. [Run] reads it so
// net/http's own protocol errors reach the same sink as the shell's
// record, rather than the standard logger's stderr.
func (s *Shell) Logger() *slog.Logger {
	return s.logger
}

// refuseEchoMethods answers TRACE and TRACK with 405 before the request
// reaches subtree dispatch, so neither method reaches a registered
// subtree. TRACE echoes the request back to the
// client by definition, which hands anything an intermediary added,
// cookies and identity headers included, to whoever can make the
// browser send it; TRACK is IIS's undocumented synonym, and a server
// that refuses one while answering the other has refused nothing.
// net/http never answers either on its own.
//
// No Allow header accompanies the 405. The shell knows only that these
// two methods are refused everywhere, not which methods any subtree
// accepts, and an Allow header naming a guessed set
// would be a claim about handlers this step has never seen.
func refuseEchoMethods(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodTrace || r.Method == "TRACK" {
			markRefusal(r, http.StatusMethodNotAllowed)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// terminated dispatches a request to the registered subtree that
// claims its path and answers every other path with
// [http.NotFoundHandler]. The 404 is wired here explicitly rather than
// left to [http.ServeMux]'s own default, which answers the same status
// by coincidence of the mux's behavior rather than by the shell's
// decision. Subtrees register patterns that end on a segment boundary,
// so a path sharing a subtree's bytes without the boundary is claimed
// by none.
func (s *Shell) terminated() http.Handler {
	notFound := http.NotFoundHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h, pattern := s.mux.Handler(r)
		if pattern == "" {
			notFound.ServeHTTP(w, r)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// statusCapture is a http.ResponseWriter wrapper that records the
// status written, so logRefusal can decide whether the wrapped chain
// refused a request by observing the response it actually sent, never
// by reading the chain's construction. logRefusal wraps the whole
// chain, so statusCapture reaches every handler the shell dispatches
// to. strip.go keeps the websocket Upgrade handshake intact, so
// statusCapture stays transparent to streaming and upgraded
// connections: Unwrap, Flush, and Hijack exist so a streaming handler
// reaches the real writer, whether through [http.ResponseController]
// or by asserting [http.Flusher] or [http.Hijacker] itself.
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
// would silently downgrade every file [http.ServeContent] serves to a
// userspace copy, since ServeContent copies through [io.CopyN], which
// reaches [io.Copy] with a source that has no WriteTo.
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
// a handler rejected a malformed form. A 404 for a path no subtree
// claims is likewise no refusal, and emits no record.
//
// Positioned outside strip, normalization, and subtree dispatch alike,
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
			"path", boundedPath(r.URL.Path),
			"status", capture.status,
			"identity_spelling_present", identitySpellingPresent(r),
		)
	})
}

// maxLoggedPathBytes bounds the request path logRefusal writes. The
// refusal it records is the path bound's own refusal among others, and
// that one arrives with a path bounded by nothing but
// http.Server.MaxHeaderBytes: a 1 MiB request line is refused with 414
// and would otherwise write 1 MiB to the logger for every request an
// attacker sends, turning a cheap refusal into a log-volume amplifier.
// 256 bytes is enough to identify which surface was probed, which is
// the record's whole job.
const maxLoggedPathBytes = 256

// boundedPath returns path unchanged when it fits maxLoggedPathBytes,
// and otherwise its first maxLoggedPathBytes bytes with a marker naming
// the truncation and the full length. The length is kept because it is
// the one fact the truncated prefix destroys, and an oversized path is
// exactly the shape a reader wants the size of.
func boundedPath(path string) string {
	if len(path) <= maxLoggedPathBytes {
		return path
	}
	return fmt.Sprintf("%s...[truncated, %d bytes]", path[:maxLoggedPathBytes], len(path))
}

// identitySpellingPresent reports whether r's header carries any key
// the strip's group 1 removal matches, by the same [foldedKey]
// comparison against the same [identityKeys] set, so logRefusal asserts
// against what the strip removes rather than a second set written out
// here.
//
// It tests the key's presence in the map, not the value it holds. A
// header sent with an empty value is still a header the client sent,
// and the record exists to say whether one arrived at all.
func identitySpellingPresent(r *http.Request) bool {
	for k := range r.Header {
		if _, ok := identityKeys[foldedKey(k)]; ok {
			return true
		}
	}
	return false
}
