# dubplate status

## Current state
The reshape runs now. Three tasks move the repo onto the workstation's `building-from-specs` process: task 1 fixes the
shutdown race in `internal/serve`, task 2 rewrites the Claude files and the backlog, and task 3 adds `govulncheck`, three
fuzz targets, the race detector in CI, and a `dependabot.yml`. The spec is
`~/Projects/initiatives/cairn-dubplate-reshape/specs/2026-11-22-dubplate-reshape.md`. The `structure` work waits for
the three tasks. Nothing a member or an operator sees changes.

Round `engine` (PR #17) is done: the matcher, MusicBrainz client, tag reader, and `tag` package each state a fact
once, and the citation leg of `scripts/check.sh` gates.

Work `structure-base` (PR #18, merged) moved code only: `_salvage` left the tree, `internal/engine/store` became
`internal/store`, `internal/id` parses ids, and a store test names the owner of each of 19 tables.
Work `structure-keys` (PR #19, merged) moved every setting into the package that reads it, each with a `config.go`, a
`Register` function, and readers. `cmd/dubplate-server` declares and registers nothing, and an import test holds
configuration to `config.go` and `doctor.go` in seven packages. Output is unchanged.

`intake-1` is the last product work on master. A passing album waits in `matching` with no exit until
`intake-2`'s match probe lands. The run order is in BACKLOG, "Rounds to 0.8". Work still to do sits in BACKLOG
under its `#round:<name>` or `#pass:<name>` tag.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
Task 1, the shutdown fix, and task 2, the Claude files and the backlog, build first. Task 3 follows once both have
merged. Then `structure-doctor`, the third piece of `structure`, restarts from a spec; `intake-2` takes #159.
Before any deploy that carries `cleanup-store`, check read-only that the box store holds no restore-intent record (#148).

## Open decisions
- The repo is public on GitHub and stays so (Geoff, 2026-11-10). Open under it: BACKLOG #96's question of whether
  the Unidecode-derived table in git history needs a lawyer's reading or a history rewrite.
- `intake-2`'s first task is the three rulings of #92.
- R-6d21 overrides the settings-home rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- About 41 stale `DUBPLATE_*` exports in `~/.local/secrets` make `check` fail outside `env -i`. They need pruning on
  the workstation.
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.x are majors held for Geoff (BACKLOG #131). The `modernc.org/sqlite`
  1.62.0 minor stays out because it changes the busy timeout and fails two `store` tests (#129).
