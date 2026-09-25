# Visual identity draft tokens for 5c-ii Task 4

Draft token values from Geoff's visual identity brainstorm (session dubplate-04,
2026-09-24), sent to the 5c-ii conductor ahead of the spec
`docs/superpowers/specs/2026-09-24-visual-identity-design.md`. The spec is the source of
truth once it lands; these values let Task 4 land replaceable theme data now. Geoff ruled
the look in that session: a light `label` theme and a dark `label-night` theme (cream,
black, and red), Alfa Slab One for display, Work Sans for body, IBM Plex Mono for captions,
a hard ink offset shadow as the depth rule, and plain copy with no record puns.

## Model

The two-axis model stays: the match block `dubplate` and the palettes `label` and
`label-night`, composed per DL R3.6 as `dubplate-label` and `dubplate-label-night`. Task 4's
config keys do not change.

## Match block `dubplate`

```css
--radius-selector: 2rem; --radius-field: .25rem; --radius-box: .25rem;
--size-selector: .25rem; --size-field: .25rem; --border: 2px; --depth: 0; --noise: 1;
```

## Palette `label` (color-scheme: light)

| Token | Fill | Content |
| --- | --- | --- |
| base-100 / 200 / 300 | #f4ead3 / #eadcbc / #dccaa3 | base-content #1f1b16 |
| primary | #c1352b | #fff6ea |
| secondary | #e0a526 | #1f1b16 |
| accent | #2f6b3a | #f4ead3 |
| neutral | #1f1b16 | #f4ead3 |
| info | #2e5f8a | #ffffff |
| success | #2f6b3a | #ffffff |
| warning | #e0a526 | #1f1b16 |
| error | #a8261d | #ffffff |

## Palette `label-night` (color-scheme: dark)

| Token | Fill | Content |
| --- | --- | --- |
| base-100 / 200 / 300 | #1c1916 / #24201c / #342e28 | base-content #efe4cc |
| primary | #c1352b | #fff6ea |
| secondary | #e5ad33 | #1c1916 |
| accent | #5a9e66 | #1c1916 |
| neutral | #110f0d | #efe4cc |
| info | #6d9cc7 | #1c1916 |
| success | #5a9e66 | #1c1916 |
| warning | #e5ad33 | #1c1916 |
| error | #e0604f | #1c1916 |

## Contrast, as measured by the brainstorm session

Every fill and content pair passes 4.5:1 in both palettes. base-content passes on every base
at 10.6:1 or higher. Two colors fail as text on the page ground, though both pass as fills,
so each gets an ink value (W2.x):

- `label`: warning as text is 1.83:1. Warning ink is #765000 (6.02 on base-100, 5.30 on
  base-200, 4.46 on base-300, so never text on base-300). Secondary on light is the same
  mustard and is fill-only.
- `label-night`: primary as text is 3.18:1. Primary ink is #ec7b6a (6.34, 5.86, 4.85 on
  base-100, 200, 300). Links and red text on the dark theme use the ink, never
  `--color-primary`.

Every other semantic color passes 4.5:1 as text on base-100 in both palettes.

## Default selection

System preference picks the palette: light selects `dubplate-label`, dark selects
`dubplate-label-night`. A member's manual toggle overrides it through W5.10's cookie. With
no preference signal, the default is `dubplate-label`.

## Not DaisyUI variables (5c-iii component CSS)

- The hard ink offset shadow: `box-shadow: 4px 4px 0 var(--color-neutral)`, 3px on the
  record. It picks the right near-black per palette with no extra palette name (R2.1).
- Paper grain on album sleeves: an SVG-noise overlay at about 20% opacity, separate from
  `--noise: 1`.
- The cream record rim.

## Fonts (5c-iii)

Alfa Slab One (display only), Work Sans (body, variable), and IBM Plex Mono (captions and
metadata), all OFL, embedded as woff2 through go:embed.
