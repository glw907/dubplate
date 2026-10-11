# Dependency sweep, 2026-10-11

Scope: the Go module, `web/`, the Python oracle venv, and the pinned binaries (rclone, golangci-lint, Playwright,
libxml2). The opening sweep of `admin-api-1`, one day after `2026-10-10-dependency-sweep.md`. The next sweep starts
from the "after" column.

## Before and after

| Package | Before | After | Kind |
|---|---|---|---|
| numpy | 2.5.3 | 2.5.4 | Python venv patch |
| platformdirs | 4.12.4 | 4.13.0 | Python venv minor |
| typescript | 6.0.3 | 6.0.3 | npm MAJOR 7.0.2 held (BACKLOG) |
| @types/node | 24.19.2 | 24.19.2 | npm MAJOR 26.6.5 held, follows the Node line |
| rclone (workstation) | v1.74.3 | v1.74.3 | binary; v1.75.2 recorded, not installed (writes outside the worktree and into the dotfiles manifests) |

Already latest, verified today: every `go.mod` requirement (`go list -m -u` reports nothing for a direct
requirement; `modernc.org/libc` v1.77.1 moves only with sqlite, whose v1.60.1 is current), every `web/package.json`
entry (`npm outdated` reports only the two majors above; vite 8.3.4, svelte 5.57.2, kit 3.0.1, daisyui 5.7.47,
tailwindcss 4.3.3, playwright 1.64.0 by `npm view`), golangci-lint v2.14.0, and libxml2 2.15.4 (held by `brew pin`,
not touched). Playwright stays at 1.64.0 because a newer one needs its matching Chromium in the user cache.

`go list -m -u all` also reports graph-only modules that `go.mod` does not list (x/net, x/mod, x/tools, yaml/v3,
testify, and others). They stay where minimal version selection puts them, as in the 2026-10-04 and 2026-10-10 records.

## Survey

numpy 2.5.3 to 2.5.4 and platformdirs 4.12.4 to 4.13.0 are transitive pins of beets, used only by the workstation
oracle (`make derive`), which the gate never runs. Features to leverage: none; the repo calls neither directly.
Gate risk: none, since the gate never imports the venv.

No capability a bump adds needs this pass's spec: none (nothing filed).
