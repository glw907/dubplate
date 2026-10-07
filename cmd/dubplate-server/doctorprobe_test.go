package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// doctorRows runs doctor over environ against the production registry
// and returns its rows keyed by the first column, the count of lines
// printed, and Execute's error. Every row the tests below read is
// unique by key.
func doctorRows(t *testing.T, environ []string) (rows map[string]string, lines int, err error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := newDoctorCmd(environ, productionRegistry(), &stdout, &stderr)
	cmd.SetOut(&stdout)
	err = cmd.Execute()
	rows = make(map[string]string)
	for line := range strings.SplitSeq(strings.TrimRight(stdout.String(), "\n"), "\n") {
		lines++
		if len(line) > 33 {
			rows[strings.TrimRight(line[:32], " ")] = line[33:]
		}
	}
	return rows, lines, err
}

// withEnv returns environ with each "KEY=value" entry replacing any
// entry for the same key.
func withEnv(environ []string, entries ...string) []string {
	for _, entry := range entries {
		key, _, _ := strings.Cut(entry, "=")
		environ = append(dropKey(environ, key), entry)
	}
	return environ
}

// TestDoctorBoundaryRows_Baseline is the green baseline every
// perturbation below departs from by one value: doctorEnviron exits 0
// and every boundary row states its own status. Its serve port is one
// the test binary reserved and nothing listens on, so the probe's dial
// is refused and the row reads "no listener" without failing doctor.
func TestDoctorBoundaryRows_Baseline(t *testing.T) {
	rows, lines, err := doctorRows(t, doctorEnviron(t))
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil on the green baseline", err)
	}
	want := map[string]string{
		"client ip":      "direct: the peer must be the browser; a proxy in front must be configured as trusted",
		"web":            "ok",
		"bind addresses": "127.0.0.1 present",
		"boundary probe": "127.0.0.1 no listener",
	}
	for key, status := range want {
		if got := rows[key]; got != status {
			t.Errorf("row %q = %q, want %q", key, got, status)
		}
	}
	reg := productionRegistry()
	wantLines := len(reg.OwnedWall()) + len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) +
		len(reg.Thresholds()) + len(reg.IdentitySources()) + storeRowLines + membersRowLines + authStoreRowLines + backupDoctorRowsOff +
		unidecodeRowLines + probeRowLines + len(reg.Settings())
	if lines != wantLines {
		t.Errorf("doctor printed %d lines, want %d", lines, wantLines)
	}
}

// TestDoctorBoundaryRows_Perturbations runs doctor on doctorEnviron
// plus exactly one change per case and pins the row that change
// moves, so no exit 1 comes from an unrelated row.
func TestDoctorBoundaryRows_Perturbations(t *testing.T) {
	missingCSS := filepath.Join(t.TempDir(), "absent.css")

	tests := []struct {
		name       string
		env        []string // entries replacing the baseline's
		drop       string   // a baseline key removed
		wantFail   bool
		row        string
		wantStatus string // exact status, when non-empty
		wantHas    []string
	}{
		{
			name: "trusted proxy with two addresses",
			env: []string{
				"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
				"DUBPLATE_TRUSTED_PROXY_ADDRS=203.0.113.7, 2001:db8::1",
				"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
			},
			row:        "client ip",
			wantStatus: "trusted proxy: X-Forwarded-For from 2 addresses",
		},
		{
			name: "trusted proxy with one address, a duplicate collapsing",
			env: []string{
				"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
				"DUBPLATE_TRUSTED_PROXY_ADDRS=203.0.113.7,203.0.113.7",
				"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
			},
			row:        "client ip",
			wantStatus: "trusted proxy: X-Forwarded-For from 1 address",
		},
		{
			name: "trusted proxy naming a network",
			env: []string{
				"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
				"DUBPLATE_TRUSTED_PROXY_ADDRS=203.0.113.0/24",
				"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
			},
			wantFail: true,
			row:      "client ip",
			wantHas:  []string{"invalid: ", "an address, not a network"},
		},
		{
			name:     "unknown theme match",
			env:      []string{"DUBPLATE_THEME_MATCH=other"},
			wantFail: true,
			row:      "web",
			wantHas:  []string{"invalid: ", `must be exactly "dubplate"`},
		},
		{
			name:     "operator CSS file missing",
			env:      []string{"DUBPLATE_OPERATOR_CSS_FILE=" + missingCSS},
			wantFail: true,
			row:      "web",
			wantHas:  []string{"invalid: ", "DUBPLATE_OPERATOR_CSS_FILE"},
		},
		{
			name:       "loopback address other than 127.0.0.1",
			env:        []string{"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.2"},
			row:        "bind addresses",
			wantStatus: "127.0.0.2 present",
		},
		{
			name:       "address on no interface",
			env:        []string{"DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1"},
			wantFail:   true,
			row:        "bind addresses",
			wantStatus: "192.0.2.1 not on any interface",
		},
		{
			name:     "empty bind entry",
			env:      []string{"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,,::1"},
			wantFail: true,
			row:      "bind addresses",
			wantHas:  []string{"invalid: ", "empty entry"},
		},
		{
			name:       "an absent bind address is not probed",
			env:        []string{"DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1"},
			wantFail:   true,
			row:        "boundary probe",
			wantStatus: "192.0.2.1 not probed (not on any interface)",
		},
		{
			name: "every local address a trusted proxy",
			env: []string{
				"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
				"DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1",
				"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
				"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1",
			},
			row:        "boundary probe",
			wantStatus: "127.0.0.1 no listener (a trusted proxy, identity only); " + wantAllTrustedUnproven,
		},
		{
			name:     "an invalid serve prefix fails the web row",
			env:      []string{"DUBPLATE_SERVE_PREFIX=dubplate"},
			wantFail: true,
			row:      "web",
			wantHas:  []string{"invalid: ", "must begin with a separator"},
		},
		{
			name:       "an invalid serve prefix leaves the probe unprobed",
			env:        []string{"DUBPLATE_SERVE_PREFIX=dubplate"},
			wantFail:   true,
			row:        "boundary probe",
			wantStatus: "not probed",
		},
		{
			name:       "site name unset: client ip not checked",
			drop:       "DUBPLATE_SITE_NAME",
			wantFail:   true,
			row:        "client ip",
			wantStatus: "not checked",
		},
		{
			name:       "site name unset: web not checked",
			drop:       "DUBPLATE_SITE_NAME",
			wantFail:   true,
			row:        "web",
			wantStatus: "not checked",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			base := doctorEnviron(t)
			if tt.drop != "" {
				base = dropKey(base, tt.drop)
			}
			rows, _, err := doctorRows(t, withEnv(base, tt.env...))
			if (err != nil) != tt.wantFail {
				t.Fatalf("Execute() error = %v, wantFail %v", err, tt.wantFail)
			}
			got := rows[tt.row]
			if tt.wantStatus != "" && got != tt.wantStatus {
				t.Errorf("row %q = %q, want %q", tt.row, got, tt.wantStatus)
			}
			for _, sub := range tt.wantHas {
				if !strings.Contains(got, sub) {
					t.Errorf("row %q = %q, want it to contain %q", tt.row, got, sub)
				}
			}
		})
	}
}

// TestDoctorBoundaryRows_NeutralBrandingLogsNothingToStdout uses a
// branding neutral that is a color function, the one shape
// themeColors logs for, and proves two facts: newWebRoutes does write
// its record for this fixture, and doctor's output carries neither
// that record nor an extra line.
func TestDoctorBoundaryRows_NeutralBrandingLogsNothingToStdout(t *testing.T) {
	const branding = "DUBPLATE_BRANDING=--color-neutral: oklch(20% 0.02 250)"
	environ := withEnv(doctorEnviron(t), branding)

	var logs bytes.Buffer
	prefix, _, _, err := resolveServeConfig(environ)
	if err != nil {
		t.Fatalf("resolveServeConfig: %v", err)
	}
	if _, err := newWebRoutes(environ, prefix, fixtureSiteName, privateFiles{}, slog.New(slog.NewTextHandler(&logs, nil))); err != nil {
		t.Fatalf("newWebRoutes: %v", err)
	}
	if logs.Len() == 0 {
		t.Fatal("fixture state: newWebRoutes wrote no record for a color-function neutral, so this test proves nothing")
	}

	_, baselineLines, err := doctorRows(t, doctorEnviron(t))
	if err != nil {
		t.Fatalf("baseline Execute() error = %v", err)
	}
	rows, lines, err := doctorRows(t, environ)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil for a neutral branding value", err)
	}
	if lines != baselineLines {
		t.Errorf("doctor printed %d lines, want the baseline's %d", lines, baselineLines)
	}
	if got := rows["web"]; got != "ok" {
		t.Errorf("web row = %q, want ok", got)
	}
}

func TestBindAddrsRowStatus(t *testing.T) {
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
			status, failed := bindAddrsRowStatus(tt.environ, tt.addrs, tt.err)
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed != tt.wantFailed {
				t.Errorf("failed = %v, want %v", failed, tt.wantFailed)
			}
		})
	}
}

// freePort returns a loopback port this test listened on at
// 127.0.0.1:0 and released, for a server the test starts itself.
// reservedPort is the shared port nothing may ever listen on, so a test
// that needs a listener takes its own.
func freePort(t *testing.T) int {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve a loopback port: %v", err)
	}
	port := listenerPort(t, l)
	if err := l.Close(); err != nil {
		t.Fatalf("release the reserved port: %v", err)
	}
	return port
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

// TestDoctorBoundaryProbe_EndToEnd runs the serve verb itself, the
// shell newServeShell builds served by serve.Run on 127.0.0.1, and
// then doctor against the same port: the probe crosses a real socket
// into the running server's strip, and every count comes back zero.
func TestDoctorBoundaryProbe_EndToEnd(t *testing.T) {
	port := freePort(t)
	environ := withEnv(doctorEnviron(t), "DUBPLATE_SERVE_PORT="+strconv.Itoa(port))

	ctx, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() {
		cmd := newServeCmd(environ, productionRegistry(), io.Discard, io.Discard)
		served <- cmd.ExecuteContext(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		if err := <-served; err != nil {
			t.Errorf("serve returned %v, want nil after cancel", err)
		}
	})

	// Fixture state: the route answers a plain GET on the port before
	// doctor runs, so a "refused" row is the strip's answer and never
	// a server that had not started.
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/dubplate/" + serve.ProbeRouteName
	client := &http.Client{Transport: &http.Transport{}, Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	deadline := time.Now().Add(10 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("fixture state: plain GET answered %d, want 200", resp.StatusCode)
			}
			break
		}
		select {
		case err := <-served:
			t.Fatalf("serve returned before answering: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("fixture state: the route never answered: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	seedDoctorAdmin(t, environ)
	rows, _, err := doctorRows(t, environ)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil when the strip refuses every header", err)
	}
	want := "127.0.0.1 refused; " + probeTrailer
	if got := rows["boundary probe"]; got != want {
		t.Errorf("boundary probe row = %q, want %q", got, want)
	}
}

// TestDoctorBoundaryProbe_ListenerAnswers puts a server on the port
// doctor dials that answers the route's path with a canned reply, one
// per case: each count the strip should have zeroed, an answer that is
// not 200, and a body that is not the counts.
func TestDoctorBoundaryProbe_ListenerAnswers(t *testing.T) {
	const trailer = "; " + probeTrailer
	tests := []struct {
		name       string
		status     int
		body       string
		wantFail   bool
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
			wantFail:   true,
			wantStatus: "127.0.0.1 passed 2 identity header(s)" + trailer,
		},
		{
			name:       "forwarded-address headers passed",
			status:     http.StatusOK,
			body:       `{"identity":0,"forwarded":3,"sentinel":0}`,
			wantFail:   true,
			wantStatus: "127.0.0.1 passed 3 forwarded-address header(s)" + trailer,
		},
		{
			name:       "the sentinel passed",
			status:     http.StatusOK,
			body:       `{"identity":0,"forwarded":0,"sentinel":1}`,
			wantFail:   true,
			wantStatus: "127.0.0.1 passed the Connection-listed sentinel" + trailer,
		},
		{
			name:       "every group passed",
			status:     http.StatusOK,
			body:       `{"identity":1,"forwarded":1,"sentinel":1}`,
			wantFail:   true,
			wantStatus: "127.0.0.1 passed 1 identity header(s) and 1 forwarded-address header(s) and the Connection-listed sentinel" + trailer,
		},
		{
			name:       "a 404",
			status:     http.StatusNotFound,
			body:       "404 page not found\n",
			wantFail:   true,
			wantStatus: "127.0.0.1 unexpected answer: 404 Not Found",
		},
		{
			name:       "an unparseable body",
			status:     http.StatusOK,
			body:       "<html>",
			wantFail:   true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
		{
			name:       "an empty object",
			status:     http.StatusOK,
			body:       `{}`,
			wantFail:   true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
		{
			name:       "JSON null",
			status:     http.StatusOK,
			body:       `null`,
			wantFail:   true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
		{
			name:       "an object missing two counts",
			status:     http.StatusOK,
			body:       `{"identity":0}`,
			wantFail:   true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/dubplate/"+serve.ProbeRouteName {
					http.NotFound(w, r)
					return
				}
				w.WriteHeader(tt.status)
				_, _ = io.WriteString(w, tt.body)
			}))
			defer srv.Close()
			port := listenerPort(t, srv.Listener)

			rows, _, err := doctorRows(t, withEnv(doctorEnviron(t), "DUBPLATE_SERVE_PORT="+strconv.Itoa(port)))
			if (err != nil) != tt.wantFail {
				t.Fatalf("Execute() error = %v, wantFail %v", err, tt.wantFail)
			}
			if got := rows["boundary probe"]; got != tt.wantStatus {
				t.Errorf("boundary probe row = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}

// TestDoctorBoundaryProbe_WedgedListener puts a listener on the port
// that completes the handshake and never answers. The kernel accepts
// the connection into the backlog whether or not Accept is called, so
// the dial succeeds and the read is what times out: the row reads a
// timeout and fails doctor, and the run ends within the response bound
// rather than hanging.
func TestDoctorBoundaryProbe_WedgedListener(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	port := listenerPort(t, l)

	start := time.Now()
	rows, _, err := doctorRows(t, withEnv(doctorEnviron(t), "DUBPLATE_SERVE_PORT="+strconv.Itoa(port)))
	elapsed := time.Since(start)
	if err == nil {
		t.Fatal("Execute() error = nil, want a failure for a listener that never answers")
	}
	if got, want := rows["boundary probe"], "127.0.0.1 unexpected answer: timeout"; got != want {
		t.Errorf("boundary probe row = %q, want %q", got, want)
	}
	if limit := probeResponseBound + 2*time.Second; elapsed > limit {
		t.Errorf("doctor took %v, want under %v: the response bound did not end the wait", elapsed, limit)
	}
}

// TestBoundaryProbeRowStatus_UndeclaredPortNeverDials covers a
// registry that does not declare the serve port: the row reads "not
// probed" naming the key, adds no failure, and never dials. The
// environ names a listener's port so a dial, were one made, lands on a
// socket this test watches rather than the default 8080. Load refuses
// the port key as unknown on such a registry, so cfg comes from the
// same environ without it; the row reads the environ it is handed.
func TestBoundaryProbeRowStatus_UndeclaredPortNeverDials(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = l.Close() }()
	var accepted atomic.Int32
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			accepted.Add(1)
			_ = conn.Close()
		}
	}()

	reg := config.NewRegistry()
	base := replaceKey(baseEnviron(), "DUBPLATE_STATE_DIR", t.TempDir())
	cfg, err := config.Load(base, reg)
	if err != nil {
		t.Fatalf("fixture state: config.Load on a fresh registry: %v", err)
	}
	environ := withEnv(base, "DUBPLATE_SERVE_PORT="+strconv.Itoa(listenerPort(t, l)))

	status, failed := boundaryProbeRowStatus(t.Context(), environ, reg, cfg)
	if want := "not probed: DUBPLATE_SERVE_PORT is not registered"; status != want {
		t.Errorf("status = %q, want %q", status, want)
	}
	if failed {
		t.Error("failed = true, want no failure of the row's own")
	}
	if n := accepted.Load(); n != 0 {
		t.Errorf("the row dialed the listener %d time(s), want none", n)
	}
}

// TestBoundaryProbeRowStatus_NoFailureOfItsOwn pins the states in
// which the row reports why it did not probe and leaves the failure to
// the row that owns the cause.
func TestBoundaryProbeRowStatus_NoFailureOfItsOwn(t *testing.T) {
	reg := productionRegistry()
	baseline := doctorEnviron(t)
	cfg, err := config.Load(baseline, reg)
	if err != nil {
		t.Fatalf("fixture state: config.Load on the baseline: %v", err)
	}
	// Load accepts a network as the trusted proxy value, so cfg is
	// non-nil and the refusal comes from the row's own resolve.
	networkProxy := withEnv(baseline,
		"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
		"DUBPLATE_TRUSTED_PROXY_ADDRS=203.0.113.0/24",
		"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
	)
	networkProxyCfg, err := config.Load(networkProxy, reg)
	if err != nil {
		t.Fatalf("fixture state: config.Load with a network as the trusted proxy: %v", err)
	}

	tests := []struct {
		name       string
		environ    []string
		cfg        *config.Config
		wantStatus string
	}{
		{
			name:       "an absent bind address, which the bind row fails",
			environ:    withEnv(baseline, "DUBPLATE_SERVE_BIND_ADDRS=192.0.2.1"),
			cfg:        cfg,
			wantStatus: "192.0.2.1 not probed (not on any interface)",
		},
		{
			name:       "an invalid bind list, which the bind row fails",
			environ:    withEnv(baseline, "DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,,::1"),
			cfg:        cfg,
			wantStatus: "not probed",
		},
		{
			name:       "an invalid prefix, which the web row fails",
			environ:    withEnv(baseline, "DUBPLATE_SERVE_PREFIX=dubplate"),
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
			status, failed := boundaryProbeRowStatus(t.Context(), tt.environ, reg, tt.cfg)
			if status != tt.wantStatus {
				t.Errorf("status = %q, want %q", status, tt.wantStatus)
			}
			if failed {
				t.Error("failed = true, want the failure left to the row that owns the cause")
			}
		})
	}
}

// TestBoundaryProbeRowStatus_MixedTrustedBind binds two loopback
// addresses, one of them a trusted proxy: both are dialed on the
// reserved port, where nothing listens, the trusted one is named as
// probed for identity only, and the row joins both parts without
// failing. One address is not a trusted proxy, so the row carries no
// all-trusted clause.
func TestBoundaryProbeRowStatus_MixedTrustedBind(t *testing.T) {
	reg := productionRegistry()
	environ := withEnv(doctorEnviron(t),
		"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
		"DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1",
		"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
		"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,127.0.0.2",
	)
	cfg, err := config.Load(environ, reg)
	if err != nil {
		t.Fatalf("fixture state: config.Load: %v", err)
	}

	status, failed := boundaryProbeRowStatus(t.Context(), environ, reg, cfg)
	if want := "127.0.0.1 no listener (a trusted proxy, identity only), 127.0.0.2 no listener"; status != want {
		t.Errorf("status = %q, want %q", status, want)
	}
	if failed {
		t.Error("failed = true, want no failure")
	}
}

// wantAllTrustedUnproven is the clause the probe row adds when every bind
// address is a trusted proxy, written out here rather than read from
// the production constant so a changed clause fails these tests.
const wantAllTrustedUnproven = "forwarded and Connection-listed removal unproven: every local address is a trusted proxy"

// trustedLoopback is the same-host front's configuration: 127.0.0.1 is
// the one bind address and the one trusted proxy.
var trustedLoopback = []string{
	"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
	"DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1",
	"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
	"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1",
}

// TestDoctorBoundaryProbe_TrustedLoopbackEndToEnd runs the serve verb
// behind the same-host front's configuration and doctor against it.
// The route answers a trusted peer with the identity count alone, so
// the row proves group 1 at 127.0.0.1, says so, and reads the rest as
// unproven rather than passing it, and doctor exits 0.
func TestDoctorBoundaryProbe_TrustedLoopbackEndToEnd(t *testing.T) {
	port := freePort(t)
	environ := withEnv(doctorEnviron(t), append(slices.Clone(trustedLoopback), "DUBPLATE_SERVE_PORT="+strconv.Itoa(port))...)

	ctx, cancel := context.WithCancel(t.Context())
	served := make(chan error, 1)
	go func() {
		cmd := newServeCmd(environ, productionRegistry(), io.Discard, io.Discard)
		served <- cmd.ExecuteContext(ctx)
	}()
	t.Cleanup(func() {
		cancel()
		if err := <-served; err != nil {
			t.Errorf("serve returned %v, want nil after cancel", err)
		}
	})

	// Fixture state: the route answers before doctor runs, so the row
	// reads the strip's answer and never a server that had not started.
	url := "http://127.0.0.1:" + strconv.Itoa(port) + "/dubplate/" + serve.ProbeRouteName
	client := &http.Client{Transport: &http.Transport{}, Timeout: 2 * time.Second}
	defer client.CloseIdleConnections()
	deadline := time.Now().Add(10 * time.Second)
	for {
		resp, err := client.Get(url)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("fixture state: plain GET answered %d, want 200", resp.StatusCode)
			}
			break
		}
		select {
		case err := <-served:
			t.Fatalf("serve returned before answering: %v", err)
		default:
		}
		if time.Now().After(deadline) {
			t.Fatalf("fixture state: the route never answered: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}

	seedDoctorAdmin(t, environ)
	rows, _, err := doctorRows(t, environ)
	if err != nil {
		t.Errorf("Execute() error = %v, want nil when the strip refuses every identity header", err)
	}
	want := "127.0.0.1 refused (a trusted proxy, identity only); " + wantAllTrustedUnproven + "; " + probeTrailer
	if got := rows["boundary probe"]; got != want {
		t.Errorf("boundary probe row = %q, want %q", got, want)
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
		wantFail   bool
		wantStatus string
	}{
		{
			name:       "identity refused",
			body:       `{"identity":0}`,
			wantStatus: "127.0.0.1 refused" + tail + "; " + probeTrailer,
		},
		{
			name:       "identity headers passed",
			body:       `{"identity":2}`,
			wantFail:   true,
			wantStatus: "127.0.0.1 passed 2 identity header(s)" + tail + "; " + probeTrailer,
		},
		{
			name:       "no identity count",
			body:       `{"forwarded":0,"sentinel":0}`,
			wantFail:   true,
			wantStatus: "127.0.0.1 unexpected answer: 200 OK with an unparseable body" + tail,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/dubplate/"+serve.ProbeRouteName {
					http.NotFound(w, r)
					return
				}
				_, _ = io.WriteString(w, tt.body)
			}))
			defer srv.Close()
			port := listenerPort(t, srv.Listener)

			environ := withEnv(doctorEnviron(t), append(slices.Clone(trustedLoopback), "DUBPLATE_SERVE_PORT="+strconv.Itoa(port))...)
			rows, _, err := doctorRows(t, environ)
			if (err != nil) != tt.wantFail {
				t.Fatalf("Execute() error = %v, wantFail %v", err, tt.wantFail)
			}
			if got := rows["boundary probe"]; got != tt.wantStatus {
				t.Errorf("boundary probe row = %q, want %q", got, tt.wantStatus)
			}
		})
	}
}

// TestDoctorBoundaryProbe_SendsTheConfiguredHeader configures a
// client-address header outside the fixed family, from a proxy that is
// not a bind address, and records what the probe request carried: the
// configured header is the one [serve.ClientIP] reads, so a probe that
// never sent it could not prove the strip removes it.
func TestDoctorBoundaryProbe_SendsTheConfiguredHeader(t *testing.T) {
	var got atomic.Value
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.Store(r.Header.Get("X-House-Client-Address"))
		_, _ = io.WriteString(w, `{"identity":0,"forwarded":0,"sentinel":0}`)
	}))
	defer srv.Close()
	port := listenerPort(t, srv.Listener)

	environ := withEnv(doctorEnviron(t),
		"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
		"DUBPLATE_TRUSTED_PROXY_ADDRS=203.0.113.7",
		"DUBPLATE_TRUSTED_PROXY_HEADER=X-House-Client-Address",
		"DUBPLATE_SERVE_PORT="+strconv.Itoa(port),
	)
	if _, _, err := doctorRows(t, environ); err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}
	if v, _ := got.Load().(string); v != "probe" {
		t.Errorf("probe request X-House-Client-Address = %q, want %q", v, "probe")
	}
}
