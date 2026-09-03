# dubplate-server: greenfield build spec

Status: revision 2.1, ratified 2026-09-02. Revision 1 was drafted
from the ratified rulings
(`docs/superpowers/research/2026-09-02-greenfield-server-rulings.md`),
read by Geoff, and then put through two adversarial reviews
(`research/2026-09-02-greenfield-spec-reviews/architecture.md` and
`universality.md`). Revision 2 folded both reviews, the session fold
notes, and the beets research program; revision 2.1 records Geoff's
resolutions of the review round's three open rulings (the "Rulings on
the review round" section). The spec is settled; rung plans build
from it.

Sources folded: the organizer spec
(`2026-09-01-dubplate-organizer-design.md`, amended through
2026-09-02), the login spec (`2026-09-02-magic-link-login-design.md`)
as source material, both login-spec adversarial reviews
(`research/2026-09-02-login-spec-reviews/`) as auth birth
requirements, the Navidrome integration research, the conformance
corpus report (`2026-09-01-beets-conformance-corpus.md`), the
uncertain-match UX report (`2026-09-01-uncertain-match-ux.md`), the
matcher mechanics study (`2026-09-02-beets-matcher-mechanics.md`),
the evolution lessons report
(`2026-09-02-beets-evolution-lessons.md`, lessons U1 through U7), and
the ratified reject-copy string authority
(`research/passb-copy-final.md`).

## What this initiative builds

dubplate-server is built fresh, foundation first. It is the universal
server of the organizer spec: intake and admission, the organizing
engine with a native Go MusicBrainz matcher, the library store, the
review and decision queues, backup and restore, notifications, the
admin API, the documented client API, and login. It is not an
evolution of the current code. The existing Go and web trees are
salvage: adoption material, admitted piecewise on merit, retired
after launch.

The organizer spec remains the governing architecture. Its success
bar, purview, seams, custody policy, extensibility fork, settled
forks, harness design, and beauty section apply unchanged. This spec
re-derives what greenfield changes: the ladder, the adoption story,
and the launch shape. Section "What this spec supersedes" lists the
exact deltas, and on any conflict between this spec's requirements
and the login spec's design, this spec wins.

The family system is not live. The box runs Navidrome, cloudflared,
and the four guardian timers protecting the seeded library, and it
stays that way until the new server launches complete.

## The launch bar

The server launches only when it is the complete organizer-spec
server. The launch gate has four parts, and all four hold at once:

1. **Completeness.** Intake, engine, store, review and decision
   queues, backup and restore, notifications, admin API, client API,
   and login all exist and pass their rung gates. No component ships
   as a stub to be finished after launch.
2. **The matcher bar.** The organizer spec's two-tier success bar
   holds over the conformance corpus: zero false confident matches
   (tier one, no adjudication available), and zero unadjudicated
   cases where beets is right and Go is wrong (tier two). Both tiers
   count cases, never percentages. Two refinements, ruled 2026-09-02:
   tier one is measured against release truth rather than against the
   oracle where the oracle is known-defective (ruling 2), and tier
   one additionally holds over the stock-config stratum, not only the
   estate config (ruling 3).
3. **A clean shadow record.** The shadow run defined below is clean:
   zero unadjudicated divergences and zero false confident matches
   against live beets decisions, over the floor ruling 1 defines.
4. **Layout parity.** A dry-run diff proves the new pipeline's
   computed paths equal the on-disk layout of the live library before
   anything writes. This part has its own evidence machinery (rung
   10) and is first evaluated there, ahead of the cutover, not inside
   it.

There is no partial launch and no beets fallback inside the new
server. If the matcher cannot clear the bar, the honest outcome is
that the launch waits: the interim estate keeps serving Geoff's own
listening, and the family stays offline. That is why the evidence
track runs first: the matcher is the initiative's existential risk,
and every spine rung built after a failed matcher would be inventory,
not product.

The relationship with beets outlives the launch. The initiative
succeeds or fails on adopting beets' battle-tested matching behavior,
and the harness is built so dubplate can keep evolving with it: the
version-bump differential job survives as a standing tool, so when a
newer beets improves its matcher, the divergence surfaces as a
harness finding and is adopted deliberately, adjudicated under the
success bar rather than followed automatically. The evolution study
adds the caveat the bar must carry: pinned beets 2.13.1 contains
known matching defects, so the oracle is authoritative about behavior
and not about truth, and the corpus machinery below separates the
two. Everything else the server does is comparatively well-understood
engineering; the evidence track is where the initiative's uncertainty
lives, and the ladder prices it that way.

Improving on beets is in scope, on one discipline (Geoff,
2026-09-02): improvements never ride the port. The port proves the
copy; an improvement is a deliberate divergence, adjudicated and
recorded with written reasoning in the same artifacts the corpus
uses. Known oracle defects are the first improvement class and enter
pre-launch through the known-oracle-bug ledger (ruling 2);
everything else queues as post-launch candidates, fed by the corpus,
the evolution study's recurring-pain classes, and the open-issues
survey. The component model and the corpus are what make this safe:
each candidate lands as a component change whose exact blast radius
the corpus shows before it touches the library.

## Universality

No machinery specific to one estate, anywhere. The login spec's four
tests apply to every rung:

- No hostname, path, email, or personal name outside
  `internal/config`. Site-facing copy derives from configured
  identity (`SiteName`, admin contact), the rule the reject emails
  follow.
- Drivers own driver knowledge. The streamer bypass paths, trusted
  header name, injection rules, logout wiring, post-login
  destination, and provisioning choreography live with the Navidrome
  driver, never in core routing.
- Every tunable is a threshold with a sane default: session lifetime,
  link expiry, rate limits, match thresholds, enumeration behavior.
- The design makes sense with no Cloudflare account: plain SMTP plus
  any streamer, tunnel optional, forward-auth proxies optional in
  front, real-client-IP handling configured rather than assumed.

### The config wall, derived

The organizer's rule is: the required-no-defaults wall covers what
the server owns; integrations get explicit enable flags, and a flag
makes its correlated config required-when-enabled. This spec applies
the rule rather than restating it, because the estate's current
required set is not the wall: a server that will not start without an
R2 bucket fails the fourth universality test at the first required
value.

- **Owned, required, no defaults:** the library root, the state
  directory, `SiteName`, the admin identity, and—because login is
  server purview (ruling 7)—an active identity source. Magic-link
  is the shipped default source; whichever source is active, its
  correlated config is required: for magic-link that is the two
  secrets, `DUBPLATE_PUBLIC_URL`, and an SMTP transport; for a
  forward-auth front it is the front's verification config. Exactly
  one identity source is active per deployment, so login has one wall
  shape, owned-with-a-selectable-source, not two.
- **Integrations, flag-enabled, required-when-enabled:** the streamer
  driver (the Navidrome variables move behind its flag, superseding
  the organizer's "stay required with no default"), the SSO proxy
  capability (requires the streamer flag), backup targets (the rclone
  remote pair; the server starts without one and `doctor` says so
  plainly), monitoring (pings), the trusted-proxy mode (header name
  and trusted peer as one required-when-enabled pair—proxyless
  operation is chosen, never fallen into by a silent fallback), and a
  forward-auth front.
- **Compat discipline (evolution lesson U5):** unknown or superseded
  config keys are load errors that name their replacement, never a
  silently degraded start. beets 2.4.0 silently dropped MusicBrainz
  for users with customized plugin lists; that failure class is
  refused at the wall.

Rung 1 derives the complete table from this classification and it
becomes the config core's contract. The estate's
`env/dubplate.env.template` is an instance of the wall, never its
definition.

## The Go bar

The bar is exemplary, not conforming. Stdlib idiom is the default
vocabulary, exported surfaces read like stdlib APIs, and the ops seam
is the in-repo shape to imitate. Reviewers are empowered to fail
awkward-but-working code. The organizer spec's split between gated
properties and judgment properties applies as written, including its
bounded adjacency rule.

One priority ordering inside the engine is explicit (Geoff,
2026-09-02): matching rules are structured to be as easy as possible
to alter, even at the expense of raw speed. Speed and efficiency
matter, and they are secondary to code that is easy to evolve. The
structural consequence for rung 3: the distance model is a set of
small, named penalty components with declared weights, each carrying
its own unit tests, composed by a scoring core that treats them
uniformly, so changing a rule touches one component and its cases
rather than a fused loop. This ordering is what keeps the
evolve-with-beets commitment cheap: an adjudicated upstream
improvement lands as a component change, not a rewrite. The scale
supports the trade, since matching cost is dominated by rate-limited
MusicBrainz lookups, not CPU.

Server before TUI is a hard boundary. No TUI and no local-binary work
until the server is fully formed. The ops seam is still designed for
two implementations from day one, because retrofitting the job model
is what the organizer spec forbids; only the server-side
implementation is built in this initiative, and the parity story is
stated honestly at rung 6.

## Two tracks, dependency-merged

Approach A, ratified. One executor works the rungs sequentially; the
two tracks are dependency lanes, not parallel executors.

- **Evidence track**: corpus and harness, then the Go matcher, then
  shadow. Independent of the spine rungs 5 through 13 (all three
  depend on rung 1's config core and binary skeleton), and scheduled
  first, for two reasons. The matcher is the risk that decides
  whether the initiative launches at all. And shadow evidence
  accumulates in calendar time rather than executor time, so the
  earlier the shadow deploys, the more evidence the build collects
  while the spine is built.
- **Architecture spine**: config core, identity and auth, ops seam
  and store, intake, pipeline, streamer drivers and SSO, decision
  queue, versioned API.

The tracks meet at the integration rung, where the matcher wires into
the pipeline and the launch gate is evaluated. The launch cutover and
its close-out end the initiative: beets database migration, box swap,
and—after a deprecation window—retirement of the interim estate
and the salvage tree.

### Shadow: the standing oracle

The still-running interim estate is the shadow oracle. From the
shadow rung until launch, the new matcher runs against the material
the box's live beets pipeline imports, both verdicts are logged under
the estate config, and divergences are adjudicated by the owner into
the same known-acceptable-divergence artifact the corpus uses.

The material is named honestly: the wind-down retired every
contributor intake path, so what feeds the interim import timer is
what Geoff adds for his own listening, and nothing else. The shadow
window is therefore long and thin—more calendar time than the
organizer's 30-day window, over owner-curated rather than contributor
material. Per ruling 1, the floor is a configurable duration
threshold with a 30-day default, every divergence that arrives is
adjudicated, and the observed album count is reported at the launch
gate rather than gated; the corpus's real-catalog and adversarial
strata carry the evidentiary weight a thin shadow cannot.

The oracle is also coarser than a full verdict stream: the box
imports with `beet import -q` and `quiet_fallback: skip`, so live
beets emits a binary strong-or-skip signal, with no recommendation
level and no distance for skipped albums. The shadow log records what
the oracle can actually attest, tier-one evidence rides the applied
subset, and owner adjudication covers the skips.

Shadow constraints: the deployment is additive and read-only with
respect to the library, shadow lookups spend deferred background
MusicBrainz budget through the shared rate limiter, and every box
write follows the quiet-box discipline. The observation mechanism
(how the runner sees the import timer's material) is designed in the
shadow rung's plan; the deployment machinery it needs is priced in
rung 4's band, not assumed.

## Repo pivot and tree shape

At rung 1 the repo pivots once, from the tree as it exists on disk—
the pivot list below was re-derived after review against `ls`, not
the documentation:

- The pre-pivot `cmd/`, `internal/`, `web/`, and `e2e/` trees move to
  `_salvage/`, which the Go toolchain excludes. `e2e/` moves with the
  code it tests: it imports `internal/lock` and `internal/importer`
  and builds `./cmd/dubplate`, so it cannot outlive the move, and its
  coverage of the frozen binary lives on at the pre-pivot tag. That
  tag is the buildable, testable fallback if an emergency guardian
  fix is ever needed during the build.
- The estate operations files stay live and gated: `scripts/`,
  `systemd/`, `cloud-init/`, `compose.yaml`, `config/`, `env/`. They
  are the estate's operational record—they run the interim box
  today and the new box after cutover. They are not an adopter's
  deployment kit; adopter-facing packaging remains the packaging
  initiative's work, after this one. Rung 1 audits this set rather
  than enumerating directories: `config/filebrowser.yaml` belongs to
  a retired integration and goes; there is no bats suite (retired
  2026-08-31 with the shell layer), so nothing claims one.
- Gate rewiring is stated as what it is: the `Makefile`'s Go and web
  legs currently cover the old tree (`check` runs vet, lint, test,
  tidy-check, and three web legs today). Rung 1 repoints the Go legs
  at the greenfield tree, removes the web legs until rung 7 rebuilds
  `web/`, excludes `_salvage/` from every gate leg that walks files
  (the vale comment check included), and repoints the deploy build
  target at `cmd/dubplate-server` so the box deploy path works from
  rung 1 onward—rung 4 depends on it.
- The greenfield tree takes the canonical names immediately, so the
  new code lives its whole life under its final identity and the
  close-out deletes `_salvage/` instead of renaming anything.

The greenfield tree, shaped by the ops-seam architecture:

```
cmd/dubplate-server/     cobra entry, thin
internal/config/         the universal wall, defaults, thresholds
internal/engine/         pure library: no HTTP, no terminal knowledge
    matcher/             penalty components, scoring core, verdicts
    musicbrainz/         WS client, rate limiter, response translation
    tagio/               FLAC tag read and write
    store/               SQLite store, migrations, snapshot backup
    pipeline/            admission, layout, collection routing
    shadow/              the shadow runner and divergence log
internal/ops/            operation vocabulary: interface plus wire types
internal/auth/           tokens, sessions, login, rate limiting
internal/streamer/       driver seam
    navidrome/           reference driver
internal/intake/         tus sessions, gate policy, journeys
internal/reject/         the copy contract (payloads and emails share it)
internal/notify/         optional mail seam
internal/serve/          HTTP: handlers, ingress proxy, embedded web
web/                     Svelte reference client, fresh tree
```

The tree is a shape, not a lockfile: each rung's spec may split or
merge leaves below this level, and the ops seam's package boundaries
are the ones reviewers hold hardest. The `dubplate-server` cobra
surface is re-derived at rung 1 from the organizer spec's subcommand
list, minus `beet` passthrough, which never exists on the new server.
The shadow runner is a first-class leaf because it outlives its rung:
it is the migration-shadowing capability any beets-migrating adopter
wants, not scaffolding.

**Reuse stance.** Proven leaf packages are adoption candidates,
admitted on merit under fresh review at the rung that needs them:
`atomicfile` and the gate policy as code, under the exemplary-Go bar.
Two candidates are data contracts rather than code, and they are
adopted as contracts: the reject copy strings come verbatim from the
ratified string authority (`research/passb-copy-final.md`, the
content-review-passed deck), and changes to them are register-gated,
never a side effect of a code bar; the legacy ledger's on-disk format
is captured as a documented fixture at rung 10, while the salvage
tree is still legible, because the migrator must read it and may
import nothing from `_salvage/`. The uploader's fidelity-passed web
components, tus engine, and checks are source material for the
intake rung. Adoption means copying into the greenfield tree under
the exemplary-Go and idiomatic-Svelte bars; nothing is imported from
`_salvage/`.

## Auth birth requirements

The login spec is source material; its build-on-old-serve framing,
its one-compact-pass build order, and—wherever the requirements
below conflict with it—its design are superseded. Its two
adversarial reviews are birth requirements: the auth and streamer
rungs are born with these constraints, they never exist in a weaker
form, and the rung plans walk the full review texts
(`research/2026-09-02-login-spec-reviews/security.md` and
`universality.md`) as checklists, minors included. The load-bearing
requirements, grouped:

**The trust boundary, serve-as-ingress.** Serve strips every inbound
identity-header spelling on ingress, before routing, for every path,
the `/rest/*` and `/share/*` bypasses included; injection happens
only on the session-gated proxy path. The strip set is driven by the
driver's configured header name plus a fixed extras list, underscore
variants included. The streamer's trusted source is pinned to serve's
single observed address, never a bridge subnet, and no other network
path may reach the streamer. The ingress flip precedes enabling the
streamer's externalized auth, stated as an ordering constraint
wherever the cutover is planned.

**The trust boundary, forward-auth.** The forward-auth topology gets
its own boundary, not an inherited one, because in that deployment
serve is not in the streamer's request path and every serve-side
guarantee above is inert. Requirements: the driver publishes its
strip set to the adopter, not only to serve's ingress; the adopter's
proxy must strip that set before the verify subrequest and before the
streamer; the streamer trusts only that proxy; and `dubplate doctor`
proves the boundary by verifying that a spoofed identity header from
a non-trusted source is rejected. The forward-auth documentation
carries this as setup steps, and the verify endpoint's docs state
plainly that the boundary is the adopter's proxy.

**One owned prefix.** dubplate's entire HTTP surface, API included,
lives under exactly one configured prefix. The streamer owns
everything else. The auth boundary is decided by configuration, never
by mux-miss. This rule binds the serve design from rung 1.

**Link consumption.** The magic link GET renders an interstitial with
the token in a form field and `Cache-Control: no-store`; POST
consumes, origin-checked. Consumed nonces are durable in the state
directory with TTL pruning; verify the MAC, then expiry, then insert
the nonce. Consumption failures return one generic outcome, with the
distinction only in server logs. Tokens never appear in URLs that get
logged, which the GET/POST split is what makes true.

**Sessions.** HMAC-signed `__Host-` cookie, `Secure`, `HttpOnly`,
`SameSite=Lax`, stripped before proxying to the streamer. The
`__Host-` prefix makes the session host-scoped by definition, so
same-origin—one hostname fronting serve and, under SSO, the
proxied streamer—is a stated precondition of the SSO capability. A
split-host adopter runs without the SSO capability or behind their
own forward-auth front; no cookie-domain widening is offered, because
a parent-domain cookie is the weaker design and the demand is
hypothetical. An absolute lifetime cap bounds rolling refresh,
refresh counts same-origin requests only, and a per-email valid-after
epoch in the state directory gives per-person revocation of the
dubplate session. Offboarding is stated at its true size: the epoch
revokes dubplate; streamer credentials survive it, because `/rest/*`
bypasses serve by design, so the runbook carries an explicit
streamer-deprovision step, and a driver deprovision capability is a
candidate when a driver's admin API supports one. The verified
identity carries its assertion channel (cookie, header, or service
token) through the verify-identity seam, so the organizer's rule that
admin verbs reject cookie-derived identity is enforceable at the
seam. The seam exists as an interface from the auth rung, with
magic-link as its reference implementation and forward-auth fronts as
alternates.

**Secrets.** `DUBPLATE_MAGIC_LINK_SECRET` and
`DUBPLATE_SESSION_SECRET` carry a length floor enforced at config
load, domain-separation labels inside the MAC input, unambiguous
token serialization, and a `_PREVIOUS` rotation overlap so rotation
is routine rather than an estate-wide sign-out. `DUBPLATE_PUBLIC_URL`
is required whenever magic-link is the active identity source; link
hostnames never derive from request headers.

**CSRF, stated once.** Every cookie-carrying, state-changing surface
takes the origin check: the login POST, the consumption POST, logout,
and the proxy path. This is the organizer's blanket CSRF requirement
landing as one rule at the auth rung rather than an enumeration per
endpoint.

**Rate limits and abuse.** Limits key on the real client IP, read
from the trusted-proxy pair when the proxy mode is enabled and from
`RemoteAddr` when it is not; the mode is explicit configuration, and
`doctor` reports which is live. Emails are validated, length-capped,
and lowercased before keying. Every auth endpoint body is size-capped,
limiter structures are bounded and evicting, and every auth event
(link requested, consumed, failed, rate-limit trip, header-strip hit)
is logged with the real client IP. The enumeration-kindness answer
becomes a flag, kind by default. `/rest/*` is recorded as what the
design makes it: an anonymous surface reachable through serve whose
authentication is the streamer's own; serve is its only reachable
path, and the runbook names the residual.

**Identity and roster.** The roster validates at load: unique names,
addresses parsed with `net/mail.ParseAddress`, no control characters,
bounded length. The streamer principal is split from the inbox
directory name: an optional per-contributor `streamer_username` with
startup and `doctor` collision checks. The principal—whichever
field goes on the wire—is validated as a stable identifier:
ASCII-printable, unique, bounded; display names stay free.

**One identity source.** Exactly one identity source is active per
deployment, selected by config. A forward-auth front (Access
included) is an integration with an enable flag that makes its config
required-when-enabled; the inactive source is not consulted at all.

**Post-login destination.** The destination is a driver capability
with a core default, because it carries driver semantics (Navidrome's
auto-provisioning fires on the player root): with SSO enabled the
driver declares it, without SSO the dubplate root serves, and with no
streamer the journey's honest terminal state stands. The
redirect-target validation rule (single-decode, relative-only, no
scheme, no `..`, no backslash variants) travels with the mechanism.

**Proxy hygiene** (streamer rung). `httputil.ReverseProxy` via
`Rewrite` with `SetXForwarded`; identity injected inside `Rewrite`;
session cookie and hop-by-hop hazards stripped; path normalization
rejects raw/decoded mismatches and `..` segments before prefix
routing; per-route write-timeout handling for SSE and audio streams;
identity-injected responses are `Cache-Control: private` and never
edge-cached; the origin check covers the proxy path, and the proxy
refuses to inject on non-idempotent requests lacking a same-origin
signal. The streamer's address is configured once, in driver config,
with per-mode semantics stated there (the proxy target under SSO, the
public base URL without it); the album deep-link capability derives
from that configuration, and no second ad-hoc URL enters the wall.

## The ladder

Rungs are sequential with one executor and no overlap. Magnitude
bands are relative and carry no time units: M is one pass under one
plan at roughly the repo's recent pass ceilings, S is well under
that, L splits into multiple passes at its own planning. Each rung
gets its own spec and plan when its turn comes, and each plan header
carries a token ceiling and checkpoint interval. Evidence-track rungs
are marked [E].

1. **Foundation** (M). The repo pivot and salvage freeze from the
   re-derived list above, with the ops-file audit; the greenfield
   `internal/config` implementing the derived wall table, the
   enable-flag discipline, the threshold registry, and the
   loud-compat rule (U5); the `cmd/dubplate-server` skeleton with its
   re-derived subcommand surface; `check` and `doctor` scaffolds; the
   gate and deploy build path rewired as stated in the pivot section;
   ROADMAP restated for the pivot.
2. **Corpus and harness** (L) [E]. The corpus case grows from the
   review round: a case is input (with item order pinned), frozen
   MusicBrainz responses, beets config, the evaluation date, and the
   expected verdict plus the expected applied tag set—with the
   corpus manifest pinning the beets, unidecode, jellyfish, and lap
   versions, since transliteration tables sit inside the distance
   arithmetic. Deliverables: the frozen-verdict fast gate with no
   Python in it; the version-bump differential job against pinned
   beets and a local stub, with everything that knows how to invoke
   beets isolated behind one adapter package (U7, priced against the
   planned compatibility-shedding beets 3.0); the capture tool, whose
   fidelity requirements include the pseudo-release double fetch, the
   maximal-practical `inc=` raw-JSON capture (U6), the
   `musicbrainz.host` stub seam, and the client's hard 0.25s
   inter-request floor in differential wall-time budgets; the
   assignment-tie detector (perturb and re-solve; tie cases are
   excluded or pinned, because equally-optimal assignments change
   final VA distances); the threshold-proximity flag (any case whose
   best distance or candidate gap sits within about 1e-6 of a verdict
   boundary is flagged, because strict-versus-inclusive comparisons
   admit no epsilon); the known-oracle-bug ledger (U1, pending Open
   ruling 2) seeded with the study's known-defect patterns as
   adversarial cases; the component-test derivation tooling that
   turns beets' own distance tests into per-component expected
   values; and two U8 requirements on the fixture format itself:
   the cassette request-matching rule is stated explicitly, because
   MusicBrainz query construction is conformance surface, and the
   format admits recall-extension cases from day one (fixtures for
   lookups beets never issued, conformance still checked on the
   shared candidate subset), so a post-launch candidate-retrieval
   improvement is graded by the same harness rather than escaping
   it. The registered MusicBrainz application identity and
   User-Agent land here, and the quarterly re-capture chore is
   documented here. Strata per the research report: roughly 150 to
   300 cases, about 40% structural, 50% real catalog, 10%
   adversarial, sized against tier one; the estate config is the
   primary oracle and the stock-config stratum's gating role is Open
   ruling 3.
3. **Go matcher** (L) [E]. The penalty components and scoring core
   per the evolvability ruling; album-level matching with track
   assignment pinned to the pinned-beets solver's tie behavior on
   corpus cases; the WS-response-to-matcher-input translation layer
   as a named deliverable—key normalization, per-medium track
   extraction, pregap and data/video-track handling, alias-driven
   title and artist rewriting, and release-event selection, all of
   which move verdicts before any distance is computed; stable
   candidate ordering; distance agreement at the bit-identical
   target (fixed accumulation order, compared at 1e-12 relative,
   with the rung-2 proximity flag guarding the boundaries); the
   verified-complete tunable surface from the mechanics study—
   thresholds, distance weights, preferred, ignored, required,
   ignored_media, track-length grace and max, data- and video-track
   ignoring, search limit, source mismatch penalty, timid, languages
   and ignored alias types, extra tags, and the verdict-to-action
   settings—with the evaluation date an explicit input, never a
   clock read; and the shared MusicBrainz rate limiter shipping the
   per-caller budget interface, with the priority policy re-verified
   at rung 9 when its real callers exist. Behind the harness until
   the bar holds.
4. **Shadow** (M to build, continuous to launch) [E]. The shadow
   runner under `engine/shadow` and its subcommand; the divergence
   log (durable in the state directory; the store arrives later) and
   the adjudication artifact; and the deployment the band now prices:
   a hardened systemd unit per vps-conventions with pre-created
   `ReadWritePaths`, its environment file, and the deploy.sh path
   rung 1 repointed. The observation mechanism is designed in this
   rung's plan. From here the evidence clock runs unattended while
   the spine builds.
5. **Identity and auth** (L). `internal/auth` built to the birth
   requirements: tokens, sessions with epoch and absolute cap,
   durable nonces, the login and consumption pages (Go-templated, no
   JavaScript, shared design language), SMTP send, rate limiting, the
   CSRF rule, the verify-identity seam with provenance, the
   forward-auth verify endpoint with its own trust-boundary
   documentation, and a minimal serve shell hosting it under the one
   owned prefix. The `web-auth-security-reviewer` gate runs inside
   the rung.
6. **Ops seam and store** (M). The ops vocabulary: one Go interface
   plus JSON-serializable request and response types, the job model
   and explicit cancellation from the first version, and the
   operation-availability mechanism—the organizer's named guard
   against ad-hoc parity opt-outs—as part of the interface. The
   SQLite store on the modernc driver with migrations that run
   forward on open; paths stored library-relative and byte-preserving
   from the first schema (U3), with intake's non-UTF8-filename policy
   stated where paths enter; the per-track metadata storage decision
   (typed columns by migration versus a side table) made here,
   because rung 11's built-ins consume it (U4); snapshot backup and
   restore against rclone remotes behind the backup flag. The
   contract-test suite arrives with an honest scope: it exercises the
   ops interface in-process and over HTTP through an in-test client;
   full two-implementation parity enforcement completes post-launch
   when `ops/client` exists, and until then the suite is the wire
   contract's regression net, not a parity proof.
7. **Intake** (M). tus transport, the gate policy, roster loading
   with the birth-requirement validation, journey records, the reject
   copy contract adopted verbatim from the string authority with
   register-gated changes, and the uploader reference web client
   built from the salvaged components under fresh review. Format
   acceptance is CONFIG, not policy baked in (Geoff, 2026-09-03): an
   accepted-format allowlist in the gate policy, the estate set to
   FLAC-only, other adopters free to admit lossy formats; the
   shipped default and the reject-copy wording per format are this
   rung's design details, and the tag I/O scope (rung 8) plus the
   matcher are format-agnostic upstream of it. The `web/`
   legs rejoin the repo gate here. The album is the verdict unit: one
   failing track rejects the batch, and no partial album is ever
   admitted. This rung carries an attended moment the process section
   counts: the uploader UI gates on the visual-fidelity chain,
   Geoff's before/after included. Theming ruling (Geoff, 2026-09-03),
   binding here and on every later web surface: theming is data,
   never code. Every component is authored against DaisyUI theme
   variables from day one (no hardcoded colors); the theme-variable
   contract is the documented customization API (class names and DOM
   are not); three tiers ship over the build: per-user theme choice
   (data-theme), operator branding as config-rendered CSS variables
   at serve time (no Node, no rebuild, per universality), and an
   optional operator CSS file served after the bundle. User-supplied
   components or template overrides stay ruled out (ruling 8's
   no-plugin-ABI fork). Starting theme roster (Geoff, 2026-09-03; a
   starting roster, revisable cheaply until the picker ships): dark
   (the shipped default for adopters) and light as the
   system-following pair, dim, night, black (the high-contrast
   dark), nord, coffee, retro (the personality pick), plus a custom
   gruvbox theme authored from the published Gruvbox palette
   standard - the estate's configured default, chosen to pair
   exactly with Navidrome's built-in Gruvbox Dark (Geoff,
   2026-09-03). Standing principle from the same ruling: custom
   themes are a variable block by design, so the roster grows a
   MATCHING theme for each streamer a driver ships (nord and
   dracula already exist as shared standards on both sides;
   future drivers bring their pairing on demand), keeping any
   estate visually coherent from uploader to player. The
   cross-platform evidence and the lean popularity-weighted
   roster (gruvbox pair, nord, dracula, catppuccin; the rest
   deferred to the escape hatch) live in
   research/2026-09-03-theme-collection-brief.md, a standing
   input to this rung's plan. Added requirement (Geoff,
   2026-09-03): the roster also carries a matching theme for
   each target streamer's FACTORY DEFAULT, since the common
   adopter never re-themes the streamer - one matching
   Navidrome's Dark (its verified default) and one matching
   Jellyfin's dark, both extracted from the apps' own open
   theme sources as variable blocks. Theming is more than color
   (Geoff, 2026-09-03): a theme is the FULL DaisyUI variable
   block (palette, radii, border weight, depth, noise, sizing)
   plus a declared font pairing; fonts ship embedded (woff2 via
   go:embed, never a CDN, per universality and offline use), and
   matched themes match the streamer's typography where its
   fonts are open (Navidrome: Roboto; Jellyfin: Noto Sans).
   Motion stays global and prefers-reduced-motion-aware, not
   per-theme. Config model (Geoff, 2026-09-03): theming is TWO
   orthogonal picks, the system MATCH (the structural layer:
   typography, radii, depth, sizing - navidrome, jellyfin, or
   dubplate's own) and the color PALETTE - composed at runtime
   as separate variable layers (palette blocks set only color
   variables, match blocks only structural ones), so the axes
   multiply without multiplying authored themes. Estate config:
   match navidrome, palette gruvbox. How DaisyUI's bundled
   built-ins reconcile into the two-axis model is a rung 7
   design detail.
8. **Tag I/O** (M). Pure-Go FLAC tag writing for the server scope,
   proven against the corpus's expected applied tag sets from rung 2,
   plus the wazero-taglib spike whose recorded decision the
   post-launch local organizer will consume. Writing tags is the
   pipeline's primary side effect, so this rung precedes the pipeline
   rather than folding into it. The rung's plan opens with the
   write-side close read the mechanics study scoped out (the
   apply-metadata field mapping and the mediafile layer).
9. **Pipeline** (L). Admission, tagging, layout, and collection
   routing on the engine and store; the notifications and monitoring
   seams as optional integrations, built immediately ahead of the
   pipeline that consumes them; the repair verbs (query, retag
   against a named MusicBrainz release, move), which are launch-bar
   admin surface because the new server has no beets toolkit to lean
   on; the concurrency model decision (per-album lease or explicit
   queue, with the constraint that interactive work must not starve
   timer imports), settled in this rung's spec; and re-verification
   of the rate limiter's caller-priority policy against its real
   callers. The honest terminal state "imported, no player
   configured" lands here as journey vocabulary (a relocation from
   the organizer's streamer-seam assignment, listed in the
   supersessions).
10. **Migration and layout parity** (M). The launch bar's fourth
    part gets its evidence machinery here, ahead of the cutover: the
    migrator against the organizer's four failure modes (BLOB paths
    never round-tripped through Go strings, NULL singleton album
    IDs, ledger reconciliation, the no-move dry-run); the legacy
    ledger format captured as a documented fixture while `_salvage/`
    is legible; layout-parity fixtures proving the beets path-format
    DSL and query-prefixed collection routing reproduce the estate's
    computed paths—engine work, because any beets-migrating
    adopter faces the same question; a corpus-driven path-parity
    fixture (carried in 2026-09-02, Geoff): render the path formats
    over rung 2's corpus AppliedFields records through pinned beets
    and through the new engine, and diff the renderings
    case-for-case, so organization parity is evidence rather than
    review; and the first read-only no-move
    diff against the live library (read-only listing access, R2
    snapshots as the alternative source). A non-empty diff here is a
    rung 9 defect found cheap; the same assertion re-runs at rung 16
    as a gate, not a discovery.
11. **Built-ins and lifecycle hooks** (M). The built-in set
    re-derived against the plugin survey and the estate's real
    config, under the fixed policies: scrapers refused,
    fingerprinting and ReplayGain probe-backed. Lifecycle hooks at
    the named pipeline events, JSON on stdin in the ops wire
    vocabulary, failure never failing an import.
12. **Streamer driver and SSO** (L). The driver seam (scan trigger,
    scan status, album deep link, capability reporting,
    driver-declared post-login destination) with the Navidrome
    reference driver owning the bypass path set, the trusted header,
    the externalized-auth config names verified against the pinned
    Navidrome version, logout wiring, and provisioning choreography
    with the first-admin `doctor` preflight; the streamer address
    semantics resolved in driver config as stated in the birth
    requirements; and the ingress reverse proxy built to the
    proxy-hygiene requirements, behind the SSO enable flag, with the
    availability coupling recorded at the seam. The band is L
    because the rung carries the initiative's densest security
    surface, including three carried blockers from the login
    security review. Estate theming note (Geoff, 2026-09-03): set
    Navidrome's default UI theme to its built-in Gruvbox Dark via
    config, pairing with the estate's custom DaisyUI gruvbox theme
    (ladder entry 7); verified against v0.63.2's shipped theme
    roster. Navidrome has no custom CSS hook and proxy-side CSS
    injection stays refused as fragile; the shared published
    palette standard is what makes the pair exact. Target-system
    ruling (Geoff, 2026-09-03): the streamer targets are Navidrome
    and Jellyfin; the Subsonic family is demoted from
    supported-by-design to reachable-through-the-seam-on-demand
    (rationale and evidence in ROADMAP's streamer note and the
    theme-collection brief).
13. **Decision and review queues** (M). Persisted decision records
    serializing candidates, distances, and penalty breakdowns; the
    queue verbs (`list pending`, `answer`, `defer`) with a minimal
    CLI answering path; the custody-policy enum (`strict` default,
    `owner-queue`, `contributor-picks`) and the classical-appeal
    action that rides every setting; and the admin review queue, the
    distinct low-volume side door for timer-path material.
14. **Versioned API and admin spine** (M). The contract document,
    `/api/v1` under the owned prefix, the additive-within-version
    policy, the authorization model (AdminEmail, header-only admin
    verbs enforced through provenance, service tokens for unattended
    machine verbs), and non-browser token acquisition on the session
    vocabulary, so native clients and the post-launch TUI have a
    credential path.
15. **Integration** (M). The tracks meet: the matcher wires into the
    pipeline, the contract suite runs at its stated scope, the
    conformance gate and shadow record are evaluated against the
    launch bar, and every page on the user path is rendered and read
    before any attended ask. The domain reviewer fan-out for the
    whole surface runs here.
16. **Launch cutover** (M). The box swap under a written, exercised
    rollback; the beets database archived to R2; systemd, deploy.sh,
    compose, and cloud-init converged for the new server per
    vps-conventions (the estate's record, per the pivot section, not
    an adopter kit); secrets minted through the age-store flow; the
    ingress-before-ExtAuth ordering honored; the committed
    universality breach (`UIWelcomeMessage`) resolved through
    configured `SiteName`; the no-move assertion re-run as a gate.
    The rollback target outlives the rung: beets, its database
    snapshot, and the old units stay resident and disabled through a
    deprecation window, the pre-cutover estate is re-installable
    from the pre-pivot tag, and reverting restores the snapshot and
    re-imports post-swap sessions from the session records, per the
    organizer's rollback shape. One attended gate: Geoff verifies
    login, upload, and player on his devices in a single sitting. On
    his yes, the family launch.
17. **Launch close-out** (S, after the deprecation window). The
    interim estate's remnants and `_salvage/` retire, the beets
    deprecation window closes, ROADMAP and the organizer spec
    reconcile, and the initiative's post-mortem and ledger entries
    land.

## Rulings on the review round

All three resolved by Geoff, 2026-09-02.

1. **Shadow floor, ruled through the universal-adopter lens.**
   Geoff's ruling asked what most users of the system would want, and
   the answer dissolves the estate-shaped floor. What every user
   inherits is the corpus: it ships with the product and is the
   matcher's standing guarantee. Shadow is the local confidence tool
   an adopter runs against their own live beets before trusting the
   swap, sized by their own material, which makes a fixed album count
   estate idiosyncrasy where a tunable belongs. Resolution: the
   shadow window is a configurable duration floor, default 30 days;
   every divergence that arrives is adjudicated; the observed album
   count is reported at the launch gate, not gated; the corpus's
   real-catalog and adversarial strata carry the evidentiary weight.
2. **Oracle-bug policy (U1), accepted as proposed.** Tier one is
   measured against release truth; where the oracle is
   known-defective, the case enters the known-oracle-bug ledger with
   an adjudicated expected verdict, and reproducing the oracle's
   defect fails the gate. Conformance is to beets' behavior except
   where beets is documented wrong.
3. **Stock-config gating, accepted as proposed.** Tier one (zero
   false confident matches) also holds over the stock-config stratum;
   tier two remains estate-scoped. The launch claim is honest for
   both audiences without doubling the adjudication load.

## What this spec supersedes

Against the organizer spec, which otherwise governs:

- The build-order ladder (its rungs 1 through 11) is replaced by the
  ladder above. Its rung 10 (two-binary split, local TUI) and the
  remote-TUI half of rung 11 move past launch, unscheduled here.
- Staged adoption through the `internal/beets` subprocess seam is
  superseded. The greenfield server never embeds beets; the fallback
  for a failed matcher is the interim estate continuing to serve
  playback while the family launch waits.
- The `beet` passthrough and its retirement condition are moot on the
  new server. The repair verbs are launch-bar surface.
- The shadow window is superseded per ruling 1.
- The header's preconditions (uploader ships, prod-polish completes)
  are superseded by the pivot.
- "Cloudflare Access is the reference implementation" of the
  verify-identity seam flips: magic-link is the reference, Access is
  an optional forward-auth front (the login spec's amendment 1,
  folded here).
- The unassigned streamer-seam extraction, its config-wall change,
  and its journey-vocabulary amendment are all assigned: rungs 12, 1,
  and 9 respectively. The Navidrome variables stop being
  required-with-no-default and move behind the streamer flag.
- The admin-spine rung's auth-seam generalization splits: the seam
  itself is born at rung 5; non-browser token acquisition stays at
  rung 14.

Against the login spec: the build-on-old-serve framing and the
one-compact-pass build order are superseded, and wherever this spec's
birth requirements conflict with that document's design (memory-only
nonces, rotation-only revocation, the `/api/*` routing rule, the /24
trusted source, roster `name` on the wire, enumeration kindness as
fixed behavior), the birth requirements win. A superseded-status note
at the top of the login spec points readers here. Its design content
otherwise survives as source material for rungs 5 and 12.

## Box safety during the build

The box is production for Geoff's own listening. The guardian timers
and R2 backups protect the library and beets database throughout.
Rungs 1 through 3, 5 through 9, and 11 through 15 touch the box not
at all; rung 4 deploys the additive shadow runner under its own unit;
rung 10's no-move diff needs read-only access to the library listing
(or the R2 snapshot); rung 16 is the one rung that changes what the
box serves, and it carries the quiet-box check, the `--check`-first
deploy discipline, and the exercised rollback. Every executor
entering this repo checks for live executors first.

## Out of scope

Everything the organizer spec excludes: the web admin page,
multi-tenancy, in-process plugins, scrapers, two-way sync and the
mirror, copy localization, native mobile clients. The local
`dubplate` binary, the TUI in both modes, the two-binary split, and
publish are post-launch work, outside this initiative. The packaging
initiative remains sequenced after this one, and adopter-facing
deployment artifacts are its work, not rung 16's.

## Process

Each rung opens with its own spec where the ladder marks decisions
unsettled, and a plan via writing-plans; L rungs split at their own
planning. Execution runs the implementer, diff-reviewer, gate chain
per task, with the domain fan-out at pass ends and the workflow mode
at six or more tasks. Both budgets are scored at every pass close.

Owner attention is priced honestly rather than claimed away. The
initiative carries one attended gate (rung 16's single-sitting
verification), one attended visual check (rung 7's uploader
before/after, per the visual-fidelity mandate), and a standing
adjudication stream: corpus divergences at rungs 2 and 3, shadow
divergences from rung 4 onward, each batch presented at checkpoints
rather than as it occurs. An unadjudicated divergence blocks the
launch gate, so adjudication batches are part of the critical path
and are kept small and regular. Everything else that needs Geoff is
batched questions at checkpoints.

Sessions run in arcs (Geoff's process ruling, 2026-09-02, folded
after rung 1): foundation (rung 1), evidence (2-4), spine (5-9),
surfaces (10-14), launch (15-17), one session per arc with no clears
or re-brainstorms inside one and a close-out (post-mortem, STATUS
handoff) between. Plans stay just-in-time and are drafted and
adversarially reviewed in the shadow of the previous rung's
execution; front-loading plans is refused on the evidence that plan
staleness was the review rounds' dominant defect class. Workflow mode
is a standing grant for execution and prep across the whole build,
with human questions batched and unblocked work continuing when the
owner is unavailable. CLAUDE.md's "Build process" section carries the
operational form every session and subagent loads.
