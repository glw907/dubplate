package serve

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
)

// refusedPort returns a loopback port the test binary reserved once
// (listened on 127.0.0.1:0, recorded, closed) and never listens on
// again, so a probe that dials it reaches no listener, never the host's
// 127.0.0.1:8080.
var refusedPort = sync.OnceValue(func() int {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("reserve a loopback port: %v", err))
	}
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok {
		panic(fmt.Sprintf("reserved listener at %v, want a TCP address", l.Addr()))
	}
	port := addr.Port
	if err := l.Close(); err != nil {
		panic(fmt.Sprintf("release the reserved port: %v", err))
	}
	return port
})

// rowEnviron returns the wall plus the entries, replacing any earlier
// entry for the same key.
func rowEnviron(entries ...string) []string {
	environ := wallEnviron()
	for _, entry := range entries {
		key, value, _ := strings.Cut(entry, "=")
		environ = withKey(environ, key, value)
	}
	return environ
}

// probeEnviron is rowEnviron aimed at the reserved port, where nothing
// listens.
func probeEnviron(entries ...string) []string {
	return rowEnviron(append([]string{"DUBPLATE_SERVE_PORT=" + strconv.Itoa(refusedPort())}, entries...)...)
}

// loadRowConfig loads environ against the registry the row tests share.
func loadRowConfig(t *testing.T, environ []string) *config.Config {
	t.Helper()
	cfg, err := config.Load(environ, configRegistry())
	if err != nil {
		t.Fatalf("fixture state: config.Load: %v", err)
	}
	return cfg
}

// listenerPort returns the port of a listener bound on 127.0.0.1.
func listenerPort(t *testing.T, l net.Listener) int {
	t.Helper()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok || !addr.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Fatalf("fixture state: listener at %v, want 127.0.0.1", l.Addr())
	}
	return addr.Port
}

// The two literals below are written out here rather than read from the
// production constants, so a changed clause fails these tests.
const (
	wantProbeTrailer       = "proves this server's strip, not a proxy's"
	wantAllTrustedUnproven = "forwarded and Connection-listed removal unproven: every local address is a trusted proxy"
)

// trustedLoopback is the same-host front's configuration: 127.0.0.1 is
// the one bind address and the one trusted proxy.
var trustedLoopback = []string{
	"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
	"DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1",
	"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
	"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1",
}

func TestClientIPStatus(t *testing.T) {
	proxyEnv := func(addrs string) []string {
		return []string{
			"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
			"DUBPLATE_TRUSTED_PROXY_ADDRS=" + addrs,
			"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
		}
	}
	tests := []struct {
		name       string
		entries    []string
		nilConfig  bool
		wantStatus string
		wantHas    []string
		wantFailed bool
	}{
		{
			name:       "Load refused the environment",
			nilConfig:  true,
			wantStatus: "not checked",
		},
		{
			name:       "the class off is direct",
			wantStatus: "direct: the peer must be the browser; a proxy in front must be configured as trusted",
		},
		{
			name:       "two addresses",
			entries:    proxyEnv("203.0.113.7, 2001:db8::1"),
			wantStatus: "trusted proxy: X-Forwarded-For from 2 addresses",
		},
		{
			name:       "a duplicate collapsing to one address",
			entries:    proxyEnv("203.0.113.7,203.0.113.7"),
			wantStatus: "trusted proxy: X-Forwarded-For from 1 address",
		},
		{
			name:       "a network instead of an address",
			entries:    proxyEnv("203.0.113.0/24"),
			wantHas:    []string{"invalid: ", "an address, not a network"},
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg *config.Config
			if !tt.nilConfig {
				cfg = loadRowConfig(t, rowEnviron(tt.entries...))
			}
			status, failed := clientIPStatus(cfg)
			if tt.wantStatus != "" && status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			for _, sub := range tt.wantHas {
				if !strings.Contains(status, sub) {
					t.Errorf("status = %q, want it to contain %q", status, sub)
				}
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

func TestBindAddrsStatus(t *testing.T) {
	ipNet := func(cidr string) net.Addr {
		ip, n, err := net.ParseCIDR(cidr)
		if err != nil {
			t.Fatal(err)
		}
		n.IP = ip
		return n
	}
	errNetlink := errors.New("route ip+net: netlinkrib: address family not supported by protocol")

	tests := []struct {
		name       string
		environ    []string
		addrs      []net.Addr
		err        error
		wantStatus string
		wantFailed bool
	}{
		{
			name:       "unset is the loopback default",
			addrs:      []net.Addr{ipNet("127.0.0.1/8")},
			wantStatus: "127.0.0.1 present",
		},
		{
			name:       "an address an interface holds",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=198.51.100.4"},
			addrs:      []net.Addr{ipNet("198.51.100.4/24")},
			wantStatus: "198.51.100.4 present",
		},
		{
			name:       "an IPAddr entry counts the same as an IPNet",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=198.51.100.4"},
			addrs:      []net.Addr{&net.IPAddr{IP: net.ParseIP("198.51.100.4")}},
			wantStatus: "198.51.100.4 present",
		},
		{
			name:       "an IPv4-mapped interface address matches its IPv4 form",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=198.51.100.4"},
			addrs:      []net.Addr{&net.IPNet{IP: net.ParseIP("::ffff:198.51.100.4"), Mask: net.CIDRMask(128, 128)}},
			wantStatus: "198.51.100.4 present",
		},
		{
			name:       "any loopback address counts present though no interface lists it",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.2"},
			addrs:      []net.Addr{ipNet("127.0.0.1/8")},
			wantStatus: "127.0.0.2 present",
		},
		{
			name:       "an absent address",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1"},
			addrs:      []net.Addr{ipNet("198.51.100.4/24")},
			wantStatus: "192.0.2.1 not on any interface",
			wantFailed: true,
		},
		{
			name:       "every absent address is named on the one row",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1,2001:db8::9,127.0.0.1"},
			addrs:      []net.Addr{ipNet("198.51.100.4/24")},
			wantStatus: "192.0.2.1 not on any interface, 2001:db8::9 not on any interface, 127.0.0.1 present",
			wantFailed: true,
		},
		{
			name:       "the netlink read failing names RestrictAddressFamilies",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1"},
			err:        errNetlink,
			wantStatus: "not checked: " + errNetlink.Error() + " (the unit's RestrictAddressFamilies may deny the netlink read)",
			wantFailed: true,
		},
		{
			name:       "an unspecified address is refused",
			environ:    []string{"DUBPLATE_SERVE_BIND_ADDRS=0.0.0.0"},
			wantStatus: `invalid: "0.0.0.0": the wildcard address binds every interface: invalid bind address`,
			wantFailed: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, failed := bindAddrsStatus(tt.environ, tt.addrs, tt.err)
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// cannedRoute serves reply for the probe route's path and a 404 for any
// other, on a port the returned environ names, and returns that
// environ: the probe row's input with entries applied.
func cannedRoute(t *testing.T, reply func(w http.ResponseWriter), entries ...string) []string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/dubplate/"+ProbeRouteName {
			http.NotFound(w, r)
			return
		}
		reply(w)
	}))
	t.Cleanup(srv.Close)
	port := listenerPort(t, srv.Listener)
	return rowEnviron(append([]string{"DUBPLATE_SERVE_PORT=" + strconv.Itoa(port)}, entries...)...)
}

// TestDoctorBoundaryProbe_ListenerAnswers puts a server on the port the
// probe dials that answers the route's path with a canned reply, one
// per case: each count the strip should have zeroed, an answer that is
// not 200, and a body that is not the counts.
func TestDoctorBoundaryProbe_ListenerAnswers(t *testing.T) {
	const trailer = "; " + wantProbeTrailer
	tests := []struct {
		name       string
		status     int
		body       string
		wantFailed bool
		wantStatus string
	}{
		{
			name:       "every count zero",
			status:     http.StatusOK,
			body:       `{"identity":0,"forwarded":0,"sentinel":0}`,
			wantStatus: "127.0.0.1 refused" + trailer,
		},
		{
			name:       "identity headers passed",
			status:     http.StatusOK,
			body:       `{"identity":2,"forwarded":0,"sentinel":0}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 passed 2 identity header(s)" + trailer,
		},
		{
			name:       "forwarded-address headers passed",
			status:     http.StatusOK,
			body:       `{"identity":0,"forwarded":3,"sentinel":0}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 passed 3 forwarded-address header(s)" + trailer,
		},
		{
			name:       "the sentinel passed",
			status:     http.StatusOK,
			body:       `{"identity":0,"forwarded":0,"sentinel":1}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 passed the Connection-listed sentinel" + trailer,
		},
		{
			name:       "every group passed",
			status:     http.StatusOK,
			body:       `{"identity":1,"forwarded":1,"sentinel":1}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 passed 1 identity header(s) and 1 forwarded-address header(s) and the Connection-listed sentinel" + trailer,
		},
		{
			name:       "a 404",
			status:     http.StatusNotFound,
			body:       "404 page not found\n",
			wantFailed: true,
			wantStatus: "127.0.0.1 unexpected answer: 404 Not Found",
		},
		{
			name:       "an unparseable body",
			status:     http.StatusOK,
			body:       "<html>",
			wantFailed: true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
		{
			name:       "an empty object",
			status:     http.StatusOK,
			body:       `{}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
		{
			name:       "JSON null",
			status:     http.StatusOK,
			body:       `null`,
			wantFailed: true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
		{
			name:       "an object missing two counts",
			status:     http.StatusOK,
			body:       `{"identity":0}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := cannedRoute(t, func(w http.ResponseWriter) {
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			})
			status, failed := boundaryProbeStatus(t.Context(), environ,loadRowConfig(t, environ))
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// wedgedPort puts a listener on a loopback port that completes the
// handshake and never answers: the kernel accepts the connection into
// the backlog whether or not Accept is called, so a dial succeeds and
// the read is what times out.
func wedgedPort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = l.Close() })
	return listenerPort(t, l)
}

// TestDoctorBoundaryProbe_WedgedListener points the probe at a listener
// that never answers: the row reads a timeout and fails, and the run
// ends within the response bound rather than hanging.
func TestDoctorBoundaryProbe_WedgedListener(t *testing.T) {
	environ := rowEnviron("DUBPLATE_SERVE_PORT=" + strconv.Itoa(wedgedPort(t)))
	cfg := loadRowConfig(t, environ)

	start := time.Now()
	status, failed := boundaryProbeStatus(t.Context(), environ,cfg)
	elapsed := time.Since(start)
	if want := "127.0.0.1 unexpected answer: timeout"; status != want {
		t.Errorf("status = %q, want %q", status, want)
	}
	if !failed {
		t.Error("failed = false, want a failure for a listener that never answers")
	}
	if limit := probeResponseBound + 2*time.Second; elapsed > limit {
		t.Errorf("the probe took %v, want under %v: the response bound did not end the wait", elapsed, limit)
	}
}

// TestboundaryProbeStatus_ParentDeadlineBoundsTheAnswer gives the
// probe a context that expires well inside the response bound: the
// answer's deadline is the context's, so the wait ends with it.
func TestBoundaryProbeStatus_ParentDeadlineBoundsTheAnswer(t *testing.T) {
	environ := rowEnviron("DUBPLATE_SERVE_PORT=" + strconv.Itoa(wedgedPort(t)))
	cfg := loadRowConfig(t, environ)
	ctx, cancel := context.WithTimeout(t.Context(), 100*time.Millisecond)
	defer cancel()

	start := time.Now()
	status, _ := boundaryProbeStatus(ctx, environ, cfg)
	elapsed := time.Since(start)
	if want := "127.0.0.1 unexpected answer: timeout"; status != want {
		t.Errorf("status = %q, want %q", status, want)
	}
	if limit := probeResponseBound - 500*time.Millisecond; elapsed > limit {
		t.Errorf("the probe took %v, want under %v: the caller's earlier deadline did not end the wait", elapsed, limit)
	}
}

// TestBoundaryProbeStatus_NoFailureOfItsOwn pins the states in
// which the row reports why it did not probe and leaves the failure to
// the row that owns the cause.
func TestBoundaryProbeStatus_NoFailureOfItsOwn(t *testing.T) {
	baseline := probeEnviron()
	cfg := loadRowConfig(t, baseline)
	// Load accepts a network as the trusted proxy value, so cfg is
	// non-nil and the refusal comes from the row's resolve.
	networkProxy := probeEnviron(
		"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
		"DUBPLATE_TRUSTED_PROXY_ADDRS=203.0.113.0/24",
		"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
	)
	networkProxyCfg := loadRowConfig(t, networkProxy)

	tests := []struct {
		name       string
		environ    []string
		cfg        *config.Config
		wantStatus string
	}{
		{
			name:       "an absent bind address, which the bind row fails",
			environ:    probeEnviron("DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1"),
			cfg:        cfg,
			wantStatus: "192.0.2.1 not probed (not on any interface)",
		},
		{
			name:       "an invalid bind list, which the bind row fails",
			environ:    probeEnviron("DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,,::1"),
			cfg:        cfg,
			wantStatus: "not probed",
		},
		{
			name:       "an invalid prefix, which the web row fails",
			environ:    probeEnviron("DUBPLATE_SERVE_PREFIX=dubplate"),
			cfg:        cfg,
			wantStatus: "not probed",
		},
		{
			name:       "a trusted proxy value that does not resolve, which the client ip row fails",
			environ:    networkProxy,
			cfg:        networkProxyCfg,
			wantStatus: "not probed",
		},
		{
			name:       "Load refused the environment",
			environ:    baseline,
			cfg:        nil,
			wantStatus: "not probed",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, failed := boundaryProbeStatus(t.Context(), tt.environ, tt.cfg)
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed {
				t.Error("failed = true, want the failure left to the row that owns the cause")
			}
		})
	}
}

// TestboundaryProbeStatus_MixedTrustedBind binds two loopback
// addresses, one of them a trusted proxy: both are dialed on the
// reserved port, where nothing listens, the trusted one is named as
// probed for identity only, and the row joins both parts without
// failing. One address is not a trusted proxy, so the row carries no
// all-trusted clause.
func TestBoundaryProbeStatus_MixedTrustedBind(t *testing.T) {
	environ := probeEnviron(
		"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
		"DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1",
		"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
		"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,127.0.0.2",
	)
	status, failed := boundaryProbeStatus(t.Context(), environ,loadRowConfig(t, environ))
	if want := "127.0.0.1 no listener (a trusted proxy, identity only), 127.0.0.2 no listener"; status != want {
		t.Errorf("status = %q, want %q", status, want)
	}
	if failed {
		t.Error("failed = true, want no failure")
	}
}

// TestDoctorBoundaryProbe_TrustedAnswers puts a canned listener at the
// trusted loopback address. A trusted peer's answer holds the identity
// count alone, so the row reads that count and nothing else, and a
// body without it is still not the route's answer.
func TestDoctorBoundaryProbe_TrustedAnswers(t *testing.T) {
	const tail = " (a trusted proxy, identity only); " + wantAllTrustedUnproven
	tests := []struct {
		name       string
		body       string
		wantFailed bool
		wantStatus string
	}{
		{
			name:       "identity refused",
			body:       `{"identity":0}`,
			wantStatus: "127.0.0.1 refused" + tail + "; " + wantProbeTrailer,
		},
		{
			name:       "identity headers passed",
			body:       `{"identity":2}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 passed 2 identity header(s)" + tail + "; " + wantProbeTrailer,
		},
		{
			name:       "no identity count",
			body:       `{"forwarded":0,"sentinel":0}`,
			wantFailed: true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body" + tail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := cannedRoute(t, func(w http.ResponseWriter) {
				_, _ = io.WriteString(w, tt.body)
			}, trustedLoopback...)
			status, failed := boundaryProbeStatus(t.Context(), environ,loadRowConfig(t, environ))
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// TestDoctorBoundaryProbe_SendsTheConfiguredHeader configures a
// client-address header outside the fixed family, from a proxy that is
// not a bind address, and records what the probe request carried: the
// configured header is the one [ClientIP] reads, so a probe that never
// sent it could not prove the strip removes it.
func TestDoctorBoundaryProbe_SendsTheConfiguredHeader(t *testing.T) {
	var got atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Store(r.Header.Get("X-House-Client-Address"))
		_, _ = io.WriteString(w, `{"identity":0,"forwarded":0,"sentinel":0}`)
	}))
	defer srv.Close()

	environ := rowEnviron(
		"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
		"DUBPLATE_TRUSTED_PROXY_ADDRS=203.0.113.7",
		"DUBPLATE_TRUSTED_PROXY_HEADER=X-House-Client-Address",
		"DUBPLATE_SERVE_PORT="+strconv.Itoa(listenerPort(t, srv.Listener)),
	)
	status, failed := boundaryProbeStatus(t.Context(), environ,loadRowConfig(t, environ))
	if failed {
		t.Fatalf("failed = true (status %q), want the strip's refusal to pass", status)
	}
	if v, _ := got.Load().(string); v != "probe" {
		t.Errorf("probe request X-House-Client-Address = %q, want %q", v, "probe")
	}
}
