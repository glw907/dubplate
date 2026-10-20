# `internal/config` architecture read (5c-iii merge side)

Read at HEAD `db1e111`, no plan in context, per CLAUDE.md's "Reviewers" (one
`go-architecture-reader` per touched Go package at each pass's merge).

**Verdict: workmanlike.** Density 0.74 against `flag`'s 0.38.

## Findings

- **Exported surface with no outside caller.** `OwnedWall`/`ReservedClasses` package functions
  (`keys.go:27-30,108-111`); `AddSuperseded`/`Superseded` (`registry.go:458,469`);
  `AddReservedClass` (`registry.go:165`); `IntegrationClass.Rung` is set but never read
  (`keys.go:186`).
- **File split by deliverable rather than by concern.** Six free lookup functions take
  `*Registry` as their first argument; they read as methods declared in the wrong place.
- **Duplication.** `declares()` restated on both `IntegrationClass` and `IdentitySource`; a
  `checkKeys` helper duplicated across the `Add*` loops; owned key names live as constants in
  both `config.go:390-393` and `keys.go:21-24`; `ParseString` and `validateThresholdDefault`
  each restate a parser rather than sharing one; `config.go:258-262` repeats `Enabled`'s doc
  comment.
- **Comments.** The package doc (`config.go:6-46`) is stale: it calls backup an integration
  class, which it no longer is. `config.go:48-53`'s numbering is off by one (labeled fifth,
  actually sixth paragraph). Sixteen process-citation trailers. Stale prose at
  `registry.go:264-266`, `keys.go:181-183`, `thresholds.go:10-11`.
- **Tests.** `keys_test.go:144` is misnamed for what it asserts; `thresholds_test.go:166`
  carries a process citation.

## Best code in the package

- `config.go:183-201` — the `get`/`require` dedup.
- `config.go:231-250` — `requireSiteName`'s check ordering, with no value echoed on refusal.
- `registry.go:24-31` — the panic rationale.

## Routing

The architecture cleanup pass takes the exported-surface, duplication, and comment findings.
One note goes to rung 6's 6c: 6c edits this registry, and its Task 0/2 must not build on the
stale package doc; the lookup-method reshape (free functions to methods) is adjacent to 6c's
own edits and worth doing in the same pass if 6c's diff already touches the neighborhood.
