package serve

import (
	"context"
	"encoding/json"
	"net/http"
	"net/netip"
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
// strip exists to remove reached the route. Identity counts the keys
// group 1 matches, Forwarded counts group 2's, and Sentinel is 1 when
// the Connection-listed sentinel arrived. It carries counts only, never
// a header value or an address. A trusted peer receives the identity
// count alone, as [NewProbeRoute] describes.
type ProbeCounts struct {
	Identity  int `json:"identity"`
	Forwarded int `json:"forwarded"`
	Sentinel  int `json:"sentinel"`
}

// NewProbeRequest builds the one probe request every caller sends: a
// GET for the probe route under prefix on host, carrying every group 1
// spelling in both forms, every group 2 name the strip built for proxy
// removes, the configured client-address header included, and the
// sentinel listed in Connection. Each header is added with
// [http.Header.Add], which canonicalizes the name exactly as the
// receiving server's parser does, so the request carries one canonical
// key per distinct name. host is a host:port the URL can carry.
func NewProbeRequest(ctx context.Context, prefix Prefix, host string, proxy TrustedProxyConfig) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+host+prefix.String()+"/"+ProbeRouteName, nil)
	if err != nil {
		return nil, err
	}
	for _, name := range identitySpellings {
		req.Header.Add(name, "probe")
	}
	for _, name := range forwardedAddressSet(proxy.header) {
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
// a loopback address or one of bindAddrs. bindAddrs is the resolved
// list the server listens on, each parsed and unmapped the same way,
// since net/http reports an IPv4 peer of a dual-stack listener in
// mapped form; an entry that does not parse permits no peer.
// [ClientIP] is never consulted, so no forwarded header can make a
// peer local.
//
// A permitted peer that is also one of proxy's trusted proxies
// receives a JSON object holding the identity count alone. The strip
// keeps group 2 for that peer by design, so a forwarded count would
// read as a failure it is not, and group 1 applies to every peer, so
// the identity count still proves the strip. A same-host front relays
// strangers from that address, and what reaches them is that one
// count, zero while the strip works.
//
// The forwarded count reads the strip's own group 2 set for proxy, the
// configured client-address header included.
func NewProbeRoute(bindAddrs []string, proxy TrustedProxyConfig) http.Handler {
	local := make(map[netip.Addr]struct{}, len(bindAddrs))
	for _, a := range bindAddrs {
		if addr, err := netip.ParseAddr(a); err == nil {
			local[addr.Unmap()] = struct{}{}
		}
	}
	identity := identitySpellings
	forwarded := forwardedAddressSet(proxy.header)
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
		if !addr.IsLoopback() && !bound {
			refuse.ServeHTTP(w, r)
			return
		}
		if r.Method != http.MethodGet {
			refuse.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-store")
		if proxy.trusts(peer) {
			_ = json.NewEncoder(w).Encode(struct {
				Identity int `json:"identity"`
			}{countPresent(r.Header, identity)})
			return
		}
		_ = json.NewEncoder(w).Encode(ProbeCounts{
			Identity:  countPresent(r.Header, identity),
			Forwarded: countPresent(r.Header, forwarded),
			Sentinel:  countPresent(r.Header, []string{probeSentinel}),
		})
	})
}

// countPresent returns how many keys of h carry at least one value and
// name one of names under [foldedKey], the comparison group 1's removal
// uses, so any spelling the strip would match is a spelling the count
// sees. Two spellings of one name are two keys and count twice.
func countPresent(h http.Header, names []string) int {
	want := make(map[string]struct{}, len(names))
	for _, name := range names {
		want[foldedKey(name)] = struct{}{}
	}
	n := 0
	for k, values := range h {
		if _, ok := want[foldedKey(k)]; ok && len(values) > 0 {
			n++
		}
	}
	return n
}
