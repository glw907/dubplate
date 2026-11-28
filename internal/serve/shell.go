package serve

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"unicode"
)

// Shell is the dubplate HTTP shell. A consumer registers subtrees under
// the owned prefix, then [Shell.Handler] composes the root handler.
type Shell struct {
	prefix   Prefix
	strip    strip
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

// NewShell builds a [Shell] that owns prefix and reads client addresses
// by proxy. logger receives the shell's refusal record; a nil logger
// falls back to [slog.Default].
func NewShell(prefix Prefix, proxy TrustedProxyConfig, logger *slog.Logger) *Shell {
	if logger == nil {
		logger = slog.Default()
	}
	return &Shell{
		prefix:   prefix,
		strip:    newStrip(proxy),
		origin:   newOriginCheck(),
		logger:   logger,
		mux:      http.NewServeMux(),
		subtrees: make(map[string]struct{}),
	}
}

// Register mounts h under the owned prefix as the subtree name: both
// prefix+"/"+name and everything under prefix+"/"+name+"/" reach h.
//
// Register panics, naming the subtree, on a name already registered, on
// a name that would widen the mount past one subtree, and on a call
// after [Shell.Handler]: each is a mistake in the binary's own wiring,
// so it fails at construction rather than at request time.
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
//     (normalize refuses a space as an unescaped byte), so a name
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

// Handler returns the root handler, composed in the order the package
// documentation lists. Registration ends when Handler is called.
func (s *Shell) Handler() http.Handler {
	s.handlerBuilt = true

	chain := []func(http.Handler) http.Handler{
		s.securityHeaders,
		s.logRefusal,
		s.warnForwardedFromLocalPeer,
		s.strip.wrap,
		normalize,
		refuseEchoMethods,
		s.origin.wrap,
	}
	h := s.dispatch()
	for _, step := range slices.Backward(chain) {
		h = step(h)
	}
	return h
}

// warnForwardedFromLocalPeer wraps next, the strip, and in direct mode
// logs one warning for the shell the first time a loopback or private
// peer sends a forwarded-address header. That request is the signature
// of a same-host or same-network front left in direct mode: every
// client it relays arrives from the front's own address, so the whole
// group shares one rate-limit key. Outside direct mode it returns next
// unchanged.
//
// It counts the headers the strip is about to remove, so it must run
// before the strip. The private-address test decides only whether to
// log, never whether to trust.
func (s *Shell) warnForwardedFromLocalPeer(next http.Handler) http.Handler {
	if s.strip.proxy.mode() != proxyDirect {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.noteForwardedFromLocalPeer(r)
		next.ServeHTTP(w, r)
	})
}

func (s *Shell) noteForwardedFromLocalPeer(r *http.Request) {
	peer, ok := peerAddr(r)
	if !ok || !peer.IsLoopback() && !peer.IsPrivate() {
		return
	}
	n := countPresent(r.Header, s.strip.forwarded)
	if n == 0 {
		return
	}
	s.forwardedWarning.Do(func() {
		s.logger.Warn("serve: a loopback or private peer sent a forwarded-address header in direct mode; a front on this host or network must be configured as a trusted proxy, or every client it relays shares its address and its rate limits",
			"peer", peerHost(r),
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

// refuseEchoMethods answers TRACE and TRACK with 405, so neither
// reaches a registered subtree. TRACE echoes the request back to the
// client, which hands anything an intermediary added, cookies and
// identity headers included, to whoever can make the browser send it;
// TRACK is IIS's synonym.
//
// No Allow header accompanies the 405: the shell does not know which
// methods any subtree accepts.
func refuseEchoMethods(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodTrace || r.Method == "TRACK" {
			refuse(w, r, http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// dispatch serves a request with the registered subtree that claims its
// path and answers every other path with [http.NotFoundHandler].
func (s *Shell) dispatch() http.Handler {
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

// refusalContextKey is the context key logRefusal files its per-request
// [refusalRecord] under.
type refusalContextKey struct{}

// refusalRecord is the per-request slot the shell's own refusing
// layers write their status to. A zero status means the shell refused
// nothing, whatever status a handler beyond it went on to write.
type refusalRecord struct {
	status int
}

// refuse answers r with status and records that the shell itself
// refused it, so [Shell.logRefusal] writes its record. The record is a
// no-op outside the composed chain.
func refuse(w http.ResponseWriter, r *http.Request, status int) {
	if rec, ok := r.Context().Value(refusalContextKey{}).(*refusalRecord); ok {
		rec.status = status
	}
	http.Error(w, http.StatusText(status), status)
}

// logRefusal wraps next and, once it returns, writes one log record for
// each request the shell itself refused, naming whether an identity
// spelling was present on it.
//
// The record is keyed on the refusal the refusing layer filed, never
// on the response status: a subtree answering 400 to its own bad input
// is not an ingress refusal, and a 404 for a path no subtree claims is
// none either.
//
// The header check runs after next returns and reads the header map
// next shared, so it sees the strip's removal when the strip ran
// before the refusal and the client's own headers when it did not.
func (s *Shell) logRefusal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rec := &refusalRecord{}
		r = r.WithContext(context.WithValue(r.Context(), refusalContextKey{}, rec))
		next.ServeHTTP(w, r)

		if rec.status == 0 {
			return
		}
		s.logger.Info("serve: refused request",
			"path", boundedPath(r.URL.Path),
			"status", rec.status,
			"identity_spelling_present", identitySpellingPresent(r),
		)
	})
}

// maxLoggedPathBytes bounds the request path logRefusal writes. A path
// is bounded otherwise only by the server's header ceiling, so an
// unbounded record would turn every cheap 414 refusal into a megabyte
// of log.
const maxLoggedPathBytes = 256

// boundedPath returns path unchanged when it fits maxLoggedPathBytes,
// and otherwise its first maxLoggedPathBytes bytes with a marker naming
// the truncation and the full length.
func boundedPath(path string) string {
	if len(path) <= maxLoggedPathBytes {
		return path
	}
	return fmt.Sprintf("%s...[truncated, %d bytes]", path[:maxLoggedPathBytes], len(path))
}

// identitySpellingPresent reports whether r's header holds any key the
// strip's identity removal matches, an empty value included.
func identitySpellingPresent(r *http.Request) bool {
	for k := range r.Header {
		if _, ok := identityKeys[foldedKey(k)]; ok {
			return true
		}
	}
	return false
}
