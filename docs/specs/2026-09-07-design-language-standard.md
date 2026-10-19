# dubplate design-language standard

Status: revision 2, 2026-09-07. This is the Phase 0c deliverable of the
parallel-build approach
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md:56`), a
promotion of the theme-variable contract, the two-axis composition
model, and the font delivery mechanism out of rung 7 and into a standard
every web surface is measured against.

**Revision 2, 2026-09-07.** Revision 1 went through two independent
adversarial reviews, merged and folded here
(`docs/superpowers/research/2026-09-07-design-language-standard-review.md`).
Both lenses filed the same blocker: revision 1's CSP was adopted
verbatim from the pre-pivot Vite stack and blocks the inline scripts
SvelteKit generates, so rung 5's login page would have loaded and never
booted. R5.6 and G6 are rewritten around that, and R4.2 loses the CSP
argument it leaned on. Fourteen majors and eight minors fold besides.
The largest are the variable contract's arithmetic, which said 29 in one
clause and 32 in three others and now says 32 everywhere; the `data-theme`
value rung 5 has to stamp, which no clause named; DaisyUI's own
`theme-controller` selector, which outranks R4.1's document order; three
gates that could not be run as written; and 2.0's hook 5, which R4.2
lands and revision 1 called neutral. Nothing the reviews raised was
rejected. The review fold itself asked no new ruling of the conductor;
the rulings below answer the six questions revision 1 already carried.

**Rulings, conductor, 2026-09-07.** The six readings revision 1 left
provisional (R1.6, section 6's gate rejoin point, R3.2, R3.3, R4.2, and
R5.5) are ruled under the unattended-window grant; each clause below
carries its own ruled mark and wording. One owner note rides to the
rung 5 sitting: the served CSP's `style-src 'unsafe-inline'` relaxation.

The standard is normative. Each clause is a requirement or an explicit
non-requirement, with the source it derives from. A clause carrying
**Provisional, conductor ruling owed** is a recommendation applied to
unblock authoring; it holds until the conductor rules otherwise.

Sources, cited below by these short names:

| Short name | File |
|---|---|
| BS | `docs/superpowers/specs/2026-09-02-greenfield-server-build.md` |
| PA | `docs/superpowers/specs/2026-09-03-parallel-build-approach.md` |
| S20 | `docs/superpowers/specs/2026-09-07-streamer-2-0-design.md` |
| TB | `docs/superpowers/research/2026-09-03-theme-collection-brief.md` |
| CM | `CLAUDE.md` |

The DaisyUI variable names in section 2 were read from the published
`daisyui@5.7.17` package, not from documentation or memory:
`https://cdn.jsdelivr.net/npm/daisyui@5.7.17/themes.css` for the theme
blocks and `.../theme/index.js` for the plugin. All 35 built-in blocks
declare the same 29 names, measured. No `node_modules` tree exists in
this repo, so the URL is the citation a later reader re-verifies against.
The only DaisyUI pin the tree carries is `_salvage/web/package.json:21`,
daisyui 5.7.25, which the salvage adoption record declines as rung 7's
choice. The contract below is asserted for DaisyUI 5, and rung 7
re-reads it against the version it pins.

## 1. Status and scope

**R1.1 (binding surfaces).** This standard binds rung 5's login and
consumption pages, every surface rung 7 authors, rung 12's driver
pairing as section 3 and section 7 leave it, and every web surface built
after them. That includes 2.0's device page and share management, which
are pages in the same web app as rung 5's login (S20:303-306), and 2.1's
first-party web player (S20:433). Sources: BS:643 ("binding here and on
every later web surface"), PA:56.

**R1.2 (what rung 7 keeps).** Rung 7 owns the shipped theme roster, the
visual authoring of every surface including rung 5's login pages, and
the attended before/after. This document is not rung 7's plan and
decides none of those. Source: PA:56.

**R1.3 (contract, not classes).** The theme-variable contract is the
documented customization API. Class names and DOM structure are not, and
no clause here promises either. Source: BS:645-647.

**R1.4 (theming is data).** A theme is authored as declarations only. No
theme introduces a selector, a rule, a plugin, or code of any kind.
User-supplied components and template overrides stay ruled out. Sources:
BS:643-645, BS:650-652, TB:202-206.

**R1.5 (universality).** No clause may require a rebuild, a Node
runtime, or network access at the adopter's site. Sources: CM:7
("no site-specific value, path, or identity default outside
`internal/config`"), CM web rule (CM:133-143), BS:648-650.

**R1.6 (the login-page stack).** BS:600 describes rung 5's login and
consumption pages as Go-templated with no JavaScript. CM:142 says rung
5's login page is the first page built under the SvelteKit static rule,
adopted 2026-09-07 (BS:16-18). **Ruled, conductor 2026-09-07: SvelteKit
static, client-rendered, served by the Go binary.** The later CLAUDE.md
rule governs over BS:600, and the close of rung 5 records the build-spec
correction. Sections 4 and 5 hold up under either reading with one
exception: nothing they require runs JavaScript before first paint, so a
Go-templated page would render correctly, and R4.2's theme picker is the
one mechanism that would need a different implementation, since it
changes the attribute after hydration and stores the pre-rung-6 pick in
the browser.

## 2. The theme-variable contract

**R2.1 (the complete variable set).** A composed theme declares exactly
these 32 names, no more and no fewer, in three groups. Twenty-one belong
to the palette axis and eleven to the structural axis (section 3). The
first 29 are DaisyUI's own block; the three font names are Tailwind 4's,
and R2.7 explains why they sit on the structural axis.

Palette axis, 21 declarations:

```
color-scheme
--color-base-100      --color-base-200      --color-base-300
--color-base-content
--color-primary       --color-primary-content
--color-secondary     --color-secondary-content
--color-accent        --color-accent-content
--color-neutral       --color-neutral-content
--color-info          --color-info-content
--color-success       --color-success-content
--color-warning       --color-warning-content
--color-error         --color-error-content
```

Structural axis, 11 declarations. The first eight are DaisyUI's:

```
--radius-selector  --radius-field  --radius-box
--size-selector    --size-field
--border           --depth         --noise
--font-sans        --font-serif    --font-mono
```

Source: the published DaisyUI 5.7.17 `themes.css`, where 28 distinct
custom properties each appear exactly 35 times alongside 35
`color-scheme` declarations, one identical set per built-in block; the
same DaisyUI list appears at TB:120-126.

**R2.2 (block metadata).** A theme block also carries DaisyUI's
authoring metadata: `name`, and where each applies `default`,
`prefersdark`, `color-scheme`, and `root`. The plugin destructures those
five keys and spreads the remainder into the emitted block
(`daisyui@5.7.17` `theme/index.js`). `name`, `default`, `prefersdark`,
and `root` are authoring inputs that never reach the block, so they are
not theme variables and are not counted in R2.1. `color-scheme` is the
one that is both: destructured as metadata and re-emitted as a CSS
property, which is why it counts among R2.1's 21 declarations while not
being a custom property.

**R2.3 (semantic tokens only).** A component reads color, radius,
border, depth, and size through the DaisyUI semantic token or the
DaisyUI utility that resolves to one. A component never names a
palette value. This covers every CSS color function, every hex literal,
every named CSS color, and Tailwind's own color utilities, numbered
(`bg-slate-800`, `text-red-500`) and bare (`bg-white`, `text-black`)
alike. Sources: BS:645 ("no hardcoded colors"),
`_salvage/web/src/app.css:8-9` ("only DaisyUI's own theme tokens are set
here, never a DaisyUI internal class name").

**R2.4 (what "no hardcoded colors" means as a gate).** The check is a
script. A bare grep cannot express it, because the position a token
appears in decides whether it is a color. It scans `web/src` for the
extensions `.svelte`, `.ts`, `.js`, `.css`, and `.html`, exempting
`web/src/lib/themes/` alone, and fails on any of four patterns:

1. A color function call: `rgb(`, `rgba(`, `hsl(`, `hsla(`, `oklch(`,
   `oklab(`, `lab(`, `lch(`, `hwb(`, or `color(`.
2. A hex literal matching `#[0-9a-fA-F]{3,8}` inside a CSS declaration
   value or inside the text of a `class` or `style` attribute. A
   fragment href and an element id sit outside both, which is what
   "in a color position" means mechanically.
3. A Tailwind color utility on a color-bearing prefix (`bg-`, `text-`,
   `border-`, `fill-`, `stroke-`, `ring-`, `shadow-`, `divide-`,
   `outline-`, `decoration-`, `accent-`, `caret-`, `from-`, `via-`,
   `to-`), whether the suffix is a numbered family
   (`(slate|gray|zinc|neutral|stone|red|orange|amber|yellow|lime|green|emerald|teal|cyan|sky|blue|indigo|violet|purple|fuchsia|pink|rose)-(50|[1-9]00|950)`)
   or one of the bare four, `white`, `black`, `transparent`, `current`.
4. A CSS named color used as a color value. The check carries the CSS
   named-color list and matches it against declaration values whose
   property is color-bearing.

`web/src/lib/themes/` is the only exempt path, and it is exempt because
it is the layer that defines the values. An exemption anywhere else is a
gate failure, not a comment. Source: BS:645, CM gate rule ("a check
either gates or does not exist").

**R2.5 (completeness is enforced, never inherited).** A block missing a
declaration is a gate failure. See R3.5 for why silence is the wrong
failure mode.

**R2.6 (contrast floor).** Within a composed theme, three conditions
hold. Each `-content` value clears 4.5:1 against its own paired color.
`--color-base-content` clears 4.5:1 against `--color-base-100`. And
each of `--color-primary`, `--color-secondary`, `--color-accent`,
`--color-neutral`, `--color-error`, `--color-success`, `--color-warning`,
and `--color-info` clears 4.5:1 as text against `--color-base-100`.

The third condition exists because DaisyUI paints those colors as text
directly rather than routing through `-content`: the `-soft` modifiers
use the color itself as the text color
(`.alert-soft{color:var(--alert-color)}`, `.badge-soft` the same), and
the `text-*` utilities do the same for a link or a label. The mechanism
is not specific to the four state colors, so the floor is not either.

A pairing that cannot meet the floor records an exception in its own
file, in the fixed single-line form G9 parses, directly above the
declaration. Where the lower bar is deliberate, the
exception names WCAG's 3:1 UI-component and large-text tier as its
ground, the way the salvage tree already does for primary
(`_salvage/web/src/app.css:26-30`: "primary's white content clears
WCAG's >=3:1 UI-component/large-text tier, matching the ratified
mockup's exact sample"). Ratios are computed on sRGB relative luminance
per WCAG 2.2, which means an `oklch()` value is converted before it is
measured.

Tiers 2 and 3 sit outside this floor. An operator who overrides a
branding variable or ships a CSS file can break any pairing here, and
that risk is the operator's, as R4.4 already states for the file.
Source: `_salvage/web/src/app.css:32-47` (the U13 F8.1 finding, measured
at 2.2 to 2.9:1 before correction).

**R2.7 (how a theme declares its font pairing).** R2.1's three font
names are Tailwind 4's own font namespace, and DaisyUI's plugin spreads
any non-metadata key straight into the emitted block, so a theme block
overrides them the same way it sets a color. They belong to the
structural axis because typography is structure (BS:685-686). Each value
is a font stack whose last entry is a CSS generic family, so a glyph the
embedded face lacks still paints. The declaration names families only.
It never carries a URL, and it never carries an `@font-face` rule.
Sources: TB:288-292, R1.4.

**R2.8 (motion is not a theme axis).** Motion is global and honors
`prefers-reduced-motion`. No theme declares motion. Source: BS:683-684,
TB:294-296.

## 3. The two-axis composition model

The erratum retired streamer matching and left the axes standing. The
structural axis targets dubplate's own design only, the matched themes
are not authored, and the estate config is `match dubplate, palette
gruvbox`. Source: BS:5-18.

**R3.1 (the two axes).** Theming is two orthogonal picks. The MATCH axis
carries the structural layer: R2.1's eleven structural declarations. The
PALETTE axis carries R2.1's 21 palette declarations. The two sets are
disjoint and their union is the complete contract. Sources: BS:684-691,
TB:297-308.

**R3.2 (the match axis has one value).** `dubplate` is the only match
value that exists. Config accepts it and rejects any other value with an
error naming the accepted set. **Ruled, conductor 2026-09-07: keep the
key.** The axis stays in config rather than collapsing to a constant,
for config-schema stability: an adopter who authors a second structural
block, and a later reader of the config file, both meet an axis that is
one-valued today. A constant would have to be re-widened into a key
later. The reason revision 1 gave, that 2.0's built-in streamer arrives
as one more driver (BS:15-16), does not survive the erratum: that
sentence is about the driver seam, and section 8 of this document rules
that an in-process streamer has no theme to match.

**R3.3 (how config expresses it).** Two string-valued environment keys
under the `DUBPLATE_` namespace. Both carry defaults and neither is a
wall row, because a missing theme pick has a correct answer and a
missing library root does not. A defaulted key is a threshold, so both
register through the existing threshold registry
(`internal/config/thresholds.go:40-42`), which no lane edits
`internal/config/keys.go` to reach; rung 1b's wall-row and
reserved-class work (PA:54) is a different surface these keys never
touch. The registry today parses `KindInt` and `KindDuration` only
(`internal/config/thresholds.go:15-18`), so a string kind is owed before
these keys land. **Ruled, conductor 2026-09-07: the string threshold
kind is rung 1b's registry work, landing beside the `KindFloat` carry.**

**R3.4 (composition, and DaisyUI's built-ins).** The layers compose as
separate variable sets. Where a palette sourced from a DaisyUI built-in
carries structural declarations, the match layer's declarations win.
Whichever mechanism rung 7 chooses, that ordering is the requirement,
and a test proves it by reading a composed page's computed
`--radius-box` against the match layer's value while a built-in palette
is active.

Two measured facts narrow rung 7's choice without deciding it. DaisyUI's
plugin emits every block, built-in and custom alike, at
`[data-theme="<name>"]`, one block per value; and R1.4 forbids a theme
from introducing a selector of its own, so a match layer authored as
data cannot land on a second attribute. Together they point at
composition performed when the bundle is built, from two authored
sources, which is what keeps BS:689-690's "the axes multiply without
multiplying authored themes" true. Build-time composition happens on the
workstation and costs the adopter nothing, so it is consistent with
R1.5. How the built-ins reconcile into the two-axis model stays rung 7's
design detail, as the ruling left it. Source: BS:691-693, TB:305-308.

**R3.5 (a missing variable fails loudly).** A palette block missing a
palette declaration, or a match block missing a structural one, fails
the gate. It does not fall back. DaisyUI's plugin merges a custom block
with a built-in only when the block's `name` matches that built-in
(`daisyui@5.7.17` `theme/index.js`, the `allThemes[name]` branch); a
block under any other name simply emits fewer declarations, and the
cascade then serves whatever the compiled default theme set. Measured,
the default's block is the one selector group of 35 that carries a bare
`:root`, which is the mechanism that lets an incomplete block inherit
silently. The result is a repaint that is close enough to pass review
and wrong (TB:222-227), and a missing semantic token in particular
paints a rejected album the same gray as an accepted one
(`_salvage/web/src/app.css:13-15`). The composed set is authored data,
so completeness is checkable before it ships. Silence is the one failure
mode this contract cannot afford.

**R3.6 (the composed identity is one `data-theme` value).** A composed
theme has one name, spelled `<match>-<palette>`, and that name is the
whole of what R4.2 stamps and what a picker sets. The estate's is
`dubplate-gruvbox`. A single `data-theme` value selects a single DaisyUI
block, so the two axis picks reach the document as one composed
identity, in one attribute. Rung 5 builds the stamping against this
clause and needs nothing further from rung 7's mechanism.

**R3.7 (the structural axis carries the touch-target floor).** The
`dubplate` match block sets `--size-field` so that a plain `.btn`
renders at 44px or more. Sizing is structural, so the floor belongs to
the match block rather than to each palette. At DaisyUI 5's own scale
`0.275rem` meets it, where the framework default of `0.25rem` yields
40px; the floor is the requirement and the value is the version's answer
to it. `.btn-sm` computes from a smaller multiple and still needs its
own `min-h-11` at each call site. Source: `_salvage/web/src/app.css:58-63`
(U13 F5.3 and F5.9).

> Erratum, 2026-09-25 (pass 5c-ii close): **the estate's composed
> identities are `dubplate-label` and `dubplate-label-night`, not
> `dubplate-gruvbox`.** Geoff's visual identity ruling
> (`docs/superpowers/specs/2026-09-24-visual-identity-design.md`)
> replaced the `gruvbox` palette with `label` (light) and
> `label-night` (dark) before Task 4 of pass 5c-ii dispatched. Every
> place this document names `dubplate-gruvbox` or `palette gruvbox`
> as the estate's configuration (section 3's "the estate config is
> `match dubplate, palette gruvbox`", and R3.6's "The estate's is
> `dubplate-gruvbox`") reads as the two composed identities above; the
> config keys and their grammar are unchanged.

## 4. The three tiers

The build spec ships three tiers over the build: per-user theme choice
via `data-theme`, operator branding as config-rendered CSS variables at
serve time, and an optional operator CSS file served after the bundle.
Source: BS:647-650.

**R4.1 (precedence, lowest to highest).**

1. The composed theme. The per-user pick selects which composed theme
   applies; it is the selector for this layer rather than a layer of its
   own.
2. Operator branding variables.
3. The operator CSS file.

Precedence is established by document order, with the bundle stylesheet
first, the generated branding stylesheet second, and the operator file
last. Document order decides only between declarations of equal
specificity, and the two that matter are equal: DaisyUI emits a theme at
`[data-theme="<name>"]`, which is (0,1,0), and the branding stylesheet
declares at `:root`, which is (0,1,0) as well. No clause uses a
specificity trick to reorder them.

**R4.2 (tier 1, per-user choice).** The pick is a `data-theme` attribute
on the document element, carrying R3.6's composed identity. The Go
binary renders the attribute into the `index.html` it serves, from the
operator's configured default and, once a session identity and the rung
6 store exist, from the signed-in member's stored pick. That stored pick
is a per-user row and keys on rung 6's durable surrogate user id, never
on a roster username or an address, which is 2.0's hook 5
(S20:367-369). The client changes the attribute after hydration when the
member picks a different theme. Rendering the attribute server-side is
what removes the flash of the wrong theme, and it is available because
the binary already serves the document.

The rendered document is per-member, so it is served `no-cache` while
the hashed bundle assets keep their immutable caching, matching the
salvage precedent (`_salvage/internal/serve/serve.go:283-293`).

**Ruled, conductor 2026-09-07:** the pick lives in `localStorage` until
rung 6, then is replaced by the stored per-member value rather than kept
alongside it.

**R4.3 (tier 2, operator branding).** Branding colors arrive as a
stylesheet the Go binary generates from config at serve time and serves
at a fixed path, linked after the bundle stylesheet. The file declares a
`:root` block containing only names from R2.1's set, and it is empty
when nothing is configured.

The three candidate mechanisms were weighed against R1.5:

- A build-time value is declined. It makes the adopter rebuild the
  bundle to change a color, which needs Node on the adopter's machine.
- An embedded config endpoint is declined for color. The page would
  paint unbranded, then repaint after a fetch and a script run, and the
  repaint is visible on every load.
- A generated stylesheet is adopted. It is the ratified wording
  ("config-rendered CSS variables at serve time (no Node, no rebuild,
  per universality)", BS:648-650), it blocks first paint the way any
  stylesheet does, and it needs no JavaScript.

Branding text and images (site name, logo) ride the config endpoint the
app already reads, because neither is a CSS variable and neither causes
a color repaint.

**R4.4 (tier 3, the operator CSS file).** An optional file, named in
config, served after the bundle and after the branding stylesheet. It
has the last word. It is the escape hatch for anything the roster does
not carry (TB:151-157). It is explicitly outside the contract of R1.3: a
file that reaches for a DaisyUI class name or the DOM may break on a
DaisyUI upgrade, and that is the operator's risk to take.

**R4.5 (how the served document is produced).** SvelteKit runs in static
mode with client-side rendering only, no page is server-rendered by
SvelteKit, and server-only SvelteKit files are forbidden and gate-checked
(CM:133-139). The Go serve layer is what produces the served document:
it reads the embedded `index.html` and renders it through one small
documented rewrite that stamps R4.2's `data-theme` value and adds the
tier 2 and tier 3 `<link>` elements the SvelteKit build did not write.
The testable form is a named placeholder token in `app.html` that the
build carries through to `index.html` and the serve layer substitutes.
Every mechanism in this section lives in that rewrite or in the client,
never in a SvelteKit server file.

**R4.6 (no `theme-controller` input).** Every DaisyUI theme block
carries a second selector branch,
`:root:has(input.theme-controller[value="<name>"]:checked)`, whose
specificity is (0,4,1), because `:has()` takes the specificity of its
most specific argument. A page carrying a checked `theme-controller`
input therefore lets tier 1 outrank tier 2 at any link order, which
inverts R4.1 silently. No shipped surface uses that pattern. The pick is
a stamped attribute under R4.2, so the input is redundant, and tier 2
and tier 3 are emitted at a selector no lower than the active theme's.

## 5. Embedded font delivery

**R5.1 (embedded, never a CDN).** Every face ships as woff2 inside the
binary through `go:embed`, served same-origin by the Go binary. No
stylesheet, document, or theme references an external font origin. The
reason is universality and offline use: the box has no obligation to
reach `fonts.gstatic.com`, and an adopter behind a filtered network gets
the same page the estate does. Sources: BS:679-681, PA:56.

**R5.2 (the license bar).** A face ships only under an open license that
permits redistribution and embedding: SIL Open Font License 1.1, Apache
License 2.0, MIT, or a public-domain dedication (CC0 1.0, Unlicense). A
face whose only lawful delivery is a hosted service is not eligible,
whatever its price. The tree carries each face's license text verbatim
and a manifest naming the family, the version, the upstream source URL,
and the license URL. The faces sit in `web/static/fonts/`, the license
texts in `web/static/fonts/licenses/`, and the manifest at
`web/static/fonts/FONTS.md`. G4 is scoped to font files for that reason,
so a license text and a manifest in that tree are not gate failures. The
manifest follows the evidence discipline the 2.0 spec set for prior art,
where a license cell cites the LICENSE file by URL because remembered
license strings were wrong twice (S20:395-397).

**R5.3 (the `@font-face` rules).** Every declaration sets
`font-display: swap`, so a slow first paint shows fallback text rather
than nothing. Every `src` is a same-origin path with `format('woff2')`
and no second format. Every declaration states `font-weight` and
`font-style` explicitly, including for a variable font, which declares
its weight range.

**R5.4 (subsetting).** A face is never subset below the coverage it
ships with. Library metadata is arbitrary Unicode, and a Latin-only
subset paints tofu on a real album title. Splitting one face into
`unicode-range` slices is permitted and encouraged for a large family,
because every slice still ships embedded and the browser fetches only
what the page paints. Trimming glyphs out of the shipped set is not
permitted. R2.7's generic-family tail is the backstop for a glyph no
shipped face covers. This clause is verified by review and carries no
automated gate, because a script cannot tell a trimmed face from a face
that shipped that way.

**R5.5 (how many faces).** A theme ships at most two families and at
most four woff2 files in total, counting a variable font as one file per
family. The budget exists because the bytes live in the binary and every
adopter pays them. **Ruled, conductor 2026-09-07: two families, four
woff2 files, no whole-roster byte ceiling.** A ceiling is added only when
a measured roster needs one.

**R5.6 (the served CSP).** The served header is a directive set. It is
stated here by directive and not as a literal string, because two of its
values depend on what the SvelteKit build emits:

- `default-src 'self'`, which every unlisted directive falls back to.
- `script-src 'self'` plus a hash for each inline script the served
  document carries, and no `'unsafe-inline'`.
- `style-src 'self' 'unsafe-inline'`.
- `frame-ancestors 'none'`, `base-uri 'none'`, `object-src 'none'`,
  `form-action 'self'`.
- No `font-src`. Embedded fonts need no relaxation, and `font-src` falls
  back to `default-src`, so a single-origin default already forbids an
  external font origin at runtime. Adding a font origin to this header is
  a standard violation, not a configuration choice.

Three consequences of the stack shape, stated so a later reader does not
re-derive them. SvelteKit generates inline scripts in the documents it
emits, which is what `kit.csp` exists to cover, so `script-src` has to
permit them or the page loads and never boots. `adapter-static`
prerenders every page and nonces are forbidden on prerendered pages, so
`kit.csp.mode` is `'hash'`, and SvelteKit's own `<meta http-equiv>`
policy and this header combine restrictively, which means the header
cannot be stricter for scripts than the meta policy is. R4.5's rewrite
stamps an attribute and adds `<link>` elements, leaving every inline
script's bytes untouched, so a hash computed over the embedded document
stays valid for the served one. Whether the hashes reach the header from
a build-time manifest or from hashing the embedded document at startup
is rung 5's choice; G6 asserts the outcome either way.

`style-src` carries `'unsafe-inline'` because two mechanisms the stack
depends on need inline CSS that no build-time hash can cover. DaisyUI's
`radial-progress` reads its `--value` from a style attribute, and most
Svelte transitions work by creating an inline `<style>` element at
runtime. Banning both to tighten a directive whose risk is CSS-shaped
costs more than it buys, which is why `script-src` stays hash-pinned and
`style-src` does not. Source: the salvage header
(`_salvage/internal/serve/serve.go:125`) for the four hardening
directives and the single-origin default, adopted; its silence on
`script-src` and `style-src` declined, because it served a stack whose
document carried no inline script.

## 6. Gates

The `web/` leg rejoins the repo gate at rung 7 today (BS:638,
`scripts/check.sh:5`). **Ruled, conductor 2026-09-07: the leg rejoins at
rung 5,** since rung 5's login page is the first web surface. CLAUDE.md's
"rejoins at rung 7" line and the `scripts/check.sh` header comment are
corrected at rung 5's plan, not here.

Each check below either gates or does not exist. None is guarded by a
`command -v` skip. Source: CM gate rule.

G3, G4, G5, G6, and G7 read a built bundle or a served document, so they
run in the `web/` leg, which builds one. Each of them fails with a named
message when the embedded tree is the placeholder-only tree a never-built
checkout carries (`_salvage/web/embed.go` records why that tree exists),
rather than passing on an empty iteration. G1, G2, G8, G9, and G10 read
source and run wherever the leg does.

**G1.** The R2.4 script over `web/src`, exempting
`web/src/lib/themes/` alone. Fails on any of R2.4's four patterns.

**G2.** Variable completeness. A script parses every theme file, asserts
each palette block declares exactly R2.1's 21 palette names and each
match block exactly R2.1's 11 structural names, and fails on a missing
or extra name. It also asserts that each of the three font values ends
in a CSS generic family (R2.7) and that the `dubplate` match block's
`--size-field` meets R3.7's floor. Implements R2.5, R2.7, R3.5, R3.7.

**G3.** No external origin. Fails on `http://`, `https://`, or a
protocol-relative `//` inside any `url()`, `@import`, `<link>`, or
`@font-face src` in `web/src` and in the built bundle. Implements R5.1.

**G4.** Font delivery. Every font file under `web/static/fonts/` ends in
`.woff2` and no other font format is present, and the check skips the
license texts and the manifest R5.2 places there; every `@font-face` sets
`font-display`, `font-weight`, and `font-style`; the per-theme face count
is within R5.5's budget.

**G5.** Embedding. A Go test resolves every `@font-face` src in the
built bundle against the embedded filesystem and fails on a path that
does not exist there. It asserts a nonzero face count first, so a
placeholder-only embedded tree fails by name and no empty iteration
counts as a pass. This is the check that catches a font that builds
locally and 404s on the box.

**G6.** Headers. A Go test asserts R5.6's directive set by name and
value, asserts that `script-src` carries a hash for every inline
`<script>` in the served document and no `'unsafe-inline'`, asserts that
`style-src` carries `'unsafe-inline'`, and asserts no `font-src`
directive is present. The inline-script assertion is what catches a CSP
that would block the page's own bootstrap.

**G7.** Tier order and precedence. A test asserts the served document
links the bundle stylesheet, then the branding stylesheet, then the
operator file, in that order; that the document is served `no-cache`
while a hashed asset is served immutable; and that a branding variable
set in tier 2 wins in the computed value against the active theme's own
declaration of the same name. The computed-value half is what catches a
specificity inversion a link-order assertion cannot see. Implements
R4.1, R4.2, R4.6.

**G8.** Branding surface. A Go test asserts the generated branding
stylesheet declares only names drawn from R2.1's set and emits an empty
file when nothing is configured. Implements R4.3.

**G9.** Contrast. A test computes R2.6's ratios from each composed
theme's declared values, converting to sRGB first, and fails on a pair
below the floor that carries no recorded exception. An exception is one
line directly above the declaration, in the fixed form
`/* contrast-exception: <variable> <measured ratio> <reason> */`, and
the test parses it. A marker whose pairing passes the floor is itself a
gate failure, so a stale exemption cannot survive a palette edit.

**G10.** Structural precedence. The R3.4 test, reading a computed
structural variable under a built-in palette, against the composed
identity R3.6 defines.

The SvelteKit server-file check (CM:137-139) already exists as a
separate rule and is not restated here.

## 7. What this standard does not decide

The shipped theme roster. Rung 7 authors it, with TB section (b) as its
standing input.

The login page's look, the uploader's look, or any surface's visual
design. Rung 7 owns visual authoring and the attended before/after
(PA:56).

Anything the erratum retired. No theme matches Navidrome's or
Jellyfin's factory default, the matched typography note (Roboto, Noto
Sans) is void, and TB's addenda 1 through 3 stand only as history where
they name a streamer match (BS:5-18). TB's cross-platform matrix remains
useful as palette provenance and nothing more.

What the standard owes rung 12: nothing beyond one boundary rule.
Rung 12 is trimmed to the driver seam, SSO, roster sync, and the
proxy-hygiene ingress, and its Navidrome theming note is dropped
(BS:12-15). The driver seam therefore declares no theming capability,
and a driver never receives, sets, or negotiates a dubplate theme
variable. This standard rules the other half of that boundary itself,
since no source states it: an external streamer is themed by whoever
operates it, through that streamer's own configuration, and dubplate
neither matches it nor is matched to it. The seam's own capability model
supports the rule, since theming is simply not among the capabilities a
driver declares (S20:370-373), and the erratum retired the matching that
would have required one (S20:441-444). The driver-declared post-login
destination crosses a visual boundary between two separately themed
applications; that crossing is documented as a boundary rather than
smoothed over, because smoothing it would reintroduce the matching the
erratum retired.

## 8. The 2.0 hook statement and the salvage adoption record

### Forward hooks

This standard touches two hooks from S20:354 onward.

**Hook 5, a stable identity (rungs 5 and 6).** R4.2's stored theme pick
is a per-user row, which is exactly what the hook covers. The standard
lands it by requiring the row to key on rung 6's durable surrogate user
id, never on the roster's renameable handle, so rung 6's schema task
inherits the requirement instead of discovering it. Source:
S20:367-369, with per-user state its own 2.0 slice at S20:347.

**Hook 6, the driver seam stays universal (rung 12).** The standard
lands it by declaring theming out of the seam entirely (section 7). An
in-process streamer has no theme to match, and 2.0's built-in streamer
serves dubplate's own surfaces, which R1.1 already binds. Source:
S20:370-373.

Hooks 1 through 4 and 7 through 14 are neutral to this standard. They
concern schema rows, tag writing, layout portability, and migrations,
and none reaches the presentation layer. The 2.1 first-party web player
(S20:433) is not a numbered hook, and it is the reason R1.1 binds "every
later web surface" rather than an enumerated page set.

### Salvage adoption record

`_salvage/web` is read-only adoption material, admitted on merit
(CM, the greenfield paragraph). Each candidate is adopted or declined
with a reason.

| Candidate | Verdict | Reason |
|---|---|---|
| `src/app.css` `@plugin 'daisyui/theme'` block shape, one declaration per variable, no selectors | Adopt | It is exactly what R1.4's theming-is-data rule describes, already proven to compile against DaisyUI 5. |
| The `navidrome-light` and `navidrome-dark` themes themselves | Decline | The erratum retired streamer matching (BS:5-11). The values have no successor. |
| "Only DaisyUI's own theme tokens, never a DaisyUI internal class name" (`app.css:8-9`) | Adopt | Becomes R2.3 and R1.3. It is the same rule the build spec states as the customization API. |
| The missing-semantic-token finding, U13 F4.1 and F4.4 (`app.css:13-15`) | Adopt | Becomes the justification for fail-loud completeness, R3.5. A gray reject chip is the concrete cost. |
| The `-soft` contrast finding, U13 F8.1 (`app.css:32-47`) | Adopt, and widen | Becomes R2.6's third condition and gate G9. The measured 2.2 to 2.9:1 is the evidence. The mechanism it records is generic to every color DaisyUI paints as text, so R2.6 covers eight colors where the finding measured four. |
| The primary 3:1 carve-out (`app.css:26-30`) | Adopt as a shape, not as a value | R2.6 admits a documented exception at WCAG's UI-component tier. The blue itself belonged to a retired Navidrome match, so the value has no successor. |
| `--size-field: 0.275rem` for a 44px touch target, U13 F5.3 and F5.9 (`app.css:58-63`) | Adopt as R3.7 | Sizing is the match axis, so the 44px floor belongs to the `dubplate` match block rather than to each palette. R3.7 states the floor and treats the value as DaisyUI 5's answer to it. |
| `securityHeaders` CSP (`internal/serve/serve.go:125`) | Adopt in part | The four hardening directives and the single-origin default carry over into R5.6. The header's silence on `script-src` and `style-src` does not: it served a Vite document with no inline script, and SvelteKit's has one. |
| `web/embed.go` `//go:embed all:dist` plus the committed `dist/placeholder.txt` | Adopt the pattern | Fonts ride the same embedded bundle. The placeholder lesson stands: a bare `dist/` fails `//go:embed` on a fresh checkout, so the rebuilt tree keeps an equivalent, and G5 fails by name against that tree instead of passing on it. |
| Static asset cache rules (`internal/serve/serve.go:283-293`) | Adopt | `no-cache` on the document and immutable on hashed assets is exactly what R4.2's per-member rendering needs. |
| Standalone Svelte 5 with Vite, `index.html` plus `main.ts`, no SvelteKit | Decline | Superseded by the SvelteKit static-mode rule (CM:133-136, BS:16-18). |
| `package.json` pins: daisyui 5.7.25, tailwindcss 4.3.3 | Adopt the major line, decline the exact pins | DaisyUI 5 and Tailwind 4 are the contract's basis. The exact versions are rung 7's to choose against the rebuild. |
| Font mechanism | Nothing to adopt | The salvage tree ships no font and no `@font-face`; `web/public` holds only `placeholder.txt`. Section 5 is new work in full. |
