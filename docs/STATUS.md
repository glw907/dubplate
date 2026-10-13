# dubplate status

## Current state
Master's head is the `admin-api-1` merge (`8477287`, PR #4) until the `intake-1` PR merges. Nothing deploys. The
landed rungs, the run order, and the next pass are in ROADMAP, "Path to 0.8". Work a later pass carries sits in BACKLOG
under the `#rung:<name>` tag of the rung that takes it. PR #4's "Found, not done" list is triaged: items 2, 3, and 6 to
11 are BACKLOG #75 to #82, and items 4 and 5 are dropped; item 1 landed in `intake-1`.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
The `intake-1` PR is open for Geoff's read (session `intake-1`, id `930921a1-a634-461c-91be-416fd394c230`, branch
`worktree-intake-1`). It builds the intake core: a member's tus upload, the gate, the verdicts, and the `/api/v1`
submission operations. `intake-1` must not deploy before `intake-2`: a passing album waits in `matching` with no exit
until `intake-2`'s match probe lands. The next pass is `intake-2`; its planner reads BACKLOG #83 to #85.

## Open decisions
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.5 are majors held for Geoff (the 2026-10-11 dependency sweep).
