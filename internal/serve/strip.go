package serve

import (
	"io"
	"net/http"
	"net/textproto"
	"slices"
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
// Names are written in [textproto.CanonicalMIMEHeaderKey] form, which
// is not always the casing the citing project prints. Header names are
// case-insensitive and every removal below canonicalizes both sides, so
// the two spellings name the same header; the citations keep the
// project's own casing.
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
//   - X-Original-Method and X-Rewrite-URL are quoted widely in
//     third-party auth_request material, but no page owned by nginx,
//     ingress-nginx, Traefik, or authentik was found stating either, so
//     both are dropped rather than shipped on recollection.
//
// The security review's second round added these, each with the page
// that states it:
//
//   - X-WEBAUTH-USER: Grafana's auth.proxy default header_name, the
//     header Grafana reads as the authenticated user.
//     https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/auth-proxy/,
//     read 2026-09-08. It is also Gitea's
//     REVERSE_PROXY_AUTHENTICATION_USER default.
//     https://docs.gitea.com/administration/config-cheat-sheet/,
//     read 2026-09-08.
//   - X-authentik-username, X-authentik-email, X-authentik-groups,
//     X-authentik-uid, X-authentik-jwt: authentik's proxy provider sets
//     them on the request the upstream application receives, and its
//     Traefik integration lists them in authResponseHeaders.
//     https://docs.goauthentik.io/add-secure-apps/providers/proxy/ and
//     https://docs.goauthentik.io/add-secure-apps/providers/proxy/server_traefik/,
//     both read 2026-09-08.
//   - Tailscale-User-Login, Tailscale-User-Name: tailscale serve adds
//     them to the request it proxies to the local backend.
//     https://tailscale.com/kb/1312/serve, read 2026-09-08.
//   - X-Original-URL: authentik's nginx forward-auth configuration sets
//     it so the outpost can route the authentication request, which
//     makes it a value a downstream application reads as the request's
//     own target.
//     https://docs.goauthentik.io/add-secure-apps/providers/proxy/server_nginx/,
//     read 2026-09-08.
//   - X-Forwarded-Uri, X-Forwarded-Method: Traefik's ForwardAuth
//     middleware sets both on the request it sends to the
//     authentication server.
//     https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/forwardauth/,
//     read 2026-09-08.
//   - Cf-Access-Authenticated-User-Email, dropped by the previous round
//     as uncitable, is cited now: Cloudflare's own Access tutorial
//     reads it off the incoming request as the email Access
//     authenticated.
//     https://developers.cloudflare.com/cloudflare-one/tutorials/access-workers/,
//     read 2026-09-08.
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
	"X-Forwarded-Uri",
	"X-Forwarded-Method",
	"Cf-Access-Jwt-Assertion",
	"Cf-Access-Authenticated-User-Email",
	"X-Webauth-User",
	"X-Authentik-Username",
	"X-Authentik-Email",
	"X-Authentik-Groups",
	"X-Authentik-Uid",
	"X-Authentik-Jwt",
	"Tailscale-User-Login",
	"Tailscale-User-Name",
	"X-Original-Url",
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
//
// The last four names are the client-address headers real fronts set,
// each cited by the front that sets it. A client reaching the shell
// directly can spell any of them, and a handler that reads one would
// take the client's own claim as the front's:
//
//   - Cf-Connecting-Ip: Cloudflare "provides the client IP address
//     connecting to Cloudflare to the origin web server".
//     True-Client-Ip: the same page's Enterprise-plan equivalent, which
//     "provides the original client IP address to the origin web
//     server".
//     https://developers.cloudflare.com/fundamentals/reference/http-headers/,
//     read 2026-09-08.
//   - Fastly-Client-Ip: Fastly adds it, set to client.ip, when a
//     request arrives without one.
//     https://www.fastly.com/documentation/reference/http/http-headers/Fastly-Client-IP,
//     read 2026-09-08.
//   - X-Azure-Clientip: Azure Front Door appends it to every request it
//     sends to the backend, carrying "the client IP address associated
//     with the request being processed".
//     https://learn.microsoft.com/en-us/azure/frontdoor/front-door-http-headers-protocol,
//     read 2026-09-08.
//
// Two candidates the review named are dropped. X-Client-Ip and
// X-Cluster-Client-Ip are quoted constantly in client-IP-extraction
// libraries, but no page owned by a front that sets either was found,
// so neither ships on recollection. Both stay candidates for a later
// rung that finds a citation.
var forwardedAddressHeaders = []string{
	"X-Forwarded-For",
	"X-Real-Ip",
	"Forwarded",
	"X-Forwarded-Host",
	"X-Forwarded-Proto",
	"X-Forwarded-Prefix",
	"Cf-Connecting-Ip",
	"True-Client-Ip",
	"Fastly-Client-Ip",
	"X-Azure-Clientip",
}

// group3KeepSet is R-S's carve-out: canonical header names
// stripConnectionListed never deletes, whatever a request's Connection
// header lists. Cookie, Origin, and Sec-Fetch-Site are end-to-end
// headers a client has no business declaring hop-by-hop, and 5b's
// origin check reads the latter two.
//
// The set holds header names and nothing else. The shell's own
// session-cookie name was once compared against it as though it were
// one, which it is not: a cookie name lives inside the Cookie header's
// value, so keeping Cookie here is already what protects it, and
// canonicalizing a cookie name as a header key compared two different
// kinds of name.
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

	// forwarded is group 2's removal set: forwardedAddressHeaders plus
	// the trusted-proxy configuration's own header name, derived once at
	// construction. The configured name has to join the set, since an
	// adopter who configures a name outside the fixed family
	// (Cf-Connecting-Ip, True-Client-Ip) would otherwise leave the one
	// header [ClientIP] actually reads free for any client to set.
	forwarded []string

	// forwardedCanon is forwarded in canonical form, the shape
	// stripConnectionListed compares a Connection token against.
	forwardedCanon map[string]struct{}
}

// NewStrip builds a [Strip]. header is the driver's configured
// identity header name, a construction value that is empty at rung 5
// since no driver exists yet; rung 12 supplies it through the driver
// seam. sessionCookie is the shell's own session-cookie name (R-S), a
// construction value that does not exist until 5b; the strip itself
// never removes a cookie, and [Shell] reads this name to remove that
// one cookie on the way out to the fallthrough. proxy
// is Task 1's trusted-proxy configuration and the sole input to group
// 2's carve-out. exemption is the mechanism a later front configures;
// the zero value exempts nothing.
func NewStrip(header, sessionCookie string, proxy TrustedProxyConfig, exemption FrontExemption) Strip {
	forwarded := forwardedAddressSet(proxy.header)
	canon := make(map[string]struct{}, len(forwarded))
	for _, name := range forwarded {
		canon[textproto.CanonicalMIMEHeaderKey(name)] = struct{}{}
	}
	return Strip{
		header:         header,
		sessionCookie:  sessionCookie,
		proxy:          proxy,
		exemption:      exemption,
		forwarded:      forwarded,
		forwardedCanon: canon,
	}
}

// forwardedAddressSet returns group 2's names: the fixed family plus
// configured, the trusted-proxy header an adopter named, when that name
// is not already one of them.
func forwardedAddressSet(configured string) []string {
	if configured == "" {
		return forwardedAddressHeaders
	}
	for _, name := range forwardedAddressHeaders {
		if sameHeaderName(name, configured) {
			return forwardedAddressHeaders
		}
	}
	return append(slices.Clone(forwardedAddressHeaders), configured)
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
// section has actually arrived. The up-front delete removes an
// announced key before the body is read; net/http's mergeSetHeader
// re-adds that key once the trailer's value arrives, so the read-time
// scrub is what removes a value smuggled after the body. The one
// exemption a configured front may carry is honored in both passes
// (R-Q, R-S's sibling for group 1).
//
// The body is wrapped unconditionally, never only when the request
// announced a Trailer header. A chunked request may send a trailer
// section it never announced: r.Trailer is nil at wrap time, and
// mergeSetHeader installs the wire trailer's own map wholesale at
// end of stream. Arming the scrub on the announcement would leave that
// request's identity headers standing, which is the whole strip
// bypassed by omitting one header.
func (s Strip) stripIdentity(r *http.Request) {
	exempt := s.exemption.headerFor(PeerAddress(r))
	if exempt != "" {
		collapseToLast(r.Header, exempt)
	}

	names := identitySpellings(s.header)
	for _, name := range names {
		if exempt != "" && sameHeaderName(name, exempt) {
			continue
		}
		r.Header.Del(name)
		r.Trailer.Del(name)
	}
	scrubTrailerOnRead(r, names, exempt)
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
	r.Body = &trailerScrubbingBody{ReadCloser: r.Body, req: r, names: names, exempt: exempt}
}

// trailerScrubbingBody deletes names, other than exempt, from its
// request's Trailer the moment the wrapped reader signals end of
// stream, which is when net/http has just finished populating that
// map with the values the client sent.
//
// It holds the request rather than the Trailer map it had at wrap
// time. For a request that announced no trailer, r.Trailer is nil
// until end of stream and mergeSetHeader then assigns the wire
// trailer's own map to the field; a captured nil map would leave the
// scrub deleting from nothing while the request carried the values.
type trailerScrubbingBody struct {
	io.ReadCloser
	req    *http.Request
	names  []string
	exempt string
}

func (b *trailerScrubbingBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err == io.EOF {
		b.scrub()
	}
	return n, err
}

// Close closes the wrapped body and scrubs again. net/http's own body
// Close drains the remaining chunked stream inside its own reader
// rather than through this wrapper's Read, so a handler that closes the
// body without reading it reaches end of stream, and mergeSetHeader
// populates r.Trailer, with the Read-armed scrub never having fired.
// Without this method that handler would go on to read a smuggled
// identity value straight out of r.Trailer.
func (b *trailerScrubbingBody) Close() error {
	err := b.ReadCloser.Close()
	b.scrub()
	return err
}

// scrub deletes every name but exempt from the request's Trailer. It is
// idempotent, so the two arming points may both fire on one request.
func (b *trailerScrubbingBody) scrub() {
	for _, name := range b.names {
		if b.exempt != "" && sameHeaderName(name, b.exempt) {
			continue
		}
		b.req.Trailer.Del(name)
	}
}

// stripForwardedAddress removes group 2 unless trusted-proxy mode is
// on and the request's peer address is one of the configured proxies
// (R27's carve-out), read through Task 1's TrustedProxyConfig rather
// than a second derivation of the peer address.
func (s Strip) stripForwardedAddress(r *http.Request) {
	if s.trustsPeer(r) {
		return
	}
	for _, name := range s.forwarded {
		r.Header.Del(name)
	}
}

// trustsPeer reports whether trusted-proxy mode is on and r arrived
// from one of the configured proxies. Groups 2 and 3 both turn on this
// one answer, so they cannot disagree about which peer is the front.
func (s Strip) trustsPeer(r *http.Request) bool {
	return s.proxy.trusts(PeerAddress(r))
}

// stripConnectionListed removes group 3: every header the request's
// own Connection header lists, with two carve-outs (R-O). A listed
// Upgrade keeps both Upgrade and Connection, since a websocket
// handshake sends Connection: Upgrade alongside Upgrade: websocket and
// the listed removal would otherwise delete the handshake it needs.
// group3KeepSet is never removed by this rule (R-S), whatever
// Connection says, and neither is the one header a configured front is
// exempt from group 1 on: an attacker who could name that header in
// Connection would delete the front's own identity assertion, turning
// a header a trusted peer set into a request the shell reads as
// anonymous.
//
// The same argument covers group 2's names when the peer is trusted.
// Group 2 preserves the forwarded-address headers a trusted front sets,
// and a client behind that front can put "X-Forwarded-For" in its own
// Connection header; deleting it here would undo the preservation and
// collapse [ClientIP] to the front's own address, which is the whole
// trusted-proxy carve-out defeated by one token. The keep applies only
// to a trusted peer: from any other peer group 2 has already deleted
// those names, so there is nothing left for this rule to keep.
func (s Strip) stripConnectionListed(r *http.Request) {
	tokens := connectionTokens(r.Header.Values("Connection"))
	if len(tokens) == 0 {
		return
	}

	exempt := ""
	if header := s.exemption.headerFor(PeerAddress(r)); header != "" {
		exempt = textproto.CanonicalMIMEHeaderKey(header)
	}
	keepForwarded := s.trustsPeer(r)

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
		if keepForwarded {
			if _, keep := s.forwardedCanon[canon]; keep {
				continue
			}
		}
		if exempt != "" && canon == exempt {
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
