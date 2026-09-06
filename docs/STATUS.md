# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. **Rung 3
(3a, 3b, 3c, 3d) is CLOSED.** `internal/engine/{matcher,musicbrainz,tag}`
(scoring, verdicts, translation, client, limiter, and `tag.Album`, the
matcher's public entry point) are green. `internal/harness/conform` grades
tier 0 at 72/72, translation at 50/50, requests at 215/303/48, tier 1 at
215/294/294/214, and tier 2 at 215 verdicts (`ModeCompat`) plus ten
`defect_free` rows bit-identical (`ModeDefault`), all zero-miss or
bit-identical. Read `docs/HISTORY.md`'s 3a-3d entries before rung 4. Family
system NOT live; box (5.78.148.42) runs Navidrome + cloudflared + guardian
timers only.

## Rules for a new session

Read CLAUDE.md's Build process, then the parallel-approach spec (rev 2),
then this file. Live-executor check is the lane registry, not `pgrep`;
the box is production.

## Lane registry

- lane-1 (main tree): **3d closed; next action: rung 4 spec and plan prep
  (the shadow) under the 2026-09-05 grant**, ruling B option (b): rung 4
  deploys at functional completeness with every divergence stamped with
  the matcher commit. Flag routes to the conductor through 2026-09-07
  09:00; window cap 15M, about 5.15M spent (3d's full-pass total).
  lane-2 (worktree): BLOCKED until Phase 0b/0c land. Doc lane: IDLE.

## Open items (carried)

- Rung 3 carries **twenty-one declared corpus-coverage gaps** (3a three,
  3b six, 3c one, 3d eleven), owned in HISTORY; none blocks rung 4.
- Four-knob wall registration (`StrongRecThresh`, `MediumRecThresh`,
  `RecGapThresh`, `SearchLimit`) is the follow-on for rung 3's first pass
  after rung 1b; `AdminEmail` contact and five 3b leftovers, none blocking.
- `profile_hashes` (m1) and the response-status extension stay open since
  2b/2c; `Case.Extra` producer and `MergePseudoRelease`'s `slices.Clone`
  fix are re-owned to rung 9.
- 17 unledgered tied candidate assignments (`ties.json` measures
  `Candidates[0]` only) filed for lane E1; owner ask to re-disposition
  `bug-assignment-rejection-gap` to `no-correct-answer` (R1).
- Adjudication batch 1 fully ruled (18 oracle-wrong, 3 oracle-correct,
  zero pending). Rewrite post-mortem owed; two albums await review in
  `/srv/music/review/2026-08-31/geoff/`; FILEBROWSER_ADMIN_PASSWORD
  orphaned; stale failed units.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 35.1M across rungs 2 and 3 (per-pass figures in HISTORY).
