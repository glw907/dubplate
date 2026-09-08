---
name: daisyui-alert-grid-single-wrapper
description: DaisyUI 5 .alert is display:grid; flex-col/items-start on it is inert. Fix is one wrapper div, not more flex utilities.
metadata:
  type: feedback
---

DaisyUI 5's `.alert` class is `display:grid; grid-auto-flow:column; place-items:center
start`, and gains `grid-template-columns:auto minmax(auto,1fr)` the moment it has a
second direct child (`&:has(>:nth-child(2))`). `flex-col items-start` on the alert div
does nothing (those are flexbox properties on a grid container), so a heading plus two
paragraphs plus a button row renders as four separate grid columns (dubplate U13 F6.1;
this is exactly the shape the [[daisyui-semantic-theme-tokens]]-style engine mechanic
list should probably carry too, since every cairn/dubplate site reaches for
`.alert flex-col items-start` reflexively).

**Fix**: give `.alert` exactly one child, a wrapper `<div class="flex flex-col
items-start gap-N">` holding the heading/paragraphs/actions, so the grid's own
`:has(>:nth-child(2))` selector never fires and normal flex layout runs inside the
wrapper. Do not reach for `alert-vertical`: it also sets `text-align:center` and
`justify-items:center`, which is wrong for left-aligned multi-paragraph body copy (only
right for a short centered banner).

**Why**: dubplate's built rejected/pushed-back alerts rendered as newspaper columns,
one word per line, 20 lines tall, in every OutcomeAlert/Journey/History alert with
2+ children, confirmed against the real DaisyUI 5.7.25 CSS, not a hunch.

**How to apply**: any `.alert` with more than a heading, in any cairn/dubplate-family
Svelte 5 + DaisyUI 5 site, needs the single-child-wrapper shape. Grep for `alert
flex-col` or `alert .*items-start` as the drift signal.
