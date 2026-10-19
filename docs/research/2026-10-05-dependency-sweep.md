# Dependency sweep, 2026-10-05

Task 0 of the navidrome-removal pass, run at `7ad24cc` on master. Scope as in the 2026-10-04 sweep
(`2026-10-04-dependency-sweep.md`): the Go module, `web/`, the Python oracle venv, and the pinned
binaries. `_salvage/` is out of scope. The next sweep starts from the "after" column.

## Outcome

One bump landed.

| Package | Before | After | Kind | Commit |
|---|---|---|---|---|
| @sveltejs/kit | 3.0.0 | 3.0.1 | npm patch | `cf3299f` |

3.0.1's only change: relative imports in generated types work under `nodenext` module resolution.
The repo's `tsconfig` does not use `nodenext`, so nothing in range touches a call site. The full gate was green
before the commit (it covered the whole tree, the STATUS edit included).

Everything else is at its newest non-major, verified against the registries today: go.mod's modules
(`go list -m -u all` reports only graph-only modules that go.mod does not list: x/net, x/tools, yaml/v3,
pprof, go-md2man, check.v1, cc/ccgo/gc), Go 1.27.1, `web/` (svelte 5.57.1, svelte-check 4.7.6, daisyui
5.7.47, vite 8.3.2, tailwindcss and @tailwindcss/vite 4.3.3, vite-plugin-svelte 7.3.1, adapter-static
4.0.0, playwright 1.63.0), every pin in `tools/venv/requirements.txt` (beets 2.14.1 and all transitive
pins), golangci-lint v2.14.0. `npm audit`: 0 vulnerabilities.

## Held

| Package | Held at | Latest | Unblocks when |
|---|---|---|---|
| typescript 7 | 6.0.3 | 7.0.2 | svelte-check and @sveltejs/kit declare a `^7` peer range, and sveltejs/language-tools#3063 is closed |
| @types/node 26 | 24.19.1 | 26.6.4 | Node 26 reaches LTS and the workstation's mise Node moves to 26; bump both in one change |
| rclone v1.75.1 (workstation) | v1.74.3 (PATH) | v1.75.1 | Fedora ships v1.75.x for the workstation binary the backup tests drive; the box takes it at 5d |

No PATH-resolved binary was bumped, so slice 2's worktree gate is unaffected.

## Survey: take-now and file decisions

Nothing in range provides something the repo hand-rolls. No take-now item. The two candidates carried from
the 2026-10-04 record stand: the stale "adds a module dependency" comment at
`internal/engine/musicbrainz/query.go:235-238` is corrected at that package's next touch, and sqlite
`StrictPragmas` stays optional. Go-side gate risks recorded on 2026-10-04 (rclone `lsjson` exit 3) are
unchanged, since rclone did not move.

## The baseline count (spec, rung 12, "The streamer's configuration")

Built at `cf3299f` (the commit after the sweep). Command, with `$SCRATCH` directories for every path-valued key
and synthetic sample values for the rest:

```
env -i HOME="$HOME" PATH=/usr/bin:/bin DUBPLATE_LIBRARY_ROOT=$SCRATCH/lib DUBPLATE_STATE_DIR=$SCRATCH/state \
  DUBPLATE_ROSTER_FILE=$SCRATCH/state/roster/roster.json DUBPLATE_SITE_NAME=Sample \
  DUBPLATE_ADMIN_EMAIL=admin@example.com DUBPLATE_IDENTITY_SOURCE=magic-link \
  DUBPLATE_PUBLIC_URL=https://music.example.com DUBPLATE_MAGIC_LINK_SECRET=<48 hex> \
  DUBPLATE_SESSION_SECRET=<48 hex> DUBPLATE_SMTP_HOST=smtp.example.com DUBPLATE_SMTP_PORT=587 \
  DUBPLATE_SMTP_USER=sample DUBPLATE_SMTP_PASSWORD=<sample> DUBPLATE_SMTP_FROM=music@example.com \
  ./dubplate-server check
```

Three rounds. The bare run reported five keys unset, the second nine, and the third printed
`configuration: ok` (and `store: absent`). Fourteen required keys, matching the review's run at `7036ebc`.
`DUBPLATE_TRUSTED_PROXY_ENABLED` is not required (the check passes without it), so it is outside the fourteen;
the template carries it as an explicit `false` literal.

| Class | Count | Keys |
|---|---|---|
| Env-template literal | 6 | `DUBPLATE_LIBRARY_ROOT`, `DUBPLATE_STATE_DIR`, `DUBPLATE_SITE_NAME`, `DUBPLATE_IDENTITY_SOURCE`, `DUBPLATE_ROSTER_FILE`, `DUBPLATE_PUBLIC_URL` |
| Generated or minted secret | 2 | `DUBPLATE_MAGIC_LINK_SECRET`, `DUBPLATE_SESSION_SECRET` |
| Operator-typed | 6 | `DUBPLATE_ADMIN_EMAIL`, `DUBPLATE_SMTP_HOST`, `DUBPLATE_SMTP_PORT`, `DUBPLATE_SMTP_USER`, `DUBPLATE_SMTP_PASSWORD`, `DUBPLATE_SMTP_FROM` |

The baseline count is the operator-typed class: **6**. Reading note: the template's literals for
`DUBPLATE_SITE_NAME` (`907`) and `DUBPLATE_PUBLIC_URL` (`https://music.907.life`) are this site's values, so a
stranger's install types those two as well (8 under that reading). The six-key and eight-key readings are both
auditable from the table. The two minted secrets are rendered from the age store by `deploy.sh`; the six
operator-typed rows are bare rows resolved from the same store.

Uncounted non-key inputs a default install also supplies by name: the admin bootstrap's `member add`
arguments (it creates the roster file, which `doctor` reports missing until then), and the
`dubplate-server unidecode install` step.
