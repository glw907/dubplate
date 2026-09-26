# Rung 6, Pass 6c (The Routed Schema and the Settings Home) Implementation Plan

Status: **draft for approval, revision 2, 2026-09-25.** Revision 1 (`8d0eed8`) went through a
three-lens `spec-plan-review` round (contract, mechanics, and risk; the reviews are committed at
`098f8e5` under `docs/superpowers/research/2026-09-25-rung-6c-plan-review-*.md`). Revision 2 folds
them; the fold record, one disposition per finding, is
`docs/superpowers/research/2026-09-25-rung-6c-plan-fold.md`, and it is this plan's review history.
Three rulings for Geoff ("Rulings awaiting owner"): the ceiling and lane placement, the exit code of
`doctor` and `check`, and the backup keys' protective minimums. Every answer is executable as
written.

**Ceiling:** 5.4M execution (flag 4.32M), planning recorded against the rung and outside it.
**Checkpoint interval:** each segment boundary, and the close. **Segments:** A (Tasks 0 to 3), B
(Tasks 4 and 5, the simplifier round, Task 6, the merge). **Tasks:** seven, Task 0 and the close
included, the count the spec's band table drew (spec `:1140`).

**Lane placement (Ruling 1, R-6c1): 6c can run tonight in lane 2, and it need not wait for 5d.** It
runs in a fresh lane 2 worktree cut from master after 6b-ii-b merges, on two conditions Task 0 checks:
no lane 1 edit under `internal/config` is uncommitted or in flight when the worktree is cut, and 5d
does not start executing until 6c merges. "Lane condition, answered" below carries the evidence.

Sources, in order of authority: the ratified rung 6 spec
(`docs/superpowers/specs/2026-09-23-rung-6-ops-seam-store.md`, revision 2.2 with errata E1 to E11),
above all its 6c pass section (`:1062-1084`), "The settings home" (`:718-800`), "Schema invariants"
(`:433-464`), "The schema rung 6 lands" (`:466-508`), "Package territory" (`:193-215`), "The 2.0
forward hooks" (`:877-893`), "Prior art" (`:895-917`), "Planning-miss prevention" (`:919-939`),
"Band and ceilings" (`:1121-1150`), and "Contested surfaces" (`:1152-1161`); the configuration homes
spec (`docs/superpowers/specs/2026-09-11-configuration-homes-design.md`, CH), its routed list
("What this spec routes elsewhere"), "The server's own knobs," and its seed paragraph (CH `:95-101`);
the 0.8 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, which still says
2.0), "Forward hooks the current build owes" (`:683-768`) and "Prior art" (`:770-806`); the repo
`CLAUDE.md`; the 6b-i plan's R-6bi8 (`docs/superpowers/plans/2026-09-24-rung-6b-i-ops-seam-backup.md:604-623`);
and `docs/HISTORY.md`'s 6a, 6b-i, and 6b-ii-a entries.

**Code pointers were resolved on master at `5b8f05e` and on branch `lane-2-6b-ii-b` at `30a66ac`**
(6b-ii-b's Task 5A; segment B2 was still running, with warm edits under `internal/backup`), and
re-checked for this revision on master at `098f8e5` and the same branch head. Every pointer into
`internal/backup` is cited by symbol, never by line, because 6b-ii-b is rewriting that package now:
at `30a66ac` the three backup constants sit at `prune.go:20`, `:24`, and `:28`, where master has them
at `:23`, `:27`, and `:31`. The branch also adds `Handle.Query` (`internal/engine/store/conn.go:115`
at `30a66ac`), a read-pool multi-row read this plan relies on (R-6c8). 6c executes only after 6b-ii-b
merges, and Task 0 re-resolves every pointer against the merged master.

**What this pass is.** It lands every table CH routes to rung 6 that 6a did not (spec criteria 1 to
3), gives the config registry the four pieces of per-key metadata and the correlated-rule list CH
asks for (criteria 4, 5, and 11), builds `internal/settings`, the accessor with cross-process
invalidation, the closed seed-rule table, and the audit (criteria 4 to 10), re-homes 6b-i's five
constants as the first store-homed keys, with `backup.Job` refusing any structurally invalid value
it is handed (criterion 13, R-6c13), and teaches `backup`, `check`, and `doctor` to evaluate the seed
rule read-only (criteria 8, 12, and 14). It also carries 6b-i's routed risk R13: the migration
runner's `SAVEPOINT` check, which turns a migration file carrying `COMMIT; BEGIN;` from silent into a
refusal.

**What this pass is not.** `serve` does not open the store, apply the seed rule, or read a setting:
that wiring, the seed rule's export, and the migration runner's production caller are 6d's (spec
`:213-216`, `:740-741`). No auth threshold moves to the store, and `authConfig` is untouched (6d's
seed handoff, spec `:773-800`). No admin write path, no mount, no re-authentication gate (rung 14).
No `Identity.UserID`, no roster import (6d). No share table or column (hook 15's 2026-09-25
amendment). No new module dependency. No `systemd/`, `scripts/`, `env/`, `web/`, `config/`, or
`internal/config/keys.go` change, and **no deploy of any kind.** No default value changes (R-6c11).
A task that believes it needs any of these stops and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the pass-execute chain per
> task, sequential, in the segments "Execution model" names. The implementer is
> `dubplate-implementer`. The conductor consumes structured reports and never reads a diff, a test
> log, or a gate transcript inline. Subagents never dispatch agents, so `code-simplifier` is the
> conductor's dispatch. The conductor owns `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the
> lane registry, and the contested-surface rows. The live-executor check is the lane registry,
> never `pgrep -f Projects/dubplate`.

---

## Rulings awaiting owner

Three questions. Everything else this plan decides is a question with one architecturally correct
answer, recorded with its reason under "Resolutions this plan takes."

**1. Does pass 6c run under a 5.4M execution ceiling, in lane 2 tonight once 6b-ii-b merges?**
Recommendation: yes. Reason: "Token ceiling" projects 4.33M from 6b-i's and 6b-ii-a's recorded
actuals plus the work the review fold added, and "Lane condition, answered" shows 5c leaves
`internal/config` alone for the rest of its plan while 5d cannot start before its plan is approved.

What each answer builds: yes sets the flag at 4.32M, the projection itself, so any overrun raises it
and the conductor asks one combined question before the merge-side reads; the pass runs tonight
after 6b-ii-b merges (and after the transliteration pass, if the conductor keeps that order in lane
2). No holds the spec-derived execution figure of 4.5M (flag 3.6M); cumulative spend with the fix
reserve is about 3.40M after Task 5 and crosses 3.6M during the simplifier round, so the conductor
asks after Task 5, before that round. "No" to the lane half means 6c waits for 5d's merge and runs
after it, with Task 0 re-reading 5d's `internal/config` and `doctor.go` edits.

**2. Do `doctor` and `check` exit nonzero when the seed rule would refuse `serve`'s start?**
The refusing outcomes are an admin row the env contradicts, a value in force after the seed that is
outside the binary's bounds (R-6c7), an env value or a stored value that does not parse under its
key's kind (R-6c5), and a correlated rule the post-seed set breaks. Recommendation: yes. Reason:
`doctor` and `check` both answer "can this deployment start," and 6d's `deploy.sh` runs the new
binary's `check` before restarting `serve` (spec `:395-396`), so a `check` that exits 0 on a refusal
lets a deploy restart a server that then will not start. The spec's reason for "without failing"
("so a timer-run backup never fails over an auth threshold it does not read," spec `:744-745`)
holds for `backup`, which never fails on the evaluation under either answer. A pending re-seed, an
unseeded env value, or a pending return to default is a disagreement, not a refusal, and never
fails any verb under either answer.

What each answer builds: yes makes `doctor` and `check` exit 1 on a refusing outcome, each naming
the key, both values or the bound, and the action that resolves it; this deviates from the spec's
"`check` included" sentence (spec `:743-745`) and owes it a dated erratum (D3), and Task 6's handoff
tells 6d that `deploy.sh` halts on `check`'s nonzero exit before restarting `serve`. No keeps the
spec's text: both verbs print the same refusal line and exit on today's rules alone, and 6d's
`deploy.sh` must parse the line to catch a refusal before restarting.

**3. Do the three backup keys take protective minimums: a snapshot keep of at least 3, an archive
retention of at least 168h, and a guard fraction of at least 0.25?** Recommendation: yes. Reason:
the floors are the only barrier against a one-typo seed at 6d and an admin write at rung 14, and each
revision 1 floor admits a value that destroys recovery material. A keep of 1 lets one run after an
upgrade delete every pre-upgrade remote snapshot, which is the rollback path of spec rulings 7 and
10. An archive retention of 24h purges within a day the files a bad sync moved into `archive/`. A
fraction such as `0.01` turns the short-tree guard off in effect. Each proposed floor still admits a
deliberately small deployment, and the defaults (14, `2160h`, `0.5`) clear all three. The ceilings
are harmless either way and do not change.

What each answer builds: yes sets R-6c11's bounds to keep 3 to 1000, archive retention 168h to
87600h, and guard fraction 0.25 to 1 (inclusive, NaN and both infinities refused); an env seed below
a floor refuses `serve`'s start at 6d, and `doctor` and `check` report it now (under Ruling 2). No
keeps revision 1's bounds: keep 1 to 1000, archive retention 24h to 87600h, guard fraction above 0
and at most 1, which needs the bounds data to carry an exclusive floor for the float kind. Either
answer leaves R-6c13's structural check in `internal/backup` (keep at least 1, fraction in (0, 1],
retention above 0) unchanged, since that check guards the code's own invariants, not the policy. A
split by channel (floors for admin writes only) is not offered: the registry cannot express a
per-channel bound at 6c.

---

## Lane condition, answered

The spec gives 6c "lane 2 worktree if 5c and 5d leave `internal/config` untouched; otherwise after
5d" (spec `:1062`, `:1173-1174`). The purpose is that two lanes never edit `internal/config`
concurrently. The evidence, read at `5b8f05e` and re-read at `098f8e5`:

- **5c's committed work never touched it.** `git log -- internal/config` shows its last commit is
  `6e810c4` (6b-i Task 6). No 5c-i, 5c-ii, or 5c-iii commit appears.
- **5c-iii and 5c-iv plan no edit to it.** The 5c-iii plan, which carries 5c-iv as Tasks 7 to 11,
  says "No edit to `internal/config`, `internal/auth`, `internal/identity`, `internal/engine`, or
  `cmd/dubplate-server/{check,doctor,store,backup}.go`" (`plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md:63-64`)
  and "It writes nothing under `internal/config`" (`:496`).
- **The in-flight review fold is the one open pressure, and it is live.** The security review's L2
  asks for a `DUBPLATE_SITE_NAME` grammar "at `internal/config/config.go:352`, `keys.go:23`"
  (`research/2026-09-25-rung-5c-iii-review-security.md:21-23`). At `098f8e5` the main tree carries
  warm, uncommitted edits to `internal/config/config.go` and `config_test.go` (a `requireSiteName`
  check), so L2 is being worked in lane 1 now. The review ties L2 to a 5c-iv task, so it may land
  either as a 5c-iii fold commit or inside 5c-iv. Task 0 blocker 3 reads L2's disposition in the
  5c-iii security fold record: if L2 is committed on master before the cut, 6c rebases onto it; if it
  is routed into 5c-iv, the conductor either lands it as a standalone commit before the cut or adds
  a contested row for `config.go` and holds 5c-iv's L2 edit until 6c merges.
- **5d cannot run tonight.** Its plan is drafted (`af7d941`, revision 1, pre-review) for Geoff's
  morning review (`docs/STATUS.md`, "Overnight 2026-09-25"). 5d's scope (R5S's pass table) adds the
  forward-auth and trusted-proxy integration classes, which under the rung 1b contract register from
  `cmd/dubplate-server`, and `doctor` probe rows in `doctor.go`, which 6c also edits. The conductor
  therefore holds 5d's launch until 6c merges, or 5d's plan walls off `internal/config`,
  `doctor.go`, and `check.go`; the contested-surface rows below record it.

So 6c runs before 5d under those two conditions, and the spec's "otherwise after 5d" branch applies
only if a lane 1 edit under `internal/config` is in flight at launch.

---

## Token ceiling: 5.4M execution (Ruling 1)

**Basis.** 6b-i (`docs/HISTORY.md`, its "Both budgets"): about 5.4M through the simplifier round,
with Task 0 at about 0.08M and the simplifier round at about 0.16M, over ten build tasks (1S, 1T,
1A, 1B, 2 to 7) and six fix rounds, so (5.40 − 0.08 − 0.16) / 10 ≈ **0.52M per task**, fix rounds
included; its merge ritual 0.08M and three architecture reads 0.24M, so **0.08M per read**. 6b-ii-a:
about 1.95M in all for Task 0 and three tasks, with its plan pricing the close at 0.99M, so
(1.95 − 0.99 − 0.11) / 3 ≈ **0.28M per task** with one fix round. 6c has no data-destructive path
of its own beyond Task 4's backup inputs and no security surface, so it prices a `sonnet` task at
0.45M to 0.60M and an `opus` task at 0.55M to 0.70M, with fix rounds reserved separately.

**What the review fold added, per task.** Task 1: the unique savepoint name, the kept commit branch,
and three state rows (+0.05M). Task 3: the post-seed bound split, the value and seed-value rule, five
state rows, the generation-first read, and the admin-source test (+0.05M). Task 4: `backup.Job`'s
structural refusal with its test helper (R-6c13), the injected readers, the discriminating per-key
fixtures, and the guard's message (+0.10M). Task 5: the facts entry point, the older-store states,
and criterion 14's WAL fixture (+0.05M). Task 2 is net neutral: bounds as data add a little, and
`Load` no longer parses settings and no raw reader is exported.

1. Task 0 (`opus`, fresh context): **0.10M**.
2. Task 1, the routed schema and the runner's `SAVEPOINT` check (`sonnet`): **0.50M**.
3. Task 2, the registry metadata (`opus`, the config contract): **0.55M**.
4. Task 3, `internal/settings` (`opus`, the novel logic): **0.70M**.
5. Task 4, the five constants re-homed (`sonnet`): **0.55M**.
6. Task 5, `backup`, `check`, and `doctor` (`sonnet`, the widest state table): **0.60M**.
7. Fix rounds beyond those rates: **0.40M**.
8. The close side: the simplifier round with its review (0.20M), Task 6 (0.15M), the merge ritual
   with a possible merged-range review (0.10M), and six architecture reads at 0.08M
   (`internal/engine/store`, `internal/config`, `internal/settings`, `internal/ops/local`,
   `internal/backup`, `cmd/dubplate-server`: 0.48M): **0.93M**.
9. **The projection is 4.33M.**

In two lines: 0.10 + 0.50 + 0.55 + 0.70 + 0.55 + 0.60 + 0.40 + 0.93 = 4.33M projected;
4.33M / 0.8 = 5.41M, so a **5.4M** ceiling with its 80% flag at **4.32M**.

The spec's own row was 5.6M including 1.07M of plan preparation and 10% for the config contract
(spec `:1140`); its execution-only equivalent is (0.2 + 7 × 0.46 + 0.6) × 1.1 ≈ 4.42M, which Ruling 1's
"no" rounds to 4.5M. Seven tasks is under ten, so the pass does not split.

Deliverable count: twenty-one across Tasks 1 to 6, plus Task 0's report. Tasks 1, 2, 3, and 5 carry
four each (16); Task 4 carries three, Task 6 two. No task passes four.

---

## Global constraints

- **A fresh git worktree in lane 2**, cut from master after 6b-ii-b merges. The conductor writes the
  lane registry row before Task 0. Every task, gate, and commit runs inside it. The worktree needs
  `npm ci --prefix web` before its first gate.
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`, and no task opens ssh to the
  box.
- **Package territory, stated as a wall.** This pass creates `internal/settings` (every file and its
  tests). It edits `internal/engine/store` (Task 1: two migrations, their fixtures, the runner's
  savepoint, the invariant test; Task 5: inspect's facts), `internal/config` (Task 2: the registry,
  a new declaration file, and `config.go`'s namespace classification), `internal/ops/local` and
  `internal/backup` (Task 4), and `cmd/dubplate-server`'s `store.go`, `backup.go`, `check.go`,
  `doctor.go`, and their tests (Tasks 4 and 5). **It writes nothing under `internal/auth`,
  `internal/identity`, `internal/serve`, `internal/atomicfile`, `internal/notify`, `internal/harness`,
  `internal/ops` (the seam package), `internal/ops/opshttp`, the rest of `internal/engine`, `web/`,
  `systemd/`, `scripts/`, `env/`, `config/`, `Makefile`, `CLAUDE.md`, `corpus/`, `tools/`, or
  `_salvage/`,** and no `internal/config/keys.go`, `serve.go`, `web.go`, `auth.go`, or `main.go`
  edit beyond what "Contested surfaces" names (none is expected: `registerStoreConfig` and
  `registerBackupConfig` already sit in `newProductionRegistry`). It adds no store read method:
  `Handle.Query` arrives with 6b-ii-b (Task 0 item 11). It writes no `docs/` file outside Task 6's
  drafts.
- **Imports follow the spec's table** (`:197-205`): `internal/settings` imports `config`, `store`, and
  the standard library; `internal/ops/local` takes no new internal import, because
  `cmd/dubplate-server` injects the settings it reads as reader functions (R-6c11, the spec's own
  pattern for 6d's auth keys, `:209-210`; the table's "`settings` from 6c" is a permission, not a
  requirement); `internal/backup` still imports only `store` and `atomicfile`; `internal/config`
  never imports `store` or `settings`; the store reads no setting's meaning (it may return raw rows,
  as its `Facts.LastSuccess` already returns raw job rows).
- **No new module dependency.** `go.mod` and `go.sum` stay byte-identical.
- **The store's rules carry over** (6b-i's global constraints): every store write goes through
  `Handle.Update`; times are integer UTC microseconds; every DSN comes from the one builder
  (`internal/engine/store/conn.go:187` on master); new code matches no raw SQLite result code outside
  the store's classifiers.
- **Nothing is exported ahead of a cross-package caller.** At the close every new export has a
  production caller outside its own package, with exactly the exceptions exit gate 13 names: the
  registry's correlated-rule entry point (the spec assigns it to 6c; its first production rule is 6d's
  nonce-retention rule), the `Sensitive` field's `true` value (6d declares the first sensitive keys),
  and the apply-mode type's "immediately" and "restart" values (every 6c key is "next operation";
  6d's auth keys use both). The seed rule's apply stays unexported until 6d's `serve` (spec
  `:215-216`).
- **No default changes.** Every re-homed key carries 6b-i's value unchanged (R-6c11). A task whose
  research argues for a different value reports it as a decision it did not make.
- **Every state a check meets names its report,** asserted by text, never by "an error occurred".
  Every proof names its fixture state (selected, populated, seeded) so it cannot pass vacuously.
- **Assertion discipline at the repo's bar.** No silent-success fakes, no vacuous assertions, no
  expectation derived from the code under test. Every hard-coded expectation cites the spec, CH, a
  quoted document, or a measurement this plan records.
- **`go-conventions` before any Go**, at the exemplary-stdlib bar. `ctx` first where a call can
  block. The em dash is banned in Go comments. No process citation in a shipped comment.
- **No `t.Parallel()` in `internal/engine/store`** (6a's umask rule).
- **The gate is `cairn-run-gate 'bash <worktree>/scripts/check.sh'`,** run with the worktree as its
  working directory and re-issued on exit 75 until it prints `gate exit:`. Every full gate is heavy
  (Chromium, about 3 to 4 minutes, about 560 MiB peak). Anything reported as evidence runs with
  `-count=1` and reports wall time. Specific files only. Commit footer
  `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>` for an `opus` task and the dispatch's
  footer otherwise.

---

## Behaviors this plan relies on, quoted or measured

From `go doc` at `go1.27.1 linux/amd64`, `modernc.org/sqlite v1.59.0` from the module cache, the
tree's own doc comments, and three probes run on 2026-09-25 in a scratch module against that driver
(the probe source is not committed; each task that relies on a measurement re-runs it as a test). The
mechanics review re-ran every probe and added two savepoint rows (`2026-09-25-rung-6c-plan-review-mechanics.md`,
minor 1).

- `go doc database/sql.DB.QueryRowContext`: "If the query selects no rows, the *Row.Scan will return
  ErrNoRows." The accessor's absent-row state is `sql.ErrNoRows` from `Scan`, never a nil error with
  zero values. `Handle.QueryRow` (`internal/engine/store/conn.go:108`) runs on the read pool, which
  opens with `_query_only=1` (spec `:320-322`). Its doc: "A read on the read pool sees the last
  committed state but holds no lock against a write that follows it, so it serves reports; a read
  that a write decides on runs inside that write's [Handle.Update] callback instead."
- `Handle.Query` (lane 2, `conn.go:115` at `30a66ac`): "Query runs query on h's read pool, as
  [Handle.QueryRow] does, and returns every row. The caller closes the rows, which hold one of the
  pool's connections until then." Each read-pool call is its own autocommit read, so two calls share
  no snapshot; R-6c8's generation-first rule is built on that.
- `go doc strconv.ParseFloat`: "ParseFloat recognizes the string "NaN", and the (possibly signed)
  strings "Inf" and "Infinity" as their respective special floating point values." Every comparison
  against NaN is false, so a bound written as `v < floor || v > ceiling` admits NaN. Every float
  validator refuses NaN and both infinities by name.
- `go doc time.ParseDuration`: "Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h"." No
  day unit exists, so a 90-day default is written `2160h`.
- `go doc strconv.Atoi`: "Atoi is equivalent to ParseInt(s, 10, 0), converted to type int." It
  accepts a sign, so every integer validator states its floor.
- `go doc errors.Join`: "Join returns nil if every value in errs is nil," and the result "may be
  inspected with Is and As." The seed evaluation joins every refusal so one run reports all of them.
- `PRAGMA data_version` "is only meaningful to compare ... returned by the same database connection
  at two different points in time" (sqlite.org/pragma.html, quoted at spec `:724-726`), and the read
  pool hands out different connections, so invalidation reads the settings generation row instead.
- `store.Inspect`'s doc (`inspect.go:157-165`): it "creates nothing, the lock file included, never
  migrates, and reads the database read-only, so the database file's bytes and mode are unchanged
  afterward; SQLite may create the -wal and -shm files beside a database that has none. It takes the
  store's lock shared and does not wait for it: a lock held exclusively reports StateBusy." And
  `Report.Facts` "is set for an older or current store whose version holds the jobs table, and nil
  otherwise: a newer store's schema is not this binary's to read."
- `OpenWrite` refuses an older store: `"store %s at version %d, binary at version %d: start serve to
  migrate"` (`open.go:175`), and `OpenServe` migrates before it returns, so no `*store.Handle`
  exists on a store with pending migrations (R-6c9's pre-seed read).
- The env lookup rule. `config`'s wall: "get returns name's value, treating an empty value as unset"
  (`config.go:183`). `doctor.go`'s `lookupEnviron` already carries that rule outside `config`
  ("a later entry winning over an earlier one for the same name, and treats an empty value as unset,
  [config]'s own wall.get rule," `doctor.go:397-399`), the precedent for `internal/settings` reading a
  store-homed key's env value itself (R-6c2).
- **Measured, the savepoint probe.** Inside an `_txlock=immediate` transaction, `SAVEPOINT runner`,
  then a body, then `RELEASE runner`:
  | Body | `RELEASE runner` | `Commit` | Tables after |
  | --- | --- | --- | --- |
  | `CREATE TABLE a(...)` | nil | nil | 1 |
  | `CREATE TABLE a(...); COMMIT; BEGIN; CREATE TABLE b(...)` | `no such savepoint: runner (1)` | nil | 2 |
  | `CREATE TABLE a(...); COMMIT;` | `no such savepoint: runner (1)` | `cannot commit - no transaction is active (1)` | 1 |
  | `CREATE TABLE a(...); ROLLBACK;` | `no such savepoint: runner (1)` | `cannot commit - no transaction is active (1)` | 0 |
  | a table plus `CREATE TRIGGER t AFTER INSERT ON a BEGIN SELECT 1; END;` | nil | nil | 1 |
  | `CREATE TABLE a(...); COMMIT; SAVEPOINT runner;` (mechanics review) | nil (it commits the body's transaction) | `cannot commit - no transaction is active` | 1 |
  | `CREATE TABLE a(...); COMMIT; BEGIN; SAVEPOINT runner;` (mechanics review) | nil | nil | 1, committed outside the runner |

  So the `RELEASE` failure catches the three self-terminating shapes that do not reopen the
  runner's own savepoint name, the silent `COMMIT; BEGIN;` included, and a trigger's `BEGIN ... END`
  body passes. A body that re-declares the savepoint's name defeats the check, which is why R-6c12
  names the savepoint something no migration writes and keeps `commitMigration`'s "no transaction is
  active" branch. The real runner rolls back after a failed `RELEASE` (`rollbackMigration`,
  `migrate.go:337`); for `COMMIT; BEGIN; CREATE TABLE b`, that rollback undoes `b`, `a` stays
  committed, and `user_version` is unchanged. For the `COMMIT;` shape the rollback joins a second
  error ("cannot rollback - no transaction is active"), so assertions match the file name and the
  `RELEASE` failure, never the whole text. The check detects; it does not undo the part committed
  before the file's own `COMMIT`, which the scratch store in the gate and the pre-migration copy on a
  real store cover.
- **Measured, the constraint probe** (`_foreign_keys=1`, `STRICT` tables), extended codes from
  `sqlite.Error.Code()`, whose doc reads "Code returns the sqlite result code for this error":
  a Go `string` bound into a `STRICT` `BLOB` path column answers 3091, "cannot store TEXT value in
  BLOB column" (the datatype check runs before 6a's path `CHECK`, so it answers 3091 for a valid path,
  `../x`, or empty alike); a `CHECK (visibility IN ('group','private'))` violation answers 275; a
  `NULL` into a `NOT NULL` `at` answers 1299; an orphan foreign key answers 787. The constants are
  `SQLITE_CONSTRAINT_DATATYPE = 3091`, `SQLITE_CONSTRAINT_CHECK = 275`,
  `SQLITE_CONSTRAINT_NOTNULL = 1299`, and `SQLITE_CONSTRAINT_FOREIGNKEY = 787` in the driver's
  generated `lib/sqlite.go` (`:3214-3224`).
- **Read in the tree.** 6a already declares `collections.visibility TEXT NOT NULL CHECK (visibility IN
  ('group', 'private'))` (`internal/engine/store/migrations/0003_catalog.sql`), so criterion 2 is a
  proof over an existing constraint, not a new column. Invariant 5 (`checkAttributeBags`,
  `invariant_test.go`) refuses any table that pairs a foreign key with a `name`, `key`, or `attr`
  column and a `value` column, so a per-member key-value table cannot pass the gate.
- **Not quoted, so declared rather than relied on.** Whether a `STRICT` table's non-integer
  `PRIMARY KEY` refuses `NULL` is not quoted here; every `BLOB` primary key in `0005` declares
  `NOT NULL` explicitly, and Task 1 proves it with a state row.
- **Navidrome's per-user settings, consulted for behavior only.** `model/user_props.go` declares
  `Put(ctx, userId, key, value)` and `Get`, and its table is `user_props (user_id ... references user
  on update cascade on delete cascade, key varchar not null, value varchar)`
  (`db/migrations/20210625223901_add_referential_integrity_to_user_props.go`, read 2026-09-25). That
  is the attribute bag invariant 5 forbids and the cascade invariant 7 forbids; see "Prior art."

---

## Privileged fields and their channels

| Field | Channels that may set it at 6c | What refuses every other channel |
| --- | --- | --- |
| A settings row's source `admin` | None in production. Tests write one directly through `Handle.Update` with literal SQL (spec `:765-766`). | No production function writes `admin`; the apply writes `'env-seed'` as a SQL literal and never binds `source` as a parameter (R-6c9). Exit gate 12 greps the pass's diff, and Task 3 asserts the `admin` row count unchanged after every apply path. |
| A settings row's `seed_value` | On an `env-seed` row, the apply, which writes it equal to `value` (R-6c7). On an `admin` row, rung 14's write path, which records the seed the admin value replaced; tests write it by literal SQL. | Nothing compares against it: every comparison reads `value` (R-6c7). |
| A settings row's source `env-seed`, and a row's removal back to the default | The seed apply alone, unexported at 6c, called only by `internal/settings`' own tests until 6d's `serve`. | Every other verb evaluates read-only (R-6c9); `backup`, `check`, and `doctor` hold no write path to `settings`. |
| A `settings_audit` row | The seed apply, in the same transaction as the row it records. | No other writer exists; a refused apply rolls back its audit rows with everything else (criterion 9). |
| The settings generation | The seed apply bumps it inside every transaction that changes a settings row. | Nothing else writes it; a test that writes a fixture row bumps it by literal SQL in the same transaction. |
| Every settings row, audit row, and the generation, wholesale | `restore` (6b-ii-b) replaces the database, so it installs the backup's `admin` rows, audit history, and generation with no audit row in the current timeline, and the generation can move backward. | Safe because restore holds the store's lock exclusively, which excludes every live handle and so every warm cache; a handle opened after the restore reads the restored generation fresh. Task 6's handoff carries the consequence to 6d. |
| A store-homed key's home, apply mode, sensitive marker, and bounds | The declaring rung's own file, at registration (`AddSetting` or its final name), in `cmd/dubplate-server`. | Never the env, never the store: the registry is code. A malformed declaration panics at registration (R-6c2). |
| A store-homed key's env value | Read only by `internal/settings`' evaluation, from the environ slice its caller passes (R-6c2). `doctor` renders the env value from the evaluation's outcome, never through `lookupEnviron`. | No function accepts a `Setting` and returns its env value; exit gate 11 greps each key's literal. |
| Rows in `member_overrides`, `member_preferences`, `member_quota`, `group_quota`, `device_credentials`, `move_intents` | None at 6c: each table's first writer is a later rung (R-6c6). | No production code writes them; tests populate fixtures only. |

---

## Resolutions this plan takes rather than discovers

Each resolution names its site. A task landing the surface a resolution decides writes nothing the
resolution does not settle; a question the plan left open is a stop-and-report, never a guess.

### R-6c1: the lane placement

As "Lane condition, answered" argues. The spec's condition is honored in substance: no two lanes
edit `internal/config` concurrently. It is recorded as spec reading D4 so a reader of the spec's
literal "5c and 5d leave `internal/config` untouched" finds why a pass that runs before 5d's plan
is approved satisfied it.

### R-6c2: a store-homed key is its own declaration family

Site: `internal/config`. A store-homed key is declared as a distinct type (working name `Setting`),
registered through its own entry point into a new registry family, never as a `Threshold` with a
home field. `Setting` does not embed `Threshold`, and no method or field yields one. The home is the
family: a key registered as a threshold stays on the wall and reads from the env exactly as today
(criterion 11's first half), and a key registered as a setting is store-homed.

What the compiler guarantees is exactly this: no `Parse*` function (`internal/config/thresholds.go:84`
onward) and no other function accepts a `Setting`. It cannot stop a hand-built `Threshold` literal
carrying a store-homed key's name, because `Parse*` accepts an unregistered `Threshold` (mechanics
review, minor 3). Exit gate 11 covers that residue: a `git grep` of each store-homed key's literal
shows it only in its `Setting` declaration and tests. The spec's "the compiler forbids a consumer
bypass" (`:735-736`) is read that way (D9).

A `Setting` carries its key, its kind (the existing `ThresholdKind`), its default in string form, its
accepted-form doc, its apply mode, its sensitive marker, and its bounds **as data**: a floor and a
ceiling in string form, parsed under the kind at registration, with NaN and both infinities always
refused for a float. Registration builds the value validator from that data, so the validator takes
a value and no environment by construction (spec `:730-733`), and every refusal and `doctor` row
prints the bound from the same data. (Under Ruling 3's "no," the float kind's floor is exclusive;
under "yes," every bound is inclusive.) The family claims across families like every other, so a
setting key that collides with a threshold, a wall row, a class key, or a superseded name panics
naming both (the `claim` rule, `registry.go`). Registration also panics when the default does not
parse under its kind, when a bound does not parse, when the floor exceeds the ceiling, when the
default falls outside its bounds, or when the apply mode is the zero value: each is a programmer's
mis-declaration, on the registry's existing panic precedent.

`internal/config` exports no reader of a setting's env value. The seed evaluation in
`internal/settings` reads it from the environ slice its caller passes, under the wall's rule (a later
entry wins; empty equals unset), on `doctor.go`'s `lookupEnviron` precedent.

### R-6c3: the metadata lives on settings only

Apply mode (immediately, next operation, restart), the sensitive marker, and the bounds land on
`Setting` alone. The fifteen wall thresholds keep their shape: their bounds stay in the sources'
`Validate` hooks until 6d moves the auth keys, and the serve and shadow keys stay on the wall at
restart (spec `:791-795`). Hook 18 (a description and a sensitive mark on every registered key) is
rung 14's and is not built here. Recorded with R-6c2's reading as D9.

### R-6c4: correlated rules are registry-level and proved with a test pair

Site: `internal/config` (declaration) and `internal/settings` (evaluation). The registry holds a
list of named correlated rules, each over two or more registered settings and a function of their
resolved post-seed values that returns an error naming every key it compares, on the
`sourceKeyError` precedent (`cmd/dubplate-server/auth.go`). A rule naming an unregistered or wall key
panics at registration. None of the five keys 6c re-homes correlates with another, and the one real
rule the spec names, nonce retention against link expiry, moves from the magic-link source's
`Validate` hook with 6d's re-home (spec `:787`). Criterion 5 is therefore proved in
`internal/settings` with two test-registered settings shaped as that rule (retention at least the
expiry), and 6d re-runs criterion 5 over the real pair (Task 6's handoff). The entry point is exit
gate 13's named exception.

### R-6c5: `Load` knows store-homed keys and never parses them

Site: `internal/config/config.go`. `rejectNamespace` (`config.go:243` at `5b8f05e`) classifies a set
store-homed key as declared, beside its threshold branch, so setting one is not "unknown
configuration key." `Load` does not parse a store-homed key's env value, exactly as it does not parse
a threshold today: "A registered threshold is a declared name whose own Parse* call resolves and
validates it at the point of use" (`rejectNamespace`'s threshold branch). A setting's point of use is
the seed rule, so an unparseable env value is a refusing outcome of the evaluation: `serve` refuses
at 6d, `doctor` and `check` report it (and fail under Ruling 2), and `backup` logs it and runs with
the stored value. The spec's "any | unparseable | a wall error, as today" row (`:763`) is read by
today's precedent, the parse error raised where the value is used (D7). An eager parse in `Load`
would fail every timer-run backup over an env typo in a key backup never applies, the failure the
spec's "without failing" sentence exists to prevent (`:744-745`).

### R-6c6: the routed schema's shape

Site: `internal/engine/store/migrations/`. Two migrations, numbered after the highest on master at
Task 0 (today `0003_catalog.sql`, so `0004` and `0005`). Table names are working names (spec
`:472-473`); the contents below are the contract.

- **The settings home** (`0004`): `settings`, keyed by `key` (TEXT primary key, `NOT NULL`), with the
  value (TEXT), the source (`CHECK (source IN ('env-seed', 'admin'))`, since a `default` source
  stores no row), the seed value, the actor, and `updated_at`; `settings_audit` with an opaque `id`,
  `at` and `kind` both `NOT NULL`, the actor, the channel, the key, and the old and new values
  (either nullable, for a seed from nothing and a removal); and `settings_generation`, one row, whose
  primary key is a column not named `id` or `*_id` with `CHECK` pinning it to one value, holding an
  integer generation the migration inserts at 0. `kind` and `channel` carry no `CHECK`, on
  `jobs.kind`'s precedent: rung 14 adds values, and a `STRICT` table's `CHECK` changes only by a
  rebuild.
- **CH's per-member, credential, quota, and journal tables** (`0005`):
  - `member_overrides`, keyed by `user_id` (a `BLOB NOT NULL` primary key referencing `users`), with
    the four overrides CH names (CH, "Two surfaces, one API"): a quota in bytes, may-upload,
    may-create-collections, and an upload concurrency cap, each nullable (NULL means no override)
    with a `CHECK` on its domain (non-negative bytes, 0 or 1 flags, a cap of at least 1), and
    `updated_at`.
  - `member_preferences`, keyed by `user_id` the same way, with `collapse_duplicates` (0 or 1,
    default 0), the one preference CH names for the rung 6 schema ("The keys are named now so the
    rung 6 schema carries them," CH, "Rules that cut across the verbs"), and `updated_at`. Theme,
    language, ReplayGain mode, sort order, notification choices, and the default collection each
    migrate in with the rung that first reads them, under CLAUDE.md's 2026-09-12 refinement and U4
    ("Adding a built-in is a migration and a typed accessor in the built-in's own rung," spec
    `:527`), the same reasoning the spec's ruling 2 applied to the S20 columns.
  - `member_quota`, keyed by `user_id` the same way, with the charged bytes (non-negative) and
    `updated_at`; and `group_quota`, one row on `settings_generation`'s singleton pattern, with the
    physical total. Reservations and true-up are rung 7's first consumer's to add.
  - `device_credentials`, an opaque `id`, the credential hash (`BLOB NOT NULL UNIQUE`), `user_id`
    referencing `users`, a server-minted device id (`BLOB NOT NULL UNIQUE`), `created_at`,
    `last_seen_address` (nullable TEXT), `expires_at`, and `revoked_at` (nullable) (CH, "The writer
    is a device credential").
  - `move_intents`, an opaque `id`, `at` and `kind` both `NOT NULL`, `old_path` and `new_path` as
    `BLOB NOT NULL` under 6a's path `CHECK` (relative, no empty, `.`, or `..` component, no NUL;
    E5), `granularity` with `CHECK (granularity IN ('file', 'directory'))`, and `state` with
    `CHECK (state IN ('pending', 'done'))`, the replay rule "source absent and target present means
    done" (CH, "The verbs") being the consumer's.
  - No foreign key declares `ON DELETE CASCADE` (invariant 7; a member's rows survive a tombstone).
    No share table or column (hook 15's amendment).
- **Invariant 6's exempt list** gains `settings`, `settings_generation`, and `group_quota`, the three
  the spec names beside the ledger (`:455-458`), and **also `member_overrides`,
  `member_preferences`, and `member_quota`** (spec deviation D1). Each is a one-to-one extension of
  `users` keyed by the parent's `BLOB` id, the canonical relational shape; an extra opaque `id` on
  each would be a second identity for one row that nothing references. The invariant's purpose,
  that no identifier is a rowid (S20:166-167), holds, since a `BLOB` primary key in a `STRICT`
  table is never a rowid alias and the "no `INTEGER PRIMARY KEY` named `id` or `*_id`" half still
  applies to every table (`checkEntityIDs`, `invariant_test.go:391-412`).

Each migration ships a populated fixture under `testdata/fixtures/` (R-6a15), populated as of its
version: a user, a collection, one row in every new table, and, for `0004`, an `env-seed` row, an
`admin` row whose `value` and `seed_value` differ, two audit rows, and a generation above 0.

### R-6c7: what a settings row means, and what the seed rule compares

No row means the source is `default`, and reads resolve the registry's current default, so a
re-researched default reaches every store (spec `:746-748`, criterion 7). A row stores the value as
the text it was seeded or written with.

- **Which column is compared.** An `env-seed` row holds `value = seed_value`; the apply writes both.
  On an `admin` row, `value` is the admin's value and `seed_value` the seed it replaced (rung 14's
  write path sets it). Every comparison the seed rule makes reads `value`: CH's "If the row is
  `admin` and the env value differs, the load fails" (CH `:100-101`) compares the value in force.
- **Typed equality.** Two values are equal when they parse to the same typed value under the key's
  kind, so an env value of `60m` against a stored `1h` is "equal to the seed" and changes nothing;
  the stored text is not rewritten.
- **Bounds apply to the value in force after the seed.** The spec's seed rule "validates each key's
  bounds and every correlated rule against the post-seed set" (`:740-742`), so the closed table's
  "any | stored value outside the binary's bounds" row refuses only when the out-of-bounds stored
  value stays in force: an `admin` row, or an `env-seed` row whose env value is equal to it. An
  out-of-bounds `env-seed` row that an in-bounds env value re-seeds, or that an absent env removes,
  is healed and audited. Otherwise an upgrade that narrows a bound would refuse `serve`'s start with
  no env value able to clear it (risk review M2). Recorded as D8.
- **Parse failures refuse.** An env value or a stored value that does not parse under its kind is a
  refusing outcome naming the key and the reason (R-6c5).
- The actor of a seed is `system` and its channel `env-seed` (spec `:747`).

### R-6c8: the accessor

Site: `internal/settings`. An accessor is bound to one `*store.Handle` and the registry, is safe for
concurrent use, and reads a setting by its declaration, one typed read per kind. Each read first
reads the generation on the read pool (`Handle.QueryRow`, one indexed read); an unchanged generation
serves the cached value, and a changed one reloads every row through `Handle.Query`.

**The generation is read before the rows, and the cache is stored under that earlier generation.**
Each read-pool call is its own autocommit read with no shared snapshot, so a write can land between
the two reads. Reading the generation first means a racing write can only leave newer rows cached
under an older generation, which the next read sees as changed and reloads; it can never cache older
rows under a newer generation, which would serve a stale value until some later write. Criterion 10's
sequential test cannot catch the wrong order, so the order is a Task 3 diff-reviewer checklist line.

A missing row resolves the current default. A stored value that fails to parse or falls outside the
binary's bounds is an error naming the key and the bound, never a silent default. A generation read
that fails, including a missing generation row (`sql.ErrNoRows` from `Scan`), is an error naming
`settings_generation`, never read as 0; the cache is never trusted without a successful generation
read.

A consumer reads a setting before it enters `Handle.Update`, not inside the callback. The spec's
rule that "every read a write decides on runs inside the callback" (`:355`) governs rows the write
changes. A setting read a moment before the write is the value in force for that operation, which is
exactly the apply mode "next operation" the five keys declare (R-6c11). A "restart"-mode key served
through a reloading accessor would apply live; no 6c key uses that mode, and Task 6 hands the rule to
6d.

### R-6c9: evaluation is exported, the apply is not

Site: `internal/settings`. One unexported core takes the registry, the environ slice, and the stored
rows in `internal/settings`' own row type, and returns one outcome per registered setting from the
closed table (spec `:751-763`, read per R-6c7), the correlated-rule refusals, and the orphan rows,
without writing. Each outcome carries the key, the stored row (if any), the env value (raw, if set),
the value in force, whether it refuses, and the report text; `doctor` renders the env value from the
outcome, never through its own `lookupEnviron`. Refusals join with `errors.Join`.

Two exported entry points share that core, each with its first production caller in Task 5:

- **From a handle**: reads every row through `Handle.Query` (the read pool, never an `Update`
  transaction) and evaluates. Its caller is `backup`, from its write-mode handle.
- **From facts**: converts `store.Facts`' raw settings rows (R-6c10) into the core's row type and
  evaluates; when the facts mark the settings table absent (a store older than `0004`), every key
  resolves to its default and each outcome says the store holds no settings yet. Its callers are
  `check` and `doctor`.

The row type the store exports lives in `store` (it cannot import `settings`); the conversion lives
in `internal/settings`, which imports `store`.

**The pre-seed read is the facts entry point's value in force.** The migration runner's local
snapshot cap reads "the stored row or the registry default" (spec `:748-749`) before the migration
it caps, and no `*store.Handle`, so no accessor, exists on a store with pending migrations (`open.go`
refuses it in write mode; `OpenServe` migrates before returning). `Inspect` needs no handle and
returns the raw rows for any older store whose version holds the settings table, so 6d's `serve`
inspects, resolves `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` through the facts entry point, and passes the
value as `ServeOptions.Keep`. At 6c the key's consumer is `doctor`'s row (D10).

**The apply** runs inside one `IMMEDIATE` transaction through `Handle.Update`, **reads every settings
row through `tx.Query` inside the callback** (spec `:355`; a read-pool read followed by the write
would let a rung 14 admin write land between them and be overwritten), evaluates, and writes every
outcome's change, its audit rows, and one generation bump, or refuses and writes nothing. It writes
`source` as the SQL literal `'env-seed'`, never as a bound parameter. It stays unexported until 6d's
`serve` exports it.

### R-6c10: inspect mode returns the raw settings rows

Site: `internal/engine/store/inspect.go`. `Facts` (`:126` at `5b8f05e`) gains the settings rows as
stored (key, value, source, seed value, actor, time), the generation, and an explicit marker that
the store's version predates the settings table, so "table absent" is never confused with "table
empty." A `settingsVersion` constant sits beside `jobsVersion` (`:121`) on its precedent. `readFacts`
(`:223`) reads the settings rows and generation on the inspect DSN through one connection it shares
with the jobs read (today `lastSuccesses` opens and closes its own, `inspect.go:262-271`; the task
moves that open into `readFacts` so both reads use it). Facts carry settings rows for an older store
whose version holds the table as well as for a current one. The store interprets nothing:
`internal/settings` evaluates. So `check` and `doctor` stay read-only (criterion 14). The fields are
exported with Task 5, their first caller.

### R-6c11: the five keys, their homes, defaults, and bounds

Site: `cmd/dubplate-server/store.go` (`DUBPLATE_JOB_RETENTION`, `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`)
and `cmd/dubplate-server/backup.go` (the three backup keys), registered by the existing
`registerStoreConfig` and `registerBackupConfig`, never in `internal/config/keys.go`. **The defaults
are 6b-i's values, carried unchanged** (R-6bi8), each definition's comment citing that source (BS's
researched-default rule, spec `:768-771`). If Task 4's research argues for a different value, the
task reports it as a decision it did not make and the conductor routes it to a ruling; the defaults
for the snapshot keep and the archive retention decide how much remote history a run deletes, which
is irreversible. The only new default is the local keep's, which R-6bi8 never set.

| Key | Kind | Default | Bounds (Ruling 3 yes / no) | Apply mode | Consumer at 6c |
| --- | --- | --- | --- | --- | --- |
| `DUBPLATE_JOB_RETENTION` | duration | `2160h` (90 days) | 24h to 87600h (either answer) | next operation | `ops/local`'s prune at each job start (replacing `jobs.go`'s `retention`) |
| `DUBPLATE_BACKUP_GUARD_FRACTION` | float | `0.5` | yes: 0.25 to 1; no: above 0, at most 1; NaN and infinities refused either way | next operation | each backup run's guard (replacing `guard.go`'s `guardFraction`) |
| `DUBPLATE_BACKUP_SNAPSHOT_KEEP` | int | `14` | yes: 3 to 1000; no: 1 to 1000 (the latest is never pruned, spec ruling 10) | next operation | each run's prune (replacing `prune.go`'s `snapshotKeep`; `othersKeep` derives from it) |
| `DUBPLATE_BACKUP_ARCHIVE_RETENTION` | duration | `2160h` | yes: 168h to 87600h; no: 24h to 87600h | next operation | each run's prune (replacing `archiveRetention`) |
| `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` | int | `3` | 1 to 100 (`ServeOptions.Keep` "must be at least 1," `open.go:92-94`) | next operation (the next migration) | `doctor`'s row, through the facts entry point; 6d's `serve` passes it as `ServeOptions.Keep` |

6b-i's R-6bi8 said 6d declares `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` "and 6c re-homes it after,"
written when 6c was expected after 6d. 6c now runs first, so 6c declares it as a store-homed key
whose first production reader is `doctor`, which answers R-6bi8's `unused` concern; 6d wires it into
`serve`. Its default of 3 is a working value: a rollback needs the newest copy, and two older copies
cover a bad upgrade found a release late, each a full store whose size `doctor` already shows; Task
4 records a source or keeps 3 with this reason.

**How the values reach their consumers.** The declarations live in package `main`, which
`internal/ops/local` cannot import. So `cmd/dubplate-server`'s `runBackup` builds the accessor and
injects reader functions in `local.Config`, one per value `ops/local` needs, each a closure over the
accessor and its declaration (the spec's pattern for 6d's auth keys, `:209-210`). `ops/local`
imports no `settings`, and its tests pass plain functions. `start` (or `StartBackup` for the three
backup values) calls every reader it needs **before** `insert` commits the queued record, so a reader
error fails the start naming the key and leaves no job record; the binder receives the three backup
values as arguments and returns no error, as today (`jobs.go:37`). `internal/backup` still imports no
`settings`: its three values become fields of `backup.Job`, which is one run's inputs (`backup.go`,
"A Job is one backup run's inputs").

### R-6c12: the migration runner's savepoint

Site: `internal/engine/store/migrate.go`. `applyInTx` (`:298`) runs each file's body between
`SAVEPOINT` and `RELEASE` under a name no migration writes (working name `dubplate_runner`), so a body
that ends its own transaction fails at `RELEASE`, naming the file, before the version write and the
ledger row. It covers the rebuild path too, which calls the same function. `commitMigration`'s "no
transaction is active" branch (`:355`) **stays**, and its comment cites the probe row it still
catches: a body that ends its transaction and re-declares the savepoint's name passes `RELEASE`, and
the branch is then the only detection. The runner does not scan file bodies for the savepoint name:
the unique name removes an accidental collision, and a deliberate evasion inside a reviewed in-repo
migration is the reviewer's to catch.

### R-6c13: `backup.Job` refuses a structurally invalid value and never defaults one

Site: `internal/backup`. Once the three constants become `backup.Job` fields, a field left at its
zero value would silently change the run: a guard fraction of 0 turns the short-tree guard off
(`float64(files) < 0*float64(tracks)` is never true, `guard.go:86`); an archive retention of 0 purges
every archive directory but the run's own (`own.Sub(stamp) <= 0`, `prune.go:154`); a keep of 0 makes
`othersKeep` −1 and panics at `others[othersKeep:]` (`prune.go:123`). So `Job.Run` validates the
three fields **before any rclone launch**, the `LookPath` included, and fails with a named failure
code (working name `job-invalid`) naming the field: keep at least 1, fraction in (0, 1] and not NaN,
archive retention above 0. `internal/backup` never supplies a default: a second copy of the registry
default on the data-loss path is exactly what the registry exists to prevent. This check is the
code's own invariant; the policy bounds are the registry's (R-6c11, Ruling 3). The guard's refusal
text names the fraction in use instead of "fewer than half" (`guard.go:87`, asserted at
`job_test.go:547`).

---

## Spec deviations and readings that take effect on approval

| # | Spec site | Change | Why |
| --- | --- | --- | --- |
| D1 | `:455-458`, invariant 6's exempt list | adds `member_overrides`, `member_preferences`, `member_quota` | R-6c6 |
| D2 | `:495-496`, `member_overrides` and `member_preferences` | the columns are CH's named four and `collapse_duplicates`; other preferences arrive with their consumers | R-6c6 |
| D3 | `:743-745`, "`check` included" | on Ruling 2's yes, `doctor` and `check` fail on a refusing outcome; a disagreement still fails nothing | Ruling 2 |
| D4 | `:1062`, `:1173-1174`, the lane condition | read as "no concurrent edit to `internal/config`," satisfied by sequencing | R-6c1 |
| D5 | 6b-i plan R-6bi8 | 6c declares `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`, read at 6c by `doctor` through the facts entry point; 6d wires it into `serve` | R-6c9, R-6c11 |
| D6 | criterion 5's "the correlated retention rule" | proved with a test pair shaped as the rule; the real rule moves at 6d, which re-runs criterion 5 over the real pair | R-6c4 |
| D7 | `:763`, "any \| unparseable \| a wall error, as today" | `Load` classifies a store-homed key without parsing it; an unparseable env or stored value is a refusing outcome of the seed evaluation, raised where the value is used | R-6c5 |
| D8 | `:761`, "any \| stored value outside the binary's bounds" | the bound applies to the value in force after the seed; an out-of-bounds `env-seed` row the env re-seeds or removes is healed and audited | R-6c7 |
| D9 | `:734-736` and criterion 11, "the compiler forbids a consumer bypass" | the compiler forbids passing a `Setting` to any function that reads the env; a hand-built `Threshold` carrying the key is caught by exit gate 11's grep; wall keys carry no apply mode or marker until hook 18 | R-6c2, R-6c3 |
| D10 | criterion 13, "each of 6b-i's five constants reads through the accessor" | the local keep's pre-seed read resolves from `Inspect`'s raw rows through the facts entry point, since no handle exists before the migration it caps; its `OpenServe` prune proof moves to 6d, the first caller | R-6c9 |

Task 6 drafts each as a paste-ready dated erratum.

---

## Review focus

The input classes most likely to bite, each pinned by a test in the task that owns it.

1. **A settings write in one process and a cached read in another.** Expected: the next read sees
   the write. Pinned in Task 3 (criterion 10); the generation-first order is a reviewer line.
2. **A seed set that is valid only together** (both correlated keys rise in one deploy). Expected:
   accepted; one key alone past the stored other refused, nothing written. Pinned in Task 3.
3. **A refused seed** (admin conflict, a bound, a parse failure, a correlated rule). Expected: no
   row, audit row, or generation change. Pinned in Task 3 (criterion 9).
4. **A float bound fed `NaN` or `Inf`.** Expected: refused by name. Pinned in Tasks 2 and 4.
5. **A verb that must not write** (`backup`, `check`, `doctor`) meeting an env that disagrees with
   the store. Expected: rows and file bytes unchanged, the disagreement reported. Pinned in Task 5.
6. **A migration body that commits itself.** Expected: refused naming the file, no ledger row.
   Pinned in Task 1.
7. **A store-homed key read through the wall.** Expected: no function accepts a `Setting`; no
   literal of the key outside its declaration. Checked by the diff-reviewer in Task 2 and by exit
   gate 11.
8. **A `backup.Job` with a zero or invalid retention, keep, or fraction.** Expected: refused naming
   the field before rclone runs. Pinned in Task 4.
9. **A narrowed bound over an old `env-seed` row.** Expected: an in-bounds env heals it. Pinned in
   Task 3.

What a column means stays a reviewer's question (spec `:462-464`): every migration diff carries the
checklist line "a path column not named as one, or a source identifier under another name, passes
the invariant test; read each new column's meaning."

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched by the conductor after the worktree is cut. **Its output is
blockers with evidence, not fixes.** It writes no code and edits no file. It grades a moved pointer
by content: content that survives at any offset or under a new name is a note naming the new site,
and content that is gone is a blocker.

**Blockers: a "no" stops the launch.**

1. **Has 6b-ii-b merged, and is the worktree fresh from master?** Report `git worktree list`, the
   lane registry rows, and `git status` in both trees.
2. **Do the surfaces 6b-ii-b changed still carry what this plan re-homes?** Report, by symbol, at
   HEAD: `internal/backup`'s `snapshotKeep`, `othersKeep`, `archiveRetention`, and `guardFraction`
   and every reader of each (6b-ii-b's Tasks 5A to 6 rewrote `prune.go`, `rclone.go`, and
   `backup.go` and added `fetch.go`; report whether restore or its fetch now reads any of the four);
   `backup.Job`'s fields and `Job.Run`'s order up to its first rclone launch; `guard.go`'s
   "fewer than half" text and its assertion; `ops/local`'s `retention` and its reader, `start`,
   `insert`, the `binder` type, and `Runner.backupJob`; `store.ServeOptions`, whose `Keep` field must
   survive beside 6b-ii-b's new `Reconcile`; whether `cmd/dubplate-server/restore.go` passes a `Keep`
   or any of the five values.
3. **Is `internal/config` quiet?** `git log <this plan's commit>..HEAD -- internal/config` and the
   main tree's `git status`: any lane 1 commit there is a note naming what changed; any uncommitted
   edit under `internal/config` in the main tree is a blocker (R-6c1; at `098f8e5` a
   `requireSiteName` edit was warm there). Read the 5c-iii security fold record's disposition of L2:
   if it is routed into 5c-iv and not yet committed, report it so the conductor lands it before the
   cut or adds the contested row and holds 5c-iv's edit. Report the lane registry's lane 1 row.
4. **Has 5d started?** If a 5d plan is approved and its lane is open, report whether its plan walls
   off `internal/config`, `doctor.go`, and `check.go`; a 5d executor writing any of them is a
   blocker (R-6c1).
5. **What is the highest migration on master?** 6b-ii-b's segments at `30a66ac` added none; report
   `ls internal/engine/store/migrations/` so Task 1 numbers from it.
6. **Does no `internal/settings` package, no settings table, and no store-homed declaration exist?**
7. **Did `Handle.Query` merge with 6b-ii-b?** Report its signature and doc at HEAD. R-6c8 and R-6c9
   read every row through it; its absence is a blocker, since this pass adds no store read method.

**Notes, not blockers:**

8. Report 6b-ii-b's Task 8 handoff "to 6c" items (its draft or the folded HISTORY entry) and route
   each to the task it touches, or report that it names none.
9. Report at HEAD: `newProductionRegistry`'s registration calls; `registerStoreConfig` and
   `registerBackupConfig`; `newDoctorCmd`'s parameters (it takes the registry) and `doctor.go`'s row
   order (the store row, then the backup rows) and its line-count tests' `wantTotal` expressions;
   `check.go`'s store row; `invariant6Exempt`; `applyInTx`, `rollbackMigration`, and
   `commitMigration`; `Inspect`, `Facts`, `readFacts`, `lastSuccesses`, `jobsVersion`, and the eight
   `store.State` values; `Handle.QueryRow`; `rejectNamespace`, `thresholdOf`, `AddThreshold`, and
   `claim`.
10. Report 6b-ii-b's actual spend from its HISTORY entry; the conductor re-derives this plan's
    ceiling from it before Task 1 (spec `:1147-1149`).
11. Report the conductor's recorded outcome of Rulings 1, 2, and 3.

**Deliverable count: 1** (the report).

---

## Task 1: The routed schema and the runner's savepoint

Model: `sonnet`. Depends on Task 0. **Lands hook 15's routed tables and hook 4 on
`settings_audit` and `move_intents`** (both carry `at` and `kind`); consumes hook 5 (every per-member
row keys on the surrogate id). Hardens hook 7 (R-6c12). Hook 20 weighed: neither event table records
a serve event, so neutral.

**Files:**
- Create: `internal/engine/store/migrations/0004_settings.sql`, `0005_routed.sql` (numbers from Task
  0 item 5), and their fixtures under `internal/engine/store/testdata/fixtures/`
- Create: violation fixtures under `testdata/violations/` as the criteria need
- Modify: `internal/engine/store/migrate.go` (R-6c12) and `migrate_test.go`
- Modify: `internal/engine/store/invariant_test.go` (the exempt list, D1) and `schema_test.go`

**Outcomes:**

- **The two migrations**, exactly R-6c6's contents, each with its populated fixture.
- **The exempt list** as R-6c6 and D1 state, with a comment naming each entry's class (keyed or
  singleton).
- **The savepoint** (R-6c12), under its unique name, with `commitMigration`'s branch kept and its
  comment citing the probe row it still catches.
- **The constraint proofs** for criteria 1 to 3.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Invariant test | every migration applied to an empty database | no violation |
| Invariant test | each migration's populated fixture migrated to head | no violation |
| Invariant test | a migration with no fixture | the subtest fails naming the migration (existing) |
| `move_intents` insert | a Go `string` in `old_path` or `new_path` | refused, extended code 3091 |
| `member_overrides` insert | a `user_id` with no `users` row | refused, extended code 787 |
| `member_overrides` insert | a `NULL` `user_id` | refused, extended code 1299 |
| `collections` insert | `visibility` `public`, and `''` | refused, extended code 275, each |
| `collections` insert | `group`, `private` | accepted, each |
| `settings_audit`, `move_intents` insert | `at` NULL, `kind` NULL | refused, extended code 1299, each of the four |
| `settings` insert | source `default` | refused, extended code 275 |
| `settings_generation` insert | a second row | refused, extended code 275 |
| Migration apply | a synthetic file with `COMMIT; BEGIN;` | refused naming the file and the `RELEASE` failure; no ledger row for it; `user_version` unchanged |
| Migration apply | a synthetic file ending `COMMIT;`, and one ending `ROLLBACK;` | refused naming the file, each (asserted on the file name and the `RELEASE` failure, since the rollback may join a second error) |
| Migration apply | a synthetic file ending `COMMIT; SAVEPOINT <the runner's name>;` | refused by `commitMigration`'s branch, naming the file |
| Migration apply | a synthetic file with a `CREATE TRIGGER ... BEGIN ... END` body | applied, ledger row written |

**Acceptance criteria:**

- Spec criterion 1: the invariant test passes over the full schema with a fixture per migration
  (five), and the two refusal rows above pass. **Fixture state:** a current store with a user row, so
  the orphan row's only fault is the missing parent; a `move_intents` row otherwise valid, so the
  `string` bind is the only fault.
- Spec criteria 2 and 3: the rows above, each by its own subtest, the `NOT NULL` half also asserted
  from `pragma_table_info` (`notnull = 1`).
- Each savepoint row asserts the store's `user_version` and ledger after the refusal, never a non-nil
  error alone.
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Drop `STRICT` from `move_intents` → the `string` bind row and invariant 1 both fail.
2. Remove `member_overrides` from the exempt list → the invariant test fails naming it (proves D1's
   entry is load-bearing).
3. Remove the `RELEASE` → the `COMMIT; BEGIN;` row passes silently, so its test fails.
4. Drop `NOT NULL` from `settings_audit.at` → invariant 3 and the insert row both fail.

**Prior art:** `modernc.org/sqlite` (adopted, spec); `pressly/goose` and `golang-migrate` (declined,
spec); Navidrome's `user_props` (consulted, declined, R-6c6 and "Prior art").

**Deliverable count: 4** (the settings migration, the routed migration, the exempt list with the
constraint proofs, the savepoint).

---

## Task 2: The registry's store-homed family

Model: `opus`. Depends on Task 0; **independent by files of Task 1.** **Lands hook 15's metadata half**
(apply mode and the sensitive marker on store-homed keys); hook 18 is not built (R-6c3).

**Files:**
- Create: `internal/config/settings.go` (the declaration type, its bounds data and the validator
  registration builds from them, the apply-mode type, the correlated rule type) and `settings_test.go`
- Modify: `internal/config/registry.go` (the family, its entry points, the rule list, the family
  name constant) and `registry_test.go`
- Modify: `internal/config/config.go` (`rejectNamespace`, R-6c5) and `config_test.go`

**Outcomes:**

- **The declaration family** (R-6c2) with its registration panics and its accessor returning the
  settings in declaration order, on `Thresholds()`'s precedent; bounds as data, the validator built
  at registration.
- **The apply-mode type** with its three named values and a zero value that registration refuses;
  its `String` names each (on `ThresholdKind.String`'s precedent).
- **The correlated-rule list** (R-6c4) with its registration panics.
- **`Load`'s classification** (R-6c5): a set store-homed key is declared, and `Load` never parses it.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Registration | a well-formed setting | registered; listed in order |
| Registration | a key already a threshold, wall row, class key, or superseded name | panic naming both families |
| Registration | a default that does not parse under its kind | panic naming the key and the default |
| Registration | a floor or ceiling that does not parse, or a floor above the ceiling | panic naming the key and the bound |
| Registration | a default outside its own bounds | panic naming the key and the bound |
| Registration | a zero apply mode | panic naming the key and the missing field |
| Registration | a key without the `DUBPLATE_` prefix | panic (the existing prefix rule) |
| Rule registration | a rule over registered settings | registered |
| Rule registration | a rule naming an unregistered key, or a threshold | panic naming the key |
| `Load` | a store-homed key unset or empty | no error |
| `Load` | set, parsing or not, in bounds or not | no error; not "unknown configuration key" (the evaluation reports a parse failure, R-6c5) |
| Validator | a value at the floor, at the ceiling, and one past each | accepted, accepted, refused naming the bound, refused naming the bound |
| Validator | a float setting given `NaN`, `Inf`, `-Inf` | refused, each by name (test-registered float) |

**Acceptance criteria:**

- Spec criterion 11: a threshold registered with no home still resolves from the env through
  `ParseDuration` (an existing threshold, asserted unchanged). The store-homed half: the
  diff-reviewer confirms `Setting` does not embed `Threshold`, no method or field yields one, and no
  function accepts a `Setting` and returns an env value; the implementer's report states it (D9).
- Criterion 4's registry half: the validator is built from the declared bounds data and its signature
  takes a value and no environment.
- The production registry still constructs without panicking (the existing gate-resident test).
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Let `rejectNamespace` skip the store-homed branch → a set store-homed key reports "unknown
   configuration key," so its row fails.
2. Let `Load` parse a set store-homed key → the "set, not parsing" `Load` row fails.
3. Drop the cross-family claim for the new family → the threshold-collision row fails.
4. Accept a zero apply mode → its panic row fails.

**Prior art:** none adopted; the registry is the repo's own (rung 1b). A layered-config library
(viper, koanf) is out of scope: the wall and its compat rules are ratified.

**Deliverable count: 4** (the family with its bounds data, the apply mode, the rule list, `Load`'s
classification).

---

## Task 3: `internal/settings`: the accessor, the seed rule, and the audit

Model: `opus`. Depends on Tasks 1 and 2. **Lands hook 15's settings home**, writing only through hook
9's one write method.

**Files:**
- Create: `internal/settings/` (`doc.go`, the accessor, the evaluation core and its from-handle
  entry point, the apply, and their tests; the from-facts entry point is Task 5's, since the store's
  row type lands there)

**Outcomes:**

- **The accessor** (R-6c8), with the generation-first check and typed reads.
- **The evaluation** (R-6c9): the core over `internal/settings`' own row type, with the env value
  read from the environ slice under the wall's rule, one outcome per registered setting from the
  closed table read per R-6c7, the correlated-rule refusals, and the orphan rows, joined with
  `errors.Join` where refusals are plural; the from-handle entry point exported.
- **The apply** (R-6c9), unexported: rows read through `tx.Query` inside one `IMMEDIATE`
  transaction through `Handle.Update`, every change written with its audit row (actor `system`,
  channel `env-seed`) and one generation bump, or nothing written on any refusal; `source` written as
  the literal `'env-seed'`; each re-seed and removal logged through a `*slog.Logger` the caller
  passes.
- **The seed table's rows**, each by its own test.

**States each check meets, and its report (the spec's closed table, `:751-763`, read per R-6c7, one
test per row):**

| Stored row | Env value | Outcome | Asserted |
| --- | --- | --- | --- |
| none | absent | no row; reads resolve the default | row count 0; read equals default |
| none | present | seeds a row, source `env-seed`, `value = seed_value`, audited | row, one audit row, generation +1 |
| `env-seed` | equal (typed equality, R-6c7, including `60m` against `1h`) | nothing changes | row, audit count, generation unchanged |
| `env-seed` | different | re-seeded, logged, audited | new value in both columns, one audit row with both values, log line |
| `env-seed` | absent | row removed, reads resolve the default, logged, audited (ruling 13) | row count 0, audit row with new value NULL, log line |
| `env-seed` | set to the empty string | as absent (the wall's empty-equals-unset rule): row removed, audited | row count 0, audit row |
| `admin` (`value` ≠ `seed_value`) | equal to `value` | nothing changes | unchanged |
| `admin` (`value` ≠ `seed_value`) | equal to `seed_value`, so different from `value` | refused naming both values and the action that resolves it | nothing written |
| `admin` | absent | the admin value stands | unchanged; read equals the admin value |
| `env-seed`, stored value outside the bounds | an in-bounds env value | re-seeded (healed), audited | new value, one audit row |
| `env-seed`, stored value outside the bounds | absent | removed (healed), audited | row count 0, audit row |
| `env-seed`, stored value outside the bounds | equal to the stored value | refused naming the key and the bound | nothing written |
| `admin`, stored value outside the bounds | any | refused naming the key and the bound | nothing written |
| a row for an unregistered key | any | ignored by the apply; reported as an orphan by the evaluation | row unchanged; orphan listed |
| any | an env value that does not parse under its kind | refused naming the key and the parse reason (R-6c5) | nothing written |
| a stored value that does not parse under its kind (a hand-edited row) | absent | refused naming the key and the reason | nothing written |
| two keys refusing for different reasons | in one seed set | both named in the joined error | nothing written |

Plus: an env value outside the bounds is refused naming the key and the bound, nothing written
(criterion 4); the accessor over an out-of-bounds or unparseable stored row errors naming the key;
the accessor over a store whose `settings_generation` row is missing errors naming the table and
serves no value.

**Acceptance criteria:**

- Spec criterion 4: the env-outside-bounds row. **Fixture state:** a current store with no row, so
  the only reason for refusal is the bound.
- Spec criterion 5: two test-registered duration settings, **expiry registered before retention**,
  with defaults that satisfy the rule alone (expiry `15m`, retention `48h`), and a rule "retention at
  least the expiry" (R-6c4). Stored `env-seed` rows: retention `48h`, expiry `15m`. Passing case:
  env expiry `72h` and retention `96h`, both present, applies. Failing case: env expiry `72h` and
  retention `48h`, **both present**, refuses naming both keys and writes nothing. **Fixture state:**
  both rows seeded and both env values present, so the failing case's only fault is the post-seed
  pair (never a removed row falling back to a default).
- Spec criterion 6: each row above by its own test.
- Spec criterion 7: an accessor under a registry whose default is `14` reads 14 from a store with no
  row; after closing, a registry declaring the same key with default `20`, over the same file, reads
  20, and the table still has no row.
- Spec criterion 9: a seed writes exactly one audit row, actor `system`, channel `env-seed`; a refused
  apply (the admin-conflict row) writes none, and leaves the generation unchanged.
- Spec criterion 10: two handles (`store.OpenWrite` twice on one state directory, the store created
  with `store.OpenServe`); an accessor on handle A reads and caches V1; an apply through handle B
  writes V2; A's next read returns V2. **Fixture state:** A's cache is warm (asserted by a read before
  the write).
- The admin-source guard: a test runs the apply over every seed-table state and asserts the count of
  `source = 'admin'` rows equals the fixture's own after each.
- Reviewer checklist lines for this task's diff-reviewer: every reload reads the generation before
  the rows and caches under that earlier generation (R-6c8); the apply reads rows through `tx.Query`
  inside the callback, never on the read pool (R-6c9); no comparison reads `seed_value` (R-6c7).
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Skip the generation check → criterion 10 fails.
2. Omit the generation bump from the apply → criterion 10 fails.
3. Store a row for a `default` source → criterion 7 fails.
4. Evaluate the correlated rule over the stored rows instead of the post-seed set → criterion 5's
   failing case applies, so it fails.
5. Validate after each key's seed in declaration order (incremental validation) → criterion 5's
   passing case refuses, since expiry is registered first.
6. Write the audit row outside the apply's transaction → criterion 9's refused case finds a row.
7. Compare the env against `seed_value` → the admin "different" row applies, so it fails.
8. Apply bounds to the stored value regardless of the env → the healed-by-re-seed row refuses, so it
   fails.

**Prior art:** CH's Lidarr and Sonarr split (a small config file for what the process needs before it
starts, the database for the rest) is the model this rung implements; no library is adopted.

**Deliverable count: 4** (the accessor, the evaluation, the apply with its audit, the seed table's
tests).

---

## Task 4: The five constants, re-homed

Model: `sonnet`. Depends on Task 3. **Neutral** (hook 10's backup is a consumer, unchanged).

**Files:**
- Modify: `cmd/dubplate-server/store.go` and `backup.go` (the five declarations, R-6c11; `runBackup`
  builds the accessor and injects the reader functions into `local.Config`), and their tests
- Modify: `internal/ops/local/local.go` and `jobs.go` (the reader functions in `Config`; every read
  before `insert`; the binder taking the three backup values), and their tests
- Modify: `internal/backup/guard.go`, `prune.go`, `backup.go` (the three constants become `Job`
  fields; `Job.Run`'s structural check, R-6c13; the guard's message), and their tests, with one test
  helper that fills the three fields with R-6c11's defaults, its literals citing R-6c11

**Outcomes:**

- **The five declarations** with R-6c11's kinds, defaults (unchanged from 6b-i), bounds (per Ruling
  3), and apply modes, each default's source cited in its comment; the four `internal/backup` and
  `ops/local` constants deleted, with `othersKeep` derived from the run's own keep value.
- **The reads**: `ops/local` calls its injected readers for `DUBPLATE_JOB_RETENTION` and, for a
  backup, the three backup values, all before `insert`; a reader error fails the start naming the
  key and leaves no job record.
- **`backup.Job`'s structural check** (R-6c13) before any rclone launch, and the guard's message
  naming the fraction in use.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Job start | no settings row | prune uses 90 days (the default) |
| Job start | a stored row for retention | prune uses the stored value |
| Job start | a stored retention outside the bounds | the start fails naming the key and the bound; no job record |
| Backup start | a stored backup key outside its bounds | the start fails naming the key and the bound; no job record |
| Backup run | a stored row for each backup key | the run's guard, keep, and archive retention use it |
| `Job.Run` | keep 0, fraction 0, fraction NaN, fraction above 1, archive retention 0, each alone | refused with the named code naming the field; the stand-in records no rclone call |
| Guard | a short tree under a configured fraction | the refusal text names that fraction |

**Acceptance criteria:**

- Spec criterion 13, the "reads through the accessor" half, per key: a stored `env-seed` row with a
  non-default value reaches the key's consumer, each over a fixture holding an item the default keeps
  and the stored value removes, so a consumer still reading the default fails:
  - retention: stored `48h`, and a finished job record aged 72h, which the prune deletes;
  - remote keep: stored `3` (the floor under either answer to Ruling 3), and five remote snapshots
    in the stand-in, pruned to three;
  - archive retention: stored `168h` (the floor under either answer), and an archive directory 200h
    past its move, which the prune purges;
  - guard: the admin half below.
  The local keep's half is Task 5's `doctor` row (D10).
- Spec criterion 13, the admin half: an `admin` row setting `DUBPLATE_BACKUP_GUARD_FRACTION` to `0.9`
  makes the run abort with no rclone invocation recorded by the stand-in, where the default `0.5`
  lets the same tree through (6b-i criterion 12). **Fixture state:** a populated skeleton with 10
  catalog tracks and a library tree holding 6 regular files, so 0.5 passes and 0.9 aborts.
- The existing 6b-i tests for criteria 12, 15, and 17 pass with only their construction updated: the
  three fields set through the test helper, and the readers passed to `local.New`.
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Keep reading the old constant for the guard → the admin-row test fails.
2. Read retention once at `local.New` instead of at each start → a test writing a row after `New`
   and before a start fails.
3. Remove `Job.Run`'s structural check → the zero-fraction row runs rclone, so it fails.
4. Read the backup values inside the binder, after `insert` → the out-of-bounds backup-key row finds
   a job record, so it fails.

**Deliverable count: 3** (the declarations, the reads, `backup.Job`'s structural check).

---

## Task 5: `backup`, `check`, and `doctor` evaluate the seed rule

Model: `sonnet`. Depends on Task 4. **Lands hook 15's marker display**; otherwise neutral.

**Files:**
- Modify: `internal/engine/store/inspect.go` (R-6c10) and `inspect_test.go`
- Create or modify: the from-facts entry point in `internal/settings` and its test
- Modify: `cmd/dubplate-server/backup.go`, `check.go`, `doctor.go`, and their tests

**Outcomes:**

- **Inspect's facts** carry the raw settings rows, the generation, and the table-absent marker, read
  on one shared inspect connection (R-6c10).
- **`backup`** evaluates through the from-handle entry point and logs one warning line per
  disagreement or refusal through its existing logger, naming the key and "serve applies it at its
  next start" or the refusal's action; it never fails on the evaluation and never writes a settings
  row.
- **`check`** evaluates through the from-facts entry point, prints one `settings:` line per non-quiet
  outcome, and on Ruling 2's yes exits nonzero on a refusing outcome.
- **`doctor`** evaluates through the from-facts entry point and prints one row per registered setting
  after the backup rows, in declaration order, and one row per orphan; on Ruling 2's yes a refusing
  row fails the run. The env value it shows comes from the outcome. The line-count tests derive the
  new rows from the registry. `newDoctorCmd` already takes the registry, so the goldens pass a test
  registry with no production test hook.

**`inspect`'s facts, per state:**

| Store | Facts' settings |
| --- | --- |
| current | the rows, the generation |
| older, version holds the settings table (migrated through `0004`, `0005` pending) | the rows, the generation |
| older, version predates the settings table (through `0003`) | marked table absent |
| newer, and every state `Inspect` returns no facts for | no facts (unchanged) |

**`doctor`'s settings row, per state (the house golden form: one literal expected line per state in
`doctor_test.go`; goldens build a `store.Report` by hand, `store_test.go:66`'s precedent):**

| Store | Row | Env | Row renders |
| --- | --- | --- | --- |
| current, or older with the settings table | none | absent | the default, source `default`, the apply mode |
| current, or older with the settings table | none | present | the default, source `default`, the apply mode, "env `<v>` seeds at serve's next start" |
| current, or older with the settings table | `env-seed` | equal | the value, source `env-seed`, the apply mode |
| current, or older with the settings table | `env-seed` | different | the stored value, `env-seed`, "env `<v>` re-seeds at serve's next start" |
| current, or older with the settings table | `env-seed` | absent | the stored value, `env-seed`, "returns to the default at serve's next start" |
| current, or older with the settings table | `admin` | equal or absent | the value, source `admin` |
| current, or older with the settings table | `admin` | different | refusing: both values and the resolving action |
| current, or older with the settings table | an out-of-bounds value that stays in force (R-6c7) | any | refusing: the key and the bound |
| current, or older with the settings table | an out-of-bounds `env-seed` row the env heals | in bounds, or absent | the stored value, "re-seeds" or "returns to the default" at serve's next start |
| current, or older with the settings table | a correlated pair | breaks its rule after the seed | refusing, on each key's row, naming both |
| current, or older with the settings table | any | does not parse | refusing: "invalid: `<reason>`" from the outcome |
| current, or older with the settings table | a stored value that does not parse | any | refusing: "invalid: `<reason>`" from the outcome |
| absent, or older without the settings table | n/a | absent / present | the default, source `default`, "store holds no settings yet"; with an env value, "env `<v>` seeds at serve's next start" |
| restore in progress, busy, not a store, refused, or newer | n/a | any | the default, "store not checked (`<state>`)"; with an env value, "env `<v>` not checked against the store" |
| current, or older with the settings table | an unregistered key | n/a | an orphan row: the key, "no longer registered; ignored" |

Each row that carries a sensitive setting appends the `sensitive` marker; the marker's golden uses a
test-registered sensitive setting, since no production key is sensitive until 6d. The local keep's
row with a stored `env-seed` value of `2` renders `2`, which is criterion 13's local-keep half (D10).

**Acceptance criteria:**

- `inspect`: each facts state above by its own test in `inspect_test.go`, the older stores built
  through `migrationSourceThrough` (the existing in-package precedent at `inspect_test.go:44`).
  **Fixture state:** the older-with-table store holds a settings row, so an empty result fails.
- The from-facts entry point: a facts value holding a seeded row resolves that value; a facts value
  marked table absent resolves every key to its default with the "no settings yet" report.
- Spec criterion 8: `backup` against a store holding an `env-seed` row for
  `DUBPLATE_BACKUP_SNAPSHOT_KEEP` of `14`, with the env at `20`, leaves every `settings`,
  `settings_audit`, and generation row byte-identical (dumped before and after), logs the
  disagreement naming the key, and runs with 14. **Fixture state:** a seeded row, so an apply would
  write.
- Spec criterion 12: every `doctor` state above as a golden, the orphan row included.
- Spec criterion 14: `check` with the env disagreeing with a seeded row prints the disagreement and
  leaves the database file's SHA-256 and mode unchanged, and the `-wal` file either absent or 0 bytes.
  **Fixture state:** every handle that seeded the store is closed and no `-wal` holds frames before
  the first hash, taken again after `check` returns; a seeded row and a divergent env, so an apply
  would write.
- On Ruling 2's yes: `check` and `doctor` exit 1 on the admin-conflict row and exit 0 on a pending
  re-seed; `backup` exits by its job's outcome alone in both.
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Let `backup` call the apply → criterion 8's byte comparison fails.
2. Let `check` call the apply through a write-mode open → criterion 14's hash comparison fails once
   that open closes and checkpoints.
3. Drop the orphan row → its golden fails.
4. On Ruling 2's yes, exit 0 on a refusing outcome → the exit row fails.

**Deliverable count: 4** (inspect's facts, `backup`, `check`, `doctor` with the from-facts entry
point).

---

## Task 6: The pass close, and the 6c to 6d handoff

Model: `sonnet`. Depends on Tasks 0 to 5 and the simplifier round.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/superpowers/research/handoffs/draft-STATUS-6c.md`
- Create: `docs/superpowers/research/handoffs/draft-HISTORY-6c.md`

**Outcomes:**

- **The HISTORY draft**: what landed per task with commit ids, the gate evidence with `-count=1` wall
  times, R-6c1 to R-6c13 with where each landed, the rulings' outcomes, the declared coverage gaps,
  both budgets, and "what a later pass would be wrong to rediscover," carrying at minimum the
  savepoint table, the typed-equality and value-not-seed-value rules (R-6c7), bounds on the value in
  force (R-6c7), the generation-first read (R-6c8), the settings-read-before-the-write rule (R-6c8),
  the pre-seed read through the facts (R-6c9), `backup.Job`'s structural check (R-6c13), and why the
  per-member tables are exempt (D1). Ruling 13's residual on data-loss keys: when an env line is
  dropped, the key returns to its default, and if the default is lower, the next backup prunes the
  difference. The errata D1 to D10, paste-ready and dated.
- **The handoff**:
  - To 6d: exporting the apply and wiring it into `serve` after the open and recovery and before the
    listener; resolving `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` through `Inspect` and the from-facts
    entry point before `OpenServe` and passing it as `ServeOptions.Keep`, with the `OpenServe` prune
    proof over a store at `0004` with a seeded keep row and `0005` pending (D10); moving the
    nonce-retention rule into the registry's list and re-running criterion 5 over the real
    nonce-retention and link-expiry pair (D6); re-homing the auth keys as settings with `sensitive`
    on both lifetimes and the cap; a "restart"-mode key read once at start, never through a reloading
    accessor (R-6c8); on Ruling 2's yes, `deploy.sh` halting on `check`'s nonzero exit before
    restarting `serve`; after a `restore`, the seed may refuse on a restored `admin` row and
    `env-seed` rows re-seed.
  - To 5d and 6d: between 6c's merge and 6d's, a deployed binary's `doctor` says an env value "seeds
    at serve's next start" while `serve` seeds nothing; the template carries none of the five keys,
    so the window matters only for a hand-set value.
  - To rung 14: the admin write path (source `admin`, `seed_value` recording the seed it replaced,
    its channel, its audit kind), and a resolving channel that works while `serve` is down for an
    `admin` row that refuses `serve`'s start, such as a CLI verb.
  - To rung 7: the reservations on `member_quota` and the default-collection preference. To the
    rungs that read them: the remaining preference columns.

**Acceptance criteria:** every count in the drafts was produced by a command the report pastes;
every commit id resolves; the STATUS draft is at most 60 lines. Gate green as a formality.

**Deliverable count: 2** (the HISTORY draft with the errata and the handoff, the STATUS draft).

---

## The exit gate

The gate green in the worktree and again on master after the merge, plus every assertion below.

1. Spec criteria 1, 2, and 3: Task 1.
2. Spec criterion 4: Tasks 2 (the validator built from bounds data) and 3 (the refusal).
3. Spec criteria 5, 6, 7, 9, and 10: Task 3.
4. Spec criterion 8: Task 5.
5. Spec criterion 11: Task 2, the store-homed half as D9 reads it, confirmed by the diff-reviewer.
6. Spec criterion 12: Task 5.
7. Spec criterion 13: Task 4, and the local keep's half in Task 5 (D10).
8. Spec criterion 14: Task 5.
9. Every state row above has a named test.
10. `go list -deps ./internal/config` shows neither `store` nor `settings`; `internal/backup` and
    `internal/ops/local` import no `settings`; `internal/settings` imports only `config` and `store`
    among internal packages.
11. No code outside a key's `Setting` declaration and tests names a store-homed key: `git grep -n`
    of each of the five keys' literals, excluding `*_test.go`, shows only its declaration, pasted.
    `internal/config` exports no reader of a setting's env value.
12. No production code writes source `admin`: `git diff <base>..HEAD -- internal cmd
    ':(exclude)*_test.go'` filtered to added lines matching `'admin'` or `"admin"`, pasted, every hit
    triaged with a one-line reason (a read path, `doctor`'s rendering, the `.sql` `CHECK`); the
    apply's writes carry `'env-seed'` as a literal; Task 3's admin-count test is green.
13. Every symbol the pass exports is listed with a `git grep` showing its production caller outside
    its own package, pasted. The named exceptions are exactly the correlated-rule entry point (6d),
    the sensitive marker's `true` value (6d), and the apply mode's "immediately" and "restart" values
    (6d).
14. `serve.go` is unchanged: `git diff --stat <base>..HEAD -- cmd/dubplate-server/serve.go` is empty.
15. `check.go` and `doctor.go` open the store only through `store.Inspect`: `git grep -n
    'store.Open' -- cmd/dubplate-server/check.go cmd/dubplate-server/doctor.go` is empty, pasted.

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's ownership table before the lane
launches (spec `:1152-1161`).

| Surface | Owner | 6c's position | Contention |
| --- | --- | --- | --- |
| `internal/config` (`registry.go`, `config.go`, the new `settings.go`) | rung 1b's contract | Task 2 | the 5c-iii security L2 grammar (`requireSiteName`, warm in the main tree at `098f8e5`) lands before the cut, or, if routed into 5c-iv, holds until 6c merges; the rest of 5c-iv writes none (its plan `:63`, `:496`); 5d holds until 6c merges or walls it off (R-6c1) |
| `internal/config/keys.go` | rung 1 | untouched | the site-name grammar may edit it; not 6c's |
| `internal/engine/store` migrations | rung 6 | Task 1: `0004`, `0005` | a lane that lands a migration first makes 6c renumber, and the rebase is non-fast-forward |
| `internal/engine/store/inspect.go`, `migrate.go` | rung 6 | Tasks 1, 5 | none after 6b-ii-b merges |
| `internal/settings` (new) | rung 6 | Tasks 3, 5 | none |
| `internal/ops/local`, `internal/backup` | rung 6 | Task 4 | none after 6b-ii-b merges |
| `cmd/dubplate-server/doctor.go`, `check.go` | rung 1's skeleton | Task 5 | the transliteration pass adds a `doctor` row and runs in lane 2 in sequence; 5d's probe rows wait for 6c's merge |
| `cmd/dubplate-server/store.go`, `backup.go` | rung 6 | Tasks 4, 5 | none |
| `cmd/dubplate-server/main.go` | rung 1 | untouched (expected) | the transliteration pass adds an `AddCommand`; a textual conflict keeps both |
| `go.mod`, `go.sum` | all | untouched | none |
| The state directory layout | rung 4 | unchanged | none |
| `docs/` ledgers | conductor | Task 6 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: confirm 6b-ii-b merged and its "to 6c" handoff is read; record the planning spend
   against the rung; cut the worktree from master; write the lane registry row and the
   contested-surface rows; arm the runaway guard; check AC power.
2. Before Task 1: rule Task 0's blockers, re-derive the ceiling from 6b-ii-b's actual (Task 0 item
   10), and record Rulings 1, 2, and 3.
3. At segment A's boundary (after Task 3): STATUS; re-derive.
4. After Task 5: one `code-simplifier` round over the pass's accepted Go, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, its review one `diff-reviewer` per package.
   No domain reviewer fan-out: 6c touches no `web/` and no auth surface, so review scales to its
   risk (CLAUDE.md, "Review scales to rung risk"). Each reviewer dispatch asks for correctness gaps
   ranked by consequence, not polish; a fold may refuse a finding whose fix costs more than the risk
   it removes, with a one-line reason.
5. The merge ritual, with one `diff-reviewer` on the merged result whenever the rebase is not a
   fast-forward, then **one `go-architecture-reader` per touched Go package**
   (`internal/engine/store`, `internal/config`, `internal/settings`, `internal/ops/local`,
   `internal/backup`, `cmd/dubplate-server`), no plan in context.
6. Record 6c's ceiling against the actual, and re-derive 6d's figure before its plan is drafted.
7. Release the hold on 5d's launch (R-6c1).
8. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task (`dubplate-implementer` at effort `high`, then
`diff-reviewer` at `claude-opus-5-5` effort `high`, then the gate), **sequential**, each segment
boundary on a task's own green committed gate. One re-dispatch on `fix`; a second `fix` is the
conductor's decision.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 0, 1, 2, 3 | STATUS; re-derive |
| **B** | 4, 5, then the simplifier round, Task 6 | the merge ritual, six architecture reads, the ledger fold |

Segment A closes after Task 3 because the accessor's invalidation and the seed apply are the pass's
novel logic, and a measurement there that contradicts this plan needs a ruling before Tasks 4 and 5
build on it.

**Models.** `sonnet` for Tasks 1, 4, 5, and 6. **`opus` for Tasks 2 and 3, dispatched as
`claude-opus-5-5`**: Task 2 changes the config contract every rung registers through, and Task 3 is
the cross-process invalidation and the transactional seed. Task 0 runs at `opus` in fresh context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1.
- Tasks 1 and 2 are independent by files; both precede Task 3. They run sequentially, because one
  tree runs one executor.
- Task 3 → Task 4 → Task 5 → the simplifier round → Task 6.

**Budget:** the 5.4M ceiling and 4.32M flag (Ruling 1), planning recorded separately, STATUS at each
boundary, spend scored at close against the ceiling and against zero attended interactions after
approval.

---

## The 0.8 forward hooks, per task

Per CLAUDE.md's build-for-0.8 rule. The streamer spec's forward-hooks list (`:683-768`) was read in
full, the 2026-09-25 amendments to hooks 15, 17, and 20 included.

| Task | Hook |
| --- | --- |
| 0 | Neutral. |
| 1 | Lands hook 15's routed tables and hook 4 on `settings_audit` and `move_intents`; consumes hook 5; hardens hook 7. Hook 15's share rows stay reserved (no table, no column). Hook 20 weighed: no serve event, neutral. Hooks 1, 2, 3, and 11 stay with their computing rungs (spec ruling 2). |
| 2 | Lands hook 15's metadata (apply mode, sensitive marker) on store-homed keys. Hook 18 (every key's description and sensitive mark) is rung 14's, not built. |
| 3 | Lands hook 15's settings home, through hook 9's write method. |
| 4 | Neutral. |
| 5 | Lands hook 15's marker in `doctor`. |
| 6 | Neutral. |

No new hook was discovered.

---

## Prior art consulted

Scored on CLAUDE.md's four criteria. The spec's table (`:895-917`) carries the evidence for the rows
it scored; this pass adopts nothing new, so no new license, maintenance, or adoption evidence is
required. Of CLAUDE.md's list, the entries with a 6c surface are the driver, goose, and Navidrome.

| Entry | Verdict for 6c |
| --- | --- |
| `modernc.org/sqlite` v1.59.0 | **Adopt** (spec). The constraint codes and the savepoint behavior this plan relies on are measured against it. |
| `pressly/goose` v3.28.0 | **Decline** (spec: its silent-ahead `Up`). R-6c12 closes the one gap the hand-rolled runner had that a library might have covered. |
| `golang-migrate/migrate` v4 | **Decline** (spec). |
| Navidrome's schema (GPL, behavior only) | **Consulted, declined.** Its per-user settings are a key-value `user_props` table with `on delete cascade`, which invariants 5 and 7 forbid; typed per-member columns replace it (R-6c6). Its `property` table (instance-wide key-value) is the shape `settings` shares, which invariant 5 permits because `settings` carries no foreign key. |
| Lidarr and Sonarr's config split (CH) | **Adopt the model** (CH's own source): the wall for what the process needs before it starts, the store for tunables. No code. |
| viper, koanf | **Decline:** the registry and its compat rules are ratified (rung 1b); a layering library would duplicate them. |
| gonic, OpenSubsonic, taglib, mewkiz/flac, Picard, chromaprint | No 6c surface. |

---

## Not verified in this revision

- The 5c-iii fold's disposition of security L2: where the site-name grammar lands (it was warm and
  uncommitted in the main tree at `098f8e5`), or whether it is routed into 5c-iv. Task 0 blocker 3.
- 5d's plan: whether it walls off `internal/config`, `doctor.go`, and `check.go`. Task 0 blocker 4;
  the conductor holds 5d's launch otherwise.
- 6b-ii-b's final shape of `internal/backup`, `Handle.Query`'s merge, and its handoff "to 6c." Task
  0 blockers 2 and 7 and item 8.
- The defaults: 6b-i's values are carried unchanged; the local keep's 3 is a working value Task 4
  sources or keeps with its stated reason.
- The savepoint and constraint probes ran outside the gate (the mechanics review re-ran them); Task
  1's tests are their in-gate form.
- Whether a `STRICT` table's `BLOB` primary key refuses `NULL` without `NOT NULL`: not relied on;
  `0005` declares it and Task 1 proves it.
