# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative.
**Rung 2 is closed** (215 corpus cases, file-level applied-tag
ledger, two standing chore docs; detail in `docs/HISTORY.md`'s 2c/2d
entries). **Rung 3, pass 3a is closed**: `internal/engine/matcher`
(21 distance components, `Distance`, `StringDist`, `Tunables`) and
`internal/harness/conform` (tier 0's typed `components.json` decoder,
the M1 defaults cross-check, the architecture checks) both exist and
are green, tier 0 replaying 72/72 committed rows bit-identical. The
coverage table, evidence record, and the ten rediscover-guards live
in `docs/HISTORY.md`'s 3a entry. Read it before starting 3b or 3c.
The two-code-lane model governs (CLAUDE.md;
specs/2026-09-03-parallel-build-approach.md rev 2). The family system
is NOT live until launch. The box (5.78.148.42) runs Navidrome +
cloudflared + guardian timers only. Remote: github.com/glw907/dubplate
(private); pushed at checkpoints.

## Rules for a new session

Read CLAUDE.md's Build process, then the parallel-approach spec, then
this file. The live-executor check is the lane registry below, not `pgrep`.
Box is production; quiet-box check before writes.

## Lane registry

- lane-1 (code, main tree): EXECUTING pass 3b (translation layer) per
  `docs/superpowers/plans/2026-09-04-rung-3b-translation.md` (revision 2,
  review folded, four conductor rulings taken 2026-09-04). Chain: Task 0
  pre-task, then Tasks 1-8 via pass-execute (sequential, this tree), then
  code-simplifier, then Task 9's close. Ceiling 6M, flag 4.8M, re-derive
  after Task 3. Run `wf_6521a55d-5f9` (Tasks 1-8). Checkpoint at Task 4
  (17:49): Tasks 1-4 accepted (946cc21, a67d5a0+ae2b56e, 7fcb5b4+7092090,
  b8460e0; Tasks 2 and 3 each took one fix round); Task 5 in flight.
  Spend through Task 4 ≈1.8M by transcript volume, ≈450K per task with
  fix rounds; projected close ≈4.5-5M, so the 6M ceiling stands and the
  Task 2/3 split is not taken. If interrupted: read the pass-execute journal, verify the
  last committed task's review, relaunch fresh from the next task.
  Next after 3b: 3c; 3d needs all three.
- lane-2 (code, worktree): BLOCKED until Phase 0b (rung 1b) and 0c land.
- doc lane (main tree): IDLE. Phase 0a is complete.

## Open items (carried)

- `profile_hashes` (m1) and the response-status extension stay open,
  carried since 2b/2c (`docs/HISTORY.md`'s 2c and 2d entries).
- The three M8-declared corpus-coverage gaps (`match.required`,
  `match.ignored`, `preferred.original_year`) are recorded with an owner
  in 3a's HISTORY entry. Neither blocks 3b.
- The four-knob wall registration is a named follow-on task for the
  first rung 3 pass after rung 1b lands (knobs and rationale in 3a's
  HISTORY entry); `internal/config/keys.go` stays untouched until then.
- Adjudication batch 1 is fully ruled: 18 oracle-wrong, 3 oracle-correct in
  `corpus/analysis/adjudications.json`, zero pending; the launch gate's item is clear.
- Rewrite post-mortem owed; two albums in /srv/music/review/2026-08-31/geoff/ await
  review; FILEBROWSER_ADMIN_PASSWORD orphaned; stale failed units.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |

Spend: 2a ≈1.75M/5M; 2b ≈5.8M/6M; 2c ≈8.2M/8M; 2d ≈1.6M/2.5M; 3a ≈2.84M/6M.
