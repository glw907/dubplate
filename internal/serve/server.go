package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"
)

// Named timeouts and MaxHeaderBytes for the http.Server C6 constructs
// (R-H, R-V). Each carries the reason it is sized the way it is, so the
// number is a decision recorded once rather than an unexplained literal
// repeated at every call site.
const (
	// readHeaderTimeout bounds how long a client has to finish sending
	// request headers before the connection is abandoned, independent
	// of readTimeout below (net/http applies it before the request line
	// is even fully parsed).
	readHeaderTimeout = 10 * time.Second

	// readTimeout (R-V) bounds a request body trickled a byte at a
	// time, the case none of the other three timeouts reach: a slow
	// client can hold a connection open indefinitely by draining its
	// own send buffer one byte per round trip, and this is what closes
	// it. It is the constant the salvage server carried at
	// _salvage/internal/serve/serve.go:600 that revision 1 of this plan
	// dropped by describing that server as setting a write timeout
	// alone. It must stay generous enough for 5b's two bodies that
	// trickle in over this connection: the login-request POST and the
	// magic-link-consumption POST, neither of which streams, but both
	// of which are ordinary request bodies this timeout must not cut
	// short under everyday latency.
	readTimeout = 30 * time.Second

	// idleTimeout bounds how long a keep-alive connection may sit idle
	// between requests, so a client that opens a connection and never
	// sends another request releases the file descriptor instead of
	// holding it until the OS's own (much longer) default.
	idleTimeout = 120 * time.Second

	// maxHeaderBytes (R-H) is set explicitly rather than inherited from
	// net/http's own package default, so the ceiling on header-driven
	// memory an unauthenticated client can force the server to hold is
	// a decision this package states rather than a value read from
	// http.DefaultMaxHeaderBytes.
	maxHeaderBytes = 1 << 20 // 1 MiB

	// shutdownGrace bounds Run's own drain wait for an in-flight
	// request once ctx is done. A supervising unit's own stop timeout
	// (5d's concern, not this package's) is the real backstop, so this
	// only needs to be generous, not exact.
	shutdownGrace = 25 * time.Second

	// defaultWriteTimeout protects every ordinary response, the ones
	// that never call the extension mechanism below. An SSE stream or
	// an audio response can legitimately run past 30 seconds, and a
	// single fixed write timeout would kill both, which is the mistake
	// the salvage server made (_salvage/internal/serve/serve.go:601,
	// applied at :632). The mechanism that saves the streaming case is
	// http.ResponseController.SetWriteDeadline: a streaming handler can
	// extend or clear its own deadline before this default expires,
	// while every handler that never calls it stays bound by this
	// default.
	defaultWriteTimeout = 30 * time.Second
)

// newHTTPServer builds the http.Server C6 constructs over handler,
// carrying all four named timeouts above and the explicit
// MaxHeaderBytes. It sets no Addr and opens no listener: Run pairs it
// with the listener set Listen opens.
//
// writeTimeout is a parameter rather than a read of
// defaultWriteTimeout, and [Run] is the one production caller that
// supplies it, from that constant. The deadline-extension test needs a
// write timeout it can exhaust in milliseconds, and passing one in is
// what lets every timeout stay a constant: a package-level var a test
// assigns to is shared mutable state between tests that may run in any
// order.
//
// ErrorLog is wired to logger through [slog.NewLogLogger] at warn
// level. net/http writes its own protocol errors (a bad request line, a
// TLS handshake failure, a handler panic) to that logger, and the
// zero value sends them to the standard logger's stderr, where the
// shell's own structured sink never sees them. Warn is the level: these
// are conditions an operator wants surfaced, and net/http emits one per
// malformed connection rather than per request.
func newHTTPServer(handler http.Handler, logger *slog.Logger, writeTimeout time.Duration) *http.Server {
	return &http.Server{
		Handler:           handler,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
}

// Listen opens one TCP listener per address in addrs, all on port, each
// carrying the freebindControl hook. Binding exactly this list, never
// the all-interfaces wildcard address (C8), is what refuses a
// connection to any address not named in it, since nothing listens
// there. A failure partway through the set closes every listener
// already opened before Listen returns, so a partial bind leaks no
// file descriptor. The listeners Listen returns carry their
// actually-bound addresses, readable through Addr() (R-G), for 5d's
// doctor row to report later.
func Listen(ctx context.Context, addrs []string, port int) ([]net.Listener, error) {
	lc := net.ListenConfig{Control: freebindControl}
	listeners := make([]net.Listener, 0, len(addrs))
	for _, addr := range addrs {
		l, err := lc.Listen(ctx, "tcp", net.JoinHostPort(addr, strconv.Itoa(port)))
		if err != nil {
			for _, opened := range listeners {
				_ = opened.Close()
			}
			return nil, fmt.Errorf("serve: listen %s: %w", addr, err)
		}
		listeners = append(listeners, l)
	}
	return listeners, nil
}

// Run opens the freebind listener set named by addrs and port, then
// serves handler across all of them (C6) until ctx is done, then drains
// within shutdownGrace before returning. ctx is Run's first parameter:
// Task 7 calls Run from the serve verb, passing the resolved
// configuration values and Task 5's composed root handler.
//
// logger receives net/http's own protocol errors. A caller passes
// [Shell.Logger] so those errors and the shell's refusal record share
// one sink; a nil logger falls back to [slog.Default], the same way
// [NewShell] does.
func Run(ctx context.Context, logger *slog.Logger, handler http.Handler, addrs []string, port int) error {
	if logger == nil {
		logger = slog.Default()
	}
	listeners, err := Listen(ctx, addrs, port)
	if err != nil {
		return err
	}
	return serveUntilDone(ctx, newHTTPServer(handler, logger, defaultWriteTimeout), listeners)
}

// serveUntilDone runs srv across every listener until ctx is done, then
// calls http.Server.Shutdown with a fresh shutdownGrace-bounded context
// derived via context.WithoutCancel: deriving directly from the
// already-canceled ctx would hand Shutdown a context that is already
// done, so the drain would never wait for an in-flight request. Split
// out from Run so a test can exercise the drain behavior against a
// minimal handler and a listener set it already knows the address of,
// without going through Run's own address resolution.
func serveUntilDone(ctx context.Context, srv *http.Server, listeners []net.Listener) error {
	errCh := make(chan error, len(listeners))
	for _, l := range listeners {
		go func(l net.Listener) {
			err := srv.Serve(l)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
				return
			}
			errCh <- nil
		}(l)
	}

	var serveErr error
	select {
	case <-ctx.Done():
	case serveErr = <-errCh:
	}

	// The shutdown runs on the failure path too, never only on the
	// ctx.Done one. One listener's Serve failing leaves every other
	// listener in the set still accepting, and returning here without
	// shutting down would leave a half-dead server bound to its
	// addresses for as long as the process lived. Shutdown closes the
	// listeners it is still serving, and net/http's own Serve closes
	// the failed one as it returns, so the whole set is released.
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownGrace)
	defer cancel()
	shutdownErr := srv.Shutdown(shutdownCtx)

	if serveErr != nil {
		return fmt.Errorf("serve: %w", serveErr)
	}
	if shutdownErr != nil {
		return fmt.Errorf("serve: shutdown: %w", shutdownErr)
	}
	return nil
}
