# Rung 5c-iii plan review: mechanics and feasibility

Reviewed: `docs/plans/2026-09-25-rung-5c-iii-pages-fonts-rendered-gates.md`,
revision 1 (`fc52232`). Lens: does each mechanism behave as the plan states? Every claim
below comes from a probe run on 2026-09-25 or a quoted source. The probes ran under
`/tmp/claude-1000/.../scratchpad/mech/` and `.../scratchpad/pw/`, against go1.27.1,
Node v24.20.0, `@sveltejs/kit@2.70.3`, `daisyui@5.7.46`, `playwright@1.63.0`, Chromium
153.0.8010.12, and fontTools 4.66.0. Findings are ranked by consequence, not polish.

**Counts:** 0 blockers, 5 majors, 9 minors, 2 owner forks.

The plan has no "could not verify" section. The mechanics it left unmeasured are measured
here: Chromium memory, which binary `launch()` runs, favicon fetching, the font `src` form,
and the conversion's byte comparison. Only the light-lane fit for Tasks 1 and 2 stays
unmeasured.

---

## Majors

### M1. The font `src` cannot be both relative and served from `web/build/fonts/`

- **Location:** plan `:397-402` (R-5ciii1 "Where they are served"), `:546-547` (R-5ciii6's
  gate), `:690-692` (Task 1 mutation 4), `:849-850` (Task 4's seeded `url(/fonts/x.woff2)`).
- **Defect:** R-5ciii1 says the `@font-face` `src` "is written relative" and still lands on the
  copied public file. The probe the plan cites shows the reverse. The public-directory
  rewrite happens only for a **root-absolute** source URL. A relative URL into `static/` gets
  hashed into `_app/immutable/` as a second copy, which is the option the plan declined. So
  Task 1 has two outcomes. It writes `url('/fonts/…')`, which works, and Task 4's
  root-absolute gate then fails the real tree on its CSS `url(/…)` rule. Or it writes a
  relative path, which silently ships the declined design: G5 still passes against the hashed
  copy, and the `fonts` route serves files that no stylesheet references.
- **Evidence:** the planner's own probe source is `src/app.css:20` in `scratchpad/webp`:
  `src: url("/fonts/b.woff2")`. That source becomes `url(../../../fonts/b.woff2)` in
  `_app/immutable/assets/0.*.css`. My probe (`mech/wb2`) changed it to
  `url("../static/fonts/b.woff2")`. The build then emitted `build/_app/immutable/assets/b.Cfh5bZHl.woff2`
  **and** `build/fonts/b.woff2`, and the stylesheet read `url(./b.Cfh5bZHl.woff2)`.
- **Proposed fold:** R-5ciii1 states that the source is root-absolute (`url('/fonts/<name>.woff2')`,
  Vite's public-path form) and that the build rewrites it relative. R-5ciii6's CSS rule then
  checks the **built** stylesheet for a root-absolute `url(`, where one is the real defect,
  and no longer checks `web/src` CSS. The alternative is an allowlist of exactly
  `lib/themes/fonts.css`'s `/fonts/` URLs. Task 1 mutation 4 inverts: a `src` pointing into
  `../static/` must fail G5 or a new G4 row, because a hashed face in `_app/immutable/` means
  the manifest's file is not the file that gets served. Task 4's seeded `url(/fonts/x.woff2)`
  becomes a seed in the built CSS.

### M2. The losslessness proof as written fails on a correct lossless conversion

- **Location:** `:383-384` (R-5ciii1: "yields every table byte-identical to the source
  TTF's"), `:672-673` (Task 1 acceptance), `:684-685` (mutation 1).
- **Defect:** a `--no-glyf-transform` WOFF2 does not round-trip every table byte-identically.
  fontTools always drops `DSIG`, and it always rewrites `head`, setting flags bit 11, which
  also changes `checkSumAdjustment`. As written, the acceptance criterion cannot pass. An
  implementer might then conclude that no lossless path exists, or relax the check ad hoc.
  Separately, mutation 1 holds only for Alfa Slab One. Under the default transform, Work
  Sans's `glyf` and `loca` round-trip byte-identically too.
- **Evidence:** `mech/ft/cmp.py` output, both faces:
  `only in src: ['DSIG']`; `differing tables: ['head']`; `head 54 54 diff byte offsets:
  [8, 9, 10, 11, 16]`; `head.flags src=0x3 rt=0x803`. fontTools
  `ttLib/woff2.py:218-219`: "always drop DSIG table, since the encoding process can invalidate
  it". `:302-305`: "Set bit 11 of 'head' table flags to indicate that the font has undergone a
  lossless modifying transform". Both source `DSIG` tables are 8-byte stubs. Default-transform
  round trip: Alfa differs in `['glyf', 'head', 'loca']`, Work Sans in `['head']` only. The
  `--no-glyf-transform` sizes were measured at Alfa **39,668** bytes and Work Sans **138,000**
  bytes.
- **Proposed fold:** the comparison asserts three things. Every table except `DSIG` and
  `head` is byte-identical. `DSIG` is absent from the output, and its source is the 8-byte
  empty stub. `head` differs only at bytes 8-11 (`checkSumAdjustment`) and in flags bit 11.
  Mutation 1 is scoped to Alfa Slab One. Whether the DSIG drop is still "unchanged except for
  WOFF compression" under OFL FAQ 2.2.1 is owner fork F1.

### M3. Headless Chromium never fetches a favicon, so Task 4's favicon row is vacuous

- **Location:** `:846-848` (Task 4 acceptance: "the favicon request goes to
  `<prefix>/favicon.svg`, never `/favicon.ico` (request log)"), `:859` (mutation 4),
  `:469` and `:489` (`chromium.launch()` and the `executablePath()` refusal).
- **Defect:** `chromium.launch()` in its default headless mode runs `chromium-headless-shell`,
  not Chromium. The headless shell requests no favicon, whether the link is present or absent.
  The row therefore passes on nothing, and mutation 4 can never fail. Under full Chromium's
  headless mode (`channel: 'chromium'`), the icon is fetched and the implicit `/favicon.ico`
  is requested. That implicit request reaches only the **server**, never Playwright's
  `page.on('request')`, so a request log built from Playwright events also misses it. The
  plan's missing-browser refusal has a related gap: `chromium.executablePath()` names the
  full Chromium binary, not the headless shell that `launch()` actually runs.
- **Evidence:** `pw/fav.mjs`. Default launch: server log `["/p/withicon","/p/noicon"]`, with
  no icon request in either case. `{channel:'chromium'}`: server log
  `["/p/withicon","/p/favicon.svg","/p/noicon","/favicon.ico"]`, while `page.on('request')`
  saw `["/p/withicon","/p/favicon.svg","/p/noicon"]`. `playwright-core` `coreBundle.js:43375`:
  `return options.headless ? "chromium-headless-shell" : "chromium";`. `:39662`:
  `executablePath() { return registry.findExecutable(this._name).executablePath() }` (the
  `chromium` entry). `node -e` printed
  `/var/home/glw907/.cache/ms-playwright/chromium-1243/chrome-linux64/chrome`.
- **Proposed fold:** R-5ciii4 launches with `channel: 'chromium'`. That is still Playwright's
  own download (W12.6 holds), and it runs the same binary the user-path paint comes from. The
  measured peak was 827 MB, against 745 MB for the headless shell. The missing-browser
  refusal becomes the `launch()` failure itself, whose message is "Executable doesn't exist
  at …", mapped to the named-install hint. It does not become an `executablePath()` stat. The
  request log that Task 4's rows read is the Go test's own handler wrapper around the shell,
  which sees every request, including one outside the prefix.

### M4. The relative favicon link goes outside the prefix after a client navigation

- **Location:** `:542-546` (R-5ciii6: the `%sveltekit.assets%` favicon is "relative …
  therefore no root `/favicon.ico` request"); Task 4's favicon row.
- **Defect:** Chromium re-resolves `<link rel="icon">` against the new URL after a
  same-document navigation. `login/link.html` ships `href="../favicon.svg"`. After a client
  navigation from `<prefix>/login/link` to a depth-1 route, the icon resolves one level above
  the prefix and is requested there. That request escapes the prefix and falls through to
  whatever the shell mounts outside it, which is the root half of S12. The signed-in path
  plausibly moves from the interstitial to members, and R-5cii19 already relies on a client
  `goto` to members.
- **Evidence:** `pw/fav2.mjs` (full-Chromium headless). The page loaded at `/p/login/link?t=abc`
  with `href="../favicon.svg"`. The native `replaceState` stripped the query, and then
  `pushState('/p/members')` ran. Server log:
  `["/p/login/link?t=abc","/p/favicon.svg","/p/favicon.svg","/favicon.svg"]`.
- **Proposed fold:** Task 4 has the Go head rewrite, which already runs per variant and knows
  the prefix, insert `<link rel="icon" href="<prefix>/favicon.svg" type="image/svg+xml">` as an
  absolute same-origin path. `app.html` carries no icon link. The alternative is rendering
  the link in the root layout's `<svelte:head>` through `asset()`, whose client-side base is
  absolute once computed. Add a rendered row: load `login/link`, client-navigate to members,
  and assert that the server log shows no request outside the prefix.

### M5. The CSP-violation row's named seed never reaches a browser

- **Location:** `:490-493` (W12.1 refusals: "any `securitypolicyviolation` event"),
  `:770-772` (Task 3 acceptance: "a page with a deliberately unhashed inline script (the CSP
  violation row)").
- **Defect:** 5c-ii's `serve.DocumentCSP` refuses startup on an unhashed inline script, and
  `checkVariantCSP` repeats the check per variant. The seeded document therefore fails at
  `newServeShell`, before any page loads. The row "fails", but the violation listener it was
  meant to prove is never exercised: the vacuity class that Review focus 5 names.
- **Evidence:** `internal/serve/csp.go`, `DocumentCSP` doc comment: "Before building it,
  DocumentCSP hashes every inline script in doc and refuses when one is missing from the
  meta's list". `cmd/dubplate-server/web.go`, `checkVariantCSP`. A runtime violation is
  observable. `pw/csp.mjs` served a page with `default-src 'self'; script-src 'self'` and an
  off-origin `<img>`. An `addInitScript` listener recorded `[ 'img-src http://127.0.0.2:9/x.png' ]`,
  and `page.on('console')` saw `error: Loading the image … violates the following Content
  Security Policy`.
- **Proposed fold:** seed a violation that passes startup, such as an off-origin `<img src>` or
  an inline `onclick=` attribute in a scratch build, and assert that the row fails naming the
  directive. Keep the unhashed-script seed as a separate, correctly labeled row: "startup
  refuses".

---

## Minors

1. **The Content-Type mutation cannot be killed as described** (`:686-688`). With the header
   dropped, `ServeContent` still answers `font/woff2` two ways: from the extension
   (`/etc/mime.types:font/woff2 woff2`, `/usr/share/mime/globs2:50`) and from the sniffer
   (`net/http/internal/sniff.go:178`). A table that maps the extension to something wrong
   beats the sniffer. Probe `mech/mime`: `mime.AddExtensionType(".woff2",
   "application/x-seeded")` changed an unset header to `application/x-seeded`. **Fold:** the
   test seeds a wrong host mapping with `mime.AddExtensionType` and asserts `font/woff2`. The
   quote of Go's table itself is accurate: `mime/type.go:116-120` ends its `.w` entries at
   `.webp`.
2. **The first-paint cookie states need the right `addCookies` form** (`:583-587`). JS-disabled
   contexts cannot write `document.cookie`. `context.addCookies` with a `url` field throws for
   a `__Host-` name (`Protocol error (Storage.setCookies): Invalid cookie fields`, both with
   and without `secure`). `{domain:'127.0.0.1', path:'/', secure:true}` works, and the server
   saw `cookie=[__Host-dubplate_theme=label]` (`pw/ac.mjs`). **Fold:** name that form in
   R-5ciii8.
3. **The swap toggle's focus ring surrounds a 7×12 px invisible input** (Task 9, `:1079-1088`).
   DaisyUI's `.swap input{appearance:none;border:none}` leaves the checkbox focusable, and
   Space toggles it. However, `:focus-visible` paints on the input's own box (measured 7.2×12
   px), not on the 44 px control. Task 9 carries no focus row, and Task 5's rows run only on
   the stubs. **Fold:** Task 9 moves the ring to the control
   (`.swap:has(input:focus-visible)`), suppresses it on the input, and adds a toggle ring
   row. Probe: `pw/swap.mjs`.
4. **The `version.json` mismatch seed needs a named mechanism** (`:770`). The test embeds the
   same `web/build` it compares against, so editing the file changes both sides. The seed
   works only by building the test binary (`go test -c -tags rendered`), editing
   `web/build/_app/version.json`, and then running the binary. **Fold:** say so.
5. **A Chromium launch is not the gate's memory risk** (`:494-495`). The measured peak for ten
   full-page screenshots, 320 to 2560 wide in both schemes, was 745 MB with the headless shell
   and 827 MB with `channel:'chromium'`, under a `MemoryMax=3G` unit
   (`systemd-run --wait` "Memory peak"). The 7G/8G heavy lane has room. **Unmeasured:**
   whether Tasks 1 and 2's full gate (vite build, svelte-check, `go test ./...`,
   golangci-lint) fits the light lane's 3G cap. Tasks 1 and 2 should report their peaks, or
   run heavy.
6. **The process lifecycle is sound, and one detail is worth stating.** In the probe
   (`mech/goexec`), `exec.CommandContext` killing `node` at an 8s deadline returned `signal:
   killed` at 8.02 s, and no Chromium process survived under either binary: Chromium exits
   when its pipe to Node closes. `go test`'s default 10-minute timeout will bind at Task
   10's matrix. **Fold:** the `web-rendered` target passes an explicit `-timeout`.
7. **`--scope-changed` never runs `make check`** (`:483-486`). `scripts/check.sh:295-308`
   runs `make vet lint tidy-check` and a hand-listed `go test` in that form. Hanging the leg
   off `make check` therefore does not reach it, and Task 3 must wire it into that branch
   explicitly. The plan implies this. It should say so, so a diff-reviewer checks both
   branches.
8. **Branding a non-hex `--color-neutral` would newly refuse startup** (`:339`, `:521-524`).
   Under 5c-ii, `DUBPLATE_BRANDING: --color-neutral: oklch(...)` is accepted. Under this plan,
   the hex grammar refuses it. That is a behavior change to an operator key, not only a
   hardening. See owner fork F2.
9. **The site name has no first paint** (`:527-528`). It is read "synchronously at mount", so
   the prerendered `h1` is empty until hydration, and empty under JS-disabled first-paint
   contexts. That is harmless for R-5ciii8's background rows, but no row covers it. The
   rendered identity check ("its `h1` text") must wait for hydration, not
   `domcontentloaded`. **Fold:** state it in R-5ciii4's un-booted-page refusal.

---

## Owner forks

**F1. Does a WOFF2 that drops the 8-byte `DSIG` stub and sets `head` bit 11 count as
"unchanged except for WOFF compression" (OFL FAQ 2.2.1) for Alfa Slab One's Reserved Font
Name?**

- **Option A:** yes. The WOFF2 encoder, not the converter, makes both edits, and the dropped
  `DSIG` carries no signature (8 bytes).
- **Option B:** no. Ship Alfa Slab One under a renamed family, which the OFL requires for a
  modified RFN font.
- **Option C:** take a WOFF2 the upstream project itself publishes, if one exists.

Recommendation: **A**, recorded in `FONTS.md` with the M2 comparison output.

**F2. What happens when `DUBPLATE_BRANDING` sets a non-hex `--color-neutral`?**

- **Option A:** startup refuses, as the plan states. This breaks a config that 5c-ii accepts.
- **Option B:** the head carries no `theme-color` meta, and startup logs one line.
- **Option C:** the head keeps the bundle's own per-theme values and ignores the branding
  neutral for the meta.

Recommendation: **B**. It fails soft on a cosmetic attribute, keeps every prior config
booting, and no operator value reaches an attribute outside the grammar.

---

## Checked and sound (do not re-litigate)

- **Go's MIME table.** It has no built-in `.woff2`, and the fallback reads host tables. `go
  doc`-level quotes hold (`mime/type.go:116-120`; `type_unix.go:21-31`; the sniff signature
  at `sniff.go:178`). A strong ETag with a zero modtime answers 304 to a matching
  `If-None-Match` (probe `mech/mime`).
- **The `__Host-` cookie over loopback http.** It works when set through `document.cookie`,
  and so does the server's `Secure` session `Set-Cookie` (`pw/probe.mjs`, `pw/sc.mjs`).
  Dropping `Secure` or `Path=/` leaves `document.cookie` empty, so Task 9 mutation 2 holds.
- **Playwright `colorScheme`.** The quotes hold (`types.d.ts:11331-11336`, `:2786`).
  `playwright@1.63.0` has no install scripts (`scripts {}`), so `npm ci` downloads no browser.
  Both `chromium-1243` and `chromium_headless_shell-1243` are present.
- **Kit's `replaceState`.** It stores `[PAGE_URL_KEY]: page.url.href`
  (`client.js:2571-2577`). The native-history warning sits inside `if (DEV && BROWSER)`
  (`:87`) and is a `console.warn`. Kit's other history writes (`:380`, `:2949`, `:3257`) pass
  no URL or pass the current `location`. The skip link's hash click compares against
  `location` (`:2745`), so it does not re-navigate to the stripped URL.
- **CSS Cascade 5 §6.4.3.** "Within each layer, nested layers are sorted in appearance order,
  and style rules without further nesting are similarly added to an implicit sub-layer after
  the explicitly nested layers." A normal declaration whose layer comes last wins: "for
  normal rules the declaration whose cascade layer is last wins". A rule placed directly in
  `@layer utilities` therefore outranks `utilities.daisyui.l1.l2.l3`, and `@layer components`
  loses to it (Task 5 mutation 1 holds). The built bundle's layer order is `properties,
  theme, base, components, utilities{daisyui.l1.l2.l3 …}`.
- **The built `--color-neutral`.** It is emitted as `#1f1b16` and `#110f0d`, twice each (the
  composed theme and the `system.css` branch). The built-ins emit `oklch(14% .005 285.823)`,
  so the reader must scope to the `dubplate-*` blocks.
- **Reproducible builds.** With `kit.version.name` unset, two builds differ
  (`e56a08…` against `f665c6…`). With it fixed, they match (`a31d77…` twice), so R-5ciii7
  holds.
- **First paint.** A JS-disabled screenshot taken at `domcontentloaded`, against a stylesheet
  delayed by 1.5 s, returned the post-CSS pixel `[17,15,13]` at 1583 ms. Chromium paints no
  frame before render-blocking CSS, and `page.evaluate` still works in a JS-disabled context.
- **G10.** Its discriminator is real: `dubplate` `--radius-box: 0.25rem` against built-in
  `light`'s `.5rem`.
- **`go:embed all:build`.** It accepts `WorkSans[wght].woff2`.
- **Tooling.** `golangci-lint` 2.14.0 has `--build-tags`, and `rsvg-convert` 2.63.2 is
  installed.
- **Kit's `asset()`/`resolve()`.** Both prefix the runtime base, which the bootstrap computes
  as `new URL(".", location)` or `new URL("..", location)`.
- **The mark SVGs.** They carry no `<text>`, font, or external reference.
- **IBM Plex Mono.** `@ibm/plex-mono@2.5.0` declares `OFL-1.1`. The file is 49,248 bytes,
  version 2.005, and IBM ships it already as WOFF2.
