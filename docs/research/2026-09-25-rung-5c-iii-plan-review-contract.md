# Rung 5c-iii plan review: contract and criteria lens

Reviewed: `docs/superpowers/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`,
revision 1 (`fc52232`). One lens only: is every promise testable, does each acceptance
criterion name its fixture state and the reason it would fail, and can any criterion, gate, or
named mutation pass vacuously. Checked against the visual identity spec revision 1.2 (VI), the
5c-ii HISTORY entry and its handoff draft
(`docs/superpowers/research/handoffs/draft-HISTORY-5c-ii.md`, "The 5c-iii handoff"), DL
section 6, and CLAUDE.md's planning-miss rules and its "a check either gates or does not
exist" rule. Code facts were read at the working tree's HEAD.

Findings are ranked by consequence. A finding whose fix costs more than the risk it removes
may be refused at the fold.

**Counts:** 1 blocker, 10 majors, 15 minors. No new owner fork. All six rulings are rulable
on the evidence given, with the caveats under "The six rulings".

## Blocker

### B1. The rendered leg passes when its test never runs

- **Location:** plan `:479-486` (R-5ciii4, "Where it runs"), `:747-800` (Task 3), exit gate
  item 3 at `:1171`.
- **Defect:** the `web-rendered` target runs `go test -tags rendered -count=1 -run
  TestRendered ./cmd/dubplate-server`. When `-run` matches no test, `go test` prints `ok ...
  [no tests to run]` and exits 0. Three ordinary slips produce that state: a misspelled
  `//go:build rendered` line, a renamed test function, or a tag dropped from the Makefile.
  Each one leaves the rendered leg reporting success without launching a browser. Every
  W12.1 refusal the plan lists (`:487-493`) lives inside the test, so none of them can catch
  a test that was never compiled or selected. This is the "browser that never launched" shape
  the review focus names (`:602-603`), and it escapes the refusal set by construction.
- **Proposed fold:** the Makefile target runs the test with `-json` (or `-v`) and fails
  unless the output carries a pass line for `TestRendered` itself, plus one line per expected
  subtest (see M1). Add to Task 3's acceptance criteria: "the target fails by name when the
  build tag is misspelled in a scratch copy", and add that edit as a named mutation.

## Majors

### M1. The zero-row refusal is the only floor; a short matrix passes

- **Location:** `:493` ("a row matrix that ran zero rows"), `:771-773`, `:1132-1135`.
- **Defect:** the refusal catches zero rows only. A loop that silently drops a page, a
  scheme, or a state still passes as long as one row ran. The plan also leaves open which
  side counts the rows. If the Node script counts its own rows, then a broken script is the
  only witness to its own coverage. That breaks the no-expectation-from-the-code-under-test
  rule (`:315-318`).
- **Proposed fold:** the Go test holds the declared row manifest (row IDs as literals in the
  test, derived from this plan's named rows). The script reports each row as structured
  output (ID, fixture-state assertion result, property result). The Go side fails on any ID
  that is missing, extra, or reported without its fixture-state assertion. The zero-row
  refusal then falls out as a special case. Task 10's matrix (M9) uses the same mechanism.

### M2. First paint compares the page with itself

- **Location:** `:583-587` (R-5ciii8, "First paint"), `:776-777`, exit gate item 5 at
  `:1174`.
- **Defect:** the expected value is "the stamped or media-resolved `--color-base-100`". If
  the row reads that value from the page under test, it passes under whatever palette
  applies, including the wrong one. Examples: a `label` stamp under dark emulation that
  resolves to `label-night`, or a `system` branch declared in the wrong order. Named
  mutation 5 (`:790-791`) fails only if the expectation is a literal. The plan also quotes no
  source for two behaviors the row relies on: that a screenshot taken at `domcontentloaded`
  shows the first painted frame, and how the cookie reaches a context with JavaScript
  disabled (`context.addCookies` with `secure: true` against `http://127.0.0.1`; the probe at
  `:497-502` measured `document.cookie` only). This breaks the planning-miss rule that every
  relied-on behavior must be quoted.
- **Proposed fold:** pin the six expectations to VI's hexes as literals with their source
  line: `#f4ead3` (`label`, VI `:71`) and `#1c1916` (`label-night`, VI `:85`). The state
  table is: no cookie + light gives `#f4ead3`; no cookie + dark gives `#1c1916`; cookie
  `label` under either emulation gives `#f4ead3`; cookie `label-night` under either gives
  `#1c1916`. The task quotes the Playwright docs for `addCookies` on a secure cookie over
  loopback and for the screenshot's frame timing, or probes both and records the result.

### M3. Task 4's rendered site-name row has no page to render on

- **Location:** `:841` (the states table's last row), `:846-848`, `:854-855` (mutation 1),
  Task 4 Files at `:815-818`, and R-5ciii5 at `:527-529`.
- **Defect:** the site name's first reader is Task 8's login heading. At Task 4 the stubs
  show "Log in", "Confirm login", and "Members" (`web/src/routes/*/+page.svelte`), and Task
  4's file list does not include any route component. Three things break as a result: "the
  rendered heading shows the literal characters", "the site name renders as text", and
  mutation 1's "the rendered heading row fails". Either none of them can be met, or the
  implementer edits a route outside its territory and ships a site-name reader ahead of its
  caller. Ruling 1's reason that "every Task 1 through 5 surface is exercised by the
  rendered harness against the 5c-ii stubs" (`:75-76`) is false for this surface.
- **Proposed fold:** Task 4 proves the meta through Go rows only: escaped bytes in every
  variant for `"`, `<`, and `&`, plus a parse-back row that reads the attribute with
  `golang.org/x/net/html` or `html.UnescapeString` and asserts it equals the input. Move the
  rendered heading row and mutation 1 to Task 8. Task 8's acceptance criteria gain a fixture
  site name carrying `"`, `<`, and `&` (at `example.test` scope, universal). The fixture
  state is "the served meta asserted present before the heading is read".

### M4. A branding neutral that R-5cii14 already accepts now refuses startup, or has no named report

- **Location:** `:339` (privileged-field row), `:520-523` (R-5ciii5), `:836-839` (states
  table).
- **Defect:** the `theme-color` value must match `^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`, and "a
  value outside the grammar refuses startup". The branding grammar, however, accepts 4-digit
  and 8-digit hex and `oklch()`, `rgb()`, and `hsl()` colors
  (`internal/serve/branding.go:73-75`, `:165`). An operator whose `DUBPLATE_BRANDING` sets
  `--color-neutral: oklch(...)` is valid under 5c-ii and meets a state the table does not
  list. The literal reading of `:339` makes serve refuse to start on a configuration that
  worked before this pass. That is a silent contract change to R-5cii14, which Ruling 7 of
  5c-ii settled as "refuse only a value outside the branding grammar".
- **Proposed fold:** add the state row "branding sets `--color-neutral` to a
  branding-grammar value outside the hex grammar" with a named report. Recommended report:
  the meta carries the branding value as-is. `theme-color` accepts any CSS color, and the
  branding grammar is already closed and attribute-safe, so the meta can reuse
  `checkBrandingValue`'s acceptance and then apply `html.EscapeString`. Keep the hex grammar
  for the bundle-read path only, since the minifier's output form is the risk there. The
  alternative (omit the meta and log once) is also acceptable. Either way, the task must
  name the report and add a Go row for it.

### M5. Task 5's rendered rows have no fixture host, and the exit gate's "every control" is proven only on stubs

- **Location:** `:871-876` (Files), `:908-915` (acceptance), `:598-599` (review focus 3),
  exit gate item 8 at `:1178`.
- **Defect:** the focus row requires a `btn-secondary`, `btn-warning`, `btn-neutral`,
  `btn-primary`, and `input` on "each stub". The stubs carry an `h1` and nothing else.
  "Shadows, borders, and the keylines read back per theme" names no element (no stub has a
  card, alert, badge, or framed list) and no expected value. The plan's default of
  extending the stubs "with no copy change" means shipping a control sampler in production
  pages, which the pass rules out ("No ... styleguide", `:46-47`). The fallback of a fixture
  route under `web/src/routes/` would ship that route's chunks in the embedded bundle. The
  larger gap: after Tasks 8 and 9 replace the stubs, no criterion re-runs the focus row over
  the real pages' controls (the fieldset input, submit, interstitial control, toggle, and
  sign-out). Exit gate item 8, "on every control", therefore rests on controls that no
  longer exist.
- **Proposed fold:** (a) Task 5's rows inject their fixture controls into a hydrated stub
  through `page.evaluate`, which creates DOM with the real bundle's classes and stylesheet.
  Nothing ships, and each injected element is asserted present by role before any
  measurement. (b) Each surface row names its literal expectation from VI's depth table
  (`:204-206`) and the refinement table (`:329-330`). One example: a `label` button's shadow
  is `2px 2px 0` of neutral at 55%, read back as painted pixels at the offset. (c) Rewrite
  the focus row as "every tabbable element on every served page", enumerated by DOM query
  with a nonzero count asserted per page. It runs in the manifest from Task 5 on, so Tasks 8
  and 9 inherit it, and the rows for Ruling 4 and for the 44px floor ride the same
  enumeration.

### M6. Task 8's "each login outcome" names fixtures for three outcomes of nine

- **Location:** `:1051-1053`; outcomes at `internal/auth/handlers.go:52-62`; the deck at
  `:976-981`.
- **Defect:** the criterion names sources for `sent` (fulfilled), `failed` (real), and
  `unknown` (real). It names none for `requested`, `rate_limited`, `invalid_address`,
  `too_large`, `malformed`, or the cross-origin 403. A page that maps every non-`sent`
  answer to one generic line passes Task 7's global coverage test and Task 8's three rows.
  The live-region class for `unknown` (polite or assertive) is also unstated. The fixture's
  answer mode is unnamed too, and it decides which outcomes are real: `requested` is
  produced only under `GenericAnswer`, and `failed` answers synchronously only under the
  kind answer (`handlers.go:183-188`, `:266-273`).
- **Proposed fold:** add a state table (outcome, source, expected deck key, live-region
  politeness) with one row per `requestLink` and `consumeLink` outcome. Sources: real under
  the named kind-answer fixture where the server can produce the outcome, `route.fulfill`
  with the `handlers.go:157-164` shape otherwise. Each row asserts the rendered text equals
  the deck constant for that key. The key is read from the deck module, which is fine,
  because the property under test is the mapping. Name `GenericAnswer=false` in R-5ciii4's
  fixture environment.

### M7. The non-admin row names no way to navigate client-side

- **Location:** `:1096-1097`, `:1105-1106`, review focus 6 at `:604-605`; R-5cii19.
- **Defect:** `<prefix>/members` composes the admin gate before the document, so
  `page.goto` of the members URL with a member session returns the terminator's plain-text
  404. That response is not the error page's sentence, and it would pass "no element
  containing the admin's address" trivially. The threat the row exists for is SvelteKit's
  client router rendering the public members component. No page a member can reach links to
  members (the signed-in state has "no destination", `:986-987`), and `goto` from
  `$app/navigation` is not reachable from `page.evaluate`. As written, the row either
  cannot run or runs against the wrong response.
- **Proposed fold:** name the mechanism. Sign in the member through a minted link. On the
  hydrated interstitial (identity asserted), inject `<a href="members">` via `page.evaluate`
  and click it, so kit's link interception performs the client navigation. Assert that the
  request log shows no document request for `members` (proof the navigation was
  client-side) and one roster `GET` answered 404. Then assert the not-found sentence, no
  admin address, and no add-member `form` or `fieldset` rendered. The last assertion covers
  "the page renders nothing it did not fetch" (`:310-312`). Add the member session to the
  privileged-field table's harness row (`:343`).

### M8. Task 9 mutation 4 cannot fail

- **Location:** `:1114`.
- **Defect:** "Add an admin checkbox → the API answers `admin_not_grantable` and the add row
  fails." An unchecked checkbox is not a successful control, so the form does not submit it.
  The add row never checks it, the POST carries three fields, and `added` comes back. The
  mutation survives, so the row does not guard a privileged-field channel.
- **Proposed fold:** add two assertions. First, the add form contains no control named
  `admin` (DOM query). Second, the intercepted POST body's key set is exactly `address`,
  `display_name`, `streamer_username` (or a subset without the optional key). The mutation
  then fails on the DOM query. Keep the API's own refusal as defense in depth, not as the
  row's proof.

### M9. The render-and-read can pass with a partial matrix, unread cells, and an unread fold

- **Location:** `:1132-1145` (Task 10); the `see-it-before-geoff-tests-it` standing bar.
- **Defect:** four vacuous paths exist. (1) "The row count is asserted against the matrix"
  and "the matrix count equals pages × states × 5 × 2" compare the matrix with itself. The
  pages and states are "every page and every state Tasks 8 and 9 name", which is never
  enumerated, so a one-page matrix passes. (2) No capture asserts its fixture state (page
  identity plus the state's distinguishing text) before capture, so a screenshot of an
  un-booted page, an error, or a spinner counts. (3) The reader gets contact sheets. A
  2560px full-page capture downscaled into a sheet cannot show a 2px focus ring or a keyline
  defect, and the reader's verdict carries no per-cell coverage, so "accept" can mean
  "looked at thumbnails". (4) After the fold, the matrix "re-runs", but nobody reads the
  re-captured pages before they go to Geoff (`:1234-1235`).
- **Proposed fold:** (1) the plan enumerates the capture list as literals: page and state
  pairs, with the expected count written as a number. The Go manifest (M1) holds it. (2)
  Every capture row asserts its fixture state first. (3) The reader reads every page and
  state at native resolution at 390 and 1440 in both schemes, with contact sheets for the
  other widths. Its report lists every cell it read, and the conductor rejects a report
  that is missing a cell. (4) After any fold, one fresh read of the changed cells comes
  before Geoff's look.

### M10. Reproducibility is an exit-gate claim with a one-time proof

- **Location:** `:549-561` (R-5ciii7), `:778`, mutation 1 at `:783`, exit gate item 6 at
  `:1175`.
- **Defect:** the only gate row compares the served `version.json` with the built one. Both
  come from the same embedded tree, so that row passes whether or not the build is
  deterministic. The two-build comparison is "a one-time evidence row in Task 3's report".
  Mutation 1 ("leave `kit.version.name` unset → the two-build comparison differs") therefore
  names a check no gate runs. After Task 3, a change that brings back nondeterminism (a
  `Date.now()` anywhere in config, an unsorted directory walk) passes every gate. The claim
  "Two builds of one tree are byte-identical" is listed as an exit-gate assertion without a
  gate behind it.
- **Proposed fold:** either gate it (the `web` target builds a second time into a scratch
  directory and compares tree hashes, at a cost of about 7s from 5c-ii's measured build
  times), or relabel exit gate item 6 as evidence recorded once and drop mutation 1's
  gate-shaped wording. Recommendation: gate it. Also, R-5ciii7's input walk must state a
  sorted order: `fs.readdir` promises no order, and packaging on another filesystem is the
  stated reason for the design.

## Minors

- **m1 (`:784-785`, `:769-773`).** Mutation 2 is caught by "the diff-reviewer", not a test.
  The W12.1 refusals are each proved by a one-time scratch seed. Fold: keep one
  gate-resident negative row for the cheapest refusals. For example, point the script at a
  nonexistent executable path through an env override and assert a nonzero exit plus the
  message. Mutation 2 then fails that row.
- **m2 (`:686-688`).** Mutation 2 (dropping the explicit `Content-Type`) survives on this
  workstation: `/etc/mime.types:1798` maps `woff2`, so `ServeContent`'s host lookup supplies
  `font/woff2` anyway. Fold: pick the "header set before `ServeContent`" form now (a
  recording `ResponseWriter` asserting the header is present at `WriteHeader`) rather than
  leaving it to the task.
- **m3 (`:524`, `:1086`, `:339`).** The toggle "updates the `theme-color` meta", but no value
  source is named. The `label` variant's document carries one meta, and a literal
  `#110f0d` in a component fails G1. No criterion checks the update, and the
  privileged-field table does not list the toggle as a writer of the meta. Fold: the value
  is `--color-neutral` read from computed style after the attribute change, which also
  honors branding. Add a row that reads the meta's content after the click, and add the
  channel to the table.
- **m4 (`:1100-1102`).** The toggle row names "a starting stamp" without naming which ones.
  Fold: name the four starts (`system` under light and under dark emulation, cookie `label`,
  cookie `label-night`), each with the expected checked state and the value written.
  `system` under dark starts checked, and the click writes `label`.
- **m5 (`:998-1000`, `:988-991`).** The outcome-coverage test parses Go source for
  constants. A regex that matches nothing passes. Fold: assert the parsed set's size and
  that it contains named members (`signed_out`, `admin_not_grantable`,
  `write_unconfirmed`). The AI-tells scanner similarly reports and asserts a nonzero
  scanned-string count.
- **m6 (`:449-451`, `:722-727`).** G9's "every direct-text color" is not a closed list.
  Fold: name it (base-content, and each `*-content` measured against its own fill only)
  and name which colors are excluded as fill-only under W2.6's marker.
- **m7 (`:556-558`).** The version hash's input set omits `svelte.config.js`, `vite.config`,
  and `tsconfig.json`. A config-only change that alters output keeps the old version name.
  Fold: include them, or hash the built output instead.
- **m8 (`:136-137`, `:915`).** Ruling 4's "yes builds" row ("no icon-only control carries a
  primary fill") has no icon-only control to inspect until Task 9's toggle. In 5c-iii it
  passes on an empty set. Fold: M5's per-page enumeration carries it into 5c-iv with a
  nonzero-count assertion there.
- **m9 (`:484-485`).** `--scope-changed` runs the rendered leg only for `web/`,
  `internal/serve/`, and `cmd/dubplate-server/`. The 5c-iv rows exercise `internal/auth`'s
  request, consume, and roster handlers, so a change there skips the rows that cover it.
  Fold: add `internal/auth/` and `internal/identity/`.
- **m10 (`:462-464`).** The fixture environment does not name the answer mode (M6), the
  link and consume rate limits, or how many sign-ins the manifest spends. Rows in both
  schemes across Tasks 8 and 9 sign in repeatedly from `127.0.0.1` and can trip
  `rate_limited` partway through the run. Fold: name the limits the fixture sets and the
  sign-in count.
- **m11 (`:1163-1185`).** The 5c-ii carry "the prefix proof's hydration half is evidence
  from a scratch probe, not a gate" is closed only implicitly, by the rendered rows running
  under a prefix. Fold: name it in the exit gate so the close can map it.
- **m12 (`:1138`, `:1144`).** The touch-target floor must assert a nonzero control count
  per page, or a page with no controls found passes.
- **m13 (`:1045-1049`, `:1060`).** The I2 row checks request origins, not the token's
  absence from request URLs and headers during load. Mutation 3's "ordering row" is not
  among the acceptance criteria. Fold: assert that no request's URL or headers contain the
  token, and that `location.search` is read empty before the click (the check that is the
  ordering row).
- **m14 (`:1040-1041`).** The error page has no acceptance criterion (identity type,
  surface, and the sentence). Fold: one rendered row per status, reached by client
  navigation to an unmatched path using the same injected-link mechanism as M7.
- **m15 (`:157`, `:263`, `:273`).** Ruling 6 says "about 13.4M". The two projections sum
  to 13.25M (7.15 + 6.10). This is cosmetic, but a ruling's number should match its own
  table.

## The six rulings

1. **The split.** Rulable. The evidence (task count, the cut at a Files seam, the precedent
   of 5c-ii) is sufficient. Caveat: one supporting reason, "every Task 1 through 5 surface
   is exercised by the rendered harness against the 5c-ii stubs", is false for the site
   name (M3) and the component classes (M5) until those folds land. The recommendation
   survives the correction.
2. **Playwright 1.63.0.** Rulable. License, engines, the pinned Chromium revision present
   in the cache, cadence, adoption, and the declined `@playwright/test` are all evidenced.
3. **Login without the photo.** Rulable. VI's phone layout already rules the omitted-photo
   state, and the licensing rules make a placeholder unshippable. This is a product choice
   for Geoff, and the options are complete.
4. **WCAG 1.4.11 reading.** Rulable on the quoted W3C Understanding text and VI's measured
   5.15:1 label contrast. Caveat: its "yes builds" row is vacuous in 5c-iii (m8). Strictly,
   this is a conformance reading that could be settled on evidence without an owner ruling.
   Presenting it as a ruling is harmless.
5. **Defer the sleeve, grain, rim, and Permanent Marker to rung 7.** Rulable. The
   no-consumer argument and the Apache-2.0 license fact are both evidenced, and DL R5.2's
   admission of Apache 2.0 keeps the face eligible.
6. **The reservation raise.** Rulable. The per-task rate re-derives from 5c-ii's recorded
   9.2M actual, and the arithmetic checks except for m15's rounding.

## Checked and sound

- DL G5's placeholder-tree refusal and its nonzero face count are carried into Task 1
  verbatim (`:668-679`).
- The G10 reinterpretation includes the control that keeps it from passing vacuously:
  stamping built-in `light` and seeing `--radius-box` move (`:579-582`).
- G7's without-branding control row (`:570-571`) keeps the with-branding row from reading a
  default.
- I2's `history.state` concern is grounded in kit's source (`:1032-1036`), and mutation 1
  (`:1058`) fails for a real reason.
- The toggle's cookie-write rows read served bytes after a reload (`:1100-1101`), not only
  the client attribute.
- The handoff's carried items (G4, G5, G7 computed, G10, a11y W4 through W6, the 1.4.11
  question, S11, S12, I2, A4, the `data:` noise-texture confirmation) each have a task.
  The sleeve, grain, rim, and marker items the handoff lists are routed by Ruling 5.
