---
name: daisyui-soft-pastel-vs-saturated
description: DaisyUI 5 semantic theme tokens need to be saturated colors; alert-soft/badge-soft handle the tinted-surface look on their own.
metadata:
  type: feedback
---

DaisyUI 5's `alert-soft`/`badge-soft` modifiers already compute their own
tinted background from the base `--color-error`/`-success`/`-warning`/
`-info`/`-primary` token via `color-mix`. If those tokens are themselves
pastel/light-content pairs (a common mistake when hand-authoring oklch
values without checking against a reference), the soft surfaces come out
looking flat/pastel-on-pastel instead of the intended "dark-tinted
background, colored light text" look, because `-soft` blends toward the
base color, not toward vividness.

**Why:** U13 fold F7 found dubplate/web/src/app.css's dark theme with
--color-primary at 85% L (near-white) and --color-error/-warning/-info all
70-85% L -- pastel by construction, not a soft-modifier problem. Swapping
in saturated MUI 600/700-family reads (e.g. primary oklch(66% 0.169 249),
error oklch(57% 0.2 26)) fixed both the solid buttons/step discs AND the
soft alert/badge surfaces at once, with no markup change needed beyond
adding the `-soft` class itself.

**How to apply:** When a visual-fidelity task reports "pastel surface with
dark text" on a DaisyUI theme, check `app.css`'s token *values* first
(oklch L above ~70% for a color meant to read as vivid is the tell) before
assuming the component markup needs the soft/outline/dash modifiers
adjusted. Verify by rendering the actual compiled CSS against sample
markup (a throwaway static HTML file link-rel'd to `dist/assets/*.css`,
screenshotted with `chromium-shot`) rather than trying to drive the real
app through file-upload flows just to reach a terminal alert state --
sample markup verifies the CSS/token behavior directly and is much
cheaper than reproducing the interaction that reaches that state.

See also [[daisyui-semantic-theme-tokens]] for the missing-token variant
of this same class of bug.
