# Rung 14a plan review: lens P-C, domain risk

Target: `docs/superpowers/plans/2026-10-10-rung-14a-admin-api.md` at `fae3de3` ("the plan"), against the approved spec
revision 2.1 ("the spec"). Lens: auth, sessions, and data integrity, plus the unattended-run rules. Code read at
`fae3de3` to check the plan's claims; no staleness check against HEAD. The spec's design, the member cap's default,
TypeScript 7, and `@types/node` are settled and not reopened.

Findings rank correctness gaps by consequence. Polish is omitted.

## Verified sound (no finding)

- **Epoch reads are not cached.** `EpochStore.readNormalized` reads the file on every call (`internal/auth/epoch.go:183-200`),
  and `VerifyIdentity` checks the epoch and the roster on every request (`internal/auth/seam.go:95-123`). A CLI
  `member remove` or re-add while `serve` runs therefore takes effect at the removed member's next request. The
  roster and the epoch share one normalizer (`roster.go:147-149` calls `identity.NormalizeAddress`).
- **`Handle.Update` never re-runs its callback** (`internal/engine/store/write.go:63-95`). A bump inside the callback
  runs once, and a failed commit after it leaves the member signed out, which is the safe direction.
- **The serve unit grants the auth leaf read-write** (`systemd/dubplate-server-serve.service:42`). A web removal's
  bump can write the epoch file on the box. The runbook's `systemd-run` form grants the same leaf
  (`docs/server-runbook.md:190`).
- **An unregistered subtree answers the shell's 404** (`internal/serve/serve.go:248-258`), so removing the `roster`
  registration leaves no route open.
- **The add path's actor re-read already checks the tombstone** (`roster.go:380-386`, `tombstoned_at IS NULL`). V6's
  move to `UserID` keeps that clause.
- **The live smoke runs on the workstation, and the box is never contacted** (plan:96-99). That meets CLAUDE.md's
  safety rails, and 6d's smoke set the precedent (HISTORY:331-337).

## Findings

### D1. major: the security read after segment B leaves out the roster

**Location:** plan:555-557 (Task 11, step 1).

**Defect.** The after-B read covers "the write path, the gate, the notice, the mount, and the session check". Task 4
lands in segment B and carries the pass's highest-consequence code: the removal, the demotion, the last-admin guard,
the epoch ordering inside `Update`, the re-add bump, V6's in-write actor re-read, and the member cap. That task is not
in this read's scope. The after-C read covers only "removal's CLI half". Task 3's authorizer, which holds the
privileges table and the member-versus-admin split, is in neither read. Neither does any other read name the epoch
ordering, which is where a removed member could regain access.

**Fold.** Add Task 4 (`internal/auth/roster.go`, `epoch.go`) and Task 3's authorizer and privileges rows
(`internal/ops/opshttp/authorize.go`, `catalog.go`) to the after-B read's scope. Name the questions it must answer: a
removed or re-added member's old session, the guard's place in the transaction, the bump's place relative to the guard
and the commit, and the cap's channels.

### D2. major: the stop rules contradict each other on a second `fix`, and the close's failures have no rule

**Location:** plan:56-58, plan:84-85, and plan:90-91.

**Defect.** Lines 90-91 stop the run on "a `fix` still standing after its one fix round on a real defect" (pass-core:
"a real defect after its one fix round"). Lines 57-58 and 84-85 instead let the conductor, alone, weigh an upshift on
a second `fix` for Tasks 4 and 7 and split a task "whose second `fix` names a seam". An unattended conductor that meets
a second `fix` on Task 4 has two contradictory instructions. Following the wrong one either halts a recoverable run
or grinds an auth defect without stopping. The close's own failures also have no verdict: a red live auth smoke, a
render-and-read defect, or a red merge-ritual gate.

**Fold.** Pick one rule. For example: on a second `fix`, the conductor may make one upshifted re-dispatch on Opus
(Tasks 4 and 7; Task 6 is already Opus) or one seam split, and a `fix` still standing after that stops the run. Then
add the close's rule: a red smoke step or a page defect becomes one fix task through the chain under `auth-data`,
followed by a re-run of the failing smoke steps, and a second red stops the run before the merge ritual.

### D3. major: the ceiling rule is silent on the close, the B-boundary security read, and the merge ritual

**Location:** plan:80-82.

**Defect.** At the 8.4M flag, "the current segment finishes, and no new segment starts". Three readings are left open:

- whether the close (Task 11) counts as a segment;
- whether the after-B security read and its fold, which run between segments, still run;
- what happens when the 10.5M stop lands mid-merge-ritual, between `--ff-only` and the on-master gate, which leaves
  master advanced and unverified.

The arithmetic makes the first reading likely. Plan:38-40 puts 3.2M of overhead, mostly the close's thirteen reads,
the smoke, and the merge, on top of 6.9M of tasks, so the flag most likely falls during segment C or the close.

**Fold.** Add four rules:

- The close is not a new segment. Once the close starts, it runs to the hard ceiling.
- The after-B security read and its fold belong to segment B, so segment C never starts without them. If the flag
  stops the run at boundary B, STATUS names the security read as the first resume step.
- The merge ritual starts only when the spend left covers it (about two full gates plus one `diff-reviewer`). Once
  started, it is atomic with respect to the ceiling: it finishes, or it stops before `--ff-only`.
- STATUS names the step reached.

### D4. major: Task 2 migrates the sensitive mark with no proof that the existing marks survive

**Location:** plan:306-324 (Task 2, class `tool`).

**Defect.** The new closed mark decides two things: which writes need the re-authentication gate (Task 6) and which
writes send the notice (AC13). Task 2 re-declares the mark on every existing key. Today `Sensitive: true` is set on
`DUBPLATE_MAGIC_LINK_EXPIRY`, `DUBPLATE_SESSION_LIFETIME`, and `DUBPLATE_SESSION_ABSOLUTE_CAP`
(`cmd/dubplate-server/auth.go:126-155`). Task 2's acceptance proves only that a missing description or a zero mark
panics. If one of those keys is mis-marked as not sensitive, every test still passes, and a borrowed cookie can then
change the absolute cap with no fresh sign-in and no notice. Under `tool`, coverage gaps do not block, so the
diff-reviewer has no bar to catch it.

**Fold.** Keep the class. Add one acceptance row: a table test over the production registry asserts the exact set of
store-homed keys marked sensitive (the three above, plus `DUBPLATE_REAUTH_WINDOW` and `DUBPLATE_MEMBER_CAP` once
Tasks 4 and 6 add them). Add one mutation that flips one key's mark. The cost is one test.

### D5. major: no task proves that a reset writes its audit row

**Location:** plan:406-413 (Task 5) and plan:495-500 (Task 8).

**Defect.** The spec makes `settings_audit` the durable record of every sensitive change: "the `settings_audit` row
is the durable record", because a crash can lose the notice (spec:280-285). The spec says `reset_setting` "audits
the same way" (spec:244), and `settings reset` writes "its audit row" (spec:291-292). AC11 does not repeat the audit
clause, and neither Task 5's nor Task 8's acceptance or mutations name it. A reset that skips its audit row, or writes
it outside the transaction, passes every listed proof.

**Fold.**

- Task 5's acceptance adds the reset's audit row: kind, actor, channel, key, the old value, and an empty new value,
  written in the same `Update`.
- Its mutations add "skip the audit row on reset" and "write the audit row after the commit".
- Task 8's acceptance asserts the CLI reset's row with the command-line actor and channel.

### D6. minor: Task 4's Files list omits two files the task must edit

**Location:** plan:356-358.

**Defect.** Task 4 declares `DUBPLATE_MEMBER_CAP` in a new `api.go`. Every settings file registers through
`newProductionRegistry` (`cmd/dubplate-server/main.go:99-108`), so `main.go` needs a new `register...` line. Without
it, the seed rule, `list_settings`, and `set_setting` never see the key. Tests built on a test registry still pass. Task
4 also changes `AddMember`'s cookie-channel actor (V6) and adds a `member_cap` outcome. The roster API's handler is
an `AddMember` caller (`internal/auth/admin.go:265`, mapped at `:305`) and stays live until Task 10, so it must
change in Task 4. That file is a contested surface (spec:524).

**Fold.** Add `cmd/dubplate-server/main.go` (one registration call) and `internal/auth/admin.go` (the roster API's
call site and outcome map only, nothing in `RequireAdmin`) to Task 4's Files. Add an acceptance row: the production
registry lists `DUBPLATE_MEMBER_CAP`, and from Task 6 on, `DUBPLATE_REAUTH_WINDOW` too.

### D7. minor: two ordering mutations are missing from Task 4

**Location:** plan:387-389.

**Defect.**

- The epoch file sits outside the store's rollback. A bump placed before the last-admin guard therefore survives the
  guard's rollback. A sole admin's refused self-removal would then still sign them out of every browser. The spec
  orders the bump "after the tombstone and the guard" (spec:350), but no mutation pins that order.
- No mutation pins the cap's count to the transaction, so two concurrent adds at cap minus one could both land.

**Fold.**

- Add the mutation "bump before the guard". Its proof: a refused last-admin removal leaves the admin's session
  verifying.
- Add the mutation "count live members outside the `Update`". Its proof reuses AC16's second-connection lock fixture.

### D8. minor: the settings write does not fail closed when the cookie channel arrives with no guard

**Location:** plan:396-399 (Task 5) and plan:168-170 (P10).

**Defect.** The in-write admin re-read is a guard parameter, "absent on the command line". If `internal/ops/admin`
passes no guard on a cookie-channel call, the write silently skips the re-read. AC16's settings clause would catch
that only if the test reaches the nil path.

**Fold.** Task 5's write function refuses a non-command-line channel that arrives with no guard. Add a mutation:
accept a cookie-channel write with no guard.

### D9. minor: P6's correlated rule leans on a non-strict precedent and adds a start-time state no table names

**Location:** plan:149-154 (P6) and plan:218-230 (the states table).

**Defect.** P6 refuses a window "at or above" the absolute cap, "on the precedent of `sessionLifetimeWithinCap`".
That precedent's check is `durationAtLeast`, which refuses only `l < s` (`cmd/dubplate-server/auth.go:276-291`). It
admits equality, so reusing it as cited builds the wrong rule. This is a recalled behavior, not a quoted one. The rule
also adds a start-time refusal the states table does not name. `DUBPLATE_SESSION_ABSOLUTE_CAP`'s floor is `1s`
(`auth.go:146-155`), so two environments fail:

- An environment whose cap is at or below `2h` fails the seed rule at `serve`'s first start on 14a, against the
  window's default.
- An environment whose cap is at or below `5m` has no valid window at all.

The box is not live, so this is a defined report, not a migration hazard.

**Fold.**

- Quote `durationAtLeast` and state that the new rule is strict. AC12's fixture gets a window equal to the cap.
- Add a row to the states table: `serve` start with a cap at or below the window refuses, naming both keys; a cap at
  or below `5m` names the window's floor.

### D10. minor: the order of the cap read and the bootstrap check is open

**Location:** plan:225-228 and plan:376-378.

**Defect.** The plan says "A cap that cannot be read refuses" and "The first-admin bootstrap is never refused". If the
add path reads the cap before it checks for an empty roster, a malformed stored `DUBPLATE_MEMBER_CAP` row refuses the
bootstrap. This question has one correct answer: PS:140 says "never refused".

**Fold.** State that the bootstrap check comes first and reads no cap. Add a fixture: an empty roster beside an
out-of-bounds cap row, where the bootstrap lands.

### D11. minor, OWNER FORK: a command-line raise of the cap and PS's notice

**Location:** plan:172-180 (the spec deviation) and plan:216.

**Defect.** PS:145-147 says "Raising the cap from the command line is audited under the command-line channel and
notifies the admin". At 14a the command-line path to the cap is the environment file plus the seed rule, which the
spec rules mails no one (spec:284-285, AC13: "A seed-rule change mails no one"). The plan takes the cap into 14a
without reconciling the two texts, so one ratified sentence goes unbuilt with no record.

**Options:**

- (a) The Task 11 docs fold adds a PS erratum: at 0.8 a command-line raise is the seed rule, audited and not mailed,
  because box access is the stronger credential, which is PS's own reason for skipping re-authentication.
- (b) The seed rule mails the live admins when it changes `DUBPLATE_MEMBER_CAP`. This adds a send path to `serve`'s
  start.

**Recommendation:** (a). It matches spec AC13 and adds no code.

### D12. minor: the states table omits a removed admin re-added over the web

**Location:** plan:218-230.

**Defect.** A tombstoned row is reactivated with every field taken from the candidate (`roster.go:408-417`), and a
cookie-channel candidate never carries the flag. A removed admin re-added through `add_member` therefore returns as a
member. That is correct, since the grant is CLI-only, but no row or fixture pins it. A reactivation that preserved the
old row's `admin` column would re-grant admin from the web.

**Fold.** Add the state "add over `add_member` of a removed admin's address: reactivated without the flag". Give it a
Task 4 fixture: a tombstoned admin row re-added on the cookie channel.

### D13. minor: the live smoke skips the cross-process operator path and the cap, and HISTORY drops a routing

**Location:** plan:565-571 and plan:585-591.

**Defect.** Review focus 4 (plan:241-242) is a CLI `member remove` while `serve` runs. The smoke exercises only the
web removal, so the cross-process path, with the CLI's epoch write and `serve`'s per-request read, is never run
live. The smoke also never meets `member_cap`. The 6d smoke left the `systemd-run` sandboxed verbs to rung 16
(HISTORY:337, :402). 14a adds three verbs that write the auth leaf, and HISTORY's planned entry does not route them.

**Fold.**

- Add a smoke step: `member remove` from the CLI while `serve` runs, after which the member's next request is refused.
- Add a smoke step: a cap seeded at the live count, so an add is refused on both channels. Each step is a few seconds.
- HISTORY routes the sandboxed `member remove`, `member demote`, and `settings reset` to rung 16 beside `member add`.

### D14. minor: two of Task 10's fixtures cannot reach their state as written

**Location:** plan:542-547.

**Defect.**

- A roster with two admins produces no `last_admin` refusal until one admin has been demoted. The proof needs a
  sequence, and the harness's roster reset (`cmd/dubplate-server/rosterreset_test.go`) has to restore the roster
  afterwards.
- The rendered harness runs the real binary under an environment (`rendered_test.go:369-391`) with no clock
  injection, and the window's floor is `5m`. A stale session therefore needs a cookie minted in the Go test with a
  backdated anchor. As written, the stale-session fixture names neither that cookie nor how to make it. A harness
  that cannot reach the state may be widened, and the harness's shape is rung 7's territory (plan:105).

**Fold.**

- Name the sequence: demote the other admin, then demote self.
- Name the stale-session mechanism: a session cookie signed with the fixture secret, with its anchor older than the
  window, added as an appended row. Otherwise prove the state in a component test against a stubbed client answer.

### D15. minor: the smoke's evidence can predate the merged code

**Location:** plan:565-575.

**Defect.** The smoke runs at step 5 and the rebase at step 7. When the rebase is not a fast-forward and master
gained `internal/auth`, `internal/serve`, or `cmd/dubplate-server/serve.go` changes in between, the smoke has not run
on what merges. The merge ritual's `diff-reviewer` is a static read.

**Fold.** When the rebase is not a fast-forward and touches those paths, re-run the smoke's auth steps on the rebased
head before `--ff-only`.

## Proportionality (this lens)

- **The class assignments fit.** Tasks 4 to 8 are `auth-data`, Task 6 runs on Opus, Task 10 is `engine-logic`, and
  Task 9 is `tool`. Task 2's `tool` class fits once D4's one table test lands.
- **The two security reads fit.** Each is followed by one fold, and their placement after B and after C fits the
  risk once D1 widens the B read. No full adversarial cycle is owed: CLAUDE.md reserves that for rungs 5, 12, and 16.
- **The workstation smoke is proportionate.** Its `5m` window wait is the one deliberate clock cost. Starting that
  wait first and running the other steps inside it removes most of it.

## Summary

The review found 15 findings: 0 blockers, 5 majors, and 10 minors, one of which is an owner fork. The top three are:

- **D1:** the after-B `web-auth-security-reviewer` read leaves out Task 4's roster code, which holds the removal, the
  epoch ordering, the re-add bump, the last-admin guard, and the cap. It also leaves out Task 3's authorizer.
- **D2:** the unattended stop rules contradict each other on a second `fix`, and they give no verdict for a red smoke
  or page defect at the close.
- **D3:** the ceiling rule is silent on the close, the B-boundary security read, and a mid-merge-ritual stop. The
  flag will most likely fall during segment C or the close.

The other two majors are D4, where Task 2's sensitive-mark migration has no proof that the existing sensitive keys keep
their mark, and D5, where no proof covers the reset's audit row, which is the durable record. The plan's core domain
mechanics hold up against the code: the epoch is read per request across processes, the callback runs once, the auth
leaf is granted, and an unregistered route answers 404.
