# Rung 5c-iv Task 0: the staleness pre-task, second run

Scope: the plan's Task 0 as amended for its second run (`plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`,
"At its second run, before Task 7"). Items 1, 3, 4, and 5 run as written. Item 2 reads every `file:line` in Tasks 7
through 11 and in the Segment D refresh note. Item 6 confirms VI revision 1.3.

**Verdict: no blocker.** One moved pointer whose content survives (a note). Item 3 was not run, because the
dispatch forbids network access. The conductor owns that check.

## Item 1: lane and tree

- HEAD `61bfdc4` ("docs: checkpoint lane 1 at 5c-iv's launch"). The only commit after the refresh read's `131283c`
  changes `docs/STATUS.md` alone (`git diff --stat 131283c..HEAD`: 1 file).
- The lane registry (`docs/STATUS.md`, "Lane registry") names lane 1 as 5c-iv, segment D-i (Task 0, Task 7), running
  as `wf_9508455c-ca8`. The matcher pass the first-run item named is closed, so no ordering question remains. Lane 2
  runs 6b-ii-b in `.claude/worktrees/lane-2-6b-ii-b` (`git worktree list`).
- Warm changes present, none in lane 1's territory: `tools/photos/sorter.html` (modified; the photo document lane owns
  it), `.claude/agent-memory/dubplate-implementer/` (one modified file and four untracked), and two untracked 2026-09-04
  research files. No file under `web/`, `cmd/`, `internal/`, `Makefile`, or `scripts/` is dirty.

## Item 2: citations in Tasks 7 through 11 and the refresh note

Every pointer below was read with `sed -n` at HEAD. Each one resolves to the content the plan describes, except the
note at the end.

| Citation | Content at HEAD |
| --- | --- |
| `internal/auth/handlers.go:52-62` | 11 outcome constants, `sent` through `signed_out` |
| `handlers.go:157-164` | `type answer struct` with `outcome`, `admin_contact,omitempty`, and `adminContact` |
| `internal/auth/admin.go:35-46` | 12 outcome constants, `added` through `write_unconfirmed` |
| `internal/auth/token.go:118` | `func MintLink(secret []byte, address string, now time.Time, expiry time.Duration) (string, LinkToken, error)` |
| `token.go:172` | `func LinkURL(publicOrigin, prefix, token string) (string, error)` |
| `cmd/dubplate-server/auth.go:113-118` | `magicLinkExpiryThreshold`, `Default: "15m"` |
| `cmd/dubplate-server/serve.go:42-43` | `requestLinkRouteName = "request-link"`, `consumeLinkRouteName = "consume-link"` |
| `cmd/dubplate-server/web.go:50` | `var themeColorRE = regexp.MustCompile(...)`, the pattern the plan quotes |
| `web.go:581` | `case themeColorRE.MatchString(brandNeutral):` |
| `web.go:295-298`, `:485-488` | the two favicon path lists |
| `rendered_test.go:36` | `const renderedScriptTimeout = 300 * time.Second` |
| `rendered_test.go:44-68`, `:47-50` | the four fixture limits (consumptions per client `100`) and `type renderedRow` with `linkRequests` and `consumptions` |
| `rendered_test.go:164-169` | `type loggedRequest` with method, path, query, and headers, no status |
| `rendered_test.go:328-353` | `func checkRateBudget`, summing the manifest's two counts against the limits |
| `web/scripts/rendered/run.ts:77-80` | `PAGES` with `heading: 'Log in'` and `heading: 'Confirm login'` |
| `run.ts:864` | `const MEMBERS_HEADING = 'Members';` |
| `run.ts:1343`, `:1367` | `el.parentElement?.closest('.input, .select, .textarea')` in `focusedInfo` and `settleFocusTransitions` |
| `web/src/lib/styles/components.css:172` | `:where(.btn-outline, .btn-dash) {` (border only, line 173) |
| `components.css:203-211` | the flat-button rule, `box-shadow: none` on `.btn-ghost` |
| `components.css:223` | `:where(.btn-link) {` with `color: var(--ink-primary)` |
| `web/scripts/gates/theme-bundle.ts:352` | inside the `prefers-color-scheme` block check, as the refresh note says |
| `theme-bundle.ts:378-381` | the `theme-controller` refusal |
| `node_modules/daisyui/components/button.css` | `color:var(--btn-rest-fg,var(--btn-color,var(--color-base-content)))`, daisyUI 5.7.46 |
| `node_modules/daisyui/components/swap.css` | `transition-property:transform,rotate,opacity`, `transition-duration:.2s`, `input{appearance:none;border:none}` |
| `lib.dom.d.ts:1204-1206` | `interface GetAnimationsOptions { subtree?: boolean; }` |
| kit `client.js:2573` | `[PAGE_URL_KEY]: page.url.href,` (kit 2.70.3) |
| kit `client.js:90-114` | `if (DEV && BROWSER) {` with the `replaceState` warning |
| rung 5 spec `:1415-1433`, `:1452-1460`, `:1270-1273` | the pages, the deck paragraph, and the post-login destination |
| 5c-i plan `:2495-2497` | the sign-out copy that never claims revocation |

The refresh note's file-level claims also hold:

- `web/src/lib/` holds only `styles/` and `themes/`.
- No `{@html` exists in `web/src`.
- The stubs carry the literal `· dubplate` titles.
- `.gitignore` ignores `/web/build/*` and `/web/node_modules/`, and no `web/.gitignore` exists.
- `png.ts` exports `decodePNG`, `pixelRGBA`, and `pixelHex`, with no encoder.
- `RENDERED_TIMEOUT := 7m` sits at `Makefile:106`.
- cairn's `check-admin-prose.mjs` is 285 lines.
- The Node 24.20.0 errors match those the note quotes: `ERR_UNSUPPORTED_TYPESCRIPT_SYNTAX` for an `enum` and
  `ERR_MODULE_NOT_FOUND` for `$lib/...`.

**Note (moved, content survives).** Task 8's "S20 `:390-396`" (the device page lives in the same web app) pointed at
the streamer spec as of `b9e281b`. At HEAD that paragraph ("### Web surface", "The device page and share management
are pages in the same web app as rung 5's login") sits at `:606-611`. The citation is on a neutral 2.0-hook line and
drives no work.

**Note (mechanics).** Task 7 appends two lines after `root-absolute.ts` "in the existing `; \` chain". That line
(`Makefile:99`) ends the recipe with no trailing `; \`, so the append adds one to it. The recipe runs under
`set -e` (`Makefile:44`), so a failing appended gate fails `make web`.

## Item 3: pins

**Not run.** `npm view`, PyPI, and the `google/fonts` lookups need network access, which this dispatch forbids. The
conductor runs this check before Task 7 if it still wants it. For reference, the local pins in `web/package.json`
are: `@sveltejs/adapter-static` 3.0.10, `@sveltejs/kit` 2.70.3, `@sveltejs/vite-plugin-svelte` 7.3.1,
`@tailwindcss/vite` 4.3.3, `@types/node` 24.13.6, `daisyui` 5.7.46, `playwright` 1.63.0 (installed 1.63.0),
`svelte` 5.57.1, `svelte-check` 4.7.6, `tailwindcss` 4.3.3, `typescript` 6.0.3, and `vite` 8.3.1.

## Item 4: Playwright browsers

Present: `~/.cache/ms-playwright/` holds `chromium-1243` and `chromium_headless_shell-1243`, plus the older `-1234`
pair and `ffmpeg-1011`.

## Item 5: the prefix probe

The probe was re-run against the real bundle, rebuilt with `npm run build --prefix web` (exit 0). A scratch Node
server mounted it under `/p/q`, mapping `/p/q/login`, `/p/q/login/link`, and `/p/q/members` to their `.html`
documents and serving `_app/` assets. Headless Chromium from Playwright 1.63.0 read each page:

- `/p/q/login`: `h1` "Log in", no response of 400 or above, no page error.
- `/p/q/login/link`: `h1` "Confirm login", clean.
- `/p/q/members`: `h1` "Members", clean.
- A client navigation from `/p/q/login` to `members` reached `/p/q/members` with `h1` "Members" and made no
  document request.

The stubs render an empty body under `ssr = false`, so a visible `h1` proves hydration ran. The gate's rendered leg
also covers `login` and `login/link` hydration under the `/rendered` prefix on the real Go shell.

## Item 6: VI revision

The status line at `docs/superpowers/specs/2026-09-24-visual-identity-design.md:3` reads "Status: revision 1.3".
The file's last commit is `45b80cb` ("land the visual identity spec's three errata (revision 1.3)"), so no later
revision exists. Each re-pointed citation reads as the refresh note says:

- `:120-124`: the header halo, `box-shadow: 0 0 0 6px var(--color-base-100)`.
- `:138-141`: the `theme-color` metas, `#1f1b16` and `#110f0d`.
- `:141`: the toggle, with accessible name "Dark theme".
- `:142`: the login page has no toggle.
- `:141-146`: the toggle's cookie, and the ban on `.theme-controller`.
- `:247`: Phosphor bold icons.
- `:268-271`: the login split.
- `:318-319`: the group's name as the heading, and the subline.
- `:340`: curly quotes and apostrophes.
- `:346`: placeholder text at 62%.
- `:347`: header links' hover, and the 44px size.
