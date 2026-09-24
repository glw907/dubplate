# Rung 6a plan review: domain risk (data integrity and failure)

Target: `docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md` at `11f11c2`
(revision 1). Sources: the rung 6 spec, revision 2.2; the 2.0 streamer spec's forward
hooks; `CLAUDE.md`. One lens: what a crash, a second process, a partial write, a full
disk, an upgrade, a restore, a foreign file, a wrong owner, or a network filesystem does
to the store, and whether a 6a seam forces a later rung into rework.

Every behavior below that the plan does not already quote was measured on 2026-09-23
against `modernc.org/sqlite v1.59.0` (SQLite 3.53.4) and `go1.27.1`. The probes are
under `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/82a7130f-dc63-4c92-b308-9af3d3bb0dde/scratchpad/6a-risk-probes/`
(`main.go`, `bt/main.go`) and may be gone.

**Counts:** 1 blocker, 4 major (one an OWNER FORK), 12 minor. Ranked by consequence.

The plan is strong on the risks it names. The commit rule, the version fence, the
conversion procedure, the 0600 pre-create, and the OFD decline are all correct and
measured. The findings below are the states it does not name and the one mechanism the
probes contradict.

---

## Blocker

### D1: after an automatic rollback, the callback's later statements autocommit

**Location:** R-6a10 (`:412-433`); Task 3 outcomes and states table (`:881-920`); Task 3
criteria (`:924-944`).

**Defect.** R-6a10 and Task 3 make the write method safe by recording the first
statement error and refusing to commit. That is not enough when SQLite has already
rolled the transaction back. The connection then drops back to autocommit mode, so every
statement the callback issues after the swallowed error commits by itself. The plan's
own probe shows the precondition: after `SQLITE_FULL`, `Commit` answers "no transaction
is active." This probe shows the consequence, on the driver's `*sql.Tx`, one connection,
`_txlock=immediate`:

```
insert into small values(2)            -> <nil>
insert into t values(zeroblob(100000)) -> database or disk is full (13)
insert into small values(3)            -> <nil>
Rollback                               -> cannot rollback - no transaction is active (1)
rows in small afterward                -> [1 3]
```

Row 2 was rolled back, and row 3 committed outside any transaction. A callback that
swallows a disk-full error and keeps going therefore commits a partial, non-atomic write
while the write method reports failure. The spec names four codes that can do this:
`SQLITE_FULL`, `SQLITE_IOERR`, `SQLITE_INTERRUPT`, and `SQLITE_NOMEM`
(`lang_transaction.html` section 3, spec `:327-329`). The states-table row "returns nil
after swallowing a statement error → rolled back" (`:919`) is false for these codes. This
is the commit rule every later rung inherits, and hook 9's projector joins it.

**Fold.** Once the transaction handle records an error, every later statement method
returns that recorded error without reaching SQLite, so nothing can run in autocommit
mode. Add a Task 3 criterion in the probe's shape. The fixture holds a table with one
row. `max_page_count` is pinned at the current `page_count`, so a small row still fits in
an existing page. The callback runs a successful insert, an oversized insert that fails
with code 13, and then a small insert whose error it ignores. The test asserts that
neither small row exists afterward. Add a named mutation for this rule: "pass statements
through after a recorded error" makes the late row persist. Record `IOERR`, `NOMEM`, and
`INTERRUPT` as reviewer-read rows. The runner (Task 4) returns on a file's first error
and runs no version or ledger statement after it; state that rule as well.

---

## Major

### D2: the interrupted-create states are wider than "zero bytes, no `-wal`"

**Location:** R-6a5 table (`:346-358`); R-6a6 (`:365-375`); Task 6 order
(`:1271-1274`); Task 1 leaf create (`:612-616`).

**Defect.** R-6a6 treats a crash between the `O_EXCL` pre-create and the open as the only
interrupted create. The write-class DSN carries `_journal_mode=WAL`, so the open itself
writes the file. A child process that opened a pre-created empty file with the
write-class DSN and then exited without closing left this state:

```
db size 4096, -wal absent, -shm absent
mode=ro read: application_id 0, user_version 0, sqlite_schema rows 0
```

Every crash between the open and migration 1's commit leaves this 4096-byte header-only
file. The R-6a5 table has no row for it. Read literally ("`application_id` is not
dubplate's"), it is "not a store" in every mode, so serve refuses its own interrupted
create forever and the fix is a hand deletion. The same gap covers the following states,
none of which R-6a5, Task 1, or Task 6 names:

- a zero-byte file with a `-wal` beside it;
- a leaf that exists without `store.lock` or `db/` (a crash inside the leaf create);
- a leaf created by a second process between this process's stat and its create (two
  starts, or serve against 6b-ii's restore at first boot).

The lock lives inside the leaf, so the leaf create cannot run under the lock.

**Fold.** Add one R-6a5 row: any file whose `application_id` and `user_version` are both
0 and whose schema is empty, whatever its size, is an interrupted create. Serve completes
it under the exclusive lock. Inspect and write report "not a store" with the fix "start
serve to finish creating it." Make the leaf create idempotent (tolerate `EEXIST` on each
directory; open `store.lock` with `O_CREATE` and no `O_EXCL`), then take the exclusive
lock and re-stat before the database pre-create. Name the report for a leaf with no
`store.lock` in inspect and write mode: "not a store, start serve to finish creating it."

### D3: an existing but unopenable database can read as "absent," and doctor exits 0

**Location:** Task 6 inspect (`:1275-1282`); R-6a5 table (`:346-358`); Task 7 exit
table (`:1416-1421`); report state set (`:1262-1264`).

**Defect.** The plan's probe records `mode=ro` of an absent file as `unable to open
database file (14)`. An existing file that the process cannot read returns the same code:

```
chmod 000 on an existing store, then mode=ro read -> unable to open database file (14)
```

The plan never says that "absent" comes only from `os.Stat` returning `ENOENT`. It also
names no report for an open or read error other than code 26. The unnamed cases:

- code 14 on an existing file (EACCES, or a root-owned file inside an owner-owned leaf);
- code 8 (a `-shm` the process cannot write, or a hot rollback journal left in a
  post-restore DELETE-mode file, which `mode=ro` cannot roll back);
- code 10 (I/O) or code 11 (corrupt) on the header read;
- a marked store whose ledger table is missing ("no such table");
- `os.Stat` failing with anything other than `ENOENT`.

An implementer who maps code 14 to absent makes `doctor` and `check` exit 0 with "absent"
over a store that exists. Ruling 2 treats absent as benign.

**Fold.** State it in R-6a5: absent is decided only by `os.Stat` returning
`errors.Is(err, fs.ErrNotExist)`, and no SQLite code maps to absent. Every other stat or
open or read error on a present file reports "refused" (or a new "unreadable" state)
naming the path and the error, and it exits nonzero under Ruling 2. Add one Task 6
criterion: a 0000-mode database in an owner-owned leaf reports the refusal, never absent.

### D4: staging cleanup under a shared lock deletes another holder's live file, and serve never cleans

**Location:** Task 6 write mode (`:1283-1286`), staging cleanup (`:1298-1299`), serve
mode (`:1287-1297`); Task 4 snapshot (`:1000-1010`); Task 1 pre-create (`:680-681`).

**Defect.** The failure has two halves.

1. **The 6b-i rework trap.** Write mode runs staging cleanup after taking the lock
   **shared** (`:1286`). Under the spec, `backup` (write mode, shared) writes its
   `VACUUM INTO` copy into `staging/` (spec `:585-591`), and every `member` verb also
   opens in write mode under a shared lock. A `member add` during a backup therefore
   removes the backup's in-flight staging file. If the removal falls between the
   pre-create and `VACUUM INTO`, SQLite creates the target fresh at 0644. If it falls
   later, the rename fails and the backup fails spuriously. 6a fixes the cleanup's lock
   mode, so 6b-i inherits the race or has to rework the rule.
2. **Serve wedges after a crash mid-snapshot.** Serve mode's procedure lists no staging
   cleanup. The snapshot pre-creates `staging/pre-migrate-vN.db` with `O_EXCL`, which
   "fails if it exists." A crash between the pre-create and the rename, or during the
   integrity-failure removal, leaves that file. Every later serve start then refuses the
   migration, and serve cannot start until someone deletes the file by hand.

**Fold.** Staging cleanup runs only under the **exclusive** lock: in serve mode's step 3,
before the create or the snapshot, and only with no intent record. Write mode does not
clean. At 6b-i, a shared-lock stager names its file by instance id, and recovery removes
only files whose owner lease is free (a 6b-i handoff line in Task 8). Add a Task 6
criterion: a leftover `staging/pre-migrate-vN.db` beside an older store is removed, and
serve then migrates. Change mutation 4's wording to match.

### D5 (OWNER FORK): serve silently creates an empty store when the leaf exists but the database is gone

**Location:** R-6a5 row "absent … created (serve)" (`:350`); Task 6 serve mode
(`:1287-1297`).

**Defect.** Serve decides "create" from the database file's absence alone. Consider a
leaf that already holds `snapshots/pre-migrate-v*.db` or a `pre-restore-*.db`. Its
database can go missing through a hand deletion, a mistaken `mv`, a restore crash outside
the intent's window, or a disk repair. Serve then starts on a fresh empty catalog with no
refusal. `doctor` said "absent, exit 0" the whole time. After 6b-i, the next backup
uploads an empty snapshot as "latest," and ruling 10's retention begins to age out the
good ones. The spec never distinguishes "never created" from "was here and is gone."

**Options.**

- (a) Serve creates only when the leaf is absent, or when the leaf holds no database and
  `snapshots/` is empty (a first create or an interrupted one). Otherwise it refuses,
  naming the newest snapshot and `restore --from-file`. Inspect reports "missing," not
  "absent," when snapshots exist, and exits nonzero.
- (b) Keep the plan's rule: create whenever the file is absent.
- (c) Require an explicit `serve --create-store` for any create after the first.

**Recommendation:** (a). It costs one directory read in serve's step 3 and one inspect
branch, and it closes the only path in 6a that trades a real catalog for an empty one
without a word. (c) adds an attended step to every fresh install.

---

## Minor

### D6: no test pins a shipped migration's bytes

`:989-999`, and spec `:392-393` ("append-only once merged"). An editor's trailing-newline
or whitespace fix to a merged file changes its SHA-256. The ledger check then refuses
every existing store at the first start after the deploy. The refusal fails closed, so no
data is lost, but serve is down until a rollback. The invariant test builds from empty,
so it cannot see the edit. **Fold:** a test table of file name to SHA-256 for every
embedded migration, so an edit fails the gate and a reviewer reads the change.

### D7: transaction control inside a migration file breaks the runner's atomicity

Task 4 (`:989-995`). Measured: `Exec("create table e3(a); commit; create table e4(a)")`
inside a `*sql.Tx` returns nil, both tables persist, and `Rollback` answers "no
transaction is active." A file carrying `COMMIT`, `END`, `ROLLBACK`, `SAVEPOINT`/`RELEASE`,
`VACUUM`, or `ATTACH` commits partial work, and the version and ledger writes that follow
run in autocommit mode. Also measured, and worth quoting in the plan: a multi-statement
`Exec` stops at the first failing statement (a later `create table e2` did not run).
**Fold:** a test over the embedded set refuses those statements. The runner treats a
"no transaction is active" answer from `Commit` after a clean apply as a hard error
naming the file.

### D8: the snapshot's verification open must not rewrite the copy; assert the marker

Task 4 (`:1002-1004`). Measured: a `VACUUM INTO` copy has header bytes 18 and 19 equal
to 1 and 1 (rollback-journal mode) and keeps `application_id` and `user_version`. If the
integrity-check open uses a DSN with `_journal_mode=WAL`, it converts the copy after
verification and leaves `-wal` and `-shm` in `staging/`. **Fold:** the verification open
uses `mode=ro`. Criterion 9's test also asserts the copy's `application_id`, so a
restored pre-migration copy never reads as "not a store."

### D9: an over-range busy timeout silently becomes zero

R-6a3 (`:307-317`); Task 7 (`:1389-1401`). The driver validates `_busy_timeout` with
`strconv.ParseInt(busyTimeout, 10, 64)` (`sqlite.go:307-311`). SQLite then drops values
above int32. Measured: `2147483647` reads back unchanged, while `3600000000` (1000h) and
`99999999999999` read back as `0`, which turns every contended write into an immediate
busy error. **Fold:** the caller refuses a value above 2147483647 ms, naming the key.
That adds a "too large" row beside "0s" and "-1s."

### D10: retention's unnamed states, and the snapshot-deletion channel

R-6a11 (`:442-445`); Task 4 (`:1022-1023`, `:1051-1052`); privileged-fields table
(`:245-254`). The plan names no report for these states:

- `keep` less than 1, which would delete the copy just taken;
- a file in `snapshots/` that does not parse as `pre-migrate-v<N>.db`, such as 6b-ii's
  `pre-restore-<stamp>.db` or an operator's file.

**Fold:** refuse `keep < 1` at the runner. Retention matches only its own exact pattern
and never deletes or errors on anything else. Add a privileged-fields row: "deleting a
file in `snapshots/`: the runner's retention alone at 6a, by version, `pre-migrate`
names only."

### D11: ledger and `user_version` disagreement is unnamed

Task 4 ledger states (`:1038-1040`). The plan names no report for these states:

- a version at or below `user_version` that has no ledger row;
- a ledger row above `user_version`.

Both writes commit in one transaction, so either state means tampering or a bug. The
ledger's version key makes a re-apply fail at the insert and roll back, so the state is
safe, but no report is named. **Fold:** refuse, naming both, in every mode.

### D12: `wal_checkpoint(TRUNCATE)` reports busy in a result row, not as an error

Task 4 (`:1009-1010`). The pragma returns a row whose first column is 1 when it could not
complete, and it returns no Go error. At 6a the pragma runs under the exclusive lock, so
this is theoretical. 6b-i runs the same snapshot function under a shared lock with serve
writing. **Fold:** name the state: busy is a logged warning, never a snapshot failure,
and the function reads the row rather than discarding it.

### D13: `journal_size_limit` is not a WAL bound during a transaction

R-6a14 (`:474-477`); R-6a11 free space (`:448-452`). The plan's own quote says the pragma
truncates the WAL "after a checkpoint." A migration that rewrites a large table grows the
WAL toward the table's size within its one transaction. The free-space check can pass,
and the migration then hits `SQLITE_FULL` with the new snapshot occupying the space. The
rollback is clean, so this is a wording and expectation defect only. **Fold:** rename the
constant "the post-checkpoint WAL size." State that the check covers the snapshot, and
that a mid-migration disk-full is the named disk-full refusal with the store at the prior
version.

### D14: a serve create run as root makes a root-owned leaf

Task 1 owner check (`:642-650`); Task 6 serve create. The owner check guards an existing
leaf only. A first `serve` run as root (a hand run, or a container default) creates a
root-owned leaf. The service user is then refused with the fix `runuser -u root --`,
which is the wrong fix. **Fold:** the create requires the effective uid to equal the
state directory's owner, and otherwise refuses naming that owner.

### D15: the filesystem probe and owner check see the leaf, not `db/`

Task 1 (`:651-661`). A `db/` that is a symlink or a mount onto NFS passes the probe run on
the leaf. **Fold:** probe the directory that holds the database, and refuse a symlinked
leaf or `db/` (`os.Lstat`), naming the path.

### D16: classifying through a write-class DSN writes to a foreign file

Task 6 order (`:1271-1274`) says the file is classified "before any connection opens,"
but reading `application_id` needs a connection, and the plan does not say which DSN.
Opening a DELETE-mode foreign database, or a newer post-restore store, with the
write-class DSN runs `PRAGMA journal_mode=WAL` first and rewrites its header before the
refusal. **Fold:** every mode classifies through the inspect DSN (`mode=ro`). The
write-class and serve DSNs open only after the file classifies as a dubplate store.

### D17: the privileged-fields table misses the store's creation race and the leaf's mode

`:245-254`. "The store's creation: the serve mode alone" holds, but 6b-ii's restore also
creates a database file from a snapshot (spec `:187-188`). The row should name restore as
a later channel, so the D2 race fold has a stated counterpart. Leaf and subdirectory modes
(0700) name only the leaf. **Fold:** state 0700 for `db/`, `snapshots/`, and `staging/`
too, and add restore (6b-ii) to the creation row as a future channel.

---

## Checked and sound

These were checked and need no fold.

- **Exported surface (R-6a2).** One inspect function, its options, and its report, each
  read by `doctor` and `check`. Nothing is exported ahead of a caller.
- **2.0 hooks.** No 6a schema choice forecloses hooks 1 to 15. The composite
  `(album_id, collection_id)` key makes a later album move between collections need
  `PRAGMA defer_foreign_keys` inside the transaction, which the write method's exec
  permits, so no rework follows. The typed collection columns and the surrogate id match
  hooks 5, 8, and 15.
- **Upgrade and downgrade.** The newer-store refusal in every mode, the per-file
  transaction, and the conversion procedure's re-read cover a binary older than the
  store and a restore that wins the conversion gap.
- **The four planning-miss checks.** Quoted behaviors are present everywhere except the
  items above: multi-statement `Exec` (D7), code 14's ambiguity (D3), and the
  busy-timeout range (D9). The unnamed states are D2, D3, D10, D11, and D12. The
  privileged-channel gaps are D10 and D17. Fixture states are named throughout. D1's
  criterion carries its own fixture, since a test that swallows no error could pass
  vacuously.
