# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative. Rung
3 (3a-3d) is closed at the rung 3d figures. **Rung 4a is closed**:
passes 4a-i and 4a-ii are both in the tree and green.
`internal/engine/shadow` carries the whole shadow runner,
`cmd/dubplate-server` carries the `shadow` verb's three subcommands,
and rung 4's exit gate runs one end-to-end `observe` pass over a
fourteen-directory inbox. Rung 4's six owner rulings are ratified
(Geoff, 2026-09-07, all as recommended). Read HISTORY's 4a-ii entry
before touching `internal/engine/shadow`.

Streamer 2.0 is a ratified design (ROADMAP `streamer-2-0`); its
forward hooks bind rungs 5 through 9 and 12 via CLAUDE.md.

The family system is not live; the box (5.78.148.42) runs Navidrome +
cloudflared + guardian timers only, and it is production. A new session
reads CLAUDE.md's Build process, then the parallel-approach spec (rev 2),
then this file; the live-executor check is the lane registry.

## Next action

Pass 4b, the rung 4 deploy, in the main tree. Its plan is owed
just-in-time and does not exist yet. 4b cannot close without an owner
sitting inside it (ruling 6), and its close records `started-at`,
because the launch date is that date plus the 720h floor. HISTORY's
4a to 4b handoff is what the plan builds from.

## Lane registry

- lane-1 (main tree): IDLE, pending 4b's plan.
- lane-2 (worktree): BLOCKED until Phase 0b/0c.
- doc lane: IDLE, one candidate queued (the 2.0 study rung). Whether to
  run it before 4b is the conductor's call.

## Open items (carried)

- Thirty declared coverage gaps, rung 3's twenty-one plus rung 4a's
  nine, owned in HISTORY, none blocking. Read the total, never
  re-count it.
- Four-knob wall registration follows rung 1b, now joined by rung 4's
  `KindFloat` need; `AdminEmail` and five 3b leftovers, none blocking.
  `profile_hashes` (m1) and the response-status extension open since
  2b/2c; `Case.Extra` and `MergePseudoRelease`'s `slices.Clone` to rung 9.
- 4a-i code untouched by 4a-ii still carries "provisional, owner
  ruling N open" marks; all six are ratified, retiring them is mechanical.
- 17 unledgered tied candidates filed for lane E1; owner ask to
  re-disposition `bug-assignment-rejection-gap` to `no-correct-answer`.
- Adjudication batch 1 fully ruled. Rewrite post-mortem owed; two albums
  await review in `/srv/music/review/2026-08-31/geoff/`;
  FILEBROWSER_ADMIN_PASSWORD orphaned; stale failed units.

## Pinned versions (looked up 2026-08-30)

| Component | Image / package | Tag / version |
|---|---|---|
| Navidrome | `deluan/navidrome` | `0.63.2` |
| cloudflared | `cloudflare/cloudflared` | `2026.8.2` |
| beets (box, via uv) | PyPI `beets` | `2.13.1` |
| mediafile (corpus manifest, 2d) | PyPI `mediafile` | `0.17.0` |
| unidecode (matcher table, 3a) | PyPI `unidecode` | `1.4.0` |
| lap (matcher tie mirroring, 3d) | PyPI `lap` | `0.5.13` |

Spend to date: about 42.8M through 4a-ii (HISTORY has per-pass figures).
