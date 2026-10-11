# dubplate status

## Current state
Master's head is the alignment merge (`551bf4f`, PR #2); the last recorded green full gate is at `f0a2056`. Nothing is
paused and nothing deploys. The landed rungs, the run order, and the next pass are in ROADMAP, "Path to 0.8". Work a
later pass carries sits in BACKLOG under the `#rung:<name>` tag of the rung that takes it.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
The `admin-api-1` plan is ready to launch on branch `worktree-admin-api-1-plan` once Geoff answers one question. The
plan is `docs/superpowers/plans/2026-10-11-admin-api-1.md`, with its launch prompt beside it and its review, fold, and
verification records at `docs/superpowers/research/2026-10-11-admin-api-1-plan-*.md`. The planning session
(`admin-api-1-plan`, id `6dc7ffa2-8f8d-497d-a22f-465faceb2125`) has ended.

**Question for Geoff (pass-core's clock line).** The plan's honest estimate is 695 task-minutes (585 on the path).
That is over 660, twice the design's 330, the line past which a plan comes back to Geoff before launch. The plan
review's proofs and recorded gate times drive the rise, and the brief itself expected "toward 7 hours" with
`auth-data` fix rounds. Approve 695 as planned? **Recommendation: yes.** The per-task clock stop still caps each task
at twice its estimate. The alternatives are folding T11 into T10 (about 675 to 680, design whole), or that plus moving
the route inventory to `admin-api-2` (about 655 to 660, dropping work PS:580-581 gives this pass). Detail: the fold
record, "Rulings for Geoff".

On a yes, Geoff or a conducting session that is not worktree-isolated runs the plan's "Launch" block from the main
checkout: its checks first, then `claude --bg --model sonnet --effort medium --name admin-api-1` with the prompt read
by `git show worktree-admin-api-1-plan:docs/superpowers/plans/2026-10-11-admin-api-1-prompt.txt`. The launched
session replaces this section with its own in-flight line.

## Open decisions
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
