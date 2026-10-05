# web

SvelteKit in static mode (`adapter-static`, `ssr = false`, no
fallback): every route prerenders to a plain HTML file under `build/`,
which `dubplate-server` embeds and serves. There is no Node runtime on
the box; `npm run build` runs on a workstation and the output ships as
static assets.

## Building

```sh
npm ci --prefix web
npm run build --prefix web
```

`npm run check --prefix web` runs `svelte-check` under strict
TypeScript, including the gate scripts under `scripts/` (see
`tsconfig.json`'s `include`).

## Rules

- No server-only SvelteKit file (`+page.server.*`, `+layout.server.*`,
  `+server.*`, `hooks.server.*`) and no form action. `scripts/check.sh`
  gates this from the first commit.
- No hardcoded color outside `src/lib/themes/`, and no external origin
  in source or in the built bundle. Both are gated scripts, not review
  conventions.
- `web/build/` is not committed except its placeholder file, which
  keeps a fresh checkout's `//go:embed` directive compiling before
  anyone has run the build.

See the repo's `CLAUDE.md` for the full web stack rule.
