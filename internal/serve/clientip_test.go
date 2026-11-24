package serve

import (
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
)

// addrs parses each of ss with [netip.MustParseAddr], the parsed form
// [NewTrustedProxyConfig] takes.
func addrs(ss ...string) []netip.Addr {
	out := make([]netip.Addr, 0, len(ss))
	for _, s := range ss {
		out = append(out, netip.MustParseAddr(s))
	}
	return out
}

// Documentation-reserved fixture addresses (RFC 5737, RFC 3849): the
// three IPv4 TEST-NET blocks and the IPv6 documentation prefix. Never
// an estate address.

func TestPeerAddress(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		header     string
		want       string
	}{
		{
			name:       "ipv4 peer with port",
			remoteAddr: "203.0.113.10:54321",
			want:       "203.0.113.10",
		},
		{
			name:       "ipv6 peer bracketed with port",
			remoteAddr: "[2001:db8::1]:54321",
			want:       "2001:db8::1",
		},
		{
			name:       "no port falls back to RemoteAddr verbatim",
			remoteAddr: "not-an-address",
			want:       "not-an-address",
		},
		{
			// Threat-model claim 3: a client-address header naming a
			// different address never changes PeerAddress's answer,
			// even though the peer here would be an allowlisted proxy
			// under a live TrustedProxyConfig. PeerAddress takes no
			// TrustedProxyConfig at all, so mode cannot reach it.
			name:       "claim 3: header naming a different address is ignored",
			remoteAddr: "203.0.113.10:54321",
			header:     "198.51.100.5",
			want:       "203.0.113.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.header != "" {
				r.Header.Set("X-Forwarded-For", tt.header)
			}

			if got := PeerAddress(r); got != tt.want {
				t.Errorf("PeerAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		proxies    []netip.Addr
		header     string // the configured header name; empty means mode off
		remoteAddr string
		headerVal  string
		want       string
	}{
		{
			name:       "mode off returns peer regardless of header",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  "198.51.100.5",
			want:       "203.0.113.10",
		},
		{
			// The peer here is NOT in the allowlist; the header names
			// an address that IS, alongside an unlisted leading hop.
			// A test that conflated peer and header trust, or that
			// skipped the peer check and read the header regardless,
			// would report "198.51.100.9" here instead of the peer;
			// this row cannot pass by conflating the two.
			name:       "non-allowlisted peer: header is discarded, never read into the answer",
			proxies:    addrs("198.51.100.5"),
			header:     "X-Forwarded-For",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  "198.51.100.9, 198.51.100.5",
			want:       "203.0.113.10",
		},
		{
			name:       "allowlisted peer: single-hop header is trusted",
			proxies:    addrs("203.0.113.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  "198.51.100.5",
			want:       "198.51.100.5",
		},
		{
			name:       "allowlisted peer: empty header falls back to peer, not empty string",
			proxies:    addrs("203.0.113.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  "",
			want:       "203.0.113.10",
		},
		{
			name:       "allowlisted peer: header of only commas and spaces falls back to peer",
			proxies:    addrs("203.0.113.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  " , , ",
			want:       "203.0.113.10",
		},
		{
			// Three hops, two of them configured proxies, the leftmost
			// attacker-supplied. Scanning right to left skips both
			// trailing proxy hops and lands on the attacker's own
			// value: the narrower half of the claim, written out
			// rather than computed from the input.
			name:       "three-hop: attacker is the nearest non-proxy hop, gets its own value",
			proxies:    addrs("192.0.2.1", "192.0.2.2"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.2:1111",
			headerVal:  "198.51.100.9, 192.0.2.1, 192.0.2.2",
			want:       "198.51.100.9",
		},
		{
			// Three hops: a forged leftmost entry the attacker set
			// itself, the real client address a trusted proxy
			// recorded in the middle, and the trusted proxy's own
			// address on the right. A left-to-right reader reports
			// the forged entry; right-to-left correctly skips the
			// proxy and stops at the real client.
			name:       "three-hop: right-to-left protects the real client from a forged leftmost entry",
			proxies:    addrs("192.0.2.9"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.9:1111",
			headerVal:  "198.51.100.1, 203.0.113.50, 192.0.2.9",
			want:       "203.0.113.50",
		},
		{
			name:       "ipv6 peer and header",
			proxies:    addrs("2001:db8::1"),
			header:     "X-Forwarded-For",
			remoteAddr: "[2001:db8::1]:5555",
			headerVal:  "2001:db8::2",
			want:       "2001:db8::2",
		},
		{
			name:       "trusted peer: the proxy's own rightmost hop is skipped",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "203.0.113.9, 192.0.2.10",
			want:       "203.0.113.9",
		},
		{
			// The rightmost untrusted hop does not parse, so the scan
			// ends there and answers the peer. Continuing leftward
			// would hand the answer to a value a client chose.
			name:       "trusted peer: a malformed rightmost hop ends the scan at the peer",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "203.0.113.9, garbage",
			want:       "192.0.2.10",
		},
		{
			name:       "trusted peer: a malformed hop left of the answer is never reached",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "garbage, 203.0.113.9",
			want:       "203.0.113.9",
		},
		{
			// netip.ParseAddrPort: "both the address and the port must
			// be numeric"; the address half is kept.
			name:       "trusted peer: an IPv4 hop with a port keeps the address",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "203.0.113.9:4711",
			want:       "203.0.113.9",
		},
		{
			name:       "trusted peer: a bracketed IPv6 hop with a port keeps the address",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "[2001:db8::1]:443",
			want:       "2001:db8::1",
		},
		{
			// Brackets with no port fail netip.ParseAddr and
			// netip.ParseAddrPort ("missing ]", measured), so the hop
			// is malformed.
			name:       "trusted peer: a bracketed IPv6 hop without a port is malformed",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "[2001:db8::1]",
			want:       "192.0.2.10",
		},
		{
			name:       "trusted peer: a zoned hop is malformed",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "fe80::1%eth0",
			want:       "192.0.2.10",
		},
		{
			// netip.ParseAddrPort accepts a zone inside the brackets
			// (measured: "[fe80::1%eth0]:443" parses with zone eth0),
			// so the zone refusal has to follow the second parse too.
			name:       "trusted peer: a zoned hop with a port is malformed",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "[fe80::1%eth0]:443",
			want:       "192.0.2.10",
		},
		{
			name:       "untrusted peer: a well-formed header is never read",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "198.51.100.20:1111",
			headerVal:  "203.0.113.9",
			want:       "198.51.100.20",
		},
		{
			// netip.Addr.String formats a mapped address with a
			// "::ffff:" prefix; the hop is unmapped first, so the
			// answer is the dotted form a limiter keys on.
			name:       "trusted peer: an IPv4-mapped hop answers in dotted form",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "::ffff:203.0.113.9",
			want:       "203.0.113.9",
		},
		{
			// A mapped hop naming a configured proxy is that proxy, so
			// it is skipped like its dotted spelling.
			name:       "trusted peer: an IPv4-mapped spelling of the proxy is skipped",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.10:1111",
			headerVal:  "203.0.113.9, ::ffff:192.0.2.10",
			want:       "203.0.113.9",
		},
		{
			// The peer arrives in mapped form on a dual-stack socket;
			// it is compared after the same parse and unmap, so the
			// configured dotted proxy still trusts it.
			name:       "mapped peer: compared unmapped against the configured proxy",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "[::ffff:192.0.2.10]:1111",
			headerVal:  "203.0.113.9",
			want:       "203.0.113.9",
		},
		{
			name:       "unparseable peer: never trusted",
			proxies:    addrs("192.0.2.10"),
			header:     "X-Forwarded-For",
			remoteAddr: "not-an-address",
			headerVal:  "203.0.113.9",
			want:       "not-an-address",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTrustedProxyConfig(tt.proxies, tt.header)

			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.headerVal != "" && tt.header != "" {
				r.Header.Set(tt.header, tt.headerVal)
			} else if tt.headerVal != "" {
				r.Header.Set("X-Forwarded-For", tt.headerVal)
			}

			if got := ClientIP(r, cfg); got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestProxyModeString(t *testing.T) {
	tests := []struct {
		name    string
		proxies []netip.Addr
		header  string
		want    ProxyMode
	}{
		{name: "zero value is direct", want: ModeDirect},
		{name: "configured header is trusted proxy", proxies: addrs("203.0.113.10"), header: "X-Forwarded-For", want: ModeTrustedProxy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTrustedProxyConfig(tt.proxies, tt.header)
			if got := cfg.Mode(); got != tt.want {
				t.Errorf("Mode() = %v, want %v", got, tt.want)
			}
		})
	}

	if got := ModeDirect.String(); got != "direct" {
		t.Errorf("ModeDirect.String() = %q, want %q", got, "direct")
	}
	if got := ModeTrustedProxy.String(); got != "trusted proxy" {
		t.Errorf("ModeTrustedProxy.String() = %q, want %q", got, "trusted proxy")
	}
}

// TestClientIPReadsEveryHeaderLineAsOneChain proves the forwarding
// header is read across every line it arrived on, right to left, and
// not from the first line alone. A proxy that adds its own field line
// rather than extending the existing one is ordinary, and RFC 9110
// makes the repeated lines equivalent to one comma-joined value; a
// reader that took only the first would take the leftmost,
// attacker-controlled end of the chain as the whole of it.
func TestClientIPReadsEveryHeaderLineAsOneChain(t *testing.T) {
	const peer = "192.0.2.9"

	tests := []struct {
		name    string
		proxies []netip.Addr
		lines   []string
		want    string
	}{
		{
			// The first line is the attacker's own forged entry; the
			// real chain arrived on the second. Reading line one alone
			// reports 198.51.100.1.
			name:    "the nearest hop is on the last line",
			proxies: addrs(peer),
			lines:   []string{"198.51.100.1", "203.0.113.50, " + peer},
			want:    "203.0.113.50",
		},
		{
			// Every hop on the last line is a configured proxy, so the
			// scan continues backward into the line before it.
			name:    "a fully trusted last line continues into the earlier one",
			proxies: addrs(peer),
			lines:   []string{"203.0.113.50", peer},
			want:    "203.0.113.50",
		},
		{
			// No line carries an untrusted hop, so the answer is the
			// peer rather than an empty rate-limiter key.
			name:    "every hop on every line trusted falls back to the peer",
			proxies: addrs(peer, "192.0.2.8"),
			lines:   []string{"192.0.2.8", peer},
			want:    peer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := NewTrustedProxyConfig(tt.proxies, "X-Forwarded-For")

			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = peer + ":1111"
			for _, line := range tt.lines {
				r.Header.Add("X-Forwarded-For", line)
			}

			if got := ClientIP(r, cfg); got != tt.want {
				t.Errorf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestCheckProxyHeader covers each refusal family by one or more
// members, and the accepted names an adopter's front actually sets.
func TestCheckProxyHeader(t *testing.T) {
	tests := []struct {
		name    string
		header  string
		wantErr string // a substring of the refusal; empty means accepted
	}{
		{"cloudflare's client header", "Cf-Connecting-Ip", ""},
		{"lowercase spelling of an accepted name", "x-forwarded-for", ""},
		{"an adopter's own name", "X-House-Client-Address", ""},
		{"empty", "", "empty header name"},
		{"a space", "X Bad", "token set"},
		{"a colon", "X-Bad:", "token set"},
		{"a non-ASCII letter", "X-Bäd", "token set"},
		{"cookie, group 3's keep set", "Cookie", "origin check"},
		{"origin, group 3's keep set", "origin", "origin check"},
		{"sec-fetch-site, group 3's keep set", "Sec-Fetch-Site", "origin check"},
		{"sec_fetch_site, the underscore spelling", "Sec_Fetch_Site", "origin check"},
		{"host", "Host", "end-to-end"},
		{"authorization", "authorization", "end-to-end"},
		{"connection", "Connection", "framing"},
		{"content-length", "Content-Length", "framing"},
		{"transfer-encoding", "Transfer-Encoding", "framing"},
		{"forwarded", "Forwarded", "for= parameters"},
		{"x-forwarded-host, an address-less forwarded name", "X-Forwarded-Host", "no client address"},
		{"x-forwarded-proto, an address-less forwarded name", "x-forwarded-proto", "no client address"},
		{"x-forwarded-prefix, an address-less forwarded name", "X-Forwarded-Prefix", "no client address"},
		{"x_forwarded_host, its underscore spelling", "X_Forwarded_Host", "no client address"},
		{"content-type, which the roster form needs", "Content-Type", "request bodies"},
		{"x-real-ip, an address-bearing forwarded name", "X-Real-Ip", ""},
		{"true-client-ip, an address-bearing forwarded name", "True-Client-Ip", ""},
		{"remote-user, an identity header", "Remote-User", "identity header"},
		{"remote_user, its underscore spelling", "Remote_User", "identity header"},
		{"a mixed-case identity header", "cf-access-jwt-assertion", "identity header"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckProxyHeader(tt.header)
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("CheckProxyHeader(%q) = %v, want nil", tt.header, err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Fatalf("CheckProxyHeader(%q) = %v, want an error containing %q", tt.header, err, tt.wantErr)
			}
		})
	}
}
