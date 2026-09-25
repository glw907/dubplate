# Rung 6, Pass 6b-i (The Ops Seam and Backup) Implementation Plan

Status: **revision 2.1, 2026-09-24; folded twice, awaiting approval.**
Revision 1 (`b36871c`) took three lens reviews (contract C1 to C29,
mechanics M1 to M14 and F1, risk R1 to R16), folded into revision 2
(`5782d15`). Revision 2.1 folds the fresh-context verification read (V1 to
V10, `docs/superpowers/research/2026-09-24-rung-6b-i-plan-fold-verification.md`)
and the store findings of the 6a architecture read that shape a signature
6b-i exports or binds to (`docs/superpowers/research/2026-09-24-rung-6a-architecture-reads.md`,
items 1 to 4). The fold record
(`docs/superpowers/research/2026-09-24-rung-6b-i-plan-fold-record.md`)
carries one disposition per finding in both folds. Four rulings for Geoff
are open ("Rulings for Geoff").

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

**Code pointers are resolved at `0b6da0f` on master.** 6a merged at
`8736528`; its rebase rewrote the branch, so revision 2's base `aa3b13e`
survives as `0b6da0f` with an identical store tree (`git diff aa3b13e
0b6da0f -- internal/engine/store` is empty). Revision 1 read `dedc1b7`;
since then the simplifier commit `f85e7be` turned the leaf-first free
functions into methods on `leaf` (`createLeaf` is now `create`, `statLeaf`
is `stat`, `probeLeaf` is `probe`, and `probeTarget`, `classifyLockless`,
`checkIntent`, and `besideCopy` are methods too), and `0b6da0f` renamed the
comment text to match. Every pointer below uses the new names and reads the
committed tree (`git show 0b6da0f:<path>`). Task 0 re-resolves every
pointer and grades a moved pointer by content.

**What this pass is.** Pass 6b-i builds the ops seam and backup, the first
long-running operation. It creates `internal/ops` (the interface, the wire
types, the error codes, the privileges, and the catalog), `internal/ops/local`
(the job runner, leases, recovery, and cross-process cancel), and
`internal/backup` (the guards, the rclone transport, the snapshot send, and
pruning). It exports the store's write mode, write method, and a lease
primitive to their first callers, and, on Ruling 4's yes, serve mode's open
to the tests that need a current store. It lands the backup integration class, the
`backup` verb, and `doctor`'s backup rows. The spec's pass table assigns
exactly this set (spec `:148-151`, criteria `:990-1033`).

**What this pass is not.** No HTTP projection, no authorizer, and no
contract suite: those are 6b-ii. No `restore` verb and no intent-record
format: those are 6b-ii. No settings table and no seed rule: 6b-i's
constants stay constants until 6c re-homes them. No `serve` wiring and no
roster import: those are 6d. **Nothing in production creates or migrates a
store in this pass**; a test is the only caller that creates one. `backup`
opens in write mode, which refuses an absent
store, so on every box `backup` refuses until 6d. No `systemd/`,
`scripts/deploy.sh`, or `env/` change, no backup timer, and **no deploy of
any kind.** A task that believes it needs any of these stops and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the
> pass-execute chain per task, sequential, in the four segments "Execution
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

Four yes-or-no questions. The spec's fourteen rulings and 6a's four stand
and are not reopened, with one qualification: approving this plan also
approves the spec deviations listed below, and two of them, R-6bi15 and
R-6bi16, qualify rulings 10 and 7 in ways an adopter can see (snapshots
accumulate past 14 while the clock is behind, and a backup reference must
be removed by hand to start a fresh store).

**1. Does 6b-i run under an 8.6M execution ceiling, with planning recorded
separately?** Recommendation: yes. Reason: the folds split Task 1 in two,
added the store-surface Task 1S, and grew Tasks 3 to 5, which brings the
projection to 6.45M ("Token ceiling"). The spec's 5.9M row is a total that
includes 0.95M of planning and a 0.6M close, while 6a's projected close side
is about 2.1M. What each answer builds: yes sets the flag at 6.88M, 0.43M,
about one task, above the projection. No holds 5.9M as a total, which sits below the
projection, so the flag fires on plan and the conductor asks mid-pass.

**2. Does `doctor` ask rclone itself whether the backup remote supports
server-side move or copy, running `rclone backend features
--low-level-retries 1` under a 10 s bound?** Recommendation: yes. Reason:
rclone's own probe answers for the installed binary and every backend,
crypt included, while a copied feature table goes stale and cannot see
through a crypt remote (R-6bi9). A probe measured a B2 remote with bad
credentials failing in 2.1 s with `401 bad_auth_token` under
`--low-level-retries 1`, against 13.7 s at the default of 10, so the flag
turns the probe's most useful catch into a named failure rather than a
timeout. What each answer builds: yes runs one bounded subprocess per
`doctor` run when the flag is on, which may contact the provider for some
backends; a timeout reads "not checked" and never fails `doctor`. No builds
a static table keyed on `RCLONE_CONFIG_<NAME>_TYPE` from rclone's published
"Optional Features" table, with a crypt remote resolved to the type it
wraps; it makes no network call and reports "unknown backend" for any type
missing from the table.

**3. Does the `backup` verb exit with the codes below?** 0 succeeded, 1
failed or refused before start, 69 unavailable, 75 conflict, and 130
canceled by any means. Recommendation: yes. Reason: every existing verb
exits 1 on any error (`cmd/dubplate-server/main.go:54`), so 1 keeps that
meaning, and the three others take `sysexits.h`'s `EX_UNAVAILABLE` (69,
"service unavailable") and `EX_TEMPFAIL` (75, "temp failure; user is
invited to retry") and the shell's 128-plus-SIGINT convention (130). Rung
16's timer and rung 9's ping read them. 130 covers a cancel from SIGINT,
SIGTERM, SIGHUP, or another process's cancel operation alike, since the job
record names who canceled and one canceled code keeps the timer simple; the
help text says "canceled by any means". What each answer builds: yes fixes
these in the verb's help and in main's exit path. No uses 2, 3, and 4 for
unavailable, conflict, and canceled, which the shell and cobra do not
reserve but which carry no convention an operator recognizes.

**4. Does Task 1A export serve mode's open now, ahead of 6d's `serve`, so
the tests of `internal/ops/local`, `internal/backup`, and
`cmd/dubplate-server` can build a current store?** Recommendation: yes.
Reason: every criterion from Task 3 on names a current store as its
fixture, and the store's only export today is `Inspect`. Write mode refuses
an absent store, and a package outside the store cannot reach the
unexported create or migration runner, so a test-support package such as
`storetest` would need this same export to work. Serve mode's open is the
surface 6d exports anyway; exporting it one pass early adds no new API,
and its only 6b-i callers are tests. What each answer builds: yes exports
the open and its options type with a nil-logger refusal (6a's carried
item), names it as exit gate 11's one test-only export, and leaves the
settle-loop bound with 6d. No keeps the open unexported and has the store's
own test write a current leaf template under
`internal/engine/store/testdata/leaf/`, which a store test fails on when
the template drifts from head, and which each consuming test copies into a
temporary directory and chmods to the leaf's modes, since git keeps none of
them. On no, the conductor amends Task 1A's Files before dispatch to add
the template, its regeneration test, and the copy-and-chmod helper, and
names the helper's home, which the three consuming packages' tests
import.

### Spec deviations that take effect on approval

These change the ratified spec's meaning. Approving this plan approves them;
the close drafts the erratum text (Task 8).

- R-6bi2: the lease primitive lives in `internal/engine/store`, and
  recovery policy lives in `ops/local`. The spec's package table names
  leases in both rows (`:200` and `:202`).
- R-6bi5: 6b-i declares five of the seven sentinels and the local actor
  kind only. `forbidden`, `internal`, and the network actor built from
  `identity.Identity` land in 6b-ii with the projection that first raises
  or builds them, so `internal/ops` does not import `internal/identity` at
  6b-i.
- R-6bi6: the backup's store copy is staged, verified, and sent from
  `staging/`, and removed after the send. It is never renamed into
  `snapshots/`. The spec's "renames it, fsyncs the directory by hand"
  (`:624-626`) holds for the pre-migration copy only.
- R-6bi6: a backup start, holding the backup kind lease, removes every
  `staging/backup-*.db`. Erratum E7 says write mode never removes staging;
  the kind lease proves no backup, orphaned transfer included, is live, so
  this is the one write-mode removal, and it touches no other name.
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
- R-6bi15: the prune keeps the run's own snapshot by name and the 13 newest
  of the rest, and skips every prune step when the run's stamp is not
  strictly the newest on the remote. Ruling 10's "keep the newest 14"
  holds on a sane clock.
- R-6bi16: a successful backup leaves one reference file,
  `snapshots/backup-<stamp>.ref`, which the lost-database guard (erratum
  E11) reads as a copy beside the database. Ruling 7's deletion of local
  copies stands; the reference keeps E11's reach after it.

---

## Primary inputs

Read these before Task 1A, in this order.

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
3. **`internal/engine/store` as shipped at `0b6da0f`**: `leaf.go`,
   `lock.go`, `conn.go`, `write.go`, `migrate.go`, `snapshot.go`,
   `open.go`, `inspect.go`, and `migrations/0002_jobs.sql`.
4. **This plan's quoted behaviors.** Each is quoted from `go doc` at
   `go1.27.1 linux/amd64`, a man page, rclone v1.74.3's installed manual
   (`man rclone`), or a probe run on 2026-09-24 under
   `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/89b265c3-0b9e-4d9c-b998-df7b9970c32a/scratchpad/`
   (`6bi-p1/`, `6bi-p2/`), the mechanics review's probes under
   `/tmp/claude-1000/rv6bi/`, and the fold's under
   `/tmp/claude-1000/fold6bi/`. The probes may be gone; each task re-proves what
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

The four rulings above, at the plan-approval gate. Nothing else is
expected. What would become one: a new module dependency (this pass adds
none); a threshold default changed on a task's own judgment; any change to
what the box runs.

---

## Token ceiling: 8.6M execution for 6b-i (Ruling 1)

The basis is 6a's recorded spend (`docs/STATUS.md`, "Next action" 1, and
the 6a HISTORY draft's "Both budgets"): about 1.66M for Tasks 0 to 3 and
about 1.35M for Tasks 4 to 7, zero fix rounds in segment B, and a
re-derived total of about 5.1M. So Tasks 0 to 7 cost about 3.0M, and the
close side (the tidy fix, the simplifier round, segment C, the merge
ritual, and two architecture reads) is projected at about 2.1M. Planning is
recorded against the rung and outside this ceiling, as 6a's Ruling 1 set.

**The arithmetic.**

1. Task 0 at `opus` in fresh context: **0.20M**.
2. Eight implementation tasks (1A, 1B, and 2 to 7) at 6a's measured 0.40M
   ((3.0M − 0.2M) / 7): **3.20M**. Revision 1 carried seven; the first fold
   split Task 1 because it passed four deliverables.
3. Task 1S, the store-surface reshape the architecture read routes here
   (second fold): a `sonnet` task of mechanical signature changes inside
   one package, with its existing tests to lean on, priced below the
   measured average at **0.25M**.
4. Two costs 6a did not carry: the first subprocess tests against a real
   `rclone`, with a stand-in in two signal modes, and the re-executed test
   binary that makes a second process real (Tasks 3, 4, and 6). Add
   **0.35M**.
5. One fix round, since 6b-i's cross-process paths are new: **0.25M**.
6. The close side at 6a's projection, plus three more architecture reads
   (`internal/ops`, `internal/ops/local`, `internal/backup`, beside
   `internal/engine/store` and `cmd/dubplate-server`): **2.20M**.
7. **The projection is 6.45M.**

A ceiling of **8.6M** puts the 80% flag at **6.88M**, 0.43M above the
projection, so the flag marks an overrun rather than firing on plan. The
pre-named split in Task 3 adds about 0.40M and brings the projection to
6.85M, just under the flag; one fix round beyond the one priced crosses
it, and the conductor then asks at that boundary under the 80% rule. The
spec's row reads 5.9M: 0.95M planning, 0.2M Task 0, nine tasks at 0.46M,
and a 0.6M close. Re-derive at the segment A boundary from Tasks 1S to 2's
reports, and again at segment B's from Task 3's, where the pre-named split
shows.

Checkpoint: STATUS at every segment boundary (after Tasks 2, 5, and 7) and
at the close. Every task reports its spend, its mutation outcomes, and
its measured gate wall time.

Deliverable count: thirty-three across Tasks 1S to 8, plus Task 0's
report. Tasks 3, 4, 5, 6, and 8 carry four each (20); Tasks 1A, 1B, and 7
carry three each (9); Tasks 1S and 2 carry two each (4). No task passes
four.

---

## Global constraints

- **Lane 2, a git worktree.** The conductor creates it from master after 6a
  merges, before Task 0, and writes the lane registry row. Every task, gate,
  and commit runs inside the worktree.
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`, and
  no task ssh's to the box.
- **Package territory, stated as a wall.** This pass creates `internal/ops`,
  `internal/ops/local`, and `internal/backup` (every file and `testdata/`).
  It edits `internal/engine/store` (Tasks 1S, 1A, and 1B only). Task 4 edits
  `internal/ops/local` to bind the backup body (R-6bi17). It creates
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
  the standard library only at 6b-i (its `internal/identity` import lands
  with 6b-ii's network actor, R-6bi4); `internal/ops/local`
  imports `ops`, `store`, and `backup`; `internal/backup` imports `store`
  and the standard library. The engine has no HTTP knowledge and reads no
  environment. `internal/config` imports none of the new packages.
- **The store's rules carry over.** No process that holds a store
  connection opens the database, `-wal`, or `-shm` file through `os`
  (R-6a4); every size comes from `os.Stat`. Every DSN comes from the one
  builder (R-6a17). Times are integer UTC microseconds bound as `int64`
  (R-6a8). Every store write goes through the write method, whose callback
  receives the store-owned transaction handle (R-6a10). New code matches
  no raw SQLite result code; it reuses the store's existing classifiers. A
  task that finds it must match one lands a single `primaryCode(err) (int,
  bool)` helper with `modernc.org/sqlite/lib`'s `sqlite3.SQLITE_*`
  constants first, and routes every existing match site through it
  (architecture read item 8); otherwise that cleanup stays with the
  architecture cleanup pass.
- **Nothing is exported ahead of a cross-package caller.** Within the pass,
  an export whose production caller lands in a later task is permitted, and
  the exported-unused leg warns meanwhile; at the close every export has a
  production caller (exit gate 11), with two named exceptions: serve mode's
  open, whose 6b-i callers are tests (Ruling 4), and the catalog operation,
  which spec criterion 8 requires at 6b-i and whose first production caller
  is 6b-ii's projection. Surfaces with no 6b-i criterion and no 6b-i caller
  (`forbidden`, `internal`, and the network actor) wait for 6b-ii.
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
  or a process-wide signal handler. A test that signals a re-executed
  process starts it with `SysProcAttr.Setpgid`, so a group signal never
  reaches `go test` itself.
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
| A job's `state` | Its owning runner (queued to running to a terminal state); the cancel operation (queued to canceled only); recovery (queued or running to interrupted, only when the owner's lease is free or absent). | Every state write names the state it expects in its `WHERE`, so a terminal state never changes; a late terminal update to an interrupted record matches no row and is refused and logged (criterion 5). |
| A job's `finished_at` | Every transition into a terminal state, in the same statement as `state`: the runner's, recovery's, and cancel's of a queued record. | No other statement writes it; retention and `doctor` read it, and a terminal record with it NULL is a defect a Task 3 row pins. |
| A job's `started_at` and `progress` | The owning runner alone: `started_at` in the queued-to-running transition, `progress` at each phase the body reports and in the terminal succeeded transition (a prune failure's text, R-6bi15). | Recovery and cancel never write either. |
| A job's `updated_at` | Every statement that writes the record, in the same statement. | None writes the record without it. |
| A job's `cancel_requested` | The cancel operation alone. | The runner reads it and never writes it. |
| A job's `failure_code` and `failure_message` | The owning runner, in the transition to failed. | No request type carries either. |
| Deleting a job record | Retention at each job start, which never removes the newest succeeded record of a kind (Task 3). | No operation deletes a record. |
| A lease file in `leases/` | Its owning runner (never the store handle) creates and holds it, and its close removes it while still holding it; recovery removes another's only after acquiring it, which proves its owner dead. | Recovery never removes a file it could not lock. |
| The inherited descriptors | The kind-lease value hands out close-on-exec duplicates (`F_DUPFD_CLOEXEC`) of the kind lease's and its handle's store-lock descriptors for a transfer child's `ExtraFiles` (R-6bi12). | The caller owns and closes only the duplicates; the store's own descriptors are never exposed, so no caller can release the store lock early. |
| A backup staging file (`staging/backup-<instance>.db`) | The backup job whose instance names it; a backup start, holding the kind lease, removes every `staging/backup-*.db` before its own copy (R-6bi6). | Serve mode's cleanup also removes it, under the exclusive lock (R-6a7). Recovery and every other write-mode path leave staging alone. |
| The backup reference (`snapshots/backup-<stamp>.ref`) | A backup job after its send succeeds, through the store (R-6bi16); it replaces the previous reference. | The local deletion step never touches it; only an operator removes it by hand. |
| The run's stamp | The job's start, from the wall clock in UTC at microsecond resolution (R-6bi8). | Every prune decision checks it against the remote's stamps before deleting anything (R-6bi15). |
| Deleting a remote snapshot or archive directory | The backup job's prune step, after a successful run (Task 5), with every argument built from a stamp that round-trips through the layout (R-6bi14). | The run's own snapshot is never pruned, by name (R-6bi15). |
| Deleting a local pre-migration or pre-restore copy | The backup job's local deletion, after a successful run, only for copies older than the snapshot it sent (ruling 7, Task 5); the migration runner's retention with backup off (6a). | Neither touches a name outside `pre-migrate-v<N>.db` and `pre-restore-<stamp>.db`. |
| `DUBPLATE_BACKUP_ENABLED`, `DUBPLATE_BACKUP_REMOTE` | The environment, through the registry's backup integration class (Task 6). | The store and `internal/backup` read no environment; `cmd/dubplate-server` passes both values. |
| The library root (`DUBPLATE_LIBRARY_ROOT`) | The environment, through the registry key rung 1 declared; `cmd/dubplate-server` reads it and passes it to the constructor (R-6bi18). | `internal/backup` reads no environment; the job resolves the root's symlinks once, and the guard and rclone both use the resolved path. |

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
close so the fewest commits touch the shared files. 6b-i no longer imports
`internal/identity` (R-6bi4), so 5c-ii's edits there do not reach this
pass.

### R-6bi2: the store owns the lease primitive; `ops/local` owns recovery

The spec's package table names leases in the store's row and in
`ops/local`'s. The store already owns every `flock` in the leaf
(`lock.go:84`, `tryLock`), the leaf's paths (`leaf.go:30-38`), and the
holder list from `/proc/locks`. So the store exports one lease primitive:
take an instance lease under `leases/`, take a kind lease
(`leases/kind-<kind>`), and try an instance's lease. `ops/local`
owns what a lease means for a job record: the recovery table and the
conflict answer.

**Ownership, in one sentence:** the lease primitive returns lease values;
the store handle takes no lease itself; `ops/local`'s runner takes, holds,
and releases the instance lease for its own life; and the descriptor method
(R-6bi12) sits on the kind-lease value, which is taken through a handle, so
it can duplicate that handle's store lock as well. `leases/` joins the leaf (R-6a13): serve mode's create
makes it at 0700, and the leaf's completeness check counts it. No
production leaf exists before 6d, so no leaf in the field lacks it.

Five rules close the lease's life cycle.

- **A clean close removes its own lease.** The instance lease's release
  removes `leases/<instance>` while still holding the lock, then closes
  it. Without this every run leaves one file behind forever.
- **A try that wins the lock and then finds the file gone reads free.**
  Two processes can open the same dead lease file; the first locks it,
  removes it, and closes it, and the second's `flock` then succeeds on the
  unlinked inode, so its `os.Remove` answers not-exist (probed, mechanics
  M9). That answer means another recovery already proved the owner dead.
- **No blind sweep.** Recovery tries only the owners its records name, and
  a record is inserted only after its owner's lease is held, so a lease
  file created but not yet locked is never tried. A sweep of every
  `leases/` file would race exactly that window and remove a live
  process's lease. A process that crashes with no open record leaves one
  small file; that residual is accepted.
- **The kind lease's busy error is a sentinel** the store exports, so
  `ops/local` tells a conflict from any other lease error with `errors.Is`.
- **Holder pids are "acquired by", never "held by".** `/proc/locks`
  records the pid that took a `flock`, which after the parent's death is a
  dead pid while an inheriting child holds the lock (probed, mechanics M8),
  and on btrfs the list is empty (`lock.go`, `busyError`'s doc). So no
  answer depends on a pid, and no test asserts one.

`flock(2)` gives the property every in-process test relies on: "If a
process uses open(2) (or similar) to obtain more than one file descriptor
for the same file, these file descriptors are treated independently by
flock(). An attempt to lock the file using one of these file descriptors
may be denied by a lock that the calling process has already placed via
another file descriptor." So a second `ops/local` in one test process sees
the first's lease as held, and recovery in the owning process never reads
its own lease as free.

### R-6bi3: the exported write surface

Tasks 1A and 1B export, each with its first caller in this pass:

- the write-mode open (caller: `ops/local`'s constructor, through the
  `backup` verb) and the handle's close;
- the write method and its transaction handle (caller: `ops/local`);
- the handle's read-only query path for the reads a decision does not
  rest on (caller: `ops/local`'s status read);
- the lease primitive and its busy sentinel (R-6bi2);
- a method on the kind-lease value that returns close-on-exec duplicates
  of the kind lease's and its handle's store-lock descriptors for a
  transfer child (R-6bi12; caller: `internal/backup`'s transport);
- the stamp layout's format and parse pair (R-6bi8; callers: the
  reference in the store and `internal/backup`'s prune);
- the id function;
- the backup copy (R-6bi6) and the backup reference (R-6bi16) (caller:
  `internal/backup`);
- on Ruling 4's yes, serve mode's open and its options type (callers: the
  new packages' tests; first production caller 6d's `serve`).

The migration runner and the pre-migration snapshot stay unexported until
6d. A read a write decides on runs inside the write callback, never on the
read path (spec `:355`). Each export lands with its first caller, and its
doc comment, and the doc comment of every sentinel it returns, is
corrected to describe the `errors.Is` callers that now exist (architecture
read item 4). Task 1S reshapes the signatures first, so nothing is
exported with a test-only parameter.

### R-6bi4: the actor is a method parameter; 6b-i builds the local kind

The interface takes the acting identity as a parameter of every method
(spec `:538-541`). An actor is a verified `identity.Identity` (a network
caller) or a local operator named by the operating-system user (a CLI
verb). 6b-i has no network caller, so it builds the local kind only; 6b-ii
adds the network kind and its `internal/identity` import with the
projection that first builds one. The job record's `actor` column holds a
string with the kind as its prefix: `local:<username>` at 6b-i, and
`member:<principal>` from 6b-ii. The local username comes from
`os/user.Current`, and falls back to `uid:<n>` when the uid has no passwd
entry, the shape `ownerName` already uses (`leaf.go`). `ops/local` enforces
no privilege: the projection's authorizer does, in 6b-ii, and a local
operator is the leaf's owner.

### R-6bi5: the operation vocabulary at 6b-i

Four operations, each in the catalog with a privilege from the closed set
`member` or `admin`: start a backup (`admin`), read a job's status
(`admin`), cancel a job (`admin`), and list the catalog (`member`). Job
status and cancel are `admin` because the only kind at 6b-i is backup;
rung 7, which adds member-started jobs, revisits them. A job handle is the
job id as 32 lowercase hex characters on the wire. The spec's error
vocabulary is seven sentinels, each with a stable wire code. 6b-i declares
the five it raises: `unavailable`, `not_found`, `conflict`,
`invalid_request`, and `canceled`. `forbidden` and `internal` are declared
by 6b-ii with the projection and contract suite that first raise them.

### R-6bi6: the backup's store copy is sent from staging and never kept

The job pre-creates `staging/backup-<instance>.db` at 0600, runs `VACUUM
INTO ?` on a maintenance connection, verifies the copy through the inspect
DSN (integrity `ok`, the binary's version, the store marker), sends it with
rclone, and removes it. The remote is its home. A rename into `snapshots/`
would put a backup copy beside ruling 7's pre-migration copies, which the
lost-database guard and `doctor`'s count read, and would grow the leaf by
one store per run. Its loss before the send costs a re-run, never data, so
no local durability step buys anything. The pre-migration path keeps its
rename and hand fsync (6a Task 4).

A backup start, once it holds the kind lease, removes every
`staging/backup-*.db` before it stages its own. The kind lease proves no
other backup is live, an orphaned transfer included, since the orphan
inherits the lease (R-6bi12). This replaces a per-instance cleanup in
recovery, which could never clear a stage its own live instance left after
a failed removal (risk R11, mechanics M14).

The copy ends with `wal_checkpoint(TRUNCATE)` on the same maintenance
connection, after `PRAGMA busy_timeout=0` on that connection alone. SQLite
documents that a FULL or stronger checkpoint "blocks new database writers
while it is pending" while it waits on readers through the busy handler. A
probe measured a writer stalled 1.93 s behind a checkpoint waiting its 2 s
busy timeout on a held reader, and neither waiting once the checkpoint's
busy timeout was 0 (mechanics M2). With serve live under the shared lock,
the store's 15 s busy timeout would otherwise stall every serve write for
up to 15 s per backup. The checkpoint's busy result row is logged, never a
failure (6a handoff item 4). The pragma and the checkpoint share one
connection, held through `db.Conn`.

### R-6bi7: `--max-delete 0` is the zero-track guard

The spec left open whether rclone counts a move into `--backup-dir` against
`--max-delete` (`:647-649`). A probe (`6bi-p1/`, rclone v1.74.3, `local`
backend) settled it, and the mechanics review re-ran it. After a first sync
of three files, one source file was removed and `rclone sync src
remote/library --backup-dir remote/archive/s1 --max-delete 0` ran: it
exited 7, logged `Failed to sync with 2 errors: last error was: failed to
delete 1 files`, archived nothing, and left `remote/library` holding all
three files. So a move into the backup directory counts as a delete, and
the fallback abort is not built. The same probe measured a modified file
being archived under `--max-delete 0` (`Server Side Moves: 1`), so the
limit bounds deletions only, never replacements. Task 4 pins both facts
against the real binary. The measurement covers the `local` backend only;
a copy-only backend (S3, B2) archives by server-side copy and delete, which
no probe exercised ("Not verified").

rclone's manual, quoted: `--max-delete` "tells rclone not to delete more
than N files. If that limit is exceeded then a fatal error will be
generated and rclone will stop the operation in progress." Exit code 7 is
"Fatal error (one that more retries won't fix, like account suspended)
(Fatal errors)." On `--backup-dir`: "The remote in use must support
server-side move or copy and you must use the same remote as the
destination of the sync." That requirement is why `doctor` checks the
backend's features.

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
2); and the stamp layout `20060102T150405.000000Z` (UTC, fixed width, so it
sorts lexicographically). The stamp layout has one home, the store: it
exports a format and parse pair (R-6bi3), which the reference (Task 1B),
6b-ii's `pre-restore-<stamp>.db`, and `internal/backup`'s remote listings
all use, so no second copy can drift. Format always converts to UTC first,
since `Format` on a non-UTC time writes local digits beside a literal
`Z`. Revision 1's second-resolution layout let two
runs in one second share a stamp, so `copyto` overwrote a snapshot and
`--backup-dir` wrote into the first run's archive directory, which criteria
11 and 15's back-to-back runs could meet in the gate.

### R-6bi9: the backend features check

rclone's manual: "You can also discover information about the backend
using ... `rclone backend features remote:`." A probe (`6bi-p1/`, re-run by
the mechanics review) measured its JSON `Features` object: a `local` remote
reads `Copy` false, `Move` true; an `http` remote reads both false; a
`crypt` remote over `local` reads `Move` true; an `s3` remote reads `Copy`
true, `Move` false; a remote with no `_TYPE` exits 1 with `didn't find
section in config file`; a crypt remote wrapping an undefined remote exits
1 naming the wrapped remote; a crypt remote whose `_REMOTE` is a plain path
reads `Copy` false, `Move` true and exits 0. Every offline backend answered
in 0.10 to 0.25 s. rclone's website table (`rclone.org/overview/`,
"Optional Features", read 2026-09-24) agrees for `local` (Copy ✖, Move ✔)
and `http` (both ✖), and lists S3 and B2 as Copy ✔, Move ✖, so "lacks
server-side move and copy" means both false. The check runs only with the
flag on, after the remote validator (R-6bi14) and the `_TYPE` check pass,
with `--low-level-retries 1` (Ruling 2).

### R-6bi10: exit codes and `main`'s exit path

Ruling 3's codes. `main` today prints any error and exits 1
(`main.go:54`). Task 6 adds one branch: an error carrying an exit code
(one unexported type in `package main`, matched with `errors.As`) exits
with that code, and every other error keeps exit 1. The verb's help lists
the five codes and says 130 means canceled by any means.

### R-6bi11: the verb survives its own signal

`executeWithSignalContext` cancels the verb's context on SIGINT or SIGTERM
(`main.go`, quoting `signal.NotifyContext`: the context "is marked done ...
when one of the listed signals arrives"). The job must not run under that
context, or a signal would kill rclone and bypass the record. So the verb
starts the job on a context derived from `context.WithoutCancel`, and on
the signal it calls the cancel operation with a fresh deadline context.
The verb also treats SIGHUP as a cancel, because a closed terminal sends it
and its default action would kill the verb while rclone, in its own process
group (R-6bi12), survives.

Three rules keep the verb honest about when the work ends (risk R4).

- **The verb exits when its runner's work ends, not when the record turns
  terminal.** `ops/local`'s close waits for its own jobs' goroutines, whose
  stage removal and kind-lease release follow the terminal write, and the
  verb exits only after that close returns.
- **A failed cancel write still cancels.** When the cancel operation's
  write fails (the store busy past the busy timeout), the verb, which owns
  the runner in-process, cancels the job's context directly through the
  runner, and reports the write error.
- **A refused queued-to-running transition ends the job before its body
  runs** (Task 3), so a cancel that lands first never leaves a sync
  running under a canceled record.

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

**What `Wait` returns after a cancel, and the precedence it forces.**
`os/exec`, quoted from `Cmd.Cancel`'s doc: "If the command exits with a
non-success status, or Cancel returns an error that wraps
os.ErrProcessDone, Wait and similar methods continue to return the
command's usual exit status." The mechanics review measured it: a child
SIGINT killed returns `signal: interrupt` with `ExitCode()` -1, a child
that exits 3 on SIGINT returns `exit status 3`, and a real rclone
interrupted mid-sync returns `exit status 130` within 20 ms, leaving no
partial file. So the outcome is decided in this order, and every states
table below follows it:

1. The child exits nonzero, or dies by any signal, after the job's context
   is done: **canceled**. An exit 0 is a success whatever the context: the
   job continues, and its next progress callback observes the cancel, so
   a send that lands just before a cancel still writes its reference.
2. The child died by a signal nobody sent through `Cancel` (the OOM killer,
   an operator's `kill`): **failed**, code `rclone-signal-<NAME>`, from
   `ExitError`'s `syscall.WaitStatus`.
3. Any other nonzero exit: **failed**, code `rclone-exit-<N>`.

`doctor`'s features probe takes the same rule: a deadline hit reads "not
checked", whatever the exit.

The child runs with `SysProcAttr.Setpgid` ("sets the process group ID of
the child to Pgid, or, if Pgid == 0, to the new child's process ID"), so a
terminal's SIGINT reaches the verb alone and the job ends canceled through
its record, never failed because rclone died first. `Pdeathsig` is declined:
its doc says "the signal is sent on thread termination, which may happen
before process termination" (go.dev/issue/27505).

An orphaned child is still possible when the parent is killed outright. So
the child inherits the kind lease's and the store lock's descriptors
through `ExtraFiles` ("additional open files to be inherited by the new
process ... If non-nil, entry i becomes file descriptor 3+i"). Go opens
every file close-on-exec (`os/file_unix.go`: `open(name,
flag|syscall.O_CLOEXEC, ...)`), so nothing but `ExtraFiles` reaches the
child. The store hands out duplicates made with `fcntl(F_DUPFD_CLOEXEC)`,
never plain `dup(2)`: a fold probe on go1.27.1 read `FD_CLOEXEC` clear on a
`syscall.Dup` result and set on an `F_DUPFD_CLOEXEC` one, and a plain
duplicate would leak both locks into any concurrent `exec` once 6d runs
backup inside `serve`. `os/exec` clears the flag in the child for each
`ExtraFiles` slot (`syscall/exec_linux.go`, pass 2: "The new fd is created
NOT close-on-exec, which is exactly what we want," and an `F_SETFD` clear
when the descriptor already sits at its slot). The transport owns the
duplicates and closes them after `Wait` (R-6bi3), so no caller can close the store's own
descriptor and drop its lock early. `flock(2)`: "Locks created by flock()
are associated with an open file description ... the lock is released
either by an explicit LOCK_UN operation on any of these duplicate file
descriptors, or when all such file descriptors have been closed." An orphan
therefore keeps a second backup and a restore out until it exits (probed:
a lock inherited by a child survived the parent's SIGKILL). The orphan's
window is bounded only loosely: rclone is a Go program, and `os/signal` says
"A write to a broken pipe on file descriptors 1 or 2 ... will cause the
program to exit with a SIGPIPE signal," so an orphan whose stderr pipes to
the dead verb exits at its next stderr write, which at rclone's default
log level may be the sync's end; a partway sync resumes at the next run.
So the orphan window can last a whole sync, and the inherited locks are
what keep it safe. For rung 16's unit, the close records a runbook line:
`KillMode=mixed` sends SIGTERM "to the main process while the subsequent
SIGKILL signal ... is sent to all remaining processes"
(`systemd.kill(5)`), which lets the verb cancel through the record first.

### R-6bi13: `doctor` reads backup facts through the inspect report

`doctor` opens nothing for writing. The inspect report gains read-only
fields for a store that classifies at or above version 2 and at or below
the binary's version: the newest succeeded job's finish time and final
`progress` text per kind, the `-wal` file's size, the free bytes on the
leaf's filesystem, and the count and total size of the `.db` copies in
`snapshots/` (the backup reference is not a copy and is not counted).
Sizes come from `os.Stat` and `statfs`. An absent `-wal` reads 0; a
`statfs` failure reads "unknown" and never fails `doctor`. A newer store
reports none of them, since this binary cannot trust its schema. The job
query runs on the inspect DSN.

### R-6bi14: one remote validator, and the join rule

`DUBPLATE_BACKUP_REMOTE` is validated in one function in `internal/backup`,
which both `doctor` (Task 7) and the job's first phase (Task 4) call; the
job fails with `remote-invalid` before any guard. The rule: a name before
the first colon matching `[A-Za-z0-9_]+`. A value with no colon is refused,
because rclone reads it as a local path, and a relative one writes a full
library copy under the working directory (risk R3, probed). The `_TYPE`
check is `doctor`'s, since the job reads no environment.

The destinations are never built by string concatenation. The value splits
at its first colon into a name and a path, and each destination is the
name, a colon, and `path.Join(path, sub)` for `sub` in `library`,
`archive/<stamp>`, and `snapshots/<stamp>.db`. So `name:` gives
`name:library`, `name:dir` and `name:dir/` both give `name:dir/library`,
and `name:/abs` keeps its absolute path. Revision 1's `<remote>/library`
turned `name:` into `name:/library`, an absolute path from the backend's
root, and on a bucket backend three different buckets (mechanics M5).
rclone reads an empty path as the backend's root, so on the `local`
backend `name:library` is relative to the working directory (probed), and
on a bucket backend the path's first element is the bucket; the runbook
line says the path part must name where the backup lives, bucket included.

Every `deletefile` and `purge` argument is built only from a stamp that
parses back through the layout, so an empty or malformed stamp can never
produce `rclone purge <remote>/archive/`. `purge` ignores filters, which
is why this is a construction rule and a tested one.

### R-6bi15: pruning never trusts the clock alone

"The run's own snapshot is the newest" held only on a sane clock (contract
C7, risk R2). If 14 or more remote stamps are later than the run's (a clock
stepped back, a reset RTC, a restored box writing to the same remote),
keep-14-newest deletes the snapshot the run just sent. So the prune:

- lists `snapshots/` and requires the run's own `<stamp>.db` in the
  listing; absent, it skips every prune step;
- requires the run's stamp to be strictly greater than every other parsed
  stamp; otherwise it skips every prune step, local deletion included;
- keeps the run's own snapshot by name, and the 13 newest of the rest;
- measures archive age from the run's stamp.

A skip is recorded as the succeeded record's final `progress` text and
logged; the job still succeeds, since the backup landed. A forward clock
jump still purges archive directories early; that residual is recorded
under "Not verified". A listing of an absent directory is empty: rclone
v1.74.3's `lsjson` on a missing directory exits 3 with `directory not
found` (probed), and a remote has no `archive/` until a run first archives
a file.

### R-6bi16: the backup reference keeps the lost-database guard alive

Erratum E11 refuses to create an empty store when `snapshots/` holds a
pre-migration or pre-restore copy (`open.go`, `besideCopy` and
`newestCopy`); it is the leaf's only memory that a store once existed.
Ruling 7 deletes those copies after each successful backup, so on a box
with backup on, E11 would never fire again, and a lost database file would
let serve create an empty store whose snapshots then displace every good
one within 14 runs (risk R1). So a successful backup, after its send and
before its local deletion, writes `snapshots/backup-<stamp>.ref` through
the store, holding the stamp and the remote, durably (temp file, fsync,
rename, directory fsync, as the pre-migration path does), and removes the
previous reference after the new one lands. `newestCopy` counts a
reference, and `besideCopy`'s refusal for one names `restore --snapshot
<stamp>`. Ruling 7's deletion stands and criterion 16 holds unchanged. A
regression floor on remote pruning (the other half of R1's chain, a
logically emptied store) routes to rung 7, the first writer of tracks.

### R-6bi17: the job body contract, and who wires it

`ops/local` imports `backup` (spec table `:198-205`), so the binding lives
in `ops/local`, and Task 4, which creates the body, adds it there.

Task 3 defines the body's contract without `backup` in view: a function
taking `ctx` and a progress callback. The callback writes the phase text to
`progress` and is a cancel-poll point: when `cancel_requested` is set it
cancels the job's context and returns the context's error. A body fails
with an error the runner reads a failure code and message from through
`errors.As` on a small interface `ops/local` declares; an error without
one records failure code `job-error`. A succeeded body may return a final
progress text, which the terminal transition writes.

Task 4 binds the backup kind to `backup`'s body in `internal/ops/local`.
Every test that needs a real runner around the backup body lives in the
external test package `backup_test`, since an in-package test of `backup`
importing `ops/local` is an import cycle once `ops/local` imports
`backup`.

### R-6bi18: the library root's channel, resolved once

`DUBPLATE_LIBRARY_ROOT` is a registered key (rung 1).
`cmd/dubplate-server` reads it through the wall and passes it to the
constructor with the flag and the remote. The job resolves it once with
`filepath.EvalSymlinks` in its first phase, and both the count guard and
rclone use the resolved path. `path/filepath.WalkDir`, quoted: "WalkDir
does not follow symbolic links", and it does not follow a symlinked root
either: a probe walked a symlink to a directory of two files and counted
0. rclone follows a symlinked root and syncs its contents (risk R7,
probed). Without the resolve, a library mount reached through a symlink
aborts every run as empty.

---

## Review focus

These are the input classes most likely to bite, each pinned by a test in
the task that owns it.

1. **A second process that owns a running job.** Expected: an open in
   another process leaves it alone, while the same open marks a dead
   owner's record interrupted. Pinned in Task 3.
2. **A process killed mid-job.** Expected: the next open marks its records
   interrupted and removes its lease file, and a late update from anywhere
   is refused. Pinned in Task 3.
3. **A wrong or wiped library mount.** Expected: no rclone invocation.
   Pinned in Task 4.
4. **A zero-track catalog over a populated remote.** Expected: a run that
   would archive a file ends failed. Pinned in Task 4.
5. **SIGINT at a terminal mid-sync.** Expected: the job ends canceled, the
   transfer receives exactly one SIGINT (the runner's, never the
   terminal's), a stand-in that ignores it is killed at `WaitDelay`, and the
   verb exits 130. Pinned in Task 6.
6. **A prune that deletes the latest snapshot.** Expected: never, on any
   clock. Pinned in Task 5.
7. **A backup remote an adopter configured wrong.** Expected: `doctor`
   names the defect before the first run, and the job refuses a malformed
   value before any guard. Pinned in Tasks 4 and 7.
8. **A lost database file on a box whose backups deleted every local
   copy.** Expected: serve's create refuses, naming the backup reference.
   Pinned in Task 1B.

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched by the conductor after 6a merges
and before any implementer. Reads the spec, the 6a plan's handoff, and this
plan against the merged tree. **Its output is blockers with evidence, not
fixes.** It writes no code and edits no file.

**Every check runs against the real HEAD, and the task reports that HEAD.**
This plan's code pointers were resolved at `0b6da0f` on master, after the
simplifier's renames. Grade a moved pointer by content: a pointer whose
content survives at any offset or under a new name is a note naming the new
site, and a pointer whose content is gone is a blocker.

**The first six items are blockers: a "no" stops the launch.**

1. **Is the worktree fresh from master, with 6a merged (at `8736528`)?**
   Report `git worktree list`, the lane registry rows, and `git status` in
   both trees. *(Expected: lane 1 on 5c-ii in the main tree, lane 2's
   worktree fresh from master.)*
2. **Does anything after `0b6da0f` reshape a surface this plan names?**
   Report the diff from `0b6da0f` to HEAD for
   `internal/engine/store` and `cmd/dubplate-server/store.go`, by function.
   Task 1S's reshape (architecture read items 1 to 3) has not run yet, so
   the `euid` parameters, `inspect`, `migrateOptions.backup`, and
   `migrate`'s `fs.FS` parameter are expected as read.
   The renames `f85e7be` made (`createLeaf` to `leaf.create`, `statLeaf`
   to `leaf.stat`, `probeLeaf` to `leaf.probe`, and the other leaf-first
   functions to methods) are already mapped here and are not blockers. A
   later change to the behavior or signature of `update`, `writeTx`,
   `openWrite`, `openServe`, `serveOptions`, `takeSnapshot`, `copyInto`,
   `verifyCopy`, `checkFreeSpace`, `tryLock`, `lockHolders`, `leaf.create`,
   `leaf.stat`, `besideCopy`, `newestCopy`, `Report`, or `storeRowStatus`
   is a blocker until the conductor maps it.
3. **Do `internal/ops`, `internal/ops/local`, and `internal/backup` not
   exist, and does no file import `os/exec` for `rclone`?**
4. **Is the "backup targets" reserved class still at `keys.go` with its six
   names, and is none of `DUBPLATE_BACKUP_ENABLED`,
   `DUBPLATE_BACKUP_REMOTE`, or the six retired names declared anywhere
   else?** Report `git grep` for each, and that `DUBPLATE_LIBRARY_ROOT` is
   still a registered key.
5. **Has 5c-ii touched `internal/config`, `doctor.go`, `check.go`,
   `store.go`, or `internal/engine/store`?** Any commit is a blocker for
   R-6bi1.
6. **Does `rclone` resolve on the gate's `PATH`, and at what version?**
   Report `rclone version`. A version other than v1.74.3 is a note, and
   Task 4 re-runs R-6bi7's probe and R-6bi15's `lsjson` exit-3 probe
   against it.

**Notes (not blockers):**

7. Report `newProductionRegistry`'s body, `newRootCmd`'s `AddCommand`
   block, and `main`'s exit path at HEAD.
8. Report `doctor.go`'s row order and `TestDoctorCmd_LineCountFollowsRegistry`'s
   arithmetic at HEAD, which Task 7 extends.
9. Report `migrations/0002_jobs.sql`'s columns at HEAD, which Task 3 writes
   against; a changed column is a blocker for Task 3 alone.
10. Report the spec pointers this plan cites, which errata may have moved.
11. Report whether the templates still carry the six retired names
    (`env/dubplate.env.template:49-53` and `:89`), which the template
    classification test must read as superseded after Task 6.
12. Report Ruling 4's outcome as the conductor recorded it, since Task 1A's
    export list depends on it.

**Deliverable count: 1** (the report).

---

## Task 1S: The store's signatures, reshaped before 6b-i binds to them

Model: `sonnet`. Depends on Task 0. **Neutral on 2.0 hooks.** The 6a
architecture read (`docs/superpowers/research/2026-09-24-rung-6a-architecture-reads.md`,
`internal/engine/store`, items 1 to 3) found three test-only seams in the
signatures 6b-i is about to export or bind to. This task removes them while
the store has no caller outside its package, so no later pass reshapes an
exported signature. It changes no behavior.

**Files:**
- Modify: `internal/engine/store/leaf.go`, `open.go`, `inspect.go`,
  `migrate.go`, and their tests (`leaf_test.go`, `open_test.go`,
  `inspect_test.go`, `migrate_test.go`, `snapshot_test.go`, and any other
  test that passes `euid` or an `fs.FS` to `migrate`)

**Outcomes:**

- **One uid seam** (item 1). `euid int` leaves `leaf.create`,
  `checkOwner`, `leaf.stat`, `openWrite`, `openServe`, and `inspect`
  (`leaf.go:65`, `:100`; `open.go:101`, `:172`, `:308`;
  `inspect.go:119-124`). The package reads the effective uid through one
  package variable, `var geteuid = os.Geteuid`, the stdlib seam form it
  already uses for `statfsFunc` (`fsprobe.go:43`). `inspect` folds into
  `Inspect`. Tests that simulate another uid swap the variable and restore
  it with `t.Cleanup`; the store's no-`t.Parallel` rule already makes that
  safe.
- **The backup setting reaches the runner** (item 2).
  `migrateOptions.backup` (`migrate.go:62`) has no production setter:
  `settle` builds the runner's options without it (`open.go:261`), and the
  tests default it to true, the opposite of production. Serve mode's
  options gain the same field and `settle` passes it through, so 6d's
  `serve` wires `DUBPLATE_BACKUP_ENABLED` to it without reshaping the type
  Ruling 4 may export. Its zero value stays production's: backup off, the
  runner prunes pre-migration copies to `keep`. The test helper
  (`migrate_test.go:63-65`, "keeps backup on, so no prune runs unless a row
  asks") defaults to the zero value, and each row that relies on no prune
  sets the field on by name, so the default no longer hides production's
  behavior.
- **`migrate` takes the parsed set** (item 3). `migrate` takes the
  `[]migration` its caller parsed, in place of `src fs.FS`
  (`migrate.go:90`, `:94`; `open.go:174`, `:261`), so `openServe` no
  longer parses the embedded set twice. Tests with a synthetic set call
  `readMigrations` themselves, so the near-miss marker refusal (Task 1B)
  still fires at every test's read.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Effective uid | owns the state directory | as today |
| Effective uid | differs (the test swaps `geteuid`) | refused naming the owner, as today |
| Serve options' backup field | zero (off) | the runner prunes to `keep` after a migration |
| Serve options' backup field | on | the runner keeps every pre-migration copy |
| `migrate`'s input | a parsed set | applied as today; no second parse |

**Acceptance criteria:**

- `git grep -n 'euid int' -- internal/engine/store` finds nothing, and no
  unexported `inspect` remains beside `Inspect`.
- Every store test that passed before passes, with only its call sites
  changed; no test's expected value changes.
- A new row drives serve mode's open (in-package) over a populated store
  one version back with two planted older pre-migration copies and
  `keep` 1: with the backup field off, one copy survives; with it on, all
  three do. **Fixture state:** the older store and both planted copies.
- The uid-mismatch rows refuse through the swapped `geteuid`.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Call `os.Geteuid()` directly in `leaf.create` → the swapped-uid
   create row does not refuse.
2. Leave the backup field out of `settle`'s runner options → the
   backup-on serve row prunes to one copy.
3. Re-read the embedded set inside `migrate` → no behavior changes; report
   it reviewer-read (a double parse only).

**2.0 hook:** neutral. **Prior art:** the stdlib's own package-variable
seam form (`statfsFunc` in this package).

**Deliverable count: 2** (the uid seam with `Inspect`'s fold, the runner's
options with the backup field wired and the parsed set).

---

## Task 1A: The store's surface for 6b-i: exports, leases, and the inherited descriptors

Model: `opus`. Depends on Task 1S. **Exports hook 9's write method to its
first caller.**

**Files:**
- Create: `internal/engine/store/lease.go`, `internal/engine/store/lease_test.go`
- Modify: `internal/engine/store/leaf.go`, `open.go`, `write.go`,
  `conn.go`, `lock.go`, `id.go`, `doc.go`, and their tests

**Outcomes:**

- **The exports** R-6bi3 names except the backup copy and reference (Task
  1B), each documented at the Go Doc Comments bar, with every exported
  sentinel's doc naming the `errors.Is` callers that now exist
  (architecture read item 4). The write method's doc keeps its measured
  contract (R-6a18). On Ruling 4's yes, serve mode's
  open and its options type are exported, and the open refuses a nil
  logger (6a's carried item); the settle-loop bound stays with 6d.
- **`leases/`** joins the leaf (R-6bi2): created by serve mode at 0700,
  counted by the completeness check. A write-mode open of a leaf without
  `leases/` refuses naming the missing directory and "start serve to finish
  creating it," the lockless leaf's wording.
- **The lease primitive** (R-6bi2). An instance lease is an exclusive
  `flock` on `leases/<instance>`, where the instance is 32 lowercase hex
  characters from the id function, created 0600 with `O_CREATE|O_EXCL`,
  and returned as a lease value its taker holds until it releases it: the
  store handle takes none, and `ops/local`'s runner is its one owner
  (R-6bi2). Its release removes the file while holding it, then closes
  it. A kind lease is an exclusive, non-blocking `flock` on
  `leases/kind-<kind>`, created if absent and never removed, whose busy
  error wraps an exported sentinel and names the pids that acquired it, or
  "held by another process" when `/proc/locks` names none. Trying another
  instance's lease returns held, free (acquired, then the file removed
  while held, then closed; a not-exist answer from that removal also reads
  free), or absent (no file), and refuses on any other error naming the
  path.
- **The inherited descriptors** (R-6bi12): one method on the kind-lease
  value returns duplicates of the kind lease's and its handle's store-lock
  descriptors, made with `fcntl(F_DUPFD_CLOEXEC)`, as `*os.File` values the
  caller owns and closes.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| `leases/` | present, 0700 | proceed |
| `leases/` | absent in an existing leaf | write mode refuses, naming the directory and "start serve to finish creating it" |
| Instance lease | created and locked | held until its taker releases it |
| Instance lease | the file already exists (an id collision) | refused, naming the path; never reused |
| Instance lease | released cleanly | the file removed while held, then closed |
| Kind lease | free | held for the job's life |
| Kind lease | held by another descriptor | busy (the sentinel), naming acquirer pids or "held by another process" |
| Try another's lease | held | held; nothing removed |
| Try another's lease | free | free; the file removed while held |
| Try another's lease | free, but the removal answers not-exist | free |
| Try another's lease | absent | absent |
| Try another's lease | open fails with anything but not-exist | refused, naming the path and the error |
| Descriptors | returned | close-on-exec duplicates; closing them leaves the store's lock and the kind lease held |
| Serve open (Ruling 4's yes) | nil logger | refused before any file is touched |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Two handles in one test process: the second's kind lease on `backup` is
  busy while the first holds it (matched with `errors.Is` on the
  sentinel), and free after the first releases. The second's try of the
  first's instance lease reads held. **Fixture state:** a store at the
  binary's version, both handles open, each with an instance lease taken
  through the primitive. The busy message is asserted by
  either form, never by pid, since the gate may run on btrfs.
- After the first handle's lease descriptor is closed without removal (a
  dead owner), a try reads free and the file is gone; a second try reads
  absent.
- The removal race: two descriptors open the same dead lease file; the
  first try removes it; a try through the second descriptor's path then
  reads free, not refused.
- An instance lease's release leaves `leases/` without its file, and
  opening and closing a store handle creates no file in `leases/`.
- A re-executed child started with the returned duplicates in its
  `ExtraFiles` keeps the kind lease busy after the parent closes every
  descriptor it holds, and the lease reads free after the child exits.
  Closing the duplicates in the parent leaves the store lock held. Each
  returned descriptor reads `FD_CLOEXEC` set through `F_GETFD`.
- The lease files are 0600 and `leases/` is 0700 under a 0022 umask.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Take the kind lease shared → the two-handle busy row passes both.
2. Remove the lease file before acquiring it → report whether the free row
   still holds, and name the race (a live owner's file removed) as a
   reviewer-read row if no test fires.
3. Treat absent as held → the dead-owner second-try row fails.
4. Refuse on a not-exist removal → the removal-race row fails.
5. Skip the release's removal → the clean-close row fails.
6. Return the store's own descriptor instead of a duplicate → closing it
   in the parent releases the store lock (report the row that fires).
7. Duplicate with plain `dup(2)` → the `FD_CLOEXEC` row fails.
8. Take an instance lease in the handle's open → the no-file-on-open row
   fails.

**2.0 hook:** exports hook 9's write method. **Prior art:** none beyond
`flock(2)`.

**Deliverable count: 3** (the exports with serve's open on Ruling 4's yes,
`leases/` and the lease primitive, the inherited descriptors).

---

## Task 1B: The store's surface for 6b-i: the backup copy, the reference, and the carried items

Model: `opus`. Depends on Task 1A. **Lands 2.0 hook 10** (the backup copy
by `VACUUM INTO`).

**Files:**
- Create: `internal/engine/store/stamp.go`, `stamp_test.go`
- Modify: `internal/engine/store/snapshot.go`, `open.go`, `migrate.go`,
  `inspect.go`, `doc.go`, and their tests

**Outcomes:**

- **The backup copy** (R-6bi6): free space for one copy plus the WAL bound
  (6a's `checkFreeSpace`), pre-create `staging/backup-<instance>.db` at
  0600, `VACUUM INTO ?` on a maintenance connection, verify through the
  inspect DSN, and return the path. The caller removes it. It ends with
  `wal_checkpoint(TRUNCATE)` on the same connection after `PRAGMA
  busy_timeout=0`, logging a busy row. A second exported step removes every
  `staging/backup-*.db`; the backup job calls it holding the kind lease.
- **The stamp layout** (R-6bi8): an exported format and parse pair, the
  layout's one home. Format converts to UTC first; parse accepts only the
  fixed-width layout with a literal `Z`.
- **The backup reference** (R-6bi16): an exported step writes
  `snapshots/backup-<stamp>.ref` durably and removes the previous
  reference after the new one lands. It refuses a stamp that does not
  parse through the layout. `newestCopy` counts a reference, and
  `besideCopy`'s refusal for one names `restore --snapshot <stamp>`.
- **The inspect report's new fields** (R-6bi13).
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
| Backup copy | free space short | refused, naming the need, the free bytes, and the path; no stage created |
| Backup copy | the stage already exists | refused, naming the path |
| Backup copy | fails verification | refused; the stage removed |
| Backup copy | checkpoint reports busy | logged at once, without waiting the busy timeout; the copy stands |
| Stage sweep | `staging/backup-*.db` present | removed |
| Stage sweep | any other name | untouched |
| Stamp | a non-UTC time formatted | the UTC digits |
| Stamp | a missing, short, or long fraction, or a zone other than `Z` | parse refuses |
| Reference | written | one `backup-<stamp>.ref`, the previous one removed |
| Reference | a stamp that does not parse | refused; nothing written |
| Serve's create | the database absent, a reference present | refused, naming the reference and `restore --snapshot <stamp>` |
| Inspect fields | `-wal` absent | size 0 |
| Inspect fields | `statfs` fails | free space "unknown" |
| Inspect fields | no `.db` copy | count 0, size 0 |
| Pre-migration rename | fails | refused; the stage removed; the store at N |
| Rebuild marker | exact | a rebuild |
| Rebuild marker | a near miss (`\r`, trailing space or tab) | the set refused, naming the file |
| Rebuild marker | any other first line | an ordinary file |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criterion 9's store half, as two fixtures, since one open write
  transaction holds SQLite's single write lock and would stop a looping
  writer (mechanics M1, contract C5). **Fixture A:** a writer commits
  through the write method in a loop across the copy; the test records the
  committed count at a barrier just before the copy starts; the copy
  passes `integrity_check`, reads the binary's version and the marker, and
  holds at least that many rows. **Fixture B:** a second handle holds an
  uncommitted insert of one marked row across the copy, with no looping
  writer; the copy finishes and lacks the marked row. **Fixture state:** a
  populated `jobs` table in both.
- The checkpoint row: with a read transaction held open on a second
  handle, a writer started during the copy's checkpoint commits within 1 s,
  far below the 15 s busy timeout, and the busy row is logged.
- The stage is 0600 under a 0022 umask.
- Free space short and stage already exists are each driven by their real
  trigger (6a's free-space seam; a planted stage); verification failure is
  reviewer-read, since no production-real trigger exists.
- The sweep removes two planted `backup-*.db` stages and leaves a planted
  `pre-migrate-v1.db` stage alone.
- The stamp rows: a time in a non-UTC zone formats to its UTC digits and
  round-trips; each malformed form the table names is refused.
- The reference: a second write leaves exactly one reference; with the
  database file removed from a leaf whose `snapshots/` holds only a
  reference, serve's create refuses naming it and creates nothing.
- The rename failure: through the unexported migration runner after the
  leaf is created, since `leaf.create` chmods every leaf directory back to
  0700 (`leaf.go`), `snapshots/` made unwritable, a migration from a
  populated store at N refuses, and `staging/` is empty afterward.
- The near-miss rows: a synthetic set whose file 2 begins with the marker
  plus `\r\n`, and one with a trailing space, are each refused naming the
  file; the exact marker still runs as a rebuild.
- The inspect report's new fields read back for a current store with a
  succeeded `backup` record carrying a final `progress` text, a
  1-byte-or-larger `-wal`, two copies and one reference in `snapshots/`
  (the count reads 2); an absent `-wal` reads 0; a newer store reports none
  of them.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Run the backup copy on the write pool → report the effect on the looping
   writer (6a's stall), or name it reviewer-read.
2. Drop the checkpoint's `busy_timeout=0` → the checkpoint row's writer
   waits the busy timeout.
3. Keep the stage on a rename failure → the rename-failure row fails.
4. Compare the marker after `strings.TrimSpace` and accept it → the
   near-miss rows run a rebuild instead of refusing.
5. Leave the reference out of `newestCopy` → serve's create builds an
   empty store in the reference row.
6. Open the database file through `os` to size it → report which R-6a4 row
   fires, or name it reviewer-read.

**2.0 hook:** lands hook 10. **Prior art:** SQLite `VACUUM INTO`, adopted;
the driver's backup API (`conn.go:1208-1219` in the module) stays the
recorded alternative.

**Deliverable count: 3** (the backup copy with its checkpoint and sweep,
the stamp layout with the reference and the guard's reading of it, the two
carried items with the report fields).

---

## Task 2: `internal/ops`: the interface, wire types, codes, privileges, and catalog

Model: `sonnet`. Depends on Task 0. **Independent by files of Tasks 1A and
1B.** **Lands the ops-seam half of 2.0 hook 9** (the seam a later projector
joins).

**Files:**
- Create: `internal/ops/doc.go`, `ops.go`, `errors.go`, `catalog.go`, and
  their tests

**Outcomes:**

- **One interface** with R-6bi5's four operations. Each method takes `ctx`,
  the actor (R-6bi4), and a request type, and returns a response type and
  an error. Request and response types are JSON-serializable, with
  explicit `json` tags in snake case, and none carries an actor field. The
  actor type has the local kind only at 6b-i, built from an operating-system
  username, and renders as `local:<username>`.
- **The job model's types:** the handle, the six states (queued, running,
  succeeded, failed, canceled, interrupted) with a terminal predicate, and
  a status type carrying the state, the kind, the phase, the failure code
  and message, and the times.
- **The error vocabulary:** R-6bi5's five sentinels, each wrapped by an
  error type that carries its stable wire code and, for unavailable, a
  reason. Callers match with `errors.Is`.
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
| Local actor | an empty username | construction refuses |

**Acceptance criteria:**

- Spec criterion 1: a table of every catalog entry asserts its privilege is
  `member` or `admin`; a catalog built with a zero privilege fails, naming
  the operation.
- Every sentinel's wire code is pinned by a table whose expected strings are
  R-6bi5's literals, and a test fails when a sentinel has no code.
- A JSON round trip of every request and response type is lossless, and
  decoding `{"actor":"someone"}` into the start request leaves it equal to
  its zero value (spec criterion 2's type half).
- `internal/ops` imports only the standard library (checked by `go list
  -deps`).
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

Model: `opus`. Depends on Tasks 1A and 2. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/ops/local/doc.go`, `local.go`, `jobs.go`,
  `recovery.go`, and their tests, with a `TestMain` that re-executes the
  test binary as a second process

**Outcomes:**

- **The constructor** opens nothing itself. It takes a write-mode handle,
  the backup configuration (flag, remote, and library root, R-6bi18), a
  logger that must be non-nil, and the transfer's parameters, takes its
  instance lease, and runs recovery. Its close waits for its own jobs'
  goroutines to return, then releases its instance lease (R-6bi11).
- **Recovery** (spec table `:560-565`): for every queued or running record,
  try its owner's lease. Held: leave it. Free or absent: mark it
  interrupted in one write, with a `WHERE` naming both states it may hold,
  setting `finished_at` and `updated_at` in the same statement. Recovery
  runs at construction and again before every exclusive start. It never
  touches staging (R-6bi6).
- **The body contract** (R-6bi17): a function taking `ctx` and a progress
  callback; the callback writes `progress` and is a cancel-poll point; a
  failure's code and message are read through `errors.As` on a small
  interface this package declares, and any other error records
  `job-error`. Task 4 binds the backup kind's body.
- **The runner.** A start checks availability first (a flag-off backup
  returns unavailable and writes nothing), then takes the kind lease. On
  busy it runs recovery before answering, so a dead owner's record is
  interrupted first: a remaining queued or running record of the kind
  returns conflict naming its handle, and none returns conflict with "a
  previous run's transfer is still running" and no pid (R-6bi2). With the
  lease held it runs recovery, prunes retention, inserts the record as
  queued with the actor, and runs the job on a goroutine under a context
  the runner cancels. Each transition writes the state it expects in its
  `WHERE`. The queued-to-running transition sets `started_at`; a terminal
  transition sets `finished_at`. A terminal transition that matches no row
  is refused and logged at warning with the job's handle and both states.
  A queued-to-running transition that matches no row (a cancel landed
  first) ends the goroutine before the body runs, logged at info.
- **Cancel** writes `cancel_requested`. On a queued record it writes
  canceled directly, with `finished_at`. On a terminal record it writes
  nothing and reports the state. On a running record owned by any process,
  the owning runner reads the column at each progress callback and every
  cancel poll interval while the body runs, and cancels the job's context;
  the job then ends canceled. A job that finishes before it sees the
  request keeps its own state, and the cancel response says which state
  the job ended in. The runner also exposes an in-process cancel of its
  own job's context, which the verb uses when its cancel write fails.
- **Retention** at each start deletes terminal records whose finish time is
  older than 90 days, except the newest succeeded record of each kind.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Record, owner lease | held | left alone, whatever its kind |
| Record, owner lease | free | interrupted, `finished_at` set; the lease file removed |
| Record, owner lease | absent | interrupted, `finished_at` set |
| Record, owner lease | unreadable | recovery refused, naming the path; the constructor fails |
| Record | terminal | unchanged |
| Late terminal update | the record is interrupted | refused, logged at warning; the record stays interrupted |
| Queued to running | the record is already canceled | the body never runs; logged at info |
| Start | flag off | unavailable, reason "backup integration disabled"; no record, no kind lease |
| Start | kind lease busy, a live queued or running record | conflict naming its handle |
| Start | kind lease busy, no live record after recovery | conflict, "a previous run's transfer is still running" |
| Cancel | unknown handle | not found |
| Cancel | malformed handle | invalid request |
| Cancel | queued | canceled, `finished_at` set |
| Cancel | running | `cancel_requested` set; the owner ends it canceled |
| Cancel | terminal | no write; reports the state |
| Status | unknown handle | not found |
| Body | fails with a coded error | failed, its code and message |
| Body | fails with any other error | failed, `job-error` |
| Close | a job still running | waits for its goroutine |
| Retention | no record past 90 days | nothing deleted |
| Retention | the only succeeded record of a kind is past 90 days | kept |
| Retention | an interrupted record past 90 days | deleted |

**Acceptance criteria:**

- Spec criterion 3, with a positive control: a re-executed child process
  opens `ops/local` over the store, starts a job whose body blocks, and
  signals readiness on a pipe; a second non-terminal record whose owner
  lease is free is planted; the parent opens its own `ops/local`; the
  child's record still reads running and the planted record reads
  interrupted. **Fixture state:** a current store; the child's job
  running, its lease held; one dead-owner record.
- Spec criterion 4: a queued and a running record planted with an owner
  whose lease file exists unlocked, plus one with no lease file, read
  interrupted with `finished_at` set after an open, and the unlocked lease
  file is gone. **Fixture state:** three records, all non-terminal, and one
  terminal record naming the same dead owner as the unlocked-lease queued
  record, which stays unchanged.
- Spec criterion 5: the child is killed with SIGKILL mid-job, the parent's
  recovery marks it interrupted, and a terminal update issued afterward
  through the runner's own transition is refused and logged; the record
  stays interrupted. The log line is asserted by message and handle.
- Spec criterion 6: from a second `ops/local`, cancel on the child's
  running job sets `cancel_requested`, and the child's record ends
  canceled. The child's body reports one phase, then blocks for 5 s,
  longer than the 2 s poll interval, then returns success. Cancel on a
  queued record whose owner lease the child holds ends it canceled (a
  free-owner queued record would be interrupted by the canceller's own
  recovery first). Cancel on a succeeded record reports succeeded and
  writes nothing.
- The cancel-first row: an in-package test inserts a queued record,
  cancels it (canceled), then drives the runner's unexported run step for
  it, with no test-only seam in the runner; the body's ran flag stays false
  and the record reads canceled.
- Spec criterion 7: a second start while the child's backup runs returns
  conflict naming the child's handle.
- Spec criterion 8, the runner's half: flag off, the catalog reports
  backup unavailable with the exact reason, and a start writes no record
  and leaves `leases/` holding only the instance lease; flag on, the
  catalog reports available.
- Spec criterion 2's record half: the start's actor parameter is what the
  record's `actor` column holds.
- Spec criterion 17: three terminal records, one succeeded at 100 days, one
  failed at 100 days, one succeeded at 10 days, planted through the write
  method: the next start deletes both 100-day records and keeps the 10-day
  one, since only the newest succeeded record is protected. A second
  fixture with one succeeded record at 100 days keeps it. A third with an
  interrupted record at 100 days, written by recovery, deletes it.
- After the constructor's close, `leases/` holds no instance file of its
  own.
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
   record reads canceled while its job still runs (report the body's state
   at that moment).
5. Retain by count instead of by the newest succeeded per kind → the
   single-record fixture deletes it.
6. Protect every succeeded record → criterion 17's 100-day succeeded
   record survives.
7. Answer the conflict before recovery → report whether a dead owner's
   handle is named, or name it reviewer-read.
8. Poll `cancel_requested` only at progress callbacks → criterion 6's child
   ends succeeded instead of canceled.
9. Skip recovery at construction → criterion 3's dead-owner record stays
   running.
10. Run the body after a refused queued-to-running → the cancel-first
    row's ran flag is true.
11. Leave `finished_at` unset on recovery → the interrupted retention row
    keeps the record.

**Prior art:** none adopted. A job queue library would own the table this
spec already designed (`0002_jobs.sql`).

**Deliverable count: 4** (the constructor with recovery and close, the
runner with its transitions and body contract, cancel, retention).
**Pre-named split:** if the implementer reports that the four cannot land
in one green gate, cancel and retention move to their own task between this
one and Task 4.

---

## Task 4: `internal/backup`: the guards, the transport, and the job

Model: `opus`. Depends on Tasks 1B and 3. **Neutral on 2.0 hooks** (hook
1's file hash drives backup change detection later; ruling 2 defers it to
rung 7).

**Files:**
- Create: `internal/backup/doc.go`, `backup.go`, `guard.go`, `remote.go`,
  `rclone.go`, and their tests, with a recording stand-in built from the
  test binary
- Modify: `internal/ops/local/jobs.go` and its test (the backup kind's
  binding, R-6bi17)

**Outcomes:**

- **The job body** runs, in order: resolve the rclone binary with
  `exec.LookPath`, validate the remote (R-6bi14), resolve the library root
  (R-6bi18), run the guards, sync the library tree, sweep
  `staging/backup-*.db` (Task 1B, under the kind lease the runner holds),
  take the backup copy, send it, and remove the stage. Each phase reports
  through the progress callback, where the runner reads
  `cancel_requested`. `ops/local` binds the backup kind to this body.
- **The guards** (spec `:638-649`), against the skeleton through unexported
  queries: a missing library root aborts. With catalog tracks, a tree whose
  regular-file count falls below half the track count aborts, and rclone
  runs with `--max-delete` equal to the track count. With zero tracks, an
  empty root aborts, and rclone runs with `--max-delete 0` (R-6bi7). A
  guard abort writes failed with a guard code and never starts rclone.
- **The remote validator and join** (R-6bi14), one function `doctor` also
  calls.
- **The transport** runs `rclone sync <root> <name>:<path>/library
  --backup-dir <name>:<path>/archive/<stamp> --max-delete N`, then `rclone
  copyto <stage> <name>:<path>/snapshots/<stamp>.db` with `-M
  --metadata-set run_id=<handle> --metadata-set sync_completed=<RFC 3339
  UTC>`, each destination joined by R-6bi14's rule. A probe (`6bi-p2/`)
  measured both metadata keys reading back through `rclone lsjson -M` on a
  `local` remote. Each invocation runs under R-6bi12's process shape, with
  the store's duplicated descriptors in `ExtraFiles`, closed after `Wait`.
  The outcome follows R-6bi12's precedence. A failure's message is the
  last stderr line, bounded to 512 bytes.
- **The layout** under the remote is `library/`, `archive/<stamp>/`, and
  `snapshots/<stamp>.db`, one stamp per run, taken at the job's start in
  R-6bi8's layout.
- **The stand-in** records its argv to a file the test names, writes its
  pid to another, exits with a code the test sets, and runs in one of two
  signal modes: one honors SIGINT (exits 130), and one ignores it, so only
  `WaitDelay`'s `Kill` ends it. In both modes it appends one line to its
  record file for every SIGINT it receives, so a test counts the signals
  that reached it rather than timing its survival.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| rclone | not on `PATH` | failed, code `rclone-missing`, before any other phase |
| Remote | no colon, or a name outside `[A-Za-z0-9_]` | failed, code `remote-invalid`, no invocation |
| Remote | `name:` | destinations `name:library`, `name:archive/<stamp>`, `name:snapshots/<stamp>.db` |
| Remote | `name:dir` or `name:dir/` | destinations under `name:dir/` |
| Library root | absent, or a dangling symlink | failed, code `library-missing`, no invocation |
| Library root | a symlink to a directory | resolved; the guard and rclone both use the target |
| Library root | a regular file, not a directory | failed, code `library-missing` |
| Library root | unreadable | failed, code `library-unreadable`, naming the error |
| Zero tracks | empty root | failed, code `library-empty`, no invocation |
| Zero tracks | non-empty root | sync with `--max-delete 0` |
| Tracks | file count below half the track count | failed, code `library-short`, naming both counts, no invocation |
| Tracks | at or above half | sync with `--max-delete` equal to the track count |
| Sync | exit 0 | continue |
| Sync | the job's context done, then a nonzero exit or any signal death | canceled; the child gone by `WaitDelay` |
| Sync | the job's context done, then exit 0 | continue; the next progress callback observes the cancel |
| Sync | killed by a signal not sent through `Cancel` | failed, `rclone-signal-<NAME>` |
| Sync | exit 7 on `--max-delete` | failed, `rclone-exit-7` |
| Sync | any other nonzero | failed, `rclone-exit-<N>` |
| Copy | refused (Task 1B's states) | failed, code `snapshot-failed`, naming the reason |
| Send | nonzero | failed by the same precedence; the stage removed |
| Stage removal | fails | logged; the job's state stands; the next start's sweep removes it |

**Acceptance criteria:**

- Spec criterion 10: a run against a real `rclone` with a `local` remote in
  a temporary directory leaves every library file under `library/`, one
  object under `snapshots/`, and that object passes `integrity_check` and
  reads the binary's version; `rclone lsjson -M` on it reads back `run_id`
  equal to the job's handle and a `sync_completed` that parses as RFC 3339.
  **Fixture state:** a store with two tracks and a library root holding
  their two files.
- Spec criterion 11: after a first run, one library file is deleted and a
  second run leaves it under `archive/<stamp>/` of the second run's stamp,
  not deleted. **Fixture state:** a populated catalog, so `--max-delete` is
  the track count.
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
  before the job began (Task 1B's fixture A, driven through the job).
- The join rows: for `name:`, `name:dir`, and `name:dir/`, the stand-in's
  recorded argv carries exactly the destinations the table names.
- The symlink rows: a root reached through a symlink to a two-file
  directory passes the guard and syncs both files; a dangling symlink ends
  `library-missing` with no invocation.
- The cancel precedence: the SIGINT-honoring stand-in, canceled mid-sync,
  ends the job canceled though its exit is 130. A stand-in killed by
  SIGKILL from the test, with no cancel, ends failed with
  `rclone-signal-KILL`.
- R-6bi12's orphan row, in `backup_test`: a re-executed child runs the job
  through a real runner against the SIGINT-ignoring stand-in; the child is
  SIGKILLed; a runner built after the kill starts a second backup, which
  returns conflict with "a previous run's transfer is still running"
  (asserted by text, never by pid); after the stand-in exits, a start
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
5. Drop `Setpgid` → Task 6's group-signal row records two SIGINT lines
   instead of one (the row that fires lives in Task 6, since it needs the
   verb as the group leader).
6. Omit `ExtraFiles` → the orphan row's second start succeeds while the
   stand-in lives.
7. Leave `WaitDelay` unset → Task 6's criterion 20 leaves the
   SIGINT-ignoring stand-in alive (report here if any test fires).
8. Join destinations by string → the `name:` join row's argv reads
   `name:/library`.
9. Walk the unresolved root → the symlink row ends `library-empty`.
10. Classify by the exit status before the context → the cancel
    precedence row ends `rclone-exit-130`.

**Prior art:** rclone as a probed subprocess, adopted (spec). The salvage
tier's missing-root guard is adopted in reasoning, and its top-level-entry
check is replaced by the count guard (`_salvage/internal/backup/backup.go:22-34`);
nothing is imported.

**Deliverable count: 4** (the guards with the root's resolution, the
transport with its precedence, the job body and its binding, the layout
with its stamp and the remote's validator and join).

---

## Task 5: Pruning: remote snapshots, the archive, the reference, and local copies

Model: `sonnet`. Depends on Task 4. **Neutral on 2.0 hooks.**

**Files:**
- Create: `internal/backup/prune.go`, `internal/backup/prune_test.go`
- Modify: `internal/backup/backup.go` (the reference and the prune phase
  after a successful send)

**Outcomes:**

- **The reference** (R-6bi16): after the send succeeds, the job writes the
  backup reference through the store, before any prune step.
- **The prune's safety rule** (R-6bi15): list `snapshots/` with `rclone
  lsjson`; skip every prune step, local deletion included, unless the
  listing holds the run's own `<stamp>.db` and the run's stamp is strictly
  greater than every other parsed stamp. A listing that exits 3 (`directory
  not found`) is empty.
- **Remote snapshots** (ruling 10): keep the run's own snapshot by name and
  the 13 newest of the rest, and delete the others with `rclone
  deletefile`. A name that does not parse as `<stamp>.db` is left alone.
- **The archive:** list `archive/`, and delete every `archive/<stamp>/`
  directory whose stamp is more than 90 days before the run's stamp with
  `rclone purge`. A directory name that does not parse as a stamp is left
  alone. Every `deletefile` and `purge` argument is built from a parsed
  stamp through R-6bi14's join.
- **Local copies** (ruling 7): delete every `pre-migrate-v<N>.db` and
  `pre-restore-<stamp>.db` in `snapshots/` whose modification time is
  earlier than this run's start. A copy written after the job began was not
  captured by it and stays. The reference is never touched.
- **A skipped or failed prune** is logged and written as the succeeded
  record's final `progress` text, and the job still ends succeeded, since
  the backup itself landed. `doctor`'s last-success row shows that text
  (Task 7).

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Listing | `snapshots/` or `archive/` absent (exit 3) | empty; no failure |
| Listing | the run's own snapshot missing | every prune skipped; the reason in `progress` |
| Listing | a stamp later than or equal to the run's | every prune skipped; the reason in `progress` |
| Remote snapshots | 14 or fewer, own included | nothing deleted |
| Remote snapshots | more than 14 | own kept by name, the 13 newest others kept, the rest deleted |
| Remote snapshots | a name that is not a stamp | untouched |
| Archive | more than 90 days before the run's stamp | purged |
| Archive | within 90 days | kept |
| Archive | a name that is not a stamp | untouched |
| Local copies | older than the run's start | deleted |
| Local copies | newer than the run's start | kept |
| Local copies | the reference, or any other name | untouched |
| Any prune step | rclone or `os.Remove` fails | logged; the job succeeded with the failure in `progress` |
| The run | failed or canceled before the send | no reference, no prune of any kind |

**Acceptance criteria:**

- Spec criterion 15: a real `local` remote pre-seeded with 15 stamped
  snapshots, an unparseable `notes.db`, two archive directories at 100 and
  10 days, and one named `keep-me`: after a successful run the remote holds
  exactly the run's own snapshot and the 13 newest seeded ones, `notes.db`,
  the 10-day archive, and `keep-me`.
- The clock row: the same remote seeded with 15 snapshots stamped one day
  in the future: after a successful run all 16 snapshots survive, and the
  record's `progress` names the skip.
- The first-run row: a remote with no `archive/` and no `snapshots/`: the
  run succeeds and its `progress` names no prune failure.
- Spec criterion 16: a leaf holding `pre-migrate-v1.db` and
  `pre-restore-20260101T000000.000000Z.db` dated before the run, a
  `pre-migrate-v2.db` dated after its start, and an operator file
  `notes.txt`: after a successful run only `pre-migrate-v2.db`,
  `notes.txt`, and the run's reference survive in `snapshots/`.
- The failed-run row: a remote seeded with 15 old snapshots and zero
  catalog tracks with a root missing one remote file, run against the real
  binary (criterion 13's shape): the run fails, all 15 snapshots and every
  local copy survive, and no reference is written.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Keep the 14 newest including own, by stamp alone → the clock row
   deletes the run's own snapshot.
2. Prune on a failed run → the failed-run row loses a snapshot.
3. Treat exit 3 on a listing as a failure → the first-run row names a
   prune failure.
4. Delete local copies by name pattern `*.db` → report whether an
   unrelated `.db` fixture is lost.
5. Measure archive age from the wall clock at prune time → report the row
   that fires, or name it reviewer-read.
6. Write the reference before the send → the failed-run row finds one.

**Deliverable count: 4** (the reference, the safety rule with remote
snapshot pruning, archive pruning, local deletion).

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
  `ops/local` with the flag, the remote, and `DUBPLATE_LIBRARY_ROOT`
  (R-6bi18), starts a backup as a local operator, polls its status, closes
  the runner, which waits for the job's goroutine (R-6bi11), and exits by
  Ruling 3. It prints the job's handle when it starts and its terminal
  state when it ends.
- **Signals** by R-6bi11: SIGINT, SIGTERM, and SIGHUP each call the cancel
  operation on a fresh deadline context derived from
  `context.WithoutCancel`. When that write fails, the verb cancels the job
  through the runner in-process and reports the write error. The verb
  keeps polling until the record is terminal and the runner's close
  returns.

**States each check meets, and its report:**

| Check | State | Report | Exit |
| --- | --- | --- | --- |
| Wall | a retired name set | the load error naming its message | 1 |
| Wall | flag set, remote unset | the load error naming the remote key | 1 |
| Wall | flag unparseable | the load error naming the flag | 1 |
| Store | absent | "no store ... start serve first"; nothing created | 1 |
| Store | any refusal of the write mode | its reason | 1 |
| Start | flag off | unavailable, reason "backup integration disabled" | 69 |
| Start | conflict | the running job's handle, or the orphaned-transfer text | 75 |
| Job | succeeded | the handle and "succeeded", plus any prune text | 0 |
| Job | failed | the handle, the code, the message | 1 |
| Job | canceled, by a signal or another process's cancel | the handle and "canceled" | 130 |
| Signal | the cancel write fails | the write error; the job canceled in-process | 130 |
| Job | interrupted (unreachable in-process, since the verb's own runner holds the lease; kept so the state set is closed) | the handle and "interrupted" | 1 |

**Acceptance criteria:**

- Spec criterion 18: `backup` against a state directory with no leaf exits
  1 naming "start serve first", and the directory listing is identical
  before and after.
- Spec criterion 19: setting each of the six retired names fails the load
  with its message verbatim, each by its own table row. The template
  classification test (`templatekeys_test.go`) still passes, now reading
  the six names as superseded.
- Spec criterion 20: the verb runs as a re-executed process started with
  `Setpgid` (its own group leader), against the SIGINT-ignoring stand-in,
  which writes its pid to a file the test names. Three rows send SIGINT,
  SIGTERM, and SIGHUP to the verb's pid alone; each ends the job canceled,
  the stand-in's pid is gone once `WaitDelay` has passed, and the verb
  exits 130. A fourth row sends SIGINT to the verb's process group, as a
  terminal does, and after the verb exits 130 asserts that the stand-in's
  record file holds exactly one SIGINT line: the runner's `Cancel`. Without
  `Setpgid` the terminal's group signal also reaches the stand-in and adds
  a second line, so the row fails by count, with no timing window to
  flake. (Revision 2's 500 ms liveness check passed with `Setpgid` dropped,
  since an ignoring stand-in survives a group SIGINT either way.)
- Each exit code maps to its outcome as the verb's help states: a table of
  the five outcomes, each driven by its real trigger (flag off, a held kind
  lease, stand-in exits 0 and 7, SIGINT), asserts the code and that the help
  text names it, including "canceled by any means".
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Start the job on the verb's own context → criterion 20 ends failed, not
   canceled.
2. Cancel on the done context → the write fails at once and the
   in-process fallback masks it in the outcome, so criterion 20's rows
   also assert that the verb prints no cancel-write error.
3. Ignore SIGHUP → the SIGHUP row's verb dies by the signal's default and
   exits 129, not 130.
4. Map every error to exit 1 → the exit-code table fails.
5. Exit when the record turns terminal, before the runner's close → report
   whether the stage survives in any row, or name it reviewer-read.
6. Leave the reserved class in `keys.go` → registration panics (report the
   test that fires).
7. Drop `Setpgid` from the transport (Task 4's mutation 5) → the group
   row's record file holds two SIGINT lines.

**Deliverable count: 4** (the class, the supersessions with the reserved
class's removal, the verb with its exit codes, signal handling).

---

## Task 7: `doctor`'s backup rows

Model: `sonnet`. Depends on Tasks 1B and 6. **Neutral on 2.0 hooks.**

**Files:**
- Modify: `cmd/dubplate-server/doctor.go`, `doctor_test.go`,
  `cmd/dubplate-server/store.go`, `store_test.go`, `backup.go`,
  `backup_test.go`

**Outcomes:**

- **The remote rows**, only with the flag on: the value passes R-6bi14's
  validator (`internal/backup`), the same function the job calls.
  `RCLONE_CONFIG_<NAME>_TYPE` (uppercased name) must be set. For `crypt`,
  `RCLONE_CONFIG_<NAME>_REMOTE` must be set; when it names a remote
  (`<other>:...` with a non-empty name), that remote's own `_TYPE` must be
  set; when it is a plain path (no colon) or a connection string (starting
  with `:`), the wrapped check is skipped and the features probe decides,
  since rclone accepts both (probed). The features probe runs last
  (Ruling 2), with `--low-level-retries 1`, under R-6bi12's precedence: a
  deadline hit reads "not checked" whatever the exit. A non-crypt type
  carries ruling 14's warning, never a failure.
- **The store rows** from R-6bi13: last backup success age or "never",
  with the record's final `progress` text when it names a skipped or
  failed prune; the `-wal` size; free space; and the local copy count and
  size.
- **Every row renders in a golden** from a constructed input, as 6a's store
  row does (R-6a16).

**States each check meets, and its report:**

| Check | State | Row | Exit |
| --- | --- | --- | --- |
| Flag | off | backup: "unavailable: backup integration disabled" | 0 |
| Remote | no colon | "invalid remote: want <name>:<path>" | 1 |
| Remote name | empty (`:path`) | "invalid remote: want <name>:<path>" | 1 |
| Remote name | a character outside letters, digits, underscore | names the character and rclone's rule | 1 |
| Remote path | empty (`name:`) | accepted | 0 |
| `_TYPE` | unset | names `RCLONE_CONFIG_<NAME>_TYPE` | 1 |
| Crypt | `_REMOTE` unset | names `RCLONE_CONFIG_<NAME>_REMOTE` | 1 |
| Crypt | wrapped remote's `_TYPE` unset | names the wrapped remote and its key | 1 |
| Crypt | `_REMOTE` a plain path or a connection string | wrapped check skipped; the probe decides | as the probe |
| rclone | not on `PATH` | "rclone not found" | 1 |
| Features | `Copy` and `Move` both false | names the backend and rclone's `--backup-dir` requirement | 1 |
| Features | the probe exits nonzero before the deadline | its last stderr line | 1 |
| Features | the deadline hit, any exit | "not checked: rclone backend features timed out" | 0 |
| Crypt | the type is not `crypt` | the ruling 14 warning appended | 0 |
| Last success | no succeeded backup record | "never" | 0 |
| Last success | a record | its age, rounded to minutes | 0 |
| Last success | a record whose `progress` names a prune skip or failure | its age and that text | 0 |
| WAL size | `-wal` absent | "0 B" | 0 |
| WAL size | present | its size | 0 |
| Free space | `statfs` succeeded | the free bytes | 0 |
| Free space | `statfs` failed | "unknown" | 0 |
| Local copies | none | "0 copies" | 0 |
| Local copies | one or more | the count and total size | 0 |
| Store | absent, older, newer, or not a store | "not checked" on the store-fact rows | as the store row |

**Acceptance criteria:**

- Spec criterion 21: one golden line per row state above, each from a
  constructed input, plus three end-to-end rows against a real `rclone`: a
  valid `local` remote, an `http` remote at a loopback URL
  (`http://127.0.0.1:9/`; the fold's probe measured `rclone backend
  features` answering it with both false, exit 0, in 0.27 s with nothing
  listening), both features false, and a crypt remote whose wrapped remote is undefined.
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
5. Refuse a crypt `_REMOTE` holding a plain path → its golden exits 1.

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
  gate's wall time with the rclone tests), the resolutions R-6bi1 to
  R-6bi18 with where each landed, the four rulings and their outcomes, the
  declared coverage gaps, both budgets, and "what a later pass would be
  wrong to rediscover," which carries at minimum R-6bi7's measurement,
  R-6bi12's orphan shape with its SIGPIPE-bounded window and the `Wait`
  precedence, the checkpoint's `busy_timeout=0` (R-6bi6), the remote join
  (R-6bi14), and the `lsjson` exit 3 (R-6bi15).
- **The errata owed**, paste-ready: the spec deviations, and the spec's
  "Not verified" rows this pass settled (`--max-delete`, the feature table,
  `os/exec`'s wording, rclone's exit codes and its exit 130 on SIGINT). The
  fold record's "Owed errata" section lists them.
- **The 6b-ii handoff:** the ops interface and codes the projection
  mirrors; `forbidden`, `internal`, and the network actor with its
  `internal/identity` import, which 6b-ii declares with their first caller;
  the catalog operation's first production caller; the lease and
  kind-lease semantics restore must respect; the staging rule; the remote
  layout and join restore reads; the backup reference, which a restore
  from a snapshot should replace.
- **Forward items for later rungs:** to 6d, a store identity minted at
  create, so a backup can refuse a remote another store writes to (risk
  R9's option b); to rung 7, per-collection and per-extension counting in
  the guard, a `--max-delete` sized as `ceil((1 − fraction) × tracks)`
  with an operator flag for a planned mass change, and a regression floor
  on remote pruning (risk R10's option b, risk R1's option b); to 6c, the
  `SAVEPOINT` check on migration files that also catches `COMMIT; BEGIN;`
  (risk R13).
- **The CLAUDE.md Gate line** for `rclone`, and the runbook lines owed to
  rung 16: `KillMode=mixed`; rclone pinned at or above the measured
  v1.74.3 on the box; rclone variables set by hand until the template rows
  land; a crypt remote recommended; one remote path per server; the remote
  value's path part names where the backup lives, bucket included; the
  non-empty wrong-mount gap; with zero tracks, any library deletion fails
  every run with exit 7 until tracks exist, remedied by archiving the file
  on the remote by hand with `rclone moveto`; removing a backup reference
  by hand is how an operator deliberately starts a fresh store; and a
  restore before rung 12 loses stars and playlists.

**Acceptance criteria:** every count in the drafts was produced by a
command the report pastes; every commit id resolves; the STATUS draft is at
most 60 lines. No gate change, so `bash scripts/check.sh` green as a
formality.

**Deliverable count: 4** (the HISTORY draft, the errata, the handoff with
the forward items and the Gate and runbook lines, the STATUS draft).

---

## 6a's carried review items, disposed

- **A migration file carrying its own `COMMIT` is caught only at Commit,
  after its tail ran in autocommit** (`migrate.go:340`). **Declined for
  6b-i.** The gate applies every embedded file through the runner against
  scratch stores (`TestSchemaSkeletonSatisfiesInvariants`,
  `TestSchemaMigrationsCarryPopulatedFixtures`), so an embedded `COMMIT`
  fails the gate before merge, and its tail only ever runs against a
  scratch database. Refusing it before the tail runs needs a SQL tokenizer:
  a `CREATE TRIGGER` body legitimately carries `BEGIN ... END`, so a keyword
  scan refuses valid files. The driver exposes no autocommit query
  (`conn.go`, v1.59.0, read 2026-09-24). One case passes the gate
  silently: a file carrying `COMMIT; BEGIN;`, whose runner `Commit` then
  commits the file's own second transaction and returns nil (risk R13). A
  `SAVEPOINT` issued before the body and released after it turns that case
  loud without a tokenizer; 6b-i adds no migration, so it routes to 6c,
  the next pass that adds one (Task 8's forward items). The per-file
  `diff-reviewer` checklist line (6a Task 4) stays the net until then.
- **A failed snapshot rename leaves staging occupied until serve's
  cleanup** (`snapshot.go:73`). **A criterion in Task 1B.**
- **The rebuild marker matches exact bytes only** (`migrate.go:163`). **A
  criterion in Task 1B:** a near miss is refused at `readMigrations`,
  which every test calls, so a bad embedded file fails the gate.
- **Serve's unbounded settle loop and unguarded logger**
  (`open.go`, `openServe`). **The logger guard lands in Task 1A on Ruling
  4's yes**, since an exported open must refuse a nil logger; **the settle
  loop routes to 6d**, as the conductor directed. 6b-i's own logger
  parameter refuses nil at construction (Task 3).

---

## The exit gate

`bash scripts/check.sh` green in the worktree and again on master after
the merge, plus every assertion below. Each is a spec 6b-i criterion
(`:990-1033`).

1. Criterion 1: Task 2.
2. Criterion 2: Tasks 2 and 3.
3. Criteria 3 to 8 (the job model and availability): Task 3.
4. Criterion 9 (the snapshot's content): Tasks 1B and 4.
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
    package and tests (the exported-unused leg's report, pasted), with
    exactly two named exceptions: serve mode's open (Ruling 4's yes; its
    6b-i callers are tests) and the catalog operation (spec criterion 8
    requires it; 6b-ii's projection is its first production caller).

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's
ownership table before the lane launches.

| Surface | Owner | 6b-i's position | Contention |
| --- | --- | --- | --- |
| `internal/engine/store` | rung 6 | Tasks 1S, 1A, and 1B | none; 5c-ii leaves it untouched |
| `internal/ops`, `internal/ops/local`, `internal/backup` (new) | rung 6 | Tasks 2 to 5 | none |
| `cmd/dubplate-server/main.go` | rung 1 | Task 6: one registration line, one `AddCommand` entry, main's exit branch | **5c-ii Task 5 adds one line in `newProductionRegistry`**; a textual conflict keeps both in landing order; the second lander's rebase is non-fast-forward |
| `cmd/dubplate-server/doctor.go`, `doctor_test.go` | rung 1's skeleton | Task 7 | 5c-ii adds threshold rows through the registry; `TestDoctorCmd_LineCountFollowsRegistry` derives them, and the second lander regenerates goldens in its rebase |
| `cmd/dubplate-server/store.go`, `backup.go` (new) | rung 6 | Tasks 6, 7 | none |
| `internal/config/keys.go` | rung 1b's contract | Task 6 removes the "backup targets" reserved class | 5c-ii leaves `internal/config` untouched; 6c's lane condition is about 5c and 5d, so this edit does not change it |
| `scripts/check.sh`, `Makefile`, `web/` | rung 5 | untouched | after 5c-ii merges, the lane's rebase needs `npm ci --prefix web` |
| `CLAUDE.md` Gate section | conductor | Task 8 drafts the `rclone` line | conductor-committed in the main tree |
| The state directory layout | rung 4 | `leases/` joins the store leaf; `snapshots/` gains the backup reference | 6d grants the leaf on the box |
| `docs/` ledgers | conductor | Task 8 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: confirm 6a merged; record the planning spend against the
   rung; create the lane 2 worktree from master; write the lane registry
   row and the contested-surface rows; arm the runaway guard; check AC power.
2. Before Task 1S: rule Task 0's blockers and record the four rulings'
   outcomes.
3. At segment A's boundary: STATUS; re-derive the ceiling from Tasks 1S to
   2. At segment B's: STATUS; re-derive again with Task 3's report, and
   take the pass split below if the projection passes the ceiling.
4. At segment C's boundary: STATUS; one `code-simplifier` round over Tasks
   1S to 7's accepted Go, pasting `.claude/instructions/simplifier-brief.md`
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

Workflow mode: the pass-execute chain per task, **sequential**, in **four
segments** of at most four tasks (Task 1S made revision 2's segment A five
long, so segment B splits at the pass's pre-named split point). Every task runs `dubplate-implementer`, then `diff-reviewer`
(`claude-opus-5-5`), then the gate, with one re-dispatch on `fix` and a
second `fix` as the conductor's decision. Every reviewer dispatch asks for
correctness gaps ranked by consequence, not polish.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1S, 1A, 1B, 2 | STATUS. Re-derive the ceiling. |
| **B** | 3, 4, 5 | STATUS. Re-derive the ceiling; the pass split, if it must. |
| **C** | 6, 7 | STATUS. The simplifier round. |
| **D** | 8 | The close drafts, the merge ritual, the architecture reads, the ledger fold. |

Each boundary sits on a task's own green committed gate.

**Models.** `sonnet` by default. **Four upshifts to `opus`, all
unconditional**, dispatched as `claude-opus-5-5`: Task 1A (the lease
primitive, whose wrong variants pass every single-descriptor test, and the
inherited descriptors); Task 1B (the backup copy, the data path hook 10
names, and the reference the lost-database guard reads); Task 3
(cross-process recovery and cancel, whose races a single-process test
passes); and Task 4 (the guards and `--max-delete`, the data-loss path,
and R-6bi12's orphan shape). Task 1S stays `sonnet`: its changes are
mechanical and its existing tests pin behavior. Task 0 runs at `opus` in
fresh context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1S.
- Task 1S → Task 1A (the reshaped signatures Task 1A exports).
- Task 1A → Task 1B (the lease and the descriptors the copy's callers
  use).
- Tasks 1A and 2 → Task 3 (the write surface and leases; the interface).
- Tasks 1B and 3 → Task 4 (the copy; the runner and the body contract).
- Task 4 → Task 5 (the job the prune phase joins).
- Tasks 3, 4, and 5 → Task 6 (the verb runs the whole job).
- Tasks 1B and 6 → Task 7 (the report fields; the class).
- Tasks 0 to 7 → Task 8.
- **Independent by files:** Task 2 of Tasks 1S, 1A, and 1B. They still run
  sequentially, because one tree runs one executor and the module gate
  runs over `./...`.

**Where this pass splits, if it must:** at the Task 5 / Task 6 boundary,
segment B's end, if segment A's or B's re-derivation projects past the
ceiling. Tasks 6 and 7 then
open a pass 6b-i-2, because the verb and `doctor` rows are what give the
exports their production caller; 6b-i-1 would close with its exports
warned by the exported-unused leg, which the conductor records. Inside the
pass, Task 3 carries its pre-named split.

**Budget:** 8.2M execution (Ruling 1), flag 6.56M, planning recorded
separately, STATUS at each boundary, spend scored at close against the
ceiling and against zero attended interactions after approval.

---

## Prior art consulted

Scored on CLAUDE.md's four criteria; the spec's table (`:895-915`) carries
the evidence for the rows it scored.

| Entry | Verdict for 6b-i |
| --- | --- |
| rclone v1.74.3, subprocess | **Adopt** (spec). This plan adds R-6bi7's, R-6bi9's, R-6bi12's, R-6bi14's, and R-6bi15's measurements. |
| `librclone` | **Decline** (spec): rclone's whole dependency tree for one verb. |
| SQLite `VACUUM INTO` | **Adopt** for the backup copy (Task 1B); the driver's backup API stays the recorded alternative. |
| `internal/atomicfile` | **Not used** at 6b-i: the store writes its own reference file durably, as its pre-migration path does, and the intent record is 6b-ii's. |
| `_salvage/internal/backup` | **Adopt the missing-root guard in reasoning**; the count guard replaces its top-level check; import nothing. |
| A Go job-queue library | **Decline**: it would own a table the spec already designed, and none speaks cross-process `flock` leases. Not measured further, since no candidate was in the spec's list. |
| OpenSubsonic error codes | **Consulted, declined as a model** (Task 2). |
| Navidrome, gonic, taglib, mewkiz/flac, Picard, chromaprint, goose, golang-migrate | No 6b-i surface. |

---

## Not verified in this revision

- Whether every rclone backend accepts `--metadata-set`; a `local` remote
  on tmpfs did, and Task 4's criterion 10 re-proves it there. The manual
  read here does not say whether a backend without metadata support drops
  the flag or errors; if one errors, the send fails loudly on that backend,
  which Task 7's features row does not predict.
- Whether `rclone backend features` contacts the provider for every
  backend; Ruling 2's 10 s bound and `--low-level-retries 1` cover the case
  where it does (a bad-credential B2 remote measured 2.1 s).
- `--max-delete`'s counting of `--backup-dir` moves was measured on the
  `local` backend only; a copy-only backend (S3, B2) archives by
  server-side copy and delete, which no probe exercised.
- The box's rclone version: nothing in 6b-i reads it; rung 16's runbook
  pins it at or above v1.74.3.
- A forward clock jump still purges archive directories early (R-6bi15);
  the safety rule catches a backward jump only.
- The 6a close-side spend (about 2.1M) is STATUS's re-derived projection,
  not a recorded actual; segment A re-derives this ceiling from it.
