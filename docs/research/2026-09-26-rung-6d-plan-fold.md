# Rung 6d plan fold: revision 1 to revision 2

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md`, revision 1 at `5afe46b`,
folded to revision 2. Reviews folded:
`2026-09-26-rung-6d-plan-review-contract.md` (C), `-mechanics.md` (X), and `-risk.md` (R).

## Summary

- **47 finding IDs** across the three lenses: C 10 major and 9 minor, X 6 and 9, R 3 and 10. After
  merging duplicates, **40 distinct findings**, plus **1 the fold found** while verifying (F1).
- **Folded: 39**, three of them with a part refused (C-M10, X-m9, R-m10). **Refused: 1** (R-m6).
  **Owner forks: 0.** Also folded: the risk lens's unnumbered planning-miss note (the limiter's
  request count) and F1. The fold made no ruling of a settled question.
- **Tasks: ten, unchanged.** No task or deliverable was added, so the ceiling stays **11.5M, flag
  9.2M**, and the pass stays at the spec's ten-task line without crossing it. The added rows sit
  inside existing deliverables; Task 0's re-derivation from 5d's actual is where a heavier rate
  shows.
- **Rulings: the three from revision 1, unchanged.** The risk lens agreed with Rulings 2 and 3 on
  the evidence, and no finding moved Ruling 1's basis.

Every finding was checked against the tree at `96752b8` (the reviewed code is unchanged since
`5afe46b`) or against the reviewer's recorded probe. Two mechanisms the fold adds were proven here:
`systemctl try-restart` and the `systemd-run` exit status were quoted from the installed man pages,
and the trigger fault fixture was probed through Python's `sqlite3` (a `BEFORE INSERT` trigger raising
on the third insert inside `BEGIN IMMEDIATE` left both tables empty after `ROLLBACK`).

## Why the fold refused so little

Most findings name a state the plan left open or a proof that passes vacuously, which is the class of
defect the repo's planning-miss rules target and the class every 5b-ii fix round came from. Each such
fold is a table row or one clause, far cheaper than the fix round it prevents. The fold did decline
the parts that would buy a test-only seam, a brittle shell-order grep, or code for a state that
cannot exist.

## Conflicting folds, resolved at the root

- **The keep read (X-M1 against R-m2).** X-M1 proposed refusing any out-of-bounds stored keep row
  before migrating. R-m2 proposed taking the post-seed value. Neither alone is right. Revision 2
  follows 6c's R-6c7 and 6c-E8 (the bound applies to the value in force): an in-bounds row passes its
  own value (spec `:411-414`), an out-of-bounds row refuses only when the outcome refuses, and a healed
  row passes its healed value. `evaluate.go`'s `outcome` doc was read to confirm the heal rule.
- **Listener proofs (C-M1).** The fold took the finding and declined its mechanism: an injected
  listen function is a test-only seam, which `go-architecture-reader` grades against. The
  preparation function (`newServeShell`) is already the unit before `serve.Run`, and `serve.Run` is
  the only listener, so its return is the observable moment (R-6d2).

## Dispositions

| IDs | Finding | Disposition |
| --- | --- | --- |
| C-M1 | No criterion observes the listener | Folded: R-6d2 names the preparation function as the proof point; Task 1 and Task 2 criteria and mutations rewritten to it. The injected listen seam is declined (test-only seam). |
| C-M2 | Lock wait proven only for an exclusive holder | Folded: Task 1's two shared-holder rows and the unconditional-exclusive mutation. |
| C-M3, X-M5 | Prune proof vacuous; the second run cannot prune | Folded: Task 1's two separate backup-off `0004` fixtures, the "skip the prune" mutation, and the `migrate.go:121-123` reason. |
| C-M4 | Reactivation can re-grant admin on the cookie channel | Folded: R-6d10 item 2, the privileged-fields `users.admin` row, a Task 3 row, and a mutation. |
| C-M5, R-M1, X-m7(b) | `AddMember`'s actor admin check and at-least-one-admin rule dropped | Folded: R-6d10 items 1 and 4 (inside `Update`, quoted), Task 3 rows and mutations, and two review-focus lines. The "read through `QueryRow`" mutation became a diff-reviewer check by name, because no deterministic test catches that race. |
| C-M6, R-m5, X-m7(a) | Tombstoned rows hold both `UNIQUE` columns | Folded: R-6d10 item 3 (the username read covers tombstoned rows; the reactivated row's own name is no collision), the schema quoted, Task 3 rows and an API test, and Task 5's tombstoned-address refusal (R-6d11). |
| C-M7 | Criterion 1's admin proof passes when the flag is dropped | Folded: Task 5 criterion 1 asserts per-row flags, the record's hash, count, and actor, and the cookie-channel authority; 6d-R12 records the spec clause's reading. |
| C-M8 | Two Task 5 mutations have no failing fixture | Folded: the valid dry-run criterion and the trigger fault fixture (probed). |
| C-M9, X-m6, R-m1 | Import admits usernames the write path refuses; no field-rule state; "naming both entries" | Folded: R-6d11 runs each entry through `validateWriteCandidate`, names positions under `validateRoster`'s quoted rule, and Task 4 keeps the parse helpers; Task 5 row and mutation. |
| C-M10 | `deploy.sh` promises have no proof | Folded in part: the state-dir leg is extended to gate the unprefixed grant and the leaves array; `systemd-run`'s exit status is quoted. **Refused in part:** a gate grep for the step order inside `deploy.sh` is a brittle pseudo-behavior test the repo's shell rule excludes, and `deploy.sh --check` needs ssh to the box; the order goes to the diff-reviewer by name. |
| C-m1 | Reconcile-root mutation needs differing roots | Folded: Task 1's reconcile fixture state. |
| C-m2 | `doctor` string proof lacks the pre-start render; `backup.go:274` untested | Folded: Task 2's before-and-after assertion, for both. |
| C-m3 | Out-of-bounds row and the admin conflict's resolving action | Folded: Task 2 rows (under 6c-E8's reading, and naming `outcome`'s existing "unset KEY, or set it to V"). |
| C-m4 | `Valid`'s zero-id refusal unproven | Folded: Task 3 row and mutation. |
| C-m5 | Task 4 fixtures, order, pre-`0003` store, foreign uid, the owner check unquoted | Folded: Task 4 rows, `List`'s order in R-6d10, the foreign-uid criterion, and `leaf.create`/`checkOwner` in the code pointers. |
| C-m6 | Dry-run aggregation across layers; mode-0000 under root; display name | Folded: Task 5 criteria and R-6d11's display-name line. |
| C-m7 | Which reader fails, and the sweep's report | Folded: Task 7's per-mode failing readers and R-6d9's sweep rule. |
| C-m8 | Unquoted behaviors | Folded: the behaviors table gains `QueryRow`, the owner check, `systemd-run`, and `validateRoster`. |
| C-m9 | Criterion 4 tests four verbs, not every verb | Folded: one table test over the registered command tree. |
| X-M1, R-m2 | The keep extraction gives the wrong number; out-of-bounds rows pass or strand | Folded: R-6d4 rewritten (see "Conflicting folds"); Task 1 rows and a mutation. |
| X-M2 | `parse` has no `KindString` arm, so every restart after a string seed refuses | Folded: R-6d7, Task 6 Files, row, and mutation. |
| X-M3 | Criterion 9's "then B at a new start" contradicts the seed rule | Folded: Task 7 unsets the env at the restart and adds the refusal sub-case; 6d-R9. Owed spec erratum E-a below. |
| X-M4 | The raised-cap direction is vacuous without a refresh | Folded: Task 7 defines it through a refresh, with `clampExpiry` quoted. |
| X-M6 | The Task 6 to Task 7 seam is not independently green | Folded: Task 6's interim-read outcome and Files. |
| X-m1 | The re-entry warning would fire on every start | Folded: R-6d16 warns on a second or later settle only; Task 1 criterion. |
| X-m2 | `check`'s holder line would list its own pid | Folded: Task 8 outcome, states, and mutation; `/proc/locks` quoted from the probe. |
| X-m3 | `Valid`'s change breaks two `internal/ops` fixtures | Folded: Task 3 Files and contested surfaces. |
| X-m4, R-m4(b) | Members facts need version guards | Folded: R-6d13 and Task 4 rows and mutation. |
| X-m5 | The owner refusal recommends `runuser` | Folded: R-6d15 and Task 8 Files (`leaf.go` and the four tests asserting the hint). |
| X-m8 | Serve's report needs restore's head | Folded: R-6d16 passes the stamp and the `RestoreIntent`; Task 1's reconcile row asserts the head. |
| X-m9 | Contexts for writes and nonce readers | Folded in part: `Insert`, `Sweep`, and `RunSweep`'s function gain `ctx` (R-6d9, Task 7). **Refused in part:** the write deadline is already set by both callers (`admin.go:255`, `member.go:95`, both verified), and `Lookup` reads the pool, which needs none. |
| R-M2 | The import drops the writable-by-others refusal | Folded: R-6d11 (quoted), the privileged-fields row, Task 5 row and mutation. |
| R-M3 | "Restart `serve`" would start the unit 5d keeps stopped; `check` order | Folded: new R-6d19 (`try-restart`, quoted; render, install, `check`, halt, `try-restart`), Task 8, Task 0 item 2, 6d-R10. |
| R-m3 | The runbook deletes the only rollback input | Folded: R-6d18 and Task 8 keep the file until 5d's `.prev` copies are superseded. |
| R-m4(a) | "Leaf present and empty" state missing | Folded: Task 1 row. |
| R-m6 | The settle loop has no overall bound | **Refused.** A bound's proof needs a test-only stub for `settle`, and revision 2's second-settle warning (X-m1) already makes a spin loud on every pass rather than silent; the cost exceeds the risk of a hypothetical `settle` bug. |
| R-m7 | Read ordering in the link handlers | Folded: R-6d9's orderings, Task 7 criteria and two mutations, and a review-focus line. |
| R-m8 | `linkExpiryCeiling` must stay the registered ceiling | Folded: R-6d9 (quoted), Task 7 outcome and mutation. |
| R-m9 | No proof that `check` fails on an auth refusal | Folded: Task 6 row and mutation. |
| R-m10 | Import-record concurrency; the pre-6d restore report line | Folded in part: R-6d11 places the record read inside the callback, and Task 5 adds the concurrent-imports test (a schema singleton was not needed). **Refused in part:** the spec's "the report says to run `member import`" (`:715-716`) is not built, because no pre-6d store or snapshot can exist (`OpenWrite` refuses an absent store and 5d never starts `serve`); recorded as 6d-R11, owed erratum E-c. |
| R planning-miss item 4 | The limiter proof needs a request count | Folded: Task 7's fixture state names a count that trips A and not B. |
| F1 (fold) | `rendered_test.go` calls `newServeShell` (`:420-454`), whose signature Tasks 1, 4, 6, and 7 change, while the plan forbade editing that file | Folded: the global constraint now allows edits at those call sites only; each task's Files and contested surfaces name them. **Corrected in the second fold (B1): the premise was a misreading.** `:420-454` are doc comments naming `newServeShell`; the calls go through `serve_test.go`'s `buildServeShellWithLogger`. The file's real 6d exposure is its roster-file harness. |

## Owed errata to ratified documents

The fold edits no ratified document. Task 9 applies these, with 6d-E1 from revision 1, to the rung 6
spec (`docs/specs/2026-09-23-rung-6-ops-seam-store.md`):

- **E-a (6d-R9), criterion 9.** "With the env holding A and an in-package admin row holding B ... a
  limiter built at the next start uses B" cannot hold under the spec's own seed rule (`:759`): env A
  beside admin row B refuses the start. The next start runs with the env unset or set to B.
- **E-b (6d-R10), criterion 14.** "Names the holder before restarting `serve`" reads as `check` naming
  the holder, then `systemctl try-restart`, which is a no-op until rung 16 enables the unit.
- **E-c (6d-R11), `:715-716`.** The pre-6d-snapshot restore report line is not built: no pre-6d store
  or snapshot can exist.
- **E-d (6d-R12), criterion 1.** "An imported admin can grant admin through the CLI" proves nothing
  about the import, since the CLI grant checks no member's flag; 6d proves the imported flags and the
  imported admin's cookie-channel authority instead.
- **E-e (6d-R4, widened), criterion 2 and the import table (`:820-831`).** Four refusal states join
  the spec's table (an empty file, a file writable by others, an entry failing a field rule, a
  tombstoned address), and "naming both entries" names positions.

## Findings the fold could not verify

- X-M4's session probe and X-m2's `/proc/locks` probe were not re-run; their evidence is kept from
  the mechanics review, and the fold read `session.go`'s `clampExpiry` and `VerifySession` to confirm
  the first.
- `AddCorrelatedRule`'s refusal of an unregistered key (X-M6's fold) is taken from the mechanics
  review, not re-read.
- The trigger fault fixture is probed on Python's `sqlite3`, not `modernc.org/sqlite`; the plan's
  "Not verified" list carries it.

## Second fold: revision 2 to revision 2.1

Target: the plan at revision 2 (`adb9644`), folded to revision 2.1 against the verification read
(`2026-09-26-rung-6d-plan-fold-verification.md`, committed at `314c0c8`): 1 blocker, 3 majors, and 4
minors. Every finding was re-checked against committed `HEAD` (`314c0c8`) before folding;
`rendered_test.go` moved a few lines since `adb9644` (5c-iv Task R2), so the plan cites its
`314c0c8` lines. **Folded: 8. Refused: 0. Owner forks: 0.** No task was added, so the ten-task line
and the 11.5M ceiling (flag 9.2M) stand, and the three rulings are unchanged: every fold adds or
moves files inside existing tasks, and none touches a ruling's premise.

Evidence gathered for the fold, each quoted or probed, never recalled:
- `rendered_test.go` at `314c0c8`: `startRenderedServer` and `startFixtureServer` call
  `buildServeShellWithLogger` (`:428`, `:452`); `renderedEnviron` (`:372-399`) writes a roster with no
  `streamer_username` and sets `DUBPLATE_ROSTER_FILE` (`:383`); `rosterResetServer` (`:613-659`)
  rewrites that file; `members/roster-file` (`:918-927`) reads it; `mintLinkFor` (`:535`) calls
  `resolveAuthConfig`, which Tasks 6 and 7 replace and delete. `serve_test.go:66-89`'s builder calls
  `resolveAuthConfig` and `newServeShell`.
- `run.ts:4569-4574`: `resetCaptureRoster` POSTs and reads `{ members: number }`; `:5061` compares
  each capture state against it.
- `flock(2)` (installed man page): "If a process uses open(2) (or similar) to obtain more than one
  file descriptor for the same file, these file descriptors are treated independently by flock()."
  `tryLock` opens a fresh descriptor per attempt (`lock.go:85-90`).
- Roster callers: `grep` finds `Lookup` at `handlers.go:210`, `:366` and `seam.go:114`, and `List` at
  `handlers.go:308`, `admin.go:206`, `member.go:162`, `doctor.go:173`. `Identity.Valid`
  (`identity.go:208-211`) checks no id today; `RequireAdmin` refuses on `!v.Identity.Valid()`
  (`admin.go:92`, `:121`).
- `write.go:68-69`: `Update` returns `errNoDeadline` when `ctx.Deadline()` is unset; `apply`
  (`apply.go:29-31`) passes its caller's context straight to `Update`; `memberVerbWriteWait` is 30s
  (`member.go:26`, used at `:95`).
- `open.go:656`: an absent store refuses "no store at %s: start serve first".
- 5d plan: Task 6's rendered-doctor test lives in `templatekeys_test.go` (`:1055-1097`), with the
  roster-row and `DUBPLATE_ROSTER_FILE`-row assertions the verification names; step 14 (`:1377-1380`)
  is the supervised start; `--apply` swaps the binary before installing `server.env` (`:1256-1257`).
  The repo's template at `HEAD` has no `DUBPLATE_ROSTER_FILE` row yet (only a comment at `:72`); 5d
  adds it.
- `grep -n DUBPLATE_ROSTER_FILE cmd/dubplate-server/*_test.go` finds the key in seven test files'
  environments, not only `rendered_test.go`.

| ID | Finding | Disposition |
| --- | --- | --- |
| B1 | `rendered_test.go`'s harness is built on the roster file; the global constraint forbade the rework; F1's premise was wrong | **Folded.** Confirmed at `314c0c8`. The global constraint now reads "changes only as the store move and the settings move require". Task 4's Files take the rework: the two fixture members seeded into the store with streamer usernames; `rosterResetServer` resetting `users` (the first POST's live ids are the baseline, later POSTs delete rows outside it) through its own write-mode handle or the shell's, keeping `{"members": n}` so `run.ts` and R-6d17 hold; `members/roster-file` reading `users`. Two proofs and a mutation added. The code pointer and Task 1's Files are corrected (Task 1 touches `serve_test.go`'s builders, not `rendered_test.go`); the F1 row above is corrected. Found while verifying and folded with it: `mintLinkFor` calls `resolveAuthConfig`, so Tasks 6 and 7 name it; `run.ts:3938`'s stale comment joins R-6d17's routed note. The in-process second handle is quoted from `flock(2)`, not probed, and the plan's "Not verified" list says so. |
| M1 | Task 3 is not green alone: `ctx` breaks six callers outside its Files, and `Valid`'s zero-id refusal breaks `RequireAdmin` until Task 4 | **Folded, by moving the `Valid` change.** Task 3's Files gain the `ctx` threading at the six call sites the verification names (confirmed by `grep`). `Valid`'s zero-id refusal, its state row, its mutation, and the two `internal/ops` fixtures move to Task 4, where `serve` first reads `users`. Moving `serve`'s roster construction into Task 3 was weighed and declined: it would drag the rendered harness rework and the member verbs into Task 3, which is the identity surface's `opus` task. Neither task's deliverable count changes. |
| M2 | The template keeps `DUBPLATE_ROSTER_FILE` from Task 5 to Task 8; 5d's rendered-doctor test is unnamed | **Folded.** R-6d12 now drops the template row in Task 5's commit, with the header comment. Task 0 item 2 names the test. Task 4's Files take its members-row assertions (an absent store reads "no store yet", exit unaffected). Task 5's Files take its key assertions: the fixture roster and the deleted-row assertion retire, and a re-added row fails naming the supersession message, which is Task 5's new mutation. Criterion 14's template half moves from Task 8 to Task 5. Widened on the same evidence: every test environment that sets the key fails the wall load from Task 5, so Task 5's Files name the seven test files' helpers. |
| M3 | The 6d upgrade cannot run: no store exists, `try-restart` is a no-op, `member import` refuses an absent store; Owner touchpoints said "restart `serve`" | **Folded.** R-6d18 adds one supervised start after the deploy, in 5d step 14's form: `systemctl start`, then `doctor` (the store row at the binary's version; the members row failing naming `member import`, which is expected here), then `systemctl stop`, with the unit left disabled. The import steps follow the stop, through `systemd-run` with the store leaf in `ReadWritePaths`. Task 8's runbook outcome and render-and-read check this order. Owner touchpoints no longer says "restart", and agrees with R-6d19. |
| m1 | R-M3's `try-restart`-to-`restart` mutation dropped without a recorded refusal | **Folded as a one-grep check.** The state-dir leg fails when a `systemctl` line in `deploy.sh` runs `restart`, `start`, or `enable` (whole words, so `try-restart` passes) on the serve unit, or when the enabled timer list names it. R-M3's mutation is restored against it. The step-order grep stays refused (C-M10). |
| m2 | The seed apply and `member import` are new `Update` callers with no stated deadline | **Folded.** R-6d5 names a 30s write wait on `memberVerbWriteWait`'s precedent for the seed apply, and `member import` uses `memberVerbWriteWait` itself. `Update`'s behaviors row now quotes what the deadline bounds. Task 2 adds the outcome and a mutation (dropping the deadline fails every writing row with `errNoDeadline`). |
| m3 | R-6d19's "render the env, install the binary" reverses 5d's install order | **Folded.** R-6d19 and Task 8 now state the rule that matters: `check` runs after both the binary swap and the `server.env` install, in 5d's existing order. |
| m4 | The fault fixture catches only one variant of "the record outside the transaction" | **Folded.** Task 5 names the variant the fixture catches (the record in its own `Update` before the inserts) and rewrites the mutation to it. The after-the-inserts variant, which the fixture cannot see, goes to the diff-reviewer by name. |

### What the second fold could not verify

- The reset server's second handle in the serve shell's process is quoted, not probed (see B1).
- 5d's rendered-doctor test does not exist yet (5d has not executed), so its Task 4 and Task 5 edits
  are specified against the 5d plan's text; Task 0 item 2 reports the landed test.
- Whether 5d's `deploy.sh` holds the serve unit's name in a variable, which m1's presence check
  must match, waits on 5d's landed file; Task 0 item 2 reports it.

## Prose fold: revision 2.1 to revision 2.2

Source: `2026-09-26-rung-6d-plan-prose-review.md` (one `prose-voice-reviewer`, Opus, over the plan
at `985ff3a`). The fold took all 3 blockers, all 11 warnings, and all 6 suggestions, and refused
none. It changes no fact, ruling, task, deliverable, or criterion; the one number that moved is
W4's variant arithmetic, which the review asked to recompute. Every citation fix was checked against
the spec, CH, or the code before it was written. Between `20f4a7c` and `985ff3a` the specs,
`internal/`, `scripts/`, and every `cmd/` file but `rendered_test.go` (cited at `314c0c8`) are
unchanged, so those line numbers hold at either base.

| ID | Finding | Disposition |
| --- | --- | --- |
| B1 | Ruling 3's yes/no inverted in the header, Ruling 1, and the variants | **Folded.** All three now read "a 'no' on Ruling 2 or on Ruling 3", matching Ruling 3, 6d-R8, and Task 9. |
| B2 | 6d-R11's reason said 5d never starts `serve` | **Folded** with the review's text: 5d's one supervised start runs the pre-6d binary, which never opens a store, and `OpenWrite` refuses an absent one. Confirmed against 5d plan step 14 (`:1377`). |
| B3 | 6d-R2 and 6d-R11 cite wrong spec lines | **Folded.** 6d-R2 now cites `:839-840` and 6d-R11 `:713-714`; both quotes confirmed at those lines. |
| W1 | Four more wrong citations | **Folded, one corrected.** CH `:83-84`, spec `:843-846`, and spec `:1128-1130` beside `:1141` confirmed as given. The `flock(2)` quote starts at spec `:374` ("Converting a lock") and reaches "not guaranteed to be atomic" at `:375`, so the row cites `:374-375`, not the review's `:375-376`. |
| W2 | Citation bases mixed | **Folded, premise corrected.** The members-page lines the plan cited (`:216-224`, `:66-68`) were right at `20f4a7c`; R2 at `92147bc` moved them, and they are now re-cited at `985ff3a` (`:312-318`, `:95-97`). The `run.ts` lines were already at `985ff3a` (R2 moved them too). The Sources line and the code-pointers heading name each base, `check.sh`'s "at HEAD" became the default base, and "Not verified" gains the `web/` line. |
| W3 | Task 4's fixture-state claim paired the wrong mutations | **Folded** with the review's text; "longer than it" became "longer than that count". |
| W4 | Ceiling variants omit the split cost; "this table" is a list | **Folded.** Each variant adds about 0.5M for the split: one task projects 10.29M, ceiling 13.0M, flag 10.4M; both project 11.17M, ceiling 14.0M, flag 11.2M, totals across both halves. Ruling 1 now points to "the 'Token ceiling' arithmetic". |
| W5 | Tasks 6 and 7 overlap on `mintLinkFor` | **Folded.** Task 6 moves `mintLinkFor` off `resolveAuthConfig`; Task 7 confirms no caller remains; the global constraint says the same. |
| W6 | Owner touchpoints contradicted itself | **Folded** as three numbered steps and the review's closing sentence. |
| W7 | Ruling labels collide; review labels undefined | **Folded.** "Spec ruling 9", "the rung 5 review's minor finding 14", "the spec's erratum row 10", and "the 6c plan's Ruling 2" at each site (four, one in Task 0 split across two sentences). Task 5's "(ruling 8)" became "(spec ruling 8)" on the same rule. A Sources line defines B1, M1, M2, C-M10, R-M3, D10, S20, and R5S, and says a bare "Ruling N" is this plan's. |
| W8 | Rulings 2 and 3 hard to scan | **Folded.** Each opens with a one-sentence definition (the dual-key epoch from the spec's "The epoch"; the store identity from 6b-i's R9 option b), gives its reasons as a short list, and keeps its recommendation and both "what each answer builds" paragraphs. |
| W9 | Two procedures are run-on prose | **Folded.** R-6d19's order is four numbered steps; Task 8's runbook outcome is a four-step install order and a five-step upgrade section. |
| W10 | The execution model contradicted Task 8's header | **Folded** with the review's text. |
| W11 | Stale parenthetical in Task 8's Files | **Folded** with the review's text. |
| S1 | `store.go:280-284` should be `:286-298` | **Folded as `:285-298`.** `settingsChangeNote`'s doc comment begins at `:285` and already says "serve's next start". Also fixed the same citation in Task 2's Files. |
| S2 | The demote grep "finds only comments" | **Folded.** It finds comments and tests (`admin_test.go`, `roster_test.go`). |
| S3 | "will rewrite most of the files this plan names first" | **Folded.** |
| S4 | "planning recorded against the rung and outside it" | **Folded.** |
| S5 | Split the status line | **Folded.** The status line carries the revision and state; the history is a list. |
| S6 | Task 3's Files omit `identity.go`'s `Principal` doc comment | **Folded.** Confirmed at `identity.go:190-193` ("so rung 6 can substitute a durable surrogate id here"). |

Refused: none.
