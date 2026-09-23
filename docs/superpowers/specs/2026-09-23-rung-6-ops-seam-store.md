# Rung 6: the ops seam and the store

Status: **revision 2, draft for the owner's ratification**, 2026-09-23. Revision 1 (`12a6941`) drew
four fresh-context adversarial reviews (`e97d761`, 113 findings); this revision folds all of them,
and `research/2026-09-23-rung-6-spec-fold-record.md` gives every finding's disposition and the long
arguments this spec only states. No owner ruling below is given yet. Tree citations were re-read at
`e97d761`; the staleness pre-task each pass opens with re-checks the ones that pass leans on.

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
and then a store snapshot, sent to any rclone remote behind an integration flag, and an offline
restore that swaps one file under a restore-intent record and reconciles the catalog against the
tree before `serve` may open. Rung 6 also lands the schema the ratified configuration homes spec
routes to it, the settings home with an env-seed rule that only `serve` applies, and, once rung 5
closes, the surrogate user id and an explicit roster import. The ladder drew it as band M; it is
five passes (band L), and "Band and ceilings" gives the arithmetic.

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
an erratum line listed in "Errata this spec owes"; none is silent.

---

## Owner rulings

Fourteen yes-or-no questions change what an adopter sees, configures, or is asked for, ordered by
how much build each answer changes. Rulings 1 through 7 are the draft's, restated where the
reviews showed them unbuildable; 8 through 14 are forks the reviews raised. Everything else is a
technical decision closed in its own section.

**1. Does the roster move into the store through an explicit one-shot import verb, after which
`DUBPLATE_ROSTER_FILE` is a superseded name?** Recommendation: yes. **[CH]** Pass 6d adds
`dubplate-server member import --roster PATH`. It runs under the store lock held exclusive, in
one transaction, refuses a users table that is not empty or a store that already holds an import
record, mints every member's surrogate id, and writes an import record (the file's SHA-256, the
member count, the time, and the operating-system actor) in the same transaction. The same commit
removes the key from the magic-link source's `Required` list (`cmd/dubplate-server/auth.go:44-46`)
and registers it through `AddSuperseded` with a message naming the verb, so a set key is an
ordinary load error that needs no store state. Reason: the draft's import-on-open needed the key
set for one boot and a load error after it, which `config.Load` cannot express because it never
sees the store, and which crash-loops `serve` (`systemd/dubplate-serve.service:45-55`). CH makes
members store state (CH:67-70), and one transaction must own a membership and its id
(R5S:1153-1158).
What it changes for an adopter: the upgrade to 6d is one deploy, whose template no longer carries
the key, plus one command; afterward, members are added by the web page or the `member` verbs.

**2. Do CH's routed tables land at 6c while S20's hook columns 1, 2, 3, and 11 arrive with the rungs
that compute them?** Recommendation: yes. **[CH]** The rule applied is "a table or column lands
with its first writer," and CH's routed tables are the owner's explicit exception, because CH
routes them to rung 6 by name (CH:637-645) and its visibility ruling puts that column on the row at
rung 6 (CH:195). The catalog skeleton (collections, albums, tracks, users) lands with them because
the routed tables hold foreign keys into it. The audio key and file hash (hook 1, rung 7), the
folded sort key (hook 2, rung 9), the art hash (hook 3, rung 9), and the gains (hook 11, rung 9)
arrive by those rungs' migrations under the invariants below. Reason: S20 assigns those four
columns to "rung 6 schema" (S20:363-372, S20:392), and CLAUDE.md's 2026-09-12 refinement says a
hook is built by its first consumer; the refinement is the later ruling, so a yes owes S20 a dated
erratum on the four hooks. A no lands the four as nullable columns at 6c. What it changes: nothing
an adopter sees.

**3. Do the ten auth thresholds move to the store with env seeds, while the serve keys, the shadow
keys, and the store's busy timeout stay on the wall?** Recommendation: yes. **[CH]** CH moves every
tunable with a default into the store (CH:66) and keeps on the wall what the process needs before
it can start (CH:74-75). The serve prefix, bind addresses, port, and the busy timeout are read
before the store opens. The shadow floor and budget are read by the shadow unit, which is not
granted the store leaf and retires at launch. Reason: moving the shadow keys buys a store grant
for a unit that dies at launch; the category is new to CH, so a yes owes CH an erratum line. What
it changes: after 6d, `doctor` shows each auth threshold with its source.

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
(rclone's `remote:path` form). The credentials are env names rclone itself reads, rendered from the
age store like every other secret, so CH's "the wall holds every secret" (CH:52-57) stays true
with no erratum and no hand-edited `rclone.conf`. The five R2 names and `DUBPLATE_BACKUP_STAGING`
(`internal/config/keys.go:60-71`) become supersessions. Reason: the R2 names fail BS's fourth
universality test (BS:192-194); a credential in an rclone config file has no route through the age
store, which fails the stranger-setup goal. The alternatives are a wall key naming an rclone config
file that `deploy.sh` renders, or the draft's position with a CH erratum narrowing "every secret."
What it changes: an adopter configures any rclone backend the way rclone documents its env form.

**6. Is restore an offline verb?** Recommendation: yes. `dubplate-server restore` refuses while
any process holds the store, and it is not an operation of the ops interface, so no network channel
reaches it. Reason: replacing a database file under an open connection "results in behavior that is
undefined and probably undesirable" (sqlite.org/howtocorrupt.html section 2.5). What it changes: a
restore means stopping the server first.

**7. Does the store keep one pre-migration snapshot per from-version, deleted only after a
successful backup postdates it?** Recommendation: yes. Each is named by the version it was taken
at (`pre-migrate-v4.db`), a higher version never overwrites a lower one, and an absent snapshot
refuses the migration. Reason: the draft's single snapshot was overwritten by the next upgrade
exactly when a clean-committing migration was semantically wrong. What it changes: the state
directory holds one store copy per upgrade not yet backed up.

**8. Does the roster import refuse a member with no `streamer_username`, naming each one, rather
than deriving a handle?** Recommendation: yes. S20 hook 5 makes the handle required, unique, and
stable (S20:376-378), and the roster makes it optional today (`internal/auth/roster.go:109`).
`check` evaluates the import read-only before the deploy and names every member that would fail.
Reason: the handle names the member's existing streamer account; a derived one creates a fresh
account that orphans the member's history. What it changes: an adopter fills the field first.

**9. Does re-adding a removed member reactivate the same surrogate id?** Recommendation: yes. The
address index is unique across live and tombstoned rows, and re-adding the address clears the
tombstone on the same row. Reason: stars and history key on the id, the address is the login
proof, and the old epoch record keeps a pre-removal cookie revoked. A no mints a new id and keeps
the address unique across live rows only; the dual-key epoch read in "Identity in the store" keeps
that safe too. What it changes: a returning member gets their library state back.

**10. Does dubplate prune dated snapshots and the remote archive itself, by a store-homed
retention?** Recommendation: yes. Reason: a `local`, `sftp`, or `webdav` remote has no lifecycle policy, so
leaving pruning to one fails BS's universality bar. Pruning never removes the latest snapshot, and
the archive default exceeds the backup-failure detection window.

**11. Is the latest snapshot the only supported restore of store and tree together?**
Recommendation: yes. The remote tree is only ever the latest, with superseded files under the
archive; an older snapshot pairs with a newer tree. An older snapshot restores the store alone under
an explicit flag, and the reconciliation report names every row without a file. Reason:
point-in-time recovery from the archive is a feature no ruling asks for.

**12. Does a disaster restore rotate the session and magic-link secrets rather than back up the auth
leaf?** Recommendation: yes. The restore verb's final report and the runbook require the rotation.
Reason: epoch and nonce records stay files in the auth leaf (R5S:533-539) and are not in the
backup, so a restored box would forget every revocation; rotation invalidates every outstanding
credential, including revocations made after the last backup, which a restored leaf would miss.
What it changes: every member logs in again after a disaster restore.

**13. When an env-seeded key is removed from the env, does its row revert to the default, logged?**
Recommendation: yes. **[CH]** Reason: an operator who deletes a key expects the default, and a row
that keeps the old seed is the silent outcome CH forbids (CH:85-86). The alternative fails the
load.

**14. Does `doctor` warn when the backup remote is not an rclone `crypt` remote?** Recommendation:
yes, a warning and never a refusal. Reason: after 6d the snapshot holds member addresses; a refusal
fails adopters whose provider encrypts at rest.

**Not asked: rung 9's concurrency model** (BS:760-764). A lease or a queue sits above the store's
serialized writers. The store's two constraints on rung 9: no write transaction spans a MusicBrainz
lookup or a file copy, and library mutations continue during a backup's sync unless rung 9's lease
pauses them, so restore's reconciliation report is the net.

---

## What rung 6 builds

| Deliverable | Pass | Proved by |
| --- | --- | --- |
| `internal/engine/store`: open (inspect and writing), the leaf, the lock, three connection classes, the write function, migrations, snapshots | 6a | 6a criteria 1-14 |
| The migration set, its applied-migration ledger, and the schema-invariant test | 6a, then 6c | 6a 8-12, 6c 1-2 |
| The job records table and `doctor`'s store row and `check`'s inspect report, the store's first production callers | 6a | 6a 13-14 |
| `internal/ops`: interface, wire types, error codes, privileges, the catalog, the job model | 6b-i | 6b-i 1-8 |
| `internal/ops/local`: the runner, leases, recovery, cross-process cancel | 6b-i | 6b-i 3-8 |
| `internal/backup`: the snapshot job, the rclone transport, the guards, the remote layout | 6b-i | 6b-i 9-14 |
| The backup integration class and its supersessions; the `backup` verb; `doctor`'s backup row | 6b-i | 6b-i 15-18 |
| `internal/ops/opshttp` (unexported) with its authorizer, and the contract suite | 6b-ii | 6b-ii 1-4 |
| The `restore` verb: the intent record, the one-file swap, library restore, reconciliation | 6b-ii | 6b-ii 5-11 |
| The catalog skeleton and every table CH routes to rung 6 **[CH]** | 6c | 6c 1-4 |
| Registry metadata: home, apply mode, `sensitive`, bounds, correlated rules **[CH]** | 6c | 6c 5-6 |
| `internal/settings`: accessor, generation, seed rule, audit; job retention re-homed **[CH]** | 6c | 6c 7-15 |
| The import verb, the users-backed roster, `Identity.UserID`, the dual-key epoch read | 6d | 6d 1-8 |
| The auth thresholds re-homed with accessor reads at every call site | 6d | 6d 9-11 |
| The store leaf's deployment: the serve unit's grant, `deploy.sh`, the state-dir leg, the template | 6d | 6d 12-13 |

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
  check land with their first caller at rung 7 or rung 9.
- No hook column whose value a later rung computes (ruling 2), and no `collections` row version for
  CH's `if_version` check (CH:256-257): rung 14's shape edit is its first writer.
- No default collection row. The first rung that writes a collection (rung 7's uploader or rung 10's
  migrator, whichever lands first) creates it, owned by the first admin.
- No backup schedule and no schedule key. CH's class lists "backup targets and schedule" (CH:56);
  the schedule's consumer is a systemd timer, so rung 16 declares the key with the timer, and rung
  16's carried list gains both. rclone's credential names reach the template at rung 16's
  convergence; 6b-i declares the class and documents the names.
- No backup pings: rung 9's monitoring seam (`internal/config/keys.go:93-100`) reads the job's
  terminal state. No runbook: rung 16 authors it (CH:661-662), and each pass records the lines it
  owes (restore size, rollback cost, rotation, crypt, roster file deletion) in its close.
- No re-authentication gate or sensitive-change notice (rung 14, CH:657-660), and no URL-valued
  store-homed key; the write path's validator is where rung 14's range check attaches (CH:132-133).
- No record from rung 4's or rung 5's leaves moves into the store except the roster (ruling 1);
  epoch and nonce records stay files (R5S:533-539). No last-admin guard of its own: 5c builds it,
  and 6d re-establishes it inside the write transaction.
- No reconciliation of `DUBPLATE_ADMIN_EMAIL` against `users.admin`. After 6d, `users.admin` is
  authoritative for authorization; the contact row's reconciliation is rung 14's (BS:833).

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
| `internal/backup` | snapshot job, rclone transport, guards, restore staging, reconciliation | `store`, `atomicfile` |
| `internal/settings` **[CH]** | accessor, seed rule, audit | `config`, `store` |

`internal/ops` imports no server package (RE:116), the engine has no HTTP knowledge (BS:375), and
`internal/config` never imports `store` or `settings`.

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
Leftover staging files are removed at every writing open.

Every verb that opens the store refuses when its effective uid differs from the leaf's owner,
naming the owner and the fix (`runuser -u <owner> -- dubplate-server ...`); a root-run verb would
otherwise leave a root-owned `-shm` or database that `serve` (`User=music`) cannot open.

The lock is an advisory `flock` on `store.lock`, released when the process exits. `serve`,
`backup`, and the `member` verbs hold it shared; `serve` holds it exclusive while it migrates;
`restore` and `member import` hold it exclusive; `check` and `doctor` take it shared and
non-blocking. `flock` cannot name a holder, so a refusal says "held by another process" and lists
the pids `/proc/locks` shows for the lock file's inode when that file is readable.

WAL needs every process on one host (sqlite.org/wal.html section 1). The store probes the leaf's
filesystem with `statfs`, refuses NFS, CIFS, SMB2, and 9p, and warns on FUSE, because sshfs and
rclone mounts are remote while mergerfs is local and common on home servers. The probe is a
package variable the test substitutes, which is the one test seam rung 6 declares, justified
because no unit test can mount NFS.

### Two opens

The **inspect open** is read-only (`mode=ro`), takes the lock shared and non-blocking, never
creates, and never migrates. It reports absent, restore in progress, older (naming the pending
migrations), newer (naming both versions), or current. A read-only WAL open may create `-shm`
(sqlite.org/wal.html section 5), which the uid check keeps owned correctly. `check` and `doctor`
use it.

The **writing open** creates an absent store only in `serve` and the `member` verbs; `backup`
refuses an absent store, because a mistyped state directory would otherwise upload an empty dated
snapshot that "latest" then selects. A writing open that finds a `restore-intent` record refuses
and names `restore --resume`. A writing open other than `serve` refuses a store whose version is
not the binary's, naming both versions and "start serve to migrate."

### Connections and transactions

2.0 hook 9 asks for one writer, WAL with a `busy_timeout`, and a seam a later projector joins so
its delta commits with its event (S20:387-389). The store opens three connection classes over one
file:

- The **write pool** holds one connection with `_txlock=immediate`, so every write transaction
  takes the write lock first and the busy timeout governs the wait.
- The **read pool** opens with `_query_only=1`, so no caller writes outside the write function.
  A write there would fail instantly, because the busy handler does not run on a deferred
  read-to-write upgrade (verified; see "Behaviors a task relies on").
- The **maintenance connection** is a separate `*sql.DB` opened for one job, a snapshot or a
  foreign-keys-off table rebuild, and closed afterward, so it never returns to a pool.
  `VACUUM INTO` fails on a `query_only` connection and inside a transaction, and running it on the
  write pool would stall every in-process writer for the snapshot's length.

Every per-connection setting rides the DSN, because `database/sql` opens connections at will, as
modernc's shorthand keys, which the driver validates before applying any: `_busy_timeout`, `_foreign_keys=1`, `_journal_mode=WAL`, `_synchronous=FULL`, and
`_query_only` or `_txlock` per class (`sqlite.go:307-391` at v1.59.0). `journal_size_limit` has no
shorthand and rides `_pragma`, which the driver executes verbatim and cannot pre-validate
(`sqlite.go:305-306`), so 6a proves every setting reads back on a fresh pooled connection.

`synchronous` is `FULL`. Under `NORMAL`, "a transaction committed in WAL mode ... might roll back
following a power loss" (sqlite.org/pragma.html), which breaks every pattern where file work
follows a commit: CH's move intent, the import record, a job record. `FULL` also meets `VACUUM
INTO`'s sync condition (sqlite.org/lang_vacuum.html) and is this build's compiled WAL default.
`journal_size_limit` bounds the WAL's size after a checkpoint, and every snapshot ends with
`wal_checkpoint(TRUNCATE)`, because a checkpoint "does not normally truncate the WAL file"
(sqlite.org/wal.html).

The store exposes one way to write: a function that runs a callback inside one `IMMEDIATE`
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

The busy timeout is a wall threshold with a researched default (ruling 3), declared in rung 6's own
file.

### Migrations

2.0 hook 7 asks that the mechanism be chosen with an evolving schema in view (S20:383-384). The
runner is hand-rolled over `embed.FS` and `PRAGMA user_version`, as S20's prior-art row records
(S20:435); "Prior art" re-scores the alternatives. Its contract:

- Only `serve` migrates, at start, before its listener opens. It takes the lock exclusive; when a
  shared holder exists it waits and logs the holder at an interval. The serve unit sets no
  `TimeoutStartSec`, so the 6a plan quotes systemd's default and bounds the wait below it. After migrating it downgrades to shared.
  A binary deployed while an older `serve` runs therefore never migrates under it.
- Each numbered SQL file applies in its own `IMMEDIATE` transaction that reads `user_version`,
  applies the file, formats the new version into `PRAGMA user_version = N` (a pragma takes no bound
  parameter), and records the file's name and SHA-256 in an applied-migration ledger table. A
  failed migration leaves the prior version, schema, and ledger.
- A ledger row whose name or hash differs from the binary's file at that version is refused, naming
  both; a store once opened by an unmerged worktree build is caught here.
- A store whose `user_version` exceeds the binary's highest migration is refused, naming both. This
  is CH's "laptop ahead" direction (CH:165-168) and covers a newer snapshot restored under an older
  binary.
- A store is non-empty when `user_version > 0`. Before migrating a non-empty store the runner
  checks free space for one full copy plus margin, writes `pre-migrate-v<N>.db` through the
  maintenance connection into `staging/`, verifies it, renames it into `snapshots/`, and fsyncs the
  directory by hand (ruling 7). Any failure refuses the migration and names the snapshot path.
- A rebuild migration that needs foreign keys off runs on the maintenance connection, because
  enforcement "may only be enabled or disabled when there is no pending BEGIN or SAVEPOINT"
  (sqlite.org/pragma.html). It runs `PRAGMA foreign_key_check` before commit, refuses to commit on
  any row, and closes the connection.
- Every migration file ships with a fixture that populates the schema as of its version. The
  invariant test fails on a migration with no fixture. 6a's proof is migration 1 over an empty
  store; 6c's is the first populated one.
- Files are append-only once merged. A lane that rebases onto another lane's migration renumbers
  its own, and the rebase is non-fast-forward, so merge ritual step 5 runs.

Rolling back to an older binary meets the newer-store refusal; the way back is that binary's own
`restore` of the matching pre-migration snapshot, which loses every write since the upgrade. The
runbook states the cost.

### Schema invariants

One test applies every migration to an empty database and to each migration's populated fixture,
then inspects the result through the catalog (`pragma_table_list`, `pragma_table_info`,
`pragma_foreign_key_list`, `pragma_index_list`), never by parsing SQL. Each rule below is stated as
the catalog fact the test checks, and one violating fixture per rule proves the rule can fail for
its named reason:

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
6. Every table's primary key is a `BLOB` column named `id`; no column named `id` or ending in `_id`
   is an `INTEGER PRIMARY KEY`. Alternate keys (MBIDs, the audio key) are exempt (S20:166-167).
   A store test proves the id function mints 16 random bytes, never equal to a rowid.
7. No foreign key to `users` or `collections` declares `ON DELETE CASCADE`, since CH requires
   tombstones and audit history to survive (CH:519-521).

What a column means stays a reviewer's question: a path column not named as one, or a source
identifier under another name, passes the test. The 6c plan and every later rung's migration task
carry that as a `diff-reviewer` checklist line, not a claimed proof.

### The schema rung 6 lands

Pass 6a lands the job records and the applied-migration ledger. Pass 6c lands the rest **[CH]**.
Table names are working names the 6c plan may change; the contents are the contract.

- `users`: the surrogate id (hook 5, S20:376-378); the normalized address, uniquely indexed across
  live and tombstoned rows under ruling 9; the display name; `streamer_username`, required and
  unique under ruling 8; the admin flag; a tombstone time (CH:519-521); timestamps. Created empty;
  6d fills it.
- `collections`: name; a server-allocated `root_path`, unique and never client-supplied; `owner`,
  a non-null reference to users; `visibility`, with a CHECK admitting only 1.0's single value as CH
  names it; `writer` with the S20 hook 8 domain, this server or a named local binary (S20:385-386);
  the writer epoch; the mirror generation; the last-mirror time; the duplicate-collapse flag; the
  layout template; the tagging policy; and the admission overrides (CH:173-181, CH:637-642).
  Share rows are reserved in this spec only: no table, no column, no check (CH:535-538).
- `albums` and `tracks`: opaque ids, the collection reference, the track's `rel_path` relative to
  its collection's `root_path` under U3, the MusicBrainz release and recording ids under invariant
  4, the album-and-track relation, and timestamps.
- `settings` and `settings_audit`: one row per store-homed key with its value, source, seed value,
  actor, and time (CH:79-87), and one audit row per write with `at`, `kind`, actor, channel, key,
  and old and new values (CH:130-132). Apply mode and `sensitive` are registry metadata, not
  columns, which reads CH:638 together with BS's registry erratum (BS:39-40).
- `member_overrides` and `member_preferences`: per-member rows keyed to users (CH:122-125, CH:161).
- `device_credentials`: the credential hash, the user, the device id, creation, last-seen address,
  expiry, and revocation (CH:543-548).
- `quota_counters`: the per-member counter and the group total (CH:488-500).
- `move_intents`: `at`, `kind`, the old and new paths as `BLOB`, a granularity (file or directory),
  and a state, with the replay rule "source absent and target present means done" (CH:451-458).
  Rung 16's relocation is a named consumer.

Root allocation supports rung 16's relocation (BS:131-141, CH:662-664): every collection's root is
allocated from its id, with no column for a grandfathered root. If rung 12's Navidrome check fails
and Geoff rules for grandfathered roots through rung 10's plan, that ruling adds its column then.
Track paths relative to their collection's root make a relocation one row update and one subtree
move rather than a rewrite of every track row.

### Paths (U3)

Every path is relative and byte-preserving from the first schema (BS:662-664, EVO:356-366): a
collection's `root_path` is library-relative and a track's `rel_path` is relative to that root.
A path crosses the store boundary as `[]byte`, and invariant 1's `STRICT` rule enforces it in
SQLite, not only in Go. Paths compare bytewise. The store refuses a path that is absolute, has an
empty, `.`, or `..` component, or carries NUL. The build compiles with
`SQLITE_LIKE_DOESNT_MATCH_BLOBS`, so `LIKE` never matches a path; rung 9's query verb says how it
searches paths. Rung 7's gate rules on non-UTF-8 filenames and canonicalizes symlinks once
(EVO:363-366, EVO:600-602); whatever it admits, the store returns unchanged.

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
one) also holds a kind lock for the job's life, so a second start returns a typed conflict naming
the running job. Recovery runs at every writing open and before any exclusive start:

| Record | Owner lease | Outcome |
| --- | --- | --- |
| queued or running | held | left alone, whatever its kind |
| queued or running | free | marked interrupted |
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
read them. Terminal records are pruned by retention (6b-i: a fixed constant; 6c: a store-homed key
**[CH]**); pruning never removes the newest successful record of a kind, which `doctor`'s
last-backup age reads.

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

A backup job syncs the library tree, then writes a store snapshot, then sends the snapshot. The
snapshot comes second so every file the catalog named when the sync began is on the remote or in
its archive. The pair is still not atomic: a change committed during the sync can leave a row
whose file is not on the remote, which restore's reconciliation report names. The snapshot records
the run id and the sync's completion time in its object metadata.

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
lexicographically. Pruning follows ruling 10.

The guards run before rclone. A missing library root aborts. So does a tree whose regular-file
count falls below a store-homed fraction of the catalog's track count, and when the catalog holds no
tracks, an empty root aborts. rclone runs with `--max-delete` sized from the same count. The salvage
guard checked only the top level (`_salvage/internal/backup/backup.go:22-34`, `:125-128`); a wiped
mount under surviving collection directories passes it, so the fresh review reverses that argument.
Every rclone invocation runs under `exec.CommandContext`; the 6b-i plan quotes `os/exec` on what
cancellation sends and how `WaitDelay` bounds the wait, and rclone's documentation on sync,
`--backup-dir`, `--max-delete`, and exit codes.

The gate runs a real `rclone` against a `local` backend in a temporary directory, and a missing
`rclone` fails the test, never skips it. A recording stand-in whose exit code each case sets covers
the remote-failure and SIGINT cases; it is never an always-succeeds fake.

Backup does not preserve hardlinks, so a restored library is larger than its source (CH:410-414).
Rung 16's relocation plan states how the next backup handles a moved subtree, by rclone's rename
tracking or a planned re-upload.

### The restore verb

`restore` is offline (ruling 6). Its interface: `--snapshot <stamp>` or `--snapshot latest` names
an object under `snapshots/`, or `--from-file PATH` names a local file such as a pre-migration
snapshot; `--library` also restores the tree; `--into-nonempty` permits a non-empty library root;
`--older` permits a snapshot other than the latest, store only and never with `--library`
(ruling 11); `--resume` continues a recorded
restore. Its flow:

1. Take the lock exclusive, check the uid, and check free space.
2. Fetch or read the snapshot into `staging/` at 0600, run `integrity_check`, and refuse one whose
   `user_version` exceeds the binary's.
3. If a current store exists, open it, run `wal_checkpoint(TRUNCATE)` and `journal_mode=DELETE`,
   close it, and verify no `-wal` or `-shm` remains, so the swap moves one file. A current store
   that cannot open is refused unless the operator names `--discard-current`, which moves the
   unopened files aside together and names them.
4. Write the `restore-intent` record through `atomicfile.WriteStream`, naming the pre-restore and
   incoming names and whether the library half is owed.
5. Rename the current database to `snapshots/pre-restore-<stamp>.db`, rename the snapshot into
   `db/`, and fsync both directories. The next writing open restores WAL mode through its DSN.
6. With `--library`, copy `library/` from the remote into the library root, refusing a non-empty
   root without `--into-nonempty`.
7. Run the reconciliation report: rows without files and files without rows, written to a file in
   the leaf and summarized on stdout.
8. Remove the intent record, and print the ruling 12 rotation instruction.

A crash at any step leaves the intent record, and every writing open refuses until `restore
--resume` finishes the recorded steps. A pre-restore copy follows ruling 7's retention. Restoring a
pre-6d snapshot under a 6d binary leaves the users table empty, and the report says to run `member
import`.

---

## The settings home **[CH]**

### The accessor and the registry metadata

Every store-homed key reads through the settings accessor, never from a long-lived struct
(CH:88-92). Invalidation crosses processes: the store holds a settings generation integer that
every settings write bumps inside its transaction, and the accessor reads it, one indexed read,
before trusting its cache. `PRAGMA data_version` is per connection and cannot serve.

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
any consumer reads a setting. It applies every seed in one `IMMEDIATE` transaction and validates
each key's bounds and every correlated rule against the post-seed set, so raising the link expiry
and the nonce retention together in one deploy succeeds. Every other verb, `check` included,
evaluates the rule read-only and reports a disagreement without failing, so a timer-run backup never
fails over an auth threshold it does not read. A seed or re-seed writes an audit row with actor
`system` and channel `env-seed`. A `default` source stores no value; reads resolve the registry's
current default, so a re-researched default reaches every existing store. The closed table:

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

Pass 6c's own tunables are the first store-homed consumers: job retention and the backup guard's
fraction, each with a researched default recorded where it is defined (BS:236-248).

### The seed handoff

Fifteen thresholds are registered through the env today: rung 4's two shadow keys, pass 5a's three
serve keys, and rung 5's ten auth keys (`cmd/dubplate-server/auth.go:99-200`). All ten auth values
resolve once into `authConfig` (`auth.go:402-416`) and reach the auth core as construction
arguments, which is the long-lived struct CH forbids. So the apply-mode column is a **required
behavior** 6d builds, not a description of today's code: 6d replaces `authConfig` with accessor
reads at each call site 5c and 5d wire.

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
| `DUBPLATE_SHADOW_FLOOR`, `_BUDGET` | `shadow.go:73-88` | wall | restart | unchanged; the shadow unit is not granted the store leaf |

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

**The import** follows ruling 1. It evaluates read-only first, and `check` runs the same evaluation
before a deploy. Its states: the file absent, unreadable, or malformed refuses naming the file; a
duplicate address refuses naming both entries; a member with no `streamer_username` refuses naming
each one (ruling 8); a populated users table or an existing import record refuses and says so; a
valid file imports. 6d removes `deploy.sh`'s pre-create of the roster leaf (R5S:198-200, PA:218)
and leaves the file in place; the runbook tells the operator to delete it after verifying the
import. Rolling back past 6d needs the old template and the old file, and members added since then
are re-added through the verb afterward.

**The resolution point.** R5S ruled that the store moves behind the roster's interface without
touching a call site (R5S:194-196, R5S:287-292, PA:209), and `internal/auth/roster.go:113-116`
already frames the substitution as one constructor. The resolution point (`Lookup`,
`roster.go:286`) reads the users table. `identity.Identity` gains a `UserID` field carrying the
surrogate id, which per-user rows key on. `Principal` stays the streamer handle: rung 12 injects
the principal into the proxy header (R5S:1243-1245), and Navidrome creates an account for any
unknown proxy user, so an opaque id there would orphan every member's history. The write path mints
the id for a member added by the page or the verb. The rule 5b-ii ruled, that `AddMember` refuses
the admin flag on every channel except 5c's CLI grant, carries over, and 5c's last-admin guard and
the `streamer_username` collision check run inside the write transaction, so a concurrent
self-demotion of the last two admins leaves one. The member-change audit record stays the log
record rung 5 built (`roster.go:316`); CH routes no member-audit table.

**The epoch** gets no bulk move. Records are named `sha256(kind, value)`
(`internal/auth/store.go:105-114`), which is one-way, so a re-key cannot reach an address removed
before 6d, and a copy races a concurrent bump from another process (`EpochStore.mu` is in-process
only, `internal/auth/epoch.go:66-72`). Instead, the revocation check reads both records, the one
keyed by the session's address and the one keyed by its resolved `UserID`, and takes the later
epoch, permanently. New bumps write the `UserID` key. Every epoch write takes an advisory lock on
the epoch directory, reads the stored value, and writes only a later one, so no process moves an
epoch backward. The records stay files (R18). This closes the reviews' scenarios: a rollback
to a 5d binary, a restored pre-6d snapshot, a re-added member, and a crash mid-way, since there is
no mid-way. R5S's "change of one derivation" (R5S:1689-1690) held for the code only.

---

## The 2.0 forward hooks

Per CLAUDE.md's build-for-2.0 rule, each pass states per task whether it lands a hook or is
neutral; a hook is built by the rung that first consumes it.

| Hook (S20) | Rung 6's position |
| --- | --- |
| 1, audio key and file hash (:363) | **Deferred** to rung 7's migration by ruling 2 (S20 erratum owed). Invariant 6 exempts both as alternate keys. |
| 2, folded sort key (:370) | **Deferred** to rung 9's migration by ruling 2 (S20 erratum owed). |
| 3, art blob (:372) | **Deferred** to rung 9's migration by ruling 2 (S20 erratum owed). |
| 4, timestamped events (:374) | **The rule lands at 6a** as invariant 3; `settings_audit` (6c) is its first table. |
| 5, stable identity (:376) | **Lands at 6c and 6d.** 6c creates users; 6d mints the id, adds `UserID`, and reads both epoch keys. |
| 6, universal driver seam (:379) | Neutral. |
| 7, schema migrations (:383) | **Lands at 6a.** |
| 8, collection writer (:385) | **Lands at 6c** as `writer` with its domain, epoch, and generation **[CH]**. |
| 9, transaction model (:387) | **Lands at 6a.** One writer, WAL, `FULL`, the write function. |
| 10, live-file backup consistency (:390) | **Lands at 6b-i**, by `VACUUM INTO`. |
| 11, gains (:392) | **Deferred** to rung 9's migration by ruling 2 (S20 erratum owed). |
| 12, portable layout | Neutral. Rung 7. |
| 13, sort tags; 14, ID3 version | Neutral. Rungs 8 and 9. |
| 15, configuration homes (:401) | **Lands at 6c** for the schema, the settings home, and the marker; the two mounts and the re-authentication gate are rung 14's **[CH]**. |

---

## Prior art

Scored on CLAUDE.md's four criteria. Versions and dates are from `go list -m` in a scratch module
on 2026-09-23; licenses from each module's `LICENSE`; importer counts from pkg.go.dev. The
mechanics review's measurements are at `/tmp/claude-1000/r6mech/` and summarized in the fold record.

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
| Navidrome's schema | Users and per-user tables | GPL, behavior only | | | | **Consult** at 6c's plan for behavior only; no code. |

The other entries CLAUDE.md names (OpenSubsonic, gonic, taglib, mewkiz/flac, Picard, chromaprint)
have no rung 6 surface.

---

## Risk model

The store holds the only record the greenfield server keeps of what it admitted. Its failure
classes, ordered by cost, each with the guards named above:

| Failure | Guarded by |
| --- | --- |
| data loss by restore | the exclusive lock, the one-file swap after a truncating checkpoint, the intent record, a crash injected at each step |
| data loss by backup (a wiped or wrong mount archives the remote) | the missing-root guard, the store-aware count guard, `--max-delete`, ruling 10's retention floor |
| a migration under a live older process | `serve`-only migration under the exclusive lock; the version check in every write transaction |
| a bad migration | per-migration transactions, the ledger, per-version snapshots (ruling 7), the newer-store refusal, the foreign-key check, populated fixtures |
| foreign keys silently off; lost durability | the DSN, the maintenance connection that never returns to a pool, the readback proof; `synchronous=FULL` |
| disk full | a free-space check before every full copy, `journal_size_limit`, the truncating checkpoint, the named error, `doctor`'s WAL and free-space rows; CH's upload floor stays rung 7's |
| a live job marked interrupted; settings flap | per-process leases taken before insert; `serve`-only seeding in one transaction |
| revocation lost at 6d; a root-owned store file | the permanent dual-key read and forward-only epoch writes; the uid check |

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
| WAL on one host, read-only opens creating `-shm`, growth without `journal_size_limit`; the hot-journal rule; statement-level undo | sqlite.org/wal.html sections 1 and 5; howtocorrupt.html section 1.4; lang_transaction.html section 3 |
| a `BLOB` column storing TEXT outside `STRICT` | verified by the mechanics review |
| `exec.CommandContext` cancellation and `WaitDelay` | the plan quotes `os/exec` |
| rclone sync, `--backup-dir`, `--max-delete`, exit codes, `RCLONE_CONFIG_*` | the plan quotes rclone's documentation |
| `flock` release on exit and `/proc/locks` | the plan quotes `flock(2)` and `proc(5)` |

### Every state a check can meet, with its report

| Check | States and reports |
| --- | --- |
| inspect open | absent, restore in progress, older (pending migrations named), newer (both versions), current, lock contended ("held by another process" with pids), network filesystem (refused), FUSE (warned), wrong uid (refused naming the owner) |
| writing open | absent: created by `serve` and `member`, refused by `backup`; zero bytes or not a database: refused; version differs in a non-`serve` verb: refused naming both; intent record present: refused naming `--resume`; ledger mismatch: refused naming both |
| migration | snapshot written, snapshot write fails (refused naming the path), free space short (refused), shared holder present (`serve` waits and logs) |
| job recovery | the four rows of the recovery table |
| backup | flag off: unavailable, no record; store absent: refused; root missing, root empty with no tracks, or count below the fraction: aborted before rclone; already running: conflict; rclone failure: failed with its exit status; succeeded |
| restore | the eight flow steps, each resumable; store held: refused; snapshot absent, corrupt, or newer: refused; current store unopenable: refused without `--discard-current`; non-empty root: refused without `--into-nonempty` |
| seed rule | the eleven rows of the table above |
| roster import | the six states in "Identity in the store" |
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
| a user's admin flag | the CLI grant alone, as 5b-ii ruled |
| a collection's `root_path` and `writer` | the server alone; no request type carries either |
| backup | the CLI verb, and over HTTP only behind rung 14's authorizer |
| restore and import | their CLI verbs alone, offline under the exclusive lock |

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
   the read pool) reads back on a freshly opened pooled connection.
5. A write through the read pool fails with a read-only error; an orphan insert through the write
   function fails with `SQLITE_CONSTRAINT_FOREIGNKEY` by its extended code.
6. A callback that swallows a failed statement's error does not commit; a nested write call fails
   fast; a write with no deadline is refused.
7. The write function fails closed when `user_version` is changed under it.
8. A migration failing after its version write leaves `user_version`, schema, and ledger unchanged;
   a ledger hash mismatch is refused naming both.
9. Migrating a populated version-N fixture leaves `pre-migrate-vN.db` with `user_version` N and the
   fixture's rows; a snapshot write that fails leaves the store at N with an error naming the path;
   a second migration never overwrites the lower-version snapshot.
10. A foreign-keys-off rebuild migration is followed by an orphan insert through the write function
    that fails.
11. Each invariant fails against its violating fixture for its named catalog fact, including a Go
    `string` bound into a `STRICT` path column.
12. A zero-byte file and a non-database file are each refused as not a store. A non-`serve`
    writing open of an older store refuses naming both versions; `serve` with a
    second handle holding the lock shared waits rather than migrating; a newer store is refused.
13. The inspect open against an absent, an older, and a newer store leaves the state directory
    byte-identical and reports each state by name; `doctor`'s store row renders each in a golden.
14. The `statfs` seam reports NFS (refused), FUSE (warned), and ext4 (opened); a uid mismatch is
    refused naming the owner; the database and a snapshot are created 0600.

### 6b-i: the ops seam and backup (lane 2 worktree)

1. Every catalog entry's privilege is `member` or `admin`; the zero value fails construction.
2. A body carrying an actor field is ignored, and the job records the caller's identity.
3. A record owned by a live second process (a second `ops/local` over the same store holding its
   lease) survives another process's open.
4. A queued and a running record whose owner lease is free read interrupted after an open.
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
    the fraction each abort with no rclone invocation recorded by the stand-in.
13. A stand-in exit 7 makes the job failed with the code in its record.
14. `backup` with an absent store refuses and creates nothing.
15. Setting each retired R2 name, and `DUBPLATE_BACKUP_STAGING`, fails the load with its
    supersession message.
16. `backup` receiving SIGINT mid-sync ends the job canceled, leaves no stand-in process after
    `WaitDelay`, and exits with the canceled code.
17. Each exit code maps to its outcome as the verb's help states.
18. `doctor`'s backup row renders never backed up, last success age, unavailable, and the crypt
    warning (ruling 14) in goldens.

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
7. A crash injected after each flow step leaves the intent record; a writing open refuses naming
   `--resume`; `--resume` completes.
8. A newer, corrupt, or absent snapshot is refused; an older one is refused without `--older`.
9. A tree restore into an empty root yields the backed-up files; into a non-empty root without
   `--into-nonempty` it refuses naming the flag.
10. The reconciliation report names a planted row without a file and a planted file without a row.
11. A restored file is owned by the leaf's owner at 0600.

### 6c: the schema and the settings home **[CH]** (lane 2 worktree if 5c and 5d leave `internal/config` untouched; otherwise after 5d)

1. The invariant test passes over the full schema, with the first populated-fixture migration.
2. An orphan collection owner is refused; a Go `string` bound into `rel_path` is refused.
3. A duplicate address, live or tombstoned, is refused under ruling 9's index.
4. `visibility` refuses any value but 1.0's.
5. A seed outside its bounds, validated through the registry's value validator with no environment,
   fails `serve`'s start.
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
14. A terminal record older than the retention is pruned at the next open, and the newest
    successful record of each kind survives.
15. `check` evaluates the seed rule read-only and leaves the store byte-identical.

### 6d: identity in the store and the handoff (after rung 5's 5d merges)

1. `member import` imports a valid roster once, writes the import record, and a second run refuses.
2. Each import refusal state reports as stated, each by its own fixture.
3. With `DUBPLATE_ROSTER_FILE` set, every verb fails the load naming `member import`.
4. A verified identity's `UserID` is the surrogate id and its `Principal` the streamer handle.
5. A member added through the write path receives a new id; the admin flag arrives only through
   the CLI grant; re-adding a removed address reactivates the same id (ruling 9).
6. A concurrent self-demotion of the last two admins leaves one admin.
7. A session revoked by an address-keyed bump before 6d, and one revoked by a `UserID`-keyed bump
   after, both stay revoked; a session revoked under neither key stays valid.
8. Two processes bumping one epoch concurrently never leave a stored value lower than either bump.
9. With the env holding A and an in-package admin row holding B, a link minted after the write
   carries B's expiry, a session's next refresh uses B's lifetime, the next request enforces B's
   absolute cap, and a limiter built at the next start uses B.
10. Each re-homed key seeds from the env and `doctor` shows its source and the `sensitive` marker on
    both lifetimes and the cap.
11. `check` names every roster member missing a `streamer_username` before a deploy.
12. `systemd-analyze verify` is clean, the state-dir leg covers the `store` leaf, and the unit's
    `ReadWritePaths` names it.
13. The template carries no `DUBPLATE_ROSTER_FILE` row, and `deploy.sh` pre-creates the `store` leaf
    and no longer the roster leaf.

### Band and ceilings

The ladder drew rung 6 as band M, one pass (BS:656; bands at BS:563-567). The ratified CH routes a
schema and a settings home here **[CH]**, and rung 5 hands rung 6 the surrogate id, the epoch
obligation (R5S:1809-1811), and the roster's move behind its interface (R5S:194-196, PA:209). Five
passes make it band L, which the build spec's ladder entry owes an erratum.

The ceilings use 5b-ii's measured rate, the nearest analog: segments A through C spent about 4M
across Tasks 1-8 including fix rounds, about 0.5M per task. Plan preparation adds about 0.85M per
pass (0.83M at 5a, 0.87M at 8a and 8b), the staleness pre-task about 0.2M, and the close
(simplifier, architecture readers, fold) about 0.6M. The draft priced tasks at 8b's chain rate
without fix rounds, and its 6d basis double-counted costs already inside 5a's actual. Two lessons
bind: a pass that changes the config contract overran (1b), and a pass over ten tasks overran
(5b-i).

| Pass | Tasks | Arithmetic | Ceiling | Flag (80%) |
| --- | --- | --- | --- | --- |
| 6a | 7 | 0.85 + 0.2 + 7 × 0.5 + 0.6 = 5.15 | 5.2M | 4.2M |
| 6b-i | 8 | 0.85 + 0.2 + 8 × 0.5 + 0.6 = 5.65 | 5.7M | 4.6M |
| 6b-ii | 5 | 0.85 + 0.2 + 5 × 0.5 + 0.6 + 0.3 targeted review = 4.45 | 4.5M | 3.6M |
| 6c | 8 | 0.85 + 0.2 + 8 × 0.5 + 0.6 = 5.65, plus 10% for a config-contract change | 6.2M | 5.0M |
| 6d | 7 | 0.85 + 0.2 + 7 × 0.5 + 0.6, plus 0.7 for security review inside the chain (5b-ii's pattern) = 5.85 | 5.9M | 4.7M |

The total is 27.5M against the draft's 17M; the difference is the fix-round rate and the scope the
reviews added. The phase ceiling and each lane's sub-ceiling are set at lane launch (PA section 4).
Each ceiling is re-derived from the previous pass's actual before the next plan is drafted, and a
plan that exceeds ten tasks splits the pass, names the follow-on, and re-derives both ceilings.

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
| `cmd/dubplate-server/main.go` (`newProductionRegistry`, `AddCommand`) | 6a registers the busy timeout; 6b-i the class and `backup`; 6b-ii `restore`; 6c its keys; 6d `member import`; rung 5's 5c adds the `member` group in the same block | a textual conflict keeps both lines in landing order, and merge ritual step 5 runs |
| `cmd/dubplate-server/doctor.go`, `check.go` | 6a the store row and the inspect report; 6b-i the backup row; 6c the settings rendering and read-only seed evaluation; 6d the import evaluation | rungs 5, 9, 13 add rows the same way |
| `go.mod`, `go.sum` | every lane adds its own requirements | a rebase conflict is resolved by `go mod tidy`, and merge ritual step 5 runs |
| The state directory layout | rung 4; 6a adds the `store` leaf | rung 16 converges |
| `internal/auth` roster and epoch, `cmd/dubplate-server/auth.go`, `internal/identity` | rung 5; 6d after 5d merges, with no concurrent auth-writing lane | rungs 7, 12, 14 |
| `systemd/dubplate-serve.service`, `scripts/deploy.sh`, the state-dir leg in `scripts/check.sh`, `env/dubplate-server.env.template` | 6d after 5d merges | rung 16 converges |

---

## Process and review

Each pass opens with a staleness pre-task, then runs the implementer, `diff-reviewer`, and gate
chain per task in segments of three to four tasks on green commits. Rung 6 is not one of the three
rungs that keep the full adversarial cycle, so review scales to risk. The data-loss paths earn one
targeted fresh-context read at 6b-ii's close, over the migration runner, the snapshot, and restore,
by an `opus` reviewer briefed with the risk model. Pass 6d runs `web-auth-security-reviewer` inside
its chain. At each merge the conductor dispatches one `go-architecture-reader` per touched package,
and every `code-simplifier` dispatch carries `.claude/instructions/simplifier-brief.md` verbatim.

Passes 6a through 6c are disjoint from rung 5's territory except the rows above, so they run as
lane 2 beside 5c and 5d. Pass 6d waits for 5d. Rung 6 edits the serve unit, `deploy.sh`, and the
template in the repo at 6d; the deploy rides the next ruled deploy, and the no-deploy carry in
`docs/STATUS.md` holds until 5d lifts it.

---

## Errata this spec owes

Each lands, dated 2026-09-23 or at ratification, in the named document when Geoff rules this spec.

1. **BS ladder entry 6** (BS:656): band M to L, five passes, for the reasons in "Band and ceilings."
2. **BS "Box safety during the build"** (BS:934-935): rungs 5 and 6 edit deploy units and scripts
   in the repo, and their deploy rides the next ruled deploy; the box sees them only then.
3. **BS config wall** (BS:218-220): the backup class is one remote plus a flag, not a pair, and its
   credentials are rclone's own env names (conditional on ruling 5).
4. **CH wall homing** (CH:66, CH:74-75): keys read before the store opens (the serve keys, the
   store's busy timeout) and keys read by a unit not granted the store leaf (the shadow pair) stay
   on the wall (conditional on ruling 3).
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
10. **R5S epoch** (R5S:1687-1690): the re-key is a permanent dual-key read, not a rename.
11. **S20 hooks 1, 2, 3, 11** (S20:363-372, S20:392): schema lands with the computing rung's
    migration (conditional on ruling 2).
12. **S20 prior art**: the `ncruces/go-sqlite3` row is a wasm2go translation, not wazero.
13. **BS U3** (BS:662-664): a track path is relative to its collection's library-relative root
    (clarifying).

---

## What depends on the configuration homes spec

Each passage marked **[CH]** exists because of CH's routes: rulings 1, 2, 3, and 13; the schema
tables other than the job records and the ledger, and the root-allocation paragraph; the settings
home, the registry metadata, and the seed table; pass 6c in full and 6d's re-homing half; the hook
table's rows for hooks 8 and 15; the contested rows for `internal/settings` and the registry files;
and job retention as a store-homed key. The foreign-key DSN rule does not depend on CH: CH requires
it (CH:524-526), and hook 9 and the schema's own references require it independently.

---

## Not verified in this revision

- `zombiezen.com/go/sqlite`'s license and importer count; its verdict rests on neither.
- The exact wording of `database/sql`, `os/exec`, rclone, `flock(2)`, and `proc(5)` that the
  planning-miss table sends each plan to quote.
- The line citations into rung 5's files, which 5c and 5d may move; `internal/auth/seam.go` was
  uncommitted in lane 1 when this revision was written.
- systemd's default start timeout, which bounds `serve`'s migration wait.
