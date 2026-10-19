# 5c-iii web review (svelte-reviewer)

Range `6e6fd12..HEAD` over `web`, graded against the visual identity spec revision 1.3 and the plan.
Cascade claims checked against DaisyUI 5.7.46's component CSS and the built bundle.

- **B1 (blocker, converges with a11y B1).** `components.css:139-150`: `--badge-bg: transparent` leaves
  `color: var(--badge-fg)`; text vanishes (label: error/info/success `#ffffff` on base-100 1.20:1,
  primary/accent/neutral cream on cream; label-night: six statuses 1.00:1). Fix: `--badge-ink` per
  status driving text and border, `background-image: none` (DaisyUI's noise), secondary and neutral
  fall back to base-content, and a rendered row measuring painted badge text at 4.5:1 per status per
  theme.
- **W1 (converges with a11y W3).** `components.css:73-77,110-121`: the `.btn` border and shadow rules
  beat DaisyUI's `:disabled`, `.btn-disabled`, `.btn-link`, and `.btn-outline.btn-primary` resets.
  Exclude or restate them; add a disabled-button row.
- **W2.** `components.css:130-135`: `.input { border-color }` suppresses DaisyUI's `validator`,
  `input-error`, and `input-success` borders; the validator hint paints raw `--color-error` as text.
  Restate the states after the base rule and set `.validator-hint { color: var(--ink-error) }`.
- **W3.** `components.css:48-201`: the rules share the `utilities` sublayer with Tailwind's generated
  utilities and come later, so `class="btn shadow-none"` and similar lose. Wrap selectors in
  `:where()` so markup utilities win by specificity while the rules still beat DaisyUI by layer.
- **W4.** `components.css:160-163`: `.display` sets no slab face (R-5ciii2, VI type table). Add
  `font-family: var(--font-serif); font-weight: 400`, optionally a `document.fonts.check` row.
- **W5.** `components.css:165-176`: the page-title rule frames the title (3px above, 1px below); VI
  and the plan put a 3px-over-1px rule under the title. Draw both lines in `::after`.
- **W6 (converges with a11y W5).** `g1.ts:25` refuses `text-*-content`; end with `(?![\w-])`.
- **Suggestions.** S1 the wrapped-input double ring (converges with a11y W2). S2 the global
  `a:hover` underline also hits `a.btn` (converges with a11y W4). S3 `.list` frame should scope to
  `.list.rounded-box` or carry the radius. S4 `link-primary` and `.btn-link` paint raw fills; map to
  inks or refuse in G1. S5 press feel (translate vs shadow transition; ghost buttons; a backwards
  comment at `run.ts:1298-1303`). S6 badge radius against the specimen. S7 label-night keyline on
  base-300 2.50:1 (converges with a11y W8). S8 `sourceVersion()` misses `web/scripts/`, which
  Tailwind scans. S9 gate precision: `g4.ts:128-134` checks only presence of `font-display`; G5's
  comment says "exactly one"; `root-absolute.ts` sink gaps (converges with security M1). S10 no
  rendered row proves the three faces load under the prefix. S11 stale VI line citations in
  `inks.css:9`, `g9.ts`, `surface.css:7-8` (cite section names); the stale `components.css` header;
  `FONTS.md`'s truncated commit. S12 stub copy "Log in" goes with Task 8. S13 the input inset shadow
  against the specimen.

Checked sound: fonts resolve under any prefix; descriptors; mark and favicon bytes; every ink and
surface token; font stacks; the focus rule's layer placement; the reduced-motion override; G9 and
G2's pinned rows; the Task 9 deferrals.
