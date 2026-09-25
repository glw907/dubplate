# Rung 5, Pass 5c-iii (Fonts, the Identity's Component Layer, the Rendered Gates, and the Three Real Pages) Implementation Plan

Status: **revision 4, 2026-09-25: reviewed; awaiting owner approval (Rulings 1-10).** The
three `spec-plan-review` lenses (contract, mechanics, risk; `f0e3c00`) are folded, one
disposition per finding, in `docs/superpowers/research/2026-09-25-rung-5c-iii-plan-fold.md`.
The verification read (`2026-09-25-rung-5c-iii-plan-fold-verification.md`, `bbba7a8`) is
folded in the same record's "Fold 2" section. The prose review
(`2026-09-25-rung-5c-iii-plan-prose-review.md`) is folded in its "Prose fold" section. No
ruling below is answered.

Sources. Code pointers were resolved at `b9e281b`.

- The 5c-ii close (`docs/HISTORY.md`, top entry, drafted in
  `docs/superpowers/research/handoffs/draft-HISTORY-5c-ii.md`). Its "What stayed open",
  "Declared coverage gaps", and "The 5c-iii handoff" sections are this pass's scope.
- The 5c-ii plan (`docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`,
  revision 2.4). This pass builds on its R-5cii2 (the relative base), R-5cii7 (the CSP
  builder), R-5cii8 (cache directives), R-5cii15 (where each DL gate lands), R-5cii19 (the
  members route is public, the API is the boundary), R-5cii21 (exact paths), and R-5cii22 (the
  theme stamp and cookie).
- The three 5c-ii reviews
  (`docs/superpowers/research/2026-09-24-rung-5c-ii-review-{svelte,security,a11y}.md`).
- The visual identity spec (`docs/superpowers/specs/2026-09-24-visual-identity-design.md`,
  revision 1.2, `782d864`, VI below), the look every page builds to.
- The design-language standard (`docs/superpowers/specs/2026-09-07-design-language-standard.md`,
  DL below).
- The web surface standard (`docs/superpowers/specs/2026-09-12-web-surface-standard.md`,
  revision 2.1, W below).
- The rung 5 spec (`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`): "The pages"
  (`:1413-1466`), "Consumption" (`:1051-1080`), and "The post-login destination"
  (`:1268-1273`).
- The 5c-i plan's page outline
  (`docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md:2494-2511`).
- The 2.0 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`): "Forward
  hooks" (`:450`) and "Prior art" (`:522`).

**Owner rulings in force.** dubplate carries its own visual identity and matches no streamer
(Geoff, 2026-09-24): no reference screenshots, no visual-fidelity chain. VI is the look (Geoff,
2026-09-25, CLAUDE.md's web stack rule): every `web/` task reads it, and it outranks any styling
habit. Every user-path page is rendered and read before Geoff sees it
(`see-it-before-geoff-tests-it`).

**What this pass is.** 5c-ii shipped the delivery path with three structural stubs. This pass
makes the stubs real and proves them in a browser. It lands six things:

1. The three ruled faces as embedded woff2 files, with G4 and G5.
2. VI revision 1.2's theme data (five inks, the surface tokens), with per-surface contrast
   rows.
3. The rendered leg of the gate: Playwright against the real Go shell, carrying G7's computed
   half, G10, and the first-paint proof per preference and cookie.
4. The document head: the `theme-color` metas, the site name, the robots meta on the two auth
   documents, the mark's SVG files, and the favicon.
5. VI's component layer: the one focus rule, the ink utilities, the printed surface, and the
   refinement rules.
6. 5c-iv's work: the string deck with its register gate and W12.3's AI-tells rule; the login
   page, the consumption interstitial, and the add-member page with the signed-in header and
   its theme toggle; and the render-and-read at W4.1's five viewports in both schemes.

**What this pass is not.** No period photograph and no photo pipeline (Ruling 3). No album
surface, so no sleeve, sleeve grain, record rim, no-art sleeve, or Permanent Marker face
(Ruling 5). No grammar layer, component set, styleguide, or W12 rule beyond the ones named
here (W section 15). No `doctor` row, env template row, runbook row, or `ReadWritePaths`:
those are 5d's. No edit to `internal/config`, `internal/auth`, `internal/identity`,
`internal/engine`, or `cmd/dubplate-server/{check,doctor,store,backup}.go`. **No deploy of any
kind.** A task that believes it needs any of these stops and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process", the pass-execute chain
> per task, sequential, in the segments "Execution model" names. The implementer is
> `dubplate-implementer`. The conductor consumes structured reports and never reads a diff, a
> test log, or a gate transcript inline. Subagents never dispatch agents, so `code-simplifier`,
> the register gate, the render-and-read reader, and the pass-end reviewers are the
> conductor's dispatches. The conductor owns `docs/STATUS.md`, `docs/HISTORY.md`,
> `ROADMAP.md`, the lane registry, and the contested-surface rows. **This pass runs in lane 1,
> the main tree.** The live-executor check is the lane registry.

---

## Rulings awaiting owner

Ten yes/no questions, each with a recommendation and what each answer builds. Rulings 1
through 6 are revision 1's; Rulings 7 through 10 are forks the three plan-review lenses raised
(fold record, `docs/superpowers/research/2026-09-25-rung-5c-iii-plan-fold.md`). The rung 5
spec's rulings, the 5c-i and 5c-ii plans' rulings, VI's owner rulings, and W's section 16 stand
and are not reopened.

**1. Does the remaining 5c scope split again, into 5c-iii (Tasks 1, 2, 3a, 3b, 4, 5, and the
close, Task 6) and 5c-iv (Tasks 7 through 11), under ceilings of 10.8M and 8.5M?**
Recommendation: yes.

**Reason:** the scope is ten build tasks plus two closes. The cut sits after Task 5, where the
Files change from shared infrastructure (fonts, theme data, the rendered harness, the document
head, the component CSS) to the pages that consume it. The seam is sequential, not
file-disjoint: four files cross it, and Task 6's handoff names them. Each half gets its own
pass-end review: the infrastructure is read before three pages build on it, the argument 5c-ii's
own split made. Neither half exports ahead of a caller: every Task 1 through 5 surface is
exercised by the rendered harness, against the 5c-ii stubs or against controls the harness
injects into a hydrated stub (Task 5). The site name is the exception: its first reader is Task
8's login heading, so Task 4 proves the meta by Go rows alone and Task 8 carries the rendered
heading row. The `Mark` component is built with its first caller, Task 8's login page.

**What yes builds:** Tasks 1 through 6 as 5c-iii, closed by Task 6. 5c-iv executes Tasks 7
through 11 of this same plan after a just-in-time refresh at 5c-iii's close (one
verification-lens read of Tasks 7 through 11 against the 5c-iii close, and its fold) and then
its own Task 0. This is a declared deviation from CLAUDE.md's just-in-time rule (below). The
rung 5 spec's 5c row and W section 15 take a third split erratum at 5c-iii's close. **What no
builds:** one pass of twelve tasks under a 17.4M ceiling (flag 13.92M); Task 6 becomes a
STATUS checkpoint with no reviews, and the pass-end reviews run once, after Task 10.

**2. Does `playwright` join `web/package.json` as an exact `devDependency` at 1.63.0?**
Recommendation: yes.

**Reason:** W12.4 names Playwright as the rendered driver, and W12.6 rules that "Playwright
drives its own downloaded Chromium rather than the workstation's layered `chromium-browser`".
5c-ii's owner-touchpoint list makes any web dependency beyond R-5cii16's pins a question. The
evidence was measured on 2026-09-25. `npm view playwright version` answers `1.63.0`, published
2026-09-04, and 1.60 through 1.63 landed monthly. `playwright-core@1.63.0`'s `package.json`
declares `"license": "Apache-2.0"` and `"engines": {"node": ">=20"}`. Its `browsers.json` pins
`chromium` revision `1243` (`browserVersion` `153.0.8010.12`). Both `chromium-1243` and
`chromium_headless_shell-1243` are already present in `~/.cache/ms-playwright/`. The package has
no install scripts, so `npm ci` downloads no browser. npm counted 69,104,951 downloads in the
week ending 2026-09-21. The harness launches with `channel: 'chromium'` (R-5ciii4), which runs
Playwright's own downloaded Chromium, so W12.6 holds. The package is the library only;
`@playwright/test`'s runner is declined because the harness is one script driven from a Go test
(R-5ciii4), and a second runner is a second config to keep in step.

**What yes builds:** Task 3a as written. **What no builds:** the rendered leg cannot exist as a
gate, since W12.1 forbids a skipped check; G7's computed half, G10, the first-paint proof,
and the render-and-read carry forward again.

**3. Does the login page ship without the period photograph, as a single column (mark,
wordmark, the site name as heading, one email field, the sign-in button), until the photo
curation pass supplies licensed photos?** Recommendation: yes.

**Reason:** VI ruling 5 puts a rotating photograph on the login page, but VI's "Photography"
section scopes curation as "a separate pass with no fixed cap", and no photo, manifest, or
CREDITS file exists. Shipping a placeholder photo would ship an unlicensed or unrepresentative
image. VI's login description allows the photo to be omitted on a phone ("the photo appears as
a shorter band above it or is omitted", VI `:260-261`), so the single column reuses a state VI
already names. VI still lists "the stacked login" under "Open for rung 7" (`:356`), shown but
not ruled, so the desktop single column is a new layout that this ruling adopts. Error pages
keep VI's plain sentence without the closed-club photo, for the same reason.

**What yes builds:** Task 8's login page as the single column, with the two-column split
routed to the curation pass. **What no builds:** 5c-iv blocks on the curation pass, or Geoff
names one shipped photo and its license for Task 8 to carry with its film-edge caption.

**4. Does VI's reading of WCAG 1.4.11 stand: a `label-night` primary button on `base-200` or
`base-300` is identified by its label text, so it takes no `--ink-primary` keyline?**
Recommendation: yes; the quoted W3C text settles it, and VI asked for the ruling.

**Reason:** VI (`:117-121`) asked the 5c-ii accessibility review to confirm or refute the
reading; the review measured the fills (2.94:1 on `base-200`, 2.43:1 on `base-300`) and left
the reading open. The W3C's Understanding document for SC 1.4.11 (WCAG 2.2, fetched
2026-09-25) settles it: "This success criterion does not require that controls have a visual
boundary indicating the hit area. If a control has visible content (such as text or a
sufficiently contrasting icon), which helps users identify the presence of the control, then a
border or other indication of the overall boundary of the hit area is not required, as is
[sic] therefore not subject to non-text contrast requirements." A primary button's text measures
5.15:1 on the fill (VI `:119`). The reading fails only for an icon-only primary control, and
this pass ships none: the theme toggle is a ghost `swap` whose icons take
`--color-neutral-content` on the neutral header bar (14.31:1 on `label`, 15.15:1 on
`label-night`; Task 9). VI's refinement row still gives color-filled objects a black keyline
on `label-night` (`:330`) as a design choice, not a conformance fix.

**What yes builds:** no keyline beyond VI's refinement row; Task 9's rendered row asserts the
toggle is present and that no icon-only control takes a primary fill. **What no builds:** a
2px `--ink-primary` keyline on every `label-night` primary fill over `base-200` and
`base-300` (VI `:120-121`). It lands in Task 5's printed surface, read on the injected
`btn-primary` over each of the two grounds, and Task 8 adds a row reading it on the login
page's primary button when that button sits on either ground.

**5. Do the sleeve grain, the record rim, the no-art die-cut sleeve, and the Permanent Marker
face land with rung 7's album surfaces rather than in this pass?** Recommendation: yes.

**Reason:** VI's "Depth and surface" says its rules "belong to pass 5c-iii", but the grain and
the rim decorate album sleeves and the record inside them, and the marker face serves only the
no-art sleeve (VI `:149-153`, `:187-192`). No 5c-iii or 5c-iv surface renders an album, so
the CSS would ship with no instance to render and read, and CLAUDE.md's 2026-09-12 refinement
builds a surface with its first consumer. The mark carries its own rim inside its SVG, so the
mark needs no component rule. One fact for rung 7: Permanent Marker is **Apache-2.0**, not OFL
as VI `:149` states (`google/fonts` files it under `apache/permanentmarker/` with a
`LICENSE.txt`, and `@fontsource/permanent-marker@5.3.0` declares `Apache-2.0`); DL R5.2
admits Apache 2.0, so the face stays eligible and VI takes an erratum.

**What yes builds:** this plan as written, with the four items and VI's license erratum
carried to rung 7's plan. **What no builds:** Task 5 gains a fifth deliverable (the sleeve
component CSS and the fourth woff2 file) with a fixture-only render, and a pre-named split.

**6. Is the 5c reservation raised: 5c-iii and 5c-iv together project about 14.8M of execution
(8.30M and 6.50M) against 5c-ii's provisional 5.0M for this scope?** Recommendation: yes.

**Reason:** the ceiling re-derives from the last actual, as 5c-ii's did (5c-ii plan, "Token
ceiling"). 5c-ii projected 7.0M and closed at about 9.2M over eight tasks, about 1.15M per
task once its review folds (about 1.18M), simplifier, close, and architecture reads are
amortized ("Token ceiling"). 5c-ii's provisional 5.0M assumed six tasks at 0.725M (4.35M) plus
adders. This scope is twelve tasks with the first
browser tooling, after the fold split Task 3 and added 5c-iv's refresh read. Rung 5's
reservation moves from about 41M to about 51M.

**What yes builds:** the ceilings below. **What no builds:** the conductor cuts scope to fit
the old figure; the render-and-read (Task 10) and the AI-tells rule would be the first
candidates, and each is a ratified obligation, so the cut takes an erratum.

**7. Does the signed-in header's focus ring take W9.1's page-colored halo (a `base-100` halo
under the `--ink-primary` outline, spreading past the ring), so one ring color clears 3:1 on
the neutral bar and VI's "one focus-ring color" (`:5`) stays literally true?**
Recommendation: yes.

**Reason:** VI puts the header on neutral ("Neutral is the header bar", `:109`). On `label`,
`--ink-primary` `#9d2f26` against neutral `#1f1b16` measures 2.34:1 (risk lens), and with a
2px offset the gap between control and ring is also neutral, so the ring fails the 3:1 focus
floor the W3C Understanding document for SC 1.4.11 states ("the visual focus indicator for a
component must have sufficient contrast against the adjacent background when the component is
focused"). VI's "clears the 3:1 focus-appearance floor on every surface" (`:114`) measured the
three bases only. On `label-night` the same ring measures 6.93:1. W9.1 prescribes a
page-colored halo where a ground shares the ring's hue ("Where a tinted ground shares the
ring's hue, a page-colored halo sits under the ring"). The bar fails on luminance instead, and
the same halo fixes it.

**What yes builds:** Task 9 scopes the halo to the header controls' `:focus-visible` state as
`box-shadow: 0 0 0 6px var(--color-base-100)` under the 2px outline at its 2px offset. The
halo fills the offset gap, lies under the ring, and extends 2px past the ring's outer edge, so
`base-100` is the ring's neighbor on both sides and the bar never touches it. Task 5's focus
row samples one pixel just inside the ring and one just outside it, both at 3:1 or better.
Under yes, both header samples read `base-100`: 6.12:1 on `label` (`#9d2f26` on `#f4ead3`)
and 6.34:1 on `label-night` (`#ec7b6a` on `#1c1916`). Without the halo the outer sample reads
the bar at 2.34:1, so the header rows fail. **What no builds:** a header-scoped ring in
`--ink-primary`'s `label-night` value `#ec7b6a` under both themes, with no halo. The header
redefines `--ink-primary` in its own scope, so the inherited row's color assertion holds
unchanged (every header foreground takes `--color-neutral-content`, so no header text reads
the ink). Both samples then read the bar: 6.20:1 on `label`'s `#1f1b16` and 6.93:1 on
`label-night`'s `#110f0d`. The header-scoped value lives in `web/src/lib/themes/inks.css`,
since G1 refuses a literal outside `lib/themes/` (`web/scripts/gates/g1.ts:11`), and Task 9's
Files gain that file. VI takes an erratum: the ring color is one per surface class, not one
everywhere.

**8. Do W4.10's two tiers stand in this pass (the rendered rule fails below 24px on the
activation region and reports below 44px as advisory), with VI's "Header nav links and the theme
toggle meet 44px" (`:333`) as its own error row for the header controls?** Recommendation: yes.

**Reason:** W4.10 is ruled ("conductor 2026-09-12") with the promotion reserved: "Geoff decides
whether 44 becomes an error tier once rung 7's component set exists." Revision 1's seed (a
30px button failing) promoted 44 to an error floor without declaring it. DL R3.7 already
carries 44 on the structural axis through `--size-field`, so a plain `.btn` renders at 45px
(VI `:64-65`).

**What yes builds:** Task 10's touch-target rule at W4.10's tiers with SC 2.5.8's five
exceptions and the square test, seeded with a 20px control; Task 9's 44px row on the header
controls. **What no builds:** 44 becomes the error tier now, a declared deviation from W4.10's
ruled tiers, and every control on every page must clear it.

**9. Does a WOFF2 that drops the source's 8-byte empty `DSIG` stub and sets `head` flags bit 11,
and so rewrites `checkSumAdjustment`, count as "unchanged except for WOFF compression" (OFL FAQ
2.2.1) for Alfa Slab One's Reserved Font Name?** Recommendation: yes.

**Reason:** the mechanics lens measured it (`mech/ft/cmp.py`): fontTools' WOFF2 encoder drops
`DSIG` unconditionally (`ttLib/woff2.py:218-219`: "always drop DSIG table, since the encoding
process can invalidate it") and sets `head` bit 11 (`:302-305`: "Set bit 11 of 'head' table
flags to indicate that the font has undergone a lossless modifying transform"), which also
changes `checkSumAdjustment`. The encoder makes both edits, not the converter, and both source
`DSIG` tables are 8-byte stubs that carry no signature. Every other table round-trips
byte-identically under `--no-glyf-transform`. Work Sans carries no Reserved Font Name in its
`OFL.txt` copyright line, so the question binds Alfa Slab One alone.

**What yes builds:** Task 1 ships the converted Alfa Slab One under its own name, with the
comparison output and this ruling recorded in `FONTS.md`. **What no builds:** Task 1 ships a
WOFF2 the upstream project itself publishes if Task 0 finds one at a pinned commit; else Alfa
Slab One ships under a renamed family, as the OFL requires of a modified font carrying a
Reserved Font Name. Under the rename, Task 1's comparison admits the `name` table as a fourth
expected difference and `FONTS.md` records the rename command; any other difference still
stops the task.

**10. When `DUBPLATE_BRANDING` sets `--color-neutral` to a functional color (`oklch()`,
`rgb()`, `hsl()`), does the document carry no `theme-color` meta, with one startup log line
naming the reason?** Recommendation: yes.

**Reason:** R-5cii14 (5c-ii Ruling 7) admits a branding color as hex of 3, 4, 6, or 8 digits or
as `oklch(`, `rgb(`, or `hsl(` (`internal/serve/branding.go:73-75`, `:165`). Revision 1's hex
grammar refused startup on the functional forms, which would stop a configuration 5c-ii
accepts. Omitting the meta fails soft on a cosmetic attribute, keeps every prior configuration
booting, and puts no operator value into an attribute outside a hex grammar. A branding neutral
in any R-5cii14 hex form enters the meta under both answers (R-5ciii5).

**What yes builds:** R-5ciii5's functional-color state row: no `theme-color` meta in any
variant, one `slog` line at startup, and the toggle leaves the absent meta absent. **What no
builds:** the meta carries the functional value, since HTML `theme-color` takes a CSS color and
the branding grammar already refuses `"`, `<`, `>`, `'`, `;`, and `\`; the value is escaped
with `html.EscapeString` regardless, and the toggle writes the branding value.

---

### Spec deviations that take effect on approval

Approving this plan approves these. The close drafts each erratum (Tasks 6 and 11).

- **DL R5.5 ("at most two families and at most four woff2 files").** VI's type section is
  Geoff's ruling and names three families (Alfa Slab One, Work Sans, IBM Plex Mono) plus a
  fourth face for one use (Permanent Marker), four files. R5.5's family count was a conductor
  ruling (2026-09-07), so VI governs: R5.5 reads "four woff2 files", with no family count.
  G4's face-count row checks four files. VI `:141` and `:152` cite R5.4 for the file budget;
  the clause is R5.5.
- **R-5cii8's members row.** The 5c-ii plan's table reads `private, no-cache`; Fold B's L4
  moved the members document to `no-store` (`cmd/dubplate-server/web.go:80`). The code is
  right; the plan row is superseded.
- **W4.1's "The both-schemes half binds from rung 7, because R5's R36 gives rung 5 one composed
  theme."** Rung 5 now ships two composed themes (VI ruling 8), so the render-and-read
  (Task 10) covers both schemes.
- **The rung 5 spec's add-member "admin checkbox" (`:1438`).** R-5c13 (5c-i) already refuses an
  `admin` field on the web channel and the API enforces it (`internal/auth/admin.go:43`,
  `outcomeAdminNotGrantable`); the page ships no admin control. This is the 5c-i outline's
  reading (`2026-09-23-rung-5c-pages-member-verb.md:2501-2503`), recorded here so the spec
  line takes its erratum.
- **R-5cii22's cookie attribute set.** The 5c-ii plan names `Path=/; Secure; SameSite=Lax`
  and no lifetime, which makes the pick a session cookie. W5.10 says "the pick is lost only
  with the cookie jar", so the toggle writes `Max-Age=31536000` as well (R-5ciii9). R-5cii22
  takes an erratum; the server's reading of the cookie is unchanged.
- **VI `:128-131`'s `theme-color` under a functional-color branding neutral.** Under Ruling
  10's yes, such a deployment carries no `theme-color` meta. Under its no, the meta carries the
  functional value. Either way VI's hex pair describes the unbranded bundle only.
- **CLAUDE.md's "plans are just-in-time per lane, never front-loaded", for 5c-iv.** Under
  Ruling 1's yes, 5c-iv executes Tasks 7 through 11 of this plan rather than a plan drafted
  after 5c-iii. The mitigation is the refresh at 5c-iii's close: one verification-lens read
  of Tasks 7 through 11 against the 5c-iii close and its review fold, folded before 5c-iv's
  Task 0 (conductor's action 5).

The risk lens also listed W4.10's 44px floor, W7.14 through W7.16 and W6.11, and R-5cii14's
grammar as undeclared. Revision 2 conforms to W4.10 unless Ruling 8 answers no, implements
W6.11 and W7.14 through W7.16 in Tasks 7 and 8, and routes R-5cii14 to Ruling 10.

---

## Primary inputs

Read these before Task 1, in this order.

1. **VI in full.** Every `web/` task reads it before writing, and the pass-end reviewers grade
   against it.
2. **The 5c-ii HISTORY entry's handoff**, and the three 5c-ii reviews. Their "checked and
   sound" lists are this pass's do-not-re-litigate record.
3. **The 5c-ii plan's R-5cii2, R-5cii7, R-5cii8, R-5cii14, R-5cii19, R-5cii21, R-5cii22.**
4. **DL** sections 5 and 6 (R5.1 through R5.6; G1 through G10) and R3.4, R3.5, R3.7, R4.1,
   R4.6.
5. **W** sections 3 (W3.3, W3.6), 4 (W4.1, W4.10), 5 (W5.9, W5.10), 6 (W6.11), 7 (W7.1, W7.2,
   W7.14, W7.15, W7.16), 8 (W8.4), 9 (W9.1), 12 (W12.1, W12.3's AI-tells row, W12.4, W12.5,
   W12.6), and 15.
6. **The rung 5 spec's "The pages" and "Consumption".**
7. **This plan's quoted behaviors.** Every relied-on library, browser, stdlib, or license
   behavior below is quoted from its source, a `go doc`, or a scratch probe run on 2026-09-25
   against `go1.27.1 linux/amd64`, Node v24.20.0, `@sveltejs/kit@2.70.3`, `daisyui@5.7.46`,
   `playwright@1.63.0` with Chromium 153.0.8010.12, and fontTools 4.66.0. The probes lived
   under
   `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/bc83700f-945c-4755-8eac-4c458ffb90b6/scratchpad/`
   (`pw/`, `webp/`, `fonts/`, and the mechanics lens's `mech/`) and may be gone; the lens
   reports at `docs/superpowers/research/2026-09-25-rung-5c-iii-plan-review-*.md` carry their
   outputs. Each task re-proves what it relies on through its own gate-resident test, and a
   task that relies on a behavior this plan does not quote confirms it the same way and
   reports what it read.

One tree fact every `web/` task needs: `web/src/routes/+layout.ts` sets `prerender = true` and
`ssr = false`, so a prerendered document carries no page markup and no `<svelte:head>` content
until hydration (the built `web/build/login.html` holds no `<title>`). Anything that must be in
the served bytes before a script runs (a meta, an icon link) is a Go head insertion
(R-5ciii5), never `<svelte:head>`.

---

## Goal

After 5c-iii, the served bundle paints in dubplate's own type and component layer, from
embedded faces a gate proves are present, pinned, licensed, and served; every served document
carries the `theme-color` metas, the site name, and the icon link its stamp calls for; and a
rendered leg in the gate proves, in a real browser against the real Go shell, that tier 2 wins
in the computed value, that the match layer's structural values hold under every composed
identity, and that first paint agrees with the stamp under each emulated preference and each
cookie. After 5c-iv, the login page, the consumption interstitial, and the add-member page are
real, their strings live in one deck that passed the register gate and the AI-tells rule, the
interstitial spends the token only on a click and drops it from the address bar, and every
page has been rendered at W4.1's five viewports in both schemes and read before any attended
check.

---

## Owner touchpoints

The ten rulings above, at the plan-approval gate. The render-and-read's screenshots go to
Geoff with the 5c-iv close as the first attended look at the pages, never before the fresh
read in Task 10 has passed. What would become a touchpoint: any web dependency beyond
R-5cii16's pins and Ruling 2's; any Go module dependency (this pass adds none); a VI value a
task believes wrong; a copy line the register gate cannot settle; any change to what the box
runs.

---

## Token ceiling

**Basis.** 5c-ii's recorded actual (`docs/HISTORY.md`, top entry, "Both budgets"): about 9.2M
across eight tasks, which amortizes Task 0, the fix rounds, the two review-fold runs (about
0.6M and 0.58M), the simplifier (about 0.11M), the close drafts (about 0.17M), and three
architecture reads (about 0.33M). That is **1.15M per task**. It excludes the conductor's own
turns and the planning, recorded against the rung.

**5c-iii (Tasks 1, 2, 3a, 3b, 4, 5, 6).**

1. Seven tasks at 1.15M is **8.05M**. Revision 1's Task 3 took its pre-named split in the fold
   (the reproducibility gate and the row manifest grew its harness deliverable), so G7 and G10
   are Task 3b.
2. The first browser tooling: a Chromium launch inside every full gate from Task 3a on, first
   screenshots an implementer reads, and a heavy gate lane: add **0.30M**.
3. Two `opus` tasks of seven (Tasks 3a and 4), a lower share than 5c-ii's three of eight: add
   **0.00M**.
4. Three `go-architecture-reader` dispatches (`internal/serve`, `cmd/dubplate-server`, `web`)
   against 5c-ii's five: **-0.05M**.
5. **Projection 8.30M.** A **10.8M** ceiling puts the 80% flag at **8.64M**, 0.34M above the
   projection, so the flag marks an overrun rather than firing on plan. One pre-named split
   (Task 5's, about 1.15M) crosses the flag, and the conductor asks at that
   boundary under the 80% rule.

**5c-iv (Tasks 7 through 11).**

1. Five tasks at 1.15M is **5.75M**, plus 5c-iv's own Task 0 re-run, inside the per-task rate.
2. The just-in-time refresh at 5c-iii's close (one verification-lens read of Tasks 7 through
   11 and a small fold): add **0.25M**.
3. The deck's register gate (one `prose-voice-reviewer` and a small fold): add **0.15M**.
4. The render-and-read's fresh reader over 150 captures, 60 of them at native resolution, plus
   one re-read of the changed cells after its fold: add **0.35M**.
5. **Projection 6.50M.** A **8.5M** ceiling puts the flag at **6.80M**, 0.30M above the
   projection.

**Together:** 14.80M projected, against 5c-ii's provisional 5.0M (Ruling 6).

**Under Ruling 1's "no":** twelve tasks, 13.80M, plus the browser, register-gate, and reader
adders (0.80M) less the architecture reads' saving (0.05M), less the second Task 0 and one
review round (about -1.0M), with no refresh read: about 13.55M, and a 17.4M ceiling (flag
13.92M).

**Checkpoint interval:** STATUS at every segment boundary (after Task 2, after Task 3b, after
Task 5, after 5c-iii's reviews, at each close, and at the same seams in 5c-iv), at most three
tasks apart. Every task reports its spend, its mutation outcomes, and its measured gate wall
time.

**Deliverable count:** forty-five across Tasks 1 through 11 (Task 3a carries three, Task 3b
two, every other task four), plus each Task 0's report. None passes four.

---

## Global constraints

- **Lane 1, the main tree, on master.** No task reads, writes, or runs a command inside
  `.claude/worktrees/`. The matcher word-boundary plan
  (`docs/superpowers/plans/2026-09-25-matcher-word-boundary.md`, "Global constraints") also
  claims lane 1, and CLAUDE.md keeps corpus and harness work in the main tree. This pass
  launches only when the lane registry shows lane 1 free and names the order; the conductor
  writes that order before either launches (recommended: the matcher pass first, since it is
  small and needs the main tree, then this pass).
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`, and no task ssh's to
  the box.
- **Package territory, stated as a wall.** This pass writes: the `web/` tree (`src/`, `static/`,
  `scripts/`, `svelte.config.js`, `package.json`, `package-lock.json`); `internal/serve`
  (`document.go`, `csp.go` only if a directive test needs it, and their tests);
  `cmd/dubplate-server` (`web.go`, `web_test.go`, `serve.go` for the site-name, prefix, and
  route wiring, `serve_test.go`, and the new build-tagged `rendered_test.go`); `web/embed.go`'s
  test file; `Makefile`; `scripts/check.sh`; `CLAUDE.md`'s gate paragraph; `.gitignore`;
  `.gitleaks.toml`, only if a binary face trips it (Task 1). **It writes nothing under
  `internal/config`, `internal/auth`, `internal/identity`, `internal/engine`, `internal/notify`,
  `internal/harness`, `internal/atomicfile`, `systemd/`, `env/`, `config/`, `corpus/`, `tools/`,
  or `_salvage/`,** and no `docs/` file outside the close drafts. A task that believes it needs
  one stops and reports.
- **No new Go module dependency.** `go.mod` and `go.sum` stay byte-identical.
- **The web dependency set is R-5cii16's pins plus `playwright@1.63.0` under Ruling 2.** The
  pins as committed: `daisyui` is `5.7.46` (`web/package.json`), not the 5c-ii plan's `5.7.45`;
  Task 0 records the drift. No ESLint, Prettier, `@playwright/test`, Fontsource, or font package
  enters `package.json`; the IBM Plex file is taken from its npm tarball once, by hand
  (R-5ciii1), and the package is not a dependency.
- **Server-only SvelteKit files are forbidden**, as 5c-ii's check gates.
- **Nothing sensitive in a page component.** The members page component ships in public
  `_app/immutable/` chunks and a client `goto` renders it without the gate (R-5cii19). The
  roster API is the authorization boundary; the page renders nothing it did not fetch.
- **Every page link, fetch, and asset reference composes from the runtime base** (R-5ciii6);
  no root-absolute URL literal ships (S12).
- **Assertion discipline at the repo's bar.** No vacuous assertion, no expectation derived
  from the code under test. Every hard-coded expectation cites VI, DL, W, the rung 5 spec, a
  quoted document, or a measured probe. Every rendered row names the fixture state it needs
  and asserts that state before it asserts the property. Every rendered row is a literal ID in
  the Go-side row manifest (R-5ciii4).
- **Skills before code.** `go-conventions` before any Go; `svelte-conventions`,
  `ts-conventions`, and the `daisyui` skill before any component, TypeScript, or markup. A
  stock DaisyUI component wins unless VI rules otherwise. No process citation in a shipped
  comment. The em dash is banned in Go comments.
- **Universality.** No estate hostname, path, email, group name, or personal name anywhere in
  `web/` or a fixture. Fixture names use `example.test`.
- **The gate is `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`**,
  re-issued on exit 75 until it prints `gate exit:`, **in the default heavy lane for every
  task: never set `CAIRN_GATE_LANE=light`.** From Task 3a on the gate launches a browser;
  Tasks 1 and 2 launch none, but their Vite build, `svelte-check`, `go test ./...`, and
  golangci-lint have no measured fit under the light lane's 3G cap, so they run heavy and
  report their peak memory. Anything reported as evidence runs with `-count=1` and reports
  wall time. Specific files only. Commit footer `Co-Authored-By: Claude Opus 5.5
  <noreply@anthropic.com>` for an `opus` task and the dispatch's footer otherwise.

---

## Privileged fields and their channels

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| The composed `data-theme` stamp | Unchanged from 5c-ii (R-5cii22): the operator keys, then a valid `__Host-dubplate_theme` cookie. **New writer of the cookie:** the toggle (Task 9), in the page, with `Path=/; Secure; SameSite=Lax; Max-Age=31536000` (R-5ciii9) and a value from `label` or `label-night` only. | The server still never writes the cookie and still ignores a value outside the stampable set. The toggle writes no other name and no other value; a rendered row reads the cookie through the context's `cookies()` after each click. |
| The `theme-color` meta values | At startup: the built bundle's own `--color-neutral` per composed theme, validated against `^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`; or, when `DUBPLATE_BRANDING` sets `--color-neutral`, that R-5cii14-validated value if it is hex, else Ruling 10's answer (R-5ciii5). In the page: the toggle, which writes the computed `--color-neutral` after it changes the stamp (Task 9). | Any request component. A bundle value outside the hex grammar refuses startup, naming the theme; a branding value never refuses startup here, since R-5cii14 already validated it. |
| The site name in a document | `DUBPLATE_SITE_NAME` (a required wall key, `internal/config/keys.go:23`, `config.go:352`), read once at serve startup, escaped with `html.EscapeString` into one meta element (R-5ciii5). | Any request component. The client reads it as text, and Svelte's text interpolation escapes it again. |
| The roster's admin contact on the login page | The request-link answer's `admin_contact` (`internal/auth/handlers.go:157-164`, `omitempty`), a roster display name and address shown to anyone who submits an unknown address under the kind answer. | Rendered as Svelte text only: never `{@html}`, never an unencoded `mailto:` href. A nil contact (no admin on the roster) renders its own deck line (Task 7). |
| A session, through the interstitial | The consume POST (`<prefix>/consume-link`, field `token`, `internal/auth/handlers.go:21`), sent only by a click on the interstitial's one control (I2). | A page load, a prefetch, a link scanner's GET: the document's load performs no request that carries the token in its URL or headers (Task 8's rendered row). |
| The roster write | Unchanged: the roster API behind `RequireAdmin`. The page adds a caller, never a channel; it sends `address`, `display_name`, and `streamer_username` only (`admin.go:25-31`). | The API's own refusals (5c-ii Task 1). The page carries no control named `admin`, and a rendered row checks the POST body's key set (Task 9). |
| The rendered harness's sessions | Magic links the harness mints with `auth.MintLink` (`internal/auth/token.go:118`) over a fixture secret, one for the fixture admin and one for the fixture member, each consumed through the real interstitial (R-5ciii4). | The harness adds no bypass, build-tagged or otherwise, to any production path; W11.8's dev fixture stays rung 7's. |

---

## Resolutions this plan takes rather than discovers

A task landing a resolution writes nothing about it in a shipped comment; the close maps each
resolution to its landing commit.

### R-5ciii1: how the faces are obtained, pinned, licensed, and served

**Three files ship in 5c-iii**, one per VI family: Alfa Slab One (one file), Work Sans (one
variable file, `wght` 100 to 900), and IBM Plex Mono at 400 (one file). The fourth R5.5 file,
Permanent Marker, lands with rung 7 (Ruling 5).

**Sources, pinned by commit or version:**

| Face | Upstream file | Pin | License in the source |
| --- | --- | --- | --- |
| Alfa Slab One | `google/fonts` `ofl/alfaslabone/AlfaSlabOne-Regular.ttf` | commit `ae38321d5a5ff3a2125357402dd332998e6bf41b` (the latest touching that directory, 2026-02-26) | `ofl/alfaslabone/OFL.txt`: "Copyright 2016 The Alfa Slab One Project Authors ... with Reserved Font Name "Alfa Slab"." |
| Work Sans | `google/fonts` `ofl/worksans/WorkSans[wght].ttf` | commit `8b0a1d0f5983c89bc2b93f1b5fb55f9e252744b5` (2026-03-12); upstream `weiweihuanghuang/Work-Sans` at `b35c810...` per `upstream_info.md` | `ofl/worksans/OFL.txt`: "Copyright 2019 The Work Sans Project Authors" |
| IBM Plex Mono | `@ibm/plex-mono@2.5.0` tarball, `package/fonts/complete/woff2/IBMPlexMono-Regular.woff2` (49,248 bytes) | npm version `2.5.0` | `package/LICENSE.txt`: "Copyright © 2017 IBM Corp. with Reserved Font Name "Plex"" |

Fontsource (`@fontsource/alfa-slab-one@5.3.0`, `@fontsource-variable/work-sans@5.3.0`) is
declined: each ships its face as `latin`, `latin-ext`, and `vietnamese` slices (measured from
the tarballs), three files per face, which breaks R5.5's four-file budget. R5.4 calls
`unicode-range` slices permitted, so the budget is the whole reason. The Google Fonts CSS API
is declined for the same slicing and because R5.1 forbids a hosted service.

**Conversion, and what lossless means here.** Two sources are TTF; IBM ships Plex Mono as
WOFF2 already, so it is taken as is. The OFL FAQ (openfontlicense.org, fetched 2026-09-25),
item 2.2: "A change in font format normally is considered modification, and Reserved Font
Names (RFNs) cannot be used. Because of the design of the WOFF and WOFF2 formats, however, it
is possible to create a WOFF/WOFF2 version that is not considered modification". Item 2.2.1:
"only if the original font data remains unchanged except for WOFF compression, and
WOFF-specific metadata is either omitted altogether or present and includes, unaltered, the
contents of all equivalent metadata in the original font." Alfa Slab One carries a Reserved
Font Name, so a lossy conversion would force a rename. The conversion runs once on the
workstation, `uvx --from 'fonttools[woff]==4.66.0' fonttools ttLib.woff2 compress
--no-glyf-transform -o <out>.woff2 <in>.ttf` (fontTools 4.66.0 is the newest on PyPI, MIT;
the `--help` text reads "`--no-glyf-transform` Do not transform glyf (and loca) tables").

The encoder itself makes two edits no flag avoids, measured by the mechanics lens
(`mech/ft/cmp.py`, both faces): it drops `DSIG` (fontTools `ttLib/woff2.py:218-219`: "always
drop DSIG table, since the encoding process can invalidate it"), and it sets `head` flags bit
11 (`:302-305`: "Set bit 11 of 'head' table flags to indicate that the font has undergone a
lossless modifying transform"), which also changes `checkSumAdjustment`. The measured round
trip: `only in src: ['DSIG']`, `differing tables: ['head']`, `head` byte offsets `[8, 9, 10,
11, 16]`, flags `0x3` to `0x803`. Whether those two edits keep Alfa Slab One's name is Ruling
9. **The task proves losslessness in that sense**: `fonttools ttLib.woff2 decompress` of each
output, compared table by table with the source TTF, shows every table except `DSIG` and
`head` byte-identical; `DSIG` absent from the output and 8 bytes long in the source; and
`head` differing only at bytes 8 through 11 (`checkSumAdjustment`) and in flags bit 11. The
report pastes the comparison. Measured `--no-glyf-transform` sizes: Alfa Slab One 39,668
bytes, Work Sans 138,000 bytes. No Python file is committed; the command lives in the
manifest.

**Pinned bytes.** `web/static/fonts/FONTS.md` (DL R5.2's manifest) names, per file: family,
font version (the `name` table's ID 5, measured "Version 2.000" for Alfa Slab One and
"Version 2.012" for Work Sans), upstream URL at its pinned commit or version, the license URL,
the conversion command, the comparison's result, Ruling 9's answer for Alfa Slab One, and the
file's SHA-256. G4 recomputes each SHA-256 and fails on a mismatch, so a swapped file cannot
pass. License texts sit verbatim in `web/static/fonts/licenses/`.

**Where they are served.** DL R5.2's placement stands: `web/static/fonts/` copies to
`web/build/fonts/` (measured in the probe build), and a new `fonts` route serves each `.woff2`
under `<prefix>/fonts/`. The `@font-face` `src` is **written root-absolute in the source**,
`url('/fonts/<name>.woff2')`, which is Vite's public-directory form, and the build rewrites it
relative to the built stylesheet: the probe's `url("/fonts/b.woff2")` emerged as
`url(../../../fonts/b.woff2)` in `_app/immutable/assets/0.*.css`, so the reference resolves
under any prefix (S12's font half). A relative source does not work: the mechanics lens's probe
(`mech/wb2`) wrote `url("../static/fonts/b.woff2")`, and the build emitted a second, hashed copy
at `build/_app/immutable/assets/b.Cfh5bZHl.woff2` beside `build/fonts/b.woff2`, with the
stylesheet pointing at the hashed copy. So G5 requires every built `src` to resolve under
`fonts/` to a manifest file, and a face under `_app/immutable/` fails it. The alternative,
fonts under `web/src/lib/fonts/` hashed into `_app/immutable/` on purpose, is declined because
it moves the license texts and the manifest out of the embedded tree and needs a DL R5.2 and
G4 erratum.

**Caching.** R-5cii8 already classes the fonts `no-cache`. `serve.NewAssets` serves through
`http.ServeContent` with a zero `modtime` (`internal/serve/document.go:230`) and no `ETag`, so a
`no-cache` revalidation has nothing to compare and refetches every byte. `go doc
net/http.ServeContent`: "If the caller has set w's ETag header formatted per RFC 7232, section
2.3, ServeContent uses it to handle requests using If-Match, If-None-Match, or If-Range." The
`fonts` route sets a strong ETag from each file's SHA-256, computed once at startup, so a
revalidation answers 304 (the mechanics lens's probe `mech/mime` confirms the 304 with a zero
modtime).

**Content type.** Go's built-in extension table has no `.woff2` entry (`mime/type.go:116-120`
ends its `.w` entries at `.webp`), and the fallback reads host files (`mime/type_unix.go`:
`/etc/mime.types` and others), which differ between the workstation and the box. The sniffer
does know it (`net/http/internal/sniff.go:178`: `&exactSig{[]byte("wOF2"), "font/woff2"}`),
but a host table that maps the extension wrongly beats the sniffer: the mechanics lens's probe
seeded `mime.AddExtensionType(".woff2", "application/x-seeded")` and an unset header came back
`application/x-seeded`. The route sets `Content-Type: font/woff2` explicitly so no host table
decides, and its test seeds that wrong mapping to prove it.

**CSP.** No change: DL R5.6's "No `font-src`" holds, since `font-src` falls back to
`default-src 'self'` and the faces are same-origin.

### R-5ciii2: the three face roles map onto DL's three font names

DL R2.1's structural axis carries `--font-sans`, `--font-serif`, and `--font-mono`, each ending
in a generic family (R2.7, gated by G2). VI's roles map: `--font-sans` is Work Sans (body,
labels, controls, h3 and smaller in semibold); `--font-serif` is Alfa Slab One (h1, h2, the
wordmark; a slab is a serif), falling back through `ui-serif, serif`; `--font-mono` is IBM Plex
Mono (data only). A `.display` class carries the slab. The values change in `dubplate.css`'s
two match blocks and `system.css`'s two branches in one commit, so G2's parity rows hold.

### R-5ciii3: VI revision 1.2's theme data lands as data, with two new layers

- **Five inks** per composed theme and in both `system.css` branches, from VI's ink table
  (`:100-106`): `--ink-primary`, `--ink-accent` (success text uses it; VI "also success"),
  `--ink-info`, `--ink-error`, `--ink-warning`. `label`: `#9d2f26`, `#2d6136`, `#2d5a81`,
  `#a4261d`, `#754f00`. `label-night`: `#ec7b6a`, `#66a46e`, `#6d9cc7`, `#e37866`, `#e5ad33`
  (the info and warning inks, `#6d9cc7` and `#e5ad33`, may be written `var(--color-info)` and
  `var(--color-warning)`, since each equals its fill; W2.7's "The ink defaults to the fill").
- **Surface tokens**, a fourth layer beside the inks and like them outside the DaisyUI blocks
  (so DL R2.1's 21 names hold), in a new `web/src/lib/themes/surface.css` under `@layer base`,
  one rule per composed theme and both `system.css` branches: the shadow color and strengths (VI
  `:204-206`: neutral at 55% and 45% on `label`; black at 65% and 60% on `label-night`), the
  `label-night` keylines (VI `:330`: `#706a5f` for base-surface objects, black for color-filled
  ones), and the selection ground (VI `:331`). The placeholder opacity (VI `:332`) waits for
  rung 7: no page in this plan carries a `placeholder` attribute. Literal colors live here
  because G1 exempts `lib/themes/` alone (`web/scripts/gates/g1.ts:11`). The implementer names
  the set; the names are a closed list G2 checks per rule.
- **A11y W4 closes as gate rows.** G9 measures a closed list against its grounds. Every ink
  against `base-100`, `base-200`, and `base-300` of its own theme (VI: "every ink passes 4.5:1
  on base-100, base-200, and base-300, so no component needs to know which surface it sits
  on"). `base-content` against the same three bases. Each `*-content` color against its own
  fill only (`primary`, `secondary`, `accent`, `neutral`, `info`, `success`, `warning`,
  `error`; VI: "Every fill and its content color pass 4.5:1 in both palettes"). The fills
  themselves are never measured as text, and the two colors VI marks fill-only (`label` secondary,
  `label-night` neutral, VI `:108`, W2.6's marker) are named in the script as excluded with
  that reason. A second G9 row pins the ten ink values against VI's table, quoted into the
  script with its source line, so an ink that drifts from VI fails even when it still passes
  contrast. That is the row that pins `#9d2f26` and `#754f00`, today unpinned.

### R-5ciii4: the rendered leg is a build-tagged Go test that drives a Playwright script

W12.4: "Driven by Playwright against `dubplate-server serve` already answering". The shape:

- **The server is the real shell, in process.** `cmd/dubplate-server/rendered_test.go`, under
  `//go:build rendered`, builds the shell through `newServeShell` (`serve.go:243`) against a
  fixture environment on `127.0.0.1:0` with a non-default prefix, so every rendered row also
  proves the prefix. It embeds the bundle the `web` target just built, so the stale-binary
  shape of W12.1 cannot arise; the harness still asserts the served `_app/version.json` equals
  `web/build/_app/version.json` byte for byte.
- **The fixture environment, named.** A temp state dir; a fixture roster holding one admin and
  one member at `example.test`; fixed test secrets; an unreachable SMTP port; the kind
  unknown-address answer (`GenericAnswer` false, `cmd/dubplate-server/serve.go:374`, W7.15's
  default); and the four `auth.Limits` counts (`internal/auth/limit.go:37-42`:
  `LinkRequestsPerAddress`, `LinkRequestsPerClient`, `ConsumptionAttemptsPerClient`,
  `LinkMailCeiling`) set through their configuration keys to values the task names. The row
  manifest states how many link requests and consumptions a full run spends, and the Go side
  asserts each count stays below its limit, so no row meets a `rate_limited` it did not ask
  for.
- **The row manifest lives in Go.** `rendered_test.go` holds every rendered row's ID as a
  literal, each derived from a row this plan names. The script reports one structured line
  per row (its ID, the fixture-state assertion's result, the property's result, a detail
  string). The test runs one `t.Run` per manifest ID and fails on an ID the script never
  reported, an ID the manifest does not hold, and a row reported without its fixture-state
  result. A script that ran zero rows fails every manifest ID, so the zero-row refusal is a
  special case. Every later task that adds rows adds their IDs to the manifest.
- **The target proves the test ran.** `go test` answers `ok ... [no tests to run]` and exits 0
  when `-run` matches nothing (measured on this tree:
  `go test -count=1 -run NoSuchTestXyz ./internal/atomicfile` printed that line and exit 0), so
  a misspelled `//go:build rendered`, a renamed test, or a dropped `-tags` would pass
  silently. The `web-rendered` target therefore runs `go test -json` and fails unless the
  stream carries an event whose `Action` is `pass` and whose `Test` is exactly `TestRendered`
  (not a subtest). `go doc cmd/test2json` defines the stream: "a newline-separated sequence of
  TestEvent objects", with `Action` "one of a fixed set of action descriptions", including
  "pass - the test passed" and "skip - the test was skipped or the package contained no
  tests". The target keeps `go test`'s own exit status (the shell's `pipefail`, or the stream
  written to a file first), and the task quotes the event line it matched from its own run.
- **The browser work is one Node script**, `web/scripts/rendered/run.ts`, launched by the test
  under `exec.CommandContext` with the base URL, any minted links, and the request-log
  endpoint in its environment. It launches with `chromium.launch({ channel: 'chromium' })`.
  The default headless launch runs `chromium-headless-shell` (`playwright-core`
  `coreBundle.js:43375`: `return options.headless ? "chromium-headless-shell" : "chromium";`),
  which fetches no favicon at all (mechanics lens, `pw/fav.mjs`), so a favicon row under it
  passes on nothing. `types/types.d.ts` documents the channel: 'Use "chromium" to [opt in to new
  headless mode]'. It is still Playwright's own download (`chromium-1243`, present), so W12.6
  holds; its measured peak was 827 MB against the headless shell's 745 MB.
- **The request log is the Go side's.** Playwright's `page.on('request')` never sees a
  browser-initiated `/favicon.ico` fetch, which reaches only the server (`pw/fav.mjs`: the
  server logged `/favicon.ico`, the page's request events did not). `rendered_test.go` wraps
  the shell's handler in a recorder that logs every request reaching the listener (method,
  path, query, and headers), including any outside the prefix, and serves that log to the
  script from a second loopback listener the test owns. The shell's own handler gains no path.
- **Sessions come from the real login path.** The test mints a link with `auth.MintLink` and
  `auth.LinkURL` (`token.go:118`, `:172`) over the fixture link secret, one per sign-in the
  manifest names, for the fixture admin or the fixture member; the script opens it and clicks
  the interstitial's control, so Task 8's I2 rows and every signed-in row ride the production
  consume path. Before Task 8 lands, the stub interstitial has no control, so Tasks 3a through
  5 run only unauthenticated rows.
- **Server-answer states the fixture cannot produce** are rendered by fulfilling the one
  request with Playwright's `route.fulfill` and the JSON shape `internal/auth/handlers.go:157-164`
  defines; a row that uses a fulfilled answer says so in its ID. Task 8's state table names
  each outcome's source.
- **Where it runs.** A `web-rendered` Makefile target runs `go test -json -tags rendered
  -count=1 -timeout <T> -run TestRendered ./cmd/dubplate-server` under the check above, plus
  `go vet -tags rendered` and `golangci-lint run --build-tags rendered` over that package,
  since the default `vet` and `lint` targets never compile a tagged file. `<T>` is explicit
  because `go test`'s default ten-minute timeout binds at Task 10's matrix; each task that
  grows the matrix sets it to at least twice its measured wall time and reports both. `make
  check` runs the target after `test`. `scripts/check.sh`'s full form always runs it (W12.6:
  the pass-end gate, and "a per-task gate only for a task that moves rendered paint or touches
  `web/`"). `--scope-changed` calls `make check` only when the change set touches
  `internal/harness/`, `internal/engine/`, or `corpus/` (`scripts/check.sh:210-221`).
  Otherwise it runs `make vet lint tidy-check` and `go test` over every package but the
  conform suite (`:296-308`), so Task 3a wires the leg into that branch explicitly: it runs
  when the change set touches `web/`, `internal/serve/`, `internal/auth/`,
  `internal/identity/`, or `cmd/dubplate-server/{web,serve,auth,rendered_test}.go` and their
  tests, and prints the scope line either way. `--reduced` skips it and says so.
- **It gates, never skips (W12.1).** Each of these states exits nonzero with a message naming
  it:
  - The browser does not launch. The script maps a launch error whose message contains
    `Executable doesn't exist at` to the hint `npx --prefix web playwright install chromium`.
    `playwright-core@1.63.0` throws that message from its browser registry
    (`lib/coreBundle.js:33180`). Any other launch error exits nonzero with its own message.
  - The server does not answer.
  - A page answers outside 2xx.
  - A page's final URL differs from the one requested (W12.1's redirect trap).
  - A page settles without its declared identity (its `h1` text) within the timeout: the
    un-booted-page shape. It is judged after hydration, since `ssr = false` leaves the
    prerendered body empty until the client boots.
  - The served `version.json` differs from the build's.
  - Any page raises a `securitypolicyviolation` event or a console error.
  - A manifest ID is missing, extra, or reported without its fixture state.
- **Two refusals stay resident in the gate as negative rows**, so they cannot rot: the test runs
  the script once more with `PLAYWRIGHT_BROWSERS_PATH` set to an empty temporary directory the
  test creates, and asserts a nonzero exit carrying the install hint. The registry reads that
  variable when the module loads (`coreBundle.js:33120`), so the launch takes the same registry
  path a cleared cache or a Playwright bump takes. A scratch probe on 2026-09-25
  (`pw/missing.mjs`, `channel: 'chromium'`) exited 1 with `browserType.launch: Executable
  doesn't exist at <dir>/chromium-1243/chrome-linux64/chrome`. An `executablePath` override
  (`launch()`'s option, `types/types.d.ts:25418-25422`) is not the proof: it throws from
  `resolveExecutablePath` (`coreBundle.js:39930`) as "Failed to launch … because executable
  doesn't exist at …", a different message with no install box. And one row injects an
  off-origin `<img>` into a hydrated page through `page.evaluate` and asserts the violation
  listener records an `img-src` violation. The mechanics lens observed a runtime `img-src`
  violation from an off-origin `<img>` in served markup (`pw/csp.mjs`); the task confirms by
  probe that an injected element's load raises the same event and reports it, and falls back to
  the scratch-build seed if it does not.
- **The containment W12.6 names** is `cairn-run-gate`'s own transient user scope (heavy lane,
  7G high, 8G max), so the leg adds no second `systemd-run`. The mechanics lens measured ten
  full-page screenshots at 320 to 2560 pixels in both schemes at 827 MB peak under
  `channel: 'chromium'`.
- **Process lifecycle.** The mechanics lens's probe (`mech/goexec`) had `exec.CommandContext`
  kill `node` at an 8s deadline (`signal: killed` at 8.02 s), and no Chromium process
  survived under either binary.

**Measured probe (2026-09-25, Chromium 153 through `playwright@1.63.0`):** a page on
`http://127.0.0.1` and on `http://localhost` that sets `document.cookie = '__Host-t=label-night;
Path=/; Secure; SameSite=Lax'` reads it back and the server receives `cookie=[__Host-t=label-night]`
on reload, and `matchMedia('(prefers-color-scheme: dark)').matches` is true under
`newContext({ colorScheme: 'dark' })`. So the harness can exercise the `__Host-` cookie over
loopback http, and the toggle's write works there. The Playwright type docs state
`colorScheme` "supported values are `'light'` and `'dark'` ... Passing `null` resets emulation
to system defaults. Defaults to `'light'`" and that `'no-preference'` "is deprecated"
(`playwright-core@1.63.0` `types/types.d.ts:11330-11336`, `:2786`). A no-preference state
therefore cannot be emulated; R-5cii22's fallback for it is the unconditional light branch,
which the `'light'` row covers. A context with JavaScript disabled cannot write
`document.cookie`, so the first-paint rows seed the cookie with `context.addCookies`: the form
with a `url` field throws for a `__Host-` name ("Protocol error (Storage.setCookies): Invalid
cookie fields"), and `{ name, value, domain: '127.0.0.1', path: '/', secure: true }` works, the
server receiving `cookie=[__Host-dubplate_theme=label]` (mechanics lens, `pw/ac.mjs`).

### R-5ciii5: the document head is the Go rewrite's, per variant

The stamp's variants are built once at startup (`serve.BuildDocumentVariants`,
`internal/serve/document.go:57`, which takes the document's `name`), so the head additions
ride them:

- **`theme-color`.** VI `:128-131`: "`#1f1b16` for light and `#110f0d` for dark, as a `media`
  pair when following the system, and as one value matching the stamped palette when the
  cookie overrides it." The `system` variant carries `<meta name="theme-color"
  media="(prefers-color-scheme: light)" content="…">` and the dark twin; the `label` and
  `label-night` variants carry one meta each. Values come from the bundle's own
  `--color-neutral` per composed theme (the same startup read that finds the theme selectors,
  `cmd/dubplate-server/web.go:473-503`), scoped to the `dubplate-*` blocks, since the built-ins
  emit `oklch(14% .005 285.823)` for theirs; the built bundle emits `#1f1b16` and `#110f0d`,
  twice each (mechanics lens). A bundle value must match `^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`
  before it enters an attribute, and a value outside it refuses startup naming the theme. With
  `DUBPLATE_BRANDING` setting `--color-neutral` (A4: branding flattens both palettes), a value
  in any R-5cii14 hex form (`hexColorRE`, `internal/serve/branding.go:73`: 3, 4, 6, or 8
  digits) puts one meta at it in every variant; a functional value takes Ruling 10's answer.
  The toggle (Task 9) updates the meta in place when it changes the stamp, writing the
  computed `--color-neutral` read after the attribute change, so the value honors branding and
  no color literal enters a component (G1).
- **The site name.** VI ruling 4: "the login heading is the group's name". The name is
  `DUBPLATE_SITE_NAME`, a required wall key. `newServeShell` passes `cfg.SiteName` to the web
  routes, and the rewrite inserts `<meta name="dubplate-site-name" content="…">`, escaped with
  `html.EscapeString`, into every variant. The page reads it at mount; with `ssr = false` the
  prerendered body is empty until hydration, so the heading appears with the rest of the page
  and no request precedes it. A config endpoint (DL R4.3's "config endpoint") is declined
  here: it would be a fetch and a repaint for one string, and no endpoint exists. Go rows
  prove the escape by reading the attribute back: `go doc html.UnescapeString`:
  "UnescapeString(EscapeString(s)) == s always holds".
- **The robots meta (W7.15: "Both auth pages carry `noindex, nofollow`").** The rewrite inserts
  `<meta name="robots" content="noindex, nofollow">` into every variant of the `login` and
  `login/link` documents, chosen by the document name. It is a Go insertion because
  `<svelte:head>` content is absent from the served bytes under `ssr = false`.
- **The icon links.** The rewrite inserts `<link rel="icon" href="<prefix>/favicon.svg"
  type="image/svg+xml">` and the 32px PNG's link into every variant, with the prefix
  `newServeShell` mounts under. A relative link breaks after a client navigation: Chromium
  re-resolves `<link rel="icon">` against the new URL after a same-document navigation, and the
  mechanics lens's probe (`pw/fav2.mjs`) loaded `/p/login/link?t=abc` with
  `href="../favicon.svg"`, stripped the query, pushed `/p/members`, and the server logged a
  request for `/favicon.svg`, outside the prefix. A prefix-absolute link is composed from the
  runtime prefix, so it is S12-clean. `app.html` carries no icon link.
- **Every head insertion is outside every script body**, so 5c-ii's per-variant CSP recheck
  (`checkVariantCSP`, `web.go:297`) keeps passing; a mutation proves the recheck still fires.

### R-5ciii6: every URL composes from the runtime base (S12)

Kit's `resolve` (`@sveltejs/kit/types/index.d.ts:3470-3490`, since 2.26): "Resolve a pathname
by prefixing it with the base path, if any". Its argument type is the app's own page paths
(`.svelte-kit/non-ambient.d.ts:42`: `"/" | "/login" | "/login/link" | "/members"`), so the Go
API paths (`request-link`, `consume-link`, `sign-out`, `roster`,
`cmd/dubplate-server/serve.go:42-49`) go through one helper that joins `resolve('/')` with the
endpoint name. Static files go through kit's `asset` (`:3454-3469`, since 2.26); both prefix
the runtime base, which the bootstrap computes as `new URL(".", location)` or `new URL("..",
location)` (mechanics lens). The icon links are the Go rewrite's (R-5ciii5). A new static gate
(`web/scripts/gates/root-absolute.ts`) fails on any root-absolute `href=`, `src=`, `action=`,
`fetch('/…')`, or `goto('/…')` literal in `web/src`, and on any root-absolute CSS `url(/…)` in
the **built** stylesheets under `web/build/_app/immutable/assets/`. It does not scan CSS
`url()` in `web/src`, because the `@font-face` source must be root-absolute for Vite to rewrite
it (R-5ciii1); a root-absolute URL that survives into the built CSS is the real defect.

### R-5ciii7: builds are reproducible (S11)

Kit (`types/index.d.ts:884-903`): the version name "must be deterministic (e.g. a commit ref
rather than `Math.random()` or `Date.now().toString()`), otherwise defaults to a timestamp of
the build." The timestamp reaches a chunk (the current build's `version.json` value
`1790346892862` appears in `_app/immutable/chunks/Y8ElIxAE.js`), so every build renames chunks
and moves the inline script's hash. `kit.version.name` becomes a SHA-256, computed in
`svelte.config.js` with `node:crypto`, over `web/package-lock.json`, `web/svelte.config.js`,
`web/vite.config.*`, `web/tsconfig.json`, and every file under `web/src/` and `web/static/`.
The walk sorts by relative path, `/`-separated, in byte order, and hashes each path with its
bytes, since `fs.readdir` promises no order and packaging on another filesystem is the reason
for the design. Kit's own example (`git rev-parse HEAD`) is declined because a source tarball
without `.git` would fail to build, and packaging is a goal (VI "Licensing"). The mechanics
lens measured the premise: with `kit.version.name` unset two builds differ (`e56a08…` against
`f665c6…`), and with it fixed they match (`a31d77…` twice).

**It gates.** The `web` target hashes the `web/build` tree (sorted paths and bytes), builds
again, and fails unless the second hash equals the first. The contract lens estimates about 7s
from 5c-ii's build times; Task 3a reports the measured cost. The rendered leg's
served-against-built `version.json` row stays, since it catches a stale embed, a different
defect.

### R-5ciii8: G7's computed half and G10, defined against the fixtures they need

- **G7, computed half** (DL: "a branding variable set in tier 2 wins in the computed value
  against the active theme's own declaration of the same name"). Two fixture servers: one with
  `DUBPLATE_BRANDING` setting `--color-primary` to a value no theme declares, one without. With
  branding, the computed `--color-primary` on the document element equals the branding value
  under each stamp (`dubplate-label`, `dubplate-label-night`, `dubplate-system` under both
  emulated preferences). Without branding, it equals each theme's own value, which proves the
  first row reads a real override rather than a default. A third fixture adds an operator file
  declaring the same name, and the operator value wins (tier 3 over tier 2, R4.1). Colors are
  compared by reading back painted pixels (W12.5: "Every color comparison paints the computed
  color on a canvas and reads sRGB bytes back").
- **G10** (DL R3.4). No shipped palette is sourced from a DaisyUI built-in, so R3.4's literal
  subject does not exist. The built-ins stay in the bundle (`app.css:3`, `themes: light
  --default, dark`), which W5.9 keeps on purpose, and `light --default` is the bare `:root`
  block R3.5 names as the silent-inheritance path. G10 therefore reads all eleven structural
  names' computed values under each composed identity and asserts each equals the `dubplate`
  match block's declared value, then stamps the built-in `light` on the same page client-side
  and asserts `--radius-box` changes to `light`'s own value (`dubplate` `0.25rem` against
  `light`'s `.5rem`, mechanics lens), so the row cannot pass on a page where no built-in could
  leak.
- **First paint** (R-5cii22's rendered half): six states, each expectation a literal from VI,
  never read from the page under test.

  | Cookie | Emulation | Expected painted background | Source |
  | --- | --- | --- | --- |
  | none | light | `#f4ead3` | VI `:71`, `label` base-100 |
  | none | dark | `#1c1916` | VI `:85`, `label-night` base-100 |
  | `label` | light | `#f4ead3` | VI `:71` |
  | `label` | dark | `#f4ead3` | VI `:71` |
  | `label-night` | light | `#1c1916` | VI `:85` |
  | `label-night` | dark | `#1c1916` | VI `:85` |

  Each row asserts the stamped attribute in the served bytes first, then samples the document
  element's painted background from a screenshot taken at `domcontentloaded` in a context with
  JavaScript disabled, the cookie seeded by R-5ciii4's `addCookies` form. The mechanics lens
  measured the timing: a JS-disabled screenshot at `domcontentloaded`, against a stylesheet
  delayed 1.5 s, returned the post-CSS pixel `[17,15,13]` at 1583 ms, because Chromium paints no
  frame before render-blocking CSS; `page.evaluate` still works in a JS-disabled context.

### R-5ciii9: the theme cookie persists

W5.10: "the pick is lost only with the cookie jar". R-5cii22 names `Path=/; Secure;
SameSite=Lax` and no lifetime, which would make the pick a session cookie lost at browser
close. The toggle writes `Max-Age=31536000` (one year) with the other three attributes. The
server's reading of the cookie is unchanged. `document.cookie` does not expose expiry, so
Task 9's row reads the cookie's `expires` through the browser context's `cookies()` API.

---

## Review focus

1. **A token in the interstitial's URL** is read once, dropped from the address bar before
   any other work, and sent only on a click. Neither `location.href` nor `history.state`
   carries it afterward, including after the skip link and the consume click, and no
   request's URL or headers carry it. Pinned in Task 8.
2. **A root-absolute URL** in any page, fetch, stylesheet, or icon link under a prefix,
   including after a client navigation. Pinned in Tasks 1 and 4.
3. **A fill-only or low-contrast color as a focus ring** (a11y W5): every tabbable element on
   every served page shows the one `--ink-primary` ring at 3:1 or better against its adjacent
   painted pixels, including on the neutral header bar (Ruling 7). Pinned in Tasks 5 and 9.
4. **Operator text into markup**: the site name and a branding neutral in the head, and the
   roster's admin contact on the login page. Pinned in Tasks 4 and 8.
5. **A rendered gate that passes on nothing**: a test never compiled or selected, an
   un-booted page, a missing browser, a short or empty row manifest, a favicon row under a
   browser that fetches no favicon. Pinned in Task 3a.
6. **The members page rendered to a non-admin** by client navigation: it shows the not-found
   copy, no roster, and no add form. Pinned in Task 9.

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched before any implementer, and again before Task 7 if
Ruling 1 is yes (after the conductor's refresh of Tasks 7 through 11). Reads this plan against
the tree as shipped. **Its output is blockers with evidence, not fixes.** It edits no tracked
file. A moved pointer whose content survives is a note; a pointer whose content is gone is a
blocker.

1. **Lane 1 free, the tree clean** of warm changes the conductor did not author, and the lane
   registry naming this pass's order against the matcher pass. Report HEAD.
2. **Every `file:line` this plan cites** resolves to the content described, including
   `web/src/app.css:3`, `web/src/lib/themes/inks.css:14-24` (two inks today),
   `web/src/routes/+layout.ts` (`prerender = true`, `ssr = false`),
   `web/scripts/gates/g1.ts:11-12`, `web/scripts/gates/theme-bundle.ts:352`,
   `internal/serve/document.go:21,57,211-232`, `internal/serve/csp.go:32-36`,
   `internal/serve/branding.go:73-75,165`,
   `cmd/dubplate-server/web.go:57-80,156-162,297,473-503`,
   `cmd/dubplate-server/serve.go:42-62,243,289-311,374`,
   `internal/auth/handlers.go:20-21,52-62,157-164`, `internal/auth/admin.go:25-46`,
   `internal/auth/limit.go:37-42`, `internal/auth/token.go:55,64,118,172`, `Makefile:42-87`,
   `scripts/check.sh:1-60,209-221,296-313`.
3. **Every pin**: `npm view <pkg> version` for each R-5cii16 package, `playwright`, and
   `@ibm/plex-mono`; PyPI's newest `fonttools`; the `google/fonts` commits R-5ciii1 names still
   carry the named files. Report every newer version; a newer minor or patch is taken through
   the `dependency-upgrade` skill by the conductor before Task 1, a major is a question.
   Report also whether the Alfa Slab One upstream project publishes a WOFF2 at a pinnable
   commit (Ruling 9's "no" answer uses it).
4. **Playwright's browsers** for 1.63.0 (`chromium-1243` and `chromium_headless_shell-1243`)
   are present under `~/.cache/ms-playwright/`; if absent, report it (the install is the
   conductor's action).
5. **The prefix probe**: build, serve under a non-default prefix, and confirm the three stubs
   hydrate (the 5c-ii probe, re-run).
6. **VI is still revision 1.2**, and its ink table, focus rule, and refinement table read as
   R-5ciii3 quotes them.

---

## Segment A (5c-iii)

## Task 1: The faces, their licenses and manifest, G4, and the `fonts` route with G5

Model: `sonnet`. **Independent of Task 2 by files except `web/src/app.css`'s one import line**;
runs sequentially, one tree.

**Files:**
- Create: `web/static/fonts/AlfaSlabOne-Regular.woff2`, `web/static/fonts/WorkSans[wght].woff2`
  (or a bracket-free name the task states), `web/static/fonts/IBMPlexMono-Regular.woff2`,
  `web/static/fonts/licenses/*.txt`, `web/static/fonts/FONTS.md`,
  `web/src/lib/themes/fonts.css` (the `@font-face` rules), `web/scripts/gates/g4.ts`
- Modify: `web/src/app.css` (one import), `Makefile` (G4 in the `web` target),
  `cmd/dubplate-server/web.go` and `web_test.go` (the `fonts` route), `web/embed_test.go`
  (G5), `.gitleaks.toml` only if a binary face trips it (report it)

**Outcomes:**
- **The faces and their record (R-5ciii1):** three files obtained, converted where needed
  with R-5ciii1's comparison passing, license texts verbatim, and `FONTS.md` with every field
  R-5ciii1 names, including Ruling 9's answer.
- **The `@font-face` rules (DL R5.3):** each with `font-display: swap`, a root-absolute
  public-path `src` (`url('/fonts/<name>.woff2')`, which the build rewrites relative; R-5ciii1)
  with `format('woff2')` and no second format, explicit `font-weight` (`100 900` for Work Sans)
  and `font-style`.
- **G4 (DL section 6), extended:** every font file under `web/static/fonts/` ends in `.woff2`
  and no other font format is present, skipping `licenses/` and `FONTS.md`; every `@font-face`
  sets the three descriptors; the file count is within four (the R5.5 deviation); each file's
  SHA-256 equals its manifest row, and every manifest row names an existing file.
- **The `fonts` route and G5:** `<prefix>/fonts/<name>.woff2` answers GET and HEAD for each
  embedded face with `Content-Type: font/woff2`, `Cache-Control: no-cache`, and a strong ETag,
  answering 304 to a matching `If-None-Match`; every other path under it (a directory, a
  licence text, `FONTS.md`, a missing name, a trailing slash) answers the terminator's 404
  (R-5cii21). G5 (a Go test over the built bundle) resolves every `@font-face` `src` in the
  built stylesheet against the embedded tree, asserting a nonzero face count first, and
  requires each to land under `fonts/` on a file the manifest names, so a hashed copy under
  `_app/immutable/` fails it.

**Acceptance criteria:**
- The report pastes the per-table comparison of each converted file against its TTF, in
  R-5ciii1's three-part form (every table but `DSIG` and `head` identical; `DSIG` absent and an
  8-byte stub in the source; `head` differing only at bytes 8 through 11 and flags bit 11), and
  each file's size and SHA-256. A comparison showing any other difference stops the task.
- G4 fails by name on: a `.ttf` added beside the faces; a `@font-face` missing
  `font-display`; a fifth file; a manifest SHA-256 edited by one character. **Fixture state:**
  each seeded in a scratch copy, the real tree passing first.
- G5 fails naming the missing build on the placeholder-only tree, and naming the path when one
  face file is removed from the embedded tree. **Fixture state:** the real built bundle
  carrying three `@font-face` rules, asserted.
- The route rows: each state above, plus a HEAD, a 405 with `Allow: GET, HEAD`, and a 304.
  The `Content-Type` row first seeds `mime.AddExtensionType(".woff2", "application/x-seeded")`
  (R-5ciii1), then asserts `font/woff2`. **Fixture state:** the seeded mapping asserted in
  force (`mime.TypeByExtension(".woff2")` answers the seed) before the request.
- `bash scripts/check.sh` green, heavy lane, with the gate's peak memory reported.

**Named mutations:**
1. Convert Alfa Slab One with the default glyf transform → the comparison reports changed
   `glyf` and `loca` tables beside `head`, and the task does not ship that file. (Work Sans's
   `glyf` and `loca` round-trip identically even under the default transform, per the
   mechanics lens, so this mutation is Alfa Slab One's alone.)
2. Drop the explicit `Content-Type` → the seeded-mapping row reads `application/x-seeded` and
   fails.
3. Serve `FONTS.md` through the route → the 404 row for it fails.
4. Write one `src` relative into `../static/fonts/` → the build emits a hashed copy under
   `_app/immutable/assets/`, and G5 fails naming the `src` that does not land under `fonts/`.

**2.0 hook: neutral.**

**Prior art:** Fontsource and the Google Fonts CSS API declined (R-5ciii1); `@ibm/plex-mono`
adopted as a file source only (OFL-1.1, 107,174 npm downloads in the week ending 2026-09-21,
IBM-maintained); fontTools adopted as a one-time conversion tool (MIT, 4.66.0, 2026-09-23).

**Deliverable count: 4** (the faces with their record, the `@font-face` rules, G4, the route
with G5).

---

## Task 2: VI 1.2's theme data: the font stacks, five inks, the surface tokens, and per-surface G9

Model: `sonnet`. Depends on Task 1 (the family names).

**Files:**
- Create: `web/src/lib/themes/surface.css`
- Modify: `web/src/lib/themes/dubplate.css`, `inks.css`, `system.css`, `web/src/app.css` (one
  import), `web/scripts/gates/g2.ts`, `g9.ts`, `theme-bundle.ts`

**Outcomes:**
- **The stacks (R-5ciii2)** in both match blocks and both `system.css` branches, each ending in
  a generic family.
- **The five inks (R-5ciii3)** per composed theme and branch; `inks.css`'s header cites VI,
  not the draft tokens file.
- **The surface tokens (R-5ciii3)**, a closed set per theme and branch.
- **The gates:** G2's parity rows cover the new inks and every surface token; `theme-bundle.ts`
  checks the composed themes carry every ink and surface name (5c-ii's `INK_NAMES` widened);
  G9 measures each ink and each direct-text color on all three bases of its theme and pins the
  ten ink values to VI's table (a11y W4).

**Acceptance criteria:**
- G9's per-surface rows report every pair on R-5ciii3's closed list, with the worst case for
  each ink matching VI's table column (4.54 / 4.85 and so on) to two decimals, and name the two
  fill-only exclusions with their reason. **Fixture state:** three bases and every fill per
  theme read from the theme files, asserted present, and the measured pair count asserted
  equal to the list's size.
- `dubplate-system`'s parity holds for every new name in both branches.
- `bash scripts/check.sh` green, heavy lane, with the gate's peak memory reported.

**Named mutations:**
1. Set `label`'s `--ink-primary` to `#c1352b` (the fill) → the `base-300` row fails (VI: the
   fill is why the ink exists).
2. Set `label`'s `--ink-warning` to `#744e00` (4.60:1 on `base-300`, so every contrast row
   still passes) → the pinned-value row fails alone.
3. Drop `--ink-info` from `system.css`'s dark branch → the parity row names it.
4. Move a surface token into a `daisyui/theme` block → G2's 21-name count fails.

**2.0 hook: neutral.** **Deliverable count: 4** (the stacks, the inks, the surface tokens, the
gate rows).

---

## Segment B (5c-iii)

## Task 3a: The rendered leg: the harness, the reproducibility gate, and first paint

Model: `opus`. Depends on Task 2. Upshift reason: W12.1's refusal set is exactly the class of
check that passes vacuously when built quickly, and the harness composes the real shell.

**Files:**
- Create: `cmd/dubplate-server/rendered_test.go` (`//go:build rendered`),
  `web/scripts/rendered/run.ts` (and siblings it needs)
- Modify: `web/package.json`, `web/package-lock.json` (Ruling 2), `web/svelte.config.js`
  (R-5ciii7), `Makefile` (`web-rendered`, the second-build comparison in `web`; `check` runs
  `web-rendered`), `scripts/check.sh` (the full form, the `--scope-changed` branch, the
  `--reduced` line, the header comment), `CLAUDE.md`'s gate paragraph ("The rendered (browser)
  gates join this leg at rung 5c-iii" becomes present tense), `.gitignore` (the harness's
  screenshot output directory)

**Outcomes:**
- **The harness (R-5ciii4):** the in-process shell on the named fixture environment, the Go
  row manifest, the `-json` pass check, `channel: 'chromium'`, the Go-side request log, every
  W12.1 refusal including the redirect trap and the two resident negative rows, the explicit
  `-timeout`, and the leg wired into the full form and the `--scope-changed` branch.
- **The reproducibility gate (R-5ciii7):** the deterministic version name over the sorted
  input set, and the `web` target's second-build comparison.
- **The first-paint rows (R-5ciii8)**, the six states against VI's literal hexes, plus a
  no-CSP-violation row over every page that also answers the 5c-ii carry "confirm rendered in
  5c-iii" for DaisyUI's `data:` noise texture (svelte W4, a11y S7).

**Acceptance criteria:**
- Each refusal fails by name, the real tree passing first:
  - The `web-rendered` target with the build tag misspelled in a scratch copy (no
    `TestRendered` pass event).
  - The resident missing-browser row (`PLAYWRIGHT_BROWSERS_PATH` at an empty temporary
    directory).
  - A stub with its `h1` removed in a scratch build (the identity refusal, judged after
    hydration).
  - A served `version.json` that differs from the build's. The seed builds the test binary
    with `go test -c -tags rendered`, edits `web/build/_app/version.json`, then runs the
    binary, since the embed and the file are otherwise the same bytes.
  - The resident injected off-origin `<img>` (a runtime `img-src` violation).
  - A page whose final URL differs from the requested one. The task names a request the shell
    redirects, or seeds one in a scratch copy.
  - A row removed from the script (the manifest names the missing ID).
  - A row added to the script only (the manifest names the extra ID).
  - A row reported without its fixture-state result.
- A deliberately unhashed inline script in a scratch build is a separate row labeled
  "startup refuses": `serve.DocumentCSP` refuses it at `newServeShell` before any page loads
  (`internal/serve/csp.go`: "DocumentCSP hashes every inline script in doc and refuses when
  one is missing from the meta's list"), so it proves the startup check, not the listener.
- First paint: six rows, each asserting the stamped attribute in the served bytes, then the
  painted pixel against R-5ciii8's table.
- The `web` target's second-build comparison passes on the real tree and fails with
  `kit.version.name` unset.
- `--scope-changed` runs the leg for a change set touching `web/` and prints its skip line for
  one touching only `internal/atomicfile/`; the report pastes both scope lines.
- The report pastes the matched `TestRendered` pass event, the manifest's row count, its link
  request and consumption counts against the fixture limits, the gate's wall time before and
  after the leg, the second build's added time, the gate's peak memory, and the `-timeout`
  value.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Leave `kit.version.name` unset → the `web` target's second-build comparison fails.
2. Make the harness skip when the browser is absent → the resident missing-browser row,
   which expects a nonzero exit carrying the install hint, fails.
3. Misspell `//go:build rendered` → the `web-rendered` target finds no `TestRendered` pass
   event and fails.
4. Declare the `system.css` dark branch before the light one → the no-cookie, dark-emulation
   first-paint row reads `#f4ead3` against `#1c1916` and fails.
5. Drop one row from the script → the manifest names its ID as never reported.

**2.0 hook: neutral.** **Prior art:** Playwright adopted (Ruling 2); `@playwright/test`
declined (R-5ciii4); cairn's rendered runner declined for the structural reason W12.8 records.

**Deliverable count: 3** (the harness, the reproducibility gate, first paint). Revision 1's
pre-named split is taken: G7 and G10 are Task 3b.

---

## Task 3b: G7's computed half and G10

Model: `sonnet`. Depends on Task 3a (the harness and its manifest). The fixture servers are
new environments for the harness Task 3a built, not new mechanism.

**Files:**
- Modify: `cmd/dubplate-server/rendered_test.go` (the fixture servers, the manifest IDs),
  `web/scripts/rendered/run.ts` (the rows)

**Outcomes:**
- **G7's computed half (R-5ciii8):** the with-branding, without-branding, and operator-file
  fixture servers, and their rows under every stamp and both emulations.
- **G10 (R-5ciii8):** eleven structural names per composed identity, and the built-in `light`
  control.

**Acceptance criteria:**
- G7's with-branding rows and without-branding rows both run, and the operator row runs.
  **Fixture state:** each server's `DUBPLATE_BRANDING` and operator file asserted from the
  served document's tier links before any color is read.
- G10 reads eleven names per identity and sees `--radius-box` change from `0.25rem` to `.5rem`
  under `light`.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Put `branding.css` before the bundle stylesheet in `tierHrefs` → G7's computed row fails
   while the 5c-ii link-order row also fails (both halves see it).
2. Remove `--radius-box` from the `dubplate-label` match block and from G2's expectation →
   G10 reads `light`'s value through `:root` and fails.

**2.0 hook: neutral.** **Deliverable count: 2** (G7, G10).

---

## Segment C (5c-iii)

## Task 4: The document head: `theme-color`, the site-name and robots metas, the mark and icon links, and the root-absolute gate

Model: `opus`. Depends on Task 3a (its rendered rows use the harness). **Independent of Tasks
3b and 5 by files except the manifest and `run.ts`.** Upshift reason: two operator-supplied
values enter markup (R-5ciii5), the class Task 6 of 5c-ii upshifted for.

**Files:**
- Create: `web/static/favicon.svg`, `web/static/favicon.png` (32px),
  `web/static/mark/mark-{large,medium,small}.svg` (copied from
  `docs/superpowers/specs/2026-09-24-visual-identity/`, unmodified),
  `web/scripts/gates/root-absolute.ts`
- Modify: `internal/serve/document.go` and its test (the head insertions ride
  `BuildDocumentVariants`; today its only insertion is `insertTierLinks`, applied once before
  the per-palette loop, `document.go:57-71`, and `theme-color` differs per palette, so the task
  adds a per-variant insertion inside the loop and widens the signature for the prefix and the
  site name; the robots meta keys on the document names `login.html` and `login/link.html`,
  `web.go:57-58`), `cmd/dubplate-server/web.go`, `web_test.go`, `serve.go` (passing
  `cfg.SiteName` and the mount prefix), `serve_test.go`, `Makefile` (the new gate),
  `cmd/dubplate-server/rendered_test.go` (the manifest IDs), `web/scripts/rendered/run.ts`
  (the head rows). `web/src/app.html` is not modified: it carries no icon link (R-5ciii5).

**Outcomes:**
- **`theme-color` (R-5ciii5)** in every variant, per the stamp and branding states, with
  Ruling 10's answer for a functional branding neutral.
- **The site-name and robots metas (R-5ciii5):** the site name as one escaped meta in every
  variant; `noindex, nofollow` in every variant of the `login` and `login/link` documents only.
- **The mark and icon links (VI "The mark"):** the three SVGs served as static files, the
  favicon `mark-small.svg` with a 32px PNG rendered once by `rsvg-convert` (installed, 2.63.2;
  the command in the task report), the prefix-absolute icon links the Go rewrite inserts, and Go
  routes answering their exact paths `no-cache` with a strong ETag. The `Mark` component that
  picks among the three SVGs is built with its first caller, in Task 8. G1 does not scan
  `web/static/` (`g1.ts:10`), where the mark's fixed brand colors live; the mark SVGs carry no
  `<text>`, font, or external reference (mechanics lens).
- **The root-absolute gate (R-5ciii6)** in the `web` target, over `web/src` markup and
  TypeScript and the built stylesheets.

**States each check meets:**

| Check | State | Report |
| --- | --- | --- |
| Neutral read from the bundle | present, hex | accepted |
| Neutral read from the bundle | absent for a stampable palette | startup refuses, naming the theme |
| Neutral read from the bundle | not matching the hex grammar (an `oklch()` a later minifier emits) | startup refuses, naming the value's theme, not the value |
| Branding sets `--color-neutral` | hex of 3, 4, 6, or 8 digits (R-5cii14's `hexColorRE`) | every variant carries one meta at it; a row each for `#1f1b16` and `#1f1b16ff` |
| Branding sets `--color-neutral` | `oklch(20% 0 0)`, a functional color R-5cii14 admits | Ruling 10: yes, no meta in any variant and one startup log line; no, one meta at the escaped value |
| Site name | plain text | escaped into `content` |
| Site name | carrying `"`, `<`, `&` | escaped; the attribute read back with `html.UnescapeString` equals the input |
| Robots meta | the `login` and `login/link` documents | present once in every variant |
| Robots meta | every other document | absent |

**Acceptance criteria:**
- Go rows for every state above, over the real built bundle, and the CSP recheck still passes
  for every variant. The site name has no rendered row here: its first reader is Task 8's
  login heading, and the stubs render no site name.
- Rendered rows (full Chromium, R-5ciii4): the `system` variant's metas switch with emulation
  (`page.evaluate` reads the matching meta); on each page the Go request log shows at least one
  icon request on `<prefix>/favicon.svg` or `<prefix>/favicon.png`, whichever Chromium picks
  from the two inserted links, and no request outside the prefix; a client-navigation row loads
  the `login/link` stub, injects `<a href="../members">` through `page.evaluate`, clicks it, and
  finds no request outside the prefix. **Fixture state:** for the navigation row, the log holds
  no document request for `members` and `location.pathname` ends in `/members`, proving the
  navigation was client-side.
- The root-absolute gate fails by name on a seeded `href="/members"` and `fetch('/roster')` in
  a scratch copy of `web/src`, and on a seeded `url(/fonts/x.woff2)` in a scratch copy of a
  built stylesheet, and passes the real tree.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Insert the site name unescaped → the `"` row's read-back no longer equals the input.
2. Place the `theme-color` insertion inside the first `<script>` → the CSP recheck refuses
   startup.
3. Emit a media pair on the `label` variant → the cookie-override row finds two metas.
4. Drop the Go icon-link insertion → the request log shows `/favicon.ico` (full Chromium
   requests it when no icon link exists, `pw/fav.mjs`).
5. Write the icon link depth-relative in the `login/link.html` variants (`../favicon.svg`)
   → the client-navigation row logs a request outside the prefix after the navigation to
   `members`.
6. Refuse startup on a functional branding neutral (revision 1's grammar) → the `oklch` row
   fails under either answer to Ruling 10.

**2.0 hook: neutral.** **Deliverable count: 4** (`theme-color`, the site-name and robots
metas, the mark and icon links, the root-absolute gate).

---

## Task 5: VI's component layer: the focus rule, the ink utilities, the printed surface, and the refinement rules

Model: `sonnet`. Depends on Task 3a (its proofs are rendered rows). **Independent of Task 4
by files except the manifest and `run.ts`.**

**Files:**
- Create: `web/src/lib/styles/components.css` (or a name the task states), imported once from
  `app.css`
- Modify: `web/src/app.css`, `cmd/dubplate-server/rendered_test.go` (the manifest IDs),
  `web/scripts/rendered/run.ts` (the component rows)

**Where the proofs run.** No page, route, or stub gains a control for this task: a sampler in
a shipped page is the styleguide this pass rules out, and a fixture route would ship its
chunks in the embedded bundle and need a Go route (R-5cii21). The rows inject a closed fixture
control set (a `btn-primary`, a `btn-secondary`, a `btn-warning`, a `btn-neutral`, a ghost
button, a default button, an `input`, a card, an alert, and a badge, each with the bundle's
own classes) into a hydrated stub through `page.evaluate`, under the real stylesheet and the
real stamp, and assert each present by role before measuring. Nothing ships, and the rows
survive Tasks 8 and 9 unchanged.

**Where the rules sit.** `daisyui@5.7.46` emits its components inside `@layer utilities {
@layer daisyui.l1.l2.l3 { ... } }` (`node_modules/daisyui/components/button.css`, first
bytes). The button's outline color is `outline-color:var(--btn-color,var(--color-base-content))`
and the input's focus is `outline:2px solid var(--input-color)` (`components/input.css`). The
task quotes CSS Cascade 5's rule for declarations placed directly in a parent layer against
those in its nested sublayers, and places VI's overrides where they outrank DaisyUI's sublayer
without being unlayered (tier 3's operator file must keep the last word, DL R4.4).

**Outcomes:**
- **The one focus rule (VI `:112-115`; a11y W5):** every focusable element's `:focus-visible`
  ring is a 2px `--ink-primary` outline at a 2px offset, overriding DaisyUI's per-button
  `--btn-color` outline and the input's `--input-color` outline. W3.6's exception stands:
  `main:focus` suppresses the ring on the skip target alone. No checkbox, toggle, or radio
  takes a fill-only modifier (VI `:115`).
- **The ink utilities (W2.7):** `text-ink-primary`, `-accent`, `-info`, `-error`, `-warning`
  (Tailwind `@utility`), the only way markup reaches a status text color.
- **The printed surface (VI `:199-214`, `:224-229`, `:329-330`):** buttons `2px 2px 0` and
  cards, alerts, and framed lists `3px 3px 0` of the surface tokens' shadow color with no blur;
  a pressed button translates 2px into its shadow; ghost buttons take no border or shadow;
  controls carry a 2px `base-content` border and inputs a faint inset shadow; badges are
  squared outlines in IBM Plex Mono, uppercase, `.08em` tracking; the `label-night` keylines;
  the three radius values.
- **The refinement rules this pass's pages use (VI `:320-334`):** `.display` at line-height 1.1
  with `text-wrap: balance`; `text-wrap: pretty` on paragraphs and alert text; the three mono
  sizes and two trackings as utilities; the label rule under page titles (a 3px line over a 1px
  line); `::selection`; hover states and pointer cursors on header links; motion stops under
  `prefers-reduced-motion: reduce`. Rules for surfaces no page renders (tabs, tracklists, the
  film edge, the catalog line) wait for rung 7, with the rest of Ruling 5's list.

**Acceptance criteria (rendered rows, both schemes):**
- **The focus row covers every tabbable element on every served page**, enumerated by DOM query,
  plus the injected set. For each, tabbing to it paints a ring whose read-back color equals
  `--ink-primary`'s under that theme. The row samples one painted pixel just inside the ring and
  one just outside it, and each measures 3:1 or better against the ring (the Understanding
  document for SC 1.4.11: "the visual focus indicator for a component must have sufficient
  contrast against the adjacent background when the component is focused"). **Fixture state:** a
  nonzero tabbable count asserted per page (the injected set supplies it while the stubs carry
  none), and each injected control present by role. The row is a manifest entry from this task
  on, so Tasks 8 and 9 inherit it over the real pages' controls and the header (Ruling 7).
- A pressed button's box moves 2px down and right; under `reducedMotion: 'reduce'` it does not.
- The surface rows read back literal expectations computed from VI's tokens, never from the
  page: a `label` button's shadow at `2px 2px 0` is neutral `#1f1b16` at 55% composited over
  `base-100` (VI `:204`); a card's and an alert's `3px 3px 0` at 45% (`:205`); on
  `label-night`, black at 65% and 60% (`:206`), a default button's and a card's keyline
  `#706a5f`, and a primary button's and an alert's keyline black (`:330`); a control's 2px
  `base-content` border; a ghost button's absent border and shadow. Each row names the pixel
  it samples and its expected sRGB bytes.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Author the focus rule in `@layer components` → the `btn-secondary` ring reads back the
   mustard fill (1.83:1) and the row fails.
2. Write a status text as `text-warning` → a static row (added to G1's neighbour or the
   root-absolute gate's scanner, the task states which) refuses DaisyUI status text utilities
   outside the ink utilities.
3. Put the shadow color as a literal in `components.css` → G1 fails.
4. Drop the reduced-motion rule → the reduced-motion row fails.

**2.0 hook: neutral.** **Prior art:** VI's refinement specimen CSS
(`docs/superpowers/research/2026-09-24-identity-refinement/specimen-refined.css`) read as the
worked form of each rule, adopted where it uses DaisyUI's own classes, declined where it styles
a surface rung 7 owns.

**Pre-named split:** the refinement rules move to their own task after this one.

**Deliverable count: 4** (the focus rule, the ink utilities, the printed surface, the
refinement rules). Ruling 4's icon-only row moves to Task 9, where the first icon-only
control exists.

---

## Task 6: The 5c-iii close (under Ruling 1's yes)

Model: `sonnet`. Depends on Tasks 0 through 5 (3a and 3b included), the simplifier round,
and the pass-end reviews with their fold.

**Files (drafts only; the conductor folds and commits):**
`docs/superpowers/research/handoffs/draft-HISTORY-5c-iii.md`,
`docs/superpowers/research/handoffs/draft-STATUS-5c-iii.md`

**Outcomes:**
- **The HISTORY draft:** per task with commit ids; the gate evidence, including the rendered
  leg's wall time and the build's peak memory under the heavy lane; the resolutions and
  rulings; the review dispositions; the declared gaps; both budgets; and "what a later pass
  would be wrong to rediscover".
- **The errata owed, paste-ready,** from this plan's deviations and the fold record's "Errata
  owed" list:
  - the third split in the rung 5 spec's 5c row and W section 15;
  - DL R5.5;
  - W4.1's both-schemes line;
  - the rung 5 spec's admin checkbox;
  - VI `:141`'s and `:152`'s clause number;
  - VI `:149`'s Permanent Marker license;
  - VI `:114`'s focus-floor claim per Ruling 7, and under its no, the ring color per surface
    class;
  - VI `:128-131` per Ruling 10;
  - R-5cii8's members row;
  - R-5cii22's cookie lifetime.
- **The 5c-iv handoff:** what Tasks 7 through 10 build on. It names the four files that cross
  the Ruling 1 cut (`cmd/dubplate-server/rendered_test.go`, `web/scripts/rendered/run.ts`,
  `Makefile`, and `web/src/lib/styles/components.css`), each with its state at the close. It
  lists every 5c-iii review fold that touches a surface they consume, as the input to
  conductor's action 5's refresh read.
- **The STATUS draft,** under 60 lines.

**Acceptance criteria:** every count produced by a command the report pastes; every commit id
resolves; `bash scripts/check.sh` green as a formality.

**Deliverable count: 4.**

---

## Segment D (5c-iv)

5c-iv's tasks below take the just-in-time refresh at 5c-iii's close (conductor's action 5)
before 5c-iv's Task 0 runs; the refresh may amend any of them.

## Task 7: The string deck, the AI-tells rule, and the register gate's input

Model: `sonnet`.

**Files:**
- Create: `web/src/lib/copy/deck.ts`, `web/scripts/gates/ai-tells.ts`
- Modify: `Makefile` (the rule in the `web` target)

**Outcomes:**
- **The deck** (rung 5 spec `:1452-1460`): every user-facing string of the three pages, the
  error page, and the signed-in header, as named constants:
  - the login subline ("Sign in with your email. We'll send you a link.", VI `:304-305`, with
    curly apostrophes per VI `:326`);
  - one line per request-link outcome (`sent`, `requested`, `unknown` with its admin contact
    and `unknown` with none, `rate_limited`, `invalid_address`, `too_large`, `malformed`,
    `failed`; `handlers.go:52-62`), the cross-origin 403, and a network failure;
  - the sent moment's text and its escape hatch back to the form (W6.11, W7.15);
  - the interstitial's statement, its control, and one line per consume outcome (`signed_in`,
    `refused`, `rate_limited`, `too_large`, `malformed`, `failed`);
  - the sign-out control and outcome, which never claims revocation (the 5b-ii bound,
    `2026-09-23-rung-5c-pages-member-verb.md:2495-2497`);
  - the add-member form's labels and hints, and one specific line per roster outcome
    (`admin.go:35-46`);
  - the roster list's empty state (W7.1, W7.2) and its plural-aware count (W8.4);
  - the error page's two sentences;
  - the toggle's accessible name, "Dark theme" (VI `:131`);
  - every page `<title>`, one per page and state where the state changes it (W7.14), in VI's
    "Sign in" register, replacing the stubs' "Log in" and "Confirm login".

  The signed-in state is the honest terminal state the spec names for no streamer
  (`spec:1270-1273`): a plain confirmation, no destination.
- **The coverage test:** every outcome value the handlers answer has a deck line, asserted by a
  test that reads the Go outcome constants' values from `internal/auth/handlers.go` and
  `admin.go` and compares them with the deck's keys. So a parser that matches nothing cannot
  pass, the test first asserts the parsed set contains `signed_out`, `admin_not_grantable`, and
  `write_unconfirmed` and holds at least the 23 values the two files declare at `b9e281b` (11
  in `handlers.go:52-62`, 12 in `admin.go:35-46`). Every title and the nil-contact line are
  covered the same way.
- **The AI-tells rule (W12.3; W section 15: "The AI-tells rule also runs inside 5c"):** a port
  of cairn's `scripts/checks/check-admin-prose.mjs` (MIT, 285 lines; the blocking lists
  "ported verbatim from prose-guard") as a gate script scanning every string in `web/src`
  markup and literals, with a `--list` mode. It enforces no register, and it reports and
  asserts a nonzero scanned-string count.
- **The register gate's input:** the deck's strings in one reviewable file, handed to the
  conductor's `prose-voice-reviewer` dispatch against the Microsoft editor standard (the 5c-i
  mail-copy precedent), whose findings a small fold task applies. Copy never names a mechanism
  the member cannot see, never puns (VI ruling 4), and a reject, security, or rate-limit line
  is never playful (W1.4).

**Acceptance criteria:** the coverage test passes and fails by name on a deleted line; the
AI-tells rule fails by name on a seeded "Moreover" and an em dash in a scratch copy;
`bash scripts/check.sh` green, heavy lane.

**Named mutations:** 1. Delete the `rate_limited` line → the coverage row names it. 2. Seed
"delve" in a component string → the AI-tells rule fails. 3. Write "you have been signed out
everywhere" → the register reviewer is asked to catch it; the deck test checks no sign-out
line contains "everywhere" or "revoked". 4. Break the constant parser's pattern → the
named-members assertion fails.

**2.0 hook: neutral.** **Deliverable count: 4** (the deck, the coverage test, the AI-tells
rule, the register input).

---

## Task 8: The login page, the consumption interstitial, and the error page

Model: `sonnet`. Depends on Task 7. **Independent of Task 9 by files except the manifest and
`run.ts`**; Task 9's header reads the `Mark` component this task creates.

**Files:**
- Modify: `web/src/routes/login/+page.svelte`, `web/src/routes/login/link/+page.svelte`,
  `web/src/routes/+error.svelte`, `cmd/dubplate-server/rendered_test.go` (the fixture site
  name, the manifest IDs), `web/scripts/rendered/run.ts`
- Create: a `Mark` component under `web/src/lib/components/`; a small API helper under
  `web/src/lib/` (R-5ciii6) if Task 7 did not

**Outcomes:**
- **The login page** (rung 5 spec `:1415-1433`; VI `:258-261` under Ruling 3; W6.11, W7.14,
  W7.15, W7.16): the `Mark` and the wordmark in Alfa Slab One, the site name as the `h1` over
  the label rule, the subline, a DaisyUI `fieldset` with `legend`, `label`, and `input
  type="email"` with `autocomplete="email"`, `validator` and a space-reserving
  `validator-hint`, no `placeholder` attribute, and one primary submit. Each outcome renders
  per the state table below, a fresh result superseding a stale one so two outcome elements
  never stack. The sent state is a centered moment, not an alert, with an escape hatch back
  to the form carrying the typed address; when it unmounts the focused form, focus moves
  deliberately (W7.16). The `<title>` follows the state (W7.14). The admin contact renders as
  Svelte text only. The server's copy stays authoritative over `validator`. No theme toggle
  (VI `:132`). The `Mark` component, built here for its first caller, picks among Task 4's
  three SVGs by rendered size (56px and up large, 21 to 55 medium, 20 and below small) through
  kit's `asset()`, with empty `alt`; the wordmark text beside it carries the name.
- **The interstitial** (rung 5 spec "Consumption"; security I2): on mount it reads `t` from
  `location`, then **before any other work** replaces the history entry with the same URL less
  its query. Kit's `$app/navigation` `replaceState` stores `page.url.href` (the token URL) in
  the history entry's state (`client.js:2573`, `[PAGE_URL_KEY]: page.url.href`), so it is not
  used for this; the page calls `history.replaceState(history.state, '', <URL without query>)`,
  keeping kit's index keys, and the task quotes kit's dev-only warning (`client.js:90-114`,
  inside `if (DEV && BROWSER)`) to show the production build is unaffected. The page never
  reads kit's `page.url`, which still holds the token after the native call. It states what is
  about to happen and carries one control; only a click POSTs `token` to `consume-link`. It
  loads nothing from a third party (CSP `default-src 'self'` already forbids it). A missing
  token renders the `refused` copy with no request.
- **The error page:** VI's plain sentence in the identity's type and surface, no photo
  (Ruling 3).
- **Every page carries W3.6's skip link and `<main id="main" tabindex="-1">`.**

**The outcome states (W7.15's three classes).** "Real" is produced by the fixture environment
(R-5ciii4: the kind answer, an unreachable SMTP port); "fulfilled" is `route.fulfill` with the
`handlers.go:157-164` shape; the network failure is an aborted request (Playwright's
`route.abort`, which the task quotes from `types/types.d.ts`). Error and warning announce
assertively; plain status announces politely.

| Outcome | Source | Class | Rendering |
| --- | --- | --- | --- |
| `sent` | fulfilled | the sent moment | focus moves to the moment's heading |
| `requested` | fulfilled (the generic answer only) | the sent moment | as `sent` |
| `unknown`, admin contact present | real (an address not on the fixture roster) | plain status | polite, the contact as text |
| `unknown`, no admin contact | fulfilled (`admin_contact` omitted) | plain status | polite, its own deck line |
| `failed` | real (the unreachable SMTP port) | warning | assertive |
| `rate_limited` | fulfilled | plain status | polite |
| `invalid_address` | real if a value `type="email"` admits fails the server's normalization, else fulfilled; the task states which | error | assertive |
| `too_large`, `malformed` | fulfilled | error | assertive |
| cross-origin 403 | fulfilled with status 403 | error | assertive |
| network failure | aborted request | error | assertive |
| `signed_in` | real (a minted link) | plain status | polite |
| `refused` | real (the stripped URL reloaded, or no token) | error | assertive |
| `rate_limited`, `too_large`, `malformed`, `failed` (consume) | fulfilled | plain status for `rate_limited`, error for the rest | as their class |

**Acceptance criteria (rendered, both schemes, under the prefix):**
- I2, in order: after the interstitial loads with a minted token, `location.search` is empty
  and `JSON.stringify(history.state)` does not contain the token, both read before the click
  (the ordering row); the Go request log shows no request to `consume-link` before the click
  and exactly one after, and no logged request carries the token in its URL or headers; both
  history reads hold again after the skip link is activated and after the consume click.
  **Fixture state:** a real minted link for the fixture admin, asserted valid by the row that
  clicks it and receives `signed_in`.
- A reload of the stripped URL shows the `refused` copy and sends nothing.
- One row per state-table line: the rendered text equals the deck constant for that outcome
  (the key read from the deck module, since the property is the mapping), and the live
  region's politeness and the class match the table. After two submits, exactly one outcome
  element exists. The escape hatch returns to the form with the typed address in the field.
  After the sent moment mounts, `document.activeElement` is its heading. The `<title>` equals
  the deck's title for each state.
- The site-name heading: the fixture's `DUBPLATE_SITE_NAME` carries `"`, `<`, and `&` (at
  `example.test` scope, universal), and the `h1` text equals it literally with no element
  created from it. **Fixture state:** the served `dubplate-site-name` meta asserted present
  before the heading is read. The fixture admin's display name carries `<` the same way, and
  the `unknown` row shows it as text.
- The error page: one row per status the client can reach, by injecting a link to an
  unmatched path into a hydrated page and clicking it; the page shows VI's sentence in the
  slab and base surface with the skip link first in tab order.
- One `h1` per page (W12.4's `h1-count`), the skip link first in tab order, and no
  `placeholder` attribute on any input.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Use `$app/navigation`'s `replaceState` → the `history.state` row finds the token.
2. POST on mount → the request log finds a consume before the click.
3. Strip the query after the fetch instead of before → the ordering row finds a non-empty
   `location.search` at its read.
4. Drop `autocomplete="email"` → a static row over the login markup fails.
5. Map every non-`sent` answer to one generic line → the state-table rows for `failed` and
   `rate_limited` fail.
6. Insert the site name with `{@html}` → the heading row finds an element created from it.

**2.0 hook: neutral** (the device page will live in this app, VI and S20 `:390-396`; nothing
here builds for it). **Pre-named split:** the error page and its rows move to their own task
after this one. **Deliverable count: 4** (login with the `Mark` component, interstitial, error
page, the rendered rows).

---

## Task 9: The add-member page, the signed-in header, and the theme toggle

Model: `sonnet`. Depends on Task 7, and on Task 8 for the `Mark` component. **Otherwise
independent of Task 8 by files except the manifest and `run.ts`.**

**Files:**
- Create: a `SiteHeader` component (mark, wordmark, nav, toggle, sign-out) under
  `web/src/lib/components/`
- Modify: `web/src/routes/members/+page.svelte`, `web/src/lib/styles/components.css` (the
  header's focus treatment), `cmd/dubplate-server/rendered_test.go` (the member sign-in, the
  manifest IDs), `web/scripts/rendered/run.ts`, and under Ruling 7's no,
  `web/src/lib/themes/inks.css` (the header-scoped `--ink-primary`)

**Outcomes:**
- **The signed-in header (VI `:131-136`):** a neutral bar with the `Mark` and wordmark, and at
  the right end of the nav the toggle, an icon-only DaisyUI `swap` driven by its hidden
  checkbox (the stock mechanism, `daisyui` skill "swap" rule "Use only a hidden checkbox to
  control the swap state"), sun and moon Phosphor bold icons (VI `:237`), accessible name "Dark
  theme", checked when the stamp resolves dark. Every header foreground (the toggle's icons,
  the nav links, sign-out, and the wordmark) takes `--color-neutral-content`, since
  `base-content` on the `label` bar measures 1.00:1 (risk lens). The focus ring on the bar
  follows Ruling 7. The swap's checkbox stays focusable but DaisyUI's `.swap input` is
  `appearance:none; border:none` at about 7 by 12 px, so the ring moves to the control with
  `.swap:has(> input:focus-visible)` and is suppressed on the input (W9.1's descendant path;
  mechanics lens `pw/swap.mjs`). Choosing a state writes `__Host-dubplate_theme=label|label-night;
  Path=/; Secure; SameSite=Lax; Max-Age=31536000` (R-5ciii9), sets `data-theme` on the document
  element (DL R4.2: "The client changes the attribute after hydration"), and leaves exactly one
  `theme-color` meta, without `media`, whose content is the computed `--color-neutral` (the
  shape the server stamps for that cookie; under Ruling 10's yes, an absent meta stays absent).
  Never a `.theme-controller` (DL R4.6; the theme-bundle gate already refuses the class,
  `theme-bundle.ts:352`). The toggle and each nav link meet 44px (VI `:333`). A sign-out
  control POSTs `sign-out` and returns to the login page.
- **The roster list:** the members the API returns, each with display name, address, and
  streamer username, and a squared-outline badge marking an admin; the plural-aware count; the
  empty state. It stacks at narrow widths (W4.4, W4.6).
- **The add-member form:** `fieldset` of address, display name, and an optional streamer
  username, no `placeholder` attributes; no control named `admin` (the spec deviation above);
  one primary submit; each roster outcome renders its specific deck line, and `added` refreshes
  the list.
- **A non-admin reaching the page by client navigation** (R-5cii19) sees the not-found
  sentence, no roster, and no add form, since the API's GET answers the terminator's 404. The
  task names the mechanism (in-page copy from the deck, or kit's `error(404)` from a
  client-side load, quoting kit's documentation for it).

**Acceptance criteria (rendered, both schemes):**
- The toggle, from four starts, each asserted before the click: `system` under light emulation
  (unchecked; the click writes `label-night`), `system` under dark (checked; writes `label`),
  cookie `label` (unchecked; writes `label-night`), cookie `label-night` (checked; writes
  `label`). After each click, the context's `cookies()` holds exactly the chosen value with an
  `expires` about a year out; the attribute changed; the one `theme-color` meta reads `#1f1b16`
  for `label` or `#110f0d` for `label-night` (VI `:128-131`); and a reload's served bytes carry
  the matching stamp. **Fixture state:** the signed-in admin from a real consumed link.
- The header: each foreground's painted contrast against the bar is 3:1 or better; the toggle
  and each nav link measure 44px or more (VI `:333`, Ruling 8); the focus row (Task 5) covers
  every header control, its ring box contains the toggle's icon box, and its two samples clear
  3:1 in the header under Ruling 7's answer. Ruling 4's row: the toggle is present, and no
  icon-only control carries a primary fill.
- An add with a fresh address shows `added` and the list grows by one; a repeat shows
  `duplicate_address`; the fixture roster file on disk shows the write (real API). The add form
  holds no control named `admin` (DOM query), and the intercepted POST body's key set is a
  subset of `address`, `display_name`, `streamer_username`.
- The non-admin row: sign the fixture member in through a minted link; on the hydrated
  interstitial showing `signed_in`, inject `<a href="../members">` through `page.evaluate` and
  click it. **Fixture state:** the Go request log holds no document request for `members` and
  one roster `GET` answered 404, proving a client navigation. Then the not-found sentence is
  visible, no element contains the admin's fixture address, and no `form`, `fieldset`, or
  roster list renders.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Use `.theme-controller` on the checkbox → the theme-bundle gate fails.
2. Write the cookie without `Secure` → the browser refuses a `__Host-` cookie, and the reload
   row finds the old stamp.
3. Render the roster from a cached copy after a 404 → the non-admin row finds an address.
4. Add an `admin` checkbox → the DOM query for a control named `admin` fails (an unchecked box
   is not submitted, so the API refusal alone would never see it).
5. Drop `Max-Age` → the `cookies()` row finds a session cookie.
6. Set the toggle's icons in `base-content` → the header contrast row fails on `label`.

**2.0 hook: neutral.** Hook 5 (the stable identity) is untouched: the toggle's cookie carries
no identity (W section 15), and `streamer_username` stays optional per R-5c13 until rung 12.
**Pre-named split:** the header and the toggle with their rows move to their own task after
this one. **Deliverable count: 4** (the header with the toggle, the list, the form, the
rendered rows).

---

## Segment E (5c-iv)

## Task 10: The render-and-read

Model: `sonnet` for the capture; the read is the conductor's dispatch.

**Files:**
- Modify: `cmd/dubplate-server/rendered_test.go` (the capture manifest, the link count),
  `web/scripts/rendered/run.ts` (the capture matrix and the two floors)
- Create: `docs/superpowers/research/2026-09-25-rung-5c-iv-render-and-read.md` (the reader's
  findings, committed by the conductor)

**Outcomes:**
- **The capture matrix:** fifteen page states, each at 320, 390, 768, 1440, and 2560 CSS
  pixels (W4.1), in `label` and `label-night`, full page, to the ignored output directory,
  with a contact sheet per page: **150 captures**. The states, as literals in the Go manifest:
  login with the empty form; the sent moment (fulfilled); `unknown` with its admin contact
  (real); `failed` (real); `rate_limited` (fulfilled); an error (`malformed`, fulfilled); the
  interstitial ready (a minted token); the interstitial `refused` (no token); the interstitial
  `signed_in`; members populated (admin); members empty (a fulfilled roster answer); members
  after `added`; members after `duplicate_address`; members as a non-admin (not found); and
  the error page (an unmatched path by client navigation). Every capture asserts its fixture
  state first (the page's `h1` and the state's distinguishing deck text) and fails rather than
  capture an un-booted page, an error, or a spinner. The test mints one link per capture that
  consumes one and asserts the count against the fixture limits.
- **The viewport-overflow floor (W12.4):** at 390 and 320 nothing paints past the viewport,
  gated on the document scrolling.
- **The touch-target floor (W4.10, at Ruling 8's tiers):** the square test on each activation
  region at every width, failing below 24px and reporting below 44px as advisory, with SC
  2.5.8's five exceptions: Inline and Spacing implemented, Equivalent, User Agent Control, and
  Essential as suppressions with a named reason. A nonzero control count is asserted per page.
- **The read:** the conductor dispatches one fresh-context reader (`opus`, no plan in context)
  with VI's path, W section 4, the native-resolution captures at 390 and 1440 in both schemes
  (60 images), and the contact sheets for 320, 768, and 2560, asking for defects ranked by
  consequence and a list of every cell it read. The conductor rejects a report missing a cell.
  A small fold task applies what the conductor accepts, the matrix re-runs, and one fresh read
  of the changed cells precedes Geoff's look.

**Acceptance criteria:** the manifest holds the fifteen states and the run reports 150 captures,
each with its fixture state asserted; the overflow floor fails by name on a seeded 400px-wide
element and the touch floor on a seeded 20px control, each passing the real tree; the reader's
report lists all 60 native cells and all 90 sheet cells (the fifteen states at 320, 768, and
2560 in both schemes), and its verdict is accept or its accepted findings are folded,
re-captured, and re-read; `bash scripts/check.sh` green, heavy lane, with the `-timeout` reset
to at least twice the measured wall time.

**2.0 hook: neutral.** **Deliverable count: 4** (the matrix, the overflow floor, the touch
floor, the read and its fold).

---

## Task 11: The 5c-iv close

Model: `sonnet`. As Task 6, over Tasks 7 through 10: the HISTORY and STATUS drafts, the
errata (the rung 5 spec's 5c row closed; W section 15's rendered and AI-tells lines as
landed; any erratum the refresh read or the reviews added), and the handoff to 5d (the
`deploy.sh` clean-build row now includes the rendered leg's absence on the box, since nothing
Node or Playwright ships) and to rung 7 (Ruling 5's four items, the photo column under Ruling
3, W12.4's remaining rules, W4.10's 44px promotion question, and both palettes' placeholder
opacity: VI `:332`'s 62% on `label`, which no 5c page consumes, and `label-night`'s, which VI
leaves at DaisyUI's default of 4.38:1 on base-100 per the risk lens).
**Deliverable count: 4.**

---

## The exit gate

`bash scripts/check.sh` green on master in the full form, plus each assertion below and its
proving task.

1. Three faces embedded, pinned by SHA-256, licensed, served `font/woff2`; G4 and G5 fail by
   name on their seeds (Task 1).
2. Every ink passes 4.5:1 on all three bases of its theme and equals VI's table; every pair on
   G9's closed list measured (Task 2).
3. The rendered leg runs in every full gate and in `--scope-changed` for its trigger set,
   fails unless `TestRendered` itself passed, holds its row manifest, and refuses each W12.1
   shape (Task 3a).
4. Tier 2 wins in the computed value under every stamp; tier 3 wins over tier 2; the match
   layer's structural values hold under every composed identity (Task 3b).
5. First paint equals VI's base-100 literal in all six states (Task 3a).
6. Two builds of one tree are byte-identical, gated in the `web` target (Task 3a).
7. `theme-color` follows the stamp and branding; the site name is escaped; the auth documents
   carry `noindex, nofollow`; no request leaves the prefix, including after a client
   navigation, and none goes to `/favicon.ico`; no root-absolute URL ships (Task 4).
8. Every tabbable element on every served page shows the 2px `--ink-primary` ring at a 2px
   offset, at 3:1 or better against its adjacent pixels (Tasks 5 and 9).
9. The interstitial drops the token before any request, keeps it out of every request, and
   spends it only on a click (Task 8).
10. The toggle writes a persistent cookie, the meta follows, and the next render stamps it; a
    non-admin sees no roster and no add form (Task 9).
11. Every deck line exists for every outcome and title; the AI-tells rule and the register gate
    pass (Task 7).
12. Every page state rendered at five widths in both schemes, floors green, read and folded
    (Task 10).
13. The 5c-ii carry "the prefix proof's hydration half is evidence from a scratch probe, not a
    gate" is closed: every rendered row runs under a non-default prefix and judges identity
    after hydration (Task 3a).

**Deferred, by name:** the photo column and error-page photos (Ruling 3); the sleeve, grain,
rim, no-art sleeve, and Permanent Marker (Ruling 5); VI's "Open for rung 7" list; W12.4's
remaining rendered rules; W11.8's dev fixture; W4.10's 44px promotion (Ruling 8); both
palettes' placeholder opacity.

---

## Contested surfaces

| Surface | Owner | This pass's position | Contention |
| --- | --- | --- | --- |
| Lane 1 (the main tree) | the lane registry | the whole pass | **the matcher word-boundary plan also claims lane 1**; the conductor writes the order in the registry before either launches |
| `web/` | rung 5 births it; rung 7 authors visuals | every task | none open; rung 7 inherits |
| `cmd/dubplate-server/web.go`, `serve.go`, their tests | rung 5 | Tasks 1, 3a, 4 | 6b-ii's lane 2 writes `cmd/dubplate-server/restore.go` and `main.go`, disjoint files in the same package |
| `cmd/dubplate-server/rendered_test.go` (new) | rung 5 | Task 3a onward | none |
| `internal/serve/document.go` | rung 5 | Task 4 | none |
| `Makefile`, `scripts/check.sh`, `CLAUDE.md` gate paragraph | shared | Tasks 1, 3a, 4, 7 | **after Task 3a lands, every full gate on every lane launches Chromium and needs `npm ci --prefix web` plus Playwright's browsers**, and runs heavy. 6b-ii's `--scope-changed` gates stay browser-free (its files are outside the trigger set), but its merge ritual's full gates do not; the conductor tells 6b-ii's conductor through the registry |
| `go.mod`, `go.sum` | shared | untouched | none |
| `docs/` ledgers | conductor | Tasks 6 and 11 draft only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record planning spend against the rung; write the lane registry's order for
   this pass and the matcher pass, the lane 1 row, and the contested-surface rows; arm the
   runaway guard; AC power; confirm the gate string runs heavy for every task.
2. Before Task 1: rule Task 0's blockers; take any minor or patch pin bump through
   `dependency-upgrade`.
3. Before Task 3a: `npm ci --prefix web` after the lockfile changes; if Task 0 found a browser
   absent, `npx --prefix web playwright install chromium`, recorded in the workstation's
   install manifest per the global rule. After Task 3a merges: one registry line to 6b-ii's
   conductor that its full gates now run heavy and need Playwright's browsers.
4. After Task 5, under Ruling 1's yes (under its no, these reviews run once, at action 8):
   one `code-simplifier` round over Tasks 1 through 5, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, reviewed one `diff-reviewer` per
   package; then `svelte-reviewer` and `daisyui-a11y-reviewer`, each given VI's path and asked
   for conformance to it (tokens, inks, the component layer, the copy register) graded beside
   correctness, and `web-auth-security-reviewer` over the head insertions and the new routes;
   each asked for correctness gaps ranked by consequence; one fold dispatch that may refuse a
   finding whose fix costs more than the risk it removes, with a one-line reason.
5. At Task 6: one `go-architecture-reader` each for `internal/serve`, `cmd/dubplate-server`,
   and `web`, no plan in context. Re-derive 5c-iv's figure from 5c-iii's actual. Then the
   just-in-time refresh: one verification-lens read (`claude-opus-5-5`, `spec-plan-review`'s
   verification lens) of Tasks 7 through 11 against the 5c-iii close, its review fold, and
   the tree, and one fold of what it finds, before 5c-iv's Task 0.
6. After Task 7: the register gate (one `prose-voice-reviewer`, Microsoft editor standard) and
   its fold.
7. At Task 10: the render-and-read reader, its fold, and the fresh read of the changed cells.
8. After Task 10: the simplifier round over Tasks 7 through 10; the three reviewers again,
   `web-auth-security-reviewer` pointed at the interstitial's token handling (I2), the
   toggle's cookie write, and the admin contact's rendering; the fold.
9. At Task 11: the architecture reads for the packages 5c-iv touched; the render-and-read's
   contact sheets go to Geoff with the close, the first attended look.
10. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, one invocation per segment.
Every task runs `dubplate-implementer`, then `diff-reviewer` (`claude-opus-5-5`), then the
gate, with one re-dispatch on `fix` and a second `fix` as the conductor's decision.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1, 2 | STATUS. |
| **B** | 3a, 3b | STATUS. The registry line to 6b-ii after 3a merges. |
| **C** | 4, 5 | STATUS. The simplifier, the three reviewers, the fold. |
| **Close** | 6 | The 5c-iii close; the architecture reads; the ledger fold; the refresh read of Tasks 7 through 11 and its fold. |
| **D** | 7, 8, 9 | 5c-iv's Task 0 first. STATUS. The register gate after Task 7. |
| **E** | 10, 11 | The read, its fold, and the re-read; the simplifier; the reviewers; the close. |

Each boundary sits on a task's own green committed gate.

**Models.** `sonnet` at effort `high` by default. **Two upshifts to `opus`, unconditional**:
Task 3a (the harness's refusal set) and Task 4 (operator text into markup). Task 0 runs at
`opus` in fresh context.

**Dependencies, stated exactly:**
- Task 0 first; its blockers ruled before Task 1.
- Task 1 → Task 2 → Task 3a → Task 3b.
- Task 3a → Task 4 and Task 3a → Task 5. **Tasks 3b, 4, and 5 are independent by files except
  the row manifest and `run.ts`, which each extends**; they run sequentially because one tree
  runs one executor.
- Tasks 1 through 5, the simplifier, and the reviews → Task 6.
- Task 6 → the refresh read → 5c-iv's Task 0 → Task 7.
- Task 7 → Task 8 → Task 9. **Tasks 8 and 9 are independent by files except the manifest,
  `run.ts`, and the `Mark` component Task 8 creates and Task 9's header reads.**
- Tasks 8 and 9 → Task 10 → the reviews → Task 11.

**Where this pass splits further, if it must:** at the Task 4 / Task 5 boundary, if segment
C's re-derivation projects past the flag. Task 5 then opens 5c-iv, since Tasks 8 and 9 are its
first consumers.

**Budget:** 10.8M (5c-iii) and 8.5M (5c-iv) under Ruling 1, flags 8.64M and 6.80M; planning
recorded separately; STATUS at each boundary; spend scored at each close against its ceiling
and against zero attended interactions after approval.
