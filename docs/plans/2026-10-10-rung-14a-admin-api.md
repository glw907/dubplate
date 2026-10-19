# Rung 14a (the Admin API's First Half) Implementation Plan

Status: **revision 2.1, 2026-10-10: the three `spec-plan-review` lenses folded, then the verification read's V1 to
V10** (`docs/research/2026-10-10-rung-14a-plan-fold.md`, "the plan fold", and its "Second fold"). Amended
2026-10-10 for the gate scope pass: the per-task gate, its clock, the worktree cut, and lane launch step 8.

**Spec:** `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md`, revision 2.1 at `c7c3fbb` ("the spec").
Its fold record is `docs/research/2026-10-10-rung-14a-spec-fold.md` ("the fold"); this plan places the
fold's owed items V4 to V7, doctor's warning split, errata 1 to 8, and the conductor-owed list. Task sketch, ceiling,
and clock come from `docs/research/2026-10-10-next-pass-planning-brief.md` section 6 ("the brief").
The plan cites the spec and never restates it. Spec abbreviations (BS, RP, SS, CH, R6, PA) carry over; PS is the
group-scale principles spec (`docs/specs/2026-09-25-group-scale-principles-design.md`). Code pointers are
at `dd161b3`; the pre-flight before each segment re-resolves them.

**Goal.** Land the `/api/v1` cookie mount with its contract document, the admin settings write path behind the
re-authentication gate, member removal and demotion over the CLI and the API, hook 18, and the two pages that call
them.

**Architecture.** `internal/ops` gains a second interface, the admin seam, beside the job seam. A new
`internal/ops/admin` package implements it for `serve` over the roster (`internal/auth`) and a new settings write
function (`internal/settings`). `opshttp` gains its first exported surface, the mount, and owns every answer under
`<prefix>/api`. The CLI verbs call the roster and the settings write function directly, which are the shared write
path (spec, "The mount").

**Pass class:** `auth-data`. Overrides: Task 0 and Task 2 are `tool`; Task 9 is `tool` (a read-only `doctor` row, an
`Inspect` fact, and runbook text, with no write path); Tasks 10a and 10b are `engine-logic` with the `web/` rendered
leg. Task 10c, the roster API's retirement, keeps `auth-data`.
**Ceiling:** 11.8M execution tokens, flag 9.4M (80%). Planning and plan review sit outside it.
**Clock estimate:** about 6.8 hours of execution, excluding owner pauses; `auth-data` fix rounds push it toward 8.3
(re-priced below for the gate scope pass).
**Checkpoint interval:** four tasks at most: each segment boundary and the close.
**Segments:** Task 0 (conductor, on master). A: Tasks 1, 2, 3. B: Tasks 4, 5, 6, 7, then the first security read and
its fold. C: Tasks 8, 9, 10a. D: Tasks 10b, 10c. Then the close, Task 11. Boundaries A, B, and C end on a green full
gate on the heavy lane; D's yields to the close gate (pass-core, "Closing" step 2).

### Ceiling and clock, re-derived

The brief's 9.5M priced ten tasks (Task 0 and nine build tasks) at 0.6M plus 3.0M of overhead. Revision 2.1 of the
spec adds work the brief did not price: the admin seam and the unavailable job seam (V1), the `error.detail` field and
two wire codes, the keyword walk, the re-add bump, the roster API's retirement into `list_members` and `add_member`,
the runbook and `doctor` edits, and the member cap this plan takes (below). It drops the older-store write mode and
every migration. The plan fold split revision 1's Task 10 into three tasks, because the retirement reaches the deck
gate, the `RequireAdmin` test table, `serve_test.go`, and the rendered harness's members rows (the plan fold, B1).

The re-cut is twelve build tasks plus Task 0: 13 x 0.6M = 7.8M, an Opus premium on Task 6 of 0.3M, about 0.2M for the
fixtures the plan fold added to Tasks 2 and 4 to 8, and an overhead of 3.1M. The overhead is the brief's 3.0M plus
the architecture read for `internal/ops/admin` and `internal/engine/store`, less the read `internal/identity` no
longer takes (Task 11). The sum is 11.4M; 11.8M leaves most of one more fix round.

The clock follows the brief's rates: thirteen tasks at about 22 minutes (2b's 20, plus the rendered leg every task
here runs), three boundary gates at about 9 minutes, and a close of about 1.65 hours. The close carries thirteen
reviewer reads, one combined fold, the simplifier, the live smoke, the render-and-read, and two full gates (the close
gate and the on-master gate; the plan fold removed a third).

**Re-priced for the gate scope pass.** Every 14a task still runs the rendered leg: each touches a package the
rendered test binary links (`internal/{ops, identity, config, auth, settings}`, `cmd/dubplate-server`, `web`), so the
scoped form's skip never applies here and the 22-minute rate holds for Tasks 1 to 8. The savings fall on Tasks 9 to
10c, which the old form ran with the conform suite whole and the bundle built twice:
- Old: 4 tasks x (conform about 150s cold + a second bundle build, estimated at 30s) = 4 x 3.0 = 12 minutes.
- New: conform's closure (`go list -deps -test ./internal/harness/conform`) excludes `internal/engine/store`, so the
  suite replays from Go's cache after one cold run, about 150s at the first gate in the fresh worktree (Task 1;
  without `-trimpath` the build action IDs include the package directory). The bundle builds once per gate: Tasks
  10a to 10c edit `web/scripts/rendered/run.ts` or `web/scripts/gates/deck-coverage.ts`, which are gate machinery, so
  their scoped run falls back to the full form and runs `make check`, which builds the bundle once while `go test`
  reads from the cache. Total 2.5 + 4 x 0.5 = 4.5 minutes, against 12 old plus the 2.5 that Task 1 would not have
  paid, so the net is about 9.5 minutes.
- So 7 hours (420 minutes) becomes about 410 minutes, or 6.8 hours. The boundary gates and the close are full-form and
  unchanged, and so is the fix-round increment (1.5 hours); the allowance moves from 8.5 to 8.3 hours with the base.
  The 11.8M token ceiling stands: the gates change wall time, not the tokens a task spends.

## Tree, branch, and execution

**Tree.** Lane 1 runs in its own worktree, `.claude/worktrees/lane-1-rung-14a`, on branch `lane-1-rung-14a`, cut from
master after the gate scope pass and Task 0 have landed. This departs from CLAUDE.md's "the primary code lane ... share[s] the tree", for three
reasons. pass-core requires the pass's branch or worktree, "never the default branch's checkout". The merge ritual
(rebase, `--ff-only`) presumes a branch. And the main tree stays on master for the rung 7 planning lane and the
conductor's ledger commits. The CLAUDE.md rule exists so two code lanes never share one gate's `./...`, and this
layout keeps that. The pass needs no venv, no corpus, and no beets, so nothing pins it to the main tree. dubplate has
no CI (`.github/workflows` is absent), so pass-core's draft-PR shadow does not apply.

**Executor:** `dubplate-implementer`, `model: sonnet` (the runner passes only a model, `pass-execute.js:857`, so the
effort is the agent's own), except Task 6 on `model: opus` from the start (the gate's decision point, the notice's
ordering, and seven operations over two write paths).

**Per-task gate:** the runner's classifier (`scripts/checks/gate-tier.mjs`, run from the worktree) computes it from
the base the runner recorded before the task's first dispatch: `bash <worktree>/scripts/check.sh --scope-changed
--base <sha>`. The plan's string stays as the fallback when the classifier refuses or the base is unusable:
`cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/.claude/worktrees/lane-1-rung-14a/scripts/check.sh --scope-changed'`.
Either form is re-issued on exit 75 until it prints `gate exit:`, on the heavy lane for every task. With `--base`, the
change set is the task's own diff; the fallback string still diffs from `git merge-base master HEAD`, so its set is
cumulative over the pass. Task 1 edits `internal/identity/identity.go`, and every task after it touches a package the
rendered test binary links (`go list -deps -test -tags rendered ./cmd/dubplate-server`; `opshttp` joins at Task 7's
mount, with no `check.sh` edit), so every task's scoped gate runs the rendered leg and launches Chromium. The conform
suite is no longer scoped by hand: it replays from Go's test cache unless its inputs changed. Tasks 10a to 10c edit gate machinery (`web/scripts/rendered/**`, `web/scripts/gates/**`), so their scoped run falls back to the full form.
`pass-execute`'s `tool` class defaults to the light lane (`PASS_CLASSES.tool.gateLane`), so Tasks 2 and 9 carry
`gateLane: "heavy"` (HISTORY 6d: "Pass `gateLane: "heavy"` for a gate that launches Chromium"); both still launch the
browser, because Task 2 edits `internal/config` and `cmd/dubplate-server`, and Task 9 edits `cmd/dubplate-server` and
`internal/engine/store`. `reducedGate` is the `--reduced` form, which under `auth-data` only a comment-only round
takes. Tasks 10a and 10b still set their own `reducedGate` to the per-task gate string: `engine-logic` lets a test-only
round reduce (`testOnlyReduces`, `pass-execute.js:310-316`), and `--reduced` never runs the rendered leg, which is the
leg whose rows those rounds fix. (`--reduced` now runs `svelte-check` on a `web/` change, so that is no longer part of
the reason. The rows live in `web/scripts/rendered/**` and `web/scripts/gates/**`, which are gate machinery and
promote a reduced run to the full form anyway; a round that touches only `cmd/dubplate-server/rendered_test.go` would
not, hence the override stays.) Boundaries take `check.sh` (full, heavy); the close takes `check.sh --merge` in the
worktree, then `check.sh` on master.

**Mode:** `pass-execute` by name, one invocation per segment, sequential, with `repo` set to the worktree path,
`implementer: "dubplate-implementer"`, `reviewer: "diff-reviewer"`, `reviewerModel: "claude-opus-5-5"`, `maxFix: 1`,
`stopOnEscalate: true`, `passClass: "auth-data"`, and each task's own `passClass`, `gateLane`, `reducedGate`, and
`model` as above. The plan's gate string is passed as `gate` and acts as the classifier's fallback; the runner reads the
classifier from the worktree, so it names the worktree's `check.sh`. No task is marked independent: every task from 3 on reads the one before.

### The unattended run

The conductor runs Task 0, segments A to D, and the close without an owner touchpoint, holding the sleep inhibitor,
the stall guard, and the fallback `/loop` wake-up per `~/.claude/docs/unattended-work-guards.md`.

**The hard ceiling.**
- At the flag (9.4M) the in-flight chain finishes, STATUS is written with the resume prompt, the current segment
  finishes, and no new segment starts. The first security read and its fold belong to segment B, so segment C never
  starts without them; a flag stop at boundary B names that read as the first resume step.
- The close is not a new segment: once it starts, it runs to the hard ceiling.
- The merge ritual (Task 11, steps 4 to 8, from the merge token) starts only with at least 0.5M left. Once started it finishes, or it stops
  before `--ff-only`, never between `--ff-only` and the on-master gate.
- At 11.8M the conductor stops after the in-flight chain returns. STATUS is written at each boundary and before any
  stop, naming the step reached.

**The conductor rules alone, recording each in its notes and later in HISTORY:** accept; one re-dispatch on `fix`;
after that round, one more dispatch on a `fix` still standing, either an Opus upshift (Tasks 4, 7, and 10c first; Task
6 is already Opus) or a split on the seam the finding names (Task 4's pre-approved seam is the member cap, which
spec deviation 1 shows is separable: its code, fixtures, and mutations move to a Task 4b); a flake the rerun rule covers; a pre-flight claim that is
wrong at HEAD, amended in the plan before dispatch; any decision a task marks "the implementer's"; a reviewer finding
refused under CLAUDE.md's fold rule, with its one-line reason; and a dependency held by Task 0's rules.

**These stop the run (STATUS written, then stop):** the hard ceiling; a `fix` still standing after the extra dispatch
above; an `escalate` the conductor cannot resolve inside the task's files; a warm change in the worktree the conductor
did not dispatch, or another executor writing a surface the contested-surfaces table gives to 14a; a gate red on a
commit whose diff does not explain it, after one re-run; a security finding at High or above that needs a design
change; an architectural question the spec does not settle (pass-core); a close-gate red that the simplifier's revert
does not clear; a smoke or page defect still red after the close's one fold (Task 11, step 2); and a red on-master
gate, with STATUS naming the commit and no revert without Geoff.

**Owner-gated steps, batched at the end:** none needs Geoff's go. The live auth smoke and the render-and-read of both
pages are Claude's verification (pass-core, "Verification a plan parks for the owner is Claude's"), run at the close
before Geoff sees anything. Their evidence and the pages' captures go to Geoff in the close's one summary. The
summary also lists the PS:145-147 erratum (spec deviation 1) as decided, with its reason: the spec's `:284`, "Neither
the seed rule nor the command line sends a notice". Nothing deploys, and the box is never contacted.

## Global constraints

- **Territory.** The spec's contested-surfaces table ("Contested surfaces with the rung 7 lane") is this lane's
  ownership record, by reference, amended by the second spec deviation below. 14a writes nothing in `internal/serve`,
  edits no `RequireAdmin` line, and leaves `post.ts`, `SiteHeader.svelte`'s structure, and the rendered harness's
  shape to rung 7.
- **No migration**, no new module dependency outside Task 0's sweep, and kin-openapi in `_test.go` files of `opshttp`
  only (spec, "The contract document"; the no-link test stays green).
- **Layering.** `internal/ops` and `opshttp` import no `internal/auth` or `internal/settings`; `internal/auth` and
  `internal/settings` import no `internal/ops`, and `internal/auth` imports no `internal/settings` or
  `internal/config` (it imports neither today). Only `internal/ops/admin` and `cmd/dubplate-server` join them (fold,
  V1's verification).
- **No export ahead of a caller** (CLAUDE.md, 2026-09-12). Every new export names its production caller in the task
  that adds it.
- **Config keys** are declared from 14a's own file, `cmd/dubplate-server/api.go`, registered through one
  `newProductionRegistry` line (`main.go:99-108`), never by editing `internal/config/keys.go`'s rows beyond hook 18's
  descriptions (CLAUDE.md, "Facts agents get wrong").
- **Go** under `go-conventions`; **web** under `svelte-conventions`, `ts-conventions`, the `daisyui` skill, and the
  visual identity spec (`docs/specs/2026-09-24-visual-identity-design.md`); copy in plain terms.
- **Every proof names its fixture state**, and every `auth-data` task names its mutations; the implementer reports
  each failing. `web/` has no component test runner (no vitest in `web/package.json`), so every page proof is a named
  rendered manifest row that says what it intercepts and what it reaches for real.

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
  (`admin.go:110-133`), and reports a refused session distinctly from a server failure. The verifier builds the
  actor from `Session.AnchoredAt`, never `IssuedAt`, which a refresh moves (quoted below).
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
  `reauthentication_required` a page keeps the admin's pending change, and its sign-in link opens the sign-in page in
  a new tab (`target="_blank" rel="noopener"`), so the pending change survives in the first. The page offers a retry.
  The cookie the link sets in the other tab serves the retry. The copy says to sign in from this browser.
- **P6. The re-authentication window** is `DUBPLATE_REAUTH_WINDOW`: a duration setting, store-homed, sensitive, apply
  mode next-operation. Its default is `2h`, GitHub sudo mode's window (spec D10's source). Its floor is `5m`: the
  anchor is set when the link is consumed (`session.go:80-86`), so the round trip that matters runs from that click
  to the retry, and the floor keeps a mis-set window from locking the web out of resetting it. Its ceiling is
  `24h`. A correlated rule, registered as `sessionLifetimeWithinCap` is (`cmd/dubplate-server/auth.go:325`), refuses
  a window at or above `DUBPLATE_SESSION_ABSOLUTE_CAP`. It is strict: `durationAtLeast` admits equality (quoted
  below), so the rule is not that helper reused as is.
- **P7. The member cap** (spec deviation, below) is `DUBPLATE_MEMBER_CAP`: an integer setting, store-homed, sensitive,
  default `20` (PS ruling 4 at `:729`, ruled as recommended per PS's status line at `:3`). Its floor is `1` and its
  ceiling `10000`, because `AddSetting` panics on "a kind with no bound" (`internal/config/registry.go:410-414`) and
  PS lets the operator raise it to any practical number.
- **P8. Wire names.** A member is `user_id`, 32 lowercase hex characters. The `error.detail` enumeration is every
  roster-write outcome the members page keys on (`web/src/routes/members/+page.svelte:46-54`), plus `last_admin`,
  `environment_sets`, `busy` (V5), and `member_cap`. `admin_not_grantable` and `unknown_field` leave it, since a
  decoder that ignores unknown keys gives neither a cause (V5; `opshttp.go:141-144`).
- **P9. doctor's `DUBPLATE_ADMIN_EMAIL` warning has two messages, not three** (the fold's over-ceremony item 2). A
  removed member and an address no member holds name one action, `member add <address> --admin` or pointing the
  variable at a live admin, since `addUser` reactivates a tombstoned row with the flag (`roster.go:395-419`). A live
  non-admin names another: no promote verb exists and `member add` refuses a live address, so the only action is to
  point the variable at a live admin. All three states still warn (AC21). A warning is the row's status text with
  `failed` false, as the members row's advisory status already is (`doctor.go:223`); doctor gains no new row state.
- **P10. The roster's in-write admin check is one exported roster function over a `*store.Tx`.** The settings write
  takes it as a guard parameter from `internal/ops/admin`, so the users table's SQL stays in `internal/auth` and
  `internal/settings` imports no `internal/auth`.
- **P11. The roster's write dependencies reach only the write paths.** The epoch store, a clock, and the cap reach
  the add, remove, and demote paths; `member list` and every read build no epoch store and open no auth leaf. The cap
  arrives as an `auth.Setting[int]` built in `cmd/dubplate-server` from the accessor, on the precedent of
  `durationSetting` and `authBodyCap` (`cmd/dubplate-server/auth.go:499-516`); `serve.go` and `member.go` (which
  already holds `reg`) both build it. The cap reader is called before `Update`, as the accessor's contract requires
  (quoted below), and its result and error are held. Inside `Update`, a roster with no live member takes the
  bootstrap and ignores the held result (PS:140, "never refused"); otherwise a held error refuses with `internal`, and
  the live count is taken there. No roster read happens outside the transaction, since "no live admin" is decided
  inside `addUser`'s `Update` today (`roster.go:426-432`).
- **P12. `user_id` parsing** accepts exactly 32 characters that equal their own lowercase hex re-encoding. The empty
  string and the all-zero id (the `UserID` zero value, "never resolved", `identity.go:205-208`) answer
  `invalid_request`.

### Spec deviation 1: the member cap lands here

PS routes "Member cap on every add channel" to "Rung 14 settings; the roster's add path" (`:580`), with acceptance
row A1 (`:619`) and the default ruled at 20 (ruling 4, `:729`; ruled at `:3`). The spec omits it, and 14a rebuilds the
web add channel as `add_member`. Shipping that channel without the cap would leave the ratified mechanism unbuilt
behind a new door. 14b builds collections and no add channel, so 14a is the cap's rung. The cap rides Tasks 4, 9, and
10a and adds no task. If the conductor routes it to 14b instead, those three tasks drop their cap clauses and nothing
else moves. `member import` retired at 6d and `dubplate import navidrome` does not exist, so A1's channels at 14a are
`member add` and `add_member`. PS:145-147's "raising the cap from the command line ... notifies the admin" has no 14a
channel: the command-line route to the cap is the environment file and the seed rule, which sends no notice (the
fold's C9 and erratum 5; spec AC13). Task 11 records a PS erratum saying so.

### Spec deviation 2: the roster API's retirement edits rung 7 surfaces in place

The spec's contested-surfaces rows give 14a only appended rows in the rendered harness and no part of the deck gates.
Retiring the roster API cannot stay inside that grant:
- `web/scripts/gates/deck-coverage.ts:155-167` parses `admin.go`'s outcome constants and hard-requires
  `admin_not_grantable` and at least 21 outcomes, so the handlers' removal turns the `web` leg red.
- `internal/auth/admin_test.go` builds its `RequireAdmin` refusal table on `NewRosterAPI` (`:75`, `:265`).
- `cmd/dubplate-server/serve_test.go` drives `/roster` in about ten tests (`:810-1208`).
- The members rows and helpers in `web/scripts/rendered/run.ts` (`:866`, `:1050-1090`, `:3086-3151`) and
  `cmd/dubplate-server/rendered_test.go` (`:218-219`, `:860`) intercept or prove the roster API.
14a therefore edits these in place, in Tasks 10a, 10b, and 10c: the members rows and their helpers, deck-coverage's
roster outcome source, its detail-constant pattern (10a), its `REQUIRED_TITLES` list (10b, `deck-coverage.ts:43`),
and the host of the `RequireAdmin` table (every refusal row kept). It changes no `RequireAdmin` line,
no harness shape, and no other deck rule. Rung 7 is a planning lane today, so this is a coordination note: the lane
registry row records it (lane launch, step 3), and the spec owes an erratum (the plan fold).

## Behaviors this plan relies on, quoted

The spec's quotes stand (spec, "Planning-miss prevention", first bullet). Added here, read on 2026-10-10:

- `go doc encoding/hex.DecodeString`: "DecodeString expects that src contains only hexadecimal characters and that
  src has even length." It accepts uppercase: `hex.DecodeString("ABCD")` returned `[171 205] <nil>` on go1.27.2, and
  `hex.DecodeString("")` returns an empty slice and nil. P12's length check is therefore separate from the
  re-encoding check.
- `go doc strings.CutPrefix`: "If s doesn't start with prefix, CutPrefix returns s, false." The mount's cut branches
  on `found`.
- `go doc context.WithValue`, quoted under P2.
- `internal/auth/session.go:76-86`: `IssuedAt` "is when the session was last issued ... A refresh moves it";
  `AnchoredAt` "is the session's original issue time ... every check that must reason about when the session was
  born reads the anchor rather than IssuedAt, which a refresh advances."
- `go doc ./internal/engine/store Handle.Update`: "Update runs fn inside one IMMEDIATE transaction ... ctx must carry
  a deadline, which bounds the wait for the write pool's one connection inside this process. The wait on another
  process's write lock is bounded by the busy timeout alone." AC16's concurrency proof and V5's `busy` rest on it.
- `internal/engine/store/conn.go:34-36`: "ErrBusy wraps a failure after which SQLite's write lock stayed held past the
  busy timeout." `additionOutcome` already maps it and `context.DeadlineExceeded` to busy (`admin.go:304-311`).
- `internal/settings/accessor.go:20-21`: "A setting read before a write enters [store.Handle.Update] is the value in
  force for that operation; read it outside the callback."
- `internal/auth/store.go:49-58`: `Open` "creates the leaf and both record subdirectories under stateDir if they do
  not already exist" and "refuses to start when the leaf itself already has a mode wider than dirPerm".
- `cmd/dubplate-server/auth.go:276-291`: `durationAtLeast` refuses only when `l < s`, so it admits equality.
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
| the actor's anchor | the verifier, from `Session.AnchoredAt`; never `IssuedAt`, a request field, or a refresh |
| the notice's recipients | the roster's live admins, read after the commit; never a request field or the environment |
| `DUBPLATE_MEMBER_CAP` and `DUBPLATE_REAUTH_WINDOW` | as every sensitive store-homed setting: `set_setting` inside the window, `settings reset`, the seed rule |

| Call | State | Answer |
|---|---|---|
| any path under `<prefix>/api` | `api`, `api/`, `api/v1`, `api/v2/...` | 404, JSON `not_found`, before verification |
| any `/api/v1` call | verification fails on a setting read or a roster lookup | 500 `internal`, logged at Error |
| any `/api/v1` call | a valid session, same-origin | admitted, and the refreshed cookie set |
| any write | the store's write lock outwaits the busy timeout or the deadline | `unavailable`, detail `busy`, nothing written |
| any member call | `user_id` empty, uppercase, of the wrong length, not hex, or all zeros | `invalid_request` |
| add (`add_member`, `member add`) | live members at the cap, a new or re-added address | `conflict`, detail `member_cap`, naming the setting and removal |
| add | live members at cap minus one | commits |
| add | `member add --admin` on a roster with no live member, holding cap-many tombstoned rows, beside a cap row that cannot be read | commits; the bootstrap ignores the held cap read (PS bootstrap) |
| add | the cap cannot be read, roster not empty | `internal`, nothing written |
| add over `add_member` | a removed admin's address | reactivated without the flag |
| `member add`, `remove`, `demote` | `auth.Open` refuses (a leaf wider than `0700`, a subdirectory it cannot create) | refused naming the verb and the leaf, nothing written; `member list` opens no auth leaf |
| `serve` start | the window at or above `DUBPLATE_SESSION_ABSOLUTE_CAP` (the `2h` default against a cap at or below `2h`, with `DUBPLATE_SESSION_LIFETIME` lowered to at most the cap so `sessionLifetimeWithinCap` passes) | the seed rule refuses, naming `DUBPLATE_REAUTH_WINDOW` and the cap; with a cap at or below `5m` the message names the window's floor |
| the notice | the send fails or outwaits its deadline | logged; the write stays committed |
| `doctor`'s members row | live members above the cap | warns, without failing (PS A1) |
| `doctor`'s members row | the cap cannot be read | the count, with "cap not read", without failing |
| `doctor`'s `DUBPLATE_ADMIN_EMAIL` row | the environment refused, or the store absent, unvouched, older than `UsersVersion`, or not inspected | the row's status as today, no standing warning |

## Review focus

The inputs most likely to bite a person, each pinned in its owning task:

1. A stale session mid-edit: the change survives, and a retry after a sign-in in another tab commits (Tasks 7, 10b).
2. The admin demotes or removes themself while another admin remains: the next call answers `forbidden` or
   `unauthenticated` and the page shows that copy, never a crash or a stale list (Tasks 4, 7, 10a).
3. A key the serving environment sets: read-only on the page, and a write against it refused before the next start
   could be (Tasks 5, 10b).
4. A CLI `member remove` while `serve` runs and the page adds the same address: one write wins, the other answers
   `busy`, `not_found`, or a duplicate, never a half-written row (Tasks 4, 8; the close's smoke).
5. A re-added member's old session in another browser stays refused, and a fresh sign-in at least one second later
   verifies (Task 4).

## Lane launch (the conductor)

Before Task 0:
1. **Live executors.** The lane registry shows no code lane; `git status` in the main tree shows only the
   `dubplate-implementer` agent-memory edits and two 2026-09-04 research files, none code. Confirm no session owns
   them before Task 0 writes master.
2. **Guards:** the runaway guard, the sleep inhibitor sized to the run's end, the stall guard, the `/loop` wake-up,
   and AC power.

After Task 0, before Task 1:
3. **Contested surfaces written.** Copy the spec's contested-surfaces rows into STATUS's lane registry (the fold's
   conductor-owed item), naming lane 1 as 14a's owner, with spec deviation 2's amendment: 14a edits in place the
   harness's members rows and helpers, deck-coverage's roster outcome source, detail-constant pattern, and
   `REQUIRED_TITLES` list, and the `RequireAdmin` table's host.
4. **Lane registry row:** lane 1, the worktree, the branch, this plan, the sub-ceiling (the ceiling).
5. **Provisioning.** `npm ci --prefix web` in the worktree; `make unidecode-wheel` and `make subsonic-xsd` verified
   in the user cache; `xmllint` at libxml 2.15.4 on `PATH`. No venv.
6. **Throughput baseline:** Task 0's timed full gate on master, quoted. The worktree is cut from that commit, and the
   first per-task gate proves its provisioning.
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
- The record lands at `docs/research/2026-10-1x-dependency-sweep.md`.

**Acceptance.** The sweep commit passes `check.sh` (full, heavy) on master, timed for the baseline; the record lists
each version with its evidence and each held item with its unblock condition.

## Segment A

### Task 1: The network actor on `UserID`

**Files:** `internal/ops/ops.go`, `ops_test.go`; `internal/identity/identity.go` (the doc comment at `:209-211`); the
`opshttp` tests that build a network actor (`helpers_test.go`, `opshttp_test.go`, `contract_test.go`).
**Hook:** lands hook 5 in part. **Gate:** the per-task gate.

**Interfaces.** Produces `ops.NetworkActor(id identity.Identity, anchor time.Time)` and the read-only accessors for
the `UserID`, the principal, and the anchor, which Tasks 6 and 7 consume. `LocalActor` carries none of them.

**Outcomes.** The wire form is `member:<lowercase hex UserID>` (spec, "The network actor (D5)"), and the two comments
the spec names (`identity.go:209-211`, `ops.go:42-46`) say so.

**Acceptance.** AC6's actor clause at the `ops` level: two identities with one `UserID` and different principals
render one wire form. **Fixture:** principals that differ, so a principal-built form fails. The settings-audit half
of AC6 is Task 6's; its log clause is Task 7's.

**Mutations:** build the wire form from `Principal`; encode the id in uppercase; drop the anchor from the actor
(the accessor test fails).

### Task 2: Hook 18, description and sensitive mark

**Files:** `internal/config/settings.go`, `keys.go`, `thresholds.go`, `registry.go` and their tests; every
declaration site: `internal/config/keys.go`'s wall rows and `cmd/dubplate-server/auth.go`, `backup.go`, `doctor.go`,
`proxy.go`, `serve.go`, `store.go`, `web.go`; a test in `cmd/dubplate-server` over the production registry.
**Class:** `tool`, `gateLane: "heavy"`. **Hook:** lands hook 18. **Gate:** the per-task gate.

**Interfaces.** Produces the closed mark type (no valid zero value) and a description on `Setting`, `Threshold`,
`WallRow`, and each integration-class key. Tasks 6, 7, and 10b read both through the registry.

**Outcomes.** As the spec, "Hook 18", says. `IntegrationClass` names its keys as declarations carrying a description
and a mark, not bare strings (`keys.go:155-160`). `Setting.Doc` keeps its meaning. Every existing key gets its
description and mark in this task, as the contested-surfaces row requires of the first segment. Every store-homed
setting keeps its pre-14a sensitivity: `DUBPLATE_MAGIC_LINK_EXPIRY`, `DUBPLATE_SESSION_LIFETIME`, and
`DUBPLATE_SESSION_ABSOLUTE_CAP` are sensitive (`cmd/dubplate-server/auth.go:126-155`), and no other is. If rung 7
has merged keys first, this task marks them too.

**Acceptance.** AC8. A table test over the production registry pins the exact set of store-homed keys marked
sensitive; Tasks 4 and 6 add `DUBPLATE_MEMBER_CAP` and `DUBPLATE_REAUTH_WINDOW` to it. **Fixture:** per family (setting,
threshold, wall row, integration-class key), one declaration missing a description and one with a zero mark, each
panicking with the key's name; the existing construction test builds the binary's whole registry.

**Mutations:** accept a zero mark on one family; accept an empty description on one family (each family's proof
fails alone); flip one store-homed key's mark (the pin fails).

### Task 3: The ops layer: the admin seam, the codes, and the unavailable job seam

**Files:** `internal/ops/` (a new file for the admin seam and its request and response types, a new file for the
unavailable job seam, `catalog.go`, `errors.go`, and their tests); `internal/ops/opshttp/opshttp.go`, `authorize.go`,
and their tests. **Hook:** lands hook 15 in part (the settings operations' vocabulary). **Gate:** the per-task gate.

**Interfaces.** Produces `ops.AdminSeam` with the seven operations (spec, "The mount", the operation list); its
request and response types; the two codes' sentinels; an error carrying an optional `detail`, with P8's values as
exported constants of the one shape `Detail<Name> = "<value>"`, declared together in one `internal/ops` file (Task
10a's deck gate parses that shape; `internal/ops/admin` reads them, so they are exported); and the unavailable job seam's type. Task 6 implements the admin
seam; Task 7 mounts the handler.

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
needed; `internal/auth/admin.go` (the roster API's `AddMember` call, one outcome constant, and the outcome map only,
`:35-45`, `:265`, and `:305`, nothing in `RequireAdmin`); `web/src/lib/copy/deck.ts` (one `ROSTER_OUTCOMES` line); every other `NewRoster` caller (`admin_test.go`, `seam_test.go`,
`cmd/dubplate-server/serve_test.go`, `rosterreset_test.go`); `cmd/dubplate-server/serve.go` and `member.go`, for the
roster's write dependencies (P11); a new `cmd/dubplate-server/api.go` holding `DUBPLATE_MEMBER_CAP`'s declaration
(P7) and `main.go`'s one registration line. **Hook:** lands hook 5 in part (the in-write actor re-read by `UserID`);
the epoch stays on the address (D6). **Gate:** the per-task gate.

**Interfaces.** Produces the roster's remove and demote methods naming the target by `UserID`, the exported in-write
admin check (P10), and the cap's read on the add path. Consumes the cap as an `auth.Setting[int]` (P11).

**Outcomes.**
- Removal and demotion as the spec's "Member removal and demotion", "The epoch bump and R-6d20 (D6)", and "The
  last-admin guard (D7)" sections say, inside one `Update` each, the bump after the tombstone and the guard.
- The reactivation branch of `addUser` bumps the address's epoch in the same callback and order (fold C3). A
  cookie-channel reactivation of a removed admin's address returns without the flag.
- **V6:** `AddMember`'s cookie-channel actor is the acting `UserID`, re-read by id inside the write as removal does.
  The command line keeps the operating-system username for its log record. `internal/auth` imports no
  `internal/ops`. The roster API's handler passes the new actor until Task 10c removes it.
- **V7:** the code comment at the `ErrEpochBackward` branch states the reason: the stored epoch is later than now, so
  committing loses no revocation a write could make. The accepted residual is a session anchored between that epoch
  and the clock step, which no bump can reach. Task 11's HISTORY entry repeats it.
- `Subscribe`'s doc comment gains the must-not-block contract (spec, last paragraph of "The epoch bump").
- P11's shape: `member add` opens the auth leaf through `auth.Open` for the re-add bump, and reports its refusal
  naming the verb and the leaf; `member list` opens none.
- **The cap** (P7, P11): the add path counts live members, admins included, inside the add transaction, and refuses at
  the cap with `member_cap` naming the setting and removal. A tombstoned row does not count. A reactivation counts.
  The cap is read and held before `Update` (P11); inside it, a roster with no live member takes the bootstrap and
  ignores the held result. A cap that cannot be read on a non-empty roster refuses with `internal`, never a default.
- The roster API's interim answer for the cap, until Task 10c retires it: a new `outcomeMemberCap = "member_cap"`
  in `admin.go`, mapped to 409, with one `ROSTER_OUTCOMES` line in `deck.ts`, because deck-coverage requires a deck
  line for every outcome it parses from `admin.go` (`deck-coverage.ts:179-183`) and runs in every `make web`.
- `DUBPLATE_MEMBER_CAP` is registered in the production registry and joins Task 2's sensitive pin.

**Acceptance.** AC15 and AC16's roster halves. PS A1's roster rows: at cap minus one the add lands, at the cap it is
refused, and the bootstrap lands. The states table's remove, demote, and add rows at the roster level.
**Fixtures:**
- AC15's: a preference, an override, and a quota row written first; the removal's epoch record deleted before the
  re-add; the clock injected one second past the re-add's bump.
- AC15's bump failures: an epoch store whose bump fails with a non-backward error, after which the member is still
  live; and a clock behind the stored epoch, so `ErrEpochBackward` returns and the removal commits.
- AC16's: the second connection holds the write lock with one demotion uncommitted. The other demotion starts, and
  the test commits the first after a fixed hold well inside the busy timeout. The busy timeout is set far above the
  hold, so the correct code never answers `busy`. The implementer reports the guard-outside mutant failing in ten of
  ten runs. No store hook is added (the fold refused a barrier hook).
- A roster at the cap holding one tombstoned row, so a count that includes tombstones fails.
- The bootstrap: an empty roster holding cap-many tombstoned rows, beside a cap reader that errors, takes
  `member add --admin`.
- An unreadable cap on a non-empty roster: the reader errors, and nothing is written.
- An add over the roster API at the cap, answering 409 `member_cap`.
- Two concurrent adds at cap minus one, under AC16's lock fixture: exactly one lands.
- A tombstoned admin row re-added on the cookie channel returns as a member.
- A sole admin's refused self-removal leaves that admin's session verifying.

**Mutations:** move the guard outside the transaction (the concurrency proof fails); bump before the guard (the
refused self-removal fixture fails); bump after the commit; commit despite a non-backward bump error; skip the re-add
bump; roll back on `ErrEpochBackward`; delete overrides or quota on removal; read the actor's flag by address or
outside the `Update`; count tombstoned rows toward the cap; count live members outside the `Update`; refuse on the
held cap error before the empty-roster test inside `Update` (the bootstrap fixture fails); keep the old row's admin flag on a cookie-channel reactivation.

### Task 5: The settings write path

**Files:** `internal/settings/` (a new write file and its test, and `evaluate.go` for BACKLOG #25 at `:265`).
**Hook:** neutral (hook 9: the write rides the one-writer seam). **Gate:** the per-task gate.

**Interfaces.** Produces the exported set and reset write functions. Each takes the registry, the serving
environment, the actor's wire form and channel, and the in-write guard (P10, absent on the command line). Each
returns what it wrote or why not, the source in force, the apply mode, and whether the key is sensitive. Neither takes
a mailer: the notice lives only in `internal/ops/admin`, which the CLI verbs and the seed rule do not import, so
AC13's command-line and seed-rule clauses hold by construction. Task 6 and Task 8 are the callers.

**Outcomes.** As the spec's "The settings write path" through "Rate limits stay restart-applied" say: one validator
(the written key's own check, then the seed rule's per-key comparison over `Outcome.Err`, never `Evaluation.Err`), the
`admin` source, `seed_value`, the generation bump, one `settings_audit` row per committed set or reset (spec:244: the
reset "audits the same way", with an empty new value), written inside the same `Update`, the environment refusals,
and the no-op answers. BACKLOG #25: the parse failure is wrapped, not the nil `Validate` result.

**Acceptance.** AC10 and AC11 at the function level, the reset's audit row (kind, actor, channel, key, the old value,
an empty new value), and the states table's `set_setting`, `reset_setting`, and per-key comparison rows. **Fixtures:**
- an `env-seed` row, so `seed_value` records the seed;
- an `admin` row, so it keeps the earlier `seed_value`;
- a store already holding one out-of-bounds `admin` row beside a valid write to another key (AC11's last clause);
- two keys under one correlated rule;
- written test-first and failing at HEAD, an `admin` row whose value does not parse, whose refusal wraps the parse
  error (`errors.Is`; BACKLOG #25).

**Mutations:** compare `Evaluation.Err` instead of per-key outcomes (the existing-refusal fixture fails); skip the
written key's own bounds check; write over an environment-set key; bump the generation on a no-op; overwrite
`seed_value` on an `admin` row; skip the audit row on reset; write the audit row after the commit.

### Task 6: The admin seam's implementation, the gate, and the notice

**Files:** a new `internal/ops/admin/` package and its tests; `cmd/dubplate-server/api.go` (P6's declaration and its
correlated rule). **Model:** opus. **Hook:** lands hook 15 in part (settings and members on the cookie mount's seam).
**Gate:** the per-task gate.

**Interfaces.** Consumes Task 1's actor accessors, Task 3's admin seam and codes, Task 4's roster methods, and Task
5's write functions. Produces the admin seam for `serve`, which Task 7 mounts.

**Outcomes.**
- The seven operations, each over the roster or the write functions. `list_settings` answers as the spec's "Hook 18"
  last paragraph says, with no wall or integration value ever on the wire.
- The re-authentication gate decided here, after decoding, from the actor's anchor accessor and the key's mark, for
  exactly the calls the spec's "The re-authentication gate" names. An unreadable window refuses with `internal`.
- The notice as the spec's "The sensitive-change notice" says: after `Update` returns, under its own deadline, to
  every live admin, logged on failure. It names the key, the actor's principal, and the old and new values.
- `user_id` parsing per P12. **V5:** `store.ErrBusy` and `context.DeadlineExceeded` from any write answer
  `unavailable` with detail `busy`.
- `DUBPLATE_REAUTH_WINDOW` is registered and joins Task 2's sensitive pin; its correlated rule is strict (P6).

**Acceptance.** AC12 at the seam (the refresh clause is Task 7's), AC13, AC6's settings-audit half, AC9's content
half, AC16's settings-write actor clause, AC17's seam half (the add request type has no admin field), and every
states-table row reachable at the seam. **Fixtures:**
- AC12's actors anchored inside and outside the window;
- a window reader that errors;
- a window equal to the absolute cap (refused) and one a second below it (accepted), and the `serve`-start seed
  refusal with a cap at or below the window. Each lowers `DUBPLATE_SESSION_LIFETIME` to at most the cap, since any
  cap of `24h` or less already fails `sessionLifetimeWithinCap` against the lifetime's `2160h` default
  (`auth.go:260-263`); each refusal's assertion names `DUBPLATE_REAUTH_WINDOW`;
- AC13's two live admins and one removed admin behind a recording mailer;
- a mailer that fails, to prove the write stays;
- a mailer that blocks until released, beside a concurrent writer whose deadline would expire if the send held the
  write;
- AC6's member renamed by a direct store write between two writes;
- a held write lock under a short deadline, against one settings write and one member write, each answering `busy`;
- `user_id` empty, uppercase, 30 characters, non-hex, and all zeros.

**Mutations:** gate on a time other than the actor's anchor accessor; skip the gate for `add_member`; send the notice
inside the write callback (the blocking-mailer fixture fails); mail a removed admin; fall back to a default window on
a read failure; map `ErrBusy` to `internal`; make the window rule non-strict (the equal-to-cap fixture fails); accept
an uppercase `user_id`; drop the length check.

### Task 7: The mount and the contract document

**Files:** `internal/ops/opshttp/` (the exported mount, `openapi.yaml`, the contract tests); a new
`internal/auth/` file for P2's session check and its test; `cmd/dubplate-server/api.go` (registration of `api`
through `Shell.Register`, and the verifier); `cmd/dubplate-server/serve.go` (one call into `api.go`); the route
inventory test if no lane has landed it (PS A2). **Hook:** lands hook 15 (the cookie mount). **Gate:** the per-task
gate.

**Interfaces.** Consumes everything above. Produces the `/api/v1` surface the pages call in Tasks 10a and 10b.

**Outcomes.**
- The mount, the carrier, the verifier, and the 401/500 split per P2 and P3; the actor's anchor from
  `Session.AnchoredAt`. One log record per admitted operation with actor, principal (its own field, AC6), channel,
  and `client_ip`. No bearer header is ever read.
- The document per the spec's "The contract document (D4)", with every operation, body, code, and detail, the version
  policy, and no `security` requirement.
- The contract tests per AC1 and AC2: operation-path parity, the keyword and `format` walk, `IncludeResponseStatus:
  true` with every other silent pass failing (on the `conformance.go:96-136` precedent), the closed loaded copy, and
  the three named proofs.
- **V4:** one test marshals the job response values (every state, a failure, every timestamp) and validates each
  against the document's schema directly, since no routed answer carries them on this mount. It also asserts that
  each job schema's property set equals the Go type's JSON field set, so a queue-position property cannot enter the
  document (AC7).
- **The route inventory (PS A2):** if no lane has landed it, this task lands it over the real serve shell, every route
  with its auth class (public, session, admin), and no delegated row (PS erratum, 2026-10-05). If rung 7 landed it
  first, this task adds `api` and keeps `roster` until Task 10c removes it.

**Acceptance.** AC1 to AC5; AC6's log clause (a test reads the record's principal field); AC7's document half; AC12's
refresh clause; AC17's HTTP half (a body naming `admin` adds the member without the flag); AC21's first clause; AC9's
raw-body sentinel scan over every `/api/v1` answer, refusals included; AC18 for every HTTP-reachable state; the states
rows above for `api` paths and the 500. **Fixtures:**
- sentinel values on every wall and integration row;
- a live member session and a live admin session;
- a session whose cap setting is made unreadable (500);
- no cookie (401), and a removed member's session (401);
- a same-origin request past the refresh point, which sets the refreshed cookie;
- an admin removing themself while another admin remains, whose next call answers 401;
- a session whose anchor is outside the window and whose `IssuedAt` is inside it after one same-origin refresh,
  driven through the real mount against a sensitive `set_setting`: `reauthentication_required`;
- `DUBPLATE_ADMIN_EMAIL` set to a live non-admin's address, whose `list_members` answers `forbidden`.

**Mutations:** build the actor from `Session.IssuedAt` (the refresh fixture fails); answer a gate refusal with 500;
answer a setting-read failure with 401; drop `IncludeResponseStatus`; leave a response object open in the loaded copy
(the rename proof passes, so the test fails); answer an unknown `api` path in plain text; read an `Authorization`
header.

**After Task 7 (segment B's last step):** the first security read (Task 11, step 1) and its fold.

## Segment C

### Task 8: The CLI verbs

**Files:** `cmd/dubplate-server/member.go`, `member_test.go`; a new `cmd/dubplate-server/settings.go` and its test;
`main.go` (registration); `internal/settings/evaluate.go` (the three seed-rule messages name the verb); the file
holding `serve`'s pre-open older-store refusal (`cmd/dubplate-server/store.go` or `serve.go`). **Hook:** neutral.
**Gate:** the per-task gate.

**Outcomes.**
- `member remove <address>` and `member demote <address>` over the roster methods, resolving the address to the id
  first; a target removed between the read and the write answers not found. Both open the auth leaf as Task 4's
  `member add` does, as the runbook's `systemd-run` form already grants (`docs/server-runbook.md:190`).
- `settings reset <key>` per the spec's "The resolving verb", through `store.OpenWrite`, writing its audit row with
  the command-line actor and channel.
- BACKLOG #12: every required flag's usage says "(required)", on the four verbs.
- The seed rule's three `admin`-row messages name `settings reset`, and `serve`'s pre-open refusal of an older store
  over an `admin` row names the previous-release recovery.

**Acceptance.** AC14; AC15's and AC16's command-line halves; the states table's `settings reset` rows; the CLI
reset's audit row; each listed verb's usage text carries "(required)" on its required flags. **Fixtures:**
- a current store holding each of the three refusing `admin` rows, then `serve` starting;
- two out-of-bounds rows reset one at a time;
- an older store, a newer store, and a restore intent;
- an admin `DUBPLATE_MAGIC_LINK_EXPIRY` row of `5m`, then an admin `DUBPLATE_NONCE_RETENTION` row of `10m`: resetting
  the expiry restores its `15m` default, above the retention, so the reset breaks `nonceRetentionCoversExpiry` and
  is refused (the rule refuses only retention below expiry, `auth.go:255-259`, `:282`).

**Mutations:** open through anything but `OpenWrite`; skip the validator on the CLI reset (the correlated-rule
fixture fails); drop the verb name from one seed-rule message; skip the CLI reset's audit row.

### Task 9: `DUBPLATE_ADMIN_EMAIL`, doctor, and the runbook

**Files:** `cmd/dubplate-server/doctor.go`, `doctor_test.go`; `internal/engine/store/inspect.go` and its test;
`docs/server-runbook.md`. **Class:** `tool`, `gateLane: "heavy"`. **Hook:** neutral. **Gate:** the per-task gate.

**Outcomes.**
- `Inspect` reports the roster standing of one address its caller names (live admin, live non-admin, removed, none),
  and no address enters `store.Facts` (`inspect.go:166-171` carries counts only; spec, last paragraph of
  "`DUBPLATE_ADMIN_EMAIL` (D9)").
- `doctor`'s `DUBPLATE_ADMIN_EMAIL` row per AC21 and P9: the warning is the row's status text with `failed` false.
  The `OwnedWall` loop adds the row before `storeRow` runs `Inspect` (`doctor.go:64-66`, `:99`); rows print at the
  end (`:160-164`), so the row's status is resolved from the store row's report with the print order unchanged. The
  row's states follow the states table above.
- The members row warns when live members exceed `DUBPLATE_MEMBER_CAP` (PS A1, "an over-cap roster loads with a
  `doctor` row"), and appends "cap not read" when the cap cannot be read; neither fails.
- The runbook per the spec's three edits (step 10, the key description at `:59-60`, the restore line), plus one line
  naming `member remove`, `member demote`, and `settings reset` as verbs run through the same `systemd-run` form.

**Acceptance.** AC21's `doctor` and runbook clauses. **Fixtures:** a store per standing (a live admin, a live
non-admin, a tombstoned row, no row); an empty roster and an absent store, which add no standing warning; an unvouched
store; a roster one above a lowered cap; an unreadable cap row.

**Mutations:** fail the exit status on the warning; let an empty roster add the address warning; merge the live
non-admin message into the removed one.

### Task 10a: The members page onto `/api/v1`

**Files:** `web/src/routes/members/+page.svelte`; a new `web/src/lib/api/` client module for `/api/v1` (14a's);
`web/src/lib/copy/deck.ts` (detail-keyed rows appended); `web/scripts/gates/deck-coverage.ts` (the detail
enumeration's coverage); the members rows and helpers in `web/scripts/rendered/run.ts` (`:866`, `:1050-1090`,
`:3086-3151`) and `cmd/dubplate-server/rendered_test.go` (`:218-219`, `:860`), retargeted (spec deviation 2).
**Class:** `engine-logic`; the scoped gate runs the rendered leg; `reducedGate` is the per-task gate (`--reduced` never runs the rendered leg). **Hook:**
neutral (consumes hook 15). **Gate:** the per-task gate.

**Outcomes.**
- The members page lists, adds, removes, and demotes over `/api/v1` through the new client, confirms before a removal
  or demotion, and shows each refusal's own copy, the member cap's included. Each add refusal moves focus to its
  field. The client reads `error.code` on a JSON 403; only the origin check's plain-text 403 is cross-site.
- deck-coverage requires a deck line for every P8 detail value, parsed from Task 3's file by a second pattern for
  the `Detail<Name>` shape beside the existing one, which matches only unexported `outcomeX` names
  (`deck-coverage.ts:29`). The roster API stays live, and its outcome source stays in the gate, until Task 10c.
- The existing members rows prove the same behaviors over `/api/v1`: the unreachable row, the malformed-answer rows,
  the non-admin row (which now meets `list_members`'s JSON `forbidden`), the console-noise allowlist, and the proof
  that the script's write reached the real API. The roster reset server writes the store directly and stays.

**Acceptance.** AC19. **Rendered rows** (each names what it intercepts and what it reaches for real):
- a JSON 403 renders the code's copy, never the cross-site copy;
- each refusal's copy: last admin, re-authentication needed, forbidden, signed out, and member cap;
- each add refusal moves focus to its field;
- the confirm step blocks the call until confirmed;
- last admin, against the fixture's one admin (`seedRenderedMembers`, `rosterreset_test.go:38-43`): that admin
  demotes self through the real API, which answers `last_admin` and writes nothing, so no reset is owed. The roster
  reset server only deletes rows outside its three-address baseline (`rosterreset_test.go:99-104`); it restores no
  admin flag, tombstone, or epoch, so no row demotes or removes a baseline address;
- a removal through the real API, after the confirm step, of an address outside the reset baseline that the same row
  added first, so the reset deletes it and no baseline address takes an epoch bump. A successful demotion needs a
  second admin, which the web cannot create, so it is proved in Go (Task 4) and in the close's smoke.

### Task 10b: The server settings page

**Files:** a new `web/src/routes/server-settings/+page.svelte`; the `web/src/lib/api/` client module (the settings
operations); `web/src/lib/copy/deck.ts` (rows and `TITLES`); `web/scripts/gates/deck-coverage.ts` (its
`REQUIRED_TITLES` list only, `:43`, since the gate fails a `TITLES` key the list does not name, `:69-72`); the route
inventory Task 7 lands (`server-settings`, admin class); `SiteHeader.svelte` (one appended navigation entry); `cmd/dubplate-server/web.go` (the
document row, appended); `serve.go` (the page's registration behind `RequireAdmin`); appended rendered rows.
**Class:** `engine-logic`; the scoped gate runs the rendered leg; `reducedGate` is the per-task gate (`--reduced` never runs the rendered leg). **Hook:**
neutral (consumes hook 15). **Gate:** the per-task gate.

**Outcomes.** The server settings page per AC20 and P4, with P5's stale-session flow. It follows the visual identity
spec and passes the `web/` leg's static and rendered gates.

**Acceptance.** AC20. **Rendered rows:**
- every key with its description and mark;
- an environment-set key, shown as set in the environment file with no edit or reset;
- a write's apply-mode line;
- the stale-session flow: the first sensitive write is intercepted and fulfilled with a JSON 403
  `reauthentication_required`; the field still holds the edit; the sign-in link opens a new tab; and Retry, reaching
  the real API, commits. The write is `DUBPLATE_MEMBER_CAP` at its default, `20` (the fixture sets
  `DUBPLATE_MAGIC_LINK_EXPIRY` in the environment, `rendered_test.go:386`, so that key is not writable), and the
  same row then resets it through `reset_setting`, because the harness has no settings reset and a later capture
  cell must not meet an `admin` row. The notice's send fails against the fixture's closed SMTP port and is logged.
  The real gate's refusal is proved in Go (Task 7) and in the close's smoke.

### Task 10c: The roster API's retirement

**Files:** `internal/auth/admin.go` (the roster API's handlers and its outcome constants removed; no `RequireAdmin`
line changes); `internal/auth/admin_test.go` (the `RequireAdmin` refusal table rehosted on a stub handler);
`cmd/dubplate-server/serve.go` (the `roster` registration at `:359` removed); `cmd/dubplate-server/serve_test.go` (the
`/roster` tests at `:810-1208` turned into 404 proofs, or moved to `/api/v1` where they prove the gate); the route
inventory (`roster` dropped); `web/scripts/gates/deck-coverage.ts` (the roster outcome source, its must-have list, and
its floor count re-derived) and `deck.ts` (the roster rows no page reads, removed). **Class:** `auth-data`. **Hook:**
neutral. **Gate:** the per-task gate.

**Outcomes.** `roster` answers 404, and BACKLOG #23 closes with it (spec, "Riders"). `RequireAdmin` keeps every
refusal row, each still proved.

**Acceptance.** AC17's `roster` 404, AC22's #23. **Fixtures:** the rehosted table's every refusal row; `roster`,
`roster/`, and `roster/x` under an admin session.

**Mutations:** for each rehosted refusal row, admit its case (the row fails); leave the `roster` registration (the 404
proof fails).

## Close

### Task 11: Reviews, verification, merge, and the ledgers

The conductor's steps, in order (pass-core, "Closing"; CLAUDE.md, "Lanes" and "Reviewers"). Every code-changing fold
lands before the close gate.

1. **The first security read, at boundary B** (segment B's last step): `web-auth-security-reviewer` over Task 3's
   authorizer and privileges rows (`opshttp/authorize.go`, `catalog.go`), Task 4's roster and epoch code, the write
   path, the gate, the notice, the mount, and the session check. It answers: does a removed or re-added member's old
   session stay refused; is the guard inside the transaction; is the bump after the guard and before the commit; and
   does the cap hold on both channels. Then one fold task.
2. **The close's reads, then one fold.** In parallel where independent:
   - the second security read over Tasks 8, 10a, 10b, and 10c (the CLI verbs, the pages, the retirement, and the
     rehosted `RequireAdmin` table);
   - `svelte-reviewer` and `daisyui-a11y-reviewer` over both pages, each passed the visual identity spec's path and
     asked to grade conformance (tokens, inks, the component layer, the copy register) beside correctness;
   - one `go-architecture-reader` per touched package, no plan in context: `internal/ops`, `opshttp`,
     `internal/ops/admin`, `internal/auth`, `internal/settings`, `internal/config`, `internal/engine/store`, and
     `cmd/dubplate-server`. `internal/identity` takes none: Task 1 changes one doc comment there, and the reader
     grades architecture;
   - the **live auth smoke** on the workstation, on the real binary under `env -i`, with the loopback SMTP sink (6d's
     method, HISTORY 6d, "The gate evidence"). Start the `5m` window wait first and run the other steps inside it.
     Steps: sign-in; `catalog` with the three job operations unavailable; `list_settings`; a non-sensitive write; a
     sensitive write from a fresh session, with the notice at the sink; the window seeded at its `5m` floor and
     outwaited, then a sensitive write refused; `add_member`, `remove_member` (the removed member's next request
     refused), a re-add (the old session still refused), `demote_member`, and the last-admin refusal; `member remove`
     from the CLI while `serve` runs (the member's next request refused); the cap seeded at the live count, so an add
     is refused on both channels; `roster` and `api/v2/x` answering 404; and `settings reset` with `serve` stopped
     clearing an out-of-bounds `admin` row written directly;
   - **render and read** both pages in both palettes at phone and desktop widths, in each refusal state, graded
     fresh-context.

   Then one combined fold task through the chain under `auth-data`; it may refuse a finding with a one-line reason.
   The failing smoke steps and pages re-run after it; a second red stops the run before the merge ritual.
3. **Simplify once:** `code-simplifier:code-simplifier` over the pass's Go, TypeScript, and Svelte, with
   `.claude/instructions/simplifier-brief.md` pasted verbatim, then its `diff-reviewer`.
4. **Merge token,** then rebase onto master.
5. **Close gate:** `check.sh --merge` (the full gate plus the `master..HEAD` gitleaks scan, heavy) in the worktree. It
   is pass-core's close gate and the merge ritual's in-tree gate at once. On red, revert the simplifier commit and
   re-gate first.
6. **If the rebase was not a fast-forward:** one `diff-reviewer` on the merged result; and if master's new commits
   touch `internal/auth`, `internal/serve`, or `cmd/dubplate-server/serve.go`, the smoke's auth steps re-run on the
   rebased head.
7. **`--ff-only`** onto master.
8. **`check.sh` on master.**
9. **Docs fold** (one `sonnet` fold agent, one independent `diff-reviewer` over its diff):
   - Errata 1 to 8 (the fold, "Owed errata" and "continued"), each a dated erratum note in its ratified document:
     R6:913-915, R6:589-590, R6:943-944, R6 criterion 7 with erratum row 10, CH:171-172, PA:173-215, BS:983, and
     R6:582. PS gains a note that the member cap landed at 14a, and an erratum to PS:145-147: 14a has no
     command-line set verb, and the seed rule, the command-line route to the cap, sends no notice (the fold's C9 and
     erratum 5).
   - The spec's owed erratum from the plan fold: the contested-surfaces rows for `web/` and `internal/auth/admin.go`
     gain spec deviation 2's in-place edits.
   - BACKLOG: #12, #23, and #25 closed; C14's rung 16 check filed (the fold's conductor-owed list).
   - STATUS, present tense: the routing of `serve`'s job runner, instance lease, and job recovery to the first rung
     whose page starts a job (erratum 3); 14b's trigger; the contested-surfaces rows released; lane 1 idle; and, as
     an open decision, whether rung 7 and a running rung 12 slice may both write `internal/serve` (the fold's
     conductor-owed list).
   - HISTORY: what landed per task; what the gates caught; both security reads' dispositions; the fold-rule trial
     (any refused spec or plan finding that proved a real defect, or none); V7's residual; the sandboxed
     `member remove`, `member demote`, and `settings reset` routed to rung 16 beside `member add`; "what a later pass
     would be wrong to rediscover" (the 401/500 split, the anchor never `IssuedAt`, the per-key comparison, the
     epoch ordering and the re-add bump, why the job operations answer `unavailable`, the uppercase-hex and empty-id
     traps, the bootstrap ignoring the held cap read, and P5's no-return-parameter reason); and the pass score: tokens
     against 11.8M, attended time, and clock against about 6.8 hours (total, gate time, lock wait, and rework with
     each red's cause). The gate time and lock wait come from `cairn-run-gate --records <toplevel> <branch>` (present in
     the script's source, checked 2026-10-10 by grep).
   - **The review-growth measure:** the spec grew from 419 to 651 lines across two folds (55%), past the 25% signal,
     and the plan from 609 to its revision 2 length (the plan fold's measures). The entry names the cause (the
     states table, the contested-surfaces table, the quoted kin-openapi behaviors, V1's admin seam, and the plan
     fold's fixtures and Task 10 split) and whether the growth bought execution-time savings.
10. **Commit** specific files; pre-bake STATUS's resume prompt; close the session.

## Prior art consulted

The spec's "Prior art" table is this plan's, by reference: kin-openapi adopted in tests only; go-mail through
`internal/notify`; `x/time/rate` kept with live apply ruled out; the stdlib MACs, sqlite, and goose neutral; GitHub
sudo mode adopted for its rule and P6's default, declining its timer reset; and the OpenSubsonic specification,
Navidrome, gonic, taglib, mewkiz/flac, Picard's tables, and chromaprint neutral. The member cap adopts no library.

## Rulings for Geoff

None. The member cap's placement is a sequencing call taken above with its reason, not a product fork: PS ruled the
mechanism and its default. The cap's command-line notice (the plan fold, D11) is settled by the spec's AC13 and the
fold's C9, and Task 11 records the PS erratum. Spec deviation 2 is a coordination note while rung 7 only plans.

## Errata (filled during execution)
