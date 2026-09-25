# Ledger: color, surface, depth, and cohesion lens

Benchmark: specimen.html + specimen.css (rev 1), tiles at 1440/768/390, plus probes
`bench/probe.html`, `bench/probe2.html` (renders `probe.png`, `probe2.png`, `zz.png`).
Contrast figures are WCAG ratios computed from the spec's hex values (`bench/cr.py`).

## Measurements

- Shadow values in the CSS: 5. Outer: `2px 2px 0` (buttons, record drop-shadow) and
  `3px 3px 0` (cards, alerts, sleeves). Inset: input `2px 2px 0` at 10% (top-left inner shade,
  same light source), tab bar `0 -4px 0`. One light source (top-left) throughout.
- Shadow legibility: light button shadow 3.66:1 on base-100, card shadow 2.76:1. Dark button
  shadow 1.14:1, card shadow 1.13:1 (black at any alpha cannot beat 1.2:1 on `#1c1916`).
- Radius values in use: 4. `.25rem` (tokens, `rounded-box`, `rounded-sm`, which is `.25rem` in
  Tailwind 4), `.2rem` (badge override), `2px` (sleeve jacket), `0` (tabs, film edge).
- Dark keylines: input 13:1 (full cream), card 2.81:1 (cream 35%, translucent), default button
  1.20:1 (`#000`), default button fill against ground 1.08:1.
- Reds at rest per view: library 3 (mark label, active tab bar, primary button; record labels
  sit hidden behind sleeves), album page 2 (mark, Play), login 2, rising to 4 with an expired
  link (mark, focus ring, error alert, primary button). Error vs primary: 1.29:1 apart (light),
  1.56:1 (dark); dark error fill `#e0604f` has luminance 0.248 against primary's 0.141, so the
  error block outshouts the action.
- Film-edge caption: `#e8c25a` at .85 on `#0d0c0a` = effective `#c7a74e`, 8.4:1. Palette
  secondary on the same carrier: 8.9:1 (`#e0a526`), 9.6:1 (`#e5ad33`).
- No-art album: the handwritten label is hidden at rest AND on hover (probe: after the 14% slide
  and 40deg turn the label is still under the jacket), so the one use of the marker face is never
  seen. The title at `font-size="1.55"` also overflows the label's chord (renders as "ARVEY'S GHOS").

## Ledger

| knob | verdict | evidence | exact change | felt impact |
|---|---|---|---|---|
| No-art album: hidden label | recommend | Label never visible, rest or hover (probe.png sleeve 2); blank tan card reads as a missing image. A die-cut sleeve is how plain sleeves and test pressings actually ship. | `.sl:has(.nolabel) { filter: drop-shadow(3px 3px 0 var(--shadow-card)); } .sl:has(.nolabel) .rec svg { filter: none; } .jacket.nolabel { box-shadow: none; -webkit-mask: radial-gradient(circle at 71% 50%, #0000 18.5%, #000 18.8%); mask: radial-gradient(circle at 71% 50%, #0000 18.5%, #000 18.8%); }` The hole centre is the label centre (`.rec` left 24% + 47%, top 3% + 47%); radius 18.5% just clears the 17.0% label so a sliver of run-out shows. The shadow moves to the wrapper because a mask clips the element's own box-shadow and filter. Phone top-peek variant: recompute the centre from that layout's `.rec` offsets. Verified in probe2.png / zz.png, both palettes. | high |
| No-art album: title fit | recommend | `font-size="1.55"` overflows the 5.8-unit label at y=14.6 (chord 11.3 units). | In the record SVG `<text>`: `font-size="1.2"`, add `textLength="9" lengthAdjust="spacingAndGlyphs"` only when the measured title exceeds 9 units; titles that would compress past ~80% wrap to two lines at `y="13.0"` and `y="14.5"`, clear of the spindle. | high (part of the above) |
| Dark keyline for base-surface objects | recommend | Default button in dark: fill 1.08:1, `#000` border 1.20:1, shadow 1.14:1; the button's edge nearly vanishes while cards beside it carry a 2.81:1 cream edge. Two rules for one kind of object. | `[data-theme="label-night"] .btn:not(.btn-primary):not(.btn-ghost), [data-theme="label-night"] .card { border-color: color-mix(in oklab, var(--color-base-content) 40%, var(--color-base-100)); }` (= `#706a5f`, 3.26:1 on base-100, 3.02:1 on base-200; solid, so it no longer blends with the fill beneath). Inputs keep full base-content, as in light. Primary buttons keep `#000`. | med |
| Dark keyline for colour-filled objects | recommend | Alerts in dark take the card's translucent 35% cream border; over a gold fill it paints a lighter gold ring (`#e8c069`), over coral a pale ring. Primary button already uses `#000`, which reads as a printed keyline. | `[data-theme="label-night"] .alert { border-color: #000; }` Rule: in dark, a colour fill takes the black keyline; a base surface takes the 40% cream keyline. | med |
| Error alert register (ink discipline) | recommend | Error fill sits 1.29:1 (light) / 1.56:1 (dark) from primary: a near-miss red beside the action it asks the member to retry. In dark it is the loudest mass on screen. Badges already print status as an outline. | `.alert-error { background: color-mix(in oklab, var(--color-error) 14%, var(--color-base-100)); color: var(--color-base-content); border-color: var(--color-error); }` (text 11.8:1 light on `#ebd2bd`, 11.7:1 dark on `#37231e`; keyline 5.9:1 / 5.0:1). Warning alert keeps its solid gold fill; gold does not compete with the action. Red fill is then reserved for the primary action. | med |
| "Recently added" list frame | recommend | `list bg-base-100 rounded-box border-2 border-base-300` with no shadow is the stock DaisyUI list; it is the only framed object that follows neither the object rule (ink keyline + 3px shadow) nor the seam rule (base-300 divides regions, never frames objects). It sits directly above an inked, raised alert. | `.list.rounded-box { border: 2px solid var(--color-base-content); box-shadow: 3px 3px 0 var(--shadow-card); }` (dark: covered by the 40% keyline rule above; add `.list.rounded-box` to that selector). Drop `border-base-300` from the markup. | med |
| Film-edge caption ink | recommend | `#e8c25a` is a paler, less saturated gold outside the palette; next to a gold warning alert or badge it reads as a near-miss. The carrier `#0d0c0a` is a third near-black beside `#0b0a09` (vinyl) and `#110f0d` (neutral dark). | `.filmedge { background: #0b0a09; } .filmedge .edge { color: color-mix(in oklab, var(--color-secondary) 85%, #0b0a09); opacity: 1; }` (≈8.4:1, same brightness as today, palette hue). Film and vinyl then share one black. | low-med |
| `::selection` | recommend | Unstyled: the browser's blue is the only cool, off-palette colour a member can summon. Red would add a competing red. | `::selection { background: color-mix(in oklab, var(--color-secondary) 45%, var(--color-base-100)); color: var(--color-base-content); }` (light `#ebcb85`, text 10.9:1; dark `#765c23`, cream text 5.0:1): a grease-pencil highlight. | low |
| Radius: badge | recommend | `.2rem` is a fourth radius, 0.8px off the `.25rem` token; drift, not intent. | `.badge { border-radius: var(--radius-selector); }` | low |
| Radius: cover thumbnails | recommend | Thumbnail is a sleeve at 44px, but it takes the chrome radius (`.25rem` = 9% of its side) while the 190px sleeve takes `2px` (1%). At list size it reads as an app icon. | Thumbnails and the no-art thumbnail: `rounded-[2px]` in place of `rounded-sm`. Two radius families remain: 2px for printed covers, `.25rem` for chrome, 0 for rules and film. | low |
| Sleeve radius 2px vs spec `.25rem` | already-right | The spec's "every corner" line governs the DaisyUI match block (controls). A sleeve is card stock, crisper than a control; 2px vs 4px is the one deliberate print/chrome split. | none (the thumbnail change above aligns the only stray). | n-a |
| Depth: light source and offset family | already-right | Every outer shadow falls down-right with zero blur; the input inset shades top-left; the pressed button collapses into its own 2px. Sleeve, card, alert share 3px; button and record share 2px (the record is thinner and sits under the sleeve). | none | n-a |
| Depth in dark | already-right | Shadows measure 1.13:1: they read as a faint thickening of the bottom-right edge, not as a shadow. No black can do better on `#1c1916`, and the owner asked for less present shadows; the keyline carries the edge (after the 40% change). | none | n-a |
| Record drop-shadow alpha (55% on a 2px offset) | already-right | Next to a near-black disc the alpha difference from 45% is invisible. | none | n-a |
| Sleeves carry no keyline | already-right | Cover art has its own edge; an ink frame round a cover would read as a UI card, not a sleeve. The shadow plus the lip gradient define it, including white covers. | none | n-a |
| Border strategy (light) | already-right | Two meanings, used consistently: base-content 2px = an object or a typographic rule (controls, cards, tabs strip, label rule); base-300 2px = a seam between page regions (aside edge, section dividers, 1.35:1). | none (the list is the one stray; see above) | n-a |
| Surface ladder | already-right | base-100 = paper (page, cards, list); base-200 = panel and back-of-sleeve (aside, tracklist card, default button fill); base-300 = seams and blank stock (dividers, no-art sleeve). Each step holds one meaning per view. | none | n-a |
| Red count at rest | already-right | 3 on the library, 2 on the album page and login; record labels stay behind the sleeve until hover. The only overflow is the error-alert state, handled above. | none | n-a |
| Grain | already-right | One language, three materials, coarse to fine with scale: film (freq .7, alpha .34), sleeve board (.9, .2, warm gray), control paper (DaisyUI `--fx-noise`, 1.34, .2; visible only as a fine speckle at 4x zoom). All fractal noise at low alpha, no blend modes. | none | n-a |
| Dark devices: record rim, sleeve lip, label rule, leaders, tab rule | already-right | Rim at 55% cream keeps each disc visible on `#1c1916`; lip gradient still reads on dark covers; the 3+1 rule and tab strip stay full cream; leaders at 35% currentColor stay legible and quiet. | none | n-a |
| Image frames (sleeves, thumbnails, photos, film edge) | already-right (bar the thumbnail radius) | Covers print as sleeves; photos always sit in the film-edge carrier; a film edge inside a card takes the card's clipped `.25rem` corner, the full-bleed login edge is square to the viewport. | none beyond the thumbnail radius | n-a |
| No-art thumbnail in lists | already-right | A die-cut dot at 44px reads as a bullseye (probe.png); the spec rules the record illegible at row size, and the row text names the album. | none (take the 2px radius only) | n-a |

## AI-default and template check

Mostly off-template. The palette sits near the cream, serif, and terracotta cluster, but four
things move it off: the slab display face (not an editorial serif), the 2px ink keylines with
hard offset shadows, the black header bar, and the record, catalog-line, and film-edge devices.
The 3+1 label rule is heavier than a broadsheet hairline and sits with `.25rem` corners and
printed shadows, so the broadsheet cluster does not apply. Cards are inked and squared, not
uniform soft-rounded tiles.

Two elements read as template today. The "Recently added" list is the stock DaisyUI list
(`rounded-box`, faint base-300 border, no shadow); the smallest move off it is the card rule
above. The blank tan no-art sleeve reads as a generic image placeholder; the die-cut move above
fixes it.
