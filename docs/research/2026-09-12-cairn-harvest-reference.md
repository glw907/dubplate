# cairn harvest reference for the web surface standard

Companion to `docs/superpowers/specs/2026-09-12-web-surface-standard.md`
(the standard). The standard carries the requirements. This file carries
the evidence: where each clause came from, what cairn measured, which
cairn rulings were reversed before adoption, and what was read and
declined.

Read this file when you ask why a clause exists, when you re-harvest
cairn after a cairn release, or when you want to check a citation. A rung
implementer does not need it.

## How to re-verify a citation

Sources were read on 2026-09-12 against `~/Projects/cairn-cms` at commit
`446d5d5`, package version 0.96.0. That is a workstation path, so a
reader without it resolves the same files through the cairn repository at
that commit. Line numbers drift with the repo; the commit does not.

Citation corrections applied in the standard's revision 2 are marked
below with the lens A finding that caught them.

## Short names

| Short name | Location under `cairn-cms/` |
|---|---|
| ADS | `docs/internal/admin-design-system.md` |
| PDS | `docs/internal/public-design-system.md` |
| CI | `docs/internal/code-idioms.md` |
| AUD | `docs/reference/cairn-audit.md` |
| CRAFT | `skills/cairn-admin-screens/references/craft.md` |
| FORM | `skills/cairn-admin-screens/references/form-anatomy.md` |
| GRADE | `skills/cairn-admin-screens/references/grader-prompt.md` |
| TK | `src/lib/admin-toolkit/`, `src/lib/components/`, `src/lib/design/`, and `scripts/build/` (CIT-8) |
| WM | `templates/waymark/src/` |
| RULES | `src/lib/audit/rules/` |

Four files cited as `TK` do not sit under the two directories revision 1
named (CIT-8): `admin-css.input.css` is at `scripts/build/`,
`grammar-tokens.ts` at `src/lib/design/`, `admin-sheet-inventory.test.ts`
at `src/tests/unit/`, and `playwright.config.ts` exists only at
`examples/showcase/`. The `TK` row above is widened to cover the first
three; the playwright config is cited by its full path.

dubplate's own documents are cited in the standard by their own short
names (DL, R5, S20) and are not cairn sources.

## Per-clause sources

| Clause | Sources |
|---|---|
| W1.1 | owner ruling 1 |
| W1.2 | owner ruling 2; ADS:56-57 (cairn's own register line, as the shape) |
| W1.3 | owner ruling 3 |
| W1.4 | `docs/superpowers/research/passb-copy-final.md` (the ratified reject-copy deck); CLAUDE.md, "Reject wording is contract", for verbatim sharing |
| W2.1 | ADS:296-323; WM `theme.css:203-212` |
| W2.2 | FORM:45-73 |
| W2.3 | `scripts/build/admin-css.input.css:186-210` |
| W2.4 | ADS:305-323; `scripts/build/admin-css.input.css:125-165` |
| W2.5 | WM `theme.css:33-39`; `scripts/build/admin-css.input.css:170-205` (the `@utility` half, added at revision 2 per C-7) |
| W2.6 | AUD:75; RULES `static/grammar-boundary.ts` |
| W2.7 | ADS:207-221; `scripts/build/admin-css.input.css:85-100`; WM `tokens.css:36-40` |
| W2.8 | `src/lib/design/grammar-tokens.ts:1-8` |
| W2.9 | PDS:108-135; WM `theme.css:195-212` |
| W2.10 | WM `chassis/README.md:84-95` and `:96-110`; ADS:340-344 |
| W3.1 | ADS:122-128 |
| W3.2 | ADS:155-162; ADS:471 (CIT-7, replacing ADS:466-467); PDS:157-162 |
| W3.3 | TK `PageHeader.svelte:3-14`; ADS:459-462 |
| W3.4 | ADS:1024-1031 |
| W3.5 | ADS:363-366; ADS:139-151 |
| W3.6 | PDS:153-157 (CIT-15); WM `(site)/+layout.svelte:72-97`; WM `site.css:46-53` |
| W3.7 | WM `+error.svelte:1-49`; WM `chassis/README.md:169-199`, whose rule is scoped to an unmatched path (CIT-4) |
| W3.8 | WM `composition.css:117-138` |
| W3.9 | WM `composition.css:15-138`; WM `prose.css:71-112`; WM `chassis/README.md:152-159` |
| W3.10 | WM `composition.css:99-103` |
| W4.1 | ADS:28-29; RULES `rendered/viewport-overflow.ts:5-6` |
| W4.2 | ADS:43-44 (CIT-12); ADS:22-27; ADS:45-49 |
| W4.3 | ADS:58-60 |
| W4.4 | CRAFT:87-97; CRAFT:137-144 (the two disagree on the stacked gap, CIT-11) |
| W4.5 | CRAFT:98-109; CRAFT:145-224; WM `prose.css:347-356`; `docs/internal/record/2026-07-craft-chapter-acceptance.md:212-280`; Roselli, "Tables, CSS Display Properties, and ARIA" |
| W4.6 | ADS:833-840 |
| W4.7 | ADS:494-501; ADS:518 (the two-pane collapse, CIT-13) |
| W4.8 | WM `SiteHeader.svelte:26-27`, `:106-111` |
| W4.9 | PDS:135-146; WM `site.css:34-44` |
| W4.10 | RULES `rendered/touch-targets.ts:2-35`; WM `prose.css:693-694`; ADS:203-204; WCAG 2.2 SC 2.5.8 and SC 2.5.5 |
| W4.11 | ADS:196-200 |
| W4.12 | ADS:1054-1066; PDS:303-315; CRAFT:77-81 |
| W5.1 | TK `AdminTable.svelte:1-24`; TK `ExpandableRow.svelte:56-74` |
| W5.2 | TK `PageHeader.svelte:21-23`; TK `AdminTable.svelte:10-12`; TK `ExpandableRow.svelte:94-96` |
| W5.3 | TK `ExpandableRow.svelte:8-14`; TK `ToolbarDisclosure.svelte:11-19` |
| W5.4 | CI:142-145; TK `ToolbarDisclosure.svelte:133-227`; WM `Banner.svelte:7-9`; Svelte docs for `$props.id()`, `$state.raw`, and `{@attach}` |
| W5.5 | TK `FieldLabel.svelte:57-61` |
| W5.6 | TK `pagination-window.ts:1-6`; TK `index.ts:11-14` |
| W5.7 | TK `AdminTable.svelte:8-10`; WM `SiteHeader.svelte:166-168` |
| W5.8 | ADS:76-83; AUD:80; the compiled layer order, measured (see below) |
| W5.9 | DL R4.2, R4.1; ADS:70-75 read and declined as a cairn embedding choice (C-2) |
| W5.10 | DL R4.2, R4.5; SvelteKit CSP documentation and kit issue 10510 (C-3) |
| W5.11 | TK `StatusChip.svelte:144-151` |
| W5.12 | TK `ToolbarDisclosure.svelte:3-9`, `:85-91`, `:179-276` |
| W5.13 | ADS:526-530; TK `ToolbarDisclosure.svelte:293-317`; daisyUI dropdown docs; MDN Popover API |
| W5.14 | CI:146-150; TK `dialog-origin.ts:1-20`; ADS:489-493 |
| W5.15 | ADS:902-912; TK `ListToolbar.svelte:48-59`, `:198-265`; ARIA APG table and radio patterns |
| W5.16 | ADS:595-615; TK `DeleteDialog.svelte:88-91`; Tailwind Preflight; Scott O'Hara on lists and Safari |
| W5.17 | AUD:73; ADS:843-849; TK `StatusChip.svelte:14-18`; daisyUI 5.7.20 source |
| W5.18 | ADS:222-238; `src/lib/components/cairn-admin.css:1016-1019` and `:1054-1057` |
| W5.19 | CI:91 |
| W5.20 | WM `site.css:55-81`; WM `(site)/+layout.svelte:55-63` |
| W5.21 | ADS:345-349; Tailwind source-detection docs |
| W5.22 | `src/lib/components/cairn-admin.css` (cairn's pinned unlayered set) |
| W5.23 | `admin-toolkit/index.ts:12-14` (the retired wrappers, as the counter-evidence) |
| W6.1 | FORM:13-43; ADS:263-284 |
| W6.2 | TK `FieldLabel.svelte:17-30`; FORM:114-118; HTML Standard, the `label` element |
| W6.3 | FORM:82-129; TK `FieldLabel.svelte:5-12` |
| W6.4 | FORM:131-140; R5's page descriptions |
| W6.5 | TK `MediaCaptureCard.svelte:12-15`; TK `typed-confirm.ts:1-5`; ADS:917-923 |
| W6.6 | TK `FieldInput.svelte:130-147`, `:208-212` |
| W6.7 | CI:22-27; CI:46-50 |
| W6.8 | ADS:809-821 |
| W6.9 | ADS:556-562 |
| W6.10 | WCAG 2.2 SC 2.5.7 |
| W6.11 | WCAG 2.2 SC 3.3.7 and SC 1.3.5; R5:1406 |
| W7.1 | ADS:448-458; TK `EmptyState.svelte:8-21`; TK `ConceptList.svelte:449-461` |
| W7.2 | ADS:448-450 |
| W7.3 | ADS:378-380; TK `MediaHeroField.svelte:496-497` |
| W7.4 | ADS:511-516 |
| W7.5 | TK `MediaInsertPopover.svelte:358-378`; TK `PreviewBanner.svelte:15-18` |
| W7.6 | TK `ConceptList.svelte:265-310`; ADS:956-959; CI:155-156; WordPress core ticket 36853 (the nonce character) |
| W7.7 | ADS:917-923; TK `DeleteDialog.svelte:71-115`; TK `MediaUploadDialog.svelte:3-6`; ARIA APG alertdialog |
| W7.8 | ADS:546-551; daisyUI `button.css` |
| W7.9 | ADS:570-591; ADS:938-945 |
| W7.10 | TK `StatusChip.svelte:3-25`, `:77-84`; TK `ManageEditors.svelte:10-12` |
| W7.11 | ADS:946-963, retargeted at rung 13 (G-4) |
| W7.12 | ADS:932-937 |
| W7.13 | ADS:964-973 |
| W7.14 | TK `ConfirmPage.svelte:54-57`; TK `LoginPage.svelte:57-66`, `:135-154` |
| W7.15 | TK `LoginPage.svelte:4-5`, `:46-56`, `:72`, `:101-149`; TK `ConfirmPage.svelte:3-4`; R5 fork 4 |
| W7.16 | ADS:809-815 |
| W7.17 | ADS:990-997; ARIA in HTML for `progress` |
| W7.18 | WCAG 2.2 SC 3.3.8 and technique G218; R5:1401-1410 |
| W7.19 | `docs/superpowers/specs/2026-08-31-dubplate-uploader-design.md:32-53`, `:222-255` |
| W8.1 | ADS:1148-1164 |
| W8.2 | `docs/internal/record/2026-07-15-admin-ux-audit.md:158-171` (cairn's three leaks, replaced by dubplate's own list) |
| W8.3 | ADS:1148-1164 |
| W8.4 | TK `format.ts:98-121` |
| W8.5 | `exemplar-list.md:92-94` (CIT-6); WCAG 2.2 SC 2.5.3 |
| W8.6 | ADS:657-659 |
| W8.7 | TK `format.ts:1-38`; CRAFT:73-76; ADS:1295-1299 |
| W8.8 | CI:36-39; TK `media-upload-outcome.ts:46-58` |
| W8.9 | TK `MediaUploadDialog.svelte:46-50`; `research/passb-copy-final.md`; R5:1428-1438 |
| W9.1 | WM `tokens.css:156-181`; PDS:184-188 |
| W9.2 | AUD:76; TK `ExpandableRow.svelte:175-182`; RULES `static/focus-parity.ts:12-16` |
| W9.3 | AUD:77-78; WM `prose.css:866-877`; daisyUI 5.7.20 emitted sheet, counted |
| W10.1 | CI:53-59 |
| W10.2 | CI:96-108 |
| W10.3 | CI:68-72 |
| W10.4 | CI:151-154 |
| W10.5 | WM `SiteHeader.svelte:8-9`, `:50-52` |
| W10.6 | WM `icons.ts:2-24`; WM `SiteHeader.svelte:135-156`; WM `prose.css:577-587` |
| W10.7 | WM `tsconfig.json:1-12`; WM `svelte.config.js:21-55` |
| W10.8 | CI:174-182 |
| W11.1 | `design/README.md:7-28` |
| W11.2 | `design/README.md:30-60`; PDS:280-283 |
| W11.3 | WM `(site)/styleguide/+page.svelte:1-15`, `:24-95`, `:368-379` |
| W11.4 | `skills/cairn-admin-screens/SKILL.md:88-109` |
| W11.5 | GRADE:11-20, `:38-42`, `:60-62`, `:127`, `:133` for the capture set and the reader's discipline; `docs/internal/record/2026-07-craft-chapter-acceptance.md:255-264` for k=3 and the 2-of-3 bar (CIT-5) |
| W11.6 | `docs/internal/record/2026-08-07-vertical-alignment-harvest-findings.md:124,128` for the 51-pixel measurement under a 120 floor (CIT-9); ADS:1288-1290 for the baseline-update rule |
| W11.7 | TK `PageHeader.test.ts:6-10`, `:97-136`; CI:135-137; TK `ToolbarDisclosure.test.ts`; WM `admin-link.test.ts:5-14` |
| W11.8 | `packages/cairn-cms-dev/README.md:1-50`; WM `vite.config.ts:6-31` |
| W11.9 | `design-your-site.md:76-99` |
| W11.10 | `docs/internal/record/2026-07-17-waymark-final-design-review-audit.md:54-157` |
| W12.1 | AUD:36-50, `:141-143`, `:167-221`; `src/lib/audit/bin.ts:59-69`; `run.ts:40-42`; `rendered/identity.ts:13-33`, `:56-76`; `rendered.ts:145-198`, `:261-271` |
| W12.2 | AUD:38-41, `:88-113`, `:244`; `engine-rulings.md:4694-4700`; `src/lib/audit/suppress.ts:21` |
| W12.3 | RULES `static/*`; `src/lib/audit/run.ts:29`; `package.json` `files` (which excludes `scripts/`) |
| W12.4 | RULES `rendered/*`; AUD:225-250 for the tiering |
| W12.5 | AUD:252-255 |
| W12.6 | `scripts/checks/check-touch-targets.mjs:25-28`; `scripts/test/contained.mjs`; `rendered/bootstrap.ts:170-186` |
| W12.7 | WM `package.json:5-14`; WM `.prettierrc`; `vitest.config.ts:27-160`; `eslint.config.js:8-58` |
| W12.8 | `LICENSE:1` (MIT); `package.json` dependency list; `bin.ts:63`; `config.ts:68-110` |
| W12.9 | WCAG 2.2 |
| W13.1 | daisyUI 5 release notes; DL R1.3 |
| W13.2 | ADS:1266-1293; `cairn-admin.css:1014-1015` |
| W13.3 | `src/tests/unit/admin-sheet-inventory.test.ts:6-16`; `OfficeList.svelte:16-21`; `docs/internal/record/2026-09-08-polish-inputs/admin-sweep.md:41`, `:94` |
| W13.4 | ADS:1266-1293; `docs/HISTORY.md:612-620` |
| W13.5 | W11.3's route; `admin-css-build.test.ts` |
| W13.6 | "What's new in Svelte: September 2026" |
| W13.7 | DL header and R1.5 |
| W13.8 | DL R1.5, R4.4 |
| W13.9 | DL R2.1, R3.7, R4.3; daisyUI 5 release notes |

## Measurements the standard states without their working

**The compiled cascade layer order.** Compiling stock
`@import "tailwindcss"; @plugin "daisyui";` through
`@tailwindcss/postcss` 4.3.3 with daisyUI 5.7.20 emits
`@layer properties;` then `@layer theme, base, components, utilities;`,
and daisyUI's component rules land in nested sublayers of `utilities`
(`utilities.daisyui`, `.l1`, `.l1.l2`, `.l1.l2.l3`). A layer's own direct
content beats its sublayers, which is why Tailwind's utilities still beat
daisyUI and why an `@layer components` rule loses to both. This is the
measurement behind W5.8, W5.18, and W5.22. cairn's own fix is unlayered
and says so at `cairn-admin.css:1016-1019`.

**The control-edge mixes.** daisyUI's `.input` sets
`--input-color: color-mix(in oklab, var(--color-base-content) 20%, #0000)`
and `.checkbox` borders at the same fallback. cairn measured that at
1.492:1 light and 1.773:1 dark, and a 55% mix at 3.586:1 light and
4.959:1 dark (ADS:222-238). `.toggle` measured 3.118:1 light and 4.370:1
dark and needed no change.

**The status fills.** cairn measured `--color-warning` as small text on
the base surface at about 2.2:1 (ADS:207-215). That is the evidence for
W2.7's fill-versus-ink split.

**The touch-target history.** cairn's rendered gate ran at 44px and
raised 138 findings, almost all against daisyUI's own `btn-sm` (32px) and
`btn-xs` (24px), then moved to the 24px activation-region floor
(`rendered/touch-targets.ts`). dubplate does not inherit that argument,
because DL R3.7 forbids the condition: `.btn-sm` carries its own
`min-h-11` at each call site. What transfers is the second argument only,
that a token check cannot see a rendered 20px checkbox or a label that
activates it.

**The motion population.** daisyUI 5.7.20's emitted sheet carries 391
transitions at `.2s`, 141 at `.3s`, 141 at `.1s`, 14 at 75ms, 12 at 20ms,
10 at `.4s`, 6 at 50ms, 3 at `.5s`, one at `1s`, and one at `15s`, and
exactly one `prefers-reduced-motion` block. That is why W9.3 binds
authored CSS and why the reduced-motion guarantee is a rendered
assertion rather than a source rule.

**The Tailwind spacing collision.** With
`@theme { --spacing-xs; --spacing-xl; --spacing-2xl }`, Tailwind 4.3.3
compiles `.max-w-xs`, `.max-w-xl`, and `.max-w-2xl` to the spacing values
while `.max-w-sm` still reads `--container-sm`. Spacing is read before
container. The collision is a namespace-design fact, so it covers every
key Tailwind's `--container-*` set defines, not the three cairn happened
to declare.

**The `type-*` namespace.** Compiling
`@theme { --gap-label: .25rem; --type-body: 1rem; }` with candidates
`gap-label` and `type-body` emits `.gap-label` and nothing for
`type-body`. `--gap-*` is a Tailwind namespace; `--type-*` is not. cairn
declares all thirteen of its roles through `@utility`
(`scripts/build/admin-css.input.css:170-205`). That is W2.5's correction.

**The pixel-diff floor.** A 1.5px icon shift measured 51 differing pixels
under a 120-pixel tolerance floor
(`record/2026-08-07-vertical-alignment-harvest-findings.md:124,128`).

## Reversals not imported

cairn's tree records rulings its earlier documents and its shipped skill
still state in their first form. The standard takes the later form of
each, and a reader of the cairn sources should too.

- The chip tone dot and the bounded/quiet register pair were retired for
  the fused three-register set (ADS:394-423). `craft.md:49-54` and the
  grader prompt's item (c) still teach the first form.
- `text-box: trim-both` was specified as a silent default, measured, and
  not landed: it is inert on flex containers, it moves about 6px across
  25 sites and quietly re-rules a ratified spacing scale, and it breaks
  the type-role contract (ADS:1113-1137). It is also not Baseline, with
  no Firefox implementation as of mid-2026, which is the durable reason.
  The workstation note that calls it a cairn default is stale. dubplate
  does not adopt it.
- The rendered touch floor moved from 44 to 24 (W4.10).
- The trailing create button below a list was removed as a duplicate of
  the header action (ADS:459-462).
- `FieldRow`, `SelectInput`, and `TextInput` were retired for zero
  consumers (`admin-toolkit/index.ts:12-14`); the recipe is a label
  wrapper plus a bare control. This is the evidence against a full
  wrapper set, which is why W5.23 mandates only the coupled families.
- The `.list`-as-bulleted-list ruling was rewritten after its first claim
  about marker suppression measured wrong; the zero gap came from
  `padding-inline-start: 0` (ADS:595-615). Note that cairn's sheet omits
  Preflight, so that ruling was written for a Preflight-absent world;
  dubplate keeps Preflight and W5.16 states the consequence.
- The audit viewport matrix checked 390 alone until an adversarial review
  found it undershot the bar; 320 was added.
- `badge-ghost` was retired from cairn's tree and is still compiled into
  the shipped sheet through a compatibility safelist
  (`admin-css-safelist.ts:95-102`), so "refused" and "absent" came apart.
  That is the drift W5.17's version stamps and falsifiable tests exist to
  catch.

## What was read and declined

| Item | Reason |
|---|---|
| SvelteKit form actions, the `form` prop, `CsrfField`, the cookie-at-SSR chrome state | dubplate is static and client-rendered; the Go server owns CSRF; the accessibility and copy patterns built on top survive the swap |
| `csrf.checkOrigin: false` | Hands CSRF to a runtime dubplate does not have |
| The prose content surface (`prose.css`) as a whole, the islands runtime, `?url` stylesheet imports, `rel="external"` on admin links | Markdown delivery and the cairn preview frame; dubplate renders no markdown |
| The chassis boundary gate that parses a README table | dubplate has one app, not a theme market; the mechanism (derive an allowlist from the doc that documents it) is noted for the config-homes boundary |
| `@source inline(...)` safelisting | Exists because a consumer's markup is never scanned; dubplate scans its own |
| Fontsource imports | DL section 5 embeds woff2 with a license manifest |
| `@lucide/svelte` | W10.6 hand-authors path data for the embedded bundle |
| The office and desk breakpoints, the type-scale numbers, the gap values, the 4/8px ladder's steps, the register line, the flat chrome band's fill and shadow posture, the 85vh dialog cap, the middle-dot separator | cairn design choices; the shapes are adopted, the values are rung 7's |
| The docs-figure gate (`check-visuals`) | Docs-specific |
| `data-theme` on a bare inner wrapper (ADS:70-75) | A consequence of cairn scoping its whole sheet so it can embed inside a host page. dubplate owns the document, and DL R4.2 puts the attribute on the document element |
| The whole-sheet class inventory (`admin-sheet-inventory.test.ts`) | cairn's shipped sheet is a de facto public API for consumer markup. dubplate's sheet ships inside a binary and no consumer writes markup against it, so W13.3 narrows the inventory to the classes dubplate's own markup writes |
| `norms-bands` | Grades a tree against cairn's own measured control geometry. Owner ruling 1 forbids importing cairn's look |
| `panel-width` | Measures `ExpandableRow`, a cairn component dubplate does not ship |
| `screen-anatomy` as an engine rule | Its content is cairn's office and desk anatomy read off cairn's shell classes. The one-`h1`-per-page half becomes a small dubplate rule |
| `gap-scale` | Its subject is bracket utilities, which W2.4 forbids outright. Gate W2.4 directly |
| `token-resolution` | DL G2 already asserts each theme file's exact declared name set |
| `relational-spacing` | Advisory, expensive, welded to `--cairn-gap-*`. Revisit past rung 7 |
| A full DaisyUI wrapper component set | cairn retired three such wrappers for zero consumers. W5.23 mandates recipes only for the families carrying a refused default, an edge fix, or a measured floor |

## Engine rules welded to cairn

These are the rules a dependency cannot subset and a port must fork. The
count matters, because revision 1 named two.

Rendered: `screen-anatomy` (`.card-shell`, the `.drawer` desk exemption),
`container-inset-asymmetry` (`.card-shell, .list, .modal-box`),
`relational-spacing` (`--cairn-gap-*` under `[data-theme='cairn-admin']`),
`form-font-parity` (the same theme selector), `panel-width`
(`ExpandableRow`), `norms-bands` (cairn's measured geometry).

Static: `type-scale`, `gap-scale`, and `grammar-boundary` (all three
import cairn's `GRAMMAR_TOKENS` and filter to `--cairn-type-*` or
`--cairn-gap-*`), and `stock-default-hazards` in part (its guarded-control
check keys on the class `cairn-btn-guarded`).

Harness constants: the redirect trap's hardcoded `/admin/login`
(`rendered.ts:75`), the `cairn-admin-theme` cookie, and the suppression
directive marker `cairn-audit-disable-next-line` (`suppress.ts:21`).

Four of W12.3's rows have no engine source at all, because cairn's
`scripts/` directory is not in the published package's `files` list:
`check-admin-prose.mjs`, `check-editor-quotes.mjs`, `check-idioms.mjs`,
and `check-public-tokens.mjs`.

## The rendered engine and a client-rendered bundle

The rendered runner compares a JavaScript-disabled capture against the
settled one, on title and landmark (`rendered/identity.ts:13-33`,
`:56-76`). For a client-rendered bundle the first capture is the shell,
which carries no `<main>`, so the identities never match and
`rendered.page-identity-mismatch` fires at error tier on every page with
zero rules run (`rendered.ts:195-198`, `rendered/findings.ts:199-212`).
The alternative path is worse: the runner stores the JavaScript-disabled
identity, so for a client-rendered app every stored identity is the same
shell and the redirect-trap refusal fires unconditionally at exit 2
(`rendered.ts:145`, `:157-160`, `:261-271`).

That is the measurement behind W12.8's ruling that a port is the primary
path for the rendered half.

## Why revision 2 rules as it does

The standard states its rules and one reason each. The longer arguments
the seven lenses supplied sit here.

**The inline theme script (W5.10).** SvelteKit hashes only the inline
scripts it generates and never parses `app.html`, so a hand-added theme
script would be absent from SvelteKit's own `<meta http-equiv>` policy.
The meta policy and the Go header intersect, so the browser blocks the
script even when the header allows it, and every load flashes the wrong
theme. `%sveltekit.nonce%` fails the build under prerendering (kit issue
10510) and no `%sveltekit.hash%` placeholder exists. The cookie route
removes the hash, the sync burden, and one bump-rehearsal item, and DL
R4.5 already rewrites the document per request.

**The theme element (W5.9).** cairn scopes its whole sheet under a theme
root so it can embed inside a host page, and `ADS:70-75` records the
consequence: every scoped rule is a descendant selector, so a class on
the theme element never matches. dubplate owns the document. Moving
`data-theme` off the document element would make the branding
stylesheet's `:root` declarations lose to inherited-then-shadowed values
rather than compete at equal specificity, killing DL's tiers 2 and 3.
daisyUI's own `root:` config key defaults to `:root` for the same reason.

**DaisyUI 5's rename (W13.1, W13.9).** The last DaisyUI major renamed
every theme variable (`--p`, `--b1`, `--b2`, `--b3`, `--bc`,
`--rounded-btn`, `--border-btn`), removed `--animation-btn`,
`--animation-input`, and `--btn-focus-scale`, introduced `--size-field`,
`--size-selector`, and `--border`, and moved configuration from
`tailwind.config.js` to the CSS `@plugin` directive. Class removals in
the same major were comparatively few (`form-control`, `btn-group`,
`input-group`, `card-compact`, `btm-nav`, and the menu and avatar
modifier prefixes). So "variables are the stable contract" is the
opposite of what the evidence shows.

**Table display properties (W4.5).** Chrome 80 stopped dropping table
semantics for `display: flex`, `grid`, `inline-block`, and `contents`;
Firefox fixed `display: block` by Firefox 62; WebKit fixed its table
display bugs in Safari 17. `display: contents` on a table or its parts is
the case still broken across engines. Roselli, "Tables, CSS Display
Properties, and ARIA".

**SC 2.5.8's exceptions (W4.10).** Inline, Spacing, Equivalent, User
Agent Control, and Essential, quoted in full at the W3C's Understanding
page. The criterion is a square-fits test, not an area test, and carries
no viewport qualifier.

**The live-region nonce (W7.6).** The canonical field report is WordPress
core ticket 36853, which settled on U+00A0 because "while other invisible
characters have been tested, the no-break space looked safer", and which
records clear-then-set as the more reliable mechanism across NVDA, JAWS,
and VoiceOver. No vendor or W3C source states U+200B's behavior in all
three, and U+200B reaches a braille display as a character in some
stacks.

**The guarded control (W7.8).** daisyUI's `button.css` applies
`pointer-events: none` (and a 20% text mix that fails W2.7's floor) to
`.btn:is(:disabled, [disabled], [aria-disabled="true"])`. The selector is
`[aria-disabled="true"]`, not `[aria-disabled]`, and the behavior arrived
in 5.6.

**Gate cost (W12.6).** Measured on this workstation: `bash
scripts/check.sh` on master is 2m 04.6s, of which gitleaks is 1m 45s;
`make build` is 5.1s warm; a bare Playwright floor in cairn's install
(launch plus three pages in two themes with empty documents and only the
settle wait) is 9.2s, 12.8s including node start. The real rendered leg
adds per-stop focus comparison, per-element canvas readbacks, and two
viewport resizes per page.

**5c's arithmetic (section 15).** R5 prices 5c at fifteen deliverables
and 4.5M, derived as about 2.5M of fixed prep plus about 0.115M of chain
per deliverable, plus fork 1's two additions and one extra reviewer read.
Revision 1's routing added roughly eleven further items, about 1.3M at
that grain, with no re-derivation. That is why revision 2 cuts the 5c
list to the ratified deliverables.
