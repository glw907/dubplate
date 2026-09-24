# Rung 5, Pass 5c-ii (The Admin Gate, the Roster API, and the `web/` Tree Born Gated) Implementation Plan

Status: **revision 2.2, 2026-09-24, prose folded, awaiting owner approval.**
Revision 1 (`3ebfe3c`) took three lens reviews (`140e3cd`:
`docs/superpowers/research/2026-09-24-rung-5c-ii-plan-review-contract.md`,
`-mechanics.md`, `-risk.md`); revision 2 (`74066b1`) folded them and
took a fresh-context verification read (`cebcda2`:
`docs/superpowers/research/2026-09-24-rung-5c-ii-plan-fold-verification.md`,
V1 through V9), which revision 2.1 (`0bb3722`) folded. Revision 2.2
folds the prose review
(`docs/superpowers/research/2026-09-24-rung-5c-ii-plan-prose-review.md`,
P1 through P15). The fold record, one disposition per finding in all
three folds, is
`docs/superpowers/research/2026-09-24-rung-5c-ii-plan-fold-record.md`.
Seven rulings for Geoff are open ("Rulings for Geoff").

Sources: the approved 5c-i plan
(`docs/superpowers/plans/2026-09-23-rung-5c-pages-member-verb.md`,
revision 2.2), whose "Pass 5c-ii" section specifies this pass's first two
tasks and whose resolutions R-5c12, R-5c13, R-5c16, and R-5c19 govern
them; pass 5c-i's close (`docs/HISTORY.md`, top entry) with its security
minors, its refusal of n5, and its carried items; the 5c-i security
review (`docs/superpowers/research/2026-09-24-rung-5c-i-security-review.md`);
the rung 5 spec (`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`,
revision 2.1 plus the 5c-i close's errata); the design-language
standard (`docs/superpowers/specs/2026-09-07-design-language-standard.md`,
DL below); the web surface standard
(`docs/superpowers/specs/2026-09-12-web-surface-standard.md`, revision
2.1, ratified 2026-09-23); the parallel-build approach's ownership table
(`docs/superpowers/specs/2026-09-03-parallel-build-approach.md`); and the
2.0 streamer spec (`docs/superpowers/specs/2026-09-07-streamer-2-0-design.md`),
"Forward hooks" and "Prior art". The approved 6a plan
(`docs/superpowers/plans/2026-09-23-rung-6a-store-foundation.md`) is the
structural exemplar and the concurrent lane. Code pointers were resolved
at `6de96a7`.

**Owner ruling in force (Geoff, 2026-09-24, `6de96a7`).** dubplate
replaces Navidrome at 2.0 and carries its own visual identity. No web
surface matches a streamer, no reference screenshots exist, and the
visual-fidelity chain applies to no rung. The render-and-read gate still
applies: every user-path page is rendered and read by the lane before
Geoff sees it.

**What this pass is.** Pass 5c-ii lands the Go half of rung 5's admin
surface and the delivery half of its web surface. Task 1 builds the admin
gate and the roster API. Task 2 mounts them in the serve shell, proves
the spec's cross-surface exit rows through the real shell, and closes
security n5 in the address normalizer. Task 3 births `web/` as a
SvelteKit static app with its gate leg in the same commit, so no `web/`
file ever lands ungated. Task 4 authors the one composed theme as data.
Task 5 lands the embed path and the R4.5 document rewrite's theme stamp
under a runtime prefix. Task 6 lands the generated branding stylesheet,
the operator CSS file, and the rewrite's two tier links beside the
routes they point at. Task 7 lands the served header set (CSP, `nosniff`,
`Referrer-Policy`, and HSTS), the cache rules, and the document routes,
the admin-gated add-member document among them. Task 8 is the close.

**What this pass is not.** No page content: the three routes ship as
structural stubs, and the login page, its string deck, the consumption
interstitial, and the add-member page are 5c-iii's (R-5cii1). No font
faces, no G4, no G5. No browser in the gate: G7's computed-value half and
G10 are 5c-iii's, with the rendered tooling. No `doctor` probe row, no
`ReadWritePaths`, no env template row, no trusted-proxy setting, no
runbook row, and none of the nine magic-link keys: those are 5d's. No
edit to `internal/config`, `internal/engine/store`, or
`cmd/dubplate-server/check.go`. **No deploy of any kind.** A task that
believes it needs any of these stops and reports.

> For agentic workers: workflow mode per CLAUDE.md's "Build process",
> the pass-execute chain per task, sequential, in the four segments
> "Execution model" names. The implementer is `dubplate-implementer`.
> The conductor consumes structured reports and never reads a diff, a
> test log, or a gate transcript inline. Subagents never dispatch
> agents, so `code-simplifier` and the three pass-end reviewers are the
> conductor's dispatch. The conductor owns `docs/STATUS.md`,
> `docs/HISTORY.md`, `ROADMAP.md`, the lane registry, and the
> contested-surface rows. **This pass runs in lane 1, the main tree,
> beside 6a in lane 2's worktree.** The live-executor check is the lane
> registry, never `pgrep -f Projects/dubplate`.

---

## Rulings for Geoff

Seven questions, each with a recommendation. The rung 5 spec's rulings,
the 5c-i plan's six rulings, and the web surface standard's section 16
stand and are not reopened. The fold record gives each question's
history (`docs/superpowers/research/2026-09-24-rung-5c-ii-plan-fold-record.md`).

**1. Does the remaining 5c scope split into 5c-ii and 5c-iii, with 5c-ii
under a 9.4M execution ceiling?** Recommendation: yes.

**Reason:** the remaining 5c scope is thirteen tasks plus Task 0, past
this build's twelve-task line. 5c-ii takes the Go-side and delivery
work: eight tasks, projected at 6.75M ("Token ceiling"). 5c-iii takes
the fonts, the three pages with the string deck, the rendered gates, and
the render-and-read: about six tasks, provisionally 5.0M. The cut sits
after the served-header task, the seam the 5c-i outline named. The split
also gives the delivery path its own pass-end review before any page
builds on it.

**What yes builds:** this plan as written. At the close, the rung 5
spec's 5c row and the web surface standard's section 15 take a second
split erratum. 5c's share of the rung reservation moves from about 18M
(the 5c-i plan's Ruling 6) to about 22M. Rung 5's reservation moves from
about 37M to about 41M. **What no builds:** one pass over all thirteen
tasks. Thirteen tasks at 0.725M (9.43M) plus both halves' adders (0.95M
and 0.60M) projects 10.98M, and 1.3 times that is a 14.3M ceiling with
its 80% flag at 11.4M.

**2. Does the served header set carry `Strict-Transport-Security:
max-age=31536000`, with no `includeSubDomains` and no `preload`?**
Recommendation: yes.

**Reason:** review finding 18 routes HSTS into this pass. A browser that
sees HSTS applies it to the whole host for the full age, so the shape is
visible to adopters and slow to reverse. The shape comes from RFC 6797
(R-5cii6). `includeSubDomains` would reach sibling hosts that an adopter
runs under the same domain, and dubplate does not own those hosts.
`preload` asks browsers to ship the host in their source, which only the
domain's owner can decide.

Under every answer, the header is sent only when `DUBPLATE_PUBLIC_URL`
uses https on a host that is not loopback. R-5cii6 gives the
correctness reason, so the exception is not asked.

**What yes builds:** R-5cii6 as written, with a named constant for the
age. **What no builds:** no HSTS header, and finding 18 stays open. A
configurable age or an `includeSubDomains` flag is an answer Geoff can
name instead.

**3. Does Task 4 author the theme now, as the `dubplate` match file and
the `gruvbox` palette file, without waiting on a UI mockup session?**
Recommendation: yes.

**Reason:** the three routes ship as unstyled structural stubs under
every answer, and the page design is 5c-iii's (R-5cii1). Geoff may run a
mockup session for those pages outside the tree, and its output feeds
5c-iii's plan. DL section 3 records the estate's configuration as `match
dubplate, palette gruvbox`. R36 requires the theme as data before G2 and
G9 can run. Every page binds to variable names only (spec, "The pages"),
so either file can be replaced without a markup change. G9 gates
contrast, so a replacement palette passes the same check.

**What yes builds:** Task 4 as written. **What no builds:** Task 4 waits
on the mockup session and takes its palette as input.

**4. Do `DUBPLATE_BRANDING` and `DUBPLATE_OPERATOR_CSS_FILE` default to
the literal `none`, meaning "not configured"?** Recommendation: yes.

**Reason:** the registry panics on an empty `KindString` default
(`internal/config/registry.go:322-325`). `ParseString` documents why
(`internal/config/thresholds.go:123-126`): "the one invalid Default is
the empty string, because wall.get's empty-equals-unset convention would
otherwise make an empty-string default indistinguishable from no default
at all". `cmd/dubplate-server/auth.go:33-36` records the same
constraint. 5c-ii may not edit `internal/config` (R-5c17), so the fix
lives in the keys' shape, and adopters see that shape (review C1).

**What yes builds:** both keys stay `KindString` thresholds, as the
spec's "Presentation (R14)" says. `doctor` prints their effective value,
and an unset or empty value resolves to `none`. `none` means no branding
declarations and no operator file. It cannot shadow a real file, because
R-5cii13 admits only paths ending in `.css`. **What no builds:** the two
keys become the `Optional` rows of a new integration class with an
enable flag, on the rotation secrets' precedent (`auth.go:30-36`).
`doctor` stops printing them, and the spec's "Each is `KindString`"
takes an erratum.

**5. Does `@types/node` join the web pins at an exact 24.x version, with
`web/scripts/**/*.ts` in the type-checked set?** Recommendation: yes.

**Reason:** `svelte-check` checks only what kit's generated
`.svelte-kit/tsconfig.json` includes: kit's generated type files, the
two vite configs, and `src/`, `test/`, and `tests/` (R-5cii16). Node's
type stripping does no type checking. So revision 1's "the gate scripts
are TypeScript under `ts-conventions`" gated nothing (review M2).
Including the scripts was measured, and it fails with `Cannot find name
'process'` without Node's types. The package is types-only and a
devDependency. The newest 24.x on 2026-09-24 is 24.13.6 (`npm view
@types/node@24 version`), matching the Node 24 floor.

**What yes builds:** R-5cii16 gains the row. `svelte.config.js`'s
`kit.typescript.config` hook appends the scripts to kit's generated
include list, so the gate type-checks every gate script and still checks
all of `src/` (R-5cii16, "How the scripts join the set"). **What no
builds:** the gate scripts are written as `.js` with JSDoc, and the
`ts-conventions` claim is dropped. Moving G1, G3, and the server-file
check into Go tests is an answer Geoff can name instead.

**6. Does the header set keep `Referrer-Policy: no-referrer`, recording
login from a browser without `Sec-Fetch-Site` as an unsupported
residual?** Recommendation: yes.

**Reason:** the Fetch standard serializes a non-CORS POST's `Origin` as
`null` under `no-referrer`. `CrossOriginProtection.Check` falls back to
`Origin` only when `Sec-Fetch-Site` is absent, and `null` fails its host
comparison. So a browser that sends no `Sec-Fetch-Site` (Safari before
16.4, iOS 15 devices) could fail every POST: request-link, consume, and
add (review R5). `go doc net/http.CrossOriginProtection` says
`Sec-Fetch-Site` is "available in all browsers since 2023", so only old
devices are exposed. WebKit's exact serialization is unverified.
`no-referrer` is the spec's value (R32), so yes needs no erratum.

**What yes builds:** R-5cii6 as written. The residual goes into the
close's gap list and is carried to 5d's runbook row. **What no builds:**
`Referrer-Policy: same-origin`, with an erratum to R32. It keeps every
cross-origin referrer empty and lets `Origin` serialize normally. The
token URL then reaches dubplate itself in `Referer` on the consumption
page's same-origin requests, and dubplate logs no `Referer`.

**7. Does a branding value outside the grammar, or an operator file that
is missing or refused, stop serve startup?** Recommendation: yes.

**Reason:** at 5c-ii a refused startup stops login only. A restart
surfaces the error at once. The box runs nothing yet. At rung 12 the
same refusal would also stop the streamer fallthrough, so a color typo
or a mount that is not ready would take the music server down (review
R7). The security property holds under either answer: a refused value
never reaches CSS.

**What yes builds:** R-5cii13 and R-5cii14 as written. The question is
carried to rung 12's plan, where availability is the larger stake.
**What no builds:** serve starts with an empty branding stylesheet and
no operator link, and logs the refusal at error level. Every refusal row
in Task 6 becomes a degrade row, and 5d's `doctor` row fails loudly.

---

### Spec deviations that take effect on approval

Approving this plan approves these. The close drafts the erratum text
(Task 8).

- R-5cii1: the 5c row splits a second time, into 5c-ii and 5c-iii.
- R-5cii3: the 5c-i outline's line that the gate's refusal "must track
  the served 404 document once W3.7's 404 lands" is withdrawn. W3.7
  gives every owned-path 404 to the Go mux, so no served 404 document
  exists for the gate to track.
- R-5cii12: the spec 5c row's "a rejected theme match value failing
  config load" reads as the serve verb's resolution refusing it. A
  threshold has no validation hook in `internal/config`, and R-5c17's
  rule keeps that package untouched. `doctor` reports the value without
  judging it; that row is carried to 5d.
- R-5cii15: G4, G5, G7's computed-value half, and G10 land in 5c-iii,
  with the fonts and the rendered tooling.
- R-5cii10: `go.mod` gains one `ignore` directive, and `.gitleaks.toml`
  gains one allowlist path.
- R-5cii20: the spec's embed-path line "Every served response carries
  `X-Content-Type-Options: nosniff` and `Referrer-Policy: no-referrer`
  beside the CSP" (`:1393-1394`) reads as: every served response carries
  the first two, and every served document carries the CSP. A 404, a
  JSON answer, an asset, or a stylesheet runs no script, and the gate's
  404 must equal the terminator's, which carries no CSP.

---

## Primary inputs

Read these before Task 1, in this order.

1. **The 5c-i plan's "Pass 5c-ii" section and resolutions R-5c12,
   R-5c13, R-5c16, R-5c19, and R-5c20.** Tasks 1 and 2 adopt its
   specification, updated where the tree moved (Task 0 item 2).
2. **The 5c-i security review**, in full. Its "attacked and found sound"
   list and the 5b-ii review's list are this pass's do-not-re-litigate
   record. S1 and S2 added four roster refusals that Task 1's outcome
   table must answer.
3. **The rung 5 spec**, by heading: "The web surface" (`:1337`), "The
   embed path" (`:1383`), "The pages" (`:1413`), "The gate leg"
   (`:1474`), "Packages and their contracts" (`:1486`), "Presentation
   (R14)" (`:1626`), R13 (`:464`), R14 (`:474`), R24 (`:587`), R32
   (`:671`), R36 (`:731`), R38 (`:773`), the threat model (`:1940`),
   the 2.0 hooks (`:1872`), the pass split (`:2472`), the contested
   surfaces (`:2499`), and risks 6, 7, 12, and 14 (`:2660` onward).
4. **DL**, sections 2 through 6: the 32-name contract, R3.2 through
   R3.7, R4.1 through R4.6, R5.6, and G1 through G10.
5. **The web surface standard**, section 15 (the clauses bound to 5c),
   W3.7 (404s), W5.9 and W5.10 (theme mechanics), W12.6 (where the
   rendered leg runs), and section 19.
6. **This plan's quoted behaviors.** Every relied-on stdlib, library,
   RFC, or tool behavior below is quoted from `go doc`, the package
   source, a man page, or a scratch probe run on 2026-09-24 against
   `go1.27.1 linux/amd64`, Node v24.20.0, npm 11.19.0, and the npm
   tarballs of `@sveltejs/kit@2.70.3` and `@sveltejs/adapter-static@3.0.10`.
   The probes live under
   `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/89b265c3-0b9e-4d9c-b998-df7b9970c32a/scratchpad/`
   (`kit/`, `n5/`, `ig/`, `npmls/`, `rfc6797.txt`, `dcp.txt`) and may be
   gone. Each task re-proves what it relies on through its own
   gate-resident test. A task that relies on a behavior this plan does
   not quote confirms it the same way and reports what it read.

---

## Goal

After this pass, an admin can list and add members through a
cookie-authenticated API whose every server answer to a non-admin
equals the answer to an unregistered path. (The route's existence is
public: the client router's manifest names `/members` in ungated `_app`
chunks, R-5cii19. The API is the boundary.) The `web/` tree exists, builds to a static bundle,
and is gated from its first commit. The Go binary serves each
prerendered document under any configured prefix, stamped with the
composed theme and linked to the branding and operator stylesheets in
tier order. Every response carries the header set, every document
carries a CSP whose script hashes cover its own inline scripts, and
HSTS follows the configured scheme. What 5c-iii adds is content: the
fonts, the pages, and the rendered proof.

---

## Owner touchpoints

The seven rulings above, at the plan-approval gate. Nothing else is
expected; no attended check occurs inside 5c-ii, since the pages it
serves are stubs (Ruling 3). What would become an owner touchpoint: a
new Go module dependency (this pass adds none); a web dependency beyond
the pinned set (R-5cii16, with Ruling 5 settling `@types/node` before
Task 3 dispatches); a threshold default changed on a task's own
judgment; any change to what the box runs.

---

## Token ceiling: 9.4M execution for 5c-ii (Ruling 1)

`spec:2548-2549`'s rule governs: each ceiling is re-derived from the
previous pass's actual. The basis is 5c-i's recorded actual, **about
5.8M over its eight tasks** (Tasks 1 through 7 and the close;
`docs/HISTORY.md`, top entry, "Conductor fold notes"): **0.725M per
task**. The rate amortizes Task 0, the fix rounds, the simplifier round,
the mail-copy register gate, the pass-end security read, its two-task
fold (about 0.78M), and three architecture reads. It excludes the
conductor's own turns, as the 5c-i figure does. Planning is spent
before Task 0 dispatches and is recorded against the rung, outside this
ceiling.

**The arithmetic for 5c-ii.**

1. Eight tasks (Tasks 1 through 7 and the close) at 0.725M is
   **5.80M**.
2. Three of the eight run at `opus` (Tasks 1, 6, and 7), the same share
   as 5c-i: add **0.00M**.
3. The first Node tree in the repo: a first install, a bundle build and
   `svelte-check` inside every gate from Task 3 on, and build output an
   implementer reads for the first time. Add **0.30M**.
4. Two pass-end reviewers beyond 5c-i's one (`svelte-reviewer` and
   `daisyui-a11y-reviewer`), over a small tree of stubs and theme data:
   add **0.15M**.
5. Five `go-architecture-reader` dispatches (`internal/auth`,
   `internal/identity`, `internal/serve`, `cmd/dubplate-server`, `web`)
   against 5c-i's three: add **0.10M**.
6. The review fold (revision 2) adds named states and proof rows
   without adding a task: Task 1's GET, HEAD, 405, `failed`, query, and
   Content-Type rows; Task 3's per-state seeds and the hidden-lockfile
   rows; Task 6's absorption of the rewrite's tier links with their
   anchor states and the operator path's refusals; Task 7's route-state,
   placeholder-tree, and minted-link rows. About 0.1M each: add
   **0.40M**.
7. **The projection is 6.75M.**

A ceiling of **9.4M** puts the 80% flag at **7.52M**, 0.77M above the
projection, about one task at the measured rate. The flag then marks a
real overrun rather than firing on plan. Either pre-named task split
(Task 3's gates to Task 4, or Task 7's document routes to their own
task) adds about 0.73M and brings the projection to 7.48M, just under
the flag; both together cross it, and the conductor asks at that
boundary under the 80% rule. Re-derive at the segment A and segment B
boundaries.

**5c-iii's provisional figure** is about 5.0M of execution: six tasks
at 0.725M (4.35M), the first rendered tooling (+0.25M), the string
deck's register gate (+0.15M), and the render-and-read (+0.20M), with
about 0.8M of planning. 5c's three passes then project about 17.6M of
execution (5.8M, 6.75M, and 5.0M). Planning adds about 4.3M: 5c-i's
1.5M estimate, since no actual is recorded
(`docs/superpowers/research/2026-09-23-rung-5c-plan-fold-record.md:511`);
5c-ii's 1.0M plus 1.0M for proposing this split (`:503-505`); and
5c-iii's 0.8M. That is about 22M against Ruling 6's 18M, which counted
5c-i's planning the same way (Ruling 1).

**Checkpoint interval:** STATUS at every segment boundary (after Task 2,
after Task 5, after segment C's reviews, and at the close), at most
three tasks apart. Every task reports its spend, its mutation outcomes,
and its measured gate wall time.

**Deliverable count:** thirty-one across Tasks 1 through 8, plus Task
0's report. Tasks 1, 2, 3, 5, 6, 7, and 8 carry four each; Task 4
carries three. No task passes four.

---

## Global constraints

- **Lane 1, the main tree, on master.** The conductor writes the lane
  registry row before Task 0. Lane 2 runs 6a in
  `.claude/worktrees/lane-2-6a`; **no task reads, writes, or runs a
  command inside `.claude/worktrees/`.**
- **No deploy.** No task runs `scripts/deploy.sh`, not even `--check`,
  and no task ssh's to the box.
- **Package territory, stated as a wall.** This pass writes:
  `internal/auth` (`admin.go` born; `handlers.go` for the `readForm`
  lift; their tests; `roster.go` for `actorRefusesRune`'s doc comment
  only, which n5 would otherwise leave false); `internal/identity/identity.go`
  and its test (n5);
  `internal/serve` (new files for the document, asset, and header
  handlers; `serve.go` for the header step; their tests);
  `cmd/dubplate-server` (`serve.go` edited, `web.go` born, `main.go`'s
  one registration line; their tests); the `web/` tree; `scripts/check.sh`;
  `Makefile`; `CLAUDE.md`'s gate paragraph; `.gitignore`;
  `.gitleaks.toml`; and `go.mod`'s one `ignore` line. **It writes
  nothing under `internal/config`, `internal/engine`, `internal/notify`,
  `internal/harness`, `internal/atomicfile`, `systemd/`, `env/`,
  `config/`, `corpus/`, `tools/`, or `_salvage/`, and nothing in
  `cmd/dubplate-server/check.go`, `check_test.go`, `doctor.go`,
  `doctor_test.go`, or `store.go`.** It writes no `docs/` file outside
  Task 8's drafts. A task that believes it needs one stops and reports.
- **No new Go module dependency.** `go.mod`'s only change is the
  `ignore ./web/node_modules` directive (R-5cii10). `go.sum` stays
  byte-identical.
- **The web dependency set is exactly R-5cii16's pins**, with
  `@types/node` among them only under Ruling 5. No other npm package
  enters `web/package.json`. `tus-js-client` stays out until
  rung 7, and no ESLint, Prettier, or Playwright package enters in 5c-ii.
- **`internal/auth` reads no environment and renders no HTML**, and
  **`internal/serve` imports neither `internal/auth` nor
  `internal/config`** (spec, "Packages and their contracts"). The
  composition of the gate over a document handler happens in
  `cmd/dubplate-server`.
- **Server-only SvelteKit files are forbidden**: no `+page.server.*`,
  `+layout.server.*`, `+server.*`, `hooks.server.*`, and no form action.
  Task 3's check gates it from the first commit.
- **Assertion discipline at the repo's bar.** No silent-success fakes, no
  vacuous assertions, no expectation derived from the code under test.
  Every hard-coded expectation cites the spec, DL, the web surface
  standard, an RFC, or a quoted document. Every criterion names the
  fixture state it needs.
- **A security property is proved by behavior, never by structure**
  (risk 3).
- **No token, cookie value, secret, or URL query string reaches a log
  record or an error string.** A document route logs its path and never
  `RawQuery`, because `<prefix>/login/link` arrives carrying the token.
- **Addresses reach exactly three outputs:** the roster audit record,
  `member list`'s stdout, and the roster API's GET answer to a verified
  admin. No other log record and no error string carries one.
- **Universality.** No estate hostname, path, email, or personal name
  anywhere, the web tree included. Every tunable is a threshold with a
  default or a named constant whose value this plan states.
- **Skills before code.** `go-conventions` before any Go, at the
  exemplary-stdlib bar. **Web tasks invoke `svelte-conventions`,
  `ts-conventions`, and the `daisyui` skill before writing a component,
  a TypeScript file, or markup.** The em dash is banned in Go comments.
  No process citation (a pass, a task, a ruling id, a plan line) in a
  shipped comment, in Go, TypeScript, Svelte, or CSS.
- **The gate is `bash scripts/check.sh`, green at every task end**, run
  through `cairn-run-gate`, re-issued on exit 75 until it prints `gate
  exit:`. Tasks 1 and 2 may run under `CAIRN_GATE_LANE=light`. From
  Task 3 on, the gate builds the bundle, and Task 3 reports whether the
  build's peak memory fits the light lane's 3G cap; the conductor sets
  the lane from that report. No gate in 5c-ii launches a browser.
  Anything reported as evidence runs with `-count=1` and reports wall
  time. Specific files only. Commit footer `Co-Authored-By: Claude Opus
  5.5 <noreply@anthropic.com>` for an `opus` task and the dispatch's
  footer otherwise.

---

## Privileged fields and their channels

CLAUDE.md's planning-miss rule asks that every privileged field name
every channel that may set it.

| Field | Channels that may set it | What refuses every other channel |
| --- | --- | --- |
| A member's `admin` flag | `dubplate-server member add --admin` through `identity.ChannelCommandLine` (5c-i). An operator's hand edit, validated at reload, with no audit record (5c-i's declared gap 2). | The roster API refuses any request whose `r.Form` carries an `admin` key, body or query, any value, empty included (Task 1). The store refuses the flag on every channel but the command line (5c-i Task 6). The seam reads the flag from the roster record, never the request (found sound). |
| The authority to write the roster through the web | A cookie-channel actor whose record in the disk copy, re-read under the store's lock, holds the flag (5c-i Task 6), reached only through the admin gate (Task 1; exported as `RequireAdmin` in Task 7). | A stale gate decision: `AddMember` re-checks the actor under the lock. A request that bypassed the gate: the inner handler refuses on its own when its context carries no valid admin identity (Task 1). |
| The roster's contents as a read (every address, admin flag, and streamer username) | The roster API's GET and HEAD to a verified admin (Task 1); `member list` on the command line (5c-i); 5b-ii's kind unknown-address answer, which hands out the one admin contact the spec names. | Every other requester of `<prefix>/roster` answers the gate's 404 (Task 1). The generic answer's setting governs the one contact the kind answer reveals (5b-ii). No log record or error string carries an address (global constraint). |
| The three member fields a POST writes | `r.PostForm` only (R-5c13): the request body. | The query string: a POST carrying any query key answers `admin_not_grantable`, `unknown_field`, or `malformed` before a field is read (R-5cii4), so an address never rides a URL into a proxy's access log. |
| The audit record's actor | The gate's verified, normalized address, passed through an unexported context key (Task 1). | Any form field, header, or query value; a spoofed `actor` field is refused as an unknown field and never read (Task 1). |
| The audit record's channel | The constant `identity.ChannelCookie` in the roster API. | Anything a request carries. |
| `streamer_username` | The roster API (Task 1) and the verb (5c-i), both through `AddMember`. A hand edit. **It becomes a privilege channel at rung 12**, carried there as 5c-i's declared gap 5; rows the API writes now persist into rung 12, so rung 12's check re-validates every existing row's principal, not only new writes (Task 8's gap list). | Nothing else writes the roster. |
| The refreshed session cookie | The gate, only after the admin check passes, and only for a request whose `Sec-Fetch-Site` is exactly `same-origin` (Task 1). | Every refusal, a verified non-admin included: the gate sets no cookie. The document handlers set no cookie. |
| The member address set (n5) | `identity.NormalizeAddress` on every channel: login, roster load, `AddMember` (Task 2). | Every channel reads through the one normalizer, so no second address path exists to forget the refusal. |
| The composed `data-theme` value stamped into a document | `DUBPLATE_THEME_MATCH` and `DUBPLATE_THEME_PALETTE`, validated at serve startup: the palette against `[a-z0-9-]+` first, then the composed name as an exact token against the theme selectors extracted from the built bundle (Task 5, R-5cii12). | Any request component: no route reads a query value, a header, or a cookie for the theme at 5c-ii. W5.10's signed-out cookie is rung 7's and is not built here. |
| The branding stylesheet's declarations | `DUBPLATE_BRANDING`, validated at serve startup: names from DL R2.1's set only, values from R-5cii14's grammar (Task 6). | Any request component. A value outside the grammar refuses startup rather than being escaped into place. |
| The operator CSS file's bytes | The file at `DUBPLATE_OPERATOR_CSS_FILE`, read once at serve startup, only when its path ends in `.css` and its content is UTF-8 with no NUL byte (Task 6, R-5cii13). | Any request-supplied path component: the route serves one in-memory value at one fixed path. |
| The CSP header's script hashes | The embedded build's own SvelteKit CSP meta, per document, read at serve startup (Task 7). | No configuration key and no request component contributes a source to `script-src`. |
| The presence of `Strict-Transport-Security` | The scheme and host of `DUBPLATE_PUBLIC_URL`, read at serve startup (Task 7, Ruling 2). The key is always set when serve runs: the magic-link source, the only declared identity source, lists it as Required (`cmd/dubplate-server/auth.go:50`), and the selector is required (`resolveIdentitySource`, `internal/config/config.go:291-296`). | The request's own scheme and every forwarded header: the ingress strip removes the forwarded family before any handler runs (5a). |

---

## Resolutions this plan takes rather than discovers

Each resolution names its site. A task landing the surface a resolution
decides writes nothing about the resolution in a shipped comment; the
close maps each resolution to its landing commit from the task reports.

### R-5cii1: the split, and where it cuts

The remaining 5c scope splits after Task 7 (Ruling 1). 5c-ii carries
the Go head (the gate, the API, the wiring, n5), the web tree born with
its gate leg, the theme as data, the embed path and rewrite, the
branding and operator stylesheets, and the served header set with the
document routes. 5c-iii carries the fonts with G4 and G5, the login page
and its string deck with the deck's register gate, the consumption
interstitial, the add-member page, the rendered sub-leg (G7's
computed-value half and G10), and the render-and-read at W4.1's five
viewports.

The cut keeps every 5c-ii export's caller inside 5c-ii (R-5cii18). It
also keeps the delivery path reviewable on its own: the pass-end
`web-auth-security-reviewer` reads the gate, the API, the header set,
the CSP, and the embed path before a page builds on them, which is the
same ordering argument the 5c-i split made for the security fold.

### R-5cii2: the runtime prefix rides SvelteKit's relative base

The 5c-i outline named the runtime prefix against SvelteKit's
build-time `paths.base` as "the pass's largest unknown". The kit source
answers it without a `reroute` hook.

- `kit.paths.relative` defaults to true (`types/index.d.ts:667-669`,
  kit 2.70.3): "If `true`, `base` and `assets` imported from
  `$app/paths` will be replaced with relative asset paths during
  server-side rendering, resulting in more portable HTML."
- During prerendering, the page renderer derives the base from the
  document's own depth and emits a client expression, not a literal
  (`src/runtime/server/page/render.js:121-133`): `const segments =
  pathname.slice(paths.base.length).split('/').slice(2); base =
  segments.map(() => '..').join('/') || '.'; base_expression = \`new
  URL(${s(base)}, location).pathname.slice(0, -1)\`;`. That branch runs
  only when `!state.prerendering?.fallback`.
- The client reads that runtime value
  (`src/runtime/app/paths/internal/client.js`: `export const base =
  __SVELTEKIT_PAYLOAD__?.base ?? __SVELTEKIT_PATHS_BASE__;`), and route
  resolution slices it off the path
  (`src/runtime/client/client.js:1641`: `url.pathname.slice(base.length)`).
- Vite's client build takes a relative base under the same condition
  (`src/exports/vite/index.js:1081,1089`: `const relative =
  kit.paths.relative !== false || !!kit.paths.assets;` then `base:
  relative ? './' : base`).
- Prerendered files are named by path
  (`src/core/postbuild/prerender.js:196-199`): `/login` is written as
  `login.html`, `/login/link` as `login/link.html`, and `/members` as
  `members.html`.
- adapter-static clears its output on every build
  (`adapter-static/index.js:62-63`: `builder.rimraf(assets);
  builder.rimraf(pages);`) and refuses dynamic routes without a
  fallback unless `strict` is false (`index.js:10-12`).

So the build uses `paths.base` empty, `paths.relative` at its default,
`trailingSlash` at its default of `'never'`, `prerender = true` and
`ssr = false` in the root layout, **and no fallback page**, since the
fallback branch is the one that keeps absolute paths. A document
computes its base from where the browser loaded it, so the same bundle
works under any prefix. **Two consequences bind the Go routes.** A
document is served only at its exact path: `<prefix>/login/` would
compute the base one level too deep, so a trailing-slash request
answers the terminator's 404 rather than the document (R-5cii21 states
the rule for every route). And the router
never rewrites a path to reach a document; it maps an exact request
path to one embedded file. The `reroute` hook and `router.type:
'hash'` are declined: the first is unneeded, and the second disables
prerendering. The mechanics review's probe at these pins confirmed it:
served under `/p/q`, all three documents hydrated in headless Chromium,
`$app/paths` `base` read `/p/q`, a `goto(base + '/members')` from
`/p/q/login` rendered the members page, and no request under the prefix
404'd. Task 0 re-runs the probe against the tree of its day, and Task 5
proves the reference half through the shell.

### R-5cii3: the gate's refusal stays `http.NotFound`

W3.7 settles the 5c-i outline's open line. "Two layers answer a 404 and
the Go mux answers first. A path under the owned prefix matching no
route returns 404 from the Go mux." The SvelteKit 404 "is a
client-router state", rendered by the root `+error.svelte`. So no
served 404 document exists, and the owned terminator's answer stays
`serve.Terminator()`, which calls `http.NotFound`
(`internal/serve/prefix.go:131-135`). The gate's refusal is that same
call, on the roster API and on the add-member document route alike.
`go doc net/http.Error`, which `NotFound` calls, states that it "sets
Content-Type to 'text/plain; charset=utf-8', and sets
X-Content-Type-Options to 'nosniff'", so both answers already carry one
header of the set. Task 7's header step wraps both identically.

### R-5cii4: the roster API's outcome table, updated for 5c-i's fold

The 5c-i outline's POST table predates 5c-i's S1 and S2. The store now
returns four refusals the table did not name
(`internal/auth/roster.go`): `errDuplicatePrincipal` (m1, line 118),
`errBlankDisplayName` (n3, line 131), `errRosterWritableByOthers` (m3,
line 241), and `errRosterWriteLanded` (m4, line 246). The table below
adds a row for each, and Task 1 adopts it.

The field set is closed. `r.Form` may carry `address`, `display_name`,
and `streamer_username`, each at most once, and nothing else. An `admin`
key keeps its own outcome (R-5c13). Any other key is `unknown_field`,
which closes `Admin`, `is_admin`, `actor`, and every spelling the
outline's `admin` rule left open. A repeated key is `malformed`, since
`url.Values.Get` would silently take the first.

**The three fields are read from `r.PostForm` only**, R-5c13's rule
restated: the admin and unknown-key checks run over `r.Form` (body and
query), and the values come from the body. A POST whose query string
carries any key that survived the admin and unknown-key checks answers
`malformed`, so `?address=...` with an empty body adds nobody and an
address never rides a URL into a tunnel or proxy access log. Once that
check passes, `r.Form` and `r.PostForm` hold the same three fields, so
the `r.PostForm` read is defense in depth: equivalent under the check,
so no mutation can tell it apart (V3).

**The Content-Type check parses the header.** `ParseForm` itself reads
the type through `mime.ParseMediaType`, and `go doc
mime.ParseMediaType` states that it "parses a media type value and any
optional parameters, per RFC 1521" and "On success, ParseMediaType
returns the media type converted to lowercase and trimmed of white
space and a non-nil map". The check compares that media type alone and
treats any returned error as a refusal, so a browser `fetch` with a `URLSearchParams`
body, which sends `application/x-www-form-urlencoded;charset=UTF-8`
(measured under Node 24's Fetch implementation), is accepted. Any other
media type, a missing header, or a header `ParseMediaType` refuses is
`malformed`: `go doc net/http.Request.ParseForm` states that "when the
Content-Type is not application/x-www-form-urlencoded, the request Body
is not read, and r.PostForm is initialized to a non-nil, empty value",
so such a request would otherwise surface as three missing fields.

**The order of the refusals** is fixed: Content-Type, the body cap
(while `readForm` parses), a repeated key, an `admin` key, a key outside
the three, a query key, and only then the fields.

**The error mapping names every sentinel**, because every validation
refusal from `AddMember` wraps both `ErrRosterRecordInvalid` and its
specific sentinel (`roster.go`, `checkAddition` and `addTo`). The mapper tests
the specific sentinels first and never branches on
`ErrRosterRecordInvalid`; a mapper that did would turn
`errActorNotAdmin` into a 400. `errRosterBusy` maps to `busy`.
`errActorNotAdmin`, `errRosterWritableByOthers`, `errInvalidActor`, and
`errAdminGrantForbidden` map to `failed`. `errNoAdmin` is unreachable on
an addition and falls to `failed` if it ever surfaces.

`errRosterWriteLanded` gets its own outcome, `write_unconfirmed`,
because the member may be on disk. A generic `failed` would invite a
retry that meets `duplicate_address`. The page copy (5c-iii) tells the
admin to reload the list before retrying.

### R-5cii5: n5 closes in the normalizer, from stdlib tables

The 5c-i close refused n5 because "refusing it needs the derived Unicode
Default_Ignorable_Code_Point table, which the standard `unicode` package
does not expose". The derived table is absent, but every input to its
derivation is present. `DerivedCoreProperties.txt` (Unicode 17.0.0,
read 2026-09-24) gives the derivation: "Generated from
Other_Default_Ignorable_Code_Point + Cf (Format characters) +
Variation_Selector - White_Space - FFF9..FFFB (Interlinear annotation
format characters) - 13430..13440 (Egyptian hieroglyph format
characters) - Prepended_Concatenation_Mark". `go doc unicode` exposes
`Other_Default_Ignorable_Code_Point`, `Variation_Selector`,
`Prepended_Concatenation_Mark`, `White_Space`, and `Cf`, at `const
Version = "17.0.0"`. The probe evaluated that formula over every rune
and counted **4174**, which is the file's own "Total code points: 4174"
for the property.

The probe also measured the review's examples. U+115F, U+1160, U+3164,
and U+FFA0 (the Hangul fillers), U+034F (CGJ), and U+FE0F (VS16) are
all in the derived set. **U+2800 (braille pattern blank) is not**: it
is `So`, a visible symbol by category that renders blank. The
normalizer refuses the derived set plus U+2800 by name.

The refusal lives in `identity.NormalizeAddress`, so login, roster load,
and every write path share it. A hand-edited roster carrying such an
address then fails to load, which is R2's fail-closed rule applied, and
no production roster exists before 5d. Display names are unchanged:
Ruling 3 of the 5c-i plan admits ZWJ and ZWNJ there, and the address
set is the one n5 names.

The table is the toolchain's. A Go release that moves to a later
Unicode version can change the set; the test pins the named code points,
the set's size (4174), and the version it measured, so a change
surfaces as a failing row rather than a silent widening, and an
over-refusal (dropping `- Prepended_Concatenation_Mark` from the
formula) fails the size row. The mechanics review confirmed
the set equals, point for point, the 27 `Default_Ignorable_Code_Point`
ranges in the file.

**Three of the ten rows are regressions, not proofs.** The current
normalizer already refuses all of `Cf` (`identity.go:70-72`), so
U+200B, U+00AD, and U+E0001 fail before the change; the other seven
pass today in both the local part and the domain (both reviews probed
it). The test labels the three as pre-existing.

`actorRefusesRune`'s doc comment (`internal/auth/roster.go:646-651`)
says it refuses "the same set [identity.NormalizeAddress] refuses", and
`validateActor`'s says a command-line actor is "free of every rune an
address refuses". After n5 both are false. Task 2 rewrites the two
comments to state what the command-line rule refuses on its own terms
(control characters and `Cf`, `Zs`, `Zl`, `Zp`) and changes no code in
the file. The command-line actor is an operating-system user name, never
an address, and a cookie actor already passes through `NormalizeAddress`,
so the n5 set reaches every address the audit record names.

### R-5cii6: the header set is the shell's outermost step, and HSTS follows the scheme (Ruling 2)

`nosniff`, `Referrer-Policy: no-referrer`, and (under Ruling 2) HSTS are
set by a new first step in `Shell.Handler`'s composition, outside the
ingress strip. So every response the shell answers carries them: a
document, an asset, a JSON answer, a normalization refusal, an origin
refusal, the terminator's 404, and the gate's 404. The rung 5 spec
requires this for "every served response" (R32), and the equality of
the gate's 404 with the terminator's depends on both passing the same
step. The fallthrough slot is inside the step at rung 5, since the slot
is empty. Rung 12's proxy reconciles the step with the streamer's own
headers; the close carries that as a rung 12 edge.

The CSP is not in this step. It is per document (R-5cii7).

`Referrer-Policy` stays `no-referrer` under Ruling 6, with old-browser
login recorded as a residual.

HSTS: `max-age=31536000`, no other directive, set only when
`DUBPLATE_PUBLIC_URL`'s scheme is https and its host is not loopback by
`isLoopbackHost` (`cmd/dubplate-server/auth.go:474-482`), fixed under
every answer to Ruling 2 (V6). RFC 6797 section 8.3 matches a Known
HSTS Host by host alone, and section 8.1.1 exempts only IP literals. So
HSTS from `https://localhost` behind a local TLS proxy would pin every
`localhost` port in that browser for a year. `validatePublicURL` accepts
`https://localhost` (`auth.go:465`), and the exception reuses R38's own
loopback set. RFC 6797 (read 2026-09-24):
section 6.1.1, "The REQUIRED 'max-age' directive specifies the number
of seconds, after the reception of the STS header field, during which
the UA regards the host ... as a Known HSTS Host"; section 7.2, "An HSTS
Host MUST NOT include the STS header field in HTTP responses conveyed
over non-secure transport"; section 8.1 processes the header only for a
response "received over a secure transport"; and section 8.1.1, a host
matching "the IP-literal or IPv4address productions" is never noted.
The process speaks plain HTTP to the tunnel or the adopter's proxy,
while the browser's leg is the one RFC 6797 governs, and R38 already
makes that leg https everywhere but localhost. So the configured scheme
is the right input, and the request's own scheme is not. `internal/serve`
takes the decision as a construction value and reads no configuration.

### R-5cii7: the CSP header is built per document from SvelteKit's own hashes

The spec leaves the hashes' source to rung 5 (`:1362-1364`). This plan
takes SvelteKit's.

- kit's CSP mode (`types/index.d.ts:384-386`): "Whether to use hashes or
  nonces to restrict `<script>` and `<style>` elements. `'auto'` will
  use hashes for prerendered pages". The build sets `kit.csp.mode:
  'hash'` and DL R5.6's directive set.
- A prerendered page carries the policy in a meta tag
  (`types/index.d.ts:374`): "When pages are prerendered, the CSP header
  is added via a `<meta http-equiv>` tag (note that in this case,
  `frame-ancestors`, `report-uri` and `sandbox` directives will be
  ignored)". kit writes it as `<meta http-equiv="content-security-policy"
  content="${escape_html(content, true)}">` (`csp.js:327`).
- kit adds no style hash when `style-src` carries `'unsafe-inline'`
  (`csp.js`: `style_needs_csp = (directive) => !!directive &&
  !directive.some((value) => value === 'unsafe-inline')`). That matters,
  because a hash in the list would disable `'unsafe-inline'` in a CSP3
  browser, which kit's own comment states: "'unsafe-inline' is ignored if
  either a hash or nonce value is present in the source list".

- kit emits the boot script as a bare tag with no attribute in hash mode
  (`src/runtime/server/page/render.js:586-588`: `` body += `\n\t\t\t<script${
  csp.script_needs_nonce ? ` nonce="${csp.nonce}"` : '' }>${init_app}</script>\n\t\t`; ``),
  and the probe build's documents each carry exactly one such script.

At startup, the Go layer reads each embedded document's CSP meta,
unescapes its `content` with `html.UnescapeString` (`go doc`: "It
unescapes a larger range of entities than EscapeString escapes"), takes
the `sha256-` sources from its `script-src`, and serves DL R5.6's full
directive set as a header with those hashes. The header carries
`frame-ancestors 'none'`, which the meta cannot. The meta stays in the
document; a header and a meta policy combine restrictively (DL R5.6),
and the two agree on every directive the meta can carry.

**Every state the extraction can meet refuses startup**, naming the
document: no CSP meta; a meta whose `script-src` is absent; a
`script-src` carrying `'unsafe-inline'`; an inline script whose
`sha256` the startup computes and does not find in the meta's list
(which covers a document with scripts and no hash, and a script added to
`app.html` outside kit's set); and a script opening tag the extractor
does not recognize, matched case-insensitively as `<script` followed by
anything but `>` (an attributed or `type="module"` inline script, which
kit does not emit, refuses rather than slipping past the scanner).

G6 (Task 7) then proves the outcome independently: the test hashes
every inline `<script>` body of each served document with
`crypto/sha256` and `base64.StdEncoding` and asserts each hash is in
that document's header. **G6's scanner is written in the test file and
shares no function with `internal/serve`'s extractor**, so one blind
spot cannot hide a script from both. A drift between kit's hash and the
bytes it emitted fails G6, which is the point. The R4.5
rewrite touches no script bytes (DL R5.6), and Task 5 proves it.

### R-5cii8: the cache directives, derived rather than copied

The salvage handler's `private` fit an authenticated surface (spec,
"The embed path"). Derived per response class here:

| Response | `Cache-Control` | Reason |
| --- | --- | --- |
| An asset under `_app/immutable/` | `public, max-age=31536000, immutable` | Content-hashed names never change content, and the asset carries no credential. |
| Any other embedded file (`_app/version.json`, and the fonts 5c-iii adds) | `no-cache` | The name is stable across builds. |
| A document (`login`) | `no-cache` | R4.5 and DL R4.2. |
| The consumption document (`login/link`) | `no-store` | R32: the one document whose URL carried a token. |
| The add-member document (`members`) | `private, no-cache` | It answers per requester, behind the gate. |
| The branding and operator stylesheets | `no-cache` | Configuration changes at restart without a name change. |
| Every JSON answer | `no-store` | The 5b-ii precedent, kept. |
| Every 404 and every non-JSON refusal | none set | `http.NotFound` sets none, and the gate's 404 must equal it. A JSON refusal (a 400 `malformed`, a 503 `busy`) is a JSON answer and carries `no-store`, as `writeAnswer` sets today (`internal/auth/handlers.go:416`). |

### R-5cii9: the embed lives in package `web`, and a build re-emits its placeholder

`go doc embed`: "Patterns may not contain '.' or '..' or empty path
elements", so the directive must sit in a Go file inside `web/`, and
the package is `github.com/glw907/dubplate/web` (`web/embed.go`), on the
salvage precedent (`_salvage/web/embed.go`). The pattern is `all:build`,
because SvelteKit writes its assets under `_app/`, and `go doc embed`
states that a directory pattern excludes "files with names beginning
with '.' or '_'" unless "a pattern begins with the prefix 'all:'".

adapter-static deletes its output on every build (R-5cii2), so a
committed placeholder inside `build/` would show as deleted after each
build. The placeholder therefore lives in `web/static/placeholder.txt`,
which the build copies into `build/` every time, and `build/placeholder.txt`
is committed as the second copy that keeps a fresh checkout compiling,
the salvage mechanism exactly. `.gitignore` ignores `web/build/*` except
that one file. No route serves the placeholder.

The embedded tree stays rooted at `build`, and `cmd/dubplate-server`
strips it with `fs.Sub` (`go doc io/fs.Sub`: "Sub does not check if the
directory currently exists"), then checks that the tree is a real build
and refuses serve startup naming `npm run build` when it is the
placeholder-only tree. The spec's reason stands: the consumer is "the
one place able to return an error rather than panicking at program
init".

### R-5cii10: the bundle builds before any Go test reads it, and three repo files learn about `web/`

- **Order.** Go tests in `web` and `cmd/dubplate-server` read the
  embedded bundle, so the bundle must be current before `go test` runs.
  The Makefile gains a `web` target (install-state check, build, and the
  static web gates), and `check` depends on it first. The target joins
  the `.PHONY` list (`Makefile:52`): it has no file prerequisites, and
  beside the `web/` directory make otherwise treats it as up to date and
  skips its recipe while `check` exits 0 (V2, measured by the
  verification probe). `check.sh`'s
  `--scope-changed` path calls `make web` before its narrowed Go
  targets, and its `--reduced` path runs the build step alone before its
  `go test`. Every form builds; only `--reduced` skips the static web
  gates, since it is never a boundary gate.
- **`go.mod` gains `ignore ./web/node_modules`.** `go help packages`:
  "File system patterns expanded with the '...' wildcard exclude ...
  Directories matching an ignore directive in a module's go.mod file".
  `node_modules` is not in that list's other entries, and the probe
  measured it: a scratch module with `web/node_modules/pkg/a.go` listed
  `example.com/ig/web/node_modules/pkg` under `go list ./...`, and with
  the directive it listed only the root, with `go vet ./...` and `go
  mod tidy` both clean. `web/.svelte-kit` begins with a dot and is
  excluded already.
- **`.gitleaks.toml` gains `web/.svelte-kit/`** in its allowlist paths.
  The full leg runs `gitleaks detect --no-git -s .`
  (`scripts/check.sh:400`), which walks untracked files. The mechanics
  review measured gitleaks 8.30.1 under `[extend] useDefault = true`:
  the built-in allowlist already skips `node_modules` and
  `package-lock.json` (0 bytes scanned), so no entry is added for them;
  without the `.svelte-kit` entry the scan read 563 KB of generated
  files and took 2.5 s against 0.35 s. `web/build/` stays scanned
  (97 KB, clean), since it is what the binary embeds. The allowlist's
  comment names the generated directory as the reason.
- **`.gitignore`** gains `/web/node_modules/`, `/web/.svelte-kit/`, and
  `/web/build/*` with `!/web/build/placeholder.txt`.

### R-5cii11: the install-state check covers three states, measured

R13 makes Node and npm hard requirements. The leg also refuses a stale
install, because a gate that builds against the wrong dependencies is a
check that does not gate. The probe measured `npm ls` (npm 11.19.0) on a
one-dependency project: exit 0 when installed; **exit 1** with
`node_modules` absent; **exit 1** ("invalid") when `package.json`'s
range excludes the installed version; and **exit 0** when the lockfile
pins a newer version inside the same range while the old one is
installed. So `npm ls` alone misses the in-range drift, which is the
normal state after a lockfile bump. The leg adds a comparison of
`web/package-lock.json`'s resolved versions against
`web/node_modules/.package-lock.json`, npm's hidden lockfile (`npm help
package-lock-json`: "npm as of v7 uses a 'hidden' lockfile present in
node_modules/.package-lock.json").

| State | Report |
| --- | --- |
| `node` or `npm` not on `PATH` | fails naming the tool and `mise` as the install path |
| Node major below 24 | fails naming the found and required versions |
| `web/node_modules` absent | fails naming `npm ci --prefix web` |
| an installed package outside its declared range | fails naming `npm ci --prefix web` |
| an entry present in both lockfiles at different versions | fails naming `npm ci --prefix web` and the first differing package |
| a lockfile entry marked `optional: true` absent from the hidden lockfile | proceeds (npm never installs another platform's binaries) |
| a lockfile entry not marked optional absent from the hidden lockfile | fails naming the package and `npm ci --prefix web` |
| installed and current | proceeds |

The comparison skips the root `""` key and compares `version` only for
entries present in both files. The mechanics review measured why the
optional row exists: `package-lock.json` held 116 entries and the hidden
lockfile 69, and the 46 missing were all `optional: true` platform
binaries (`@rolldown/binding-darwin-arm64` and its kind). A reading of
"every lockfile entry must be installed" would fail every clean
`npm ci`.

The leg never installs. An install is the lane's, per the spec's 5c
lane-launch item, and a gate that mutates `node_modules` would race a
second gate in the same tree. The one exception is Task 3's implementer,
who runs `npm install --prefix web` once against the exact pins to
produce `web/package-lock.json`; the conductor runs `npm ci --prefix web`
after Task 3 lands.

### R-5cii12: the presentation keys are four thresholds declared in `cmd/dubplate-server/web.go`

R14 declares the presentation keys in rung 5's own file with
`KindString`. The names and defaults, chosen here:

| Key | Default | Accepted |
| --- | --- | --- |
| `DUBPLATE_THEME_MATCH` | `dubplate` | exactly `dubplate` (DL R3.2) |
| `DUBPLATE_THEME_PALETTE` | `gruvbox` | a value matching `[a-z0-9-]+` whose composed name `<match>-<palette>` exists as a theme selector in the embedded bundle |
| `DUBPLATE_BRANDING` | `none` (Ruling 4) | `none`, or a declaration list under R-5cii14's grammar |
| `DUBPLATE_OPERATOR_CSS_FILE` | `none` (Ruling 4) | `none`, or a path to a file R-5cii13 admits |

**Why no default is empty.** The registry refuses an empty `KindString`
default at registration (`internal/config/registry.go:322-325`: `case
KindString: if t.Default == "" { panic(fmt.Sprintf("%s: invalid default
%q", t.Key, t.Default)) }`), and the wall reads an empty value as unset
(`internal/config/config.go:182-187`: `if !ok || v == "" { return "",
false }`). Revision 1's empty defaults would have panicked every verb at
`newProductionRegistry`. Under Ruling 4 the two optional keys take the
literal `none`; under Ruling 4's "no" they become an integration class's
`Optional` rows instead, and this table and the contested-surfaces
`doctor` row change with it.

**An empty value resolves to the default** for every key, by the wall's
convention: `DUBPLATE_THEME_MATCH=` stamps `dubplate`, and a test proves
it. No state row may promise that an empty value refuses.

**The composed-name check reads the built CSS as measured.** The
minified bundle writes the theme selector unquoted, inside a
`theme-controller` form as well:
`:is(:root:has(input.theme-controller[value=dubplate-gruvbox]:checked),[data-theme=dubplate-gruvbox])`
(mechanics review, measured at the pins). The check finds the CSS files
through each document's `rel="stylesheet"` links, extracts every
`[data-theme=<name>]` selector in its quoted and unquoted forms, and
matches the composed name as an exact token against that set. The
palette's `[a-z0-9-]+` rule runs first, so a value carrying `"`, `>`,
or whitespace never reaches the lookup or the attribute, and a prefix of
a real name (`gruv`) fails the exact match.

They register through one `registerWebConfig(reg)` call, one line in
`newProductionRegistry` (`main.go`), which is a contested line with 6a
(see "Contested surfaces"). `internal/config` is untouched (R-5c17).

**A threshold has no validation hook.** `go doc ./internal/config
Threshold` shows four fields, `Key`, `Kind`, `Default`, and `Doc`, and
no validator. So the refusal of a bad value happens in the serve verb's
resolution, before any listener opens, in the words the spec's exit row
asks for (the accepted set named). `doctor` prints each key's effective
value and does not judge it. That gap is carried to 5d's `doctor` rows,
and to 6c, whose settings table gains per-key metadata.

The keys are env thresholds at rung 5. The configuration homes spec
moves tunables to the store's settings table at 6c and makes the theme
a member preference; 6c migrates these four and is not blocked by them.

### R-5cii13: the operator CSS file is read once, bounded, and served at one path

The file is read once at serve startup, so a changed file needs a
restart. It is served at `<prefix>/operator.css` with `Content-Type:
text/css; charset=utf-8`, and the route reads no request-supplied
component (R14). Its size is bounded by `operatorCSSMaxBytes`, 262144
bytes (256 KiB), a named constant: well above any hand-written override
and well below anything that belongs in the binary's memory.

**The route is public, so the file's class is checked.** A wrong path
would otherwise publish whatever the service user can read: the roster
(mode 0600, owned by the service user, well under the bound) or a
secrets file (risk review R1). The path must end in `.css`, which no
roster, env, or secrets file the estate or an adopter's template names
does, and the content must be valid UTF-8 with no NUL byte. States:

| State | Report |
| --- | --- |
| key unset, empty, or `none` | no tier 3 `<link>`; the route answers the terminator's 404 |
| path not ending in `.css` (the roster fixture's path included) | serve refuses startup, naming the key and the rule, never the path's contents |
| path absent, not a regular file, or unreadable | serve refuses startup, naming the key and the reason, never the file's bytes |
| file over the bound | serve refuses startup, naming the key and the bound |
| content not valid UTF-8, or carrying a NUL byte | serve refuses startup, naming the key and the rule |
| file empty | served empty; the `<link>` stays |
| file present and within the bound | served as read |

### R-5cii14: branding is a declaration list with a closed grammar

`DUBPLATE_BRANDING` holds `;`-separated `name: value` pairs. The
generated stylesheet at `<prefix>/branding.css` declares one `:root`
block of those pairs and is empty when the key is unset, empty, or
`none` (DL R4.3, Ruling 4). The
stylesheet is always linked, so tier order never depends on
configuration.

- **Names** come from DL R2.1's 32-name set only; any other name
  refuses startup, naming it.
- **Values** follow per-name grammar. A `--color-*` name takes a hex
  color (`#` and 3, 4, 6, or 8 hex digits) or one of `oklch(`,
  `rgb(`, `hsl(` with numeric, `%`, `deg`, `/`, and space arguments. A
  radius, size, or border name takes a number with a `rem`, `px`, or
  `em` unit. `--depth` and `--noise` take `0` or `1`. Task 6 quotes
  DaisyUI 5's variable list for the per-name kind before relying on it.
- **Refused anywhere in a value:** `;`, `{`, `}`, `<`, `>`, `"`, `'`,
  `\`, `@`, `url(`, `var(`, `expression(`, a control character, and
  anything over 64 bytes. A refused value stops startup rather than
  being escaped into place, because an escaped value that still parses
  as CSS is the class SEC filed against the salvage render.

State table: `none` and empty key, one pair, many pairs, a 64-byte
value (accepted) beside a 65-byte one (refused), a repeated name (refused,
naming it), an empty pair from a doubled `;` (refused), a pair with no
`:` (refused), a name outside the set, and a value outside its grammar.

### R-5cii15: what lands where among DL's ten gates

Each gate lands with the artifact it reads, and 5c-ii's exit gate names
each one.

| Gate | Lands in | Reads |
| --- | --- | --- |
| Server-file check | Task 3 | `web/src` |
| G1 (no hardcoded colors) | Task 3 | `web/src` less `lib/themes/` |
| G3 (no external origin) | Task 3 | `web/src` and the built bundle |
| G2 (variable completeness) | Task 4 | the theme data files |
| G9 (contrast) | Task 4 | the composed theme's declared values |
| G8 (branding surface) | Task 6 | the generated stylesheet |
| G6 (headers) | Task 7 | each served document |
| G7, link-order and cache half | Task 7 | each served document and asset |
| G4 (font delivery), G5 (embedding) | 5c-iii | the faces 5c-iii adds |
| G7, computed-value half; G10 | 5c-iii | a rendered page (W12.6) |

G3, G6, and G7 read a built bundle or a served document, so each fails
by name against the placeholder-only tree (DL section 6). G4 and G5
would fail on an empty face set here, which is why they land with the
faces. The web surface standard's W12.6 already rules that a rendered
check runs at pass end and on tasks touching `web/`, never per commit
unconditionally.

### R-5cii16: the web dependency pins

The newest production version of each package on 2026-09-24 (`npm
view`), with one exception: TypeScript's newest is 7.0.2, but
`@sveltejs/kit@2.70.3` peers `typescript: '^5.3.3 || ^6.0.0'` and
`svelte-check@4.7.6` peers `'^5.0.0 || ^6.0.0'`, so TypeScript pins at
6.0.3, the newest in both ranges.

| Package | Version |
| --- | --- |
| `@sveltejs/kit` | 2.70.3 |
| `@sveltejs/adapter-static` | 3.0.10 |
| `@sveltejs/vite-plugin-svelte` | 7.3.1 |
| `svelte` | 5.57.1 |
| `vite` | 8.3.1 |
| `tailwindcss`, `@tailwindcss/vite` | 4.3.3 |
| `daisyui` | 5.7.45 |
| `svelte-check` | 4.7.6 |
| `typescript` | 6.0.3 |
| `@types/node` (Ruling 5 only) | 24.13.6 |

Every entry is a `devDependency`, since nothing Node ships to the box.
Exact versions, no ranges, with `package-lock.json` committed. Node is
24 or later: `node v24.20.0` runs a `.ts` file directly (measured:
`node t.ts` printed `4` for `const x: number = 2; console.log(x * 2);`),
so the gate scripts are TypeScript under `ts-conventions` with no
runner dependency. Type stripping does no type checking, and
`svelte-check` reads only kit's generated include set, so under Ruling 5
the gate scripts join that set (measured: without `@types/node` the
scripts fail on `process` and `node:fs`).

**How the scripts join the set (V1).** A child tsconfig's `include`
replaces the one it extends; it does not merge. The verification probe
measured it: `"include": ["scripts/**/*.ts"]` in `web/tsconfig.json`
checked 108 files and missed a type error seeded in `src/lib/bad.ts`.
Kit 2.70.3 generates its list in
`src/core/sync/write_tsconfig.js:58-85` (`'ambient.d.ts'`, `'env.d.ts'`,
`'non-ambient.d.ts'`, `'./types/**/$types.d.ts'`, the two vite configs,
then `${dir}/**/*.js`, `.ts`, and `.svelte` for `src/`, `test/`, and
`tests/`), and then returns `kit.typescript.config(config) ?? config`
(`:131`). Kit documents that hook as "a function that allows you to edit
the generated `tsconfig.json`. You can mutate the config (recommended)
... Note that any paths configured here should be relative to the
generated config file, which is written to `.svelte-kit/tsconfig.json`"
(`types/index.d.ts:851-861`). So `svelte.config.js` mutates, never
restates: `config.include.push('../scripts/**/*.ts')` and
`config.compilerOptions.allowImportingTsExtensions = true`. Restating
kit's list in `web/tsconfig.json` would also work today, but it would
drift silently at a kit bump that adds an entry. The flag is needed
because Node's type stripping requires a sibling import to name its
`.ts` extension, which strict checking otherwise refuses ("An import
path can only end with a '.ts' extension when
'allowImportingTsExtensions' is enabled"); kit's `noEmit: true`
permits it. Measured on a scratch copy of the probe app: 180 files
checked, a `./lib.ts` import between two scripts accepted, the seeded
`src/lib/bad.ts` error reported, and the only other errors the
`process` and `node:fs` names `@types/node` supplies. `web/tsconfig.json` sets `skipLibCheck: true`
under every answer to Ruling 5, since strict `svelte-check` otherwise
reports 41 errors inside kit's own `types/index.d.ts`. A later bump goes through the `dependency-upgrade`
skill. Task 0 re-reads each version and reports any newer one.

### R-5cii17: the review shape

R24's second `web-auth-security-reviewer` read runs at 5c-ii's end,
scoped to 5c-ii's own diff. Its surface is the gate, the roster API,
their wiring, n5, the header set, HSTS, the CSP, the embed path, the
rewrite, the branding and operator stylesheets, and the document routes.
The 5b-ii and 5c-i "found sound" lists are its do-not-re-litigate
record. `svelte-reviewer` and `daisyui-a11y-reviewer` read the `web/`
tree with the web surface standard as their lens (section 15). 5c-iii
runs all three again over its own diff, where R24's "returns for the
admin page" is met by the page itself.

### R-5cii18: nothing is exported ahead of a cross-package caller

`auth.NewRosterAPI` and `auth.AdminConfig` are called from
`cmd/dubplate-server` in Task 2. The gate is unexported (`requireAdmin`)
in Task 1, since `NewRosterAPI` composes it inside `internal/auth`, and
Task 7 exports it as `auth.RequireAdmin` in the commit that adds its
first outside caller, the `members` route. `web.Assets` is read
from `cmd/dubplate-server` (Task 5). Each new `internal/serve` export
has its caller in `cmd/dubplate-server` in the task that adds it. Every
new refusal sentinel is unexported unless a caller outside its package
branches on it.

### R-5cii19: the add-member document route composes gate, method, then document

`<prefix>/members` is `RequireAdmin` wrapping a GET-and-HEAD method
check wrapping the document handler. The order decides what a
non-admin learns: a non-admin's POST to `members` answers the gate's
404 like any other method, and only a verified admin meets the 405. The
roster API sits at `<prefix>/roster`, so the document and the API never
share a route (the 5c-i outline).

**The gate hides the members document; the route itself is public.** With `ssr = false`, the
members page component ships in public `_app/immutable/` chunks, the
client router's manifest names `/members` to every visitor, and a
client-side `goto` renders the page without requesting `members.html`
(measured). So the document's 404 keeps a non-admin from learning
whether their session is an admin's; it does not hide that the route
exists. The roster API is the authorization boundary, and nothing
sensitive may live in the page component (carried to 5c-iii's handoff).

### R-5cii20: the CSP rides documents only

R-5cii7's header is set on each document response and on nothing else.
The spec's embed-path sentence reads as all three headers on every
response (a deviation, listed above). A 404, a JSON answer, an asset, or
a stylesheet runs no script, and Task 7's mutation 5 depends on the
gate's 404 carrying no CSP, since the terminator's carries none.

### R-5cii21: every route answers its exact paths, for GET and HEAD, and nothing else

`Shell.Register` mounts both `<prefix>/<name>` and every path under
`<prefix>/<name>/` (`internal/serve/serve.go:94-96`), so each route
names what its deeper paths answer. One rule for all of them:

| Route | Exact paths it answers | Every other path under it | Methods |
| --- | --- | --- | --- |
| `login` | `<prefix>/login`, `<prefix>/login/link` | the terminator's 404, the trailing-slash forms included | GET, HEAD; any other answers 405 with `Allow: GET, HEAD` |
| `members` | `<prefix>/members` | the gate's 404 to a non-admin, the terminator's 404 to an admin | GET, HEAD after the gate (R-5cii19) |
| `roster` | `<prefix>/roster` | the terminator's 404 to everyone, from the exact-path wrap Task 2 mounts outside the gate (byte-equal to the gate's) | GET, HEAD, POST after the gate (Task 1) |
| `_app` | each embedded file under `_app/` | a directory path (`<prefix>/_app/immutable/`), the bare `<prefix>/_app`, and a missing file answer the terminator's 404; no listing, no redirect | GET, HEAD; any other answers 405 |
| `branding.css`, `operator.css` | the name itself | the terminator's 404 (`<prefix>/operator.css/x`) | GET, HEAD; any other answers 405 |

The `_app` handler is a lookup in the embedded tree, never
`http.FileServerFS`, which lists a directory and redirects a path ending
in `index.html`. A path carrying a `.` or `..` segment never reaches a
route: `Normalize` refuses it with 400 first (`internal/serve/normalize.go`,
rule 2), so a test aimed at a route uses a plain deeper segment.

---

## Review focus

These are the input classes most likely to bite, each pinned by a test
in the task that owns it.

1. **A non-admin probing the roster.** GET, POST, and a crafted `admin`
   field, as a verified non-admin. Expected: the terminator's 404 for
   every one, no `Set-Cookie`, and no byte on disk changed. Pinned in
   Tasks 1 and 2.
2. **An `admin` field spelled another way.** `Admin=on`, `admin=`,
   `?admin=`, `is_admin=1`. Expected: refused with no write. Pinned in
   Task 1.
3. **A document requested at a trailing-slash path.** Expected: the
   terminator's 404, never a document computing the wrong base. Pinned
   in Task 5.
4. **A document that gains an inline script the CSP does not cover.**
   Expected: serve refuses startup naming the document, because startup
   hashes every inline script against the meta's list (R-5cii7), and G6
   fails on its own scanner. Pinned in Task 7.
5. **Operator text shaped like CSS.** A branding value carrying `}` or
   `url(`. Expected: serve refuses startup; no stylesheet carries it.
   Pinned in Task 6.
6. **A token in the consumption document's URL.** Expected: no log
   record and no error carries the query string, and the response is
   `no-store`. Pinned in Task 7.
7. **An address with an invisible code point.** Expected: refused at
   login, at roster load, and on both write paths. Pinned in Task 2.

---

## Task 0: The staleness pre-task

Model: `opus`, fresh context, dispatched by the conductor before any
implementer. Reads this plan against the tree as shipped. **Its output
is blockers with evidence, not fixes.** It edits no tracked file. The
conductor rules each blocker before Task 1 dispatches. Every check runs
against the real HEAD, and the task reports that HEAD. A moved pointer
whose content survives is a note; a pointer whose content is gone is a
blocker.

**Blockers: a "no" stops the launch.**

1. **Is lane 1 free and lane 2 where the registry says?** Report the lane
   registry rows, `git worktree list`, and the main tree's `git status`.
   Uncommitted change under `internal/`, `cmd/`, `web/`, `scripts/`,
   `Makefile`, or `go.mod` in the main tree is a blocker. Read nothing
   inside `.claude/worktrees/` beyond `git worktree list`.
2. **Do 5c-i's surfaces match this plan?** Report, by symbol:
   `(*Roster).AddMember`'s signature; the unexported sentinel set in
   `internal/auth/roster.go` (every sentinel Task 1's outcome table maps, four of them by line in R-5cii4); `readForm`
   as a method on `*handlers`; `Seam.VerifyIdentity`'s signature and
   `VerifiedIdentity`'s four fields; `identity.Identity.Valid` refusing
   `ChannelCommandLine`; `auth.SessionCookie`'s signature;
   `serve.Terminator` calling `http.NotFound`; `subtreeNameRefusal`'s
   four refusals. A changed signature is a blocker.
3. **Has lane 2 landed anything on master?** Report `git log` for
   `cmd/dubplate-server/main.go`, `doctor.go`, `doctor_test.go`,
   `go.mod`, and `go.sum` since `6de96a7`. A landing changes Task 5's
   `main.go` line and the merge note, and is a blocker only if it
   touched a file this plan's wall names as 5c-ii's.
4. **The prefix probe (R-5cii2).** In a scratch directory outside the
   repo, build a SvelteKit app at R-5cii16's pins with `adapter-static`,
   `prerender = true`, `ssr = false`, no fallback, and three routes
   (`/login`, `/login/link`, `/members`), each setting a
   `data-hydrated` attribute from an `$effect`. Serve the output with a
   small Go program that maps `/p/q/login` to `login.html`,
   `/p/q/login/link` to `login/link.html`, and `/p/q/_app/...` to the
   assets. In headless Chromium, load each document under `/p/q` and
   report whether `data-hydrated` is set, whether any request 404s, and
   whether a client-side navigation from `/p/q/login` to
   `/p/q/members` resolves. A failure is a blocker; the conductor then
   rules between the `reroute` hook (the 5c-i outline's first
   candidate) and refusing a non-default prefix until rung 7. The
   mechanics review ran this probe green at the pins on 2026-09-24; the
   re-run guards against a moved pin. Report also the built CSS's
   theme selector form and each document's stylesheet link form, which
   R-5cii12 and Task 6's anchor rely on.
5. **Is every pin still the newest, and does TypeScript's peer range
   still exclude 7?** Report `npm view <pkg> version` for each
   R-5cii16 row (`@types/node@24` under Ruling 5) and the two peer
   ranges. A newer version is a note, and
   the conductor rules under `dependency-upgrade`; a peer change that
   admits TypeScript 7 is a note.
6. **Does the gate run cleanly at HEAD?** Run
   `CAIRN_GATE_LANE=light cairn-run-gate 'bash scripts/check.sh'` in
   the main tree and report the exit and wall time.

**Notes (not blockers):**

7. Confirm the three rejoin sentences at `CLAUDE.md:140`,
   `scripts/check.sh:5`, and `Makefile:33-34` (verified at `6de96a7`;
   the rung 5 spec's staleness list still cites `CLAUDE.md:118`).
8. Report `.gitleaks.toml`'s allowlist paths and `.gitignore`'s web
   entries (today only `_salvage/web` paths).
9. Report `check.sh`'s four forms and where each runs `make check`,
   `make vet lint tidy-check`, or `go test` (lines 276-300 at
   `6de96a7`), the sites R-5cii10 edits.
10. Report the spec headings this plan cites by line, and the web
    surface standard's section 15 clause list.
11. Report `newProductionRegistry`'s body and whether 6a's registration
    line has landed.
12. Report whether `registerServeConfig`'s three thresholds and
    `resolveServeConfig` still read as this plan assumes (R-5cii12's
    precedent).

**Deliverable count: 1** (the report).

---

## Task 1: The admin gate and the roster API

Model: `opus`. Depends on Task 0. Adopts the 5c-i plan's "5c-ii Task 1"
with R-5cii4's updates. Upshift reason: the gate is the one privileged
action's authorization, and its wrong variants (a refresh before the
check, a JSON 404, an ungated export) pass every happy-path test.

**Files:**
- Create: `internal/auth/admin.go`, `internal/auth/admin_test.go`
- Modify: `internal/auth/handlers.go`, `internal/auth/handlers_test.go`
  (lift `readForm` from a method on `*handlers` into a package function
  taking the cap, so the 5b-ii handlers and the roster API share one
  body path)

**Interfaces:**
- Consumes: `Seam.VerifyIdentity`; `SessionCookie`; `(*Roster).List`,
  `(*Roster).AddMember`; the unexported roster sentinels;
  `identity.ChannelCookie`; `identity.Identity.Valid`.
- Produces: the unexported gate `requireAdmin(cfg AdminConfig, next
  http.Handler) http.Handler`, which Task 7 exports (R-5cii18);
  `auth.NewRosterAPI(cfg AdminConfig) http.Handler`,
  which returns **the gated composition only**; `auth.AdminConfig{Seam
  Seam; Roster *Roster; ClientIP func(*http.Request) string; Now
  func() time.Time; Logger *slog.Logger; BodyCap int64}`. The verified
  identity reaches the inner handler through an unexported context key
  and accessor. The constant `rosterAPIWriteWait` (5 seconds, R-5c19)
  is copied by `NewRosterAPI` into an unexported field of the inner
  handler, which the in-package test constructs directly with a shorter
  wait; no package variable is a seam.

**Outcomes:**

- **The gate** (R-5c12): verify through the seam, passing `sameOrigin`
  true only for `Sec-Fetch-Site: same-origin`; require
  `Identity.Valid()` and `Admin`; answer every refusal with
  `http.NotFound`, the terminator's own answer (R-5cii3); set no cookie
  on any refusal; log the refusal reason and the client IP, never the
  cookie value and never an address. **Only after the admin check
  passes**, set the refreshed cookie when the seam returned one, then
  call the inner handler with the verified identity in the context.
  `go doc net/http.CrossOriginProtection` states that "the GET, HEAD, and
  OPTIONS methods are safe methods and are always allowed", so the
  shell's origin check says nothing about a GET's origin, and the gate
  reads the header itself.
- **The inner handler refuses on its own** with the same `http.NotFound`
  when its context carries no valid admin identity.
- **GET and HEAD** answer `no-store` JSON: a `members` array, each entry
  with `address`, `display_name`, `streamer_username`, and `admin`. A
  roster store failure answers 500 `failed`, never a partial list.
  Any other method answers 405 with `Allow: GET, HEAD, POST`, reached
  only by a verified admin. The API does not see its own path, since
  `AdminConfig` carries none; Task 2's mount confines it to the exact
  path (R-5cii21).
- **POST** takes a form-encoded body (R-5c13, R-5cii4): refuse a
  Content-Type whose `mime.ParseMediaType` media type is not
  `application/x-www-form-urlencoded`; cap at `BodyCap` before parsing,
  through the lifted `readForm`; refuse a repeated key, an `admin` key,
  any key outside the three, and any remaining query key, in that
  order, before reading a field; read the three fields from
  `r.PostForm`; then call `AddMember`
  under `context.WithTimeout(r.Context(), wait)` with the gate's actor,
  `ChannelCookie`, and `ClientIP`. `go doc net/http.Request.Context`
  says a server request's context is canceled only "when the client's
  connection closes, the request is canceled (with HTTP/2), or when the
  ServeHTTP method returns", so the lock wait needs its own bound.

**The POST outcome table:**

| Result | Status | `outcome` |
| --- | --- | --- |
| added | 201 | `added` |
| duplicate address | 409 | `duplicate_address` |
| duplicate streamer username | 409 | `duplicate_streamer_username` |
| duplicate principal (m1) | 409 | `duplicate_principal` |
| invalid member address (n5 included) | 400 | `invalid_address` |
| invalid display name | 400 | `invalid_display_name` |
| blank or absent display name (n3) | 400 | `blank_display_name` |
| invalid streamer username | 400 | `invalid_streamer_username` |
| missing streamer username | 400 | `missing_streamer_username` |
| an `admin` key in the body or query | 400 | `admin_not_grantable` |
| any other key outside the three | 400 | `unknown_field` |
| a repeated key, a non-form media type (`multipart/form-data; boundary=x` included), a missing or unparseable Content-Type, an unparseable body | 400 | `malformed` |
| a field key (`address`, `display_name`, `streamer_username`) in the query string | 400 | `malformed` |
| `application/x-www-form-urlencoded; charset=utf-8` with a valid body | as its fields decide | the parameter never refuses |
| body over the cap | 413 | `too_large` |
| the lock wait outlived its deadline (`errRosterBusy`) | 503 | `busy` |
| the write landed, then failed (`errRosterWriteLanded`, m4) | 500 | `write_unconfirmed` |
| `errActorNotAdmin`, `errRosterWritableByOthers`, `errInvalidActor`, `errAdminGrantForbidden`, `errNoAdmin` (unreachable on an addition), any other error | 500 | `failed` |

The mapper tests the specific sentinels and never
`ErrRosterRecordInvalid`, which every validation refusal also wraps
(R-5cii4).

The outcomes are specific because the caller is trusted
(`spec:1442-1443`); the copy a person reads is 5c-iii's deck.

**States each check meets:**

| Check | State | Report |
| --- | --- | --- |
| Cookie | absent, forged, expired, older than its epoch, member removed | the terminator's 404, no cookie |
| Roster store | fails on the seam's lookup | the terminator's 404, no cookie |
| Identity | valid, admin | inner handler |
| Identity | valid, not admin | the terminator's 404, no cookie, even when the seam refreshed |
| Identity | on `ChannelCommandLine` with `Admin` true (stub seam) | the terminator's 404 |
| `Sec-Fetch-Site` | `same-origin` | refresh honored after the check |
| `Sec-Fetch-Site` | `same-site`, `cross-site`, `none`, absent | no refresh |
| Form | each row of the outcome table | its row |

**Acceptance criteria:**

- **The gate's refusal table** is a test. Every refusal row answers the
  same status, the same headers (all but `Date`), and the same body as
  `http.NotFound` does on the same request, and carries no
  `Set-Cookie`. **Fixture state:** a roster holding both an admin and a
  non-admin, with valid sessions minted for both; the non-admin's
  request is same-origin and old enough that the seam returns a
  refreshed value, so a gate that set the cookie before the admin check
  fails the no-`Set-Cookie` assertion. The test first asserts that the
  seam returned a non-empty `RefreshedValue` for that session, so a
  session too fresh to refresh cannot pass the row vacuously. The
  non-admin rows carry an `admin=on` form value and an `X-Admin: 1`
  header, so a gate reading either (mutation 3) admits them.
- The inner handler built directly, with no identity in its context,
  answers the same `http.NotFound` bytes.
- A captured log for the forged-cookie row does not contain the cookie
  value, and no refusal log line, gate or POST, contains either fixture
  address. The `invalid_address` POST row submits an address `net/mail`
  quotes in its error (`roster.go:699-701`), so logging the parse error
  fails the row. **Fixture state:** the log sink holds at least one
  record for each request, so "contains none" reads real lines.
- **GET:** a roster holding an admin with a `streamer_username` and a
  command-line-created admin without one answers 200 with
  `Cache-Control: no-store` and a `members` array in file order, each
  entry carrying exactly the four fields with the fixture's values.
  **HEAD** answers the same status and headers with no body. **PUT**
  from an admin answers 405 with `Allow: GET, HEAD, POST`; from a
  non-admin, the gate's 404.
- **The GET store-failure row**, on the in-package inner handler with an
  admin identity in its context and the roster file corrupted after
  `OpenRoster`: 500 `failed` and no `members` key. (Through the gate the
  same fixture answers the gate's 404, since the seam's `Lookup` reads
  the same file, so the row cannot be proved through the composition.)
- **The `failed` rows**, each on the in-package inner handler with an
  admin identity in its context: (i) the actor's admin flag removed from
  the file by a hand edit after the identity was placed, which reaches
  `errActorNotAdmin` under the lock (**fixture state:** the edited
  roster still names a second admin, and the row first asserts that
  `OpenRoster` on the edited file succeeds, so the refusal is the
  under-lock actor check at `roster.go:798` and not the load's
  `errNoAdmin` at `roster.go:460`); (ii) the roster chmodded to 0666
  after `OpenRoster`, which `writeLocked`'s re-read refuses as
  `errRosterWritableByOthers`. Each asserts 500 `failed`, the bytes
  unchanged, and no audit record.
- **The Content-Type rows:** `application/x-www-form-urlencoded;
  charset=utf-8` with a valid body answers 201 `added`;
  `multipart/form-data; boundary=x` and a missing header answer 400
  `malformed`.
- **The query-field row:** a POST to `roster?address=<valid>&display_name=x&streamer_username=y`
  with an empty form body answers 400 `malformed`, the bytes unchanged.
  **Fixture state:** the query values would be accepted as a body, so
  without the query-key check the empty body reaches the field checks
  and answers a field outcome, never `malformed`. A `?admin=on`
  POST with a valid body answers `admin_not_grantable`, and a `?foo=1`
  POST with a valid body answers `unknown_field`.
- A same-origin GET with a valid admin session carries a `Set-Cookie`
  whose expiry moved forward and whose anchor did not; a `same-site` GET
  and a header-less GET carry none. **Fixture state:** a session old
  enough that refresh visibly moves its expiry.
- Every POST outcome row is a test. Each refused row asserts the roster
  file's bytes unchanged and no audit record. The `added` row asserts
  exactly one audit record naming the admin's normalized address as
  actor and `cookie` as channel. **Fixture state:** a roster file with
  at least two members, so "unchanged" compares real bytes.
- The `admin` row holds for `admin=on`, `admin=false`, and `admin=` in
  the body, and for `?admin=` in the query string. The `unknown_field`
  row holds for `Admin=on`, `is_admin=1`, and `actor=x`.
- The `write_unconfirmed` row injects a failure after the rename (the
  test's own path for `errRosterWriteLanded`, which the task names), and
  asserts the member is on disk.
- The `busy` row holds the directory lock from the test and asserts 503
  within the shortened deadline.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Answer a JSON 404 or a 403 for a verified non-admin → the
   `http.NotFound` equality row fails.
2. Set the refreshed cookie before the admin check → the non-admin row
   carries `Set-Cookie`.
3. Read the admin flag from a request field or header → the non-admin
   row reaches the inner handler.
4. Refresh on any request → the header-less row carries a cookie.
5. Check `admin` in `r.PostForm` only → the `?admin=on` row answers
   `malformed`, not `admin_not_grantable` (the query-key check still
   refuses it, so nothing writes).
6. Take the actor from the form → the `actor=x` row reaches the store.
7. Parse before capping → the over-cap row reads the whole body (a
   reader that counts bytes read).
8. Return the inner handler ungated from `NewRosterAPI` → a non-admin
   GET answers the list.
9. Skip `Identity.Valid` in the gate → the `ChannelCommandLine` stub
   row reaches the inner handler.
10. Accept keys outside the three → the `Admin=on` row writes.
11. Map `errRosterWriteLanded` to `failed` → the `write_unconfirmed` row
    fails.
12. Take the first of a repeated `address` → the repeated-key row writes.
13. Compare the raw Content-Type string → the charset row answers
    `malformed`.
14. Map `ErrRosterRecordInvalid` to 400 before the specific sentinels →
    `failed` fixture (i) answers 400.
15. Drop the query-key check → the query-field row answers a field
    outcome, not `malformed`. (Reading the fields from `r.Form` is an
    equivalent mutant under the check, R-5cii4, so it is not named.)
16. Log the address parse error on a POST refusal → the
    `invalid_address` log row finds the address.

**2.0 hook: neutral to the schema.** The gate is the one 2.0's
web admin surface reuses (streamer spec, "Web surface", revision 2.2).

**Prior art:** none external; the gate reuses the seam, the outcome
shape, and the body path the 5b-ii handlers use.

**Independent: no.** It opens the pass.

**Deliverable count: 4** (the gate with its refresh after the check,
GET, POST with its outcome table and deadline, the closed field set).

---

## Task 2: The serve wiring, the cross-surface exit rows, and n5

Model: `sonnet`. Depends on Task 1.

**Files:**
- Modify: `cmd/dubplate-server/serve.go`, `cmd/dubplate-server/serve_test.go`
- Modify: `internal/identity/identity.go`, `internal/identity/identity_test.go`
- Modify: `internal/auth/roster_test.go` (the n5 rows only: the load
  path and `AddMember` on both channels)
- Modify: `internal/auth/roster.go` (the doc comments of
  `actorRefusesRune` and `validateActor` only, R-5cii5; no code)

**Outcomes:**

- **`newServeShell` passes the seam it constructs into `AdminConfig`**
  and registers `NewRosterAPI(cfg)` as the subtree `roster`, wrapped in
  an exact-path check: a request whose `r.URL.Path` is not exactly
  `prefix.String() + "/roster"` goes to `serve.Terminator()`, and only
  the exact path reaches the gate (V4). The wrap is needed because
  `Shell.Register` mounts both the name and its subtree
  (`internal/serve/serve.go:94-96`: `s.mux.Handle(pattern, h)` and
  `s.mux.Handle(pattern+"/", h)`), and the shell dispatches through
  `s.mux.Handler(r)` (`serve.go:218`), which leaves `r.Pattern` unset,
  so the API cannot tell its exact path from a deeper one. A suffix
  test is refused: it answers the list at `<prefix>/roster/roster`. The
  gate's 404 and the terminator's are byte-equal, so a non-admin sees
  the same answer on either side of the wrap. The comment saying "No
  handler here consumes the seam" is removed. The
  body cap is the existing `DUBPLATE_AUTH_REQUEST_BODY_CAP` value.
- **The cross-surface exit rows are proved through `shell.Handler()`**,
  because a handler test cannot see the origin check or the prefix
  router.
- **n5 (R-5cii5):** `identity.NormalizeAddress` refuses any rune in the
  derived Default_Ignorable_Code_Point set and U+2800, wrapping
  `ErrInvalidAddress`, with an error naming the code point as `U+XXXX`
  and never the address.

**Acceptance criteria (every shell row through `shell.Handler()`):**

- A POST to `<prefix>/roster` with a valid admin cookie and neither
  `Origin` nor `Sec-Fetch-Site` is refused by the origin check, and the
  file is unchanged.
- The same POST with `Sec-Fetch-Site: same-origin` adds the member and
  writes one audit record.
- **A non-admin's GET and POST to `<prefix>/roster` are
  indistinguishable from the same requests to an unregistered
  `<prefix>/<unknown>`**: same status, same body, every header equal but
  `Date`, no `Set-Cookie`. The POSTs carry `Sec-Fetch-Site: same-origin`
  so both reach the gate and the terminator. With an admin session, the
  GET answers the list.
- **`member list` reads what the API wrote, and the API reads what the
  verb wrote.** **Fixture state:** the shell's roster opened before the
  verb's first write, and the verb opened after the API's write, so each
  side proves the other's reload without a restart.
- **n5 rows:** each of U+115F, U+1160, U+3164, U+FFA0, U+034F, U+FE0F,
  U+200B, U+00AD, U+E0001, and U+2800 embedded in an otherwise valid
  address is refused by `NormalizeAddress`, by the roster loader (a
  fixture file carrying one), by `AddMember` on both channels, and at
  request-link, which answers 400 `invalid_address` after
  `allowLinkClient` has charged the client (`handlers.go:194-208`).
  **Fixture state:** the same address without the code point is
  accepted by all four paths, so each refusal is the code point's. The
  task runs the table before the change and reports which rows fail:
  U+200B, U+00AD, and U+E0001 are refused today by the `Cf` rule and are
  labelled pre-existing regression rows in the table; the other seven
  must pass before the change. Each refusal's error text names the code
  point as `U+XXXX` and does not contain the address. A test asserts
  `unicode.Version` equals `"17.0.0"` and the derived set's size equals
  4174, so a toolchain move to a later Unicode version fails loudly for
  a reread and an over-refusing formula fails the size row.
- A verified admin's GET to `<prefix>/roster/x`, `<prefix>/roster/`,
  and `<prefix>/roster/roster` each answers the terminator's 404
  (R-5cii21). **Fixture state:** the same admin's GET to
  `<prefix>/roster` answers the list in the same test, so the 404s are
  the path's.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Mount the API outside the prefix → the admin GET answers 404.
2. Wrap the gate's refusal in a handler that adds a header → the
   indistinguishability row fails.
3. Construct a second roster value for the API → the verb-then-GET row
   still passes, since both reload; report that it does not fire and
   why, and the diff-reviewer confirms one roster value is shared.
4. Omit `Variation_Selector` from the derivation → the U+FE0F row
   passes the normalizer.
5. Refuse only `Cf` → the Hangul filler rows pass.
6. Drop the U+2800 rule → the braille-blank row passes.
7. Drop `- Prepended_Concatenation_Mark` from the derivation (U+0600
   through U+0605 are `Cf` and in that set) → the size row fails.
8. Confine the roster mount by `strings.HasSuffix(path, "/roster")` →
   the `<prefix>/roster/roster` row answers the list.

**2.0 hook: hook 5 (a stable identity), neutral.** The write path is
5c-i's; this task adds its second caller.

**Prior art:** Unicode's `DerivedCoreProperties.txt` formula, adopted as
a derivation from stdlib tables; `golang.org/x/text` declined, since
the stdlib carries every input and a new dependency is an owner
touchpoint.

**Independent: no** (Task 1's exports).

**Deliverable count: 4** (the mount, the cross-surface rows, the
two-way reload row, n5).

---

## Task 3: `web/` born, with its gate leg in the same commit

Model: `sonnet`. Depends on Task 0's prefix probe and Task 2. Skills:
`svelte-conventions`, `ts-conventions`, `daisyui`, before any file.

**Files:**
- Create: `web/package.json`, `web/package-lock.json`,
  `web/svelte.config.js`, `web/vite.config.ts`, `web/tsconfig.json`,
  `web/src/app.html`, `web/src/app.css`, `web/src/app.d.ts`,
  `web/src/routes/+layout.svelte`, `web/src/routes/+layout.ts`,
  `web/src/routes/+error.svelte`, `web/src/routes/login/+page.svelte`,
  `web/src/routes/login/link/+page.svelte`,
  `web/src/routes/members/+page.svelte`, `web/static/placeholder.txt`,
  `web/build/placeholder.txt`, `web/README.md`, and the gate scripts
  under `web/scripts/gates/` (server files, G1, G3, the install-state
  comparison)
- Modify: `scripts/check.sh`, `Makefile`, `CLAUDE.md` (the gate
  paragraph at `:140` and the web stack rule's `tus-js-client` line,
  which gains "from rung 7"), `.gitignore`, `.gitleaks.toml`, `go.mod`
  (the `ignore` line only)

**Outcomes:**

- **The scaffold** at R-5cii16's pins: `adapter-static` with no
  fallback; `prerender = true`, `ssr = false`, and default
  `trailingSlash` in `+layout.ts`; `kit.csp` in `'hash'` mode with DL
  R5.6's directive set; `paths` at defaults (R-5cii2); Tailwind 4
  through `@tailwindcss/vite`; DaisyUI 5 through the CSS `@plugin`
  directive with its built-in themes left enabled (W5.9), written as
  `@plugin "daisyui" { themes: light --default, dark; }`. A bare
  `@plugin "daisyui"` emits `@media (prefers-color-scheme:dark){:root:not([data-theme]){…}}`
  from its `dark --prefersdark` default, the block W5.9 forbids; the
  named line emits none and keeps `light` and `dark` for G10 (measured).
  `web/tsconfig.json` sets `skipLibCheck: true` and carries no
  `include` of its own; under Ruling 5, `svelte.config.js`'s
  `kit.typescript.config` hook appends `../scripts/**/*.ts` and sets
  `allowImportingTsExtensions` (R-5cii16, "How the scripts join the
  set"). The implementer runs
  `npm install --prefix web` once against the exact pins to produce the
  lockfile (R-5cii11). Three route
  stubs, each one `<main>` with one `<h1>` naming the page and nothing
  else (Ruling 3), and a root `+error.svelte` whose 404 is one human
  sentence (W3.7). Every `.svelte` file carries its `@component` block
  per `svelte-conventions`.
- **The leg**, in `check.sh` and the Makefile's `web` target: Node and
  npm required and named on failure; R-5cii11's install-state states;
  `npm run build --prefix web`; `svelte-check` with strict TypeScript;
  the server-file check; G1; G3. R-5cii10's order: `make check` runs
  `web` first; `--scope-changed` calls `make web` first; `--reduced`
  runs the build alone before its `go test`. `web` joins the `.PHONY`
  list at `Makefile:52` in this commit.
- **The rejoin sentences**, rewritten at all three sites to say the
  `web/` leg runs today and what it runs: `CLAUDE.md:140`,
  `scripts/check.sh:4-6`, and `Makefile:33-34` (plural there). The web
  surface standard's section 19 row for `CLAUDE.md` and `check.sh` is
  met for the static rules; its rendered half is 5c-iii's and the new
  sentences say so.
- **The repo plumbing (R-5cii10):** the `go.mod` `ignore` directive, the
  one gitleaks allowlist path with its comment, and the `.gitignore`
  entries.

**The server-file check's states:**

| State | Report |
| --- | --- |
| no forbidden file under `web/src` | passes, printing the count of files it read |
| `web/src` absent or empty | fails by name: an empty iteration is not a pass |
| one forbidden file (`+page.server.ts`, `+layout.server.js`, `+server.ts`, `hooks.server.ts`) | fails naming the file |
| several | fails naming every one, not only the first |
| an `actions` export in any `+page` file | fails naming the file |

G1's and G3's states follow the same shape: absent or empty input fails
by name, and a violation names every site. G3 reads both the source and
the built bundle; its built-bundle half fails by name against a
`web/build` holding only `placeholder.txt`.

**Acceptance criteria:**

- The build writes exactly `login.html`, `login/link.html`,
  `members.html`, `_app/`, and `placeholder.txt` into `web/build/`, and
  `git status` is clean after a build (the placeholder round-trips).
- **Every state row of every check is seeded**, and the task reports
  each failure's text. The server-file check: `web/src` absent, `web/src`
  empty, one `+page.server.ts` stub, three forbidden files at once (all
  three named), and an `actions` export. G1: a `bg-red-500` class. G3: an
  `@import url(https://...)` in source, and the placeholder-only
  `web/build`. The install state (R-5cii11): `node` off `PATH` and a
  `node` shim on `PATH` printing `v22.0.0` (the version row), a deleted
  `node_modules`, a top-level package installed outside its range, the
  hidden lockfile missing one non-optional entry, and a lockfile edited
  to pin a newer in-range version of one **transitive** package
  (**fixture state:** top-level pins are exact, so only a transitive
  entry has an in-range newer version, and `npm ls` accepts it, so the
  row proves the comparison and not `npm ls`). The clean install's
  absent `optional: true` entries are the proceed row's fixture.
- **The `make` rows (V2):** with `web/` present and
  `web/node_modules` deleted, `make check` itself exits non-zero and
  its output names `npm ci --prefix web`; the same holds for `make web`.
  **Fixture state:** the `web/` directory exists and the target names no
  prerequisite, so make runs the recipe only because of `.PHONY`.
- **The `src/` row (V1):** a type error seeded in `web/src/lib/` fails
  `svelte-check` in the gate under every answer to Ruling 5, and, under
  Ruling 5, a type error seeded in a gate script fails it too.
- `go list ./...` lists no package under `web/node_modules` (**fixture
  state:** a `.go` file planted under `web/node_modules/` for the test
  and removed after).
- The gitleaks leg's scanned byte count and wall time with a built
  `web/.svelte-kit` present, measured with and without the allowlist
  entry, reported both ways (the entry must remove the generated tree
  from the scan).
- The task reports the full gate's wall time before and after, the
  bundle build's wall time, and the build's peak resident memory
  (`/usr/bin/time -v`), and whether that fits the light lane's 3G cap.
- `bash scripts/check.sh` green in all four forms, each run once and
  reported.

**Named mutations:**

1. Guard the leg with `command -v node || skip` → the missing-Node row
   passes.
2. Run the Go leg before the build → a clean checkout's Go tests read the
   placeholder (proved in Task 5; reported here as the ordering row).
3. Check only `+page.server.*` → the `hooks.server.ts` row passes.
4. Rely on `npm ls` alone → the drift row passes.
5. Drop the `ignore` directive → the planted `.go` row lists the package.
6. Skip every lockfile entry absent from the hidden lockfile → the
   missing non-optional row passes.
7. Under Ruling 5, drop `scripts/**/*.ts` from the checked set → a
   seeded type error in a gate script passes the gate.
8. Under Ruling 5, set `"include": ["scripts/**/*.ts"]` in
   `web/tsconfig.json` in place of the hook → a type error seeded in
   `web/src/lib/` passes the gate (the `src/` row). Under every answer
   to Ruling 5 the `src/` row is seeded, so a later `include` cannot
   narrow the check unseen.
9. Drop `web` from the Makefile's `.PHONY` list → the missing
   `node_modules` row passes `make check` (V2).

**2.0 hook: neutral.** The tree is where 2.0's device page and web admin
surface arrive (spec, "The 2.0 hooks").

**Prior art:** SvelteKit with `adapter-static`, adopted per CLAUDE.md's
web rule (Geoff, 2026-09-07). The salvage placeholder mechanism,
adopted. cairn's static rules, declined for 5c-ii (W12.8 leaves port or
lift to rung 7).

**Independent: no.**

**Pre-named split:** if the four deliverables cannot land in one green
gate, G1 and G3 move to Task 4 and the server-file check stays here,
since `web/` must not exist ungated.

**Deliverable count: 4** (the scaffold, the leg, the rejoin sentences,
the repo plumbing).

---

## Task 4: The composed theme as data

Model: `sonnet`. Depends on Task 3. Skills: `daisyui`, `ts-conventions`.
Ruling 3 governs the palette.

**Files:**
- Create: `web/src/lib/themes/dubplate.css` (the match block, DL R2.1's
  11 structural names), `web/src/lib/themes/gruvbox.css` (the palette
  block, the 21 palette names), `web/scripts/gates/g2.ts`,
  `web/scripts/gates/g9.ts`
- Modify: `web/src/app.css` (the theme binding), `scripts/check.sh` or
  the Makefile's `web` target (the two new gates)

**Outcomes:**

- **The data (R36):** one `dubplate` match block and one `gruvbox`
  palette block. The match block's `--size-field` meets DL R3.7's floor.
  Each font value ends in a CSS generic family (DL R2.7), naming no face
  yet, since the faces are 5c-iii's.
- **The binding:** the bundle carries exactly one composed DaisyUI theme,
  named `dubplate-gruvbox` (DL R3.6), whose 32 variables are the two
  files' union, composed when the bundle builds with no runtime
  JavaScript. The mechanism is measured: two `@plugin "daisyui/theme"`
  blocks carrying the same `name: "dubplate-gruvbox"`, in two imported
  files, build to one merged rule holding both files' variables, in
  `@layer base` (so the unlayered `branding.css` outranks it at any link
  order). The task quotes the theme-plugin syntax from
  `daisyui@5.7.45`'s source before relying on it. No
  `:root:not([data-theme])` guard and no `prefers-color-scheme` block
  (W5.9).
- **G2 and G9** in the leg, per DL section 6.

**States:**

| Check | State | Report |
| --- | --- | --- |
| G2 | no theme file | fails by name |
| G2 | a block missing one name, or carrying one extra | fails naming the file and the name |
| G2 | a font value without a generic tail | fails naming the variable |
| G2 | `--size-field` below the floor | fails naming the value and the floor |
| G9 | a pair below its floor with no exception marker | fails naming the pair and the ratio |
| G9 | an exception marker on a pair that passes | fails naming the stale marker |
| G9 | a value G9 cannot parse | fails naming it, never skipping it |

**Acceptance criteria:**

- The built CSS contains a rule whose selector carries
  `[data-theme=dubplate-gruvbox]` (the minifier drops the quotes;
  R-5cii12 quotes the measured form, and the match accepts either form)
  declaring all 32 names, each value equal to the data file's after both
  are parsed, never compared as raw bytes, since the minifier may
  rewrite a value's spelling. **Fixture state:** the built bundle, not
  the source.
- G2 and G9 each fail by name against a seeded defect per state row,
  and the task reports each text.
- Every hard-coded ratio floor in G9 cites DL R2.6.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Let G2 treat a missing name as inherited → the missing-name row
   passes.
2. Compare contrast on raw values without converting to sRGB → a
   seeded `oklch` pair's ratio differs from its expectation. The
   expected ratio comes from an independent source recorded in the test
   beside its citation: the CSS Color 4 specification's published
   OKLab-to-sRGB sample code run in a scratch, or a published reference
   pair. Never the implementer's own arithmetic, which would share the
   defect it tests.
3. Skip unparseable values in G9 → the unparseable row passes.

**2.0 hook: hook 5 through DL R4.2, neutral.** The per-member theme row
is rung 6's.

**Prior art:** DaisyUI 5's theme plugin, adopted; the gruvbox palette's
published values, cited by URL in the data file's header.

**Independent: no** (Task 3's tree).

**Deliverable count: 3** (the data, the binding, G2 and G9).

---

## Task 5: The embed path, the theme stamp, and the theme keys

Model: `sonnet`. Depends on Task 4. Skill: `go-conventions`.

**Files:**
- Create: `web/embed.go` (package `web`), `web/embed_test.go`,
  `internal/serve/document.go`, `internal/serve/document_test.go`,
  `cmd/dubplate-server/web.go`, `cmd/dubplate-server/web_test.go`
- Modify: `web/src/app.html` (the placeholder token),
  `cmd/dubplate-server/main.go` (one line: `registerWebConfig(reg)` in
  `newProductionRegistry`), `cmd/dubplate-server/serve.go`,
  `cmd/dubplate-server/serve_test.go`

**Outcomes:**

- **The embed (R-5cii9):** `web.Assets` over `all:build`; the consumer
  in `cmd/dubplate-server` applies `fs.Sub` and refuses serve startup on
  the placeholder-only tree, naming `npm run build`.
- **The rewrite's first half (DL R4.5, W3.7):** one documented rewrite
  applied to each prerendered document, computed once at startup, not
  per request. In this task it substitutes the placeholder token in
  `app.html` with the validated composed name and changes no other
  byte. Task 6 adds the two tier links in the same function, in the
  commit that adds the routes they point at, so every task's prefix
  proof is green on its own commit.
- **The theme keys (R-5cii12):** all four keys registered through
  `registerWebConfig`, the two optional ones with Ruling 4's defaults,
  consumed in Task 6. Serve's resolution refuses a match value outside
  `{dubplate}` with an error naming the accepted set, refuses a palette
  outside `[a-z0-9-]+`, and refuses a composed name that is not an exact
  token in the bundle's extracted theme selectors, naming the name.
- **The document routes, for `login` only**, in this task, and the
  `_app` asset route, both under R-5cii21: the subtree `login` answers
  `<prefix>/login` with `login.html` and `<prefix>/login/link` with
  `login/link.html`; the subtree `_app` answers each embedded file by a
  lookup in the embedded tree. Task 7 adds `members`, the headers, and
  the cache rules.

**States:**

| Check | State | Report |
| --- | --- | --- |
| Embedded tree | placeholder only | serve refuses startup naming `npm run build` |
| Embedded tree | a document the route table names is missing | serve refuses startup naming the file |
| Placeholder token | absent from a document | serve refuses startup naming the document |
| Placeholder token | present more than once | serve refuses startup naming the document |
| Match key | `dubplate` | proceeds |
| Match key | unset or empty | resolves to `dubplate` and proceeds (the wall's convention, R-5cii12) |
| Match key | anything else | serve refuses naming the accepted set |
| Palette key | unset or empty | resolves to `gruvbox` and proceeds |
| Palette key | composed name an exact token in the bundle's theme selectors | proceeds |
| Palette key | composed name absent | serve refuses naming the composed name |
| Palette key | a prefix of a real palette (`gruv`) | serve refuses naming the composed name |
| Palette key | carrying `"`, `>`, whitespace, or an uppercase letter | serve refuses naming the `[a-z0-9-]+` rule, never echoing the value into markup |
| `login` path | exact document path | the rewritten document |
| `login` path | the trailing-slash form, or any other path under the subtree | the terminator's 404 |
| `login` method | anything but GET or HEAD | 405 with `Allow: GET, HEAD` |
| `_app` path | an embedded file | the file |
| `_app` path | `<prefix>/_app`, `<prefix>/_app/immutable/`, or a missing file | the terminator's 404, with no listing and no redirect |
| `_app` method | anything but GET or HEAD | 405 with `Allow: GET, HEAD` |

**Acceptance criteria:**

- **The prefix proof, gate-resident:** with the prefix set to a
  non-default two-segment value, each served document's every `src` and
  `href` resolves, relative to that document's own URL, to a 200
  through `shell.Handler()`. At this task every reference is one the
  build emitted, since the tier links arrive in Task 6. **Fixture
  state:** a prefix the build never saw, and the real built bundle, so a
  root-absolute reference fails.
- **The prefix proof, recorded:** the lane renders each stub under the
  default and a non-default prefix in headless Chromium, reports whether
  the client router hydrated and whether any request failed, and
  attaches the screenshots to its report. This is evidence, not a gate
  (R-5cii15).
- The rewrite changes exactly the token: a test diffs each served
  document against its embedded original and asserts every inline
  `<script>` body byte-identical.
- The served document carries `data-theme="dubplate-gruvbox"` on its
  document element, and `DUBPLATE_THEME_MATCH=` (empty) stamps the same
  value.
- A clean checkout with no build fails the embed test by name (the
  ordering row Task 3 reported).
- Every state row is a test.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Embed `build` without `all:` → the `_app` assets are missing and the
   prefix proof fails.
2. (Withdrawn, V7: a planted second token is refused at startup by the
   "present more than once" row under the correct rewrite too, so the
   mutant cannot be told apart. That row is the guard: a document can
   carry only the one token, so a document-wide replace touches no
   script.)
3. Serve a document at its trailing-slash path → the trailing-slash row
   answers 200.
4. Accept any match value → the refused-match row passes.
5. Match the composed name as a substring → the `gruv` row proceeds.
6. Serve `_app` through `http.FileServerFS` → the directory row answers
   a 200 listing.

**2.0 hook: neutral.** DL R4.2's per-member stamp keys on rung 6's
surrogate id; this task stamps the operator's default only.

**Prior art:** the salvage embed package and static handler, adopted in
reasoning (the placeholder and `fs.Sub`) and re-derived (the cache
rule, R-5cii8, and the lookup that lists nothing, R-5cii21).

**Independent: no.**

**Deliverable count: 4** (the embed, the theme stamp, the theme keys,
the `login` and `_app` routes with the prefix proof).

---

## Task 6: The branding stylesheet, the operator CSS file, and the tier links

Model: `opus`. Depends on Task 5. Upshift reason: these are "the two
places operator text becomes attribute markup and CSS declarations"
(R14), the class SEC filed against the salvage render, and a grammar
that admits one wrong character passes every well-formed fixture.

**Files:**
- Create: `internal/serve/branding.go`, `internal/serve/branding_test.go`
- Modify: `internal/serve/document.go`, `internal/serve/document_test.go`
  (the tier links), `cmd/dubplate-server/web.go`,
  `cmd/dubplate-server/web_test.go`, `cmd/dubplate-server/serve.go`,
  `cmd/dubplate-server/serve_test.go`

**Outcomes:**

- **The branding stylesheet (DL R4.3, R-5cii14):** generated once at
  startup from `DUBPLATE_BRANDING`, served at `<prefix>/branding.css`,
  one `:root` block, empty when the key is unset, empty, or `none`.
- **The operator file (DL R4.4, R-5cii13):** read once at startup from
  `DUBPLATE_OPERATOR_CSS_FILE` under its class checks, served at
  `<prefix>/operator.css`.
- **The tier links (DL R4.1, R4.5):** the rewrite inserts the tier 2
  `branding.css` link always and the tier 3 `operator.css` link only
  when the operator key names a file, **after the last kit-emitted
  `rel="stylesheet"` link**. kit links every branch node's stylesheet
  in an `ssr = false` prerender (`src/runtime/server/page/index.js:143-146`,
  "provide nodes without running load functions so that the styles and
  fonts are linked in the head before CSR takes over"; `render.js:83`
  seeds the list from `client.stylesheets`, `render.js:287` adds each
  node's, and `render.js:330-345` emits one link per entry as
  `<link href="..." rel="stylesheet">`). The stubs carry one today;
  5c-iii's pages will add more, and anchoring on the first would put
  `branding.css` ahead of later bundle CSS.
- **G8** as a Go test: the generated stylesheet declares only names from
  DL R2.1's set and is empty when nothing is configured.

**Anchor states:**

| State | Report |
| --- | --- |
| zero kit stylesheet links in a document | serve refuses startup naming the document |
| one link | the tier links follow it |
| several links | the tier links follow the last, proved on a fixture document carrying two |

**Acceptance criteria:**

- Every state row in R-5cii13, R-5cii14, and the anchor table is a test,
  and each refusal names the key and never echoes a refused value.
- **The injection table:** each of `}`, `{`, `;` inside a value,
  `url(https://example.invalid/x)`, `var(--x)`, `expression(`,
  `</style>`, a backslash escape, a control character, and a 65-byte
  value refuses startup, and a 64-byte valid value is accepted.
  **Fixture state:** each refused value sits beside one valid pair, so
  the refusal is the bad value's and not an empty-list artifact.
- **The operator path's class:** the key pointed at the test's own
  roster fixture (a `.json` path) refuses startup naming the key and the
  `.css` rule; a `.css` file carrying invalid UTF-8, and one carrying a
  NUL byte, each refuse; `none` links nothing.
- A valid multi-pair value renders byte-for-byte to the expected
  stylesheet, whose text the test states literally.
- **The routes read nothing from the request (R-5cii21):**
  `<prefix>/operator.css/x` and `<prefix>/branding.css/x` answer the
  terminator's 404; `<prefix>/operator.css?path=/etc/passwd` answers the
  configured file; `<prefix>/operator.css/../x` answers `Normalize`'s
  400 before any route runs, recorded as the normalizer's row; a POST
  to either stylesheet answers 405 with `Allow: GET, HEAD`.
- **Link order:** the served document links the bundle stylesheet(s),
  then `branding.css`, then `operator.css`, with the operator link
  present when the key names a file and absent when it is `none`, proved
  both ways, and the branding link present when the branding key is
  `none`.
- **The prefix proof extends** over both stylesheet hrefs: under the
  non-default prefix each resolves to a 200 through `shell.Handler()`.
- The rewrite changes exactly the token and the two insertions: every
  inline `<script>` body byte-identical to the embedded original.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Escape a refused value instead of refusing it → the `}` row starts.
2. Allow `var(` → the `var(--x)` row starts.
3. Read the operator file per request → a test that edits the file
   after startup sees the new bytes.
4. Accept a name outside DL R2.1's set → G8's name row fails.
5. Omit the branding `<link>` when the key is `none` → the link-order
   row fails.
6. Anchor on the first stylesheet link → the two-link fixture's order
   fails.
7. Accept any operator path → the roster-fixture row starts and serves
   the roster's bytes.

**2.0 hook: neutral.** 6c moves these keys to the settings table.

**Prior art:** DL R4.3's generated-stylesheet mechanism, adopted as
ratified; a CSS parser dependency declined, since a closed grammar that
refuses is smaller and stricter than a parser that accepts.

**Independent: no.**

**Deliverable count: 4** (the branding stylesheet, the operator file,
the tier links with their anchor, G8).

---

## Task 7: The served header set, the CSP, the cache rules, and the add-member document route

Model: `opus`. Depends on Task 6. Upshift reason: the header step moves
the shell's documented composition order, and the CSP's failure mode (a
page that loads and never boots, risk 6) passes every server-side test
that does not hash the scripts itself.

**Files:**
- Create: `internal/serve/headers.go`, `internal/serve/headers_test.go`,
  `internal/serve/csp.go`, `internal/serve/csp_test.go`
- Modify: `internal/serve/serve.go` (the new first composition step and
  its doc), `internal/serve/doc.go`, `internal/serve/document.go`,
  `internal/serve/document_test.go`, `internal/auth/admin.go` and
  `internal/auth/admin_test.go` (the gate exported as `RequireAdmin`,
  R-5cii18), `cmd/dubplate-server/serve.go`,
  `cmd/dubplate-server/serve_test.go`

**Outcomes:**

- **The header step (R-5cii6):** `nosniff`, `Referrer-Policy:
  no-referrer` (Ruling 6), and, under Ruling 2, HSTS, set by the
  shell's new outermost step on every response. `doc.go`'s composition
  list gains the step as step 0.
- **The CSP (R-5cii7, R-5cii20):** built per document at startup from
  the kit meta's hashes, after the startup hashes every inline script
  against the meta's list, and DL R5.6's directive set; set on each
  document response only.
- **The cache rules (R-5cii8)**, per response class.
- **The add-member document route (R-5cii19, R-5cii21):** the subtree
  `members`, composed in `cmd/dubplate-server` as `auth.RequireAdmin`
  over the method check over the document handler, answering the exact
  path only.
- **The log channel:** `newServeShell` takes the `*slog.Logger` it
  passes to `serve.NewShell` (today `nil`, so records reach
  `slog.Default()`, `cmd/dubplate-server/serve.go:215`), so a test
  captures the whole shell's log without swapping a process global.

**The CSP extraction's states** are R-5cii7's five refusals (no meta; no
`script-src`; `'unsafe-inline'` in it; an inline script whose hash the
meta lacks; an unrecognized script opening tag) plus the passing state,
each a test.

**Acceptance criteria:**

- **G6:** the served document set equals the route table's three
  documents, asserted first. For each, the header carries `default-src
  'self'`, `script-src 'self'` plus one `sha256-` source for every
  inline `<script>` body (hashed by the test with `crypto/sha256` and
  `base64.StdEncoding`, through a scanner written in the test file that
  shares no function with the extractor), `style-src 'self'
  'unsafe-inline'`, `frame-ancestors 'none'`, `base-uri 'none'`,
  `object-src 'none'`, `form-action 'self'`, no `'unsafe-inline'` in
  `script-src`, and no `font-src`. **Fixture state:** the real built
  bundle, which carries at least one inline script per document
  (asserted, so the row cannot pass on zero scripts).
- **G6 and G7 against the placeholder-only tree:** each runs over an
  `fstest.MapFS` holding only `placeholder.txt` and fails naming the
  missing build, so neither can pass on an empty document set.
- **G7, link-order and cache half:** the document set equals the three
  documents, asserted first; the bundle stylesheet(s), then
  `branding.css`, then `operator.css`, with no `_app/` stylesheet link
  after `branding.css`; `login.html` served `no-cache`,
  `login/link.html` `no-store`, `members.html` `private, no-cache`, one
  `_app/immutable/` asset `public, max-age=31536000, immutable`,
  `_app/version.json` `no-cache`, and both stylesheets `no-cache`.
- **Every response carries the header step:** a document, an asset, a
  stylesheet, a JSON answer, a 400 from normalization, a 403 from the
  origin check, a 405, the terminator's 404, the gate's 404, and a 404
  for a path outside the prefix (the fallthrough slot, empty at rung 5),
  through `shell.Handler()`. HSTS is present on every one with an
  `https://` public URL on a non-loopback host, and absent on every one
  with `http://localhost` and with `https://localhost` (the fixed
  loopback exception, R-5cii6). **Fixture state:** all three configurations, so
  presence and both absences are each proved.
- **The non-admin rows for `members`:** a verified non-admin's GET and
  POST to `<prefix>/members` equal the terminator's answer to
  `<prefix>/<unknown>` in status, body, and every header but `Date`,
  with no `Set-Cookie`; an admin's GET answers the stamped document; an
  admin's POST answers 405; an admin's GET to `<prefix>/members/` and
  `<prefix>/members/x` answers the terminator's 404.
- **The minted link routes and never logs (risk review R4):** a link
  built through `auth.LinkURL` with a fixture token (the real parameter
  name, `t`, `internal/auth/token.go:64`) is requested through
  `shell.Handler()` and answers the `login/link` document with
  `no-store`, which binds `internal/auth`'s `linkPath` to the document
  route. The same token also rides a GET to the trailing-slash
  `<prefix>/login/link/`, a POST to `<prefix>/login/link`, and a request
  `Normalize` refuses. Across all four, the captured shell log holds no
  line containing the token. **Fixture state:** the log sink captured at
  least one record for each request, so "no line contains" reads real
  lines.
- Every CSP extraction state is a test.
- `bash scripts/check.sh` green.

**Named mutations:**

1. Put the header step inside the prefix router → the normalization
   400 and the out-of-prefix 404 lack `nosniff` or `Referrer-Policy`.
2. Set HSTS whenever the request arrived → the `http://localhost` row
   carries it.
3. Set HSTS on every https URL → the `https://localhost` row carries it.
4. Drop one hash → G6 names the unhashed script, and startup refuses.
5. Add `'unsafe-inline'` to `script-src` → G6's no-unsafe-inline row
   fails.
6. Serve the CSP on the gate's 404 → the members indistinguishability
   row fails, since the terminator's 404 carries none.
7. Order method check before the gate → a non-admin's POST to
   `members` answers 405.
8. Mark `login/link` `no-cache` → its cache row fails.
9. Log `r.URL.String()` on a document request → the minted-link row
   finds the token.
10. Let G6 call the production extractor → the diff-reviewer rejects
    the shared function (a structural check the reviewer confirms).

**2.0 hook: neutral.** The header step is where rung 12 reconciles the
streamer's own headers (carried as an edge).

**Prior art:** the salvage header set
(`_salvage/internal/serve/serve.go:119-128`), adopted for `nosniff`,
`Referrer-Policy`, and the four hardening directives; its HSTS comment
("HSTS is a Cloudflare edge setting") declined under finding 18 and
R-5cii6. SvelteKit's `kit.csp` hash mode, adopted as the hash source.

**Independent: no.**

**Pre-named split:** if the four deliverables cannot land in one green
gate, the `members` route moves to its own task after this one, and the
gate's export moves with it.

**Deliverable count: 4** (the header step with HSTS, the CSP, the cache
rules, the `members` route).

---

## Task 8: The pass close, and the 5c-ii to 5c-iii handoff

Model: `sonnet`. Depends on Tasks 0 through 7, the simplifier round, and
the pass-end reviews with their folds.

**Files (drafts only; the conductor folds and commits):**
- Create: `docs/superpowers/research/handoffs/draft-STATUS-5c-ii.md`
- Create: `docs/superpowers/research/handoffs/draft-HISTORY-5c-ii.md`

**Outcomes:**

- **The HISTORY draft:** what landed per task with commit ids; the gate
  evidence (`go test -count=1 ./...` wall time, per-package times for
  `internal/auth`, `internal/identity`, `internal/serve`, `web`, and
  `cmd/dubplate-server`, the web leg's wall time, and the build's peak
  memory); the resolutions R-5cii1 through R-5cii21 with where each
  landed; the seven rulings and their outcomes; the review findings and
  their dispositions; the declared coverage gaps; both budgets; and
  "what a later pass would be wrong to rediscover". The gaps are at
  minimum: the theme keys' validation lives in serve's resolution, and
  `doctor` does not judge them (R-5cii12, to 5d); the prefix proof's
  hydration half is evidence, not a gate (to 5c-iii); G4, G5, G7's
  computed half, and G10 are not yet gates (to 5c-iii); the n5 set is
  pinned to Unicode 17.0.0; the header step wraps the fallthrough and
  rung 12 must reconcile it; no ESLint comment config for TypeScript
  yet (to rung 7's W12 static rules). Also carried: rung 12 re-validates
  every existing row's principal, not only new writes, since the API
  writes `streamer_username` from 5c-ii on; the shared-origin residual
  (script on the streamer's same-origin pages can call
  `<prefix>/roster` with the admin's cookie) as a rung 12 threat-model
  entry; Ruling 7's availability question to rung 12's plan; Ruling 6's
  old-browser login residual to 5d's runbook row; and to 5d, that
  `deploy.sh` builds the bundle from a clean committed tree (`npm ci`,
  then the build) before `go build` and refuses a dirty `web/`, since
  the binary embeds whatever `web/build` holds.
- **The errata owed**, drafted paste-ready: the spec's 5c row and pass
  split (the second split); the web surface standard's section 15 line;
  the spec's "failing config load" row (R-5cii12); the spec's
  embed-path sentence on the CSP (R-5cii20); R-5cii3's withdrawal
  of the 5c-i outline's 404-tracking line; the spec's staleness item
  citing `CLAUDE.md:118`; and the HISTORY 5c-i entry's n5 sentence
  ("the standard `unicode` package does not expose" the table), which
  R-5cii5 corrects.
- **The 5c-iii handoff:** the surface it builds on (the routes, the
  gate, the API's outcome values, the header step, the CSP builder that
  hashes whatever the pages add); Ruling 3's outcome and any mockup
  session output as a standing input; the fonts' candidate under R19
  (Inter 4.1, OFL 1.1, license at
  `https://github.com/rsms/inter/blob/master/LICENSE.txt`, one variable
  file per style, which meets DL R5.5 without subsetting under R5.4),
  recorded as a candidate for 5c-iii's plan to adopt or decline; the
  rendered tooling W12.6 describes; that the members page component
  ships in public chunks, so nothing sensitive may live in it and the
  API alone authorizes (R-5cii19); and that `default-src 'self'` also
  governs `img-src`, so any `url(data:...)` a DaisyUI component or
  Vite's asset inlining puts in the built CSS is blocked, a conflict
  5c-iii's first styled page meets and decides.
- **The STATUS draft** stays present tense and under 60 lines.

**Acceptance criteria:** every count in the drafts was produced by a
command the report pastes; every commit id resolves; the STATUS draft is
at most 60 lines; `bash scripts/check.sh` green as a formality.

**Deliverable count: 4** (the HISTORY draft, the errata, the handoff,
the STATUS draft).

---

## The exit gate

`bash scripts/check.sh` green on master in all four forms, plus every
assertion below. Each maps to the spec's 5c row (`:2482`) or to this
plan's scope, and each names its proving task.

1. A non-admin session receives the terminator's 404 from the roster API
   and from the add-member document, every server answer
   indistinguishable from an unregistered path's (the route's existence
   is public, R-5cii19), while an admin receives the list and the
   document (Tasks 1, 2, 7).
2. An add-member POST without a same-origin signal is denied (Task 2).
3. An add-member POST that succeeds writes one audit record naming the
   actor and the channel; one that fails validation writes none and
   leaves the file unchanged (Task 1).
4. `member list` reads what the API wrote, and the API reads what the
   verb wrote (Task 2).
5. An address carrying a default-ignorable code point or U+2800 is
   refused on every channel (Task 2).
6. The server-file check, G1, G2, G3, G6, G7's link-order and cache
   half, G8, and G9 pass, and each fails by name against a seeded
   violation or the placeholder-only tree (Tasks 3, 4, 6, 7).
7. The served document carries the composed `data-theme` value and the
   three stylesheets in tier order (Tasks 5, 6, 7).
8. Documents are `no-cache`, the consumption document is `no-store`, and
   a hashed asset is immutable (Task 7).
9. `script-src` carries a hash for every inline script and no
   `'unsafe-inline'` (Task 7).
10. `nosniff` and `Referrer-Policy: no-referrer` are on every served
    response, and HSTS follows the configured scheme and host (Task 7).
11. No log record on the consumption document route contains the token
    (Task 7).
12. A rejected theme match value fails serve's resolution, naming the
    accepted set (Task 5).
13. The bundle serves under a prefix it was never built for, with every
    reference resolving (Tasks 5, 6).
14. Every route answers its exact paths only, and no directory is listed
    (R-5cii21; Tasks 5, 6, 7).

**Deferred to 5c-iii, by name:** G4, G5, G7's computed-value half, G10,
the three pages as authored, and the render-and-read at W4.1's
viewports.

---

## Contested surfaces

The conductor writes these rows into the parallel-approach spec's
ownership table before the lane launches.

| Surface | Owner | 5c-ii's position | Contention |
| --- | --- | --- | --- |
| `cmd/dubplate-server/main.go` | rung 1's skeleton | Task 5 adds one line in `newProductionRegistry` | **6a Task 7 adds one line in the same function**; a textual conflict keeps both in landing order, and the second lander's rebase is non-fast-forward |
| `go.mod` | shared | Task 3 adds one `ignore` directive; no requirement | **6a Task 2 adds `modernc.org/sqlite`**; the lines differ, and `go mod tidy` plus the merge ritual's in-tree gate resolve any overlap |
| `go.sum` | shared | untouched | 6a's |
| `cmd/dubplate-server/doctor.go`, `doctor_test.go` | rung 1's skeleton | **untouched**; the four new thresholds add four `doctor` rows through the registry (two under Ruling 4's "no"), and `TestDoctorCmd_LineCountFollowsRegistry` derives its count | 6a Task 7 edits both; if 6a's goldens (R-6a16) pin the full output, the second lander regenerates them in its rebase |
| `cmd/dubplate-server/check.go`, `check_test.go`, `store.go` | rung 6 | untouched | 6a's |
| `internal/engine/store` | rung 6 | untouched | 6a's |
| `internal/config` | rung 1b's contract | **untouched** (R-5c17) | 6c's condition for lane 2 |
| `cmd/dubplate-server/serve.go`, `serve_test.go`, `web.go` (new) | rung 5 | Tasks 2, 5, 6, 7 | none; 6a does not edit them |
| `internal/auth`, `internal/identity`, `internal/serve` | rung 5 | Tasks 1, 2, 5, 6, 7 | none |
| `web/` (new tree) | rung 5 births it; rung 7 authors visuals | Tasks 3, 4, 5 | 5c-iii adds the fonts and the pages |
| `scripts/check.sh`, `Makefile` | rung 5 adds the web leg | Task 3 | **after 5c-ii lands on master, every full gate builds the bundle**: lane 2's next rebase needs `npm ci --prefix web` in its worktree before its in-tree gate |
| `CLAUDE.md` gate paragraph | rung 5's pass 5c | Task 3 | conductor-committed docs in the main tree: Task 3's commit is path-limited |
| `.gitleaks.toml`, `.gitignore` | shared | Task 3 appends | every lane; an append conflict resolves in landing order |
| `docs/` ledgers | conductor | Task 8 drafts only | every lane |

---

## The conductor's actions

Not tasks; no implementer performs them.

1. Before Task 0: record the planning spend against the rung; write the
   lane 1 registry row and the contested-surface rows; arm the runaway
   guard; check AC power; run `npm ci --prefix web` only after Task 3
   lands (the tree does not exist before it).
2. Before Task 1: rule Task 0's blockers, the prefix probe's above all.
3. At segment A's boundary (after Task 2): STATUS; re-derive the
   ceiling from Tasks 1 and 2.
4. At segment B's boundary (after Task 5): STATUS; set the gate lane
   from Task 3's memory report; re-derive the ceiling.
5. At segment C's end (after Task 7): one `code-simplifier` round over
   Tasks 1 through 7's accepted Go and TypeScript, pasting
   `.claude/instructions/simplifier-brief.md` verbatim, its review
   chunked one `diff-reviewer` per package; then the three pass-end
   reviewers (R-5cii17), each asked for correctness gaps ranked by
   consequence, and one fold dispatch that may refuse a finding whose
   fix costs more than the risk it removes, with a one-line reason.
6. At the close: one `go-architecture-reader` per touched Go package
   (`internal/auth`, `internal/identity`, `internal/serve`,
   `cmd/dubplate-server`, `web`), no plan in context.
7. Tell lane 2, through the lane registry, that its next rebase needs
   `npm ci --prefix web` and a gate lane set from Task 3's memory report.
8. Record 5c-ii's actual against the ceiling and re-derive 5c-iii's
   figure before 5c-iii's plan is drafted.
9. **Do not deploy.**

---

## Execution model

Workflow mode: the pass-execute chain per task, **sequential**, one
invocation per segment. Every task runs `dubplate-implementer`, then
`diff-reviewer` (`claude-opus-5-5`), then the gate, with one re-dispatch
on `fix` and a second `fix` as the conductor's decision.

| Segment | Tasks | The boundary's own work |
| --- | --- | --- |
| **A** | 1, 2 | STATUS. Re-derive the ceiling. The Go head is complete and gradable. |
| **B** | 3, 4, 5 | STATUS. Set the gate lane. Re-derive the ceiling. |
| **C** | 6, 7 | The simplifier round, the three reviewers, the fold. STATUS. |
| **D** | 8 | The close drafts, the architecture reads, the ledger fold. |

Each boundary sits on a task's own green committed gate, so each is a
resume point. Segment A's boundary is also the seam where the Files
change from Go packages to the web tree.

**Models.** `sonnet` by default, at effort `high`. **Three upshifts to
`opus`, all unconditional**, dispatched as `claude-opus-5-5`: Task 1
(the one privileged action's authorization), Task 6 (operator text into
CSS), and Task 7 (the shell's composition order and the CSP). Task 0
runs at `opus` in fresh context.

**Dependencies, stated exactly:**

- Task 0 first; its blockers are ruled before Task 1.
- Task 1 → Task 2 (the gate and the API).
- Task 2 → Task 3 (the gate lane stays light through segment A; the
  web tree starts on a green Go head).
- Task 3 → Task 4 → Task 5 → Task 6 → Task 7 (each builds on the last
  task's tree, the leg, and the document handler).
- Tasks 0 through 7, the simplifier, and the reviews → Task 8.
- **Independent by files:** Task 2's n5 half and Task 1. They still run
  sequentially, because one tree runs one executor.

**Where this pass splits further, if it must:** at the Task 5 / Task 6
boundary, if segment B's re-derivation projects past the ceiling. Tasks
6 and 7 then move to the head of 5c-iii together, since Task 6 adds the
tier links with the routes they point at and Task 7's G7 row needs
both. Task 5's commit links neither, so nothing 404s in between.

**Budget:** 9.4M execution (Ruling 1), flag 7.52M, planning recorded
separately, STATUS at each boundary, spend scored at close against the
ceiling and against zero attended interactions after approval.
