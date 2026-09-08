---
name: daisyui-semantic-theme-tokens
description: DaisyUI 5 alert/badge/step color modifiers silently no-op if a theme omits --color-error/-success/-warning/-info; how to source and contrast-check replacements from a Material palette family.
metadata:
  type: feedback
---

DaisyUI 5's `.alert-error`, `.badge-success`, `.step-error` etc. all read
`var(--color-error)` (or success/warning/info) with a `base-200` fallback.
If a custom theme block defines only base/primary/secondary/neutral (as
dubplate's `navidrome-*` themes did through U12), every semantic color
modifier computes to invalid-at-computed-value-time and silently falls
back to base-200 -- no build warning, no lint error, just a rejected
album rendering the same gray as an accepted one. `npm run check` and
`npm test` both stay green through this; only grepping the built CSS
(`grep -- '--color-error:' dist/assets/*.css`) or a visual capture
catches it.

**Why:** confirmed as U13 fold F4 (blocker F4.1/F4.4), proven three ways
including a probe render where the step spine's completed/failed discs
had no marker at all while unstarted steps (falling to the real
base-300 default) showed solid discs -- failure was *less* visible than
the untouched state.

**How to apply:** any DaisyUI custom theme in this repo (or a sibling
cairn-family site) must define all eight tokens
(error/success/warning/info x base/-content) in every theme block, not
just primary/secondary/neutral. When picking values: read the same
Material/MUI palette family the base colors were sourced from (this repo's
`app.css` header cites `ui/src/themes/{light,dark}.js`) rather than
inventing new hues, then contrast-check with the WCAG relative-luminance
formula: >=4.5:1 between a color and its own `-content` pair, >=3:1
between the color and `base-100`. A Material color that's fine in a light
theme (e.g. MUI's `red 700`/`#d32f2f` for error) is usually too dark to
clear 3:1 against a near-black `base-100` in the dark theme -- lighten to
a lighter step of the same family (e.g. `red 300`/`#e57373`) and flip the
`-content` pair to dark text, mirroring how this repo's existing
`--color-primary-content` in the dark theme already uses dark text on a
light accent.

Proof surfaces to re-check after any theme token change: the journey
step-spine discs (`step-success`/`step-error` in JourneySteps.svelte) and
the history/file-row badge chips (`badge-success`/`-error`/`-warning`) --
both go fully invisible/uncolored when a token is missing, so they make
the defect obvious at a glance.
