# Draft HISTORY entry: rung 5, pass 5c-iii

Draft only. The conductor folds this into `docs/HISTORY.md` and does not commit it as a
standalone file.

## Rung 5, pass 5c-iii (the fonts, the theme data, the rendered gates, the document head, and the component layer): closed 2026-09-25, lane 1

Executed against
`docs/superpowers/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md` (revision 4,
ratified by Geoff 2026-09-25, Rulings 1-10 all yes/as stated), in the main tree, from `fc52232`
(the plan's own drafting commit; code work starts at `83d5ae9`, which raised window 2's cap to
run 5c-iii through this close). `git log --oneline --grep="5c-iii" -i | wc -l` counts 31
commits tagged for this pass, including its plan-authorship and review commits; the code and
docs commits landing this entry's work are named below.

### What landed

Task 1 (`0f9a57d`; fix round `f5d46cb`) embedded the three ruled faces (Alfa Slab One, Work
Sans, IBM Plex Mono) as woff2, the license manifest, and the `fonts` route with G5. Ruling 9
settled the WOFF2/OFL question for Alfa Slab One's Reserved Font Name: fontTools' encoder drops
the source's empty `DSIG` stub and sets `head` bit 11 unconditionally, so the shipped face
counts as "unchanged except for WOFF compression" under the OFL FAQ.

Task 2 (`5dc0714`; fix rounds `3c6ac4c`, `253e36d`, `3a5bfef`) landed VI 1.2's theme data: the
font stacks, the five inks, the surface tokens, and per-surface G9. Two conductor-ruled second
fix rounds landed here, pinning G9's pair count, its fill-only list, and its direct-text check
against DL R2.6 independent of `DIRECT_TEXT_COLORS`.

Task 3a (`5d54a9d`) added the rendered leg: a build-tagged Go test (`rendered_test.go`) driving
a Playwright 1.63.0 script (`web/scripts/rendered/run.ts`) against a real Chromium, the
reproducibility gate (two builds compared byte for byte), and the first-paint rows. Task 3b
(`591fe53`; narrowed `3db2708`) added G7's computed half and G10 to the same leg, comparing
colors as painted pixels and narrowing the listener exemption.

Task 4 (`e320b80`) added the document head: `theme-color`, the site-name and robots metas, the
mark and icon links, and the root-absolute gate.

Task 5 (`809cf80`; fix rounds `6799598`, `610ad61`) added the component layer: the focus rule,
the ink utilities, the printed surface, and the refinement rules. The second fix round asserted
focus-row coverage by identity and pressed-state reach, a conductor-ruled second round.

The `code-simplifier` round (`4e1467e`) extracted `readThemeBlocks` (G2) and `measureRatio`
(G9) so duplicated luminance/ratio preambles share one implementation, and had `run.ts`'s
`paintedPrimary` resolve `--color-primary` in the browser rather than re-implementing the
canvas-paint logic inline. No behavior change; every assertion, row, and manifest entry held.

### The pass-end reviews and their fold

Three reviewers ran and are recorded at `944e235`:
`docs/superpowers/research/2026-09-25-rung-5c-iii-review-svelte.md` (1 blocker, 6 warnings, 13
suggestions), `-review-a11y.md` (1 blocker, 8 warnings, 9 suggestions), and
`-review-security.md` (2 medium, 5 low, 4 informational). The fold record
(`docs/superpowers/research/2026-09-25-rung-5c-iii-review-fold.md`, `7ca3d53`) counts 49 raw
findings, 10 repeats, 39 distinct: 30 folded (two declining a sub-part, each with its own
reason), 3 refused (I3, I4, svelte S6), and 6 routed or recorded (M2 to rung 12, svelte S12 to
Task 8, a11y S7 to Task 9 and rung 7, svelte S13 to the rung 7 design review, a11y S2 and S4 to
errata below).

The fold resumed after a power loss stopped the first fold agent before it committed anything;
the resumed agent kept and fixed that agent's warm, uncommitted `run.ts` test rows rather than
redoing them. Fold commits: `81988b1` (the component layer, 28 rendered rows, the prefix watch
on every row, the harness env allowlist — svelte's blocker, most warnings and suggestions,
security L5), `2c880db` (the static gates: G1, G4, G5's comment, G9, root-absolute), `75fde16`
(the document rewrite, the site-name grammar, the theme-color grammar, SVG CSP, variant CSP
equality — security L1 through L4, I1, I2), `8588dd0` (VI section citations, FONTS.md's full
commit, the rendered script's deadline raised to 300s after an under-load run measured past
150s). Each commit landed on a full heavy gate that exited 0. The fold's mutation proof ran
every new rendered row and Go-side check against `8da75b8`'s pre-fold code and confirmed each
failed for its stated reason before the fix turned it green (fold record, "Mutation proof"
table).

The fold's own `diff-reviewer` returned `fix`: the alert halo rule sat before the button's and
inputs' own `box-shadow` rules and lost at equal specificity, so a focused control inside a
status alert got no halo; and the fill-keyline rule painted `.btn-outline.btn-<status>`'s
border black on `label-night`, so an outline or dash button lost its edge. The fix round
(`db1e111`) moved the halo after every `box-shadow` a focusable carries and painted an outline
or dash button's border in its status ink. It also restored `staticFile`'s doc comment,
extended root-absolute's srcset sink to a later candidate in the list, and recorded itself in
the fold record. A second `diff-reviewer` accepted `db1e111`, reproducing both red-first
proofs: against the pre-fix order the alert rows failed at `1.22:1`/`1.33:1` (label) and
`1.03:1`/`1.02:1` (label-night); against the pre-fix border rule every status outline measured
`[0,0,0,255]` on `label-night`. `db1e111` added `settleFocusTransitions`, which waits out
DaisyUI's 0.2s `box-shadow` transition before the focus walk measures a ring.

Simplifier round 2 (`3fd6c0b`) extracted `paintedProperty` from three repeated closures in
`run.ts`; no behavior change.

Three merge-side `go-architecture-reader` reads ran at HEAD `db1e111`, all "workmanlike":
`docs/superpowers/research/2026-09-25-rung-5c-iii-architecture-internal-serve.md`,
`-internal-config.md`, `-cmd-dubplate-server.md`. Each names every finding with its `file:line`
and a routing line; the headline findings and their routes are in "What stayed open" below.

### Gate evidence

The rendered leg's wall time rose through the pass: about 17s before Task 3a, to about 55s on
an idle machine after the fold's added rows, with one run under another lane's load measured
past 150s before it was killed (the script's deadline is now 300s, `RENDERED_TIMEOUT` 7m). A
standalone rendered-test run at `db1e111` measured 107.4s. The full heavy gate
(`bash scripts/check.sh` through `cairn-run-gate`) runs about 3-4 minutes with about 560 MiB
peak, launching Chromium on every invocation since Task 3a.

### What stayed open, and where it goes

**Architecture-read findings, routed beyond the cleanup pass:**

- `internal/serve`: `NewTrustedProxyConfig` (`clientaddr.go:25`) has no production caller,
  `cmd/dubplate-server/serve.go:244` declares a zero `TrustedProxyConfig`, and
  `ProxyMode`/`ModeDirect`/`ModeTrustedProxy`/`String` (`clientaddr.go:36-52`) are unused. Routed
  to 5d-i Task 1, whose Task 0 must read this finding before wiring the trusted-proxy class.
  Rung 12 hooks shipped early (`SetFallthrough` at `serve.go:140`, fields `:31-32`;
  `NewFrontExemption` at `strip.go:242`, no caller; `NewStrip`'s header param at `strip.go:300`,
  always `""`) are recorded for rung 12, or for the cleanup pass to unexport now.
- `internal/config`: the package doc (`config.go:6-46`) calls backup an integration class,
  which is stale; `keys.go:186`'s `IntegrationClass.Rung` field is set but never read. Noted to
  6c's Task 0/2, since 6c edits this registry and must not build on the stale doc.
- `cmd/dubplate-server`: `rendered_test.go`'s `linkRequests`/`consumptions` fields
  (`:44-68,328-353`) are never set, so `checkRateBudget` cannot fail — routed to 5c-iv's refresh
  read, since Task 8 adds login-POST rows that spend a link and give the check something to
  fail against. The magic-link key names kept in agreement by hand across four files
  (`serve.go:25-37`, `auth.go:46-60/268-285/365`, `doctor.go:144`) route to 5d-i Task 6 or the
  cleanup pass.

**Review items recorded, not folded:**

- a11y S2 and svelte's related note (per-variant `color-scheme`): a `label` cookie stamp still
  carries `color-scheme: light dark` (W5.10) rather than the resolved scheme. Erratum owed to
  VI.
- a11y S4 (a W5.8 placement note): the component layer beats DaisyUI from `@layer utilities`'
  implicit sublayer rather than declaring itself "unlayered by necessity" as W5.8 describes;
  only the reduced-motion kill switch needs to be unlayered. Erratum owed to the web surface
  standard.
- svelte S13 and the related keyline observation: the input inset shadow and `label-night`'s
  keyline-vs-specimen edge color are recorded for the rung 7 uploader's attended design review.

### Errata owed

Draft text, paste-ready into each named document. The conductor applies these as erratum
blocks, not as edits to the ratified prose above them.

**Already landed, not owed by this close.** VI's revision 1.3 (`45b80cb`, landed before this
pass's Task 6) already carries three of the errata this plan's Task 6 section lists: the clause
number at VI `:141`/`:152` (R5.4 corrected to R5.5), Permanent Marker's license at VI `:149`
(Apache-2.0, not OFL), and the header focus ring's page-colored halo per Ruling 7 at VI `:114`.
No further edit to VI is needed for these three.

**1. The rung 5 spec's 5c row: the third split.** Append after the existing 2026-09-25 erratum
block at `docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md:2470-2486` (which records
5c-ii's own split into 5c-ii and 5c-iii):

> Erratum, 2026-09-25 (pass 5c-iii close): **the 5c-iii row splits again, into 5c-iii and
> 5c-iv.** The erratum above described 5c-iii as carrying "the fonts, the three pages with
> their string deck, the rendered gates, and the render-and-read" under its own plan's ceiling.
> `docs/superpowers/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`'s Ruling 1
> splits that scope again at the served-header-versus-pages boundary: 5c-iii carries the fonts,
> the theme data, the rendered harness and its reproducibility and first-paint gates, the
> document head, and the component layer, and closed under a 10.8M ceiling at about
> [SPEND-PLACEHOLDER, conductor fills] in subagent tokens (this pass's own HISTORY entry has the
> figure). 5c-iv carries the string deck, the three pages, and the render-and-read, under this
> plan's own remaining tasks (7 through 11) and an 8.5M ceiling, taken after a just-in-time
> refresh read at this close rather than a freshly drafted plan (Ruling 1's declared deviation
> from CLAUDE.md's just-in-time rule). 5c's four-pass total reservation moves from about 22M
> (5c-ii's erratum) to about 14.8M for 5c-iii and 5c-iv together (8.30M and 6.50M projected);
> rung 5's own reservation moves from about 41M to about 51M (this plan's Ruling 6).

**2. The web surface standard's section 15 line.** Append to the existing 2026-09-25 erratum
block at `docs/superpowers/specs/2026-09-12-web-surface-standard.md:1790-1794` (the "second
split" erratum):

> Erratum, 2026-09-25 (pass 5c-iii close): **the third split.** 5c-iii itself split into
> 5c-iii and 5c-iv at the served-header-versus-pages boundary (rung 5 spec's matching erratum
> above). "5c-iii carries its own re-derived ceiling" now names two plans' remaining scope, not
> one: 5c-iii closed under 10.8M, and 5c-iv carries the same plan's Tasks 7-11 under 8.5M.

**3. `docs/superpowers/specs/2026-09-07-design-language-standard.md`'s R5.5.** Append after
`R5.5` (`:475-479`):

> Erratum, 2026-09-25 (pass 5c-iii close): **R5.5 carries no family count.** The visual identity
> spec (Geoff's ruling) names three families (Alfa Slab One, Work Sans, IBM Plex Mono) plus a
> fourth face, Permanent Marker, for one use — four families, four woff2 files. R5.5's "at most
> two families" was a conductor ruling (2026-09-07) that VI's owner ruling supersedes. R5.5 now
> reads "a theme ships at most four woff2 files in total, counting a variable font as one file
> per family," with no family count. `web/scripts/gates/g4.ts`'s face-count row checks four
> files.

**4. The web surface standard's W4.1.** Append after `W4.1` (`:340-346`):

> Erratum, 2026-09-25 (pass 5c-iii close): **the both-schemes half no longer binds from rung
> 7.** Rung 5 now ships two composed themes (VI ruling 8: `label` and `label-night`), so the
> render-and-read (5c-iv Task 10) covers both schemes at rung 5, not rung 7.

**5. The rung 5 spec's add-member "admin checkbox".** Append after the add-member page
description at `docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md:1436-1445`:

> Erratum, 2026-09-25 (pass 5c-iii close): **the add-member page ships no admin checkbox.**
> R-5c13 (5c-i) already refuses an `admin` field on the web channel, and the API enforces it
> (`internal/auth/admin.go:43`, `outcomeAdminNotGrantable`). This was already the 5c-i outline's
> reading (`2026-09-23-rung-5c-pages-member-verb.md:2501-2503`); this erratum records it against
> the spec line itself, which still names the checkbox. 5c-iv's Task 9 ships the page with no
> admin control.

**6. The 5c-ii plan's R-5cii8.** Append after R-5cii8's cache-directive section
(`docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md:856-873`):

> Erratum, 2026-09-25 (pass 5c-iii close): **the members document row is `no-store`, not
> `private, no-cache`.** This plan's cache table named `private, no-cache` for the members
> document, but Fold B of 5c-ii's own review moved it to `no-store`
> (`cmd/dubplate-server/web.go:80`, security L4). The code is right; this plan's row is
> superseded.

**7. The 5c-ii plan's R-5cii22.** Append after R-5cii22's cookie-attribute description
(`docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md:544` and its
neighboring rows):

> Erratum, 2026-09-25 (pass 5c-iii close): **the theme cookie carries a one-year `Max-Age`.**
> This plan named `Path=/; Secure; SameSite=Lax` with no lifetime, making the pick a session
> cookie. The web surface standard's W5.10 ("the pick is lost only with the cookie jar")
> requires a persistent cookie, so 5c-iii's toggle (Task 9's `Set-Cookie`, R-5ciii9) writes
> `Max-Age=31536000` as well. The server's reading of the cookie is unchanged.

**8. VI's `:128-131` under a functional-color branding neutral.** Append after the
`theme-color` paragraph at `docs/superpowers/specs/2026-09-24-visual-identity-design.md:138-140`:

> Erratum, 2026-09-25 (pass 5c-iii close): **the hex pair describes the unbranded bundle only.**
> Under `DUBPLATE_BRANDING` setting `--color-neutral` to a functional color (`oklch()`, `rgb()`,
> `hsl()`), Ruling 10 (this pass) has the document carry no `theme-color` meta at all, with one
> `slog` startup line naming the reason. A branding neutral in a hex form (3, 4, 6, or 8 digits)
> still enters the meta as before. `#1f1b16` and `#110f0d` remain the values for every other
> configuration.

### Budgets

Tokens: 10.8M ceiling, closed at about [SPEND-PLACEHOLDER, conductor fills]M in subagent
tokens. Attended time: [SPEND-PLACEHOLDER, conductor fills] planning misses, [SPEND-PLACEHOLDER,
conductor fills] execution sittings after plan approval (Ruling 7's header halo color was
Geoff's own ruling on a genuine fork, not a sitting).

### What a later pass would be wrong to rediscover

- The rendered leg's wall-time growth (17s to about 55s idle, past 150s under load) is a
  property of the fixture's row count, not a regression to chase; the deadline is now 300s.
- `settleFocusTransitions` exists because DaisyUI transitions `box-shadow` over 0.2s; any new
  focus-row assertion added later must wait it out or it reads mid-transition.
- The alert halo's ordering (after every `box-shadow` a focusable carries) is load-bearing:
  the two rules tie at zero specificity, and source order alone decides the winner.
- An outline or dash button's border paints in its status ink, never the fill keyline; the fill
  keyline is black-adjacent on `label-night` and erases the button's edge if applied there.
- Outline and dash button *text* still paints DaisyUI's raw `--btn-color`
  (`web/src/lib/styles/components.css:172`), unfixed and latent: label secondary and warning
  measure 1.83:1, label-night neutral 1.09:1, label-night primary 3.18:1. This must land before
  any page adopts `.btn-outline` or `.btn-dash` with a low-contrast status — route to 5c-iv's
  refresh read as a precondition for Tasks 8 and 9, else to rung 7. The fixture that would catch
  a regression here covers `.btn-outline` but not `.btn-dash`.
- Other latent, unfixed items found during the fold and not yet gated: the ink border shows
  over DaisyUI's hover fill on outline buttons; a wrapped `<label class="input">` inside a
  status alert gets the halo on the inner input, not the wrapper; `settleFocusTransitions` awaits
  with no bound, so a paused animation would hang the focus walk until the script's 300s
  deadline; root-absolute's new `srcset` pattern false-positives on a base64 data URI that
  happens to start with `/`.
- `internal/serve`'s trusted-proxy machinery (`NewTrustedProxyConfig`, `ProxyMode`) exists but
  has no production caller; a pass that assumes trusted-proxy mode is live must check
  `cmd/dubplate-server/serve.go`'s wiring first.

## 5c-iv handoff

Full detail above ("What stayed open", "Errata owed"). This section names what 5c-iv's Tasks 7
through 11 build on, for conductor action 5's refresh read.

**The four files crossing the Ruling 1 cut, and their state at `db1e111`:**

- `cmd/dubplate-server/rendered_test.go` — the build-tagged rendered-leg test. Carries the
  request-log prefix watch (every row fails if a request left the owned prefix since the
  previous row), the fonts/coverage/focus/surface/link/badge/display rows from Tasks 3a, 3b,
  and 5's fold, and the touch-target seed row from Ruling 8. `linkRequests`/`consumptions`
  fields exist but nothing sets them yet (the architecture-read finding above) — 5c-iv's Task 8
  is the first task that spends a link and gives `checkRateBudget` a row to fail against; the
  refresh read decides whether Task 8 wires these fields or whether they're dropped until then.
- `web/scripts/rendered/run.ts` — the Playwright harness. Carries `settleFocusTransitions`, the
  env allowlist (`PATH`, `HOME`, `TMPDIR`, `LANG`, `LC_ALL`, `TZ`, three `XDG_*` vars,
  `PLAYWRIGHT_BROWSERS_PATH`), `paintedProperty`/`paintedColor`/`paintedPrimary`, and the fixture
  page's full component set. 5c-iv's pages are real routes the harness will navigate to instead
  of (or beside) the injected fixture; Task 8's login-page rows and Task 9's header/toggle rows
  are new callers of this harness, not new copies of it.
- `Makefile` — the `web` target's rule set, including the rendered leg's now-300s deadline.
  Task 7 modifies the same target to add the register-gate rule; the refresh read should confirm
  the two edits compose without reordering an existing rule.
- `web/src/lib/styles/components.css` — the component layer: the focus rule (with the alert-halo
  ordering fix), ink utilities, the printed surface, refinement rules, and the still-latent
  outline/dash text-color gap named above. Every 5c-iv page consumes this layer as-is; Task 8 and
  9 must not re-declare a selector this file already owns (G1's `-content` guard and the
  `:where()` specificity discipline both gate on it).

**5c-iii review-fold items that touch a Tasks 7-10 surface, beyond the four files:**

- a11y W8 / svelte S7 (G9, the `label-night` keyline on `base-300`): carries a
  `contrast-exception` marker naming `base-300` because no page places a control there yet.
  Task 9's add-member page or Task 8's login page, if either places a control on `base-300`,
  must either clear 3:1 or extend the marker's justification.
  the header halo per Ruling 7 stays with Task 9 specifically (the header is Task 9's).
- a11y S7 / svelte S7 routing: mono utilities and the header link hover are routed to Task 9
  (header) and rung 7 (mono surfaces) — Task 9's refresh read should confirm the header hover
  state exists before the render-and-read.
- svelte S12 / a11y S9 (stub copy): routed to Task 8 directly — the login page's copy must not
  ship the fixture's placeholder strings.
- L2's `{@html}` note: routed to Task 7 — the site name renders by text interpolation only,
  never `{@html}`, in the string deck's consuming components.
- The touch-target rule (Ruling 8): seeded in the rendered leg with a 20px control; Task 10's
  full render-and-read is where every real control on every real page is measured against it,
  and Task 9's header controls take their own 44px row.
