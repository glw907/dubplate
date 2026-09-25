# Draft HISTORY entry: rung 5, pass 5c-ii

Draft only. The conductor folds this into `docs/HISTORY.md` and does not commit it as a
standalone file.

## Rung 5, pass 5c-ii (the admin gate, the roster API, and the `web/` tree born gated): closed 2026-09-25, lane 1

Executed against `docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`
(revision 2.4, ratified by Geoff 2026-09-24, revision 2.2's seven rulings all yes, with revisions
2.3 and 2.4 amending the theme surface for Geoff's visual identity ruling), in the main tree, from
`6de96a7` (code pointers resolved). Task 0's report is not separately filed; no blocker surfaced.

### What landed

Task 1 (`5637ecb`, fix round `7769d25`) added the admin gate and the roster API in `internal/auth`:
`admin.go`, the `RequireAdmin` wrapper, `NewRosterAPI`, and the outcome table for GET, HEAD, and
POST, none of which ever takes the `admin` flag from a request.

Task 2 (`e9f7367`) mounted the roster API in `internal/serve` and closed the n5 address refusal in
`internal/identity`: the derived `Default_Ignorable_Code_Point` set (4174 code points, Unicode
17.0.0) plus U+2800, refused in `NormalizeAddress` so every channel (login, roster load,
`AddMember`) shares one refusal.

Task 3 (`66f6e14`, fix round `493869b`) bore `web/`: the SvelteKit static-mode tree at the pinned
dependency versions, the three route stubs (login, login/link, members), and the gate leg in the
same commit (`scripts/check.sh`, the Makefile, `CLAUDE.md`'s gate paragraph, `go.mod`'s one
`ignore` directive, `.gitignore`, `.gitleaks.toml`).

Task 4 (`4092ec0`, fix round `4c544a2`) landed the theme as data: the `dubplate` match block and
the `label` and `label-night` palettes from the visual identity draft tokens (Ruling 3's
amendment), plus the `dubplate-system` value.

Task 5 (`4489716`; widened `d9d0eca`; fix round `b0e01d6`; fix `bdc41fb`) added the embed path, the
theme stamp, and the theme keys: `DUBPLATE_THEME_MATCH` and `DUBPLATE_THEME_PALETTE` validated at
serve startup, the `__Host-dubplate_theme` cookie reader, and the proof that no log record carries
the cookie value.

Task 6 (`f65d4f6`) added the branding stylesheet, the operator CSS file, and the tier links.

Task 7 (`8aa5fdc`; simplified `2424b9a`) added the served header set, the CSP built per document
from SvelteKit's own hashes, the cache directives, and the members document route.

### The pass-end reviews and their fold

Three reviewers ran over `0c2745f..2424b9a` and committed at `71de914`:
`svelte-reviewer` (`docs/superpowers/research/2026-09-24-rung-5c-ii-review-svelte.md`),
`web-auth-security-reviewer` (`-review-security.md`), and `daisyui-a11y-reviewer`
(`-review-a11y.md`).

The fold ran in two rounds. Fold A closed the gate-script findings across two commits
(`d236d03`, then a second `fix` round `cc4649c`): Svelte B1 (the vacuous-gate bug under a path
containing a space or a percent-encoded character), W1 through W3, W5 through W7; a11y warnings
1, 2, 3, and 6; security S3 and S6. A reviewer second-fixed the round, and the conductor ruled one
more targeted round, F-A3 (`1306ae3`): `checkNameConfinement` now passes the enclosing layer name,
not a boolean, so every rule setting a theme token name is confined to `@layer base` under any
wrapper (`:host`, `:where(:root)`, a stock `[data-theme=light|dark]` block, or the theme-controller
branch), except Tailwind's own preflight rule; and `STOCK_THEME_CONTROLLER`'s alternation-precedence
bug is fixed. F-A3 was accepted.

Fold B (`3acbdc0`, `opus`) closed: security M1 (the operator CSS file, previously read through
`os.Stat` then a separate open, which followed a symlink and let a `roster.json` symlink named
`x.css` pass every check; the fix opens the file once with `O_RDONLY|O_NOFOLLOW`, `fstat`s the
descriptor, resolves the parent with `filepath.EvalSymlinks`, and refuses a path matching the
roster or anything under `StateDir`, with symlink, parent-symlink, and FIFO test rows); L1 (bounded
method logging); L2 (`isLoopbackHost` widened to `LOCALHOST`, `localhost.`, and `*.localhost` per
RFC 6761); L3 (CSP recomputed per served variant); L4 (the members document moved to `no-store`);
Svelte W4 (`img-src 'self' data:` added to both the kit config and `internal/serve/csp.go`, which
resolves the handoff's open conflict between `default-src` and `img-src`: both now name `data:`
explicitly, so DaisyUI's noise texture and Vite's inlined assets are no longer blocked); W8 (a Go
test pinning the meta and header CSP directives equal); a11y W7 (a `<title>` on every route and the
error page); the identity spec's label ink values (`--ink-primary` `#9d2f26`, `--ink-warning`
`#754f00`); and the corrected `(0,4,1)` specificity comment in `system.css` (was claimed `(0,1,0)`).

`code-simplifier` ran once over the fold (`1a7cc9e`): one stale `:host` doc comment on
`isConfinedPart` corrected.

### What stayed open, and where it goes

Security M2 (the roster API shares an origin with the Navidrome fallthrough: once
`SetFallthrough` is called, any script on that origin could read the roster or POST members with
an admin's session) is not live today and carries to rung 12 as a question for Geoff: a separate
origin or step-up confirmation for roster writes, until 2.0 retires Navidrome.

Security L5 (no per-actor throttle on the roster POST) and L6 (n5 can refuse a pre-existing roster
whole; a doctor row is owed) carry with no assigned pass yet; both are candidates for the same 5d
runbook pass that already carries the other security residuals below.

M1's race guards (the symlink-swap fix) have no mutation test: a between-syscalls swap of the
target is not deterministically testable, so the fix is proved by code reading, not by a red test
turned green. Recorded as a declared gap, not a defect.

A11y warnings 4, 5, and 6 stay open. Warning 4 (G9 measures against base-100 only; several
direct-text and fill colors fail on base-200 or base-300) has no per-surface gate row yet; it is
carried to 5c-iii, which the visual identity spec's own contrast table (built from the 5c-ii
review's numbers) already narrows. The reviewer's WCAG 1.4.11 question, whether a `label-night`
primary fill's own contrast identifies the control or the button's label text does, is also
carried to 5c-iii. Warning 5 (fill-only colors as invisible focus rings) and warning 6 (inks do not
follow DaisyUI's `theme-controller` branch) are answered by the visual identity spec's ruled focus
rule (one 2px `--ink-primary` outline everywhere, never a `--btn-color` outline) and its ban on the
`.theme-controller` pattern; 5c-iii builds to that rule.

Svelte S11 (builds are not reproducible; `kit.version.name` is a timestamp, so the CSP hash changes
per build) and S12 (a 5c-iii note on absolute hrefs escaping the prefix, and the root favicon
request) carry to 5c-iii.

Security I2 (the login/link page must consume the token only on a click, drop `?t=` with
`history.replaceState`, and load no third party) carries to 5c-iii, which is the pass that first
builds that page.

Everything else in the three reviews (every remaining suggestion, and the security review's info
items I1, I3, I4, I5) was read and left as reported: none blocks, and none is assigned a
destination beyond its own review file.

### The gate evidence

`bash scripts/check.sh` (`cairn-run-gate`) at HEAD (`1a7cc9e`): **exit 0**, 226-line log. The web
leg's build step reports `vite build` at 616ms (client) and 6.18s (server), and `svelte-check`
completing in about 5ms of its own reported wall time (`1790346906233` to `1790346906238`); the
static gates (`server-files`, `g1`, `g3`, `g2`, `g9`, `theme-bundle`) each report "passed" with no
separately timed wall clock. `go vet` and `golangci-lint run ./...` report 0 issues. gitleaks
scanned about 115.10 MB in 53.8s and found no leaks.

`go test -count=1 ./...` at the same commit, cache cleared by `-count=1`: **2m25.197s** wall
(`user 3m27.728s`, `sys 0m13.866s`). `internal/harness/conform` dominates at 142.257s; the next
largest are `internal/harness` at 10.998s and `internal/backup` at 9.848s (6b-i's package, merged
alongside). The five packages this pass's plan names: `cmd/dubplate-server` **52.351s**,
`internal/auth` **6.221s**, `internal/serve` **1.175s**, `internal/identity` **0.079s**, `web`
**0.004s** (both figures from the same `-count=1` run; a second isolated run over just these five
packages agreed within noise: `cmd/dubplate-server` 52.236s, `internal/auth` 6.334s, `internal/serve`
1.145s, `internal/identity` 0.060s, `web` 0.006s).

**Could not verify:** the build's peak memory. No task recorded a measurement, and no gate step
instruments it. `CAIRN_GATE_LANE=light`'s 3G cap was never at issue: no gate step in 5c-ii launches
a browser, per the plan's own rule, so the light-lane question the plan asked Task 3 to answer
did not need answering with a number.

The two staged legs report unchanged from 5c-i in kind: 326 process citations in Go comments
(`CHECK_CITATIONS=warn`) and 128 exported-unused `internal/` identifiers (`CHECK_EXPORTED_UNUSED=warn`),
neither gating.

### The resolutions (R-5cii1 through R-5cii22)

| Resolution | What it settles | Where it landed |
|---|---|---|
| R-5cii1 | The remaining 5c scope splits again, after Task 7, into 5c-ii and 5c-iii | This pass's own scope (Ruling 1) |
| R-5cii2 | The runtime prefix rides SvelteKit's relative base; no `reroute` hook, no hash router | Task 3 (`web/svelte.config.js`), proved by Task 5's reference half |
| R-5cii3 | The gate's refusal stays `serve.Terminator()` / `http.NotFound`; no served 404 document exists to track | Task 1, Task 2 (`internal/serve/prefix.go`) |
| R-5cii4 | The roster API's outcome table, extended for 5c-i's `errDuplicatePrincipal`, `errBlankDisplayName`, `errRosterWritableByOthers`, `errRosterWriteLanded`; `r.PostForm` only, any query key refuses | Task 1 (`internal/auth/roster.go`, `admin.go`) |
| R-5cii5 | n5 closes in the normalizer from stdlib Unicode tables, not a missing derived table | Task 2 (`internal/identity/identity.go`) |
| R-5cii6 | The header set is the shell's outermost step; HSTS follows `DUBPLATE_PUBLIC_URL`'s scheme, `max-age=31536000`, no `includeSubDomains`, no `preload` (Ruling 2) | Task 7 |
| R-5cii7 | The CSP is built per document from SvelteKit's own generated hashes | Task 7 (`internal/serve/csp.go`) |
| R-5cii8 | Cache directives are derived per route class, not copied from the salvage handler | Task 7 |
| R-5cii9 | The embed lives in package `web`; a build re-emits its placeholder | Task 3, Task 5 |
| R-5cii10 | The bundle builds before any Go test reads it; `go.mod`, `.gitignore`, `.gitleaks.toml` learn about `web/` | Task 3 |
| R-5cii11 | The install-state check covers three measured states (absent, stale, current) | Task 3 |
| R-5cii12 | The theme keys are thresholds with no validation hook in `internal/config`; serve's own resolution refuses a rejected value, and `doctor` reports without judging | Task 5 |
| R-5cii13 | The operator CSS file is read once at startup, bounded, served at one fixed path | Task 6, hardened by Fold B's M1 fix |
| R-5cii14 | Branding is a `;`-separated declaration list with a closed grammar; no request-supplied value reaches CSS | Task 6 |
| R-5cii15 | G4, G5, G7's computed-value half, and G10 land in 5c-iii, with the fonts and the rendered tooling | Deferred to 5c-iii by name |
| R-5cii16 | The web dependency pins, each the newest production version on 2026-09-24; `@types/node` joins under Ruling 5 | Task 3 |
| R-5cii17 | The review shape: a second `web-auth-security-reviewer` read at this pass's own close, scoped to its own diff | This pass's reviews (`71de914`) |
| R-5cii18 | Nothing is exported ahead of a cross-package caller | Applied throughout; confirmed by the security review's exported-surface check |
| R-5cii19 | `<prefix>/members` composes the admin gate, the GET-and-HEAD method check, then the document; the route's existence is public, the API is the boundary | Task 7 |
| R-5cii20 | Every served response carries `nosniff` and `no-referrer`; only a served document carries the CSP | Task 7 |
| R-5cii21 | Every route answers its exact paths for GET and HEAD only; no directory listing | Tasks 5, 6, 7 |
| R-5cii22 | The default follows the system preference in CSS; a cookie picks among stamped variants; the Go reader lands here, the toggle lands in 5c-iii | Task 4 (data), Task 5 (cookie reader and stamp) |

### The seven rulings and their outcomes

1. **The split, 9.4M ceiling for 5c-ii.** Ruled yes. Built as written; the spec's 5c row and the
   web surface standard's section 15 take a second split erratum (below).
2. **HSTS `max-age=31536000`, no `includeSubDomains`, no `preload`.** Ruled yes. Built as R-5cii6.
3. **Task 4 authors the theme now, without a mockup session.** Ruled yes, then amended: Geoff's
   live visual identity ruling (session dubplate-04, 2026-09-24) replaced `gruvbox` with `label`
   and `label-night` before Task 4 dispatched. Task 4 built the amended values.
4. **`DUBPLATE_BRANDING` and `DUBPLATE_OPERATOR_CSS_FILE` default to the literal `none`.** Ruled
   yes. Built as R-5cii13, R-5cii14.
5. **`@types/node` joins the web pins at an exact 24.x version.** Ruled yes. Built as R-5cii16.
6. **`Referrer-Policy: no-referrer` stays, recording the old-Safari residual.** Ruled yes. The
   residual (a browser sending no `Sec-Fetch-Site` could fail a same-origin POST under
   `no-referrer`'s `Origin: null` serialization) is carried to 5d's runbook row.
7. **A branding value outside the grammar, or a missing or refused operator file, stops serve
   startup.** Ruled yes. Built as R-5cii13, R-5cii14; carried to rung 12's plan, where the same
   refusal would also stop the Navidrome fallthrough.

### Declared coverage gaps

At minimum, carried forward from this pass's own plan and the reviews:

- The theme keys' validation lives in serve's resolution, and `doctor` does not judge them
  (R-5cii12), to 5d.
- The prefix proof's hydration half is evidence from a scratch probe, not a gate; to 5c-iii.
- G4, G5, G7's computed half, and G10 are not yet gates; to 5c-iii (R-5cii15).
- The n5 set is pinned to Unicode 17.0.0; a later Go toolchain's Unicode version can change it
  (R-5cii5's own "the table is the toolchain's").
- The header step wraps the fallthrough; rung 12 must reconcile the header set with Navidrome once
  mounted.
- No ESLint comment config for TypeScript yet; to rung 7's W12 static rules.
- The first-paint agreement between the theme stamp and the CSS is proved statically at Task 5;
  the rendered half is 5c-iii's (R-5cii22).
- Rung 12 must re-validate every existing roster row's `streamer_username` principal, not only new
  writes, since the API writes it from 5c-ii on.
- The shared-origin residual (security M2) is a rung 12 threat-model entry.
- Ruling 7's availability question goes to rung 12's plan.
- Ruling 6's old-browser login residual goes to 5d's runbook row.
- The note that a branding color applies in both color schemes (identity spec amendment A4) goes
  to 5d's runbook row.
- To 5d: `deploy.sh` must build the bundle from a clean committed tree (`npm ci`, then the build)
  before `go build`, and must refuse a dirty `web/`, since the binary embeds whatever `web/build`
  holds.
- Security M1's race guards, L5, L6 (above, "What stayed open").
- A11y warnings 4-6, the WCAG 1.4.11 question, Svelte S11, S12, security I2 (above, "What stayed
  open").

### Both budgets

**Tokens.** Ceiling 9.4M, 80% flag at 7.52M. Spend was about 7.4M through the three pass-end
reviews (`71de914`); Geoff approved continuing past the flag (2026-09-25). Beyond that: the first
fold run (`wf_02fe1440-91c`, Fold A across `d236d03` and `cc4649c`, about 0.6M), the second fold run
(`wf_78a4751c-f0d`, F-A3 and F-B across `1306ae3` and `3acbdc0`, about 0.58M), and the simplifier
round (`1a7cc9e`, about 0.11M). That puts spend at roughly 8.7M before this close task, all figures
approximate. Task 8's own cost is not yet known; the conductor's fold adds it.

**Attended time.** One planning question at plan approval (the seven rulings, all yes on first
ask), one live design session for the visual identity ruling that amended Ruling 3 mid-flight, and
one combined question at the 80% flag (continue past it: yes). No task-level check-in.

### What a later pass would be wrong to rediscover

- The runtime prefix problem R-5c-i's outline called "the pass's largest unknown" has one answer,
  proved against the kit source and a headless-browser probe: `paths.relative` at its default with
  no `reroute` hook and no fallback page makes every prerendered document self-locating from
  wherever it is served, at the cost that every route must answer its exact path (a trailing slash
  answers the terminator's 404, not the document).
- `img-src` needs `data:` explicitly in both the kit CSP config and the Go CSP builder; DaisyUI's
  noise texture and Vite's asset inlining both emit `data:` URLs that `default-src 'self'` alone
  blocks. Fixed once, at Fold B; a later pass adding a component with an inline image should not
  re-diagnose this.
- The theme-bundle gate scripts must confine a theme-token-setting rule to `@layer base` under
  every wrapper a browser recognizes (`:host`, `:where(:root)`, a stock `[data-theme]` block, the
  theme-controller branch), not just a bare `:root`. F-A3's fix generalizes this; a later gate
  script touching `theme-bundle.ts` should extend the confinement check, not narrow it back to one
  selector shape.
- The operator CSS file's read path is a security-sensitive seam even though it looks like static
  file serving: `os.Stat` follows symlinks, so a stat-then-open pattern with a suffix check on the
  configured path (not the resolved path) can be defeated by a symlink named `x.css` pointing at
  the roster file. The fix (open with `O_NOFOLLOW`, `fstat` the descriptor, `EvalSymlinks` the
  parent, refuse a path matching the roster or `StateDir`) is the pattern any future
  operator-supplied-file feature should copy, not reinvent.
- Every gate script's `main()` guard must not compare a percent-encoded `import.meta.url` against
  a raw `process.argv[1]` path; a repo path containing a space or a percent-encoded character makes
  every such gate pass silently by never running. Fold A's fix is the pattern for any future gate
  script.

## Errata owed

Draft text, paste-ready into each named document. The conductor applies these as erratum blocks,
not as edits to the ratified prose above them.

**1. The rung 5 spec's 5c row and pass split (the second split).** Append after the existing
2026-09-24 erratum block at `docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md:2438-2472`:

> Erratum, 2026-09-25 (pass 5c-ii close): **the 5c-ii row splits again, into 5c-ii and 5c-iii.**
> The erratum above described 5c-ii as carrying "the pages, the theme, the embed path, and the
> gate's `web/` leg rejoin" under its own plan's ceiling. `docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`'s
> Ruling 1 splits that scope again at the served-header task's boundary: 5c-ii carries the
> Go-side and delivery work (the admin gate, the roster API, the `web/` tree's birth and gate leg,
> the theme as data, the embed path, the branding and operator stylesheets, the served header
> set) and closed under a 9.4M ceiling at about 8.7M in subagent tokens (this pass's own HISTORY
> entry has the figure). 5c-iii carries the fonts, the three pages with their string deck, the
> rendered gates, and the render-and-read, under its own plan's ceiling. 5c's three-pass total
> reservation moves from about 18M (5c-i's Ruling 6) to about 22M; rung 5's own reservation moves
> from about 37M to about 41M.

**2. The web surface standard's section 15 line.** Append to the existing 2026-09-24 erratum
block at `docs/superpowers/specs/2026-09-12-web-surface-standard.md:1748` (the "5c ceiling
unchanged" erratum):

> Erratum, 2026-09-25 (pass 5c-ii close): **the second split.** 5c-ii itself split into 5c-ii and
> 5c-iii at the served-header task's boundary (rung 5 spec's matching erratum). "5c-ii's ceiling
> is its own plan's" now names two plans, not one: 5c-ii closed under 9.4M, and 5c-iii carries its
> own re-derived ceiling.

**3. The rung 5 spec's "failing config load" row (R-5cii12).** In the 5c pass-split table's 5c row
(`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md`, "The pages, the `member` verb, and
the gate leg" row), the exit-gate clause "a rejected theme match value failing config load with
the accepted set named" reads as an erratum:

> Erratum, 2026-09-25 (pass 5c-ii close): **the theme match value is refused by serve's own
> resolution, not by `internal/config`.** `internal/config`'s threshold registry has no validation
> hook (R-5c17 keeps that package untouched across 5c-i and 5c-ii), so the check that a rejected
> `DUBPLATE_THEME_MATCH` or `DUBPLATE_THEME_PALETTE` value fails, naming the accepted set, lives in
> `cmd/dubplate-server/web.go`'s serve-startup resolution (R-5cii12). `doctor` prints the
> configured value without judging it; that gap is carried to 5d.

**4. The rung 5 spec's embed-path sentence on the CSP (R-5cii20).** At
`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md:1393-1394`:

> Erratum, 2026-09-25 (pass 5c-ii close): **"Every served response carries `X-Content-Type-Options:
> nosniff` and `Referrer-Policy: no-referrer` beside the CSP" reads as two clauses, not one.**
> Every served response (a 404, a JSON answer, an asset, a stylesheet) carries the first two
> headers. Only a served document carries the CSP, because only a document runs script, and the
> gate's 404 must equal the terminator's answer exactly, which carries no CSP (R-5cii20).

**5. R-5cii3's withdrawal of the 5c-i outline's 404-tracking line.** The 5c-i outline's line that
the gate's refusal "must track the served 404 document once W3.7's 404 lands" is withdrawn: W3.7
gives every owned-path 404 to the Go mux, and no served 404 document exists for the gate to track.
(No separate document carries this line to correct; recorded here as the disposition.)

**6. DL R3.6's "The estate's is `dubplate-gruvbox`" and DL section 3's `palette gruvbox`
(R-5cii22).** At `docs/superpowers/specs/2026-09-07-design-language-standard.md` (the "two-axis
composition model" heading, and R3.6):

> Erratum, 2026-09-25 (pass 5c-ii close): **the estate's composed identities are `dubplate-label`
> and `dubplate-label-night`, not `dubplate-gruvbox`.** Geoff's visual identity ruling
> (`docs/superpowers/specs/2026-09-24-visual-identity-design.md`) replaced the `gruvbox` palette
> with `label` (light) and `label-night` (dark) before Task 4 of pass 5c-ii dispatched. Every
> place this document names `dubplate-gruvbox` or `palette gruvbox` as the estate's configuration
> reads as the two composed identities above; the config keys and their grammar are unchanged.

**7. W5.10's "this clause lands at rung 7" (R-5cii22).** At
`docs/superpowers/specs/2026-09-12-web-surface-standard.md` (W5.10):

> Erratum, 2026-09-25 (pass 5c-ii close): **the cookie reader and the toggle split across two
> passes.** W5.10 as ratified describes the theme cookie mechanism as one piece of rung 7's work.
> Pass 5c-ii's R-5cii22 lands the Go-side cookie reader and the theme stamp resolution now
> (`cmd/dubplate-server/web.go`, Task 5); the toggle that writes the cookie (the DaisyUI `swap`,
> the accessible name "Dark theme") lands in 5c-iii.

**8. A clarifying erratum to W5.9 (identity spec amendment A11).** At
`docs/superpowers/specs/2026-09-12-web-surface-standard.md` (W5.9, the sentence "A system-following
pick is a third value the Go layer resolves and stamps"):

> Erratum, 2026-09-25 (pass 5c-ii close): **the system-following pick is stamped by the Go layer
> and then resolved by a media block, not resolved by the Go layer alone.** The Go layer stamps the
> single value `dubplate-system`; a CSS block under `@media (prefers-color-scheme: dark)`, keyed to
> that stamped value at `(0,1,0)` inside `@layer base`, resolves it to `label` or `label-night`
> variables in the browser. W5.9's two reasons against a media block (dead code, since DL R4.5
> stamps every document; and a specificity that would outrank the branding stylesheet at any link
> order) do not apply to a block keyed to a stamped value rather than an unconditional
> `:root:not([data-theme])` guard: it is not dead code, because `dubplate-system` is a real stamped
> value, not the guard's "no attribute at all" case, and at `(0,1,0)` it does not outrank branding
> at any link order.

**9. The rung 5 spec's staleness item citing `CLAUDE.md:118`.** No new erratum text is owed here
beyond confirming the existing staleness pre-task rows
(`docs/superpowers/specs/2026-09-08-rung-5-identity-auth.md:2626`, `CLAUDE.md:118`,
`scripts/check.sh:5`, `Makefile:33-34`) still describe the gate leg correctly after Task 3's rejoin;
Task 3's own commit rewrote all three, so this item is closed, not carried.

**10. The HISTORY 5c-i entry's n5 sentence, corrected by R-5cii5.** In `docs/HISTORY.md`'s 5c-i
entry (top entry as of this pass's start), wherever it states that n5 was refused because
"refusing it needs the derived Unicode Default_Ignorable_Code_Point table, which the standard
`unicode` package does not expose": that reasoning is now superseded. `go doc unicode` exposes
every input the derivation needs (`Other_Default_Ignorable_Code_Point`, `Variation_Selector`,
`Prepended_Concatenation_Mark`, `White_Space`, `Cf`), and R-5cii5 derives the 4174-code-point set
from them, matching `DerivedCoreProperties.txt` (Unicode 17.0.0) exactly. The conductor amends the
5c-i entry's sentence in place rather than adding a new erratum block, since HISTORY is not a
ratified spec.

**11. Web surface standard section 19 entry: the identity spec withdraws W1.3's illustrations.**
Add a row to `docs/superpowers/specs/2026-09-12-web-surface-standard.md`'s section 19 table:

> | W1.3's vocabulary illustrations ("Nothing cut yet", "Pressed and filed") | Geoff's visual
> identity ruling 4 withdraws them: the identity lives in the visuals and the photographs, and
> copy carries no record-shop puns | W1.3 |

## The 5c-iii handoff

See the STATUS draft's "5c-iii handoff" pointer. The full handoff (the surface 5c-iii builds on,
the visual identity spec as standing input, the theme work R-5cii22 routes forward, and the
carried review items) is written out in this file's "What stayed open" and "Declared coverage
gaps" sections above; nothing further is owed beyond what is already there. In one place: 5c-iii
builds on the routes, the gate, the roster API's outcome values, the header step, and the CSP
builder that hashes whatever pages it adds; it reads the visual identity spec revision 1.2
(`782d864`) as its standing input, with the draft tokens file as its superseded predecessor; and
it carries the theme toggle, the `theme-color` metas, the mark and favicon, the ink utilities, the
rendered first-paint proof, the four ruled font families (Alfa Slab One, Work Sans, IBM Plex Mono,
all OFL, as woff2), the component CSS the spec names outside DaisyUI's variables (the hard offset
shadow, the sleeve grain, the record rim, the no-art die-cut sleeve), the branding-flattens-both-
palettes rule (A4), and the component sheet under review with Geoff.
