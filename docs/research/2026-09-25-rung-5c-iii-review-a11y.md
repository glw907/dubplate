# 5c-iii UI and accessibility review (daisyui-a11y-reviewer)

Range `6e6fd12..HEAD` over `web`, `internal/serve`, `cmd/dubplate-server/web.go`, graded against the
visual identity spec revision 1.3 and the W-rules. Measured in Chromium against the built bundle.

- **B1 (blocker).** `web/src/lib/styles/components.css:145-156`: `.badge { --badge-bg: transparent }`
  keeps DaisyUI's `color: var(--badge-fg)` (the fill's `-content` color), so badge text vanishes:
  `badge-warning` on label-night is 1.00:1; on label, primary 1.12, info/success/error 1.20,
  accent 1.00; on label-night, accent/info/success/error 1.00. A plain `.badge`'s base-200 border on
  base-100 is about 1.1:1 (W5.18). Fix: `color: var(--badge-ink, var(--color-base-content));
  border-color: var(--badge-ink, color-mix(in oklab, var(--color-base-content) 55%, transparent))`, a
  per-status `--badge-ink` (primary to `--ink-primary`, accent and success to `--ink-accent`, info,
  error, warning to their inks; secondary and neutral to base-content or refused), and a rendered row
  checking badge text at 4.5:1 in both themes (`run.ts:996-1000` checks existence only).
- **W1.** `components.css:58-61`: the `--ink-primary` ring on colored fills is 1.03 to 2.34 (label)
  and 1.05 to 2.00 (label-night), so focus inside an `.alert-*` is invisible. Extend the base-100 halo
  to focusables inside filled containers (`.alert :focus-visible { box-shadow: 0 0 0 6px
  var(--color-base-100) }`), add focus-row samples inside each alert variant, and a VI erratum ("every
  surface" measured the bases only).
- **W2.** `components.css:58`: DaisyUI 5's wrapped input (`<label class="input">`) and `.select` draw
  two rings in two colors; `.file-input:focus` keeps a base-content ring. Fix with `--input-color:
  var(--ink-primary)` on `:is(:focus,:focus-within)` and `outline: none` on the inner
  `:focus-visible`; add a wrapped-input fixture.
- **W3.** `components.css:75-80,89-92`: disabled buttons and `.btn-link` get the raised shadow and
  keyline. Fix: no shadow and a transparent border for `:is(:disabled,[disabled],.btn-disabled)` and
  `.btn-link`, and no `:active` translate.
- **W4.** `components.css:192`: links in prose look like body text until hover (1.4.1; `--ink-primary`
  against body text is 2.34 and 2.19, under G183's 3:1); `a:hover` also underlines `a.btn`; no
  `:focus-visible` sibling (W9.2). Use DaisyUI's `.link` (underlined at rest) with `text-ink-primary`,
  and scope the hover rule to `.link:is(:hover,:focus-visible)` or `a:not(.btn)`.
- **W5.** `web/scripts/gates/g1.ts:25`: the status-text regex blocks `text-primary-content` and
  `text-neutral-content`, DaisyUI's contrast-safe pairings Task 9 needs. End the pattern with
  `(?![\w-])`.
- **W6.** `components.css:131-135`: `.checkbox`'s edge is the 20% mix (about 1.49:1). Add `.checkbox`,
  `.radio`, `.toggle`, and `.file-input` to the 2px base-content keyline rule.
- **W7.** `components.css:196-200`: reduced motion covers only the press; W9.3's global kill switch
  (unlayered `!important` in W5.22's file, with a rendered row) doesn't exist, and Task 9's swap
  rotation would run under it. Land it before Task 9.
- **W8.** G9 has no rows for keylines, the ring against neutral or fills, or the selection ground;
  `--keyline-base` `#706a5f` on label-night base-300 is 2.50 (under 3:1). The focus row never runs
  the `dubplate-system` stamp. Add rows, or record the base-300 keyline as an exception.
- **Suggestions.** S1 `.btn-ghost { border-style: none }` breaks forced colors; use a transparent
  border and no press translate. S2 a `label` cookie stamp still carries `color-scheme: light dark`
  (W5.10); emit per variant, with a VI erratum. S3 the header comment in `components.css:4` is stale.
  S4 record the `@layer utilities` placement as a W5.8 erratum. S5 `0.6875rem` for the badge size. S6
  a `text-ink-success` alias. S7 mono utilities and header-link hover are deferred to Task 9 or rung 7.
  S8 allow only 3- or 6-digit hex for `theme-color`. S9 stub copy ("Log in", straight apostrophe) goes
  to Task 8.

Checked correct: the ring overrides DaisyUI's bare-control outlines; `main:focus-visible` follows
W3.6; the inks and fill-only markers; theme-color per variant; font URLs under the prefix; 45px
button height.
