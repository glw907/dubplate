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

// probeSentinel is the header the probe sends and lists in Connection.
// Only a strip removes it: net/http's server leaves a Connection-listed
// header in place.
const probeSentinel = "Dubplate-Probe-Sentinel"

// probeCounts is the probe route's answer: how many keys of group 1
// and group 2 reached the route, and 1 when the sentinel did. It
// carries counts only, never a header value or an address.
type probeCounts struct {
	Identity  int `json:"identity"`
	Forwarded int `json:"forwarded"`
	Sentinel  int `json:"sentinel"`
}

// newProbeRequest builds the probe request: a GET for the probe route
// under prefix on host, a host:port, carrying every group 1 spelling,
// every group 2 name the strip built for proxy removes, the configured
// client-address header included, and the sentinel listed in
// Connection.
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
// from an unknown path.
//
// A permitted peer is one whose [peerAddr] is loopback or one of
// bindAddrs. [ClientIP] is never consulted, so no forwarded header can
// make a peer local.
//
// A permitted peer that is also a trusted proxy receives the identity
// count alone: the strip keeps group 2 for it by design, so a forwarded
// count would read as a failure it is not. A same-host front relays
// strangers from that address, and what reaches them is that one
// count, zero while the strip works.
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
		peer, ok := peerAddr(r)
		if !ok {
			refuse.ServeHTTP(w, r)
			return
		}
		_, bound := local[peer]
		if !peer.IsLoopback() && !bound {
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

// countPresent returns how many keys of h carry a value and name one
// of names under [foldedKey], so any spelling the strip would match is
// counted. Two spellings of one name count twice.
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

// The boundary probe's bounds, so a host that drops the handshake and a
// listener that never answers both end the row in bounded time.
const (
	probeDialBound     = 2 * time.Second
	probeResponseBound = 2 * time.Second
)

// probeAnswerCap bounds how much of the route's answer doctor reads; the
// counts object is a few dozen bytes.
const probeAnswerCap = 4 << 10

// probeResult is one address's part of the probe row. answered is true
// when the route returned its counts, so the row's trailer applies.
type probeResult struct {
	status   string
	failed   bool
	answered bool
}

// probeAddress probes the server listening on ip and port. The dial
// binds ip as its local address, so the server sees a peer the route
// permits. trusted reports whether ip is one of proxy's addresses.
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

// probeExchange writes the probe request on conn and reads the route's
// counts before ctx's deadline. For a trusted peer only the identity
// count is required. The error is the text the row prints after
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

var errProbeTimeout = errors.New("timeout")

func answerError(err error) error {
	if isTimeout(err) {
		return errProbeTimeout
	}
	return err
}

func isTimeout(err error) bool {
	var ne net.Error
	return errors.As(err, &ne) && ne.Timeout()
}

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
