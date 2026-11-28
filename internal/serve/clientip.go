package serve

import (
	"errors"
	"net"
	"net/http"
	"net/netip"
	"net/textproto"
	"slices"
	"strings"
)

// TrustedProxyConfig is the set of peer addresses permitted to report
// a client address, and the header they report it in. The zero value
// is direct mode, under which [ClientIP] is always the TCP peer.
type TrustedProxyConfig struct {
	proxies map[netip.Addr]struct{}
	header  string
}

// newTrustedProxyConfig builds a [TrustedProxyConfig] that trusts
// header only from one of proxies; an empty header leaves direct mode.
// Each of proxies must be unmapped and carry no zone, the form a peer
// and a hop are compared in, or it never matches.
func newTrustedProxyConfig(proxies []netip.Addr, header string) TrustedProxyConfig {
	set := make(map[netip.Addr]struct{}, len(proxies))
	for _, p := range proxies {
		set[p] = struct{}{}
	}
	return TrustedProxyConfig{proxies: set, header: header}
}

// validateProxyHeader reports why name cannot be the trusted-proxy
// client address header, or nil when it can. name must be an RFC 9110
// token, since [http.Header] never matches a name
// [textproto.CanonicalMIMEHeaderKey] leaves uncanonicalized. A header is
// also refused when the strip removes it before [ClientIP] reads it,
// when adding it to group 2 would strip a header the shell reads from
// every untrusted peer, or when it carries no address list: that last
// kind keys every client on the proxy's address, so one client's flood
// spends the whole group's rate limits. The match is by [foldedKey].
func validateProxyHeader(name string) error {
	if name == "" {
		return errors.New("empty header name")
	}
	if strings.ContainsFunc(name, func(r rune) bool { return !isTokenChar(r) }) {
		return errors.New("not an HTTP header name: it holds a character outside RFC 9110's token set")
	}
	if reason, ok := refusedProxyHeaders[foldedKey(name)]; ok {
		return errors.New(reason)
	}
	return nil
}

var refusedProxyHeaders = newRefusedProxyHeaders()

func newRefusedProxyHeaders() map[string]string {
	refused := map[string]string{
		"Content-Type":      "the shell reads it to parse request bodies, and naming it would strip it from every untrusted peer",
		"Host":              "an end-to-end header that carries no client address",
		"Connection":        "a framing header that carries no client address",
		"Content-Length":    "a framing header that carries no client address",
		"Transfer-Encoding": "a framing header that carries no client address",
		"Authorization":     "an end-to-end header that carries no client address",
	}
	for _, h := range forwardedAddressHeaders {
		if h.refusal != "" {
			refused[h.name] = h.refusal
		}
	}
	for name := range hopByHopExempt {
		refused[name] = "the shell reads it for sessions or the origin check, and naming it would strip it from every untrusted peer"
	}
	for _, name := range identityHeaders {
		refused[textproto.CanonicalMIMEHeaderKey(name)] = "an identity header the strip removes before the client address is read"
	}
	return refused
}

// isTokenChar reports whether r is an RFC 9110 tchar: a letter, a
// digit, or one of "!#$%&'*+-.^_`|~".
func isTokenChar(r rune) bool {
	switch {
	case 'a' <= r && r <= 'z', 'A' <= r && r <= 'Z', '0' <= r && r <= '9':
		return true
	}
	return strings.ContainsRune("!#$%&'*+-.^_`|~", r)
}

// proxyMode is which client-IP source is live for a
// [TrustedProxyConfig].
type proxyMode int

const (
	proxyDirect proxyMode = iota
	proxyTrusted
)

func (m proxyMode) String() string {
	if m == proxyTrusted {
		return "trusted proxy"
	}
	return "direct"
}

func (c TrustedProxyConfig) mode() proxyMode {
	if c.header == "" {
		return proxyDirect
	}
	return proxyTrusted
}

// peerHost returns r.RemoteAddr with its port stripped, or verbatim
// when it does not parse as host:port.
func peerHost(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// peerAddr returns the TCP peer address of r, unmapped to match the
// configured proxies and bind addresses, and false when [peerHost] is
// not an IP address. No header or configuration changes its answer,
// which is what makes it the sole input to every trust decision in this
// package.
func peerAddr(r *http.Request) (netip.Addr, bool) {
	addr, err := netip.ParseAddr(peerHost(r))
	if err != nil {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}

// ClientIP reports the client address the shell holds accountable for
// r. In direct mode it is the TCP peer as [peerHost] prints it.
//
// The configured header is read only when the peer is one of cfg's
// proxies. Its hops are scanned right to left, across every line it
// arrives on, since RFC 9110 makes repeated lines one comma-joined
// value; the answer is the first hop that is not itself a configured
// proxy, so entries prepended on the left never reach past the trusted
// chain. Each hop is an address, or an address and port; one with a
// zone, or that parses as neither, is malformed and ends the scan. The
// answer is the hop's unmapped address in [netip.Addr.String] form. A
// malformed, empty, or all-proxy header falls back to peerHost, never
// an empty string, which would be one rate-limiter bucket every
// attacker shares.
func ClientIP(r *http.Request, cfg TrustedProxyConfig) string {
	if peer, ok := peerAddr(r); ok && cfg.trusts(peer) {
		if hop, ok := cfg.rightmostUntrusted(r.Header.Values(cfg.header)); ok {
			return hop.String()
		}
	}
	return peerHost(r)
}

// trusts reports whether peer, a [peerAddr] answer, is a configured
// proxy in trusted mode.
func (c TrustedProxyConfig) trusts(peer netip.Addr) bool {
	if c.header == "" {
		return false
	}
	_, ok := c.proxies[peer]
	return ok
}

// rightmostUntrusted returns the rightmost hop across values that is
// not a configured proxy, and false when the chain is empty, all
// proxies, or that hop is malformed. A malformed hop ends the scan
// rather than being skipped: an attacker who could make the scan step
// over a planted malformed value would choose the answer outright.
func (c TrustedProxyConfig) rightmostUntrusted(values []string) (netip.Addr, bool) {
	for _, value := range slices.Backward(values) {
		for _, raw := range slices.Backward(strings.Split(value, ",")) {
			hop := strings.Trim(raw, " \t")
			if hop == "" {
				continue
			}
			addr, ok := parseHop(hop)
			if !ok {
				return netip.Addr{}, false
			}
			if _, isProxy := c.proxies[addr]; isProxy {
				continue
			}
			return addr, true
		}
	}
	return netip.Addr{}, false
}

// parseHop parses one hop as an address, or an address and port whose
// address is kept, so a front that appends the client's port does not
// collapse every client onto one limiter key. A zone, which no client
// address carries, makes the hop malformed. The address is returned
// unmapped.
func parseHop(hop string) (netip.Addr, bool) {
	addr, err := netip.ParseAddr(hop)
	if err != nil {
		addrPort, err := netip.ParseAddrPort(hop)
		if err != nil {
			return netip.Addr{}, false
		}
		addr = addrPort.Addr()
	}
	if addr.Zone() != "" {
		return netip.Addr{}, false
	}
	return addr.Unmap(), true
}
