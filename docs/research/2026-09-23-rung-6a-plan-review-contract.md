# Rung 6a Plan Review: Contract and Criteria

Target: `docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md` at `11f11c2`
(revision 1). Source of truth: the rung 6 spec, revision 2.2, section "The pass split", 6a
(`:911-949`), and the repo `CLAUDE.md`. Lens: contract and criteria only. Line numbers below are
the plan's lines at `11f11c2`.

**Verdict.** Every one of the spec's sixteen 6a criteria maps to a task criterion, and the exit
gate's map (`:1504-1529`) is complete. The plan's additions beyond the spec are declared as
resolutions, and two of them are raised as rulings. The quoting discipline is strong: flock,
`/proc/locks`, the magic numbers, the DSN order, `VACUUM INTO`, the pragma rollback, and
`crypto/rand.Read` are all quoted or probed. The gaps are four criteria whose fixture state as
written cannot fail under the mutation they claim to catch, and two unnamed store states that
real crashes produce.

Counts: 0 blockers, 6 majors, 7 minors.

---

## Majors

### C1 (major): criterion 8's fixture fails before the version write, so the promise is vacuous

**Location:** plan `:1057-1062` (Task 4 criterion), against `:989-996` (the runner's order) and
mutation 1 at `:1084-1085`.

**Defect.** Spec criterion 8 says "an injected migration failing **after** its version write." The
runner's order (`:989-994`) is: read `user_version`, apply the file, set `PRAGMA user_version = N`,
insert the ledger row. The plan's fixture is "a synthetic file that sets a table, then fails on a
later statement in the same file". That fails inside the file, **before** the runner's version
write, so the named state is never reached. Mutation 1 ("set `user_version` in a separate
transaction") does not fire when that separate transaction would run after the file, because the
file never gets that far. The atomicity promise at the data-integrity core is then unproved.

**Fold.** Two fixture rows. (a) A synthetic file that runs its own `PRAGMA user_version = <N>`
first and then fails on a later statement. (b) A failure injected after the runner's own version
write: for example, a synthetic set whose earlier file plants a ledger row at version N, so the
runner's ledger insert for N hits the primary key. Each row asserts `user_version`, the
`sqlite_schema` rows, and the ledger rows equal their pre-run values. Name which row mutation 1
fires on.

### C2 (major): criterion 15's "bytes unchanged" passes on a checkpointed store under the mutation it claims to catch

**Location:** plan `:1334-1341` (Task 6, criterion 15 in-package half), mutation 6 at `:1362-1363`,
and R-6a4 at `:320-323`.

**Defect.** R-6a4 hashes the database file only after every handle in the test process has closed.
The last close of a WAL database checkpoints and removes `-wal`, so the fixture store has an empty
WAL. Opening that store read-write and closing it leaves the main file's bytes unchanged. So the
"bytes unchanged" assertion passes even when inspect opens without `mode=ro`, and mutation 6 cannot
fire through this row. The spec's reason for the invariant (a read-only WAL open, `wal.html`
section 5) only matters when the WAL holds frames not yet in the main file.

**Fold.** Name the fixture state as "a hot WAL". The store's last writer is a re-executed test
binary killed before close (the same real-process-death technique the spec's restore step hook
uses), so `-wal` holds committed frames that are not yet in the main file. Hash the main file, run
inspect, then hash again. Assert the report reads the WAL's committed rows, which proves inspect
really read through the WAL. Apply the same state to the older and newer rows. State that mutation
6 fires on this row.

### C3 (major): the classification table has no row for serve's interrupted create after the WAL switch

**Location:** plan R-6a5 table `:340-350`, R-6a6 `:356-367`, and Task 6 `:1288-1301`.

**Defect.** Serve's create pre-creates a zero-byte file, then opens it through the write-pool DSN,
whose `_journal_mode=WAL` writes the database header, and then runs migration 1. A crash between
the header write and migration 1's commit leaves a file with nonzero size, `application_id` 0,
`user_version` 0, and an empty schema. The table is ambiguous for this file. Row 4 ("`application_id`
is not dubplate's") classifies it as not a store, and serve then refuses it permanently. Row 5
implies that zero, zero, and an **empty** schema is some other state, but it never says which.
R-6a6 rescues only the zero-byte case. So a crash during first start can leave a store that
refuses to start until someone deletes a file by hand, and no row names that report.

**Fold.** Add a row: a file with nonzero size, `application_id` 0, `user_version` 0, and zero
schema objects is an interrupted create. Serve completes it under the exclusive lock. Inspect and
write mode report it as not a store, naming "start serve to finish creating it". Make row 4
explicit that any nonzero foreign id, or 0 with a nonempty schema, is not a store. Add a fixture
for each: set the header with `PRAGMA journal_mode=WAL` alone, and use a foreign database with
`user_version` set but no `application_id`, a common shape. Also name the report for a zero-byte
file with a `-wal` beside it: refuse it and never truncate.

### C4 (major): a partly created leaf is unnamed in every mode

**Location:** plan Task 1 states table `:627-643`, Task 6 order `:1276-1279`, and Task 6 criterion 13
`:1314-1321`.

**Defect.** The states tables cover a leaf that is absent, present with the wrong owner, or not a
0700 directory. They cover no leaf that is present but missing `store.lock`, `db/`, or the database
file. Serve's create makes the leaf and its contents in several steps, so a crash can leave any of
these, and an operator's `rm` can too. Two contract questions have no answer. (a) Inspect takes the
lock before it classifies (`:1277`). If `store.lock` is absent, does inspect create it, which would
be a write from `doctor`, or report something? (b) Criterion 13's "creates nothing" fixture uses a
missing leaf only (`:1316`). The more dangerous absent state is a present leaf with no database
file. There, the only guard is the pre-open `os.Stat`, because the driver opens every class with
`SQLITE_OPEN_CREATE` (`conn.go:102`, v1.59.0, quoted by the plan at `:726-728`). A mutation that
skips classification when the leaf exists would create a database, and no test fires.

**Fold.** Add rows to Task 6's table for a leaf that lacks `store.lock`, lacks `db/`, or lacks the
database file, with a report per mode. The recommended rows: inspect reports absent and creates
nothing, lock file included; write mode refuses naming "start serve first"; serve completes the
leaf under the exclusive lock. Extend criterion 13's fixture to cover a present leaf with no
database file, asserting the directory listing is unchanged. See C9 for defense in depth.

### C5 (major): criterion 14's fixture names no ordering, and the re-read cannot fail when the versions match

**Location:** plan `:1322-1333` (Task 6, criterion 14) and mutation 2 at `:1355-1356`.

**Defect.** Three gaps each let the test pass vacuously or flake.
(a) "Waits and does not migrate (the version is unchanged while it waits)" gives no event that
proves serve has reached step 2's wait. A version check that runs before serve reaches step 2
passes trivially.
(b) "A second descriptor blocked in a **blocking** exclusive `flock` before serve mode releases its
shared lock" also has no ordering. A goroutine cannot show that it is parked in the kernel, and the
plan's own constraint (`:541-543`) forbids a sleep. Nothing pauses serve between step 1 and step 2.
The spec allows only three seams, and none of them sits here.
(c) The criterion does not say which version the replacement store holds. If the replacement is at
the stale version, mutation 2 ("skip the re-read") changes nothing observable.
The check "no descriptor points at the renamed file" also relies on `/proc/self/fd` semantics that
the plan never quotes. After a rename, `readlink` returns the new path, so comparing against the
original path passes under the mutation. That breaks the planning-miss quoting rule.

**Fold.** (a) Order on serve's holder log line: the `*slog.Logger` parameter feeds a handler that
signals a channel, and the version is read after that signal. (b) Build the gap
deterministically. The test holds shared on descriptor A. Serve runs step 1, releases, and polls
exclusive; the log signal shows it is waiting behind A. The test then issues a blocking exclusive
on descriptor B. It confirms B is pending by the `->` waiter line in `/proc/locks`, quoted from
`proc_locks(5)`. Then it releases A. Report whether B or serve's 50 ms poll wins. If serve wins,
the test retries or reports it as flaky rather than passing. (c) The replacement store sits at a
different version, for example current where the original was older, and the test asserts serve
applied no migration to it. For the descriptor check, compare device and inode from `os.Stat` of
each `/proc/self/fd/N`, which opens nothing, against the renamed file's stat. Quote `proc_pid_fd(5)`
or probe it.

### C6 (major): criterion 12 does not assert which index refused the duplicate

**Location:** plan `:1188-1191` (Task 5, criterion 12) and mutation 5 at `:1216-1217`.

**Defect.** `users` carries three unique columns: `id`, the address, and `streamer_username`, which
uses `NOCASE`. The fixture inserts a row "directly in the test" and then a duplicate. If the
duplicate also repeats the `streamer_username` or the `id`, the insert fails for the wrong reason.
The tombstoned case then passes under mutation 5, the partial index on live rows. Ruling 9 is an
owner ruling, and this is its only proof.

**Fold.** The duplicate row differs from the existing one in every column except the address. The
test asserts extended code 2067 (`SQLITE_CONSTRAINT_UNIQUE`) and that the message names the address
column (SQLite's form: `UNIQUE constraint failed: users.<column>`). Quote that message from a probe.
Apply the same rule to the two 787 rows at `:1199-1200`: the parent album and the collection each
exist, so the only failure is the one the row names.

---

## Minors

### C7 (minor): the ctx bound on another process's write lock is recalled, not quoted

**Location:** plan `:914` (Task 3 states row "the busy error at the busy timeout, or the context's
error if sooner").

**Defect.** The row assumes that cancelling a context interrupts SQLite's busy-handler wait. The
module does call `sqlite3_interrupt` when a context is cancelled (`all_test.go:4589-4680`,
v1.59.0). That test also shows the driver **discards** a file-backed connection after an interrupt,
so the write pool's one connection is replaced. Neither fact is quoted, and nobody measured whether
an interrupt cuts short a busy-handler sleep. That breaks the planning-miss rule on quoting.

**Fold.** Probe it: another handle holds `IMMEDIATE`, the context deadline is 100 ms, and the busy
timeout is 2 s. Quote the result, or drop "or the context's error if sooner" from the row. Record
the connection discard, so a later task does not assume that per-connection state such as
`max_page_count` survives.

### C8 (minor): serve's shared wait in step 1 has no bound and no report

**Location:** plan `:1300` (Task 6 table, "Lock | exclusive held by another | ... | waits for its
shared lock").

**Defect.** Every other wait names its bound and what it reports on expiry. This one names neither.
A restore that holds the exclusive lock indefinitely would hang serve silently. The spec's
procedure (`:344-358`) also leaves step 1 unbounded, so this is a gap the spec shares.

**Fold.** Bound step 1 by the migration wait and report the holders on expiry, the same as step 2.
Or state that the wait is unbounded and logs the holders at the interval, so 6d inherits the
decision in writing.

### C9 (minor): the privileged-field table leaves out SQLite's own create channel

**Location:** plan `:584-587` ("The store's creation" row).

**Defect.** The row names serve mode as the only channel. But every connection class the driver
opens carries `SQLITE_OPEN_CREATE` (`conn.go:102`, v1.59.0), and so does the snapshot verification
open at `:996-1000`. Each of those is a channel that can create a file at a wrong or absent path.
Only the classification step refuses them. The same gap weakens the snapshot verification: an empty
file passes `integrity_check` with `ok`.

**Fold.** List the driver's create flag as a channel in the table. Put SQLite's URI `mode=rw`
(quoted from `uri.html`: opened for read-write, never created) on the write, read, and maintenance
DSNs for an existing store, and open the snapshot verification copy with `mode=ro`. Keep the
pre-open `os.Stat`. Have the snapshot verification also assert `user_version` N and the marker.
Serve's create keeps its own DSN without `mode=rw`.

### C10 (minor): R-6a6 changes a ratified criterion, and the plan defers it to the close

**Location:** plan R-6a6 `:356-367` and Task 8 erratum (2) at `:1459-1460`.

**Defect.** Spec criterion 13 says a zero-byte file is refused as not a store, with no mode named.
The plan has serve complete that file instead. The change is sound, but it is a spec deviation
decided before execution, and the plan queues it as an erratum drafted after execution. R-6a3 does
the same to the fold record's `main.go` row.

**Fold.** Name R-6a3 and R-6a6 in the approval message as spec deviations that take effect on
approval. The close still drafts the erratum text. This is not an owner fork: it is a record of what
the conductor approves.

### C11 (minor): two assertions are vacuous by construction and are not declared as gaps

**Location:** plan `:1337-1338` ("`-wal` or `-shm` owned by the leaf owner") and Task 8's gap list
at `:1450-1455`.

**Defect.** In a test, every file is owned by the test's uid, so the ownership assertion cannot
fail. The inspect connection class (`:747`) also has no settings readback. Criterion 4 names the
three classes, so the plan does not owe one, but the `mode=ro` and `_query_only=1` it relies on are
unproved except by C2's row.

**Fold.** Add "leaf-owner ownership of `-wal` and `-shm`" to Task 8's declared coverage gaps as a
reviewer-read row. Optionally, add an inspect-class readback of `query_only`.

### C12 (minor): tests that set the umask are process-wide

**Location:** plan `:652-654` and `:1342-1344`.

**Defect.** `syscall.Umask` changes the whole process. A parallel test that creates files while the
umask is changed sees a different mode, and the 0600 rows could pass or fail for the wrong reason.

**Fold.** State that the umask tests never call `t.Parallel()`, and that no test in the package runs
in parallel with them. The simplest form is no `t.Parallel()` anywhere in the files that change the
umask, plus a line in the criterion.

### C13 (minor): the ledger states miss a row for a missing ledger entry

**Location:** plan Task 4 states table `:1027-1043`.

**Defect.** The table covers a row that disagrees and a row with no binary file. It does not cover
a ledger that has fewer rows than `user_version` (for example, version 3 with rows 1 and 2 only). A
tampered or hand-restored store can produce that.

**Fold.** Add the row "a version at or below `user_version` has no ledger row: refused, naming the
version" to the table. One fixture proves it.

---

## Areas that are sound

- **Spec-to-plan map.** All sixteen criteria are placed. Criterion 3 is proved at both the
  connection and the write-method level. Criteria 9 and 15 are split across tasks with the halves
  named. The exit gate adds criterion 11 ("no production create") and states how it is proved.
- **Promises beyond the spec** are each declared as a resolution with a reason. The ones that change
  behavior are the `application_id` marker (R-6a5), the path CHECK constraints (R-6a9), serve's
  refusal of any intent record at 6a (R-6a7), SMB added to the refused list, and the `doctor` and
  `check` exit rule (Ruling 2). The exit rule and the ceiling are correctly raised as owner rulings
  and are not decided inside the plan.
- **Quoting.** The flock semantics, `/proc/locks` fields, the magic numbers, the driver's DSN order
  and its lack of `_pragma` validation, `_txlock` scoped to read-write transactions, the pragma
  rollback, the `VACUUM INTO` target rules, and `crypto/rand.Read` are all quoted or probed with
  their values.
- **Fixture states** are named for criteria 1, 2, 4 (forced fresh connections), 6 (a successful
  first statement, so a plain `*sql.Tx` fails the row), 7, 9, 10, and 16.
- **The privileged-field table** is complete for 6a's fields apart from C9. Recording the fields
  that have "none at 6a" as a writer is the right defense against adding a writer by accident.
- **The arithmetic** holds: 8 tasks × 0.533M plus 0.35M comes to 4.62M; the flag is 5.12M at 80% of
  6.4M; and the deliverables total 31.
