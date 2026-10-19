# Rung 14a plan review, lens P-A (contract and criteria)

Target: `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md` at `fae3de3` ("the plan"). Spec: revision 2.1
("the spec"). Fold record: `research/2026-10-10-rung-14a-spec-fold.md` ("the fold"). Read beside them:
`pass-core/SKILL.md`, the repo CLAUDE.md, `scripts/check.sh:163-275`, `cmd/dubplate-server/rendered_test.go`,
`web/scripts/rendered/run.ts`, `web/package.json`, `internal/auth/session.go:70-141`, `internal/auth/seam.go:83-115`,
`internal/auth/admin.go:105-135`, `cmd/dubplate-server/auth.go:120-325`, `internal/config/thresholds.go:40-50`, and
PS:126-151, :575-585, :615-625.

## Coverage map

Every spec AC and owed item has an owning task. The gaps are in proofs and fixtures, not in assignment.

| Item | Owner(s) in the plan | Verdict |
|---|---|---|
| AC1-AC4 | Task 7 | covered |
| AC5 | Tasks 3, 7 | covered; 401 fixtures thin (PA-9) |
| AC6 | Tasks 1, 6 | the log-line clause has no owner (PA-10) |
| AC7 | Tasks 3, 7 (V4) | the queue-field clause has no test (PA-15) |
| AC8 | Task 2 | covered; the zero-mark fixture is not per family (PA-4) |
| AC9, AC10, AC11 | Tasks 5, 6, 7 | covered |
| AC12 | Task 6 | the refresh clause is vacuous at the seam (PA-1); the window's own fixtures are missing (PA-12) |
| AC13 | Task 6 | the command-line and seed-rule clauses have no proof (PA-11) |
| AC14 | Task 8 | covered; one mutation has no fixture that catches it (PA-13) |
| AC15 | Tasks 4, 8 | the injected bump failure fixture is missing (PA-7) |
| AC16 | Tasks 4, 6, 8 | the concurrency proof's mutation needs an ordering point (PA-8) |
| AC17 | Tasks 6, 7, 10 | covered |
| AC18 | Tasks 2, 4-9 | covered by reference; HTTP fixtures thin (PA-9) |
| AC19, AC20 | Task 10 | proofs name a test runner that does not exist, and the existing harness rows break (PA-2) |
| AC21 | Task 9 | the "no code path grants" clause has no proof (PA-11) |
| AC22 | Tasks 5, 8, 10 | #25 and #12 carry no proof (PA-14) |
| V4, V6, V7 | Tasks 7, 4, 4 + 11 | covered |
| V5 | Task 6, P8 | no fixture or mutation (PA-12) |
| doctor warning split | P9, Task 9 | covered, with a reason that meets the fold's test |
| errata 1-8 | Task 11 step 9 | all eight placed, each at its cited line |
| conductor-owed | lane launch 2, Task 11 step 9 | the `internal/serve` ruling question is not carried (PA-16) |
| PS A1 (member cap) | Tasks 4, 9, 10 | covered; one row is unreachable (PA-13) |
| PS A2 (route inventory) | Task 7 | covered |

## Findings

### PA-1. major. Plan:440, :427, :468-475. The re-authentication proof can pass vacuously, and its real channel has no mutation.

**Defect.** The freshness of a write is a privileged field. The spec sets it through one channel only: "the mount's
session middleware, from the verified session's anchor, carried on the actor" (spec:480). `VerifyIdentity` returns a
`Session` carrying both `IssuedAt` ("A refresh moves it") and `AnchoredAt` (`internal/auth/session.go:76-86`). So the
one place that can pick the wrong time is Task 7's verifier, built in `api.go` from P2's session check, where it
calls `ops.NetworkActor(id, anchor)`. Task 6 owns AC12, and its "refreshed after its anchor aged out" fixture and its
"gate on `IssuedAt` instead of the anchor" mutation sit at the seam. There the test builds the actor itself, and
`ops.Actor` exposes only the anchor (Task 1). The mutation is unreachable at the seam, and the refresh fixture proves
nothing about a refresh. Task 7's mutations do not include passing `IssuedAt` as the anchor. A verifier that did so
would pass every listed test, and staying signed in would renew freshness, which the spec rules out (spec:260-261).

**Fold.** Move AC12's refresh clause to Task 7. Fixture: a session whose anchor is outside the window and whose
`IssuedAt` is inside it after one same-origin refresh, driven through the real mount against a sensitive
`set_setting`. Expected answer: `reauthentication_required`. Add the Task 7 mutation "build the actor from
`Session.IssuedAt`", which the fixture must catch. Drop the `IssuedAt` mutation from Task 6, or restate it as "the
seam gates on a time other than `Actor`'s anchor accessor".

### PA-2. major. Plan:524-547; spec:529. Retiring `roster` breaks existing rendered rows that Task 10 does not list, and the proofs name a test kind the repo lacks.

**Defect.** The rendered harness is built on the roster API today. These rows and helpers drive or intercept `**/roster`:
`members/unreachable` and `members/malformed-roster/*` (`web/scripts/rendered/run.ts:3098`, `:3151`), the
console-noise regex naming `roster` (`run.ts:866`), the non-admin row's "intentional 404 from `roster`" (`run.ts:861`,
`:1050-1090`), the `rosterResetServer` (`rendered_test.go:799`), and the proof that "the POST the script drove was the
real roster API" (`rendered_test.go:860`). After Task 10 these rows fail the rendered leg or prove nothing, so Task
10's gate cannot go green on the listed files. The plan lists only "the rendered harness's rows for the new document",
and the territory (plan:104-105; spec:529) gives the harness's shape to rung 7 with 14a "appending rows only". Editing
these rows is a write to a rung 7 surface that the contested-surfaces table does not grant.

Also, Task 10's proofs say "fails a component or rendered test" (plan:546). `web/package.json` has no component test
runner (no vitest), so only the rendered leg can carry them.

**Fold.** Add to Task 10's files the members rows and helpers in `run.ts` and `rendered_test.go` named above, retargeted
to `/api/v1`, and amend the contested-surfaces row in the lane registry (lane launch step 2): "14a owns the harness's
`members/*` rows and their reset server; rung 7 owns the harness's shape." Restate each Task 10 proof as a named
rendered manifest row: the JSON 403, each refusal's copy (last admin, reauthentication, forbidden, signed out, member
cap), the add-refusal focus moves (AC19's second clause, which has no proof today), the confirm step, the
environment-set row, and the apply-mode line (AC20's clause, which has no proof today). Each row says what it
intercepts and what it reaches for real.

### PA-3. major. Plan:557-595. Close ordering: a Go fold lands after the merge ritual's gates.

**Defect.** Step 8 dispatches the architecture reads and "one fold task" after step 7 has run `check.sh --merge`,
`--ff-only`, and `check.sh` on master. A Go fold from step 8 then lands on master, or on a branch already merged,
with no full gate and no `diff-reviewer`. Step 4's UI reviewers and step 1's segment C security read can also yield
folds after step 3's full gate. Those folds are gated only if they land before step 7. The plan does not say so.
Holding the reads until after the merge also costs clock: a fold after the merge needs a second merge ritual, which is
two more full gates.

**Fold.** Reorder the close so every code-changing fold precedes one gate:
1. The segment C security read and its fold.
2. The architecture reads and their fold. They need no plan context, so they can run in parallel with the UI
   reviewers.
3. The UI reviewers and their fold.
4. The simplifier.
5. The full gate in the worktree (the close gate).
6. The live smoke and the render-and-read.
7. The merge ritual.

Only the docs fold follows the merge. If CLAUDE.md's "at each pass's merge" wording seems to require reads after the
merge, read it as "before the merge's gate".

Over-ceremony in the same step: `internal/identity` changes only a doc comment (Task 1). Drop its architecture read
and say why in the plan, which saves one Opus read. The ceiling's "two more architecture reads" already prices
`internal/engine/store`, which Task 9 touches for real.

### PA-4. major. Plan:306-324. Hook 18's mark conversion has no proof that it keeps each gated key's sensitivity.

**Defect.** Task 2 replaces `Setting.Sensitive bool` with the closed mark type. The `tool` class carries no mutation
mandate, and its only proofs are about registration panics. For store-homed settings the mark decides Task 6's
re-authentication gate and the notice. Three auth settings are `Sensitive: true` today
(`cmd/dubplate-server/auth.go:134`, `:144`, `:154`). The two new keys (P6, P7) are declared sensitive. If the
conversion maps one store-homed key to not-sensitive, a borrowed cookie can change it with no fresh sign-in and no
notice, and no listed test fails. Separately, the fixture names "one with a zero mark" and not one per family. A
family whose `Add*` lacks the zero-mark panic then passes. AC8 asks for a panic for a zero mark in every family.

**Fold.** Add to Task 2 one table test that pins every store-homed setting's mark to its pre-14a `Sensitive` value
(magic link expiry, session lifetime, and absolute cap are sensitive, and the rest are not). Name the mutation "flip
one store-homed key's mark". Tasks 4 and 6 extend the table with `DUBPLATE_MEMBER_CAP` and `DUBPLATE_REAUTH_WINDOW`,
both sensitive. Make the zero-mark fixture per family, as the description fixture already is. The `tool` class can
stay, since the change is mechanical apart from this one pin.

### PA-5. minor. Plan:143-148, :236. P5's "keeps the admin's pending change" has no mechanism and no proof.

**Defect.** Review focus 1 claims "the change survives". The interstitial offers no onward link
(`login/link/+page.svelte:15-17`). If the "sign in" link navigates the page's own tab to `/login`, the pending change
is lost with the component. The plan names neither a new tab nor any persistence, and no Task 10 proof covers the
survival.

**Fold.** P5 states the link opens the sign-in page in a new tab (`target="_blank" rel="noopener"`). The copy keeps
"sign in from this browser". One rendered row proves that after a `reauthentication_required` answer, the field
still holds the edit and Retry re-sends it.

### PA-6. minor. Plan:361-363, :375-378. The roster's cap read is missing from Task 4's interface.

**Defect.** Task 4 says "the roster constructor takes the epoch store and a clock" and "consumes nothing new". The
add path must still read `DUBPLATE_MEMBER_CAP`, a store-homed setting, and refuse when it cannot read it. The auth
package reads settings through `Setting[T]` readers (`seam.go:83`), and `member.go:97` and `:133` build the roster
without one. Unless the plan names how the cap reaches the roster on both channels, the implementer picks the shape.
That choice decides whether `member list` needs a reader too, and whether a read failure on the command line fails
closed (PS:138, "a channel that can't reach the store fails closed").

**Fold.** Task 4's interface states that the constructor also takes the cap's `Setting[int]` reader, that `serve` and
the member verbs build it from the same registry, and that a read error refuses the add with `internal`.

### PA-7. minor. Plan:383-389. Task 4 has no fixture for "bump after the commit", and no mutation for a commit on a non-backward bump error.

**Defect.** In a sequential test, "bump after the commit" ends in the same state as the correct order. Only AC15's
injected bump failure separates them, and Task 4's fixture list omits it. The states row "any other bump error: rolls
back" has no mutation of its own either.

**Fold.** Add two fixtures: an epoch store whose `Bump` returns a non-backward error, after which the member is still
live, and one that returns `ErrEpochBackward`, after which the removal commits. Add the mutation "commit despite a
non-backward bump error".

### PA-8. minor. Plan:384-387; spec AC16. The concurrency proof's mutation needs a named ordering point.

**Defect.** AC16's proof must fail against a guard moved outside the transaction. That mutant reads the admin count
before `Update`, and so sees two admins only if its read runs before the held transaction commits. The fold refused a
barrier hook. The plan names no other signal that the second demotion has read and is waiting. A test that commits on
a timer may commit first, and the mutant then passes.

**Fold.** The fixture names its ordering point. Two forms work. If the held transaction runs through the same `Handle`,
the test polls the write pool's `sql.DBStats.WaitCount` (quote `go doc database/sql DBStats`) until the second
`Update` is waiting. Otherwise, the implementer reports the mutant failing in ten of ten runs. Either way the
implementer's mutation report states which form it used.

### PA-9. minor. Plan:468-471. Task 7's fixtures do not reach the 401 classes, the refresh row, or self-removal.

**Defect.** P3's split and the plan's added states rows need more fixtures than "a live member session, a live admin
session, a session whose cap setting is made unreadable". Without a refusal-class fixture, a mutant that answers a
removed member's session with 500 passes. Without a refresh fixture, the row "a valid session, same-origin: admitted,
and the refreshed cookie set" goes unproven. Review focus 2's HTTP half also has no fixture.

**Fold.** Add four fixtures: no cookie, which answers 401; a removed member's session, 401; a same-origin request past
the refresh point, which sets the refreshed cookie; and an admin removing themself while another admin remains, whose
next call answers 401. Add the mutation "answer a gate refusal with 500".

### PA-10. minor. Plan:296, :435, :455. AC6's third clause has no owner.

**Defect.** "The log line carries the principal as its own field" (AC6; spec:183) sits in no task's acceptance. Task 7's
log record lists only "actor, channel, and `client_ip`". The actor is now the hex form, so a reader of the log loses
the principal.

**Fold.** Task 7's log record adds `principal`, and its acceptance cites AC6's log clause, with a test that reads the
record.

### PA-11. minor. Plan:435, :495, :518. AC13's command-line and seed-rule clauses, and AC21's first clause, have no proof.

**Defect.** "A command-line reset mails no one", "a seed-rule change mails no one", and "no code path reads
`DUBPLATE_ADMIN_EMAIL` to grant or authorize admin" are criteria, and no task names their fixtures.

**Fold.** For AC13, Task 8's acceptance states that `settings reset` builds no notifier, with a test that wires a
recording mailer and runs the verb. If it holds by construction, the plan says so in one line with the import fact. For
AC21, Task 9 or Task 7 adds a fixture with `DUBPLATE_ADMIN_EMAIL` set to a live non-admin's address, whose
`list_members` call answers `forbidden`.

### PA-12. minor. Plan:427-442. Task 6 names no fixtures for V5's `busy`, the unreadable window, the window's correlated rule, or the blocking mailer.

**Defect.** Several mutations and states rows have no fixture that meets them:
- "Fall back to a default window on a read failure" needs a window reader that errors.
- V5's row ("any write ... `unavailable`, detail `busy`") is in P8 and the states table, with no fixture and no
  mutation.
- AC12's last clause, "refuses a window at or above the absolute cap", and P6's floor and ceiling are named in no
  fixture.
- "Send the notice inside the write callback" needs a blocking mailer and a concurrent writer, and the list names
  only a recording mailer and a failing one.

**Fold.** Add these fixtures:
- an erroring window reader;
- a held write lock under a short deadline, against one settings write and one member write, each answering `busy`,
  with the mutation "map `ErrBusy` to `internal`";
- a window equal to the cap and one a second below it;
- a mailer that blocks until released, with a concurrent writer whose deadline would expire.

### PA-13. minor. Plan:227, :389, :499. Two proofs cannot fail as written.

**Defect.** (a) The states row "`member add --admin` on a roster with no live member commits whatever the cap" and the
mutation "refuse the bootstrap at the cap" are unreachable. P7's floor is 1, so a roster with zero live members is
never at the cap. A tombstone-counting mutant is already caught by the tombstone fixture. (b) Task 8's mutation "skip
the validator on the CLI reset" is caught by none of its fixtures. Every listed reset removes a refusal, and none adds
one.

**Fold.** For (a), keep the bootstrap fixture as PS A1's positive row. Restate its reachable form: an empty roster
holding cap-many tombstoned rows takes the bootstrap. Drop the unreachable mutation. For (b), add a fixture where
deleting an `admin` row breaks a correlated rule, for example a nonce-retention reset against an admin-raised
magic-link expiry (`nonceRetentionCoversExpiry`). The reset is then refused.

### PA-14. minor. Plan:403-404, :491. AC22's #25 and #12 riders carry no proof.

**Fold.** Task 5 adds a test-first fixture: an `admin` row whose value does not parse. The refusal wraps the parse error
(`errors.Is` against it), and the test fails at HEAD. Task 8 asserts "(required)" in each listed verb's usage text.

### PA-15. minor. Plan:462-463, :468. AC7's "no job wire type carries a queue position or order field" has no test.

**Defect.** V4 validates Go job values against the document. An extra optional property in the document's job schema,
such as a queue position, would not fail it.

**Fold.** The V4 test also asserts that each job schema's property set equals the Go type's JSON field set.

### PA-16. minor. Plan:251-252, :584-585. One conductor-owed item and one PS clause are not carried.

**Defect.** (a) The fold's conductor-owed list ends with an open question: whether rung 7 and a running rung 12 slice
may both write `internal/serve`. It "still needs its ruling", and the plan does not carry it to STATUS. (b) PS:145-147
says raising the cap from the command line "notifies the admin". The spec's erratum 5 says the command line sends no
notice, and 14a builds no CLI set verb. Task 11 step 9's PS note says only that the cap landed at 14a.

**Fold.** Step 9's STATUS bullet carries (a) as an open decision. The PS note adds one line: 14a has no command-line
set verb, and erratum 5 rules the command line sends no notice.

### PA-17. minor (over-ceremony). Plan:60-68, :42-44. The rendered leg runs on every task. The cost is priced and largely unavoidable.

**Defect.** `check.sh --scope-changed` diffs from `git merge-base master HEAD` (`check.sh:177-193`), so the rendered leg
runs from the first task that touches a trigger path (`check.sh:246-259`) to the end. Task 2 already edits
`cmd/dubplate-server/auth.go`, `serve.go`, and `web.go`, and Task 4 edits `internal/auth`. Moving Task 1's comment
edit therefore saves at most Task 1's run. The plan prices this at about 22 minutes a task. The blast-radius rule in
pass-core would want the task's own diff, and `check.sh` has no base override.

**Fold.** No plan change is required. File a BACKLOG item for a `check.sh` base override (for example `HEAD~1` on a
task's own commit) as a gate-economy improvement, so later `auth-data` passes can scope to each task's own diff.

## Summary

Seventeen findings: 0 blockers, 4 majors, and 13 minors, one of them over-ceremony. Every spec AC (1-22), owed item
(V4-V7), the doctor warning split, all eight errata, and PS A1/A2 have an owning task. The defects are proofs that can
pass vacuously or fixtures that are missing. The top three:

1. **PA-1.** AC12's refresh clause and its `IssuedAt` mutation sit at the seam, where the test builds the actor, so they
   prove nothing. The one channel that sets freshness, Task 7's verifier, has no mutation.
2. **PA-2.** Task 10 retires `roster`, but more than five existing rendered rows and the harness's reset server depend
   on it. These sit outside Task 10's file list and outside 14a's contested-surface grant, and Task 10's proofs name a
   component-test kind the repo does not have.
3. **PA-3.** The close puts the architecture reads and their Go fold after the merge ritual's gates. Every
   code-changing fold should come before the close gate.
