---
name: daisyui-size-field-touch-targets
description: DaisyUI v5's --size-field theme token controls the 44px touch-target floor for .btn/.input/.select, but not .btn-sm
metadata:
  type: project
---

DaisyUI v5's plain `.btn` computes its height as `--size-field * 10`;
`.btn-sm` as `--size-field * 8`; `.btn-lg` as `--size-field * 12`. The
default `--size-field: 0.25rem` puts a plain `.btn` at 40px, under the
44px touch-target floor (WCAG 2.5.5 AAA / platform touch guidance).

**Why:** dubplate's uploader (`web/src/app.css`) was built with inputs
correctly meeting 44px via an explicit `min-h-11` utility, but every
button was left at DaisyUI's default and missed it (found and fixed
U13 fold F5.3/F5.9).

**How to apply:** setting `--size-field: 0.275rem` in a DaisyUI v5
theme block lands a plain `.btn` at exactly 44px (10 * 0.275rem =
2.75rem), for free, across every plain-`.btn` call site in the app
(inputs/selects that already carry their own `min-h-11` are unaffected,
since 0.275rem also grows their native height to the same 44px). It
does *not* fix `.btn-sm` (8 * 0.275rem = 2.2rem = 35.2px still), which
still needs an explicit `min-h-11` override at each call site. Check
both `@plugin 'daisyui/theme'` blocks (light and dark) if the site has
two, as dubplate does; the token is per-theme, not global.
