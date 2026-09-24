# Rung 5c-ii plan review: contract and criteria lens

Target: `docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`
at `3ebfe3c` (revision 1, draft). Sources read: the repo `CLAUDE.md`, the
rung 5 spec (5c row, "The pages", "The gate leg"), the web surface
standard (section 15, W3.7, W5.9, W5.10), the 5c-i plan (R-5c12 through
R-5c20), DL R5.6 and G6, and the tree at `3ebfe3c`
(`internal/auth/roster.go`, `handlers.go`, `token.go`, `seam.go`,
`internal/identity/identity.go`, `internal/serve/prefix.go`, `serve.go`,
`normalize.go`, `internal/config/thresholds.go`, `registry.go`,
`config.go`, `cmd/dubplate-server/serve.go`, `auth.go`), plus the plan's
own kit 2.70.3 tarball in the session scratchpad.

The lens asks four things. Can every promise be tested? Does each
criterion name its fixture and the reason it would fail? Can any
criterion pass vacuously? Does the plan meet CLAUDE.md's four
planning-miss items? Findings are ranked by consequence. One probe was
run: `identity.NormalizeAddress` at HEAD against the ten n5 code points,
in a scratch module copy of `identity.go`. No repo file was touched.

**Counts:** 1 blocker, 9 major, 9 minor. No owner fork. C1's fold picks
an adopter-visible config shape, so it lands in the plan revision that
Geoff approves.

---

## Blocker

### C1. Two of the four presentation keys cannot register, and one state row cannot be reached

- **Severity:** blocker
- **Location:** plan `:699-708` (R-5cii12's table), `:733` (R-5cii13
  "key empty"), `:1405-1410`, `:1426-1427` (Task 5's "anything else,
  empty included" row)

**Defect.** R-5cii12 declares `DUBPLATE_BRANDING` and
`DUBPLATE_OPERATOR_CSS_FILE` as `KindString` thresholds with an empty
default. The registry panics on that case (`internal/config/registry.go:322-325`):
`case KindString: if t.Default == "" { panic(fmt.Sprintf("%s: invalid default %q", ...)) }`.
`ParseString` documents the reason (`thresholds.go:124-128`): "the one
invalid Default is the empty string, because wall.get's
empty-equals-unset convention would otherwise make an empty-string
default indistinguishable from no default at all". The repo already
records this constraint (`cmd/dubplate-server/auth.go:33-36`: "a
threshold's KindString default must be non-empty"). Task 5's one-line
`registerWebConfig(reg)` in `newProductionRegistry` would therefore
panic in every verb. The wall forbids the `internal/config` edit that
would lift the rule.

The same convention (`config.go:182-187`, `if !ok || v == ""` returns
unset) makes `DUBPLATE_THEME_MATCH=` resolve to the default `dubplate`.
So the match key's "empty included → serve refuses" row can never occur.
A test written to it either fails, or passes against something other
than what it names. R-5cii12 quotes `go doc ./internal/config Threshold`
but not the registry's validation, which is the planning-miss item
"quoted, never recalled".

**Proposed fold.** Quote `registry.go:322-325` and `config.go:182-187`
in R-5cii12. Then pick one mechanism for the two optional keys:
- (a) Non-empty sentinel defaults: `DUBPLATE_BRANDING=none` and
  `DUBPLATE_OPERATOR_CSS_FILE=none`, with `none` meaning absent and
  refused as a literal path.
- (b) The keys as an integration class's `Optional` rows, which is the
  precedent the two rotation secrets use (`auth.go:30-36`). `doctor`
  would then no longer print them, and the contested-surfaces row's
  "four `doctor` rows" changes.

Recommendation: (a). It keeps `doctor` reporting the effective value,
which R14 and R-5cii12 want. Either choice changes the adopter-visible
default, which the plan's own "Owner touchpoints" names, so it belongs
in the revision Geoff approves. Delete the match key's "empty included"
clause, or restate it as "empty resolves to the default" with a test
that proves it.

---

## Major

### C2. The Content-Type refusal does not say how the header is compared, and a string-equality check refuses every real page POST

- **Severity:** major
- **Location:** `:476-482` (R-5cii4), `:1003-1006`, `:1028`

**Defect.** The plan refuses "a body whose Content-Type is not
`application/x-www-form-urlencoded`" but never says how the header is
compared. A browser `fetch` with a `URLSearchParams` body sends
`application/x-www-form-urlencoded;charset=UTF-8`. An implementer who
compares with `r.Header.Get(...) == "application/x-www-form-urlencoded"`
passes every Go fixture, which sends the bare type. The same check then
refuses 5c-iii's add-member page as `malformed`. No acceptance row names
a parameterized header, so the defect ships green.

**Proposed fold.** State that the check parses the header with
`mime.ParseMediaType` (the parser `ParseForm` itself uses; quote `go doc
mime.ParseMediaType`) and compares the media type alone. Add two rows,
each with its fixture: `application/x-www-form-urlencoded; charset=utf-8`
is accepted, and `multipart/form-data; boundary=x` answers `malformed`.
Add the mutation "compare the raw header string → the charset row
answers `malformed`".

### C3. The roster API reads its fields from `r.Form`, which contradicts R-5c13 and puts addresses in URLs

- **Severity:** major
- **Location:** `:472-476` (R-5cii4), `:327-329` (the three-outputs
  constraint), and R-5c13 in the 5c-i plan (`:737-739`)

**Defect.** R-5c13, which the plan says governs Task 1, reads the three
fields "from `r.PostForm` only, on 5b-ii's precedent that nothing is
read from a query string". R-5cii4 closes the field set over `r.Form`,
the body and query together, and neither Task 1's outcomes nor its
states say which one the fields come from. Nothing names the state "a
field present only in the query". Read through `r.Form.Get`, a POST to
`<prefix>/roster?address=...&display_name=...&streamer_username=...`
with an empty form body adds a member. That puts the address in the
request URL, which tunnel and proxy access logs record. The global
constraint that addresses reach exactly three outputs then fails
somewhere this pass cannot see.

**Proposed fold.** Restate R-5c13: the three fields are read from
`r.PostForm`, and the admin and unknown-key checks run over `r.Form`. Add
a state row and test: any of the three field keys present in
`r.URL.Query()` answers `malformed` (or `unknown_field`, whichever the
revision rules) with the file bytes unchanged. Keep the `?admin=` row as
it is. The field fixture is a query-only address with an empty body, so
a `PostForm` read that falls back to `Form` fails the row.

### C4. The `failed` row has no fixture, and the sentinel wrapping sets a trap the mutations miss

- **Severity:** major
- **Location:** `:1031-1032`, `:360` (the privileged-fields row
  "A stale gate decision: `AddMember` re-checks the actor under the
  lock"), `:1066-1070`

**Defect.** Every validation refusal from `AddMember` wraps both
`ErrRosterRecordInvalid` and its specific sentinel (`roster.go:772-779`,
`:800-811`). That includes `errActorNotAdmin`, which the table sends to
500 `failed`. A mapper that tests `errors.Is(err, ErrRosterRecordInvalid)`
before the specific sentinels turns the actor-not-admin refusal into a
400. No mutation catches that, because nothing gives the `failed` row a
fixture. The privileged-fields table calls the under-lock re-check the
second net for the one privileged write, and no Task 1 or Task 2
criterion proves the API reaches it. `errRosterWritableByOthers` is also
raised on the read path (`roster.go:319`), so its fixture is not obvious.
A chmod changes no mtime, so the gate's cached `Lookup` passes and
`writeLocked`'s re-read refuses. An implementer has to find that timing.

**Proposed fold.** Give the `failed` row two named fixtures:
- (i) An in-package test that builds the inner handler with a context
  identity whose disk record has had its admin flag removed by a hand
  edit after the gate's check. It expects 500 `failed`, bytes unchanged,
  and no audit record.
- (ii) The roster chmodded to 0666 after the gate's `Lookup` cached it.

Add the mutation "map `ErrRosterRecordInvalid` to 400 before the
specific sentinels → fixture (i) answers 400". List every roster
sentinel that the table sends to `failed` by name: `errActorNotAdmin`,
`errRosterWritableByOthers`, `errInvalidActor`, and `errNoAdmin`
(unreachable on add, so say so). Also map `errRosterBusy` to the `busy`
row by name.

### C5. GET, HEAD, and 405 carry outcomes but no acceptance criterion, and the GET store-failure row cannot be reached through the gate

- **Severity:** major
- **Location:** `:998-1002` (outcomes), `:1050-1079` (acceptance)

**Defect.** The GET answer is the contract 5c-iii's list page builds on.
It should carry a `no-store` JSON `members` array with four named
fields, "500 `failed`, never a partial list" on a store failure, and
405 with `Allow: GET, HEAD, POST`. None of that has an acceptance
criterion or a mutation. The store-failure row cannot be reached through
`RequireAdmin`. The gate's seam `Lookup` reads the same roster, so a
broken file answers the gate's 404 before `List` runs. A test written
through the composed handler therefore proves the gate, not the row.

**Proposed fold.** Add these criteria:
- A GET row with the fixture "a roster holding an admin with a
  streamer_username and a CLI-created admin without one". It asserts
  each member's four fields, `Cache-Control: no-store`, and file order.
- A HEAD row.
- A PUT row answering 405 with the exact `Allow` value, reached only by
  an admin.
- The store-failure row, built on the in-package inner handler with an
  admin identity in context and a roster file corrupted after `OpenRoster`.

### C6. Task 5's criteria need routes and a key that land in Task 6

- **Severity:** major
- **Location:** `:1405-1407`, `:1435-1439`, `:1448-1451`; Task 6
  mutation 5 (`:1527-1528`)

**Defect.** Task 5's rewrite inserts `branding.css` and `operator.css`
links. Its gate-resident prefix proof requires "every `src` and `href`
[to resolve] to a 200", but those two routes land in Task 6. So the
criterion cannot pass as written. Task 5 also requires "the operator
link only when the key is set, proved both ways", while it declares the
branding keys "here as well and consumed in Task 6". Nothing in Task 5
reads the operator key. The plan's own split note (`:1832-1833`) admits
that the routes 404 in between. The criteria were never cut to match.

**Proposed fold.** Scope Task 5's prefix proof to every reference but
the two tier-2 and tier-3 hrefs, and state the exclusion. Move the
operator key's resolution into Task 5, or move the "operator link only
when set" row into Task 6. Task 6 then extends the prefix proof over
both stylesheet hrefs. Name the fixture for Task 5's link-order row as
the branding key unset, so Task 6's mutation 5 has a row to break.

### C7. The rewrite's insertion point assumes exactly one bundle stylesheet link

- **Severity:** major
- **Location:** `:1399-1404` (the rewrite), `:1418-1431` (states)

**Defect.** The tier order depends on "inserts the tier 2 and tier 3
`<link>` elements after the bundle stylesheet". The plan never quotes
how kit links stylesheets in an `ssr = false` prerender. kit 2.70.3 does
link them, and here is the source:
- `src/runtime/server/page/index.js:143-146` calls `render_response`
  with "provide nodes without running load functions so that the styles
  and fonts are linked in the head before CSR takes over".
- `render.js:83` seeds `stylesheets` from `client.stylesheets`, and
  `render.js:287` adds each node's.
- `render.js:330-345` emits one `<link rel="stylesheet">` per entry.

A document therefore carries one link per stylesheet chunk: the entry,
the layout, and each page with component CSS. The stubs have one chunk
today. 5c-iii's pages will add more. A rewrite that anchors on the first
link puts `branding.css` ahead of later bundle CSS, and that inverts DL
R4.1's tiers without failing any 5c-ii test. The states table names no
zero-link or many-link row.

**Proposed fold.** Quote the three kit sites above. Define the anchor as
"after the last kit-emitted `rel="stylesheet"` link". Add states:
- Zero stylesheet links: serve refuses startup, naming the document.
- One link: the anchor is that link.
- Several links: the anchor is the last one, proved on a fixture
  document with two.

Add the mutation "anchor on the first link → the two-link fixture's
order fails".

### C8. The composed-name check and Task 4's criterion rest on an unquoted build output, and the palette value that becomes attribute markup has two of its states named

- **Severity:** major
- **Location:** `:702`, `:1329-1335`, `:1352-1354`, `:1408-1410`,
  `:1428-1429`; the privileged-fields row at `:366`

**Defect.** Three behaviors are relied on without a quote:
- How DaisyUI 5's theme plugin composes one theme from two separate
  files. A CSS `@plugin` block cannot import a file, so the mechanism is
  a real unknown. Task 4 defers the quote to the implementer.
- The selector form the minified build emits. Tailwind 4 builds
  minify, and a minifier may drop the quotes from `[data-theme="..."]`.
  DaisyUI also emits the name inside a `theme-controller` `:has()`
  selector.
- Whether values survive minification. Task 4's "with the data files'
  values" assumes byte identity.

Task 5's "composed name present in the bundle" check reads that same
output. A substring search would accept a palette value that is a prefix
of a real one (`gruv` against `dubplate-gruvbox`). R14 names the theme
value as one of "the two places operator text becomes attribute markup".
Its states are "present" and "absent" only: nothing names empty, a
prefix, or a value carrying `"`, `>`, or whitespace.

**Proposed fold.** Require a Task 0 or Task 4 measurement. Build once,
record the emitted selector form and value form in the report, and quote
DaisyUI 5.7.45's theme-plugin source for the two-file composition. Write
Task 4's criterion against the measured form, comparing values after
parsing. At serve resolution, refuse any palette outside `[a-z0-9-]+`
before the lookup, and match the composed name as an exact token
against the set of selectors extracted. Add these states, each a test:
- Empty resolves to the default (C1).
- A prefix of a real palette is refused.
- A value carrying `"` is refused by the charset rule.

### C9. Exit item 6 promises placeholder-tree failures that no task proves, and G6 can pass on zero documents

- **Severity:** major
- **Location:** `:783-785` (R-5cii15), `:1576-1584` (G6), `:1585-1589`
  (G7), `:1713-1715` (exit gate item 6)

**Defect.** The spec's exit row and R-5cii15 say G3, G6, and G7 "fail by
name against the placeholder-only tree". No acceptance criterion in
Task 3 or Task 7 runs any of them against that tree. Task 3 seeds G3 in
source only, and G3's built-bundle half has no seed. G6's guard,
"at least one inline script per document (asserted first)", runs per
document, so it holds trivially when the document set is empty. G7 has
the same weakness.

**Proposed fold.** Task 7 adds one row for G6 and one for G7. Each runs
over an `fstest.MapFS` holding only `placeholder.txt`, and each asserts
a failure that names the missing tree. Both also assert first that the
document set equals the route table's three documents. Task 3 adds a
seed for G3's built-bundle half: an empty `web/build` other than the
placeholder must fail by name.

### C10. The plan never says how inline scripts are found, so G6's independence has no contract

- **Severity:** major
- **Location:** `:585-593` (R-5cii7), `:1571-1584`

**Defect.** Two parts of the plan must find inline `<script>` bodies:
the startup extraction's fourth state, and G6. The standard library has
no HTML tokenizer (`html` only escapes). `golang.org/x/net` is not in
`go.mod`, and the plan forbids a new module dependency. The plan quotes
nothing about the shape of kit's emitted script tags. It also does not
forbid G6 from calling the production scanner. If one hand-written
scanner serves both and misses a variant (for example
`<script type="module">`, an uppercase tag, or a body that contains
`</`), the header and G6 both omit the script, G6 passes, and the page
never boots. That is risk 6, which the plan calls Task 7's upshift
reason.

**Proposed fold.** Quote kit's script emission from `render.js`: the
exact tag and attribute forms it writes. State that G6's scanner is
written in the test file and shares no function with `internal/serve`'s
extractor. Add a G6 fixture document that carries an attributed and an
uppercase script tag, so a naive scanner fails.

---

## Minor

### C11. Three of the ten n5 rows are already refused at HEAD, and the refusal outcome at request-link is not named

- **Location:** `:1156-1165`, `:1134-1137`

The scratch probe ran `NormalizeAddress` at `3ebfe3c` over `a<cp>b@example.com`
and `ab@exa<cp>mple.com`. U+200B, U+00AD, and U+E0001 are refused today
by the existing `Cf` rule (`identity.go:70-72`). The other seven are
accepted. The fixture state, "the same address without the code point is
accepted", does not show that each row proves the new rule.

**Fold:**
- Require the task to report which rows fail before the change, and
  label the three as regression rows.
- Assert that the error names `U+XXXX`, which is the outcome the plan
  states and no criterion checks.
- Assert that the derived set's size is 4174, so an over-refusal
  mutation (dropping `- White_Space` or `- Prepended_Concatenation_Mark`)
  fails too.
- Name request-link's answer: 400 `invalid_address`, after
  `allowLinkClient` charges the client (`handlers.go:194-208`).

### C12. Task 6's operator-route fixture never reaches the handler

- **Location:** `:1515-1517`

`<prefix>/operator.css/../x` is refused with 400 by `Normalize`
(`normalize.go`, rule 2) before any route runs. The criterion's
"terminator's 404 or the same file" then fails, or the test proves
`Normalize` instead of the route.

**Fold:** use `<prefix>/operator.css/x`. `Register` mounts both `name`
and `name + "/"` (`serve.go:94-96`), so that path does reach the handler.
Add the same row for `branding.css`. Add an accepted 64-byte branding
value beside the refused 65-byte one.

### C13. Review focus 4 promises a startup refusal that R-5cii7's states do not deliver

- **Location:** `:865-867` against `:585-588`

Startup refuses only when a document carries an inline script "while the
meta lists no hash". A script outside kit's set in a document whose meta
does list hashes (for example a script added to `app.html`) passes
startup. G6 still catches it in the gate. Align the two: either startup
hashes every inline script and refuses one that is missing from the
meta, or review focus 4 says "G6 fails" alone.

### C14. The asset and document routes leave some request states unnamed

- **Location:** `:1411-1416`, `:1430-1431`

Name these states:
- `<prefix>/_app/immutable/`, a directory: 404 with no listing. An
  `http.FileServerFS` lists directories by default.
- `<prefix>/_app` bare.
- A missing asset: equal to the terminator's 404.
- A non-GET, non-HEAD request to `login`, `login/link`, `_app`, and the
  two stylesheets: 405 or 404, ruled.

R-5cii19 fixes the method check for `members` only.

### C15. Task 3 seeds one row per check, not one per state

- **Location:** `:1247-1256`, `:681-688`, `:1263-1269`

The rows for an absent or empty `web/src`, several forbidden files,
Node missing from `PATH`, Node below 24, and an out-of-range install
have no seeded proof. Mutation 1 relies on the missing-Node row, which
no acceptance item lists.

**Fold:** seed and report every state row, as Task 4 does. The drift
fixture must be a transitive package: top-level pins are exact, so an
"in-range" newer lockfile pin of a top-level package describes a state
that never occurs.

### C16. The log and header-step rows cover the success path only

- **Location:** `:1601-1603`, `:1590-1595`, `:1060-1061`

The token row exercises only the served 200. Also exercise
`?token=` on the trailing-slash 404, a POST to `login/link`, and a
normalization refusal. Extend Task 1's log criterion to the POST refusal
rows: `net/mail` quotes its input, per `roster.go:699-701`, so an
`invalid_address` refusal that logs the parse error leaks the submitted
address. Add an out-of-prefix 404 to the header-step list. Mutation 1
cites "the fallthrough 404", but no row lists one.

### C17. Task 1's refresh fixture should assert its own precondition

- **Location:** `:1055-1059`, `:1087-1088`, `:996-997`

The fixture should assert that the seam returned a non-empty
`RefreshedValue` for the non-admin's session before it asserts no
`Set-Cookie`. Otherwise a session too fresh to refresh passes the row
vacuously. For mutation 3 to fire, the non-admin rows need to carry an
`admin` form value or header. "The inner handler refuses on its own"
needs a row: the inner handler with no identity in context answers
`http.NotFound`.

### C18. The HSTS states skip an unset public URL

- **Location:** `:370`, `:1593-1595`

If any selected identity source leaves `DUBPLATE_PUBLIC_URL` unset,
name that state (HSTS absent) and test it. If none can, say so with the
source line that requires the key.

### C19. Task 4's contrast expectation is computed by hand

- **Location:** `:1364-1366`

"Computed by hand" makes the implementer's own arithmetic the expected
value for mutation 2's `oklch` row. Cite an independent computation
instead, such as a published reference pair or a named library's output
recorded in the test.

---

## Areas that hold under this lens

- The gate's refusal-equality criterion compares against `http.NotFound`
  on the same request. Its fixture (a same-origin non-admin old enough
  to refresh) makes the Set-Cookie mutation fire.
- The busy, `write_unconfirmed`, and over-cap rows each name a mechanism
  that exists in the tree: the directory lock, the `writeRosterStream`
  variable seam at `roster.go:251`, and the counting reader.
- The privileged-fields table is complete apart from C3's query channel
  and C8's palette charset.
- R-5cii2, R-5cii6, R-5cii10, and R-5cii11 quote or measure every
  behavior they rely on. The two-way reload row names its fixture order.
- Task 2 mutation 3 is honestly declared as a non-firing mutation.
- The cache table (R-5cii8) and the G7 cache row map one to one.
