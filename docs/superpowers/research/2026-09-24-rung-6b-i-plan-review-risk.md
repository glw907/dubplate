# Rung 6b-i plan review: domain risk (data integrity and failure)

Target: `docs/superpowers/plans/2026-09-24-rung-6b-i-ops-seam-backup.md` at `b36871c`,
revision 1. Lens: data integrity and failure risk. Sources read: the plan in full; the rung 6
spec revision 2.2 ("Layout, ownership, and the lock", "Three open modes", "The ops seam",
"Backup and restore", the owner rulings, and 6b-i's criteria); the fold record's risk model;
and 6a's store code at `lane-2-6a` head `aa3b13e` (`lock.go`, `snapshot.go`, `open.go`,
`write.go`, `migrate.go`, `0002_jobs.sql`). Two rclone facts were re-probed against
rclone v1.74.3 on 2026-09-24 (R3, R5, R7).

Plan line numbers below are `plan:<line>` at `b36871c`.

## Counts

| Severity | Count |
| --- | --- |
| Blocker | 0 |
| Major | 1 |
| Medium | 8 |
| Low | 7 |

Owner forks: R1, R9, R10.

The job model holds up well under this lens. The lease design (instance lease before insert,
kind lease inherited by the orphan, the `WHERE`-guarded transitions) survives kill -9, a
concurrent second process, and a second backup. R-6bi7's `--max-delete 0` measurement is sound
for the backend it was measured on. The decline of the migration-file COMMIT guard holds for the
case it names (R13 is one narrow gap). The findings cluster in two places: what pruning may
delete (R1, R2, R9), and the edges of the remote and the library root that the proofs never
reach (R3, R5, R7).

## Findings

### R1 (Major, OWNER FORK): Task 5's local deletion disables 6a's lost-database guard, and count-only retention then lets 14 runs prune every good snapshot

**Location:** plan:1059-1062 (local copies), plan:1051-1054 (keep 14), plan:304; the guard is
`open.go` `besideCopy`/`newestCopy` at `lane-2-6a` (spec erratum E11).

**Defect.** E11 refuses to create an empty store when `snapshots/` holds a `pre-migrate-v<N>.db`
or `pre-restore-*.db` copy. It is the leaf's only memory that a store once existed. Task 5 deletes
every such copy older than the run's start after each successful backup. So on every box with
backup on, `snapshots/` is empty after the first backup, and E11 never fires again. The failure
chain then runs to completion without a guard:

1. The database file is lost (operator `rm`, a disk fault, a botched manual move).
2. `serve` (6d) finds no database and no copy beside it, and creates an empty store, as E11 allows.
3. The daily backup (rung 16's timer) succeeds. With zero tracks, `--max-delete 0` keeps the
   library safe, and an unchanged tree deletes nothing, so the sync passes. It sends a snapshot of
   the empty store as the newest.
4. Ruling 10 keeps the 14 newest by stamp. After 14 runs, every snapshot of the real store is gone.

A logically damaged store (a bad migration, a bug that deletes rows) takes the same step 3 and
4 path with no lost file at all. Ruling 7 was ruled on 2026-09-23. E11 landed on 2026-09-24.
Nobody weighed this interplay.

**Proposed fold (the fork).**

- (a) **Recommended.** A successful backup leaves a small reference file in `snapshots/`
  (`backup-<stamp>.ref`, holding the stamp and the remote). E11's `newestCopy` treats it as a copy
  beside the database and refuses, naming `restore --snapshot <stamp>`. The Task 1 territory
  covers it. This keeps ruling 7's deletion as ruled and restores E11's reach.
- (b) Add to (a) a regression floor on remote pruning. Record the store's page count, or its
  track and user counts, as snapshot metadata. Skip the prune (logged; the job still succeeds)
  when the new snapshot holds less than the guard fraction of the newest kept snapshot's count.
- (c) Grandfather-father-son retention (keep 14 daily plus the newest per week and month). This
  changes ruling 10's meaning.
- (d) Accept, and write a runbook line.

Recommendation: (a) in Task 5 now. File (b) against rung 7, the first writer of tracks.

### R2 (Medium): Prune trusts the wall clock. A backward jump prunes the run's own snapshot, and a forward jump purges the whole archive

**Location:** plan:446 (stamp), plan:1051-1058, plan:303, plan:962-963.

**Defect.** "The run's own snapshot is the newest, so it is never pruned" (plan:1053) is an
assumption, never a check. It fails in three cases:

- **Backward clock jump.** When 14 or more existing stamps are later than the run's stamp, the
  run's own snapshot is deleted right after it lands. Every later run's snapshot follows until real
  time passes the bad stamps, while `restore --snapshot latest` restores a stale store.
- **Forward clock jump.** Archive retention is "older than 90 days" from now. It purges every
  `archive/<stamp>/` at once. The archive is the only remote copy of every file the library deleted
  or replaced.
- **Stamp collision.** A backward jump, or two runs inside one second (the tests' back-to-back
  runs in criteria 11 and 15), reuses a stamp. `copyto` then overwrites a snapshot, and
  `--backup-dir` writes into an existing `archive/<stamp>/`, so criterion 11 can pass against the
  first run's directory.

The same gap covers a remote whose listing omits the new object (eventual consistency), which
the plan also never checks.

**Proposed fold.** Before any prune step, require that the `lsjson` listing contains the run's own
`<stamp>.db` at the stage's size, and that its stamp is strictly greater than every other parsed
stamp. Otherwise skip every prune, local deletion included, with the reason in the job message
(logged; the job succeeds). Measure archive age from the run's stamp. Refuse to start a run
whose stamp is not strictly after the newest remote stamp, with the failure code `clock-behind`.
Add one criterion 15 row with a future-stamped snapshot present, which ends with nothing pruned.
Record the forward-jump residual in "Not verified".

### R3 (Medium): The remote value is validated only by `doctor`, and the join rule is unspecified

**Location:** plan:952-956, plan:1207-1213, Task 6 wall rows (plan:~1150).

**Defect.** The job builds `<remote>/library`, `<remote>/archive/<stamp>`, and
`<remote>/snapshots/...` with no stated join rule, and neither the wall nor the job validates the
value. A 2026-09-24 re-probe confirms three hazards:

- **`name:`.** A naive join gives `name:/library`, an absolute path from the backend's root
  (rclone treats `loc:/x` as `/x`). `name:library` is relative, so the backup lands somewhere the
  operator did not name.
- **No colon (`backup`, `/mnt/x`).** rclone reads the value as a local path. Under an interactive
  run, a relative value writes a full library copy under the working directory, on the box's own
  disk, and can fill it.
- **An empty stamp or a join bug.** `rclone purge <remote>/archive/`, or `rclone purge <remote>`,
  deletes the whole archive or the whole remote. `purge` ignores filters.

**Proposed fold.** Put one validator in `internal/backup` that holds `doctor`'s rule: a name before
the first colon, `[A-Za-z0-9_]+`, and a non-empty `_TYPE`. Both `doctor` and the job's first phase
call it, and the job fails with `remote-invalid` before any guard. State the join: the remote
root, plus `/` only when the path part is non-empty and has no trailing slash. Add state rows for
`name:`, `name:path`, `name:path/`, and no colon. Build every `purge` and `deletefile` argument only
from a stamp that round-trips through the layout, and add a test asserting each argv ends in
`/archive/<valid stamp>` or `/snapshots/<valid stamp>.db`.

### R4 (Medium): The verb exits on the record's terminal state, not on the job's end, and a refused `queued → running` has no stated outcome

**Location:** plan:825-834, plan:1133-1141, plan:519-524 (R-6bi11).

**Defect.** Three paths let the process exit, or the record lie, while work continues:

- **A cancel that lands first.** A cancel from another process can land on the queued record
  before the owner's `queued → running` transition. The transition then matches no row, and the
  plan says only "refused and logged". If the body still runs, rclone syncs under a record that
  reads canceled. The verb polls, sees a terminal state, and exits 130 while its goroutine is
  mid-sync. Because rclone runs in its own process group, it continues as an orphan holding the
  kind lease and the store lock. The operator saw "canceled".
- **Work after the terminal write.** In the normal order, stage removal and the kind-lease release
  follow the terminal write. A verb that exits on the terminal state cuts them off, and the stage
  leaks until another process's recovery.
- **A failed cancel write.** The cancel write can fail (the store busy past the busy timeout).
  The job then never sees the cancel and runs to exit 0. `NotifyContext` keeps swallowing later
  SIGINTs, so the operator's only exit is SIGKILL, which makes the orphan R-6bi12 describes.

**Proposed fold.** A refused `queued → running` ends the goroutine before the body runs, and a
Task 3 row proves it: a cancel lands between the insert and the transition, and the stand-in
records no invocation. `ops/local` exposes a close that waits for its own jobs' goroutines. The
verb exits only after that wait returns, not when the record turns terminal. When the verb's
cancel write fails, the verb, which owns the runner in-process, cancels the job's context
directly. Add both to Task 6's states table.

### R5 (Medium): An absent `archive/` makes every early run report a prune failure, and no exit code carries a prune failure

**Location:** plan:1051-1069, plan:1063-1064 (prune failure), the Ruling 3 codes (plan:82-93).

**Defect.** Re-probed on rclone v1.74.3: `rclone lsjson <dir>` on a missing directory exits 3
("directory not found"). A remote has no `archive/` until a run first archives a file, which on a
growing library can be many runs. So on backends that model directories (`local`, `sftp`), every
early run logs a prune failure and puts it in the job message. The planning-miss rule asks for the
absent state, and the archive states table (plan:1073-1075) has no "absent" row. Criterion 15's
fixture pre-seeds two archive directories, so the gate never meets this state. A prune failure
exits 0 (succeeded), so rung 16's timer and rung 9's ping never see one. Operators learn to ignore
the message, and a real prune failure later (an unbounded remote) goes unnoticed.

**Proposed fold.** Treat exit 3 on the `archive/` listing as empty. Add an "absent" row for both
`archive/` and `snapshots/`, with a first-run fixture that has no `archive/`. Surface a prune
failure in `doctor` (the newest succeeded record's message carries it) so it is visible without
changing Ruling 3's codes.

### R6 (Medium): Criterion 3's proof passes with recovery disabled

**Location:** plan:865-868.

**Defect.** The criterion asserts that the child's running record "still reads running" after the
parent opens. An implementation whose constructor skips recovery entirely also passes. The plan's
review focus 1 names this as the top input class, and the proof has no positive control. This is
planning-miss item 4: the fixture state that keeps the test from passing vacuously is missing.

**Proposed fold.** In the same fixture, plant a second non-terminal record whose owner lease is
free. After the parent's open, assert that it reads interrupted while the child's record reads
running. Name mutation 8: "skip recovery at construction → the dead-owner record stays running."

### R7 (Medium): The library root has no named channel, and a symlinked root falsely aborts every run

**Location:** plan:810-813 (the constructor's inputs), plan:946-951, plan:966-977, plan:287-306 (the
privileged fields table).

**Defect.** Two gaps:

- **No channel.** The constructor takes "the backup configuration (flag and remote)". The job
  syncs `<root>`, but no task, table, or row names where the root comes from. The candidate is
  `DUBPLATE_LIBRARY_ROOT`, registered in `internal/config`. The privileged fields table omits it,
  though it decides what the sync may archive (planning-miss item 3). An implementer could read the
  environment in `internal/backup`, which the global constraints forbid, or sync the wrong
  directory.
- **A symlinked root.** Re-probed: rclone follows a symlinked root and syncs its contents.
  `go doc path/filepath.WalkDir` says "WalkDir does not follow symbolic links", and it Lstats the
  root. A count guard built on `WalkDir` counts zero under a symlinked root, so every run aborts
  with `library-empty` or `library-short`. A mount reached through a symlink is common on home
  servers. The states table has no symlink row, and the plan quotes neither behavior (planning-miss
  item 1).

**Proposed fold.** Name the channel. `cmd/dubplate-server` reads the library root through the
registry and passes it to the constructor, and a privileged fields row records that. Resolve the
root once with `filepath.EvalSymlinks` before both the guard and rclone, so both see one
directory. Add state rows for a symlink to a directory and a dangling symlink (`library-missing`),
and quote both behaviors.

### R8 (Medium): R-6bi7's guard is measured on one backend and one binary, and the box's rclone is never checked

**Location:** plan:405-423, plan:219-222 (the gate prerequisite), plan:1461-1470 ("Not verified").

**Defect.** Until rung 7 writes tracks, the zero-track guard is the only barrier between a
non-empty wrong mount and archiving the whole remote library. After 90 days that archive is
purged. The guard rests on one probe: the `local` backend, rclone v1.74.3, on the workstation.
The gate pins the behavior against the workstation's rclone. Production runs the box's rclone,
whose version nothing records or checks. On a backend without server-side move (S3 and B2, per
R-6bi9), `--backup-dir` goes through server-side copy and delete, which the probe did not exercise.

**Proposed fold.** Add two "Not verified" rows: "`--max-delete` counting of backup-dir moves,
measured on `local` only" and "the box's rclone version." Give `doctor` an rclone version row that
warns below the measured version. Route a runbook line to rung 16 that pins rclone at or above
v1.74.3 on the box. If a probe against a copy-only backend is cheap (an `s3`-compatible local
server is not), record it. Otherwise the note stands.

### R9 (Medium, OWNER FORK): Two servers sharing one remote path prune each other's snapshots, and "latest" may name the other's store

**Location:** plan:952-963, plan:1051-1054.

**Defect.** Nothing ties a remote path to a store. An adopter who points a second server at the
same path causes three problems. The second server's sync meets the first's files (mostly stopped
by `--max-delete`). Its runs count toward the 14-keep, and each prune deletes the other server's
snapshots. `restore --snapshot latest` (6b-ii) may fetch the other store. The second server can be
a test box, a rebuilt VPS beside the old one, or a second group on one bucket. Ease of setup by a
stranger is a stated product goal, so this configuration is plausible.

**Fork.**

- (a) Accept, and write a runbook line: one remote path per server.
- (b) **Recommended.** At its first successful run, the backup writes `<remote>/dubplate-store`
  naming a store identity. Every later run refuses with `remote-foreign` when the file names
  another store. This needs a store identity, which 6a may not carry: a random id written at
  create, which 6d's `serve` owns. So land the reader and writer in 6b-i if a store id exists at
  merge, and file it for 6b-ii otherwise.
- (c) Prune only snapshots whose metadata `run_id` names a job in this store's `jobs` table.
  Retention deletes job records after 90 days, so this is fragile.

### R10 (Low, OWNER FORK): `--max-delete` equal to the track count adds no bound beyond the count guard, and the guard counts every file across all collections

**Location:** plan:946-951, plan:976.

**Defect.** With tracks, the count guard passes at half the track count, counting every regular
file (art, `.nfo`, lyrics) across the whole root. `--max-delete` then permits archiving as many
files as the catalog has tracks, which is the whole library. A wiped collection mount under a
larger surviving collection passes both. A tree whose audio is wiped but whose artwork survives
passes too. Nothing is lost at once, since everything is archived, but it is lost after 90 days
unnoticed. At 6b-i every box has zero tracks, so the practical exposure starts at rung 7.

**Fork.**

- (a) Keep the plan as written. Rung 16's one-time relocation, a mass rename, needs a large
  delete bound.
- (b) Count per collection root, and only files with admitted extensions (FLAC at 1.0). Size
  `--max-delete` as `ceil((1 - fraction) × tracks)`, with an explicit operator flag for a planned
  mass change.
- (c) Keep (a) in 6b-i, and route (b) and the relocation interplay to rung 7's plan, the first
  writer of tracks.

Recommendation: (c), recorded in Task 8's handoff.

### R11 (Low): A backup stage named per instance blocks every later backup in a long-lived instance

**Location:** plan:393, plan:671, plan:983.

**Defect.** When stage removal fails ("logged; the job's state stands"), the next backup in the
same instance meets "the stage already exists" and refuses. Recovery never removes a stage whose
instance lease is held, which here is its own. In 6d's long-lived `serve`, backups then fail until
a restart. At 6b-i the verb is one-shot, so this bites first at 6d or rung 14.

**Proposed fold.** At job start, while it holds the kind lease, remove every
`staging/backup-*.db`. The kind lease proves that no other backup, orphan included, is live, which
is a stronger proof than the per-instance lease check. Otherwise, name the stage by job id.

### R12 (Low): An instance lease file is never removed on a clean close, and a removal race is not a stated state

**Location:** plan:631-641, plan:662-668.

**Defect.** The plan holds the instance lease "for the handle's life", but says nothing about
removing it at close. Recovery removes a lease only when a non-terminal record names it. So
`leases/` grows by one file per verb run and per `serve` start, forever. The "try another's lease,
free" path removes the file while holding it. A concurrent recovery that opened the same file
before the unlink can then win the lock and get ENOENT on its own removal. No row says that reads
as free and not refused.

**Proposed fold.** The handle's close unlinks its own lease file while it still holds the lock,
then closes it. The try path treats ENOENT on removal as free. Add a row and one test for each.

### R13 (Low): The COMMIT guard decline holds, with one gap: `COMMIT; BEGIN;` passes the gate silently

**Location:** plan:1303-1314; `migrate.go` `commitMigration` at `lane-2-6a`.

**Assessment.** The decline's reasoning is sound for its named case. A file ending in `COMMIT`
makes `tx.Commit()` answer "no transaction is active". That fails the gate's full-set apply on a
scratch store before merge, and a tokenizer would reject valid `CREATE TRIGGER ... BEGIN ... END`
bodies. A non-empty production store also has its pre-migration copy.

The gap is a file carrying `COMMIT; BEGIN;`. The runner's `Commit` then commits the file's own
second transaction and returns nil. The gate stays green, and the version write and the ledger row
land in a different transaction from the file's first half. If a statement after the inner `BEGIN`
fails in production, the first half stays committed at the old `user_version`, and the retry fails.

**Proposed fold (refusable).** Detection needs no tokenizer. Issue `SAVEPOINT runner` before the
body and `RELEASE runner` after it. Any `COMMIT` or `ROLLBACK` in the file destroys the savepoint,
so `RELEASE` fails loudly on the scratch store. It fires after the tail runs, like today's check,
but it turns the silent case loud. 6b-i adds no migration, so routing this to 6c, the next pass
that adds one, is also reasonable. The decline record should name the `COMMIT; BEGIN;` case either
way.

### R14 (Low): Criterion 20's fixture can signal the test itself, and it never exercises `WaitDelay`'s kill

**Location:** plan:1167-1172, plan:1041-1042 (mutation 7).

**Defect.** Two fixture gaps:

- **The process group.** The second row sends SIGINT "to the verb's process group". Unless the
  test starts the re-executed verb with `Setpgid`, the verb shares the test binary's group, and the
  signal reaches `go test` itself.
- **The kill path.** A stand-in that exits on SIGINT ends before `WaitDelay`, so "no child
  survives `WaitDelay`" passes with `WaitDelay` unset. Mutation 7 is left as "report if any test
  fires".

**Proposed fold.** State two fixture states. The test starts the verb as its own group leader. One
row's stand-in ignores SIGINT, so only `WaitDelay`'s `Kill` ends it, and mutation 7 then fails that
row.

### R15 (Low): With zero tracks, any deletion in the library fails every backup with exit 7 and no operator escape

**Location:** plan:405-423, plan:973-974.

**Defect.** Between 6d and rung 7 or rung 10, the store has zero tracks. Removing or renaming one
library file then makes every run exit 7 until tracks exist. No flag permits it. The failure is
loud and in the safe direction, so this is an operability gap and loses no data.

**Proposed fold.** Add a runbook line to Task 8's handoff that names the state and its remedy
(archive the file on the remote by hand with `rclone moveto`). Do not add a flag in 6b-i.

### R16 (Low): R-6bi3's export list omits the descriptors R-6bi12 passes to the child, and the plan omits one spec deviation

**Location:** plan:353-364, plan:500-505, plan:98-123.

**Defect.** Two gaps:

- **The descriptors.** `ExtraFiles` needs the store lock's and the kind lease's `*os.File`. The
  store's `storeLock.file` is unexported, and R-6bi3 names no accessor. Exporting the raw file lets
  a caller close it and release the store lock early.
- **The deviation.** Recovery in `ops/local`, a write-mode process, removes `staging/backup-*`
  files. Erratum E7 says "the write mode never does" remove staging. The instance-lease proof
  makes this safe, but the plan's deviation list omits it.

**Proposed fold.** Export one method that returns `dup(2)` copies of both descriptors for
inheritance, and state it in R-6bi3. Add "write-mode removal of a dead instance's backup stage" to
"Spec deviations that take effect on approval".

## The four planning-miss checks

| Item | Result |
| --- | --- |
| Behavior quoted, never recalled | Mostly met: `flock`, `os/exec`, `SysProcAttr`, `--max-delete`, and backend features are quoted. Unquoted: `WalkDir`'s symlinked root against rclone's (R7), `lsjson` exit 3 on an absent directory (R5), rclone's `name:` against `name:/` path semantics and `purge`'s filter-free delete (R3), and rclone's refusal of `--backup-dir` without server-side move or copy. That refusal is what makes the job safe on such a backend without `doctor`, so quote it in Task 4. |
| Every state names its report | Gaps: an absent `archive/` or `snapshots/` (R5), a symlinked root (R7), the remote value's forms (R3), a refused `queued → running` and a failed cancel write (R4), and ENOENT on lease removal (R12). |
| Every privileged field names every channel | Gaps: the library root (R7), and the run's stamp. The stamp decides what is pruned, and its only channel is the wall clock (R2). |
| Every proof names its fixture state | Gaps: criterion 3 has no positive control (R6), and criterion 20 needs the process group and a stand-in that ignores SIGINT (R14). Criterion 15 never meets the absent `archive/` state (R5). |

## Checked and found sound

- **A kill -9 of the verb mid-sync.** The instance lease is not inherited, so recovery marks the
  record interrupted. The orphan keeps the kind lease and the store lock, so a second start
  conflicts and names the holders, and a restore waits. Task 4's orphan row proves both halves.
- **A second backup.** The kind lease is taken before recovery and the insert, and criterion 7
  covers it.
- **A late terminal update.** Every transition carries a `WHERE` state guard, and criterion 5 has a
  real SIGKILL fixture.
- **A network drop.** A partial sync archives and never deletes. rclone skips deletions after I/O
  errors. A failed send prunes nothing (Task 5's failed-run row).
- **A full local disk.** The free-space check runs before the stage. SQLITE_FULL removes the stage.
  The stage is never kept (R-6bi6).
- **A remote that lies about features.** `doctor` reports it. The job does not depend on the
  answer, because rclone refuses `--backup-dir` itself. Quote that refusal, per the table above.
- **Local copy deletion under clock drift.** Deletion compares each copy's mtime with the run's
  start, so a backward jump keeps copies, which is the safe direction.
