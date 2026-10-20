# Ledger: typography and rhythm (felt-refinement audit, one lens)

Benchmark: `specimen.html` + `specimen.css` (rev 1). Measurements come from computed styles in
headless Chromium at 1440px and 390px (`bench/measure.cjs`, `measure-1440.json`,
`measure-390.json`), plus 3x crops (`bench/z-*.png`).

Distinct rendered sizes (label board, text-bearing elements, marker SVG excluded): 13.
Work Sans 12/14/16/18 (4, clean Tailwind steps). Alfa Slab One 20/24/30/36 (4, about a 1.2 ratio).
IBM Plex Mono 10.5/10.88/12/12.8/13.6 (5 sizes for 3 roles, the one ad-hoc family).

Recommendations first, ordered by felt impact; then everything judged already right.

| knob | verdict | evidence | exact change | felt impact |
|---|---|---|---|---|
| Film-edge caption overflow on phones | recommend | At 390px the dancers card's bottom edge has scrollWidth 311 against clientWidth 282; `white-space:nowrap; overflow:hidden` clips "PUBLIC DOMAIN" to "PUBLIC DON" (tile-390-1). A clipped credit is a clipped CC BY attribution, not only a looks problem. | `.filmedge .edge { flex-wrap: wrap; white-space: normal; column-gap: 12px; row-gap: 0; }` `.filmedge .edge > span { white-space: nowrap; }` `.filmedge .edge > span:last-child { margin-left: auto; }` (wraps the right-hand credit onto a second edge line, still right-aligned, only when it cannot fit) | high |
| Tracklist heading register | recommend | `<h2 class="font-semibold">` renders Work Sans 600 at 14px (card-body sets `.875rem`), the same size as the track titles under it and smaller than the 18px artist line. The spec sets h2 in Alfa Slab One; this is the one h2 that is not, and it reads as a stock DaisyUI card title. | `<h2 class="display text-xl">Tracklist</h2>` (the same 20px step as "Recently added"; no rule, matching the other in-card h2s). Hierarchy becomes h1 36 > h2 20 Alfa > artist 18 Work Sans > body 14. | med |
| Display line-height (and so the rule-under-title gap) | recommend | `.display` inherits Tailwind's per-size line-heights: 36/40 (Library), 36/45 (album and login h1, `leading-tight` = 1.25, looser than the default), 20/28, 24/32, 30/36. The label rule sits `mt-1` below the line box, so the glyph-to-rule gap varies (about 4px under "Library", about 6.5px under "Heart of the Congos", 4px of half-leading under the 20px h2). The two-line login h1 is also looser than a heavy slab wants. | `.display { line-height: 1.1; }` and drop `leading-tight` from the three h1s. Every Alfa size then carries the same proportional half-leading, so the 3px+1px rule lands at one optical distance under every title. | med |
| Curly quotes and apostrophes | recommend | Straight marks in the product copy: `Lee "Scratch" Perry`, `isn't`, `doesn't`, `We'll` (x3), `Garvey's` (x3 plus the marker label SVG), `Can't`. The dash in `1938–48` is already a true en dash; the `·` separators are correct. | `Lee “Scratch” Perry`; `isn’t`, `doesn’t`, `We’ll`, `Garvey’s`, `Can’t`. Rule for rung 7: copy strings and tag-derived titles render with U+2019/U+201C/U+201D (a smartypants pass on server-side copy; album titles from tags are shown as stored). | med |
| Mono size ladder | recommend | Five Plex Mono sizes: film edge 10.5px (set in px), badge .68rem = 10.88px, catalog .75rem = 12px, track number .8rem = 12.8px, duration .85rem = 13.6px. Two of them sit in one tracklist row beside a 14px title, three sizes in one line. | `.badge { font-size: .6875rem; }` `.filmedge .edge { font-size: .6875rem; }` `.tl .n, .tl .t { font-size: .8125rem; }` (catalog stays .75rem). Five sizes become three (11, 12, 13); 13px Plex Mono matches 14px Work Sans optically because Plex's cap height is about 5% taller. | med |
| Headings wrap balance | recommend | Unbalanced breaks: login h1 "The Hollies Road / library" (1440), album h1 "Heart of the / Congos" (390). | `.display { text-wrap: balance; }` (gives "The Hollies / Road library", "Heart of / the Congos"). | med |
| Widows in short prose | recommend | At 390px: "…for everyone in the / group.", "…or it has / moved.", "Request a / new one." (tile-390-2). | `p, .alert > span { text-wrap: pretty; }` | low-med |
| Catalog line break on phones | recommend | At 390px the line breaks to "… 11 TRACKS ·" / "1977", orphaning the year and splitting nothing useful. | Non-breaking spaces inside each segment (`FLAC&nbsp;16/44.1`, `11&nbsp;tracks`) plus `.catno { text-wrap: balance; }` so it breaks as "DBP-0412 · FLAC 16/44.1 ·" / "11 TRACKS · 1977". | low-med |
| Mono caps tracking system | recommend | Computed: badge .06em at 10.88px = 0.65px; catalog .08em at 12px = 0.96px; film edge .14em at 10.5px = 1.47px. The badge is smaller than the catalog line yet tracked tighter, the inverse of how small caps should track. | `.badge { letter-spacing: .08em; }` so badge and catalog share one "label caps" value. The film edge keeps .14em as its own deliberate register (wide-tracked edge print), giving a two-value system: .08em label caps, .14em edge print. | low |
| Track number column alignment | recommend | `.tl .n` is left-aligned in a 1.8rem box, so "9" and "10" align on their first digit rather than their units. Durations are mono and flush right, so they already column. | `.tl .n { width: auto; min-width: 2ch; text-align: right; }` (keep the `.5rem` gap). | low |
| Mark beside wordmark | already-right | 3x crop: mark centre at y≈139.3, lowercase x-height centre of "dubplate" at y≈139.7 (header). The mark aligns to the x-height's optical centre, not the ascender box. The 34px mark against 24px type and 36px against 30px on login keep the same relationship. | none (phone header: the mark span shrinks to 0 width; that is a flex-shrink layout defect for the layout lens, `flex: none` on the mark span) | n-a |
| Dotted leader vertical position | already-right | Leader dots sit about 4-6px above the baseline, level with the colon in "4:02" and the `·` separators in the catalog line. It reads as a middot leader, which agrees with the separator system; moving it to the baseline would make it the only baseline dot on the page. | none | n-a |
| Tabular figures | already-right / n-a | Every columned digit (durations, track numbers, catalog number, badge 16/44.1) is Plex Mono, so `tabular-nums` is inherent. No columned figures in Work Sans. | none | n-a |
| Alfa Slab placement and step | already-right | Alfa appears only on h1 (36), h2 (20 and 24), and the wordmark (24 header, 30 login). The 20/24/30/36 ladder is about a 1.2 ratio. h2 at 20 in a 270px aside and 24 on the full-width state cards reads as one family; the state card headings are their page's lead heading. | none | n-a |
| Display letter-spacing | already-right | `.display` .01em = 0.36px at 36px; Alfa is set tight, so a hair of positive tracking keeps the counters open. | none | n-a |
| Body line-height | already-right | 14/20 (1.43), 14/21 in cards (1.5), 16/24, 18/28, 12/16 on the three-line list meta, 14/17.5 on the two-line album captions. All sit inside 1.25–1.5, tighter for larger and shorter text. | none | n-a |
| Measure | already-right | Longest wrapping paragraph measures 65ch at 1440 ("Upload the first one…", 494px). The album description container allows 85ch but carries one short line; the phone paragraphs run 30–42ch. | none today; if rung 7 adds album notes, cap them at `max-width: 65ch`. | n-a |
| Heading margins | already-right | Title → rule 4px, rule → catalog 8px, header block → tabs 20px, tabs → grid 24px; aside h2 → rule 4px → list 16px. Tight above the rule, looser below, as the device wants. | none (the display line-height change above makes the 4px consistent) | n-a |
| Vertical rhythm | already-right | Every gap measured sits on a 4px grid (4, 8, 12, 16, 20, 24, 28). No off-grid values; the 20 and 28 steps are purposeful groupings, not drift. | none | n-a |
| Line-weight system | already-right | 1px (rule's thin line), 2px (control and card borders, tab strip, leader, badge border, focus outline), 3px (rule's heavy line, card shadow offset), 4px (active tab bar stacked on the 2px strip). A 1-2-3-4 scale with each step owning a role; the active tab's 4px red over the 2px black echoes the label rule. | none | n-a |
| Badge beside text | already-right | Badge centred on the list row (`items-center`); its cap height centres in the 2px box at 3x. | none | n-a |

## AI-default and template check

Nothing in the type layer reads as a template, with two local exceptions that the recommendations
above remove: the Tracklist heading is a stock DaisyUI card title, and the straight quotes read as
unedited default copy. Against the named clusters:

- Cream, serif, and terracotta: the surface overlaps it (cream ground, a slab display face, a
  brick red). What pulls it out is that the red is a record-label red paired with mustard and a
  Rasta green, the slab is a heavy wood-type Alfa rather than an editorial serif, and the mono
  catalog data and record devices give it a subject. It sits near the cluster and is not of it.
- Broadsheet (hairline rules, zero radius): the 3px+1px rule and 2px borders are printed-label
  weights, not hairlines, and the radius is .25rem. Not this cluster.
- Uniform rounded cards and gradient heroes: no gradients. The empty, error, and form cards share
  one bordered and shadowed card, which is a composition question for another lens.
- A cluster the brief does not name is the neo-brutalist SaaS look (thick black borders plus hard
  offset button shadows). The spec's reduced-strength, warm-ink shadows keep dubplate on the
  printed side of it; this is ruled and stays as is.
