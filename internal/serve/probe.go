package serve

import (
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
	"net/textproto"
)

// ProbeRouteName is the boundary probe's subtree name under the owned
// prefix.
const ProbeRouteName = "boundary-probe"

// probeSentinel is the header the probe request sends and lists in its
// own Connection header. Only a strip removes it: net/http's server
// leaves a Connection-listed header in place, which the route's
// no-strip test proves over a real socket.
const probeSentinel = "Dubplate-Probe-Sentinel"

// ProbeCounts is the probe route's answer: how many of the headers the
// strip exists to remove reached the route. Identity counts group 1's
// distinct canonical names, Forwarded counts group 2's, and Sentinel is
// 1 when the Connection-listed sentinel arrived. It carries counts
// only, never a header value or an address.
type ProbeCounts struct {
	Identity  int `json:"identity"`
	Forwarded int `json:"forwarded"`
	Sentinel  int `json:"sentinel"`
}

// probeIdentityNames is group 1 as the probe sends and counts it: every
// identity spelling the strip removes while no driver header is
// configured, which is every deployment the shell serves today.
func probeIdentityNames() []string {
	return identitySpellings("")
}

// NewProbeRequest builds the one probe request every caller sends: a
// GET for the probe route under prefix on host, carrying every group 1
// spelling in both forms, every group 2 name, and the sentinel listed
// in Connection. Each header is added with [http.Header.Add], which
// canonicalizes the name exactly as the receiving server's parser does,
// so the request carries one canonical key per distinct name. host is a
// host:port the URL can carry.
func NewProbeRequest(ctx context.Context, prefix Prefix, host string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+host+prefix.String()+"/"+ProbeRouteName, nil)
	if err != nil {
		return nil, err
	}
	for _, name := range probeIdentityNames() {
		req.Header.Add(name, "probe")
	}
	for _, name := range forwardedAddressHeaders {
		req.Header.Add(name, "probe")
	}
	req.Header.Add(probeSentinel, "probe")
	req.Header.Add("Connection", probeSentinel)
	return req, nil
}

// NewProbeRoute returns the boundary probe route's handler. It answers
// a GET from a permitted peer with the [ProbeCounts] of the request it
// received, and every other request, from any peer, with
// [Terminator]'s 404, so a stranger cannot tell the route from an
// unknown path. The peer is checked before the method.
//
// A permitted peer is one whose [PeerAddress], parsed and unmapped, is
// a loopback address or one of bindAddrs, and is not one of proxy's
// trusted proxies. bindAddrs is the resolved list the server listens
// on, each parsed and unmapped the same way, since net/http reports an
// IPv4 peer of a dual-stack listener in mapped form; an entry that does
// not parse permits no peer. [ClientIP] is never consulted, so no
// forwarded header can make a peer local, and excluding the trusted
// set keeps a configured same-host proxy from relaying strangers to
// the route.
//
// Each count reads [http.Header.Values], which canonicalizes the name
// it is given, so an underscore spelling is counted under the same
// canonical key the server's parser stored it under.
func NewProbeRoute(bindAddrs []string, proxy TrustedProxyConfig) http.Handler {
	local := make(map[netip.Addr]struct{}, len(bindAddrs))
	for _, a := range bindAddrs {
		if addr, err := netip.ParseAddr(a); err == nil {
			local[addr.Unmap()] = struct{}{}
		}
	}
	identity := probeIdentityNames()
	refuse := Terminator()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		peer := PeerAddress(r)
		addr, err := netip.ParseAddr(peer)
		if err != nil {
			refuse.ServeHTTP(w, r)
			return
		}
		addr = addr.Unmap()
		_, bound := local[addr]
		if (!addr.IsLoopback() && !bound) || proxy.trusts(peer) {
			refuse.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodGet {
			refuse.ServeHTTP(w, r)
			return
		}

		counts := ProbeCounts{
			Identity:  countPresent(r.Header, identity),
			Forwarded: countPresent(r.Header, forwardedAddressHeaders),
			Sentinel:  countPresent(r.Header, []string{probeSentinel}),
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		_ = json.NewEncoder(w).Encode(counts)
	})
}

// countPresent returns how many distinct canonical names among names
// carry at least one value in h.
func countPresent(h http.Header, names []string) int {
	present := make(map[string]struct{}, len(names))
	for _, name := range names {
		if len(h.Values(name)) > 0 {
			present[textproto.CanonicalMIMEHeaderKey(name)] = struct{}{}
		}
	}
	return len(present)
}
