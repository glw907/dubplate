# Rung 6, Pass 6a (The Store Foundation) Implementation Plan

Status: **revision 2.2, 2026-09-23; approved by Geoff 2026-09-24 with
all four rulings yes.**
Revision 1 (`11f11c2`) took three lens reviews (contract, mechanics, and
risk). Their folds, refusals, and owner forks are recorded in
`docs/research/2026-09-23-rung-6a-plan-fold-record.md`.
Revision 2 (`1d69ab8`) took a fresh-context verification read
(`docs/research/2026-09-23-rung-6a-plan-fold-verification.md`,
V1 to V9); revision 2.1 folds it, and the fold record's "Second fold"
section gives one disposition per finding. Revision 2.2 folds a prose and
fact review (`docs/research/2026-09-23-rung-6a-plan-prose-review.md`,
B1 to B2, W1 to W10, S1 to S11); the fold record's "Prose review fold"
section gives one disposition per finding. Four rulings for Geoff are
open ("Rulings for Geoff").

Sources: the ratified rung 6 spec
(`docs/specs/2026-09-23-rung-6-ops-seam-store.md`, revision
2.2, all fourteen rulings yes, last touched at `b406f93`) and its fold
record (`docs/research/2026-09-23-rung-6-spec-fold-record.md`,
whose "Revision 2.2" section carries the contested-surface rows and the
ceiling arithmetic); the 2.0 streamer spec
(`docs/specs/2026-09-07-streamer-2-0-design.md`, "Forward
hooks" and "Prior art"); the configuration homes spec, revision 3
(`docs/specs/2026-09-11-configuration-homes-design.md`); the
approved 5c-i plan (`docs/plans/2026-09-23-rung-5c-pages-member-verb.md`),
which runs in lane 1 beside this pass; and `docs/HISTORY.md`'s top entry
(pass 5b-ii) for the spend basis. Code pointers were resolved at
`cb9ecc3`; the Go tree is unchanged since `08befea`.

**What this pass is.** Pass 6a builds `internal/engine/store`: the store
leaf and its lock, the SQLite driver and three connection classes, the
one write method, the migration runner with its pre-migration snapshots,
the first three migrations with the schema-invariant test, and the three
open modes. It lands the first production callers of the store, `doctor`'s
store row and `check`'s inspect report, and the busy-timeout threshold
they read. The pass is Go only and runs in the lane 2 worktree.

**What this pass is not.** No `internal/ops`, no job runner, no lease, no
`backup`, and no `restore`: those are 6b-i and 6b-ii. No settings table,
no registry metadata, and no seed rule: those are 6c. No roster import,
no `Identity.UserID`, and no `serve` wiring: those are 6d. **Nothing in
production creates or migrates a store in this pass.** `doctor` and
`check` open it read-only; the serve mode is tested in-package and gains
its caller at 6d. No `systemd/`, `scripts/`, or `env/` change, and **no
deploy of any kind.** A task that believes it needs any of these stops
and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process",
> the pass-execute chain per task, sequential, in the three segments
> "Execution model" names. The implementer is `dubplate-implementer`. The
> conductor consumes structured reports and never reads a diff, a test
> log, or a gate transcript inline. Subagents never dispatch agents, so
> `code-simplifier` is the conductor's dispatch. The conductor owns
> `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the lane registry,
> and the contested-surface rows. **This pass runs in lane 2, a git
> worktree, beside 5c-i in lane 1.** The live-executor check is the lane
> registry, never `pgrep -f Projects/dubplate`.

---

## Rulings for Geoff

Four yes-or-no questions. The spec's fourteen rulings stand and are not
reopened.

**1. Does 6a run under a 6.8M execution ceiling, with planning recorded
separately?** **Ruled yes (Geoff, 2026-09-24).** Recommendation: yes. Reason: the spec's 5.5M figure is a
total that includes 0.95M of planning and a 0.46M chain rate, while
5b-ii's recorded actual is 0.533M per task with everything amortized, the
basis the approved 5c-i plan uses. On that basis, with the review fold's
added state rows, 6a projects to 4.92M of execution ("Token ceiling").
What each answer builds: yes sets the flag at 5.44M, about one task above
the projection, and moves rung 6's reservation up by about 2.3M before
6b-i re-derives. No holds the spec's 5.5M total, a 4.55M execution
ceiling that sits under the projection, so the flag fires on plan and the
conductor asks mid-pass.

**2. Does `doctor` exit 0 when the store is absent, older, or current,
and nonzero for every other state?** **Ruled yes (Geoff, 2026-09-24).** Recommendation: yes. Reason: nothing
creates a store in production until 6d, so a nonzero exit on "absent"
would fail 5d's carried "doctor exits 0" exit gate on every box, and an
older store is the normal state after an upgrade that `serve` migrates.
An interrupted create exits nonzero because it is not yet a store `serve`
can start through: it needs a completion step, and its report names the
fix. **What yes builds:** `doctor` and `check` exit nonzero for a newer
store, a restore in progress, a file that is not a store (an interrupted
create included), a lock held exclusively, a refusal of any kind (owner,
filesystem, ledger, an unreadable file), and exit 0 for absent, older,
and current (R-6a12). **What no builds:** every state but "current"
exits nonzero, and 5d's exit gate then needs a pre-created store, which
only `serve` may create.

**3. Does `serve` refuse to create or complete a store when
`snapshots/` holds a copy and the database is absent or an interrupted
create?** **Ruled yes (Geoff, 2026-09-24).** Recommendation: yes.

**Reason:** serve creates a store when the database file is absent, and
completes one when the file is zero bytes with no `-wal` or header-only
and unmarked (R-6a6). A database lost to a hand deletion, a mistaken
`mv`, a failed copy, a `: > store.db`, or a disk repair meets one of
those rows, and would be replaced by an empty catalog without a word;
after 6b-i the next backup would upload that empty store as the latest.
The check costs one directory read in serve's step 3 and one inspect
branch. A genuine first create or interrupted create never meets a
copy, because a store at version 0 takes no snapshot (R-6a11), so the
check costs a fresh install nothing.

**What yes builds:** serve, holding the exclusive lock, reads
`snapshots/` before it creates a store or completes an interrupted
create. With a `pre-migrate-v<N>.db` or `pre-restore-*.db` present, serve
refuses naming the newest copy and `restore --from-file`, leaves the
database file as it found it, and inspect reports "refused" with the same
text, which exits nonzero under Ruling 2 (R-6a5). With no copy present, a
lost database is still created or completed empty; the guard covers only
a leaf that kept a copy.

**What no builds:** revision 1's rule stands. Serve creates whenever the
database file is absent and completes every interrupted create. A third
option, an explicit `serve --create-store` for any create after the
first, is not recommended, because it adds an attended step to every
fresh install.

**4. Does a caller's deadline bound only the in-process wait, leaving
SQLite's wait on another process's write lock bounded by the busy timeout
alone?** **Ruled yes (Geoff, 2026-09-24).** Recommendation: yes. Reason: a probe measured a `BeginTx` with a
300 ms context and a 3 s busy timeout returning `database is locked (5)`
after 3.01 s, because SQLite's busy handler does not check the interrupt
the context triggers (R-6a18). The spec already says the in-process wait
is bounded by the context "and never by `busy_timeout`"; at group scale
the writers are one `serve` and short CLI verbs, so a 15 s worst case is
rare and visible. What each answer builds: yes documents the measured
behavior, pins it with a test, and adds no mechanism. No makes the write
method pin a `*sql.Conn` and set `PRAGMA busy_timeout` to the smaller of
the time left and the configured value before each `BEGIN IMMEDIATE`,
which adds a statement per write and hand-manages the pool's one
connection; Task 3 then carries a fifth deliverable and splits. That
split costs about 0.53M and brings the projection to 5.45M, just past
Ruling 1's 5.44M flag, so the conductor asks at that boundary under the
80% rule.

### Spec deviations that take effect on approval

These change the ratified rung 6 spec's meaning. Approving this plan
approves them; the close drafts the erratum text (Task 8), and the fold
record lists them as owed errata.

- R-6a3: the migration wait registers at 6d, not 6a (the spec fold
  record's `main.go` contested row).
- R-6a2: the spec's "6a exports the open" reads as the inspect mode.
- R-6a6: an interrupted create (a zero-byte file, or an unmarked
  header-only file) is completed by serve mode, and criterion 13's "not a
  store" holds for the inspect and write modes.
- R-6a7: only serve mode, holding the exclusive lock, removes staging
  leftovers; the spec's "every writing open" does not.
- R-6a9: the path refusal lands as schema CHECKs at 6a; the Go validator
  lands with rung 7's first writer.
- Task 1: the filesystem probe refuses SMB, Ceph, and AFS beside the
  spec's list, runs on `db/` when it exists, and a create requires the
  effective uid to own the state directory.
- Task 6: serve's step 1 shared wait is bounded by the migration wait and
  logs the holders, as step 2 does.
- Ruling 1's yes: the ceiling table row's figure becomes 6.8M execution
  plus planning.
- Ruling 3's yes: serve refuses a create or a completion when
  `snapshots/` holds a copy, naming it and `restore --from-file`, and
  inspect reports it as refused.
- Ruling 4's yes: the write method's contract reads that the busy
  timeout alone bounds the wait on another process's write lock, and a
  context does not shorten it.

---

## Primary inputs

Read these before Task 1, in this order.

1. **The rung 6 spec, revision 2.2.** The sections this pass implements:
   "What rung 6 builds" and "What rung 6 does not build" (`:141-189`),
   "Package territory" (`:193-216`), "The store" in full (`:220-494`:
   layout, ownership, and the lock; the three open modes; connections and
   transactions; migrations; schema invariants; the schema rung 6 lands;
   paths), the 2.0 hooks (`:842-856`), the prior-art table (`:860-880`),
   the privileged-field table (`:884-902`), and 6a's sixteen criteria
   (`:911-949`).
2. **The fold record's "Revision 2.2" section**: the verification ledger
   (which behaviors were measured and where), the risk model, and the
   contested-surface rows the conductor writes into PA's ownership table.
3. **The configuration homes spec**, "Collections carry their shape"
   (`:184-208`) and the tombstone rule (`:515-526`), which fix what the
   `collections` and `users` rows carry.
4. **This plan's quoted behaviors.** Every relied-on library, stdlib, or
   kernel behavior below is quoted from `go doc`, the module source in
   `$(go env GOMODCACHE)`, a man page, or a scratch probe run on
   2026-09-23 against `go1.27.1 linux/amd64` and `modernc.org/sqlite
   v1.59.0` (SQLite 3.53.4). The probes live under
   `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/82a7130f-dc63-4c92-b308-9af3d3bb0dde/scratchpad/`
   (`6a-probes/` for revision 1; `6a-review-probes/`, `6a-risk-probes/`,
   and `6a-fold-probes/` for the review and its fold) and may be gone; each task re-proves what it relies on through its own
   gate-resident test. A task that relies on a behavior this plan does not
   quote confirms it the same way and reports what it read.

---

## Goal

After this pass, a store can exist and be trusted. One SQLite file opens
through three connection classes whose every setting is proved on a fresh
connection. One write method commits only a clean callback and refuses a
store whose version moved under it. A migration either applies whole or
leaves the store as it was, and a non-empty store is copied before it
migrates. The schema-invariant test binds every later rung's migration to
rules SQLite's own catalog decides. `doctor` and `check` report the
store's state on the box without changing a byte of it.

---

## Owner touchpoints

The four rulings above, at the plan-approval gate. Nothing else is
expected. What would become one: a second new module dependency (this
pass adds exactly `modernc.org/sqlite`); a threshold default changed on a
task's own judgment; any change to what the box runs.

---

## Token ceiling: 6.8M execution for 6a (Ruling 1)

The basis is 5b-ii's recorded "about 6.4M" (`docs/HISTORY.md`, top entry,
"Both budgets") over the twelve tasks it bought: **0.533M per task**, with
Task 0, fix rounds, simplifiers, the close, and the conductor's turns
amortized in the rate. This is the basis the approved 5c-i plan uses.
Planning is spent before Task 0 dispatches, so the conductor records its
actual against the rung and outside this ceiling.

**The arithmetic.**

1. Eight tasks (Tasks 1 through 7 and the close) at 0.533M is **4.27M**.
2. Three of the eight run at `opus` (Tasks 3, 4, and 6), the same share
   as 5c-i: add **0.05M**.
3. The first SQLite code in the repo: a new module dependency, a
   first-build compile of the transpiled driver inside every gate run,
   and the driver's error model met for the first time. Add **0.20M**.
4. Two `go-architecture-reader` dispatches (`internal/engine/store`,
   `cmd/dubplate-server`): **0.10M**.
5. The review fold (revision 2): about two dozen named states and three
   reworked proofs (criterion 8's failure point, criterion 14's ordering,
   and criterion 15's hot WAL), concentrated in Tasks 3, 4, and 6, with no
   new task and no deliverable count changed. Add **0.30M**, about half a
   task at the measured rate.
6. **The projection is 4.92M.**

A ceiling of **6.8M** puts the 80% flag at **5.44M**, 0.52M above the
projection, about one task at the measured rate, so the flag marks a real
overrun rather than firing on plan. Either pre-named task split (Task 6's
serve mode, or Task 3's under Ruling 4's no) adds about 0.53M and brings
the projection to 5.45M, just past the flag, so the conductor asks at
that boundary under the 80% rule rather than absorbing it. Both splits
together project 5.98M, still under the ceiling. The spec's table row
reads 5.5M, which is 0.95M of planning plus a 4.55M execution figure
(5.5M minus 0.95M) at 0.46M per task; the 0.46M rate subtracts planning
and Task 0 that 5b-ii's recorded actual already amortizes. Re-derive at
the segment A boundary from Tasks 1 through 3's reports.

Checkpoint: STATUS at every segment boundary, after Task 3, after Task 7,
and at the close. Every task reports its spend, its mutation outcomes,
and its measured gate wall time.

Deliverable count: thirty-one across Tasks 1 through 8, plus Task 0's
report. Tasks 1 through 6 and Task 8 carry four each; Task 7 carries
three. No task passes four. The review fold changed no task's count;
Task 6's four are re-cut (staging cleanup joins the serve mode, and the
shared classification takes its place).

---

## Global constraints

- **Lane 2, a git worktree.** The conductor creates it from master before
  Task 0 and writes the lane registry row. Every task, gate, and commit
  runs inside the worktree. `scripts/check.sh` runs `make check` over
  `./...`, so two executors in one tree would each gate the other's
  half-written work.
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`,
  and no task ssh's to the box.
- **Package territory, stated as a wall.** This pass creates
  `internal/engine/store` (every file, its `migrations/` directory, and
  its `testdata/`). It creates `cmd/dubplate-server/store.go` and its
  test, and edits `cmd/dubplate-server/doctor.go`, `check.go`, `main.go`
  (one line in `newProductionRegistry`), and their tests. It edits
  `go.mod` and `go.sum` for one requirement. **It writes nothing under
  `internal/auth`, `internal/identity`, or `internal/config`**, and
  nothing under `internal/serve`, `internal/notify`, `internal/harness`,
  `internal/atomicfile`, the rest of `internal/engine`, `web/`,
  `systemd/`, `scripts/`, `env/`, `config/`, `Makefile`, `CLAUDE.md`,
  `corpus/`, `tools/`, or `_salvage/`. It writes no `docs/` file outside
  Task 8's drafts. A task that believes it needs one stops and reports.
- **One new module dependency: `modernc.org/sqlite v1.59.0`**, the newest
  version (`go list -m -versions`, 2026-09-23) and the one the spec
  scored. It brings `golang.org/x/sys` and the `modernc.org/*` runtime as
  requirements; no task imports `golang.org/x/sys` directly. File locks
  use stdlib `syscall.Flock`, and the filesystem probe uses stdlib
  `syscall.Statfs`.
- **The engine has no HTTP knowledge and reads no environment.**
  `internal/engine/store` takes every value it needs as a parameter. It
  imports the driver, `internal/atomicfile` only if a task needs it, and
  the standard library. It never imports `internal/config`.
- **No process that holds a store connection opens the database, `-wal`,
  or `-shm` file through `os`** (R-6a4). Metadata comes from `os.Stat`,
  which opens nothing.
- **Every DSN comes from one builder that escapes the path, and no DSN
  can create a database** (R-6a17). The pre-create is the only way a
  database file comes to exist.
- **Nothing is exported ahead of a cross-package caller** (R-6a2).
- **Assertion discipline at the repo's bar.** No silent-success fakes, no
  vacuous assertions, no expectation derived from the code under test.
  Every hard-coded expectation cites the spec, a quoted document, or a
  probe this plan records. Every criterion names the fixture state it
  needs, so a test cannot pass on an empty store.
- **Universality.** No estate hostname, path, email, or personal name
  anywhere. Every tunable is a threshold with a default or a named
  unexported constant whose value this plan states.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar. `ctx`
  first where a call can block, which includes every lock wait and every
  statement. Errors lowercase and context-first. Sentinels only where a
  caller branches, asserted through `errors.Is`. **The em dash is banned
  in Go comments.** No process citation (a pass, a task, a ruling id, a
  plan line) in a shipped comment; the citation leg is staged and warns
  today.
- **`-race` cost.** The driver is transpiled C; tests that open many
  connections under `-race` are slow. Each task reports its package's
  test wall time, and no test sleeps where a lock or channel can order
  events.
- **No `t.Parallel()` in `internal/engine/store`.** `syscall.Umask`
  changes the whole process, and the 0600 rows (Tasks 1, 4, and 6) set
  it, so a parallel test could create a file under the wrong umask.
- **The gate is `bash scripts/check.sh`, green at every task end**, run
  through `cairn-run-gate` inside the worktree, re-issued on exit 75
  until it prints `gate exit:`. Anything reported as evidence runs with
  `-count=1` and reports wall time. Specific files only. Commit footer
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` for an `opus`
  task and the dispatch's footer otherwise.

---

## Privileged fields and their channels

CLAUDE.md's planning-miss rule asks that every privileged field name every
channel that may set it. At 6a most privileged fields exist as columns
with no writer; each row says so, so no task adds a writer by accident.

| Field | Channels that may set it at 6a | What refuses every other channel |
| --- | --- | --- |
| `user_version` and the migration ledger | The migration runner alone, inside the serve mode's exclusive lock (Tasks 4, 6). | The write method refuses any callback whose store version differs from the binary's (Task 3). The read pool is `query_only` (Task 2). The inspect mode opens `mode=ro` (Task 6). |
| The store's creation | The serve mode alone (Task 6), through the pre-create, which has no production caller until 6d. 6b-ii's restore is the later second channel, and places a file from a snapshot. | The driver opens every class with `SQLITE_OPEN_CREATE` (`conn.go`, v1.59.0), so SQLite itself is a channel; every writing DSN carries `mode=rw`, which a probe measured refusing an absent file with `unable to open database file (14)` and creating nothing, and inspect carries `mode=ro` (R-6a17). The write mode refuses an absent store naming "start serve first". Classification by `os.Stat` runs before any open. Under Ruling 3's yes, serve also refuses a create beside a snapshot. |
| `application_id` (the store marker, R-6a5) | Migration 1 alone. | No other migration or code path writes it; the invariant test asserts its value over every fixture. |
| `users.admin`, `users.id`, `users.streamer_username` | **None at 6a.** The table lands empty; 6d's import and write path are its first writers. | No Go accessor for `users` exists (spec `:167-169`). |
| `collections.root_path`, `collections.writer`, `collections.owner` | **None at 6a.** Rung 7 or rung 10 creates the first row. | No Go accessor exists. The path CHECK constraints refuse a malformed root at the schema (R-6a9). |
| A job's state, owner, actor, and `cancel_requested` | **None at 6a.** 6b-i's runner is the first writer. | No Go accessor exists. |
| The store leaf's owner, and its directories' modes | The process that creates the leaf, which is the serve mode (Task 6), and only when its effective uid owns the state directory. The leaf, `db/`, `snapshots/`, and `staging/` are each 0700. | Every mode refuses a leaf whose owner differs from the effective uid (Task 1). A create run as a uid that does not own the state directory refuses, naming that owner (Task 1). |
| `staging/` contents | The snapshot writes a file there (Task 4). Serve mode alone removes leftovers, holding the exclusive lock with no intent record (R-6a7). | The write mode never removes a staging file. |
| Deleting a file in `snapshots/` | The runner's retention alone at 6a, by version, over `pre-migrate-v<N>.db` names only (R-6a11). | Retention ignores every other name; `keep` below 1 is refused. |
| `DUBPLATE_STORE_BUSY_TIMEOUT` | The environment, through the registry (Task 7). | The store reads no environment; the caller passes the value. |

---

## Resolutions this plan takes rather than discovers

Each resolution names its site. A task landing the surface a resolution
decides writes nothing about the resolution in a shipped comment; the
close maps each resolution to its landing commit from the task reports.

### R-6a1: lane 2, and the package territory against 5c-i

5c-i edits `internal/auth`, `internal/identity`, and
`cmd/dubplate-server` (`auth.go`, `doctor.go`, `serve.go`, `main.go`,
`member.go`), and leaves `internal/config` and `go.mod` untouched
(5c-i's R-5c17 and Global Constraints). 6a's new package shares nothing
with it. The overlap is `cmd/dubplate-server/doctor.go`, `doctor_test.go`,
and `main.go`, all in Task 7, which is why Task 7 is last before the
close: the fewer 6a commits touch those files, the smaller the rebase.
5c-i does not touch `check.go` (its plan's file lists name none). The
merge ritual's step 5 `diff-reviewer` runs on whichever lane rebases
second, because the rebase will not be a fast-forward.

### R-6a2: the exported surface is the inspect mode and its report

The spec (`:213-216`): "6a exports the open, whose production callers are
`doctor` and `check`, and tests the write method and the serve mode
in-package." This plan reads "the open" as the inspect mode, the only
mode those two callers use. Exported at 6a: one inspect function, its
options (the state directory and the busy timeout), its report type with
a closed state set, and nothing else. **Unexported at 6a:** the handle
type, the write mode, the serve mode, the write method and its
transaction handle, the migration runner, the snapshot function, the id
function, and every sentinel no caller outside the package branches on.
6b-i exports the write method with `ops/local` and the write mode with
`backup`; 6d exports the serve mode with `serve`. `cmd/dubplate-server`
branches on the report's state, never on an error's identity, so the
store exports no sentinel at 6a.

### R-6a3: the busy timeout registers at 6a; the migration wait at 6d

The fold record's contested row for `main.go` says "6a registers the busy
timeout and the migration wait." The busy timeout has a 6a production
reader: the inspect mode's DSN, which `doctor` and `check` pass. The
migration wait's only reader is the serve mode, whose production caller
is 6d's `serve`. Task 7 registers `DUBPLATE_STORE_BUSY_TIMEOUT` in
`cmd/dubplate-server/store.go` (the rung's own file, per CLAUDE.md), and
the serve mode takes the migration wait as a parameter; 6d registers
`DUBPLATE_STORE_MIGRATION_WAIT` with `serve`. Registering it at 6a would
put an env key on `doctor`'s table that no production code reads, which
misleads an adopter. This is an owed erratum ("Spec deviations that take
effect on approval").

**The busy timeout's default is 15 seconds**, a `KindDuration` threshold.
The researched basis: Navidrome's default DSN carries
`_busy_timeout=15000&_journal_mode=WAL&_foreign_keys=on&synchronous=normal`
(`consts/consts.go`, `DefaultDbPath`, read on master 2026-09-23). It is
the estate's own streamer on the same SQLite, and a busy timeout bounds
how long one process waits for another's write, which at group scale is
milliseconds. The store adopts Navidrome's value and declines its
`synchronous=normal` (the spec rules `FULL`).

**Its range is refused by the store row, not at load.** `internal/config`
resolves a threshold "at its point of use rather than at Load"
(`config.go`, the package comment), and `doctor`'s `thresholdStatus`
prints `d.String()` for any duration that parses, so the threshold row
shows `0s` or `-1s` as given. The registry has no bounds metadata until
6c, and the one existing bounds channel (an identity source's `Validate`
hook) lives in `auth.go`, a 5c-i file. The caller in
`cmd/dubplate-server/store.go` checks the range before any open, and the
store row carries the refusal, naming the key and the bound. The floor is
1 ms, because `_busy_timeout` takes integer milliseconds and a positive
duration under 1 ms would truncate to 0, no busy wait at all. The ceiling
is 2147483647 ms: the driver parses the value with `strconv.ParseInt(...,
10, 64)` (`sqlite.go`, v1.59.0), and a risk probe measured `2147483647`
reading back unchanged while `3600000000` and `99999999999999` read back
as `0`. Both bounds are named unexported constants.

### R-6a4: no `os` descriptor on a live database file; OFD locking declined

`modernc.org/sqlite`'s `OFDLocking` documentation (`ofd.go`, v1.59.0): "A
POSIX record lock belongs to the (process, inode) pair: the kernel drops
every POSIX lock the process holds on a file at any close(2) of any
descriptor of that file. Code as innocent as `f, _ := os.Open(dbPath) ...
f.Close()` anywhere in the process ... therefore silently strips SQLite's
own transaction locks and leaves the database unprotected against other
processes." OFD locks are opt-in, and the maintainer's changelog says
"Off by default, and staying off until the mode has real-world mileage";
they also leave WAL's `-shm` coordination on POSIX locks. The store
declines them and takes the rule instead: **no code in the store opens
the database, `-wal`, or `-shm` file through `os` while any connection to
it may be open in the process.** Every "is this a store" and "did the
bytes change" check uses `os.Stat` or a SQLite connection. The two
permitted `os` opens are the `O_CREATE|O_EXCL` pre-create of a database
file that does not yet exist (no connection can hold it) and the
pre-create of a snapshot target, a different inode. Tests that hash the
database file's bytes do so only after every handle in the test process
is closed, and say so in their fixture state.

### R-6a5: a store carries a marker, and every file state has one report

Migration 1 sets `PRAGMA application_id` to a fixed dubplate value
(chosen by Task 5 and asserted by the invariant test), inside its own
transaction. The probe measured that `application_id` and `user_version`
written inside a transaction roll back with it, and commit with it.

**How a file is classified.** Restore-in-progress and busy are decided
before classification runs, at Task 6's "Every mode's order" steps 3 and
6. "Absent" is decided only by `os.Stat` returning an error for which
`errors.Is(err, fs.ErrNotExist)` holds. No
SQLite code maps to absent: a risk probe measured a `mode=ro` open of an
existing file with mode 0000 returning `unable to open database file
(14)`, the same code an absent file returns. Every mode reads the marker,
the version, the schema count, and the ledger through the inspect DSN
(`mode=ro`), and the write and serve DSNs open only after the file
classifies as a dubplate store or an interrupted create. A writing DSN
runs `_journal_mode=WAL` on open, which would rewrite a foreign
rollback-journal file's header before the refusal.

**The states a leaf and its database file can meet, and the report every
open gives:**

| Leaf or file | Report |
| --- | --- |
| no leaf, or a leaf (with or without `store.lock`, `db/`, `snapshots/`, or `staging/`) holding no database file and no copy in `snapshots/` | absent (inspect, which creates nothing, the lock file included); refused naming "start serve first" (write); created, or the leaf completed and then created (serve) |
| a leaf whose database file is absent, zero bytes with no `-wal`, or header-only and unmarked, beside a `pre-migrate-v<N>.db` or `pre-restore-*.db` in `snapshots/` | Ruling 3: under yes, refused in every mode, naming the newest copy and `restore --from-file`, with the database file unchanged; under no, as the row for the file's own state |
| a leaf missing `store.lock` while a database file exists | not a store, naming the missing lock file and "start serve to finish creating it" (inspect and write, which read nothing without the lock and never create the lock file); completed under the exclusive lock (serve) |
| zero bytes, no `-wal`, and no copy in `snapshots/` (Ruling 3) | not a store, naming "start serve to finish creating it" (inspect and write); an interrupted create that serve completes (R-6a6) |
| zero bytes with a `-wal` beside it | not a store, naming both files; no mode truncates or removes either |
| any size, `application_id` 0, `user_version` 0, no schema object, and no copy in `snapshots/` (Ruling 3) | an interrupted create: not a store, naming "start serve to finish creating it" (inspect and write); completed by serve (R-6a6) |
| SQLite answers `SQLITE_NOTADB` (code 26) | not a store, naming the path |
| a SQLite database whose `application_id` is neither 0 nor dubplate's | not a store, naming the path and the foreign id |
| `application_id` 0 with any schema object or a nonzero `user_version` | not a store (a foreign database never marked) |
| present, but `os.Stat` fails with anything but not-exist, or the open or first read fails with any code but 26 (14 on a present file, 8, 10, 11) | refused, naming the path and the error |
| marked, and the ledger table is missing | refused, naming the path |
| marked, `user_version` below the binary's highest | older, naming each pending migration file |
| marked, `user_version` above the binary's highest | newer, naming both versions |
| marked, current, ledger agrees | current |
| marked, a ledger row whose name or hash differs from the binary's file | refused, naming both |
| marked, a version at or below `user_version` with no ledger row, or a ledger row above `user_version` | refused, naming the version and `user_version` |

**"Newer" is decided on `user_version` alone, before the ledger is
read.** A marked store whose `user_version` is above the binary's highest
reports newer, naming both versions, whatever its ledger holds, because
the binary cannot judge versions above its head: a newer binary leaves
ledger rows for versions this binary has no file for, and those rows are
expected, not a disagreement. The ledger rows of this table run only when
`user_version` is at or below the binary's highest.

The inspect report's closed state set holds every row: absent, restore in
progress, busy, not a store, refused, older, newer, current. A ledger
disagreement and an unreadable file are "refused" with their reason.

The revision 1 probe measured the zero-byte and code-26 rows: a
zero-byte file opened `mode=ro` reads `user_version` 0 and an empty
schema with no error, and a file of repeated text returns `file is not a
database (26)` on the first read. The mechanics review measured a 5-byte
file, 99 zero bytes, and a bare "SQLite format 3" header each returning
code 26. The header-only row is R-6a6's.

### R-6a6: an interrupted create is completed only by serve mode

The serve mode creates a store by making the leaf, pre-creating the
database file at 0600 with `O_CREATE|O_EXCL`, opening it, and applying
migration 1. The open itself writes page 1. Three probes (the mechanics
review's D and D-2, and the risk review's child-process probe, distinct
from the risk lens's own finding D2) measured a
pre-created empty file opened with the write DSN reading 4096 bytes with
no `-wal`, `application_id` 0, `user_version` 0, and no schema object,
both after a rolled-back transaction and after a process exit without a
close. So a crash anywhere from the pre-create to migration 1's commit
leaves one of two files: zero bytes (before the open) or header-only
(after it). Under the exclusive lock, with no `-wal` beside a zero-byte
file, the serve mode treats either as its own interrupted create and
applies every migration from version 0, with no snapshot, since the store
is empty. The revision 1 probe measured that SQLite creates a WAL
database inside a pre-created empty 0600 file and gives `-wal` and `-shm`
the same 0600. The inspect and write modes report either as "not a
store" with the fix. Under Ruling 3's yes, serve first reads `snapshots/`,
and a copy there makes either file "refused" in every mode rather than a
create to complete. A foreign SQLite file that is also empty and
unmarked is indistinguishable from a header-only create, and serve adopts
it; nothing is lost, because it holds nothing.

The spec's criterion 13 ("a zero-byte file ... refused as not a store")
holds for the inspect and write modes. This is an owed erratum.

### R-6a7: a restore-intent record at 6a is presence only, and staging cleanup runs only in serve mode

The intent record's format and its `reconcile-pending` state are 6b-ii's.
At 6a: every mode checks for `restore-intent` in the leaf at Task 6's
"Every mode's order" step 3, after the stat and the owner check and
before any other step. Inspect reports "restore in progress." Write mode
refuses naming `restore --resume`. **Serve mode also refuses**, naming
`restore --resume`, until 6b-ii teaches it the one state it accepts.
6b-ii refines both rules with the format.

**Staging cleanup runs in serve mode alone, holding the exclusive lock,**
in step 3 after the intent check and before the create or the snapshot,
and only when no intent record exists, so at 6a it can never remove a
file an intent names. The write mode never cleans. Two reasons, one each
way. Under the spec, `backup` (a write-mode verb under the shared lock)
stages its `VACUUM INTO` copy in `staging/`, and every `member` verb also
opens in write mode under a shared lock, so a write-mode cleanup would
remove a live backup's staging file. And the snapshot pre-creates its
staging file with `O_EXCL`, so a crash mid-snapshot leaves a file that
would refuse every later migration if serve never cleaned. The spec's
"Every writing open ... remove[s] staging leftovers" (`:235-236`) narrows
to serve mode; this is an owed erratum. 6b-i's shared-lock stager names
its file by instance id, and its recovery removes only a file whose
owner's lease is free (Task 8's handoff).

### R-6a8: times are integer UTC microseconds

Every time column is `INTEGER NOT NULL` (or nullable where the spec makes
the time optional, such as a tombstone) holding `time.Time.UnixMicro()`
in UTC. Code binds an `int64`, never a `time.Time`, so no driver time
format (`_time_format`, `_inttotime`) enters the schema's contract.
Invariant 3's `at` column follows the same rule.

### R-6a9: the path rules are CHECK constraints; the Go validator waits for rung 7

The spec (`:480-487`): "The store refuses a path that is absolute, has an
empty, `.`, or `..` component, or carries NUL." The refusal lands in the
schema: `collections.root_path` and `tracks.rel_path` each carry a CHECK
over the column's bytes, because at 6a no Go code writes a path, and a Go
validator with no caller would be exported ahead of its caller. The probe
measured one form that works over a `STRICT` `BLOB` column with
`LIKE_DOESNT_MATCH_BLOBS` compiled in: a non-empty length, a first byte
that is not `/`, no NUL byte (`instr` with `x'00'`), and no `//`, `/./`,
or `/../` byte sequence in the value framed by a `/` byte at each end
(`instr` over the framed value cast back to `BLOB`). It accepted
`Artist/Album/01 Track.flac`, `A/..B/.c/d..flac`, and a non-UTF-8 name,
and refused an absolute path, an empty value, a NUL anywhere, `a/./b`,
`a/../b`, `../b`, `a/.`, `a//b`, `a/b/`, and `.`. Rung 7's gate adds the
Go-side validator with its first writer.

### R-6a10: the write method records every statement error itself

The probe measured two facts that make a plain `*sql.Tx` unsafe to hand
to a callback:

- A callback that swallows a failed statement's error still commits the
  rest. Inside one `IMMEDIATE` transaction, an insert that succeeded,
  then an orphan insert that failed with `FOREIGN KEY constraint failed
  (787)`, then `Commit` returned nil, and the first row was committed.
- SQLite may end a transaction on its own after a statement fails, and
  the spec names four codes that can do this: `SQLITE_FULL`,
  `SQLITE_IOERR`, `SQLITE_INTERRUPT`, and `SQLITE_NOMEM`
  (`lang_transaction.html` section 3, spec `:327-329`). With `PRAGMA
  max_page_count = 3`, an oversized insert failed with `database or disk
  is full (13)`, and the following `Commit` failed with `cannot commit -
  no transaction is active (1)`, with no row committed.

A third fact, measured by the risk review and re-run by the fold probe
(`6a-fold-probes/`, P1), makes recording alone insufficient: **after
SQLite rolls a transaction back, the connection is in autocommit mode,
and every later statement commits by itself.** With `max_page_count`
pinned at the current `page_count`, a callback ran `insert into small
values(2)` (nil), an oversized insert (`database or disk is full (13)`),
then `insert into small values(3)` (nil); `Rollback` answered `cannot
rollback - no transaction is active (1)`, and `small` then held rows 1
and 3. Row 3 committed outside any transaction, one of the four codes
named above.

So the callback receives a store-owned transaction handle, never a
`*sql.Tx`. Its statement methods (an exec, a single-row query whose scan
error is recorded, and a multi-row query whose iteration and close errors
are recorded) remember the first error, **and once an error is recorded
every later statement method returns that error without reaching
SQLite**, so nothing the callback issues afterward can run in autocommit
mode. The write method commits only when the callback returned nil
**and** the handle recorded no error. When SQLite already rolled back,
the rollback's "no transaction is active" answer is not a second error;
the write method returns the statement's error, and for code 13 it
returns the named disk-full error.

The callback's signature carries the marked context:
`func(ctx context.Context, tx <handle>) error`. The nested-write mark
(Task 3) catches a nested call only when it uses that `ctx`; a callback
that closes over its caller's outer `ctx` waits for the one connection
until its deadline instead. The package comment states that limit.

### R-6a11: one pre-migration snapshot per migrating run

A run that finds the store at version N and the binary at M > N writes
one `pre-migrate-vN.db`, before the first pending file applies. If a
`pre-migrate-vN.db` already exists (an earlier run at N that failed),
the new copy replaces it by rename, because the newest state at version N
is the rollback point. A copy of any other version is never touched by
the write. With backup off, the retention step then keeps the newest
`keep` copies by version number, where `keep` is the runner's parameter
(the caller passes the stored row or the registry default from 6c; at 6a
the only callers are tests). Retention selects only names that parse as
`pre-migrate-v<N>.db`; it never deletes, and never errors on, any other
file in `snapshots/`, such as 6b-ii's `pre-restore-<stamp>.db` or an
operator's file. The runner refuses a `keep` below 1, which would delete
the copy it just took. A store at version 0 is empty and takes no
snapshot.

**Free space** is checked before the copy: `f_bavail × f_frsize` from
`statfs(2)` must cover `page_count × page_size` plus the post-checkpoint
WAL size (R-6a14). The check covers the snapshot only. A migration that
rewrites a large table grows the WAL inside its one transaction past that
size, so a mid-migration disk-full is possible after the check passes; it
returns the named disk-full error with the store at the prior version. `statvfs(3)` states `f_blocks` is "Size of fs in f_frsize
units"; the probe measured `f_bsize` and `f_frsize` equal on both local
filesystems and matching `df`, and the store uses `f_frsize`.

### R-6a12: `doctor`'s store row is one line, always (Ruling 2)

The row prints for every state, absent included, so the line-count
arithmetic in `TestDoctorCmd_LineCountFollowsRegistry` gains one
constant line that is not registry-derived. That test names the constant
and asserts it, so a second store row fails it. The row names no path
beyond the state directory's leaf and no address. The exit rule is Ruling
2's: absent, older, and current exit 0; every other state exits nonzero.
`check` prints its store line after `configuration: ok` and applies the
same rule. A FUSE warning rides the row's text and does not change the
exit.

### R-6a13: the leaf's `leases/` directory lands with 6b-i

The serve mode's create makes the leaf (0700), `store.lock`, `db/`,
`snapshots/`, and `staging/`. `leases/` is 6b-i's first consumer and
lands there.

### R-6a14: three named constants, with their values

- **The post-checkpoint WAL size**: `journal_size_limit` of 64 MiB,
  applied through `_pragma` on every writing class (the inspect class
  writes nothing and omits it). SQLite's `wal.html` section 6: "The
  checkpoint does not normally truncate the WAL file (unless the
  journal_size_limit pragma is set)." It bounds the WAL after a
  checkpoint, never during a transaction. Adopters never set it; the
  free-space rule adds it to the space a snapshot needs.
- **The lock poll interval**: 50 ms, the interval every bounded lock wait
  polls `LOCK_NB` at (the salvage lock package's shape, adopted in
  reasoning, as 5c-i did).
- **The holders log interval**: 5 s. A bounded lock wait that logs its
  holders (serve mode's shared and exclusive waits, Task 6) emits the
  first holders record on its first failed poll, then one record every
  5 s until the wait ends. The first record needs no interval to pass, so
  a test ordered on it waits one poll, not 5 s.

### R-6a15: every migration ships with a populated fixture

A fixture is a SQL file of inserts under `internal/engine/store/testdata/`
that populates the schema as of its migration's version, one per
migration file, named after it. The invariant test applies every
migration to an empty database and, for each migration, migrates to that
version, applies its fixture, then migrates to the head; it fails on a
migration with no fixture. A fixture never carries a real address or
name.

### R-6a16: `doctor`'s goldens render from constructed reports

`cmd/dubplate-server` can construct an inspect report (exported data) but
cannot build an older or newer store without the unexported runner. So
the golden test renders one exact line per state from a constructed
report, and two end-to-end cases run the real inspect mode against a
state directory with no leaf (absent) and one whose `db/` holds a
zero-byte database (not a store). The in-package tests prove that the
inspect mode produces every state from a real fixture (Task 6).

### R-6a17: one DSN builder escapes the path, and no DSN creates

SQLite parses a `file:` URI, so a path carrying `#`, `?`, or `%` opens
the wrong file when it is concatenated in. The mechanics review measured
`"file:" + dir + "/s.db"` with a directory named `hash#01` or `q?mark`
creating a database at the truncated path `hash` or `q` with no error.
`t.TempDir()` keeps `#` and `%` in its names (`testing.go`, `allowed`),
and Go names a duplicate subtest `name#01`, so the gate itself can meet
this. One unexported builder forms every DSN as `(&url.URL{Scheme: "file",
Path: p}).String()` followed by `?` and the class's keys. The fold probe
(P2) measured it: under a directory named `a#b%41?c` it produced
`file:///.../a%23b%2541%3Fc/s.db?mode=rw&_journal_mode=WAL&_busy_timeout=500`,
the exec succeeded, `busy_timeout` read back 500, `journal_mode` read
`wal`, the database was at the exact path (`a#b%41?c/s.db`), and no
entry appeared at a truncated path (`a`) beside the directory.

Every writing class (write pool, read pool, maintenance) carries
`mode=rw`, and the inspect class carries `mode=ro`. No DSN can create a
database. The serve mode's create, which pre-creates the file first,
needs no DSN that can create, and a missing `os.Stat` check can no
longer turn an open into a create. The fold probe (P3) measured `mode=rw`
on an absent file returning `unable to open database file (14)` and
creating nothing, and (P2) `mode=rw` on a pre-created empty file opening
it at that exact path in WAL mode (`journal_mode` read `wal`). The
snapshot's verification open uses the inspect DSN.

### R-6a18: a caller's deadline does not shorten SQLite's busy wait (Ruling 4)

The mechanics review measured a `BeginTx` with a 300 ms context and
`_busy_timeout=3000` returning `database is locked (5) (SQLITE_BUSY)`
after 3.01 s. The driver does interrupt a statement when its context is
done (`sqlite.go`, `interruptOnDone`, v1.59.0), and its test
`TestInMemoryDBSurvivesContextCancel` (`all_test.go`) states: "File-backed
databases keep the existing behaviour: an interrupted connection is still
discarded, since the underlying data lives on disk." A write pool's one
connection is replaced after an interrupt, and per-connection state such
as a test's `max_page_count` does not survive it. The interrupt does not
cut the busy handler's wait. Under Ruling 4's yes, the write method's
contract reads: the context bounds the in-process wait for the one write
connection, and the busy timeout alone bounds the wait on another
process's write lock. A test pins the measured behavior. Under no, Task 3
gains the per-call busy timeout and splits (Ruling 4).

---

## Review focus

These are the input classes most likely to bite, the ones no happy path
exercises, each pinned by a test in the task that owns it.

1. **A store opened by the wrong user.** `doctor` run as root on the box.
   Expected: refused, naming the owner and `runuser -u <owner> --`, with
   no file created or chowned. Pinned in Task 1.
2. **A migration that commits cleanly but is wrong.** Expected: the
   pre-migration copy exists, is verified, and survives the next upgrade.
   Pinned in Task 4.
3. **A callback that ignores an error.** Expected: nothing commits.
   Pinned in Task 3.
4. **A lock conversion that a restore wins.** Expected: the serve mode
   holds no descriptor to the renamed file and re-reads the version.
   Pinned in Task 6.
5. **A foreign SQLite file at the store's path.** Expected: "not a
   store," never a migration into it, and its bytes untouched by the
   refusal. Pinned in Task 6.
6. **A callback that keeps going after a disk-full.** Expected: nothing
   it issues afterward commits in autocommit mode. Pinned in Task 3.
7. **A crash during the first create or mid-snapshot.** Expected: the
   next serve start completes the create or clears the staging leftover;
   it never refuses forever. Pinned in Task 6.

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched by the conductor before any
implementer. Reads the spec, the fold record's "Revision 2.2" section,
and this plan against the tree as shipped. **Its output is blockers with
evidence, not fixes.** It writes no code and edits no file. The conductor
rules each blocker before Task 1 dispatches.

**Every check runs against the real HEAD, and the task reports that
HEAD.** This plan's code pointers were resolved at `cb9ecc3`. Grade a
moved pointer by content: a pointer whose content survives at any offset
is a note naming the new line, and a pointer whose content is gone is a
blocker.

**The first five items are blockers: a "no" stops the launch.**

1. **Is the worktree set up and free, and is lane 1 where the registry
   says?** Report `git worktree list`, the lane registry rows, and `git
   status` in both trees. Report any uncommitted change under `internal/`,
   `cmd/`, or `go.mod` in the worktree. *(Expected: lane 1 on 5c-i in the
   main tree, lane 2's worktree fresh from master.)*
2. **Is `internal/engine/store` absent, and does no file in the module
   import `modernc.org/sqlite`?** A lane that landed either first changes
   every task.
3. **Has 5c-i touched `internal/config`, `go.mod`, or `check.go`?**
   Report the commit list for each since `08befea`. Any commit is a
   blocker for R-6a1's territory.
4. **Is `modernc.org/sqlite v1.59.0` still the newest version, and does
   it build with the repo's toolchain?** Report `go list -m -versions modernc.org/sqlite`, and whether a scratch module outside the worktree that imports it compiles. A newer version is a blocker only if its changelog touches
   the DSN parser, the `Error` type, or locking; otherwise it is a note
   and the conductor rules which to take under the `dependency-upgrade`
   skill.
5. **Does the gate run cleanly from the worktree?** Run `bash
   scripts/check.sh` through `cairn-run-gate` in the worktree at its
   starting HEAD and report the wall time and the exit. The HISTORY note
   "gitleaks scans the worktree from the main tree" (the 2026-09-08
   two-lane run) says the gitleaks leg may read the wrong tree; report
   which tree it scanned.

**Notes (not blockers):**

6. Report `newProductionRegistry`'s body and `newRootCmd`'s `AddCommand`
   block at HEAD, the lines Task 7 extends, and whether 5c-i's lane has
   already added its `member` line.
7. Report `TestDoctorCmd` and `TestDoctorCmd_LineCountFollowsRegistry` by
   name and the arithmetic each asserts, the property R-6a12 changes.
8. Report `check_test.go`'s assertions on `check`'s stdout, which Task 7
   extends.
9. Report `DUBPLATE_STATE_DIR`'s handling in `config.Load` and `doctor`
   (`lookupEnviron`), since Task 7 reads it on both paths.
10. Report the fold record's contested-surface rows by heading, and
    confirm the `main.go` row still says 6a registers the migration wait
    (R-6a3 declines that).
11. Report `internal/atomicfile`'s exported functions, and confirm that
    neither `WriteFile` nor `WriteStream` renames an existing file into
    place (the spec's reason the snapshot fsyncs its directory by hand).
12. Report the spec pointers this plan cites (`:141-189`, `:193-216`,
    `:220-494`, `:842-856`, `:860-880`, `:884-902`, `:911-949`), which
    document lane 1's errata may have moved.

**Deliverable count: 1** (the report).

---

## Task 1: The store leaf: layout, lock, owner check, and filesystem probe

Model: `sonnet`. Depends on Task 0. **Independent by files of Task 2.**

The leaf is pure filesystem work, with no driver, so it lands first and
every later task opens through it.

**Files:**
- Create: `internal/engine/store/doc.go` (the package comment),
  `internal/engine/store/leaf.go`, `internal/engine/store/lock.go`,
  `internal/engine/store/fsprobe.go`, and a `_test.go` for each

**Interfaces (all unexported):**
- Produces: a leaf value built from a state directory, naming
  `store.lock`, `db/`, the database path, `snapshots/`, `staging/`, and
  `restore-intent`; a lock value over `store.lock` with shared-nonblocking,
  shared-bounded, and exclusive-bounded acquisition and a release; an
  owner check; a filesystem probe held in a package variable (spec seam
  1); a pre-create function that makes a file at 0600 with
  `O_CREATE|O_EXCL` and fails if it exists.

**Outcomes:**

- **The leaf** is `<state dir>/store`, 0700, holding the names the spec
  lists (`:224-233`) except `leases/` (R-6a13). A function creates the
  leaf, `store.lock`, and the three directories (each 0700) for the serve
  mode's create; nothing else creates them. **The create is idempotent**,
  because a crash can stop it partway and a second process can race it:
  each directory tolerates an existing directory, and `store.lock` opens
  with `O_CREATE` and no `O_EXCL`. The lock lives inside the leaf, so the
  leaf create cannot run under the lock; the serve mode takes the
  exclusive lock after it and re-stats before the database pre-create.
  **A create requires the effective uid to own the state directory**, and
  otherwise refuses naming that owner, so a first run as root cannot
  leave a root-owned leaf whose refusal would then name `runuser -u root`.
  Inspect and write open `store.lock` without `O_CREATE`; a missing lock
  file is a state (R-6a5), never a file they make.
- **The lock** is an `flock` on `store.lock`, a file outside `db/` that
  nothing renames. `man 2 flock` (read 2026-09-23): "Locks created by
  flock() are associated with an open file description"; "If a process
  uses open(2) (or similar) to obtain more than one file descriptor for
  the same file, these file descriptors are treated independently by
  flock()"; "A shared or exclusive lock can be placed on a file regardless
  of the mode in which the file was opened"; and "A call to flock() may
  block if an incompatible lock is held by another process. To make a
  nonblocking request, include LOCK_NB". Every bounded wait polls
  `LOCK_NB` at the R-6a14 interval until `ctx` ends, and returns a busy
  error naming the holders when it does. The shared-nonblocking form
  (inspect) makes one attempt.
- **A refusal names the holders.** `flock` names none, so the lock reads
  `/proc/locks` when it is readable and lists the pid of every line whose
  device and inode match `store.lock`. `proc_locks(5)` (read 2026-09-23):
  field 5 is "The PID of the process that owns the lock"; field 6 is
  "Three colon-separated subfields that identify the major and minor
  device ID of the device containing the filesystem where the locked file
  resides, followed by the inode number of the locked file"; and "the
  list of locks shown in /proc/locks is filtered to show just the locks
  for the processes in the PID namespace ... for which the /proc
  filesystem was mounted." The probe measured a shared `flock` on a file
  with device 60 printing `FLOCK  ADVISORY  READ 54300 00:3c:12157465 0
  EOF`: the major and minor are hexadecimal and the inode decimal. The
  split of `Stat_t.Dev` into major and minor follows glibc's encoding,
  quoted from `/usr/include/bits/sysmacros.h` (read 2026-09-23): major is
  `(dev & 0x00000000000fff00) >> 8` or'd with `(dev & 0xfffff00000000000)
  >> 32`, and minor is `dev & 0xff` or'd with `(dev & 0x00000ffffff00000)
  >> 12`. A line carrying `->` is a queued waiter, not a holder, and its
  fields shift by one (the mechanics review measured `96: -> FLOCK
  ADVISORY WRITE 63012 00:3c:...`), so the parse skips it. An unreadable
  `/proc/locks`, or no matching line, yields "held by another process"
  with no pid list. **On btrfs the list is always empty:** the mechanics
  review measured `stat` reporting a per-subvolume device (`0:3a`) while
  `/proc/locks` printed the superblock's (`00:25`). The fallback text is
  honest there, and the gap is declared (Task 8).
- **The owner check.** Every mode that finds an existing leaf compares
  the leaf's owner uid with `os.Geteuid()`, at Task 6's "Every mode's
  order" step 2, before the restore-intent check. A
  mismatch refuses, naming the owner (from `os/user.LookupId`, which
  returns "an error ... of type UnknownUserIdError" when "a user cannot be
  found", in which case the name is `uid:<n>`) and the fix `runuser -u
  <owner> -- dubplate-server ...`. An absent leaf reports absent with no
  uid check. The comparison is an unexported function of the leaf's
  stat and an effective uid, so a test can refuse without root; that is
  ordinary dependency injection, not a fourth seam.
- **The filesystem probe** calls `syscall.Statfs` on `db/` when it
  exists, else the leaf, else the leaf's parent, since the database's own
  directory is what WAL depends on and `statfs` follows a symlinked `db/`
  to its target. It classifies `f_type` against
  `/usr/include/linux/magic.h` (read 2026-09-23): `NFS_SUPER_MAGIC
  0x6969`, `CIFS_SUPER_MAGIC 0xFF534D42`, `SMB2_SUPER_MAGIC 0xFE534D42`,
  and `V9FS_MAGIC 0x01021997` refuse, because WAL needs every process on
  one host (the spec, citing `wal.html` section 1);
  `FUSE_SUPER_MAGIC 0x65735546` warns; every other type opens. Three more
  multi-host filesystems refuse on the same ground, beyond the spec's
  list, over four constants: `SMB_SUPER_MAGIC 0x517B` (the retired
  smbfs), `CEPH_SUPER_MAGIC 0x00c36400`, and AFS's two, `AFS_SUPER_MAGIC
  0x5346414F` and `AFS_FS_MAGIC 0x6B414653` (`magic.h` line 56, the
  second AFS constant it defines; line 7 carries the first). The probe
  measured
  `tmpfs` (`0x1021994`) and `btrfs` (`0x9123683e`) on this workstation,
  so tests on `t.TempDir()` must open.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Leaf | absent | absent, no owner check |
| Leaf | present, owner equals euid | proceed |
| Leaf | present, owner differs, owner resolvable | refused, naming the owner's name and the `runuser` fix |
| Leaf | present, owner differs, owner unknown | refused, naming `uid:<n>` |
| Leaf | present but a regular file, or mode not 0700 | refused, naming the path and the expected shape |
| Leaf create | a directory or `store.lock` already exists (a partial create, or a racing process) | tolerated; the create completes |
| Leaf create | the state directory is owned by another uid | refused, naming that owner; nothing created |
| Lock | free | acquired |
| Lock | shared held, shared requested | acquired |
| Lock | exclusive held, shared nonblocking requested | busy, naming holders |
| Lock | incompatible held past `ctx` | busy, naming holders; nothing changed |
| Lock | `/proc/locks` unreadable, or no device match (btrfs) | busy, "held by another process", no pids |
| Lock | a `->` waiter line on the lock file | skipped; not named as a holder |
| Filesystem | NFS, CIFS, SMB, SMB2, 9p, Ceph, AFS | refused, naming the type |
| Filesystem | FUSE | opens, with a warning naming FUSE |
| Filesystem | anything else | opens |
| Pre-create | absent target | created at 0600 whatever the umask |
| Pre-create | target exists | refused, the existing file untouched |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criterion 16's first two parts: the seam set to NFS refuses, to
  FUSE warns, and to ext4 (`0xEF53`) opens. **Fixture state:** the seam
  replaced for the test and restored by `t.Cleanup`, and a real leaf on
  `t.TempDir()` so the probe is the only difference. The owner-mismatch
  row refuses naming the owner, with the injected euid one above the
  test's own.
- A pre-create under a process umask of 0022 lands 0600. **Fixture
  state:** the test sets `syscall.Umask(0o022)` and restores the prior
  value, since the test process's umask is otherwise unknown.
- A held-lock test takes an exclusive `flock` on a second descriptor and
  asserts the shared-nonblocking attempt returns busy **naming the test
  process's own pid** (the one fixture that proves the `/proc/locks`
  match is real), and a bounded shared attempt returns busy after a short
  `ctx`. **Fixture state:** the lock file lives in a directory made by
  `os.MkdirTemp("/dev/shm", ...)` and removed by `t.Cleanup`, because
  `t.TempDir()` may sit on btrfs, where no line matches. The fold probe
  (P9) measured `/dev/shm` as tmpfs (`0x1021994`) and the computed
  `00:1b:<inode>` found on the process's own `FLOCK ADVISORY WRITE` line.
  The same test plants a synthetic `->` line through a parse function
  over the file's text and asserts it names no holder.
- The leaf create run twice, and run after a partial create that left
  `store.lock` and `db/` only, completes both times with every directory
  0700; a create against a state directory whose owner differs (the
  injected owner comparison, as the owner check) refuses and creates
  nothing. **No test in the package calls `t.Parallel()`** (the umask
  rows).
- A released lock lets the waiter in: a bounded exclusive waiter blocked
  behind a shared holder acquires after the holder releases, ordered by a
  channel, never a sleep.
- Doc comments carry no em dash; `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Compare the leaf's owner with the real uid instead of the effective
   uid → no test fires under a normal test run; report it as a
   reviewer-read row.
2. Skip the owner check when the leaf exists → the mismatch row passes.
3. Match `/proc/locks` on inode alone → no test fires on one device;
   report it as a reviewer-read row and name the collision it opens.
4. Parse the device fields as decimal → the own-pid row names no pid.
5. Treat FUSE as refused → the FUSE row fails.
6. Create with `os.Create` → the umask row lands 0644.
7. Parse `->` lines as holders → the synthetic waiter row names a pid.
8. Create the leaf's directories with `os.Mkdir` failing on an existing
   one → the second-create row fails.

**2.0 hook: neutral.** **Prior art:** adopts the salvage lock package's
poll-under-`ctx` shape (`_salvage/internal/lock/lock.go`) in reasoning
with stdlib `syscall.Flock`, as 5c-i did; declines `golang.org/x/sys/unix`
for its magic constants and `Flock`, since the stdlib carries both needs
and a direct import would make an indirect requirement direct.

**Deliverable count: 4** (the leaf and its create, the lock with its
holder report, the owner check, the filesystem probe with the
pre-create).

---

## Task 2: The driver and the three connection classes

Model: `sonnet`. Depends on Task 1 for the leaf (a handle opens the
leaf's database path). **Lands 2.0 hook 9's WAL, `FULL`, and busy
timeout.**

**Files:**
- Modify: `go.mod`, `go.sum` (one requirement, `modernc.org/sqlite
  v1.59.0`)
- Create: `internal/engine/store/conn.go`, `internal/engine/store/conn_test.go`,
  `internal/engine/store/driver_test.go` (the adoption conditions)

**Interfaces (all unexported):** a handle holding the write pool (one
connection), the read pool, and a function that opens a maintenance
connection for one job and returns it to be closed; a DSN builder per
class; a typed busy error and a typed read-only error, each derived from
the driver's `*sqlite.Error` code.

**Outcomes:**

- **Every per-connection setting rides the DSN**, because `database/sql`
  opens connections at will. The DSN is a `file:` URI. The driver's
  `newConn` opens with `SQLITE_OPEN_READWRITE|SQLITE_OPEN_CREATE|
  SQLITE_OPEN_FULLMUTEX|SQLITE_OPEN_URI` (`conn.go`, v1.59.0), so URI
  parameters such as `mode=ro` reach SQLite. The shorthand keys and their
  validation, quoted from `driver.go`: "Each value is validated against
  the same set github.com/mattn/go-sqlite3 accepts (case-insensitive); an
  unrecognized value fails the connection with an error instead of being
  silently ignored. The keys are applied in a fixed order ...
  _busy_timeout and _auto_vacuum first ..., then the _pragma values, then
  the remaining keys, and _query_only last." And: "each _pragma value is
  executed verbatim (with PRAGMA prepended) and is not validated."

  Every DSN comes from R-6a17's one builder, which escapes the path.

  | Class | DSN keys |
  | --- | --- |
  | write pool (`SetMaxOpenConns(1)`) | `mode=rw`, `_busy_timeout`, `_foreign_keys=1`, `_journal_mode=WAL`, `_synchronous=FULL`, `_txlock=immediate`, `_pragma=journal_size_limit(<bound>)` |
  | read pool | the same without `_txlock`, plus `_query_only=1` |
  | maintenance (its own `*sql.DB`, one connection, closed after use) | `mode=rw`, `_busy_timeout`, `_foreign_keys=1`, `_synchronous=FULL`, `_txlock=immediate`, `_pragma=journal_size_limit(<bound>)` |
  | inspect (first used by Task 4's snapshot verification, then Task 6) | `mode=ro`, `_busy_timeout`, `_query_only=1` |

  The maintenance class carries `_txlock=immediate` because `newTx`
  issues a plain `begin` when the begin mode is empty (quoted below); the
  mechanics review measured a maintenance transaction that read, met
  another connection's commit, and then failed its write at once with
  `database is locked (517)` (`SQLITE_BUSY_SNAPSHOT`, no busy wait).
  `VACUUM INTO` runs outside `BeginTx`, so the key does not touch it.

- **`_txlock=immediate` governs only read-write transactions.** The
  driver's `newTx` (`tx.go`): `sql := "begin"; if !opts.ReadOnly &&
  c.beginMode != "" { sql = "begin " + c.beginMode }`. The probe measured
  a `ReadOnly` transaction on the write pool taking no write lock (a
  second handle's write proceeded at once). The write method therefore
  never passes `ReadOnly`, and no reader uses the write pool.
- **The typed busy error** wraps any `*sqlite.Error` whose `Code()` is 5
  (`SQLITE_BUSY`). The probe measured a second handle's insert, with
  `_busy_timeout=400`, returning `database is locked (5) (SQLITE_BUSY)`
  after 400 ms while the first handle held an `IMMEDIATE` transaction, and
  a `BEGIN` succeeding once it released. The driver enables extended
  result codes (`c.extendedResultCodes(true)` in `newConn`), so the
  comparison masks the primary code (`code & 0xff`) and names the
  extended code in the message; Navidrome's `db.go` makes the same
  distinction, noting that `SQLITE_BUSY_SNAPSHOT` "only a new transaction
  clears."
- **The read-only error** wraps code 8 (`SQLITE_READONLY`): the probe
  measured an insert through a `_query_only=1` connection returning
  `attempt to write a readonly database (8)`.
- **The adoption conditions** (the spec's prior-art verdict, S20's
  condition) are gate-resident tests: FTS5's trigram tokenizer answers a
  substring query, and a reader holding a transaction keeps its snapshot
  across a second handle's commit. The probe measured both on SQLite
  3.53.4 (`ENABLE_FTS5` and `ENABLE_SNAPSHOT` in `compile_options`): a
  `tokenize='trigram'` table matched `'emian'` to `Bohemian Rhapsody`,
  and a read-pool transaction counted 2 rows before and after a commit
  that a fresh read counted as 3. The test also asserts
  `LIKE_DOESNT_MATCH_BLOBS` in `compile_options`, which the spec's path
  rule relies on.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| DSN | a shorthand value outside the driver's set | the driver's own error at open; a test proves one |
| DSN | a path carrying `#`, `?`, or `%` | the file opens at that exact path (R-6a17) |
| DSN | a writing class against an absent file | `unable to open database file (14)`; nothing created |
| Readback | a setting reads back other than set | test failure naming class and pragma |
| Write on read pool | any write | the read-only error |
| Two handles | second writer within the busy timeout | waits, then commits |
| Two handles | second writer past the busy timeout | the busy error |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criteria 1 and 2, as described, each on a populated fixture (two
  or more rows) so a vacuous match fails.
- Spec criterion 4: every setting (journal mode, foreign keys,
  synchronous, busy timeout, `journal_size_limit`, `query_only`) reads
  back on a freshly opened connection of each class, the maintenance
  connection included. The inspect class reads back its busy timeout and
  `query_only` 1, the only keys it sets; a verification probe
  (`6a-fold2-probe/`) measured a `mode=ro&_busy_timeout=700&_query_only=1`
  open of a WAL file reading back `700`, `1`, and `wal`. **Fixture state:** each read pool connection is
  forced fresh (the test holds one connection with `DB.Conn` and opens a
  second, so the second is new), because a reused connection would pass
  on settings a single `PRAGMA` set once. The probe's readback on write
  and read classes returned `wal`, `1`, `2`, the timeout, `67108864`, and
  `0` or `1`.
- Spec criterion 5's first half: a write through the read pool returns
  the read-only error by `errors.Is`.
- Spec criterion 3 at the connection level: two independent handles on
  one file, the first holding an `IMMEDIATE` transaction for half the busy
  timeout then committing, the second waiting and committing; held past
  the timeout, the second returns the busy error. **Fixture state:** the
  busy timeout set short (a few hundred milliseconds) in both DSNs, and
  the hold ordered by a channel.
- Every class opens under a directory named `a#b%41?c` and its file
  lands at that exact path, with nothing created beside the directory
  (the fold probe's P2 shape). A writing class against an absent file
  returns code 14 and leaves the directory listing unchanged.
- The busy error classifies an extended `SQLITE_BUSY`: a second handle's
  deferred read-then-write after a first handle's commit returns 517
  (the mechanics review's probe E shape, on a test-opened handle whose
  DSN has no `_txlock`, so its `BEGIN` is deferred), and the error
  matches the typed busy error by `errors.Is`. A maintenance transaction
  under the same sequence takes the write lock at its `BEGIN`, so the
  other handle's write waits behind it and the maintenance write never
  returns 517, proving its `_txlock=immediate`.
- `go.mod` gains exactly one direct requirement; `go mod tidy` leaves it
  unchanged. The task reports the gate's wall time before and after, since
  the driver's first compile lands in every cold gate.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Drop `_foreign_keys=1` from the read pool's DSN → the readback row
   fails for that class.
2. Set `journal_size_limit` once with a statement instead of through
   `_pragma` → the fresh-connection readback fails.
3. Drop `_query_only=1` → the read-pool write succeeds.
4. Compare the busy error against the full extended code → the 517 row
   passes unclassified.
5. Drop `_txlock=immediate` from the write pool → no row fires here;
   Task 3's deferred-upgrade row fires (report it forward).
6. Concatenate the path into the DSN → the `a#b%41?c` row creates a file
   at a truncated path.
7. Drop `mode=rw` from a writing class → the absent-file row creates a
   database.
8. Drop `_txlock=immediate` from the maintenance class → the maintenance
   row returns 517.

**2.0 hook: lands hook 9's connection half** (one writer, WAL,
`busy_timeout`, `FULL`). **Prior art:** adopts `modernc.org/sqlite`
v1.59.0 on the spec's scoring, re-verified 2026-09-23: BSD-3-Clause
(`LICENSE`), released 2026-09-15 (the module's `.info`), newest on the
proxy, `go 1.25.0` in its `go.mod`, requiring `golang.org/x/sys v0.47.0`
and four `modernc.org` modules. The fallback, `ncruces/go-sqlite3`, stays
recorded and is not taken, since both adoption conditions hold.
`zombiezen.com/go/sqlite` and `mattn/go-sqlite3` stay declined on the
spec's grounds. Declines the driver's `OFDLocking` (R-6a4).

**Deliverable count: 4** (the dependency with its adoption tests, the
escaping DSN builder with the classes' DSNs and the readback, the typed
busy error, the typed read-only error). The fold adds rows, not a
deliverable.

---

## Task 3: The write method

Model: `opus`. Depends on Task 2. **Lands 2.0 hook 9's seam.**

This is the one way to write. Every later rung's transaction runs
through it, so a wrong commit rule here is wrong everywhere.

**Files:**
- Create: `internal/engine/store/write.go`, `internal/engine/store/write_test.go`

**Interfaces (all unexported):** the write method on the handle, taking
`ctx` and a callback that receives the store-owned transaction handle
(R-6a10); the transaction handle's three statement methods; the named
errors below.

**Outcomes:**

- **It commits only when the callback returns nil and the handle recorded
  no statement error**, and rolls back otherwise (R-6a10). SQLite's
  `lang_transaction.html` section 3, as the spec quotes, lets a statement
  fail and the transaction continue; the probe measured exactly that.
- **After a recorded error, no statement reaches SQLite** (R-6a10): every
  later statement method returns the recorded error, because a
  transaction SQLite already rolled back leaves the connection in
  autocommit mode.
- **It requires a deadline.** A `ctx` with no deadline
  (`ctx.Deadline()` reports false) is refused before any wait. The
  in-process wait for the one write connection is bounded by `ctx`:
  `database/sql`'s `DB.BeginTx` doc says "The provided context is used
  until the transaction is committed or rolled back. If the context is
  canceled, the sql package will roll back the transaction", and
  `DB.Conn`'s doc says it "will block until either a connection is
  returned or ctx is canceled." The wait on another process's write lock
  is bounded by the busy timeout alone (R-6a18, Ruling 4).
- **A nested call fails fast.** The probe measured that a second
  statement on a one-connection pool, issued while a transaction holds
  that connection, blocks rather than failing. So the write method marks
  its `ctx` with an unexported key and passes the marked `ctx` to the
  callback (R-6a10's signature), and a call whose `ctx` carries the mark
  returns a nested-write error at once.
- **Its first statement reads `user_version`** and returns a version
  error naming both values when it differs from the binary's highest
  migration. The binary's version is a handle field set at open, so tests
  can open a handle against a synthetic set.
- **Named errors.** Code 13 returns a disk-full error naming the store.
  A foreign-key failure (extended code 787, `SQLITE_CONSTRAINT_FOREIGNKEY`,
  measured by the probe as `FOREIGN KEY constraint failed (787)`) is
  distinguishable by its extended code. The busy error is Task 2's.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Deadline | none | refused, no wait |
| Deadline | expires waiting for the in-process connection | the context's error, wrapped |
| Deadline | shorter than the busy timeout, waiting on another process's write lock | the busy error at the busy timeout; the context does not shorten SQLite's wait (R-6a18, measured) |
| Nesting | a call inside a callback, on the callback's `ctx` | the nested-write error, at once |
| Nesting | a call inside a callback, on a closed-over outer `ctx` | waits until that `ctx` ends (documented limit, R-6a10) |
| Version | equal | the callback runs |
| Version | changed under the handle | the version error naming both; nothing commits |
| Callback | returns nil, no statement error | committed |
| Callback | returns an error | rolled back; that error returned |
| Callback | returns nil after swallowing a statement error | rolled back; the statement error returned |
| Callback | issues statements after a recorded error | each returns the recorded error; none reaches SQLite |
| Statement | SQLite already rolled back (code 13) | the disk-full error; the rollback's "no transaction is active" answer suppressed |
| Statement | SQLite already rolled back (`IOERR`, `NOMEM`, `INTERRUPT`) | the statement's error; reviewer-read, since no test can produce these on demand |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criterion 5's second half: an orphan insert through the write
  method fails with extended code 787. **Fixture state:** a parent table
  and a child with a foreign key, the parent row absent.
- Spec criterion 6: a callback that swallows a failed statement's error
  does not commit, **proved by a first statement that succeeded** and is
  absent afterward (the probe's shape, so the test fails under a plain
  `*sql.Tx`); a nested call on the callback's `ctx` returns the
  nested-write error well inside the outer deadline; a write with no
  deadline is refused.
- The autocommit row (R-6a10, the fold probe's P1 shape): **Fixture
  state:** a table holding one row, and `max_page_count` pinned at the
  current `page_count` inside the callback so a small row still fits in
  an existing page. The callback runs a small insert that succeeds, an
  oversized insert that fails with code 13, and a second small insert
  whose error it ignores, then returns nil. Afterward the table holds
  only the fixture's row. The pragma is per-connection and an interrupt
  discards the connection (R-6a18), so the test closes the handle rather
  than trusting a reset.
- Ruling 4's row, under yes: a second handle holds an `IMMEDIATE`
  transaction; a write with a 100 ms deadline and a busy timeout of a few
  hundred milliseconds returns the busy error at no less than the busy
  timeout. The test names the measured behavior it pins.
- Spec criterion 7: the write method fails closed when `user_version` is
  changed under it. **Fixture state:** a second handle bumps
  `user_version` between the first handle's open and its write.
- Spec criterion 3 through the write method: two handles on one file,
  both writing through their write methods, the first holding its
  callback for half the busy timeout; the second waits and commits.
- A disk-full row: a callback that sets `PRAGMA max_page_count` low and
  then inserts past it returns the disk-full error, and no row commits.
  (The probe's form; the pragma is per-connection, so the test resets it
  or closes the handle.)
- The single-row query's scan error and the multi-row query's iteration
  error are each recorded: a callback that ignores a failing scan does
  not commit.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Hand the callback `*sql.Tx` and commit on a nil return → the swallow
   row commits the first insert.
2. Drop the deadline check → the no-deadline row waits.
3. Drop the nesting mark → the nested row blocks until the outer
   deadline.
4. Read `user_version` on the read pool before the transaction → the
   changed-version row commits.
5. Return the rollback's error after code 13 → the disk-full row returns
   "no transaction is active".
6. Record `Exec` errors but not `Scan` errors → the scan row commits.
7. Remove `_txlock=immediate` from the write pool (Task 2's DSN) → the
   two-handle row's second writer returns busy at once instead of
   waiting (the deferred upgrade skips the busy handler, measured by the
   spec's mechanics review). Report whether it fires.
8. Pass statements through to SQLite after a recorded error → the
   autocommit row keeps the late small row.
9. Hand the callback the caller's `ctx` instead of the marked one → the
   nested row blocks until the outer deadline.

**2.0 hook: lands hook 9's seam**, the callback a later projector joins
so its delta commits with its event. **Prior art:** none; `database/sql`
alone.

**Deliverable count: 4** (the commit rule with the recording handle and
its no-statement-after-error rule, the deadline and nesting rules, the
version check, the named errors). Ruling 4's no would add a fifth, the
per-call busy timeout, and split it into its own task.

---

## Task 4: The migration runner and pre-migration snapshots

Model: `opus`. Depends on Task 3 (the handle and its maintenance
connection). **Lands 2.0 hook 7.**

**Files:**
- Create: `internal/engine/store/migrate.go`, `internal/engine/store/snapshot.go`,
  `internal/engine/store/migrate_test.go`, `internal/engine/store/snapshot_test.go`

**Interfaces (all unexported):** the runner as a function taking `ctx`,
the handle, an `fs.FS` of numbered SQL files (spec seam 2), whether backup
is on, and the local keep count; the snapshot function used by the runner
now and by 6b-i's backup later; a free-space check.

**Outcomes:**

- **Files are numbered SQL files**, applied in numeric order, each in its
  own `IMMEDIATE` transaction on the write pool that reads `user_version`,
  applies the file, sets `PRAGMA user_version = N` by formatting `N` into
  the statement, and inserts the ledger row (version, file name, SHA-256
  of the file's bytes). The probe measured a bound pragma parameter
  failing (`near "?": syntax error (1)`), and `user_version`, DDL, and
  `application_id` all rolling back with their transaction.
- **A file's first error ends it.** The runner returns on the first error
  from applying the file and runs no version or ledger statement after
  it. The risk review measured that a multi-statement `Exec` stops at its
  first failing statement. A `Commit` that answers "no transaction is
  active" after a clean apply is a hard error naming the file, since it
  means the file carried its own transaction control (the risk review
  measured `create table e3(a); commit; create table e4(a)` inside a
  `*sql.Tx` returning nil and persisting both tables). Migration files
  carry no `COMMIT`, `END`, `ROLLBACK`, `SAVEPOINT`, `RELEASE`, `VACUUM`,
  or `ATTACH`; the diff-reviewer checks each new file for them.
- **The version check runs first, before the ledger is read** (R-6a5):
  a `user_version` higher than the binary's highest is refused naming
  both versions, the same "laptop ahead" case the configuration homes
  spec names (`:180-183`), whatever the ledger holds.
- **The ledger check** runs next, only when `user_version` is at or below
  the binary's highest, and before any apply: every ledger row's name and
  hash must equal the binary's file at that version, every version from 1
  through `user_version` must have a row, and no row may sit above
  `user_version`, or the runner refuses naming both sides.
- **The snapshot** (R-6a11) runs before the first pending file of a
  non-empty store: free-space check, pre-create `staging/pre-migrate-vN.db`
  at 0600 (Task 1), `VACUUM INTO ?` on the maintenance connection with
  the target bound as a parameter (the mechanics review measured `VACUUM
  INTO ?` succeeding with the target `it's copy.db`, where a concatenated
  string breaks on the `'`), open the copy through the inspect DSN
  (`mode=ro`, R-6a17) and run `PRAGMA integrity_check` expecting the
  single row `ok` and read `user_version` expecting N and
  `application_id` expecting the marker, close it, rename it into
  `snapshots/`, then open `snapshots/` and `Sync` it (the directory
  entry). The verification open is `mode=ro` because a copy is a
  rollback-journal database (the risk review read header bytes 18 and 19
  as 1 and 1), and a writing DSN's `_journal_mode=WAL` would convert it
  and leave `-wal` and `-shm` in `staging/`; and because an empty file
  also passes `integrity_check`, so the version and marker reads are what
  prove the copy. The copy is durable without a hand `Sync` of the file:
  SQLite's source says "For a VACUUM INTO, the pager-flags are set to the
  same values as they are for the database being vacuumed, except that
  PAGER_CACHESPILL is always set" (`lib/sqlite_g_0000000000000003.go`,
  v1.59.0), so the maintenance class's `_synchronous=FULL` governs it. `internal/atomicfile` has no rename of
  an existing file (`WriteFile` and `WriteStream` each write their own
  temp file), so the rename and the directory fsync are by hand, as the
  spec says (`:589-592`). Any failure refuses the migration with the
  store at N and names the snapshot path. The snapshot ends with
  `PRAGMA wal_checkpoint(TRUNCATE)` on the store.
- **Quoted and measured for the snapshot.** `lang_vacuum.html`, as the
  spec quotes: `VACUUM INTO` yields "a consistent snapshot of the original
  database"; its target "must not previously exist, or else it must be an
  empty file"; and "A VACUUM will fail if there is an open transaction on
  the database connection that is attempting to run the VACUUM." The
  probe measured: `VACUUM INTO` into a pre-created empty 0600 file
  succeeds and keeps 0600; `VACUUM INTO` on a `query_only` connection
  fails with code 8; a target that already holds a database fails with
  `output file already exists (1)`; a 4096-byte non-database target fails
  with code 26; the copy passes `integrity_check`; and `PRAGMA
  wal_checkpoint(TRUNCATE)` returned `0 0 0` and left a 0-byte `-wal`.
- **Retention** runs after a successful migration with backup off, keeping
  the newest `keep` copies by version over `pre-migrate-v<N>.db` names
  only, and the runner refuses `keep` below 1 (R-6a11).
- **A foreign-keys-off rebuild** is a migration file carrying a marker
  the runner reads (Task 4 chooses the marker's form and documents it in
  the package comment). It runs on the maintenance connection with
  `PRAGMA foreign_keys = OFF` issued **before** its `BEGIN IMMEDIATE`,
  runs `PRAGMA foreign_key_check` before commit, refuses to commit on any
  row, and closes the connection. `pragma.html`, as the spec quotes:
  enforcement "may only be enabled or disabled when there is no pending
  BEGIN or SAVEPOINT." The probe measured `PRAGMA foreign_keys = 0`
  inside a transaction leaving the value at 1.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Ledger | agrees | proceed |
| Ledger | a row's name or hash differs | refused, naming both |
| Ledger | a row for a version with no binary file | refused, naming the version |
| Ledger | a version at or below `user_version` with no row | refused, naming the version and `user_version` |
| Ledger | a row above `user_version` | refused, naming the row and `user_version` |
| A file | `Commit` answers "no transaction is active" after a clean apply | a hard error naming the file |
| Version | 0 (empty) | apply all, no snapshot |
| Version | N below M | snapshot, then apply N+1 through M |
| Version | equal | nothing applies |
| Version | above M | refused, naming both; decided before the ledger is read, so no ledger row fires |
| Free space | enough | proceed |
| Free space | short | refused, naming the need, the free bytes, and the path |
| Snapshot | copy fails `integrity_check`, or reads a version other than N or no marker | refused, copy removed from staging, store at N |
| Snapshot | same-version copy exists | replaced by rename |
| Snapshot | the staging target already exists | refused, naming the path (serve's cleanup runs before the runner, so this is a bug state at 6a) |
| A file | fails mid-apply | that file's transaction rolled back; the version, schema, and ledger at the prior file |
| Rebuild | `foreign_key_check` returns rows | refused, naming the table; nothing committed |
| Retention | backup on | nothing deleted |
| Retention | backup off, more than `keep` copies | the oldest by version deleted |
| Retention | a file in `snapshots/` whose name is not `pre-migrate-v<N>.db` | untouched, no error |
| Retention | `keep` below 1 | refused before any apply |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criterion 8, in three rows, each asserting that `user_version`,
  the `sqlite_schema` rows, and the ledger rows equal their pre-run
  values. **Fixture state for all three:** a store at the synthetic set's
  version 1.
  - **After the runner's version write** (the criterion's own words):
    synthetic file 2 creates a table and a `BEFORE INSERT` trigger on the
    ledger table that runs `SELECT RAISE(ABORT, 'injected')`, so the
    runner's `PRAGMA user_version = 2` executes and its ledger insert then
    fails. The fold probe (P8) measured this shape: the insert failed with
    `constraint failed: injected (1811)`, and after the rollback
    `user_version` read 1 and the table and trigger were gone. Mutation 1
    fires on this row.
  - **Inside the file, after the file's own version write:** synthetic
    file 2 runs `PRAGMA user_version = 2`, creates a table, then fails on
    a later statement.
  - **A ledger hash mismatch** is refused naming both hashes: a store
    migrated to version 2, then a set whose file 2 differs by one byte.
- Spec criterion 9 on synthetic sets: migrating a populated store at N
  leaves `pre-migrate-vN.db` whose `user_version` is N and which holds the
  populated rows; a snapshot write that fails (the `staging/` directory
  made unwritable) leaves the store at N and the error names the path; a
  second migration from N+1 leaves the `vN` copy byte-identical; with
  backup off and `keep` 2, three migrations leave the newest two copies
  and a planted `pre-restore-x.db` untouched. The kept copy reads
  `user_version` N and the marker. **Fixture state:** the store holds
  rows before the first migration, so an empty copy fails.
- The ledger rows: a store at version 3 whose ledger holds rows 1 and 2
  only, and a store at version 1 with a planted row 2, are each refused
  naming the version (rows planted by a raw connection).
- A synthetic file carrying `commit` mid-file yields the hard error
  naming the file. (Criterion 9's embedded-set half, the job-records
  fixture migrated through the skeleton, is Task 5's.)
- Spec criterion 10: after an injected foreign-keys-off rebuild migration,
  an orphan insert through the write method fails with 787, proving the
  maintenance connection never returned to a pool with enforcement off. A
  second injected rebuild that leaves an orphan is refused by
  `foreign_key_check` with nothing committed.
- A free-space row through a stubbed statfs result (spec seam 1's
  variable) refuses naming the numbers.
- Every snapshot and every staged file is 0600.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Set `user_version` in a separate transaction from the file → the
   failure-after-version row leaves the new version.
2. Skip the ledger hash comparison → the mismatch row migrates.
3. Take the snapshot after the first file → the snapshot holds N+1.
4. Run `VACUUM INTO` on the write pool → report the effect on a
   concurrent in-process writer (the spec's stall), or name it a
   reviewer-read row.
5. Write the copy straight into `snapshots/` → a failed copy is visible
   there (report whether the failing-snapshot row fires).
6. Issue `PRAGMA foreign_keys = OFF` inside the transaction → the rebuild
   runs with enforcement on (report what the synthetic rebuild does).
7. Return the rebuild connection to the pool → the orphan insert
   succeeds.
8. Delete by modification time instead of version → report whether the
   retention row fires.
9. Verify the copy through the write DSN → the copy gains `-wal` and
   `-shm` in `staging/` (report whether the 0600 row or a listing row
   fires).
10. Match retention on a `*.db` glob → the planted `pre-restore-x.db` is
    deleted.
11. Skip the ledger-completeness check → the version-3-with-two-rows
    store migrates.

A reviewer-read row: dropping `_synchronous=FULL` from the maintenance
class weakens the copy's durability, which no test can observe.

**2.0 hook: lands hook 7**, the migration mechanism chosen with an
evolving schema in view. **Prior art:** hand-rolled over `embed.FS` and
`PRAGMA user_version`, as S20's prior-art row records. Declines `goose`
(an older binary's `Provider.Up` on a newer database returns no error, the
silent-ahead case this runner refuses) and `golang-migrate` (its modernc
driver pinned to v1.18.1), both on the spec's measured grounds. Adopts
SQLite's `VACUUM INTO`; the driver's backup API (`backup.go`) stays the
recorded alternative.

**Deliverable count: 4** (the runner with its ledger and refusals, the
snapshot, retention, the foreign-keys-off rebuild). The fold adds rows,
not a deliverable.

---

## Task 5: The first three migrations, their fixtures, and the invariant test

Model: `sonnet`. Depends on Task 4. **Lands 2.0 hooks 4, 5 (the table),
and 8.**

**Files:**
- Create: `internal/engine/store/migrations/0001_ledger.sql`,
  `0002_jobs.sql`, `0003_catalog.sql` (working names), `internal/engine/store/migrations.go`
  (the `embed.FS`), `internal/engine/store/id.go`,
  `internal/engine/store/invariant_test.go`, `internal/engine/store/schema_test.go`
- Create: `internal/engine/store/testdata/fixtures/` (one populated
  fixture per migration) and `internal/engine/store/testdata/violations/`
  (nine violating fixtures across the seven invariants)

**Outcomes:**

- **Migration 1, the ledger**: the applied-migration table keyed by
  version (exempt from invariant 6), recording itself, and `PRAGMA
  application_id` set to dubplate's marker (R-6a5).
- **Migration 2, the job records**: one entity table (`BLOB` id, the
  opaque job handle), with `kind` and `at` `NOT NULL` (invariant 3), a
  state column with a CHECK admitting exactly queued, running, succeeded,
  failed, canceled, and interrupted; the owner instance id; the actor as
  text; `cancel_requested` as a 0-or-1 integer with a CHECK; progress as
  text; a failure code and message; and started, updated, and finished
  times (R-6a8). The table is the contract 6b-i writes; its contents
  follow the spec's job model (`:513-545`). The one index 6b-i's pruning
  and `doctor`'s last-backup read need, kind plus state plus finish time,
  lands here too.
- **Migration 3, the catalog skeleton** (spec `:443-457`):
  - `users`: `id` (`BLOB`, 16 bytes, hook 5); the normalized address, `TEXT
    NOT NULL UNIQUE` across live and tombstoned rows (ruling 9); the display
    name; `streamer_username`, `TEXT NOT NULL UNIQUE COLLATE NOCASE`
    (ruling 8), which matches 5c-i's `strings.EqualFold` uniqueness because
    5c-i's write-path shape is ASCII (R-5c6) and `NOCASE` folds ASCII; the
    admin flag with a 0-or-1 CHECK; a nullable tombstone time; created and
    updated times.
  - `collections`: `id`; `name`; `root_path` `BLOB NOT NULL UNIQUE` with
    R-6a9's CHECK; `owner` `BLOB NOT NULL` referencing `users(id)`;
    `visibility` with a CHECK admitting `group` and `private`; `writer`
    with a CHECK admitting `server` and `local` (hook 8's domain; the
    named local binary's reference column arrives at 6c with
    `device_credentials`, by `ALTER TABLE ... ADD COLUMN ... REFERENCES`,
    which the probe measured SQLite permits on a `STRICT` table with a
    `NULL` default and enforces afterward); the writer epoch and the
    mirror generation as integers defaulting to 0; the nullable
    last-mirror time; the duplicate-collapse flag; the layout template and
    tagging policy as text; the admission overrides as typed nullable
    columns (a format allowlist as text and a match-threshold override as
    a real), never an attribute bag.
  - `albums`: `id`; `collection_id` referencing `collections(id)`;
    `mb_release_id`; created and updated times; a UNIQUE index on `(id,
    collection_id)`.
  - `tracks`: `id`; `album_id` and `collection_id`, together a composite
    foreign key onto `albums(id, collection_id)`, so a track's collection
    is its album's; `rel_path` `BLOB NOT NULL` with R-6a9's CHECK;
    `mb_recording_id`; created and updated times; a UNIQUE index on
    `(collection_id, rel_path)`.
  - No foreign key declares `ON DELETE CASCADE` (invariant 7).
- **The id function** mints 16 bytes from `crypto/rand.Read`. `go doc
  crypto/rand.Read` (go1.27.1): "Read fills b with cryptographically
  secure random bytes. It never returns an error, and always fills b
  entirely."
- **The invariant test** applies every migration to an empty database and
  to each migration's populated fixture (R-6a15), then checks the seven
  rules (spec `:401-432`) through `pragma_table_list`,
  `pragma_table_info`, `pragma_foreign_key_list`, and `pragma_index_list`,
  never by parsing SQL. The probe measured `pragma_table_list.strict`
  reading 1 for a `STRICT` table and 0 otherwise. Rule 6's exemption list
  is named in the test and holds the ledger alone at 6a; 6c adds its own
  entries with its tables.
- **Nine violating fixtures across the seven rules** prove each rule can
  fail for its named catalog fact: a non-`STRICT` table; a `path` column
  declared `TEXT`; a foreign key to a missing column; a foreign key to a
  non-unique parent; an `_events` table missing `at`; a bare `release_id`
  column; a foreign key plus `key` plus `value` table; an `INTEGER PRIMARY
  KEY` named `id`; and an `ON DELETE CASCADE` onto `users`.

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criterion 11: each invariant fails against its violating fixture
  **for its named fact** (the test asserts the failure message names the
  rule and the table), and passes over the skeleton. A Go `string` bound
  into `tracks.rel_path` is refused; the probe measured `cannot store
  TEXT value in BLOB column c.path (3091)` on a `STRICT` table and
  `typeof` text with no error on an ordinary one.
- Spec criterion 12: a duplicate address is refused by the `users` index,
  once against a live row and once against a tombstoned one. **Fixture
  state:** the existing row inserted directly in the test, since no
  accessor exists, and the duplicate differs from it in every column but
  the address, so no other unique column can refuse it. The test asserts
  extended code 2067 and the message naming the address column; the fold
  probe (P5) measured SQLite's form as `constraint failed: UNIQUE
  constraint failed: users.address (2067)`, and `users.streamer_username`
  for a case-folded username duplicate.
- Spec criterion 9's embedded half: migrating the populated job-records
  fixture (version 2) forward through the skeleton leaves
  `pre-migrate-v2.db` with `user_version` 2 and the fixture's job rows.
- The path CHECKs refuse each R-6a9 bad form on both columns and accept
  the three good forms, as table rows.
- A track whose `collection_id` differs from its album's is refused with
  787. A second collection owned by a missing user is refused with 787.
  **Fixture state:** the album, both collections, and the owning user of
  the first collection exist, so the only failure is the one the row
  names.
- A table of file name to SHA-256 pins every embedded migration's bytes,
  so an edit to a merged file (a trailing newline included) fails the
  gate instead of every existing store's ledger check at the next start.
  A new migration adds its row.
- The id function test asserts 16 bytes and two successive values
  differing, and an insert of a minted id into an entity table leaves
  `rowid` unequal to the id's integer reading; the test states that this
  proves only that the id is not the rowid, not randomness.
- `application_id` reads the marker over every fixture.
- The `diff-reviewer` checklist line from spec `:430-432`: a path column
  not named as one, or a source identifier under another name, passes the
  test, so the reviewer reads each new column's meaning.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Drop `STRICT` from `tracks` → invariant 1 fails on the skeleton.
2. Make the albums unique index cover `id` alone → the composite foreign
   key fails with "foreign key mismatch" at insert (invariant 2 names it).
3. Drop `NOT NULL` from the job table's `at` → invariant 3 fails.
4. Declare `ON DELETE CASCADE` on `collections.owner` → invariant 7
   fails.
5. Index `users.address` partially (live rows only) → the tombstoned
   duplicate row inserts.
6. Remove one migration's fixture → the invariant test fails naming the
   migration.
7. Drop the `/../` term from the path CHECK → the `a/../b` row inserts.
8. Append a newline to `0002_jobs.sql` → the hash-pin row fails.

**2.0 hooks: lands hook 4** (invariant 3 binding the job table first),
**hook 5's table** (the surrogate id column and mint function; the
resolution to it is 6d's), and **hook 8** (`writer`, the epoch, the
generation). **Deferred by ruling 2:** hooks 1, 2, 3, and 11. **Prior
art:** consulted Navidrome's
`user` table for behavior only (`db/migrations/20200130083147_create_schema.go`,
read 2026-09-23): an opaque `id` primary key beside a unique `user_name`,
the same surrogate-plus-handle split hook 5 takes. No code adopted (GPL).

**Deliverable count: 4** (the three migrations with their hash pins, the
fixtures, the invariant test with its violations, the id function).

---

## Task 6: The three open modes

Model: `opus`. Depends on Task 5 (the embedded set fixes the binary's
version). **Neutral on 2.0 hooks** (it enforces hook 7's two refusal
directions).

**Files:**
- Create: `internal/engine/store/open.go`, `internal/engine/store/inspect.go`,
  `internal/engine/store/open_test.go`, `internal/engine/store/inspect_test.go`

**Interfaces:**
- Produces, **exported** (R-6a2): one inspect function taking `ctx` and
  its options (state directory, busy timeout) and returning a report; the
  report type with a closed state set (absent, restore in progress, busy,
  not a store, refused, older, newer, current), the binary's and the
  store's versions, pending migration names, holder pids, and a warning.
- Produces, unexported: the classification every mode shares, the
  write-mode open, and the serve-mode open taking the migration wait, the
  keep count, and a `*slog.Logger` as parameters, with its staging
  cleanup.

**Outcomes:**

- **Every mode's order**, the one sequence R-6a7 and Task 1 point at:
  1. Stat the leaf and its four names (`store.lock`, `db/`, `snapshots/`,
     `staging/`). An absent leaf makes inspect report absent and write
     refuse.
  2. Owner check, when the leaf exists.
  3. Restore-intent check (R-6a7).
  4. Serve only: run Task 1's idempotent leaf create whenever the leaf or
     any of the four names is missing, under Task 1's state-directory
     owner rule, so a present but partial leaf is completed before the
     lock is taken.
  5. Filesystem probe.
  6. Take the lock in its mode (inspect and write open `store.lock`
     without `O_CREATE`).
  7. Classify (R-6a5).
  8. Open the mode's DSN.
- **The classification is one function every mode shares** (R-6a5). It
  decides absence by `os.Stat` alone (`fs.ErrNotExist`), reads the size
  and the `-wal` beside a zero-byte file by `os.Stat` (R-6a4), and reads
  the marker, the version, the schema count, and the ledger through the
  inspect DSN (`mode=ro`). No writing DSN opens a file that has not
  classified as a dubplate store or an interrupted create, because a
  writing DSN's `_journal_mode=WAL` rewrites a rollback-journal file's
  header on open.
- **Inspect** takes the lock shared and non-blocking, classifies, and
  reports. The revision 1 probe measured a `mode=ro` open of an absent
  file failing with `unable to open database file (14)` and creating
  nothing; and a `mode=ro` open of a WAL database with no `-wal` or
  `-shm` beside it reading successfully, leaving the database file's
  bytes and mode unchanged and creating `-wal` and `-shm`, which
  `wal.html` section 5 permits when the directory is writable. The fold
  probe (P4) measured a hot WAL: a child process wrote three rows with
  `wal_autocheckpoint(0)` and exited through `os.Exit` without a close,
  leaving a 12392-byte `-wal`; a `mode=ro` open then read all three rows
  and left the main file's hash and the `-wal` unchanged, while the same
  open without `mode=ro` changed the main file's hash and removed the
  `-wal` on close.
- **Write mode** takes the lock shared, bounded by `ctx`; refuses an
  absent store naming "start serve first" and creates nothing; refuses
  any version but the binary's, naming both and "start serve to migrate".
  It never removes a staging file (R-6a7).
- **Serve mode** follows the spec's four-step procedure (`:344-358`):
  1. Take the lock shared, classify, and open. This wait is bounded by
     the migration wait and logs the holders the same way step 2 does, so
     a restore holding the exclusive lock never hangs serve silently.
  2. If not current, close every connection, release, and take exclusive
     bounded by the migration wait, logging holders through a
     `*slog.Logger` parameter on the first failed poll, then at R-6a14's
     holders log interval.
  3. Holding exclusive, re-stat and re-classify, remove staging leftovers
     when no intent record exists (R-6a7), then create, complete an
     interrupted create (R-6a6), or migrate. A create pre-creates the
     database at 0600 (Task 1) and applies every migration from version
     0. Under Ruling 3's yes, a create and the completion of an
     interrupted create each first read `snapshots/` and refuse beside a
     copy.
  4. Close every connection, convert to shared, and return to step 1.

  `flock(2)`, as the spec quotes: "Converting a lock ... is not
  guaranteed to be atomic: the existing lock is first removed, and then a
  new lock is established. Between these two steps, a pending lock
  request by another process may be granted."

**States each check meets, and its report:** the R-6a5 table, plus:

| Check | State | Inspect | Write | Serve |
| --- | --- | --- | --- | --- |
| Intent record | present | restore in progress | refused, naming `restore --resume` | refused, naming `restore --resume` (R-6a7) |
| Lock | exclusive held by another | busy, naming holders | busy after `ctx` | waits up to the migration wait for its shared lock, logging holders; then exits naming them |
| Lock | shared held by another, serve needs exclusive | n/a | n/a | waits up to the migration wait, logging holders; then exits naming them |
| Conversion | another process wins the exclusive lock in the gap | n/a | n/a | holds no connection; re-reads after the winner releases |
| Staging | a leftover file, no intent | untouched | untouched | removed in step 3, then the create or migration proceeds |
| Staging | a leftover file beside an intent | untouched | untouched | untouched; refused per R-6a7 |
| Owner, filesystem | refused | refused | refused | refused |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criterion 13: a zero-byte file and a text file are each refused as
  not a store in write mode; write mode against an absent store refuses
  naming "start serve first" and **the leaf is still absent afterward**;
  against an older store (the embedded set's first two migrations
  applied, through the runner on a sub-`fs.FS`) refuses naming both
  versions; a newer store is refused naming both versions. **Newer
  fixture's write:** a raw connection sets `user_version` to head+1 and
  inserts that version's ledger row, as a newer binary would, so a
  ledger-first classification reports "refused" and fails the row. A foreign SQLite database with its own
  `application_id`, and a foreign rollback-journal database with
  `user_version` set and no `application_id`, are each refused as not a
  store by write and serve modes with the file's bytes unchanged (hashed
  after every handle closes, R-6a4).
- A present leaf with no database file: with `store.lock` and `db/`
  present, write mode refuses naming "start serve first" and inspect
  reports absent; with `db/` present and no `store.lock`, inspect reports
  absent. **Each asserts the leaf's recursive listing is identical before
  and after**, the lock file included.
- The interrupted creates (R-6a6). **Fixture states:** a zero-byte
  database with no `-wal`; and a header-only database built the way the
  mechanics review's D-2 was (pre-create, open the write DSN, begin, roll
  back, close), asserted to be 4096 bytes before the row runs. Inspect
  and write report each as not a store naming "start serve to finish
  creating it"; serve completes each to current. A zero-byte database
  with a `-wal` beside it is not a store in every mode, and neither file
  changes.
- An unreadable database: a 0000-mode database file in an owner-owned
  leaf reports "refused" naming the path, never absent. **Fixture
  state:** the gate runs unprivileged; under root no mode denies the
  read, and the test fails saying so rather than passing.
- Spec criterion 14, in two rows. **Fixture state for both:** the store
  is older, so serve mode must convert; every other SQLite connection to
  the file in the test process is closed, because SQLite's unix VFS
  defers a descriptor's close while another connection on the inode holds
  locks ("If there are outstanding locks, do not actually close the file
  just yet", `lib/sqlite_g_000000000001ffff.go:12569`, v1.59.0); and
  the test's own lock is a second descriptor on `store.lock`, which
  `flock(2)` treats as a second process (quoted in Task 1).
  - **The wait:** the test holds shared on descriptor A, then starts serve
    mode with a logger whose handler signals a channel on the holders
    record. Serve takes shared, reads the older version, releases, and
    fails its exclusive poll against A. On the signal, the test asserts
    the version is unchanged, releases A, and asserts serve migrates to
    current.
  - **The conversion:** the same start, but on the signal serve holds no
    lock and no connection, and the test converts A to exclusive with
    `LOCK_EX|LOCK_NB` and asserts the conversion succeeded (the fold
    probe, P7, measured it succeeding 2000 of 2000 times against a tight
    `LOCK_NB` poller on a second descriptor; a failure fails the test
    rather than passing it). Holding it, the test renames the database
    aside and places a store at the binary's current version, built in
    another directory and closed. It then `os.Stat`s every
    `/proc/self/fd/N` and asserts none is `os.SameFile` with the renamed
    file. The fold probe (P6) measured `os.Stat` of an fd link matching
    the renamed file by `os.SameFile` while `readlink` returned the new
    name, so a path comparison would pass under the mutation. The test
    releases A and asserts serve returned current with the placed store's
    ledger and schema unchanged, which a serve that skipped the re-read
    would have migrated.
- Spec criterion 15, in-package half: inspect against an absent, an
  older, and a newer store reports each state by name, and afterward the
  database file's bytes and mode are unchanged. **Fixture state: a hot
  WAL** for the older and newer rows. After the store is built and
  closed, the test re-executes its own binary (`os.Args[0]` with
  `-test.run` naming a helper and an environment variable naming the
  store); the child makes the write that decides the state with
  `wal_autocheckpoint(0)` on its connection and exits through `os.Exit`
  without a close, so the `-wal` holds that frame. **Older:** the store
  is built at version 1, and the child applies migration 2 through the
  runner on a sub-`fs.FS` holding the first two files, leaving the store
  at 2 with its ledger complete. **Newer:** the store is built at the
  head, and the child sets `user_version` to head+1 and inserts that
  version's ledger row, as a newer binary would. The test hashes the
  main file with no handle open, runs inspect, and hashes again; the
  report's state proves inspect read through the WAL. Every other report
  state is produced once from a real fixture: restore in progress (an
  empty `restore-intent` file), busy (an exclusive `flock` held by the
  test), not a store, refused (owner), current.
- Spec criterion 16's third part: serve mode's create leaves the database
  at 0600 under a 0022 umask, and a snapshot likewise (Task 4's row
  re-run under this fixture).
- Staging: a leftover `staging/pre-migrate-v1.db` beside an older store
  is removed by serve, which then migrates; beside an intent record it is
  untouched and serve refuses; write mode leaves a planted staging file
  untouched.
- Ruling 3's row: under yes, a leaf with no database file and a
  `snapshots/pre-migrate-v1.db` makes serve refuse naming that copy and
  `restore --from-file`, and inspect report "refused", with the listing
  unchanged; under no, serve creates. A second row: under yes, a
  zero-byte database with no `-wal` beside `snapshots/pre-migrate-v1.db`
  makes serve refuse naming that copy and inspect report "refused", and
  the database is still zero bytes afterward; under no, serve completes
  it to current.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Keep the connection open across the conversion → the `/proc/self/fd`
   row finds a descriptor on the renamed file.
2. Skip the re-read after taking exclusive → the conversion row migrates
   the placed current store from the stale version.
3. Let write mode create → the absent rows leave a leaf or a database.
4. Run staging cleanup in write mode → the write-mode staging row loses
   its planted file.
5. Treat an unmarked empty database as not a store in serve mode → the
   header-only row refuses instead of completing.
6. Open inspect without `mode=ro` → the hot-WAL row's hash changes (P4
   measured this).
7. Map code 14 to absent → the 0000 row reports absent.
8. Classify through the write DSN → the foreign rollback-journal row's
   bytes change.
9. Skip serve's staging cleanup → the leftover row refuses the snapshot.
10. Under Ruling 3's yes, create or complete without reading
    `snapshots/` → the missing-database row creates an empty store, and
    the zero-byte-beside-copy row completes one.
11. Read the ledger before deciding "newer" → criterion 13's and
    criterion 15's newer rows report "refused" instead of newer.

A reviewer-read row: detecting a zero-byte file by opening it, which
R-6a4 forbids and no test observes.

**Prior art:** none beyond Task 1's lock shape.

**Deliverable count: 4** (the classification every mode shares, the
inspect mode and its exported report, the write mode, the serve mode's
procedure with its create, completion, and staging cleanup). Revision 1
counted staging cleanup as its own deliverable; it moves into serve mode
(R-6a7), and the shared classification takes its place. **Pre-named
split:** if the implementer reports that the four cannot land in one
green gate, the serve mode moves to its own task between this one and
Task 7, which depends only on the inspect mode.

---

## Task 7: `doctor`'s store row, `check`'s inspect report, and the busy timeout

Model: `sonnet`. Depends on Task 6. **Neutral on 2.0 hooks.**

This task lands the first production callers of the store.

**Files:**
- Create: `cmd/dubplate-server/store.go` (the threshold's declaration and
  registration function, and the report-to-line rendering),
  `cmd/dubplate-server/store_test.go`
- Modify: `cmd/dubplate-server/main.go` (one registration call in
  `newProductionRegistry`), `cmd/dubplate-server/doctor.go`,
  `cmd/dubplate-server/doctor_test.go`, `cmd/dubplate-server/check.go`,
  `cmd/dubplate-server/check_test.go`

**Outcomes:**

- **`DUBPLATE_STORE_BUSY_TIMEOUT`** registers as a `KindDuration`
  threshold, default `15s`, in `cmd/dubplate-server/store.go` (R-6a3),
  with a registration function `newProductionRegistry` calls after
  `registerAuthConfig`. `doctor` shows it on the threshold family's
  ordinary line.
- **`doctor`** prints one `store` row after the identity-source rows and
  the roster row, in the pinned `%-32s %s` form, for every state
  (R-6a12). The state directory comes from `lookupEnviron(environ,
  "DUBPLATE_STATE_DIR")`, so the row prints when `config.Load` fails;
  with the key unset, the row reads "not checked" and the owned row
  already reports "missing". A busy timeout that fails to parse is
  reported by the threshold row, as every duration threshold is today,
  and the store row reads "not checked". A busy timeout that parses but
  falls outside R-6a3's range (below 1 ms, or above 2147483647 ms) leaves
  the threshold row showing the parsed value, as `thresholdStatus` does
  for every duration, and **the store row carries the refusal**: "not
  checked: DUBPLATE_STORE_BUSY_TIMEOUT must be at least 1ms" or "at most
  2147483647ms". `doctor`'s exit follows Ruling 2, and an out-of-range
  value exits 1.
- **`check`** runs the inspect mode after `config.Load` succeeds, prints
  `store: <state>` after `configuration: ok`, and exits by Ruling 2. An
  unparseable or out-of-range busy timeout fails `check` naming the key
  and the bound before the inspect mode runs.
  Main stays the only stderr writer (`check.go`'s existing contract).
- **The rendering** is one function from a report to a status string, in
  `store.go`, shared by both verbs: absent; older, naming the pending
  migration files; newer, naming both versions; current, naming the
  version; restore in progress, naming `restore --resume`; busy, naming
  holder pids or "held by another process"; not a store, naming why;
  refused, naming the reason and its fix. A FUSE warning appends to the
  status.

**States each check meets, and its report:** every report state above,
plus:

| Check | State | `doctor` row | Exit |
| --- | --- | --- | --- |
| State dir | unset | `not checked` | the owned row's failure already exits 1 |
| Busy timeout | unparseable | `not checked`; the threshold row names the reason | 1 |
| Busy timeout | parses, below 1 ms or above 2147483647 ms | `not checked: ...` naming the key and the bound; the threshold row shows the value | 1 |
| Report | absent, older, current | the state | 0 unless another row fails |
| Report | any other state | the state | 1 |

**Acceptance criteria a diff-reviewer verifies from the diff:**

- Spec criterion 15's golden half, under R-6a16: one exact expected line
  per report state from a constructed report, plus the two end-to-end
  cases (no leaf, and a zero-byte `db/` file).
- `TestDoctorCmd_LineCountFollowsRegistry` names the store row as one
  constant line and fails if a second store line prints (R-6a12).
- `check` against an absent store prints `configuration: ok` then
  `store: absent` and exits 0; against a zero-byte database prints the
  not-a-store line and exits nonzero. **Fixture state:** a full valid
  environment, so only the store differs.
- A `doctor` run against an absent store creates nothing under the state
  directory (the directory listing is identical before and after).
- The busy-timeout threshold row shows `15s` by default and the set value
  when set. With `0s`, `-1s`, `500us`, and `1000h` (3600000000 ms, the
  risk review's measured over-range value), the store row reads `not checked` naming the key and the
  bound, `doctor` exits 1, and `check` fails naming the key before any
  inspect. `2147483647ms` passes.
- No new exported symbol in `cmd/dubplate-server` (package `main`);
  `internal/engine/store`'s exports are exactly R-6a2's, each with a
  caller here.
- `bash scripts/check.sh` green.

**Named mutations this task reports:**

1. Print the row only on failure (the roster row's pattern) → the absent
   golden and the line-count row fail.
2. Exit nonzero on absent → the absent `check` row fails.
3. Read `DUBPLATE_STATE_DIR` from the loaded config only → the row
   vanishes on a load failure (report the test that fires).
4. Call the write-mode open instead of the inspect mode → the absent `check` row reports a refusal instead of `store: absent`.
5. Check only `d > 0` → the `500us` row passes the store a zero busy
   timeout.
6. Drop the ceiling → the `1000h` row opens with a busy timeout that
   reads back as 0.

**Deliverable count: 3** (the threshold, the shared rendering with
`doctor`'s row, `check`'s report).

---

## Task 8: The pass close, and the 6a to 6b-i handoff

Model: `sonnet`. Depends on Tasks 0 through 7 and the simplifier round.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/research/handoffs/draft-STATUS-6a.md`
- Create: `docs/research/handoffs/draft-HISTORY-6a.md`

**Outcomes:**

- **The HISTORY draft** carries what landed per task with commit ids, the
  gate evidence (`go test ./...` wall time and the store package's time
  with `-count=1`, and the gate's cold and warm wall time with the
  driver), the resolutions R-6a1 through R-6a18 with where each landed,
  the four rulings and their outcomes, the declared coverage gaps, both
  budgets, and "what a later pass would be wrong to rediscover." The gaps
  are at minimum: the uid check's real-uid mutation (reviewer-read); the
  `/proc/locks` inode-only mutation (reviewer-read); the holder list's
  empty result on btrfs, where `stat` and `/proc/locks` disagree on the
  device; the zero-byte detection by `os.Stat` (reviewer-read); the
  leaf-owner ownership of `-wal` and `-shm` in criterion 15, which cannot
  fail when every test file has the test's uid; the `IOERR`, `NOMEM`, and
  `INTERRUPT` rollback rows (reviewer-read); the copy's durability under
  `_synchronous=FULL` (reviewer-read); and the OFD decline with its
  reason.
- **The errata owed**, drafted paste-ready for the conductor from the
  fold record's "Owed errata" list, with each ruling's outcome applied.
- **The 6b-i handoff**:
  1. The surface 6b-i builds on: the write method and its transaction
     handle, which 6b-i exports with `ops/local`; the write mode,
     exported with `backup`; the snapshot function; the job table's
     columns; `leases/`, which 6b-i creates.
  2. The intent record rule 6b-ii refines (R-6a7).
  3. The staging rule 6b-i inherits: a shared-lock stager names its file
     by instance id, and its recovery removes only a file whose owner's
     lease is free, since serve's cleanup runs only under the exclusive
     lock.
  4. The snapshot's `wal_checkpoint(TRUNCATE)` result row, which 6b-i
     must read because under a shared lock with serve writing it can
     report busy in the row's first column rather than as a Go error
     (logged, never a snapshot failure).
  5. A `VACUUM INTO` copy is a rollback-journal database, so 6b-ii's
     restore sets WAL on the placed file.
  6. An interrupted file-backed connection is discarded by the driver
     (R-6a18).
  7. The runbook lines owed: a root-run verb's refusal and the `runuser`
     fix; a create refused because the state directory's owner differs;
     the FUSE warning; the btrfs holder gap.
- **The STATUS draft** stays present tense and under 60 lines.

**Acceptance criteria:** every count in the drafts was produced by a
command the report pastes; every commit id resolves; the STATUS draft is
at most 60 lines. No gate change, so `bash scripts/check.sh` green as a
formality.

**Deliverable count: 4** (the HISTORY draft, the errata, the handoff, the
STATUS draft).

---

## The exit gate

`bash scripts/check.sh` green in the worktree and again on master after
the merge, plus every assertion below. Each is a spec 6a criterion
(`:911-949`), and each task above names the test that proves it.

1. Criteria 1 and 2 (the driver's adoption conditions): Task 2.
2. Criterion 3 (two handles, the busy wait and the busy error): Tasks 2
   and 3.
3. Criterion 4 (every DSN setting on every class): Task 2.
4. Criterion 5 (the read-only error; the orphan insert's 787): Tasks 2
   and 3.
5. Criteria 6 and 7 (the write method's commit rule, nesting, deadline,
   and version check): Task 3.
6. Criteria 8, 9, and 10 (the runner, the snapshots, the rebuild): Tasks
   4 and 5.
7. Criteria 11 and 12 (the invariants, the address index): Task 5.
8. Criteria 13 and 14 (not a store, the write-mode refusals, the lock
   conversion): Task 6.
9. Criterion 15 (inspect leaves the file unchanged; `doctor`'s goldens):
   Tasks 6 and 7.
10. Criterion 16 (the filesystem probe, the owner check, 0600): Tasks 1,
    4, and 6.
11. Nothing in the pass creates or migrates a store outside a test: the
    only production callers of `internal/engine/store` are `doctor` and
    `check`, both through the inspect mode.

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's
ownership table before the lane launches, beside the fold record's
"Revision 2.2" rows.

| Surface | Owner | 6a's position | Contention |
| --- | --- | --- | --- |
| `internal/engine/store` (new) | rung 6 | Tasks 1-6 | none; later rungs append migrations with fixtures |
| `go.mod`, `go.sum` | shared | Task 2 adds `modernc.org/sqlite` | 5c-i leaves both untouched; a later conflict resolves by `go mod tidy` and merge ritual step 5 |
| `cmd/dubplate-server/main.go` | rung 1 | Task 7's one registration line | **5c-i Task 7 adds its `AddCommand` line**; a textual conflict keeps both in landing order |
| `cmd/dubplate-server/doctor.go`, `doctor_test.go` | rung 1's skeleton, rung 5's rows | Task 7's store row and line-count constant | **5c-i Tasks 4, 5, and 6 edit both**; non-fast-forward rebase, merge ritual step 5 |
| `cmd/dubplate-server/check.go`, `check_test.go` | rung 1 | Task 7 | 5c-i does not edit them |
| `cmd/dubplate-server/store.go` (new) | rung 6 | Task 7 | none |
| `internal/config` | rung 1b's contract | **untouched** | 6c's condition for lane 2 |
| `internal/auth`, `internal/identity` | rung 5 | **untouched** | 6d only, after 5d |
| The state directory layout | rung 4 | the `store` leaf, created only by the serve mode | 6d grants it on the box |
| `docs/` ledgers | conductor | Task 8 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record the planning spend's actual against the rung;
   create the lane 2 worktree from master; write the lane registry row
   and the contested-surface rows; arm the runaway guard; check AC power.
2. Before Task 1: rule Task 0's blockers and record the four rulings'
   outcomes.
3. At segment A's boundary: STATUS; re-derive the ceiling from Tasks 1-3.
4. At segment B's boundary: STATUS; one `code-simplifier` round over
   Tasks 1-7's accepted Go, pasting `.claude/instructions/simplifier-brief.md`
   verbatim, its review chunked one `diff-reviewer` per package.
5. The merge ritual (merge token, rebase onto master, full gate in the
   worktree, `--ff-only`, full gate on master, and one `diff-reviewer` on
   the merged result, since 5c-i's edits to `doctor.go` and `main.go`
   make the rebase non-fast-forward), then **one `go-architecture-reader`
   per touched Go package** (`internal/engine/store`,
   `cmd/dubplate-server`), no plan in context.
6. Record the spec's 6a ceiling against the actual, and re-derive 6b-i's
   figure from it before 6b-i's plan is drafted.
7. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, in
**three segments**. Every task runs `dubplate-implementer`, then
`diff-reviewer` (`claude-opus-5-5`), then the gate, with one re-dispatch
on `fix` and a second `fix` as the conductor's decision.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1, 2, 3 | STATUS. Re-derive the ceiling. |
| **B** | 4, 5, 6, 7 | STATUS. The simplifier round. |
| **C** | 8 | The close drafts, the merge ritual, the architecture reads, the ledger fold. |

Each boundary sits on a task's own green committed gate, so each is a
resume point.

**Models.** `sonnet` by default. **Three upshifts to `opus`, all
unconditional**, dispatched as `claude-opus-5-5`: Task 3 (the commit rule
every later write inherits, whose wrong variants pass every happy-path
test), Task 4 (the migration runner and snapshot, the data-loss path the
risk model names first), and Task 6 (the lock-conversion procedure, whose
race a single-process test passes). Task 0 runs at `opus` in fresh
context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1.
- Task 1 → Task 2 (the leaf's database path).
- Task 2 → Task 3 (the handle and its classes).
- Task 3 → Task 4 (the write pool, the maintenance connection).
- Task 4 → Task 5 (the runner the invariant test drives).
- Task 5 → Task 6 (the embedded set's version).
- Task 6 → Task 7 (the exported inspect mode).
- Tasks 0-7 → Task 8.
- **Independent by files:** Task 1 and Task 2's dependency and adoption
  tests. They still run sequentially, because one tree runs one executor
  and the module gate runs over `./...`.

**Where this pass splits, if it must:** at the Task 5 / Task 6 boundary,
if segment A's re-derivation projects past the ceiling. Tasks 6 and 7
then move together to the head of 6b-i, so the inspect mode never ships
without its callers (R-6a2). Inside the pass, Task 6 carries a pre-named
task split: its serve mode moves to its own task before Task 7 if the
implementer reports that the four deliverables cannot land in one green
gate. That split stays inside segment B and adds about 0.53M, which
brings the projection to 5.45M, just past the 5.44M flag, so the
conductor asks at that boundary under the 80% rule ("Token ceiling").

**Budget:** 6.8M execution (Ruling 1), flag 5.44M, planning recorded
separately, STATUS at each boundary, spend scored at close against the
ceiling and against zero attended interactions after approval.
