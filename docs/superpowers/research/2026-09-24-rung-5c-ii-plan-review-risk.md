# Pass 5c-ii plan review: domain risk lens (security and failure)

Target: `docs/superpowers/plans/2026-09-24-rung-5c-ii-admin-gate-roster-api-web.md`
at `3ebfe3c` (revision 1, draft). Read against the rung 5 spec, DL R5.6
and G6 through G8, the 5c-i plan's R-5c12, R-5c13, R-5c19, and its 5c-ii
outline, the 5c-i security review's found-sound list, and the Go tree on
master (`internal/auth`, `internal/identity`, `internal/serve`,
`cmd/dubplate-server`). Kit behavior was checked against the plan's own
kit 2.70.3 tarball in the session scratchpad.

## Verdict

No blocker. The plan's core security design holds: the gate's order, the
404 equality, the closed field set, and the actor source are sound. So
are the cache table and the CSP extraction states. One major: the
operator CSS route can publish any file the service user can read. The
twelve minors are named states, dropped or synthetic proof rows, and
carries. Two of them are owner forks.

Counts: **0 blocker, 1 major, 12 minor.**

## Findings, ranked by consequence

### R1. MAJOR: the operator CSS route publishes any readable file, unauthenticated

- **Location:** plan:722-737 (R-5cii13), plan:1498-1518 (Task 6).
- **Defect:** `DUBPLATE_OPERATOR_CSS_FILE` accepts "a path to a readable
  regular file within" 256 KiB. The route then serves those bytes
  publicly at `<prefix>/operator.css`, with no session check. The roster
  file (mode 0600, owned by the service user, so readable) fits the
  bound, and so do most files under the state directory. So does any
  secrets file the unit can read. A stranger adopter pastes the wrong
  path, or copies the roster line in the env template. The server then
  starts cleanly and publishes every member's address to the internet.
  No state row and no test covers this class. Likelihood is low; the
  consequence is PII or secret disclosure with no signal.
- **Proposed fold:** add three refusals to R-5cii13's state table, each
  a Task 6 test. First, the path must end in `.css`. Second, a path that
  resolves (`filepath.EvalSymlinks`) to the roster file or under
  `StateDir` refuses startup, naming the key. Third, content that is not
  valid UTF-8 or carries a NUL byte refuses. **Fixture state:** point
  the key at the test's own roster fixture and assert the startup
  refusal. These are cheap, and they turn a silent disclosure into a
  named startup error.

### R2. MINOR: the non-form Content-Type refusal has no named state for a media-type parameter

- **Location:** plan:476-482 (R-5cii4), plan:1003-1006 (Task 1 POST).
- **Defect:** Task 1 refuses "a non-form Content-Type" as `malformed`.
  A browser `fetch` with a `URLSearchParams` body sends
  `application/x-www-form-urlencoded;charset=UTF-8`. An equality check on
  the header string refuses every real add from the 5c-iii page, while
  every Go test that sends the bare type passes. The stdlib's own
  `parsePostForm` parses the header through `mime.ParseMediaType`. This
  is planning-miss item 2: the check can meet more states than the plan
  names.
- **Proposed fold:** quote `mime.ParseMediaType`. Add a state row: the
  form type with a `charset` parameter proceeds. Add a mutation:
  compare the raw header string, and the charset row answers
  `malformed`.

### R3. MINOR: deeper subtree paths and method states are unnamed for five routes, and Task 6's traversal row proves the normalizer

- **Location:** plan:841-846 (R-5cii19), plan:1411-1416 (Task 5 routes),
  plan:1515-1517 (Task 6), `internal/serve/serve.go:95-96`.
- **Defect:** `Shell.Register` mounts both `<prefix>/<name>` and
  everything under `<prefix>/<name>/`. Task 5 states the exact-path rule
  for `login` only.
  - **Unnamed routes:** nothing says what `<prefix>/roster/x`,
    `<prefix>/members/`, `<prefix>/branding.css/x`, or
    `<prefix>/operator.css/x` answer. A trailing-slash `members` served
    to an admin computes its base one level too deep (R-5cii2).
  - **Unnamed methods:** the method rule is unnamed for the `login`,
    `_app`, and stylesheet routes. So is `_app`'s directory handling: an
    `http.FileServer` lists `<prefix>/_app/` and redirects `index.html`.
  - **The vacuous row:** Task 6's `<prefix>/operator.css/../x` row never
    reaches the route. `Normalize` refuses a `..` segment with 400
    first (`internal/serve/normalize.go`, rule 2). So the row proves the
    normalizer, and it contradicts its own criterion ("the terminator's
    404 or the same file").
- **Proposed fold:** state one rule for every document, asset, and
  stylesheet route. Only the exact path answers, any deeper path is the
  terminator's 404, and a directory under `_app` is the terminator's 404
  with no listing. Only GET and HEAD are admitted, and the refusal status
  is named. Give the roster API the same exact-path rule. Replace Task
  6's traversal row with `<prefix>/operator.css/x`, which must equal the
  terminator's answer.

### R4. MINOR: the consumption-route proofs use a synthetic `?token=` and drop the minted-link row

- **Location:** plan:1601-1603 (Task 7 token row), plan:1141-1165
  (Task 2), `internal/auth/token.go:55,64`.
- **Defect:** the link parameter is `t` (`linkTokenParam = "t"`), not
  `token`. A mutation that logs `r.URL.Query().Get("t")` passes a test
  that sends `?token=`. The 5c-i outline's Task 2 also carried a row that
  followed a link minted by the real request-link handler through the
  shell to the owned `login` stub. This plan drops it and places it
  nowhere. `linkPath` in `internal/auth` and the document route in
  `cmd/dubplate-server` are two constants in two packages, and nothing
  gate-resident binds them. If they drift, every login link answers 404
  while the gate stays green.
- **Proposed fold:** in Task 7, mint the link through the real
  `LinkURL` path and request it through `shell.Handler()`. Assert that it
  answers the `login/link` document with `no-store`, and that no
  captured log line contains the token. **Fixture state:** a log sink
  that captured at least one record for the request, so "no line
  contains" reads real lines.

### R5. MINOR, OWNER FORK: `Referrer-Policy: no-referrer` can make an old browser's same-origin POST carry `Origin: null`

- **Location:** plan:526-537 (R-5cii6), `internal/serve/origin.go`.
- **Defect:** the Fetch standard serializes a non-CORS POST's `Origin`
  as `null` under the `no-referrer` policy. A same-origin `fetch` is
  non-CORS (response tainting "basic"). `CrossOriginProtection.Check`
  (go1.27.1 source) falls back to Origin only when `Sec-Fetch-Site` is
  absent, then requires `url.Parse(origin).Host == req.Host`. `"null"`
  fails that check. So a browser without `Sec-Fetch-Site` (Safari before
  16.4, which covers iOS 15 devices a group plausibly still owns) could
  fail every POST once this pass adds the header: request-link, consume,
  and add. The go doc itself says Sec-Fetch-Site is "available in all
  browsers since 2023", so the exposure is old devices only. WebKit's
  exact serialization is unverified; a scratch probe settles it.
- **Options:** (a) keep `no-referrer` (the spec's value) and record
  old-browser login as an unsupported residual; (b) switch to
  `same-origin`, which keeps every cross-origin referrer empty. Under
  (b) the token URL never leaves the origin and the Origin header
  serializes normally. The one change is that same-origin asset and
  consume requests from the consumption page carry the token in
  `Referer` to dubplate itself, which logs no Referer.
- **Recommendation:** (a), with the residual written into the runbook
  and the close, unless the conductor's probe shows the refusal on a
  device the group actually uses.

### R6. MINOR: the rewrite's anchor has two unnamed states

- **Location:** plan:1399-1404 and 1418-1431 (Task 5).
- **Defect:** the rewrite inserts the tier 2 and 3 links "after the
  bundle stylesheet". Kit 2.70.3 renders an `ssr = false` shell with
  every branch node's stylesheets linked (`runtime/server/page/index.js`,
  the comment "provide nodes without running load functions so that the
  styles and fonts are linked in the head"). That means one link for the
  layout's CSS and one more for any page that imports its own. So the
  anchor can be absent or plural. Inserting after the first of several
  puts branding before a bundle sheet, and the bundle sheet then wins
  the tie. That is DL R4's tier order inverted, and 5c-ii has no
  computed-value half to see it.
- **Proposed fold:** add two state rows. Zero bundle stylesheet links
  refuses startup, naming the document. Several links place the
  insertion after the last. Assert in G7 that no `_app/` stylesheet link
  follows `branding.css`.

### R7. MINOR, OWNER FORK: a cosmetic configuration error stops every surface

- **Location:** plan:734-735 (R-5cii13), plan:755-759 (R-5cii14).
- **Defect:** a branding value outside the grammar, or an operator file
  that is missing at boot, refuses serve startup. Today that means
  login. At rung 12 it means the streamer fallthrough too: the whole
  music server goes down over a color typo, or over a file on a mount
  that is not ready yet. `doctor` does not judge these keys until 5d
  (R-5cii12). The security property (a refused value never reaches CSS)
  holds either way.
- **Options:** (a) refuse startup, as planned; (b) serve an empty
  branding stylesheet or no operator link, log at error level, and let
  5d's `doctor` row fail loudly.
- **Recommendation:** (a) for 5c-ii, since a restart surfaces the error
  at once and the box runs nothing yet. Carry the question to rung 12's
  plan, where the fallthrough makes availability the larger stake.

### R8. MINOR: two rung 12 carries need sharper wording now that the web writes the roster

- **Location:** plan:363 (privileged table, `streamer_username`),
  plan:1661-1667 (Task 8's gap list).
- **Defect:** this pass opens the first network channel that writes
  `streamer_username`, and rows written now persist into rung 12. 5c-i's
  gap 5 reads "a principal must not map a non-admin member onto a
  streamer admin account". A rung 12 check applied only at write time
  misses every row the API wrote earlier. Separately, the roster API's
  CSRF defense trusts any `Sec-Fetch-Site: same-origin` request. At rung
  12 the streamer's own UI is same-origin: the session cookie is
  stripped from requests going to the fallthrough, but script on the
  streamer's pages can still call `<prefix>/roster` with the admin's
  `__Host-` cookie. A script-execution bug in the streamer then becomes
  a roster write. Neither residual is in the spec's threat-model list.
- **Proposed fold:** Task 8's gap list states that rung 12 re-validates
  every existing row's principal, not only new writes. It also names
  the shared-origin residual (streamer script reaches the admin API) as
  a rung 12 threat-model entry.

### R9. MINOR: the privileged-fields table has no row for reading the roster

- **Location:** plan:357-370.
- **Defect:** this is planning-miss item 3. The roster's full address
  list, with admin flags and streamer usernames, is a privileged read.
  Its channels are the gate's GET, `member list` on the command line,
  and 5b-ii's kind unknown-address answer, which hands out one admin
  contact. The table covers only write authority.
- **Proposed fold:** add the row, with the refusal column naming the
  gate's 404 and the generic answer's setting.

### R10. MINOR: n5 falsifies a shipped comment the wall forbids fixing, and three of its rows are not proofs of the change

- **Location:** plan:1134-1137 and 1156-1165 (Task 2),
  `internal/auth/roster.go:647-649`, `internal/identity/identity.go:70`.
- **Defect:**
  - **The comment:** `actorRefusesRune`'s doc says it refuses "the
    same set [identity.NormalizeAddress] refuses". After n5 that is
    false, and Task 2's Files list does not include `roster.go`. So the
    implementer either stops or leaves a wrong comment.
  - **The rows:** `addressRefusesRune` already refuses all of `Cf`. So
    the U+200B, U+00AD, and U+E0001 rows pass before the change. They
    are regressions, not proofs, and the task's mutation evidence could
    over-count.
- **Proposed fold:** add `internal/auth/roster.go` (the comment only)
  to Task 2's Files, or have `actorRefusesRune` say what it refuses on
  its own. Label the three `Cf` rows as pre-existing in the test table.

### R11. MINOR: an undeclared deviation on where the CSP is served

- **Location:** plan:139-157 (the deviations list), plan:537, plan:1615
  (mutation 5).
- **Defect:** the spec's embed-path section says "Every served response
  carries `X-Content-Type-Options: nosniff` and `Referrer-Policy:
  no-referrer` beside the CSP". The plan serves the CSP on documents
  only, and mutation 5 forbids it on the gate's 404. The narrowing is
  right (a 404 or a JSON answer runs no script), but it is not on the
  approve-with-plan deviations list.
- **Proposed fold:** add it as an R-5cii item and to Task 8's errata.

### R12. MINOR: the embedded bundle is whatever `web/build` holds at build time

- **Location:** plan:611-635 (R-5cii9), plan:1661-1667 (Task 8 carries).
- **Defect:** `web/build/` is gitignored, and the binary embeds whatever
  is there. A binary built from a workstation carries a bundle built
  from the working tree, including uncommitted `web/src` edits. No
  reviewer has read that JavaScript. The CSP stays self-consistent,
  since it is extracted from the same bundle, so nothing fails.
- **Proposed fold:** carry to 5d: `deploy.sh` builds the bundle from a
  clean committed tree (`npm ci`, then the build) before `go build`, and
  refuses a dirty `web/`.

### R13. MINOR: the members 404 hides admin status, not the page's existence

- **Location:** plan:199-202 (Goal), plan:839-846, spec "The pages".
- **Defect:** with `ssr = false`, the members page component ships in
  public `_app/immutable/` chunks. The client router's manifest names
  `/members` to every visitor, and 5c-iii's page will name the roster
  API's path. So the gate's 404 keeps a non-admin from learning whether
  their own session is an admin's. It does not hide that the route
  exists. The API remains the real boundary, which the plan builds
  correctly.
- **Proposed fold:** say so in the 5c-iii handoff: nothing sensitive
  may live in the page component, and the API alone authorizes. Scope
  the Goal's "cannot tell from an unregistered path" to the server's
  answers.

## Notes, not findings

- **CSP and `data:` URIs.** `default-src 'self'` also governs `img-src`,
  so any `url(data:...)` a DaisyUI component or Vite's asset inlining
  emits in the built CSS is blocked. The stubs use no component, so
  5c-ii is unaffected. A G3 or G6 assertion that the built CSS carries no
  `url(data:` would move the conflict to the pass that can decide it,
  rather than letting it surface as a broken render in 5c-iii.
- **Gate log volume.** Every anonymous probe of `<prefix>/roster` or
  `<prefix>/members` writes a gate refusal line, where the terminator
  writes none. journald's rate limit bounds it. If it proves noisy, log
  the no-cookie reason at debug level.

## Planning-miss audit (the four items)

1. **Quoted behaviors.** Strong overall: `ParseForm`,
   `CrossOriginProtection`'s safe methods, `http.Error`'s headers, RFC
   6797, `go doc embed`, and the kit source are all quoted. Unquoted
   behaviors the plan relies on: `mime.ParseMediaType` in `ParseForm`
   (R2), `CrossOriginProtection`'s Origin fallback under the new
   Referrer-Policy (R5), `http.FileServer`'s listing and redirect
   behavior (R3), and kit's `ssr = false` stylesheet linking (R6).
2. **Named states.** The gaps: the media-type parameter (R2), deeper
   subtree paths and methods (R3), the rewrite anchor count (R6), and
   the operator path's file class (R1). The CSP extraction, branding
   grammar, theme keys, install state, and server-file tables are
   complete for their checks.
3. **Channels for privileged fields.** The read authority is missing
   (R9), and `streamer_username`'s persistence into rung 12 needs
   stating (R8). The admin flag, actor, channel, refresh cookie, theme,
   branding, CSP, and HSTS rows are complete and match the code (the
   flag is read from the roster record at `seam.go`; the actor is
   re-checked under the lock at `roster.go` `addTo`).
4. **Non-vacuous fixtures.** Strong: the gate refusal table's
   refreshed non-admin, the two-member roster for byte comparison, both
   HSTS configurations, G6's nonzero-script precheck, and the prefix the
   build never saw. Weak: the synthetic token parameter (R4), the three
   pre-existing `Cf` rows (R10), and Task 6's traversal row, which the
   normalizer answers (R3).

## Attacked and found sound (do not re-litigate in fold)

- **The gate's order.** The seam runs, then `Identity.Valid`, then
  `Admin`, then the refresh `Set-Cookie`, then the inner handler. The
  non-admin row's fixture (same-origin and old enough to refresh)
  catches a premature cookie. The seam reads `Admin` from the roster
  record, and `Identity.Valid` refuses `ChannelCommandLine`.
- **404 equality.** Both the gate and the terminator call
  `http.NotFound` and both pass through the new outermost header step,
  so the equality survives Task 7. Mutation 5 guards the CSP.
- **The closed field set over `r.Form`.** It covers the query string and
  repeated keys (malformed before the admin check). A capital `Admin`
  falls to `unknown_field`, and `admin_not_grantable` still has
  `AddMember`'s channel refusal behind it.
- **The actor and channel sources.** The actor comes from the gate's
  context value and is re-validated as an admin under the directory
  lock. The channel is a constant. `logAddition` already logs the
  normalized actor (5c-i n4).
- **Concurrency and partial writes.** The API adds a caller to 5c-i's
  sound path (the `flock` before `r.mu`, atomicfile, re-read, and
  `errRosterWriteLanded`). The 5-second `context.WithTimeout` makes
  `busy` reachable. `write_unconfirmed` is correctly distinct from
  `failed`.
- **HSTS gated on the configured scheme.** The ingress strip removes
  forwarded headers, so no request can switch it on. The go doc for
  `CrossOriginProtection` names HSTS as the mitigation for its own
  HTTP-to-HTTPS gap.
- **CSP extraction.** All four refusal states refuse startup, and G6
  hashes the served bytes independently, so a kit hash drift fails by
  name.
- **The branding grammar.** It refuses rather than escapes, and the
  injection table puts each bad value beside a valid pair. The operator
  can already write arbitrary same-origin CSS in tier 3, so the grammar
  guards against malformed configuration, not a hostile operator. That
  is the right weight.
- **Cache table.** `private` on the members document keeps an admin's
  refreshed `Set-Cookie` out of shared caches, and `no-store` covers the
  token document and JSON. A 404 sets nothing.
- **Path traversal into the embed.** `Normalize` refuses dot segments,
  encoded bytes, and backslashes before any route. `fs.Sub` with
  `io/fs` path validation cannot reach outside `build`. The `_app`
  subtree cannot reach a document (subject to R3's listing rule).
