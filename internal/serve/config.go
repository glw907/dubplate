package serve

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/glw907/dubplate/internal/config"
)

// bindAddrsThreshold registers as a string, since the registry has no
// list kind; [ParseConfig] parses the list.
var bindAddrsThreshold = config.Threshold{
	Key:         "DUBPLATE_SERVE_BIND_ADDRS",
	Kind:        config.KindString,
	Default:     "127.0.0.1",
	Doc:         "a comma-separated list of bind addresses, for example 127.0.0.1,::1",
	Description: "the addresses the server listens on, comma-separated",
	Sensitivity: config.Sensitive,
}

var portThreshold = config.Threshold{
	Key:         "DUBPLATE_SERVE_PORT",
	Kind:        config.KindInt,
	Default:     "8080",
	Doc:         "an integer TCP port from 1024 through 65535",
	Description: "the TCP port the server listens on",
	Sensitivity: config.Sensitive,
}

// portFloor is Linux's default net.ipv4.ip_unprivileged_port_start:
// the serve unit withholds CAP_NET_BIND_SERVICE, so a lower port fails
// at bind, and port 0 would listen wherever the kernel picks.
const (
	portFloor   = 1024
	portCeiling = 65535
)

// ValidatePort checks DUBPLATE_SERVE_PORT against [portFloor] and
// [portCeiling], so doctor and every verb refuse an out-of-range port
// before serve binds. Every refusal is a [config.KeyError].
func ValidatePort(environ []string) error {
	port, err := config.ParseInt(environ, portThreshold)
	if err != nil {
		return err
	}
	if port < portFloor {
		return &config.KeyError{Key: portThreshold.Key, Reason: fmt.Sprintf(
			"%d is below %d: the serve unit withholds CAP_NET_BIND_SERVICE, so it cannot bind a privileged port", port, portFloor)}
	}
	if port > portCeiling {
		return &config.KeyError{Key: portThreshold.Key, Reason: fmt.Sprintf("%d is above %d", port, portCeiling)}
	}
	return nil
}

var prefixThreshold = config.Threshold{
	Key:         "DUBPLATE_SERVE_PREFIX",
	Kind:        config.KindString,
	Default:     "/dubplate",
	Doc:         "a single path segment beginning with a separator, for example /dubplate",
	Description: "the single path segment every page lives under",
	Sensitivity: config.Sensitive,
}

// Register adds the serve shell's three thresholds to reg.
func Register(reg *config.Registry) {
	reg.AddThreshold(bindAddrsThreshold)
	reg.AddThreshold(portThreshold)
	reg.AddThreshold(prefixThreshold)
}

// ParsePrefix reads DUBPLATE_SERVE_PREFIX against environ and returns
// the [Prefix] it names, for a caller that needs the prefix alone.
func ParsePrefix(environ []string) (Prefix, error) {
	value, err := config.ParseString(environ, prefixThreshold)
	if err != nil {
		return Prefix{}, err
	}
	return newPrefix(value)
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
	prefix, err := ParsePrefix(environ)
	if err != nil {
		return Config{}, err
	}

	bindRaw, err := config.ParseString(environ, bindAddrsThreshold)
	if err != nil {
		return Config{}, err
	}
	addrs, err := parseBindAddrs(bindRaw)
	if err != nil {
		return Config{}, err
	}

	port, err := config.ParseInt(environ, portThreshold)
	if err != nil {
		return Config{}, err
	}

	return Config{Prefix: prefix, BindAddrs: addrs, Port: port}, nil
}

var errInvalidBindAddr = errors.New("invalid bind address")

// parseBindAddrs parses raw's comma-separated addresses. An empty,
// unparseable, or zoned entry is refused rather than dropped, which
// would bind fewer interfaces than the operator named, and so is an
// unspecified address, which binds every interface.
func parseBindAddrs(raw string) ([]netip.Addr, error) {
	parts := strings.Split(raw, ",")
	addrs := make([]netip.Addr, 0, len(parts))
	for _, part := range parts {
		text := strings.TrimSpace(part)
		if text == "" {
			return nil, fmt.Errorf("%q: empty entry: %w", raw, errInvalidBindAddr)
		}
		addr, err := netip.ParseAddr(text)
		if err != nil || addr.Zone() != "" {
			return nil, fmt.Errorf("%q: %w", text, errInvalidBindAddr)
		}
		if addr.Unmap().IsUnspecified() {
			return nil, fmt.Errorf("%q: the wildcard address binds every interface: %w", text, errInvalidBindAddr)
		}
		addrs = append(addrs, addr)
	}
	return addrs, nil
}

const (
	trustedProxyAddrsKey  = "DUBPLATE_TRUSTED_PROXY_ADDRS"
	trustedProxyHeaderKey = "DUBPLATE_TRUSTED_PROXY_HEADER"
)

// trustedProxyClass is the trusted proxy integration. With its flag
// off the server runs in direct mode, where no forwarded header is
// ever read.
var trustedProxyClass = config.IntegrationClass{
	Name: "trusted proxy",
	Flag: config.KeyDecl{
		Key:         "DUBPLATE_TRUSTED_PROXY_ENABLED",
		Description: "turns on trusted-proxy mode, where a header from a trusted front names the client address",
		Sensitivity: config.Sensitive,
	},
	Required: []config.KeyDecl{
		{Key: trustedProxyAddrsKey, Description: "the addresses of the fronts permitted to report a client address, comma-separated", Sensitivity: config.Sensitive},
		{Key: trustedProxyHeaderKey, Description: "the header the trusted fronts report the client address in", Sensitivity: config.Sensitive},
	},
}

// RegisterTrustedProxy adds [trustedProxyClass] to reg.
func RegisterTrustedProxy(reg *config.Registry) {
	reg.AddIntegrationClass(trustedProxyClass)
}

// ParseTrustedProxy builds the [TrustedProxyConfig] cfg configures:
// the zero value, direct mode, when the class is off. Serve startup and
// doctor both call it, so the two cannot disagree about a value.
// [config.Load] has already checked the flag and which keys are set;
// this checks each value's grammar and reports every refusal at once.
func ParseTrustedProxy(cfg *config.Config) (TrustedProxyConfig, error) {
	class := trustedProxyClass.Name
	if !cfg.IntegrationEnabled(class) {
		return TrustedProxyConfig{}, nil
	}

	addrs, errs := parseProxyAddrs(cfg.IntegrationValue(class, trustedProxyAddrsKey))
	header := strings.TrimSpace(cfg.IntegrationValue(class, trustedProxyHeaderKey))
	if err := validateProxyHeader(header); err != nil {
		errs = append(errs, fmt.Errorf("%s %q: %w", trustedProxyHeaderKey, header, err))
	}
	if len(errs) > 0 {
		return TrustedProxyConfig{}, errors.Join(errs...)
	}
	return newTrustedProxyConfig(addrs, header), nil
}

// parseProxyAddrs parses value's comma-separated entries, unmapped to
// match [peerAddr] and [parseHop], and returns one refusal per entry it
// cannot accept. Trust is pinned to observed addresses, so a network
// is refused rather than widened to every host in it. A zone, which no
// peer address carries, and an unspecified address, a bind wildcard,
// are refused too.
func parseProxyAddrs(value string) ([]netip.Addr, []error) {
	var addrs []netip.Addr
	var errs []error
	for i, entry := range strings.Split(value, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			errs = append(errs, fmt.Errorf("%s: entry %d is empty", trustedProxyAddrsKey, i+1))
			continue
		}
		addr, reason := parseProxyAddr(entry)
		if reason != "" {
			errs = append(errs, fmt.Errorf("%s %q: %s", trustedProxyAddrsKey, entry, reason))
			continue
		}
		addrs = append(addrs, addr)
	}
	return addrs, errs
}

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
