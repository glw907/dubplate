# Rung 6d plan review: domain risk (revision 2.3)

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md` at `4b2cddf` (revision 2.3).
Lens: data integrity, auth, and failure. Pass class `auth-data`. Every claim below was checked against the
tree at `4b2cddf` with `grep`, `sed`, and `git show`; nothing is from memory. No box access.

Counts: 0 blocker, 2 major, 8 minor (one minor is an owner fork).

## What holds

These were checked and need no fold:

- **Cross-process write safety.** `Handle.Update` runs one IMMEDIATE transaction
  (`internal/engine/store/write.go:44`, `conn.go:146` `_txlock=immediate`). So a `member add` from the
  CLI while `serve` runs, and two concurrent `member add --admin` bootstraps, serialize on SQLite's
  write lock. R-6d10's "every read inside the callback" rule is then sufficient for the actor check
  and the at-least-one-admin check.
- **No silent admin grant.** The admin flag reaches `users` only through the CLI grant. A cookie-channel
  reactivation writes 0. `RequireAdmin` reads `Admin` from a per-request `Lookup`
  (`internal/auth/seam.go:112`, `admin.go:92`). The `users` schema is `STRICT` with
  `id BLOB NOT NULL PRIMARY KEY` and `CHECK (admin IN (0, 1))` (`0003_catalog.sql:1-10`), so a zero
  id cannot be stored, and Task 4's `Valid` zero-id refusal cannot lock out a resolved member.
- **No lockout on upgrade.** Sessions live in the `auth` leaf, keyed by address, and the seam resolves
  them with `Lookup` per request. The re-added admin's existing session therefore resolves to the new
  row. Between steps 2 and 4 of the upgrade, `serve` is stopped, so the empty-`users` window is never
  live.
- **The store leaf on the box.** `deploy.sh` makes `/var/lib/dubplate` `music:music 0755`
  (`scripts/deploy.sh:264-266`), and `leaf.create` requires the euid to own the state directory
  (`leaf.go:64-78`). The serve unit runs as `music`, so the supervised start creates the store. An
  empty pre-created leaf classifies as `StateAbsent` (`inspect.go:19-21`). The plan's "absent" rows
  therefore cover the real fresh-install state.
- **Deploy order.** The staged `unidecode install` carries only `DUBPLATE_STATE_DIR`
  (`deploy.sh:361-373`). The superseded key in the old env therefore cannot fail the staged step
  before the env is re-rendered.
- **Ruling 3 (store identity to rung 16).** I agree. One box, one remote, and backup has never run
  there.
- **Ruling 4 (drop `member import`).** I agree. `d1e93e9`'s roster is keyed `streamer_username`
  (`git show d1e93e9:internal/auth/roster.go:37-109`). `parseRoster` refuses unknown members
  (`roster.go:340`). The import would have no reader.
- **R-6d20.** `grep -rn '\.Bump(' --include=*.go internal cmd`, outside tests, is empty. No epoch
  record exists to dual-key.
- **R-6d21.** The premise holds: `check.go:62-68` evaluates only when `r.Facts != nil`, and
  `evaluateTableAbsent` never reads environ (`evaluate.go:150-163`). Without it, Task 6 would
  regress `check` on a fresh install.
- **R-6d19.** `deploy.sh` never starts a unit (`:27-36`, `:422-426`). Moving the halt and the holder
  line to rung 16 strands nothing.

## Findings

### R1 (major): the rollback note sends the operator to a procedure the rollback binary cannot run

Location: plan `:753-756` (R-6d18, "Rollback past 6d") and Task 8's runbook outcome (`:1451`).

**Defect.** The upgrade's rollback copies are taken on the box immediately before the apply, so the
`.prev` binary is `d1e93e9` (STATUS `:8`). That binary predates the rename. Its `member add` takes
`--streamer-username` (`git show d1e93e9:cmd/dubplate-server/member.go:101`), and its loader reads
only `streamer_username`. The plan's note says to re-create the roster with "the pre-6d runbook's own
section (this runbook at `0cb33ae`, `:216-250`)". That section runs `member add ... --username`
(`docs/server-runbook.md:238-240`), which `d1e93e9` rejects as an unknown flag. A file hand-written
in the new shape fails `parseRoster`'s unknown-member refusal. An operator who rolls back after a
failed 6d start, following the runbook, cannot sign anyone in. The correct step is trivial: the
upgrade moved the file aside and never changed it.

**Fold.** In R-6d18 and Task 8's runbook outcome, make the rollback step "move `roster.json` back
from its aside name". Give the exact `mv`, since upgrade step 3 keeps the file intact. Drop the
reference to the `0cb33ae` re-create section. Add one line to Task 8's render-and-read check: the
rollback step names the aside file and no `member add`.

### R2 (major): the reconciler that 6d exports and wires into `serve` still tells the operator to run `member import`

Location: `internal/backup/reconcile.go:166-168`. Plan `:845` (6d-R11) and Task 1 (`:932`,
`:943`).

**Defect.** `report.head` writes `"the users table is empty: run member import"` whenever
`users` is empty (`reconcile.go:57-61`, `:166-167`). 6d-R11 reads the spec's restore-report sentence
as "not built". In fact it is already built, and Task 1 exports this reconciler and makes `serve`
its production caller. Under Ruling 4's yes, the verb does not exist, so the restore-recovery report
names a nonexistent command. A restore of a snapshot taken before the first admin is a realistic way
to reach an empty `users` table. The plan's Goal (`:158`) is "every operator string ... is true",
and the plan names no task that touches this string.

**Fold.** Task 1's files already include `internal/backup/reconcile.go`. Add an outcome: the
empty-`users` line names `member add ADDRESS --admin`. Add a reconcile-row state with an empty
`users` table that asserts the new text. Restate 6d-R11 as "the existing line is reworded to name
`member add --admin`". Under Ruling 4's no, the line keeps `member import`.

### R3 (minor): the runbook paragraph that calls the seed note false is not in R-6d18's edit list

Location: plan `:735-760` (R-6d18). `docs/server-runbook.md:340-345`.

**Defect.** The runbook says that "seeds at serve's next start" is "not yet true" and tells the
operator to "clear it from the environment". After Task 2, that advice is wrong. After Task 6, it
would have an operator delete auth tuning they meant to keep. R-6d18 lists the install order, the
upgrade section, rollback, `:265`, and `:773`, but not this paragraph. Task 2's string re-read
covers only `backup.go` and `store.go`.

**Fold.** Add the paragraph to R-6d18's list: replace it with the seed rule as `serve` now applies it.
Add the paragraph to Task 8's render-and-read check.

### R4 (minor): the local keep falls to the default when `Inspect` saw a busy store or an intent

Location: plan `:522-523` (R-6d4, "No keep row, or no facts ... busy ... restore intent"). Task 1's
mutation list (`:986-991`).

**Defect.** The keep is read before `OpenServe` from a pre-lock `Inspect`. Two of the no-facts states
are transient: busy, meaning another holder has the lock, and a restore intent. `serve` then waits,
migrates, and prunes pre-migration copies to the registry default of 3, even when the stored row says
10. That prune deletes the rollback copies the operator asked to keep. No production state reaches
it at 6d (the box has no store, and no migration is pending beyond `0005`). It becomes reachable at
the first future migration.

**Fold.** `ServeOptions` has no production caller yet, so its signature may change (R-6d16 already
changes it). Take one of two options, the conductor's call:
1. Make `Keep` a function `OpenServe` evaluates after it holds the lock, over the handle.
2. In R-6d4, rule that a busy or intent `Inspect` re-inspects once after `OpenServe`'s wait, before
   the migration.

Add one Task 1 state: an older store whose `Inspect` reports busy, with a stored keep of 5, prunes to
5.

### R5 (minor): a start that the seed will refuse still migrates the store first

Location: plan `:478-480` (R-6d2) and `:530-536` (R-6d5).

**Defect.** Today the wall load refuses a bad auth value before `serve` touches the store. After Task
6, the refusal moves to the seed, which runs after `OpenServe` has already migrated. On a future
upgrade, a typo in the env therefore migrates the store forward and then refuses. The rollback binary
then refuses the newer store, and the runbook's rollback copies exclude the store
(`server-runbook.md:817-820`). R-6d4 already runs `Inspect` and `EvaluateFacts` before the open, and
R-6d21 makes that evaluation report the seed's refusals for every store state it can vouch for.

**Fold.** In R-6d2, refuse before `OpenServe` when the pre-open evaluation, over present facts or
R-6d21's empty row set, already refuses. The seed apply after the open stays the authority. Add one
Task 2 state: an older store plus an unparseable env value refuses with the store still at its
version and no pre-migration copy written.

### R6 (minor): the privileged-fields table omits restore, and R-6d22's grep proves inserts only

Location: plan `:443-457`, `:805-814` (R-6d22), and `:1205-1207` (Task 4's grep).

**Defect.** `restore` replaces the database wholesale (6b), so it is a channel for every `users` field,
`admin` included. Planning-miss item 3 asks for every channel by name. Restoring a snapshot taken
when a since-demoted member was admin is the one real way the flag returns. R-6d22's proof also greps
only `'INSERT INTO users'`. That grep misses `UPDATE users` (AddMember's own reactivation path is an
UPDATE), `REPLACE INTO`, and `INSERT OR REPLACE`. A stray `UPDATE users SET admin` elsewhere would
pass the proof.

**Fold.**
- Add `restore` (an operator verb, the whole row set from the snapshot) to the `users.admin`,
  `users.username`, and "any `users` row" lines.
- Widen Task 4's grep to `grep -rniE '(insert|replace|update|delete)[^;]*\busers\b' --include=*.go
  internal cmd`, outside tests. It should find only `AddMember`'s statements and
  `reconcile.go`'s count.

### R7 (minor, OWNER FORK): member removal has no channel from 6d until an unscheduled verb

Location: plan `:84-85` ("no member removal") and R-6d10 (`:636-641`).

**Defect.** Before 6d, an operator could remove or demote a member by editing `roster.json`, and the
code acknowledges hand edits (`roster.go:754`). After 6d, removal requires raw SQL against the live
store. No spec, ROADMAP line, or STATUS row schedules a removal verb. STATUS names only "the revoke
verb" and slice 6's "offboarding deprovision", which presume one. A compromised or departed member
cannot be cut off by any supported means at 0.8.0 unless a rung picks it up.

**Options.**
- (a) Route `member remove` (tombstone) to a named pre-launch rung, such as rung 14 beside the admin
  write path or slice 6 beside offboarding. Record it in STATUS as launch-blocking. Cost: nothing in
  6d.
- (b) Build a minimal CLI `member remove` in 6d. It would tombstone in one `Update` and carry the
  last-admin guard R-6d10 routes. Cost: about 0.5M, and it crosses the ten-task line.
- (c) Document an interim `sqlite3` tombstone statement in the runbook. This needs `sqlite3` on the
  box, unverified.

**Recommendation:** (a). The box has one member and nothing is live before 0.8.0. The routing must be
explicit, though, so launch cannot ship without it.

### R8 (minor): the upgrade section names neither the aside path, nor where the admin's fields come from, nor a re-run

Location: plan `:168-177` (Owner touchpoints) and `:748-752` (R-6d18, the upgrade section).

**Defect.** Step 4 needs the admin's address, display name, and username. The only record of them is
the file step 3 moves. A re-run after a rollback (R1) meets a live row: `member add` then refuses it
as a duplicate, which is correct but unstated. Planning-miss item 2 asks for each state a check can
meet.

**Fold.** In the upgrade section:
- Step 3 names the aside path (`roster.json.pre-6d`, or the re-plan's name).
- Step 4 says to read `address`, `display_name`, and `streamer_username` from that file.
- One line covers a re-run: the store already holds the admin, so `member add` refuses as a
  duplicate, and `member list` showing that admin is the pass condition.

### R9 (minor): the behaviors table quotes `Update` without the word the concurrency argument rests on

Location: plan `:375`.

**Defect.** R-6d10 and the review checklist (`:862-863`) rely on reads inside `Update` being atomic
with the write across processes. That holds only because the transaction is IMMEDIATE
(`write.go:44`, `conn.go:146`). The quoted row omits it, so planning-miss item 1 is unmet for the
one property that keeps a concurrent CLI `member add` and a page add from both passing the
admin check.

**Fold.** Prefix the row's quote with "Update runs fn inside one IMMEDIATE transaction on the write
pool". Cite `conn.go:146`.

### R10 (minor, over-ceremony): Task 1 re-proves `OpenServe` behavior the store package already tests

Location: plan `:946-964` and `:973-975`.

**Defect.** About eight of Task 1's fifteen rows re-test `internal/engine/store` behavior through the
preparation function. These include newer-than-binary, any other intent, E11, the interrupted
create, the exclusive lock held past the wait, and a current store under a shared holder. The store
package already covers them: `TestOpenServe_LockHeldPastMigrationWait`, `_WaitsForExclusive`,
`_CurrentKeepsSharedLock`, `_LostDatabaseBesideBackupRef`, and `_ReconcilePendingRefusals` in
`open_test.go` and `intent_test.go`. The three lock rows also need re-exec child processes with
1s waits. The serve layer adds only four things: the options it passes (keep extraction,
`MigrationWait`, `Reconcile`, `Logger`), the refusal-to-nil-shell mapping, and the handle's close.

**Fold.** Keep the serve-specific rows:
- absent, and the pre-created empty leaf;
- the four keep and prune rows;
- the reconcile row;
- one lock row, proving `MigrationWait` from the wall reaches `OpenServe`;
- the foreign-owner row;
- one refusal row (newer) for the nil-shell mapping.

Cite the store tests for the rest. Saving: about 0.2M and several seconds of gate time per run.

## Planning-miss items

1. **Library behavior quoted.** Met, except R9.
2. **Check states named.** Met for every task table. The gaps are the upgrade section's re-run
   (R8) and the reconciler's empty-`users` state (R2).
3. **Privileged fields' channels.** Restore is missing (R6). Everything else is named.
4. **Proof fixture states.** Met. Tasks 1, 3, 4, and 6 each name the state that keeps the
   assertion from passing vacuously.

## Proportionality

The 10.5M ceiling, the three segments, and the two security reads are proportionate for an
`auth-data` pass that moves the identity source. The appendix costs nothing unless Ruling 4 is no.
R10 is the one task-level over-specification found. Task 8's store-specific leg check is cheap, so
no finding is raised.
