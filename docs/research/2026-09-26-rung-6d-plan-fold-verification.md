# Rung 6d plan fold: verification read

Target: `docs/plans/2026-09-26-rung-6d-identity-serve-store.md`, revision 2, against the
fold record `2026-09-26-rung-6d-plan-fold.md`, the three lens reviews, revision 1 at `5afe46b`, and
the approved 5d plan. The reader took no part in the drafting or the fold. Code was read at committed
`HEAD` (`adb9644`); the working tree's warm `web/` and `rendered_test.go` edits belong to another lane
and were not read.

**Counts: 1 blocker, 3 major, 4 minor.** Every blocker and major is a task that cannot go green as
scoped, or a runbook step that cannot run on the box. None reopens a ruling.

## 1. Did the 19 majors close?

All 19 closed at the cited location, with two qualifications.

| ID | Closed at | Verdict |
| --- | --- | --- |
| C-M1 | R-6d2 `:298-312`; Task 1 criterion `:685-690`; Task 2 mutations `:751-753` | Closed. The listener mutations are gone; the preparation function's return is the proof point. |
| C-M2 | Task 1 rows `:679-681`, mutation `:706` | Closed. |
| C-M3, X-M5 | Task 1 rows `:672-673`, criterion `:694-696` | Closed. The counts work: 2 + 1 = 3 pruned to 1, and 3 + 1 = 4 pruned to 3. `migrate.go:121-123` confirmed. |
| C-M4 | R-6d10 item 2 `:428-430`; privileged fields `:271`; Task 3 row `:798`; mutation `:824` | Closed. |
| C-M5, R-M1 | R-6d10 items 1 and 4 `:425-436`; Task 3 rows `:804-805`; review focus `:591-592` | Closed. The `AddMember` doc quote was checked against `roster.go:753-756`. |
| C-M6 | R-6d10 item 3 `:431-434`; Task 3 rows `:797`, `:800`; API test `:818-820`; Task 5 row `:929` | Closed. |
| C-M7 | Task 5 criterion 1 `:934-938`; 6d-R12 | Closed. |
| C-M8 | Task 5 criteria `:945-949` | Closed. The trigger probe was re-run here and reproduces: both tables hold 0 rows after `ROLLBACK`. See minor m4. |
| C-M9 | R-6d11 `:462-469`; Task 5 row `:923`, mutation `:961` | Closed. |
| C-M10 | Task 8 criteria `:1131-1140` | Folded in part and refused in part. The refusal is reasonable (question 5). |
| X-M1 | R-6d4 `:326-346` | Closed by the R-m2 synthesis. The in-bounds row uses the pre-seed row value, and an out-of-bounds row uses the healed value in force. This matches `outcome`'s doc (`evaluate.go:269-274`). |
| X-M2 | R-6d7 `:374-376`; Task 6 row `:1008`, mutation `:1024` | Closed. `parse` quoted correctly (`evaluate.go:367-378`). |
| X-M3 | Task 7 `:1065-1067`; 6d-R9 | Closed. |
| X-M4 | Task 7 `:1062-1064` | Closed. The `clampExpiry` and `VerifySession` quotes were checked at `session.go:144`, `:215`. |
| X-M6 | Task 6 interim read `:993-997` | Closed. The fold does not repeat this defect at 6d's other seams, though; see major M1. |
| R-M2 | R-6d11 `:462-465`; Task 5 row `:920`, mutation `:960` | Closed. `readRoster`'s check was confirmed at `roster.go:318-320`. |
| R-M3 | R-6d19 `:546-556`; Task 8 `:1107-1110` | Closed on `try-restart` and the step order. The review's named mutation (swap `try-restart` for `restart`) was dropped without a recorded refusal; see minor m1. |

## 2. Findings, ranked by consequence

### Blocker

**B1. `rendered_test.go` needs a fixture rework that the global constraint forbids, and F1's premise is wrong.**
Location: global constraint `:261-262`; code pointer `:193-194`; Task 1, 4, 6, 7 Files
(`:656`, `:849`, `:984`, `:1042`); fold record F1 (`fold.md:92`).
Defect: `rendered_test.go` never calls `newServeShell`. Its two call sites (`:424`, `:448`) call
`buildServeShellWithLogger`, which lives in `serve_test.go:66-89`. A signature change therefore lands
in `serve_test.go`, which is already "their tests". So the F1 allowance permits an edit that is never
needed. The edit the file does need is forbidden, because the rendered harness is built on the roster
file:
- `renderedEnviron` (`:364-379`) writes `roster.json` with no `streamer_username` and sets
  `DUBPLATE_ROSTER_FILE`. After Task 5, that key fails every verb's wall load. After Task 4, the file
  is never read.
- `rosterResetServer` (`:609-650`) resets the capture matrix by rewriting that file.
- The `members/roster-file` subtest (`:914-923`) proves the add-member write by reading that file.

The rendered test is in the full gate, so Task 4 goes red, then Task 5, and no task may fix it.
Fold: replace the constraint with "`rendered_test.go` changes only as the store move requires". Add
the rework to Task 4's Files:
- seed the two fixture members into the store with streamer usernames;
- make `rosterResetServer` reset the `users` table, keeping its HTTP contract (a POST answers
  `{"members": n}`) so `web/scripts/rendered/run.ts` needs no change and R-6d17 holds;
- have `members/roster-file` read `users` instead of the file.

Add the `DUBPLATE_ROSTER_FILE` removal to Task 5's Files. Correct the code pointer at `:193-194` and
the F1 row.

### Major

**M1. Task 3 is not green on its own: its signature and `Valid` changes break callers it does not own.**
Location: Task 3 Files `:768-775` and outcomes `:779-783`; R-6d10 `:414-418`.
Defect: R-6d10 adds `ctx` to `Lookup` and `List`, but Task 3's Files name only `roster.go` and
`seam.go` among the callers. The other production callers are `handlers.go:210,308,366`,
`admin.go:206`, `member.go:162`, and `doctor.go:173`, and none of them is Task 3's. Second, Task 3
makes `Valid` refuse a zero `UserID` on the cookie channel, but `serve` stays on the file-backed
roster until Task 4. That roster has no id, so its `Lookup` yields a zero `UserID`. `RequireAdmin`
refuses on `!v.Identity.Valid()` (`admin.go:92`, `:121`), so every admin-gated test over `serve`,
the rendered harness included, fails between Task 3 and Task 4. This is X-M6's defect class (a seam
that is not green on its own) at a seam the fold did not re-check.
Fold: Task 3's Files gain the `ctx` threading at those six call sites (a mechanical change). Then
either move `serve`'s roster construction into Task 3, or defer `Valid`'s zero-id refusal and its
mutation to Task 4, where `serve` first reads `users`.

**M2. Task 5 supersedes `DUBPLATE_ROSTER_FILE` three tasks before the template drops it, and 5d's rendered-doctor test is unnamed.**
Location: Task 5 `:912`; Task 8 `:1115`; R-6d12 `:480`; Task 0 item 2 `:616-623`; 5d plan Task 6
(`2026-09-25-rung-5d-boundary-deployment-close.md:1060-1095`).
Defect: 5d lands a gate-resident rendered-doctor test. It renders `env/dubplate-server.env.template`,
runs `doctor`, asserts exit 0, and asserts exit 1 "naming the `roster` row when the roster file is
absent". It also asserts that deleting the `DUBPLATE_ROSTER_FILE` row fails naming the missing key.
Task 5 makes that key superseded while the template keeps its row until Task 8, so from Task 5 to Task
8 the template renders an environment every verb refuses, and 5d's test fails. Task 4's change from
the roster row to a members row already breaks its roster-row assertions. No task names the test, and
Task 0 item 2 does not ask for it. `TestTemplateKeysClassifiedAgainstProductionRegistry` stays green,
since a superseded key classifies, so the break shows only in 5d's test.
Fold: move the template row's removal into Task 5's commit, which R-6d12's "same commit" wording
already implies. Name 5d's rendered-doctor test in Task 0 item 2. Add it to Task 4's Files for the
members-row assertions and to Task 5's Files for the key assertions.

**M3. The 6d upgrade path cannot run on the box: the store does not exist and nothing starts `serve`.**
Location: Owner touchpoints `:109-110` ("restart `serve`"); Task 8 runbook outcome `:1116-1119`;
R-6d19 `:546-556`; R-6d11 `:575`.
Defect: after 5d, the box has no store. 5d's step 14 started the pre-6d binary once, and that binary
never opened a store. 6d-R11 relies on exactly this fact. The upgrade section reads "remove the
roster row, deploy, `member import --dry-run`, `member import`", and `deploy.sh`'s `try-restart` is a
no-op on the stopped unit. `member import` opens through `OpenWrite`, which refuses an absent store
("start serve first"), so both import steps refuse. Owner touchpoints' "restart `serve`" is the step
R-6d19 forbids `deploy.sh` to take, and it names no alternative.
Fold: after the deploy, the upgrade section adds one supervised start in 5d step 14's form:
`systemctl start`, then `doctor` showing the store created at head, then `systemctl stop`. The unit
stays disabled. The import steps follow the stop, run through `systemd-run` with the store leaf in
`ReadWritePaths`. Change Owner touchpoints to match. Task 8's render-and-read checks this order.

### Minor

- **m1. R-M3's mutation was dropped as though folded** (`fold.md:83`; Task 8 mutations `:1146-1148`).
  C-M10's refusal covers a step-order grep, but not the presence check R-M3 asked for. A leg check
  that `deploy.sh` never runs `systemctl restart`, `start`, or `enable` on the serve unit is one
  grep on the leg's precedent, and it guards the production-box outcome R-6d19 exists to prevent.
  Add it with its mutation, or record the refusal.
- **m2. Two new `Update` callers have no stated deadline** (R-6d5 `:348-356`; Task 5 `:906`).
  `Update` refuses a context without a deadline (`errNoDeadline`, `write.go:23-25`, `:69`). The
  unexported `apply` (`apply.go:29`) sets none, and `serve`'s start context carries none, so the
  first run of Task 2 fails until someone adds one. The same holds for `member import`. Fold: name
  the deadline, on `memberVerbWriteWait`'s precedent, in R-6d5 and Task 5.
- **m3. R-6d19's "render the env, install the binary" reverses 5d's install order in the wording**
  (`:552-553`). 5d's `--apply` swaps the binary before it installs `server.env`
  (5d plan `:1255-1260`). The rule that matters is that `check` runs after both installs. Fold: say
  "after the binary swap and the `server.env` install", so Task 8 does not reorder 5d's steps to
  match the wording.
- **m4. The fault fixture catches only one variant of the "record outside the transaction" mutation**
  (Task 5 `:947-949`, `:958`). A record written in its own `Update` after the users insert never
  runs when the insert faults, so the fixture passes under that variant. Fold: name the variant, a
  record written in its own transaction before the inserts.

## 3. New mechanisms: quoted or probed?

Every added row of the behaviors table (`:226-236`) was checked against source, a man page, or a
re-run probe, and each holds:
- `Handle.QueryRow`: `conn.go:107-111`, verbatim.
- The `users` schema: `0003_catalog.sql:1-10`.
- `readRoster`'s mode refusal: `roster.go:318-320`.
- `validateRoster`'s doc: `roster.go:531-537`, two lines off the cited `:529-535`, which is trivial.
- `AddMember`'s doc: `roster.go:751-756`.
- `clampExpiry` and `VerifySession`: `session.go:144`, `:215`.
- `linkExpiryCeiling`: `nonce.go:69-76`; today's caller passes the constant `magicLinkExpiryCeiling`
  (`serve.go:255`).
- `parse`: `evaluate.go:367-378`.
- `outcome`'s heal doc: `evaluate.go:269-274`.
- `Inspect`'s early returns and version constants: `inspect.go:122-127`, `:206`, `:227`.
- `claim`'s panic: `registry.go:126-131`.
- `AddSuperseded`: `registry.go:530-544`.
- `systemd-run`'s EXIT STATUS, `systemctl try-restart`, and `runuser`: each quote matches the
  installed man page.
- The trigger rollback and `/proc/locks` listing both shared holders: both probes re-run here and
  reproduce.

The Ruling 2 premise greps also reproduce: `.Bump(` has no production caller, and no demote or remove
path exists. No mechanism was stated from memory. The one factual error is F1's call-site claim,
covered by B1.

## 4. Can the owner rule each Ruling on the evidence?

Yes, all three. Ruling 1's basis (6c's 4.5M and 0.75M) matches `docs/HISTORY.md:216-222`, and the
arithmetic holds. B1, M1, and M2 add files to existing tasks, not tasks, so the ten-task line stands.
Task 0's re-derivation should expect Task 4 to be heavier. Ruling 2's premise was re-verified. Ruling
3 states what each answer builds and where the exposure begins.

## 5. Are the refusals reasonable?

- **R-m6 (no settle-loop bound): reasonable.** A bound's proof needs a test-only stub for `settle`,
  and the second-settle warning makes a spin loud. The residual case is `settle` returning nil below
  head, a bug in tested code, which the refusal honestly weighs as the smaller risk.
- **C-M10's partial refusal: reasonable** for the order grep and for `deploy.sh --check`. The cheap
  presence check in m1 is separate from the refused order check.
- **X-m9's partial refusal: reasonable** for `AddMember` (`admin.go:255` and `member.go:95` set the
  deadline) and for `Lookup`. It does not cover the two new callers in m2.
- **R-m10's partial refusal: reasonable.** No pre-6d store can exist, and M3 relies on the same fact.
