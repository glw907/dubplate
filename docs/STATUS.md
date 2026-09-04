# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative.
**Rung 2 is closed**: 215 committed corpus cases (7 smoke, 61
structural, 94 real-catalog, 17 adversarial, 36 stock-config), the
file-level applied-tag ledger (`corpus/analysis/file_tags.json`, 12
subjects), and the two standing chore docs
(`docs/corpus-version-bump-differential.md`,
`docs/corpus-recapture-chore.md`). The rung 8 handoff lives in
`corpus/README.md`'s "Handoff to rung 8" section. Six known-oracle-bug
hypotheses are seeded (one `oracle-correct`, five pending). The
two-code-lane model governs (CLAUDE.md;
specs/2026-09-03-parallel-build-approach.md rev 2). The family system
is NOT live until launch. The box (5.78.148.42) runs Navidrome +
cloudflared + guardian timers only. Remote: github.com/glw907/dubplate
(private); pushed at checkpoints (Geoff, 2026-09-03).

## Rules for a new session

Read CLAUDE.md's Build process, then the parallel-approach spec, then
this file. The live-executor check is the lane registry below, not
`pgrep`. Box is production; quiet-box check before writes.

## Lane registry

- lane-1 (code, main tree): NEXT ACTION is the rung 3 pass-3a planning
  gate — rung 3 (the Go matcher) executes only after Geoff reads its
  revision-2 spec (`docs/superpowers/specs/2026-09-03-rung-3-go-matcher.md`).
- lane-2 (code, worktree): BLOCKED until Phase 0a/0b/0c land.
- doc lane (main tree): IDLE. Phase 0a is complete
  (research/2026-09-04-reverse-edge-completion.md: fourteen SOFT,
  one HARD — lane V waits on rung 9's concurrency ruling; permit/
  forbid section governs lane launches). The rung 3 spec is folded
  at revision 2 (review at research/2026-09-04-rung-3-spec-review.md).
  Lane-2 now waits only on Phase 0b (rung 1b) and 0c.

## Open items (carried)

- **Geoff: the pending adjudication batch** (launch-gate blocking, not
  pass-close blocking — 19 rows across 5 unruled hypotheses:
  featuring-pattern, remix/parenthetical, assignment-rejection-gap,
  non-Latin alias mis-selection, locale alias ignored;
  `docs/superpowers/research/2026-09-03-adjudication-batch-1.md`) and
  **the rung 3 spec read** before lane-1 executes rung 3.
- `profile_hashes` (m1), the response-status extension, and Tier C
  exclusions (`match.required`/`match.ignored`,
  `preferred.original_year`) stay open, carried since 2b/2c
  (`docs/HISTORY.md`'s 2c and 2d entries have detail).
- Rewrite post-mortem owed; two albums in /srv/music/review/2026-08-31/geoff/
  await review; FILEBROWSER_ADMIN_PASSWORD orphaned; stale failed units (cosmetic).

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |

Spend: 2a ≈1.75M/5M; 2b ≈5.8M/6M; 2c ≈8.2M/8M; 2d ≈1.6M/2.5M.
