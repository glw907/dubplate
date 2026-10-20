# admin-api-1 plan fold: verification read

Target: `docs/superpowers/plans/2026-10-11-admin-api-1.md` and `2026-10-11-admin-api-1-prompt.txt` at `f546d61` (the
fold commit), against the pre-fold plan at `6b4cd97`, the fold record
(`2026-10-11-admin-api-1-plan-fold.md`), the three reviews, and the approved spec (revision 2.1). Reader: Opus 5.5,
fresh context, no part in the fold. Code read in the worktree at `f546d61` (code unchanged since `551bf4f`).

**Counts:** 0 blockers, 0 majors, 7 minors (3 `correctness`, 4 `optional`). One OWNER FORK, which is the fold's own
clock question, restated with one more option.

**Question 4, in one line:** the arithmetic holds (695 task-minutes, 585 on the path, against a line of 660), no
estimate is padded, and Geoff can rule on the evidence given. The "No" option should also name one cut that keeps the
design whole (T11 folded into T10). Both spec readings are faithful readings of the spec's governing sentence and
change no design.

## 1. Did each blocker and major close where the fold record says?

There were no blockers. All 18 majors close at the cited locations (contract F1 to F6, mechanics M1 to M5, risk RM1
to RM7). One residual sits inside RM2's fold (finding V1 below). It fails closed, so RM2 still counts as closed.

| Major | Closed at (plan line) | Code check |
|---|---|---|
| F1 / RM4 (in-flight demotion) | T3 in-flight fixture :488-491; T5 direct-seam and in-flight fixtures :579-581; mutations :497, :599-600 | Holds. A read before `Update` goes through the read pool and sees the uncommitted demotion as not yet made, so the outside-the-`Update` mutant admits and the fixture kills it |
| F2 (removal bump) | T3 AC15 fixture :475-476; mutation "skip the removal bump" :496 | `go doc ./internal/auth EpochStore` has `Read` |
| F3 / M6 (busy, two arms) | Behaviors :278-285; T5 fixture :592-595; both mutations :600 | The `go doc ./internal/engine/store Handle.Update` text is quoted in full and matches. `write_test.go:573-585` gives `DeadlineExceeded` and `:588-606` gives `ErrBusy`. `go doc ... OpenWrite` reads `func OpenWrite(ctx, stateDir string, busyTimeout time.Duration)` |
| F4 / M4 (detail set) | P7 :226-233; T6 parity test :620-622; mutation :672 | `admin.go:312-337` maps the seven add refusals; `handlers.go:57` is `outcomeInvalid = "invalid_address"`; `+page.svelte:46-54` holds `duplicate_principal`; `go doc ./internal/ops Codes` exists |
| F5 (deck guard) | T9 :750-753; T11 :823-825; mutation :782, :832 | `deck-coverage.ts:29` is `/\boutcome[A-Za-z]+\s*=\s*"([^"]*)"/g`; `:158-169` holds three names and the floor of 21 |
| F6 (unfixtured states) | Each gap is now a fixture line: T3 :478-480, T4 :530-532, T5 :576-582, T6 :654-655, T7 :698-708, T8 :734 | Every row F6 listed now has a named fixture |
| M1 / RM6 (unexported sentinels) | P9 :245-247; T3 :455-456, acceptance :474 | `roster.go:39-105`: every sentinel is lower-case, and the comment at `:55-59` says only in-package code branches on them. `errMemberNotFound` also sits in `gateRefusals` (`admin.go:142-150`), which T3's rename reaches because `admin.go` is in T3's files |
| M2 (route inventory) | T6 :628-634; files :643-645; mutations :673 | `go doc ./internal/serve Shell` lists the methods `EnableHSTS`, `Handler`, `Logger`, `Register`. Every production `Register` call is in `serve.go:344-366` or `web.go:186-194` |
| M3 (identity-source keys) | T2 :417-420; T5 :548; T6 :650-652 | `auth.go:66-83`: eight `Required` keys and two `Optional`, ten in all; `registry.go:352` `AddIdentitySource` |
| M5 (estimates) | Header :45-52 | See question 4 |
| RM1 / RM2 (resume, pair crash) | Records :139-145; Per-task review :135; Resume :165-183; T6 :641; Close 2 :841-843; prompt :3-11, :56-61 | Closed. The residual is V1 |
| RM3 (rows read through `tx`) | T4 :510-513, fixture :525-527, mutation :536 | `apply.go:26-44` reads through `tx`; `evaluate.go:110-121` `Evaluate` reads the read pool |
| RM5 (generation bump) | T4 :513-514, fixture :528, mutation :537-538; T7 :704 | `accessor.go:93-110` reloads only on a generation change |
| RM7 (restore channel) | States :337; T8 :724-725, acceptance :731-732 | `internal/backup/doc.go:1-9` (the auth leaf is not backed up) |

## 2. Contradictions, build order, shared files

- **Plan against prompt.** No contradiction. Both say the same for each of these: the models (T3 and T5 on `opus`,
  the rest on `sonnet`), the order, the `dispatch`, `review`, and `security-read` lines, the clock file, the branch-only
  stop, the spec path, and the instruction to read the cited sections first.
- **Plan against spec.** No contradiction. The two spec readings are judged under question 4. P4 now cites AC20 for
  the route (spec:613-614). The members-page link (M26) is navigation that AC20 leaves open.
- **Build order.** Every task builds on what the tasks before it land. T3 and T4 meet on P9's one type,
  `func(context.Context, *store.Tx) error`, with no shared file. T5 creates `api.go` and T6 extends it, in sequence.
  T10 and T11 each edit T6's inventory test, also in sequence.
- **Pair halves sharing a file.** None. I checked each pair's file list after the fold:
  - T1 and T2: `internal/ops/**` and `internal/identity` against `internal/config/**`, the cmd declaration sites, and
    the `internal/settings` tests.
  - T3 and T4: `internal/auth/**` with the cmd roster callers against `internal/settings/`.
  - T7 and T8: `member.go`, the new `settings.go`, `main.go`, `evaluate.go`, `store.go`, and their tests against
    `doctor.go`, `inspect.go`, and the runbook.

  One file is missing from T2's list (V2), but no other half touches it.

## 3. New mechanisms: quoted or probed

Each new mechanism checks out:

- **The cmd-local registrar.** It rests on the `Shell` method set, which I re-read with `go doc`.
- **`context.WithoutCancel`.** The quote matches `go doc`.
- **`strconv.NumError`.** `go doc` shows `Err` and `Unwrap`.
- **go-mail v0.8.1.** In the module cache, `client.go` dials through `DialToSMTPClientWithContext(ctx)` in
  `DialAndSendWithContext` (near `:1376`), and `:1226` is `connection.SetDeadline(time.Now().Add(c.connTimeout))`.
- **The SQLite trigger.** The fold probed it with the `sqlite3` command-line tool, not with the modernc driver. Two
  in-tree tests already inject a `RAISE(ABORT)` trigger through the driver: `Tx.Exec` at
  `internal/ops/local/recovery_test.go:123`, and `internal/engine/store/migrate_test.go:349`. So the mechanism works
  where T4 needs it. Citing that precedent would be better evidence than the probe; that is optional, and nothing
  depends on it.
- **The `go/parser` read.** It has a precedent: `internal/harness/conform/arch_test.go:48` calls
  `parser.ParseFile(..., parser.ImportsOnly)`.
- **The `pgrep` pattern (probed again).** With `grep -E`, the pattern matches `.../worktrees/admin-api-1/foo` and a
  bare `.../admin-api-1` at end of line. It does not match `.../admin-api-1-plan`. It also matches a hypothetical
  `admin-api-10`, which is harmless.
- **The `claude agents --json` filter (probed).** Every current row carries `cwd`, `name`, `state`, and `status`, so
  the `jq` `test` never meets a null.

I did not probe `git ls-remote --exit-code`; git's manual documents exit status 2 for no matching ref.

## 4. The clock question and the two spec readings

**Arithmetic.**

- Task-minutes: 40+30+50+60+50+65+100+40+30+55+55+30+90 = 695.
- Path: 40 + max(30,50) + max(60,50) + 65 + 100 + max(40,30) + 55 + 55 + 30 + 90 = 585.
- Design estimate: brief section 6 (lines 329-334) gives "about 5.5 hours" (330), so the line is 660. The before and
  after rows in the fold record match `6b4cd97:…:44-49`.

**Padding.** I found none:

- **T6 (100).** It carries the mount, the session check, an 11-operation document with kin-openapi tests and three
  proofs, V4, the route inventory with its recorder, 11 mutations, a full gate on the heavy lane, and the security
  read with its re-read. That is close to mechanics' measured 90 plus the fold's fixtures.
- **T2 (+15) and T3 (+15).** They buy, respectively, the ten identity-source keys plus seven listed test files, and
  the in-flight, epoch-read, and re-add-policy fixtures.
- **The Close (90).** It holds two `--merge` gates of up to 30 minutes each, the reviewers, the render matrix, and
  the smoke. If anything it is tight.

**Honest folds.** pass-core says "fold a single-deliverable task into a neighbor":

- **T11 into T10.** Both run inline in the session, and both are full-form gates, since `web/scripts/gates/*` is
  gate machinery (`check.sh:364-370`). They already share `deck.ts`, `deck-coverage.ts`, `serve.go`, and the
  inventory test. Merged, they keep the design whole and save one full gate and one per-task review, about 15 to 20
  minutes. That reaches about 675 to 680, not 660.
- **T8 into T7.** This saves one gate in task-minutes but lengthens the path, so it is not worth recommending.
- **Moving the inventory.** The fold record's own cut (the route inventory to `admin-api-2`) saves about 20 minutes.
  Spec:530 does give the inventory to "the first lane to register a route", and 14a is that lane, so this cut is a
  scope choice, not a free one.

**Can Geoff rule on the evidence?** Yes. The record gives the number, the reason, and each option's cost. Two facts
would make it complete (V4):

- the T10/T11 merge as a cut that keeps the design whole;
- that the brief itself expects "toward 7 hours" (420 minutes, a line of 840) once `auth-data` fix rounds are
  counted. The plan header cites this; the ruling does not.

**Spec reading 1: hook 18 and the identity-source keys. Faithful, no design change.**

- Spec:195 opens with "Every registered key carries a description and a sensitive mark". The list after the colon
  (settings, thresholds, wall rows, and integration-class keys) names the families SS asked for.
- AC9 needs `list_settings` to answer "every registered key's name, home, description, apply mode, and mark", and the
  ten identity-source keys are registered (`registry.go:352`). Without T2's extension, T5 cannot meet AC9.
- Spec:204-207 gives an environment-only key's mark the job of flagging "a credential", and the only credentials in
  the registry are identity-source keys.
- The extension adds work (T2 +15) and no new surface.

**Spec reading 2: "any removal or demotion" in the runbook. Faithful, no design change.** The spec's reason is that a
restore brings back state made after the snapshot (spec:339, :394-395), and that reason covers a demotion the same
way. The line is runbook prose, not code, and AC21 only asks that the restore line read as the spec describes. The
same reason also covers settings writes (V5).

## Findings, ranked by consequence

### V1. minor, correctness. Resume's candidate rule matches the session's own branch, so every mid-pair resume stops

- **Location:** plan:170-180 ("For each `dispatch` line …"); prompt:7-9.
- **Defect:** the session commits, then dispatches, so its own branch `worktree-admin-api-1` sits at the dispatch
  sha. `git merge-base --is-ancestor <dispatch sha> worktree-admin-api-1` holds, so the session's own branch is a
  candidate. `git log --format=%s <dispatch sha>..worktree-admin-api-1` prints nothing. The third arm then reads "it
  has no `T<n>:` commit: a stop". A crash mid-pair therefore always stops, even when both halves committed.
  Earlier `dispatch` lines are also ambiguous. Every later subagent branch descends from an earlier dispatch sha, so
  for the `T1+T2` line, a committed but unmerged T3 branch "carries its `T1:` commit and is not an ancestor of
  `HEAD`", and the rule says to merge it under the wrong line, with the wrong `--base`. This fails closed: no work is
  lost and no review is skipped. It costs one attended stop that RM2's fold meant to avoid.
- **Proposed fold:** "Only the last `dispatch` line can be unsettled, since the session merges and reviews a dispatch
  before the next. Classify that line's candidates, excluding the session's own branch."

### V2. minor, correctness. T2's file list misses `templatekeys_test.go`

- **Location:** plan:428-431 (T2 Files).
- **Defect:** `cmd/dubplate-server/templatekeys_test.go:99` reads `class.Required` and `:109` reads `src.Required`
  and `src.Optional` as `[]string`. T2 turns both into declarations. The test stops compiling unless T2 edits it, and
  T2 lists it nowhere. Rm7's grep looked for the `Add*` and declaration forms, not for readers of the fields. T1 does
  not touch the file, so this is no pair conflict. It does break the plan's file list, and an executor that holds to
  the list could stop on it.
- **Proposed fold:** add `cmd/dubplate-server/templatekeys_test.go` (the readers of `.Required` and `.Optional`) to
  T2's Files.

### V3. minor, correctness. No fixture proves a sensitive reset sends the notice

- **Location:** plan:587-590 (T5 notice fixtures), :597-601 (mutations); spec:280 ("Every committed `admin`-source
  write or reset of a sensitive key").
- **Defect:** the recording-mailer fixture names a write. The reset has only the zero-send counter-case, a
  non-sensitive reset. A mutant that sends no notice on `reset_setting` passes every named fixture.
- **Proposed fold:** the recording-mailer fixture runs one sensitive `set_setting` and one sensitive
  `reset_setting`, each sending one message per live admin. Add the mutation "send no notice on a sensitive reset".

### V4. minor, optional. OWNER FORK: the clock ruling omits a cut that keeps the design whole

- **Location:** fold record :125-134 ("Rulings for Geoff").
- **Defect:** the "No" option lists only scope cuts. It leaves out T11 folded into T10 (about 15 to 20 minutes, no
  scope change, by pass-core's own "fold a single-deliverable task into a neighbor"). It also leaves out the brief's
  "toward 7 hours" with `auth-data` fix rounds (a line of 840 on that figure), which the plan header cites.
- **Options:**
  - **(a) Approve 695 as planned.** Recommended. Every added minute buys a proof a lens ranked major, and the
    per-task clock still bounds each task.
  - **(b) Approve with T11 merged into T10** (about 675 to 680). This saves one full gate and one review, and both
    halves stay in one commit series that the session reads.
  - **(c) Return the plan for (b) plus moving the inventory to `admin-api-2`** (about 655 to 660). This is a scope
    cut against spec:530.
- **Proposed fold:** add the two facts and option (b) to the ruling, so Geoff sees every cut that keeps the design
  whole. The recommendation stays (a).

### V5. minor, optional. The runbook's restore line stops at privilege, while the same reason covers settings

- **Location:** plan:337 (States: "restoring a reset row"), :724-725 (T8).
- **Defect:** the fold widened the line from "removal" to "removal or demotion" on the reason that a restore brings
  back state from before the snapshot. That reason also brings back any settings write or reset made since, which
  the plan's own States row names. A sensitive setting tightened since the snapshot (a shorter session lifetime, say)
  returns at its looser value, and the runbook does not say so. Widening further is a third reading of the spec, so
  it is optional.
- **Proposed fold:** either the line says "re-run any removal, demotion, or settings change made since the
  snapshot", or the fold record says why the line stops at privilege: membership is what the spec's sentence guards.

### V6. minor, optional. AC9's sentinel fixture says "free text", then lists ports, and gives no validity rule

- **Location:** plan:650-652 (T6 AC9 fixture).
- **Defect:** the line says the sentinel goes on each key "whose value is free text", then names "each port and host".
  A port is not free text, and it must still pass its validator. A secret's sentinel must meet the 32-byte floor
  (`auth.go`, `secretLengthFloor`), or the server will not start. An implementer can work this out, but the wording
  invites a sentinel that cannot be set.
- **Proposed fold:** "a distinct sentinel on every non-store-homed key except the closed enumerations, each valid for
  its key's validator (each port a distinct unused number, each secret at least 32 bytes)".

### V7. minor, optional. Over-ceremony, ranked by clock and token cost

1. **The A/B session choreography in the smoke** (plan:851-859). It costs clock: repeated magic-link sign-ins
   through the sink every 5 minutes, across a smoke that runs well past 5 minutes. A cheaper equivalent: run every
   gated step under the default `2h` window, then restart `serve` once with the window seeded at `5m`. Sessions
   survive a restart, so session A, signed in at the start, is past the window, and its sensitive write is refused.
   That is one restart instead of N sign-ins, and it proves the same D10 clause. It saves perhaps 5 to 10 minutes of
   the Close.
2. **Every dispatch gets "Conventions every task follows" verbatim** (plan:155-158; prompt:42-43). It costs tokens:
   the section is now about 100 lines, and its Records, Clock, Resume, Per-task review, Stops, and Dispatches bullets
   bind the session, not the subagent. The prompt's eight rules already restate the subagent's share. Across five
   dispatched tasks (four dispatches), that is roughly 4,000 to 6,000 tokens of instruction a subagent cannot act on,
   which also raises the chance it acts on the session's rules (writing a clock file, appending to the PR notes). A
   cheaper equivalent: pass the bullets that bind a subagent (Gate, base, Setup, Test-first, Mutation proofs, Fixture
   states, Layering, No export, Config keys, No migration, Skills, Commits) and leave the rest to the session.

Nothing else the fold added reads as ceremony:

- The `dispatch`, `review`, and `security-read` lines are one line each and carry real recovery state.
- The security reviewer re-reading its own fix is one dispatch on an `auth-data` pass.
- The trimmed render matrix lowered the clock.
