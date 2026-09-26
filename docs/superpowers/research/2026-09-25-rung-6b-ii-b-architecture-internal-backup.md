# `internal/backup` architecture read (6b-ii-b merge side)

Read at HEAD `960e760`, no plan in context, per CLAUDE.md's "Reviewers" (one
`go-architecture-reader` per touched Go package at each pass's merge).

**Verdict: sound with nits.** Density 0.288 against `os/exec`'s 0.670 and
`archive/zip`'s 0.258.

## Findings

- **Exported ahead of caller.** `Report`, `Reconcile`, and `(*Report).Text`
  (`reconcile.go:31,55,149`) have no caller: `store.OpenOptions.Reconcile` is
  never set outside a test. **TAKE before merge:** unexport now, per the house
  rule; re-export with serve's hook at 6d. Landed at `bd34f9a`.
- **Cross-package duplication.** `remote.go:31`'s `validName` restates
  `cmd/dubplate-server/backup.go:374`'s `firstInvalidRemoteNameRune` (itself
  re-cut at `:328`); `ParseRemote` names the rune and `Remote.Name()` already
  exists to share it. `backup.go:132` and `prune.go:200` each restate the
  store leaf's layout by hand.
- **File split.** `restore.go` holds three unrelated concerns in one file: the
  flow, the library half, and the report step. Two unrelated `checkLibrary`
  functions exist (`guard.go:46`, `restore.go:505`); the library half belongs
  in its own `library.go` with the source-side one renamed
  `source.checkLibrary`. The failure type belongs beside `Job.Run`, not where
  it currently sits.
- **Comments.** `doc.go:1`'s package doc describes backup only, though the
  package now also owns restore. **TAKE before merge:** corrected. Landed at
  `a66b7d1`. `restore.go`'s step-number citations (`:222,:282,:314,:367,:432,
  :498,:622`) name plan step numbers instead of what each step does. **TAKE
  before merge:** reworded. Landed at `dcb1f47`.
- **Within-package duplication.** `Text`/`summary` both restate a `render
  (limit)` preamble; `withCode` and `newTransport` are near-identical helper
  shapes; `Reconcile` restates root classification and drops
  `resolveRoot`'s not-a-directory check (carried: call `resolveRoot` instead).
- **Other.** `signalNames`'s table restates `unix.SignalName` (a judgment
  call, not a clear fix); `RestoreConfig.Now` is a test-only seam (a package
  var `now` would drop the field).

## Best code in the package

- `rclone.go:143`: classifies `ctx.Err()` before any other error path.
- `prune.go:83`: `checkSafety`'s guard shape.
- `fetch.go:96-150`: `fetch` and `syncStage`.

## Routing

The three TAKE items landed pre-merge (`bd34f9a`, `a66b7d1`, `dcb1f47`).
Everything else routes to 6d (the `Reconcile` wiring) or the architecture
cleanup pass.
