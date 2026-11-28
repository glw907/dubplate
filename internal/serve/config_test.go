package serve

import (
	"errors"
	"net/netip"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/glw907/dubplate/internal/config"
)

// wallEnviron returns the owned wall settings this registry declares, the environment
// every [config.Load] in this file starts from.
func wallEnviron() []string {
	return []string{
		"DUBPLATE_LIBRARY_ROOT=/srv/music/library",
		"DUBPLATE_STATE_DIR=/var/lib/dubplate",
		"DUBPLATE_SITE_NAME=Test Family",
		"DUBPLATE_ADMIN_EMAIL=admin@example.org",
	}
}

// configRegistry builds the registry this package's two Register
// functions fill, over the config package's own seed.
func configRegistry() *config.Registry {
	reg := config.NewRegistry()
	Register(reg)
	RegisterTrustedProxy(reg)
	return reg
}

// withKey returns a copy of environ with name set to value, replacing
// any earlier entry.
func withKey(environ []string, name, value string) []string {
	out := slices.DeleteFunc(slices.Clone(environ), func(kv string) bool {
		n, _, _ := strings.Cut(kv, "=")
		return n == name
	})
	return append(out, name+"="+value)
}

// TestParseTrustedProxy walks every state the trusted proxy class's
// three keys can meet, through config.Load and then the resolver, the
// order serve startup and doctor both run them in. A row names exactly
// one of three outcomes: a Load refusal, a resolver refusal, or the
// configuration the resolver returns.
//
// Each wanted configuration is built from the row's own literal
// addresses, already in the dotted, unmapped form the set compares in,
// so a resolver that skipped a trim or an unmap returns a set that
// differs from it.
func TestParseTrustedProxy(t *testing.T) {
	const (
		flag   = "DUBPLATE_TRUSTED_PROXY_ENABLED"
		addrs  = "DUBPLATE_TRUSTED_PROXY_ADDRS"
		header = "DUBPLATE_TRUSTED_PROXY_HEADER"
	)

	tests := []struct {
		name       string
		set        map[string]string // keys layered over wallEnviron; "" sets the key empty
		wantLoad   []string          // substrings of the Load refusal
		wantErr    []string          // substrings of the resolver refusal
		wantConfig TrustedProxyConfig
	}{
		{
			name:       "flag absent is mode direct",
			wantConfig: TrustedProxyConfig{},
		},
		{
			name:       "flag false is mode direct",
			set:        map[string]string{flag: "false"},
			wantConfig: TrustedProxyConfig{},
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
			wantConfig: newTrustedProxyConfig(
				[]netip.Addr{netip.MustParseAddr("192.0.2.10"), netip.MustParseAddr("198.51.100.7")},
				"Cf-Connecting-Ip",
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			environ := wallEnviron()
			for name, value := range tt.set {
				environ = withKey(environ, name, value)
			}

			cfg, err := config.Load(environ, configRegistry())
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

			got, err := ParseTrustedProxy(cfg)
			if len(tt.wantErr) > 0 {
				if err == nil {
					t.Fatalf("ParseTrustedProxy() error = nil, want a refusal containing %q", tt.wantErr)
				}
				for _, want := range tt.wantErr {
					if !strings.Contains(err.Error(), want) {
						t.Errorf("ParseTrustedProxy() error = %v, want it to contain %q", err, want)
					}
				}
				if !reflect.DeepEqual(got, TrustedProxyConfig{}) {
					t.Errorf("ParseTrustedProxy() returned a configuration alongside its refusal, want the zero value")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseTrustedProxy() error = %v, want nil", err)
			}
			if !reflect.DeepEqual(got, tt.wantConfig) {
				t.Errorf("ParseTrustedProxy() = %+v, want %+v", got, tt.wantConfig)
			}
		})
	}
}

// TestParseTrustedProxy_RefusalsJoin proves the resolver joins its
// refusals with errors.Join rather than stopping at the first, so an
// operator reads every bad entry in one run.
func TestParseTrustedProxy_RefusalsJoin(t *testing.T) {
	environ := wallEnviron()
	environ = withKey(environ, "DUBPLATE_TRUSTED_PROXY_ENABLED", "true")
	environ = withKey(environ, "DUBPLATE_TRUSTED_PROXY_ADDRS", "0.0.0.0, proxy.example.test")
	environ = withKey(environ, "DUBPLATE_TRUSTED_PROXY_HEADER", "Cookie")
	cfg, err := config.Load(environ, configRegistry())
	if err != nil {
		t.Fatalf("config.Load() error = %v, want nil", err)
	}

	_, err = ParseTrustedProxy(cfg)
	var joined interface{ Unwrap() []error }
	if !errors.As(err, &joined) {
		t.Fatalf("ParseTrustedProxy() error = %v (%T), want an errors.Join result", err, err)
	}
	if got := len(joined.Unwrap()); got != 3 {
		t.Errorf("joined refusals = %d, want 3 (two entries and the header): %v", got, err)
	}
}

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name       string
		environ    []string
		wantErr    error
		wantAddrs  []string
		wantPort   int
		wantPrefix string
	}{
		{
			name:       "defaults",
			environ:    nil,
			wantAddrs:  []string{"127.0.0.1"},
			wantPort:   8080,
			wantPrefix: "/dubplate",
		},
		{
			name: "configured",
			environ: []string{
				"DUBPLATE_SERVE_BIND_ADDRS=127.0.0.1,::1",
				"DUBPLATE_SERVE_PORT=9090",
				"DUBPLATE_SERVE_PREFIX=/library",
			},
			wantAddrs:  []string{"127.0.0.1", "::1"},
			wantPort:   9090,
			wantPrefix: "/library",
		},
		{
			name:    "invalid prefix",
			environ: []string{"DUBPLATE_SERVE_PREFIX=no-leading-separator"},
			wantErr: errInvalidPrefix,
		},
		{
			name:    "unparseable bind list",
			environ: []string{"DUBPLATE_SERVE_BIND_ADDRS=not-an-address"},
			wantErr: errInvalidBindAddr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfig(tt.environ)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParseConfig() error = %v, want it to wrap %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseConfig() error = %v, want nil", err)
			}
			if addrs := addrTexts(got.BindAddrs); !slices.Equal(addrs, tt.wantAddrs) {
				t.Errorf("addrs = %v, want %v", addrs, tt.wantAddrs)
			}
			if got.Port != tt.wantPort {
				t.Errorf("port = %d, want %d", got.Port, tt.wantPort)
			}
			if got.Prefix.String() != tt.wantPrefix {
				t.Errorf("prefix = %q, want %q", got.Prefix.String(), tt.wantPrefix)
			}
		})
	}
}

// TestParsePrefix proves the prefix reader ignores the other two serve
// keys: a bind list ParseConfig refuses leaves the prefix readable.
func TestParsePrefix(t *testing.T) {
	tests := []struct {
		name    string
		environ []string
		want    string
		wantErr error
	}{
		{name: "the default", want: "/dubplate"},
		{name: "configured", environ: []string{"DUBPLATE_SERVE_PREFIX=/library"}, want: "/library"},
		{
			name:    "a refused bind list does not reach it",
			environ: []string{"DUBPLATE_SERVE_PREFIX=/library", "DUBPLATE_SERVE_BIND_ADDRS=not-an-address", "DUBPLATE_SERVE_PORT=nope"},
			want:    "/library",
		},
		{name: "invalid", environ: []string{"DUBPLATE_SERVE_PREFIX=library"}, wantErr: errInvalidPrefix},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePrefix(tt.environ)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParsePrefix() error = %v, want it to wrap %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePrefix() error = %v, want nil", err)
			}
			if got.String() != tt.want {
				t.Errorf("ParsePrefix() = %q, want %q", got.String(), tt.want)
			}
		})
	}
}

// addrTexts returns each of addrs in its printed form, so a test
// compares parsed addresses against literals.
func addrTexts(addrs []netip.Addr) []string {
	texts := make([]string, 0, len(addrs))
	for _, a := range addrs {
		texts = append(texts, a.String())
	}
	return texts
}

func TestParseBindAddrs(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr bool
	}{
		{
			name: "single address",
			raw:  "127.0.0.1",
			want: []string{"127.0.0.1"},
		},
		{
			name: "several addresses",
			raw:  "127.0.0.1,::1",
			want: []string{"127.0.0.1", "::1"},
		},
		{
			name: "whitespace around a separator",
			raw:  "127.0.0.1 , ::1",
			want: []string{"127.0.0.1", "::1"},
		},
		{
			name: "an IPv6 address in its long form",
			raw:  "0:0:0:0:0:0:0:1",
			want: []string{"::1"},
		},
		{
			name:    "empty entry",
			raw:     "127.0.0.1,,::1",
			wantErr: true,
		},
		{
			name:    "unparseable entry",
			raw:     "127.0.0.1,not-an-address",
			wantErr: true,
		},
		{
			name:    "an address carrying a zone",
			raw:     "fe80::1%eth0",
			wantErr: true,
		},
		// The wildcard rows: each parses as an IP address, so nothing
		// but the unspecified check refuses it.
		{
			name:    "the IPv4 wildcard alone",
			raw:     "0.0.0.0",
			wantErr: true,
		},
		{
			name:    "the IPv6 wildcard alone",
			raw:     "::",
			wantErr: true,
		},
		{
			name:    "the IPv4 wildcard beside a real address",
			raw:     "127.0.0.1,0.0.0.0",
			wantErr: true,
		},
		{
			name:    "the IPv4-mapped IPv6 wildcard",
			raw:     "::ffff:0.0.0.0",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBindAddrs(tt.raw)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseBindAddrs(%q) error = nil, want error", tt.raw)
				}
				if !errors.Is(err, errInvalidBindAddr) {
					t.Errorf("parseBindAddrs(%q) error = %v, want it to wrap errInvalidBindAddr", tt.raw, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseBindAddrs(%q) error = %v, want nil", tt.raw, err)
			}
			if texts := addrTexts(got); !slices.Equal(texts, tt.want) {
				t.Errorf("parseBindAddrs(%q) = %v, want %v", tt.raw, texts, tt.want)
			}
		})
	}
}

// TestValidatePort proves the port's bound at the function the binary
// runs beside config.Load: 0 would listen on an ephemeral port, and a
// port below 1024 cannot bind under the serve unit's empty capability
// set, so both refuse naming the capability, while 1024 and 65535, the
// bound's own ends, pass.
func TestValidatePort(t *testing.T) {
	tests := []struct {
		port    string
		wantErr string // the whole refusal; empty means the port passes
	}{
		{port: "0", wantErr: "DUBPLATE_SERVE_PORT: 0 is below 1024: the serve unit withholds CAP_NET_BIND_SERVICE, so it cannot bind a privileged port"},
		{port: "80", wantErr: "DUBPLATE_SERVE_PORT: 80 is below 1024: the serve unit withholds CAP_NET_BIND_SERVICE, so it cannot bind a privileged port"},
		{port: "1023", wantErr: "DUBPLATE_SERVE_PORT: 1023 is below 1024: the serve unit withholds CAP_NET_BIND_SERVICE, so it cannot bind a privileged port"},
		{port: "1024"},
		{port: "65535"},
		{port: "65536", wantErr: "DUBPLATE_SERVE_PORT: 65536 is above 65535"},
	}
	for _, tt := range tests {
		t.Run(tt.port, func(t *testing.T) {
			err := ValidatePort([]string{"DUBPLATE_SERVE_PORT=" + tt.port})
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidatePort() error = %v, want nil", err)
				}
				return
			}
			if err == nil || err.Error() != tt.wantErr {
				t.Errorf("ValidatePort() error = %v, want %q", err, tt.wantErr)
			}
		})
	}
}
