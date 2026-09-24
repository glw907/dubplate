# Rung 5c-ii plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`
at `3ebfe3c` (revision 1, draft). Lens: does every mechanism behave as
the plan states, and can each task's gate go green on its own commit?
Reviewed 2026-09-24 on `go1.27.1 linux/amd64`, Node v24.20.0, npm
11.19.0, golangci-lint 2.13.2, gitleaks 8.30.1, and Chromium headless.

Every claim below comes from a probe or a quoted source. The probes are
under `/tmp/claude-1000/-var-home-glw907-Projects-dubplate/89b265c3-0b9e-4d9c-b998-df7b9970c32a/scratchpad/rv/`:
`app/` is a SvelteKit app at R-5cii16's exact pins, `srv/` is a prefix
server, `igm/` is a scratch module for the `ignore` directive, `gl/` is
the gitleaks probe, and `n5/` and `cur/` hold the Unicode probes.

**Counts:** 0 blockers, 4 majors, 11 minors. Two owner forks (M2, m4).

## What the probes confirmed

These are the load-bearing claims that held. They are listed so that
no fold reopens them.

- **R-5cii2 works under a runtime prefix.** The build used
  `adapter-static`, `prerender = true`, `ssr = false`, no fallback, and
  default `paths` and `trailingSlash`. Served under `/p/q`, all three
  documents hydrated in headless Chromium. `$app/paths` `base` read
  `/p/q`. A `goto(base + '/members')` from `/p/q/login` rendered the
  members page. The server logged no 404 under the prefix. The only 404
  was the browser's own `/favicon.ico`, which is outside the prefix.
- **Build output.** The build writes exactly `login.html`,
  `login/link.html`, `members.html`, `_app/` (with `_app/version.json`),
  and `placeholder.txt`. References are relative (`./_app/…` and
  `../_app/…`). The boot script computes `new URL("..", location)`.
  The build took 11.2 s wall time with a peak RSS of 766 MB, well under
  the 3G light-lane cap.
- **CSP, R-5cii7.** In hash mode, each prerendered document carries one
  inline `<script>` and a `<meta http-equiv="content-security-policy">`.
  Its `script-src` holds the `sha256-` of that script's body. A Python
  recomputation matched all three documents. `style-src` carries
  `'unsafe-inline'` and no hash. When `frame-ancestors` is configured,
  kit drops it from the meta silently, with no build warning. The meta
  policy did not block hydration.
- **The `app.html` token.** `%dubplate.theme%` survives into each
  document exactly once, on `<html>`, and nowhere under `_app/`. Kit
  prints no warning for it.
- **Theme composition, R36 and Task 4.** Two `@plugin "daisyui/theme"`
  blocks with the same `name: "dubplate-gruvbox"`, in two imported
  files, came out as one merged rule holding both files' variables. The
  theme rule sits in `@layer base`, so the unlayered `branding.css`
  outranks it at any link order.
- **G3 is reachable.** A seeded `@import url(https://…)` does not fail
  the build. It survives into the built CSS, so G3 sees it in the
  source and in the bundle alike.
- **The `ignore` directive.** `go help packages` lists "Directories
  matching an ignore directive in a module's go.mod file". A scratch
  module was set up with `ignore ./web/node_modules`, a real 93 MB
  `node_modules`, a planted uncompilable `.go` file, and `//go:embed
  all:build` over a real bundle. In that module, `go list ./...` omits
  `node_modules`, and `go vet`, `golangci-lint run ./...` (2.13.2),
  `go mod tidy`, and the embed all run clean.
- **The stdlib quotes.** The quotes of `net/http.Error`,
  `CrossOriginProtection`, `Request.ParseForm`, `embed` (`all:` and
  invalid names), and `io/fs.Sub` match `go doc` verbatim.
  `CrossOriginProtection.Check` reads `Sec-Fetch-Site` before `Origin`
  (`csrf.go`). So a same-origin POST that carries `Origin: null` under
  `Referrer-Policy: no-referrer` is still admitted.
- **n5, R-5cii5.** On `unicode.Version` "17.0.0", the formula yields
  4174 code points. The set also equals, point for point, the 27
  `Default_Ignorable_Code_Point` ranges in Unicode 17.0.0's
  `DerivedCoreProperties.txt` (0 missing, 0 extra). U+2800 is not in
  the set. The current normalizer (`identity.go:47-72`) already refuses
  U+200B, U+00AD, and U+E0001, which are `Cf`. It admits U+115F,
  U+1160, U+3164, U+FFA0, U+034F, U+FE0F, and U+2800 in both the local
  part and the domain. So Task 2's mutations 4, 5, and 6 fire. The
  three `Cf` rows are regression rows, not proofs of the new rule.
- **The pins.** `npm view` confirms every R-5cii16 version is the
  newest. TypeScript's newest is 7.0.2, and `typescript@6` tops out at
  6.0.3. The peer ranges are kit `^5.3.3 || ^6.0.0` and svelte-check
  `^5.0.0 || ^6.0.0`, and the peers of vite-plugin-svelte 7.3.1 and
  `@tailwindcss/vite` admit vite 8. `npm install` of the whole set
  raised no peer errors.
- **Node 24 runs `.ts` directly.** Type stripping ran a two-file gate
  script that imports with an explicit `.ts` extension, and printed no
  warning. A non-erasable construct, such as an `enum`, throws.

## Findings

### M1 (major): Task 5 cannot pass its own prefix proof

**Location:** plan lines 1399-1403, 1435-1439, 1449-1451, and 1828-1833.

**Defect:** Task 5 inserts the tier 2 and tier 3 `<link>` elements
(`branding.css`, and `operator.css` when the key is set). Its prefix
proof requires every `src` and `href` in each served document to
resolve to a 200 through `shell.Handler()`. The `branding.css` and
`operator.css` routes land in Task 6. On Task 5's commit, the
`branding.css` href therefore answers the terminator's 404, and the
proof fails. The plan's own split note concedes the gap: "The routes
would answer the terminator's 404 in between". Task 5's link-order row,
"operator link only when the key is set, proved both ways", also relies
on a file key that Task 6 validates.

**Fold:** scope Task 5's prefix proof to the references the build
emitted. Task 6 extends it to `branding.css` and `operator.css`. The
alternative is to have Task 5 register `branding.css` as the empty
stylesheet and leave the grammar to Task 6. Either way, state which
task makes the `href` return 200.

### M2 (major, OWNER FORK): the TypeScript gate scripts are never type-checked, and checking them needs a dependency outside the pins

**Location:** plan lines 810-815 (R-5cii16), 1208-1211, and 1229-1232.

**Defect:** svelte-check only checks what kit's generated
`.svelte-kit/tsconfig.json` includes: `src/**`, `test/**`, `tests/**`,
and the vite config. It does not include `web/scripts/gates/`, and
Node's type stripping does no type checking. So "the gate scripts are
TypeScript under `ts-conventions`" gates nothing.

Adding `scripts/**/*.ts` to the include was measured. It fails with
`Cannot find name 'node:fs'` and `Cannot find name 'process'`, because
`@types/node` is not in the pin set. R-5cii16 and "Owner touchpoints"
make any added package an owner question, so Task 3 would stop
mid-execution.

A related measurement: under the pin set, TS 6 with no `@types/node`
and strict svelte-check reports 41 errors inside
`@sveltejs/kit/types/index.d.ts` ("Cannot find name 'http'") unless the
tsconfig sets `skipLibCheck: true`. That is the sv template's default,
but the plan never names it.

**Owner fork:**

- (a) Add `@types/node` at an exact 24.x pin and include
  `scripts/**/*.ts` in the checked set. **Recommended:** it is
  types-only, it is a devDependency, and it makes the gate scripts meet
  the plan's own bar.
- (b) Keep the pin set and declare the gate scripts untyped. Write them
  as `.js` with JSDoc and drop the `ts-conventions` claim.
- (c) Write G1, G3, and the server-file check as Go tests, which the Go
  gate already type-checks and which need no Node types.

Whichever option is chosen, R-5cii16 or Task 3 must name the tsconfig's
`skipLibCheck: true`.

### M3 (major): the install-state comparison has an unnamed state that every correct install hits

**Location:** plan lines 665-688 (R-5cii11), 1265-1269.

**Defect:** In the probe, `package-lock.json` holds 116 package
entries and `node_modules/.package-lock.json` holds 69. The 46 entries
missing from the hidden lockfile are all `optional: true` platform
binaries, such as `@rolldown/binding-darwin-arm64` and
`@rolldown/binding-android-arm64`. npm never installs them on
linux-x64. No package appeared only in the hidden lockfile, and no
shared entry differed in version.

The table row "the lockfile pins a version the install does not carry"
reads as "every lockfile entry must be present". Read that way, the
comparison fails on every clean `npm ci`. The first gate run finds it,
and the likely hasty fix is to ignore every absent entry, which would
also hide a missing non-optional package.

**Fold:** add two rows to R-5cii11's table:

- "An `optional: true` entry is absent from the hidden lockfile: the
  check proceeds."
- "A non-optional entry is absent: the check fails, naming it."

Compare `version` only for entries present in both files, and skip the
root `""` key. Task 3's drift fixture already exercises the shared-entry
path.

### M4 (major): the non-form Content-Type refusal will refuse 5c-iii's own page if it compares strings

**Location:** plan lines 477-482 (R-5cii4) and 1003-1006 (Task 1, POST).

**Defect:** The plan quotes `ParseForm` for why a non-form body must
be `malformed`, but it never says how the type is compared. A `fetch`
whose body is `URLSearchParams` sends
`application/x-www-form-urlencoded;charset=UTF-8`. This was measured as
`new Request(…, {body: new URLSearchParams(…)}).headers.get('content-type')`
under Node 24, which follows the Fetch standard's body extraction.
That is the natural way for the 5c-iii add-member page to submit.

A literal equality check against `application/x-www-form-urlencoded`
passes every Task 1 fixture that uses the bare type. It then answers
`malformed` to the real page. `ParseForm` itself accepts the parameter,
because it parses the header through `mime.ParseMediaType`.

**Fold:** state that the check parses the header with
`mime.ParseMediaType` and compares the media type. Add two outcome
rows: the `;charset=UTF-8` variant is accepted, and
`multipart/form-data` and a missing Content-Type are `malformed`. Add a
mutation: "compare the raw header → the charset row answers
`malformed`."

### m1 (minor): gitleaks already skips `node_modules`, so Task 3's byte-count criterion proves nothing

**Location:** plan lines 656-661 and 1273-1274.

**Defect:** With `[extend] useDefault = true`, gitleaks 8.30.1 scanned
0 bytes of `web/node_modules` and 0 bytes of `package-lock.json`,
because the built-in allowlist already covers both. The entry that
matters is `web/.svelte-kit/`. Without it, gitleaks scanned 563 KB, and
the full probe scan took 2.5 s against 0.35 s with it. `web/build/`
scanned 97 KB clean. As written, the "within 5% with and without
`node_modules`" criterion is true before Task 3 changes anything.

**Fold:** drop the `node_modules` allowlist entry or mark it
belt-and-braces. Re-aim the criterion at `.svelte-kit`, measured with
and without the entry. Correct the allowlist comment's reason.

### m2 (minor): DaisyUI's default plugin emits the block W5.9 forbids

**Location:** plan lines 1223-1225 and 1333-1335.

**Defect:** A bare `@plugin "daisyui"`, with its built-in themes left
enabled, emits `@media (prefers-color-scheme:dark){:root:not([data-theme]){…}}`
from its `dark --prefersdark` default, as measured in the built CSS.
Task 4 then says "no `:root:not([data-theme])` guard and no
`prefers-color-scheme` block". A reviewer reading W5.9 against the
bundle will flag the build output itself.

**Fold:** name the plugin line in Task 3 as
`@plugin "daisyui" { themes: light --default, dark; }`. With it, the
built CSS has 0 `prefers-color-scheme` rules, and `light` and `dark`
stay available for G10.

### m3 (minor): the built selector is unquoted

**Location:** plan lines 702 and 1352-1354.

**Defect:** The minified bundle writes
`:is(:root:has(input.theme-controller[value=dubplate-gruvbox]:checked),[data-theme=dubplate-gruvbox])`,
without quotes. Task 4's criterion names the literal
`[data-theme="dubplate-gruvbox"]`, which a text search will not find.
R-5cii12's startup check ("exists as a theme block in the embedded
bundle") has the same exposure. A needle written with quotes refuses
startup on every correct bundle. The plan also never says how Go finds
the hashed CSS file (`_app/immutable/assets/0.<hash>.css`).

**Fold:** match `[data-theme=<name>]` in both quoted and unquoted forms,
or assert through the document's own stylesheet `<link>`. Name the file
discovery in R-5cii12, either as the document's `rel="stylesheet"` link
or as a walk of `_app/immutable/assets/*.css`.

### m4 (minor, OWNER FORK on Ruling 2): HSTS on `https://localhost` pins every localhost port

**Location:** plan lines 99-117 and 539-551.

**Defect:** RFC 6797 section 8.3 matches a Known HSTS Host by host
alone: step 1 is "Extract from the URI any substring described by the
host component". Section 8.1.1 exempts only IP-literal and IPv4
hosts. `validatePublicURL` accepts `https://localhost`
(`cmd/dubplate-server/auth.go:465`). An adopter who tries dubplate
behind a local TLS proxy would then pin every `http://localhost:*`
service in that browser for a year. RFC 6797's IP-literal exemption
already makes `https://127.0.0.1` harmless.

**Owner fork:**

- (a) Keep the ruling as drafted: HSTS whenever the scheme is https.
- (b) HSTS only when the scheme is https and `isLoopbackHost` is false.
  **Recommended:** the function already exists
  (`auth.go:474-482`), and the carve-out matches R38's loopback set.

### m5 (minor): the operator-path criterion expects the wrong refusal

**Location:** plan lines 1515-1517.

**Defect:** `<prefix>/operator.css/../x` never reaches the route.
Normalize refuses dot segments with 400 (`internal/serve/normalize.go`,
rule 2, and the "Every rule but the length bound refuses with 400"
ruling). The criterion as written, "the terminator's 404 or the same
file", fails against correct code.

**Fold:** expect Normalize's 400 for the `..` row. Add
`<prefix>/operator.css/x` as the row that reaches the route and must
answer the terminator's 404.

### m6 (minor): sub-path states for `roster`, `members`, and `_app` are unnamed

**Location:** plan lines 841-846, 996-1002, and 1411-1415.

**Defect:** `Shell.Register` mounts `name` and every path under
`name/` (`internal/serve/serve.go:90-94`). Task 5 names the
non-exact-path row for `login` only. For a verified admin, the plan
never says what `<prefix>/roster/x`, `<prefix>/members/`, or
`<prefix>/members/x` answer. A `members/` document would also compute
the wrong base (R-5cii2).

`_app` has no handler named. If Task 5 uses `http.FileServerFS`, a
directory request such as `<prefix>/_app/immutable/` answers a 200
directory listing, and a request for `…/index.html` answers a 301.
Neither is the terminator's 404, which W3.7 assigns to an owned path
that matches nothing.

**Fold:** add rows:

- A non-exact path under `roster` or `members`, for a verified admin,
  answers the terminator's 404.
- Under `_app`, a directory path and a missing file answer the
  terminator's 404, and no directory is listed.

### m7 (minor): `RequireAdmin` is exported six commits before its first cross-package caller

**Location:** plan lines 832-834 (R-5cii18) and 971-975.

**Defect:** R-5cii18 says `RequireAdmin` is called from
`cmd/dubplate-server` in Tasks 2 and 7. But Task 2 mounts only
`NewRosterAPI`, which composes the gate inside `internal/auth`. The
first outside caller is Task 7's `members` route. That conflicts with
CLAUDE.md's "nothing is exported ahead of a caller". The
exported-unused leg warns on it for Tasks 1 through 6, though it does
not gate.

**Fold:** keep the gate unexported until Task 7 exports it, or correct
R-5cii18's claim.

### m8 (minor): n5 leaves a false comment in a file outside Task 2's wall

**Location:** plan lines 1121-1123; `internal/auth/roster.go:646-651`.

**Defect:** `actorRefusesRune`'s doc says it refuses "the same set
[identity.NormalizeAddress] refuses". Once n5 lands, that is false,
because the address set grows by the derived ignorables and U+2800.
`roster.go` is not in Task 2's Files, so the implementer must either
leave a false comment or stop.

**Fold:** add `internal/auth/roster.go` to Task 2's Files, for that
comment only.

### m9 (minor): Task 3 has to install, but two passages say only the lane installs

**Location:** plan lines 690-692, 1202-1203, and 1765-1768.

**Defect:** Task 3 creates `web/package-lock.json`, and its gate
builds against `node_modules`. The implementer must therefore run
`npm install`. R-5cii11 says "An install is the lane's", and conductor
action 1 runs `npm ci` "only after Task 3 lands".

**Fold:** state that Task 3's implementer runs
`npm install --save-exact --prefix web` to produce the lockfile. The
leg itself still never installs.

### m10 (minor): the gate on the members document hides the response, not the route

**Location:** plan lines 200-202 and 1701-1704.

**Defect:** `_app/immutable/entry/app.*.js` names the `/members` route
(measured). A client-side navigation renders the members page from
ungated `_app` chunks without requesting `members.html`. That was
measured too: `goto` from `/login` produced the members `<h1>` with no
document request. Exit-gate item 1's "indistinguishable" is true of
the HTTP responses only. The API is the real boundary, and that is
sound.

**Fold:** add one sentence to R-5cii19 stating this, so that the
pass-end security reader does not chase route discovery through the
bundle.

### m11 (minor): the token-log test has no named way to capture the shell's log

**Location:** plan lines 1601-1603.

**Defect:** `newServeShell` builds the shell with
`serve.NewShell(prefix, strip, nil)` (`cmd/dubplate-server/serve.go:215`),
so its records go to `slog.Default()`. Capturing "a captured log from
the whole shell" then means swapping a process global, which races any
parallel test in the package.

**Fold:** name the channel. Either `newServeShell` takes the logger it
passes to `NewShell`, or the row runs without `t.Parallel` and
restores `slog.Default` in `t.Cleanup`.

## Planning-miss items (CLAUDE.md's four)

- **Quoted behaviors:** these are mostly quoted and correct. Two
  mechanisms that tasks rely on are unquoted: M4 (Content-Type
  parsing) and M3 (what the hidden lockfile omits).
- **Named states:** M3 (absent optional package) and m6 (sub-paths and
  `_app` directories) are missing.
- **Privileged-field channels:** the table is complete. For HSTS
  presence, the configured scheme is the only channel. m4 is a product
  refinement, not a missing channel.
- **Non-vacuous fixture states:** the fixture states are well named.
  One criterion is vacuous by accident: m1, the gitleaks byte count.

## Task order

Every task can go green on its own commit except Task 5 (M1). If M2's
fork is left open, Task 3 stops for an owner touchpoint.
