# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative.
Rungs 1 (foundation) and 2's passes 2a-2b are closed; pass 2c's
real-catalog stratum is complete: corpus at 162 cases, 94 of the
dossier's 96 primary rows (both misses recorded with reasons), all
twelve capture batches accepted, zero divergences owed, everything
re-deriving byte-exact. The two-code-lane execution model is
ratified, adversarially folded, and governs from here (CLAUDE.md's
re-keyed Build process section;
specs/2026-09-03-parallel-build-approach.md revision 2). The family
system is NOT live until launch. The box (5.78.148.42) runs
Navidrome + cloudflared + guardian timers only. Remote:
github.com/glw907/dubplate (private); 98 commits unpushed - push on
Geoff's word.

## Rules for a new session in this repo

1. Read CLAUDE.md's Build process (re-keyed 2026-09-03), then the
   parallel-approach spec above, then this file. The rulings doc
   (research/2026-09-02-greenfield-server-rulings.md) grounds it all.
2. The build spec (specs/2026-09-02-greenfield-server-build.md,
   rev 2.1 as amended 2026-09-03) owns the ladder; the organizer
   spec governs purview; specs go to Geoff BEFORE adversarial review.
3. The live-executor check is the lane registry below, not pgrep.
4. Box is production; quiet-box check before writes.

## Lane registry

- lane-1 (code, main tree): EXECUTING 2c tasks 16-17 (adversarial
  strata, one sequential pass-execute workflow, launched
  2026-09-03). Next after: 18 (locale + #6358), 19a/19b
  (stock-config), 20 (close + Geoff's adjudication batch); all
  sequential (corpus-ledger contention). Then 2d
  (plans/2026-09-03-rung-2d-file-level.md rev 2, ready).
- lane-2 (code, worktree): BLOCKED until Phase 0a (fifteen
  reverse-edge cells, research dispatch), 0b (rung 1b config
  wall-row registry, own small plan), 0c (theme-variable contract +
  embedded fonts) land per the folded proposal.
- doc lane (main tree): IDLE.

## Open items (carried)

- Geoff at 2c close: the adjudication batch (four hypothesis rows +
  any strata divergences). bug-threshold-boundary already disposed
  oracle-correct by Geoff, in the ledger.
- Rewrite post-mortem still owed; two albums in
  /srv/music/review/2026-08-31/geoff/ await manual review;
  FILEBROWSER_ADMIN_PASSWORD orphaned in age store; stale failed
  units on the box (cosmetic).

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |

Spend: 2a ≈1.75M/5M; 2b ≈5.8M/6M; 2c ≈5.4M/8M so far.
