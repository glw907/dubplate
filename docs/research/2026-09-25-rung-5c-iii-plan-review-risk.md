# Rung 5c-iii plan review: domain risk lens

Subject: `docs/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`, revision 1
(`fc52232`). Lens: domain risk (security, accessibility, identity conformance, delivery). Read
against VI revision 1.2, the rung 5 spec, W revision 2.1, DL, CLAUDE.md, the 5c-ii plan's
R-5cii14/19/21/22, the tree at HEAD, and the queued 6b-ii and matcher plans.

**Counts: 0 blockers, 8 majors, 16 minors.** Findings are ranked by consequence within each
severity. The plan's security spine is sound: the token path, the site-name escape, the fonts
route's allow-list, and the CSP story all hold up. The majors are two accessibility failures that
the plan's own text would build, one ratified W section the plan never reads, a branding
regression, lane contention, and three gate-shape gaps.

## Checked and sound

- **Ruling 4's WCAG quote is verbatim.** The W3C Understanding document for SC 1.4.11 (WCAG 2.2,
  fetched for this review) reads: "This success criterion does not require that controls have a
  visual boundary indicating the hit area. If a control has visible content (such as text or a
  sufficiently contrasting icon), which helps users identify the presence of the control, then a
  border or other indication of the overall boundary of the hit area is not required, as is
  therefore not subject to non-text contrast requirements." The reading holds for a text-labelled
  `label-night` primary button (5.15:1 text on fill, recomputed). The same document also says
  "the visual focus indicator for a component *must* have sufficient contrast against the
  adjacent background when the component is focused." That sentence is the basis for M1.
- **The CSP needs no change.** `internal/serve/csp.go:32-36` already carries `img-src 'self'
  data:` (DaisyUI's noise texture), `default-src 'self'` covers the same-origin fonts (no
  `font-src`, per DL G6), and `form-action 'self'` covers both POSTs. Neither the toggle's
  attribute and meta writes nor the `<img>` mark adds a source.
- **`Referrer-Policy: no-referrer` is served on every response** (`internal/serve/headers.go:38`),
  so the interstitial's font, CSS, and favicon subresource loads cannot carry `?t=` in a
  `Referer`. The rung 5 spec's consumption residual (`:1071-1077`) stays bounded.
- **Site-name escaping.** `html.EscapeString` into a double-quoted `content` attribute, read back
  as text and escaped again by Svelte. Mutation 1 proves it.
- **The fonts route.** An exact-name allow-list, an explicit `font/woff2`, a strong SHA-256 ETag,
  and a 404 for `FONTS.md`, the licences, directories, and trailing slashes are the right shape.
- **I2's core.** Kit's `$app/navigation` `replaceState` stores `[PAGE_URL_KEY]: page.url.href`
  (`client.js:2573`), so the plan correctly avoids it. Kit's own initial `history.replaceState`
  (`client.js:380-386`) passes `''` as the URL and stores no URL.
- **Citations spot-checked, all resolve:** `web.go:80` (`no-store`), `web.go:297`
  (`checkVariantCSP`), `serve.go:42-49` and `:243`, `document.go:57` and `:229`, `token.go:118`
  and `:172`, `keys.go:23`, `config.go:352`, `admin.go:27`, `g1.ts:10-12`, `theme-bundle.ts:352`,
  `app.css:3`, and VI `:112-121`, `:128-136`, `:141`, `:149-153`, `:199-214`, `:237`, `:304-305`,
  `:320-334`. VI `:141` does say "R5.4's four files" where R5.5 is meant, so the erratum is right.
- **Ruling 6 per-task arithmetic.** 9.2M over 8 tasks is 1.15M per task, and 9.4 × 0.8 = 7.52.
  See m13 for the reservation sum.

## Majors

### M1. The one focus ring fails WCAG 1.4.11 on the `label` header bar

- **Location:** plan `:887-891` (Task 5, the one focus rule) and `:1079-1088` (Task 9, the header
  controls).
- **Defect:** VI puts the signed-in header on neutral: "Neutral is the header bar" (`:109`), and
  the mark sits "on the dark header" (`:160`). On `label`, `--ink-primary` `#9d2f26` against
  neutral `#1f1b16` measures **2.34:1** (computed for this review). With a 2px offset, the gap
  between the ring and the control is also neutral, so the ring's only adjacent color is
  2.34:1. That fails the 3:1 focus-indicator requirement quoted above. VI's "clears the 3:1
  focus-appearance floor on every surface" (`:114`) measured only the three bases, never
  neutral. On `label-night` the same ring measures 6.93:1 and passes. Task 5's acceptance row
  compares the ring's color to `--ink-primary` and would pass on this failure.
- **Proposed fold:** add a rendered row. For each header control under each theme, it asserts
  the ring's contrast against the painted pixels next to it at 3:1 or better. The fix for the
  row itself is the owner fork below.
- **OWNER FORK (identity):** how the header ring clears 3:1.
  - (a) W9.1's halo: "Where a tinted ground shares the ring's hue, a page-colored halo sits
    under the ring." A 2px `base-100` halo between the control and the ink ring keeps one ring
    color. Ink on cream measures above 6:1.
  - (b) A header-scoped ring in `--ink-primary`'s `label-night` value (6.93:1 on `#1f1b16`).
  - (c) A header-scoped ring in the mustard secondary, which measures about 8:1 on the bar.
  - **Recommendation: (a).** It is W's own mechanism, and it keeps VI's "one focus-ring
    color" literally true.

### M2. The toggle as the plan describes it is invisible on `label`

- **Location:** plan `:132` (Ruling 4: "the theme toggle is a ghost `swap` in `base-content`")
  and `:1079-1083`.
- **Defect:** base-content `#1f1b16` on the `label` neutral bar `#1f1b16` is 1.00:1. The sun and
  moon icons are the control's only identification. A "sufficiently contrasting icon" is exactly
  the Understanding document's condition, so this fails SC 1.4.11 and fails at use. Ruling 4's
  argument also leans on this sentence ("the theme toggle is a ghost `swap` in base-content").
- **Proposed fold:** every header foreground (the toggle icons, nav links, sign-out, and the
  wordmark) takes `--color-neutral-content`. The measured pairs are 14.31:1 on `label` and
  15.15:1 on `label-night`. Correct Ruling 4's sentence to read "in `neutral-content`". Add a
  rendered row that asserts the icon's painted contrast against the bar at 3:1 or better.

### M3. W7.15's ratified login-surface obligations are missing from Task 8

- **Location:** plan `:206-207` (Primary inputs list W sections 3, 4, 5, 8, 9, 12, and 15, and
  omits section 7) and `:1022-1042` (Task 8).
- **Defect:** W7.15 (`web-surface-standard.md:1059-1076`) is ratified, and the plan implements
  none of the following:
  - **"Both auth pages carry `noindex, nofollow`."** Neither the plan nor the tree has one
    (grep of `web/src`, `cmd/dubplate-server`, `internal/serve`: zero hits).
  - **Three outcome states "distinguished, never collapsed":** "a send failure (warning), a
    throttle (plain status), and an unexpected failure (error)". The plan collapses these into
    polite versus assertive only.
  - **"A fresh result supersedes a stale one so two alerts never stack."**
  - **"The sent state is a centered moment, not an alert, with an escape hatch back to the form
    for a mistyped address, carrying the address (W6.11)."** W6.11 grounds this in SC 3.3.7
    (Level A).

  W7.14 (the page `<title>` swaps per state) and W7.16 (focus after an async unmount, which
  applies when the form unmounts for the sent state) are missing too.
- **Proposed fold:**
  - Add W section 7 (W6.11, W7.14, W7.15, W7.16) to Primary inputs.
  - Task 8 gains: the robots meta on both auth documents; the three-class outcome mapping
    (`failed` is warning, `rate_limited` is plain status, and `malformed`, `too_large`, and a
    network failure are error); supersession; and the sent moment with its escape hatch
    pre-filled.
  - Rendered rows: the robots meta is present on both documents; the escape hatch returns the
    typed address; after two submits exactly one outcome element exists.
  - The robots meta is a head insertion like R-5ciii5's, so it can ride `app.html` per route
    through `<svelte:head>`. It must stay outside script bodies.
  - This keeps Task 8 at four deliverables if the robots meta joins the login deliverable.

### M4. The `theme-color` grammar refuses branding values that R-5cii14 accepts

- **Location:** plan `:339` (privileged fields: "A value outside the grammar refuses startup")
  and `:519-523` and `:836-839` (R-5ciii5 and its state table).
- **Defect:** R-5cii14 (5c-ii plan `:1101-1106`) admits a `--color-*` branding value as hex with
  3, 4, 6, or 8 digits, or as `oklch(`, `rgb(`, or `hsl(`. R-5ciii5 checks the meta value against
  `^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$` and refuses startup outside it. An operator whose
  `DUBPLATE_BRANDING` sets `--color-neutral: oklch(20% 0 0)` or `#1f1b16ff` today starts a
  server; after Task 4, the same config refuses startup. The deviation from R-5cii14 is not
  declared, and the state table names only "a grammar-valid hex" for branding. That breaks the
  planning-miss rule, which requires each state to name its report.
- **Proposed fold:**
  - Keep the hex grammar for the bundle read, where the value is the build's own.
  - For branding, accept any value R-5cii14 already validated. Its refusal set excludes `"`,
    `<`, `>`, `'`, `;`, and `\`, and HTML `theme-color` takes any CSS color. Escape it with
    `html.EscapeString` regardless.
  - Add the state rows "branding neutral as `oklch(...)`: one meta at it" and "branding neutral
    as 8-digit hex: one meta at it".

### M5. Two passes claim lane 1, and the heavy-lane spillover to 6b-ii is unplanned

- **Location:** plan `:58-59` and `:289` ("This pass runs in lane 1, the main tree"), `:1195-1203`
  (contested surfaces), and the matcher plan
  `docs/plans/2026-09-25-matcher-word-boundary.md:306-310`.
- **Defect:** The matcher pass also runs "The main tree, as lane 1, after 5c-ii closes", and
  CLAUDE.md pins it there ("Corpus and harness work never moves to a worktree"). Neither plan
  names the other or an order. One tree runs one executor, so the second pass to launch is
  blocked. 6b-ii runs in a lane 2 worktree and writes `cmd/dubplate-server`
  (`restore.go`, `main.go`), which is inside 5c-iii's `--scope-changed` trigger set (`:484-485`).
  Once Task 3 lands on master and 6b-ii rebases, every 6b-ii task gate launches Chromium in the
  machine-wide heavy lane. That queue is shared with cairn sessions. Neither 6b-ii's ceiling nor
  its gate note (`:367`) prices this. The contested-surfaces row says "the conductor tells every
  lane", but it names neither pass.
- **Proposed fold:**
  - State the order: the matcher pass first, since it is small and on master's current tree, and
    5c-iii after it closes. Alternatively, move 5c-iii to a worktree, since it is not corpus or
    harness work.
  - Write the lane registry rows for that order.
  - Give 6b-ii's conductor one line: after 5c-iii's Task 3 merges, 6b-ii's gates run heavy and
    need Playwright's browser. Either narrow the rendered leg's scope trigger to `web/`,
    `internal/serve/`, and the named `cmd/dubplate-server` files (`web.go`, `serve.go`,
    `rendered_test.go`), or have 6b-ii re-price its gates.

### M6. Task 5's rendered proofs rest on fixture controls with no durable home

- **Location:** plan `:874-876` and `:908-915`.
- **Defect:** Task 5's focus rows need a `btn-secondary`, `btn-warning`, `btn-neutral`,
  `btn-primary`, and an `input`, "each control present on the page". The default is "to extend
  the stubs with no copy change", and the fallback is a fixture route under `web/src/routes/`.
  Either choice ships test controls in the production bundle: a new route also needs a Go route
  (R-5cii21 exact paths), which makes it a public page. Tasks 8 and 9 then rewrite those stubs.
  The real pages carry no secondary, warning, or neutral button, so Task 5's rows either break
  at Task 8 or lose their subject. Neither outcome is planned.
- **Proposed fold:** the harness injects a closed fixture control set into a served page with
  `page.evaluate`. That page runs under the real bundle stylesheet and the real stamp, and
  injection is not subject to the page's CSP. Each row asserts that the injected controls are
  present by role before tabbing. No markup ships, and the rows survive Tasks 8 and 9 unchanged.

### M7. The touch-target floor contradicts W4.10's ruled tiers (undeclared deviation)

- **Location:** plan `:1138` and `:1143-1144` ("fails by name on ... a 30px button").
- **Defect:** W4.10 (`:453-486`), "Ruled, conductor 2026-09-12": the rendered rule "fails below
  24 on the activation region and reports below 44 as advisory. Geoff decides whether 44 becomes
  an error tier once rung 7's component set exists." A 30px button passes that rule. W4.10 also
  requires SC 2.5.8's five exceptions and the square test, and the plan names neither. Either the
  seed is wrong or the plan promotes 44 to an error tier, which is Geoff's call and is not
  listed among the declared deviations.
- **Proposed fold:** seed a 20px control for the error floor. Carry the five exceptions and the
  square measurement. Add a VI-scoped row: VI `:333` requires "Header nav links and the theme
  toggle meet 44px", and that row is VI conformance, not the general floor.
- **OWNER FORK:**
  - (a) Keep W4.10's tiers, plus VI's 44px row for the header controls.
  - (b) Promote 44 to the error tier now.
  - **Recommendation: (a).** W4.10 already reserves the promotion for rung 7's component set.

### M8. The theme cookie has no lifetime, contrary to W5.10

- **Location:** plan `:338` and `:1084-1085` (`Path=/; Secure; SameSite=Lax`, inherited from
  R-5cii22).
- **Defect:** With no `Max-Age` or `Expires`, the cookie is a session cookie. A dark choice is
  lost at browser close, and the next load stamps the operator default. W5.10 says "the pick is
  lost only with the cookie jar". The toggle is also "a standing override from the first
  explicit choice on". R-5cii22 left the attribute set open, and 5c-iii is the first writer.
- **Proposed fold:** write `Max-Age=31536000`. Task 9's rendered row asserts the attribute
  through the context's `cookies()` API, since `document.cookie` does not expose expiry.

## Minors

- **m1. The swap's focus ring lands on the wrong box.**
  - **Location:** `:887-891` and `:1080-1082`.
  - **Defect:** DaisyUI's swap (`components/swap.css`) keeps the checkbox in the grid as
    `appearance:none; border:none` at its default roughly 13px size. "Every focusable element's
    `:focus-visible` ring" therefore paints around a small invisible box inside the 44px label,
    and the Task 5 row would pass.
  - **Proposed fold:** use W9.1's descendant path, `.swap:has(> input:focus-visible)` rings the
    label. The row asserts the ring's box contains the icon's box.
- **m2. The `unknown` outcome's admin contact is unprivileged-table text with a missing state.**
  - **Location:** `:977-978` and `:336-343`.
  - **Defect:** `admin_contact` (`handlers.go:157-164`) carries a roster display name and
    address into a page any stranger can load. The privileged-fields table omits it, and the
    nil-contact state (`omitempty`, meaning no admin on the roster) has no deck line.
  - **Proposed fold:** add a table row stating that it renders as Svelte text only, never
    `{@html}` and never an unencoded `mailto:`. Add the nil state to the deck coverage test.
    Optionally, add a static gate that refuses `{@html` in `web/src`.
- **m3. The I2 rows check history only at load.**
  - **Location:** `:1045-1050`.
  - **Defect:** After the native `replaceState`, kit's in-memory `page.url` still holds the
    token, and kit's `reset_focus` calls `history.replaceState(history_state, '', url)` with
    kit's URL (`client.js:3257`).
  - **Proposed fold:** re-assert `location.search` and `history.state` after activating the
    skip link and after the consume click. Forbid reading `page.url` on the interstitial.
- **m4. `label-night` placeholders fail 1.4.3.**
  - **Location:** `:902`.
  - **Defect:** DaisyUI's `.5` placeholder opacity measures 4.38:1 on `label-night` base-100. VI
    `:332` fixes only `label`.
  - **Proposed fold:** ship no placeholders on these pages (labels carry the purpose), or add a
    night opacity that clears 4.5:1. Either way, G9 or a rendered row measures it. A VI value
    question goes to Geoff per `:240-241`.
- **m5. Ruling 3's reason overstates VI.**
  - **Location:** `:111-112` ("a ruled state, not a new design").
  - **Defect:** VI `:356` lists "Phone layouts: ... the stacked login" under "Open for rung 7",
    shown but not ruled.
  - **Proposed fold:** reword the reason as "VI's photography section allows the photo to be
    omitted on phones". The ruling stays with Geoff.
- **m6. The deck omits the per-state `<title>` strings.**
  - **Location:** `:974-987`.
  - **Defect:** W7.14 requires the title to swap per state. The stubs' "Log in" and "Confirm
    login" also conflict with VI's "Sign in" register.
  - **Proposed fold:** add every title to the deck and to its coverage test.
- **m7. W section 7 is absent from Primary inputs.**
  - **Location:** `:206`.
  - **Defect:** This is M3's root cause, and W7.1 and W8.4 are cited without it.
  - **Proposed fold:** add it.
- **m8. Fixture rate limits and link counts are unnamed.**
  - **Location:** `:461-477` and `:1132-1135`.
  - **Defect:** Task 10's matrix renders real `failed` and `unknown` outcomes across 5 widths × 2
    schemes from one client IP. Each consumed link is single-use.
  - **Proposed fold:** the fixture names its link-request limits, or those captures fulfil the
    answer. The test mints one link per interstitial capture and asserts the count.
- **m9. W12.1's redirect-trap shape is missing.**
  - **Location:** `:487-493`.
  - **Proposed fold:** add a row that fails when the final URL differs from the requested one.
- **m10. Ruling 4's icon-only row is vacuous in Task 5.**
  - **Location:** `:915`.
  - **Defect:** No icon-only control exists until Task 9.
  - **Proposed fold:** move the row to Task 9, and have it assert the toggle's presence first.
- **m11. The Fontsource decline misreads R5.4.**
  - **Location:** `:366-370`.
  - **Defect:** R5.4 calls `unicode-range` slices "permitted and encouraged". The valid reason is
    R5.5's file budget.
  - **Proposed fold:** drop the "less than the upstream coverage R5.4 requires" clause.
- **m12. 5c-iv executes a plan drafted before 5c-iii's reviews (undeclared deviation).**
  - **Location:** `:78-80`.
  - **Defect:** CLAUDE.md: plans are "just-in-time per lane, never front-loaded". Task 0's re-run
    checks pointers, not design drift that 5c-iii's review fold may cause.
  - **Proposed fold:** list the deviation, and add one verification-lens read of Tasks 7 through
    11 against the 5c-iii close before 5c-iv's Task 0.
- **m13. The Ruling 6 sum is off.**
  - **Location:** `:157` ("about 13.4M").
  - **Defect:** The two projections sum to 13.25M (7.15 + 6.10).
  - **Proposed fold:** correct the figure.
- **m14. The version hash inputs are incomplete.**
  - **Location:** `:555-557`.
  - **Defect:** The hash omits `svelte.config.js`, `vite.config.*`, and `tsconfig.json`, and it
    specifies no file ordering or path normalization. A config-only change would keep a stale
    version name.
  - **Proposed fold:** add the three config files to the hash inputs, and specify a sorted,
    normalized file order.
- **m15. Tasks 1 and 2 on the light lane are unmeasured.**
  - **Location:** `:327-328`.
  - **Defect:** The light lane caps at 2G high and 3G max, and those gates still run the Vite
    build and svelte-check.
  - **Proposed fold:** Task 0 measures the peak memory of the current `web` leg, or Tasks 1 and 2
    run heavy.
- **m16. The non-admin members state is under-specified.**
  - **Location:** `:1096-1097`.
  - **Defect:** The row checks for no roster. It does not check that the add form is hidden, and
    it does not say how the page renders "the error page's not-found sentence": a `+page.ts`
    `error(404)` or in-page copy.
  - **Proposed fold:** name the mechanism. Assert that no `form` and no roster element render.

## Deviations the plan takes

| Deviation | Declared? |
| --- | --- |
| DL R5.5: two families becomes four files with no family count | Yes (`:176-181`) |
| R-5cii8's members row: `no-cache` becomes `no-store` | Yes (`:182-184`) |
| W4.1's both-schemes line binds now | Yes (`:185-187`) |
| Rung 5 spec's add-member admin checkbox | Yes (`:188-192`) |
| VI ruling 5 (login photo) and error-page photos deferred | Yes, as owner Ruling 3 |
| VI "Depth and surface ... belong to pass 5c-iii" (sleeve, grain, rim, marker) deferred | Yes, as owner Ruling 5. It defers nothing this pass renders, so it is sound |
| W4.10's 24px error / 44px advisory tiers becoming a 44px error floor | **No** (M7) |
| W7.15, W7.14, W6.11 login-surface obligations | **No, omitted** (M3) |
| R-5cii14's branding grammar narrowed for `--color-neutral` | **No** (M4) |
| W5.10's persistent pick | **No** (M8, inherited from R-5cii22) |
| CLAUDE.md's just-in-time plan rule for 5c-iv | **No** (m12) |

Rulings 3 and 5 defer nothing that VI requires of a surface this pass builds, except the login
photo. That deferral is correctly an owner question. The error-page photo rides the same ruling.
Every other VI item this pass's pages use (the focus rule, inks, printed surface, badges, label
rule, selection, reduced motion, 44px header targets, curly quotes, the mark, the favicon, and
`theme-color`) is planned. M1 and M2 are the places where building the plan's text literally
would miss VI's intent.

## Delivery notes (no finding)

- Playwright stays workstation-only: `devDependencies`, the browser in `~/.cache/ms-playwright`,
  and `deploy.sh` building with `make build-linux` (`scripts/deploy.sh:338`), which never runs
  the rendered leg. Ruling 2's evidence (license, engines, cadence, downloads) meets CLAUDE.md's
  adoption bar.
- Ruling 1's cut after Task 5 is at a genuine Files seam, and each half stays within four
  deliverables per task.
