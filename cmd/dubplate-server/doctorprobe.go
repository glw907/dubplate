package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"syscall"
	"time"

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

// The boundary probe's bounds: the dial and the response each get
// their own, so a host that drops the handshake and a listener that
// accepts and never answers both end the row in bounded time.
const (
	probeDialBound     = 2 * time.Second
	probeResponseBound = 2 * time.Second
)

// probeAnswerCap bounds how much of the route's answer doctor reads.
// The counts object is a few dozen bytes; anything past this is not
// the route's answer.
const probeAnswerCap = 4 << 10

// probeTrailer ends every row in which the route answered. The probe
// dials the server directly, so it proves what this server strips and
// says nothing about a proxy in front of it.
const probeTrailer = "proves this server's strip, not a proxy's"

// trustedProbeNote follows each trusted address's part of the row: the
// route answers a trusted peer with the identity count alone.
const trustedProbeNote = " (a trusted proxy, identity only)"

// allTrustedUnproven is the row's clause when every bind address is a
// trusted proxy: no address could prove the forwarded-address or the
// Connection-listed removal, so the row says so instead of passing
// them.
const allTrustedUnproven = "forwarded and Connection-listed removal unproven: every local address is a trusted proxy"

// boundaryProbeRowStatus dials each configured bind address on the
// configured port, from that same address, sends the one probe request
// [serve.NewProbeRequest] builds, and reports what the running
// server's probe route counted. Per address: "refused" when every
// count is zero; "passed ..." naming each group that reached the
// route, which fails the row; "no listener" when the dial is refused
// or times out, which does not fail, since doctor runs before the
// server starts as well as beside it; "not probed (not on any
// interface)" when the host does not hold the address, which the bind
// addresses row already fails; and "unexpected answer: ..." for a
// status other than 200, a body that is not the counts, or a
// connection that accepts and never answers, which fails.
//
// An address in the trusted proxy set is probed for group 1 alone,
// since the strip keeps the forwarded-address headers for a trusted
// peer by design and the route answers it with the identity count
// only; its part carries a note saying so. When every address is one,
// the row adds a clause naming the forwarded-address and
// Connection-listed removal unproven, which is no failure: a same-host
// front configured as trusted is the recommended topology, and the
// identity count is all a probe from that address can prove. The probe
// request carries the trusted proxy configuration's own header beside
// the fixed forwarded-address family, so an untrusted address proves
// the strip removes the one header the client address is read from.
//
// The row probes nothing, and adds no failure of its own, when reg
// does not declare the serve port (a registry built for a test of
// another family), when cfg is nil because [config.Load] refused
// environ, or when the serve keys or the trusted proxy values do not
// resolve: each of those causes is reported, and failed, by the row
// that owns it.
func boundaryProbeRowStatus(ctx context.Context, environ []string, reg *config.Registry, cfg *config.Config) (status string, failed bool) {
	declared := slices.ContainsFunc(reg.Thresholds(), func(th config.Threshold) bool {
		return th.Key == servePortThreshold.Key
	})
	if !declared {
		return "not probed: " + servePortThreshold.Key + " is not registered", false
	}
	if cfg == nil {
		return "not probed", false
	}
	prefix, addrs, port, err := resolveServeConfig(environ)
	if err != nil {
		return "not probed", false
	}
	proxy, trusted, err := trustedProxy(cfg)
	if err != nil {
		return "not probed", false
	}

	parts := make([]string, 0, len(addrs))
	trustedCount, answered := 0, false
	for _, text := range addrs {
		ip, err := netip.ParseAddr(text)
		if err != nil {
			// The bind addresses row reports the same disagreement
			// between net.ParseIP and netip.ParseAddr, and fails.
			parts = append(parts, text+" not probed")
			continue
		}
		ip = ip.Unmap()
		_, isTrusted := trusted[ip]
		result := probeAddress(ctx, prefix, ip, port, proxy, isTrusted)
		part := text + " " + result.status
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

// trustedProxy returns the trusted proxy configuration cfg resolves to
// and its set of proxy addresses, empty when the class is off, or
// [resolveTrustedProxy]'s refusal when its values do not resolve.
func trustedProxy(cfg *config.Config) (serve.TrustedProxyConfig, map[netip.Addr]struct{}, error) {
	proxy, err := resolveTrustedProxy(cfg)
	if err != nil {
		return serve.TrustedProxyConfig{}, nil, err
	}
	set := make(map[netip.Addr]struct{})
	if proxy.Mode() == serve.ModeDirect {
		return proxy, set, nil
	}
	addrs, _ := parseProxyAddrs(cfg.IntegrationValue(trustedProxyIntegrationClass.Name, trustedProxyAddrsKey))
	for _, a := range addrs {
		set[a] = struct{}{}
	}
	return proxy, set, nil
}

// probeResult is one address's part of the probe row.
type probeResult struct {
	status string
	failed bool
	// answered is true when the route returned its counts, so the
	// row's trailer applies.
	answered bool
}

// probeAddress probes the server listening on ip and port with the
// probe request built for proxy. The dial binds ip as its local
// address, so the peer the server sees is the address it was
// configured to listen on, which the route permits. trusted reports
// whether ip is one of proxy's addresses, whose answer carries the
// identity count alone.
func probeAddress(ctx context.Context, prefix serve.Prefix, ip netip.Addr, port int, proxy serve.TrustedProxyConfig, trusted bool) probeResult {
	dialer := net.Dialer{
		Timeout:   probeDialBound,
		LocalAddr: &net.TCPAddr{IP: ip.AsSlice()},
	}
	host := net.JoinHostPort(ip.String(), strconv.Itoa(port))
	conn, err := dialer.DialContext(ctx, "tcp", host)
	switch {
	case err == nil:
	case errors.Is(err, syscall.ECONNREFUSED), isTimeout(err):
		return probeResult{status: "no listener"}
	case errors.Is(err, syscall.EADDRNOTAVAIL):
		return probeResult{status: "not probed (not on any interface)"}
	default:
		return probeResult{status: "unexpected dial error: " + err.Error(), failed: true}
	}
	defer func() { _ = conn.Close() }()

	counts, err := probeExchange(ctx, conn, prefix, host, proxy, trusted)
	if err != nil {
		return probeResult{status: "unexpected answer: " + err.Error(), failed: true}
	}
	return countsResult(counts)
}

// probeExchange writes the probe request built for proxy on conn and
// reads the route's counts, all within [probeResponseBound]. For a
// trusted peer only the identity count is required, and the other two
// read as zero whatever the body holds. The error it returns is the
// text the row prints after "unexpected answer: ".
func probeExchange(ctx context.Context, conn net.Conn, prefix serve.Prefix, host string, proxy serve.TrustedProxyConfig, trusted bool) (serve.ProbeCounts, error) {
	if err := conn.SetDeadline(time.Now().Add(probeResponseBound)); err != nil {
		return serve.ProbeCounts{}, err
	}
	req, err := serve.NewProbeRequest(ctx, prefix, host, proxy)
	if err != nil {
		return serve.ProbeCounts{}, err
	}
	if err := req.Write(conn); err != nil {
		return serve.ProbeCounts{}, answerError(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return serve.ProbeCounts{}, answerError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return serve.ProbeCounts{}, errors.New(resp.Status)
	}
	unparseable := fmt.Errorf("%s with an unparseable body", resp.Status)
	var answer probeAnswer
	if err := json.NewDecoder(io.LimitReader(resp.Body, probeAnswerCap)).Decode(&answer); err != nil {
		if isTimeout(err) {
			return serve.ProbeCounts{}, errProbeTimeout
		}
		return serve.ProbeCounts{}, unparseable
	}
	if answer.Identity == nil {
		return serve.ProbeCounts{}, unparseable
	}
	if trusted {
		return serve.ProbeCounts{Identity: *answer.Identity}, nil
	}
	if answer.Forwarded == nil || answer.Sentinel == nil {
		return serve.ProbeCounts{}, unparseable
	}
	return serve.ProbeCounts{
		Identity:  *answer.Identity,
		Forwarded: *answer.Forwarded,
		Sentinel:  *answer.Sentinel,
	}, nil
}

// probeAnswer is [serve.ProbeCounts] as the probe reads it, with each
// count a pointer so a key the body leaves out stays nil. A 200 whose
// body lacks any count, or is JSON null, did not come from the route,
// and reading it as all zeros would report a strip that never ran.
type probeAnswer struct {
	Identity  *int `json:"identity"`
	Forwarded *int `json:"forwarded"`
	Sentinel  *int `json:"sentinel"`
}

// errProbeTimeout is the answer error for a connection that accepted
// and did not answer within [probeResponseBound].
var errProbeTimeout = errors.New("timeout")

// answerError maps a deadline expiry to [errProbeTimeout] and passes
// any other error through.
func answerError(err error) error {
	if isTimeout(err) {
		return errProbeTimeout
	}
	return err
}

// isTimeout reports whether err is a [net.Error] that timed out.
func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// countsResult renders the route's counts: "refused" when every
// count is zero, and otherwise each group that reached the route,
// which fails the row.
func countsResult(c serve.ProbeCounts) probeResult {
	var passed []string
	if c.Identity > 0 {
		passed = append(passed, fmt.Sprintf("%d identity header(s)", c.Identity))
	}
	if c.Forwarded > 0 {
		passed = append(passed, fmt.Sprintf("%d forwarded-address header(s)", c.Forwarded))
	}
	if c.Sentinel > 0 {
		passed = append(passed, "the Connection-listed sentinel")
	}
	if len(passed) == 0 {
		return probeResult{status: "refused", answered: true}
	}
	return probeResult{status: "passed " + strings.Join(passed, " and "), failed: true, answered: true}
}
