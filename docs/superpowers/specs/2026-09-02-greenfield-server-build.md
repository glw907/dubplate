# dubplate-server: greenfield build spec

Status: drafted 2026-09-02 from the ratified rulings
(`docs/superpowers/research/2026-09-02-greenfield-server-rulings.md`)
and the inputs they list. Awaiting Geoff's review. Per the process rule
of 2026-09-02, no adversarial review round and no rung plan exists or
runs until he has read this document.

Sources folded: the organizer spec
(`2026-09-01-dubplate-organizer-design.md`, amended through
2026-09-02), the login spec (`2026-09-02-magic-link-login-design.md`)
as source material, both login-spec adversarial reviews
(`research/2026-09-02-login-spec-reviews/`) as auth birth
requirements, the Navidrome integration research, the conformance
corpus report (`2026-09-01-beets-conformance-corpus.md`), and the
uncertain-match UX report (`2026-09-01-uncertain-match-ux.md`).

## What this initiative builds

dubplate-server is built fresh, foundation first. It is the universal
server of the organizer spec: intake and admission, the organizing
engine with a native Go MusicBrainz matcher, the library store, the
review and decision queues, backup and restore, notifications, the
admin API, the documented client API, and login. It is not an
evolution of the current code. The existing `internal/` and `web/`
trees are salvage: adoption material, admitted piecewise on merit,
retired at launch.

The organizer spec remains the governing architecture. Its success
bar, purview, seams, custody policy, extensibility fork, settled
forks, harness design, and beauty section apply unchanged. This spec
re-derives what greenfield changes: the ladder, the adoption story,
and the launch shape. Section "What this spec supersedes" lists the
exact deltas.

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
   (tier one, no adjudication available), and zero unadjudicated cases
   where beets is right and Go is wrong (tier two). Both tiers count
   cases, never percentages.
3. **A clean shadow record.** The shadow run defined below is clean:
   zero unadjudicated divergences and zero false confident matches
   against live beets decisions.
4. **The no-move assertion.** A dry-run diff proves the new pipeline's
   computed paths equal the on-disk layout of the live library before
   anything writes.

There is no partial launch and no beets fallback inside the new
server. If the matcher cannot clear the bar, the honest outcome is
that the interim estate keeps running and the launch waits. That is
why the evidence track runs first: the matcher is the initiative's
existential risk, and every spine rung built after a failed matcher
would be inventory, not product.

The relationship with beets outlives the launch. The initiative
succeeds or fails on adopting beets' battle-tested matching behavior,
and the harness is built so dubplate can keep evolving with it: the
version-bump differential job survives as a standing tool, so when a
newer beets improves its matcher, the divergence surfaces as a
harness finding and is adopted deliberately, adjudicated under the
success bar rather than followed automatically. Everything else the
server does is comparatively well-understood engineering; the
evidence track is where the initiative's uncertainty lives, and the
ladder prices it that way.

## Universality

No machinery specific to one estate, anywhere. The login spec's four
tests apply to every rung:

- No hostname, path, email, or personal name outside
  `internal/config`. Site-facing copy derives from configured identity
  (`SiteName`, admin contact), the rule the reject emails follow.
- Drivers own driver knowledge. The streamer bypass paths, trusted
  header name, injection rules, logout wiring, and provisioning
  choreography live with the Navidrome driver, never in core routing.
- Every tunable is a threshold with a sane default: session lifetime,
  link expiry, rate limits, match thresholds, enumeration behavior.
- The design makes sense with no Cloudflare account: plain SMTP plus
  any streamer, tunnel optional, forward-auth proxies optional in
  front, real-client-IP handling configured rather than assumed.

Integrations get explicit enable flags, and a flag makes its
correlated config required-when-enabled. The required-no-defaults wall
covers only what the server owns.

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
implementation is built in this initiative.

## Two tracks, dependency-merged

Approach A, ratified. One executor works the rungs sequentially; the
two tracks are dependency lanes, not parallel executors.

- **Evidence track**: corpus and harness, then the Go matcher, then
  shadow. Independent of the spine from day one, and scheduled first,
  for two reasons. The matcher is the risk that decides whether the
  initiative launches at all. And shadow evidence accumulates in
  calendar time rather than executor time, so the earlier the shadow
  deploys, the more free evidence the build collects.
- **Architecture spine**: config core, identity and auth, ops seam and
  store, intake, pipeline, streamer drivers and SSO, decision queue,
  versioned API.

The tracks meet at the integration rung, where the matcher wires into
the pipeline and the launch gate is evaluated. The launch cutover
closes the initiative: beets database migration, box swap, and
retirement of the interim estate and the salvage tree.

### Shadow: the standing oracle

The still-running interim estate is the shadow oracle. From the shadow
rung until launch, the new matcher runs against the same material the
box's live beets pipeline imports, both verdicts are logged under the
estate config, and divergences are adjudicated by the owner into the
same known-acceptable-divergence artifact the corpus uses. This
supersedes the organizer spec's 30-day, 20-album shadow window with
something strictly stronger: the window is the whole build, and the
organizer's floor (at least 30 days and at least 20 real albums,
whichever completes later) is kept as the minimum the launch gate
accepts.

Shadow constraints: the deployment is additive and read-only with
respect to the library, shadow lookups spend deferred background
MusicBrainz budget through the shared rate limiter, and every box
write follows the quiet-box discipline. The shadow mechanism (how the
runner observes the import timer's material) is designed in the shadow
rung's plan, not here.

## Repo pivot and tree shape

At rung 1 the repo pivots once, mechanically:

- The pre-pivot `cmd/`, `internal/`, and `web/` trees move to
  `_salvage/`, which the Go toolchain and the gate exclude. Salvage is
  reading material, frozen at the pivot commit. The pre-pivot tag is
  the buildable fallback if an emergency guardian fix is ever needed
  during the build.
- The estate operations files stay live and gated, because they run
  the interim box today and the new box after cutover: `scripts/`,
  `tests/*.bats`, `systemd/`, `cloud-init/`, `compose.yaml`,
  `config/`, `env/`.
- The greenfield tree takes the canonical names immediately, so the
  new code lives its whole life under its final identity and the
  launch deletes `_salvage/` instead of renaming anything.

The greenfield tree, shaped by the ops-seam architecture:

```
cmd/dubplate-server/     cobra entry, thin
internal/config/         the universal wall, defaults, thresholds
internal/engine/         pure library: no HTTP, no terminal knowledge
    matcher/             distance and penalty model, candidates, verdicts
    musicbrainz/         WS client, shared rate limiter, cassette replay
    tagio/               FLAC tag read and write
    store/               SQLite store, migrations, snapshot backup
    pipeline/            admission, layout, collection routing
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

**Reuse stance.** Proven leaf packages are adoption candidates,
admitted on merit under fresh review at the rung that needs them:
`atomicfile`, the reject copy contract, the gate policy, the ledger
format. The uploader's fidelity-passed web components, tus engine,
checks, and reject copy machinery are source material for the intake
rung. Adoption means copying into the greenfield tree under the
exemplary-Go and idiomatic-Svelte bars; nothing is imported from
`_salvage/`.

## Auth birth requirements

The login spec is source material; its build-on-old-serve framing is
superseded. Its two adversarial reviews are birth requirements: the
auth and streamer rungs are born with these constraints, they never
exist in a weaker form, and the rung plans walk the full review texts
(`research/2026-09-02-login-spec-reviews/security.md` and
`universality.md`) as checklists, minors included. The load-bearing
requirements, grouped:

**The trust boundary.** Serve strips every inbound identity-header
spelling on ingress, before routing, for every path, the `/rest/*` and
`/share/*` bypasses included; injection happens only on the
session-gated proxy path. The strip set is driven by the driver's
configured header name plus a fixed extras list, underscore variants
included. The streamer's trusted source is pinned to serve's single
observed address, never a bridge subnet, and no other network path may
reach the streamer. The ingress flip precedes enabling the streamer's
externalized auth, stated as an ordering constraint wherever the
cutover is planned.

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
`SameSite=Lax`, stripped before proxying to the streamer. An absolute
lifetime cap bounds rolling refresh, refresh counts same-origin
requests only, and a per-email valid-after epoch in the state
directory gives per-person revocation and an offboarding step that
actually offboards. The verified identity carries its assertion
channel (cookie, header, or service token) through the verify-identity
seam, so the organizer's rule that admin verbs reject cookie-derived
identity is enforceable at the seam. The seam exists as an interface
from the auth rung, with magic-link as its reference implementation
and forward-auth fronts as alternates.

**Secrets.** `DUBPLATE_MAGIC_LINK_SECRET` and
`DUBPLATE_SESSION_SECRET` carry a length floor enforced at config
load, domain-separation labels inside the MAC input, unambiguous token
serialization, and a `_PREVIOUS` rotation overlap so rotation is
routine rather than an estate-wide sign-out. `DUBPLATE_PUBLIC_URL` is
required when login is enabled; link hostnames never derive from
request headers.

**Rate limits and abuse.** Limits key on the real client IP, read from
a configured trusted-proxy header honored only when the peer is the
configured proxy, with `RemoteAddr` fallback for proxyless adopters.
Emails are validated, length-capped, and lowercased before keying.
Every auth endpoint body is size-capped, limiter structures are
bounded and evicting, and every auth event (link requested, consumed,
failed, rate-limit trip, header-strip hit) is logged with the real
client IP. The enumeration-kindness answer becomes a flag, kind by
default.

**Identity and roster.** The roster validates at load: unique names,
addresses parsed with `net/mail.ParseAddress`, no control characters,
bounded length. The streamer principal is split from the inbox
directory name: an optional per-contributor `streamer_username` with
startup and `doctor` collision checks, because one string doing
filesystem, streamer-account, and auth-principal duty is the exact
idiosyncrasy the universality ruling forbids.

**One identity source.** Exactly one identity source is active per
deployment, selected by config. A forward-auth front (Access included)
is an integration with an enable flag that makes its config
required-when-enabled; the inactive source is not consulted at all.
Adopters with their own proxy get a forward-auth verify endpoint
rather than being forced through serve's proxy topology.

**Proxy hygiene** (streamer rung). `httputil.ReverseProxy` via
`Rewrite` with `SetXForwarded`; identity injected inside `Rewrite`;
session cookie and hop-by-hop hazards stripped; path normalization
rejects raw/decoded mismatches and `..` segments before prefix
routing; per-route write-timeout handling for SSE and audio streams;
identity-injected responses are `Cache-Control: private` and never
edge-cached; the origin check covers the proxy path, and the proxy
refuses to inject on non-idempotent requests lacking a same-origin
signal.

## The ladder

Rungs are sequential with one executor and no overlap. Magnitude bands
are relative and carry no time units: M is one pass under one plan at
roughly the repo's recent pass ceilings, S is well under that, L
splits into multiple passes at its own planning. Each rung gets its
own spec and plan when its turn comes, and each plan header carries a
token ceiling and checkpoint interval. Evidence-track rungs are marked
[E].

1. **Foundation** (M). The repo pivot and salvage freeze; the
   greenfield `internal/config` with the universal wall, enable-flag
   discipline, and threshold registry; the `cmd/dubplate-server`
   skeleton with its re-derived subcommand surface; `check` and
   `doctor` scaffolds; the gate rewired to cover the greenfield tree
   and exclude `_salvage/`; ROADMAP restated for the pivot.
2. **Corpus and harness** (L) [E]. The quadruple corpus (input, frozen
   MusicBrainz responses, beets config, expected verdict), the
   frozen-verdict fast gate with no Python in it, the version-bump
   differential job against pinned beets and a local stub, and the
   capture tool. The registered MusicBrainz application identity and
   User-Agent land here, and the quarterly re-capture chore is
   documented here. Strata per the research report: roughly 150 to 300
   cases, about 40% structural, 50% real catalog, 10% adversarial,
   with the adversarial stratum sized against tier one.
3. **Go matcher** (L) [E]. The distance and penalty model, album-level
   matching, stable candidate ordering, numeric distance agreement
   within the epsilon rung 2's plan states, beets' full tunable
   surface (thresholds, distance weights, preferred, ignored,
   required), and the shared MusicBrainz rate limiter with per-caller
   budgets in which a user-present probe outranks a batch import.
   Behind the harness until the bar holds.
4. **Shadow** (S to build, continuous to launch) [E]. The shadow
   runner deployed to the box under the constraints above, the
   divergence log, and the adjudication artifact. From here the
   evidence clock runs unattended while the spine builds.
5. **Identity and auth** (M; the plan may split it). `internal/auth`
   built to the birth requirements: tokens, sessions with epoch and
   absolute cap, durable nonces, the login and consumption pages
   (Go-templated, no JavaScript, shared design language), SMTP send,
   rate limiting, the verify-identity seam with provenance, the
   forward-auth verify endpoint, and a minimal serve shell hosting it
   under the one owned prefix. The `web-auth-security-reviewer` gate
   runs inside the rung.
6. **Ops seam and store** (M). The ops vocabulary: one Go interface
   plus JSON-serializable request and response types, with the job
   model and explicit cancellation from the first version. The SQLite
   store on the modernc driver, migrations that run forward on open,
   snapshot backup and restore against rclone remotes, and the
   contract-test suite skeleton that will hold both implementations to
   parity.
7. **Intake** (M). tus transport, the gate policy, roster loading with
   the birth-requirement validation, journey records, the reject copy
   contract in payloads, and the uploader reference web client built
   from the salvaged components under fresh review. The `web/` build
   joins the repo gate here. The album is the verdict unit: one
   failing track rejects the batch, and no partial album is ever
   admitted.
8. **Tag I/O** (M). Pure-Go FLAC tag writing for the server scope,
   plus the wazero-taglib spike whose recorded decision the post-launch
   local organizer will consume. Writing tags is the pipeline's primary
   side effect, so this rung precedes the pipeline rather than folding
   into it.
9. **Pipeline** (L). Admission, tagging, layout, and collection
   routing on the engine and store; the notifications and monitoring
   seams as optional integrations, built immediately ahead of the
   pipeline that consumes them; the repair verbs (query, retag against
   a named MusicBrainz release, move), which are launch-bar admin
   surface because the new server has no beets toolkit to lean on; and
   the concurrency model decision (per-album lease or explicit queue,
   with the constraint that interactive work must not starve timer
   imports), settled in this rung's spec. The honest terminal state
   "imported, no player configured" lands here as journey vocabulary.
10. **Built-ins and lifecycle hooks** (M). The built-in set re-derived
    against the plugin survey and the estate's real config, under the
    fixed policies: scrapers refused, fingerprinting and ReplayGain
    probe-backed. Lifecycle hooks at the named pipeline events, JSON on
    stdin in the ops wire vocabulary, failure never failing an import.
11. **Streamer driver and SSO** (M). The driver seam (scan trigger,
    scan status, album deep link, capability reporting) with the
    Navidrome reference driver owning the bypass path set, the trusted
    header, the externalized-auth config names verified against the
    pinned Navidrome version, logout wiring, and provisioning
    choreography with the first-admin `doctor` preflight. The ingress
    reverse proxy built to the proxy-hygiene requirements, behind the
    SSO enable flag, with the availability coupling recorded at the
    seam.
12. **Decision and review queues** (M). Persisted decision records
    serializing candidates, distances, and penalty breakdowns; the
    queue verbs (`list pending`, `answer`, `defer`) with a minimal CLI
    answering path; the custody-policy enum (`strict` default,
    `owner-queue`, `contributor-picks`) and the classical-appeal
    action that rides every setting; and the admin review queue, the
    distinct low-volume side door for timer-path material.
13. **Versioned API and admin spine** (M). The contract document,
    `/api/v1` under the owned prefix, the additive-within-version
    policy, the authorization model (AdminEmail, header-only admin
    verbs enforced through provenance, service tokens for unattended
    machine verbs), and non-browser token acquisition on the session
    vocabulary, so native clients and the post-launch TUI have a
    credential path.
14. **Integration** (M). The tracks meet: the matcher wires into the
    pipeline, the full contract suite runs, the conformance gate and
    shadow record are evaluated against the launch bar, and every page
    on the user path is rendered and read before any attended ask. The
    domain reviewer fan-out for the whole surface runs here.
15. **Launch cutover** (M). The migrator against the organizer spec's
    four failure modes (BLOB paths never round-tripped through Go
    strings, NULL singleton album IDs, ledger reconciliation, the
    no-move dry-run against the live library); the beets database
    archived to R2; the box swap under a written, exercised rollback
    (re-PUT the prior tunnel ingress, disable externalized auth,
    restore the beets snapshot); systemd, deploy.sh, compose, and
    cloud-init rebuilt for the new server per vps-conventions; secrets
    minted through the age-store flow; the ingress-before-ExtAuth
    ordering honored; the interim estate and `_salvage/` retired; the
    committed universality breach (`UIWelcomeMessage`) resolved
    through configured `SiteName`. One attended gate: Geoff verifies
    login, upload, and player on his devices in a single sitting. On
    his yes, the family launch.

## What this spec supersedes

Against the organizer spec, which otherwise governs:

- The build-order ladder (its rungs 1 through 11) is replaced by the
  ladder above. Its rung 10 (two-binary split, local TUI) and the
  remote-TUI half of rung 11 move past launch, unscheduled here.
- Staged adoption through the `internal/beets` subprocess seam is
  superseded. The greenfield server never embeds beets; the fallback
  for a failed matcher is the interim estate continuing to run.
- The `beet` passthrough and its retirement condition are moot on the
  new server. The repair verbs are launch-bar surface.
- The shadow window is superseded by the whole-build shadow, with the
  organizer floor kept as the launch-gate minimum.
- The header's preconditions (uploader ships, prod-polish completes)
  are superseded by the pivot.

Against the login spec: the build-on-old-serve framing and its
one-compact-pass build order are superseded. Its design content
survives as source material for rungs 5 and 11 under the birth
requirements, and its three filed amendments plus the missing fourth
(identity provenance in the verify seam) are folded into this spec
rather than filed separately.

## Box safety during the build

The box is production for Geoff's own listening. The guardian timers
and R2 backups protect the library and beets database throughout.
Rungs 1 through 3 and 5 through 13 touch the box not at all; rung 4
deploys the additive shadow runner; rung 15 is the one rung that
changes what the box serves, and it carries the quiet-box check, the
`--check`-first deploy discipline, and the exercised rollback. Every
executor entering this repo checks for live executors first.

## Out of scope

Everything the organizer spec excludes: the web admin page,
multi-tenancy, in-process plugins, scrapers, two-way sync and the
mirror, copy localization, native mobile clients. The local
`dubplate` binary, the TUI in both modes, the two-binary split, and
publish are post-launch work, outside this initiative. The packaging
initiative remains sequenced after this one.

## Process

Each rung opens with its own spec where the ladder marks decisions
unsettled, and a plan via writing-plans; L rungs split at their own
planning. Execution runs the implementer, diff-reviewer, gate chain
per task, with the domain fan-out at pass ends and the workflow mode
at six or more tasks. Both budgets are scored at every pass close.
The one attended gate of the initiative is rung 15's; everything else
that needs Geoff is batched questions at checkpoints.
