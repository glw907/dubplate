package main

import (
	"bytes"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

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
		len(reg.Thresholds()) + len(reg.IdentitySources()) + storeRowLines + membersRowLines + authDirRowLines + backupDoctorRowsOff +
		unidecodeRowLines + probeRowLines + musicVolumeRowLines + len(reg.Settings())
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
	listen, err := serve.ParseConfig(environ)
	if err != nil {
		t.Fatalf("serve.ParseConfig: %v", err)
	}
	if _, err := newWebRoutes(environ, listen.Prefix,fixtureSiteName, privateFiles{}, slog.New(slog.NewTextHandler(&logs, nil))); err != nil {
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

// listenerPort returns the port of a listener bound on 127.0.0.1.
func listenerPort(t *testing.T, l net.Listener) int {
	t.Helper()
	addr, ok := l.Addr().(*net.TCPAddr)
	if !ok || !addr.IP.Equal(net.IPv4(127, 0, 0, 1)) {
		t.Fatalf("fixture state: listener at %v, want 127.0.0.1", l.Addr())
	}
	return addr.Port
}

// placeholderWebRefusal opens the web row's status when the embedded
// build holds only the placeholder, the state of a tree where
// `npm run build` has not run.
const placeholderWebRefusal = "invalid: the embedded web build tree holds only the placeholder"

// startServeOnHeldPort serves the shell the serve verb builds, over a
// fixture web tree, on a loopback port this test holds from the moment
// it picks the number, and returns the environ doctor runs against:
// environ with extra applied and DUBPLATE_SERVE_PORT set to that port.
// Holding the listener leaves no gap in which another test's server
// could take the port and answer the route first. The fixture tree
// lets a checkout whose web/build holds only the placeholder, which
// serve itself refuses, still run the probe across a real socket.
// The plain GET it makes before returning proves the route answers, so
// a "refused" row is the strip's answer and never a server that had
// not started. Cleanup stops the server and closes the build.
func startServeOnHeldPort(t *testing.T, extra ...string) []string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("hold a loopback port: %v", err)
	}
	port := listenerPort(t, l)
	// serve refuses a staging directory or library root that does not
	// exist, so both are sibling directories here.
	environ := withMusicVolume(t, withEnv(doctorEnviron(t), append(slices.Clone(extra), "DUBPLATE_SERVE_PORT="+strconv.Itoa(port))...))

	cfg, err := loadConfig(environ, productionRegistry())
	if err != nil {
		_ = l.Close()
		t.Fatalf("loadConfig: %v", err)
	}
	listen, err := serve.ParseConfig(environ)
	if err != nil {
		_ = l.Close()
		t.Fatalf("serve.ParseConfig: %v", err)
	}
	srv, err := newServeShellVia(t.Context(), environ, productionRegistry(), listen.Prefix, listen.BindAddrs, cfg, discardTestLogger(),
		func(s *serve.Shell) registrar { return s }, newFixtureBuildFS(1, 1))
	if err != nil {
		_ = l.Close()
		t.Fatalf("newServeShellVia: %v", err)
	}
	t.Cleanup(func() {
		if err := srv.Close(); err != nil {
			t.Errorf("close the serve build: %v", err)
		}
	})
	httpSrv := &http.Server{Handler: srv.shell.Handler()}
	done := make(chan error, 1)
	go func() { done <- httpSrv.Serve(l) }()
	t.Cleanup(func() {
		if err := httpSrv.Close(); err != nil {
			t.Errorf("close the server: %v", err)
		}
		if err := <-done; !errors.Is(err, http.ErrServerClosed) {
			t.Errorf("Serve returned %v, want http.ErrServerClosed", err)
		}
	})

	client := &http.Client{Transport: &http.Transport{}, Timeout: 5 * time.Second}
	defer client.CloseIdleConnections()
	resp, err := client.Get("http://127.0.0.1:" + strconv.Itoa(port) + "/dubplate/" + serve.ProbeRouteName)
	if err != nil {
		t.Fatalf("fixture state: the route never answered: %v", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("fixture state: plain GET answered %d, want 200", resp.StatusCode)
	}
	return environ
}

// doctorErrBesidesWeb is err unless the web row alone explains it: in a
// tree whose build holds only the placeholder, doctor's web row fails
// and the command exits 1 whatever the row under test says. The row
// under test carries its own failure in its status, which each caller
// compares exactly.
func doctorErrBesidesWeb(rows map[string]string, err error) error {
	if err != nil && strings.HasPrefix(rows["web"], placeholderWebRefusal) {
		return nil
	}
	return err
}

// TestDoctorBoundaryProbe_EndToEnd serves the shell newServeShell
// builds on a held 127.0.0.1 port and runs doctor against that port:
// the probe crosses a real socket into the running server's strip, and
// every count comes back zero.
func TestDoctorBoundaryProbe_EndToEnd(t *testing.T) {
	environ := startServeOnHeldPort(t)
	seedDoctorAdmin(t, environ)
	rows, _, err := doctorRows(t, environ)
	if err := doctorErrBesidesWeb(rows, err); err != nil {
		t.Errorf("Execute() error = %v, want nil when the strip refuses every header", err)
	}
	want := "127.0.0.1 refused; " + wantProbeTrailer
	if got := rows["boundary probe"]; got != want {
		t.Errorf("boundary probe row = %q, want %q", got, want)
	}
}

// wantAllTrustedUnproven is the clause the probe row adds when every bind
// address is a trusted proxy, written out here rather than read from
// the production constant so a changed clause fails these tests.
const wantAllTrustedUnproven = "forwarded and Connection-listed removal unproven: every local address is a trusted proxy"

// wantProbeTrailer ends every probe row in which the route answered.
const wantProbeTrailer = "proves this server's strip, not a proxy's"

// trustedLoopback is the same-host front's configuration: 127.0.0.1 is
// the one bind address and the one trusted proxy.
var trustedLoopback = []string{
	"DUBPLATE_TRUSTED_PROXY_ENABLED=true",
	"DUBPLATE_TRUSTED_PROXY_ADDRS=127.0.0.1",
	"DUBPLATE_TRUSTED_PROXY_HEADER=X-Forwarded-For",
	"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1",
}

// TestDoctorBoundaryProbe_TrustedLoopbackEndToEnd serves the shell in
// the same-host front's configuration and runs doctor against it. The
// route answers a trusted peer with the identity count alone, so the
// row proves group 1 at 127.0.0.1, says so, and reads the rest as
// unproven rather than passing it, and doctor exits 0.
func TestDoctorBoundaryProbe_TrustedLoopbackEndToEnd(t *testing.T) {
	environ := startServeOnHeldPort(t, trustedLoopback...)
	seedDoctorAdmin(t, environ)
	rows, _, err := doctorRows(t, environ)
	if err := doctorErrBesidesWeb(rows, err); err != nil {
		t.Errorf("Execute() error = %v, want nil when the strip refuses every identity header", err)
	}
	want := "127.0.0.1 refused (a trusted proxy, identity only); " + wantAllTrustedUnproven + "; " + wantProbeTrailer
	if got := rows["boundary probe"]; got != want {
		t.Errorf("boundary probe row = %q, want %q", got, want)
	}
}
