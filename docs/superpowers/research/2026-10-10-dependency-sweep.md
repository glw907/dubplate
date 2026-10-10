# Dependency sweep, 2026-10-10

Scope: the Go module, `web/`, the Python oracle venv, and the pinned binaries
(rclone, golangci-lint, Playwright, libxml2). Release notes were read for
exactly the range below. The next sweep starts from the "after" column.
Opening sweep of the gate scope pass.

## Before table

| Package | Current | Target | Kind |
|---|---|---|---|
| golang.org/x/sys | v0.48.0 | v0.49.0 | Go direct |
| golang.org/x/crypto | v0.57.0 | v0.58.0 | Go indirect (go-mail NTLM `md4`) |
| golang.org/x/text | v0.42.0 | v0.43.0 | Go indirect (go-mail precis) |
| vite | 8.3.3 | 8.3.4 | npm patch |
| @types/node | 24.19.1 | 24.19.2 | npm patch (24.x line) |
| playwright | 1.63.0 | 1.64.0 | npm minor, pinned binary |
| rolldown bindings (transitive) | 1.2.12 | 1.2.13 | npm, moved with vite |
| @oxc-project/types (transitive) | 0.152.0 | 0.153.0 | npm, moved with vite |
| platformdirs | 4.12.3 | 4.12.4 | Python venv patch |
| typescript | 6.0.3 | 7.0.2 | npm MAJOR, held |
| @types/node | 24.x | 26.6.5 | npm MAJOR, held |
| rclone (workstation) | v1.74.3 | v1.75.2 | binary, held |

Already latest, verified: cobra v1.10.2, go-mail v0.8.1, x/time v0.16.0,
kin-openapi v0.149.0, modernc.org/sqlite v1.60.1, modernc.org/libc v1.77.1
(moves only with sqlite), golangci-lint v2.14.0, @sveltejs/kit 3.0.1,
adapter-static 4.0.0, svelte 5.57.2, svelte-check 4.7.6,
@sveltejs/vite-plugin-svelte 7.3.1, tailwindcss and @tailwindcss/vite 4.3.3,
daisyui 5.7.47, beets 2.14.1 and every other venv pin. libxml2 stays held by
`brew pin libxml2` and is not touched.

`go list -m -u all` also reports graph-only modules that go.mod does not list
and no package of the build links (x/net, x/sync, x/mod, x/term, x/tools,
yaml/v3, pprof, ccgo/v4, cc/v4, gc/v3, go-md2man, testify, pty, regexp2, and
others). They stay where minimal version selection puts them, as in the
2026-10-04 record. `go mod tidy` moved x/sync and x/tools sums as a
side effect of x/crypto and x/sys; no package of ours imports them.

## Go

### golang.org/x/sys v0.48.0 to v0.49.0

The only direct import is `cmd/dubplate-server/serve.go`. The range carries
platform constant and syscall-table churn. Nothing relevant; features to
leverage: none.

### golang.org/x/crypto v0.57.0 to v0.58.0

Only `md4` is linked (go-mail NTLM). Nothing in range touches it. Nothing.

### golang.org/x/text v0.42.0 to v0.43.0

Linked through go-mail's `secure/precis`; the repo imports no x/text package.
Nothing relevant.

## web/

### vite 8.3.3 to 8.3.4

Fixes: preload correctness with `chunkImportMap` on a non-root base, waiting
for in-flight stylesheets in the preload helper, lightningcss preprocessor
resolution, a rolldown-related dependency refresh (rolldown 1.2.12 to 1.2.13).
Gate risk: the dependency refresh can change emitted bytes; the byte-for-byte
leg compares two builds of one source, and the CSP-hash agreement re-derives
from the build. Both ran in the gate. Features to leverage: none (the
`bundled-dev` additions are dev-server only).

### @types/node 24.19.1 to 24.19.2

Type-only patch on the line that matches Node 24.20.0. Nothing.

### playwright 1.63.0 to 1.64.0

Features: `page.webmcp`, video frame rate, `Locator.within()`, test runner
additions. None replaces anything `web/scripts/rendered/run.ts` hand-rolls.
Practice to change: the `screen` property is now forwarded from device
descriptors. `run.ts` does not use `devices[...]`, so nothing changes.
The bundled Chromium moves from 153 to 156 (revision 1248, already in the
workstation cache). Gate risk: the rendered leg's computed-value expectations
(`web/scripts/rendered/run.ts`, the 2026-09-25 probe comments at the
Chromium-153 notes) were re-run under 156 in the full gate. The two probe
comments name the version they were measured on and stay as dated records.
`CLAUDE.md` still says "Playwright 1.63.0" in the Gate section; the conductor
updates it.

## Python oracle venv

### platformdirs 4.12.3 to 4.12.4

A patch; the derivation scripts set `BEETSDIR` before beets imports, so beets
never resolves a platformdirs config path (see the 2026-10-04 record). It
cannot change beets output. `make derive` was not run (it needs the network);
run it once before the next oracle-dependent rung and expect no diff.

## Binaries

### rclone v1.74.3 to v1.75.2

The workstation binary is Fedora's `/usr/bin/rclone` at v1.74.3, outside this
repo's control. Held under the 2026-10-04 ruling until Fedora ships v1.75.x.
The v1.75.2 notes point to the changelog only. When it is taken, re-measure
`lsjson` exit 3 for an empty directory (internal/backup/rclone.go) as the
2026-10-04 record says.

### golangci-lint

v2.14.0 is the newest release. Nothing to move.

## Refactor candidates

No package in range provides something the repo hand-rolls. The earlier filed
items (sqlite `StrictPragmas`, the stale `x/text/cases` rationale at
internal/engine/musicbrainz/query.go) stand unchanged.

## Majors held

| Package | Held at | Unblocks when |
|---|---|---|
| typescript 7.0.2 | 6.0.3 | svelte-check (4.7.6 declares `typescript ^5 \|\| ^6`) and @sveltejs/kit (3.0.1 declares `^6`) declare a `^7` peer range, and sveltejs/language-tools#3063 is closed. Geoff's 2026-10-10 ruling takes it only if svelte-check supports it; it does not, so it is held. |
| @types/node 26.6.5 | 24.19.2 | Node 26 reaches LTS and the workstation's mise Node moves to 26; bump both in one change. Until then it stays on the newest 24.x. |
| rclone v1.75.2 | v1.74.3 | Fedora ships v1.75.x for the workstation binary; the box takes it at 5d. |
