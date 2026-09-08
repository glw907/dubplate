# dubplate status

## Current state

The greenfield dubplate-server build is the one active initiative.
**Rung 3 (3a-3d) is closed**; `internal/engine/{matcher,musicbrainz,tag}`
and `internal/harness/conform` are green at the rung 3d figures (see
HISTORY). **Rung 4's spec is in place and its six owner rulings are ratified**
(Geoff, 2026-09-07, all as recommended: integration-class wall, release
under dispute, both compat modes, 720h floor, 60-request cap, deploy as
an owner sitting).

**Pass 4a-i (Tasks 0-9, rev 2.1) is closed at its intermediate close.**
`internal/atomicfile`, `matcher.QuietImportTunables`, the shadow's
config wall, the FLAC reader, the store, capture, reconcile, the
transports, the re-score and classification are in the tree and
green. Read HISTORY's "Rung 4, pass 4a-i" entry before touching
`internal/engine/shadow`. The family system is not live; box
(5.78.148.42) runs Navidrome + cloudflared + guardian timers only. A
new session reads CLAUDE.md's Build process, then the
parallel-approach spec (rev 2), then this file; live-executor check is
the lane registry, not `pgrep`, and the box is production.

## Lane registry

- lane-1 (main tree): **EXECUTING 4a-ii post-chain fix rounds**. The
  chain (wf_c97fd506-dda, 16 agents, 1.32M) accepted Tasks 10-12
  (09353f0..3c229f0); Task 13 (f7cd7eb, f7f82b7) drew a second `fix`
  and Task 14 (9c956b8) escalated (mutations 4 and 6 unfired, proofs
  written as commentary). Conductor-ruled fixes run sequentially, 13
  then 14 (opus), then code-simplifier and the Task 15 close. Then 4b
  (deploy, owner sitting, ruling 6). 4a-ii ceiling 3M (Geoff, 2026-09-07), flag at
  2.4M to Geoff; chain args in the project handoff dir
  (`4a-ii-chain-args.json`).
  lane-2 (worktree): BLOCKED until Phase 0b/0c. Doc lane: IDLE.

## Open items (carried)

- Rung 3's **21 declared corpus-coverage gaps** (3a three, 3b six, 3c
  one, 3d eleven), owned in HISTORY, none blocking. Rung 4's own gap:
  the config-import rule (C3) is guarded by the in-chain reviewer, not
  an architectural test.
- Four-knob wall registration follows rung 1b; `AdminEmail` and five
  3b leftovers, none blocking. `profile_hashes` (m1) and the
  response-status extension open since 2b/2c; `Case.Extra` and
  `MergePseudoRelease`'s `slices.Clone` re-owned to rung 9.
- 17 unledgered tied candidates filed for lane E1; owner ask to
  re-disposition `bug-assignment-rejection-gap` to `no-correct-answer`.
- Adjudication batch 1 fully ruled. Rewrite post-mortem owed; two
  albums await review in `/srv/music/review/2026-08-31/geoff/`;
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

Spend to date: about 40.1M across rungs 2, 3, 4a-i (HISTORY has
per-pass figures).
