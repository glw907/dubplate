# Greenfield build-spec adversarial review (universality and adoption), 2026-09-02

Spec: `docs/superpowers/specs/2026-09-02-greenfield-server-build.md`

Read in full: the spec under review, the ratified rulings
(`research/2026-09-02-greenfield-server-rulings.md`), the login spec
(`specs/2026-09-02-magic-link-login-design.md`) and both its adversarial reviews
(`research/2026-09-02-login-spec-reviews/security.md`, `universality.md`), and the
organizer spec (`specs/2026-09-01-dubplate-organizer-design.md`). Grounded against
`CLAUDE.md`, `docs/STATUS.md`, `env/dubplate.env.template`, `config/` (`navidrome.toml`,
`contributors.yaml`, `filebrowser.yaml`), `scripts/`, `tests/`, and the current
`internal/` package list.

**The fold is largely faithful.** All six security blockers and all three universality
criticals survive at full strength, several with better wording than the originals: the
strip requirement gained "on ingress, before routing, for every path", the trusted source
generalized from `172.30.0.1/32` to "serve's single observed address", and the `/api/*`
collision became a stated architectural rule ("One owned prefix", spec:244). Of the
security review's thirteen majors, eleven land verbatim in substance. The findings below
are what the fold dropped, what it reworded into a weaker claim, and what the ladder's
universality claim does not survive contact with.

## Ranked findings

**[Critical] The forward-auth topology — the spec's own answer for the universal adopter — inherits no trust boundary at all.**
Spec:236-242 states the boundary in serve's terms: "Serve strips every inbound
identity-header spelling on ingress, before routing, for every path... The streamer's
trusted source is pinned to serve's single observed address... and no other network path
may reach the streamer." Every clause of that guarantee is conditional on serve being the
streamer's ingress. Spec:299-300 then opens the other topology: "Adopters with their own
proxy get a forward-auth verify endpoint rather than being forced through serve's proxy
topology." In that deployment serve is not in the streamer's request path — the adopter's
nginx/Traefik/Caddy is — so nothing in this spec strips a client-supplied `Remote-User`
before it reaches the streamer, and the "single observed address" the streamer trusts is
the adopter's proxy, which forwards whatever the internet sent it unless the adopter
independently knows to strip. The security review verdicted the equivalent gap as "remote
unauthenticated account takeover" (security.md:12, 116). The fold merged the security
review's fix and the universality review's fix and never re-derived the first for the
topology the second introduces. The requirement the spec owes: the forward-auth path
declares its own boundary (the proxy strips the driver's header set before the verify
subrequest and before the streamer, the streamer trusts only that proxy, and `doctor`
proves a spoofed header from a non-trusted source is rejected), and the driver publishes
its strip set to the adopter rather than only to serve's ingress.

**[High] `__Host-` is adopted from the security review, which silently forecloses the split-host deployment the universality review filed as [High] — and the trade is nowhere recorded.**
Spec:257 requires an "HMAC-signed `__Host-` cookie". The `__Host-` prefix forbids a
`Domain` attribute by definition, so the session is host-only. universality.md:17 raised
exactly this deployment: "Split uploader and streamer onto separate hostnames — a natural
adopter choice — and nothing works: a host-scoped cookie is not sent to the sibling
host... Either state same-origin as a hard precondition of the SSO capability, or add
cookie-domain config plus a redirect-target allowlist." The spec takes neither branch. It
never mentions hostname scoping, cookie domain, or the split-host case. Worse, the escape
hatch it does provide does not escape: a forward-auth adopter on a split host has the same
problem, because the proxy's verify subrequest for a request to the streamer host carries
no cookie set on serve's host. The spec's own universality test — "The design makes sense
with no Cloudflare account" (spec:92) — is satisfied while the more common adopter
topology is quietly unbuildable. One sentence fixes it: same-origin is a stated
precondition of the SSO capability, and split-host adopters get a documented alternative
(or the config).

**[High] The `/rest/*` bypass sits outside every control the birth requirements build, and "an offboarding step that actually offboards" overclaims because of it.**
Spec:261-262: "a per-email valid-after epoch in the state directory gives per-person
revocation and an offboarding step that actually offboards." The epoch offboards a member
from dubplate. It does not offboard them from the streamer, and the security review named
precisely why (security.md:49): "removing someone from `config/contributors.yaml` kills
their dubplate access... but does **not** kill their streamer access, because
spec:128-131 has every member self-set a Navidrome app password and spec:46-50 bypasses
`/rest/*` entirely... add... an explicit deprovision runbook step." The login spec then
puts that step out of scope by name (login:199-202, "Navidrome account lifecycle beyond
auto-creation... stay in the streamer's own admin UI"). The greenfield spec neither folds
the deprovision requirement nor restates the exclusion, and rung 11's driver-seam verb
list (spec:391-395: scan trigger, scan status, album deep link, capability reporting,
bypass paths, trusted header, ExtAuth config names, logout wiring, provisioning
choreography) has no deprovisioning counterpart to the provisioning it does fund. The same
structural fact leaves security.md:104's residual unaddressed: `/rest/*` is now the one
anonymous, unrate-limited surface the design creates, and the spec places no abuse
requirement on it while requiring serve to be its only reachable path. Either the seam
gains a deprovision capability, or the claim shrinks to what it is: per-person revocation
of the dubplate session, with streamer deprovisioning an owner runbook step.

**[High] The config wall is asserted as a rule and never derived, so rung 1 has no way to know what it must produce.**
Spec:96-98: "Integrations get explicit enable flags, and a flag makes its correlated config
required-when-enabled. The required-no-defaults wall covers only what the server owns."
That is the organizer's rule restated (organizer:84-86), and it is where the adoption
claim actually lives. Nothing in the ladder applies it. Rung 1 (spec:322-327) says only
"the greenfield `internal/config` with the universal wall, enable-flag discipline, and
threshold registry" — the wall's *contents* are never derived, and the estate's current
required set is what an implementer will reach for. `env/dubplate.env.template` documents
that set: "internal/config.Load requires every DUBPLATE_* name below except the
DUBPLATE_PING_* block", which today makes five `DUBPLATE_R2_*` values, five SMTP values,
and three Navidrome values required with no default. Three consequences the spec must
resolve rather than leave to rung 1's implementer:
- **R2 is Cloudflare.** The spec's own fourth test is "The design makes sense with no
  Cloudflare account" (spec:92) while the organizer places Storage among the five
  integrations the server "integrates with, and never owns" (organizer:82-104) and rung 6
  lists "snapshot backup and restore against rclone remotes" (spec:359-360) with no enable
  flag. A server that will not start without an R2 bucket fails the test the spec opens
  with.
- **Navidrome's variables.** organizer:649-652 states they "stay required with no default"
  until the seam extraction is scheduled. This spec schedules it (rung 11) without saying
  so in the supersession list.
- **Login's own status.** Spec:274 conditions `DUBPLATE_PUBLIC_URL` on "when login is
  enabled", which implies an enable flag for a surface ruling 7 places in server purview
  ("Login machinery is server purview", rulings:39). Owned surfaces take the
  required-no-defaults wall; integrations take flags. Login is being given both shapes at
  once.

**[High] Beets path-format parity is unfunded, and the launch bar's fourth part is therefore discovered inside the one irreversible rung.**
The launch gate has four parts (spec:44-59). Parts 2 and 3 have funded evidence machinery
(rungs 2, 3, 4). Part 4 — "A dry-run diff proves the new pipeline's computed paths equal
the on-disk layout of the live library before anything writes" — has none. Rung 9
(spec:375-376) delivers "Admission, tagging, layout, and collection routing" with no
parity requirement, and rung 2's corpus is a *matcher* corpus by construction (the
quadruple is "input, frozen MusicBrainz responses, beets config, expected verdict",
spec:328-330), which carries no path assertion at all. Meanwhile the layout the assertion
must reproduce is a beets DSL: `CLAUDE.md` records "Collections route via a beets flexible
attribute with query-prefixed path formats", and the organizer warns "A one-character
difference from `config/beets-config.yaml`'s query-prefixed path formats re-lays-out the
entire library on the first post-swap run" (organizer:452-454). Compounding it, the
migrator moved. The organizer built it at rung 5 and kept the swap "deliberately thin,
because rungs 1 through 6 built and proved everything it needs" (organizer:610-611); this
spec folds the migrator, the four failure modes, and the no-move dry-run into rung 15
(spec:418-421), the rung that also performs the irreversible box swap. The organizer's
stated reason for the ordering ("Rungs 3 through 6 sit in front of it for that reason",
organizer:546-547) is inverted without being listed as superseded. This is also the
adopter's problem, not only the estate's: any adopter migrating from beets faces the same
layout parity question, which makes it engine work, not cutover work.

**[High] The conformance corpus's primary oracle is one estate's beets config, and the launch bar is measured against it.**
Launch bar part 2 (spec:50-53) requires the two-tier success bar to hold "over the
conformance corpus". The organizer defines whose corpus that is: "The primary oracle
config is the estate's `config/beets-config.yaml`, because the estate is what the swap
risks. A stock-config stratum is secondary, and it is what says anything about a stranger
adopter" (organizer:478-481). Rung 2 (spec:328-336) restates the strata and the quadruple
and drops the primary/secondary distinction entirely. Under the greenfield ruling the
framing has to change: this initiative is no longer a swap that risks one estate, it is a
universal product whose central ruling is "No machinery specific to one estate, anywhere"
(spec:81). As written the existential-risk gate is passed by proving conformance under one
family's tunables, while the stratum that speaks to an adopter gates nothing. Say what the
stock-config stratum must clear for launch, or say plainly that adopter-config conformance
is post-launch work and accept that the matcher bar is estate-scoped evidence.

**[Medium] Two salvage artifacts are needed as data and strings, and the reuse stance only supplies a mechanism for code.**
Spec:216-221 admits "`atomicfile`, the reject copy contract, the gate policy, the ledger
format" as adoption candidates, then defines adoption as "copying into the greenfield tree
under the exemplary-Go and idiomatic-Svelte bars; nothing is imported from `_salvage/`".
Both bars are code bars, and two of the four items are not code problems.
- *The reject copy contract.* `CLAUDE.md`: "Reject wording is contract
  (`internal/reject` / `friendly_reason`): the uploader page and emails share it verbatim."
  The strings have a ratified authority — `research/passb-copy-final.md`, self-described as
  "the string authority", content-review gate PASS 2026-09-01 — which appears in neither
  the spec's "Sources folded" list (spec:9-16) nor rung 7's deliverables (spec:363-369).
  Nothing says whether the strings are adopted verbatim or re-derived, and no register
  review owns them at any rung. A Go bar applied to a copy deck will silently reword a
  contract.
- *The ledger format.* The greenfield tree (spec:188-206) has no ledger package, yet rung
  15 requires "ledger reconciliation" (spec:420) against the organizer's third migrator
  failure mode ("the ledger must be reconciled or every contributor's journey history
  breaks its links at the swap", organizer:446-449). That is a read of the *legacy*
  ledger's on-disk format, from code that may import nothing from `_salvage/` and whose
  ledger package does not exist. The format is a frozen data contract, not an adoption
  candidate, and it needs a rung that captures it while the salvage tree is still legible.

**[Medium] The trusted-proxy pair reproduces the exact half-configured footgun the wall rule exists to prevent.**
Spec:278-281: "Limits key on the real client IP, read from a configured trusted-proxy
header honored only when the peer is the configured proxy, with `RemoteAddr` fallback for
proxyless adopters." Two correlated values (the header name and the trusted peer) plus a
silent fallback. An adopter who configures the header and not the peer, or the peer
incorrectly, gets the fallback: every request keys on the proxy's address, which is
security.md:24's original blocker restored ("five link requests per hour for the entire
family"), and nothing announces it. This is the same shape as the Access pair the
universality review filed (universality.md:43-44) and this spec correctly fixed
(spec:295-298). Apply the same treatment: the trusted-proxy header and peer are one
required-when-enabled pair under a proxy-mode flag, the proxyless mode is chosen rather
than fallen back into, and `doctor` reports which mode is live.

**[Medium] Roster validation is specified for `name` while the identity header value is now `streamer_username`, and the charset floor was softened.**
Spec:287-293: "The roster validates at load: unique names, addresses parsed with
`net/mail.ParseAddress`, no control characters, bounded length. The streamer principal is
split from the inbox directory name: an optional per-contributor `streamer_username`..."
The validation sentence governs names and addresses; the value that now goes on the wire
is `streamer_username`, which the sentence does not reach. security.md:67 asked for the
constraint on the header value specifically ("Require at roster load: unique names,
ASCII-printable, no control characters, bounded length — and say plainly that the header
value must be a stable unique identifier, not a display name"). The fold kept "no control
characters" and dropped "ASCII-printable". Go's header-value validation admits high bytes,
so a non-ASCII `streamer_username` passes serve and is mangled or rejected downstream by
whatever the adopter's streamer and proxy do with it, which is a per-member hard failure
with no diagnostic. State the validation against the principal, not the display name, and
keep the ASCII floor.

**[Medium] The post-login destination is absent from the spec entirely, and it is load-bearing in three separate review findings.**
The word "redirect" does not appear in the spec. The login spec made the destination
carry driver semantics ("the post-login redirect deliberately lands on the player root so
auto-creation fires", login:121-123), and both reviews built on it: universality.md:50
called that "a core redirect target load-bearing for one driver's auto-provisioning; a
driver without auto-create needs an explicit provisioning verb instead", and
security.md:77 spelled out the redirect-validation bypass class. Rung 11's "provisioning
choreography" (spec:395) is the only plausible hook and it does not say the destination is
driver-declared. Three configurations need three different answers and the spec gives
none: with SSO the player root, with a streamer but no SSO the dubplate root, and with no
streamer at all the honest terminal the spec does fund at rung 9 ("imported, no player
configured", spec:384). Make the post-login destination a driver capability with a core
default, and carry the redirect-target validation rule with it.

**[Medium] Three config values will describe one streamer's address with three different semantics, and the spec resolves none of them.**
universality.md:41 filed this against the login spec and it is unaddressed here. The
estate has `DUBPLATE_NAVIDROME_URL` (admin API) and `DUBPLATE_PUBLIC_NAVIDROME_URL`
(documented in `env/dubplate.env.template` as "the externally reachable Navidrome address
the journey UI links into"); the login spec adds a driver "proxy target" (login:151-155).
The greenfield spec names the driver capability that consumes the public one — "album deep
link" (spec:392) — and never says where its base URL comes from. The answer differs per
configuration: same-origin under the SSO proxy, a separate origin without SSO, absent with
no streamer. Left unresolved this is a site-specific URL that will be reintroduced into
the greenfield config at rung 11, in a build whose first universality test is "No hostname,
path, email, or personal name outside `internal/config`" (spec:84).

**[Medium] Rung 4 puts a second binary on a live box whose ops files serve the frozen one, and the spec defers only the observation mechanism.**
Spec:166-167 defers "The shadow mechanism (how the runner observes the import timer's
material)" to the rung's plan, which is reasonable. The deployment half is not deferred,
it is unmentioned. From rung 1 the pivot freezes `cmd/`, `internal/`, `web/` into
`_salvage/` (spec:172-176) while `scripts/`, `systemd/`, `compose.yaml`, `config/`, and
`env/` "stay live and gated" (spec:177-180) — and those files describe the frozen binary.
`env/dubplate.env.template`'s own header reads "internal/config.Load requires every
DUBPLATE_* name below", a sentence that after rung 1 refers to `_salvage/internal/config`.
Rung 4 then deploys a greenfield binary with a *different* config wall (rung 1's) to the
same box, which means deploy.sh renders two walls, systemd carries a second unit under
vps-conventions, and two binaries coexist until rung 15 rebuilds everything. None of that
appears in "Box safety during the build" (spec:459-465), which says only "rung 4 deploys
the additive shadow runner". Related and smaller: the shadow runner has no leaf in the tree
(spec:188-206) and no stated fate at launch, though the spec is explicit that the
version-bump differential job survives as a standing tool (spec:70-75). An adopter
migrating off beets wants the shadow capability as much as the estate does.

**[Medium] The supersession list is incomplete: at least three organizer statements and most of the login spec's design now conflict silently.**
Spec:435-455 lists five organizer deltas and two login-spec deltas. Unlisted, and now
contradicted:
- organizer:99, "Cloudflare Access is the reference implementation" of the verify-identity
  seam. The spec flips it (spec:265-267) under the login spec's amendment 1, and states
  the amendments are "folded into this spec rather than filed separately" (spec:453-455),
  which means the organizer document is never corrected and keeps reading the old way while
  "remain[ing] the governing architecture" (spec:29).
- organizer:649-652, "One deliverable stays unassigned: the streamer-driver seam
  extraction, with its config-wall change and its journey vocabulary amendment... Until one
  arrives, the Navidrome environment variables stay required with no default." This spec
  assigns all three (rungs 11, 1, 9).
- organizer:101-103, "Generalizing today's Access coupling into this seam, non-browser
  token acquisition included, is a deliverable of the admin-spine rung." The spec splits it
  across rung 5 (the seam) and rung 13 (token acquisition).
On the login spec, spec:450-452 supersedes only "the build-on-old-serve framing and its
one-compact-pass build order" and keeps the rest as "source material for rungs 5 and 11".
But the birth requirements overturn its *design*, not its framing: memory-only nonces
(login:89-93), 90-day rolling sessions revoked only by rotation (login:98-106), the
stateless cookie, routing rule 2's `/api/*` (login:51-52), the `/24` trusted source
(login:64-66), enumeration kindness as fixed behavior, and roster `name` as the header
value (login:118-120) are all now wrong. A rung-5 implementer told to treat that document
as source material will read superseded design as design. Mark the superseded statements
in the login spec, or state that the birth requirements win on every conflict.

**[Medium] The estate ops files are kept "live and gated" without saying whether they are the estate's record or an adopter's kit, and rung 15 rebuilds them rather than universalizing them.**
Spec:177-180 keeps `scripts/`, `systemd/`, `cloud-init/`, `compose.yaml`, `config/`, `env/`
live "because they run the interim box today and the new box after cutover". Both stated
purposes are this estate's. `CLAUDE.md` sets a different bar for the same files: "Ease of
setup by a stranger is a primary product goal: no site-specific value, path, or identity
default outside `internal/config`" — and today those files carry `ND_SPOTIFY_ID`, a
`musicbox` network, `config/cloudflared-tunnel-config.json`, and `UIWelcomeMessage =
"Wright Family Music Server"` (`config/navidrome.toml:13`). Rung 15's wording is "systemd,
deploy.sh, compose, and cloud-init rebuilt for the new server per vps-conventions"
(spec:422-424), which is a rebuild for one box, performed in the rung with the least
headroom, and it will bake estate shape into the artifacts a stranger clones. The spec's
own out-of-scope section offers the honest resolution — "The packaging initiative remains
sequenced after this one" (spec:474) — so say it: the ops tree is the estate's operational
record through launch, adopter-facing deployment artifacts are packaging-initiative work,
and rung 15 rebuilds without claiming universality it is not funded to deliver. Silence
here is what lets a rung-15 implementer believe the launch satisfies the stranger goal.

**[Low] Rung 15 is banded M while carrying nine deliverables, the only irreversible step, and the only attended gate.**
Spec:418-431 lists the migrator against four failure modes, the R2 archive, the box swap,
a written and exercised rollback, systemd, deploy.sh, compose, and cloud-init rebuilds,
secrets minting, the ingress ordering, two retirements, the `UIWelcomeMessage` fix, and
Geoff's verification. The band definition is "M is one pass under one plan at roughly the
repo's recent pass ceilings" (spec:316-317), which puts this alongside rung 1's pivot. The
workstation pass-sizing rule asks for a deliverable count at dispatch and a plain statement
when it passes roughly four. The organizer's equivalent rung was banded S and was
deliberately thin *because* everything it needed was pre-proven (organizer:607-611); this
one is fatter and pre-proves less. Band it L, or move the migrator and the ops rebuild
ahead of the swap, which the path-format finding above wants anyway.

**[Low] The organizer's blanket CSRF requirement is not restated, and the spec names origin checks in only two places.**
The word "CSRF" does not appear. organizer:128 requires "CSRF defenses scoped to
cookie-carrying browsers", and universality.md:25 filed its absence as [High] in the pass
that introduces the estate's first first-party cookie. The spec covers the consumption POST
(spec:254, "POST consumes, origin-checked") and the proxy path (spec:308-310). The login
POST and logout are not covered, and an unprotected login POST is a mail-bomb trigger
against any roster address. security.md:36 also noted that the existing `requireSameOrigin`
sits inside the `/upload` mount and covers nothing new — moot for greenfield code, but the
positive requirement it implies is the one worth carrying: every cookie-carrying
state-changing surface takes the origin check, stated once at the auth rung rather than
enumerated per endpoint.

**[Low] The ops keep-list names an artifact class that no longer exists and keeps a retired integration's config.**
Spec:178-180 lists "`scripts/`, `tests/*.bats`, `systemd/`, `cloud-init/`, `compose.yaml`,
`config/`, `env/`" as the live-and-gated set. `tests/` is empty — there are no `.bats`
files in the repo — so the gate's shell-behavior pin that `CLAUDE.md` describes ("behavior
is pinned by `tests/*.bats`") is naming nothing, and the pivot list inherits the claim
uninspected. `config/` is kept wholesale and still contains `filebrowser.yaml` for the
integration the wind-down retired on 2026-09-02 (rulings:65-74). The keep-list wants one
audit pass at rung 1, not a directory enumeration.

**[Low] The sources list omits the initiative's largest matcher input.**
Spec:9-16 lists the folded sources: the organizer spec, the login spec, both login reviews,
the Navidrome research, the conformance corpus report, and the uncertain-match UX report.
`research/2026-09-02-beets-matcher-mechanics.md` is not among them — 649 lines of
source-level read against pinned beets `v2.13.1`, written the same morning, and the
substrate for rung 3's entire deliverable list ("The distance and penalty model,
album-level matching, stable candidate ordering... beets' full tunable surface",
spec:337-343). Rung 3 is the rung the spec itself calls the initiative's existential risk
(spec:64-66). Its research should be named where the ladder points an implementer at it.

## The three I would fix first

1. **The forward-auth trust boundary (Critical).** The spec opens a second topology in one
   sentence and carries no boundary requirement for it, while the boundary it does carry is
   written entirely in serve's terms. This is a paragraph now and an account-takeover class
   later, and it is the finding where the universality fix and the security fix were merged
   without either being re-derived against the other.
2. **Derive the config wall (High).** Rung 1 writes the wall once and every rung after it
   inherits the result. Right now the ladder tells that implementer the rule and not the
   answer, and the answer sitting in `env/dubplate.env.template` fails the spec's own
   fourth universality test at the first required R2 value. Naming the owned set, the
   integration set, and login's own status costs a table now; retrofitting it after twelve
   rungs build against the wrong wall is the most expensive kind of universality debt.
3. **Fund layout parity and move the migrator off the irreversible rung (High).** One of
   the four launch-gate conditions has no evidence machinery anywhere in the ladder, and it
   is evaluated inside the single rung that writes to the live library. The matcher's risk
   was priced correctly and scheduled first; the layout's identical risk was not priced at
   all. Give path-format parity a corpus and a rung ahead of rung 15, which restores the
   organizer's stated reason for putting the proving rungs in front of the swap.

**Verdict: the fold is faithful and the ladder is sound, and the spec is not ready to build until the forward-auth topology gets a trust boundary, the config wall is actually derived, and layout parity is funded — three gaps that all sit where the universality claim meets the estate the build lives inside.**
