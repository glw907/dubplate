# Rung 6 spec review: mechanics and feasibility

Reviewer lens: does every mechanism the draft relies on behave as stated. Target:
`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md` at `12a6941` (cited below as `spec:N`).
Evidence was gathered on 2026-09-23 in a scratch module under `/tmp/claude-1000/r6mech/`, using
`modernc.org/sqlite v1.59.0` (SQLite 3.53.4), `pressly/goose/v3 v3.28.0`,
`golang-migrate/migrate/v4 v4.20.1`, `ncruces/go-sqlite3 v0.35.6`, golangci-lint 2.13.2, Go 1.27.1,
sqlite.org and rclone.org pages fetched the same day, and the repo tree at `105ad50`.

## Verdict

The driver-level claims mostly hold. `_txlock=immediate` exists, the `_pragma` sort puts
`busy_timeout` first, foreign keys ride every pooled connection, FTS5 trigram works,
`LIKE_DOESNT_MATCH_BLOBS` is compiled in, and `VACUUM INTO` gives a consistent snapshot. The sqlite.org
quotes are accurate. The HISTORY actuals the ceilings cite are quoted correctly. The spec does not
break on the driver. It breaks one layer up, where its own mechanisms meet a second process, a second
user, or a second rung's code, and it breaks in four places. Migrate-on-open runs under a *shared* lock
while an older serve process is live. The roster-file ruling needs a store fact that `config.Load`
cannot see by construction. Nothing in SQLite enforces the `BLOB` path rule. I proved empirically
that the same path bytes stored once as TEXT and once as BLOB both pass a UNIQUE index. The epoch
re-key cannot move records whose address it no longer knows. Two further defects are silent:
foreign keys switched off for a table rebuild leak back into the write pool, and the env-seed rule
re-seeds from whichever process boots with a different environment. One claim was listed as
unverified: whether the busy handler runs on a deferred upgrade. It is now verified, and it does
not run. The pass split undercounts 6b, and 6d's ceiling double-counts its basis. Two findings are
blockers, sixteen are major, and fifteen are minor. Two findings carry an OWNER FORK.

## Findings

### 1. Blocker: forward migration runs under a shared lock beside a live older process

**Spec:** 197-199 ("A process that serves or writes holds it shared"), 204-206 (`serve`, `backup`,
and `member` create and open the store), 244-252 (the runner migrates on open and snapshots first),
248-250 (a newer store is refused at open).

**Defect:** Every writing verb migrates on open, and every writing verb holds the lock *shared*. Take
the deploy window: `deploy.sh` has swapped the binary but serve has not restarted yet. A timer-run
`backup` in that window migrates the store under the running older serve. So does an operator's
`member add` after 6d. The newer-store refusal is checked only at open, so the old serve never sees
it. The old serve keeps writing against a schema its code does not know: it misses new NOT NULL
columns and new constraints. The pre-migration snapshot is also not atomic with the migration. The
other process can commit between the `VACUUM INTO` and the first migration transaction, and that
commit is then missing from the "recoverable" snapshot.

**Evidence:** flock semantics tested (`/tmp/claude-1000/r6mech/flk`): two `LOCK_SH` holders coexist
by design. The spec's own risk model names the two processes (spec:592-595, "Two processes write the
store: serve and a timer-run `backup`").

**Proposed fold:** Migration requires the store lock *exclusive*, taken non-blocking. A writing verb
that finds pending migrations while another process holds the lock shared refuses and names the
reason ("store needs migration N→M; stop serve"). Otherwise it holds exclusive through the snapshot
and every migration, then downgrades to shared. Separately, the write function checks
`PRAGMA user_version` inside each `IMMEDIATE` transaction against the compiled maximum, and fails
closed on a store that moved ahead under a live process. That check is one integer read per write.

### 2. Blocker, OWNER FORK (touches ruling 6): "after the import `DUBPLATE_ROSTER_FILE` is a load error" is not implementable as written

**Spec:** 91-101 (ruling 6), 508-510 (the import runs "on the first open that finds the roster file
set and the users table empty"), 644 (the state table: "file set and users populated: the ruling 6
load error"; "file unset and users empty: `doctor` reports no members").

**Defect:** `DUBPLATE_ROSTER_FILE` is a `Required` key of the magic-link identity source today
(`cmd/dubplate-server/auth.go:44-46`), so an unset key is already a load error. The
"file unset" row of spec:644 contradicts the tree. The new rule makes the key's legality depend on
whether the users table is populated. `config.Load` runs before the store opens, and the spec itself
forbids `config` from importing `store` (spec:180-181). So the condition cannot be a *load* error.
The lifecycle also cannot converge in one deploy. The import needs the key set. The next start with
the same rendered env then hits "file set and users populated" and fails. The estate needs one deploy
with the key and a second deploy without it, and the template has to change between them. That
contradicts spec:497-499 ("no template row removed").

**Evidence:** `auth.go:44-46` (`Required: []string{"DUBPLATE_ROSTER_FILE", ...}`), and
`config.Load` in `internal/config/config.go:346`, which resolves the wall from `environ` alone.

**Proposed fold, for the owner:** Replace the implicit import-on-open with an explicit one-shot verb,
`dubplate-server member import --roster PATH`, which runs in one transaction and refuses a populated
users table. 6d then registers `DUBPLATE_ROSTER_FILE` through `AddSuperseded` with a message naming
the verb, and removes it from the source's `Required` list in the same commit. That makes it a true
load error with no store dependency, and the upgrade becomes one deploy plus one command. If the
owner keeps import-on-open, the key must first move from `Required` to `Optional`. The "populated
plus set" refusal must then be a store-open error that `check` also evaluates. Name it as such, not
as a load error.

### 3. Major: a `BLOB` declaration does not enforce U3; a Go string silently stores TEXT and defeats UNIQUE

**Spec:** 268-269 (invariant 1: "named with a `path` suffix and declared `BLOB`"), 316-322 (U3:
"crosses the store boundary as `[]byte`... Paths compare bytewise"), 312 (`root_path` unique).

**Defect:** In an ordinary (non-STRICT) table, a column declared `BLOB` has BLOB affinity. It stores
whatever storage class the caller binds. One `string` bind anywhere in a later rung stores TEXT. A
bytewise lookup by `[]byte` then misses that row, and a UNIQUE index lets the TEXT form and the BLOB
form of the same path coexist. Invariant 1 inspects only the declared type, so it passes this
schema. The "no driver conversion can touch it" guarantee holds only as a Go coding discipline, and
nothing enforces it.

**Evidence (exp run):**
```
C typeof when bound as string: text
C match by []byte param: 0
C unique violated by same bytes as blob? err: <nil>
C strict insert string into BLOB: constraint failed: cannot store TEXT value in BLOB column trs.path (3091)
C check typeof insert string: constraint failed: CHECK constraint failed: typeof(path)='blob' (275)
```

**Proposed fold:** Every table is `STRICT`, or every path column carries
`CHECK (typeof(col) = 'blob')`. Invariant 1 then asserts the enforcement through the catalog: the
`strict` column of `pragma_table_list`, or the CHECK text. Add a proof that binds a Go `string` into
a path column and expects a constraint failure.

### 4. Major: switching foreign keys off for a table rebuild leaks back into the one-connection write pool

**Spec:** 253-257 ("The runner holds one dedicated connection for such a migration"), 600-601 (the
risk "Foreign keys silently off", guarded "by the DSN rule").

**Defect:** The DSN applies `foreign_keys` only when a connection opens. The write pool holds one
connection. A dedicated `db.Conn` that runs `PRAGMA foreign_keys = OFF` and then `Close()`s returns
that same connection to the pool with enforcement still off. Every later write in the process then
runs unenforced. This is exactly the silent failure spec:600 names, and the DSN rule does not guard
it.

**Evidence (exp run):**
```
G fk after dedicated conn returned to write pool: 0
G orphan insert via write pool after leak, err: <nil>
```

**Proposed fold:** Before releasing a rebuild connection, the runner either re-enables foreign keys
and reads the value back, or discards the connection by returning `driver.ErrBadConn` from
`(*sql.Conn).Raw`. Add a proof: after an FK-off migration, an orphan insert through the ordinary
write function fails.

### 5. Major: the read pool's mode is unspecified, and both choices have a sharp edge

**Spec:** 211-213 ("The read pool serves concurrent readers"), 216-220, 408-415 (the snapshot), and
771-772 (listed as not verified: whether the busy handler runs on a failed deferred upgrade).

**Defect and evidence:** The unverified item is now verified. The busy handler does *not* run on a
deferred read-to-write upgrade: a `busy_timeout` of 2000 returned `SQLITE_BUSY` in 0s. The
sqlite.org busy-handler page gives the rule: "If SQLite determines that invoking the busy handler
could result in a deadlock, it will go ahead and return SQLITE_BUSY."
```
E  second writer waited 531ms exec=<nil> commit=<nil>          (IMMEDIATE, two pools: waits)
E2 deferred upgrade after foreign commit: err=database is locked (517) after 0s   (BUSY_SNAPSHOT)
E3 deferred upgrade while writer holds lock: err=database is locked (5) after 0s
```
Any write that reaches the read pool, which has no `_txlock`, therefore fails instantly under
contention. The obvious guard is to open the read pool with `_query_only=1`. But
`VACUUM INTO` then fails on that pool:
```
F vacuum into on query_only conn: attempt to write a readonly database (8)
F vacuum into inside tx: cannot VACUUM from within a transaction (1)
```
Running the snapshot on the write pool instead would hold the process's only write connection for
the whole snapshot, which stalls every in-process writer.

**Proposed fold:** State three connection classes. The write pool is one connection with
`_txlock=immediate`. The read pool uses `_query_only=1`. The snapshot uses one dedicated,
non-query-only connection outside both pools, with no open transaction. Move the busy-handler row
from "Not verified" into the quoted-behaviors table with the result above.

### 6. Major: cross-process cancellation and non-exclusive job recovery are undefined

**Spec:** 353-368 (cancellation is "an explicit operation, never a reliance on context cancellation
reaching across a process"; "A job runs in the process that started it"; recovery marks interrupted
"every running record whose exclusive lock no live process holds"), 400-401 (the contract suite).

**Defect:** Job records are shared through the store, so serve's `ops/local` lists the backup a CLI
process started. Once rung 14 mounts HTTP, nothing reaches that job with a cancel: the job's context
lives in the other process, and the spec names no mechanism. The recovery rule covers only exclusive
kinds. A running non-exclusive record has no lock to probe. At open, the recovering process cannot
tell a dead job from one live in another process. The ordering is also unstated. If the runner
inserts the `running` row before it takes the kind lock, a concurrent opener sees "running, lock
free" and marks a live job interrupted, and "a terminal state never changes" then strands it.

**Evidence:** flock test: a lock is visible across processes, and within one process across separate
open file descriptions (`flock f2 EX nb same process: resource temporarily unavailable`). The
mechanism is therefore sound for exclusivity, but the spec gives non-exclusive jobs no lock at all.

**Proposed fold:** Every job holds a lock for its life, whether per job or per process instance, and
the runner takes it *before* inserting the record. Cancellation writes a `cancel_requested` column
that the owning runner polls at its progress points. The contract suite adds a case where the job
runs in a second process, or in a second `ops/local` over the same store, and the cancel comes from
the first.

### 7. Major: the env-seed rule runs in every process that boots, with different environments

**Spec:** 453-466 ("The env-seed rule runs once at boot... before any consumer reads a setting"; an
`env-seed` row with a different env value "re-seeds the row"), 444-446 (version invalidation).

**Defect:** Serve, the timer-run `backup` unit, and the operator's CLI verbs each boot, and each
carries its own environment. CH itself says the failure-ping unit has "its own env" (CH:61-62). Any
process whose env disagrees re-seeds the row, and the next serve boot re-seeds it back. The value
flip-flops by whichever process started last, and no error is ever raised. Cache invalidation is the
second half. A write from a CLI process must invalidate serve's in-memory cache, so the version has
to be read from the store. `PRAGMA data_version` is per connection, and a pool hands out different
connections. The spec does not say where the version lives.

**Evidence:** spec:459-463 as written; CH:61-62; sqlite.org pragma.html on `data_version` ("provides
an indication that the database file has been modified" on *this* connection).

**Proposed fold:** Only `serve` applies the seed rule, or only a process that holds a named
"seeding" role. Every other verb evaluates it read-only, the way `check` does (spec:454-455), and
reports a disagreement. The version is a settings-generation integer bumped inside each settings
write transaction. The accessor reads it, at one indexed read, to decide whether its cache is stale.

### 8. Major: a row "written at its default" pins the old default after an upgrade

**Spec:** 459 ("no row / absent → the row is written at its default, source `default`"), 461
("`default` / absent → nothing changes"), 474-475 (researched defaults).

**Defect:** A later release that re-researches a default never reaches an existing store. The row
holds the old value with source `default`, and the table says nothing changes.

**Proposed fold:** A `default` row stores no value, or the accessor ignores the stored value for
source `default` and resolves the code default. Add a proof: change a registered default and reopen
the store, and the new default wins.

### 9. Major: "a consistent snapshot of the store plus the library tree" overclaims

**Spec:** 15-16, 44-51 (ruling 1), 408-419 (snapshot first, then `rclone sync` of the tree).

**Defect:** The store snapshot is consistent, but the pair is not. The tree sync can run for hours
after the snapshot. Any rung 9 move that commits in that window leaves the snapshot pointing at path
A while the remote holds the file at path B. `sync --backup-dir` moves A into the dated archive. A
restore then yields rows whose files are missing from the library. The move-intent journal
(spec:304, 598-599) covers crash consistency inside one box, not a backup's two-phase capture.

**Evidence:** rclone docs, `--backup-dir`: "for any files which would have been updated or deleted
will be stored in remote:old".

**Proposed fold:** Pick one approach and state it. The backup job holds the library-mutation lease
that rung 9 must honor. Or the snapshot is taken *after* the tree sync, and the restore runbook
reconciles rows against the archive. Or restore runs a named reconcile that reports orphans. Soften
spec:15 to what is guaranteed. This decides rung 9's lease surface, so it belongs in the forward
constraint spec:118 already gives rung 9.

### 10. Major: the salvage empty-library guard does not survive per-collection roots

**Spec:** 417-420 ("a missing library root and an empty one each abort before rclone runs"),
585-586.

**Defect:** The salvage guard checks only top-level entries, and argues "an empty top level and an
empty tree are the same fact" (`_salvage/internal/backup/backup.go:125-128`). That argument runs one
way only. A wiped mount whose collection directories still exist, which is likely with
server-allocated roots (spec:288, 306-312), passes the guard with zero files. `sync` then moves the
whole remote library into the archive, and "pruning that archive is the remote's own lifecycle
policy" (spec:50-51) deletes it on schedule.

**Proposed fold:** Make the guard store-aware. Abort when the store records N tracks and the tree
walk finds far fewer regular files. Also pass rclone's `--max-delete` bound, derived from the same
count. The fresh review the spec promises should record this reversal of the salvage argument.

### 11. Major: the epoch re-key cannot move every record, and it races concurrent bumps

**Spec:** 519-525 ("No member's revocation is lost, including across a crash mid-way"), 604, 645.

**Defect:** Records are named `sha256(kind, address)` (`internal/auth/store.go:109-114`), which is
one-way. The re-key can map only the addresses in the roster at import. An address removed from the
roster before 6d leaves a record nobody can re-key. After the completion marker, the old key is not
consulted. If that address is re-added and receives a new surrogate id, its pre-removal sessions
verify again, because `Session` carries `Address` (`internal/auth/session.go:71-74`) and the new key
has no epoch. The re-key "writes the new name before removing the old one." A bump that lands on the
new key first is then overwritten backward by the copy. `EpochStore.mu` is in-process only
(`epoch.go:67-72`), and CLI processes can bump.

**Proposed fold:** Drop the bulk move. The revocation check computes both digests from the
session's address and its resolved surrogate id, and it takes the later epoch. New bumps write the
surrogate key. The address-keyed read is never retired, or it is retired only when the session cap
has elapsed since 6d. This deletes the crash-safety machinery and its four-state proof, and it loses
nothing.

### 12. Major: `Principal` becoming the surrogate id breaks the streamer's user mapping

**Spec:** 510-512 (the resolution point "returns the surrogate id as the identity's `Principal`").

**Defect:** R5S says rung 12 "injects the resolved principal inside `Rewrite` on the proxy path"
(R5S:1243-1245), and `identity.go` documents `Principal` as "what names this member to a downstream
consumer." Navidrome creates an account for any unknown proxy-header user. An opaque surrogate id
there creates a fresh, unreadable Navidrome user per member and orphans that member's history.

**Evidence:** navidrome.org docs, external authentication: "If a user is successfully authenticated
by the proxy but does not exist in the Navidrome database, it will be created with a random
password."

**Proposed fold:** `identity.Identity` gains a `UserID` field for the surrogate. `Principal` stays
the streamer handle, `streamer_username`, required after 6d. Per-user rows key on `UserID`. The
`identity.go` comment "rung 6 can substitute a durable surrogate id here" is then corrected in 6d.

### 13. Major: the apply-mode column describes pure functions, but the tree captures every value once

**Spec:** 483-493 (the seed-handoff table), 497-498 ("a statement of what the code does").

**Defect:** All ten auth values resolve once into `authConfig` (`auth.go:402-416`). The doc comment
says the struct passes them to the auth core "as plain construction arguments," which is the
"long-lived cached struct" CH forbids (CH:89-90). `MintLink`, `RefreshSession`, and `VerifySession`
take the value as a parameter, so "next operation" and "immediately" hold only if the handler layer
reads the accessor per request. 5c and 5d build that layer, and today `resolveAuthConfig` has no
production caller. The table's rows are requirements on unwritten code, not statements of what the
code does. Meeting them means 6d dismantles `authConfig` and rewires every consumer 5c/5d wires. The
7-task estimate does not count that work.

**Proposed fold:** Relabel the column as a required behavior. Add a 6d task that replaces
`authConfig` with accessor reads at each call site. Add a proof per non-restart row: change the
store value with no restart, and the next operation sees it.

### 14. Major: the registry has no place for the bounds that every settings write must check

**Spec:** 445-451 (the registry "gains three pieces of metadata": home, apply mode, sensitive),
468-469 ("Every write validates the value against the key's bounds and against every correlated
rule"), 488 (the correlated retention rule "moves from the source's `Validate` hook").

**Defect:** `config.Threshold` is `{Key, Kind, Default, Doc}` (`internal/config/thresholds.go:45-50`).
All ten auth bounds live in `validateMagicLinkSource`, which reads `environ`
(`auth.go:230-280`, for example `checkIntBounds(errs, environ, authRequestBodyCapThreshold, ...)`).
A store write has no environment to hand it. 6c's registry change therefore needs a fourth metadata
item: a per-key validator, or bounds, plus a registry-level correlated-rule list. Without it, the
6c exit row "a seed outside its bounds fails the load" has nothing to call.

**Proposed fold:** Name the fourth metadata item in spec:445-451 and in 6c's scope. The re-homing in
6d moves each bound from the source hook into that registry entry.

### 15. Major: CLI verbs run as a different user poison the store's files

**Spec:** 193-195 (leaf 0700, files 0600), 204-206, 425-433 (restore), 454-455 (`check` opens
read-only).

**Defect:** The serve unit runs as `User=music` (`systemd/dubplate-serve.service:50`). No spec says
which user runs `restore`, `doctor`, `check`, or the `member` verbs. An operator who runs them as
root does damage. `restore` renames a root-owned 0600 snapshot into place. A read-only open with no
`-shm` present creates a root-owned `-shm`, because wal.html requires write access to create it. In
both cases serve can no longer open the store. SQLite creates the `-wal` and `-shm` files with the
database file's mode, so the owner problem propagates.

**Evidence:** sqlite.org wal.html: "The opening process must have write privileges for '-shm'...
or else write access on the directory." `grep` finds no run-as-user guidance in R5S, CH, or BS.

**Proposed fold:** Every verb that opens the store refuses when the effective uid differs from the
leaf's owner, and names the fix (`runuser -u <owner> -- dubplate-server ...`). `restore` asserts the
restored file's owner and mode before its rename.

### 16. Major: 6b is undercounted at 8 tasks, which breaches the spec's own ten-task rule

**Spec:** 675 (6b's scope), 696 (8 tasks, 4.5M), 700-701 (a pass over ten tasks splits).

**Defect:** 6b's scope, counted one deliverable per task at the granularity 5a, 8a, and 8b used,
covers ten or more units:

1. The ops types, sentinels, and codes.
2. The catalog with availability and privilege.
3. The job model with locks, recovery, and cancellation.
4. `ops/local`.
5. The projection with its authorizer.
6. The contract suite.
7. The rclone transport with its guards.
8. The snapshot.
9. The offline restore with the three-file move and library restore.
10. The two verbs, with SIGINT polling.
11. The integration class and its supersessions.
12. `doctor`'s row.

Finding 6's cross-process cancel adds one more unit. 5b-i ran ten tasks and overran (HISTORY:578-590).

**Proposed fold:** Split now, at spec time. 6b-i covers the ops seam, the job model, and the
contract suite, using a no-op or fake long-running operation. 6b-ii covers backup, restore, and the
verbs. Re-derive both ceilings from 8a's per-agent rate.

### 17. Major: the invariant test's rules 3 to 6 are not decidable from the catalog as written

**Spec:** 263-278, 675 ("each invariant fails against its violating fixture").

**Defect:** `pragma_table_info` and `pragma_foreign_key_list` return names, declared types, nullness,
and key targets. Rule 4 ("namespaced by its source"), rule 5 ("attribute bag"), and rule 6 ("opaque
random value") each need a naming convention to become checkable. Rule 3 binds only tables whose
names end in `_events`, so a `play_history` table evades it. With a heuristic, the "violating
fixture" proof only shows that the heuristic catches the one violation it was written for. Rule 2
is also too weak. SQLite accepts a foreign key whose parent column exists but is neither PRIMARY KEY
nor UNIQUE, and then fails every insert.

**Evidence:**
```
fk to non-unique parent insert: SQL logic error: foreign key mismatch - "c" referencing "p" (1)
fk_list cols: [id seq table from to on_update on_delete match]
```

**Proposed fold:** Give each rule its decidable proxy in the spec. Rule 3: a table carrying a `kind`
column is an event table and must carry a non-null `at`. Rule 4: every `*_mbid` column is named
`mb_<entity>_id`, and a bare `release_id` or `recording_id` is refused. Rule 5: refuse a table
matching (FK, TEXT `name` or `key`, untyped `value`). Rule 6: a column named `id` or `*_id` that is
not an FK must not be `INTEGER PRIMARY KEY`. Rule 2 adds the check that each FK's parent columns
are the PK or carry a UNIQUE index. Where a rule is only a convention, the spec says so.

### 18. Major: 6a lands an exported store package with no production caller

**Spec:** 674 (6a's scope, with no cmd consumer), 204-206; CLAUDE.md's 2026-09-12 refinement
("nothing is exported ahead of a caller").

**Defect:** 6a merges `internal/engine/store`'s exported API on its own, and the first production
caller is 6b's `backup` verb. The exported-unused leg (`scripts/check.sh:465-522`) warns today and
gates later. The 6a exit row "`check` never creates one" also needs a `check.go` edit, and the
contested-surface rows (spec:705-726) do not list `check.go`.

**Proposed fold:** 6a includes `doctor`'s store row (absent, version, filesystem type) and `check`'s
read-only absent report as its production callers. Add `cmd/dubplate-server/check.go` to the
contested rows.

### 19. Minor: SQLite creates the database and the `VACUUM INTO` output at 0644, not 0600

**Spec:** 193-195.

**Evidence (`perm` run, umask 0022):** `snap.db 644`, `fresh.db 644`. `-wal` and `-shm` inherit the
main file's mode (`600` when the main file is pre-created at 0600). An empty target pre-created
with `O_EXCL` at 0600 keeps 0600 through `VACUUM INTO`, and sqlite.org permits it: "must not
previously exist, or else it must be an empty file."

**Fold:** Pre-create the database file and every snapshot target with `O_CREATE|O_EXCL`, mode 0600.
Add a proof that asserts the modes.

### 20. Minor: the `synchronous` level must be FULL, not merely "explicit"

**Spec:** 227-229.

**Evidence:** pragma.html: "WAL mode does lose durability. A transaction committed in WAL mode with
synchronous=NORMAL might roll back following a power loss." With NORMAL, a committed move intent
can vanish after its file has moved, which breaks the journal pattern spec:598-599 relies on. The
build's own default is `DEFAULT_WAL_SYNCHRONOUS=2` (FULL), read from `pragma compile_options`.

**Fold:** State `synchronous=FULL`.

### 21. Minor: the three-file restore move is not atomic

**Spec:** 427-432.

**Defect:** A crash between the three renames can leave the old `-wal` beside the new database,
which is the mispairing that howtocorrupt section 1.4 warns about.

**Fold:** Restore opens the current store with the lock held exclusive, runs
`PRAGMA wal_checkpoint(TRUNCATE)`, closes the store, and verifies that no `-wal` or `-shm`
remains. Only then does it move the single database file aside and rename the snapshot into place.

### 22. Minor: DSN mechanics, citation drift, and pragma binding

**Spec:** 222-227, 564.

**Defect:** The fail-before-partial-change property the prior-art row praises holds only for the
shorthand keys. The module says: "_pragma values are the one exception: they are executed verbatim
and cannot be checked here, so a malformed _pragma can still fail partway" (`sqlite.go`, in
`applyQueryParams`). The spec sets the settings as `_pragma` values. The `_pragma` sort sits at
`sqlite.go:432-456`, not 425-447. The migration runner also cannot bind a parameter to
`PRAGMA user_version = ?`:
`SQL logic error: near "?": syntax error`.

**Fold:** Use `_journal_mode=WAL&_foreign_keys=1&_synchronous=FULL&_busy_timeout=N&_txlock=immediate`.
Their apply order is busy_timeout, `_pragma`, foreign_keys, journal_mode, then synchronous. Fix
the citation. The runner formats the version integer into the statement.

### 23. Minor: `atomicfile` has no function that renames an existing file

**Spec:** 413-414 ("an atomic rename"), 572, 597-598 (cites `atomicfile.go:45-55` as the directory
fsync).

**Defect:** Lines 45-55 are `WriteStream`'s doc comment. `WriteFile` does not fsync the directory.
Neither function can rename a file that `VACUUM INTO` already wrote, short of copying it through a
writer.

**Fold:** The store does `os.Rename` plus a directory fsync itself, or `atomicfile` gains an exported
rename helper. If it gains one, `atomicfile` becomes a contested surface and needs a row.

### 24. Minor, OWNER FORK: remote layout, and whether the snapshot is encrypted at rest

**Spec:** 53-60 (one `remote:path`), 417-423.

**Evidence:** rclone docs, `--backup-dir`: "You must use the same remote as the destination of the
sync. The backup directory must not overlap the destination directory without it being excluded by a
filter rule."

**Fold:** Fix the layout derived from the one spec: `<remote:path>/library`,
`<remote:path>/archive/<date>`, and `<remote:path>/snapshots/<stamp>.db`. **Owner fork:** after 6d
the snapshot holds member addresses. Decide whether the runbook recommends, or `doctor` warns
without, an rclone `crypt` remote.

### 25. Minor: in-process writers bypass the busy timeout, and nested writes deadlock

**Spec:** 232-235.

**Defect:** Inside one process, a second writer queues in `database/sql`'s pool for the single
connection. That wait is bounded only by its context, never by `busy_timeout`. A write function
called from inside another write callback waits on itself forever.

**Fold:** The write function requires a context with a deadline. The spec forbids nesting, and a
test proves a nested call fails fast.

### 26. Minor: `backup` should not create an absent store

**Spec:** 204-206.

**Defect:** A mistyped `DUBPLATE_STATE_DIR` creates an empty store. The run then uploads an empty
dated snapshot, and a restore of "the latest" would pick it.

**Fold:** Only `serve` and the `member` verbs create the store. `backup` refuses an absent store.

### 27. Minor: the network-filesystem refusal needs a defined set

**Spec:** 199-202, 674.

**Defect:** FUSE is ambiguous: sshfs and rclone mount are remote, while mergerfs is local and common
on home music servers. A blanket refusal of FUSE blocks legitimate adopters. The test also needs a
`statfs` seam, which the architecture reader grades as a test-only seam.

**Fold:** Refuse the NFS, CIFS, SMB2, and 9p magic numbers, and warn on FUSE. Name the seam's
justification.

### 28. Minor: flock cannot name the holder, and readers take no lock

**Spec:** 63-67 ("the refusal names the holder"), 198, 642.

**Evidence:** the flock test; `/proc/locks` lists a FLOCK entry with its pid and inode, and that is
the only source of the holder.

**Fold:** Name the holder by matching the lock file's inode in `/proc/locks`, or narrow the promise
to "held by another process." `doctor` and `check` also take the lock shared and non-blocking. They
report "restore in progress" rather than read a file that `restore` may be swapping.

### 29. Minor: the acting identity must not ride the wire type

**Spec:** 346-350.

**Defect:** "Every request carries the acting identity" conflicts with "never from a request body
field" when the request type is the wire type.

**Fold:** The identity is a method parameter, or a field tagged `json:"-"`, and the contract suite
asserts that a body carrying an actor is ignored.

### 30. Minor: `goose`'s decline rests on the wrong mechanism

**Spec:** 566.

**Evidence:** Module graph pruning means an importer never inherits goose's nineteen direct
requirements. A scratch program using goose's `Provider` with modernc gained 4 small modules
(interpolate, go-retry, multierr, x/sync), 12 `go.sum` lines, no database driver, and no moby. The
binary grew from 9.7MB to 12.0MB. The real disqualifier is behavioral. An older binary's
`Provider.Up` on a newer database returns no error:
`older binary up on newer db: 0 <nil>` / `db version 2`. goose also keeps its own version table,
not `user_version`.

**Fold:** Keep the verdict. Replace the reasoning with the measured cost and the silent-ahead
behavior that the runner must refuse (spec:248-250). golang-migrate's go.mod pins
`modernc.org/sqlite v1.18.1` and `mattn/go-sqlite3`. Under MVS our pin wins, so "version pairing
becomes ours to track" really means "its driver is tested against a four-year-old modernc."

### 31. Minor: the ncruces fallback row is stale

**Spec:** 565.

**Evidence:** `ncruces/go-sqlite3 v0.35.6`'s README: "It wraps a Wasm build of SQLite, and uses
wasm2go to translate it to Go." Its go.mod requires `go-sqlite3-wasm/v6`, and wazero appears nowhere
in its requirements.

**Fold:** Correct the row, and the matching S20 prior-art row.

### 32. Minor: 6d's ceiling basis double-counts and ignores the closest analog

**Spec:** 698.

**Defect:** "5a's actual plus its security review (0.18M) and fold (0.22M)" adds costs already
inside 5a's 3.9M (HISTORY:1735-1739). The auth-handler pass 5b-ii carries a 7.5M ceiling and a
projection near 7M (STATUS). It is the nearer analog for a pass that edits auth, adds security
review inside the chain, and rewires consumers (finding 13).

**Fold:** Re-derive 6d from 5b-ii's actual once HISTORY records it, and restate the arithmetic.

### 33. Minor: job retention arrives a pass after the job runner

**Spec:** 367-368, 473-474, 675-676.

**Defect:** 6b's runner writes terminal records with no retention sweep until 6c. That is
harmless, but unstated.

**Fold:** Say so in 6b's scope, or land a fixed retention constant at 6b that 6c re-homes.

## Claims verified as stated

`_txlock=immediate` exists (`sqlite.go:385-391`). The `_pragma` sort puts `busy_timeout` first and
orders the rest lexicographically. Foreign keys are on for a second pooled connection. FTS5 trigram
answers a substring query. The build carries `ENABLE_FTS5`, `ENABLE_SNAPSHOT`, and
`LIKE_DOESNT_MATCH_BLOBS`, both in `lib/sqlite_linux_amd64.go` and in `pragma compile_options`.
`user_version` and the DDL roll back with a failed transaction. A multi-statement `Exec` error rolls
back cleanly. Under `IMMEDIATE`, a second writer across pools waits and succeeds. `VACUUM INTO`
refuses inside a transaction and refuses an existing non-empty target. Its output passes
`integrity_check` and opens in rollback-journal mode. `mode=ro` does not create an absent file, and
the default open does. Every quoted sqlite.org passage matches the live pages: lang_vacuum,
lang_transaction, pragma, wal section 1, and howtocorrupt sections 1.2, 1.4, and 2.5. The seed
table's line citations into `auth.go` (99, 105, 111, 117, 166-184, 190, 196) are exact. The
HISTORY actuals for 8b, 8a, 5a, 1b, and 5b-i are quoted correctly. `rung 5` holds no reserved class
in `keys.go`, so 6b's `keys.go` edit is uncontested in practice. golangci-lint's `unused`, with
`exported-fields-are-used: false`, does not flag a JSON-marshaled wire type's unread field
(probe: `0 issues`), so the ops wire types are safe under the linter.
