package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// The four boundary rows doctor prints after the unidecode row and
// before the settings rows. They are fixed rows rather than registry
// declarations, so reg alone never accounts for them.
const (
	clientIPDoctorRow      = "client ip"
	webDoctorRow           = "web"
	bindAddrsDoctorRow     = "bind addresses"
	boundaryProbeDoctorRow = "boundary probe"
)

// foldRefusal renders err as one status line: its message with every
// newline an [errors.Join] carries folded to "; ".
func foldRefusal(err error) string {
	return strings.ReplaceAll(err.Error(), "\n", "; ")
}

// clientIPRowStatus reports how the server learns a client's address:
// "direct" when the trusted proxy class is off, "trusted proxy:
// <header> from <n> address(es)" when it is on and
// [resolveTrustedProxy] accepts it, and "invalid: " plus every
// refusal otherwise, which fails the row. cfg is nil when
// [config.Load] refused the environment, and the row then reads "not
// checked" and adds no failure of its own, since Load's own report
// already fails doctor. It names no proxy address: an address is
// deployment topology a terminal need not show, and the count is
// enough to see the mode.
func clientIPRowStatus(cfg *config.Config) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	proxy, err := resolveTrustedProxy(cfg)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	if proxy.Mode() == serve.ModeDirect {
		return proxy.Mode().String(), false
	}
	class := trustedProxyIntegrationClass.Name
	header := strings.TrimSpace(cfg.IntegrationValue(class, trustedProxyHeaderKey))
	addrs, _ := parseProxyAddrs(cfg.IntegrationValue(class, trustedProxyAddrsKey))
	distinct := make(map[netip.Addr]struct{}, len(addrs))
	for _, addr := range addrs {
		distinct[addr] = struct{}{}
	}
	noun := "addresses"
	if len(distinct) == 1 {
		noun = "address"
	}
	return fmt.Sprintf("%s: %s from %d %s", proxy.Mode(), header, len(distinct), noun), false
}

// webRowStatus runs the web resolution serve startup runs,
// [newWebRoutes] with the same inputs, so doctor and serve cannot
// disagree about the theme match, the palette, the branding value, or
// the operator CSS file. It reads the embedded bundle and the operator
// file and writes nothing. The logger discards: the one record
// newWebRoutes writes for a branding neutral that drops the
// theme-color meta is serve's to show, and doctor's stdout is the
// table alone. A refusal reads "invalid: " plus the error and fails
// the row; a nil cfg ([config.Load] refused the environment) reads
// "not checked" and adds no failure of its own.
func webRowStatus(environ []string, cfg *config.Config) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	prefixValue, err := config.ParseString(environ, servePrefixThreshold)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	prefix, err := serve.NewPrefix(prefixValue)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	private := privateFiles{
		roster:   cfg.IdentitySourceValue(rosterFileKey),
		stateDir: cfg.StateDir,
	}
	discard := slog.New(slog.NewTextHandler(io.Discard, nil))
	if _, err := newWebRoutes(environ, prefix, cfg.SiteName, private, discard); err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	return "ok", false
}

// bindAddrsRowStatus reports, for each address
// DUBPLATE_SERVE_BIND_ADDRS configures, whether the host holds it:
// "<address> present" when it is a loopback address or equals the IP
// of one of ifaceAddrs, and "<address> not on any interface"
// otherwise, which fails the row. The listener binds with IP_FREEBIND,
// so an address the host lacks raises no bind error and this row is
// the only place the absence shows. A loopback address counts present
// because Linux routes all of 127.0.0.0/8 to the loopback device while
// [net.InterfaceAddrs] lists only 127.0.0.1/8.
//
// ifaceAddrs and ifaceErr are [net.InterfaceAddrs]'s own results, so
// every state is testable without a seam. That call reads the address
// table over netlink, which a unit's RestrictAddressFamilies without
// AF_NETLINK denies, so an error reads "not checked: " plus the error
// and fails the row, naming that setting as the likely cause. An
// unset key is the registered loopback default; a value
// [serve.SplitBindAddrs] refuses reads "invalid: " plus its refusal.
func bindAddrsRowStatus(environ []string, ifaceAddrs []net.Addr, ifaceErr error) (status string, failed bool) {
	raw, err := config.ParseString(environ, serveBindAddrsThreshold)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	addrs, err := serve.SplitBindAddrs(raw)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	if ifaceErr != nil {
		return fmt.Sprintf("not checked: %v (the unit's RestrictAddressFamilies may deny the netlink read)", ifaceErr), true
	}

	held := make(map[netip.Addr]struct{}, len(ifaceAddrs))
	for _, a := range ifaceAddrs {
		if ip, ok := interfaceIP(a); ok {
			held[ip] = struct{}{}
		}
	}

	parts := make([]string, 0, len(addrs))
	for _, text := range addrs {
		ip, err := netip.ParseAddr(text)
		if err != nil {
			// SplitBindAddrs accepted the entry through net.ParseIP, so a
			// parse failure here is a disagreement between the two
			// parsers, reported rather than assumed away.
			return fmt.Sprintf("invalid: %q: %v", text, err), true
		}
		ip = ip.Unmap()
		if _, ok := held[ip]; ok || ip.IsLoopback() {
			parts = append(parts, text+" present")
			continue
		}
		parts = append(parts, text+" not on any interface")
		failed = true
	}
	return strings.Join(parts, ", "), failed
}

// interfaceIP returns the unmapped IP a [net.InterfaceAddrs] entry
// carries, and false for an address type that carries none.
func interfaceIP(a net.Addr) (netip.Addr, bool) {
	var ip net.IP
	switch v := a.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	default:
		return netip.Addr{}, false
	}
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

// boundaryProbeRowStatus is the boundary probe row's status.
func boundaryProbeRowStatus() (status string, failed bool) {
	return "not probed", false
}
