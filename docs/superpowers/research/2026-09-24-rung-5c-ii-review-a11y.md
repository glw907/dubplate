# Rung 5c-ii pass-end review: DaisyUI and accessibility

Reviewer: `daisyui-a11y-reviewer` (Opus 5.5), over `0c2745f..2424b9a` in `web/`, against the
visual identity spec, W2.6 and W2.7, DL R2.6, R4.6, and G9, and plan revision 2.4. It ran g2, g9,
and theme-bundle on a scratch copy with three mutation probes; `svelte-check` reports 0 errors.
Every contrast figure in the spec and the markers recomputes correctly (label primary 4.61,
secondary and warning 1.83, `--ink-warning` 6.02/5.30/4.46; label-night primary 3.18, neutral
1.09, `--ink-primary` 6.34 and 4.85; every `-content` pair 4.96 or better; base-content 10.61 or
better). Verdict: accept with warnings. No blockers.

## Warnings

1. **theme-bundle does not compute the cascade** (`theme-bundle.ts:84-93`). An unlayered
   `:root{--color-primary:#0000ff;--ink-warning:#fff}` appended to `app.css` passes g2, g9, and
   theme-bundle while overriding every theme in a browser (mutation M2, proven). Fail on any rule
   setting a theme name without a `[data-theme=dubplate-*]` compound or outside `@layer base`, or
   soften the header until 5c-iii's rendered G7 half.
2. **theme-bundle never checks the composed themes' inks** (`theme-bundle.ts:187-193`). Deleting
   the `inks.css` import passes all three gates (mutation M1, proven). Add `INK_NAMES` and
   `reference.inks.get(composed)`.
3. **G9 never checks a marker's recorded ratio or its 3:1 tier** (`g9.ts:158-168`). label-night
   primary at 1.33:1 with its marker kept passes (mutation M3, proven). Fail on a ratio mismatch
   and below 3.0 when the reason cites that tier.
4. **G9 measures against base-100 only** (`g9.ts:205-242`, `inks.css:14`). label `--ink-primary`
   is 4.06 on base-200 and 3.42 on base-300; direct-text colors below 4.5 on base-300 (label
   accent/success 3.96, info 4.17, error 4.40; night accent/success 4.16, error 3.80);
   label-night primary as a fill is 2.94 on base-200 and 2.43 on base-300, under its marker's
   3:1. Add per-surface G9 rows, or record the restrictions in the spec and the 5c-iii handoff.
5. **Fill-only colors become invisible focus rings.** DaisyUI 5.7.46's button outline is
   `var(--btn-color, ...)`: label `btn-secondary` and `btn-warning` 1.83, label-night
   `btn-neutral` 1.09, label-night `btn-primary` on base-200 2.94 (WCAG 1.4.11, 2.4.7). A 5c-iii
   risk: forbid those modifiers or set the focus outline to base-content; markers should say
   "never text, never a focus or selector indicator".
6. **Inks do not follow DaisyUI's `theme-controller` branch** (`inks.css:13,18`). A swap using
   `.theme-controller` flips the palette but not the inks (`#765000` on night base 2.43,
   `#c1352b` 3.18). DL R4.6 bans the pattern; gate it and put the ban in the 5c-iii toggle brief.
7. **No `<title>` on any route** (WCAG 2.4.2, Level A). Add a unique `<svelte:head><title>`.
8. **Wrong specificity claim** (`system.css:13-14`, plan line 1321): the built DaisyUI blocks are
   (0,4,1), not (0,1,0).

## Suggestions

S1 `app.css:9-12` misstates import ordering (`var()` resolves at computed-value time); S2 G9
skips missing names silently; S3 G9 drops `oklch()` alpha; S4 the `dark:` scan skips `.css`; S5
descendant selectors count as root rules in `selectorHasDataTheme` (`g2.ts:324-328`); S6 G2 reads
only one `@layer base` block; S7 measure the noise texture in 5c-iii's rendered G10; S8 a
`<noscript>` line; S9 an unstamped document renders stock light (kept on purpose); S10 stamp the
color-scheme meta per palette if 5c-iii inlines critical CSS.

## Confirmed correct

`lang="en"`; zoomable viewport; one `<main>` and one `<h1>` per stub; no v3 directives; DaisyUI
configured through `@plugin`; no stock `prefers-color-scheme` block; the system branches'
parity; 45px buttons and 24px selectors (WCAG 2.5.8).
