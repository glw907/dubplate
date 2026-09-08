package serve

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"syscall"
	"testing"
	"time"
)

// --- the http.Server construction (C6, R-H, R-V) ---------------------

// TestNewHTTPServerTimeouts asserts every named timeout constant and
// MaxHeaderBytes lands in the http.Server literal newHTTPServer builds.
// The read timeout is also checked against a literal 30 seconds, R-V's
// own value, independent of the constant it is wired from.
func TestNewHTTPServerTimeouts(t *testing.T) {
	srv := newHTTPServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}), slog.Default(), defaultWriteTimeout)

	if srv.ReadHeaderTimeout != readHeaderTimeout {
		t.Errorf("ReadHeaderTimeout = %v, want the readHeaderTimeout constant %v", srv.ReadHeaderTimeout, readHeaderTimeout)
	}
	if srv.ReadTimeout != 30*time.Second {
		t.Errorf("ReadTimeout = %v, want R-V's 30s", srv.ReadTimeout)
	}
	if srv.ReadTimeout != readTimeout {
		t.Errorf("ReadTimeout = %v, want the readTimeout constant %v", srv.ReadTimeout, readTimeout)
	}
	if srv.WriteTimeout != 30*time.Second {
		t.Errorf("WriteTimeout = %v, want the 30s defaultWriteTimeout Run passes", srv.WriteTimeout)
	}
	// A second construction with a different write timeout proves the
	// field is wired from the parameter rather than from a constant
	// read inside newHTTPServer, which the row above alone could not
	// tell apart.
	if short := newHTTPServer(http.NotFoundHandler(), slog.Default(), 5*time.Millisecond); short.WriteTimeout != 5*time.Millisecond {
		t.Errorf("WriteTimeout = %v for a 5ms argument, want the argument to reach the field", short.WriteTimeout)
	}
	if srv.IdleTimeout != idleTimeout {
		t.Errorf("IdleTimeout = %v, want the idleTimeout constant %v", srv.IdleTimeout, idleTimeout)
	}
	if srv.MaxHeaderBytes != maxHeaderBytes {
		t.Errorf("MaxHeaderBytes = %d, want the maxHeaderBytes constant %d", srv.MaxHeaderBytes, maxHeaderBytes)
	}
}

// --- the write-deadline extension mechanism (spec:975-981) -----------

// TestWriteDeadlineExtension proves the write timeout newHTTPServer
// wires is real, and that http.ResponseController.SetWriteDeadline,
// exposed to the fallthrough handler by newHTTPServer's ordinary
// construction, lets a handler survive past it. It passes a short
// write timeout through the constructor, so both rows exercise
// newHTTPServer's actual wiring path (proving a dropped WriteTimeout
// field would let the no-extension row succeed instead of fail) at a
// duration the gate can run in a second rather than thirty.
//
// This is a timing test. The margins are wide by a large multiple
// rather than tight, so a loaded machine's scheduling delay cannot
// carry either row across its own threshold: the handler sleeps five
// times the timeout it must exceed, and the extension it sets is
// twenty times that timeout and four times the sleep.
func TestWriteDeadlineExtension(t *testing.T) {
	const testWriteTimeout = 100 * time.Millisecond
	const sleepPastTimeout = 500 * time.Millisecond
	const extendedDeadline = 2 * time.Second

	tests := []struct {
		name    string
		extend  bool
		wantErr bool
	}{
		{
			name:    "extends past the default write deadline and succeeds",
			extend:  true,
			wantErr: false,
		},
		{
			name:    "does not extend, the default write deadline cuts it",
			extend:  false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.extend {
					if err := http.NewResponseController(w).SetWriteDeadline(time.Now().Add(extendedDeadline)); err != nil {
						t.Errorf("SetWriteDeadline: %v", err)
					}
				}
				time.Sleep(sleepPastTimeout)
				_, _ = w.Write([]byte("ok"))
			})

			srv := httptest.NewUnstartedServer(handler)
			srv.Config = newHTTPServer(handler, slog.Default(), testWriteTimeout)
			srv.Start()
			defer srv.Close()

			start := time.Now()
			resp, err := srv.Client().Get(srv.URL)
			elapsed := time.Since(start)

			if tt.wantErr {
				if err == nil {
					body, _ := io.ReadAll(resp.Body)
					_ = resp.Body.Close()
					t.Fatalf("Get succeeded without extension, body %q, want a write-deadline failure (elapsed %s)", body, elapsed)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get failed despite extension: %v (elapsed %s)", err, elapsed)
			}
			body, readErr := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			if readErr != nil || string(body) != "ok" {
				t.Fatalf("body = %q, read err %v, want \"ok\"", body, readErr)
			}
		})
	}
}

// --- the IP_FREEBIND listener set (R20, R-M) --------------------------

// TestListenBindsExactAddresses asserts Listen binds exactly the
// addresses named, none substituted for the all-interfaces wildcard
// address (C8), and that every bound listener's address is readable
// afterward (R-G). One row binds two IPv4 loopback addresses in a
// single call, proving a second address bound alongside the first; the
// other binds an IPv6 loopback address, since the hook's socket level
// (IPPROTO_IP) is the same for both families and this is the assertion
// that proves it.
func TestListenBindsExactAddresses(t *testing.T) {
	tests := []struct {
		name  string
		addrs []string
	}{
		{name: "two IPv4 loopback addresses", addrs: []string{"127.0.0.1", "127.0.0.2"}},
		{name: "an IPv6 loopback address", addrs: []string{"::1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			listeners, err := Listen(context.Background(), tt.addrs, 0)
			if err != nil {
				t.Fatalf("Listen(%v, 0): %v", tt.addrs, err)
			}
			defer func() {
				for _, l := range listeners {
					_ = l.Close()
				}
			}()

			if len(listeners) != len(tt.addrs) {
				t.Fatalf("len(listeners) = %d, want %d (one per address, the second bound alongside the first)", len(listeners), len(tt.addrs))
			}
			for i, l := range listeners {
				addr := l.Addr()
				if addr == nil {
					t.Fatalf("listeners[%d].Addr() = nil, want the actually-bound address", i)
				}
				host, _, err := net.SplitHostPort(addr.String())
				if err != nil {
					t.Fatalf("listeners[%d].Addr() = %q, want a host:port pair: %v", i, addr.String(), err)
				}
				// The bound host must equal the address asked for, not
				// merely be some address that is not the wildcard: a
				// listener bound to a different loopback address than
				// the one configured would pass a wildcard-only check
				// while accepting traffic nobody asked it to.
				bound := net.ParseIP(host)
				if bound == nil {
					t.Fatalf("listeners[%d] bound host %q, want an IP address", i, host)
				}
				if want := net.ParseIP(tt.addrs[i]); !bound.Equal(want) {
					t.Fatalf("listeners[%d] bound host %q, want exactly the requested %q", i, host, tt.addrs[i])
				}
			}
		})
	}
}

// TestListenPartialFailureClosesOpened proves a failure partway through
// the address list closes every listener already opened, so a partial
// bind leaks nothing. It pre-occupies a fixed port on 127.0.0.2, then
// asks Listen for that same port on 127.0.0.1 (which succeeds first)
// and 127.0.0.2 (which collides and fails). 127.0.0.2 is not the
// documentation-reserved address the Control-hook test below uses.
func TestListenPartialFailureClosesOpened(t *testing.T) {
	occupied, err := net.Listen("tcp", "127.0.0.2:0")
	if err != nil {
		t.Fatalf("pre-occupy 127.0.0.2: %v", err)
	}
	defer func() { _ = occupied.Close() }()
	_, portStr, err := net.SplitHostPort(occupied.Addr().String())
	if err != nil {
		t.Fatalf("split pre-occupied address %q: %v", occupied.Addr().String(), err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		t.Fatalf("parse pre-occupied port %q: %v", portStr, err)
	}

	_, err = Listen(context.Background(), []string{"127.0.0.1", "127.0.0.2"}, port)
	if err == nil {
		t.Fatalf("Listen succeeded, want an error binding the already-occupied 127.0.0.2:%d", port)
	}

	// If the first listener (127.0.0.1:port) were left open, rebinding
	// it here would fail with EADDRINUSE.
	leak, err := net.Listen("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		t.Fatalf("127.0.0.1:%d still bound after Listen's failure, want it closed: %v", port, err)
	}
	_ = leak.Close()
}

// --- the Control hook, behaviorally (R20) -----------------------------

const freebindTestAddr = "192.0.2.77" // RFC 5737 TEST-NET-3, on no interface

// TestControlHookOpensAnUnconfiguredAddress asserts the boot property:
// a listen on freebindTestAddr succeeds with freebindControl and fails
// with EADDRNOTAVAIL from a bare net.ListenConfig. The socket option is
// then read back with syscall.GetsockoptInt as a second, cheaper
// assertion.
func TestControlHookOpensAnUnconfiguredAddress(t *testing.T) {
	t.Run("with the hook, the address binds", func(t *testing.T) {
		listeners, err := Listen(context.Background(), []string{freebindTestAddr}, 0)
		if err != nil {
			t.Fatalf("Listen(%q, 0) with the hook: %v, want success (the boot property)", freebindTestAddr, err)
		}
		defer func() { _ = listeners[0].Close() }()

		tcpListener, ok := listeners[0].(*net.TCPListener)
		if !ok {
			t.Fatalf("listener type = %T, want *net.TCPListener", listeners[0])
		}
		raw, err := tcpListener.SyscallConn()
		if err != nil {
			t.Fatalf("SyscallConn: %v", err)
		}
		var (
			optValue int
			optErr   error
		)
		if err := raw.Control(func(fd uintptr) {
			optValue, optErr = syscall.GetsockoptInt(int(fd), syscall.IPPROTO_IP, syscall.IP_FREEBIND)
		}); err != nil {
			t.Fatalf("raw.Control: %v", err)
		}
		if optErr != nil {
			t.Fatalf("GetsockoptInt(IPPROTO_IP, IP_FREEBIND): %v", optErr)
		}
		if optValue == 0 {
			t.Fatalf("IP_FREEBIND read back as 0, want nonzero (the option freebindControl set)")
		}
	})

	t.Run("without the hook, the same address fails with EADDRNOTAVAIL", func(t *testing.T) {
		lc := net.ListenConfig{}
		_, err := lc.Listen(context.Background(), "tcp", net.JoinHostPort(freebindTestAddr, "0"))
		if err == nil {
			t.Fatalf("Listen(%q, 0) without the hook succeeded, want EADDRNOTAVAIL (the address is on no interface)", freebindTestAddr)
		}
		if !errors.Is(err, syscall.EADDRNOTAVAIL) {
			t.Fatalf("Listen error = %v, want errors.Is(err, syscall.EADDRNOTAVAIL)", err)
		}
	})
}

// --- Run: ctx first, and the drain (C6) -------------------------------

// TestRunTakesCtxFirstAndReturnsPromptlyWhenAlreadyDone calls Run with an
// already-canceled context, positionally, ctx first: Run opens the
// listener set, observes ctx.Done immediately, and shuts down without
// ever accepting a connection. Beyond that behavior, this call site is
// the mutation backstop for Run's own signature: taking ctx off Run and
// reading a package context instead fails to compile here.
func TestRunTakesCtxFirstAndReturnsPromptlyWhenAlreadyDone(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	if err := Run(ctx, slog.Default(), handler, []string{"127.0.0.1"}, 0); err != nil {
		t.Fatalf("Run(already-canceled ctx, ...) = %v, want nil", err)
	}
}

// TestRunDrainsInFlightRequestAfterCtxDone proves Run's shutdown path
// waits for an in-flight request instead of cutting it when ctx is
// already done, which is only possible if the shutdown timeout context
// is derived with context.WithoutCancel: deriving straight from the
// already-canceled ctx would hand Shutdown a context already done, and
// Shutdown returns immediately once no connection remains active
// without ever waiting, so a dropped WithoutCancel would still pass a
// test with no in-flight connection. This one keeps a handler blocked
// mid-request across the cancel to rule that out.
func TestRunDrainsInFlightRequestAfterCtxDone(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
		w.WriteHeader(http.StatusOK)
	})

	ctx, cancel := context.WithCancel(context.Background())
	listeners, err := Listen(ctx, []string{"127.0.0.1"}, 0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	addr := listeners[0].Addr().String()

	runErr := make(chan error, 1)
	go func() {
		runErr <- serveUntilDone(ctx, newHTTPServer(handler, slog.Default(), defaultWriteTimeout), listeners)
	}()

	respErr := make(chan error, 1)
	go func() {
		resp, err := http.Get("http://" + addr) //nolint:noctx // deliberately outlives the request's own ctx
		if err != nil {
			respErr <- err
			return
		}
		_ = resp.Body.Close()
		respErr <- nil
	}()

	<-started
	cancel() // ctx.Done fires while the request is still in flight

	select {
	case err := <-respErr:
		t.Fatalf("in-flight request returned before the handler released it: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(release)

	if err := <-respErr; err != nil {
		t.Fatalf("in-flight request failed during drain: %v", err)
	}
	if err := <-runErr; err != nil {
		t.Fatalf("serveUntilDone returned %v, want nil", err)
	}
}

// TestServeUntilDoneShutsDownWhenOneServeFails: one listener's Serve
// returning an error must not leave the rest of the set accepting.
// serveUntilDone shuts the server down on that path too, so every
// address it bound is free the moment it returns; returning the error
// without a shutdown would leave a half-dead server holding its
// addresses for the life of the process.
//
// The failure is induced by closing one listener before serveUntilDone
// runs: net/http's Serve returns net.ErrClosed for it, which is not
// http.ErrServerClosed and so reaches the error path.
func TestServeUntilDoneShutsDownWhenOneServeFails(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	listeners, err := Listen(context.Background(), []string{"127.0.0.1", "127.0.0.2"}, 0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	survivor := listeners[1].Addr().String()
	if err := listeners[0].Close(); err != nil {
		t.Fatalf("close the listener whose Serve must fail: %v", err)
	}

	// Not an already-canceled context: the return must come from the
	// failing Serve, never from ctx.Done racing it.
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	err = serveUntilDone(ctx, newHTTPServer(handler, slog.Default(), defaultWriteTimeout), listeners)
	if err == nil {
		t.Fatal("serveUntilDone returned nil, want the failing listener's error")
	}
	if !errors.Is(err, net.ErrClosed) {
		t.Fatalf("serveUntilDone error = %v, want it to wrap net.ErrClosed", err)
	}

	// If the surviving listener were still accepting, rebinding its
	// address here would fail with EADDRINUSE.
	rebound, err := net.Listen("tcp", survivor)
	if err != nil {
		t.Fatalf("%s still bound after the failure path, want the shutdown to have released it: %v", survivor, err)
	}
	_ = rebound.Close()
}

// levelCapturingHandler records each record's level and message, the
// two facts the ErrorLog assertion turns on.
type levelCapturingHandler struct {
	levels   []slog.Level
	messages []string
}

func (h *levelCapturingHandler) Enabled(context.Context, slog.Level) bool { return true }

func (h *levelCapturingHandler) Handle(_ context.Context, r slog.Record) error {
	h.levels = append(h.levels, r.Level)
	h.messages = append(h.messages, r.Message)
	return nil
}

func (h *levelCapturingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *levelCapturingHandler) WithGroup(string) slog.Handler      { return h }

// TestNewHTTPServerErrorLogReachesTheSlogSink proves net/http's own
// protocol errors land in the shell's logger rather than the standard
// logger's stderr. The zero-value ErrorLog sends a malformed request
// line, a TLS handshake failure, and a handler panic to log.Default,
// where the operator's structured sink never sees them.
func TestNewHTTPServerErrorLogReachesTheSlogSink(t *testing.T) {
	capture := &levelCapturingHandler{}
	srv := newHTTPServer(http.NotFoundHandler(), slog.New(capture), defaultWriteTimeout)

	if srv.ErrorLog == nil {
		t.Fatal("srv.ErrorLog = nil, want a logger writing to the shell's own sink")
	}
	srv.ErrorLog.Print("http: bad request line")

	if len(capture.levels) != 1 {
		t.Fatalf("records = %d, want 1", len(capture.levels))
	}
	if capture.levels[0] != slog.LevelWarn {
		t.Errorf("record level = %v, want %v", capture.levels[0], slog.LevelWarn)
	}
	if capture.messages[0] != "http: bad request line" {
		t.Errorf("record message = %q, want the text net/http wrote", capture.messages[0])
	}
}
