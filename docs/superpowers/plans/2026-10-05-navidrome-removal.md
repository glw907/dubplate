# Navidrome removal pass Implementation Plan

Status: **revision 1, 2026-10-05, drafted; not yet reviewed.** It carries no ruling for Geoff: every fork it
touches was ruled in the spec. It takes eight readings of its own ("Readings this plan takes"), each flagged for
plan review.

**Spec:** `docs/superpowers/specs/2026-10-05-rungs-12-16-replan-design.md`, revision 2.1 at `33b465c`, approved by
Geoff 2026-10-05 with all three owner rulings ruled yes ("the spec"). Its "The removal pass" section and its four
subsections are this plan's scope.
**Fold record:** `docs/superpowers/research/2026-10-05-rungs-12-16-replan-fold.md`.
**Parallel document pass:** `docs/superpowers/plans/2026-10-05-replan-document-pass.md` (revision 2, `1c37e40`),
which shares the main tree. Its "Coordination with `navidrome-removal`" section binds this plan too (below).
**Out of this pass, by the spec:** the archive-preservation chore is the document pass's Task 0 ("Preserving the
archive (now)" orders it before any other box work); the runbook's forward-auth topology section, the roster step,
`IDEAS.md`, `ROADMAP.md`, `CLAUDE.md`, and every spec erratum are the document pass's; the
architecture-cleanup inventory (the identity-source default, `doctor`'s flag chain) stays with the cleanup pass.
**No deploy.** The serve unit stays disabled and no store exists on the box (spec, "The removal pass").

**Pass class:** `tool` (Go: `go-conventions`, an Opus reviewer, a `go-architecture-reader` per touched package).
Task 1 is `auth-data`: it edits the ingress strip, the origin check's surroundings, and the session cookie's last
remover, so its re-pointed tests take a mutation proof each. Tasks 4a and 4b are `sweep` (a behavior-preserving
rename with grep post-conditions). The close runs the union: `tool`'s architecture reads, `auth-data`'s
`web-auth-security-reviewer` and live auth smoke, and `sweep`'s spot captures (the members page).
**Executor:** `dubplate-implementer`, `model: sonnet`, effort `high`, for Tasks 1 to 4b and Task 5's ledger
edits. No task upshifts: the spec names every symbol each task removes or renames, and the plan names each test's
disposition, so no task carries novel logic. A second `fix` verdict on Task 1 is the one upshift trigger the
conductor should weigh first (`model: opus`), since its failure mode is a vacuous re-pointed test.
**Task 0** is the conductor's, through the `dependency-upgrade` skill (implementers carry no Skill tool).
**Per-task gate:** `cairn-run-gate 'bash /var/home/glw907/Projects/dubplate/scripts/check.sh'` (the requested
`cairn-run-gate 'bash scripts/check.sh'`, with the absolute path because an agent's working directory resets),
re-issued on exit 75 until it prints `gate exit:`. Heavy lane: the web leg launches Chromium. Never
`CAIRN_GATE_LANE=light`. Every commit lands on master (below), so every commit takes the full form, and
`pass-execute`'s `reducedGate` is set to the same full string.
**Review chain:** `diff-reviewer` (`claude-opus-5-5`) per task against that task's acceptance criteria, at the
task's class bar. At the close: one `code-simplifier:code-simplifier` dispatch over the changed Go, TypeScript, and
Svelte, pasting `.claude/instructions/simplifier-brief.md` verbatim, with one `diff-reviewer` read of its diff; one
`go-architecture-reader` per touched Go package (list in Task 5); `web-auth-security-reviewer` over the
`internal/serve` and `cmd/dubplate-server/serve*.go` diff, asking for correctness gaps ranked by consequence,
vacuous re-pointed tests first; `svelte-reviewer` and `daisyui-a11y-reviewer` on one shared brief scoped to the
rename's `web/` diff, passing `docs/superpowers/specs/2026-09-24-visual-identity-design.md` and asking for
conformance to it (the copy register above all) graded beside correctness. Every reviewer dispatch asks for
correctness gaps ranked by consequence, not polish; a fold may refuse a finding with a one-line reason.
**Ceiling:** 8.0M execution tokens, flag 6.4M (80%). Planning and plan review are recorded outside it.
**Checkpoint interval:** four tasks (the default), met at each segment boundary (three tasks apart) and the close.
STATUS is written at each.
**Segments:** Task 0 (conductor, before any segment). A: Tasks 1, 2a, 2b. B: Tasks 3, 4a, 4b. Then the close's
reviews and Task 5. Every boundary lands on a green commit.
**Execution mode:** seven implementer tasks, so each segment runs through `pass-execute` by name, one invocation
per segment, sequential, with `implementer: "dubplate-implementer"`, `reviewer: "diff-reviewer"`, `maxFix: 1`,
`stopOnEscalate: true`, `passClass: "tool"`, each task's own `passClass` where it differs, `gate` and
`reducedGate` both the full string above.
**Independence:** Task 4a is file-independent of Tasks 1 to 3 and 4b. Task 2b depends on 2a (it prunes what only
`shadow` used). Task 3 follows 2a (both edit `doctor_test.go`). Task 1 precedes 2a (both edit
`cmd/dubplate-server/serve.go` and `serve_test.go`). Task 4b follows 1 and 3 (it edits `internal/auth`). All tasks
run one at a time in the main tree regardless: the gate runs over `./...` and `CLAUDE.md` forbids a second
executor in one tree. Independence only frees the conductor to reorder 4a around a stalled task.

## Where it runs, and the shared tree

Lane 1, the main tree, committing straight to master, as the lane registry's lane 1 row says and as the document
pass plan assumes. No branch and no merge ritual: there is no second tree to merge from. (`pass-core` prefers a
branch; the repo's lane practice and the document pass's coordination rules both key on master, and a branch
would split the shared tree's history for no isolation gain.)

The document pass's "Sharing the main tree" rules hold for this pass in reverse:

- A segment here does not launch while a document segment is in flight. At each launch, the last document task
  is committed and `git status` shows no warm change outside this pass's files, except the files below.
- The tree carries warm files this pass did not author (at plan time: `.claude/agent-memory/dubplate-implementer/*`
  and two untracked `docs/superpowers/research/2026-09-04-*.md` files). No task stages, edits, or deletes them.
  Each task stages only its named files, never `git add -A`.
- A red gate whose failing leg sits in a document-lane file is not a task failure here: the implementer reports
  it and stops without committing.
- The slice 2 lane works in its own worktree. Its `kin-openapi` addition to `go.mod` rebases onto Task 0's
  commits (spec ruling 5), which land on master first.

## Readings this plan takes

Each is a method or placement call, decided here and flagged for plan review.

1. **The exported-unused criterion is read as "the removal leaves no hit of its own".** The spec's "The staged
   exported-unused leg reports no hit in `internal/serve`" cannot hold as written. Run in isolation at `1c37e40`,
   the leg reports nine `internal/serve` names: `NewFrontExemption` (this pass deletes it) and eight unrelated to
   Navidrome (`ErrInvalidBindAddr`, `ProxyMode`, `RewriteDocument`, `ThemeCookieName`, `ErrInvalidPrefix`,
   `Owns`, and `statusCapture`'s `Flush` and `Hijack`). `Flush` and `Hijack` satisfy `http.Flusher` and
   `http.Hijacker` and cannot be unexported, which the name-based leg cannot see; the spec keeps both. The other
   six predate the re-plan and belong to the architecture-cleanup inventory, which the spec puts out of scope.
   Task 1's criterion is therefore: the leg's `internal/serve` list after Task 1 is the base list minus
   `NewFrontExemption`, with no new name.
2. **`musicbrainz.ClassBackground` and `ErrUnavailable` are pruned.** The spec keeps them "only if rung 9's spec
   names a consumer". No rung 9 spec exists at HEAD (`ls docs/superpowers/specs/` has none), so none names one.
   Rung 9 re-adds either with its consumer in the same commit, the simplifier brief's pattern 4.
3. **`config.IntegrationClass.Optional` and the registry's Optional-threshold pairing skip are removed as a shadow
   orphan the spec's list missed.** The shadow class is the only production class declaring `Optional`
   (`backup.go:32-39` and `proxy.go:25-33` declare none), and `registry.go:205-215` names the shadow class as the
   pairing's reason. Ruling 6 ("a removal reshapes what remains") governs. If the task finds a second production
   use, it keeps both and reports.
4. **The runbook's `--streamer-username` lines are renamed here.** The rename's grep acceptance names
   `docs/server-runbook.md`, and a runbook documenting a flag that no longer exists is a defect at the commit that
   removes it. The document pass owns everything else in that file (the roster step, the topology section), per
   its plan's file territory.
5. **The beets-migration `IDEAS.md` entry is the document pass's.** Its plan places it (its Task 8, after the
   merge). This pass's HISTORY entry records the pre-deletion commit the entry cites.
6. **A comment-only package still takes a `go-architecture-reader`.** Ruling 6 says every touched package is
   graded. A package touched only in `_test.go` files takes none, since the reader grades non-test source.
7. **The nonce sweep's "driver" wording is rephrased** (`cmd/dubplate-server/serve.go:293-295` and `:497`, the
   "interval driver"), so the spec's prose grep over `serve*.go` holds without an exemption list. It is a
   different sense of the word, and one word is cheaper than a standing exemption.
8. **Two `0003` preconditions were read at plan time** (below), before the document pass's archive chore ran.
   Both reads were read-only listings. Task 4a's re-confirmation runs only after that chore's STATUS record
   exists, the order the document pass requires.

## Behaviors this plan relies on, quoted

| Behavior | Quote | Where relied on |
| --- | --- | --- |
| `go doc net/http NotFoundHandler` | "NotFoundHandler returns a simple request handler that replies to each request with a “404 page not found” reply." | Task 1: `Handler` passes it as `Prefix.Route`'s `next` |
| `go doc io Copy` | "If src implements WriterTo, the copy is implemented by calling src.WriteTo(dst). Otherwise, if dst implements ReaderFrom, the copy is implemented by calling dst.ReadFrom(src)." | Task 1: `statusCapture.ReadFrom`'s restated rationale |
| `net/http/fs.go:430` (Go 1.27.1) and `io/io.go:363-364` | `io.CopyN(w, sendContent, sendSize)`; `CopyN` is `written, err = Copy(dst, LimitReader(src, n))` | Task 1: `http.ServeContent` reaches `io.Copy` with a `LimitReader` source, which has no `WriteTo`, so the writer's `ReadFrom` is consulted |
| `go doc net/http CrossOriginProtection.AddInsecureBypassPattern` | "AddInsecureBypassPattern permits all requests that match the given pattern. ... Only requests that match the pattern directly are permitted." "AddInsecureBypassPattern panics if the pattern conflicts with one already registered" | Task 1: the bypass stays path-keyed; its doc keeps "registers twice" |
| `internal/serve/origin.go:103-116` | `allowed` returns true for GET, HEAD, and OPTIONS; for any other method with neither `Sec-Fetch-Site` nor `Origin` it returns whether `o.bypass` resolves `bypassMatch` | Task 1: a POST outside the prefix carrying neither header answers 403 before routing |
| `go doc encoding/json/v2 RejectUnknownMembers` | "RejectUnknownMembers specifies that unknown members should be rejected when unmarshaling a JSON object." | Task 4b: `parseRoster` (`internal/auth/roster.go:340`) refuses a file carrying `streamer_username` with `ErrRosterUnparseable` |
| `go doc encoding/json Decoder.DisallowUnknownFields` | "DisallowUnknownFields causes the Decoder to return an error when the destination is a struct and the input contains object keys which do not match any non-ignored, exported fields in the destination." | Task 2b: `readJSON` (`internal/harness/analysis/ledger.go:19-20`) refuses a row carrying a removed shadow field |
| `go doc github.com/spf13/cobra Command.MarkFlagRequired` | "causes your command to report an error if invoked without the flag." | Task 4b: `member add` without `--username` |
| `internal/config/config.go:311`, `:315`, `:338` | `"%s: reserved for %s, arrives at rung %d"`; `"%s: %s"` (a superseded key's message); `"%s: unknown configuration key"` | Tasks 2a and 3: a deleted row refuses as unknown; a retired row refuses with its message. No key moves into the required set |
| `internal/engine/store/migrate.go:237-238` | `if name != m.name \|\| sum != m.sum { return fmt.Errorf("ledger version %d records %s (sha256 %s), this binary has %s (sha256 %s)", ...` | Task 4a: any store migrated under the old `0003` fails loudly at open |
| `internal/engine/store/schema_test.go:289-300` | `migrationHashes` pins each file's SHA-256, "computed once by an independent sha256sum run over the merged file, never by this package's own readMigrations" | Task 4a: the `0003` re-pin |
| `internal/engine/store/inspect.go:19-21` | "StateAbsent is a state directory with no store yet: no leaf, or a leaf holding no database file and no copy in snapshots/." | The `0003` precondition: no `store` leaf means doctor's `store` row reads `absent` |
| SQLite `NOCASE` (sqlite.org/datatype3.html) | "the 26 upper case characters of ASCII are folded to their lower case equivalents before the comparison is performed." | Task 4a: `users.username` keeps `NOT NULL UNIQUE COLLATE NOCASE` |

## The `0003` preconditions, read at plan time

Read-only ssh to the box (`5.78.148.42`), 2026-10-05, at plan time:

- `ls -la /var/lib/dubplate/` lists `auth`, `roster`, and `unidecode`; `ls /var/lib/dubplate/store` reports "No such
  file or directory". No leaf, so doctor's `store` row reads `absent` (`inspect.go:19-21`).
- `grep -oE '^(DUBPLATE_BACKUP[A-Z_]*|RCLONE_CONFIG_[A-Z_]*)=' /etc/dubplate/server.env` prints nothing (names
  only, no values), and `systemctl list-timers --all | grep -ci dubplate` prints `0`. Backup was never enabled, so
  no remote holds a store snapshot.
- `dubplate-server-serve.service` is installed and `disabled`.

Task 4a re-confirms both before its commit (conductor, read-only, after the archive chore) and cites the evidence in
its commit message. If either fails, Task 4a becomes an amending migration `0006` and the in-place edit is not made.

## Privileged fields and their channels

| Field | Every channel that may set it after this pass | Never |
| --- | --- | --- |
| An identity header (group 1, `identityExtras`) reaching a handler | none: the strip deletes every spelling from every peer, in headers and trailers | a trusted proxy (trusted for the client address alone), a configured front (the exemption is deleted), a driver header (deleted) |
| A forwarded-address header (group 2) reaching a handler | a peer named by the trusted-proxy class (`DUBPLATE_TRUSTED_PROXY_*`), unchanged | any other peer |
| The session cookie | minted by `internal/auth`'s handlers only; read by owned handlers | a bypassed handler (the bearer bypass's contract) |
| An origin-check bypass pattern | `originCheck.addBearerBypass`, package code only; no production caller today (`git grep addBearerBypass` hits `origin.go` and tests) | configuration, a request |
| A member's admin flag | `member add --admin` (owner-run CLI) and the roster file on disk | the roster API (`admin_not_grantable`), unchanged |
| A member's username (the principal's source) | `member add --username`, the roster API POST's `username` field (admin session), and the roster file on disk; `users.username` has no production writer until 6d | any other form key (`unknown_field`) |

## Task 0: The dependency sweep and the baseline count

Conductor. Before segment A. **0.8 hooks:** neutral.

**Outcomes:**
- The `dependency-upgrade` skill's sweep over `go.mod`, `web/package.json`, `tools/venv/requirements.txt`, and
  the pinned binaries (rclone, golangci-lint, Playwright), per `CLAUDE.md` "Current dependencies". Minors and
  patches land as their own gate-green commits on master; a major waits for Geoff and is recorded with its unblock
  condition. A sweep with nothing to move is one line in the pass record. The last sweep was
  `docs/superpowers/research/2026-10-04-dependency-sweep.md`.
- The spec's baseline (rung 12, "The streamer's configuration"): the count of values a default install's operator
  supplies by hand, at the commit after the sweep. Method: build the binary; run `dubplate-server check` with an
  empty environment; supply every key it reports as "required setting unset"; repeat until it reports none. The
  count is the number of keys supplied, recorded with their names. At plan time two rounds reported five then nine
  keys (fourteen so far, not run to the end).
- Report to the slice 2 lane's conductor that the sweep is on master.

**Acceptance:** the pass record names the sweep's outcome (each bump, or "nothing to move") and the count with its
key list and commit.

## Task 1: `internal/serve` loses the front, the fallthrough, and the driver header

Pass class: `auth-data`. Model: `sonnet`, effort `high`. **0.8 hooks:** neutral. Hook 6's fallthrough half
retires; nothing is built. The bearer bypass and `statusCapture`'s streaming methods stay for in-process
consumers.

**Files:**
- Modify: `internal/serve/strip.go`, `serve.go`, `doc.go`, `prefix.go`, `normalize.go`, `origin.go`, `probe.go`,
  `server.go`, and their tests (`strip_test.go`, `serve_test.go`, `prefix_test.go`, `probe_test.go`,
  `headers_test.go`, `server_test.go`)
- Modify: `cmd/dubplate-server/serve.go`, `serve_test.go`, `web_test.go` (the `NewStrip` calls at `:439`, `:641`,
  `:772`)

**Outcomes:**
- Deleted: `FrontExemption`, `NewFrontExemption`, `headerFor`, `collapseToLast` (`strip.go:233-267`, `:381-392`);
  the exemption threaded through `stripIdentity`, `deleteIdentity`, `scrubTrailerOnRead`,
  `trailerScrubbingBody.exempt`, and `stripConnectionListed` (`strip.go:353-379`, `:416-498`, `:520-578`).
- Deleted: `Shell.SetFallthrough`, `fallthroughHandler`, `fallthroughSet`, `stripSessionCookie`, `removeCookie`,
  `withoutCookie` (`serve.go:33-34`, `:137-153`, `:288-376`), `Strip.sessionCookie`. `Handler` passes
  `http.NotFoundHandler()` as `Prefix.Route`'s `next`. `Prefix.Route` stays; its doc states the owned-versus-unowned
  decision on its own terms, and a later root route arrives as a named route the shell owns.
- Deleted: the driver `header` parameter and `Strip.header`. `NewStrip(proxy TrustedProxyConfig) Strip` takes one
  argument. Group 1's name set no longer takes a parameter: `identityNames`, `identitySpellings`, `identityKeys`,
  `identitySpellingPresent`, and `probeIdentityNames` (`strip.go:205-230`, `:405-414`; `serve.go:524-541`;
  `probe.go:32-37`) reshape to a set derived once, so no caller passes `""`.
- `originCheck.addBearerBypass` stays. Its doc names its consumers by function, never by rung (the admin API's
  bearer mount, and the OpenSubsonic API's non-GET requests), and keeps its two contract sentences: a bypassed
  handler must not read the session cookie, and the pattern registers twice because either registration alone
  leaves one form refusing. It states that it is path-keyed.
- `statusCapture`'s `ReadFrom`, `Flush`, `Unwrap`, `Hijack`, and group 3's `Upgrade` carve-out stay, their
  rationale restated for in-process streaming (`http.ServeContent` reaches `io.Copy`, quoted above).
- Every comment in `internal/serve` and `cmd/dubplate-server/serve*.go` describes the package as it stands: the
  package doc's composition order and extension point (`doc.go:1-54`), `Prefix`'s and `Terminator`'s docs
  (`prefix.go:16-21`, `:123-135`), `Normalize`'s rules 1, 2, and 4 (`normalize.go:25-27`, `:46-68`),
  `defaultWriteTimeout` (`server.go:58-66`), `identityExtras` (`strip.go:11-14` and the Navidrome bullet at
  `:33-37`; Remote-User keeps its Authelia citation), `newServeShell`'s doc (`cmd/dubplate-server/serve.go:299-312`),
  and `servePrefixThreshold` (`:169-175`), which makes no claim about a sibling server. The nonce sweep's
  "interval driver" wording is rephrased (Reading 7). No new comment carries a rung, task, or plan citation.

**Test dispositions** (the spec requires one per test; every re-pointed test keeps its assertion):

| Test | Disposition |
| --- | --- |
| `internal/serve`: `TestSetFallthroughTwicePanics` | Delete: its sole subject is the slot |
| `TestSessionCookieRemovedAtFallthroughPresentAtOwnedPrefix` | Keep only its owned half, renamed to state it: the session cookie reaches a registered subtree intact. The fallthrough half is deleted (sole subject: the removal) |
| `TestSessionCookieOtherCookiesSurviveAtFallthrough`, `TestSessionCookieRawHeaderSurvivesVerbatim`, `TestSessionCookieDuplicateNamedCookieBothRemoved`, `TestSessionCookieRemovedFromEveryCookieLine`, `TestSessionCookieRemovedFirstInLineLeavesNoLeadingSpace`, `TestSessionCookieOnlyCookieOnALineDropsTheLine`, `TestSessionCookieRemovedWhenTheNameCarriesSpace`, `TestSessionCookieSpacedNameLeavesTheOthersReadable` | Delete: each one's sole subject is the cookie removal |
| `TestOrderStripBeforeFallthrough` | Delete: no handler exists outside the prefix, and `TestOrderStripBeforeOwnedSubtree` keeps the order assertion |
| `TestOrderRawVsDecodedRefusedBeforeAnyHandler` | Re-point: the outside-prefix row asserts 400 (a 404 cannot produce it); the owned spy's "never invoked" check stays |
| `TestOrderPrefixRequiresSegmentBoundary` | Re-point: the owned spy is never invoked and the status is 404; the positive "reaches `next`" half stays proven by `prefix_test.go`'s `Route` table, which passes its own `next` spy |
| `TestRefusalLogsIdentityAlreadyGone` | Re-point: drop the fallthrough recorder; the 400 and the record assertions stand |
| `TestEmptyFallthroughAnswers404` | Re-scope and rename: a path outside the prefix answers 404 |
| `TestResponseControllerFlushReachesRealWriter` | Keep the owned half; delete the fallthrough half (the owned half proves the same writer) |
| `TestReadFromReachesRealWriter` | Re-point at a registered spy subtree under the prefix; the `used` and body assertions stand |
| `TestUnclaimedOwnedPathReachesTheTerminator` | Re-point: drop the fallthrough recorder; the 404 and owned-not-invoked checks stand |
| `TestEchoMethodsRefusedBeforeRouting` | Re-point: the two outside rows keep 405 and "no handler reached" against the owned spy |
| `TestOriginCheckCoversSubtreeAndFallthrough` | Rename (subtree and outside the prefix); rows unchanged, since 403 and 404 already separate the check from the destination |
| `TestIdentitySpellingPresentReadsKeyPresence`, `newTestShell`, every `NewStrip("", ...)` call | Signature only |
| `strip_test.go`: `TestStripWrapExemption`, `TestStripWrapEmptyExemptionExemptsNothing`, `TestStripWrapExemptedHeaderSurvivesConnectionListing`, `TestStripWrapExemptionKeepsOnlyTheExactSpelling` | Delete: each one's sole subject is the exemption or the driver header |
| `TestStripWrapRemovesEveryIdentitySpelling`, `TestStripWrapRemovesMixedSeparatorSpellings` | Keep; drop the driver header's entries from `expectedIdentitySpellings` and the "driver's configured name" row |
| `prefix_test.go`'s `terminalFallthrough`, `headers_test.go:13`, `server_test.go:57` comments | Rename to name `next` or "outside the prefix" |
| `cmd/dubplate-server`: `TestNewServeShell_RemovesTheSessionCookieAtTheFallthrough` | Delete: its sole subject is the removal |
| `TestNewServeShell_StateChangingFallthroughRequestIsDenied` | Re-point and rename: a cookie-carrying POST outside the prefix with neither header answers 403 (a 404 cannot produce it); the spy goes |
| `TestNewServeShell_MintedLinkRoutesToRegisteredLoginSubtree` | Re-point: drop the fallthrough sentinel; the 200 from the login subtree is a status a 404 cannot produce. Its fixture-state comment is rewritten |
| `TestNewServeShell_StripCarriesNoFrontExemption` | Rename to its surviving property: a trusted proxy is still identity-stripped. Body unchanged |

The generic hardening keeps every test it has: the trusted-proxy configuration, the origin check (the bearer bypass
included), and group 1 and group 2 stripping.

**States a request meets after the task, and its answer:**

| Request | Answer |
| --- | --- |
| GET outside the prefix | 404 from `http.NotFoundHandler()`, no handler invoked |
| POST outside the prefix, neither `Sec-Fetch-Site` nor `Origin` | 403 from the origin check, logged as the shell's refusal |
| TRACE or TRACK anywhere | 405, logged |
| escaped and decoded paths disagree, inside or outside | 400, logged |
| owned, unclaimed | 404 from `Terminator`, not logged |
| owned, registered | the subtree, with every identity spelling stripped and the session cookie intact |

**Acceptance criteria:**
- `go doc -all ./internal/serve | grep -inE 'front ?exemption|forward-auth|driver|fallthrough|exemption'` prints
  nothing. The same grep over `cmd/dubplate-server/serve*.go` prints nothing outside rung and history citations in
  tests. Generic "in front of" prose stays.
- `git grep -nE 'FrontExemption|SetFallthrough|removeCookie|withoutCookie|collapseToLast|sessionCookie' -- internal/serve cmd/dubplate-server`
  returns nothing.
- `go doc ./internal/serve NewStrip` prints `func NewStrip(proxy TrustedProxyConfig) Strip`.
- The exported-unused leg's `internal/serve` list is the base list minus `NewFrontExemption` (Reading 1).
- Every row of the dispositions table is done as written; the report lists each test by name with its disposition.
- **Named mutations, each reported failing:** swap `http.NotFoundHandler()` for a handler answering 200 (the 404
  re-scope and the origin check's same-origin outside row fail); drop `s.origin.wrap` from `Handler` (both 403
  outside rows fail); drop `refuseEchoMethods` (the outside TRACE row fails); delete `statusCapture.ReadFrom` (the
  re-pointed `TestReadFromReachesRealWriter` fails); skip group 1's deletion of `Remote-User` (the renamed
  trusted-proxy test fails). **Fixture state:** each re-pointed probe asserts its spy was reached, or asserts a
  status a 404 cannot produce, so none passes against a request that never routed.
- The gate green.

## Task 2a: The shadow program leaves

Pass class: `tool`. Model: `sonnet`, effort `high`. Depends on Task 1. **0.8 hooks:** neutral.

**Files:**
- Delete: `cmd/dubplate-server/shadow.go`, `shadow_test.go`, `internal/engine/shadow/` (with `testdata/`)
- Modify: `cmd/dubplate-server/main.go` (`:11`, `:39-51`, `:112-113`, `:143`, `:152-158`), `main_test.go`
  (`:40-50`), `doctor_test.go` (`:95-110`, `:1066-1113`), `serve.go` (`:68-73`, `:182-187`), `serve_test.go`
  (`:252-280`), `templatekeys_test.go` (`:35`), `unidecode.go` (`:35`, `:75-90`), `unidecode_test.go` (`:42`)
- Modify: `internal/config/registry.go` (`:197-230`), `keys.go` (`:172-199`), `config_test.go` (`:443-444`),
  `keys_test.go` (`:136-165`), `thresholds_test.go` (`:119-129`), and the config tests that exercise
  `Optional` (Reading 3)
- Modify: `Makefile` (`:3-8`)

**Outcomes:**
- The `shadow` verb, `internal/engine/shadow`, the `shadow` integration class, and its five keys and two
  thresholds are gone. A `DUBPLATE_SHADOW_*` key in the environment now refuses as "unknown configuration key"
  (`config.go:338`).
- `setUnidecodeTransliterator` is deleted. `unidecode install` and doctor's `unidecode` row stay; their comments
  name the import pipeline's scoring as the consumer, without a rung number.
- `matcherCommit` stays as build identity. Its rationale in `main.go` (`:39-51`, `:152-158`) and the `Makefile`
  (`:3-8`) states what it identifies, with no shadow refusal.
- `IntegrationClass.Optional` and the Optional-threshold pairing skip go (Reading 3), with their tests; the
  fixture class keeps its Required rows.
- Doctor's shadow fixtures: the table test at `doctor_test.go:95-110` registers no shadow class.
  `TestDoctorCmd_ShadowRows`' two generic properties (a duration threshold renders as a duration; an integration
  row reports its flag) are re-pointed at a production registration (a backup duration threshold and the backup
  class), unless an existing test already proves one, which the report names.
- `TestNewServeCmd_WiredIntoRootCmd` (`serve_test.go:252-280`) reads its anchor against another production verb.
- `thresholds_test.go` uses a fixture key name, not `DUBPLATE_SHADOW_FLOOR`.
- The pre-deletion commit (the parent of this task's commit) is in the task's report, for HISTORY and the
  document pass's `IDEAS.md` entry.

**Acceptance criteria:**
- `git grep -nIiE 'engine/shadow|DUBPLATE_SHADOW|registerShadow|newShadowCmd' -- cmd internal scripts Makefile`
  returns nothing. (The spec's full pattern is closed by Task 2b.)
- `go list -deps ./cmd/dubplate-server | grep -E 'internal/(harness|engine/shadow)'` is empty.
- A test proves `DUBPLATE_SHADOW_ENABLED=true` in the environment fails `config.Load` with "unknown configuration
  key". **Fixture state:** an environment otherwise valid, so the shadow key is the only error.
- **Named mutations:** drop `newServeCmd` from `AddCommand` (the re-pointed wiring test fails); render a
  duration threshold through the int path (the re-pointed doctor row fails).
- The gate green.

## Task 2b: The shadow program's orphans

Pass class: `tool`. Model: `sonnet`, effort `high`. Depends on Task 2a. **0.8 hooks:** neutral.

**Files:**
- Modify: `internal/harness/analysis/adjudication.go`, `adjudication_test.go`
- Modify: `internal/engine/musicbrainz/limiter.go`, `limiter_test.go`, `client.go`, `client_test.go`
- Modify (comments only): `internal/engine/tag/tag.go` (`:58-67`, `:161`), `internal/engine/tagio/doc.go`
  (`:34`), `internal/engine/tagio/container.go` (`:68`), `internal/engine/matcher/quiet.go` (`:5`),
  `internal/engine/store/leaf.go` (`:14`), `internal/auth/store.go` (`:28`, `:120`); test comments in
  `internal/atomicfile/atomicfile_test.go` (`:153`), `internal/engine/tagio/container_test.go` (`:271`),
  `internal/harness/conform/defaults_test.go` (`:205`)
- Modify: `internal/engine/tag/tag_test.go` (`:274`, the "503 propagates" row)

**Outcomes:**
- `internal/harness/analysis` is case-only: `AdjudicationSubject` keeps `CaseID` alone, `AdjudicationEntry` loses
  `MatcherCommit` and `MatcherConfigHash`, and the validator's shadow branches go. The JSON shape of a case row is
  unchanged, so `corpus/analysis/adjudications.json` loads as it is.
- `musicbrainz.ClassBackground` and `ErrUnavailable` are pruned (Reading 2). The limiter's doc and its
  anti-starvation reasoning are restated for two classes. Retry exhaustion still returns an error that wraps the
  last failure; the tag test's 503 row asserts that and that it is not `ErrNotFound`.
- `internal/engine/tagio` stays intact, its doc naming the import pipeline's reader as its consumer without a rung
  number. Its exported names now have no importer, so the exported-unused leg warns on them; HISTORY records it.
- Every remaining comment naming the shadow, its store, or its observation is rewritten to state the code's reason
  without it.

**States the adjudication check meets:**

| Row | Report |
| --- | --- |
| `case_id` set, known case | accepted |
| `case_id` set, unknown case | refused naming the id (unchanged) |
| carrying `album`, `release_mbid`, `matcher_commit`, or `matcher_config_hash` | refused at decode as an unknown field (`ledger.go:19-20`) |
| no `case_id` | refused naming the missing subject |

**Acceptance criteria:**
- The spec's full grep, `git grep -nIiE 'engine/shadow|DUBPLATE_SHADOW|registerShadow|newShadowCmd|shadow (observe|program|runner|store|integration|class|leaf)' -- cmd internal scripts Makefile`,
  returns nothing. CSS `box-shadow`, the rendered gate's `component/surface/shadow/...` ids, and the "shadow" idiom
  stay.
- `grep -cE 'release_mbid|matcher_commit|matcher_config_hash' corpus/analysis/adjudications.json` prints `0`
  (verified at plan time), and the corpus tests that load it pass.
- A test row loads a ledger carrying `release_mbid` and asserts the decode refusal. **Named mutation:** remove
  `dec.DisallowUnknownFields()` and the row fails.
- **Named mutation:** disable the limiter's reservation (every token by strict priority) and the restated
  anti-starvation test fails.
- The gate green.

## Task 3: Configuration retirements and the remaining stale prose

Pass class: `tool`. Model: `sonnet`, effort `high`. Depends on Task 2a. **0.8 hooks:** neutral. Hook 6's
"driver declines capabilities" motivation leaves the config docs; the empty-Required behavior it described stays,
justified on its own terms.

**Files:**
- Modify: `internal/config/keys.go` (`:80-104`, `:128-155`), `keys_test.go` (`:69`), `config.go` (`:15-30`,
  `:121`, `:423-424`), `config_test.go` (`:458-460`, `:567-582`)
- Modify: `internal/auth/seam.go` (`:17-22`), `session.go` (`:12-19`), `handlers.go` (`:104-107`), `token.go`
  (`:160-163`), and the roster comments that describe an external streamer (`roster.go:43-51`, `:112-117`,
  `:824-827`)
- Modify: `cmd/dubplate-server/store.go` (`:17-22`), `doctor_test.go` (`:1048-1063`)

**Outcomes:**
- Deleted rows: the reserved `streamer` class (`DUBPLATE_NAVIDROME_URL`, `_ADMIN_USER`, `_ADMIN_PASSWORD`) and the
  monitoring class's `DUBPLATE_PING_NAVIDROME_URL`.
- `DUBPLATE_PUBLIC_NAVIDROME_URL` and `DUBPLATE_NAVIDROME_DB` read "retired: dubplate integrates with no external
  streamer", after the `DUBPLATE_TUNNEL_TOKEN` precedent. `DUBPLATE_ACCESS_TEAM_DOMAIN` and `DUBPLATE_ACCESS_AUD`
  read as retired, with no front named as a destination (Owner ruling 2).
- `auth.Seam`'s doc names no alternate implementation. `SessionCookieName`'s doc states same-origin's reason
  without an SSO capability. `handlers.go` and `token.go` describe the link's path as owned by the shell, with no
  fallthrough slot. The roster comments describe the built-in streamer resolving a username case-insensitively,
  or the column's `COLLATE NOCASE`, rather than a downstream streamer.
- The `DUBPLATE_STORE_BUSY_TIMEOUT` comment (`store.go:17-22`) states the 15s default's reason without the retired
  system.
- Doctor's reserved-class order test (`doctor_test.go:1048-1063`) pins table order against two surviving classes.

**States a retired or deleted key meets:**

| Key in the environment | Report |
| --- | --- |
| `DUBPLATE_NAVIDROME_URL`, `_ADMIN_USER`, `_ADMIN_PASSWORD`, `DUBPLATE_PING_NAVIDROME_URL` | "unknown configuration key" |
| `DUBPLATE_PUBLIC_NAVIDROME_URL`, `DUBPLATE_NAVIDROME_DB` | the "retired: dubplate integrates with no external streamer" message |
| `DUBPLATE_ACCESS_TEAM_DOMAIN`, `DUBPLATE_ACCESS_AUD` | the retired message |
| none of them | no report; `check`'s required set is unchanged (ruling 7) |

**Acceptance criteria:**
- With Task 1, the spec's prose criterion holds in full:
  `go doc -all ./internal/auth` and `go doc -all ./internal/config`, each piped to
  `grep -inE 'front ?exemption|forward-auth|driver|fallthrough|exemption'`, print nothing.
- `git grep -niE 'navidrome' -- cmd internal ':!*_test.go'` returns only the two retired rows' key names.
- A table test covers each row of the states table above. **Named mutation:** delete the
  `DUBPLATE_PUBLIC_NAVIDROME_URL` superseded row and its row reports "unknown" instead of the message.
- Ruling 7: the required-key count `check` reports with an empty environment matches Task 0's count.
- The gate green.

## Task 4a: The store column becomes `username`

Pass class: `sweep`. Model: `sonnet`, effort `high`. Independent of every other task. **0.8 hooks:** realigns hook
5 (the users row's handle is named `username`); builds nothing new.

**Precondition:** the conductor's read-only re-confirmation of "The `0003` preconditions" passes, after the
document pass's archive chore is recorded in STATUS. On a failure the task is re-cut as migration `0006`, a
`RENAME COLUMN`, and the conductor amends this task before dispatch.

**Files:**
- Modify: `internal/engine/store/migrations/0003_catalog.sql` (`:5`)
- Modify: `internal/engine/store/testdata/fixtures/0003_catalog.sql` (`:5`), `0004_settings.sql` (`:8`),
  `0005_routed.sql` (`:9`)
- Modify: `internal/engine/store/schema_test.go` (`:194-238`, `:297`)

**Outcomes:**
- `users.username TEXT NOT NULL UNIQUE COLLATE NOCASE`, edited in place under the one-time erratum the document
  pass records in the rung 6 spec.
- `migrationHashes["0003_catalog.sql"]` re-pinned from an independent `sha256sum` over the edited file. The commit
  message quotes the `sha256sum` line and cites the precondition evidence (this plan's section and the
  re-confirmation's date).

**Acceptance criteria:**
- `git grep -nIiE 'streamer[_ -]?username' -- internal/engine/store` returns nothing.
- `sha256sum internal/engine/store/migrations/0003_catalog.sql` equals the pinned value, and
  `TestSchemaMigrationHashPins` passes.
- The schema test's case-folded duplicate row asserts "UNIQUE constraint failed: users.username". **Named
  mutation:** drop `COLLATE NOCASE` and that row fails. **Fixture state:** the first insert succeeds before the
  case-folded second is attempted.
- The gate green.

## Task 4b: The roster, the API, the CLI, and the page say "username"

Pass class: `sweep`. Model: `sonnet`, effort `high`. Depends on Tasks 1 and 3. **0.8 hooks:** realigns hook 5's
field name; neutral otherwise.

**Files:**
- Modify: `internal/auth/roster.go`, `admin.go`, `seam.go`, and their tests (`roster_test.go`, `admin_test.go`,
  `seam_test.go`)
- Modify: `cmd/dubplate-server/member.go`, `doctor.go` (`:91`), and the tests naming the old key or flag
  (`member_test.go`, `doctor_test.go`, `serve_test.go`, `backup_test.go`, `restore_test.go`)
- Modify: `internal/backup/job_test.go`, `reconcile_test.go`, `restore_test.go`
- Modify: `web/src/routes/members/+page.svelte`, `web/src/lib/copy/deck.ts`, `web/scripts/rendered/run.ts`
- Modify: `config/roster.example.json`, `docs/server-runbook.md` (`:138`, `:141`, `:164` only)

**Outcomes:**
- `StreamerUsername` becomes `Username`, `streamer_username` becomes `username` as the roster file's key and the
  API's JSON and form key, and every `streamerUsername*` identifier, sentinel, and constant is renamed
  (`roster.go:37-52`, `:105-141`, `:187`, `:533-613`, `:739`, `:812-832`; `admin.go:22-45`, `:77`, `:180`, `:253`).
- The flag is `--username` and stays required (`member.go:60-106`). `member list`'s column head reads `USERNAME`
  (`member.go:172-186`).
- The wire codes are `duplicate_username`, `invalid_username`, and `missing_username`; `duplicate_principal`'s copy
  drops "streamer".
- The roster stays at `schema_version` 1, with no version gate and no compatibility shim (spec).
- The members page: the field's label is "Username"; its hint says it is the name the member types into a music
  app, then the shape rule; the field's `id`, `name`, and hint `id` drop "streamer"; `FIELD_FOR_OUTCOME` and the
  page's state follow. The copy follows the visual identity spec's "Copy" section (Microsoft editor standard).
  The task checks Microsoft's A-Z word list for "user name": if it spells two words, `deck.ts` records the
  departure beside the label, citing ruling 3 and the music apps' own label. The report states what the guide
  says and the URL read.
- `run.ts` follows every renamed id, label key, outcome key, and mocked member field.
- `config/roster.example.json` and the runbook's two bootstrap commands use `username` and `--username`.
- `Principal` stays the username exactly as today; 6d's revision 2.3 re-justifies it (spec).

**States each check meets:**

| Input | Answer |
| --- | --- |
| roster file with `username` | loads |
| roster file with `streamer_username` only, or both keys | `ErrRosterUnparseable` (`RejectUnknownMembers`) |
| roster file member with no username | loads; the principal is the address (the loader's looser check, unchanged) |
| API POST with a valid `username` | `added` |
| API POST carrying `streamer_username` | `unknown_field` |
| API POST with no `username` / a bad shape / a case-folded duplicate / another member's address | `missing_username` / `invalid_username` / `duplicate_username` / `duplicate_principal` |
| API POST carrying `admin` | `admin_not_grantable` (unchanged) |
| `member add` with `--streamer-username` | cobra's unknown-flag error |
| `member add` without `--username` | cobra's required-flag error |

**Acceptance criteria:**
- `git grep -nIiE 'streamer[_ -]?username' -- cmd internal web/src web/scripts config docs/server-runbook.md`
  returns nothing, and `git grep -n 'streamer-hint' -- web` returns nothing.
- `TestRosterExampleFile` (`roster_test.go:440-451`) passes against the renamed example.
- A test row per states-table line. **Named mutations:** decode the roster without `RejectUnknownMembers` (the
  old-key row loads and fails the test); add `streamer_username` to `rosterFields` (the `unknown_field` row
  fails); drop `MarkFlagRequired("username")` (the required-flag row fails). **Fixture state:** the old-key roster
  row carries an otherwise valid member, so the refusal can only be the unknown key.
- The rendered gate's members walk passes with the new ids and label (`run.ts:3555-3655`, `:4130-4210`,
  `:5035-5080`), and the deck-coverage and ai-tells gates pass.
- The gate green.

## Task 5: The close

Pass class: `docs` for its own edits. Model: `sonnet`. One fold agent drafts and commits, then one independent
`diff-reviewer` read of its diff.

The conductor runs, in order, before dispatching Task 5:

1. `code-simplifier:code-simplifier` over the pass's changed Go, TypeScript, and Svelte, with
   `.claude/instructions/simplifier-brief.md` pasted verbatim, then one `diff-reviewer` read of its diff.
2. `web-auth-security-reviewer` over the Task 1 diff (`internal/serve`, `cmd/dubplate-server/serve*.go`), asking
   for correctness gaps ranked by consequence, vacuous re-pointed tests first. One fold task if it returns a
   blocking finding.
3. `svelte-reviewer` and `daisyui-a11y-reviewer` on one shared brief scoped to Task 4b's `web/` diff, with the
   visual identity spec's path.
4. One `go-architecture-reader` per touched package, no plan in context: `internal/serve`,
   `cmd/dubplate-server`, `internal/config`, `internal/auth`, `internal/engine/store`,
   `internal/engine/musicbrainz`, `internal/harness/analysis`, `internal/engine/tag`, `internal/engine/tagio`,
   `internal/engine/matcher` (Reading 6). Findings that are the removal's own go to a fold task; others are filed
   to the cleanup inventory.
5. The settle steps: `auth-data`'s live auth smoke is the rendered gate's sign-in walk against a real serve shell,
   which every full gate runs (no deploy happens in this pass); `sweep`'s spot capture is the rendered gate's
   members-page capture in both themes, read by the conductor.
6. Update `.claude/agents/dubplate-implementer.md:60`, which still names `internal/engine/shadow` (a rule lives
   where it executes).

**Task 5's outcomes:**
- The pass's HISTORY entry, newest first: what landed per task; the gate evidence; each reviewer's dispositions;
  both budgets (tokens against 8.0M; attended time); and "What a later pass would be wrong to rediscover": the
  bearer bypass stays on purpose with its consumers; `0003` was edited in place under the read preconditions, and
  any older store fails at its ledger check; the roster carries no compatibility shim and the first
  roster-reading deploy takes the runbook step; `tagio` warns on the exported-unused leg until the import pipeline
  imports it; `readFLAC` is salvage for rung 9 at `internal/engine/shadow/libraryread.go` in the pre-deletion
  commit; `ClassBackground`, `ErrUnavailable`, and `IntegrationClass.Optional` were pruned for a consumer to
  re-add; Task 0's operator-supplied count. It names any finding a fold refused that proved real in execution, or
  says none did.
- STATUS, present tense: the pass closed; the lane 1 row; the next actions (6d's revision 2.3, the document pass's
  segment B); the moot carries the removal closed (`FrontExemption`'s dual-stack match, the configured-header
  carve-out).
- Out-of-scope findings from any dispatch go to `BACKLOG.md`, verified first.

## Contested surfaces

- With the document pass (main tree): `docs/server-runbook.md` (this pass edits three lines in Task 4b; the
  document pass edits the rest after this pass merges), `docs/STATUS.md` and `docs/HISTORY.md` (this pass's close
  writes first; the document pass writes after), `IDEAS.md` (the document pass's only).
- With the slice 2 lane (worktree): `go.mod` and `go.sum` (Task 0 lands first; slice 2 rebases).
- With 6d (not running): every `internal/auth`, `cmd/dubplate-server`, and `internal/engine/store` file here. 6d
  moves to revision 2.3 after this pass (spec).

## The conductor's actions

1. Before Task 0: the live-executor check (lane registry, `git status`, document lane state), the lane registry
   row, the runaway guard, AC power. Commit this plan and point STATUS at it.
2. Task 0 through the `dependency-upgrade` skill; record the count.
3. Before each segment: one `haiku` pre-flight that checks every file:line and claim the segment's tasks cite at
   HEAD; amend the plan, then launch.
4. Before Task 4a: the read-only `0003` re-confirmation, after the archive chore's STATUS record.
5. At each segment boundary: STATUS; the 6.4M flag.
6. The close's reviews and reads (Task 5's list), then Task 5, then close the session.

## Prior art consulted

None engaged. The pass adopts no library and builds no capability; the streamer spec's prior-art list is weighed
by the slices that build.

## Not verified in this revision

- Line citations are at `1c37e40`. Task 0's sweep and the document lane's commits can move them; each segment's
  pre-flight re-resolves its tasks' citations.
- Microsoft's A-Z entry for "user name" was not fetched (the guide's URL for it returned 404); Task 4b checks it.
- The operator-supplied count was probed for two rounds only; Task 0 runs it to the end.
- The rendered gate's sign-in walk is relied on as the live auth smoke; the pre-flight confirms it mints and
  consumes a real link.
