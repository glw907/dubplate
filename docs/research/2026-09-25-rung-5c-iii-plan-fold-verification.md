# Rung 5c-iii plan fold 1: verification read

Subject: `docs/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`, revision
2 (`0aeb644`), against revision 1 (`fc52232`), the fold record
(`2026-09-25-rung-5c-iii-plan-fold.md`), and the three lens reviews (`-contract`, `-mechanics`,
`-risk`). The reader had no part in the fold. Findings are correctness gaps ranked by
consequence; a finding whose fix costs more than the risk it removes may be refused.

**Verdict: the fold closes. 0 blockers, 2 majors, 9 minors.** Every lens blocker and major
closed at the cited location except one geometry gap in Ruling 7's recommended answer, which
the fold introduced when it turned risk M1 into a ruling. The second major is a new mechanism
(the resident missing-browser row) stated partly from memory: its message differs from the one
the plan maps. Both folds are one or two sentences. No task passes four deliverables. The task
order builds. All ten rulings are rulable; Rulings 4, 7, and 9 each have a "no" build with a
small gap (minors 3, 4, 5).

## Majors

### V-M1. Ruling 7's "yes" names no halo geometry, and the focus row names no sample side

- **Location:** Ruling 7 (`:189-207`); Task 5 acceptance, the focus row (`:1287-1295`); Task 9
  outcomes (`:1543`) and acceptance (`:1574-1577`).
- **Defect:** the ruling's build is "a 2px `base-100` ring under the `--ink-primary` outline".
  The focus rule is a 2px outline at a 2px offset. The natural implementation of a 2px halo,
  `box-shadow: 0 0 0 2px var(--color-base-100)`, fills only the offset gap. That leaves cream
  inside the ink ring and the neutral bar outside it. The row asks for the ring's "painted
  contrast against the pixels adjacent to it", but it does not say which side. Recomputed here:
  `#9d2f26` against `#1f1b16` is 2.34:1, and against `#f4ead3` it is 6.12:1. Sampling the
  outer side makes the recommended answer fail its own row, which costs a Task 9 fix round.
  Sampling the inner side passes wherever a halo exists, so a row that picks the easier side
  proves nothing about the bar. The fold record says "the header rows fail without the ruling's
  treatment". That holds only for the outer side, and there the treatment as worded fails too.
- **Proposed fold:** name the geometry and the sample points in one sentence each.
  (a) Ruling 7's "yes" is a halo that spreads past the ring (for example a 6px `base-100`
  spread beneath the 2px-offset outline), so cream lies on both sides of the ink. (b) The row
  samples one pixel just inside the ring and one just outside, and asserts both at 3:1 or
  better. If Geoff prefers the halo inside the ring only, the ruling states the other WCAG
  reading instead: the halo plus the ring is the indicator, measured as the halo against the
  bar (14.31:1). The row then samples the halo's outer edge against the bar. Either reading
  works if the plan names it.

### V-M2. The resident missing-browser row takes a different code path and message from the refusal it guards

- **Location:** R-5ciii4, "It gates" (`:695-697`) and "Two refusals stay resident"
  (`:704-707`); Task 3a acceptance (`:1078`) and mutation 2 (`:1105-1106`).
- **Defect:** the plan maps the launch error "Executable doesn't exist at …" to the install
  hint. It proves that mapping with an `executablePath` naming no file. In
  `playwright-core@1.63.0`, those two states throw different errors from different functions.
  The scratchpad install at `scratchpad/pw/node_modules/playwright-core` shows both:
  - The real missing browser goes through the registry. `lib/coreBundle.js:33180` throws
    ``Executable doesn't exist at ${e}`` with the install box.
  - The override goes through `resolveExecutablePath`. `:39930` throws ``Failed to launch
    ${this._name} because executable doesn't exist at ${executablePath}``, which has a
    lowercase "executable" and no install box.
  A script that maps the quoted message fails the resident row on the real tree. A script
  widened to map any launch error passes the row without ever exercising the registry path.
  The registry path is the one a Playwright bump or a cleared cache actually hits. The fold
  also quotes `types/types.d.ts:17650-17657` for `executablePath`. That is
  `launchPersistentContext`'s option. `launch()`'s own option is documented at `:25418-25422`
  with the same wording, so the quote is harmless but mis-anchored.
- **Proposed fold:** the resident row runs the script with `PLAYWRIGHT_BROWSERS_PATH` set to an
  empty temporary directory. The registry reads that variable at `coreBundle.js:33120`. The row
  then hits the registry's own "Executable doesn't exist at" throw, the same state a missing
  browser produces. It asserts a nonzero exit that carries the hint. Drop the `executablePath`
  override. Re-anchor the type quote to `:25418-25422`, or drop it.

## Minors

1. **Task 4 mutation 5 names a mutation that does not escape the prefix** (`:1226-1227`).
   A Go-inserted `href="favicon.svg"` on `login/link.html` resolves to
   `<prefix>/login/favicon.svg` at load and to `<prefix>/favicon.svg` after the client
   navigation to `members`. Both stay inside the prefix, so the client-navigation row does not
   fail. The per-page "log shows `<prefix>/favicon.svg`" row does catch the mutation on the
   `login/link` page. The escaping form the mechanics lens probed is the depth-relative
   `../favicon.svg` on a depth-2 document. **Fold:** mutation 5 writes the depth-relative href
   (`../favicon.svg` in `login/link.html`) and names the client-navigation row.
2. **Task 4's favicon row assumes Chromium fetches the SVG when an SVG and a 32px PNG link
   are both present** (`:1208-1209`). The mechanics probe had one SVG link only, and no quote
   says which candidate Chromium picks. **Fold:** the row asserts that at least one icon
   request lands on either inserted prefix-absolute path and that none leaves the prefix, or
   Task 4 probes the selection and records it.
3. **Ruling 7's "no" answer breaks the inherited focus row** (`:204-207`, `:1289`). The row
   asserts the ring "equals `--ink-primary`'s under that theme". A header ring in
   `label-night`'s ink value on `label` fails that assertion unless the header redefines
   `--ink-primary` locally. **Fold:** the "no" build adds "the header scopes `--ink-primary`
   (or the row reads the ring token the header declares)".
4. **Ruling 9's "no" rename path contradicts Task 1's stop rule** (`:237-241`, `:970`).
   Renaming a Reserved Font Name edits the `name` table. Task 1 stops on "any other
   difference", and no conversion tool is committed. **Fold:** under "no" with no upstream
   WOFF2, the comparison admits the `name` table, and the rename command is recorded in
   `FONTS.md`. Alternatively, the rename is a re-plan stop.
5. **Ruling 4's "no" build names no landing task** (`:154-156`). **Fold:** "Task 5's printed
   surface", with a Task 8 row on the login primary.
6. **Task 10's sheet-cell count does not reconcile** (`:1649-1650`). 150 captures less 60
   native leaves 90. "45 sheet cells" holds only if one cell is a state and width pair showing
   both schemes. **Fold:** define the cell, or write 90.
7. **Ruling 1's "no" and conductor's action 4 disagree.** Under "no", the pass-end reviews run
   once, after Task 10 (`:92-93`). Action 4 runs the reviews after Task 5 unconditionally
   (`:1740`). **Fold:** prefix action 4 with "Under Ruling 1's yes".
8. **The split's "Files seam" is sequential, not disjoint** (`:78-79`). `rendered_test.go`,
   `run.ts`, `Makefile`, and `components.css` (created by Task 5, modified by Task 9) cross
   the cut. No race exists, because 5c-iv starts after 5c-iii closes. **Fold:** Task 6's
   handoff names these four files as 5c-iv's inherited surfaces, so the refresh read checks
   them.
9. **Nits, one line each.** 5c-iii's ceiling note cites "Task 5's or Task 8's" split
   (`:384-385`), but Task 8 belongs to 5c-iv. VI citations sit one line off: "Dark theme" is
   VI `:131`, not `:132`, and "The login page has no toggle" is `:132`, not `:133`. Task 0
   catches both. Task 3a is dense for a three-deliverable task, but its count holds. A
   pre-named split that moves the reproducibility gate out would be cheap.

## Question 1: did each blocker and major close where cited?

| Finding | Closed? | Where verified |
| --- | --- | --- |
| Contract B1 and M1 (vacuous pass; the manifest) | Yes. The Go manifest holds literal IDs, one `t.Run` each, and fails on missing, extra, or state-less rows; the target requires a `pass` event whose `Test` is exactly `TestRendered`, and keeps the exit status. | `:639-656`; Task 3a acceptance `:1076-1086`; mutations 3 and 5 |
| Contract M2 (first paint pinned) | Yes. Six literal rows cite VI `:71` and `:85`, both checked against VI; the `addCookies` form and the frame timing are quoted from the mechanics probes. The default `DUBPLATE_THEME_PALETTE` is `system` (`web.go:108`), so the no-cookie rows hold with the fixture at default. | `:842-859`; `:731-735`; Task 3a mutation 4 (checked against `system.css`: light unconditional, dark under `@media`) |
| Contract M3 (Task 4's site-name row) | Yes. Go rows with an `html.UnescapeString` read-back; the rendered heading row and mutation moved to Task 8 with `"`, `<`, `&`. | `:1198-1206`, `:1220`; Task 8 `:1495-1499`, mutation 6 |
| Contract M4 and risk M4 (Ruling 10) | Yes. Hex forms enter under either answer; the functional form has a named report under each answer; the privileged-field row no longer refuses a branding value. | `:243-258`, `:477`, `:1196-1197`, mutation 6 |
| Contract M5 and risk M6 (injected controls) | Yes. A closed injected set, asserted by role; the focus row spans every tabbable element on every page and is inherited by Tasks 8 and 9. Its sample side is open (V-M1). | `:1247-1254`, `:1287-1295` |
| Contract M6 (outcome table) | Yes. Thirteen rows name source, class, and politeness; `GenericAnswer` false is named. | `:1458-1478`; `:632` |
| Contract M7 (non-admin client navigation) | Yes. A minted member sign-in, an injected `../members` link, a Go-log proof of client navigation, and the no-form and no-roster assertions. The 404 is `RequireAdmin`'s `http.NotFound` (`internal/auth/admin.go:128`). | `:1583-1588` |
| Contract M8 (mutation 4) | Yes. A DOM query for a control named `admin` plus a POST key-set check. | `:1580-1582`, `:1596-1597` |
| Contract M9 (render-and-read) | Yes. Fifteen literal states, 150 captures, fixture state per capture, 60 native cells, a per-cell list the conductor enforces, and a fresh re-read. The sheet count is open (minor 6). | `:1622-1651` |
| Contract M10 (reproducibility gated) | Yes. The `web` target builds twice and compares tree hashes; the input walk is sorted; mutation 1 now names a gate. | `:815-819`; Task 3a `:1093-1094`, `:1104` |
| Mechanics M1 (font `src`, built-CSS gate) | Yes. The source is root-absolute; the gate scans built stylesheets only; G5 fails a hashed copy; mutation 4 inverted; Task 4's seed goes in a built stylesheet. | `:546-559`, `:792-797`, `:991-992`, `:1214-1216` |
| Mechanics M2 (WOFF2 comparison) | Yes. Three-part comparison, measured sizes, mutation 1 scoped to Alfa Slab One; the license question is Ruling 9. | `:524-537`, `:967-970`, `:984-987` |
| Mechanics M3 (`channel`, Go-side log) | Yes. `getExecutableName` at `coreBundle.js:43370-43376` returns `"chromium"` for the alias channel, and the `types.d.ts` channel line is present (`:25395`, `:17549`). The Go recorder sees out-of-prefix requests. The missing-browser proof is V-M2. | `:657-671` |
| Mechanics M4 (prefix-absolute icon link) | Yes. The Go rewrite inserts the link and `app.html` carries none; today's `app.html` has no icon link and the built `login.html` has no `<title>`, both confirmed. Mutation 5's wording is minor 1. | `:772-779`; Task 4 `:1207-1213` |
| Mechanics M5 (CSP seed) | Yes. The injected off-origin `<img>` is the runtime row, honestly marked as confirmed by probe; the unhashed script is relabeled "startup refuses". | `:708-712`, `:1087-1090` |
| Risk M1 (header focus ring, Ruling 7) | Partly. The ruling and the row exist, but the geometry and the sample side are unnamed (V-M1). | `:189-207`, `:1289-1295` |
| Risk M2 (toggle `neutral-content`) | Yes. Every header foreground takes it, with a contrast row and mutation 6; Ruling 4's sentence is corrected. The 14.31:1 figure recomputes. | `:148-150`, `:1541-1542`, `:1574`, `:1599` |
| Risk M3 (W7.14 through W7.16, W6.11) | Yes. W section 7 joins the inputs; titles, the sent moment, the escape hatch, supersession, and the focus move are in Tasks 7 and 8; the robots meta is a Go insertion. W7.15's text matches the plan's quotes (`web-surface-standard.md:1059-1076`). | `:313-315`, `:768-771`, `:1370-1386`, `:1432-1442`, `:1489-1494` |
| Risk M5 (lane 1 ordering) | Yes. Named in global constraints, contested surfaces, and conductor's action 1. The narrowed trigger set is disjoint from 6b-ii's written files (`main.go`, `restore.go`, `store.go`, `internal/backup`, `internal/ops`, per its Files lines). | `:418-424`, `:1716`, `:1721`, `:1731-1733` |
| Risk M7 (44px, Ruling 8) | Yes. W4.10's tiers, the five exceptions, the square test, a 20px seed, and VI's 44px header row. W4.10's text matches (`:453-490`). | `:209-222`, `:1636-1639`, `:1575` |
| Risk M8 (cookie `Max-Age`) | Yes. Read through `cookies()`, with a mutation, and declared against R-5cii22. W5.10's "lost only with the cookie jar" matches (`:628-629`). | `:283-286`, `:861-867`, `:1570-1571`, `:1598` |

## Question 2: contradictions, task order, and the split

- **Order builds.** 1, 2, 3a, 3b, 4, 5 is sequential. 3b, 4, and 5 share only `rendered_test.go`
  and `run.ts`, each extending them. The 3a and 3b split is clean: 3b only adds fixture
  servers and rows to the harness 3a builds.
- **Pre-named splits.** Tasks 5, 8, and 9 each name theirs, and each names a clean seam.
- **Deliverables.** The count of 45 reconciles: ten tasks at four, plus 3a at three and 3b at
  two. No task is over four.
- **Arithmetic.** The ceilings recompute: 8.30M, 6.50M, 14.80M, 13.55M, and the flags 8.64M,
  6.80M, and 13.92M.
- **Contradictions found.** Minor 7 (Ruling 1's "no" against action 4) and minor 9's ceiling
  note. No stale revision 1 residue: the 30px seed, 13.4M, `executablePath()`, and
  `<svelte:head>` for a served meta survive only as historical mentions.
- **The Ruling 1 cut.** It is sequential, not file-disjoint (minor 8). Each half names its close
  (Tasks 6 and 11). It also names its reviewers (actions 4 and 8) and its architecture reads
  (actions 5 and 9).

## Question 3: new mechanisms against their sources

| Claim | Checked | Result |
| --- | --- | --- |
| `go test` exits 0 with no test matched; the test2json stream | `go doc cmd/test2json` | Quotes verbatim: "a newline-separated sequence of TestEvent objects", "one of a fixed set of action descriptions", "pass - the test passed", "skip - the test was skipped or the package contained no tests". |
| `channel: 'chromium'` | `playwright-core@1.63.0` in `scratchpad/pw/node_modules`. Playwright is not in `web/node_modules`, since Ruling 2 is unanswered. | The type line is present. The source maps the alias to `"chromium"`. |
| `executablePath` refusal | The same install | The message differs from the mapped one (V-M2). |
| `+layout.ts` sets `ssr = false` | The file | `prerender = true; ssr = false` confirmed. The built `login.html` has no `<title>` (count 0). |
| The Go head-insertion surface | `internal/serve/document.go:57-95` | Today the only insertion is `insertTierLinks`, anchored after the last stylesheet link and applied once before the per-palette loop. `theme-color` differs per palette, so Task 4 must add a per-variant insertion inside the loop and widen the signature (prefix, site name). Task 4's Files list `document.go`, so this is in scope. The document names are `login.html` and `login/link.html` (`web.go:57-58`), which the robots meta's name check must use. |
| W4.10, W5.10, W7.14 through W7.16, W9.1 | `web-surface-standard.md` | Each quote matches. W9.1's halo condition ("a tinted ground shares the ring's hue") does not literally describe the neutral bar. Ruling 7 adopts the mechanism, which is Geoff's call. |
| `html.UnescapeString` | `go doc` | Verbatim. |
| 2.34:1 and 14.31:1 | Recomputed | 2.338:1 and 14.31:1. Ink on cream is 6.12:1. |

## Question 4: can each ruling be decided, and does the plan build under each answer?

| Ruling | Rulable | Builds under yes | Builds under no |
| --- | --- | --- | --- |
| 1 split | Yes | Yes | Yes, with minor 7 |
| 2 Playwright | Yes | Yes | Yes (the leg does not exist) |
| 3 no photo | Yes | Yes | Yes (a block, or a named photo) |
| 4 WCAG 1.4.11 | Yes | Yes | Minor 5 |
| 5 defer rung 7 items | Yes | Yes | Yes (a pre-named split) |
| 6 reservation | Yes | Yes | Yes (cut candidates named) |
| 7 header ring | Yes on evidence | V-M1 | Minor 3 |
| 8 touch tiers | Yes | Yes | Yes |
| 9 WOFF2 and RFN | Yes | Yes | Minor 4 |
| 10 functional neutral | Yes | Yes | Yes |
