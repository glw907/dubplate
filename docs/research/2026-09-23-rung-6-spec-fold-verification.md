# Rung 6 spec fold verification

Target: `specs/2026-09-23-rung-6-ops-seam-store.md` revision 2 at `80ab7ff` (cited as R6:line),
against the fold record, the four reviews, and revision 1 at `12a6941`. Reader: fresh context,
adversarial, 2026-09-23. Tree citations checked at `80ab7ff`, modernc at the module cache's
v1.59.0, `flock(2)` from the local man page.

**Verdict: not ready for the owner's sitting as written, but close.** The fold is real. Every
blocker in the four reviews is fixed at its root, and I found no finding claimed folded that was
not addressed somewhere. The convergent defects (roster import, shared-lock migration, job
recovery, epoch re-key, three-file restore, seeding in every process) are gone. Nine of the
fourteen rulings (2 through 7, 11, 13, 14) can be ruled today. Four cannot, because the evidence
the owner would rule on is wrong or incomplete. Ruling 9's "no" branch claims a safety property
the epoch design does not give. Ruling 8's pre-deploy `check` contradicts 6d criterion 3. Ruling
1 understates the upgrade and omits the import as an admin-flag channel. Ruling 10 has no
deliverable in any pass. Separately, the fold introduced one structural defect: 6b-i's backup
guard and 6b-ii's reconciliation read catalog tables that land at 6c, so the pass order cannot
build two criteria as written. Several new mechanisms were stated from memory: the lock
downgrade after migration contradicts `flock(2)`, and the start-timeout bound cites the
incumbent's unit, which R5S rules 6d may not touch. A narrow fold of findings 1 through 13 makes
it sitting-ready. None needs new research beyond quoting two man pages.

Counts: 1 blocker, 12 major, 15 minor.

---

## Findings

### 1. BLOCKER. 6b-i and 6b-ii consume catalog tables that land at 6c

- **Location:** R6:558-560 (the store-aware count guard and `--max-delete` read "the catalog's
  track count"), R6:597-598 (restore step 7 reconciles rows against files), R6:901-902 (6b-i
  criterion 12), R6:930 (6b-ii criterion 10), R6:403 and R6:183 (the catalog lands at 6c),
  R6:69-71 (ruling 2's reason for the skeleton).
- **Defect:** The passes run 6a, 6b-i, 6b-ii, then 6c. At 6a the store holds only the job
  records and the ledger (R6:403). The count guard and restore's reconciliation both need
  `tracks` and `collections`, which do not exist until 6c. So 6b-i criterion 12 ("a root whose
  file count falls below the fraction") and 6b-ii criterion 10 ("a planted row without a file")
  have no table to plant into. The guard's fraction is also "store-homed" (R6:559), yet the
  settings home lands at 6c; only job retention gets a 6b-i constant (R6:502). Ruling 2's stated
  reason for landing `albums` and `tracks` at 6c is false: "the routed tables hold foreign keys
  into it" (R6:70-71), yet no CH-routed table references albums or tracks. Their real first
  consumers are the 6b guard and reconciliation, which is the pass-order defect again.
- **Evidence:** R6:403 against R6:558 and R6:597. The routed tables listed at R6:406-429 carry
  foreign keys to `users` and `collections` only.
- **Proposed fix:** Move the catalog skeleton (users, collections, albums, tracks, with the
  first populated fixture) into 6a or the head of 6b-i, and keep the CH-routed tables at 6c. Give
  the guard fraction a 6b-i constant that 6c re-homes, as job retention does. Correct ruling 2's
  reason. Also state what `--max-delete` is when the catalog holds zero tracks, which is every
  box until rung 7 or rung 10 writes one. Re-derive the 6a and 6c task counts and ceilings.

### 2. MAJOR. Ruling 9's "no" branch claims a safety property the epoch design does not give

- **Location:** R6:132-133 ("the dual-key epoch read ... keeps that safe too"), R6:714-720.
- **Defect:** After 6d, "New bumps write the `UserID` key" (R6:719). Under a "no", removing
  member A after 6d bumps only A's old `UserID` record. Re-adding A mints a new id. A stolen
  pre-removal cookie carries only the address (`internal/auth/session.go:71-90`, as IN15
  showed). It resolves to the new id, which has no record, and to the address key, which holds
  only pre-6d bumps. The revoked cookie verifies. This is IN15's exact scenario, and it survives
  under "no".
- **Evidence:** R6:718-720 against R6:132-133.
- **Proposed fix:** Either a removal bump writes both the `UserID` key and the address key
  (then the "no" claim holds), or ruling 9 says plainly that "no" requires that extra write. Add
  that case to 6d criterion 7.

### 3. MAJOR. Ruling 8's pre-deploy `check` contradicts 6d criterion 3

- **Location:** R6:125 ("`check` evaluates the import read-only before the deploy"), R6:692-693,
  R6:961 (criterion 3), R6:974 (criterion 11).
- **Defect:** Criterion 3 makes every verb fail the load while `DUBPLATE_ROSTER_FILE` is set.
  Before the 6d deploy, the rendered env still carries that key, so a 6d `check` fails the load.
  After the deploy, the key is gone, and `check` has no path to the roster file. No mechanism is
  stated for how `check` finds the roster or which binary runs it "before a deploy". The owner
  would rule ruling 8 partly on this mitigation.
- **Evidence:** R6:961 against R6:974. `member import` takes the path as `--roster PATH`
  (R6:51); `check` is given nothing equivalent.
- **Proposed fix:** Give `check` a `--roster PATH` evaluation mode that is exempt from the
  superseded-key load error, or make the evaluation `member import --dry-run --roster PATH`.
  Rewrite criterion 11 against the named mechanism and state where in the upgrade it runs.

### 4. MAJOR. The import is an admin-flag channel, and the privileged-field table omits it

- **Location:** R6:839 ("a user's admin flag | the CLI grant alone, as 5b-ii ruled"), R6:708-709,
  R6:49-61 (ruling 1).
- **Defect:** The roster file carries `admin` per member (`internal/auth/roster.go:110`,
  `Admin bool json:"admin"`). The import must carry it, or the box has no admin after 6d. The
  privileged-field table, which the planning-miss rule exists to make complete, names the CLI
  grant alone. An implementer following R6:839 drops the flags; one following the file keeps
  them. That is the ambiguity class that caused every 5b-ii fix round.
- **Evidence:** `roster.go:106-111`. R6:839.
- **Proposed fix:** Add "the import verb, from the roster file's `admin` field, once" to the
  admin-flag row, and add to 6d criterion 1 that an imported admin is an admin. State whether
  the import requires at least one admin in the file.

### 5. MAJOR. Ruling 1 understates the upgrade, and IN1's partial refusal is unsound

- **Location:** R6:62-63 ("one deploy ... plus one command"), R6:52-53, R6:261-262 (`serve`
  holds the lock shared), R6:212 of the fold record (the IN1 partial refusal).
- **Defect:** `member import` takes the lock exclusive, and `serve` holds it shared for life. So
  the upgrade is deploy, stop `serve`, import, then start `serve`. Between the deploy and the
  import, `serve` runs on an empty users table and nobody can log in. The ruling's adopter line
  says none of this. Worse, if anyone runs a 5c `member add` in that window (a runbook
  bootstrap, or an operator adding themselves first), the import then refuses forever, because
  it "refuses a users table that is not empty" (R6:52). Every roster member must then be
  re-added by hand. The fold refused IN1's "member writes refuse until the import row exists"
  because "any set key is already a load error". That reason does not cover this case: the key
  is unset after the deploy, and the add succeeds.
- **Evidence:** R6:52 and R6:261-262. The fold record's IN1 row.
- **Proposed fix:** State the real sequence and the login outage in ruling 1's "what it changes".
  Then close the dead end one of three ways: `member` writes refuse until an import record exists
  or the operator declares a fresh install; the import merges into a non-empty table, refusing
  only on address collision; or the import runs with `serve` stopped as part of `deploy.sh`.

### 6. MAJOR. The post-migration downgrade to shared contradicts `flock(2)`

- **Location:** R6:340 ("After migrating it downgrades to shared").
- **Defect:** `flock(2)`: "Converting a lock (shared to exclusive, or vice versa) is not
  guaranteed to be atomic: the existing lock is first removed, and then a new lock is
  established. Between these two steps, a pending lock request by another process may be
  granted." A `restore` waiting for the exclusive lock can win that gap. `serve` still has the
  connections it migrated through open, so `restore` swaps the file under them. That is the case
  ruling 6 cites (howtocorrupt section 2.5). `serve` then writes into the renamed-aside
  pre-restore copy. This mechanism was stated from memory, against the repo's planning-miss rule.
- **Evidence:** `man 2 flock`, quoted above.
- **Proposed fix:** Close every store connection before converting, then reopen and re-run the
  version check after taking the shared lock. Alternatively, hold the exclusive lock through the
  whole process life of the migration and take the shared lock on a second lock file first. Add
  a 6a criterion: an exclusive waiter granted during the conversion never leaves `serve` with a
  handle to the old inode.

### 7. MAJOR. The serve unit cited is the incumbent's, which R5S rules 6d may not touch

- **Location:** R6:58-59 (`systemd/dubplate-serve.service:45-55`), R6:257-258 (`User=music`),
  R6:339 ("The serve unit sets no `TimeoutStartSec`"), R6:1025 (the contested row gives 6d
  `systemd/dubplate-serve.service`).
- **Defect:** That file runs `/usr/local/bin/dubplate serve` from `/etc/dubplate/env`
  (`systemd/dubplate-serve.service:51-52`), the pre-pivot binary. R5S:1708-1713: "The incumbent
  `systemd/dubplate-serve.service` is the pre-pivot binary's unit and is left untouched until
  rung 16 retires it, so rung 5's unit takes a distinct file name." The contested row therefore
  directs 6d to edit a file a ratified spec freezes. The crash-loop and start-timeout arguments
  rest on a unit the greenfield server never runs under. The timeout argument also fails on its
  own terms: that unit is `Type=simple`, which systemd considers started at exec, so no start
  timeout bounds a migration wait at all.
- **Evidence:** `systemd/dubplate-serve.service:49-52`; R5S:1708-1713; R5S:2332 (5d builds the
  new unit).
- **Proposed fix:** Name 5d's greenfield unit (file name as 5d lands it) in the contested row and
  in every citation. Replace the timeout sentence with the real exposure: a timer-run backup
  holds the lock shared for the whole sync, hours on a first run, so a post-deploy migration
  leaves `serve` down until the backup ends. State the rule: the wait is bounded and fails
  loudly, or a deploy with pending migrations refuses while a backup runs. Quote systemd's
  `Type=` semantics for whichever unit type 5d chooses.

### 8. MAJOR. No pass owns wiring `serve` to the store, and a 5d deploy can crash `serve`

- **Location:** R6:338-341 (only `serve` migrates, at start), R6:626-629 (only `serve` seeds),
  R6:172-188 (the builds table), R6:1020-1025 (the contested rows), R6:188 and R6:1040-1042 (the
  store leaf's grant waits for 6d).
- **Defect:** `serve` must open the store, hold the lock, migrate before its listener opens, and
  at 6c apply the seed rule. No builds row, task list (fold record, "How the ceilings were
  derived"), or contested row names `cmd/dubplate-server/serve.go`. Rung 5's 5c and 5d edit that
  file heavily in parallel. There is also a sequencing hazard. If 6a merges before 5d, 5d's
  deploy builds from a master whose `serve` opens the store, while 5d's unit grants only the
  auth and roster leaves. The store leaf is neither pre-created nor in `ReadWritePaths` until
  6d, so `serve` fails at start on the box. A second gap follows: the store's exported write
  function has no production caller until 6b-i, which is exactly what CLAUDE.md's "nothing is
  exported ahead of a caller" forbids. The builds table's claim that `doctor` and `check` are
  the store's first production callers covers only the inspect open.
- **Evidence:** `grep serve.go` over R6 finds only the seed-handoff citation at R6:675.
- **Proposed fix:** Assign `serve`'s store wiring to a named 6a task and add a `serve.go`
  contested row shared with 5c and 5d. Do one of three things: record in PA's ownership table
  that 5d's unit and `deploy.sh` carry the store leaf; gate 6a's merge behind 5d; or keep
  `serve` from opening the store until 6d. State the write function's first production caller,
  or keep it unexported until 6b-i.

### 9. MAJOR. Rulings 7 and 10 have no deliverable, task, or criterion

- **Location:** R6:115-120 (ruling 7: snapshots "deleted only after a successful backup
  postdates it"), R6:136-139 (ruling 10: dubplate prunes dated snapshots and the archive),
  R6:602 (pre-restore copies follow ruling 7), R6:883-955 (no pruning criterion outside job
  records), fold record 6b-i and 6c task lists.
- **Defect:** A "yes" on either ruling is built by nobody. No pass deletes a pre-migration or
  pre-restore snapshot, and no pass prunes the remote `snapshots/` or `archive/`. Ruling 10 also
  names no keys and no defaults, and it references a "backup-failure detection window" defined
  nowhere. Its retention is "store-homed", which does not exist at 6b-i (the same gap as finding
  1 and CT19). Without local deletion, ruling 7's snapshots grow without bound on every box whose
  backup is off, which is the default.
- **Evidence:** A search for "prune" and "delete" in R6's criteria finds only 6c criterion 14
  (job records).
- **Proposed fix:** Add a task and criteria to 6b-i: local snapshot deletion after a postdating
  success, and remote pruning that never removes the latest snapshot. Give each a constant that
  6c re-homes. Define the detection window, for example as rung 9's failure-ping interval times
  a factor, or remove the clause. State what happens to ruling 7's snapshots when backup is
  disabled.

### 10. MAJOR. The inspect open's "byte-identical" criteria contradict the spec's own open semantics

- **Location:** R6:276-277 ("A read-only WAL open may create `-shm`"), R6:878-879 (6a criterion
  13: the state directory stays "byte-identical"), R6:955 (6c criterion 15).
- **Defect:** A store `serve` has closed cleanly has no `-wal` or `-shm`. A `mode=ro` open of it
  may create them, per R6:276 and sqlite.org/wal.html section 5, and a read-only connection does
  not delete them at close. So "byte-identical" fails for the older and newer cases, or it
  passes only if the implementer reads the version without SQLite. The criterion does not say
  which.
- **Evidence:** R6:276-277 against R6:878 and R6:955.
- **Proposed fix:** Define the invariant as "the database file's bytes and mode are unchanged;
  `-wal` and `-shm` may appear, owned by the leaf owner". Alternatively, read `user_version`
  from the header (offset 60) and the ledger through a real open only when the store is current.
  Quote wal.html section 5 in the planning-miss table rather than paraphrasing it.

### 11. MAJOR. 6a's migration criteria need injected migration sets, which the spec denies exist

- **Location:** R6:268-270 ("the one test seam rung 6 declares" is the `statfs` probe),
  R6:359-361 ("6a's proof is migration 1 over an empty store"), R6:866-872 (6a criteria 8-10),
  R6:925-926 (6b-ii criterion 7, a crash injected after each step).
- **Defect:** At 6a the embedded set holds migration 1 only. Criterion 9 migrates "a populated
  version-N fixture" and leaves `pre-migrate-vN.db`, which needs a migration N+1. Criterion 8
  needs a migration that fails after its version write. Criterion 10 needs a foreign-keys-off
  rebuild migration. None exists at 6a, so each needs a synthetic migration set passed to the
  runner, which is a seam. 6b-ii's per-step crash injection is a second seam. R6:359-361 also
  contradicts criterion 9 directly. The architecture reader grades test-only seams, and the spec
  pre-empts that review with a false "one seam" claim.
- **Evidence:** R6:268-270, R6:359-361, R6:868-870.
- **Proposed fix:** Declare the runner's migration source as an `fs.FS` parameter of an
  unexported function, which is ordinary dependency injection, and name how crash injection
  works (a subprocess killed at a named step, or a step hook). Drop "the one test seam", and
  reconcile R6:359-361 with criterion 9.

### 12. MAJOR. Restore across schema versions is unspecified, and a partial refusal rests on it

- **Location:** R6:587-590 (step 3 opens the current store), R6:597-598 (step 7 reconciles),
  R6:280-284 (the writing open refuses a version other than the binary's), R6:365-367 (the
  rollback path), R6:602-604 (a pre-6d snapshot under 6d), and the fold record's IN9 partial
  refusal.
- **Defect:** Two cases are unstated.
  - **Rollback.** An older binary restores its `pre-migrate-vN.db` over a newer current store.
    Step 3 opens that newer store. The spec does not say whether that open bypasses the version
    refusal or counts as "cannot open", which would need `--discard-current`. IN9's refusal of
    `--no-migrate` ("the older binary's own `restore --from-file` ... already rolls back without a
    flag") holds only if step 3 accepts a newer current store.
  - **Older snapshot.** Step 7 queries the restored store with the binary's catalog queries.
    Only `serve` migrates, so a restored older snapshot (a pre-6c store, or ruling 11's `--older`)
    is reconciled against a schema the queries may not match.
- **Evidence:** R6:587-590 and R6:597-598, read against R6:280-284.
- **Proposed fix:** Step 3 opens with a raw connection that skips the version check, since a
  checkpoint needs no schema. Step 7 runs only when the restored version equals the binary's.
  Otherwise the report is deferred to `serve`'s first start after migrating, recorded in the
  intent record. Add both cases to 6b-ii's criteria.

### 13. MAJOR. The staging cleanup can delete the snapshot a restore intent names

- **Location:** R6:254 ("Leftover staging files are removed at every writing open"), R6:283
  (a writing open refuses on an intent record), R6:585 and R6:591-594 (restore stages the
  incoming snapshot, then writes the intent).
- **Defect:** Order is unstated. If cleanup runs before the intent check, a writing open after a
  crash between steps 4 and 5 deletes the incoming file the intent names, and `--resume` cannot
  complete the swap. The fold added both rules without relating them.
- **Evidence:** R6:254 and R6:283.
- **Proposed fix:** The intent check runs first, and cleanup never touches a file an intent
  record names. Add that crash point to 6b-ii criterion 7.

### 14. MINOR. The maintenance connection's DSN is unstated, and `VACUUM INTO`'s fsync depends on it

- **Location:** R6:297-300, R6:311, R6:859-860.
- **Defect:** lang_vacuum.html: "if the PRAGMA synchronous setting of the original database is
  NORMAL or FULL, then SQLite invokes fsync() ... on the output database." The setting that
  counts is the maintenance connection's. Criterion 4 reads settings back on pooled connections
  only.
- **Proposed fix:** State that the maintenance DSN carries `_synchronous=FULL` and
  `_busy_timeout`, and add it to criterion 4.

### 15. MINOR. Lease files have no cleanup, and the kind lock has no named file

- **Location:** R6:478-482.
- **Defect:** Each process start creates `leases/<instance-id>`, and nothing removes it, so the
  directory grows forever. The exclusive kind lock's file is unnamed.
- **Proposed fix:** Recovery removes a lease file whose lock it acquired, after marking that
  owner's records. Name the kind lock, for example `leases/kind-<kind>`.

### 16. MINOR. Invariant 6 forces a `BLOB id` key onto keyed and singleton tables without saying so

- **Location:** R6:391-393; `settings` (R6:419), the ledger, the settings generation
  (R6:613-614), and `quota_counters`' group total (R6:426).
- **Defect:** "Every table's primary key is a `BLOB` column named `id`" gives `settings` a random
  id plus a unique key column, and the generation integer a one-row table with a random id. That
  may be intended, but the 6c plan will meet it as a surprise. The alternate-key exemption also
  exempts nothing, because the rule constrains primary keys only.
- **Proposed fix:** State the consequence, or exempt keyed and singleton tables by a named list
  in the test. Reword the exemption.

### 17. MINOR. The `visibility` CHECK conflicts with CH's API-level gate

- **Location:** R6:411, R6:938.
- **Defect:** CH:205-207 puts the refusal of "private" in the rung 14 API, conditional on the
  driver's capability report. A CHECK admitting only 1.0's value makes that conditional
  unreachable without a table rebuild. The spec's own open question notes the value was not
  re-read (fold record, "Unresolved").
- **Proposed fix:** Admit both CH values in the CHECK and leave the refusal to rung 14, or state
  the rebuild cost that "private" then carries.

### 18. MINOR. Errata are incomplete for the epoch, goose, and later rungs' scope

- **Location:** R6:1069 (erratum 10), R6:1072 (erratum 12), R6:93 and R6:206-209.
- **Defect:** The re-key language also stands at R5S:336-338 (R4), R5S:538-539 (R18), R5S:714
  (R35), and R5S:1809-1811, and erratum 10 names only R5S:1687-1690. R35's notification half
  (`EpochStore.subscribers`, keyed by address) has no stated key after 6d. S20:435's goose
  rationale ("ten database drivers") is now known wrong (ME30) and gets no erratum. Rung 12
  gains the streamer database in the backup set, and rung 16 gains the schedule key, the timer,
  and the rclone template rows. BS:844-851 names none of them, and the spec records no erratum
  or carried-list home.
- **Proposed fix:** Widen erratum 10 to every re-key passage and state the notification's key.
  Add the goose rationale to erratum 12. Add ladder-entry errata, or name the carried list, for
  rungs 12 and 16.

### 19. MINOR. Ruling 3 asks the owner to re-rule what CH already ratified

- **Location:** R6:78-85.
- **Defect:** CH:66 already homes the auth thresholds in the store, and the busy timeout is
  forced onto the wall by the DSN. The only real fork is the shadow pair's wall homing, which is
  the one part that owes a CH erratum.
- **Proposed fix:** Narrow ruling 3 to the shadow pair, and state the rest as applied.

### 20. MINOR. Rulings 10, 11, 13, and 14 lack a "What it changes" line, and ruling 12 is garbled

- **Location:** R6:136-161, R6:150-151.
- **Defect:** Rulings 1 through 9 each say what the adopter sees; four of the new ones do not.
  Ruling 12 says rotation "invalidates every outstanding credential, including revocations made
  after the last backup". A revocation is not a credential, so the clause reads backward.
- **Proposed fix:** Add the four lines. Rewrite as "rotation invalidates every outstanding
  credential, so revocations made after the last backup hold too".

### 21. MINOR. Some deliverables still have no criterion

- **Location:** R6:502 (6b-i's fixed retention constant), R6:788 (`doctor`'s WAL and free-space
  rows), R6:174-176 (the job records table at 6a), R6:898-906 (supersession messages).
- **Defect:** 6b-i prunes by a constant with no criterion. `doctor`'s WAL-size and free-space
  rows appear only in the risk model. CT15 asked for each retired name's message verbatim, and
  criterion 15 checks "its supersession message" without stating one.
- **Proposed fix:** Add one criterion per item, and write the messages in the spec or name the
  plan as their home.

### 22. MINOR. The uid refusal will fail a root-run `doctor` in `deploy.sh`

- **Location:** R6:256-258, R6:819.
- **Defect:** `deploy.sh` connects as `root@` (`scripts/deploy.sh:114`), and R5S:2332 has 5d run
  `dubplate-server doctor` against the rendered environment. After 6a that `doctor` refuses on
  uid. The refusal is right, but the cross-lane obligation is unrecorded.
- **Proposed fix:** Record in the contested rows that `deploy.sh` invokes store-opening verbs
  through `runuser -u <owner>`.

### 23. MINOR. Package territory omits the new edges into existing packages

- **Location:** R6:228-238.
- **Defect:** 6d's users-backed roster and accessor reads add `internal/auth` edges to `store` and
  `settings`, or require `cmd` to inject them. The table covers only new packages.
- **Proposed fix:** State the 6d edges or the injection shape.

### 24. MINOR. The rclone credential names depend on a remote name the spec does not fix

- **Location:** R6:97-107, R6:207-209.
- **Defect:** `RCLONE_CONFIG_<NAME>_*` varies with the adopter's remote name, and a `crypt`
  remote (ruling 14) needs a second remote's names. A template row cannot name an arbitrary
  `<NAME>`, and `doctor` cannot check required-when-enabled presence without it.
- **Proposed fix:** Fix the remote name (for example `dubplate`, with a `crypt` wrapper as
  `dubplate-crypt`), or state how `doctor` derives `<NAME>` from `DUBPLATE_BACKUP_REMOTE`.

### 25. MINOR. `rclone` becomes a gate prerequisite with no record in CLAUDE.md's Gate section

- **Location:** R6:567-569.
- **Defect:** A missing `rclone` fails the test by design, which makes `rclone` a gate
  prerequisite, and CLAUDE.md's Gate section lists every other tool the gate needs.
- **Proposed fix:** Route one line to CLAUDE.md's Gate section at 6b-i's close.

### 26. MINOR. The per-task rate is not traceable to a committed number

- **Location:** R6:987-988; fold record, "How the ceilings were derived".
- **Defect:** "Segments A through C spent about 4M across Tasks 1-8." Committed STATUS records
  segments A and B only, about 3.3M over six tasks, which is 0.55M per task. At 0.55M, the 35
  tasks add about 1.75M to the 27.5M total. The arithmetic in the table is otherwise exact: each
  row, each 80% flag, and the 27.5M total check out against the stated rates and the fold
  record's task counts. Every pass is under ten tasks.
- **Proposed fix:** Cite the HISTORY or STATUS line that records segment C, or use 0.55M.

### 27. MINOR. 6d's re-homing task likely exceeds one deliverable

- **Location:** fold record, 6d row ("re-homing and the `authConfig` rewire").
- **Defect:** One task moves ten keys' bounds into the registry and replaces `authConfig` with
  accessor reads at every call site 5c and 5d wire. ME13 flagged this rewire as uncounted work,
  and the fold counts it as half a task.
- **Proposed fix:** Split it into re-homing and rewiring (8 tasks, still under ten), and add
  0.5M.

### 28. MINOR. The roster import's "six states" does not match the list

- **Location:** R6:826, R6:692-696.
- **Defect:** The list reads as seven states (absent, unreadable, malformed, duplicate address,
  missing handle, populated or recorded, valid) or five, depending on grouping. 6d criterion 2
  proves "each import refusal state", so the count sets the test count.
- **Proposed fix:** Enumerate the states in a table and fix the count.

---

## Check 1: blockers and majors, claimed against revision 2

Every blocker (CT1-CT5, IN1, IN3, CS1, ME1, ME2) and every major is addressed at the place the
fold record names. Most original failure scenarios no longer occur. Four survive in part:

- **IN1 (b), `member add` before the import.** The scenario survives through a different door.
  See finding 5.
- **IN15, a re-added member.** The scenario survives on ruling 9's "no" branch. See finding 2.
- **IN6 and ME10, the store-aware guard.** The guard is folded in prose, but it cannot be built
  in its pass. See finding 1.
- **CT16, pruning.** The ruling exists and nothing builds it. See finding 9.

No criterion I checked passes vacuously in the way the reviews found. Criteria 6a-13 and 6c-15
are now contradictory rather than vacuous (finding 10).

## Check 4: the nine partial refusals

| ID | Sound? | Note |
| --- | --- | --- |
| CT2 | Yes | A name heuristic proves only itself. |
| CT24 | Yes | The contested table carries the shared files. |
| IN1 | **No, in part** | The write gate guarded the add-before-import dead end, which a superseded key does not cover (finding 5). |
| IN6 | Yes | Sound once finding 1 is fixed. Until rung 7 the catalog is empty, so a non-empty wrong mount passes. Say so in the runbook line. |
| IN9 | **Conditional** | Holds only if restore step 3 accepts a newer current store (finding 12). |
| IN11 | Yes | |
| IN20 | Yes | |
| CS9 | Yes | PA:235 sets no number; it sets a per-phase combined ceiling with lane sub-ceilings. |
| CT14 | Yes | A closed store is production-real. |

## Check 3: owner rulings

All fourteen are yes-or-no decidable. Every review fork is carried: the seventeen fork IDs map
onto rulings 1, 2, 4, 5, 7, 8, 9, 10, 11, 12, 13, and 14, and none is missing. Four rulings
carry defective evidence: 1 (findings 4, 5), 8 (finding 3), 9 (finding 2), and 10 (finding 9).
Ruling 3 is mostly already decided (finding 19). Ruling 6 is close to technical, given the quoted
SQLite hazard, but asking costs little.

## Check 6: citations new in revision 2, spot-checked

Forty-eight were checked, and all resolve to what the spec claims except the serve unit (finding
7):

- **The ratified specs:** CH:52-57, CH:56, CH:66-70, CH:74-75, CH:85-86, CH:107-110,
  CH:132-133, CH:195, CH:256-257, CH:461-462, CH:637-645, CH:661-662; S20:166-167, S20:363-372,
  S20:376-378, S20:385-386, S20:392; BS:39-40, BS:218-220, BS:656, BS:833, BS:934-935;
  OS:243-245; R5S:194-200, R5S:287-292, R5S:533-539, R5S:1153-1158, R5S:1243-1245,
  R5S:1552-1555, R5S:1687-1690, R5S:1809-1811; PA:209, PA:218.
- **The tree:** `auth.go:44-46`, `auth.go:230-280`, `auth.go:402-416`; `roster.go:109`,
  `roster.go:113-116`, `roster.go:286`; `auth/store.go:105-114`; `epoch.go:66-72`;
  `atomicfile.go:82-89` (inside `WriteStream`); `keys.go:60-71`, `keys.go:93-100`;
  `_salvage/internal/backup/backup.go:91-104`; `env/dubplate-server.env.template:110-117`.
- **modernc v1.59.0:** `sqlite.go:305-306` (the `_pragma` exception), and `sqlite.go:307-391`
  (the shorthand keys `_busy_timeout`, `_foreign_keys`, `_journal_mode`, `_synchronous`,
  `_query_only`, `_txlock`; no `journal_size_limit` shorthand).
