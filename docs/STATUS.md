# dubplate status

## Current state
The reshape is done. `dubplate` builds by the workstation's `building-from-specs` process. Task 1 (PR #20) fixed the
shutdown race in `internal/serve`, task 2 (PR #21) rewrote the Claude files and the backlog, and task 3 (PR #22) added
`govulncheck` to the gate, three fuzz targets, the race detector in CI, and a `dependabot.yml` that watches the GitHub
Actions pins. GitHub's secret scanning, push protection, and Dependabot alerts are on. The spec is
`~/Projects/initiatives/cairn-dubplate-reshape/specs/2026-11-22-dubplate-reshape.md`. Nothing a member or an operator
sees changed.

Milestone `engine` (PR #17) is done: the matcher, MusicBrainz client, tag reader, and `tag` package each state a fact
once, and the citation leg of `scripts/check.sh` gates.

Work `structure-base` (PR #18, merged) moved code only: `_salvage` left the tree, `internal/engine/store` became
`internal/store`, `internal/id` parses ids, and a store test names the owner of each of 19 tables.
Work `structure-keys` (PR #19, merged) moved every setting into the package that reads it, each with a `config.go`, a
`Register` function, and readers. `cmd/dubplate-server` declares and registers nothing, and an import test holds
configuration to `config.go` and `doctor.go` in seven packages. Output is unchanged.

`intake-1` is the last product work on master. A passing album waits in `matching` with no exit until
`intake-2`'s match probe lands. The run order is in BACKLOG, "Milestones to 0.8". Work still to do sits in BACKLOG
under its `#milestone:<name>` tag.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
The overnight spec is `~/Projects/initiatives/dubplate-server/specs/2026-11-22-overnight-design.md` (approved
2026-11-22). `dev-friction` merged as PR #24 (`0632ecf`): the driver is `modernc.org/sqlite` v1.63.0, certifi is
2026.11.20, `TestBusyTimeout` waits on signals, lint runs queue, and `scripts/check.sh` unsets every `DUBPLATE_*`
name. `structure-doctor` still runs, moving the eleven health checks out of `cmd/dubplate-server` (#123, #124, #143,
#144, #147, #156, #172, #173). Its pull request takes `gh pr update-branch` and one green CI run before it merges. The session Geoff talks to applies the spec's backlog
text after each merge. `intake-2` takes #159 next.
Before any deploy that carries `cleanup-store`, check read-only that the box store holds no restore-intent record (#148).

## Open decisions
- The repo is public on GitHub and stays so (Geoff, 2026-11-10). Open under it: BACKLOG #96's question of whether
  the Unidecode-derived table in git history needs a lawyer's reading or a history rewrite.
- `intake-2`'s first task is the three rulings of #92.
- R-6d21 overrides the settings-home rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- The gate unsets the workstation's 16 `DUBPLATE_*` exports, but a bare `go test` still passes them to child
  verbs (BACKLOG #176). Pruning `~/.local/secrets` stays a workstation chore.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.x are majors held for Geoff (BACKLOG #130).
- Geoff accepted on 2026-11-22 that a write under lock contention ends at its deadline on the new driver (#129).
