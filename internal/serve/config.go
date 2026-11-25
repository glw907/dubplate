package serve

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/glw907/dubplate/internal/config"
)

// BindAddrsThreshold is the comma-separated list of addresses the
// serve verb binds, loopback only by default. The registry has no list
// kind, so it registers as a string and [ParseConfig] parses the list.
var BindAddrsThreshold = config.Threshold{
	Key:         "DUBPLATE_SERVE_BIND_ADDRS",
	Kind:        config.KindString,
	Default:     "127.0.0.1",
	Doc:         "a comma-separated list of bind addresses, for example 127.0.0.1,::1",
	Description: "the addresses the server listens on, comma-separated",
	Sensitivity: config.Sensitive,
}

// PortThreshold is the TCP port the serve verb binds on every
// configured address.
var PortThreshold = config.Threshold{
	Key:         "DUBPLATE_SERVE_PORT",
	Kind:        config.KindInt,
	Default:     "8080",
	Doc:         "an integer TCP port from 1024 through 65535",
	Description: "the TCP port the server listens on",
	Sensitivity: config.Sensitive,
}

// portFloor and portCeiling bound DUBPLATE_SERVE_PORT. The
// floor is Linux's default net.ipv4.ip_unprivileged_port_start: the
// serve unit's empty CapabilityBoundingSet= withholds
// CAP_NET_BIND_SERVICE, so a lower port fails at bind, and port 0 would
// listen on whatever ephemeral port the kernel picks.
const (
	portFloor   = 1024
	portCeiling = 65535
)

// ValidatePort checks DUBPLATE_SERVE_PORT against
// [portFloor] and [portCeiling], naming the capability a
// privileged port would need. The binary runs it beside [config.Load],
// so doctor and every verb refuse an out-of-range port before serve
// ever binds. A value that does not parse returns [config.ParseInt]'s
// [config.KeyError].
func ValidatePort(environ []string) error {
	port, err := config.ParseInt(environ, PortThreshold)
	if err != nil {
		return err
	}
	if port < portFloor {
		return &config.KeyError{Key: PortThreshold.Key, Reason: fmt.Sprintf(
			"%d is below %d: the serve unit withholds CAP_NET_BIND_SERVICE, so it cannot bind a privileged port", port, portFloor)}
	}
	if port > portCeiling {
		return &config.KeyError{Key: PortThreshold.Key, Reason: fmt.Sprintf("%d is above %d", port, portCeiling)}
	}
	return nil
}

// PrefixThreshold is the one owned path prefix: every route the
// shell registers lives under it, and every path outside it answers
// 404. Its default, /dubplate, is one segment so the segment-aware
// match has a boundary to find, and carries no estate value.
var PrefixThreshold = config.Threshold{
	Key:         "DUBPLATE_SERVE_PREFIX",
	Kind:        config.KindString,
	Default:     "/dubplate",
	Doc:         "a single path segment beginning with a separator, for example /dubplate",
	Description: "the single path segment every page lives under",
	Sensitivity: config.Sensitive,
}

// Register adds the serve shell's three thresholds to reg,
// a plain function: newRootCmd builds a fresh, independent registry on
// every construction, so registering onto it can never collide with a
// prior process's own registration.
func Register(reg *config.Registry) {
	reg.AddThreshold(BindAddrsThreshold)
	reg.AddThreshold(PortThreshold)
	reg.AddThreshold(PrefixThreshold)
}

// Config is the shell's listen configuration, read from the three
// thresholds [Register] adds.
type Config struct {
	Prefix    Prefix
	BindAddrs []netip.Addr
	Port      int
}

// ParseConfig reads the three thresholds against environ. It opens no
// listener, so a caller can refuse an invalid prefix or bind list
// without touching the network.
func ParseConfig(environ []string) (Config, error) {
	prefixValue, err := config.ParseString(environ, PrefixThreshold)
	if err != nil {
		return Config{}, err
	}
	prefix, err := NewPrefix(prefixValue)
	if err != nil {
		return Config{}, err
	}

	bindRaw, err := config.ParseString(environ, BindAddrsThreshold)
	if err != nil {
		return Config{}, err
	}
	addrs, err := parseBindAddrs(bindRaw)
	if err != nil {
		return Config{}, err
	}

	port, err := config.ParseInt(environ, PortThreshold)
	if err != nil {
		return Config{}, err
	}

	return Config{Prefix: prefix, BindAddrs: addrs, Port: port}, nil
}

// ErrInvalidBindAddr is the sentinel [ParseConfig] wraps when an entry
// of the bind address list is refused.
var ErrInvalidBindAddr = errors.New("invalid bind address")

// parseBindAddrs parses raw's comma-separated addresses, each trimmed of
// surrounding space. An empty entry, an entry that is not an IP address,
// and an entry carrying a zone are refused rather than dropped, since a
// dropped entry would bind fewer interfaces than the operator named.
// The unspecified addresses are refused too: each binds every
// interface, which [Listen] never does.
func parseBindAddrs(raw string) ([]netip.Addr, error) {
	parts := strings.Split(raw, ",")
	addrs := make([]netip.Addr, 0, len(parts))
	for _, part := range parts {
		text := strings.TrimSpace(part)
		if text == "" {
			return nil, fmt.Errorf("%q: empty entry: %w", raw, ErrInvalidBindAddr)
		}
		addr, err := netip.ParseAddr(text)
		if err != nil || addr.Zone() != "" {
			return nil, fmt.Errorf("%q: %w", text, ErrInvalidBindAddr)
		}
		if addr.Unmap().IsUnspecified() {
			return nil, fmt.Errorf("%q: the wildcard address binds every interface: %w", text, ErrInvalidBindAddr)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

// The trusted proxy class's two Required keys.
const (
	TrustedProxyAddrsKey  = "DUBPLATE_TRUSTED_PROXY_ADDRS"
	TrustedProxyHeaderKey = "DUBPLATE_TRUSTED_PROXY_HEADER"
)

// TrustedProxyClass is the trusted proxy integration: its
// enable flag, the comma-separated addresses of the fronts permitted
// to report a client address, and the one header they report it in,
// both required while it is on. With the flag unset or false the
// server runs in direct mode, where the TCP peer is the client
// address and no forwarded header is ever read.
var TrustedProxyClass = config.IntegrationClass{
	Name: "trusted proxy",
	Flag: config.KeyDecl{
		Key:         "DUBPLATE_TRUSTED_PROXY_ENABLED",
		Description: "turns on trusted-proxy mode, where a header from a trusted front names the client address",
		Sensitivity: config.Sensitive,
	},
	Required: []config.KeyDecl{
		{Key: TrustedProxyAddrsKey, Description: "the addresses of the fronts permitted to report a client address, comma-separated", Sensitivity: config.Sensitive},
		{Key: TrustedProxyHeaderKey, Description: "the header the trusted fronts report the client address in", Sensitivity: config.Sensitive},
	},
}

// RegisterTrustedProxy adds [TrustedProxyClass] to reg.
func RegisterTrustedProxy(reg *config.Registry) {
	reg.AddIntegrationClass(TrustedProxyClass)
}

// ParseTrustedProxy builds the [TrustedProxyConfig] cfg
// configures: the zero value, which is direct mode, when the class is
// off, and otherwise the parsed address set and the checked header
// name. Serve startup and doctor both call it, so the two cannot
// disagree about a value.
//
// [config.Load] has already refused an unparseable flag, a Required
// key left unset while the flag is on, and either key set while it is
// off. What is left is each value's own grammar: every address entry
// that is empty, is not an IP address, is a network, carries a zone,
// or is unspecified is refused by name, and so is a header
// [ValidateProxyHeader] refuses. Every refusal is reported at once.
func ParseTrustedProxy(cfg *config.Config) (TrustedProxyConfig, error) {
	class := TrustedProxyClass.Name
	if !cfg.IntegrationEnabled(class) {
		return TrustedProxyConfig{}, nil
	}

	addrs, errs := ParseProxyAddrs(cfg.IntegrationValue(class, TrustedProxyAddrsKey))
	header := strings.TrimSpace(cfg.IntegrationValue(class, TrustedProxyHeaderKey))
	if err := ValidateProxyHeader(header); err != nil {
		errs = append(errs, fmt.Errorf("%s %q: %w", TrustedProxyHeaderKey, header, err))
	}
	if len(errs) > 0 {
		return TrustedProxyConfig{}, errors.Join(errs...)
	}
	return NewTrustedProxyConfig(addrs, header), nil
}

// ParseProxyAddrs parses value's comma-separated entries, each trimmed
// of surrounding space and unmapped, and returns one refusal per entry
// it cannot accept.
//
// An entry is unmapped because net/http's RemoteAddr and a forwarding
// header's hop are both compared in unmapped form, and
// [netip.Addr.String] keeps the "::ffff:" prefix a mapped entry was
// written with. Trust is pinned to observed addresses, so a network is
// refused rather than widened to every host in it. A zone names an
// interface on this host, which no peer address carries, and an
// unspecified address is a bind wildcard, never a peer.
func ParseProxyAddrs(value string) ([]netip.Addr, []error) {
	var addrs []netip.Addr
	var errs []error
	for i, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			errs = append(errs, fmt.Errorf("%s: entry %d is empty", TrustedProxyAddrsKey, i+1))
			continue
		}
		addr, reason := parseProxyAddr(entry)
		if reason != "" {
			errs = append(errs, fmt.Errorf("%s %q: %s", TrustedProxyAddrsKey, entry, reason))
			continue
		}
		addrs = append(addrs, addr)
	}
	return addrs, errs
}

// parseProxyAddr parses one non-empty entry, returning its unmapped
// address, or the reason it is refused.
func parseProxyAddr(entry string) (netip.Addr, string) {
	addr, err := netip.ParseAddr(entry)
	if err != nil {
		if _, err := netip.ParsePrefix(entry); err == nil {
			return netip.Addr{}, "an address, not a network"
		}
		return netip.Addr{}, "not an IP address"
	}
	if addr.Zone() != "" {
		return netip.Addr{}, "carries a zone"
	}
	addr = addr.Unmap()
	if addr.IsUnspecified() {
		return netip.Addr{}, "the unspecified address"
	}
	return addr, ""
}
