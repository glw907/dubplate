# Rung 5, Pass 5b-i (The Auth Core: the Config Contract, the Token, the Session, and the State Records) Implementation Plan

**Revision 1, 2026-09-08.** Drafted just in time from pass 5a's own
handoff (`docs/HISTORY.md:379-408`), against the rung 5 identity-and-auth
spec at revision 2.1
(`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`), whose
errata now carry 5a's nine corrections (`spec:2050-2157`) and whose six
product forks Geoff ruled on 2026-09-08 (`spec:172-271`).

**Revision 2, 2026-09-08.** Folded from the two-lens fresh-context plan
review at
`docs/superpowers/research/2026-09-08-rung-5b-plan-review.md`, whose
merged 31-item "Verified sound" list is this pass's
do-not-re-litigate record. Thirty-seven raised findings merged to
twenty-eight, all folded, none rejected. Three blockers folded: the
revocation epoch now reads the session's **anchor**, which refresh never
moves, because refresh advances the issue time and a refresh after an
epoch bump would otherwise un-revoke the session permanently (Task 6,
Task 9, A5, A8); the per-source validation hook now receives the
environment snapshot alongside the resolved values, so R31's correlated
threshold check has a landing site inside `Load` (Task 2, Task 4, A2);
and Task 3's territory now carries the three test files its `keys.go`
deletions break or make vacuous, with Task 0 item 10's survey rewritten from registry
constructors to a grep for the deleted literals. Nine majors folded: the
address normalizer lands at Task 1 rather than being assumed by four
tasks; the token's wire encoding is stated rather than left to the
implementer; the nonce record carries consumed-at plus the resolved
retention; a malformed nonce record is retained rather than deleted, so a
crash-truncated record cannot fail open; the gitleaks conjunction that
keeps a golden literal green is a constraint on four tasks;
`spec:1081`'s same-origin refresh clause is carried; R33, R-V and R24
are routed back to the passes that read them; the territory wall names
every file its tasks write; and Task 0 carries sixteen of the spec's
seventeen staleness bullets with the seventeenth named and deferred.
Sixteen minors folded, four of them pointer corrections. **Nothing was
rejected**, and one reviewer measurement made a fold cheaper than the
review proposed: the gitleaks rule needs a key-like identifier **and** a
high-entropy value, so a visibly synthetic fixture secret needs no
rename and only the MAC-bearing golden literals do.

**Task count and ceiling are unchanged.** Task 0 plus ten tasks,
eleven deliverables. Task 1 moves from two deliverables to three, for the
normalizer. Task 0 moves from twenty-four items to twenty-five. The
ceiling stays **5.0M** with the flag at **4.0M**, and the arithmetic
table gains one line for the fold rows, moving the expectation from 4.49M
to **4.61M** against 0.39M of remaining forecast. The rung reservation's
baseline is corrected: the overage is **5.5M against the live 15.75M**,
not 6.75M against the superseded 14.5M, because 5a's close already
recorded 1.25M of that difference.

**Nine folds are conductor-owed rather than applied silently**, each
marked "conductor-owed" inline where it sits: the epoch's anchor check
(A8, Task 9); the hook's environment snapshot (A2, Task 2); the token's
wire encoding (A4, Task 5); the nonce record's carried value (A7, Task
8); the malformed-record retention rule (Task 8); the R33, R-V and R24
routing (primary input 1, Inheritance); the 5.5M overage against the
reservation (the ceiling section, conductor action 5); R38's widening to
the loopback set (Task 4); and the lane registry's home in
`docs/STATUS.md` (Global Constraints, Task 0 item 1).

**This plan takes the pass split the spec pre-authorized.** `spec:2179`
states the trigger: "If 5b's plan exceeds fifteen tasks, split the pass
rather than the tasks," with the cut point "between the token, session,
and state records on one side and the roster, limiters, and seam on the
other, which is where the risk changes from cryptographic construction to
input validation." An honest decomposition of `spec:2168`'s whole 5b row
lands seventeen items. The split is taken here, at the spec's own cut.
**This plan is 5b-i and lands Task 0 plus ten tasks. The follow-on is
5b-ii**, whose scope, exit gate, and provisional ceiling are named below
and whose plan is drafted just in time from this pass's handoff. **The
split, both ceilings, and the second security read the split creates are
conductor-owed decisions**, argued where each belongs.

**Band: L.** `spec:2161-2163` sets rung 5 at L band across four passes,
and the split makes 5b two of them without changing the band. Three of
rung 5's passes change a contract, and 5b-i's contract is `Load`'s own:
the wall stops being four hard-coded requires and starts requiring what
the registry carries.

**What this pass is.** 5b-i builds `internal/identity`, the first half of
`internal/auth`, and the config-core change that makes rung 5's keys
real. It lands the verified-identity leaf, the wall's requirement
mechanism and the per-source declaration family, rung 5's own key
declarations with both `internal/config/keys.go` retirements, the
magic-link token, the `__Host-` session cookie, the state-directory leaf
with its digest naming rule, the create-exclusive consumed-nonce store,
and the revocation epoch with the seam 2.0's credential store subscribes
to.

**What this pass is not.** The roster store and its write path, the three
rate limiters, `internal/notify`, the origin check, the verify-identity
seam, and the auth handlers are **5b-ii's, in full**. No page, no
`member` verb, no systemd unit, no deploy path, and no `doctor` probe row
is built here; those are 5c's and 5d's. **A task that believes it needs a
roster, a limiter, a mail send, an origin check, a handler, a page, a
unit, or a probe row stops and reports.**

> **For agentic workers:** workflow mode per CLAUDE.md's "Build process"
> section, the pass-execute chain per task, sequential. The implementer is
> `dubplate-implementer`, registered at
> `.claude/agents/dubplate-implementer.md`. The conductor consumes
> structured reports and never reads a diff, a test log, or a gate
> transcript inline. Subagents never dispatch agents, so `code-simplifier`
> after each accepted task and the in-pass `web-auth-security-reviewer`
> after the chain are both the conductor's dispatch (`spec:2455-2457`).
> The per-task cadence is CLAUDE.md's git-conventions rule, which every
> task's own commit invokes; `spec:2455-2457` says the conductor owns the
> dispatch and puts it at each pass close.
> The conductor owns the ledgers (`docs/STATUS.md`, `docs/HISTORY.md`,
> `ROADMAP.md`), the lane registry, the contested-surface rows, and the
> batch commit of anything a task drafts into them. **This pass runs in
> LANE 2, a git worktree at `.claude/worktrees/lane-2-5b` on branch
> `lane-2-5b`**, concurrently with pass 8c in the main tree. No task runs
> from the main tree, and no second executor enters the worktree. **The
> live-executor check is the lane registry in `docs/STATUS.md`, never
> `pgrep -f Projects/dubplate`**, which matches every lane once a worktree
> exists.

---

## Primary inputs

Read these before Task 1, in this order.

1. **`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`,
   revision 2.1.** The pass-split table's 5b row (`spec:2168`) is this
   plan's scope statement, less the half this plan hands 5b-ii. The
   sections 5b-i implements are "What rung 5 builds" (`:93-148`), the
   ruled product forks (`:172-271`), the rulings (`:273-770`), "The auth
   core" for its token, consumption, and session subsections
   (`:1012-1101`), "Packages and their contracts" (`:1465-1496`), "Config
   keys" in full (`:1498-1656`), "The state directory" (`:1658-1704`),
   the 2.0 hook table (`:1841-1859`), the prior-art table (`:1861-1907`),
   the threat model (`:1909-1980`), the errata (`:1982-2157`), the
   contested surfaces (`:2186-2218`), the band and ceiling
   (`:2220-2277`), the staleness pre-task (`:2281-2345`), the risks
   (`:2347-2440`), the process section (`:2442-2490`), and universality
   (`:2492-2523`). Rulings **R4, R5, R6, R10, R11, R17, R18, R24, R26,
   R29, R31, R32, R33, R34, R35, R38** are the ones 5b-i reads. R1, R2,
   R3, R7, R8, R9, R15, R16, R25, R28, R30, R37 belong to 5b-ii; R12,
   R13, R14, R19, R20, R21, R22, R36 belong to 5c and 5d. **Revision 1
   routed R33 and R24 to 5c and 5d**, against `HISTORY:393-394` for the
   first and against R-5b16's own reading for the second; the
   Inheritance section dispositions R33 and R-5b16 corrects R24
   (**conductor-owed**, review M-g).
2. **`docs/HISTORY.md`'s "Rung 5, pass 5a" entry (`:5-543`)**, and inside
   it, at minimum: "The exported surface" (`:175-188`), "The declared
   coverage gaps" (`:323-378`), "The 5a to 5b handoff" (`:379-408`),
   "Forward carries" (`:409-426`), "What a later pass would be wrong to
   rediscover" (`:484-507`), and "The merge ritual" (`:531-543`). **The
   handoff is binding on this plan**, and the Inheritance section below
   states what each clause costs a task.
3. **`docs/HISTORY.md`'s rung 1b entry, the contract paragraph
   (`:1491-1507`)**, in the words that bind: the wall is registry
   contents, `Load` takes the registry it enforces against, a rung
   declares its keys in its own file inside the package that owns them,
   no lane adds a name to `internal/config/keys.go`, and a lane removes
   only its own reserved class from that file in the same commit that
   declares the class in its own package. 5a exercised the removal half
   once, for the `serve` class. **5b-i exercises it for the `auth and
   identity` class and, for the first time anywhere, for a pending wall
   row.**
4. **`docs/superpowers/research/2026-09-08-rung-5-spec-inventory.md`
   (INV), sections 1 (`:60-228`), 3 (`:388-522`), 4 (`:523-611`), 5
   (`:612-718`), 6 (`:719-825`), and its thirteen surprises
   (`:1187-1256`).** Section 3 is the config surface as it exists beside
   what the birth requirements ask for, and surprises 1, 2, 3, and 7 are
   the four this pass answers or forwards. **Do not re-derive what INV
   settles. Cite it, and verify any line carried forward**, because
   several of INV's `keys.go` pointers moved when 5a deleted the `serve`
   class.
5. **`docs/superpowers/research/2026-09-08-rung-5-auth-prior-art.md`
   (PRA), in full.** Its sessions and token sections (`:46-183`) are the
   adopt-or-decline evidence this pass records; its verdict table
   (`:23-44`) is the summary; its "Provisional rulings" (`:655-677`) and
   "Evidence notes and unverified items" (`:678-717`) are what a task
   cites rather than re-researches. The Prior art section below carries
   the verdict per entry with the criterion that decided it.
6. **`docs/superpowers/research/2026-09-08-rung-5a-security-review.md`**,
   the twenty-four findings and their dispositions. **This is the
   do-not-re-litigate record.** One residual is routed to this pass:
   finding 7, an absolute-form request target sets `r.Host` to the
   client's chosen authority, so **every URL 5b builds comes from the
   configured public origin and never from `r.Host`**. Findings 16, 22,
   23 and 24 are 5c's, 5d's, and the wall contract's, and this pass
   reopens none of them.
7. **`docs/superpowers/research/2026-09-08-rung-5a-plan-review.md`'s
   "Verified sound" list (`:705-860`), thirty items.** A pass plan does
   not reopen an item on it without new evidence. Items 6, 11, 12, 13,
   16, 23, 24, 29 and 30 bear directly on this pass: R27's two-address
   split stands as landed, the registry contract's removal half is
   followed as `HISTORY:1491-1507` words it, `Registry.claim` panics on a
   duplicate across families, `doctor`'s two line-count assertions derive
   from the registry rather than from a literal, `main.go`'s doc comment
   already names every verb it should, Task 0's shape is right, the
   worktree and merge-ritual rules are stated in three places, deliverable
   counts are stated per task, and the review shape is the in-chain
   `diff-reviewer` plus one fresh-context reviewer with the threat model
   as its rubric.
8. **`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` (S20),
   the forward-hooks section (`:354-383`) and the identity-and-app-passwords
   section (`:234-252`).** Hook 5 (`S20:367-369`) is the hook this pass
   lands, and `S20:245-246` is the requirement behind R35: every app
   password carries the user's valid-after epoch, so a credential store
   must be able to **read** the epoch. Being told it changed is not enough.
9. **`docs/superpowers/specs/2026-09-03-parallel-build-approach.md` (PA),
   its lane table (`PA:96-101`), its scheduling rule (`PA:80`), and its
   contested-surface ownership table (`PA:175-203`).** `PA:100` gives
   lane A `internal/auth`, the `internal/serve` shell, and
   `internal/notify`. `PA:177` is the config-wall row this pass edits
   against, and `PA:202` is the `internal/config` row.
10. **The tree, read-only, before writing anything:**
    `internal/config/config.go` (the package doc's five classes at
    `:6-42`, `Config` at `:66-118`, the `wall` type and `require` at
    `:120-178`, `rejectNamespace` at `:200-239`, and `Load` at
    `:241-287`); `internal/config/registry.go` (`AddWallRow` at `:138`,
    `AddReservedClass` at `:163`, `AddIntegrationClass` at `:211`,
    `AddThreshold` at `:267`, `requirePrefix` at `:106`, `claim` at
    `:119`, `panicDuplicate` at `:95`, `validateThresholdDefault` at
    `:306`); `internal/config/thresholds.go` (the kind constants at
    `:16-24`, `Threshold` at `:45`, `thresholdOf` at `:112`, `ParseString`
    at `:128`); `internal/config/keys.go` (the pending wall row at `:24`,
    the `auth and identity` reserved class at `:60-73`, the two Access
    supersession messages at `:159-162`, `DUBPLATE_CONTRIBUTORS_FILE`'s at
    `:165-166`); `cmd/dubplate-server/main.go` (`newRootCmd` at `:82-102`,
    the three register calls at `:83-86`, the `AddCommand` block at
    `:94-100`); `cmd/dubplate-server/serve.go` (the threshold declarations
    at `:11-73`, `newServeCmd` at `:122-144`, the `NewStrip` call at
    `:136`); `cmd/dubplate-server/doctor.go` (`newDoctorCmd` at `:23-71`,
    `thresholdStatus` at `:116-150`); `cmd/dubplate-server/doctor_test.go`
    (the extra-line assertion at `:130`, `TestDoctorCmd_LineCountFollowsRegistry`
    at `:151`); `cmd/dubplate-server/serve_test.go`'s `productionRegistry`
    (`:19-25`); `cmd/dubplate-server/templatekeys_test.go` in full;
    `internal/atomicfile/atomicfile.go` (`WriteFile` at `:16`, `WriteStream`
    at `:55-91` with the directory fsync at `:82-89`);
    `internal/engine/shadow/store.go` (`storeDirName` at `:41`, `StoreDir`
    at `:46`, `OpenStore` at `:53`, `validID` at `:66`);
    `env/dubplate-server.env.template` and `env/dubplate.env.template`;
    `scripts/check.sh`.
11. **`_salvage/`, read-only, on merit, never imported.** **There is no
    `_salvage/internal/auth`.** The pre-pivot tree carried no token, no
    session, and no nonce store, because identity was Cloudflare Access's.
    The nearest material is `_salvage/internal/access/access.go`, a
    Cloudflare Access JWT verifier whose JWKS-cooldown reasoning is worth
    reading once and whose dependency R26 declines
    (`golang-jwt/jwt/v5` enters `go.mod` only in the commit that builds an
    Access front). `_salvage/internal/config/contributors.go` is the
    fail-open roster loader R2 supersedes, and it is **5b-ii's** input,
    not this pass's.
12. **`CLAUDE.md` in full**, its "Facts agents get wrong" list in
    particular, and `docs/STATUS.md`'s lane state, which today is prose
    inside the Next action section (review m13).
13. **`~/.claude/skills/go-conventions/SKILL.md` before any Go.** This
    pass writes Go only. No Python, no Svelte, no TypeScript, no shell,
    no unit file, so `vps-conventions`, `svelte-conventions`, and
    `ts-conventions` are all out of scope.

---

## Goal

Ship the auth core's cryptographic half and the config contract that
makes rung 5 configurable at all.

Land `internal/identity` as a dependency-free leaf holding the verified
identity and its assertion channel, so `internal/ops` can carry a
verified identity at rung 6 without importing a server package. Land the
config core's requirement mechanism, so a registered owned row is a
required row and a selector declared in a rung's own file cannot report
green while unset. Land the per-source declaration family, so the eleven
names magic-link needs are required by mechanism rather than by prose,
accepted by the wall rather than reported unknown, and reported per
source by `doctor`. Land rung 5's own declarations with the `auth and
identity` reserved class and the pending identity-source wall row both
retired in the commit that replaces them. Land the magic-link token with
its unambiguous serialization, its two labels, its fixed verification
order, and its verify-only previous secret. Land the `__Host-` session
cookie whose rolling refresh moves the expiry and never the anchor. Land
the state-directory leaf with one digest naming rule and one `validID`
check covering both records, the create-exclusive consumed-nonce insert
that a replay cannot survive across a restart, and the revocation epoch
with the read-plus-notification seam a 2.0 credential store subscribes
to.

Nothing here reads a roster, limits a rate, sends mail, checks an origin,
serves a route, renders a page, writes a unit, touches the box, or reaches
the live network. **Nothing here adds a module dependency**: R26's two are
both consumed at 5b-ii.

---

## Inheritance from pass 5a, applied

Each clause of `HISTORY:379-408` and the security review's own additions,
with the task that carries it.

- **5b mounts auth handlers through `Shell.Register`.** No handler exists
  at 5b-i, so the mount lands at **5b-ii**. This plan's territory wall
  forbids writing `internal/serve` at all.
- **`ClientIP` is 5a's and is never re-derived.** The three limiter keys
  read it at 5b-ii. **No task in this plan derives a client address**, and
  a task that believes it needs one stops and reports.
- **The origin check inserts after the strip and after normalization**
  (R-F). 5b-ii's, with the insertion point already documented in 5a's
  composition. This plan neither builds it nor reserves a slot for it.
- **The session-cookie name is taken at `Shell` construction** (R-E), and
  5a wired `""` there (`cmd/dubplate-server/serve.go:136`). **Task 6 wires
  the real name**, which retires 5a's declared coverage gap 4.
- **`Origin` and `Sec-Fetch-Site` survive ingress** (R-S), and Go's
  `CrossOriginProtection.Check` returns nil when both are absent. Carried
  verbatim into the 5b-ii handoff; no 5b-i surface reads either header.
- **Every URL builds from the configured public origin, never `r.Host`**
  (security review finding 7). **Task 5 holds it**, because the magic
  link's hostname is the first URL rung 5 builds and `spec:1033-1034`
  already requires it. The rule travels to 5b-ii as a constraint on every
  later URL.
- **R33's post-login destination rule is 5b's** (`HISTORY:393-394`,
  `spec:680-692`). **Conductor-owed**, disposed here rather than
  reassigned: rung 5 accepts no caller-supplied destination, **the token
  carries no destination field at 5b-i**, and any later destination rides
  inside the MAC-covered token rather than as a free query parameter on
  the consumption POST (`spec:690-692`). **Task 5 holds the token half**,
  and the rule travels to 5b-ii and 5c as a constraint on the token
  format rather than as new work.
- **The 30-second read timeout and the `http.MaxBytesReader`
  alternative** (R-V, `HISTORY:391-392`). Carried verbatim into the 5b-ii
  handoff, because `spec:1590`'s auth request body cap is 5b-ii's and the
  timeout is what bounds a body before the cap exists. No 5b-i surface
  reads a request body.
- **`Referrer-Policy` and the served header set are 5c's.** Nothing in
  this pass sets a response header.
- **The moved `cmd/dubplate-server/templatekeys_test.go` classifies
  against the production registry, which 5b's new keys join.** **Task 3
  holds it**, and its acceptance criteria name the two templates by path.

**The merge review's three carries, ruled.**

| Carry | Ruling | Reason |
| --- | --- | --- |
| `templatekeys_test.go` classifies a pending wall row as owned where `Load` would refuse it (`classifyKey`, `:95-100`) | **Landed here, Task 3** | Task 3 removes the last pending row, which would close the divergence by accident and leave the latent disagreement standing for the next rung that files one. The fix is one condition, and the task that retires the row is the task that owes it. |
| `productionRegistry` (`serve_test.go:19-25`) restates `newRootCmd`'s register calls by hand | **Landed here, Task 3** | Task 3 adds a fourth register call, which is the second time the hand-restatement would have to be maintained in step. One shared constructor in the binary's own package, called by `newRootCmd` and by the test, removes the class. |
| `internal/serve/server.go:161` passes the listener as a goroutine parameter, a pre-1.22 idiom | **Forwarded to 5d** | 5b-i writes no `internal/serve` file, and opening one for a behavior-preserving idiom fix widens the diff a fresh-context security reviewer reads for no gain. 5d already reads `server.go` for the bound-address `doctor` row (R-G), so the fix lands beside a reason to be in the file. |

---

## Owner touchpoints: none

5b-i raises no attended question. `spec:2484-2490` names rung 5's three
expected touchpoints, and all three sit elsewhere: the six product forks
are ruled and closed, the `style-src 'unsafe-inline'` sitting rides with
5c's pages, and the box deploy is 5d's. **Anything else is a conductor
ruling** (`spec:2488`).

**What would become an owner touchpoint, so a task recognizes one rather
than absorbing it.**

- **Any secret minted.** Rung 5's three secrets (`DUBPLATE_MAGIC_LINK_SECRET`,
  `DUBPLATE_SESSION_SECRET`, `DUBPLATE_SMTP_PASSWORD`) are minted at rung
  16 through the age store (`spec:1632-1638`). **5b-i declares their names
  and mints nothing.** A task that finds itself wanting a real secret
  value stops and reports.
- **Any bare `NAME=` row added to `env/dubplate-server.env.template`.**
  That file is a manifest `deploy.sh` renders from the age store, and it
  hard-fails the whole render on one unresolved name (CLAUDE.md, "Facts
  agents get wrong"). **Task 3 adds exactly one row and it is a literal**
  (R-5b7 below), which resolves without the store.
- **Any adopter-visible default.** The four threshold defaults, the
  session cookie's name, the secret length floor, and the state
  directory's leaf name are what an adopter inherits when they configure
  nothing. R-5b8 through R-5b12 rule all of them as applied, and each
  states what would change under a different answer. A task changes none
  on its own judgment.
- **Any change to what the box runs.** 5b-i deploys nothing and enables
  nothing. **One consequence needs stating rather than discovering**:
  once Task 2 and Task 3 land, `config.Load` requires
  `DUBPLATE_IDENTITY_SOURCE`, and `dubplate-server shadow` calls `Load`
  against the same registry (`spec:1648-1656`). The box's live shadow
  timers keep running the deployed binary and are unaffected. **The
  rendered `/etc/dubplate/server.env` carries the selector from Task 3
  onward**, because R-5b7 puts it in the template as a literal and
  `deploy.sh` renders that file from the template, so the row itself is
  not what a deploy waits on. **`Load` then requires the nine
  correlated magic-link keys** (A2, `spec:1522-1536`), none of which the
  greenfield template carries today and three of which are secrets rung
  16 mints (`spec:1632-1638`). That whole list is what 5d's exit gate
  satisfies (`spec:2170`), and **no deploy runs during rung 5's open
  passes.**

**A discovered question is a finding to report, never a mid-pass ask.**
Every stop condition below resolves to the conductor, and the
80%-of-ceiling flag routes there too.

---

## The pass split, taken

`spec:2179-2185` pre-authorized this call and gave the cut point.

**Why the honest decomposition exceeds fifteen.** `spec:2168`'s 5b row
carries: the identity leaf; token serialization; the session cookie; the
nonce store; the epoch store; the revocation seam as a read plus a
notification; the roster store's load, validation, admin flag, write
path, and single resolution point; the consumption re-checks;
three rate limiters plus a bounded evicting map; the origin check with
its bypass; the verify-identity seam; `internal/notify` with its
transport and CRLF guard; the reserved-class removal; the wall row's
retirement; the `Load` edit; the correlated validations; and three
supersession rewrites. At the repo's task granularity, where a task ends
in an independently testable deliverable a reviewer could reject while
accepting its neighbor, that is seventeen items. Compressing it to
fifteen would put five and six deliverables in single tasks, which is the
"splitting tasks instead of the pass" failure mode CLAUDE.md names, and
`spec:2392-2397`'s risk 8 is filed against exactly this pass.

**The cut, as the spec drew it.**

| | Scope | Exit gate | Ceiling |
| --- | --- | --- | --- |
| **5b-i** (this plan) | `internal/identity` born. `internal/auth` born with the token, the session, and the two state records. The config core's requirement mechanism and per-source declaration family. Rung 5's key declarations with both `keys.go` retirements, the three supersession rewrites, and one template literal row. Four thresholds and the load-time validations. | `bash scripts/check.sh` green, plus: a replayed nonce is refused across a process restart; a token whose MAC fails never reaches the nonce store; a session older than the epoch is refused, and so is a link minted before an epoch bump; refresh never moves the anchor and a session past the cap is refused; a configuration whose nonce retention floor is below the link expiry fails to load; `doctor` exits nonzero and names the selector when `DUBPLATE_IDENTITY_SOURCE` is unset. | **5.0M**, flag 4.0M |
| **5b-ii** (plan owed, just in time from this pass's handoff) | The roster store with B9's validation, its admin flag, its atomic write path, its audit record, and its single resolution point (R1, R2, R3, R37). The three-key limiters with their bounded evicting map (R16, R28). `internal/notify` with its SMTP transport and CRLF guard (R7, R8, R9). The origin check with its bearer bypass (R15). The verify-identity seam with provenance (R17). The auth handlers mounted through `Shell.Register`, with the consumption-time roster and epoch re-checks (R30). The remaining thresholds. | The remaining assertions of `spec:2168`: the limiter map's bound under a flood; N distinct addresses from one client address tripping the per-IP link limit; a forged client-address header limited and logged under its peer address; the roster's four load refusals and its three write properties; a protected cookie-authenticated POST without a same-origin signal denied while a bypassed bearer path is not; a recipient or subject carrying CRLF refused by the seam; a link for an address removed from the roster refused at consumption (R30). | **4.5M** provisional, re-derived from 5b-i's actual |

**Two assertions in `spec:2168`'s exit gate belong to neither half**, and
the close drafts the correction. "A link for an address removed from the
roster" is refused by the consumption-time roster re-check, which is
5b-ii's, **so it is written into 5b-ii's exit-gate cell above and named
in the handoff** rather than left to a prose sentence with no carrier. "A
verify subrequest from such a peer is refused" grades the forward-auth
verify endpoint, which `spec:2170` gives to **5d**.

**What the split costs, stated plainly.** Prep, the staleness pre-task,
the security read, and the merge ritual are each paid twice. The ceiling
arithmetic below carries the whole of it, and the conductor rules the
overage against R23's reservation.

**What the split buys.** The security read that grades a token, a session
cookie, and a create-exclusive nonce insert is a different read from the
one that grades a roster's input validation, a limiter's bound, and an
origin check's bypass. `spec:2172-2177` made 5a its own pass on exactly
that argument, and R24 already added a second `web-auth-security-reviewer`
read at 5c for a surface the earlier read could not cover
(`spec:577-586`). The same reasoning applies here, one step earlier.

---

## Token ceiling: 5.0M (re-derived from 5a's actual)

`spec:2233-2236` grants the re-derivation in the same sentence that sets
the figures: "Each ceiling is re-derived from the previous pass's actual
report before the next plan is drafted." The previous pass's actual report
is 5a's (`HISTORY:519-530`), and it prices a pass of nearly this shape.

**The basis.** 5a closed at about **3.90M** against a 4.25M ceiling, for
Task 0 plus eight chain tasks in one worktree, pure Go, no corpus and no
network: plan preparation 0.83M, Task 0 0.13M, the chain 1.68M across 18
agents, the post-chain fix round 0.27M, the security review 0.18M, the
security fold 0.22M, the merge review 0.09M, arguments and commits 0.30M,
conductor turns 0.20M. Its chain rate is **0.21M per task**. 5b-i is the
same shape with two more tasks, a heavier read, and a denser security
surface.

**The arithmetic, anchored on 5a's actual rather than summed from
scratch.**

| Line | Basis | Figure |
| --- | --- | --- |
| 5a's measured total, Task 0 plus eight chain tasks in a worktree | `HISTORY:519-530` | 3.90M |
| Two more chain tasks, ten here against eight there | 5a's measured chain rate, 1.68M over eight tasks | +0.42M |
| Heavier prep | Three packages against one, a config-core contract change, twelve key names against three, and a spec section set twice 5a's | +0.10M |
| A longer staleness pre-task | Sixteen of the spec's seventeen bullets plus the nine this pass adds, twenty-five items against 5a's nineteen at 0.13M | +0.02M |
| A denser security read and fold | The token, the session, and the nonce store are the pass's cryptographic core, where 5a's shell was a chain of wrappers | +0.05M |
| The plan review's own folds | Twenty-eight findings adding acceptance rows to nine tasks, one deliverable to Task 1, and one item to Task 0, on 5a's own +0.10M for folds across six tasks | +0.12M |
| **Expected total** | | **4.61M** |

**The ceiling is ruled at 5.0M, conductor 2026-09-08, as applied.** A
4.5M ceiling puts the 80% flag at 3.6M, which on a 4.61M expectation
fires around Task 7 of ten, mid-chain, where the conductor's only
available action is to continue. That is the arithmetic announcing
itself rather than a signal, and rung 1b's overrun is the recorded case
of a flag crossing inside the chain with no conductor action
(`spec:2225-2227`). At 5.0M the flag lands at **4.0M**, past the chain
and inside the close and the reviewer, where a stop is cheap and the
choice is real.

**Flag at 80%, which is 4.0M.** Finish the task in flight, write STATUS,
and send one combined question to the conductor. **The conductor's action
when it fires is the choice this plan already names**: whether the in-pass
`web-auth-security-reviewer` runs here or its read moves to 5b-ii's, and
whether the close absorbs the spec corrections or hands them forward. The
arithmetic to check at that point is 4.0M against a 4.61M expectation,
leaving 0.61M of forecast against 1.0M of headroom.

**Re-derive the ceiling from Task 3's actual report.** Task 3 is the
declarations task, the one whose blast radius across templates, fixtures
and tests this plan predicts rather than measures. If the re-derivation
passes 5.5M, the conductor names the split at the Task 6 / Task 7
boundary, where the work changes from cryptographic construction to
durable records on disk.

**Recorded against the rung reservation.** R23 (`spec:569-575`) set rung 5
at 14.5M across four passes, and 5a's close already moved the reservation
to **15.75M**, recording the 1.25M difference then
(`spec:2042-2050`). **15.75M is therefore the live baseline**, and
computing against 14.5M now would re-record that same 1.25M. With 5b
split, the rung's figures become 5a at a 4.25M ceiling and about 3.90M
actual, 5b-i 5.0M, 5b-ii 4.5M provisional, 5c 4.5M, 5d 3.0M, taking the
ceiling for each open pass: **21.25M against the live 15.75M**. The
**5.5M** difference is recorded against the reservation rather than taken
from another pass, because 5c and 5d each own surfaces this pass cannot
price and the spec already schedules their re-derivation from actuals.
**This is a conductor-owed ruling**, and the honest statement behind it
is that R23 priced undivided 5b's seventeen items at what 5a spent on
eight.

**Checkpoint: STATUS after every four tasks**, so after Task 3, after
Task 7, and at the close. Every task reports its own spend, its mutation
outcomes, and its measured gate wall time, and the conductor sizes the
next task against it. **Every task boundary is a resumable stopping
point** (`spec:2275-2277`), because pass 8a's chain met a session window
cap before its own ceiling would have bound.

**Deliverable count: eleven.** Ten task deliverables plus the close. No
task carries more than four, and Task 2 and Task 3 carry four each and
are named below as the two to watch. Revision 2's folds move Task 1 from
two internal deliverables to three, for the address normalizer, and add
no task.

---

## Global Constraints

- **Lane 2, a git worktree at `.claude/worktrees/lane-2-5b`, branch
  `lane-2-5b`.** `spec:2168`'s tree column for 5b is "Worktree", and
  CLAUDE.md's Lanes section is the rule behind it: the gate runs the
  module leg over `./...` plus two whole-repo legs, so two code lanes need
  two trees. **Lane 1 runs pass 8c in the main tree, writing
  `internal/engine/tagio`, concurrently with this pass.** The lane
  registry entry is written before the first dispatch, and the
  live-executor check is that registry. **`docs/STATUS.md` carries lane
  state as prose inside its Next action section rather than under a
  labeled heading** (review m13), so the conductor either adds a labeled
  "Lane registry" block before the first dispatch or reads Next action as
  the registry. **Conductor-owed**, and either answer keeps the rule that
  `pgrep -f Projects/dubplate` is never the check.
- **Package territory, stated as a wall.** This pass writes
  **`internal/identity`** (born this pass, including its tests),
  **`internal/auth`** (born this pass, its token, session, and store files
  only), **`internal/config`, each file for the change its task names, its
  tests included** (`config.go`, `registry.go`, `thresholds.go`,
  `keys.go`, and the `config_test.go`, `registry_test.go` and
  `keys_test.go` sites a task's own change reaches),
  **one new file under `cmd/dubplate-server`** (the auth declarations),
  **`main.go`'s register calls and the shared registry constructor**,
  **`serve.go`'s one-line cookie-name wiring**, **`serve_test.go`'s
  `productionRegistry` and its environments**,
  **`templatekeys_test.go`'s `classifyKey`**, **`doctor.go`'s per-source
  row loop with `doctor_test.go`'s fifth-family assertion and its literal
  spot check**, and **one literal row in
  `env/dubplate-server.env.template`**. It writes **nothing** under
  `internal/serve`, `internal/engine`, `internal/harness`,
  `internal/atomicfile`, `corpus/`, `web/`, `systemd/`, `scripts/`,
  `tools/`, `Makefile`, or `_salvage/`, and no `docs/` file outside Task
  10's conductor-owned drafts. A task that believes it needs one of those
  stops and reports.
- **`internal/serve` is closed to every other lane until rung 5 closes**
  (`PA:80`), and **5b-i is not the lane that opens it**. Task 0 item 2 is
  the blocking check that no lane has written it since 5a's merge.
- **No roster, no limiter, no mail, no origin check, no handler, no page,
  no unit.** Every one is named in `spec:2168` and every one lands at
  5b-ii, 5c, or 5d. **A task that reaches for one stops and reports.**
- **`internal/auth` and `internal/identity` read configuration through
  values passed in, never from the environment** (`spec:1471-1479`).
  Neither package imports `internal/config`. `internal/identity` imports
  no other internal package at all (`spec:1467-1469`, R17). Resolution
  happens in `cmd/dubplate-server`.
- **Module dependencies: none added.** `go.mod` carries cobra and its two
  indirects and stays that way (`go.mod:5-10`). R26 (`spec:591-600`)
  budgets rung 5 exactly two, `github.com/wneessen/go-mail` and
  `golang.org/x/time/rate`, and **both are 5b-ii's**. Every mechanism here
  is stdlib: `crypto/hmac`, `crypto/sha256`, `crypto/rand`,
  `crypto/subtle`, `encoding/base64`, `net/mail`, `net/url`, `os`,
  `time`. **A task that reaches for `go get` stops and reports.**
- **Assertion discipline, at the repo's bar.** No silent-success fakes,
  no vacuous assertions, **no expectation derived from the code under
  test**. Every hard-coded expectation comes from the spec, from an RFC,
  or from documentation a task cites by URL and date. A test that asserts
  a MAC by recomputing it with the same function proves nothing; a golden
  vector written out and checked in is what proves it.
- **A cryptographic property is proved by behavior, never by structure.**
  A test that reads the verification function's statement order proves the
  code says what it says. A test that presents a token with a broken MAC
  and asserts the nonce directory is still empty proves the order. Risk 3
  (`spec:2364-2369`) is the reason.
- **Prefer fixtures that are obviously not credentials** (`spec:2216-2218`).
  Every test secret is a visibly synthetic literal long enough to pass the
  floor, every address is documentation-reserved or invented, and no
  fixture is shaped like a real key. **gitleaks scans the tracked
  worktree with `--no-git`**, and a flagged fixture is a fixture chosen
  wrong before it is an allowlist entry.
- **The gitleaks rule is a conjunction, and each half is an escape.**
  Measured on the gate's own binary and its own invocation
  (`scripts/check.sh:110`), `generic-api-key` fires only when a key-like
  identifier sits on the same line as a value whose entropy clears the
  threshold, which `check.sh:104-109` states and
  `internal/config/keys_test.go:59-65` already applies in the tree. **A
  fixture secret escapes through its value**: a visibly synthetic literal
  does not clear the entropy threshold under any identifier. **A golden
  literal that must carry entropy, because it carries a MAC, escapes
  through its identifier**: no `token`, `key`, `secret` or `apikey` in the
  name that sits beside it. Measured, `goldenToken` fires at entropy 5.17
  and `goldenVector`, `wantEncoded`, `nonceDigest` and
  `goldenSessionCookie` do not. **The precedent is rename rather than
  suppress**, so `.gitleaksignore` stays untouched.
- **Universality.** No estate hostname, path, email, catalog value, or
  personal name anywhere (`spec:2492-2500`). **Every tunable is a
  threshold with a default** (`spec:2506-2508`). No address lives in
  configuration at all, because fork 1 moved the admin contact onto the
  roster record (`spec:1547-1555`).
- **Drivers own driver knowledge** (`spec:2501-2504`). Nothing in this
  pass names a streamer, a header, or a path a driver would supply.
- **`go-conventions` before any Go**, at the standing exemplary-stdlib
  bar (`idiomatic-beautiful-go`). Concrete types over interfaces; `any`
  not `interface{}`; no `init()`; guard clauses over nesting; `ctx` first
  where a call can block or reach a filesystem; errors lowercase and
  context-first with `%w` only where a caller branches. **The em dash is
  banned in Go comments**, and vale scans every new file before it is
  staged.
- **Sentinels where a caller branches.** The auth core's refusals are Go
  errors, so `errors.Is` has real surface here: a replayed nonce, an
  expired token, a failed MAC, and an epoch-refused session are four
  outcomes a caller distinguishes. Each wraps a package sentinel and each
  test asserts through `errors.Is` rather than on message text.
- **No token, no secret, and no cookie value ever reaches a log record or
  an error string** (`spec:1181-1185`, `spec:1946`). An error naming a
  token says which check failed, never what the value was. This is a
  criterion on every task from Task 5 onward.
- **The gate is `bash scripts/check.sh`, green at every task end**, run
  **in the worktree**. Never commit red. Specific files only. Commit
  footer `Co-Authored-By: Claude <noreply@anthropic.com>`.
- **Every task stamps its resolutions' markers.** A task landing the
  surface a resolution decides writes `// ruled, conductor 2026-09-08,
  R-5bN` on the symbol that carries the decision, in the repo's existing
  convention (`internal/config/registry.go:137`, `keys.go:35`,
  `cmd/dubplate-server/serve.go:25`). Task 10's resolution table is then a
  grep rather than a reconstruction.
- **A sub-second test run tells you nothing.** Anything reported as
  evidence runs with `-count=1` and reports wall time.
- **The staleness pre-task runs first.** Task 0, below. Its output is
  blockers with evidence, not fixes; the conductor rules each one before
  Task 1 dispatches.

---

## Contested surfaces

`spec:2186-2218` is the source. The conductor writes these into the
parallel-approach spec's ownership table before the lane launches, per
lane-launch checklist item 2.

| Surface | Owner | Consumers | 5b-i's position |
| --- | --- | --- | --- |
| `internal/identity` | rung 5, created by it | rung 6's ops request types, rung 12's driver, 2.0's credential store | **Created whole, Task 1.** Nothing later in rung 5 changes it. |
| `internal/auth` | rung 5, created by it; no other lane writes it | rung 7 and rung 14 through the seam; rung 12 through the seam only | **Created, half.** The token, the session, and the two state records land here; the roster, the limiters, the origin check, and the seam land at 5b-ii in the same package. |
| `internal/config/keys.go` edits | rung 5 removes the `auth and identity` reserved class and the pending identity-source wall row, and rewrites the three supersession messages its own keys replace | every later rung, by the same rule | **All three edit classes land at Task 3**, in the commit that declares their replacements. 5a took the `serve` class already. |
| `internal/config/config.go`, `Load`'s owned-row resolution | rung 1's contract; rung 5 adds the identity-source require and the per-source correlated resolution (R29) | every rung that adds a required owned row | **Task 2, whole.** This is the pass's contract change. |
| `internal/config/registry.go` and `thresholds.go` | rung 1b's registry | every rung declaring keys | **Task 2 adds one declaration family and its accessor.** No existing family changes shape. |
| `cmd/dubplate-server` root command and `doctor` rows | rung 1's skeleton; rung 5 adds `serve`, the `member` verb group, and the probe row kind | 6, 9, 13 | **Task 3 adds one register call and one new file.** The `member` verb is 5c's; the probe rows are 5d's. |
| `env/dubplate-server.env.template` | rung 4 by its own header; rung 5 adds commented pending rows only | rung 16 mints | **Task 3 adds one literal row**, which the row-state header admits and which R-5b7 argues. The other six non-secret rows stay 5d's. |
| `internal/serve` | rung 5, created at 5a; rung 12 thereafter, with no concurrent serve-writing lane | 6, 7, 12, 14 | **Untouched by 5b-i.** 5b-ii inserts the origin check and mounts handlers; 5d takes the merge carry on `server.go:161`. |
| `.gitleaksignore` | rung 1b retired its one fingerprint; the lane that needs an allowed finding appends and names it in its pass report | every lane | **5b-i expects to append nothing** and reports it if it must. |
| `go.mod` | shared | every lane | **Untouched by 5b-i.** |
| `docs/` ledgers | conductor-owned | every rung | A pass drafts and never commits. |

**Lane 1 runs pass 8c concurrently, in the main tree, writing
`internal/engine/tagio`.** Its territory and 5b-i's are disjoint by
package. The two lanes share the tree through `.gitleaksignore`,
`go.mod`, `cmd/dubplate-server`, and `docs/`. **5b-i writes
`cmd/dubplate-server`**, so the rebase at the merge ritual is the one
place the lanes can collide, and Task 0 item 3 reports on it before the
pass starts.

**The merge ritual closes the lane**, stated here because the close task
executes it: the conductor takes the merge token, rebases the worktree
branch onto master, runs the full gate in-tree, merges `--ff-only`, runs
the full gate on master, and dispatches **one `diff-reviewer` over the
merged result whenever the rebase was not a fast-forward**. The second
code slot is released at this pass close rather than at the rung close
(`spec:2273-2275`), so the worktree is removed once master is green.
**gitleaks runs against the worktree from the main tree**, per the
two-lane run mechanics 5a recorded.

---

## The auth core's contract, as outcomes

Every clause below is a property the pass ships. Each names the task that
lands it and the spec clause that requires it. A task builds to this
section; it does not re-derive it.

### A1. A registered owned row is a required row

`Load` requires every owned-now wall row the registry carries, and a row
carrying a `Pending` status is neither required nor read (`keys.go:10-13`).
Rung 1's four rows keep their behavior exactly, because they are the four
rows `NewRegistry` seeds. A rung that registers a wall row in its own file
gets a requirement from the registration, with no edit to
`internal/config`. Every missing row is reported once, joined with every
other failure, in the words `doctor` already prints. **Task 2**
(`spec:623-636`, risk 11 at `spec:2409-2415`).

### A2. One identity source is selected by configuration, and its correlated set is required by mechanism

A rung registers a source by name, with the keys that source requires,
the keys it accepts and never requires, and a validation hook. **The hook
receives the source's resolved values and the environment snapshot in
`os.Environ()` form** (**conductor-owed**, review B2), so a rung's own
file can validate a registered threshold's value, and a relation between
two of them, without putting a rung 5 name in `internal/config`. Without
the snapshot R31's floor has no landing site: a threshold resolves at its
point of use through `config.ParseDuration`
(`internal/config/thresholds.go:91-93`) and `Load` reads no threshold
value at all (`config.go:248-287`), and a name cannot be declared in two
families because `Registry.claim` panics across them
(`registry.go:119-124`). The registration also names the selector key, so
`internal/config` carries no rung 5 name of its own. `Load` requires the
selector when at least one source is registered, refuses a value naming
no registered source with an error listing the registered names, requires
the selected source's required keys, reads its optional keys, and runs its
hook. A name in any registered source's set classifies as declared rather
than unknown, which is the wall's sixth classification. `Config` carries
the active source's name and an accessor for its resolved values, on
`Config.IntegrationValue`'s own precedent (`config.go:66-118`). **Task 2**
(`spec:1538-1546`, R5 at `spec:340-353`, R29 at `spec:623-636`).

### A3. Rung 5's names are declared in rung 5's own file, and both retirements land in the same commit

`DUBPLATE_IDENTITY_SOURCE` is registered as an owned wall row, and the
pending `{Name: "identity source"}` row is deleted from
`internal/config/keys.go` in the same commit (`keys.go:24`). The
magic-link source is registered with nine required names
(`DUBPLATE_MAGIC_LINK_SECRET`, `DUBPLATE_SESSION_SECRET`,
`DUBPLATE_PUBLIC_URL`, `DUBPLATE_ROSTER_FILE`, `DUBPLATE_SMTP_HOST`,
`DUBPLATE_SMTP_PORT`, `DUBPLATE_SMTP_USER`, `DUBPLATE_SMTP_PASSWORD`,
`DUBPLATE_SMTP_FROM`) and two optional names
(`DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS`, `DUBPLATE_SESSION_SECRET_PREVIOUS`),
and the `auth and identity` reserved class is deleted in the same commit
(`keys.go:60-73`). Three supersession messages are rewritten: the two
Access rows now say the seam admits a forward-auth front and that its
configuration arrives with the rung that builds one, and
`DUBPLATE_CONTRIBUTORS_FILE` now names `DUBPLATE_ROSTER_FILE` rather than
rung 6's store. **Task 3** (R1 at `spec:280-284`, R5, `spec:1502-1514`,
`spec:1620-1628`).

### A4. The magic-link token cannot be forged, shifted, replayed, or outlived

The token carries a normalized login address, an expiry, a mint time, and
a 128-bit nonce from `crypto/rand`. **Every field is length-prefixed in
the serialization, not in the MAC input alone**, so no value can be
shifted across a boundary by a separator inside it, which is what
`spec:1017-1019` claims. **The wire encoding is stated rather than left
to the implementer** (**conductor-owed**, review M-b): each field is
base64url-encoded before any separator is applied, so a field value can
never carry the separator and a dotted address round-trips. A version
label and a domain-separation label sit inside the MAC input as package
constants with their reason documented. **The token carries no
destination field**, and any later destination rides inside the
MAC-covered token rather than as a free query parameter (R33,
`spec:690-692`); the version label is what admits one without changing
the meaning of an existing field. The MAC is HMAC-SHA256, verified
with `crypto/subtle.ConstantTimeCompare`. **The verification order is
fixed: MAC, then expiry, then the nonce insert**, because inserting first
lets an unauthenticated caller fill the nonce directory.
`DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS` verifies and never mints. The link's
hostname derives from `DUBPLATE_PUBLIC_URL` and never from a request
header. **Task 5** (`spec:1014-1034`, R32 at `spec:661-679`, security
review finding 7).

### A5. A session cannot outlive its absolute cap, and rolling refresh cannot extend it

One HMAC-signed cookie named with the `__Host-` prefix carries the
normalized address, the issue time, an anchor time, and an expiry, under
its own domain-separation label. Its attributes are `Secure`, `HttpOnly`,
`SameSite=Lax`, `Path=/`, and no `Domain`, which the `__Host-` prefix
requires and which makes the cookie host-scoped by definition. **Rolling
refresh moves the expiry and never the anchor**, so the absolute cap
survives refresh with no server-side state. **Refresh moves the issue
time too, so the anchor is the session's original issue time and the only
field that survives a refresh unchanged.** Every check that must reason
about when the session was born reads the anchor, the epoch's included
(A8). **Refresh counts same-origin requests only** (`spec:1081-1082`), so
a cross-origin subresource load cannot hold a session open to its cap on
a machine its owner has walked away from.
`DUBPLATE_SESSION_SECRET_PREVIOUS` verifies and never mints. **Task 6**
(`spec:1071-1084`, R38 at `spec:758-767`).

### A6. Both state records are named by one derivation and confined to one root

A record's name is the hex of a SHA-256 digest, over the nonce for one
record and over the normalized login address for the other, and both pass
one `validID` check refusing an id that is empty, `.`, `..`, or carrying a
separator. The leaf is one package-level constant so the unit's
`ReadWritePaths` grant is a single static path, and one function joins it
so the name has one home. Files are 0600 under 0700 directories. A test
proves that every write both records take, and the attacker-shaped
`..`-bearing id they refuse, stay confined to the root passed in. **Task
7** (R34 at `spec:694-705`, R10 at `spec:413-420`, `spec:1681-1704`,
`internal/engine/shadow/store.go:37-70` as the shape adopted).

### A7. A consumed nonce cannot be consumed twice, across a restart or across any accepted configuration

The insert opens the record with `O_CREATE|O_EXCL`, syncs the file, and
syncs the containing directory, so first-writer-wins is a property of the
filesystem call rather than of a check-then-write. **The record carries
consumed-at plus the resolved `DUBPLATE_NONCE_RETENTION`, never the
token's own expiry** (**conductor-owed**, review M-c), and the sweep
prunes on that carried value, at open and on an interval. Config load
refuses a retention below the link expiry, naming both keys, so the
record always outlives every token that could bear its nonce; under the
other reading the retention key has no runtime consumer at all and R31
guards a mechanism the design does not have. **A record whose payload
does not parse is retained rather than removed**, until its file mtime is
older than the resolved retention (**conductor-owed**, review M-d),
because a crash between the create and the write leaves a zero-length
record whose token may still verify. **Tasks 4 and 8** (R11 at
`spec:421-444`, R31 at `spec:650-660`, `spec:1664-1668`).

### A8. The revocation seam exposes a read as well as a notification

The epoch record is one per normalized login address, written through
`internal/atomicfile.WriteStream` because overwriting is exactly what a
revocation does and because only `WriteStream` fsyncs the containing
directory after the rename (`atomicfile.go:82-89`). The seam exposes a
**read** returning the current epoch for a resolved identity, which is the
authority, plus a change notification, which is the convenience. **A
session whose anchor predates its address's epoch is refused**
(**conductor-owed**, review B1), and so is a token whose mint time
predates it. The anchor rather than the issue time, because rolling
refresh moves the issue time (A5): a session refreshed once after a bump
would carry a post-bump issue time and be permitted from then on, which
is R30's own failure mode (`spec:643-645`) reproduced on every request.
The anchor is what refresh never moves, so no ordering rule between the
check and the refresh is load-bearing. The address key is recorded as rung 6's
re-key obligation rather than left to be discovered. **Task 9** (R4 at
`spec:333-339`, R11, R35 at `spec:706-715`, `S20:245-246`).

---

## Resolutions this plan takes rather than discovers

Sixteen. Each is a question the spec leaves open or answers in a form the
tree contradicts. Each states what was decided, why, and what would change
under a different answer. **Each is ruled by the conductor at dispatch**,
on 5a's precedent, and none counts as an execution sitting. The marker
convention is `// ruled, conductor 2026-09-08, R-5bN`, continuing into
5b-ii from R-5b17.

### R-5b1: the pass splits, and this plan is 5b-i

**Decided:** take `spec:2179-2185`'s cut. 5b-i carries the identity leaf,
the config contract, the token, the session, and the two state records.
5b-ii carries the roster, the limiters, the mail seam, the origin check,
the verify seam, and the handlers. **Conductor-owed.** Under the other
answer, one fifteen-task pass carries five and six deliverables per task
and one security read grades both a MAC construction and a roster's input
validation.

### R-5b2: the auth keys are declared in `cmd/dubplate-server/auth.go`

**Decided:** on R-C's own precedent from 5a (`cmd/dubplate-server/serve.go:11-25`)
and rung 4's `shadow.go` before it. Declaring inside `internal/auth` would
make the auth package import `internal/config`, which `spec:1471-1479`
forbids. Declaring inside `internal/config` would violate the rung 1b
contract (`HISTORY:1501-1505`). Under the other answer, the package that
mints tokens also reads the environment, and the seam rung 6 substitutes
behind grows a configuration dependency.

### R-5b3: all twelve names are declared at 5b-i, though six of their consumers arrive at 5b-ii

**Decided:** the `auth and identity` reserved class is one indivisible
literal, and the rung 1b contract requires its deletion in the commit
that declares its replacement. `DUBPLATE_ROSTER_FILE` and the five SMTP
names therefore land here, declared and required-when-selected, with
their consumers arriving one pass later. This is what a registry
declaration is for: `doctor` reports the name, `Load` requires it under
the selector, and no code reads it yet. Under the other answer the class
is deleted while half its names classify as unknown, and a correctly
configured deployment fails to load.

### R-5b4: `Load` requires every owned-now wall row, rather than naming rung 5's selector by hand

**Decided:** R29 (`spec:623-636`) asks for "a `w.require` call for
`DUBPLATE_IDENTITY_SOURCE`" inside `internal/config`. The generic form
discharges the same requirement without putting a rung 5 name in the
config core, and it closes risk 11 (`spec:2409-2415`) by construction for
every later rung rather than for this one key. Rung 1's four rows are
unaffected, because they are the four rows `NewRegistry` seeds and no
other wall row exists. **The close drafts the correction to R29's
wording.** Under the other answer, `internal/config` names a rung 5 key,
the rung 1b contract acquires an exception, and the next rung to register
a wall row rediscovers that registering does not require.

### R-5b5: the per-source correlated set is a registry declaration family, not a widened integration class and not a hard-coded switch

**Decided:** the registration carries the selector key, the source name,
its required keys, its optional keys, and a validation hook over the
resolved values. `spec:1546-1548` refuses to widen the integration family
for one consumer, because an integration gates on a boolean flag and a
selector carries a value. A hard-coded switch inside `Load` would put nine
rung 5 key names in `internal/config`. The hook is what keeps the value
rules (a secret's length floor, a URL's scheme, a duration's relation to
another) in the rung's own file while still failing at load, joined with
every other error, in the words `doctor` and `check` already print. Under
the other answer, either the registry grows a boolean-flag shape that
cannot express a value, or the config core learns rung 5's vocabulary.

### R-5b6: the two rotation-overlap names are the source's optional keys, never thresholds

**Decided:** `doctor` prints a `KindString` threshold's **effective
value** (`cmd/dubplate-server/doctor.go:142-147`), and
`DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS` carries a secret. Declaring either
as a threshold would print a secret on an operator's terminal.
`validateThresholdDefault` also refuses an empty `KindString` default
(`registry.go:306-329`), which an unset-by-default rotation key needs, so
the threshold family is wrong on both counts. The source's optional set
accepts the name, never requires it, and prints no value. Under the other
answer, a rotation key is either a printed secret or a sentinel default
that a reader has to be told about.

### R-5b7: `DUBPLATE_IDENTITY_SOURCE` lands as a literal row in `env/dubplate-server.env.template` in the task that makes it owned-now

**Decided:** `TestTemplateKeysCoverEveryOwnedNowRow`
(`cmd/dubplate-server/templatekeys_test.go:74-89`) is a hard gate
asserting that every owned-now row appears in the greenfield template. A
pass that registers an owned-now row without its template row commits
red. The row is `DUBPLATE_IDENTITY_SOURCE=magic-link`, a literal in the
template's second row state (`env/dubplate-server.env.template:8-16`),
which needs no mint and resolves without the age store. **This
contradicts `spec:2205`'s "commented pending rows only" and `spec:2170`'s
assignment of the template rows to 5d**, and the close drafts the
correction. The other **six** non-secret names stay 5d's, because none is
a wall row and no gate reaches them. `spec:1640-1645` says "eight" and
enumerates seven, and `spec:2170` gives 5d "its eight literal rows"; the
close drafts both count corrections. Under the other answer, the gate is
red from Task 3 until 5d.

### R-5b8: the secret length floor is 32 bytes

**Decided:** HMAC-SHA256's key should be at least the hash output length,
which RFC 2104 section 3 states as the recommendation and which is 32
bytes for SHA-256. The check counts bytes of the raw configured value
after no transformation, so an adopter pasting a base64 rendering of 32
random bytes passes with room to spare. The runbook's minting instruction
is 32 bytes from `crypto/rand`, base64-encoded. Under a shorter floor, a
short secret is accepted and the construction's security rests on a value
an adopter chose casually.

### R-5b9: the session cookie is named `__Host-dubplate_session`

**Decided:** the `__Host-` prefix is required by A5 and carries its own
attribute rules. The remainder names the product rather than the estate,
which universality requires, and it is stable across adopters so a
support answer is the same everywhere. 5a's shell takes the name as a
construction value (R-E), and Task 6 wires this one. Under a configurable
name, an adopter gains a knob nobody needs and the shell's own removal
step gains a failure mode when the two disagree.

### R-5b10: the four thresholds this pass declares, and their defaults

**Decided:** `DUBPLATE_MAGIC_LINK_EXPIRY` (duration, `15m`),
`DUBPLATE_SESSION_LIFETIME` (duration, `2160h`, which is 90 days),
`DUBPLATE_SESSION_ABSOLUTE_CAP` (duration, `4320h`, which is 180 days),
and `DUBPLATE_NONCE_RETENTION` (duration, `24h`). The first three are
product fork 3's ruled values (`spec:216-229`). The fourth is not a fork
value, and `24h` is chosen because the record is a digest name and a few
bytes of expiry, the sweep is cheap, and a day of retention survives a
clock adjustment or a paused process without weakening single use. The
remaining thresholds of `spec:1587-1593` are declared at 5b-ii, in the
pass whose code reads them. Under a shorter retention default, R31's floor
does more work than it should; under a longer one, nothing breaks and the
directory grows.

### R-5b11: the state leaf is `auth`, with one subdirectory per record

**Decided:** `<DUBPLATE_STATE_DIR>/auth` is the one leaf, holding
`nonce/` and `epoch/` beneath it, so the serve unit's `ReadWritePaths`
grant is one static path (`spec:1691-1697`,
`internal/engine/shadow/store.go:37-47` as the shape adopted). Fork 1's
roster leaf is a second, separate leaf declared at 5b-ii, so the grant
covering the auth records covers nothing else (`spec:1672-1676`). Under a
per-record leaf, the unit gains two grants and rung 4's single-static-path
argument stops holding.

### R-5b12: the record digest is full SHA-256 hex, untruncated

**Decided:** 64 hex characters, no truncation. A truncated digest invites
a collision argument that a nonce store cannot afford to lose, and the
record names are never read by a person, so length costs nothing. Under a
truncated name, the store's single-use property depends on a birthday
bound nobody measured.

### R-5b13: the token rides in the link's query string, never in a path segment

**Decided:** R32 (`spec:661-679`) bounds the token's exposure with a
consumption route that logs its path and never `RawQuery`. That mechanism
only bounds a token that is in the query. A path-borne token would be
written by every access log and by 5a's own refusal record
(`internal/serve/serve.go:446`, the bounded path). The interstitial reads
the token from `window.location` and the consumption POST carries it in a
form field (`spec:1038-1044`), both 5c's. Under a path-borne token, R32's
own mitigation stops applying and the pass would owe a new one.

### R-5b14: the nonce sweep's interval is a package constant, and its driver arrives at 5b-ii

**Decided:** the retention floor is the tunable an adopter has a reason to
change, and the sweep's period is not; it is one hour, a package constant
with its reason stated. 5b-i ships the sweep as a method plus the sweep at
open, because both are testable without a server. The goroutine that calls
it on the interval belongs to the serve verb's lifecycle, which 5b-ii
wires alongside the handlers. Under a threshold, an adopter gains a knob
whose wrong value has no visible symptom.

### R-5b15: 5b-i writes no `internal/serve` file

**Decided:** the origin check (R-F), the handler mounts, and the merge
carry on `server.go:161` all move to 5b-ii or 5d. A pass that opens
`internal/serve` for one line widens the diff a fresh-context security
reviewer reads over the shell it already cleared. The one exception is
`cmd/dubplate-server/serve.go`'s construction call, which is the binary's
wiring rather than the shell. Under the other answer, this pass's security
read reopens 5a's surface.

### R-5b16: a fresh-context `web-auth-security-reviewer` runs at 5b-i's own pass end, and again at 5b-ii's

**Decided:** R24 (`spec:577-586`) budgets one read "inside 5b", written
before the split existed. The split creates two surfaces that no single
read covers well: a MAC construction, a cookie's refresh arithmetic, and
a create-exclusive insert on one side, and a roster's input validation, a
limiter's bound, and an origin check's bypass on the other. R24 already
set the precedent by adding a second read at 5c for a surface the earlier
read could not cover. **Conductor-owed**, priced in both ceilings, and
drafted as a spec correction at this pass's close. Under one read, the
half that lands second is graded by a reviewer holding the half that
landed first.

---

## The threat model, as the in-pass reviewer's rubric

The reviewer receives `spec:1909-1980`, this section, and the diff. It
does not receive this plan's task decomposition. A finding that
contradicts the spec's threat model is a finding against the spec.

**What this pass protects.** The two HMAC secrets. The single-use property
of a consumed nonce. The absolute cap on a session's life. The revocation
epoch's authority over both a session and an outstanding link.

**Who the adversary is, for this pass's surface.** An unauthenticated
visitor holding a link that was mailed to someone else, or an expired one,
or one whose fields have been shifted. A person on a shared device holding
a live session past an offboarding. A mail gateway that follows links. A
process on the box that can read the state directory.

**The claims this pass's code must hold**, drawn from `spec:1936-1958`:

1. A token verifies before it is looked up, so a caller with a broken MAC
   cannot write a nonce record.
2. A nonce cannot be consumed twice, across a process restart, and across
   any configuration `Load` accepts.
3. A session cannot outlive the absolute cap, no refresh moves the
   anchor, and refresh counts same-origin requests only.
4. An epoch bump refuses every session older than it, including a session
   minted from a link that predates the bump **and a session refreshed
   after the bump**, because the check reads the anchor rather than the
   issue time refresh advances.
5. A token never reaches a log record and never appears in an error
   string.
6. A record name derived from an address cannot escape its root, and the
   address does not appear in a filename.
7. A secret below the floor fails at load rather than at the first
   verification.
8. Every URL is built from the configured public origin, never from
   `r.Host` (security review finding 7).

**Attacks the rubric names, and what answers each.** A shifted-field
forgery, answered by length-prefixed serialization. A cross-purpose
replay of a session value as a link value, answered by the two
domain-separation labels. A timing oracle on the MAC comparison, answered
by `crypto/subtle`. A nonce-directory flood by an unauthenticated caller,
answered by the fixed verification order. A replay across a restart,
answered by `O_CREATE|O_EXCL` plus two fsyncs. A replay after a sweep,
answered by R31's floor at load. A session extended past its cap by
repeated refresh, answered by the untouched anchor. A revoked person's
outstanding link, answered by the epoch comparison against the mint time.
A revoked person's live session kept alive by refresh, answered by the
epoch comparison against the anchor.
A path-traversal record id derived from admin-authored roster data,
answered by the digest plus `validID`.

**What this pass cannot answer, and does not claim to.** Rate limiting,
enumeration, origin checking, mail injection, and roster validation are
5b-ii's whole surface. The reviewer is told so, and a finding against one
of them is recorded as a 5b-ii requirement rather than as a 5b-i defect.

---

## Task 0: The staleness pre-task

Model: **`opus`**, **fresh context, zero prior context**, dispatched by
the conductor before any implementer. Reads the rung 5 spec and this plan
against the tree as shipped. **Its output is blockers with evidence, not
fixes**; it writes no code and edits no file. The conductor rules each
blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports what that
HEAD is.** This plan's `spec:`, `HISTORY:`, `INV:`, `PA:`, `PRA:`, and
`S20:` pointers were resolved at `6a9570d`. A citation that has moved
since is a finding for this task rather than a surprise for an
implementer.

**Grade a moved pointer by whether its content moved.** A pointer off by
fewer than about ten lines whose cited content still resolves to the same
subject is a note in the report. A pointer landing on a different subject,
or naming content that no longer exists, is a blocker.

**Four pointer corrections are already folded from the plan review, so
confirm them rather than rediscover them.** The ceiling erratum is
`spec:2042-2050`, not `spec:2078-2087`, which is the hop-by-hop and
trailer errata. The `NewStrip` call is
`cmd/dubplate-server/serve.go:136`, not `:137`, which is the `NewShell`
call. Fork 3 runs `spec:216-229`, not `spec:210-227`. "Anything else is a
conductor ruling" is `spec:2488`, not `spec:2489-2490`. **Report any of
the four that has moved again.**

The checklist is **sixteen of the spec's seventeen bullets**
(`spec:2281-2345`), reduced to nothing, with the design-language
default-theme bullet (`spec:2334-2337`) named and deferred to 5c on item
20, extended with the nine this pass adds. Twenty-five items. **The first
three are blockers: a "no" stops the launch.**

1. **Is the lane registry consistent with where this pass intends to run,
   and is the target tree free?** `docs/STATUS.md`'s lane registry is the
   live-executor check, **never `pgrep -f Projects/dubplate`, which
   matches every lane once a worktree exists**. *(Expected at drafting,
   read from `docs/STATUS.md`'s Next action section, which is where lane
   state lives today: lane-1 IDLE with pass 8c's plan as its next action
   in the main tree; lane-2 IDLE, the `lane-2-5a` worktree torn down at
   5a's merge, about to take this pass at
   `.claude/worktrees/lane-2-5b`; `git worktree list` showing one tree at
   master.)* **Report any live executor in the
   target worktree, any uncommitted change under `internal/config`,
   `cmd/dubplate-server`, or `internal/serve`, and confirm the worktree is
   at master's HEAD.**
2. **Has any other lane written `internal/serve` since 5a's merge?**
   `PA:80` forbids it until rung 5 closes. *(Expected: no commit touching
   it since `307001b`.)* **Report the commit list.**
3. **Has lane 1's pass 8c touched 5b-i's territory?** *(Expected: no. 8c
   absorbs the shadow's reader into `internal/engine/tagio`.)* **Report
   any 8c commit touching `internal/config`, `cmd/dubplate-server`,
   `go.mod`, `.gitleaksignore`, or `env/`**, since those are the surfaces
   where the two lanes could collide and where the merge ritual's rebase
   would stop being a fast-forward.
4. **Does `internal/config/keys.go` still carry the `auth and identity`
   reserved class and the pending identity-source wall row, at the shapes
   this spec cites?** *(Expected: the pending row at `keys.go:24`, the
   `auth and identity` class at `keys.go:60-73` with eight keys and rung
   5. **The `serve` class the spec's own bullet still names is gone**, 5a
   removed it, so that half of `spec:2297-2300` is already stale.)*
   **Report both line ranges and the stale spec bullet.**
5. **Is `DUBPLATE_CONTRIBUTORS_FILE`'s supersession message still the one
   naming rung 6's store?** *(Expected: `keys.go:165-166`. The spec cites
   `keys.go:177-178` at `spec:1507`, R1 itself cites `keys.go:176-177` at
   `spec:284`, and INV cites `:176-177`; all three moved when 5a deleted
   the `serve` class.)* **Report the real line and all three stale
   pointers**, because the close drafts a correction for each. Also report the two Access rows *(expected:
   `keys.go:159-162`, cited as `:171-176` at `spec:1506`)*.
6. **Does the registry still expose `AddWallRow`, `AddReservedClass`,
   `AddIntegrationClass`, `AddThreshold`, and the kind constants, with
   `IntegrationClass.Enabled` still parsing a boolean flag?** *(Expected:
   `registry.go:138`, `:163`, `:211`, `:267`; `thresholds.go:16-24`;
   `keys.go:227`.)* **Report `AddWallRow`'s status-row carve-out
   (`registry.go:130-152`)**, which is what Task 3's deletion of the
   pending row interacts with, and **`validateThresholdDefault`'s
   `KindString` case (`registry.go:306-329`)**, which R-5b6 turns on.
7. **Is `cmd/dubplate-server/main.go`'s single `AddCommand` block still
   the wiring point, and does `doctor` still derive every row from the
   registry?** *(Expected: `newRootCmd` at `:82-102`, three register calls
   at `:83-86`, one `AddCommand` block at `:94-100` carrying five
   subcommands; `newDoctorCmd` at `doctor.go:23-71`, four family loops, no
   per-rung knowledge.)* **Report the doc comment's subcommand list**,
   which already names every verb rung 5 adds.
8. **Do `doctor`'s tests still pin the output's line count, and what does
   this pass's arithmetic do to them?** *(Expected: the extra-line
   assertion at `doctor_test.go:130` and
   `TestDoctorCmd_LineCountFollowsRegistry` at `:151`.)* **Compute and
   report**: this pass removes one reserved class, removes one pending
   wall row, adds one wall row, adds four thresholds, and adds one
   per-source row, for a net of plus four on the production registry.
   **Then report which tests use `config.NewRegistry()` and which use
   `productionRegistry()`**, because a family the fresh registry never
   populates passes both assertions vacuously. **If either assertion
   would fail, that is a blocker for Task 2 and the conductor rules it
   before dispatch.**
9. **Does `internal/config/config.go`'s `Load` still hard-code its four
   owned requires, and does `Config` still carry no generic wall
   accessor?** *(Expected: `Load` at `:248-287` with four `w.require`
   calls plus the generic integration loop; `Config` at `:66-118` with
   `IntegrationEnabled` and `IntegrationValue` as the accessor
   precedent.)* **Report the package doc comment's five-class list
   (`:6-42`)**, which Task 2 extends to six, and **the `Config` doc
   comment's "the identity-source selector joins its fields at its own
   rung" sentence (`:67-68`)**, which Task 3 lands.
10. **Which gate-resident test sites do Task 2's requirement change and
    Task 3's two `keys.go` deletions break?** **Survey by grep for the
    two deleted literals, `"identity source"` and `"auth and identity"`,
    and for every owned-now wall row a test registers**, never by which
    registry constructor a test calls: a test that builds no registry at
    all reads the package table directly, and the fresh registry is
    seeded from the same table the deletions empty. *(Expected, five
    sites, from the plan review's own measurement at `6a9570d`:
    `internal/config/keys_test.go:5-21`,
    `TestOwnedWall_IdentitySourceRow`, fails outright when `keys.go:24`
    goes; `internal/config/keys_test.go:75`, `TestReservedClassOf`'s
    `DUBPLATE_SMTP_HOST` row, fails when `keys.go:59-73` goes;
    `cmd/dubplate-server/doctor_test.go:243-258`,
    `TestDoctorCmd_LiteralSpotCheck`, asserts the literal `identity
    source` line against `config.NewRegistry()` and fails when the row
    goes; `internal/config/config_test.go:303-316`,
    `TestLoad_RegisteredThresholdIsNotUnknown`, runs against
    `classificationRegistry()` whose `DUBPLATE_CLASSIFY_OWNED` wall row
    (`config_test.go:195`) is unset, and fails at **Task 2** once every
    owned-now row is required; `internal/config/registry_test.go:302-307`,
    `TestCollisionGate_IdentitySourceCarveOut`, goes vacuous rather than
    failing. Also expected:
    `cmd/dubplate-server/serve_test.go:165` and `:238` build through
    `productionRegistry()`, so their environments gain
    `DUBPLATE_IDENTITY_SOURCE`; `check_test.go` and `shadow_test.go` are
    expected to be untouched.)* **Report the exact list and count, and
    treat any mismatch as a blocker for Task 3 rather than a note**,
    because Task 3 cannot reach a green gate outside its declared
    territory. This is Task 2's and Task 3's blast radius, and 5a's one
    planning miss was this same class (`HISTORY:526-529`).
11. **Do both environment templates still list what this pass's
    classification change touches?** *(Expected:
    `env/dubplate.env.template:41-45` lists the five `DUBPLATE_SMTP_*`
    names, which classify today as the `auth and identity` reserved class
    and must classify as the magic-link source's required keys after Task
    3; `:65` lists `DUBPLATE_CONTRIBUTORS_FILE`; `:114-115` lists the two
    Access names; `env/dubplate-server.env.template` lists no rung 5
    name.)* **Report both lists.** **A blocker if the frozen template's
    SMTP names would classify as unknown after the reserved class is
    deleted**, because `TestTemplateKeysClassifiedAgainstProductionRegistry`
    reads that template by path.
12. **Do `productionRegistry` and `classifyKey` still carry the two merge
    carries this pass lands?** *(Expected: `serve_test.go:19-25` restates
    `newRootCmd`'s register calls by hand; `templatekeys_test.go:95-100`
    returns "owned wall" for any wall row including a `Pending` one, where
    `ownedNow` (`keys.go:36-43`) would refuse it.)* **Report both.**
13. **Does `thresholdStatus` still print a `KindString` threshold's
    effective value?** *(Expected: `cmd/dubplate-server/doctor.go:142-147`,
    `config.ParseString` then the value.)* **R-5b6 turns on this
    exactly**, so report the code path in full.
14. **Is `internal/serve`'s `NewStrip` signature still
    `NewStrip(header, sessionCookie string, proxy TrustedProxyConfig,
    exemption FrontExemption)`, and does `cmd/dubplate-server/serve.go`
    still pass `""` for the cookie name?** *(Expected: `strip.go:300` and
    `serve.go:136`.)* Task 6 wires the real name (R-5b9).
15. **Does `internal/atomicfile` still offer both entry points, and does
    `WriteStream` still fsync the containing directory?** *(Expected:
    `WriteFile` at `:16`, `WriteStream` at `:55-91`, the directory open
    and `Sync` at `:82-89`. The spec cites `:45-56` and INV cites
    `:45-55`; both moved.)* **Report the real ranges**, because R11's
    whole argument rests on the second half.
16. **Does `internal/engine/shadow/store.go` still carry the shape this
    pass adopts?** *(Expected: `storeDirName` at `:41`, `StoreDir` at
    `:46`, `OpenStore` at `:53`, `validID` at `:66`, and the 0755/0644
    permissions R10 supersedes for auth state.)* **Report all five.**
17. **Is the state-directory agreement check still hard-coded to
    `systemd/dubplate-shadow.service`?** *(Expected: `scripts/check.sh:153-168`.)*
    R12 generalizes it **at 5d**. **5b-i writes no unit**, so report the
    shape and confirm this pass's arithmetic does not move it.
18. **Do `CLAUDE.md:120`, `scripts/check.sh:5`, and `Makefile:33` still
    carry the "rejoins at rung 7" sentence?** **Pass 5c corrects all
    three, not this pass.** Report the three lines and change none.
19. **Is `_salvage/internal/notify` unchanged, and does
    `_salvage/internal/serve/serve.go` still carry `requireSameOrigin`,
    `mountBasePath`, `staticHandler`, and the single write timeout?**
    Both are **5b-ii's** inputs. **Also confirm that
    `_salvage/internal/auth` does not exist** and report what the nearest
    salvage material actually is, since the prompt naming it is the kind
    of pointer that turns into a phantom adoption.
20. **Has the design-language standard changed since revision 2**, in
    particular R2.1's names, R4.5's rewrite, R5.6's directive set, or the
    ten gates? **5c's**, reported only. `spec:2334-2337`'s pair (does the
    standard still ship no default theme block, and does DL section 7
    still give the roster to rung 7) is **5c's too, named here and
    deferred rather than carried**, because R36 is the ruling that reads
    it.
21. **Does `scripts/deploy.sh` still deliver `config/contributors.yaml` to
    `/etc/dubplate/contributors.yaml`?** R37 leaves that delivery alone, so
    a change here is a signal rather than a blocker.
22. **Is `.gitleaksignore` still the tracked empty file rung 1b left**,
    and has any lane appended to it? **Report the line count without
    quoting any fixture literal.**
23. **Confirm on the worktree's own toolchain, measured rather than
    remembered:** `crypto/subtle.ConstantTimeCompare`'s presence and
    signature; that `os.OpenFile` with `os.O_CREATE|os.O_EXCL` returns an
    error satisfying `errors.Is(err, os.ErrExist)` for an existing file;
    that `(*os.File).Sync` on a directory handle succeeds on this
    filesystem; and the `go` directive in `go.mod` *(expected `go 1.27.0`
    at `go.mod:3`)*. **Report any divergence**, because Task 7's and Task
    8's central rows rest on the second and third.
24. **Is `go test ./...`'s wall time still what this plan sizes gate runs
    against?** Run it once at `-count=1` in the worktree and report the
    total and `bash scripts/check.sh`'s total. *(5a measured `go test
    ./...` at 2m33.785s and the gate near five minutes cold.)*
25. **Does `env/dubplate-server.env.template` still carry the three row
    states in its header, and does `deploy.sh` still abort the render on
    one unresolved name?** (`spec:2325-2327`, restored here because
    **R-5b7 rests its whole argument on both facts** and no other item
    reaches them; item 11 reads what the templates list, not the header.)
    *(Expected: the three row states at
    `env/dubplate-server.env.template:8-16`, with the literal state at
    `:12-13` and the commented state noted as "none below yet";
    `deploy.sh` hard-failing the whole render and listing every name it
    could not resolve.)* **Report both by line.** A change here is a
    blocker for Task 3, because the row state R-5b7 writes into would no
    longer exist.

**Deliverable count: 1** (the blocker report, each item with its evidence
and each blocker with the file and line that proves it). No gate run
beyond item 24's timing measurement; the task writes nothing.

**2.0 hooks: neutral.** The pre-task builds no surface.

**Prior art: none consulted.** The task reads the tree.

**Independent: no.** Every task depends on it.

---

## Task 1: The identity leaf

Model: `sonnet`. Depends on Task 0.

`internal/identity` is born first, because it is the vocabulary Tasks 5
through 9 and all of 5b-ii speak, and because RE cell (5,6) requires the
spec to name the package and forbids `internal/ops` importing
`internal/auth` (R17, `spec:526-532`).

**Files:**
- Create: `internal/identity/identity.go`
- Create: `internal/identity/identity_test.go`

**Interfaces:**
- Consumes: nothing.
- Produces: a verified-identity value carrying the normalized login
  address, the resolved principal, the display name, the name of the
  identity source that asserted it, and the assertion channel; a channel
  type with the three values the seam distinguishes; and **the address
  normalizer every other task calls**. Tasks 5, 6 and 9 consume the
  normalizer; Tasks 6 and 9 consume the value; 5b-ii's seam constructs
  it; 5b-ii's roster and limiters key on the normalizer's output; rung
  6's ops request types carry it.

**Outcomes:**

- The package imports **no other internal package** and no third-party
  package, which is the whole reason it exists (`spec:1467-1469`).
- The verified identity carries five fields, and **the resolved principal
  is its own field, distinct from the login address**. That separation is
  2.0's hook 5: rung 6 substitutes a durable surrogate id in one place
  without touching a call site (`S20:367-369`, `spec:1153-1160`).
- The assertion channel is a named type with exactly three values, cookie,
  header, and service token (`spec:1221-1224`), each with a documented
  meaning, and a `String` method whose output is what an audit record and
  a log line carry.
- The zero value is not a valid identity, and the package says so: a
  caller cannot accidentally treat an unpopulated value as verified.
- **The package owns address normalization**, because `spec:1161-1163`
  requires it "before any keying, any logging, and any rate-limit lookup"
  and `internal/identity` is the one leaf every consumer already imports
  (review M-a). One function parses the value as a single RFC 5322
  address with `net/mail`, lowercases it, and refuses a value above a
  stated length cap, the cap written out as a literal with its reason.
  `net/mail` is stdlib, so the no-internal-import rule is untouched, and
  `internal/config/config.go:162-173` is the in-tree idiom being
  followed. Revision 1 assumed a caller had normalized, which left the
  token minter and the session minter accepting whatever they were given
  while only the epoch store normalized, and a divergence between the
  address in a MAC and the address in a digest is a silent revocation
  miss rather than an error.
- **The normalizer's output is what the identity's address field
  carries**, so a populated identity value cannot hold an unnormalized
  address.
- The doc comments state what the package is for and what it must never
  gain, so the next rung reading it learns the dependency rule from the
  package rather than from the spec.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The import block contains no `github.com/glw907/dubplate/internal/...`
  path and no third-party path, provable by reading it in full.
- Five fields, named and documented, with the principal separate from the
  address.
- Three channel values, no more, each documented; `String` covers all
  three and the test enumerates them by literal rather than by ranging
  over the type's own list.
- A test proves the zero value is refused by whatever check the package
  offers for it, with the expectation written out rather than derived from
  the code under test.
- Every fixture address is documentation-reserved or invented.
- The normalizer's rows cover a case difference collapsing to one value,
  the length cap at the boundary and one byte over, an unparseable value,
  and an address carrying a `.` in its local part surviving unchanged
  apart from case. Each expectation is written out rather than derived
  from the code under test, and the cap literal appears in the test.
- A test asserts that the address field of a constructed identity equals
  the normalizer's output for the same input, so the two cannot drift.
- Doc comments carry no em dash. `bash scripts/check.sh` green in the
  worktree.

**Named mutations this task reports:**

1. Fold the principal into the address field → the hook-5 test fails,
   because the two values are asserted to differ in a row where a member's
   principal is not their address.
2. Add a fourth channel value → the enumeration test fails by name.
3. Make the zero value pass the validity check → the zero-value row
   fails.
4. Lowercase without parsing → the unparseable-value row fails, reporting
   an address accepted that `net/mail` refuses.
5. Compare the cap against the rune count rather than the byte length →
   the boundary row fails on a multi-byte fixture.

**2.0 hook: lands hook 5's first half** (`spec:1848`, `S20:367-369`). The
resolved principal is a field of its own, which is the single place rung 6
substitutes a surrogate id, and the type is what 2.0's credential store
reads.

**Prior art:** none consulted for the type itself. **Declines**
`golang-jwt/jwt/v5`'s claims vocabulary as a model (MIT, v5.3.1 published
2026-01-28, 15,146 importers, PRA:114-140): the claim set is designed for
a token crossing a trust boundary, and this value never leaves the
process. Gitea's server-side session record is recorded as the road not
taken (PRA:568-578).

**Independent: yes**, by files and by imports. It runs first because
everything downstream names it.

**Deliverable count: 3** (the type and its channel, the address
normalizer, the tests). Gate green; commit.

---

## Task 2: The config core's requirement mechanism and the per-source declaration family

Model: **`opus`**. Depends on Task 0.

This is the pass's contract change, and it lands with **no rung 5 key
name anywhere in `internal/config`**. Everything it builds is proved
against a fixture source the config package's own tests declare.

**Files:**
- Modify: `internal/config/config.go` (the package doc's class list, the
  `Config` type and its accessor, `rejectNamespace`, `Load`)
- Modify: `internal/config/registry.go` (one declaration family and its
  accessor)
- Modify: `internal/config/config_test.go`,
  `internal/config/registry_test.go`
- Modify: `cmd/dubplate-server/doctor.go` (one row loop),
  `cmd/dubplate-server/doctor_test.go`

**Interfaces:**
- Consumes: nothing from this pass.
- Produces: a registry entry point that declares an identity source (the
  selector key, the source name, the required keys, the optional keys, and
  a validation hook **taking the source's resolved values and the
  environment snapshot in `os.Environ()` form**), an accessor listing
  declared sources, a `Config` field carrying the active source's name,
  and a `Config` accessor returning a resolved per-source value by key.
  Task 3 is the first caller and Task 4 writes the first hook.

**Outcomes (A1, A2):**

- **`Load` requires every owned-now wall row the registry carries**, and a
  row carrying `Pending` is neither required nor read. Rung 1's four rows
  behave exactly as before, because they are the rows `NewRegistry` seeds.
  A missing row is reported once, never twice, and joined with every other
  failure.
- A rung declares an identity source from its own file. The declaration
  names the selector key, so `internal/config` names nothing rung-specific.
- **With no source declared, the selector is not required**, which is what
  keeps `NewRegistry`-only callers and `internal/config`'s own tests
  working unchanged.
- With one or more sources declared, `Load` requires the selector, refuses
  a value naming no declared source with an error listing the declared
  names, requires the selected source's required keys, reads its optional
  keys, and runs the selected source's validation hook.
- **The hook takes the source's resolved values and the environment
  snapshot** (**conductor-owed**, review B2). Without the snapshot a
  rung cannot validate a registered threshold's value, or a relation
  between two of them, from its own file: a threshold resolves at its
  point of use through `config.ParseDuration`
  (`internal/config/thresholds.go:91-93`), `Load` reads no threshold
  value (`config.go:248-287`), and a name cannot be declared as both a
  threshold and a source key because `Registry.claim`
  (`registry.go:119-124`) panics across families. R31's floor is the
  validation that needs it, and it is one of this pass's own exit-gate
  assertions.
- **The hook's errors join the rest**, so a deployment with a short secret
  and a missing SMTP host learns both at once.
- **`rejectNamespace` gains a sixth classification**: a name belonging to
  any declared source's required or optional set is a declared name, so
  setting it is never reported unknown. The package doc's numbered class
  list (`config.go:6-42`) gains the source class as its sixth entry.
  **The doc edit also reconciles the two countings the tree keeps apart**
  (review m15): `thresholds.go:104-111` calls `thresholdOf` "the fifth
  classification the wall recognizes" while the package doc's numbered
  five never carries thresholds as a class of their own. One sentence in
  the package doc says where a registered threshold sits relative to the
  numbered list, so the two readings agree by inspection.
- `Config` carries the active source's name and answers a lookup for a
  resolved per-source value by key, on `IntegrationValue`'s precedent
  (`config.go:66-118`). **No per-source field is added**, so a second
  source needs no second field.
- **`doctor` prints one row per declared source**, reporting active,
  inactive, or unsatisfied, deriving the row from the registry the way
  every other row derives.
- **A registration collision panics** through the registry's existing
  `claim`, so a key declared in two families is a loud programmer error at
  registration (`registry.go:95`, `:119`).

**Acceptance criteria a diff-reviewer verifies from the diff:**

- No `DUBPLATE_IDENTITY_SOURCE`, no `DUBPLATE_MAGIC_LINK_*`, no
  `DUBPLATE_SMTP_*`, and no `DUBPLATE_ROSTER_FILE` literal appears
  anywhere under `internal/config`, provable by grep, which is the rung 1b
  contract's own test.
- Every test in this task declares a **fixture** source on scratch
  `DUBPLATE_TEST_SCRATCH_*` names, on
  `TestDoctorCmd_LineCountFollowsRegistry`'s own precedent
  (`doctor_test.go:151-175`).
- A row registered with `Pending` set is still neither required nor read,
  proved by a test that registers one and asserts `Load` succeeds with it
  unset.
- A wall row registered without `Pending` is required, proved by a test
  that registers one, leaves it unset, and asserts `Load` names it exactly
  once in the joined error.
- The selector is unrequired with zero sources declared and required with
  one, proved by two rows of the same test rather than by two tests that
  could drift.
- A selector value naming no declared source fails with an error naming
  the declared set, and the expectation is written out rather than built
  from the registry under test.
- A declared source's required key missing fails; its optional key missing
  succeeds; its optional key set is readable from `Config`.
- **The hook can validate a registered threshold's value**, proved
  against a fixture: a threshold registered on a
  `DUBPLATE_TEST_SCRATCH_*` name, a fixture hook that refuses one value
  of it by reading the environment snapshot, and an assertion that the
  refusal joins `Load`'s other errors rather than replacing them. This is
  what makes Task 4's R31 check reachable.
- **`classificationRegistry`'s own environment satisfies its owned-now
  wall row.** `internal/config/config_test.go:195` registers
  `DUBPLATE_CLASSIFY_OWNED` and `:303-316` asserts `Load` returns nil
  against it, so requiring every owned-now row breaks that test unless
  its environment gains the row. The task reports the change with its
  reason.
- A name in a declared source's set is accepted by `rejectNamespace`,
  proved by a test that sets one and asserts `Load` reports no unknown-key
  error for it.
- **`TestDoctorCmd_LineCountFollowsRegistry` gains a fifth family**,
  proving that declaring one source costs exactly one `doctor` line
  printed in its own block. Without this the new family passes both
  existing assertions vacuously, because both run against a fresh
  registry that declares no source.
- The package doc comment lists six classes and the sixth is described in
  the same register as the other five.
- No existing test's expectation is weakened, and the task reports the
  full list of tests it touched with the reason for each.
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Keep `Load`'s four hard-coded requires and add nothing → the
   registered-wall-row row fails, reporting a row registered and not
   required, which is risk 11's exact shape.
2. Require a `Pending` row → the pending row's test fails, and every
   existing `Load` test that satisfies four rows fails with a demand for a
   variable literally spelled `identity source`, which is the trap
   `spec:344-346` names.
3. Require the selector with zero sources declared → `internal/config`'s
   own tests and every `NewRegistry`-based command test fail.
4. Skip the validation hook when the required keys are all present → the
   hook row fails, reporting a configuration accepted that the source
   refuses.
5. Pass the resolved values alone and drop the environment snapshot →
   the fixture-threshold row fails to compile, which is the point: the
   signature is what makes R31's check expressible.
6. Report a source-set name as unknown → the sixth-classification row
   fails, and the frozen template's SMTP names would fail Task 3's
   template test.
7. Print the resolved value in `doctor`'s per-source row → the row's own
   assertion fails, and the reason is R-5b6's: a source's required set
   carries secrets.

**2.0 hook: neutral.** The config core carries no 2.0 surface. It keeps
hook 6 open by naming no driver.

**Prior art:** none consulted. The mechanism is the repo's own registry,
and rung 1b's contract (`HISTORY:1491-1507`) is the authority.

**Independent: yes**, by files. It touches no package Task 1 writes and
consumes nothing from it. It runs early because Task 3 is its first
caller.

**Deliverable count: 4** (the requirement mechanism, the source family
with its hook, `Config`'s field and accessor, `doctor`'s row). **This is
one of the two tasks to watch**; if its review returns `fix` twice, the
conductor's split point is between the requirement mechanism and the
source family.

---

## Task 3: Rung 5's declarations, and both `keys.go` retirements

Model: `sonnet`. Depends on Tasks 0 and 2.

The rung 1b contract's removal half, exercised for a reserved class and,
for the first time anywhere, for a pending wall row.

**Files:**
- Create: `cmd/dubplate-server/auth.go`
- Create: `cmd/dubplate-server/auth_test.go`
- Modify: `internal/config/keys.go` (two deletions, three message
  rewrites)
- Modify: `internal/config/keys_test.go` (the two tests the deletions
  break), `cmd/dubplate-server/doctor_test.go` (the literal spot check's
  identity-source line), `internal/config/registry_test.go` (the
  carve-out test the deletions make vacuous)
- Modify: `internal/config/config.go` (the `Config` doc comment's pending
  sentence and the package doc's owned-row paragraph)
- Modify: `cmd/dubplate-server/main.go` (the shared registry constructor
  and its fourth register call)
- Modify: `cmd/dubplate-server/serve_test.go` (`productionRegistry` and
  its environments), `cmd/dubplate-server/templatekeys_test.go`
  (`classifyKey`)
- Modify: `env/dubplate-server.env.template` (one literal row)

**Interfaces:**
- Consumes: Task 2's source declaration entry point.
- Produces: `registerAuthConfig(reg)`, called from the binary's shared
  registry constructor; the twelve declared names; and the constructor
  itself, which `newRootCmd` and `productionRegistry` both call.

**Outcomes (A3):**

- `DUBPLATE_IDENTITY_SOURCE` is registered as an owned wall row with no
  `Pending`, and **the `{Name: "identity source", Pending: ...}` row is
  deleted from `keys.go` in the same commit** (`keys.go:24`, R5).
- The magic-link source is registered with its nine required names and its
  two optional names, and **the `auth and identity` reserved class is
  deleted from `keys.go` in the same commit** (`keys.go:60-73`).
- The three supersession messages are rewritten:
  `DUBPLATE_CONTRIBUTORS_FILE` names `DUBPLATE_ROSTER_FILE` rather than
  rung 6's store, and the two Access rows say that the seam admits a
  forward-auth front and that its configuration arrives with the rung that
  builds one (`spec:1620-1628`). **No supersession entry is deleted.**
- **`env/dubplate-server.env.template` gains exactly one row**,
  `DUBPLATE_IDENTITY_SOURCE=magic-link`, in the template's literal row
  state, with a comment stating why it is a literal and not a store
  lookup (R-5b7).
- **The binary has one registry constructor**, called by `newRootCmd` and
  by the test helper, so a fourth register call is added once beside
  `registerShadowIntegration`, `registerShadowThresholds` and
  `registerServeConfig` (`main.go:84-86`, merge carry 2).
- **`classifyKey` no longer reports a `Pending` wall row as owned**, and
  it gains a branch for a declared source's key set (merge carry 1, and
  the sixth classification's test-side counterpart).
- `internal/config`'s two doc comments stop describing a pending fifth
  wall row that no longer exists.
- **The keys.go diff is two deletions and three message rewrites, and
  nothing else.**

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The `keys.go` diff contains no addition other than replacement message
  text, provable by reading it in full. **No name is added to `keys.go`**,
  which is the contract's other half.
- Both templates classify with no unknown name, asserted by
  `TestTemplateKeysClassifiedAgainstProductionRegistry` over
  `env/dubplate.env.template` and `env/dubplate-server.env.template`.
  **The frozen template's five `DUBPLATE_SMTP_*` names now classify as
  the magic-link source's required keys**, where they classified as a
  reserved class before, and a test row names that transition
  explicitly.
- `TestTemplateKeysCoverEveryOwnedNowRow` passes with the new owned-now
  row, and the task reports the row's line in the template.
- `doctor` exits nonzero and names `DUBPLATE_IDENTITY_SOURCE` when it is
  unset against the production registry. **This is `spec:2168`'s exit-gate
  assertion, and it lands in this task.**
- `doctor` exits zero against an environment carrying the selector and the
  nine required keys, and the task reports the full row list once.
- Setting `DUBPLATE_MAGIC_LINK_SECRET` with the selector unset is a load
  error naming the selector rather than an unknown-key error.
- `classifyKey` returns empty for a hypothetical `Pending` row, proved by
  a test that registers one, and the test states in a comment that no
  production registry carries a pending row after this task.
- **`TestOwnedWall_IdentitySourceRow` (`internal/config/keys_test.go:5-21`)
  is deleted rather than weakened**, because the row it asserts no longer
  exists and a test rewritten to assert its absence would pin nothing.
  The task reports the deletion with that reason.
- **`TestReservedClassOf`'s `DUBPLATE_SMTP_HOST` row
  (`internal/config/keys_test.go:75`) moves to the magic-link source's
  required set**, asserted through the same classification path the
  template test uses, with the transition named in the row.
- **`TestDoctorCmd_LiteralSpotCheck`'s identity-source line
  (`cmd/dubplate-server/doctor_test.go:255-258`) is replaced by the
  owned-now selector row printed against the production registry**, so
  the spot check keeps a hard-coded literal to check and does not lose
  its anti-vacuity purpose (`doctor_test.go:240-242`). The
  `DUBPLATE_ADMIN_EMAIL` line is untouched.
- **`TestCollisionGate_IdentitySourceCarveOut`
  (`internal/config/registry_test.go:302-307`) registers its own fixture
  status row** rather than reading the production table, so the carve-out
  it proves stays proved after the last production pending row is gone.
- `productionRegistry` calls the shared constructor rather than restating
  register calls, and `newRootCmd` calls the same constructor.
- Every environment in `serve_test.go` that satisfies the wall gains the
  selector, and the task reports the count against Task 0 item 10's
  prediction.
- No secret value appears in any fixture; every fixture secret is a
  visibly synthetic literal.
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Declare the twelve names without deleting the reserved class →
   `Registry.claim` panics at registration on the eight colliding keys
   (`registry.go:95`, `:119`), and the registry test reports it.
2. Delete the reserved class without declaring the source → the frozen
   template's five SMTP names classify as unknown and
   `TestTemplateKeysClassifiedAgainstProductionRegistry` fails, naming
   each.
3. Clear the pending row's `Pending` field instead of deleting the row →
   `Load` demands a variable literally spelled `identity source`, and
   every wall test fails with that name in the error.
4. Register the selector and skip the template row →
   `TestTemplateKeysCoverEveryOwnedNowRow` fails, naming the row. **This
   is R-5b7's whole argument, fired.**
5. Rewrite `DUBPLATE_CONTRIBUTORS_FILE`'s message by deleting the entry →
   the frozen template's `:65` name classifies as unknown.
6. Leave `productionRegistry` restating the register calls → nothing
   fails, which the task reports as an unfired mutation with its reason:
   the carry is a maintenance defect rather than a behavioral one, and the
   proof is that a fourth call was added in one place instead of two.
7. Leave `TestCollisionGate_IdentitySourceCarveOut` reading the
   production table → nothing fails, and that is the defect: with no
   pending row left, the test passes without exercising the carve-out at
   all. The task reports it as the vacuity the fixture row closes, which
   is the no-vacuous-assertions rule fired against a test rather than
   against production code.

**2.0 hook: neutral.** Configuration names no 2.0 surface. The roster
key's declaration is what 5b-ii's store reads, and hook 5's write half
lands there.

**Prior art:** none consulted. The registry contract is the authority.

**Independent: no.** It consumes Task 2's declaration family.

**Deliverable count: 4** (the declarations file, the two `keys.go`
retirements with the three message rewrites, the template row, the two
merge carries). **This is the second task to watch**, because its blast
radius crosses two templates, six test files, and the binary's wiring.
Task 0 item 10 measures that radius before dispatch and a mismatch is a
blocker rather than a note.

---

## Task 4: The magic-link thresholds and the source's load-time validation

Model: `sonnet`. Depends on Tasks 0, 2, and 3.

**Files:**
- Modify: `cmd/dubplate-server/auth.go`, `cmd/dubplate-server/auth_test.go`

**Interfaces:**
- Consumes: Task 2's validation hook, **which takes the source's resolved
  values and the environment snapshot**, so this task's hook can reach a
  registered threshold's value; Task 3's source registration.
- Produces: the four thresholds by key, and the magic-link source's
  validation hook. Tasks 5, 6, 8 and 9 read the resolved values through a
  resolver this task also produces, on `resolveServeConfig`'s precedent
  (`cmd/dubplate-server/serve.go:81-106`).

**Outcomes (A7, R-5b8, R-5b10):**

- Four thresholds are declared with the keys, kinds, and defaults R-5b10
  rules, each with a `Doc` naming the accepted form the way rung 4's and
  5a's do.
- The validation hook refuses, **naming the key in each case**: a
  `DUBPLATE_MAGIC_LINK_SECRET` or `DUBPLATE_SESSION_SECRET` shorter than
  32 bytes; the same for either `_PREVIOUS` value when it is set; a
  `DUBPLATE_PUBLIC_URL` failing any of three conditions; and a
  `DUBPLATE_NONCE_RETENTION` below `DUBPLATE_MAGIC_LINK_EXPIRY` (R31),
  read from the environment snapshot because both are thresholds rather
  than source keys.
- **The URL check is three conditions, not two** (review m7). The value
  parses; it carries a non-empty host, because `net/url` accepts a bare
  hostname with `err == nil` and an empty `Host`, and a URL with no host
  builds a path-only link that defeats security review finding 7; and its
  scheme is `https` unless its host is loopback. **The comparison reads
  the parsed hostname rather than the host**, since
  `url.Parse("http://localhost:8080").Host` is `localhost:8080` and only
  `Hostname()` is `localhost`.
- **The loopback carve-out covers the whole loopback set, not the name
  `localhost` alone** (**conductor-owed**, review m6). R38's reason is the
  browser rule, and the W3C Secure Contexts definition of a potentially
  trustworthy origin covers `127.0.0.0/8` and `::1` as well as the name,
  so refusing a loopback-IP adopter refuses them for a reason that does
  not apply. The check accepts a hostname equal to `localhost` or one
  whose parsed address reports loopback, and the task cites the W3C
  definition by URL and read date. The close drafts the correction to
  R38 (`spec:758-767`).
- **Every refusal joins the rest of `Load`'s errors**, so `check` and
  `doctor` report a short secret and a bad URL together.
- **No error text carries a secret value.** A short secret is reported by
  its key and by the floor, never by its content or its length in a form
  that narrows a guess.
- One resolver returns the values the auth core takes as plain
  construction arguments, so `internal/auth` never reads the environment.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Each threshold's default is the literal R-5b10 names, and the test
  asserts each by literal rather than by reading the declaration.
- A secret one byte below the floor fails and one byte at the floor
  passes, both rows present, with the boundary stated as inclusive in the
  doc comment.
- `https://example.test` passes; `http://example.test` fails;
  `http://localhost:8080` passes; `http://127.0.0.1:8080` passes;
  `http://192.168.1.10:8080` fails; `example.test` fails for an empty
  host; `://bad` fails for the parse, which is a value `net/url`
  measurably refuses where `not a url` and `example.test` both parse
  cleanly; each with the key named.
- A retention equal to the link expiry passes and one below it fails,
  proving R31's floor is inclusive.
- A test asserts that no error string produced by the hook contains any
  fixture secret's value, by searching the joined error text for each
  fixture literal.
- Every fixture secret is a visibly synthetic literal of at least 32
  bytes, and no fixture is shaped like a real key. **A visibly synthetic
  value is what keeps gitleaks quiet here**, since the rule needs both a
  key-like identifier and a high-entropy value (Global Constraints).
- `bash scripts/check.sh` green in the worktree, gitleaks included.

**Named mutations this task reports:**

1. Compare the secret's rune count instead of its byte length → a
   multi-byte fixture row fails, reporting a short key accepted.
2. Make the floor exclusive at 32 → the at-the-floor row fails.
3. Allow `http` on any host → the plain-HTTP row fails, and the reason is
   R38's: a browser refuses a `Secure` cookie over plain HTTP on any
   origin but a potentially trustworthy one, so the login would complete
   and sign nobody in.
4. Compare `url.Host` rather than `url.Hostname()` → the
   `http://localhost:8080` row fails, because the port is part of the
   host.
5. Drop the non-empty-host condition → the `example.test` row fails,
   reporting a configured origin accepted that builds no origin at all.
6. Compare retention and expiry with the inequality reversed → the equal
   row and the below row both fail.
7. Include the offending value in the error → the no-secret-in-errors row
   fails.

**2.0 hook: neutral.** Thresholds carry no 2.0 surface.

**Prior art:** RFC 2104 section 3 is cited by URL and read date for the
key-length recommendation behind R-5b8. No library consulted.

**Independent: no.** It consumes Tasks 2 and 3.

**Deliverable count: 3** (the four thresholds, the validation hook, the
resolver). Gate green; commit.

---

## Task 5: The magic-link token

Model: **`opus`**. Depends on Tasks 0, 1, and 4.

The pass's first cryptographic construction, and the one where a subtly
wrong reading produces a token that verifies every test and forges under
an attacker who knows the format.

**Files:**
- Create: `internal/auth/token.go`
- Create: `internal/auth/token_test.go`
- Create: `internal/auth/doc.go`

**Interfaces:**
- Consumes: Task 1's identity vocabulary **and its address normalizer**,
  which the minter runs the caller's address through rather than
  assuming a caller did; Task 4's resolved values as plain construction
  arguments.
- Produces: a minter taking the current secret and the link expiry, a
  verifier taking the current and previous secrets, the token's parsed
  fields including its mint time, the sentinels a caller branches on, and
  the link-URL builder that takes the configured public origin. Task 8
  consumes the nonce; Task 9 consumes the mint time; 5b-ii's consumption
  handler consumes all of it.

**Outcomes (A4):**

- The token carries the normalized login address, an expiry, a mint time,
  and a 128-bit nonce from `crypto/rand`. **The minter normalizes through
  Task 1's function**, so the address a MAC covers is the address Task 9
  digests.
- **Every field is length-prefixed in the serialization**, the MAC input
  included, so no value can be shifted across a boundary by a separator
  inside it. That is `spec:1017-1019`'s claim about the serialization
  rather than about the MAC input alone. The doc comment states the
  attack this defeats.
- **The wire encoding is stated: each field is base64url-encoded before
  any separator is applied** (**conductor-owed**, review M-b), so a field
  value can never carry the separator. Without this the obvious shape,
  joining fields with `.`, breaks every address whose local part carries a
  dot, which is the majority of real addresses; the MAC then fails and the
  person cannot log in, with every ordinary test green.
- **The token carries no destination field** (R33, `spec:690-692`), and
  the version label exists so a later field can be added without changing
  the meaning of an existing one. Rung 12 introduces a destination if it
  wants one, inside the MAC-covered token rather than as a free query
  parameter.
- **A version label and a domain-separation label are inside the MAC
  input**, both package constants, each with its reason documented. The
  domain-separation label differs from the session cookie's, so a session
  value presented as a link value fails the MAC.
- The MAC is HMAC-SHA256 and the comparison is
  `crypto/subtle.ConstantTimeCompare`.
- **Verification order is MAC, then expiry, and the nonce is returned for
  the caller to insert afterwards.** The package's own API makes the wrong
  order awkward: the verifier returns the nonce only on a token that has
  already passed the MAC and the expiry.
- `DUBPLATE_MAGIC_LINK_SECRET_PREVIOUS` verifies and never mints. A token
  that verifies under the previous secret alone is accepted, and the
  minter has no path to it.
- **The link URL is built from the configured public origin**, never from
  a request value, and the token rides in the query string (R-5b13,
  security review finding 7).
- Distinct sentinels for a failed MAC, an expired token, and a malformed
  token, so a caller branches with `errors.Is` and the one generic
  user-facing outcome is the caller's choice rather than the package's
  (`spec:1043-1044`).
- **No error string and no log record carries the token, the nonce, or
  the secret.**

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The package imports no `internal/config` path, provable by reading the
  import block.
- **A checked-in golden vector**: a fixed secret, a fixed address, a fixed
  expiry, a fixed mint time, and a fixed nonce, with the expected encoded
  token written out as a literal. The test compares against the literal
  rather than recomputing the MAC with the same function, which is the
  assertion-discipline rule this pass leans on hardest. **The identifier
  beside every such literal must not read as a key name**, because the
  literal carries a MAC and therefore clears gitleaks' entropy threshold;
  measured, `goldenToken` fires and `goldenVector` does not (Global
  Constraints).
- **A golden vector for an address carrying the encoding's separator**,
  written out as a literal, round-tripping through mint and verify. This
  is the row that fails under a separator-joined encoding.
- **A golden vector for an address carrying `+` and `=`**, written out as
  a literal, proving the encoding and the query both survive them.
  Measured, `url.JoinPath` plus `Query().Encode()` renders
  `?t=AAAA%2B%2F%3Dabc`, so base64 padding and slashes survive the link.
- The minter's signature offers no destination parameter, provable by
  reading it, and the doc comment records that the version label is where
  a later field would land.
- A field-shifting row: two different field triples that would serialize
  identically without length prefixes are asserted to produce different
  MACs, and the row states which shift it defeats.
- A cross-purpose row: a value minted under the session's
  domain-separation label fails the link verifier, and the row names the
  label rather than asserting a bare failure.
- A one-bit flip in the MAC fails; a one-bit flip in the address fails;
  each with the sentinel asserted through `errors.Is`.
- An expired token fails with the expiry sentinel **after** the MAC check,
  proved by a token that is both expired and MAC-broken reporting the MAC
  sentinel.
- The verifier returns no nonce for a token that fails either check,
  proved by a row that asserts the returned nonce is the zero value.
- A token minted under the current secret and one minted under the
  previous secret both verify; the minter's signature offers no previous
  secret at all.
- The link URL for a configured origin `https://example.test/` is written
  out as a literal, carries the token in the query, and changes when the
  configured origin changes.
- A test searches every error string the package produces for each
  fixture's token, nonce, and secret literal, and asserts none appears.
- Every fixture secret is a visibly synthetic literal at or above the
  floor; every address is documentation-reserved or invented.
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Concatenate fields without length prefixes → the field-shifting row
   fails.
2. Join the fields with a separator without encoding each field first →
   the separator-bearing-address golden row fails, reporting a token that
   cannot be parsed back for the majority of real addresses.
3. Drop the domain-separation label from the MAC input → the
   cross-purpose row fails once Task 6 lands, and this task reports the
   mutation as one whose blast radius spans two tasks.
4. Compare MACs with `==` or `bytes.Equal` → nothing functional fails,
   which the task reports as an **unfired mutation with its reason**: a
   timing property has no gate-resident test, and the criterion is the
   reviewer's reading of the comparison call. The doc comment states why
   the constant-time call is there.
5. Check the expiry before the MAC → the both-broken row fails, reporting
   the expiry sentinel where the MAC sentinel is wanted.
6. Return the nonce on a failed verification → the zero-value row fails,
   and Task 8's fixed-order row fails too.
7. Build the link from a request host rather than the configured origin →
   the URL row fails; the criterion also holds security review finding 7
   for the whole pass.

**2.0 hook: neutral.** The token carries the login address, and 2.0 reads
neither the token nor its fields. The mint time it carries is what Task 9
compares against the epoch, which is where hook 5 lands.

**Prior art:** **adopts** stdlib `crypto/hmac`, `crypto/sha256`,
`crypto/rand`, and `crypto/subtle` (stdlib, exemplar quality; the birth
requirements specify the construction down to the label, which is under
fifty lines and no supply-chain edge; PRA:105-113, `spec:1871`).
**Declines `golang-jwt/jwt/v5`** for minting (MIT, v5.3.1 published
2026-01-28, 15,146 importers, good posture on `alg` confusion;
PRA:114-140): algorithm agility is a liability for a token minted and
consumed by one process, and R26 holds the dependency for the commit that
builds a Cloudflare Access front. **Declines `o1egl/paseto`** (MIT,
v1.0.0 published 2018-11-08, its own README deprecating protocol v1;
PRA:141-155): fails maintenance outright. **Declines
`aidanwoods.dev/go-paseto`** (MIT, v1.6.0 published 2025-12-27, 235
importers, genuinely well made; PRA:156-175): PASETO's value is
interoperability across a trust boundary this token never crosses.

**Independent: no.** It consumes Task 4's resolved values and Task 1's
vocabulary.

**Deliverable count: 3** (the token, the link builder, the tests). Gate
green; commit.

---

## Task 6: The session cookie

Model: **`opus`**. Depends on Tasks 0, 1, 4, and 5.

**Files:**
- Create: `internal/auth/session.go`
- Create: `internal/auth/session_test.go`
- Modify: `cmd/dubplate-server/serve.go` (the cookie name passed to
  `NewStrip`), `cmd/dubplate-server/serve_test.go`

**Interfaces:**
- Consumes: Task 1's identity vocabulary **and its address normalizer**,
  Task 4's lifetime and cap, Task 5's label convention.
- Produces: the cookie name constant, a minter, a verifier taking current
  and previous secrets, a refresher, the parsed session's issue time and
  anchor, and the sentinels. Task 9 consumes the issue time; 5b-ii's
  session middleware consumes the rest.

**Outcomes (A5):**

- One HMAC-signed cookie named `__Host-dubplate_session` (R-5b9),
  carrying the normalized address, the issue time, an anchor time, and an
  expiry, under its own domain-separation label.
- Attributes are `Secure`, `HttpOnly`, `SameSite=Lax`, `Path=/`, and **no
  `Domain`**, which the `__Host-` prefix requires and which is why
  same-origin is a stated precondition of the SSO capability
  (`spec:1076-1082`).
- **Rolling refresh moves the expiry and never the anchor.** A session
  refreshed any number of times is refused once the anchor plus the
  absolute cap has passed.
- **Refresh moves the issue time, so the anchor is the session's original
  issue time**, documented on the type in exactly those words. Every
  check that must reason about when the session was born reads the
  anchor, and Task 9's epoch check is the one that would otherwise read a
  field refresh advances (review B1).
- **Refresh counts same-origin requests only** (`spec:1081-1082`),
  documented on the refresher as a stated precondition of its contract.
  5b-ii's middleware is the caller that enforces it, and the handoff
  carries the rule so it is inherited rather than rediscovered. Without
  it a cross-origin subresource load holds a session open to its 180-day
  cap on a machine its owner has walked away from, which is the
  shared-device adversary the threat model names.
- **The address the session carries is the normalizer's output** (Task
  1), so the value a MAC covers and the value the epoch store digests
  cannot diverge.
- `DUBPLATE_SESSION_SECRET_PREVIOUS` verifies and never mints.
- The value's serialization follows the token's rule exactly: length
  prefixes and both labels inside the MAC input, with the
  domain-separation label distinct from the link's.
- **The serve verb constructs the shell with the real cookie name**
  (`cmd/dubplate-server/serve.go:136`), which retires 5a's declared
  coverage gap 4.
- No error string and no log record carries the cookie value or the
  secret.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The cookie name is the literal R-5b9 rules, asserted by a test that
  writes it out, and the same constant is what the serve verb passes.
- Every attribute is asserted on a written cookie: `Secure` true,
  `HttpOnly` true, `SameSite=Lax`, `Path=/`, and **`Domain` empty**, the
  last with a comment naming the `__Host-` rule.
- **A refresh row that refreshes repeatedly across the whole lifetime and
  asserts the anchor is byte-identical each time**, with the expected
  anchor written out rather than read from the value under test.
- A session whose anchor plus the cap has passed is refused with the cap
  sentinel, even when its own expiry is in the future. That combination is
  the row that proves the cap is real.
- A session past its expiry but inside its cap is refused with the expiry
  sentinel.
- A link token presented as a session value fails, and a session value
  presented as a link token fails, both asserted through the sentinels
  and both naming the labels in the row.
- A session minted under the previous secret verifies; the minter offers
  no previous secret.
- A test searches every error string for the fixture cookie values and
  secrets and asserts none appears.
- `serve_test.go` asserts the shell is constructed with the cookie name
  rather than the empty string.
- **A behavioural row closes 5a's gap 4** (review m8): a request carrying
  `__Host-dubplate_session` alongside another cookie reaches the
  fallthrough with the session cookie absent and the other cookie
  readable, run against the shell as the serve verb constructs it. The
  gap's property is behavioural (`HISTORY:340-341`,
  `internal/serve/strip.go:293-296`), and 5a proved the removal for an
  arbitrary name; this row proves it for the real one. The row lives in
  `serve_test.go`, so no `internal/serve` file is opened (R-5b15).
- **A row proving refresh moves the issue time and not the anchor
  separately**, with both fields read out and the anchor asserted against
  a written-out literal, so the two cannot be confused by a reader or by
  Task 9.
- Every checked-in cookie literal carries an identifier that does not
  read as a key name, since a session value carries a MAC and clears
  gitleaks' entropy threshold (Global Constraints).
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Move the anchor on refresh → the repeated-refresh row fails, and the
   task reports it as the mutation whose absence silently removes the
   absolute cap (`spec:2370-2376`, risk 4).
2. Drop `Secure` or set a `Domain` → the attribute rows fail by name, and
   the reason is that a browser refuses a `__Host-` cookie carrying
   either.
3. Reuse the link's domain-separation label → the two cross-purpose rows
   fail.
4. Check the cap against the issue time rather than the anchor → the
   refreshed-past-the-cap row fails.
5. Leave refresh moving the anchor as well as the issue time → the
   repeated-refresh row fails, which is mutation 1 read from the other
   side and the reason the two fields are asserted separately.
6. Leave the serve verb passing `""` → `serve_test.go`'s construction row
   fails and the behavioural fallthrough row fails, so 5a's gap 4 stays
   open on both counts.

**2.0 hook: neutral, with a recorded forward.** The session keys on the
normalized address, and rung 6's re-key onto the surrogate id changes the
seam rather than the cookie's shape. Task 9's epoch carries the same
obligation, and R4 records it as rung 6's rather than leaving it to be
discovered.

**Prior art:** **adopts** stdlib `net/http` cookies (stdlib; the design is
a stateless HMAC cookie plus a per-address epoch file, so there is nothing
for a session library to manage; PRA:46-54, `spec:1870`). **Declines
`gorilla/sessions`** (BSD-3, newest release v1.4.0 dated 2024-08-20, 8,382
importers; PRA:55-73): its value is hiding the cookie serialization this
rung must state exactly. **Declines `alexedwards/scs` on shape** (MIT,
v2.9.0, stated stable and maintained, 3,701 importers; PRA:75-93): it
manages a server-side store, and rung 5 has none by design. **Records
Gitea's server-side store as the road not taken** (PRA:568-578): a
server-side store is what you build when sessions must be listable and
individually revocable, and rung 5 buys per-person revocation with an
epoch file at a fraction of the machinery.

**Independent: no.** It consumes Task 5's label convention and Task 4's
values.

**Deliverable count: 3** (the session value, the cookie-name wiring, the
tests). Gate green; commit.

---

## Task 7: The state-directory leaf, the naming rule, and the confinement argument

Model: `sonnet`. Depends on Task 0. Revision 1 declared a dependency on
Task 4 that its own Interfaces line, its independence mark, and the
execution model's dependency list all contradict: the state directory is
rung 1's `Config.StateDir` (`internal/config/config.go:75`), not one of
Task 4's four thresholds (review m5).

The substrate both records sit on, born once so the nonce store and the
epoch store cannot disagree about where a file goes or what a name may
contain.

**Files:**
- Create: `internal/auth/store.go`
- Create: `internal/auth/store_test.go`

**Interfaces:**
- Consumes: nothing from this pass beyond the state directory as a plain
  string.
- Produces: the leaf-name constant, the join function, an opener creating
  the directories it needs, the digest derivation, and the `validID`
  check. Tasks 8 and 9 build their records on it.

**Outcomes (A6):**

- **One package-level constant names the leaf**, `auth`, with the reason
  stated: the serve unit's `ReadWritePaths` entry is a single static path
  (`spec:1691-1697`). One function joins it to the state directory, so the
  name has one home rather than a literal per call site.
- The opener creates the leaf and both record subdirectories at open,
  with **directories 0700 and files 0600** (R10), stated in the code with
  the reason, rather than left to the unit's umask.
- **The record name is the hex of a full SHA-256 digest** (R-5b12), over
  the value the caller supplies. One derivation serves both records, which
  is what makes rung 6's re-key a change of one function rather than a
  rename of every record (R34, `spec:694-705`).
- `validID` refuses an id that is empty, `.`, `..`, or carries a separator
  in either direction, on rung 4's precedent
  (`internal/engine/shadow/store.go:66-70`), so a name that is not a
  digest is a loud bug rather than a path traversal.
- **A test proves every write stays confined to the root passed in**, and
  that the attacker-shaped `..`-bearing id is refused. That test is the
  static argument the unit's grant will rest on at 5d, exactly as rung 4's
  is (`spec:1698-1704`).

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The leaf constant and the join function each appear once, and no other
  file in the package builds a path by hand, provable by grep for the
  literal.
- Permissions are asserted by `os.Stat` on a created directory and a
  created file, with `0700` and `0600` written out.
- The digest is asserted against a **checked-in vector**: a fixed input
  and its expected 64-character hex, written out rather than recomputed by
  the test.
- `validID` rows cover empty, `.`, `..`, a forward slash, a backslash, and
  a valid digest, each asserted by literal.
- **The confinement test writes through every entry point the package
  offers and asserts, by walking the filesystem, that no file exists
  outside the root.** It also asserts the `..`-bearing id is refused
  before any file is created.
- A test asserts the normalized address does not appear in any created
  filename, which is the other half of R34's argument.
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Name the record from the value rather than its digest → the
   no-address-in-filenames row fails, and the confinement test fails for
   an address whose local part carries `..`.
2. Truncate the digest → the vector row fails.
3. Create directories 0755 and files 0644, inheriting rung 4's shape →
   both permission rows fail, and the reason is R10's: the shadow store's
   permissions are the wrong inheritance for identity state.
4. Accept an id containing a separator → the `validID` rows fail and the
   confinement test fails.
5. Build a path by joining in a caller rather than through the one
   function → nothing fails today, which the task reports as an **unfired
   mutation with its reason**: the criterion is the grep, and the cost is
   paid at 5d when the unit's grant needs one static path.

**2.0 hook: lands hook 5's re-key affordance.** R34's digest derivation is
what makes rung 6's substitution of a surrogate id a change of one
function rather than a rename of every record (`spec:700-703`,
`S20:367-369`).

**Prior art:** **adopts the shape of `internal/engine/shadow/store.go`**
wholesale (in-repo, rung 4's own, gate-proven): one constant per leaf, one
join function, creation at open, and `validID`. It is adoption of a
pattern rather than a dependency, and the permissions are the one clause
R10 supersedes. No third-party library consulted, because the surface is
`os` and `crypto/sha256`.

**Independent: yes**, by files. It writes `store.go` only and consumes
nothing from Tasks 1, 5, or 6.

**Deliverable count: 3** (the leaf and opener, the naming rule, the
confinement test). Gate green; commit.

---

## Task 8: The consumed-nonce store and its sweep

Model: **`opus`**. Depends on Tasks 0, 5, and 7.

**Files:**
- Create: `internal/auth/nonce.go`
- Create: `internal/auth/nonce_test.go`

**Interfaces:**
- Consumes: Task 7's opener, naming rule, and permissions; Task 5's
  nonce; **Task 4's resolved `DUBPLATE_NONCE_RETENTION`**, which is what
  the record carries and what the sweep prunes on. It does not consume
  Task 5's expiry, which is the token's rather than the record's (review
  M-c).
- Produces: an insert that fails when the record exists, a sweep, and the
  replay sentinel. 5b-ii's consumption handler calls the insert after the
  verifier returns.

**Outcomes (A7):**

- **The insert opens the record with `O_CREATE|O_EXCL`, syncs the file,
  and syncs the containing directory** (R11, `spec:421-444`). An existing
  record is the replay sentinel, distinguishable by `errors.Is` from every
  I/O failure.
- Neither `atomicfile.WriteFile` nor `atomicfile.WriteStream` is used
  here, and the doc comment states why: both rename over the target, which
  silently overwrites the record whose whole purpose is
  first-writer-wins.
- **The record carries consumed-at plus the resolved
  `DUBPLATE_NONCE_RETENTION`**, never the token's own expiry
  (**conductor-owed**, review M-c), and **the sweep prunes on that carried
  value** rather than on the file's mtime or on a global cutoff (R31).
  R31's floor then holds retention at or above the link expiry, so the
  record always outlives every token that could bear its nonce. Under the
  other reading the record dies exactly when its token dies,
  `DUBPLATE_NONCE_RETENTION` has no runtime consumer at all, and R31
  guards a mechanism the design does not have.
- The sweep runs at open and is offered as a method for a caller to run on
  an interval. **The interval driver is 5b-ii's** (R-5b14).
- **A sweep that meets a malformed record retains it and reports the
  count**, removing it only once its file mtime is older than the
  resolved retention (**conductor-owed**, review M-d). Measured, the
  insert creates the record exclusively and then writes it, so a crash in
  between leaves a zero-length file whose token may still verify;
  deleting it on sight would make that token consumable again for the
  rest of its expiry. A record whose expiry is unknown is the one whose
  single-use property must be assumed rather than discarded. The mtime is
  the one filesystem timestamp this store reads, and the doc comment says
  why.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- The insert's open flags are asserted by behavior: inserting the same
  nonce twice returns the replay sentinel on the second call, with the
  first call's file byte-identical afterwards.
- **A replay across a simulated restart**: insert, close the store, open a
  fresh store over the same root, insert again, assert the sentinel. This
  is `spec:2168`'s first exit-gate assertion and it lands here.
- **A fixed-order row**: a token whose MAC fails is passed through the
  full consumption path this task can reach, and the nonce directory is
  asserted empty afterwards by counting entries. This is `spec:2168`'s
  second exit-gate assertion, and it is held here rather than in Task 5
  because only this task can count the directory.
- The directory `Sync` call is present and its error is not discarded,
  provable by reading the function.
- A record whose carried expiry has passed is removed by the sweep and one
  whose expiry is in the future survives, with both rows in the same test
  and the boundary case at exactly the expiry stated as retained.
- **A sweep row proving R31's floor matters**: with a retention below the
  link expiry, a still-valid token's nonce would be swept, and the row
  asserts Task 4's load-time refusal is what prevents that configuration
  from existing. The row cites Task 4 rather than duplicating its check.
- A malformed record is retained and counted, and the sweep completes.
- **A crash-truncated record row**: write a zero-length record by hand,
  run the sweep, and assert a second insert of that nonce still returns
  the replay sentinel. Then advance the record's mtime past the retention
  and assert the sweep removes it.
- A record survives past its token's expiry when retention exceeds it,
  which is the row that proves the record carries retention rather than
  the token's expiry.
- No nonce value appears in any error string. Every checked-in nonce or
  digest literal carries an identifier that does not read as a key name
  (Global Constraints).
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Insert through `atomicfile.WriteFile` → the replay row fails, because
   the rename overwrites and the second insert succeeds. This is risk 3's
   exact shape (`spec:2364-2369`).
2. Check for the record's existence and then create it → the row still
   passes single-threaded, so the task adds a **concurrent-insert row**:
   two goroutines inserting the same nonce, exactly one succeeding. That
   row is what fires.
3. Drop the directory `Sync` → nothing fails, which the task reports as an
   **unfired mutation with its reason**: a crash-consistency property has
   no gate-resident test, the criterion is the reviewer's reading, and the
   doc comment carries the argument.
4. Sweep on file mtime → the carried-expiry rows fail, because a record
   written late with a short expiry survives and one written early with a
   long expiry is removed.
5. Delete a malformed record on sight → the crash-truncated-record row
   fails, reporting a token consumable twice after a crash. Aborting the
   sweep on one instead fails the sweep-completes row.

**2.0 hook: neutral.** The nonce record is self-expiring and rung 6
absorbs neither record (R18, `spec:533-540`).

**Prior art:** none adopted. `internal/atomicfile` is examined and
**declined for this record** with the reason recorded, which is R11's own
finding: INV weighed `WriteFile` against `WriteStream`
(`INV:557-566`) and neither is correct for a first-writer-wins record.
The `O_CREATE|O_EXCL` plus two-fsync shape is the stdlib's own answer.

**Independent: no.** It consumes Tasks 5 and 7.

**Deliverable count: 3** (the insert, the sweep, the tests). Gate green;
commit.

---

## Task 9: The revocation epoch and its seam

Model: `sonnet`. Depends on Tasks 0, 1, 5, 6, and 7.

The pass's headline 2.0 hook, and the record that makes offboarding real.

**Files:**
- Create: `internal/auth/epoch.go`
- Create: `internal/auth/epoch_test.go`

**Interfaces:**
- Consumes: Task 7's opener and naming rule, Task 1's identity **and its
  address normalizer**, Task 5's mint time, **Task 6's anchor** rather
  than its issue time.
- Produces: a bump, a read returning the current epoch for a resolved
  identity, a change notification a subscriber registers for, and the
  two refusal checks a caller applies to a session and to a token. 5b-ii's
  session middleware and consumption handler call both checks; 2.0's
  credential store subscribes to the seam.

**Outcomes (A8):**

- One record per normalized login address, named through Task 7's digest
  (R4, R34), written through `internal/atomicfile.WriteStream` because
  overwriting is what a revocation does and because only `WriteStream`
  fsyncs the containing directory after the rename
  (`atomicfile.go:82-89`, R11).
- **The seam is a read plus a change notification** (R35,
  `spec:706-715`). The read returns the current epoch for a resolved
  identity and is the authority, because that is what a credential store
  stamps on a password it mints and what a later check compares against
  (`S20:245-246`). The notification is the convenience half.
- **A session whose anchor predates its address's epoch is refused**
  (**conductor-owed**, review B1), and **a token whose mint time predates
  it is refused too**. The anchor rather than the issue time, because
  rolling refresh moves the issue time (Task 6): a session refreshed once
  after a bump would carry a post-bump issue time and be permitted from
  then on, permanently, which is R30's own failure mode (`spec:643-645`)
  reproduced on every request rather than once at mint. The anchor is
  what refresh never moves, so no ordering rule between the epoch check
  and the refresh is load-bearing. Without the token check, an
  outstanding link survives an offboarding for the full link expiry and
  mints a session the epoch check then honors (`spec:637-648`).
- **The address is normalized through Task 1's function** before it is
  digested, so the key the epoch is stored under and the key a session
  carries cannot diverge.
- An address with no epoch record has no revocation, which is the
  ordinary case, and the read reports it as such rather than as an error.
- The address key is documented **in the code** as the one per-user record
  deliberately keyed on an address, with rung 6's re-key recorded as rung
  6's obligation (R4, `spec:333-339`).
- A subscriber that is not running when a bump happens still learns the
  truth on its next read, which is the property a publish-only event
  cannot offer.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- A bump followed by a read returns the bumped value, across a fresh store
  opened over the same root, so the record's durability is proved rather
  than its in-memory copy.
- **A session minted before a bump is refused and one minted after is
  accepted**, with both rows and the boundary case at exactly the epoch
  stated as refused, both read through the anchor. This is `spec:2168`'s
  third exit-gate assertion.
- **A session refreshed after a bump is still refused**, with the refresh
  performed between the bump and the check, so the row fails outright
  under an issue-time check. This is the row that holds threat-model
  claim 4.
- **A link minted before a bump is refused**, asserted through the token's
  mint time rather than through its expiry, and the row states that the
  attacker chooses when to click.
- The epoch check normalizes the address before keying, proved by a row
  where the session's address differs in case from the bumped address and
  is still refused. A non-normalized key silently fails to revoke
  (`spec:2370-2376`, risk 4).
- The notification reaches a registered subscriber, and **a test proves
  the read is authoritative without it**: a bump with no subscriber
  registered, followed by a fresh store's read, returns the bumped value.
- `WriteStream` is the call, provable by reading the function, with the
  directory-fsync reason in the doc comment.
- The confinement test from Task 7 covers this record's writes.
- No address appears in a filename, re-asserted here for this record.
  Every checked-in digest literal carries an identifier that does not
  read as a key name (Global Constraints).
- `go.mod` unchanged. `bash scripts/check.sh` green in the worktree.

**Named mutations this task reports:**

1. Write through `atomicfile.WriteFile` → nothing functional fails, which
   the task reports as an **unfired mutation with its reason**: the
   difference is the directory fsync, the failure needs a power cut, and
   R11's argument is that the weaker call fails open on the one file whose
   whole purpose is to close access. The doc comment carries it.
2. Check only the session's issue time and not the token's mint time →
   the link-before-the-bump row fails, which is R30's whole finding.
3. Key the epoch on the raw address → the case-difference row fails.
4. Check the epoch against the session's issue time rather than its
   anchor → the refreshed-after-the-bump row fails, reporting a revoked
   session permitted for the rest of its cap.
5. Make the comparison exclusive at the epoch → the boundary row fails.
6. Offer publication without a read → the no-subscriber row fails, and
   the task reports that S20's credential store is what the row stands
   for.

**2.0 hook: lands hook 5.** `S20:245-246` requires every app password to
carry the user's valid-after epoch so a roster-level revoke invalidates
every device at once, and R35's read is what a credential store stamps
from. The re-key onto rung 6's surrogate id changes the key of both halves
together (R4), and Task 7's digest is what makes that a change of one
function.

**Prior art:** none adopted. **Adopts `internal/atomicfile.WriteStream`**
(in-repo, rung 4's, gate-proven, and the only entry point that fsyncs the
directory). **Records Gitea's per-session revocation as the alternative
architecture** (PRA:568-578) with the reason rung 5 declines it: a
server-side store is what you build when sessions must be listable, and
the epoch buys per-person revocation at a fraction of the machinery.

**Independent: no.** It consumes Tasks 5, 6, and 7.

**Deliverable count: 3** (the record, the seam, the two refusal checks).
Gate green; commit.

---

## Task 10: The pass close, and the 5b-i to 5b-ii handoff

Model: `sonnet`. Depends on Tasks 0 through 9.

**Files:**
- Modify: `docs/STATUS.md`, `docs/HISTORY.md` (drafted for the
  conductor's batch commit; **the implementer never commits a ledger**)
- Draft, handed to the conductor in the task return: the spec
  corrections, the coverage-gap list, and the 5b-ii handoff

**Outcomes:**

- The HISTORY entry in the house shape: what landed, what the gate caught,
  and **what a later pass would be wrong to rediscover**.
- The auth core's contract coverage, A1 through A8, each with its Go site,
  what proves it, and, for a clause no test can reach, the stated reason.
- **The sixteen resolutions listed with their markers**, each row naming
  the file and symbol the marker sits on and stating what execution
  confirmed or contradicted. The markers are `// ruled, conductor
  2026-09-08, R-5bN`, stamped by the task that landed each, so this table
  is a grep.
- The declared coverage gaps recorded as gaps, on one criterion: a surface
  this pass itself builds that no test in this pass can reach. **The
  expected list, to be corrected against what execution found:**
  1. **A constant-time comparison has no gate-resident proof** (Task 5,
     mutation 3). The criterion is the reviewer's reading of the call.
  2. **Crash consistency rests on two fsync calls rather than on a test**
     (Tasks 8 and 9). A power cut is not a thing this gate reproduces.
  3. **No handler, no limiter, and no origin check exercises any of this**
     (R-5b1). 5b-ii is the first real caller of the token, the session,
     and both stores.
  4. **The per-source declaration family has exactly one declared source**
     (Task 2). Its two-source behavior is proved only against a fixture
     source, and the second real source arrives with a forward-auth front
     nobody has built.
  5. **Nothing validates a registered threshold's value except the two
     this pass's hook reaches** (Task 4, and security review finding 24).
     Task 2's hook takes the environment snapshot, so
     `DUBPLATE_NONCE_RETENTION` and `DUBPLATE_MAGIC_LINK_EXPIRY` are
     validated at load and their relation with them (R31). Finding 24's
     residual stands for every other registered threshold, which `Load`
     still accepts with any value, and that is a property of the rung 1b
     wall contract rather than a gap this pass created.
  6. **The sweep's interval driver does not exist** (R-5b14), so the sweep
     is proved by direct call rather than by elapsed time.
  **The count is stated and joined to the running total**, which stands at
  sixty-three after 5a (`docs/STATUS.md`'s carried list), so rung 15 does
  not re-count.
- **The spec corrections the pass owes**, drafted for the conductor:
  - **R23's ceiling table and the pass split** (R-5b1): 5b becomes 5b-i
    and 5b-ii, with both ceilings and the rung total against R23's
    reservation.
  - **R24's review shape** (R-5b16): a fresh-context
    `web-auth-security-reviewer` read at each half's pass end.
  - **R29's wording** (R-5b4): `Load` requires every owned-now wall row
    rather than naming `DUBPLATE_IDENTITY_SOURCE` by hand, which closes
    risk 11 for every rung rather than for one key.
  - **The env template's row assignment** (R-5b7): one literal row lands
    at 5b-i rather than at 5d, because a gate-resident test demands it.
    `spec:2205`'s "commented pending rows only" and `spec:2170`'s row
    assignment both take the correction.
  - **Two of `spec:2168`'s exit-gate assertions belong elsewhere**: the
    roster-removal half of the third is 5b-ii's, and the verify-subrequest
    half of the eighth is 5d's, since `spec:2170` gives the endpoint to
    5d.
  - **The staleness pre-task's `serve` reserved-class bullet**
    (`spec:2297-2300`) is stale, because 5a removed that class.
  - **The `keys.go` pointers at `spec:1506`** (`:177-178` and `:171-176`)
    moved to `:165-166` and `:159-162`.
  - **The `atomicfile` pointer at `spec:434`** (`:45-56`) resolves to
    `:55-91`, with the directory fsync at `:82-89`.
  - **The epoch's session check reads the anchor** (review B1).
    `spec:1085-1086` says "a session whose issue time predates the epoch
    for its address is refused", and rolling refresh moves the issue time
    (`spec:1080-1081`), so as written an epoch bump is undone by the next
    refresh. The anchor is the session's original issue time and the
    field refresh never moves.
  - **The token's field list gains the mint time.** `spec:1016-1017`
    enumerates three fields where R30 (`spec:640-642`) requires a fourth,
    and the landed token carries four.
  - **The serialization claim covers the wire form.** `spec:1017-1019`
    states it and the landed token states the encoding, so a field value
    can never carry the encoding's separator.
  - **The nonce record's carried value.** `spec:1664-1668` says the
    record carries "its expiry" without saying whose; the record carries
    consumed-at plus the resolved retention, which is the only reading
    under which R31's floor has a mechanism to guard.
  - **R38's carve-out covers the loopback set** (`spec:758-767`), not the
    name `localhost` alone, because the browser rule R38 reasons from is
    the W3C potentially-trustworthy-origin definition.
  - **The non-secret name count.** `spec:1640-1645` says "eight" and
    enumerates seven; `spec:2170` gives 5d "its eight literal rows". One
    lands at 5b-i, so six remain for 5d.
  - **R33 and R24 belong to 5b.** `HISTORY:393-394` assigns R33 to this
    pass and R-5b16 corrects R24's one-read-inside-5b wording, so the
    ruling routing in the spec's own reading of who reads what is
    restated.
  - **R1's own `keys.go` pointer at `spec:284`** (`:176-177`) moved to
    `:165-166`, the same move as `spec:1507`'s.
  - **The ceiling erratum's own pointer.** This plan cited
    `spec:2078-2087` for it in revision 1; it is `spec:2042-2050`, and
    the rung reservation's live baseline is the 15.75M that erratum set.
  - **Any correction Task 0 found** in this plan's own pointers.
- **The 5b-ii handoff**, naming what 5b-ii inherits rather than rebuilds:
  the identity leaf, its assertion channel, and **its address normalizer,
  which every keying, every logging, and every rate-limit lookup runs
  through** (`spec:1161-1163`); the token's verifier and
  the nonce it returns, to be inserted after verification and never
  before; **the token's version label as the only place a later field,
  a destination included, may be added** (R33, `spec:690-692`); the
  session's minter, verifier, and refresher, with the rule that refresh
  never moves the anchor, that the anchor is the session's original
  issue time, and that **refresh counts same-origin requests only**
  (`spec:1081-1082`); the epoch's two refusal checks, **the session check
  reading the anchor and never the issue time**, applied to a session on
  every request and to a token at consumption, **plus the roster re-check
  that refuses a link for an address removed from the roster** (R30,
  `spec:2168`);
  the seam's read as the authority a credential store stamps from; the
  state-directory opener and its naming rule, which the roster's own leaf
  sits beside rather than inside; `registerAuthConfig` and the declared
  key set, which 5b-ii's remaining thresholds join rather than re-derive;
  the constraint that `internal/auth` never reads the environment; 5a's
  own carries, unchanged, namely that `ClientIP` is never re-derived, that
  the origin check inserts after the strip and after normalization, that
  `Origin` and `Sec-Fetch-Site` survive ingress while
  `CrossOriginProtection.Check` returns nil when both are absent, and that
  every URL builds from the configured public origin and never `r.Host`;
  the merge carry on `internal/serve/server.go:161`, forwarded to 5d;
  **the 30-second read timeout and the `http.MaxBytesReader` alternative**
  (R-V, `HISTORY:391-392`), which is what bounds a request body until
  `spec:1590`'s cap lands; and
  the two module dependencies R26 budgets, both of which 5b-ii is the
  first pass to take.
- **The evidence, not a bare "green"**: the mutation table across Tasks 1
  through 9 with every outcome, the deliberately unfired ones included and
  each with its reason; the measured wall times for `internal/auth`,
  `internal/identity`, `go test ./...`, and `bash scripts/check.sh`; the
  count of existing tests this pass changed with the reason for each; and
  confirmation that `go.mod` and `.gitleaksignore` are untouched.
- **Both budgets scored**: tokens against the 5.0M ceiling, and attended
  time as two counts. **The target is zero planning misses and zero
  execution sittings.**

**Acceptance criteria:**

- Every gap in the list is a gap this pass created or inherited, with the
  task and the mutation that names it.
- The mutation table has a row per named mutation across every task, and
  **every unfired mutation carries its reason**.
- `docs/STATUS.md` stays present tense and under 60 lines after the edit,
  with the pass ledger **moved** to HISTORY rather than summarized.
  STATUS sits at exactly 60 lines today, so the edit trades lines rather
  than adding them.
- No ledger is committed by the implementer, and the lane registry entry
  is the conductor's.

**2.0 hook: neutral.** The close records hook 5 as landed by Tasks 1, 7,
and 9, and hook 6 as untouched by this pass.

**Prior art:** none consulted. The close reads the pass.

**Independent: no.**

**Deliverable count: 4** (the HISTORY entry, the STATUS edit, the drafted
corrections, the 5b-ii handoff).

---

## The conductor's actions at the pass end

Stated here because they are not tasks and no implementer performs them.

1. **`code-simplifier` over each accepted task's changed Go**, dispatched
   after that task's review accepts and before the close. Subagents cannot
   dispatch agents, so this is the conductor's dispatch every time
   (`spec:2455-2457`). The per-task cadence is CLAUDE.md's
   git-conventions rule rather than that sentence, which places the
   dispatch at the pass close.
2. **One fresh-context `web-auth-security-reviewer` over the whole auth
   core** (R-5b16), dispatched after Task 10's gate and before the merge
   ritual. **Its input is the threat-model section of this plan, the
   spec's own threat model (`spec:1909-1980`), and the diff.** It does not
   receive this plan's task decomposition. Its findings are folded or
   refused with the reason recorded, and **that record is the
   do-not-re-litigate list for 5b-ii, 5c, and 5d** (`spec:2472-2478`).
3. **The merge ritual, which closes the lane.** Merge token, rebase onto
   master, full gate in the worktree, `--ff-only` merge, full gate on
   master, and **one `diff-reviewer` over the merged result whenever the
   rebase was not a fast-forward**. `cmd/dubplate-server` is the one
   surface where lane 1's 8c could make it non-fast-forward, and Task 0
   item 3 reports on it before the pass starts.
4. **Release the second code slot.** The worktree is removed once master
   is green, because the slot passes at a pass close rather than a rung
   close (`spec:2273-2275`).
5. **Rule the conductor-owed decisions** this plan names: the split
   (R-5b1), the 5.0M ceiling and its 4.0M flag, **the 5.5M overage
   against the live 15.75M reservation**, the second security read
   (R-5b16), and the nine folds revision 2 marks conductor-owed inline
   (the epoch's anchor check, the hook's environment snapshot, the
   token's wire encoding, the nonce record's carried value, the
   malformed-record retention rule, the R33/R-V/R24 routing, the
   reservation figure, R38's loopback widening, and the lane registry's
   home in `docs/STATUS.md`).
6. **Re-derive 5b-ii's ceiling from 5b-i's actual report** before 5b-ii's
   plan is drafted, and record the rung total against R23's reservation.

---

## Execution model

Workflow mode per the standing grant: the pass-execute chain
(`~/.claude/workflows/pass-execute.js`) per task, **strictly sequential**,
with `repo` pointed at the lane-2 worktree. Nothing here runs in parallel,
and **no second executor enters the worktree**. Every task runs
`dubplate-implementer`, then `diff-reviewer` (`claude-opus-5`), then the
gate, with one re-dispatch on `fix` and a second `fix` as the conductor's
decision. **The runaway guard is armed at launch. Subagents never dispatch
agents.**

**Models.** `sonnet` by default. **Three named upshifts to `opus`:**

- **Task 2, the config core.** The mechanism is under-specified by design:
  R29 names an outcome and the tree offers three shapes that reach it,
  and the wrong one either puts rung 5's vocabulary in the config core or
  leaves risk 11 open for the next rung. Its blast radius crosses every
  caller of `Load`.
- **Task 5, the token.** Length-prefixed serialization, two labels inside
  the MAC input, and a fixed verification order are three properties whose
  wrong reading produces code that passes every ordinary test. This is
  "novel correctness-critical logic the plan cannot fully specify" in
  CLAUDE.md's own words.
- **Task 6, the session.** The anchor's invariance under refresh is the
  single arithmetic property that silently removes the absolute cap when
  it is wrong, and `spec:2370-2376` files it as risk 4.
- **Task 8, the nonce store.** Also `opus`. The create-exclusive insert
  plus two fsyncs is the mechanism risk 3 exists for, and the failing
  variant passes any test that checks only that the record exists.

**Task 0 also runs at `opus`**, fresh context, as the brief requires: a
staleness check that misses a moved citation costs more than the model
difference, and plan staleness is the build's named dominant defect class.

**One named runner-up**, upshifted on a condition rather than by policy.
**Task 3, the declarations**, if Task 2's review returns `fix`. Task 3's
correctness rests on the family Task 2 landed, and its own blast radius
crosses two templates and four test files. It stays `sonnet` unless that
evidence weakens.

**Review.** Rung 5 keeps the **full fresh-context adversarial cycle**
(CLAUDE.md's build-process section, `spec:2459-2478`), and 5b-i's share is
the in-chain `diff-reviewer` per task plus **one fresh-context
`web-auth-security-reviewer` at the pass end** (R-5b16), dispatched by the
conductor with the threat model as its rubric. The two-lens cycle over the
whole subsystem is 5d's.

**Dependencies and order, stated exactly:**

- **Task 0 runs first, in fresh context, before any implementer
  dispatch.** Its blockers are ruled by the conductor before Task 1.
- Task 2 → Task 3 → Task 4. The declarations need the family, and the
  hook needs the registration.
- Task 4 → Task 5 → Task 6. The token needs its expiry and its floor; the
  session follows the token's label convention.
- Task 7 → Task 8. The nonce record needs the leaf and the naming rule.
- Tasks 5, 6, 7 → Task 9. The epoch compares against both a mint time and
  an issue time and writes through the leaf.
- Tasks 0 through 9 → Task 10.
- **Tasks 1, 2, and 7 are mutually independent by files**, writing
  `internal/identity`, `internal/config`'s four files with their tests
  plus `cmd/dubplate-server/doctor.go` and `doctor_test.go`, and
  `internal/auth/store.go`, and consuming nothing from each other. **They
  still run sequentially**, because one tree runs one executor and the
  module gate runs over `./...` regardless. The marks exist so a conductor
  can reorder them when a dependency stalls, never so a second executor
  enters the worktree.

**Where this pass splits, if it splits again.** Task 0 plus ten tasks is
under the fifteen-task trigger, so **the recommendation is to keep 5b-i as
one pass.** The cut point, if the Task 3 re-derivation puts the ceiling
past 5.5M, is the **Task 6 / Task 7 boundary**, where the work changes
from cryptographic construction in memory to durable records on disk.
Tasks 0 through 6 would be 5b-i and Tasks 7 through 10 would be 5b-i-b.
**Two reasons argue against taking it.** The first half alone ships a
token and a session with nothing to make either single-use or revocable,
which is a surface no reviewer can grade for the property that matters.
And the pass-end security read would land on a construction with no
durability, which defeats the reason the token and the state records were
put in the same half in the first place.

**Budget:** 5.0M ceiling, re-derived here from 5a's actual and ruled by
the conductor 2026-09-08, against R23's 4M for undivided 5b. Rung 5's
combined reservation stands at 21.25M against the live 15.75M
(`spec:2042-2050`), the **5.5M** difference recorded rather than taken
from another pass. **Flag at 80%, which is
4.0M**, routed to the conductor; re-derive after Task 3; STATUS after Task
3, after Task 7, and at the close. Spend is scored at close against the
ceiling and against **zero attended interactions**, which is this pass's
target. The sixteen resolutions, and the nine conductor-owed folds
revision 2 carries, are ruled at dispatch rather than at the
close, on 5a's precedent; **none counts as an execution sitting**, and
each carries its marker in the code so a later ruling is a grep.
