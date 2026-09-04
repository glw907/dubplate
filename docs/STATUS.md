# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative.
Rungs 1 (foundation) and 2's passes 2a-2b are closed; pass 2c is
complete through Task 17: corpus at 178 cases, real-catalog stratum
done, adversarial stratum done (all four probed hypotheses updated
with evidence, every disposition pending-adjudication). Task 16
(9d87622): both string-heuristic defects reproduced on all ten
cases, verdicts moved strong-to-medium, defect-free distances on
the record. Task 17 (6e68391 + fold e805540 + c685894): six
assignment-gap cases land wrong-but-optimal mappings under
strong/apply; the alias class's four captured cases were reviewer-
caught duplicates of committed real-nonlatin cases, ruled dropped,
evidence re-pointed and claim-pinned on the twins; reproduction
narrowed to the _parse_artist_credits entity-name-over-credit
mechanism; _preferred_alias unreachability deferred to Task 18's
locale probe. The two-code-lane model governs (CLAUDE.md;
specs/2026-09-03-parallel-build-approach.md rev 2). The family
system is NOT live until launch. The box (5.78.148.42) runs
Navidrome + cloudflared + guardian timers only. Remote:
github.com/glw907/dubplate (private); commits unpushed - push on
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

- lane-1 (code, main tree): EXECUTING 2c tasks 18, 19a, 19b, 20
  (one sequential pass-execute workflow; Geoff approved finishing
  the pass past the 80% flag, hard stop well before 9M). Then 2d
  (plans/2026-09-03-rung-2d-file-level.md rev 2, ready). Overnight
  grants (Geoff, 2026-09-03 evening): push master at checkpoints;
  2d pre-authorized past its 80% flag up to ~130% of its re-derived
  ceiling, held with STATUS beyond that; rung 3 spec drafted tonight
  but rung 3 executes only after Geoff's read.
- lane-2 (code, worktree): BLOCKED until Phase 0a (fifteen
  reverse-edge cells, research dispatch), 0b (rung 1b config
  wall-row registry, own small plan), 0c (theme-variable contract +
  embedded fonts) land per the folded proposal.
- doc lane (main tree): IDLE.

## Open items (carried)

- Geoff at 2c close: the adjudication batch (tasks 16-17 added 18
  pending rows: 10 string-heuristic, 6 assignment, 2 alias
  re-pointed at real-nonlatin twins; plus any earlier strata rows).
  bug-threshold-boundary already disposed oracle-correct, in the
  ledger.
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

Spend: 2a ≈1.75M/5M; 2b ≈5.8M/6M; 2c ≈6.5M/8M through Task 17
(tasks 16-17 chain 694k, escalation folds 276k + 124k).
