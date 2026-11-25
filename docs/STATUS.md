# dubplate status

## Current state
The reshape is done (PRs #20 to #22, spec
`~/Projects/initiatives/cairn-dubplate-reshape/specs/2026-11-22-dubplate-reshape.md`): `govulncheck`, fuzz targets,
the race detector in CI, Dependabot, secret scanning, and push protection. Milestone `engine` (PR #17) is done, and
the citation leg of `scripts/check.sh` gates. Nothing a member or an operator sees changed.

Work `structure-base` (PR #18, merged) moved code only: `_salvage` left the tree, `internal/engine/store` became
`internal/store`, `internal/id` parses ids, and a store test names the owner of each of 19 tables.
Work `structure-keys` (PR #19, merged) moved every setting into the package that reads it, each with a `config.go`, a
`Register` function, and readers. `cmd/dubplate-server` declares and registers nothing, and an import test holds
configuration to `config.go` and `doctor.go` in seven packages. Output is unchanged.
Work `dev-friction` (PR #24, merged 2026-11-23) took `modernc.org/sqlite` v1.63.0 and certifi 2026.11.20, made
`TestBusyTimeout` wait on signals, queued lint runs, and made `scripts/check.sh` unset every `DUBPLATE_*` name.
Work `structure-doctor` (PR #25, merged 2026-11-23) moved the eleven health checks into the packages that know each
fact; `cmd/dubplate-server/doctor.go` only places the rows, and the output is byte-identical. Both followed
`~/Projects/initiatives/dubplate-server/specs/2026-11-22-overnight-design.md`.
Work `lean-pattern` (PR #26, merged 2026-11-23) moved the binding rules into 15 decision records in
`docs/decisions/` and cut `CLAUDE.md` to 85 lines. The gate now runs dupl, gocognit, forbidigo, `deadcode -test`,
a 200-line limit on `CLAUDE.md`, and a status check on each record.
Work `complexity-cleanup` (PR #27, merged 2026-11-23) split ten functions under gocognit's bar of 30; only
`commitVerdict` and `walkZIP` keep an exclusion.

`intake-1` is the last product work on master. A passing album waits in `matching` with no exit until
`intake-2`'s match probe lands. The run order is in BACKLOG, "Milestones to 0.8". Work still to do sits in BACKLOG
under its `#milestone:<name>` tag.

The box is unchanged: the dubplate-server binary `d1e93e9`, the serve unit installed and disabled, the `auth`,
`roster`, and `unidecode` leaves, one admin member, and `/srv/music/library` (6.6 GB).

## Next action
The architecture cleanup chain (`docs/specs/2026-11-23-architecture-cleanup.md`, approved by Geoff on 2026-11-23) paused
at the weekly usage ceiling on 2026-11-24. Tasks 1 to 3 are merged: `vocabulary` (#28), `wire-contract` (#29), and
`server-boundaries` (#30, merged with Geoff's waiver of the Current check for one STATUS-only commit). Task 4,
`server-access`, was stopped with five committed, unpushed commits in `.claude/worktrees/server-access` (branch
`worktree-server-access`, built on #30's old head); relaunch it from the spec's goal text with `master` as its base,
or let a session finish from that worktree. The chain runs in the session Geoff talks to, from this checkout, by
building-from-specs "Chains": after the weekly reset (2026-11-27, 23:00 America/Anchorage) the ceiling drops to 85
percent, and that session re-arms the spec's `/loop 15m` line. A second session in this checkout makes
`claude-worktree-busy` refuse every launch. Clear `/tmp` first: the gate's `go test` leg failed on a full tmpfs.

`intake-2` takes #159 after the chain's last task, `web`, merges. PR #25 lists findings its readers left unbuilt under `## For Geoff`, not yet filed in BACKLOG.
Before any deploy that carries `cleanup-store`, check read-only that the box store holds no restore-intent record (#148).

## Open decisions
- The repo is public on GitHub and stays so (Geoff, 2026-11-10). Open under it: BACKLOG #96's question of whether
  the Unidecode-derived table in git history needs a lawyer's reading or a history rewrite.
- `intake-2`'s first task is the three rulings of #92.
- R-6d21 overrides the settings-home rule that the server never reads its environment. Geoff may reopen it;
  nothing waits on it.
- The 0.8 and 0.9 launches are FLAC-only. Whether the group needs a non-FLAC format sooner is still open.
- The gate unsets the workstation's 16 `DUBPLATE_*` exports; a bare `go test` still passes them (BACKLOG #176).
- One Run-D.M.C. frame (`commons:139197098`) and five maybes wait in `photos/.candidates/` for Geoff's call.
- TypeScript 7.0.2 and `@types/node` 26.6.x are majors held for Geoff (BACKLOG #130).
- Geoff accepted on 2026-11-22 that a write under lock contention ends at its deadline on the new driver (#129).

## Launch prompt
Start the architecture cleanup chain from `docs/STATUS.md`, "Next action", by building-from-specs.
