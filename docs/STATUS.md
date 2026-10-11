# dubplate status

## Current state
Master's head is the lean-cutover merge (`8269d27`); its last recorded green full gate is at `f0a2056`. Nothing is
paused and nothing deploys. The landed rungs, the run order, and the next pass are in ROADMAP, "Path to 0.8". Work a
later pass carries sits in BACKLOG under the `#rung:<name>` tag of the rung that takes it.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

The alignment pass merged as PR #2 (`551bf4f`).

## Next action
The `admin-api-1` re-plan is in flight: session `admin-api-1-plan` (id `6dc7ffa2-8f8d-497d-a22f-465faceb2125`),
branch `worktree-admin-api-1-plan`, writing `docs/superpowers/plans/2026-10-11-admin-api-1.md` and its launch prompt.

## Open decisions
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
