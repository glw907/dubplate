# Rung 5c-iii: pass-end review fold record

This record gives one disposition for every finding in the three pass-end reviews of rung 5c-iii.
Each finding is folded (with its commit), refused (with a one-line reason), or routed (to a named
task or rung). A finding raised by two reviews folds once, under the first ID listed. The standard
is the visual identity spec (VI), revision 1.3. The fold ran on master from `8da75b8`, resumed after
a power loss stopped the first fold agent before it committed anything.

Reviews, all in `docs/research/`:

- `2026-09-25-rung-5c-iii-review-svelte.md` (svelte): 1 blocker, 6 warnings, 13 suggestions.
- `2026-09-25-rung-5c-iii-review-a11y.md` (a11y): 1 blocker, 8 warnings, 9 suggestions.
- `2026-09-25-rung-5c-iii-review-security.md` (security): 2 medium, 5 low, 4 informational.

Commits:

| Commit | Group |
|---|---|
| `81988b1` | The component layer, 28 rendered rows, the prefix watch on every row, the harness env allowlist |
| `2c880db` | The static gates: G1, G4, G5's comment, G9, root-absolute |
| `75fde16` | The document rewrite, the site-name grammar, the theme-color grammar, SVG CSP, variant CSP equality |
| `8588dd0` | VI section citations, FONTS.md's full commit, the rendered script's deadline |

Each commit landed on a full heavy gate (`scripts/check.sh` through `cairn-run-gate`) that exited 0.

## Mutation proof

The new rendered rows ran first against `8da75b8`'s CSS, with the new manifest rows in place. Every
component row went red; the failure text below is abridged from that run. The fixes then turned
every row green. The fonts rows passed at `8da75b8` (S10 is coverage, not a defect), so they were
proven by a mutation instead.

| Row | Failure against `8da75b8` |
|---|---|
| `component/badge/*` | `rendered-fx-badge-primary: text [255,246,234,255] ... text 1.16:1 on its ground`; label-night info `1.30:1`; every ground off base-100 (the noise layer) |
| `component/focus/*/*` | `rendered-fx-wrapped-input: a second ring on the wrapped control itself`; `a:rendered-fx-alert-info: ring contrast inside 1.21:1, outside 1.22:1` |
| `component/surface/flat/*` | `rendered-fx-btn-disabled: box-shadow ... 2px 2px 0px 0px, border solid rgb(31, 27, 22)` |
| `component/surface/ghost/*` | `box-shadow none, border none rgb(31, 27, 22) (want none, solid rgba(0, 0, 0, 0))` |
| `component/pressed/flat` | `rendered-fx-btn-ghost moved (2.00, 2.00); rendered-fx-btn-link moved (2.00, 2.00)` |
| `component/surface/keyline-controls/*` | `rendered-fx-checkbox [30,25,20,51], rendered-fx-toggle [32,28,22,128]` |
| `component/input-states/*` | `rendered-fx-input-error border [31,27,22,255] (want #a8261d)`; hint `[168,38,29,255] (want #a4261d)` |
| `component/link/*` | `rendered-fx-link-warning: [224,165,38,255] at 1.83:1`; `rendered-fx-anchor-btn hovered text-decoration-line underline` |
| `component/utility-wins` | `rounded-full badge radius 0px (want nonzero)` |
| `component/display-face` | `.display font-family "Work Sans", ...` |
| `component/page-title` | `::after ... border-top 1px, border-bottom 0px; the title's own border-top 3px` |
| `component/reduced-motion/kill-switch` | `over 1ms: [{"what":"rendered-fx-btn-default","s":0.2}, ...]` |
| `fonts/ibm-plex-mono` (mutation: `fonts.css` naming a missing file) | `load statuses ["rejected: A network error occurred."]; GET /rendered/fonts/IBMPlexMono-Regular.woff2 logged false` |
| The prefix watch (mutation: a row fetching `/leak-probe`) | `requests outside /rendered/: ["GET /fonts/IBMPlexMono-Missing.woff2","GET /leak-probe"]` |

The same fonts mutation also showed the prefix watch's value. The unresolvable face stayed
root-absolute, and the watch failed every row that loaded it.

Go and static-gate mutations:

- **G9 keyline row.** Before the marker landed: `dubplate-label-night (--color-base-300): --keyline-base measures 2.50:1 against 3:1 and no marker records it`.
- **G1.** The old pattern matched `text-primary-content` and `text-neutral-content`. The new one
  admits both and still refuses `text-primary`, `text-error`, and `text-warning/50`.
- **G4.** A rule with `font-display: block`, one with a second `woff` source, and one with a
  `truetype` source each report a problem. The shipped rules report none.
- **root-absolute.** Each new sink matched on a root-absolute literal: `location.href =`,
  `location.assign`, `location.replace`, `window.open`, `new URL`, `new Request`, `sendBeacon`,
  `formaction=`, and `srcset=`. A relative argument to any of them did not match.
- **L1 and I2.** The new tests failed first. A missing build placeholder beside a site name that
  carried one passed (`error = <nil>`). A site name or href carrying it failed only as `appears 2
  times`. A stylesheet link past `</head>` passed.
- **L2.** Reverting `requireSiteName` to `require` fails 8 of the grammar test's 11 subtests.
- **L4.** `Content-Security-Policy = "", want "default-src 'none'; style-src 'unsafe-inline'; sandbox"` on each SVG.
- **I1.** Disabling the equality check fails with `checkVariantCSP: got nil error, want the policy mismatch refusal`.
- **S8 (theme-color).** The eight-digit and color-function cases failed against the old `#` prefix check.

## Blockers

- **svelte B1 = a11y B1 (badge text vanishes).** Folded, `81988b1`. Text and border read a
  per-status `--badge-ink`, with primary on `--ink-primary`, accent and success on `--ink-accent`,
  and info, warning, and error on their own inks. Secondary, neutral, and the plain badge fall back
  to base-content. The badge's fill is transparent, and `background-image: none` drops DaisyUI's
  noise. The `component/badge/<state>` row measures painted text at 4.5:1 or better for all nine
  badges on both cookie themes and the system stamp in either preference.

## Warnings

- **a11y W1 (the ring on colored fills).** Folded, `81988b1`. A focusable inside an `.alert` takes
  `box-shadow: 0 0 0 6px var(--color-base-100)`. The focus walk samples a link inside every alert
  variant. **VI erratum owed.** VI's "Themes" says `--ink-primary` "clears the 3:1 focus-appearance
  floor on the three bases". It records `label`'s neutral header as a fourth surface that falls
  short. The status alert fills are more such surfaces: 1.03 to 2.34 on `label`, and 1.05 to 2.00
  on `label-night`. The erratum should extend the header's halo sentence to "focusables inside a
  filled container (the header, every status alert)". This fold did not edit VI. **Fix round:**
  the halo rule first sat before the button's and the inputs' own `box-shadow` rules, and at equal
  zero specificity the later rule won, so a focused button or input inside an alert got no halo.
  The rule now follows every `box-shadow` a focusable carries. The fixture adds a button and an
  input inside an info alert, and the walk requires both by role. Against the old order, the focus
  rows failed with `rendered-fx-alert-btn: ring contrast inside 1.22:1, outside 1.33:1` on `label`
  and `inside 1.03:1, outside 1.02:1` on `label-night`.
- **a11y W2 = svelte S1 (wrapped input and select double rings).** Folded, `81988b1`.
  `--input-color` is `--ink-primary` on a control's `:focus` or `:focus-within`, so DaisyUI draws
  its own ring in the ink on any focus. A bare input inside a wrapper drops its own ring. The fixture
  adds `<label class="input">`, and the walk requires no ring on the inner input.
- **a11y W3 = svelte W1 (disabled and link buttons look raised).** Folded, `81988b1`. Ghost, link,
  `.btn-disabled`, and `:disabled` or `[aria-disabled]` buttons get a transparent solid border, no
  shadow, and no press translate. Rows: `component/surface/flat/<theme>` and `component/pressed/flat`.
  **Fix round:** the fold missed svelte W1's outline half. The fill keyline rule also set the border
  of `.btn-outline.btn-<status>`, which is black on `label-night`'s near-black base, so the button
  lost its only edge. An outline or dash button's border now paints the status ink, or
  base-content for a status with none. Row: `component/surface/keyline-outline/<theme>`, which
  failed before the fix with every status outline border at `[0,0,0,255]` on `label-night`.
- **a11y W4 = svelte S2 and S4 (links).** Folded, `81988b1`. The global `a:hover` rule is gone.
  Prose links use DaisyUI's `.link`, which is underlined at rest, with `text-ink-primary`.
  `link-<status>` paints the status ink, and `.btn-link` paints `--ink-primary`. `.link-hover` gains
  a `:focus-visible` twin for W9.2 parity, and `a.btn` stays plain on hover.
- **svelte W2 (validator and state borders).** Folded, `81988b1`. The error and success borders for
  input, select, textarea, and file-input, plus DaisyUI 5.7.46's own validator selectors, are
  restated after the keyline. `.validator-hint` is set in `--ink-error`.
- **svelte W3 (`:where()` specificity).** Folded, `81988b1`. Every selector in the component layer
  sits in `:where()`. DaisyUI 5.7.46 declares `.alert`'s `border-color` outside its sublayers, so
  the fill keyline reaches alerts through `--alert-border-color` rather than a heavier selector. The
  `component/utility-wins` row proves `rounded-full` beats the layer's squared badge.
- **svelte W4 (`.display` sets no slab face).** Folded, `81988b1`. `font-family: var(--font-serif);
  font-weight: 400`, proven by `component/display-face`.
- **svelte W5 (the page-title rule frames the title).** Folded, `81988b1`. The rule sits under the
  title as one `::after`: a 3px line, a 2px gap, and a 1px line. The VI specimen agrees. Proven by
  `component/page-title`.
- **svelte W6 = a11y W5 (G1's `-content` regex).** Folded, `2c880db`. The pattern ends with
  `(?![\w-])`.
- **a11y W6 (selection-control keylines).** Folded, `81988b1`. `.checkbox`, `.radio`, `.toggle`,
  and `.file-input` join the keyline rule. Proven by `component/surface/keyline-controls/<theme>`.
- **a11y W7 (the global reduced-motion kill switch).** Folded, `81988b1`. It is W9.3's unlayered
  `!important` block in the new `web/src/lib/styles/daisyui-overrides.css` (W5.22's file), imported
  last. Proven by `component/reduced-motion/kill-switch`: no element of the fixture set runs a
  transition or animation over 1ms under `reducedMotion: 'reduce'`. It is the first entry in W5.8's
  pinned unlayered set. The `unlayered-count` ratchet gate W12 names does not exist yet; it goes to
  Task 9 with the header work.
- **a11y W8 = svelte S7 (G9 rows, the `label-night` keyline on base-300).** Folded, `2c880db`
  (G9) and `81988b1` (the focus walk). G9 measures `--keyline-base` against the three bases at
  3:1 and base-content on the resolved OKLab `--selection-ground` at 4.5:1 (11.10 on `label`, 5.44 on `label-night`). The pair count moves
  from 66 to 74. `label-night`'s keyline on base-300 at 2.50:1 carries a `contrast-exception`
  marker naming base-300, since no page places a control on base-300 yet. G9 refuses a marker whose
  ratio or base matches no failing pair. The focus walk now runs the `dubplate-system` stamp in both
  preferences on both pages. The ring against the neutral header stays with Task 9's header halo
  (Ruling 7).

## Security

- **M1 (sinks outside the source scan).** Folded, `81988b1` and `2c880db`. Every rendered row now
  fails when any request the server logged since the previous row left the owned prefix. A final
  sweep refuses a request logged after the last row. root-absolute scans the listed sinks, and
  since the fix round a root-absolute URL in a later `srcset` candidate too. The built-JS scan is
  **refused**: the per-row request watch covers the code paths the rendered rows exercise, and a
  static scan of minified dependency code would flag inert strings. The G7 branding and operator
  fixture servers sit outside the request log, so requests to those two bases are outside M1's
  watch. Sharing the main log would interleave their records with the offsets each row reads.
- **M2 (CSP scope under a same-origin streamer).** Routed to rung 12, which adds the fallthrough
  streamer: scope `script-src`, `connect-src`, and `form-action` to `<origin><prefix>/`.
- **L1 (the placeholder in operator text).** Folded, `75fde16`. `BuildDocumentVariants` runs
  `RewriteDocument` on the build's bytes before inserting the head. An inserted value that carries
  the placeholder (the site name, a prefix in any href) is refused by name.
- **L2 (`DUBPLATE_SITE_NAME` grammar).** Folded, `75fde16`. Valid UTF-8, no C0, DEL, or C1
  control, no bidirectional embedding, override, or isolate, and at most 100 bytes. The refusal
  names the key and never echoes the value. The grammar lives in rung 1's own `Load`
  (`requireSiteName`), and `keys.go` is untouched. The placeholder refusal is L1's, at the
  document layer. Routed to Task 7: render the site name by text interpolation, never `{@html}`.
- **L3 (the branding neutral trusted on a `#` prefix).** Folded in part, `75fde16`. A branding
  neutral reaches `theme-color` only when it matches the three- or six-digit grammar, now named
  `themeColorRE`. Anything else drops the meta and logs once. The single-parse half is **refused**:
  the second read re-splits a string `BrandingStylesheet` already validated, and moving it would
  change `serve`'s API for no reduction in risk.
- **L4 (SVG responses carry no CSP).** Folded, `75fde16`. Every `image/svg+xml` response carries
  `default-src 'none'; style-src 'unsafe-inline'; sandbox`.
- **L5 (the harness passes the whole environment).** Folded, `81988b1`. The script inherits only
  `PATH`, `HOME`, `TMPDIR`, `LANG`, `LC_ALL`, `TZ`, the three `XDG_*` directories, and
  `PLAYWRIGHT_BROWSERS_PATH`, plus the fixture's own variables.
- **I1 (the variant CSP is discarded).** Folded, `75fde16`. `checkVariantCSP` requires each
  variant's policy to equal the served header.
- **I2 (the anchor may sit outside `<head>`).** Folded, `75fde16`. The last stylesheet link must
  precede the first `</head>`.
- **I3 (operator CSS is unlayered).** Refused: unlayered by design, so DL R4.4 gives the operator
  the last word.
- **I4 (the harness request log on loopback).** Refused: it serves only under the `rendered` build
  tag on loopback, and L5 now keeps secrets out of the script's environment.

## Suggestions

- **svelte S3 (the list frame).** Folded, `81988b1`. The frame carries `--radius-box`, so every
  `.list` is a framed box like a card.
- **svelte S5 = a11y S1 (press feel and the ghost button).** Folded, `81988b1`. Ghost buttons keep
  a transparent solid border, which survives forced colors, and do not translate. The pressed-row
  comment now reads correctly: DaisyUI transitions `box-shadow` and `transform` but not `translate`.
- **svelte S6 (badge radius against the specimen).** Refused: VI owner ruling 12 says "squared
  outlines", which outranks the specimen's slight rounding. The radius stays 0.
- **svelte S7.** Folded with a11y W8.
- **svelte S8 (`sourceVersion()` misses `web/scripts/`).** Folded, `81988b1`, by hashing
  `scripts/`. `@source not` is declined: the rendered fixture's markup there is what puts the
  classes it measures (`badge-accent`, `link-warning`, `validator`, and others) into the bundle.
  Excluding it would leave those rows measuring CSS that never ships.
- **svelte S9 (gate precision).** Folded, `2c880db`. G4 requires `font-display: swap` and one
  `url()` in `format('woff2')`. G5's comment now matches its set check. root-absolute's sinks fold
  under security M1.
- **svelte S10 (fonts under the prefix).** Folded, `81988b1`. `fonts/<face>` rows load each face
  and log its GET from inside the prefix, with the request log opened before the page loads.
- **svelte S11 (stale citations).** Folded, `8588dd0` (VI section names in `inks.css`,
  `surface.css`, `g9.ts`, and `run.ts`, plus FONTS.md's full Work Sans commit
  `b35c81086186162164947bd39574683073d9b268`) and `81988b1` (the `components.css` header).
- **svelte S12 = a11y S9 (stub copy).** Routed to Task 8.
- **svelte S13 (the input inset shadow against the specimen).** Recorded for the rung 7 uploader
  design review, the attended before/after. A related observation for the same sitting: the
  specimen draws `label-night`'s input edge near base-content, while the layer draws every control
  edge in `--keyline-base` (`#706a5f` on `label-night`). VI's dark-keylines row names buttons,
  cards, and lists only.
- **a11y S2 (per-variant `color-scheme`).** Routed, VI erratum owed. A `label` cookie stamp still
  carries `color-scheme: light dark` (W5.10). Emitting per variant changes VI's "Every document
  carries `<meta name="color-scheme" content="light dark">`".
- **a11y S3 (the stale `components.css` header).** Folded with svelte S11.
- **a11y S4 (a W5.8 placement erratum).** Recorded, erratum owed to the web surface standard. W5.8
  says a rule that must beat DaisyUI is "unlayered by necessity". The component layer instead beats
  DaisyUI from `@layer utilities`' own implicit sublayer, which also keeps the operator's tier 3
  file winning. The kill switch is the one rule here that needs to be unlayered.
- **a11y S5 (the badge size).** Folded, `81988b1`: `0.6875rem`.
- **a11y S7 (mono utilities, header link hover).** Routed to Task 9 (the header) and rung 7 (the
  mono surfaces).
- **a11y S6 (a `text-ink-success` alias).** Folded, `81988b1`. It is an alias of `--ink-accent`
  (VI: "`--ink-accent` (also success)"), and G1 admits it.
- **a11y S8 (three- or six-digit hex for `theme-color`).** Folded with L3, `75fde16`.

## Counts

The three reviews raise 49 findings. Ten repeat another review's finding: a11y B1, W2, W3, W5, S1,
S3, and S9, and svelte S2, S4, and S7. That leaves 39 distinct findings:

- 30 folded. Two of them decline a sub-part, each with its reason above: M1's built-JS scan and
  L3's single parse.
- 3 refused: I3, I4, and svelte S6.
- 6 routed or recorded: M2 (rung 12), svelte S12 (Task 8), a11y S7 (Task 9 and rung 7), svelte S13
  (the rung 7 design review), a11y S2 (a VI erratum), and a11y S4 (a web surface standard
  erratum). L2's `{@html}` note also goes to Task 7.

Owed errata: VI for W1's halo scope and for a11y S2's `color-scheme`, and the web surface standard
for a11y S4's layer placement.

## Owed by the conductor

- The `code-simplifier` pass over `81988b1..8588dd0`. A subagent cannot dispatch one.
- The `go-architecture-reader` reads for `internal/serve`, `internal/config`, and
  `cmd/dubplate-server` at the merge.
- The rendered leg now runs about 55s on an idle machine, up from about 17s. Under another lane's
  load it ran past 150s, so the script's deadline is now 300s.
