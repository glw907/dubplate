# Visual identity draft tokens for 5c-ii Task 4

Draft token values from Geoff's visual identity brainstorm (session dubplate-04,
2026-09-24), sent to the 5c-ii conductor ahead of the spec
`docs/specs/2026-09-24-visual-identity-design.md`. The spec is the source of
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

## Toggle and default binding (second message from the brainstorm session)

- **The toggle** lives in the app header's nav, right-aligned after the last nav link, as an
  icon-only DaisyUI `swap` (sun and moon) with the accessible name "Dark theme", on every
  signed-in page. The login page has no toggle: it follows the system preference or the
  member's stored cookie. No design reason ties the toggle to 5c-ii.
- **Three states in effect**: follow system (the default), light, and dark. The UI shows two,
  and choosing either one writes the cookie. Returning to "follow system" can wait for a
  later settings page.
- **Head metas**: `<meta name="color-scheme" content="light dark">`, so browser-drawn UI
  matches before the theme applies. A media pair of theme colors,
  `<meta name="theme-color" media="(prefers-color-scheme: light)" content="#1f1b16">` and
  `<meta name="theme-color" media="(prefers-color-scheme: dark)" content="#110f0d">`, so
  mobile browser chrome matches the header bar (neutral) in each palette. When the cookie
  overrides the system choice, the server stamps a single `theme-color` that matches the
  stamped palette, in place of the media pair.
- **No flash**: the stamped `data-theme` and the body background agree on first paint in
  both palettes.
- **The mark** (a record in the label's red and black) is still being tuned. The favicon is
  an SVG with a 32px PNG fallback, sent when final. 5c-ii can ship without one.

## Settled after the 5c-ii amendment (third message from the brainstorm session)

- `--size-field: 0.28125rem`, which replaces the draft's `.25rem` and the amendment's interim
  `0.275rem`. Buttons come out at 45px, clearing DL R3.7, on a step DaisyUI's theme guidance
  sanctions for larger fields. `--size-selector` stays `0.25rem`.
- `label-night`'s `--color-neutral` is fill-only: it takes the R2.6 exception marker and no
  text ink. Neutral is the header bar and the offset-shadow color, never text.
- `label`'s `--color-secondary` (1.83:1) is fill-only the same way, with the R2.6 marker.
- Warning keeps `--ink-warning`, because warning copy appears as text.
- `DUBPLATE_THEME_PALETTE=system` is the adopter-facing spelling; the spec uses it.
- 5c-iii's component CSS may grow from a component sheet under review with Geoff: printed
  shadows on buttons and cards, ink borders, a red focus ring, mono uppercase badges,
  underline tabs, a thick-thin rule under headings, a mono catalog line, and dotted-leader
  tracklists. Exact rules follow if he approves.

## Superseded by the visual identity spec (2026-09-24, `ecba52a`)

The spec `docs/specs/2026-09-24-visual-identity-design.md` has landed and
supersedes this file. Where the two differ, the spec wins. One value changed for 5c-ii
Task 4: `--radius-selector` is `.25rem`, not `2rem`, because Geoff ruled squared badges, so
every corner is square. The rest of Task 4's values stand as settled above:
`--size-field: .28125rem`; `label` secondary and `label-night` neutral fill-only with the
W2.6 marker; `--ink-warning` #765000 (label) and `--ink-primary` #ec7b6a (label-night); the
`dubplate-system` default, the metas, and `DUBPLATE_THEME_PALETTE=system`.

## Spec revision 1.1 (2026-09-24): the inks shift

The semantic `--color-*` values are fills only; text in a semantic color uses its ink, and
every ink passes 4.5:1 on base-100, base-200, and base-300. Two values 5c-ii already shipped
change for `label`: `--ink-primary` `#9d2f26` (was `#c1352b`) and `--ink-warning` `#754f00`
(was `#765000`); 5c-ii's pass-end fold applies them. 5c-iii adds the remaining inks
(`--ink-accent`, `--ink-info`, `--ink-error` in both palettes; success uses accent), one focus
outline everywhere (2px `--ink-primary` at a 2px offset), and G9 rows on all three surfaces. The
toggle writes the cookie and the server stamps `data-theme`; never a `.theme-controller`. The
spec's Themes section is the source.
