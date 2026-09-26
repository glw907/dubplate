# Rung 5, Pass 5d (The Boundary, the Deployment, and the Close) Implementation Plan

Status: **revision 1, drafted 2026-09-25 for `spec-plan-review`; no ruling answered.** Geoff reviews it
after the three plan lenses and their fold. It executes in lane 1 (the main tree) after 5c-iv closes,
which has not run yet, so Task 0 re-verifies every pointer against the tree 5c-iv leaves.

Token ceiling: **8.5M execution, flag 6.8M** (Ruling 1; arithmetic in "Token ceiling"). Checkpoint
interval: STATUS at every segment boundary, at most three tasks apart. Segments: **A** (Tasks 1, 2, 3),
**B** (Tasks 4, 5, 6), **C** (Tasks 7, 8, the adversarial cycle, Task 9), **Sitting** (Task 10,
attended, under Ruling 3), **Close** (Task 11). Planning spend is recorded against the rung, outside the
ceiling.

Sources. Code pointers were resolved at `4c766c3` on master, with warm uncommitted edits from the
5c-iii fold under `web/`, `Makefile`, and `cmd/dubplate-server/rendered_test.go`; nothing this plan
cites was among them. Box facts were read once, read-only, on 2026-09-25.

- The rung 5 spec (`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`, "the spec" below): the
  pass-split table's 5d row (`:2519`), R6 (`:363-392`), R12 (`:454-462`), R21 and R22 (`:566-577`),
  R27 (`:612-624`), "The trust boundary" (`:795-903`), "Integration classes" and "The environment
  template" (`:1635-1686`), "The state directory" (`:1690-1733`), "Deployment" (`:1737-1779`), the
  threat model (`:1950-2009`), "Errata the close applies" (`:2013-2463`), the staleness pre-task
  (`:2626-2700`), risks 10 and 11 (`:2751-2762`), and "Process" (`:2799-2837`).
- `docs/STATUS.md`'s carried "No deploy until 5d" item, and its "To rung 12" and "To rung 14" items.
- `docs/HISTORY.md`: 5a's forward carries (findings 22 and 23, R-G, R-I; `:3857-3885`, `:4045-4056`),
  5b-ii's disposition (major 5, minors 9 and 14; `:2244-2253`, `:2405-2413`), 5c-i's runbook carries
  (`:1489`, `:1729-1731`), and 5c-ii's declared gaps (`:555-608`).
- The 5c-iii plan (`docs/superpowers/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`): its
  scope wall (`:60-65`) and Task 11's handoff to 5d (`:1768-1777`).
- The rung 6 spec (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`): the 6c lane
  condition (`:1062`, `:1173-1176`), 6d's criteria 13 and 14 (`:1115-1117`), the settings metadata
  (`:729-736`), and "Not verified" (`:1263-1265`).
- The streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, which still says 2.0
  for the 0.8 streamer): "Forward hooks" (`:683-768`) and "Prior art" (`:770-806`).
- The two login-spec reviews (`docs/superpowers/research/2026-09-02-login-spec-reviews/security.md`,
  SEC, and `universality.md`, UNI), and the rung 5 security review records under
  `docs/superpowers/research/` (5a, 5b-i, 5b-ii, 5c-i, 5c-ii, 5c-iii), the do-not-re-litigate list.
- The translit plan (`docs/superpowers/plans/2026-09-25-translit-table-at-setup.md`, approved), which
  also edits `scripts/deploy.sh` and `cmd/dubplate-server/doctor.go` from lane 2.

**What this pass is.** Rung 5's last pass. It gives the serve shell its two remaining configuration
surfaces, gives `doctor` rows whose status comes from a probe of the running server, writes the
deployment (the serve unit, its grant, `deploy.sh`'s pre-creates and clean web build, the environment
template's rows, and a generalized state-dir check), writes the operator runbook and the trust-boundary
documentation, maps all 50 SEC and UNI findings to where each was absorbed, runs the rung's full
fresh-context adversarial cycle over the whole subsystem, and closes the rung. Under Ruling 3's yes it
ends with the first production deploy of the greenfield server, as one attended sitting.

**What this pass is not.** It builds no streamer driver, reverse proxy, fallthrough, SSO header, or
roster sync (rung 12). It mounts no API and no bearer channel (rung 14). It builds no revoke verb and
does not close review minor 14 (epoch monotonicity across processes), which the rung 6 spec gives to
6d's criterion 8. It takes none of STATUS's "To rung 12" items (the streamer-admin mapping, the shared
origin M2, Ruling 7's availability question, the `streamer_username` re-validation, the header step's
fallthrough reconciliation, the streamer database in a restore) and none of the "To rung 14" items (the
projection's mount, the production authorizer, the bearer channel, Ruling 3's 403, security L5 to L7).
It does not take L5 (a roster-POST throttle) or L6 (an n5 whole-roster doctor row), which stay
unassigned in STATUS. It adds no Go module dependency and no web dependency. **It edits nothing under
`internal/config`** (see "Global constraints"). It writes nothing under `web/`, `internal/auth`
(except under Ruling 2's no), `internal/identity`, `internal/notify`, `internal/engine`, `internal/ops`,
`internal/backup`, `compose.yaml`, `config/` (the example roster already landed at 5b-ii), `corpus/`, `tools/`, or `_salvage/`. **Nothing deploys and nothing writes the box before Task
10**, and Task 10 exists only under Ruling 3's yes.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the pass-execute chain per task,
> sequential within each segment. The implementer is `dubplate-implementer`. The conductor consumes
> structured reports and never reads a diff, a test log, or a gate transcript inline. Subagents never
> dispatch agents, so `code-simplifier`, the two adversarial lenses, the prose reader, and the
> architecture reads are the conductor's dispatches. The conductor owns `docs/STATUS.md`,
> `docs/HISTORY.md`, `ROADMAP.md`, the lane registry, and the contested-surface rows. **This pass runs
> in lane 1, the main tree.** The live-executor check is the lane registry, never `pgrep -f
> Projects/dubplate`.

---

## Rulings

Three yes/no questions, each with a recommendation and what each answer builds. Everything else this
plan decides is in "Resolutions this plan takes", each with its reason, because each has one
architecturally correct answer (Geoff, 2026-09-25: "You don't need to ask questions for the sake of
ceremony"). The rung 5 spec's rulings, the 5a through 5c-iv plans' rulings, and the rung 6 spec's
rulings stand and are not reopened.

**1. Does this pass run under an 8.5M execution ceiling (flag 6.8M)?** Recommendation: yes.

**Reason:** "Token ceiling" derives a 6.72M projection from 5c-i's per-task actual, the most recent
rung 5 Go pass with a security read and fold, plus five named adders. The ceiling is the projection
divided by 0.8, rounded up to the build's 0.5M grain, so the flag sits at the projection and marks an
overrun rather than firing on plan. The spec's 3M (R23) predates every rung 5 actual and is superseded
by the re-derivation R23 itself requires ("Each ceiling is re-derived from the previous pass's actual
report"). **What yes builds:** the pass as written, ten tasks plus Task 0, one pass. **What no builds:**
the same tasks under whatever ceiling Geoff names; below about 6.7M the conductor asks at the Task 6
boundary whether to split the pass there (the named cut: 5d-i is Tasks 0 to 6, 5d-ii is Tasks 7 to 11),
since the build tasks and the review-and-close tasks are file-disjoint at that seam. Under Ruling 2's
no, the ceilings are those in "Token ceiling", "Under Ruling 2's no".

**2. Does the forward-auth verify endpoint move from 5d to rung 12?** Recommendation: yes.

**Reason:** four facts, each in the tree or a ratified document. First, it has no consumer before rung
12: no adopter exists before launch (rung 16), and the estate reaches serve through the tunnel as
ingress, never through a forward-auth proxy. CLAUDE.md's 2026-09-12 refinement says a hook "is built by
the rung that first consumes it, and nothing is exported ahead of a caller." Second, the endpoint's
answer is driver knowledge the spec's own "Universality" section gives to rung 12: "The header name,
the injection rules, the bypass paths, the logout wiring, and the post-login destination live with the
driver at rung 12." A 5d endpoint would pick its answer header (`Remote-User` is the field's usual
name) with no driver to agree with. Third, STATUS carries "no non-admin maps onto a streamer admin" to
rung 12, and a verify endpoint in front of a streamer that auto-creates accounts from a trusted header
opens exactly that path before its guard exists. Fourth, an unguarded verify endpoint is an identity
oracle (R21's own words), so deferring it removes the one new anonymous-reachable identity surface from
the first production deploy. Rung 12 keeps the full adversarial cycle (CLAUDE.md, "Review scales to
rung risk"), so the endpoint is not reviewed less by moving. **What yes builds:** no verify endpoint,
no forward-auth verify integration class, and no keys (declaring keys for code that does not exist is
what the reserved-class mechanism is for, and rung 1b's contract forbids a reserved class whose keys
the rung that declares it will not build). The close writes a spec erratum moving the 5d row's first
clause and 5b-i's moved "a verify subrequest from such a peer is refused" assertion to rung 12, adds one
line to STATUS's "To rung 12" item, and the trust-boundary documentation (Task 7) states the forward-
auth topology as a future integration with the requirements BS:504-514 already fixes. Ten tasks, one
pass. **What no builds:** Task 4V (below, "Task 4V (under Ruling 2's no)") at `opus`, eleven tasks,
and the pass splits at the Task 6 boundary into 5d-i (Tasks 0 to 6 with 4V) and 5d-ii (Tasks 7 to 11),
per the ten-task rule, with the ceilings in "Token ceiling".

**3. Does 5d end with an attended first-deploy sitting that mints the two HMAC secrets, installs the
serve unit disabled, bootstraps the admin, and lifts the no-deploy carry?** Recommendation: yes.

**Reason:** the spec's exit gate asks for "`deploy.sh --check` rendering with no unresolved name and
`dubplate-server doctor` exiting 0 against the rendered environment." That is impossible while the two
magic-link secrets are commented pending rows, as the spec's template section planned ("rung 16 mints
all three"): R29 makes `Load` require every Required key of the selected source, so `doctor` exits 1
naming them. It is worse on the box: the live shadow timer (`dubplate-shadow.timer`, active) calls
`config.Load` against the same registry (the spec's own cross-rung note, `:1674-1679`), and the box's
`/etc/dubplate/server.env` today carries no `DUBPLATE_IDENTITY_SOURCE` row at all (read 2026-09-25: ten
rows, the rung 4 set). Any deploy of a binary built after 5b-i therefore breaks the live shadow unless
the template carries the whole magic-link set with real values. Without the mint, the no-deploy carry
lasts to rung 16, and the translit pass's staged deploy (due before packaging) and every shadow binary
update wait with it. The SMTP password is already minted under its exact name (`DUBPLATE_SMTP_PASSWORD`,
registry row, 2026-08-31), so only the two HMAC keys are new. Minting needs one 1Password approval, which
is Geoff's touch, so it belongs in the attended sitting rather than the unattended pass. The spec
already names the 5d box deploy as an owner touchpoint (`:2835-2836`, risk 10). **What yes builds:**
Task 6 writes the two secrets as bare rows, and Task 10 runs the sitting script in "Task 10". **What no
builds:** Task 6 writes the two secrets as commented pending rows, Task 10 does not exist, the exit
gate's box assertions become Task 6's gate-resident proof alone, and the no-deploy carry moves to rung
16 with a STATUS line naming the live shadow's dependency on it.

---

## Primary inputs

The implementer reads, per task, only what the task's Files and Outcomes name, plus:

- `CLAUDE.md` in full (stack rules, the gate, safety rails, "Facts agents get wrong").
- `go-conventions` before any Go; `vps-conventions` before `systemd/`, `scripts/deploy.sh`, or
  `scripts/check.sh`; `golang-spf13-cobra` before any command wiring; `writing-voice` (Google developer
  documentation standard) before the runbook.
- `.claude/instructions/simplifier-brief.md`, pasted verbatim into every `code-simplifier` dispatch.

---

## Goal

After 5d, the serve shell reads a trusted proxy's client-address header only when an operator
configured one and only from that proxy's address; `doctor` reports the live client-IP mode, the web
surface's startup resolution, the configured bind addresses against the host's interfaces, and a live
boundary probe that dials the running listener and says plainly when none answers; the repo carries a
hardened greenfield serve unit whose grant names exactly the auth leaf and the roster leaf, a
`deploy.sh` that pre-creates both, installs the unit without enabling it, and refuses to ship a binary
built from a dirty `web/`; a state-dir check that derives its expectations from every greenfield unit
rather than naming one; a template whose render passes `doctor`; an operator runbook; a disposition
table covering all 50 SEC and UNI findings; a folded two-lens adversarial review; the rung's errata
applied; and, under Ruling 3's yes, the greenfield binary and unit on the box with the admin
bootstrapped and `doctor` green.

---

## Owner touchpoints

The three rulings, at the plan-approval gate. Under Ruling 3's yes, the Task 10 sitting: Geoff approves
the one 1Password read that mints the two secrets, supplies his own display name and streamer username
for the admin bootstrap (the streamer username should equal his Navidrome login, since rung 12 maps it),
watches the quiet-box check, and approves the apply after reading `deploy.sh --check`. What would become
a touchpoint: any Go module or web dependency; any change to `compose.yaml` or the tunnel; any write to
the box outside Task 10; a finding from either adversarial lens the fold would refuse on a
correctness-critical claim (the conductor's overrule then takes CLAUDE.md's `xhigh`-then-`fable` path
and records it).

---

## Token ceiling

**Basis.** 5c-i's recorded actual (`docs/HISTORY.md:1692-1698`): about 5.3M in subagent tokens across
seven build tasks, which amortizes Task 0, the fix rounds, the simplifier round, one register gate, one
pass-end `web-auth-security-reviewer` read, and its two-task fold. That is **0.76M per build task**,
review and fold included. 5c-i is the right basis: the most recent rung 5 pass whose surface is Go in
`cmd/dubplate-server` and `internal/serve` plus a security read, which is 5d's shape. 5c-ii's 1.15M per
task carried a new web tree and three reviewers, which 5d does not; 6b-ii-a's 0.65M carried no security
read.

1. Seven build and document tasks (1, 2, 3, 5, 6, 7, 8) at 0.76M: **5.32M**. Task 4 does not exist under
   Ruling 2's yes; Task 9's fold is inside the per-task rate, as 5c-i's fold was.
2. The second adversarial lens (universality), which 5c-i did not run: **+0.35M**.
3. Both lenses read the whole rung 5 subsystem (four packages, `cmd/dubplate-server`'s auth and serve
   files, the unit, and `web/`'s three pages), not one pass's diff: **+0.25M**.
4. The heavy gate lane on every task (Chromium, about 3-4 min and about 560 MiB per full gate since 5c-iii
   Task 3a), which lengthens each fix round's wait and report: **+0.30M**, 5c-iii's own adder for the
   same cause.
5. Task 11's close drafts plus two `go-architecture-reader` dispatches (`internal/serve`,
   `cmd/dubplate-server`) and the sitting's evidence subagents under Ruling 3: **+0.50M**.

**Projection 6.72M.** A **8.5M** ceiling puts the flag at **6.8M**, 0.08M above the projection. Rung 5's
reservation moves from about 41M (5c-ii's close) by 5.5M over the spec's 3M line for 5d, to about 46.5M.

**Under Ruling 2's no.** Task 4V adds one task at `opus`, so the pass splits at the Task 6 boundary.
5d-i (Tasks 0 to 6 with 4V): six tasks at 0.76M (4.56M), the heavy lane (+0.30M), and one close draft
with two architecture reads (+0.25M): projection 5.11M, **ceiling 6.4M, flag 5.12M**. 5d-ii (Tasks 7 to
11): two tasks at 0.76M (1.52M), the two lens adders (+0.60M), the close and the sitting (+0.50M):
projection 2.62M, **ceiling 3.3M, flag 2.64M**. Together 7.73M.

**Deliverable count:** Tasks 1, 2, 3, 5, 6 carry four each; Tasks 7, 8, 9, 11 carry three or four;
Task 10 is a conductor-run sitting. None passes four.

---

## Global constraints

- **Lane 1, the main tree, on master**, after 5c-iv's close is committed and the lane registry shows
  lane 1 free with this pass named. No task reads, writes, or runs a command inside
  `.claude/worktrees/`. The spec's "Worktree, released at the close" for 5d is superseded by the lane
  registry; the close records it.
- **`internal/config` is untouched, stated plainly for the 6c lane condition.** The rung 6 spec runs 6c
  in lane 2 "only if 5c and 5d leave `internal/config` untouched" (`:1062`). 5d declares its integration
  class in its own file in `cmd/dubplate-server` (CLAUDE.md: "Config keys are declared through the
  registry in each rung's own file"), through the existing `AddIntegrationClass` (`internal/config/
  registry.go:213`), with no reserved class to retire (`keys.go` carries no trusted-proxy class; read at
  `4c766c3`). `config.IntegrationClass` has no value-validation hook (`internal/config/keys.go:184-190`),
  so value validation lives in `cmd/dubplate-server`'s resolver, which serve startup and `doctor` both
  call, on R-5cii12's precedent for the theme keys. Rung 6's own plan adds a per-key value validator to
  the registry at 6c (`:729-736`); moving 5d's checks onto it is 6c's or the architecture cleanup pass's,
  never 5d's. `internal/config` stays byte-identical across the pass, checked at the close by `git diff
  --stat <5d base>..HEAD -- internal/config` answering empty.
- **Package territory, stated as a wall.** This pass writes: `internal/serve` (`clientaddr.go`,
  `strip.go`, a new probe file, and their tests); `cmd/dubplate-server` (a new `proxy.go`, `serve.go`,
  `doctor.go`, a new doctor-probe file, `main.go`'s registration line, and their tests, plus
  `templatekeys_test.go`); `systemd/` (one new unit); `scripts/deploy.sh`; `scripts/check.sh`;
  `env/dubplate-server.env.template`; `docs/server-runbook.md` (new); the disposition table and review
  records under `docs/superpowers/research/`; the close drafts; and, at the close only, the errata
  targets. Under Ruling 2's no, Task 4V adds `internal/auth` (one handler file and its test) and
  `cmd/dubplate-server` (a new `forwardauth.go`). A task that believes it needs anything else stops and
  reports.
- **No Go module or web dependency.** `go.mod`, `go.sum`, `web/package.json`, and
  `web/package-lock.json` stay byte-identical. `golang.org/x/net/http/httpguts` is declined for header-
  name validation (R26 fixed rung 5's dependency count at two); a token check against RFC 9110's `tchar`
  set is a few lines.
- **No deploy before Task 10.** No task runs `scripts/deploy.sh` in apply mode and no task writes the
  box. Task 6 and Task 10's preparation may run `scripts/deploy.sh --check`, which the script documents
  as reporting every step without applying, and the box stays read-only (`ssh -o BatchMode=yes` for a
  read). The incumbent `/usr/local/bin/dubplate` and its units are never touched.
- **Assertion discipline at the repo's bar.** No vacuous assertion, no expectation derived from the code
  under test. Every hard-coded expectation cites the spec, a quoted man page or `go doc`, or a measured
  probe. Every test names the fixture state it needs and asserts that state before the property. Every
  named mutation is proved to fail the gate and then reverted.
- **Skills before code**, as "Primary inputs". No process citation in a shipped comment. The em dash is
  banned in Go comments.
- **Universality.** No estate hostname, address, or path in Go code or a Go fixture; fixtures use
  `example.test` and the documentation ranges (192.0.2.0/24, 198.51.100.0/24, 2001:db8::/32). Estate
  values live in the template, the unit, and `deploy.sh`, which already carry them.
- **The gate is `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`**, re-issued
  on exit 75 until it prints `gate exit:`, **in the default heavy lane for every task: never set
  `CAIRN_GATE_LANE=light`.** Every full gate launches Chromium and runs about 3-4 minutes at about 560
  MiB. Evidence runs use `-count=1` and report wall time. Specific files only. Commit footer
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` for an `opus` task, the dispatch's footer
  otherwise.

---

## Privileged fields and their channels

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| The trusted-proxy set and header name | `DUBPLATE_TRUSTED_PROXY_ENABLED`, `DUBPLATE_TRUSTED_PROXY_ADDRS`, `DUBPLATE_TRUSTED_PROXY_HEADER`, read once from the environment at serve startup through Task 1's resolver; `doctor` reads the same resolver. | No request component reaches the resolver. A value the resolver refuses stops serve before any listener opens (`newServeCmd`'s own "refuses at startup" contract, `serve.go:162-171`). |
| The reported client IP (R27) | `RemoteAddr` always. The configured header only when the peer address, compared as a `netip.Addr`, is in the configured set, and only its rightmost hop that parses as an address and is not itself a configured proxy. | A header from any other peer is never read (`clientaddr.go:107-109`). A malformed rightmost untrusted hop ends the scan and yields the peer address, so no hop left of it is ever reached (Task 1). Group 2 strips every forwarded-address header from an untrusted peer (`strip.go:460-466`). |
| The peer address (the only input to a trust decision) | `RemoteAddr` alone (`clientaddr.go:62-74`). | No header, no configuration. Unchanged by 5d. |
| The front exemption (one identity header a front may keep) | None at 5d. `newServeShell` passes `serve.FrontExemption{}`. | A Task 1 test asserts the shell's strip carries the zero exemption, so every identity spelling is stripped on every path from every peer. |
| The boundary probe's answer | The probe route, answering only a peer whose address is loopback or one of the configured bind addresses (a local dial); every other peer meets the terminator's 404. | The answer carries counts only, never a header value or an address. An internet client through the tunnel arrives from the proxy's own address, which is neither, so it sees the 404 (Task 3). |
| The two state leaves' ownership and mode | `deploy.sh` (as root on the box) creating each with owner `music`, mode 0700. | `auth.Open` refuses a leaf wider than 0700 (`internal/auth/store.go:53-71`, `dirPerm = 0o700` at `:32`); the roster loader refuses a group- or other-writable file (`roster.go:318-319`). |
| The two HMAC secrets and the SMTP password | The workstation age store, rendered by `deploy.sh` into `/etc/dubplate/server.env` (0600 root:root). | No other channel. No value appears in the repo, a log, a test, or a report; checks are name-only. |
| (Ruling 2's no) The verify answer's identity header | Task 4V's handler alone, from the seam's roster-resolved principal, only to a permitted peer. | Every inbound spelling of it is stripped by group 1 before the handler; a non-permitted peer meets the terminator's 404. |

---

## Resolutions this plan takes

Each is the plan's own decision with its reason. The adversarial lenses may challenge any of them.

**R-5d1. The trusted-proxy class's shape.** Name `trusted proxy`; flag `DUBPLATE_TRUSTED_PROXY_ENABLED`;
Required `DUBPLATE_TRUSTED_PROXY_ADDRS` (comma-separated addresses) and `DUBPLATE_TRUSTED_PROXY_HEADER`
(one header name). Declared in a new `cmd/dubplate-server/proxy.go`, registered from
`newProductionRegistry` (`main.go:108-117`). **Reason:** R22 names exactly this shape (an enable flag
plus the proxy addresses and the client-address header, required when enabled); the file follows
`backup.go`'s precedent (`:40-54`).

**R-5d2. Address grammar.** Each entry is trimmed and parsed with `netip.ParseAddr`, then unmapped. `go
doc net/netip.ParseAddr`: "The string s can be in dotted decimal ("192.0.2.1"), IPv6 ("2001:db8::68"),
or IPv6 with a scoped addressing zone". `go doc net/netip.Addr.Unmap`: "if ip is an IPv6 address
wrapping an IPv4 address, it returns the wrapped IPv4 address." Unmapping matters because `go doc
net/netip.Addr.String` warns "unlike package net's IP.String method, IPv4-mapped IPv6 addresses format
with a "::ffff:" prefix", while `net.IP.String` (the form `RemoteAddr` carries) prints "dotted decimal
... if ip is an IPv4 or IP4-mapped IPv6 address". Refused, each naming the entry: an empty entry, a
value that does not parse (a hostname), a network in CIDR form (reason: "an address, not a network";
the spec pins trust to observed addresses, never a subnet, `:839-840`), an address with a zone, and an
unspecified address. **The four states:** flag absent or false with no Required key set is mode direct;
flag true with `ADDRS` absent or empty is `Load`'s own missing-key report for an enabled class
(`config.go:374-387`; `lookupEnviron`'s rule makes an empty value unset); a malformed entry is refused
naming it; many bad entries are joined with `errors.Join` so every one is named at once. A duplicate
entry collapses into the set and is not an error.

**R-5d3. Header grammar.** The header name must be an RFC 9110 token (letters, digits, and
`!#$%&'*+-.^_`|~`), since `go doc net/textproto.CanonicalMIMEHeaderKey` says a key with "invalid header
field bytes as defined by RFC 9112 ... is returned without modifications", which would leave a
malformed name uncanonicalized and unmatched. Refused by name, with a reason each: `Cookie`, `Origin`,
and `Sec-Fetch-Site` (group 3's keep set, `strip.go:201-205`; adding one to group 2 would strip it from
every untrusted peer and break sessions and the origin check); `Host`, `Connection`, `Content-Length`,
`Transfer-Encoding`, and `Authorization` (end-to-end or framing headers no front reports an address in);
`Forwarded` (RFC 7239's `for=` grammar is not an address list, so every hop would fail R-5d4's parse);
and any identity spelling group 1 strips (`strip.go:11-63`, both forms), since group 1 would delete it
before `ClientIP` reads it. Comparison is case-insensitive through canonical form, and an underscore
spelling of a refused dash name is refused too. The same four states as R-5d2.

**R-5d4. The hop rule (review minor 9, SEC finding 22).** `ClientIP` scans hops right to left as today
(`clientaddr.go:121-135`), but each non-empty hop must parse with `netip.ParseAddr` and carry no zone.
A configured proxy is skipped; the first hop that is not a proxy is returned in unmapped
`netip.Addr.String` form; **a hop that fails to parse ends the scan and returns the peer address**, so an
attacker who plants a malformed value cannot make the scan continue leftward into values they chose.
`TrustedProxyConfig`'s set becomes `map[netip.Addr]struct{}` (5a's forward carry 1), and `PeerAddress`
is compared after the same parse and unmap. A `RemoteAddr` that does not parse is never trusted.

**R-5d5. Doctor's four new rows print unconditionally, and both line-count tests take a named
constant.** R6 requires the plan to say which way the collision goes. The rows are `client ip`, `web`,
`bind addresses`, and `boundary probe`, printed after the backup rows. They are not registry
declarations, so registering them through the registry would invent a family for four fixed rows; the
store row set this precedent (`storeRowLines`, `doctor_test.go:217-235`). `TestDoctorCmd` and
`TestDoctorCmd_LineCountFollowsRegistry` each add a `probeRowLines` constant equal to 4, hard-coded,
never derived from the rows' own slice.

**R-5d6. The web row runs serve's own web resolution.** STATUS's "the theme keys' doctor row" and the
5c-ii erratum ("`doctor` prints the configured value without judging it; that gap is carried to 5d")
close by calling `newWebRoutes` (`web.go:218`), the exact function serve startup calls
(`serve.go:303-306`), with the same inputs, reading the bundle and never writing. It judges the theme
match, the palette, the branding value, and the operator CSS file in one place, so doctor and serve
cannot disagree. A refusal reads `invalid: <error>` and fails doctor. When the prefix or the site name
cannot be resolved (so `Load` already failed), the row reads `not checked` and adds no failure of its
own. **Reason:** a per-key validator map would be a second copy of R-5cii12's checks; the security
review's minor 12 class ("`doctor` green for sign-in-breaking configurations") is exactly a doctor that
checks less than startup does.

**R-5d7. The bind-addresses row.** For each configured address (`serve.SplitBindAddrs`,
`bindaddrs.go:37-56`): `present` when it equals the IP of an interface address from `net.InterfaceAddrs`
("returns a list of the system's unicast interface addresses") or is a loopback address; `not on any
interface` otherwise, which fails doctor. **Reason:** R20's `IP_FREEBIND` makes the bind succeed for an
absent address, so the listener raises no error and the spec gives this row the job (`:1601-1607`). A
loopback address counts present because Linux routes 127.0.0.0/8 to `lo` while `InterfaceAddrs` lists
only `127.0.0.1/8`. The four states: bind threshold unset is the registered default (loopback) and
`present`; empty or malformed is `SplitBindAddrs`'s refusal as `invalid: ...`, failing; one absent
address fails naming it; many absent addresses are all named on the one row.

**R-5d8. The boundary probe observes through a local-only probe route.** The spec leaves the probe's
observation unstated. A route under the owned prefix, named `boundary-probe`, answers `GET` with a
fixed JSON object of three counts: identity headers that reached it (group 1), forwarded-address
headers (group 2), and a probe-owned sentinel header the request listed in `Connection` (group 3), with
`Cache-Control: no-store`. It answers only a peer that is loopback or one of the configured bind
addresses; every other peer, and every other method, meets the terminator's 404 through the same
`exactPath`/`methodOnly` wrapping the other routes use. The probe client and the route live in one
`internal/serve` file, so the header lists stay unexported and the client can never send a set the
route does not count. `doctor` dials each configured address on the configured port with
`net.Dialer.LocalAddr` set to that same address (`go doc net.Dialer`: "LocalAddr is the local address to
use when dialing an address ... If nil, a local address is automatically chosen"), so the peer the shell
sees is known, with a 2-second timeout. **Reason:** the spec's R6 wants a live dial that "cannot fail
for any deployment reason" to be replaced by one that can; only a response the running process computes
after its own strip can tell doctor whether the strip ran. Restricting it to local peers keeps the
tunnel from exposing a trusted-proxy configuration's shape.

**R-5d9. The probe row's states, each named.** Per address: `refused` (all three counts zero; group 2
may be nonzero only when the dial address is itself a configured trusted proxy, reported as `group 2 not
provable from a trusted proxy address`); `no listener` when the dial is refused or times out, which is
**not** a failure, since the runbook runs doctor before the unit starts and the sitting runs it with
the unit stopped; `passed N identity header(s)` or `passed the Connection-listed sentinel`, failing;
`unexpected answer: <status>` for any non-200 or unparseable body, failing. The row ends with `proves
this server's strip, not a proxy's` (R6: "it cannot prove an adopter's proxy strips one"). When the
bind configuration is invalid the row reads `not probed` and adds no failure beyond R-5d7's.

**R-5d10. The unit's name, type, and body.** `systemd/dubplate-server-serve.service`: it names the
binary and the verb, and it is distinct by more than one letter from the incumbent
`dubplate-serve.service`, which stays untouched until rung 16 (spec `:1739-1742`). `Type=exec`: `man
systemd.service`, "systemctl start command lines for exec services will report failure when the
service's binary cannot be invoked successfully", which `Type=simple` does not. The body inherits the
shadow unit's hardening block verbatim minus nothing (the shadow already dropped `AF_NETLINK`,
`dubplate-shadow.service:84-89`), `User=music`, `UMask=0027`,
`EnvironmentFile=/etc/dubplate/server.env`, `ExecStart=/usr/local/bin/dubplate-server serve`,
`Restart=on-failure`, `RestartSec=5`, `StartLimitIntervalSec=300` with `StartLimitBurst=30` (the
2026-09-01 ruling the incumbent records), `OnFailure=dubplate-ping-fail@serve.service` wired and inert,
`After=docker.service` without `network-online.target` (R20), `TimeoutStopSec=15` argued from a request
in flight on a page a person is watching (serve's shutdown drains in-flight requests), and
`ReadWritePaths=/var/lib/dubplate/auth /var/lib/dubplate/roster` with no `-` or `+` prefix. `man
systemd.exec`: paths "may be prefixed with "-", in which case they will be ignored when they do not
exist", which would turn a missing pre-create into an `EROFS` at first write instead of the namespace
failure. An `[Install]` section with `WantedBy=multi-user.target` exists so a later `systemctl enable`
works, and `deploy.sh` never enables it (risk 10).

**R-5d11. The generalized state-dir check (R12).** `deploy.sh` gains one array naming every greenfield
leaf under `DUBPLATE_LIB_DIR`, `DUBPLATE_SERVER_LEAVES=(shadow auth roster)`. A pre-create loop runs
`install -d -o music -g music -m 0700` for `auth` and `roster` (R10's mode, which `auth.Open` enforces);
the shadow leaf's existing pre-create line (`deploy.sh:503-505`) stays as it is, and the array still
names it so the check covers it. `scripts/check.sh`'s state-dir leg selects every unit under
`systemd/` whose `ExecStart=` names `/usr/local/bin/dubplate-server`, and for each `ReadWritePaths=`
entry asserts: no `-` or `+` prefix; the entry is the template's `DUBPLATE_STATE_DIR` or a leaf directly
under it; and the leaf is in `DUBPLATE_SERVER_LEAVES`. It also asserts `DUBPLATE_LIB_DIR` equals the
template's `DUBPLATE_STATE_DIR` (today's second half). Each failure names the unit, the entry, and the
missing piece. **The four states:** no greenfield unit found fails (a selector that matches nothing must
not pass); a unit with no `ReadWritePaths=` passes for that unit (a read-only unit is legal); a
malformed entry fails naming it; many bad entries are all reported before the leg fails. **Reason:**
data beats a third hard-coded pair (R12), and 6d adds the `store` leaf by one array word (rung 6 spec
criterion 14).

**R-5d12. `systemd-analyze` becomes hard-required.** `scripts/check.sh:631-643` guards the verify leg
with `command -v systemd-analyze ... skipping`, which CLAUDE.md forbids ("never guard one with `command
-v || skip`"). The guard is removed and a missing tool fails by name, on the Node precedent (R13).
**Reason:** the exit gate's first assertion is "`systemd-analyze verify` clean on the new unit".

**R-5d13. `deploy.sh` builds the web bundle from a clean tree.** Before `make build-linux`, `deploy.sh`
refuses when `git status --porcelain -- web/` is non-empty (naming the paths), then runs `npm ci
--prefix web` and the web build, so the embedded bundle is the committed source's. Nothing Node or
Playwright ships to the box: the rendered leg is a workstation gate only. Under `--check` the build still
runs, on the binary step's own precedent (`deploy.sh:317-320`). **Reason:** 5c-ii's handoff ("the binary
embeds whatever `web/build` holds") and 5c-iv's Task 11 row.

**R-5d14. The template's rows.** `DUBPLATE_ROSTER_FILE=/var/lib/dubplate/roster/roster.json` and
`DUBPLATE_PUBLIC_URL=https://music.907.life` as literals (not secrets, never minted; the hostname is the
tunnel's own, `config/cloudflared-tunnel-config.json`, and it carries no path per R-5c9). The four
non-secret SMTP names (`DUBPLATE_SMTP_HOST`, `_PORT`, `_USER`, `_FROM`) and `DUBPLATE_SMTP_PASSWORD` as
**bare rows**, correcting the spec's "literal" for the four: the age store already holds all five under
these exact names (the registry's note on "NEW non-secret values (SMTP host/port/user/from ...)"), so
"none is ever minted" is stale, a bare row keeps a personal address out of a repo that goes public at
0.8, and the flat-namespace hazard that forced `DUBPLATE_STATE_DIR` to a literal
(`env/dubplate-server.env.template:38-53`) does not apply, because the incumbent and the greenfield
server use the one transport with the same values. `DUBPLATE_MAGIC_LINK_SECRET` and
`DUBPLATE_SESSION_SECRET` as bare rows under Ruling 3's yes, commented pending rows under its no.
`DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS` and `DUBPLATE_SESSION_SECRET_PREVIOUS` as commented rows (Optional
keys, used only across a rotation). `DUBPLATE_TRUSTED_PROXY_ENABLED=false` as an explicit literal,
because proxyless operation is "chosen, never fallen into" (BS:278-280); the estate's own values (the
cloudflared container's address, which Docker allocates dynamically on `musicbox_default` today, and
`Cf-Connecting-Ip`) wait for the cutover, which must first pin that container's address in
`compose.yaml`. The header's "No row here is a secret" paragraph is rewritten. The serve bind and port
rows are not added: the unit stays disabled and loopback is the right default until the cutover binds
the bridge gateway.

**R-5d15. The runbook's home.** A new `docs/server-runbook.md`, the greenfield server's operator
runbook, beside the incumbent's `docs/bringup-runbook.md`, which is not edited. **Reason:** the spec
requires a runbook and trust-boundary documentation (`:1771-1774`, `:884-903`, `:2865-2872`), the
incumbent file documents a retiring estate, and a section inside it would be deleted with it at rung 16.

---

## Review focus

5d is auth-risk and deploy-risk, so it keeps the spec's full fresh-context adversarial cycle (R24,
`:2815-2825`): two independent lenses over the whole rung 5 subsystem, `web-auth-security-reviewer`
for security and one `opus` universality reader, each given the spec's threat model (`:1950-2009`)
and Task 8's disposition table, each grading the implementation against the claims rather than against
taste, and each asked for correctness gaps ranked by consequence, not polish. The six focus points the
lenses must each address by name:

1. **A configured header changing a trust decision.** Every trust decision reads the peer address;
   the client-address header affects only rate-limit keys and logs, from a configured proxy only, and a
   malformed hop cannot steer the scan (Tasks 1 and 3).
2. **An identity header surviving ingress** on any path, from any peer, in either spelling, including
   through a `Connection` listing (Task 3's probe and the in-process strip tests).
3. **`doctor` green on a configuration serve would refuse**, or green while the listener passes a
   spoofed header (Tasks 2 and 3).
4. **A grant wider or narrower than the writes.** The unit's `ReadWritePaths` against every write the
   serve process makes, the leaves' modes against `auth.Open`'s refusal, and the namespace failure when
   a pre-create is missing (Task 5).
5. **A secret in the repo, a log, a report, or a rendered file with the wrong mode** (Tasks 6 and 10).
6. **A deploy that breaks the live shadow or starts the new unit** (Tasks 5, 6, and 10).

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched before any implementer. Reads this plan against the tree
5c-iv leaves. **Its output is blockers with evidence, not fixes.** It edits no tracked file. A moved
pointer whose content survives is a note; a pointer whose content is gone is a blocker.

1. **Lane 1 free and the tree clean** of warm changes the conductor did not author; the lane registry
   names this pass; **5c-iv's close is committed on master.** Report HEAD and 5c-iv's close commit.
2. **5c-iv's handoff to 5d** (5c-iii plan Task 11, `:1768-1777`, and 5c-iv's HISTORY entry): every
   item it names for 5d, and whether any lands outside this plan's tasks. The clean-build row
   (R-5d13) must be among them.
3. **What 5c-iii and 5c-iv changed in 5d's files:** `cmd/dubplate-server/serve.go` (the shell
   construction at `:243-315`, in particular the zero `proxyConfig` at `:244`, the strip at `:245`, and
   `clientIP` at `:283`), `cmd/dubplate-server/web.go` (`newWebRoutes`'s signature at `:218` and its
   side-effect freedom), `internal/serve/` (`clientaddr.go`, `strip.go:227-340`, `serve.go:84-116`),
   and `cmd/dubplate-server/doctor.go` with `doctor_test.go` (the row order at `:38-127`, the constants
   `storeRowLines` and `backupDoctorRowsOff`).
4. **What lane 2 merged since `4c766c3`:** 6b-ii-b (restore) and the translit pass each may have edited
   `doctor.go`, `doctor_test.go`, `main.go`, `scripts/deploy.sh` (translit's staged transition), and
   `scripts/check.sh`. Report every doctor row and line-count constant they added, whether translit's
   `unidecode/` leaf is pre-created by `deploy.sh` and granted to any unit, and any new unit under
   `systemd/` whose `ExecStart=` names `dubplate-server` (R-5d11 must cover it).
5. **Whether 6c has merged or is running**, and whether `internal/config` has changed since
   `4c766c3`. A 6c value validator changes nothing here (see "Global constraints"); report it.
6. **`scripts/check.sh`**: the systemd leg (`:631-643`), the state-dir leg (`:645-658`), and the heavy
   gate's current wall time and peak memory.
7. **`scripts/deploy.sh`**: `UNIT_FILES` (`:64`), the binary step (`:304-366`), the state-dir pre-creates
   (`:436-505`), `DUBPLATE_LIB_DIR` (`:503`), and `TIMER_UNITS` (`:525-526`).
8. **`env/dubplate-server.env.template`** still carries the ten rows read at `4c766c3` and nothing of the
   magic-link set but `DUBPLATE_IDENTITY_SOURCE`.
9. **Age-store names, name-only:** `DUBPLATE_SMTP_HOST`, `_PORT`, `_USER`, `_FROM`, and `_PASSWORD` are
   present in `~/.local/secrets`; `DUBPLATE_MAGIC_LINK_SECRET` and `DUBPLATE_SESSION_SECRET` are absent.
   Never print a value.
10. **Box, read-only** (`ssh -o BatchMode=yes root@5.78.148.42`): systemd is still 257; `music` is uid
    2000; `/var/lib/dubplate` holds only `shadow` (plus `unidecode` if translit deployed); nothing listens
    on 8080; `/etc/dubplate/server.env`'s row names; `dubplate-shadow.timer` active.
11. **Every other `file:line` this plan cites** resolves to the content described, including the spec
    pointers in "Sources" and `internal/auth/store.go:32,53-71`, `internal/auth/roster.go:318-319`,
    `internal/config/registry.go:213`, `internal/config/keys.go:184-190`, `internal/config/config.go:374-387`,
    `cmd/dubplate-server/main.go:108-117`, and `cmd/dubplate-server/backup.go:40-54`.
12. **The spec's staleness questions** (`:2641-2700`) that still apply to 5d: the state-dir check still
    hard-coded; `doctor` still deriving registry rows; `.gitleaksignore` still empty or its appends named.

---

## Segment A

## Task 1: The trusted-proxy integration class and the address grammar

Model: `opus` (the trust input to every rate-limit key; novel validation the plan specifies by outcome).

**Files:**
- Create: `cmd/dubplate-server/proxy.go`, `cmd/dubplate-server/proxy_test.go`
- Modify: `internal/serve/clientaddr.go`, `internal/serve/clientaddr_test.go`, `internal/serve/strip.go`
  (the proxy set's type only), `internal/serve/strip_test.go` as the type change requires,
  `cmd/dubplate-server/serve.go` (the shell's construction), `cmd/dubplate-server/main.go` (one
  registration line), `cmd/dubplate-server/serve_test.go`

**Outcomes:**
- The `trusted proxy` integration class per R-5d1, registered from `newProductionRegistry`.
- A resolver in `proxy.go` that reads the class from a `*config.Config` (`IntegrationEnabled`,
  `IntegrationValue`, as `backup.go:61-62` does) and returns a `serve.TrustedProxyConfig`, applying
  R-5d2 and R-5d3 and joining every refusal. Serve startup calls it before any listener opens; Task 2's
  doctor row calls the same function.
- `TrustedProxyConfig` keyed by `netip.Addr`, `PeerAddress` compared after parse and unmap, and
  `ClientIP` applying R-5d4. The exported surface does not grow except as the new resolver needs;
  `NewTrustedProxyConfig` takes parsed addresses.
- `newServeShell` builds the strip and every handler's `ClientIP` from the resolved configuration,
  keeps `serve.FrontExemption{}`, and its doc comment's "No trusted proxy is configured yet" paragraph
  (`serve.go:239-242`) is replaced by the real contract.

**Acceptance criteria:**
- A table test over the resolver covers each R-5d2 and R-5d3 state by name: flag absent (mode direct),
  flag `false` (direct), flag `true` with each Required key absent and empty (the `Load` report), each
  refused address shape (hostname, `192.0.2.0/24`, `fe80::1%eth0`, `0.0.0.0`, `::`, an empty entry
  from `a,,b`), each refused header (`Cookie`, `Forwarded`, `Remote-User`, `Remote_User`, `X Bad`), a
  two-bad-entries value naming both, and a valid `192.0.2.10, ::ffff:198.51.100.7` with `Cf-Connecting-Ip`
  yielding a set whose second member reads `198.51.100.7`.
- `ClientIP` rows, each with its fixture peer and header stated first: a trusted peer with
  `203.0.113.9, 192.0.2.10` (proxy rightmost) yields `203.0.113.9`; a trusted peer with
  `203.0.113.9, garbage` yields the peer; a trusted peer with `garbage, 203.0.113.9` yields
  `203.0.113.9`; an untrusted peer with any header yields the peer; a trusted peer with an
  IPv4-mapped hop yields dotted form; two header lines read as one chain (`clientaddr.go:96-101`).
- A shell built by `newServeShell` from an environment enabling the class, driven through
  `httptest` with `RemoteAddr` set to the configured proxy, keys the link-request limiter on the
  header's client (the per-client limit trips for that client and not for a second client behind the
  same proxy); the same request from an unconfigured peer keys on the peer. **Fixture state:** a
  fixture roster and a stub mailer, the limiter's per-client limit set to 1.
- The shell's strip carries the zero front exemption: a request from the configured proxy bearing
  `Remote-User` reaches a registered handler without it.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Continue the scan past a malformed hop → the `garbage, 203.0.113.9`-shaped row inverted
   (`203.0.113.9, garbage`) returns the leftmost value and fails.
2. Compare the peer as a string, not a `netip.Addr` → the `::ffff:` proxy row fails.
3. Drop the CIDR refusal → the `192.0.2.0/24` row fails.
4. Build the shell with the zero `TrustedProxyConfig` again → the limiter row fails.

**0.8 hook:** lands hook 19's client-IP half (a direct mode that trusts no forwarded header, chosen
explicitly), since this class is its first consumer; the listener half stays rung 12's. Hook 18
(description and sensitive mark on every key) is neutral: `IntegrationClass` carries neither, and the
field is rung 14's to add. **Deliverable count: 4** (the class, the resolver, the netip set with the
hop rule, the shell wiring).

---

## Task 2: Doctor's client-IP, web, and bind-address rows

Model: `sonnet`. Depends on Task 1.

**Files:**
- Create: `cmd/dubplate-server/doctorprobe.go` (the three offline rows' status functions and Task 3's
  row), `cmd/dubplate-server/doctorprobe_test.go`
- Modify: `cmd/dubplate-server/doctor.go` (the four rows' print calls after the backup rows and the
  exit condition), `cmd/dubplate-server/doctor_test.go` (`probeRowLines`, R-5d5)

**Outcomes:**
- `client ip`: `direct` or `trusted proxy: <header> from <n> address(es)`, or `invalid: <joined
  refusal>` (failing), from Task 1's resolver. It names no address, since a proxy address is
  deployment topology a terminal need not show; the count is enough to see the mode.
- `web` per R-5d6.
- `bind addresses` per R-5d7.
- `boundary probe` prints `not probed` in this task (Task 3 replaces it), so the line count is final
  here.
- `newDoctorCmd`'s exit condition (`doctor.go:129`) gains the three rows' failures; its doc comment's
  exit-status sentence lists them.

**Acceptance criteria:**
- Each row's states, each by its own test with its fixture stated first: `client ip` direct, trusted,
  and invalid (a CIDR entry) with exit 1; `web` ok, an unknown theme match value with exit 1 naming the
  accepted set, a missing operator CSS file with exit 1, and `not checked` when `DUBPLATE_SITE_NAME` is
  unset (exit 1 from `Load`, and the row's own text proves it added no second reason); `bind addresses`
  with `127.0.0.1` present, `127.0.0.2` present (loopback), `192.0.2.1` failing, and `127.0.0.1,,::1`
  invalid.
- `TestDoctorCmd` and `TestDoctorCmd_LineCountFollowsRegistry` pass with `probeRowLines = 4`
  hard-coded; `TestDoctorCmd_AbsentStoreCreatesNothing` still passes (the web row writes nothing).
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Print the web row without calling `newWebRoutes` (the old unjudged value) → the unknown-theme row
   exits 0 and fails.
2. Treat a loopback address as absent → the `127.0.0.2` row fails.
3. Derive `probeRowLines` from the rows' own slice → replace with the literal and show the derivation
   mutation (one row dropped) passes the derived count and fails the literal one.

**0.8 hook: neutral.** The 0.8 slice 6 admin credential list names "a `doctor` row [that] names the gap
in the default mode"; `client ip` is that row's present form, and 0.8 extends its text, never its
mechanism. **Deliverable count: 4.**

---

## Task 3: The boundary probe route and its doctor row

Model: `opus` (a live-dial proof that must be able to fail; the spec rejected an in-process probe for
exactly that reason).

**Files:**
- Create: `internal/serve/probe.go` (the route handler, the probe client, and the header set they
  share), `internal/serve/probe_test.go`
- Modify: `cmd/dubplate-server/serve.go` (mount the route), `cmd/dubplate-server/doctorprobe.go`,
  `cmd/dubplate-server/doctorprobe_test.go`

**Outcomes:**
- The route and client per R-5d8, the row's states per R-5d9. The probe request carries every group 1
  spelling in both forms (sent with the exact map keys, since `http.Header.Set` would canonicalize an
  underscore key), every group 2 name, and the sentinel listed in `Connection`.
- `doctor` dials each configured bind address on the configured port.
- The route's local-peer rule reads the same resolved bind list serve listens on, passed at
  construction; the route never reads configuration itself.

**Acceptance criteria:**
- In-package, against the real `Strip` composed by `NewShell`: a probe through the shell reports all
  counts zero; the same probe sent to the route handler **without** the strip reports every count
  nonzero (proving the route can see what the strip removes, so a zero is evidence).
- End to end in `cmd/dubplate-server`: the real shell from `newServeShell` served by `serve.Run` on
  `127.0.0.1` at a port the test reserved, `doctor` run against an environment naming that port, and
  the row reads `refused`. **Fixture state:** the listener answering, asserted by one plain `GET` to
  the route before doctor runs.
- With nothing on the reserved port, the row reads `no listener` and doctor's exit status is unchanged
  by the row.
- With a test server on the port whose handler answers the route's JSON with a nonzero identity count
  (a listener that did not strip), the row fails and doctor exits 1.
- A request to the route from a peer outside the local set gets the terminator's 404, byte-equal to an
  unknown path's; `POST` gets the same 404.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Mount the route outside the strip (a root-level handler) → the end-to-end row reads a nonzero count
   and fails.
2. Report `no listener` as a failure → the no-listener row's exit status changes and fails.
3. Accept any peer at the route → the outside-peer 404 row fails.
4. Send headers through `Header.Set` → the underscore spellings arrive canonicalized; the in-package
   no-strip row's identity count falls below the list's length and fails.

**0.8 hook: neutral.** **Deliverable count: 4** (the route, the client, the row, the end-to-end proof).

---

## Segment B

## Task 4V (under Ruling 2's no only): The forward-auth verify endpoint

Model: `opus`. Under Ruling 2's yes this task does not exist.

**Files:**
- Create: `cmd/dubplate-server/forwardauth.go` and its test; `internal/auth/verify.go` and its test
- Modify: `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/main.go`

**Outcomes:**
- The `forward-auth verify` integration class: flag `DUBPLATE_FORWARD_AUTH_VERIFY_ENABLED`, Required
  `DUBPLATE_FORWARD_AUTH_VERIFY_ADDRS` under R-5d2's grammar (R21).
- A `verify` route under the owned prefix, mounted only when the class is enabled: to a permitted peer
  (compared as `netip.Addr`) it calls the seam's `VerifyIdentity` with `sameOrigin` false (no refresh,
  no `Set-Cookie`, no state change) and answers 200 with `Remote-User` carrying the resolved principal
  and `Cache-Control: no-store`, or 401 with no body; to any other peer, the terminator's 404. It never
  reads the body. The task quotes the forward-auth documentation of Caddy (`forward_auth`), Traefik
  (`forwardAuth`), nginx (`auth_request`), and Authelia, each read on the day, for which method each
  sends, and answers every method the shell admits alike; it quotes `internal/serve/origin.go` for
  whether a non-safe method from the proxy meets the origin check and registers the route's exemption
  if it does, with the reason (no state change, peer-gated).
- Task 7's runbook gains the forward-auth setup steps BS:504-514 requires, the proxy's strip set
  published from group 1's list.

**Acceptance criteria:** a permitted peer with a valid session gets 200 and the principal; the same
peer with no cookie, an expired session, a session for a removed member, and a session older than the
epoch each get 401; a forged `Remote-User` from the permitted peer with no cookie gets 401 (the
forged value never reaches the answer); a non-permitted peer with a valid session gets the 404; the
class disabled mounts no route (404); no answer carries `Set-Cookie`. `bash scripts/check.sh` green.

**Named mutations:** pass `sameOrigin` true (a `Set-Cookie` appears); compare the peer through
`ClientIP` instead of `PeerAddress` (a non-permitted peer behind a configured trusted proxy naming a
permitted address in the header gets 200); mount the route when disabled.

**0.8 hook:** lands hook 5's read side (the principal is `streamer_username`, required on write since
5c-i). **Deliverable count: 4.**

---

## Task 5: The serve unit, the pre-creates, the clean web build, and the state-dir check

Model: `sonnet`. Depends on Task 1 (the binary the unit runs must carry the class).

**Files:**
- Create: `systemd/dubplate-server-serve.service`
- Modify: `scripts/deploy.sh`, `scripts/check.sh`

**Outcomes:**
- The unit per R-5d10, with a comment per inherited block stating its reason, on the shadow unit's
  precedent, and a derivation comment above `ReadWritePaths=` naming the two leaves' sources
  (`auth.StoreDir`, `internal/auth/store.go:36-39`, and the template's roster row's directory).
- `deploy.sh`: the unit joins `UNIT_FILES` (rsynced and daemon-reloaded, never enabled or started, and
  `TIMER_UNITS` unchanged); `DUBPLATE_SERVER_LEAVES` and its pre-create loop per R-5d11, with the
  comment naming every unit that depends on each leaf; the clean web build per R-5d13.
- `scripts/check.sh`: the generalized state-dir leg per R-5d11 and the hard-required
  `systemd-analyze` per R-5d12.
- **Evidence, recorded in the task report, not a gate leg:** on the workstation, a user-level
  transient unit (`systemd-run --user --wait --collect`) carrying the new unit's hardening properties,
  with `ReadWritePaths=` remapped to two scratch leaves, runs `dubplate-server member add --admin`
  against a scratch roster path and then `dubplate-server serve` for a few seconds against a fixture
  environment; the roster file appears and the auth leaf gains its subdirectories. The same run with
  the roster leaf absent fails with systemd's own text, measured on this workstation 2026-09-25 with
  systemd 259: "Failed to set up mount namespacing: <path>: No such file or directory" and "status=226/NAMESPACE"
  (`man systemd.exec`'s exit-code table: "226 EXIT_NAMESPACE Failed to set up mount, UTS, or IPC
  namespacing"). The box repeats the positive half in Task 10.

**Acceptance criteria:**
- `systemd-analyze verify` over `systemd/*` clean modulo the expected box-only `ExecStart` paths
  (the leg's existing filter, `check.sh:634`).
- The state-dir leg passes the real tree and fails by name on each of: `roster` removed from
  `DUBPLATE_SERVER_LEAVES`; `auth` removed; a `-/var/lib/dubplate/roster` entry; an entry outside the
  state directory (`/var/lib/other/x`); a nested entry (`/var/lib/dubplate/auth/nonces`); every unit's
  `ExecStart` pointed elsewhere (no greenfield unit found).
- `shellcheck` clean on both scripts.
- `bash scripts/deploy.sh --check` (read-only; the box is reached for diffs only) reports the unit's
  install, the two new pre-creates, and the web build, and refuses by name when a scratch edit makes
  `web/` dirty. It still aborts at the render on the names Task 6 has not yet added, which the report
  records as expected.
- The evidence run's two outcomes are in the report with their exit codes and systemd's text.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:** the six state-dir seeds above, each failing the gate by name; restore the
`command -v systemd-analyze` guard with the tool hidden from `PATH` → the gate fails naming the tool.

**0.8 hook: neutral.** 6d adds the `store` leaf by one array word (rung 6 spec criterion 14) and
retires `roster` (erratum 8). **Deliverable count: 4** (the unit, `deploy.sh`, the state-dir leg, the
namespace evidence).

---

## Task 6: The environment template and the rendered-doctor proof

Model: `sonnet`. Depends on Tasks 1, 2, 3 (doctor's final row set) and Task 5.

**Files:**
- Modify: `env/dubplate-server.env.template`, `cmd/dubplate-server/templatekeys_test.go`

**Outcomes:**
- The rows per R-5d14, each with a comment giving its row state and reason, and the header's
  secret paragraph rewritten.
- A gate-resident test that parses the greenfield template the way `deploy.sh`'s `render_template`
  does (a literal keeps its value, a bare row takes a fixture value, a commented row is omitted, a line
  with no `=` fails the test), adds a fixture state directory, a fixture roster naming one admin at
  mode 0600, and a fixture bundle state as the gate already builds it, runs `doctor` against that
  environment, and asserts exit 0 with every row's status stated. Fixture values are obviously not
  credentials (`example.test` addresses, a 32-byte repeated-letter secret) and trip no gitleaks rule.

**Acceptance criteria:**
- The rendered-doctor test exits 0 with the fixture roster present, and exits 1 naming the `roster`
  row when the roster file is absent (the runbook's "doctor before the bootstrap" state, stated in the
  test's name).
- Deleting the `DUBPLATE_ROSTER_FILE` row fails the test naming the magic-link source's missing key;
  deleting the trusted-proxy flag row leaves it passing (the class is off by default), which the test
  asserts deliberately so a future reader does not mistake the explicit `false` for a requirement.
- Under Ruling 3's no: the two secrets are commented, and the test instead asserts exit 1 naming
  exactly those two keys.
- `TestTemplateKeysClassifiedAgainstProductionRegistry` and `TestTemplateKeysCoverEveryOwnedNowRow`
  pass unchanged.
- `bash scripts/deploy.sh --check` renders with no unresolved name **except the two secrets under
  Ruling 3's yes** (they are minted at Task 10), and the report quotes the abort's name list to prove it.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:** make the template parser keep a commented row (the `_PREVIOUS` secrets' fixture
values join the environment; a short fixture there makes doctor fail and proves the omission is
tested); drop the fixture roster's admin flag (the roster row fails, exit 1).

**0.8 hook: neutral.** **Deliverable count: 3** (the rows, the header, the rendered-doctor proof).

---

## Segment C

## Task 7: The server runbook and the trust-boundary documentation

Model: `sonnet`, Google developer documentation standard (`writing-voice`). Independent of Tasks 1 to 6
by files; it describes their behavior, so it runs after them.

**Files:**
- Create: `docs/server-runbook.md` (R-5d15)

**Outcomes, one section per read, each opening with a one-paragraph brief:**
- **Install order** (spec `:1771-1774`): install, render the environment, run `dubplate-server member add
  --admin` as the service user through `systemd-run` carrying the unit's properties (the shadow unit's
  operator rule, `dubplate-shadow.service:60-64`), run `doctor`, start the unit.
- **The doctor rows**: what each new row proves and cannot, `no listener` before start, and the
  boundary probe's limit.
- **Mail**: two worked SMTP examples, one with no Cloudflare account (a mailbox provider's submission
  port with STARTTLS) and one through Cloudflare Email Service's SMTP submission endpoint (R7), and the
  outbound-465 note; the mail-ceiling lockout lever and the kind-mode per-client bound (5c-i's carry,
  `HISTORY:1729-1731`).
- **The trust boundary**: the peer address and the reported client IP (R27), trusted-proxy
  configuration with the Authelia and Miniflux corroboration the spec names (`:894-901`), the HTTPS
  precondition (R38), and the forward-auth topology stated as a future integration under Ruling 2's
  yes (its requirements from BS:504-514: the proxy strips the published set before the verify
  subrequest and before the streamer, and the streamer trusts only that proxy), or the setup steps
  under its no.
- **Browser and presentation notes**: Ruling 6's old-browser `Referrer-Policy` residual; a branding
  color applies in both color schemes (identity spec amendment A4).
- **Backup rows** carried from 6a and 6b-i: `KillMode=mixed` for a unit that runs rclone, rclone
  v1.74.3 or later, a crypt remote, and the exit-7 remedy (`HISTORY:884-887`).

**Acceptance criteria:** every section exists with its brief; every command in it was run by the
implementer against the workstation or a fixture and matches the binary's `--help` (the report lists
each); no estate hostname or personal address appears; one `prose-voice-reviewer` read (the
conductor's dispatch) and its fold; `bash scripts/check.sh` green (the gate reads no Markdown, so this
proves the commit is clean).

**0.8 hook: neutral.** **Deliverable count: 3** (the runbook, the trust-boundary section, the prose
fold).

---

## Task 8: The SEC and UNI disposition table

Model: `opus`. Independent of every code task by files; it may run as a document lane beside Segment A
if the registry allows, and must land before the adversarial cycle.

**Files:**
- Create: `docs/superpowers/research/2026-09-25-rung-5-sec-uni-disposition.md`

**Outcomes:** one row per finding, all 50 (SEC's 6 blockers, 13 majors, and 12 minors; UNI's 3
criticals, 6 highs, 7 mediums, and 3 lows, per the spec's count at `:2003-2006`): the finding's own id
and a one-line statement; the birth-requirement clause or ruling that absorbed it (a spec pointer); the
code and the test that prove it at HEAD (`file:line` and test name, grepped, never recalled); and a
status of absorbed, absorbed with a recorded residual (the residual stated), or carried (to which rung,
citing the carry). A closing section lists the residue in one place.

**Acceptance criteria:** the row count is 50 and each review text's own finding count matches its
rows; every cited test exists (`go test -list` output in the report); every residual cites where it is
recorded; no finding is marked absorbed without a test or an explicit "by design" pointer.

**0.8 hook: neutral.** **Deliverable count: 3** (the table, the residue list, the verification).

---

## The adversarial cycle (conductor's actions, between Tasks 8 and 9)

1. One `code-simplifier` round over Tasks 1 to 6's Go and shell, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, reviewed by one `diff-reviewer`.
2. Two fresh-context lenses, dispatched in one message, neither given this plan:
   - **Security**: `web-auth-security-reviewer` (`claude-opus-5-5`, effort `high`) over the whole rung 5
     subsystem at HEAD: `internal/serve`, `internal/auth`, `internal/identity`, `internal/notify`,
     `cmd/dubplate-server`'s serve, auth, member, doctor, and proxy files, the new unit, `deploy.sh`'s
     new steps, the template, and `web/`'s three pages.
   - **Universality**: one `opus` reader (effort `high`), the lens that caught UNI's three criticals,
     over the same tree plus the runbook, asking whether a stranger with no Cloudflare account, a
     different mail provider, a different proxy, or IPv6 can install and operate it from the runbook
     alone, and whether any estate value leaks outside `internal/config`, the template, the unit, and
     `deploy.sh`.
   Each receives the spec's threat model (`:1950-2009`), Task 8's table, the six review-focus points,
   and the list of prior review records as the do-not-re-litigate baseline (a finding on a "verified
   sound" item needs new evidence). Each returns correctness gaps ranked by consequence, with
   `file:line` and a reproduction where one is possible.
3. Both reports are committed as
   `docs/superpowers/research/2026-09-25-rung-5d-review-{security,universality}.md`.

## Task 9: The adversarial fold

Model: `opus` (the fold rules on security findings; CLAUDE.md lets a fold refuse a finding whose fix
costs more than the risk it removes).

**Files:** whatever each accepted finding's fix touches, inside this pass's territory wall; the merged
disposition record `docs/superpowers/research/2026-09-25-rung-5d-review-fold.md`.

**Outcomes:** every finding from both lenses gets exactly one disposition, folded (with the commit and
the test that fails without it) or refused (with a one-line reason), in the merged record, which
becomes the rung's final do-not-re-litigate list. A finding outside the territory wall is carried to
the rung that owns the surface, named in the record. A fix that would touch `internal/config` is
carried, never made (the 6c lane condition).

**Acceptance criteria:** the record covers every finding in both reports (the counts match); each fold
has a test that fails with the fix reverted; `bash scripts/check.sh` green, heavy lane. **If the
lenses return more than about six findings needing code**, the conductor splits the fold into two
tasks at a file-disjoint seam rather than growing one task, and says so at the checkpoint.

**0.8 hook: neutral.** **Deliverable count: up to 4.**

---

## Task 10 (under Ruling 3's yes): The first-deploy sitting

**Attended; conductor-run with Geoff present. No implementer.** Every box write happens here and
nowhere else. Each step's output is recorded in the close.

1. **Preconditions on the workstation:** master green at Task 9's commit; the lane registry shows no
   other executor in lane 1 and lane 2 not mid-merge; AC power.
2. **The quiet-box check** (read-only): no import, backup, or shadow run in progress (`systemctl
   list-units --state=activating 'dubplate-*'`), no upload traffic, and the shadow timer's next firing
   more than ten minutes away. If the box is not quiet, stop and reschedule.
3. **The mint** (one 1Password approval, Geoff's touch): `DUBPLATE_MAGIC_LINK_SECRET` and
   `DUBPLATE_SESSION_SECRET` through `~/.dotfiles/scripts/secrets/secret-set.sh`, each a fresh random
   value of at least the source's length floor (`secretLengthFloor`, `auth.go`) generated inside the
   command and never echoed; each documented in `~/.dotfiles/secrets/registry.md` with its scope (the
   greenfield server's HMAC keys, box-only, not a Worker secret) and its rotation note (the `_PREVIOUS`
   overlap doubles the count at rotation, spec `:1662-1664`); `sync.sh --verify`. No value is printed.
4. **The local doctor, pre-deploy:** the template rendered locally against `~/.local/secrets` in a
   subshell with a scratch state directory and a scratch roster, `dubplate-server doctor` run against
   it, every row read. This proves the estate's real SMTP and URL values pass the magic-link
   source's validation before anything reaches the box. Values are never printed.
5. **`bash scripts/deploy.sh --check`**: renders with no unresolved name; the report is read in full,
   the unit is installed and not enabled, the two pre-creates appear, the binary is stamped, and
   nothing touches the incumbent binary or its units. Geoff approves the apply.
6. **`bash scripts/deploy.sh`** (apply).
7. **Box verification:** the two leaves exist as `music:music 0700`, with their SELinux context read
   once (`ls -Zd`; `var_lib_t` expected, vps-conventions); `systemctl is-enabled
   dubplate-server-serve.service` answers `disabled`; the shadow timer still active and its next run
   succeeds (the journal read after its next firing, or a `systemctl start dubplate-shadow.service` run
   Geoff approves).
8. **The admin bootstrap:** `dubplate-server member add --admin <DUBPLATE_ADMIN_EMAIL> --name
   <Geoff's display name> --streamer-username <Geoff's Navidrome login>`, run as `music` through
   `systemd-run --wait --pipe --uid=music --gid=music -p EnvironmentFile=/etc/dubplate/server.env -p
   ProtectSystem=strict -p ReadWritePaths=/var/lib/dubplate/roster`, then `member list` the same way.
9. **Doctor against the rendered environment**, the same way: exit 0, every row read; `boundary probe`
   reads `no listener`.
10. **One supervised start:** `systemctl start dubplate-server-serve.service`; `doctor` again, with
    `boundary probe` reading `refused` and `bind addresses` reading `127.0.0.1 present`; the journal
    shows the listener and no error; `systemctl stop dubplate-server-serve.service`. The unit stays
    disabled; nothing routes to it (the tunnel still serves Navidrome only).
11. **Lift the carry:** the conductor edits STATUS's "No deploy until 5d" item to record what deployed
    and what the next deploy may carry.

If any step fails, the sitting stops at that step, the box is left as the last successful step left
it (every step before the apply is read-only, and the apply is `deploy.sh`'s idempotent set), and the
failure becomes a finding for a follow-up task, never an improvised fix on the box.

---

## Task 11: The rung 5 close

Model: `sonnet`.

**Files:** the close drafts under `docs/superpowers/research/handoffs/`; the errata targets named below.

**Outcomes:**
- **The 5d HISTORY entry and the rung 5 summary**: what landed per task, the gate evidence, the
  adversarial cycle's counts and dispositions, the sitting's record (under Ruling 3's yes), declared
  coverage gaps, both budgets scored, and "What a later pass would be wrong to rediscover" (at least:
  the hop rule stops at a malformed hop; doctor's web row is serve's own resolution; the state-dir leg
  reads `DUBPLATE_SERVER_LEAVES`; the unit is installed disabled; the estate's trusted-proxy values
  wait for the cloudflared address pin).
- **The STATUS draft**: present tense, rung 5 closed, the lane registry line, the next action, the
  carries 5d hands on: to the cutover (rung 16): the trusted-proxy estate values and the `compose.yaml`
  address pin for cloudflared, the serve bind to the bridge gateway, and enabling the unit; to rung 12
  under Ruling 2's yes: the verify endpoint; to 6c/6d: moving 5d's value checks onto the registry's
  validator.
- **The errata**, each dated and pasted where the spec's "Errata the close applies" section says,
  skipping any already applied (grep first): the build spec's ladder entry 5 (page stack), ladder
  entry 7 (gate rejoin and roster), B12 (proxy hygiene), B4 (tokens and logged URLs), and every other
  paste-ready block in that section. Plus the rung 5 spec's own 5d-close errata: the 5d row as
  executed (lane 1, not a worktree; the example-roster clause already dropped); the SMTP rows bare, not
  literal (R-5d14); the secrets minted at 5d (Ruling 3's yes) or carried (no); the verify endpoint
  moved (Ruling 2's yes); R23's reservation moved to about 46.5M; and the pass-split table's 5d exit
  gate's "the unit starting ... failing with a named mount-namespacing error" recorded as Task 5's
  workstation evidence plus Task 10's box start.
- **The contested-surface rows** released: `internal/serve` and `internal/auth` pass to rung 12's
  ownership per the spec's table (`:2545-2546`).
- **Two `go-architecture-reader` dispatches** (conductor's): `internal/serve` and `cmd/dubplate-server`,
  no plan in context; findings routed to the architecture cleanup pass.

**Acceptance criteria:** every carried item in STATUS's "No deploy until 5d" line is closed or routed
by name; `git diff --stat <5d base>..HEAD -- internal/config` is empty; `bash scripts/check.sh` green.

**Deliverable count: 4** (HISTORY, STATUS, errata, contested surfaces).

---

## The exit gate

`bash scripts/check.sh` green on master, plus each assertion and its proving task. The spec's 5d row
(`:2519`) is the source; each clause maps here.

1. `systemd-analyze verify` clean on the new unit, and the tool hard-required (Task 5).
2. The generalized state-dir check failing when either leaf is removed from `deploy.sh` (Task 5).
3. The unit starting with the roster leaf writable and failing with systemd's named mount-namespacing
   error when its pre-create is removed (Task 5's evidence; Task 10's box start under Ruling 3's yes).
4. `doctor` reporting the live client-IP mode, the bound addresses, and a boundary-probe row that
   dials the running listener and reports plainly when none answers (Tasks 2 and 3).
5. `deploy.sh --check` rendering with no unresolved name and `doctor` exiting 0 against the rendered
   environment (Task 6's gate-resident proof; Task 10's steps 5 and 9 under Ruling 3's yes).
6. The disposition table covering all 50 findings (Task 8).
7. The adversarial cycle's findings folded or refused with reasons (Task 9).
8. `internal/config` untouched (Task 11's check).

**Deferred, by name:** the verify endpoint (Ruling 2's yes); the estate's trusted-proxy values and the
cloudflared address pin (rung 16); minor 14 (6d criterion 8); L5 and L6 (unassigned).

---

## Contested surfaces

| Surface | Owner | This pass's position | Contention |
| --- | --- | --- | --- |
| Lane 1 (the main tree) | the lane registry | the whole pass | 5c-iv must close first; the conductor writes the order |
| `cmd/dubplate-server/doctor.go`, `doctor_test.go`, `main.go` | shared | Tasks 1, 2, 3 | lane 2's 6b-ii-b and translit (doctor rows, a verb, registration lines) and 6c (doctor's settings rendering); a non-fast-forward merge triggers the merge ritual's re-review. Task 0 reports what merged |
| `scripts/deploy.sh` | shared | Task 5 | translit's staged transition (lane 2); 6d later edits the leaf array |
| `scripts/check.sh` | shared | Task 5 | every lane; the state-dir leg is 5d's |
| `env/dubplate-server.env.template` | rung 4's file; rung 5 adds rows | Task 6 | 6d removes the roster row |
| `systemd/` | shared | Task 5 (one new file) | 6d edits this unit's grant |
| `internal/serve` | rung 5 until this close, then rung 12 | Tasks 1, 3 | none open (PA:80) |
| `internal/config` | rung 1's contract | untouched | 6c may run in lane 2 concurrently |
| The box | production | Task 10 only | the quiet-box check; no other lane deploys during 5d |
| `docs/` ledgers | conductor | Task 11 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record planning spend against the rung; write the lane registry's lane 1 row and the
   contested-surface rows; arm the runaway guard; AC power; confirm the heavy gate string.
2. Before Task 1: rule Task 0's blockers.
3. At each segment boundary: STATUS (task ledger, decisions, spend, next task); the 80% flag check.
4. After Task 7: the `prose-voice-reviewer` read and its fold inside Task 7.
5. Between Tasks 8 and 9: the adversarial cycle as written above.
6. Task 10 under Ruling 3's yes, attended.
7. At Task 11: the two architecture reads; score both budgets; close the session.

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, one invocation per segment. Every
implementer task runs `dubplate-implementer`, then `diff-reviewer` (`claude-opus-5-5`, effort `high`),
then the gate, with one re-dispatch on `fix` and a second `fix` as the conductor's decision.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 0, 1, 2, 3 | STATUS. Task 8 may start beside A as a document lane if the registry allows. |
| **B** | (4V under Ruling 2's no), 5, 6 | STATUS. Under Ruling 2's no this boundary closes 5d-i. |
| **C** | 7, 8, the adversarial cycle, 9 | STATUS; the simplifier; the two lenses; the fold. |
| **Sitting** | 10 (Ruling 3's yes) | attended |
| **Close** | 11 | the architecture reads; the ledgers |

Each boundary sits on a task's own green committed gate.

**Models.** `sonnet` at effort `high` by default. **Upshifts to `opus`, unconditional:** Task 0 (fresh
context), Task 1, Task 3, Task 4V, Task 8, Task 9, and the universality lens.

**Dependencies, stated exactly:** Task 0 before any task. Task 1 → Task 2 → Task 3. Task 1 → Task 5.
Tasks 2, 3, 5 → Task 6. Tasks 1 to 6 → Task 7. Task 8 depends on nothing but must precede the cycle.
The simplifier and the lenses → Task 9 → Task 10 → Task 11. **Tasks 5 and 2 are independent by files**
but run sequentially because one tree runs one executor.

**Where this pass splits, if it must:** at the Task 6 boundary (5d-i: Tasks 0 to 6; 5d-ii: Tasks 7 to
11), under Ruling 2's no by rule, or under its yes if segment B's re-derivation projects past the flag.

---

## Prior art consulted

Per CLAUDE.md, from the streamer spec's list (`:770-806`) and the rung 5 prior-art study
(`docs/superpowers/research/2026-09-08-rung-5-auth-prior-art.md`).

| Entry | Verdict for 5d, with reason |
| --- | --- |
| stdlib `net/netip` | **Adopt** for the proxy set and the hop parse: value-type addresses with `Unmap` give the canonical comparison `net.IP` does not (R-5d2, SEC finding 22). |
| stdlib `net.Dialer` with `LocalAddr`, `net.InterfaceAddrs` | **Adopt** for the probe's known source and the bind row (R-5d7, R-5d8). |
| `golang.org/x/net/http/httpguts` | **Decline**: a new module dependency against R26's count of two, for a token check of a few lines. |
| `golang.org/x/time/rate` | Already adopted at 5b (R26); 5d changes only the key it receives. |
| Authelia (Apache-2.0) forwarded-headers documentation | **Adopt as behavior and citation** for the runbook's trust-boundary section (spec `:894-899`). No code read. |
| Miniflux `AUTH_PROXY_HEADER` with `TRUSTED_REVERSE_PROXY_NETWORKS` | **Adopt as shape corroboration** for R22's class; **decline** its network (CIDR) form, since the spec pins trust to addresses, never a subnet. |
| `coreos/go-systemd` (sd_notify) | **Decline**: `Type=exec` reports an exec failure without a readiness protocol, and adding a dependency for readiness no consumer needs is ahead of a caller. |
| Navidrome's reverse-proxy authentication (GPL-3.0) | **Behavior only, and only under Ruling 2's no**, for what a streamer does with `Remote-User`; never read for code. |
| Caddy `forward_auth`, Traefik `forwardAuth`, nginx `auth_request` documentation | **Consulted under Ruling 2's no** (Task 4V quotes each for the method it sends); not consulted under yes. |
| OpenSubsonic, taglib, wazero, modernc sqlite, goose, mewkiz/flac, Picard, chromaprint | **Neutral**: none touches 5d's surface. |

---

## Not verified in this revision

- Whether translit's `unidecode/` leaf needs a pre-create in `DUBPLATE_SERVER_LEAVES` or is written by
  its install verb as root; Task 0 item 4 reads translit's landed `deploy.sh`.
- The exact text `systemd-run --user` prints on the box's systemd 257 for the namespace failure (the
  workstation's 259 was measured); the box repeats only the positive half, so the exact wording on 257
  is never asserted.
- Whether `validateSMTPFrom` accepts the age store's `DUBPLATE_SMTP_FROM` value; Task 10 step 4 proves
  it before the deploy, and a refusal there is a finding, not a deploy.
- The heavy gate's wall time after 5c-iv, which sets each task's `-timeout`.
