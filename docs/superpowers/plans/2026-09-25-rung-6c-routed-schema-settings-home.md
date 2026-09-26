# Rung 6, Pass 6c (The Routed Schema and the Settings Home) Implementation Plan

Status: **draft for approval, revision 1, 2026-09-25.** Two rulings for Geoff ("Rulings awaiting
owner"). Ruling 1 is the ceiling and the lane placement; both answers are executable as written. No
`spec-plan-review` round has run on this revision yet; the conductor runs it before the approval
sitting, and its fold record is this plan's review history.

**Ceiling:** 5.1M execution (flag 4.08M), planning recorded against the rung and outside it.
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
("What this spec routes elsewhere") and "The server's own knobs"; the 0.8 streamer spec
(`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`, which still says 2.0), "Forward hooks
the current build owes" (`:683-768`) and "Prior art" (`:770-806`); the repo `CLAUDE.md`; the 6b-i
plan's R-6bi8 (`docs/superpowers/plans/2026-09-24-rung-6b-i-ops-seam-backup.md:604-623`); and
`docs/HISTORY.md`'s 6a, 6b-i, and 6b-ii-a entries.

**Code pointers were resolved on master at `5b8f05e` and on branch `lane-2-6b-ii-b` at `30a66ac`**
(6b-ii-b's Task 5A; segment B2 was still running, with warm edits under `internal/backup`). Every
pointer into `internal/backup` is cited by symbol, never by line, because 6b-ii-b is rewriting that
package now: at `30a66ac` the three backup constants sit at `prune.go:20`, `:24`, and `:28`, where
master has them at `:23`, `:27`, and `:31`. 6c executes only after 6b-ii-b merges, and Task 0
re-resolves every pointer against the merged master.

**What this pass is.** It lands every table CH routes to rung 6 that 6a did not (spec criteria 1 to
3), gives the config registry the four pieces of per-key metadata and the correlated-rule list CH
asks for (criteria 4, 5, and 11), builds `internal/settings`, the accessor with cross-process
invalidation, the closed seed-rule table, and the audit (criteria 4 to 10), re-homes 6b-i's five
constants as the first store-homed keys (criterion 13), and teaches `backup`, `check`, and `doctor`
to evaluate the seed rule read-only (criteria 8, 12, and 14). It also carries 6b-i's routed risk R13:
the migration runner's `SAVEPOINT` check, which turns a migration file carrying `COMMIT; BEGIN;` from
silent into a refusal.

**What this pass is not.** `serve` does not open the store, apply the seed rule, or read a setting:
that wiring, the seed rule's export, and the migration runner's production caller are 6d's (spec
`:213-216`, `:740-741`). No auth threshold moves to the store, and `authConfig` is untouched (6d's
seed handoff, spec `:773-800`). No admin write path, no mount, no re-authentication gate (rung 14).
No `Identity.UserID`, no roster import (6d). No share table or column (hook 15's 2026-09-25
amendment). No new module dependency. No `systemd/`, `scripts/`, `env/`, `web/`, `config/`, or
`internal/config/keys.go` change, and **no deploy of any kind.** A task that believes it needs any of
these stops and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the pass-execute chain per
> task, sequential, in the segments "Execution model" names. The implementer is
> `dubplate-implementer`. The conductor consumes structured reports and never reads a diff, a test
> log, or a gate transcript inline. Subagents never dispatch agents, so `code-simplifier` is the
> conductor's dispatch. The conductor owns `docs/STATUS.md`, `docs/HISTORY.md`, `ROADMAP.md`, the
> lane registry, and the contested-surface rows. The live-executor check is the lane registry,
> never `pgrep -f Projects/dubplate`.

---

## Rulings awaiting owner

Two questions. Everything else this plan decides is a question with one architecturally correct
answer, recorded with its reason under "Resolutions this plan takes."

**1. Does pass 6c run under a 5.1M execution ceiling, in lane 2 tonight once 6b-ii-b merges?**
Recommendation: yes. Reason: "Token ceiling" projects 4.08M from 6b-i's and 6b-ii-a's recorded
actuals, and "Lane condition, answered" shows 5c leaves `internal/config` alone for the rest of its
plan while 5d cannot start before its plan is approved in the morning.

What each answer builds: yes sets the flag at 4.08M, the projection itself, so any overrun raises it
and the conductor asks one combined question before the merge-side reads; the pass runs tonight
after 6b-ii-b merges (and after the transliteration pass, if the conductor keeps that order in lane
2). No holds the spec-derived execution figure of 4.5M (flag 3.6M); cumulative spend with the fix reserve is
about 3.15M after Task 5 and crosses 3.6M during the close, so the conductor asks after Task 5, before the simplifier round. "No"
to the lane half means 6c waits for 5d's merge and runs after it, with Task 0 re-reading 5d's
`internal/config` and `doctor.go` edits.

**2. Do `doctor` and `check` exit nonzero when the seed rule would refuse `serve`'s start?**
The refusing outcomes are an admin row the env contradicts, a stored or env value outside the
binary's bounds, and a correlated rule the post-seed set breaks. Recommendation: yes. Reason:
`doctor` and `check` both answer "can this deployment start," and 6d's `deploy.sh` runs the new
binary's `check` before restarting `serve` (spec `:395-396`), so a `check` that exits 0 on a refusal
lets a deploy restart a server that then will not start. The spec's reason for "without failing"
("so a timer-run backup never fails over an auth threshold it does not read," spec `:744-745`)
holds for `backup`, which never fails on the evaluation under either answer. A pending re-seed, an
unseeded env value, or a pending return to default is a disagreement, not a refusal, and never
fails any verb under either answer.

What each answer builds: yes makes `doctor` and `check` exit 1 on a refusing outcome, each naming
the key, both values or the bound, and the action that resolves it; this deviates from the spec's
"`check` included" sentence (spec `:743-745`) and owes it a dated erratum (D3). No keeps the spec's
text: both verbs print the same refusal line and exit on today's rules alone, and 6d's `deploy.sh`
must parse the line to catch a refusal before restarting.

---

## Lane condition, answered

The spec gives 6c "lane 2 worktree if 5c and 5d leave `internal/config` untouched; otherwise after
5d" (spec `:1062`, `:1173-1174`). The purpose is that two lanes never edit `internal/config`
concurrently. The evidence, read at `5b8f05e`:

- **5c's committed work never touched it.** `git log -- internal/config` shows its last commit is
  `6e810c4` (6b-i Task 6). No 5c-i, 5c-ii, or 5c-iii commit appears.
- **5c-iii and 5c-iv plan no edit to it.** The 5c-iii plan, which carries 5c-iv as Tasks 7 to 11,
  says "No edit to `internal/config`, `internal/auth`, `internal/identity`, `internal/engine`, or
  `cmd/dubplate-server/{check,doctor,store,backup}.go`" (`plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md:63-64`)
  and "It writes nothing under `internal/config`" (`:496`).
- **The in-flight review fold is the one open pressure.** The security review's L2 asks for a
  `DUBPLATE_SITE_NAME` grammar "at `internal/config/config.go:352`, `keys.go:23`"
  (`research/2026-09-25-rung-5c-iii-review-security.md:21-23`). The fold's warm edits at
  `5b8f05e` are `cmd/dubplate-server/rendered_test.go` and `web/scripts/rendered/run.ts` only, so it
  has not landed there. The fold precedes 5c-iii's close, which precedes 5c-iv, while 6c follows
  6b-ii-b's segments B2 and B3 and its merge. Wherever the grammar lands, it is committed on master
  before 6c's worktree is cut, so 6c rebases onto it rather than racing it. Task 0 blocker 3 checks
  this.
- **5d cannot run tonight.** Its plan is being drafted tonight as a document lane for Geoff's morning
  review (`docs/STATUS.md`, "Overnight 2026-09-25"). 5d's scope (R5S's pass table) adds the
  forward-auth and trusted-proxy integration classes, which under the rung 1b contract register from
  `cmd/dubplate-server`, and `doctor` probe rows in `doctor.go`, which 6c also edits. The conductor
  therefore holds 5d's launch until 6c merges, or 5d's plan walls off `internal/config`,
  `doctor.go`, and `check.go`; the contested-surface rows below record it.

So 6c runs before 5d under those two conditions, and the spec's "otherwise after 5d" branch applies
only if a lane 1 edit under `internal/config` is in flight at launch.

---

## Token ceiling: 5.1M execution (Ruling 1)

**Basis.** 6b-i (`docs/HISTORY.md`, its "Both budgets"): about 5.4M through the simplifier round,
with Task 0 at about 0.08M and the simplifier round at about 0.16M, over ten build tasks (1S, 1T,
1A, 1B, 2 to 7) and six fix rounds, so (5.40 − 0.08 − 0.16) / 10 ≈ **0.52M per task**, fix rounds
included; its merge ritual 0.08M and three architecture reads 0.24M, so **0.08M per read**. 6b-ii-a:
about 1.95M in all for Task 0 and three tasks, with its plan pricing the close at 0.99M, so
(1.95 − 0.99 − 0.11) / 3 ≈ **0.28M per task** with one fix round. 6c has no data-destructive path
and no security surface, so it prices a `sonnet` task at 0.45M and an `opus` task at 0.55M to
0.65M, with fix rounds reserved separately.

1. Task 0 (`opus`, fresh context): **0.10M**.
2. Task 1, the routed schema and the runner's `SAVEPOINT` check (`sonnet`): **0.45M**.
3. Task 2, the registry metadata (`opus`, the config contract): **0.55M**.
4. Task 3, `internal/settings` (`opus`, the novel logic): **0.65M**.
5. Task 4, the five constants re-homed (`sonnet`): **0.45M**.
6. Task 5, `backup`, `check`, and `doctor` (`sonnet`, the widest state table): **0.55M**.
7. Fix rounds beyond those rates: **0.40M**.
8. The close side: the simplifier round with its review (0.20M), Task 6 (0.15M), the merge ritual
   with a possible merged-range review (0.10M), and six architecture reads at 0.08M
   (`internal/engine/store`, `internal/config`, `internal/settings`, `internal/ops/local`,
   `internal/backup`, `cmd/dubplate-server`: 0.48M): **0.93M**.
9. **The projection is 4.08M.**

In two lines: 0.10 + 0.45 + 0.55 + 0.65 + 0.45 + 0.55 + 0.40 + 0.93 = 4.08M projected;
4.08M / 0.8 = 5.10M, so a **5.1M** ceiling with its 80% flag at **4.08M**.

The spec's own row was 5.6M including 1.07M of plan preparation and 10% for the config contract
(spec `:1140`); its execution-only equivalent is (0.2 + 7 × 0.46 + 0.6) × 1.1 ≈ 4.42M, which Ruling 1's
"no" rounds to 4.5M. Seven tasks is under ten, so the pass does not split.

Deliverable count: twenty across Tasks 1 to 6, plus Task 0's report. Tasks 1, 2, 3, and 5 carry four
each (16); Task 4 carries two, Task 6 two. No task passes four.

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
  `registerBackupConfig` already sit in `newProductionRegistry`). It writes no `docs/` file outside
  Task 6's drafts.
- **Imports follow the spec's table** (`:197-205`): `internal/settings` imports `config`, `store`, and
  the standard library; `internal/ops/local` gains `settings`; `internal/backup` still imports only
  `store` and `atomicfile`; `internal/config` never imports `store` or `settings`; the store reads no
  setting's meaning (it may return raw rows, as its `Facts.LastSuccess` already returns raw job rows).
- **No new module dependency.** `go.mod` and `go.sum` stay byte-identical.
- **The store's rules carry over** (6b-i's global constraints): every store write goes through
  `Handle.Update`; times are integer UTC microseconds; every DSN comes from the one builder
  (`internal/engine/store/conn.go:187`); new code matches no raw SQLite result code outside the
  store's classifiers.
- **Nothing is exported ahead of a cross-package caller.** At the close every new export has a
  production caller outside its own package, with exactly the exceptions exit gate 13 names: the
  registry's correlated-rule entry point (the spec assigns it to 6c; its first production rule is 6d's
  nonce-retention rule) and the `Sensitive` field's `true` value (6d declares the first sensitive
  keys). The seed rule's apply stays unexported until 6d's `serve` (spec `:215-216`).
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

From `go doc` at `go1.27.1 linux/amd64`, `modernc.org/sqlite v1.59.0` from the module cache, and
three probes run on 2026-09-25 in a scratch module against that driver (the probe source is not
committed; each task that relies on a measurement re-runs it as a test).

- `go doc database/sql.DB.QueryRowContext`: "If the query selects no rows, the *Row.Scan will return
  ErrNoRows." The accessor's absent-row state is `sql.ErrNoRows` from `Scan`, never a nil error with
  zero values. `Handle.QueryRow` (`internal/engine/store/conn.go:108`) runs on the read pool, which
  opens with `_query_only=1` (spec `:320-322`).
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
- **Measured, the savepoint probe.** Inside an `_txlock=immediate` transaction, `SAVEPOINT runner`,
  then a body, then `RELEASE runner`:
  | Body | `RELEASE runner` | `Commit` | Tables after |
  | --- | --- | --- | --- |
  | `CREATE TABLE a(...)` | nil | nil | 1 |
  | `CREATE TABLE a(...); COMMIT; BEGIN; CREATE TABLE b(...)` | `no such savepoint: runner (1)` | nil | 2 |
  | `CREATE TABLE a(...); COMMIT;` | `no such savepoint: runner (1)` | `cannot commit - no transaction is active (1)` | 1 |
  | `CREATE TABLE a(...); ROLLBACK;` | `no such savepoint: runner (1)` | `cannot commit - no transaction is active (1)` | 0 |
  | a table plus `CREATE TRIGGER t AFTER INSERT ON a BEGIN SELECT 1; END;` | nil | nil | 1 |

  So the `RELEASE` failure catches all three self-terminating shapes, the silent `COMMIT; BEGIN;`
  included, and a trigger's `BEGIN ... END` body passes. It detects; it does not undo the part
  committed before the file's own `COMMIT`, which the scratch store in the gate and the pre-migration
  copy on a real store cover.
- **Measured, the constraint probe** (`_foreign_keys=1`, `STRICT` tables), extended codes from
  `sqlite.Error.Code()`, whose doc reads "Code returns the sqlite result code for this error":
  a Go `string` bound into a `STRICT` `BLOB` path column answers 3091, "cannot store TEXT value in
  BLOB column"; a `CHECK (visibility IN ('group','private'))` violation answers 275; a `NULL` into a
  `NOT NULL` `at` answers 1299; an orphan foreign key answers 787. The constants are
  `SQLITE_CONSTRAINT_DATATYPE = 3091`, `SQLITE_CONSTRAINT_CHECK = 275`,
  `SQLITE_CONSTRAINT_NOTNULL = 1299`, and `SQLITE_CONSTRAINT_FOREIGNKEY = 787` in the driver's
  `lib/sqlite_linux_amd64.go`.
- **Read in the tree.** 6a already declares `collections.visibility TEXT NOT NULL CHECK (visibility IN
  ('group', 'private'))` (`internal/engine/store/migrations/0003_catalog.sql`), so criterion 2 is a
  proof over an existing constraint, not a new column. Invariant 5 (`checkAttributeBags`,
  `invariant_test.go`) refuses any table that pairs a foreign key with a `name`, `key`, or `attr`
  column and a `value` column, so a per-member key-value table cannot pass the gate.
- **Navidrome's per-user settings, consulted for behavior only.** `model/user_props.go` declares
  `Put(ctx, userId, key, value)` and `Get`, and its table is `user_props (user_id ... references user
  on update cascade on delete cascade, key varchar not null, value varchar)`
  (`db/migrations/20210625223901_add_referential_integrity_to_user_props.go`, read 2026-09-25). That
  is the attribute bag invariant 5 forbids and the cascade invariant 7 forbids; see "Prior art."

---

## Privileged fields and their channels

| Field | Channels that may set it at 6c | What refuses every other channel |
| --- | --- | --- |
| A settings row's source `admin` | None in production. Tests write one directly through `Handle.Update` with literal SQL (spec `:765-766`). | No exported or unexported production function writes `admin`; the seed apply writes only `env-seed` rows and removals. Exit gate 12 greps for it. |
| A settings row's source `env-seed`, and a row's removal back to the default | The seed apply alone, unexported at 6c, called only by `internal/settings`' own tests until 6d's `serve`. | Every other verb evaluates read-only (R-6c9); `backup`, `check`, and `doctor` hold no write path to `settings`. |
| A `settings_audit` row | The seed apply, in the same transaction as the row it records. | No other writer exists; a refused apply rolls back its audit rows with everything else (criterion 9). |
| The settings generation | The seed apply bumps it inside every transaction that changes a settings row. | Nothing else writes it; a test that writes a fixture row bumps it by literal SQL in the same transaction. |
| A store-homed key's home, apply mode, sensitive marker, and bounds | The declaring rung's own file, at registration (`AddSetting` or its final name), in `cmd/dubplate-server`. | Never the env, never the store: the registry is code. A malformed declaration panics at registration (R-6c2). |
| A store-homed key's env value | Read by one function in `internal/config` whose only production caller is `internal/settings` (R-6c2). | No `Parse*` wall getter accepts a store-homed declaration; that is a compile error (criterion 11). |
| Rows in `member_overrides`, `member_preferences`, `member_quota`, `group_quota`, `device_credentials`, `move_intents` | None at 6c: each table's first writer is a later rung (R-6c6). | No production code writes them; tests populate fixtures only. |

---

## Resolutions this plan takes rather than discovers

Each resolution names its site. A task landing the surface a resolution decides writes nothing the
resolution does not settle; a question the plan left open is a stop-and-report, never a guess.

### R-6c1: the lane placement

As "Lane condition, answered" argues. The spec's condition is honored in substance: no two lanes
edit `internal/config` concurrently. It is recorded as spec reading D4 so a reader of the spec's
literal "5c and 5d leave `internal/config` untouched" finds why a pass that runs before 5d's plan
exists satisfied it.

### R-6c2: a store-homed key is its own declaration family

Site: `internal/config`. A store-homed key is declared as a distinct type (working name `Setting`),
registered through its own entry point into a new registry family, never as a `Threshold` with a
home field. So "a store-homed key has no wall getter" (spec `:734-736`) is a compile-time fact:
`ParseInt`, `ParseDuration`, `ParseFloat`, and `ParseString` (`internal/config/thresholds.go:84`
onward) take a `Threshold`, and no function returns a `Setting`'s env value as a typed Go value.
The home is the family: a key registered as a threshold stays on the wall and reads from the env
exactly as today (criterion 11's first half), and a key registered as a setting is store-homed.

A `Setting` carries its key, its kind (the existing `ThresholdKind`), its default in string form, its
accepted-form doc, its apply mode, its sensitive marker, and its bounds as a validator that takes a
value and no environment (spec `:730-733`). The family claims across families like every other, so
a setting key that collides with a threshold, a wall row, a class key, or a superseded name panics
naming both (the `claim` rule, `registry.go`). Registration also panics when the default does not
parse under its kind, when the default fails its own validator, when the apply mode is the zero
value, or when the validator is nil: each is a programmer's mis-declaration, on the registry's
existing panic precedent.

The seed rule reads a setting's env value through one `internal/config` function that returns the
raw string and whether it is set, with the wall's empty-equals-unset rule (`config.go`'s
`wall.get`). Its name and doc say the seed rule is its reader, and exit gate 11 pastes the
`git grep` showing `internal/settings` as its only non-test caller.

### R-6c3: the metadata lives on settings only

Apply mode (immediately, next operation, restart), the sensitive marker, and the value validator
land on `Setting` alone. The fifteen wall thresholds keep their shape: their bounds stay in the
sources' `Validate` hooks until 6d moves the auth keys, and the serve and shadow keys stay on the
wall at restart (spec `:791-795`). Hook 18 (a description and a sensitive mark on every registered
key) is rung 14's and is not built here.

### R-6c4: correlated rules are registry-level and proved with a test pair

Site: `internal/config` (declaration) and `internal/settings` (evaluation). The registry holds a
list of named correlated rules, each over two or more registered settings and a function of their
resolved post-seed values that returns an error naming every key it compares, on the
`sourceKeyError` precedent (`cmd/dubplate-server/auth.go`). A rule naming an unregistered or wall key
panics at registration. None of the five keys 6c re-homes correlates with another, and the one real
rule the spec names, nonce retention against link expiry, moves from the magic-link source's
`Validate` hook with 6d's re-home (spec `:787`). Criterion 5 is therefore proved in
`internal/settings` with two test-registered settings shaped as that rule (retention at least the
expiry). The entry point is exit gate 13's named exception.

### R-6c5: `Load` knows store-homed keys, parses them, and never bounds them

Site: `internal/config/config.go`. `rejectNamespace` (`config.go:243`) classifies a set store-homed
key as declared, beside its threshold branch, so setting one is not "unknown configuration key."
`Load` parses every set store-homed key's env value under its kind and reports a failure as a wall
error naming the key and its doc, the spec's "any | unparseable | a wall error, as today" row
(`:763`). `Load` never applies a setting's bounds: a bound is the seed rule's, so a verb that does
not read an auth threshold never fails on one.

### R-6c6: the routed schema's shape

Site: `internal/engine/store/migrations/`. Two migrations, numbered after the highest on master at
Task 0 (today `0003_catalog.sql`, so `0004` and `0005`). Table names are working names (spec
`:472-473`); the contents below are the contract.

- **The settings home** (`0004`): `settings`, keyed by `key` (TEXT primary key), with the value (TEXT),
  the source (`CHECK (source IN ('env-seed', 'admin'))`, since a `default` source stores no row), the
  seed value, the actor, and `updated_at`; `settings_audit` with an opaque `id`, `at` and `kind` both
  `NOT NULL`, the actor, the channel, the key, and the old and new values (either nullable, for a
  seed from nothing and a removal); and `settings_generation`, one row, whose primary key is a
  column not named `id` or `*_id` with `CHECK` pinning it to one value, holding an integer generation
  the migration inserts at 0. `kind` and `channel` carry no `CHECK`, on `jobs.kind`'s precedent:
  rung 14 adds values, and a `STRICT` table's `CHECK` changes only by a rebuild.
- **CH's per-member, credential, quota, and journal tables** (`0005`):
  - `member_overrides`, keyed by `user_id` (a `BLOB` primary key referencing `users`), with the four
    overrides CH names (CH, "Two surfaces, one API"): a quota in bytes, may-upload, may-create-
    collections, and an upload concurrency cap, each nullable (NULL means no override) with a
    `CHECK` on its domain (non-negative bytes, 0 or 1 flags, a cap of at least 1), and `updated_at`.
  - `member_preferences`, keyed by `user_id` the same way, with `collapse_duplicates` (0 or 1,
    default 0), the one preference CH names for the rung 6 schema ("The keys are named now so the
    rung 6 schema carries them," CH, "Rules that cut across the verbs"), and `updated_at`. Theme,
    language, ReplayGain mode, sort order, notification choices, and the default collection each
    migrate in with the rung that first reads them, under CLAUDE.md's 2026-09-12 refinement and U4
    ("Adding a built-in is a migration and a typed accessor in the built-in's own rung," spec
    `:527`), the same reasoning the spec's ruling 2 applied to the S20 columns.
  - `member_quota`, keyed by `user_id`, with the charged bytes (non-negative) and `updated_at`; and
    `group_quota`, one row on `settings_generation`'s singleton pattern, with the physical total.
    Reservations and true-up are rung 7's first consumer's to add.
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
  applies to every table.

Each migration ships a populated fixture under `testdata/fixtures/` (R-6a15), populated as of its
version: a user, a collection, one row in every new table, and, for `0004`, an `env-seed` row, an
`admin` row, two audit rows, and a generation above 0.

### R-6c7: what a settings row means

No row means the source is `default`, and reads resolve the registry's current default, so a
re-researched default reaches every store (spec `:746-748`, criterion 7). A row stores the value as
the text it was seeded or written with. Two values are equal when they parse to the same typed value
under the key's kind, so an env value of `60m` against a stored `1h` is "equal to the seed" and
changes nothing; the stored text is not rewritten. The actor of a seed is `system` and its channel
`env-seed` (spec `:747`).

### R-6c8: the accessor

Site: `internal/settings`. An accessor is bound to one `*store.Handle` and the registry, is safe for
concurrent use, and reads a setting by its declaration, one typed read per kind. Each read first
reads the generation on the read pool (one indexed read); an unchanged generation serves the cached
value, and a changed one reloads. A missing row resolves the current default. A stored value that
fails to parse or falls outside the binary's bounds is an error naming the key and the bound, never
a silent default. A generation read that fails is an error; the cache is never trusted without a
successful generation read.

A consumer reads a setting before it enters `Handle.Update`, not inside the callback. The spec's
rule that "every read a write decides on runs inside the callback" (`:355`) governs rows the write
changes. A setting read a moment before the write is the value in force for that operation, which is
exactly the apply mode "next operation" the five keys declare (R-6c11).

### R-6c9: evaluation is exported, the apply is not

Site: `internal/settings`. One exported evaluation takes the registry, the environment, and the stored
rows, and returns one outcome per registered setting from the closed table (spec `:751-763`), the
correlated-rule refusals, and the orphan rows, without writing. `backup` (from its write-mode handle),
`check`, and `doctor` (both from inspect mode's facts, R-6c10) are its production callers. The apply,
which runs that evaluation inside one `IMMEDIATE` transaction and writes every outcome's change, its
audit rows, and one generation bump, or refuses and writes nothing, stays unexported until 6d's
`serve` exports it. The pre-seed read (the migration runner's local snapshot cap reads "the stored row
or the registry default," spec `:748-749`) is the accessor itself.

### R-6c10: inspect mode returns the raw settings rows

Site: `internal/engine/store/inspect.go`. `Facts` (`:126`) gains the settings rows as stored (key,
value, source, seed value, actor, time) and the generation, read by `readFacts` (`:223`) on the
read-only connection it already opens, when the store's version holds the settings table (the
`jobsVersion` precedent at `:121`). The store interprets nothing: `internal/settings` evaluates. So
`check` and `doctor` stay read-only (criterion 14) and open no second connection. The field is
exported with Task 5, its first caller.

### R-6c11: the five keys, their homes, defaults, and bounds

Site: `cmd/dubplate-server/store.go` (`DUBPLATE_JOB_RETENTION`, `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`)
and `cmd/dubplate-server/backup.go` (the three backup keys), registered by the existing
`registerStoreConfig` and `registerBackupConfig`, never in `internal/config/keys.go`. The defaults
are 6b-i's working values (R-6bi8) unless Task 4's research gives a recorded reason to change one,
and each definition's comment cites its source (BS's researched-default rule, spec `:768-771`).

| Key | Kind | Default | Bounds | Apply mode | Consumer at 6c |
| --- | --- | --- | --- | --- | --- |
| `DUBPLATE_JOB_RETENTION` | duration | `2160h` (90 days) | 24h to 87600h | next operation | `ops/local`'s prune at each job start (replacing `jobs.go`'s `retention`) |
| `DUBPLATE_BACKUP_GUARD_FRACTION` | float | `0.5` | above 0, at most 1; NaN and infinities refused | next operation | each backup run's guard (replacing `guard.go`'s `guardFraction`) |
| `DUBPLATE_BACKUP_SNAPSHOT_KEEP` | int | `14` | 1 to 1000 (the latest is never pruned, spec ruling 10) | next operation | each run's prune (replacing `prune.go`'s `snapshotKeep`; `othersKeep` derives from it) |
| `DUBPLATE_BACKUP_ARCHIVE_RETENTION` | duration | `2160h` | 24h to 87600h | next operation | each run's prune (replacing `archiveRetention`) |
| `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` | int | `3` | 1 to 100 (`ServeOptions.Keep` "must be at least 1," `open.go:92-94`) | next operation (the next migration) | `doctor`'s row; 6d's `serve` passes it as `ServeOptions.Keep` |

6b-i's R-6bi8 said 6d declares `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` "and 6c re-homes it after,"
written when 6c was expected after 6d. 6c now runs first, so 6c declares it as a store-homed key
whose first production reader is `doctor`, which answers R-6bi8's `unused` concern; 6d wires it into
`serve`. Its default of 3 is a working value: a rollback needs the newest copy, and two older copies
cover a bad upgrade found a release late, each a full store whose size `doctor` already shows; Task
4 records a source or keeps 3 with this reason. `internal/backup` still imports no `settings`: its
three values become fields of `backup.Job`, which is one run's inputs (`backup.go`, "A Job is one
backup run's inputs"), filled by `ops/local` from the accessor at each job start.

### R-6c12: the migration runner's savepoint

Site: `internal/engine/store/migrate.go`. `applyInTx` (`:298`) runs each file's body between
`SAVEPOINT` and `RELEASE`, so a body that ends its own transaction fails at `RELEASE`, naming the
file, before the version write and the ledger row. It covers the rebuild path too, which calls the
same function. The savepoint probe above makes `commitMigration`'s "no transaction is active" branch
(`:355`) unreachable for a file's own `COMMIT` or `ROLLBACK`, so the task removes that branch or
states why a residual case still reaches it, never leaving dead code.

---

## Spec deviations and readings that take effect on approval

| # | Spec site | Change | Why |
| --- | --- | --- | --- |
| D1 | `:455-458`, invariant 6's exempt list | adds `member_overrides`, `member_preferences`, `member_quota` | R-6c6 |
| D2 | `:495-496`, `member_overrides` and `member_preferences` | the columns are CH's named four and `collapse_duplicates`; other preferences arrive with their consumers | R-6c6 |
| D3 | `:743-745`, "`check` included" | on Ruling 2's yes, `doctor` and `check` fail on a refusing outcome; a disagreement still fails nothing | Ruling 2 |
| D4 | `:1062`, `:1173-1174`, the lane condition | read as "no concurrent edit to `internal/config`," satisfied by sequencing | R-6c1 |
| D5 | 6b-i plan R-6bi8 | 6c declares `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`; 6d wires it | R-6c11 |
| D6 | criterion 5's "the correlated retention rule" | proved with a test pair shaped as the rule; the real rule moves at 6d | R-6c4 |

Task 6 drafts each as a paste-ready dated erratum.

---

## Review focus

The input classes most likely to bite, each pinned by a test in the task that owns it.

1. **A settings write in one process and a cached read in another.** Expected: the next read sees
   the write. Pinned in Task 3 (criterion 10).
2. **A seed set that is valid only together** (both correlated keys rise in one deploy). Expected:
   accepted; one key alone past the stored other refused, nothing written. Pinned in Task 3.
3. **A refused seed** (admin conflict, a bound, a correlated rule). Expected: no row, audit row, or
   generation change. Pinned in Task 3 (criterion 9).
4. **A float bound fed `NaN` or `Inf`.** Expected: refused by name. Pinned in Tasks 2 and 4.
5. **A verb that must not write** (`backup`, `check`, `doctor`) meeting an env that disagrees with
   the store. Expected: rows and file bytes unchanged, the disagreement reported. Pinned in Task 5.
6. **A migration body that commits itself.** Expected: refused naming the file, no ledger row.
   Pinned in Task 1.
7. **A store-homed key read through the wall.** Expected: does not compile. Checked by the
   diff-reviewer in Task 2 and by exit gate 11.

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
   `backup.Job`'s fields; `ops/local`'s `retention` and its reader `Runner.insert`, and
   `Runner.backupJob`; `store.ServeOptions`, whose `Keep` field must survive beside 6b-ii-b's new
   `Reconcile`; whether `cmd/dubplate-server/restore.go` passes a `Keep` or any of the five values.
3. **Is `internal/config` quiet?** `git log <this plan's commit>..HEAD -- internal/config` and the
   main tree's `git status`: any lane 1 commit there is a note naming what changed (the 5c-iii
   fold's `DUBPLATE_SITE_NAME` grammar, if it landed there); any uncommitted edit under
   `internal/config` in the main tree is a blocker (R-6c1). Report the lane registry's lane 1 row.
4. **Has 5d started?** If a 5d plan is approved and its lane is open, report whether its plan walls
   off `internal/config`, `doctor.go`, and `check.go`; a 5d executor writing any of them is a
   blocker (R-6c1).
5. **What is the highest migration on master?** 6b-ii-b's segments at `30a66ac` added none; report
   `ls internal/engine/store/migrations/` so Task 1 numbers from it.
6. **Does no `internal/settings` package, no settings table, and no store-homed declaration exist?**

**Notes, not blockers:**

7. Report 6b-ii-b's Task 8 handoff "to 6c" items (its draft or the folded HISTORY entry) and route
   each to the task it touches, or report that it names none.
8. Report at HEAD: `newProductionRegistry`'s registration calls; `registerStoreConfig` and
   `registerBackupConfig`; `doctor.go`'s row order (the store row, then the backup rows) and its
   line-count tests' `wantTotal` expressions; `check.go`'s store row; `invariant6Exempt`;
   `applyInTx` and `commitMigration`; `Inspect`, `Facts`, `readFacts`, `jobsVersion`;
   `Handle.QueryRow`; `rejectNamespace`, `thresholdOf`, `AddThreshold`, and `claim`.
9. Report 6b-ii-b's actual spend from its HISTORY entry; the conductor re-derives this plan's ceiling
   from it before Task 1 (spec `:1147-1149`).
10. Report the conductor's recorded outcome of Rulings 1 and 2.

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
- **The savepoint** (R-6c12), with the dead branch resolved.
- **The constraint proofs** for criteria 1 to 3.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Invariant test | every migration applied to an empty database | no violation |
| Invariant test | each migration's populated fixture migrated to head | no violation |
| Invariant test | a migration with no fixture | the subtest fails naming the migration (existing) |
| `move_intents` insert | a Go `string` in `old_path` or `new_path` | refused, extended code 3091 |
| `member_overrides` insert | a `user_id` with no `users` row | refused, extended code 787 |
| `collections` insert | `visibility` `public`, and `''` | refused, extended code 275, each |
| `collections` insert | `group`, `private` | accepted, each |
| `settings_audit`, `move_intents` insert | `at` NULL, `kind` NULL | refused, extended code 1299, each of the four |
| `settings` insert | source `default` | refused, extended code 275 |
| Migration apply | a synthetic file with `COMMIT; BEGIN;` | refused naming the file; no ledger row for it; `user_version` unchanged |
| Migration apply | a synthetic file ending `COMMIT;`, and one ending `ROLLBACK;` | refused naming the file, each |
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
- Create: `internal/config/settings.go` (the declaration type, the apply-mode type, the correlated
  rule type, the seed value reader) and `settings_test.go`
- Modify: `internal/config/registry.go` (the family, its entry points, the rule list, the family
  name constant) and `registry_test.go`
- Modify: `internal/config/config.go` (`rejectNamespace` and `Load`, R-6c5) and `config_test.go`

**Outcomes:**

- **The declaration family** (R-6c2) with its registration panics and its accessor returning the
  settings in declaration order, on `Thresholds()`'s precedent.
- **The apply-mode type** with its three named values and a zero value that registration refuses;
  its `String` names each (on `ThresholdKind.String`'s precedent).
- **The correlated-rule list** (R-6c4) with its registration panics.
- **`Load`'s classification and parse** (R-6c5), and the seed value reader (R-6c2).

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Registration | a well-formed setting | registered; listed in order |
| Registration | a key already a threshold, wall row, class key, or superseded name | panic naming both families |
| Registration | a default that does not parse under its kind | panic naming the key and the default |
| Registration | a default its own validator refuses | panic naming the key |
| Registration | a zero apply mode, or a nil validator | panic naming the key and the missing field |
| Registration | a key without the `DUBPLATE_` prefix | panic (the existing prefix rule) |
| Rule registration | a rule over registered settings | registered |
| Rule registration | a rule naming an unregistered key, or a threshold | panic naming the key |
| `Load` | a store-homed key unset or empty | no error |
| `Load` | set and parsing under its kind | no error; nothing bounds it |
| `Load` | set and not parsing | an error naming the key and its doc |
| `Load` | set and parsing but outside its bounds | no error (the seed rule's) |
| Seed value reader | unset, empty, set | not set, not set, the raw string |
| Validator | a float setting given `NaN`, `Inf`, `-Inf` | refused, each by name (test-registered float) |

**Acceptance criteria:**

- Spec criterion 11: a threshold registered with no home still resolves from the env through
  `ParseDuration` (an existing threshold, asserted unchanged). The store-homed half is compile-time:
  the diff-reviewer confirms no `Parse*` function accepts the new type and no function returns a
  typed value from the env for it, and the implementer's report states it.
- Criterion 4's registry half: the validator's signature takes a value and no environment, and a
  test calls it with no environment in scope.
- The production registry still constructs without panicking (the existing gate-resident test).
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Let `rejectNamespace` skip the store-homed branch → a set store-homed key reports "unknown
   configuration key," so its row fails.
2. Let `Load` apply bounds → the "outside its bounds" row fails.
3. Drop the cross-family claim for the new family → the threshold-collision row fails.
4. Accept a zero apply mode → its panic row fails.

**Prior art:** none adopted; the registry is the repo's own (rung 1b). A layered-config library
(viper, koanf) is out of scope: the wall and its compat rules are ratified.

**Deliverable count: 4** (the family, the apply mode, the rule list, `Load` with the seed reader).

---

## Task 3: `internal/settings`: the accessor, the seed rule, and the audit

Model: `opus`. Depends on Tasks 1 and 2. **Lands hook 15's settings home**, writing only through hook
9's one write method.

**Files:**
- Create: `internal/settings/` (`doc.go`, the accessor, the evaluation, the apply, and their tests)

**Outcomes:**

- **The accessor** (R-6c8), with the generation check and typed reads.
- **The evaluation** (R-6c9): one outcome per registered setting from the closed table, the
  correlated-rule refusals, and the orphan rows, joined with `errors.Join` where refusals are
  plural; exported.
- **The apply** (R-6c9), unexported: evaluation inside one `IMMEDIATE` transaction through
  `Handle.Update`, every change written with its audit row (actor `system`, channel `env-seed`) and
  one generation bump, or nothing written on any refusal; each re-seed and removal logged through a
  `*slog.Logger` the caller passes.
- **The seed table's eleven rows**, each by its own test.

**States each check meets, and its report (the spec's closed table, `:751-763`, one test per row):**

| Stored row | Env value | Outcome | Asserted |
| --- | --- | --- | --- |
| none | absent | no row; reads resolve the default | row count 0; read equals default |
| none | present | seeds a row, source `env-seed`, audited | row, one audit row, generation +1 |
| `env-seed` | equal (typed equality, R-6c7, including `60m` against `1h`) | nothing changes | row, audit count, generation unchanged |
| `env-seed` | different | re-seeded, logged, audited | new value, one audit row with both values, log line |
| `env-seed` | absent | row removed, reads resolve the default, logged, audited (ruling 13) | row count 0, audit row with new value NULL, log line |
| `admin` | equal | nothing changes | unchanged |
| `admin` | different | refused naming both values and the action that resolves it | nothing written |
| `admin` | absent | the admin value stands | unchanged; read equals the admin value |
| any | stored value outside the bounds | refused naming the key and the bound | nothing written |
| a row for an unregistered key | any | ignored by the apply; reported as an orphan by the evaluation | row unchanged; orphan listed |
| any | unparseable | a wall error (Task 2's `Load` row); the evaluation, handed one, reports it without writing | nothing written |

Plus: an env value outside the bounds is refused naming the key and the bound, nothing written
(criterion 4); the accessor over an out-of-bounds stored row errors naming the key and the bound.

**Acceptance criteria:**

- Spec criterion 4: the env-outside-bounds row. **Fixture state:** a current store with no row, so
  the only reason for refusal is the bound.
- Spec criterion 5: with two test-registered duration settings and a rule "retention at least the
  expiry" (R-6c4), stored `env-seed` rows retention `48h` and expiry `15m`; env expiry `72h` and
  retention `96h` together applies; env expiry `72h` with retention unchanged refuses naming both keys
  and writes nothing. **Fixture state:** both rows seeded, so the second case's only fault is the
  post-seed pair.
- Spec criterion 6: each of the eleven rows above by its own test.
- Spec criterion 7: an accessor under a registry whose default is `14` reads 14 from a store with no
  row; after closing, a registry declaring the same key with default `20`, over the same file, reads
  20, and the table still has no row.
- Spec criterion 9: a seed writes exactly one audit row, actor `system`, channel `env-seed`; a refused
  apply (the admin-conflict row) writes none, and leaves the generation unchanged.
- Spec criterion 10: two handles (`store.OpenWrite` twice on one state directory, the store created
  with `store.OpenServe`); an accessor on handle A reads and caches V1; an apply through handle B
  writes V2; A's next read returns V2. **Fixture state:** A's cache is warm (asserted by a read before
  the write).
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Skip the generation check → criterion 10 fails.
2. Omit the generation bump from the apply → criterion 10 fails.
3. Store a row for a `default` source → criterion 7 fails.
4. Validate each key's bounds against the pre-seed set rather than the post-seed set → criterion 5's
   passing case fails.
5. Write the audit row outside the apply's transaction → criterion 9's refused case finds a row.

**Prior art:** CH's Lidarr and Sonarr split (a small config file for what the process needs before it
starts, the database for the rest) is the model this rung implements; no library is adopted.

**Deliverable count: 4** (the accessor, the evaluation, the apply with its audit, the seed table's
tests).

---

## Task 4: The five constants, re-homed

Model: `sonnet`. Depends on Task 3. **Neutral** (hook 10's backup is a consumer, unchanged).

**Files:**
- Modify: `cmd/dubplate-server/store.go` and `backup.go` (the five declarations, R-6c11), and their
  tests
- Modify: `internal/ops/local/local.go` and `jobs.go` (the accessor in `Config`; the retention read
  at each job start; the backup job's three fields filled from the accessor), and their tests
- Modify: `internal/backup/guard.go`, `prune.go`, `backup.go` (the three constants become `Job`
  fields), and their tests
- Modify: `cmd/dubplate-server/backup.go`'s `runBackup` to build the accessor and pass it to
  `local.New`

**Outcomes:**

- **The five declarations** with R-6c11's kinds, defaults, bounds, and apply modes, each default's
  source cited in its comment; the four `internal/backup` and `ops/local` constants deleted, with
  `othersKeep` derived from the run's own keep value.
- **The reads**: `ops/local` reads `DUBPLATE_JOB_RETENTION` before each job's insert and the three
  backup values before building each `backup.Job`; an accessor error fails the start naming the key.

**States each check meets, and its report:**

| Check | State | Report |
| --- | --- | --- |
| Job start | no settings row | prune uses 90 days (the default) |
| Job start | a stored row for retention | prune uses the stored value |
| Job start | a stored value outside the bounds | the start fails naming the key and the bound; no job record |
| Backup run | a stored row for each backup key | the run's guard, keep, and archive retention use it |
| `doctor` (Task 5) | `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` | its row shows value, source, apply mode |

**Acceptance criteria:**

- Spec criterion 13, the "reads through the accessor" half, per key: a test that a stored `env-seed`
  row with a non-default value reaches the key's consumer (the prune cutoff; the guard; the remote
  keep count; the archive cutoff; for the local keep, the accessor's value passed as
  `ServeOptions.Keep` to `store.OpenServe` over a store with pending migrations prunes pre-migration
  copies to that count, 6a criterion 9's setup).
- Spec criterion 13, the admin half: an `admin` row setting `DUBPLATE_BACKUP_GUARD_FRACTION` to `0.9`
  makes the run abort with no rclone invocation recorded by the stand-in, where the default `0.5`
  lets the same tree through (6b-i criterion 12). **Fixture state:** a populated skeleton with 10
  catalog tracks and a library tree holding 6 regular files, so 0.5 passes and 0.9 aborts.
- The existing 6b-i tests for criteria 12, 15, and 17 pass unchanged against the defaults.
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Keep reading the old constant for the guard → the admin-row test fails.
2. Read retention once at `local.New` instead of at each start → a test writing a row after `New`
   and before a start fails.

**Deliverable count: 2** (the declarations, the reads).

---

## Task 5: `backup`, `check`, and `doctor` evaluate the seed rule

Model: `sonnet`. Depends on Task 4. **Lands hook 15's marker display**; otherwise neutral.

**Files:**
- Modify: `internal/engine/store/inspect.go` (R-6c10) and `inspect_test.go`
- Modify: `cmd/dubplate-server/backup.go`, `check.go`, `doctor.go`, and their tests

**Outcomes:**

- **Inspect's facts** carry the raw settings rows and the generation (R-6c10).
- **`backup`** evaluates from its write-mode handle and logs one warning line per disagreement or
  refusal through its existing logger, naming the key and "serve applies it at its next start" or
  the refusal's action; it never fails on the evaluation and never writes a settings row.
- **`check`** evaluates from the inspect facts, prints one `settings:` line per non-quiet outcome, and
  on Ruling 2's yes exits nonzero on a refusing outcome.
- **`doctor`** prints one row per registered setting after the backup rows, in declaration order, and
  one row per orphan; on Ruling 2's yes a refusing row fails the run. The line-count tests derive the
  new rows from the registry.

**`doctor`'s settings row, per state (the house golden form: one literal expected line per state in
`doctor_test.go`):**

| Store | Row | Env | Row renders |
| --- | --- | --- | --- |
| current | none | absent | the default, source `default`, the apply mode |
| current | none | present | the default, source `default`, the apply mode, "env `<v>` seeds at serve's next start" |
| current | `env-seed` | equal | the value, source `env-seed`, the apply mode |
| current | `env-seed` | different | the stored value, `env-seed`, "env `<v>` re-seeds at serve's next start" |
| current | `env-seed` | absent | the stored value, `env-seed`, "returns to the default at serve's next start" |
| current | `admin` | equal or absent | the value, source `admin` |
| current | `admin` | different | refusing: both values and the resolving action |
| current | any | a stored or env value outside the bounds | refusing: the key and the bound |
| current | a correlated pair | breaks its rule after the seed | refusing, on each key's row, naming both |
| absent, older, busy, or unreadable | n/a | any | the default or the env value, "store not checked (`<state>`)" |
| current | an unregistered key | n/a | an orphan row: the key, "no longer registered; ignored" |
| any | any | unparseable | "invalid: `<reason>`" from `Load`'s status (the threshold rows' precedent) |

Each row that carries a sensitive setting appends the `sensitive` marker; the marker's golden uses a
test-registered sensitive setting, since no production key is sensitive until 6d.

**Acceptance criteria:**

- Spec criterion 8: `backup` against a store holding an `env-seed` row for
  `DUBPLATE_BACKUP_SNAPSHOT_KEEP` of `14`, with the env at `20`, leaves every `settings`,
  `settings_audit`, and generation row byte-identical (dumped before and after), logs the
  disagreement naming the key, and runs with 14. **Fixture state:** a seeded row, so an apply would
  write.
- Spec criterion 12: every `doctor` state above as a golden, the orphan row included.
- Spec criterion 14: `check` with the env disagreeing with a seeded row leaves the database file's
  SHA-256 and mode unchanged and prints the disagreement. **Fixture state:** a seeded row and a
  divergent env, so an apply would write.
- On Ruling 2's yes: `check` and `doctor` exit 1 on the admin-conflict row and exit 0 on a pending
  re-seed; `backup` exits by its job's outcome alone in both.
- `cairn-run-gate 'bash <worktree>/scripts/check.sh'` green.

**Named mutations this task reports:**

1. Let `backup` call the apply → criterion 8's byte comparison fails.
2. Read the settings in `check` through a write-mode open → criterion 14's hash comparison or its lock
   row fails; report which.
3. Drop the orphan row → its golden fails.
4. On Ruling 2's yes, exit 0 on a refusing outcome → the exit row fails.

**Deliverable count: 4** (inspect's facts, `backup`, `check`, `doctor`).

---

## Task 6: The pass close, and the 6c to 6d handoff

Model: `sonnet`. Depends on Tasks 0 to 5 and the simplifier round.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/superpowers/research/handoffs/draft-STATUS-6c.md`
- Create: `docs/superpowers/research/handoffs/draft-HISTORY-6c.md`

**Outcomes:**

- **The HISTORY draft**: what landed per task with commit ids, the gate evidence with `-count=1` wall
  times, R-6c1 to R-6c12 with where each landed, the rulings' outcomes, the declared coverage gaps,
  both budgets, and "what a later pass would be wrong to rediscover," carrying at minimum the
  savepoint table, the typed-equality rule (R-6c7), the settings-read-before-the-write rule (R-6c8),
  and why the per-member tables are exempt (D1). The errata D1 to D6, paste-ready and dated.
- **The handoff**: to 6d, exporting the apply and wiring it into `serve` after the open and
  recovery and before the listener, passing `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` as
  `ServeOptions.Keep`, moving the nonce-retention rule into the registry's list, re-homing the auth
  keys as settings with `sensitive` on both lifetimes and the cap; to rung 14, the admin write path
  (source `admin`, its channel, its audit kind); to rung 7, the reservations on `member_quota` and
  the default-collection preference; to the rungs that read them, the remaining preference columns.

**Acceptance criteria:** every count in the drafts was produced by a command the report pastes;
every commit id resolves; the STATUS draft is at most 60 lines. Gate green as a formality.

**Deliverable count: 2** (the HISTORY draft with the errata and the handoff, the STATUS draft).

---

## The exit gate

The gate green in the worktree and again on master after the merge, plus every assertion below.

1. Spec criteria 1, 2, and 3: Task 1.
2. Spec criterion 4: Tasks 2 (the validator's signature) and 3 (the refusal).
3. Spec criteria 5, 6, 7, 9, and 10: Task 3.
4. Spec criterion 8: Task 5.
5. Spec criterion 11: Task 2, the store-homed half a compile-time fact the diff-reviewer confirmed.
6. Spec criterion 12: Task 5.
7. Spec criterion 13: Task 4.
8. Spec criterion 14: Task 5.
9. Every state row above has a named test.
10. `go list -deps ./internal/config` shows neither `store` nor `settings`; `internal/backup`
    imports no `settings`; `internal/settings` imports only `config` and `store` among internal
    packages.
11. The seed value reader's only non-test caller is `internal/settings`: the `git grep`, pasted.
12. No production code writes source `admin`: `git grep -n "'admin'" -- 'internal/**/*.go'
    ':!*_test.go'` shows only the schema's `CHECK` and read paths, pasted.
13. Every symbol the pass exports is listed with a `git grep` showing its production caller outside
    its own package, pasted. The named exceptions are exactly the correlated-rule entry point (6d)
    and the sensitive marker's `true` value (6d).
14. `serve.go` is unchanged: `git diff --stat <base>..HEAD -- cmd/dubplate-server/serve.go` is empty.

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's ownership table before the lane
launches (spec `:1152-1161`).

| Surface | Owner | 6c's position | Contention |
| --- | --- | --- | --- |
| `internal/config` (`registry.go`, `config.go`, the new `settings.go`) | rung 1b's contract | Task 2 | lane 1's 5c-iv writes none (its plan `:63`, `:496`); the 5c-iii fold's site-name grammar lands before the cut or waits; 5d holds until 6c merges or walls it off (R-6c1) |
| `internal/config/keys.go` | rung 1 | untouched | the site-name grammar may edit it; not 6c's |
| `internal/engine/store` migrations | rung 6 | Task 1: `0004`, `0005` | a lane that lands a migration first makes 6c renumber, and the rebase is non-fast-forward |
| `internal/engine/store/inspect.go`, `migrate.go` | rung 6 | Tasks 1, 5 | none after 6b-ii-b merges |
| `internal/settings` (new) | rung 6 | Task 3 | none |
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
   9), and record Rulings 1 and 2.
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

**Budget:** the 5.1M ceiling and 4.08M flag (Ruling 1), planning recorded separately, STATUS at each
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

- The 5c-iii fold's disposition of security L2: where the site-name grammar lands, or whether it is
  refused or routed. Task 0 blocker 3.
- 5d's plan: whether it walls off `internal/config`, `doctor.go`, and `check.go`. Task 0 blocker 4;
  the conductor holds 5d's launch otherwise.
- 6b-ii-b's final shape of `internal/backup` and its handoff "to 6c." Task 0 blocker 2 and item 7.
- The researched defaults: 6b-i's working values are carried; Task 4 records a source for each or
  keeps it with its stated reason.
- The savepoint and constraint probes ran outside the gate; Task 1's tests are their in-gate form.
