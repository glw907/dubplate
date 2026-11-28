package serve

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"sync"
	"time"
)

// Timeouts and the header ceiling for the server [Run] builds.
const (
	// readHeaderTimeout bounds how long a client has to finish sending
	// request headers before the connection is abandoned, independent
	// of readTimeout below (net/http applies it before the request line
	// is even fully parsed).
	readHeaderTimeout = 10 * time.Second

	// readTimeout bounds a request body trickled a byte at a
	// time, the case none of the other three timeouts reach: a slow
	// client can hold a connection open indefinitely by draining its
	// own send buffer one byte per round trip, and this is what closes
	// it. It must stay generous enough for the two bodies that
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

	// maxHeaderBytes bounds the header memory an unauthenticated client
	// can make the server hold.
	maxHeaderBytes = 1 << 20 // 1 MiB

	// shutdownGrace bounds Run's drain wait for an in-flight request
	// once ctx is done. The supervising unit's stop timeout is the real
	// backstop, so this only needs to be generous, not exact.
	shutdownGrace = 25 * time.Second

	// defaultWriteTimeout bounds every ordinary response. An SSE stream
	// or an audio response can legitimately run longer: such a handler
	// extends or clears its own deadline through
	// [http.ResponseController.SetWriteDeadline] before this one
	// expires.
	defaultWriteTimeout = 30 * time.Second
)

// newHTTPServer builds the server Run serves handler with. It sets no
// Addr and opens no listener: Run pairs it with the listener set listen
// opens.
//
// ErrorLog is wired to logger through [slog.NewLogLogger] at warn
// level. net/http writes its own protocol errors (a bad request line, a
// TLS handshake failure, a handler panic) to that logger, and the
// zero value sends them to the standard logger's stderr, where the
// shell's own structured sink never sees them. Warn is the level: these
// are conditions an operator wants surfaced, and net/http emits one per
// malformed connection rather than per request.
func newHTTPServer(handler http.Handler, logger *slog.Logger) *http.Server {
	return &http.Server{
		Handler:           handler,
		ErrorLog:          slog.NewLogLogger(logger.Handler(), slog.LevelWarn),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      defaultWriteTimeout,
		IdleTimeout:       idleTimeout,
		MaxHeaderBytes:    maxHeaderBytes,
	}
}

// listen opens one TCP listener per address in addrs, all on port, each
// carrying the freebindControl hook. Binding exactly this list, never
// the all-interfaces wildcard address, is what refuses a
// connection to any address not named in it, since nothing listens
// there. A failure partway through the set closes every listener
// already opened before listen returns, so a partial bind leaks no
// file descriptor. The listeners listen returns carry their
// actually-bound addresses, readable through Addr().
func listen(ctx context.Context, addrs []netip.Addr, port int) ([]net.Listener, error) {
	lc := net.ListenConfig{Control: freebindControl}
	listeners := make([]net.Listener, 0, len(addrs))
	for _, addr := range addrs {
		l, err := lc.Listen(ctx, "tcp", net.JoinHostPort(addr.String(), strconv.Itoa(port)))
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
// serves handler across all of them until ctx is done, then drains
// within shutdownGrace before returning. The serve verb calls Run with
// the resolved configuration values and its composed root handler.
//
// logger receives net/http's own protocol errors. A caller passes
// [Shell.Logger] so those errors and the shell's refusal record share
// one sink; a nil logger falls back to [slog.Default], the same way
// [NewShell] does.
func Run(ctx context.Context, logger *slog.Logger, handler http.Handler, addrs []netip.Addr, port int) error {
	if logger == nil {
		logger = slog.Default()
	}
	listeners, err := listen(ctx, addrs, port)
	if err != nil {
		return err
	}
	return serveUntilDone(ctx, newHTTPServer(handler, logger), listeners)
}

// serveUntilDone runs srv across every listener until ctx is done, then
// calls http.Server.Shutdown with a fresh shutdownGrace-bounded context
// derived via context.WithoutCancel: deriving directly from the
// already-canceled ctx would hand Shutdown a context that is already
// done, so the drain would never wait for an in-flight request.
func serveUntilDone(ctx context.Context, srv *http.Server, listeners []net.Listener) error {
	errCh := make(chan error, len(listeners))
	var serving sync.WaitGroup
	for _, l := range listeners {
		serving.Go(func() {
			err := srv.Serve(l)
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				errCh <- err
				return
			}
			errCh <- nil
		})
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

	// Shutdown does not wait for a Serve that had not yet registered its
	// listener: that Serve closes the listener only as it returns. Waiting
	// for every Serve here is what makes the addresses free on return.
	serving.Wait()

	if serveErr != nil {
		return fmt.Errorf("serve: %w", serveErr)
	}
	if shutdownErr != nil {
		return fmt.Errorf("serve: shutdown: %w", shutdownErr)
	}
	return nil
}
