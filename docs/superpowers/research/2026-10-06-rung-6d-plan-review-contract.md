# Rung 6d plan review: contract and criteria lens (revision 2.3)

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md` at `4b2cddf`, pass class
`auth-data`. Lens: is every promise testable, does each criterion name its fixture and its failure
reason, can any pass vacuously; proportionality; the four planning-miss items. Weighted toward the
parts the rebase record (`2026-10-06-rung-6d-plan-rebase.md`) lists as changed. Every claim below was
checked against the tree at `4b2cddf` (`0cb33ae` plus the two plan documents; `git diff --name-only
0cb33ae..HEAD` shows no code change).

Counts: 0 blocker, 2 major, 8 minor (correctness); 2 minor (ceremony). No owner fork.

## Correctness gaps, ranked by consequence

### C1 (major): `doctor` never prints a supersession message, so Task 5's criterion 4 and its template assertion cannot pass as written

Location: plan `:1266-1269` (criterion 4: "runs every verb that loads the wall and asserts each fails
the load with the message verbatim"), `:1243-1245` (`TestRenderedTemplateDoctor` "asserts a rendered
env carrying the row fails the load with the supersession message"), `:1278-1279` (the mutation "re-add
the template's row (`TestRenderedTemplateDoctor` fails naming the supersession message)"), and the
states table `:1262`.

Defect: `rejectNamespace` reports a superseded key as `"%s: %s"` (`internal/config/config.go:302-305`),
and `doctor` folds every Load error into a per-key map (`rowStatuses`, `doctor.go:369-391`), then
prints rows only for owned wall rows, reserved and integration classes, thresholds, identity sources,
and its fixed rows (`doctor.go:51-195`). After Task 5, `DUBPLATE_ROSTER_FILE` is none of those, so its
entry lands in the map and is never printed. `doctor` exits "configuration is incomplete"
(`doctor.go:198-201`) with every printed row reading ok. The verbs that return Load's error (`check`
`:30-33`, `serve` `:239-242`, `member` `:74`, `:150`, `backup` `:194`, `restore` `:152`) do name it;
`doctor`, the runbook's diagnosis verb, does not. The template test runs `doctor`
(`templatekeys_test.go:315-317`), so its supersession assertion has nothing to match. The repo's one
existing superseded-key test goes through `backup` (`backup_test.go:320`), which is why this never
surfaced. An operator whose hand-kept env still carries the row gets a failing `doctor` that names no
cause.

Proposed fold: Task 5 adds one conditional `doctor` line that prints every Load error no printed row
claimed (supersession, reserved-class, and unknown-key leaves), on the roster row's old precedent of
printing only on failure, so R10's unconditional count is unchanged. Criterion 4's table then asserts
the verbatim message for `doctor` on that line and for the other six verbs on their returned error.
The "deleting the roster row" replacement and the re-add mutation assert the same line. This adds
roughly a quarter of a deliverable. Task 5 stays at three.

### C2 (major): R-6d23's named baseline has no test that kills its own named mutation

Location: plan `:1210-1215` (the rendered-leg criterion and its fixture state), `:1223-1225`
(mutation "take the baseline from the table at the first POST (a test that posts before the
add-member write answers 2 where the named set answers 3 after it)").

Defect: the only stated proof is the full rendered leg, and its fixture state says the add-member row
has already written `renderedNewMemberAddress` before the first reset. Under that order a first-POST
snapshot holds the same three rows as the named set, so the mutation passes the rendered leg. The
test the mutation cites, POST then add then POST, is not among Task 4's criteria. This is the one item
5c-iv routed here ("must make the snapshot explicit"), and as written it can land unproved.

Proposed fold: add a Go criterion for `rosterResetServer` alone, over a store holding the two
seeded rows. It POSTs and expects 2, adds `renderedNewMemberAddress` and one `capture-*` row, then
POSTs again. It expects 3 and the capture row gone, asserted by a `users` read. Name that test as the
mutation's killer. The rendered leg stays as the integration proof.

### C3 (minor): the keep-row-of-1 fixture does not fix the env, so the `Outcome.Value` mutation can survive

Location: plan `:953` (state "at `0004`, an `env-seed` keep row of 1, two pre-migration copies, backup
off"), `:987-988` (mutation "take `Keep` from `Outcome.Value` ... with the env unset").

Defect: the mutation is killed only when the env is unset, because then the seed removes the row and
`Outcome.Value` is the default 3. The state row does not name the env. An implementer who sets
`DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP=1` beside an `env-seed` row of 1, the natural pairing, makes
`Outcome.Value` 1, and the mutation passes.

Proposed fold: the state row reads "... backup off, the keep key unset in the env."

### C4 (minor): the handle-close mutation names a test no criterion builds

Location: plan `:941-942` (outcome "the verb closes the handle when `serve.Run` returns"), `:990-991`
(mutation "drop the handle's close (a test that runs `restore` after the verb returns fails on
`ErrLockBusy`)").

Defect: R-6d2 confines the start-order proofs to the preparation function, and no test today runs
`newServeCmd` through `serve.Run` to return (`serve_test.go` has only `RefusesInvalidConfigWithoutListening`
and the wiring test). A test that calls the preparation function's returned close itself leaves the
verb's own close unproved, and the mutation survives. The same gap covers a refusal after `OpenServe`
succeeds (Task 2's seed refusal): nothing asserts that the error path releases the handle.

Proposed fold: one criterion. A Task 2 seed-refusal row asserts that `store.OpenWrite` (or an
exclusive `restore` open) succeeds right after the preparation function returns its error. For the
verb's close, either run the verb on a reserved port with a cancelled context and assert the same
afterward, or record the verb's `defer` as diff-reviewer-checked by name. Process exit releases the
`flock` in production, so the second option is proportionate.

### C5 (minor): the body cap's "next operation" read has no proof

Location: plan `:1391-1413` (Task 7 criteria), `:1416-1420` (mutations).

Defect: each apply mode is proven through a representative. The expiry and lifetime go through
criterion 9, the cap through the lowered-cap row, and the answer, lifetime, cap, and retention
through failing readers. `DUBPLATE_AUTH_REQUEST_BODY_CAP` appears in none of them. A mutation that
caches the body cap at construction passes every Task 7 criterion, and the review checklist
(`:868-870`) does not name it either.

Proposed fold: either add the body cap to the failing-reader list (a failing body-cap reader fails
the link request with the same 500 for a member and a stranger), or add one criterion-9 sub-case: an
`admin` row B lowers the cap, and a body between B and A is refused on the next request. The failing
reader is cheaper.

### C6 (minor): Task 8's misspelled-path row miscounts its refusals

Location: plan `:1463`.

Defect: with the unit granting `auth stor` and the array `(auth store)`, the existing check fires
(`stor` not in the array, `check.sh:710-713`), the new reverse check fires (`store` is granted by no
unit), and the new store check fires (the serve unit grants no `store`). That is three refusals, not
"twice". The criterion says to quote each refusal, so an implementer who sees three may read the
mismatch as a defect in the new checks.

Proposed fold: "fails three times: the existing array check, the reverse check, and the store check."

### C7 (minor): the R-6d21 and members-row tables leave the version-1 store unnamed

Location: plan `:798-803`, `:1338`, `:1190`, and `:1194`.

Defect: `Inspect` returns no facts below `jobsVersion` (2) (`inspect.go:240-242`) and sets
`SettingsAbsent` only for versions 2 and 3 (`readFacts`, `:271-277`). "Below `0004`" therefore covers
two different report shapes. A version-1 store has `Facts == nil`, so `check` (`check.go:62-68`)
still evaluates nothing, and the members row falls into none of the six rows. ("Older than `0003`"
assumes facts; "not read" lists refused, busy, newer, intent, and not-a-store.) No box can hold such a
store, because nothing has ever created one (`OpenServe` has no production caller at `0cb33ae`), so
the consequence is low. Still, the tables do not say which report a version-1 store gets.

Proposed fold: define both rules by `Report.Version`, not by `Facts`: "a store whose version is below
`settingsVersion`, facts or not". The members row's "older than `0003`" row then covers version 1 too.
Name version 1 in each table, or state that the `0003` fixture stands for both.

### C8 (minor): the `Valid` change breaks a third fixture, and Task 6 leaves `auth_test.go`'s table orphaned

Location: plan `:1146-1148` (the "two test fixtures that then need an id"), `:1301-1307` versus
`:1376-1378`.

Defect: (a) `internal/auth/admin_test.go:123-131`'s `fixtureAdmin` is a cookie-channel identity with
no id. Once Task 4's `Valid` refuses a zero `UserID`, `RequireAdmin` (`admin.go:92`, `:121`) refuses it,
which the Files list misses. The gate catches this, so it costs a fix round and nothing worse. (b) Task
6 replaces `resolveAuthConfig`'s env parse, because the ten `config.Threshold` values it passes to
`config.ParseDuration` and `ParseInt` (`auth.go:518-563`) stop being thresholds. Its Files list names
every test caller except `auth_test.go:830-890`, whose table Task 7 retires. Task 6 cannot be green
with that table still calling the old parse.

Proposed fold: (a) name `fixtureAdmin` beside the two `internal/ops` fixtures. (b) Move the
`auth_test.go` table's retirement and `resolveAuthConfig`'s deletion into Task 6, where the accessor
read replaces it, and leave Task 7 deleting only `authConfig`.

### C9 (minor): the hint criterion's "source string" scope is ambiguous, and `unidecode.go` carries the same hint

Location: plan `:1472-1473` ("no store test or source string names `runuser`"), `:700-703`.

Defect: `cmd/dubplate-server/unidecode.go:269` (and `unidecode_test.go:141`, `:198`) carry the same
`runuser -u <owner> -- dubplate-server ...` hint, outside Task 8's Files. Read repo-wide, the criterion
fails. Read as the store package, it passes. The unidecode verb reads `DUBPLATE_STATE_DIR` from the
environment (`unidecode.go:111`), which a root shell under `runuser` also lacks, so R-6d15's reason
applies to it as well.

Proposed fold: scope the criterion to `internal/engine/store`, and add one routed line in Task 9
naming `unidecode.go:269` for the next pass that touches the verb. Folding it into Task 8 is also
cheap (one string and two test substrings), but it widens the Files list.

### C10 (minor): criterion 4's "every verb" enumeration can go vacuous

Location: plan `:1266-1269`.

Defect: `newRootCmd` captures `os.Environ()` at construction (`main.go:126-135`), and `version` and
`unidecode` never call `config.Load`. A test that "runs every verb that loads the wall" has to pick
that set somehow. A hand-picked list passes when a later verb is missed, and a filter on "fails to
load" passes everything.

Proposed fold: the test names the seven loading leaves (`check`, `doctor`, `serve`, `member add`,
`member list`, `backup`, `restore`). It walks the command tree and asserts the leaf set equals those
seven plus the named non-loaders, so a new verb fails the test until someone classifies it.

## Planning-miss items: verdict

- **Quoted, not recalled:** met. Every behavior row I checked matches the tree: `rejectNamespace`'s
  empty rule, `claim`, `AddSuperseded`, `Inspect`'s early returns, `OpenWrite`'s refusals, `buildValidator`,
  and the `users` schema. The `runuser(1)` row supports R-6d15's conclusion, though its sentence is
  about which variables `runuser` sets, not about loading a file. The caller's environment passes
  through, and an operator's root shell carries no `DUBPLATE_*` keys, so the conclusion holds. One
  unquoted reliance is C1's: how `doctor` renders a Load error. That is where the miss sits.
- **Every state names its report:** met, except C7's version-1 store and C1's `doctor` row.
- **Privileged fields:** met. The table covers every `users` column that carries authority, and the
  grep criterion (`:1205-1207`) makes the single-insert rule checkable. The tests' direct
  `INSERT`/`DELETE` (tombstone fixtures, the reset server) are excluded by the "outside tests" scope.
  That exclusion is correct.
- **Fixture state named:** met for the substantive criteria (criterion 5's address-differs-from-username
  fixture, the reactivation's own old username, the no-admin golden's tombstoned admin, R-6d21's
  no-leaf and `0003` fixtures). C2 and C3 are the two places where the named fixture is the wrong one
  for its mutation.

## Over-ceremony, ranked by clock and token cost

### O1 (minor): Task 8 proves five failing leg states by full heavy-lane gate runs

Location: plan `:1466-1469` ("a full `check.sh` run if the leg cannot run alone").

Cost: the state-dir leg has no standalone entry point (`check.sh:657-737` sits inside the script), so
each of the five failing states becomes a full gate run. That is about 106s and Chromium per run, plus
quoting and reverting. Three of the five states exercise checks that already exist and were proven
when they landed.

Fold: demonstrate only the two new checks and the misspelled path, which shows all three together.
Run the leg's block alone (copied into a scratchpad script against the edited tree) rather than the
whole gate. The criterion keeps "the committed files pass" through the task's one real gate run.

### O2 (minor): Task 1 re-proves `OpenServe`'s own refusals through the preparation function

Location: plan `:948-964` and `:971-972` ("Every refusal row asserts an error, a nil shell, and the
store unchanged").

Cost: five of the fifteen rows belong to `OpenServe` and are already tested in `internal/engine/store`:
newer, any other intent, E11's copy-without-database, the interrupted create, and the foreign uid.
Wiring each into a cmd-level fixture duplicates those tests and adds about five subtests and their
fixtures to the plan's heaviest-rate task. `serve`'s wiring adds only one property on a refusal: it
returns an error and no shell.

Fold: keep one representative `OpenServe` refusal (newer) for the error-and-nil-shell property. Keep
the rows where `serve`'s own code decides the outcome (the keep extraction, the lock wait and holder,
the reconcile, the empty leaf), and cite the store package's tests for the rest.

## Not findings

The ceiling arithmetic (8.13M, the variants, and the 28 deliverables) re-adds correctly. Ruling 4 is
already posed to Geoff with its "no" content in the appendix, so this lens adds no fork. R-6d21's
reversal of 6c's "never reads environ" is a conductor design with one correct answer (avoiding the
regression), not an owner fork. The pass class and the two security reads match the identity surface.
