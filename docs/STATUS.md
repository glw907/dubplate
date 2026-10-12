# dubplate status

## Current state
Master's head is the `admin-api-1` merge (`8477287`, PR #4); its close gate (`gate.sh --merge`) was green on
`e6440e9`. Nothing deploys. The landed rungs, the run order, and the next pass are in ROADMAP, "Path to 0.8". Work a
later pass carries sits in BACKLOG under the `#rung:<name>` tag of the rung that takes it. PR #4's "Found, not done"
list is triaged: item 1 rides in `intake-1`, items 2, 3, and 6 to 11 are BACKLOG #75 to #82, and items 4 and 5 are
dropped.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
In flight: `intake-1-plan` (session `ffa5bf00-81bc-4fe4-bbdc-4941e3e13624`) is planning `intake-1` on branch
`worktree-intake-1-plan`.

## Open decisions
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.5 are majors held for Geoff (the 2026-10-11 dependency sweep).
