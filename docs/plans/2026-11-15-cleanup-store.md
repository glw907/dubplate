# Pass plan: the restore seal, and the store and config seams cut down

Initiative: dubplate-server
Round: cleanup (pass 3 of 5, under the conductor `conductor-dubplate`)
Pass: cleanup-store

Design: `/var/home/glw907/Projects/initiatives/dubplate-server/specs/2026-11-15-cleanup-round-design.md`, approved by
Geoff in the Approval Queue on 2026-11-16 (document `20261116-000308-88f4d3`). Read in full: "Calls made here" (11
and 12), "Scope", "How it fails", pass 3's row under "Passes" and the paragraph after it, "What each merge changes",
"Acceptance" (pass 3, and the two every-pass items), "Standing answers" 13 to 20, and "Backlog verdicts" (#116 to
#119, #26). The findings are `docs/research/2026-11-10-whole-repo-review.md` lines 142-183 (S1 to S14), and its
"Leave alone" list is lines 284-288. Where this plan and the design differ, the design wins and the difference is a
stop. Pass 2's rulings are in `docs/archive/plans/2026-11-15-cleanup-gate.md` `## Rulings`; none of them binds a file
this pass touches.

Worktree: `/var/home/glw907/Projects/dubplate/.claude/worktrees/cleanup-store`, branch `cleanup-store`, cut from
`origin/cleanup-gate` at `209821fd12752f4a5a44e783a0fe490df3d566dd`. Pass 2 (PR #13) has not landed, so this branch
is stacked on it: **the PR's base is `cleanup-gate`, not `master`**, and the round's close packet merges #13 first.
Class: `auth-data` (design call 11). The PR never lands unattended; it waits for the round's close packet with passes
4 and 5. Review lens: one `standards:plan-reviewer` read of this plan (record under "Review record"); a task-scope
`standards:diff-reviewer` read on each `auth-data` task (T2, T7); and at the close, the whole-branch
`standards:diff-reviewer` and one `standards:web-auth-security-reviewer` read. Executor: `dubplate-cleanup-store`,
`sonnet` at `medium` with the Fable advisor; T7 goes to one `opus` subagent. Design path minutes: 130. This plan's
path: 240 (see "Path"), under the 260-minute park line (standing answer 14).

## Pre-flight

No step needs Geoff. The pass adds no secret, no sign-in, no dialog, and no deploy, and it reaches no live account.

Permissions checked: read `~/.local/state/claude-denials.jsonl` on 2026-11-16 for the last 30 days. Twelve entries
carry a dubplate cwd: eleven "Classifier unavailable" refusals in pass 2's planning session (`grep`, `sed`, `cat`,
`git show`, `ls` over its job dir), which are classifier outages and no rule, and one "[Unauthorized Access]" on a
`curl` download of a libxml2 tarball in pass 1. No task here downloads anything. Every command in the tasks (`go
test`, `go test -race`, `grep`, `git`, `gh pr checks`, `bash scripts/gate.sh`) has run in passes 1 and 2 without a
denial. No `.claude/settings*` rule is wanted.

Reviewed: one `standards:plan-reviewer` read; its findings and the fold are under "Review record".

Evidence: the mechanism table and the recorded searches below. Each row was read or run against
`209821f` by this planner on 2026-11-16.

## What the pass changes

A `--from-file` restore seals its stage the way a fetched restore does: the file and `staging/` are both synced
before `Record` writes the intent that names the stage. One session method does it for both paths. Then the review's
cuts land: one package prunes pre-migrate copies, the restore intent has one model, the settings row has one type,
lock acquisition has one wait, `atomicfile` has one write path, per-kind parsing and environment lookup each have
one home in `config`, a typed `config.KeyError` carries the key, `Threshold` gains bounds, two uncalled exports and
two unused lease states go, every `StateRefused` refusal keeps its cause, the named redundant tests go, and the
conductor trailers leave `config`. BACKLOG #26's residue (the accessor's mutex) folds in.

## Rules that bind every task

1. **Files.** Pass 3's row only: `internal/engine/store/**` except `internal/engine/store/migrations/`,
   `internal/backup/**`, `internal/settings/**`, `internal/config/**`, `internal/atomicfile/**`,
   `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/serve_store_test.go`, `cmd/dubplate-server/doctor.go`,
   `BACKLOG.md` (closing this pass's items and filing finds), and this plan and its prompt (moved to
   `docs/archive/plans/` at the close). A call site a changed signature forces is taken and listed under
   `## Decisions` (standing answer 15); the forced sites this plan already knows are named in each task. Anything
   else is filed through `standards:log-issue` in BACKLOG and the PR body. Never touch `docs/STATUS.md` (the
   conductor's; wanted text goes in the PR body under `## STATUS text`) or `.claude/settings*` ("`.claude/settings*`
   is never an executor's file: record the rule the pass wants for the close packet and go on"). Never touch a file
   under `internal/engine/store/migrations`, `internal/config/keys.go`'s key rows, or anything in the review's
   "Leave alone" list: `store/write.go:69-227` and `store/swap.go:126-235` are the two in this row (standing answer
   16). A cut that would touch them is left and the PR says so.
2. **No key change.** No config key is added, renamed, or removed. "Config keys are declared through the registry in
   each owner's file, never by editing `internal/config/keys.go`." Deleting the uncalled `OwnedWall` and
   `ReservedClasses` functions and comment trailers in `keys.go` (T9, T11) touches no key row.
3. **The box.** The round adds no feature, no migration, and no config key, and it deploys nothing. No task reaches
   the box (`5.78.148.42`), by ssh or otherwise.
4. **Gates.** The fast lane is `bash scripts/gate.sh --scope-changed --base 209821fd12752f4a5a44e783a0fe490df3d566dd`
   (see D2). The close is `bash scripts/gate.sh --merge`. Each runs bare from the worktree root with the Bash tool's
   `timeout: 600000`, re-issued on exit 75 until it prints `gate exit:`. Never pipe a gate. Never commit red.
   `RUN_GATE_IF_BUSY=defer` is never used on T2 or T7 (`auth-data`).
5. **Commits.** One commit per task (a review fix may add one), the task id first in the subject, imperative mood,
   specific files only (never `-A`), ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never
   commit a secret or a `.env` value; read each diff by eye before staging.
6. **Pushes.** Only `git push origin cleanup-store`. The draft PR opens on T1's push: `gh pr create --draft --base
   cleanup-gate`. After each push, watch CI as a background Bash task, bare: `gh pr checks <n> --watch --fail-fast`,
   with `gh run watch <run id> --exit-status` as the fallback. A red stops the line when it arrives. A red outside
   this pass's files is fixed in this branch and listed under `## Decisions` (standing answer 19).
7. **Unchanged behavior.** Every task but T1 and T5's cancel fix is a cut, proven by the tests that existed before
   the pass. A test is deleted only where the review names it (S13), or where it tests only a function the review
   names for deletion (S5, S11); the PR lists each deleted test or subtest beside its finding (item 8). A kept test
   whose call a cut renames keeps every assertion: rename the call, never drop a check.
8. **Records.** PR notes at `<git dir>/pr-notes.md`, first line `launch-base=<sha>`. Each task appends its
   test-first failures, mutation results, review line, gate exits, and anything found outside scope.
9. **Expected values** are literals from this plan or counts computed independently (`grep -c` over the start
   commit), never a figure the change itself printed.
10. **Mutation proofs are run.** Apply each, run the named test, record the failure, revert. One that doesn't fail
    is a finding fixed in the same task.
11. **Smallest thing that meets acceptance.** No abstraction, export, flag, or fallback the design doesn't name.
    Every new export names its production caller in the task that adds it.
12. **Go.** Load `standards:go-conventions` before the first Go edit. "Idiomatic, architecturally beautiful Go" is an
    acceptance dimension of every task. Comments say what the code does; no em dash in a comment or commit (Vale
    reads no Go comment today, BACKLOG #135, so each task runs the grep in T13's check on its changed files).
13. **Layering.** No package gains an import edge. `config` and `identity` import only the standard library; `store`
    imports `atomicfile` and not `config`; `settings` imports `config` and `store`; `backup` imports `store`. The
    import-graph tests (`docs/standing-rules.md`, "Gate legs") and `go list -deps` hold this. T5 (`settings.Row` as
    an alias of `store.SettingsRow`) keeps the edge `settings -> store`, which exists.

## Mechanisms

Each row was read at `209821f`.

| # | Mechanism | Source | What a task relies on |
| --- | --- | --- | --- |
| 1 | `StageFile` copies into a pre-created stage and returns. `copyFile` syncs the file (`out.Sync()`) and never syncs `staging/` | `store/restore.go:328-356`, `395-407` | T1: the bug |
| 2 | The fetch path chmods the stage, then `syncStage` opens it `O_RDWR` (so a FIFO fails) and syncs it, then opens `staging/` and syncs that | `backup/fetch.go:~120`, `129-151` | T1: the seal's steps, moved into store |
| 3 | `verifyStage` re-chmods to 0600 and removes the `-wal` and `-shm` files, and never recreates the stage | `store/restore.go:439-466` | T1: a seal before verify stays valid; `TestRestoreSession_VerifyStageChmods` pins the chmod |
| 4 | backup's restore order: `StageFile` or `fetch` (233, 256), then `VerifyStage` (261), then `Record` (291) | `backup/restore.go` | T1: the line the PR cites for item 1 |
| 5 | `TestFetch_SyncFailureRemovesStage` plants a FIFO, since "fsync on one fails with EINVAL", and asserts `errors.Is(err, syscall.EINVAL)` and the stage removed | `backup/fetch_test.go:550-573` | T1: the one existing seam that observes a sync |
| 6 | `syncDir(dir)` syncs a directory | `store/snapshot.go:205-215` | T1, T6 |
| 7 | `pruneLocal` lists `p.localDir`, skips directories and non-matching names, removes a file whose ModTime is before `p.start`, and joins errors. `localDir` is computed from the stage path | `backup/prune.go:63,163-194`, `backup/backup.go:174-175` | T2 |
| 8 | `parsePreMigrateName` requires the prefix, the `.db` suffix, an `Atoi` version of at least 1, and a round-trip match; `IsPreMigrateCopy` wraps it | `store/snapshot.go:24-31,251-268` | T2 |
| 9 | The ratified store spec: "With backup off, it keeps the newest LOCAL_SNAPSHOT_KEEP copies"; with backup on, copies stay until a backup captures them. `migrate.go:135` prunes only `if !opts.backup`; `TestOpenServe_BackupFieldControlsPruneOnMigrate` pins both | `docs/specs/2026-09-23-rung-6-ops-seam-store.md:97-99,440,1040`; `store/open_test.go:204-243` | T2: `ServeOptions.Backup` stays (D3) |
| 10 | `intentRecord` holds `intentSource` and `*libraryOwed`; `restoreIntent()` rebuilds `RestoreIntent` from them; `Record` builds them from a `RestoreIntent` | `store/intent.go:58-109`, `store/restore.go:578-626`, `store/open.go:371` | T3 |
| 11 | `settings.Row` and `store.SettingsRow` carry the same fields; `EvaluateFacts` copies one into the other; `Facts.SettingsGeneration` has no production reader, and `inspect_test.go:440-441,462-463` assert it | `settings/evaluate.go:~31,132-141`; `store/inspect.go:194-235` | T5 |
| 12 | Four lock entry points: `acquireSharedNonblocking` (a `tryLock` with `LOCK_SH`), `acquireShared` and `acquireExclusive` (both `acquireBounded`), `acquireBounded` (a poll loop), plus `waitLock`'s own poll loop with holder logging. Callers: `open.go:157`, `open.go:250,263`, `inspect.go:277`, `restore.go:137` | `store/lock.go:40-100`, `store/open.go:380-418` | T4 |
| 13 | `waitLock` derives `context.WithTimeout(ctx, opts.MigrationWait)` and on `ctx.Done()` returns "wait %s for the store lock %s: %w" with the busy error, even when the parent was cancelled. `go doc context.Context`: "If Done is closed, Err returns a non-nil error explaining why: DeadlineExceeded if the context's deadline passed, or Canceled if the context was canceled" | `store/open.go:386-418` | T4: the cancel bug |
| 14 | `WriteFile` syncs the temp file and renames; `WriteStream` also syncs the directory | `atomicfile/atomicfile.go:16-43,55-91` | T6 |
| 15 | Per-kind parsing: `validateThresholdDefault` (switch), `buildValidator` (switch over `strconv.Atoi`, `time.ParseDuration`, `parseFiniteFloat`), `parseThreshold[T]` with `ParseInt`/`ParseDuration`/`ParseFloat`/`ParseString`, `settings.parse` (uses `strconv.ParseFloat`), and the accessor's typed reads | `config/registry.go:287-308`, `config/settings.go:131-147`, `config/thresholds.go:64-136`, `settings/evaluate.go:358-373`, `settings/accessor.go:43-62` | T7 |
| 16 | Environment lookup three times with one semantics (a later entry wins; an empty value is unset): `config.get` over the wall, `settings.lookupEnv`, `doctor.lookupEnviron` | `config/config.go:~185-205`, `settings/evaluate.go:344-353`, `cmd/dubplate-server/doctor.go:684-695` | T7 |
| 17 | `doctor.rowStatuses` flattens `loadErr`, tests each leaf with `errors.Is(leaf, config.ErrMissing)`, and splits every line with `strings.Cut(line, ": ")`; `splitKeyReason` does the same at 539, 546, 553, 563 and `serve.go:116`; 622 uses `errors.AsType[*sourceKeyError]`. The comment at 630-634 claims every refusal `validateMagicLinkSource` returns is keyed, which is false | `cmd/dubplate-server/doctor.go:58,475-500,539-563,622-634,697-707` | T7 |
| 18 | The leaves that reach `rowStatuses`: `config.Load`'s wall errors (`config.go:210` `require`, 260, 282, 317, 322, 326, 349, 371, 438) and the source `Validate` errors it appends (388-389), which are `auth.go`'s `sourceKeyError` (349-362, used at 378, 475-498) and its keyed errors (418, 421, 436, 439), `config.go:226` `requireEmail`, `serve.go`'s `validateServePort` (114-126), and `intake.go`'s `resolveStaging` (445, 449) via `doctor.go:58`. `auth.go:297,301` sit in `durationAtLeast`, a correlated rule, and never reach it | as cited | T7: a start list; T7 rebuilds it from `rowStatuses`'s inputs and converts every leaf |
| 19 | `Threshold` has no bounds; `Setting` has `Floor` and `Ceiling` strings enforced by `boundedValidator` and checked at registration | `config/thresholds.go`, `config/settings.go:70-71,131-147` | T8 |
| 20 | `config.OwnedWall()` (`keys.go:94`) and `config.ReservedClasses()` (`keys.go:151`) are package functions with no production caller (doctor uses the `Registry` methods, `doctor.go:79,87,207`); `Registry.Superseded()` (`registry.go:527`) has none | as cited | T9 |
| 21 | `LeaseState` has `LeaseHeld`, `LeaseFree`, `LeaseAbsent`; `TryLease` returns one; the one production caller tests only `state == store.LeaseHeld` | `store/lease.go:34-130`; `internal/ops/local/recovery.go:33-37` | T9 |
| 22 | `stateError{state, msg}` has no `Unwrap`; `refusal` formats with `fmt.Sprintf`; `Inspect` reads the state with `errors.AsType[*stateError]`. `go doc errors.AsType`: "finds the first error in err's tree that matches the type E", the tree being err "followed by the errors obtained by repeatedly calling its Unwrap() error or Unwrap() []error method" | `store/open.go:35-46`, `store/inspect.go:497-498` | T10 |
| 23 | The accessor's `value` locks `a.mu` and holds it across the `settings_generation` query and `readRows` | `settings/accessor.go:87-121` | T12 |

## Recorded searches

Run at `209821f` from the worktree root.

| Search | Hits | Use |
| --- | --- | --- |
| `grep -rn IsPreMigrateCopy internal cmd --include='*.go'` | 4 files: `backup/prune.go`, `backup/prune_test.go`, `store/snapshot.go`, `cmd/dubplate-server/serve_store_test.go` | item 2's start |
| `grep -c 'strings.Cut(line, ": ")' cmd/dubplate-server/doctor.go` | 1 | item 3's start |
| `grep -rn "refusal(StateRefused" internal/engine/store/*.go \| grep -v _test \| wc -l` | 46 (inspect 17, open 17, restore 12) | item 4's frame; the design counts 35 that format a cause with `%v` |
| `grep -c 'ruled, conductor' internal/config/*.go` | config.go 2, keys.go 4, registry.go 8, thresholds.go 1 (15) | item 5's start |
| `grep -n 'arrives at rung' internal/config/*.go` | 1 (`config.go:322`) | item 5 |
| `git diff --stat origin/master -- internal/engine/store/migrations` | empty | item 6 |
| `grep -rn 'errors.Is(\|errors.As' cmd/dubplate-server internal/{backup,auth,ops,serve,settings} --include='*.go'`, non-test, filtered to `store.`, `ErrLockBusy`, `Canceled`, `DeadlineExceeded`, `ErrNotExist`, `ErrNoRows` | 33 | T10: the three that can meet a store refusal are `cmd/dubplate-server/restore.go:195` (`store.ErrLockBusy`), `internal/ops/errors.go:176` (`context.Canceled`), and `internal/ops/admin/admin.go:124` (`store.ErrBusy`, `context.DeadlineExceeded`) |
| `grep -rln 'func syncDir' internal --include='*.go'` | 3: `store/snapshot.go`, `auth/nonce.go`, `intake/gate/zip.go` | T6: the last two are outside the row and stay |
| `grep -ln 'go func\|sync.WaitGroup\|wg.Go' internal/settings/*_test.go` | 1 (`write_test.go`); no accessor test reads concurrently | T12 |

## Tasks

Task order puts the bug first. A pair's two halves have disjoint files; the session merges both, runs the fast lane,
and pushes. Every task's check includes the fast lane at exit 0.

### T1: One seal step for both restore paths (20 min, `ordinary`, session)

Finding S1 and S3 (BACKLOG #116, #117's S3 part). Outcome: `RestoreSession` gains one unexported-or-exported seal
step (exported only because `backup.fetch` calls it; that is its production caller) that sets the stage to mode 0600,
opens it `O_RDWR`, syncs it, closes it, then calls `syncDir(s.leaf.stagingDir)`, with `%w` on each cause. `StageFile`
calls it after the copy, and `copyFile` drops its own `Sync`. `backup.fetch` calls it in place of `syncStage`, which
is deleted with its chmod; a failed seal still removes the stage on both paths. `VerifyStage` keeps its chmod.
`fetch.go:101`'s doc comment, which cites `[syncStage]`, names the seal instead.

Files: `internal/engine/store/restore.go`, `internal/backup/fetch.go`, their tests only where a call is renamed.

Test-first, where a seam exists: before the change, run `TestFetch_SyncFailureRemovesStage` and record it green;
it must stay green after, now through the store's seal. `TestRestoreSession_StageFile` (`restore_test.go:509-531`)
already asserts mode 0600 after `StageFile` and pins the mode through the move; no new test is added. No test observes the directory sync: the store's tests have no seam for it, and the pass adds no
test-only seam (item 1). The PR names the call by file and line: the seal call in `StageFile`, and the
`backup/restore.go` line where `Record` follows `VerifyStage`, so the reviewer can confirm the order.

Mutation: remove the file `Sync` from the seal; `TestFetch_SyncFailureRemovesStage` must fail. Record it.

Check: `grep -n 'func syncStage' internal/backup/fetch.go` prints nothing; `grep -n 'Sync()' internal/engine/store/restore.go`
shows the seal's file sync and no sync in `copyFile`; `go test ./internal/engine/store/... ./internal/backup/...`;
the fast lane.

### T2: One package prunes pre-migrate copies (15 min + 5 review, `auth-data`, session)

Finding S2 (#117). Outcome: `Handle.PrunePreMigrateCopies(before time.Time) error` lists `h.leaf.snapshotsDir`,
skips directories, names `parsePreMigrateName` refuses, and the pre-restore copy, and removes each copy whose ModTime
is before `before`, joining errors as `pruneLocal` does. Its production caller is backup's prune, which drops
`pruneLocal`, its `localDir` field, and `backup.go:174`'s path arithmetic. `IsPreMigrateCopy` is deleted.
`ServeOptions.Backup` stays (D3), so `serve.go` may need no change; say so in the PR if it doesn't.

Files: `internal/engine/store/snapshot.go` (or the file holding the handle's snapshot methods), its test,
`internal/backup/prune.go`, `internal/backup/backup.go`, `internal/backup/prune_test.go`,
`cmd/dubplate-server/serve_store_test.go`.

Tests: `TestPrune_LocalCopies`, `TestPrune_KeepsPreRestoreCopy`, and `TestPrune_LocalCopyNames`
(`prune_test.go:275-318`) stay and pass through the new method. `TestPrune_LocalCopyNames` already asserts each row's
removal effect at 310, so only its direct `IsPreMigrateCopy` check at 307-308 goes; the PR names those lines beside S2. `serve_store_test.go`'s `preMigrateCopies`
helper gets a test-local filter (prefix `pre-migrate-v` and suffix `.db`), and its `wantCopies` literals stay.

Test-first: `auth-data`. Write one store test of `PrunePreMigrateCopies` first (an old copy removed, a new copy and a
pre-restore copy kept, named by literal file names) and record it failing to compile at the start commit. Mutation:
make the method keep every file; `TestPrune_LocalCopyNames` and the store test must fail.

Check: `grep -rn IsPreMigrateCopy internal cmd --include='*.go'` prints nothing; `grep -rn 'PrunePreMigrateCopies'
internal/backup --include='*.go' | grep -v _test` prints one call; the per-task `diff-reviewer` reads the commit.

### T3 and T6: a pair (25 min + 5 merge)

#### T3: One model of a restore's intent (25 min, `ordinary`, pair half)

Finding S4 (#117). Outcome: `RestoreIntent` becomes a field of `intentRecord`, with json tags; `intentSource`,
`libraryOwed`, and `restoreIntent()` go, and `Record` stores the caller's `RestoreIntent` as given. The library half
is owed exactly when the record says so today (`LibraryRoot != ""` stands where `Library != nil` stood); keep every
state's meaning. The record is decoded with `DisallowUnknownFields` (`intent.go:119`), and today's nested `source`
and `library` objects cannot come from one flat struct, so the wire shape changes to flat keys (D8). The design
allows it ("free today because no store is live"); `Format` keeps its value. Other sites that read the old fields:
`open.go:371`, `restore.go:764-774` (`Intent()`), `restore.go:820` (`r.Library != nil`), and the record's validation
in `intent.go`, whose refusal texts (`intent.go:159-163`) name the new keys.

Files: `internal/engine/store/intent.go`, `internal/engine/store/restore.go`, `internal/engine/store/open.go` (line
371 only), `internal/engine/store/intent_test.go`, the 14 fixtures in `internal/engine/store/testdata/intent/`, and the
forced intent-record JSON literals in `cmd/dubplate-server/settings_test.go:307-309`, `store_test.go:160-170`,
`member_test.go:142-144`, and `doctor_test.go:857-859` (D8).

Tests: every fixture and literal moves to the new keys with the same values, and every test keeps its assertion. The
wantErr literals at `intent_test.go:100-101` take the new refusal text and still assert the same refusal; the struct
literals at `intent_test.go:38,70,86-89,354,485,487` take the new field paths. The PR lists each changed literal.

Check: `grep -n 'intentSource\|libraryOwed\|restoreIntent()' internal/engine/store/*.go` prints nothing;
`grep -rln '"source": {' internal/engine/store/testdata cmd/dubplate-server` prints nothing; the store and cmd tests
pass.

#### T6: One atomic write path (10 min, `ordinary`, pair half)

Finding S7 (#117). Outcome: `atomicfile.WriteFile` is implemented on `WriteStream`, so it also syncs the directory.
Its signature and doc stay apart from saying it syncs the directory.

Files: `internal/atomicfile/atomicfile.go`, its test.

Check: the existing `atomicfile` tests pass; `go test ./...` passes for every importer (the fast lane).

### T4 and T12: a pair (20 min + 5 merge)

#### T4: One bounded lock wait, which reports a cancel as a cancel (20 min, `ordinary`, pair half)

Finding S6 (#117). Outcome: `lock.go` keeps `tryLock` and one bounded wait. `acquireSharedNonblocking`,
`acquireShared`, `acquireExclusive`, and `acquireBounded` go; `inspect.go:277` and `restore.go:137` call `tryLock`,
and `open.go:157`, `250`, and `263` call the one wait. The wait takes an optional bound (zero means the caller's
context alone, as `OpenWrite` at `open.go:157` needs, since it has no `MigrationWait`) and an optional logger for the
holder lines `waitLock` logs today. With a zero bound it derives no child context, so `blockedSignalingCtx`
(`lock_test.go:17-33`) still signals at the first blocking `Done`. When the parent context ends, the wait returns
`fmt.Errorf("...: %w: %w", ctx.Err(), ErrLockBusy)`: `errors.Is` finds both the parent's cause and `ErrLockBusy`, so
the busy assertions at `lock_test.go:93-109`, `open_test.go:354`, and `open_test.go:1098` stay true, and only the
wait's own timeout carries the "wait %s for the store lock" text and the holders. `cmd/dubplate-server/restore.go:193-196`
tests `ctx.Err()` before `ErrLockBusy`, so its answer does not move.

Files: `internal/engine/store/lock.go`, `internal/engine/store/open.go`, `internal/engine/store/inspect.go` (line 277
only), `internal/engine/store/restore.go` (line 137 only, if it changes), and the tests whose calls are renamed
(`lock_test.go:80,100,123,133,142`, `open_test.go:354,1098`, `restore_test.go:355`), which keep every assertion.

Test-first: the cancel fix is a bug. Write one test that holds the lock, starts the wait with a parent context it
then cancels, and asserts `errors.Is(err, context.Canceled)`, `errors.Is(err, ErrLockBusy)`, and that the error text
does not contain `"wait "`; record it failing at the start commit. Mutations: return the timeout's text on every
`ctx.Done()`, and drop `ErrLockBusy` from the parent-end error; the new test fails on the first, and the new test and
`lock_test.go:93-109` fail on the second.

Check: `grep -cE 'func (acquireShared|acquireExclusive|acquireBounded|acquireSharedNonblocking)\b'
internal/engine/store/lock.go` prints 0; the store tests pass.

#### T12: The settings accessor reads outside its mutex (10 min, `ordinary`, pair half)

BACKLOG #26's residue. Outcome: `value` reads the generation, then the rows when the cached generation differs,
without holding `a.mu`, and takes the mutex only to read or swap the cache. The comment on read order stays true:
rows are cached under the generation read before them, so a racing older swap is reloaded by the next read.

Files: `internal/settings/accessor.go`, `internal/settings/accessor_test.go`.

Test: add one test that runs reads from several goroutines while another handle writes, and asserts each read
returns either the old or the new literal value. Run `go test -race ./internal/settings/...` and record its exit.
That test cannot fail on the mutex's scope; item 7's proof of scope is the function body. The read-order argument
holds without the mutex: each cached row set was read after the generation it is tagged with, and generations only
rise, so a stale swap is reloaded by the next read. The PR states this and does not claim the race run proves it.

Check: item 7: `sed -n '/^func (a \*Accessor) value/,/^}/p' internal/settings/accessor.go` shows no `QueryRow` or
`readRows` call between a `Lock` and its `Unlock`; `go test -race ./internal/settings/...` exits 0.

### T5: One settings row type (10 min, `ordinary`, session)

Finding S5 (#117), the review's cheap form. Outcome: `settings.Row` becomes `type Row = store.SettingsRow`, and
`EvaluateFacts` drops its field-by-field copy. The field docs on `settings.Row` (`evaluate.go:30-44`) move onto
`store.SettingsRow` (`inspect.go:225-235`). `Facts.SettingsGeneration` is deleted with its read in
`readSettingsFacts`, and its two assertions at `inspect_test.go:440-441,462-463` go with it; the PR names them beside
S5. Any other test of the field is a stop.

Files: `internal/settings/evaluate.go`, `internal/engine/store/inspect.go`, `internal/engine/store/inspect_test.go`.

Check: `grep -rnw SettingsGeneration internal cmd --include='*.go'` prints nothing (`-w` keeps
`TestSchemaSettingsGenerationSingleton` out); the settings and store tests pass.

### T7: One parse, one lookup, one keyed error (45 min + 5 review, `auth-data`, one `opus` subagent)

Findings S8 and S10 (#118, #119), with design call 12. Outcome:

- `ThresholdKind` gains one parse method returning the parsed value or an error; `validateThresholdDefault`,
  `buildValidator`, `parseThreshold`, and the accessor's per-kind parsers use it, and floats there go through
  `parseFiniteFloat`. `settings.parse` (`evaluate.go:358-373`) keeps `strconv.ParseFloat` for floats: `outcome`
  (`evaluate.go:265`) parses stored rows that never passed `Validate` and heals an out-of-bounds row (comment at
  250-253), so a finite-only parse there would refuse a stored `Inf` row it heals today. Its int, duration, and string
  arms use the kind's method. The PR records this under `## Decisions` (D9).
- `config.Lookup(environ []string, key string) (string, bool)` is exported with the wall's semantics (a later entry
  wins; an empty value is unset). `settings.lookupEnv` is deleted and its callers (`evaluate.go:256`,
  `write.go:111,159`) call `config.Lookup`. `doctor.lookupEnviron` keeps its name and becomes a one-line call to
  `config.Lookup`: it has eleven callers in pass 5's files (`backup.go:430,438,503,512`, `unidecode.go:82,147`,
  `store.go:275,298,470,488`, `settings.go:62`), and pass 5 inlines them (D10, filed). `config.get`
  (`config.go:183-202`) stays on the `w.env` map that `rejectNamespace` walks; its comment says the map and `Lookup`
  share one rule. `Lookup`'s production callers are `settings` and `doctor.go`.
- `config.KeyError{Key, Reason string}` with an `Err` field for the wrapped sentinel (so `errors.Is(err,
  config.ErrMissing)` still holds) and `Unwrap`. Its `Error()` prints `Key + ": " + Reason` byte for byte, so every
  error text a test or a member sees is unchanged. First, rebuild mechanism 18's list from `rowStatuses`'s real
  inputs: every error `config.Load` returns or appends into doctor's `loadErr`, and every leaf `doctor.go:58` adds.
  The planner's list is a start, not the answer: `config.Load`'s wall and parser errors (`require` at 210,
  `requireEmail` at 226, and 260, 282, 317, 322, 326, 349, 371, 438), `auth.go`'s `sourceKeyError` (deleted; its
  uses become `config.KeyError`) and its keyed errors at 418, 421, 436, 439, `serve.go`'s `validateServePort`, and
  `intake.go`'s `resolveStaging` errors at 445 and 449. `auth.go:297,301` sit in `durationAtLeast`, a correlated
  rule that never reaches `rowStatuses`, and stay. Each leaf on the rebuilt list returns a `KeyError`; the `auth.go`,
  `intake.go`, and `doctor_test.go` sites are forced and go under `## Decisions`. `validateSMTPPort` and
  `validateSMTPFrom` become `KeyError`s too, so the `errors.AsType` at `doctor.go:622` now matches them: confirm with
  a doctor output test that the earlier dedup sweep leaves each reported once.
- `doctor.go`'s `rowStatuses` reads the key and reason with `errors.AsType[*config.KeyError]`, and `splitKeyReason`
  and the `strings.Cut` split go. The comment at 630-634 is corrected to what the code does. A leaf that is not a
  `KeyError` is reported the way `rowStatuses` reports an unkeyed line today; if no such path exists after the
  change, say so in the PR.

Files: `internal/config/{config.go,thresholds.go,registry.go,settings.go}` and their tests,
`internal/settings/{evaluate.go,accessor.go,write.go}`, `cmd/dubplate-server/doctor.go`, `cmd/dubplate-server/serve.go`,
and the forced sites `cmd/dubplate-server/auth.go`, `cmd/dubplate-server/intake.go`, and
`cmd/dubplate-server/doctor_test.go:817-832` (its fixture wraps `&sourceKeyError{...}`; it wraps a `config.KeyError`
and keeps its assertion). `web.go:270,278,288,549,651` and `intake.go:325` build keyed text at route or bundle build
time, not through `config.Load`, and stay; the PR says so.

Tests: new tests for `config.Lookup` (later wins, empty unset, absent) and `config.KeyError` (`Error()` text against a
literal, `errors.Is` to `ErrMissing`, `errors.AsType`). Test-first: `auth-data`; write them first and record them
failing to compile at the start commit. Every `doctor_test` output test passes unchanged: that is the proof doctor's
output did not move. Mutation: swap `Key` and `Reason` where `rowStatuses` reads them; a doctor output test must
fail. (A mutation of `Error()` would not reach doctor once `rowStatuses` reads the fields.)

Check: item 3: `grep -c 'strings.Cut(line, ": ")' cmd/dubplate-server/doctor.go` prints 0; `grep -rn
'func lookupEnv\b\|type sourceKeyError\|func splitKeyReason' internal cmd --include='*.go'` prints nothing; the
per-task `diff-reviewer` reads the commit.

### T8 and T10: a pair (20 min + 5 merge)

#### T8: `Threshold` carries bounds (15 min, `ordinary`, pair half)

Finding S9 (#118). Outcome: `Threshold` gains `Floor` and `Ceiling` strings, as `Setting` has. Config's `Parse*`
refuse a value outside them with a `KeyError` whose reason names the bound, and registration refuses a threshold
whose default falls outside its bounds, the way `Setting` is checked. Their reader this pass is config itself, by the
design's split: pass 5 sets them on the `cmd` declarations and removes the hand checks (`store.go:34-53,133-180`,
`serve.go:98-126`), so no declaration sets them here (D6).

Files: `internal/config/thresholds.go`, `internal/config/registry.go`, their tests.

Tests: a table over each kind with a value below, at, and above each bound, expected results as literals.

Check: the new table passes; every existing config test passes.

#### T10: Every `StateRefused` refusal keeps its cause (20 min, `ordinary`, pair half)

Finding S12 (#117). Outcome: `stateError` keeps the formatted error it was built from and gains `Unwrap`, so a
refusal built with `%w` keeps its cause in the chain; `refusal` keeps its signature for every state. Two mechanics:
`refusal` builds its error with `fmt.Errorf`, since `go vet` flags `%w` through a `fmt.Sprintf` wrapper; and `Unwrap`
returns that stored error itself, never `errors.Unwrap` of it, which is nil for a site with two `%w` verbs.
`withError` (`inspect.go:494-501`) still finds the outer `stateError` first, since `errors.AsType` searches the tree
in pre-order. Each of the 35
`refusal(StateRefused, …)` sites that formats a cause with `%v`, continuation lines included, uses `%w`. The PR names
any refusal site it removes.

Before the change, read the three handlers in "Recorded searches" (`restore.go:195`, `ops/errors.go:176`,
`ops/admin/admin.go:124`) and record in the PR whether a cause newly reachable through `Unwrap` (`context.Canceled`,
`context.DeadlineExceeded`, `store.ErrBusy`, `store.ErrLockBusy`, `fs.ErrNotExist`) flips any branch. A flip that
changes an answer is fixed at its cause in this task or is a stop.

Files: `internal/engine/store/{open.go,inspect.go,restore.go}` and their tests.

Test: one test that builds a `StateRefused` refusal over a sentinel and asserts `errors.Is` finds the sentinel and
`Inspect`'s state read still sees `StateRefused`. Mutation: drop `Unwrap`; the test must fail.

Check: item 4: `grep -rn -A1 'refusal(StateRefused' internal/engine/store/*.go | grep -v _test | grep -c '%v'`
prints 0.

### T9: Uncalled exports and unused lease states go (10 min, `ordinary`, session)

Finding S11 (#118). Outcome: the package functions `config.OwnedWall` and `config.ReservedClasses` are deleted.
`Registry.Superseded` stays (D11): `cmd/dubplate-server/intake_test.go:67` and `templatekeys_test.go:95` read it for
real assertions, and both are pass 5's files. `LeaseState` goes: `TryLease` returns whether the lease is held as a `bool` (the
name is the executor's call under go-conventions), and the forced site is `internal/ops/local/recovery.go:33-37`.

Files: `internal/config/keys.go` (the two functions only), `internal/engine/store/lease.go`, their tests, and the
forced site.

Tests: a test that calls `OwnedWall()` or `ReservedClasses()` to reach the
table for another assertion (`TestOwnedWall_RequiredNowRows` at `keys_test.go:14`, and `describe_test.go:93`) reads
the table through `NewRegistry()`'s methods or the package var instead and keeps every assertion. The two
`slices.Clone` tests of the deleted functions (`TestOwnedWall_ReturnsCopy` at `keys_test.go:30`,
`TestReservedClasses_ReturnsCopy` at `keys_test.go:97`) go here and are named beside S11 and S13. The `LeaseFree` and `LeaseAbsent` assertions in `lease_test.go:169-180` and `212` become `false`
assertions. The loop at 169-180 already stats the file after each call; it gains a stat before the first call that
asserts the file exists, so the Free-then-Absent distinction is still checked and no kept test loses an assertion.

Check: `grep -rn 'func OwnedWall\|func ReservedClasses\|LeaseState' internal cmd --include='*.go'` prints nothing.

### T11: The named tests and the trailers go (15 min, `ordinary`, session)

Findings S13 and S14 (#118). Outcome: delete exactly these, after reading each to confirm it matches the finding:
`TestThresholds_ReturnsCopy` (`thresholds_test.go:37`, a duplicate of `registry_test.go:83-89`),
`TestSettings_ReturnsCopy` (`settings_test.go:42`), the other two of the four `slices.Clone` tests having gone in
T9; `TestNewRegistry_CarriesRung1Declarations`
(`registry_test.go:31-41`); and the 2^-64 check in `schema_test.go:340-361` (the migrate, insert, and rowid block at the end of
`TestID`, not the whole test; its doc comment at 327-330 drops the rowid sentence, and the length and
distinctness checks stay). Remove the 15 `// ruled, conductor` trailers, the 12-line deleted-rows comment at `keys_test.go:47-58` (lines 39-45
are a live gitleaks note and stay),
and "arrives at rung %d" from `config.go:322`'s error text, so the error reads "reserved for %s". The comment that
remains says what the code does.

Files: `internal/config/*.go`, `internal/config/*_test.go`, `internal/engine/store/schema_test.go`.

Check: item 5: `grep -c 'ruled, conductor' internal/config/*.go` prints 0 for each file; `grep -rn 'arrives at rung'
internal --include='*.go'` prints nothing; a test that asserts the 322 text is updated to the new literal and named
in the PR.

### T13: The close (35 min, session)

1. `bash scripts/gate.sh --merge` on the head, bare.
2. Em-dash check over the branch's Go files: `git diff --name-only 209821fd12752f4a5a44e783a0fe490df3d566dd...HEAD
   -- '*.go' | xargs grep -n '—'` prints nothing.
3. Whole-branch `standards:diff-reviewer` against the design and this plan, loading `go-conventions`, and one
   `standards:web-auth-security-reviewer` read on the `cmd/dubplate-server` diff and the restore seal. Both are asked
   for correctness gaps ranked by consequence. One batched fix chain, quoting each finding verbatim; a second read
   only on a blocker. A security read short of `accept` after the fold is a stop.
4. `bash scripts/gate.sh --merge` again on the final head if the fold changed anything.
5. Close this pass's BACKLOG items (#116, #117, #118, #119, #26's residue) with the PR link; file each find, and
   file for pass 5: inline `doctor.lookupEnviron` (D10), decide `Registry.Superseded` (D11), and the rung text in
   `doctor.go:88` and `doctor_test.go:163`, which item 5's grep over `internal` does not reach.
6. `git mv docs/plans/2026-11-15-cleanup-store.md docs/plans/2026-11-15-cleanup-store-prompt.txt docs/archive/plans/`.
7. The PR body: an opening paragraph per `~/.claude/docs/voice/written-for-people.md`, `Round: cleanup`, `Pass:
   cleanup-store`, the `Review:` and `Calls:` lines, evidence for every acceptance item, each deleted test beside its
   finding, `## Decisions`, `## Planning misses`, `## STATUS text`, the score, the land test's result (waits for the
   round's close packet), and the closing `🤖 Generated with [Claude Code](https://claude.com/claude-code)` line.
8. Mark the PR ready once the head's check is green.

## Standing answers

| # | Stop | Answer |
| --- | --- | --- |
| A1 | A cut would change a test's assertion, not only its call | Keep the assertion. If the cut cannot keep it, leave that cut, file it, and say so in the PR. |
| A2 | T10's read finds a branch that flips on a newly reachable cause | Fix it at the cause in T10 if the fix stays in the row; a fix outside the row is a stop. |
| A3 | T7 finds a leaf reaching `rowStatuses` outside mechanism 18 | Convert it as a forced site and list it under `## Decisions`. |
| A4 | T1's stage test fails at the start commit because `StageFile` leaves another mode today | Record it as test-first evidence; the seal fixes it. |
| A5 | CI red outside the pass's files | Fix in the branch, list under `## Decisions` (standing answer 19). |
| A6 | `--base 209821f…` falls back to the full form | Expected only when the change set holds machinery, which no task touches. Record the line and go on. |

## Decisions for the executor

- D1: The PR's base is `cleanup-gate`, stacked on PR #13 (conductor note). The close packet merges #13 first.
- D2: The fast lane names `--base 209821fd12752f4a5a44e783a0fe490df3d566dd`. Without it, `check.sh` measures from
  the merge base with `master` (`7db541b`), which pulls pass 2's machinery (`scripts/*`, `Makefile`, `go.mod`,
  `.gitleaks.toml`) into the change set and promotes every scoped run to the full form. The adapter's "other passes
  pass no `--base`" assumes a branch cut from master; this branch is not.
- D3: `ServeOptions.Backup` stays, against S2's text. The ratified store spec (`2026-09-23-rung-6-ops-seam-store.md`
  97-99, 440, 1040) keeps the newest `LOCAL_SNAPSHOT_KEEP` copies only when backup is off, and
  `TestOpenServe_BackupFieldControlsPruneOnMigrate` pins it. Dropping the flag would change behavior and delete a
  test the review does not name. Listed for Geoff at the close.
- D4: T1's seal is one `RestoreSession` method in store, called by `StageFile` and by `backup.fetch`. It is exported
  because backup calls it.
- D5: `config.KeyError` wraps its sentinel in an `Err` field so `errors.Is(err, config.ErrMissing)` holds, and its
  text is byte-identical to today's `"<key>: <reason>"`.
- D6: `Threshold.Floor` and `Ceiling` are read by config's parsers this pass and set by pass 5's declarations, by the
  review's split of S9.
- D8: The restore intent record's wire shape goes flat (`remote`, `snapshot`, `file`, `library_root`,
  `into_nonempty`) because one tagged `RestoreIntent` cannot emit today's nested objects. The alternative, custom
  `MarshalJSON` and `UnmarshalJSON` that rebuild the nesting, is more code than the duplicate model S4 removes. The
  store's fixtures and four cmd test literals change with it (T3); no store is live, per the design.
- D9: `settings.parse` keeps `strconv.ParseFloat` for floats, so a stored non-finite row is still healed as it is
  today; every other float parse is finite-only (T7).
- D10: `doctor.lookupEnviron` stays as a one-line call to `config.Lookup`, because its other ten callers are in pass
  5's files. Pass 5 inlines them; filed in BACKLOG at T13.
- D11: `Registry.Superseded` stays, against S11's text: two pass-5 cmd tests read it for real assertions. Filed for
  pass 5, which owns those tests.
- D7: BACKLOG #78 is not taken (see "Planning misses"). #129 (sqlite v1.62 needs `go.mod`) is not taken: `go.mod` is
  pass 2's machinery file and outside this row; it goes back untagged for the next round's sweep. #133 (the
  `TestBusyTimeout` flake) is not planned; if it reds a run, standing answer 19 covers it.

## Path

| Step | Minutes |
| --- | --- |
| T1 | 20 |
| T2, with its review | 20 |
| T3 and T6, merged | 30 |
| T4 and T12, merged | 25 |
| T5 | 10 |
| T7, with its review | 50 |
| T8 and T10, merged | 25 |
| T9 | 10 |
| T11 | 15 |
| T13 | 35 |
| Total | 240 `[estimate]` |

The design's figure is 130. This path is under the 260-minute park line, so the pass runs; the PR records the
difference under `## Decisions` with the actual.

## Acceptance

Copied verbatim from the design's pass 3 list, plus the two every-pass items.

| # | Item | Task | Evidence |
| --- | --- | --- | --- |
| 1 | A `--from-file` restore syncs the `staging/` directory before the intent record names the stage. The PR names the call by file and line, and the reviewer confirms the order. A test proves it only where a seam the store's tests already use can observe the sync; the pass adds no test-only seam to production code. Today: `store/restore.go:403` syncs the file only. | T1, T13 | the seal call's file and line, the `Record` line, the reviewer's order line, `TestFetch_SyncFailureRemovesStage` and its mutation |
| 2 | One package prunes pre-migrate copies: `grep -rn IsPreMigrateCopy internal cmd --include='*.go'` prints nothing, and `Handle.PrunePreMigrateCopies` has a production caller. Today: four files match, one of them `cmd/dubplate-server/serve_store_test.go`. | T2 | the two greps |
| 3 | `config.Lookup` and `config.KeyError` exist with tests, `Threshold` carries `Floor` and `Ceiling`, and `doctor.go` reads keys from `config.KeyError`: `grep -c 'strings.Cut(line, ": ")' cmd/dubplate-server/doctor.go` prints 0. Today: none of the three exists, and the count is 1. | T7, T8 | the grep, the new tests, the struct |
| 4 | No refusal drops its cause. Each of the 35 sites that format a cause with `%v` wraps it with `%w`, continuation lines included, and the PR names any refusal site it removes. Today: 35 of 46 sites. | T10 | the `-A1` grep at 0 |
| 5 | `grep -c 'ruled, conductor' internal/config/*.go` prints 0 for each file, and no error text says "arrives at rung". Today: 15 trailers and one error line (`config.go:322`). | T11 | both greps |
| 6 | `git diff --stat origin/master -- internal/engine/store/migrations` prints nothing. | T13 | the command's empty output |
| 7 | The settings accessor no longer holds its mutex across a database read, and its tests pass under `go test -race`. Today: `internal/settings/accessor.go:88-112` does. | T12 | the function body and the race run's exit |
| 8 | Every test the pass deletes is named in the PR with the review finding that names it. No other test is deleted, and the reviewer confirms that no kept test lost an assertion. | T2, T5, T9, T11, T13 | the PR's list; `git diff 209821f...HEAD -- '*_test.go' \| grep '^-func Test'` against it; the reviewer's line |
| 9 | The security reviewer returns accept. | T13 | its verdict |
| 10 | The PR's check is green on its head. | T13 | `gh pr checks <n>` |
| 11 | `bash scripts/gate.sh --merge` exits 0 on the close head. | T13 | the `gate exit: 0` line |

## Planning misses

- `[stale-fact]` BACKLOG #78 places doctor's address normalizer in `doctor.go`; it now sits in
  `cmd/dubplate-server/store.go:296-303` (`adminAddress`), beside `config.requireEmail` (`config.go:218-229`). Moving
  both onto `identity.NormalizeAddress` changes what `Load` accepts, which is a behavior change this round does not
  make, and `store.go` is pass 5's file. It is left and its BACKLOG text reworded for pass 5.
- `[conflict]` The base prompt's fast lane has no `--base`; the stacked branch makes every such run full (D2).
- `[conflict]` S2 says to drop `ServeOptions.Backup`; the ratified spec and an unnamed test keep it (D3).

## Review record

One `standards:plan-reviewer` read (opus at medium) on 2026-11-16, against `209821f`. Verdict: not ready, four
blockers, all from incomplete call-site and test-site lists. Each blocker was checked by grep before folding.

| Finding | Fold |
| --- | --- |
| B1: `doctor.lookupEnviron` has ten callers in pass 5's files, `settings.lookupEnv` two in `write.go`, and `sourceKeyError` one in `doctor_test.go` | Taken: `lookupEnviron` stays as a delegate (D10), `write.go` and `doctor_test.go` join T7's files |
| B2: the intent record is strict JSON, so T3 breaks 14 fixtures, wantErr texts, struct literals, and four cmd JSON literals | Taken: flat wire shape with every fixture and literal listed as forced (D8). The reviewer's "tags that reproduce today's shape" cannot emit nested objects from a flat struct |
| B3: T4's cancel rule flips three `ErrLockBusy` assertions, and a child context defeats `blockedSignalingCtx` | Taken: the parent-end error wraps both `ctx.Err()` and `ErrLockBusy`; a zero bound derives no child; signature spelled out |
| B4: `Registry.Superseded` has two cmd test callers and one in a test T11 deletes after T9 | Taken: `Superseded` stays (D11), so T9 compiles in order |
| M1: `settings.parse` reads unvalidated rows, so finite-only parsing changes healing | Taken: `settings.parse` keeps `strconv.ParseFloat` (D9) |
| M2: T7's mutation misses doctor; mechanism 18 lists `auth.go:297,301` and misses `config.go:226` | Taken: mutate `rowStatuses`; T7 rebuilds the list from its inputs; SMTP dedup confirmed by a doctor test |
| M3: T2's table move and T9's lease rewrite would lose assertions | Taken: `TestPrune_LocalCopyNames` stays and loses only 307-308; the lease loop gains a pre-call stat |
| M4: `%w` through `Sprintf` fails vet; `errors.Unwrap` loses two-`%w` causes | Taken: both mechanics stated in T10 |
| Minor 1: T5's grep matches a schema test; field docs lost | Taken: `-w`; docs move to `store.SettingsRow` |
| Minor 2: `keys_test.go:40-58` cuts a live gitleaks note | Taken: 47-58 only |
| Minor 3: T1's new test duplicates `TestRestoreSession_StageFile` | Taken: cite it, add none; `fetch.go:101` doc noted |
| Minor 4: the race test cannot fail on the lock's scope | Taken: T12 says what proves item 7 |
| Minor 5: `config.get` over `Lookup` re-scans | Taken: `get` stays on its map |
| Notes 6 to 8 (web.go sites stay, doctor rung text, T6 importers) | Taken: T7 text, T13 step 5 filing; T6 unchanged |

Path after the fold: 240 minutes, under the 260 park line.
