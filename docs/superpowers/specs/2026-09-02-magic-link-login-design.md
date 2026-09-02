# Magic-link login and streamer SSO: design

Status: SUPERSEDED as design, 2026-09-02. This document is source
material for the greenfield build spec
(2026-09-02-greenfield-server-build.md), whose auth birth
requirements win on every conflict with the design below (nonce
durability, revocation, routing rule 2, the trusted source, the
wire identity, enumeration behavior, and the build order among
them). Read it for intent and topology, never as requirements.

Original status: brainstormed and ratified in session 2026-09-02
(Geoff + Claude). Blocks the family launch: Pass B's uploader is deployed but
the URL goes to no family member until this ships. Supersedes
Cloudflare Access on the family path. Files three amendments against
the organizer spec (2026-09-01-dubplate-organizer-design.md), listed
at the end.

## What this builds

One branded email login for the whole music estate. A family member
enters their email on a login page served by `dubplate serve`, clicks
the magic link that arrives, and holds a session that opens both the
uploader and the streamer's web player. Native phone apps and public
share links keep working unchanged. Cloudflare Access retires from the
family path.

## Universality constraint (Geoff, 2026-09-02)

The server is a universal tool, and this subsystem must carry no
machinery specific to one estate. The tests, applied to every piece:

- No hostname, path, email, or personal name outside `internal/config`.
  Login-page contact copy derives from configured admin identity, the
  same rule the reject emails follow.
- The streamer SSO proxy is a capability of the streamer-driver seam,
  configured by enable flag. Navidrome is the reference driver, never a
  hardcoded assumption; the trusted header's name and the injection
  rules live with driver code.
- Every tunable is a threshold with a sane default: session lifetime,
  link expiry, request rate limits.
- The design must make sense for an adopter with no Cloudflare
  account: plain SMTP plus any streamer, tunnel optional, forward-auth
  proxies optional in front.

## Topology

`dubplate serve` becomes the ingress for the music hostname. The
estate's tunnel (or any adopter's reverse proxy) routes the whole host
to serve. Serve handles the uploader path and the auth endpoints
itself and reverse-proxies everything else to the streamer.

Routing rules, in order:

1. `/rest/*` and `/share/*` pass through to the streamer completely
   untouched: no session check, no identity header. Native Subsonic
   clients authenticate with the streamer directly, and public share
   links stay public. (Navidrome documents these as the two paths that
   must bypass externalized auth.)
2. The uploader path and `/api/*` (dubplate's own API) are served by
   dubplate, session-gated.
3. Auth endpoints (`/login`, the link-consumption endpoint, logout)
   are served by dubplate, unauthenticated by nature, rate-limited.
4. Everything else proxies to the streamer. With a valid session, the
   proxy strips any client-supplied identity header and injects the
   trusted header from the verified session. Without one, serve
   redirects to the login page itself; the streamer's SPA never sees
   the unauthenticated request (Navidrome's SPA ignores redirect
   headers on 401, so the proxy must own this).

The proxy uses Go's `httputil.ReverseProxy`. Server-sent events
(Navidrome's `/api/events`) stream correctly through it without extra
configuration. The streamer is configured to trust only the proxy's
source range (`ND_EXTAUTH_TRUSTEDSOURCES` pinned to the compose
gateway CIDR the estate already pins; an adopter pins their own).

Availability note, recorded deliberately: the proxy puts serve in the
streaming path, so serve down means the web player is down. Native
apps via `/rest/*` are only down if serve is down too, since serve
fronts the host. The capability is optional: without it the streamer
keeps its own login page and loses only SSO.

## Login flow

- The login page is a Go-served template sharing the uploader's
  compiled CSS and design language. The unauthenticated surface
  carries no JavaScript.
- The visitor enters an email. If it matches the roster
  (case-insensitive, the existing `config.Roster` rules), serve signs
  a one-time magic-link token and emails it through the configured
  SMTP. If it does not match, the page says so kindly and points at
  the configured admin contact, the same register as the uploader's
  roster-absent page. That is a deliberate kindness-over-enumeration
  choice, recorded here: the roster is a handful of family addresses,
  and a mistyped email deserves an answer, not silence. The rate
  limits below bound what enumeration could probe.
- The magic link is an HMAC-signed token (`DUBPLATE_MAGIC_LINK_SECRET`)
  carrying the email, an expiry (default 15 minutes), and a nonce.
  Consumption is one-time: serve keeps a small consumed-nonce set with
  TTL, which survives within a process lifetime; a restart forgiving a
  replay inside the 15-minute window is an accepted family-scale
  tradeoff, recorded here.
- Consuming the link sets the session cookie and redirects to the
  originally requested path (validated as same-origin relative, never
  an absolute URL) or the player root.

## Sessions

Stateless, HMAC-signed cookie (`DUBPLATE_SESSION_SECRET`): email,
issue time, expiry. Default lifetime 90 days, rolling (refreshed on
use past the halfway mark). `Secure`, `HttpOnly`, `SameSite=Lax`.
Revocation is secret rotation, which signs out everyone; per-user
revocation is deliberately out of scope at family scale and recorded
as a future organizer concern (the decision-queue rung's session
vocabulary can revisit it).

The uploader's existing identity plumbing consumes the session the
same way it consumed the Access JWT: the verify-identity seam yields a
verified email, and everything downstream (ownership checks, ledger
attribution, roster lookup) is unchanged. `internal/access` retires
from the family path; the seam interface stays so any forward-auth
front (Access included) can layer back in front for adopters who want
edge hardening.

## Streamer SSO mapping

The trusted header value is the contributor's roster `name`, the same
identity that names their inbox directory. The streamer driver owns
the header name and injection (Navidrome: `Remote-User` via ExtAuth).
Navidrome auto-creates a user on first web-app load with a random
password and a regular role; the post-login redirect deliberately
lands on the player root so auto-creation fires. The estate's admin
account predates this change, so the first-user-becomes-admin rule is
already spent; the runbook documents that ordering for fresh adopters.

Phone apps need streamer credentials (reverse-proxy auth cannot cover
them; Navidrome ignores Subsonic credentials only when the header is
present, and the proxy never injects on `/rest/*`). Self-serve: the
streamer's user editing stays enabled and family members set their app
password in the player's own settings; help copy points there.

## Security requirements

- Serve strips any inbound trusted-header spelling on every request
  before its own evaluation (the documented parameter-pollution risk).
- Rate limits, on by default: link requests per email and per source
  address (defaults on the order of 5 per hour), link consumption
  attempts per source. These graduate from the deferred-minors list to
  requirements because unauthenticated traffic now reaches serve.
- Timing-safe comparisons on token verification; single-use nonces as
  above; tokens never logged.
- The login endpoints emit machine-readable errors consistent with the
  API contract; copy follows the copy contract register.
- The `web-auth-security-reviewer` gate runs over the whole subsystem
  before the attended gate, and the pass-end fan-out includes it.

## Config

New required-for-serve values (the serve wall grows by exactly the
subsystem's needs): `DUBPLATE_MAGIC_LINK_SECRET`,
`DUBPLATE_SESSION_SECRET` (both minted via the age-store flow).
New optional thresholds with defaults: link expiry, session lifetime,
rate limits. New streamer-driver config: SSO enable flag, proxy
target, trusted-header name (driver default). Retired from the wall:
`DUBPLATE_ACCESS_TEAM_DOMAIN`, `DUBPLATE_ACCESS_AUD` (the seam keeps
optional support for a forward-auth front, but it is no longer
required config). The estate's Access application on the uploader path
is deleted at cutover.

## Package shape

`internal/auth`: tokens, sessions, the login handlers, rate limiting.
`internal/serve` consumes it through the verify-identity seam and
hosts the streamer proxy as a distinct component. The boundary is
drawn so the organizer initiative's engine extraction inherits it: the
admin-spine rung later builds device-code token acquisition on the
same session vocabulary, per the organizer spec.

## Build order

One compact pass, tasks roughly: (1) `internal/auth` tokens + sessions
+ tests; (2) login page template + endpoints + rate limiting; (3) the
streamer proxy with the bypass rules + header hygiene; (4) uploader
middleware swap (session replaces Access JWT) + config changes; (5)
Navidrome ExtAuth config + compose/tunnel rewiring + deploy.sh; (6)
security review gate + fresh login-flow captures (every page on the
user path rendered and read before the owner sees it). Then ONE
combined attended gate: Geoff verifies login, uploader, and player on
his devices in a single sitting; on his yes, the U12b FileBrowser
cutover, the Access app deletions, and both passes' closes run
unattended.

## Amendments filed against the organizer spec

1. Reference auth flips: dubplate's magic-link login becomes the
   verify-identity seam's reference implementation; Cloudflare Access
   demotes to optional edge hardening.
2. The streamer-driver seam gains an optional web-SSO capability
   (header-asserted identity; Navidrome ExtAuth is the reference,
   Jellyfin's forward-auth is a candidate); capability absent means no
   SSO, nothing else changes.
3. The availability coupling of the SSO proxy is recorded where the
   driver seam is specified.

## Out of scope

Per-user session revocation (future organizer concern). OIDC or
password auth (the seam permits them later; nothing here builds them).
A device-code flow for non-browser clients (admin-spine rung).
Navidrome account lifecycle beyond auto-creation (role changes and
removals stay in the streamer's own admin UI).
