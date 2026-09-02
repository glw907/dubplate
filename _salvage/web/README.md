# dubplate uploader web

The Svelte 5 / DaisyUI single-page app `dubplate serve` embeds and serves
under `/upload/`. See `docs/superpowers/plans/2026-08-31-dubplate-uploader.md`
(task U9 and later) for the design this implements.

## Dev loop

```sh
npm ci
npm run dev
```

`npm run dev` starts Vite's dev server. Its proxy (`web/vite.config.ts`)
forwards `/upload/api` and `/upload/files`—the exact paths
`internal/serve/serve.go`'s `buildMux` mounts the JSON API and the tus
protocol under—to `http://localhost:8080`, the port `vite.config.ts`
names as `devBackend`.

In a second terminal, run the real backend on that port:

```sh
DUBPLATE_SERVE_PORT=8080 dubplate serve --dev-identity you@example.com
```

`--dev-identity` skips Access verification and treats every request as
that email; it refuses to start unless every bind address is loopback.
Every other `DUBPLATE_*` variable `internal/config.Load` requires still
needs a value (see that file)—set them in a workstation `.env` your
shell sources, never hardcoded here.

## Checks

- `npm run check`—`svelte-check`, must report 0 errors and 0 warnings.
- `npm run test`—`vitest run`, the unit suite (`*.test.ts` next to the
  code it covers) against a stubbed `TusDriver`/`fetch`, never the real
  network.
- `npm run build`—the production Vite build to `web/dist`, base
  `/upload/`, that `go:embed` (`web/embed.go`) ships in the binary.

`scripts/check.sh` runs all three (via `npm ci` first) as part of the
repo-wide gate; `make check` runs that script. A fresh checkout ships
`web/dist/placeholder.txt` (see `.gitignore`) so `go build` stays green
before anyone has run `npm run build` once.

## Pinned dependencies

Versions verified against the npm registry 2026-09-01
(`docs/superpowers/research/passb-stack-brief.md`); each pin's own note:

- `svelte` `5.57.0`—runes mode; no peer deps of its own.
- `vite` `8.2.2`—current major, Rolldown-backed bundler.
- `@sveltejs/vite-plugin-svelte` `7.3.0`—peers `vite ^8.0.0-beta.7 || ^8.0.0`, `svelte ^5.46.4`, both satisfied above.
- `tailwindcss` `4.3.3`—CSS-first config, no `tailwind.config.js`.
- `@tailwindcss/vite` `4.3.3`—peer `vite ^5.2.0 || ^6 || ^7 || ^8`.
- `daisyui` `5.7.25`—wired from CSS via `@plugin "daisyui";`, not a JS import or `tailwind.config.js` registration (that pattern is v3/v4-Tailwind only and no-ops under Tailwind 4).
- `tus-js-client` `4.3.1`—the stable `latest` dist-tag. **Named watch item:** a `5.0.0-pre2` exists under the `next` tag; do not move to it until it publishes stable, since it is pre-release and this pin is deliberate.
- `svelte-check` `4.7.6`, `typescript` `6.0.3`, `vitest` `4.1.11`, `jsdom` `30.0.1`—tooling only, tracked loosely with the pins above rather than independently pinned to a brief entry.
- `@testing-library/svelte` `5.4.2`—Svelte 5 component rendering for the `*.test.ts` suites, peer `svelte ^5.0.0` satisfied above.
- `@testing-library/jest-dom` `7.0.1`—the `toBeInTheDocument`/`toHaveClass` matchers those same suites assert with.
