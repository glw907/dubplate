package serve

import (
	"io"
	"net/http"
	"net/textproto"
	"strings"
)

// identityExtras is the fixed set of trusted-header-SSO header names
// group 1 removes in addition to the driver's configured header name.
// Rung 12 declares a driver's header through the seam [NewStrip] reads
// at construction; this list never gains a driver-specific name of its
// own. Every name here is documented by the project that owns it as a
// header it sets or reads to carry an authenticated identity into a
// backend application; identitySpellings derives each name's
// underscore twin, so a new entry is one line.
//
// Every candidate this task could cite, and every one it dropped:
//
//   - Remote-User, Remote-Name, Remote-Email, Remote-Groups: Authelia's
//     trusted-header-SSO response headers, injected by the reverse
//     proxy into the request the backend application receives.
//     https://www.authelia.com/integration/trusted-header-sso/introduction/,
//     read 2026-09-08.
//   - Remote-User is also Navidrome's own default ExtAuth.UserHeader,
//     the header the streamer rung 12 proxies reads for externalized
//     authentication.
//     https://www.navidrome.org/docs/usage/integration/authentication/,
//     read 2026-09-08.
//   - X-Forwarded-User, X-Forwarded-Email, X-Forwarded-Preferred-Username,
//     X-Forwarded-Groups: oauth2-proxy's pass_user_headers set (on by
//     default), forwarded to the upstream application.
//     X-Forwarded-Access-Token: oauth2-proxy's pass_access_token
//     header, carrying the OAuth access token to the upstream
//     application.
//     https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview,
//     read 2026-09-08.
//   - Cf-Access-Jwt-Assertion: Cloudflare Access's documented
//     application-token request header, set on every request Access
//     forwards to the origin.
//     https://developers.cloudflare.com/cloudflare-one/identity/authorization-cookie/validating-json/,
//     read 2026-09-08.
//
// Dropped candidates, each with the reason it was not admitted:
//
//   - Miniflux's AUTH_PROXY_HEADER names no fixed header at all; its
//     own documentation states the value is empty by default and
//     adopter-configured, so it contributes no name here. Its explicit
//     pairing with TRUSTED_REVERSE_PROXY_NETWORKS is the shape R27
//     already takes, not a header spelling.
//     https://miniflux.app/docs/configuration.html, read 2026-09-08.
//   - oauth2-proxy's X-Auth-Request-User, X-Auth-Request-Email,
//     X-Auth-Request-Groups, X-Auth-Request-Preferred-Username, and
//     X-Auth-Request-Access-Token are documented as response headers
//     oauth2-proxy returns to the proxy under --set-xauthrequest for
//     Nginx's auth_request mode, not headers the project documents as
//     arriving on an inbound request to the backend application.
//   - Cf-Access-Client-Id and Cf-Access-Client-Secret authenticate a
//     Cloudflare Access service token: a caller who lacks the secret
//     cannot forge them, so removing them defends against nothing this
//     strip is for and would break service-token integrations.
//   - Cf-Access-Authenticated-User-Email is widely quoted in
//     third-party material, but this task found no page under
//     developers.cloudflare.com stating it, so it is dropped rather
//     than shipped on recollection.
var identityExtras = []string{
	"Remote-User",
	"Remote-Name",
	"Remote-Email",
	"Remote-Groups",
	"X-Forwarded-User",
	"X-Forwarded-Email",
	"X-Forwarded-Preferred-Username",
	"X-Forwarded-Groups",
	"X-Forwarded-Access-Token",
	"Cf-Access-Jwt-Assertion",
}

// forwardedAddressHeaders is group 2: the address- and origin-asserting
// headers a client could otherwise use to spoof its own network
// position. Authelia's forwarded-headers documentation requires a
// reverse proxy to remove and replace X-Forwarded-For arriving from an
// untrusted client
// (https://www.authelia.com/integration/proxies/forwarded-headers/,
// read 2026-09-08). R-R extends the rule to the whole X-Forwarded-*
// family, since each remaining member steers an absolute URL a
// downstream handler builds from client-asserted input.
var forwardedAddressHeaders = []string{
	"X-Forwarded-For",
	"X-Real-Ip",
	"Forwarded",
	"X-Forwarded-Host",
	"X-Forwarded-Proto",
	"X-Forwarded-Prefix",
}

// group3KeepSet is R-S's carve-out: canonical names stripConnectionListed
// never deletes, whatever a request's Connection header lists. Cookie,
// Origin, and Sec-Fetch-Site are end-to-end headers a client has no
// business declaring hop-by-hop, and 5b's origin check reads the
// latter two.
var group3KeepSet = map[string]struct{}{
	"Cookie":         {},
	"Origin":         {},
	"Sec-Fetch-Site": {},
}

// identitySpellings returns every name group 1 removes: header (the
// driver's configured identity header name, empty at rung 5) and every
// name in identityExtras, each in its dash form and, where a dash
// makes the two forms distinct, its underscore form too. Both
// spellings come from one source, so a new name is one line of data.
func identitySpellings(header string) []string {
	names := make([]string, 0, len(identityExtras)+1)
	if header != "" {
		names = append(names, header)
	}
	names = append(names, identityExtras...)

	spellings := make([]string, 0, len(names)*2)
	for _, name := range names {
		spellings = append(spellings, name)
		if underscored := strings.ReplaceAll(name, "-", "_"); underscored != name {
			spellings = append(spellings, underscored)
		}
	}
	return spellings
}

// FrontExemption names the one identity header spelling a configured
// front may exempt from group 1's removal, and the peer addresses from
// which the exemption applies. The zero value exempts nothing, which
// is every deployment rung 5 ships: no front exists before rung 12
// (spec:849-862).
type FrontExemption struct {
	header  string
	proxies map[string]struct{}
}

// NewFrontExemption builds a [FrontExemption] that exempts header, in
// its dash form only, from group 1's removal when the request's peer
// address is one of proxies. Every other identity spelling, header's
// underscore twin included, is still stripped. An empty header exempts
// nothing.
func NewFrontExemption(header string, proxies []string) FrontExemption {
	set := make(map[string]struct{}, len(proxies))
	for _, p := range proxies {
		set[p] = struct{}{}
	}
	return FrontExemption{header: header, proxies: set}
}

// headerFor returns the header this exemption exempts for a request
// whose peer address is peer, or "" when the exemption is empty or
// peer is not one of its configured proxies.
func (e FrontExemption) headerFor(peer string) string {
	if e.header == "" {
		return ""
	}
	if _, trusted := e.proxies[peer]; !trusted {
		return ""
	}
	return e.header
}

// Strip is the rung 5 ingress strip (C2): a wrapper that removes every
// identity-header spelling, the forwarded-address family, and any
// header a request's own Connection header lists, before the request
// reaches the mux. It clears both r.Header and r.Trailer, so a value
// smuggled in an announced trailer is removed on the same pass as one
// arriving in a header field (R-P). Task 5 composes Strip as the
// outermost step; Strip wraps the mux, never a handler, which is what
// makes the removal hold for every path without enumerating them.
type Strip struct {
	header        string
	sessionCookie string
	proxy         TrustedProxyConfig
	exemption     FrontExemption
}

// NewStrip builds a [Strip]. header is the driver's configured
// identity header name, a construction value that is empty at rung 5
// since no driver exists yet; rung 12 supplies it through the driver
// seam. sessionCookie is the shell's own session-cookie name (R-S), a
// construction value that does not exist until 5b; a Strip built with
// no session-cookie name adds nothing beyond the fixed keep-set. proxy
// is Task 1's trusted-proxy configuration and the sole input to group
// 2's carve-out. exemption is the mechanism a later front configures;
// the zero value exempts nothing.
func NewStrip(header, sessionCookie string, proxy TrustedProxyConfig, exemption FrontExemption) Strip {
	return Strip{header: header, sessionCookie: sessionCookie, proxy: proxy, exemption: exemption}
}

// Wrap returns an [http.Handler] that clears the strip's three groups
// from r before calling next.
func (s Strip) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.stripIdentity(r)
		s.stripForwardedAddress(r)
		s.stripConnectionListed(r)
		next.ServeHTTP(w, r)
	})
}

// stripIdentity removes group 1: every identity spelling from r.Header
// and r.Trailer on the same pass (R-P), then wraps r.Body so the same
// names are deleted from r.Trailer again once the client's trailer
// section has actually arrived. The up-front delete removes the
// announced key before the body is read; net/http's mergeSetHeader
// re-adds that key once the trailer's value arrives, so the read-time
// scrub is what removes a value smuggled after the body. The one
// exemption a configured front may carry is honored in both passes
// (R-Q, R-S's sibling for group 1).
func (s Strip) stripIdentity(r *http.Request) {
	exempt := s.exemption.headerFor(PeerAddress(r))
	if exempt != "" {
		collapseToLast(r.Header, exempt)
	}

	// A request that announces a Trailer name populates r.Trailer with
	// that name up front, mapped to nil; the length check has to run
	// before the deletion loop empties the map, or scrubTrailerOnRead
	// below would wrongly conclude no trailer was announced at all.
	announcedTrailer := len(r.Trailer) > 0

	names := identitySpellings(s.header)
	for _, name := range names {
		if exempt != "" && sameHeaderName(name, exempt) {
			continue
		}
		r.Header.Del(name)
		r.Trailer.Del(name)
	}
	if announcedTrailer {
		scrubTrailerOnRead(r, names, exempt)
	}
}

// collapseToLast rewrites h so name carries only the last value among
// any it currently has (R-Q). A front that appends rather than
// replaces produces a request carrying the exempted name twice, and
// Header.Get returns the first of those; this rule decides which of
// the two a downstream handler sees.
func collapseToLast(h http.Header, name string) {
	values := h.Values(name)
	if len(values) < 2 {
		return
	}
	h.Set(name, values[len(values)-1])
}

func sameHeaderName(a, b string) bool {
	return textproto.CanonicalMIMEHeaderKey(a) == textproto.CanonicalMIMEHeaderKey(b)
}

// scrubTrailerOnRead wraps r.Body so that once the client's trailer
// section has arrived, every name in names, other than exempt, is
// deleted from r.Trailer again. net/http's mergeSetHeader (transfer.go)
// copies the wire trailer's values into r.Trailer unconditionally once
// the body reaches io.EOF, which re-adds any name this task's up-front
// delete removed; wrapping the body is the one point that reliably runs
// after that merge.
func scrubTrailerOnRead(r *http.Request, names []string, exempt string) {
	if r.Body == nil {
		return
	}
	r.Body = &trailerScrubbingBody{ReadCloser: r.Body, trailer: r.Trailer, names: names, exempt: exempt}
}

// trailerScrubbingBody deletes names, other than exempt, from trailer
// the moment the wrapped reader signals end of stream, which is when
// net/http has just finished populating trailer with the values the
// client sent.
type trailerScrubbingBody struct {
	io.ReadCloser
	trailer http.Header
	names   []string
	exempt  string
}

func (b *trailerScrubbingBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err == io.EOF {
		for _, name := range b.names {
			if b.exempt != "" && sameHeaderName(name, b.exempt) {
				continue
			}
			b.trailer.Del(name)
		}
	}
	return n, err
}

// stripForwardedAddress removes group 2 unless trusted-proxy mode is
// on and the request's peer address is one of the configured proxies
// (R27's carve-out), read through Task 1's TrustedProxyConfig rather
// than a second derivation of the peer address.
func (s Strip) stripForwardedAddress(r *http.Request) {
	if s.proxy.Mode() == ModeTrustedProxy {
		if _, trusted := s.proxy.proxies[PeerAddress(r)]; trusted {
			return
		}
	}
	for _, name := range forwardedAddressHeaders {
		r.Header.Del(name)
	}
}

// stripConnectionListed removes group 3: every header the request's
// own Connection header lists, with two carve-outs (R-O). A listed
// Upgrade keeps both Upgrade and Connection, since a websocket
// handshake sends Connection: Upgrade alongside Upgrade: websocket and
// the listed removal would otherwise delete the handshake it needs.
// group3KeepSet and the shell's own session-cookie name are never
// removed by this rule (R-S), whatever Connection says.
func (s Strip) stripConnectionListed(r *http.Request) {
	tokens := connectionTokens(r.Header.Values("Connection"))
	if len(tokens) == 0 {
		return
	}

	upgrade := false
	for _, token := range tokens {
		if textproto.CanonicalMIMEHeaderKey(token) == "Upgrade" {
			upgrade = true
			break
		}
	}

	for _, token := range tokens {
		canon := textproto.CanonicalMIMEHeaderKey(token)
		if upgrade && (canon == "Upgrade" || canon == "Connection") {
			continue
		}
		if _, keep := group3KeepSet[canon]; keep {
			continue
		}
		if s.sessionCookie != "" && canon == textproto.CanonicalMIMEHeaderKey(s.sessionCookie) {
			continue
		}
		r.Header.Del(token)
	}
}

// connectionTokens flattens every value a request sent under the
// Connection key (a client may repeat the header, and each instance
// may itself be comma-separated) into the individual header names it
// lists.
func connectionTokens(values []string) []string {
	var tokens []string
	for _, v := range values {
		for _, hop := range splitHops(v) {
			if token := trimHop(hop); token != "" {
				tokens = append(tokens, token)
			}
		}
	}
	return tokens
}
