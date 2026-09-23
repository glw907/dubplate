# Rung 6: the ops seam and the store

Status: **revision 2.1, draft for the owner's ratification**, 2026-09-23. Revision 1 (`12a6941`)
drew four fresh-context adversarial reviews (`e97d761`, 113 findings), folded into revision 2
(`80ab7ff`); a verification read of revision 2 (`7c1b46b`, 28 findings) is folded here.
`research/2026-09-23-rung-6-spec-fold-record.md` gives every finding's disposition and the long
arguments this spec only states. No owner ruling below is given yet. Tree citations were re-read at
`7c1b46b`; the staleness pre-task each pass opens with re-checks the ones that pass leans on.

Rung 6 builds dubplate-server's library store and its operation vocabulary. The store is one
SQLite file on the pure-Go modernc driver, opened in WAL mode with `synchronous=FULL` and one
writer per process, migrated forward only by `serve` under an exclusive lock behind a
pre-migration snapshot, and guarded by a schema-invariant test whose rules SQLite's own catalog
can decide: path columns as enforced bytes in `STRICT` tables, entity ids that are never rowids,
timestamped event tables, and foreign keys that resolve and never cascade. The ops seam is one Go
interface with JSON-serializable request and response types, a job model whose liveness is a
per-process lease and whose cancellation crosses processes through the store, and an operation
catalog that carries availability and required privilege, so the organizer spec's parity rule has
a mechanism from the first version. Backup and restore are the first operations: the library tree
and then a store snapshot, sent to any rclone remote behind an integration flag, with dubplate's
own pruning, and an offline restore that swaps one file under a restore-intent record and
reconciles the catalog against the tree before `serve` may open. Rung 6 also lands the schema the
ratified configuration homes spec routes to it, the settings home with an env-seed rule that only
`serve` applies, and, once rung 5 closes, the surrogate user id, an explicit roster import, and
`serve`'s own use of the store. The ladder drew it as band M; it is five passes (band L), and
"Band and ceilings" gives the arithmetic.

Written against the ratified build spec (`specs/2026-09-02-greenfield-server-build.md`, BS), the
organizer spec (`specs/2026-09-01-dubplate-organizer-design.md`, OS), the configuration homes spec
(`specs/2026-09-11-configuration-homes-design.md`, revision 3, ratified 2026-09-23 at `fd122d9`,
CH), the 2.0 streamer spec (`specs/2026-09-07-streamer-2-0-design.md`, revision 2.2, S20), the
parallel-build approach (`specs/2026-09-03-parallel-build-approach.md`, PA), the rung 5 spec
(`specs/2026-09-08-rung-5-identity-auth.md`, R5S), the beets evolution lessons
(`research/2026-09-02-beets-evolution-lessons.md`, EVO), the dependency analysis
(`research/2026-09-03-rung-dependency-analysis.md`, DEP), and the reverse-edge completion
(`research/2026-09-04-reverse-edge-completion.md`, RE).

Every passage that exists because CH routes it here carries the mark **[CH]**, indexed in "What
depends on the configuration homes spec." Every change this spec makes to a ratified document is
an erratum line listed in "Errata this spec owes"; none is silent. Every mechanism the spec relies
on is quoted from its documentation or man page, or listed in "Not verified in this revision."

---

## Owner rulings

Fourteen yes-or-no questions change what an adopter sees, configures, or is asked for, ordered by
how much build each answer changes. Rulings 1 through 7 are the draft's, restated where the
reviews showed them unbuildable; 8 through 14 are forks the reviews raised. Everything else is a
technical decision closed in its own section.

**1. Does the roster move into the store through an explicit one-shot import verb, after which
`DUBPLATE_ROSTER_FILE` is a superseded name?** Recommendation: yes. **[CH]** Pass 6d adds
`dubplate-server member import --roster PATH [--dry-run]`. It runs in one `IMMEDIATE` transaction
under the store lock held shared, like every `member` verb, so `serve` keeps running. It refuses
when an import record exists. It merges into a users table that already holds members: an entry
whose address, handle, and admin flag match an existing row is that member; any other address or
handle collision refuses the whole import, naming both. It carries each member's `admin` flag from
the file and refuses when neither the file nor the table holds an admin. It mints every new
member's surrogate id and writes an import record (the file's SHA-256, the member count, the time,
and the operating-system actor) in the same transaction. `--dry-run` reports every refusal at once
and writes nothing. The same commit removes the key from the magic-link source's `Required` list
(`cmd/dubplate-server/auth.go:44-46`) and registers it through `AddSuperseded`, with a message
naming the verb, so a set key is an ordinary load error that needs no store state. Reason: the
draft's import-on-open needed the key set for one boot and a load error after it, which
`config.Load` cannot express because it never sees the store; every later start would fail the
load. CH makes members store state (CH:67-70), and one transaction must own a membership and its
id (R5S:1153-1158). What it changes for an adopter: the upgrade is the 6d deploy, whose template no
longer carries the key, then `member import --dry-run`, then `member import`. From the deploy's
restart until the import commits, no roster member can log in. A member added through the page or
the verb in that window merges rather than blocking the import.

**2. Do CH's routed tables land at 6c, the catalog skeleton at 6a, and S20's hook columns 1, 2, 3,
and 11 with the rungs that compute them?** Recommendation: yes. **[CH]** The rule applied is "a
table or column lands with its first reader or writer." The skeleton (users, collections, albums,
tracks) lands at 6a, because its first readers are 6b-i's store-aware backup guard and 6b-ii's
restore reconciliation, and CH's routed tables hold foreign keys into users and collections. CH's
routed tables land at 6c as the owner's explicit exception, because CH routes them to rung 6 by
name (CH:637-645) and puts the visibility column on the row at rung 6 (CH:195). The audio key and
file hash (hook 1, rung 7), the folded sort key (hook 2, rung 9), the art hash (hook 3, rung 9),
and the gains (hook 11, rung 9) arrive by those rungs' migrations under the invariants below.
Reason: S20 assigns those four columns to "rung 6 schema" (S20:363-372, S20:392), and CLAUDE.md's
2026-09-12 refinement, the later ruling, says a hook is built by its first consumer; a yes owes S20
a dated erratum. A no lands the four as nullable columns at 6c. What it changes: nothing an adopter
sees.

**3. Do the shadow floor and budget stay on the wall?** Recommendation: yes. The rest of the
threshold homing is CH's ratified rule applied, not a fork: the ten auth thresholds are tunables
with defaults and move to the store (CH:66); the serve prefix, bind addresses, and port, the
store's busy timeout, and `serve`'s migration-lock wait are read before the store opens and stay
on the wall (CH:74-75). The shadow pair is the one real fork. Its consumer is the shadow unit,
which is not granted the store leaf and retires at launch, and "a unit not granted the store leaf"
is a wall category CH does not name, so a yes owes CH an erratum line. Reason: a store grant for a
unit that dies at launch buys nothing. What it changes: the shadow pair stays in the env template.

**4. Does a backup contain the library tree and a store snapshot, with the streamer's own database
added by rung 12?** Recommendation: yes. One run syncs the tree and then snapshots the store. The
snapshot alone restores nothing: the store records where every file lives, and the files are the
custody. The salvage backup covered three things, `library.db`, `navidrome.db`, and the tree
(`_salvage/internal/backup/backup.go:91-104`); at 1.0 the streamer's database is the only home of
members' stars and playlists (CH:461-462), so rung 12, which lands the Navidrome driver, adds that
database to the backup set. Until rung 12 the loss is stated in the runbook. Reason: OS names
backup and restore as one capability (OS:77-81). What it changes: remote storage grows with the
library.

**5. Is a backup target one flag plus one remote key, with credentials in rclone's own
`RCLONE_CONFIG_<NAME>_*` environment form rendered by `deploy.sh` from the age store?**
Recommendation: yes. The class declares `DUBPLATE_BACKUP_ENABLED` and `DUBPLATE_BACKUP_REMOTE`
(rclone's `remote:path` form). `<NAME>` is the remote name before the colon, uppercased, because
rclone's manual says "you must write the name in uppercase in the environment variable" and "if
you want to create a remote using environment variables you must create the ..._TYPE variable."
So `doctor` checks, with the flag on, that `RCLONE_CONFIG_<NAME>_TYPE` is set. The credentials are
env names rclone itself reads, rendered from the age store like every other secret, so CH's "the
wall holds every secret" (CH:52-57) stays true with no erratum and no hand-edited `rclone.conf`.
The five R2 names and `DUBPLATE_BACKUP_STAGING` (`internal/config/keys.go:60-71`) become
supersessions. Reason: the R2 names fail BS's fourth universality test (BS:192-194), and a
credential in an rclone config file has no route through the age store. The alternatives are a
wall key naming an rclone config file that `deploy.sh` renders, or the draft's position with a CH
erratum narrowing "every secret." What it changes: an adopter configures any rclone backend the way
rclone documents its env form.

**6. Is restore an offline verb?** Recommendation: yes. `dubplate-server restore` refuses while
any process holds the store, and it is not an operation of the ops interface, so no network channel
reaches it. Reason: replacing a database file under an open connection "results in behavior that is
undefined and probably undesirable" (sqlite.org/howtocorrupt.html section 2.5). What it changes: a
restore means stopping the server first.

**7. Does the store keep one pre-migration snapshot per from-version, deleted only after a
successful backup postdates it, and capped at the newest few when backup is off?**
Recommendation: yes. Each is named by the version it was taken at (`pre-migrate-v4.db`), a higher
version never overwrites a lower one, and a failed snapshot write refuses the migration. After each
successful backup, the job deletes every local pre-migration and pre-restore snapshot older than
that backup's snapshot. With the backup integration off, the store keeps the newest
`DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` (a 6b-i constant, re-homed at 6c) and deletes older ones at
the next migration. Reason: the draft's single snapshot was overwritten by the next upgrade exactly
when a clean-committing migration was semantically wrong. What it changes: the state directory
holds one store copy per upgrade not yet backed up, and `doctor` reports their count and size.

**8. Does the roster import refuse a member with no `streamer_username`, naming each one, rather
than deriving a handle?** Recommendation: yes. S20 hook 5 makes the handle required, unique, and
stable (S20:376-378), and the roster makes it optional today (`internal/auth/roster.go:109`).
`member import --dry-run` names every member that would fail, and a refused import leaves `serve`
running. Reason: the handle names the member's existing streamer account; a derived one creates a
fresh account that orphans the member's history. What it changes: an adopter fills the field before
the import, or the login outage ruling 1 describes lasts until they do.

**9. Does re-adding a removed member reactivate the same surrogate id?** Recommendation: yes. The
address index is unique across live and tombstoned rows, and re-adding the address clears the
tombstone on the same row. Reason: stars and history key on the id, the address is the login proof,
and the old epoch record keeps a pre-removal cookie revoked. A no mints a new id and keeps the
address unique across live rows only; that stays safe because every revocation bump after 6d writes
both the address key and the `UserID` key ("Identity in the store"), so a pre-removal cookie that
resolves to the new id is still refused by the address record. What it changes: a returning member
gets their library state back.

**10. Does dubplate prune dated snapshots and the remote archive itself?** Recommendation: yes.
After each successful run the backup job keeps the newest `DUBPLATE_BACKUP_SNAPSHOT_KEEP` objects
under `snapshots/` and deletes `archive/<stamp>/` directories older than
`DUBPLATE_BACKUP_ARCHIVE_RETENTION`, each a 6b-i constant with a researched default, re-homed at 6c.
Pruning never removes the latest snapshot. Reason: a `local`, `sftp`, or `webdav` remote has no
lifecycle policy, so leaving pruning to one fails BS's universality bar. What it changes: remote
storage stays bounded with no provider configuration.

**11. Is the latest snapshot the only supported restore of store and tree together?**
Recommendation: yes. The remote tree is only ever the latest, with superseded files under the
archive, so an older snapshot pairs with a newer tree. An older snapshot restores the store alone
under an explicit flag, and the reconciliation report names every row without a file. Reason:
point-in-time recovery from the archive is a feature no ruling asks for. What it changes: an
operator who restores an older snapshot gets the store as it was and a list of what the tree no
longer holds.

**12. Does a disaster restore rotate the session and magic-link secrets rather than back up the
auth leaf?** Recommendation: yes. The restore verb's final report and the runbook require the
rotation. Reason: epoch and nonce records stay files in the auth leaf (R5S:533-539) and are not in
the backup, so a restored box would forget every revocation. Rotation invalidates every
outstanding credential, so revocations made after the last backup hold too, which a restored leaf
would miss. What it changes: every member logs in again after a disaster restore.

**13. When an env-seeded key is removed from the env, does its row revert to the default, logged?**
Recommendation: yes. **[CH]** Reason: an operator who deletes a key expects the default, and a row
that keeps the old seed is the silent outcome CH forbids (CH:85-86). The alternative fails the
load. What it changes: deleting a key from the age store is a supported way back to the default.

**14. Does `doctor` warn when the backup remote is not an rclone `crypt` remote?** Recommendation:
yes, a warning and never a refusal, read from `RCLONE_CONFIG_<NAME>_TYPE`. Reason: after 6d the
snapshot holds member addresses; a refusal fails adopters whose provider encrypts at rest. What it
changes: the runbook recommends a `crypt` remote, and `doctor` says when one is absent.

**Not asked: rung 9's concurrency model** (BS:760-764). A lease or a queue sits above the store's
serialized writers. The store's two constraints on rung 9: no write transaction spans a MusicBrainz
lookup or a file copy, and library mutations continue during a backup's sync unless rung 9's lease
pauses them, so restore's reconciliation report is the net.

---

## What rung 6 builds

| Deliverable | Pass | Proved by |
| --- | --- | --- |
| `internal/engine/store`: the open (inspect, write, serve modes), the leaf, the lock, three connection classes, the write method, migrations, snapshots | 6a | 6a criteria 1-15 |
| The migration set (job records, ledger, catalog skeleton), the ledger, and the schema-invariant test | 6a, then 6c | 6a 8-13, 6c 1 |
| `doctor`'s store row and `check`'s inspect report, the open's first production callers | 6a | 6a 14 |
| `internal/ops`: interface, wire types, error codes, privileges, the catalog, the job model | 6b-i | 6b-i 1-8 |
| `internal/ops/local`: the runner, leases, recovery, cross-process cancel; the write method's first caller | 6b-i | 6b-i 3-8 |
| `internal/backup`: the snapshot job, the rclone transport, the guards, the layout, pruning | 6b-i | 6b-i 9-17 |
| The backup integration class and its supersessions; the `backup` verb; `doctor`'s backup rows | 6b-i | 6b-i 18-21 |
| `internal/ops/opshttp` (unexported) with its authorizer, and the contract suite | 6b-ii | 6b-ii 1-4 |
| The `restore` verb: the intent record, the one-file swap, library restore, reconciliation | 6b-ii | 6b-ii 5-13 |
| Every table CH routes to rung 6 **[CH]** | 6c | 6c 1-4 |
| Registry metadata: home, apply mode, `sensitive`, bounds, correlated rules **[CH]** | 6c | 6c 5-6 |
| `internal/settings`: accessor, generation, seed rule, audit; 6b-i's constants re-homed **[CH]** | 6c | 6c 7-15 |
| The import verb, the users-backed roster, `Identity.UserID`, the dual-key epoch | 6d | 6d 1-8 |
| The auth thresholds re-homed with accessor reads at every call site | 6d | 6d 9-10 |
| `serve` opening the store: migrate, seed, then listen | 6d | 6d 11-12 |
| The store leaf's deployment: 5d's serve unit's grant, `deploy.sh`, the state-dir leg, the template | 6d | 6d 13-14 |

## What rung 6 does not build

- No HTTP route. The projection stays unexported, because backup is an admin operation and rung 14
  owns the authorization model (BS:831-837). Rung 14 exports it and mounts it on CH's two mounts,
  cookie and bearer (CH:107-110); the authorizer takes the mount's channel, so one projection serves
  both.
- No `ops/client`; it is post-launch (BS:668-672). The contract suite is the wire contract's
  regression net until then, not a parity proof.
- No Go accessor for a table whose first caller is a later rung. Rungs 7, 9, 13, and 14 add the
  accessors for collections, the catalog, quotas, the move-intent journal, decision records,
  preferences, overrides, and device credentials. The root-allocation function and the no-nesting
  check land with their first caller at rung 7 or rung 9. 6b-i's guard and 6b-ii's reconciliation
  read the skeleton through unexported queries in `internal/backup`.
- No hook column whose value a later rung computes (ruling 2), and no `collections` row version for
  CH's `if_version` check (CH:256-257): rung 14's shape edit is its first writer.
- No default collection row. The first rung that writes a collection (rung 7's uploader or rung 10's
  migrator, whichever lands first) creates it, owned by the first admin.
- No backup schedule and no schedule key. CH's class lists "backup targets and schedule" (CH:56);
  the schedule's consumer is a systemd timer, so rung 16 declares the key with the timer. rclone's
  credential rows reach the template at rung 16's convergence; 6b-i declares the class and documents
  the names. Both are erratum 14's carried-list lines for rung 16.
- No backup pings: rung 9's monitoring seam (`internal/config/keys.go:93-100`) reads the job's
  terminal state. No runbook: rung 16 authors it (CH:661-662), and each pass records the lines it
  owes (restore size, rollback cost, rotation, crypt, roster file deletion, and the empty-catalog
  guard gap) in its close.
- No re-authentication gate or sensitive-change notice (rung 14, CH:657-660), and no URL-valued
  store-homed key; the write path's validator is where rung 14's range check attaches (CH:132-133).
- No record from rung 4's or rung 5's leaves moves into the store except the roster (ruling 1);
  epoch and nonce records stay files (R5S:533-539). No last-admin guard of its own: 5c builds it,
  and 6d re-establishes it inside the write transaction.
- No reconciliation of `DUBPLATE_ADMIN_EMAIL` against `users.admin`. After 6d, `users.admin` is
  authoritative for authorization; the contact row's reconciliation is rung 14's (BS:833).
- No production open of the store before 6d except `doctor` and `check`'s inspect open. `serve`
  opens the store at 6d, the pass that also grants its leaf, so a 5d deploy built from a master
  carrying 6a through 6c never has `serve` touch an ungranted leaf.

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
`store`, for the users-backed roster behind the roster's existing constructor; it never imports
`settings`, because `cmd/dubplate-server` injects each auth threshold as an accessor function.

Exported surface follows its first caller. 6a exports the open, whose production callers are
`doctor` and `check`, and tests the write method and the serve mode in-package. 6b-i exports the
write method with `ops/local`, its first caller. 6d exports the serve mode and the seed rule with
`serve`, their first caller.

---

## The store

### Layout, ownership, and the lock

The store lives under one leaf of `DUBPLATE_STATE_DIR` named `store`, following the `shadow` and
`auth` leaves (`internal/engine/shadow/store.go:39-49`, `internal/auth/store.go:12-40`). The leaf
holds `store.lock`, a `db/` directory with the database and its `-wal` and `-shm` files,
`snapshots/` (pre-migration and pre-restore copies), `staging/`, `leases/`, and at most one
`restore-intent` record. The lock file sits outside `db/` and nothing ever renames it, so every
process locks the same inode. The leaf is 0700. SQLite creates files at 0644 under a 0022 umask and
gives `-wal` and `-shm` the main file's mode, so the store pre-creates the database file and every
snapshot target with `O_CREATE|O_EXCL` at 0600 (SQLite permits an empty target for `VACUUM INTO`).

Every writing open checks for a `restore-intent` record first. Only then does it remove staging
leftovers, and it never removes a file an intent record names.

Every verb that opens an existing leaf refuses when its effective uid differs from the leaf's owner,
naming the owner and the fix (`runuser -u <owner> -- dubplate-server ...`); a root-run verb would
otherwise leave a root-owned `-shm` or database that `serve` cannot open. An absent leaf reports
absent with no uid check. After 6d, `deploy.sh`, which connects as root (`scripts/deploy.sh:114`),
runs every store-opening verb through `runuser -u <owner>`.

The lock is an advisory `flock` on `store.lock`, released when the process exits. `backup` and the
`member` verbs, `import` included, hold it shared; `restore` holds it exclusive; `check` and
`doctor` take it shared and non-blocking; `serve` follows the procedure in "Migrations." `flock`
cannot name a holder, so a refusal says "held by another process" and lists the pids `/proc/locks`
shows for the lock file's inode when that file is readable.

WAL needs every process on one host (sqlite.org/wal.html section 1). The store probes the leaf's
filesystem with `statfs`, refuses NFS, CIFS, SMB2, and 9p, and warns on FUSE, because sshfs and
rclone mounts are remote while mergerfs is local and common on home servers.

Rung 6 declares three test seams, each justified because the real condition cannot be produced in
a unit test: the `statfs` probe, a package variable (no test can mount NFS); the migration runner's
source, an `fs.FS` parameter of an unexported function, which is ordinary dependency injection
(6a's embedded set holds no failing or rebuilding migration); and restore's step hook, an
unexported function variable called between flow steps, which a test sets to exit a re-executed
test binary at a named step, so each crash is a real process death.

### Three open modes

The **inspect mode** is read-only (`mode=ro`), takes the lock shared and non-blocking, never
creates, and never migrates. It reports absent, restore in progress, older (naming the pending
migrations), newer (naming both versions), or current. It does not leave the leaf untouched: a
read-only WAL open needs "the -shm and -wal files already exist and are readable" or "write
permission on the directory containing the database so that the -shm and -wal files can be created"
(sqlite.org/wal.html section 5). The invariant it keeps is that the database file's bytes and mode
never change; `-wal` and `-shm` may appear, owned by the leaf owner. `check` and `doctor` use it.

The **write mode** creates an absent store only in the `member` verbs; `backup` refuses an absent
store, because a mistyped state directory would otherwise upload an empty dated snapshot that
"latest" then selects. A write-mode open that finds a `restore-intent` record refuses and names
`restore --resume`. A write-mode open refuses a store whose version is not the binary's, naming
both versions and "start serve to migrate."

The **serve mode** is the only one that migrates or creates a store at start (from 6d). It accepts
one intent state, `reconcile-pending`, which "The restore verb" describes.

### Connections and transactions

2.0 hook 9 asks for one writer, WAL with a `busy_timeout`, and a seam a later projector joins so
its delta commits with its event (S20:387-389). The store opens three connection classes over one
file:

- The **write pool** holds one connection with `_txlock=immediate`, so every write transaction
  takes the write lock first and the busy timeout governs the wait.
- The **read pool** opens with `_query_only=1`, so no caller writes outside the write method. A
  write there would fail at once, because the busy handler does not run on a deferred read-to-write
  upgrade (verified; see "Behaviors a task relies on").
- The **maintenance connection** is a separate `*sql.DB` opened for one job, a snapshot or a
  foreign-keys-off table rebuild, and closed afterward, so it never returns to a pool. Its DSN
  carries `_synchronous=FULL`, `_busy_timeout`, and `_foreign_keys=1`; a rebuild turns foreign keys
  off after open. `VACUUM INTO` fails on a `query_only` connection and inside a transaction, and
  running it on the write pool would stall every in-process writer for the snapshot's length.

Every per-connection setting rides the DSN, because `database/sql` opens connections at will, as
modernc's shorthand keys, which the driver validates before applying any: `_busy_timeout`,
`_foreign_keys=1`, `_journal_mode=WAL`, `_synchronous=FULL`, and `_query_only` or `_txlock` per
class (`sqlite.go:307-391` at v1.59.0). `journal_size_limit` has no shorthand and rides `_pragma`,
which the driver executes verbatim and cannot pre-validate (`sqlite.go:305-306`), so 6a proves
every setting reads back on a fresh connection of every class.

`synchronous` is `FULL`. Under `NORMAL`, "a transaction committed in WAL mode ... might roll back
following a power loss" (sqlite.org/pragma.html), which breaks every pattern where file work
follows a commit: CH's move intent, the import record, a job record. `VACUUM INTO` fsyncs its
output only "if the PRAGMA synchronous setting of the original database is NORMAL or FULL"
(sqlite.org/lang_vacuum.html), and the setting that counts is the maintenance connection's.
`journal_size_limit` bounds the WAL's size after a checkpoint, and every snapshot ends with
`wal_checkpoint(TRUNCATE)`, because a checkpoint "does not normally truncate the WAL file"
(sqlite.org/wal.html).

The store exposes one way to write: a method that runs a callback inside one `IMMEDIATE`
transaction. That callback is hook 9's seam. Its contract:

- It commits only when the callback returns nil and no statement inside failed, and rolls back
  otherwise. SQLite may undo one statement and continue the transaction on `SQLITE_FULL`,
  `SQLITE_IOERR`, `SQLITE_INTERRUPT`, or `SQLITE_NOMEM` (sqlite.org/lang_transaction.html section
  3), so a swallowed statement error never commits. "Already rolled back" is that error.
- Every read a write decides on runs inside the callback, never on the read pool.
- It requires a context with a deadline, since an in-process wait for the one write connection is
  bounded by the context and never by `busy_timeout`. A nested call fails fast.
- Its first statement reads `user_version` and fails closed on a value other than the binary's.
- `SQLITE_FULL` surfaces as a named disk-full error.

The busy timeout is a wall threshold with a researched default, declared in rung 6's own file.

### Migrations

2.0 hook 7 asks that the mechanism be chosen with an evolving schema in view (S20:383-384). The
runner is hand-rolled over an `fs.FS` of numbered SQL files and `PRAGMA user_version`, as S20's
prior-art row records (S20:435); production passes the embedded set, and tests pass synthetic sets.
"Prior art" re-scores the alternatives.

**Only `serve` migrates, and it never holds a store connection across a lock conversion.**
`flock(2)` says: "Converting a lock (shared to exclusive, or vice versa) is not guaranteed to be
atomic: the existing lock is first removed, and then a new lock is established. Between these two
steps, a pending lock request by another process may be granted." So a waiting `restore` can win
any conversion, and `serve` must hold no handle to the database file when it might. The procedure:

1. Take the lock shared, open, and read the version. If it is current, keep the lock and continue.
2. Otherwise close every connection, release, and take the lock exclusive, waiting at most
   `DUBPLATE_STORE_MIGRATION_WAIT` (a wall threshold with a researched default) and logging the
   holders at an interval. On expiry, exit nonzero naming the holders.
3. Holding it exclusive, reopen, re-read the version (a restore may have run in the gap), and
   migrate.
4. Close every connection, convert to shared, and return to step 1, so a restore that won the
   conversion is met by a fresh open and a fresh version check.

No start timeout bounds that wait. 5d's serve unit chooses its own `Type=`; under `Type=simple`,
systemd "will consider the unit started immediately after the main service process has been
forked off," and `TimeoutStartSec=` bounds only a daemon that must "signal start-up completion"
(systemd.service(5)). The real exposure is a timer-run backup holding the lock shared for a whole
sync, hours on a first run, after a deploy that brings pending migrations. So `serve`'s wait is the
threshold above, after which systemd's restart policy retries. 6d's `deploy.sh` also runs the new
binary's `check` before restarting `serve`; when it reports the store older and the lock held, it
names the holder and asks the operator to wait for or stop the backup. The login outage during the
wait is a runbook line.

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
- A store is non-empty when `user_version > 0`. Before migrating a non-empty store the runner
  checks free space for one full copy plus margin, writes `pre-migrate-v<N>.db` through the
  maintenance connection into `staging/`, verifies it, renames it into `snapshots/`, and fsyncs the
  directory by hand (ruling 7). Any failure refuses the migration and names the snapshot path.
- A rebuild migration that needs foreign keys off runs on the maintenance connection, because
  enforcement "may only be enabled or disabled when there is no pending BEGIN or SAVEPOINT"
  (sqlite.org/pragma.html). It runs `PRAGMA foreign_key_check` before commit, refuses to commit on
  any row, and closes the connection.
- Every migration file ships with a fixture that populates the schema as of its version, and the
  invariant test fails on a migration with no fixture. 6a embeds three migrations in order: the
  ledger (which records itself), the job records, and the catalog skeleton. Its populated-fixture
  proof migrates the job-records fixture forward through the skeleton.
- Files are append-only once merged. A lane that rebases onto another lane's migration renumbers
  its own, and the rebase is non-fast-forward, so merge ritual step 5 runs.

Rolling back to an older binary meets the newer-store refusal. The way back is that binary's own
`restore --from-file` of its matching pre-migration snapshot, which accepts a newer current store
("The restore verb," step 3) and loses every write since the upgrade. The runbook states the cost.

### Schema invariants

One test applies every migration to an empty database and to each migration's populated fixture,
then inspects the result through the catalog (`pragma_table_list`, `pragma_table_info`,
`pragma_foreign_key_list`, `pragma_index_list`), never by parsing SQL. Each rule is the catalog
fact the test checks, and one violating fixture per rule proves the rule can fail for its named
reason:

1. Every table is `STRICT` (`pragma_table_list.strict = 1`). A column named `path` or ending in
   `_path` is declared `BLOB`. A `STRICT` table refuses a Go `string` bound into a `BLOB` column,
   which a plain `BLOB` declaration does not: a TEXT and a BLOB copy of the same bytes both pass a
   UNIQUE index in an ordinary table.
2. Every foreign key resolves to an existing table and column, and its parent columns are the
   primary key or carry a UNIQUE index; SQLite accepts a key to a non-unique parent and then fails
   every insert with "foreign key mismatch."
3. A table with a `kind` column, or whose name ends in `_events`, `_audit`, or `_journal`, carries
   `at` and `kind`, both `NOT NULL`. This is 2.0 hook 4 (S20:374-375).
4. MusicBrainz identifiers are named `mb_<entity>_id`. A column named `release_id`,
   `recording_id`, `artist_id`, `release_group_id`, `track_id`, or `work_id` that is not a foreign
   key is refused (EVO:575-578, the #604 class).
5. No table pairs a foreign key column with a TEXT column named `name`, `key`, or `attr` and a column
   named `value` (U4's attribute bag).
6. An entity table's primary key is a `BLOB` column named `id`, and no column named `id` or ending
   in `_id` is an `INTEGER PRIMARY KEY` (S20:166-167). Keyed and singleton tables are exempt by a
   list named in the test: `settings` (keyed by key), the ledger (keyed by version), the settings
   generation (one row), and the group quota total (one row). A store test proves the id function
   mints 16 random bytes, never equal to a rowid.
7. No foreign key to `users` or `collections` declares `ON DELETE CASCADE`, since CH requires
   tombstones and audit history to survive (CH:519-521).

What a column means stays a reviewer's question: a path column not named as one, or a source
identifier under another name, passes the test. Every rung 6 migration task and every later
rung's carries that as a `diff-reviewer` checklist line, not a claimed proof.

### The schema rung 6 lands

Pass 6a lands the job records, the ledger, and the catalog skeleton. Pass 6c lands CH's routed
tables **[CH]**. Table names are working names a plan may change; the contents are the contract.

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
  `sensitive` are registry metadata, not columns, which reads CH:638 together with BS's registry
  erratum (BS:39-40).
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
move rather than a rewrite of every track row.

### Paths (U3)

Every path is relative and byte-preserving from the first schema (BS:662-664, EVO:356-366): a
collection's `root_path` is library-relative and a track's `rel_path` is relative to that root. A
path crosses the store boundary as `[]byte`, and invariant 1's `STRICT` rule enforces it in SQLite,
not only in Go. Paths compare bytewise. The store refuses a path that is absolute, has an empty,
`.`, or `..` component, or carries NUL. The build compiles with `SQLITE_LIKE_DOESNT_MATCH_BLOBS`, so
`LIKE` never matches a path; rung 9's query verb says how it searches paths. Rung 7's gate rules on
non-UTF-8 filenames and canonicalizes symlinks once (EVO:363-366, EVO:600-602); whatever it admits,
the store returns unchanged.

### Per-track metadata (U4)

Typed columns added by migration, one per field a built-in or the engine owns, with no attribute
bag (invariant 5). Adding a built-in is a migration and a typed accessor in the built-in's own
rung. EVO weighs it this way ("the weak-typing complaint corroborates U4," EVO:489-494), and the
cost is the no-plugin-ABI fork OS settled (OS:338-352).

---

## The ops seam

### Vocabulary and privileges

`internal/ops` defines one interface. Its request and response types are JSON-serializable and
are the source of the API's wire types (OS:230-236). The acting identity is a method parameter,
never a field of the wire type: a verified `identity.Identity` for a network caller, or a local
operator named by the operating-system user for a CLI verb (R5S:1141-1143). The contract suite
proves that a body carrying an actor field is ignored.

Every catalog entry declares a privilege from a closed set, `member` or `admin`, with no zero
value. The error vocabulary is a small set of sentinels, each with a stable wire code:
unavailable (with a reason), not found, conflict, invalid request, forbidden, canceled, and
internal. Callers match with `errors.Is` in Go and by code over HTTP.

### The job model

Long-running operations return a job handle, an opaque random value (OS:241-248). States are
queued, running, succeeded, failed, canceled, and interrupted; a terminal state never changes.
Progress is polled, and the interface leaves room for a local push layer (OS:243-245).

Liveness is per process, not per kind. Every process that runs jobs holds a lease: an exclusive
`flock` on `leases/<instance-id>`, held for the process's life. The runner takes its lease before
it inserts any record, and every record names its owner instance. An exclusive kind (backup is
one) also holds `leases/kind-<kind>` exclusive for the job's life, so a second start returns a
typed conflict naming the running job. Recovery runs at every write-mode open and before any
exclusive start:

| Record | Owner lease | Outcome |
| --- | --- | --- |
| queued or running | held | left alone, whatever its kind |
| queued or running | free | marked interrupted; recovery then removes the lease file it acquired |
| terminal | any | unchanged |
| a late terminal update to a record already marked interrupted | any | refused and logged |

Cancellation is an explicit operation that writes `cancel_requested` on the record; it never
relies on a context crossing a process. The owning runner polls the column at its progress points
and cancels the job's context. Canceling a queued job moves it straight to canceled; canceling a
terminal job is a no-op that reports the state; a running job that finishes first ends in its own
state, and the response says which.

A job runs in the process that started it. The CLI `backup` verb starts a backup through
`ops/local`, polls it, cancels it on SIGINT or SIGTERM through the explicit operation, and exits
with a distinct code per outcome: 0 for succeeded, and distinct nonzero codes for failed,
canceled, unavailable, and conflict. The 6b-i plan fixes the numbers against the existing verbs'
convention and records them in the verb's help, because rung 16's timer and rung 9's failure ping
read them. Terminal records are pruned at every write-mode open once older than
`DUBPLATE_JOB_RETENTION` (a 6b-i constant, re-homed at 6c **[CH]**); pruning never removes the
newest successful record of a kind, which `doctor`'s last-backup age reads.

### Operation availability

The catalog lists every operation with its availability, the reason when unavailable, and its
privilege. It is the seam's mechanism against ad-hoc parity opt-outs and never shares a term with
the streamer seam's driver capabilities (OS:273-277). Every asymmetry it reports is named in a spec
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
body), forbidden (a denying authorizer), canceled (canceling a running job), and internal (the
store closed under the implementation, a production-real trigger with no test-only constructor).
Its scope is BS's: the wire contract's regression net, not a parity proof (BS:668-672).

---

## Backup and restore

### The backup job

A backup job syncs the library tree, then writes a store snapshot, then sends the snapshot, then
prunes. The snapshot comes second so every file the catalog named when the sync began is on the
remote or in its archive. The pair is still not atomic: a change committed during the sync can
leave a row whose file is not on the remote, which restore's reconciliation report names. The
snapshot records the run id and the sync's completion time in its object metadata.

2.0 hook 10 forbids a byte copy of a live WAL database (S20:390-391; sqlite.org/howtocorrupt.html
section 1.2). The snapshot is `VACUUM INTO` on the maintenance connection into a pre-created 0600
file in `staging/`, which yields "a consistent snapshot of the original database"
(sqlite.org/lang_vacuum.html). It meets the hook's intent though the hook names the backup API or a
checkpoint-then-copy; the backup API (`conn.go:1208-1219`) stays the recorded alternative. The job
checks free space for one full copy first, runs `integrity_check` on the result, renames it, fsyncs
the directory by hand (`atomicfile` has no rename of an existing file; `WriteFile` does not fsync
the directory, `WriteStream` does, at `internal/atomicfile/atomicfile.go:82-89`), and ends with
`wal_checkpoint(TRUNCATE)`.

The remote layout under `DUBPLATE_BACKUP_REMOTE` is `library/` for the tree, `archive/<stamp>/`
for files the sync replaced or deleted (rclone requires the backup directory on the same remote and
not overlapping the destination), and `snapshots/<stamp>.db`. Stamps are UTC and sort
lexicographically. After a successful run the job prunes the remote by ruling 10 and deletes local
snapshots by ruling 7.

The guards run before rclone, against the skeleton 6a landed. A missing library root aborts. So
does a tree whose regular-file count falls below `DUBPLATE_BACKUP_GUARD_FRACTION` (a 6b-i constant,
re-homed at 6c) of the catalog's track count. With zero catalog tracks, every box's state until
rung 7 or rung 10 writes one, an empty root aborts, and rclone runs with `--max-delete 0`, so the
remote is never archived while the catalog knows no track. Otherwise rclone runs with
`--max-delete` sized from the track count. rclone's manual: "This tells rclone not to delete more
than N files. If that limit is exceeded then a fatal error will be generated and rclone will stop
the operation in progress." Until the catalog holds tracks, a non-empty wrong mount passes the
guards, which is a runbook line. The salvage guard checked only the top level
(`_salvage/internal/backup/backup.go:22-34`, `:125-128`); a wiped mount under surviving collection
directories passes it, so the fresh review reverses that argument. Every rclone invocation runs
under `exec.CommandContext`; the 6b-i plan quotes `os/exec` on what cancellation sends and how
`WaitDelay` bounds the wait, and rclone's documentation on sync, `--backup-dir`, and exit codes.

The gate runs a real `rclone` against a `local` backend in a temporary directory, and a missing
`rclone` fails the test, never skips it; 6b-i's close routes that prerequisite to CLAUDE.md's Gate
section. A recording stand-in whose exit code each case sets covers the remote-failure and SIGINT
cases; it is never an always-succeeds fake.

Backup does not preserve hardlinks, so a restored library is larger than its source (CH:410-414).
Rung 16's relocation plan states how the next backup handles a moved subtree, by rclone's rename
tracking or a planned re-upload.

### The restore verb

`restore` is offline (ruling 6). Its interface: `--snapshot <stamp>` or `--snapshot latest` names
an object under `snapshots/`, or `--from-file PATH` names a local file such as a pre-migration
snapshot; `--library` also restores the tree; `--into-nonempty` permits a non-empty library root;
`--older` permits a remote snapshot other than the latest, store only and never with `--library`
(ruling 11); `--resume` continues a recorded restore. Its flow:

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
trusting its cache. `PRAGMA data_version` is per connection and cannot serve.

The registry gains four pieces of metadata per key: its home (wall or store), its apply mode
(immediately, next operation, or restart), the `sensitive` marker (CH:128-129, CH:638), and its
bounds as a validator that takes a value rather than an environment. It also gains a registry-level
list of correlated rules. Today's bounds live in the source's `Validate` hook and read `environ`
(`cmd/dubplate-server/auth.go:230-280`), which a store write cannot supply. A key registered
without a home stays on the wall. A store-homed key has no wall getter: the seed rule is the only
reader of its env value, so the compiler forbids a consumer bypass. `doctor` shows each store-homed
key with its value, source, apply mode, and marker.

### The seed rule

Only `serve` applies the seed rule, at start after the wall loads and the store opens and before
any consumer reads a setting (6c builds it, 6d wires it into `serve`). It applies every seed in one
`IMMEDIATE` transaction and validates each key's bounds and every correlated rule against the
post-seed set, so raising the link expiry and the nonce retention together in one deploy succeeds.
Every other verb, `check` included, evaluates the rule read-only and reports a disagreement without
failing, so a timer-run backup never fails over an auth threshold it does not read. A seed or
re-seed writes an audit row with actor `system` and channel `env-seed`. A `default` source stores
no value; reads resolve the registry's current default, so a re-researched default reaches every
existing store. The closed table:

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

Pass 6c re-homes 6b-i's five constants as the first store-homed keys, each with a researched default
recorded where it is defined (BS:236-248): `DUBPLATE_JOB_RETENTION`,
`DUBPLATE_BACKUP_GUARD_FRACTION`, `DUBPLATE_BACKUP_SNAPSHOT_KEEP`,
`DUBPLATE_BACKUP_ARCHIVE_RETENTION`, and `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`.

### The seed handoff

Fifteen thresholds are registered through the env today: rung 4's two shadow keys, pass 5a's three
serve keys, and rung 5's ten auth keys (`cmd/dubplate-server/auth.go:99-200`). All ten auth values
resolve once into `authConfig` (`auth.go:402-416`) and reach the auth core as construction
arguments, which is the long-lived struct CH forbids. So the apply-mode column is a **required
behavior** 6d builds, not a description of today's code: 6d replaces `authConfig` with accessor
functions `cmd/dubplate-server` injects at each call site 5c and 5d wire.

| Key | Registered | Home after 6d | Required apply mode | Rule for state already issued |
| --- | --- | --- | --- | --- |
| `DUBPLATE_MAGIC_LINK_EXPIRY` | `auth.go:99` | store | next operation | a link keeps the expiry it was minted with (`MintLink`, `internal/auth/token.go:118`) |
| `DUBPLATE_SESSION_LIFETIME` | `auth.go:105` | store | next operation | applies at each session's next refresh (`RefreshSession`, `session.go:132`) |
| `DUBPLATE_SESSION_ABSOLUTE_CAP` | `auth.go:111` | store | immediately | verification reads the live cap (`VerifySession`, `session.go:206`) at every request, in both directions |
| `DUBPLATE_NONCE_RETENTION` | `auth.go:117` | store | next operation | the next sweep uses it; its correlated rule with the link expiry moves from the source's `Validate` hook to the registry's list |
| the four rate limits | `auth.go:166-184` | store | restart | the limiters are built once (`NewLimiter`, `internal/auth/limit.go:62`); a live apply is rung 14's option |
| `DUBPLATE_AUTH_REQUEST_BODY_CAP` | `auth.go:190` | store | next operation | the next request reads it |
| `DUBPLATE_UNKNOWN_ADDRESS_ANSWER` | `auth.go:196` | store | next operation | the next answer reads it |
| `DUBPLATE_SERVE_PREFIX`, `_BIND_ADDRS`, `_PORT` | `serve.go:35-59` | wall | restart | unchanged; read before the store opens |
| `DUBPLATE_SHADOW_FLOOR`, `_BUDGET` | `shadow.go:73-88` | wall (ruling 3) | restart | unchanged; the shadow unit is not granted the store leaf |

6d declares both lifetimes and the absolute cap `sensitive`, so rung 14's re-authentication gate
covers them from its first commit. None of the ten auth thresholds appears in the environment
template, which carries only the shadow pair among thresholds
(`env/dubplate-server.env.template:110-117`); the handoff mints nothing and removes no template row
except ruling 1's `DUBPLATE_ROSTER_FILE`. The plan re-reads each cited line at its staleness
pre-task, since 5c and 5d may move them.

---

## Identity in the store

Pass 6d opens only after rung 5's pass 5d merges, because every file it edits is rung 5's until
then. `web-auth-security-reviewer` runs inside its chain.

**The import** follows ruling 1. Its states, each with its report:

| Roster file and store | Report |
| --- | --- |
| the file absent or unreadable | refused, naming the file |
| the file malformed | refused, naming the file and the parse error |
| two entries with one address | refused, naming both entries |
| an entry with no `streamer_username` | refused, naming each such entry (ruling 8) |
| no admin in the file or the table | refused, naming the rule |
| an entry colliding with a different existing member on address or handle | refused, naming both |
| an import record already present | refused, naming the record's time and actor |
| every entry valid, matching rows merged | imported; the record written |

That is eight states, seven refusals and one success; 6d criterion 2 proves each refusal by its own
fixture. 6d removes `deploy.sh`'s pre-create of the roster leaf (R5S:198-200, PA:218) and leaves the
file in place; the runbook tells the operator to delete it after verifying the import. Rolling back
past 6d needs the old template and the old file, and members added since then are re-added through
the verb afterward.

**The resolution point.** R5S ruled that the store moves behind the roster's interface without
touching a call site (R5S:194-196, R5S:287-292, PA:209), and `internal/auth/roster.go:113-116`
already frames the substitution as one constructor. The resolution point (`Lookup`,
`roster.go:286`) reads the users table. `identity.Identity` gains a `UserID` field carrying the
surrogate id, which per-user rows key on. `Principal` stays the streamer handle: rung 12 injects
the principal into the proxy header (R5S:1243-1245), and Navidrome creates an account for any
unknown proxy user, so an opaque id there would orphan every member's history. The write path mints
the id for a member added by the page or the verb. The rule 5b-ii ruled, that `AddMember` refuses
the admin flag on every channel except 5c's CLI grant, carries over, and the import is the one other
admin channel, once, from the file. 5c's last-admin guard and the `streamer_username` collision
check run inside the write transaction, so a concurrent self-demotion of the last two admins leaves
one. The member-change audit record stays the log record rung 5 built (`roster.go:316`); CH routes
no member-audit table.

**The epoch** gets no bulk move. Records are named `sha256(kind, value)`
(`internal/auth/store.go:105-114`), which is one-way, so a re-key cannot reach an address removed
before 6d, and a copy races a concurrent bump from another process (`EpochStore.mu` is in-process
only, `internal/auth/epoch.go:66-72`). Instead, the revocation check reads both records, the one
keyed by the session's address and the one keyed by its resolved `UserID`, and takes the later
epoch, permanently. Every bump after 6d writes both keys, so a removal is refused through the
address record even when the address later resolves to a different id. Every epoch write takes an
advisory lock on the epoch directory, reads the stored value, and writes only a later one, so no
process moves an epoch backward. The revocation notification (R35) keeps its address argument,
because sessions carry the address, and gains the `UserID`. The records stay files (R18). This
closes the reviews' scenarios: a rollback to a 5d binary, a restored pre-6d snapshot, a re-added
member under either answer to ruling 9, and a crash mid-way, since there is no mid-way.

**`serve` and the store.** 6d wires `serve` to the store in `cmd/dubplate-server/serve.go`: the
serve-mode open and its migration procedure, then recovery, then the seed rule, then the listener.

---

## The 2.0 forward hooks

Per CLAUDE.md's build-for-2.0 rule, each pass states per task whether it lands a hook or is
neutral; a hook is built by the rung that first consumes it.

| Hook (S20) | Rung 6's position |
| --- | --- |
| 1, audio key and file hash (:363) | **Deferred** to rung 7's migration by ruling 2 (S20 erratum owed). Both are alternate keys, not primary keys. |
| 2, folded sort key (:370) | **Deferred** to rung 9's migration by ruling 2 (S20 erratum owed). |
| 3, art blob (:372) | **Deferred** to rung 9's migration by ruling 2 (S20 erratum owed). |
| 4, timestamped events (:374) | **The rule lands at 6a** as invariant 3; `settings_audit` (6c) is its first table. |
| 5, stable identity (:376) | **Lands at 6a and 6d.** 6a creates users; 6d mints the id, adds `UserID`, and reads both epoch keys. |
| 6, universal driver seam (:379) | Neutral. |
| 7, schema migrations (:383) | **Lands at 6a.** |
| 8, collection writer (:385) | **Lands at 6a** as `writer` with its domain, epoch, and generation **[CH]**. |
| 9, transaction model (:387) | **Lands at 6a.** One writer, WAL, `FULL`, the write method. |
| 10, live-file backup consistency (:390) | **Lands at 6b-i**, by `VACUUM INTO`. |
| 11, gains (:392) | **Deferred** to rung 9's migration by ruling 2 (S20 erratum owed). |
| 12, portable layout | Neutral. Rung 7. |
| 13, sort tags; 14, ID3 version | Neutral. Rungs 8 and 9. |
| 15, configuration homes (:401) | **Lands at 6c** for the routed schema, the settings home, and the marker; the two mounts and the re-authentication gate are rung 14's **[CH]**. |

---

## Prior art

Scored on CLAUDE.md's four criteria. Versions and dates are from `go list -m` in a scratch module
on 2026-09-23; licenses from each module's `LICENSE`; importer counts from pkg.go.dev. The
mechanics review's measurements are summarized in the fold record.

| Entry | Role | License | Quality | Maintenance | Adoption | Verdict |
| --- | --- | --- | --- | --- | --- | --- |
| `modernc.org/sqlite` | The driver | BSD-3-Clause | Transpiled C under a hand-written `database/sql` layer. Its DSN parser validates every shorthand key before applying any (`sqlite.go:290-305`); `_pragma` values are executed verbatim and are the one exception (`:305-306`). SQLite 3.53.4 with FTS5, snapshots, and `LIKE_DOESNT_MATCH_BLOBS`. | v1.59.0, 2026-09-15; ten releases since 2026-05-10. 25 open issues, oldest 2020-09-05. Bus factor 1. | 3,518 importers | **Adopt**, meeting S20's condition: a gate test proves the FTS5 trigram tokenizer and WAL with a reader held across a commit. |
| `github.com/ncruces/go-sqlite3` | The fallback | MIT | A Wasm build of SQLite translated to Go by wasm2go; no wazero in its requirements | v0.35.6, 2026-09-23 | 96 importers; gonic uses it | **Record as the fallback** if the trigram or WAL test fails. Still v0. |
| `github.com/pressly/goose/v3` | Migrations | MIT | Readable, `embed.FS` provider | v3.28.0, 2026-09-02; 133 open issues, oldest 2019-01-10 | 1,725 importers | **Decline.** Graph pruning keeps its cost small (four modules, about 2.3MB of binary), but an older binary's `Provider.Up` on a newer database returns no error, the silent-ahead case the runner must refuse, and it keeps its own version table. |
| `github.com/golang-migrate/migrate/v4` | Migrations | MIT | Driver-per-package | v4.20.1, 2026-09-09 | 5,219 importers | **Decline.** Its SQLite driver is tested against modernc v1.18.1 and pulls cgo `mattn/go-sqlite3`. |
| `zombiezen.com/go/sqlite` | A non-`database/sql` API | ISC (not re-read) | Ergonomic | v1.4.2, 2025-05-23 | not measured | **Decline.** It leaves `database/sql` for a second vocabulary. |
| `github.com/mattn/go-sqlite3` | The common driver | MIT | Mature | v1.14.52 | very high | **Decline** on OS's no-cgo fork (OS:412-413). |
| rclone, as a subprocess | Backup transport | MIT | The estate's proven tool | v1.74.3 on the workstation | the estate runs it | **Adopt** as a probed subprocess. `librclone` in-process is declined: rclone's whole dependency tree for one verb. |
| SQLite `VACUUM INTO` | Snapshots | public domain | Guarantee quoted above | SQLite's own | universal | **Adopt**; the backup API is the recorded alternative. |
| `internal/atomicfile` | The restore-intent record | in-repo | house bar | in-repo | rungs 4 and 5 | **Adopt** for records; renames of SQLite-written files fsync by hand. |
| `_salvage/internal/backup` and `restore` | Guards and flow | salvage tier | salvage | frozen | ran on the box | **Adopt the missing-root guard**; the empty-top-level argument is reversed. Import nothing. |
| Navidrome's schema | Users and per-user tables | GPL, behavior only | | | | **Consult** at 6a's and 6c's plans for behavior only; no code. |

The other entries CLAUDE.md names (OpenSubsonic, gonic, taglib, mewkiz/flac, Picard, chromaprint)
have no rung 6 surface.

---

## Risk model

The store holds the only record the greenfield server keeps of what it admitted. Its failure
classes, ordered by cost, each with the guards named above:

| Failure | Guarded by |
| --- | --- |
| data loss by restore | the exclusive lock, the one-file swap after a truncating checkpoint, the intent record checked before staging cleanup, a crash injected at each step |
| data loss by backup (a wiped or wrong mount archives the remote) | the missing-root guard, the store-aware count guard, `--max-delete` (0 with an empty catalog), ruling 10's retention floor |
| a migration under a live older process, or a restore in a lock conversion | `serve`-only migration under the exclusive lock, no connection held across a conversion, the version check in every write transaction |
| a bad migration | per-migration transactions, the ledger, per-version snapshots (ruling 7), the newer-store refusal, the foreign-key check, populated fixtures |
| foreign keys silently off; lost durability | the DSN, the maintenance connection that never returns to a pool, the readback proof; `synchronous=FULL` on every class |
| disk full | a free-space check before every full copy, `journal_size_limit`, the truncating checkpoint, the named error, local snapshot deletion, `doctor`'s WAL, free-space, and snapshot rows; CH's upload floor stays rung 7's |
| a live job marked interrupted; settings flap | per-process leases taken before insert; `serve`-only seeding in one transaction |
| revocation lost at 6d; a root-owned store file | the permanent dual-key read, bumps to both keys, and forward-only writes; the uid check |

---

## Planning-miss prevention

Every 5b-ii fix round was an ambiguity its plan left open (CLAUDE.md, "Planning-miss prevention").
The equivalent questions for rung 6 are closed here or in the rulings.

### Behaviors a task relies on, quoted rather than recalled

| Behavior | Where it is settled |
| --- | --- |
| modernc's shorthand DSN keys, their validation, and the `_pragma` exception | quoted above from v1.59.0; the plan re-quotes |
| `database/sql` opening connections at will | the DSN rule; the plan quotes the package documentation |
| the busy handler on a deferred upgrade | verified: it does not run; `SQLITE_BUSY` returns at once ("if SQLite determines that invoking the busy handler could result in a deadlock, it will go ahead and return SQLITE_BUSY") |
| foreign keys not toggling inside a transaction; `user_version` and DDL rolling back; no bound parameter in a pragma | quoted above; the last two verified by the mechanics review and repeated by 6a's proofs |
| `VACUUM INTO`'s consistency, its sync condition, and its refusals inside a transaction and on `query_only` | quoted above and verified |
| `synchronous=NORMAL` losing a commit on power loss | quoted above |
| WAL on one host; a read-only open needing existing `-shm` and `-wal` or a writable directory | sqlite.org/wal.html sections 1 and 5, quoted above |
| the hot-journal rule; statement-level undo | howtocorrupt.html section 1.4; lang_transaction.html section 3 |
| a `BLOB` column storing TEXT outside `STRICT` | verified by the mechanics review |
| `flock` conversion not being atomic | `flock(2)`, quoted above |
| `Type=simple` and `TimeoutStartSec=` | systemd.service(5), quoted above; 5d's unit type is re-read at 6d's staleness pre-task |
| `--max-delete`; `RCLONE_CONFIG_<NAME>_*` naming | rclone's manual, quoted above |
| `exec.CommandContext` cancellation and `WaitDelay` | the plan quotes `os/exec` |
| rclone sync, `--backup-dir`, exit codes | the plan quotes rclone's documentation |
| `flock` release on exit and `/proc/locks` | the plan quotes `flock(2)` and `proc(5)` |

### Every state a check can meet, with its report

| Check | States and reports |
| --- | --- |
| inspect mode | absent (no uid check), restore in progress, older (pending migrations named), newer (both versions), current, lock contended ("held by another process" with pids), network filesystem (refused), FUSE (warned), wrong uid (refused naming the owner) |
| write mode | absent: created by `member`, refused by `backup`; zero bytes or not a database: refused; version differs: refused naming both; intent record present: refused naming `--resume`; ledger mismatch: refused naming both |
| serve mode | current: continue; older: the four-step procedure; lock wait expired: exit nonzero naming the holders; snapshot write or free space fails: refused naming the path; `reconcile-pending` intent: migrate, reconcile, remove |
| job recovery | the four rows of the recovery table |
| backup | flag off: unavailable, no record; store absent: refused; root missing, root empty with no tracks, or count below the fraction: aborted before rclone; already running: conflict; rclone failure (including `--max-delete` exceeded): failed with its exit status; succeeded, then pruned |
| restore | the eight flow steps, each resumable; store held: refused; snapshot absent, corrupt, or newer: refused; current store newer: accepted; current store unopenable: refused without `--discard-current`; non-empty root: refused without `--into-nonempty`; restored version older: `reconcile-pending` |
| seed rule | the eleven rows of the table above |
| roster import | the eight states of the import table |
| epoch read | address key only, `UserID` key only, both with either later, neither |

### Privileged fields and every channel that may set them

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
fixture state it needs and the reason it would fail, so none passes vacuously.

### 6a: the store foundation (lane 2 worktree)

1. The trigram tokenizer answers a substring query over a populated FTS5 table.
2. A reader holding a transaction open still sees its snapshot after a writer on a second store
   handle commits.
3. Two independent store handles on one file: the second `IMMEDIATE` writer waits while the first
   holds its transaction for half the busy timeout and then commits; held past the timeout, the
   second returns the typed busy error.
4. Every DSN setting (WAL, foreign keys, `FULL`, busy timeout, `journal_size_limit`, `query_only` on
   the read pool) reads back on a freshly opened connection of each class, the maintenance
   connection included.
5. A write through the read pool fails with a read-only error; an orphan insert through the write
   method fails with `SQLITE_CONSTRAINT_FOREIGNKEY` by its extended code.
6. A callback that swallows a failed statement's error does not commit; a nested write call fails
   fast; a write with no deadline is refused.
7. The write method fails closed when `user_version` is changed under it.
8. An injected migration failing after its version write leaves `user_version`, schema, and ledger
   unchanged; a ledger hash mismatch is refused naming both.
9. Migrating the populated job-records fixture forward leaves `pre-migrate-vN.db` with
   `user_version` N and the fixture's rows; a snapshot write that fails leaves the store at N with an
   error naming the path; a second migration never overwrites the lower-version snapshot.
10. An injected foreign-keys-off rebuild migration is followed by an orphan insert through the
    write method that fails.
11. Each invariant fails against its violating fixture for its named catalog fact, including a Go
    `string` bound into a `STRICT` path column, and passes over the skeleton.
12. A zero-byte file and a non-database file are each refused as not a store. A write-mode open of
    an older store refuses naming both versions; a newer store is refused.
13. Serve mode with a second handle holding the lock shared waits and does not migrate; after that
    handle releases, it migrates. A second process that requests the exclusive lock while serve
    mode converts is granted it, and serve mode then holds no handle to the file that process
    renamed, and re-reads the version.
14. The inspect mode against an absent, an older, and a newer store leaves the database file's bytes
    and mode unchanged, with any `-wal` or `-shm` owned by the leaf owner, and reports each state
    by name; `doctor`'s store row renders each in a golden.
15. The `statfs` seam reports NFS (refused), FUSE (warned), and ext4 (opened); a uid mismatch is
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
9. A snapshot taken while another handle commits in a loop passes `integrity_check`, holds every
   row committed before it began, and holds no row from an uncommitted transaction.
10. A successful run against a real `rclone` `local` backend leaves every library file under
    `library/` and the snapshot under `snapshots/`.
11. A file deleted from the library appears under `archive/<stamp>/`, not deleted.
12. A missing root, an empty root with no catalog tracks, and a root whose file count falls below
    the fraction of a populated skeleton's tracks each abort with no rclone invocation recorded by
    the stand-in.
13. With zero catalog tracks and a non-empty remote library, a sync that would archive a file fails
    under `--max-delete 0` and the job ends failed.
14. A stand-in exit 7 makes the job failed with the code in its record.
15. After a successful run, the remote keeps exactly the newest `SNAPSHOT_KEEP` snapshots, including
    the latest, and no archive directory older than `ARCHIVE_RETENTION`.
16. After a successful run, every local pre-migration snapshot older than its snapshot is deleted;
    with the flag off, only the newest `LOCAL_SNAPSHOT_KEEP` survive the next migration.
17. A terminal record older than `DUBPLATE_JOB_RETENTION` is pruned at the next open, and the newest
    successful record of each kind survives.
18. `backup` with an absent store refuses and creates nothing.
19. Setting each retired name fails the load with its message verbatim: for the five R2 names,
    "superseded by DUBPLATE_BACKUP_REMOTE; configure the remote in rclone's RCLONE_CONFIG_<NAME>_*
    environment form"; for `DUBPLATE_BACKUP_STAGING`, "superseded: backup staging now lives in the
    store leaf under DUBPLATE_STATE_DIR".
20. `backup` receiving SIGINT mid-sync ends the job canceled, leaves no stand-in process after
    `WaitDelay`, and exits with the canceled code; each exit code maps to its outcome as the verb's
    help states.
21. `doctor`'s backup rows render never backed up, last success age, unavailable, a missing
    `RCLONE_CONFIG_<NAME>_TYPE`, the crypt warning (ruling 14), and the store's WAL size, free
    space, and local snapshot count and size, each in a golden.

### 6b-ii: the projection and restore (lane 2 worktree)

1. The contract suite holds one case per sentinel and fails when a sentinel has none; every case
   passes in-process and over HTTP.
2. A denying authorizer returns forbidden over HTTP and writes no job record; a nil authorizer
   fails construction.
3. A job started over HTTP is polled and canceled over HTTP and ends canceled with its work stopped.
4. A cancel for a job running in another `ops/local` over the same store succeeds through the
   projection.
5. `restore` refuses while a real reader holds the store.
6. With a live store whose WAL holds uncheckpointed rows, restore checkpoints, swaps one file, and
   the pre-restore copy holds those rows.
7. A crash injected through the step hook after each flow step leaves the intent record; a
   write-mode open refuses naming `--resume`; `--resume` completes.
8. A crash between steps 4 and 5 followed by a write-mode open leaves the staged incoming snapshot
   in place, and `--resume` completes the swap.
9. A newer, corrupt, or absent snapshot is refused; an older remote one is refused without
   `--older`.
10. An older binary's `restore --from-file` of its pre-migration snapshot over a newer current store
    succeeds, and the swapped-in store opens under that binary.
11. A restored older-version snapshot leaves a `reconcile-pending` intent; serve mode migrates it,
    writes the report, and removes the intent, and a write-mode open before then refuses.
12. A tree restore into an empty root yields the backed-up files; into a non-empty root without
    `--into-nonempty` it refuses naming the flag; the reconciliation report names a planted row
    without a file and a planted file without a row.
13. A restored file is owned by the leaf's owner at 0600.

### 6c: the routed schema and the settings home **[CH]** (lane 2 worktree if 5c and 5d leave `internal/config` untouched; otherwise after 5d)

1. The invariant test passes over the full schema, with a populated fixture for every migration.
2. A Go `string` bound into `move_intents`' path columns is refused; an orphan override row is
   refused.
3. A duplicate address, live or tombstoned, is refused under ruling 9's index.
4. `visibility` refuses any value but CH's two.
5. A seed outside its bounds, validated through the registry's value validator with no environment,
   fails the seed rule.
6. The correlated retention rule passes when both keys rise in one seed set and fails when only the
   expiry rises past the stored retention.
7. Each row of the seed table behaves as stated, each by its own test.
8. A registered default changed between two opens reaches a store with no row.
9. `backup` started with a divergent env leaves every settings row unchanged and reports the
   disagreement.
10. A seed writes one audit row with actor `system` and channel `env-seed`; a refused write writes
    none.
11. A write through a second store handle invalidates the first handle's cached value.
12. A key registered without a home still reads from the env; a store-homed key has no wall getter
    (a compile-time fact the plan states and the diff-reviewer checks).
13. `doctor` renders value, source, apply mode, and marker, and an orphan row, in goldens.
14. Each of 6b-i's five constants reads through the accessor, and an admin row for one changes the
    behavior its 6b-i criterion proves.
15. `check` evaluates the seed rule read-only and leaves the database file's bytes unchanged.

### 6d: identity, the handoff, and `serve`'s store (after rung 5's 5d merges)

1. `member import` imports a valid roster once, carries each entry's admin flag, writes the import
   record, and a second run refuses; an imported admin can grant admin through the CLI.
2. Each of the seven import refusal states reports as its table row states, each by its own
   fixture, and `--dry-run` reports all of a file's refusals at once and writes nothing.
3. A member added through the verb before the import merges when the file's entry matches it and
   refuses naming both when it differs.
4. With `DUBPLATE_ROSTER_FILE` set, every verb fails the load naming `member import`.
5. A verified identity's `UserID` is the surrogate id and its `Principal` the streamer handle.
6. A member added through the write path receives a new id; the admin flag arrives only through the
   CLI grant or the import; re-adding a removed address reactivates the same id (ruling 9); a
   concurrent self-demotion of the last two admins leaves one admin.
7. A session revoked by an address-keyed bump before 6d, and one revoked by a bump after 6d, both
   stay revoked; a pre-removal session for a member removed after 6d and re-added under a new id
   (ruling 9's "no" branch, forced in-package) stays revoked; a session revoked under neither key
   stays valid.
8. Two processes bumping one epoch concurrently never leave a stored value lower than either bump.
9. With the env holding A and an in-package admin row holding B, a link minted after the write
   carries B's expiry, a session's next refresh uses B's lifetime, the next request enforces B's
   absolute cap, and a limiter built at the next start uses B.
10. Each re-homed key seeds from the env, and `doctor` shows its source and the `sensitive` marker
    on both lifetimes and the cap.
11. `serve` against an older store migrates before its listener accepts a connection, and its lock
    wait past `DUBPLATE_STORE_MIGRATION_WAIT` exits nonzero naming the holder.
12. `serve` applies the seed rule before the first request, and a request handled after start reads
    the seeded value.
13. `systemd-analyze verify` is clean on 5d's serve unit, whose `ReadWritePaths` names the `store`
    leaf; the state-dir leg covers the leaf.
14. The template carries no `DUBPLATE_ROSTER_FILE` row; `deploy.sh` pre-creates the `store` leaf and
    no longer the roster leaf, runs store-opening verbs through `runuser`, and names the holder
    before restarting `serve` onto an older store whose lock is held.

### Band and ceilings

The ladder drew rung 6 as band M, one pass (BS:656; bands at BS:563-567). The ratified CH routes a
schema and a settings home here **[CH]**, and rung 5 hands rung 6 the surrogate id, the epoch
obligation (R5S:1809-1811), and the roster's move behind its interface (R5S:194-196, PA:209). Five
passes make it band L, which the build spec's ladder entry owes an erratum.

The ceilings use 5b-ii's measured rate, the nearest analog. STATUS records its spend at about 5.6M
of 7.5M after Task 10, about 0.55M per task including fix rounds. Plan preparation adds about 0.85M
per pass (0.83M at 5a, 0.87M at 8a and 8b), the staleness pre-task about 0.2M, and the close
(simplifier, architecture readers, fold) about 0.6M. Two lessons bind: a pass that changes the
config contract overran (1b), and a pass over ten tasks overran (5b-i).

| Pass | Tasks | Arithmetic | Ceiling | Flag (80%) |
| --- | --- | --- | --- | --- |
| 6a | 8 | 0.85 + 0.2 + 8 × 0.55 + 0.6 = 6.05 | 6.1M | 4.9M |
| 6b-i | 9 | 0.85 + 0.2 + 9 × 0.55 + 0.6 = 6.6 | 6.6M | 5.3M |
| 6b-ii | 5 | 0.85 + 0.2 + 5 × 0.55 + 0.6 + 0.3 targeted review = 4.7 | 4.7M | 3.8M |
| 6c | 7 | 0.85 + 0.2 + 7 × 0.55 + 0.6 = 5.5, plus 10% for a config-contract change | 6.1M | 4.9M |
| 6d | 9 | 0.85 + 0.2 + 9 × 0.55 + 0.6, plus 0.7 for security review inside the chain = 7.3 | 7.3M | 5.8M |

The total is 30.8M. The phase ceiling and each lane's sub-ceiling are set at lane launch (PA
section 4). Each ceiling is re-derived from the previous pass's actual before the next plan is
drafted, and a plan that exceeds ten tasks splits the pass, names the follow-on, and re-derives
both ceilings.

### Contested surfaces

The conductor writes these rows into PA's ownership table (PA:173) before lane 2 launches.

| Surface | Owner | Consumers |
| --- | --- | --- |
| `internal/engine/store` and its migration set (new) | rung 6's 6a owns the runner, ledger, and invariant test; every later rung appends migrations with fixtures | a rebase onto another lane's migration renumbers the rebasing lane's files, and merge ritual step 5 runs |
| `internal/ops`, `internal/ops/local`, `internal/backup` (new) | rung 6's 6b-i | rungs 7, 9, 13, 14 add operations; rung 12 adds the streamer database to the backup set; rung 16 schedules backup |
| `internal/ops/opshttp` (new) | rung 6's 6b-ii, unexported | rung 14 exports and mounts it |
| `internal/settings` (new) **[CH]** | rung 6's 6c | every rung that declares a store-homed key; rung 14 |
| `internal/config/registry.go`, `thresholds.go` **[CH]** | rung 1b's contract; 6c adds the four metadata items and the correlated-rule list additively | rung 5's lane does not edit these files while 6c runs; if it must, 6c waits |
| `internal/config/keys.go` | 6b-i removes the "backup targets" reserved class in the commit that declares its replacement, registering the retired names through `AddSuperseded`; 6d supersedes `DUBPLATE_ROSTER_FILE` | every later rung, by the same rule |
| `cmd/dubplate-server/main.go` (`newProductionRegistry`, `AddCommand`) | 6a registers the busy timeout and the migration wait; 6b-i the class, its constants, and `backup`; 6b-ii `restore`; 6d `member import`; rung 5's 5c adds the `member` group in the same block | a textual conflict keeps both lines in landing order, and merge ritual step 5 runs |
| `cmd/dubplate-server/doctor.go`, `check.go` | 6a the store row and the inspect report; 6b-i the backup rows; 6c the settings rendering and read-only seed evaluation | rungs 5, 9, 13 add rows the same way |
| `cmd/dubplate-server/serve.go` | rung 5 through 5d; 6d then adds the store open, recovery, and the seed rule before the listener | rungs 7, 12, 14 |
| `go.mod`, `go.sum` | every lane adds its own requirements | a rebase conflict is resolved by `go mod tidy`, and merge ritual step 5 runs |
| The state directory layout | rung 4; 6a adds the `store` leaf | rung 16 converges |
| `internal/auth` roster and epoch, `cmd/dubplate-server/auth.go`, `internal/identity` | rung 5; 6d after 5d merges, with no concurrent auth-writing lane | rungs 7, 12, 14 |
| 5d's greenfield serve unit (a distinct file from the frozen `systemd/dubplate-serve.service`, R5S:1708-1713), `scripts/deploy.sh`, the state-dir leg in `scripts/check.sh`, `env/dubplate-server.env.template` | 6d after 5d merges; `deploy.sh`'s store-opening verbs run through `runuser -u <owner>` | rung 16 converges |

---

## Process and review

Each pass opens with a staleness pre-task, then runs the implementer, `diff-reviewer`, and gate
chain per task in segments of three to four tasks on green commits. Rung 6 is not one of the three
rungs that keep the full adversarial cycle, so review scales to risk. The data-loss paths earn one
targeted fresh-context read at 6b-ii's close, over the migration runner, the lock procedure, the
snapshot, and restore, by an `opus` reviewer briefed with the risk model. Pass 6d runs
`web-auth-security-reviewer` inside its chain. At each merge the conductor dispatches one
`go-architecture-reader` per touched package, and every `code-simplifier` dispatch carries
`.claude/instructions/simplifier-brief.md` verbatim.

Passes 6a through 6c are disjoint from rung 5's territory except the rows above, so they run as
lane 2 beside 5c and 5d, and none of them makes `serve` open the store. Pass 6d waits for 5d. Rung 6
edits 5d's serve unit, `deploy.sh`, and the template in the repo at 6d; the deploy rides the next
ruled deploy, and the no-deploy carry in `docs/STATUS.md` holds until 5d lifts it.

---

## Errata this spec owes

Each lands, dated 2026-09-23 or at ratification, in the named document when Geoff rules this spec.

1. **BS ladder entry 6** (BS:656): band M to L, five passes, for the reasons in "Band and ceilings."
2. **BS "Box safety during the build"** (BS:934-935): rungs 5 and 6 edit deploy units and scripts
   in the repo, and their deploy rides the next ruled deploy; the box sees them only then.
3. **BS config wall** (BS:218-220): the backup class is one remote plus a flag, not a pair, and its
   credentials are rclone's own env names (conditional on ruling 5).
4. **CH wall homing** (CH:74-75): keys read by a unit not granted the store leaf (the shadow pair)
   stay on the wall (conditional on ruling 3).
5. **CH wall list** (CH:52-54): "the roster file" retires at 6d; members are store state and the
   roster file is an import source only (conditional on ruling 1).
6. **CH integration classes** (CH:56): the backup schedule key is declared by rung 16 with its
   timer, not by rung 6.
7. **CH routed tables** (CH:638): apply mode and the sensitive flag are registry metadata, and the
   settings table carries per-row state (clarifying, read with BS:39-40).
8. **R5S roster key** (R5S:1552-1555, R5S:198-200): `DUBPLATE_ROSTER_FILE` stops being the roster's
   key at 6d, and the roster leaf's pre-create retires (conditional on ruling 1); PA:218's leaf row
   follows.
9. **R5S surrogate id** (R5S:1153-1158) and the `identity.go` field comment: the surrogate id is
   `Identity.UserID`, and `Principal` stays the streamer handle.
10. **R5S epoch** (R4 at R5S:336-338, R18 at R5S:538-539, R35 at R5S:714, R5S:1687-1690, and
    R5S:1809-1811): the re-key is a permanent dual-key read with bumps to both keys, not a rename;
    R35's notification keeps its address argument and gains the `UserID`.
11. **S20 hooks 1, 2, 3, 11** (S20:363-372, S20:392): schema lands with the computing rung's
    migration (conditional on ruling 2).
12. **S20 prior art** (S20:435 and the fallback row): ncruces is a wasm2go translation, not wazero,
    and goose's decline rests on its silent-ahead `Up`, not on "ten database drivers."
13. **BS U3** (BS:662-664): a track path is relative to its collection's library-relative root
    (clarifying).
14. **BS ladder entries 12 and 16** (BS:844-851 for 16): rung 12 adds the streamer's database to the
    backup set (ruling 4); rung 16 declares the backup schedule key with its timer and renders the
    `RCLONE_CONFIG_*` template rows (ruling 5).

---

## What depends on the configuration homes spec

Each passage marked **[CH]** exists because of CH's routes: rulings 1, 2, and 13; the routed tables
landed at 6c, and the root-allocation paragraph; the settings home, the registry metadata, and the
seed table; pass 6c in full and 6d's re-homing half; the hook table's rows for hooks 8 and 15; the
contested rows for `internal/settings` and the registry files; and job retention as a store-homed
key. The foreign-key DSN rule does not depend on CH: CH requires it (CH:524-526), and hook 9 and the
schema's own references require it independently.

---

## Not verified in this revision

- `zombiezen.com/go/sqlite`'s license and importer count; its verdict rests on neither.
- The exact wording of `database/sql`, `os/exec`, rclone's sync, `--backup-dir`, and exit codes, and
  `proc(5)`, which the planning-miss table sends each plan to quote.
- Whether rclone counts a file moved into `--backup-dir` against `--max-delete`; 6b-i criterion 13
  settles it against the real binary, and the plan quotes the answer.
- 5d's serve unit's file name and `Type=`, which 5d has not landed; 6d's staleness pre-task reads
  both.
- The line citations into rung 5's files, which 5c and 5d may move.
