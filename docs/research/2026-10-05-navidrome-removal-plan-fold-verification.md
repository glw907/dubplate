# Navidrome removal plan: fold verification

Target: `docs/superpowers/plans/2026-10-05-navidrome-removal.md`, revision 2 at `13bb9db`.
Fold record: `docs/research/2026-10-05-navidrome-removal-plan-fold.md`.
Method: a fresh-context read of the plan, the fold diff (`git diff HEAD~1`), the three reviews, the spec (rev 2.1),
and `~/.claude/workflows/pass-execute.js`. Each claim below was checked against the tree or by running it. The
reviewer brief's warning applies: only gaps that affect correctness or the stated requirements are ranked above
minor, and none here reaches that bar.

**Totals:** 9 findings. 0 blocker, 0 major, 9 minor.

## Question 1: did the cited findings close where the fold says?

Each spot check holds.

| Claim | Verified |
| --- | --- |
| PC2: `gateLane` and `reviewerModel` are real args that override the class defaults | Yes. `pass-execute.js:359` computes `(t.gateLane \|\| a.gateLane \|\| (cls && cls.gateLane)) === "light"`, so `a.gateLane: "heavy"` beats `tool`'s `gateLane: "light"` (`:267`). Any value other than `"light"` works; the header comment (`:65-67`) documents only `"light"`, so "heavy" is a convention, not a recognized value. `:569` reads `a.reviewerModel \|\| cls.reviewerModel`, so the Opus override beats `sweep`'s `"sonnet"` (`:252`). `auth-data` exists (`:226`). |
| PC1: `TestRoute` separates the destinations | Yes. `prefix_test.go:162-166` is the neighbor row (`want: terminalFallthrough`). The loop (`:198-224`) builds distinct `registered` and `fellThrough` recorders and asserts exactly one was called. With `Owns` as a bare `HasPrefix`, `/anchorfoo` reaches `registered` and the row fails. `Route` calls `Owns` (`prefix.go:151`). |
| Reading 2: `client.go:392` wraps `ErrUnavailable` | Yes: `return nil, fmt.Errorf("%w: %w", ErrUnavailable, lastErr)`. `tag_test.go:274` asserts it. `run.go:271` is the only non-test `ClassBackground` caller. |
| PC4: the `env -i` method is runnable | Yes. A binary built with `go build ./cmd/dubplate-server`, run as `env -i HOME="$HOME" PATH=/usr/bin:/bin ./dps check`, prints five `required setting unset` rows and exits 1. That matches the review's first round. `"configuration: ok"` is `check.go:34`. `ErrMissing`'s text is `config.go:79`. |
| PC6: the uncapped leg | Runnable as `{ echo "work_dir=<scratch>"; sed -n 566,611p scripts/check.sh; } \| bash`. At HEAD, `^internal/serve` gives exactly the nine names Reading 1 lists, `NewFrontExemption` among them. |
| PC12: pflag | `pflag@v1.0.10/errors.go:40` is `return fmt.Sprintf("unknown flag: --%s", e.name)`. |
| PC5, PR10 | `main.go:110` defines `newProductionRegistry`, which calls `registerShadowIntegration` at `:112`. `serve_test.go:97` defines `productionRegistry`. `auth_test.go:73-74` and `:239-241` are the `_PREVIOUS` rows. |
| PC11, PM11, PC9 | `corpus/README.md:2288-2304` documents the `album` plus `release_mbid` subject. `BACKLOG.md:42-43` names `--streamer-username`. |
| PM3, PM6 | `cookieValue` (`:289`) and `cookieLines` (`:475`) are called only from tests the table deletes. `TestStripWrapEmptyExemptionExemptsNothing` and `TestStripWrapRemovesMixedSeparatorSpellings` are the remaining `X-Driver-Identity` users, and the table disposes of both. |
| Task 1 Flush mutation | Holds by source. `net/http/responsecontroller.go` `Flush` walks `FlushError`, `Flusher`, `rwUnwrapper`, and else returns `errNotSupported()`. The plan did not quote this (PV7). |

## Findings

### PV1 (minor): Task 2b's comment scope contradicts its file list

- **Where:** plan `:401` (Files), `:414-416` and `:433-434` (Outcomes).
- **Defect:** The fold moved Task 2b to comments only and dropped `limiter_test.go` and `client_test.go` from the
  Files list. The outcome still says "Every remaining comment naming the shadow ... is rewritten", and "The
  limiter's behavior and tests are unchanged". `limiter_test.go:39-109` uses `"shadow"` as a budget label and
  says "shadow waiter" in its comments. The cited ranges also miss two shadow comments in the listed files:
  `limiter.go:68` ("rung 4's shadow are three real concurrent ...") and `client.go:340` ("rung 4's shadow budget").
  Ruling 6's "reads as if it had never had the removed part" leaves these, and the segment pre-flight will not
  catch them, because the cited ranges are accurate as far as they go.
- **Proposed fold:** Files: `limiter.go` (`:11-41`, `:68`), `client.go` (`:74-84`, `:340`), and `limiter_test.go`
  (comments, plus the `"shadow"` label renamed to a neutral `"background"`, assertions unchanged). Alternatively,
  drop the line ranges and say "every shadow comment in these files".

### PV2 (minor): Task 2a's named mutation cannot be applied at the task's commit

- **Where:** plan `:386-387`.
- **Defect:** "re-add `registerShadowIntegration(reg)` to `newProductionRegistry`" refers to a function the same
  commit deletes with `shadow.go`. As written, the mutation does not compile.
- **Proposed fold:** "Run the new test at the pre-deletion commit (the task's base). It fails there, since the
  shadow class is registered and the key is not unknown. The report shows that failure."

### PV3 (minor): Task 2b's "before Task 2a" leg run has no tree to run in

- **Where:** plan `:419-421`.
- **Defect:** Task 2b runs after Task 2a has committed. Inside one `pass-execute` segment, the 2b implementer has no
  pre-2a tree. `git stash` is forbidden, and a second worktree is lane machinery. The runner does not pass 2a's
  report to 2b.
- **Proposed fold:** Segment A's haiku pre-flight (conductor action 3) records the uncapped list at HEAD, which is
  Task 1's base, and the conductor carries it in Task 2b's notes. Task 2b diffs its own after-run against that
  list. Reading 1's serve list already pins Task 1's delta.

### PV4 (minor): Task 3's ruling 7 comparison names no channel for Task 0's set

- **Where:** plan `:502-503` and `:228-229`.
- **Defect:** Task 3's acceptance compares "key by key" with Task 0's set. Task 0's set lives in "the pass record",
  which is not a committed file the implementer can read. Nothing carries the set or the sample values into the
  dispatch.
- **Proposed fold:** Conductor action 4 also carries Task 0's classified set and its sample environment in Task 3's
  dispatch notes, as it already does for Task 4a's re-confirmation date.

### PV5 (minor, predates the fold): the doctor reserved-class order test becomes partly vacuous

- **Where:** plan `:481` (Task 3: "pins table order against two surviving classes"); `doctor_test.go:1048-1063`;
  `keys.go:59-105`.
- **Defect:** The test exists to catch a sort by rung (R15: queues at rung 13 print before streamer at rung 12).
  Once `streamer` is deleted, the table reads intake 7, pipeline 9, monitoring 9, queues 13. That is already rung
  order, so a stable rung sort produces table order, and no pair of surviving classes can fail under the
  regression the test was written for. The plan names no mutation. The planning-miss rule ("each proof names the
  fixture state it needs so the test cannot pass vacuously") applies. No reviewer raised this.
- **Proposed fold:** Pin `pipeline` before `monitoring` (same rung, table order), with the named mutation "sort by
  name", which fails it. The test's doc states that a rung sort can no longer be told apart from the table.
  Alternatively, delete the order assertion and record that its subject left with the streamer class.

### PV6 (minor): Task 5's slot beside document segment B is unstated

- **Where:** plan `:86-88`, `:625-626`; document plan `:35-37`, `:84` (Task 10 "After the removal pass's own
  close").
- **Defect:** Document segment B launches once close steps 1 to 4 commit. Task 10 of the document pass waits for
  this pass's Task 5. "A segment here does not launch while a document segment is in flight" then holds Task 5
  until the document pass's Tasks 5 to 8 invocation ends. That order works, because the document pass's Tasks 9
  and 10 run outside the invocation. The plan does not state it, and a conductor could run Task 5 inside the
  document invocation.
- **Proposed fold:** Add one sentence to Task 5: "Task 5 runs before document segment B launches, or after its
  Tasks 5 to 8 invocation ends, never during it."

### PV7 (minor, question 3): two behaviors stated from memory

- **Where:** plan `:334-335` (`http.ResponseController` "then returns `http.ErrNotSupported`") and `:293`
  ("`unused` lints test files").
- **Defect:** Both are true. `responsecontroller.go`'s `Flush` falls to `errNotSupported()`, and `.golangci.yml`
  sets no `run.tests: false`. Neither is quoted, against the planning-miss rule. Every other new mechanism the
  fold added is quoted or cited to a line: pflag, the runner lines, `check.sh:566-618`, and `client.go:392`.
- **Proposed fold:** Add two rows to "Behaviors this plan relies on, quoted". One quotes `ResponseController.Flush`'s
  source switch. The other quotes golangci-lint's `run.tests` default, with a note that the repo's config does not
  override it.

### PV8 (minor, question 5): Reading 2 reverses the spec's literal default and surfaces only in HISTORY

- **Where:** plan `:5-6` ("carries no ruling for Geoff"), `:114-125`, `:636-637`.
- **Defect:** The spec's sentence ends "Otherwise they are pruned". No rung 9 spec exists (`ls
  docs/specs/`), so on the natural reading the condition fails and the symbols go. The fold keeps
  both. Its architectural case is sound. `ErrUnavailable` is the client's own wrapped sentinel, so "prune" cannot
  sensibly mean deleting it, and the limiter's rotation is three-class. The no-ceremony ruling covers taking the
  decision. Still, an approved spec's stated default is reversed, and Geoff would see that only in HISTORY.
- **Proposed fold:** No owner question. Add one line to the close's STATUS outcome naming the departure and its
  reason, so the next reader of the spec's sentence finds it.

### PV9 (minor): Task 3's coverage claim for `registry.go:289` is false

- **Where:** plan `:497-498`.
- **Defect:** The source grep pattern has `front ?exemption`, not bare `exemption`, so it cannot see
  `registry.go:289`'s "the one exemption". That file is Task 2a's, which deletes `isOptionalIntegrationKey`, so
  nothing is missed in practice.
- **Proposed fold:** Drop `registry.go:289` from the parenthesis, or cite Task 2a as the task that removes it.

## Question 2: contradictions, green-at-each-commit order, followable sequencing

- **Order builds green.** Task 1 removes `NewStrip`'s three parameters and edits every caller it lists: the
  `internal/serve` tests, `cmd/dubplate-server/serve.go:324`, `serve_test.go:1353`, and `web_test.go:439`, `:641`,
  `:772`. That matches `git grep 'NewStrip('`. Task 2a then deletes the shadow package, and Task 2b runs on
  orphans whose code stays. Task 3 follows 2a on `doctor_test.go`. Task 4b follows 1 and 3 on `seam.go` and
  `serve_test.go`. Task 4a is independent. No commit depends on a later one.
- **Followable by `pass-execute`.** Every conductor decision now falls before a segment launches: the `0003`
  re-confirmation, the A-Z fetch, and the `0006` re-cut. A segment has no between-task window, and this order
  respects that. The one under-stated slot is Task 5 (PV6).
- **Contradictions the fold introduced:** PV1 (Files against Outcome) and PV2 (a mutation the same commit makes
  impossible). The sibling-plan edit to "the merge" matches this plan's "Where it runs" bullet and Task 5's last
  paragraph.
- **One ambiguity left open:** the `0006` path triggers when the archive chore "stopped". The plan does not say
  what happens if the chore simply has not run when segment A ends. It reads as "wait". The document plan orders
  the chore "now, first", so this is unlikely to bite and is not ranked.

## Question 3: mechanisms or behaviors stated from memory

Only PV7. Each other mechanism the fold added cites a line that checks out against the tree or the runner.

## Question 4: over-engineering the fold added (571 to 684 lines), ranked by cost

1. **Task 2b's before-and-after leg confirmation** (`:419-421`). This is the costliest addition: it needs a pre-2a
   tree (PV3), and its only output is a HISTORY sentence. PV3's fold makes it one pre-flight command.
   Alternatively, drop "before": the table plus the after-run is enough to name any unlisted name.
2. **Close step 7 re-runs the key iteration** (`:623`) after Task 3 has already proved ruling 7. No task after Task
   3 touches `internal/config`. The re-run is cheap, but it repeats Task 3. Keep it only if the simplifier is
   expected to touch config.
3. **The "Shared heavy lock" header paragraph** (`:57-60`) restates `pass-gate-economy.md`'s rule. It costs four
   lines and no execution, so leave it.
4. **Task 0's three-way key classification and non-key inputs** (`:221-225`). The cost is small, and it settles
   which "operator supplies by hand" reading the spec meant, so it is not over-engineering.

Nothing here rises above minor. The fold's own measure is accurate: it proposed no new mechanism.

## Question 5: owner and plan boundaries

- **Correctly deferred to the owner:** Reading 6 keeps the ten architecture readers on ruling 6's text, and the
  three refusals are right. A major dependency bump waits for Geoff.
- **Correctly taken by the plan:** Reading 3 (`IntegrationClass.Optional`), Reading 4 (the runbook lines), the
  `auth-data` upgrade of Task 4b, and the Task 0 classification are method or placement calls.
- **Borderline:** Reading 2 (PV8). It is decidable under the no-ceremony ruling, but it reverses an approved
  spec's stated default, so it should surface in STATUS as well as in HISTORY.
- Nothing the plan decides belongs to Geoff, and nothing it asks Geoff is the plan's to decide.
