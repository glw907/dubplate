# Rung 6c plan fold record: revision 1 to revision 2

The plan is `docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md`. Revision 1
was `8d0eed8`, and revision 2 is the working-tree draft this record accompanies. The fold covers
three lens reviews, committed at `098f8e5`:

- `2026-09-25-rung-6c-plan-review-contract.md` (C): 0 blockers, 8 majors, 10 minors.
- `2026-09-25-rung-6c-plan-review-mechanics.md` (X): 0 blockers, 2 majors, 13 minors.
- `2026-09-25-rung-6c-plan-review-risk.md` (R): 1 blocker, 6 majors, 7 minors.

Every finding was checked against the tree before it was acted on. The checks covered master at
`098f8e5` and the read-only branch `lane-2-6b-ii-b` at `30a66ac`, which merges before 6c runs. The
authorities were the rung 6 spec (`2026-09-23-rung-6-ops-seam-store.md`), the configuration homes
spec (CH), and the repo `CLAUDE.md`.

## Counts

There are 47 finding IDs. They fold into 31 distinct defects, because 16 IDs duplicate another
lens.

| Disposition | Finding IDs |
| --- | --- |
| Folded | 45 |
| Refused | 1 (C-m7) |
| Owner fork | 1 (R-M6, now Ruling 3) |

Four folded findings carry a **partial refusal**, where one proposed mechanism was declined and the
defect was still fixed. They are C-M6, R-M4, R-m4, and X-m1, each noted in its row.

**Why so little was refused.** Anthropic's guidance warns that a gap-hunting reviewer reports gaps
even in sound work. Here, though, every major and minor was a real defect on inspection, not a
matter of taste:

- a mechanism that cannot be built (no handle before a migration);
- a proof that passes vacuously (the WAL hash, the grep, criterion 5's ambiguous env);
- a state with no named report;
- a claim the tree contradicts (the constants file, "open no second connection," "tests pass
  unchanged").

Most folds cost one state row or one sentence. Refusal pressure went to mechanisms instead. Each
convergent defect got its cheapest root fix, and four heavier proposals were declined in part: a new
store read-transaction method, a closed source type, a migration-body scan, and an older-store
observable that `cmd` cannot build.

## Verification notes that shaped the fold

- **`Handle.Query` exists on lane 2** (`conn.go:115` at `30a66ac`), a read-pool multi-row read. That
  settles the multi-row cluster with no store edit in 6c. Task 0 gains blocker 7 to confirm it merged.
- **`Inspect` takes the lock shared and non-blocking**, and an exclusive holder reports `StateBusy`
  (`inspect.go:157-165`). So C-M6's "succeeds under a held exclusive lock" observable does not
  exist.
- **Only `internal/engine/store`'s in-package tests can build an older store**
  (`migrationSourceThrough`; `inspect_test.go:44` does it today). Inspect's older-store states are
  provable there. `cmd` goldens build a `store.Report` by hand.
- **The site-name grammar is live**: at `098f8e5` the main tree carries warm, uncommitted
  `requireSiteName` edits to `internal/config/config.go` and `config_test.go`. That confirms R-m1's
  premise. Task 0 blocker 3 already treats an uncommitted `internal/config` edit as a blocker.
- **B1's three zero-value consequences reproduce by reading `30a66ac`.** The first is
  `float64(files) < guardFraction*float64(tracks)` at `guard.go:86`. The second is
  `own.Sub(stamp) <= archiveRetention` at `prune.go:154`. The third is `others[othersKeep:]` at
  `prune.go:123`.
- **`rejectNamespace`'s threshold branch** states that a threshold is "a declared name whose own
  Parse* call resolves and validates it at the point of use." `backup.go:122` returns `Load`'s error.
  That confirms R-M1.
- **The driver constants sit in `lib/sqlite.go:3214-3224`.** `lib/sqlite_linux_amd64.go` has no
  `SQLITE_CONSTRAINT` name (X-m2 confirmed).
- **`doctor.go`'s `lookupEnviron` already duplicates `wall.get`'s rule outside `config`.** That is
  the precedent for `internal/settings` reading the env value itself (C-m6).

## Dispositions

### Blocker

| IDs | Disposition | Where revision 2 carries it |
| --- | --- | --- |
| R-B1 (with X-m11) | **Folded.** A zero-valued `backup.Job` field disables the guard, purges the archive, or panics. `Job.Run` refuses a structurally invalid keep, fraction, or retention before any rclone launch, with a named code naming the field. It never supplies a default. | New R-6c13. Task 4: Files (the test helper), Outcomes, the `Job.Run` state row, named mutation 3, and the criterion "pass with only their construction updated." Review focus item 8. Deliverables 2 → 3. |

### Majors

| IDs | Disposition | Where revision 2 carries it |
| --- | --- | --- |
| C-M1, X-M1, R-M3 | **Folded at the root.** No handle, and so no accessor, exists before the migration the local keep caps. The pre-seed read is now the from-facts entry point's value in force, over `Inspect`'s raw rows. A store without the settings table resolves every key to its default ("store holds no settings yet"). The `OpenServe` prune proof moves to 6d, the first caller, because no package in 6c can build it: store tests cannot import `settings`, and `cmd` has no older store. | R-6c9 ("The pre-seed read"), R-6c10 (the table-absent marker), R-6c11 consumer column, D5, new D10, Task 4 criterion (the local keep's half moves to Task 5), Task 5 (the inspect facts table, the doctor local-keep golden), Task 6's 6d handoff. |
| C-M2, X-m8 | **Folded.** In criterion 5's failing case, both env values are present (expiry `72h`, retention `48h`). The test defaults satisfy the rule alone (`15m`, `48h`), and expiry is registered first. Mutation 4 is replaced by two: the rule evaluated over the stored rows, and incremental validation. | Task 3 criterion 5, named mutations 4 and 5. |
| C-M3 | **Folded.** An `env-seed` row holds `value = seed_value`, and every comparison reads `value` (CH `:100-101`). On an `admin` row, `seed_value` is the seed it replaced, set by rung 14. Admin fixtures set `value ≠ seed_value`, and the "different" env equals `seed_value`. | R-6c7, the privileged-fields `seed_value` row, the `0004` fixture, Task 3's admin rows, mutation 7, the reviewer line. |
| C-M4, C-M5, R-M4, X-m13 (its multi-row half) | **Folded, with a partial refusal.** Every reload reads through lane 2's `Handle.Query`, which `backup`'s evaluation also uses for its orphans. The generation is read first, and the cache is stored under that earlier generation. A race can then only cache newer rows under an older generation, which the next read reloads. **Refused in part:** R-M4's option (a), a new read-transaction method on the store. The generation-first order is correct without a shared snapshot, and it adds no export. | R-6c8, R-6c9 (the from-handle entry point), global constraints (no store read method), Task 0 blocker 7, Task 3 reviewer line, quoted behaviors (the `QueryRow` and `Query` docs). |
| C-M6, X-m5 | **Folded, with a partial refusal.** Criterion 14 hashes the main file and checks that `-wal` is absent or 0 bytes after `check` returns. Before the first hash, every seeding handle is closed and no `-wal` holds frames. The load-bearing mutation is "`check` calls the apply through a write-mode open," which the mechanics probe shows changes the main file once that open closes. **Refused in part:** C-M6's older-store observable. `cmd` cannot build an older store, and the held-exclusive-lock variant reports `StateBusy`. A cheaper structural check replaces it: exit gate 15 greps that `check.go` and `doctor.go` never call `store.Open*`. | Task 5 criterion 14, named mutation 2, exit gate 15. |
| C-M7 | **Folded.** Each per-key fixture holds an item the default keeps and the stored value removes. Retention is `48h` against a 72h record. Remote keep is `3` against five snapshots. Archive retention is `168h` against a 200h archive. The guard uses its admin half. The values sit at floors valid under either answer to Ruling 3. | Task 4 criterion 13. |
| C-M8, R-m4, X-m4 | **Folded, with a partial refusal.** The apply writes `'env-seed'` as a SQL literal and never binds `source`. Exit gate 12 greps the pass's diff over `internal/` and `cmd/` for added `'admin'` or `"admin"` and triages every hit; a whole-tree grep would drown in the existing `internal/auth` and `internal/ops` hits. Task 3 asserts the `admin` row count unchanged after every apply path. **Refused in part:** R-m4's closed unexported source type, which the literal rule and the count test already cover. | Privileged-fields `admin` row, R-6c9, Task 3 (the admin-source guard), exit gate 12. |
| X-M2 (and R-m7's backup-key row) | **Folded.** `ops/local` cannot name package `main`'s declarations. `runBackup` injects reader functions in `local.Config`, the spec's pattern for 6d's auth keys (`:209-210`), so `ops/local` imports no `settings`. Every read happens before `insert`, and the binder takes the three values as arguments. A new state row covers an out-of-bounds backup key (fails the start, no job record). | R-6c11 ("How the values reach their consumers"), global constraints (imports), exit gate 10, Task 4 Files, Outcomes, states, and mutation 4. |
| R-M1 | **Folded.** `Load` classifies a store-homed key as declared and never parses it, by today's threshold precedent. An unparseable env value is a refusing outcome of the evaluation, so `backup` logs it and runs. | R-6c5, new D7, Ruling 2's refusing outcomes, Task 2 (`Load` rows, mutation 2), Task 3 (the unparseable row), Task 5 (the "invalid" rows read from the outcome). |
| R-M2 | **Folded.** Bounds apply to the value in force after the seed (spec `:740-742`). An out-of-bounds `env-seed` row that the env re-seeds or removes is healed and audited. The row refuses only while the out-of-bounds value stays in force. Rung 14 gets a handoff line: a resolving channel that works while `serve` is down. | R-6c7, new D8, Task 3 (three split rows, mutation 8), Task 5 doctor rows, review focus item 9, Task 6's rung 14 handoff. |
| R-M5 | **Folded (as directed).** 6b-i's defaults carry unchanged. If research argues for a change, the task reports it for a ruling and never makes it. The only new default is the local keep's. | "What this pass is not," global constraints ("No default changes"), R-6c11, Task 4 Outcomes, "Not verified." |
| R-M6 | **Owner fork: Ruling 3.** Protective minimums of keep ≥ 3, archive retention ≥ 168h, and fraction ≥ 0.25. Recommendation: yes (option b). | "Rulings awaiting owner" item 3. R-6c11's bounds column carries both answers. R-6c2 notes that the "no" answer needs an exclusive float floor. |

### Minors

| IDs | Disposition | Where revision 2 carries it |
| --- | --- | --- |
| C-m1, X-m9, R-m7 (its other rows) | **Folded.** New state rows: an empty env as absent; two refusals joined; a stored value that does not parse; a missing generation row as an error naming the table. X-m9's "`Load` with two unparseable values" is moot under R-M1 and is covered by the joined-refusal row. R-m7's "restart"-mode line goes to the 6d handoff. | Task 3 states and "Plus," R-6c8, Task 6's 6d handoff. |
| C-m2 | **Folded.** 6d re-runs criterion 5 over the real nonce-retention and link-expiry pair. | R-6c4, D6, Task 6's 6d handoff. |
| C-m3 | **Folded.** The apply mode's "immediately" and "restart" values are named exceptions. | Global constraints, exit gate 13. |
| C-m4, X-m6, R-m6 | **Folded at the root.** Facts carry settings rows for any store whose version holds the table, older included. `doctor` evaluates whenever they do, and the eight `store.State` values are enumerated. `readFacts` shares one inspect connection between the jobs and settings reads, which corrects the false "open no second connection" claim. | R-6c10, Task 5 (the inspect facts table, doctor rows), Task 0 item 9. |
| C-m5 | **Folded.** A handoff line covers the 6c-to-6d window, where `doctor`'s "seeds at serve's next start" is untrue. | Task 6 handoff ("To 5d and 6d"). |
| C-m6, X-m3 | **Folded.** `internal/config` exports no raw env reader. `internal/settings` reads the environ slice itself on `lookupEnviron`'s precedent. The compile-time claim is narrowed to "no function accepts a `Setting`." Exit gate 11 greps each key's literal. R-6c2 and R-6c3 are recorded as reading D9. | R-6c2, R-6c3, new D9, privileged-fields env-value row, Task 2 criterion 11, exit gate 11. |
| C-m7 | **Refused.** Moot once bounds are data (X-m12). Registration builds the validator from declared strings, so no user-supplied function exists that could call `os.Getenv`. | n/a |
| C-m8 | **Folded.** Every `BLOB` primary key in `0005` declares `NOT NULL`, with a state row for a `NULL` `user_id` (1299). The `STRICT` behavior is marked "not quoted, not relied on." | R-6c6, quoted behaviors, Task 1 states, "Not verified." |
| C-m9 | **Folded.** The fold verified that `newDoctorCmd` already takes the registry (`doctor.go:28`), so the goldens pass a test registry. The plan now states it. | Task 5 Outcomes, Task 0 item 9. |
| C-m10 | **Folded.** On Ruling 2's yes, 6d's `deploy.sh` halts on `check`'s nonzero exit. | Ruling 2 "what each answer builds," Task 6's 6d handoff. |
| X-m1 | **Folded, with a partial refusal.** The savepoint gets a name no migration writes. `commitMigration`'s branch stays, citing the probe row it still catches. Both new probe rows join the table. Assertions match the file name and the `RELEASE` failure. **Refused in part:** the body scan for the savepoint name. The unique name removes an accidental collision, and a deliberate evasion inside a reviewed in-repo migration is the reviewer's to catch. | R-6c12, the savepoint probe table, Task 1 states and outcomes. |
| X-m2 | **Folded.** The constants citation is corrected to the generated `lib/sqlite.go:3214-3224`. | Quoted behaviors. |
| X-m7 | **Folded.** The store exports its own raw row type in `Facts` (Task 5). `internal/settings` owns the core's row type and the conversion. There are two exported entry points: from-handle (`backup`) and from-facts (`check`, `doctor`). | R-6c9, Task 3 Files, Task 5 Files and deliverable. |
| X-m10, R-m5 | **Folded.** The guard's refusal text names the fraction in use, and `job_test.go:547` changes with it. | R-6c13, Task 4 Outcomes and state row, Task 0 blocker 2. |
| X-m12 | **Folded.** Bounds are declared as data (floor and ceiling strings parsed under the kind; NaN and infinities always refused). Registration builds the validator, and every refusal and `doctor` row prints the bound from the same data. | R-6c2, Task 2 Files, Outcomes, states, and criterion 4. |
| X-m13 (its doctor half) | **Folded.** `doctor` renders the env value from the evaluation's outcome, never through `lookupEnviron`. | R-6c9, privileged-fields env-value row, Task 5 Outcomes. |
| R-m1 | **Folded.** Blocker 3 reads the 5c-iii security fold's L2 disposition. If L2 is routed into 5c-iv, the conductor lands it before the cut, or adds the contested row and holds 5c-iv's edit. | "Lane condition, answered," Task 0 blocker 3, contested surfaces. |
| R-m2 | **Folded.** The apply reads rows through `tx.Query` inside the callback. | R-6c9, Task 3 Outcomes and reviewer line. |
| R-m3 | **Folded.** A privileged-fields row for `restore` records the lock argument. The 6d handoff notes that the seed may refuse on a restored `admin` row. HISTORY carries ruling 13's residual on data-loss keys. | Privileged fields, Task 6 (HISTORY and handoff). |

The review also named "the guard's 'fewer than half' text" as a convergence point. X-m10 and R-m5
cover it above. R-B1's third fold item ("pass with the fields set to the defaults") and X-m11 are the
same defect, folded with R-B1.

## Rulings as they now read

1. **Ceiling and lane.** Does 6c run under a 5.4M execution ceiling (flag 4.32M), in lane 2 once
   6b-ii-b merges? Recommendation: yes. It is re-derived because the fold added work.
2. **`doctor` and `check` exit code.** Do both exit nonzero when the seed rule would refuse
   `serve`'s start? The refusing outcomes now include a parse failure and an out-of-bounds value
   that stays in force. Recommendation: yes. On yes, D3's erratum is owed and 6d's `deploy.sh`
   halts on the exit.
3. **Protective minimums (new, from R-M6).** Do the backup keys take the floors keep ≥ 3, archive
   retention ≥ 168h, and guard fraction ≥ 0.25? Recommendation: yes. Yes builds those floors, which
   refuse a lower seed at 6d and are reported by `doctor` and `check` now. No keeps revision 1's
   floors (1, 24h, above 0) and needs an exclusive float floor.

No settled question was made a ruling. R-M1, R-M2, X-M2's route, and R-M4's order each had one
architecturally correct answer. Each was taken and recorded as a resolution or a reading.

## Ceiling arithmetic

The fold added +0.05M to Task 1, +0.05M to Task 3, +0.10M to Task 4, and +0.05M to Task 5. Task 2 is
net neutral.

- Projection: 0.10 + 0.50 + 0.55 + 0.70 + 0.55 + 0.60 + 0.40 + 0.93 = **4.33M**.
- Ceiling: 4.33 / 0.8 = 5.41, so the ceiling is **5.4M** with the flag at **4.32M**.
- Previous: 4.08M projected, a 5.1M ceiling.
- Ruling 1's "no" figure stays 4.5M (flag 3.6M). Cumulative spend with the reserve is now about
  3.40M after Task 5, so it crosses 3.6M during the simplifier round.

## Errata owed

This fold edits no ratified document. Task 6 drafts each erratum paste-ready and dated.

**New or changed by this fold, against the rung 6 spec:**

- **D7** (`:763`): an unparseable value is the evaluation's refusing outcome, not a `Load` error.
- **D8** (`:761`): the bounds row applies to the value in force after the seed.
- **D9** (`:734-736`, criterion 11): "the compiler forbids a consumer bypass" narrows to "no
  function accepts a `Setting`," plus the grep. Wall keys carry no metadata until hook 18.
- **D10** (criterion 13): the local keep resolves from `Inspect` facts, and its `OpenServe` proof
  moves to 6d.
- **D6** (criterion 5): extended with 6d's re-proof over the real pair.

**Changed against the 6b-i plan:**

- **D5** (R-6bi8): the local keep's 6c reader is `doctor` through the facts entry point.

**Carried from revision 1, unchanged:** D1, D2, and D4 against the rung 6 spec, and D3 on Ruling 2's
yes.

CH needs no erratum. R-6c7's reading matches CH `:95-101`.
