# Rung 5 Spec Inventory

Research report, 2026-09-08, prepared for the rung 5 spec's authoring,
against HEAD `8ff65bd99f982bda940db19420ee80f834fcf9d1`. Lane-1 was
running the pass 8a chain in this same tree while this was written, and
8a closed before it was finished, so `internal/engine/tagio`,
`docs/plans/2026-09-07-rung-8a-tagio-writer.md`, and
`docs/specs/2026-09-07-rung-8-tag-io.md` were moving
underneath it. Nothing below reads any of them, and no gate was run.
The `docs/HISTORY.md` citations are against the file as it stands after
8a's close entry landed, which shifted every earlier entry by 408 lines.
Every other citation is a line read at the HEAD above.

Scope, quoted from the build spec's ladder entry
(`docs/specs/2026-09-02-greenfield-server-build.md:619-627`):

> **Identity and auth** (L). `internal/auth` built to the birth
> requirements: tokens, sessions with epoch and absolute cap, durable
> nonces, the login and consumption pages (Go-templated, no JavaScript,
> shared design language), SMTP send, rate limiting, the CSRF rule, the
> verify-identity seam with provenance, the forward-auth verify endpoint
> with its own trust-boundary documentation, and a minimal serve shell
> hosting it under the one owned prefix. The `web-auth-security-reviewer`
> gate runs inside the rung.

One clause of that entry is already superseded. "Go-templated, no
JavaScript" is overturned by the Phase 0c standard's R1.6
(`docs/specs/2026-09-07-design-language-standard.md:88-100`),
which rules the login and consumption pages as SvelteKit static pages
rendered client-side and served by the Go binary, on CLAUDE.md's later
web rule (`CLAUDE.md:135-143`). The close of rung 5 records the
build-spec correction. Section 7 carries what that swap costs.

Short names used below, matching the design-language standard's own
table so citations read the same across both documents:

| Short name | File |
|---|---|
| BS | `docs/specs/2026-09-02-greenfield-server-build.md` |
| MD | `docs/specs/2026-09-02-magic-link-login-design.md` |
| PA | `docs/specs/2026-09-03-parallel-build-approach.md` |
| RE | `docs/research/2026-09-04-reverse-edge-completion.md` |
| DL | `docs/specs/2026-09-07-design-language-standard.md` |
| S20 | `docs/specs/2026-09-07-streamer-2-0-design.md` |
| SEC | `docs/research/2026-09-02-login-spec-reviews/security.md` |
| UNI | `docs/research/2026-09-02-login-spec-reviews/universality.md` |

Sections 1 through 8 are the as-landed state of everything the spec
touches, measured against the tree. Section 9 is the review shape.
Section 10 is budget evidence and a pass-split proposal. The document
closes with surprises and candidate rulings.

Every ruling this inventory raises is applied as its recommendation and
marked **provisional, conductor ruling owed**. A ruling that changes
what an adopter or a family member sees, configures, or is asked for is
marked **product fork** and repeated in the closing section for Geoff.

---

## 1. The birth requirements and the magic-link design, clause by clause

### 1.1 The twelve birth-requirement clauses

BS:393-523 is the authority. Its opening paragraph (BS:394-404) states
the standing that governs everything below: the login spec is source
material, its build-on-old-serve framing and its one-compact-pass build
order are superseded, and both adversarial reviews are birth
requirements the rung plan walks as checklists, minors included.

| # | Clause | BS lines | Status against the tree |
|---|---|---|---|
| B1 | The trust boundary, serve-as-ingress | 405-415 | Open. No `internal/serve` exists. |
| B2 | The trust boundary, forward-auth | 416-427 | Open, and it needs an active `doctor` that does not exist (see 1.4). |
| B3 | One owned prefix | 428-432 | Open, and no config key names the prefix (see section 3). |
| B4 | Link consumption | 433-440 | Open. Overturns MD:96-104 on nonce durability. |
| B5 | Sessions | 441-463 | Open. Overturns MD:106-122 on revocation and cookie shape. |
| B6 | Secrets | 464-471 | Names reserved, values unminted (see section 3). |
| B7 | CSRF, stated once | 472-477 | Open. Salvage carries a proven implementation. |
| B8 | Rate limits and abuse | 478-490 | Open. Overturns MD:88-95 on enumeration. |
| B9 | Identity and roster | 491-498 | **Contradicted by the tree** (see section 5). |
| B10 | One identity source | 499-503 | Half-landed: the wall carries the pending row. |
| B11 | Post-login destination | 504-511 | Open, and it belongs to a driver rung 5 does not build. |
| B12 | Proxy hygiene | 512-523 | Marked "(streamer rung)" and half shell-level (see section 2). |

### 1.2 The clauses the tree already answers

**B10's pending row landed at rung 1 and survived rung 1b.**
`internal/config/keys.go:19-25` carries `ownedWall` with
`{Name: "identity source", Pending: "pending (arrives at rung 5)"}` as
its fifth row, and `ownedNow` (`keys.go:32-43`) is what keeps `Load`
from requiring it. `doctor` prints the pending status verbatim
(`cmd/dubplate-server/doctor.go:33-38`, `cmp.Or(row.Pending, ...)`).
Rung 5 is the rung that clears the `Pending` field, and the moment it
does, `Load` starts requiring a row named `identity source`. That name
is a status label rather than an environment key, which R16's carve-out
exempts from the `DUBPLATE_` prefix check
(`internal/config/registry.go:137`). The spec must state what clearing
it means in mechanical terms, because a bare clear would make `Load`
demand an environment variable literally spelled `identity source`. The
plausible shape is that the row's `Name` changes to the selector key
rung 5 declares, and the pending row retires with it.

**B6's names are reserved and unclaimed.** `keys.go:60-73` holds the
`auth and identity` reserved class at rung 5 with eight keys. Setting
any of them today is a load error naming the class and the rung
(`ReservedClass`, `keys.go:45-53`).

**B7 has a proven implementation in salvage.**
`_salvage/internal/serve/serve.go:130-177` is `requireSameOrigin`, whose
doc comment records the ordering discovery that matters: `Sec-Fetch-Site`
is checked first and honored on its own, with `Origin` compared against
`r.Host` as the fallback, because an `Origin`-only check 403s every dev
request through a proxy that rewrites `Host`. A request bearing neither
header is refused. That is exactly B7's blanket rule, already written
once, and SEC:36 records the defect the greenfield must not repeat: the
salvage check sat inside the `/upload` mount and covered no new surface.

### 1.3 The clauses the tree contradicts

**B9 against `keys.go:176-177`.** The birth requirement says the roster
validates at load, with unique names, `net/mail.ParseAddress`, no
control characters, bounded length, and an optional `streamer_username`
with startup and `doctor` collision checks. The supersession map says
`DUBPLATE_CONTRIBUTORS_FILE` is "retired: the roster loads from the
store (rung 6), not a config-supplied file". RE:(5,6) says rung 5 needs
no store, "by deliberate design". Rung 5 therefore has a roster
requirement, a validation contract, and no source. Section 5 carries the
whole finding; it is the sharpest single item in this inventory.

**BS:621-622 against DL R1.6.** Covered in the header. The build spec's
ladder entry still says Go-templated with no JavaScript.

**BS:638 and its two siblings against DL section 6.** The build spec
puts the `web/` gate leg's rejoin at rung 7 (BS:646-647, inside ladder
entry 7). DL section 6 (`DL:513-517`) rules the leg rejoins at rung 5
and names two copies to correct: CLAUDE.md's line and `scripts/check.sh`'s
header. There is a third. `Makefile:33` carries "The web legs rejoin at
rung 7, when web/ is rebuilt against the greenfield server" directly
above the `check` target. The three live copies are `CLAUDE.md:118`,
`scripts/check.sh:5`, and `Makefile:33`, plus the build spec's own
ladder text, which takes a dated erratum rather than an edit.

### 1.4 The clauses that need a mechanism nothing in the tree has

**B2 asks `doctor` to prove a boundary.** BS:423-425: "`dubplate doctor`
proves the boundary by verifying that a spoofed identity header from a
non-trusted source is rejected." B8 asks for a second active report
(BS:479-481): the client-IP mode is explicit configuration, "and
`doctor` reports which is live." Today `doctor` is a pure config-wall
reporter. `newDoctorCmd` (`cmd/dubplate-server/doctor.go:23-71`) walks
four registry families, prints one `%-32s %s` line each, and returns a
single short error when `config.Load` is unsatisfied. It makes no
network call, runs no probe, and has no vocabulary for a check that
passes or fails on evidence rather than on configuration. Rung 5 either
introduces an active-check row kind into `doctor` or states plainly that
B2's proof is a separate verb. This is a design decision the ladder
entry does not name and the reverse-edge study did not surface, because
RE:(5,4) described rung 5's doctor additions as "rows", which is the
vocabulary that already exists.

**B12's own header is misleading.** BS:512 marks proxy hygiene
"(streamer rung)", and PA:77 corrects it: the header-strip set, the path
normalization ahead of prefix routing, and the per-route write timeouts
"are all shell-level" and therefore rung 5's. Only the
`httputil.ReverseProxy` half is rung 12's. Section 2 splits the clause.

### 1.5 The magic-link design, read for intent

MD carries a SUPERSEDED banner (MD:3-9) naming six points where the
birth requirements win. Read as intent, its sections divide cleanly into
what rung 5 inherits and what it must not.

| MD section | Lines | Disposition |
|---|---|---|
| What this builds | 18-25 | Intent, intact. One branded email login for the whole estate. |
| Universality constraint | 27-43 | Intact and promoted: BS:148-164 restates the four tests. |
| Topology | 45-80 | Intact as topology. Routing rule 2's `/api/*` is a review blocker (SEC:20, UNI:7). |
| Login flow | 82-104 | Intent intact; nonce durability overturned by B4. |
| Sessions | 106-122 | Cookie shape and revocation overturned by B5. |
| Streamer SSO mapping | 124-139 | Wire identity overturned by B9. Section 8 carries the rest. |
| Security requirements | 141-154 | Absorbed into B1, B7, B8, and the review gate. |
| Config | 156-167 | Absorbed into the reserved classes; the Access rows are already superseded (`keys.go:169-174`). |
| Package shape | 169-176 | Intact: `internal/auth` plus `internal/serve` consuming it through the seam. |
| Build order | 178-190 | Overturned. "One compact pass" is an L-band rung at multiple passes (BS:536-538, UNI:61). |
| Amendments | 192-202 | Intact as amendments against the organizer spec. |
| Out of scope | 204-210 | Half overturned: per-user revocation is now a birth requirement (B5). |

The single most useful thing MD still carries is its availability note
(MD:76-80), which prices the SSO proxy honestly: serve in the streaming
path means serve down is the web player down, while `/rest/*` survives
as long as serve fronts the host at all. Rung 5's spec should carry that
sentence forward, because the capability it describes is optional and
the rung that builds it is 12.

### 1.6 What landed since either document was written

Five things, each of which changes an assumption one of the two
documents made.

The **registry** (rung 1b, closed 2026-09-08, `docs/HISTORY.md:413-500`).
The config wall stopped being a closed table. A rung declares its keys
in its own file, and "a lane removes only its own reserved class from
that file, in the same commit that declares the class in its own
package (rung 5 is first, taking `auth and identity` and `serve`)"
(`HISTORY:494-497`). Section 3 carries the mechanics.

The **design-language standard** (Phase 0c, DL, revision 2). It binds
rung 5's pages by name at DL:61-65 and hands rung 5 four concrete build
obligations: stamp the composed `data-theme` (R3.6, DL:310-317), produce
the served document through one documented rewrite (R4.5, DL:400-411),
generate the tier 2 branding stylesheet from config (R4.3, DL:371-392),
and serve the CSP directive set (R5.6, DL:470-500).

The **SvelteKit static rule** (`CLAUDE.md:135-143`, ruled as M11 at
S20:449-452). Rung 5's login page is the first page built under it.

**Pass 4b's env template and units.**
`env/dubplate-server.env.template` now exists with the three row states
written into its own header (`:8-16`), and
`systemd/dubplate-shadow.service` is the greenfield unit precedent.
Section 4 carries both.

**The state-directory layout.** `/var/lib/dubplate` is the greenfield
state directory, deliberately separate from the incumbent's
`/var/lib/musicbox` (template `:38-53`), with `shadow` as its first leaf.

---

## 2. The serve shell

### 2.1 What PA requires the spec to state

PA:79 is the requirement, quoted in full because the spec has to
discharge it clause by clause:

> Rung 5 owns the serve shell, the mux, and the ingress strip, and its
> spec must name the root-level composition order (ingress strip, then
> path normalization, then prefix routing, then fallthrough), declare the
> extension point every later consumer adds a file at, and state which of
> those points the driver's proxy occupies.

PA:80 adds the scheduling rule: no two lanes write `internal/serve`
until rung 5 closes, and rung 12 owns the shell thereafter with no
concurrent serve-writing lane. PA:178 awards the shell, mux, and ingress
strip to rung 5 with rungs 6, 7, 12, and 14 named as consumers who "add
files at the declared extension point".

Three consequences the spec must make explicit.

**The composition order is a middleware chain, not a mux feature.** The
strip runs before routing "for every path, the `/rest/*` and `/share/*`
bypasses included" (BS:405-408), so it wraps the mux rather than living
in any handler. Path normalization runs next and also wraps the mux,
because `http.ServeMux` does its own path cleaning and redirect
internally; a normalization check registered as a route would run after
the decision it is meant to guard. The spec should state that ordering
as a property with a test, since it is invisible in the route table.

**The extension point is what rungs 6, 7, 12, and 14 will each add a
file at.** No code exists to constrain its shape. `internal/` today holds
`atomicfile`, `config`, `engine`, and `harness` only. The candidate
shape that satisfies both PA:79 and the one-owned-prefix rule is a
registration seam under the owned prefix plus exactly one root-level
fallthrough slot the driver claims. `Mount(prefix, http.Handler)` was
already examined and rejected at PA:77 for the driver's case.

**The driver's proxy occupies the fallthrough, not a mount.** PA:77:
"rung 12 is the ingress reverse proxy and by the one-owned-prefix rule
the streamer owns everything else, so it is the root mux's fallthrough
rather than a mounted subtree." RE:(14,12) adds the constraint from the
other side: rung 14 must terminate the owned prefix so the later
fallthrough cannot capture an unknown API path. Rung 5 builds the
terminator, since it owns the mux.

### 2.2 The header-strip set

BS:405-410 defines the set as "the driver's configured header name plus
a fixed extras list, underscore variants included". At rung 5 no driver
exists, so the configured half is empty and the fixed half is all rung 5
can ship. SEC:42 is the review finding behind the underscore clause:
header spellings to strip are not enumerated, "and the underscore class
is the one that bites adopters". The spec owes an enumerated default
list and a stated rule for how a driver contributes its own name later,
without rung 12 editing rung 5's strip code.

A second strip target is rung 5's own: BS:441-443 requires the session
cookie be stripped before proxying to the streamer, and SEC:63 filed the
salvage design's failure to do it as a major.

### 2.3 Path normalization

BS:514-516: "path normalization rejects raw/decoded mismatches and `..`
segments before prefix routing". SEC:33 is the finding: a non-bypass
path can be smuggled through a bypass prefix. The salvage tree carries
the same defect class already proven once at a different layer.
`mountBasePath` (`_salvage/internal/serve/serve.go:82-95`) documents
that `http.StripPrefix` does a plain string trim rather than a
segment-aware one, so an unguarded `/uploadfoo` would trim to `foo` and
pass through. That is the identical bug shape one level down from the
one B12 names, and it is worth citing in the spec as evidence that the
class is real in this codebase rather than hypothetical.

### 2.4 Per-route write timeouts

The salvage server sets one global write timeout:
`_salvage/internal/serve/serve.go:599-602` declares
`httpReadHeaderTimeout` 10s, `httpWriteTimeout` 30s, and
`httpIdleTimeout` 120s, applied to a single `http.Server` at `:628-633`.
A 30s global write timeout kills an SSE stream and any audio response
longer than 30 seconds, which is why B12 asks for per-route handling.
The stdlib mechanism is `http.ResponseController`'s `SetWriteDeadline`,
which lets a streaming handler extend or clear its own deadline while
the server's default protects everything else. The spec should name the
mechanism and assign the default, and the plan should verify the
behavior against a live stream rather than against the constant.

Rung 5 has no SSE and no audio route of its own. The requirement is
nonetheless rung 5's, because the deadline policy is a property of the
`http.Server` the shell constructs, and rung 12 cannot change it without
writing the shell.

### 2.5 Wiring a serve verb into the binary

`cmd/dubplate-server/main.go:63-80` is `newRootCmd`. It builds the
registry once, visibly (`reg := config.NewRegistry()` at `:64`), calls
`registerShadowIntegration(reg)` and `registerShadowThresholds(reg)` at
`:65-66`, and passes `reg` into the three subcommands that enforce
against it in one `AddCommand` block at `:74-79`. The doc comment at
`:1-20` carries the subcommand surface with a rung per verb, and `:12`
already lists "serve: run the HTTP server (rung 5)". Rung 5 adds
`newServeCmd(...)` to the block, its own `registerAuth*` and
`registerServe*` calls above it, and nothing else. RE:(5,4) predicted
contention here with rung 4; rung 4's shadow verb has landed, so the
contention is spent.

`doctor` needs no change to display rung 5's new classes. It derives
every row from the registry (`doctor.go:33-61`), and `rowStatuses`
(`:83-105`) parses `config.Load`'s joined error into per-key text with
no per-rung knowledge. The active checks of section 1.4 are the only
`doctor` work rung 5 owes, and they do not fit the existing row shape.

### 2.6 The serve unit

`systemd/dubplate-serve.service` is the incumbent's unit, not a
greenfield one. It execs `/usr/local/bin/dubplate serve` (`:52`) and
reads `EnvironmentFile=/etc/dubplate/env` (`:51`), both of which are the
pre-pivot binary and manifest. `scripts/deploy.sh:59-66` still ships it
in `UNIT_FILES`, and `deploy.sh:53` records that it is deployed but
deliberately not enabled. Rung 5's unit is a new file against
`/usr/local/bin/dubplate-server serve` and `/etc/dubplate/server.env`,
the pair `systemd/dubplate-shadow.service:20-21` established.

What a greenfield serve unit inherits from the incumbent, each with the
line that argues it:

- The whole hardening block, `:99-125`, which
  `dubplate-shadow.service:61-79` already carried forward verbatim minus
  `AF_NETLINK`. A greenfield serve unit runs no CPython subprocess, so
  it drops `AF_NETLINK` for the same reason the shadow unit did
  (`dubplate-shadow.service:80-85`).
- The restart policy and its 2026-09-01 ruling, `:36-46`. A literal
  `StartLimitIntervalSec=0` means the unit never enters failed, so
  `OnFailure=` never fires and a crash-looping user-facing surface
  alerts nobody. 300s over 30 restarts is the landed answer.
- `TimeoutStopSec=15` (`:57-63`), argued from a client mid-PATCH. Rung 5
  has no upload, so the value needs its own reasoning rather than a copy.
- The `OnFailure=dubplate-ping-fail@<slug>.service` idiom (`:28-35`),
  wired and inert. `dubplate-shadow.service:48-59` records precisely why
  it stays inert on the greenfield side: the whole `DUBPLATE_PING_*`
  class is reserved for rung 9 (`keys.go:118-125`), and the ping helper
  reads the incumbent's environment file and execs the frozen incumbent
  binary. A greenfield serve unit inherits the same inertness.
- The network-online question, answered in the negative at `:12-26`. The
  incumbent takes `After=docker.service` rather than
  `network-online.target` because its listener binds the compose bridge
  gateway IP, with `IP_FREEBIND` set through
  `net.ListenConfig.Control` for boot resilience
  (`_salvage/internal/serve/serve.go:561-584`). Whether the greenfield
  serve binds a bridge address at all is a rung 5 decision that the unit
  then follows, and it depends on whether the streamer proxy exists at
  rung 5, which it does not.

`ReadWritePaths` is the one line a greenfield serve unit cannot copy.
Section 4 derives it.

---

## 3. Config

### 3.1 The two reserved classes rung 5 owns

`internal/config/keys.go:60-85`, verbatim in content:

| Key | Class | Kind | Secret? |
|---|---|---|---|
| `DUBPLATE_MAGIC_LINK_SECRET` | auth and identity | string | Yes, minted |
| `DUBPLATE_SESSION_SECRET` | auth and identity | string | Yes, minted |
| `DUBPLATE_PUBLIC_URL` | auth and identity | string | No |
| `DUBPLATE_SMTP_HOST` | auth and identity | string | No |
| `DUBPLATE_SMTP_PORT` | auth and identity | int | No |
| `DUBPLATE_SMTP_USER` | auth and identity | string | No |
| `DUBPLATE_SMTP_PASSWORD` | auth and identity | string | Yes, minted |
| `DUBPLATE_SMTP_FROM` | auth and identity | string | No |
| `DUBPLATE_SERVE_BIND_ADDRS` | serve | string list | No |
| `DUBPLATE_SERVE_PORT` | serve | int | No |

The `serve` class carries its own rationale in a code comment
(`keys.go:74-79`): the shell binds its listener in the same rung as
auth, and the wall keeps them apart because "a bind address is not an
identity concern".

### 3.2 What the ten keys do not cover

Six requirement groups have no reserved name.

**The owned prefix.** B3 (BS:428-432) says dubplate's entire HTTP
surface lives under exactly one configured prefix and that "the auth
boundary is decided by configuration, never by mux-miss". No key names
it. The `serve` class has two keys and neither is a prefix.

**The identity-source selector.** B10 (BS:499-503) requires exactly one
active source selected by config. The wall carries the pending row and
no key.

**The three flag-enabled classes the build spec names and
`internal/config` lacks.** Rung 1b's Resolution R13
(`docs/plans/2026-09-07-rung-1b-config-registry.md:818-830`)
recorded them without declaring them: an SSO proxy capability, a
trusted-proxy mode, and a forward-auth front, all named at BS:167-175
as flag-enabled integrations. The forward-auth front appears in the tree
only inside two supersession messages
(`keys.go:169-174`, both Access rows redirected to "the forward-auth
front's own config, arriving at rung 5"), and the other two appear
nowhere. R13 filed all three "to rungs 5 and 12's own spec work"
(`HISTORY:549-550`). B2 makes the forward-auth front rung 5's, since
rung 5 builds the verify endpoint. The trusted-proxy mode is rung 5's by
B8. The SSO proxy capability is rung 12's.

**The thresholds.** BS:161-162 requires every tunable be a threshold
with a sane default, naming session lifetime, link expiry, rate limits,
and enumeration behavior. Thresholds are their own registration family
(`registry.go:33-46`) declared in the rung's own file, exactly as
`registerShadowThresholds` does at `cmd/dubplate-server/main.go:66`.
None is reserved and none needs to be. The `_PREVIOUS` rotation-overlap
names B6 requires (BS:466-468) are in the same position: rung 5 declares
them itself.

**The theme keys.** DL R3.3 (`DL:261-273`) says two string-valued
environment keys express the match and palette picks. DL does not say
which rung declares them, and rung 5 is the rung that stamps the
composed `data-theme` (R3.6, `DL:315-317`). R4.3's branding variables
and R4.4's operator CSS file path are two more config surfaces rung 5's
serve layer reads. `KindString` exists for exactly this shape, landed at
rung 1b Task 5 (`HISTORY:436-440`).

**The roster's source.** Section 5.

### 3.3 The registry contract

Rung 1b's close states the rule as a contract change
(`HISTORY:487-499`): the wall is registry contents, `Load` takes the
registry it enforces against, a rung declares its keys in its own file
inside the package that owns them, and no lane adds a name to
`internal/config/keys.go`. Removal is the half rung 5 exercises first. A
lane "removes only its own reserved class from that file, in the same
commit that declares the class in its own package".

Mechanically, that means one commit that deletes the `auth and identity`
literal from `keys.go:60-73` and adds a `registerAuth(reg)` function in
`cmd/dubplate-server` or in `internal/auth`, plus a second for the
`serve` class. `NewRegistry` seeds rung 1's own declarations
(`registry.go:63-77`), and every `Add*` entry point runs `requirePrefix`
before any claim (`registry.go:106-110`), so a rung 5 key without the
`DUBPLATE_` prefix panics at registration rather than registering
unreachable. The collision gate panics rather than returning an error,
on the stdlib precedent recorded at `registry.go:24-30`, and a
gate-resident test constructs the production registry exactly as the
binary does.

One detail worth naming in the plan: `AddThreshold` skips the claim for
a key that is also an integration class's `Optional` member (the R5
pairing, `registry.go:244-249` and `:279-291`). Rung 5's flag-enabled
classes will have optional members that are also thresholds, so this
pairing fires for the first time outside rung 4's shadow class.

### 3.4 The env template and the minting rule

`env/dubplate-server.env.template` carries the three row states in its
own header (`:8-16`): a bare `NAME=` resolves from the workstation age
store at render time or `deploy.sh` aborts, a `NAME=value` line is a
literal for a value that will never be minted, and a commented-out
`NAME=` line is a name whose value arrives later, on the
`config.WallRow.Pending` precedent. The header notes that no commented
row exists yet.

`scripts/deploy.sh:176-200` is `render_template`, and its comment
records the hazard that produced the literal rows: a bare row resolves
through indirect expansion against the shell's own environment, which
already carries the incumbent's exported values, so a bare
`DUBPLATE_STATE_DIR=` would silently inherit `/var/lib/musicbox`.
`render_missing` accumulates across both templates so one run reports
every unresolvable row together.

RE:(5,4) settled the collision this creates. Rung 4's deploy is merge
point M1, `deploy.sh` hard-fails the whole render on one unresolved
name, and the build spec puts minting at rung 16. The paste-ready row
from that cell is the rule rung 5's spec restates: rung 5 adds no name
it has not minted, a name whose value arrives later lands as a commented
pending row, never as a live `NAME=` line, and rung 16 mints the auth
secrets through the age-store flow.

The template's header also records that no row in it is a secret today
and therefore no name in it is registered at
`~/.dotfiles/secrets/registry.md` (`:18-25`). Rung 5's three secrets are
the first that will be, and CLAUDE.md's rule applies to them in full:
`secret-set.sh` mints, the registry records scope and rotation, and only
names ever appear in the repo. The `_PREVIOUS` overlap doubles the count
at rotation time, which the registry entry should state when it is
written.

---

## 4. The state directory

### 4.1 What rung 5 writes

Two records, both named in the birth requirements. Consumed nonces are
durable in the state directory with TTL pruning, and the verification
order is fixed: verify the MAC, then expiry, then insert the nonce
(BS:437-439). A per-email valid-after epoch gives per-person revocation
of the dubplate session (BS:455-456). RE:(5,6) confirms rung 5 needs no
store, no migrations, and no ops vocabulary, and that both records are a
deliberate decoupling from rung 6's SQLite.

### 4.2 Rung 4's layout precedent

`internal/engine/shadow/store.go` is the model, and its shape is worth
copying wholesale.

- One package-level constant names the single leaf directory, with the
  reason stated: `storeDirName = "shadow"`, "the one directory under the
  state directory the shadow writes, so 4b's `ReadWritePaths` is a
  single static path" (`store.go:37-39`).
- `StoreDir(stateDir string) string` joins it once, so "the directory
  name has one home rather than a literal per call site"
  (`store.go:41-47`).
- `OpenStore(root string)` creates the subdirectories it needs and
  returns the store, with directories 0755 and files 0644 per
  go-conventions (`store.go:49-59`).
- `validID` refuses `""`, `.`, `..`, and anything containing a slash or
  backslash, because ids are digest-derived, "so anything else is a bug
  rather than input, and this check makes it a loud one rather than a
  path traversal" (`store.go:61-70`).
- Every write goes through `atomicfile.WriteFile`, so a reader never
  sees a partial file and two writers never contend (`store.go:17-21`).

`internal/atomicfile` offers two entry points. `WriteFile(path, data,
perm)` (`atomicfile.go:12-16`) writes a temp file in the target's own
directory, syncs, and renames. `WriteStream(path, perm, write)`
(`atomicfile.go:45-55`) does the same without buffering the content in
full and additionally fsyncs the containing directory after the rename.
Nonce and epoch records are small, so `WriteFile` is the fit, with one
caveat: `WriteStream` is the one that fsyncs the directory entry, and a
nonce record that survives the crash it exists to defend against wants
that guarantee. The spec should state which it uses and why.

### 4.3 The permission question rung 4 did not have to answer

The shadow store writes 0644 files. Rung 5's records are identity state:
a consumed-nonce set and a per-person revocation epoch. Neither is a
secret in the sense the age store means, and neither is public. The unit
runs as `music` with `UMask=0027`
(`systemd/dubplate-shadow.service:79`). Recommendation, provisional: 0600
for the auth leaf's files and 0700 for its directory, stated in the spec
with the reason, because the default inherited from the shadow
precedent would be the wrong one silently.

### 4.4 How the leaf grant generalizes

`systemd/dubplate-shadow.service:23-40` is the argument, and it is worth
reading in full before writing rung 5's unit. Three claims hold it up.
`TestStoreWritesNothingOutsideItsRoot` is "the static argument that every
write the store makes, and the attacker-shaped `..`-bearing id it
refuses, stay confined to the `OpenStore` root passed in". `deploy.sh`
pre-creates the exact path before the unit starts, without which startup
fails with "Failed to set up mount namespacing". And the literal is
derived rather than chosen: it is `DUBPLATE_STATE_DIR` plus `/shadow`,
the same join `shadow.StoreDir` makes at runtime.

Three copies move together, and `scripts/check.sh:154-167` gates their
agreement. The check reads `DUBPLATE_STATE_DIR` out of the template,
`ReadWritePaths=` out of `systemd/dubplate-shadow.service`, and
`DUBPLATE_LIB_DIR` out of `deploy.sh`, then fails when they disagree.
`scripts/deploy.sh:489-505` is the pre-create, which creates
`/var/lib/dubplate` as a side effect of creating `/var/lib/dubplate/shadow`
and chowns both to `music`.

A rung 5 serve unit adds a fourth literal the gate cannot see, because
the check names `dubplate-shadow.service` explicitly. Rung 5 either
generalizes the check over every greenfield unit or adds a second
hard-coded pair. Recommendation, provisional: generalize, since rungs 6,
7, and 9 each add a state-directory consumer and the third copy of a
one-unit check is the point at which it should have been data.

One more grant is likely. If rung 5 serves an operator CSS file named in
config (DL R4.4), that path is a read, not a write, and
`ProtectSystem=strict` permits reads outside `ReadWritePaths`. No extra
grant is needed for it.

---

## 5. The roster

### 5.1 Where members come from today

`config/contributors.yaml` is the file, with `name`, `email`, and `role`
per contributor. Its header carries the finding that made the name field
load-bearing: `name` IS the contributor's inbox directory name, matched
as an equal string, "so display names do not belong here (found live
2026-08-31, in the retired shell layer: 'Geoff Wright' never matched
inbox/geoff, so no uploader notice ever sent)". The last line of that
header says the uploader initiative replaces the file with email-keyed
identity.

`scripts/deploy.sh:420-426` installs it to
`/etc/dubplate/contributors.yaml` at 0640 root:music, and the incumbent
reads it through `DUBPLATE_CONTRIBUTORS_FILE`.

The salvage loader is `_salvage/internal/config/contributors.go`. It
builds a `Roster` with `byName` and `byEmail` maps (`:23-56`) and offers
`EmailFor` (`:61`), `FirstName` (`:74`), `Has` (`:83`), and
`DirNameFor` (`:93`), with email matching case-insensitive. Its
missing-file behavior is the part that does not survive contact with the
birth requirements: an absent file "returns an empty `Roster` and a nil
error" (`:33-39`). That is a fail-open, and B9 requires the roster
validate at load.

### 5.2 What the birth requirements ask for

BS:491-498, in full: unique names; addresses parsed with
`net/mail.ParseAddress`; no control characters; bounded length; the
streamer principal split from the inbox directory name via an optional
per-contributor `streamer_username` with startup and `doctor` collision
checks; and the principal, whichever field goes on the wire, validated
as a stable identifier that is ASCII-printable, unique, and bounded,
with display names left free.

UNI:13 is the review finding that produced the split: the design "fuses
three identities into the roster's `name` and leaves an adopter no field
to separate them". The three are the inbox directory name, the streamer
principal, and the human display name.

### 5.3 The contradiction

`internal/config/keys.go:176-177` retires `DUBPLATE_CONTRIBUTORS_FILE`
with the message "retired: the roster loads from the store (rung 6), not
a config-supplied file". Setting that key today is a load error carrying
that text.

Three documents then disagree about who owns the roster loader and where
it reads from. PA:180 awards the roster loader to rung 5, noting it was
"assigned twice today, to 5 and 7". BS:645-646, ladder entry 7, still
says "roster loading with the birth-requirement validation" is rung 7's,
which RE:(5,7) already flagged as unamended build-spec text. And
`keys.go` says the source is rung 6's store, which rung 5 does not have.

Rung 5 cannot defer the question. It needs a roster to answer whether a
submitted address may receive a link (MD:86-95), and it needs the
principal to inject later. The options, with the recommendation:

1. **A config-supplied YAML file, a new key, superseding the
   supersession.** Cheapest to build, matches the adopter's mental model
   of a small text file, and needs a new `DUBPLATE_ROSTER_FILE` key plus
   an amendment to the `keys.go` retirement message.
2. **A file at a fixed path under the state directory.** No new key, but
   an adopter editing identity by hand inside a directory the server
   otherwise owns is a worse story, and it collides with rung 6's later
   ownership of that directory.
3. **Wait for rung 6.** Refused by the phase plan: rung 5 runs in phase
   1 and rung 6 opens phase 2 (PA:97-113).

Recommendation, provisional, and a **product fork**: option 1, with the
key declared in rung 5's own file, the supersession message rewritten to
name the new key rather than the store, and the spec stating explicitly
that rung 6's store may later become the roster's home without changing
the loader's interface.

### 5.4 What hook 5 needs rung 5 to keep open

S20:367-369: "A stable identity, rung 5 and rung 6. Rung 6 mints a
durable surrogate user id that per-user rows key on, and rung 5 resolves
a roster identity to it. The roster's `streamer_username` becomes
required, unique, and stable rather than optional, and it is the
renameable handle the surrogate id protects against."

Two things follow for rung 5's design, and one tension.

Rung 5 keeps a single resolution point where a login address becomes an
identity value, so rung 6 can substitute a surrogate id there without
touching call sites. DL R4.2 already leans on the same hook from the
presentation side: the stored theme pick keys on rung 6's surrogate id
and lives in `localStorage` until then (`DL:351-370`).

Rung 5 should also treat `streamer_username` as required-shaped even
while it is optional, since 2.0 makes it required and a roster written
without it is a migration rung 6 pays for.

The tension: B5 keys the revocation epoch on the email
(BS:455-456), while hook 5 says per-user rows key on the surrogate id
and never on an address. Both are right for their moment, since the
epoch must exist before any user id does. Recommendation, provisional:
rung 5 keys the epoch on the normalized login address, states in the
spec that it is the one per-user record deliberately keyed that way, and
records the re-key as a rung 6 obligation so rung 6 inherits it instead
of discovering it.

---

## 6. Mail

### 6.1 The seam is rung 5's, and three documents said otherwise

PA:179 awards `internal/notify` to rung 5, "it must send mail to deliver
magic links, and its deliverables already carry SMTP send", with rung 9
demoted to a consumer. RE:(5,9) found the live contradiction and settled
it: the build spec's ladder entry 9 and the dependency analysis's Wave B
territory column are errata, and rung 9 builds the monitoring seam and
integrates the mail seam it does not own. RE:(5,7) adds the shape
requirement: rung 5 births notify "as a general send seam rather than a
magic-link mailer", because a notify seam designed as a magic-link
mailer is the predictable defect when rung 7's reject emails and rung
9's notifications arrive.

### 6.2 The salvage mailer

`_salvage/internal/notify/notify.go` is 96 lines. Its package comment
(`:1-6`) records what it replaced and the contract it honors: a native
TLS client rather than shelling out to `sendmail(1)`, with implicit TLS
on port 465 and STARTTLS on any other port, matching the msmtprc the box
had. `Mailer` wraps a `config.SMTPConfig` (`:26-42`), `New` returns one
(`:44-48`), and `Send(ctx, to, subject, body)` (`:50-92`) synthesizes
From, Date, and Message-ID, RFC 2047 encodes the subject, sends the body
quoted-printable, logs one slog record on success, and leaves the
caller to decide whether a failure aborts anything.

It depends on `github.com/wneessen/go-mail`. Two of its fields exist
only for its own tests (`tlsConfig` and `forceImplicitTLS`), each with a
comment saying so, which is the shape go-conventions asks for rather
than an injected interface.

The adoption question the plan must answer under CLAUDE.md's four
criteria is whether go-mail earns its place against `net/smtp`.
`net/smtp` is frozen and carries no message construction at all, so
choosing it means hand-rolling MIME, RFC 2047 encoding, and
quoted-printable, which is exactly the surface where a hand-roll goes
wrong quietly. Recommendation, provisional: adopt go-mail, with license,
release cadence, open-issue age, and importer evidence recorded in the
plan per CLAUDE.md's adoption bar, since the salvage code is proven
against the estate's own provider and the alternative is bespoke MIME.

SEC:94 is the one security finding that lands directly on this package:
SMTP construction from roster values needs explicit CRLF rejection. A
header-injection guard on the recipient address belongs in the seam
rather than in each caller.

### 6.3 The transport options, priced for a stranger

**SMTP.** The universality test is stated twice, at MD:41-43 and again
at BS:163-164: "The design makes sense with no Cloudflare account: plain
SMTP plus any streamer, tunnel optional." The adopter supplies five
values, already reserved as keys: host, port, user, password, and from.
Any provider works. The operational cost is one mail account and one
app password, and the failure mode most adopters hit is a provider that
rejects the From address, which is a configuration error the seam can
report clearly.

**Fastmail JMAP.** `~/.claude/instructions/fastmail-api.md` documents
the workstation's access: a session endpoint at
`https://api.fastmail.com/jmap/session` and an API endpoint at
`https://api.fastmail.com/jmap/api/`, both Bearer-authenticated with
`FASTMAIL_API_TOKEN`. The workstation's poplar client proves a Go JMAP
sender is real rather than theoretical: `jmap/emailsubmission.go:113`
implements `EmailSubmission/set`, `jmap/jmap.go:17` declares the
submission capability URI, and `jmap/conformance_jt_test.go:706` asserts
the two-method submission shape RFC 8621 section 7.5 defines. As a
dubplate transport it fails universality outright. It binds the adopter
to one vendor, and it buys nothing over SMTP, since Fastmail also speaks
SMTP.

**A Cloudflare Worker.** The estate already runs one for failure pings.
`~/Projects/pings/wrangler.jsonc` declares a `send_email` binding named
`EMAIL`, and `~/Projects/pings/src/email.ts:28-42` is `sendAlert`,
which renders text and HTML bodies and calls `email.send({from, to,
subject, text, html})` with a fixed `FROM_ADDRESS` of `pings@907.life`.
The binding is Email Routing's, and its constraint is decisive here:
destinations must be verified Email Routing addresses. The pings repo
recorded the verification step as part of provisioning
(`~/Projects/pings/docs/HISTORY.md:47-48`, the destination address "came
back already verified") and even noted the dashboard reports a
successful send as "dropped" (`~/Projects/pings/README.md:70-73`). A
magic-link mailer sends to whatever address a family member typed, so a
verified-destinations transport cannot serve it. Beyond that, requiring
a Cloudflare account breaks the design's own fourth universality test.

Recommendation, provisional, and a **product fork**: SMTP is the only
transport rung 5 builds. The seam is one small interface with a single
send method taking `ctx` first, so a later transport is an
implementation rather than a refactor, and no transport beyond SMTP is
built until an adopter needs one.

### 6.4 What the seam owes its later consumers

Rung 7's reject emails share wording with the uploader page verbatim
(CLAUDE.md's copy-contract fact), so the seam carries a body and a
subject as data and owns no copy of its own. Rung 9 integrates
monitoring, which is pings rather than mail. The seam's own error
doctrine is already written in the salvage comment: notify is
best-effort margin rather than a per-run failure. Rung 5 breaks that in
one place, and the spec should say so plainly. A magic link that fails
to send is a login that fails, so the login handler treats a send error
as a user-visible failure while every other caller keeps the
best-effort contract.

---

## 7. The web surface

### 7.1 What the 0c standard binds, clause by clause

Every clause below binds rung 5 directly. R1.1 (`DL:61-65`) makes it
explicit: the standard binds rung 5's login and consumption pages first
and every later web surface after them.

| Clause | Lines | What rung 5 must do |
|---|---|---|
| R1.6 | 88-100 | Build the pages as SvelteKit static, client-rendered, served by the Go binary. Record the build-spec correction at the rung's close. |
| R2.1 to R2.7 | 103-234 | Consume the 32-name variable contract. Read only semantic tokens, never a DaisyUI internal class name. |
| R3.3 | 261-273 | Two string-valued environment keys carry the match and palette picks. |
| R3.6 | 310-317 | Stamp one composed `data-theme` value, spelled `<match>-<palette>`. "Rung 5 builds the stamping against this clause." |
| R3.7 | 318-327 | The `dubplate` match block's `--size-field` yields a 44px touch target. |
| R4.1 | 335-350 | Bundle stylesheet, then branding stylesheet, then operator file, in document order. |
| R4.2 | 351-370 | The Go binary renders the attribute into the served `index.html`. The pick lives in `localStorage` until rung 6. The document is served `no-cache` while hashed assets stay immutable. |
| R4.3 | 371-392 | Generate the branding stylesheet from config at serve time, at a fixed path, empty when nothing is configured. |
| R4.4 | 393-399 | Serve an optional operator CSS file named in config, last. |
| R4.5 | 400-411 | Produce the served document through one documented rewrite over the embedded `index.html`, substituting a named placeholder token. |
| R4.6 | 412-420 | Ship no `theme-controller` input. |
| R5.1 to R5.5 | 424-469 | Embed woff2 through `go:embed`, no external origin, license texts and a manifest in the tree, at most two families and four files. |
| R5.6 | 470-500 | Serve the CSP directive set. |

R5.6 is the clause carrying the one owner note DL routes to the rung 5
sitting (`DL:29-31`): the served CSP's `style-src 'unsafe-inline'`
relaxation. The reason is recorded at `DL:492-500`. DaisyUI's
`radial-progress` reads `--value` from a style attribute and most Svelte
transitions create an inline `<style>` element at runtime, so no
build-time hash can cover them. `script-src` stays hash-pinned with no
`'unsafe-inline'`, which is the directive whose failure would let the
page load and never boot.

Two consequences of the stack the spec should restate rather than
rediscover. SvelteKit emits inline scripts, so `kit.csp` must be
configured and `script-src` must permit them by hash. `adapter-static`
prerenders every page and nonces are forbidden on prerendered pages, so
`kit.csp.mode` is `'hash'`. Whether the hashes reach the header from a
build-time manifest or from hashing the embedded document at startup is
rung 5's choice, and G6 asserts the outcome either way (`DL:487-491`).

### 7.2 CLAUDE.md's web rule

`CLAUDE.md:135-143`. SvelteKit static mode, Svelte 5 runes, DaisyUI 5,
`adapter-static` with client-side rendering only, built to static assets
embedded via `go:embed` and served by the Go binary as the only server.
Server-only SvelteKit files are forbidden and the gate checks that none
exists. `tus-js-client` is pinned at 4.3.x, which is rung 7's concern.
Node is workstation build-time only, and nothing Node ships to the box.
The last sentence names rung 5's login page as the first page built
under the rule.

### 7.3 The gate's web leg

There is no web leg today. `scripts/check.sh` runs `make check`,
shellcheck, ruff, gitleaks, vale, `docker compose config`, and a
pattern-filtered `systemd-analyze verify`, and its header says the web
leg rejoins at rung 7 (`:5-6`). `Makefile:33` says the same above the
`check` target, whose recipe is `vet lint test tidy-check`
(`Makefile:34`). `CLAUDE.md:118` is the third copy.

DL section 6 rules the rejoin at rung 5 and specifies ten gates.
G1, G2, G8, G9, and G10 read source. G3, G4, G5, G6, and G7 read a built
bundle or a served document, so they need the leg to build one, and each
"fails with a named message when the embedded tree is the
placeholder-only tree a never-built checkout carries" rather than
passing on an empty iteration (`DL:522-527`).

That requirement collides with the repo's own gate rule, which rung 5
must resolve rather than skirt: a check either gates or does not exist,
and none may be guarded by a `command -v` skip. The web leg needs Node
on the machine running it. `scripts/check.sh:19-38` already establishes
the pattern for a hard tool requirement: each of go, golangci-lint,
shellcheck, ruff, and gitleaks fails the run by name when absent, while
`systemd-analyze` and Docker are the two that skip. Recommendation,
provisional: node and npm join the hard-required list, with the error
message naming mise as the install path, since the alternative is a gate
that silently stops checking every web clause the moment a machine lacks
Node.

The SvelteKit server-file check (`CLAUDE.md:137-139`) is a separate rule
DL does not restate (`DL:509-510`), and it is rung 5's to build, since
rung 5 creates the first `web/` tree that could violate it.

### 7.4 The salvage web build

`_salvage/web/package.json` pins Svelte 5.57.0, Vite 8.2.2, Tailwind
4.3.3, DaisyUI 5.7.25, TypeScript 6.0.3, vitest 4.1.11, and
`tus-js-client` 4.3.1, with `@sveltejs/vite-plugin-svelte` rather than
SvelteKit. DL's salvage adoption record (`DL:648-668`) already ruled on
every candidate this tree offers, and the two that matter to rung 5's
mechanics are the embed pattern and the cache rules.

`_salvage/web/embed.go` is 24 lines and carries the lesson in its
package comment: `npm run build` writes `index.html` and a hashed
`assets/` directory into `dist/`, and `dist/placeholder.txt` is a
separately committed copy of `public/placeholder.txt` "so this package
still builds on a fresh checkout that has never run the web build (a
bare, empty `dist/` fails `//go:embed` at compile time)". `Assets` is
declared `//go:embed all:dist` and stays rooted at `dist`, because the
consumer strips the prefix with `io/fs.Sub` itself, "since that is the
one place able to return the error a missing directory would produce
instead of panicking at program init". Both halves are worth carrying
forward verbatim in reasoning.

`_salvage/internal/serve/serve.go:279-297` is `staticHandler`, whose two
cache rules DL adopts at `DL:665`: a hashed `/assets/*` entry is
immutable forever, and every other path is `no-cache`. The comment
records why `private` rather than `public` was correct there, which is a
detail rung 5 must re-derive rather than copy, since its login page is
served to an unauthenticated visitor.

### 7.5 What the serve shell owes the web surface

R4.5's rewrite is the seam. The Go layer reads the embedded
`index.html`, substitutes a named placeholder token with the composed
`data-theme` value, and adds the tier 2 and tier 3 `<link>` elements the
SvelteKit build did not write. That makes the static handler two
handlers rather than one: a document renderer for `index.html` and an
immutable asset server for everything hashed. The extension point of
section 2.1 must accommodate rung 7 swapping the embedded bundle for a
larger one without touching either.

Rung 5 also owns three serve-side surfaces RE:(5,7) awarded it: the
`go:embed` static asset path in `internal/serve`, the embedded woff2
fonts themselves, and the config-rendered CSS variable injection point.
The fonts are new work in full; DL:667 records that the salvage tree
ships no font and no `@font-face`. R5.2's license bar and manifest
(`DL:431-444`) apply, with the faces at `web/static/fonts/`, license
texts at `web/static/fonts/licenses/`, and the manifest at
`web/static/fonts/FONTS.md`.

One thing rung 5 must not do: author a theme roster entry. RE:(5,7)'s
second paste-ready row says the standard ships one default block and
rung 7's roster supersedes it without touching rung 5's templates,
because the pages bind to variable names only.

---

## 8. Streamer SSO

### 8.1 What MD's mapping said, and what survives

MD:124-139 maps the trusted header value to the contributor's roster
`name`, the same identity that names their inbox directory; puts the
header name and injection with the streamer driver, with Navidrome's
`Remote-User` via ExtAuth as the reference; lands the post-login
redirect on the player root so Navidrome's auto-creation fires; and
sends phone apps to the streamer's own settings for an app password.

B9 overturns the first clause. UNI:13 is the finding: the design fuses
three identities into one field. The wire identity is now the validated
principal, which is `streamer_username` when present and the name
otherwise, ASCII-printable, unique, and bounded.

The rest survives as rung 12's material rather than rung 5's. The 2.0
erratum trims rung 12 to "the driver seam, SSO, roster sync, and the
proxy-hygiene ingress" (S20:441-444), and DL section 7 adds the
boundary rule that the driver seam declares no theming capability at all.

### 8.2 What rung 5 must leave in place

Six things, each traceable to a birth requirement.

The ingress strip runs for every path before routing, with the set
driven by the driver's configured header name plus rung 5's fixed extras
list (BS:405-410). Rung 5 ships the extras and the mechanism by which a
driver contributes its name.

The fallthrough slot exists and is empty. Rung 12's proxy occupies it
(PA:77), and rung 14 terminates the owned prefix so the fallthrough
cannot capture an unknown API path (RE:(14,12)).

The session cookie carries the `__Host-` prefix, which makes it
host-scoped by definition and makes same-origin a stated precondition of
the SSO capability (BS:441-447). No cookie-domain widening is offered.

The verify-identity seam exists as an interface from this rung, with
magic-link as its reference implementation and forward-auth fronts as
alternates, and the verified identity carries its assertion channel,
which is cookie, header, or service token (BS:457-463). That provenance
is what makes the organizer's rule enforceable that admin verbs reject
cookie-derived identity.

The post-login destination is a driver capability with a core default
(BS:504-510). Without SSO the dubplate root serves. The redirect-target
validation rule travels with the mechanism: single-decode,
relative-only, no scheme, no `..`, no backslash variants.

The revocation seam exists for a later credential store to subscribe to.
S20:245-246: "Rung 5's identity layer exposes the revocation seam a
credential store subscribes to", so that a roster-level revoke
invalidates every app password 2.0 mints.

### 8.3 What rung 5 must not build

No `httputil.ReverseProxy`, no Navidrome header name, no ExtAuth
configuration, no trusted-source pin, no roster sync, no app passwords,
and no driver package. The one-owned-prefix rule (BS:428-432) is what
makes the omission safe: dubplate's whole surface lives under one
configured prefix, the auth boundary is configuration rather than
mux-miss, and a rung 5 deployment with no streamer simply serves nothing
at the fallthrough.

### 8.4 The hook statement rung 5's plan owes

CLAUDE.md requires every rung plan to state per task whether it lands a
2.0 forward hook or is neutral. For rung 5 the hook list resolves to
one landed hook and one adjacent.

Hook 5, a stable identity (S20:367-369), is landed by rung 5's roster
resolution point, its treatment of `streamer_username`, and the epoch
key decision of section 5.4.

Hook 6, the driver seam stays universal (S20:370-373), is rung 12's, and
rung 5 lands nothing against it beyond leaving the fallthrough and the
strip set as data. DL already discharged its presentation half by
declaring theming out of the seam (`DL:643-647`).

Hooks 1 through 4 and 7 through 14 are schema, tagging, layout, and
migration concerns that rung 5 does not reach.

---

## 9. Review shape

Rung 5 is one of three rungs that keep the full fresh-context
adversarial cycle. CLAUDE.md's build-process section names them: "auth
(5), streamer/SSO (12), and cutover (16) keep the full fresh-context
adversarial cycle", while mechanical rungs ride the in-chain
`diff-reviewer`. The ladder entry adds that "the `web-auth-security-reviewer`
gate runs inside the rung" (BS:626-627), which means inside rather than
at the pass-end fan-out.

BS:398-404 sets the second obligation. Both login-spec reviews are birth
requirements, and "the rung plans walk the full review texts as
checklists, minors included". The two texts are not small.
`SEC` carries 6 blockers (`:11-26`), 13 majors (`:33-69`), and 12 minors
(`:76-109`), and closes with a verdict of not ready to build.
`UNI` carries 3 criticals (`:7-13`), 6 highs (`:16-31`), 7 mediums
(`:34-52`), and 3 lows (`:55-61`). That is 50 items, and the plan owes a
disposition for each. Most are already discharged by the birth
requirements themselves, so the honest form is a table mapping each
finding to the birth-requirement clause that absorbed it, with the
residue called out.

The reviewers CLAUDE.md names for this surface are
`web-auth-security-reviewer` for serve and JWT, `svelte-reviewer` and
`daisyui-a11y-reviewer` for `web/`, and `diff-reviewer` per plan task.

The repo's own precedent for a fresh-context adversarial cycle is the
U13 round, recorded at `HISTORY:3788-3790`: 57 agents and 45 confirmed
serious findings from five reviewers with adversarial verification,
folded as F1 through F8. That round found a defect no gate could have,
namely that the tag probe never existed and no main-lane upload could
reach the library.

Recommendation, provisional, for rung 5's shape. Each pass runs the
standard implementer, `diff-reviewer`, gate chain. The auth-core pass
adds `web-auth-security-reviewer` inside the pass rather than after it,
because the ladder entry says so. The login-pages pass adds
`svelte-reviewer` and `daisyui-a11y-reviewer`. The rung's final pass
runs the full cycle over the whole subsystem with at least two
independent fresh-context lenses, one security and one universality,
mirroring the pair that produced SEC and UNI, since the universality
lens caught three criticals the security lens did not.

---

## 10. Budget evidence and a pass split

### 10.1 What the ledger says about L-band work

Rung 3 is the only L-band rung this build has finished. It ran as four
passes, and rung 8 is running as at least two.

| Pass | Spend | Ceiling | Source |
|---|---|---|---|
| 3a | 2.84M | 6M | `HISTORY:3243-3247` |
| 3b | 4.82M | 6M | `HISTORY:2928-2935` |
| 3c | about 4.9M | 6M | `HISTORY:2551-2557` |
| 3d | about 5.15M | 6M, re-derived to an expected 4.8M | `HISTORY:2141-2147` |
| 4a-i | about 2.69M | 3M sub-ceiling | `HISTORY:1660-1662` |
| 4a-ii | about 2.7M | 3M | `HISTORY:1380-1388` |
| 4b (unattended half) | about 2.6M | 3M | `HISTORY:911-915` |
| 1b | about 3.7M | 2.5M | `HISTORY:739-746` |
| 8a | about 3.6M | 4.5M | `HISTORY:403-411` |

Two patterns are worth carrying into rung 5's ceilings.

Plan preparation is the largest single line in every rung 3 pass. 3c
spent 2.55M on prep against 1.74M for the whole task chain
(`HISTORY:2551-2554`), and 3d spent 2.51M on prep against a 5.15M total
(`HISTORY:2141-2143`). 3b's entry states the lesson plainly: without the
plan-prep workflow, execution closed at 2.7M (`HISTORY:2932-2938`).

Ceilings are overrun when the pass is a contract change rather than a
build. 1b ran 3.7M against 2.5M, and its own entry records that the 2.0M
flag was crossed inside the chain with no conductor action, "a conductor
miss this entry records rather than omits" (`HISTORY:743-746`). Rung 5
changes contracts in at least three places: the serve shell's extension
point, the verify-identity seam, and the roster's source.

### 10.2 The proposed split

Four passes, sequenced so each one's deliverable is testable without the
next. Ceilings are per pass and include plan preparation, since that is
where the evidence says the spend goes.

**5a, the serve shell and ingress. Ceiling 3M.** `internal/serve` born:
the composition order of PA:79 with the strip, the normalization, and
the prefix terminator; the declared extension point; the empty
fallthrough; the `http.Server` construction with the per-route
write-deadline policy; the `serve` reserved class removed from
`keys.go` and declared in rung 5's own file, with the prefix key and the
bind keys; the `serve` verb wired into `newRootCmd`. No auth, no pages.
Acceptance is the shell's own tests plus the strip and normalization
properties.

**5b, the auth core. Ceiling 4M.** `internal/auth` born: token
serialization with domain separation and the length floor, sessions with
the `__Host-` cookie, the absolute cap and rolling refresh, durable
nonces and the per-email epoch in the state directory, the roster loader
and its validation, rate limiting with the client-IP mode, the CSRF
rule, and the verify-identity seam with provenance. `internal/notify`
born as a general send seam with the SMTP transport. The `auth and
identity` reserved class removed and declared. The
`web-auth-security-reviewer` gate runs inside this pass. This is the
largest pass and the one whose scope most wants watching.

**5c, the login pages and the web leg. Ceiling 4M.** `web/` born as a
SvelteKit static app under CLAUDE.md's rule: the login page, the
consumption interstitial, the `go:embed` path, the R4.5 rewrite, the
tier 2 branding stylesheet, the fonts with their licenses and manifest,
and the CSP. The gate's web leg rejoins, with DL's G1 through G10 and
the server-file check, and the three "rejoins at rung 7" copies are
corrected. `svelte-reviewer` and `daisyui-a11y-reviewer` run inside.

**5d, the boundary, the deployment, and the close. Ceiling 3M.** The
forward-auth verify endpoint with its own trust-boundary documentation;
the forward-auth front's config class; the active `doctor` checks of
section 1.4; the serve unit, its `ReadWritePaths` grant, the deploy
pre-create, and the generalized state-dir gate; the pending env rows;
the full fresh-context adversarial cycle over the whole subsystem; the
build-spec errata and the rung close.

Total 14M across four passes, against rung 3's four-pass actual of about
17.7M. The estimate is lower because rung 5 has no corpus, no oracle,
and no external API budget, and higher per pass than rung 4's because
three of its four passes change a contract.

Two scheduling notes for the plan header. RE:(5,3)'s paste-ready
constraint says the second code slot passes at a pass close rather than
a rung close, so lane A releases the worktree at each of these four
merges. And pass 8a's chain ran under a session window cap that STATUS
recorded mid-pass as binding before the pass's own 4.5M ceiling would
have, which is the reason to plan every task boundary in these four
passes as a resumable stopping point.

---

## Surprises

1. **The roster has no source.** `keys.go:176-177` retires the
   contributors file with "the roster loads from the store (rung 6)",
   RE:(5,6) says rung 5 needs no store by deliberate design, and B9
   requires a validated roster at rung 5. Three documents also disagree
   about who owns the loader. This is the single largest gap in the rung.

2. **`doctor` is a reporter and the birth requirements need a prover.**
   B2 asks `doctor` to reject a spoofed identity header from a
   non-trusted source, and B8 asks it to report which client-IP mode is
   live. `cmd/dubplate-server/doctor.go` walks registry families and
   prints `%-32s %s` lines. It has no vocabulary for a check whose result
   depends on evidence.

3. **Clearing the pending wall row is not a one-line change.** Removing
   `Pending` from `{Name: "identity source"}` (`keys.go:24`) would make
   `Load` require an environment variable literally spelled
   `identity source`. The row has to become a real key or retire.

4. **Three copies of "the web leg rejoins at rung 7" exist and the 0c
   standard names two.** `CLAUDE.md:118`, `scripts/check.sh:5`, and
   `Makefile:33`. DL:513-517 lists the first two.

5. **B12 is filed under the streamer rung and half of it is rung 5's.**
   BS:512 marks proxy hygiene "(streamer rung)"; PA:77 says the strip
   set, the normalization, and the write timeouts "are all shell-level".
   A reader of the build spec alone would defer all of it to rung 12.

6. **No config key names the one owned prefix.** B3 makes the auth
   boundary a configuration decision, and the `serve` reserved class
   holds a bind-address key and a port key.

7. **Three integration classes the build spec names have no reserved
   class**, and rung 1b deliberately declined to declare them (R13). The
   forward-auth front exists in the tree only as text inside two
   supersession messages.

8. **The revocation epoch and 2.0's hook 5 disagree on the key.** B5
   keys the epoch on the email; S20:367-369 says per-user rows key on the
   surrogate id and never on an address. Both are correct for their
   moment, and nothing records the re-key as rung 6's obligation.

9. **The salvage roster loader fails open.** A missing contributors file
   returns an empty roster and a nil error
   (`_salvage/internal/config/contributors.go:33-39`). Adopting it
   unchanged would let a deployment with a missing roster start and
   refuse every login as unknown.

10. **The salvage server's single 30s write timeout would break SSE and
    long audio.** `_salvage/internal/serve/serve.go:599-602` and
    `:628-633`. B12's per-route requirement exists because of exactly
    this, and rung 5 sets the policy even though rung 12 owns the routes
    that need it.

11. **The state-dir agreement gate is hard-coded to the shadow unit.**
    `scripts/check.sh:154-167` names `systemd/dubplate-shadow.service`
    literally. A serve unit adds a fourth copy the gate cannot see.

12. **The gate's web leg needs Node and the repo forbids skip guards.**
    Five of DL's ten gates read a built bundle, the gate rule says a
    check either gates or does not exist, and Node ships to no box. The
    workstation requirement has to be explicit.

13. **The magic-link design's `/api/*` routing rule was a blocker in both
    reviews.** MD:59-60 hands `/api/*` to dubplate; SEC:20 and UNI:7 both
    file it as collision with the streamer's own API. B3's one-owned-prefix
    rule resolves it, and a reader working from MD alone would not know.

---

## Candidate rulings

Each is applied as its recommendation and marked provisional. The five
marked **product fork** are repeated in the closing section.

**C1. The roster's source is a config-supplied file with its own key.**
Rung 5 declares `DUBPLATE_ROSTER_FILE`, rewrites the
`DUBPLATE_CONTRIBUTORS_FILE` supersession message to name it, and states
that rung 6's store may become the roster's home later without changing
the loader's interface. **Product fork**: it is the file an adopter
edits to add a family member.

**C2. The roster loader fails closed.** A missing or unparseable roster
file is a load error naming the path, superseding the salvage behavior.

**C3. `streamer_username` is written as required-shaped from the
start.** It stays optional in the schema per B9, and rung 5 validates it
as the stable principal wherever present, so 2.0's promotion to required
costs no migration.

**C4. The revocation epoch keys on the normalized login address, and the
re-key is recorded as rung 6's obligation.** The spec states it as the
one per-user record deliberately keyed on an address.

**C5. The pending identity-source row retires into a real selector
key.** Rung 5 declares the selector, removes the `Pending` row in the
same commit, and states the mechanism in the close.

**C6. `doctor` gains an active-check row kind.** B2's boundary proof and
B8's client-IP mode report land as `doctor` rows whose status comes from
a probe rather than from `config.Load`. The alternative, a separate
verb, is declined because the birth requirement names `doctor`.

**C7. SMTP is the only transport rung 5 builds.** The seam takes one
method with `ctx` first. JMAP and any Worker-based transport are
declined on universality, with the Email Routing verified-destination
constraint recorded as the concrete reason the estate's own Worker
cannot serve. **Product fork**: it decides what an adopter configures to
receive mail.

**C8. `github.com/wneessen/go-mail` is adopted, with evidence.** The
plan records license, release cadence, open-issue age, and importer
count per CLAUDE.md's adoption bar. `net/smtp` is declined because it
carries no message construction.

**C9. A magic-link send failure is user-visible.** The seam keeps its
best-effort contract for every other caller, and the login handler
treats a send error as a failed login.

**C10. Auth state files are 0600 under a 0700 leaf.** The shadow store's
0644 default is the wrong inheritance for identity state.

**C11. The state-dir agreement check is generalized over greenfield
units.** `scripts/check.sh` derives the unit list rather than naming
`dubplate-shadow.service`.

**C12. Node and npm join the gate's hard-required tools.** The web leg
fails by name when they are absent, never skips.

**C13. Rung 5 declares the two theme keys and the branding config.**
DL R3.3's match and palette keys, R4.3's branding values, and R4.4's
operator CSS path are declared in rung 5's own file with `KindString`,
since rung 5's serve layer is what reads them.

**C14. The rung splits into four passes** as section 10.2 proposes, with
ceilings 3M, 4M, 4M, and 3M, and the worktree released at each pass
close.

**C15. The review shape is per-pass domain reviewers plus a two-lens
fresh-context cycle at 5d.** `web-auth-security-reviewer` runs inside
5b, the Svelte pair inside 5c, and 5d runs security and universality
lenses over the whole subsystem.

**C16. Session lifetime, link expiry, and the rate limits ship as
thresholds with the design's own defaults**: 15 minutes for a link, 90
days rolling for a session with an absolute cap the spec names, and on
the order of five link requests per hour per address. **Product fork**:
every one of these is felt by a family member.

**C17. Enumeration stays kind by default.** BS:487 makes it a flag, and
the design's reasoning holds: the roster is a handful of addresses and a
mistyped email deserves an answer. **Product fork**: it is what the
login page says to a stranger.

**C18. The first-admin question is documented rather than enforced.**
MD:130-133 records that the estate's admin predates SSO and that fresh
adopters get the ordering in a runbook; SEC:100 files "first user
becomes admin" as documentation rather than a control. Rung 5 builds no
admin bootstrap. **Product fork**: it decides what a new adopter does
first.

---

## For Geoff: the product forks

Five of the rulings above change what a person sees, configures, or is
asked for. Each carries a recommendation.

1. **Where the roster lives** (C1). Recommendation: a YAML file at a
   configured path, the same shape as today's `config/contributors.yaml`,
   with the store as a later home.
2. **The mail transport** (C7). Recommendation: SMTP only, five config
   values, any provider. The estate's own Cloudflare Worker cannot serve
   magic links because Email Routing sends only to verified destinations.
3. **Session lifetime and link expiry** (C16). Recommendation: 15-minute
   links, 90-day rolling sessions with an absolute cap, five link
   requests per hour per address, all as thresholds an adopter can change.
4. **What the login page says to an address that is not on the roster**
   (C17). Recommendation: keep the kind answer that points at the
   configured admin contact, with the flag available for an adopter who
   wants silence.
5. **Who becomes the first admin** (C18). Recommendation: rung 5 builds
   no bootstrap and the runbook carries the ordering, which is the
   position both the design and the security review already took.
