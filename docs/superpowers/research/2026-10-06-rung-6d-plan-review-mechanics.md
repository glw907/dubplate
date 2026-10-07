# Rung 6d plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-26-rung-6d-identity-serve-store.md`, revision 2.3 at `4b2cddf`.
Lens: mechanics and feasibility, plus proportionality and the four planning-miss items. Reviewer:
Opus 5.5, high effort, 2026-10-06.

Method. `git diff --stat 0cb33ae..4b2cddf` touches only three docs, so every code citation was read at
`4b2cddf` in the tree. Probes ran in a scratch copy of `cmd/` and `internal/` (`git archive 4b2cddf`)
under the session scratchpad, never in the repo. Findings are ranked by consequence. Counts: 0
blockers, 2 majors, 9 minors.

## Verdict

The plan builds. R-6d21 and R-6d22 hold mechanically. The ceiling arithmetic and the deliverable count
check out. The box section matches STATUS and `deploy.sh`. Two majors need a fold before approval.
Task 1's older-store rows cannot be built from `cmd/dubplate-server` as written (probed), and most of
them re-prove `OpenServe` behavior the store package already proves. The runbook keeps a paragraph
that Task 2 makes false and that tells operators to clear their settings.

## Findings

### M1 (major): Task 1 cannot build its older-store fixtures, and most of its rows duplicate store tests

**Location:** plan `:946-991` (Task 1's state table, criteria, and mutations).

**Defect.** Six rows need a genuine store at `0004` that `serve`'s preparation function then migrates:
the two prune rows, the two out-of-bounds keep rows, the older-store lock row, and the reconcile row.
The tests live in `cmd/dubplate-server`, which has no seam to build one. `buildStore` and
`migrationSourceThrough` are unexported test helpers in `internal/engine/store`
(`open_test.go:221`, `invariant_test.go:518`). The only `cmd` precedent lowers `user_version` and cuts
the ledger (`restore_test.go:747-775`, `storeFile`). It leaves 0005's tables in place, so a migration
re-runs `0005_routed.sql` and fails.

**Evidence.** Probe: create a store with `OpenServe`, then apply `storeFile`'s statement
(`DELETE FROM migrations WHERE version > 4; PRAGMA user_version = 4`), then call `OpenServe` again:

```
inspect state=older reason="" facts=true
second OpenServe err=migrate .../dubplate.db: apply 0005_routed.sql: SQL logic error: table member_overrides already exists (1)
```

**Duplication.** Several rows re-prove what the store package already proves at the right altitude:

- the shared-held, exclusive-held, and current-shared lock rows
  (`TestOpenServe_LockHeldPastMigrationWait` `:860`, `TestOpenServe_CurrentKeepsSharedLock` `:1081`);
- the interrupted create (`TestOpenModes_InterruptedCreate` `:518`);
- E11 (`TestOpenModes_LostDatabaseBesideCopy` `:747`);
- the foreign owner (`TestOpenModes_OwnerAndFilesystemRefused` `:622`);
- the other intents (`TestOpenModes_RestoreIntent` `:692`);
- prune by `Keep` and `Backup` (`TestOpenServe_BackupFieldControlsPruneOnMigrate` `:209`).

The store's own lock test holds the lock from a second descriptor in one process. Task 1 adds a
re-executed lock-holder child, which needs a new `TestMain` mode in `backup_test.go:75`, only so "the
named pid differs" (see m5).

**Proposed fold.** Task 1 proves what `cmd` decides:

1. Extract R-6d4's keep decision as a pure function over the `store.Report`, the environment, and
   the registry. Table-test every R-6d4 bullet on synthetic `Facts`: in-bounds row, out-of-bounds
   `admin` row refusing, healed `env-seed` row, no row, no facts.
2. Write one wiring test per path. A refusal from `OpenServe` yields an error and no shell. A
   lowered `newer` store is enough, since `storeFile`'s cut works for that direction. `MigrationWait`,
   `Keep`, and `Backup` reach `OpenServe`.
3. Prove D10 end to end once, setting through to the prune count. Use a cmd helper that lowers the
   store and drops 0005's six tables (`member_overrides`, `member_preferences`, `member_quota`,
   `group_quota`, `device_credentials`, `move_intents`). Probe: Inspect reports `older` with facts,
   and `OpenServe` with `Keep` 1 migrates and leaves `pre-migrate-v4.db` alone. The helper must
   name its dependence on the 0005 table list, because every later migration (rung 12's slices)
   extends it. The same helper serves the reconcile row.
4. Cite the store tests above for the rest, and drop those rows from Task 1's table.

Cost saved: roughly eight subtests and a new re-exec child mode, about 0.2M of Task 1's 0.88M, and the
fix round the infeasible fixtures would otherwise cost.

### M2 (major): the runbook's settings paragraph becomes false at Task 2, and Task 8 does not touch it

**Location:** `docs/server-runbook.md:340-345`. Plan R-6d18 (`:735-760`), Task 8 (`:1439`, `:1474-1478`),
and Goal (`:158`).

**Defect.** The runbook says: "**That statement is not yet true.** `serve` does not apply the seed
today ... treat such a row as a hand-set value that has no effect, and clear it from the
environment." Task 2 wires the seed, so this paragraph becomes false. An operator who follows it after
6d removes settings that now take effect. R-6d21 also changes what the absent-store rows print: an
out-of-bounds env value now fails `doctor` and `check`. The runbook's text at `:334-339` still
describes only "store holds no settings yet". The Goal says "Every operator string that says 'serve
applies it at its next start' is true." R-6d18's list and Task 8's render-and-read checks cover the
install order, the upgrade, the rollback note, `:265`, and `:773`, but not this section.

**Proposed fold.** Add a fifth bullet to R-6d18. "The settings rows" section drops the "not yet
true" paragraph and states R-6d21's behavior: before the first start, an env value outside its bounds
fails `doctor` and `check`. Task 8's render-and-read adds this section to what it checks.

### m1 (minor): `reconcile.go:167` already tells an operator to run `member import`

**Location:** `internal/backup/reconcile.go:167` (`"the users table is empty: run member import\n"`);
plan 6d-R11 (`:845`).

**Defect.** 6d-R11 says the spec's "the report says to run `member import`" is "Not built". It is
built. Under Ruling 4's yes, no `member import` verb exists, so a restore report over an empty
`users` table names a command that does not exist.

**Fold.** Task 1 already edits the reconciler (`:932`). The line names `member add --admin` under
Ruling 4's yes. 6d-R11's wording becomes "built; Task 1 rewords it".

### m2 (minor): `keys.go:126` names the key Task 5 retires, and Task 5's grep finds it

**Location:** `internal/config/keys.go:126-127`; plan Task 5 criterion (`:1274-1275`).

**Defect.** The pre-pivot `DUBPLATE_CONTRIBUTORS_FILE` supersession reads "renamed to
`DUBPLATE_ROSTER_FILE`", which points at a key 6d supersedes. `grep -rn DUBPLATE_ROSTER_FILE
--include=*.go cmd internal`, outside tests, returns this line as well as the new registration. Task
5's acceptance then fails as written.

**Fold.** Task 5 rewords `keys.go:126` to name `member add --admin`. That is a message edit, not a key
declaration, so CLAUDE.md's `keys.go` rule does not bar it. The grep criterion then names both lines.

### m3 (minor): Task 4's file list misses `fixtureAdmin` in `internal/auth/admin_test.go`

**Location:** plan Task 4 (`:1146-1148`, "the two test fixtures that then need an id").

**Defect.** `internal/auth/admin_test.go:123-132` (`fixtureAdmin`) is a cookie-channel identity with no
id. The tests hand it through `stubSeam` to `RequireAdmin`, which checks `Valid()` (`admin.go:92`,
`:121`). Once `Valid` refuses a zero `UserID`, those tests answer 404. There are three such fixtures,
not two.

**Fold.** Name the third fixture in Task 4's Files list.

### m4 (minor): R-6d21's state set omits a store below the jobs table, and NOT NULL is not non-empty

**Location:** R-6d21 (`:789-803`), Task 6's states (`:1337-1341`), R-6d10 (`:629`).

**R-6d21 holds.** `settings.evaluate(reg, environ, nil)` (`evaluate.go:212-269`) runs the per-key rule
and every correlated rule over an empty row set. That is exactly what `serve`'s first `apply` computes
against a freshly migrated, empty `settings` table (`apply.go:29-50`). An absent store and an empty
pre-created leaf both report `StateAbsent`. Probe: `inspect empty leaf: state=absent facts=false`;
`OpenWrite` answers "start serve first" and `OpenServe` creates. So the prediction is sound.

**Gap.** `Inspect` returns `StateOlder` with `Facts` nil below `jobsVersion` (`inspect.go:242`).
`check` branches on `r.Facts != nil` (`check.go:62`), so an implementation keyed on `Facts` reaches
absent and `0002`-`0003` but not `0001`. Task 6's table names absent and `0003` only. No box holds a
`0001` store, so the consequence is small.

**Fold.** R-6d21 keys on state (absent, or older without a settings table, `Facts` present or not).
Task 6's table adds the `0001` row.

**Separately (R-6d10, `:629`).** "Every row carries a username (`NOT NULL`, quoted)" overstates the
schema. `username TEXT NOT NULL UNIQUE COLLATE NOCASE` (`0003_catalog.sql:5`) admits `''`. The
guarantee is R-6d22's single write path, not the schema. The consequence is fail-closed, since
`Valid` refuses an empty `Principal`.

**Fold.** Reword R-6d10 to cite R-6d22.

### m5 (minor): the lock rows assert a pid that `/proc/locks` does not always yield

**Location:** Task 1 (`:960-975`).

**Defect.** `busyError` names "held by pid N" only when `/proc/locks` resolves the lock. Otherwise it
says "held by another process" (`lock.go:99-108`), and its doc says this always happens on btrfs. The
workstation's `/tmp` is tmpfs today, but `$HOME` is btrfs, and a `TMPDIR` change would flip the rows.
The repo's precedent accepts either form (`lease_test.go:66`).

**Fold.** Use the precedent's two-form assertion. This disappears if M1's fold drops the lock rows.

### m6 (minor): privileged fields omit restore as a channel for `users` rows

**Location:** the privileged-fields table (`:443-460`), R-6d22 (`:805-814`).

**Defect.** `restore` replaces the whole database, so it sets every `users` field, `admin` included.
The table names `AddMember` alone for "any `users` row". Under Ruling 3's yes, the refusal of a
foreign snapshot waits for rung 16. Until then, a snapshot from another store restores its members
and admins. The trust boundary is the operator, so the risk is small. The security reviewer is briefed
with this table, though, and will not look at the restore path.

**Fold.** Add a row: "`restore` (whole-database replace; operator-run; foreign-snapshot refusal is
rung 16's under Ruling 3)". R-6d22's "by construction" claim names it.

### m7 (minor): Task 8's misspelled-path row fails three checks, not two

**Location:** Task 8's state table (`:1463`).

**Defect.** With the unit granting `/var/lib/dubplate/stor` and the array `(auth store)`, three checks
fail: the existing check (`stor` not in the array, `check.sh:716`), the new store check, and the
new reverse check (`store` in the array, granted by no unit). A report asserting "twice" is wrong.

**Fold.** Change "twice" to "three times" and name the reverse check.

### m8 (minor): Task 6 and Task 7 disagree on when `auth_test.go`'s `resolveAuthConfig` table goes

**Location:** Task 6 (`:1301-1307`), Task 7 (`:1376-1378`).

**Defect.** Task 6 turns the ten thresholds into `config.Setting` values. `resolveAuthConfig` calls
`config.ParseDuration(environ, <Threshold>)` (`auth.go:518-563`), which no longer compiles against a
`Setting`. Task 6 also says "every test caller stops calling `resolveAuthConfig`". Yet the table
caller at `auth_test.go:881` retires only in Task 7. Task 6 cannot be green with that table in place.

**Fold.** Move the table's retirement, and `resolveAuthConfig`'s deletion, into Task 6. Task 7 keeps
`authConfig`'s deletion.

### m9 (minor): the rendered fixture's two seeded members can tie on `created_at`

**Location:** Task 4's harness bullet (`:1149-1158`), R-6d10 (`:642-643`).

**Defect.** `List` orders by `created_at`, then `id`. A harness that seeds both fixture members in one
pass with one clock reading orders them by random id, so the rendered list order varies from run to
run.

**Fold.** Task 4 seeds the two members with distinct `created_at` values, admin first, through
`AddMember` or with explicit stamps.

### Over-ceremony

- **O1 (minor, clock time).** Task 8 proves each failing leg state "by applying its one edit ... a full
  `check.sh` run if the leg cannot run alone" (`:1466-1469`). It cannot run alone: the leg is inline
  inside a block of `check.sh` (`check.sh:657-737`). Five states mean five heavy-lane gates
  (about 106s each, Chromium), plus their transcripts, for shell-text assertions. **Fold.** Task 8
  moves the leg into a function, or into `scripts/check-state-dir.sh` that `check.sh` calls, so each
  mutation runs in seconds. The gate still runs it unconditionally.
- **O2 (minor, ambiguity).** Task 8's "no store test or source string names `runuser`" (`:1472-1473`)
  would also catch `cmd/dubplate-server/unidecode.go:269` and its tests. That verb does not load the
  wall, and Task 8 does not list it. **Fold.** Scope the criterion to `internal/engine/store`.

## Checked and sound

- **Citations.** About 40 cited `file:line` references were spot-checked at `4b2cddf` and say what
  the plan claims. They include `member.go:26/:95`, `admin.go:38/:255/:308`, `roster.go:118/:141/:340`,
  `serve.go:247/:375/:384`, `doctorprobe.go:104`, `backup.go:274`, `write.go:68`, `migrate.go:121`,
  `open.go:122/:440/:656`, `config.go:289`, `auth.go:304`, `doctor.go:198`, `leaf.go:113`, and the
  four `runuser` test assertions. One is off by file: `:440` sits in `open.go` (`statLeaf`), and the
  bullet's placement already says so.
- **Grep premises.** `\.Bump(` and `INSERT INTO users` have no production hit. The demote grep finds
  only `roster.go:754` and `admin.go:299`. Six production `Lookup`/`List` callers sit outside Task 3's
  files, matching the plan.
- **Task 5's mutation.** `AddIdentitySource` claims Required keys under the identity-source family
  (`registry.go:395`). `AddSuperseded` claims under the superseded family (`:507`). Registering both
  panics in `claim` (`:125-130`), so the mutation is real. The empty-value row matches
  `rejectNamespace` (`config.go:288-290`).
- **R-6d5.** `apply` sets no deadline (`apply.go:29-31`), so the 30s wrap is needed and the mutation
  is real.
- **R-6d3.** The unit is `Type=exec`, so a five-minute in-process lock wait never trips a start
  timeout. `Restart=on-failure`, `RestartSec=5`, and `StartLimitBurst=30` per 300s are as quoted.
- **Box.** `deploy.sh` creates `/var/lib/dubplate` as `music:music 0755` (`:266`). So
  `leaf.create`'s state-directory owner check (`leaf.go:72-77`) passes for `serve` under `User=music`.
  An empty pre-created leaf reads `absent` and `OpenServe` creates over it (probed). `deploy.sh`'s
  only `systemd-run` is the unidecode install (`:361`), and it starts no unit. The upgrade's four
  steps match STATUS's box (binary `d1e93e9`, unit disabled, `auth`/`roster`/`unidecode` leaves).
- **Ceiling.** The arithmetic holds: 0.25 + 5.28 + 0.88 + 0.70 + 1.02 = 8.13; 8.13 / 0.8 = 10.16,
  rounded up to 10.5M. The deliverables total 28 across Tasks 1 to 8.
- **Task order.** Each task's inputs exist when it runs (Task 3's interim dual backend, Task 4's
  `Valid` change, Task 5 after the callers move, Task 6's interim read), apart from m8.
- **Proportionality.** `auth-data` fits Tasks 0 to 8. Two security reads at the B and C seams are
  the spec's allowance. Beyond M1 and O1, no task's test mandate exceeds its risk.

No owner fork arises under this lens. R-6d21 has one correct answer: predict `serve`'s refusal.
