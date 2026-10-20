# Dependency sweep, 2026-10-07

Task 0 of rung 6d (`docs/plans/2026-09-26-rung-6d-identity-serve-store.md`), run at
`7cb9722` on master. Scope as in the 2026-10-05 sweep (`2026-10-05-dependency-sweep.md`): the Go
module, `web/`, the Python oracle venv, and the pinned binaries. `_salvage/` is out of scope. The next
sweep starts from the "after" column.

## Outcome

Two direct bumps and one lockfile refresh landed, each in its own commit with the full gate green.

| Package | Before | After | Kind | Commit |
|---|---|---|---|---|
| svelte | 5.57.1 | 5.57.2 | npm patch | `f58a559` |
| vite | 8.3.2 | 8.3.3 | npm patch | `0782059` |
| acorn (transitive) | 8.18.0 | 8.19.0 | lockfile refresh | the commit before this record |
| magic-string (transitive) | 1.4.2 | 1.4.3 | lockfile refresh | the commit before this record |
| nanoid (transitive) | 3.3.19 | 3.3.20 | lockfile refresh | the commit before this record |
| postcss (transitive) | 8.5.28 | 8.5.29 | lockfile refresh | the commit before this record |

The refresh deleted `web/node_modules` and `web/package-lock.json` and reinstalled, so packages that
the exact top-level pins admit moved too. No `package.json` range changed in that commit.

Everything else is at its newest non-major, verified against the registries today:

- go.mod's listed modules. `go list -m -u all` reports only graph-only modules that go.mod does not
  list (x/net, x/tools, yaml/v3, pprof, go-md2man, check.v1, cc/ccgo/gc). Go is 1.27.1, the newest.
- `web/`'s other pins (@sveltejs/kit 3.0.1, svelte-check 4.7.6, daisyui 5.7.47, tailwindcss and
  @tailwindcss/vite 4.3.3, vite-plugin-svelte 7.3.1, adapter-static 4.0.0, playwright 1.63.0).
  @types/node 24.19.1 is the newest 24.x, and typescript 6.0.3 the newest 6.x.
- All 20 pins in `tools/venv/requirements.txt`, checked against PyPI.
- golangci-lint v2.14.0.

`npm audit`: 0 vulnerabilities. No PATH-resolved binary moved. Slice 2 had not merged
`getkin/kin-openapi` at sweep time, so it is not in scope.

## Held

| Package | Held at | Latest | Unblocks when |
|---|---|---|---|
| typescript 7 | 6.0.3 | 7.0.2 | svelte-check and @sveltejs/kit declare a `^7` peer range, and sveltejs/language-tools#3063 is closed |
| @types/node 26 | 24.19.1 | 26.6.4 | Node 26 reaches LTS and the workstation's mise Node moves to 26; bump both in one change |
| rclone v1.75.1 (workstation) | v1.74.3 (PATH, Fedora `rclone-1.74.3-1.fc44`) | v1.75.1 | Fedora ships v1.75.x for the workstation binary the backup tests drive; the box takes it at 5d |

## Survey: take-now and file decisions

svelte 5.57.2 is runtime and compiler bug fixes. Two of them add compile errors: an `{:else}` followed
by another `{:else}` or `{:else if}`, and a snippet declaration that redeclares a parameter. A third
refuses `let:` on a component with a `children` snippet. `svelte-check` and the static build passed,
so no component meets any of them.

vite 8.3.3 is dev-server hardening (`fs.serve` checks `?vite-wasm-instance`; `safeModulePaths` stores
ids) plus a `transformIndexHtml` filename fix and a launch-editor bump. The repo ships the static
build only, and the rendered gate's two-build byte comparison passed.

Nothing in range provides something the repo hand-rolls. No take-now item, and nothing to file. The
two candidates carried from the 2026-10-04 record stand: the stale "adds a module dependency" comment
at `internal/engine/musicbrainz/query.go:235-238` is corrected at that package's next touch, and
sqlite `StrictPragmas` stays optional. The Go-side gate risk recorded on 2026-10-04 (rclone `lsjson`
exit 3) is unchanged, because rclone did not move.
