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
| F1 (fold) | `rendered_test.go` calls `newServeShell` (`:420-454`), whose signature Tasks 1, 4, 6, and 7 change, while the plan forbade editing that file | Folded: the global constraint now allows edits at those call sites only; each task's Files and contested surfaces name them. |

## Owed errata to ratified documents

The fold edits no ratified document. Task 9 applies these, with 6d-E1 from revision 1, to the rung 6
spec (`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`):

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
