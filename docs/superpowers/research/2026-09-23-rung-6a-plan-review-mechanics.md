# Rung 6a plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md` at `11f11c2` (revision 1).
Lens: does every mechanism behave as stated. Evidence comes from module source under
`$(go env GOMODCACHE)/modernc.org/sqlite@v1.59.0`, `go doc` on go1.27.1, and new probes under
`.../scratchpad/6a-review-probes/` (`main.go`, `lk/`, `race/`, `sh/`, `vi/`), run 2026-09-23 against
modernc.org/sqlite v1.59.0 on this workstation. Line numbers are plan lines unless marked otherwise.

**Counts: 0 blockers, 7 majors, 9 minors.** The plan's order is buildable as given, and the
worktree territory against 5c-i holds. The drafter's quotes that I re-read are accurate: the DSN
key order and validation (`sqlite.go:281-470`), `newTx`'s `begin` rule (`tx.go:20-30`),
`extendedResultCodes(true)` (`conn.go:113`), the OFD text (`ofd.go:31-39`, CHANGELOG), `go doc`
for `crypto/rand.Read`, `DB.BeginTx`, `DB.Conn`, `os/user.LookupId`, `syscall.Statfs_t` (it carries
`Frsize`), and the `magic.h` constants. The majors are states the plan does not name and proofs that
cannot fire as written.

## Majors

### M1 (major): a `file:` DSN built by concatenation opens the wrong file

Location: `:748-749` ("The DSN is a `file:` URI"), and every DSN in `:761-766`, plus the
snapshot's integrity open (`:1003`).

Defect: nothing in the plan escapes the path. SQLite parses the URI, so `#` starts a fragment, `?`
starts the query, and `%XX` is decoded. modernc also splits on the first `?` (`conn.go:62`,
`strings.IndexRune(dsn, '?')`). Probe C built `"file:" + dir + "/s.db" + ws`:

```
C dir "sp ace":  exec err=<nil>; intended file exists=true
C dir "hash#01": exec err=<nil>; intended file exists=false; ... entries include "hash"
C dir "pct%41x": exec err=unable to open database file (14); intended file exists=false
C dir "q?mark":  exec err=<nil>; intended file exists=false; ... entries include "q"
```

With `#` or `?` in the path, SQLite silently created a database at a truncated path. That breaks
"inspect creates nothing" and the universality goal. It also reaches the gate: `t.TempDir()` keeps
`#` and `%` in its directory name (`testing.go:1610`, `const allowed = "!#$%&()+,-.=@^_{}~ "`), and
Go names a duplicate subtest `name#01`.

Fold: one unexported DSN builder escapes the path as a URI path (`(&url.URL{Scheme: "file", Path:
p}).String()` or an equivalent that encodes `%`, `?`, and `#`), and every class uses it. Add a Task 2
row that opens under a directory named `a#b%41?c` and asserts the file lands at that exact path, with
nothing created at `a`. Add the mutation "concatenate the path", which that row fires.

### M2 (major): the likeliest interrupted create is a 4096-byte unmarked file, and no state names it

Location: R-6a5's table (`:346-358`) and R-6a6 (`:365-375`).

Defect: R-6a6 says a crash "between the two" (the pre-create and the open) leaves a zero-byte file.
But the open itself writes page 1. Probe D pre-created an empty 0600 file and opened the write DSN
without creating a table:

```
D journal_mode after open: wal
D after open (conn live): [ 4096 -rw-------] [-wal absent] ...
D ro: app_id 0 uv 0 schema 0 jm wal
D2 (migration-1-shaped tx rolled back, as a crash would): [ 4096 -rw-------] [-wal absent] [-shm absent]
```

So a crash at any point from the first connection open through migration 1's commit leaves a
non-zero file with `application_id` 0, `user_version` 0, and an empty schema. The table has no row
for that state. The zero-byte row is decided by size, and the "any object" row needs a schema object.
The spec (`:377`) says "A store is non-empty when `user_version > 0`". An implementer who reads the
table as "unmarked means not a store" leaves `serve` refusing on every start after the crash, with
manual deletion the only fix.

Fold: add a row: "a SQLite database with `application_id` 0, `user_version` 0, and an empty schema
(any size)" is an interrupted create. Serve mode completes it under the exclusive lock. Inspect and
write mode report "not a store, naming 'start serve to finish creating it'". Restate R-6a6's premise:
the zero-byte state covers only the pre-create-to-open window. Add a Task 6 row built the way D2 is
built (open the write DSN, begin, roll back, close), so the file is 4096 bytes when the row runs.

### M3 (major): serve mode never clears `staging/`, so a crash during a snapshot blocks every later migration

Location: Task 6 serve outcome (`:1287-1297`), staging cleanup (`:1283-1286`, `:1298-1299`), and
Task 4's snapshot pre-create (`:1001-1002`) with its states (`:1045-1048`).

Defect: the snapshot pre-creates `staging/pre-migrate-vN.db` with `O_CREATE|O_EXCL`, and the "target
exists" row refuses (`:681`). Only write mode runs staging cleanup (`:1286`). Serve mode's procedure
never runs it. A crash or kill during `VACUUM INTO` or `integrity_check` leaves the staged file in
place. Task 4's "copy fails `integrity_check` → copy removed" row covers a failure the process
survives, not a crash. On the next start, serve mode's pre-create meets the leftover and refuses. At
6a no production write-mode verb exists to clear it, so the box cannot migrate until someone deletes
the file by hand. The spec already rules this case (`:235`): "Every writing open checks for a
`restore-intent` record first. Only then does it remove staging leftovers." Serve mode is a writing
open.

Fold: serve mode runs staging cleanup in step 3, holding the exclusive lock, after the intent check
and before create or migrate. Add Task 4 or Task 6 rows: "a staged `pre-migrate-vN.db` left by an
earlier run → removed, then the snapshot proceeds", and "a staged file with an intent present →
untouched, and serve refuses per R-6a7".

### M4 (major): the conversion proof in criterion 14 cannot be ordered as written, and loses the race about 299 times in 300

Location: Task 6 criterion 14 (`:1321-1332`) and mutations 1 and 2 (`:1352-1355`).

Defect: the plan has a blocking exclusive `flock` queued before serve mode releases its shared lock,
then granted in the gap. Two problems follow. First, the test has no event that tells it serve mode
holds the shared lock, so it cannot queue the waiter at the right time without a sleep. Second, the
kernel does not hand the lock to a woken waiter; the waiter must re-run the lock attempt. Serve
mode's own `LOCK_EX|LOCK_NB`, issued right after its release, usually wins. Probe `race/` queued the
waiter (confirmed by its `->` line in `/proc/locks`), then released shared and tried an immediate
non-blocking exclusive:

```
serve's immediate LOCK_NB won 299, blocked waiter won 1 of 300
```

So the row almost never exercises the gap. When serve mode wins, it migrates, and mutations 1 and 2
survive.

Fold: order the row through the `*slog.Logger` parameter serve mode already takes (`:1290`). The test
holds its own shared lock S on descriptor A before calling serve mode. Serve mode reads the older
version, closes, releases, and its exclusive poll fails against S. Serve mode then logs its "waiting
for exclusive" line, and a test handler signals a channel on it. At that point serve mode holds no
lock and no connection, so the test converts S to exclusive on descriptor A with
`LOCK_EX|LOCK_NB`. That conversion succeeds deterministically. Holding it, the test renames and
replaces the file, checks `/proc/self/fd`, and releases. This keeps the spec's property (a
competitor wins the gap) with no sleep and no race. The fixture state must also say that every other
SQLite connection to that file in the test process is closed. SQLite's unix VFS defers closing a
descriptor while another connection on the inode holds locks (`sqlite_g_000000000001ffff.go:12569`:
"If there are outstanding locks, do not actually close the file just yet ... add the file descriptor
to pInode->pUnused list"), and a deferred descriptor would fail the `/proc/self/fd` row for the
wrong reason.

### M5 (major): `ctx` does not bound SQLite's busy wait

Location: Task 3's states table row "Deadline | expires waiting on another process's write lock |
the busy error at the busy timeout, or the context's error if sooner" (`:913`), and the outcome at
`:885-892`.

Defect: the "context's error if sooner" half is false. Probe A held an `IMMEDIATE` transaction on
one handle, then called `BeginTx` on a second handle with `_busy_timeout=3000` and a 300 ms context:

```
A waiter BeginTx after 3.01s: err=database is locked (5) (SQLITE_BUSY)
```

The context expired at 300 ms and the call still returned at the busy timeout. `newTx` runs `begin
immediate` through the busy handler, which does not check the interrupt the context triggers. The
spec says only that the in-process wait is bounded by the context (`spec:331-332`), which is true. At
the 15 s default, a caller with a 2 s deadline waits 15 s. A test written to the plan's row either
fails or asserts nothing.

Fold: correct the row to "the busy error at the busy timeout; the context does not shorten SQLite's
busy wait (measured)", and pin that with a test. Choosing between that and a per-call busy timeout is
OWNER FORK A (below).

### M6 (major): criterion 8's fixture fails before the version write, so mutation 1 cannot fire

Location: Task 4 criterion 8 (`:1056-1061`), the runner's order (`:989-993`), and mutation 1
(`:1083-1084`).

Defect: the runner "reads `user_version`, applies the file, sets `PRAGMA user_version = N` ..., and
inserts the ledger row". The fixture is "a synthetic file that sets a table, then fails on a later
statement in the same file". That failure happens inside the file, before the runner's version write.
Mutation 1 moves the version write into a separate transaction. Under that mutation, the failed file
still never reaches the version write, so the row passes and the mutation survives. The spec
criterion says "failing **after** its version write" (`spec:931`). The only step after the version
write is the ledger insert.

Fold: make the failure land after the version write. For example, the synthetic file creates a
`BEFORE INSERT` trigger on the ledger table that runs `SELECT RAISE(ABORT, 'injected')`, so the
runner's own ledger insert fails after `user_version = N` has executed. Assert that `user_version`,
`sqlite_schema` (the trigger and table gone), and the ledger are unchanged. Also keep the in-file
failure row as its own case.

### M7 (major): no mechanism exists for the busy timeout's positive-value refusal

Location: R-6a3 (`:314-317`, "a load error naming the key, raised by the caller"), and Task 7
(`:1399-1401`, `:1419`, `:1436-1437`).

Defect: thresholds are not parsed at `config.Load`. The package comment says a threshold "resolves
at its point of use rather than at Load" (`internal/config/config.go:48-53`). `doctor` renders a
threshold row through `thresholdStatus` (`cmd/dubplate-server/doctor.go:212-231`). That function
prints `d.String()` for any duration that parses, so `0s` and `-1s` render as `0s` and `-1s`, with
no reason attached. The only existing bounds channel is an identity source's `Validate` hook feeding
`rowStatuses`. rung 5's `durationFloor` uses it through `validateMagicLinkSource`
(`auth.go:72-76`, `:225-`). That hook sits in `auth.go`, a 5c-i file, and belongs to another
concern. So "the threshold row reports the reason" has no channel inside Task 7's territory. An
implementer will improvise, touch `auth.go` or `internal/config` (both walled), or stop. Separately,
`_busy_timeout` takes integer milliseconds, so a positive duration under 1 ms (for example `500us`)
converts to 0, which means no busy wait at all.

Fold: say it outright. The threshold row shows the parsed value, as every duration threshold does
today. The store row carries the refusal: "not checked: DUBPLATE_STORE_BUSY_TIMEOUT must be at
least 1ms". `check` fails naming the key before inspect runs. The floor is 1 ms, a named constant.
Remove "a load error" from R-6a3. Correct Task 7's criterion to assert the store row's text rather
than the threshold row's.

## Minors

### m1: the maintenance class has no `_txlock`, so its "BEGIN IMMEDIATE" is really `BEGIN`

Location: `:765` and `:1024-1032`. `newTx` issues plain `begin` when `beginMode` is empty
(`tx.go:22-25`). Probe E: a maintenance-DSN transaction read, another connection committed, and the
maintenance write then failed at once with `database is locked (517)` (`SQLITE_BUSY_SNAPSHOT`, no
busy wait). Fold: add `_txlock=immediate` to the maintenance DSN, which `VACUUM INTO` does not
affect because it runs outside `BeginTx`. Alternatively, state that the rebuild issues `BEGIN
IMMEDIATE` itself on a pinned `*sql.Conn`. Extend criterion 4's readback to the begin mode.

### m2: `/proc/locks` matching misses on btrfs, and `->` waiter lines shift the fields

Location: `:628-641`, `:694-698`, `:711`. Probe `lk/`, one exclusive `flock` per directory:

```
/tmp (tmpfs):              stat maj:min 0:3c  /proc/locks 00:3c   match
/var/home/glw907 (btrfs):  stat maj:min 0:3a  /proc/locks 00:25   no match
```

btrfs reports a per-subvolume device to `stat` but the superblock's device in `/proc/locks`, so on
btrfs the holder list is always empty. A queued blocking waiter prints as `96: -> FLOCK ADVISORY
WRITE 63012 00:3c:...`, which moves the pid to field 6. The stdlib has no `major`/`minor` helper,
and the plan declines `x/sys/unix`, so the `dev_t` split (glibc's encoding) is hand-written and not
quoted. Fold: quote the encoding, skip `->` lines (they are waiters, not holders), and state the
fixture for the own-pid row: `t.TempDir()` on `$TMPDIR` or `/tmp`, reported by `statfs` type. Keep
"no pids" as the named btrfs outcome, and put the btrfs gap in the HISTORY coverage list. The box's
filesystem decides whether this matters in production.

### m3: the inspect report's closed state set lacks three states the plan names or implies

Location: `:1262-1264`, against `:358` and `:651`. R-6a5 says inspect reports a ledger mismatch "as
its own state", but the closed set has no such state. There is also no state for a SQLite error
other than 26 (for example 11 `SQLITE_CORRUPT`, 10 I/O, or 14 on a permission failure), or for a
state directory whose parent `statfs` fails with `ENOENT`. The same gap covers a leaf that exists
without `store.lock` or `db/` (a crash during the leaf create). Fold: name each state's report and
its Ruling 2 exit, and make the leaf create idempotent (`MkdirAll`, and `store.lock` opened with
`O_CREATE` only by serve). Short files behave as the plan expects: probe `sh/` returned `file is not
a database (26)` for a 5-byte file, 99 zero bytes, and a bare "SQLite format 3" header.

### m4: the nested-write mark must reach the callback's `ctx`

Location: `:893-897`. The mark catches a nested call only if that call uses the marked `ctx`. A
callback that closes over the caller's outer `ctx` waits for the one connection until its deadline.
Fold: give the callback the signature `func(ctx context.Context, tx *txHandle) error` with the
marked `ctx`. Document that the closure form is not caught. Make the criterion's nested call use the
passed `ctx`.

### m5: `VACUUM INTO` durability and target binding are unquoted

Location: `:1000-1010`. Two supporting facts need quoting. The copy is synced because `VACUUM INTO`
takes the source connection's safety level: SQLite source, "For a VACUUM INTO, the pager-flags are
set to the same values as they are for the database being vacuumed" (`lib/sqlite_freebsd_386.go:24556`,
the same amalgamation). So the maintenance DSN's `_synchronous=FULL` is load-bearing; name a
mutation that drops it, or `Sync` the copy by hand. The target should be bound: probe `vi/` ran
`VACUUM INTO ?` with the target `it's copy.db` and succeeded, where string concatenation breaks on
the `'`. For the 6b-ii handoff: the copy is a rollback-journal database. Probe B read header bytes
18 and 19 as 1,1 and `journal_mode` as `delete`, so a restore that places it must set WAL.

### m6: Task 2's mutation 4 has a firing row

Location: `:841-844`. Probe E produces an extended `SQLITE_BUSY` (517) from a deferred read-then-write
on a second handle. Fold: add that row, so the full-code mutation fires instead of becoming a
reviewer-read row.

### m7: retention must match only `pre-migrate-v*.db`

Location: `:1022-1023`, `:1052`. `snapshots/` also holds pre-restore copies (spec `:227`). Fold:
retention selects only by the `pre-migrate-v<N>.db` pattern. Add a row with a planted non-matching
file that survives retention.

### m8: `journal_size_limit` on "every class" disagrees with the DSN table

Location: R-6a14 (`:474-477`) against the inspect row (`:766`). Fold: say "every writing class".

### m9: other shared-network filesystems

Location: `:651-661`. `magic.h` also carries `CEPH_SUPER_MAGIC 0x00c36400` and `AFS_SUPER_MAGIC
0x5346414F`, both multi-host filesystems where WAL's one-host rule fails the same way. Fold: add both
to the refused list, as the plan already does for SMB, and record them in erratum 6.

## Planning-miss audit (the four items)

1. **Quoted, never recalled.** Mostly met. Unquoted: `dev_t` decoding (m2), `VACUUM INTO`'s sync
   inheritance (m5), SQLite's busy-handler/`ctx` interaction (M5), and `config`'s point-of-use
   threshold rule (M7).
2. **Every state names its report.** Missed: the 4096-byte unmarked create (M2), a stale staging
   file under serve (M3), and the non-26 SQLite errors, the ledger-mismatch state, and the partial
   leaf (m3).
3. **Privileged fields name every channel.** The table at `:245-254` is complete for 6a's fields.
   One addition: `staging/` contents are written by the snapshot and removed by write-mode cleanup
   and (after M3) serve-mode cleanup, and by nothing else.
4. **Proofs name their fixture state.** Missed: criterion 8's failure point (M6), criterion 14's
   ordering and the other-connection precondition (M4), and the own-pid row's filesystem (m2).

## OWNER FORK A: a caller's deadline shorter than the busy timeout (from M5)

- **Option 1 (recommended): accept it and document it.** The busy timeout bounds cross-process
  waits, and a caller's deadline bounds only the in-process wait. The writers at group scale are one
  `serve` and short CLI verbs, so a 15 s worst case is rare and visible. No new mechanism.
- **Option 2: per-call busy timeout.** The write method pins a `*sql.Conn` and sets `PRAGMA
  busy_timeout` to min(time left on the deadline, the configured value) before `BEGIN IMMEDIATE`.
  This honors deadlines, but it adds a per-write statement and turns the pool into a hand-managed
  connection.
