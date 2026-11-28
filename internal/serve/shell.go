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

	// forwardedWarning makes the direct-mode warning once per shell,
	// which is once per process.
	forwardedWarning sync.Once

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
// it may. Each refused shape fails silently rather than visibly, which
// is what makes a panic the right answer: an empty name mounts the
// whole prefix, a "/" mounts a deeper path than written, a brace is a
// mux wildcard, and whitespace, which normalize refuses in a request
// path, mounts a pattern nothing can reach.
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

// warnForwardedFromLocalPeer, in direct mode, logs one warning the
// first time a loopback or private peer sends a forwarded-address
// header: the signature of a nearby front left untrusted, which puts
// every client it relays on one rate-limit key. It counts headers the
// strip removes, so it must run before the strip. The private-address
// test decides only whether to log, never whether to trust.
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
// [slog.Default] when NewShell was given none.
func (s *Shell) Logger() *slog.Logger {
	return s.logger
}

// refuseEchoMethods answers TRACE and its IIS synonym TRACK with 405,
// without an Allow header, since the shell does not know any subtree's
// methods. TRACE echoes the request, handing anything an intermediary
// added, cookies and identity headers included, to whoever can make the
// browser send it.
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

type refusalContextKey struct{}

// refusalRecord is the per-request slot the shell's refusing layers
// write their status to; zero means the shell refused nothing.
type refusalRecord struct {
	status int
}

// refuse answers r with status and records that the shell refused it,
// so [Shell.logRefusal] writes its record.
func refuse(w http.ResponseWriter, r *http.Request, status int) {
	if rec, ok := r.Context().Value(refusalContextKey{}).(*refusalRecord); ok {
		rec.status = status
	}
	http.Error(w, http.StatusText(status), status)
}

// logRefusal writes one log record for each request the shell itself
// refused, naming whether an identity spelling was present. It keys on
// the filed refusal, never the response status, which a subtree's own
// 400 or an unclaimed path's 404 would also produce. The header check
// runs after next returns, so it sees the strip's removal when the
// strip ran before the refusal.
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

// maxLoggedPathBytes bounds the request path logRefusal writes, which
// would otherwise turn every cheap 414 refusal into a megabyte of log.
const maxLoggedPathBytes = 256

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
