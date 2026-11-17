# dubplate status

<!-- round-report cleanup -->
**Round cleanup is done.** PRs #13 to #16 merged on 2026-11-17 on Geoff's approve of `20261117-064744-b8dd9b`, and the
approved allow rules landed in `.claude/settings.json` (`7151edb`). Nothing has deployed; the box is unchanged. Report:
`~/Projects/initiatives/dubplate-server/reports/cleanup-report.md`.
<!-- /round-report -->

## Current state
Round `engine` (PR #17) is done: the matcher, MusicBrainz client, tag reader, and `tag` package each state a fact
once, the limiter is gone, the client caps what it reads at 32 MiB, the harness grades the pattern table it
names, and the citation leg of `scripts/check.sh` gates. No distance, recommendation, action, parsed tag, or
written byte changed, and `make derive` passes. One failure path is stricter: a write refuses a source that
shrank between the parse and the copy, where it once renamed the short file into place. Nothing deploys.

`intake-1` is the last product pass on master. A passing album waits in `matching` with no exit until
`intake-2`'s match probe lands. The run order is in BACKLOG, "Rounds to 0.8". Work a later
pass carries sits in BACKLOG under its `#round:<name>` or `#pass:<name>` tag.

The cleanup rounds come from the 2026-11-10 whole-repo review (`docs/research/2026-11-10-whole-repo-review.md`),
regrouped by `~/Projects/initiatives/dubplate-server/specs/2026-11-15-cleanup-round-design.md`.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
The next round is the one BACKLOG's "Rounds to 0.8" names after `engine`; `intake-2` takes the pacing and
transliteration decisions filed as #159.
Before any deploy that carries `cleanup-store`, check read-only that the box store holds no restore-intent record (#148).

## Coordination
The workstation's Claude infrastructure is under the `workstation-refinement` initiative
(`~/Projects/initiatives/workstation-refinement/`, plan under `plans/`). Until Geoff approves it, the record sits
only on the local branch `workstation-refinement`; read the copy on main when it exists there. Its status line
governs: `proposed` settles nothing, and `active` means Geoff approved it. G16 and G17 sit in its "Pass 3: process",
and the `cairn-run-gate` changes in "Pass 1: fixes and settings". dubplate edits nothing in `~/.dotfiles`, the
standards plugin, the global `CLAUDE.md`, or `pass-core`; a finding that needs one goes to that initiative. The
seam, as agreed with its planning session on 2026-11-10:

- `gate-1` goes ahead on G1 to G15 and G18. G14 and G15 change the `dubplate-pass` adapter only.
- G16 (STATUS carries no fact git knows) holds until that plan is `active`, because it proposes one home for the
  one-executor check that reads STATUS's session ids. A pass that reaches G16 first leaves it in BACKLOG #108.
- G17 (Close removes the worktree and branch) belongs to that plan as a `pass-core` step. dubplate adds nothing.
- BACKLOG #109 (findings X1 to X6) is proposed in that plan.
- `cairn-run-gate` keeps its lanes and flags. That plan proposes additions only: a nesting marker, a preflight that
  names `npm ci --prefix web`, and shorter default output. A pass that depends on one checks that it has landed.

Each pass reads that status line at its design.

## Open decisions
- The repo is public on GitHub and stays so (Geoff, 2026-11-10). Open under it: BACKLOG #96's question of whether
  the Unidecode-derived table in git history needs a lawyer's reading or a history rewrite.
- `intake-2`'s first task is the three rulings of #92.
- R-6d21 overrides the settings-home pass's rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.x are majors held for Geoff (BACKLOG #131). The `modernc.org/sqlite`
  1.62.0 minor stays out because it changes the busy timeout and fails two `store` tests (#129).
