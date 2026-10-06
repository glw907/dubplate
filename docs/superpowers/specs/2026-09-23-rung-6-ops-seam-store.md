# Rung 6: the ops seam and the store

Status: **revision 2.2, ratified by Geoff 2026-09-23**: all fourteen owner rulings below are ruled
yes, as recommended. The errata this spec owes land in the documents they name. Its review history, every finding's disposition, and the arguments this spec only states
are in `research/2026-09-23-rung-6-spec-fold-record.md`. Erratum 2026-10-06: a one-time exception to the
append-only rule for `0003_catalog.sql` (see the erratum block under "Migrations").

Rung 6 builds dubplate-server's library store and its operation vocabulary. The store is one
SQLite file on the pure-Go modernc driver. It runs in WAL mode with `synchronous=FULL` and one writer
per process. Only `serve` creates or migrates it, under an exclusive lock and behind a
pre-migration snapshot. A schema-invariant test binds every later rung's migrations to rules
SQLite's own catalog can decide: path columns as enforced bytes in `STRICT` tables, entity ids that
are never rowids, timestamped event tables, and foreign keys that resolve and never cascade.

The ops seam is one Go interface with JSON-serializable request and response types. Its job model
takes liveness from a per-process lease and carries cancellation across processes through the
store. Its operation catalog records availability and required privilege, so the organizer spec's
parity rule has a mechanism from the first version. Backup and restore are the first operations.
A backup syncs the library tree, then snapshots the store, to any rclone remote behind an
integration flag, and prunes its own history. Restore is offline: it swaps one file under a
restore-intent record and reconciles the catalog against the tree before `serve` may open.

Rung 6 also lands the schema the ratified configuration homes spec routes to it and the settings
home, with an env-seed rule that only `serve` applies. Once rung 5 closes, it lands the surrogate
user id, an explicit roster import, and `serve`'s own use of the store. The ladder drew it as band
M; it is five passes (band L).

Written against the ratified build spec (`specs/2026-09-02-greenfield-server-build.md`, BS), the
organizer spec (`specs/2026-09-01-dubplate-organizer-design.md`, OS), the configuration homes spec
(`specs/2026-09-11-configuration-homes-design.md`, revision 3, ratified 2026-09-23 at `fd122d9`,
CH), the 2.0 streamer spec (`specs/2026-09-07-streamer-2-0-design.md`, revision 2.2, S20), the
parallel-build approach (`specs/2026-09-03-parallel-build-approach.md`, PA), the rung 5 spec
(`specs/2026-09-08-rung-5-identity-auth.md`, R5S), the beets evolution lessons
(`research/2026-09-02-beets-evolution-lessons.md`, EVO), and the reverse-edge completion
(`research/2026-09-04-reverse-edge-completion.md`, RE). Tree citations were read at `7597052`.

Every passage that exists because CH routes it here carries the mark **[CH]**. Every change to a
ratified document is an erratum line in "Errata this spec owes." Every mechanism the spec relies on
is quoted from its documentation, backed by a recorded measurement, or listed in "Not verified in
this revision."

---

## Owner rulings

**Ruled, Geoff, 2026-09-23: yes to all fourteen, as recommended.**

Fourteen yes-or-no questions, ordered by how much build each answer changes. Thirteen change what
an adopter sees, configures, or is asked for; ruling 2 changes only when schema lands. Each carries a
recommendation, one reason, and what it changes. The mechanics behind each live in the body.

**1. Does the roster move into the store through a one-time `member import` command, after which
`DUBPLATE_ROSTER_FILE` is a retired setting?** Recommendation: yes. Reason: members are store state
under CH (CH:67-70), and the configuration loader cannot see the store, so the file cannot stay a
live setting after the move. What it changes: the 6d upgrade is a deploy that drops the setting,
then `member import --dry-run`, then `member import`. No roster member can log in between `serve`'s
restart and the import. A member added through the CLI verb in that window merges rather than
blocking the import.

**2. Do S20's hook columns 1, 2, 3, and 11 land with the rungs that compute them, rather than as
nullable columns at 6c?** Recommendation: yes. Reason: CLAUDE.md's 2026-09-12 refinement, which
postdates S20, says a hook is built by its first consumer, so the audio key, file hash, sort key,
art hash, and gains arrive with rungs 7 and 9. What it changes: when those columns appear; nothing
an adopter sees. A yes owes S20 a dated erratum (S20:363-372, S20:392).

**3. Do the shadow floor and budget stay in the environment rather than moving to the store?**
Recommendation: yes. Reason: their only reader is the shadow unit, which has no access to the store
and retires at launch. What it changes: the shadow pair stays in the env template. A yes owes CH an
erratum naming this wall category.

**4. Does a backup contain the library tree and a store snapshot, with the streamer's own database
added by rung 12?** Recommendation: yes. Reason: the store only records where files live, so a
snapshot alone restores nothing, and at 1.0 the streamer's database is the only home of members'
stars and playlists (CH:461-462). What it changes: remote storage grows with the library. Until rung
12, a restore loses stars and playlists, which the runbook states.

**5. Is a backup target one on/off flag and one rclone remote name, with credentials in rclone's own
environment variables rendered from the age store?** Recommendation: yes. Reason: it works with any
rclone backend, and it keeps every secret in the age store, as CH requires (CH:52-57). What it
changes: an adopter sets `DUBPLATE_BACKUP_ENABLED`, `DUBPLATE_BACKUP_REMOTE`, and rclone's
`RCLONE_CONFIG_<NAME>_*` variables. The template rows for rclone's variables arrive at rung 16;
until then the runbook sets them by hand. The six R2-era settings retire.

**6. Is restore an offline command that refuses while the server runs?** Recommendation: yes.
Reason: replacing a database file under an open connection "results in behavior that is undefined
and probably undesirable" (sqlite.org/howtocorrupt.html section 2.5). What it changes: a restore
means stopping the server first.

**7. Before each upgrade, does the store keep a copy of itself, one per version, until a backup has
safely captured the upgraded store?** With backup off, it keeps the newest
`DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` copies. Recommendation: yes. Reason: a migration that commits
cleanly but is wrong needs a copy from before it, and that copy must survive the next upgrade. What
it changes: the state directory holds one store copy per upgrade not yet backed up, shown by
`doctor`. Rolling back uses the local copy while it survives, then a pre-upgrade remote snapshot
while ruling 10 keeps one.

**8. Does the roster import refuse a member with no streamer username rather than inventing one?**
Recommendation: yes. Reason: the username names the member's existing streamer account, and an
invented one creates a fresh account that orphans the member's history. What it changes: an adopter
fills the field before importing. `member import --dry-run` names every member missing one, and a
refused import leaves `serve` running with the login outage of ruling 1.

**9. Does re-adding a removed member give them back the same identity?** Recommendation: yes.
Reason: stars and history are keyed on that identity, and the address is the member's login proof.
What it changes: a returning member gets their library state back. Either answer keeps a pre-removal
login refused.

**10. Does dubplate prune old remote snapshots and old archived files itself?** Recommendation: yes.
Reason: many rclone backends, `local` among them, have no lifecycle policy to prune for it. What it
changes: remote storage stays bounded with no provider configuration. The newest
`DUBPLATE_BACKUP_SNAPSHOT_KEEP` snapshots and `DUBPLATE_BACKUP_ARCHIVE_RETENTION` of archive stay;
the latest snapshot is never pruned.

**11. Is the latest snapshot the only supported way to restore the store and the library together?**
Recommendation: yes. Reason: the remote keeps only the latest library tree, so an older snapshot
pairs with a newer tree. What it changes: an older snapshot restores the store alone, behind a flag,
with a report of every file the tree no longer holds.

**12. After a disaster restore, are the session and magic-link secrets rotated instead of backing up
the login records?** Recommendation: yes. Reason: rotation invalidates every outstanding credential,
so revocations made after the last backup hold too, which restored login records would miss. What it
changes: every member logs in again after a disaster restore.

**13. When an operator removes a seeded setting from the environment, does the setting return to its
default, with a log line?** Recommendation: yes. **[CH]** Reason: an operator who deletes a setting
expects the default, and keeping the old value silently is what CH forbids (CH:85-86). What it
changes: deleting a setting from the age store is a supported way back to the default. The
alternative refuses to start `serve` until the operator restores the setting or accepts the default.

**14. Does `doctor` warn when the backup remote is not encrypted with an rclone `crypt` remote?**
Recommendation: yes, a warning and never a refusal. Reason: after 6d the snapshot holds member
addresses, but a refusal would fail adopters whose provider encrypts at rest. What it changes: the
runbook recommends a `crypt` remote, and `doctor` says when one is absent.

**Not asked: rung 9's concurrency model** (BS:760-764). A lease or a queue sits above the store's
serialized writers. The store's two constraints on rung 9: no write transaction spans a MusicBrainz
lookup or a file copy, and library mutations continue during a backup's sync unless rung 9's lease
pauses them, so restore's reconciliation report is the net.

---

## What rung 6 builds

| Deliverable | Pass | Proved by |
| --- | --- | --- |
| `internal/engine/store`: the open (inspect, write, serve modes), the leaf, the lock, three connection classes, the write method, migrations, snapshots | 6a | 6a criteria 1-16 |
| The migration set (ledger, job records, catalog skeleton) and the schema-invariant test | 6a, then 6c | 6a 8-13, 6c 1 |
| `doctor`'s store row and `check`'s inspect report, the open's first production callers | 6a | 6a 15 |
| `internal/ops`: interface, wire types, error codes, privileges, the catalog, the job model | 6b-i | 6b-i 1-8 |
| `internal/ops/local`: the runner, leases, recovery, cross-process cancel; the write method's first caller | 6b-i | 6b-i 3-8 |
| `internal/backup`: the snapshot job, the rclone transport, the guards, the layout, pruning | 6b-i | 6b-i 9-17 |
| The backup integration class and its supersessions; the `backup` verb; `doctor`'s backup rows | 6b-i | 6b-i 18-21 |
| `internal/ops/opshttp` (unexported) with its authorizer, and the contract suite | 6b-ii | 6b-ii 1-4 |
| The `restore` verb: the intent record, the one-file swap, library restore, reconciliation | 6b-ii | 6b-ii 5-13 |
| Every table CH routes to rung 6 **[CH]** | 6c | 6c 1-3 |
| Registry metadata: home, apply mode, `sensitive`, bounds, correlated rules **[CH]** | 6c | 6c 4-5 |
| `internal/settings`: accessor, generation, seed rule, audit; 6b-i's constants re-homed **[CH]** | 6c | 6c 6-14 |
| The import verb, the users-backed roster, `Identity.UserID`, the dual-key epoch | 6d | 6d 1-8 |
| The auth thresholds re-homed with accessor reads at every call site | 6d | 6d 9-10 |
| `serve` opening the store: migrate, seed, then listen | 6d | 6d 11-12 |
| The store leaf's deployment: 5d's serve unit's grant, `deploy.sh`, the state-dir leg, the template | 6d | 6d 13-14 |

## What rung 6 does not build

- No HTTP route and no `ops/client`. The projection stays unexported until rung 14 owns the
  authorization model (BS:831-837) and mounts it on CH's cookie and bearer mounts (CH:107-110); the
  authorizer takes the mount's channel. The client is post-launch (BS:668-672).
- No Go accessor for a table whose first caller is a later rung: rungs 7, 9, 13, and 14 add them,
  and the root-allocation function and no-nesting check land at rung 7 or 9. 6b's guard and
  reconciliation read the skeleton through unexported queries in `internal/backup`.
- No hook column a later rung computes (ruling 2), and no `collections` row version for CH's
  `if_version` check (CH:256-257), whose first writer is rung 14.
- No default collection row: the first rung that writes a collection (rung 7 or rung 10) creates
  it, owned by the first admin.
- No backup schedule, schedule key, or rclone template rows: rung 16 lands them with the timer
  (CH:56; erratum 14). No backup pings: rung 9's monitoring seam (`internal/config/keys.go:93-100`)
  reads the job's terminal state.
- No runbook: rung 16 authors it (CH:661-662), and each pass records its owed lines at its close
  (restore size, rollback cost, rotation, crypt, rclone variables by hand, roster file deletion, the
  empty-catalog guard gap).
- No re-authentication gate or sensitive-change notice (rung 14, CH:657-660), and no URL-valued
  store-homed key; rung 14's range check attaches at the write path's validator (CH:132-133).
- No state leaves rung 4's or rung 5's leaves except the roster (ruling 1); epoch and nonce records
  stay files (R5S:533-539). 5c builds the last-admin guard, and 6d re-establishes it in the write
  transaction. After 6d `users.admin` is authoritative, and reconciling `DUBPLATE_ADMIN_EMAIL` is
  rung 14's (BS:833).
- Nothing creates or migrates the store in production before 6d. `doctor` and `check` open it
  read-only, `backup` refuses an absent store, and `restore` can create one only from a snapshot,
  which cannot exist before 6d. `serve` opens the store at 6d, the pass that grants its leaf, so a
  5d deploy from a master carrying 6a through 6c never touches an ungranted leaf.

---

## Package territory

Every package rung 6 creates is new, so its lanes write no file another lane owns except the rows
"Contested surfaces" names.

| Package | Role | Imports it may take |
| --- | --- | --- |
| `internal/engine/store` | the database, connections, migrations, snapshots, leases | `atomicfile`, the driver |
| `internal/ops` | interface, wire types, job model, catalog, error codes | `identity` only |
| `internal/ops/local` | in-process implementation and job runner | `ops`, `store`, `backup`; `settings` from 6c |
| `internal/ops/opshttp` | the projection and the contract suite | `ops` |
| `internal/backup` | snapshot job, rclone transport, guards, pruning, restore, reconciliation | `store`, `atomicfile` |
| `internal/settings` **[CH]** | accessor, seed rule, audit | `config`, `store` |

`internal/ops` imports no server package (RE:116), the engine has no HTTP knowledge (BS:375), and
`internal/config` never imports `store` or `settings`. At 6d, `internal/auth` gains one edge, to
`store`, for the users-backed roster behind the roster's existing constructor. It never imports
`settings`: `cmd/dubplate-server` injects each auth threshold as an accessor function. The store
itself never reads a setting; a caller passes it any store-homed value it needs.

Exported surface follows its first caller. 6a exports the open, whose production callers are
`doctor` and `check`, and tests the write method and the serve mode in-package. 6b-i exports the
write method with `ops/local`, its first caller. 6d exports the serve mode and the seed rule with
`serve`, their first caller.

**Erratum 2026-09-24 (6a plan approval, E3):** "6a exports the open" means the inspect mode,
its options (the state directory and the busy timeout), and its report type with a closed
state set. The handle, the write mode, the serve mode, the write method, and the migration
runner stay unexported until their own first caller.

---

## The store

### Layout, ownership, and the lock

The store lives under one leaf of `DUBPLATE_STATE_DIR` named `store`, following the `shadow` and
`auth` leaves (`internal/engine/shadow/store.go:39-49`, `internal/auth/store.go:12-40`). The leaf
holds `store.lock`, a `db/` directory with the database and its `-wal` and `-shm` files,
`snapshots/` (pre-migration and pre-restore copies), `staging/`, `leases/`, and at most one
`restore-intent` record. The lock file sits outside `db/` and nothing ever renames it, so every
process locks the same inode. The leaf is 0700. The mechanics review measured that SQLite creates
the database and a `VACUUM INTO` target at 0644 under a 0022 umask, and gives `-wal` and `-shm` the
main file's mode. So the store pre-creates the database file and every snapshot target with
`O_CREATE|O_EXCL` at 0600; SQLite permits an empty `VACUUM INTO` target ("must not previously exist,
or else it must be an empty file," sqlite.org/lang_vacuum.html).

Every writing open checks for a `restore-intent` record first. Only then does it remove staging
leftovers, and it never removes a file an intent record names.

**Erratum 2026-09-24 (6a plan approval, E7):** only serve mode, holding the exclusive lock with
no intent record present, removes staging leftovers. The write mode never does; `backup` and
every `member` verb stage their own files under a shared lock, and a write-mode cleanup would
remove a live one.

Every verb that opens an existing leaf refuses when its effective uid differs from the leaf's owner,
naming the owner and the fix (`runuser -u <owner> -- dubplate-server ...`); a root-run verb would
otherwise leave a root-owned `-shm` or database that `serve` cannot open. An absent leaf reports
absent with no uid check. After 6d, `deploy.sh`, which connects as root (`scripts/deploy.sh:114`),
runs every store-opening verb through `runuser -u <owner>`.

**Erratum 2026-09-24 (6a plan approval, E9):** a create also requires the effective uid to own
the state directory, and otherwise refuses naming that owner before anything is created.

The lock is an advisory `flock` on `store.lock`, released when the process exits:

| Verb | Lock |
| --- | --- |
| `backup`, every `member` verb (`import` included) | shared |
| `restore` | exclusive |
| `check`, `doctor` | shared, non-blocking |
| `serve` | the procedure in "Migrations" |

`flock` cannot name a holder, so a refusal says "held by another process" and lists the pids
`/proc/locks` shows for the lock file's inode when that file is readable.

WAL needs every process on one host (sqlite.org/wal.html section 1). The store probes the leaf's
filesystem with `statfs`, refuses NFS, CIFS, SMB2, and 9p, and warns on FUSE, because sshfs and
rclone mounts are remote while mergerfs is local and common on home servers.

**Erratum 2026-09-24 (6a plan approval, E6):** the probe also refuses SMB, Ceph, and AFS
(`SMB_SUPER_MAGIC`, `CEPH_SUPER_MAGIC`, `AFS_SUPER_MAGIC`, and `AFS_FS_MAGIC`, four constants),
on the same multi-host ground. It probes `db/` when it exists, else the leaf, else the leaf's
parent, since `statfs` follows a symlinked `db/` to its target.

Rung 6 declares three test seams. Each exists because no unit test can produce the real condition:

- The `statfs` probe is a package variable, because no test can mount NFS.
- The migration runner's source is an `fs.FS` parameter of an unexported function, ordinary
  dependency injection, because 6a's embedded set holds no failing or rebuilding migration.
- Restore's step hook is an unexported function variable called between flow steps. A test sets it
  to exit a re-executed test binary at a named step, so each crash is a real process death.

### Three open modes

The **inspect mode** is read-only (`mode=ro`), takes the lock shared and non-blocking, never
creates, and never migrates. It reports absent, restore in progress, older (naming the pending
migrations), newer (naming both versions), or current. A read-only WAL open needs "the -shm and -wal
files already exist and are readable" or "write permission on the directory containing the database
so that the -shm and -wal files can be created" (sqlite.org/wal.html section 5). So the invariant it
keeps is that the database file's bytes and mode never change; `-wal` and `-shm` may appear, owned
by the leaf owner. `check` and `doctor` use it.

The **write mode** never creates a store. A write-mode open of an absent store refuses and names
"start serve first," so only `serve` ever creates or migrates. A write-mode open that finds a
`restore-intent` record refuses and names `restore --resume`. A write-mode open refuses a store whose
version is not the binary's, naming both versions and "start serve to migrate." `backup` and the
`member` verbs use it.

The **serve mode** is the only mode that creates a store or migrates one (from 6d). It accepts one
intent state, `reconcile-pending`, which "The restore verb" describes.

**Erratum 2026-09-24 (6a plan approval, E11):** holding the exclusive lock, serve mode first
reads `snapshots/` before it creates a store or completes an interrupted create (a zero-byte
database with no `-wal`, or a header-only unmarked one). With a `pre-migrate-v<N>.db` or
`pre-restore-*.db` present, it refuses naming the newest copy and `restore --from-file`, leaves
the database file unchanged, and inspect reports it as refused. A lost database with no copy
beside it is still created or completed empty; the guard covers only a leaf that kept a copy.

### Connections and transactions

2.0 hook 9 asks for one writer, WAL with a `busy_timeout`, and a seam a later projector joins so
its delta commits with its event (S20:387-389). The store opens three connection classes over one
file:

- The **write pool** holds one connection with `_txlock=immediate`, so every write transaction
  takes the write lock first and the busy timeout governs the wait.
- The **read pool** opens with `_query_only=1`, so a write there fails at once with a read-only
  error, and no caller writes outside the write method. Without it, a write on a read connection
  would upgrade a deferred transaction. SQLite skips the busy handler on that upgrade (measured by
  the mechanics review: `SQLITE_BUSY` in 0s with a 2000ms timeout), so the write would return
  `SQLITE_BUSY` under contention instead of waiting.
- The **maintenance connection** is a separate `*sql.DB` for one snapshot or foreign-keys-off
  rebuild, closed afterward so it never returns to a pool. Its DSN carries `_synchronous=FULL`,
  `_busy_timeout`, and `_foreign_keys=1`. `VACUUM INTO` needs it: "A VACUUM will fail if there is an
  open transaction on the database connection that is attempting to run the VACUUM"
  (sqlite.org/lang_vacuum.html), it fails on `query_only` (measured by the mechanics review), and on
  the write pool it would stall every in-process writer.

Every per-connection setting rides the DSN, because `database/sql` opens connections at will, as
modernc's shorthand keys, which the driver validates before applying any: `_busy_timeout`,
`_foreign_keys=1`, `_journal_mode=WAL`, `_synchronous=FULL`, and `_query_only` or `_txlock` per
class (`sqlite.go:307-391` at v1.59.0). `journal_size_limit` has no shorthand and rides `_pragma`,
which the driver executes verbatim and cannot pre-validate (`sqlite.go:305-306`), so 6a proves every
setting reads back on a fresh connection of every class.

`synchronous` is `FULL`. Under `NORMAL`, "a transaction committed in WAL mode ... might roll back
following a power loss" (sqlite.org/pragma.html), which breaks every pattern where file work follows
a commit: CH's move intent, the import record, a job record. `VACUUM INTO` fsyncs its output only
"if the PRAGMA synchronous setting of the original database is NORMAL or FULL"
(sqlite.org/lang_vacuum.html), and the setting that counts is the maintenance connection's. "The
checkpoint does not normally truncate the WAL file (unless the journal_size_limit pragma is set)"
(sqlite.org/wal.html section 6): `journal_size_limit` bounds the WAL after each checkpoint, and every
snapshot ends with `wal_checkpoint(TRUNCATE)` to return it to zero bytes.

The store exposes one way to write: a method that runs a callback inside one `IMMEDIATE`
transaction. That callback is hook 9's seam. Its contract:

- It commits only when the callback returns nil and no statement inside failed, and rolls back
  otherwise. SQLite may undo one statement and continue the transaction on `SQLITE_FULL`,
  `SQLITE_IOERR`, `SQLITE_INTERRUPT`, or `SQLITE_NOMEM` (sqlite.org/lang_transaction.html section
  3), so a swallowed statement error never commits. "Already rolled back" is that error.
- Every read a write decides on runs inside the callback, never on the read pool.
- It requires a context with a deadline, since an in-process wait for the one write connection is
  bounded by the context and never by `busy_timeout`. A nested call fails fast.

  **Erratum 2026-09-24 (6a plan approval, E10):** measured and pinned by test: the busy timeout
  alone bounds the wait on another process's write lock, and a caller's context does not shorten
  it, because the driver discards an interrupted file-backed connection rather than cutting the
  busy handler's own wait.
- Its first statement reads `user_version` and fails closed on a value other than the binary's.
- `SQLITE_FULL` surfaces as a named disk-full error.

The busy timeout is a wall threshold with a researched default, declared in rung 6's own file.

### Migrations

2.0 hook 7 asks that the mechanism be chosen with an evolving schema in view (S20:383-384). The
runner is hand-rolled over an `fs.FS` of numbered SQL files and `PRAGMA user_version`, as S20's
prior-art row records (S20:435); production passes the embedded set, and tests pass synthetic sets.

`serve` never holds a store connection across a lock conversion. `flock(2)` says: "Converting a lock
(shared to exclusive, or vice versa) is not guaranteed to be atomic: the existing lock is first
removed, and then a new lock is established. Between these two steps, a pending lock request by
another process may be granted, with the result that the conversion either blocks, or fails if
LOCK_NB was specified." A waiting `restore` can win any conversion, so `serve` follows this
procedure:

1. Take the lock shared, open, and read the version. If it is current, keep the lock and continue.
   **Erratum 2026-09-24 (6a plan approval, E8):** this shared wait is bounded by
   `DUBPLATE_STORE_MIGRATION_WAIT` and logs the holders, as step 2 does.
2. Otherwise close every connection, release, and take the lock exclusive, waiting at most
   `DUBPLATE_STORE_MIGRATION_WAIT` (a wall threshold with a researched default) and logging the
   holders at an interval. On expiry, exit nonzero naming the holders.
3. Holding it exclusive, reopen, re-read the version (a restore may have run in the gap), and create
   or migrate.
4. Close every connection, convert to shared, and return to step 1, so a restore that won the
   conversion is met by a fresh open and a fresh version check.

Under `Type=simple` systemd "will consider the unit started immediately after the main service
process has been forked off" (systemd.service(5)), so no start timeout bounds that wait. The real
exposure is a backup holding the lock for a whole sync, hours on a first run, after a deploy that
brings migrations; 6d's `deploy.sh` runs the new binary's `check` first and, when the store is older
and the lock held, names the holder before restarting `serve`.

The runner's other rules:

- Each file applies in its own `IMMEDIATE` transaction that reads `user_version`, applies the file,
  formats the new version into `PRAGMA user_version = N` (a pragma takes no bound parameter), and
  records the file's name and SHA-256 in an applied-migration ledger table. A failed migration
  leaves the prior version, schema, and ledger.
- A ledger row whose name or hash differs from the binary's file at that version is refused, naming
  both; a store once opened by an unmerged worktree build is caught here.
- A store whose `user_version` exceeds the binary's highest migration is refused, naming both. This
  is CH's "laptop ahead" direction (CH:165-168), and it covers a newer snapshot restored under an
  older binary.
- A store is non-empty when `user_version > 0`. Before migrating a non-empty store the runner checks
  free space for one full copy plus margin, writes `pre-migrate-v<N>.db` through the maintenance
  connection into `staging/`, verifies it, renames it into `snapshots/`, and fsyncs the directory by
  hand (ruling 7). Any failure refuses the migration and names the snapshot path. With backup off,
  the runner then deletes all but the newest `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` copies, using the
  value its caller passes (the stored row or the registry default; a changed env value reaches it
  after the next start's seed).
- A rebuild migration that needs foreign keys off runs on the maintenance connection, because
  enforcement "may only be enabled or disabled when there is no pending BEGIN or SAVEPOINT"
  (sqlite.org/pragma.html). It runs `PRAGMA foreign_key_check` before commit, refuses to commit on
  any row, and closes the connection.
- Every migration file ships with a fixture that populates the schema as of its version, and the
  invariant test fails on a migration with no fixture. 6a embeds three migrations in order: the
  ledger (which records itself), the job records, and the catalog skeleton. Its populated-fixture
  proof migrates the job-records fixture forward through the skeleton.
- Files are append-only once merged. A lane that rebases onto another lane's migration renumbers its
  own, and the rebase is non-fast-forward, so merge ritual step 5 runs.

> Erratum, 2026-10-06 (rungs 12 to 16 re-plan spec, `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`,
> revision 2.1, Geoff, 2026-10-05, owner rulings 1 to 3 all yes): a one-time exception to the append-only
> rule above. `0003_catalog.sql` and its fixtures `0003_catalog.sql`, `0004_settings.sql`, and
> `0005_routed.sql` under `internal/engine/store/testdata/fixtures/` were edited in place to rename the
> `users.streamer_username` column to `username`, and for that rename only. The exception rests on two
> preconditions, both checked read-only against the box on 2026-10-06:
>
> - Doctor's `store` row reads `absent`: the state directory holds no store leaf, so no store was ever
>   migrated under the old `0003`.
> - Backup was never enabled (`server.env` carries no backup keys and no `dubplate-*` timer exists), so the
>   remote holds no store snapshot.
>
> Master commit `2f4d0524836d1f88d2377f7b6e29f7e5aec7ac39` made the edit and re-pinned `migrationHashes`'
> `0003_catalog.sql` entry to `42834b8f62df9727bea686bbc02030f1b9af0de2a6c61a59050be72399c7b7fd`. Any older
> store fails loudly at open: the runner compares each ledger row's sha256 against the embedded file and
> refuses the open on a mismatch. The append-only rule's text is unchanged, and no later edit in place
> follows from this one.

Rolling back to an older binary meets the newer-store refusal. While the local pre-migration
snapshot survives ruling 7, the way back is that binary's `restore --from-file`, which accepts a
newer current store ("The restore verb," step 3). After a backup has deleted it, the way back is
`restore --older --snapshot <stamp>` of a pre-upgrade remote snapshot, while ruling 10 still keeps
one. Either path loses every write since the upgrade, which the runbook states.

### Schema invariants

One test applies every migration to an empty database and to each migration's populated fixture,
then inspects the result through the catalog (`pragma_table_list`, `pragma_table_info`,
`pragma_foreign_key_list`, `pragma_index_list`), never by parsing SQL. Each rule is the catalog fact
the test checks, and one violating fixture per rule proves the rule can fail for its named reason:

1. Every table is `STRICT` (`pragma_table_list.strict = 1`). A column named `path` or ending in
   `_path` is declared `BLOB`. The mechanics review measured that an ordinary table stores a Go
   `string` bound into a `BLOB` column as TEXT, so a TEXT and a BLOB copy of the same bytes both pass
   a UNIQUE index, while a `STRICT` table refuses the bind.
2. Every foreign key resolves to an existing table and column, and its parent columns are the primary
   key or carry a UNIQUE index. The mechanics review measured that SQLite accepts a key to a
   non-unique parent and then fails every insert with "foreign key mismatch."
3. A table with a `kind` column, or whose name ends in `_events`, `_audit`, or `_journal`, carries
   `at` and `kind`, both `NOT NULL`. This is 2.0 hook 4 (S20:374-375).
4. MusicBrainz identifiers are named `mb_<entity>_id`. A column named `release_id`, `recording_id`,
   `artist_id`, `release_group_id`, `track_id`, or `work_id` that is not a foreign key is refused, the
   class of beets bug #604 in which a generic id column mixed sources (EVO:575-578).
5. No table pairs a foreign key column with a TEXT column named `name`, `key`, or `attr` and a column
   named `value`. That shape is the attribute bag U4 forbids.
6. An entity table's primary key is a `BLOB` column named `id`, and no column named `id` or ending in
   `_id` is an `INTEGER PRIMARY KEY` (S20:166-167). Keyed and singleton tables are exempt by a list
   named in the test: `settings` (keyed by key), the ledger (keyed by version), the settings
   generation (one row), and the group quota total (one row). A store test proves the id function
   mints 16 random bytes, never equal to a rowid.
7. No foreign key to `users` or `collections` declares `ON DELETE CASCADE`, since CH requires
   tombstones and audit history to survive (CH:519-521).

What a column means stays a reviewer's question: a path column not named as one, or a source
identifier under another name, passes the test. Every rung 6 migration task, and every later rung's,
carries that as a `diff-reviewer` checklist line.

### The schema rung 6 lands

Pass 6a lands the ledger, the job records, and the catalog skeleton (users, collections, albums,
tracks). The skeleton lands first because its first readers are 6b-i's store-aware backup guard and
6b-ii's restore reconciliation, and CH's routed tables hold foreign keys into users and collections.
Pass 6c lands CH's routed tables **[CH]**, because CH routes them to rung 6 by name (CH:637-645) and
puts the visibility column on the row at rung 6 (CH:195). Table names are working names a plan may
change; the contents are the contract.

- `users` (6a): the surrogate id (hook 5, S20:376-378); the normalized address, uniquely indexed
  across live and tombstoned rows under ruling 9; the display name; `streamer_username`, required
  and unique under ruling 8; the admin flag; a tombstone time (CH:519-521); timestamps. It stays
  empty until 6d.
- `collections` (6a): name; a server-allocated `root_path`, unique and never client-supplied;
  `owner`, a non-null reference to users; `visibility`, with a CHECK admitting both of CH's values,
  group and private, since rung 14's API is what refuses private while the driver lacks per-user
  folder access (CH:205-207); `writer` with the S20 hook 8 domain, this server or a named local
  binary (S20:385-386); the writer epoch; the mirror generation; the last-mirror time; the
  duplicate-collapse flag; the layout template; the tagging policy; and the admission overrides
  (CH:173-181, CH:637-642). Share rows are reserved in this spec only: no table, no column, no check
  (CH:535-538).
- `albums` and `tracks` (6a): opaque ids, the collection reference, the track's `rel_path` relative
  to its collection's `root_path` under U3, the MusicBrainz release and recording ids under
  invariant 4, the album-and-track relation, and timestamps.
- `settings` and `settings_audit` (6c): one row per store-homed key with its value, source, seed
  value, actor, and time (CH:79-87), and one audit row per write with `at`, `kind`, actor, channel,
  key, and old and new values (CH:130-132), plus the settings generation row. Apply mode and
  `sensitive` are registry metadata, which reads CH:638 together with BS's registry erratum
  (BS:39-40).
- `member_overrides` and `member_preferences` (6c): per-member rows keyed to users (CH:122-125,
  CH:161).
- `device_credentials` (6c): the credential hash, the user, the device id, creation, last-seen
  address, expiry, and revocation (CH:543-548).
- `quota_counters` (6c): the per-member counter and the group total (CH:488-500).
- `move_intents` (6c): `at`, `kind`, the old and new paths as `BLOB`, a granularity (file or
  directory), and a state, with the replay rule "source absent and target present means done"
  (CH:451-458). Rung 16's relocation is a named consumer.

Root allocation supports rung 16's relocation (BS:131-141, CH:662-664): every collection's root is
allocated from its id, with no column for a grandfathered root. If rung 12's Navidrome check fails
and Geoff rules for grandfathered roots through rung 10's plan, that ruling adds its column then.
Track paths relative to their collection's root make a relocation one row update and one subtree
move.

### Paths (U3, the build spec's byte-path rule)

Every path is relative and byte-preserving from the first schema (BS:662-664, EVO:356-366): a
collection's `root_path` is library-relative and a track's `rel_path` is relative to that root. A
path crosses the store boundary as `[]byte`, and invariant 1's `STRICT` rule enforces it in SQLite
as well as in Go. Paths compare bytewise. The store refuses a path that is absolute, has an empty,
`.`, or `..` component, or carries NUL. The build compiles with `SQLITE_LIKE_DOESNT_MATCH_BLOBS`, so
`LIKE` never matches a path; rung 9's query verb says how it searches paths. Rung 7's gate rules on
non-UTF-8 filenames and canonicalizes symlinks once (EVO:363-366, EVO:600-602); whatever it admits,
the store returns unchanged.

**Erratum 2026-09-24 (6a plan approval, E5):** the path refusal lands as schema CHECK constraints
at 6a, since no Go code writes a path yet. The Go-side validator lands with rung 7's first writer.

### Per-track metadata (U4, the build spec's typed-metadata rule)

Typed columns added by migration, one per field a built-in or the engine owns, with no attribute
bag (invariant 5). Adding a built-in is a migration and a typed accessor in the built-in's own rung.
EVO weighs it this way ("the weak-typing complaint corroborates U4," EVO:489-494), and the cost is
the no-plugin-ABI fork OS settled (OS:338-352).

---

## The ops seam

### Vocabulary and privileges

`internal/ops` defines one interface. Its request and response types are JSON-serializable and are
the source of the API's wire types (OS:230-236). The acting identity is a method parameter and never
a field of the wire type: a verified `identity.Identity` for a network caller, or a local operator
named by the operating-system user for a CLI verb (R5S:1141-1143). The contract suite proves that a
body carrying an actor field is ignored.

Every catalog entry declares a privilege from a closed set, `member` or `admin`, with no zero value.
The error vocabulary is a small set of sentinels, each with a stable wire code: unavailable (with a
reason), not found, conflict, invalid request, forbidden, canceled, and internal. Callers match with
`errors.Is` in Go and by code over HTTP.

### The job model

Long-running operations return a job handle, an opaque random value (OS:241-248). States are
queued, running, succeeded, failed, canceled, and interrupted; a terminal state never changes.
Progress is polled, and the interface leaves room for a local push layer (OS:243-245).

Liveness is per process. Every process that runs jobs holds a lease: an exclusive `flock` on
`leases/<instance-id>`, held for the process's life. The runner takes its lease before it inserts
any record, and every record names its owner instance. An exclusive kind (backup is one) also holds
`leases/kind-<kind>` exclusive for the job's life, so a second start returns a typed conflict naming
the running job. Recovery runs at every write-mode open and before any exclusive start:

| Record | Owner lease | Outcome |
| --- | --- | --- |
| queued or running | held | left alone, whatever its kind |
| queued or running | free | marked interrupted; recovery then removes the lease file it acquired |
| terminal | any | unchanged |
| a late terminal update to a record already marked interrupted | any | refused and logged |

Cancellation is an explicit operation that writes `cancel_requested` on the record; it never relies
on a context crossing a process. The owning runner polls the column at its progress points and
cancels the job's context. Canceling a queued job moves it straight to canceled; canceling a
terminal job is a no-op that reports the state; a running job that finishes first ends in its own
state, and the response says which.

A job runs in the process that started it. The CLI `backup` verb starts a backup through
`ops/local`, polls it, cancels it on SIGINT or SIGTERM through the explicit operation, and exits with
a distinct code per outcome: 0 for succeeded, and distinct nonzero codes for failed, canceled,
unavailable, and conflict. The 6b-i plan fixes the numbers against the existing verbs' convention and
records them in the verb's help, because rung 16's timer and rung 9's failure ping read them. At each
job start, `ops/local` prunes terminal records older than `DUBPLATE_JOB_RETENTION` (a 6b-i constant,
re-homed at 6c **[CH]**), reading it through the accessor after `serve`'s seed; pruning never removes
the newest successful record of a kind, which `doctor`'s last-backup age reads.

### Operation availability

The catalog lists every operation with its availability, the reason when unavailable, and its
privilege. It is the seam's mechanism against ad-hoc parity opt-outs and never shares a term with the
streamer seam's driver capabilities (OS:273-277). Every asymmetry it reports is named in a spec
(OS:279-285). This spec names two:

- **Backup is config-gated.** With the flag off it is unavailable, reason "backup integration
  disabled," and a start returns the unavailable code and writes no job record.
- **Restore is not an operation.** It is an offline verb (ruling 6), so it is not an asymmetry.

### The HTTP projection and the contract suite

The projection maps each operation to a handler over the same types a future `ops/client`
unmarshals. It stays unexported, and its constructor requires an authorizer, taking the mount's
channel, and fails on nil. It consults the authorizer with each operation's declared privilege
before dispatch; a denial returns forbidden and has no side effect.

The contract suite runs each case against `ops/local` directly and through the projection over
`httptest`. It holds one case per sentinel and fails when a sentinel has none: unavailable (flag
off), not found (an unknown job id), conflict (a second backup start), invalid request (a malformed
body), forbidden (a denying authorizer), canceled (canceling a running job), and internal (the store
closed under the implementation, a production-real trigger with no test-only constructor). It is the
wire contract's regression net (BS:668-672).

---

## Backup and restore

### The backup job

A backup job syncs the library tree, then writes a store snapshot, then sends the snapshot, then
prunes. The snapshot comes second so every file the catalog named when the sync began is on the
remote or in its archive. The pair is still not atomic: a change committed during the sync can leave
a row whose file is not on the remote, which restore's reconciliation report names. The snapshot
records the run id and the sync's completion time in its object metadata.

2.0 hook 10 forbids a byte copy of a live WAL database (S20:390-391; sqlite.org/howtocorrupt.html
section 1.2). The snapshot is `VACUUM INTO` on the maintenance connection into a pre-created 0600
file in `staging/`, which yields "a consistent snapshot of the original database"
(sqlite.org/lang_vacuum.html). It meets the hook's intent, which names the backup API or a
checkpoint-then-copy; the backup API (`conn.go:1208-1219`) stays the recorded alternative. The job
checks free space for one full copy first, runs `integrity_check` on the result, renames it, fsyncs
the directory by hand (`atomicfile` has no rename of an existing file; `WriteFile` does not fsync the
directory, and `WriteStream` does, at `internal/atomicfile/atomicfile.go:82-89`), and ends with
`wal_checkpoint(TRUNCATE)`.

The remote layout under `DUBPLATE_BACKUP_REMOTE` is `library/` for the tree, `archive/<stamp>/` for
files the sync replaced or deleted, and `snapshots/<stamp>.db`. Stamps are UTC and sort
lexicographically. rclone's manual sets the archive's conditions: "The remote in use must support
server-side move or copy and you must use the same remote as the destination of the sync. The backup
directory must not overlap the destination directory without it being excluded by a filter rule."
So `doctor` reports a remote whose backend lacks server-side move and copy, and 6b-i's plan quotes
rclone's per-backend feature table. After a successful run the job prunes the remote by ruling 10 and
deletes local snapshots older than this run's snapshot by ruling 7.

The guards run before rclone, against the skeleton 6a landed:

- A missing library root aborts.
- With catalog tracks, a tree whose regular-file count falls below `DUBPLATE_BACKUP_GUARD_FRACTION`
  (a 6b-i constant, re-homed at 6c **[CH]**) of the track count aborts, and rclone runs with
  `--max-delete` sized from the track count. rclone's manual: "This tells rclone not to delete more
  than N files. If that limit is exceeded then a fatal error will be generated and rclone will stop
  the operation in progress."
- With zero catalog tracks, every box's state until rung 7 or rung 10 writes one, an empty root
  aborts, and rclone runs with `--max-delete 0`. Whether rclone counts a move into `--backup-dir`
  against that limit is unverified; 6b-i criterion 13 settles it against the real binary. If it does
  not count, the job instead aborts a zero-track run whose remote `library/` is non-empty.

Until the catalog holds tracks, a non-empty wrong mount passes the guards, which is a runbook line. A
wiped mount under surviving collection directories passes a top-level check, which is why the count
guard replaces salvage's (`_salvage/internal/backup/backup.go:22-34`, `:125-128`). Every rclone
invocation runs under `exec.CommandContext`; the 6b-i plan quotes `os/exec` on what cancellation
sends and how `WaitDelay` bounds the wait, and rclone's documentation on sync and exit codes.

The gate runs a real `rclone` against a `local` backend in a temporary directory; a missing
`rclone` fails the test, and 6b-i's close adds it to CLAUDE.md's Gate section. A recording stand-in
whose exit code each case sets covers the remote-failure and SIGINT cases.

**Credentials.** rclone reads a remote from `RCLONE_CONFIG_<NAME>_*`, where "you must write the name
in uppercase in the environment variable," and "if you want to create a remote using environment
variables you must create the ..._TYPE variable." The name "can only contain letters, digits, or the
_ (underscore) character," and "you can only set the options of the immediate backend, so
RCLONE_CONFIG_MYS3CRYPT_ACCESS_KEY_ID has no effect, if myS3Crypt is a crypt remote based on an S3
remote" (rclone's manual). So `<NAME>` is the remote name in `DUBPLATE_BACKUP_REMOTE` before the
colon, uppercased. With the flag on, `doctor` refuses a name holding any other character, checks that
the named remote's `_TYPE` is set, and, for a `crypt` remote, that the remote it wraps is defined
too. Ruling 14's warning reads the same `_TYPE`.

Backup does not preserve hardlinks, so a restored library is larger than its source (CH:410-414).
Rung 16's relocation plan states how the next backup handles a moved subtree.

### The restore verb

`restore` is offline (ruling 6). Its flags:

| Flag | Meaning |
| --- | --- |
| `--snapshot <stamp>` or `--snapshot latest` | an object under `snapshots/` |
| `--from-file PATH` | a local file, such as a pre-migration snapshot |
| `--library` | also restore the tree |
| `--into-nonempty` | permit a non-empty library root |
| `--older` | permit a remote snapshot other than the latest, store only, never with `--library` (ruling 11) |
| `--discard-current` | move an unopenable current store aside |
| `--resume` | continue a recorded restore |

Its flow:

1. Take the lock exclusive, check the uid, and check free space.
2. Fetch or read the snapshot into `staging/` at 0600, run `integrity_check`, and refuse one whose
   `user_version` exceeds the binary's.
3. If a current store exists, open it on a raw connection that skips the version check, since a
   checkpoint needs no schema; a newer current store is accepted, which is the rollback path. Run
   `wal_checkpoint(TRUNCATE)` and `journal_mode=DELETE`, close it, and verify no `-wal` or `-shm`
   remains, so the swap moves one file. A current store that cannot open is refused unless the
   operator names `--discard-current`, which moves the unopened files aside together and names them.
4. Write the `restore-intent` record through `atomicfile.WriteStream`, naming the pre-restore and
   incoming files and whether the library half is owed.
5. Rename the current database to `snapshots/pre-restore-<stamp>.db`, rename the snapshot into
   `db/`, and fsync both directories. The next open restores WAL mode through its DSN.
6. With `--library`, copy `library/` from the remote into the library root, refusing a non-empty
   root without `--into-nonempty`.
7. If the restored version equals the binary's, run the reconciliation report (rows without files
   and files without rows), write it to a file in the leaf, summarize it on stdout, and remove the
   intent. Otherwise set the intent's state to `reconcile-pending`; `serve`'s first start migrates
   the store, runs the report, and removes the intent, and no other writing open proceeds until it
   has.
8. Print the ruling 12 rotation instruction.

A crash at any step before the intent is removed leaves it, and every write-mode open refuses until
`restore --resume` finishes the recorded steps. A pre-restore copy follows ruling 7's retention.
Restoring a pre-6d snapshot under a 6d binary leaves the users table empty, and the report says to
run `member import`.

---

## The settings home **[CH]**

### The accessor and the registry metadata

Every store-homed key reads through the settings accessor, never from a long-lived struct
(CH:88-92). Invalidation crosses processes: the settings generation row holds an integer that every
settings write bumps inside its transaction, and the accessor reads it, one indexed read, before
trusting its cache. `PRAGMA data_version` cannot serve, because "it is only meaningful to compare the
'PRAGMA data_version' values returned by the same database connection at two different points in
time" (sqlite.org/pragma.html), and a pool hands out different connections.

The registry gains four pieces of metadata per key: its home (wall or store), its apply mode
(immediately, next operation, or restart), the `sensitive` marker (CH:128-129, CH:638), and its
bounds as a validator that takes a value rather than an environment. It also gains a registry-level
list of correlated rules. Today's bounds live in the source's `Validate` hook and read `environ`
(`cmd/dubplate-server/auth.go:230-280`), which a store write cannot supply. A key registered without
a home stays on the wall. A store-homed key has no wall getter: the seed rule is the only reader of
its env value, so the compiler forbids a consumer bypass. `doctor` shows each store-homed key with
its value, source, apply mode, and marker.

### The seed rule

Only `serve` applies the seed rule, at start after the wall loads and the store opens and before any
request (6c builds it, 6d wires it into `serve`). It applies every seed in one `IMMEDIATE`
transaction and validates each key's bounds and every correlated rule against the post-seed set, so
raising the link expiry and the nonce retention together in one deploy succeeds. Every other verb,
`check` included, evaluates the rule read-only and reports a disagreement without failing, so a
timer-run backup never fails over an auth threshold it does not read. A seed or re-seed writes an
audit row with actor `system` and channel `env-seed`. A `default` source stores no value; reads
resolve the registry's current default, so a re-researched default reaches every existing store. A
consumer that runs before the seed, the migration runner's local snapshot cap, reads the stored row
or the registry default. The closed table:

| Row | Env value | Outcome |
| --- | --- | --- |
| none | absent | no row; reads resolve the default |
| none | present | the value seeds a row, source `env-seed`, audited |
| `env-seed` | equal to the seed | nothing changes |
| `env-seed` | different | re-seeded, logged, audited |
| `env-seed` | absent | the row is removed and reads resolve the default, logged and audited (ruling 13) |
| `admin` | equal | nothing changes |
| `admin` | different | `serve` fails to start, naming both values and the action that resolves it |
| `admin` | absent | the admin value stands |
| any | stored value outside the binary's bounds | `serve` fails to start, naming the key and the bound |
| a row for a key no longer registered | any | ignored, and `doctor` reports it |
| any | unparseable | a wall error, as today |

No channel sets source `admin` at rung 6. The write path that does stays unexported until rung 14,
and the conflict row is proved in-package by a test that writes an admin row directly.

Pass 6c re-homes 6b-i's five constants as the first store-homed keys **[CH]**, each with a
researched default recorded where it is defined (BS:236-248): `DUBPLATE_JOB_RETENTION`,
`DUBPLATE_BACKUP_GUARD_FRACTION`, `DUBPLATE_BACKUP_SNAPSHOT_KEEP`,
`DUBPLATE_BACKUP_ARCHIVE_RETENTION`, and `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`.

### The seed handoff

Fifteen thresholds are registered through the env today: rung 4's two shadow keys, pass 5a's three
serve keys, and rung 5's ten auth keys (`cmd/dubplate-server/auth.go:99-200`). All ten auth values
resolve once into `authConfig` (`auth.go:402-416`) and reach the auth core as construction arguments,
which is the long-lived struct CH forbids. So 6d builds the apply-mode column below as a required
behavior, replacing `authConfig` with accessor functions `cmd/dubplate-server` injects at each call
site 5c and 5d wire; today's code applies every auth threshold at restart.

| Key | Registered | Home after 6d | Required apply mode | Rule for state already issued |
| --- | --- | --- | --- | --- |
| `DUBPLATE_MAGIC_LINK_EXPIRY` | `auth.go:99` | store | next operation | a link keeps the expiry it was minted with (`MintLink`, `internal/auth/token.go:118`) |
| `DUBPLATE_SESSION_LIFETIME` | `auth.go:105` | store | next operation | applies at each session's next refresh (`RefreshSession`, `session.go:132`) |
| `DUBPLATE_SESSION_ABSOLUTE_CAP` | `auth.go:111` | store | immediately | verification reads the live cap (`VerifySession`, `session.go:206`) at every request, in both directions |
| `DUBPLATE_NONCE_RETENTION` | `auth.go:117` | store | next operation | the next sweep uses it; its correlated rule with the link expiry moves from the source's `Validate` hook to the registry's list |
| the four rate limits | `auth.go:166-184` | store | restart | the limiters are built once (`NewLimiter`, `internal/auth/limit.go:62`); a live apply is rung 14's option |
| `DUBPLATE_AUTH_REQUEST_BODY_CAP` | `auth.go:190` | store | next operation | the next request reads it |
| `DUBPLATE_UNKNOWN_ADDRESS_ANSWER` | `auth.go:196` | store | next operation | the next answer reads it |
| `DUBPLATE_SERVE_PREFIX`, `_BIND_ADDRS`, `_PORT` | `serve.go:35-59` | wall | restart | unchanged; read before the store opens (CH:74-75) |
| `DUBPLATE_SHADOW_FLOOR`, `_BUDGET` | `shadow.go:73-88` | wall (ruling 3) | restart | unchanged; the shadow unit is not granted the store leaf |

The auth rows apply CH's rule that tunables with defaults live in the store (CH:66); the busy
timeout and the migration wait join the serve keys on the wall. 6d declares both lifetimes and the
absolute cap `sensitive` for rung 14's re-authentication gate. The template carries no auth
threshold (`env/dubplate-server.env.template:110-117`), so the handoff removes only ruling 1's
`DUBPLATE_ROSTER_FILE` row.

---

## Identity in the store

Pass 6d opens only after rung 5's pass 5d merges, because every file it edits is rung 5's until
then. `web-auth-security-reviewer` runs inside its chain.

### The import

`dubplate-server member import --roster PATH [--dry-run]` follows ruling 1. It runs in one
`IMMEDIATE` transaction under the store lock held shared, like every `member` verb, so `serve` keeps
running. It carries each entry's `admin` flag, mints every new member's surrogate id, and writes an
import record (the file's SHA-256, the member count, the time, and the operating-system actor) in the
same transaction. An entry whose address, handle, and admin flag match an existing row is that
member. `--dry-run` reports every refusal at once and writes nothing. The same commit removes the key
from the magic-link source's `Required` list (`cmd/dubplate-server/auth.go:44-46`) and registers it
through `AddSuperseded` with a message naming the verb, so a set key is an ordinary load error that
needs no store state. Its states:

| Roster file and store | Report |
| --- | --- |
| the file absent or unreadable | refused, naming the file |
| the file malformed | refused, naming the file and the parse error |
| two entries with one address | refused, naming both entries |
| two entries with one handle | refused, naming both entries |
| an entry with no `streamer_username` | refused, naming each such entry (ruling 8) |
| no admin in the file or the table | refused, naming the rule |
| an entry colliding with a different existing member on address or handle | refused, naming both |
| an import record already present | refused, naming the record's time and actor |
| every entry valid, matching rows merged | imported; the record written |

That is nine states, eight refusals and one success; 6d criterion 2 proves each refusal by its own
fixture. 6d removes `deploy.sh`'s pre-create of the roster leaf (R5S:198-200, PA:218) and leaves the
file in place; the runbook tells the operator to delete it after verifying the import. Rolling back
past 6d needs the old template and the old file, and members added since then are re-added through
the verb afterward.

### The resolution point

R5S ruled that the store moves behind the roster's interface without touching a call site
(R5S:194-196, R5S:287-292, PA:209), and `internal/auth/roster.go:113-116` already frames the
substitution as one constructor. The resolution point (`Lookup`, `roster.go:286`) reads the users
table. `identity.Identity` gains a `UserID` field carrying the surrogate id, which per-user rows key
on. `Principal` stays the streamer handle, because rung 12 injects the principal into the proxy
header (R5S:1243-1245) and Navidrome's external-authentication documentation says "if a user is
successfully authenticated by the proxy but does not exist in the Navidrome database, it will be
created with a random password," so an opaque id there would orphan every member's history.

The write path mints the id for a member added by the page or the verb. From 6d, the add-member page
and verb require `streamer_username` and refuse a duplicate by name. The rule 5b-ii ruled, that
`AddMember` refuses the admin flag on every channel except 5c's CLI grant, carries over, and the
import is the one other admin channel, once, from the file. 5c's last-admin guard and the
`streamer_username` collision check run inside the write transaction, so a concurrent self-demotion
of the last two admins leaves one. The member-change audit record stays the log record rung 5 built
(`roster.go:316`); CH routes no member-audit table.

### The epoch

The epoch gets no bulk move. Records are named `sha256(kind, value)`
(`internal/auth/store.go:105-114`), which is one-way, so a re-key cannot reach an address removed
before 6d, and a copy races a concurrent bump from another process (`EpochStore.mu` is in-process
only, `internal/auth/epoch.go:66-72`). Instead, the revocation check reads both records, the one
keyed by the session's address and the one keyed by its resolved `UserID`, and takes the later epoch,
permanently. Every bump after 6d writes both keys, so a removal is refused through the address record
even when the address later resolves to a different id. Every epoch write takes an advisory lock on
the epoch directory, reads the stored value, and writes only a later one, so no process moves an
epoch backward. Rung 5's revocation notification (R35, R5S:714) keeps its address argument, because
sessions carry the address, and gains the `UserID`. The records stay files, as rung 5's R18 rules
(R5S:538-539). With no copy there is no half-finished state for a crash to leave.

### `serve` and the store

6d wires `serve` to the store in `cmd/dubplate-server/serve.go`: the serve-mode open and its
migration procedure, then recovery, then the seed rule, then the listener.

---

## The 2.0 forward hooks

Per CLAUDE.md's build-for-2.0 rule, each pass states per task whether it lands a hook or is neutral;
a hook is built by the rung that first consumes it.

- **Lands at 6a:** hook 4 as invariant 3, binding the job-records table first (`settings_audit` at
  6c is its first event table); hook 5's users table; hook 7, migrations; hook 8, `writer` with its
  domain, epoch, and generation **[CH]**; hook 9, one writer, WAL, `FULL`, and the write method.
- **Lands at 6b-i:** hook 10, by `VACUUM INTO`.
- **Lands at 6c:** hook 15's routed schema, settings home, and marker; its two mounts and
  re-authentication gate are rung 14's **[CH]**.
- **Lands at 6d:** hook 5's surrogate id, `UserID`, and the dual-key epoch.
- **Deferred by ruling 2, each owing an S20 erratum:** hook 1, the audio key and file hash, both
  alternate keys (rung 7); hooks 2, 3, and 11, the sort key, art hash, and gains (rung 9).
- **Neutral:** hook 6 (rung 12), hook 12 (rung 7), hooks 13 and 14 (rungs 8 and 9).

---

## Prior art

Scored on CLAUDE.md's four criteria. Versions and dates are from `go list -m` in a scratch module on
2026-09-23; licenses from each module's `LICENSE`; importer counts from pkg.go.dev.

| Entry | Role | License | Quality | Maintenance | Adoption | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `modernc.org/sqlite` | The driver | BSD-3-Clause | Transpiled C under a hand-written `database/sql` layer. Its DSN parser validates every shorthand key before applying any (`sqlite.go:290-305`); `_pragma` values are executed verbatim and are the one exception (`:305-306`). SQLite 3.53.4 with FTS5, snapshots, and `LIKE_DOESNT_MATCH_BLOBS`. | v1.59.0, 2026-09-15; ten releases since 2026-05-10. 25 open issues, oldest 2020-09-05. Bus factor 1. | 3,518 importers | **Adopt**, meeting S20's condition: a gate test proves the FTS5 trigram tokenizer and WAL with a reader held across a commit. |
| `github.com/ncruces/go-sqlite3` | The fallback | MIT | A Wasm build of SQLite translated to Go by wasm2go; no wazero in its requirements | v0.35.6, 2026-09-23 | 96 importers; gonic uses it | **Record as the fallback** if the trigram or WAL test fails. Still v0. |
| `github.com/pressly/goose/v3` | Migrations | MIT | Readable, `embed.FS` provider | v3.28.0, 2026-09-02; 133 open issues, oldest 2019-01-10 | 1,725 importers | **Decline.** Graph pruning keeps its cost small (four modules, about 2.3MB of binary), but an older binary's `Provider.Up` on a newer database returns no error, the silent-ahead case the runner must refuse, and it keeps its own version table. |
| `github.com/golang-migrate/migrate/v4` | Migrations | MIT | Driver-per-package; its `database/sqlite` driver imports only modernc | v4.20.1, 2026-09-09 | 5,219 importers | **Decline.** Its modernc driver is tested against modernc v1.18.1, 41 minor versions behind the one this rung adopts, and it keeps its own version table. |
| `zombiezen.com/go/sqlite` | A non-`database/sql` API | ISC (not re-read) | Ergonomic | v1.4.2, 2025-05-23 | not measured | **Decline.** It leaves `database/sql` for a second vocabulary. |
| `github.com/mattn/go-sqlite3` | The common driver | MIT | Mature | v1.14.52 | very high | **Decline** on OS's no-cgo fork (OS:412-413). |
| rclone, as a subprocess | Backup transport | MIT | The estate's proven tool | v1.74.3 on the workstation | the estate runs it | **Adopt** as a probed subprocess. `librclone` in-process is declined: rclone's whole dependency tree for one verb. |
| SQLite `VACUUM INTO` | Snapshots | public domain | Guarantee quoted above | SQLite's own | universal | **Adopt**; the backup API is the recorded alternative. |
| `internal/atomicfile` | The restore-intent record | in-repo | house bar | in-repo | rungs 4 and 5 | **Adopt** for records; renames of SQLite-written files fsync by hand. |
| `_salvage/internal/backup` and `restore` | Guards and flow | salvage tier | salvage | frozen | ran on the box | **Adopt the missing-root guard**; the count guard replaces the top-level check. Import nothing. |
| Navidrome's schema | Users and per-user tables | GPL, behavior only | | | | **Consult** at 6a's and 6c's plans for behavior only; no code. |

The other entries CLAUDE.md names (OpenSubsonic, gonic, taglib, mewkiz/flac, Picard, chromaprint)
have no rung 6 surface.

---

## Planning-miss prevention

CLAUDE.md's four planning-miss rules are closed in place: behaviors are quoted or measured where
used (the fold record keeps the ledger and the risk model), each check's states sit in its own
section's table or flow, and each proof names its fixture in its criterion. Privileged fields and
every channel that may set them:

| Field | Channels |
| --- | --- |
| settings source `admin` | none at rung 6; rung 14's admin API alone |
| settings source `env-seed`, and `default` by removal | `serve`'s seed rule alone |
| a job's actor | the implementation, from its caller's method parameter |
| a job's terminal state | its owning runner; recovery, only when the owner's lease is free |
| `cancel_requested` | the cancel operation |
| a user's surrogate id | the store's mint, at import and in the write path; never a request or the roster file |
| a user's admin flag | the CLI grant, as 5b-ii ruled; the import verb, from the roster file's `admin` field, once |
| a collection's `root_path` and `writer` | the server alone; no request type carries either |
| backup | the CLI verb, and over HTTP only behind rung 14's authorizer |
| restore and import | their CLI verbs alone; restore offline under the exclusive lock |

---

## The pass split

Each pass's exit gate is `check.sh` green plus the numbered criteria. Each criterion names the
fixture state it needs and the reason it would fail.

### 6a: the store foundation (lane 2 worktree)

1. The trigram tokenizer answers a substring query over a populated FTS5 table.
2. A reader holding a transaction open still sees its snapshot after a writer on a second store
   handle commits.
3. Two independent store handles on one file: the second `IMMEDIATE` writer waits while the first
   holds its transaction for half the busy timeout and then commits; held past the timeout, the
   second returns the typed busy error.
4. Every DSN setting (WAL, foreign keys, `FULL`, busy timeout, `journal_size_limit`, `query_only` on
   the read pool) reads back on a freshly opened connection of each class, the maintenance connection
   included.
5. A write through the read pool fails with a read-only error; an orphan insert through the write
   method fails with `SQLITE_CONSTRAINT_FOREIGNKEY` by its extended code.
6. A callback that swallows a failed statement's error does not commit; a nested write call fails
   fast; a write with no deadline is refused.
7. The write method fails closed when `user_version` is changed under it.
8. An injected migration failing after its version write leaves `user_version`, schema, and ledger
   unchanged; a ledger hash mismatch is refused naming both.
9. Migrating the populated job-records fixture forward leaves `pre-migrate-vN.db` with `user_version`
   N and the fixture's rows; a snapshot write that fails leaves the store at N with an error naming
   the path; a second migration never overwrites the lower-version snapshot; with backup off, only
   the newest `LOCAL_SNAPSHOT_KEEP` copies survive a migration.
10. An injected foreign-keys-off rebuild migration is followed by an orphan insert through the write
    method that fails.
11. Each invariant fails against its violating fixture for its named catalog fact, including a Go
    `string` bound into a `STRICT` path column, and passes over the skeleton.
12. A duplicate address, live or tombstoned, is refused by the `users` index (ruling 9).
13. A zero-byte file and a non-database file are each refused as not a store. A write-mode open of an
    absent store refuses naming "start serve first" and creates nothing; of an older store, refuses
    naming both versions; a newer store is refused.

    **Erratum 2026-09-24 (6a plan approval, E2):** a zero-byte file and a header-only unmarked
    file are interrupted creates that serve mode completes. "Refused as not a store" holds for
    the inspect and write modes only.
14. Serve mode with a second handle holding the lock shared waits and does not migrate; after that
    handle releases, it migrates. A second process that requests the exclusive lock while serve mode
    converts is granted it, and serve mode then holds no handle to the file that process renamed, and
    re-reads the version.
15. The inspect mode against an absent, an older, and a newer store leaves the database file's bytes
    and mode unchanged, with any `-wal` or `-shm` owned by the leaf owner, and reports each state by
    name; `doctor`'s store row renders each in a golden.
16. The `statfs` seam reports NFS (refused), FUSE (warned), and ext4 (opened); a uid mismatch is
    refused naming the owner; the database and a snapshot are created 0600.

### 6b-i: the ops seam and backup (lane 2 worktree)

1. Every catalog entry's privilege is `member` or `admin`; the zero value fails construction.
2. A body carrying an actor field is ignored, and the job records the caller's identity.
3. A record owned by a live second process (a second `ops/local` over the same store holding its
   lease) survives another process's open.
4. A queued and a running record whose owner lease is free read interrupted after an open, and the
   dead owner's lease file is gone.
5. A late terminal update to an interrupted record is refused and logged.
6. Canceling from a second `ops/local` sets `cancel_requested`, and the owning runner ends the job
   canceled; canceling a queued job ends it canceled; canceling a finished job reports its state.
7. A second backup start returns conflict naming the running job.
8. With the flag off, the catalog reports unavailable with the exact reason and a start writes no
   record; with the flag on, the catalog reports available.
9. A snapshot taken while another handle commits in a loop passes `integrity_check`, holds every row
   committed before it began, and holds no row from an uncommitted transaction.
10. A successful run against a real `rclone` `local` backend leaves every library file under
    `library/` and the snapshot under `snapshots/`.
11. A file deleted from the library appears under `archive/<stamp>/`, not deleted.
12. A missing root, an empty root with no catalog tracks, and a root whose file count falls below the
    fraction of a populated skeleton's tracks each abort with no rclone invocation recorded by the
    stand-in.
13. With zero catalog tracks and a non-empty remote `library/`, a run that would archive a file ends
    failed, by `--max-delete 0` or by the fallback abort, whichever the real binary requires; the
    plan records which.
14. A stand-in exit 7 makes the job failed with the code in its record.
15. After a successful run, the remote keeps exactly the newest `SNAPSHOT_KEEP` snapshots, including
    the latest, and no archive directory older than `ARCHIVE_RETENTION`.
16. After a successful run, every local pre-migration and pre-restore snapshot older than its
    snapshot is deleted.
17. A terminal record older than `DUBPLATE_JOB_RETENTION` is pruned at the next job start, and the
    newest successful record of each kind survives.
18. `backup` with an absent store refuses and creates nothing.
19. Setting each retired name fails the load with its message verbatim: for the five R2 names,
    "superseded by DUBPLATE_BACKUP_REMOTE; configure the remote in rclone's RCLONE_CONFIG_<NAME>_*
    environment form"; for `DUBPLATE_BACKUP_STAGING`, "superseded: backup staging now lives in the
    store leaf under DUBPLATE_STATE_DIR".
20. `backup` receiving SIGINT mid-sync ends the job canceled, leaves no stand-in process after
    `WaitDelay`, and exits with the canceled code; each exit code maps to its outcome as the verb's
    help states.
21. `doctor`'s backup rows render, each in a golden: never backed up, last success age, unavailable,
    a remote name with an invalid character, a missing `_TYPE`, a `crypt` remote whose wrapped remote
    is undefined, a backend without server-side move or copy, the crypt warning (ruling 14), and the
    store's WAL size, free space, and local snapshot count and size.

### 6b-ii: the projection and restore (lane 2 worktree)

1. The contract suite holds one case per sentinel and fails when a sentinel has none; every case
   passes in-process and over HTTP.
2. A denying authorizer returns forbidden over HTTP and writes no job record; a nil authorizer fails
   construction.
3. A job started over HTTP is polled and canceled over HTTP and ends canceled with its work stopped.
4. A cancel for a job running in another `ops/local` over the same store succeeds through the
   projection.
5. `restore` refuses while a real reader holds the store.
6. With a live store whose WAL holds uncheckpointed rows, restore checkpoints, swaps one file, and
   the pre-restore copy holds those rows.
7. A crash injected through the step hook after each flow step leaves the intent record; a write-mode
   open refuses naming `--resume`; `--resume` completes.
8. A crash between steps 4 and 5 followed by a write-mode open leaves the staged incoming snapshot in
   place, and `--resume` completes the swap.
9. A newer, corrupt, or absent snapshot is refused; an older remote one is refused without `--older`;
   an unopenable current store is refused without `--discard-current`.
10. An older binary's `restore --from-file` of its pre-migration snapshot over a newer current store
    succeeds, and the swapped-in store opens under that binary.
11. A restored older-version snapshot leaves a `reconcile-pending` intent; serve mode migrates it,
    writes the report, and removes the intent, and a write-mode open before then refuses.
12. A tree restore into an empty root yields the backed-up files; into a non-empty root without
    `--into-nonempty` it refuses naming the flag; the reconciliation report names a planted row
    without a file and a planted file without a row.
13. The swapped-in database file is owned by the leaf's owner at 0600.

### 6c: the routed schema and the settings home **[CH]** (lane 2 worktree if 5c and 5d leave `internal/config` untouched; otherwise after 5d)

1. The invariant test passes over the full schema, with a populated fixture for every migration; a
   Go `string` bound into `move_intents`' path columns is refused; an orphan override row is refused.
2. `visibility` refuses any value but CH's two.
3. `settings_audit` and `move_intents` carry `NOT NULL` `at` and `kind`.
4. A seed outside its bounds, validated through the registry's value validator with no environment,
   fails the seed rule.
5. The correlated retention rule passes when both keys rise in one seed set and fails when only the
   expiry rises past the stored retention.
6. Each row of the seed table behaves as stated, each by its own test.
7. A registered default changed between two opens reaches a store with no row.
8. `backup` started with a divergent env leaves every settings row unchanged and reports the
   disagreement.
9. A seed writes one audit row with actor `system` and channel `env-seed`; a refused write writes
   none.
10. A write through a second store handle invalidates the first handle's cached value.
11. A key registered without a home still reads from the env; a store-homed key has no wall getter (a
    compile-time fact the plan states and the diff-reviewer checks).
12. `doctor` renders value, source, apply mode, and marker, and an orphan row, in goldens.
13. Each of 6b-i's five constants reads through the accessor, and an admin row for one changes the
    behavior its 6b-i criterion proves.
14. `check` evaluates the seed rule read-only and leaves the database file's bytes unchanged.

### 6d: identity, the handoff, and `serve`'s store (after rung 5's 5d merges)

1. `member import` imports a valid roster once, carries each entry's admin flag, writes the import
   record, and a second run refuses; an imported admin can grant admin through the CLI.
2. Each of the eight import refusal states reports as its table row states, each by its own fixture,
   and `--dry-run` reports all of a file's refusals at once and writes nothing.
3. A member added through the verb before the import merges when the file's entry matches it and
   refuses naming both when it differs.
4. With `DUBPLATE_ROSTER_FILE` set, every verb fails the load naming `member import`.
5. A verified identity's `UserID` is the surrogate id and its `Principal` the streamer handle.
6. A member added through the write path receives a new id; the page and the verb refuse a missing or
   duplicate `streamer_username` by name; the admin flag arrives only through the CLI grant or the
   import; re-adding a removed address reactivates the same id (ruling 9); a concurrent
   self-demotion of the last two admins leaves one admin.
7. A session revoked by an address-keyed bump before 6d, and one revoked by a bump after 6d, both
   stay revoked; a pre-removal session for a member removed after 6d and re-added under a new id
   (ruling 9's "no" branch, forced in-package) stays revoked; a session revoked under neither key
   stays valid.
8. Two processes bumping one epoch concurrently never leave a stored value lower than either bump.
9. With the env holding A and an in-package admin row holding B, a link minted after the write
   carries B's expiry, a session's next refresh uses B's lifetime, the next request enforces B's
   absolute cap, and a limiter built at the next start uses B.
10. Each re-homed key seeds from the env, and `doctor` shows its source and the `sensitive` marker on
    both lifetimes and the cap.
11. `serve` against an absent store creates it, and against an older store migrates, before its
    listener accepts a connection; its lock wait past `DUBPLATE_STORE_MIGRATION_WAIT` exits nonzero
    naming the holder.
12. `serve` applies the seed rule before the first request, and a request handled after start reads
    the seeded value.
13. `systemd-analyze verify` is clean on 5d's serve unit, whose `ReadWritePaths` names the `store`
    leaf; the state-dir leg covers the leaf.
14. The template carries no `DUBPLATE_ROSTER_FILE` row; `deploy.sh` pre-creates the `store` leaf and
    no longer the roster leaf, runs store-opening verbs through `runuser`, and names the holder before
    restarting `serve` onto an older store whose lock is held.

### Band and ceilings

The ladder drew rung 6 as band M, one pass (BS:656; bands at BS:563-567). The ratified CH routes a
schema and a settings home here **[CH]**, and rung 5 hands rung 6 the surrogate id, the epoch
obligation (R5S:1809-1811), and the roster's move behind its interface (R5S:194-196, PA:209). Five
passes make it band L, which the build spec's ladder entry owes an erratum.

Each ceiling is plan preparation, plus a 0.2M staleness pre-task, plus tasks at a 0.46M chain rate,
plus a 0.6M close. The chain rate comes from 5b-ii: STATUS records about 5.6M spent after Task 10,
and HISTORY's pass spends include plan preparation and Task 0, so (5.6 − 0.85 − 0.2) / 10 ≈ 0.46M.
Plan preparation ranged 0.83M to 1.07M across 5a, 8a, 8b, 1b, and 4b (HISTORY:1735, 2220, 2627,
2961, 3136); the table uses 0.95M, and 1.07M for 6c, since 1b is the config-contract analog. The fold
record shows each row's arithmetic.

| Pass | Tasks | Ceiling | Flag (80%) |
| --- | --- | --- | --- |
| 6a | 8 | 5.5M | 4.4M |
| 6b-i | 9 | 5.9M | 4.7M |
| 6b-ii | 5 | 4.4M (includes 0.3M targeted review) | 3.5M |
| 6c | 7 | 5.6M (includes 10% for a config-contract change) | 4.5M |
| 6d | 9 | 6.6M (includes 0.7M security review in the chain) | 5.3M |

**Erratum 2026-09-24 (6a plan approval, E4):** the 6a row reads 6.8M execution plus planning
recorded separately, with the flag at 5.44M, the approved rung 6a plan's Ruling 1 yes. The 5.5M
and 4.4M figures above are superseded for 6a; the other rows are unchanged.

The total is 28.0M. The phase ceiling and each lane's sub-ceiling are set at lane launch (PA
section 4). Each ceiling is re-derived from the previous pass's actual before the next plan is
drafted, and a plan that exceeds ten tasks splits the pass, names the follow-on, and re-derives both
ceilings.

### Contested surfaces

Before lane 2 launches, the conductor writes the contested-surface rows in the fold record's
"Revision 2.2" section into PA's ownership table (PA:173). The rows name every file a rung 6 pass
shares with another lane: `internal/config`'s registry and `keys.go`, `main.go`'s registration
block, `doctor.go` and `check.go`, `serve.go` (6d only), `go.mod`, the state directory layout, rung
5's auth and identity files (6d only), and 5d's serve unit, `deploy.sh`, the state-dir leg, and the
template (6d only).

---

## Process and review

Each pass opens with a staleness pre-task, then runs the implementer, `diff-reviewer`, and gate
chain per task in segments of three to four tasks. Review scales to risk: the data-loss paths earn
one targeted fresh-context `opus` read at 6b-ii's close (the migration runner, the lock procedure,
the snapshot, and restore, briefed with the fold record's risk model), and 6d runs
`web-auth-security-reviewer` inside its chain. At each merge the conductor dispatches one
`go-architecture-reader` per touched package, and every `code-simplifier` dispatch carries
`.claude/instructions/simplifier-brief.md` verbatim.

Passes 6a and 6b run as lane 2 beside 5c and 5d, and none of them makes `serve` open the store. 6c
joins lane 2 only if 5c and 5d leave `internal/config` untouched; otherwise it waits for 5d. Pass 6d
waits for 5d. Rung 6 edits 5d's serve unit, `deploy.sh`, and the template in the repo at 6d; the
deploy rides the next ruled deploy, and the no-deploy carry in `docs/STATUS.md` holds until 5d lifts
it.

---

## Errata this spec owes

Each lands, dated 2026-09-23 or at ratification, in the named document when Geoff rules this spec.
"If yes" marks an erratum that depends on a ruling.

| # | Document | Change |
| --- | --- | --- |
| 1 | BS:656 | Ladder entry 6: band M to L, five passes |
| 2 | BS:934-935 | Rungs 5 and 6 edit deploy files in the repo; their deploy rides the next ruled one |
| 3 | BS:218-220 | The backup class is one remote plus a flag, with rclone env credentials (if ruling 5 is yes) |
| 4 | CH:74-75 | Keys read by a unit not granted the store leaf (the shadow pair) stay on the wall (if ruling 3 is yes) |
| 5 | CH:52-54 | "The roster file" retires from the wall list; the file is an import source only (if ruling 1 is yes) |
| 6 | CH:56 | The backup schedule key is declared by rung 16 with its timer |
| 7 | CH:638 | Apply mode and the sensitive flag are registry metadata (clarifying, with BS:39-40) |
| 8 | R5S:1552-1555, R5S:198-200, PA:218 | The roster key and the roster leaf's pre-create retire at 6d (if ruling 1 is yes) |
| 9 | R5S:1153-1158, `identity.go` | The surrogate id is `Identity.UserID`; `Principal` stays the streamer handle |
| 10 | R5S:336-338 (R4), :538-539 (R18), :714 (R35), :1687-1690, :1809-1811 | The re-key is a permanent dual-key read with bumps to both keys; R35's notification gains the `UserID` |
| 11 | S20:363-372, S20:392 | Hooks 1, 2, 3, and 11 land with the computing rung's migration (if ruling 2 is yes) |
| 12 | S20:435 and the fallback row | ncruces is a wasm2go translation; goose's decline rests on its silent-ahead `Up` |
| 13 | BS:662-664 | A track path is relative to its collection's library-relative root (clarifying) |
| 14 | BS:844-851 and rung 12's entry | Rung 12 adds the streamer database to the backup set; rung 16 declares the schedule key and renders the `RCLONE_CONFIG_*` rows |

**Applied 2026-09-23:** rows 1-8 and 10-14 landed in the named documents, inline and dated. Row
9 landed its R5S half only. The `identity.go` half waits for pass 5c-i to merge, because 5c-i's
Task 6 writes `internal/identity`; the rung 6 pass that introduces `Identity.UserID` lands it. Notes
per row:

- Row 1: `specs/2026-09-02-greenfield-server-build.md`, a new dated erratum block after the
  2026-09-12 erratum, covering rows 1, 2, 3, 13, and 14 together (the ladder list itself is
  left as history, per this file's own convention of recording supersession in the front-matter
  erratum block rather than rewriting the numbered ladder).
- Row 2: same block, box-safety correction.
- Row 3: same block, backup-class correction.
- Row 4: `specs/2026-09-11-configuration-homes-design.md`, inline blockquote after the wall
  paragraph naming the backup schedule and failure-ping classes.
- Row 5: same file, inline blockquote replacing the roster file in the owned-wall list.
- Row 6: same file, inline blockquote after the wall paragraph naming the backup schedule key.
- Row 7: same file, the rung 6 forward-hook bullet under "What this spec routes elsewhere",
  moving apply mode and the sensitive flag out of the settings-table column list.
- Row 8: `specs/2026-09-03-parallel-build-approach.md`'s state-directory-layout matrix row, and
  `specs/2026-09-08-rung-5-identity-auth.md` at the `DUBPLATE_ROSTER_FILE` claim and the
  roster-leaf pre-create claim.
- Row 9: `specs/2026-09-08-rung-5-identity-auth.md`, inline blockquote at the single-resolution-
  point passage, spec half only. **Deferred: `internal/identity`'s `Identity.UserID` /
  `Principal` naming, owned by rung 6's Task 6.**
- Row 10: `specs/2026-09-08-rung-5-identity-auth.md`, inline blockquotes at R4, R18, R35, the
  digest passage, and "Edges to later rungs": the re-key is a permanent dual-key read with
  bumps to both keys, and R35's notification gains the `UserID`.
- Row 11: `specs/2026-09-07-streamer-2-0-design.md`, a blockquote after the forward-hooks list
  naming hooks 1, 2, 3, and 11.
- Row 12: same file, the `ncruces/go-sqlite3` and `pressly/goose` prior-art table rows.
- Row 13: `specs/2026-09-02-greenfield-server-build.md`, same erratum block as rows 1-3 (the
  track-path clarification).
- Row 14: same file, same erratum block (the backup-set and schedule-key correction to ladder
  entries 12 and 16).

**Applied 2026-09-24 (6a plan approval):** E1-E11 landed inline, dated, in the documents they
name. Notes per row:

- E1: `research/2026-09-23-rung-6-spec-fold-record.md`, the `main.go` contested-surface row.
- E2: this spec, criterion 13.
- E3: this spec, "Package territory."
- E4: this spec, the ceiling table's 6a row, Ruling 1 yes.
- E5: this spec, "Paths."
- E6: this spec, the filesystem probe paragraph under "Layout, ownership, and the lock."
- E7: this spec, "Layout, ownership, and the lock," the staging-cleanup sentence.
- E8: this spec, "Migrations," the serve procedure's step 1.
- E9: this spec, "Layout, ownership, and the lock," the ownership paragraph.
- E10: this spec, "The store exposes one way to write," the deadline bullet, Ruling 4 yes.
- E11: this spec, "Three open modes," the serve mode bullet, Ruling 3 yes.

---

## Not verified in this revision

- `zombiezen.com/go/sqlite`'s license and importer count; its verdict rests on neither.
- The exact wording of `database/sql`, `os/exec`, rclone's sync and exit codes, rclone's per-backend
  feature table, and `proc(5)`, which each plan quotes.
- Whether rclone counts a file moved into `--backup-dir` against `--max-delete`; 6b-i criterion 13
  settles it against the real binary.
- Which rclone backends support server-side move or copy; `doctor` reports it per remote, and the
  6b-i plan quotes the feature table.
- 5d's serve unit's file name and `Type=`, which 5d has not landed; 6d's staleness pre-task reads
  both.
- The line citations into rung 5's files, which 5c and 5d may move.
