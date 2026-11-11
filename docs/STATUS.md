# dubplate status

## Current state
`intake-1` is the last product pass on master. Nothing deploys: a passing album waits in `matching` with no exit
until `intake-2`'s match probe lands. The landed rungs and the run order are in BACKLOG, "Path to 0.8". Work a later
pass carries sits in BACKLOG under its `#rung:<name>` or `#pass:<name>` tag.

A whole-repo review on 2026-11-10 (`docs/research/2026-11-10-whole-repo-review.md`) set a series of passes ahead of
`intake-2`, approved by Geoff the same day: `gate-1`; then `cleanup-engine`, `cleanup-store`, and `web-shell-1` at
once; then `cleanup-ops`; then `intake-2`. BACKLOG #104 to #127 hold the findings by pass.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
`gate-1`, in a new session and worktree. Launch prompt:

> Run the `gate-1` pass in dubplate through the `dubplate-pass` skill. It is runner class. Its scope is BACKLOG's
> items tagged `#pass:gate-1` (#34, #97, #104 to #108) and the "gate-1" section of
> `docs/research/2026-11-10-whole-repo-review.md`. Start with finding G1, per-leg timing, and order the other cuts
> by what it measures. Before the design, read "Coordination" in `docs/STATUS.md` and the status line of
> `~/Projects/initiatives/workstation-refinement/`. Bring the design for one approval.

## Coordination
The workstation's Claude infrastructure is under its own plan: the `workstation-refinement` initiative
(`~/Projects/initiatives/workstation-refinement/`, plan under `plans/`). Until Geoff approves it, the record sits
only on the local branch `workstation-refinement`, in the worktree
`~/Projects/initiatives/.claude/worktrees/workstation-refinement/`; read the copy on main when it exists there. Its
status line governs: `proposed` means nothing in it is settled, and `active` means Geoff approved it. In the plan,
G16 and G17 sit in "Pass 3: process", and the `cairn-run-gate` changes in "Pass 1: fixes and settings". dubplate edits nothing in `~/.dotfiles`, the standards
plugin, the global `CLAUDE.md`, or `pass-core`; a finding that needs one goes to that initiative. The seam, as agreed
with its planning session on 2026-11-10:

- `gate-1` goes ahead on G1 to G15 and G18. G14 and G15 change the `dubplate-pass` adapter only. Under G15, each
  dispatch keeps one sentence telling the subagent to read the cited range first.
- G16 (STATUS carries no fact git knows) holds until that plan is `active`, because it proposes one home for the
  one-executor check that reads STATUS's session ids. A pass that reaches G16 first leaves it in BACKLOG #108.
- G17 (Close removes the worktree and branch) belongs to that plan as a `pass-core` step. dubplate adds nothing.
- BACKLOG #109 (findings X1 to X6) is proposed in that plan.
- `cairn-run-gate` keeps its lanes and flags. That plan proposes additions only: a nesting marker, a preflight that
  names `npm ci --prefix web`, and shorter default output. A pass that depends on one checks that it has landed.

Every pass in the series reads that status line at its design, and reports a workstation-layer finding to the
initiative instead of fixing it here.

## Open decisions
- The re-plan sitting (#127) is due before `pipeline` is planned: rung 10's migrator against a folder-admit verb,
  streamer slices 3 and 5 first with a private deploy, and three scope candidates for after 0.8.
- The repo is public on GitHub and stays so (Geoff, 2026-11-10). Open under it: BACKLOG #96's question of whether
  the Unidecode-derived table in git history needs a lawyer's reading or a history rewrite.
- `intake-2`'s first task is the three rulings of #92.
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.5 are majors held for Geoff (the 2026-10-11 dependency sweep).
