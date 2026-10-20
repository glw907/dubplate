# `cmd/dubplate-server` restore surface architecture read (6b-ii-b merge side)

Read at HEAD `960e760`, no plan in context, per CLAUDE.md's "Reviewers" (one
`go-architecture-reader` per touched Go package at each pass's merge).

**Verdict: sound with nits.** `restore.go` density 0.184.

## Findings

- **Correctness.** A stdout failure swallowed the secret-rotation instruction
  along with the rest of the report (`restore.go:201-204`), and the report
  file itself never carries the rotation line. **TAKE before merge:** repeat
  the instruction on stderr so a lost stdout doesn't lose it. Landed at
  `00739d3`.
- **Duplication, already drifted.** `restoreFlags.check` (`restore.go:45-99`)
  and backup's own `RestoreConfig.check`
  (`internal/backup/restore.go:103-121`) had diverged: the CLI refused
  `--discard-current` with `--resume`, but backup silently dropped
  `DiscardCurrent` on resume instead of refusing the combination
  (`:181-183`). **TAKE before merge:** `backup.Restore` now refuses the same
  combination the CLI does. Landed at `418a8aa`. Other duplication, carried:
  exit constants restated in both files (`restore.go:23-26` vs
  `backup.go:71-75`); two separate SIGHUP mechanisms (`restore.go:169-173`,
  `backup.go:183-210`), which want one path through
  `executeWithSignalContext` (`main.go:97`); `restoreRotation` spells the
  four secret names a fifth time instead of reusing `serve.go:27-30`'s
  constants; a third copy of the `parseErr`/`rangeErr` unpack
  (`restore.go:154-160`); the test-side exit mapping is copied twice
  (`restore_test.go:425-434`, `backup_test.go:471-480`).

## Best code in the package

- `restore.go:192-200`: checks `ctx.Err()` before `ErrLockBusy`.
- `restore.go:175,208-222`: the best-effort writer.
- `restore.go:71-74,146-149`: the flag check runs before load, pinned by a
  test.

## Routing

The rotation-on-stderr fix and the discard/resume drift closure (backup now
refuses the combination too) landed pre-merge (`00739d3`, `418a8aa`).
Everything else routes to the architecture cleanup pass.
