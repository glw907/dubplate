# Rung 5: identity and auth

Status: **revision 2.1, product forks ruled by Geoff 2026-09-08.** Pass
5a closed 2026-09-08 (`internal/serve`, merged at `307001b`) and its
close applies the errata dated 2026-09-08 in "Errata the close
applies", below: R23's ceiling, R24's review shape, the session-cookie
pass assignment, the hop-by-hop scope, the trailer-coverage claim, the
`X-Forwarded-*` family, rule 1's mechanism and rule 3's shape, the read
timeout, and the template-classification test's home.

Revision 2.1 folds Geoff's rulings on the six product forks. Fork 1
reverses its recommendation: adding a group member is a web-interface
action, built at rung 5. The roster becomes a file-backed store behind
one loader interface with an atomic write path, an admin flag per
member, a `dubplate-server member` verb group that bootstraps the first
admin on a fresh install, and an admin-only add-member page beside the
login page. Forks 2, 3, 4, and 6 are ruled as recommended: SMTP
submission alone with five configured values, every lifetime and limit
as a registered threshold, the kind unknown-address answer with a flag
for the generic one, and no typed-code fallback at rung 5. Fork 5's "no
admin bootstrap" is retired, because the CLI verb is the bootstrap. R1,
R23, R24, and R37 are amended to match fork 1 and carry Geoff's date;
the other thirty-four rulings are ruled as this revision applied them.
The pass split still runs four passes. Pass 5c grows by two deliverables
to fifteen, and its ceiling moves from 4M to 4.5M.

Revision 2 folds two fresh-context adversarial reviews, one through the
web-authentication-security lens and one through the
spec-integrity-and-scope lens, merged and re-verified at
`docs/superpowers/research/2026-09-08-rung-5-spec-review.md`. Three
blockers, twelve majors, and fourteen minors are folded; the merged
review's "Verified sound" list is the do-not-re-litigate record for
every rung 5 pass plan. The three blockers were a rate-limit key that no
clause made unspoofable, a link-request limiter with no per-client-IP
key at all, and an identity-source selector that the landed registry
requires nowhere and `doctor` reports green while it is unset. Rulings
R27 through R38 are new in this revision, R6, R11, R14, and R15 are
amended in place, and the pass split is unchanged. Two claims revision 1
made are withdrawn as false and replaced with what is true: "the wall
proves presence" (R29) and "tokens never appear in a URL that gets
logged" (R32).

Four owner rulings already given are recorded below as ruled: the build
carries no calendar schedule and no launch date; rung 5's login and
consumption pages are SvelteKit static pages served by the Go binary (0c
R1.6); the `web/` gate leg rejoins the repo gate at rung 5; and the
string threshold kind is rung 1b's work, landed. Six questions changed
what an adopter or a member sees, configures, or is asked for,
and Geoff ruled all six on 2026-09-08; the fork section carries each
ruling beside the recommendation it accepted or reversed. Every further
ruling this spec raises is now ruled as this revision applied it, and
the ruling section records who ruled each.

Rung 5 is one of the three rungs that keep the full fresh-context
adversarial cycle (CLAUDE.md, the build-process section). The review
shape is stated in the process section, and the threat model the
reviewers grade against is its own section.

Written against the ratified build spec
(`docs/superpowers/specs/2026-09-02-greenfield-server-build.md`,
revision 2.1, BS) and its auth birth requirements (BS:393-523), the
magic-link login design
(`docs/superpowers/specs/2026-09-02-magic-link-login-design.md`, MD,
superseded as design and read for intent), the parallel-build approach
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`,
revision 2, PA), the reverse-edge completion
(`docs/superpowers/research/2026-09-04-reverse-edge-completion.md`, RE),
the design-language standard
(`docs/superpowers/specs/2026-09-07-design-language-standard.md`,
revision 2, DL), the 2.0 streamer design
(`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, S20), the
two login-spec adversarial reviews
(`docs/superpowers/research/2026-09-02-login-spec-reviews/security.md`,
SEC, and `universality.md`, UNI), the rung 5 spec inventory
(`docs/superpowers/research/2026-09-08-rung-5-spec-inventory.md`, INV),
the rung 5 prior-art study
(`docs/superpowers/research/2026-09-08-rung-5-auth-prior-art.md`, PRA),
and the merged revision 1 review
(`docs/superpowers/research/2026-09-08-rung-5-spec-review.md`, REV).

Tree citations are read at `be274346add7abc339bd6ca0cf4c3b3071dd1fdf`.
Citations carried from INV and PRA were read at
`8ff65bd99f982bda940db19420ee80f834fcf9d1`; the staleness pre-task
re-checks the ones this rung leans on.

Nothing here amends the build spec, the launch bar, or the ladder. Two
clauses of ladder entry 5 are already superseded by later ratified
rulings, and the rung's close writes both as dated errata rather than
silent edits. The errata section carries their wording.

---

## What rung 5 builds

- **`internal/auth`**, the magic-link identity source: token minting and
  verification, the session cookie, durable consumed nonces, the
  per-person revocation epoch, the roster store with its
  birth-requirement validation and its atomic write path, the admin flag
  each member carries, the rate limiters, the origin check, and the
  verify-identity seam with assertion-channel provenance.
- **`internal/identity`**, a dependency-free leaf holding the verified
  identity type and its assertion channel, so that `internal/ops` can
  carry a verified identity at rung 6 without importing a server
  package. RE cell (5,6) requires this spec to name that package.
- **`internal/serve`**, the shell: the root composition order, the
  ingress header strip, path normalization, prefix routing under the one
  owned prefix, the empty fallthrough slot, the declared extension point
  every later consumer adds a file at, the `http.Server` construction
  with its deadline policy, the embedded-asset path, and the
  served-document rewrite DL R4.5 specifies.
- **`internal/notify`**, a general mail send seam with one SMTP
  transport, adopted from `_salvage/internal/notify` under fresh review
  and restructured so rung 7's reject mail and rung 9's notifications
  are callers rather than special cases.
- **`web/`**, born as a SvelteKit static app under CLAUDE.md's web rule:
  the login page and its own string deck, the link-consumption
  interstitial, the admin-only add-member page fork 1 rules, the theme
  variable binding DL section 2 defines, one composed theme authored as
  data, and the embedded woff2 faces with their licenses and manifest.
- **The `member` verb group**, `dubplate-server member add --admin
  <email>` and `member list`, which write and read the roster store from
  the command line. The verb is how a fresh install gets its first
  admin, and it is the only path that grants the admin flag to an
  account that has never signed in (fork 1).
- **The forward-auth verify endpoint**, off by default, with its own
  trust-boundary documentation stating that the boundary is the
  adopter's proxy.
- **Active `doctor` checks**: the boundary proof B2 names and the
  client-IP mode report B8 names, as rows whose status comes from a
  probe.
- **Config**, through rung 1b's registry: the two reserved classes
  removed from `internal/config/keys.go` and declared in rung 5's own
  file in the same commit, the pending identity-source wall row retired
  into a real selector key, the `internal/config/config.go` edit that
  makes `Load` require and resolve it, the thresholds with their
  defaults and their two correlated validations, and the two integration
  classes B8 and B2 need.
- **Deployment**: the greenfield serve unit, its `ReadWritePaths` grant
  over both the auth leaf and the roster leaf fork 1 makes writable, the
  `deploy.sh` pre-create of each, the commented pending rows in the
  environment template, and the generalized state-directory agreement
  check.
- **The gate's `web/` leg**, rejoining at this rung with DL's ten gates
  and the SvelteKit server-file check, plus the three live copies of
  "the web leg rejoins at rung 7" corrected.

---

## Owner rulings, recorded

**O1. No calendar schedule.** The build carries no launch date. Nothing
in this rung expresses a deadline, and its thresholds are durations
measured from an event rather than dates. (Geoff, recorded as the rung 4
erratum.)

**O2. The login pages are SvelteKit static.** DL R1.6, ruled 2026-09-07:
SvelteKit static, client-rendered, served by the Go binary. BS:621-622's
"Go-templated, no JavaScript" is superseded, and the rung's close
records the correction.

**O3. The `web/` gate leg rejoins at rung 5.** DL section 6, ruled
2026-09-07. Rung 5 corrects `CLAUDE.md:118`, `scripts/check.sh:5`, and
the third copy DL does not name, `Makefile:33-34`.

**O4. The string threshold kind is rung 1b's, and it landed.**
`KindString` and `KindFloat` arrived at rung 1b Task 5, with threshold
validation moved to registration time. Rung 5's string-valued thresholds
register against the shipped kind.

---

## The product forks, ruled

Six questions changed what a person sees, configures, or is asked for.
Geoff ruled all six on 2026-09-08. Each carries the recommendation this
spec made and the ruling that answered it.

**Fork 1: where the roster lives.** Recommendation: a file at a
configured path, `DUBPLATE_ROSTER_FILE`, edited by hand and reloaded by
a restart.

**Ruled (Geoff, 2026-09-08): reversed. Adding a group member is a
web-interface action, and rung 5 builds it.** The roster is a
file-backed store behind one loader interface, with an atomic write path
through `internal/atomicfile` and an admin flag on every member. Three
surfaces reach it. The `dubplate-server member` verb group writes it
from the command line, which is how a fresh install gets its first
admin. An admin-only add-member page in rung 5's pages pass writes it
from the browser, beside the login page. The loader reads it, and it
re-reads after any write so a running server never serves a stale
roster. The web never grants the admin flag to a first login, so the
only path to the first admin is the CLI verb.

What it changes for an adopter: after the one CLI bootstrap, adding a
member is a form, and no restart follows. Rung 6 later moves the
store behind the same interface without touching a call site, which is
the same promise the recommendation made. R37 answers who places the
file on a fresh deployment, and fork 1 changes that answer: the service
writes the file, so `deploy.sh` pre-creates the directory it lives in
and the CLI verb creates the file, and no deploy ever overwrites it.

> Erratum, 2026-09-23 (rung 6 spec, erratum 8, ruling 1 yes): this pre-create retires at rung
> 6d. Once `member import` moves the roster into the store, the roster leaf and its
> directory are no longer needed and `deploy.sh` stops pre-creating them.

**Fork 2: the mail transport.** Recommendation: authenticated SMTP
submission, and only that. Five configured values, any provider. What it
changes for an adopter: they need one mailbox and one app password, and
no Cloudflare account. For the estate the recommendation is Fastmail
with an app password on port 587 with STARTTLS, because a salvage
comment dated 2026-08-31 records Hetzner blocking outbound 465 from the
box. The estate's own pings Worker cannot serve magic links: Email
Routing sends only to verified destination addresses, and a magic link
goes to whatever address a person typed.

**Ruled (Geoff, 2026-09-08): as recommended.** SMTP submission only,
five configured values, any provider. The estate uses Fastmail on port
587 with STARTTLS.

**Fork 3: session lifetime, link expiry, and the rate limits.**
Recommendation: a 15-minute link, a 90-day rolling session under a
180-day absolute cap, five link requests per hour per address, twenty
link requests per hour per client address, thirty consumption attempts
per hour per client address, and a global ceiling of sixty outbound link
mails per hour. What it changes for a member: how often they sign
in again, and how quickly a mistyped address stops getting mail. Every
one is a threshold with a default, so an adopter changes it in the
environment file. Revision 1 carried the first, second, third, and fifth
of these and no per-client-IP link limit, which left roster enumeration
and outbound mail unbounded across distinct addresses (R28).

**Ruled (Geoff, 2026-09-08): as recommended.** Every one of the seven
values ships as a registered threshold carrying that default.

**Fork 4: what the login page says to an address that is not on the
roster.** Recommendation: keep the kind answer, pointing at the
configured admin contact, with a flag for an adopter who wants a generic
outcome. What it changes: a stranger learns whether an address is known.
The roster is a handful of addresses, a mistyped email deserves an
answer, and the mitigation the kind answer rests on is the pair of limit
keys R28 sets: the per-address key bounds mail to one person, and the
per-client-IP key plus the global send ceiling bound enumeration across
distinct addresses. Revision 1 rested the same recommendation on a
per-address key alone, which bounded nothing an enumerator does. PRA
records that PocketBase takes the opposite default, so this is a real
fork rather than a formality.

**Ruled (Geoff, 2026-09-08): as recommended.** The kind answer names the
configured admin contact, and a threshold flag switches the whole
deployment to the generic outcome. Fork 1 supplies the contact: it is
the roster's admin entry, so no config key carries an address.

**Fork 5: who becomes the first admin.** Recommendation: rung 5 builds
no admin bootstrap, and the runbook carries the ordering.

**Ruled (Geoff, 2026-09-08): retired by fork 1.** The `dubplate-server
member add --admin` verb is the bootstrap, and it is the only path that
grants the admin flag to an account that has never signed in. What
remains of the recommendation is the streamer half, which fork 1 does
not touch: a fresh adopter still creates the streamer's own admin
account before the SSO capability is enabled, and the runbook still
carries that ordering.

**Fork 6: whether the login page offers a typed code as a fallback.**
Recommendation: no, at rung 5. The GET interstitial plus the POST
consumption already defends the failure mode a code would address, a
link-scanning mail gateway, and the estate relies on that same defense
in the pings Worker. A typed code also reintroduces the transcription
burden WCAG 2.2's accessible-authentication guidance warns about, which
is the criterion an emailed link satisfies.

**Ruled (Geoff, 2026-09-08): as recommended.** No typed-code fallback at
rung 5.

---

## Rulings

Each is marked **ruled, conductor 2026-09-08, as applied**, meaning the
conductor accepts it in the form this revision folded it. The four that
fork 1 contradicted are amended to match the fork and carry Geoff's
date. Rulings that are also product forks say so.

**R1. The roster is a file-backed store with a write path, behind one
loader interface.** *Ruled, Geoff 2026-09-08.* Rung 5 declares
`DUBPLATE_ROSTER_FILE` in its own file and rewrites the
`DUBPLATE_CONTRIBUTORS_FILE` supersession message, which today says the
roster loads from rung 6's store (`keys.go:176-177`). The new message
names the new key.

The interface is the whole ruling. One type opens the file, validates
it, answers a lookup, lists the roster, and adds a member; every caller
speaks to that interface and none of them knows the format. Rung 6
substitutes its store behind the same interface without touching a call
site, which is what makes building the roster at rung 5 cheap rather
than a thing rung 6 pays for twice.

The write path adds an address, a display name, an optional
`streamer_username`, and an admin flag, validates the whole record
before it writes, writes the file atomically through
`internal/atomicfile`, and re-reads so the in-memory roster the server
answers from is the file it just wrote. A write that fails validation
changes no byte on disk. Concurrent writes serialize behind one mutex
inside the store, because two admins adding members at once through the
web must not lose one of them to a rename.

The file's format is what a program writes rather than what a person
edits by hand, so it is JSON through `encoding/json`. A program cannot
preserve the comments a YAML roster invites, and reading YAML would take
a module dependency R26 does not budget. `config/roster.example.json`
ships as the adopter's starting point, and the runbook shows the CLI
verb rather than the file. **Product fork 1, reversed by Geoff.**

**R2. The roster loader fails closed.** *Ruled, conductor 2026-09-08, as
applied.* A missing, unreadable, or unparseable roster file is a load
error naming the path. This supersedes the salvage loader, which returns
an empty roster and a nil error for a missing file
(`_salvage/internal/config/contributors.go:33-39`). A fail-open roster
lets a deployment start and refuse every login as unknown, which reads
to a member as a broken login rather than a broken configuration.
Fork 1's write path does not soften the rule. The store re-reads the
file it just wrote, and a write whose re-read fails to parse is an error
the caller sees rather than a roster the running server has already
accepted.

**R3. `streamer_username` is written as required-shaped from the
start.** *Ruled, conductor 2026-09-08, as applied.* It stays optional in
the schema per B9. Wherever present it is validated as the stable
principal: ASCII-printable, unique across the roster, bounded in length,
and collision-checked against every roster name at startup, in `doctor`,
and now at write time, since fork 1 gives the collision a live path in.
The write-time check runs against the roster as loaded, inside the same
lock the write holds, so two simultaneous adds cannot both pass it.
2.0's hook 5 promotes it to required, and a roster written without this
validation is a migration rung 6 pays for.

**R4. The revocation epoch keys on the normalized login address.**
*Ruled, conductor 2026-09-08, as applied.* This is the one per-user
record deliberately keyed on an address, because the epoch must exist
before any durable user id does. The re-key to rung 6's surrogate id is
recorded here as rung 6's obligation, so rung 6 inherits it rather than
discovering it. S20:367-369 is the constraint on the other side.

> Erratum, 2026-09-23 (rung 6 spec, erratum 10): the rung 6 spec settles this obligation as a
> permanent dual-key read, not a one-time swap. Rung 6 reads the epoch by the normalized
> address and by the surrogate `UserID`, and a revocation bumps both keys together, so a
> record minted before the surrogate id existed and a record minted after it both resolve.

**R5. The pending identity-source wall row retires into a real selector
key.** *Ruled, conductor 2026-09-08, as applied.* `ownedWall` carries
`{Name: "identity source", Pending: "pending (arrives at rung 5)"}`
(`keys.go:19-25`). Clearing the `Pending` field alone would make `Load`
require an environment variable literally spelled `identity source`.
Rung 5 removes that row from `keys.go` and registers
`DUBPLATE_IDENTITY_SOURCE` as a wall row in its own file, in the same
commit, on the same rule rung 1b wrote for reserved classes. Revision 1
then said "the wall proves presence", which is false against the landed
registry: registering a row claims its name and creates no requirement.
R29 states the `internal/config/config.go` edit that makes presence
real. The auth package proves the value names a built source and fails
closed on anything else.

**R6. `doctor` gains a probe row kind.** *Ruled, conductor 2026-09-08,
as applied.* B2's boundary proof and B8's client-IP mode report land as
rows whose status comes from running a check rather than from
`config.Load`. The alternative, a separate verb, is declined because the
birth requirement names `doctor`.

The boundary probe dials the service's own configured bind address over
loopback with a spoofed header set and reports what came back. Revision
1 ran it in-process through `httptest`, which cannot fail for any
deployment reason: not a bind address exposing an unstripped path, not a
proxy in front re-adding a header, not a unit running an older binary.
BS:422-425 asks `doctor` to prove that a spoofed header from a
non-trusted source is rejected, and a source is a deployment concept, so
a probe compiled into the process it inspects is a build-time assertion
wearing an operator-facing label. The in-process check survives as the
gate-resident test, and the `doctor` row is the live dial. When no
listener answers, the row reports that plainly rather than passing. The
row's output still states what it proved and what it cannot: it proves
this deployment refuses a spoofed header, and it cannot prove an
adopter's proxy strips one.

Both probe rows meet two gate-resident assertions rung 1b left behind.
`doctor_test.go:129-130` fails on any line beyond the owned, reserved,
and threshold rows, and `TestDoctorCmd_LineCountFollowsRegistry`
(`doctor_test.go:151-175`) asserts the total equals the registry's own
counts and re-asserts it after one added row. The per-source satisfied
row this spec's config section promises meets the same pair. The plan
states which way it goes, either registering the probe and per-source
rows through the registry so the derivation holds, or extending both
tests, and it may not discover the collision at implementation time.

**R7. SMTP is the only transport rung 5 builds.** *Ruled, conductor
2026-09-08, as applied.* The seam is one method taking `ctx` first. JMAP
is declined: the best third-party Go client is v0.5.3 with nine
importers, poplar hand-rolled its own rather than depend on one, and
Fastmail publishes SMTP submission anyway. A relay Worker of our own is
declined: it adds a deployable, a shared secret, and a config key class
to buy what authenticated SMTP submission already gives. Cloudflare
Email Service is adopted as documentation, not code, because its SMTP
submission endpoint fits the five reserved SMTP keys exactly. **Product
fork 2.**

**R8. `github.com/wneessen/go-mail` is adopted.** *Ruled, conductor
2026-09-08, as applied.* MIT, LICENSE fetched 2026-09-08, v0.8.1
published 2026-07-09, 237 importers, small stdlib-leaning dependency
footprint, implicit TLS and STARTTLS policies, TLS 1.2 minimum.
`_salvage/internal/notify` already runs it in production on the box's
reject-mail path and passed the repo gate. `net/smtp` is declined as the
whole answer: it is documented as frozen, it has no implicit-TLS dial,
and it composes no message, so choosing it means hand-rolling RFC 2047
subject encoding, quoted-printable, Message-ID, and Date.

**R9. A magic-link send failure is a failed login.** *Ruled, conductor
2026-09-08, as applied.* The seam keeps its best-effort contract for
every other caller, which is what the salvage package's own comment
records. The login handler is the one caller that treats a send error as
user-visible, because a link that never sends is a login that never
happens.

**R10. Auth state files are 0600 under a 0700 leaf.** *Ruled, conductor
2026-09-08, as applied.* The shadow store's 0644 files and 0755
directories are the wrong inheritance for a consumed nonce set and a
revocation epoch. The roster file joins them under fork 1, because a
file the service writes is a state file whatever its contents. The unit
runs as `music` with `UMask=0027`, and the permissions are stated in the
store's own code rather than left to the umask.

**R11. The consumed-nonce insert is create-exclusive, not an atomic
rewrite.** *Ruled, conductor 2026-09-08, as applied.*
`atomicfile.WriteFile` writes a temp file and renames over the target,
which silently overwrites an existing record. A nonce record whose whole
purpose is first-writer-wins must fail when the record already exists,
so the insert opens the record with `O_CREATE|O_EXCL`, syncs the file,
and syncs the containing directory. INV's read of this question weighed
`WriteFile` against `WriteStream`, and neither is correct for the nonce.

The epoch record is the opposite case, because overwriting is exactly
what a revocation does, and it goes through `atomicfile.WriteStream`
rather than `WriteFile`. Both rename over the target. Only `WriteStream`
fsyncs the containing directory afterwards
(`internal/atomicfile/atomicfile.go:45-56`), and a rename that has not
reached the directory can be lost in a crash. For this record the
pre-crash state is the un-revoked one, so the weaker call fails open on
the one file whose whole purpose is to close access. The two records
then agree on durability.

The roster's write path (R1) takes `WriteStream` on the same argument.
It is a whole-file rewrite, overwriting is what adding a member does,
and a rename that never reaches the directory would drop a member an
admin watched the page accept.

**R12. The state-dir agreement check is generalized over greenfield
units.** *Ruled, conductor 2026-09-08, as applied.*
`scripts/check.sh:153-168` names `systemd/dubplate-shadow.service`
literally. Rung 5 selects every unit whose `ExecStart` names
`/usr/local/bin/dubplate-server`, asserts each `ReadWritePaths` entry is
the environment template's `DUBPLATE_STATE_DIR` or a leaf directly under
it, and asserts `deploy.sh` pre-creates each leaf. Rungs 6, 7, and 9
each add a state-directory consumer, and a third hard-coded pair is the
point at which the check should have been data.

**R13. Node and npm join the gate's hard-required tools.** *Ruled,
conductor 2026-09-08, as applied.* Five of DL's ten gates read a built
bundle. The repo rule is that a check either gates or does not exist,
and `scripts/check.sh:19-38` already fails by name for go,
golangci-lint, shellcheck, ruff, and gitleaks. Node joins that list,
with the error message naming mise as the install path. The alternative
is a gate that stops checking every web clause the moment a machine
lacks Node. Nothing Node ships to the box, and the box never runs this
gate.

**R14. Rung 5 declares the presentation keys its serve layer reads.**
*Ruled, conductor 2026-09-08, as applied.* DL R3.3's two theme keys,
R4.3's branding values, and R4.4's operator CSS file path are declared
in rung 5's own file with `KindString`, because rung 5's serve layer is
what reads them and rung 5 stamps the composed `data-theme` value R3.6
defines.

Each is validated before it reaches a served byte. The match value is
checked against the accepted set and refused with an error naming that
set, which discharges DL R3.2, and the composed `data-theme` value is
checked against the known match and palette names before substitution.
Branding values are validated against the value grammar their variables
accept and escaped when rendered into the generated stylesheet; DL's G8
already gates the names, and revision 1 constrained nothing about the
values. The operator CSS file is read from the configured path only,
served with a fixed content type, and never resolved from any
request-supplied component. These are the two places operator text
becomes attribute markup and CSS declarations, which is the class SEC
filed against the salvage page's `fmt.Appendf` render.

**R15. The origin check denies a protected POST carrying neither
`Origin` nor `Sec-Fetch-Site`.** *Ruled, conductor 2026-09-08, as
applied.* `net/http.CrossOriginProtection` (Go 1.25, and the module is
on `go 1.27`) is the mechanism, wrapped so that the documented default
of allowing header-less requests is tightened.

The rule's scope is cookie-authenticated state-changing requests, and a
request authenticated by a bearer credential rather than by the session
cookie is exempt. Fork 1's add-member POST is the highest-value member
of that scope, since it is the one request in rung 5 that grants another
person access, and it is covered by the shell rule rather than by
anything the page does. Revision 1 justified the blanket form with "all
four protected surfaces are browser-only by construction", which is
false of the proxy path: it carries `/rest/*`, an anonymous surface for
native clients that send neither `Origin` nor `Sec-Fetch-Site`, and
S20:284-285 puts the whole 2.0 OpenSubsonic surface under the same
prefix. Rung 14's non-browser token acquisition inherits the same shell
rule. The mechanism for the exemption is `AddInsecureBypassPattern`,
measured present on the installed toolchain alongside `Check`,
`Handler`, and `SetDenyHandler`. A bypassed path may not read the
session cookie for authentication, which is what keeps the exemption
from becoming a hole. Two consequences are stated so rung 12 inherits a
decided contract rather than a wrapper it rewrites.
`CrossOriginProtection` always allows GET, HEAD, and OPTIONS, so a
native client's state-changing GET is never checked and the streamer's
own authentication is what stands behind it. The fallthrough is empty at
rung 5, and until rung 12 fills it the rule denies a cookie-carrying
state-changing request there with no bypass pattern registered.

`gorilla/csrf` is declined: it carries advisory GO-2025-3884 against its
trusted-origin handling, which is the exact dependency this rung should
not take when the stdlib implements the same check.

**R16. The rate limiter map is bounded and evicting, and a test proves
it.** *Ruled, conductor 2026-09-08, as applied.*
`golang.org/x/time/rate` supplies the bucket. The keyed map, its size
bound, and its eviction are rung 5's own code, because an unbounded map
of limiters keyed by attacker-supplied addresses is the
memory-exhaustion bug B8 names. A gate-resident test floods the map with
distinct keys and asserts the bound holds.

**R17. The verified-identity type lives in `internal/identity`.**
*Ruled, conductor 2026-09-08, as applied.* RE cell (5,6) requires the
spec to name the package and forbids `internal/ops` importing
`internal/auth`. A dependency-free leaf holding the verified identity
and its assertion channel satisfies both: rung 6's ops vocabulary
imports the leaf, and the server mechanism stays in `internal/auth`.

**R18. The nonce and epoch records stay in the state directory after
rung 6's store lands.** *Ruled, conductor 2026-09-08, as applied.* RE
cell (5,6) records that no document answers this and that answering it
"no" is legitimate and worth writing down. The records are small,
self-expiring, and needed before the store opens, so no migration is
owed and rung 6 absorbs neither. The one change rung 6 makes is R4's
re-key of the epoch, permanent and dual-keyed rather than a one-time
swap (erratum 2026-09-23, rung 6 spec, erratum 10).

**R19. Rung 5 ships one font family and does not choose it for taste.**
*Ruled, conductor 2026-09-08, as applied.* DL R5.2's license bar,
manifest, and license texts apply in full, and the plan records each
face's license URL. Rung 7 owns visual authoring and the attended
before/after, so rung 7 replaces the face without touching the delivery
mechanism.

**R20. The listener sets `IP_FREEBIND` through `net.ListenConfig`.**
*Ruled, conductor 2026-09-08, as applied.* The salvage server does this
for boot resilience when the configured address is not yet present.
Adopting the mechanism keeps the unit off `network-online.target` and
removes a boot-order failure mode, and it costs one `Control` function.
The address rung 12 will bind is not known at rung 5, which is the case
the mechanism protects.

**R21. The forward-auth verify endpoint is off by default.** *Ruled,
conductor 2026-09-08, as applied.* An unguarded verify endpoint is an
identity oracle. It is an integration class with an enable flag whose
correlated configuration is required-when-enabled, per B10's rule for
integrations.

**R22. The trusted-proxy client-IP mode is an integration class, not a
threshold.** *Ruled, conductor 2026-09-08, as applied.* B8 asks for
explicit configuration and a `doctor` report of which mode is live. An
enable flag with required correlated keys gives both, and it removes the
need for a separate mode threshold whose value could disagree with the
configuration around it.

**R23. The four passes and their ceilings** are 5a at 3M, 5b at 4M, 5c
at 4.5M, and 5d at 3M, 14.5M in total, with the second code slot's
worktree released at each pass close. *Ruled, Geoff 2026-09-08.* Fork 1
moves 5c alone: the admin page and the `member` verb group take its
deliverable count from thirteen to fifteen, and the pass-split section
shows the arithmetic that lifts its ceiling by 0.5M. The pass-split
section carries the scope and the exit gates.

**R24. The review shape** is per-pass domain reviewers plus a two-lens
fresh-context adversarial cycle at 5d. *Ruled, Geoff 2026-09-08.*
`web-auth-security-reviewer` runs inside 5b, because the ladder entry
says the gate runs inside the rung. `svelte-reviewer` and
`daisyui-a11y-reviewer` run inside 5c. Fork 1 adds a second
`web-auth-security-reviewer` read, at 5c, over the admin page and the
roster write path behind it. That surface is a cookie-authenticated,
privilege-granting write, which is the shape the reviewer exists for,
and it does not exist at 5b for the 5b read to cover.

**R25. Enumeration stays kind by default**, as a threshold rather than a
compile-time choice. *Ruled, conductor 2026-09-08, as applied.*
**Product fork 4.**

**R26. Rung 5 adds exactly two module dependencies**,
`github.com/wneessen/go-mail` and `golang.org/x/time/rate`. *Ruled,
conductor 2026-09-08, as applied.* Sessions, tokens, and the origin
check are stdlib. `golang-jwt/jwt/v5` enters `go.mod` only in the commit
that builds a Cloudflare Access forward-auth front, and never for
minting dubplate's own tokens. The roster's format is stdlib JSON for
the same reason (R1): a YAML roster would take a third dependency this
count does not have, and revision 2 named a YAML file while budgeting no
parser for it.

**R27. The peer address and the reported client IP are two values with
two jobs.** *Ruled, conductor 2026-09-08, as applied.* The peer address
is `RemoteAddr`. It is the only input to any trust decision, and it is
what the trusted-proxy allowlist and the verify endpoint's
permitted-addresses list compare against. The reported client IP derives
from the configured client-address header only when the peer address
matches a configured proxy address, and it is used for rate limiting and
logging alone. Revision 1 carried BS:478-481's sentence without the
condition SEC's own blocker required, which left the limit key, the
audit trail, and the verify endpoint's allowlist all reading an
attacker-chosen value. The full mechanism, the parse rule, and the
strip-set additions are in the trust-boundary section.

**R28. The link-request limiter carries two keys and a global ceiling.**
*Ruled, conductor 2026-09-08, as applied.* Per submitted address and per
client IP, with either one tripping, plus a global bound on outbound
link mail per interval. A single per-address key bounds nothing an
enumerator does, because an attacker who never repeats an address never
trips it, and every roster hit sends real mail. The threshold list
carries all three, and the pair is the mitigation fork 4's kind answer
rests on. **Product fork 3.**

**R29. Rung 5 edits `internal/config/config.go`.** *Ruled, conductor
2026-09-08, as applied.* `Load` requires exactly four hard-coded owned
rows and resolves nothing else except an enabled integration class's
`Required` keys (`internal/config/config.go:241-279`), and `Config`
carries no generic wall accessor. Registering a wall row claims its name
and creates no requirement, so a registered-only selector leaves `Load`
returning nil and `doctor` printing "ok" for an unset key
(`cmd/dubplate-server/doctor.go:34`). Rung 5 adds a `Config` field and a
`w.require` call for `DUBPLATE_IDENTITY_SOURCE`, and resolves the active
source's correlated set through per-source required rows in `Load`, so
the eleven magic-link keys are required by the same mechanism rather
than by prose. The edit is 5b's, and it is a contract change to rung 1's
config core, which is why it is a contested surface rather than a chore.

**R30. Consumption re-resolves the roster and the epoch.** *Ruled,
conductor 2026-09-08, as applied.* The POST resolves the token's address
against the loaded roster and fails with the one generic outcome when it
is absent, and it compares the token's own mint time against the
address's epoch and refuses a token minted before it. Without both
checks an outstanding link survives offboarding: it verifies, passes
expiry, inserts a fresh nonce, and mints a session whose issue time
postdates the epoch, which the epoch check then permits for the full
session lifetime. The window is the link expiry and the attacker chooses
when to click. The per-request session path also re-checks roster
membership, which is what the salvage design did
(`_salvage/internal/config/contributors.go:93`).

**R31. The nonce sweep prunes on the record's own carried expiry, and
the separate retention threshold is a floor.** *Ruled, conductor
2026-09-08, as applied.* Revision 1 carried two pruning mechanisms and
did not say which governs, and one reading lets an adopter who raises
the link expiry keep a shorter retention default and have every consumed
nonce pruned while its token still verifies. Config load validates that
retention is greater than or equal to the link expiry and fails by name
when it is not, on the secret-length-floor precedent. Single use is the
property R11 spends a create-exclusive insert and two fsyncs to buy, and
no accepted configuration may defeat it.

**R32. The served surface carries `nosniff` and `Referrer-Policy:
no-referrer`, and the consumption route is `no-store`.** *Ruled,
conductor 2026-09-08, as applied.* Revision 1 claimed that a token never
reaches a URL that gets logged and named the GET/POST split as what
makes it true. The claim is false and it originates at BS:437-439: the
magic link is a URL carrying the token, so the GET request line reaches
the access log, the tunnel's log, and browser history. The split closes
the prefetch-burn and login-CSRF halves of SEC's blocker and not the
third. What rung 5 does instead: the consumption route logs the path and
never `RawQuery`; the whole served surface carries `nosniff` and
`Referrer-Policy: no-referrer` as a named serve responsibility, adopting
the salvage server's header set
(`_salvage/internal/serve/serve.go:119-128`) that DL R5.6 does not
cover; the consumption route's document is served `no-store` while every
other document is `no-cache` under R4.5; and the client-rendered
interstitial clears the token from the address bar after reading it. The
residual is recorded rather than denied. The close writes an erratum
against BS:437-439.

**R33. The redirect rule is stated positively, and rung 5 accepts no
caller-supplied destination.** *Ruled, conductor 2026-09-08, as
applied.* After a single decode, accept only a value whose first byte is
`/` and whose second byte is neither `/` nor a backslash, and reject any
value carrying a scheme, a `..` segment, a control character, or a
percent-encoded separator. BS:509-511's negative enumeration admits
`//evil.example`, which is relative-only and scheme-free and which every
browser treats as absolute. On the second half: the destination at rung
5 is driver-declared or the core default, never a value a caller
supplies, so MD:102-104's originally-requested path does not travel the
mail round trip. Rung 12 introduces the parameter if it wants one, and
it carries it inside the MAC-covered token rather than as a free query
parameter on the consumption POST.

**R34. Both state records are named through one digest derivation and
one `validID` check.** *Ruled, conductor 2026-09-08, as applied.* The
record name is the hex of a digest, over the nonce for the nonce record
and over the normalized address for the epoch, and both pass the same id
check. Revision 1 gave the naming rule to the nonce store alone and
named the epoch record from the address itself, which is a
path-traversal shape derived from admin-authored roster data and which
puts the roster in cleartext filenames. The digest is also what makes
R4's re-key onto rung 6's surrogate id a change of one derivation rather
than a rename of every record. The confinement test covers both records,
so the unit's `ReadWritePaths` grant rests on one static argument.

**R35. The revocation seam is a read plus a change notification.**
*Ruled, conductor 2026-09-08, as applied.* The read returns the current
epoch for a resolved identity, and it is the authority. The notification
is the convenience that lets a subscriber act promptly. A publish-only
event cannot satisfy S20:245-246, which requires every app password to
carry the user's valid-after epoch: a subscriber that is not running
never learns of the event, a restart loses it, and a credential store
minting a password afterwards has no way to read the epoch it must
stamp. R4's re-key at rung 6 changes the key of both halves together.

> Erratum, 2026-09-23 (rung 6 spec, erratum 10): the re-key is permanent and dual-keyed (see
> R4's erratum above), and the change notification's payload gains the surrogate `UserID`
> alongside whatever it already carried, so a subscriber minting a password can stamp the
> right epoch without a second lookup.

**R36. Rung 5 authors one composed theme as data, not as a roster
entry.** *Ruled, conductor 2026-09-08, as applied.* DL revision 2 ships
no default block, and DL section 7 gives the shipped roster to rung 7,
so revision 1's "rung 7's roster supersedes the standard's default
block" named nothing. RE cell (5,7) recorded the gap as the cell's
substance: the pages render unstyled without at least one block and rung
5 must not author a roster entry. Rung 5 authors one `dubplate` match
block and one palette block as data under `web/src/lib/themes/`,
sufficient for G2's completeness and `--size-field` floor, G9's contrast
ratios, and G10's structural precedence. Rung 7's roster supersedes both
files without touching rung 5's markup, because the pages bind to
variable names only.

**R37. The roster file lives in the state directory, and the CLI verb
places it.** *Ruled, Geoff 2026-09-08.* Fork 1 makes the service the
roster's writer, which moves both the file's home and the question of
who places it.

Revision 2 had `deploy.sh` deliver the roster to
`/etc/dubplate/contributors.yaml` at 0640 `root:music`, on the path it
already uses for the incumbent importer (`scripts/deploy.sh:420-426`).
Neither half survives a written roster. `/etc` is read-only to the unit
under `ProtectSystem`, and a file `deploy.sh` re-delivers on every
deploy would discard every member added through the web since the last
one.

The roster instead sits in its own leaf of `DUBPLATE_STATE_DIR`, owned
by `music`, 0600 under a 0700 leaf on R10's rule, inside the serve
unit's `ReadWritePaths` grant. `deploy.sh` pre-creates the leaf and
never writes the file. `DUBPLATE_ROSTER_FILE` points there for the
estate, and an adopter may point it anywhere their unit can write.

A fresh install therefore starts with no roster, which R2 makes a load
error, and `dubplate-server member add --admin` is what clears it: the
verb creates the file when it is absent and refuses to create a second
one. The repo ships `config/roster.example.json` with invented names as
a reference for the shape, never as a file `deploy.sh` copies. The
estate's incumbent `config/contributors.yaml` stays where it is until
rung 16 retires the incumbent importer, and rung 5 neither reads nor
moves it. The runbook carries the adopter's ordering: install, render
the environment, run the verb, start the unit.

**R38. The deployment is reached over HTTPS, and `doctor` says so.**
*Ruled, conductor 2026-09-08, as applied.* The `__Host-` prefix requires
`Secure`, and a browser refuses a `Secure` cookie over plain HTTP on any
origin but localhost. An adopter running dubplate over HTTP on a LAN
would get a login flow that completes, sets no cookie, and signs nobody
in. Config load refuses a `DUBPLATE_PUBLIC_URL` whose scheme is not
https unless its host is localhost, and `doctor` carries the row, so the
failure surfaces at configuration time. The universality section states
the precondition alongside its no-Cloudflare claim, and the terminator
stays the adopter's choice.

---

## The trust boundary

### Serve as ingress

Serve strips every inbound identity-header spelling on ingress, before
routing, for every path, including the `/rest/*` and `/share/*`
bypasses. The strip runs as a wrapper around the mux rather than inside
any handler, which is what makes "every path" true without enumerating
paths.

The strip set has two halves. The driver's configured header name is the
first, and at rung 5 no driver exists, so that half is empty and the
mechanism by which a driver contributes its name is what rung 5 ships. A
driver declares its header name through the driver seam at rung 12, and
the shell reads the declared name at construction. Rung 12 never edits
rung 5's strip code.

The second half is a fixed extras list this spec proposes and the rung 5
plan verifies name by name against the documentation of each reference
implementation it names. The list covers the trusted-header spellings
the forward-auth field uses, each in both its dash form and its
underscore form. The underscore class is the one SEC:42 names as the one
that bites adopters, and the Go mechanism behind it is worth stating
because it is invisible: `textproto.CanonicalMIMEHeaderKey` uppercases
the letter after each dash and treats the underscore as an ordinary
token character, so `Remote_User` and `Remote-User` are two distinct map
keys and a delete of one leaves the other standing. The plan asserts
that behavior with a test rather than asserting it from memory, and the
strip is written so a new spelling is one line of data.

A third fixed group joins the set: `X-Forwarded-For`, `X-Real-IP`, and
`Forwarded`. They are stripped on every path unless trusted-proxy mode
is enabled and the peer address matches, which is the carve-out R27
defines and the next subsection states in full. SEC's proxy-hygiene
major asks for their ingress removal by name, and revision 1's strip set
covered identity spellings alone.

Hop-by-hop hygiene belongs with the strip, for the same reason the strip
does: it must hold for every path. Serve removes every header a
request's own `Connection` header lists, on ingress, before routing.
Without it a client sending `Connection: Remote-User` causes rung 12's
freshly injected identity header to be deleted before it leaves, a
downgrade any attacker can induce. BS:512-519 names hop-by-hop hazards
inside the proxy-hygiene clause, and revision 1's B12 erratum split that
clause without assigning them. The injection-time consequence stays rung
12's, inside `Rewrite`.

Rung 5's own cookie is a fourth strip target. The session cookie is
removed from any request before it is proxied to the streamer
(BS:441-443). SEC:63 filed the salvage design's failure to do this as a
major, and rung 5 ships the removal even though the proxy that would
need it arrives at rung 12, because the removal belongs to the shell.

The streamer's trusted source is pinned to serve's single observed
address, never a bridge subnet, and the ingress flip precedes enabling
the streamer's externalized auth. Both are rung 12 and rung 16 concerns
that this spec records so they are not rediscovered.

### Two addresses, two jobs

R27's rule is the spine of every trust decision below, so it is stated
here in full rather than inside a rate-limit paragraph.

**The peer address** is `RemoteAddr`, the address the connection came
from. It is the only input to any trust decision in the shell. The
trusted-proxy allowlist compares against it, the verify endpoint's
permitted-addresses list compares against it, and no configured header
can change it.

**The reported client IP** is derived, and only for rate limiting and
logging. When trusted-proxy mode is off, it is the peer address. When it
is on, the shell reads the configured client-address header if and only
if the peer address matches one of the configured proxy addresses, and
otherwise falls back to the peer address. A multi-hop value parses right
to left, taking the rightmost hop that is not itself a configured proxy
address, so an attacker prepending entries cannot choose the result.

The distinction is what makes three separate guarantees true at once: a
limit key an attacker cannot pick, an audit trail that records who
actually connected, and a verify endpoint that an internet client cannot
reach by naming the proxy's address in a header. Revision 1 carried
BS:478-481's sentence about reading the client IP "from the
trusted-proxy pair" and never said what the pair is checked against,
which is the condition SEC's rate-limit blocker asked for by name and
which BS dropped when it absorbed the finding.

The forward-auth exemption follows the same shape. Exactly one identity
header may ever be exempted from the strip, it is the one the active
front's configuration names, and the exemption applies only when the
peer address matches that front's configured addresses. Every other
spelling is still stripped, the exempted header's underscore twin
included. While magic-link is the active source the exemption is
impossible, on BS:499-503's rule that the inactive source is never
consulted. No front ships at 1.0, so this is the mechanism a later front
configures rather than code rung 5 builds. It is written here because
rung 5 fixes the strip's shape and PA:80 closes `internal/serve` to
other lanes afterwards.

### The forward-auth topology

In a forward-auth deployment serve is not in the streamer's request
path, so every serve-side guarantee above is inert. That topology gets
its own boundary and its own documentation, whose plain statement is
that the boundary is the adopter's proxy. The setup steps say that the
driver publishes its strip set to the adopter, that the adopter's proxy
must strip that set before the verify subrequest and before the
streamer, and that the streamer trusts only that proxy.

External corroboration is worth citing in that documentation, because it
tells an adopter this is the field's normal answer rather than our
invention. Authelia's proxy documentation requires a reverse proxy to
remove and replace `X-Forwarded-For` arriving from an untrusted client,
warning that a rule keyed on it can otherwise be hijacked, and it notes
that a cloud proxy must be made to do the same. Miniflux pairs its
`AUTH_PROXY_HEADER` with an explicit `TRUSTED_REVERSE_PROXY_NETWORKS`
setting, the same shape as R22's integration class.

`doctor`'s boundary probe (R6) proves dubplate's half.

### One owned prefix

dubplate's entire HTTP surface lives under exactly one configured
prefix, and the streamer owns everything else. The auth boundary is
decided by configuration, never by a mux miss. MD's routing rule 2,
which hands `/api/*` to dubplate, was a blocker in both adversarial
reviews (SEC:20, UNI:7) precisely because it collides with the
streamer's own API. B3's rule resolves it, and a reader working from MD
alone would not know that.

Rung 5 declares `DUBPLATE_SERVE_PREFIX` with a default. Under the prefix
sit the login page, the consumption interstitial, logout, the static
assets, the generated branding stylesheet, and, when enabled, the verify
endpoint. Rung 14's API lands under the same prefix, and rung 14
terminates it so the later fallthrough cannot capture an unknown API
path (RE cell (14,12)). Rung 5 builds the terminator, since rung 5 owns
the mux: a path under the prefix that matches no route returns 404 from
dubplate and never reaches the fallthrough.

---

## The serve shell

### Composition order

PA:79 requires this spec to name the root-level composition order, and
the order is a property of the chain rather than a feature of the mux.

1. **The ingress strip.** It wraps everything, so no route can be
   reached with an identity header still attached.
2. **Path normalization.** It rejects a raw and decoded mismatch, any
   `..` segment, any empty segment, and any backslash, before prefix
   routing. Revision 1 carried the first two, and SEC's path-smuggling
   major asks for all four. The backslash matters on the fallthrough,
   where rung 12 forwards a raw path to a streamer whose own router may
   read a backslash as a separator, which is the smuggling shape one
   level down that rule 3 cites `mountBasePath` for. It also wraps the
   mux, because `http.ServeMux` performs its own path cleaning and
   redirect internally; a normalization check registered as a route
   would run after the decision it exists to guard.
3. **Prefix routing.** The owned prefix is matched segment-aware, never
   by a plain string trim, and the comparison is case-sensitive bytes on
   segment boundaries. A case-insensitive match would widen the owned
   surface past what an adopter configured. The salvage tree already
   proved this defect class one level down: `mountBasePath` documents
   that `http.StripPrefix` does a plain trim, so an unguarded
   `/uploadfoo` trims to `foo` and passes through. SEC:33 is the same
   shape at the bypass prefixes.
4. **The fallthrough.** Empty at rung 5. Rung 12's reverse proxy
   occupies it, which is why `Mount(prefix, http.Handler)` was rejected
   at PA:77: by the one-owned-prefix rule the streamer owns everything
   else, so the driver is the root mux's fallthrough rather than a
   mounted subtree.

The order is invisible in a route table, so the plan proves it with
tests: a request bearing every spelling in the strip set reaches a
handler with none of them; a request whose raw path and decoded path
disagree is refused before any route runs; a path that string-prefixes
the owned prefix without matching a segment boundary does not enter the
owned surface.

### The extension point

Rungs 6, 7, 12, and 14 each add a file at a declared extension point.
The shape rung 5 ships is a registration seam for the owned prefix plus
exactly one root-level fallthrough slot.

The seam's contract, stated as behavior: a consumer registers a subtree
under the owned prefix by name at construction time, in the binary's own
wiring, and the shell owns the mux the registration lands in. Two
registrations of the same subtree are a programmer error and fail loudly
at construction, on the same precedent rung 1b's registry cites for its
collision gate. The fallthrough slot takes at most one handler, and a
deployment with no streamer leaves it empty, which is what makes rung 5
safe to ship without a driver.

The static surface is two handlers rather than one, because DL R4.5
splits them: a document renderer for `index.html` and an immutable asset
server for everything hashed. The extension point must let rung 7 swap
the embedded bundle for a larger one without touching either.

### Deadlines

The salvage server sets one global 30-second write timeout on a single
`http.Server`. That value kills an SSE stream and any audio response
longer than 30 seconds, which is why B12 asks for per-route handling.
Rung 5 sets the policy even though rung 12 owns the routes that need it,
because the policy is a property of the `http.Server` the shell
constructs and rung 12 cannot change it without writing the shell.

The mechanism is `http.ResponseController.SetWriteDeadline`, which lets
a streaming handler extend or clear its own deadline while the server's
default protects everything else. Rung 5 sets a read-header timeout, an
idle timeout, and a default write timeout, and it exposes the extension
mechanism to the fallthrough handler. Rung 5 has no streaming route of
its own, so the plan verifies the mechanism against a test handler that
writes past the default rather than against the constant.

### Wiring

`newRootCmd` builds the registry visibly, registers rung 4's shadow
declarations, and passes the registry into the subcommands in one
`AddCommand` block (`cmd/dubplate-server/main.go:63-80`). The doc
comment at the top of that file already lists "serve: run the HTTP
server (rung 5)". Rung 5 adds its own `registerAuthConfig(reg)` and
`registerServeConfig(reg)` calls above the block, and `newServeCmd(...)`
and fork 1's `newMemberCmd(...)` inside it, following rung 4's precedent
of declaring in `cmd/dubplate-server/shadow.go`. The `member` verb group
takes its subcommands the way cobra means it to, one parent with `add`
and `list` under it, and it reaches the roster through the same store
interface the server uses rather than through a second parser. RE cell
(5,4) predicted contention here with rung 4, and rung 4's shadow verb
has landed, so the contention is spent.

`doctor` needs no change to display rung 5's new classes, because it
derives every row from the registry
(`cmd/dubplate-server/doctor.go:23-71`). The probe rows of R6 are the
only `doctor` work rung 5 owes, and they do not fit the existing row
shape.

---

## The auth core

### The magic-link token

The token carries a normalized login address, an expiry, and a 128-bit
random nonce from `crypto/rand`. The serialization is unambiguous: every
field is length-prefixed, so no value can be shifted across a boundary
by a separator inside it. A version label and a domain-separation label
are inside the MAC input, both as package constants with their reason
documented, which is where the birth requirement wanted them. The MAC is
HMAC-SHA256, and verification uses `crypto/subtle.ConstantTimeCompare`.

Verification order is fixed by B4 and by the security of the nonce
store: verify the MAC, then the expiry, then insert the nonce. Inserting
first would let an unauthenticated caller fill the nonce directory.

`DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS` is accepted for verification and
never for minting, which is what makes rotation routine rather than an
estate-wide sign-out. A secret shorter than the length floor is a config
load error, not a runtime surprise.

Link hostnames derive from `DUBPLATE_PUBLIC_URL` and never from a
request header.

### Consumption

The magic link is a GET that renders an interstitial carrying the token,
served `Cache-Control: no-store`. Under R1.6 the document is
client-rendered, so the page reads the token from `window.location`
rather than receiving it in a server-rendered hidden field, and it
clears the token from the address bar as soon as it has read it. The
POST consumes, origin-checked. Consumption failures return one generic
outcome, with the distinction only in the server log.

Verification is not the whole of consumption. The POST also resolves the
token's address against the loaded roster, and it compares the token's
mint time against that address's epoch (R30). Without both, an
outstanding link outlives an offboarding by the full link expiry and
mints a session the epoch check then honors.

What the GET/POST split buys, and what it does not, is stated exactly
(R32). It closes the prefetch-burn and login-CSRF halves of SEC's
consumption blocker. It does not keep the token out of a URL, because
the link is a URL. The consumption route logs the path and never
`RawQuery`, the route's response is `no-store` where every other served
document is `no-cache`, and the whole served surface carries
`Referrer-Policy: no-referrer` so a subresource load cannot carry the
token off-origin. The residual, a token visible to a mail provider's
link scanner and in browser history, is bounded by single use and a
15-minute default expiry and is recorded in the residuals rather than
denied.

The interstitial is also the defense against a link-scanning mail
gateway, and the estate already relies on the same shape: the pings
Worker makes its failure endpoint POST-only so a scanner cannot trigger
it.

### The session cookie

One HMAC-signed cookie, named with the `__Host-` prefix, carrying the
normalized address, the issue time, an anchor time, and an expiry, under
its own domain-separation label. Attributes: `Secure`, `HttpOnly`,
`SameSite=Lax`, `Path=/`, and no `Domain`, which the `__Host-` prefix
requires and which makes the cookie host-scoped by definition. That is
why same-origin is a stated precondition of the SSO capability. A
split-host adopter runs without SSO or behind their own forward-auth
front, and no cookie-domain widening is offered.

Rolling refresh moves the expiry and never the anchor, so the absolute
cap survives refresh with no server-side state. Refresh counts
same-origin requests only. `DUBPLATE_SESSION_SECRET_PREVIOUS` verifies
and never mints, as with the link secret.

Per-person revocation is the epoch file: a session whose issue time
predates the epoch for its address is refused. Offboarding is stated at
its true size, and the runbook says it plainly. The epoch revokes
dubplate. Streamer credentials survive it, because `/rest/*` bypasses
serve by design, so the runbook carries an explicit streamer-deprovision
step and a driver deprovision capability is a later candidate.

S20:245-246 names the forward edge: rung 5's identity layer exposes the
revocation seam a credential store subscribes to, so that a roster-level
revoke invalidates every app password 2.0 mints. The seam is a read plus
a change notification (R35). The read returns the current epoch for a
resolved identity and is the authority, because that is what a
credential store stamps on a password it mints and what a later check
compares against. The notification is the convenience half. Revision 1
offered publication alone, which no subscriber can rely on across a
restart and which gives a store minting a password after the event no
way to read what it must stamp.

### The roster

The store reads and writes a file named by `DUBPLATE_ROSTER_FILE` (R1),
validates at load, and fails closed (R2). Validation is B9's list:
unique names, addresses parsed with `net/mail.ParseAddress`, no control
characters, bounded lengths, and the optional `streamer_username`
validated as a stable principal with startup, `doctor`, and write-time
collision checks.

A member record carries the address, the display name, the optional
`streamer_username`, and an admin flag. The flag is the whole of rung
5's authorization vocabulary, since rung 5 has exactly one privileged
action, and rung 6's ops verbs are where a richer model would arrive if
one is ever wanted.

Three callers reach the store, and one interface serves all three. The
`dubplate-server member` verb group adds and lists from the command
line. The admin-only add-member page adds from the browser. The auth
core reads for every login and every request. The interface is what
makes rung 6's substitution a change of one constructor, and it is why
no caller knows the file is JSON.

Every write validates the whole record first, takes the store's lock,
re-checks the address and the principal against the roster as loaded,
writes atomically through `internal/atomicfile.WriteStream` (R11), and
re-reads. A write that fails validation changes no byte. A write that
succeeds is visible to the next request with no restart, which is the
property the fork was ruled for.

The address is validated on the way in as strictly as on the way out.
`net/mail.ParseAddress` accepts it, it normalizes before it is stored,
it is refused if the normalized form already sits in the roster, and it
carries no control character and no more than the bounded length. The
`streamer_username` takes the same treatment as R3 states, because a
principal that collides is a principal two people share.

Every roster change writes one audit record: who made it, through which
assertion channel, which address was added, whether the admin flag was
set, and the reported client IP. The CLI verb writes the same record
with the operating-system user in the actor field, since a command line
has no verified identity to name. That record is the only trace a roster
change leaves, because the file itself carries no history.

UNI:13 is the finding the field exists for: the design fused three
identities into the roster's `name`, and an adopter had no field to
separate them. The three are the inbox directory name, the streamer
principal, and the human display name. Display names stay free. The
principal, whichever field goes on the wire, is ASCII-printable, unique,
and bounded.

The store exposes exactly one resolution point where a login address
becomes an identity value. That single point is 2.0's hook 5: rung 6
substitutes a durable surrogate id there without touching a call site.
Fork 1 gives that hook a second half. A member added through the page or
the verb is where a surrogate id would be minted, so rung 6 replaces one
derivation on the read side and one on the write side, both inside the
store.

> Erratum, 2026-09-23 (rung 6 spec, erratum 9, spec half only): the surrogate id this passage
> describes is named `Identity.UserID`; `Principal` stays the name for the streamer handle
> (`streamer_username`), and the two are never conflated. The `internal/identity` code change
> this naming implies is deferred to rung 6's own lane, since rung 5's `identity.go` is a
> package this document does not edit.

Email matching is case-insensitive on a normalized, length-capped,
lowercased address, and the normalization happens before any keying, any
logging, and any rate-limit lookup.

### Rate limits, abuse, and logging

Limits key on the reported client IP as R27 derives it, never on a
header the peer address does not authorize. `doctor` reports which mode
is live (R6). Buckets come from `golang.org/x/time/rate`; the keyed map
is rung 5's, bounded and evicting, with the eviction proved under a
flood (R16).

Link requests carry two keys, per submitted address and per client IP,
and either one trips (R28). A global ceiling bounds outbound link mail
per interval besides. One key alone is not a limit on enumeration: an
attacker walking distinct addresses never repeats an address, so a
per-address bucket never fills, and every roster hit sends real mail
until the provider's quota answers instead of the server. Consumption
attempts keep their single per-client-IP key.

Every auth endpoint body is size-capped before it is read. Every auth
event is logged with the reported client IP: link requested, link
consumed, consumption failed, rate limit tripped, and header-strip hit.
Tokens and secrets never enter a log record, and the consumption route
logs its path without `RawQuery`.

`/rest/*` is recorded as what the design makes it: an anonymous surface
reachable through serve whose authentication is the streamer's own.
Serve is its only reachable path, and the runbook names the residual.

### CSRF

One rule covers every cookie-authenticated, state-changing surface: the
login POST, the consumption POST, logout, and the cookie-carrying half
of the proxy path. R15's wrapped `CrossOriginProtection` is the
mechanism, stricter than the stdlib default in the one documented
respect. A request authenticated by a bearer credential rather than by
the session cookie is exempt through `AddInsecureBypassPattern`, and a
bypassed path may not read the session cookie for authentication. That
exemption is what `/rest/*`, rung 14's non-browser token acquisition,
and 2.0's OpenSubsonic surface need, and revision 1's "browser-only by
construction" justification did not survive them.

SEC:36 records the defect this must not repeat. The salvage check sat
inside the `/upload` mount and covered no new surface. Stating the rule
once, at the shell, is what makes a surface added at rung 7 or rung 14
covered by construction. The salvage implementation's ordering discovery
is worth carrying: `Sec-Fetch-Site` is honored on its own, with `Origin`
against `Host` as the fallback, because an `Origin`-only check refuses
every development request through a proxy that rewrites `Host`.

Double-submit tokens are declined as redundant. The consumption POST
already carries the token in a form field, and the `__Host-` prefix,
`SameSite=Lax`, and the origin check each do a distinct job on a
browser-only surface.

### The verify-identity seam

The seam exists as an interface from this rung, with magic-link as its
reference implementation and forward-auth fronts as alternates. The
verified identity it yields carries the login address, the resolved
principal, the display name, the identity source that asserted it, and
the assertion channel, which is cookie, header, or service token.

Provenance is the point. The organizer's rule that admin verbs reject
cookie-derived identity is enforceable at the seam only if the channel
travels with the identity, and rung 6's ops request types carry the same
value. The type lives in `internal/identity` (R17), so `internal/ops`
never imports `internal/auth`.

Fork 1 carves one exception out of that rule, and the carve-out is
narrow enough to state exactly. Adding a member is a web action Geoff
ruled for, so the add-member handler accepts a cookie-derived identity
whose roster record carries the admin flag. Nothing else does. Rung 6's
ops verbs keep the rule unchanged, granting the admin flag to a first
login stays impossible, and the exception buys its own mitigations: the
origin check covers the POST (R15), the flag is read from the roster
rather than from anything the request carries, and every change writes
an audit record. The seam is where the flag is read, so a later
forward-auth front asserts admin through the same value rather than
through a second mechanism.

Rung 12 consumes the seam in two ways: it reads a verified identity to
decide whether to proxy, and it injects the resolved principal inside
`Rewrite` on the proxy path. Neither reaches into `internal/auth`.

### The post-login destination

The destination is a driver capability with a core default, because it
carries driver semantics. With SSO enabled the driver declares it,
without SSO the dubplate root serves, and with no streamer the journey's
honest terminal state stands. Rung 5 accepts no caller-supplied
destination, so nothing travels the mail round trip and the consumption
POST carries no destination parameter (R33). MD:102-104 redirected to
the originally requested path, and that is rung 12's to introduce if it
wants it, carried inside the MAC-covered token.

The redirect-target validation rule travels with the mechanism and is
rung 5's to build. It is stated positively, because BS:509-511's
negative enumeration admits the form SEC named first: after a single
decode, accept only a value whose first byte is `/` and whose second
byte is neither `/` nor a backslash, and reject any value carrying a
scheme, a `..` segment, a control character, or a percent-encoded
separator. `//evil.example` is relative-only and scheme-free, and every
browser reads it as absolute.

---

## The mail seam

`internal/notify` is rung 5's, per the contested-surface table. The
build spec's ladder entry 9 and the dependency analysis's Wave B
territory column are errata on this point, which RE cell (5,9) settled:
rung 9 builds the monitoring seam and integrates the mail seam it does
not own.

The seam is general by construction (RE cell (5,7)). It carries a
subject and a body as data and owns no copy of its own, because rung 7's
reject wording is a code contract in `internal/reject` shared verbatim
with the uploader page. One method, `ctx` first, one SMTP
implementation.

The seam owns the header-injection guard. SEC:94 is the finding: SMTP
construction from roster values needs explicit CRLF rejection, and the
guard belongs in the seam rather than in each caller. The guard covers
every header-valued argument, the subject as well as the recipient,
because a subject is a header value and rung 7's reject mail and rung
9's notifications will pass subjects derived from filenames and tag
text. The From value is validated at config load rather than at send.

The salvage package is the starting point under fresh review at the
exemplary-Go bar. It synthesizes From, Date, and Message-ID, RFC 2047
encodes the subject, sends the body quoted-printable, and logs one
record on success. Its two test-only fields each carry a comment saying
so, which is the shape go-conventions asks for rather than an injected
interface.

Error doctrine is R9: best-effort for every caller except the login
handler.

The adopter-facing transport table lives in the runbook, and its four
rows all reach the same Go code. Two worked examples ship: a mailbox
provider's app password, and Cloudflare Email Service's SMTP submission
at `smtp.mx.cloudflare.net`, which uses the literal username `api_token`
and an API token as the password. No default host ships, because a
shipped default host would be a site-specific value in the wrong place.

One item is unverified and the plan verifies it before the cutover
sitting: whether `907.life` DNS currently authorizes both the Fastmail
and the Cloudflare senders. The pings Worker already sends as
`pings@907.life` through Cloudflare while the domain's mail is hosted at
Fastmail.

---

## The web surface

### What the standard binds

DL R1.1 binds rung 5's three pages first and every later web surface
after them. The clauses rung 5 discharges are R1.6, R2.1 through R2.7
(consume the 32-name variable contract and read only semantic tokens),
R3.2 through R3.7, R4.1, R4.3 through R4.6, R5.1 through R5.5, and R5.6.
R3.2 is discharged in config, where R14's validation refuses any match
value but `dubplate` with an error naming the accepted set. R3.4 and
R3.5 are discharged by G2 and G10 over the theme R36 authors.

Two clauses are deferred with their reason. R4.2's server half is rung
5's: the Go binary stamps the composed identity into the document it
serves. R4.2's client half, the per-member pick living in `localStorage`
until rung 6 replaces it, waits for the picker, and rung 5 builds no
picker. Revision 1 claimed R4.2 whole and omitted R3.2, R3.4, and R3.5
in either direction.

The mechanics worth restating so they are not rediscovered. SvelteKit
emits inline scripts, so `kit.csp` must be configured and `script-src`
must permit them by hash. `adapter-static` prerenders every page and
nonces are forbidden on a prerendered page, so `kit.csp.mode` is
`'hash'`. R4.5's rewrite stamps an attribute and adds `<link>` elements,
leaving every inline script's bytes untouched, so a hash computed over
the embedded document stays valid for the served one. Whether the hashes
reach the header from a build-time manifest or from hashing the embedded
document at startup is rung 5's choice, and DL's G6 asserts the outcome
either way.

The served CSP is a directive set: `default-src 'self'`; `script-src
'self'` plus a hash per inline script and no `'unsafe-inline'`;
`style-src 'self' 'unsafe-inline'`; `frame-ancestors 'none'`, `base-uri
'none'`, `object-src 'none'`, `form-action 'self'`; and no `font-src`,
because embedded fonts need no relaxation and adding a font origin is a
standard violation rather than a configuration choice. The `style-src`
relaxation is DL's one owner note riding to the rung 5 sitting:
DaisyUI's `radial-progress` reads `--value` from a style attribute and
most Svelte transitions create an inline `<style>` element at runtime,
so no build-time hash can cover them.

One consequence for the security review, named while it is cheap: MD's
property that the unauthenticated surface carries no JavaScript is spent
under R1.6. The `web-auth-security-reviewer` is pointed at that change
explicitly rather than left to find it.

### The embed path

The Go layer reads the embedded `index.html`, substitutes a named
placeholder token with the composed `data-theme` value R14 validates,
and adds the tier 2 and tier 3 `<link>` elements the SvelteKit build did
not write. The document is served `no-cache` and hashed assets are
immutable, with one exception R32 rules: the consumption route's
document is served `no-store`, because it is the one response whose URL
carried a token.

Every served response carries `X-Content-Type-Options: nosniff` and
`Referrer-Policy: no-referrer` beside the CSP, adopting the salvage
server's header set (`_salvage/internal/serve/serve.go:119-128`). DL
R5.6 rules the CSP directives and is silent on the rest, and revision 1
listed the CSP alone among serve's responsibilities.

The salvage embed package carries two lessons worth adopting in
reasoning. A separately committed placeholder file keeps the package
compiling on a fresh checkout that has never run the web build, because
a bare empty `dist/` fails `//go:embed` at compile time. The embedded
filesystem stays rooted where the build wrote it, and the consumer
strips the prefix with `io/fs.Sub`, because that is the one place able
to return an error rather than panicking at program init.

The salvage static handler's cache rules are adopted with one
re-derivation. It marked responses `private`, which was correct for an
authenticated surface. Rung 5's login page is served to an
unauthenticated visitor, so the plan derives the directive rather than
copying it.

### The pages

The login page is one email field and one submit control, built from
DaisyUI 5's `fieldset`, `legend`, `label`, and `input`, with `validator`
and `validator-hint` for inline validation. The hint reserves its space
so the layout does not shift when it appears. `validator` is client-side
convenience only, and the server's generic outcome copy, the
roster-absent copy, and the rate-limit copy stay authoritative.

Accessibility is a design property of this flow rather than an
afterthought. Email link authentication is a listed sufficient technique
for WCAG 2.2's accessible-authentication criterion, and the same
guidance warns that manual transcription of a code fails it, which is
the argument for link-plus-interstitial over a typed code. The email
field carries `autocomplete="email"`. The "check your email" outcome
reaches assistive technology through a polite live region without moving
focus, and an error uses an assertive one.
`autocomplete="one-time-code"` does not apply, because the token arrives
in the link and rides a hidden field.

The consumption interstitial states what is about to happen and carries
one control that performs the POST.

The add-member page is rung 5's third page and its only authenticated
one. It lists the roster and offers a form of an address, a display
name, an optional streamer username, and an admin checkbox. A signed-in
member without the admin flag never reaches it: the route answers 404
rather than 403, so the page's existence is not a fact a non-admin
learns. The form is built from the same DaisyUI `fieldset` idiom as the
login page, its server-side outcomes are specific rather than generic
since the caller is already trusted, and a duplicate address, a
colliding principal, and an unparseable address each say which one it
was. The list marks who is an admin. The page offers no removal control,
because offboarding is the epoch bump plus a streamer deprovision and
the runbook owns that sequence.

The page's copy joins the string deck beside the login page's, under the
same editor standard and the same register gate.

Rung 5's page copy has an authority of its own, and it is not rung 7's.
The generic outcome, the roster-absent line, and the rate-limit line are
the build's first user-facing strings. Rung 5 authors them against the
writing-voice editor standard (Microsoft), with a register gate inside
5c, and the strings live beside the pages as a named deck so rung 7
inherits something to supersede. Rung 7's reject copy contract is a
different artifact for a different surface, and revision 1's forward
reference to it left rung 5's own copy unowned.

All three pages bind to variable names only, so rung 7's roster
supersedes rung 5's theme files without touching rung 5's markup. Rung 5
authors one `dubplate` match block and one palette block as data (R36)
and no roster entry. DL revision 2 ships no default block, so revision
1's sentence about the standard's default block named nothing, and G2,
G9, and G10 had nothing to read.

Every page on the user path is rendered and read by the lane before any
attended check, per the standing `see-it-before-geoff-tests-it` bar. The
visual-fidelity chain does not apply here: it governs work that must
match an existing reference, and rung 7 owns the uploader's fidelity
gate and the attended before/after.

### The gate leg

The leg builds the bundle and runs DL's G1 through G10, plus the
SvelteKit server-file check, which DL does not restate and which is rung
5's to build because rung 5 creates the first `web/` tree that could
violate it. G3, G4, G5, G6, and G7 read a built bundle or a served
document, and each fails with a named message when the embedded tree is
the placeholder-only tree, rather than passing on an empty iteration.
R13 makes Node a hard requirement of the gate.

---

## Packages and their contracts

`internal/identity`. Holds the verified identity and the assertion
channel and nothing else. It imports no other internal package, which is
what lets `internal/ops` carry a verified identity at rung 6 without
reaching into a server package.

`internal/auth`. Mints and verifies magic-link tokens and session
cookies, owns the nonce and epoch stores, owns the roster store behind
one interface that reads, validates, and writes, limits request rates,
and implements the verify-identity seam. It reads configuration through
values passed in, never from the environment directly. It exposes the
revocation seam. It renders no HTML and registers no route, so the
handlers it provides are values the shell mounts.

`internal/serve`. Owns the shell, the mux, the ingress strip, path
normalization, prefix routing, the fallthrough slot, the extension
point, the `http.Server` and its deadlines, the embedded assets, the
document rewrite, the generated branding stylesheet, and the served
response headers, the CSP among them. It consumes `internal/auth`
through the seam.

`internal/notify`. Sends a message with a subject and a body to one
address over SMTP, guards against header injection, and reports failure
to its caller.

`web/`. A SvelteKit static app with client-side rendering only, no
server-only files, built to assets the Go binary embeds. Its
`@component` blocks carry purpose, contract, and failure mode, and its
props speak through typed JSDoc, per the repo's Svelte comment standard.

---

## Config keys

Every name below is declared in rung 5's own file through rung 1b's
registry. The two reserved classes are removed from
`internal/config/keys.go` in the same commit that declares them, which
is the removal half of the rung 1b contract, exercised here for the
first time.

Rung 5 also rewrites three supersession messages in that file, for
`DUBPLATE_CONTRIBUTORS_FILE` (`keys.go:177-178`) and the two Access rows
(`keys.go:171-176`). The rung 1b contract as written permits only
reserved-class removal (`docs/HISTORY.md:494-497`), so the extension is
stated rather than assumed, on the same footing R5 states it for the
pending wall row: a lane may rewrite the supersession message of a name
its own keys replace, in the commit that declares the replacement. The
contested-surface row names all three edit classes so a later lane
inherits the rule rather than the precedent.

### The wall

| Key | Status | Secret |
| --- | --- | --- |
| `DUBPLATE_IDENTITY_SOURCE` | Required. Replaces the pending `identity source` row (R5). Accepted values are the sources the binary implements, which at rung 5 is magic-link alone. | No |

### Required when magic-link is the active source

| Key | Status | Secret |
| --- | --- | --- |
| `DUBPLATE_MAGIC_LINK_SECRET` | Required. Length floor enforced at load. | Yes, minted through the age store |
| `DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS` | Optional. Verifies, never mints. | Yes, minted at rotation |
| `DUBPLATE_SESSION_SECRET` | Required. Length floor enforced at load. | Yes, minted |
| `DUBPLATE_SESSION_SECRET_PREVIOUS` | Optional. Verifies, never mints. | Yes, minted at rotation |
| `DUBPLATE_PUBLIC_URL` | Required. Link hostnames derive from it. Its scheme must be https unless its host is localhost (R38). | No |
| `DUBPLATE_ROSTER_FILE` | Required (R1). The service writes this file as well as reading it, so it names a path inside `DUBPLATE_STATE_DIR` for the estate and a unit-writable path anywhere for an adopter (R37). | No |
| `DUBPLATE_SMTP_HOST` | Required. | No |
| `DUBPLATE_SMTP_PORT` | Required. | No |
| `DUBPLATE_SMTP_USER` | Required. | No |
| `DUBPLATE_SMTP_PASSWORD` | Required. | Yes, minted |
| `DUBPLATE_SMTP_FROM` | Required. | No |

The registry's integration family gates on a boolean flag, and an
identity-source selector with a value is a different shape, so rung 5
does not widen rung 1b's contract for one consumer. It edits `Load`
instead (R29). `Load` requires the selector by name and resolves the
active source's correlated set through per-source required rows, in one
place, and `Config` gains the field that carries the answer.
`internal/auth` still reads configuration through values passed in.
`doctor` carries a row per source reporting satisfied or unsatisfied,
which meets the same line-count assertions R6's probe rows meet.

Fork 1's admin surface declares no key of its own. The admin flag lives
on the roster record, so the page's authorization reads the roster the
server already loaded, and fork 4's admin contact copy derives from the
roster's admin entry rather than from a configured address.
`DUBPLATE_ROSTER_FILE` stays the only key the roster needs, and the
universality rule that no address lives outside `internal/config` is
satisfied more strictly than before, since now no address lives in
configuration at all.

> Erratum, 2026-09-23 (rung 6 spec, erratum 8, ruling 1 yes): this key does not stay. Rung 6d's
> `member import` moves the roster into the store, after which `DUBPLATE_ROSTER_FILE` is a
> retired setting; the file becomes an import source only.

Revision 1 declined the `Load` edit and rested on "the wall proves the
selector is present". It does not. `Registry.AddWallRow`
(`internal/config/registry.go:138-152`) claims the name and creates no
requirement, `Load` requires four hard-coded rows and nothing else
(`internal/config/config.go:241-279`), and `doctor` prints "ok" for any
row `Load` never mentions (`cmd/dubplate-server/doctor.go:34`). Without
R29 the selector and its eleven correlated keys are required by prose
alone, and a deployment missing all twelve exits 0 reporting green.

### Serve

| Key | Status | Secret |
| --- | --- | --- |
| `DUBPLATE_SERVE_BIND_ADDRS` | Threshold defaulting to loopback. | No |
| `DUBPLATE_SERVE_PORT` | Threshold with a default. | No |
| `DUBPLATE_SERVE_PREFIX` | Threshold with a default. The one owned prefix (B3). | No |

The bind default is loopback, and the reason ships with it. The login
surface is anonymous-reachable and the box is production, so SEC's
Access-retirement minor asks for the bridge gateway and loopback and
never `0.0.0.0`. An adopter reaches the service through the tunnel or
through their own proxy, which the runbook documents. One residual
belongs beside it: R20's `IP_FREEBIND` is what lets the listener bind an
address that is not yet present, and it therefore suppresses the
boot-time error a misconfigured address would otherwise raise. `doctor`
carries a row reporting the addresses actually bound, so the
misconfiguration surfaces there instead.

### Thresholds

Session lifetime, session absolute cap, link expiry, link requests per
hour per address, link requests per hour per client address, consumption
attempts per hour per client address, the global outbound link-mail
ceiling per hour, the auth request body cap, the nonce retention floor,
and the unknown-address behavior (R25). Product fork 3 sets the first
seven values. Each is a threshold with a default, per the universality
constraint's third test.

Two of them are correlated rather than free. The nonce retention floor
must be greater than or equal to the link expiry and config load refuses
the pair when it is not (R31), because the sweep prunes on each record's
carried expiry and a shorter retention would let a consumed nonce
disappear while its token still verifies.

### Presentation (R14)

The theme match and theme palette keys DL R3.3 defines, the branding
values R4.3 renders into the generated stylesheet, and the operator CSS
file path R4.4 serves last. Each is `KindString`. The branding keys are
restricted to names drawn from DL R2.1's set, because R4.3 requires the
generated sheet to declare only those names and to be empty when nothing
is configured.

### Integration classes

**Trusted proxy** (R22). An enable flag plus the proxy addresses and the
client-address header, required when enabled. Off by default, so
`RemoteAddr` is the client address until an adopter says otherwise.

**Forward-auth verify** (R21). An enable flag plus the addresses
permitted to make the verify subrequest, required when enabled. Off by
default.

### Names rung 5 does not declare

The SSO proxy capability is rung 12's. A forward-auth front placed in
front of dubplate is not built at 1.0, so its keys are not declared; the
two superseded Access rows keep their supersession and their message is
rewritten to say that the seam admits a front and that its configuration
arrives with the rung that builds one. Declaring keys for code that does
not exist is what the reserved-class mechanism is for, and inventing a
rung number for it would be worse than saying so plainly.

### The environment template

`env/dubplate-server.env.template` is rung 4's file by its own header,
and `deploy.sh` hard-fails the whole render on one unresolved name. Rung
5 adds no name it has not minted. Every rung 5 secret lands as a
commented pending row, on the `config.WallRow.Pending` precedent, and
rung 16 mints all three through the age-store flow. The registry entry
written at minting states that the `_PREVIOUS` overlap doubles the count
at rotation time.

Rung 5's eight required non-secret names get a row state too, which
revision 1 left unassigned. `DUBPLATE_IDENTITY_SOURCE`,
`DUBPLATE_ROSTER_FILE`, `DUBPLATE_PUBLIC_URL`, and
`DUBPLATE_SMTP_HOST/PORT/USER/FROM` land as literal `NAME=value` lines,
the template's second row state, because none is a secret and none is
ever minted.

That has a cross-rung consequence worth stating once. `dubplate-server
shadow` calls `config.Load` against the same registry
(`cmd/dubplate-server/shadow.go:257`, `:324`, `:361`), so once R29 makes
the selector genuinely required, the box's live shadow verb fails at
`Load` until `/etc/dubplate/server.env` carries the row, and `deploy.sh`
renders that file from this template. Any newly required owned row
changes what the live unit needs at its next deploy, which is why 5d's
exit gate asserts a clean render and a green `doctor` against it.

---

## The state directory

Rung 5 writes two records under one leaf of `DUBPLATE_STATE_DIR`, and
fork 1 adds the roster as a third writable file under a second leaf.
Rung 4 owns the layout both sit in.

**Consumed nonces.** One record per consumed nonce, inserted
create-exclusively (R11), carrying its expiry. A sweep prunes on that
carried expiry, at startup and on an interval, and the retention
threshold is a floor config load holds at or above the link expiry
(R31).

**The revocation epoch.** One record per normalized address (R4),
written with `atomicfile.WriteStream` so the rename reaches the
directory (R11).

**The roster** (fork 1). One file, at the path `DUBPLATE_ROSTER_FILE`
names, under its own leaf so that the grant covering the auth records
covers nothing else. It is 0600 under a 0700 leaf (R10), rewritten whole
through `atomicfile.WriteStream` (R11), and re-read after each write
(R1). It has one fixed name rather than a derived one, so the naming
rule below does not reach it.

**One naming rule for the two records.** A record name is the hex of a
digest, over the nonce for one and over the normalized address for the
other, and both pass one `validID` check refusing an id that is empty,
`.`, `..`, or carrying a separator, on rung 4's precedent. Revision 1
gave the rule to the nonce store alone and named the epoch record from
the address itself. An address local part may legally carry `/`, `.`,
and `..`, so that is a path-traversal shape derived from admin-authored
roster data, and it also writes the roster into cleartext filenames. The
digest is what makes R4's re-key onto rung 6's surrogate id a change of
one derivation rather than a rename of every record (permanent and
dual-keyed; erratum 2026-09-23, rung 6 spec, erratum 10, see R4 above).

The layout follows rung 4's store wholesale in shape: one package-level
constant per leaf so each `ReadWritePaths` entry is a static path, one
function joining it so the name has one home, and creation of the
directories the store needs at open. Permissions are R10's, stated in
the code with the reason.

A test proves that every write the three files take, and the
attacker-shaped `..`-bearing id the two records refuse, stay confined to
the root passed in. That test is the static argument the unit's grant
rests on, exactly as rung 4's is, and it reaches the epoch store because
the epoch store now has the id rule the argument needs.

---

## Deployment

The greenfield serve unit is a new file against
`/usr/local/bin/dubplate-server serve` and `/etc/dubplate/server.env`,
the pair the shadow unit established. The incumbent
`systemd/dubplate-serve.service` is the pre-pivot binary's unit and is
left untouched until rung 16 retires it, so rung 5's unit takes a
distinct file name.

What the unit inherits, each with its reason:

- The hardening block the shadow unit already carried forward, minus
  `AF_NETLINK`, which a greenfield serve runs no CPython subprocess to
  need.
- The restart policy and its 2026-09-01 ruling. A literal
  `StartLimitIntervalSec=0` means the unit never enters failed, so
  `OnFailure=` never fires and a crash-looping user-facing surface
  alerts nobody.
- The `OnFailure=dubplate-ping-fail@<slug>.service` idiom, wired and
  inert for the same reason the shadow unit records: the whole
  `DUBPLATE_PING_*` class is reserved for rung 9.
- `ReadWritePaths` naming exactly the auth leaf and the roster leaf,
  each pre-created by `deploy.sh` before the unit starts, without which
  startup fails with a mount-namespacing error.

Two values get their own reasoning rather than a copy. `TimeoutStopSec`
was argued from a client mid-PATCH, which rung 5 does not have; rung 5's
argument is a request in flight on a page a person is looking at.
Ordering does not take `network-online.target`, because R20's
`IP_FREEBIND` removes the boot-order failure the target would guard.

The state-dir agreement check is generalized in the same pass (R12).

`deploy.sh` gains the unit and the two pre-creates, and nothing else.
Fork 1 keeps that list closed rather than widening it. The roster is a
file the service writes (R37), so `deploy.sh` creates the leaf it lives
in, sets the leaf's ownership and mode, and never writes the file
itself; a deploy that re-delivered a roster would discard every member
added through the web since the last one. The estate's incumbent
`config/contributors.yaml` delivery (`scripts/deploy.sh:420-426`) is
untouched, because it serves the incumbent importer until rung 16
retires it, and `DUBPLATE_ROSTER_FILE` points somewhere else entirely.
`config/roster.example.json` ships with invented names as a reference
for the shape, and the runbook carries the adopter's ordering: install,
render the environment, run `dubplate-server member add --admin`, start
the unit. The box is production, so the deploy waits for a quiet-box
check, and it is named here as an owner touchpoint rather than assumed.

---

## What rung 5 does not build

- The streamer driver, its reverse proxy, its header name, its ExtAuth
  configuration, its trusted-source pin, and its roster sync. All are
  rung 12's. The fallthrough slot stays empty and the strip set stays
  data.
- Any store. Rung 5 needs none by deliberate design, and adopting a
  session library that manages one would introduce a store ahead of rung
  6 in the one rung the parallel spec says must not depend on it.
- A forward-auth front in front of dubplate, Cloudflare Access included.
  The seam admits one. No front is implemented, and no JWT dependency
  enters `go.mod` at this rung.
- The API. Rung 14 owns `/api/v1` and terminates the owned prefix.
- The uploader, its tus transport, its journey records, and the reject
  copy contract. All are rung 7's.
- App passwords, device pages, and share management. All are 2.0's, and
  they consume the revocation seam rung 5 exposes.
- The theme roster, the visual authoring of these pages, and the
  attended before/after. All are rung 7's.
- Member removal, roster editing beyond adding, and any authorization
  model richer than one admin flag. Fork 1 rules the add path in and
  nothing else; offboarding stays the epoch bump plus the streamer
  deprovision the runbook carries.
- A typed-code login fallback (fork 6).

---

## What rung 5 needs from earlier rungs

**From rung 1**, the module, the gate, `internal/config`, and the
`cmd/dubplate-server` skeleton with its subcommand surface. Landed.

**From rung 1b**, the registry: `AddWallRow`, `AddReservedClass`,
`AddIntegrationClass`, `AddThreshold`, `KindString`, the collision gate,
and the rule that a lane removes only its own reserved class from
`keys.go` in the commit that declares it elsewhere. Landed 2026-09-08.

**From rung 4**, `internal/atomicfile`, the state-directory layout, the
environment template with its three row states, the greenfield unit
precedent, and the `deploy.sh` pre-create pattern. Landed.

**From Phase 0c**, the design-language standard: the 32-name variable
contract, the composed `data-theme` identity, the three tiers, the font
delivery mechanism, the CSP, and the ten gates. Landed.

**From rungs 2, 3, and 8**, nothing. The dependency analysis states it
plainly: those rungs contribute nothing to auth, and rung 5 spends no
MusicBrainz budget, needs no venv, and reads no corpus artifact.

---

## Edges to later rungs

**Rung 6** consumes `internal/identity` and takes two obligations: the
epoch's re-key onto its surrogate user id (R4), and the surrogate id
itself at the roster's single resolution point (R3). It absorbs neither
state record (R18). Its contract suite runs `httptest` against its own
handler until rung 12 opens the shell again, which is the cost the
dependency analysis already priced.

> Erratum, 2026-09-23 (rung 6 spec, erratum 10): the re-key obligation is a permanent
> dual-key read on both the normalized address and the surrogate `UserID`, with a
> revocation bumping both keys, not a one-time swap; R35's notification gains the
> `UserID` in its payload.

**Rung 7** adds files at the extension point, embeds its own bundle
through the same static path, authors the theme roster, and owns the
visual authoring of rung 5's three pages and the attended before/after.
It consumes the roster store and the mail seam.

**Rung 9** integrates the mail seam it does not own and builds the
monitoring seam. It adds one journey state.

**Rung 12** takes the fallthrough slot, contributes its driver's header
name to the strip set, injects the resolved principal inside `Rewrite`,
declares the post-login destination, and owns `internal/serve` from its
opening. MD's availability note travels with it: the proxy puts serve in
the streaming path, so serve down is the web player down, while native
clients on `/rest/*` are down only if serve is down as well.

**Rung 14** publishes `/api/v1` under the owned prefix and terminates
the prefix so the fallthrough cannot capture an unknown API path. Its
non-browser token acquisition builds on rung 5's session vocabulary.

**Rung 16** mints the three secrets, runs the ingress flip before
enabling the streamer's externalized auth, and deletes the estate's
Access application from the group path.

---

## The 2.0 hooks

Per CLAUDE.md's build-for-2.0 rule, each rung 5 pass states per task
whether it lands a hook or is neutral.

| Hook | Rung 5's position |
| --- | --- |
| 5, a stable identity (S20:367-373) | **Landed by passes 5b and 5c.** The roster's single resolution point is where rung 6 substitutes the surrogate id; `streamer_username` is validated as required-shaped now (R3); the epoch's address key is recorded as rung 6's re-key obligation (R4), and the digest derivation R34 rules makes that re-key a change of one function; and the revocation seam exposes a read as well as an event (R35), which is what a credential store needs to stamp an app password with the epoch S20:245-246 requires. Fork 1 gives the hook its write half: a member arrives through the store's own write path, so that is where rung 6 mints a surrogate id rather than deriving one, and the `streamer_username` collision check runs there too rather than only at startup. The write path lands at 5b and its two callers land at 5c. |
| 6, the driver seam stays universal | Neutral, and kept open. Rung 5 leaves the fallthrough empty and the strip set as data, and DL already discharged the presentation half by declaring theming out of the seam. |
| Every other hook | Neutral. They are schema, tagging, layout, and migration concerns rung 5 does not reach. |

Two adjacent 2.0 surfaces rung 5 makes cheap rather than lands. The
device page and share management are pages in the same web app as the
login page, so the `web/` tree rung 5 births is the tree they arrive in.
The credential store subscribes to the revocation seam rung 5 exposes,
which is what lets a roster-level revoke invalidate every device at
once.

---

## Prior art

Scored on CLAUDE.md's four criteria. Figures are PRA's, read 2026-09-08
from pkg.go.dev and the projects' own pages; one LICENSE file was
fetched directly, for the one new module dependency recommended for
adoption.

| Entry | Role | License | Quality | Maintenance | Adoption | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| stdlib `net/http` cookies | The whole session mechanism | stdlib | Exemplar | n/a | n/a | **Adopt.** The design is a stateless HMAC cookie plus a per-address epoch file, so there is nothing for a session library to manage. |
| stdlib `crypto/hmac`, `crypto/sha256`, `crypto/rand`, `crypto/subtle` | Token minting and verification | stdlib | Exemplar | n/a | n/a | **Adopt.** The birth requirements specify the construction down to the domain-separation label, which is under fifty lines and no supply-chain edge. |
| `gorilla/sessions` | Session store plumbing | BSD-3 | Turns on shape, not code | v1.4.0, 2024-08-20 | 8,382 importers | Decline. Its value is hiding the cookie serialization this rung must state exactly. |
| `alexedwards/scs` | Server-side session manager | MIT | Reads as a house-bar library | v2.9.0, stable and maintained | 3,701 importers | Decline on shape. It manages a store, and rung 5 has none by design. |
| `golang-jwt/jwt/v5` | Token minting | MIT | Good posture on alg confusion | v5.3.1, 2026-01-28 | 15,146 importers | Decline for minting; **hold** as the Cloudflare Access verification dependency, entering `go.mod` only in the commit that builds that front. Algorithm agility is a liability for a token minted and consumed by one process. |
| `o1egl/paseto` | Token format | MIT | n/a | v1.0.0, 2018-11-08; its own README deprecates protocol v1 | n/a | Decline. Fails maintenance outright. |
| `aidanwoods.dev/go-paseto` | Token format | MIT | Genuinely well made | v1.6.0, 2025-12-27 | 235 importers | Decline. PASETO's value is interoperability across a trust boundary this token never crosses. |
| `wneessen/go-mail` | SMTP send and message composition | MIT (LICENSE fetched 2026-09-08) | Small stdlib-leaning footprint, implicit TLS and STARTTLS policies, TLS 1.2 minimum; already runs in `_salvage/internal/notify` past the repo gate | v0.8.1, 2026-07-09; still v0 | 237 importers | **Adopt** (R8), with the salvage package as the in-repo starting point under fresh review. Adoption is the weak cell and is stated honestly. |
| stdlib `net/smtp` | SMTP send | stdlib | Frozen by its own documentation | n/a | n/a | Decline as the whole answer. No implicit-TLS dial and no message composition. |
| `emersion/go-smtp` | SMTP | MIT | Well regarded | v0.25.0, 2026-08-13 | 404 importers | Decline. Its differentiator is the server half, which rung 5 never needs. |
| JMAP from Go | Mail transport | MIT | n/a | `go-jmap` v0.5.3, 2025-02-01 | 9 importers | Decline. poplar hand-rolled its own rather than depend on one, and Fastmail publishes SMTP submission anyway. |
| Cloudflare Email Service, SMTP submission | Adopter mail option | n/a | n/a | Beta as of 2026-09-08 | n/a | **Adopt as documentation, not code.** It reaches the same five reserved SMTP keys. |
| Cloudflare Email Service REST API or Workers binding | Adopter mail option | n/a | n/a | n/a | n/a | Decline. Identical pipeline, a second code path, no gain. |
| A relay Worker of our own | Mail transport | in-estate | Proven shape at `pings@907.life` | n/a | n/a | Decline. A second deployable, a shared secret, and a new key class for no gain since SMTP submission published. |
| `golang.org/x/time/rate` | The token bucket | BSD-3 | Go team; the house exemplar | v0.16.0, 2026-08-19 | 14,348 importers | **Adopt**, agreeing with S20's existing verdict. The bounded evicting keyed map is ours (R16). |
| stdlib `net/http.CrossOriginProtection` | The origin check | stdlib | Implements the birth requirement's rule | Go 1.25; module on `go 1.27` | n/a | **Adopt**, wrapped to deny when neither `Origin` nor `Sec-Fetch-Site` is present (R15). |
| `gorilla/csrf` | CSRF | BSD-3 | Carries advisory GO-2025-3884 on trusted-origin validation | v1.7.3, 2025-01-23, not the module's latest | 1,103 importers | Decline. |
| Double-submit CSRF tokens | CSRF | n/a | n/a | n/a | n/a | Decline as redundant on a browser-only surface. |
| DaisyUI 5 `fieldset`, `input`, `validator` | The login form | The stack's own | Documented idiom with a non-shifting hint | n/a | n/a | **Adopt**, with `validator` treated as client-side convenience only. |
| WCAG 2.2 technique G218, email link authentication | The accessibility argument | n/a | Listed sufficient technique for SC 3.3.8 | n/a | n/a | **Adopt as the argument**, with `autocomplete="email"` and a live region for the outcome. |
| `autocomplete="one-time-code"` | The interstitial | n/a | n/a | n/a | n/a | Decline. The token rides a hidden field and is never typed. |

Four self-hosted projects were read for lessons only, none
transliterated. Grafana shipped magic links and is removing them, on the
stated grounds that the feature stayed experimental with minimal usage
and that enterprises are moving to phishing-resistant authentication;
the lesson is about audience, since dubplate's magic link competes with
a reused password rather than with an identity provider. PocketBase
treats enumeration as the design problem and returns an opaque id even
for an unknown address, which makes fork 4 a real fork rather than a
formality. Ory Kratos stresses that a good delivery provider is the
feature's reliability, which is why the transport table is not
incidental. Gitea keeps sessions server-side behind a cookie, the road
not taken: a server-side store is what you build when sessions must be
listable and individually revocable, and rung 5 buys per-person
revocation with an epoch file at a fraction of the machinery.

---

## The threat model

The reviewers grade against this section, and a finding that contradicts
it is a finding against the spec rather than against an implementation.

**What is protected.** Access to the group's uploader and, under SSO,
to the streamer's web player. The roster's addresses, and now its
contents: fork 1 makes the roster a thing an attacker would want to
write, not only to read. The two HMAC secrets. The SMTP credential.

**Who the adversary is.** An unauthenticated visitor reaching serve from
the internet, since the login surface is public by construction. A mail
gateway that follows links. A person on a shared device. A misconfigured
or hostile proxy in front of serve. A signed-in member without the admin
flag, reaching for the roster, is in scope, which fork 1 changed:
revision 2 put a member acting outside their own account out of
scope at group scale, and that stops being honest once one member can
grant another access. The rest of that exclusion stands, because rung 5
has exactly one privileged action and nothing else a member could
usefully reach for.

**Entry points.** The login POST, the link GET, the consumption POST,
the add-member POST, logout, the verify endpoint when enabled, the
static assets, and the fallthrough, which is empty at rung 5. The
`member` verb group is not an entry point: it runs as a local process
under an operating-system user the unit's own hardening already bounds.

**The claims the design makes.** No inbound identity header survives
ingress on any path. A path cannot be smuggled through a prefix. No
configured header can change a trust decision, because every trust
decision reads the peer address. A token verifies before it is looked
up, and a nonce cannot be consumed twice across a restart or across any
accepted retention configuration. A session cannot outlive the absolute
cap, and an epoch bump refuses every session older than it, including a
session minted from a link that predates the bump. A state-changing
cookie-authenticated request without a same-origin signal is refused. A
limiter map cannot grow without bound, and neither a single address nor
a single client can drive unbounded outbound mail. A token never reaches
a log record. Only a request carrying a session for a member whose
roster record holds the admin flag can change the roster, that flag is
read from the roster rather than from the request, no web path sets it
on a first login, a roster change that fails validation leaves the file
byte-identical, and every roster change that succeeds writes an audit
record naming the actor and the channel.

**The residuals, stated rather than hidden.** `/rest/*` is an anonymous
surface whose authentication is the streamer's own, reachable only
through serve, and the runbook names it. A dubplate revoke does not
revoke streamer credentials, which is why offboarding carries a
streamer-deprovision step. A magic link is as trustworthy as the mailbox
it lands in and the transport that carries it. The link URL carries the
token, so it reaches a mail provider's link scanner and the recipient's
browser history; single use, a 15-minute default expiry, `no-referrer`,
and a consumption route that never logs `RawQuery` are what bound it,
and the GET/POST split is not. Enumeration is kind by default (fork 4).
A shared device holds a session until it expires or an epoch bump
refuses it, and on an admin's shared device that session can add a
member. An admin is trusted, and rung 5 buys no re-authentication step
in front of the one privileged action; the audit record is what makes
the use visible afterwards. `doctor`'s boundary probe proves this
deployment's own refusal and cannot prove an adopter's proxy strips
anything.

**The two review texts are birth requirements.** SEC carries 6 blockers,
13 majors, and 12 minors, and UNI carries 3 criticals, 6 highs, 7
mediums, and 3 lows: 50 items. The rung 5 plans walk both texts as
checklists, minors included, and the honest form is a table mapping each
finding to the birth-requirement clause that absorbed it, with the
residue called out. That table is 5d's deliverable and the adversarial
cycle's input.

---

## Errata the close applies

**Build spec ladder entry 5, the page stack.** Paste-ready:

> Erratum, 2026-09-08 (rung 5's close): "the login and consumption pages
> (Go-templated, no JavaScript, shared design language)" is superseded.
> Design-language standard R1.6, ruled 2026-09-07, makes them SvelteKit
> static pages, client-rendered, built to assets the Go binary embeds
> and serves, on CLAUDE.md's web rule. Nothing else in this entry
> changes.

**Build spec ladder entry 7, the gate rejoin and the roster.**
Paste-ready:

> Erratum, 2026-09-08 (rung 5's close): the `web/` gate leg rejoins at
> rung 5 rather than rung 7, per the design-language standard's section
> 6, because rung 5's login page is the first web surface. The same
> correction applies to the gate-rewiring bullet at BS:332-333, whose
> "removes the web legs until rung 7 rebuilds `web/`" is the second
> in-document copy. The roster with the birth-requirement validation is
> rung 5's, per the parallel-build approach's contested-surface table,
> and it is a store with a write path rather than a loader alone, since
> Geoff ruled on 2026-09-08 that adding a member is a web-interface
> action built at rung 5. Rung 7 consumes it.

**Build spec B12, proxy hygiene.** Paste-ready:

> Erratum, 2026-09-08 (rung 5's close): the "(streamer rung)" marker on
> proxy hygiene is half wrong. The header-strip set, the ingress removal
> of `Connection`-listed hop-by-hop headers, the path normalization
> ahead of prefix routing, and the per-route write-deadline policy are
> shell-level and land at rung 5. Only the `httputil.ReverseProxy` half,
> its identity injection, its hop-by-hop consequence inside `Rewrite`,
> and its cache-control rule are rung 12's.

**Build spec B4, tokens and logged URLs.** Paste-ready:

> Erratum, 2026-09-08 (rung 5's close): "Tokens never appear in URLs
> that get logged, which the GET/POST split is what makes true" is
> false. The magic link is a URL carrying the token, so the GET request
> line reaches the access log, any intermediary's log, and browser
> history. The GET/POST split closes the security review's prefetch-burn
> and login-CSRF halves and not its third. What holds instead: the
> consumption route logs its path and never `RawQuery`, the served
> surface carries `Referrer-Policy: no-referrer`, the consumption
> response is `no-store`, and the remaining exposure is a recorded
> residual bounded by single use and a short expiry.

Three in-tree copies of the gate-rejoin sentence are corrected in pass
5c rather than at the close, because the leg lands there:
`CLAUDE.md:118`, `scripts/check.sh:5`, and `Makefile:33`. DL's section 6
names only the first two.

**Pass 5a's close, 2026-09-08, adds the corrections below.** Pass 5a
(`internal/serve`, merged at `307001b`) closed with a fresh-context
`web-auth-security-reviewer` read against this spec's own threat model
(`spec:1903-1975`); its findings and their dispositions are the
do-not-re-litigate record at
`docs/superpowers/research/2026-09-08-rung-5a-security-review.md`.

> Erratum, 2026-09-08 (pass 5a's close): **R23's ceiling table**
> (`spec:563-567`) is corrected. 5a's ceiling was re-derived from rung
> 8a's actual before 5a's own plan was drafted, from 3M to **4.25M**,
> per the re-derivation rule R23 already states ("each ceiling is
> re-derived from the previous pass's actual report before the next
> plan is drafted"). Rung 5's combined reservation moves from 14.5M to
> **15.75M**, the 1.25M difference recorded against the reservation
> rather than taken from another pass. 5a's actual spend closed at
> about 3.9M against the 4.25M ceiling.

> Erratum, 2026-09-08 (pass 5a's close): **R24's review shape**
> (`spec:571-580`) gains 5a's own reviewer. The conductor ruled
> (R-J, 5a's plan) that a fresh-context `web-auth-security-reviewer`
> read runs at 5a's own pass end, not only inside 5b and at 5c: the
> shell's composition order is the artifact the whole trust boundary
> rests on, and `spec:2043-2047`'s own split-point argument is the
> reason a security read of it happens on its own rather than
> alongside the token, session, nonce-store, roster, and limiter
> surfaces 5b also carries. R24 otherwise stands as written.

> Erratum, 2026-09-08 (pass 5a's close): **the session-cookie removal's
> pass assignment** is settled. Neither `spec:2035-2041`'s 5a row nor
> its 5b row named which pass removes rung 5's own session cookie from
> a request before it reaches the fallthrough, though `spec:808-813`
> requires the removal. Ruled R-E: 5a builds the removal, at the
> fallthrough boundary, parameterized by the cookie name as a
> construction value 5b supplies once the name exists.

> Erratum, 2026-09-08 (pass 5a's close): **the hop-by-hop scope**
> (`spec:796-806`) is narrowed to what 5a actually built. Serve removes
> every header a request's own `Connection` header lists, with two
> carve-outs: a `Connection` naming `Upgrade` keeps both `Upgrade` and
> `Connection` standing, so a websocket handshake survives the listed
> removal, and a fixed keep-set (`Cookie`, `Origin`, `Sec-Fetch-Site`,
> and the shell's own session-cookie name) is never removed by the
> listed rule regardless of what a client's `Connection` header names.
> The fixed RFC 7230 hop-by-hop set (`Keep-Alive`, `Transfer-Encoding`,
> `TE`, `Trailer`, `Proxy-Authorization`, `Proxy-Authenticate`) stays
> unremoved by 5a's own code, on the finding that
> `httputil.ReverseProxy` already removes that fixed set on the way out
> (rung 12's concern, ruled R-O).

> Erratum, 2026-09-08 (pass 5a's close): **the strip covers trailer
> fields, not header fields alone** (`spec:772-812` names only
> `r.Header`). Every identity spelling in group 1's set is deleted from
> `r.Trailer` on the same pass as `r.Header` (ruled R-P), because
> `net/http` moves an announced trailer key out of `r.Header` and into
> `r.Trailer` once the body is consumed, which a header-only strip
> cannot reach. The pass's first threat-model claim now names both
> maps.

> Erratum, 2026-09-08 (pass 5a's close): **group 2 extends to the whole
> `X-Forwarded-*` family** (`spec:795-800` names only
> `X-Forwarded-For`, `X-Real-IP`, and `Forwarded`). Ruled R-R:
> `X-Forwarded-Host`, `X-Forwarded-Proto`, and `X-Forwarded-Prefix` join
> the three, carrying the same trusted-proxy carve-out, because each
> remaining member steers an absolute URL a downstream handler (5b's
> or rung 12's) builds from client-asserted input.

> Erratum, 2026-09-08 (pass 5a's close): **rule 1's mechanism and rule
> 3's shape** (`spec:918-928`, `spec:918-920` in particular) are
> settled. Ruled R-D: rule 1 refuses when `r.URL.EscapedPath() !=
> r.URL.Path`, a blanket ban on percent-encoding in the path rather
> than a narrower `RawPath`-based test, because `RawPath` is populated
> only when the escaped form differs from the default encoding of
> `Path` and is undefined for an unencoded request. The strictness is
> named rather than absorbed as a permanent default: rung 12 revisits
> it against the streamer's real client paths, share links, cover-art
> paths, and any path a Subsonic client percent-encodes, with measured
> evidence before relaxing it. Rule 3, stated at `spec:918-928` and in
> the composition order (`spec:907-944`) as "any empty segment", is
> corrected to a doubled-separator rule: a single trailing separator is
> admitted (`/`, `/dubplate/`, and `/dubplate/x/` all pass), and the
> empty-segment reading contradicted C4's own routing of the bare
> prefix. **The normalization rule count moves from four to six**: the
> composition order's rule 2 (`spec:907-944`) and the pass-split
> table's 5a row (`spec:2058`) both said "four-rule normalization";
> the landed rule set adds a length bound (R-H, answered `414`) and a
> refusal for a request target whose path does not begin with a
> separator (R-U, admitting an absolute-form target while refusing a
> `CONNECT` request's empty path).

> Erratum, 2026-09-08 (pass 5a's close): **the read timeout**
> (`spec:965-981`, which names a read-header timeout, an idle timeout,
> and a default write timeout) gains a fourth. Ruled R-V: a 30-second
> read timeout, adopted at the salvage server's own value, bounds a
> request body trickled a byte at a time, since neither the read-header
> timeout nor the idle timeout covers that shape. If a later rung wants
> a streaming upload route, the alternative is `http.MaxBytesReader` on
> the body plus `http.ResponseController.SetReadDeadline` per route,
> carried to 5b as a named option rather than adopted here.

> Erratum, 2026-09-08 (pass 5a's close): **the template-classification
> test's home, and the wall's fifth classification.** `Wiring`
> (`spec:982-1005`) does not say where the test proving every template
> name classifies against the production registry lives. It lives in
> `cmd/dubplate-server`, not `internal/config`: the rung 1b registry
> contract composes the production registry as each rung declares its
> own keys in its own file, so `internal/config` alone can no longer
> classify a template's names without reporting a declared name as
> unknown once a later rung's file exists only in the binary package.
> `internal/config` keeps a generic classification-mechanics test
> against a fixture registry it declares for itself.
> `rejectNamespace`'s wall gains a fifth classification alongside
> owned, reserved, superseded, and integration-class rows: a registered
> threshold is a declared name, so `Load` accepts a value set against
> it rather than reporting it unknown. Rung 4's thresholds classified
> through their integration class and never exercised this branch;
> `DUBPLATE_SERVE_PREFIX`, the first standalone threshold rung 5
> declares, is what makes 5a the first pass to need it. Recorded as a
> residual (security-review finding 24, above): no per-threshold value
> validation runs at `Load` time beyond the registry's own
> default-presence check, a property of the rung 1b wall contract
> rather than a 5a gap.

**Pass 5b-i's close, 2026-09-12, adds the corrections below.** Pass 5b
split into 5b-i (`internal/identity`, `internal/auth`'s config-facing
core, and the state records; merged at `1d4ed29`) and 5b-ii (the
roster, limiters, and mail seam). 5b-i's own review disposition table
is `docs/HISTORY.md`'s pass entry, not repeated here.

> Erratum, 2026-09-12 (5b-i close): **R23's ceiling table and the pass
> split** (`spec:569-575`; R-5b1). 5b becomes 5b-i and 5b-ii, each with
> its own ceiling, tracked against R23's live 15.75M reservation (set
> at 5a's close, `spec:2042-2050`) rather than 5b's original 4M.
> 5b-i's ceiling was 5.0M against an actual of roughly 5.5M; the
> overage is recorded against the reservation rather than funded from
> another pass. The pass-split table's 5b row (`spec:2168`) and its
> exit gate now describe 5b-i's own scope; 5b-ii inherits the roster,
> the limiters, and the mail seam, per "The 5b-i to 5b-ii handoff"
> below the pass ledger entry this close writes.

> Erratum, 2026-09-12 (5b-i close): **R24's review shape**
> (`spec:577-580`; R-5b16) is corrected. "`web-auth-security-reviewer`
> runs inside 5b, because the ladder entry says the gate runs inside
> the rung" described a single pass that no longer exists once 5b
> splits. As ruled: a fresh-context `web-auth-security-reviewer` read
> runs at each half's own pass end, 5b-i's and 5b-ii's, not one read
> "inside 5b" as originally stated. 5b-i's own read was not run this
> window for budget and is owed to 5b-ii's window as its first item.

> Erratum, 2026-09-12 (5b-i close): **R29's wording**
> (`spec:623-635`) understates what landed. `Load` requires every
> owned-now wall row generically, by iterating `reg.ownedWall` and
> requiring every row whose `Pending` is empty
> (`internal/config/config.go:362-365`), rather than a hard-coded
> `w.require` call for `DUBPLATE_IDENTITY_SOURCE` by name. This closes
> risk 11 (an owned-now row a rung declares but `Load` never requires)
> for every rung that adds one, not only for the selector.

> Erratum, 2026-09-12 (5b-i close): **the environment template's row
> assignment** (`spec:1640-1645`, `spec:2205`; R-5b7) is corrected.
> One literal row, `DUBPLATE_IDENTITY_SOURCE=magic-link`, lands at
> 5b-i rather than at 5d, because `TestTemplateKeysCoverEveryOwnedNowRow`
> is a hard gate the config contract change makes fail otherwise. The
> contested-surfaces table's "rung 5 adds commented pending rows only"
> (`spec:2205`) holds for the three secrets and is corrected for this
> one non-secret row. The other six non-secret names
> (`DUBPLATE_ROSTER_FILE`, `DUBPLATE_PUBLIC_URL`, and
> `DUBPLATE_SMTP_HOST/PORT/USER/FROM`) stay 5d's, per the count
> correction below.

> Erratum, 2026-09-12 (5b-i close): **two of `spec:2168`'s exit-gate
> assertions belong elsewhere.** Of the semicolon-separated list in the
> 5b row's exit gate, the roster-removal half of the third assertion
> ("a link for an address removed from the roster") is 5b-ii's, since
> the roster does not exist at 5b-i; and the verify-subrequest half of
> the eighth assertion ("a verify subrequest from such a peer is
> refused") is 5d's, since `spec:2170` gives the forward-auth verify
> endpoint to 5d. 5b-i's own exit gate is the remainder of that list,
> minus both named halves.

> Erratum, 2026-09-12 (5b-i close): **the staleness pre-task's `serve`
> reserved-class bullet** (`spec:2297-2300`) is stale as of 5a's close.
> "Does `internal/config/keys.go` still carry the `auth and identity`
> and `serve` reserved classes...? Rung 5 removes all three" no longer
> describes the tree once any pass after 5a starts: 5a already removed
> the `serve` class and declared its own file. Only the `auth and
> identity` class and the pending wall row remained for 5b-i to remove,
> and 5b-i removed them (`internal/config/keys.go`,
> `cmd/dubplate-server/auth.go`).

> Erratum, 2026-09-12 (5b-i close): **the `keys.go` pointers at
> `spec:1506`** (`keys.go:177-178` for `DUBPLATE_CONTRIBUTORS_FILE`'s
> supersession message, `keys.go:171-176` for the two Access rows) are
> stale. Regenerated by grep against the tree at this close, the
> supersession message is `internal/config/keys.go:151-152` and the two
> Access rows are `:145-148`.

> Erratum, 2026-09-12 (5b-i close): **the `atomicfile` pointer at
> `spec:434`** (`internal/atomicfile/atomicfile.go:45-56`) resolves to
> `:55-91`. The directory fsync that makes `WriteStream` durable across
> a crashed rename is at `:82-89`.

> Erratum, 2026-09-12 (5b-i close): **the epoch's session check reads
> the anchor, not the issue time** (review finding B1). `spec:1085-1086`
> states "a session whose issue time predates the epoch for its address
> is refused." As landed, `EpochStore.CheckSession`
> (`internal/auth/epoch.go:165-181`) reads `Session.AnchoredAt`, never
> `IssuedAt`. Refresh moves `IssuedAt` but never the anchor
> (`internal/auth/session.go:130-132`), so a check that read the field
> refresh moves would let a session refreshed once after a revocation
> bump carry a post-bump issue time and pass permanently.

> Erratum, 2026-09-12 (5b-i close): **the token's field list gains the
> mint time.** `spec:1016-1017` enumerates three fields (address,
> expiry, nonce) where R30 (`spec:640-642`) requires a fourth, the
> mint time, so consumption can compare it against the address's
> epoch. As landed, `LinkToken` carries all four
> (`internal/auth/token.go:89-106`), and `linkFieldCount` is 5 on the
> wire once the version label joins them (`internal/auth/token.go:49`).

> Erratum, 2026-09-12 (5b-i close): **the serialization claim covers
> the wire form, not field independence.** `spec:1017-1019` states
> "every field is length-prefixed, so no value can be shifted across a
> boundary by a separator inside it." As landed, each field is
> base64url-encoded before being joined with `.`
> (`internal/auth/token.go:38-45`), so no field value can carry the
> separator; the fields are not length-prefixed.

> Erratum, 2026-09-12 (5b-i close): **the nonce record's carried
> value.** `spec:1664-1668` says the record carries "its expiry"
> without saying whose. As landed
> (`internal/auth/nonce.go:216-240`), the record carries the
> consumed-at time plus the resolved retention duration, and the sweep
> derives the removal moment as their sum; no bare expiry field exists.

> Erratum, 2026-09-12 (5b-i close): **R38's carve-out covers the
> loopback set** (`spec:758-767`, `spec:1530`), not the literal name
> `localhost` alone. `validatePublicURL`
> (`cmd/dubplate-server/auth.go:171-184`) refuses a non-https scheme
> "unless its host is loopback," which is the wider set a host resolves
> to, matching `DUBPLATE_SERVE_BIND_ADDRS`'s own use of the term rather
> than a single literal.

> Erratum, 2026-09-12 (5b-i close): **the non-secret name count.**
> `spec:1640-1645` says "eight" and enumerates seven
> (`DUBPLATE_IDENTITY_SOURCE`, `DUBPLATE_ROSTER_FILE`,
> `DUBPLATE_PUBLIC_URL`, and the four `DUBPLATE_SMTP_*` names); the
> count was already off by one before this close. `spec:2170` gives 5d
> "its eight literal rows." With `DUBPLATE_IDENTITY_SOURCE` landing at
> 5b-i (above), six literal rows remain for 5d:
> `DUBPLATE_ROSTER_FILE`, `DUBPLATE_PUBLIC_URL`, and the four
> `DUBPLATE_SMTP_*` names.

> Erratum, 2026-09-12 (5b-i close): **R33 and R24 belong to 5b's
> scope, not 5a's.** The 5a-to-5b handoff record
> (`docs/HISTORY.md`, rung 5 pass 5a's close) shows R33's post-login
> destination rule (`spec:1247-1265`) was assigned to 5a's Task 3
> before being corrected to 5b's; it landed in `internal/auth` at
> 5b-i. R24's own wording correction is R-5b16, above. The pass-split
> table's 5b row (`spec:2168`) is corrected to include the
> post-login-destination redirect validation among 5b-i's built
> deliverables, alongside the seams it already names.

> Erratum, 2026-09-12 (5b-i close): **R1's own `keys.go` pointer at
> `spec:284`** (`keys.go:176-177`) is stale for the same reason as the
> pointer above: regenerated by grep, `DUBPLATE_CONTRIBUTORS_FILE`'s
> supersession message is at `internal/config/keys.go:151-152`.

> Erratum, 2026-09-12 (5b-i close): **the ceiling erratum's own
> pointer.** A plan citing this rung's ceiling correction gave
> `spec:2078-2087` for it in revision 1 of its own text; the correct
> pointer is `spec:2042-2050`, above, and the rung reservation's live
> baseline is the 15.75M that erratum set, not a figure recomputed from
> the wrong paragraph.

> Erratum, 2026-09-12 (5b-i close): **R-5b14's own wording** describes
> the nonce sweep's one-hour interval constant as landing at 5b-i. As
> executed, `spec:1664-1668`'s "at startup and on an interval" is only
> half built at 5b-i: `OpenNonceStore` sweeps once at open and `Sweep`
> is a callable method (`internal/auth/nonce.go:80-91`, `:143`), but
> no interval constant exists and nothing calls `Sweep` on a schedule.
> The interval driver, the one-hour constant included, is 5b-ii's in
> full, matching R-5b14's own operative clause ("the driver arrives at
> 5b-ii").

**Pass 5b-ii's close, 2026-09-23, adds the corrections below.** Pass
5b-ii (the roster, the mail seam, the rate limiters, the origin check,
and the auth handlers, in the main tree from `1d4ed29`) closed with a
fresh-context `web-auth-security-reviewer` read over both halves of 5b;
its findings and their dispositions are the do-not-re-litigate record at
`docs/superpowers/research/2026-09-23-rung-5b-ii-security-review.md`.

> Erratum, 2026-09-23 (5b-ii close): **the 5b-i security read ran.**
> The R24 erratum above (`spec:2174-2181`) says 5b-i's own read "was not
> run this window for budget and is owed to 5b-ii's window as its first
> item." It was written before the read ran. The read ran on 2026-09-12
> at `1d4ed29` and is recorded at
> `docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md`
> (R-5b18). 5b-ii's own pass-end read, scoped to `1d4ed29..08befea`,
> re-checked all twenty of its findings alongside 5b-ii's own surface.

> Erratum, 2026-09-23 (5b-ii close): **R33's redirect validation did
> not land at 5b-i.** The erratum above (`spec:2287-2296`) states that
> R33's post-login destination rule "landed in `internal/auth` at
> 5b-i" and adds the redirect validation to the pass-split table's 5b
> row. No such function exists in `internal/auth`; 5b-ii's Task 0 and
> its close both confirm it by grep. As ruled (R-5b26), no
> redirect-target validator is built at rung 5: R33 travels as a
> constraint to rung 12, and the validator lands with rung 12's
> destination parameter or not at all. The 5b row's added deliverable
> is withdrawn.

> Erratum, 2026-09-23 (5b-ii close): **5b-ii's ceiling and the rung
> reservation.** 5b-ii ran against a 7.5M ceiling rather than the 4.5M
> provisional that 5b-i's plan carried, and closed at about 6.4M. Rung 5
> now stands at 24.4M (5a's 3.9M actual, 5b-i's 5.5M actual, 5b-ii's
> 7.5M ceiling, 5c's 4.5M, and 5d's 3.0M) against the live 15.75M
> reservation (`spec:2042-2050`). The 8.65M difference is recorded
> against the reservation rather than taken from another pass.

> Erratum, 2026-09-23 (5b-ii close): **the example roster file lands at
> 5b-ii, not 5d.** The pass-split table's 5d row assigns "the example
> roster file (R37)" to 5d. 5b-ii's Task 4 created
> `config/roster.example.json` beside the roster loader (R-5b20),
> because the loader test is what keeps the example from rotting. The
> 5d row drops the example-file clause; its two `deploy.sh`
> pre-creates stay.

> Erratum, 2026-09-23 (5b-ii close): **the gate-rejoin sentence in
> `CLAUDE.md` moved.** The paragraph above that names the three in-tree
> copies of the rejoin sentence gives `CLAUDE.md:118`. The sentence is
> now at `CLAUDE.md:140`, in the Gate section. The other two pointers,
> `scripts/check.sh:5` and `Makefile:33`, still resolve; `Makefile:33`
> reads "The web legs rejoin," plural.

---

## The pass split

L band, **four passes**. Each pass's deliverable is testable without the
next, and three of the four change a contract, which is the shape INV's
budget evidence says overruns a ceiling.

| Pass | Scope | Exit gate | Tree |
| --- | --- | --- | --- |
| **5a** The serve shell and ingress | `internal/serve` born: the composition order with the strip (identity spellings, the forwarded-address group, and hop-by-hop removal), the four-rule normalization, the case-sensitive segment-aware prefix match, and the prefix terminator; the peer-address and reported-client-IP derivation (R27); the declared extension point; the empty fallthrough; the `http.Server` with its deadline policy and the `ResponseController` extension mechanism; `IP_FREEBIND` on the listener; the `serve` reserved class removed from `keys.go` and declared in rung 5's own file with the prefix and bind keys; `newServeCmd` wired into the `AddCommand` block. No auth, no pages. | `bash scripts/check.sh` green, plus: a request bearing every strip-set spelling in both dash and underscore form reaches a handler with none of them; a raw-versus-decoded mismatch, a `..` segment, an empty segment, and a backslash are each refused before routing; a path that string-prefixes the owned prefix without a segment boundary does not enter the owned surface, and `/Prefix` reaches the fallthrough; a `Connection`-listed header is removed on ingress; an unknown path under the prefix returns 404 and does not reach the fallthrough; a handler that writes past the default write deadline succeeds after extending it. | Worktree |
| **5b** The auth core and the mail seam | `internal/identity` and `internal/auth` born: token serialization with its labels and length floor, the `__Host-` session with rolling refresh under an absolute anchor, the create-exclusive nonce store and the epoch store under the state directory, the roster store with B9's validation, its admin flag, its atomic write path and its single resolution point (R1), the consumption-time roster and epoch re-checks, the three-key rate limiters with their bounded evicting map, the origin check with its bearer-credential bypass, the verify-identity seam with provenance, and the revocation seam as a read plus a notification. `internal/notify` born as a general send seam with the SMTP transport and its CRLF guard. The `auth and identity` reserved class removed and declared, the pending wall row retired into the selector key, the `internal/config/config.go` edit that requires and resolves it (R29), the correlated-threshold validations, and the three supersession-message rewrites. `web-auth-security-reviewer` runs inside this pass. | `check.sh` green, plus: a replayed nonce is refused across a process restart; a token whose MAC fails never reaches the nonce store; a session older than the epoch is refused, and so is a link minted before an epoch bump and a link for an address removed from the roster; refresh never moves the anchor and a session past the cap is refused; the limiter map's bound holds under a flood of distinct keys; N distinct addresses from one client address trip the per-IP link limit and send no further mail; a configuration whose nonce retention floor is below the link expiry fails to load; a request from a non-allowlisted peer carrying a forged client-address header is limited and logged under its peer address, and a verify subrequest from such a peer is refused; `doctor` exits nonzero and names the selector when `DUBPLATE_IDENTITY_SOURCE` is unset; a roster with a duplicate principal, a control character, or an unparseable address fails to load, and a missing file fails to load; a roster write whose record fails validation leaves the file byte-identical, a concurrent pair of writes both survive, and the store answers from the written file with no restart; a protected cookie-authenticated POST with neither `Origin` nor `Sec-Fetch-Site` is denied while a bypassed bearer path is not; a recipient or a subject carrying CRLF is refused by the seam. | Worktree |
| **5c** The pages, the `member` verb, and the gate leg. Fifteen deliverables. | `web/` born as a SvelteKit static app: the login page with its own string deck, the consumption interstitial, the admin-only add-member page (fork 1), the one composed theme authored as data (R36), the theme binding, the embedded woff2 faces with licenses and manifest, the `go:embed` path, the R4.5 rewrite, the generated branding stylesheet, the operator CSS file, and the served header set (CSP, `nosniff`, `Referrer-Policy`). The `dubplate-server member` verb group, `add` and `list`, wired into the `AddCommand` block over 5b's store (fork 1). The gate's `web/` leg rejoins with G1 through G10 and the server-file check, Node becomes a hard requirement, and the three rejoin sentences are corrected. `svelte-reviewer` and `daisyui-a11y-reviewer` run inside, and `web-auth-security-reviewer` returns for the admin page (R24). | `check.sh` green with the new leg, plus: G1 through G10 passing, each failing by name against a placeholder-only embedded tree; the served document carrying the composed `data-theme` value and the three stylesheets in tier order; the document served `no-cache`, the consumption route's response served `no-store`, and a hashed asset immutable; `script-src` carrying a hash for every inline script and no `'unsafe-inline'`; `nosniff` and `Referrer-Policy: no-referrer` on every served response; no log record emitted on the consumption route containing the token; a rejected theme match value failing config load with the accepted set named; a non-admin session receiving 404 from the add-member route while an admin session receives the page; an add-member POST without a same-origin signal denied; an add-member POST that succeeds writing one audit record naming the actor and the channel, and one that fails validation writing none and leaving the file unchanged; `member add --admin` against an absent roster creating it and refusing to create a second, and `member list` reading what the page wrote; all three pages rendered and read by the lane. | Worktree |
| **5d** The boundary, the deployment, and the close | The forward-auth verify endpoint and its integration class, with its own trust-boundary documentation; the trusted-proxy integration class; `doctor`'s probe rows and its bound-address row; the serve unit, its `ReadWritePaths` over both leaves, the two `deploy.sh` pre-creates and the example roster file (R37), and the generalized state-dir check; the env template's commented pending rows and its eight literal rows; the SEC and UNI disposition table; the full fresh-context adversarial cycle; the errata and the rung close. | `check.sh` green, plus: `systemd-analyze verify` clean on the new unit; the generalized state-dir check failing when either leaf is removed from `deploy.sh`; the unit starting with the roster leaf writable and failing with a named mount-namespacing error when its pre-create is removed; `doctor` reporting the live client-IP mode, the bound addresses, and a boundary-probe row that dials the running listener and reports plainly when none answers; `deploy.sh --check` rendering with no unresolved name and `dubplate-server doctor` exiting 0 against the rendered environment; the disposition table covering all 50 review findings; the adversarial cycle's findings folded or refused with reasons. | Worktree, released at the close |

**The split point that matters is between 5a and 5b.** The shell's
composition order is the artifact the whole trust boundary rests on, and
reviewing it in the same pass as the token and session mechanics would
bury it. PA:79 already says as much from the other side: building the
ingress strip as an unplanned chore would strip the adversarial cycle
from that single artifact.

**If 5b's plan exceeds fifteen tasks, split the pass rather than the
tasks.** The cut point is between the token, session, and state records
on one side and the roster, limiters, and seam on the other, which is
where the risk changes from cryptographic construction to input
validation. The conductor takes that call at 5b's plan authoring, names
the follow-on 5b-ii, and re-derives both ceilings.

### Contested surfaces

The conductor writes these into the parallel-approach spec's ownership
table before the lane launches, per lane-launch checklist item 2.

| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/serve` shell, mux, ingress strip (amends the existing row) | rung 5, created by it; rung 12 thereafter, with no concurrent serve-writing lane | 6, 7, 12, 14 add files at the declared extension point |
| `internal/auth` (new) | rung 5, created by it; no other lane writes it | rung 7 and rung 14 consume the seam; rung 12 consumes it through the seam only |
| `internal/identity` (new) | rung 5, created by it | rung 6's ops request types, rung 12's driver, 2.0's credential store |
| `internal/notify` (amends the existing row) | rung 5, as a general send seam rather than a magic-link mailer, adopted from `_salvage/internal/notify` on merit | rung 7's reject copy, rung 9's monitoring integration |
| The roster store, its interface, and its write path | rung 5 | rung 7 reads it; rung 5's own `member` verb and admin page write it; rung 6 may later back it with its store behind the same interface |
| `web/` (new tree) | rung 5 births it with three pages, the admin page among them; rung 7 authors visuals and the theme roster and owns the attended before/after | rung 7, rung 12's pairing, 2.0's device page |
| The embedded static asset path, the woff2 faces, and the branding injection point | rung 5 | rung 7 embeds its bundle and the operator file through the same path |
| `internal/config/keys.go` edits | rung 5 removes the `auth and identity` and `serve` reserved classes and the pending identity-source wall row, each in the commit that declares its replacement elsewhere, and rewrites the three supersession messages its own keys replace | every later rung, by the same rule |
| `internal/config/config.go`, `Load`'s owned-row resolution | rung 1's contract; rung 5 adds the identity-source require and the per-source correlated resolution (R29) | every rung that adds a required owned row |
| `.gitleaksignore` | rung 1b retired its one fingerprint; the lane that needs an allowed finding appends and names it in its pass report (RE cell (5,2)) | every lane; a concurrent append is a non-fast-forward rebase and triggers the merge ritual's re-review |
| `cmd/dubplate-server` root command and `doctor` rows (amends the existing row) | rung 1's skeleton; rung 5 adds `serve`, fork 1's `member` verb group, and the probe row kind | 6, 9, 13 |
| The state directory layout (amends the existing row) | rung 4 owns the layout; rung 5 adds two leaves, the auth records' and fork 1's roster leaf, and their permissions policy | rung 6's store supersedes neither record |
| `env/dubplate-server.env.template` (amends the existing row) | rung 4 by its own header; rung 5 adds commented pending rows only | rung 16 mints |
| `scripts/check.sh` | rung 5 adds the `web/` leg, the Node requirement, and the generalized state-dir check | every lane, and after 5c every full gate builds the bundle |
| `CLAUDE.md` and `Makefile` rejoin sentences | rung 5's pass 5c | every reader |

The conductor also owes the lane registry entry, the sub-ceiling against
the phase ceiling, the runaway guard, and the `docs/STATUS.md`
lane-registry update.

One lane-launch item joins the checklist at 5c. `git worktree add`
checks out tracked files only, which is why item 5 provisions a venv,
and `web/node_modules` is the same shape: any tree that runs the full
gate after 5c installs the web dependencies first. Prefer fixtures that
are obviously not credentials in the auth packages, since an allowlist
entry there is the one a later reader is least able to re-audit.

### Band and ceiling

INV's budget table is the evidence. Every rung 3 pass whose spend
included plan authoring closed between 2.84M and 5.15M, plan preparation
was the largest single line in each, and the one pass that closed at
2.7M carried prep of 0.26M because its plan was already written. Rung 1b
overran its 2.5M ceiling at 3.7M, and its own entry records the 2.0M
flag crossing inside the chain with no conductor action.

**Propose 3M for 5a, 4M for 5b, 4.5M for 5c, and 3M for 5d**, 14.5M in
total, against rung 3's four-pass actual of about 17.7M. The estimate is
lower because rung 5 has no corpus, no oracle, and no external API
budget. It is higher per pass than rung 4's because three of the four
change a contract: the shell's extension point at 5a, the
verify-identity seam and the roster's source at 5b, and the gate's own
leg at 5c. Each ceiling is re-derived from the previous pass's actual
report before the next plan is drafted.

**5c's ceiling is re-derived here, because fork 1 grows it.** The
arithmetic, from INV's budget table:

- Revision 2 set 5c at 4M for thirteen deliverables.
- Plan preparation does not scale with deliverable count. It was 2.55M
  at 3c and 2.51M at 3d, so about 2.5M of any pass's spend is fixed
  prep.
- That leaves about 1.5M of chain across thirteen deliverables, or about
  0.115M each.
- Fork 1 adds two: the admin-only add-member page, and the
  `dubplate-server member` verb group. Two at 0.115M is 0.23M.
- R24 adds a second `web-auth-security-reviewer` read at 5c, over a
  cookie-authenticated privilege-granting write. Price it at one
  reviewer pass, 0.25M.
- 4M plus 0.23M plus 0.25M is 4.48M, which rounds on this build's 0.5M
  grain to **4.5M**. Rung 5's total moves from 14M to 14.5M.

5b's count and ceiling do not move. Its roster deliverable grows a write
path rather than gaining a deliverable, and 1b's overrun is the standing
reason to watch it rather than to widen it.

The split does not move in revision 2 or 2.1, and only 5c's figure does.
The folded work sits inside the passes that already owned its surfaces:
5a takes three more normalization rules and the hop-by-hop strip, 5c
takes one composed theme, one string deck, and fork 1's two new
deliverables, 5d takes the roster leaf and a live boundary probe. 5b is
the exception worth naming, because it absorbs the
`internal/config/config.go` contract change (R29), the consumption
re-checks, the third limiter key, and the correlated-threshold
validations. It was already the largest pass with the densest security
surface, so its named cut point is more likely to bind than it was at
revision 1, and the conductor states its deliverable count at dispatch.
One per-gate cost is new rather than moved: after 5c every full gate
builds the web bundle, which every lane pays.

Two scheduling notes for every plan header. The second code slot passes
at a pass close rather than a rung close, so lane A releases the
worktree at each of these four merges. And every task boundary is
planned as a resumable stopping point, because pass 8a's chain met a
session window cap before its own ceiling would have bound.

---

## The staleness pre-task

Every rung 5 pass opens with a staleness pre-task before its first
implementer dispatch. One fresh-context agent reads this spec and the
pass plan against the tree as shipped and reports blockers with evidence
rather than fixes. The conductor rules each before the first task
dispatches. Plan staleness is the build's named dominant defect class.

The first is a blocker; a "no" stops the launch.

- **Is the lane registry consistent with where this pass intends to run,
  and is the target tree free?** The live-executor check is the
  registry, never `pgrep -f Projects/dubplate`, which matches every lane
  once a worktree exists.
- **Has any other lane written `internal/serve`?** PA:80 forbids it
  until rung 5 closes, and a violation invalidates 5a's whole design.
- **Does `internal/config/keys.go` still carry the `auth and identity`
  and `serve` reserved classes and the `{Name: "identity source",
  Pending: ...}` wall row, at the shapes this spec cites?** Rung 5
  removes all three.
- **Is `DUBPLATE_CONTRIBUTORS_FILE`'s supersession message still the one
  naming rung 6's store?** R1 rewrites it.
- **Does the registry still expose `AddWallRow`, `AddReservedClass`,
  `AddIntegrationClass`, `AddThreshold`, and `KindString`, with
  `IntegrationClass.Enabled` still parsing a boolean flag?** The
  selector design in the config section turns on that last fact.
- **Is `cmd/dubplate-server/main.go`'s single `AddCommand` block still
  the wiring point, and does `doctor` still derive every row from the
  registry?** The probe row kind is an addition to that shape.
- **Is the state-directory agreement check still hard-coded to
  `systemd/dubplate-shadow.service`, and are the three literals still
  the template, the unit, and `deploy.sh`?** R12 generalizes it.
- **Do `CLAUDE.md:118`, `scripts/check.sh:5`, and `Makefile:33-34` still
  carry the "rejoins at rung 7" sentence?** Pass 5c corrects all three.
- **Has the design-language standard changed since revision 2**, in
  particular R2.1's 32 names, R4.5's rewrite, R5.6's directive set, or
  the ten gates?
- **Is `_salvage/internal/notify` unchanged**, and does
  `_salvage/internal/serve/serve.go` still carry `requireSameOrigin`,
  `mountBasePath`, `staticHandler`, and the single write timeout this
  spec reasons from?
- **Has `internal/atomicfile` gained the streaming write rung 8 was to
  add**, and is `WriteFile` still rename-over? R11 turns on the second
  half.
- **Does `env/dubplate-server.env.template` still carry the three row
  states in its header, and does `deploy.sh` still abort the render on
  one unresolved name?**
- **Does `Load` still hard-code its four owned requires, and does
  `Config` still carry no generic wall accessor?** R29 is the edit that
  makes the selector real, and its shape depends on both facts.
- **Do `doctor`'s tests still pin the output's line count**, at
  `doctor_test.go:129-130` and `TestDoctorCmd_LineCountFollowsRegistry`?
  R6's probe rows and the per-source rows meet both.
- **Does the design-language standard still ship no default theme block,
  and does DL section 7 still give the roster to rung 7?** R36 authors
  the block rung 5's pages render against, and revision 1 assumed 0c had
  shipped one.
- **Does `scripts/deploy.sh` still deliver `config/contributors.yaml` to
  `/etc/dubplate/contributors.yaml`** for the incumbent importer? R37
  leaves that delivery alone and puts rung 5's roster in the state
  directory instead, so a change here is a signal rather than a blocker.
- **Is `.gitleaksignore` still the tracked empty file rung 1b left**,
  and has any lane appended to it?

---

## Risks

**1. The composition order is right in the route table and wrong in the
chain.** A normalization check registered as a route runs after
`http.ServeMux` has already cleaned and redirected, and the mistake is
invisible in review. Mitigation: 5a's exit gate asserts the order
through behavior rather than through structure, with one test per
property, and the shell exposes no way to register a check as a route.

**2. A strip set that misses a spelling.** The underscore class is the
one the field reports as biting adopters, and Go's header
canonicalization keeps the two spellings distinct. Mitigation: the
extras list is data with both spellings generated from one source, the
test enumerates the whole set, and the driver's contribution goes
through the same data path rather than through rung 12 editing rung 5's
code.

**3. A nonce store that forgives a replay.** An atomic rename overwrites
silently, and a store that lost the exclusivity would pass every test
that checks only "the record exists". Mitigation: R11's create-exclusive
insert, a test that consumes the same link twice across a simulated
restart, and the fixed verify-then-insert order.

**4. Session mechanics that look right and are not.** A refresh that
moves the anchor silently removes the absolute cap; an epoch check keyed
on a non-normalized address silently fails to revoke. Mitigation: both
are exit-gate assertions at 5b, with hard-coded expectations rather than
expectations derived from the code under test.

**5. The roster fails open.** The salvage loader does exactly this.
Mitigation: R2, with the missing-file case as a named test rather than
an implied one.

**6. The CSP blocks the page's own bootstrap.** DL revision 1's
inherited CSP would have let rung 5's login page load and never boot,
which is the blocker both design-language reviews filed. Mitigation:
DL's G6 asserts a hash per inline script and no `'unsafe-inline'` in
`script-src`, and the lane renders and reads both pages rather than
inferring from a green test.

**7. The gate leg silently stops checking.** Five of DL's ten gates read
a built bundle, and a `command -v node || skip` would turn a missing
tool into a green build. Mitigation: R13's hard requirement, plus each
bundle-reading gate failing by name against a placeholder-only tree.

**8. Scope accretion in 5b.** It is the largest pass, it carries the
densest security surface, and every addition to it is defensible alone.
Mitigation: the named cut point and the 5b-ii instruction above, plus
the conductor stating the pass's deliverable count at dispatch.

**9. A ratified document contradicts this spec.** Three already do on
the roster's owner, one does on the page stack, and one does on proxy
hygiene's rung. Mitigation: the errata section, written at the close
into the named documents, and the staleness pre-task's question about
each.

**10. The first production deploy of a user-facing surface.** The box is
production and the serve unit is new. Mitigation: 5d's deploy runs
`deploy.sh --check` first, waits for a quiet-box check, and is named as
an owner touchpoint rather than assumed. The unit is deployed and left
disabled until the cutover rung, on the incumbent unit's own precedent.

**11. A required key nothing requires.** The registry claims a name and
creates no requirement, so a selector declared and never wired reads as
done in every review and reports green in `doctor`. Revision 1 shipped
exactly that design. Mitigation: R29's `Load` edit, 5b's exit-gate
assertion that `doctor` exits nonzero on the unset selector, and the
staleness question about `Load`'s four hard-coded requires.

**12. A gate that reads a theme file nobody wrote.** G2, G9, and G10
read theme data, and DL's empty-iteration guard covers G3 through G7
alone. A rung 5 that authors no block would pass three gates on empty
iterations. Mitigation: R36 authors one composed theme as data, and 5c's
exit gate names the gates rather than asserting the leg is green.

**13. A roster write that loses a member or admits one.** Fork 1 puts a
writer in front of the file every trust decision reads. A rewrite that
races another rewrite drops a member silently; validation run outside
the write's lock lets two colliding principals both pass; a rename that
never reaches the directory loses an add an admin watched succeed.
Mitigation: one lock covering validate, write, and re-read (R1),
`WriteStream` rather than `WriteFile` (R11), and 5c's exit-gate
assertions that a concurrent pair of writes both survive and that a
failed write leaves the file byte-identical.

**14. The admin check passes for a member who is not one.** Rung 5's one
privileged action is authorized by a flag on a roster record, and a
check reading that flag from the session rather than from the roster
would grant it to anyone who could shape a cookie. Mitigation: the flag
is read from the roster at request time and never carried in the cookie,
the route answers 404 to a non-admin so its existence leaks nothing, and
R24 sends `web-auth-security-reviewer` back over the page at 5c.

---

## Process

Each pass gets its own plan through writing-plans, drafted just in time
in the shadow of the previous pass's execution and never front-loaded.
Plans carry a token ceiling and a checkpoint interval, and state per
task whether the task lands a 2.0 hook or is neutral. Execution runs the
implementer, `diff-reviewer`, gate chain per task, in workflow mode at
six or more tasks. The gate is `bash scripts/check.sh`, green at every
task end, specific files only, never commit red. Both budgets are scored
at every pass close.

`go-conventions` before any Go. `svelte-conventions` and
`ts-conventions` before any component or TypeScript. `vps-conventions`
before the unit and the deploy edits at 5d. Subagents cannot dispatch
agents, so the conductor owns the `code-simplifier` dispatch at each
pass close.

**The adversarial cycle.** Rung 5 keeps the full fresh-context cycle.
Its shape:

- Inside 5b, `web-auth-security-reviewer` reads the auth core, because
  the ladder entry says the gate runs inside the rung rather than at the
  pass-end fan-out.
- Inside 5c, `svelte-reviewer` and `daisyui-a11y-reviewer` read `web/`.
- At 5d, at least two independent fresh-context lenses read the whole
  subsystem, one security and one universality, mirroring the pair that
  produced SEC and UNI. The universality lens is not optional: it caught
  three criticals the security lens did not, including the identity
  fusion that produced B9.
- Both lenses receive this spec's threat-model section and the SEC and
  UNI disposition table as their input, and they grade the
  implementation against the claims rather than against taste.
- Every finding is folded or refused with its reason recorded in the
  merged review document, which becomes the do-not-re-litigate list for
  every later rung 5 pass. REV is the first such document, and its
  "Verified sound" list already carries that force: a pass plan does not
  re-open an item on it without new evidence.

The repo's precedent for this cycle is the U13 round, which ran five
reviewers with adversarial verification and produced 45 confirmed
serious findings, including one no gate could have caught.

**Owner touchpoints rung 5 expects, and only these.** The six product
forks, ruled 2026-09-08 and closed. The rung 5 sitting DL routes the
`style-src 'unsafe-inline'` note to, which is the one remaining item
from that batch. The 5d box deploy, which waits for a quiet-box check.
Anything else is a conductor ruling.

---

## Universality

Nothing in this rung is site-specific. No hostname, path, email, or
personal name lives outside `internal/config`, and fork 1 tightens that:
the login page's contact copy derives from the roster's admin entry, so
no address sits in configuration at all. The repo's example roster
carries invented names, and the estate's own roster is written by the
CLI verb on the box rather than committed.

Drivers own driver knowledge. The header name, the injection rules, the
bypass paths, the logout wiring, and the post-login destination live
with the driver at rung 12, and rung 5 ships the mechanism by which a
driver contributes them rather than any driver's answer.

Every tunable is a threshold with a default: session lifetime, the
absolute cap, link expiry, the rate limits, the enumeration behavior,
the owned prefix, and the theme picks.

The design makes sense with no Cloudflare account. Plain SMTP submission
plus any streamer, with the tunnel optional and a forward-auth proxy
optional in front. One precondition rides with that claim, and revision
1 omitted it: the deployment is reached over HTTPS. The session cookie
carries the `__Host-` prefix and `Secure`, which a browser refuses over
plain HTTP on any origin but localhost, so an adopter serving over HTTP
on a LAN would get a login that completes and signs nobody in. The
terminator stays the adopter's choice, a tunnel or their own proxy, and
config load plus a `doctor` row surface the mistake at configuration
time (R38). Real-client-IP handling is configured rather than assumed.
The estate's own mail transport is one configuration choice among the
supported ones, reversible in a deploy, and the runbook carries two
worked examples so an adopter with no Cloudflare account and an adopter
with one both have a path.
