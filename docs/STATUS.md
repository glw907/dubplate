# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative.
Rungs 1 and 2's passes 2a-2b are closed; **pass 2c is closed** at 215
committed corpus cases (7 smoke, 61 structural, 94 real-catalog, 17
adversarial, 36 stock-config). Six known-oracle-bug hypotheses are
seeded (one `oracle-correct`, five pending); the adjudication batch
(commit `f5bb1bd`, synced at `890bb70`) is prepared at `research/2026-09-03-
adjudication-batch-1.md`. The two-code-lane model governs (CLAUDE.md;
specs/2026-09-03-parallel-build-approach.md rev 2). The
family system is NOT live until launch. The box (5.78.148.42) runs
Navidrome + cloudflared + guardian timers only. Remote:
github.com/glw907/dubplate (private); pushed at checkpoints per the
standing grant (Geoff, 2026-09-03).

## Rules for a new session in this repo

1. Read CLAUDE.md's Build process, then the parallel-approach spec,
   then this file. The build spec (rev 2.1) owns the ladder; specs go
   to Geoff before adversarial review.
2. The live-executor check is the lane registry below, not `pgrep`.
   Box is production; quiet-box check before writes.

## Lane registry

- lane-1 (code, main tree): EXECUTING pass 2d, file-level derivation
  (`plans/2026-09-03-rung-2d-file-level.md` rev 2; Task 13's scope,
  the audio-fixture route, and the `reconcile.go`/`versionProbe` edit
  are pre-ruled, see `docs/HISTORY.md`'s 2c entry). Grants: push
  master at checkpoints; 2d pre-authorized past its 80% flag to ~130%
  of its re-derived ceiling; rung 3 executes only after Geoff's read.
- lane-2 (code, worktree): BLOCKED until Phase 0a/0b/0c land.
- doc lane (main tree): EXECUTING the rung 3 spec's adversarial
  review (specs/2026-09-03-rung-3-go-matcher.md, six rulings already
  ratified by Geoff and out of the review's scope). Also ratified
  2026-09-03: the library-organization config direction
  (research/2026-09-03-library-organization-config-direction.md,
  syntax Candidate B), standing input to the organizing-engine rung.

## Open items (carried)

- **Geoff: the pending adjudication batch**, launch-gate blocking but
  pass-close non-blocking — 19 rows across 5 unruled hypotheses
  (featuring-pattern, remix/parenthetical, assignment-rejection-gap,
  non-Latin alias mis-selection, locale alias ignored), presented at
  `docs/superpowers/research/2026-09-03-adjudication-batch-1.md`.
- Tier C exclusions, `match.required`/`match.ignored`, and
  `preferred.original_year` stay unreachable under every profile,
  carried since 2b (`docs/HISTORY.md`'s 2c entry has detail).
- Rewrite post-mortem owed; two albums in
  /srv/music/review/2026-08-31/geoff/ await review;
  FILEBROWSER_ADMIN_PASSWORD orphaned; stale failed units (cosmetic).

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |

Spend: 2a ≈1.75M/5M; 2b ≈5.8M/6M; 2c ≈8.2M against the 8M
re-derived ceiling (12M pre-re-derivation; flag overrun approved).
