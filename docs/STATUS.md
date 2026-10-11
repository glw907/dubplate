# dubplate status

## Current state
Master's head is the alignment merge (`551bf4f`, PR #2); the last recorded green full gate is at `f0a2056`. Nothing is
paused and nothing deploys. The landed rungs, the run order, and the next pass are in ROADMAP, "Path to 0.8". Work a
later pass carries sits in BACKLOG under the `#rung:<name>` tag of the rung that takes it.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
In flight: the `admin-api-1` pass, session `admin-api-1`, id `6baa2add-e9c2-4ac4-92a3-3e5054aef3d8`, branch
`worktree-admin-api-1`. The plan is `docs/superpowers/plans/2026-10-11-admin-api-1.md`; Geoff approved its 695-minute
estimate before launch.

## Open decisions
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
