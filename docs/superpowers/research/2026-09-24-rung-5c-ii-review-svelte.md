# Rung 5c-ii pass-end review: Svelte

Reviewer: `svelte-reviewer` (Opus 5.5), over `git diff 0c2745f HEAD -- web/` at `2424b9a`. It
read every file, the built output, `internal/serve/csp.go`, the Makefile `web` target, and plan
revision 2.4's R-5cii2, R-5cii22, and Task 3, and probed the vacuity findings in scratch space.
No server-only SvelteKit file exists; runes, `$app/state`, props, and migration patterns are
clean; the config, prerender under a relative base, and the stubs are correct.

## Blocker

- **B1.** Every gate script passes silently when the repo path contains a space or any
  percent-encoded character. The main guard ``import.meta.url === `file://${process.argv[1]}` ``
  compares a percent-encoded URL with a raw path, so `main()` never runs and node exits 0
  (`server-files.ts:60`, `g1.ts:123`, `g2.ts:645`, `g3.ts:71`, `g9.ts:257`,
  `theme-bundle.ts:534`, `install-state.ts:74`). Proven in `scratchpad/a b/`. Fix: move the
  shared parser into `lib.ts` and call `main()` unconditionally, or use `import.meta.main`
  (Node 24.2+, raising the Makefile floor), or compare against
  `pathToFileURL(realpathSync(process.argv[1])).href`.

## Warnings

- **W1.** Dropping the `inks.css` import passes every gate (`theme-bundle.ts:441-452` checks
  only structural and palette names for the composed themes). Add `INK_NAMES`.
- **W2.** G9 never compares a `contrast-exception` marker's recorded ratio with the measured one
  (`g9.ts:158-168`; parsed at `g2.ts:79`). Fail on a mismatch; consider enforcing a floor
  named in the reason.
- **W3.** G9 accepts markers on `-content` pairs and `--color-base-content` (`g9.ts:215`, `:219`),
  which are always text. Pass `null` and report any such marker.
- **W4.** DaisyUI's `--noise: 1` texture is a `data:` SVG background image; `img-src` falls back to
  `default-src 'self'`, so the CSP (meta and Go header) blocks it on every component
  (`dubplate.css:35,50`, `system.css:27,65`, `svelte.config.js:12-21`, `csp.go:29-31`). Add
  `img-src 'self' data:` in both, or set `--noise: 0`; confirm rendered in 5c-iii.
- **W5.** G1 misses `style:color="#ff0000"`, `style:background-color="red"`, SVG `fill=`, a
  multi-line `class` value, and hex in `.ts` (`g1.ts:81`, `:69`). Flag any hex token outside
  comments and `lib/themes/`; cover `style:<prop>=` and `fill=`/`stroke=`/`stop-color=`; scan
  `.svg`.
- **W6.** G3 misses `src=` loads and does not scan `.ts` (`g3.ts:14`, `:19-20`).
- **W7.** The server-file check misses `src/hooks.server/index.ts` and a custom
  `kit.files.hooks.server` (`server-files.ts:15-23`; kit `filesystem.js:171-196`).
- **W8.** CSP directives live in two places that can drift (`svelte.config.js:12-21`,
  `csp.go:29-31`); leave only `script-src` in the kit config or add a Go test pinning equality.

## Suggestions

S1 theme-bundle's effective value ignores layers and specificity; S2 the (0,1,0) claim in
`system.css:13-14` is wrong for the built `:is(...)` selector (0,4,1); S3 G2 and G9 read only the
first `@layer base`; S4 `walk()` swallows every `readdirSync` error; S5 add
`erasableSyntaxOnly: true`; S6 type the layout props with `LayoutProps`; S7 the `dark:` scan skips
`.css`; S8 the placeholders point at themselves; S9 `TestNoFlashStaticHalf` hardcodes two of
three documents; S10 process citations in `embed_test.go`; S11 builds are not reproducible
(`kit.version.name` is a timestamp, so the CSP hash changes per build); S12 a 5c-iii note on
absolute hrefs escaping the prefix and the root favicon request.
