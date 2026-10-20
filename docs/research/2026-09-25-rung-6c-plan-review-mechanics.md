# Rung 6c plan review: mechanics and feasibility

Plan under review: `docs/plans/2026-09-25-rung-6c-routed-schema-settings-home.md` at
`8d0eed8`, revision 1. Lens: does every mechanism behave as the plan states, and can each task
build in the stated order against master at `8d0eed8` and `lane-2-6b-ii-b` at `30a66ac`? Every
claim below was checked with a scratch probe against `modernc.org/sqlite v1.59.0` from the module
cache (`go1.27.1 linux/amd64`), with `go doc`, or by reading the source. Findings are ranked by
consequence. Polish is left out.

**Counts:** 0 blockers, 2 majors, 13 minors. No owner forks: each finding has one correct answer,
and this review proposes it.

## What was verified and holds

- **The savepoint table (plan `:245-258`) reproduces exactly.** All five rows give the stated
  `RELEASE`, `Commit`, and table count under `_txlock=immediate`. The real runner rolls back after a
  failed `RELEASE` (`rollbackMigration`, `migrate.go:337`), so the probe also ran that flow. For
  `COMMIT; BEGIN; CREATE TABLE b`, the rollback undoes `b` and `a` stays committed, with
  `user_version` unchanged. The Task 1 row "no ledger row, `user_version` unchanged" holds.
- **The four constraint codes hold**, and the datatype check runs before `CHECK`. A Go `string` bound
  into a `STRICT` `BLOB` column carrying 6a's full path `CHECK` (copied from `0003_catalog.sql`)
  answers 3091 whether the string is a valid path, `../x`, or empty. A `..` component in a `BLOB`
  answers 275. `NULL` into `at` or `kind` answers 1299. An orphan `user_id` answers 787.
  `visibility` `public` and `''` answer 275, and so does source `default`. A second singleton row
  answers 275 under `CHECK (singleton = 1)`, and a duplicate answers 1555. So the Task 1 fixture
  "a `move_intents` row otherwise valid" does isolate the bind as the only fault.
  `(*Error).Code` doc: "Code returns the sqlite result code for this error."
- **Invalidation across two handles works as R-6c8 states.** Two independent pools opened on one WAL
  file with the repo's DSN keys: pool A reads generation 0, pool B commits a row plus a bump, and
  pool A's next `SELECT` on its reused idle connection reads 1. `OpenWrite` twice in one process is
  feasible because it takes `LOCK_SH` through `syscall.Flock` on its own descriptor
  (`lock.go:49`, `:89`). Two shared locks coexist.
- **`check`'s read-only open leaves the database file's bytes and mode unchanged.** Probe: after an
  `inspectDSN`-shaped open (`mode=ro&_query_only=1`), a read, and a close, the file's SHA-256 and
  mode are the same. SQLite creates `-wal` (0 bytes) and `-shm` beside it, which `Inspect`'s own doc
  already states (`inspect.go:157-161`).
- **The invariant checks behave as R-6c6 assumes.** Invariant 3 (`checkKindAt`) catches the
  `_audit` suffix and any `kind` column. Invariant 5 passes `settings` because it carries no foreign
  key. Invariant 6 refuses a `user_id` primary key unless the table is exempt, so mutation 2 kills.
- **The quoted docs are accurate:** `QueryRowContext`, `ParseFloat` on NaN and Inf, `ParseDuration`
  units, `Atoi`, and `errors.Join`.
- **Cited pointers resolve at `8d0eed8`:** `conn.go:108` and `:187`, `thresholds.go:84`,
  `config.go:243`, `inspect.go:121`, `:126`, and `:223`, `migrate.go:298` and `:355`,
  `open.go:92-94`, and `prune.go:23/27/31`, with `:20/24/28` on the lane. So do the 5c-iii plan's
  `:63-64` and `:496`, the security review's `:21-23`, and R-6bi8's `:604-623`. The last commit
  under `internal/config` is `6e810c4`.
- **Lane 2 at `30a66ac`:** restore and fetch read none of the four backup constants.
  `ServeOptions.Keep` survives beside the new `Reconcile`. No migration was added. `migrate.go` is
  untouched, and `inspect.go` gains only the intent refusal. No production caller of `OpenServe`
  exists on either tree.

## Majors

### M1. The pre-seed read cannot be the accessor, and criterion 13's local-keep proof cannot be built

Location: plan `:444-445` (R-6c9, "The pre-seed read ... is the accessor itself"), `:470`, and
`:821-825` (Task 4 criterion 13, the local keep).

Defect: the accessor is bound to a `*store.Handle`, and a handle exists only after an open.
`OpenWrite` refuses an older store ("start serve to migrate", `open.go:175`). `OpenServe`
migrates before it returns. So no accessor can read `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` before the
migration that value caps. A store below the settings migration has no `settings` table at all, so
an accessor read there fails. The test as written ("the accessor's value passed as
`ServeOptions.Keep` to `store.OpenServe` over a store with pending migrations") also cannot be built
in any package:

- Only `internal/engine/store`'s own tests can build an older store, through
  `migrationSourceThrough` (`invariant_test.go:506`).
- That package's in-package test cannot import `internal/settings`, which imports `store`. Go
  refuses the import cycle.
- `cmd/dubplate-server` and `internal/ops/local` have no older-store fixture. The only one is a
  hand-built `store.Report` (`store_test.go:66`).

Task 4's implementer would hit a stop-and-report on this row.

Fold:

- **R-6c9.** The pre-seed read is the evaluation's resolved value over `Inspect`'s raw rows (R-6c10).
  It is the registry default when the store's version predates the settings table.
- **Task 4, criterion 13's local-keep half.** Prove that a stored `env-seed` row reaches `doctor`'s
  row, the key's only 6c consumer. Add a `cmd` test that resolves the value from `Inspect` facts over
  a current store.
- **Task 6's handoff.** Hand the `OpenServe` prune proof to 6d, `serve` being the first caller, with
  this mechanism named. 6a criterion 9 already proves that `Keep` prunes.

### M2. `ops/local` has no way to name the four declarations, and the backup keys' read point is unstated

Location: plan `:458-460` and `:477-479` (R-6c11), `:794-807` (Task 4 Files and Outcomes), and
`:811-817` (Task 4 states).

Defect: R-6c11 declares the five `config.Setting` values in `cmd/dubplate-server/store.go` and
`backup.go`, which are package `main`. The accessor "reads a setting by its declaration" (R-6c8),
but `internal/ops/local` cannot reference a package `main` variable, and neither can its tests.
The plan names neither the declarations' home nor how they reach the runner.

Second, the backup body is built by a binder, `binder func(kl *store.KindLease, h ops.Handle) body`
(`jobs.go:37`). `start` calls it after `insert` has committed the queued record
(`local.go`, `j.body = bind(kl, j.handle)`), and the binder returns no error. An accessor error on a
backup key read inside the binder therefore lands after a job record exists. The state table gives
"no job record" only for retention, and names no row for an out-of-bounds backup key.

Fold:

- **How the declarations reach the runner.** `local.Config` carries the accessor plus the four
  declarations, passed by `runBackup`, or `cmd/dubplate-server` injects typed reader functions. The
  latter is the spec's own pattern for 6d's auth keys ("`cmd/dubplate-server` injects each auth
  threshold as an accessor function," spec `:209-210`). Pick one in the plan.
  `internal/ops/local`'s tests then register their own settings.
- **The read point.** `start` (or `StartBackup`) reads all four before `insert`. The binder takes
  the three backup values as arguments.
- **Task 4's state table** gains "Backup start | a stored backup key outside its bounds | the start
  fails naming the key and the bound; no job record."

## Minors

1. **The `commitMigration` branch stays reachable, and a same-named savepoint evades the check.**
   Location: plan `:486-488`. Probe: a body ending `COMMIT; SAVEPOINT runner;` passes `RELEASE`,
   which commits the body's own transaction. The version write and the ledger insert then autocommit,
   and `Commit` answers "cannot commit - no transaction is active." So the branch at `migrate.go:355`
   is still the only detection there, and removing it loses that. A body ending
   `COMMIT; BEGIN; SAVEPOINT runner;` passes `RELEASE` and `Commit` both, with the first part
   committed outside the runner: silent. Fold: keep the branch, citing this probe row. Name the
   savepoint something no migration would write (for example `dubplate_runner`), and have
   `readMigrations` refuse a file whose body contains that name, case-insensitively. Add both shapes
   to Task 1's state table. Also note that for the `COMMIT;` and `END;` shapes, `rollbackMigration`
   joins a second error, "cannot rollback - no transaction is active." Assertions should match the
   file name and the `RELEASE` failure, never the whole text.
2. **The constants' file is misnamed.** Location: plan `:264-266`. `SQLITE_CONSTRAINT_CHECK = 275`,
   `_DATATYPE = 3091`, `_FOREIGNKEY = 787`, and `_NOTNULL = 1299` sit in the driver's
   `lib/sqlite.go:3214-3224`. That file is generated "from the per-target sqlite_*.go files", and
   `lib/sqlite_linux_amd64.go` carries no `SQLITE_CONSTRAINT` name. Fold: correct the citation.
3. **The compile-time claim is narrower than stated.** Location: plan `:306-314`, `:289`, and
   `:686-689`. Probe: `cfg.Threshold(s)` from a `Setting` fails to compile ("cannot convert"). But
   `ParseString(env, cfg.Threshold{Key: s.Key, Default: s.Default})` compiles and reads the env, and
   so does any `Threshold` literal carrying a store-homed key: `Parse*` accepts an unregistered
   `Threshold`. Fold: state the fact as "no function accepts a `Setting`." The diff-reviewer confirms
   three things: `Setting` does not embed `Threshold`, no method or field yields one, and a
   `git grep` shows each of the five keys' literal only in its `Setting` declaration. That grep is
   pasted at exit gate 11.
4. **Exit gate 12's grep misses a bound parameter and points at the wrong file type.** Location:
   plan `:957-958` and `:284`. The `CHECK` lives in a `.sql` file, which the `*.go` grep never
   shows. A writer that binds `source` through `?` from a Go constant `"admin"` passes the grep.
   Fold: the apply's write statements inline `'env-seed'` as a SQL literal and never bind `source`
   as a parameter. The grep covers both `'admin'` and `"admin"` under `internal/settings`. Task 3
   asserts the count of `source = 'admin'` rows unchanged after every apply path.
5. **Criterion 14's hash can pass vacuously.** Location: plan `:892-894` and `:902-903`. Probe: a
   write through a read-write pool leaves the main file's hash unchanged until the last connection
   closes and checkpoints. A write-mode open that only reads never changes it. So mutation 2 (a read
   through a write-mode open) can survive both of its named checks, and a fixture with a live `-wal`
   hides a write. Fold: the fixture state reads "every handle closed and no `-wal` present before the
   first hash." Mutation 2 becomes "`check` calls the apply through a write-mode open", which the
   probe shows changes the main file once that open closes.
6. **R-6c10's connection claim and the older-store rule contradict the tree and each other.**
   Location: plan `:447-454` and `:877`. `readFacts` opens no connection: `lastSuccesses` opens and
   closes its own (`inspect.go:262-271`). R-6c10 reads settings for an older store whose version
   holds the table, while `doctor`'s table prints "store not checked" for every older store. The
   closed state set also includes restore in progress (lane 2), not a store, refused, and newer,
   which "unreadable" leaves unnamed. Fold: open one inspect connection in `readFacts` for both
   reads, or drop "open no second connection." Rule "older with the settings table" one way in both
   places. Enumerate the eight `store.State` values in the `doctor` row.
7. **The evaluation's row type has no home.** Location: plan `:438-440` and `:449-454`. `store`
   cannot import `settings`, and Task 3 lists no `store` file. So the rows `check` and `doctor` get
   from `Facts` need a type in `store`, exported only at Task 5, while Task 3's exported evaluation
   takes rows. Fold: Task 3 defines the evaluation's input row in `internal/settings`. Task 5 adds a
   raw `store` row type to `Facts`, and `internal/settings` or `cmd` converts. Name which.
8. **Criterion 5's failing case is ambiguous.** Location: plan `:753-757`. "Retention unchanged" read
   as absent from the env removes the `env-seed` row (the table's `env-seed`/absent row). The
   post-seed retention is then the test default, and the case passes or fails on that default. Fold:
   the env carries retention `48h`, equal to its stored seed, and the test-registered retention
   default is below `72h`.
9. **States with no named report.** Location: plan `:599-614`, `:667-682`, and `:730-747`. These
   are missing: a stored value that does not parse under its kind (the accessor, the evaluation, and
   `doctor`'s row); a missing generation row; two refusals in one evaluation, the "many bad" case
   `errors.Join` exists for; and `Load` with two unparseable setting values. Fold: one row each.
10. **The guard's refusal text turns false.** Location: plan `:826-829`.
    `guard.go` says "fewer than half the catalog's %d tracks" (`guard.go:87`), and `job_test.go:547` asserts it. At
    fraction `0.9` that text is wrong. Fold: Task 4 names the new wording, derived from the fraction,
    and updates the assertion.
11. **"The existing 6b-i tests pass unchanged" cannot hold literally.** Location: plan `:830`.
    `guard`'s signature gains the fraction (`guard_test.go:140` calls it). Every `local.New` caller
    must pass the new required accessor, unless a nil accessor falls back to the defaults, a silent
    default the plan forbids. Fold: "pass with only their construction updated."
12. **The validator's value type is open.** Location: plan `:316-317` and `:690-691`. One `Setting`
    struct across four kinds needs `func(any) error`, `func(string) error`, or a generic type. The
    refusal must also name "the bound", which an opaque function cannot print. Fold: declare the
    bounds as data parsed under the kind (a floor and a ceiling as strings, with NaN and both
    infinities always refused for a float). Registration builds the validator from them, and
    `doctor` and the refusals print the bound from the same data.
13. **The plan never names two reads it depends on.** Location: plan `:327` and `:869-871`, then
    `:854-856`. `doctor`'s "env `<v>` seeds" must render the value from the evaluation's outcome.
    `doctor.go`'s own `lookupEnviron` would be a second reader of a store-homed key's env value,
    breaking "the seed rule is the only reader" (spec `:734-735`). Next, `backup` evaluates "from its
    write-mode handle", which needs every settings row, orphans included. Outside `Update`, only lane
    2's new `Handle.Query` (`conn.go`, `30a66ac`) returns more than one row. Fold: name the outcome
    as `doctor`'s source for the env value. Name `Handle.Query` as Task 5's read, and add a Task 0
    item confirming it merged.

## Planning-miss items, per task

- **Quoted, never recalled.** Tasks 1 and 2 are covered by the probes and quotes. The gaps are the
  constants' file (minor 2), the dead-branch claim (minor 1), and the compile-time claim (minor 3).
- **Every state names its report.** The gaps are minors 6 and 9, plus M2's backup-key row.
- **Every privileged-field channel is named.** The table at `:280-290` is complete except that exit
  gate 12 proves too little (minor 4). A store-homed key's env value has a second potential reader
  in `doctor` (minor 13).
- **Every proof names its fixture state.** The gaps are criterion 5 (minor 8), criterion 14's WAL
  state (minor 5), and criterion 13's local keep, which cannot be built (M1).

## Build order

Tasks 1 and 2 are independent by files and build on master plus the lane 2 merge. Task 3 builds on
both through `QueryRow` and `Update`. Tasks 4 and 5 build once M2 names the declarations' route and
minor 13 names `Handle.Query`. Lane 2 still had warm uncommitted edits at `30a66ac`
(`internal/engine/store/restore.go` and `internal/backup/restore.go`). Task 0's blocker 2 covers
re-resolving `internal/backup` after they land.
