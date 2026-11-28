package serve

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// ProbeRouteName is the boundary probe's subtree name under the owned
// prefix.
const ProbeRouteName = "boundary-probe"

// probeSentinel is the header the probe request sends and lists in its
// own Connection header. Only a strip removes it: net/http's server
// leaves a Connection-listed header in place, which the route's
// no-strip test proves over a real socket.
const probeSentinel = "Dubplate-Probe-Sentinel"

// probeCounts is the probe route's answer: how many of the headers the
// strip exists to remove reached the route. Identity counts the keys
// group 1 matches, Forwarded counts group 2's, and Sentinel is 1 when
// the Connection-listed sentinel arrived. It carries counts only, never
// a header value or an address. A trusted peer receives the identity
// count alone, as [NewProbeRoute] describes.
type probeCounts struct {
	Identity  int `json:"identity"`
	Forwarded int `json:"forwarded"`
	Sentinel  int `json:"sentinel"`
}

// newProbeRequest builds the one probe request every caller sends: a
// GET for the probe route under prefix on host, carrying every group 1
// spelling in both forms, every group 2 name the strip built for proxy
// removes, the configured client-address header included, and the
// sentinel listed in Connection. Each header is added with
// [http.Header.Add], which canonicalizes the name exactly as the
// receiving server's parser does, so the request carries one canonical
// key per distinct name. host is a host:port the URL can carry.
func newProbeRequest(ctx context.Context, prefix Prefix, host string, proxy TrustedProxyConfig) (*http.Request, error) {
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

// NewProbeRoute returns the boundary probe route's handler, mounted at
// [ProbeRouteName] under prefix. It answers a GET for exactly that path
// from a permitted peer with the [probeCounts] of the request it
// received, and every other request, from any peer, with
// [http.NotFoundHandler]'s 404, so a stranger cannot tell the route
// from an unknown path. The path is checked first, then the peer, then
// the method.
//
// A permitted peer is one whose [peerIP], parsed and unmapped, is
// a loopback address or one of bindAddrs, the addresses the server
// listens on, each unmapped the same way, since net/http reports an
// IPv4 peer of a dual-stack listener in mapped form. [ClientIP] is
// never consulted, so no forwarded header can make a peer local.
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
func NewProbeRoute(prefix Prefix, bindAddrs []netip.Addr, proxy TrustedProxyConfig) http.Handler {
	path := prefix.String() + "/" + ProbeRouteName
	local := make(map[netip.Addr]struct{}, len(bindAddrs))
	for _, a := range bindAddrs {
		local[a.Unmap()] = struct{}{}
	}
	identity := identitySpellings
	forwarded := forwardedAddressSet(proxy.header)
	refuse := http.NotFoundHandler()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			refuse.ServeHTTP(w, r)
			return
		}
		peer := peerIP(r)
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
		_ = json.NewEncoder(w).Encode(probeCounts{
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

// The boundary probe's bounds: the dial and the response each get a
// bound, so a host that drops the handshake and a listener that
// accepts and never answers both end the row in bounded time.
const (
	probeDialBound     = 2 * time.Second
	probeResponseBound = 2 * time.Second
)

// probeAnswerCap bounds how much of the route's answer doctor reads.
// The counts object is a few dozen bytes; anything past this is not
// the route's answer.
const probeAnswerCap = 4 << 10

// probeResult is one address's part of the probe row.
type probeResult struct {
	status string
	failed bool
	// answered is true when the route returned its counts, so the
	// row's trailer applies.
	answered bool
}

// probeAddress probes the server listening on ip and port with the
// probe request built for proxy. The dial binds ip as its local
// address, so the peer the server sees is the address it was
// configured to listen on, which the route permits. trusted reports
// whether ip is one of proxy's addresses, whose answer carries the
// identity count alone.
func probeAddress(ctx context.Context, prefix Prefix, ip netip.Addr, port int, proxy TrustedProxyConfig, trusted bool) probeResult {
	dialer := net.Dialer{
		Timeout:   probeDialBound,
		LocalAddr: &net.TCPAddr{IP: ip.AsSlice()},
	}
	host := net.JoinHostPort(ip.String(), strconv.Itoa(port))
	conn, err := dialer.DialContext(ctx, "tcp", host)
	switch {
	case err == nil:
	case errors.Is(err, syscall.ECONNREFUSED), isTimeout(err):
		return probeResult{status: "no listener"}
	case errors.Is(err, syscall.EADDRNOTAVAIL):
		return probeResult{status: "not probed (not on any interface)"}
	default:
		return probeResult{status: "unexpected dial error: " + err.Error(), failed: true}
	}
	defer func() { _ = conn.Close() }()

	exchange, cancel := context.WithTimeout(ctx, probeResponseBound)
	defer cancel()
	counts, err := probeExchange(exchange, conn, prefix, host, proxy, trusted)
	if err != nil {
		return probeResult{status: "unexpected answer: " + err.Error(), failed: true}
	}
	return countsResult(counts)
}

// probeExchange writes the probe request built for proxy on conn and
// reads the route's counts, all before ctx's deadline: the caller
// bounds the exchange by bounding ctx. For a trusted peer only the
// identity count is required, and the other two read as zero whatever
// the body holds. The error it returns is the text the row prints after
// "unexpected answer: ".
func probeExchange(ctx context.Context, conn net.Conn, prefix Prefix, host string, proxy TrustedProxyConfig, trusted bool) (probeCounts, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return probeCounts{}, err
		}
	}
	req, err := newProbeRequest(ctx, prefix, host, proxy)
	if err != nil {
		return probeCounts{}, err
	}
	if err := req.Write(conn); err != nil {
		return probeCounts{}, answerError(err)
	}
	resp, err := http.ReadResponse(bufio.NewReader(conn), req)
	if err != nil {
		return probeCounts{}, answerError(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return probeCounts{}, errors.New(resp.Status)
	}
	unparseable := fmt.Errorf("%s with an unparseable body", resp.Status)
	var answer probeAnswer
	if err := json.NewDecoder(io.LimitReader(resp.Body, probeAnswerCap)).Decode(&answer); err != nil {
		if isTimeout(err) {
			return probeCounts{}, errProbeTimeout
		}
		return probeCounts{}, unparseable
	}
	if answer.Identity == nil {
		return probeCounts{}, unparseable
	}
	if trusted {
		return probeCounts{Identity: *answer.Identity}, nil
	}
	if answer.Forwarded == nil || answer.Sentinel == nil {
		return probeCounts{}, unparseable
	}
	return probeCounts{
		Identity:  *answer.Identity,
		Forwarded: *answer.Forwarded,
		Sentinel:  *answer.Sentinel,
	}, nil
}

// probeAnswer is [probeCounts] as the probe reads it, with each count a
// pointer so a key the body leaves out stays nil. A 200 whose body
// lacks any count, or is JSON null, did not come from the route, and
// reading it as all zeros would report a strip that never ran.
type probeAnswer struct {
	Identity  *int `json:"identity"`
	Forwarded *int `json:"forwarded"`
	Sentinel  *int `json:"sentinel"`
}

// errProbeTimeout is the answer error for a connection that accepted
// and did not answer within [probeResponseBound].
var errProbeTimeout = errors.New("timeout")

// answerError maps a deadline expiry to [errProbeTimeout] and passes
// any other error through.
func answerError(err error) error {
	if isTimeout(err) {
		return errProbeTimeout
	}
	return err
}

// isTimeout reports whether err is a [net.Error] that timed out.
func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

// countsResult renders the route's counts: "refused" when every
// count is zero, and otherwise each group that reached the route,
// which fails the row.
func countsResult(c probeCounts) probeResult {
	var passed []string
	if c.Identity > 0 {
		passed = append(passed, fmt.Sprintf("%d identity header(s)", c.Identity))
	}
	if c.Forwarded > 0 {
		passed = append(passed, fmt.Sprintf("%d forwarded-address header(s)", c.Forwarded))
	}
	if c.Sentinel > 0 {
		passed = append(passed, "the Connection-listed sentinel")
	}
	if len(passed) == 0 {
		return probeResult{status: "refused", answered: true}
	}
	return probeResult{status: "passed " + strings.Join(passed, " and "), failed: true, answered: true}
}
