package main

import (
	"errors"
	"fmt"
	"net/netip"
	"strings"

	"github.com/glw907/dubplate/internal/config"
	"github.com/glw907/dubplate/internal/serve"
)

// The trusted proxy class's two Required keys.
const (
	trustedProxyAddrsKey  = "DUBPLATE_TRUSTED_PROXY_ADDRS"
	trustedProxyHeaderKey = "DUBPLATE_TRUSTED_PROXY_HEADER"
)

// trustedProxyIntegrationClass is the trusted proxy integration: its
// enable flag, the comma-separated addresses of the fronts permitted
// to report a client address, and the one header they report it in,
// both required while it is on. With the flag unset or false the
// server runs in direct mode, where the TCP peer is the client
// address and no forwarded header is ever read.
var trustedProxyIntegrationClass = config.IntegrationClass{
	Name: "trusted proxy",
	Rung: 5,
	Flag: config.KeyDecl{
		Name:        "DUBPLATE_TRUSTED_PROXY_ENABLED",
		Description: "turns on trusted-proxy mode, where a header from a trusted front names the client address",
		Mark:        config.Sensitive,
	},
	Required: []config.KeyDecl{
		{Name: trustedProxyAddrsKey, Description: "the addresses of the fronts permitted to report a client address, comma-separated", Mark: config.Sensitive},
		{Name: trustedProxyHeaderKey, Description: "the header the trusted fronts report the client address in", Mark: config.Sensitive},
	},
}

// registerProxyConfig adds trustedProxyIntegrationClass to reg.
func registerProxyConfig(reg *config.Registry) {
	reg.AddIntegrationClass(trustedProxyIntegrationClass)
}

// resolveTrustedProxy builds the [serve.TrustedProxyConfig] cfg
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
// [serve.CheckProxyHeader] refuses. Every refusal is reported at once.
func resolveTrustedProxy(cfg *config.Config) (serve.TrustedProxyConfig, error) {
	class := trustedProxyIntegrationClass.Name
	if !cfg.IntegrationEnabled(class) {
		return serve.TrustedProxyConfig{}, nil
	}

	addrs, errs := parseProxyAddrs(cfg.IntegrationValue(class, trustedProxyAddrsKey))
	header := strings.TrimSpace(cfg.IntegrationValue(class, trustedProxyHeaderKey))
	if err := serve.CheckProxyHeader(header); err != nil {
		errs = append(errs, fmt.Errorf("%s %q: %w", trustedProxyHeaderKey, header, err))
	}
	if len(errs) > 0 {
		return serve.TrustedProxyConfig{}, errors.Join(errs...)
	}
	return serve.NewTrustedProxyConfig(addrs, header), nil
}

// parseProxyAddrs parses value's comma-separated entries, each trimmed
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
