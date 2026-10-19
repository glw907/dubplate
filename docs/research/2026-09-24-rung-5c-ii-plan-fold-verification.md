# Rung 5c-ii plan fold verification

Target: `docs/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`,
revision 2 at `74066b1`, read against revision 1 (`3ebfe3c`), the fold
record, and the three lens reviews. Fresh context: this reader had no
part in the fold. Every claim below was checked against `go doc`
(go1.27.1), the tree at `74066b1`, the kit 2.70.3 tarball, or a scratch
probe under `/tmp/claude-1000/verify5cii/`.

**Verdict.** The fold closes every blocker and major at the location the
fold record cites. It introduces one mechanism stated without proof that
silently narrows the gate (V1), and one pre-existing gate bypass survived
all three lenses (V2). No blocker. Counts: **0 blocker, 2 major, 7
minor.** V3, V4, and V5 were introduced or left open by the fold. V7
through V9 predate it.

## 1. Did each blocker and major close where the fold record says?

Yes, all fourteen closed.

| ID | Closed at | Check |
| --- | --- | --- |
| C1 | Ruling 4 (`:155-178`); R-5cii12 "Why no default is empty" and "An empty value resolves" (`:921-934`); Task 5 states (`:1836`, `:1838`) | The registry and wall quotes match `registry.go:322-325` and `config.go:184`. No state row still promises that an empty value refuses. Under either answer to Ruling 4 the keys register. |
| C2, M4 | R-5cii4 (`:615-628`); Task 1 table (`:1339-1341`), Content-Type rows (`:1408-1411`), mutation 13 | The `mime.ParseMediaType` quote matches `go doc`. `net/http/request.go:1297` parses with the same function. Node 24 `Request` sends `application/x-www-form-urlencoded;charset=UTF-8` (re-measured). |
| C3 | R-5cii4 "The three fields" (`:608-613`); privileged row (`:484`); query-field row (`:1412-1415`); mutation 15 | Closed in behavior, since the query-key refusal keeps an address out of the URL. The row's proof is misaimed (V3). |
| C4 | R-5cii4 "The error mapping" (`:634-642`); "The `failed` rows" (`:1401-1407`); mutation 14 | The sentinels and lines 118, 131, 241, and 246 match `roster.go`. Fixture (i) is missing one fixture state (V5). |
| C5 | Task 1 GET, HEAD, and PUT rows and the store-failure row (`:1389-1400`) | Closed. The store-failure row is on the inner handler, as the review asked. |
| C6, M1 | Task 5 outcomes (`:1807-1813`) and prefix proof (`:1852-1858`); Task 6 (`:1924-1935`, `:1974-1975`); split note (`:2345-2349`) | Closed. Task 5 inserts no link, and Task 6 lands the links with their routes. |
| C7 | Task 6 tier links and anchor states (`:1924-1945`); mutation 6; Task 7 G7 (`:2065-2068`) | The kit quotes match the tarball: `render.js:83`, `:287`, `:330-345`, and `add_stylesheet` emitting `<link href="${href}" ${attributes.join(' ')}>`. The `index.js:143-146` citation is off by three lines (the comment is at `:140-141`). The content is present. |
| C8 | R-5cii12 "The composed-name check" (`:936-946`); Task 4 (`:1727-1735`, `:1752-1758`); Task 5 palette states | Re-measured on the probe build. The selector is unquoted inside the `theme-controller` `:is()` form. Two same-named `daisyui/theme` blocks merge into one rule inside `@layer base`. |
| C9 | Task 7 G6 and G7 placeholder rows (`:2062-2064`); Task 3 G3 built half (`:1638-1639`, `:1650-1651`) | Closed. |
| C10 | R-5cii7 kit emission quote and the fifth refusal (`:760-782`); G6 independence (`:787-789`); mutation 10 | The `render.js:586-588` quote matches. The probe build's `login.html` carries exactly one bare `<script>`. The refused fixture part is argued soundly. |
| R1 | R-5cii13 class rules and states (`:973-988`); privileged row (`:492`); Task 6 operator-path row and mutation 7 | Closed. The `EvalSymlinks` refusal is reasoned and proportionate. |
| M2 | Ruling 5 (`:180-200`); R-5cii16 (`:1069-1075`); Task 3 (`:1602-1604`); mutation 7 | The owner fork is honest. The 41 errors and the `process` and `node:fs` failures were re-measured. **The "yes" build is under-specified (V1).** |
| M3 | R-5cii11 table (`:883-900`); Task 3 seeds; mutation 6 | Re-measured on the probe lock: 116 entries, 69 hidden, and all 46 missing carry `optional+dev`. |

## 2. Contradictions and task order

The segment order builds: Tasks 1 and 2 are Go only, Task 3 births the
leg, Task 5 links nothing it does not route, and Task 6 routes what it
links. No new `internal/config` edit is needed, and no subtree name
collides with the three auth routes (`serve.go:41-43`). `doctor_test.go`
pins no full output, so Task 5's four new rows need no edit to it. The
ceiling arithmetic checks out (6.75M, flag 7.52M), and so do the
deliverable count (31) and the fold record's tallies (47 = 43 + 4).
The contradictions found are V3, V4, V7, and V9.

## 3. Mechanisms: quoted or proven?

These check out against the source or a probe:

- `mime.ParseMediaType` and `Request.ParseForm`, `go doc` verbatim.
- The `none` sentinel, `registry.go:322-325` and `config.go:182-187`.
- `resolveIdentitySource` requiring the selector, and
  `DUBPLATE_PUBLIC_URL` as Required (`auth.go:50`).
- `isLoopbackHost` (`auth.go:474-482`) and `validatePublicURL`
  accepting `https://localhost` (`:465`).
- RFC 6797 sections 6.1.1, 7.2, 8.1, 8.1.1, and 8.3, in
  `rfc6797.txt`.
- The route-states rule's `Register` double mount (`serve.go:95-96`) and
  `Normalize` rule 2.
- kit's stylesheet and script emission.
- The skipLibCheck count of 41, re-run.
- The `@types/node` 24.13.6 pin and the TypeScript peer ranges,
  `npm view`.
- The n5 derivation, from a Go probe: size 4174 at Unicode 17.0.0,
  4187 without the Prepended_Concatenation_Mark term, and 13 runes in
  both Cf and PCM. U+2800 is outside the set, and all ten examples
  parse through `net/mail` in both the local part and the domain.
- `CrossOriginProtection`, `Request.Context`, `fs.Sub`,
  `html.UnescapeString`, `go help packages`, and `go doc embed`.

**Not proven: Ruling 5's include (V1).**

## Findings

### V1. major: including `scripts/**/*.ts` replaces kit's include list, so the gate stops type-checking `src/`

- **Location:** R-5cii16 (`:1069-1075`); Task 3 scaffold (`:1602-1604`);
  Ruling 5 "What yes builds" (`:193-194`); Task 3 mutation 7 (`:1684-1685`).
- **Defect:** the plan says only that `web/tsconfig.json` "includes
  `scripts/**/*.ts`". A child tsconfig's `include` replaces the one it
  extends. It does not merge. The probe proved it:
  - `"include": ["scripts/**/*.ts"]` over `extends:
    ./.svelte-kit/tsconfig.json` checked 108 files. A seeded `const n:
    number = "x"` in `src/lib/bad.ts` went unreported.
  - Kit's own include list checked 178 files and reported the error.

  So the obvious reading of the plan leaves every `src/` file
  unchecked, and the gate stays green. Mutation 7 still fires, because
  scripts stay checked, so nothing catches the regression. The probe
  also showed a second gap: a gate script importing a sibling `./lib.ts`
  fails strict checking ("An import path can only end with a '.ts'
  extension when 'allowImportingTsExtensions' is enabled"). Node's type
  stripping requires that extension. The mechanics reviewer's own probe
  (`rv/app/tsconfig.scripts.json`) restated kit's list and set the flag.
  The fold carried neither.
- **Proposed fold:** in R-5cii16 and Task 3, state that `include` in
  `web/tsconfig.json` replaces kit's list. Restate kit's entries relative
  to `web/` (`.svelte-kit/ambient.d.ts`, `.svelte-kit/non-ambient.d.ts`,
  `.svelte-kit/types/**/$types.d.ts`, `vite.config.ts`,
  `src/**/*.{js,ts,svelte}`), then add `scripts/**/*.ts`. Set
  `allowImportingTsExtensions: true`, which kit's `noEmit: true`
  permits. Add a mutation: a type error seeded under `src/` must fail the
  gate after the include change. Ruling 5's "What yes builds" should name
  both settings, since Geoff is approving that build.

### V2. major: a Make target named `web` beside the `web/` directory never runs unless it is `.PHONY`

- **Location:** R-5cii10 "Order" (`:839-844`); Task 3 "The leg"
  (`:1610-1615`); `Makefile:52` (the existing `.PHONY` list).
- **Defect:** the plan adds a `web` target with no file prerequisites,
  and `check` depends on it. With a `web/` directory present, make
  treats the target as up to date and skips its recipe. The probe
  proved it: `check: web` printed only `CHECK`, and adding `.PHONY: web`
  printed `BUILDING` first. Without the entry, `make check` and
  `--scope-changed`'s `make web` skip the install-state check, the
  build, `svelte-check`, and every static web gate, and still exit 0.
  This is the "a check that does not gate" class. It predates the fold,
  and no lens raised it. Task 3's seeds would catch it only if the
  implementer runs them through `make`.
- **Proposed fold:** Task 3 adds `web` to `Makefile:52`'s `.PHONY` list.
  Add one row: with a `web/` directory present and `node_modules`
  deleted, `make check` itself fails naming `npm ci --prefix web`.
  Add one mutation: dropping `web` from `.PHONY` makes that row pass.

### V3. minor: mutation 15 cannot fire, and two neighbors misstate what they prove

- **Location:** mutation 15 (`:1460`); the query-field row's fixture
  state (`:1414-1415`); mutation 5 (`:1444`); privileged row (`:484`).
- **Defect:** R-5cii4 refuses as `malformed` any query key that
  survives the admin and unknown-key checks. That leaves no query key
  among the three fields. After the check, the three fields in `r.Form`
  equal the three in `r.PostForm`, so reading from `r.Form` is an
  equivalent mutant. The row's claim that such a read "adds the
  member" is false: the query-key check refuses first. Mutation 5 does
  fire, but the query `?admin=` answers `malformed`; it does not write.
  Row `:484` says a query key answers "`malformed` or
  `admin_not_grantable`", but `?foo=1` answers `unknown_field`. An
  `opus` implementer who must prove every named mutation meets a
  survivor on Task 1.
- **Proposed fold:** replace mutation 15 with "drop the query-key check
  → the query-field row answers a field outcome, not `malformed`". State
  that the `r.PostForm` read is defense in depth, equivalent under the
  check. Correct mutation 5's symptom, and add `unknown_field` to row
  `:484`.

### V4. minor: nothing names how the roster API answers only its exact path

- **Location:** Task 1 (`:1309-1310`); Task 2 (`:1535`); R-5cii21 roster
  row (`:1138`); `AdminConfig` (`:1281-1282`).
- **Defect:** the m6 fold asks the API to answer only `<prefix>/roster`.
  Two facts leave the mechanism open:
  - `AdminConfig` carries no prefix or path.
  - The shell dispatches through `s.mux.Handler(r)`
    (`internal/serve/serve.go:218`), which leaves `r.Pattern` unset.

  An improvised `HasSuffix(path, "/roster")` passes the `roster/x` row
  but answers the list at `<prefix>/roster/roster`. The exposure is
  admin-only, so this is minor.
- **Proposed fold:** name the mechanism. One option: Task 2 wraps the
  mount in `cmd/dubplate-server` with an exact-path check. The gate's
  404 and the terminator's are byte-equal, so order does not show.
  Another option: `AdminConfig` carries the route path. Add a
  `<prefix>/roster/roster` row.

### V5. minor: `failed` fixture (i) can pass without reaching `errActorNotAdmin`

- **Location:** `:1402-1404`.
- **Defect:** if the hand edit clears the only admin's flag, the
  re-read under the lock fails with `errNoAdmin` (`roster.go:460`,
  which also wraps `ErrRosterRecordInvalid`). The API still answers 500
  `failed`, so the row passes, but it proves the load refusal and not
  the under-lock actor check the privileged-fields table relies on
  (`:482`).
- **Proposed fold:** the fixture's roster keeps a second admin after
  the edit. The row first asserts that `OpenRoster` on the edited file
  succeeds.

### V6. minor: Ruling 2 offers a known-wrong branch

- **Location:** Ruling 2 (`:107-135`).
- **Defect:** the loopback exception is a correctness call. RFC 6797
  section 8.3 matches by host alone, so HSTS from `https://localhost`
  pins every localhost port. That was verified. Yet "What no builds"
  still offers "drop the exception".
- **Proposed fold:** fold the exception as fixed, as the fold record
  did for m4's substance. Ruling 2 then asks only for the HSTS shape
  (`max-age`, no `includeSubDomains`, no `preload`).

### V7. minor, pre-existing: Task 5 mutation 2 contradicts the "token present more than once" row

- **Location:** Task 5 states (`:1834`); mutation 2 (`:1879-1880`).
- **Defect:** planting the token inside a script body makes the document
  carry two tokens. The state table refuses that at startup under the
  correct implementation too, so the mutation cannot be told apart from
  it.
- **Proposed fold:** define the token's one legal site as the root
  element's `data-theme` attribute value, and count occurrences outside
  script bodies. Or drop mutation 2, since the byte-identical script
  criterion (`:1864-1866`) already guards the class.

### V8. minor, pre-existing: Task 2's Files omit a site for the command-line `AddMember` n5 row

- **Location:** Task 2 Files (`:1484`) against its criterion (`:1521`).
- **Defect:** the Files list limits `roster_test.go` to "the load-path
  row for n5 only", and no member-verb test is listed.
- **Proposed fold:** widen the parenthetical to "the n5 rows".

### V9. minor, pre-existing: two cache rows contradict each other for JSON refusals

- **Location:** R-5cii8 (`:806-807`).
- **Defect:** the JSON row sets `no-store` on every JSON answer. The
  refusal row sets nothing on every refusal. A 400 `malformed` or a
  503 `busy` is both a JSON answer and a refusal. `writeAnswer`
  (`handlers.go:416`) sets `no-store` today.
- **Proposed fold:** reword the last row to "every 404 and every
  non-JSON refusal".

## 4. The rulings

Geoff can rule all seven on the evidence given. The quotes behind
Rulings 2, 4, 5, and 6 were verified, and Ruling 6 names its one
unverified point (WebKit's serialization). Only Ruling 2 carries part
that should be folded instead: its loopback exception is RFC-driven
(V6). Ruling 5's "yes" branch should carry V1's two tsconfig settings
before Geoff approves it. Otherwise he approves a build that silently
narrows the gate. Rulings 1, 3, 4, and 7 are genuine budget, taste,
adopter-configuration, and availability forks, and each recommendation
is supported.
