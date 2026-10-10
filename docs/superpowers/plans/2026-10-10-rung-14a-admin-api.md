# Rung 14a (the Admin API's First Half) Implementation Plan

Status: **revision 1, drafted 2026-10-10 from the approved spec; awaiting `spec-plan-review`.**

**Spec:** `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md`, revision 2.1 at `c7c3fbb` ("the spec").
Its fold record is `docs/superpowers/research/2026-10-10-rung-14a-spec-fold.md` ("the fold"); this plan places the
fold's owed items V4 to V7, doctor's warning split, errata 1 to 8, and the conductor-owed list. Task sketch, ceiling,
and clock come from `docs/superpowers/research/2026-10-10-next-pass-planning-brief.md` section 6 ("the brief").
The plan cites the spec and never restates it. Spec abbreviations (BS, RP, SS, CH, R6, PA) carry over; PS is the
group-scale principles spec (`docs/superpowers/specs/2026-09-25-group-scale-principles-design.md`). Code pointers are
at `c7c3fbb`; the pre-flight before each segment re-resolves them.

**Goal.** Land the `/api/v1` cookie mount with its contract document, the admin settings write path behind the
re-authentication gate, member removal and demotion over the CLI and the API, hook 18, and the two pages that call
them.

**Architecture.** `internal/ops` gains a second interface, the admin seam, beside the job seam. A new
`internal/ops/admin` package implements it for `serve` over the roster (`internal/auth`) and a new settings write
function (`internal/settings`). `opshttp` gains its first exported surface, the mount, and owns every answer under
`<prefix>/api`. The CLI verbs call the roster and the settings write function directly, which are the shared write
path (spec, "The mount").

**Pass class:** `auth-data`. Overrides: Task 0 and Task 2 are `tool`; Task 9 is `tool` (a read-only `doctor` row, an
`Inspect` fact, and runbook text, with no write path); Task 10 is `engine-logic` with the `web/` rendered leg.
**Ceiling:** 10.5M execution tokens, flag 8.4M (80%). Planning and plan review sit outside it.
**Clock estimate:** about 6 hours of execution, excluding owner pauses; `auth-data` fix rounds push it toward 7.5.
**Checkpoint interval:** four tasks at most: each segment boundary and the close.
**Segments:** Task 0 (conductor, on master). A: Tasks 1, 2, 3. B: Tasks 4, 5, 6, 7. C: Tasks 8, 9, 10. Then the close,
Task 11. Boundaries A and B end on a green full gate on the heavy lane; C's yields to the close gate (pass-core,
"Closing" step 2).

### Ceiling and clock, re-derived

The brief's 9.5M priced ten tasks (Task 0 and nine build tasks) at 0.6M plus 3.0M of overhead. Revision 2.1 adds
work the brief did not price: the admin seam and the unavailable job seam (V1), the `error.detail` field and two wire
codes, the keyword walk, the re-add bump, the roster API's retirement into `list_members` and `add_member`, the
runbook and `doctor` edits, and the member cap this plan takes (below). It drops the older-store write mode and every
migration. The re-cut is ten build tasks plus Task 0: 11 x 0.6M = 6.6M, an Opus premium on Task 6 of 0.3M, and an
overhead of 3.2M (the brief's 3.0M plus two more architecture reads, for `internal/ops/admin` and
`internal/engine/store`). The sum is 10.1M; 10.5M leaves most of one more fix round.

The clock follows the brief's rates: eleven tasks at about 22 minutes (2b's 20, plus the rendered leg every task here
runs), two boundary gates at about 9 minutes, and a close of about 1.75 hours (thirteen reviewer reads, the
simplifier, the live smoke, the render-and-read, and the merge ritual's two full gates).

## Tree, branch, and execution

**Tree.** Lane 1 runs in its own worktree, `.claude/worktrees/lane-1-rung-14a`, on branch `lane-1-rung-14a`, cut from
master after Task 0 lands. This departs from CLAUDE.md's "the primary code lane ... share[s] the tree", for three
reasons. pass-core requires the pass's branch or worktree, "never the default branch's checkout". The merge ritual
(rebase, `--ff-only`) presumes a branch. And the main tree stays on master for the rung 7 planning lane and the
conductor's ledger commits. The CLAUDE.md rule exists so two code lanes never share one gate's `./...`, and this
layout keeps that. The pass needs no venv, no corpus, and no beets, so nothing pins it to the main tree. dubplate has
no CI (`.github/workflows` is absent), so pass-core's draft-PR shadow does not apply.

**Executor:** `dubplate-implementer`, `model: sonnet`, effort `high`, except Task 6 on `model: opus` from the start
(the gate's decision point, the notice's ordering, and seven operations over two write paths). A second `fix` on
Task 4 or Task 7 is the upshift the conductor weighs first.

**Per-task gate:** `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-1-rung-14a/scripts/check.sh --scope-changed'`,
re-issued on exit 75 until it prints `gate exit:`, on the heavy lane for every task. Off master, `check.sh` diffs the
change set from `git merge-base master HEAD` (`scripts/check.sh:177-193`), so the set is cumulative over the pass.
Task 1 edits `internal/identity/identity.go`, a rendered-leg input (`:246`), and from then on every task's scoped gate
runs the rendered leg and launches Chromium. `pass-execute`'s `tool` class defaults to the light lane
(`PASS_CLASSES.tool.gateLane`), so Tasks 2 and 9 carry `gateLane: "heavy"` (HISTORY 6d: "Pass `gateLane: "heavy"` for
a gate that launches Chromium"). `reducedGate` is the `--reduced` form; under `auth-data` only a comment-only round
takes it. Boundaries take `check.sh` (full, heavy); the merge ritual takes `check.sh --merge` in the worktree, then
`check.sh` on master.

**Mode:** `pass-execute` by name, one invocation per segment, sequential, with `repo` set to the worktree path,
`implementer: "dubplate-implementer"`, `reviewer: "diff-reviewer"`, `reviewerModel: "claude-opus-5-5"`, `maxFix: 1`,
`stopOnEscalate: true`, `passClass: "auth-data"`, and each task's own `passClass`, `gateLane`, and `model` as above.
No task is marked independent: every task from 3 on reads the one before.

### The unattended run

The conductor runs Task 0, segments A to C, and the close without an owner touchpoint, holding the sleep inhibitor,
the stall guard, and the fallback `/loop` wake-up per `~/.claude/docs/unattended-work-guards.md`.

**The hard ceiling.** At the flag (8.4M) the in-flight chain finishes, STATUS is written with the resume prompt, the
current segment finishes, and no new segment starts. At 10.5M the conductor stops after the in-flight chain returns
and writes STATUS. STATUS is also written at each boundary and before any stop.

**The conductor rules alone, recording each in its notes and later in HISTORY:** accept; one re-dispatch on `fix`; an
upshift to Opus; splitting a task whose second `fix` names a seam; a flake the rerun rule covers; a pre-flight claim
that is wrong at HEAD, amended in the plan before dispatch; any decision a task marks "the implementer's"; a
reviewer finding refused under CLAUDE.md's fold rule, with its one-line reason; and a dependency held by Task 0's
rules.

**These stop the run (STATUS written, then stop):** the hard ceiling; a `fix` still standing after its one fix round
on a real defect; an `escalate` the conductor cannot resolve inside the task's files; a warm change in the worktree
the conductor did not dispatch, or another executor writing a surface the contested-surfaces table gives to 14a; a
gate red on a commit whose diff does not explain it, after one re-run; a security finding at High or above that
needs a design change; and an architectural question the spec does not settle (pass-core).

**Owner-gated steps, batched at the end:** none needs Geoff's go. The live auth smoke and the render-and-read of both
pages are Claude's verification (pass-core, "Verification a plan parks for the owner is Claude's"), run at the close
before Geoff sees anything. Their evidence and the pages' captures go to Geoff in the close's one summary. Nothing
deploys, and the box is never contacted.

## Global constraints

- **Territory.** The spec's contested-surfaces table ("Contested surfaces with the rung 7 lane") is this lane's
  ownership record, by reference. 14a writes nothing in `internal/serve`, edits no `RequireAdmin` line, and leaves
  `post.ts`, `SiteHeader.svelte`'s structure, and the rendered harness's shape to rung 7, appending rows only.
- **No migration**, no new module dependency outside Task 0's sweep, and kin-openapi in `_test.go` files of `opshttp`
  only (spec, "The contract document"; the no-link test stays green).
- **Layering.** `internal/ops` and `opshttp` import no `internal/auth` or `internal/settings`; `internal/auth` and
  `internal/settings` import no `internal/ops`. Only `internal/ops/admin` and `cmd/dubplate-server` join them (fold,
  V1's verification).
- **No export ahead of a caller** (CLAUDE.md, 2026-09-12). Every new export names its production caller in the task
  that adds it.
- **Config keys** are declared from 14a's own file, `cmd/dubplate-server/api.go`, never by editing
  `internal/config/keys.go`'s rows beyond hook 18's descriptions (CLAUDE.md, "Facts agents get wrong").
- **Go** under `go-conventions`; **web** under `svelte-conventions`, `ts-conventions`, the `daisyui` skill, and the
  visual identity spec (`docs/superpowers/specs/2026-09-24-visual-identity-design.md`); copy in plain terms.
- **Every proof names its fixture state**, and every `auth-data` task names its mutations; the implementer reports
  each failing.

## Decisions this plan takes

The spec leaves these to the plan. Each is architecture or method, decided on its evidence.

- **P1. The contract document's path** is `internal/ops/opshttp/openapi.yaml`, beside the tests that load it. Nothing
  embeds or serves it.
- **P2. The mount and the context carrier.** `opshttp` exports one constructor for the mount, its first caller
  `cmd/dubplate-server/api.go`. The mount takes the job seam, the admin seam, the prefix, a logger, and a verifier
  function `(http.ResponseWriter, *http.Request) -> (ops.Actor, ops.Privilege, error)`. It cuts `<prefix>/api/v1/`
  itself, answers every other `api` path with JSON `not_found`, then calls the verifier: an error wrapping the new
  `unauthenticated` sentinel answers 401, and any other error answers 500 `internal`. An admitted caller rides the
  request context under an unexported key type (`go doc context.WithValue`: "The provided key must be comparable and
  should not be of type string or any other built-in type ... Users of WithValue should define their own types for
  keys"), and `opshttp`'s authorizer is its only reader. The verifier is built in `api.go` from a new session check in
  `internal/auth`, which verifies through the seam, refreshes on a same-origin request as `RequireAdmin` does
  (`admin.go:110-133`), and reports a refused session distinctly from a server failure.
- **P3. The 401 and 500 split.** A verification refusal whose cause is in `gateRefusals` (`admin.go:142-150`: no
  cookie, a malformed or forged session, the cap, expiry, revocation, an unknown member) answers 401
  `unauthenticated`. A failed setting read or roster lookup is the server failing and answers 500 `internal`, logged
  at Error, as `gateRefusalReason` already classes it (`:155-165`). The spec's states table names only the first.
- **P4. The admin settings route** is `server-settings`, a single segment (`Shell.Register` refuses a separator,
  `serve.go:120-121`), behind `RequireAdmin`, with a `no-store` document row beside `members` (`web.go:104-108`).
  The page title is "Server settings". The visual identity spec reserves `settings` (`:145-146`).
- **P5. How the admin returns after a fresh sign-in.** The link interstitial's `signed_in` state offers no onward
  link, by design (`web/src/routes/login/link/+page.svelte:15-17`), and that page is rung 7's territory. The pages
  therefore add no return parameter, which would also add a redirect target to validate. On
  `reauthentication_required` a page keeps the admin's pending change, links to the sign-in page, and offers a retry.
  The cookie a link consumed in another tab of the same browser serves the retry. The copy says to sign in from this
  browser.
- **P6. The re-authentication window** is `DUBPLATE_REAUTH_WINDOW`: a duration setting, store-homed, sensitive, apply
  mode next-operation. Its default is `2h`, GitHub sudo mode's window (spec D10's source). Its floor is `5m`: the
  anchor is set when the link is consumed (`session.go:80-86`), so the round trip that matters runs from that click
  to the retry, and the floor keeps a mis-set window from locking the web out of resetting it. Its ceiling is
  `24h`, and a correlated rule refuses a window at or above `DUBPLATE_SESSION_ABSOLUTE_CAP`, on the precedent of
  `sessionLifetimeWithinCap` (`cmd/dubplate-server/auth.go:325`).
- **P7. The member cap** (spec deviation, below) is `DUBPLATE_MEMBER_CAP`: an integer setting, store-homed, sensitive,
  default `20` (PS ruling 4 at `:729`, ruled as recommended per PS's status line at `:3`). Its floor is `1` and its ceiling `10000`, because `AddSetting`
  panics on "a kind with no bound" (`internal/config/registry.go:410-414`) and PS lets the operator raise it to any
  practical number.
- **P8. Wire names.** A member is `user_id`, 32 lowercase hex characters. The `error.detail` enumeration is every
  roster-write outcome the members page keys on (`web/src/routes/members/+page.svelte:46-54`), plus `last_admin`,
  `environment_sets`, `busy` (V5), and `member_cap`. `admin_not_grantable` and `unknown_field` leave it, since a
  decoder that ignores unknown keys gives neither a cause (V5; `opshttp.go:141-144`).
- **P9. doctor's `DUBPLATE_ADMIN_EMAIL` warning has two messages, not three** (the fold's over-ceremony item 2). A
  removed member and an address no member holds name one action, `member add <address> --admin` or pointing the
  variable at a live admin, since `addUser` reactivates a tombstoned row with the flag (`roster.go:395-419`). A live
  non-admin names another: no promote verb exists and `member add` refuses a live address, so the only action is to
  point the variable at a live admin. All three states still warn (AC21).
- **P10. The roster's in-write admin check is one exported roster function over a `*store.Tx`.** The settings write
  takes it as a guard parameter from `internal/ops/admin`, so the users table's SQL stays in `internal/auth` and
  `internal/settings` imports no `internal/auth`.

### Spec deviation: the member cap lands here

PS routes "Member cap on every add channel" to "Rung 14 settings; the roster's add path" (`:580`), with acceptance
row A1 (`:619`) and the default ruled at 20 (ruling 4, `:729`; ruled at `:3`). The spec omits it, and 14a rebuilds the web add
channel as `add_member`. Shipping that channel without the cap would leave the ratified mechanism unbuilt behind a
new door. 14b builds collections and no add channel, so 14a is the cap's rung. The cap rides Tasks 4, 9, and 10 and
adds no task. If the conductor routes it to 14b instead, those three tasks drop their cap clauses and nothing else
moves. `member import` retired at 6d and `dubplate import navidrome` does not exist, so A1's channels at 14a are
`member add` and `add_member`.

## Behaviors this plan relies on, quoted

The spec's quotes stand (spec, "Planning-miss prevention", first bullet). Added here, read on 2026-10-10:

- `go doc encoding/hex.DecodeString`: "DecodeString expects that src contains only hexadecimal characters and that
  src has even length." It accepts uppercase: `hex.DecodeString("ABCD")` returned `[171 205] <nil>` on go1.27.2. The
  `user_id` parser therefore refuses any input that is not its own lowercase re-encoding.
- `go doc strings.CutPrefix`: "If s doesn't start with prefix, CutPrefix returns s, false." The mount's cut branches
  on `found`.
- `go doc context.WithValue`, quoted under P2.
- `go doc ./internal/engine/store Handle.Update`: "Update runs fn inside one IMMEDIATE transaction ... ctx must carry
  a deadline, which bounds the wait for the write pool's one connection inside this process. The wait on another
  process's write lock is bounded by the busy timeout alone." AC16's concurrency proof and V5's `busy` rest on it.
- `internal/engine/store/conn.go:34-36`: "ErrBusy wraps a failure after which SQLite's write lock stayed held past the
  busy timeout." `additionOutcome` already maps it and `context.DeadlineExceeded` to busy (`admin.go:304-311`).
- `go doc github.com/getkin/kin-openapi/openapi3filter Options`: "Set IncludeResponseStatus so ValidateResponse fails
  on response status not defined in OpenAPI spec."
- `internal/notify/notify.go:80-88`: Send "reports failure to its caller and makes no retry or best-effort policy of
  its own; that decision belongs to each caller." The notice's log-and-commit policy is the caller's (spec, "The
  sensitive-change notice").
- `internal/auth/epoch.go:84-91`: "A bump earlier than it writes nothing, notifies no one, and returns
  ErrEpochBackward ... A bump at the stored epoch's own second has nothing to move and returns nil."
- `internal/config/registry.go:410-414`: AddSetting "panics on a declaration that could never be read: a kind with no
  bound ...".

## Privileged fields and states

The spec's privileged-field table and states table (spec, "Planning-miss prevention") are this plan's, by reference.
Added rows:

| Field | Channels after 14a |
|---|---|
| the API caller on the request context | the mount, after its verifier admits; never a header or body field |
| the notice's recipients | the roster's live admins, read after the commit; never a request field or the environment |
| `DUBPLATE_MEMBER_CAP` and `DUBPLATE_REAUTH_WINDOW` | as every sensitive store-homed setting: `set_setting` inside the window, `settings reset`, the seed rule |

| Call | State | Answer |
|---|---|---|
| any path under `<prefix>/api` | `api`, `api/`, `api/v1`, `api/v2/...` | 404, JSON `not_found`, before verification |
| any `/api/v1` call | verification fails on a setting read or a roster lookup | 500 `internal`, logged at Error |
| any `/api/v1` call | a valid session, same-origin | admitted, and the refreshed cookie set |
| any write | the store's write lock outwaits the busy timeout or the deadline | `unavailable`, detail `busy`, nothing written |
| any member call | `user_id` in uppercase, of the wrong length, or not hex | `invalid_request` |
| add (`add_member`, `member add`) | live members at the cap, a new or re-added address | `conflict`, detail `member_cap`, naming the setting and removal |
| add | live members at cap minus one | commits |
| add | `member add --admin` on a roster with no live member | commits whatever the cap (PS bootstrap) |
| add | the cap cannot be read | `internal`, nothing written |
| the notice | the send fails or outwaits its deadline | logged; the write stays committed |
| `doctor`'s members row | live members above the cap | warns, without failing (PS A1) |

## Review focus

The inputs most likely to bite a person, each pinned in its owning task:

1. A stale session mid-edit: the change survives, and a retry after a sign-in in another tab commits (Tasks 6, 10).
2. The admin demotes or removes themself while another admin remains: the next call answers `forbidden` or
   `unauthenticated` and the page shows that copy, never a crash or a stale list (Tasks 4, 7, 10).
3. A key the serving environment sets: read-only on the page, and a write against it refused before the next start
   could be (Tasks 5, 10).
4. A CLI `member remove` while `serve` runs and the page adds the same address: one write wins, the other answers
   `busy`, `not_found`, or a duplicate, never a half-written row (Tasks 4, 8).
5. A re-added member's old session in another browser stays refused, and a fresh sign-in at least one second later
   verifies (Task 4).

## Lane launch (the conductor, before Task 1)

1. **Live executors.** The lane registry shows no code lane; `git status` in the main tree shows only the
   `dubplate-implementer` agent-memory edits and two 2026-09-04 research files, none code. Confirm no session owns
   them before cutting the worktree.
2. **Contested surfaces written.** Copy the spec's contested-surfaces rows into STATUS's lane registry (the fold's
   conductor-owed item), naming lane 1 as 14a's owner, before the first dispatch.
3. **Lane registry row:** lane 1, the worktree, the branch, this plan, the sub-ceiling (the ceiling).
4. **Provisioning.** `npm ci --prefix web` in the worktree; `make unidecode-wheel` and `make subsonic-xsd` verified
   in the user cache; `xmllint` at libxml 21504 on `PATH`. No venv.
5. **Guards:** the runaway guard, the sleep inhibitor sized to the run's end, the stall guard, the `/loop` wake-up,
   and AC power.
6. **Throughput baseline:** one full gate in the worktree, timed and quoted.
7. **Pre-flight** per segment (pass-core): one `haiku` dispatch checks every claim the segment's tasks make about code
   at HEAD; the conductor amends the plan, then dispatches.

## Task 0: The dependency sweep (conductor, on master)

**Class:** `tool`. **Hook:** neutral. Run in the main tree on master through the `dependency-upgrade` skill
(implementers carry no Skill tool), landed as its own commit before the worktree is cut (the contested-surfaces row
for `go.mod` and `web/package.json`).

**Outcomes.**
- `go.mod`: `golang.org/x/sys` v0.48.0 to v0.49.0 and the indirect `golang.org/x/crypto` v0.57.0 to v0.58.0 (the
  brief's movers), plus any other mover the proxy shows that day. `kin-openapi` v0.149.0, `go-mail` v0.8.1,
  `x/time` v0.16.0, and `sqlite` v1.60.1 are re-read; `modernc.org/libc` moves only with sqlite.
- `web/package.json`: every minor and patch. **TypeScript 6.0.3 to 7.0.2** is taken only if `svelte-check`'s release
  notes or source show TypeScript 7 support and the `web/` leg passes on it. Otherwise it is held, and the record
  names the unblock condition (the `svelte-check` release that supports TypeScript 7). `@types/node` stays on the
  newest 24.x, because Node runs v24.20.0. Neither is a fork.
- `tools/venv/requirements.txt` re-read.
- Pinned binaries: rclone, golangci-lint, and Playwright each checked for a newer release; libxml2 read under
  `brew pin` and bumped only by CLAUDE.md's deliberate sequence, with the XML goldens regenerated.
- A survey line per capability a bump adds: take now, file, or none.
- The record lands at `docs/superpowers/research/2026-10-1x-dependency-sweep.md`.

**Acceptance.** The sweep commit passes `check.sh` (full, heavy) on master; the record lists each version with its
evidence and each held item with its unblock condition.

## Segment A

### Task 1: The network actor on `UserID`

**Files:** `internal/ops/ops.go`, `ops_test.go`; `internal/identity/identity.go` (the doc comment at `:209-211`); the
`opshttp` tests that build a network actor (`helpers_test.go`, `opshttp_test.go`, `contract_test.go`).
**Hook:** lands hook 5 in part. **Gate:** the per-task gate.

**Interfaces.** Produces `ops.NetworkActor(id identity.Identity, anchor time.Time)` and the read-only accessors for
the `UserID` and the anchor, which Tasks 6 and 7 consume. `LocalActor` carries neither.

**Outcomes.** The wire form is `member:<lowercase hex UserID>` (spec, "The network actor (D5)"), and the two comments
the spec names (`identity.go:209-211`, `ops.go:42-46`) say so.

**Acceptance.** AC6's actor clause at the `ops` level: two identities with one `UserID` and different principals
render one wire form. **Fixture:** principals that differ, so a principal-built form fails. The settings-audit half
of AC6 is Task 6's.

**Mutations:** build the wire form from `Principal`; encode the id in uppercase; drop the anchor from the actor
(the accessor test fails).

### Task 2: Hook 18, description and sensitive mark

**Files:** `internal/config/settings.go`, `keys.go`, `thresholds.go`, `registry.go` and their tests; every
declaration site: `internal/config/keys.go`'s wall rows and `cmd/dubplate-server/auth.go`, `backup.go`, `doctor.go`,
`proxy.go`, `serve.go`, `store.go`, `web.go`. **Class:** `tool`, `gateLane: "heavy"`. **Hook:** lands hook 18.
**Gate:** the per-task gate.

**Interfaces.** Produces the closed mark type (no valid zero value) and a description on `Setting`, `Threshold`,
`WallRow`, and each integration-class key. Tasks 6, 7, and 10 read both through the registry.

**Outcomes.** As the spec, "Hook 18", says. `IntegrationClass` names its keys as declarations carrying a description
and a mark, not bare strings (`keys.go:155-160`). `Setting.Doc` keeps its meaning. Every existing key gets its
description and mark in this task, as the contested-surfaces row requires of the first segment. If rung 7 has merged
keys first, this task marks them too.

**Acceptance.** AC8. **Fixture:** one declaration per family missing a description, and one with a zero mark, each
panicking with the key's name; the existing construction test builds the binary's whole registry.

**Mutations:** accept a zero mark; accept an empty description on one family (each family's proof fails alone).

### Task 3: The ops layer: the admin seam, the codes, and the unavailable job seam

**Files:** `internal/ops/` (a new file for the admin seam and its request and response types, a new file for the
unavailable job seam, `catalog.go`, `errors.go`, and their tests); `internal/ops/opshttp/opshttp.go`, `authorize.go`,
and their tests. **Hook:** lands hook 15 in part (the settings operations' vocabulary). **Gate:** the per-task gate.

**Interfaces.** Produces `ops.AdminSeam` with the seven operations (spec, "The mount", the operation list); its
request and response types; the two codes' sentinels; an error carrying an optional `detail`; and the unavailable
job seam's type. Task 6 implements the admin seam; Task 7 mounts the handler.

**Outcomes.**
- The admin seam, the privileges table rows (all seven at admin), and the `catalog` answer joining the job seam's
  entries with one per admin operation through `BuildCatalog` (`catalog.go:111`).
- `unauthenticated` (401) and `reauthentication_required` (403) join `ops.Codes()`, and `checkStatuses` forces each.
- The error body gains optional `error.detail`, carrying P8's enumeration.
- The unavailable job seam answers `start_backup`, `job_status`, and `cancel_job` with `ErrUnavailable` and a reason,
  and its catalog reports each `available: false` with that reason. `local.Runner` gains nothing.
- `opshttp`'s handler routes the seven operations through the authorizing decorator over both seams.

**Acceptance.** AC5's authorizer half (a member caller reaches `catalog` and each admin operation answers
`forbidden`); AC7's `unavailable` and catalog half. **Fixture:** fakes for both seams that record every call, so a
refused call that reached a seam fails the test.

**Mutations:** drop one admin operation's privilege row (the catalog build fails); let a member privilege reach an
admin operation; omit a new code's status (`checkStatuses` fails); report a job operation available.

## Segment B

### Task 4: Removal, demotion, the re-add bump, and the cap in the roster

**Files:** `internal/auth/roster.go`, `roster_test.go`, `epoch.go` (`Subscribe`'s doc comment), `epoch_test.go` if
needed; `cmd/dubplate-server/serve.go` and `member.go`, for the roster constructor's new arguments only; a new
`cmd/dubplate-server/api.go` holding `DUBPLATE_MEMBER_CAP`'s declaration (P7). **Hook:** lands hook 5 in part (the
in-write actor re-read by `UserID`); the epoch stays on the address (D6). **Gate:** the per-task gate.

**Interfaces.** Produces the roster's remove and demote methods naming the target by `UserID`, the exported in-write
admin check (P10), and the cap's read on the add path. Consumes nothing new.

**Outcomes.**
- Removal and demotion as the spec's "Member removal and demotion", "The epoch bump and R-6d20 (D6)", and "The
  last-admin guard (D7)" sections say, inside one `Update` each, the bump after the tombstone and the guard.
- The reactivation branch of `addUser` bumps the address's epoch in the same callback and order (fold C3).
- **V6:** `AddMember`'s cookie-channel actor is the acting `UserID`, re-read by id inside the write as removal does.
  The command line keeps the operating-system username for its log record. `internal/auth` imports no
  `internal/ops`.
- **V7:** the code comment at the `ErrEpochBackward` branch states the reason: the stored epoch is later than now, so
  committing loses no revocation a write could make. The accepted residual is a session anchored between that epoch
  and the clock step, which no bump can reach. Task 11's HISTORY entry repeats it.
- `Subscribe`'s doc comment gains the must-not-block contract (spec, last paragraph of "The epoch bump").
- The roster constructor takes the epoch store and a clock; `serve` and `member` pass them.
- **The cap** (P7): the add path counts live members, admins included, inside the add transaction, and refuses at the
  cap with `member_cap` naming the setting and removal. A tombstoned row does not count. A reactivation counts. The
  first-admin bootstrap is never refused. A cap that cannot be read refuses, never a default.

**Acceptance.** AC15 and AC16's roster halves. PS A1's roster rows: at cap minus one the add lands, at the cap it is
refused, and the bootstrap lands. The states table's remove, demote, and add rows at the roster level.
**Fixtures:** AC15's (a preference, an override, and a quota row written first; the removal's epoch record deleted
before the re-add; the clock injected one second past the re-add's bump); AC16's (the second connection holding the
write lock with one demotion uncommitted); a roster at the cap holding one tombstoned row, so a count that includes
tombstones fails.

**Mutations:** move the guard outside the transaction (the concurrency proof fails); bump after the commit; skip the
re-add bump; roll back on `ErrEpochBackward`; delete overrides or quota on removal; read the actor's flag by address
or outside the `Update`; count tombstoned rows toward the cap; refuse the bootstrap at the cap.

### Task 5: The settings write path

**Files:** `internal/settings/` (a new write file and its test, and `evaluate.go` for BACKLOG #25 at `:265`).
**Hook:** neutral (hook 9: the write rides the one-writer seam). **Gate:** the per-task gate.

**Interfaces.** Produces the exported set and reset write functions. Each takes the registry, the serving
environment, the actor's wire form and channel, and the in-write guard (P10, absent on the command line). Each
returns what it wrote or why not, the source in force, the apply mode, and whether the key is sensitive. Task 6 and
Task 8 are the callers.

**Outcomes.** As the spec's "The settings write path" through "Rate limits stay restart-applied" say: one validator
(the written key's own check, then the seed rule's per-key comparison over `Outcome.Err`, never `Evaluation.Err`), the
`admin` source, `seed_value`, the generation bump, one `settings_audit` row, the environment refusals, and the no-op
answers. BACKLOG #25: the parse failure is wrapped, not the nil `Validate` result.

**Acceptance.** AC10 and AC11 at the function level, and the states table's `set_setting`, `reset_setting`, and
per-key comparison rows. **Fixtures:** an `env-seed` row (so `seed_value` records the seed); an `admin` row (so it
keeps the earlier `seed_value`); a store already holding one out-of-bounds `admin` row beside a valid write to another
key (AC11's last clause); and two keys under one correlated rule.

**Mutations:** compare `Evaluation.Err` instead of per-key outcomes (the existing-refusal fixture fails); skip the
written key's own bounds check; write over an environment-set key; bump the generation on a no-op; overwrite
`seed_value` on an `admin` row.

### Task 6: The admin seam's implementation, the gate, and the notice

**Files:** a new `internal/ops/admin/` package and its tests; `cmd/dubplate-server/api.go` (P6's declaration and its
correlated rule). **Model:** opus. **Hook:** lands hook 15 in part (settings and members on the cookie mount's seam).
**Gate:** the per-task gate.

**Interfaces.** Consumes Task 1's actor accessors, Task 3's admin seam and codes, Task 4's roster methods, and Task
5's write functions. Produces the admin seam for `serve`, which Task 7 mounts.

**Outcomes.**
- The seven operations, each over the roster or the write functions. `list_settings` answers as the spec's "Hook 18"
  last paragraph says, with no wall or integration value ever on the wire.
- The re-authentication gate decided here, after decoding, from the actor's anchor and the key's mark, for exactly the
  calls the spec's "The re-authentication gate" names. An unreadable window refuses with `internal`.
- The notice as the spec's "The sensitive-change notice" says: after `Update` returns, under its own deadline, to
  every live admin, logged on failure. It names the key, the actor's principal, and the old and new values.
- `user_id` parsing per the hex quote above. **V5:** `store.ErrBusy` and `context.DeadlineExceeded` from any write
  answer `unavailable` with detail `busy`.

**Acceptance.** AC12, AC13, AC6's settings-audit half, AC9's content half, AC16's settings-write actor clause,
AC17's seam half (the add request type has no admin field), and every states-table row reachable at the seam.
**Fixtures:** AC12's sessions anchored inside and outside the window, plus one refreshed after its anchor aged out;
AC13's two live admins and one removed admin behind a recording mailer; AC6's member renamed by a direct store write
between two writes; a mailer that fails, to prove the write stays.

**Mutations:** gate on `IssuedAt` instead of the anchor; skip the gate for `add_member`; send the notice inside the
write callback (a blocking mailer stalls a concurrent writer; the test fails); mail a removed admin; fall back to a
default window on a read failure; accept an uppercase `user_id`.

### Task 7: The mount and the contract document

**Files:** `internal/ops/opshttp/` (the exported mount, `openapi.yaml`, the contract tests); a new
`internal/auth/` file for P2's session check and its test; `cmd/dubplate-server/api.go` (registration of `api`
through `Shell.Register`, and the verifier); `cmd/dubplate-server/serve.go` (one call into `api.go`); the route
inventory test if no lane has landed it (PS A2). **Hook:** lands hook 15 (the cookie mount). **Gate:** the per-task
gate.

**Interfaces.** Consumes everything above. Produces the `/api/v1` surface the pages call in Task 10.

**Outcomes.**
- The mount, the carrier, the verifier, and the 401/500 split per P2 and P3. One log record per admitted operation
  with actor, channel, and `client_ip`. No bearer header is ever read.
- The document per the spec's "The contract document (D4)", with every operation, body, code, and detail, the version
  policy, and no `security` requirement.
- The contract tests per AC1 and AC2: operation-path parity, the keyword and `format` walk, `IncludeResponseStatus:
  true` with every other silent pass failing (on the `conformance.go:96-136` precedent), the closed loaded copy, and
  the three named proofs.
- **V4:** one test marshals the job response values (every state, a failure, every timestamp) and validates each
  against the document's schema directly, since no routed answer carries them on this mount.
- **The route inventory (PS A2):** if no lane has landed it, this task lands it over the real serve shell, every route
  with its auth class (public, session, admin), and no delegated row (PS erratum, 2026-10-05). If rung 7 landed it
  first, this task adds `api` and keeps `roster` until Task 10 removes it.

**Acceptance.** AC1 to AC5; AC7's document half; AC17's HTTP half (a body naming `admin` adds the member without the flag); AC9's raw-body sentinel scan over every `/api/v1` answer, refusals
included; AC18 for every HTTP-reachable state; the states rows above for `api` paths and the 500. **Fixtures:**
sentinel values on every wall and integration row; a live member session and a live admin session; a session whose
cap setting is made unreadable.

**Mutations:** drop `IncludeResponseStatus`; leave a response object open in the loaded copy (the rename proof
passes, so the test fails); answer an unknown `api` path in plain text; answer a setting-read failure with 401; read
an `Authorization` header.

## Segment C

### Task 8: The CLI verbs

**Files:** `cmd/dubplate-server/member.go`, `member_test.go`; a new `cmd/dubplate-server/settings.go` and its test;
`main.go` (registration); `internal/settings/evaluate.go` (the three seed-rule messages name the verb); the file
holding `serve`'s pre-open older-store refusal (`cmd/dubplate-server/store.go` or `serve.go`). **Hook:** neutral.
**Gate:** the per-task gate.

**Outcomes.**
- `member remove <address>` and `member demote <address>` over the roster methods, resolving the address to the id
  first; a target removed between the read and the write answers not found. The verbs open the auth leaf for the
  epoch store, as the runbook's `systemd-run` form already grants (`docs/server-runbook.md:190`).
- `settings reset <key>` per the spec's "The resolving verb", through `store.OpenWrite`.
- BACKLOG #12: every required flag's usage says "(required)", on the four verbs.
- The seed rule's three `admin`-row messages name `settings reset`, and `serve`'s pre-open refusal of an older store
  over an `admin` row names the previous-release recovery.

**Acceptance.** AC14; AC15's and AC16's command-line halves; the states table's `settings reset` rows. **Fixtures:**
a current store holding each of the three refusing `admin` rows, then `serve` starting; two out-of-bounds rows reset
one at a time; an older store, a newer store, and a restore intent.

**Mutations:** open through anything but `OpenWrite`; skip the validator on the CLI reset; drop the verb name from one
seed-rule message.

### Task 9: `DUBPLATE_ADMIN_EMAIL`, doctor, and the runbook

**Files:** `cmd/dubplate-server/doctor.go`, `doctor_test.go`; `internal/engine/store/inspect.go` and its test;
`docs/server-runbook.md`. **Class:** `tool`, `gateLane: "heavy"`. **Hook:** neutral. **Gate:** the per-task gate.

**Outcomes.**
- `Inspect` reports the roster standing of one address its caller names (live admin, live non-admin, removed, none),
  and no address enters `store.Facts` (`inspect.go:166-171` carries counts only; spec, last paragraph of
  "`DUBPLATE_ADMIN_EMAIL` (D9)").
- `doctor`'s `DUBPLATE_ADMIN_EMAIL` row per AC21 and P9. doctor has no warning state today (its row functions return
  a fail flag, `doctor.go:222-235`), so this task adds a non-failing warning state; `doctor`'s exit status ignores it.
- The members row warns when live members exceed `DUBPLATE_MEMBER_CAP` (PS A1, "an over-cap roster loads with a
  `doctor` row").
- The runbook per the spec's three edits (step 10, the key description at `:59-60`, the restore line), plus one line
  naming `member remove`, `member demote`, and `settings reset` as verbs run through the same `systemd-run` form.

**Acceptance.** AC21. **Fixtures:** a store per standing (a live admin, a live non-admin, a tombstoned row, no row),
and an empty roster, which adds no second warning; a roster one above a lowered cap.

**Mutations:** fail the exit status on the warning; let an empty roster add the address warning; merge the live
non-admin message into the removed one.

### Task 10: The members and server settings pages

**Files:** `web/src/routes/members/+page.svelte`; a new `web/src/routes/server-settings/+page.svelte`; a new
`web/src/lib/api/` client module for `/api/v1` (14a's); `web/src/lib/copy/deck.ts` (appended rows);
`SiteHeader.svelte` (one appended navigation entry); `cmd/dubplate-server/web.go` (the document row, appended),
`serve.go` (the page's registration behind `RequireAdmin`, and the `roster` registration removed); `internal/auth/admin.go`
(the roster API's handlers removed, nothing in `RequireAdmin`); the rendered harness's rows for the new document.
**Class:** `engine-logic`; the scoped gate runs the rendered leg. **Hook:** neutral (consumes hook 15).
**Gate:** the per-task gate.

**Outcomes.**
- The members page lists, adds, removes, and demotes over `/api/v1` through the new client, confirms before a removal
  or demotion, and shows each refusal's own copy, the member cap's included. Each add refusal moves focus to its
  field. The client reads `error.code` on a JSON 403; only the origin check's plain-text 403 is cross-site.
- The server settings page per AC20 and P4, with P5's stale-session flow.
- The roster API retires: `roster` answers 404, and BACKLOG #23 closes with it (spec, "Riders").
- Both pages follow the visual identity spec and pass the `web/` leg's static and rendered gates.

**Acceptance.** AC17's `roster` 404, AC19, AC20, AC22's #23. **Fixtures:** the rendered harness's admin session
against a roster with two admins and one member, so the last-admin refusal and the demote control both render; a
stale session for the re-authentication state; an environment-set key for the read-only row.

**Proofs:** a JSON 403 rendering cross-site copy fails a component or rendered test; each refusal state renders its
own copy; the confirm step blocks the call until confirmed.

## Close

### Task 11: Reviews, verification, merge, and the ledgers

The conductor's steps, in order (pass-core, "Closing"; CLAUDE.md, "Lanes" and "Reviewers"):

1. **Security read after segment B** (at that boundary, not here): `web-auth-security-reviewer` over the write path,
   the gate, the notice, the mount, and the session check, then one fold task. **After segment C:** a second read
   over removal's CLI half, the pages, and the roster API's retirement, then one fold task. Both ask for correctness
   gaps ranked by consequence; a fold may refuse a finding with a one-line reason.
2. **Simplify once:** `code-simplifier:code-simplifier` over the pass's Go, TypeScript, and Svelte, with
   `.claude/instructions/simplifier-brief.md` pasted verbatim, then its `diff-reviewer`.
3. **Full gate:** `check.sh` (heavy) in the worktree. On red, revert the simplifier commit and re-gate first.
4. **UI reviewers:** `svelte-reviewer` and `daisyui-a11y-reviewer` over both pages, each passed the visual identity
   spec's path and asked to grade conformance (tokens, inks, the component layer, the copy register) beside
   correctness.
5. **Live auth smoke** on the workstation, on the real binary under `env -i`, with the loopback SMTP sink (6d's
   method, HISTORY 6d, "The gate evidence"). It covers: sign-in; `catalog` with the three job operations unavailable;
   `list_settings`; a non-sensitive write; a sensitive write from a fresh session, with the notice at the sink; the
   window seeded at its `5m` floor and outwaited, then a sensitive write refused; `add_member`, `remove_member` (the
   removed member's next request refused), a re-add (the old session still refused), `demote_member`, and the
   last-admin refusal; `roster` and `api/v2/x` answering 404; and `settings reset` with `serve` stopped clearing an
   out-of-bounds `admin` row written directly.
6. **Render and read** both pages in both palettes at phone and desktop widths, in each refusal state, graded
   fresh-context, before Geoff sees them.
7. **Merge ritual:** merge token, rebase onto master, `check.sh --merge` in the worktree, `--ff-only`, `check.sh` on
   master, and one `diff-reviewer` on the merged result if the rebase was not a fast-forward.
8. **Architecture reads:** one `go-architecture-reader` per touched package, no plan in context: `internal/ops`,
   `opshttp`, `internal/ops/admin`, `internal/auth`, `internal/settings`, `internal/config`, `internal/identity`,
   `internal/engine/store`, and `cmd/dubplate-server`. One fold task.
9. **Docs fold** (one `sonnet` fold agent, one independent `diff-reviewer` over its diff):
   - Errata 1 to 8 (the fold, "Owed errata" and "continued"), each a dated erratum note in its ratified document:
     R6:913-915, R6:589-590, R6:943-944, R6 criterion 7 with erratum row 10, CH:171-172, PA:173-215, BS:983, and
     R6:582. PS gains a note that the member cap landed at 14a.
   - BACKLOG: #12, #23, and #25 closed; C14's rung 16 check filed (the fold's conductor-owed list).
   - STATUS, present tense: the routing of `serve`'s job runner, instance lease, and job recovery to the first rung
     whose page starts a job (erratum 3); 14b's trigger; the contested-surfaces rows released; lane 1 idle.
   - HISTORY: what landed per task; what the gates caught; both security reads' dispositions; the fold-rule trial
     (any refused spec finding that proved a real defect, or none); V7's residual; "what a later pass would be wrong
     to rediscover" (the 401/500 split, the per-key comparison, the epoch ordering and the re-add bump, why the job
     operations answer `unavailable`, the uppercase-hex trap, and P5's no-return-parameter reason); and the pass
     score: tokens against 10.5M, attended time, and clock against about 6 hours (total, gate time, lock wait, and
     rework with each red's cause).
   - **The review-growth measure:** the spec grew from 419 to 651 lines across two folds (55%), past the 25% signal.
     The entry names the cause (the states table, the contested-surfaces table, the quoted kin-openapi behaviors,
     and V1's admin seam) and whether the growth bought execution-time savings.
10. **Commit** specific files; pre-bake STATUS's resume prompt; close the session.

## Prior art consulted

The spec's "Prior art" table is this plan's, by reference: kin-openapi adopted in tests only; go-mail through
`internal/notify`; `x/time/rate` kept with live apply ruled out; the stdlib MACs, sqlite, and goose neutral; GitHub
sudo mode adopted for its rule and P6's default, declining its timer reset; and the OpenSubsonic specification,
Navidrome, gonic, taglib, mewkiz/flac, Picard's tables, and chromaprint neutral. The member cap adopts no library.

## Rulings for Geoff

None. The member cap's placement is a sequencing call taken above with its reason, not a product fork: PS ruled the
mechanism and its default.

## Errata (filled during execution)
