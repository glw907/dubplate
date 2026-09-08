package serve

import (
	"net/http/httptest"
	"testing"
)

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
		proxies    []string
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
			proxies:    []string{"198.51.100.5"},
			header:     "X-Forwarded-For",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  "198.51.100.9, 198.51.100.5",
			want:       "203.0.113.10",
		},
		{
			name:       "allowlisted peer: single-hop header is trusted",
			proxies:    []string{"203.0.113.10"},
			header:     "X-Forwarded-For",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  "198.51.100.5",
			want:       "198.51.100.5",
		},
		{
			name:       "allowlisted peer: empty header falls back to peer, not empty string",
			proxies:    []string{"203.0.113.10"},
			header:     "X-Forwarded-For",
			remoteAddr: "203.0.113.10:1111",
			headerVal:  "",
			want:       "203.0.113.10",
		},
		{
			name:       "allowlisted peer: header of only commas and spaces falls back to peer",
			proxies:    []string{"203.0.113.10"},
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
			proxies:    []string{"192.0.2.1", "192.0.2.2"},
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
			proxies:    []string{"192.0.2.9"},
			header:     "X-Forwarded-For",
			remoteAddr: "192.0.2.9:1111",
			headerVal:  "198.51.100.1, 203.0.113.50, 192.0.2.9",
			want:       "203.0.113.50",
		},
		{
			name:       "ipv6 peer and header",
			proxies:    []string{"2001:db8::1"},
			header:     "X-Forwarded-For",
			remoteAddr: "[2001:db8::1]:5555",
			headerVal:  "2001:db8::2",
			want:       "2001:db8::2",
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
		proxies []string
		header  string
		want    ProxyMode
	}{
		{name: "zero value is direct", want: ModeDirect},
		{name: "configured header is trusted proxy", proxies: []string{"203.0.113.10"}, header: "X-Forwarded-For", want: ModeTrustedProxy},
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
