# Rung 6, Pass 6b-i (The Ops Seam and Backup) Implementation Plan

Status: **revision 1, 2026-09-24; draft, unreviewed.** No lens review, fold,
or owner ruling has touched it. Three rulings for Geoff are open ("Rulings
for Geoff").

Sources: the ratified rung 6 spec
(`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`, revision
2.2, with the 2026-09-24 errata E1 to E11) and its fold record
(`docs/superpowers/research/2026-09-23-rung-6-spec-fold-record.md`); the
approved 6a plan (`docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md`,
revision 2.2), its "What this pass is not", Task 8's handoff, and conductor
action 6; the 2.0 streamer spec
(`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, "Forward hooks"
and "Prior art"); the parallel-build approach's ownership table
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`); and the
approved 5c-ii plan (`docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`),
which runs in lane 1 beside this pass.

**Code pointers were resolved at `dedc1b7`, the head of branch `lane-2-6a`,
before 6a merged.** The worktree held an uncommitted tidy fix when this plan
was drafted, so every pointer reads the committed tree
(`git show lane-2-6a:<path>`). Task 0 re-resolves every pointer after 6a
merges, and grades a moved pointer by content.

**What this pass is.** Pass 6b-i builds the ops seam and backup, the first
long-running operation. It creates `internal/ops` (the interface, the wire
types, the error codes, the privileges, and the catalog), `internal/ops/local`
(the job runner, leases, recovery, and cross-process cancel), and
`internal/backup` (the guards, the rclone transport, the snapshot send, and
pruning). It exports the store's write mode, write method, and a lease
primitive to their first callers. It lands the backup integration class, the
`backup` verb, and `doctor`'s backup rows. The spec's pass table assigns
exactly this set (spec `:148-151`, criteria `:990-1033`).

**What this pass is not.** No HTTP projection, no authorizer, and no
contract suite: those are 6b-ii. No `restore` verb and no intent-record
format: those are 6b-ii. No settings table and no seed rule: 6b-i's
constants stay constants until 6c re-homes them. No `serve` wiring and no
roster import: those are 6d. **Nothing in production creates or migrates a
store in this pass.** `backup` opens in write mode, which refuses an absent
store, so on every box `backup` refuses until 6d. No `systemd/`,
`scripts/deploy.sh`, or `env/` change, no backup timer, and **no deploy of
any kind.** A task that believes it needs any of these stops and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the
> pass-execute chain per task, sequential, in the three segments "Execution
> model" names. The implementer is `dubplate-implementer`. The conductor
> consumes structured reports and never reads a diff, a test log, or a gate
> transcript inline. Subagents never dispatch agents, so `code-simplifier`
> is the conductor's dispatch. The conductor owns `docs/STATUS.md`,
> `docs/HISTORY.md`, `ROADMAP.md`, the lane registry, and the
> contested-surface rows. **This pass runs in lane 2, a git worktree, beside
> 5c-ii in lane 1.** The live-executor check is the lane registry, never
> `pgrep -f Projects/dubplate`.

---

## Rulings for Geoff

Three yes-or-no questions. The spec's fourteen rulings and 6a's four stand
and are not reopened.

**1. Does 6b-i run under a 7.6M execution ceiling, with planning recorded
separately?** Recommendation: yes. Reason: the spec's 5.9M row is a total
that includes 0.95M of planning and prices tasks at 0.46M, while 6a's
measured record prices the close side (tidy fix, simplifier round, close,
merge ritual, and architecture reads) at about 2.1M, which the spec's 0.6M
close figure does not carry ("Token ceiling"). What each answer builds: yes
sets the flag at 6.08M, about half a task above the 5.75M projection. No
holds 5.9M as a total, an execution figure under the projection, so the flag
fires on plan and the conductor asks mid-pass.

**2. Does `doctor` ask rclone itself whether the backup remote supports
server-side move or copy, running `rclone backend features` under a 10 s
bound?** Recommendation: yes. Reason: rclone's own probe answers for the
installed binary and every backend, crypt included, while a copied feature
table goes stale and cannot see through a crypt remote (R-6bi9). What each
answer builds: yes runs one bounded subprocess per `doctor` run when the
flag is on, which may contact the provider for some backends; a timeout
reads "not checked" and never fails `doctor`. No builds a static table keyed
on `RCLONE_CONFIG_<NAME>_TYPE` from rclone's published "Optional Features"
table, with a crypt remote resolved to the type it wraps; it makes no network
call and reports "unknown backend" for any type missing from the table.

**3. Does the `backup` verb exit with the codes below?** 0 succeeded, 1
failed or refused before start, 69 unavailable, 75 conflict, and 130
canceled. Recommendation: yes. Reason: every existing verb exits 1 on any
error (`cmd/dubplate-server/main.go:51-54`), so 1 keeps that meaning, and
the three others take `sysexits.h`'s `EX_UNAVAILABLE` (69, "service
unavailable") and `EX_TEMPFAIL` (75, "temp failure; user is invited to
retry") and the shell's 128-plus-SIGINT convention (130). Rung 16's timer
and rung 9's ping read them. What each answer builds: yes fixes these in
the verb's help and in main's exit path. No uses 2, 3, and 4, which the
shell and cobra do not reserve but which carry no convention an operator
recognizes.

### Spec deviations that take effect on approval

These change the ratified spec's meaning. Approving this plan approves them;
the close drafts the erratum text (Task 8).

- R-6bi2: the lease primitive lives in `internal/engine/store`, and
  recovery policy lives in `ops/local`. The spec's package table names
  leases in both rows (`:200` and `:202`).
- R-6bi6: the backup's store copy is staged, verified, and sent from
  `staging/`, and removed after the send. It is never renamed into
  `snapshots/`. The spec's "renames it, fsyncs the directory by hand"
  (`:624-626`) holds for the pre-migration copy only.
- R-6bi7: `--max-delete 0` suffices for the zero-track case. A probe
  measured rclone v1.74.3 counting a move into `--backup-dir` against
  `--max-delete`, so the spec's fallback abort (`:647-649`) is not built.
- R-6bi8: 6b-i declares four of the five constants.
  `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` has no reader until 6d's `serve`,
  and an unused unexported constant fails `golangci-lint`'s `unused` check,
  so 6d declares it and 6c re-homes it after.
- R-6bi9: the per-backend feature check reads `rclone backend features`,
  not a copied table (Ruling 2's yes). The installed manual carries no
  per-backend table; this plan quotes the website's rows as evidence only.
- R-6bi12: the transfer child inherits the kind lease and the store's
  shared lock, and runs in its own process group.

---

## Primary inputs

Read these before Task 1, in this order.

1. **The rung 6 spec, revision 2.2.** The sections this pass implements:
   "What rung 6 builds" (`:141-160`), "Package territory" (`:193-221`),
   "The ops seam" through "The job model" and "Operation availability"
   (`:533-591`), "The backup job" and "Credentials" (`:611-672`), the 2.0
   hooks (`:877-891`), the prior-art table (`:895-915`), the privileged-field
   table (`:919-937`), and 6b-i's twenty-one criteria (`:990-1033`). Line
   numbers are as read at `aafbfdb`; Task 0 re-reads them.
2. **The 6a plan's resolutions** R-6a2 (exports), R-6a4 (no `os` descriptor
   on a live database file), R-6a7 (staging), R-6a8 (times), R-6a10 (the
   write method's recorded errors), R-6a13 (`leases/`), R-6a17 (one DSN
   builder), and R-6a18 (the busy wait); and Task 8's handoff items 1 to 7.
3. **`internal/engine/store` as shipped at `dedc1b7`**: `leaf.go`,
   `lock.go`, `conn.go`, `write.go`, `migrate.go`, `snapshot.go`,
   `open.go`, `inspect.go`, and `migrations/0002_jobs.sql`.
4. **This plan's quoted behaviors.** Each is quoted from `go doc` at
   `go1.27.1 linux/amd64`, a man page, rclone v1.74.3's installed manual
   (`man rclone`), or a probe run on 2026-09-24 under
   `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/89b265c3-0b9e-4d9c-b998-df7b9970c32a/scratchpad/`
   (`6bi-p1/`, `6bi-p2/`). The probes may be gone; each task re-proves what
   it relies on through its own gate-resident test. A task that relies on a
   behavior this plan does not quote confirms it the same way and reports
   what it read.

---

## Goal

After this pass, an operator can run `dubplate-server backup` against a
store `serve` created, and the run either leaves the library tree and a
verified store snapshot on any rclone remote, or refuses before rclone runs
with a reason an operator can act on. A second start, a cancel from another
process, a crash of either process, and a SIGINT each end in a named job
state that a later process reads correctly. `doctor` reports whether the
backup configuration can work before the first run tries it.

---

## Owner touchpoints

The three rulings above, at the plan-approval gate. Nothing else is
expected. What would become one: a new module dependency (this pass adds
none); a threshold default changed on a task's own judgment; any change to
what the box runs.

---

## Token ceiling: 7.6M execution for 6b-i (Ruling 1)

The basis is 6a's recorded spend (`docs/STATUS.md`, "Next action" 1):
about 1.66M for Tasks 0 to 3 and about 1.35M for Tasks 4 to 7, zero fix
rounds in segment B, and a re-derived total of about 5.1M. So Tasks 0 to 7
cost about 3.0M, and the close side (the tidy fix, the simplifier round,
segment C, the merge ritual, and two architecture reads) is projected at
about 2.1M. Planning is recorded against the rung and outside this ceiling,
as 6a's Ruling 1 set.

**The arithmetic.**

1. Task 0 at `opus` in fresh context: **0.20M**.
2. Seven implementation tasks at 6a's measured 0.40M ((3.0M − 0.2M) / 7):
   **2.80M**.
3. Two costs 6a did not carry: the first subprocess tests against a real
   `rclone`, and the re-executed test binary that makes a second process
   real (Tasks 3, 4, and 6). Add **0.30M**.
4. One fix round, since 6b-i's cross-process paths are new: **0.25M**.
5. The close side at 6a's projection, plus three more architecture reads
   (`internal/ops`, `internal/ops/local`, `internal/backup`, beside
   `internal/engine/store` and `cmd/dubplate-server`): **2.20M**.
6. **The projection is 5.75M.**

A ceiling of **7.6M** puts the 80% flag at **6.08M**, 0.33M above the
projection, so the flag marks an overrun rather than firing on plan. The
pre-named split in Task 3 adds about 0.40M and brings the projection to
6.15M, just past the flag, so the conductor asks at that boundary under the
80% rule. The spec's row reads 5.9M: 0.95M planning, 0.2M Task 0, nine tasks
at 0.46M, and a 0.6M close. Re-derive at the segment A boundary from Tasks 1
to 3's reports.

Checkpoint: STATUS at every segment boundary, after Task 3, after Task 7,
and at the close. Every task reports its spend, its mutation outcomes, and
its measured gate wall time.

Deliverable count: twenty-five across Tasks 1 to 8, plus Task 0's report.
Tasks 1, 3, 4, 6, and 8 carry four each; Task 2 carries two; Tasks 5 and 7
carry three each. No task passes four.

---

## Global constraints

- **Lane 2, a git worktree.** The conductor creates it from master after 6a
  merges, before Task 0, and writes the lane registry row. Every task, gate,
  and commit runs inside the worktree.
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`, and
  no task ssh's to the box.
- **Package territory, stated as a wall.** This pass creates `internal/ops`,
  `internal/ops/local`, and `internal/backup` (every file and `testdata/`).
  It edits `internal/engine/store` (Task 1 only). It creates
  `cmd/dubplate-server/backup.go` and its test, and edits
  `cmd/dubplate-server/main.go` (two registration lines and main's exit
  path), `doctor.go`, `store.go`, and their tests. It edits
  `internal/config/keys.go` for one removal: the "backup targets" reserved
  class (`keys.go:61-71`), the removal half of the rung 1b registry
  contract, in the same commit that registers its supersessions, as
  `799419a` and `ee79b5b` did for rung 5. **It writes nothing under
  `internal/auth`, `internal/identity`, `internal/serve`, `internal/notify`,
  `internal/harness`, `internal/atomicfile`, the rest of `internal/engine`,
  `web/`, `systemd/`, `scripts/`, `env/`, `config/`, `Makefile`,
  `CLAUDE.md`, `corpus/`, `tools/`, or `_salvage/`.** It writes no `docs/`
  file outside Task 8's drafts. A task that believes it needs one stops and
  reports.
- **No new module dependency.** rclone runs as a subprocess, never as a Go
  module; `librclone` is declined (spec prior art).
- **Imports follow the spec's table** (`:198-205`): `internal/ops` imports
  `internal/identity` and the standard library only; `internal/ops/local`
  imports `ops`, `store`, and `backup`; `internal/backup` imports `store`
  and the standard library. The engine has no HTTP knowledge and reads no
  environment. `internal/config` imports none of the new packages.
- **The store's rules carry over.** No process that holds a store
  connection opens the database, `-wal`, or `-shm` file through `os`
  (R-6a4); every size comes from `os.Stat`. Every DSN comes from the one
  builder (R-6a17). Times are integer UTC microseconds bound as `int64`
  (R-6a8). Every store write goes through the write method, whose callback
  receives the store-owned transaction handle (R-6a10).
- **Nothing is exported ahead of a cross-package caller.** Within the pass,
  an export whose production caller lands in a later task is permitted, and
  the exported-unused leg warns meanwhile; at the close every export has a
  production caller (exit gate 11).
- **Every store write carries a context with a deadline.** The write method
  refuses one without (6a criterion 6). A caller whose own context is
  already done, the `backup` verb after SIGINT above all, derives a fresh
  deadline context from `context.WithoutCancel` for its cancel write, or
  the write fails at once and the cancel is lost (Task 6).
- **Assertion discipline at the repo's bar.** No silent-success fakes, no
  vacuous assertions, no expectation derived from the code under test.
  Every hard-coded expectation cites the spec, a quoted document, or a probe
  this plan records. Every criterion names the fixture state it needs.
- **Universality.** No estate hostname, path, remote name, email, or
  personal name anywhere. Every tunable is a named unexported constant whose
  value this plan states, or a registered key.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar. `ctx`
  first where a call can block. Errors lowercase and context-first.
  Sentinels only where a caller branches. **The em dash is banned in Go
  comments.** No process citation in a shipped comment.
- **No `t.Parallel()` in `internal/engine/store`** (6a's umask rule). The
  new packages may run tests in parallel only where no test sets the umask
  or a process-wide signal handler.
- **The gate is `bash scripts/check.sh`, green at every task end**, run
  through `cairn-run-gate` inside the worktree and re-issued on exit 75
  until it prints `gate exit:`. After 5c-ii lands on master, the lane's
  first rebase needs `npm ci --prefix web` in the worktree before its
  in-tree gate (5c-ii conductor action 7). Anything reported as evidence
  runs with `-count=1` and reports wall time. Specific files only. Commit
  footer `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` for an
  `opus` task and the dispatch's footer otherwise.
- **A real `rclone` is a gate prerequisite from Task 4.** A test that needs
  it fails, never skips, when `exec.LookPath("rclone")` fails. The close
  routes the prerequisite to CLAUDE.md's Gate section (Task 8).

---

## Privileged fields and their channels

| Field | Channels that may set it at 6b-i | What refuses every other channel |
| --- | --- | --- |
| A job's `id` | `ops/local`'s start, through the store's id function (16 random bytes, `id.go`). | No request type carries an id on start. A cancel or a status read names an id; a malformed one is invalid request, an unknown one is not found. |
| A job's `actor` | `ops/local`'s start, from the method's actor parameter alone (R-6bi4). | No request type has an actor field; a JSON body carrying one decodes into a type with no such field (criterion 2). |
| A job's `owner_instance` | The runner that inserts the record, naming its own held lease. | The runner inserts no record before its lease is held (Task 3). |
| A job's `state` | Its owning runner (queued to running to a terminal state); the cancel operation (queued to canceled only); recovery (queued or running to interrupted, only when the owner's lease is free). | Every state write names the state it expects in its `WHERE`, so a terminal state never changes; a late terminal update to an interrupted record matches no row and is refused and logged (criterion 5). |
| A job's `cancel_requested` | The cancel operation alone. | The runner reads it and never writes it. |
| A job's `failure_code` and `failure_message` | The owning runner, in the transition to failed. | No request type carries either. |
| Deleting a job record | Retention at each job start, which never removes the newest succeeded record of a kind (Task 3). | No operation deletes a record. |
| A lease file in `leases/` | Its owning runner creates and holds it; recovery removes one only after acquiring it, which proves its owner dead. | Recovery never removes a file it could not lock. |
| A backup staging file (`staging/backup-<instance>.db`) | The backup job whose instance names it; recovery removes one only when that instance's lease is free. | Serve mode's cleanup also removes it, under the exclusive lock (R-6a7). |
| Deleting a remote snapshot or archive directory | The backup job's prune step, after a successful run (Task 5). | The latest snapshot is never pruned (ruling 10). |
| Deleting a local pre-migration or pre-restore copy | The backup job's local deletion, after a successful run, only for copies older than the snapshot it sent (ruling 7, Task 5); the migration runner's retention with backup off (6a). | Neither touches a name outside `pre-migrate-v<N>.db` and `pre-restore-<stamp>.db`. |
| `DUBPLATE_BACKUP_ENABLED`, `DUBPLATE_BACKUP_REMOTE` | The environment, through the registry's backup integration class (Task 6). | The store and `internal/backup` read no environment; `cmd/dubplate-server` passes both values. |

---

## Resolutions this plan takes rather than discovers

Each resolution names its site. A task landing the surface a resolution
decides writes nothing about the resolution in a shipped comment; the close
maps each resolution to its landing commit.

### R-6bi1: lane 2, and the package territory against 5c-ii

5c-ii edits `cmd/dubplate-server/main.go` (one line in
`newProductionRegistry`), `serve.go`, `serve_test.go`, `web.go`,
`internal/auth`, `internal/identity`, `internal/serve`, `web/`,
`scripts/check.sh`, `Makefile`, `go.mod`, `.gitleaks.toml`, and
`.gitignore`, and leaves `internal/config`, `doctor.go`, and `check.go`
untouched (5c-ii "Contested surfaces"). 6b-i's overlap is `main.go` (Task 6
adds one registration line, one `AddCommand` entry, and an exit-code branch
in `main`), and `doctor`'s output (5c-ii adds threshold rows through the
registry; 6b-i adds rows in `doctor.go`). Tasks 6 and 7 are last before the
close so the fewest commits touch the shared files. `internal/ops` imports
`internal/identity`, which 5c-ii edits; 6b-i only imports it, and Task 0
reports `Identity`'s fields at the merged head.

### R-6bi2: the store owns the lease primitive; `ops/local` owns recovery

The spec's package table names leases in the store's row and in
`ops/local`'s. The store already owns every `flock` in the leaf
(`lock.go:81`, `tryLock`), the leaf's paths (`leaf.go:28-38`), and the
holder list from `/proc/locks`. So the store exports one lease primitive:
take an instance lease under `leases/`, take a kind lease
(`leases/kind-<kind>`), and try an instance's lease, which on success proves
the owner dead and removes the file while holding it. `ops/local` owns what
a lease means for a job record: the recovery table, the conflict answer,
and the staging cleanup. `leases/` joins the leaf (R-6a13): serve mode's
create makes it at 0700, and the leaf's completeness check counts it. No
production leaf exists before 6d, so no leaf in the field lacks it.

`flock(2)` gives the property every in-process test relies on: "If a
process uses open(2) (or similar) to obtain more than one file descriptor
for the same file, these file descriptors are treated independently by
flock(). An attempt to lock the file using one of these file descriptors
may be denied by a lock that the calling process has already placed via
another file descriptor." So a second `ops/local` in one test process sees
the first's lease as held, and recovery in the owning process never reads
its own lease as free.

### R-6bi3: the exported write surface

Task 1 exports, each with its first caller in this pass: the write-mode
open (caller: `ops/local`'s constructor, through the `backup` verb); the
handle's close; the write method and its transaction handle (caller:
`ops/local`); the handle's read-only query path for the reads a decision
does not rest on (caller: `ops/local`'s status read and conflict answer);
the lease primitive (R-6bi2); the id function; and the backup snapshot
function (R-6bi6). Serve mode, the migration runner, and the pre-migration
snapshot stay unexported until 6d. A read a write decides on runs inside
the write callback, never on the read path (spec `:355`).

### R-6bi4: the actor is a method parameter with two kinds

The interface takes the acting identity as a parameter of every method
(spec `:538-541`). An actor is either a verified `identity.Identity` (a
network caller, from 6b-ii) or a local operator named by the
operating-system user (a CLI verb). The job record's `actor` column holds a
string with the kind as its prefix: `member:<principal>` or
`local:<username>`. The local username comes from `os/user.Current`, and
falls back to `uid:<n>` when the uid has no passwd entry, the shape
`ownerName` already uses (`leaf.go`). `ops/local` enforces no privilege:
the projection's authorizer does, in 6b-ii, and a local operator is the
leaf's owner.

### R-6bi5: the operation vocabulary at 6b-i

Four operations, each in the catalog with a privilege from the closed set
`member` or `admin`: start a backup (`admin`), read a job's status
(`admin`), cancel a job (`admin`), and list the catalog (`member`). Job
status and cancel are `admin` because the only kind at 6b-i is backup;
rung 7, which adds member-started jobs, revisits them. A job handle is the
job id as 32 lowercase hex characters on the wire. The error vocabulary is
the spec's seven sentinels, each with a stable wire code: `unavailable`,
`not_found`, `conflict`, `invalid_request`, `forbidden`, `canceled`, and
`internal`. 6b-i raises five of them; `forbidden` and `internal` wait for
6b-ii's projection and contract suite.

### R-6bi6: the backup's store copy is sent from staging and never kept

The job pre-creates `staging/backup-<instance>.db` at 0600, runs `VACUUM
INTO ?` on a maintenance connection, verifies the copy through the inspect
DSN (integrity `ok`, the binary's version, the store marker), sends it with
rclone, and removes it. The remote is its home. A rename into `snapshots/`
would put a backup copy beside ruling 7's pre-migration copies, which the
lost-database guard and `doctor`'s count read, and would grow the leaf by
one store per run. Its loss before the send costs a re-run, never data, so
no local durability step buys anything. The pre-migration path keeps its
rename and hand fsync (6a Task 4). The snapshot ends with
`wal_checkpoint(TRUNCATE)`, whose busy result row is logged, never a
failure (6a handoff item 4).

### R-6bi7: `--max-delete 0` is the zero-track guard

The spec left open whether rclone counts a move into `--backup-dir` against
`--max-delete` (`:647-649`). A probe (`6bi-p1/`, rclone v1.74.3, `local`
backend) settled it. After a first sync of three files, one source file was
removed and `rclone sync src remote/library --backup-dir remote/archive/s1
--max-delete 0` ran: it exited 7, logged `Failed to sync with 2 errors:
last error was: failed to delete 1 files`, archived nothing, and left
`remote/library` holding all three files. So a move into the backup
directory counts as a delete, and the fallback abort is not built. The same
probe measured a modified file being archived under `--max-delete 0`
(`Server Side Moves: 1`), so the limit bounds deletions only, never
replacements. Task 4 pins both facts against the real binary.

rclone's manual, quoted: `--max-delete` "tells rclone not to delete more
than N files. If that limit is exceeded then a fatal error will be
generated and rclone will stop the operation in progress." Exit code 7 is
"Fatal error (one that more retries won't fix, like account suspended)
(Fatal errors)."

### R-6bi8: four constants, with their values

Each is a named unexported constant at 6b-i, re-homed by 6c as a
store-homed key with a researched default. The values below are this pass's
working values; 6c's research may change them.

- **Job retention** (`DUBPLATE_JOB_RETENTION`): 90 days, in `ops/local`.
- **Guard fraction** (`DUBPLATE_BACKUP_GUARD_FRACTION`): 0.5, in
  `internal/backup`. A tree holding fewer regular files than half the
  catalog's tracks aborts.
- **Remote snapshot keep** (`DUBPLATE_BACKUP_SNAPSHOT_KEEP`): 14, in
  `internal/backup`.
- **Archive retention** (`DUBPLATE_BACKUP_ARCHIVE_RETENTION`): 90 days, in
  `internal/backup`.

`DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` is not declared here (a spec
deviation above).

Four further constants carry no re-homing: the cancel poll interval (2 s,
how often a running job re-reads `cancel_requested` while rclone runs); the
transfer's `WaitDelay` (10 s); the `doctor` features bound (10 s, Ruling
2); and the stamp layout (`20060102T150405Z`, UTC, which sorts
lexicographically).

### R-6bi9: the backend features check

rclone's manual: "You can also discover information about the backend
using ... `rclone backend features remote:`." A probe (`6bi-p1/`) measured
its JSON `Features` object: a `local` remote reads `Copy` false, `Move`
true; an `http` remote reads both false; a `crypt` remote over `local`
reads `Move` true; a remote with no `_TYPE` exits 1 with `didn't find
section in config file`; a crypt remote wrapping an undefined remote exits
1 naming the wrapped remote. rclone's website table
(`rclone.org/overview/`, "Optional Features", read 2026-09-24) agrees for
`local` (Copy ✖, Move ✔) and `http` (both ✖), and lists S3 and B2 as Copy ✔,
Move ✖, so "lacks server-side move and copy" means both false. The check
runs only with the flag on, after the name and `_TYPE` checks pass.

### R-6bi10: exit codes and `main`'s exit path

Ruling 3's codes. `main` today prints any error and exits 1
(`main.go:51-54`). Task 6 adds one branch: an error carrying an exit code
(one unexported type in `package main`, matched with `errors.As`) exits
with that code, and every other error keeps exit 1. The verb's help lists
the five codes.

### R-6bi11: the verb survives its own signal

`executeWithSignalContext` cancels the verb's context on SIGINT or SIGTERM
(`main.go`, quoting `signal.NotifyContext`: the context "is marked done ...
when one of the listed signals arrives"). The job must not run under that
context, or a signal would kill rclone and bypass the record. So the verb
starts the job on a context derived from `context.WithoutCancel`, and on
the signal it calls the cancel operation with a fresh deadline context,
then keeps polling until the record is terminal. The verb also treats
SIGHUP as a cancel, because a closed terminal sends it and its default
action would kill the verb while rclone, in its own process group
(R-6bi12), survives.

### R-6bi12: the transfer child's process group and inherited locks

rclone runs under `exec.CommandContext`. Its `Cancel` sends SIGINT to the
child, and `WaitDelay` bounds the rest. `os/exec`, quoted: "If WaitDelay is
non-zero, it bounds the time spent waiting on two sources of unexpected
delay in Wait: a child process that fails to exit after the associated
Context is canceled, and a child process that exits but leaves its I/O
pipes unclosed," and "If the child process has failed to exit ... then it
will be terminated using os.Process.Kill." `CommandContext` "sets the
command's Cancel function to invoke the Kill method on its Process, and
leaves its WaitDelay unset," so both are set explicitly.

The child runs with `SysProcAttr.Setpgid` ("sets the process group ID of
the child to Pgid, or, if Pgid == 0, to the new child's process ID"), so a
terminal's SIGINT reaches the verb alone and the job ends canceled through
its record, never failed because rclone died first. `Pdeathsig` is declined:
its doc says "the signal is sent on thread termination, which may happen
before process termination" (go.dev/issue/27505).

An orphaned child is still possible when the parent is killed outright. So
the child inherits the kind lease's and the store lock's descriptors
through `ExtraFiles`. `flock(2)`: "Locks created by flock() are associated
with an open file description ... the lock is released either by an
explicit LOCK_UN operation on any of these duplicate file descriptors, or
when all such file descriptors have been closed." An orphan therefore keeps
a second backup and a restore out until it exits. For rung 16's unit, the
close records a runbook line: `KillMode=mixed` sends SIGTERM "to the main
process while the subsequent SIGKILL signal ... is sent to all remaining
processes" (`systemd.kill(5)`), which lets the verb cancel through the
record first.

### R-6bi13: `doctor` reads backup facts through the inspect report

`doctor` opens nothing for writing. The inspect report gains read-only
fields for a store that classifies at or above version 2 and at or below
the binary's version: the newest succeeded job's finish time per kind, the
`-wal` file's size, the free bytes on the leaf's filesystem, and the count
and total size of `snapshots/` copies. Sizes come from `os.Stat` and
`statfs`. A newer store reports none of them, since this binary cannot
trust its schema. The job query runs on the inspect DSN.

---

## Review focus

These are the input classes most likely to bite, each pinned by a test in
the task that owns it.

1. **A second process that owns a running job.** Expected: an open in
   another process leaves it alone. Pinned in Task 3.
2. **A process killed mid-job.** Expected: the next open marks its records
   interrupted and removes its lease file, and a late update from anywhere
   is refused. Pinned in Task 3.
3. **A wrong or wiped library mount.** Expected: no rclone invocation.
   Pinned in Task 4.
4. **A zero-track catalog over a populated remote.** Expected: a run that
   would archive a file ends failed. Pinned in Task 4.
5. **SIGINT at a terminal mid-sync.** Expected: the job ends canceled, no
   child survives `WaitDelay`, and the verb exits 130. Pinned in Task 6.
6. **A prune that deletes the latest snapshot.** Expected: never. Pinned in
   Task 5.
7. **A backup remote an adopter configured wrong.** Expected: `doctor`
   names the defect before the first run. Pinned in Task 7.

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched by the conductor after 6a merges
and before any implementer. Reads the spec, the 6a plan's handoff, and this
plan against the merged tree. **Its output is blockers with evidence, not
fixes.** It writes no code and edits no file.

**Every check runs against the real HEAD, and the task reports that HEAD.**
This plan's code pointers were resolved at `dedc1b7` on `lane-2-6a`. Grade a
moved pointer by content: a pointer whose content survives at any offset is
a note naming the new line, and a pointer whose content is gone is a
blocker.

**The first six items are blockers: a "no" stops the launch.**

1. **Has 6a merged to master, and is the worktree fresh from it?** Report
   `git worktree list`, the lane registry rows, and `git status` in both
   trees. *(Expected: lane 1 on 5c-ii in the main tree, lane 2's worktree
   fresh from master with 6a merged.)*
2. **Does 6a's tidy fix or simplifier round change a surface this plan
   names?** Report the diff from `dedc1b7` to the merged head for
   `internal/engine/store` and `cmd/dubplate-server/store.go`, by function.
   A renamed or reshaped `update`, `writeTx`, `openWrite`, `takeSnapshot`,
   `copyInto`, `verifyCopy`, `checkFreeSpace`, `tryLock`, `lockHolders`,
   `createLeaf`, `statLeaf`, `Report`, or `storeRowStatus` is a blocker
   until the conductor maps it.
3. **Do `internal/ops`, `internal/ops/local`, and `internal/backup` not
   exist, and does no file import `os/exec` for `rclone`?**
4. **Is the "backup targets" reserved class still at `keys.go` with its six
   names, and is none of `DUBPLATE_BACKUP_ENABLED`,
   `DUBPLATE_BACKUP_REMOTE`, or the six retired names declared anywhere
   else?** Report `git grep` for each.
5. **Has 5c-ii touched `internal/config`, `doctor.go`, `check.go`,
   `store.go`, or `internal/engine/store`?** Any commit is a blocker for
   R-6bi1.
6. **Does `rclone` resolve on the gate's `PATH`, and at what version?**
   Report `rclone version`. A version other than v1.74.3 is a note, and
   Task 4 re-runs R-6bi7's probe against it.

**Notes (not blockers):**

7. Report `identity.Identity`'s fields at HEAD (`internal/identity`), which
   5c-ii may change and `internal/ops` imports.
8. Report `newProductionRegistry`'s body, `newRootCmd`'s `AddCommand`
   block, and `main`'s exit path at HEAD.
9. Report `doctor.go`'s row order and `TestDoctorCmd_LineCountFollowsRegistry`'s
   arithmetic at HEAD, which Task 7 extends.
10. Report `migrations/0002_jobs.sql`'s columns at HEAD, which Task 3 writes
    against; a changed column is a blocker for Task 3 alone.
11. Report the spec pointers this plan cites, which errata may have moved.
12. Report whether the templates still carry the six retired names
    (`env/dubplate.env.template:49-53` and `:89`), which the template
    classification test must read as superseded after Task 6.

**Deliverable count: 1** (the report).

---

## Task 1: The store's surface for 6b-i: exports, leases, and the backup copy

Model: `opus`. Depends on Task 0. **Lands 2.0 hook 10** (the backup copy by
`VACUUM INTO`), and exports hook 9's write method to its first caller.

The store gives 6b-i what it builds on, and closes two of 6a's carried
review items in the same files.

**Files:**
- Create: `internal/engine/store/lease.go`, `internal/engine/store/lease_test.go`
- Modify: `internal/engine/store/leaf.go`, `open.go`, `write.go`,
  `conn.go`, `snapshot.go`, `migrate.go`, `inspect.go`, `id.go`, `doc.go`,
  and their tests

**Outcomes:**

- **The exports** R-6bi3 names, each documented at the Go Doc Comments
  bar. The write method's doc keeps its measured contract (R-6a18). The
  inspect report gains R-6bi13's fields.
- **`leases/`** joins the leaf (R-6bi2): created by serve mode at 0700,
  counted by the completeness check. A write-mode open of a leaf without
  `leases/` refuses naming the missing directory and "start serve to finish
  creating it," the lockless leaf's wording.
- **The lease primitive.** An instance lease is an exclusive `flock` on
  `leases/<instance>`, where the instance is 32 lowercase hex characters
  from the id function, created 0600 with `O_CREATE|O_EXCL` and held for the
  handle's life. A kind lease is an exclusive, non-blocking `flock` on
  `leases/kind-<kind>`, created if absent and never removed, whose busy
  error names the holders. Trying another instance's lease returns held,
  free (acquired, then the file removed while held, then closed), or absent
  (no file), and refuses on any other error naming the path.
- **The backup copy** (R-6bi6): free space for one copy plus the WAL bound
  (6a's `checkFreeSpace`), pre-create `staging/backup-<instance>.db` at 0600,
  `VACUUM INTO ?` on a maintenance connection, verify through the inspect
  DSN, and return the path. The caller removes it. It ends with
  `wal_checkpoint(TRUNCATE)`, logging a busy row.
- **Carried item: a failed rename leaves no stage.** When the pre-migration
  copy's rename fails (`snapshot.go:73`), the stage is removed, and a
  removal error is joined to the rename error. Today a failed rename leaves
  `staging/pre-migrate-vN.db`, which refuses every later migration at that
  version until serve's cleanup runs at the next start.
- **Carried item: a near-miss rebuild marker is refused.** `readMigrations`
  refuses a set whose file's first line, with trailing spaces, tabs, and one
  `\r` removed, equals the rebuild marker while the raw line does not. The
  refusal names the file. Today such a file is not a rebuild
  (`migrate.go:163` matches exact bytes), so it runs on the write pool with
  foreign keys on.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| `leases/` | present, 0700 | proceed |
| `leases/` | absent in an existing leaf | write mode refuses, naming the directory and "start serve to finish creating it" |
| Instance lease | created and locked | held for the handle's life |
| Instance lease | the file already exists (an id collision) | refused, naming the path; never reused |
| Kind lease | free | held for the job's life |
| Kind lease | held by another descriptor | busy, naming holder pids or "held by another process" |
| Try another's lease | held | held; nothing removed |
| Try another's lease | free | free; the file removed while held |
| Try another's lease | absent | absent |
| Try another's lease | open fails with anything but not-exist | refused, naming the path and the error |
| Backup copy | free space short | refused, naming the need, the free bytes, and the path; no stage created |
| Backup copy | the stage already exists | refused, naming the path (recovery removes a dead instance's stage first) |
| Backup copy | fails verification | refused; the stage removed |
| Backup copy | checkpoint reports busy | logged; the copy stands |
| Pre-migration rename | fails | refused; the stage removed; the store at N |
| Rebuild marker | exact | a rebuild |
| Rebuild marker | a near miss (`\r`, trailing space or tab) | the set refused, naming the file |
| Rebuild marker | any other first line | an ordinary file |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Two handles in one test process: the second's kind lease on `backup` is
  busy while the first holds it, and free after the first releases. The
  second's try of the first's instance lease reads held. **Fixture state:**
  a store at the binary's version, both handles open.
- After the first handle's lease descriptor is closed without removal (a
  dead owner), a try reads free and the file is gone; a second try reads
  absent.
- A backup copy taken while a second handle commits in a loop passes the
  integrity check, reads the binary's version and the marker, holds every
  row committed before the copy began, and holds no row of a transaction
  left open during it. **Fixture state:** a populated `jobs` table, a
  writer inserting rows in a loop, and one transaction held open across
  the copy. This is spec criterion 9's store half.
- The copy's stage and the lease files are 0600, and `leases/` is 0700,
  under a 0022 umask.
- The rename failure: `snapshots/` made unwritable, a migration from a
  populated store at N refuses, and `staging/` is empty afterward.
- The near-miss rows: a synthetic set whose file 2 begins with the marker
  plus `\r\n`, and one with a trailing space, are each refused naming the
  file; the exact marker still runs as a rebuild.
- The inspect report's new fields read back for a current store with a
  succeeded `backup` record, a 1-byte-or-larger `-wal`, and two copies in
  `snapshots/`; a newer store reports none of them.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Take the kind lease shared → the two-handle busy row passes both.
2. Remove the lease file before acquiring it → report whether the free row
   still holds, and name the race (a live owner's file removed) as a
   reviewer-read row if no test fires.
3. Treat absent as held → the dead-owner second-try row fails.
4. Run the backup copy on the write pool → report the effect on the looping
   writer (6a's stall), or name it reviewer-read.
5. Keep the stage on a rename failure → the rename-failure row fails.
6. Compare the marker after `strings.TrimSpace` and accept it → the
   near-miss rows run a rebuild instead of refusing.
7. Open the database file through `os` to size it → report which R-6a4 row
   fires, or name it reviewer-read.

**2.0 hook:** lands hook 10. **Prior art:** SQLite `VACUUM INTO`, adopted;
the driver's backup API (`conn.go:1208-1219` in the module) stays the
recorded alternative.

**Deliverable count: 4** (the exports with the report fields, `leases/` and
the lease primitive, the backup copy, the two carried items).

---

## Task 2: `internal/ops`: the interface, wire types, codes, privileges, and catalog

Model: `sonnet`. Depends on Task 0. **Independent by files of Task 1.**
**Lands the ops-seam half of 2.0 hook 9** (the seam a later projector joins).

**Files:**
- Create: `internal/ops/doc.go`, `ops.go`, `errors.go`, `catalog.go`, and
  their tests

**Outcomes:**

- **One interface** with R-6bi5's four operations. Each method takes `ctx`,
  the actor (R-6bi4), and a request type, and returns a response type and
  an error. Request and response types are JSON-serializable, with
  explicit `json` tags in snake case, and none carries an actor field.
- **The job model's types:** the handle, the six states (queued, running,
  succeeded, failed, canceled, interrupted) with a terminal predicate, and
  a status type carrying the state, the kind, the phase, the failure code
  and message, and the times.
- **The error vocabulary:** seven sentinels, each wrapped by an error type
  that carries its stable wire code and, for unavailable, a reason.
  Callers match with `errors.Is`.
- **Privileges and the catalog:** a privilege type whose zero value is
  invalid; a catalog entry carries the operation's name, its privilege, its
  availability, and the reason when unavailable. Building a catalog with a
  zero-privilege entry fails (spec criterion 1).

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Catalog entry privilege | `member` or `admin` | accepted |
| Catalog entry privilege | zero | construction fails, naming the operation |
| Job handle on the wire | 32 lowercase hex | accepted |
| Job handle on the wire | empty, wrong length, uppercase, or non-hex | invalid request, naming the handle |
| Job state | any of the six | its name on the wire |
| Unavailable | constructed without a reason | construction refuses; every unavailable answer names why |

**Acceptance criteria:**

- Spec criterion 1: a table of every catalog entry asserts its privilege is
  `member` or `admin`; a catalog built with a zero privilege fails, naming
  the operation.
- Every sentinel's wire code is pinned by a table whose expected strings are
  R-6bi5's literals, and a test fails when a sentinel has no code.
- A JSON round trip of every request and response type is lossless, and
  decoding `{"actor":"someone"}` into the start request leaves it equal to
  its zero value (spec criterion 2's type half).
- `internal/ops` imports only `internal/identity` and the standard library
  (checked by `go list -deps`).
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Let the privilege's zero value mean `member` → the zero-privilege row
   passes construction.
2. Add an `actor` field to the start request → the decode row fails.
3. Accept uppercase hex handles → the handle row fails.

**2.0 hook:** lands hook 9's ops-seam half. **Prior art:** the
OpenSubsonic specification is consulted for its error-code shape and
declined as a model: its codes are numeric and per-endpoint, while this
seam's codes are the spec's seven strings shared by every operation.

**Deliverable count: 2** (the interface and wire types, the codes and the
catalog).

---

## Task 3: `internal/ops/local`: the runner, recovery, cancel, and retention

Model: `opus`. Depends on Tasks 1 and 2. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/ops/local/doc.go`, `local.go`, `jobs.go`,
  `recovery.go`, and their tests, with a `TestMain` that re-executes the
  test binary as a second process

**Outcomes:**

- **The constructor** opens nothing itself. It takes a write-mode handle,
  the backup configuration (flag and remote), a logger that must be non-nil,
  and the transfer's parameters, takes its instance lease, and runs
  recovery.
- **Recovery** (spec table `:560-565`): for every queued or running record,
  try its owner's lease. Held: leave it. Free or absent: mark it interrupted
  in one write, with a `WHERE` naming both states it may hold. Then remove
  every `staging/backup-<instance>.db` whose instance's lease is free or
  absent. Recovery runs at construction and again before every exclusive
  start.
- **The runner.** A start checks availability first (a flag-off backup
  returns unavailable and writes nothing), takes the kind lease (busy
  returns conflict naming the running job's handle, or, with no queued or
  running record of the kind, the lease holders), runs recovery, prunes
  retention, inserts the record as queued with the actor, and runs the job
  on a goroutine under a context the runner cancels. Each transition writes
  the state it expects in its `WHERE`. A transition that matches no row is
  refused and logged at warning with the job's handle and both states.
- **Cancel** writes `cancel_requested`. On a queued record it writes
  canceled directly. On a terminal record it writes nothing and reports the
  state. On a running record owned by any process, the owning runner reads
  the column at each phase boundary and every cancel poll interval while a
  transfer runs, and cancels the job's context; the job then ends canceled.
  A job that finishes before it sees the request keeps its own state, and
  the cancel response says which state the job ended in.
- **Retention** at each start deletes terminal records whose finish time is
  older than 90 days, except the newest succeeded record of each kind.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Record, owner lease | held | left alone, whatever its kind |
| Record, owner lease | free | interrupted; the lease file removed |
| Record, owner lease | absent | interrupted |
| Record, owner lease | unreadable | recovery refused, naming the path; the constructor fails |
| Record | terminal | unchanged |
| Late terminal update | the record is interrupted | refused, logged at warning; the record stays interrupted |
| Start | flag off | unavailable, reason "backup integration disabled"; no record, no lease |
| Start | kind lease busy, a queued or running record | conflict naming its handle |
| Start | kind lease busy, no such record | conflict naming the holders ("a previous run's transfer is still running") |
| Cancel | unknown handle | not found |
| Cancel | malformed handle | invalid request |
| Cancel | queued | canceled |
| Cancel | running | `cancel_requested` set; the owner ends it canceled |
| Cancel | terminal | no write; reports the state |
| Status | unknown handle | not found |
| Retention | no record past 90 days | nothing deleted |
| Retention | the only succeeded record of a kind is past 90 days | kept |
| Staging | a backup stage whose instance lease is held | kept |
| Staging | a backup stage whose instance lease is free or absent | removed |
| Staging | any other name | untouched |

**Acceptance criteria:**

- Spec criterion 3: a re-executed child process opens `ops/local` over the
  store, starts a job whose transfer blocks, and signals readiness on a
  pipe; the parent opens its own `ops/local`, and the child's record still
  reads running. **Fixture state:** a current store; the child's job
  running, its lease held.
- Spec criterion 4: a queued and a running record planted with an owner
  whose lease file exists unlocked, plus one with no lease file, read
  interrupted after an open, and the unlocked lease file is gone.
  **Fixture state:** three records, all non-terminal, one terminal record
  that stays unchanged.
- Spec criterion 5: the child is killed with SIGKILL mid-job, the parent's
  recovery marks it interrupted, and a terminal update issued afterward
  through the runner's own transition is refused and logged; the record
  stays interrupted. The log line is asserted by message and handle.
- Spec criterion 6: from a second `ops/local`, cancel on the child's running
  job sets `cancel_requested`, and the child's record ends canceled; cancel
  on a queued record ends it canceled; cancel on a succeeded record reports
  succeeded and writes nothing.
- Spec criterion 7: a second start while the child's backup runs returns
  conflict naming the child's handle.
- Spec criterion 8, the runner's half: flag off, the catalog reports
  backup unavailable with the exact reason, and a start writes no record and
  leaves `leases/` holding only the instance lease; flag on, the catalog
  reports available.
- Spec criterion 2's record half: the start's actor parameter is what the
  record's `actor` column holds.
- Spec criterion 17: three terminal records, one succeeded at 100 days, one
  failed at 100 days, one succeeded at 10 days, planted through the write
  method: the next start deletes only the 100-day failed record. A second
  fixture with one succeeded record at 100 days keeps it.
- The staging rows, each with a planted stage and a held, unlocked, or
  missing lease.
- `bash scripts/check.sh` green.

The job body in this task's tests is a test-supplied function. Task 4
supplies the backup body; until then no production caller starts a job, and
the exported-unused leg warns.

**Named mutations this task reports:**

1. Insert the record before taking the instance lease → report whether a
   concurrent recovery can mark it interrupted, or name it reviewer-read.
2. Mark interrupted without the state guard → the terminal record in
   criterion 4's fixture changes.
3. Drop the `WHERE` state guard on a terminal transition → criterion 5's
   late update overwrites interrupted.
4. Cancel a running job by writing canceled directly → criterion 6's child
   record reads canceled while its job still runs (report the transfer's
   state at that moment).
5. Retain by count instead of by the newest succeeded per kind → the
   single-record fixture deletes it.
6. Read the conflict's running handle outside the kind lease check →
   report the race, or name it reviewer-read.
7. Poll `cancel_requested` only at phase boundaries → criterion 6's child
   ends succeeded instead of canceled.

**Prior art:** none adopted. A job queue library would own the table this
spec already designed (`0002_jobs.sql`).

**Deliverable count: 4** (the constructor with recovery, the runner and its
transitions, cancel, retention). **Pre-named split:** if the implementer
reports that the four cannot land in one green gate, cancel and retention
move to their own task between this one and Task 4.

---

## Task 4: `internal/backup`: the guards, the transport, and the job

Model: `opus`. Depends on Tasks 1 and 3. **Neutral on 2.0 hooks** (hook 1's
file hash drives backup change detection later; ruling 2 defers it to rung
7).

**Files:**
- Create: `internal/backup/doc.go`, `backup.go`, `guard.go`, `rclone.go`,
  and their tests, with a recording stand-in built from the test binary

**Outcomes:**

- **The job body** runs the guards, syncs the library tree, takes the
  backup copy (Task 1), sends it, and removes the stage. Each phase is a
  progress point where the runner reads `cancel_requested`.
- **The guards** (spec `:638-649`), against the skeleton through unexported
  queries: a missing library root aborts. With catalog tracks, a tree whose
  regular-file count falls below half the track count aborts, and rclone
  runs with `--max-delete` equal to the track count. With zero tracks, an
  empty root aborts, and rclone runs with `--max-delete 0` (R-6bi7). A
  guard abort writes failed with a guard code and never starts rclone.
- **The transport** runs `rclone sync <root> <remote>/library --backup-dir
  <remote>/archive/<stamp> --max-delete N`, then `rclone copyto <stage>
  <remote>/snapshots/<stamp>.db` with `-M --metadata-set run_id=<handle>
  --metadata-set sync_completed=<RFC 3339 UTC>`. A probe (`6bi-p2/`)
  measured both metadata keys reading back through `rclone lsjson -M` on a
  `local` remote. Each invocation runs under R-6bi12's process shape. The
  rclone binary path is a parameter the caller resolves with
  `exec.LookPath`. A nonzero exit writes failed with failure code
  `rclone-exit-<N>` and the last stderr line, bounded to 512 bytes, as the
  message.
- **The layout** under the remote is `library/`, `archive/<stamp>/`, and
  `snapshots/<stamp>.db`, one stamp per run, taken at the job's start.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| rclone | not on `PATH` | failed, code `rclone-missing`, before any guard |
| Library root | absent | failed, code `library-missing`, no invocation |
| Library root | a regular file, not a directory | failed, code `library-missing` |
| Library root | unreadable | failed, code `library-unreadable`, naming the error |
| Zero tracks | empty root | failed, code `library-empty`, no invocation |
| Zero tracks | non-empty root | sync with `--max-delete 0` |
| Tracks | file count below half the track count | failed, code `library-short`, naming both counts, no invocation |
| Tracks | at or above half | sync with `--max-delete` equal to the track count |
| Sync | exit 0 | continue |
| Sync | exit 7 on `--max-delete` | failed, `rclone-exit-7` |
| Sync | any other nonzero | failed, `rclone-exit-<N>` |
| Sync | canceled | canceled; the child gone by `WaitDelay` |
| Copy | refused (Task 1's states) | failed, code `snapshot-failed`, naming the reason |
| Send | nonzero | failed, `rclone-exit-<N>`; the stage removed |
| Stage removal | fails | logged; the job's state stands |

**Acceptance criteria:**

- Spec criterion 10: a run against a real `rclone` with a `local` remote in
  a temporary directory leaves every library file under `library/`, one
  object under `snapshots/`, and that object passes `integrity_check` and
  reads the binary's version. **Fixture state:** a store with two tracks
  and a library root holding their two files.
- Spec criterion 11: after a first run, one library file is deleted and a
  second run leaves it under `archive/<stamp>/`, not deleted. **Fixture
  state:** a populated catalog, so `--max-delete` is the track count.
- Spec criterion 12: a missing root, an empty root with zero catalog
  tracks, and a root holding one file under a four-track catalog each end
  failed with their code, and the stand-in records no invocation.
- Spec criterion 13 (R-6bi7): zero catalog tracks, a remote `library/`
  holding two files, and a root holding one of them: the run ends failed
  with `rclone-exit-7`, and the remote still holds both files with nothing
  under `archive/`. The plan records that `--max-delete 0` is what
  refuses.
- Spec criterion 14: a stand-in exiting 7 ends the job failed with
  `rclone-exit-7` in `failure_code`.
- Spec criterion 9's job half: the object sent holds every row committed
  before the job began (Task 1's copy proof, driven through the job).
- R-6bi12's orphan row: a re-executed child runs the job against a stand-in
  that sleeps; the child is SIGKILLed; while the stand-in lives, a second
  start conflicts naming holders; after the stand-in exits, a start
  succeeds. **Fixture state:** a flag-on configuration and a sleeping
  stand-in.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Count directories as files → report whether the short-root row passes.
2. Run the guards after the sync → criterion 12's stand-in records an
   invocation.
3. Omit `--max-delete` with zero tracks → criterion 13's remote loses a
   file to `archive/`.
4. Take the copy before the sync → report whether criterion 10's object
   still passes; name the ordering as reviewer-read if it does.
5. Drop `Setpgid` → report whether any test fires; the terminal SIGINT
   case is Task 6's.
6. Omit `ExtraFiles` → the orphan row's second start succeeds while the
   stand-in lives.
7. Leave `WaitDelay` unset → Task 6's criterion 20 leaves a stand-in (report
   here if any test fires).

**Prior art:** rclone as a probed subprocess, adopted (spec). The salvage
tier's missing-root guard is adopted in reasoning, and its top-level-entry
check is replaced by the count guard (`_salvage/internal/backup/backup.go:22-34`);
nothing is imported.

**Deliverable count: 4** (the guards, the transport, the job body, the
layout with its stamp).

---

## Task 5: Pruning: remote snapshots, the archive, and local copies

Model: `sonnet`. Depends on Task 4. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/backup/prune.go`, `internal/backup/prune_test.go`
- Modify: `internal/backup/backup.go` (the prune phase after a successful
  send)

**Outcomes:**

- **Remote snapshots** (ruling 10): after a successful send, list
  `snapshots/` with `rclone lsjson`, keep the 14 newest by stamp, and delete
  the rest with `rclone deletefile`. The run's own snapshot is the newest,
  so it is never pruned. A name that does not parse as `<stamp>.db` is left
  alone.
- **The archive:** delete every `archive/<stamp>/` directory whose stamp is
  older than 90 days with `rclone purge`. A directory name that does not
  parse as a stamp is left alone.
- **Local copies** (ruling 7): delete every `pre-migrate-v<N>.db` and
  `pre-restore-<stamp>.db` in `snapshots/` whose modification time is
  earlier than this run's start. A copy written after the job began was not
  captured by it and stays.
- **A prune failure** is logged and recorded in the job's message, and the
  job still ends succeeded, since the backup itself landed.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Remote snapshots | 14 or fewer | nothing deleted |
| Remote snapshots | more than 14 | the oldest deleted; the newest 14 kept |
| Remote snapshots | a name that is not a stamp | untouched |
| Archive | a stamp older than 90 days | purged |
| Archive | a stamp within 90 days | kept |
| Archive | a name that is not a stamp | untouched |
| Local copies | older than the run's start | deleted |
| Local copies | newer than the run's start | kept |
| Local copies | any other name | untouched |
| Any prune step | rclone or `os.Remove` fails | logged; the job succeeded with the failure named in its message |
| The run | failed or canceled before the send | no prune of any kind |

**Acceptance criteria:**

- Spec criterion 15: a real `local` remote pre-seeded with 15 stamped
  snapshots, an unparseable `notes.db`, two archive directories at 100 and
  10 days, and one named `keep-me`: after a successful run the remote holds
  exactly the 14 newest snapshots including the run's own, `notes.db`, the
  10-day archive, and `keep-me`.
- Spec criterion 16: a leaf holding `pre-migrate-v1.db` and
  `pre-restore-x.db` dated before the run and a `pre-migrate-v2.db` dated
  after its start: after a successful run only the last survives, and an
  operator file `notes.txt` is untouched.
- A failed run (stand-in exit 7) leaves every remote and local file as it
  found them.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Keep by count before adding the run's snapshot → criterion 15 keeps 15.
2. Prune on a failed run → the failed-run row loses a file.
3. Delete local copies by name pattern `*.db` → `notes.txt` survives but
   report whether an unrelated `.db` fixture is lost.
4. Compare archive stamps as time without UTC → report the row that fires.

**Deliverable count: 3** (remote snapshot pruning, archive pruning, local
deletion).

---

## Task 6: The backup class, the supersessions, and the `backup` verb

Model: `sonnet`. Depends on Tasks 3, 4, and 5. **Neutral on 2.0 hooks.**

**Files:**
- Create: `cmd/dubplate-server/backup.go`, `cmd/dubplate-server/backup_test.go`
- Modify: `cmd/dubplate-server/main.go` (one registration line after
  `registerStoreConfig`, one `AddCommand` entry, and R-6bi10's exit branch),
  `internal/config/keys.go` (remove the "backup targets" reserved class),
  and the tests that pin the reserved-class list

**Outcomes:**

- **The integration class** `backup`: flag `DUBPLATE_BACKUP_ENABLED`,
  required `DUBPLATE_BACKUP_REMOTE`, registered from `backup.go`.
- **The supersessions,** each through `AddSuperseded` with spec criterion
  19's message verbatim: the five R2 names read "superseded by
  DUBPLATE_BACKUP_REMOTE; configure the remote in rclone's
  RCLONE_CONFIG_<NAME>_* environment form", and `DUBPLATE_BACKUP_STAGING`
  reads "superseded: backup staging now lives in the store leaf under
  DUBPLATE_STATE_DIR". The reserved class leaves `keys.go` in the same
  commit, since a name claimed by both families panics at registration
  (`registry.go`, `claim`).
- **The verb** `dubplate-server backup` loads the wall, opens the store in
  write mode (an absent store refuses and creates nothing), builds
  `ops/local`, starts a backup as a local operator, polls its status, and
  exits by Ruling 3. It prints the job's handle when it starts and its
  terminal state when it ends.
- **Signals** by R-6bi11: SIGINT, SIGTERM, and SIGHUP each call the cancel
  operation on a fresh deadline context, and the verb keeps polling until
  the record is terminal.

**States each check meets, and its report:**

| Check | State | Report | Exit |
| --- | --- | --- | --- |
| Wall | a retired name set | the load error naming its message | 1 |
| Wall | flag set, remote unset | the load error naming the remote key | 1 |
| Wall | flag unparseable | the load error naming the flag | 1 |
| Store | absent | "no store ... start serve first"; nothing created | 1 |
| Store | any refusal of the write mode | its reason | 1 |
| Start | flag off | unavailable, reason "backup integration disabled" | 69 |
| Start | conflict | the running job's handle or the holders | 75 |
| Job | succeeded | the handle and "succeeded" | 0 |
| Job | failed | the handle, the code, the message | 1 |
| Job | canceled | the handle and "canceled" | 130 |
| Job | interrupted (unreachable in-process, since the verb's own runner holds the lease; kept so the state set is closed) | the handle and "interrupted" | 1 |

**Acceptance criteria:**

- Spec criterion 18: `backup` against a state directory with no leaf exits
  1 naming "start serve first", and the directory listing is identical
  before and after.
- Spec criterion 19: setting each of the six retired names fails the load
  with its message verbatim, each by its own table row. The template
  classification test (`templatekeys_test.go`) still passes, now reading
  the six names as superseded.
- Spec criterion 20: the verb runs as a re-executed process against a
  sleeping stand-in; the test sends SIGINT to the verb's pid alone; the job
  ends canceled, no stand-in process remains after `WaitDelay`, and the
  verb exits 130. A second row sends SIGINT to the verb's process group, as
  a terminal does, and asserts the same outcome, which fails without
  `Setpgid`.
- Each exit code maps to its outcome as the verb's help states: a table of
  the five outcomes, each driven by its real trigger (flag off, a held kind
  lease, stand-in exits 0 and 7, SIGINT), asserts the code and that the help
  text names it.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Start the job on the verb's own context → criterion 20 ends failed, not
   canceled.
2. Cancel on the done context → the cancel write is refused for its
   deadline (report the error), and the job ends by rclone's own exit.
3. Ignore SIGHUP → report the stand-in's fate when the verb is killed by
   SIGHUP's default action.
4. Map every error to exit 1 → the exit-code table fails.
5. Leave the reserved class in `keys.go` → registration panics (report the
   test that fires).

**Deliverable count: 4** (the class, the supersessions with the reserved
class's removal, the verb with its exit codes, signal handling).

---

## Task 7: `doctor`'s backup rows

Model: `sonnet`. Depends on Tasks 1 and 6. **Neutral on 2.0 hooks.**

**Files:**
- Modify: `cmd/dubplate-server/doctor.go`, `doctor_test.go`,
  `cmd/dubplate-server/store.go`, `store_test.go`, `backup.go`,
  `backup_test.go`

**Outcomes:**

- **The remote rows**, only with the flag on: the remote name is the part of
  `DUBPLATE_BACKUP_REMOTE` before the first colon. A value with no colon, an
  empty name, or a name holding any character but a letter, digit, or
  underscore is refused. `RCLONE_CONFIG_<NAME>_TYPE` (uppercased name) must
  be set. For `crypt`, `RCLONE_CONFIG_<NAME>_REMOTE` must name a remote whose
  own `_TYPE` is set. The features probe runs last (Ruling 2). A non-crypt
  type carries ruling 14's warning, never a failure.
- **The store rows** from R-6bi13: last backup success age or "never",
  the `-wal` size, free space, and the local copy count and size.
- **Every row renders in a golden** from a constructed input, as 6a's store
  row does (R-6a16).

**States each check meets, and its report:**

| Check | State | Row | Exit |
| --- | --- | --- | --- |
| Flag | off | backup: "unavailable: backup integration disabled" | 0 |
| Remote | no colon | "invalid remote: want <name>:<path>" | 1 |
| Remote name | a character outside letters, digits, underscore | names the character and rclone's rule | 1 |
| `_TYPE` | unset | names `RCLONE_CONFIG_<NAME>_TYPE` | 1 |
| Crypt | wrapped remote's `_TYPE` unset | names the wrapped remote and its key | 1 |
| rclone | not on `PATH` | "rclone not found" | 1 |
| Features | `Copy` and `Move` both false | names the backend and rclone's archive requirement | 1 |
| Features | the probe exits nonzero | its last stderr line | 1 |
| Features | the probe times out at 10 s | "not checked: rclone backend features timed out" | 0 |
| Crypt | the type is not `crypt` | the ruling 14 warning appended | 0 |
| Last success | no succeeded backup record | "never" | 0 |
| Last success | a record | its age, rounded to minutes | 0 |
| Store | absent, older, newer, or not a store | "not checked" on the store-fact rows | as the store row |

**Acceptance criteria:**

- Spec criterion 21: one golden line per row state above, each from a
  constructed input, plus three end-to-end rows against a real `rclone`: a
  valid `local` remote, an `http` remote (both features false), and a crypt
  remote whose wrapped remote is undefined.
- `TestDoctorCmd_LineCountFollowsRegistry` names the backup rows' count
  as constants, one set for the flag off and one for on, and fails if a row
  prints twice.
- With the flag off, `doctor` runs no subprocess (the stand-in records no
  invocation).
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Read the name after the colon → the invalid-name golden fails.
2. Skip the wrapped remote check for crypt → its end-to-end row passes.
3. Run the features probe with the flag off → the no-subprocess row fails.
4. Treat a probe timeout as a failure → the timeout golden exits 1.

**Deliverable count: 3** (the remote rows, the store-fact rows, the
goldens and line count).

---

## Task 8: The pass close, and the 6b-i to 6b-ii handoff

Model: `sonnet`. Depends on Tasks 0 to 7 and the simplifier round.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/superpowers/research/handoffs/draft-STATUS-6b-i.md`
- Create: `docs/superpowers/research/handoffs/draft-HISTORY-6b-i.md`

**Outcomes:**

- **The HISTORY draft** carries what landed per task with commit ids, the
  gate evidence (the new packages' test wall times with `-count=1`, and the
  gate's wall time with the rclone tests), the resolutions R-6bi1 to R-6bi13
  with where each landed, the three rulings and their outcomes, the declared
  coverage gaps, both budgets, and "what a later pass would be wrong to
  rediscover," which carries R-6bi7's measurement and R-6bi12's orphan
  shape at minimum.
- **The errata owed**, paste-ready: the six spec deviations, and the spec's
  "Not verified" rows this pass settled (`--max-delete`, the feature table,
  `os/exec`'s wording, rclone's exit codes).
- **The 6b-ii handoff:** the ops interface and codes the projection
  mirrors; `internal` and `forbidden`, which 6b-ii's contract suite raises
  first; the lease and kind-lease semantics restore must respect; the
  staging rule; the remote layout restore reads.
- **The CLAUDE.md Gate line** for `rclone`, and the runbook lines owed to
  rung 16: `KillMode=mixed`, rclone variables set by hand until the
  template rows land, a crypt remote recommended, the non-empty wrong-mount
  gap, and that a restore before rung 12 loses stars and playlists.

**Acceptance criteria:** every count in the drafts was produced by a
command the report pastes; every commit id resolves; the STATUS draft is at
most 60 lines. No gate change, so `bash scripts/check.sh` green as a
formality.

**Deliverable count: 4** (the HISTORY draft, the errata, the handoff with
the Gate and runbook lines, the STATUS draft).

---

## 6a's carried review items, disposed

- **A migration file carrying its own `COMMIT` is caught only at Commit,
  after its tail ran in autocommit** (`migrate.go:339-349`). **Declined for
  6b-i.** The gate applies every embedded file through the runner against
  scratch stores (`TestSchemaSkeletonSatisfiesInvariants`,
  `TestSchemaMigrationsCarryPopulatedFixtures`), so an embedded `COMMIT`
  fails the gate before merge, and its tail only ever runs against a
  scratch database. Refusing it before the tail runs needs a SQL tokenizer:
  a `CREATE TRIGGER` body legitimately carries `BEGIN ... END`, so a keyword
  scan refuses valid files. The driver exposes no autocommit query
  (`conn.go`, v1.59.0, read 2026-09-24). 6b-i adds no migration. The
  per-file `diff-reviewer` checklist line (6a Task 4) stays the net.
- **A failed snapshot rename leaves staging occupied until serve's
  cleanup** (`snapshot.go:73`). **A criterion in Task 1.**
- **The rebuild marker matches exact bytes only** (`migrate.go:163`). **A
  criterion in Task 1:** a near miss is refused at `readMigrations`, which
  every test calls, so a bad embedded file fails the gate.
- **Serve's unbounded settle loop and unguarded logger**
  (`open.go:192-196`). **Routed to 6d**, as the conductor directed.
  6b-i's own logger parameter refuses nil at construction (Task 3).

---

## The exit gate

`bash scripts/check.sh` green in the worktree and again on master after
the merge, plus every assertion below. Each is a spec 6b-i criterion
(`:990-1033`).

1. Criterion 1: Task 2.
2. Criterion 2: Tasks 2 and 3.
3. Criteria 3 to 8 (the job model and availability): Task 3.
4. Criterion 9 (the snapshot's content): Tasks 1 and 4.
5. Criteria 10 to 14 (the job against real rclone, the guards, the exit
   code): Task 4.
6. Criteria 15 and 16 (pruning): Task 5.
7. Criterion 17 (retention): Task 3.
8. Criteria 18 to 20 (the verb, the supersessions, SIGINT): Task 6.
9. Criterion 21 (`doctor`'s rows): Task 7.
10. Nothing in the pass creates or migrates a store outside a test: the
    production callers of `internal/engine/store` are `doctor`, `check`,
    and `backup`, the last through the write mode.
11. Every symbol the pass exports has a production caller outside its own
    package and tests (the exported-unused leg's report, pasted).

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's
ownership table before the lane launches.

| Surface | Owner | 6b-i's position | Contention |
| --- | --- | --- | --- |
| `internal/engine/store` | rung 6 | Task 1 | none; 5c-ii leaves it untouched |
| `internal/ops`, `internal/ops/local`, `internal/backup` (new) | rung 6 | Tasks 2 to 5 | none |
| `cmd/dubplate-server/main.go` | rung 1 | Task 6: one registration line, one `AddCommand` entry, main's exit branch | **5c-ii Task 5 adds one line in `newProductionRegistry`**; a textual conflict keeps both in landing order; the second lander's rebase is non-fast-forward |
| `cmd/dubplate-server/doctor.go`, `doctor_test.go` | rung 1's skeleton | Task 7 | 5c-ii adds threshold rows through the registry; `TestDoctorCmd_LineCountFollowsRegistry` derives them, and the second lander regenerates goldens in its rebase |
| `cmd/dubplate-server/store.go`, `backup.go` (new) | rung 6 | Tasks 6, 7 | none |
| `internal/config/keys.go` | rung 1b's contract | Task 6 removes the "backup targets" reserved class | 5c-ii leaves `internal/config` untouched; 6c's lane condition is about 5c and 5d, so this edit does not change it |
| `internal/identity` | rung 5 | imported only | 5c-ii edits it; Task 0 reports `Identity`'s fields |
| `scripts/check.sh`, `Makefile`, `web/` | rung 5 | untouched | after 5c-ii merges, the lane's rebase needs `npm ci --prefix web` |
| `CLAUDE.md` Gate section | conductor | Task 8 drafts the `rclone` line | conductor-committed in the main tree |
| The state directory layout | rung 4 | `leases/` joins the store leaf | 6d grants the leaf on the box |
| `docs/` ledgers | conductor | Task 8 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: confirm 6a merged; record the planning spend against the
   rung; create the lane 2 worktree from master; write the lane registry
   row and the contested-surface rows; arm the runaway guard; check AC power.
2. Before Task 1: rule Task 0's blockers and record the three rulings'
   outcomes.
3. At segment A's boundary: STATUS; re-derive the ceiling from Tasks 1 to 3.
4. At segment B's boundary: STATUS; one `code-simplifier` round over Tasks
   1 to 7's accepted Go, pasting `.claude/instructions/simplifier-brief.md`
   verbatim, its review chunked one `diff-reviewer` per package.
5. The merge ritual, with one `diff-reviewer` on the merged result whenever
   the rebase is not a fast-forward, then **one `go-architecture-reader`
   per touched Go package** (`internal/engine/store`, `internal/ops`,
   `internal/ops/local`, `internal/backup`, `cmd/dubplate-server`), no
   plan in context.
6. Record the 6b-i ceiling against the actual, and re-derive 6b-ii's figure
   before 6b-ii's plan is drafted.
7. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, in **three
segments**. Every task runs `dubplate-implementer`, then `diff-reviewer`
(`claude-opus-5-5`), then the gate, with one re-dispatch on `fix` and a
second `fix` as the conductor's decision. Every reviewer dispatch asks for
correctness gaps ranked by consequence, not polish.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1, 2, 3 | STATUS. Re-derive the ceiling. |
| **B** | 4, 5, 6, 7 | STATUS. The simplifier round. |
| **C** | 8 | The close drafts, the merge ritual, the architecture reads, the ledger fold. |

Each boundary sits on a task's own green committed gate.

**Models.** `sonnet` by default. **Three upshifts to `opus`, all
unconditional**, dispatched as `claude-opus-5-5`: Task 1 (the lease
primitive, whose wrong variants pass every single-descriptor test, and the
backup copy, the data path hook 10 names); Task 3 (cross-process recovery
and cancel, whose races a single-process test passes); and Task 4 (the
guards and `--max-delete`, the data-loss path, and R-6bi12's orphan shape).
Task 0 runs at `opus` in fresh context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1.
- Tasks 1 and 2 → Task 3 (the write surface and leases; the interface).
- Tasks 1 and 3 → Task 4 (the copy; the runner).
- Task 4 → Task 5 (the job the prune phase joins).
- Tasks 3, 4, and 5 → Task 6 (the verb runs the whole job).
- Tasks 1 and 6 → Task 7 (the report fields; the class).
- Tasks 0 to 7 → Task 8.
- **Independent by files:** Tasks 1 and 2. They still run sequentially,
  because one tree runs one executor and the module gate runs over `./...`.

**Where this pass splits, if it must:** at the Task 5 / Task 6 boundary, if
segment A's re-derivation projects past the ceiling. Tasks 6 and 7 then
open a pass 6b-i-2, because the verb and `doctor` rows are what give the
exports their production caller; 6b-i-1 would close with its exports
warned by the exported-unused leg, which the conductor records. Inside the
pass, Task 3 carries its pre-named split.

**Budget:** 7.6M execution (Ruling 1), flag 6.08M, planning recorded
separately, STATUS at each boundary, spend scored at close against the
ceiling and against zero attended interactions after approval.

---

## Prior art consulted

Scored on CLAUDE.md's four criteria; the spec's table (`:895-915`) carries
the evidence for the rows it scored.

| Entry | Verdict for 6b-i |
| --- | --- |
| rclone v1.74.3, subprocess | **Adopt** (spec). This plan adds R-6bi7's and R-6bi9's measurements. |
| `librclone` | **Decline** (spec): rclone's whole dependency tree for one verb. |
| SQLite `VACUUM INTO` | **Adopt** for the backup copy (Task 1); the driver's backup API stays the recorded alternative. |
| `internal/atomicfile` | **Not used** at 6b-i: no record file is written; the intent record is 6b-ii's. |
| `_salvage/internal/backup` | **Adopt the missing-root guard in reasoning**; the count guard replaces its top-level check; import nothing. |
| A Go job-queue library | **Decline**: it would own a table the spec already designed, and none speaks cross-process `flock` leases. Not measured further, since no candidate was in the spec's list. |
| OpenSubsonic error codes | **Consulted, declined as a model** (Task 2). |
| Navidrome, gonic, taglib, mewkiz/flac, Picard, chromaprint, goose, golang-migrate | No 6b-i surface. |

---

## Not verified in this revision

- Whether every rclone backend accepts `--metadata-set`; a `local` remote
  on tmpfs did. A backend that drops it loses only the metadata, since the
  stamp names the snapshot and the job record holds the run.
- Whether `rclone backend features` contacts the provider for every
  backend; Ruling 2's 10 s bound covers the case where it does.
- rclone's behavior on SIGINT mid-sync against a real remote; the gate's
  SIGINT row runs the stand-in, and the real binary's exit is observed only
  when it exits within `WaitDelay`.
- The 6a close-side spend (about 2.1M) is STATUS's re-derived projection,
  not a recorded actual; segment A re-derives this ceiling from it.
