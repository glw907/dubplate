---
name: daisyui-soft-reads-color-token-as-text
description: DaisyUI 5's alert-soft/badge-soft set text color to the raw semantic token itself, not -content; a mid-saturation token reads dark-on-dark until you check the compiled CSS.
metadata:
  type: feedback
---

DaisyUI 5's compiled CSS (`node_modules/daisyui/components/alert.css`) shows
`.alert-soft{color:var(--alert-color,...)}` where `--alert-color` is literally
`--color-error`/`-success`/`-info`/`-warning` itself, never `-content`.
`.badge-soft` uses the same pattern. So a token tuned for use as a solid
background (MUI 600/700-family mid-saturation) becomes the actual text color
against the tinted -soft background, which on a near-black dark ground reads
around 2.2-3:1, not the state text's 4.5:1 floor.

**Why:** related to [[daisyui-soft-pastel-vs-saturated]], but the inverse
failure mode. That memory is about the base token being too pastel; this one
is about a saturated-but-mid-lightness base token being too dark once it's
used as *text* rather than a fill. Caught only by reading the alert/badge
CSS source or computing contrast numerically (oklch->linear-sRGB->relative
luminance); svelte-check and the component test suite are both blind to it.

**How to apply:** when tuning a dark theme's error/success/info/warning
tokens for a `-soft` alert/badge site, treat the base token's own lightness as
the thing that must clear 4.5:1 against the background, not the -content
pair. A Material 400-family read (lighter than 600/700) works well against a
near-black ground; -content then needs to flip to dark to stay legible on the
now-lighter solid-fill (disc/badge) usage. Separately, an alert's `<h2>`/`<p>`
inherit the alert's own color by default (no `color` set on them). For a
mockup that wants a fixed white heading regardless of alert state, give the
heading/body an explicit `text-base-content`/`text-base-content/80` override;
do not rely on retuning the token alone to fix heading contrast.
