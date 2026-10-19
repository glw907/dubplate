# Dependency sweep, 2026-10-12

Scope: the Go module, `web/`, the Python oracle venv, and the pinned binaries (rclone, golangci-lint, Playwright,
libxml2). The opening sweep of `intake-1`, one day after `2026-10-11-dependency-sweep.md`. The next sweep starts from
the "after" column.

## Before and after

| Package | Before | After | Kind |
|---|---|---|---|
| tus-js-client | absent | 4.3.1 | npm, added exact (P14); no importer until `intake-2`; newest by `npm view` |
| typescript | 6.0.3 | 6.0.3 | npm MAJOR 7.0.2 held (BACKLOG) |
| @types/node | 24.19.2 | 24.19.2 | npm MAJOR 26.6.5 held, follows the Node line |
| rclone (workstation) | v1.74.3 | v1.74.3 | binary; v1.75.2 recorded, not installed (writes outside the worktree and into the dotfiles manifests) |

Already latest, verified today: every `tools/venv/requirements.txt` pin (PyPI's newest for each of the twenty), every
`go.mod` requirement (`go list -m -u all` reports only graph-only modules that `go.mod` does not list, which stay where
minimal version selection puts them, as in the earlier records), every other `web/package.json` entry (`npm outdated`
reports only the two majors above; Playwright 1.64.0 by `npm view`), golangci-lint v2.14.0 (the latest release), and
libxml2 2.15.4 (held by `brew pin`, not touched).

Not yet in `go.mod`, recorded for T6 and T9 (P14): `github.com/mewkiz/flac` newest is v1.0.14 and
`github.com/tus/tusd/v2` newest is v2.10.1, both what the plan quotes, so neither moves the plan's citations.

## Survey

tus-js-client 4.3.1 is the pinned client for the uploader page (`intake-2`) and has no call site in this pass. Features
to leverage: none yet. Gate risk: none, since no `web/` gate checks for an unused dependency and the static bundle
imports nothing from it.

## Held majors

- TypeScript 7.0.2: unblocks when the Svelte and SvelteKit toolchain supports it (BACKLOG).
- @types/node 26.6.5: unblocks when the workstation's Node line reaches 26.
