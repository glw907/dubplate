package main

import (
	"bytes"
	"errors"
	"log/slog"
	"net"
	"path/filepath"
	"strings"
	"testing"
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
// and every boundary row states its own status.
func TestDoctorBoundaryRows_Baseline(t *testing.T) {
	rows, lines, err := doctorRows(t, doctorEnviron(t))
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil on the green baseline", err)
	}
	want := map[string]string{
		"client ip":      "direct",
		"web":            "ok",
		"bind addresses": "127.0.0.1 present",
		"boundary probe": "not probed",
	}
	for key, status := range want {
		if got := rows[key]; got != status {
			t.Errorf("row %q = %q, want %q", key, got, status)
		}
	}
	reg := productionRegistry()
	wantLines := len(reg.OwnedWall()) + len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) +
		len(reg.Thresholds()) + len(reg.IdentitySources()) + storeRowLines + backupDoctorRowsOff +
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
