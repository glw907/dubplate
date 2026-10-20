# Draft: `docs/HISTORY.md` entry for rung 6, pass 6a

Paste-ready for the conductor's fold. Placement: newest-first, above the
5c-i entry. Sections "The errata this pass owes" and "The 6b-i handoff"
live inline below as their own headings inside this same draft, per the
dispatch's "say which"; they are not separate files.

---

## Rung 6, pass 6a (the store foundation): closed 2026-09-24, lane 2

Executed against
`docs/plans/2026-09-23-rung-6a-store-foundation.md`
(revision 2.2, ratified by Geoff 2026-09-24 with all four rulings yes)
in the lane 2 worktree (`.claude/worktrees/lane-2-6a`, branch
`lane-2-6a`), from `08befea` (the Go tree unchanged since). Task 0's
staleness read found no blockers. The plan's full review history, fold
rulings, and finding dispositions live in
`docs/research/2026-09-23-rung-6a-plan-fold-record.md`.

Tasks 1 through 7 ran across two segments: A (Tasks 1-3, closed
`2d863cb`) and B (Tasks 4-7, closed `dedc1b7`), followed by one tidy fix
(`5858e6f`: tests off the host state directory, the busy-timeout-
unparseable row test, `doctor` comments, the `storeRowLines` constant),
one `code-simplifier` round (`f4f5ac4`, seven leaf-first free functions
turned into methods on `leaf`) and its comment-only review fix
(`aa3b13e`), and this close (Task 8).

Two implementer dispatches for Task 5 died on network errors with
nothing committed (the first attempt and one resume); the task
re-dispatched clean and landed at `c91bf91`, `01902c3`, `e5e2eaf`.

### The gate evidence

`go test ./...` at HEAD (`aa3b13e`), cache cleared first (`go clean
-testcache`): **2m18.926s** wall (`internal/harness/conform` at
135.110s dominates; `internal/harness` at 10.963s and `internal/auth`
at 6.283s are the next largest; every other package under 5s).

`go test -count=1 ./internal/engine/store/...` at HEAD: **4.821s** test
time, **5.690s** wall including build.

`cairn-run-gate 'bash scripts/check.sh'` at HEAD, run twice in this
close (both against a warm Go build cache): **exit 0** both times,
99-line log, **1m10.093s** wall on the second (timed) run. The gitleaks
leg scanned about 120.67 MB in 1m0.3s to 1m4.8s across the two runs and
found no leaks. `go vet` and `golangci-lint run ./...` report 0 issues.
A true cold-cache (`go clean -cache`) gate run was not taken; doing so
would rebuild the whole module and every other lane's cache, outside
this task's scope. See "Could not verify" below.

Two staged legs stayed at warning, unchanged in kind from 5c-i: 325
process citations in Go comments (`CHECK_CITATIONS=warn`) and 128
exported-unused `internal/` identifiers (`CHECK_EXPORTED_UNUSED=warn`);
the new `internal/engine/store` package contributes none to either
list, because its only exported surface (`Inspect`, `Options`,
`Report`, and the report's field set) is consumed by
`cmd/dubplate-server`.

`internal/engine/store` holds 86 top-level `func Test` definitions
(`grep -c '^func Test' internal/engine/store/*_test.go`, table-driven
where a table exists), across three migrations
(`internal/engine/store/migrations/0001_ledger.sql` through
`0003_catalog.sql`) each with its own populated fixture under
`testdata/fixtures/`.

### What landed

Task 1 (`814a891`, then `1ffc499` fix) added the store leaf: layout
(`store.lock`, `db/`, `snapshots/`, `staging/`), the `flock(2)`-backed
lock with the 50 ms poll interval and 5 s holders-log interval
(R-6a14), the owner check that refuses a create when the effective uid
does not own the state directory, and the filesystem probe that refuses
SMB, Ceph, and AFS beside the spec's named list (E6/E9). The fix round
corrected lock ordering, owner-refusal naming, and a shape check the
first review flagged.

Task 2 (`4d80763`, then `2419a6f` fix) added the `modernc.org/sqlite`
driver dependency and the three connection classes (write pool, read
pool, maintenance) plus the inspect class, each DSN built through one
escaping builder (R-6a17) so a path carrying `#`, `?`, or `%` opens the
exact file rather than a truncated one. The fix round widened DSN test
coverage and corrected `maintenanceDSN`'s doc comment.

Task 3 (`2d863cb`, `opus`) added the store's one write method,
`(*handle).update`, and its recording transaction handle
(`*writeTx`): every statement method remembers the first error, and
once one is recorded every later statement method short-circuits
without reaching SQLite, so nothing a callback issues after a disk-full
or a constraint failure can commit in SQLite's post-rollback autocommit
mode (R-6a10). The callback's context carries the nested-write mark so
a callback using its caller's outer `ctx` cannot recurse into the one
connection.

Task 4 (`8f271d9`, `opus`) added the migration runner and its
pre-migration snapshot: one `pre-migrate-vN.db` per migrating run,
replaced by rename on a retry at the same version, free space checked
against `page_count × page_size` plus the `journal_size_limit` ceiling
before the copy (R-6a11, R-6a14). This task fixed the ledger table
shape (`migrations`: version, name, sha256 lowercase hex), the store
marker value (`0x64756270`, "dubp"), and the rebuild marker line
(`-- dubplate:foreign-keys-off`) ahead of Task 5's schema, decisions
the plan left to the implementer.

Task 5 (`c91bf91`, `01902c3`, `e5e2eaf`; no fix round) added the first
three migrations (ledger, jobs, catalog) and their fixtures, the
schema-invariant test that applies every migration to an empty database
and proves each fixture survives to head, and the criterion-9
job-fixture snapshot test. `owner_instance TEXT` was named in this
task. `e5e2eaf` dropped three resolution-id citations that had reached
shipped test comments.

Task 6 (`e4ae72e`, `opus`; no fix round) added the store's three open
modes (inspect, write, serve) and the "every mode's order" the file
classification table (R-6a5) walks: restore-intent and busy decided
first, then absent, then the marker and ledger reads. This task added
`Report.Reason` and made `Inspect` return only a `*Report` (no separate
error for a classified state), and made the write mode's "refuses a
newer store" path carry its own text distinct from every other refusal.

Task 7 (`dedc1b7`, then `5858e6f` fix; no further fix round on the
tidy pass beyond what the fix commit covers) added `doctor`'s one-line
store row (always present, per R-6a12/Ruling 2), `check`'s inspect
report line, and the busy-timeout registration
(`DUBPLATE_STORE_BUSY_TIMEOUT`, 15 s default, 1 ms floor, 2147483647 ms
ceiling) in the rung's own file, `cmd/dubplate-server/store.go`. Row
texts are this task's own. The fix commit moved `doctor`/`check` tests
off the host's real state directory, added the busy-timeout-unparseable
row test, corrected `doctor` comments, and named `storeRowLines` as a
constant.

The simplifier round (`f4f5ac4`, Opus 5.5) turned seven leaf-first free
functions (`createLeaf`, `statLeaf`, `probeLeaf`, and others of the same
shape) into methods on `leaf`. Its review returned `fix` on stale
comment text the refactor left behind (`aa3b13e`: renamed leftover
`createLeaf`/`statLeaf`/`probeLeaf` references in comments).

### The four rulings and their outcomes

All four ruled **yes** by Geoff on 2026-09-24 (plan `:67-140`):

1. **The 6.8M execution ceiling, planning recorded separately.** Yes:
   the flag sits at 5.44M. See "Both budgets" below for the actual.
2. **`doctor` exits 0 for absent, older, and current; nonzero
   otherwise.** Yes: landed as R-6a12 in Task 7 (`storeRowLines`, the
   exit rule applied identically in `check`).
3. **`serve` refuses a create or completion when `snapshots/` holds a
   copy.** Yes: landed in R-6a5's copy row and R-6a6's guard, applied
   in Task 6's serve outcome (in-package only; `serve`'s production
   caller is 6d).
4. **A caller's deadline bounds only the in-process wait; the busy
   timeout alone bounds the cross-process wait.** Yes: landed as
   R-6a18, pinned by a Task 3 test against the measured `BeginTx`/
   `_busy_timeout` behavior. No per-call busy-timeout mechanism was
   added; Task 3 was not split.

### The resolutions, R-6a1 through R-6a18

Each is stated in full in the plan at its own `### R-6aN` heading; this
table names where each landed at HEAD.

| Resolution | What it decides | Landed |
| --- | --- | --- |
| R-6a1 | Lane 2's package territory against 5c-i; the overlap is `doctor.go`, `doctor_test.go`, `main.go` in Task 7 | Confirmed: Task 0 found no 5c-i commits touching `internal/config`, `go.mod`, or `check.go` since `08befea` |
| R-6a2 | The exported surface at 6a is the inspect mode, its options, and its report only | `internal/engine/store/inspect.go` (Task 6, Task 7's callers) |
| R-6a3 | The busy timeout registers at 6a; the migration wait registers at 6d | `cmd/dubplate-server/store.go` (Task 7) |
| R-6a4 | No `os` descriptor on a live database file; OFD locking declined | `internal/engine/store` throughout; the two permitted `os` opens (`O_CREATE\|O_EXCL` pre-creates) in Task 1 and Task 4 |
| R-6a5 | A store carries a marker; every file state has one report; "newer" decides on `user_version` alone before the ledger reads | `internal/engine/store/inspect.go`, `migrate.go` (Tasks 4, 6) |
| R-6a6 | An interrupted create (zero-byte or unmarked header-only) is completed only by serve mode | `internal/engine/store/open.go` (Task 6) |
| R-6a7 | A restore-intent record at 6a is presence only; staging cleanup runs only in serve mode | `internal/engine/store/open.go` (Task 6) |
| R-6a8 | Times are integer UTC microseconds (`time.Time.UnixMicro()`) | `internal/engine/store/migrations/*.sql` (Task 5) |
| R-6a9 | Path rules are CHECK constraints at 6a; the Go validator waits for rung 7 | `internal/engine/store/migrations/0003_catalog.sql` (Task 5) |
| R-6a10 | The write method records every statement error itself; a callback cannot commit in autocommit mode after a rollback | `internal/engine/store/write.go` (Task 3) |
| R-6a11 | One pre-migration snapshot per migrating run, replaced by rename on retry | `internal/engine/store/snapshot.go`, `migrate.go` (Task 4) |
| R-6a12 | `doctor`'s store row is one line, always; the exit rule follows Ruling 2 | `cmd/dubplate-server/doctor.go`, `storeRowLines` (Task 7) |
| R-6a13 | `leases/` lands with 6b-i, not this pass | Not built; confirmed absent from Task 1's leaf layout |
| R-6a14 | Three named constants: 64 MiB `journal_size_limit`, 50 ms lock poll, 5 s holders-log interval | `internal/engine/store/lock.go`, `migrate.go` (Tasks 1, 4) |
| R-6a15 | Every migration ships with a populated fixture | `internal/engine/store/testdata/fixtures/` (Task 5) |
| R-6a16 | `doctor`'s goldens render from constructed reports, plus two end-to-end cases | `cmd/dubplate-server/doctor_test.go` (Task 7) |
| R-6a17 | One DSN builder escapes the path; no DSN can create | `internal/engine/store/conn.go` (Task 2) |
| R-6a18 | A caller's deadline does not shorten SQLite's busy wait (Ruling 4) | `internal/engine/store/write.go` (Task 3) |

### The declared coverage gaps

**Eight**, all reviewer-read or otherwise unclosed by this pass, none
blocking the exit gate:

1. The uid check's real-uid mutation (reviewer-read).
2. The `/proc/locks` inode-only mutation (reviewer-read).
3. The holder list's empty result on btrfs, where `stat` and
   `/proc/locks` disagree on the device.
4. The zero-byte detection by `os.Stat` (reviewer-read).
5. The leaf-owner ownership of `-wal` and `-shm` in criterion 15, which
   cannot fail when every test file has the test's uid.
6. The `IOERR`, `NOMEM`, and `INTERRUPT` rollback rows (reviewer-read).
7. The copy's durability under `_synchronous=FULL` (reviewer-read).
8. The OFD decline with its reason (R-6a4; the store declines OFD
   locking rather than adopting it, and this gap is the declined
   mechanism itself, not a defect).

### Both budgets

Tokens: **about 3.5M in subagent tokens** before this close, the merge
ritual, and the architecture reads (conductor-supplied: Task 0 about
0.11M, segment A about 1.55M, segment B about 1.35M plus one dead
dispatch, the tidy fix about 0.13M, the simplifier round and its review
about 0.35M), against the **6.8M execution ceiling** with the flag at
**5.44M** (Ruling 1). Attended time: **zero attended interactions after
approval**. Fix rounds: Task 1 one, Task 2 one, Tasks 3 through 7 none;
the simplifier round one, comment-only.

### What a later pass would be wrong to rediscover

- **A migration file carrying its own `COMMIT` is caught only at the
  runner's `Commit` call, after its tail already ran in autocommit**
  (`migrate.go`, around the migration-apply loop). 6b-i's restore path,
  which also runs SQL against the store, inherits this: a script that
  self-commits mid-file is not refused until the runner tries to close
  the transaction it thinks is still open.
- **A failed snapshot rename leaves staging occupied until serve's
  cleanup runs** (`snapshot.go`, around the pre-create/rename pair).
  Nothing but serve mode, holding the exclusive lock with no intent
  record, clears a staging leftover (R-6a7); a write-mode caller must
  not add its own cleanup.
- **The rebuild marker line (`-- dubplate:foreign-keys-off`) matches
  exact bytes only.** A CRLF line ending or a trailing space runs a
  rebuild with foreign-key enforcement on rather than off, silently.
- **Serve's settle loop has no iteration bound, and `serveOptions.logger`
  has no nil guard.** Both are unexercised at 6a (serve's only caller is
  the in-package test); 6d, which gives serve its production caller,
  inherits both as open items rather than closed ones.
- **Mutation 4 (`VACUUM INTO` on the write class) is unobservable under
  the exclusive lock**, and **mutation 11 (reading the ledger before the
  version check) panics inside `checkLedger`'s contract rather than
  reporting refused.** Neither is a defect in the shipped code; both are
  findings about the mutation harness's reach that a later reviewer
  should not re-derive from scratch.
- **`member_test.go` and `serve_test.go` build `newRootCmd` with
  `os.Environ()` but never dispatch `doctor` or `check`.** A future test
  added to either file that expects the store row to appear must
  dispatch one of those two commands itself; building the root command
  alone proves nothing about the store row's presence.

---

## The errata this pass owes

Drafted paste-ready for the conductor from the plan fold record's
"Owed errata" table (`docs/research/2026-09-23-rung-6a-plan-fold-record.md`,
E1-E11), with each ruling's outcome applied. All four rulings are yes,
so E4, E10, and E11 take their "yes" text.

- **E1 (R-6a3).** The spec fold record's `main.go` contested row: the
  migration wait registers at 6d with `serve`, not at 6a.
- **E2 (R-6a6).** Criterion 13 (`:938`): a zero-byte file and a
  header-only unmarked file are interrupted creates that serve
  completes; "not a store" holds for the inspect and write modes only.
- **E3 (R-6a2).** "Package territory" (`:213-216`): "6a exports the
  open" means the inspect mode, its options, and its report.
- **E4 (Ruling 1, yes).** The ceiling table row (`:1098`) reads 6.8M
  execution plus planning, not the spec's original 5.5M total.
- **E5 (R-6a9).** Paths (`:480-487`): the path refusal lands as schema
  CHECKs at 6a; the Go validator lands with rung 7's first writer.
- **E6.** Filesystem probe (`:256-258`): the probe also refuses SMB,
  Ceph, and AFS. It probes `db/` when it exists, else the leaf.
- **E7 (R-6a7).** Layout (`:235-236`): only serve mode, holding the
  exclusive lock with no intent record, removes staging leftovers;
  write mode never does.
- **E8.** Serve procedure step 1 (`:351`): step 1's shared wait is
  bounded by the migration wait and logs the holders, as step 2 does.
- **E9.** Ownership (`:238-242`): a create requires the effective uid
  to own the state directory, and otherwise refuses naming that owner.
- **E10 (Ruling 4, yes).** The write method's contract (`:331-332`):
  the busy timeout alone bounds the wait on another process's write
  lock, and a context does not shorten it (measured).
- **E11 (Ruling 3, yes).** Layout and the serve procedure's create:
  serve refuses a create when `snapshots/` holds a copy, naming it and
  `restore --from-file`, and inspect reports it as refused.

---

## The 6b-i handoff

1. **The surface 6b-i builds on.** The write method
   (`(*handle).update`) and its transaction handle (`*writeTx`), which
   6b-i exports with `ops/local`; the write mode, exported with
   `backup`; the snapshot function; the job table's columns (Task 5);
   `leases/`, which 6b-i creates (R-6a13).
2. **The intent record rule 6b-ii refines** (R-6a7): at 6a, every mode
   checks only for `restore-intent`'s presence; the format and the
   `reconcile-pending` state are 6b-ii's.
3. **The staging rule 6b-i inherits**: a shared-lock stager names its
   file by instance id, and its recovery removes only a file whose
   owner's lease is free, since serve's cleanup runs only under the
   exclusive lock.
4. **The snapshot's `wal_checkpoint(TRUNCATE)` result row**, which
   6b-i must read: under a shared lock with serve writing, it can
   report busy in the row's first column rather than as a Go error
   (logged, never a snapshot failure).
5. **A `VACUUM INTO` copy is a rollback-journal database**, so 6b-ii's
   restore sets WAL on the placed file.
6. **An interrupted file-backed connection is discarded by the
   driver** (R-6a18): 6b-i's restore, which also opens file-backed
   connections, cannot assume a context-cancelled connection is reusable.
7. **The runbook lines owed**: a root-run verb's refusal and the
   `runuser` fix; a create refused because the state directory's owner
   differs; the FUSE warning; the btrfs holder gap.
