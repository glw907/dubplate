# Rung 14a plan fold: verification read

Target: `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md`, revision 2 at `a81c6bb` ("the plan"), against
revision 1 at `fae3de3`. Fold record: `research/2026-10-10-rung-14a-plan-fold.md` ("the fold"). Reviews:
`-plan-review-contract.md` (PA), `-mechanics.md` (B), `-domain.md` (D). Spec: revision 2.1. Reader: fresh context,
no part in the fold. Every claim below was checked against the tree at `a81c6bb` (code unchanged since `fae3de3`) or
in a scratch probe on go1.27.2. Findings are correctness gaps ranked by consequence; polish is omitted.

## 1. Did each blocker and major close at the cited location?

Yes, all fourteen. Residuals that the fold itself introduced are listed as findings below.

| ID | Closed at | Evidence |
|---|---|---|
| B1 | Tasks 10a, 10b, 10c (plan:667-731); spec deviation 2 (:219-232); lane launch step 3 (:323-325); Task 11 step 9 erratum (:785-786) | `deck-coverage.ts:158-167` does hard-require `admin_not_grantable` and 21 outcomes; `admin_test.go:75`, `:265` build on `NewRosterAPI`; `serve_test.go:795-1013` drive `/roster`. Residuals: V1, V5, V6, V7. |
| B2 | plan:74-77, :673, :700 | `pass-execute.js:356` honours a task's own `reducedGate`; `engine-logic` has `testOnlyReduces: true` (`:310`); `auth-data` has `false` (`:303`), so 10c never reduces a test-only round. |
| B3 | P11 (:196-202), the accessor quote (:253-254), Task 4 (:447, :463-466) | `accessor.go:20-21` quoted exactly; `auth.Setting[T]` exists (`handlers.go:75`); `durationSetting` and `authBodyCap` sit at `auth.go:499-516`. Residual: V2. |
| B4 | P11; Task 4 files (:438-443) and outcome (:461-462); states row (:292) | Callers at `member.go:97`, `:133`, `serve.go:324` confirmed; `auth.Open` creates an absent leaf (`store.go:49-58`), so the row correctly omits "leaf absent". |
| B5, PA-3 | Task 11 (:737-778) | Every code fold now precedes `check.sh --merge`; one full gate removed. `--merge` is "the full gate plus one gitleaks pass over ... master..HEAD" (`check.sh:50-59`). Residual: V8. |
| PA-1 | Task 7 outcome (:576), fixture (:601-603), mutation (:605); Task 6 mutation restated (:559) | `session.go:76-86`: "IssuedAt ... A refresh moves it"; the anchor is "the session's original issue time". |
| PA-2 | Tasks 10a-10c files; global constraints (:138-140); rendered rows (:686-692, :706-712) | No vitest in `web/package.json`. Residual: V1 (one row cannot reach its state). |
| PA-4, D4 | Task 2 (:393-405) | `Sensitive: true` appears exactly three times in non-test Go (`auth.go:134`, `:144`, `:154`), so the pin's set is right. |
| D1 | Task 11 step 1 (:740-744) | Names Task 3's authorizer, Task 4's roster and epoch code, and the four questions. |
| D2 | plan:100-112 | One rule for a standing `fix`; the smoke, page, close-gate, and on-master reds each have a verdict. |
| D3 | plan:90-98 | Four rules present. Residual: V8. |
| D5 | Task 5 (:507-508, :511-512, :522); Task 8 (:625, :631-632, :640) | Spec:244: `reset_setting` "audits the same way". |

The refusals hold. PA-8's `WaitCount` refusal is right: `go doc database/sql DBStats` says WaitCount is "The total
number of connections waited for", a pool counter, while a second handle's lock is waited on inside SQLite's busy
handler. D8's refusal is right: AC16 (spec:604-608) requires "An actor demoted before a ... settings write is
refused", which Task 6 drives through `internal/ops/admin`, the only cookie-channel caller.

## 2. Contradictions and build order

The task order builds. Task 4 creates `api.go` after Task 2's mark type exists; Tasks 6 and 7 extend it; 10a needs
Tasks 3 and 7; 10b needs Tasks 2 and 6; 10c needs 10a off `roster`. Segment D (10b, 10c) is a legal segment, and the
first security read is pinned to segment B by both the header (:30) and the ceiling rule (:92-93). The close order is
coherent. Five fold-introduced defects remain, as V1, V2, V5, V7, and V8.

## 3. Mechanisms stated from memory

Most new mechanisms were verified against source. Checked and true:
- the runner's per-task `reducedGate` and `model` (`pass-execute.js:356`, `:857`);
- `--reduced` skipping the rendered leg (`check.sh:261-263`) and `svelte-check` (`:375-378`);
- `hex.DecodeString("")` returns an empty non-nil slice and nil (probe: `"" -> [] err=<nil> reenc=""`);
- `auth.Open`'s creation behavior; `durationAtLeast` admitting equality (`auth.go:282`, `if l < s`);
- the rendered harness's `route.fulfill` interception (`run.ts:2705-2706`, `:2949-2951`);
- every quote in "Behaviors this plan relies on".

Five were stated without a check, and four of them are wrong:
- **V1.** The 10a last-admin row assumes a second fixture admin and a reset that restores the fixture. Neither exists.
- **V3.** Task 8's reset fixture cannot break the rule it names.
- **V4.** The `serve`-start fixture is decided by a different correlated rule.
- **V5.** `REQUIRED_TITLES` is placed in `deck.ts`, but it lives in `deck-coverage.ts`.
- **V6.** "The gate's existing constant parser" cannot read Task 3's constants. This one is a shape mismatch, not a
  false claim.

## 4. Owner forks

None is a genuine fork. D11's refusal holds: spec:284 (approved revision 2.1) says "Neither the seed rule nor the
command line sends a notice", so PS:145-147 is amended by a later approved text, not reopened. P5's new-tab link is
the conventional re-authentication flow and design-level, which pass-core gives to Claude. B6's removal of the doctor
warning state and P11's dependency shape are architecture. V10 asks only that the PS erratum be listed as decided in
the close summary.

## Findings

### V1. major. Task 10a's last-admin row cannot reach its state, and the reset it relies on does not restore it.

**Location:** plan:690-692 (Task 10a rendered rows); fold D14.

**Defect.** The rendered fixture seeds one admin and one member (`rosterreset_test.go:38-43`, `seedRenderedMembers`),
so there is no "other admin" to demote. The roster reset server only deletes rows outside the three-address baseline
(`rosterreset_test.go:99-104`). It restores no admin flag, clears no tombstone, and touches no epoch file. A demotion,
or a removal of a baseline row, therefore survives into later capture cells. A removal also bumps that address's
epoch, which the reset never clears. Widening the seed or the reset edits `rosterreset_test.go`, which is outside
10a's files. It also reshapes the rung 7 harness, and its `TestRosterResetServer` pins the counts 2 and 3. The
likely outcome is a fix round or an escalate at segment C's last task.

**Fold.** Restate the row against the existing fixture. The one fixture admin demotes self against the real API,
answers `last_admin`, and writes nothing, so no reset is needed. If a successful removal row is wanted, it targets an
address outside the reset baseline that the run added itself, so the reset deletes it and no baseline address takes
an epoch bump. Successful demotion stays proved in Go (Task 4) and in the close's smoke, because the web cannot
create a non-baseline admin. Change "the roster reset restores the fixture" accordingly.

### V2. minor. P11 orders the bootstrap check before the cap read, and also puts the cap read before `Update`.

**Location:** P11 (:200-202); Task 4 outcome (:465-466); mutation "read the cap before the bootstrap check" (:491).

**Defect.** There is no bootstrap check today. "No live admin" is decided inside `addUser`'s `Update`
(`roster.go:426-432`). The cap must be read before `Update` (`accessor.go:20-21`). Taken literally, "bootstrap check
first, reads no cap" plus "cap read before `Update`" forces an empty-roster read outside the transaction. That read
races a concurrent add, and an add that lands in the gap then skips the cap. The sound form calls the cap reader
before `Update` and holds its result and error. Inside `Update`, a roster with no live member takes the bootstrap and
ignores the held result. That form temporally "reads the cap before the bootstrap check", which is the named mutation,
so the reviewer may flag correct code.

**Fold.** Replace the P11 sentence and the mutation with this. The reader is called before `Update` and its error is
held. Inside `Update`, a roster with no live member takes the bootstrap and ignores the held result; otherwise a held
error refuses with `internal`. The mutation becomes "refuse on the held cap error before the empty-roster test inside
`Update`". The Task 4 fixture (:481-482) already catches it.

### V3. minor. Task 8's correlated-rule reset fixture cannot be refused.

**Location:** Task 8 fixture (:636-637); fold PA-13(b).

**Defect.** `nonceRetentionCoversExpiry` refuses only when retention is below expiry (`auth.go:255-259`, `:282`).
The expiry's ceiling is `1h` (`auth.go:104`) and the retention's default is `24h` (`:159`). Resetting the
retention row restores 24h, which covers any legal expiry. Resetting an admin-raised expiry lowers it. Neither reset
can break the rule, so the correct code fails the test, and the "skip the validator" mutation is never exercised.

**Fold.** Write an admin expiry of `5m`, then an admin retention of `10m`. Resetting the expiry restores `15m`, above
the retention, and is refused.

### V4. minor. The `serve`-start window fixture is decided by `sessionLifetimeWithinCap` first.

**Location:** states row (:293); Task 6 fixture (:549-550).

**Defect.** Any cap at or below `2h` already fails `sessionLifetimeWithinCap` against the lifetime's `2160h` default
(`auth.go:260-263`, `:139`). That holds today, without the window. The fixture as stated refuses for the wrong rule,
so dropping the window rule leaves it green unless the assertion names the window key.

**Fold.** The fixture lowers `DUBPLATE_SESSION_LIFETIME` to at most the cap, and asserts that the refusal names
`DUBPLATE_REAUTH_WINDOW`. Fix the states row's parenthetical the same way.

### V5. minor. Task 10b's files miss two edits it cannot avoid.

**Location:** Task 10b files (:696-699); spec deviation 2 (:219-232).

**Defect.** `REQUIRED_TITLES` is declared in `web/scripts/gates/deck-coverage.ts:43`, not `deck.ts`, and the gate
fails a `TITLES` key it does not list (`:69-72`). Separately, Task 7 lands the route inventory (no inventory test
exists at HEAD), and 10b registers `server-settings`, so the inventory needs a row. 10c lists the inventory, but 10b
does not.

**Fold.** Add `deck-coverage.ts` (`REQUIRED_TITLES` only) and the route inventory (`server-settings`, admin class) to
10b's files. Add `REQUIRED_TITLES` to deviation 2's list of in-place edits.

### V6. minor. Task 10a's "existing constant parser" cannot read Task 3's detail constants.

**Location:** Task 3 interfaces (:413-415); Task 10a outcome (:680-681).

**Defect.** The parser matches `/\boutcome[A-Za-z]+\s*=\s*"([^"]*)"/g` (`deck-coverage.ts:29`), which is only
unexported `outcomeX` names. Task 3's detail constants are read by `internal/ops/admin`, another package, so they
are exported and do not match.

**Fold.** Task 3 names the constants' shape (for example `Detail<Name>`), and Task 10a adds a pattern for it beside
the existing one.

### V7. minor. Task 4's cap refusal on the roster API meets deck coverage, and `deck.ts` is outside Task 4.

**Location:** Task 4 files (:439-440); fold D6 ("adds a `member_cap` outcome").

**Defect.** Every outcome constant parsed from `admin.go` needs a deck line (`deck-coverage.ts:179-183`), and the gate
runs in every `make web` (`Makefile:102`). If Task 4 adds `outcomeMemberCap` to the roster API's map, the `web` leg
goes red on a file Task 4 does not own. Leaving the cap unmapped answers 500 on the interim `/roster` channel.

**Fold.** State the roster API's interim answer for the cap in Task 4. Either add a new outcome, with `deck.ts`'s row
added to Task 4's files, or map the cap to an existing refusal outcome until 10c. Either is fine; the plan should
pick one.

### V8. minor. The ceiling rule's merge-ritual range starts one step late.

**Location:** plan:95 ("Task 11, steps 5 to 8") against :770 (step 4, "Merge token, then rebase").

**Defect.** CLAUDE.md's ritual begins with the merge token and the rebase. A 0.5M check at step 5 can stop the run
holding the token on a rebased branch.

**Fold.** Read "steps 4 to 8".

### V9. minor. Task 10b's stale-session row commits a real sensitive write that no reset undoes.

**Location:** Task 10b rendered rows (:710-712).

**Defect.** Retry commits to the real store, and the harness has no settings reset. The `admin` row persists into
every later capture cell of the settings page. The notice's send fails against the closed SMTP port
(`rendered_test.go:377-378`), which is logged and harmless.

**Fold.** Name the key and value. Either write the key's own default, so later cells render a value they would show
anyway, or follow the commit with `reset_setting` in the same row.

### V10. minor. The PS erratum amends a ratified owner document without a line in the close summary.

**Location:** plan:114-116, :782-784.

**Fold.** The close summary lists the PS:145-147 erratum as decided, with its reason (spec:284), per pass-core's
sitting rule.

## 5. Over-ceremony, ranked by cost

The fold added little that should come out. The ceiling's +1.3M is two tasks at 0.6M plus 0.2M of fixtures, less one
read, and B1 forced the tasks. The clock's +1 hour is the same two tasks and a third boundary gate. The segment rule
(three to four tasks) requires that gate once the 10 split makes thirteen tasks. The +210 plan lines are files,
fixtures, and quotes the majors named.

1. **Task 4's proof load (largest exposure, not a removal).** The task now carries nine fixtures, twelve mutations,
   the P11 refactor across eight `NewRoster` callers, and `member add`'s auth leaf, all at 0.6M on Sonnet. It is the
   likeliest fix-round site. Spec deviation 1 already shows the cap is separable. Name the cap as the pre-approved
   split seam in the extra-dispatch rule (:101-102), at no planned cost.
2. **Lane launch step 8's two BACKLOG filings** cost one verified filing dispatch at boundary A. The gate pass
   supersedes the first, so drop it then.
3. **Close step 6's conditional smoke re-run** fires only on a non-fast-forward rebase that touches the auth paths.
   It is cheap; keep it.
4. **The ten-of-ten mutant runs** on two lock fixtures cost seconds. Keep them.

The pass's largest avoidable cost predates the fold. The cumulative scoped gate makes every task's gate nearly the
full gate with Chromium, about 22 minutes a task (plan:67-71, :48), and the gate pass addresses it.

## Plan text to change when the gate classifier lands

The runner already switches to `scripts/checks/gate-tier.mjs` when it exists in the repo: it computes each task's
tier from `base..HEAD` unless the task pins `gateTier` (`pass-execute.js:56-65`, `:91-93`, `:449`).

1. **:67-71.** The per-task gate string and its cumulative-base account (rendered leg from Task 1 on, conform from
   Task 9) become the computed tier. The fixed string survives only as a pinned task's gate.
2. **:72-73.** The `gateLane: "heavy"` overrides follow the tier. Task 2 touches `serve.go`, `web.go`, and `auth.go`,
   which are rendered inputs (`check.sh:246-252`), so it stays heavy. Task 9 (`doctor.go`, `inspect.go`, the runbook)
   likely goes light.
3. **:74-77, :673, :700.** The 10a and 10b `reducedGate` overrides name "the per-task gate string", which no longer
   exists as a fixed string. Pin each task's `gateTier` to the rendered tier, or set `reducedGate` to that tier's
   string.
4. **:80-83.** The mode arguments gain the classifier decision and any `gateTier` pins. Confirm the classifier maps
   `internal/auth`, `internal/identity`, `web/`, and the serve-shell files to the rendered leg, as `check.sh:246-252`
   does, because `auth-data` owes "the e2e specs its change reaches" (pass-core).
5. **:30-32.** "D's yields to the close gate": with narrow per-task tiers, nothing full-gates the D tip before the
   close's smoke and reads, and the close rebases (step 4) before its gate. Pass-core's skip applies only "with no
   merge or rebase of the default branch in between". Keep D's boundary gate, or state why the close gate covers it.
6. **:43-51.** Re-price the clock ("thirteen tasks at about 22 minutes ... plus the rendered leg every task here
   runs") and the ceiling's gate share.
7. **:333-336.** Lane launch step 8 drops the per-task-base filing, and keeps the double `make web` filing only if
   the gate pass leaves it.
8. **:322-331.** Lane launch notes that the worktree is cut after the gate pass merges, so the runner's probe finds
   the classifier in the worktree.

## Counts

0 blockers, 1 major, 9 minors. No owner fork. Fold-introduced: V1 (D14's reset mechanism), V2 (P11 and D10's
wording), V3 (PA-13b), V4 (D9's row), V5 (B1's `REQUIRED_TITLES` placement and the split), V7 (D6), and V8 (D3).
Pre-existing, surfaced by the split: V6. Fold-adjacent: V9 (D14) and V10 (D11).
