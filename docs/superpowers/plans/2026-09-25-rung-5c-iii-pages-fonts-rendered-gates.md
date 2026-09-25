# Rung 5, Pass 5c-iii (Fonts, the Identity's Component Layer, the Rendered Gates, and the Three Real Pages) Implementation Plan

Status: **revision 1, 2026-09-25: draft, awaiting review.** Not yet through the
`spec-plan-review` cycle (three lenses, one fold, one verification read, the prose
review). No ruling below is answered.

Sources: the 5c-ii close (`docs/HISTORY.md`, top entry, drafted in
`docs/superpowers/research/handoffs/draft-HISTORY-5c-ii.md`), whose "What stayed open",
"Declared coverage gaps", and "The 5c-iii handoff" sections are this pass's scope; the 5c-ii
plan (`docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`, revision
2.4), whose R-5cii2 (the relative base), R-5cii7 (the CSP builder), R-5cii8 (cache
directives), R-5cii15 (where each DL gate lands), R-5cii19 (the members route is public, the
API is the boundary), R-5cii21 (exact paths), and R-5cii22 (the theme stamp and cookie) this
pass builds on; the three 5c-ii reviews
(`docs/superpowers/research/2026-09-24-rung-5c-ii-review-{svelte,security,a11y}.md`); the
visual identity spec (`docs/superpowers/specs/2026-09-24-visual-identity-design.md`,
revision 1.2, `782d864`, VI below), the look every page builds to; the design-language
standard (`docs/superpowers/specs/2026-09-07-design-language-standard.md`, DL below); the web
surface standard (`docs/superpowers/specs/2026-09-12-web-surface-standard.md`, revision 2.1,
W below); the rung 5 spec (`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`), "The
pages" (`:1413-1466`), "Consumption" (`:1051-1080`), and "The post-login destination"
(`:1268-1273`); the 5c-i plan's page outline
(`docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md:2494-2511`); and the 2.0
streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`), "Forward hooks"
(`:450`) and "Prior art" (`:522`). Code pointers were resolved at `b9e281b`.

**Owner rulings in force.** dubplate carries its own visual identity and matches no streamer
(Geoff, 2026-09-24): no reference screenshots, no visual-fidelity chain. VI is the look (Geoff,
2026-09-25, CLAUDE.md's web stack rule): every `web/` task reads it, and it outranks any styling
habit. Every user-path page is rendered and read before Geoff sees it
(`see-it-before-geoff-tests-it`).

**What this pass is.** 5c-ii shipped the delivery path with three structural stubs. This pass
makes the stubs real and proves them in a browser. It lands the three ruled faces as embedded
woff2 files with G4 and G5; VI revision 1.2's theme data (five inks, the surface tokens) with
per-surface contrast rows; the rendered leg of the gate (Playwright against the real Go shell),
carrying G7's computed half, G10, and the first-paint proof per preference and cookie; the
document head (the `theme-color` metas, the site name, the mark and favicon); VI's component
layer (the one focus rule, the ink utilities, the printed surface, the refinement rules); the
string deck with its register gate and W12.3's AI-tells rule; the login page, the consumption
interstitial, and the add-member page with the signed-in header and its theme toggle; and the
render-and-read at W4.1's five viewports in both schemes.

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

Six yes/no questions, each with a recommendation. The rung 5 spec's rulings, the 5c-i and
5c-ii plans' rulings, VI's owner rulings, and W's section 16 stand and are not reopened.

**1. Does the remaining 5c scope split again, into 5c-iii (Tasks 1 through 6) and 5c-iv
(Tasks 7 through 11), under ceilings of 9.4M and 8.0M?** Recommendation: yes.

**Reason:** the scope is nine build tasks plus two closes, past the brief's eight-task line.
The cut sits after Task 5, where the Files change from shared infrastructure (fonts, theme
data, the rendered harness, the document head, the component CSS) to the pages that consume
it. Each half gets its own pass-end review: the infrastructure is read before three pages
build on it, the argument 5c-ii's own split made. Neither half exports ahead of a caller:
every Task 1 through 5 surface is exercised by the rendered harness against the 5c-ii stubs.

**What yes builds:** Tasks 1 through 6 as 5c-iii, closed by Task 6. 5c-iv executes Tasks 7
through 11 of this same plan after its own staleness pre-task (Task 0's checklist re-run
against the 5c-iii close), with no second plan drafted. The rung 5 spec's 5c row and W
section 15 take a third split erratum at 5c-iii's close. **What no builds:** one pass of
eleven tasks under a 15.9M ceiling (flag 12.7M); Task 6 becomes a STATUS checkpoint with no
reviews, and the pass-end reviews run once, after Task 10.

**2. Does `playwright` join `web/package.json` as an exact `devDependency` at 1.63.0?**
Recommendation: yes.

**Reason:** W12.4 names Playwright as the rendered driver, and W12.6 rules that "Playwright
drives its own downloaded Chromium rather than the workstation's layered `chromium-browser`".
5c-ii's owner-touchpoint list makes any web dependency beyond R-5cii16's pins a question.
Evidence, measured 2026-09-25: `npm view playwright version` answers `1.63.0` (published
2026-09-04; 1.60 through 1.63 landed monthly); `playwright-core@1.63.0`'s `package.json`
declares `"license": "Apache-2.0"` and `"engines": {"node": ">=20"}`; its `browsers.json` pins
`chromium` revision `1243` (`browserVersion` `153.0.8010.12`), which is already present in
`~/.cache/ms-playwright/chromium-1243`; npm counted 69,104,951 downloads in the week ending
2026-09-21. The package is the library only; `@playwright/test`'s runner is declined because
the harness is one script driven from a Go test (R-5ciii4), and a second runner is a second
config to keep in step.

**What yes builds:** Task 3 as written. **What no builds:** the rendered leg cannot exist as a
gate, since W12.1 forbids a skipped check; G7's computed half, G10, the first-paint proof,
and the render-and-read carry forward again.

**3. Does the login page ship without the period photograph, as a single column (mark,
wordmark, the site name as heading, one email field, the sign-in button), until the photo
curation pass supplies licensed photos?** Recommendation: yes.

**Reason:** VI ruling 5 puts a rotating photograph on the login page, but VI's "Photography"
section scopes curation as "a separate pass with no fixed cap", and no photo, manifest, or
CREDITS file exists. Shipping a placeholder photo would ship an unlicensed or unrepresentative
image. VI already describes the phone layout with the photo "omitted", so the single column is
a ruled state, not a new design. Error pages keep VI's plain sentence without the closed-club
photo, for the same reason.

**What yes builds:** Task 8's login page as the single column, with the two-column split
routed to the curation pass. **What no builds:** 5c-iv blocks on the curation pass, or Geoff
names one shipped photo and its license for Task 8 to carry with its film-edge caption.

**4. Does VI's reading of WCAG 1.4.11 stand: a `label-night` primary button on `base-200` or
`base-300` is identified by its label text, so it takes no `--ink-primary` keyline?**
Recommendation: yes.

**Reason:** VI (`:117-121`) asked the 5c-ii accessibility review to confirm or refute the
reading; the review measured the fills (2.94:1 on `base-200`, 2.43:1 on `base-300`) and left
the reading open. The W3C's Understanding document for SC 1.4.11 (WCAG 2.2, fetched
2026-09-25) settles it: "This success criterion does not require that controls have a visual
boundary indicating the hit area. If a control has visible content (such as text or a
sufficiently contrasting icon), which helps users identify the presence of the control, then a
border or other indication of the overall boundary of the hit area is not required, as is
therefore not subject to non-text contrast requirements." A primary button's text measures
5.15:1 on the fill (VI `:119`). The reading fails only for an icon-only primary control, and
this pass ships none (the theme toggle is a ghost `swap` in `base-content`). VI's refinement
row still gives color-filled objects a black keyline on `label-night` (`:330`) as a design
choice, not a conformance fix.

**What yes builds:** no keyline beyond VI's refinement row; Task 5's rendered rows assert no
icon-only control takes a primary fill. **What no builds:** a 2px `--ink-primary` keyline on
every `label-night` primary fill over `base-200` and `base-300` (VI `:120-121`).

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

**6. Is the 5c reservation raised: 5c-iii and 5c-iv together project about 13.4M of execution
against 5c-ii's provisional 5.0M for this scope?** Recommendation: yes.

**Reason:** CLAUDE.md's ceiling rule re-derives from the last actual. 5c-ii projected 7.0M and
closed at about 9.2M over eight tasks, about 1.15M per task once its review folds (about
1.18M), simplifier, close, and architecture reads are amortized ("Token ceiling"). 5c-ii's
provisional 5.0M assumed six tasks at 0.725M. This scope is eleven tasks with the first
browser tooling. Rung 5's reservation moves from about 41M to about 50M.

**What yes builds:** the ceilings below. **What no builds:** the conductor cuts scope to fit
the old figure; the render-and-read (Task 10) and the AI-tells rule would be the first
candidates, and each is a ratified obligation, so the cut takes an erratum.

---

### Spec deviations that take effect on approval

Approving this plan approves these. The close drafts each erratum (Tasks 6 and 11).

- **DL R5.5 ("at most two families and at most four woff2 files").** VI's type section is
  Geoff's ruling and names three families (Alfa Slab One, Work Sans, IBM Plex Mono) plus a
  fourth face for one use (Permanent Marker), four files. R5.5's family count was a conductor
  ruling (2026-09-07), so VI governs: R5.5 reads "four woff2 files", with no family count.
  G4's face-count row checks four files. VI `:141` cites "R5.4's four files"; the clause is
  R5.5.
- **R-5cii8's members row.** The 5c-ii plan's table reads `private, no-cache`; Fold B's L4
  moved the members document to `no-store` (`cmd/dubplate-server/web.go:80`). The code is
  right; the plan row is superseded.
- **W4.1's "The both-schemes half binds from rung 7, because R5's R36 gives rung 5 one composed
  theme."** Rung 5 now ships two composed themes (VI ruling 8), so the render-and-read
  (Task 10) covers both schemes.
- **The rung 5 spec's add-member "admin checkbox" (`:1438`).** R-5c13 (5c-i) already refuses an
  `admin` field on the web channel and the API enforces it (`internal/auth/admin.go:27`,
  `outcomeAdminNotGrantable`); the page ships no admin control. This is the 5c-i outline's
  reading (`2026-09-23-rung-5c-pages-member-verb.md:2501-2503`), recorded here so the spec
  line takes its erratum.

---

## Primary inputs

Read these before Task 1, in this order.

1. **VI in full.** Every `web/` task reads it before writing, and the pass-end reviewers grade
   against it.
2. **The 5c-ii HISTORY entry's handoff**, and the three 5c-ii reviews. Their "checked and
   sound" lists are this pass's do-not-re-litigate record.
3. **The 5c-ii plan's R-5cii2, R-5cii7, R-5cii8, R-5cii19, R-5cii21, R-5cii22.**
4. **DL** sections 5 and 6 (R5.1 through R5.6; G1 through G10) and R3.4, R3.5, R4.1, R4.6.
5. **W** sections 3 (W3.3, W3.6), 4 (W4.1, W4.10), 5 (W5.9, W5.10), 8, 9 (W9.1), 12 (W12.1,
   W12.3's AI-tells row, W12.4, W12.5, W12.6), and 15.
6. **The rung 5 spec's "The pages" and "Consumption".**
7. **This plan's quoted behaviors.** Every relied-on library, browser, stdlib, or license
   behavior below is quoted from its source, a `go doc`, or a scratch probe run on 2026-09-25
   against `go1.27.1 linux/amd64`, Node v24.20.0, `@sveltejs/kit@2.70.3`, `daisyui@5.7.46`,
   and `playwright@1.63.0` with Chromium 153.0.8010.12. The probes lived under
   `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/bc83700f-945c-4755-8eac-4c458ffb90b6/scratchpad/`
   (`pw/`, `webp/`, `fonts/`) and may be gone. Each task re-proves what it relies on through
   its own gate-resident test, and a task that relies on a behavior this plan does not quote
   confirms it the same way and reports what it read.

---

## Goal

After 5c-iii, the served bundle paints in dubplate's own type and component layer, from
embedded faces a gate proves are present, pinned, licensed, and served; every served document
carries the `theme-color` metas and the site name its stamp calls for; and a rendered leg in
the gate proves, in a real browser against the real Go shell, that tier 2 wins in the computed
value, that the match layer's structural values hold under every composed identity, and that
first paint agrees with the stamp under each emulated preference and each cookie. After 5c-iv,
the login page, the consumption interstitial, and the add-member page are real, their strings
live in one deck that passed the register gate and the AI-tells rule, the interstitial spends
the token only on a click and drops it from the address bar, and every page has been rendered
at W4.1's five viewports in both schemes and read before any attended check.

---

## Owner touchpoints

The six rulings above, at the plan-approval gate. The render-and-read's screenshots go to
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

**5c-iii (Tasks 1 through 6).**

1. Six tasks at 1.15M is **6.90M**.
2. The first browser tooling: a Chromium launch inside every full gate from Task 3 on, first
   screenshots an implementer reads, and a heavy gate lane: add **0.30M**.
3. Two `opus` tasks of six (Tasks 3 and 4), a lower share than 5c-ii's three of eight: add
   **0.00M**.
4. Three `go-architecture-reader` dispatches (`internal/serve`, `cmd/dubplate-server`, `web`)
   against 5c-ii's five: **-0.05M**.
5. **Projection 7.15M.** A **9.4M** ceiling puts the 80% flag at **7.52M**, 0.37M above the
   projection, so the flag marks an overrun rather than firing on plan. One pre-named split
   (Task 3's or Task 5's, about 1.15M) crosses the flag, and the conductor asks at that
   boundary under the 80% rule.

**5c-iv (Tasks 7 through 11).**

1. Five tasks at 1.15M is **5.75M**, plus 5c-iv's own Task 0 re-run, inside the per-task rate.
2. The deck's register gate (one `prose-voice-reviewer` and a small fold): add **0.15M**.
3. The render-and-read's fresh reader over about 120 screenshots: add **0.20M**.
4. **Projection 6.10M.** A **8.0M** ceiling puts the flag at **6.40M**.

**Under Ruling 1's "no":** eleven tasks, 12.65M plus the same adders less the second Task 0
and one review round (about -1.0M), about 12.25M, and a 15.9M ceiling (flag 12.7M).

**Checkpoint interval:** STATUS at every segment boundary (after Task 2, after Task 5, after
5c-iii's reviews, at each close, and at the same seams in 5c-iv), at most three tasks apart.
Every task reports its spend, its mutation outcomes, and its measured gate wall time.

**Deliverable count:** forty-four across Tasks 1 through 11, plus each Task 0's report. Every task
carries four, and none passes four.

---

## Global constraints

- **Lane 1, the main tree, on master.** No task reads, writes, or runs a command inside
  `.claude/worktrees/`.
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`, and no task ssh's to
  the box.
- **Package territory, stated as a wall.** This pass writes: the `web/` tree (`src/`,
  `static/`, `scripts/`, `svelte.config.js`, `package.json`, `package-lock.json`);
  `internal/serve` (`document.go`, `csp.go` only if a directive test needs it, and their
  tests); `cmd/dubplate-server` (`web.go`, `web_test.go`, `serve.go` for the site-name and
  route wiring, `serve_test.go`, and the new build-tagged `rendered_test.go`); `web/embed.go`'s
  test file; `Makefile`; `scripts/check.sh`; `CLAUDE.md`'s gate paragraph; `.gitignore`.
  **It writes nothing under `internal/config`, `internal/auth`, `internal/identity`,
  `internal/engine`, `internal/notify`, `internal/harness`, `internal/atomicfile`, `systemd/`,
  `env/`, `config/`, `corpus/`, `tools/`, or `_salvage/`,** and no `docs/` file outside the
  close drafts. A task that believes it needs one stops and reports.
- **No new Go module dependency.** `go.mod` and `go.sum` stay byte-identical.
- **The web dependency set is R-5cii16's pins plus `playwright@1.63.0` under Ruling 2.** The
  pins as committed: `daisyui` is `5.7.46` (`web/package.json`), not the plan's `5.7.45`; Task
  0 records the drift. No ESLint, Prettier, `@playwright/test`, Fontsource, or font package
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
  and asserts that state before it asserts the property.
- **Skills before code.** `go-conventions` before any Go; `svelte-conventions`,
  `ts-conventions`, and the `daisyui` skill before any component, TypeScript, or markup. A
  stock DaisyUI component wins unless VI rules otherwise. No process citation in a shipped
  comment. The em dash is banned in Go comments.
- **Universality.** No estate hostname, path, email, group name, or personal name anywhere in
  `web/` or a fixture. Fixture names use `example.test`.
- **The gate is `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'`**,
  re-issued on exit 75 until it prints `gate exit:`. **From Task 3 on, the gate launches a
  browser, so it runs in the default heavy lane: never set `CAIRN_GATE_LANE=light` on it.**
  Tasks 1 and 2 may run light (no browser yet). Anything reported as evidence runs with
  `-count=1` and reports wall time. Specific files only. Commit footer `Co-Authored-By: Claude
  Opus 5.5 <noreply@anthropic.com>` for an `opus` task and the dispatch's footer otherwise.

---

## Privileged fields and their channels

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| The composed `data-theme` stamp | Unchanged from 5c-ii (R-5cii22): the operator keys, then a valid `__Host-dubplate_theme` cookie. **New writer of the cookie:** the toggle (Task 9), in the page, with `Path=/; Secure; SameSite=Lax` and a value from `label` or `label-night` only. | The server still never writes the cookie and still ignores a value outside the stampable set. The toggle writes no other name and no other value; a rendered row reads `document.cookie` after each click. |
| The `theme-color` meta values | The built bundle's own `--color-neutral` per composed theme, read once at serve startup, validated against a closed hex grammar before entering an attribute (R-5ciii5); when `DUBPLATE_BRANDING` sets `--color-neutral`, that validated value (A4's flattening). | Any request component. A value outside the grammar refuses startup, as R-5cii14's branding refusal does. |
| The site name in a document | `DUBPLATE_SITE_NAME` (a required wall key, `internal/config/keys.go:23`, `config.go:352`), read once at serve startup, escaped with `html.EscapeString` into one meta element (R-5ciii5). | Any request component. The client reads it as text, and Svelte's text interpolation escapes it again. |
| A session, through the interstitial | The consume POST (`<prefix>/consume-link`, field `token`, `internal/auth/handlers.go:21`), sent only by a click on the interstitial's one control (I2). | A page load, a prefetch, a link scanner's GET: the document's load performs no request that carries the token (Task 8's rendered row). |
| The roster write | Unchanged: the roster API behind `RequireAdmin`. The page adds a caller, never a channel; it sends `address`, `display_name`, and `streamer_username` only (`admin.go:25-27`). | The API's own refusals (5c-ii Task 1). The page never sends `admin`. |
| The rendered harness's admin session | A magic link the harness mints with `auth.MintLink` (`internal/auth/token.go:118`) over a fixture secret, consumed through the real interstitial (R-5ciii4). | The harness adds no bypass, build-tagged or otherwise, to any production path; W11.8's dev fixture stays rung 7's. |

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
the tarballs), three files per face, which breaks R5.5's four-file budget and ships less than
the upstream coverage R5.4 requires. The Google Fonts CSS API is declined for the same
slicing and because R5.1 forbids a hosted service.

**Conversion, and why it must be lossless.** Two sources are TTF. The OFL FAQ (openfontlicense.org,
fetched 2026-09-25), item 2.2: "A change in font format normally is considered modification,
and Reserved Font Names (RFNs) cannot be used. Because of the design of the WOFF and WOFF2
formats, however, it is possible to create a WOFF/WOFF2 version that is not considered
modification". Item 2.2.1: "only if the original font data remains unchanged except for WOFF
compression, and WOFF-specific metadata is either omitted altogether or present and includes,
unaltered, the contents of all equivalent metadata in the original font." Alfa Slab One
carries a Reserved Font Name, so a lossy conversion would force a rename. The conversion runs
once on the workstation, `uvx --from 'fonttools[woff]==4.66.0' fonttools ttLib.woff2 compress
--no-glyf-transform -o <out>.woff2 <in>.ttf` (fontTools 4.66.0 is the newest on PyPI, MIT;
the `--help` text reads "`--no-glyf-transform` Do not transform glyf (and loca) tables").
**The task proves losslessness**: `fonttools ttLib.woff2 decompress` of each output yields
every table byte-identical to the source TTF's, and the report pastes the comparison. The
probe's default-transform conversion measured 34,624 bytes (Alfa) and 131,296 bytes (Work
Sans); `--no-glyf-transform` output will be larger, and the task reports the sizes. No Python
file is committed; the command lives in the manifest.

**Pinned bytes.** `web/static/fonts/FONTS.md` (DL R5.2's manifest) names, per file: family,
font version (the `name` table's ID 5, measured "Version 2.000" for Alfa Slab One and
"Version 2.012" for Work Sans), upstream URL at its pinned commit or version, the license URL,
the conversion command, and the file's SHA-256. G4 recomputes each SHA-256 and fails on a
mismatch, so a swapped file cannot pass. License texts sit verbatim in
`web/static/fonts/licenses/`.

**Where they are served.** DL R5.2's placement stands: `web/static/fonts/` copies to
`web/build/fonts/` (measured in the probe build), and a new `fonts` route serves each `.woff2`
under `<prefix>/fonts/`. The `@font-face` `src` is written relative, and Vite rewrites a public
path relative to the built stylesheet: the probe's `url("/fonts/b.woff2")` emerged as
`url(../../../fonts/b.woff2)` in `_app/immutable/assets/0.*.css`, so the reference resolves
under any prefix (S12's font half). The alternative, fonts under `web/src/lib/fonts/` hashed
into `_app/immutable/` (the probe emitted `url(./a.BXklGzmz.woff2)`), is declined because it
moves the license texts and the manifest out of the embedded tree and needs a DL R5.2 and G4
erratum.

**Caching.** R-5cii8 already classes the fonts `no-cache`. `serve.NewAssets` serves through
`http.ServeContent` with a zero `modtime` (`internal/serve/document.go:229`) and no `ETag`, so a
`no-cache` revalidation has nothing to compare and refetches every byte. `go doc
net/http.ServeContent`: "If the caller has set w's ETag header formatted per RFC 7232, section
2.3, ServeContent uses it to handle requests using If-Match, If-None-Match, or If-Range." The
`fonts` route sets a strong ETag from each file's SHA-256, computed once at startup, so a
revalidation answers 304.

**Content type.** Go's built-in extension table has no `.woff2` entry (`mime/type.go`'s list
ends its `.w` entries at `.wav`, `.weba`, `.webm`, `.webp`), and the fallback reads host files
(`mime/type_unix.go`: `/etc/mime.types` and others), which differ between the workstation and
the box. The sniffer does know it (`net/http/internal/sniff.go:178`: `&exactSig{[]byte("wOF2"),
"font/woff2"}`). The route sets `Content-Type: font/woff2` explicitly so no host table decides.

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
  (the last two may be written `var(--color-info)` and `var(--color-warning)`, since they equal
  the fill; W2.7's "The ink defaults to the fill").
- **Surface tokens**, a fourth layer beside the inks and like them outside the DaisyUI blocks
  (so DL R2.1's 21 names hold), in a new `web/src/lib/themes/surface.css` under `@layer base`,
  one rule per composed theme and both `system.css` branches: the shadow color and strengths
  (VI `:204-206`: neutral at 55% and 45% on `label`; black at 65% and 60% on `label-night`),
  the `label-night` keylines (VI `:330`: `#706a5f` for base-surface objects, black for
  color-filled ones), the selection ground (VI `:331`), and the placeholder opacity (VI `:332`,
  62% on `label`). Literal colors live here because G1 exempts `lib/themes/` alone
  (`web/scripts/gates/g1.ts:11`). The implementer names the set; the names are a closed list
  G2 checks per rule.
- **A11y W4 closes as gate rows.** G9 measures every ink and every direct-text color against
  `base-100`, `base-200`, and `base-300` of its own theme (VI: "every ink passes 4.5:1 on
  base-100, base-200, and base-300, so no component needs to know which surface it sits on").
  A second G9 row pins the ten ink values against VI's table, quoted into the script with its
  source line, so an ink that drifts from VI fails even when it still passes contrast. That
  is the row that pins `#9d2f26` and `#754f00`, today unpinned.

### R-5ciii4: the rendered leg is a build-tagged Go test that drives a Playwright script

W12.4: "Driven by Playwright against `dubplate-server serve` already answering". The shape:

- **The server is the real shell, in process.** `cmd/dubplate-server/rendered_test.go`, under
  `//go:build rendered`, builds the shell through `newServeShell` (`serve.go:243`) against a
  fixture environment (temp state dir, a fixture roster holding one admin and one member at
  `example.test`, fixed test secrets, an unreachable SMTP port) on `127.0.0.1:0` with a
  non-default prefix, so every rendered row also proves the prefix. It embeds the bundle the
  `web` target just built, so the stale-binary shape of W12.1 cannot arise; the harness still
  asserts the served `_app/version.json` equals `web/build/_app/version.json` byte for byte.
- **The browser work is one Node script**, `web/scripts/rendered/run.ts`, launched by the test
  under `exec.CommandContext` with the base URL and any minted links in its environment. It
  uses `playwright`'s `chromium.launch()`.
- **An admin session comes from the real login path.** The test mints a link with
  `auth.MintLink` and `auth.LinkURL` (`token.go:118`, `:172`) over the fixture link secret; the
  script opens it and clicks the interstitial's control, so Task 8's I2 rows and every signed-in
  row ride the production consume path. Before Task 8 lands, the stub interstitial has no
  control, so Tasks 3 through 5 run only unauthenticated rows.
- **Server-answer states the fixture cannot produce** (a `sent` link mail, a `rate_limited`
  answer) are rendered by fulfilling the one request with Playwright's `route.fulfill` and the
  JSON shape `internal/auth/handlers.go:157-164` defines; a row that uses a fulfilled answer
  says so.
- **Where it runs.** A `web-rendered` Makefile target runs `go test -tags rendered -count=1 -run
  TestRendered ./cmd/dubplate-server`, plus `go vet -tags rendered` and `golangci-lint run
  --build-tags rendered` over that package, since the default `vet` and `lint` targets never
  compile a tagged file. `make check` runs it after `test`. `scripts/check.sh`'s full form
  always runs it (W12.6: the pass-end gate, and "a per-task gate only for a task that moves
  rendered paint or touches `web/`"); `--scope-changed` runs it when the change set touches
  `web/`, `internal/serve/`, or `cmd/dubplate-server/`, and prints the scope line either way;
  `--reduced` skips it and says so.
- **It gates, never skips (W12.1).** Each of these exits nonzero with a message naming the
  state: Playwright's Chromium absent (`chromium.executablePath()` names no existing file; the
  message names `npx --prefix web playwright install chromium`); the server not answering; a
  page answering outside 2xx; a page that settles without its declared identity (its `h1`
  text, or `main` for the stubs) within the timeout, the un-booted-page shape; the served
  `version.json` differing from the build's; any `securitypolicyviolation` event or console
  error on any page; and a row matrix that ran zero rows.
- **The containment W12.6 names** is `cairn-run-gate`'s own transient user scope (heavy lane,
  7G high, 8G max), so the leg adds no second `systemd-run`.

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
which the `'light'` row covers.

### R-5ciii5: the document head is the Go rewrite's, per variant

The stamp's variants are built once at startup (`serve.BuildDocumentVariants`,
`internal/serve/document.go:57`), so the head additions ride them:

- **`theme-color`.** VI `:128-131`: "`#1f1b16` for light and `#110f0d` for dark, as a `media`
  pair when following the system, and as one value matching the stamped palette when the
  cookie overrides it." The `system` variant carries `<meta name="theme-color"
  media="(prefers-color-scheme: light)" content="…">` and the dark twin; the `label` and
  `label-night` variants carry one meta each. Values come from the bundle's own
  `--color-neutral` per composed theme (the same startup read that finds the theme selectors,
  `cmd/dubplate-server/web.go:473-503`), checked against `^#[0-9a-fA-F]{3}([0-9a-fA-F]{3})?$`
  before they enter an attribute; the task quotes the form the minified bundle emits. With
  `DUBPLATE_BRANDING` setting `--color-neutral`, every variant carries one meta at the branding
  value (A4: branding flattens both palettes). The toggle updates the meta in place when it
  changes the stamp.
- **The site name.** VI ruling 4: "the login heading is the group's name". The name is
  `DUBPLATE_SITE_NAME`, a required wall key. `newServeShell` passes `cfg.SiteName` to the web
  routes, and the rewrite inserts `<meta name="dubplate-site-name" content="…">`, escaped with
  `html.EscapeString`, into every variant. The page reads it synchronously at mount, so no
  request and no flash precede it. A config endpoint (DL R4.3's "config endpoint") is declined
  here: it would be a fetch and a repaint for one string, and no endpoint exists.
- **Every head insertion is outside every script body**, so 5c-ii's per-variant CSP recheck
  (`checkVariantCSP`, `web.go:297`) keeps passing; a mutation proves the recheck still fires.

### R-5ciii6: every URL composes from the runtime base (S12)

Kit's `resolve` (`@sveltejs/kit/types/index.d.ts:3470-3490`, since 2.26): "Resolve a pathname
by prefixing it with the base path, if any". Its argument type is the app's own page paths
(`.svelte-kit/non-ambient.d.ts:42`: `"/" | "/login" | "/login/link" | "/members"`), so the Go
API paths (`request-link`, `consume-link`, `sign-out`, `roster`,
`cmd/dubplate-server/serve.go:42-49`) go through one helper that joins `resolve('/')` with the
endpoint name. Static files go through kit's `asset` (`:3454-3469`, since 2.26). The
favicon is a `<link rel="icon">` in `app.html` using `%sveltekit.assets%`, which the probe
build rendered as `href="./favicon.svg"` in `login.html` and `href="../favicon.svg"` in
`login/link.html`: relative, present before any script, and therefore no root `/favicon.ico`
request (the root half of S12). A new static gate (`web/scripts/gates/root-absolute.ts`)
fails on any root-absolute `href=`, `src=`, `action=`, `fetch('/…')`, `goto('/…')`, or CSS
`url(/…)` literal in `web/src`.

### R-5ciii7: builds are reproducible (S11)

Kit (`types/index.d.ts:884-903`): the version name "must be deterministic (e.g. a commit ref
rather than `Math.random()` or `Date.now().toString()`), otherwise defaults to a timestamp of
the build." The timestamp reaches a chunk (the current build's `version.json` value
`1790346892862` appears in `_app/immutable/chunks/Y8ElIxAE.js`), so every build renames chunks
and moves the inline script's hash. `kit.version.name` becomes a SHA-256 over
`web/package-lock.json` and every file under `web/src/` and `web/static/`, computed in
`svelte.config.js` with `node:crypto`. Kit's own example (`git rev-parse HEAD`) is declined
because a source tarball without `.git` would fail to build, and packaging is a goal (VI
"Licensing"). **Proof:** two consecutive builds of one tree produce byte-identical
`web/build` trees, a one-time evidence row in Task 3's report and a gate row that compares
the served and built `version.json` (R-5ciii4).

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
  and asserts `--radius-box` changes to `light`'s own value, so the row cannot pass on a page
  where no built-in could leak.
- **First paint** (R-5cii22's rendered half): for each of the six states (no cookie under
  light and dark emulation; cookie `label` and cookie `label-night` under each), the document
  element's painted background on the first frame equals the stamped or media-resolved
  `--color-base-100`. "First frame" is sampled from a screenshot taken at `domcontentloaded`
  with JavaScript disabled for that context, so no script can have corrected it.

---

## Review focus

1. **A token in the interstitial's URL** is read once, dropped from the address bar before
   any other work, and sent only on a click. Neither `location.href` nor `history.state`
   carries it afterward. Pinned in Task 8.
2. **A root-absolute URL** in any page, fetch, or stylesheet under a prefix. Pinned in Tasks 3
   and 4.
3. **A fill-only or low-contrast color as a focus ring** (a11y W5): every focusable control on
   every page shows the one `--ink-primary` ring. Pinned in Task 5.
4. **Operator text into markup**: the site name and a branding neutral in the head. Pinned in
   Task 4.
5. **A rendered gate that passes on nothing**: an un-booted page, a missing browser, a zero-row
   matrix. Pinned in Task 3.
6. **The members page rendered to a non-admin** by client navigation: it shows the not-found
   copy and no roster. Pinned in Task 9.

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched before any implementer, and again before Task 7 if
Ruling 1 is yes. Reads this plan against the tree as shipped. **Its output is blockers with
evidence, not fixes.** It edits no tracked file. A moved pointer whose content survives is a
note; a pointer whose content is gone is a blocker.

1. **Lane 1 free, the tree clean** of warm changes the conductor did not author. Report HEAD.
2. **Every `file:line` this plan cites** resolves to the content described, including
   `web/src/app.css:3`, `web/src/lib/themes/inks.css:14-24` (two inks today),
   `web/scripts/gates/g1.ts:11-12`, `web/scripts/gates/theme-bundle.ts:352`,
   `internal/serve/document.go:21,57,211-232`, `internal/serve/csp.go:32-36`,
   `cmd/dubplate-server/web.go:57-80,156-162,297,473-503`,
   `cmd/dubplate-server/serve.go:42-62,243,289-311`, `internal/auth/handlers.go:20-21,52-62,157-164`,
   `internal/auth/admin.go:25-46`, `internal/auth/token.go:55,64,118,172`, `Makefile:42-87`,
   `scripts/check.sh:1-60`.
3. **Every pin**: `npm view <pkg> version` for each R-5cii16 package, `playwright`, and
   `@ibm/plex-mono`; PyPI's newest `fonttools`; the `google/fonts` commits R-5ciii1 names still
   carry the named files. Report every newer version; a newer minor or patch is taken through
   the `dependency-upgrade` skill by the conductor before Task 1, a major is a question.
4. **Playwright's Chromium** for 1.63.0 (revision 1243) is present under
   `~/.cache/ms-playwright/`; if absent, report it (the install is the conductor's action).
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
- **The faces and their record (R-5ciii1):** three files obtained, converted losslessly where
  needed, license texts verbatim, and `FONTS.md` with every field R-5ciii1 names.
- **The `@font-face` rules (DL R5.3):** each with `font-display: swap`, a relative `src` with
  `format('woff2')` and no second format, explicit `font-weight` (`100 900` for Work Sans) and
  `font-style`.
- **G4 (DL section 6), extended:** every font file under `web/static/fonts/` ends in `.woff2`
  and no other font format is present, skipping `licenses/` and `FONTS.md`; every `@font-face`
  sets the three descriptors; the file count is within four (the R5.5 deviation); each file's
  SHA-256 equals its manifest row, and every manifest row names an existing file.
- **The `fonts` route and G5:** `<prefix>/fonts/<name>.woff2` answers GET and HEAD for each
  embedded face with `Content-Type: font/woff2`, `Cache-Control: no-cache`, and a strong ETag,
  answering 304 to a matching `If-None-Match`; every other path under it (a directory, a
  licence text, `FONTS.md`, a missing name, a trailing slash) answers the terminator's 404
  (R-5cii21). G5 (a Go test over the built bundle) resolves every `@font-face` `src` in the
  built stylesheet against the embedded tree, asserting a nonzero face count first.

**Acceptance criteria:**
- The report pastes the per-table byte comparison of each converted file against its TTF and
  each file's size and SHA-256.
- G4 fails by name on: a `.ttf` added beside the faces; a `@font-face` missing
  `font-display`; a fifth file; a manifest SHA-256 edited by one character. **Fixture state:**
  each seeded in a scratch copy, the real tree passing first.
- G5 fails naming the missing build on the placeholder-only tree, and naming the path when one
  face file is removed from the embedded tree. **Fixture state:** the real built bundle
  carrying three `@font-face` rules, asserted.
- The route rows: each state above, plus a HEAD, a 405 with `Allow: GET, HEAD`, and a 304.
- `bash scripts/check.sh` green.

**Named mutations:**
1. Convert with the default glyf transform → the byte comparison reports changed `glyf` and
   `loca` tables, and the task does not ship that file.
2. Drop the explicit `Content-Type` → the route row asserting `font/woff2` fails on a host
   whose mime tables lack `.woff2` (the test serves through a handler with no extension table
   available, or asserts the header was set before `ServeContent`; the task states which).
3. Serve `FONTS.md` through the route → the 404 row for it fails.
4. Point one `src` at an absolute `/fonts/...` path → the build still rewrites it relative
   (R-5ciii1's probe); the task reports whether the root-absolute gate (Task 4) or G5 catches
   it, and Task 4's gate must.

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
- G9's per-surface rows report every pair measured, with the worst case for each ink matching
  VI's table column (4.54 / 4.85 and so on) to two decimals. **Fixture state:** three bases per
  theme read from the theme files, asserted present.
- `dubplate-system`'s parity holds for every new name in both branches.
- `bash scripts/check.sh` green.

**Named mutations:**
1. Set `label`'s `--ink-primary` to `#c1352b` (the fill) → the `base-300` row fails (VI: the
   fill is why the ink exists).
2. Set `label`'s `--ink-warning` to `#765000` (the draft's value, still above 4.5:1 on
   `base-100`) → the pinned-value row fails.
3. Drop `--ink-info` from `system.css`'s dark branch → the parity row names it.
4. Move a surface token into a `daisyui/theme` block → G2's 21-name count fails.

**2.0 hook: neutral.** **Deliverable count: 4** (the stacks, the inks, the surface tokens, the
gate rows).

---

## Segment B (5c-iii)

## Task 3: The rendered leg: the harness, G7's computed half, G10, and first paint

Model: `opus`. Depends on Task 2. Upshift reason: W12.1's refusal set is exactly the class of
check that passes vacuously when built quickly, and the harness composes the real shell.

**Files:**
- Create: `cmd/dubplate-server/rendered_test.go` (`//go:build rendered`),
  `web/scripts/rendered/run.ts` (and siblings it needs)
- Modify: `web/package.json`, `web/package-lock.json` (Ruling 2), `web/svelte.config.js`
  (R-5ciii7), `Makefile` (`web-rendered`; `check` runs it), `scripts/check.sh` (the forms, the
  header comment), `CLAUDE.md`'s gate paragraph ("The rendered (browser) gates join this leg at
  rung 5c-iii" becomes present tense), `.gitignore` (the harness's screenshot output directory)

**Outcomes:**
- **The harness (R-5ciii4)** with every W12.1 refusal, on a reproducible build (R-5ciii7).
- **G7's computed half (R-5ciii8).**
- **G10 (R-5ciii8).**
- **The first-paint rows (R-5ciii8)**, the six states, plus a no-CSP-violation row over every
  page that also answers the 5c-ii carry "confirm rendered in 5c-iii" for DaisyUI's `data:`
  noise texture (svelte W4, a11y S7).

**Acceptance criteria:**
- Each W12.1 refusal fails by name: the harness pointed at a missing browser path; a document
  whose identity never appears (a stub with its `h1` removed in a scratch build); a served
  `version.json` edited in the embedded copy; a page with a deliberately unhashed inline
  script (the CSP violation row); an empty row matrix. **Fixture state:** the real tree passes
  first, then each seed fails.
- G7's with-branding rows and without-branding rows both run, and the operator row runs.
- G10 reads eleven names per identity and sees `--radius-box` change under `light`.
- First paint: six states, each asserting the stamped attribute in the served bytes before it
  samples the pixel.
- The report pastes two consecutive builds' `web/build` tree hashes, equal, and the gate's
  wall time before and after the leg.
- `bash scripts/check.sh` green, heavy lane.

**Named mutations:**
1. Leave `kit.version.name` unset → the two-build comparison differs.
2. Make the harness skip when Chromium is absent → the missing-browser row passes silently,
   and the diff-reviewer rejects the skip (W12.1).
3. Put `branding.css` before the bundle stylesheet in `tierHrefs` → G7's computed row fails
   while the 5c-ii link-order row also fails (both halves see it).
4. Remove `--radius-box` from the `dubplate-label` match block and from G2's expectation →
   G10 reads `light`'s value through `:root` and fails.
5. Declare the `system.css` dark branch before the light one → the dark-emulation first-paint
   row fails.

**2.0 hook: neutral.** **Prior art:** Playwright adopted (Ruling 2); `@playwright/test`
declined (R-5ciii4); cairn's rendered runner declined for the structural reason W12.8 records.

**Pre-named split:** if the four cannot land in one green gate, G7's computed half and G10
move to their own task after this one; the harness, the refusals, and first paint stay.

**Deliverable count: 4** (the harness with its refusals and the reproducible build, G7, G10,
first paint).

---

## Task 4: The document head: `theme-color`, the site name, the mark and favicon, and the root-absolute gate

Model: `opus`. Depends on Task 3 (its rendered rows use the harness). **Independent of Task 5
by files.** Upshift reason: two operator-supplied values enter markup (R-5ciii5), the class
Task 6 of 5c-ii upshifted for.

**Files:**
- Create: `web/static/favicon.svg`, `web/static/favicon.png` (32px),
  `web/static/mark/mark-{large,medium,small}.svg` (copied from
  `docs/superpowers/specs/2026-09-24-visual-identity/`, unmodified), a `Mark` component under
  `web/src/lib/components/`, `web/scripts/gates/root-absolute.ts`
- Modify: `web/src/app.html` (the icon links), `internal/serve/document.go` and its test (the
  head insertions ride `BuildDocumentVariants`), `cmd/dubplate-server/web.go`, `web_test.go`,
  `serve.go` (passing `cfg.SiteName`), `serve_test.go`, `Makefile` (the new gate),
  `web/scripts/rendered/run.ts` (the head rows)

**Outcomes:**
- **`theme-color` (R-5ciii5)** in every variant, per the stamp and branding states.
- **The site name (R-5ciii5)** as one escaped meta in every variant.
- **The mark and favicon (VI "The mark"):** the three SVGs served as static files, the favicon
  `mark-small.svg` with a 32px PNG rendered once by `rsvg-convert` (the command in the task
  report), both linked from `app.html` through `%sveltekit.assets%`, and Go routes answering
  their exact paths `no-cache` with a strong ETag; a `Mark` component that picks the SVG by
  rendered size (56px and up large, 21 to 55 medium, 20 and below small) through kit's
  `asset()`, with empty `alt`, the wordmark text beside it carrying the name. G1 does not scan
  `web/static/` (`g1.ts:10`), where the mark's fixed brand colors live.
- **The root-absolute gate (R-5ciii6)** in the `web` target.

**States each check meets:**

| Check | State | Report |
| --- | --- | --- |
| Neutral read from the bundle | present, hex | accepted |
| Neutral read from the bundle | absent for a stampable palette | startup refuses, naming the theme |
| Neutral read from the bundle | not matching the hex grammar (an `oklch()` a later minifier emits) | startup refuses, naming the value's theme, not the value |
| Branding sets `--color-neutral` | a grammar-valid hex | every variant carries one meta at it |
| Site name | plain text | escaped into `content` |
| Site name | carrying `"`, `<`, `&` | escaped; the rendered heading shows the literal characters |

**Acceptance criteria:**
- Go rows for every state above, over the real built bundle, and the CSP recheck still passes
  for every variant.
- Rendered rows: the `system` variant's metas switch with emulation (`page.evaluate` reads the
  matching meta); the favicon request goes to `<prefix>/favicon.svg`, never `/favicon.ico`
  (request log); the site name renders as text.
- The root-absolute gate fails by name on a seeded `href="/members"`, `fetch('/roster')`, and
  `url(/fonts/x.woff2)` in a scratch copy, and passes the real tree.
- `bash scripts/check.sh` green.

**Named mutations:**
1. Insert the site name unescaped → the `"` row breaks the attribute and the rendered heading
   row fails.
2. Place the `theme-color` insertion inside the first `<script>` → the CSP recheck refuses
   startup.
3. Emit a media pair on the `label` variant → the cookie-override row finds two metas.
4. Remove the `app.html` icon link → the request log shows `/favicon.ico`.

**2.0 hook: neutral.** **Deliverable count: 4** (`theme-color`, the site name, the mark and
favicon, the root-absolute gate).

---

## Task 5: VI's component layer: the focus rule, the ink utilities, the printed surface, and the refinement rules

Model: `sonnet`. Depends on Task 3 (its proofs are rendered rows). **Independent of Task 4 by
files.**

**Files:**
- Create: `web/src/lib/styles/components.css` (or a name the task states), imported once from
  `app.css`
- Modify: `web/src/app.css`, `web/scripts/rendered/run.ts` (the component rows, run against a
  fixture page under `web/src/routes/` only if the task shows no stub can host a control; the
  default is to extend the stubs with no copy change)

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
- **The refinement rules this pass's pages use (VI `:320-334`):** `.display` at line-height
  1.1 with `text-wrap: balance`; `text-wrap: pretty` on paragraphs and alert text; the three
  mono sizes and two trackings as utilities; the label rule under page titles (a 3px line over
  a 1px line); `::selection`; the `label` placeholder opacity; hover states and pointer
  cursors on header links; motion stops under `prefers-reduced-motion: reduce`. Rules for
  surfaces no page renders (tabs, tracklists, the film edge, the catalog line) wait for rung 7,
  with the rest of Ruling 5's list.

**Acceptance criteria (rendered rows, both schemes):**
- Tabbing to each focusable element on each stub paints a ring whose read-back color equals
  `--ink-primary`'s under that theme, including a `btn-secondary`, a `btn-warning`, a
  `btn-neutral`, a `btn-primary`, and an `input`. **Fixture state:** each control present on
  the page, asserted by role before tabbing.
- A pressed button's box moves 2px down and right; under `reducedMotion: 'reduce'` it does not.
- Shadows, borders, and the keylines read back per theme.
- No icon-only control carries a primary fill (Ruling 4).
- `bash scripts/check.sh` green.

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
refinement rules).

---

## Task 6: The 5c-iii close (under Ruling 1's yes)

Model: `sonnet`. Depends on Tasks 0 through 5, the simplifier round, and the pass-end reviews
with their fold.

**Files (drafts only; the conductor folds and commits):**
`docs/superpowers/research/handoffs/draft-HISTORY-5c-iii.md`,
`docs/superpowers/research/handoffs/draft-STATUS-5c-iii.md`

**Outcomes:** the HISTORY draft (per task with commit ids; the gate evidence including the
rendered leg's wall time and the build's peak memory under the heavy lane; the resolutions and
rulings; the review dispositions; the declared gaps; both budgets; "what a later pass would be
wrong to rediscover"); the errata owed, paste-ready (the third split in the rung 5 spec's 5c
row and W section 15; DL R5.5; W4.1's both-schemes line; the rung 5 spec's admin checkbox; VI
`:141`'s clause number; VI `:149`'s Permanent Marker license; R-5cii8's members row); the
5c-iv handoff (what Tasks 7 through 10 build on); the STATUS draft under 60 lines.

**Acceptance criteria:** every count produced by a command the report pastes; every commit id
resolves; `bash scripts/check.sh` green as a formality.

**Deliverable count: 4.**

---

## Segment C (5c-iv)

## Task 7: The string deck, the AI-tells rule, and the register gate's input

Model: `sonnet`.

**Files:**
- Create: `web/src/lib/copy/deck.ts`, `web/scripts/gates/ai-tells.ts`
- Modify: `Makefile` (the rule in the `web` target)

**Outcomes:**
- **The deck** (rung 5 spec `:1452-1460`): every user-facing string of the three pages, the
  error page, and the signed-in header, as named constants: the login subline ("Sign in with
  your email. We'll send you a link.", VI `:304-305`, with curly apostrophes per VI `:326`);
  one line per request-link outcome (`sent`, `requested`, `unknown` with its admin contact,
  `rate_limited`, `invalid_address`, `too_large`, `malformed`, `failed`,
  `handlers.go:52-62`) and the cross-origin 403; the interstitial's statement, its control,
  and one line per consume outcome (`signed_in`, `refused`, `rate_limited`, `too_large`,
  `malformed`, `failed`); the sign-out control and outcome, which never claims revocation (the
  5b-ii bound, `2026-09-23-rung-5c-pages-member-verb.md:2495-2497`); the add-member form's
  labels and hints and one specific line per roster outcome (`admin.go:35-46`); the roster
  list's empty state (W7.1, W7.2) and its plural-aware count (W8.4); the error page's two
  sentences; the toggle's accessible name, "Dark theme" (VI `:132`). The signed-in state is
  the honest terminal state the spec names for no streamer (`spec:1270-1273`): a plain
  confirmation, no destination.
- **The AI-tells rule (W12.3; W section 15: "The AI-tells rule also runs inside 5c"):** a port
  of cairn's `scripts/checks/check-admin-prose.mjs` (MIT, 285 lines; the blocking lists
  "ported verbatim from prose-guard") as a gate script scanning every string in `web/src`
  markup and literals, with a `--list` mode. It enforces no register.
- **The register gate's input:** the deck's strings in one reviewable file, handed to the
  conductor's `prose-voice-reviewer` dispatch against the Microsoft editor standard (the
  5c-i mail-copy precedent), whose findings a small fold task applies.
- **Copy never names a mechanism the member cannot see**, never puns (VI ruling 4), and a
  reject, security, or rate-limit line is never playful (W1.4).

**Acceptance criteria:** every outcome value the handlers answer has a deck line, asserted by
a test that lists the Go outcome constants' values (read from the source files) against the
deck's keys; the AI-tells rule fails by name on a seeded "Moreover" and an em dash in a
scratch copy; `bash scripts/check.sh` green.

**Named mutations:** 1. Delete the `rate_limited` line → the coverage row names it. 2. Seed
"delve" in a component string → the AI-tells rule fails. 3. Write "you have been signed out
everywhere" → the register reviewer is asked to catch it; the deck test checks no sign-out
line contains "everywhere" or "revoked".

**2.0 hook: neutral.** **Deliverable count: 4** (the deck, the coverage test, the AI-tells
rule, the register input).

---

## Task 8: The login page, the consumption interstitial, and the error page

Model: `sonnet`. Depends on Task 7. **Independent of Task 9 by files.**

**Files:**
- Modify: `web/src/routes/login/+page.svelte`, `web/src/routes/login/link/+page.svelte`,
  `web/src/routes/+error.svelte`, `web/scripts/rendered/run.ts`
- Create: a small API helper under `web/src/lib/` (R-5ciii6) if Task 7 did not

**Outcomes:**
- **The login page** (rung 5 spec `:1415-1433`; VI `:258-261` under Ruling 3): the `Mark` and
  the wordmark in Alfa Slab One, the site name as the `h1` over the label rule, the subline, a
  DaisyUI `fieldset` with `legend`, `label`, and `input type="email"` with
  `autocomplete="email"`, `validator` and a space-reserving `validator-hint`, and one primary
  submit. A `sent` or `requested` outcome reaches assistive technology through a polite live
  region without moving focus; an error through an assertive one. The server's copy stays
  authoritative over `validator`. No theme toggle (VI `:133`).
- **The interstitial** (rung 5 spec "Consumption"; security I2): on mount it reads `t` from
  `location`, then **before any other work** replaces the history entry with the same URL less
  its query. Kit's `$app/navigation` `replaceState` stores `page.url.href` (the token URL) in
  the history entry's state (`client.js:2573`, `[PAGE_URL_KEY]: page.url.href`), so it is not
  used for this; the page calls `history.replaceState(history.state, '', <URL without query>)`,
  keeping kit's index keys, and the task quotes kit's dev-only warning (`client.js:90-114`,
  inside `if (DEV && BROWSER)`) to show the production build is unaffected. It states what is
  about to happen and carries one control; only a click POSTs `token` to `consume-link`. It
  loads nothing from a third party (CSP `default-src 'self'` already forbids it). A missing
  token renders the `refused` copy with no request.
- **The error page:** VI's plain sentence in the identity's type and surface, no photo
  (Ruling 3).
- **Every page carries W3.6's skip link and `<main id="main" tabindex="-1">`.**

**Acceptance criteria (rendered, both schemes, under the prefix):**
- I2: after the interstitial loads with a minted token, `location.search` is empty and
  `JSON.stringify(history.state)` does not contain the token; the request log shows no request
  to `consume-link` before the click and exactly one after; every request's origin is the
  server's. **Fixture state:** a real minted link for the fixture admin, asserted valid by the
  row that clicks it and receives `signed_in`.
- A reload of the stripped URL shows the `refused` copy and sends nothing.
- Each login outcome renders its deck line, with the live region's politeness per class
  (`sent` fulfilled, `failed` real against the unreachable SMTP port, `unknown` real for an
  address not in the fixture roster under the kind answer).
- One `h1` per page (W12.4's `h1-count`), the skip link first in tab order.
- `bash scripts/check.sh` green.

**Named mutations:**
1. Use `$app/navigation`'s `replaceState` → the `history.state` row finds the token.
2. POST on mount → the request log finds a consume before the click.
3. Strip the query after the fetch instead of before → the ordering row fails.
4. Drop `autocomplete="email"` → a static row over the login markup fails.

**2.0 hook: neutral** (the device page will live in this app, VI and S20 `:390-396`; nothing
here builds for it). **Deliverable count: 4** (login, interstitial, error page, the rendered
rows).

---

## Task 9: The add-member page, the signed-in header, and the theme toggle

Model: `sonnet`. Depends on Task 7. **Independent of Task 8 by files.**

**Files:**
- Create: a `SiteHeader` component (mark, wordmark, nav, toggle, sign-out) under
  `web/src/lib/components/`
- Modify: `web/src/routes/members/+page.svelte`, `web/scripts/rendered/run.ts`

**Outcomes:**
- **The signed-in header (VI `:131-136`):** a neutral bar with the `Mark` and wordmark, and at
  the right end of the nav the toggle, an icon-only DaisyUI `swap` driven by its hidden
  checkbox (the stock mechanism, `daisyui` skill "swap" rule "Use only a hidden checkbox to
  control the swap state"), sun and moon Phosphor bold icons (VI `:237`), accessible name "Dark
  theme", checked when the stamp resolves dark. Choosing a state writes
  `__Host-dubplate_theme=label|label-night; Path=/; Secure; SameSite=Lax` (R-5cii22), sets
  `data-theme` on the document element (DL R4.2: "The client changes the attribute after
  hydration"), and updates the `theme-color` meta. Never a `.theme-controller` (DL R4.6; the
  theme-bundle gate already refuses the class, `theme-bundle.ts:352`). The toggle and each nav
  link meet 44px (VI `:333`). A sign-out control POSTs `sign-out` and returns to the login
  page.
- **The roster list:** the members the API returns, each with display name, address, and
  streamer username, and a squared-outline badge marking an admin; the plural-aware count; the
  empty state. It stacks at narrow widths (W4.4, W4.6).
- **The add-member form:** `fieldset` of address, display name, and an optional streamer
  username; no admin control (the spec deviation above); one primary submit; each roster
  outcome renders its specific deck line, and `added` refreshes the list.
- **A non-admin reaching the page by client navigation** (R-5cii19) sees the error page's
  not-found sentence and no roster, since the API's GET answers the terminator's 404.

**Acceptance criteria (rendered, both schemes):**
- After a toggle click, `document.cookie` holds exactly the chosen value, the attribute
  changed, and a reload's served bytes carry the matching stamp. **Fixture state:** the
  signed-in admin from a real consumed link, and a starting stamp asserted before the click.
- An add with a fresh address shows `added` and the list grows by one; a repeat shows
  `duplicate_address`; the fixture roster file on disk shows the write (real API).
- The non-admin row: a member session, `goto` to the members route, the not-found sentence
  visible, and no element containing the admin's fixture address.
- `bash scripts/check.sh` green.

**Named mutations:**
1. Use `.theme-controller` on the checkbox → the theme-bundle gate fails.
2. Write the cookie without `Secure` → the browser refuses a `__Host-` cookie, and the reload
   row finds the old stamp.
3. Render the roster from a cached copy after a 404 → the non-admin row finds an address.
4. Add an admin checkbox → the API answers `admin_not_grantable` and the add row fails.

**2.0 hook: neutral.** Hook 5 (the stable identity) is untouched: the toggle's cookie carries
no identity (W section 15), and `streamer_username` stays optional per R-5c13 until rung 12.
**Deliverable count: 4** (the header with the toggle, the list, the form, the rendered rows).

---

## Task 10: The render-and-read

Model: `sonnet` for the capture; the read is the conductor's dispatch.

**Files:**
- Modify: `web/scripts/rendered/run.ts` (the capture matrix and the overflow floor)
- Create: `docs/superpowers/research/2026-09-25-rung-5c-iv-render-and-read.md` (the reader's
  findings, committed by the conductor)

**Outcomes:**
- **The capture matrix:** every page and every state Tasks 8 and 9 name, at 320, 390, 768,
  1440, and 2560 CSS pixels (W4.1), in `label` and `label-night`, full page, to the ignored
  output directory, with a contact sheet per page. The row count is asserted against the
  matrix so a missing capture fails.
- **The viewport-overflow floor (W12.4):** at 390 and 320 nothing paints past the viewport,
  gated on the document scrolling.
- **The touch-target floor (W4.10)** on every control at every width.
- **The read:** the conductor dispatches one fresh-context reader (`opus`, no plan in
  context) with VI's path, W section 4, and the contact sheets, asking for defects ranked by
  consequence; a small fold task applies what the conductor accepts, and the matrix re-runs.

**Acceptance criteria:** the matrix count equals pages × states × 5 × 2; the two floors pass
and each fails by name on a seeded 400px-wide element and a 30px button; the reader's verdict
is accept, or its accepted findings are folded and re-captured; `bash scripts/check.sh` green.

**2.0 hook: neutral.** **Deliverable count: 4** (the matrix, the overflow floor, the touch
floor, the read and its fold).

---

## Task 11: The 5c-iv close

Model: `sonnet`. As Task 6, over Tasks 7 through 10: the HISTORY and STATUS drafts, the
errata (the rung 5 spec's 5c row closed; W section 15's rendered and AI-tells lines as
landed), and the handoff to 5d (the `deploy.sh` clean-build row now includes the rendered
leg's absence on the box, since nothing Node or Playwright ships) and to rung 7 (Ruling 5's
four items, the photo column under Ruling 3, W12.4's remaining rules). **Deliverable count:
4.**

---

## The exit gate

`bash scripts/check.sh` green on master in the full form, plus each assertion below and its
proving task.

1. Three faces embedded, pinned by SHA-256, licensed, served `font/woff2`; G4 and G5 fail by
   name on their seeds (Task 1).
2. Every ink passes 4.5:1 on all three bases of its theme and equals VI's table (Task 2).
3. The rendered leg runs in every full gate and refuses each W12.1 shape (Task 3).
4. Tier 2 wins in the computed value under every stamp; tier 3 wins over tier 2; the match
   layer's structural values hold under every composed identity (Task 3).
5. First paint agrees with the stamp in all six states (Task 3).
6. Two builds of one tree are byte-identical (Task 3).
7. `theme-color` follows the stamp and branding; the site name is escaped; no request goes to
   `/favicon.ico`; no root-absolute URL ships (Task 4).
8. Every focus ring is `--ink-primary` at 2px offset 2px, on every control (Task 5).
9. The interstitial drops the token before any request and spends it only on a click (Task 8).
10. The toggle writes the cookie, and the next render stamps it; a non-admin sees no roster
    (Task 9).
11. Every deck line exists for every outcome; the AI-tells rule and the register gate pass
    (Task 7).
12. Every page rendered at five widths in both schemes, floors green, read and folded (Task
    10).

**Deferred, by name:** the photo column and error-page photos (Ruling 3); the sleeve, grain,
rim, no-art sleeve, and Permanent Marker (Ruling 5); VI's "Open for rung 7" list; W12.4's
remaining rendered rules; W11.8's dev fixture.

---

## Contested surfaces

| Surface | Owner | This pass's position | Contention |
| --- | --- | --- | --- |
| `web/` | rung 5 births it; rung 7 authors visuals | every task | none open; rung 7 inherits |
| `cmd/dubplate-server/web.go`, `serve.go`, their tests | rung 5 | Tasks 1, 3, 4 | none; no other lane edits them |
| `cmd/dubplate-server/rendered_test.go` (new) | rung 5 | Task 3 onward | none |
| `internal/serve/document.go` | rung 5 | Task 4 | none |
| `Makefile`, `scripts/check.sh`, `CLAUDE.md` gate paragraph | shared | Tasks 1, 3, 4, 7 | **after Task 3 lands, every full gate on every lane launches Chromium and needs `npm ci --prefix web` plus Playwright's browser**; the conductor tells every lane through the registry |
| `go.mod`, `go.sum` | shared | untouched | none |
| `docs/` ledgers | conductor | Tasks 6 and 11 draft only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record planning spend against the rung; write the lane 1 registry row and
   the contested-surface rows; arm the runaway guard; AC power; confirm the gate string runs
   heavy from Task 3.
2. Before Task 1: rule Task 0's blockers; take any minor or patch pin bump through
   `dependency-upgrade`.
3. Before Task 3: `npm ci --prefix web` after the lockfile changes; if Task 0 found the
   browser absent, `npx --prefix web playwright install chromium`, recorded in the
   workstation's install manifest per the global rule.
4. After Task 5: one `code-simplifier` round over Tasks 1 through 5, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, reviewed one `diff-reviewer` per
   package; then `svelte-reviewer` and `daisyui-a11y-reviewer`, each given VI's path and asked
   for conformance to it (tokens, inks, the component layer, the copy register) graded beside
   correctness, and `web-auth-security-reviewer` over the head insertions and the new routes;
   each asked for correctness gaps ranked by consequence; one fold dispatch that may refuse a
   finding whose fix costs more than the risk it removes, with a one-line reason.
5. At Task 6: one `go-architecture-reader` each for `internal/serve`, `cmd/dubplate-server`,
   and `web`, no plan in context. Re-derive 5c-iv's figure from 5c-iii's actual.
6. After Task 7: the register gate (one `prose-voice-reviewer`, Microsoft editor standard) and
   its fold.
7. At Task 10: the render-and-read reader and its fold.
8. After Task 10: the simplifier round over Tasks 7 through 10; the three reviewers again,
   `web-auth-security-reviewer` pointed at the interstitial's token handling (I2) and the
   toggle's cookie write; the fold.
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
| **A** | 1, 2 | STATUS. The last light-lane gates. |
| **B** | 3, 4, 5 | STATUS. The simplifier, the three reviewers, the fold. |
| **Close** | 6 | The 5c-iii close; the architecture reads; the ledger fold. |
| **C** | 7, 8, 9 | 5c-iv's Task 0 first. STATUS. The register gate after Task 7. |
| **D** | 10, 11 | The read and its fold; the simplifier; the reviewers; the close. |

Each boundary sits on a task's own green committed gate.

**Models.** `sonnet` at effort `high` by default. **Two upshifts to `opus`, unconditional**:
Task 3 (the harness's refusal set) and Task 4 (operator text into markup). Task 0 runs at
`opus` in fresh context.

**Dependencies, stated exactly:**
- Task 0 first; its blockers ruled before Task 1.
- Task 1 → Task 2 → Task 3.
- Task 3 → Task 4 and Task 3 → Task 5. **Tasks 4 and 5 are independent by files**; they run
  sequentially because one tree runs one executor.
- Tasks 1 through 5, the simplifier, and the reviews → Task 6.
- Task 6 → 5c-iv's Task 0 → Task 7.
- Task 7 → Task 8 and Task 7 → Task 9. **Tasks 8 and 9 are independent by files.**
- Tasks 8 and 9 → Task 10 → the reviews → Task 11.

**Where this pass splits further, if it must:** at the Task 4 / Task 5 boundary, if segment
B's re-derivation projects past the flag. Task 5 then opens 5c-iv, since Tasks 8 and 9 are its
first consumers.

**Budget:** 9.4M (5c-iii) and 8.0M (5c-iv) under Ruling 1, flags 7.52M and 6.40M; planning
recorded separately; STATUS at each boundary; spend scored at each close against its ceiling
and against zero attended interactions after approval.
