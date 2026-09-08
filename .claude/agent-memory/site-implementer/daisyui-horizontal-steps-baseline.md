---
name: daisyui-horizontal-steps-baseline
description: DaisyUI horizontal .step centers content in a row shared by all steps; a wrapped sublabel lifts a shorter neighbor's label off the shared baseline unless each block gets self-start.
metadata:
  type: feedback
---

DaisyUI's `.steps-horizontal .step` is `grid-template-rows:40px 1fr; place-items:center`.
All `<li class="step">` siblings sit in the outer `<ul>`'s one shared grid row, so that
row's height is set by whichever step's content is tallest. Every step's own content
block is then vertically centered inside that shared, stretched height. Result: a step
whose sublabel wraps to two lines sits near the top of the (now-tall) cell, while a
neighbor with a one-line sublabel gets centered in the same tall cell and its label ends
up lower, so different labels land on different baselines even though nothing about
their own content changed (dubplate U13 F6.4, confirmed against the built captures at both a
660px iPad-portrait width and 1440px desktop).

**Fix**: add `self-start` to each step's own content wrapper, overriding the parent's
`place-items:center` on the block axis only. Labels then anchor to the top of their
cell regardless of how many lines the sublabel below them wraps to.

**Also**: if the horizontal breakpoint activates inside a narrow content column (e.g. an
app shell's own `max-w-2xl`), the *viewport* breakpoint is not the number that matters,
the *column* width is. Check what width the steps actually render at before picking
`sm:`/`md:`/`lg:steps-horizontal`: a `sm:` breakpoint can fire while the real available
width is still mobile-narrow because the column itself hasn't grown yet.

**How to apply**: any DaisyUI horizontal `.steps` with variable-length sublabels, in any
cairn/dubplate-family site, needs `self-start` on the step content block. This is a
mechanic (always right, not content-dependent), a candidate for [[daisyui-semantic-theme-tokens]]-style
engine filing if it recurs on a second repo.

**Superseded in dubplate (2026-09-01)**: dubplate's JourneySteps dropped horizontal mode
entirely per the ratified mockup (a vertical spine, stacked markers, labels to the right,
at every viewport) -- the `lg:steps-horizontal` responsive switch and the `self-start`
workaround it needed are both gone from that component now. The baseline problem above is
still real for any site that keeps horizontal mode; it just no longer applies to dubplate's
journey spine specifically.
