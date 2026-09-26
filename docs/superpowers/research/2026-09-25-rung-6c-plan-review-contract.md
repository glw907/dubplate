# Rung 6c plan review: contract and criteria

Reviewed: `docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md` at `8d0eed8`
(revision 1), through one lens: contract and criteria. The inputs were the ratified rung 6 spec's 6c
section (`2026-09-23-rung-6-ops-seam-store.md:1062-1084`), "The settings home" (`:718-800`), the schema
sections (`:433-508`), and the configuration homes spec (CH). Code pointers were checked on master:
`internal/engine/store/open.go`, `write.go`, `conn.go`, and `invariant_test.go`, `internal/config`,
`internal/backup/guard.go`, `prune.go`, and `internal/ops/local/jobs.go`.

Findings are ranked by consequence. Polish is omitted.

**Counts:** 0 blockers, 8 majors, 10 minors, no owner forks beyond the plan's own two rulings.

## Verdict

All 14 of the spec's 6c criteria map to a task, and the exit gate's table is correct (criterion 1-3
to Task 1, 4 to Tasks 2 and 3, 5-7 and 9-10 to Task 3, 8, 12, and 14 to Task 5, 11 to Task 2, 13 to
Task 4). The four planning-miss items are mostly met, and the stdlib and driver quoting is
unusually thorough. The majors fall into three groups:

- **One design claim contradicts the store's contract.** The pre-seed read of the local snapshot
  keep cannot go through the accessor.
- **Four proofs can pass vacuously.** These are criterion 5's failing case and its mutation,
  criterion 13's per-key fixtures, criterion 14's byte comparison, and exit gate 12's grep.
- **Three contract gaps leave the implementer to guess.** These are which column an admin-row
  comparison reads, the order of the generation and row reads, and the store's missing multi-row
  read.

Each has a fold of a few lines. None needs a new ruling.

## Deviations D1 to D6

| # | Holds? | Note |
| --- | --- | --- |
| D1 | Yes | `checkEntityIDs` (`invariant_test.go:391-412`) applies the "no `INTEGER PRIMARY KEY` named `id`/`*_id`" half to every table, exempt or not, so the plan's claim at `:404-406` is true of the code. A one-to-one extension keyed by the parent's `BLOB` id is a "keyed table" in the spec's own sense (`:455-456`). |
| D2 | Yes | CH names only `collapse-duplicates` as "named now so the rung 6 schema carries them" (CH, "Rules that cut across the verbs"). The other preferences are consumer-routed under the 2026-09-12 refinement. The four overrides match CH's "Two surfaces, one API" list. CH's "overrides in the settings table" was already superseded by the spec's separate `member_overrides` (`:495`), so the plan follows the ratified text. |
| D3 | Yes, if ruled | This is a legitimate owner fork, because it changes a ratified sentence. The recommendation is sound. See minor m10 for the 6d handoff it owes. |
| D4 | Yes | The spec's purpose is no concurrent `internal/config` edit. Holding 5d satisfies it. |
| D5 | Partly | The declaration holds. The claim that criterion 13's local-keep half is met "through the accessor" does not (major M1). |
| D6 | Yes, with a gap | A test pair is the only way to prove criterion 5 at 6c. No 6d criterion re-proves it over the real nonce-retention and link-expiry pair (minor m2). |
| missing | | R-6c2 and R-6c3 are readings of spec `:728-736` that are not in the table (minor m6). |

## Majors

### M1. The local snapshot keep's pre-seed read cannot use the accessor

- **Location:** plan `:444-445` (R-6c9, "The pre-seed read ... is the accessor itself"), and
  `:821-825` (Task 4, criterion 13's local-keep proof).
- **Defect:** the accessor is bound to a `*store.Handle`. `OpenWrite` refuses "a store at any
  version but the binary's" (`open.go:106-113`), and `OpenServe` migrates before it returns. So no
  handle exists on a store with pending migrations. That is exactly the store the migration runner's
  `Keep` must be read from (spec `:748-749`: "reads the stored row or the registry default"). Task 4's
  test ("the accessor's value passed as `ServeOptions.Keep` to `store.OpenServe` over a store with
  pending migrations") can't be written as stated. Either it stops, or it reads the row by some
  other path that R-6c9 doesn't sanction. The claim also misleads 6d, whose handoff (`:927`) wires
  this value into `serve`.
- **Fold:** amend R-6c9. The pre-seed read resolves from inspect mode's raw rows (R-6c10), which
  `Inspect` returns for older stores, through one `internal/settings` resolution over raw rows and
  the registry: the stored value, or the default when no row exists or the store predates `0004`.
  `doctor` is its production caller. Task 4's local-keep proof becomes the following fixture. A
  store at version 4 carries an `env-seed` row of `1` and three pre-migration copies. The resolution
  over its `Inspect` facts gives 1, and `OpenServe` with that `Keep` prunes the copies to one. A
  store at version 3 resolves the default. Record the change in D5.

### M2. Criterion 5's failing case is ambiguous, and mutation 4 can survive

- **Location:** plan `:753-757`, `:775-776`.
- **Defect:** "env expiry `72h` with retention unchanged" doesn't say whether the env holds
  retention `48h` or omits it. If it is omitted, the seed table's `env-seed | absent` row removes
  the retention row (ruling 13), and the post-seed retention becomes the test-registered default. If
  that default is at least 72h, the case applies and the test fails for the wrong reason. If the
  default is under 72h, the case refuses for the wrong reason. Mutation 4 is also ill-defined:
  - Evaluating the rule over the pre-seed stored values (48h, 15m) passes both cases' rule check.
    It breaks the failing case, not the passing one.
  - The realistic bug is incremental per-key validation. It breaks the passing case only when
    expiry is applied before retention. If retention is declared first, the mutant survives.
- **Fold:** state the failing case's env as expiry `72h` and retention `48h`, both present. Give the
  test-registered settings defaults that satisfy the rule alone. Register expiry before retention.
  Replace mutation 4 with two mutations:
  - (a) Evaluate the rule over the stored rows, not the post-seed set, which makes the failing case
    fail.
  - (b) Validate after each key's seed in declaration order, which makes the passing case fail,
    since expiry is registered first.

### M3. Which column an admin-row comparison reads is unspecified

- **Location:** plan `:365-367` (the `settings` columns: value and seed value), `:412-419` (R-6c7),
  `:739-741` (the three `admin` rows).
- **Defect:** the row carries both `value` and `seed_value`, and R-6c7 never says what the apply
  writes into `seed_value` or which column "equal" compares against. CH is explicit: "If the row is
  `admin` and the env value differs, the load fails". The admin row's value is the one compared,
  and `seed_value` is the env seed the admin overrode. The `admin` fixtures don't set
  `value != seed_value`, so an implementation comparing against `seed_value` passes all three admin
  rows vacuously. That is also a privileged field with no named channel: nothing says who may set
  `seed_value`.
- **Fold:** add to R-6c7 that `env-seed` rows hold `value = seed_value`, the apply writes both, and
  every comparison reads `value`. `seed_value` on an `admin` row is the seed it replaced, set only by
  rung 14's write path. Add `seed_value` to the privileged-field table. Every `admin` fixture sets
  `value` and `seed_value` to different values, and the env equals `seed_value` in the "different"
  row, so a mutant comparing `seed_value` fails.

### M4. The generation read order is not pinned, so a stale cache can be silent

- **Location:** plan `:423-429` (R-6c8).
- **Defect:** the generation and the row are two reads on a pooled read connection with no shared
  snapshot. Consider an accessor that loads the row and then records a generation read after the
  load. A write that lands between the two reads leaves it caching (new generation, old value)
  forever, until the next write. Criterion 10's sequential test can't catch this race, so a
  test-green implementation can carry it. That staleness is the exact failure CH forbids.
- **Fold:** state that the accessor stores the generation read before its reload, or reads the
  generation and the rows in one read transaction on one connection. Add this as a diff-reviewer
  checklist line in Task 3. A test can't reliably pin it.

### M5. The store has no multi-row read, and the task wall omits one

- **Location:** plan `:178-183` (the territory wall: store edits only in Task 1 and in Task 5's
  inspect), `:440` and `:854` (`backup` evaluates "from its write-mode handle"), `:715-716`
  (Task 3's files).
- **Defect:** `Handle` exposes `QueryRow` on the read pool and nothing multi-row
  (`conn.go:108`; `Query` exists only on `*Tx`, `write.go:161`). The evaluation needs every
  `settings` row, orphans included. From backup's handle, the only multi-row path is `Tx.Query`
  inside `Update`, an `IMMEDIATE` write transaction taken for a read-only evaluation. It contends
  with the writer and contradicts R-6c9's "without writing" in spirit. Task 3 or Task 5 will meet
  this and must stop under `:53-54`.
- **Fold:** Task 3 adds a read-pool `Handle.Query` in `conn.go`, following `QueryRow`'s shape,
  listed in its Files and in the wall, with the exported-caller rule met by `internal/settings`.
  The alternative is to have `backup` evaluate from an `Inspect` call as `check` does. The first
  fold is simpler.

### M6. Criterion 14's byte comparison and mutation 2 can pass vacuously

- **Location:** plan `:892-894`, `:902-903`.
- **Defect:** a write-mode open that performs no write leaves the main database file's bytes
  unchanged in WAL mode, so mutation 2 ("read through a write-mode open") likely survives the
  SHA-256 comparison. The plan hedges with "report which". An apply's write also lands in `-wal`
  and reaches the main file only at checkpoint. If any handle stays open when the hash is taken,
  even "check calls the apply" survives.
- **Fold:**
  - Name the load-bearing mutation as "`check` calls the apply," with the fixture state "a seeded
    row and a divergent env, so an apply would write" already stated.
  - Hash the main file and the `-wal` (presence, size, and bytes), taken after `check` returns.
  - Prove the write-mode mutation with an observable that differs. `check` succeeds against a store
    at version 4 (an older store), which `OpenWrite` refuses. Or it succeeds under a held
    exclusive `store.lock`, if inspect takes no lock (quote `Inspect`'s doc for which applies).

### M7. Criterion 13's per-key fixtures don't discriminate the default from the stored value

- **Location:** plan `:821-825`.
- **Defect:** "a stored `env-seed` row with a non-default value reaches the key's consumer" names no
  fixture state. Two examples show the gap:
  - Retention: with no finished record aged between the stored value and 90 days, the prune looks
    identical under both values.
  - Remote keep: with fewer snapshots than both values, nothing is pruned.
  So each per-key proof can pass with the consumer still reading the default. The guard's admin
  half does this right (`:826-829`).
- **Fold:** each key's fixture holds an item the default keeps and the stored value removes:
  - retention: stored `48h`, and a finished record aged 72h;
  - remote keep: stored `3`, and five remote snapshots in the stand-in;
  - archive retention: stored `48h`, and an archive 72h past its move;
  - guard: stored `0.9`, with the existing 10-track and 6-file tree;
  - local keep: per M1.

### M8. Exit gate 12's grep can't see the channel it guards

- **Location:** plan `:957-958`, and the privileged-field row at `:284`.
- **Defect:** `git grep "'admin'" -- 'internal/**/*.go'` matches only a SQL single-quoted literal.
  A Go constant `"admin"` bound as a `?` parameter, which is the idiomatic form, is invisible. So is
  anything under `cmd/`. The expected output ("shows only the schema's `CHECK`") is also wrong: the
  `CHECK` lives in a `.sql` file the pathspec excludes. The gate passes vacuously against the one
  violation it exists to catch.
- **Fold:** grep `-e "'admin'" -e '"admin"'` over `internal/` and `cmd/`, excluding tests, and paste
  every hit with a one-line reason: read paths, `doctor`'s rendering. Add a Task 3 test that runs
  the apply over every seed-table state and asserts that no row with source `admin` exists unless
  the fixture wrote it.

## Minors

- **m1. The seed table misses the empty, many-bad, and stored-unparseable states** (`:730-747`).
  The fold adds three rows:
  - An env set to the empty string beside an `env-seed` row behaves as absent: the row is removed,
    under the wall's empty-equals-unset rule. That surprise deserves a proof.
  - Two keys that refuse for different reasons in one seed set: both are named in the joined error,
    and nothing is written.
  - A stored value that no longer parses under its kind (kind changed, hand-edited row): the
    evaluation refuses naming the key, and `doctor` renders it.
- **m2. D6 and the handoff owe 6d a re-proof** (`:501`, `:926-929`). The fold adds "6d re-runs
  criterion 5 over the real nonce-retention and link-expiry pair" to D6 and to the 6d handoff. No
  6d criterion (spec `:1102-1105`) covers it.
- **m3. Exit gate 13 misses two exported values with no caller** (`:200-204`, `:959-961`). All five
  keys are "next operation", so the apply-mode type's "immediately" and "restart" values have no
  production use at 6c, and the staged exported-unused leg warns on them. The fold names them as
  exceptions beside the sensitive marker.
- **m4. `doctor` contradicts R-6c10 on an older store** (`:877`, `:449-452`). R-6c10 returns
  settings rows for any store at version 4 or later. The doctor table renders "store not checked"
  for every older store. The fold: evaluate whenever `Facts` carries settings rows, and render "not
  checked" otherwise.
- **m5. `doctor`'s text is false in the window between 6c and 6d** (`:869-872`). R-6c1 holds 5d
  until 6c merges, and 5d is the first deploy, so a deployed binary may carry 6c without 6d. In that
  window, `Load` accepts the five env keys (R-6c5), `serve` seeds nothing, and "seeds at serve's
  next start" is untrue. The template carries none of the keys, so the risk is small. The fold
  records the window in Task 6's handoff to 5d and 6d.
- **m6. R-6c2's exported raw reader weakens criterion 11, and the reading is unrecorded**
  (`:325-328`, `:686-689`, `:330-336`). The spec says "the compiler forbids a consumer bypass"
  (`:735-736`). An exported function that returns a setting's raw env string is a bypass the
  compiler permits, checked once by exit gate 11's grep and never again. The fold has two parts:
  - Let `internal/settings` read `environ` itself, one lookup under the documented
    empty-equals-unset rule. No `internal/config` export is then needed.
  - Otherwise, state the weaker guarantee in a D7 row. Put R-6c3 (wall keys carry no apply mode or
    marker until hook 18) in the same row.
- **m7. Task 2's criterion 4 test proves nothing about the environment** (`:690-691`). "A test
  calls it with no environment in scope" can't detect an `os.Getenv` inside the validator. The fold:
  `t.Setenv` sets the key to an out-of-bounds value, and the test asserts the validator's verdict on
  an in-bounds value is unchanged.
- **m8. A behavior is recalled, not quoted** (`:374-375`, `:385`). The per-member tables rely on a
  `BLOB PRIMARY KEY` refusing `NULL`. SQLite's historic quirk admits `NULL` in a non-integer primary
  key of an ordinary table. Whether `STRICT` changes that is recalled here, not quoted. The fold
  declares `NOT NULL` explicitly, or quotes sqlite.org/stricttables.html and adds a state row for a
  `NULL` `user_id`.
- **m9. `doctor`'s test seam for the sensitive marker is unnamed** (`:881-882`). The marker golden
  needs a test-registered setting, but `doctor` builds from `newProductionRegistry`. The fold states
  that the rendering function takes the registry as a parameter, so the golden passes one without a
  test-only hook in production code.
- **m10. A "yes" on Ruling 2 owes 6d a handoff line** (`:95-99`, `:926`). If Ruling 2 is yes, 6d's
  `deploy.sh` must halt on `check`'s nonzero exit before restarting `serve` (spec 6d criterion 14).
  The fold adds this to Task 6's 6d handoff, so the reason the ruling exists is carried by the pass
  that relies on it.

## Criteria coverage, for the record

| Spec 6c criterion | Plan site | Testable, fixture named, can't pass vacuously? |
| --- | --- | --- |
| 1 invariants, string bind, orphan override | Task 1 `:618-621` | Yes. The `move_intents` path `CHECK` from 6a accepts a well-shaped TEXT, so 3091 is the only fault. |
| 2 visibility | Task 1 `:608-609` | Yes (existing constraint). |
| 3 `at`/`kind` `NOT NULL` | Task 1 `:610`, `:622-623` | Yes. |
| 4 seed outside bounds | Tasks 2 and 3 `:690`, `:751-752` | Yes, with m7. |
| 5 correlated rule | Task 3 `:753-757` | No, see M2. |
| 6 seed table rows | Task 3 `:730-758` | Yes, with M3 and m1. |
| 7 default change reaches store | Task 3 `:759-761` | Yes. |
| 8 `backup` divergent env | Task 5 `:886-890` | Yes (a row dump, robust). |
| 9 audit on seed, none on refusal | Task 3 `:762-763` | Yes. |
| 10 cross-handle invalidation | Task 3 `:764-767` | Yes for the sequential case. `OpenWrite` takes a shared flock, so two handles coexist. The race is M4. |
| 11 no wall getter | Task 2 `:686-689` | Partly, see m6. |
| 12 `doctor` goldens | Task 5 `:863-891` | Yes, with m4 and m9. |
| 13 five constants through the accessor | Task 4 `:821-829` | The admin half passes. The per-key half fails (M7), and the local-keep half is infeasible (M1). |
| 14 `check` read-only | Task 5 `:892-894` | No, see M6. |
