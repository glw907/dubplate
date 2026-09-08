# Rung 5 prior art: sessions, tokens, mail, limits, CSRF, and the login pages

Status: research input for the rung 5 plan. Written 2026-09-08. Every
external claim below carries the URL it came from and the date it was
read. Verdicts marked "provisional, conductor ruling owed" are applied
as recommendations in the plan and stand until a conductor rules.

This document scores prior art under CLAUDE.md's adoption rule: license,
code quality against the house bar, maintenance (release cadence and
open-issue age), and adoption (importers and who depends on it).
Familiarity is not evidence. The 2.0 streamer spec's prior-art section
(`specs/2026-09-07-streamer-2-0-design.md`, lines 393 to 420) already
scored two entries this study touches, and the verdicts below agree with
it where they overlap.

Two facts shape almost every verdict. Rung 5 needs no store by deliberate
design, per the reverse-edge cell (5,6), so a library whose value is
store management has nothing to manage here. And the auth birth
requirements pin the token and cookie construction down to the
domain-separation label and the serialization, so a library that owns
those bytes would be owning the requirement rather than serving it.

## Verdicts at a glance

| Area | Candidate | Verdict | One-line reason |
|---|---|---|---|
| Sessions | stdlib `net/http` alone | **Adopt** | The design is a stateless HMAC `__Host-` cookie; `http.SetCookie` plus `crypto/hmac` is the whole mechanism. |
| Sessions | `gorilla/sessions` | Decline | Its value is store plumbing and cookie serialization rung 5 must own; last release 2024-08-20. |
| Sessions | `alexedwards/scs` | Decline | Excellent library, wrong shape: it is a server-side store manager, and rung 5 has no store. |
| Tokens | stdlib `crypto/hmac`, `crypto/rand`, `crypto/subtle` | **Adopt** | The birth requirements specify the construction; the stdlib supplies exactly it and nothing else. |
| Tokens | `golang-jwt/jwt/v5` | Decline for minting; hold for Access | Algorithm agility is a liability for a token that never leaves this server. Keep it only if the Cloudflare Access forward-auth front is built. |
| Tokens | `o1egl/paseto` | Decline | v1.0.0 published 2018-11-08; its own README deprecates protocol v1. |
| Tokens | `aidanwoods.dev/go-paseto` | Decline | Well-made and maintained, but PASETO's value is interoperability across a trust boundary this token never crosses. |
| Mail | `wneessen/go-mail` | **Adopt** | Message composition is the hard half, the salvage package already uses it and passed the house gate, and it covers implicit TLS and STARTTLS. |
| Mail | stdlib `net/smtp` | Decline as the whole answer | Frozen, no implicit TLS, and it composes no message. |
| Mail | `emersion/go-smtp` | Decline | Its distinguishing value is the server half, which rung 5 never needs. |
| Mail | JMAP from Go | Decline | No credible client: the best is `v0.5.3` with 9 importers, and poplar hand-rolled its own rather than depend on one. |
| Mail | Cloudflare Email Service | **Adopt as documentation, not code** | Its SMTP submission endpoint means the SMTP client already covers it, with no Worker and no new config keys. |
| Mail | A Worker relay of our own | Decline | A second deployable and a shared secret to buy what authenticated SMTP submission already gives. |
| Rate limits | `golang.org/x/time/rate` | **Adopt** | BSD-3, Go team, 14,348 importers. The bounded evicting keyed map around it is ours. |
| CSRF | stdlib `net/http.CrossOriginProtection` | **Adopt** | Added in Go 1.25, the module is on `go 1.27`, and it implements the origin check the birth requirements state. |
| CSRF | `gorilla/csrf` | Decline | Carries a live advisory, GO-2025-3884, for improper validation of trusted origins. |
| Login page | DaisyUI 5 `fieldset`, `input`, `validator` | **Adopt** | The stack's own idiom for a labelled form field, with a hint element that reserves its space. |
| Login page | WCAG 2.2 technique G218 | **Adopt as the accessibility argument** | Email link authentication is a listed sufficient technique for SC 3.3.8. |

## Sessions and cookies

The session the birth requirements describe is an HMAC-signed `__Host-`
cookie carrying an email, an issue time, and an expiry, refreshed under
an absolute lifetime cap and revoked by a per-email valid-after epoch in
the state directory. Nothing in that sentence is a session store. The
cookie is the state, the epoch file is the revocation channel, and the
verification is one HMAC and one `subtle.ConstantTimeCompare`.

### gorilla/sessions

| Criterion | Evidence |
|---|---|
| License | BSD-3-Clause, per pkg.go.dev, read 2026-09-08 (<https://pkg.go.dev/github.com/gorilla/sessions>). LICENSE file at <https://github.com/gorilla/sessions/blob/main/LICENSE>, not fetched. |
| Quality | Not assessed against the house bar; the verdict turns on shape, not on code. |
| Maintenance | Newest release v1.4.0, 2024-08-20 (<https://github.com/gorilla/sessions/releases>, read 2026-09-08). 4 open issues, 9 open pull requests, 3.2k stars (<https://github.com/gorilla/sessions>, read 2026-09-08). |
| Adoption | 8,382 importers (pkg.go.dev, read 2026-09-08). |

Adoption is high and the license is fine. The problem is that the package
exists to hide cookie serialization behind a `Values map[any]any` and to
swap stores underneath it. Rung 5 has to state its serialization
unambiguously, put a domain-separation label inside the MAC input, and
support a `_PREVIOUS` secret for rotation overlap. All three are
requirements about the exact bytes, and a library that owns those bytes
turns each requirement into an argument with the library. Two years
between releases is a secondary concern next to that.

**Verdict: decline.** Provisional, conductor ruling owed.

### alexedwards/scs

| Criterion | Evidence |
|---|---|
| License | MIT, per pkg.go.dev, read 2026-09-08 (<https://pkg.go.dev/github.com/alexedwards/scs/v2>). |
| Quality | Reads as a house-bar library from the outside: small surface, explicit middleware, documented cookie attributes. Not read line by line for this study. |
| Maintenance | v2.9.0, published 2025-04-17 per pkg.go.dev. The README states the project is stable, "actively maintained with ongoing bug fixes and essential updates", and that API changes are not expected (<https://github.com/alexedwards/scs>, read 2026-09-08). 11 open issues, 10 open pull requests, 2.6k stars. |
| Adoption | 3,701 importers (pkg.go.dev, read 2026-09-08). |

This is the better of the two libraries and still the wrong tool. SCS
communicates a session token by cookie and keeps the data server-side
across 19 or more store backends. Rung 5's design puts the data in the
cookie and keeps no server-side session at all, which is what makes cell
(5,6) resolve to "rung 5 needs no store". Adopting SCS would mean
introducing a store to hold something the cookie already carries, ahead
of rung 6, in the one rung the parallel spec says must not depend on it.

**Verdict: decline.** Provisional, conductor ruling owed.

### What to adopt instead

`net/http.SetCookie` with `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`
and the `__Host-` name prefix, plus `crypto/hmac` over an explicit
serialization. The salvage tree already carries the in-repo precedent
for the neighbouring problem: `_salvage/internal/serve/session.go`
generates upload session IDs as 128 bits from `crypto/rand` in
lowercase base32 with no padding, with a compiled pattern that every
ingress validates against. That shape (fixed alphabet, fixed length, one
validator) is worth carrying into the nonce encoding.

## Magic-link token construction

The birth requirements specify the token: an HMAC carrying email, expiry
and nonce, with a length floor on the secret, a domain-separation label
in the MAC input, unambiguous serialization, a `_PREVIOUS` rotation
overlap, verification ordered MAC then expiry then nonce insert, and one
generic failure outcome. That is a specification of a construction, and
the stdlib is the only candidate that implements exactly it.

### golang-jwt/jwt/v5

| Criterion | Evidence |
|---|---|
| License | MIT (pkg.go.dev, read 2026-09-08, <https://pkg.go.dev/github.com/golang-jwt/jwt/v5>). |
| Quality | Good, with an explicit posture on the two classic JWT failures. The README warns that "It's important that you validate the `alg` presented is what you expect", and `alg=none` is accepted only when the caller passes `jwt.UnsafeAllowNoneSignatureType` as the key (pkg.go.dev, read 2026-09-08). |
| Maintenance | v5.3.1, published 2026-01-28. |
| Adoption | 15,146 importers, the highest of any candidate in this study. |

The library is fine. The format is the wrong choice for this token. JWT's
header names its own algorithm, which is why every JWT library ships a
warning about validating `alg`, and the mitigation for a risk that exists
only because the format is negotiable is worse than a format that is not.
Our token is minted and consumed by the same process, with a secret only
that process holds, so there is no interoperability to buy.

One carve-out. `_salvage/internal/access/access.go` verifies Cloudflare
Access assertions with `golang-jwt/jwt/v5`, and the birth requirements
keep a forward-auth front (Access included) as an optional integration
behind an enable flag. If rung 5 builds that integration, the dependency
returns for that purpose alone, verifying someone else's tokens rather
than minting ours.

**Verdict: decline for minting. Hold as the Access-verification
dependency, entering `go.mod` only in the commit that builds that
integration.** Provisional, conductor ruling owed.

### o1egl/paseto

| Criterion | Evidence |
|---|---|
| License | MIT (pkg.go.dev, read 2026-09-08, <https://pkg.go.dev/github.com/o1egl/paseto>). |
| Quality | Not assessed; the maintenance evidence settles it. |
| Maintenance | v1.0.0, published 2018-11-08 (pkg.go.dev, read 2026-09-08). The README states "Version 1 of the protocol is deprecated" and names v2 as the recommended protocol version, with 4 open issues and 941 stars (<https://github.com/o1egl/paseto>, read 2026-09-08). |
| Adoption | 1,295 importers. |

A module whose newest tagged version is from 2018 and whose own README
deprecates part of what it implements fails the maintenance criterion
outright. PASETO reached protocol versions 3 and 4 after that tag.

**Verdict: decline.**

### aidanwoods.dev/go-paseto

| Criterion | Evidence |
|---|---|
| License | MIT (pkg.go.dev, read 2026-09-08, <https://pkg.go.dev/aidanwoods.dev/go-paseto>; repository <https://github.com/aidantwoods/go-paseto>). |
| Quality | Presents well. The README's stated design is versioned protocols rather than algorithm agility, "making it harder to misuse", and it implements PASETO v2, v3 and v4 for both local and public modes. |
| Maintenance | v1.6.0, published 2025-12-27. 2 open issues, 494 stars (GitHub, read 2026-09-08). |
| Adoption | 235 importers. |

This is the credible token library in the field, and it still loses to
the stdlib here. PASETO's advantage over JWT is that the version pins the
algorithm, which matters when a token crosses a trust boundary to a party
who must decide how to verify it. Our magic-link token crosses no such
boundary. The whole construction is a MAC over a serialization we already
have to specify, which is under fifty lines of `crypto/hmac`, and adding
a module to hold those fifty lines adds a supply-chain edge and a
migration story for no property we lack.

**Verdict: decline.** Provisional, conductor ruling owed.

### What to adopt instead

`crypto/hmac` with `crypto/sha256`, `crypto/rand` for the nonce, and
`crypto/subtle.ConstantTimeCompare` for verification, all stdlib
(<https://pkg.go.dev/crypto/hmac>, <https://pkg.go.dev/crypto/subtle>).
The domain-separation label and the field separator become constants the
package documents, which is where the birth requirement wanted them.

## Mail delivery from Go

Sending one short plaintext email is two problems. Getting bytes to a
relay is the easy half. Composing a message that renders correctly, with
a Message-ID, a Date, an RFC 2047 encoded subject when a name carries an
accent, and a quoted-printable body, is the half that produces bug
reports.

### stdlib net/smtp

The package documentation states plainly: "The smtp package is frozen and
is not accepting new features" (<https://pkg.go.dev/net/smtp>, read
2026-09-08). It implements SMTP with 8BITMIME, AUTH with PLAIN and
CRAM-MD5, and STARTTLS. PLAIN refuses to send credentials on a connection
that is neither TLS-protected nor to localhost, which is a good default.
It has no implicit-TLS dial, so port 465 needs `tls.Dial` plus
`smtp.NewClient` by hand, and it composes nothing: the caller writes the
headers.

Frozen is not the same as broken, and for a five-header plaintext message
the composition gap is closable. The reason not to close it by hand is
the next entry.

### wneessen/go-mail

| Criterion | Evidence |
|---|---|
| License | MIT. LICENSE file fetched 2026-09-08: "MIT License", "Copyright (c) 2022-2025 The go-mail Authors" (<https://raw.githubusercontent.com/wneessen/go-mail/main/LICENSE>). |
| Quality | The package documents "a small dependency footprint by mainly relying on the Go Standard Library and the Go extended packages", implicit TLS and configurable STARTTLS policies, a TLS 1.2 minimum default, and a wide AUTH set including SCRAM and XOAUTH2 (pkg.go.dev, read 2026-09-08). The in-repo evidence is stronger than any of that: `_salvage/internal/notify/notify.go` uses it, and that package passed the repo's own gate and comment standard. |
| Maintenance | v0.8.1, published 2026-07-09 per pkg.go.dev, read 2026-09-08. Still a v0 module. |
| Adoption | 237 importers (pkg.go.dev, read 2026-09-08). |

Adoption is the weakest cell, and it is worth stating honestly: 237
importers is two orders of magnitude below `x/time/rate`. Set against
that, the repo has already run this library in production on the box's
reject-mail path, the module has one clear maintainer with an active
release history, and the alternative is hand-rolling MIME composition
against a frozen stdlib package.

The salvage package also carries an operational fact worth keeping. Its
comment on `implicitTLSPort` records that the deployed route is port 587
with STARTTLS, "since Hetzner blocks outbound 465 by default", dated
2026-08-31 (`_salvage/internal/notify/notify.go`). Any transport
recommendation that assumes 465 works from the box is wrong until
re-checked.

**Verdict: adopt, with `_salvage/internal/notify` as the in-repo starting
point under fresh review.** Reverse-edge cells (5,7) and (5,9) both bind
this: `internal/notify` is rung 5's, and it must be built as a general
send seam rather than as a magic-link mailer, because rung 7's reject
emails and rung 9's notifications are its next two consumers.

### emersion/go-smtp

| Criterion | Evidence |
|---|---|
| License | MIT (pkg.go.dev, read 2026-09-08, <https://pkg.go.dev/github.com/emersion/go-smtp>). |
| Quality | Well regarded in the Go mail ecosystem. Not assessed line by line. |
| Maintenance | v0.25.0, published 2026-08-13. Newer than go-mail's newest tag. |
| Adoption | 404 importers. |

The package's own positioning is the reason to decline it: the
documentation notes that the stdlib's client is frozen and that go-smtp
offers "a server implementation and a number of client improvements". The
server implementation is the differentiator, and rung 5 receives no mail.
Its client half would still leave message composition to us. Adopting it
means taking the larger surface to solve the smaller half of the problem.

**Verdict: decline.**

### JMAP from Go

The workstation's own evidence is the clearest signal here. poplar talks
JMAP to Fastmail and depends on no third-party JMAP module: its `go.mod`
lists none, and `internal/backend/jmapsource/mail.go` imports
`github.com/glw907/poplar/jmap`, a package poplar wrote itself. Its send
path builds an `EmailSubmission` through that package
(`~/Projects/poplar/internal/backend/jmapsource/mail.go`, read
2026-09-08).

The best available third-party client is `git.sr.ht/~rockorager/go-jmap`:
MIT, v0.5.3 published 2025-02-01, 9 importers (pkg.go.dev, read
2026-09-08). Nine importers and a v0 tag is not a dependency to put on
the family's login path. Writing our own, as poplar did, is a
disproportionate answer to "send one short email".

Fastmail itself does not require JMAP. It publishes SMTP submission on
`smtp.fastmail.com`, port 465 with "SSL/TLS Encryption: Enabled, but not
STARTTLS" or port 587 for STARTTLS clients, authenticating PLAIN with an
app-specific password, and states "You cannot use your regular Fastmail
password"
(<https://www.fastmail.help/hc/en-us/articles/1500000278342-Server-names-and-ports>,
read 2026-09-08).

**Verdict: decline JMAP as a dubplate transport. Fastmail remains fully
available over SMTP.** Provisional, conductor ruling owed.

### Cloudflare Email Service

The estate already sends mail from a Worker. The pings Worker uses the
`send_email` binding with `FROM_ADDRESS = "pings@907.life"`
(`~/Projects/pings/src/email.ts` and `wrangler.jsonc`, read 2026-09-08),
so a Worker relay for dubplate is a shape the estate has proven.

Cloudflare's own product has moved past the point where that shape is
needed. Email Sending entered public beta on 2026-04-16, sending from
Workers or the REST API
(<https://developers.cloudflare.com/changelog/post/2026-04-16-email-sending-public-beta/>,
read 2026-09-08). On 2026-06-08 Cloudflare added authenticated SMTP
submission on `smtp.mx.cloudflare.net` port 465 with implicit TLS, AUTH
PLAIN or LOGIN, username `api_token`, and a Cloudflare API token with
"Email Sending: Edit" as the password
(<https://developers.cloudflare.com/changelog/post/2026-06-08-smtp-submission/>,
read 2026-09-08). Submissions through it enter the same pipeline as the
REST API, with automatic DKIM and ARC signing.

That endpoint fits dubplate's existing five reserved SMTP keys exactly:
`DUBPLATE_SMTP_HOST`, `_PORT`, `_USER`, `_PASSWORD`, `_FROM`
(`internal/config/keys.go`, the "auth and identity" reserved class). An
adopter on Cloudflare configures a host, a port, the literal username
`api_token`, and a token. No Worker, no binding, no code.

**Verdict: adopt as a documented adopter option reached through the SMTP
client. Decline the Worker relay and the REST API transport as code.**
Provisional, conductor ruling owed.

## Mail transport options for an adopter

The point of this table is that all four rows below reach the same Go
code. Rung 5 ships one SMTP client, and the choice of provider becomes
configuration and a runbook page.

| Option | Setup steps | Secrets it needs | Deliverability | Cost | Notes |
|---|---|---|---|---|---|
| Any SMTP relay with an app password (Fastmail is the estate's) | Create an app password in the provider, point host and port at the provider, set the From address. Fastmail: `smtp.fastmail.com`, 465 implicit TLS or 587 STARTTLS, AUTH PLAIN. | The five reserved SMTP keys; the password minted through the age store. | Rides the provider's existing SPF and DKIM for the sending domain, so nothing new to publish when the domain already sends through that provider. | Included with a mailbox the adopter already pays for. | The one option that needs no new account of any kind. Fastmail forbids the account password, app password only. |
| Cloudflare Email Service, SMTP submission | Onboard a sending domain in Email Service, mint an API token with Email Sending: Edit, set host `smtp.mx.cloudflare.net`, port 465, user `api_token`. | The same five SMTP keys; the password is the API token. | Automatic DKIM and ARC signing; Cloudflare manages sending IP reputation. Before a sending domain is onboarded, sends are limited to verified destination addresses in the account. | Workers Paid plan. 3,000 outbound emails included per month, then $0.35 per 1,000. Sends to verified destination addresses in the account are free on all plans. | Beta as of 2026-09-08. Daily limits start conservative and scale with account standing, with no published starting number. |
| Cloudflare Email Service, REST API or Workers binding | As above, plus code that speaks the API or a Worker to deploy. | An API token, or a Worker with a binding. | Identical to the SMTP row; same pipeline. | Same. | Declined for dubplate: it buys nothing over the SMTP row and costs a second code path. |
| A relay Worker of our own (the pings shape) | Deploy a Worker, give it a send binding, give dubplate a shared secret and a URL. | A shared secret, plus its rotation story, plus a new config key class. | Whatever the Worker's own sender is. | Workers pricing. | Declined. It was the right shape before Cloudflare published SMTP submission. It now adds a deployable and a secret for no gain. |

**Recommendation for the estate:** Fastmail SMTP with an app password.
The domain already sends through Fastmail, the box's outbound path
already uses port 587 with STARTTLS because of the Hetzner block on 465,
and the salvage `notify` package already speaks exactly this. It adds no
beta dependency to the family's login path.

**Default for other adopters:** none. The spec ships no default host, and
the setup guide presents SMTP submission as the one supported transport
with two worked examples, a mailbox provider's app password and
Cloudflare Email Service. That satisfies the universality constraint's
test that the design "must make sense for an adopter with no Cloudflare
account".

Unverified: whether `907.life`'s current SPF and DMARC records authorize
both the Fastmail and the Cloudflare senders. No DNS lookup was run for
this study. The rung 5 plan should verify before the cutover sitting,
because pings already sends as `pings@907.life` through Cloudflare while
mail for the domain is hosted at Fastmail.

## Rate limiting and login abuse

The requirement is per-email and per-source-address limits on link
requests and consumption attempts, keyed on the real client IP, with
"limiter structures bounded and evicting".

### golang.org/x/time/rate

| Criterion | Evidence |
|---|---|
| License | BSD-3-Clause (pkg.go.dev, read 2026-09-08, <https://pkg.go.dev/golang.org/x/time/rate>). |
| Quality | Go team code, and the exemplar the house bar points at. The package is a token-bucket `Limiter` with `Allow`, `Reserve` and `Wait`. |
| Maintenance | v0.16.0, published 2026-08-19. |
| Adoption | 14,348 importers. |

The 2.0 spec already scored this entry "Adopt" for exactly this purpose,
noting that "The repo's MusicBrainz limiter solves a different problem"
(`specs/2026-09-07-streamer-2-0-design.md` line 415). This study agrees,
and adds the part the library does not cover: `rate.Limiter` is one
bucket. The keyed map from email or IP to limiter, its size bound, and
its eviction are rung 5's to build and to test, and an unbounded map of
limiters keyed by attacker-supplied email addresses is precisely the
memory-exhaustion bug the birth requirement names.

**Verdict: adopt `x/time/rate` for the bucket. Build the bounded evicting
keyed map in `internal/auth`, with a test that proves eviction under a
flood.**

Hand-rolling the bucket is the alternative and it loses on evidence:
14,348 importers against a from-scratch implementation of a well-known
algorithm, in a rung that already carries the build's densest set of
security requirements.

## CSRF on the login form

The birth requirements state one blanket rule: every cookie-carrying,
state-changing surface takes the origin check. Four surfaces qualify: the
login POST, the consumption POST, logout, and the proxy path.

### stdlib net/http.CrossOriginProtection

Added in Go 1.25.0, and the module is on `go 1.27.0`, so it is available
without a dependency. It rejects non-safe cross-origin browser requests
using the `Sec-Fetch-Site` header, "available in all browsers since
2023", falling back to comparing the `Origin` hostname against `Host`.
GET, HEAD and OPTIONS are always allowed. It exposes
`NewCrossOriginProtection`, `AddTrustedOrigin`, `AddInsecureBypassPattern`,
`Check`, `Handler` and `SetDenyHandler`
(<https://pkg.go.dev/net/http#CrossOriginProtection>, read 2026-09-08).

One documented limitation matters for rung 5. "Requests without
Sec-Fetch-Site or Origin headers are currently assumed to be either
same-origin or non-browser requests, and are allowed." All four protected
surfaces are browser-only by construction, so the safe reading of the
birth requirement is stricter than the stdlib default: a POST to any of
them that carries neither header should be rejected rather than allowed.

**Verdict: adopt `CrossOriginProtection` as the mechanism, wrapped so
that a missing `Origin` and `Sec-Fetch-Site` pair is a denial on the four
protected POST surfaces.** Provisional, conductor ruling owed, because it
is a deliberate deviation from a stdlib default and the
`web-auth-security-reviewer` will ask about it either way.

### gorilla/csrf

| Criterion | Evidence |
|---|---|
| License | BSD-3-Clause (pkg.go.dev, read 2026-09-08, <https://pkg.go.dev/github.com/gorilla/csrf>). |
| Quality | Carries a published advisory: "GO-2025-3884: Improper validation of TrustedOrigins allows CSRF attacks in github.com/gorilla/csrf" (pkg.go.dev, read 2026-09-08). |
| Maintenance | v1.7.3, published 2025-01-23, and pkg.go.dev flags "This package is not in the latest version of its module". |
| Adoption | 1,103 importers. |

A CSRF library with a CSRF advisory against its trusted-origin handling
is the exact dependency this rung should not take, when the stdlib
implements the same check.

**Verdict: decline.**

### Double-submit tokens

Declined as redundant. The consumption POST already carries the token in
a form field by design, and adding a second synchronizer field to the
same form buys nothing that the origin check does not already give on a
browser-only surface. `__Host-` plus `SameSite=Lax` plus the origin check
is the stated design, and each of the three has a distinct job.

## The login pages: DaisyUI 5 and accessibility

Ruling 0c R1.6 governs: rung 5's login and consumption pages are
SvelteKit static, client-rendered, built to assets the Go binary embeds
and serves. This supersedes the build spec's "Go-templated, no
JavaScript" wording at line 619, and rung 5's close records the
correction along with the CLAUDE.md and `scripts/check.sh` header fix for
the web gate leg.

Name one consequence for the security review while it is cheap. The
magic-link design's property that "the unauthenticated surface carries no
JavaScript" is spent under R1.6. The `web-auth-security-reviewer` should
be pointed at that change explicitly rather than left to discover it, and
the served CSP that R5.6 specifies, with `style-src 'self' 'unsafe-inline'`,
is the owner note already queued for the rung 5 sitting.

### DaisyUI 5 form components

The stack's own idiom for a labelled field is the `fieldset` component:

```html
<fieldset class="fieldset">
  <legend class="fieldset-legend">Title</legend>
  <label class="label">Description</label>
  <input type="text" class="input" />
</fieldset>
```

The documentation adds "Add proper id and for attributes for
accessibility" (<https://daisyui.com/components/fieldset/>, read
2026-09-08).

For inline validation, the `validator` and `validator-hint` pair drives
off the browser's constraint API. The hint "will be invisible if it comes
after a validator class, then if the input is invalid, validator-hint
will get visible", and it "still occupies space even if it's invisible
because we don't want a layout shift when the hint appears"
(<https://daisyui.com/components/validator/>, read 2026-09-08). The
documented email example is one line:

```html
<input class="input validator" type="email" required placeholder="[email protected]" />
<div class="validator-hint">Enter valid email address</div>
```

**Verdict: adopt both.** One caveat for the plan: `validator` is a
client-side convenience only. The server's own generic outcome, the
roster-absent copy, and the rate-limit copy remain the authority, and
none of them may be inferred from client-side validity.

### Accessibility guidance for a magic-link flow

Magic-link login is not merely accessible, it is a named answer to a WCAG
2.2 criterion. SC 3.3.8 Accessible Authentication (Minimum) requires that
authentication avoid a cognitive function test unless an alternative
exists, and its sufficient techniques list includes "G218: Email link
authentication"
(<https://www.w3.org/WAI/WCAG22/Understanding/accessible-authentication-minimum.html>,
read 2026-09-08). The same document notes that compliance fails if the
user must manually transcribe a verification code, which is a direct
argument for the link-plus-interstitial design over a typed code.

Three concrete items for the pages:

- `autocomplete="email"` on the email field. MDN defines the value as "An
  email address" and warns against `autocomplete="off"` on login fields,
  because browsers and password managers ignore it
  (<https://developer.mozilla.org/en-US/docs/Web/HTML/Reference/Attributes/autocomplete>,
  read 2026-09-08).
- A live region for the outcome. "Check your email" and the generic
  failure message are status messages under WCAG 4.1.3 and must reach
  assistive technology without moving focus, which is `role="status"` or
  `aria-live="polite"` for the confirmation and `role="alert"` for an
  error (WCAG 2.2 SC 4.1.3, <https://www.w3.org/WAI/WCAG22/Understanding/status-messages.html>,
  not fetched directly for this study; the criterion's requirement is
  summarized from secondary sources read 2026-09-08 and should be
  confirmed against the W3C page in the plan).
- `autocomplete="one-time-code"` is defined for OTP fields but does not
  apply here, since the interstitial's token arrives in the link and is
  carried in a form field rather than typed. It becomes relevant only if
  a typed-code fallback is ever added, which is a product fork below.

## How other self-hosted projects do passwordless login

Read for lessons only. Nothing here is transliterated, and the GPL and
AGPL entries are read at arm's length.

**Grafana (Go, AGPL-3.0 core) shipped magic links and is removing them.**
Passwordless authentication by magic link or emailed one-time password
arrived behind the `passwordlessMagicLinkAuthentication` feature toggle
in November 2024, with the OTP valid for 20 minutes by default
(<https://grafana.com/docs/grafana/latest/setup-grafana/configure-access/configure-authentication/passwordless/>,
via search results read 2026-09-08). It is deprecated in v12.3 and
removed in v12.5. The stated reasons, from the deprecation issue dated
2026-02-04, are that the feature "has remained experimental for over a
year with minimal usage" and that "Enterprise platforms are moving toward
phishing-resistant authentication (FIDO2/passkeys), not magic links"
(<https://github.com/grafana/grafana/issues/117400>, read 2026-09-08).

The lesson is about audience, not about the mechanism. Grafana's users
are enterprises with an identity provider, and a magic link competes with
SSO there. dubplate's users are a family with no identity provider, where
the magic link competes with a password each member would otherwise
reuse. The evidence is still worth recording, because it is the strongest
public argument against the design and the plan should know it exists.

**PocketBase (Go, MIT) treats enumeration as the design problem.** Its
OTP flow returns an `otpId` "even if a user with the provided email
doesn't exist", and its documentation warns that OTP standing alone "could
be less secure compared to the other methods because the generated
password is usually 0-9 digits" with guessing and enumeration risk
(<https://pocketbase.io/docs/authentication/>, read 2026-09-08). Two
lessons land. First, dubplate's deliberate opposite choice, the kind
roster-absent message, is a real fork from prevailing practice and is
correctly already a flag with kindness as the default. Second, a
high-entropy link avoids the digit-guessing problem an emailed numeric
code creates, which is another argument for link over code.

**Ory Kratos (Go, Apache-2.0) uses a two-step code, not a link.** "The
first step the user is required to enter their email address or phone
number. In the second step they are prompted to enter the one-time code
sent to their email address or phone number", and the documentation
stresses that "You need a good email delivery provider to ensure that
one-time codes are delivered reliably"
(<https://www.ory.com/docs/kratos/passwordless/one-time-code>, read
2026-09-08). The delivery-provider point is the one to carry: the
transport table above is not incidental to this feature, it is the
feature's reliability.

**Miniflux (Go, Apache-2.0) has no magic link at all.** Its configuration
offers local password with `DISABLE_LOCAL_AUTH`, OAuth2 and OIDC through
`OAUTH2_PROVIDER`, WebAuthn passkeys through `WEBAUTHN`, and proxy
authentication through `AUTH_PROXY_HEADER` with
`TRUSTED_REVERSE_PROXY_NETWORKS`
(<https://miniflux.app/docs/configuration.html>, read 2026-09-08). The
pairing of a header name with an explicit trusted-network setting is the
same shape rung 5's trusted-proxy mode takes, and it is evidence that the
explicit-configuration approach is the field's normal answer rather than
an invention.

**Gitea (Go, MIT) keeps sessions server-side behind a cookie.** It uses
`gitea.com/go-chi/session`, Apache-2.0, a fork of Macaron's session
package supporting memory, file, Redis, and several databases, with
configurable cookie name, path, secure flag, domain and SameSite
(<https://pkg.go.dev/gitea.com/go-chi/session>, via search results read
2026-09-08). It is the opposite architecture to rung 5's, and it is worth
recording as the road not taken: a server-side session store is what you
build when sessions must be listable and individually revocable, and
rung 5 buys per-person revocation with a per-email epoch instead, at a
fraction of the machinery.

**Authelia (Go, Apache-2.0) states the forward-auth header rule that rung
5's second trust boundary needs.** Its proxy documentation requires that a
reverse proxy remove and replace `X-Forwarded-For` when it arrives from an
untrusted client, warning that "If this is not removed from non-trusted
proxies a user could theoretically hijack any rule that contains this
criteria to potentially skip an authentication criteria depending on how
it is configured", and it notes that with a cloud proxy such as
Cloudflare "you must ensure they do this or you configure a rule to do
it" (<https://www.authelia.com/integration/proxies/forwarded-headers/>,
read 2026-09-08). This is external corroboration of the birth
requirement that the forward-auth topology gets its own boundary and that
the driver publishes its strip set to the adopter, and it is the model
for how that documentation should read.

One further data point on the field's licensing. tinyauth, a well-starred
Go forward-auth server (8.2k stars, v5.1.0 dated 2026-06-25), is
AGPL-3.0 (<https://github.com/steveiliop56/tinyauth>, read 2026-09-08),
so it is unreadable-for-adoption under the same rule that governs
Navidrome and gonic. Whether it implements any email code login was not
established; the README fetch did not mention one, and absence of mention
is not absence of the feature.

## What this means for the 2.0 forward hooks

Per CLAUDE.md's "Build for 2.0" rule, each entry above is weighed against
the 2.0 spec's forward-hooks list.

Hook 5, a stable identity, is the only hook this study touches. It says
rung 6 mints a durable surrogate user id and rung 5 resolves a roster
identity to it, and that `streamer_username` "becomes required, unique,
and stable rather than optional". The prior-art consequence is a
constraint on the roster loader rather than on any library: the loader's
birth-requirement validation should treat `streamer_username` as the
renameable handle a surrogate id will later protect, validating it as
ASCII-printable, unique and bounded now, so rung 6 has something stable
to key on. No library in this study helps or hinders that.

Every other entry is neutral to the hooks list. Sessions, tokens, CSRF and
rate limiting are all stdlib or Go-team code with no schema surface. The
mail seam is neutral by construction as long as it is built as a general
send seam, per cell (5,9), because 2.0's notification needs arrive
through the same interface.

## Product forks for Geoff

Each carries a recommendation. None is decided here.

1. **Which mail provider the estate uses.** Recommend Fastmail SMTP with
   an app password: the domain already sends through Fastmail, the box's
   route is already 587 with STARTTLS, and it keeps a beta product off the
   family's login path. The alternative, Cloudflare Email Service SMTP, is
   the same five config keys and no code change, so this is reversible in
   a deploy rather than a rewrite.
2. **What an adopter is told to use by default.** Recommend no default
   host, with two worked setup examples, a mailbox provider's app password
   and Cloudflare Email Service. A shipped default would be a
   site-specific value in the wrong place.
3. **The From address a family member sees.** Recommend
   `DUBPLATE_SMTP_FROM` as required configuration with no default, and the
   login page's contact copy derived from the configured admin identity,
   the rule the reject emails already follow.
4. **Whether the login page offers a typed code as a fallback to the
   link.** Recommend no at rung 5. The GET interstitial plus POST
   consumption already defends the failure mode a code would address, a
   mail gateway that follows links, and the estate already relies on that
   defense elsewhere: the pings Worker makes its failure endpoint POST-only
   "so a link-scanning mail gateway can never trigger it"
   (`~/Projects/pings/README.md`, read 2026-09-08). Adding a typed code
   also reintroduces the transcription burden WCAG 3.3.8 warns about.
5. **Session lifetime and link expiry defaults.** Recommend carrying the
   magic-link design's 15-minute link expiry and revisiting the 90-day
   session default against the new absolute lifetime cap, since the cap
   changes what the default means. Both are thresholds with sane defaults
   under the universality constraint, so this is a copy-and-numbers fork
   rather than an architecture one.

## Provisional rulings the plan should carry

Each is applied as its recommendation and marked provisional, conductor
ruling owed.

- **P1.** Rung 5 adds exactly two new module dependencies:
  `github.com/wneessen/go-mail` and `golang.org/x/time/rate`. Sessions,
  tokens, and CSRF are stdlib.
- **P2.** `golang-jwt/jwt/v5` enters `go.mod` only in the commit that
  builds the Cloudflare Access forward-auth integration, and never for
  minting dubplate's own tokens.
- **P3.** Mail transport at rung 5 is authenticated SMTP submission only.
  JMAP, the Cloudflare REST API, the Workers binding, and a relay Worker
  are declined rather than deferred.
- **P4.** `internal/notify` is adopted from `_salvage/internal/notify`
  under fresh review at the exemplary-Go bar, restructured as a general
  send seam rather than a magic-link mailer, per cells (5,7) and (5,9).
- **P5.** The origin check denies a protected POST that carries neither
  `Origin` nor `Sec-Fetch-Site`, which is stricter than
  `CrossOriginProtection`'s documented default.
- **P6.** The rate-limiter map is bounded and evicting, and a gate-resident
  test proves eviction under a flood of distinct keys.

## Evidence notes and unverified items

License strings in this document come from pkg.go.dev's license detection
and from GitHub repository pages, both read 2026-09-08. Only one LICENSE
file was fetched directly, `wneessen/go-mail`, because it is the one new
module dependency this study recommends adopting. The 2.0 spec's rule
that every row cite the LICENSE file by URL is therefore met for the
adoption and unmet for the declines, which is the right place to spend
the fetches.

Marked unverified:

- `alexedwards/scs` release date. pkg.go.dev gives v2.9.0 as published
  2025-04-17; the GitHub releases page rendered a day and month without a
  year that the fetch read as July 8. Both agree v2.9.0 is newest.
- `wneessen/go-mail` release date. pkg.go.dev gives v0.8.1 as published
  2026-07-09; the GitHub releases page rendering was read as 2024. GitHub
  omits the year for current-year releases, which favours the pkg.go.dev
  date, but this was not confirmed against the tag itself.
- Most recent commit dates for `gorilla/sessions` and `alexedwards/scs`.
  The repository pages did not surface them in the fetch, so cadence is
  judged on releases alone.
- Open-issue age for every candidate. Counts were captured; the age of the
  oldest open issue was not, so the maintenance criterion here rests on
  release cadence and issue count.
- Whether `907.life` DNS currently authorizes both the Fastmail and the
  Cloudflare senders. No lookup was run.
- Cloudflare Email Sending's starting daily limit. The limits page states
  that new accounts "start with a conservative daily quota" that scales
  with account standing, without a number
  (<https://developers.cloudflare.com/email-service/platform/limits/>, read
  2026-09-08).
- Whether Hetzner still blocks outbound port 465 from the box. The claim
  comes from a salvage-tree comment dated 2026-08-31, not from a live
  test.
- Ory Kratos's default code lifetime and attempt limits. The documentation
  page read did not state them.
- Whether tinyauth implements any email code login. Not established.
- WCAG SC 4.1.3's exact wording. Summarized from secondary sources; the
  W3C understanding page was not fetched directly.
