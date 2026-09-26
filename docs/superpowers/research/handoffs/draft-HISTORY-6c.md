# Draft HISTORY entry: rung 6, pass 6c (the routed schema and settings home)

Paste-ready for the conductor's ledger fold, once the merge ritual and the architecture reads
land. The bracketed fields need the merge's own commit ids; everything else is final. Model:
`draft-HISTORY-6b-i.md`'s format, following the top entry of `docs/HISTORY.md`.

---

## Rung 6, pass 6c (the routed schema and settings home): closed 2026-09-26, lane 2

Executed against
`docs/superpowers/plans/2026-09-25-rung-6c-routed-schema-settings-home.md`
(revision 2.2) in the lane 2 worktree (`.claude/worktrees/lane-2-6c`, branch `lane-2-6c`), cut from
master at `392d1cd`. Task 0's staleness read (`20bd27a`) found no blockers: five `local.New`
callers, a stale `internal/config` package doc it noted for cleanup, and eleven moved plan
citations, none of which changed a task's scope.

Ten commits ran across two segments and a simplifier round
(`git log --oneline 392d1cd..4ad14e7 | wc -l` reports 10):

- **Segment A** (Tasks 0-3, run `wf_62218e14-e00`, about 2.9M): Task 0 (`20bd27a`) reported no
  blocker. Task 1 (`b028c0f`, then fix `9058b37`) landed the `0004_settings` and `0005_routed`
  migrations (R-6c6): `settings`, `settings_audit`, `settings_generation`, `member_overrides`,
  `member_preferences`, `member_quota`, `group_quota`, `device_credentials`, `move_intents`, each
  with its populated fixture, plus the migration runner's `SAVEPOINT`/`RELEASE` guard against a
  migration body that ends its own transaction (R-6c12), the fix asserting the savepoint-refusal
  state and dropping plan-line citations. Task 2 (`39b0cbe`, `claude-opus-5-5`) added the config
  registry's store-homed `Setting` family: its own declaration type distinct from `Threshold`
  (R-6c2), the apply-mode and sensitive-marker metadata (R-6c3), the correlated-rule list
  (R-6c4), and `Load`'s classification of a set store-homed key without parsing it (R-6c5). Task
  3 (`257617b`, `claude-opus-5-5`) added `internal/settings`: the generation-first accessor
  (R-6c8), the unexported evaluation core and its two exported entry points, from-handle and
  from-facts (R-6c9), and the audited apply inside one `IMMEDIATE` transaction, reading every
  settings row through `tx.Query` in the callback and writing `source` as the SQL literal
  `'env-seed'`.
- **Segment B** (Tasks 4, 5, the simplifier round, run `wf_a5a9ec89-e5e`, about 1.6M): Task 4
  (`5e13971`, then fix `2d3c965`) re-homed the five backup constants as store-homed settings
  (R-6c11): `DUBPLATE_JOB_RETENTION`, `DUBPLATE_BACKUP_GUARD_FRACTION`,
  `DUBPLATE_BACKUP_SNAPSHOT_KEEP`, `DUBPLATE_BACKUP_ARCHIVE_RETENTION`, and
  `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`, each carrying 6b-i's unchanged default, `backup.Job`'s
  structural check before any rclone launch (R-6c13), and the reader-function injection pattern
  into `local.Config`; the fix removed plan citations, fixed `pruneArchive`'s doc link, turned the
  reader-error test into a four-row table, and corrected the guard-fraction doc to the 0.25 floor
  (Ruling 3's yes). Task 5 (`6c4e9d1`, then `b76069d` noting the settings and backup-remote rows
  in `doctor`'s exit-status doc, then two fix rounds `25b0daf`) taught `backup`, `check`, and
  `doctor` to evaluate the seed rule (R-6c9's from-facts entry point, R-6c10's raw settings
  rows), each read-only; the first fix round addressed the diff-reviewer's findings on the seed
  rule evaluation, and a second `fix` verdict flagged one leftover `"(R-6c9)"` comment citation
  the first round missed. The conductor accepted the task and folded the removal into the
  simplifier round, noting that the gate's citation leg does not match `"R-6c\d"`-shaped patterns
  (only whole-word `rung [0-9]`, `Task [0-9]`, `(C[0-9]`, `Resolution R`, and similar), a gap for
  the comment-cleanup task. The `code-simplifier` round (`4ad14e7`) dropped a further plan
  citation and reflowed an over-long doc comment; accepted by review.

### The merge

6c merges to master at `<merge-sha, filled at the fold>`: `<rebase outcome>`, a green worktree
gate, `git merge --ff-only`, a green gate on master, `<merged-range diff-reviewer verdict, if the
rebase is not a fast-forward>`.

### The architecture reads

Six `go-architecture-reader` grades run at the merge, each with no plan in context:
`internal/engine/store`, `internal/config`, `internal/settings`, `internal/ops/local`,
`internal/backup`, and `cmd/dubplate-server`. `<findings and routing, filled at the fold>`.

### The gate evidence

`go test -count=1` per touched package: `internal/engine/store` **10.101s**, `internal/config`
**0.018s**, `internal/settings` **0.579s**, `internal/ops/local` **3.038s**, `internal/backup`
**26.899s**, `cmd/dubplate-server` **58.760s** (one combined `go test -count=1` invocation over
all six packages).

`CAIRN_GATE_LANE` unset (heavy lane) `cairn-run-gate 'bash scripts/check.sh'` at HEAD (`4ad14e7`):
first issue reported "gate process vanished without a status" (a lost run, not a red gate, per
`.claude/agent-memory/dubplate-implementer/feedback_gate-vanished-status.md`); a same-string
re-issue reported **gate exit: 0**, 645-line log. The gitleaks leg scanned about 115.4 MB in
53.5-59.1s across the two runs and found no leaks.

Two staged legs stayed at warning: **325** process citations in Go comments (unchanged from
before the pass; 6c added none) and **128** exported-unused `internal/` identifiers (up from 125
before the pass: `git list -deps` and a direct re-run of check.sh's own awk pipeline against the
base commit (`392d1cd`) and HEAD both confirm the counts). Net new: `AddCorrelatedRule`
(`internal/config/registry.go`), `ApplyImmediately` and `ApplyMode`
(`internal/config/settings.go`), and `SettingsRow` (`internal/engine/store/inspect.go`); one name
dropped from the base list, `ThresholdKind` (`internal/config/thresholds.go`), now referenced
outside its own file and tests. `ApplyImmediately` and `ApplyMode`'s "restart" value are named
exit gate 13 exceptions (their first production caller is 6d); `SettingsRow` and
`AddCorrelatedRule` are 6c's own store-homed and correlated-rule surface, exercised only by
`internal/settings`' and `internal/config`'s own tests until 6d's `serve` and rung 14's admin
write path consume them. Both counts are pasted above from `bash scripts/check.sh`'s own
reporting lines.

### The three rulings and their outcomes

All three ruled by Geoff (plan `:100-220`):

1. **Yes.** A 5.4M execution ceiling, planning recorded separately, 4.32M flag. The pass
   re-derived about 5.0M at segment A's boundary, crossing the 4.32M flag, and continued under
   the standing overnight grant.
2. **Yes.** `doctor` and `check` exit nonzero on a refusing seed outcome (D3). Erratum D3 is
   owed; 6d's `deploy.sh` halts on `check`'s nonzero exit before restarting `serve` (Task 6's
   handoff).
3. **Yes.** The protective minimums keep the bound floor above the code-level invariant floor:
   snapshot keep at least 3 (not 1), archive retention at least 168h (not 24h), guard fraction at
   least 0.25 (not "above 0"). Geoff asked for the best and most idiomatic design on both this and
   Ruling 2, and the conductor took both as stated.

### The resolutions, R-6c1 through R-6c13

Each is stated in full in the plan at its own `### R-6cN` heading; this table names where each
landed at HEAD (`4ad14e7`).

| Resolution | What it decides | Landed |
| --- | --- | --- |
| R-6c1 | The lane placement: 6c runs in lane 2 before 5d, since no two lanes edit `internal/config` concurrently | Confirmed: Task 0 found no concurrent edit since the cut |
| R-6c2 | A store-homed key is its own declaration family (`Setting`), never a `Threshold` with a home field | `internal/config/settings.go` (Task 2, `39b0cbe`) |
| R-6c3 | Apply mode, the sensitive marker, and bounds live on `Setting` alone; the twenty wall thresholds keep their shape | `internal/config/settings.go` (Task 2, `39b0cbe`) |
| R-6c4 | Correlated rules are registry-level, proved with a test pair; the real nonce-retention/link-expiry pair moves at 6d | `internal/config` (declaration) and `internal/settings` (evaluation) (Tasks 2 `39b0cbe`, 3 `257617b`) |
| R-6c5 | `Load` classifies a set store-homed key without parsing it; the seed evaluation raises a parse failure where the value is used | `internal/config/config.go` (Task 2, `39b0cbe`) |
| R-6c6 | The routed schema's shape: `0004_settings` and `0005_routed`, invariant 6's exempt list gaining the per-member tables | `internal/engine/store/migrations/` (Task 1, `b028c0f`) |
| R-6c7 | What a settings row means: `value` is what every comparison reads, typed equality, bounds on the value in force, parse failures refuse | `internal/settings` (Task 3, `257617b`) |
| R-6c8 | The accessor: generation read before the rows, cache stored under the earlier generation, a setting read before `Handle.Update`'s callback | `internal/settings/accessor.go` (Task 3, `257617b`) |
| R-6c9 | Evaluation is exported (from-handle, from-facts); the apply is unexported until 6d | `internal/settings` (Task 3, `257617b`); consumed read-only by `backup`, `check`, `doctor` (Task 5, `6c4e9d1`) |
| R-6c10 | `Inspect` returns the raw settings rows and generation, with an explicit table-absent marker | `internal/engine/store/inspect.go` (Task 5, `6c4e9d1`) |
| R-6c11 | The five keys, their homes, defaults (unchanged from 6b-i), and bounds | `cmd/dubplate-server/store.go`, `backup.go` (Task 4, `5e13971`) |
| R-6c12 | The migration runner's savepoint, under a unique name no migration writes | `internal/engine/store/migrate.go` (Task 1, `b028c0f`) |
| R-6c13 | `backup.Job` refuses a structurally invalid value before any rclone launch and never defaults one | `internal/backup` (Task 4, `5e13971`) |

### The declared coverage gaps

**Four**, all named in the plan's "Not verified in this revision" (`:1435-1444`), none blocking
the exit gate:

1. 6b-ii-b's final shape of `internal/backup`, `Handle.Query`'s merge, and its handoff "to 6c"
   (Task 0's blockers 2 and 7, and item 8), resolved with no blocker found.
2. The defaults: 6b-i's values are carried unchanged; the local keep's 3 is a working value Task 4
   sources or keeps with its stated reason (Task 4 kept 3, with the rollback/two-copy reasoning
   R-6c11 states).
3. The savepoint and constraint probes ran outside the gate during plan mechanics review; Task 1's
   tests are their in-gate form.
4. Whether a `STRICT` table's `BLOB` primary key refuses `NULL` without an explicit `NOT NULL`:
   not relied on, since `0005` declares `NOT NULL` explicitly and Task 1 proves the constraint
   directly.

### What a later pass would be wrong to rediscover

- **The savepoint table** (R-6c12): `applyInTx` wraps each migration file's body in
  `SAVEPOINT dubplate_runner` / `RELEASE dubplate_runner`, so a body that ends its own transaction
  fails at `RELEASE`, named, before the version write and the ledger row; `commitMigration`'s
  "no transaction is active" branch still catches a body that re-declares the savepoint's name and
  passes `RELEASE` anyway.
- **Typed equality and value-not-seed-value** (R-6c7): two values are equal when they parse to the
  same typed value under the key's kind, so an env value of `60m` against a stored `1h` is equal
  to the seed and changes nothing; every comparison reads `value`, never `seed_value` (that column
  exists only so an `admin` row can show the seed it replaced).
- **Bounds on the value in force** (R-6c7): the closed table's out-of-bounds row refuses only when
  the out-of-bounds stored value stays in force, an `admin` row or an `env-seed` row the env
  confirms; an out-of-bounds `env-seed` row that an in-bounds env value re-seeds, or that an
  absent env removes, is healed and audited instead of refusing serve's start on a narrowed bound
  with no env value able to clear it.
- **The generation-first read** (R-6c8): each read-pool call is its own autocommit read with no
  shared snapshot, so reading the generation before the rows means a racing write can only leave
  newer rows cached under an older generation (reloaded next time), never older rows cached under
  a newer generation (which would serve stale data indefinitely). The wrong order is invisible to
  a sequential test; it is a diff-reviewer checklist line, not a test assertion.
- **The settings-read-before-the-write rule** (R-6c8): a consumer reads a setting before it enters
  `Handle.Update`, not inside the callback, since the spec's "every read a write decides on runs
  inside the callback" rule governs rows the write changes, and a setting read a moment before the
  write is the value in force for that operation (the apply mode "next operation" the five keys
  declare). No 6c key uses "restart" mode; the accessor's live-reload rule for that mode is 6d's.
- **The pre-seed read through the facts** (R-6c9): the migration runner's local-keep cap reads
  "the stored row or the registry default" before the migration it caps, and no `*store.Handle`
  exists on a store with pending migrations, so the read resolves through `Inspect`'s raw rows and
  the facts entry point, never through the accessor.
- **`backup.Job`'s structural check** (R-6c13): a guard fraction of 0, an archive retention of 0,
  or a keep of 0 each silently changes or crashes a run (the guard never trips, every archive but
  the run's own purges, `others[othersKeep:]` panics at -1); the check runs before any rclone
  launch, `LookPath` included, and is the code's own invariant, distinct from the registry's policy
  bounds.
- **Why the per-member tables are exempt** (D1): `member_overrides`, `member_preferences`, and
  `member_quota` are each a one-to-one extension of `users` keyed by the parent's `BLOB` id, the
  canonical relational shape; an extra opaque `id` on each would be a second identity for one row
  nothing references, and invariant 6's purpose (no identifier is a rowid) holds regardless.
- **Ruling 2's residual on data-loss keys**: when an env line naming one of the five keys is
  dropped, the key's value returns to its registry default at the next seed; if the default is
  lower than the value in force, the next backup prunes the difference. No 6c code holds a floor
  above the registry's own bound to prevent this; it is a property of the seed rule, not a defect.

### Both budgets

Tokens: **about 4.5M** in subagent tokens before the merge and the architecture reads (Task 0
included in segment A's about 2.9M; segment B's about 1.6M includes Tasks 4, 5, and the
simplifier round with its review), against the **5.4M ceiling** with the **4.32M flag** (Ruling
1). The pass crossed the flag at segment A's boundary (re-derived about 5.0M, under the 5.4M
ceiling) and continued under the standing overnight grant; no attended interaction was needed to
authorize it beyond that grant. `<the merge ritual's and the architecture reads' own spend, filled
at the fold>`. Fix rounds: Task 1 one, Task 4 one, Task 5 two, Tasks 0, 2, 3 none.

---

## The errata this pass owes

Drafted paste-ready for the conductor, numbered 6c-E1 to 6c-E10, from the plan's "Spec deviations
and readings that take effect on approval" table (`:715-724`). All dated 2026-09-26.

- **6c-E1 (D1, R-6c6).** Invariant 6's exempt list (spec `:455-458`) gains `member_overrides`,
  `member_preferences`, and `member_quota` beside the ledger's three: each is a one-to-one
  extension of `users` keyed by the parent's `BLOB` id, the canonical relational shape.
- **6c-E2 (D2, R-6c6).** `member_overrides` and `member_preferences`'s columns (spec `:495-496`)
  are CH's named four overrides and `collapse_duplicates`; the remaining preferences (theme,
  language, ReplayGain mode, sort order, notification choices, default collection) arrive with
  their own consumer's rung.
- **6c-E3 (D3, Ruling 2).** On Ruling 2's yes (spec `:743-745`, "`check` included"), `doctor` and
  `check` fail on a refusing seed outcome; a disagreement between the env and a stored value that
  the seed rule can still resolve fails nothing.
- **6c-E4 (D4, R-6c1).** The lane condition (spec `:1062`, `:1173-1174`) is read as "no concurrent
  edit to `internal/config`," satisfied here by sequencing 6c before 5d rather than by calendar
  order.
- **6c-E5 (D5, R-6c9, R-6c11).** 6b-i's R-6bi8 said 6d declares `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP`;
  6c declares it instead, read at 6c by `doctor` through the facts entry point, with 6d wiring it
  into `serve`.
- **6c-E6 (D6, R-6c4).** Criterion 5's "the correlated retention rule" is proved with a test pair
  shaped as the rule (retention at least the expiry); the real nonce-retention/link-expiry pair
  moves at 6d, which re-runs criterion 5 over it.
- **6c-E7 (D7, R-6c5).** Spec `:763`'s "any | unparseable | a wall error, as today" reads as
  `Load` classifying a store-homed key without parsing it; an unparseable env or stored value is a
  refusing outcome of the seed evaluation, raised where the value is used, not eagerly in `Load`.
- **6c-E8 (D8, R-6c7).** Spec `:761`'s "any | stored value outside the binary's bounds" applies the
  bound to the value in force after the seed; an out-of-bounds `env-seed` row the env re-seeds or
  removes is healed and audited rather than refusing.
- **6c-E9 (D9, R-6c2, R-6c3).** Spec `:734-736` and criterion 11's "the compiler forbids a consumer
  bypass" is read as the compiler forbidding a `Setting` from reaching any function that reads the
  env; a hand-built `Threshold` carrying the key is caught by exit gate 11's grep instead. Wall
  keys carry no apply mode or sensitive marker until rung 14's hook 18.
- **6c-E10 (D10, R-6c9).** Criterion 13's "each of 6b-i's five constants reads through the
  accessor" is read as: the local keep's pre-seed read resolves from `Inspect`'s raw rows through
  the facts entry point, since no handle exists before the migration it caps; its `OpenServe`
  prune proof moves to 6d, the first caller.

---

## The 6c to 6d handoff

1. **Exporting the apply and wiring it into `serve`** after the open and recovery and before the
   listener; resolving `DUBPLATE_STORE_LOCAL_SNAPSHOT_KEEP` through `Inspect` and the from-facts
   entry point before `OpenServe`, passing it as `ServeOptions.Keep`, with the `OpenServe` prune
   proof over a store at `0004` with a seeded keep row and `0005` pending (D10).
2. **Moving the nonce-retention rule into the registry's correlated-rule list** and re-running
   criterion 5 over the real nonce-retention and link-expiry pair (D6).
3. **Re-homing the auth keys as settings** with `sensitive` on both lifetimes and the cap.
4. **A "restart"-mode key is read once at start, never through a reloading accessor** (R-6c8); no
   6c key uses that mode, so it is untested at 6c.
5. **On Ruling 2's yes, `deploy.sh` halts on `check`'s nonzero exit** before restarting `serve`.
6. **After a `restore`, the seed may refuse on a restored `admin` row**, and `env-seed` rows
   re-seed.

## To 5d and 6d

Between 6c's merge and 6d's, a deployed binary's `doctor` says an env value "seeds at serve's next
start" while `serve` seeds nothing (the apply is unexported until 6d exports it). The template
carries none of the five keys, so the window matters only for a hand-set value.

## To rung 14

The admin write path: source `admin`, `seed_value` recording the seed it replaced, its channel,
its audit kind; and a resolving channel that works while `serve` is down for an `admin` row that
refuses `serve`'s start, such as a CLI verb.

## To rung 7

The reservations on `member_quota` and the default-collection preference. To the rungs that read
them: the remaining preference columns (`member_preferences`' theme, language, ReplayGain mode,
sort order, notification choices, default collection).
