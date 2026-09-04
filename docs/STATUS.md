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

- lane-1 (code, main tree): IDLE after 3a's close. Next: 3b or 3c
  (mutually independent, may swap order); 3d needs all three.
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

Spend: 2a ≈1.75M/5M; 2b ≈5.8M/6M; 2c ≈8.2M/8M; 2d ≈1.6M/2.5M; 3a ≈2.26M/6M.
