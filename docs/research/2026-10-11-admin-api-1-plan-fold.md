# admin-api-1 plan fold

Target: `docs/superpowers/plans/2026-10-11-admin-api-1.md` at `6b4cd97`, and its launch prompt
`docs/plans/2026-10-11-admin-api-1-prompt.txt`. Reviews folded:
`2026-10-11-admin-api-1-plan-review-contract.md` (F1 to F19), `-review-mechanics.md` (M1 to M26), and
`-review-risk.md` (its majors M1 to M7 appear here as RM1 to RM7, its minors m1 to m12 as Rm1 to Rm12). Fold agent:
Opus 5.5, under `superpowers:receiving-code-review` and `spec-plan-review`'s "The fold". Every finding was checked
against the worktree at `6b4cd97` (code unchanged since `551bf4f`) before it was acted on. The approved design
(spec revision 2.1) is unchanged.

## Convergent roots

Each root folds once; its rows below point here.

| Root | Findings | Verified against | Folded at |
|---|---|---|---|
| A. The roster's refusals are unexported | M1, RM6 | `internal/auth/roster.go:39-105` (every sentinel lower-case; the comment at `:55-59` says only in-package code branches on them); `admin.go:312-337` maps them in-package | P9 (the sentinels named, T3 exports them, T5 is the caller); T3 outcome, files, acceptance (`errors.Is` from outside); T5 outcome (the mapping) |
| B. The detail enumeration | M4, F4 | `web/src/routes/members/+page.svelte:46-54` holds `duplicate_principal` and lacks `invalid_address`; `handlers.go:57` declares `invalid_address`, returned for adds at `admin.go:324-325`; spec:95-96 keeps each add outcome's name | P7 lists the ten values outright; T6 adds a code and detail parity test against `ops.Codes()` (exists, `go doc ./internal/ops Codes`) and the `Detail*` constants read by `go/parser`; T6 mutation "add a `Detail*` constant the document omits" |
| C. The `busy` fixture reaches one arm | F3, M6 | `go doc ./internal/engine/store Handle.Update` ("a deadline shorter than the busy timeout does not shorten it"); `write_test.go:573-606` (in-process hold gives `DeadlineExceeded`; a second handle gives `ErrBusy`, measured) | Behaviors (the full quote and the measured note); T5 fixture split into both arms, through `store.OpenWrite`'s `busyTimeout` (`go doc ./internal/engine/store`); T5 mutation "map `context.DeadlineExceeded` to `internal`" |
| D. Resume and in-flight records | M13, RM1, RM2, Rm8 | plan Records and Resume conventions: the dispatch sha and branch reach the notes only through the subagent's report; no review verdict is recorded; the prompt checks the directory alone. Probed: `git worktree list --porcelain` prints `branch refs/heads/...` per entry; `git ls-remote --exit-code --heads origin <none>` exits 2; `git rev-parse --verify --quiet` exits 1 on no branch | Records (`dispatch` line before each dispatch); Per-task review (`review T<n>=` line); T6 (`security-read=` line); Resume rewritten (launch-base recovery, three-way candidate classification, unreviewed task is in flight); Close step 2 (unreviewed tasks handed to the whole-branch reader); prompt: branch-or-directory check, the stop when the branch exists alone, the dispatch and review lines |
| E. The smoke's 5m window | M16, Rm3 | plan Close step 4 (every gated step ran inside the wait); D10 gates every member call | Close step 4: session A ages, session B runs the gated steps and re-signs near 5 minutes, A's refused write comes last |
| F. The `--base` departure is unnamed | F19, Rm11 | `dubplate-pass` SKILL.md:71 ("Pass no `--base`"); `check.sh` behavior per the mechanics review's probe | Conventions (one sentence naming the departure); Close step 7 ("Found, not done" carries the adapter amendment) |
| G. The close render matrix | F18, Rm12 | plan Close step 3 and End-to-end item 3 (two pages x two palettes x two widths x every refusal state) | End-to-end item 3 and Close step 3: base states in the full matrix, each refusal once at phone width in one palette, the re-authentication flow in both palettes; Close estimate 100 to 90 |
| H. The deck gate's vacuity guard | F5, M12 | `deck-coverage.ts:29` (the `outcomeX` pattern) and `:155-167` (three names and a floor of 21); BACKLOG #21 asks for "non-empty and names one known outcome" instead of an exact floor | T9 (the details parse guarded non-empty and naming `last_admin`, never an exact count; mutation); T11 (the guard narrows to `handlers.go` and `signed_out`; mutation); #21 joins the scope line and the Close ledger. F5's exact count is not taken, since #21 records that an exact floor reddens on every retirement |
| I. The demoted actor in flight | F1, RM4 | spec:252-253 ("so a demotion is final for writes already in flight"); every named fixture demoted before the call; `seam.go:140` reads `Admin` per request, so a mount-driven test refuses before the seam; P9 makes the guard optional by design | T3 in-flight fixture on AC16's pattern; T5 direct-seam fixtures (demoted and tombstoned rows) and the in-flight settings write; mutations "read the actor's flag outside the `Update`" and "pass no guard to the settings write". "By address" leaves the mutation list, since `ops.Actor` carries no address |
| K. The AC9 sentinel set and hook 18's families | F11, M3 | `registry.go:352` `AddIdentitySource`; `cmd/dubplate-server/auth.go:66-83` registers ten bare-string keys (the session and link secrets, the SMTP password among them); `DUBPLATE_IDENTITY_SOURCE` is closed-set (`auth.go:307`) | T2 extends the declarations and panics to `IdentitySource` keys, reading the spec's governing "every registered key" (see "Spec readings"); T5 states no environment-only value crosses; T6 names each sentinel, excludes the closed-set rows, and keeps the AC9 server apart from AC21's |

## Findings

### Contract lens

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| F1 | major | spec:252-253; plan fixtures | Folded, root I |
| F2 | major | `roster.go:257`, `handlers.go:420`; `go doc ./internal/auth EpochStore` (`Read` exists) | Folded: T3 AC15 fixture reads the epoch after the removal; mutation "skip the removal bump" |
| F3 | major | `write_test.go:573-606` | Folded, root C |
| F4 | major | `+page.svelte:46-54`, `handlers.go:57` | Folded, root B |
| F5 | major | `deck-coverage.ts:155-167`, BACKLOG #21 | Folded, root H (#21's shape over the exact count) |
| F6 | major | spec states table (spec:485-506) against each task's fixtures; each listed gap confirmed unfixtured | Gaps folded as fixture lines: not found and the non-admin demote (T3, T5), each add detail (T5), wall, unregistered, env-equal, env-sets reset (T4), the five gated calls and two ungated counter-cases (T5), roster-lookup 500 (T6), `auth.Open` refusals and `member list` (T3, T7), `serve` older store with an `admin` row (T7), doctor's older store (T8), removed target (T7). The 31-row ledger table is refused: it repeats the spec's table and the fixture lines carry the same assignment |
| F7 | minor | plan T6 mutations name a bearer read with no fixture | Folded: T6 bearer, actor-field, and `slog` capture fixtures |
| F8 | minor | `evaluate.go:265`, `:338-351`; `config/settings.go:182-186`; `go doc strconv.NumError` | Folded: Behaviors quote; T4 fixture uses an int-kind key and `strconv.ErrSyntax` |
| F9 | minor | spec:71-74; the rename proof uses a `required` field | Folded: T6 mutation annotation names the omitted-Go-field proof |
| F10 | minor | no test checks the plan's layering rules (`internal/harness/conform/arch_test.go` checks engine to harness and the server's link set only); SQLite trigger probed | Folded in part: Layering gains "`internal/settings` imports no `internal/notify`"; T4 trigger fixture and mutation annotation; T5 zero-send fixtures. Refused: a new import check, which is new machinery with no measured defect; the rule stays where the other layering rules are, read by the per-task review |
| F11 | minor | `auth.go:307`; T6's AC21 fixture names a member | Folded, root K |
| F12 | minor | the plan's contract tests read neither fact | Folded: T6 asserts the 409 on `start_backup` and the policy text |
| F13 | minor | plan T9 and T10 rows; `run.ts` structure | Folded: each row names its interception; the confirm row counts writes; the key row compares against the production registry's list; a `restart`-key row (four `ApplyRestart` keys exist); a T9 re-authentication row |
| F14 | minor | `config.go:216` lowercases; `roster.go:144-149` normalizes | Folded: T8 runbook greps and the mixed-case fixture. No mutation added: config already lowercases, so a mutation on doctor's comparison could survive vacuously |
| F15 | minor | `admin.go:113`, `:120-121` | Folded: Behaviors quote; two States rows; T6 fixtures and the mutation "refresh without the same-origin check" |
| F16 | minor | spec:280-285 sets no content; the old plan's wording | Folded: kept, and T5's recording mailer asserts the four fields |
| F17 | minor | plan:265-266 | Folded: the quotes for F3, F8, and F15 land in Behaviors; the catch-all line stays |
| F18 | minor, optional | see root G | Folded, root G (the caller named it convergent; it also lowers the clock) |
| F19 | minor, optional | see root F | Folded, root F (free, and a resumed session would otherwise "correct" the base) |

### Mechanics lens

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| M1 | major | `roster.go:39-105` | Folded, root A |
| M2 | major | `go doc ./internal/serve Shell` (method set `Register`, `Handler`, `Logger`, `EnableHSTS`); registrations at `serve.go:344-366` and `web.go:186-194`; PS:620 (A2) | Folded: T6 routes registrations through a one-method interface local to `cmd/dubplate-server` that `*serve.Shell` satisfies, a recording wrapper in the test, and the unauthenticated probe per session and admin row; `web.go` joins T6's files; mutation "drop `RequireAdmin` from one admin route" |
| M3 | major | `registry.go:352`, `auth.go:66-83` | Folded, root K |
| M4 | major | see root B | Folded, root B |
| M5 | major, optional | the plan's estimates against recorded gate times | Folded at the clock with honest re-estimates (T5 65, T6 100, T9 55, T10 55, plus the fold's own additions); the total passes the 660 line, see "Rulings for Geoff" |
| M6 | minor | `go doc ./internal/engine/store Handle.Update` | Folded, root C |
| M7 | minor | `Makefile:43-45`; `check.sh:23` | Folded: T0 stages `go.mod`, `go.sum`, and `web/package-lock.json` before its gate |
| M8 | minor | `web/package.json:21` (`"playwright": "1.64.0"`) | Folded: T0 holds a newer Playwright to the PR body; Close step 7 names it |
| M9 | minor | `store_test.go` near `:731` and `:741` assert the text; `:770` is an `env-seed` parse row | Folded: T7 files gain `store_test.go` and the evaluate tests; the parse message names the verb only for an `admin` row; fixture |
| M10 | minor | `config.go:139`, `:429-431`; `keys.go:170` | Folded: T2 files |
| M11 | minor | `web_test.go:61-65`, `:114-120` | Folded: T10 files |
| M12 | minor | see root H | Folded, root H |
| M13 | minor | see root D | Folded, root D |
| M14 | minor | `claude agents --json` keys (`cwd`, `name`, `state`, `status`); the finished `dubplate-alignment` row reads `"state":"done","status":"busy"`; adapter's bracket-class rule | Folded: the `jq` filter anchors on `/worktrees/admin-api-1$`; a `pgrep -af '[P]rojects/dubplate/.claude/worktrees/admin-api-1([^-]|$)'` line, probed to match a process under `admin-api-1/` and not under `admin-api-1-plan/` (`man pgrep`: an Extended Regular Expression); the "idle or gone" sentence goes |
| M15 | minor | plan Dispatches; the prompt | Folded: each dispatch carries the spec's path and the instruction to read the cited sections and ACs (plan and prompt) |
| M16 | minor | see root E | Folded, root E |
| M17 | minor | `notify.go:140`; go-mail v0.8.1 `client.go:1375-1391` (ctx reaches the dial) and `:1226` (`SetDeadline` from `connTimeout`) | Folded: Behaviors quote, with the finding that the context bounds the dial and the client timeout bounds the connection |
| M18 | minor | `conformance.go:27-30` imports `routers/legacy`; `gorilla/mux` only in `go.sum:19-20` | Folded: T6 names `routers/legacy` |
| M19 | minor | `run.ts` near `:3691` expects `TITLES.notFound` for the non-admin | Folded: T9 renders a list-time `forbidden` as the not-found state |
| M20 | minor, optional | P9 named no type while T3 and T4 build both sides in parallel | Folded: free, and removes a real pair-mismatch risk (P9 names `func(context.Context, *store.Tx) error`) |
| M21 | minor, optional | plan Dispatches | Refused: efficiency only; the subagent's own gate is the red signal the "returns uncommitted" rule rests on |
| M22 | minor, optional | adapter path map (SKILL.md:91-93): `runner` is `scripts/` and `.claude/` | Folded: free; the plan contradicted the adapter, and only Geoff raises a class. T0 is `ordinary` |
| M23 | minor, optional | plan Clock convention | Dropped: no risk; the Close carries an estimate and pass-core's clock covers tasks |
| M24 | minor, optional | visual identity spec :143-147 reserves no route; AC20 (spec:615-616) is the source | Folded: free citation fix in P4 |
| M25 | minor, optional | plan T1 and T5 | Folded: free, the plan's own convention; T1 and P9 name the later callers |
| M26 | minor, marked owner fork | `SiteHeader.svelte` has no navigation list (the wordmark link only); `run.ts` near `:3706-3710` asserts no "Members" link for a non-admin | Taken, not forked: option 3, a link from the members page only (P4, T10). A static page cannot tell an admin from a member, the non-admin rendered row asserts the header carries no admin link, and option 3 adds no role detection. One answer is clearly correct, so it is recorded rather than asked |

### Risk lens

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| RM1 | major | see root D | Folded, root D |
| RM2 | major | see root D | Folded, root D |
| RM3 | major | `evaluate.go:110-121` (`Evaluate` reads the read pool); `apply.go:26-44` (rows read through `tx`); `Handle.Update`'s doc confines its read-inside rule to the rows a write changes | Folded: T4 outcome reads both evaluations' rows through `tx` on `Apply`'s precedent; joint-break fixture; mutation "read the rows through `Evaluate` before `Update`" |
| RM4 | major | see root I | Folded, root I |
| RM5 | major | `accessor.go:93-110` reloads only on a generation change; T4 had no skip-the-bump mutation | Folded: T4 outcome (the bump on every committed set and reset), accessor fixture, and mutation; T7 serve-mode handle fixture |
| RM6 | major | see root A | Folded, root A |
| RM7 | major | spec:339 and `internal/backup/doc.go:1-9` (the auth leaf is not backed up); spec:394-395 | Folded: a privileged-field row for a store restore; T8's runbook line reads "any removal or demotion" (see "Spec readings") |
| Rm1 | minor | plan T6 | Folded: the security reviewer re-reads its own fix, its blocker and major findings count as `fix`, its `escalate` is a stop |
| Rm2 | minor | spec:337-338 ("in the same callback and order as removal") | Folded: T3 states the removal's policy for the re-add; mutation "roll back a removal or a re-add on `ErrEpochBackward`" |
| Rm3 | minor | see root E | Folded, root E |
| Rm4 | minor | `go.mod:8,10,13` (kin-openapi v0.149.0, go-mail v0.8.1, sqlite v1.60.1) | Folded: one sentence after Behaviors names who re-quotes after a T0 bump |
| Rm5 | minor | `auth.go:146-155` (cap floor `1s`) | Folded: the `serve`-start States row names both remedies and the floor case |
| Rm6 | minor | spec states row names wall and unregistered keys only | Folded: T4 fixture adds a threshold and an integration-class key |
| Rm7 | minor | grep of `AddSetting`, `config.Setting{`, `NewRegistry(` across tests | Folded: T2 files list each test that declares a key or builds a registry |
| Rm8 | minor | see root D | Folded, root D |
| Rm9 | minor | spec:352-353 (a commit failure after a bump leaves a live member signed out) | Folded: the `busy` States row says "no store row written" and names the bump |
| Rm10 | minor, optional | `go doc context.WithoutCancel` | Folded: free, and removes a real risk (a client that disconnects after a sensitive commit would cancel the notice); T5 and Behaviors |
| Rm11 | minor, optional | see root F | Folded, root F |
| Rm12 | minor, optional | see root G | Folded, root G |

## Spec readings (no erratum)

The approved spec is not edited, and errata on frozen documents stay unplanned (decision already taken). Two
places where the plan reads the spec past its enumerated wording, each because the spec's governing sentence and
its stated reason decide it:

1. Hook 18's family list ("settings, thresholds, wall rows, and integration-class keys", and AC8) omits
   `IdentitySource` keys. The governing sentence is "Every registered key carries a description and a sensitive
   mark", `list_settings` answers "every registered key", and those keys are registered (`registry.go:352`). T2
   covers them.
2. The runbook's restore line says "re-run any removal". The spec's reason, that a restore brings back state made
   after the snapshot, applies equally to a demotion (the users table carries the flag; the auth leaf is not backed
   up). T8 writes "any removal or demotion".

## Rulings for Geoff

1. **Approve the 695 task-minute estimate?** pass-core sends a plan back when task-minutes exceed twice the design's
   estimate: the design's 330 minutes puts the line at 660, and the folded plan sits at 695 (path about 585). The rise
   from 535 is the review's proofs on an `auth-data` pass: honest re-estimates of T5, T6, T9, and T10 (mechanics M5),
   plus the fold's own fixtures in T2, T3, T4, T7, and T8, less 10 minutes from the trimmed render matrix. The
   planning brief itself expected more than its 330 once fix rounds count: "The `auth-data` fix rounds 6d saw would
   push it toward 7 hours" (`research/2026-10-10-next-pass-planning-brief.md:334`), which is 420 minutes and a line
   of 840 on that figure.
   - **(a) Approve 695 as planned (recommended).** The per-task clock stop still bounds each task at twice its
     estimate, and every added minute buys a proof a lens ranked major or a fixture that keeps a test from passing
     vacuously.
   - **(b) Approve with T11 folded into T10**, about 675 to 680. pass-core's "fold a single-deliverable task into a
     neighbor": both run inline, both run the full gate, and they already share `deck.ts`, `deck-coverage.ts`,
     `serve.go`, and the inventory test, so the merge saves one full gate and one per-task review and keeps the design
     whole. It still sits above 660.
   - **(c) Return the plan for (b) plus moving the route inventory to `admin-api-2`**, about 655 to 660. This drops
     work the design gives this pass: PS:580-581 routes the inventory to "The next rung that registers a route", and
     the spec's surfaces table (spec:530) gives it to the first lane to register one, which is this pass. It is a
     scope cut, not a free saving.

## Measures

- Plan length: 706 lines before, 877 after. Prompt: 61 lines before, 68 after.
- Task-minutes: 535 before, 695 after (T2 35 to 50, T3 45 to 60, T4 35 to 50, T5 45 to 65, T6 50 to 100, T7 30 to
  40, T8 25 to 30, T9 35 to 55, T10 35 to 55, Close 100 to 90). Path: 445 to 585.
- Counts: 64 findings (19 contract, 26 mechanics, 19 risk). 59 folded in full, M26 among them (marked as a fork,
  taken as option 3); 3 folded in part (F6's gaps without
  its table, F10 without a new import check, F14 without a mutation); 2 dropped or refused (M21 refused, M23
  dropped). One owner question: the clock.
- New mechanisms folded, each quoted or probed: the PR-notes `dispatch`, `review`, and `security-read` lines (on the
  plan's own `launch-base=` line, for pass-core Close's "give any unread `auth-data` task its read"; defect read in
  the Records convention, where a crash mid-pair loses the dispatch sha and branch); the cmd-local registrar
  interface (`go doc ./internal/serve Shell`'s method set; PS A2's "a route is added without a row"); the SQLite
  audit-trigger fixture (probed with `sqlite3`); `context.WithoutCancel` (quoted); the `go/parser` read of the
  details file (precedent `internal/harness/conform/arch_test.go`); the `pgrep` ERE pattern (probed); `git worktree
  list --porcelain` and `git ls-remote --exit-code` (probed).
- New mechanisms refused: a layering import-check test (F10), a 31-row states ledger table (F6), a single-dispatch
  gate receipt (M21).

## Second fold

The verification read (`2026-10-11-admin-api-1-plan-fold-verification.md`, against `f546d61`) found 0 blockers, 0
majors, and 7 minors. Each was checked against the tree before it was acted on.

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| V1 | minor, correctness | the session commits before it dispatches, so its own branch sits at the dispatch sha and passes the ancestor test with no `T<n>:` commit; an earlier line's sha is an ancestor of every later subagent branch | Folded: Resume classifies only the last `dispatch` line, since each dispatch is merged and reviewed before the next, and drops the session's own branch; the prompt's resume paragraph says the same |
| V2 | minor, correctness | `cmd/dubplate-server/templatekeys_test.go:99` reads `class.Required`, `:109` reads `src.Required` and `src.Optional` as `[]string` | Folded: T2's Files |
| V3 | minor, correctness | spec:280 ("Every committed `admin`-source write or reset of a sensitive key"); the only reset fixture was the non-sensitive zero-send case | Folded: T5's recording-mailer fixture runs a sensitive `set_setting` and a sensitive `reset_setting`, each one message per live admin; mutation "send no notice on a sensitive reset" |
| V4 | minor, optional, owner fork | brief:334; PS:580-581; spec:530 | Folded into "Rulings for Geoff", ruling 1, as options (a), (b), and (c), recommendation (a) |
| V5 | minor, optional | plan States restore row; T8's runbook line | Refused: the spec's restore line guards access (a removed or demoted member regains it); a restored setting grants no one access and shows with its source on the server settings page, so a third widening of a spec sentence is not worth its reading |
| V6 | minor, optional | `secretLengthFloor = 32` (`cmd/dubplate-server/auth.go:93`); ports and hosts are validated | Folded: free, and it removes a real risk (a sentinel that stops the server makes AC9's scan vacuous). T6's AC9 fixture: every sentinel valid for its key's validator, each port a distinct valid port, each secret at least 32 bytes |
| V7 | minor, optional | plan Close step 4; Dispatches convention | Refused: the A/B smoke costs a few sign-ins against a loopback sink, while the restart variant rests on an unprobed claim that sessions survive a restart; the verbatim Conventions keep each dispatch identical to what the reviewers read, and the prompt's eight rules already name the subagent's share |

The SQLite trigger mechanism also has two in-tree precedents through the modernc driver
(`internal/engine/store/migrate_test.go:349`, `internal/ops/local/recovery_test.go:124`), which back the earlier probe.
Task-minutes are unchanged at 695: V1 to V3 and V6 are a sentence or a fixture line each.
