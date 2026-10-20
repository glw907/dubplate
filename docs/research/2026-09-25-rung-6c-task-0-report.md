# Rung 6c Task 0: the staleness pre-task report

Read on 2026-09-25 in the lane 2 worktree `.claude/worktrees/lane-2-6c` (branch `lane-2-6c`, HEAD
`392d1cd`, the same commit as master). Every item below was produced by a command in that tree; the
matched line is quoted beside its site. Plan:
`docs/plans/2026-09-25-rung-6c-routed-schema-settings-home.md`, revision 2.2.

**Verdict: 0 blockers.** Seven blocker questions answered yes. Two items are conductor bookkeeping
(the lane registry row and the leftover 6b-ii-b worktree), and eleven citations moved without
changing content. Four notes route to a build task.

## Blockers

1. **6b-ii-b merged; the worktree is fresh from master.** `git worktree list` shows the main tree,
   `lane-2-6b-ii-b`, and `lane-2-6c` all at `392d1cd`. Master's head is `392d1cd` ("Match the backup
   package doc's restore paragraph to the code"), above the 6b-ii-b Task 8 draft `05b5f55` and the
   post-review commits `bd34f9a` and `9870678`. `git status --short` is empty in `lane-2-6c` and in
   `lane-2-6b-ii-b`. In the main tree it shows only agent-memory files and two untracked
   `docs/research/2026-09-04-*.md` files, none under `internal/` or `cmd/`.
   - **Conductor bookkeeping, not a blocker.** The main tree's `docs/STATUS.md` lane registry
     (`## Lane registry`) still lists lane 2 as "worktree `lane-2-6b-ii-b` ... OPEN: 6b-ii-b ...
     then 6c (approved, `50a1cf8`) in a fresh worktree". No row names `lane-2-6c`. The plan's
     global constraints say "The conductor writes the lane registry row before Task 0." The
     `lane-2-6b-ii-b` worktree also still exists, clean, at `392d1cd`.
   - The worktree had no `web/node_modules`. `npm ci --prefix web --offline` installed it from the
     local cache ("found 0 vulnerabilities"), with no network fetch.
2. **The 6b-ii-b surfaces still carry what 6c re-homes.**
   - `internal/backup` constants. `guard.go:17` reads `const guardFraction = 0.5`. `prune.go:19`
     reads `const snapshotKeep = 14`, `prune.go:23` reads `const othersKeep = snapshotKeep - 1`,
     and `prune.go:27` reads `const archiveRetention = 90 * 24 * time.Hour`. Their only readers:
     `guard.go:86` (`if float64(files) < guardFraction*float64(tracks) {`), `prune.go:117`
     (`if len(others) <= othersKeep {`), `prune.go:122` (`for _, stamp := range others[othersKeep:] {`),
     and `prune.go:153` (`if own.Sub(stamp) <= archiveRetention {`). `grep -rn` over `internal/`
     and `cmd/` finds no other reader, test files included. **Restore and its fetch read none of the
     four.** `restore.go:510` has its own `func (s source) checkLibrary`, which does not call
     `guard`.
   - `backup.Job` (`backup.go:21`) has seven fields: `Store`, `Lease`, `Instance`, `RunID`, `Remote`,
     `LibraryRoot`, and `Logger`. `Job.Run` (`backup.go:63`) runs `progress("checking the library")`,
     then `exec.LookPath("rclone")` (`backup.go:69`), `ParseRemote`, `resolveRoot`, and
     `checkLibrary`. Its first rclone launch is `t.run(ctx, nil, "sync", ...)` after
     `progress("syncing the library")`. The prune is built at `backup.go:133`:
     `p := prune{t: t, remote: remote, stamp: stamp, localDir: localDir, start: start}`. So R-6c13's
     validation goes before `LookPath` at `:69`.
   - The guard text is unchanged: `guard.go:87` reads `"library root %s holds %d files, fewer than
     half the catalog's %d tracks"`, and `job_test.go:547` reads
     `wantMsg:  "holds 1 files, fewer than half the catalog's 4 tracks",`. `guard_test.go:131-135`
     also carries half-boundary rows ("below half", "exactly half", "odd count, just under half",
     "odd count, just over half"). Task 4's message change touches only `job_test.go:547`, but its
     fraction plumbing reaches `guard_test.go`'s table.
   - `ops/local` is unchanged in shape. `jobs.go:23` reads `const retention = 90 * 24 * time.Hour`,
     and its reader is `jobs.go:93` (`cutoff := at - retention.Microseconds()`) inside
     `insert` (`jobs.go:83`). `type binder func(kl *store.KindLease, h ops.Handle) body` is at
     `jobs.go:37`, `func (r *Runner) backupJob` at `jobs.go:41`, and `func (r *Runner) start` at
     `local.go:171`, which calls `r.insert(ctx, actor, kind, nil)` and then `j.body = bind(kl, j.handle)`.
     `StartBackup` is at `local.go:155`.
   - **`local.New` callers: five, not three.** `grep -rn 'local\.New('` finds
     `cmd/dubplate-server/backup.go:151`, `internal/ops/opshttp/contract_test.go:317` and `:825`,
     **`internal/backup/job_test.go:149`, and `internal/backup/standin_test.go:172`**. The last two
     are test files in `internal/backup`, which is inside Task 4's territory. They need the same
     reader fields Task 4 adds to `local.Config` (see note 12).
   - `store.ServeOptions` (`open.go:85`) keeps `Keep int` at `open.go:97` beside the new
     `Reconcile func(ctx context.Context, h *Handle) (report []byte, err error)` at `open.go:115`.
     Its doc reads "It must be at least 1" (`open.go:96`). The only `OpenServe` is at `open.go:219`,
     and no production caller builds a `ServeOptions{`.
   - `cmd/dubplate-server/restore.go` passes no `Keep` and none of the five values. A grep for
     `Keep`, `ServeOptions`, `RETENTION`, `SNAPSHOT_KEEP`, and `GUARD_FRACTION` in it matches nothing.
3. **`internal/config` is quiet.** `git log 8d0eed8..HEAD -- internal/config` (`8d0eed8` is the
   plan's first commit) lists one commit, `75fde16 fix(serve): fold 5c-iii's security review
   findings`. `git log 75fde16..HEAD -- internal/config` is empty. `git merge-base --is-ancestor
   75fde16 HEAD` succeeds. The main tree has no uncommitted edit under `internal/` or `cmd/`
   (`git status --short -- internal cmd` is empty). Lane 1's registry row reads "OPEN: 5c-iv ...
   segment D-i (Task 0, Task 7) running as `wf_9508455c-ca8`". The 5c-iii plan, which carries
   5c-iv, writes nothing under `internal/config`.
4. **5d-i has not launched ahead of 6c.** `git log --all 75fde16.. -- internal/config
   cmd/dubplate-server/doctor.go cmd/dubplate-server/check.go` is empty on every branch. Lane 1's
   row names 5c-iv, not 5d-i.
5. **The highest migration is `0003_catalog.sql`.** `ls internal/engine/store/migrations/` lists
   `0001_ledger.sql`, `0002_jobs.sql`, and `0003_catalog.sql`. `testdata/fixtures/` holds the same
   three names. Task 1 numbers `0004` and `0005`.
6. **No settings surface exists yet.** `internal/settings` does not exist ("No such file or
   directory"). `grep -rn -i settings internal/engine/store/migrations/` matches nothing. No Go file
   under `internal/` or `cmd/` matches `AddSetting` or `type Setting`, and no non-test file under
   `internal/engine/store` or `cmd/` mentions `settings`.
7. **`Handle.Query` merged.** At `internal/engine/store/conn.go:115`:
   `func (h *Handle) Query(ctx context.Context, query string, args ...any) (*sql.Rows, error)`, with
   the doc "Query runs query on h's read pool, as [Handle.QueryRow] does, and returns every row. The
   caller closes the rows, which hold one of the pool's connections until then." Its first
   production caller is `internal/backup/reconcile.go:98`. For R-6c9's apply, `Tx.Query` exists at
   `internal/engine/store/write.go:161` and returns the store's own `*Rows`, not `*sql.Rows`.

## Notes

8. **6b-ii-b's Task 8 handoff names nothing for 6c.** The draft
   (`docs/research/handoffs/draft-HISTORY-6b-ii-b.md`) is headed "The 6b-ii-b to 6c and
   6d handoff" (`:262`), but its lists are "To 6d", "To rung 14", and "To rung 12" only. The draft
   is now stale in one place: it says "`ServeOptions.Reconcile` and `backup.Reconcile` are `serve`'s
   own wiring". `bd34f9a` unexported that function, which is now `reconcile` at
   `internal/backup/reconcile.go:55`, with `type report struct` at `reconcile.go:31`. The store's
   pre-migration name constants are also private now. `store.IsPreMigrateCopy`
   (`snapshot.go:251`) is the export, and `prune.go:178` calls it. Neither change touches a 6c
   surface. One related item from the draft: `isLocalCopyName` no longer matches pre-restore
   copies. That changes `pruneLocal`, which is next to Task 4's prune edits but is not one of them.
9. **Symbols the build tasks use, at HEAD.**
   - `newProductionRegistry` (`main.go:108`) calls, in order: `registerShadowIntegration`,
     `registerShadowThresholds`, `registerServeConfig`, `registerAuthConfig`, `registerStoreConfig`,
     `registerWebConfig`, and `registerBackupConfig`.
   - `registerStoreConfig` (`store.go:51`) is `reg.AddThreshold(storeBusyTimeoutThreshold)`.
     `registerBackupConfig` (`backup.go:52`) is `reg.AddIntegrationClass(backupIntegrationClass)`,
     and the class is declared at `backup.go:31` with `Name: "backup"`.
   - `newDoctorCmd(environ []string, reg *config.Registry, stdout, stderr io.Writer)` is at
     `doctor.go:28`. The row order is unchanged: the store row (`doctor.go:109`, `storeRow(...)`),
     then `backupFactsRows` (`:119`), then the backup remote row. The fail line is
     `if loadErr != nil || rosterFailed || storeFailed || backupRemoteFailed {`.
   - The line-count tests use two `wantTotal` expressions. `doctor_test.go:230` is
     `ownedCount + reservedCount + integrationCount + thresholdCount + sourceCount + storeRowLines +
     backupDoctorRowsOff`, and `doctor_test.go:340` is `len(reg.OwnedWall()) +
     len(reg.ReservedClasses()) + len(reg.IntegrationClasses()) + len(reg.Thresholds()) +
     len(reg.IdentitySources()) + storeRowLines + backupDoctorRowsOn`.
   - `check.go`'s store row: `newCheckCmd` (`check.go:19`) runs `store.Inspect` and then
     `status, failed := storeRowStatus(r)` (`check.go:48`). It calls `storeRowStatus` directly, not
     `storeRow`.
   - `invariant6Exempt` (`invariant_test.go:28`) holds `ledgerTable` alone.
   - `applyInTx` is at `migrate.go:298`, `rollbackMigration` at `:337`, and `commitMigration` at
     `:349`. Its "no transaction is active" branch is at `:355`.
   - In `inspect.go`: `jobsVersion` is at `:123`, `type Facts struct` at `:128`, `func Inspect` at
     `:168`, `func readFacts` at `:228`, and `func lastSuccesses` at `:267`, which still opens and
     closes its own `sql.Open(sqliteDriverName, inspectDSN(...))`. The eight `store.State` values are
     `StateAbsent`, `StateRestoreInProgress`, `StateBusy`, `StateNotAStore`, `StateRefused`,
     `StateOlder`, `StateNewer`, and `StateCurrent` (`sed -n 20,52p inspect.go | grep -c '^	State'`
     prints 8). `StateRestoreInProgress` is 6b-ii-b's addition, so Task 5's older-store and facts
     states meet it.
   - `Handle.QueryRow` is at `conn.go:108`.
   - `rejectNamespace` is the method `func (w *wall) rejectNamespace(reg *Registry)` at
     `config.go:283`. Its threshold branch is `if thresholdOf(reg, name) {` with the comment "A
     registered threshold is a declared name whose own Parse* call resolves and validates it" at
     `config.go:321-322`. `thresholdOf` is at `thresholds.go:112`, `AddThreshold` at
     `registry.go:269`, and `claim` at `registry.go:121`.
10. **6b-ii-b's spend is not recorded yet.** The draft's "Both budgets" (`:213`) reads "Tokens: `<the
    pass's actual, filled at the fold>` against the **7.53M execution ceiling**". `docs/HISTORY.md`
    has no 6b-ii-b entry: its top entries are 5c-iii (`:5`), 6b-ii-a (`:320`), the matcher pass
    (`:524`), and 6b-i (`:979`). The conductor re-derives the ceiling from the folded actual.
11. **The rulings as recorded** (plan "Rulings", `:81-93`). Ruling 1 is **yes** (Geoff): 5.4M
    execution, flag 4.32M, lane 2 once 6b-ii-b merges. Ruling 2 is **yes** (conductor design
    decision): `doctor` and `check` exit nonzero on a refusing seed outcome, and `backup` still logs
    and runs. Ruling 3 is **yes** (conductor design decision): keep 3 to 1000, archive retention
    168h to 87600h, and guard fraction 0.25 to 1, inclusive, with NaN and both infinities refused.
    These match the dispatch notes.
12. **Routed to Task 4:** the two extra `local.New` callers in item 2
    (`internal/backup/job_test.go:149`, `internal/backup/standin_test.go:172`) must pass the new
    reader fields. They are in Task 4's package territory, so this does not widen the wall. The
    task's file list should name them.
13. **Routed to Task 2: `internal/config`'s package doc is stale** (`config.go:6-46`). Class 2 says
    "The classes are backup targets (rung 6), intake (rung 7), pipeline (rung 9), monitoring (rung
    9), decision and review queues (rung 13), and the streamer (rung 12)." But
    `keys.go`'s reserved table has no backup entry: its `Name:` rows are `intake` (`:61`),
    `pipeline` (`:72`), `monitoring` (`:81`), `decision and review queues` (`:91`), and `streamer`
    (`:98`). Backup is an integration class (`cmd/dubplate-server/backup.go:31`). The doc's "six
    classes" also leaves out registered thresholds, which `rejectNamespace` accepts as declared.
    Task 2 adds the store-homed family to that classification, so it corrects this paragraph where
    it touches the doc. Nothing should be built on the doc's current text.

## Moved citations (content intact at a new line)

| Plan cites | Now at | Matched line |
| --- | --- | --- |
| `config.go:243` (`rejectNamespace`) | `config.go:283` | `func (w *wall) rejectNamespace(reg *Registry) {` |
| `inspect.go:121` (`jobsVersion`) | `inspect.go:123` | `const jobsVersion = 2` |
| `inspect.go:126` (`Facts`) | `inspect.go:128` | `type Facts struct {` |
| `inspect.go:157-165` (Inspect doc) | `inspect.go:159-167` | `// Inspect reports the state of the store under opts.StateDir without` |
| `inspect.go:223` (`readFacts`) | `inspect.go:228` | `func readFacts(ctx context.Context, l leaf, busyTimeout time.Duration) (*Facts, error) {` |
| `inspect.go:262-271` (`lastSuccesses`' open) | `inspect.go:267-277` | `db, err := sql.Open(sqliteDriverName, inspectDSN(dbPath, busyTimeout))` |
| `open.go:175` (OpenWrite's older refusal) | `open.go:196` | `...binary at version %d: start serve to migrate"` |
| `open.go:92-94` (`Keep` "must be at least 1") | `open.go:95-97` | `// Backup is off. It must be at least 1.` |
| `conn.go:187` (the one DSN builder) | `conn.go:194` | `func buildDSN(path string, params ...string) string {` |
| `prune.go:154` at `30a66ac` (archive compare) | `prune.go:153` | `if own.Sub(stamp) <= archiveRetention {` |
| `prune.go:123` at `30a66ac` (`others[othersKeep:]`) | `prune.go:122` | `for _, stamp := range others[othersKeep:] {` |

Unmoved and confirmed: `conn.go:108` and `:115`, `migrate.go:298`, `:337`, and `:355`,
`guard.go:86` and `:87`, `job_test.go:547`, `thresholds.go:84` (`func ParseInt`), `config.go:183`
(`// get returns name's value, treating an empty value as unset.`), and `doctor.go:399`
(`// and treats an empty value as unset, [config]'s own wall.get rule.`).
