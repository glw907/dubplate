# Rung 5c-iii plan fold 1

This record folds the three `spec-plan-review` lenses into the 5c-iii plan
(`docs/superpowers/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`), taking it from
revision 1 (`fc52232`) to revision 2. The lens reports are at `f0e3c00`:
`2026-09-25-rung-5c-iii-plan-review-contract.md` (1 blocker, 10 majors, 15 minors),
`-mechanics.md` (5 majors, 9 minors, 2 owner forks), and `-risk.md` (8 majors, 16 minors, 2
owner forks, and a deviations table). Each finding ID below carries one disposition. A finding
that converges with another folds once, and its line lists every ID it covers.

**Counts:** 66 finding IDs. 59 folded, 7 routed to owner forks (four rulings: 7, 8, 9, and 10),
and 0 refused whole. One sub-proposal is refused inside a folded finding (risk m2's optional
`{@html}` gate).

## What the fold verified before acting

Each lens claim the fold relied on was checked against the tree or a primary source, never
taken on the lens's word:

- `internal/serve/branding.go:73-75` admits hex of 3, 4, 6, or 8 digits and `oklch()`,
  `rgb()`, `hsl()`, so revision 1's `theme-color` grammar would refuse a configuration 5c-ii
  accepts (contract M4, risk M4).
- `go test -count=1 -run NoSuchTestXyz ./internal/atomicfile` printed `ok ... [no tests to
  run]` and exited 0 (contract B1). `go doc cmd/test2json` defines the `-json` stream the fold
  uses.
- `web/src/routes/+layout.ts` sets `ssr = false`, and the built `web/build/login.html` carries
  no `<title>`. So `<svelte:head>` content never reaches the served bytes before a script
  runs. This changed two folds: the robots meta (risk M3) and the icon link (mechanics M4) are
  Go head insertions, not `<svelte:head>` markup.
- W4.10 (`web-surface-standard.md:453-490`) rules 24px as the error floor and 44px as advisory
  (risk M7). W7.15 (`:1059-1076`) carries the three outcome classes, supersession, the sent
  moment, and `noindex, nofollow` (risk M3). W5.10 (`:618-632`) says "the pick is lost only
  with the cookie jar" (risk M8). W9.1 (`:1182-1190`) carries the halo (risk M1).
- The matcher plan (`2026-09-25-matcher-word-boundary.md`, "Global constraints") claims lane 1,
  and `docs/STATUS.md` shows lane 1 free (risk M5). `scripts/check.sh:295-308` shows the
  `--scope-changed` branch never calls `make check` (mechanics minor 7).
- VI `:71` and `:85` give `#f4ead3` and `#1c1916` for base-100 (contract M2); VI `:109` puts
  the header on neutral and VI `:114` claims the ring clears 3:1 "on every surface" (risk M1);
  VI `:260-261` allows the photo to be omitted on a phone and VI `:356` leaves "the stacked
  login" open (risk m5); VI `:149` and `:152` carry the license and clause errors.
- `playwright-core@1.63.0` `types/types.d.ts` documents both `channel` ('Use "chromium" to opt
  in to new headless mode') and `executablePath` ("Path to a browser executable to run instead
  of the bundled one"), the two launch options the fold's new mechanisms use.
- `go doc html.UnescapeString`: "UnescapeString(EscapeString(s)) == s always holds" (contract
  M3's parse-back row).

The fold introduced no mechanism it did not quote from these sources or from a lens's
probe. Behaviors a task must still confirm are named in the plan as such (the injected
element's CSP event, `route.abort`, kit's client-side `error(404)`).

## Dispositions

### Convergent findings, folded once at the root

| IDs | Disposition | Where |
| --- | --- | --- |
| contract B1, contract M1, mechanics minor 9 | Folded. The Go test holds the row manifest as literals; the script reports each row with its fixture-state result; the target runs `go test -json` and fails without a `pass` event for `TestRendered` itself; identity is judged after hydration. | R-5ciii4 ("The row manifest lives in Go", "The target proves the test ran", the refusal list); Task 3a acceptance and mutations 3 and 5; Global constraints' assertion line |
| contract M4, risk M4, mechanics minor 8, mechanics F2 | Owner fork, **Ruling 10** (functional branding neutral: omit the meta and log one line, recommended). The hex forms R-5cii14 admits enter the meta under either answer, folded now. | Rulings; R-5ciii5; privileged-field table; Task 4 states table and mutation 6; deviation on VI `:128-131` |
| contract M5, risk M6 | Folded. Task 5 injects a closed control set into a hydrated stub through `page.evaluate`; no page or route ships a control. The focus row covers every tabbable element on every served page, with a nonzero count, so Tasks 8 and 9 inherit it. Surface rows carry literal expectations from VI. | Task 5 "Where the proofs run" and acceptance |
| contract M2, mechanics minor 2 | Folded. First paint compares against VI's base-100 literals in a six-row table; the cookie is seeded with the `addCookies` form the mechanics probe proved; the frame timing cites the mechanics probe. | R-5ciii8 first-paint table; R-5ciii4 probe paragraph; Task 3a mutation 4 |
| contract m2, mechanics minor 1 | Folded with the mechanics form: the test seeds `mime.AddExtensionType(".woff2", "application/x-seeded")` and asserts `font/woff2`. | R-5ciii1 "Content type"; Task 1 acceptance and mutation 2 |
| contract m7, risk m14, contract M10 (its sort-order clause) | Folded. The version hash covers `svelte.config.js`, `vite.config.*`, and `tsconfig.json`, walked in sorted `/`-separated byte order. | R-5ciii7 |
| contract m8, risk m10 | Folded. Ruling 4's icon-only row moves to Task 9 and asserts the toggle present first. | Ruling 4; Task 5 deliverable note; Task 9 acceptance |
| contract m10, risk m8 | Folded. The fixture environment names the kind answer and the four `auth.Limits` counts; the manifest states its link-request and consumption spend and the Go side asserts both below their limits; Task 10 mints one link per consuming capture. | R-5ciii4 "The fixture environment, named"; Task 10 |
| contract m15, risk m13 | Folded. Ruling 6's figure now matches its table: 8.30M and 6.50M, 14.80M together (revision 1's 7.15M and 6.10M summed to 13.25M, not 13.4M). | Ruling 6; Token ceiling |
| mechanics minor 3, risk m1 | Folded. The swap's ring moves to the control through `.swap:has(> input:focus-visible)` and is suppressed on the input; a row asserts the ring box contains the icon box. | Task 9 outcomes and acceptance |
| mechanics minor 5, risk m15 | Folded. Every task's gate runs heavy; Tasks 1 and 2 report peak memory. No light-lane measurement is needed once the light lane is not used. | Global constraints (the gate); Tasks 1 and 2 acceptance |
| risk M3, risk m6, risk m7 | Folded. W section 7 joins Primary inputs; the deck carries every `<title>` and the sent moment's copy; Task 8 builds the three W7.15 classes, supersession, the sent moment with its pre-filled escape hatch, and W7.16's focus move; the robots meta is a Go head insertion in Task 4, not `<svelte:head>` as the lens suggested, because `ssr = false` keeps `<svelte:head>` out of the served bytes. | Primary inputs; R-5ciii5; Task 4; Task 7; Task 8 |
| contract M7, risk m16 | Folded. The non-admin row signs the member in through a minted link, injects `<a href="../members">` on the hydrated interstitial, proves the navigation was client-side from the Go request log, then asserts the sentence, no address, no form, no roster. The not-found mechanism is the task's to name with kit's documentation quoted. | Task 9 outcomes and acceptance; privileged-field table (harness sessions) |
| contract m9, risk M5 (its scope-trigger clause) | Folded. The `--scope-changed` trigger set is `web/`, `internal/serve/`, `internal/auth/`, `internal/identity/`, and `cmd/dubplate-server/{web,serve,auth,rendered_test}.go` with their tests, so 6b-ii's `restore.go` and `main.go` stay outside it. | R-5ciii4 "Where it runs" |

### Contract lens

| ID | Disposition | Where |
| --- | --- | --- |
| B1 | Folded (convergent row above). | R-5ciii4; Task 3a |
| M1 | Folded (convergent row above). Task 10's capture list uses the same manifest. | R-5ciii4; Task 10 |
| M2 | Folded (convergent row above). | R-5ciii8 |
| M3 | Folded. Task 4 proves the site-name meta by Go rows with an `html.UnescapeString` read-back; the rendered heading row and mutation move to Task 8 with a fixture name carrying `"`, `<`, and `&`. Ruling 1's reason now names the exception. | Ruling 1; R-5ciii5; Task 4; Task 8 acceptance and mutation 6 |
| M4 | Owner fork, Ruling 10 (convergent row above). | Rulings |
| M5 | Folded (convergent row above). | Task 5 |
| M6 | Folded. Task 8 carries a state table naming each outcome's source, class, and politeness; the fixture names `GenericAnswer` false. | R-5ciii4; Task 8 |
| M7 | Folded (convergent row above). | Task 9 |
| M8 | Folded. The add form holds no control named `admin` (DOM query), and the POST body's key set is checked; mutation 4 now fails on the DOM query. | Task 9 acceptance and mutation 4 |
| M9 | Folded. Fifteen page states as literals, 150 captures, fixture state asserted per capture, native-resolution reads at 390 and 1440 in both schemes with a per-cell list the conductor checks, and a fresh re-read of changed cells after the fold. | Task 10; conductor's action 7 |
| M10 | Folded. The `web` target builds twice and compares tree hashes; the input walk is sorted. | R-5ciii7 "It gates"; Task 3a mutation 1 |
| m1 | Folded. Two refusals are gate-resident negative rows: a nonexistent `executablePath`, and an injected off-origin `<img>`. | R-5ciii4; Task 3a mutation 2 |
| m2 | Folded (convergent row above). | Task 1 |
| m3 | Folded. The toggle writes the computed `--color-neutral`; the privileged-field table lists the toggle as a writer; Task 9's row reads the meta against VI's literals. | R-5ciii5; privileged-field table; Task 9 |
| m4 | Folded. Task 9 names the four starts, each with its checked state and written value. | Task 9 acceptance |
| m5 | Folded. The coverage test asserts named members and a floor of 23 parsed values; the AI-tells scanner asserts a nonzero scanned count. | Task 7 |
| m6 | Folded. G9's list is closed: inks and `base-content` on three bases, each `*-content` on its own fill, the two fill-only colors named as excluded. | R-5ciii3; Task 2 acceptance |
| m7 | Folded (convergent row above). | R-5ciii7 |
| m8 | Folded (convergent row above). | Task 9 |
| m9 | Folded (convergent row above). | R-5ciii4 |
| m10 | Folded (convergent row above). | R-5ciii4; Task 10 |
| m11 | Folded. Exit gate item 13 names the closed 5c-ii carry. | The exit gate |
| m12 | Folded. Task 10 asserts a nonzero control count per page. | Task 10 |
| m13 | Folded. The I2 row asserts no logged request carries the token in its URL or headers, and the ordering row is `location.search` read empty before the click. | Task 8 acceptance and mutation 3 |
| m14 | Folded. One error-page row per reachable status, by an injected link to an unmatched path. | Task 8 acceptance |
| m15 | Folded (convergent row above). | Ruling 6 |

### Mechanics lens

| ID | Disposition | Where |
| --- | --- | --- |
| M1 | Folded. The `@font-face` source is root-absolute (Vite's public-path form); the root-absolute gate checks CSS `url(/…)` in the built stylesheets only; G5 requires every built `src` under `fonts/` on a manifest file; Task 1 mutation 4 now writes a relative `src` and G5 fails; Task 4's CSS seed is in a built stylesheet. | R-5ciii1 "Where they are served"; R-5ciii6; Task 1 outcomes, G5, mutation 4; Task 4 acceptance |
| M2 | Folded. The comparison is three-part (every table but `DSIG` and `head` identical; `DSIG` absent and an 8-byte stub in the source; `head` differing only at bytes 8 through 11 and flags bit 11); sizes are the measured 39,668 and 138,000 bytes; mutation 1 is Alfa Slab One's alone. The license question is F1. | R-5ciii1; Task 1 acceptance and mutation 1 |
| M3 | Folded. The harness launches with `channel: 'chromium'`; the missing-browser refusal is the launch error mapped to the install hint; request-log rows read a Go-side recorder that sees requests outside the prefix. | R-5ciii4; Ruling 2; Task 0 item 4 |
| M4 | Folded. The Go rewrite inserts prefix-absolute icon links; `app.html` carries none; Task 4 adds a client-navigation row and a mutation writing the link relative. The `<svelte:head>` alternative is not taken because `ssr = false` keeps it out of the served bytes. | R-5ciii5 "The icon links"; R-5ciii6; Task 4 |
| M5 | Folded. The CSP-violation refusal is proved by an injected off-origin `<img>` (a runtime `img-src` event); the unhashed-script seed stays as a separate row labeled "startup refuses". | R-5ciii4; Task 3a acceptance |
| minor 1 | Folded (convergent row above). | Task 1 |
| minor 2 | Folded (convergent row above). | R-5ciii4; R-5ciii8 |
| minor 3 | Folded (convergent row above). | Task 9 |
| minor 4 | Folded. The `version.json` seed builds the test binary with `go test -c -tags rendered`, edits the file, then runs the binary. | Task 3a acceptance |
| minor 5 | Folded (convergent row above). | Global constraints |
| minor 6 | Folded. The `web-rendered` target passes an explicit `-timeout` of at least twice the measured wall time, reset at Task 10. | R-5ciii4; Tasks 3a and 10 |
| minor 7 | Folded. Task 3a wires the leg into the `--scope-changed` branch explicitly, and its acceptance checks both branches' scope lines. | R-5ciii4; Task 3a |
| minor 8 | Owner fork, Ruling 10 (convergent row above). | Rulings |
| minor 9 | Folded (convergent row above). | R-5ciii4; R-5ciii5 |
| F1 | Owner fork, **Ruling 9** (the `DSIG` drop and `head` bit 11 under Alfa Slab One's Reserved Font Name; yes recommended). Task 0 now checks for an upstream-published WOFF2, which the "no" answer uses. | Rulings; Task 0 item 3; R-5ciii1 "Pinned bytes" |
| F2 | Owner fork, Ruling 10 (convergent row above). | Rulings |

### Risk lens

| ID | Disposition | Where |
| --- | --- | --- |
| M1 | Owner fork, **Ruling 7** (W9.1's halo on the header bar; yes recommended). The row the lens asked for is folded regardless: Task 5's focus row measures each ring's painted contrast against its adjacent pixels at 3:1 or better, so the header rows fail without the ruling's treatment. | Rulings; Task 5 acceptance; Task 9 acceptance |
| M2 | Folded. Every header foreground takes `--color-neutral-content`; Ruling 4's sentence is corrected; a header contrast row and a mutation are added. | Ruling 4; Task 9 outcomes, acceptance, mutation 6 |
| M3 | Folded (convergent row above). | Primary inputs; Tasks 4, 7, 8 |
| M4 | Owner fork, Ruling 10 (convergent row above). | Rulings |
| M5 | Folded. The plan names the lane 1 contention and leaves the order to the conductor's registry entry, recommending the matcher pass first; the contested-surfaces table and conductor's actions 1 and 3 carry the 6b-ii line; the scope trigger is narrowed (convergent row above). | Global constraints; Contested surfaces; conductor's actions 1 and 3 |
| M6 | Folded (convergent row above). | Task 5 |
| M7 | Owner fork, **Ruling 8** (W4.10's tiers stand with VI's 44px as a header row; yes recommended). The 20px seed, the five exceptions, the square test, and VI's header row are folded now. | Rulings; Task 9; Task 10 |
| M8 | Folded. The toggle writes `Max-Age=31536000`; the row reads `expires` through the context's `cookies()`; a mutation drops it. Declared as a deviation on R-5cii22. | R-5ciii9; privileged-field table; Task 9; deviations |
| m1 | Folded (convergent row above). | Task 9 |
| m2 | Folded in part. The admin contact joins the privileged-field table as text-only, and its nil state gets a deck line and a Task 8 row. The optional static gate refusing `{@html` is refused: Svelte's text interpolation already escapes, Task 8's mutation 6 proves the heading path, and the pass-end reviewers read every component, so a new gate script costs more than the risk it removes. | Privileged-field table; Task 7; Task 8 |
| m3 | Folded. The I2 history reads repeat after the skip link and after the consume click, and the page never reads kit's `page.url`. The lenses disagree on `client.js:3257` (risk: kit's URL; mechanics: the current location), and the repeated reads settle it in the browser. | Task 8 outcomes and acceptance |
| m4 | Folded. These pages ship no `placeholder` attribute, asserted in Tasks 8 and 9; `label-night`'s placeholder opacity is carried to rung 7 as a VI value question. | Tasks 8, 9, 11 |
| m5 | Folded. Ruling 3's reason now cites VI `:260-261` and says the desktop single column is a new layout the ruling adopts. | Ruling 3 |
| m6 | Folded (convergent row above). | Task 7 |
| m7 | Folded (convergent row above). | Primary inputs |
| m8 | Folded (convergent row above). | R-5ciii4; Task 10 |
| m9 | Folded. The redirect trap joins the refusal list and Task 3a's acceptance. | R-5ciii4; Task 3a |
| m10 | Folded (convergent row above). | Task 9 |
| m11 | Folded. The Fontsource decline now rests on R5.5's budget alone. | R-5ciii1 |
| m12 | Folded. Declared as a deviation from CLAUDE.md's just-in-time rule, with a refresh read of Tasks 7 through 11 at 5c-iii's close (see below). | Ruling 1; deviations; conductor's action 5; Task 0; Segment D preamble |
| m13 | Folded (convergent row above). | Ruling 6 |
| m14 | Folded (convergent row above). | R-5ciii7 |
| m15 | Folded (convergent row above). | Global constraints |
| m16 | Folded (convergent row above). | Task 9 |

## Rulings awaiting owner, consolidated

The plan's "Rulings awaiting owner" section carries all ten as yes/no questions with a
recommendation and what each answer builds.

1. Split into 5c-iii (Tasks 1, 2, 3a, 3b, 4, 5, 6) and 5c-iv (7 through 11) under 10.8M and
   8.5M. Recommend yes.
2. `playwright@1.63.0` as an exact devDependency, launched with `channel: 'chromium'`.
   Recommend yes.
3. The login page ships as a single column without the photograph. Recommend yes.
4. VI's WCAG 1.4.11 reading stands (no keyline on a text-labeled `label-night` primary).
   Recommend yes. The contract lens notes this is a conformance reading the W3C text settles;
   it stays a ruling because VI asked for confirmation.
5. The sleeve, grain, rim, and Permanent Marker land with rung 7. Recommend yes.
6. The 5c reservation rises to about 14.8M for this scope, rung 5 to about 51M. Recommend yes.
7. The header ring clears 3:1 through W9.1's `base-100` halo (risk M1). Recommend yes.
8. W4.10's tiers stand (24px error, 44px advisory) with VI's 44px as a header row (risk M7).
   Recommend yes.
9. A WOFF2 that drops the empty `DSIG` stub and sets `head` bit 11 keeps Alfa Slab One's
   Reserved Font Name (mechanics F1). Recommend yes.
10. A functional-color branding neutral omits the `theme-color` meta and logs one line
    (contract M4, risk M4, mechanics minor 8 and F2). Recommend yes. The contract and risk
    lenses preferred carrying the value through; that is the "no" answer's build.

## Ceilings, re-derived

Revision 1's Task 3 took its pre-named split (3a: the harness, the reproducibility gate, first
paint; 3b: G7 and G10), since the gated second build and the row manifest made its harness
deliverable a fifth. No other task passes four deliverables; Tasks 8 and 9 each gained a
pre-named split.

| Half | Tasks | Projection | Ceiling | Flag |
| --- | --- | --- | --- | --- |
| 5c-iii | 1, 2, 3a, 3b, 4, 5, 6 | 8.30M (seven at 1.15M, +0.30M browser, -0.05M reads) | 10.8M | 8.64M |
| 5c-iv | 7 through 11 | 6.50M (five at 1.15M, +0.25M refresh, +0.15M register gate, +0.35M reader) | 8.5M | 6.80M |
| Ruling 1 "no" | twelve tasks | 13.55M | 17.4M | 13.92M |

## Deviations declared

Revision 1 declared four; revision 2 keeps them and declares three more:

- R-5cii22's cookie attribute set gains `Max-Age=31536000` (W5.10; risk M8).
- VI `:128-131`'s `theme-color` does not describe a functional-color branding neutral
  (Ruling 10; risk M4).
- CLAUDE.md's just-in-time plan rule, for 5c-iv, mitigated by the refresh read at 5c-iii's
  close (risk m12).

The risk lens's other two undeclared rows need no declaration in revision 2. The plan now
conforms to W4.10 unless Ruling 8 answers no. It implements W6.11 and W7.14 through W7.16 in
Tasks 7 and 8.

## 5c-iv's plan staleness (risk m12)

5c-iv's Tasks 7 through 11 were drafted before 5c-iii's reviews. Rather than refuse, the plan
adds a just-in-time refresh at 5c-iii's close. One verification-lens read of Tasks 7 through
11 checks them against the 5c-iii close, its review fold, and the tree, and one fold applies
what it finds before 5c-iv's Task 0 runs. The refresh is priced at 0.25M in 5c-iv's ceiling.
Task 6's handoff lists every 5c-iii fold that touches a surface 5c-iv consumes, as the refresh
read's input.

## Errata owed

The fold edits no ratified document. These are owed, and Task 6 or Task 11 drafts each one
paste-ready:

- **VI `:149`:** Permanent Marker is Apache-2.0, not OFL (`google/fonts`
  `apache/permanentmarker/`, `@fontsource/permanent-marker@5.3.0`). DL R5.2 admits Apache 2.0,
  so the face stays eligible (Ruling 5).
- **VI `:141` and `:152`:** both cite "R5.4" for the four-file budget; the clause is R5.5.
- **VI `:114`:** "clears the 3:1 focus-appearance floor on every surface" was measured on the
  three bases only; on the `label` neutral header bar the ring measures 2.34:1. The erratum's
  wording follows Ruling 7.
- **VI `:5` and `:112-115` (under Ruling 7's no):** the focus-ring color is one per surface
  class, not one everywhere. The neutral header bar takes `label-night`'s `#ec7b6a` under both
  themes, declared in `web/src/lib/themes/inks.css`. Added by the prose fold.
- **VI `:128-131`:** the `theme-color` rule under a functional-color branding neutral, worded
  per Ruling 10.
- **DL R5.5:** "four woff2 files", with no family count (revision 1's declared deviation).
- **W4.1:** the both-schemes half binds at rung 5, since rung 5 ships two composed themes.
- **W section 15 and the rung 5 spec's 5c row:** the third split (Ruling 1).
- **Rung 5 spec `:1438`:** the add-member page carries no admin checkbox (R-5c13).
- **5c-ii plan R-5cii8:** the members row is `no-store`, as the code already serves.
- **5c-ii plan R-5cii22:** the toggle's cookie carries `Max-Age=31536000` (R-5ciii9).

Carried to rung 7, not an erratum yet: `label-night`'s placeholder opacity, which VI `:332`
leaves at DaisyUI's default (4.38:1 on base-100, risk lens), is a VI value question for Geoff;
and W4.10's 44px promotion, which W4.10 already reserves for rung 7.

## Fold 2

Input: the verification read
(`docs/superpowers/research/2026-09-25-rung-5c-iii-plan-fold-verification.md`, `bbba7a8`; 0
blockers, 2 majors, 9 minors) against revision 2 (`0aeb644`). The fold produces revision 3.
It reopens nothing else and changes no ruling's substance. All ten rulings stay yes/no with
their recommendations, and no task passes four deliverables.

### Verified before acting

- **Contrast figures,** recomputed as sRGB relative luminance per WCAG 2.2, with the palette
  values from VI `:68-89`: `#9d2f26` on `#1f1b16` is 2.338:1; on `#f4ead3` it is 6.121:1.
  `#f4ead3` on `#1f1b16` is 14.313:1. `#ec7b6a` on `#1c1916` is 6.340:1, on `#1f1b16` 6.202:1,
  and on `#110f0d` 6.929:1. Mustard is `#e0a526` on `#1f1b16` at 7.810:1, `#e0a526` on `#110f0d`
  at 8.725:1, `#e5ad33` on `#1f1b16` at 8.443:1, and `#e5ad33` on `#110f0d` at 9.433:1.
  Revision 2's Ruling 7 "no" figure, "6.93:1 on `#1f1b16`", paired the wrong bar: 6.93:1 is
  `label-night`'s bar, and `label`'s bar measures 6.20:1.
- **The registry path,** read in `playwright-core@1.63.0` at
  `scratchpad/pw/node_modules/playwright-core`. `coreBundle.js:33120` reads
  `PLAYWRIGHT_BROWSERS_PATH` inside the `registryDirectory` initializer. `:33180` throws
  ``Executable doesn't exist at ${e}`` with the install box. `:39930` throws "Failed to launch
  ${this._name} because executable doesn't exist at ${executablePath}". `types/types.d.ts`
  documents `launch()`'s `executablePath` at `:25418-25422`. Probe `pw/missing.mjs` ran
  `chromium.launch({ channel: 'chromium' })` with `PLAYWRIGHT_BROWSERS_PATH` set to an empty
  scratch directory. It exited 1 with `browserType.launch: Executable doesn't exist at
  <dir>/chromium-1243/chrome-linux64/chrome`.
- **The head insertion surface,** read at `internal/serve/document.go:57-71`:
  `insertTierLinks` runs once before the per-palette loop. The document names are at
  `cmd/dubplate-server/web.go:57-58`.
- **The VI lines:** "Dark theme" is VI `:131`, and "The login page has no toggle" is `:132`.

### Dispositions

| Finding | Disposition | Where |
| --- | --- | --- |
| V-M1 (halo geometry, sample side) | Folded. The Ruling 7 question names a halo that spreads past the ring. Its "yes" names `box-shadow: 0 0 0 6px var(--color-base-100)` beneath the 2px outline at its 2px offset, so `base-100` borders the ring on both sides. Task 5's focus row samples one pixel inside the ring and one outside, both at 3:1 or better. Under yes, both header samples read the halo, at 6.12:1 on `label` and 6.34:1 on `label-night`. Under no, both read the bar, at 6.20:1 and 6.93:1. Revision 2's "6.93:1 on `#1f1b16`" is corrected, and the mustard figure is stated as 7.81:1 or better. Task 9's header row now reads "its two samples clear 3:1 in the header". Fold 1's claim that "the header rows fail without the ruling's treatment" now holds, because the row takes the outer sample. | Ruling 7; Task 5 acceptance; Task 9 acceptance |
| V-M2 (missing-browser path) | Folded. The resident row sets `PLAYWRIGHT_BROWSERS_PATH` to an empty temporary directory, and the probe output is quoted. The script maps a message containing `Executable doesn't exist at` (`coreBundle.js:33180`), and any other launch error exits nonzero with its own message. The `executablePath` override is dropped and kept only as the rejected proof. Its cite is re-anchored to `launch()`'s option at `types.d.ts:25418-25422`. | R-5ciii4 "It gates" and "Two refusals"; Task 3a acceptance |
| Minor 1 (mutation 5) | Folded. Mutation 5 writes `../favicon.svg` in the `login/link.html` variants and names the client-navigation row. | Task 4 mutation 5 |
| Minor 2 (SVG or PNG) | Folded by the row, not by a probe. The row accepts at least one icon request on either inserted prefix-absolute path and none outside the prefix. Mutation 4 still fails it, because no request lands on either path. | Task 4 acceptance |
| Minor 3 (Ruling 7 "no") | Folded. The header redefines `--ink-primary` in its own scope, so the inherited color assertion holds. No header text reads the ink, because every header foreground takes `--color-neutral-content`. | Ruling 7 |
| Minor 4 (Ruling 9 "no") | Folded. Under the rename, the comparison admits the `name` table as a fourth expected difference, and `FONTS.md` records the rename command. | Ruling 9 |
| Minor 5 (Ruling 4 "no") | Folded. The keyline lands in Task 5's printed surface, read on the injected `btn-primary` over both grounds. Task 8 adds a row on the login primary when it sits on either ground. No deliverable is added. | Ruling 4 |
| Minor 6 (45 vs 90) | Folded. The count is 90 sheet cells: fifteen states at 320, 768, and 2560 in both schemes. With 60 native cells, that makes 150. | Task 10 acceptance |
| Minor 7 (Ruling 1 "no" vs action 4) | Folded. Action 4 opens "under Ruling 1's yes", and under no its reviews run once, at action 8. | Conductor's action 4 |
| Minor 8 (sequential seam) | Folded. Task 6's handoff names `rendered_test.go`, `run.ts`, `Makefile`, and `components.css` with their state at the close, and Ruling 1's reason says the seam is sequential. | Task 6 outcomes; Ruling 1 |
| Minor 9, "Task 8's split" | Folded. The 5c-iii ceiling note cites Task 5's split alone. | Token ceiling |
| Minor 9, VI cites | Folded. `:132` becomes `:131` in Task 7, and `:133` becomes `:132` in Task 8. | Tasks 7, 8 |
| Minor 9, Task 3a density | No change. The verifier confirms the count holds at three. A pre-named split would add a declaration for a task that is not over its count. | none |
| Question 3, per-variant insertion | Folded. Task 4's Files line records that `theme-color` needs a per-variant insertion inside the palette loop, with a widened signature. It also records that the robots meta keys on `login.html` and `login/link.html`. | Task 4 Files |
| Question 3, W9.1's halo condition | No change. The verifier leaves the adoption of W9.1's mechanism on a neutral ground to Geoff, and Ruling 7 already puts it to him. | Ruling 7 |

Refused: none. The only item left unchanged is Task 3a's density, which the verification read
recorded as a count that holds, not as a defect.

## Prose fold

Input: the prose review (`docs/superpowers/research/2026-09-25-rung-5c-iii-plan-prose-review.md`;
3 blockers, 5 warnings, 9 suggestions) against revision 3 (`28c983e`). The fold produces
revision 4, status "reviewed; awaiting owner approval (Rulings 1-10)". It changes no ruling's
substance, no ceiling, and no task's scope beyond W1.

### Verified before acting

- **`scripts/check.sh` (B2).** `derive_conform_affected` (`:209-221`) sets `conform_affected`
  when any changed file sits under `internal/harness/`, `internal/engine/`, or `corpus/`. Only
  then does `scope_conform` stay `full`. The `skipped` branch (`:296-308`) runs `make vet lint
  tidy-check` and `go test` over every package but the conform suite. The `full` branch
  (`:309-313`) runs `make check`, which runs `web-rendered`. So an engine-only change set does
  run the leg. Fold 1's "Verified before acting" line ("`--scope-changed` branch never calls
  `make check`") was wrong, and this fold corrects the plan rather than that record.
- **VI lines (B1, B3).** VI `:108-109` marks `label` secondary `#e0a526` fill-only, "never text,
  never a focus or selector indicator", and `:113` repeats the ban for focus. The `label-night`
  table (`:81-93`) gives info `#6d9cc7`, warning `#e5ad33`, and error `#e0604f`; the ink table
  (`:102-106`) gives the error ink `#e37866`. G1 exempts only `lib/themes/`
  (`web/scripts/gates/g1.ts:11`).
- **Contrast, recomputed** with the WCAG 2.2 sRGB relative-luminance formula: `#ec7b6a` on
  `#1f1b16` 6.202:1 and on `#110f0d` 6.929:1; `#e0a526` on `#1f1b16` 7.810:1; `#9d2f26` on
  `#1f1b16` 2.338:1 and on `#f4ead3` 6.121:1; `#ec7b6a` on `#1c1916` 6.340:1; `#765000` on
  `base-300` `#dccaa3` 4.462:1; `#744e00` on `#dccaa3` 4.597:1, on `#eadcbc` 5.459:1, and on
  `#f4ead3` 6.199:1.
- **Pointers.** `outcomeAdminNotGrantable` is `internal/auth/admin.go:43`. `http.ServeContent`
  is `internal/serve/document.go:230`. `web/package.json` pins `daisyui` `5.7.46`. The 5c-ii
  plan's provisional 5.0M is "six tasks at 0.725M (4.35M)" plus adders (`:425-426`), and its
  ceiling is re-derived under "Token ceiling" (`:371-373`). No "eight-task line" exists in any
  brief the plan names.
- **The W3C quote.** The Understanding document for SC 1.4.11, fetched 2026-09-25, reads "as is
  therefore not subject", so the `[sic]` marks the source's text.

### Dispositions

| Finding | Disposition | Where |
| --- | --- | --- |
| B1 (mustard focus option) | Folded. The mustard option is dropped: VI rules `label` secondary never a focus indicator. The "no" answer now places the header-scoped `#ec7b6a` in `web/src/lib/themes/inks.css` because G1 refuses a literal elsewhere, and Task 9's Files gain that file under the no. The VI erratum (ring color per surface class) joins "Errata owed" and Task 6's errata list. | Ruling 7; Task 9 Files; Task 6 outcomes; Errata owed |
| B2 (`--scope-changed` premise) | Folded. R-5ciii4 "Where it runs" states that `--scope-changed` calls `make check` only for a harness, engine, or corpus change, and otherwise runs the narrowed branch the leg is wired into. Task 3a's skip-line change set is `internal/atomicfile/`, which neither trigger set covers. Task 0's pointer becomes `scripts/check.sh:1-60,209-221,296-313`. | R-5ciii4; Task 3a acceptance; Task 0 item 2 |
| B3 (wrong inks) | Folded. The ink-equals-fill note names info `#6d9cc7` and warning `#e5ad33`. | R-5ciii3 |
| W1 (surfaces before a caller) | Folded by the move. The `Mark` component leaves Task 4 (its SVGs, favicon, icon links, and routes stay) and enters Task 8's Files, inside the login deliverable, so Task 8 stays at four. Task 9 now depends on Task 8 for `Mark`, and the dependency list says so. The placeholder opacity leaves both R-5ciii3's surface tokens and Task 5's refinement rules, since no page in the plan carries a `placeholder` attribute. Task 11's handoff and the exit gate's deferred list carry "both palettes' placeholder opacity" to rung 7. | Ruling 1 reason; "What this pass is"; R-5ciii3; Tasks 4, 5, 8, 9, 11; the exit gate; Execution model |
| W2 (Ruling 7 reads as a measurement) | Folded. The question is the reviewer's rewrite, citing VI `:5`. The halo is scoped to the header controls' `:focus-visible` state. The W9.1 sentence says W9.1 prescribes the halo for a shared hue, and the bar fails on luminance instead. | Ruling 7 |
| W3 (unresolvable attributions) | Folded. Ruling 6 cites the 5c-ii plan's "Token ceiling" instead of CLAUDE.md. Ruling 1 drops "the brief's eight-task line". | Rulings 1 and 6 |
| W4 (mutation not isolated) | Folded. Task 2's mutation 2 uses `#744e00`, which passes every contrast row, so the pinned-value row fails alone. | Task 2 mutation 2 |
| W5 (catalog sentences) | Folded. Sources, "What this pass is", R-5ciii4's "It gates", Task 3a's refusal criterion, Task 6's outcomes, and Task 7's deck are lists. Ruling 2's evidence carries one fact per sentence. | Those sections |
| S: Ruling 9 third edit | Folded. The question names "and so rewrites `checkSumAdjustment`". | Ruling 9 |
| S: Ruling 4 recommendation | Folded with the reviewer's wording. | Ruling 4 |
| S: Ruling 6 figure | Folded: "six tasks at 0.725M (4.35M) plus adders". | Ruling 6 |
| S: pointers | Folded: `admin.go:43`, `document.go:230`. | Deviations; R-5ciii1 |
| S: wording | Folded: "the 5c-ii plan's `5.7.45`"; "the two colors VI marks fill-only"; the exit gate's "2px `--ink-primary` ring at a 2px offset". | Global constraints; R-5ciii3; exit gate item 8 |
| S: `.gitleaks.toml` against the wall | Folded by adding it to the wall's write list, scoped to Task 1's condition. | Global constraints |
| S: `[sic]` | Folded after fetching the W3C text. | Ruling 4 |
| S: Ruling 8's "no" (optional) | Refused. The "no" already names its one build, 44px as the error tier for every control. A second "no" reading would add a fork Geoff did not raise. | none |

Refused: one optional suggestion (Ruling 8's "no"). tellgrader `--register docs` after the
fold: one finding, the intentional slop seed in Task 7's mutation 2.
