package serve

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/doctor"
)

// DoctorRows renders the network boundary's three rows: how the server
// learns a client's address, the bind addresses against this host's
// interfaces, and the boundary probe. cfg is nil when [config.Load]
// refused the environment.
func DoctorRows(ctx context.Context, environ []string, cfg *config.Config) []doctor.Row {
	clientIP := doctor.Row{Name: "client ip"}
	clientIP.Status, clientIP.Failed = clientIPStatus(cfg)
	addrs, addrsErr := net.InterfaceAddrs()
	bind := doctor.Row{Name: "bind addresses"}
	bind.Status, bind.Failed = bindAddrsStatus(environ, addrs, addrsErr)
	probe := doctor.Row{Name: "boundary probe"}
	probe.Status, probe.Failed = boundaryProbeStatus(ctx, environ, cfg)
	return []doctor.Row{clientIP, bind, probe}
}

// foldRefusal renders err as one status line: its message with every
// newline an [errors.Join] carries folded to "; ".
func foldRefusal(err error) string {
	return strings.ReplaceAll(err.Error(), "\n", "; ")
}

// directAssumption is what direct mode takes for granted: a proxy left
// untrusted in front collapses every client onto its address, so one
// client can exhaust the whole group's sign-in limits.
const directAssumption = "the peer must be the browser; a proxy in front must be configured as trusted"

// clientIPStatus reports how the server learns a client's address:
// "direct" and its assumption, "trusted proxy: <header> from <n>
// address(es)", or "invalid: " plus every refusal, which fails the row.
// A nil cfg reads "not checked" without failing, since Load's report
// already fails doctor. It counts the proxy addresses rather than
// naming them, which would print deployment topology.
func clientIPStatus(cfg *config.Config) (status string, failed bool) {
	if cfg == nil {
		return "not checked", false
	}
	proxy, err := ParseTrustedProxy(cfg)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	if proxy.mode() == proxyDirect {
		return proxy.mode().String() + ": " + directAssumption, false
	}
	noun := "addresses"
	if len(proxy.proxies) == 1 {
		noun = "address"
	}
	return fmt.Sprintf("%s: %s from %d %s", proxy.mode(), proxy.header, len(proxy.proxies), noun), false
}

// bindAddrsStatus reports, for each configured bind address, "<address>
// present" when it is loopback or the IP of one of ifaceAddrs, and
// "<address> not on any interface" otherwise, which fails the row. The
// listener binds with IP_FREEBIND, so this row is the only place a
// missing address shows. Loopback counts present because Linux routes
// all of 127.0.0.0/8 to the loopback device, while [net.InterfaceAddrs]
// lists only 127.0.0.1/8.
//
// ifaceAddrs and ifaceErr are [net.InterfaceAddrs]'s results. That call
// reads over netlink, which a unit's RestrictAddressFamilies can deny,
// so an error fails the row and names that setting.
func bindAddrsStatus(environ []string, ifaceAddrs []net.Addr, ifaceErr error) (status string, failed bool) {
	raw, err := config.ParseString(environ, bindAddrsThreshold)
	if err != nil {
		return "invalid: " + foldRefusal(err), true
	}
	addrs, err := parseBindAddrs(raw)
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
	for _, addr := range addrs {
		ip := addr.Unmap()
		if _, ok := held[ip]; ok || ip.IsLoopback() {
			parts = append(parts, addr.String()+" present")
			continue
		}
		parts = append(parts, addr.String()+" not on any interface")
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

// probeTrailer ends every row in which the route answered: the probe
// dials the server directly, so it says nothing about a proxy in front.
const probeTrailer = "proves this server's strip, not a proxy's"

const trustedProbeNote = " (a trusted proxy, identity only)"

const allTrustedUnproven = "forwarded and Connection-listed removal unproven: every local address is a trusted proxy"

// boundaryProbeStatus dials each configured bind address from that
// same address, sends the [newProbeRequest] probe, and reports what the
// running server's probe route counted. Per address: "refused" when
// every count is zero; "passed ..." naming each group that reached the
// route, which fails; "no listener" when the dial is refused or times
// out, which does not fail, since doctor also runs before the server
// starts; "not probed (not on any interface)", which the bind row
// already fails; and "unexpected answer: ..." for anything that is not
// the counts, which fails.
//
// A trusted proxy address is probed for group 1 alone, since the strip
// keeps group 2 for it by design. When every address is one, the row
// says the other removals are unproven without failing: a same-host
// trusted front is the recommended topology.
//
// The row probes nothing and adds no failure when cfg is nil or the
// serve or trusted proxy values do not resolve: the row that owns each
// cause reports and fails it.
func boundaryProbeStatus(ctx context.Context, environ []string, cfg *config.Config) (status string, failed bool) {
	if cfg == nil {
		return "not probed", false
	}
	serveCfg, err := ParseConfig(environ)
	if err != nil {
		return "not probed", false
	}
	proxy, err := ParseTrustedProxy(cfg)
	if err != nil {
		return "not probed", false
	}

	addrs := serveCfg.BindAddrs
	parts := make([]string, 0, len(addrs))
	trustedCount, answered := 0, false
	for _, addr := range addrs {
		ip := addr.Unmap()
		_, isTrusted := proxy.proxies[ip]
		result := probeAddress(ctx, serveCfg.Prefix, ip, serveCfg.Port, proxy, isTrusted)
		part := addr.String() + " " + result.status
		if isTrusted {
			trustedCount++
			part += trustedProbeNote
		}
		parts = append(parts, part)
		failed = failed || result.failed
		answered = answered || result.answered
	}
	status = strings.Join(parts, ", ")
	if len(addrs) > 0 && trustedCount == len(addrs) {
		status += "; " + allTrustedUnproven
	}
	if answered {
		status += "; " + probeTrailer
	}
	return status, failed
}
