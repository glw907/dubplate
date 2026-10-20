# Rung 6c plan fold verification (revision 2)

Target: `docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md` at `ffde470`
(revision 2), against revision 1 (`8d0eed8`), the fold record
(`2026-09-25-rung-6c-plan-fold.md`), and the three lens reviews. Code read on master (`623321c`) and on
the read-only branch `lane-2-6b-ii-b`, whose head is now `cd9afbf` (6b-ii-b Task 5B), one commit past
the `30a66ac` the fold read. The reader had no part in the fold.

**Counts: 0 blockers, 2 majors, 7 minors.** Every blocker and major the reviews raised closed at its
cited location. The four partial refusals are sound. The two majors are new, and both come from the
X-M2 fold's choice to inject reader functions through `local.Config`: one callsite the wall forbids
editing, and one criterion 13 proof whose fixture sits in a package that cannot reach the production
declarations. Each takes a sentence or two to fix. All three rulings can be ruled tonight.

---

## 1. Did each blocker and major close?

| Finding | Closed? | Where, verified |
| --- | --- | --- |
| R-B1 | Yes | R-6c13 (`:660-673`) names all three zero-value consequences, and the lines match the branch: `guard.go:86`, `prune.go:123`, `prune.go:154`. `Job.Run` validates before `LookPath` (`backup.go:69` on the branch, inside `Run` at `:63`). The fold also carries the Task 4 helper (`:1029-1030`), the state row (`:1052`), mutation 3 (`:1080`), "only their construction updated" (`:1071-1072`), and review focus item 8. |
| C-M1, X-M1, R-M3 | Yes | R-6c9's pre-seed read (`:583-589`), the table-absent marker in R-6c10 (`:600-608`), D10 (`:690`), Task 5's facts table (`:1113-1120`), the local-keep golden (`:1144-1145`), and the 6d handoff (`:1204-1206`). `Inspect` gates `readFacts` on `c.version >= jobsVersion` and `<= len(files)` (`inspect.go`, branch), so older-store facts exist as the plan says. |
| C-M2 | Yes | Criterion 5 (`:973-979`) and mutations 4 and 5 (`:1002-1005`). Worked through: incremental validation with expiry first checks 72h against the stored 48h and refuses the passing case, and the stored-rows mutant passes 48h/15m, so the failing case applies. Both mutants die. |
| C-M3 | Yes | R-6c7 (`:517-520`), the `seed_value` privileged row (`:366`), the fixture (`:509`), and the Task 3 admin rows (`:952-953`). Mutation 7 kills both admin rows, since the env equals `seed_value` in one and `value` in the other. |
| C-M4, C-M5, R-M4 | Yes | R-6c8 (`:542-547`). `Handle.Query` exists on the branch (`conn.go:115`), with the quoted doc verbatim. The generation-first argument is correct (see section 2). Task 0 blocker 7 (`:757-758`). |
| C-M6 | Yes, with minor 1 | Criterion 14 (`:1160-1164`), mutation 2 (`:1172-1173`), and exit gate 15 (`:1256-1257`). A probe confirms the mechanism (section 3). |
| C-M7 | Yes | Each per-key fixture discriminates, and each stored value sits in bounds under either answer to Ruling 3 (`:1057-1066`). The test's home is major 2. |
| C-M8 | Yes | Exit gate 12 (`:1247-1250`) covers `internal` and `cmd`, excludes tests, and includes `.sql` (so the `CHECK` hit is expected and triaged). The literal-write rule and the admin-count test (`:990-991`). |
| X-M2 | Yes, with major 1 | R-6c11's "How the values reach their consumers" (`:637-646`), the state row (`:1050`), and mutation 4 (`:1081-1082`). `start` calls `bind` after `insert` (`local.go:182-186`, branch), and the readers now move ahead of it. |
| R-M1 | Yes | R-6c5 (`:442-454`), D7, the Task 2 `Load` rows and mutation 2, the Task 3 parse rows, and the Task 5 "invalid" rows. |
| R-M2 | Yes | R-6c7's post-seed bound (`:524-530`), D8, the three split rows (`:955-958`), mutation 8, the doctor rows (`:1134-1135`), and the rung 14 handoff. |
| R-M5 | Yes | "No default changes" (`:249-250`), R-6c11 (`:615-620`), and "Not verified." |
| R-M6 | Became Ruling 3 | `:110-128`. Both answers are stated in R-6c11's table. |

**The four partial refusals.**

- **C-M6** (the older-store observable) is sound. `Inspect` takes the lock shared and non-blocking,
  and an exclusive holder reports `StateBusy`, so the held-lock variant cannot distinguish anything.
  `cmd` has no way to build an older store: `migrationSourceThrough` is an in-package test helper in
  `internal/engine/store`. The replacement grep is too narrow (minor 1).
- **R-M4** (a store read-transaction method) is sound. Generation-first ordering cannot leave a stale
  cache in place, and it adds no export.
- **R-m4** (a closed source type) is sound. The literal rule, the diff grep, and the admin-count test
  together close the channel.
- **X-m1** (a scan of migration bodies) is sound. The unique savepoint name removes accidental
  collisions, a deliberate evasion inside a reviewed in-repo migration is the reviewer's to catch, and
  the `commitMigration` branch still catches the `COMMIT; SAVEPOINT <name>;` shape.

---

## 2. Contradictions and build order

**Checked and consistent:**

- **R-6c13, D7 to D10.** They are consistent with the resolutions, the task rows, and the exit gate.
  The ceiling arithmetic checks out: the eight lines sum to 4.33M, 4.33/0.8 = 5.41, and the reserve
  total after Task 5 is 3.40M. The deliverable count is 21 (4 × 4 + 3 + 2).
- **The generation-then-rows read.** Suppose a write lands between the generation read (G1) and the
  row read. Then the cache holds newer rows under G1, and the next generation read reloads them. A
  write after the row read is also seen on the next read. The only thing that moves the generation
  backward is `restore`, and restore takes the lock exclusively. `OpenWrite` holds the lock shared
  for the handle's lifetime (`open.go:119-180`, branch), so no warm cache can survive a restore. That
  reasoning is sound.
- **Criterion 10's two handles.** `OpenWrite` takes the lock shared, so two opens on one state
  directory coexist. The fixture can be built.
- **The pre-migration read of `Inspect`'s raw rows.** It is sound for the stores it names. The
  `Facts` doc and code match R-6c10's older-store claim once `readFacts` grows the settings read. One
  edge case is unnamed (minor 6).
- **The apply's `tx.Query`.** It exists (`write.go:161`, returning the store's `*Rows`).

### Major 1: Task 4 must edit `internal/ops/opshttp/contract_test.go`, and the wall forbids it

- **Location:** plan `:222-223` (the wall: "writes nothing under ... `internal/ops/opshttp`") and
  `:1024-1030` (Task 4 Files).
- **Defect:** R-6c11 makes the reader functions required fields of `local.Config`. A nil reader would
  either panic at `start` or need a silent default, which R-6c13's reasoning forbids. `local.New` has
  two callers outside Task 4's Files, both in `internal/ops/opshttp/contract_test.go` (`:317` and
  `:825`, on the branch and on master). Both start backups (`r.StartBackup` at `:836`). Task 4 cannot
  go green without editing that file, and the global wall forbids the edit, so the Sonnet implementer
  stops and reports. The reviews' choice between "accessor plus declarations" and "injected readers"
  (X-M2) did not surface these callers.
- **Fix:** add `internal/ops/opshttp/contract_test.go` to Task 4's Files, limited to construction
  (pass the readers through a helper that returns R-6c11's defaults). Add a one-clause exception to
  the wall, and list the file in Task 0 item 2's report.

### Major 2: criterion 13's per-key proofs name fixtures that no single package can compose

- **Location:** plan `:1057-1066` and Task 4 mutation 1 (`:1077`).
- **Defect:** each per-key proof starts from "a stored `env-seed` row" and ends at a consumer
  fixture: five remote snapshots and an archive 200h past its move, both "in the stand-in."
  - The stored row reaches the consumer only through `runBackup`'s closures over the production
    declarations, and those live in package `main`.
  - The stand-in that can list snapshots and archive directories belongs to `internal/backup`'s
    tests (`useStandin`, `localRemote`). Those tests build the runner through `local.New` with plain
    functions (R-6c11: "its tests pass plain functions") and cannot import `main`.
  - `cmd/dubplate-server`'s stand-in records argv and exits 0 (`backup_test.go:32-79`), so no remote
    listing exists there.

  As a result, the retention and guard proofs can be built in `cmd` (job records and the no-call
  record are observable there), but the remote-keep and archive proofs cannot be built as one test.
  An implementer would either duplicate the wiring inside `internal/backup_test`, which leaves the
  production closures unproven, or stop. Separately, "five remote snapshots ... pruned to three" does
  not say whether the five include the run's own snapshot. Keep counts "in total, the run's own
  included" (`prune.go:18-20`), so the expected survivor set is ambiguous.
- **Fix:** split the criterion:
  - (a) In `cmd/dubplate-server`, one test per key shows that `runBackup`'s reader for that key
    returns a stored `env-seed` row's non-default value, and the default when no row exists. The
    guard's admin half and the retention prune can stay end to end in `cmd`.
  - (b) In `internal/backup` (and `ops/local` for retention), each consumer shows that it honors a
    non-default reader value over a fixture the default would keep.

  State the keep fixture as "five prior snapshots plus the run's own; three survive: the run's own
  and the two newest prior."

---

## 3. New mechanisms stated from memory

None load-bearing. Each new mechanism is quoted or tied to a task's proof:

- `Handle.Query`: quoted, and it matches the branch verbatim.
- `Inspect`'s lock and bytes: quoted, and it matches.
- The `OpenWrite` refusal: quoted. The text is at `open.go:196` on the branch (`:175` on master), and
  Task 0 re-resolves it.
- `lookupEnviron`: quoted, and it matches `doctor.go:397-410`.
- The constants file: corrected to the generated `lib/sqlite.go`.

Criterion 14 and mutation 2 depend on a WAL behavior that the plan cites only through the mechanics
review. A probe re-ran it here: `modernc.org/sqlite v1.59.0`, scratch module at
`/tmp/claude-1000/probe6c`, with the store's DSN shapes.

| Step | Main-file hash | `-wal` |
| --- | --- | --- |
| seeded, every handle closed | unchanged (baseline) | absent |
| `mode=ro`, `query_only` read (the inspect shape), closed | unchanged | 0 bytes |
| read-write open with a read only, still open, then closed | unchanged | 0 bytes, then absent |
| read-write open that writes, still open | unchanged | 4152 bytes |
| same, closed | **changed** | absent |

So criterion 14's two observables ("main hash unchanged" and "`-wal` absent or 0 bytes") catch the
write-mode apply whether or not the mutant closes its handle. The read-only inspect leaves exactly the
state the criterion admits. Mutation 2 is sound. The plan's "Behaviors" list could cite this row, and
the named mutation run is its in-gate form, so no finding is raised.

`tx.Query` is used without a quote. It exists as described, so this is not a finding.

---

## 4. Can Geoff rule the three rulings tonight?

Yes, all three, on the evidence given.

- **Ruling 1.** The arithmetic checks, and both lane facts have since moved in 6c's favor. The first
  is L2, the site-name grammar: it is now **committed on master at `75fde16`**, before the fold
  commit, so it is no longer warm. The second is 5d: its plan (`2026-09-25-rung-5d-...-close.md`) runs
  in lane 1 after 5c-iv and states `internal/config` untouched (`:225-235`). It treats `doctor.go` as
  shared (`:1035`), so 6c's hold on 5d's launch still applies as written. Minor 7 covers the stale
  wording.
- **Ruling 2.** Both branches are built out (D3, the `deploy.sh` handoff), and the refusing outcomes
  are enumerated.
- **Ruling 3.** The risk and the floors are concrete, and the default clears each floor. Criterion
  13's stored values sit at floors valid under either answer. The "no" branch has a small unbuilt
  piece in Task 2 (minor 4), which does not block the ruling.

---

## 5. The four planning-miss items on the changed tasks

| Task | Quoted behaviors | States with reports | Privileged channels | Fixture states |
| --- | --- | --- | --- | --- |
| 1 | ok (probe table) | minor 3 | n/a | ok |
| 2 | ok | minor 4 | ok | ok |
| 3 | ok | minor 2 | ok (admin literal, `seed_value`, generation, restore) | ok (criteria 5, 7, 10) |
| 4 | ok | ok (the backup-key row added) | ok | major 2 |
| 5 | ok (probe above) | minor 6 | ok (the env value from the outcome) | ok (criteria 8, 14; older-with-table holds a row) |

---

## Minors

1. **Exit gate 15 misses `doctor`'s store read** (`:1256-1257`). `doctor` inspects through
   `storeRow` in `cmd/dubplate-server/store.go:89-103`, not in `doctor.go`, and Task 5 edits
   `store.go`. A write-mode open added there passes the grep. **Fix:** add `store.go` to the
   pathspec. It has no `store.Open` call today.
2. **The accessor's behavior toward other keys' rows is unnamed** (R-6c8, `:549-553`; Task 3 "Plus",
   `:964-967`). The accessor "reloads every row," and a bad stored value "is an error naming the key."
   The plan never says whether a reload fails when a *different* key's row is unparseable or out of
   bounds, or when it meets an orphan row. An eager reload-time validation would fail every backup
   start over, say, a bad local-keep row that backup never reads. That is the failure the spec's
   "without failing" sentence exists to prevent. **Fix:** add one Task 3 row: "the accessor reads key
   A while B's stored row is unparseable or out of bounds, or an orphan exists → A's value, no error."
3. **The `COMMIT; SAVEPOINT <name>;` row names no post-state** (`:819`). Criterion `:830-831`
   requires every savepoint row to assert `user_version` and the ledger. In this shape, the version
   write and the ledger insert autocommit outside the runner before `Commit` fails. So the store
   *does* advance, unlike the `COMMIT; BEGIN;` row. An implementer copying that row's expectation
   writes a failing test. **Fix:** state that this row expects `user_version` advanced and the
   ledger row present (detection only), per the mechanics probe.
4. **Task 2's validator rows assume inclusive bounds** (`:887`). Under Ruling 3's "no," the float
   floor is exclusive (R-6c2 `:409-410`), so "at the floor → accepted" is wrong for a float, and
   Task 2 lists no exclusive-floor field or row. **Fix:** add one clause: on "no," the bounds data
   carries an exclusive-floor flag, and the float at-floor row expects "refused naming the bound."
5. **Exit gate 11's grep has no pathspec** (`:1244-1246`). A whole-tree `git grep` of each key's
   literal, excluding only `*_test.go`, returns every `docs/` hit (this plan, the spec, HISTORY).
   **Fix:** use the pathspec `-- cmd internal ':(exclude)*_test.go'`.
6. **A store at version 1 gets no facts, and the plan names no row for it** (`:1113-1120`,
   `:1139`). `Inspect` reads facts only at `version >= jobsVersion` (2). An older store at v1 reaches
   `doctor` with nil `Facts` and no table-absent marker, and it falls under neither "older without the
   settings table" (which R-6c9 keys on the marker) nor "not checked." **Fix:** one sentence. `doctor`
   treats `StateAbsent`, and `StateOlder` with nil `Facts`, as "store holds no settings yet" without
   calling the facts entry point. Or add the row.
7. **Stale lane facts** (`:144-152`, `:1381-1382`, Task 0 blocker 3). The plan says L2 is "warm,
   uncommitted" and "not verified." It landed at `75fde16` (`internal/config/config.go`, +42), before
   `ffde470`. This is harmless, since Task 0 reports it as a note, but Ruling 1's evidence reads
   weaker than it is. **Fix:** cite `75fde16` in "Lane condition, answered" and drop the "Not verified"
   bullet.

Considered and not raised: during the 6c-to-6d window, an out-of-bounds `env-seed` backup key
pending a heal shows as non-refusing in `doctor` while the accessor fails the backup start. No
production row can exist before 6d's seed, so this is left to 6d's handoff at the fold's discretion.
