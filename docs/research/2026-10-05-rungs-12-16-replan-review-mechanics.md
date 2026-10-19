# Review: rungs 12 and 16 re-plan, mechanics and architecture lens

Target: `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md` at `88c535e`.
Lens: does each stated mechanism behave as stated (verified against source), and does the result
stay clean, stdlib-shaped Go and idiomatic Svelte 5. Findings are correctness gaps ranked by
consequence, then over-ceremony ranked by cost. Polish is omitted.

Counts: 0 blocker, 7 major, 8 minor (one marked OWNER FORK), 3 over-ceremony.

## What holds

- The four `internal/serve` removals have no production caller. `cmd/dubplate-server/serve.go:324`
  is the only production `NewStrip` call: `serve.NewStrip("", auth.SessionCookieName, proxyConfig, serve.FrontExemption{})`.
  `SetFallthrough` and `addBearerBypass` are called only from tests.
- The end state is correct. `Shell.Handler` already substitutes `http.HandlerFunc(http.NotFound)` when
  no fallthrough is set (`internal/serve/serve.go:228-231`), so removing the slot leaves the same
  outward behavior.
- The dual-stack defect is real and goes with the type. `FrontExemption.headerFor` matches the raw
  peer string, with no parse and no unmap: `if _, trusted := e.proxies[peer]; !trusted` (`strip.go:263`).
- The username column rename is feasible in place. Only `0003_catalog.sql` names the column
  (`streamer_username TEXT NOT NULL UNIQUE COLLATE NOCASE`, line 5). Migrations 0004 and 0005
  reference `users (id)` only, and no non-test Go SQL reads the column yet.
- The store uses its own migration runner, not goose. Its ledger records each file's SHA-256 and
  refuses a mismatch (`internal/engine/store/migrate.go:237-238`: `ledger version %d records %s (sha256 %s), this binary has %s (sha256 %s)`).
  The spec's precondition ("no store file on the box") is therefore load-bearing, and correct to require.
- `internal/harness/analysis` and `internal/harness/cassette` do stay in use. `go list` shows
  `internal/harness/conform` importing both, and `cmd/harness`, `oracle`, `stub`, and `synth`
  importing cassette. In `cmd/dubplate-server`, only `shadow.go` and `shadow_test.go` import them.

## Findings

### M1 (major): the session-cookie removal leaves a hollow `Strip.sessionCookie` and a dead `NewStrip` parameter

Location: spec:101-102; `internal/serve/strip.go:279,300-310,318`; `internal/serve/serve.go:342-436`;
`cmd/dubplate-server/serve.go:299-304,324`.

Defect: the spec removes "the session-cookie wrapper around the fallthrough" but leaves the plumbing
that fed it. `Strip.sessionCookie` has exactly one reader, `stripSessionCookie`
(`serve.go:352`: `removeCookie(r, s.strip.sessionCookie)`). The `NewStrip` doc confirms the field's
sole purpose: "the strip itself never removes a cookie, and [Shell] reads this name to remove that one
cookie on the way out to the fallthrough" (`strip.go:304-306`). `removeCookie` and `withoutCookie`
(about 80 lines with their docs) then have no caller, and `unused` catches those two. It does not
catch the field. A keyed composite literal counts as a use, so `NewStrip(sessionCookie string, proxy ...)`
would survive the gate with a parameter that does nothing. The cmd side keeps a production
argument, `auth.SessionCookieName`, and a doc paragraph that becomes false: "The strip carries
auth.SessionCookieName, so the shell removes the server's own session cookie from any request it
hands the fallthrough" (`cmd/dubplate-server/serve.go:299-304`).

Proposed fold: name the full set in the removal list: `removeCookie`, `withoutCookie`,
`Strip.sessionCookie`, and `NewStrip`'s `sessionCookie` parameter. The strip's signature becomes
`NewStrip(proxy TrustedProxyConfig) Strip`. Add the cmd paragraph to the rewrite list. With the header,
exemption, and cookie parameters gone, `NewStrip` takes one argument. That shape is right, because
the strip's only construction input is the trusted-proxy config.

### M2 (major): `addBearerBypass` is generic hardening with a named consumer in an unchanged rung

Location: spec:104-105; `internal/serve/origin.go:194-248`;
`docs/specs/2026-09-11-configuration-homes-design.md:123-126`.

Defect: the spec counts the bypass among the "Navidrome-facing" pieces and assigns its rebuild to
slice 6. The configuration homes spec, which governs rung 14, depends on it: "The admin API is served
on two mounts, one cookie-authenticated for the web app and one bearer-authenticated for the TUI,
because rung 5's CSRF rule exempts bearer paths from origin checks and a bypassed path may not read the
session cookie." The spec lists rung 14 as unchanged (spec:53), so its premise loses its mechanism
without an erratum. The spec also says the rebuild is "keyed on `netip.Addr` where it matches peers".
The bypass matches request paths, not peers (`o.bypass.Handle(pattern, bypassMatch)`). Peer matching
belonged to `FrontExemption`.

The removal also drops two pieces of knowledge from the code. The first is the contract "it must not
read the session cookie to authenticate a request" (`origin.go:199-203`). The second is the reason
for registering twice: "either alone would leave one of the two refusing" (`origin.go:241-243`).
The in-process OpenSubsonic surface needs both. Native clients send neither `Origin` nor
`Sec-Fetch-Site`, so `allowed` refuses their POST at `origin.go:268-272`. Browser clients send
`Sec-Fetch-Site: cross-site`, so the stdlib refuses theirs. Slice 6 is also not necessarily the first
API slice to answer a POST. The streamer spec (rev 2.5, `:579-583`) names a JSON POST under the
`transcoding` extension, which arrives with slice 5.

Proposed fold, with a recommendation:
(a) **Recommended:** take the bypass off the removal list. Reshape its doc to name its consumers,
rung 14's bearer admin mount and the API surface, the same way the spec keeps the trusted-proxy config
as generic hardening (spec:213-214). The cost is the roughly 10 lines it already has, plus their test.
(b) Delete it, but name rung 14 or "the first slice that answers a non-GET on the API surface,
whichever lands first" as the rebuilder. Add an erratum to the configuration homes spec, and carry the
two rationale sentences above into streamer spec revision 3.
Either way, correct the `netip.Addr` sentence.

### M3 (major): the box roster cannot be migrated "through an existing verb or the bootstrap path"

Location: spec:152-154; `internal/auth/roster.go:340`; `cmd/dubplate-server/member.go:56-62`.

Defect: the loader is strict: `json.Unmarshal(data, &file, json.RejectUnknownMembers(true))`
(`roster.go:340`, under `encoding/json/v2`). Once the key is renamed, the old file fails to load in
`serve`, in `doctor`, and in `member add`. `member add` "creates the roster when it is absent or adds
to it when it exists" (`member.go:58-59`), and in both cases it loads the file first. No existing verb
can therefore read the old row. The bootstrap path requires the file to be absent, and removing it is
the hand edit the spec rules out.

Proposed fold: drop the requirement that this pass's deploy migrates the row (see O1). The serve unit
is not enabled until rung 16, and the box holds no store, so nothing reads the roster before then.
Route the step to rung 16's first-install runbook. That runbook removes the stale development roster
and then runs `member add --admin --username` (or 6d's `member import`, if 6d lands first and its
plan says so). The step then stays in the runbook, the repo's source of truth. If the box must carry
the renamed binary earlier, name a `deploy.sh` step that removes the file. Never ship a one-off
compatibility read.

### M4 (major): "generic hardening keeps its tests" is under-specified, because those tests probe through the fallthrough

Location: spec:114-115; `internal/serve/serve_test.go` (about 25 `SetFallthrough` calls);
`cmd/dubplate-server/serve_test.go:352,552,664`.

Defect: several hardening proofs use the fallthrough as their spy:
- `TestOriginCheckCoversSubtreeAndFallthrough` (`:988`)
- `TestEchoMethodsRefusedBeforeRouting` (`:929`, which calls `SetFallthrough(recordingHandler(&fallthroughGot))` at `:955`)
- `TestRefusalRecordOnlyForTheShellsOwnRefusals` (`:606`)
- `TestOrderRawVsDecodedRefusedBeforeAnyHandler` (`:123`) and `TestOrderPrefixRequiresSegmentBoundary` (`:160`)
- `TestRefusalLogsIdentityAlreadyGone` (`:184`)
- `TestResponseControllerFlushReachesRealWriter` (`:455`) and `TestReadFromReachesRealWriter` (`:680`)
- in cmd, `TestNewServeShell_StateChangingFallthroughRequestIsDenied` (`:546`) and `TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree` (`:629`, whose fallthrough sentinel proves the subtree was the route taken)

An implementer who swaps the spy for "expects 404" can leave a proof vacuous. Examples: a GET whose
assertion was "the fallthrough saw no identity header", or a Flush test with no handler left to flush.
This is the planning-miss class `CLAUDE.md` names: "each proof names the fixture state it needs ... so
the test cannot pass vacuously".

Proposed fold: add an acceptance line. Each test that used the fallthrough as a probe keeps its
assertion and is re-pointed at a registered spy subtree, with a "spy reached" check, or at a status
the 404 cannot also produce (403 or 405). A test is deleted only when its sole subject was the
fallthrough slot or the cookie removal: `TestSetFallthroughTwicePanics`, `TestEmptyFallthroughAnswers404`
(re-scoped to "outside the prefix answers 404"), the nine `TestSessionCookie*`, and
`TestNewServeShell_RemovesTheSessionCookieAtTheFallthrough`. The plan lists each test with its
disposition.

### M5 (major): the shadow deletion orphans more than the spec names

Location: spec:120-125.

Defect: the spec gives a general rule ("a package left with no importer outside its tests is either
deleted or recorded") and names four comment sites. These are the concrete orphans:

1. `setUnidecodeTransliterator` (`cmd/dubplate-server/unidecode.go:76-90`). Its only callers are
   `shadow.go:281` and `shadow.go:377`, and it is unexported, so `unused` fails the gate. Delete it;
   rung 9's scoring verb rebuilds it. The `unidecode install` verb and doctor's `unidecode` row stay,
   with rung 9 recorded as their consumer.
2. `internal/engine/tagio`. After the deletion it has **zero** importers, production or harness. Its
   only non-test importer is `internal/engine/shadow`. `tagio.Tags` exists for shadow alone: "Its first
   caller is rung 4's shadow reader" (`tagio/doc.go:33-34`). Keep the package, recording rung 9 as its
   consumer, because its golden generator is a gate leg. Make `Tags` unexported or remove it, under
   "nothing exported ahead of a caller", or record rung 9 as its caller.
3. `internal/harness/analysis`'s shadow-row shape. `AdjudicationSubject` carries "an album plus release
   MBID for a live shadow divergence rung 4 will write" (`adjudication.go:32-34`), alongside
   `MatcherCommit` and `MatcherConfigHash` "required on that shape" and the validator's shadow branch
   (`:210-230`). `corpus/analysis/adjudications.json` holds no `release_mbid` row. The package stays,
   but this half is a dead branch. Reshape it to case-only. That is the reshape ruling 6 asks for.
4. `matcherCommit`'s rationale (`main.go:40-52,153-158`; `Makefile:3-8`): "shadow observe REFUSES to
   run on "dev"". The stamp stays useful as a build identity. Rewrite the rationale.
5. `musicbrainz.ClassBackground` ("deferred work: rung 4's shadow, and later", `limiter.go:25`) and
   `ErrUnavailable`, whose `%w` justification is "Rung 4's shadow branches on it" (`client.go:78`),
   lose their only production consumer. Record rung 9 as the consumer, or prune.
6. About 20 further comment sites cite shadow: `config.go:28`, `keys.go:197`, `registry.go:213`,
   `auth/store.go:28,120`, `store/leaf.go:14`, `matcher/quiet.go:5`, `tag/tag.go:58,67`,
   `tagio/container.go:68`, `main.go:11`, `templatekeys_test.go:35`, and the doctor tests'
   `registerShadowIntegration` fixture (`doctor_test.go:104,1066-1110`).

Proposed fold: enumerate items 1 through 5 with a disposition each, and make item 6 a grep-clean
acceptance line (`git grep -i shadow` limited to CSS box-shadow and the deliberate "silently shadow"
idiom). Also record that `readFLAC` (`shadow/libraryread.go:24`), the FLAC-to-`matcher.ItemFields`
projection, is salvage material for rung 9's plan, reachable at the `pre-greenfield` tag or in history.

### M6 (major): the document pass omits the operator runbook and the example roster

Location: spec:168-189; `docs/server-runbook.md:138-164,615-627`; `config/roster.example.json:7,13`.

Defect: the runbook's bootstrap command is
`dubplate-server member add admin@example.org --admin --name "Ada Admin" --streamer-username ada`
(`:141`, and again at `:164`). After the flag rename, cobra refuses that command as an unknown flag.
The runbook's topology section describes "the streamer driver in a later" rung and a proxy in front
of the streamer (`:615-627`). That is the forward-auth front that ruling 4 retires. The example
roster an operator copies still says `"streamer_username"`. Ruling 7 makes these two files the
stranger's install path.

Proposed fold: add `docs/server-runbook.md` to the document changes (the flag in both commands, and
delete or rewrite the topology section) and `config/roster.example.json` to the rename list.

### M7 (major): the close omits `web-auth-security-reviewer`

Location: spec:157-166; `CLAUDE.md` "Reviewers": "`web-auth-security-reviewer` for the serve/JWT surface".

Defect: the pass edits the ingress strip, the origin check, and the cookie boundary, and it rewrites
those surfaces' tests (M4). That is where a security proof can silently go vacuous.

Proposed fold: add one `web-auth-security-reviewer` dispatch over the `internal/serve` and
`cmd/dubplate-server/serve*.go` diff. Ask it for correctness gaps ranked by consequence.

### M8 (minor): after the fallthrough goes, `Prefix.Route` and `terminated()` lose their stated reason

Location: `internal/serve/serve.go:233,322-340`; `internal/serve/prefix.go:123-157`.

Defect: with `next` always `http.NotFound`, both arms of `Prefix.Route` answer 404. The docs justify
the arrangement by the proxy: "an unknown owned path must stop here rather than reach the proxy at the
fallthrough" (`serve.go:328-329`), and "Rung 14 depends on this: an unknown API path under the owned
prefix must not reach the proxy" (`prefix.go:128-130`). The acceptance line "each remaining type,
function, and parameter earns its place" invites a fight at the architecture read.

Proposed fold, an architecture call: keep `Prefix.Route(owned, next http.Handler) http.Handler`. It is
a stdlib-shaped combinator. `Handler` passes `http.NotFoundHandler()` as `next`, and the docs are
rewritten to state the owned-versus-unowned decision on its own terms. Record in the spec that slice
8's UNI-H2 (`/` to login) arrives as a named root route the shell owns, never as a reopened
general-purpose fallthrough slot.

### M9 (minor): the in-place migration edit must re-pin the hash table, and the precondition's scope should be named

Location: spec:148-150; `internal/engine/store/schema_test.go:289-299`.

Defect: `migrationHashes` pins `"0003_catalog.sql": "f962133f..."` and exists to stop exactly this
edit: "an edit to a merged file ... then fails this table instead of every existing store's ledger
check at its next start". The gate will be red until the pin changes. Without a stated rule, an
implementer may "fix" the red gate with a new migration.

Proposed fold: the plan re-pins from an independent `sha256sum` run (the pin's own rule), and cites the
precondition evidence in the commit. The precondition check covers the store leaf and its
pre-migration copies under `/var/lib/dubplate/`, and confirms that the backup remote holds no store
snapshot. Otherwise a later restore brings back a ledger with the old sum.

### M10 (minor): the rename list is incomplete, and it leaves the API outcome codes undecided

Location: spec:141-146.

Defect: unlisted sites include:
- `web/scripts/rendered/run.ts` (20 references, including the `#add-streamer-username` and `streamer-hint` ids)
- `member.go:177`'s `STREAMER USERNAME` column
- the test files in `internal/backup` and `cmd/dubplate-server` that insert users rows
- the roster API's outcome codes `duplicate_streamer_username`, `invalid_streamer_username`, and
  `missing_streamer_username` (`admin.go:37-42`), which `deck.ts:149-155` and `+page.svelte:46-53`
  read as a wire contract

The gate forces most of these. The outcome codes are a choice.

Proposed fold: per ruling 3 ("everywhere"), the outcome codes become `duplicate_username`,
`invalid_username`, and `missing_username`. The `duplicate_principal` copy loses "streamer". The plan's
Files list names `run.ts`.

### M11 (minor): `statusCapture`'s proxy rationale is a trap for the rewrite

Location: `internal/serve/serve.go:438-496`; `strip.go:520-530`.

Defect: the docs justify `ReadFrom` by "every proxied stream rung 12 serves", and justify `Hijack`
and group 3's `Upgrade` carve-out by the proxy's websocket. A reshape that reads these as
proxy-only could delete `ReadFrom`. Slice 5's streaming through `http.ServeContent` reaches `io.Copy`,
which consults the destination's `io.ReaderFrom` to use sendfile.

Proposed fold: keep `ReadFrom`, `Flush`, and `Unwrap`, with their rationale restated for in-process
streaming. `Hijack` and the `Upgrade` carve-out cost nothing to keep. Restate them as generic
transparency and do not chase them.

### M12 (minor): the exemption threads through more than the three helpers named

Location: spec:98-99; `strip.go:381-392,465-469,526-530,546-549,573`.

Defect: the exemption also runs through `stripConnectionListed` and `trailerScrubbingBody.exempt`.
`collapseToLast` exists only for the exemption, so it is deleted outright rather than "threaded through".

Proposed fold: correct the list. The compiler and `unused` force the code. The doc sentences at
`:526-530` need the explicit mention.

### M13 (minor): the doc acceptance is scoped to `internal/serve`, but the cmd wiring carries the same stale prose

Location: spec:111-113; `cmd/dubplate-server/serve.go:279-281` ("It registers no bearer bypass"),
`:299-304` (cookie at the fallthrough), `:309-312` ("No front exemption is in force");
`TestNewServeShell_StripCarriesNoFrontExemption` (`serve_test.go:1896`).

Proposed fold: extend the acceptance line to `cmd/dubplate-server`'s serve wiring and test names. The
test's property, that a trusted proxy is still identity-stripped, stays, under a name that does not
cite a removed concept.

### M14 (minor): "each rebuilt only by the streamer slice that first consumes it" is inaccurate

Location: spec:96.

Defect: under ruling 4, `FrontExemption`, the driver header, and the fallthrough cookie removal have
no future consumer. Only the bypass has one (M2).

Proposed fold: "Three are deleted for good. The bearer bypass is handled as M2 decides."

### M15 (minor, OWNER FORK): the URL a member types into a music app

Location: `cmd/dubplate-server/serve.go:166-172`; `internal/serve/prefix.go:16-21`; streamer spec `:573-575`.

Defect: the default `/dubplate` prefix was chosen because it "collides with none of Navidrome's own
surface (/rest, /share, /app, /auth)". The one-owned-prefix rule existed so that dubplate could share
a host with a streamer that owned the root. The streamer spec mounts the API at "a sub-prefix of the
one prefix", so a member's music app gets a base URL such as `https://host/dubplate/...`. The spec's
copy ruling (spec:155) tells a member about the username they type, and the server URL is the other
string they type. Removing the fallthrough deletes the only root-level slot.

Options:
(a) Keep the sub-prefix mount, as the streamer spec has it.
(b) Serve the API at the root `/rest` through a named root route (the M8 mechanism).
(c) Let slice 1's study measure roster clients against a sub-path base URL, then decide at slice 3.

Recommendation: (c). Nothing changes in the removal pass. Add one row to "Where the carried items
land": "API base URL (sub-prefix or root): slice 1 measures, slice 3 decides".

## Over-ceremony

### O1: deploying the removal pass to the box (highest cost)

The pass's deploy (spec:153) exists only to carry the roster row across. Nothing on the box reads
the roster or the binary's serve surface before rung 16: the serve unit is not enabled, and no store
exists. The deploy costs a quiet-box check and the M3 mechanics, and it buys nothing. Fold: no deploy.
The next deploy is whichever pass first needs the box.

### O2: two web reviewers on a label rename (low cost)

`svelte-reviewer` and `daisyui-a11y-reviewer` on the members page review a field rename, a hint
string, and changed `aria-describedby` ids. The repo rule fans out by surface, so keep both. They can
share one brief, scoped to the rename diff, rather than reviewing the whole page.

### O3: the rename lands in code that 6d replaces (moderate cost, recorded, not a change)

6d moves the roster into the `users` table and touches every rename site, so its plan, which still
awaits approval, needs a rename sweep either way. Keeping the rename in this pass is still the right
order: the in-place migration edit has to land before any store exists, and 6d is the first pass to
write one. The finding records the double touch so 6d's plan revision budgets for it.
