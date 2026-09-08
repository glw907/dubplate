// Package serve implements the dubplate HTTP shell: the trusted-proxy
// client-address derivation, the ingress strip, path normalization,
// and prefix routing that sit in front of the shell's handlers.
package serve

import (
	"net"
	"net/http"
	"slices"
)

// TrustedProxyConfig carries the trusted-proxy configuration: the set
// of peer addresses permitted to report a client address, and the
// header they report it in. The zero value is mode off, under which
// [ClientIP] returns [PeerAddress] unconditionally, so RemoteAddr is
// the client address until an adopter configures otherwise.
type TrustedProxyConfig struct {
	proxies map[string]struct{}
	header  string
}

// NewTrustedProxyConfig builds a [TrustedProxyConfig] that trusts
// header only when it arrives from one of proxies. An empty header
// leaves mode off regardless of proxies.
func NewTrustedProxyConfig(proxies []string, header string) TrustedProxyConfig {
	set := make(map[string]struct{}, len(proxies))
	for _, p := range proxies {
		set[p] = struct{}{}
	}
	return TrustedProxyConfig{proxies: set, header: header}
}

// ProxyMode describes which client-IP source is live for a
// [TrustedProxyConfig], in the form rung 5d's doctor row renders
// without re-deriving it from the configuration.
type ProxyMode int

const (
	// ModeDirect is the default: PeerAddress is the client address.
	ModeDirect ProxyMode = iota
	// ModeTrustedProxy trusts a configured header from a configured
	// set of proxy addresses.
	ModeTrustedProxy
)

// String renders m for a doctor row or a log line.
func (m ProxyMode) String() string {
	if m == ModeTrustedProxy {
		return "trusted proxy"
	}
	return "direct"
}

// Mode reports which client-IP source is live for cfg.
func (c TrustedProxyConfig) Mode() ProxyMode {
	if c.header == "" {
		return ModeDirect
	}
	return ModeTrustedProxy
}

// PeerAddress returns the TCP peer address of r: RemoteAddr with its
// port stripped. No header, argument, or [TrustedProxyConfig] changes
// what it returns, which is what makes it the sole input to every
// trust decision in this package. r.RemoteAddr always carries a port
// under net/http; a value that does not parse as host:port is
// returned verbatim.
func PeerAddress(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// ClientIP reports the client address rung 5's shell holds
// accountable for r. This is a separate job from [PeerAddress]:
// PeerAddress answers who dialed the socket and never changes; with
// cfg's mode off the two answers are the same address, and with mode
// on, ClientIP is the only one of the two that consults cfg or a
// header at all.
//
// The configured header is trusted only when PeerAddress(r) is one of
// cfg's configured proxies; a header arriving from any other peer is
// discarded unread into the return value. A multi-hop header value is
// parsed right to left, taking the first hop, scanning from the
// right, that is not itself a configured proxy: each hop a trusted
// proxy relays appends to the right, so an attacker cannot reach past
// the trusted chain by prepending fake entries on the left, though an
// attacker who is the nearest hop the chain cannot verify still
// supplies the answer, since there is nothing further to check. A
// malformed or empty header falls back to PeerAddress rather than an
// empty string, since an empty rate-limiter key is one bucket every
// attacker shares.
func ClientIP(r *http.Request, cfg TrustedProxyConfig) string {
	peer := PeerAddress(r)
	if cfg.header == "" {
		return peer
	}
	if _, trusted := cfg.proxies[peer]; !trusted {
		return peer
	}
	if hop := cfg.rightmostUntrusted(r.Header.Get(cfg.header)); hop != "" {
		return hop
	}
	return peer
}

// rightmostUntrusted returns the rightmost hop in header that is not
// a configured proxy, or "" when header is empty, malformed, or every
// hop is a configured proxy.
func (c TrustedProxyConfig) rightmostUntrusted(header string) string {
	hops := splitHops(header)
	for _, raw := range slices.Backward(hops) {
		hop := trimHop(raw)
		if hop == "" {
			continue
		}
		if _, isProxy := c.proxies[hop]; isProxy {
			continue
		}
		return hop
	}
	return ""
}

// splitHops splits a comma-separated forwarding header value into its
// hops, left to right, without trimming surrounding space.
func splitHops(header string) []string {
	if header == "" {
		return nil
	}
	var hops []string
	start := 0
	for i := 0; i < len(header); i++ {
		if header[i] == ',' {
			hops = append(hops, header[start:i])
			start = i + 1
		}
	}
	return append(hops, header[start:])
}

// trimHop trims the spaces and tabs a forwarding header commonly
// carries around each comma-separated hop.
func trimHop(hop string) string {
	start := 0
	for start < len(hop) && isHopSpace(hop[start]) {
		start++
	}
	end := len(hop)
	for end > start && isHopSpace(hop[end-1]) {
		end--
	}
	return hop[start:end]
}

func isHopSpace(b byte) bool {
	return b == ' ' || b == '\t'
}
