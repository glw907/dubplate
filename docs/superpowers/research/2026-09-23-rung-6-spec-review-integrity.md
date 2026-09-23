# Rung 6 spec review: data integrity and migration risk

Target: `docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md` at `12a6941` (cited as
R6:line). Lens: crash consistency, migration, backup and restore correctness, the env-seed and
roster handoffs, and the epoch re-key. Reviewer: fresh context, adversarial, 2026-09-23. SQLite
behavior below is quoted from sqlite.org pages fetched today, not recalled.

## Verdict

The store's core is sound, and the spec quotes it correctly. Per-migration transactions,
`IMMEDIATE` writes, every pragma in the DSN, `VACUUM INTO` snapshots, and the newer-store
refusal all hold up. Every serious gap sits at a seam the SQLite journal does not cover:
between the store and files (restore's multi-file move, the epoch records, the library tree),
between processes (any writing verb migrates, seeds, and recovers jobs, not just `serve`), and
between binary versions (rollback, a snapshot overwritten on the next upgrade, restoring a
pre-6d snapshot). Two findings are blockers. The roster import cannot survive a restart
without wedging serve. Restore's three-rename move can produce the exact
database-plus-foreign-WAL pairing it cites as the corruption case. The spec should not be
ratified until both are folded. Most majors have a small, mechanical fold. Four are owner
forks: re-adding a removed member (15), point-in-time restore of an older snapshot (5), what
an env key's removal means (19), and missing streamer usernames at import (2).

## Findings

### 1. The roster import trigger wedges serve after its first boot (blocker)

- **Spec:** R6:91-101 (ruling 6), R6:508-510, R6:644 (roster import states), R6:497-499 (the
  template keeps rendering keys).
- **Scenario:** Pass 6d deploys. The env still sets `DUBPLATE_ROSTER_FILE`: `deploy.sh`
  renders it from the age store, and today it is a required member of the identity source's
  correlated set (`cmd/dubplate-server/auth.go:46`, `doctor.go:110-115`). Boot 1 finds the file
  set and `users` empty, so it imports. Any later restart finds the file set and `users`
  populated, which is ruling 6's load error. The unit has `Restart=on-failure` with a
  `StartLimitBurst=30` (`systemd/dubplate-serve.service:45-55`), so serve crash-loops and then
  stays down. The import therefore needs the key present for exactly one boot and absent
  forever after, and no deploy step can express that. Three variants make it worse:
  - (a) A crash after the import commits but before the epoch re-key finishes lands on the
    load error at the next boot, so the re-key never resumes.
  - (b) If `member add` runs before serve's first 6d boot (a runbook bootstrap step, or the
    CLI writing verb that R6:205 says creates the store), `users` is non-empty. The import
    never runs, every existing member is dropped, and the next boot fails the load.
  - (c) "First open" is any writing verb (R6:204-206). The timer-run `backup` can import while
    the old serve is still live and still writing the file through web add-member
    (`roster.go:328-`). Members added in that window are lost.
- **Evidence:** The trigger is state-inferred ("users table empty"), not recorded. No verb
  owns the import.
- **Fold:**
  - Make the import an explicit, recorded step. Add an `identity_import` row holding the
    file's SHA-256, member count, time, and actor, written in the same transaction as the
    users. The trigger is the absence of that row, never an empty table.
  - Run the import from one owner: serve at boot, or a dedicated verb under the store lock
    held exclusive.
  - Once the row exists, a key still set whose file digest matches is a `doctor` warning
    naming the removal, for at least one release. A key whose file digest differs is the
    load error, and the message names the divergence (see 20).
  - `member` writes refuse until the import row exists whenever the key is set.
  - The re-key reads its own progress marker and resumes whatever the load state.
  - Add every variant above to the "roster import" states row (R6:644).

### 2. A valid rung 5 roster can fail the store's constraints at import (major, OWNER FORK)

- **Spec:** R6:517 ("`streamer_username` becomes required, unique, and stable"), R6:508-510.
- **Scenario:** `Member.StreamerUsername` is optional today ("Empty means unset",
  `internal/auth/roster.go:91-93`, `omitempty` at :110). A group whose roster leaves it empty
  deploys 6d, and the import transaction fails on the NOT NULL constraint. Under finding 1's
  current trigger, that failure happens at boot, so serve does not start.
- **Fold:** `check` evaluates the import read-only before deploy and names each member that
  would fail. The owner fork: either the import derives a username (a deterministic slug with
  collision suffixes, logged) or it refuses with a named remediation. Either way, the gap is
  caught by `check`, never at boot.

### 3. Restore's three-file move is not crash-atomic (blocker)

- **Spec:** R6:425-433, R6:583-584 (risk model), R6:193 (the lock file lives in the same
  leaf).
- **Scenario:** Restore renames `db`, `db-wal`, and `db-shm` aside one at a time, then renames
  the snapshot in.
  - Power fails after `db` moves but before `-wal` does. The leaf now holds the old hot WAL
    with no database. The resumed or re-run restore renames the snapshot in beside that WAL,
    which is the pairing sqlite.org/howtocorrupt.html section 1.4 names.
  - With the opposite order, a crash after `-wal` moves leaves the live `db` without its WAL.
    Every committed but uncheckpointed transaction silently disappears from the live store if
    the operator abandons the restore.
- **Evidence:** Three `rename(2)` calls are three durability points. The spec cites the
  pairing hazard and then builds a procedure with windows into it.
- **Fold:**
  - Under the exclusive lock, open the current store and run
    `PRAGMA wal_checkpoint(TRUNCATE)`, then `PRAGMA journal_mode=DELETE`, then close. That
    leaves one file.
  - Write a restore-intent record (atomic, dir-fsynced) naming the pre-restore and incoming
    names. Rename `db` aside, rename the snapshot in, fsync the directory, and remove the
    intent.
  - Every open replays or refuses on an intent record it finds.
  - Alternatively, keep the database in a `store/db/` subdirectory, swap whole directories
    under the intent record, and keep the lock file outside the swapped directory. A lock
    file that moves with the directory lets a new process lock a different inode.
  - The restore test injects a crash at each step.

### 4. A half-restored library is not guarded (major)

- **Spec:** R6:425-433, R6:642.
- **Scenario:** An operator restores the store and the library tree. The store swap completes.
  The rclone copy of a multi-hundred-gigabyte tree fails or is interrupted at 40%. Nothing
  records that the restore is incomplete. `serve` starts, the catalog references files that
  are missing, and rung 9 or a driver scan acts on the gap. Under the Navidrome driver, CH
  records that tracks marked missing get purged along with their per-user stars (CH:461-465).
- **Fold:** The restore-intent record from finding 3 persists until both halves verify.
  `serve` refuses to open while it exists, and names `restore --resume`. The verify step
  compares catalog paths against the tree.

### 5. The snapshot and the tree are not a consistent pair, and restore is silent on mismatch (major, OWNER FORK)

- **Spec:** R6:44-51 (ruling 1), R6:408-423.
- **Scenario:**
  - The job snapshots the store at T0, then syncs the tree until T1. The first run takes
    hours. Uploads, moves, and removes land in between, so the remote tree reflects roughly
    T1 against a T0 catalog.
  - Snapshots are dated objects, but the tree on the remote is only ever the latest, with
    superseded files under the archive path. Restoring any snapshot older than the latest
    pairs it with a newer tree. Its moved and removed files sit in the archive, under paths
    the catalog no longer names.
  - The spec says nothing about what restore does when the halves disagree.
- **Fold:**
  - Take the store snapshot after the tree sync completes, so the catalog is never older than
    the tree. Admissions landing during the sync then show up as files-without-rows, which is
    recoverable, rather than rows-without-files.
  - Record the run id and the tree's completion time in the snapshot's metadata.
  - Add a post-restore reconciliation report (rows without files, files without rows) that
    must run before serve opens (finding 4).
  - State that only the latest snapshot pairs with the tree.
- **OWNER FORK:** Should an older dated snapshot be restorable at all, and with archive-path
  recovery? Or is only the latest pair supported?

### 6. The empty-root guard misses the likely real failure (major)

- **Spec:** R6:417-420, R6:585-586.
- **Scenario:** The library filesystem is unmounted, or a wrong mount sits at the library root,
  and a handful of files exist underneath (a `lost+found`, or a stale partial tree). The root
  exists and is non-empty, so both guards pass. `rclone sync` then archives nearly the whole
  remote. The archive makes this recoverable only while the remote's lifecycle policy (R6:51)
  has not pruned it, and nothing alerts the operator.
- **Evidence:** The store knows how many tracks the catalog holds, which is a far better
  oracle than emptiness.
- **Fold:**
  - Abort when the tree's file count is below a threshold fraction of the catalog's track
    count. Make the fraction a store-homed tunable.
  - Pass rclone `--max-delete`, sized from the same count.
  - Require that the library root is on the device `doctor` expects.
  - Have the runbook state that archive retention must exceed the backup-failure detection
    window.

### 7. Backup omits revocation state, so a disaster restore can resurrect revoked sessions (major)

- **Spec:** R6:44-51, R6:157-159 (R18 keeps nonce and epoch records as files).
- **Scenario:** The box is lost and restored to a new VPS from the store snapshot and the tree.
  The auth leaf's epoch records are not in the backup. The session and link keys come back from
  the age store unchanged, so every session cookie of an offboarded member verifies again.
  After 6d the epochs are keyed on surrogate ids, but the restored store holds those ids with
  no epoch records beside them.
- **Fold:** Either include the auth leaf in the backup set (it is small), or require the
  disaster-restore runbook to rotate the session and magic-link secrets, which invalidates
  every outstanding credential. The second fold is the simpler and stronger one. Name the
  owner: 6b's runbook.

### 8. Any writing verb migrates, under a shared lock (major)

- **Spec:** R6:197-198, R6:204-206, R6:243-259, R6:592-595.
- **Scenario:**
  - (a) A deploy installs binary N+1 while serve N keeps running until restart. The backup
    timer fires with N+1, opens the store, and migrates it under the live serve N. Serve N
    checked `user_version` only at its own open, so it now writes against a schema it does not
    know: a column rebuilt, a NOT NULL added, a trigger it cannot satisfy.
  - (b) Two N+1 processes open at once. If the runner reads `user_version` outside the
    `IMMEDIATE` transaction, the second process re-applies a migration the first committed.
  - (c) Both processes write the pre-migration snapshot. The second one's snapshot is of the
    post-migration state, and it overwrites the only pre-migration copy.
- **Fold:**
  - Migration requires the store lock held exclusive. The lock design already has an
    exclusive mode (R6:197-198).
  - Read `user_version` inside the same `IMMEDIATE` transaction that applies the migration.
  - Only `serve`, or an explicit `migrate` verb, migrates. Every other verb refuses a store
    whose version is not its own, naming both versions.
  - Add a two-process test to 6a's exit gate.

### 9. A single overwritten pre-migration snapshot does not survive a bad migration's next upgrade (major)

- **Spec:** R6:69-74 (ruling 4), R6:251-252.
- **Scenario:**
  - (a) Migration M5 commits cleanly but is semantically wrong (a bad `UPDATE`). A week later
    binary N+2 migrates M6. The runner snapshots the corrupted post-M5 store over the only good
    pre-M5 copy.
  - (b) Migrations 5 through 7 run in one open, 5 commits, and power fails. The next open
    re-snapshots state 5 over state 4.
  - (c) Recovery: `restore` of the local snapshot under the same binary is followed by
    "the first open migrates the restored store forward" (R6:431-432), which re-runs the bad
    migration immediately.
  - Rolling back to an older binary instead hits the newer-store refusal, and the only way
    around that is restoring the snapshot, losing every write since the upgrade. The spec
    never states that cost.
  - The snapshot write is also not stated as atomic, so a crash mid-write can leave neither
    the old nor the new snapshot.
- **Fold:**
  - Name each snapshot by its from-version (`pre-migrate-v4.db`). Never overwrite a snapshot
    of a lower version with a higher one. Keep a snapshot until a successful backup job
    postdates it, or keep the last K.
  - Write it through staging, rename, and dir fsync.
  - Refuse to migrate if the snapshot cannot be written, with free space checked first
    (finding 11).
  - Let `restore` accept a local pre-migration snapshot, with a `--no-migrate` path that
    requires the matching older binary.
  - Add a rollback section to the runbook that states the lost-writes cost.
  - This amends ruling 4, which is already an owner question.

### 10. `synchronous` is "explicit" but unnamed, and NORMAL breaks the intent-journal pattern (major)

- **Spec:** R6:226-229, R6:596-599.
- **Evidence:** sqlite.org/pragma.html: "A transaction committed in WAL mode with
  synchronous=NORMAL might roll back following a power loss or system crash."
- **Scenario:** CH's move verb commits the catalog change and its intent row, then does the
  filesystem work with directory fsyncs, then removes the source (CH:451-458). Under NORMAL, a
  power cut can roll back the committed intent while the renames it authorized are already
  durable. Replay finds no intent, so the files sit at the target with no catalog row, and the
  source is gone. The same class hits the roster import (surrogate ids lost while epoch
  records were already renamed onto them) and any job record.
- **Fold:** Pin `synchronous=FULL` in the spec, with the reason: file work follows commits.
  A DSN test asserts the pragma's value on a second pooled connection. FULL also satisfies
  `VACUUM INTO`'s sync condition.

### 11. Disk-full is unaddressed, and rung 6 adds several full copies of the store to the disk (major)

- **Spec:** R6:193, R6:408-415, R6:251-252. CH:494-497: "a full disk wedges the store's
  write-ahead log and takes the server down."
- **Scenario:**
  - The store leaf can hold the database, its WAL, a pre-migration snapshot, a
    `VACUUM INTO` staging copy, a restore download, and a pre-restore set, all on one
    filesystem. A backup job taking its snapshot is enough to cross the free-space floor CH
    sets for uploads, and wedge serve.
  - The WAL can grow without bound. sqlite.org/wal.html: "if … there is always at least one
    active reader, then no checkpoints will be able to complete and hence the WAL file will
    grow without bound". It also says a checkpoint "does not normally truncate the WAL file
    (unless the journal_size_limit pragma is set)."
  - `VACUUM INTO` holds a read transaction for its whole run.
- **Fold:**
  - Check the free-space floor before every full copy: at least db plus WAL size plus
    margin, and refuse by name.
  - Set `journal_size_limit`.
  - Run `wal_checkpoint(TRUNCATE)` after each snapshot.
  - Have `doctor` report WAL size and free space.
  - Map `SQLITE_FULL` to a named error.
  - Consider a small ballast file in the leaf that an operator can delete to un-wedge the
    store.
  - Route the upload floor to rung 7 as CH already does, and name 6a and 6b as owners of
    the store's own copies.

### 12. The FK-off migration connection goes back to a one-connection write pool (major)

- **Spec:** R6:253-257, R6:600-601.
- **Scenario:** The rebuild migration runs `PRAGMA foreign_keys=OFF` on "one dedicated
  connection" (R6:255-256). If that connection comes from the pool, or is returned to it, and
  a later path fails before the pragma is turned back on, the write pool's only connection
  enforces nothing for the life of the process. Every later write skips referential checks,
  which is the "Foreign keys silently off" risk the spec lists (R6:600).
- **Evidence:** sqlite.org/lang_altertable.html step 12 re-enables the pragma explicitly. A
  pooled `*sql.Conn` keeps per-connection pragma state.
- **Fold:** Run rebuild migrations on a separate `*sql.DB` that is closed afterward, or
  discard the connection (`driver.ErrBadConn`). Add a test: after a rebuild migration, an
  orphan insert through the write pool fails.

### 13. Job recovery marks live non-exclusive jobs interrupted (major)

- **Spec:** R6:363-367, R6:640.
- **Scenario:** Only exclusive kinds hold a lock (R6:364-366). A non-exclusive job (any
  operation rungs 7, 9, and 13 add) is running in serve. A `backup` CLI process opens the
  store and, "on open," marks every running record with no held lock as interrupted. That
  includes serve's live job. When the job finishes, it either violates "A terminal state never
  changes" (R6:352) or loses its result.
- **Fold:** Each process holds a per-instance lease lock (a lock file named by an instance id,
  held for the process's life). Job records carry the owner instance. Recovery marks only
  records whose owner's lease is free. The interruption test plants a non-exclusive running
  record owned by a live second process.

### 14. The epoch re-key loses revocations under rollback, concurrency, and re-import (major)

- **Spec:** R6:519-525, R6:604, R6:645.
- **Scenario:**
  - (a) Rollback. The re-key completes and writes its marker. The operator rolls back to the
    5d binary, which bumps by address digest when offboarding a member (T3), then rolls
    forward. The reader consults the new key (T1) and no longer the old key after the marker.
    A session anchored between T1 and T3 passes.
  - (b) Concurrency. `EpochStore.mu` is in-process only (`internal/auth/epoch.go:66-70`), and
    `advance` overwrites through `atomicfile.WriteStream`. If a `member` CLI process bumps the
    new key while serve's re-key copies the old record over it, the stored epoch moves
    backward. `ErrEpochBackward` is not checked across processes.
  - (c) Re-import. Restoring a pre-6d store snapshot, or any path that rebuilds `users`, mints
    new random ids (R6:509-510, invariant 6). Every re-keyed record is orphaned, and every
    revocation is lost.
- **Evidence:** "the reader consults the new key first and the old one second until a
  completion marker is written" is a first-found read, not a maximum.
- **Fold:**
  - Revocation checks read `max(epoch[id], epoch[address])` permanently. Keep old-key records
    and drop the completion marker's power to stop consulting them. The records are a few
    bytes each.
  - The re-key writes only when the new value is later than the stored one, and runs under
    the store lock held exclusive.
  - Add the rollback and re-import cases to the four-state test (R6:645).

### 15. Tombstones and address reuse can revive a revoked session (major, OWNER FORK)

- **Spec:** R6:284-286, R6:519-521.
- **Scenario:** Member A is offboarded, so `users` gets a tombstone and the epoch is bumped
  under A's surrogate id. A is later re-added with the same address. If re-adding mints a new
  id, a stolen pre-offboarding cookie for A resolves to the new id, because sessions carry
  `Address` only (`internal/auth/session.go:71-90`). The new id has no epoch record, so the
  revoked cookie verifies. The spec does not say whether the address index is unique across
  tombstones or only across live rows.
- **Fold:** State the index semantics (a partial unique index on live rows, or global).
  Finding 14's max-with-address read closes the security half either way.
- **OWNER FORK:** Is a re-added member the same person, reactivating the same id with their
  stars and history? Or a new person with a new id?

### 16. Library-relative track paths make rung 16's relocation a rewrite of every row (major)

- **Spec:** R6:294-295 (the track's "library-relative `path`"), R6:306-312, R6:304 (the
  `move_intents` contents are unspecified). Erratum at BS:131-141 and CH:662-664.
- **Scenario:** Relocating a nested collection "in one move" needs a file move and a rewrite
  of every track row in that collection. The rows and the files cannot share a transaction.
  The default collection is rooted at the library root, so its own albums cannot move in one
  rename at all. They are thousands of per-directory renames. A crash between the renames and
  the row rewrite leaves the catalog naming paths that no longer exist. The relocation also
  makes the next backup re-upload the whole library and archive the whole old copy, because
  rclone sees a delete plus a new file without `--track-renames`.
- **Fold:**
  - Store track paths relative to their collection's `root_path`, so relocation updates one
    row and moves one subtree.
  - Fix `move_intents`' contract now: old and new paths as BLOB, a granularity (file or
    directory), and a state, with an idempotent replay rule ("source absent and target
    present" means done).
  - Name rung 16's relocation as a consumer of `move_intents`.
  - Name `--track-renames`, or a post-relocation backup plan, in 6b's backup contract.

### 17. Seeds applied row by row fail valid combined changes and can partially commit (minor)

- **Spec:** R6:453-471, R6:488 (the correlated retention rule).
- **Scenario:** An operator raises `DUBPLATE_MAGIC_LINK_EXPIRY` to 2h and
  `DUBPLATE_NONCE_RETENTION` to 3h in one deploy. Seeding the expiry first validates it
  against the stored retention of 1h and fails the load. With one transaction per seed,
  earlier seeds commit before a later one fails.
- **Fold:** Apply every seed at boot in one `IMMEDIATE` transaction, and validate correlated
  rules against the post-seed set.

### 18. Every verb with its own env re-seeds, so settings flap (major)

- **Spec:** R6:453-455, R6:463.
- **Scenario:** Units read different env files today: serve and backup use `/etc/dubplate/env`,
  and shadow uses `/etc/dubplate/server.env` (`systemd/*.service` `EnvironmentFile=`). An
  operator can also run `member` from a shell with a partial env. If every writing verb applies
  the seed rule, two units with different values for one key re-seed each other on every run,
  each change logged and audited. A value that differs at "env-seed" re-seeds (R6:463), so this
  never settles.
- **Fold:** Only `serve` applies the seed rule. Every other verb evaluates it read-only, as
  `check` does (R6:454-455), and warns on disagreement.

### 19. A removed env key and a changed default both leave stale values in force (minor, OWNER FORK)

- **Spec:** R6:459-463.
- **Scenario:** An operator deletes `DUBPLATE_SESSION_LIFETIME` expecting the default back. The
  row keeps the old seeded value forever, and only `doctor` mentions it. That runs against
  CH's "never silently ignored" intent (CH:85-86). Separately, a `default` row stores its value,
  so when a later binary changes the researched default, every existing install stays pinned
  to the old one.
- **Fold:** A `default` row stores no value and resolves from the registry at read time.
- **OWNER FORK:** When an env-seeded key is removed, should the row revert to the default with
  a log line, or should the load fail?

### 20. Rollback and roster divergence are unstated (minor)

- **Spec:** R6:248-250, R6:91-101.
- **Scenario:** Rolling back past 6d brings back a binary that reads the roster file, stale
  since the import, and writes it through web add-member. Rolling forward again, the store
  ignores those additions: `users` is populated, so the key is a load error with no import
  path for the new rows.
- **Fold:** Finding 1's file digest detects the divergence and names it. Add a runbook line:
  rollback past 6d requires re-adding members through the verb afterward, or a
  `member import --merge` verb.

### 21. The write-function contract leaves partial commits and stale-read writes open (minor)

- **Spec:** R6:231-233.
- **Evidence:** sqlite.org/lang_transaction.html section 3: for `SQLITE_FULL`,
  `SQLITE_IOERR`, `SQLITE_INTERRUPT`, and `SQLITE_NOMEM`, "SQLite attempts to undo just the one
  statement … and continue with the transaction". It "might" instead roll back the whole
  transaction.
- **Scenario:** A callback swallows one statement's error, and the function commits a partial
  transaction. Context cancellation in modernc raises `SQLITE_INTERRUPT` the same way. A
  uniqueness decision read from the read pool, then written in the write pool, races another
  writer. `AddMember`'s duplicate re-check is exactly this shape once it moves to the store.
- **Fold:**
  - The function rolls back on any error the callback returns or any statement error it saw,
    and treats "already rolled back" as that error.
  - Every read a write decides on runs inside the callback.
  - The read pool opens with `query_only`.

### 22. The spec's atomicfile claim is inaccurate for the snapshot path (minor)

- **Spec:** R6:596-599, citing `atomicfile.go:45-55`.
- **Evidence:** `atomicfile.WriteFile` does not fsync the directory. Only `WriteStream` does
  (`internal/atomicfile/atomicfile.go:82-89`), and `CreateRoster` uses `WriteFile`. A
  `VACUUM INTO` output is a file SQLite writes itself, so it cannot go through either entry
  point.
- **Fold:** State that the snapshot's rename and restore's renames fsync the directory by hand,
  or add an `atomicfile` helper that renames an existing file durably. Clean leftover staging
  temp files at open.

### 23. Migration identity is `user_version` alone (minor)

- **Spec:** R6:258-259.
- **Scenario:** Lanes renumber migrations at rebase. A store once opened by an unmerged
  worktree build, such as a dev or shadow store, sits at version 7 with a different "7" than
  master's. Nothing detects the mismatch.
- **Fold:** Record each applied migration's name and content hash in a table inside its
  transaction, and refuse on mismatch.

### 24. `check` and `doctor` read without the lock, and opening a WAL store read-only is not side-effect-free (minor)

- **Spec:** R6:197-198, R6:204-206, R6:453-455.
- **Evidence:** sqlite.org/wal.html: a read-only WAL open needs `-shm` and `-wal` to already
  exist, or a writable directory to create them in. Opening a WAL store therefore creates
  files even for `check`.
- **Scenario:** `restore` can rename the database under a running `check` or `doctor`, because
  neither holds the shared lock.
- **Fold:** Both verbs take the shared lock and open with `mode=ro`. State that they may create
  `-shm`.

### 25. Foreign-key delete actions are unspecified (minor)

- **Spec:** R6:268-278, R6:284-286.
- **Scenario:** A later migration declares `ON DELETE CASCADE` from `users` or `collections`,
  and a hard delete then erases history and audit rows that CH requires to survive as
  tombstones (CH:519-521).
- **Fold:** Add invariant 7: references to `users` and `collections` declare no cascade. The
  test reads `on_delete` from `pragma_foreign_key_list`.

## Counts

Blockers 2 (1, 3). Majors 15 (2, 4-16, 18). Minors 8 (17, 19-25). Owner forks: 2, 5, 15, 19.
