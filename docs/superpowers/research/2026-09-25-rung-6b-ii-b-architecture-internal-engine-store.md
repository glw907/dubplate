# `internal/engine/store` architecture read (6b-ii-b merge side)

Read at HEAD `960e760`, no plan in context, per CLAUDE.md's "Reviewers" (one
`go-architecture-reader` per touched Go package at each pass's merge).

**Verdict: workmanlike.** Density 0.381 against `database/sql`'s 0.424.

## Findings

- **Exported ahead of caller.** `OpenServe`/`ServeOptions` (`open.go:83-116,
  219`) have no production caller; `ServeOptions.Reconcile` is set only by a
  test, so `StepServe` has no production finisher. **Routed:** settle the
  shape at 6d before serve consumes it.
- **Correctness-adjacent.** Backup's `isPreMigrateName` (`prune.go:200-210`)
  disagreed with store's own `parsePreMigrateName` (`snapshot.go:251-265`) on
  `"pre-migrate-v01.db"`. **TAKE before merge:** export one shared predicate,
  `IsPreMigrateCopy`, and unexport `PreMigratePrefix`/`Suffix` and `dbSuffix`
  (generic `.db`, no reason to export it). Landed at `9870678`.
  `PreRestorePrefix` is referenced only from a test outside its package;
  unexported in the same commit.
- **Duplication.** Two reconcile finishers (`open.go:362-371` vs
  `restore.go:866-889`); two flock poll loops (`open.go:378-410` vs
  `lock.go:64-79`); `acquireExclusive` is test-only (`lock.go:57`); handle
  assembly is repeated at four call sites, wanting one `openHandle(leaf,
  version, busy)`; `OpenRestore`/`OpenServe` share a preamble that wants a
  `prepareLeaf` helper; `readRecord` re-decodes rather than reusing a shared
  path (`restore.go:180-195`); `manualRecovery` should be inlined;
  `RestoreIntent` and its record type mirror each other; the newer-store
  message is written out twice (`restore.go:694,858`); `validate` restates
  `stepOf`; `checkSource`'s sidecar loop wants a shared `presentSidecars`
  helper.
- **File split.** `restore.go` runs 919 lines, split by deliverable (which
  task added it) rather than by concern. `open.go` is a grab bag.
  `BackupRemote` sits away from `writeBackupRef`, which it belongs beside.
  `preCreate` lives in `fsprobe.go`, a mismatched home.
- **Comments.** `doc.go` has no restore section, though restore is now most
  of the package. **TAKE before merge:** added. Landed at `a66b7d1`. Process
  notes are repeated at `lock.go:103-104,136`, `fsprobe.go:10`,
  `leaf.go:57-62,118-120`, `inspect.go:165-167`; the no-environment rationale
  is stated four separate times.
- **Tests.** `swap_test.go` computes its expected discarded paths from the
  code under test rather than a pinned literal; pin one literal per family
  to avoid a test that can't fail on a shared regression.

## Best code in the package

- `swap.go:126-213`: `swapPlan` built entirely from existing files.
- `write.go:98-135`: the `Tx` first-error latch.
- `open.go:212-269`: the flock conversion re-read.

## Routing

`IsPreMigrateCopy` and the restore doc section landed pre-merge (`9870678`,
`a66b7d1`). `OpenServe`/`Reconcile`'s shape routes to 6d. Everything else
routes to the architecture cleanup pass.
