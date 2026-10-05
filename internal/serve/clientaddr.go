// Package serve implements the dubplate HTTP shell: the trusted-proxy
// client-address derivation, the ingress strip, path normalization,
// and prefix routing that sit in front of the shell's handlers.
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

// TrustedProxyConfig carries the trusted-proxy configuration: the set
// of peer addresses permitted to report a client address, and the
// header they report it in. The zero value is mode off, under which
// [ClientIP] returns [PeerAddress] unconditionally, so RemoteAddr is
// the client address until an adopter configures otherwise.
type TrustedProxyConfig struct {
	proxies map[netip.Addr]struct{}
	header  string
}

// NewTrustedProxyConfig builds a [TrustedProxyConfig] that trusts
// header only when it arrives from one of proxies. An empty header
// leaves mode off regardless of proxies.
//
// Each of proxies must already be unmapped and carry no zone: a peer
// address and a header hop are both compared in that form, so an
// IPv4-mapped or zoned entry never matches anything. A duplicate entry
// collapses into the set.
func NewTrustedProxyConfig(proxies []netip.Addr, header string) TrustedProxyConfig {
	set := make(map[netip.Addr]struct{}, len(proxies))
	for _, p := range proxies {
		set[p] = struct{}{}
	}
	return TrustedProxyConfig{proxies: set, header: header}
}

// CheckProxyHeader reports why name cannot be a trusted-proxy client
// address header, or nil when it can. name must be an RFC 9110 token:
// [textproto.CanonicalMIMEHeaderKey] returns a name holding a space or
// another byte outside the token set unmodified, so [http.Header]
// would never match it. Past the grammar, a header is refused when the
// strip would remove it before [ClientIP] reads it, when adding it to
// the forwarded-address group would strip a header the shell depends
// on from every untrusted peer, or when it carries no address list a
// hop could parse. A header of the last kind would leave every client
// keyed on the proxy's own address, so one client's flood would spend
// the rate limits of the whole group. The comparison is case-insensitive, and a name's
// underscore spelling is refused with its dash spelling.
func CheckProxyHeader(name string) error {
	if name == "" {
		return errors.New("empty header name")
	}
	if strings.ContainsFunc(name, func(r rune) bool { return !isTokenChar(r) }) {
		return errors.New("not an HTTP header name: it holds a character outside RFC 9110's token set")
	}
	canon := textproto.CanonicalMIMEHeaderKey(strings.ReplaceAll(name, "_", "-"))
	if reason, ok := refusedProxyHeaders()[canon]; ok {
		return errors.New(reason)
	}
	return nil
}

// refusedProxyHeaders maps each canonical name [CheckProxyHeader]
// refuses to the reason it gives.
func refusedProxyHeaders() map[string]string {
	refused := map[string]string{
		"Forwarded":          "its for= parameters are not an address list, so every hop would be malformed",
		"X-Forwarded-Host":   "a forwarded header that carries the original host and no client address, so every hop would be malformed",
		"X-Forwarded-Proto":  "a forwarded header that carries the original scheme and no client address, so every hop would be malformed",
		"X-Forwarded-Prefix": "a forwarded header that carries the original path prefix and no client address, so every hop would be malformed",
		"Content-Type":       "the shell reads it to parse request bodies, and naming it would strip it from every untrusted peer",
		"Host":               "an end-to-end header that carries no client address",
		"Connection":         "a framing header that carries no client address",
		"Content-Length":     "a framing header that carries no client address",
		"Transfer-Encoding":  "a framing header that carries no client address",
		"Authorization":      "an end-to-end header that carries no client address",
	}
	for name := range group3KeepSet {
		refused[name] = "the shell reads it for sessions or the origin check, and naming it would strip it from every untrusted peer"
	}
	for _, name := range identityExtras {
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

// ProxyMode describes which client-IP source is live for a
// [TrustedProxyConfig], in the form the doctor row renders
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

// ClientIP reports the client address the shell holds accountable
// for r. This is a separate job from [PeerAddress]:
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
// supplies the answer, since there is nothing further to check.
//
// Each hop is an address, or an address and port whose address is
// kept; a hop with a zone, or one that parses as neither, is
// malformed. A malformed rightmost untrusted hop ends the scan, so no
// hop left of it is ever reached. The answer is the hop's unmapped
// address in [netip.Addr.String] form. A malformed, empty, or
// all-proxy header falls back to PeerAddress rather than an empty
// string, since an empty rate-limiter key is one bucket every attacker
// shares.
//
// Every line the header arrives on is read, not the first alone. RFC
// 9110 makes repeated field lines equivalent to one comma-joined
// value, and a proxy that appends its own line rather than extending
// the existing one is ordinary; reading the first line would take the
// leftmost, attacker-controlled end of the chain as though it were the
// whole of it.
func ClientIP(r *http.Request, cfg TrustedProxyConfig) string {
	peer := PeerAddress(r)
	if !cfg.trusts(peer) {
		return peer
	}
	if hop, ok := cfg.rightmostUntrusted(r.Header.Values(cfg.header)); ok {
		return hop.String()
	}
	return peer
}

// trusts reports whether mode is on and peer, a [PeerAddress] answer,
// is one of the configured proxies. The peer is parsed and unmapped
// first, since net/http's RemoteAddr carries an IPv4 client of a
// dual-stack listener in the mapped form ("::ffff:192.0.2.1") that a
// configured dotted address never equals. A peer that does not parse
// is never trusted.
func (c TrustedProxyConfig) trusts(peer string) bool {
	if c.header == "" {
		return false
	}
	addr, err := netip.ParseAddr(peer)
	if err != nil {
		return false
	}
	_, ok := c.proxies[addr.Unmap()]
	return ok
}

// rightmostUntrusted returns the rightmost hop across values that is
// not a configured proxy, and false when the chain is empty, entirely
// configured proxies, or its rightmost untrusted hop is malformed. The
// lines are one chain: the last line's rightmost hop is the nearest, so
// the scan runs backward through the lines and backward through each
// line's own hops.
//
// A malformed hop ends the scan rather than being skipped. Every hop
// left of the rightmost untrusted one was written by a party the chain
// cannot verify, so an attacker who could make the scan step over a
// planted malformed value would choose the answer outright.
func (c TrustedProxyConfig) rightmostUntrusted(values []string) (netip.Addr, bool) {
	for _, value := range slices.Backward(values) {
		for _, raw := range slices.Backward(splitHops(value)) {
			hop := trimHop(raw)
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

// parseHop parses one forwarding-header hop as an address, or as an
// address and port whose address is kept, since a front that appends
// the client's port would otherwise collapse every client onto one
// limiter key. A hop with a zone is malformed: a zone names an
// interface on the sender's host, which no client address carries. The
// address is returned unmapped, the form a configured proxy is stored
// in and a limiter keys on.
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
