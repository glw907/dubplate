package main

import (
	"errors"
	"net/netip"
	"reflect"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// TestResolveTrustedProxy walks every state the trusted proxy class's
// three keys can meet, through config.Load and then the resolver, the
// order serve startup and doctor both run them in. A row names exactly
// one of three outcomes: a Load refusal, a resolver refusal, or the
// configuration the resolver returns.
//
// Each wanted configuration is built from the row's own literal
// addresses, already in the dotted, unmapped form the set compares in,
// so a resolver that skipped a trim or an unmap returns a set that
// differs from it.
func TestResolveTrustedProxy(t *testing.T) {
	const (
		flag   = "DUBPLATE_TRUSTED_PROXY_ENABLED"
		addrs  = "DUBPLATE_TRUSTED_PROXY_ADDRS"
		header = "DUBPLATE_TRUSTED_PROXY_HEADER"
	)

	tests := []struct {
		name       string
		set        map[string]string // keys layered over authEnviron; "" sets the key empty
		wantLoad   []string          // substrings of the Load refusal
		wantErr    []string          // substrings of the resolver refusal
		wantConfig serve.TrustedProxyConfig
	}{
		{
			name:       "flag absent is mode direct",
			wantConfig: serve.TrustedProxyConfig{},
		},
		{
			name:       "flag false is mode direct",
			set:        map[string]string{flag: "false"},
			wantConfig: serve.TrustedProxyConfig{},
		},
		{
			name:     "flag true, addresses absent",
			set:      map[string]string{flag: "true", header: "Cf-Connecting-Ip"},
			wantLoad: []string{addrs + ": required setting unset"},
		},
		{
			name:     "flag true, addresses empty",
			set:      map[string]string{flag: "true", addrs: "", header: "Cf-Connecting-Ip"},
			wantLoad: []string{addrs + ": required setting unset"},
		},
		{
			name:     "flag true, header absent",
			set:      map[string]string{flag: "true", addrs: "192.0.2.10"},
			wantLoad: []string{header + ": required setting unset"},
		},
		{
			name:     "flag true, header empty",
			set:      map[string]string{flag: "true", addrs: "192.0.2.10", header: ""},
			wantLoad: []string{header + ": required setting unset"},
		},
		{
			name:     "flag absent, addresses set",
			set:      map[string]string{addrs: "192.0.2.10"},
			wantLoad: []string{addrs + ": set without " + flag + ", the trusted proxy integration's enable flag"},
		},
		{
			name:     "flag not a boolean",
			set:      map[string]string{flag: "yes-please", addrs: "192.0.2.10", header: "Cf-Connecting-Ip"},
			wantLoad: []string{flag + `: invalid value "yes-please", want a boolean`},
		},
		{
			name:    "a hostname",
			set:     map[string]string{flag: "true", addrs: "proxy.example.test", header: "Cf-Connecting-Ip"},
			wantErr: []string{`"proxy.example.test": not an IP address`},
		},
		{
			name:    "a network in CIDR form",
			set:     map[string]string{flag: "true", addrs: "192.0.2.0/24", header: "Cf-Connecting-Ip"},
			wantErr: []string{`"192.0.2.0/24": an address, not a network`},
		},
		{
			name:    "an address with a zone",
			set:     map[string]string{flag: "true", addrs: "fe80::1%eth0", header: "Cf-Connecting-Ip"},
			wantErr: []string{`"fe80::1%eth0": carries a zone`},
		},
		{
			name:    "the IPv4 unspecified address",
			set:     map[string]string{flag: "true", addrs: "0.0.0.0", header: "Cf-Connecting-Ip"},
			wantErr: []string{`"0.0.0.0": the unspecified address`},
		},
		{
			name:    "the IPv6 unspecified address",
			set:     map[string]string{flag: "true", addrs: "::", header: "Cf-Connecting-Ip"},
			wantErr: []string{`"::": the unspecified address`},
		},
		{
			name:    "an empty entry between two addresses",
			set:     map[string]string{flag: "true", addrs: "192.0.2.10,,192.0.2.11", header: "Cf-Connecting-Ip"},
			wantErr: []string{addrs + ": entry 2 is empty"},
		},
		{
			name:    "the cookie header",
			set:     map[string]string{flag: "true", addrs: "192.0.2.10", header: "Cookie"},
			wantErr: []string{header + ` "Cookie":`, "origin check"},
		},
		{
			name:    "the forwarded header",
			set:     map[string]string{flag: "true", addrs: "192.0.2.10", header: "Forwarded"},
			wantErr: []string{header + ` "Forwarded":`, "for= parameters"},
		},
		{
			name:    "an identity header",
			set:     map[string]string{flag: "true", addrs: "192.0.2.10", header: "Remote-User"},
			wantErr: []string{header + ` "Remote-User":`, "identity header"},
		},
		{
			name:    "an identity header's underscore spelling",
			set:     map[string]string{flag: "true", addrs: "192.0.2.10", header: "Remote_User"},
			wantErr: []string{header + ` "Remote_User":`, "identity header"},
		},
		{
			name:    "a header name holding a space",
			set:     map[string]string{flag: "true", addrs: "192.0.2.10", header: "X Bad"},
			wantErr: []string{header + ` "X Bad":`, "token set"},
		},
		{
			name: "two bad entries are both named",
			set:  map[string]string{flag: "true", addrs: "proxy.example.test, 192.0.2.0/24", header: "Cf-Connecting-Ip"},
			wantErr: []string{
				`"proxy.example.test": not an IP address`,
				`"192.0.2.0/24": an address, not a network`,
			},
		},
		{
			name:    "a bad entry and a bad header are both named",
			set:     map[string]string{flag: "true", addrs: "0.0.0.0", header: "Cookie"},
			wantErr: []string{`"0.0.0.0": the unspecified address`, header + ` "Cookie":`},
		},
		{
			name: "a valid set, a mapped entry unmapped, a duplicate collapsed",
			set:  map[string]string{flag: "true", addrs: "192.0.2.10, ::ffff:198.51.100.7,192.0.2.10", header: "Cf-Connecting-Ip"},
			wantConfig: serve.NewTrustedProxyConfig(
				[]netip.Addr{netip.MustParseAddr("192.0.2.10"), netip.MustParseAddr("198.51.100.7")},
				"Cf-Connecting-Ip",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := authEnviron(t)
			for name, value := range tt.set {
				environ = replaceKey(environ, name, value)
			}

			cfg, err := config.Load(environ, productionRegistry())
			if len(tt.wantLoad) > 0 {
				if err == nil {
					t.Fatalf("config.Load() error = nil, want a refusal containing %q", tt.wantLoad)
				}
				for _, want := range tt.wantLoad {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("config.Load() error = %v, want it to contain %q", err, want)
					}
				}
				return
			}
			if err != nil {
				t.Fatalf("config.Load() error = %v, want nil: the row's fixture must pass the wall", err)
			}

			got, err := resolveTrustedProxy(cfg)
			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatalf("resolveTrustedProxy() error = nil, want a refusal containing %q", tt.wantErr)
				}
				for _, want := range tt.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("resolveTrustedProxy() error = %v, want it to contain %q", err, want)
					}
				}
				if !reflect.DeepEqual(got, serve.TrustedProxyConfig{}) {
					t.Errorf("resolveTrustedProxy() returned a configuration alongside its refusal, want the zero value")
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveTrustedProxy() error = %v, want nil", err)
			}
			if !reflect.DeepEqual(got, tt.wantConfig) {
				t.Errorf("resolveTrustedProxy() = %+v, want %+v", got, tt.wantConfig)
			}
		})
	}
}

// TestResolveTrustedProxy_RefusalsJoin proves the resolver joins its
// refusals with errors.Join rather than stopping at the first, so an
// operator reads every bad entry in one run.
func TestResolveTrustedProxy_RefusalsJoin(t *testing.T) {
	environ := authEnviron(t)
	environ = replaceKey(environ, "DUBPLATE_TRUSTED_PROXY_ENABLED", "true")
	environ = replaceKey(environ, "DUBPLATE_TRUSTED_PROXY_ADDRS", "0.0.0.0, proxy.example.test")
	environ = replaceKey(environ, "DUBPLATE_TRUSTED_PROXY_HEADER", "Cookie")
	cfg, err := config.Load(environ, productionRegistry())
	if err != nil {
		t.Fatalf("config.Load() error = %v, want nil", err)
	}

	_, err = resolveTrustedProxy(cfg)
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatalf("resolveTrustedProxy() error = %v (%T), want an errors.Join result", err, err)
	}
	if got := len(joined.Unwrap()); got != 3 {
		t.Errorf("joined refusals = %d, want 3 (two entries and the header): %v", got, err)
	}
}
