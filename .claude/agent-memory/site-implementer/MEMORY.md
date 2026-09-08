# Agent memory: site-implementer (dubplate)

- [Commit only named paths](commit-only-named-paths.md) — a bare `git commit` here sweeps in another pass's pre-staged edits; use `--only`.
- [beets dry-run facts](beets-dry-run.md) — `--pretend` is not a matcher dry run; `-c` overlays, `BEETSDIR` replaces, `raise_on_error` separates outage from no-match.
- [systemd and Python subprocesses](systemd-python-subprocess.md) — what the hardened serve unit needed before it could run beets.
- [Admin identity in copy](admin-identity-in-copy.md) — no display-name field exists; AdminEmail alone is the established "the admin (email)" pattern.
- [DaisyUI semantic theme tokens](daisyui-semantic-theme-tokens.md) — a theme missing --color-error/-success/-warning/-info silently no-ops to gray; checks stay green, only a CSS grep or capture catches it.
- [DaisyUI --size-field touch targets](daisyui-size-field-touch-targets.md) — bumping --size-field fixes .btn's 40px default but not .btn-sm; each call site still needs min-h-11.
- [App.svelte has no store DI](app-svelte-no-store-di.md) — unlike History/Journey, App.svelte hardcodes the uploadStore singleton; App.test.ts resets it directly per test instead.
- [DaisyUI alert grid single wrapper](daisyui-alert-grid-single-wrapper.md) — `.alert` is display:grid; flex-col/items-start is inert; wrap content in one child div instead.
- [DaisyUI horizontal steps baseline](daisyui-horizontal-steps-baseline.md) — a wrapped sublabel lifts a shorter neighbor's label off the shared baseline unless each step content block gets self-start.
- [DaisyUI soft modifier needs saturated base tokens](daisyui-soft-pastel-vs-saturated.md) — alert-soft/badge-soft blend toward the base color; a pastel base token still renders pastel-soft.
- [DaisyUI -soft reads the base token as text color](daisyui-soft-reads-color-token-as-text.md) — a mid-saturation base token is fine as a fill but too dark as -soft text; heading/body need an explicit base-content override too.
