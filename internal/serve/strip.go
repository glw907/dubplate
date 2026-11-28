package serve

import (
	"io"
	"net/http"
	"net/textproto"
	"slices"
	"strings"
)

// identityHeaders is group 1: the trusted-header-SSO names the project
// that owns each documents as carrying an authenticated identity into a
// backend. Names are in [textproto.CanonicalMIMEHeaderKey] form, and
// identitySpellings adds each one's underscore twin.
//
//   - Remote-User, -Name, -Email, -Groups: Authelia,
//     https://www.authelia.com/integration/trusted-header-sso/introduction/
//   - X-Forwarded-User, -Email, -Preferred-Username, -Groups, and
//     -Access-Token: oauth2-proxy's pass_user_headers and
//     pass_access_token,
//     https://oauth2-proxy.github.io/oauth2-proxy/configuration/overview
//   - X-Forwarded-Uri, X-Forwarded-Method: Traefik ForwardAuth,
//     https://doc.traefik.io/traefik/reference/routing-configuration/http/middlewares/forwardauth/
//   - Cf-Access-Jwt-Assertion, Cf-Access-Authenticated-User-Email:
//     Cloudflare Access,
//     https://developers.cloudflare.com/cloudflare-one/identity/authorization-cookie/validating-json/
//     and https://developers.cloudflare.com/cloudflare-one/tutorials/access-workers/
//   - X-Webauth-User: Grafana's auth.proxy and Gitea's reverse-proxy
//     authentication,
//     https://grafana.com/docs/grafana/latest/setup-grafana/configure-security/configure-authentication/auth-proxy/
//     and https://docs.gitea.com/administration/config-cheat-sheet/
//   - X-Authentik-Username, -Email, -Groups, -Uid, -Jwt: authentik's
//     proxy provider, https://docs.goauthentik.io/add-secure-apps/providers/proxy/
//   - X-Original-Url: authentik's nginx forward-auth, which a backend
//     reads as the request's own target,
//     https://docs.goauthentik.io/add-secure-apps/providers/proxy/server_nginx/
//   - Tailscale-User-Login, Tailscale-User-Name: tailscale serve,
//     https://tailscale.com/kb/1312/serve
//
// Cf-Access-Client-Id and Cf-Access-Client-Secret are left out: they
// authenticate a service token a forger cannot hold, so removing them
// defends nothing and breaks service-token integrations. oauth2-proxy's
// X-Auth-Request-* names are left out too, since they are response
// headers to the proxy, not request headers to a backend. A name ships
// only when a page owned by its project states it, which keeps out
// X-Original-Method and X-Rewrite-Url.
var identityHeaders = []string{
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

// forwardedAddressHeaders is group 2: the headers asserting a client's
// address or origin, which a client reaching the shell directly could
// otherwise spoof. Authelia requires a proxy to replace an untrusted
// client's X-Forwarded-For
// (https://www.authelia.com/integration/proxies/forwarded-headers/), and
// the strip extends that to the X-Forwarded-* family, since each member
// steers an absolute URL a handler builds. The last four are the
// client-address headers fronts set:
//
//   - Cf-Connecting-Ip, True-Client-Ip: Cloudflare,
//     https://developers.cloudflare.com/fundamentals/reference/http-headers/
//   - Fastly-Client-Ip: Fastly,
//     https://www.fastly.com/documentation/reference/http/http-headers/Fastly-Client-IP
//   - X-Azure-Clientip: Azure Front Door,
//     https://learn.microsoft.com/en-us/azure/frontdoor/front-door-http-headers-protocol
//
// X-Client-Ip and X-Cluster-Client-Ip are left out: no front that sets
// either documents it.
var forwardedAddressHeaders = []forwardedHeader{
	{name: "X-Forwarded-For"},
	{name: "X-Real-Ip"},
	{name: "Forwarded", refusal: "its for= parameters are not an address list, so every hop would be malformed"},
	{name: "X-Forwarded-Host", refusal: "a forwarded header that carries the original host and no client address, so every hop would be malformed"},
	{name: "X-Forwarded-Proto", refusal: "a forwarded header that carries the original scheme and no client address, so every hop would be malformed"},
	{name: "X-Forwarded-Prefix", refusal: "a forwarded header that carries the original path prefix and no client address, so every hop would be malformed"},
	{name: "Cf-Connecting-Ip"},
	{name: "True-Client-Ip"},
	{name: "Fastly-Client-Ip"},
	{name: "X-Azure-Clientip"},
}

// forwardedHeader is one group 2 name in canonical form. refusal is
// why [validateProxyHeader] refuses the name, set for a header that
// carries no client address list, and empty for one that does.
type forwardedHeader struct {
	name    string
	refusal string
}

var forwardedAddressNames = func() []string {
	names := make([]string, 0, len(forwardedAddressHeaders))
	for _, h := range forwardedAddressHeaders {
		names = append(names, h.name)
	}
	return names
}()

// hopByHopExempt holds the names stripConnectionListed never deletes,
// whatever Connection lists: the session cookie lives in Cookie, and the
// origin check reads Origin and Sec-Fetch-Site.
var hopByHopExempt = map[string]struct{}{
	"Cookie":         {},
	"Origin":         {},
	"Sec-Fetch-Site": {},
}

// identitySpellings is every name in identityHeaders in its dash form
// and, where distinct, its all-underscore form: the set the boundary
// probe sends. Group 1 itself matches by [foldedKey], which also covers
// a name mixing the two separators.
var identitySpellings = withUnderscoreTwins(identityHeaders)

var identityKeys = foldedKeys(identityHeaders)

func withUnderscoreTwins(names []string) []string {
	spellings := make([]string, 0, len(names)*2)
	for _, name := range names {
		spellings = append(spellings, name)
		if underscored := strings.ReplaceAll(name, "-", "_"); underscored != name {
			spellings = append(spellings, underscored)
		}
	}
	return spellings
}

func foldedKeys(names []string) map[string]struct{} {
	keys := make(map[string]struct{}, len(names))
	for _, name := range names {
		keys[foldedKey(name)] = struct{}{}
	}
	return keys
}

// strip is the ingress strip. It removes group 1, every identity
// spelling, from every peer; group 2, the forwarded-address headers,
// from every peer but a trusted proxy; and group 3, every header the
// request's own Connection header lists. It wraps the mux rather than
// any handler, so the removal holds for every path.
type strip struct {
	proxy TrustedProxyConfig

	// forwarded is group 2: forwardedAddressNames plus the configured
	// client-address header, which must join the set or a name outside
	// the fixed family would leave the one header [ClientIP] reads free
	// for any client to set.
	forwarded []string

	// forwardedCanon is forwarded in canonical form, for comparing a
	// Connection token.
	forwardedCanon map[string]struct{}
}

func newStrip(proxy TrustedProxyConfig) strip {
	forwarded := forwardedAddressSet(proxy.header)
	canon := make(map[string]struct{}, len(forwarded))
	for _, name := range forwarded {
		canon[textproto.CanonicalMIMEHeaderKey(name)] = struct{}{}
	}
	return strip{
		proxy:          proxy,
		forwarded:      forwarded,
		forwardedCanon: canon,
	}
}

// forwardedAddressSet returns group 2's names: the fixed family plus
// configured, when that name is not already one of them.
func forwardedAddressSet(configured string) []string {
	if configured == "" {
		return forwardedAddressNames
	}
	for _, name := range forwardedAddressNames {
		if sameHeaderName(name, configured) {
			return forwardedAddressNames
		}
	}
	return append(slices.Clone(forwardedAddressNames), configured)
}

func (s strip) wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.stripIdentity(r)
		s.stripForwardedAddress(r)
		s.stripConnectionListed(r)
		next.ServeHTTP(w, r)
	})
}

// stripIdentity removes group 1 from r.Header and r.Trailer, then wraps
// r.Body to remove it from r.Trailer again once the trailer section
// arrives: net/http's mergeSetHeader re-adds an announced trailer key at
// end of stream, and installs an unannounced one's map wholesale. The
// body is wrapped whether or not a Trailer header was announced, since
// a chunked request may send a trailer it never announced, and arming
// the scrub on the announcement would let omitting one header bypass
// the strip.
func (s strip) stripIdentity(r *http.Request) {
	deleteIdentity(r.Header)
	deleteIdentity(r.Trailer)
	scrubTrailerOnRead(r)
}

// foldedKey returns the form group 1 matches a header key by: k with
// every underscore read as a dash, then canonicalized. A downstream
// that treats "-" and "_" alike (CGI's HTTP_ variables, nginx with
// underscores_in_headers) reads every key sharing one folded form as
// one header, and a name of three or more segments has mixed spellings
// no list enumerates.
func foldedKey(k string) string {
	return textproto.CanonicalMIMEHeaderKey(strings.ReplaceAll(k, "_", "-"))
}

// deleteIdentity deletes from h every key whose [foldedKey] is in
// identityKeys, ranging over the keys h holds so a spelling no list
// names still matches.
func deleteIdentity(h http.Header) {
	for k := range h {
		if _, ok := identityKeys[foldedKey(k)]; ok {
			delete(h, k)
		}
	}
}

func sameHeaderName(a, b string) bool {
	return textproto.CanonicalMIMEHeaderKey(a) == textproto.CanonicalMIMEHeaderKey(b)
}

func scrubTrailerOnRead(r *http.Request) {
	if r.Body == nil {
		return
	}
	r.Body = &trailerScrubbingBody{ReadCloser: r.Body, req: r}
}

// trailerScrubbingBody deletes the identity keys from its request's
// Trailer at end of stream, just after net/http has filled it. It holds
// the request rather than the map: for an unannounced trailer, r.Trailer
// is nil until net/http assigns the wire trailer's map at end of stream.
type trailerScrubbingBody struct {
	io.ReadCloser
	req *http.Request
}

func (b *trailerScrubbingBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	if err == io.EOF {
		b.scrub()
	}
	return n, err
}

// Close closes the wrapped body and scrubs again. net/http's Close
// drains a chunked body inside its own reader, never through Read, so a
// handler that closes the body unread still has r.Trailer filled with
// whatever the client sent.
func (b *trailerScrubbingBody) Close() error {
	err := b.ReadCloser.Close()
	b.scrub()
	return err
}

func (b *trailerScrubbingBody) scrub() {
	deleteIdentity(b.req.Trailer)
}

func (s strip) stripForwardedAddress(r *http.Request) {
	if s.trustsPeer(r) {
		return
	}
	for _, name := range s.forwarded {
		r.Header.Del(name)
	}
}

// trustsPeer reports whether r arrived from a configured proxy. Groups
// 2 and 3 both turn on this one answer, so they cannot disagree about
// which peer is the front.
func (s strip) trustsPeer(r *http.Request) bool {
	peer, ok := peerAddr(r)
	return ok && s.proxy.trusts(peer)
}

// stripConnectionListed removes group 3, every header the request's
// Connection header lists, with three exceptions. A listed Upgrade keeps
// Upgrade and Connection, which a websocket handshake needs. A name in
// hopByHopExempt is kept. From a trusted peer, group 2's names are kept:
// a client behind the front could otherwise list X-Forwarded-For in its
// own Connection header and collapse [ClientIP] to the front's address.
func (s strip) stripConnectionListed(r *http.Request) {
	tokens := connectionTokens(r.Header.Values("Connection"))
	if len(tokens) == 0 {
		return
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
		if _, keep := hopByHopExempt[canon]; keep {
			continue
		}
		if keepForwarded {
			if _, keep := s.forwardedCanon[canon]; keep {
				continue
			}
		}
		r.Header.Del(token)
	}
}

// connectionTokens flattens every Connection value, each possibly
// comma-separated, into the header names it lists.
func connectionTokens(values []string) []string {
	var tokens []string
	for _, v := range values {
		for hop := range strings.SplitSeq(v, ",") {
			if token := strings.Trim(hop, " \t"); token != "" {
				tokens = append(tokens, token)
			}
		}
	}
	return tokens
}
