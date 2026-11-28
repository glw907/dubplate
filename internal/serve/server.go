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
	readHeaderTimeout = 10 * time.Second

	// readTimeout closes a connection whose client trickles its body a
	// byte at a time, which no other timeout reaches. It stays generous
	// for the login and magic-link POST bodies under everyday latency.
	readTimeout = 30 * time.Second

	idleTimeout = 120 * time.Second

	// maxHeaderBytes bounds the header memory an unauthenticated client
	// can make the server hold.
	maxHeaderBytes = 1 << 20 // 1 MiB

	// shutdownGrace bounds Run's drain once ctx is done. The supervising
	// unit's stop timeout is the real backstop, so it need only be
	// generous.
	shutdownGrace = 25 * time.Second

	// defaultWriteTimeout bounds every ordinary response. An SSE stream
	// or an audio response that runs longer extends its own deadline
	// through [http.ResponseController.SetWriteDeadline] before this one
	// expires.
	defaultWriteTimeout = 30 * time.Second
)

// newHTTPServer builds the server Run serves handler with, without an
// Addr or a listener. ErrorLog sends net/http's protocol errors (a bad
// request line, a handler panic) to logger at warn level rather than to
// the standard logger's stderr, where the shell's structured sink never
// sees them.
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
// the wildcard address, is what refuses a connection to any address not
// named in it. A failure partway closes every listener already opened.
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

// Run opens the freebind listener set named by addrs and port, serves
// handler across all of them until ctx is done, then drains within
// shutdownGrace before returning.
//
// logger receives net/http's protocol errors. A caller passes
// [Shell.Logger] so those errors and the shell's refusal record share
// one sink; a nil logger falls back to [slog.Default], as in
// [NewShell].
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

// serveUntilDone runs srv across every listener until ctx is done or
// one Serve fails, then shuts srv down under a context derived with
// [context.WithoutCancel], since one derived from the canceled ctx would
// never wait for an in-flight request.
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

	// Shut down on the failure path too: one failed Serve leaves the
	// other listeners accepting, a half-dead server bound to its
	// addresses.
	shutdownCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), shutdownGrace)
	defer cancel()
	shutdownErr := srv.Shutdown(shutdownCtx)

	// Shutdown does not wait for a Serve that had not yet registered its
	// listener, which closes it only as it returns; waiting here is what
	// frees the addresses on return.
	serving.Wait()

	if serveErr != nil {
		return fmt.Errorf("serve: %w", serveErr)
	}
	if shutdownErr != nil {
		return fmt.Errorf("serve: shutdown: %w", shutdownErr)
	}
	return nil
}
