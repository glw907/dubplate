# dubplate status

## Current state
Master's head is the lean-cutover merge (`8269d27`); its last recorded green full gate is at `f0a2056`. Nothing is
paused and nothing deploys. The landed rungs, the run order, and the next pass are in ROADMAP, "Path to 0.8". Work a
later pass carries sits in BACKLOG under the `#rung:<name>` tag of the rung that takes it.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

The alignment pass is in flight on branch `worktree-dubplate-alignment`. This paragraph and the in-flight line below
appear on that branch only.

## Next action
The alignment PR is open for Geoff's read; the next action starts when it merges. Session `dubplate-alignment` (id
`394cf8c0-5100-4330-bf09-64b2412697bc`), branch `worktree-dubplate-alignment`.

Once this pass's PR merges, whoever merged it (the conducting session or Geoff) runs these four steps in order:

1. Run `git -C ~/Projects/dubplate pull --ff-only`, under pass-core's conditions: the checkout is on `master`, holds
   no uncommitted change in a file the merge touches, and has no live session per `claude agents --json`. If any
   condition fails, report the checkout stale and stop.
2. Prune the auto-memory. Back it up first with `tar czf ~/.local/share/dubplate/memory-backup-2026-10-11.tar.gz -C
   ~/.claude/projects/-var-home-glw907-Projects-dubplate memory`. Re-list the directory with `ls -l
   --time-style=+%s` and keep any file modified after T4's start epoch, which the PR body names. Write the new
   `MEMORY.md` from the PR body's proposed index. Show the delete list from the PR body's dry-run table, then delete
   those files. Re-list the directory and run the retired-phrase check over the new `MEMORY.md`.
3. Trash the three Pi 5 originals: `gio trash` on each file in `~/Projects/dubplate/reports/` that T0 copied into
   `docs/superpowers/research/2026-10-10-pi5-client/`.
4. Launch the `admin-api-1` re-plan, an unattended Opus planning session that works from the `admin-api` spec
   (`docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md`, revision 2.1, approved by Geoff on 2026-10-10;
   its "Rulings for Geoff" section reads "None"):

   ```bash
   cd ~/Projects/dubplate && claude --bg --model opus --effort high --name admin-api-1-plan "$(cat docs/superpowers/plans/2026-10-11-admin-api-1-planning-prompt.txt)"
   ```

   Its first STATUS edit replaces this list with its own in-flight line.

## Open decisions
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
