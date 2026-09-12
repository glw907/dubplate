# Rung 5, Pass 5b-ii (The Roster, the Limiters, the Mail Seam, and the Auth Handlers) Implementation Plan

Status: **revision 2.1, 2026-09-12.** Revision 2.1 applies the process
improvement plan's tier C to this plan
(`docs/superpowers/research/2026-09-12-process-improvement-plan.md`): the
four per-segment `code-simplifier` rounds become two, one at the segment C
boundary and one at D2, each round's review chunked one `diff-reviewer`
per package (Item 8's half-step), and the segment table records that its
five segments satisfy Item 16's ruled three-to-four-task sizing on green
commits. Nothing else changed. Drafted just in time from pass
5b-i's own handoff (`docs/HISTORY.md:496-530`), against the rung 5
identity-and-auth spec at revision 2.1
(`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`), whose
eighteen 5b-i errata landed at `b91bae1` and whose six product forks
Geoff ruled on 2026-09-08 (`spec:172-271`).

**Revision 2 folds three adversarial plan reviews** (contract and code
fidelity, security and threat model, execution mechanics and budget),
merged at
`docs/superpowers/research/2026-09-12-rung-5b-ii-plan-review.md`. Two
blockers, thirteen majors and eight minors were raised against revision
1; every one is folded. The task count is unchanged at twelve plus Task
0. The ceiling moved from 5.5M to **7.5M** because revision 1's
derivation rested on two figures no report carries.

**Seven folds are conductor-owed**, each marked at its own site as
`**Conductor-owed fold:**` with the question and the applied answer, so
the conductor ratifies or reverses in one edit:

1. **The ceiling.** Rule 7.5M with the flag at 6.0M, or take the Task 8 /
   Task 9 cut now at drafting time? Applied as the 7.5M ceiling, one
   pass. (Token ceiling section.)
2. **R3's `doctor` roster collision row.** Does it land in Task 4,
   widening that task's Files into `cmd/dubplate-server/doctor.go` and
   taking it to five deliverables against the plan's own four-deliverable
   cap, or route forward as a declared gap? Applied as Task 4's fifth
   deliverable. (Task 4.)
3. **R9 against fork 4's `generic` mode** (R-5b28). Does "a send failure
   is a failed login" govern a `generic` deployment, or does `generic`
   answer before the send? Applied as the narrow reading: R9 governs
   `kind`, and `generic` answers before the send. (R-5b28, Task 10.)
4. **The admin-flag invariant's enforcement point.** The store's write
   path, or 5c's caller with a gap record here? Applied as enforcement in
   the write path. (Task 5.)
5. **Task 8's model.** `opus` unconditionally, its own segment boundary,
   or plainly `sonnet`? Applied as `opus` unconditionally. (Task 8,
   Execution model.)
6. **M5's failing-first test.** May M5 ship without one if the allocation
   benchmark cannot be made to fail against the unfixed code, recorded as
   a declared gap? Applied as the named fallback the task reports. (Task
   1.)
7. **R10 against Task 2's M4.** Does M4 drop the per-key rows, preserving
   R10's one-line-per-source rule, or does R10 get amended to admit
   per-key rows for the selected source? Applied as dropping the per-key
   rows. (Task 2.)

**This plan is the second half of the split `spec:2163-2172`
pre-authorized and pass 5b-i took.** 5b-i landed the identity leaf, the
config contract, the token, the session, and the two state records, and
merged at `1d4ed29`. 5b-ii lands everything the spec's 5b row names and
5b-i did not: the roster store with its write path, the three rate
limiters, `internal/notify`, the origin check, the verify-identity seam,
and the auth handlers. Its scope statement is `spec:2168`'s 5b row less
the half 5b-i shipped, corrected by the errata at `spec:1982` onward.

**One thing changed between the split and this plan, and it reshapes the
front of the pass.** The fresh-context `web-auth-security-reviewer` read
R-5b16 owes ran on 2026-09-12 against `1d4ed29` and is recorded at
`docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md`. It
returned **one blocker, five majors, and fourteen minors**, every
CONFIRMED finding reproduced by running the landed code rather than by
reading it. Its own verdict is that "5b-ii should not be dispatched until
B1 and the three constraints are folded into its plan". They are folded,
and B1 plus the four reachable majors are **Task 1 of this pass**,
dispatched before any handler exists, because B1's fix changes
`NonceStore.Insert`'s signature and a consumption handler written against
the current two-argument form would be rewritten.

**Band: L**, unchanged. `spec:2382-2440` sets rung 5 at L band across four
passes and the split makes 5b two of them. 5b-i changed `Load`'s
contract; 5b-ii changes no contract of rung 1's, and it is the pass that
takes rung 5's two budgeted module dependencies (R26).

**What this pass is.** The roster store behind one loader interface, with
B9's validation, its admin flag, its atomic write path, its write-time
collision re-check, its audit record, and its single resolution point.
The three rate limiters over one bounded evicting keyed map, with the
global outbound-mail ceiling beside them. `internal/notify` as a general
send seam with one SMTP transport and the header-injection guard. The
origin check, wrapped stricter than the stdlib default, inserted into the
shell's composition after the strip and after normalization. The
verify-identity seam with assertion-channel provenance. The three auth
handlers, request link, consume link, and sign out, as values the shell
mounts. The nonce sweep's interval driver with its one-hour constant. The
six remaining thresholds joining `registerAuthConfig`. And the
correction of every 5b-i security finding that has a home in this
package set.

**What this pass is not.** No page, no `member` verb, no theme, no font,
no `web/` tree, no gate leg: those are 5c's. No systemd unit, no
`deploy.sh` edit, no `doctor` probe row, no verify endpoint, no env
template row beyond what already exists: those are 5d's. **No deploy of
any kind** (the carry below). **A task that believes it needs a page, a
verb, a unit, a deploy, or a probe row stops and reports.**

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section, the pass-execute chain per task, sequential, **dispatched in
> four segments with the conductor boundaries named below**. The 2026-09-12
> window proved pass-execute has no mid-run hook: a segment runs to its
> end or not at all, so every conductor action lands at a segment
> boundary and nowhere else. The implementer is `dubplate-implementer`,
> registered at `.claude/agents/dubplate-implementer.md`. The conductor
> consumes structured reports and never reads a diff, a test log, or a
> gate transcript inline. Subagents never dispatch agents, so
> `code-simplifier` and the pass-end `web-auth-security-reviewer` are both
> the conductor's dispatch (`spec:2617-2619`). The conductor owns the
> ledgers (`docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`), the lane
> registry, the contested-surface rows, and the batch commit of anything a
> task drafts. **This pass runs in LANE 1, the main tree, as the only code
> lane.** Both lanes read CLOSED in `docs/STATUS.md`'s lane registry at
> drafting time, and no second code lane is scheduled against this pass.
> **The live-executor check is that registry, never `pgrep -f
> Projects/dubplate`.**

---

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md`,
   in full.** The read R-5b16 owes for the 5b-i half, run fresh-context
   against `1d4ed29`. Its blocker B1, its majors M1 through M5, and its
   fourteen minors are **Task 1's and Task 2's whole content**, and its
   closing section, "The three findings 5b-ii's plan must carry as
   constraints", is carried verbatim in the Inheritance section below.
   Its "Attacked and found sound" list is **this pass's
   do-not-re-litigate record**: a task that reopens the MAC comparison,
   the domain separation, the verification order, the anchor arithmetic,
   the create-exclusive insert, or the `__Host-` attributes without new
   evidence is spending the pass's budget on a settled question.
   **This file is untracked at drafting time.** The conductor commits it
   before Task 1 dispatches, because this plan cites it as a primary
   input and an untracked input is a citation a later reader cannot
   resolve.
2. **`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`,
   revision 2.1.** The sections 5b-ii implements: "What rung 5 builds"
   (`:93-148`), the ruled product forks (`:172-271`), the rulings
   (`:273-770`), "Consumption" (`:1036-1068`), "The roster"
   (`:1103-1163`), "Rate limits, abuse, and logging" (`:1165-1189`),
   "CSRF" (`:1191-1215`), "The verify-identity seam" (`:1217-1245`),
   "The post-login destination" (`:1247-1267`), "The mail seam"
   (`:1269-1314`), "Packages and their contracts" (`:1465-1496`), "Config
   keys" (`:1498-1656`), "The state directory" (`:1658-1704`), the 2.0
   hook table (`:1841-1859`), the prior-art table (`:1861-1907`), the
   threat model (`:1909-1980`), **the errata in full (`:1982` onward,
   the eighteen 5b-i corrections landed at `b91bae1` included)**, the
   pass split (`:2321-2346`), the contested surfaces (`:2348-2380`), the
   band and ceiling (`:2382-2441`), the staleness pre-task
   (`:2443-2507`), the risks (`:2509-2602`), the process (`:2604-2652`),
   and universality (`:2654` onward). Rulings **R1, R2, R3, R7, R8, R9,
   R15, R16, R25, R27, R28, R30, R31, R33, R37** are the ones 5b-ii
   reads, with **R11** for the roster's write call and **R17** for the
   seam's type. R4, R5, R6, R10, R26, R29, R32, R34, R35, R38 were read
   at 5b-i and are inherited rather than reopened. R12, R13, R14, R19,
   R20, R21, R22, R36 belong to 5c and 5d.
3. **`docs/HISTORY.md`'s "Rung 5, pass 5b-i" entry (`:5-544`)**, and
   inside it, at minimum: "The exported surface" (`:168-184`), "The auth
   core's contract coverage, A1 through A8" (`:195-207`), "The
   resolutions, as ruled" (`:208-235`), "The declared coverage gaps"
   (`:368-402`), "The review record, with its deviations" (`:403-431`),
   and **"The 5b-i to 5b-ii handoff" (`:496-530`), which is binding on
   this plan**. The Inheritance section below states what each clause
   costs a task.
4. **`docs/superpowers/plans/2026-09-08-rung-5b-auth-core.md` (5b-i's
   plan), its 5b-ii cells and clauses.** Grep `5b-ii` returns the scope
   cell, the exit-gate cell, and the provisional 4.5M ceiling in the
   split table, the handoff enumeration in Task 10, and the eleven
   clauses routing a rule forward. **The exit-gate cell is this pass's
   exit gate**, extended by the two assertions the errata moved and by
   the security fold's own.
5. **`docs/superpowers/research/2026-09-08-rung-5b-plan-review.md`**, for
   the folds it routed here rather than to 5b-i: the same-origin refresh
   clause (M-e), the `MaxBytesReader` and 30-second read timeout carry
   (M-g's R-V half), the roster-removal exit-gate assertion (m9), and the
   normalizer's universal-caller rule (M-a). **Its merged "Verified
   sound" list is a do-not-re-litigate record for the plan's shape**, not
   for this pass's code.
6. **`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`
   (PA), its lane table (`PA:96-101`), its scheduling rule (`PA:80`), and
   its contested-surface ownership table (`PA:175-203`).** `PA:100` gives
   lane A `internal/auth`, the `internal/serve` shell, and
   `internal/notify`. The table's rows for `internal/auth`,
   `internal/identity`, `internal/notify`, the roster store, and
   `internal/serve` were all written at the 2026-09-12 pre-flight and are
   the rows this pass edits against.
7. **`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` (S20),
   the forward-hooks section (`:354-383`) and the identity and
   app-passwords section (`:234-252`).** Hook 5 (`S20:367-373`) is the
   hook this pass's roster and seam land; `S20:284-285` puts 2.0's whole
   OpenSubsonic surface under `/rest/*`, which is what the origin check's
   bypass mechanism exists for.
8. **The tree, read-only, before writing anything.** `internal/auth`
   in full (`token.go`, `session.go`, `store.go`, `nonce.go`, `epoch.go`,
   `doc.go`), `internal/identity/identity.go`,
   `cmd/dubplate-server/auth.go` (`registerAuthConfig` at `:108`, the
   four thresholds at `:75-100`, `secretLengthFloor` at `:68`,
   `validateMagicLinkSource` at `:124-156`, `validatePublicURL` at
   `:171-184`, `isLoopbackHost` at `:189`, `resolveAuthConfig` at
   `:207`), `internal/config/registry.go`'s `IdentitySource` (`:349-377`)
   and `Threshold` (`internal/config/thresholds.go:45-51`) with its four
   kinds (`:16-21`), `internal/serve/serve.go` (`Shell` at `:23`,
   `Register` at `:80`, `SetFallthrough` at `:136`, **`Handler`'s
   composition at `:150-163`**, `stripSessionCookie` at `:225`),
   `internal/serve/clientaddr.go`'s `ClientIP` (`:102`),
   `internal/serve/strip.go`'s `NewStrip` (`:300`),
   `internal/serve/normalize.go`'s `Normalize` (`:83`),
   `cmd/dubplate-server/serve.go` (`newServeCmd` at `:122`,
   `newServeShell` at `:158-161`), `cmd/dubplate-server/doctor.go`
   (`rowStatuses` at `:88-112`, `identitySourceStatus` at `:167-177`),
   `internal/atomicfile/atomicfile.go` (`WriteFile` at `:16`,
   `WriteStream` at `:55-91`, the directory fsync at `:82-89`),
   `env/dubplate-server.env.template`, and `scripts/check.sh`.
9. **`_salvage/`, read-only, on merit, never imported.**
   `_salvage/internal/notify/notify.go` (96 lines: `Mailer`, `New`,
   `Send`, `implicitTLSPort` at 465, and two test-only fields each
   carrying a comment saying so) is **Task 6's starting point under fresh
   review**. `_salvage/internal/config/contributors.go` is the fail-open
   roster loader R2 supersedes, and it is **Task 4's** input as the shape
   being rejected: its missing-file branch returns an empty roster and a
   nil error. `_salvage/internal/serve/serve.go`'s `requireSameOrigin`
   carries the ordering discovery `spec:1206-1210` cites, and it is
   **Task 8's** input as a lesson rather than as code.
10. **`CLAUDE.md` in full**, its "Facts agents get wrong" list in
    particular, and `docs/STATUS.md`'s lane registry and carried items,
    the no-deploy carry among them.
11. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** This
    pass writes Go only. No Python, no Svelte, no TypeScript, no shell,
    no unit file, so `vps-conventions`, `svelte-conventions`,
    `ts-conventions`, and `python-conventions` are all out of scope.

---

## Goal

Ship the half of rung 5's auth core that reads untrusted input, and make
the two halves one working login.

Land the roster store as one interface that opens, validates, answers a
lookup, lists, and adds, so that no caller knows the file is JSON and
rung 6 substitutes its own store by changing one constructor. Land its
write path under one lock with the re-read that keeps a running server
from serving a roster it did not write, and the audit record that is the
only trace a roster change leaves. Land the three limiters over one
bounded evicting map, so neither a single address nor a single client
drives unbounded outbound mail and the map cannot be grown without bound
by attacker-chosen keys. Land `internal/notify` as a general send seam
whose header-injection guard covers every header-valued argument, so
rung 7's reject mail and rung 9's notifications are callers rather than
special cases. Land the origin check once at the shell, after the strip
and after normalization, denying a cookie-authenticated state-changing
request that carries neither `Origin` nor `Sec-Fetch-Site`, with the
bearer bypass `/rest/*` and 2.0's OpenSubsonic surface need. Land the
verify-identity seam that carries the assertion channel with the
identity, so the organizer's rule about cookie-derived identity is
enforceable at one place. Land the three handlers with the consumption
order fixed in the plan rather than discovered in the code. Land the
sweep's interval driver, the six remaining thresholds, and fork 4's
kind answer.

**And land the corrections the 5b-i security read found**, before the
handlers that would otherwise be written against them.

Nothing here renders a page, registers a `member` verb, writes a systemd
unit, edits `deploy.sh`, adds a `doctor` probe row, or touches the box.

---

## Inheritance from pass 5b-i, applied

Each clause of `HISTORY:496-530` and the security review's own
constraints, with the task that carries it.

- **The identity leaf and its address normalizer**, which every keying,
  every logging, and every rate-limit lookup runs through
  (`spec:1161-1163`). **Task 1 makes the normalizer a fixed point**
  (M1), and **Tasks 4, 5, 7, 9 and 10 call it rather than assume a
  caller did**, stated in each task's Consumes line.
- **The token's verifier and the nonce it returns, to be inserted after
  verification and never before.** **Task 10 holds it**, and the
  consumption order is written out in that task rather than left to the
  implementer.
- **The token's version label is the only place a later field, a
  destination included, may be added** (R33, `spec:690-692`). **Task 10
  adds no field and accepts no destination parameter.**
- **The session's minter, verifier, and refresher, with the rule that
  refresh never moves the anchor and that refresh counts same-origin
  requests only** (`spec:1081-1082`, plan review M-e). **Task 10's
  session path applies the same-origin condition to the refresh**, which
  is the clause that sat in the crack between the two halves.
- **The epoch's two refusal checks, the session check reading the anchor
  and never the issue time**, applied to a session on every request and
  to a token at consumption, **plus the roster re-check that refuses a
  link for an address removed from the roster** (R30, `spec:2168`, plan
  review m9). **Task 10 holds all three**, and the roster re-check is one
  of this pass's exit-gate assertions.
- **The seam's `Read` as the authority a credential store stamps from.**
  **Task 9 exposes it through the verify-identity seam** without
  assuming the session store is its only reader (the configuration-homes
  threat model's rung 5 item 1).
- **The state-directory opener and its naming rule, which the roster's
  own leaf sits beside rather than inside.** **Task 4 places the roster
  file at `DUBPLATE_ROSTER_FILE`** and creates no directory; the leaf's
  `ReadWritePaths` grant and its `deploy.sh` pre-create are **5d's**
  (security review m11, which this pass records and does not build).
- **`registerAuthConfig` and the declared key set, which 5b-ii's
  remaining thresholds join rather than re-derive.** **Task 3.**
- **`internal/auth` never reads the environment** (`spec:1471-1479`).
  Every task honors it; resolution happens in `cmd/dubplate-server`.
- **5a's carries, unchanged.** `ClientIP` is 5a's and is never
  re-derived (**Task 7 and Task 10 call
  `internal/serve.ClientIP`**). The origin check inserts after the strip
  and after normalization (**Task 8**). `Origin` and `Sec-Fetch-Site`
  survive ingress, and Go's `CrossOriginProtection.Check` returns nil
  when both are absent (**Task 8 tightens exactly that**). Every URL
  builds from the configured public origin and never `r.Host` (**Task
  10**; `LinkURL` already takes the origin as an argument,
  `internal/auth/token.go:161-169`).
- **The merge carry on `internal/serve/server.go:161`**, the pre-1.22
  goroutine-parameter idiom, stays **forwarded to 5d**. Task 8 opens
  `internal/serve` for the origin check and does not widen into it.
- **The 30-second read timeout and the `http.MaxBytesReader`
  alternative** (R-V, `HISTORY:391-392`). **Task 3 declares the body cap
  threshold and Task 10 applies it**, which is what `spec:1590` asks for
  and what the read timeout bounded until now.
- **The two module dependencies R26 budgets**, `github.com/wneessen/go-mail`
  and `golang.org/x/time/rate`, **both taken in this pass**: go-mail at
  **Task 6** and x/time/rate at **Task 7**, each in the commit that first
  imports it. **No preparatory `go get` task exists**, and no third
  dependency is taken.
- **The nonce sweep's interval driver, a package constant naming its
  one-hour period, is 5b-ii's** (R-5b14, corrected by the errata).
  **Task 11.**

### The three security constraints, carried verbatim

These are the security review's own closing section, quoted rather than
paraphrased, because a paraphrase is where a constraint loses its edge.

> 1. **Fix B1 before the consumption handler is written.** `Insert` must
>    take the token's own `ExpiresAt` and write
>    `max(now+retention, tok.ExpiresAt)` into the record. A handler built
>    against the current two-argument signature will have to be rewritten.
>    The handler's own order is fixed and must be stated in the plan: MAC,
>    expiry, roster resolution, `CheckLink`, `Insert`, and only then
>    `MintSession`, with the session minted only after `Insert` returns
>    nil.
>
> 2. **Every non-nil error from the auth core denies, not only the named
>    sentinels.** M1 makes a non-sentinel error reachable for an address
>    the mint path accepted, and `EpochStore.Read` can also fail on an
>    unreadable or corrupt record. The plan must require that
>    `VerifyLink`, `VerifySession`, `CheckLink`, `CheckSession`, and
>    `Insert` be consumed as "nil means permitted, anything else means
>    refused", with the sentinel distinction used only to choose the
>    server-side log line. One generic outcome reaches the client for all
>    of them.
>
> 3. **Cap the body before the parser sees it, and key nothing on the raw
>    token string.** M5 gives 16x allocation amplification ahead of any
>    structural check, and M3 means one credential has many verifying
>    encodings. Every auth endpoint caps its body with
>    `http.MaxBytesReader` before calling into `internal/auth`, and every
>    limiter, cache, audit correlation id, or dedup key derives from the
>    decoded nonce or the normalized address, never from the token or
>    cookie string. Landing M3's canonical-encoding check in `splitSigned`
>    is the durable half of this constraint.

Constraint 1 is **Task 1's deliverable and Task 10's order**, with the
per-client consumption limiter prepended to the quoted six steps as the
order's first step; the quote is the security review's own words and the
constraint it states is unchanged, but the review left the limiter
unplaced and placement is the whole of that limiter's value (R-5b28's
neighbor, argued in Task 10). Constraint
2 is a **criterion on Tasks 9 and 10**, and a mutation in each. It is
**not** mapped to Task 11: that task's error handling is fail-closed by
construction and it carries no criterion about consuming a non-nil error
as a refusal, so naming it here would put a constraint on a task that does
not carry it, which is one re-dispatch away from being dropped.
Constraint 3 is **Task 1's `splitSigned` half, Task 3's cap threshold,
Task 7's key derivation, and Task 10's `MaxBytesReader` calls**. Its own
words are "**Every** auth endpoint caps its body", and Task 10 builds
three, so **all three carry the cap**: the request-link handler, the
consumption handler, and the sign-out handler, each with its own
criterion.

---

## Owner touchpoints: none

5b-ii raises no attended question. `spec:2646-2652` names rung 5's three
expected touchpoints and all three sit elsewhere: the six product forks
are ruled and closed, the `style-src 'unsafe-inline'` sitting rides with
5c's pages, and the box deploy is 5d's. **Anything else is a conductor
ruling.**

**What would become an owner touchpoint, so a task recognizes one rather
than absorbing it.**

- **Any secret minted.** Rung 5's three secrets are minted at rung 16
  through the age store (`spec:1632-1638`). **5b-ii mints nothing.** A
  task that wants a real secret value stops and reports.
- **Any adopter-visible default.** The six threshold defaults Task 3
  declares, the unknown-address answer's default, the limiter map's
  bound, the sweep's period, and the roster file's schema version are
  what an adopter inherits when they configure nothing. R-5b18 through
  R-5b27 rule all of them as applied, and each states what would change
  under a different answer. **A task changes none on its own judgment.**
- **Any change to what the box runs.** **Nothing deploys this pass.** The
  no-deploy carry is restated in Global Constraints, and it is unchanged
  by anything here: `env/dubplate-server.env.template:74` still selects
  `magic-link` without the nine keys the source requires, so a render
  would refuse every server verb, `shadow` included.
- **Any roster copy a member reads.** Fork 4's kind answer names the
  roster's admin entry (`spec:1553-1555`). **Task 10 returns the contact
  as data and writes no member-facing copy**; the wording is 5c's page
  work under the design-language standard.
- **Any third module dependency.** Two are budgeted and two are taken. A
  task reaching for a third stops and reports.

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition resolves to the conductor at a segment boundary, and
the 80%-of-ceiling flag routes there too.

---

## Token ceiling: 7.5M (re-derived from 5b-i's actual)

> **Conductor-owed fold:** rule the ceiling at 7.5M with the flag at
> 6.0M, or take the Task 8 / Task 9 cut now at drafting time? Applied as
> the 7.5M ceiling with one pass. Revision 1 derived 4.77M from two
> figures no report carries; the honest re-derivation from the one figure
> `docs/HISTORY.md` does carry already clears revision 1's own 6.0M split
> trigger, so the trigger and the ceiling both had to move or the pass had
> to split. The argument for one pass is unchanged and is stated at
> "Where this pass splits, if it splits again": the first half alone ships
> a roster, three limiters, a mail seam and an origin check with nothing
> that calls any of them, which is a surface no security reviewer can
> grade for the property that matters.

`spec:2395-2397` grants the re-derivation in the same sentence that sets
the figures: "Each ceiling is re-derived from the previous pass's actual
report before the next plan is drafted." The previous pass's actual report
is 5b-i's (`HISTORY:531-544`), and it prices a pass of this exact shape in
this exact package set.

**The basis, from the one figure HISTORY carries.**
`docs/HISTORY.md:533-535` is the whole of it: "this window's execution
against the 5.0M ceiling (4.0M flag) ran over, landing at roughly
**5.5M** for Tasks 1 through 10 combined with their fix and simplifier
rounds." That figure covers ten tasks with their fix rounds and the three
batched `code-simplifier` rounds. It is **not** an all-in close figure,
and `docs/HISTORY.md:540-543` says why it cannot be: "the rung 5 security
cycle and the plan-prep tail for 5b-ii were both cut from the window", so
the security read, its fold, and this plan's own preparation all sit
outside that 5.5M. Task 0 sits outside it too.

**The arithmetic.**

1. **5.5M over ten tasks is 0.55M per task**, chain plus fix rounds plus
   the simplifier, measured on this package set by the pass immediately
   before this one.
2. **Twelve tasks here is 6.6M**; add **0.13M** for Task 0, which 5a
   measured (`HISTORY:1688`), for **6.73M** of chain.
3. Add **0.25M** for the pass-end `web-auth-security-reviewer` read,
   **0.20M** for its fold, and **0.30M** for the conductor's turns and the
   merge ritual, for an expectation of **7.48M**. **Rule the ceiling at
   7.5M with the flag at 6.0M.**

**Two figures revision 1 carried are withdrawn.** "5b-i's chain cost
about 3.4M across three segments" has no source: the only 3.4M in the
repo is 5a's 80% flag against a 4.25M ceiling
(`docs/HISTORY.md:1692`). The 0.19M per task it quoted is the
workstation-wide chain rate at
`docs/superpowers/plans/2026-09-12-overnight-window.md:126-130`, measured
across four passes and 74 agents, and that line states plainly that
"Close cost is a separate line at 0.9M-1.4M per pass", so it is not an
all-in rate and the 0.309M derived from it is not one either. 5b-i was
budgeted "2.1M chain + 1.4M close" at
`docs/superpowers/plans/2026-09-12-overnight-window.md:331` and landed at
5.5M for the chain half alone, which is the measurement this ceiling is
built on.

**Why 7.5M rather than the 4.5M provisional.** 5b-i's plan set 4.5M for
5b-ii before the security read existed and before the split's second half
was decomposed. Three things moved since. The read returned a blocker and
five majors whose fold is two tasks this plan did not previously carry;
the honest decomposition of `spec:2168`'s remaining scope is twelve tasks
rather than the eight or nine a 4.5M figure implied; and 5b-i's actual
per-task rate came in at 0.55M against a planning rate of 0.19M. A
ceiling below 7.0M puts the flag inside the chain with nothing the
conductor can do but continue, which is rung 1b's recorded failure mode
(`spec:2387-2389`). At 7.5M the flag lands at **6.0M**, which on this
arithmetic falls in segment D1 or later, where the remaining choices are
real: whether the security read runs in this window or moves, and whether
the close absorbs the spec corrections or hands them forward.

**Re-derive the ceiling from Task 5's actual report**, at the segment B
boundary. Task 4 and Task 5 are the roster pair, the deliverables this
plan predicts rather than measures, and they are where a JSON schema plus
a validation list plus a locked write path can each cost more than a
token construction did. **If the re-derivation projects past 8.0M**, the
conductor names the cut at the **Task 8 / Task 9 boundary**, where the
work changes from mechanism to seam and handler, and 5b-ii-b carries
Tasks 9 through 12.

**Recorded against the rung reservation.** R23 (`spec:569-575`) set rung
5 at 14.5M across four passes and 5a's close moved the live baseline to
**15.75M** (`spec:2042-2050`). Taking actuals for the closed passes and
ceilings for the open ones, rung 5 now stands at **3.90M (5a actual) plus
5.5M (5b-i actual) plus 7.5M (5b-ii ceiling) plus 4.5M (5c) plus 3.0M
(5d), which is 24.4M against the live 15.75M**. The **8.65M** difference
is recorded against the reservation rather than taken from another pass,
because 5c and 5d each own surfaces this pass cannot price.
**Conductor-owed**, and the honest statement behind it is unchanged from
5b-i's: R23 priced undivided 5b's seventeen items at what 5a spent on
eight. **5c's and 5d's own figures are themselves stale**, carried from
the same planning rate 5b-i overran; the conductor re-derives both at this
pass's close.

**Checkpoint: STATUS at every segment boundary**, so after Task 3, after
Task 6, after Task 8, after Task 10, and at the close. Every task reports
its own spend, its mutation outcomes, and its measured gate wall time, and
the conductor sizes the next segment against it. **Every task boundary is
a resumable stopping point** (`spec:2437-2440`).

**Deliverable count: thirteen.** Twelve task deliverables plus the close.
Tasks 1, 2, 3, 5, 10 and 12 carry four each. **Task 4 carries five**, one
above the cap this plan otherwise holds, because revision 2 folded R3's
`doctor` collision row onto it; that fold is conductor-owed and marked at
Task 4's own site. Tasks 4 and 10 are the two to watch.

---

## Global Constraints

- **Lane 1, the main tree, one code lane.** `docs/STATUS.md`'s lane
  registry reads both lanes CLOSED at drafting time, and no second code
  lane is scheduled against this pass. The lane registry entry is written
  before the first dispatch and **is the live-executor check**. If a
  second code lane is later scheduled concurrently, this pass moves to a
  worktree first and no dispatch straddles the move, because the gate
  runs the module leg over `./...` plus three whole-repo legs.
- **No deploy, of anything, until 5d.** `docs/STATUS.md`'s carried item
  states why: the env template selects `magic-link` without the nine keys
  the source requires, so a render would refuse every server verb,
  `shadow` included. **No task runs `scripts/deploy.sh`, not even
  `--check`, and no task ssh's to the box.** A task that believes the box
  must change stops and reports.
- **Package territory, stated as a wall.** This pass writes
  **`internal/auth`** (`roster.go`, `limit.go`, `seam.go`, `handlers.go`
  born; `nonce.go`, `epoch.go`, `token.go`, `session.go`, `store.go`
  edited only for the security fold and the B1 signature, with their
  tests), **`internal/identity/identity.go` and its test** (the
  normalizer's fixed point alone), **`internal/notify`** (born this pass,
  with its tests), **`internal/serve`** (one new origin-check file with
  its test, plus the composition line in `serve.go`'s `Handler` and
  nothing else), **`cmd/dubplate-server/auth.go` and `auth_test.go`**
  (the six thresholds, their validations, the From validation, and the
  resolver),
  **`cmd/dubplate-server/serve.go` and `serve_test.go`** (the auth
  lifecycle and the mounts), **`cmd/dubplate-server/doctor.go` and
  `doctor_test.go`** (Task 2's M4 rendering fix and Task 4's R3 roster
  collision row, and nothing else), **`go.mod` and
  `go.sum`** (the two budgeted dependencies), and
  **`config/roster.example.json`** (one new file, which `deploy.sh` does
  not deliver; see the `config/` row in the contested-surface table). It
  writes **nothing**
  under `internal/config`, `internal/engine`, `internal/harness`,
  `internal/atomicfile`, `corpus/`, `web/`, `systemd/`, `scripts/`,
  `tools/`, `env/`, `Makefile`, or `_salvage/`, and no `docs/` file
  outside Task 12's conductor-owned drafts. **A task that believes it
  needs one of those stops and reports.**
- **`internal/config` is closed to this pass.** 5b-i landed the
  requirement mechanism and the per-source family; 5b-ii declares through
  them from `cmd/dubplate-server/auth.go` and edits neither `config.go`,
  `registry.go`, `thresholds.go`, nor `keys.go`. The rung 1b contract is
  what makes that possible, and a task that wants a config-core change
  stops and reports.
- **`internal/auth` and `internal/identity` read configuration through
  values passed in, never from the environment** (`spec:1471-1479`).
  Neither imports `internal/config`. `internal/identity` imports no other
  internal package at all (`spec:1467-1469`, R17). `internal/notify` reads
  no environment either: its transport takes its five values as a struct.
- **`internal/auth` renders no HTML and registers no route**
  (`spec:1474-1476`). The handlers it provides are `http.Handler` values
  the shell mounts, and the mounting happens in `cmd/dubplate-server`.
- **Module dependencies: exactly two, both budgeted.**
  `github.com/wneessen/go-mail` at Task 6 and `golang.org/x/time/rate` at
  Task 7, each added in the commit that first imports it, each pinned to
  the version the prior-art table records (`spec:1861-1907`), each with
  `go mod tidy` clean in the same commit. **A task that reaches for a
  third stops and reports.**
- **Assertion discipline, at the repo's bar.** No silent-success fakes,
  no vacuous assertions, **no expectation derived from the code under
  test**. Every hard-coded expectation comes from the spec, from an RFC,
  or from documentation a task cites by URL and date.
- **A security property is proved by behavior, never by structure.** A
  test that reads the consumption function's statement order proves the
  code says what it says. A test that presents a link for an address
  removed from the roster and asserts no session cookie was set proves
  the order. Risk 3 (`spec:2526-2531`) is the reason.
- **Every reproduction in the security review is a test in this pass.**
  Each of B1, M1, M2, M3 and M5 carries a reproduction the reviewer ran;
  **each becomes a gate-resident test that fails against the code as it
  stands today and passes after the fix**, and the task reports it having
  been observed failing first.
- **Prefer fixtures that are obviously not credentials**
  (`spec:2378-2380`). Every test secret is a visibly synthetic literal
  long enough to pass the floor, every address is documentation-reserved
  or invented, and no fixture is shaped like a real key. **gitleaks scans
  the tracked worktree with `--no-git`.** The rule 5b-i measured still
  holds: the `generic-api-key` rule is a conjunction of a key-like
  identifier and a high-entropy value, so a synthetic fixture escapes
  through its value and a MAC-bearing golden literal escapes through its
  identifier. **`.gitleaksignore` stays untouched**, and a flagged
  fixture is a fixture chosen wrong before it is an allowlist entry.
- **No token, no secret, and no cookie value ever reaches a log record or
  an error string** (`spec:1184`, "Tokens and secrets never enter a log
  record"; `spec:1946`, "A token never reaches a log record"). The
  consumption route logs its path and never `RawQuery`. An auth event logs
  the reported client IP, the outcome, and nothing else that identifies a
  person beyond what the spec's logging list names. This is a criterion on
  every task from Task 4 onward.
- **Addresses carry one named exception, and exactly one.** The roster
  audit record Task 5 writes names the address added, because
  `spec:1139-1141` requires it: "Every roster change writes one audit
  record: who made it, through which assertion channel, which address was
  added ...". **No other record this pass writes carries an address**, and
  no error string carries one anywhere. Revision 1 stated the constraint
  wider than either clause it cited and the extra width collided with the
  spec's own audit requirement, which is the ambiguity a task would
  otherwise have discovered while writing the only trace a roster change
  leaves.
- **One generic outcome reaches the client for every consumption
  failure** (`spec:1042-1043`, security constraint 2). The distinction
  lives in the server log alone.
- **Universality.** No estate hostname, path, email, catalog value, or
  personal name anywhere (`spec:2654` onward). **Every tunable is a
  threshold with a default.** No address lives in configuration at all,
  because fork 1 moved the admin contact onto the roster record.
- **Drivers own driver knowledge.** Nothing in this pass names a
  streamer, a header, or a path a driver would supply.
- **`go-conventions` before any Go**, at the standing exemplary-stdlib
  bar (`idiomatic-beautiful-go`). Concrete types over interfaces except
  where the spec rules one (R1's loader, R17's seam); `any` not
  `interface{}`; no `init()`; guard clauses over nesting; `ctx` first
  where a call can block or reach a filesystem or the network; errors
  lowercase and context-first with `%w` only where a caller branches.
  **The em dash is banned in Go comments**, and vale scans every new file
  before it is staged.
- **Sentinels where a caller branches**, and never as the only gate. The
  refusals this pass adds (a rate limit tripped, an address absent from
  the roster, a roster validation failure, a send failure) each wrap a
  package sentinel and each test asserts through `errors.Is`. **Security
  constraint 2 still governs the handler: a non-nil error denies whether
  or not it matches a sentinel.**
- **Every task stamps its resolutions' markers.** A task landing the
  surface a resolution decides writes `// ruled, conductor 2026-09-12,
  R-5bN` on the symbol that carries the decision, in the repo's existing
  convention (`internal/config/registry.go:375`, `// ruled, conductor
  2026-09-12, B2`; `internal/auth/nonce.go:103`, `// ruled, conductor
  2026-09-12, A7 fold 4`; `internal/auth/epoch.go:58`). **The grep Task 12
  runs is `grep -rn 'ruled, conductor 2026-09-12, R-5b'`**, so the table's
  construction is unambiguous and a marker written in any other form is
  missed. Revision 1 cited `cmd/dubplate-server/doctor.go:164` as a second
  exemplar; that file carries no `ruled` marker at all, and the citation
  form at `doctor.go:166` is `// (Resolution R-5b6).`, which this grep
  would not find.
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The gate is `bash scripts/check.sh`, green at every task end**, run
  through `cairn-run-gate '<string>'` so the implementer never polls a log
  and the transcript stays small for the reviewer. Never commit red.
  Specific files only. Commit footer `Co-Authored-By: Claude
  <noreply@anthropic.com>`.
- **A pre-flight checklist before every gate**, per CLAUDE.md's gate
  economy: no comment claims what its assertion does not prove; the
  labeled report block verbatim; no process citations in shipped
  comments; counts found, changed, deferred; re-emit before the gate.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each one before
  Task 1 dispatches.

---

## Contested surfaces

`spec:2348-2380` and `PA:175-203` are the sources. The conductor writes
these into the parallel-approach spec's ownership table before the lane
launches, per lane-launch checklist item 2.

| Surface | Owner | Consumers | 5b-ii's position |
| --- | --- | --- | --- |
| `internal/auth` | rung 5, created by it; no other lane writes it | rung 7 and rung 14 through the seam; rung 12 through the seam only | **Completed.** The roster, the limiters, the seam, and the handlers land here, beside 5b-i's token, session, and state records. Nothing in rung 5 writes this package after this pass except 5c's `member` verb, which calls the store and adds no file here. |
| `internal/identity` | rung 5, created by it | rung 6's ops request types, rung 12's driver, 2.0's credential store | **One function changed**, `NormalizeAddress`, to make it a fixed point (M1). No field, no type, and no import changes. |
| `internal/notify` | rung 5, as a general send seam rather than a magic-link mailer, adopted from `_salvage/internal/notify` on merit | rung 7's reject copy, rung 9's monitoring integration | **Born whole, Task 6.** Its subject-and-body contract is what makes rung 7 and rung 9 callers rather than special cases, and this pass owns that generality. |
| The roster store, its interface, and its write path | rung 5 | rung 7 reads it; rung 5's own `member` verb and admin page write it; rung 6 may later back it behind the same interface | **Born whole, Tasks 4 and 5.** 5c's two callers arrive next pass and add no code here. |
| `internal/serve` shell, mux, ingress strip | rung 5, created at 5a; rung 12 thereafter, with no concurrent serve-writing lane | 6, 7, 12, 14 at the declared extension point | **One new file and one composition line, Task 8.** The origin check inserts between normalization and prefix routing. The merge carry on `server.go:161` stays 5d's. |
| `cmd/dubplate-server` root command and `doctor` rows | rung 1's skeleton; rung 5 adds `serve`, the `member` verb group, and the probe row kind | 6, 9, 13 | **Task 3 adds thresholds to the existing `auth.go`, Task 2 fixes `doctor`'s rendering of a joined error, Task 11 wires the serve verb's auth lifecycle.** The `member` verb is 5c's; the probe rows are 5d's. |
| `internal/config` (`config.go`, `registry.go`, `thresholds.go`, `keys.go`) | rung 1b's registry contract | every rung declaring keys | **Untouched by 5b-ii.** 5b-i landed the mechanism; this pass declares through it. |
| `env/dubplate-server.env.template` | rung 4 by its own header; rung 5 adds commented pending rows only, corrected by the 5b-i erratum for one literal row | rung 16 mints | **Untouched by 5b-ii.** The six remaining non-secret rows and the three commented secret rows are 5d's. |
| `config/` | rung 1, as the repo's delivered configuration directory; `scripts/deploy.sh` names each file it delivers rather than syncing the directory (`deploy.sh:284`, `:426`, `:537`) | every rung that ships a config file | **One new file, Task 4**: `config/roster.example.json`, a reference for the shape that `deploy.sh` does not deliver and never will (`spec:751-752`). The incumbent `config/contributors.yaml` beside it is untouched and stays until **rung 16** retires the incumbent importer (`spec:752-754`), not 5d. **This row is new at revision 2**: `config/` had no row here and none in `PA:175-210`, and the conductor writes it into PA's table with the others. |
| `go.mod` | shared | every lane | **Two budgeted additions**, Tasks 6 and 7, each in its importing commit. |
| `.gitleaksignore` | rung 1b retired its one fingerprint; the lane that needs an allowed finding appends and names it in its pass report | every lane | **5b-ii expects to append nothing** and reports it if it must. |
| `docs/` ledgers | conductor-owned | every rung | A pass drafts and never commits. Task 12 drafts under `docs/superpowers/research/handoffs/`. |

**No concurrent code lane is scheduled against this pass.** If one is
launched, the two lanes share the tree through `go.mod`,
`.gitleaksignore`, `cmd/dubplate-server`, and `docs/`, and this pass moves
to a worktree before the second lane's first dispatch.

**The merge ritual closes the lane**, stated here because the close task
executes it: the conductor takes the merge token, rebases onto master,
runs the full gate in-tree, merges `--ff-only`, runs the full gate on
master, and dispatches **one `diff-reviewer` over the merged result
whenever the rebase was not a fast-forward**. With one lane in the main
tree, the rebase is a no-op and the ritual reduces to the two gate runs.

---

## What this pass ships, as outcomes

Every clause below is a property the pass ships. Each names the task that
lands it and the spec clause that requires it. A task builds to this
section; it does not re-derive it.

### D1. A consumed nonce's record outlives every token that could bear it

`NonceStore.Insert` takes the token's own `ExpiresAt` and writes
`max(now + retention, expiresAt)` into the record, so a retention
threshold can only ever lengthen a record's life and never shorten it
below the token that produced it. The load-time relation
(`cmd/dubplate-server/auth.go:151-153`) stays as the second net rather
than as the only one. **Task 1** (security review B1, R31 at
`spec:650-660`, R11 at `spec:421-444`).

### D2. Address normalization is a fixed point, and no credential is minted for an address that cannot be revoked

`identity.NormalizeAddress` re-parses its own output and refuses a value
whose re-parse fails or yields a different string, and it refuses any
normalized address carrying a character below 0x20, 0x7F, or a space.
The package doc's invariant, "the address inside a token's MAC and the
address in a lookup never diverge"
(`internal/identity/identity.go:31-33`), becomes true rather than
claimed. **Task 1** (security review M1 and m10, `spec:1161-1163`, R4).

### D3. The revocation epoch moves forward only

`EpochStore.Bump` reads the stored epoch under the store's own lock and
writes `max(existing, now)`, refusing a backward move rather than
recording one. Revocation is rung 5's only revocation mechanism, and a
backward clock step between an offboarding and a later bump would
otherwise reinstate every session the earlier bump refused. **Task 1**
(security review M2, `spec:1938-1942`).

### D4. One credential has exactly one verifying string

`splitSigned` decodes with a strict base64url decoder, so a field whose
byte length is not a multiple of three no longer has several encodings
that decode to identical bytes and all verify. Its split is bounded
before any structural check, so a value of separators cannot allocate a
substring header per byte. **Task 1** (security review M3 and M5,
security constraint 3).

### D5. `doctor` never prints an all-green table for a configuration `Load` refuses

`rowStatuses` flattens nested joins rather than unwrapping one level,
splits a multi-line member per line, and `identitySourceStatus` consults
the source's `Optional` keys and its hook's own failures as well as
`Required`. **The table keeps one row per declared identity source**, so
R10 stands unchanged. A gate-resident test asserts that no `doctor` run
reports a non-failing table while `Load` returns an error. **Task 2**
(security review M4, R6 at `spec:381-411`, R10 at
`cmd/dubplate-server/doctor_test.go:146-152`).

### D6. The roster fails closed, validates B9's list, and answers from one resolution point

A missing, unreadable, or unparseable roster file is a load error naming
the path, superseding the salvage loader's empty-roster-and-nil-error
(`_salvage/internal/config/contributors.go:33-39`). Validation is B9's
list: unique addresses, addresses parsed and normalized through
`identity.NormalizeAddress`, no control characters, bounded lengths, and
`streamer_username` validated as a stable principal, ASCII-printable,
unique across the roster, and bounded. **The principal's collision check
runs at all three points R3 names** (`spec:325-327`, `spec:1108-1110`):
at startup, in `doctor`, and at write time. The store exposes **exactly
one point where a login address becomes an identity value**, which is
2.0's hook 5.

**The store has two entry points and they refuse opposite things.** A
**read-open** fails closed on an absent file, which is R2. A
**write-open**, which the CLI bootstrap calls, creates the file at 0600
when it is absent and refuses an existing one, which is R37's
`dubplate-server member add --admin` ("the verb creates the file when it
is absent and refuses to create a second one", `spec:749-752`). Revision 1
stated both requirements and named only one constructor, which left a
fresh install needing a handle for a file the only constructor refuses;
the cheap wrong resolution is to soften the loader into the fail-open
shape R2 supersedes. **Task 4** (R1, R2, R3, R37, `spec:1103-1163`,
`spec:749-752`, `S20:367-373`).

### D7. A roster write is all-or-nothing, serialized, and visible without a restart

Every write validates the whole record first, takes the store's lock,
re-checks the address and the principal against the roster as loaded,
writes atomically through `internal/atomicfile.WriteStream`, and
re-reads. **A write that fails validation changes no byte on disk.** Two
concurrent writes both survive. A write whose re-read fails to parse is
an error the caller sees rather than a roster the running server has
already accepted. Every change writes one audit record naming the actor,
the assertion channel, the address added, whether the admin flag was set,
and the reported client IP; that record is the one place in this pass an
address reaches a log. **The write path refuses to set the admin flag on
any assertion channel but the command-line actor**, with its own
sentinel, which is what makes "no web path sets it on a first login"
(`spec:1949`, fork 1 at `spec:191-192`, restated at `spec:1236`) an
enforced invariant rather than a prose claim in a package the caller does
not live in. **Task 5** (R1, R2, R3, R11 at `spec:439-444`,
`spec:1129-1141`).

### D8. Mail is a general seam whose header guard covers every header-valued argument

One method, `ctx` first, taking a recipient, a subject, and a body as
data, with one SMTP transport behind it. The seam owns the
header-injection guard, and the guard covers the subject as well as the
recipient, because a subject is a header value and rung 7's reject mail
and rung 9's notifications will pass subjects derived from filenames and
tag text. The seam carries no copy of its own. Error doctrine is R9:
best-effort for every caller except the login handler, which treats a
send failure as a failed login. **Task 6** (R7, R8, R9,
`spec:1269-1314`).

### D9. Neither one address nor one client drives unbounded outbound mail, and the limiter map cannot be grown without bound

Three limiters over one bounded evicting keyed map: link requests per
address, link requests per client address, and consumption attempts per
client address, with either of the first two tripping the link request.
A global ceiling bounds outbound link mail per interval besides. **The
map's bound holds under a flood of distinct keys**, proved by a
gate-resident test. Keys derive from the normalized address or from
`internal/serve.ClientIP`, never from a raw token or cookie string
(security constraint 3), and never from a header the peer address does
not authorize (R27). **Task 7** (R16, R28, R27, `spec:1165-1189`).

### D10. One origin rule covers every cookie-authenticated state-changing surface, stated once at the shell

Go's `net/http.CrossOriginProtection`, wrapped so that a request carrying
neither `Origin` nor `Sec-Fetch-Site` is denied rather than allowed,
inserted into the shell's composition **after the ingress strip and after
path normalization** and before prefix routing, so every registered
subtree and the fallthrough are covered by construction. A request
authenticated by a bearer credential rather than by the session cookie is
exempt, through **one registration method the wrapper owns**, and **a
bypassed path may not read the session cookie for authentication**.
SEC:36's defect, a check that sat inside one mount and covered no new
surface, is what the placement exists to avoid. **Task 8** (R15,
`spec:1191-1215`).

**Why the wrapper owns the bypass set rather than delegating it.**
`CrossOriginProtection.Check` returns nil for the neither-header case
before it ever consults a bypass (the toolchain's own
`src/net/http/csrf.go:155-158` on go1.27.1, "Neither
Sec-Fetch-Site nor Origin headers are present"), and `isRequestExempt`
(`csrf.go:185-198`) is unexported over a private `c.bypass` field, so a
wrapper cannot ask the stdlib value whether a request matched a
registered pattern. A wrapper that denies whenever `Check` returns nil
and both headers are absent therefore denies a bypassed path's POST too,
and registering the pattern through `AddInsecureBypassPattern` alone
changes nothing on that path. The wrapper keeps its own pattern set in its
own `http.ServeMux` with a sentinel handler, in the shape
`csrf.go:185-190` uses, **consulted before the neither-header denial**,
and its single registration method also calls
`AddInsecureBypassPattern` so the stdlib's own cross-origin branch honors
the same pattern.

### D11. A verified identity carries the channel that asserted it, and the admin flag is read from the roster

The seam yields an `identity.Identity` carrying the login address, the
resolved principal, the display name, the source name, and the assertion
channel. The admin flag is read from the roster record at the seam, never
from anything the request carries. The seam exposes the epoch's `Read` as
the authority a later credential store stamps from, without assuming the
session store is its only reader. **Task 9** (R17, R35,
`spec:1217-1245`, `S20:245-246`).

### D12. The consumption order is fixed, and nothing mints a session before the nonce is consumed

The consumption POST runs, in this order and no other: **the per-client
consumption limiter, then MAC, expiry, roster resolution, `CheckLink`,
`Insert`, and only then `MintSession`**, with the session minted only
after `Insert` returns nil. The limiter's position is part of the order
rather than an unplaced step beside it, because a limiter after the MAC
is never reached by the failing-MAC flood it exists to bound and a limiter
after `Insert` burns a member's nonce before refusing the request. Every non-nil
error from any step denies, and one generic outcome reaches the client
for all of them. The body is capped with `http.MaxBytesReader` before the
parser sees it. The route logs its path and never `RawQuery`. **No
destination parameter is accepted**, and the post-login destination is
the core default (R33). **Task 10** (security constraints 1, 2 and 3,
R30, R32, R33, `spec:1036-1068`).

### D13. The sweep runs on a schedule the serve verb owns

A package constant names the sweep's one-hour period with its reason
stated, and the serve verb's lifecycle starts one goroutine that calls
`Sweep` on it under the verb's context and stops with it. The retention
floor stays the tunable an adopter has a reason to change; the period is
not one. **Task 11** (R-5b14 as the errata correct it,
`spec:1664-1668`).

---

## Resolutions this plan takes rather than discovers

Twelve, numbered from R-5b17, continuing 5b-i's convention. R-5b28 is
revision 2's, folded from the security review. Each is a
question the spec leaves open or answers in a form the tree contradicts.
Each states what was decided, why, and what would change under a
different answer. **Each is ruled by the conductor at dispatch**, on 5a's
and 5b-i's precedent, and none counts as an execution sitting.

### R-5b17: the 5b-i security fold is Tasks 1 and 2 of this pass, not a pass of its own

**Decided:** fold inside 5b-ii, ahead of everything else.
**Conductor-owed.** B1 changes `NonceStore.Insert`'s signature, which
Task 10 calls; M1 makes a non-sentinel error reachable on a path Task 10
branches on; M3 and M5 sit in the parser every auth endpoint reaches; M2
is the only revocation mechanism rung 5 has. Folding them first costs one
segment and makes every later task build against corrected code. A
separate remediation pass would pay a second staleness pre-task, a second
merge ritual, and a second security read for four files. Under the other
answer, the handlers are written twice.

### R-5b18: the 5b-i security read is discharged, and this pass's own read covers both halves at its end

**Decided:** the read at
`docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md` is
R-5b16's 5b-i half, delivered. `docs/HISTORY.md:411-416` and the R24
erratum both say the read was "not run this window" and is "owed to
5b-ii's window as its first item"; both were written before it ran, and
**the close drafts the correction to each**. **One fresh-context
`web-auth-security-reviewer` runs at this pass's end**, covering 5b-ii's
own surface and the corrections Tasks 1 and 2 land, per R-5b16.
**Conductor-owed.** Under the other answer the pass opens with a second
read of a surface just read, and the pass's own new surface, which is
where the input validation lives, gets the weaker grade.

### R-5b19: the roster file is a JSON object carrying a schema version and a members array

**Decided:** `{"schema_version": 1, "members": [...]}` rather than a bare
array. R1 rules JSON through `encoding/json` and says nothing about the
envelope. A version field costs one line and gives rung 6's migration
something to read; a bare array gives it nothing and makes the first
schema change a guess about what a file without a version means. Each
member record carries `address`, `display_name`, `streamer_username`
(optional), and `admin`. Under the other answer, rung 6 pays for the
omission.

### R-5b20: the roster store creates the file but never its directory

**Decided:** `DUBPLATE_ROSTER_FILE` names a path the unit can write, and
the store creates the file with 0600 through its **write-open** entry
point, which refuses an existing file, while the **read-open** entry point
fails closed on an absent one (R2). It creates no directory. R37 and fork
1 give the pre-create to `deploy.sh` and the `ReadWritePaths` grant to the
serve unit, both **5d's** (security review m11). Under the other answer,
the store silently creates a directory outside the unit's grant and the
failure surfaces as a mount-namespacing error nobody planned for, or a
fresh install finds no constructor that can create its first roster and
the loader gets softened into the fail-open shape R2 supersedes.

**The example file's home.** This pass creates
`config/roster.example.json`, which `spec:2332`'s pass-split table lists
in **5d's** row ("the example roster file (R37)"). R37 itself
(`spec:751-752`) assigns no pass, and Task 4's loader test is what keeps
the example from rotting, so the take is defensible; it is recorded here
rather than left unrecorded, and **Task 12 drafts the correction to
`spec:2332`'s 5d row**.

### R-5b21: the audit record is a structured `slog` record, and rung 5 opens no audit file

**Decided:** every roster change logs one `slog` record at info on the
server's own logger, with the actor, the assertion channel, the address
added, the admin flag, and the reported client IP as attributes. The CLI
verb (5c's) writes the same record with the operating-system user in the
actor field. `spec:1143-1149` requires the record and names no
destination. A second durable file would need its own leaf, its own
rotation story, and its own grant, all of which are 5d's surface, for a
record the journal already keeps. Under the other answer, rung 5 grows a
second state file nobody asked for.

### R-5b22: the unknown-address answer is one `KindString` threshold with two accepted values

**Decided:** `DUBPLATE_UNKNOWN_ADDRESS_ANSWER`, `KindString`, default
`kind`, accepted values `kind` and `generic`, validated at load against
that set with an error naming it. Fork 4 ruled the kind answer as the
default with "a threshold flag" switching the whole deployment
(`spec:246-250`). A boolean-named key would read as a negation at the
call site, and R25 calls it "the unknown-address behavior" rather than a
flag. **The handler returns the roster's admin contact as data**; the
copy is 5c's. Under the other answer, an adopter reads
`DUBPLATE_GENERIC_UNKNOWN_ADDRESS=false` and has to work out what true
would have meant.

### R-5b23: the limiter map's bound and the sweep's period are package constants, not thresholds

**Decided:** the map's maximum key count and the sweep's one-hour period
are package constants with their reasons stated in their doc comments.
R16 requires the bound and a test that proves it, and names no key. The
universality rule asks that every **behavior tunable** be a threshold
(`spec:1592-1593`, "Each is a threshold with a default, per the
universality constraint's third test"; the Universality section itself is
`spec:2654` onward), and neither of these is a behavior an adopter has a
reason to tune: a wrong value has no visible symptom, which is R-5b14's
own argument for the sweep period. The four rates, the ceiling and the
body cap are
thresholds, because those are the values an adopter changes. Under the
other answer, an adopter gains two knobs whose wrong settings are silent.

### R-5b24: the three limiters share one map implementation and three key spaces

**Decided:** one bounded evicting keyed map type, instantiated three
times with distinct key spaces rather than one map holding three key
kinds. A shared map with prefixed keys lets a flood of one kind evict
another kind's live buckets, which is a limiter defeating a limiter.
Three instances make each bound its own and make the flood test mean what
it says. Under the other answer, R16's bound holds and R28's two keys
stop being independent.

### R-5b25: the handlers are `http.Handler` values constructed from one dependency struct

**Decided:** `internal/auth` exposes a constructor taking the stores, the
roster, the limiters, the mailer, the resolved durations, and the
logger, returning the three handlers as values. `spec:1474-1476` requires
that the package register no route, and a dependency struct is what keeps
the mount in `cmd/dubplate-server` where the wiring already lives
(`newServeShell`, `cmd/dubplate-server/serve.go:158-161`). Under a
package-level registration, the shell's composition stops being the one
place routing is stated and 5a's whole argument weakens.

### R-5b26: rung 5 builds no redirect-target validator, because no rung 5 caller supplies a destination

**Decided:** R33's positive validation rule travels as a **constraint**
and lands no function. `spec:1258-1265` gives the rule to rung 5 "with
the mechanism", and rung 5 accepts no caller-supplied destination, so
there is no value to validate. **The erratum landed at `b91bae1` claims
R33's redirect validation "landed in `internal/auth` at 5b-i"; grep
proves otherwise**, and Task 0 item 10 confirms it: no such function
exists. **The close drafts the correction**, stating that the validator
lands with rung 12's parameter, inside the MAC-covered token. Under the
other answer, this pass ships a validated-but-uncalled function and a
reviewer has to decide whether dead security code is better than none.

### R-5b27: the origin check inserts between path normalization and prefix routing

**Decided:** in `Shell.Handler` (`internal/serve/serve.go:150-163`), the
wrap goes around the routed handler and inside `Normalize`, so the order
becomes the ingress strip, normalization, the origin check, then routing
and the fallthrough. R-F and `HISTORY`'s 5a carry require "after the
strip and after normalization"; placing it inside routing would repeat
SEC:36's defect, and placing it outside normalization would let a
`..`-bearing or raw-versus-decoded-mismatched path reach the check.
Under the other answer, a surface added at rung 7 or rung 14 is uncovered
until someone remembers.

**The bypass set is the wrapper's own.** `AddInsecureBypassPattern` alone
cannot carry the exemption, because `Check` returns nil for a
neither-header request before consulting any bypass and the stdlib's
`isRequestExempt` is unexported over a private field, so the wrapper
cannot ask it what matched. The wrapper therefore keeps its own pattern
set, and its one registration method feeds both its set and the stdlib's.
D10 carries the evidence.

### R-5b28: R9 governs the `kind` deployment, and `generic` answers before the send

> **Conductor-owed fold:** does R9's "a send failure is a failed login"
> govern a `generic` deployment, or does `generic` answer before the send?
> Applied as the narrow reading, R9 governing `kind` alone.

**Decided:** under `DUBPLATE_UNKNOWN_ADDRESS_ANSWER=generic`, the
request-link handler answers before the send completes, and a known
address and an unknown address produce byte-identical responses.
**Conductor-owed.** Two ruled clauses collide here. R9 says "a send
failure is a failed login", and fork 4's generic outcome exists to stop a
stranger learning whether an address is on the roster (`spec:244-247`).
Under `generic` with R9 applied literally, an unknown address returns the
generic outcome while a known address whose send fails returns a failure,
and that difference **is** membership. The timing channel is the same
leak: a known address runs a mint, a URL construction and an SMTP dial
before answering, and an unknown one answers immediately, which is a
network round trip to a mail provider rather than a subtle side channel.
The narrow reading keeps both rulings: R9 is the `kind` deployment's rule,
where the answer already distinguishes known from unknown and a send
failure has nothing left to leak; under `generic` the response does not
wait on the send, and a send failure is logged rather than rendered.
Under the other answer, the generic mode is a mode whose only purpose the
implementation defeats. **Task 10** holds it, with its own criterion and
its own mutation.

---

## The threat model, as the in-pass reviewer's rubric

The reviewer receives `spec:1909-1980`, this section, the 5b-i security
review's "Attacked and found sound" list, and the diff. It does not
receive this plan's task decomposition. A finding that contradicts the
spec's threat model is a finding against the spec.

**What this pass protects.** The roster's contents, which fork 1 made a
thing an attacker would want to write rather than only to read. The
outbound mail budget, and through it the addresses on the roster. The
single-use property of a consumed nonce, now across a configuration
change as well as a restart. The revocation epoch's monotonicity. The
one privileged action rung 5 has.

**Who the adversary is, for this pass's surface.** An unauthenticated
visitor reaching the login POST and the consumption POST from the
internet, since both are public by construction. An enumerator walking
distinct addresses to learn who is on the roster. A mail gateway that
follows links. A signed-in member without the admin flag reaching for the
roster. A misconfigured or hostile proxy in front of serve.

**The claims this pass's code must hold**, drawn from `spec:1936-1958`
and from the security review's own three constraints:

1. A nonce cannot be consumed twice, across a restart **and across any
   accepted retention configuration, including one tightened after a
   token was minted**.
2. Every non-nil error from the auth core denies, and one generic outcome
   reaches the client for every consumption failure.
3. No auth endpoint parses a body it has not capped, and no limiter,
   cache, or correlation id keys on a raw token or cookie string.
4. A limiter map cannot grow without bound, and neither a single address
   nor a single client drives unbounded outbound mail.
5. A state-changing cookie-authenticated request without a same-origin
   signal is refused, and a bypassed path does not read the session
   cookie for authentication.
6. A roster change that fails validation leaves the file byte-identical;
   one that succeeds writes an audit record naming the actor and the
   channel; two concurrent writes both survive.
7. A missing, unreadable, or unparseable roster file refuses the start
   rather than degrading it.
8. A link for an address removed from the roster is refused at
   consumption, and so is one minted before that address's epoch.
9. A recipient or a subject carrying CRLF is refused by the seam.
10. An epoch bump never moves an epoch backward.
11. **Only the command-line actor can set the admin flag.** The roster's
    write path refuses the flag on every other assertion channel, so "no
    web path sets it on a first login" (`spec:1949`) is enforced where the
    write happens rather than asserted in a caller's prose.
12. **Under the generic unknown-address answer, a known address and an
    unknown address are indistinguishable**, in the response bytes and in
    the time the response takes (R-5b28).

**Attacks the rubric names, and what answers each.** A replay of a link
after a retention tightening, answered by D1's `max`. A credential minted
for an address that cannot be revoked, answered by D2's fixed point. A
revocation undone by a backward clock step, answered by D3. A limiter or
audit key defeated by mutating one base64 character, answered by D4. An
allocation flood ahead of any structural check, answered by D4's bounded
split and D12's body cap. Roster enumeration across distinct addresses,
answered by D9's per-client key and global ceiling, which is the
mitigation fork 4's kind answer rests on. A limiter map grown by
attacker-chosen keys, answered by D9's bound and R-5b24's separate key
spaces. A membership oracle read out of the
generic answer's response bytes or its timing, answered by R-5b28's
equivalence requirement, which the per-client key and the ceiling bound
the rate of but do not close. A cross-site POST adding a member, answered
by D10. A member without the admin flag adding a member, answered by
D11's roster-sourced flag and by D7's channel refusal at the write path. An outstanding link surviving an offboarding, answered by D12's
roster resolution and `CheckLink`. A header-injected mail recipient or
subject, answered by D8's guard.

**What this pass cannot answer, and does not claim to.** The add-member
page and its non-admin 404 are 5c's. The forward-auth verify endpoint and
its allowlist are 5d's. The serve unit's `ReadWritePaths` grant over the
auth and roster leaves is 5d's (security review m11), so a finding that
the leaf is ungranted is recorded as a 5d requirement rather than a
5b-ii defect. The env template's nine missing keys are 5d's (security
review m13). The two declared 5b-i gaps that no test can reach, the
constant-time comparison and crash consistency, stay declared.

---

## Task 0: The staleness pre-task

Model: **`opus`**, **fresh context, zero prior context**, dispatched by
the conductor before any implementer. Reads the rung 5 spec, the security
review, and this plan against the tree as shipped. **Its output is
blockers with evidence, not fixes**; it writes no code and edits no file.
The conductor rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's `spec:`, `HISTORY:`, `PA:`, `S20:` and code
pointers were resolved at **`b91bae1`**. A citation that has moved since
is a finding for this task rather than a surprise for an implementer.

**Grade a moved pointer by whether its content moved.** A pointer off by
fewer than about ten lines whose cited content still resolves to the same
subject is a note in the report. A pointer landing on a different
subject, or naming content that no longer exists, is a blocker.

The checklist is twenty-three items, from the spec's own seventeen
bullets (`spec:2443-2507`) reduced to what this pass reaches, extended
with the twelve this pass adds. **The first four are blockers: a "no"
stops the launch.**

1. **Is the lane registry consistent with where this pass intends to run,
   and is the target tree free?** `docs/STATUS.md`'s `## Lane registry`
   table is the live-executor check, **never `pgrep -f
   Projects/dubplate`**. *(Expected: both lanes CLOSED; lane 2's worktree
   removed and branch deleted. Run `git worktree list` and report what it
   shows; the STATUS row says the worktree was "to be removed by the
   conductor", so a lingering worktree is a note and a lingering
   **executor** is a blocker.)* **Report any uncommitted change under
   `internal/auth`, `internal/identity`, `internal/serve`,
   `cmd/dubplate-server`, or `go.mod` that you did not author.**
2. **Has any other lane written `internal/serve` since 5b-i's merge?**
   `PA:80` forbids it until rung 5 closes. *(Expected: no commit touching
   it since `1d4ed29`.)* **Report the commit list.**
3. **Is the security review present and unchanged?**
   `docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md`,
   493 lines at drafting time, untracked. **Report its line count, its
   tracked status, and whether the conductor has committed it.** This
   plan cites it as a primary input; an absent or altered file is a
   blocker, because Tasks 1 and 2 are built from it.
4. **Do the six named findings still reproduce at HEAD?** Report each by
   file and line, and report any that a later commit already fixed.
   *(Expected, all six live:* B1 at `internal/auth/nonce.go:117`, which
   writes `nonceRecord(now, s.retention)` with `Insert`'s signature at
   `:104` taking only a nonce and a time; M1 at
   `internal/identity/identity.go:34-43`, returning
   `strings.ToLower(addr.Address)` with no re-parse and no character
   filter; M2 at `internal/auth/epoch.go:80-98`, writing `now` without
   reading the stored value; M3 at `internal/auth/token.go:244-258`,
   decoding each part with `base64.RawURLEncoding.DecodeString`; M4 at
   `cmd/dubplate-server/doctor.go:88-112`, unwrapping `Load`'s joined
   error exactly one level, with `identitySourceStatus` at `:167-177`
   consulting `Required` alone; M5 at `internal/auth/token.go:245`,
   calling `strings.Split` with no bound.*) **A finding that no longer
   reproduces is a blocker for the task that was to fix it**, because the
   task's test would then pass before the change.
5. **Does `cmd/dubplate-server/auth.go` still carry the shapes Task 3
   extends?** *(Expected: `secretLengthFloor` at `:68`; the four
   thresholds at `:75-100`, `KindDuration`, defaults `15m`, `2160h`,
   `4320h`, `24h`; `registerAuthConfig` at `:108`;
   `validateMagicLinkSource` at `:124-156`; `validatePublicURL` at
   `:171-184`; `isLoopbackHost` at `:189`; `resolveAuthConfig` at
   `:207`.)* **Report each, and report `resolveAuthConfig`'s current
   return shape**, which Task 3 extends and which is the signature Task
   11 calls.
6. **Does `internal/config`'s registry still expose the family Task 3
   declares through, unchanged?** *(Expected: `IdentitySource` at
   `registry.go:349-377` with `Name`, `Selector`, `Required`, `Optional`,
   and `Validate func(values map[string]string, environ []string) error`;
   `Threshold` at `thresholds.go:45-51`; the four kinds at `:16-21`,
   `KindInt`, `KindDuration`, `KindFloat`, `KindString`.)* **Report all
   three.** This pass writes nothing under `internal/config`, so a change
   here is a blocker for Task 3 rather than a chore.
7. **Is `internal/serve`'s composition still the shape Task 8 inserts
   into?** *(Expected: `Shell.Handler` at `serve.go:150-163`, building
   `routed := s.prefix.Route(...)`, `guarded := refuseEchoMethods(routed)`,
   `normalized := Normalize(guarded)`, `stripped := s.strip.Wrap(normalized)`,
   returning `s.logRefusal(stripped)`; `Register` at `:80`;
   `SetFallthrough` at `:136`; `stripSessionCookie` at `:225`.)*
   **Report the exact composition lines**, because R-5b27's insertion
   point is stated against them.
8. **Are `ClientIP` and the strip still 5a's, unchanged?** *(Expected:
   `internal/serve/clientaddr.go:102`, `internal/serve/strip.go:300`'s
   `NewStrip(header, sessionCookie string, proxy TrustedProxyConfig,
   exemption FrontExemption)`, and `cmd/dubplate-server/serve.go:158-161`
   passing `auth.SessionCookieName`.)* **No task in this pass derives a
   client address**, and a task that believes it needs one stops and
   reports.
9. **Confirm on the tree's own toolchain, measured rather than
   remembered:** that `net/http.CrossOriginProtection` exists with
   `Check`, `Handler`, `SetDenyHandler`, and `AddInsecureBypassPattern`;
   that `base64.Encoding.Strict` exists and that
   `base64.RawURLEncoding.Strict().DecodeString` refuses a value with
   non-zero trailing bits; and the `go` directive in `go.mod` *(expected
   `go 1.27.0` at `go.mod:3`)*. **Report any divergence**, because Task
   1's D4 and Task 8's D10 rest on the first two.
10. **Does any redirect-target validation function exist in
    `internal/auth`?** *(Expected: none. Grep the package for
    `redirect`, `destination`, and `Destination`.)* The erratum landed at
    `b91bae1` states that R33's redirect validation "landed in
    `internal/auth` at 5b-i". **Report what grep actually finds**, and
    report whether `MintLink` (`internal/auth/token.go:118`) takes any
    destination argument *(expected: it does not)*. R-5b26 disposes it
    and the close drafts the correction.
11. **What does this pass's registry arithmetic do to `doctor`'s
    line-count assertions?** **Compute and report**: this pass adds six
    thresholds and no wall row, no reserved class, and no integration
    class, for a net of plus six on the production registry. *(Expected
    sites: the extra-line assertion and
    `TestDoctorCmd_LineCountFollowsRegistry` in
    `cmd/dubplate-server/doctor_test.go`; report their current line
    numbers.)* **Then report which tests build through
    `productionRegistry()` and which through `config.NewRegistry()`**,
    because a family the fresh registry never populates passes both
    assertions vacuously. **If either assertion would fail, that is a
    blocker for Task 3 and the conductor rules it before dispatch.** This
    is the blast-radius class 5a's one planning miss belonged to
    (`HISTORY:526-529` of the 5a entry).
    **Also reconcile Task 2's M4 against R10 before Task 2 dispatches.**
    R10 is ruled and gate-pinned at
    `cmd/dubplate-server/doctor_test.go:146-152` and asserted at `:232-242`
    as "one per declared identity source regardless of how many keys that
    source requires". **Report whether M4 as this plan now states it
    (flattening the join and consulting `Optional`, with no per-key rows)
    leaves that assertion's meaning intact**, and report whether the
    scratch source in that test is declared but never selected, because a
    per-key expansion scoped to the selected source would pass the test
    without honoring the rule.
12. **Is the no-deploy carry still live, and does the env template still
    select a source it cannot satisfy?** *(Expected:
    `env/dubplate-server.env.template:74` sets
    `DUBPLATE_IDENTITY_SOURCE=magic-link`, and the nine keys
    `magicLinkSource.Required` names are absent from the file; security
    review m13.)* **Report the line and the grep result.** **This pass
    writes no template row**, so a change here is a signal rather than a
    blocker.
13. **Is `go.mod` still cobra-only?** *(Expected: `github.com/spf13/cobra
    v1.10.2` plus two indirects, `go.mod:5-10`.)* **Then report, without
    modifying the module, what `go list -m -versions` gives for
    `github.com/wneessen/go-mail` and `golang.org/x/time/rate`**, and
    whether the versions the prior-art table records, v0.8.1 and v0.16.0,
    are still the latest. **A newer version is a note for Tasks 6 and 7,
    not a blocker**; the plan's adoption evidence is dated 2026-09-08 and
    the task records what it actually took.
14. **Is `_salvage/internal/notify` unchanged?** *(Expected:
    `notify.go` at 96 lines with `Mailer`, `New`, `Send`,
    `implicitTLSPort = 465`, and two test-only fields each carrying a
    comment saying so, plus `notify_test.go`.)* **Report both files'
    line counts and the exported surface.** Task 6 adopts it under fresh
    review.
15. **Is `_salvage/internal/config/contributors.go` still the fail-open
    loader R2 supersedes?** *(Expected: the missing-file branch returning
    an empty roster and a nil error at `:33-39`, and the per-request
    membership re-check at `:93`.)* **Report both**, because D6 is stated
    against the first and R30's session-path re-check cites the second.
16. **Does `_salvage/internal/serve/serve.go` still carry
    `requireSameOrigin`, and what is its ordering?** *(Expected:
    `Sec-Fetch-Site` honored on its own, with `Origin` against `Host` as
    the fallback, per `spec:1206-1210`.)* **Report the function and its
    order.** Task 8 adopts the lesson and not the code.
17. **Does `internal/atomicfile` still offer both entry points, and does
    `WriteStream` still fsync the containing directory?** *(Expected:
    `WriteFile` at `:16`, `WriteStream` at `:55-91`, the directory open
    and `Sync` at `:82-89`.)* **Report the real ranges**, because D7's
    roster write rests on the second half.
18. **Do the HISTORY sections this plan cites still resolve?**
    *(Expected: the 5b-i entry at `:5-544`, its exported surface at
    `:168-184`, its contract coverage at `:195-207`, its resolutions at
    `:208-235`, its declared gaps at `:368-402`, its review record at
    `:403-431`, its errata list at `:432-495`, and the 5b-i to 5b-ii
    handoff at `:496-530`.)* **Report each range.**
19. **Did the eighteen spec errata land, and does any of them contradict
    this plan?** *(Expected: `b91bae1` adds 162 lines to
    `docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`, all
    under "Errata the close applies".)* **Report the count of dated
    erratum blocks and name any whose content this plan's text
    contradicts**, item 10's R33 claim included.
20. **Is `.gitleaksignore` still the tracked empty file rung 1b left**,
    and has any lane appended to it? **Report the line count without
    quoting any fixture literal.**
21. **Do `CLAUDE.md`, `scripts/check.sh`, and `Makefile` still carry the
    "rejoins at rung 7" sentence, and is `scripts/deploy.sh` still
    delivering `config/contributors.yaml`?** **Pass 5c corrects the first
    three and R37 leaves the fourth alone.** Report all four and change
    none.
22. **Is the gate's wall time still what this plan sizes against?** Run
    `go test ./...` once at `-count=1` and `bash scripts/check.sh` once,
    and report both totals. *(5b-i measured the full gate at 3m20s.)*
23. **Does `validateMagicLinkSource` validate `DUBPLATE_SMTP_FROM`
    today?** *(Expected: no. It checks the four secret lengths, calls
    `validatePublicURL`, and compares the nonce retention against the
    link expiry, `cmd/dubplate-server/auth.go:124-156`.
    `DUBPLATE_SMTP_FROM` is a Required key of the magic-link source at
    `auth.go:52` and no function reads it.)* **Report what the function
    validates, key by key.** Task 3 lands the check and this absence is
    the reason; a check that already exists makes that outcome a no-op
    the task reports rather than a duplicate it writes.

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
beyond item 22's measurement; the task writes nothing.

**2.0 hooks: neutral.** The pre-task builds no surface.

**Prior art: none consulted.** The task reads the tree.

**Independent: no.** Every task depends on it.

---

## Task 1: The 5b-i security fold, the blocker and the four reachable majors

Model: **`opus`**. Depends on Task 0.

The five defects the security review reproduced by running the landed
code, fixed before anything is built on top of them. This task is why
segment A exists.

**Files:**
- Modify: `internal/auth/nonce.go`, `internal/auth/nonce_test.go`
- Modify: `internal/identity/identity.go`,
  `internal/identity/identity_test.go`
- Modify: `internal/auth/epoch.go`, `internal/auth/epoch_test.go`
- Modify: `internal/auth/token.go`, `internal/auth/token_test.go`,
  `internal/auth/session_test.go` (the session parser shares
  `splitSigned`)

**Interfaces:**
- Consumes: nothing from this pass.
- Produces: `NonceStore.Insert` with the token's expiry as a third
  parameter; a fixed-point `identity.NormalizeAddress`; a monotonic
  `EpochStore.Bump`; a strict, bounded `splitSigned`. **Task 10 is the
  first caller of the new `Insert` signature**; every later task calls
  the normalizer.

**Outcomes (D1, D2, D3, D4):**

- **B1.** `Insert` takes the token's own `ExpiresAt` and writes
  `max(now + retention, expiresAt)` into the record, so a retention
  threshold can only lengthen a record's life. The load-time relation at
  `cmd/dubplate-server/auth.go:151-153` stays as the second net. The
  record's parse and the sweep read the same carried moment they read
  today; only what is written changes.
- **M1 and m10.** `NormalizeAddress` re-parses its own output and refuses
  a value whose re-parse fails or yields a different string, and refuses
  any normalized address carrying a byte below 0x20, 0x7F, or a space.
  The package doc's stated invariant becomes true. **The length cap is
  checked on the input as it is today**, and the fixed-point check is
  additional rather than a replacement.
- **M2.** `Bump` reads the stored epoch **inside the store's own mutex,
  across the read and the write**, and writes `max(existing, now)`. A
  backward move writes nothing and reports the refusal to the caller
  through a package sentinel, so a caller can distinguish "already newer"
  from "write failed". The subscriber notification fires only when the
  stored value actually moved.
- **M3.** `splitSigned` decodes with a strict base64url decoder, so one
  credential has exactly one verifying string. **The strict decoder is
  the adoption** (Prior art below); the review's re-encode-and-compare is
  the same property spelled longer.
- **M5.** The split is bounded before any structural check, so the
  existing exact-count check stays exact and a value of separators cannot
  allocate a substring header per byte.

  > **Conductor-owed fold:** may M5 ship without a failing-first test if
  > the allocation benchmark cannot be made to fail against the unfixed
  > code, recorded as a declared gap? Applied as the named fallback below,
  > which the task reports rather than decides.

  **M5's reproduction needs a named measurement, because behavior alone
  cannot separate the two forms.** `splitSigned`
  (`internal/auth/token.go:244-248`) reads
  `parts := strings.Split(value, fieldSeparator)` and refuses at
  `if len(parts) != count+1` on the next line. A separator flood returns a
  huge `len(parts)` and fails that check; a bounded split
  (`strings.SplitN` with `count+2`) returns `count+2` parts and fails the
  same check at the same line with the same sentinel. The decode loop at
  `:250-256` is never reached either way, so M3's strict decoder does not
  separate them, and `testing.AllocsPerRun` counts allocations rather than
  bytes while `strings.Split` allocates one slice plus one backing array
  regardless of element count, so an allocation **count** is flat across
  the mutation. **The assertion is
  `testing.Benchmark(...).AllocedBytesPerOp` against a threshold
  proportional to the input length**, with the threshold written out as a
  literal and its reason stated. **If the task cannot make that assertion
  fail against the pre-change code, it does not invent one**: it reports
  M5 as a defect fixed without a failing-first test, with the measurement
  it tried, and Task 12 records it in the declared coverage gaps with that
  reason. What the task must not do is leave an acceptance row the unfixed
  code already satisfies.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- **Each of the five carries the reviewer's own reproduction as a
  gate-resident test**, and the task reports having observed each one
  fail against the pre-change code before the change landed. The five:
  a 1-second retention with a longer-lived token, insert, sweep past the
  retention, and insert again refused; `"a b"@example.com` refused at
  mint rather than accepted and unrevocable; a bump at T+1m then a bump
  at T-1h leaving `CheckSession` still refusing; a minted link whose
  final character is substituted across the base64url alphabet failing
  `VerifyLink` for every substitution that is not the original; and a
  value of separators whose **allocated bytes per operation**, measured
  through `testing.Benchmark(...).AllocedBytesPerOp`, stay under a
  written-out threshold proportional to the input length. **The last row's
  assertion is on bytes and not on the returned sentinel**, because the
  unfixed code already returns the right sentinel and an assertion on it
  passes before the change.
- `Insert`'s new parameter is the token's expiry, not a precomputed
  moment, so the `max` lives in one place inside the store.
- The normalizer's idempotence is asserted over the whole existing test
  corpus, `Normalize(Normalize(x)) == Normalize(x)` for every row, plus
  the quoted-local-part row and a TAB-bearing row.
- `Bump`'s read and write are inside one critical section, proved by a
  concurrent test in which two bumps with different timestamps leave the
  later value stored, run with `-race`.
- No error string added by this task carries an address, a token, or a
  nonce.
- Every changed doc comment states the property rather than the change,
  carries no em dash, and cites no process.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Write `now.Add(retention)` rather than the `max` → the B1
   reproduction fails.
2. Drop the re-parse from the normalizer → the quoted-local-part row
   fails at `Bump` rather than at mint.
3. Drop the control-character filter → the TAB row fails.
4. Write `now` unconditionally in `Bump` → the backward-bump row fails.
5. Move `Bump`'s read outside the mutex → the `-race` concurrent row
   fails or reports a data race.
6. Decode with the non-strict decoder → the character-substitution row
   fails.
7. Restore the unbounded split → the separator-flood row's
   allocated-bytes assertion fails. **If it does not**, the measurement is
   wrong rather than the mutation, and the task reports that rather than
   recording a mutation it never observed.

**2.0 hook: neutral, with a recorded forward.** The normalizer is the
derivation rung 6 replaces on hook 5's read side, so making it a fixed
point is what lets that substitution be a change of one function rather
than an audit of every caller.

**Prior art:** **adopts `base64.Encoding.Strict`** (stdlib; documented as
rejecting non-zero trailing padding bits, which is exactly M3's
malleability). **Declines** the review's re-encode-and-compare form on
the criterion of idiom, not correctness: it is three lines where the
stdlib offers one, and a reader has to reconstruct why the comparison
exists. Cites RFC 4648 section 3.5 by name and read date for the
canonical-encoding requirement.

**Independent: no.** Task 10 calls the new signature; every later task
calls the normalizer.

**Deliverable count: 4** (the `Insert` signature and its `max`, the
normalizer's fixed point, the monotonic bump, the strict bounded parser).
Gate green; commit.

---

## Task 2: The 5b-i security fold, the minors and `doctor`'s rendering

Model: `sonnet`. Depends on Tasks 0 and 1.

Twelve findings with no cryptographic content, folded together because
each is small and every one of them touches a file this pass will not
open again.

**Files:**
- Modify: `internal/auth/session.go`, `internal/auth/session_test.go`
- Modify: `internal/auth/store.go`, `internal/auth/store_test.go`
- Modify: `internal/auth/nonce.go`, `internal/auth/nonce_test.go`
- Modify: `internal/auth/epoch.go`, `internal/auth/epoch_test.go`
- Modify: `cmd/dubplate-server/auth.go`, `cmd/dubplate-server/auth_test.go`
- Modify: `cmd/dubplate-server/doctor.go`,
  `cmd/dubplate-server/doctor_test.go`

**Interfaces:**
- Consumes: Task 1's changed signatures.
- Produces: a session cookie carrying `Max-Age`; **`RefreshSession`
  taking the absolute cap and returning an expiry clamped to the
  anchor-derived ceiling**, which is an exported signature change Tasks 9
  and 10 call (today `func RefreshSession(secret []byte, s Session, now
  time.Time, lifetime time.Duration) (string, Session)` at
  `internal/auth/session.go:130`, with the cap living one function away on
  `VerifySession`'s parameter list at `:167`); **a cookie-clearing helper**
  Task 10's sign-out handler calls; a
  store that refuses a too-wide leaf; a `RecordID` carrying a per-kind
  label; an epoch subscriber call that cannot take a revocation down; four
  bounded thresholds; a `doctor` that flattens nested joins.

**Outcomes (D5):**

- **m1 and m2.** `SessionCookie` sets `MaxAge` alongside `Expires`, and
  `RefreshSession` clamps the new expiry to `min(now + lifetime, anchor +
  cap)` so the browser is never told a value the server will refuse.
- **m9.** A cookie-clearing helper emits the same `Path`, `Secure`,
  `HttpOnly`, `SameSite`, and absent `Domain` the `__Host-` prefix
  requires, with `MaxAge` negative. **Its doc comment states plainly that
  these sessions are stateless**, so clearing the cookie does not
  invalidate the signed value and a session captured before sign-out
  keeps working until its expiry or an epoch bump. 5c's logout copy is
  bound by that sentence.
- **m3.** `Open` stats the leaf after `MkdirAll` and refuses to start on a
  mode wider than `dirPerm`, naming the path and the mode it found.
  Record **names** are digests, but a listable directory is still a
  listing of how many links are outstanding.
- **m6.** `Store.Write` is deleted. It has no caller, it is
  overwrite-shaped, and it sits beside a store whose whole correctness
  argument is that the nonce record is never overwritten.
- **m7.** `recordPath`'s refusal reports the id's length and the offending
  character class, never the value, so a future caller that skips
  `RecordID` cannot print a raw nonce into an error string
  `internal/auth/doc.go:14-15` promises never carries one.
- **m12.** `RecordID` hashes a per-kind label alongside the value, the
  same discipline `macInput` already applies, so two record kinds sharing
  a directory could not silently collide.
- **m8.** `Sweep` logs and continues past an individual unreadable
  record, and `OpenNonceStore` fails closed only on the store's own
  directory being unreadable. A record's **presence**, not its
  readability, is what blocks a replay, so one bad inode must not take
  the login surface down.
- **m14.** `Bump`'s subscriber notification runs each callback with a
  recover and a deadline, so a panicking or blocking subscriber cannot
  propagate into or stall a revocation. The write is already durable
  before the notification, which stays the right order.
- **m4.** `validatePublicURL` refuses a non-nil `u.User`, a non-empty
  `RawQuery`, and a non-empty `Fragment`, so no configured origin can
  produce a link whose visible prefix is an attacker-chosen name.
- **m5.** The four existing duration thresholds gain bounds: a positive
  floor on all four, a ceiling on the link expiry, and a ceiling on the
  session absolute cap. The threat model's mail-scanner residual is
  bounded by "single use, a 15-minute default expiry", and nothing
  enforced the second half.
- **M4.** `rowStatuses` flattens nested joins by recursing through
  `Unwrap() []error` and splitting a multi-line member per line, and
  `identitySourceStatus` consults the source's `Optional` keys and its
  hook's own failures as well as `Required`. **The table keeps one row per
  declared identity source**, whatever that source's key count.

  > **Conductor-owed fold:** does M4 drop the per-key rows, preserving
  > R10's one-line-per-source rule, or does R10 get amended to admit
  > per-key rows for the selected source? Applied as dropping the per-key
  > rows.

  Revision 1 asked for "one row per unsatisfied identity-source key",
  which revokes a ruled, gate-pinned invariant no part of the plan
  reconciled. R10 is stated at
  `cmd/dubplate-server/doctor_test.go:146-152`, "one per declared identity
  source regardless of how many keys that source requires ... ruled,
  conductor 2026-09-07, R10", and asserted at `:232-242` with the failure
  message "one line per source, not per key". **D5's load-bearing
  property does not need per-key rows**: "no `doctor` run reports a
  non-failing table while `Load` returns an error" is carried by the
  flattening and the `Optional` consultation alone. The per-key expansion
  would also have passed R10's test by accident, because that test's
  scratch source is declared but never selected, which is worse than
  failing it. The operator's need for a name rather than a verdict is met
  inside the source's single row, whose text may name the unsatisfied keys
  without adding a line; **the task states in its report which form it
  chose and confirms the line count is unchanged.**

**Acceptance criteria a diff-reviewer verifies from the diff:**

- **The M4 gate test is the load-bearing one**: with the magic-link
  source selected and only `DUBPLATE_NONCE_RETENTION` below the link
  expiry, `doctor` must not report a non-failing table. Assert the
  property generally, that no `doctor` run reports every row satisfied
  while `Load` returns an error, rather than asserting one string.
- The rendered cookie header is asserted by printing `Cookie.String()`
  and comparing against a written-out expectation, `Max-Age` included,
  not by reading the struct fields back.
- The clamp is asserted at the boundary: anchored at T, 30h cap, 24h
  lifetime, refreshed at T+23h, expiry exactly T+30h.
- **R10's line count is unchanged by this task**, asserted by running
  `TestDoctorCmd_LineCountFollowsRegistry` unmodified: adding one identity
  source of two keys still costs exactly one `doctor` line. The task
  changes neither that test's doc comment nor its failure message.
- The leaf-mode refusal is asserted against a directory created at 0755
  before `Open` runs.
- The threshold bounds are asserted at and one past each boundary, and
  each bound literal appears in the test with its reason.
- `Store.Write`'s deletion is proved by grep returning no caller.
- The subscriber recover is asserted by a subscriber that panics, with
  the bump still succeeding and the record still on disk.
- Every expectation is written out rather than derived from the code
  under test. Doc comments carry no em dash. `bash scripts/check.sh`
  green.

**Named mutations this task reports:**

1. Unwrap one level in `rowStatuses` → the M4 row reports an all-green
   table and fails.
2. Drop the `Optional` consultation in `identitySourceStatus` → the
   short-`_PREVIOUS`-secret row fails.
3. Drop the refresh clamp → the boundary row's expiry lands past the cap.
4. Drop `MaxAge` → the rendered-header row fails.
5. Restore `MkdirAll`'s no-op on an existing 0755 leaf → the leaf-mode
   row fails.
6. Remove the per-kind label from `RecordID` → the two-kind collision row
   fails.
7. Drop the recover around a subscriber → the panicking-subscriber row
   takes the test process down.

**2.0 hook: neutral.**

**Prior art:** **adopts** OWASP's Session Management cheat sheet as the
argument for an explicit `Max-Age` (cited by URL and read date), and
stdlib `net/http`'s own cookie rendering as the assertion surface.
**Declines** any session library, on 5b-i's own recorded verdict: the
design is a stateless HMAC cookie plus a per-address epoch file, and
there is nothing for a library to manage.

**Independent: no.** It follows Task 1 in the same files.

**Deliverable count: 4** (the `doctor` flattening, the session cookie's
age, clamp and clearing helper, the store's permissions, id refusal, kind
label and dead-path removal, the threshold bounds with the public-URL
tightening).

---

## Task 3: The six remaining thresholds and their load-time validations

Model: `sonnet`. Depends on Tasks 0 and 2.

The thresholds `spec:1585-1599` lists that 5b-i deferred to the pass whose
code reads them, declared through the family 5b-i landed and validated
through the hook it landed.

**Files:**
- Modify: `cmd/dubplate-server/auth.go`,
  `cmd/dubplate-server/auth_test.go`
- Modify: `cmd/dubplate-server/doctor_test.go` (the line-count arithmetic
  alone, if Task 0 item 11 reports it needs it)

**Interfaces:**
- Consumes: 5b-i's `registerAuthConfig`, `magicLinkSource`,
  `validateMagicLinkSource`, and `resolveAuthConfig`.
- Produces: six registered thresholds, their validations inside the
  existing hook, and an extended resolver returning them. **Tasks 7, 10
  and 11 consume the resolved values.**

**Outcomes:**

- Six thresholds join `registerAuthConfig`, **fork 3's four remaining
  rate values at their ruled defaults** (`spec:216-229`): link requests
  per hour per address at 5, link requests per hour per client address at
  20, consumption attempts per hour per client address at 30, and the
  global outbound link-mail ceiling per hour at 60, each `KindInt`; **the
  auth request body cap** (`spec:1590`), `KindInt`, in bytes, with a
  default stated as a literal and its reason; and **the unknown-address
  answer** (R25, fork 4), `KindString`, default `kind`, accepted values
  `kind` and `generic`. **The arithmetic, stated so the task does not
  invent a name to reach a count.** `spec:1587-1593` declares ten
  thresholds, of which fork 3 sets the first seven. 5b-i landed four of
  the ten, `DUBPLATE_MAGIC_LINK_EXPIRY`, `DUBPLATE_SESSION_LIFETIME`,
  `DUBPLATE_SESSION_ABSOLUTE_CAP` and `DUBPLATE_NONCE_RETENTION`
  (`cmd/dubplate-server/auth.go:75-100`, registered at `:108-115`), three
  of them fork 3 durations. Ten declared minus four landed is six, and
  the six are the ones listed above. The task states its final list
  against `spec:1587-1593` and reports any name the spec implies that
  this list does not carry, rather than inventing one.
- **`DUBPLATE_SMTP_FROM` is validated at config load rather than at
  send** (`spec:1289`), inside `validateMagicLinkSource`: the value parses
  as a single address through `net/mail.ParseAddress` and carries no CR
  and no LF, reported by key name and never by value. The check is absent
  today, which Task 0 item 23 confirms by reading the function. It lands
  here because this task owns `cmd/dubplate-server/auth.go` this pass and
  Task 6's Files do not reach it; **Task 6 states the division and builds
  nothing there.** The operational reason is R9: an unparseable From makes
  every send fail and R9 turns every failed send into a failed login, so a
  typo in one environment row would otherwise produce a deployment that
  starts clean, passes `doctor`, and signs nobody in.
- **Each value is validated inside `validateMagicLinkSource`**, using the
  environment snapshot the hook already receives, so a bad value fails at
  load joined with every other failure rather than at first use. Every
  rate carries a positive floor and a stated ceiling; the body cap
  carries a floor and a ceiling; the unknown-address answer is checked
  against its accepted set with an error naming that set.
- **`resolveAuthConfig` returns them**, keeping the single resolution
  point the serve verb calls. Its signature grows; a struct return
  replaces the positional list if the count makes positional returns
  unreadable, which is the task's call to make and to report.
- **No key is declared in `internal/config`**, and no wall row, reserved
  class, or integration class is added.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Every default appears in the test as a written-out literal traced to
  `spec:216-229` for the five fork-3 values, and to a stated reason for
  the other two.
- A value at each floor and each ceiling, and one past each, is asserted.
- The unknown-address answer's refusal names the accepted set, and the
  expectation is written out rather than built from the registration.
- **The From validation is asserted at its boundary**: a value that
  `net/mail.ParseAddress` refuses, a value carrying CR, and a value
  carrying LF are each refused at load, joined with every other failure
  rather than raised at first use, and a well-formed value passes. No
  refusal message carries the value.
- `doctor`'s line-count assertions pass, and the task reports the before
  and after counts it measured rather than the count it predicted.
- No `DUBPLATE_` literal this task adds appears anywhere under
  `internal/config`, provable by grep.
- Each threshold's `Doc` states the accepted form a caller reports on a
  parse error, in the shape the four existing ones use.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Drop a floor → the zero-value row for that key passes when it should
   fail.
2. Widen the unknown-address accepted set → the refusal row fails.
3. Register a threshold under a name already claimed → `Registry.claim`
   panics at registration, which the test asserts as a programmer error
   rather than a load error.
4. Return the resolved values in the wrong positional order → Task 7's
   and Task 11's callers read a rate as a cap, caught by a resolver test
   that asserts each field against a distinct environment value.
5. Validate `DUBPLATE_SMTP_FROM` at send rather than at load → the
   load-time refusal rows pass a malformed value through to the seam.

**2.0 hook: neutral.** Thresholds carry no 2.0 surface.

**Prior art:** none adopted. `spec:216-229` is the authority for five of
the six values, and the task cites it rather than re-deriving.

**Independent: no.** It consumes 5b-i's declarations and feeds Tasks 7,
10 and 11.

**Deliverable count: 4** (the six registrations, the hook's added
threshold validations, the From validation, the extended resolver).

---

> **Conductor boundary, end of segment A.** STATUS. **Confirm the five reproductions were observed
> failing before their fixes**, which is the one thing no later task can
> recover. Re-derive nothing yet; the ceiling re-derivation is segment
> B's.

---

## Task 4: The roster store, its file format, and its single resolution point

Model: `sonnet`. Depends on Tasks 0 and 1.

The read half of fork 1's store. It is born whole here so that Task 5's
write path has a validated file to write and Task 9's seam has one place
to resolve an address.

> **Conductor-owed fold:** does R3's `doctor` roster collision row land
> here, widening this task's Files into `cmd/dubplate-server/doctor.go`
> and taking it to five deliverables against this plan's own
> four-deliverable cap, or route forward as a declared gap with the pass
> that owns it? Applied as this task's fifth deliverable. R3 requires the
> check "at startup, **in `doctor`**, and now at write time"
> (`spec:325-327`, repeated at `spec:1108-1110`), revision 1 built two of
> the three and routed the third nowhere, and `spec:2332`'s 5d row gives
> 5d "`doctor`'s probe rows and its bound-address row", which a roster
> validation row is not. This task already builds the loader and its
> validation, so `doctor` only calls it and renders the refusal.

**Files:**
- Create: `internal/auth/roster.go`, `internal/auth/roster_test.go`
- Create: `config/roster.example.json`
- Modify: `cmd/dubplate-server/doctor.go`,
  `cmd/dubplate-server/doctor_test.go` (**the R3 roster row alone**, and
  nothing else in either file; Task 2's M4 rendering fix has already
  landed there)

**Interfaces:**
- Consumes: `identity.NormalizeAddress` (Task 1's fixed point),
  `identity.Identity`.
- Produces: the member record; the loader interface R1 rules; **the
  store's two entry points**, a read-open that fails closed on an absent
  file and a write-open that creates one at 0600 and refuses an existing
  one; the lookup and the list; **the one resolution point where a login
  address becomes an identity value**; the `doctor` roster row. Tasks 5, 9
  and 10 consume it; **5c's `member add --admin` bootstrap calls the
  write-open** and 5c's add-member page calls the same interface.

**Outcomes (D6):**

- **One interface, and every caller speaks to it** (R1). One type opens
  the file, validates it, answers a lookup, lists the roster, and (Task
  5) adds a member. **No caller knows the format is JSON**, which is what
  makes rung 6's substitution a change of one constructor.
- **The file is a JSON object carrying a schema version and a members
  array** (R-5b19). A member record carries the address, the display
  name, the optional `streamer_username`, and the admin flag.
- **The loader fails closed** (R2). A missing, unreadable, or unparseable
  file is a load error naming the path. The salvage loader's
  empty-roster-and-nil-error is superseded, and the task's doc comment
  says which behavior it replaces and why: a fail-open roster lets a
  deployment start and refuse every login as unknown, which reads to a
  member as a broken login rather than a broken configuration, and under
  fork 4's kind answer it then hands every visitor the admin contact.
- **The write-open is the other entry point, and it is the bootstrap's**
  (R-5b20, R37 at `spec:749-752`). It creates the file at 0600 when it is
  absent and **refuses an existing one**, which is what
  `dubplate-server member add --admin` needs on a fresh install and what
  keeps a second roster from being created over a first. It creates no
  directory. The read-open keeps refusing an absent file after the
  write-open exists; the two entry points are not two modes of one
  constructor that can be softened into each other.
- **Validation is B9's list** (`spec:1105-1109`): addresses parsed and
  normalized through `identity.NormalizeAddress`, unique after
  normalization, no control characters anywhere in the record, every
  field bounded in length with the bound written out as a literal with
  its reason, and `streamer_username` validated as a stable principal,
  ASCII-printable, unique across the roster, and bounded (R3).
- **Email matching is case-insensitive on the normalized address**, and
  normalization happens before any keying, any logging, and any
  rate-limit lookup (`spec:1161-1163`).
- **The store exposes exactly one point where a login address becomes an
  identity value**, and its doc comment names that point as the place rung
  6 substitutes a durable surrogate id (`S20:367-373`).
- **The admin flag is the whole of rung 5's authorization vocabulary**
  (`spec:1111-1115`), and the record's doc comment says so, so a later
  rung reaching for a role model knows it is adding one rather than
  extending one.
- **`doctor` carries the R3 roster row.** It calls this task's validation
  and renders the refusal, so a duplicate principal, a duplicate address,
  or an unparseable record is reported on the operator's terminal as well
  as refused at startup and at write time (`spec:325-327`,
  `spec:1108-1110`). The row names **no address and no principal value**,
  on `identitySourceStatus`'s own precedent
  (`cmd/dubplate-server/doctor.go:161-166`, "It names no value, because a
  source's required set carries secrets and doctor's table is read on a
  terminal"). It is one row, so R10's line-count arithmetic is unchanged.
- `config/roster.example.json` ships as the adopter's starting point,
  with invented addresses, and the runbook (5d's) shows the CLI verb
  rather than the file. **`deploy.sh` does not deliver it and never
  will** (`spec:751-752`, "never as a file `deploy.sh` copies"), and the
  task's own report says so, because the file lands beside
  `config/contributors.yaml`, which `deploy.sh:426` does deliver and which
  stays until **rung 16** retires the incumbent importer
  (`spec:752-754`).

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Four load refusals are asserted, each by its own sentinel and each
  naming the path: missing file, unreadable file, unparseable JSON, and a
  record failing validation.
- Validation rows cover, each with a written-out expectation: a duplicate
  address differing only in case; a duplicate `streamer_username`; a
  control character in a display name; a control character in an address;
  an unparseable address; each length bound at the boundary and one over;
  a `streamer_username` carrying a non-ASCII-printable byte.
- The resolution point is one function, provable by grep: no other
  function in the package converts an address to an identity value.
- A test asserts that a lookup for an address differing only in case
  resolves to the same member.
- **The two entry points are asserted against each other**: the read-open
  refuses an absent file; the write-open creates one at 0600 and the mode
  is asserted; the write-open refuses a file that already exists, by its
  own sentinel.
- **The `doctor` roster row is asserted twice**: a roster with a duplicate
  principal renders the refusal and `doctor` exits nonzero, and the
  rendered row contains neither an address nor a principal value, checked
  by scanning the captured output for the fixture's own literals.
- `config/roster.example.json` parses through the loader in a test, so
  the shipped example cannot rot.
- **No address from the example file or any fixture appears in a log
  record or an error string**, outside the roster audit record Task 5
  writes, which is the one exception Global Constraints names and which
  this task does not write.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Return an empty roster and a nil error for a missing file → the
   fail-closed row fails, which is the salvage behavior R2 supersedes.
2. Compare addresses without normalizing → the case-differing duplicate
   row passes when it should fail.
3. Drop the `streamer_username` uniqueness check → the duplicate-principal
   row fails.
4. Compare a length bound against the rune count rather than the byte
   length → the multi-byte boundary row fails.
5. Add a second address-to-identity derivation → the grep criterion
   fails, and hook 5's substitution stops being one function.
6. Let the write-open succeed against an existing file → the
   refuse-a-second-roster row fails.
7. Let the read-open tolerate an absent file once the write-open exists →
   the fail-closed row fails, which is the resolution under time pressure
   this task's two entry points exist to make unnecessary.
8. Drop the `doctor` roster row → the duplicate-principal row reports a
   clean table for a roster that refuses to load.

**2.0 hook: lands hook 5's read half** (`spec:1848`, `S20:367-373`). The
single resolution point is where rung 6 substitutes a surrogate id, and
`streamer_username` is validated as required-shaped now (R3) so that a
roster written today is not a migration rung 6 pays for.

**Prior art:** **adopts** stdlib `encoding/json` (R1's own reasoning: a
program writes this file, and reading YAML would take a module dependency
R26 does not budget). **Examines and declines**
`_salvage/internal/config/contributors.go` as code: its fail-open branch
is the behavior R2 supersedes, and its per-request membership re-check
(`:93`) is adopted as the **shape** Task 10's session path follows, which
is R30's own citation.

**Independent: yes**, by files, against Tasks 2 and 3. It consumes Task
1's normalizer.

**Deliverable count: 5** (the member record and file format, the
fail-closed loader with B9's validation, the interface with its two entry
points and its single resolution point, the `doctor` roster row, the
example file). **This is one above the cap this plan otherwise holds**,
and the fifth is the conductor-owed fold at the head of this task.

---

## Task 5: The roster's write path, its collision re-check, and its audit record

Model: `sonnet`. Depends on Tasks 0 and 4.

The half fork 1 reversed the recommendation for. Its callers, the
`member` verb and the add-member page, are both 5c's; this task ships the
path they call.

**Files:**
- Modify: `internal/auth/roster.go`, `internal/auth/roster_test.go`

**Interfaces:**
- Consumes: Task 4's store, validation, and resolution point;
  `internal/atomicfile.WriteStream`; `identity.Channel`.
- Produces: the add method, its audit record, and the re-read that keeps
  a running server from serving a roster it did not write. **5c's two
  callers consume it; no 5b-ii task calls it** except the tests.

**Outcomes (D7):**

- **Every write validates the whole record first, takes the store's lock,
  re-checks the address and the principal against the roster as loaded,
  writes atomically through `internal/atomicfile.WriteStream`, and
  re-reads** (`spec:1129-1135`). The re-check runs **inside the same lock
  the write holds**, so two simultaneous adds cannot both pass it (R3).
- **A write that fails validation changes no byte on disk.** The test
  proves it by comparing the file's bytes before and after, not by
  asserting an error.
- **`WriteStream` rather than `WriteFile`** (R11, `spec:439-444`):
  overwriting is what adding a member does, and only `WriteStream` fsyncs
  the containing directory after the rename, so a rename that never
  reached the directory cannot drop a member an admin watched the page
  accept.
- **A write whose re-read fails to parse is an error the caller sees**
  rather than a roster the running server has already accepted (R2's own
  clause).
- **The store creates the file when absent and never its directory**
  (R-5b20). The `ReadWritePaths` grant and the `deploy.sh` pre-create are
  5d's.
- **Every roster change writes one audit record** naming the actor, the
  assertion channel, the address added, whether the admin flag was set,
  and the reported client IP (`spec:1143-1149`). It is a structured
  `slog` record (R-5b21), and **the record is the only trace a roster
  change leaves**, because the file carries no history.
- **The web never grants the admin flag to a first login** (fork 1), and
  **this task enforces it rather than documenting it.** The write path
  takes the flag as an explicit argument, the store neither infers nor
  defaults it, and **the store refuses a write that sets the flag on any
  assertion channel but the command-line actor**, with its own sentinel.
  The channel is already in the store's hands, because the audit record
  takes it as an argument.

  > **Conductor-owed fold:** is the admin-flag invariant enforced in the
  > store's write path, or left to 5c's caller with a gap recorded here?
  > Applied as enforcement in the write path.

  The spec states the invariant three times as a property the design holds
  (`spec:1949`, "no web path sets it on a first login"; fork 1 at
  `spec:191-192`, "the only path to the first admin is the CLI verb";
  `spec:1236`, "granting the admin flag to a first login stays
  impossible"). This pass ships the roster's only write path, so a write
  path that accepts `admin: true` from the cookie channel is one handler
  bug away from a privilege grant with nothing but a doc comment in
  another package in front of it. Rung 5 has no promote path and the
  add-member page is rung 5's only web writer, so the refusal costs
  nothing a caller needs. **5c's add-member handler therefore inherits a
  refusal rather than a rule to remember**, which the 5c handoff states.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The byte-identity assertion is a real comparison of file contents
  before and after a failed write, over at least three distinct
  validation failures.
- A concurrent pair of adds, run with `-race`, both survive, and the
  final file carries both members.
- A concurrent pair of adds claiming the same `streamer_username` leaves
  exactly one, with the other refused by its sentinel.
- The re-read is asserted by a lookup immediately after a write, against
  the same store value, with no reopen.
- The audit record is asserted by capturing the logger's output and
  checking every named attribute is present, **the address included**,
  with no flag missing and no token, secret or cookie value present. This
  record is the one place in the pass an address reaches a log, per
  `spec:1139-1141` and the exception Global Constraints names.
- **A cookie-channel add carrying `admin: true` is refused by its
  sentinel and changes no byte**, asserted the same way the validation
  failures are; a command-line-channel add carrying the same flag
  succeeds.
- The file's mode is asserted as 0600 after a create, **through Task 4's
  write-open**, and the read-open still refuses a missing file after the
  write-open exists.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Write before validating → the byte-identity row fails.
2. Take the lock after the re-check → the concurrent-principal row admits
   both.
3. Use `WriteFile` rather than `WriteStream` → the durability argument's
   doc comment no longer matches the call, caught by the reviewer rather
   than by a test, and recorded as such.
4. Skip the re-read → the immediate-lookup row fails.
5. Drop the channel from the audit record → the attribute row fails.
6. Accept the admin flag on any assertion channel → the cookie-channel
   refusal row admits the write.
7. Let the read-open tolerate an absent file → the fail-closed row fails
   after the write-open exists, which is Task 4's property under the
   pressure this task creates.

**2.0 hook: lands hook 5's write half** (`spec:1848`, `S20:367-373`). A
member arrives through the store's own write path, so that is where rung
6 mints a surrogate id rather than deriving one, and the
`streamer_username` collision check runs there rather than only at
startup.

**Prior art:** **adopts `internal/atomicfile.WriteStream`** (in-repo,
rung 4's, re-landed from salvage on merit; the directory fsync at
`atomicfile.go:82-89` is the property R11 turns on). No external prior
art consulted: a locked validate-write-reread is fifteen lines and no
library's version of it would be shorter or clearer.

**Independent: no.** It consumes Task 4.

**Deliverable count: 4** (the locked write path with its re-read, the
write-time collision re-check, the audit record, the admin-flag channel
refusal).

---

## Task 6: `internal/notify`, the general send seam

Model: `sonnet`. Depends on Task 0.

The mail seam, born general rather than as a magic-link mailer, because
rung 7's reject copy and rung 9's notifications are its other callers and
`internal/reject`'s wording is a contract this package must not duplicate.

**Files:**
- Create: `internal/notify/notify.go`, `internal/notify/notify_test.go`
- Modify: `go.mod`, `go.sum` (the first of the two budgeted dependencies)

**Interfaces:**
- Consumes: nothing from this pass. **It reads no environment**: its five
  SMTP values arrive as a struct the caller fills.
- Produces: one send method, `ctx` first, taking a recipient, a subject,
  and a body. **Task 10 is its first caller**; rung 7 and rung 9 are the
  callers it is general for.

**Outcomes (D8):**

- **One method, `ctx` first, one SMTP transport** (R7). The seam carries
  a subject and a body as data and **owns no copy of its own**, because
  rung 7's reject wording is a code contract in `internal/reject` shared
  verbatim with the uploader page.
- **The seam owns the header-injection guard** (SEC:94,
  `spec:1288-1294`), and the guard covers **every header-valued
  argument**, the subject as well as the recipient, because rung 7 and
  rung 9 will pass subjects derived from filenames and tag text. A
  recipient or a subject carrying CR or LF is refused by the seam with
  its own sentinel, before anything is dialed.
- **The From value is validated at config load rather than at send**
  (`spec:1289`). **Task 3 lands that check** inside
  `validateMagicLinkSource`, which owns `cmd/dubplate-server/auth.go` this
  pass; **this task states the division and builds nothing there.** The
  seam still guards the From header it composes against CR and LF along
  with every other header-valued argument, because a guard at the seam is
  the seam's own invariant and not a substitute for the load-time refusal.
- **`github.com/wneessen/go-mail` is the transport** (R8), at the version
  Task 0 item 13 reports, with the LICENSE's MIT terms recorded in the
  task's report. Implicit TLS on port 465 and mandatory STARTTLS on any
  other port, which is the salvage package's own measured contract and
  the box's actual route.
- **Error doctrine is R9**: the seam reports failure to its caller and
  makes no policy. Best-effort is every other caller's choice; the login
  handler is the one caller that treats a send failure as a failed login,
  and that decision lives in Task 10 rather than here.
- **The success record names no recipient.** The salvage package "logs
  one record on success" (`spec:1296-1300`) and the natural attribute of
  that record is the address it went to, which Global Constraints forbid
  outside the roster audit record. **The seam's success record names the
  outcome and nothing that identifies a person**, and its doc comment says
  why, so a later caller adding a recipient attribute is adding one
  against a stated rule rather than filling an obvious gap.
- **Nothing Python, nothing Node, nothing site-specific.** No default
  host ships, because a shipped default host would be a site-specific
  value in the wrong place (`spec:1305-1306`).
- The salvage package's two test-only fields each keep a comment saying
  so, which is the shape `go-conventions` asks for rather than an
  injected interface.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- A recipient carrying CR, a recipient carrying LF, a subject carrying
  CR, and a subject carrying LF are each refused by sentinel, before any
  dial, proved by a test whose transport would fail loudly if reached.
- The implicit-TLS and STARTTLS branches are both exercised against a
  local test server, the implicit branch through the test-only field
  rather than by binding a privileged port.
- A context already cancelled returns without dialing.
- The subject's RFC 2047 encoding and the body's quoted-printable
  encoding are asserted against written-out expectations for a non-ASCII
  subject and body, cited to the RFC by number.
- `go.mod` carries exactly one new direct dependency and `go mod tidy` is
  clean in the same commit.
- The package imports no `internal/config` type and reads no environment
  variable, provable by reading the import block.
- **The success record is asserted by capturing the logger's output and
  scanning it for the recipient's own literal**, which must not appear.
- **No From validation lands in this package**, provable by grep: the
  load-time check is Task 3's, in `cmd/dubplate-server/auth.go`.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Guard the recipient and not the subject → the two subject rows fail.
2. Guard only LF → the two CR rows fail.
3. Guard after composing the message → the "before any dial" row fails,
   because the transport is reached.
4. Route port 465 through STARTTLS → the implicit-TLS row fails.

**2.0 hook: neutral.** The seam's generality is what makes rung 7 and
rung 9 callers; 2.0 adds no mail surface this pass must anticipate.

**Prior art:** **adopts `_salvage/internal/notify`** as the starting
point, restructured so the subject and body are data rather than a
magic-link template. **Judged on CLAUDE.md's four criteria, with the
evidence rather than the provenance**, because revision 1 recorded
provenance plus a deferral and a deferral is not a criterion a
diff-reviewer can grade. **License and adoption are degenerate for
in-repo code**: the file is this repo's own under this repo's terms, and
its importer count is whatever this repo makes it, so neither criterion
discriminates. **Quality**, recorded rather than promised: 96 lines,
exported surface `Mailer`, `New`, `Send`, `implicitTLSPort = 465` at
`_salvage/internal/notify/notify.go:24`, two test-only fields each
carrying a comment saying so, and `notify_test.go` beside it, all
confirmed by Task 0 item 14 against the tree before this task starts.
**Maintenance**: it ran in production on the box's reject-mail path and
passed the repo gate, and it has no upstream to track, so the maintenance
question is this repo's own. **The fresh review at the exemplary-Go bar
is a criterion this task is graded on**, not a promise: the task reports
what it changed from the salvage shape and why, line by line.
**Adopts `github.com/wneessen/go-mail`** (R8; MIT, LICENSE
fetched 2026-09-08, v0.8.1 published 2026-07-09, 237 importers, implicit
TLS and STARTTLS policies, TLS 1.2 minimum; adoption is the weak cell and
the plan states it honestly). **Declines** stdlib `net/smtp` as the whole
answer (frozen by its own documentation, no implicit-TLS dial, no message
composition, so choosing it means hand-rolling RFC 2047, quoted-printable,
Message-ID and Date). **Declines** `emersion/go-smtp` (its differentiator
is the server half). **Declines** JMAP from Go (v0.5.3, nine importers;
poplar hand-rolled its own). **Declines** a relay Worker of our own (a
second deployable, a shared secret, and a new key class for no gain).
**Adopts Cloudflare Email Service's SMTP submission as documentation, not
code** (it reaches the same five reserved SMTP keys; the runbook row is
5d's).

**Independent: yes**, by files and by imports, against Tasks 4 and 5.

**Deliverable count: 3** (the seam and its transport, the header guard,
the tests with the dependency addition).

---

> **Conductor boundary, end of segment B.** STATUS. **Re-derive the ceiling from Tasks 4 and 5's
> actual reports**, which is the measurement this plan predicted rather
> than measured. If the re-derivation projects past 8.0M, name the cut at the
> Task 8 / Task 9 boundary.

---

## Task 7: The three rate limiters and the bounded evicting map

Model: **`opus`**. Depends on Tasks 0, 1 and 3.

The memory-exhaustion bug B8 names is an unbounded map of limiters keyed
by attacker-supplied addresses, and the map is rung 5's own code by R16's
ruling. This is the task where a wrong bound is invisible until it is a
box out of memory.

**Files:**
- Create: `internal/auth/limit.go`, `internal/auth/limit_test.go`
- Modify: `go.mod`, `go.sum` (the second of the two budgeted
  dependencies)

**Interfaces:**
- Consumes: Task 3's resolved rates and ceiling; Task 1's normalizer;
  `internal/serve.ClientIP` at the call site, never re-derived here.
- Produces: one bounded evicting keyed map type; three limiters
  instantiated over it; the global outbound-mail ceiling. **Task 10 is
  the only caller.**

**Outcomes (D9):**

- **Buckets come from `golang.org/x/time/rate`** (R16, `spec:1171-1173`).
  **The keyed map, its size bound, and its eviction are this package's
  own code.**
- **Three limiters, three key spaces, one map implementation**
  (R-5b24): link requests per submitted address, link requests per client
  address, and consumption attempts per client address. **Either of the
  first two tripping trips the link request** (R28).
- **A global ceiling bounds outbound link mail per interval** besides
  (R28), so an enumerator walking distinct addresses cannot drive real
  mail until the provider's quota answers instead of the server.
- **The map's bound holds under a flood of distinct keys**, proved by a
  gate-resident test (R16's own requirement). Eviction is
  least-recently-used, and the doc comment states why a bucket for a key
  nobody has used recently is the right one to lose.
- **Keys derive from the normalized address or from the reported client
  IP**, never from a raw token or cookie string (security constraint 3),
  and never from a header the peer address does not authorize (R27). The
  limiter takes the client address as a value; **this package does not
  derive one**.
- **A tripped limit is an error with its own sentinel**, and the caller
  logs it with the reported client IP (`spec:1179-1181`). The limiter
  itself logs nothing, so one policy about what an auth event records
  lives in one place.
- The bound and the eviction policy are package constants with their
  reasons stated (R-5b23), not thresholds.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The flood test drives strictly more distinct keys than the bound and
  asserts the map's length never exceeds it, sampled during the flood and
  not only at the end.
- A test asserts that evicting a key does not reset an unrelated key's
  bucket, run across two key spaces, which is R-5b24's own property.
- The per-address and per-client link limits are each asserted to trip
  independently: N requests for one address from many clients, and N
  requests for many addresses from one client.
- The global ceiling is asserted to stop mail after its count within the
  interval regardless of which keys were used.
- Every rate in the test comes from Task 3's registered defaults by name,
  not from a literal duplicated here, and the test states where the
  expectation came from.
- The limiter's own tests run with `-race`, and the map is exercised
  concurrently.
- No test sleeps for real time: the limiter takes its clock as a value or
  the buckets are driven through `rate.Limiter`'s own
  `AllowN(t, n)` form, so a sub-second run is not the evidence.
- `go.mod` carries exactly one new direct dependency and `go mod tidy` is
  clean in the same commit.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Remove the bound → the flood row fails on the sampled length.
2. Share one map across the three key spaces → the cross-space eviction
   row fails.
3. Evict most-recently-used rather than least → the eviction row's live
   bucket disappears.
4. Key the per-client limiter on the peer address where a trusted proxy
   is configured, or on the header where it is not → **a call-site
   defect, so it is Task 10's row rather than this package's**. This
   package takes the client address as a value and derives none, so no
   test here can reach it; the task states that plainly and Task 10
   carries the forged-header criterion and this mutation.
5. Let either link key alone gate the request → the independence rows
   fail one each.

**2.0 hook: neutral.** The limiter carries no 2.0 surface, and 2.0's
`/rest/*` traffic is the streamer's own authentication rather than this
limiter's.

**Prior art:** **adopts `golang.org/x/time/rate`** (BSD-3, Go team, the
house exemplar; v0.16.0 published 2026-08-19, 14,348 importers, agreeing
with S20's existing verdict). **The bounded evicting keyed map is ours**
by R16's explicit ruling. **Declines** an LRU module (`hashicorp/golang-lru`
and its siblings) on two criteria: R26 budgets two dependencies for rung 5
and both are spent, and a `container/list`-backed LRU over a map is under
fifty lines whose bound is the property a test must prove anyway.

**Independent: yes**, by files, against Tasks 8 and 9. It consumes Tasks
1 and 3.

**Deliverable count: 3** (the bounded evicting map, the three limiters
with their key spaces, the global ceiling).

---

## Task 8: The origin check and its bearer bypass

Model: **`opus`**. Depends on Tasks 0 and 7.

> **Conductor-owed fold:** does Task 8 run at `opus` unconditionally, get
> its own segment boundary so a conditional upshift is observable, or run
> plainly at `sonnet`? Applied as `opus` unconditionally. Revision 1 read
> "upshift to `opus` if Task 7's review returns `fix`", which cannot fire:
> Tasks 7 and 8 share segment C, a `pass-execute` launch fixes each task's
> model in its args at dispatch, and the conductor sees nothing until the
> launch returns
> (`docs/superpowers/plans/2026-09-12-overnight-window.md:134-139`). The
> unconditional upshift is taken over the extra boundary because this
> task's own argument for it does not depend on Task 7's verdict: its
> correctness rests on a placement argument rather than on code volume,
> and its blast radius is the whole served surface.

One rule, stated once at the shell, covering every cookie-authenticated
state-changing surface. SEC:36 is the defect this must not repeat: the
salvage check sat inside the `/upload` mount and covered no new surface.

**Files:**
- Create: `internal/serve/origin.go`, `internal/serve/origin_test.go`
- Modify: `internal/serve/serve.go` (the composition in `Handler` alone),
  `internal/serve/serve_test.go`

**Interfaces:**
- Consumes: nothing from this pass. It reads no auth type, so
  `internal/serve` gains no import of `internal/auth`.
- Produces: the wrapped protection, its insertion into the composition,
  and **one registration method that adds a bypass pattern to both the
  wrapper's own set and the stdlib value's**. **Task 11 registers no
  pattern**, because the fallthrough is empty at rung 5.

**Outcomes (D10):**

- **`net/http.CrossOriginProtection` is the mechanism** (R15), **wrapped
  so that a protected request carrying neither `Origin` nor
  `Sec-Fetch-Site` is denied** rather than allowed, which is the one
  documented respect in which the stdlib default is too permissive for
  this rule.
- **The check inserts after the ingress strip and after path
  normalization, and before prefix routing** (R-5b27,
  `internal/serve/serve.go:150-163`), so every registered subtree and the
  fallthrough are covered by construction rather than by each mount
  remembering.
- **A request authenticated by a bearer credential rather than by the
  session cookie is exempt, through the wrapper's own bypass set**, and
  **a bypassed path may not read the session cookie for authentication**,
  which is what keeps the exemption from becoming a hole. The rule is
  stated in the file's doc comment as a contract rung 12 and rung 14
  inherit rather than a wrapper they rewrite.
- **The wrapper owns its bypass set, because `AddInsecureBypassPattern`
  alone cannot carry the exemption past the tightening.** Measured on
  go1.27.1: `Check` returns nil for the neither-header case at
  `src/net/http/csrf.go:155-158` **before** any bypass is consulted, and
  `isRequestExempt` (`csrf.go:185-198`) is unexported over a private
  `c.bypass` field, so a wrapper cannot ask the stdlib value what matched.
  A wrapper that denies whenever `Check` returns nil and both headers are
  absent would therefore deny a bypassed path's POST as well, and the
  contradiction would ship unexercised into rung 12's lap, because no test
  in this pass can reach a bypassed path. **The wrapper keeps its own
  `http.ServeMux` with a sentinel handler, in the shape `csrf.go:185-190`
  uses, consulted before the neither-header denial**, and its single
  registration method also calls `AddInsecureBypassPattern` so the
  stdlib's own cross-origin branch honors the same pattern. **The task
  re-measures both facts against the tree's toolchain and reports them**
  rather than taking this plan's word.
- **`CrossOriginProtection` always allows GET, HEAD and OPTIONS**, so a
  native client's state-changing GET is never checked and the streamer's
  own authentication stands behind it. The doc comment states this
  consequence rather than leaving rung 12 to find it.
- **The fallthrough is empty at rung 5**, and until rung 12 fills it the
  rule denies a cookie-carrying state-changing request there with no
  bypass pattern registered. The test asserts exactly that.
- The salvage implementation's ordering discovery is carried:
  `Sec-Fetch-Site` is honored on its own, with `Origin` against `Host` as
  the fallback (`spec:1206-1210`), because an `Origin`-only check refuses
  every development request through a proxy that rewrites `Host`. **Whether
  the stdlib's own `Check` already implements that order is measured and
  reported**, and the wrapper adds only what it does not.
- **Double-submit tokens are declined as redundant** (`spec:1212-1215`),
  and the file says so, because the next reader's first question is why
  there is no token.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- A protected POST carrying neither header is denied; the same POST
  carrying a same-origin `Origin` is allowed; the same POST carrying
  `Sec-Fetch-Site: same-origin` and no `Origin` is allowed; the same POST
  carrying a cross-origin `Origin` is denied. Four rows, each with a
  written-out expectation.
- A GET carrying neither header is allowed, which is the stdlib's
  documented behavior and the reason the doc comment states the
  consequence.
- **The coverage test runs against the shell as the serve verb composes
  it**, and asserts that a registered subtree and the fallthrough are both
  covered, so the property is placement rather than a wrapper's own
  behavior.
- **A pattern registered through the wrapper's own method is exempt from
  the tightening**: its POST without either header is allowed, while the
  same POST to an unregistered path is denied. Both rows run against the
  wrapper, which is the surface that can be tested; the companion rule,
  that a bypassed path may not read the session cookie, is stated in the
  test's own comment and remains a contract a reviewer reads.
- **The registration reaches the stdlib value too**, asserted by a
  registered pattern's **cross-origin** POST being allowed, which is the
  branch `AddInsecureBypassPattern` governs and the wrapper's own set does
  not.
- The insertion point is asserted by behavior, not by reading the
  composition: a request with a `..` segment is refused by normalization
  before the origin check sees it, proved by a check that records whether
  it ran.
- `internal/serve` gains no import of `internal/auth`.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Drop the tightening and use the stdlib default → the neither-header
   row is allowed and fails.
2. Move the check inside the prefix's routed handler → the fallthrough
   coverage row fails.
3. Move the check outside `Normalize` → the `..` ordering row fails.
4. Let a bypassed path read the session cookie → caught by the reviewer
   against the stated contract, and recorded as a criterion no test in
   this pass can reach, because no bearer-authenticated path exists yet.
5. Consult the wrapper's bypass set after the neither-header denial rather
   than before → the registered pattern's neither-header row is denied.
6. Register through `AddInsecureBypassPattern` alone and drop the
   wrapper's own set → the same row fails, which is the defect revision 1
   would have shipped.

**2.0 hook: neutral for hook 6, with a recorded forward.** The bypass
registration point is what 2.0's OpenSubsonic surface under `/rest/*`
consumes (`S20:284-285`), and rung 14's non-browser token acquisition
inherits the same shell rule. Nothing is built early; the affordance is
named so rung 12 does not add a second mechanism.

**Prior art:** **adopts stdlib `net/http.CrossOriginProtection`**
(implements the birth requirement's rule; Go 1.25, module on `go 1.27`),
wrapped per R15. **Declines `gorilla/csrf`** (BSD-3, v1.7.3 published
2025-01-23, 1,103 importers) on security posture: it carries advisory
GO-2025-3884 against its trusted-origin handling, which is the exact
dependency this rung should not take when the stdlib implements the same
check. **Declines double-submit tokens** as redundant on a browser-only
surface where the `__Host-` prefix, `SameSite=Lax`, and the origin check
each do a distinct job. **Examines
`_salvage/internal/serve/serve.go`'s `requireSameOrigin`** and adopts its
ordering discovery as a lesson, not its code.

**Independent: yes**, by files, against Tasks 7 and 9.

**Deliverable count: 3** (the wrapped check, its insertion into the
composition, the wrapper-owned bypass set with its one registration
method and its contract).

---

> **Conductor boundary, end of segment C.** `code-simplifier` over Tasks
> 1 through 8, its review chunked one `diff-reviewer` per package. STATUS. **This is the last boundary before the handlers**, so
> confirm that Task 1's `Insert` signature, Task 4's resolution point,
> Task 6's seam, Task 7's limiters, and Task 8's placement are all
> accepted before segment D1 dispatches: Task 10 consumes every one of
> them.

---

## Task 9: The verify-identity seam

Model: `sonnet`. Depends on Tasks 0, 1 and 4.

The seam exists as an interface from this rung, with magic-link as its
reference implementation and forward-auth fronts as alternates. Its point
is provenance.

**Files:**
- Create: `internal/auth/seam.go`, `internal/auth/seam_test.go`

**Interfaces:**
- Consumes: Task 4's roster and its resolution point; 5b-i's
  `VerifySession`, `RefreshSession`, `EpochStore.CheckSession` and
  `EpochStore.Read`; `identity.Identity` and `identity.Channel`.
- Produces: the seam interface and its magic-link implementation.
  **Task 10 consumes it; rung 6's ops request types carry its value; rung
  12 consumes it and never reaches into `internal/auth`.**

**Outcomes (D11):**

- **The seam is an interface from this rung** (`spec:1219-1221`), with
  one implementation. It yields an `identity.Identity` carrying the login
  address, the resolved principal, the display name, the identity source
  that asserted it, and **the assertion channel, which is cookie, header,
  or service token**.
- **Provenance is the point** (`spec:1225-1229`). The organizer's rule
  that admin verbs reject cookie-derived identity is enforceable at the
  seam only if the channel travels with the identity, and rung 6's ops
  request types carry the same value.
- **The admin flag is read from the roster at the seam**, never from
  anything the request carries (`spec:1233-1239`). Fork 1's carve-out is
  stated in the doc comment: the add-member handler (5c's) accepts a
  cookie-derived identity whose roster record carries the flag, and
  nothing else does.
- **The seam exposes the epoch's `Read`** as the authority a later
  credential store stamps from (R35, `S20:245-246`), **without assuming
  the session store is its only reader** (the configuration-homes threat
  model's rung 5 item 1). Being told the epoch changed is not enough; a
  credential store must be able to read it.
- **Every non-nil error denies** (security constraint 2). The seam's
  contract sentence says so on the method itself: nil means permitted,
  anything else means refused, and the sentinel distinction chooses the
  server-side log line alone.
- **The session path re-checks roster membership on every request**
  (R30's last clause, `_salvage/internal/config/contributors.go:93` as
  the shape), so a removed member's live session stops working at the
  next request rather than at its expiry.
- **The refresh counts same-origin requests only** (`spec:1081-1082`,
  plan review M-e). The seam takes the same-origin condition as an input
  rather than deriving it, because the header set is the shell's
  knowledge; **Task 10's call site supplies it**.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- A valid session for a member on the roster yields an identity whose
  channel is cookie and whose admin flag matches the roster record.
- A valid session for an address removed from the roster is refused, with
  no identity returned.
- A session refused by the epoch is refused here, and the test performs a
  refresh before the check, which is the anchor property 5b-i landed.
- A non-sentinel error from any dependency is refused, proved with a
  roster whose file has been made unreadable mid-test, which is the M1
  and `EpochStore.Read` class security constraint 2 exists for.
- The admin flag's provenance is asserted: a request whose cookie value
  encodes an admin claim does not produce an admin identity.
- `Read` is exercised without any subscriber registered, proving it is the
  authority rather than the notification's companion.
- A cross-origin request does not refresh, proved by comparing the
  session's expiry before and after.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Read the admin flag from the session value → the provenance row fails.
2. Skip the per-request roster re-check → the removed-member row fails.
3. Treat a non-sentinel error as "not refused" → the unreadable-roster row
   fails, which is the fail-open M1 predicts.
4. Refresh unconditionally → the cross-origin row fails.

**2.0 hook: lands hook 5's seam half** (`spec:1848`, `S20:367-373`). The
resolved principal is the field rung 6 substitutes a surrogate id into,
and the epoch's `Read` is what 2.0's credential store stamps an app
password from (`S20:245-246`).

**Prior art:** none adopted for the seam itself. **Declines**
`golang-jwt/jwt/v5`'s claims vocabulary as a model on 5b-i's own recorded
reasoning: the claim set is designed for a token crossing a trust
boundary, and this value never leaves the process. Gitea's server-side
session record stays the road not taken.

**Independent: yes**, by files, against Tasks 7 and 8. It consumes Tasks
1 and 4.

**Deliverable count: 3** (the seam interface, its magic-link
implementation with the roster and epoch re-checks, the channel and
admin-flag provenance).

---

## Task 10: The three auth handlers

Model: **`opus`**. Depends on Tasks 0, 1, 2, 3, 4, 6, 7 and 9. **Task 2 is
in the list** because the sign-out handler calls Task 2's cookie-clearing
helper and the session path calls Task 2's clamped `RefreshSession`.

The pass's convergence. Three handlers, one fixed order, and the whole of
the security review's three constraints applied at once.

**Files:**
- Create: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`

**Interfaces:**
- Consumes: everything this pass built, plus 5b-i's token, session, nonce
  and epoch surfaces, plus `internal/serve.ClientIP` supplied as a value
  by the caller.
- Produces: a constructor taking one dependency struct and returning the
  three handlers as `http.Handler` values (R-5b25). **Task 11 mounts
  them.**

**Outcomes (D12):**

- **The request-link handler**, in this order: cap the body with
  `http.MaxBytesReader` at Task 3's threshold; parse; normalize the
  submitted address; check the per-address limiter and the per-client
  limiter, either tripping; resolve the address against the roster; check
  the global send ceiling; mint the link; build its URL from the
  configured public origin; send through the mail seam. **Under `kind`, a
  send failure is a failed login** (R9). **Fork 4's kind answer** returns
  the roster's admin contact as data when the address is unknown and the
  unknown-address threshold is `kind` (R25, R-5b22). **No copy is written
  here.**
- **Under `generic`, a known address and an unknown address produce
  byte-identical responses, and the response does not wait on the send**
  (R-5b28). A send failure under `generic` is logged rather than rendered.
  The generic outcome has exactly one purpose, to stop a stranger learning
  whether an address is on the roster, and revision 1 left both its
  channels open: the response itself, because R9 applied literally makes a
  known address whose send fails answer differently from an unknown one,
  and the timing, because a known address runs a mint, a URL construction
  and an SMTP dial before answering while an unknown one answers at once,
  which is a network round trip to a mail provider rather than a subtle
  side channel.
- **The consumption handler**, in this order and no other: **the
  per-client consumption limiter, then MAC, expiry, roster resolution,
  `CheckLink`, `Insert`, and only then `MintSession`** (security
  constraint 1, extended at the front by revision 2). **The limiter runs
  first**, keyed on the reported client IP the caller supplies. Placement
  is the whole of this limiter's value: a token brute-forcer fails at the
  MAC on every attempt, so a limiter anywhere after the MAC is never
  reached by the flood it exists to bound, and `spec:1178-1179` gives
  consumption attempts their per-client key for exactly that traffic. A
  limiter after `Insert` is worse than useless, because it burns a
  legitimate member's nonce and then refuses the request, turning a rate
  limit into a single-use-link destroyer. The session is minted only after
  `Insert` returns nil. The body is capped before the parser sees it. The
  route logs its path and never `RawQuery` (R32). **No destination
  parameter is accepted** and the post-login destination is the core
  default (R33, R-5b26).
- **The sign-out handler caps its body with `http.MaxBytesReader`** at
  Task 3's threshold, because security constraint 3's own words are
  "**every** auth endpoint caps its body", then clears the session cookie
  through Task 2's
  helper, with the same `Path`, `Secure`, and absent `Domain` the
  `__Host-` prefix requires. **Its doc comment states what sign-out does
  not do**: these sessions are stateless, so a session captured before
  sign-out keeps working until its expiry or an epoch bump. 5c's logout
  copy must not claim more.
- **Every non-nil error denies** (security constraint 2). `VerifyLink`,
  `VerifySession`, `CheckLink`, `CheckSession` and `Insert` are consumed
  as "nil means permitted, anything else means refused", with the
  sentinel distinction choosing the server-side log line alone. **One
  generic outcome reaches the client for all of them.**
- **Nothing keys on a raw token or cookie string** (security constraint
  3). Every limiter key, log correlation, and dedup derives from the
  decoded nonce or the normalized address.
- **Every auth event is logged with the reported client IP**
  (`spec:1177-1181`): link requested, link consumed, consumption failed,
  and rate limit tripped. **No token, no secret, and no cookie value
  reaches a log record.**
- **The session path applies the same-origin condition to the refresh**
  (`spec:1081-1082`), passing it into Task 9's seam rather than deriving
  the header set here.
- **`internal/auth` registers no route** (`spec:1474-1476`). The handlers
  are values; Task 11 mounts them.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- **The order is proved by behavior, not by reading the function.** A
  token whose MAC fails leaves the nonce directory empty. A token for an
  address removed from the roster leaves the nonce directory empty and
  sets no cookie. A token whose `CheckLink` refuses leaves the nonce
  directory empty and sets no cookie. A replayed token sets no cookie on
  the second presentation. Four rows, each asserting a directory count
  and the absence of a `Set-Cookie` header.
- **N+1 consumption attempts from one client address, every one carrying
  a token whose MAC fails, are refused by the limiter rather than by the
  verifier**, asserted against the server log's outcome, which is the row
  that proves the limiter sits ahead of the MAC.
- **A body one byte over the cap is refused without the parser running on
  all three endpoints**, request link, consumption, and sign out, each
  proved by a body whose parse would otherwise succeed.
- **Under `generic`, the response to a known address and the response to
  an unknown address are compared to each other and match byte for
  byte**, the way the consumption rows already compare responses rather
  than asserting a literal. A second row asserts the same pair when the
  known address's send fails.
- Four distinct failure causes produce byte-identical client responses,
  asserted by comparing the responses to each other rather than to a
  literal, and the server log distinguishes all four.
- A non-sentinel error from the auth core denies, proved with the
  unreadable-roster fixture from Task 9.
- N distinct addresses from one client address trip the per-IP link limit
  and **send no further mail**, asserted against a counting mail seam
  fake that fails the test if called after the trip.
- **Under `kind`**, a send failure produces a failed login rather than a
  success with no mail.
- **A request from a peer outside the trusted-proxy allowlist carrying a
  forged client-address header is limited under its peer address, and the
  auth event logs the peer address**, which is exit-gate assertion 3. The
  limiter and the log record are both reachable here and nowhere else in
  the pass: Task 7 takes the client address as a value and derives none,
  and `ClientIP`'s own header handling is 5a's, covered at
  `internal/serve/clientaddr_test.go:62` and `:206`, while the "is limited
  and logged" half needs a limiter and a handler that did not exist at 5a.
- No log record in any test contains a token, a cookie value, or a
  `RawQuery`, asserted by scanning the captured log output for the
  minted token's own substring.
- The sign-out response's `Set-Cookie` is compared against a written-out
  expectation including the negative `Max-Age`.
- The handlers' constructor takes one struct and the package registers no
  route, provable by reading the file.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Insert the nonce before verifying the MAC → the empty-directory row
   fails, which is the flood the fixed order exists to prevent.
2. Mint the session before `Insert` returns nil → the replay row sets a
   second cookie.
3. Resolve the roster after `Insert` → the removed-member row consumes a
   nonce it should not have.
4. Branch on sentinels alone and permit anything else → the
   unreadable-roster row admits the request.
5. Parse before capping the body → the oversize row's parser runs.
6. Return a distinguishable outcome per failure cause → the
   byte-identity row fails.
7. Key the per-client limiter on the token string → the distinct-address
   flood row stops tripping.
8. Log `RawQuery` on the consumption route → the log-scan row fails.
9. Move the consumption limiter after the MAC check → the failing-MAC
   flood row stops tripping the limiter and trips the verifier instead.
10. Surface the send failure under `generic` → the byte-identity pair
    diverges.
11. Answer under `generic` only after the send completes → the pair still
    matches in bytes, so the row that catches it is the timing one: the
    unknown-address response and the known-address response are measured
    against each other and the known one carries the transport's own
    delay, which the test drives through a seam fake that blocks.
12. Key the per-client limiter on the peer address where a trusted proxy
    is configured, or on the header where it is not → the forged-header
    row fails. This is Task 7's mutation 4, moved here because only this
    call site derives a client address.

**2.0 hook: neutral.** The handlers carry no 2.0 surface; the identity
they produce is hook 5's and it was landed by Tasks 4 and 9.

**Prior art:** none adopted for the handlers. **Records PocketBase's
opposite default** as the road not taken under fork 4: it returns an
opaque id even for an unknown address, treating enumeration as the design
problem, which is what makes fork 4 a real fork rather than a formality
and what R28's two keys plus the global ceiling exist to mitigate under
the kind answer. **Records WCAG 2.2 technique G218** as the accessibility
argument the login flow rests on, whose page-side obligations
(`autocomplete="email"`, a live region for the outcome) are 5c's.

**Independent: no.** It consumes seven tasks.

**Deliverable count: 4** (the request-link handler with fork 4's answer,
the consumption handler with its fixed order, the sign-out handler, the
constructor and its dependency struct).

---

> **Conductor boundary, end of segment D1.** **Accept Task 10 or rule its
> second `fix`.** This is the cheapest boundary in the pass, because
> nothing is re-run to reach it, and it is the only place the conductor's
> rule on a second `fix` can land before work is built on the handlers.
> Revision 1 ran Tasks 9 through 12 in one launch, which meant a second
> `fix` on Task 10 arrived after the chain had already implemented,
> reviewed, gated and committed Task 11 against handlers the conductor had
> not accepted, and after Task 12 had drawn the close's mutation table and
> coverage-gap list from them. `pass-execute` cannot honor "a decision on
> a second `fix` verdict before every later task has already run"
> (`docs/superpowers/plans/2026-09-12-overnight-window.md:134-139`), so
> the decision needs a boundary or it does not exist.

---

## Task 11: The serve verb's auth lifecycle and the sweep's interval driver

Model: `sonnet`. Depends on Tasks 0, 3, 4, 5, 6, 7, 8, 9 and 10, which is
its Consumes line stated as a dependency set rather than a shorter one.

The wiring that makes the pass a running login, and the half of
`spec:1664-1668` the errata confirm 5b-i did not build.

**Files:**
- Modify: `cmd/dubplate-server/serve.go`,
  `cmd/dubplate-server/serve_test.go`
- Modify: `internal/auth/nonce.go`, `internal/auth/nonce_test.go` (the
  interval constant and its driver)

**Interfaces:**
- Consumes: Task 3's resolved values, Task 4's and 5's store, Task 6's
  mail seam, Task 7's limiters, Task 9's identity seam, Task 10's
  handlers, 5b-i's
  `auth.Open`, `OpenNonceStore`, `OpenEpochStore`, and
  `internal/serve.Shell.Register`.
- Produces: the serve verb's auth lifecycle and the sweep driver.
  **Nothing consumes it inside this pass**; 5c's pages and 5d's unit are
  its next readers.

**Outcomes (D13):**

- **The serve verb resolves configuration, opens the stores, builds the
  roster, the limiters, the mailer and the seam, constructs the handlers,
  and mounts them through `Shell.Register`**, in `newServeShell`'s own
  neighborhood so the verb's wiring is what a test exercises rather than a
  restatement of it (`cmd/dubplate-server/serve.go:158-161`).
- **A package constant names the sweep's one-hour period** with its
  reason stated, and **one goroutine calls `Sweep` on it under the verb's
  context and stops with it** (R-5b14 as the errata correct it). The
  sweep at open stays where 5b-i put it.
- **The sweep's own errors are logged and never fatal** after startup,
  consistent with Task 2's m8 change: a record's presence, not its
  readability, is what blocks a replay.
- **Nothing is registered as a bypass pattern** (Task 8's mechanism),
  because the fallthrough is empty at rung 5 and no bearer-authenticated
  path exists yet. The task states this rather than leaving the absence
  unexplained.
- **The verb fails before it listens** when any dependency refuses:
  `config.Load` first, as today, then the stores, the roster, and the
  mailer's configuration, each naming what refused. A login surface that
  accepts connections and fails at the first request that needs a missing
  value is what `newServeCmd`'s existing doc comment already rejects.
- **Nothing deploys.** The unit's `ReadWritePaths` grant over the auth
  leaf and the roster's directory is 5d's (security review m11), and this
  task's report names that dependency explicitly so 5d inherits it rather
  than discovering it at `ProtectSystem=strict`.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The verb's construction is exercised by a test that builds it against a
  temporary state directory and a temporary roster, asserts the three
  routes are registered under the owned prefix, and asserts an unknown
  path under the prefix returns 404.
- A missing roster file refuses the verb before it listens, with the path
  named.
- The sweep driver is proved by driving its ticker source rather than by
  sleeping: the test asserts `Sweep` ran at least once without waiting an
  hour, and the one-hour literal appears in the test with its reason.
- Cancelling the verb's context stops the driver, asserted with
  `-race` and no goroutine left behind, checked by a leak assertion
  rather than by inspection.
- The bypass-pattern absence is asserted: a cookie-carrying
  state-changing request to the fallthrough is denied.
- `doctor` and `check` still pass against the same registry, so the
  verb's new resolution has not changed what the other verbs require.
- Doc comments carry no em dash. `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Mount the handlers outside the owned prefix → the routing row fails.
2. Start the driver without the verb's context → the leak assertion
   fails.
3. Open the stores after listening → the missing-roster row accepts a
   connection before refusing.
4. Register a bypass pattern for the fallthrough → the denial row fails.

**2.0 hook: neutral.** The lifecycle is wiring; 2.0's own surfaces mount
through the same `Register` at rung 12.

**Prior art:** none adopted. The wiring follows the verb's own existing
shape.

**Independent: no.** It consumes four tasks.

**Deliverable count: 3** (the auth lifecycle, the sweep's constant and
driver, the mounts).

---

## Task 12: The pass close, and the 5b-ii to 5c handoff

Model: `sonnet`. Depends on Tasks 0 through 11.

**Files:**
- Draft: `docs/superpowers/research/handoffs/draft-STATUS-5b-ii.md` and
  `docs/superpowers/research/handoffs/draft-HISTORY-5b-ii.md`
- Draft, handed to the conductor in the task return: the spec
  corrections, the coverage-gap list, and the 5c handoff

**The ledger text is written as drafts, never to `docs/STATUS.md` or
`docs/HISTORY.md` themselves.** The conductor folds the drafts into the
real ledgers at the merge and commits them path-limited. The drafts are
committed artifacts, so they survive a restart and a fresh session can
read them. This is 5b-i's own ruled practice (M11), and 5b-i's close
recorded seven draft-content findings its reviewer escalated; the
acceptance criteria below are written against those seven.

**Outcomes:**

- The HISTORY entry in the house shape: what landed, what the gate
  caught, and **what a later pass would be wrong to rediscover**.
- **The security review's disposition table**: every one of the twenty
  findings, B1, M1 through M5, and m1 through m14, with what landed,
  where, and for the three that did not (m9's stateless-session residual,
  m11's unit grant, m13's template rows), which pass owns them and why.
  **A finding folded is a finding with a test**; a finding refused carries
  its reason.
- The pass's contract coverage, D1 through D13, each with its Go site,
  what proves it, and, for a clause no test can reach, the stated reason.
- **The twelve resolutions listed with their markers**, each row naming
  the file and symbol the marker sits on and stating what execution
  confirmed or contradicted. The markers are `// ruled, conductor
  2026-09-12, R-5bN`, stamped by the task that landed each, so this table
  is `grep -rn 'ruled, conductor 2026-09-12, R-5b'` rather than a
  reconstruction.
- **The seven conductor-owed folds, with what execution said about each.**
  Revision 2 applied each as its reviewer recommended and marked it at its
  site; the close records whether the code bore the answer out, so a
  reversal at 5c is informed rather than re-argued.
- The declared coverage gaps recorded as gaps, on one criterion: a
  surface this pass itself builds that no test in this pass can reach.
  **The expected list, to be corrected against what execution found:**
  1. **No page exercises any of this** (5c). The login page, the
     interstitial, and the add-member page are the first real callers of
     the request-link handler, the consumption handler, and the roster's
     write path.
  2. **No test exercises a real bearer-authenticated bypassed surface**
     (Task 8), because none exists until rung 12. The registration
     mechanism itself is tested against a pattern the test registers, so
     the gap is narrower than revision 1 stated it, but the companion
     rule, that a bypassed path may not read the session cookie, is still
     a contract a reviewer reads rather than a test that runs.
  3. **The serve unit's `ReadWritePaths` grant does not exist** (security
     review m11), so nothing proves the auth leaf and the roster's
     directory are writable under `ProtectSystem=strict` until 5d.
  4. **The env template does not carry the nine keys the selected source
     requires** (security review m13), so no deploy can succeed until 5d;
     this is the no-deploy carry, recorded as a gap rather than only as a
     STATUS line.
  5. **A stateless session survives sign-out** (security review m9), by
     design. Clearing the cookie does not invalidate the signed value,
     and only expiry or an epoch bump does.
  6. **The mail transport is proved against a local test server, never
     against a real provider** (Task 6). The two worked examples and the
     `907.life` DNS authorization question are the runbook's and 5d's.
  7. **M5's failing-first test, if the allocation measurement could not be
     made to fail** against the unfixed code (Task 1). Recorded only if
     that fallback fired, with the measurement the task tried and why it
     could not separate the bounded split from the unbounded one.
  **The count is stated and joined to the running total**, which stands
  at seventy-three after 5b-i, so rung 15 does not re-count.
- **The spec corrections the pass owes**, drafted for the conductor, each
  with the text it corrects:
  - **R-5b18's two**: `docs/HISTORY.md`'s 5b-i review record and the R24
    erratum both say the 5b-i security read was not run and is owed to
    5b-ii's window as its first item. It ran on 2026-09-12 and is
    recorded at
    `docs/superpowers/research/2026-09-12-rung-5b-i-security-review.md`.
  - **R-5b26's**: the erratum landed at `b91bae1` states that R33's
    redirect validation "landed in `internal/auth` at 5b-i". No such
    function exists, and Task 0 item 10 proves it by grep. The validator
    lands with rung 12's parameter or not at all.
  - **The ceiling and the reservation**: 5b-ii at 7.5M rather than the
    4.5M provisional, and rung 5 at 24.4M against the live 15.75M, with
    the derivation stated from `docs/HISTORY.md:533-535` alone.
  - **`spec:1587-1593`'s threshold list against what Task 3 declared**,
    if the two disagree in count or name. The expected agreement is ten
    declared, four landed at 5b-i, six landed here.
  - **`spec:2332`'s 5d row**, which assigns "the example roster file
    (R37)" to 5d while this pass creates `config/roster.example.json` at
    Task 4, because Task 4's loader test is what keeps it from rotting.
  - **Any correction Task 0 found** in this plan's own pointers, and any
    erratum from `b91bae1` that execution contradicted. Revision 2 already
    repointed three: the split's pre-authorization is `spec:2163-2172` and
    not `spec:2179-2185`; the universality rule this plan leans on is
    `spec:1592-1593` and not `spec:2506-2508`, which is a section break;
    and the From rule is `spec:1289` and not `spec:1294-1295`. A
    systematic drift of about five lines runs through the mail-seam
    citations besides, each still resolving to its own subject, and Task 0
    reports those as notes.
- **The 5c handoff**, naming what 5c inherits rather than rebuilds: the
  roster's interface and its write path, which the `member` verb and the
  add-member page both call rather than reimplement; **the store's
  write-open constructor, which `member add --admin` calls to create the
  first roster and which refuses to create a second**, so 5c inherits a
  call rather than reopening the package; the single
  resolution point, which 5c's pages must not bypass; **the admin flag's
  channel refusal, enforced in the write path rather than owed to 5c's
  caller**, so the add-member handler inherits a refusal rather than a
  rule to remember; the seam and its channel, which the add-member
  handler reads the admin flag through; the origin check, which covers
  5c's add-member POST by construction rather than by anything the page
  does; fork 4's kind answer as **data**, with the copy 5c's to write and
  the admin contact coming from the roster's admin entry; the
  consumption route's `no-store` and `Referrer-Policy: no-referrer`
  obligations, which are 5c's served header set; the interstitial's
  contract, that it reads the token from `window.location`, clears it
  from the address bar, and POSTs it in a form field; the logout copy's
  bound, that a stateless session survives sign-out; and the three
  remaining rejoin-sentence corrections.
- **The evidence, not a bare "green"**: the mutation table across Tasks 1
  through 11 with every outcome, the deliberately unfired ones included
  and each with its reason; the measured wall times for `internal/auth`,
  `internal/identity`, `internal/notify`, `internal/serve`, `go test
  ./...`, and `bash scripts/check.sh`; the count of existing tests this
  pass changed with the reason for each; the two module dependencies with
  the versions actually taken; and confirmation that `.gitleaksignore`
  and `env/` are untouched.
- **Both budgets scored**: tokens against the 7.5M ceiling, and attended
  time as two counts. **The target is zero planning misses and zero
  execution sittings.**

**Acceptance criteria:**

- **Every claim carries its measurement.** A bare "green" with no
  measured evidence is what 5b-i's reviewer escalated first.
- **Every mutation row's count matches its list.** A row claiming six
  mutations while listing five is what 5b-i's reviewer escalated second.
- **Every doc-comment citation names a comment that carries the argument
  attributed to it**, checked by reading the comment rather than by
  remembering it. Two false citations is what 5b-i's reviewer escalated
  third.
- **A simplifier claim carries its commits.**
- **The STATUS draft stays under the sixty-line cap** and stays present
  tense, with the pass ledger **moved** to the HISTORY draft rather than
  summarized, and **the "Awaiting Geoff's read" block is carried
  forward**, not dropped.
- Every gap in the list is a gap this pass created or inherited, with the
  task and the mutation or the finding that names it.
- Every security-review finding appears in the disposition table exactly
  once.
- No ledger is committed by the implementer, and the lane registry entry
  is the conductor's.

**2.0 hook: neutral.** The close records hook 5 as landed by Tasks 4, 5
and 9, and hook 6's affordance as named by Task 8.

**Prior art:** none consulted. The close reads the pass.

**Independent: no.**

**Deliverable count: 4** (the HISTORY draft, the STATUS draft, the
drafted corrections, the 5c handoff).

---

> **Conductor boundary, end of segment D2 and of the pass.** The
> pass-end actions below.

---

## The exit gate

`bash scripts/check.sh` green, plus every assertion below. The first
seven are `spec:2168`'s own, less the two halves the errata reassigned;
the remainder are this pass's additions.

1. The limiter map's bound holds under a flood of distinct keys.
2. N distinct addresses from one client address trip the per-IP link
   limit and send no further mail.
3. A request from a non-allowlisted peer carrying a forged
   client-address header is limited and logged under its peer address.
   **Owned by Task 10's acceptance criteria**, the only place both the
   limiter and the log record are reachable; revision 1 left this
   assertion on the gate with no task building it.
4. A roster with a duplicate principal, a control character, or an
   unparseable address fails to load, and a missing file fails to load.
5. A roster write whose record fails validation leaves the file
   byte-identical, a concurrent pair of writes both survive, and the
   store answers from the written file with no restart.
6. A protected cookie-authenticated POST with neither `Origin` nor
   `Sec-Fetch-Site` is denied while a bypassed bearer path is not.
7. A recipient or a subject carrying CRLF is refused by the seam.
8. **A link for an address removed from the roster is refused at
   consumption** (R30, plan review m9), with no nonce consumed and no
   cookie set.
9. **A consumed nonce's record outlives the token that bore it across a
   retention tightening** (security review B1).
10. **An address the mint path accepts can be revoked** (security review
    M1): every address `MintLink` accepts, `Bump` can write an epoch for.
11. **A backward `Bump` does not reinstate a refused session** (security
    review M2).
12. **One credential has exactly one verifying string** (security review
    M3), and a separator flood is refused with a bounded allocation
    (M5).
13. **`doctor` never reports a non-failing table while `Load` returns an
    error** (security review M4).
14. **Four distinct consumption failures produce byte-identical client
    responses**, distinguished only in the server log (security
    constraint 2).
15. **Under the generic unknown-address answer, a known address and an
    unknown address produce byte-identical responses**, including when
    the known address's send fails, and the response does not wait on the
    send (R-5b28).
16. **The roster's write path refuses to set the admin flag on any
    assertion channel but the command-line actor**, and the refusal
    changes no byte on disk (`spec:1949`, fork 1).
17. **A consumption attempt is refused by the limiter before its MAC is
    checked**, so a brute-forcer's flood is bounded by the limiter rather
    than absorbed by the verifier (`spec:1178-1179`).

Two assertions of `spec:2168`'s original 5b list belong to neither half
of the split and are not this gate's: the verify-subrequest refusal is
5d's, and `doctor` exiting nonzero on an unset selector was 5b-i's and
passed there.

---

## The conductor's actions at the pass end

Stated here because they are not tasks and no implementer performs them.

1. **Commit the security review** before Task 1 dispatches. It is
   untracked at drafting time and this plan cites it as a primary input.
2. **`code-simplifier` twice, not once per segment.** One round at the
   segment C boundary over Tasks 1 through 8's accepted Go, and one at D2
   over Tasks 9 through 12's. Subagents cannot dispatch agents, so both
   are the conductor's dispatch. **Each round's review is chunked by
   package territory, one `diff-reviewer` dispatch per package**, so
   tokens per reviewed line hold where they are today as the round's diff
   grows. 5b-i ran three batched per-segment rounds rather than one per
   task and **recorded the deviation**; this plan schedules the two-round
   form deliberately, per the process improvement plan's Item 8 half-step
   (`docs/superpowers/research/2026-09-12-process-improvement-plan.md`),
   and the close records it as scheduled rather than as a deviation.
3. **One fresh-context `web-auth-security-reviewer` over this pass's
   surface and Tasks 1 and 2's corrections** (R-5b16, R-5b18), dispatched
   after Task 12's gate and before the merge ritual. **Its input is the
   threat-model section of this plan, the spec's own threat model
   (`spec:1909-1980`), the 5b-i review's "Attacked and found sound" list
   as the do-not-re-litigate record, and the diff.** It does not receive
   this plan's task decomposition. Its findings are folded or refused
   with the reason recorded, and **that record is the do-not-re-litigate
   list for 5c and 5d**.
4. **The merge ritual.** With one lane in the main tree the rebase is a
   no-op, so the ritual is the full gate, the merge, and the full gate on
   master. A `diff-reviewer` over the merged result runs only if the
   rebase was not a fast-forward.
5. **Rule the conductor-owed decisions** this plan names: the security
   fold's home (R-5b17), the security read's discharge and its single
   pass-end run (R-5b18), the 7.5M ceiling with its 6.0M flag, the 8.65M
   reservation overage, the ten resolutions R-5b19 through R-5b28, and
   **the seven conductor-owed folds the header lists**, each marked at its
   own site so ratifying or reversing one is a single edit.
6. **Re-derive 5c's and 5d's ceilings from 5b-ii's actual report** before
   5c's plan is drafted, and record the rung total against R23's
   reservation. Both figures are carried from the same planning rate 5b-i
   overran by nearly three times, so neither is a measurement.
7. **Write the `config/` row into `PA:175-210`** with the other ownership
   rows at the lane launch. `config/` has no row there and had none here
   before revision 2, and this pass adds a file to it.
8. **Do not deploy.** The no-deploy carry runs until 5d. Nothing in this
   pass changes it.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain
(`~/.claude/workflows/pass-execute.js`) per task, **strictly sequential**,
**dispatched in five segments**. Every task runs `dubplate-implementer`,
then `diff-reviewer` (`claude-opus-5`), then the gate, with one
re-dispatch on `fix` and a second `fix` as the conductor's decision.
**The runaway guard is armed at launch. Subagents never dispatch
agents.**

**Segments, and why the boundaries fall where they do.** The 2026-09-12
window proved pass-execute has no mid-run hook: a segment runs to its end
or not at all, so every conductor action lands at a boundary and nowhere
else. Each boundary is therefore placed where the conductor has something
to do that the next segment depends on. **That constraint is also why
segment D splits at the Task 10 / Task 11 boundary**: the rule that a
second `fix` is the conductor's decision needs a place to be taken, and
without the split Tasks 11 and 12 would already be implemented, reviewed,
gated and committed against a handler set the conductor had not accepted.

**The segments satisfy Item 16's ruled sizing** (Geoff, 2026-09-12: three
to four tasks, every boundary on a commit the gate has already proved
green). A and B carry three tasks each. C, D1 and D2 carry two, each under
one of the rule's three named overrides: C ends at the seam where Task 9's
and Task 10's Files are disjoint from Tasks 7 and 8's, D1 ends immediately
before the pass's most expensive `fix` decision, and D2 is the pass's
remainder. Every boundary falls on a task's own green committed gate, so
each is a resume point.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1, 2, 3 | STATUS. Confirm the five reproductions were observed failing before their fixes, which no later task can recover. |
| **B** | 4, 5, 6 | STATUS. **Re-derive the ceiling from Tasks 4 and 5's actuals**, the measurement this plan predicts rather than measures. Name the Task 8 / Task 9 cut if it projects past 8.0M. |
| **C** | 7, 8 | Simplifier over Tasks 1 through 8, its review chunked per package. STATUS. Confirm every dependency Task 10 consumes is accepted, because Task 10 consumes eight tasks and a `fix` there is the pass's most expensive. |
| **D1** | 9, 10 | **Accept Task 10 or rule its second `fix`.** STATUS. Nothing is re-run to reach this boundary, and it is the only place that decision can land before Tasks 11 and 12 build on the handlers. |
| **D2** | 11, 12 | Simplifier over Tasks 9 through 12, its review chunked per package. The security read. The merge ritual. The ledger fold. |

**Models.** `sonnet` by default. **Four named upshifts to `opus`**, all
unconditional, because a `pass-execute` launch fixes each task's model in
its args at dispatch and a condition the conductor cannot observe
mid-launch is dead text:

- **Task 1, the security fold's blocker and majors.** Five defects that a
  reviewer running the code found and that ordinary tests did not, in a
  file set where the failing variant passes every test that checks only
  that the mechanism works. This is CLAUDE.md's "novel
  correctness-critical logic" in the plainest form the build has offered.
- **Task 7, the limiters.** The bound is the property B8 names and R16
  rules, and an unbounded or wrongly-evicting map passes every functional
  test of a limiter. The three key spaces' independence is a second
  property with the same shape.
- **Task 10, the handlers.** The consumption order is seven steps whose
  wrong permutation produces code that logs in correctly and consumes a
  nonce for a revoked member, and the security review's three constraints
  all land here at once.
- **Task 8, the origin check.** Its correctness rests on a placement
  argument rather than on code volume, its blast radius is the whole
  served surface, and revision 2 found it carrying a bypass mechanism that
  cannot work as revision 1 described it.

**Task 0 also runs at `opus`**, fresh context: a staleness check that
misses a moved citation costs more than the model difference, and plan
staleness is the build's named dominant defect class.

**No conditional upshift remains.** Revision 1 named Task 8 a runner-up
to be upshifted "if Task 7's review returns `fix`", and Tasks 7 and 8
share segment C, so the conductor would have seen Task 7's verdict only
after Task 8 had already run at the model its args fixed. The condition is
deleted and Task 8 runs at `opus` outright.

**Review.** Rung 5 keeps the **full fresh-context adversarial cycle**
(`spec:2621-2640`), and 5b-ii's share is the in-chain `diff-reviewer` per
task plus **one fresh-context `web-auth-security-reviewer` at the pass
end** (R-5b16, R-5b18), dispatched by the conductor with the threat model
as its rubric. The two-lens cycle over the whole subsystem is 5d's.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- Task 1 → Task 2. Both write the same files, and Task 2's clamp reads
  the arithmetic Task 1 leaves alone.
- Task 1 → Tasks 4, 5, 7, 9, 10. Every one calls the normalizer.
- **Task 2 → Task 10.** The sign-out handler calls Task 2's
  cookie-clearing helper and the session path calls its clamped
  `RefreshSession`. Revision 1 omitted this edge from Task 10's header,
  which sequential execution hid and a reorder would not have.
- **Tasks 4, 5, 6, 7, 9 → Task 11.** The verb opens the store, builds the
  limiters and the mailer, and constructs the seam before it mounts
  anything. Revision 1's header named only Tasks 0, 3, 8 and 10 while its
  Consumes line named the rest.
- Task 1 → Task 10. The consumption handler calls the new `Insert`
  signature, which is the whole reason the fold runs first.
- Task 3 → Tasks 7, 10, 11. The rates, the cap, and the unknown-address
  answer are resolved once.
- Task 4 → Task 5 → Tasks 9, 10. The write path needs the validated file;
  the seam and the handlers need the resolution point.
- Task 6 → Task 10. The login handler is the mail seam's first caller.
- Task 7 → Task 10. Three limiters, one caller.
- Task 8 → Task 11. The composition change lands before the verb mounts
  anything into it.
- Task 9 → Task 10.
- Tasks 8, 10 → Task 11.
- Tasks 0 through 11 → Task 12.
- **Tasks 4, 6, 7 and 8 are mutually independent by files**, writing
  `internal/auth/roster.go`, `internal/notify`, `internal/auth/limit.go`,
  and `internal/serve/origin.go`, and consuming nothing from each other.
  **They still run sequentially**, because one tree runs one executor and
  the module gate runs over `./...` regardless. The marks exist so a
  conductor can reorder them when a dependency stalls, never so a second
  executor enters the tree.

**Where this pass splits, if it splits again.** Task 0 plus twelve tasks
is under the fifteen-task trigger, so **the recommendation is to keep
5b-ii as one pass.** The cut point, if segment B's re-derivation projects
past 8.0M, is the **Task 8 / Task 9 boundary**, where the work
changes from mechanism to seam and handler. Tasks 0 through 8 would be
5b-ii and Tasks 9 through 12 would be 5b-ii-b. **One reason argues
against taking it**, and it is the same reason 5b-i gave for its own cut
point: the first half alone ships a roster, three limiters, a mail seam
and an origin check with nothing that calls any of them, which is a
surface no security reviewer can grade for the property that matters. The
security read would land on four mechanisms with no caller.

**Budget:** 7.5M ceiling, re-derived here from the one figure 5b-i's
actual report carries and ruled by the conductor 2026-09-12, against the
4.5M provisional 5b-i's plan carried and the 5.5M revision 1 derived from
two figures no report supports. Rung 5's combined figure stands at
**24.4M against the live 15.75M** (`spec:2042-2050`), the **8.65M**
difference recorded rather than taken from another pass. **Flag at 80%,
which is 6.0M**, routed to the conductor; re-derive at the segment B
boundary; STATUS at every segment boundary. Spend is scored at close
against the ceiling and against **zero attended interactions**, which is
this pass's target. The twelve resolutions are ruled at dispatch rather
than at the close, on 5a's and 5b-i's precedent, **together with the seven
conductor-owed folds this revision names**; **none counts as an execution
sitting**, and each resolution carries its marker in the code so a later
ruling is a grep.
