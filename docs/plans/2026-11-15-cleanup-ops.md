# Pass plan: a committed add answers success, privilege before the body, and a smaller ops and web surface

Initiative: dubplate-server
Round: cleanup (pass 5 of 5, under the conductor `conductor-dubplate`)
Pass: cleanup-ops

Design: `/var/home/glw907/Projects/initiatives/dubplate-server/specs/2026-11-15-cleanup-round-design.md`, approved by
Geoff in the Approval Queue on 2026-11-16 (document `20261116-000308-88f4d3`). Read in full: "Calls made here" (8, 15,
and 16 bind this pass), "Scope", "How it fails", pass 5's row under "Passes" (line 144), "What each merge changes",
"Acceptance" (pass 5 at lines 282-306, and the two every-pass items), "Standing answers" 13 to 20, "Rulings" (the
stacking ruling of document `20261116-163931-082d9d`), and "Backlog verdicts" (#121 to #125, #52, #77). The findings
are `docs/research/2026-11-10-whole-repo-review.md` lines 209-240 (O1 to O16), and its "Leave alone" list is lines
284-288. Where this plan and the design differ, the design wins and the difference is a stop.

Worktree: `/var/home/glw907/Projects/dubplate/.claude/worktrees/cleanup-ops`, branch `cleanup-ops`, cut by the
conductor from `origin/cleanup-web` at `16e97c4a9c6419ff786abfc509bf6cccfa1958da` (PR #15's head, stacked on #14 and
#13, all open). **The PR's base is `master`**, as the planner prompt says (D1). Class: `auth-data` (the design's row).
The PR never lands unattended; it waits for the round's close packet with passes 2, 3, and 4.

Review lens: one `standards:plan-reviewer` read of this plan (under "Review record"); a task-scope
`standards:diff-reviewer` read on every task, since every task is `auth-data` (T1, T2, and T3 by commit, T4 with T3a
and T5 with T6 by each subagent's merged range); at the close the whole-branch `standards:diff-reviewer` and one
`standards:web-auth-security-reviewer` read, each given `git diff 16e97c4...HEAD` (the base is `master`, whose diff
carries passes 2 to 4). Executor: `dubplate-cleanup-ops`, `sonnet` at `medium` with the Fable advisor. Two subagents
run beside the session: A (`opus`, T5 and T6, from T0) and B (`sonnet` at `medium`, T4 and T3a, from T1's commit).
Design path minutes: 140. This plan's path: 274 (see "Path"), 6 minutes under the 280-minute park line (standing
answer 14).

## Pre-flight

No step needs Geoff. The pass adds no secret, no sign-in, no dialog, no migration, no config key, and no deploy, and it
reaches no live account and never the box.

Permissions checked: read `~/.local/state/claude-denials.jsonl` on 2026-11-16 for the last 30 days. Fifteen entries
carry a dubplate cwd. Eleven are "Classifier unavailable" refusals, classifier outages and no rule. One "[Unauthorized
Access]" is a `curl` download of a libxml2 tarball. Three are `sed -i` edits refused in another agent's worktree
(`cmd/dubplate-server`, `internal/engine/store`), two of them by pair subagents on 2026-11-16. No task here downloads
anything, and every edit is made with the Edit or Write tool inside the session's or a subagent's worktree,
never `sed -i`. Every command the tasks run (`npm ci --prefix web`, `bash scripts/gate.sh`, `go test`, `go doc`,
`grep`, `git`, `gh pr checks`, `gh pr create`) has run in passes 1 to 4 or in this planning session without a denial.
No `.claude/settings*` rule is wanted.

Reviewed: one `standards:plan-reviewer` read; its findings and the fold are under "Review record".

Evidence: the mechanism table and the recorded searches below. Each row was read or run against `16e97c4` on
2026-11-16, by this planner or by the plan reviewer's probes.

## What the pass changes

Adding a member answers success with the stored member whenever the write committed: `Roster.AddMember` returns the
row it wrote, and the seam no longer reads it back (O6). An admin operation checks the caller's privilege before it
reads the body, so a signed-in member who is not an admin and sends a malformed body gets the privilege refusal (O1's
order). The four job and catalog operations leave `/api/v1`, its document, and `internal/ops` (O2, call 8); BACKLOG #52
stays open as the marker for the page that brings them back. A roster change through the web writes the caller's
address to its audit record, and an operator reads `local:<user>` in the roster and settings audits as it already does
in the jobs audit (O5, call 15). No type in `internal/serve` wraps the response writer to capture a status (O7). The
web app's document, asset, theme, and branding code moves out of `cmd/dubplate-server/web.go` and `internal/serve`
into one new package, `internal/webapp` (O10, call 16). The process citations under `cmd/dubplate-server`,
`internal/auth`, and `internal/serve` go. A member sees nothing change; an API client sees four operations gone and
one refusal reordered.

## Rules that bind every task

1. **Files.** Pass 5's row: `internal/{ops,auth,serve,identity,notify}/**` and `cmd/dubplate-server/**`, plus the new
   `internal/webapp/**` (acceptance item 6 places it), the one `auth-data` line in
   `.claude/skills/dubplate-pass/SKILL.md` (item 6), `BACKLOG.md` (filing finds, rewording #121 to #125 to their
   residue), and this plan and its prompt (moved to `docs/archive/plans/` at the close). A call site a changed
   signature forces is taken and listed under `## Decisions`; anything else is filed through `standards:log-issue`
   in BACKLOG and the PR body and left (standing answer 15). `scripts/check.sh` is not this pass's (the citation leg
   gates in round `engine`). `internal/settings/**` is not this pass's: T4 changes the string the command line hands
   it, never its code. Never touch `docs/STATUS.md` except to write a `## Stop` (the conductor owns it; wanted text
   goes in the PR body under `## STATUS text`), or `.claude/settings*` ("`.claude/settings*` is never an executor's
   file: record the rule the pass wants for the close packet and go on"). Never touch the review's "Leave alone" list
   (standing answer 16): of it, this pass's files hold `ops/local/jobs.go:149-236` and `auth/token.go:236-335`.
   `cmd/dubplate-server/doctor.go` moved onto `config.KeyError` in pass 3; this pass changes it only at a forced call
   site and for its two citation lines (item 7, D9).
2. **The box.** The round adds no feature, no migration, and no config key, and it deploys nothing. No task reaches
   the box (`5.78.148.42`), by ssh or otherwise. No cookie, token, or session lifetime changes.
3. **Gates.** The fast lane is `bash scripts/gate.sh --scope-changed`, with no `--base` (D2). The close is `bash
   scripts/gate.sh --merge`. Each runs bare from the worktree root with the Bash tool's `timeout: 600000`, re-issued on
   exit 75 until it prints `gate exit:`. Never pipe a gate into `tail` or `head`. Never commit red.
   `RUN_GATE_IF_BUSY=defer` is never used: every task is `auth-data`.
4. **Commits.** One commit per task (a review fix may add one), the task id first in the subject, imperative mood,
   specific files only (never `-A`), ending with `Co-Authored-By: Claude Opus 5.5 <noreply@anthropic.com>`. Never
   commit a secret or a `.env` value; read each diff by eye before staging.
5. **Pushes.** Only `git push origin cleanup-ops`. The draft PR opens on T1's push: `gh pr create --draft --base
   master`. After each push, watch CI as a background Bash task, bare: `gh pr checks <n> --watch --fail-fast`, with
   `gh run watch <run id> --exit-status` as the fallback. A red stops the line when it arrives. A red outside this
   pass's files is fixed in this branch and listed under `## Decisions` (standing answer 19).
6. **Records.** PR notes at `<git dir>/pr-notes.md` (`<git dir>` is
   `/var/home/glw907/Projects/dubplate/.git/worktrees/cleanup-ops`; a subagent uses this absolute path in its report
   and writes nothing there itself), first line `launch-base=<sha>`. Each task appends its test-first failures,
   mutation results, review line, gate exits, and anything found outside scope.
7. **Expected values** are literals from this plan or from a fixture's own input (`"erin@example.com"`,
   `"local:operator"`, `"192.0.2.1"`, `403`, `forbidden`), never a figure the code under test printed or a value
   built with the helper the code calls.
8. **Mutation proofs are run.** Apply each mutation alone, run the named test with `go test -run`, record the failing
   output's first lines with the test name, and revert. One that doesn't fail is a finding fixed in the same task.
9. **Smallest thing that meets acceptance.** No abstraction, export, flag, or fallback the design doesn't name. Every
   new export names its production caller in the task that adds it.
10. **Exemplary Go.** Load `standards:go-conventions` before the first Go edit. Idiomatic, architecturally beautiful
    Go is an acceptance dimension of every task: a correct but awkward result is a `fix`. Doc comments follow Go Doc
    Comments; a comment that a rename or a move makes false is rewritten in the same commit. No em dash in a code
    comment or commit.
11. **Kept tests.** No sign-in, session, or roster test is deleted unless the review names it. A test of a moved
    function moves with it and keeps its name and assertions. The test lists at T0 (`go test -list . <pkg>` for each
    package in the row) are the baseline item 8 compares against.

## Layering

Edges no task may add, each held by a test or a recorded grep at the close:

- `internal/serve` must not import `internal/webapp` (the shell stays below the app).
- `internal/webapp` must not import `internal/auth`, `internal/ops`, `internal/engine/...`, or `internal/settings`:
  the admin gate (`auth.RequireAdmin`) wraps its handlers in `cmd/dubplate-server`, as it does today.
- `internal/webapp` must not import `internal/config`: the config keys stay declared in `cmd/dubplate-server`
  (`registerWebConfig`, the four thresholds), which resolves them and hands `webapp` plain values.
- `internal/ops` must not import `internal/ops/opshttp` or `internal/auth` (unchanged today; T4 adds a field, never
  an import).

T5 adds one import-graph test in `internal/webapp`, modelled on `internal/intake/seam_test.go:733`, that fails when
the package's transitive imports reach any of the four packages above; its mutation is a blank import of
`internal/auth` in a non-test `webapp` file.

## Mechanisms

Each row was read at `16e97c4`.

| # | Mechanism | Source | What a task relies on |
| --- | --- | --- | --- |
| 1 | `AddMember` calls `writeMember(..., s.cfg.Roster.AddMember(ctx, candidate, rosterActor(actor), logger))`, then `added, err := s.cfg.Roster.Lookup(ctx, req.Address)`; a `Lookup` error returns as is and answers 500 for a committed add, and a retry then gets `duplicate_address` | `internal/ops/admin/members.go:35-42` | T1 |
| 2 | `Roster.AddMember` returns `error` only. A new address INSERTs under `store.NewID()`, which is not returned; a removed member's address is reactivated by UPDATE under its existing id. `logChange` writes the "roster member added" record after the commit, on the logger the caller passes | `internal/auth/roster.go` (`AddMember`, `logChange` at `:634`) | T1: the fault fires after the commit and before the old `Lookup` |
| 3 | `AddMember`'s callers: `internal/ops/admin/members.go:35`, `cmd/dubplate-server/member.go:118`, `cmd/dubplate-server/rosterreset_test.go:162`, `internal/auth/sentinel_test.go:44,59`, `internal/auth/roster_test.go:303,327,351,363,425,467,1088,1340,1400,1442,1700` | `grep -rn 'AddMember(' --include=*.go cmd internal` | T1: each is a forced call site |
| 4 | `route[Req,Resp]` runs `decode(w, r, &req, bodyCap(op))` and then `call(r.Context(), ops.Actor{}, req)`, where `call` is an `authorizingSeam` method that runs `admit` and then the seam beneath. So a malformed body answers 400 before the privilege check. `ServeHTTP` checks, in order, the route, POST, no query, and `application/json`, before `route` | `internal/ops/opshttp/opshttp.go:79-127`, `:193-218` | T2 |
| 5 | `admit(ctx, op)` returns `ops.NewNotFoundError` for an op `PrivilegeOf` doesn't know, and otherwise a forbidden answer wrapping `errAuthorizerDenied`, `errChannelMismatch`, or `errPrivilegeShort`; `respond` maps the code through `statuses` and logs `denialReason` for a forbidden answer | `authorize.go:104-125`; `opshttp.go` `respond` | T2: `route` calls `admit` itself |
| 6 | `TestServeHTTPRefusesBeforeAuthorizing` asserts zero authorizer calls and zero seam calls on every row. Rows 31-43 and 55-57, 60 (method, path, query, content type) stay at zero. Rows 44-54 and 58-59 refuse a body shape (empty, over the cap, `null`, an array, a string, a number, two objects, garbage, a wrong type, a malformed or missing handle, an array on `remove_member`, a wrong type on `reset_setting`). Under T2 only the admin and submission body rows (58, 59) are admitted first and move to a one-call table; the job handlers still decode before `authorized` until T3, so the body rows naming `start_backup`, `job_status`, or `cancel_job` stay at zero calls through T2 | `internal/ops/opshttp/opshttp_test.go:20-60`; `opshttp.go:152-187`; `helpers_test.go:218` (`fakeAuthorizer.calls`) | T2 moves rows 58 and 59; T3 retargets each job body row to a kept op, moves it into T2's one-call table, and drops the two handle rows with `validHandle` |
| 7 | `api.go:35` always builds the mount over `ops.NewUnavailableJobs(noJobRunner)`. `ops.Seam` is implemented by `authorizingSeam`, `UnavailableJobs`, and `local.Runner` (`local/local.go:92`, an assertion only). `.Catalog(` has one production call, `authorize.go:157`; `AdminCatalogEntries` and `SubmissionCatalogEntries` have one each, in the same function. The CLI calls `runner.StartBackup`, `JobStatus`, and `CancelJob` directly (`cmd/dubplate-server/backup.go:257,322,343`) and `ops.KindBackup` (`store.go:540`, `local.go:182`) | `grep -rn` over `cmd internal`, non-test | T3: the job request types, `Handle`, `ParseHandle`, `KindBackup`, and `Runner`'s three job methods stay; `Seam`, the catalog types, `Runner.Catalog`, and the two `*CatalogEntries` go |
| 8 | `internal/ops/catalog.go` holds `Privilege` (`valid`, `String`, `MarshalText`, `UnmarshalText`), the op constants (`:65`), `KindBackup` (`:73`), `privileges` (`:79`), `PrivilegeOf` (`:103`), and `CatalogEntry`, `Catalog`, `BuildCatalog`, `Entries`, `catalogEntries`. `catalog_test.go` holds `TestPrivilegeOf` and `TestPrivilege_UnmarshalText`, which follow `Privilege`, and the catalog tests, which go | as cited | T3: the kept half moves to `internal/ops/privilege.go` with its two tests in `privilege_test.go` |
| 9 | `contract_test.go` (1,023 lines) drives a mount over a real `local.Runner` with a stand-in rclone (`TestMain`); none of its helpers (`withCaller`, `denyAll`, `bounded`, `syncBuffer`, `newContract`) is referenced from another file. `document_test.go` keeps `TestDocument_PathsMatchTheCatalogBothWays` (`:213`), `goConstants`, and `enumOf`; `operationNames` (`:172`) stays but is rebuilt (mechanism 25); `TestDocument_JobValuesValidateAgainstTheSchema` (`:596`) tests only job values | `grep` over `internal/ops/opshttp/*_test.go` | T3: `contract_test.go` and the job-values test go; the paths test proves the document and the routes agree |
| 10 | `ops.Actor` is `{wire, channel, userID, anchor}`; `NetworkActor(id identity.Identity, anchor time.Time)` renders `member:<hex>`, `LocalActor(username)` renders `local:<username>`. `rosterActor` in the admin seam returns `auth.MemberActor(actor.UserID(), "")`, and its comment says the seam "is handed no request". `apiVerifier` builds the actor from `auth.CheckRequest` and logs `cfg.ClientIP(r)` beside it | `internal/ops/ops.go:25-80`; `internal/ops/admin/admin.go:97-102`; `cmd/dubplate-server/api.go:79-97` | T4 |
| 11 | `auth.Actor` is `{channel, id, name, clientIP}`. `MemberActor(id, clientIP)` (`:166`); `OperatorActor(username)` sets `clientIP: "local"` (`:173`); `String()` (`:210`) returns the bare `name` for the command line and `member:<hex>` otherwise. `logChange` writes attrs `actor` (`by.String()`), `channel`, `address`, and `client_ip` (`by.clientIP`) | `internal/auth/roster.go:152-215`, `:634` | T4: `String()` returns `"local:"+name` for an operator |
| 12 | The command line's settings audit is `settings.Author{Actor: auth.OperatorActor(resolveActor()).String(), Channel: ...}`, which `internal/settings/write.go` writes to the setting's `actor` column and its audit row. The jobs audit writes `ops.LocalActor(...)`'s `local:<user>` already | `cmd/dubplate-server/settings.go:53-56`; `internal/settings/write.go` | T4: the one `String()` change reaches both audits |
| 13 | `ClientIP` on `auth.AdminConfig` is "the client address a request is held accountable for", the shell's `serve.ClientIP` under its trusted-proxy policy; the gate's refusal records and the sign-in handlers already log it | `internal/auth/admin.go:17-19`; `handlers.go:121-125` | T4: the address the web audit carries is that one value |
| 14 | Test literals that read an operator's bare name: `internal/auth/roster_test.go:407,416,1576` and `cmd/dubplate-server/settings_test.go:436` (`actor != "operator"`). `cmd/dubplate-server/member_test.go:337` expects `actor=uid:%d`, the operator's bare name, which becomes `actor=local:uid:%d`; `internal/ops/admin/admin_test.go:278` already reads `local:test` | `grep` | T4: changes each, `member_test.go:337` included |
| 15 | `statusCapture` (`internal/serve/serve.go:260-318`, `WriteHeader` at `:275`, with `Unwrap`, `Flush`, `Hijack`, `ReadFrom`) has two users: `logRefusal` (`serve.go:357-371`), which logs `capture.status` only when a shell wrapper filed a refusal through `markRefusal`, and `LogRequests` (`document.go:261`). Every `markRefusal(r, status)` site writes the status it marks | as cited | T5: `logRefusal` logs `rec.Status` |
| 16 | `LogRequests` wraps three handlers only: login (`cmd/dubplate-server/web.go:197`), and members and server settings inside `auth.RequireAdmin` (`cmd/dubplate-server/serve.go:456-457`). Each is `newExactPathRoutes` (`web.go:851`) over `serve.NewDocument`, so the statuses that reach it are 200 (`NewDocument`), 405 with `Allow: GET, HEAD` (`methodGetHead`), and 404 (the exact-path router's not-found answer). No `http.ServeContent` and no 304 reach it; the asset handler is not wrapped. `methodGetHead` also fronts `NewStylesheet` and `NewAssets` (`document.go:193,305`), and `newExactPathRoutes` also fronts the probe through `exactPath` (`cmd/dubplate-server/serve.go:548-549`); neither is logged today | as cited | T5: only the document path logs, and no wrapper remains |
| 17 | `cmd/dubplate-server/web.go` is 860 lines: the route names and limits (`:31-92`), `webDocuments`, `privateFiles`, the four `config.Threshold` vars and `registerWebConfig` (`:124-180`), `webRoutes` and `register` (`:182-226`), `newWebRoutes`/`newWebRoutesFromFS` (`:228-382`), and from `:384` the CSP check, stylesheet and static-file handlers, fonts, icons, branding resolution, theme colors, operator CSS, theme extraction, and `newExactPathRoutes`. It imports `config`, `serve`, and `web`, never `auth`. `newExactPathRoutes`' one outside caller is `exactPath` (`serve.go:548-549`). `resolveBranding` (`:540`) and `resolveOperatorCSS` (`:642`) call `config.ParseString`, and lines 1-226 alone are 226 lines | `grep -n '^func \|^type \|^var \|^const '`; `grep -rln` | T5: the parsing stays in `cmd`, and `webRoutes`, `webDocuments`, and the theme constants move |
| 18 | `internal/serve/document.go` (440), `csp.go` (146), and `branding.go` (197) are referenced from serve's other non-test files only by a comment naming `DocumentCSP` (`doc.go`). `document.go` calls `boundedPath` and `maxLoggedPathBytes` (`serve.go:385,392`), which stay in `serve` for `logRefusal` | `grep` over `internal/serve` | T5: `webapp` needs its own bound or an export from `serve` (D6) |
| 19 | The citation leg greps the text from each line's first `//` onward in tracked non-test `.go` files under `cmd` and `internal` for `ruled, conductor\|Resolution R\|rung [0-9]\|Task [0-9]\|\(C[0-9]\|plans/.*\.md:[0-9]\|_salvage/.*:[0-9]\|Ruled, \|\bR-[0-9]+[a-z]+[0-9]*`, prints the first twenty, and warns unless `CHECK_CITATIONS=fail` | `scripts/check.sh:793-830` | T6: the count probe below reproduces it per directory |
| 20 | `go tool deadcode ./...` warns on each function no `main` reaches unless `CHECK_EXPORTED_UNUSED=fail` | `scripts/check.sh` (the leg after the citations) | T3, T5: the report for this pass's packages does not grow |
| 21 | `check.sh --scope-changed --dry-run` on this branch measures from the merge base with `master`, which spans passes 2 to 4 and holds `.gitleaks.toml`, so every fast lane runs the full gate, about 4 to 5 minutes on pavilion | `bash scripts/check.sh --scope-changed --dry-run` | D2; the path counts it |
| 22 | Outside `internal/ops`, tests use `catalog` as the generic signed-in probe: `cmd/dubplate-server/apimount_test.go:171,199,210,220,240,320,324,357,379,399,526` (the 401, refresh, remove-yourself, and member-privilege tests), `apisubmissions_test.go:117-136` (the member 403 table names the three job ops; the catalog read asserts the submission ops' privilege), `routes_test.go:59`, and `serve_test.go:509,1992,2305`. `internal/auth/request_test.go:20,41` only builds a request path that `auth.CheckRequest` never routes. `web/src` names none of the four | `grep -rln` over `cmd internal web/src` | T3a: each `cmd` site retargets to a kept op (`list_submissions` for a member probe, `list_members` for an admin one) before T3 deletes `catalog` (mechanism 28); `request_test.go` stays as is |
| 23 | `httptest.NewRequest` sets `req.RemoteAddr = "192.0.2.1:1234"`, and the api harness's `callAt` builds its requests that way. `TestAPI_RecordsTheSessionsActor` reads the "api request admitted" record from the harness's `a.logs` and asserts only a non-empty `client_ip` | `go doc -src net/http/httptest.NewRequestWithContext`; `cmd/dubplate-server/apimount_test.go:89-150,277-309` | T4: the expected address is the literal `192.0.2.1` |
| 24 | `web/build` holds only `placeholder.txt` in a fresh worktree, and `web.go:246-248` refuses it, so `go test ./cmd/dubplate-server/` fails until `make web-build` runs after `npm ci --prefix web` | `ls web/build`; `Makefile:61` | T0 and each subagent's first act |
| 25 | `document_test.go`'s `operationNames` (`:172-186`) lists the four job ops by name and calls `ops.AdminCatalogEntries` and `ops.SubmissionCatalogEntries`; `:427`, `:512-518`, `:768-769`, and `TestDocument_TheChecksCanFail` (`:784-837`) use a job or catalog op. With the two functions gone, `adminOperations` (`internal/ops/admin.go:123`) and `submissionOperations` (`submissions.go:21`) have no reader, which the `unused` linter (`.golangci.yml:10`) fails. `ops_test.go:298-302` round-trips `CatalogRequest` and `CatalogResponse` | as cited | T3: `operationNames` reads `h.routes`, the two lists go, `TheChecksCanFail` retargets to `add_member` (which declares 409) |
| 26 | `contract_test.go:561-569`'s `listMembers` leg calls `c.seam.ListMembers` on the authorizing seam, a wrapper T2 deletes; `contract_test.go:199` is a `NetworkActor` site | as cited | T2 rewrites the leg to admit and then call the admin seam (forced; T3 deletes the file); B takes `:199` |
| 27 | Cancelling the context `writeMember` derives from makes the old `Lookup` fail with an error wrapping `context.Canceled` (`internal/auth/sql.go:1327-1333`). `refusal` names no case for it, so it returns as is; `ops.CodeOf` maps `context.Canceled` to `canceled`, which answers 503 (`internal/ops/errors.go:176`), not 500. The slog handler sees `writeMember`'s derived context, so it must call the test's own `cancel` for the context the test passed | as cited | T1: the recorded base failure is "an error wrapping `context.Canceled`", whatever its status |
| 28 | At base, through `newAPIHarness`, a member session's `list_submissions` and a stale admin session's `list_members` answer 200 in place of `catalog` (`TestAPI_MemberReachesOnlyTheCatalog` and `TestAPI_RefreshKeepsTheAnchor`, edited, run, and reverted by this planner on 2026-11-16) | `go test -count=1 -run` | T3a: the retargets are valid before T3 deletes `catalog` |

## Recorded searches

Run at `16e97c4` from the worktree root.

| Search | Hits | Use |
| --- | --- | --- |
| `grep -rn 'func (.*) WriteHeader' internal/serve` | 1: `internal/serve/serve.go:275` | item 5's start |
| `wc -l cmd/dubplate-server/web.go` | 860 | item 6's start |
| `ls internal/` | `atomicfile auth backup config engine harness identity intake notify opensubsonic ops reject serve settings` | T5: `webapp` is free (`web` is the embed package at `web/`) |
| the citation probe below, per directory | `cmd/dubplate-server` 33, `internal/serve` 14, `internal/auth` 15, `internal/ops` 0, `internal/identity` 1, `internal/notify` 0 | item 7's start (the design's 12, 11, 10 is a `[stale-fact]`) |
| `wc -l internal/ops/opshttp/contract_test.go` | 1,023 | item 4's start |
| `grep -c 'operationId' internal/ops/opshttp/openapi.yaml` and the four names | the four operations present | item 4's start |
| `grep -rn 'NetworkActor(' --include=*.go cmd internal` | 1 production (`api.go:87`), 12 test sites across `ops`, `ops/admin`, `ops/opshttp`, and `internal/intake/seam_test.go:40` | T4: D5 |
| `grep -rn 'LogRequests' --include=*.go cmd internal` (non-test) | 3 call sites, 1 definition | T5 |

The citation probe, run from the worktree root (`pat` is mechanism 19's pattern, quoted):

```
for d in cmd/dubplate-server internal/serve internal/auth internal/ops internal/webapp internal/identity; do
  n=$(git ls-files --cached --others --exclude-standard -- "$d" | grep -E '\.go$' | grep -v '_test\.go$' |
    xargs -r grep -Hn '//' |
    awk 'match($0,/^[^:]+:[0-9]+:/){h=substr($0,RSTART,RLENGTH);r=substr($0,RLENGTH+1);i=index(r,"//");if(i)print h substr(r,i)}' |
    grep -cE "$pat")
  echo "$d $n"
done
```

## Tasks

Every task's check includes a task-scope `diff-reviewer` read (adapter "Per-task review"), since every task is
`auth-data`. Three streams run at once:

- The session runs T0, T1, T2, the B merge, T3, the A merge, and T7, in that order, with a fast lane after each task
  and each merge.
- Subagent A (`general-purpose`, `model: opus`, `isolation: "worktree"`) runs T5 then T6, dispatched at the end of T0
  from T0's head.
- Subagent B (`general-purpose`, `model: sonnet`, effort `medium`, `isolation: "worktree"`) runs T4 then T3a,
  dispatched right after T1's commit from T1's head, beside T2.

Each subagent follows the adapter's "Dispatches": its first act is `git merge-base --is-ancestor <dispatch sha>
HEAD`, then `npm ci --prefix web` and `make web-build` (mechanism 24). It loads `standards:go-conventions` before any
Go, commits each of its tasks, never pushes, runs no gate (D3), and returns its records in its report. Before each
dispatch the session appends `dispatch <tasks> sha=<40-hex>` to the PR notes.

The file lists are disjoint except where A9 names shared files, each split by hunk.

### T0: Setup and baselines (12 min, session)

1. The one-executor checks (adapter "Rules the pass carries"), `EnterWorktree` `path`, then `npm ci --prefix web` and
   `make web-build`, each bare.
2. Write `<git dir>/pr-notes.md` with `launch-base=<HEAD>` (expected `16e97c4` plus the plan and prompt commit).
3. Record the baselines in the PR notes: `go test -list . ./cmd/dubplate-server ./internal/auth/... ./internal/serve/...
   ./internal/ops/... ./internal/identity/... ./internal/notify/...` to `<git dir>/tests-base.txt` (expected counts:
   cmd 295, auth 148, serve 92, ops 136 across its packages, identity 10, notify 9; a different count is recorded, not
   a stop); the citation probe; the item 5 grep; `wc -l` on `web.go`; and `go tool deadcode ./...` to
   `<git dir>/deadcode-base.txt`. Deadcode lines are compared by function name, never by file path, because T3 and T5
   move functions between files and packages (`RewriteDocument` at `document.go:31` is a test-only export that moves).
4. Dispatch subagent A with T5 and T6 now.

### T1: Adding a member answers the stored member (35 min + 5 review, `auth-data`, session)

O6, item 1. Test first.

1. Write the failing test in `internal/ops/admin/members_test.go`: an add through the seam over a populated roster
   (one admin, the actor), whose `Config.Logger` carries a `slog.Handler` that calls the test's own `cancel` for the
   context the test passed to `AddMember` when it receives the "roster member added" record. `logChange` writes that
   record after the commit and before the old `Lookup`, and the handler sees `writeMember`'s derived context, so it
   must cancel the parent (mechanism 27). Want: no error, and the answer's member equals the stored row read back
   afterwards through `ListMembers` on a fresh context (address `"erin@example.com"` from the input
   `"Erin@Example.com"`, its id the stored id, never the candidate's zero id). Run it at the task's start commit: it
   must fail with an error wrapping `context.Canceled` (code `canceled`, 503, from the old `Lookup`); record the
   output. If the record arrives on a path where cancelling cannot reach the `Lookup`, use A1.
2. Add a reactivation case: a removed member's address added again answers that member's existing id (a literal
   captured from the fixture's first add, not from the code under test).
3. Change `Roster.AddMember` to return `(Member, error)`: the stored member, its id the one it inserted
   (`store.NewID()` captured) or reactivated. Take every caller in mechanism 3; `AddMember` in the seam answers from the
   returned member and drops the `Lookup`. The test callers that ignore the member write `_, err :=`.
4. Mutation, run alone and reverted: the seam answers the candidate (`added = candidate`) rather than the returned
   member; the test must fail on the id.

Files: `internal/auth/roster.go`, `internal/auth/roster_test.go`, `internal/auth/sentinel_test.go`,
`internal/ops/admin/members.go`, `internal/ops/admin/members_test.go`, `cmd/dubplate-server/member.go`,
`cmd/dubplate-server/rosterreset_test.go`.

Check: the new test passes; `go test ./internal/auth/... ./internal/ops/... ./cmd/dubplate-server/` green; the fast
lane; the task review. Commit, push, open the draft PR (`--base master`), start the CI watch, and dispatch subagent B
with T4 and T3a from this commit.

### T2: Privilege before the body (48 min + 5 review + 5 lane, `auth-data`, session)

Item 2, O1's order. Test first. BACKLOG #77 (a log line before the method check) sits in `mount.go`'s verify step,
which T2 does not reorder; it stays open (D7).

1. Write the failing test in `internal/ops/opshttp/opshttp_test.go`, placed directly after
   `TestServeHTTPRefusesBeforeAuthorizing`, through `newHandler`: a caller the fake authorizer grants
   `ops.PrivilegeMember` (through `grant`) posts each of `[`, `null`, `[]`, and an empty body, with
   `Content-Type: application/json`, to `add_member` and to `set_setting`. Want: 403, code `forbidden`, the denial log
   naming `errPrivilegeShort`'s text, one authorizer call, and zero seam calls. At the start commit each row answers
   400 `invalid_request`; record it.
2. Change `route` so it admits first: after `ServeHTTP`'s route, method, query, and content-type checks, `route` calls
   `h.seam.admit(r.Context(), op)`, answers a denial through `h.respond`, then decodes, then calls the seam beneath
   (`admin` or `subs`) with the admitted actor. The routes map binds the beneath seams' methods. The `authorizingSeam`
   keeps `admit`, `reaches`, the channel check, and the denial causes; it loses its seven admin and five submission
   wrapper methods and the `AdminSeam` and `SubmissionSeam` assertions. The job handlers and their four wrappers, with
   `authorized`, stay for T3 to delete. Rewrite the type's doc comment to say what it now is.
3. Split `TestServeHTTPRefusesBeforeAuthorizing` (mechanism 6): rows 58 and 59 (`remove_member`, `reset_setting`)
   move to a table that asserts one authorizer call and zero seam calls; every other row, the job body rows included,
   keeps zero calls until T3. The `authorize_test.go` tests that called the wrapper methods drive the handler instead,
   keeping their names and assertions (rule 11). `contract_test.go`'s `listMembers` leg (`:561-569`) admits and then
   calls the admin seam (forced, mechanism 26).
4. T2 adds no direct `ops.NetworkActor(` call in any file: a test actor comes from `grant` or `admitAnn`, so B's
   signature change merges without a compile break. T2 does not edit `helpers_test.go`, or `opshttp_test.go` between
   lines 280 and 320 (B's).
5. Mutations, each alone and reverted: (a) decode before `admit` again (the new test must fail); (b) `route` calls the
   seam with `ops.Actor{}` rather than the admitted actor; `TestServeHTTPRoutesAdminOperations` (`:134`), which
   asserts the actor the seam receives, must fail.

Files: `internal/ops/opshttp/opshttp.go`, `authorize.go`, `opshttp_test.go` (outside lines 280-320),
`authorize_test.go`, `contract_test.go` (`:561-569` only).

Check: the new test passes; `go test ./internal/ops/...` green; the fast lane; the task review. Commit and push.

### B merge (12 min, session, after T2 and B's report)

`git merge --no-edit <B's branch>`; a textual conflict is a stop (A9), and a clean merge that fails to build or test is
A11. Run the fast lane on the merge, then the task-scope `diff-reviewer` on B's range (`<T1 sha>..<B's head>`, read
against T4's and T3a's text and items 3 and 8). Copy B's records into the PR notes. Push.

### T3: The job and catalog operations go (62 min + 5 review + 5 lane, `auth-data`, session)

O2, item 4, call 8. One commit, so `document_test.go` proves the document and the server agree. Starts after the B
merge, so the `cmd` probes already name kept operations (T3a).

1. Delete `internal/ops/unavailable.go` and `unavailable_test.go`, and `internal/ops/catalog.go` and
   `catalog_test.go` after moving `Privilege` with its methods, the surviving op constants, `privileges`,
   `PrivilegeOf`, and `KindBackup` to `internal/ops/privilege.go`, and `TestPrivilegeOf` and
   `TestPrivilege_UnmarshalText` to `privilege_test.go` unchanged. `OpStartBackup`, `OpJobStatus`, `OpCancelJob`, and
   `OpCatalog` leave the constants and `privileges`.
2. In `internal/ops`, delete `Seam`, `CatalogRequest`, `CatalogResponse`, `AdminCatalogEntries`,
   `SubmissionCatalogEntries`, `adminOperations` (`admin.go:123`), and `submissionOperations` (`submissions.go:21`),
   with the `ops_test.go:298-302` round-trip rows for the catalog types (mechanism 25). Keep the job request and
   response types, `JobState`, `JobStatus`, `Handle`, and `ParseHandle`, which the command line calls. In
   `internal/ops/local`, delete `Runner.Catalog` and `var _ ops.Seam = (*Runner)(nil)`, and rewrite the two comments
   that name `ops.Seam` (`local.go:70`, `doc.go:1`).
3. In `opshttp`, delete the four job handlers, `validHandle`, the job wrappers, `authorized`, the `jobs` field, and the
   job seam parameter of `newAuthorizingSeam`, `newHandler`, and `NewMount`. Update the `handler` comment's count
   ("sixteen operations") to the routes that remain (twelve). In `cmd/dubplate-server/api.go`, drop `noJobRunner` and
   the `ops.NewUnavailableJobs` argument (a forced call site).
4. Remove `catalog`, `start_backup`, `job_status`, and `cancel_job` from `internal/ops/opshttp/openapi.yaml`, with every
   schema only they reference (A3). Delete `contract_test.go` and `TestDocument_JobValuesValidateAgainstTheSchema`.
   In `document_test.go`, `operationNames` returns the keys of a handler's `routes` map, so the paths test compares the
   document against what the mount routes; `:427`, `:512-518`, and `:768-769` drop or retarget their job and catalog
   entries; `TestDocument_TheChecksCanFail` retargets to `add_member`, which declares 409, keeping each check it makes.
5. Retarget the job rows in `opshttp_test.go`, `mount_test.go`, `carrier_test.go`, and `authorize_test.go` to an admin
   or submission op, keeping each row's assertion. Each job body row of `TestServeHTTPRefusesBeforeAuthorizing`
   retargets to a kept op and moves into T2's one-call table (mechanism 6); drop the two handle rows with
   `validHandle`. Where the `caller` type was defined in `contract_test.go`, move it to `helpers_test.go`.
6. Mutation, run alone and reverted: restore the `catalog` path in `openapi.yaml` only;
   `TestDocument_PathsMatchTheCatalogBothWays` must fail naming it.

Files: `internal/ops/{catalog.go,catalog_test.go,unavailable.go,unavailable_test.go,privilege.go,privilege_test.go,ops.go,ops_test.go,admin.go,submissions.go,doc.go}`,
`internal/ops/local/{local.go,doc.go,local_test.go}`, `internal/ops/opshttp/**`, `cmd/dubplate-server/api.go`.

Check: `test ! -e internal/ops/unavailable.go && test ! -e internal/ops/catalog.go`; `grep -nE
'^\s+/(catalog|start_backup|job_status|cancel_job):' internal/ops/opshttp/openapi.yaml` prints nothing; `grep -rn
'"catalog"\|/catalog"' cmd internal --include=*_test.go` prints only `internal/auth/request_test.go:20,41` and
`apimount_test.go:173`'s `/v2/catalog` (a path that names no v1 operation); `go test ./internal/ops/...
./cmd/dubplate-server/` green, the document test included; `go tool deadcode ./...` shows no new function name under
`internal/ops` against `deadcode-base.txt`; the fast lane; the task review. Commit and push.

### T4: Audit actors name who acted (40 min, `auth-data`, subagent B)

O5's audit half, item 3, call 15. Test first.

1. Write the failing tests:
   - In `cmd/dubplate-server/apimount_test.go`, through `newAPIHarness` (the real mount and `apiVerifier`), an admin
     calls `add_member`; the one "roster member added" record in `a.logs` carries `client_ip` equal to the literal
     `"192.0.2.1"` (mechanism 23). First confirm the harness's `ClientIP` answers the peer address with no trusted
     proxy, and record the line read. At the start commit the field is empty.
   - Change the literals in mechanism 14 that read an operator's bare name: `"local:operator"` in
     `internal/auth/roster_test.go` and `cmd/dubplate-server/settings_test.go:436`, and `actor=local:uid:%d` at
     `cmd/dubplate-server/member_test.go:337`. At the start commit they fail.
   Record each failure.
2. `ops.NetworkActor` gains the client address as a third parameter, held in an unexported field and read by one
   method, `ClientAddress`; its production callers are `apiVerifier`, which passes `cfg.ClientIP(r)` (mechanism 13),
   and `rosterActor`, which passes `actor.ClientAddress()` to `auth.MemberActor`. Rewrite `rosterActor`'s comment.
   Every test call site of `NetworkActor` takes the extra argument (forced; D5): in `opshttp` these are
   `helpers_test.go:229`, `opshttp_test.go:300,307`, `carrier_test.go:16`, `mount_test.go:53,244`, and
   `contract_test.go:199`, each a one-line hunk.
3. `auth.Actor.String()` returns `"local:" + a.name` for the command line. Read `logChange` and every other reader of
   `String()` (`grep -rn '\.String()' internal/auth cmd/dubplate-server` against the actor) to confirm no record keys on
   the bare name.
4. Sweep the two citation lines in files this task already edits: `internal/auth/roster.go:109` and
   `cmd/dubplate-server/member.go:240` (D8).
5. Mutations, each alone and reverted: (a) `rosterActor` passes `""` again (the cmd test must fail); (b)
   `String()` returns the bare name (the roster, settings, and member literals must fail).

Files: `internal/ops/ops.go`, `internal/ops/ops_test.go`, `internal/ops/admin/admin.go`, `admin_test.go`, the
`opshttp` test lines in step 2, `internal/auth/roster.go`, `roster_test.go`, `cmd/dubplate-server/api.go`,
`apimount_test.go`, `settings_test.go`, `member.go`, `member_test.go`, `internal/intake/seam_test.go` (forced, one
line).

Check, in the subagent: the new tests pass; `go test ./internal/ops/... ./internal/auth/... ./internal/intake/...
./cmd/dubplate-server/` green; commit T4. The session runs the fast lane and the review at the B merge.

### T3a: The `cmd` probes name kept operations (12 min, `auth-data`, subagent B, after T4)

Item 4's preparation and item 8, so T3 deletes `catalog` with every `cmd` test already green without it.

1. Retarget every `cmd` site in mechanism 22 to a kept operation with its assertion intact: `list_submissions` for a
   member's probe, `list_members` for an admin's (mechanism 28 ran both at base). The member 403 table in
   `apisubmissions_test.go` drops the three job ops and keeps its admin rows; its privilege check reads
   `ops.PrivilegeOf(ops.OpListSubmissions)` against the literal `ops.PrivilegeMember` in place of the catalog answer.
   `serve_test.go:509,1992,2305` are B's hunks in that file; subagent A edits other hunks of it (A9).
2. A test whose name names the catalog (`TestAPI_MemberReachesOnlyTheCatalog`) is renamed to what it now proves, and
   the old and new names go in the report for item 8's test-list diff (D12).
3. Mutation, run alone and reverted: one retargeted member probe sends an admin op instead; its test must fail on the
   status.

Files: `cmd/dubplate-server/{apimount_test.go,apisubmissions_test.go,routes_test.go}` and the three `serve_test.go`
lines.

Check, in the subagent: `grep -n '"catalog"\|/catalog"' cmd/dubplate-server/*_test.go` prints only
`apimount_test.go`'s `/v2/catalog` row; `go test ./cmd/dubplate-server/` green; commit T3a and return the report:
commit ids, the test-first and mutation records, the rename map, and any find.

### T5: One web app package, no status wrapper (subagent A, about 130 min, off the path)

Items 5 and 6, O7 and O10, call 16.

The package is `internal/webapp`. It builds the HTTP surface of the embedded web app: its documents with their CSP,
its stylesheets with the theme and branding stamped in, its static files, fonts, and icons, and the exact-path router
they sit behind. `internal/serve` keeps the transport shell it is today: listeners, the prefix, origin and
client-address policy, normalization, the strip, headers, the probe, and the refusal record. That split is the reason
for the name: `serve` is how a request reaches the process, `webapp` is what the web app answers.

1. Test first (a cut), written in `internal/serve` at the dispatch commit, since `webapp` does not exist yet: the
   request-log tests are rewritten to drive the composed document handlers rather than `LogRequests` alone. A GET to
   the login path logs one "serve: document request" record with `method=GET`, the path, and `status=200`; a POST logs
   `status=405`; an unknown path under the document router logs `status=404`; a POST to a stylesheet or asset path
   (405) and a probe 404 through `exactPath` write no record. Each passes at the dispatch commit through today's
   `LogRequests`; record the run. The tests move to `webapp` with the code in step 2, names kept. Then write the
   import-graph test (Layering).
2. Move `internal/serve/document.go`, `csp.go`, and `branding.go` with their tests to `internal/webapp`, and move
   web.go's handler logic from `:384` on (mechanism 17), with `newWebRoutesFromFS`'s body, `webRoutes`,
   `webDocuments`, and the theme constants, into `webapp`, behind one exported constructor whose production caller is
   `cmd/dubplate-server/web.go`. `webapp` takes a plain config struct of resolved values and returns the routes it
   builds; it exports the exact-path router for `exactPath` (`serve.go:548-549`), its production caller.
   `cmd/dubplate-server` keeps the route names, `registerWebConfig` and the four thresholds, and the resolution of
   each config value, `config.ParseString` in `resolveBranding` and `resolveOperatorCSS` included, and hands `webapp`
   the results. `web_test.go`'s tests of moved functions move with them, names and assertions kept (rule 11).
3. Remove `LogRequests` and its wrapper. Only the document path logs: `NewDocument` logs its 200 and its own 405 at
   the point it writes each, and the router built for the documents logs its 404, through one unexported helper in
   `webapp`. `methodGetHead` (shared with `NewStylesheet` and `NewAssets`) and the plain exact-path router behind
   `exactPath` log nothing (mechanism 16). `cmd/dubplate-server/serve.go:456-457` and `web.go:197` register the
   handlers directly. `logRefusal` logs `rec.Status` and `statusCapture` goes, with its methods.
4. Mutations, each alone and reverted: (a) drop the 405 log call (the POST row must fail); (b) the stylesheet path
   logs through the document helper (the no-record row must fail); (c) a `markRefusal` site in `serve` marks a status
   other than the one it writes (the refusal-record test must fail; name it, or add a status assertion to the
   existing one); (d) the blank import of `internal/auth` in `webapp` (the import-graph test must fail).

Files: `internal/serve/{document.go,document_test.go,csp.go,csp_test.go,branding.go,branding_test.go,serve.go,serve_test.go,doc.go}`,
`internal/webapp/**` (new), `cmd/dubplate-server/{web.go,web_test.go,serve.go,doctorprobe.go,doctorprobe_test.go,main.go}`,
and `cmd/dubplate-server/serve_test.go` outside B's three lines (A9).

Check, in the subagent: `grep -rn 'func (.*) WriteHeader' internal/serve internal/webapp` prints nothing; `wc -l
cmd/dubplate-server/web.go` under 250; `go list -deps ./internal/serve | grep internal/webapp` prints nothing; the
moved tests pass under their old names; `go test ./internal/serve/... ./internal/webapp/... ./cmd/dubplate-server/`
green; `go tool deadcode ./...` shows no new function name under `serve`, `webapp`, or `cmd/dubplate-server` against
the base report. Commit T5.

### T6: Process citations go (subagent A, about 33 min, off the path)

Item 7. After T5 in the same subagent. Rewrite or drop each comment the citation probe lists under
`cmd/dubplate-server`, `internal/auth`, `internal/serve`, `internal/webapp`, and `internal/identity` (D10), keeping
what the comment says about the code and dropping the process reference. The `web.go:762,789,832,850` citations move
into `webapp` with T5 and are swept there. `internal/auth/roster.go:109` and `cmd/dubplate-server/member.go:240` are
T4's. `auth/token.go`'s three lines (26, 44, 63) sit outside the "Leave alone" range (236-335). `doctor.go:134,278`
are comment edits only (D9).

Files: `cmd/dubplate-server/{auth.go,backup.go,doctor.go,main.go,serve.go,store.go,web.go}`,
`internal/auth/{epoch.go,nonce.go,session.go,store.go,token.go}`, `internal/serve/{freebind_linux.go,normalize.go,server.go}`,
`internal/webapp/**` (moved files), `internal/identity/identity.go`. Comment lines only.

Check: the probe prints 0 for every directory it names; `git diff --stat` touches comment lines only (`git diff -U0 |
grep '^[-+][^-+]' | grep -v '^[-+]\s*//'` prints nothing). Commit T6 and return the report: commit ids, the
test-first and mutation records, the probe and grep outputs, and any find.

### A merge (15 min, session, after T3)

`git merge --no-edit <A's branch>`; a textual conflict is a stop (A9), and a clean merge that fails to build or test is
A11. Run the fast lane on the merge, then the task-scope `diff-reviewer` on A's range, with the `webapp` reason, the
Layering section, and items 5 to 7. Copy A's records into the PR notes. Push.

### T7: The close (50 min, session)

1. Re-run the T0 measures on the head: the citation probe, the item 5 grep, `wc -l web.go`, the test lists to
   `<git dir>/tests-head.txt`, and `go tool deadcode` (compared by function name). Diff the test lists by name across
   packages (a moved test counts as kept under its name, a renamed one under D12's map); every missing name is one the
   review names (`contract_test.go`'s, the catalog and unavailable tests, the job-values test) or is a stop.
2. Add `internal/webapp/**` to the `auth-data` list at `.claude/skills/dubplate-pass/SKILL.md:215` (item 6; the one
   adapter line this pass changes).
3. Dispatch, in one message, the whole-branch `standards:diff-reviewer` and `standards:web-auth-security-reviewer`.
   Give each the design, this plan, and `git diff 16e97c4...HEAD`, and ask for correctness gaps ranked by consequence
   and for conformance to idiomatic, architecturally beautiful Go. Ask the security reviewer to confirm the admit
   order, the four removed operations, the audit address's source and trust, and that no kept sign-in, session, or
   roster test lost an assertion (item 8), with the test-list diff and D12's rename map attached.
4. Fold once, quoting each finding verbatim. Each reviewer that did not return accept gets one delta read of the
   fold; a delta's Low findings are folded once or filed. A named reviewer short of accept after the delta is a stop.
5. `BACKLOG.md`: reword #121 to #125 to their residue (D11), keep #52 and #77 open, and file each find through
   `standards:log-issue`.
6. `git mv docs/plans/2026-11-15-cleanup-ops.md docs/plans/2026-11-15-cleanup-ops-prompt.txt docs/archive/plans/`
   (both are committed at hand-over).
7. `bash scripts/gate.sh --merge` on the final head, after steps 2 to 6. Rerun after any later edit.
8. The PR body: the opening paragraph per `~/.claude/docs/voice/written-for-people.md`, `Round: cleanup`, `Pass:
   cleanup-ops`, the `Review:` and `Calls:` lines, evidence for each acceptance item, `## Decisions`, `## Calls for
   Geoff`, `## Planning misses`, `## STATUS text`, and the score. Mark it ready once the head's check is green. The PR
   waits for the round's close packet.

## Standing answers

| # | Stop | Answer |
| --- | --- | --- |
| A1 | T1's context fault cannot reach the gap between the commit and the old `Lookup` | Use the store fault named in T1 step 1 and keep the stored-row assertion only where the row survives the fault. |
| A2 | A body-shape row or an `authorize_test` assertion cannot keep its meaning under T2's order | Rewrite it to assert the new order's answer and list it under `## Decisions`; never delete it. |
| A3 | A schema in `openapi.yaml` is shared by a removed and a kept operation | Keep it. Remove only what the four operations alone reference. |
| A4 | `web.go` cannot reach under 250 lines while the config declarations stay in `cmd` | Move more handler logic into `webapp` behind the same constructor; the declarations stay (Layering). A shortfall after that is a stop. |
| A5 | A status other than 200, 404, or 405 reaches a logged document handler | Log it at the site that writes it, the same way; a wrapper never comes back. |
| A6 | CI red outside the pass's files | Fix in the branch, list under `## Decisions` (standing answer 19). |
| A7 | The fast lane prints `scope=full` | Expected on every run (D2). Record the line once and go on. |
| A8 | A task passes twice its estimate | One `## Decisions` line with the estimate and the actual, and the run goes on (the adapter's "Clock"). |
| A9 | A subagent's branch conflicts with the chain on merge | A stop: any textual conflict means a task left its list. The lists share files only in disjoint hunks. B against T2: in `opshttp_test.go` T2 owns the refusal table (`:20-60`) and its new test placed right after it, B owns `:300` and `:307`; `helpers_test.go` is B's alone (`:229`); in `contract_test.go` T2 owns `:561-569`, B owns `:199`. A against B: in `cmd/dubplate-server/serve_test.go` B owns `:509`, `:1992`, and `:2305`, and A edits other hunks. |
| A11 | A merge with no textual conflict fails to build or test because of the other side's signature change or a removed op | Fix it as a forced call site in one merge-fix commit (`merge-fix: ...`), list it under `## Decisions`, and go on. A red whose fix changes behavior is a stop. |
| A10 | A finding or a cleanup reaches past this pass's acceptance | File it through `standards:log-issue` and list it in the PR; never do it (standing answer 15). |

## Decisions for the executor

- D1: The PR's base is `master`, as the planner prompt says. The branch sits on PR #15's head, so until #13 to #15
  land the PR's diff carries passes 2 to 4; every review is given `16e97c4...HEAD`. Under the stacking ruling, the
  close document lists the PRs in the designed order.
- D2: The fast lane names no `--base`. The merge base with `master` holds pass 2's machinery, so every fast lane runs
  the full gate (mechanism 21); the path counts it.
- D3: Neither subagent runs a gate. Every fast lane is a full gate, it would serialize under `cairn-run-gate` with the
  session's (no `defer` on `auth-data`), and the session's run on each merge covers the subagent's work. Each subagent
  runs `go test`, `go vet`, and the greps on its packages.
- D13: T4 and T3a go to subagent B beside T2, so the path stays under the park line (see "Path"). T4 adds a parameter
  to `ops.NetworkActor` and T2 adds no call to it (T2 step 4), and T3a's probe retargets were run at base (mechanism
  28), so the pair is safe to run apart. T3 then starts from a tree whose `cmd` tests no longer name `catalog`.
- D4: T2 keeps the job handlers and `authorized` for one commit so T3's deletion stays one commit with the document
  change, as the planner prompt asks. Acceptance item 2 names admin operations.
- D5: The client address joins `ops.NetworkActor` as a parameter rather than a `With` method, so no network actor
  exists without one. Its test call sites, `internal/intake/seam_test.go:40` among them, are forced (standing
  answer 15).
- D6: `webapp` keeps its own path bound for the request record: an unexported copy of `boundedPath` with its 256-byte
  limit, its comment naming the shell's refusal record as the bound it matches. Exporting `serve.boundedPath` would
  add an export whose one caller is a sibling's log line; the subagent may take the export instead if the copy reads
  worse, recording the call.
- D7: BACKLOG #77 stays open: T2 reorders `route`, not the mount's verify step that logs before the method check.
- D8: T4 sweeps the two citation lines in files it edits, so T6's files stay disjoint from the chain.
- D9: `doctor.go` gets comment edits at `:134` and `:278` only, because item 7 names `cmd/dubplate-server` and the
  planner prompt keeps the rest of `doctor.go` out.
- D10: `internal/identity/identity.go:4` is in the row and in the leg that gates in round `engine`, which clears only
  the engine packages; T6 takes it so that round's gate does not trip on a file outside its own.
- D11: The rest of #121 to #125 is filed, never done: O1's merge of the mount and the handler beyond the order, O3,
  O4, O5's type merge, O8, O9, O11 to O16, and #138 to #149. Each item is reworded to its residue.
- D12: A kept test whose name names a removed operation is renamed to what it proves after the retarget, never
  deleted. The PR notes map each old name to its new one, the test-list diff in T7 counts the pair as kept, and the
  security reviewer is asked to confirm each pair kept its assertions.

## Calls

- Call 8 (design): the job operations and the catalog leave `/api/v1`, its document, and `internal/ops`, and BACKLOG
  #52 stays open to bring them back with the page that starts a job.
- Call 15 (design): an operator is `local:<user>` in the jobs, roster, and settings audits.
- Call 16 (design): the new package is `internal/webapp`, for the reason under T5.

## Calls for Geoff

- An API client sees `catalog`, `start_backup`, `job_status`, and `cancel_job` answer 404 `not_found`, and a member
  posting a malformed body to an admin operation sees 403 rather than 400. No page calls any of the four, and the web
  pages never send a malformed body.

## Path

| Step | Minutes |
| --- | --- |
| T0, with `make web-build` | 12 |
| T1, with its review | 40 |
| T2 (48, review 5, lane 5) beside B's setup, T4, and T3a (5 + 40 + 12) | 58 |
| B merge, fast lane, and review | 12 |
| T3, with its review and lane | 72 |
| Lane contention (two subagents' `go test` beside the session's gates) | 15 |
| A merge, fast lane, and review | 15 |
| T7 | 50 |
| Total | 274 `[estimate]` |

The pair counts at its larger side: T2's 58 against B's 57. Subagent A's T5 and T6 (about 163 minutes with setup) run
from the end of T0 and finish near minute 180, before the chain reaches the A merge near minute 209, so they are off
the path. Each task figure counts one fast lane at about 5 minutes (every lane is a full gate, D2), a test-first run
and its mutations under `go test -run` at under a minute each, and a task review at about 5 minutes. T3 carries the
reviewer's re-estimate (finding 6) less the `cmd` retargets that moved to T3a. T7 counts the `--merge` gate, the two
reviewers, one fold with a delta read, and a CI watch of about 7 minutes under the PR body.

The margin is thin: 274 sits 6 minutes under the 280-minute line (twice the design's 140). Pass 4 planned 180 and took
about 290. The park line binds this plan, not the run: a mid-run overrun is an A8 line under `## Decisions` with the
estimate and the actual, and the run goes on. The PR records the difference from the design's figure the same way.

## Acceptance

Copied verbatim from the design's pass 5 list, plus the two every-pass items.

| # | Item | Task | Evidence |
| --- | --- | --- | --- |
| 1 | Adding a member answers success with the member whenever the write committed. A test that makes the old second read fail passes, and fails on the base commit. Today: `ops/admin/members.go:39` re-reads and answers 500. | T1 | the recorded failing run at the start commit, the passing run, the mutation |
| 2 | A signed-in member who is not an admin and posts a malformed body to an admin operation gets the privilege refusal. Today: the body is decoded first (`opshttp.go:121-124`). | T2 | the recorded 400s at the start commit, the passing 403 rows, both mutations |
| 3 | A roster change made through the web writes the caller's address to its audit row, and an operator is `local:<user>` in the jobs, roster, and settings audits. Today: the address is always empty, and the roster and settings audits write the bare `<user>`. | T4 | the recorded failures, the passing tests, both mutations |
| 4 | `internal/ops/unavailable.go` and `internal/ops/catalog.go` are gone. `openapi.yaml` no longer lists `catalog`, `start_backup`, `job_status`, or `cancel_job`, and the document test passes. Today: both files exist, the document lists all four, and `contract_test.go` is 1,020 lines. | T3 | the two `test ! -e`, the empty `openapi.yaml` grep, the document test's pass, its mutation |
| 5 | `grep -rn 'func (.*) WriteHeader' internal/serve` prints nothing: no type wraps the response writer to capture a status. Today: one, `statusCapture` at `serve.go:275`. | T5 | the grep over `internal/serve` and `internal/webapp`, empty |
| 6 | `cmd/dubplate-server/web.go` is under 250 lines. The handler logic and `serve`'s `document.go`, `csp.go`, and `branding.go` live in one new package beside `internal/serve` (finding O10), and the adapter's `auth-data` list names that package. Today: 860 lines. | T5, T7 | `wc -l`, `ls internal/webapp`, the SKILL.md line, the import-graph test and its mutation |
| 7 | The process-citation leg, run in its warning form, reports no line under `internal/serve`, `internal/auth`, `internal/ops`, or `cmd/dubplate-server`. Today: 12, 11, and 10 lines in `serve`, `auth`, and `cmd/dubplate-server`. The leg starts to gate in round `engine`, which clears the engine packages' lines. | T4, T6 | the citation probe at 0 for each, and the close gate's citation leg output |
| 8 | The security reviewer returns accept. Every sign-in, session, and roster test that existed on the base commit still exists and passes, apart from tests the review names, and the reviewer confirms that no kept test lost an assertion. | T7 | the verdict, the test-list diff with every missing name mapped to its finding |
| 9 | The PR's check is green on its head. | T7 | `gh pr checks <n>` |
| 10 | `bash scripts/gate.sh --merge` exits 0 on the close head. | T7 | the `gate exit: 0` line |

## Planning misses

- `[stale-fact]` Item 7's "12, 11, and 10 lines in `serve`, `auth`, and `cmd/dubplate-server`" reads 14, 15, and 33 at
  `16e97c4` (recorded searches); `internal/identity` holds 1. The item stands; T4 and T6 clear all of them.
- `[stale-fact]` O11 counts 15 command constructors and 12 `SilenceUsage` sites; the tree holds 9 and about 14. O11 is
  filed (D11), so the count changes nothing here.
- `[stale-fact]` Item 4's "1,020 lines" reads 1,023. The item stands.

## Review record

One `standards:plan-reviewer` read (`opus`, mechanics and feasibility) ran on the plan at `bdc62bc` and returned six
major and seven minor findings. Each is quoted from its report with its fold.

1. "No step builds the web bundle. T0 step 1 and T5 run only `npm ci --prefix web`, but `web/build` holds only
   `placeholder.txt`, which `web.go:246-248` refuses." Folded: `make web-build` in T0 step 1 and as each subagent's
   first act after `npm ci` (mechanism 24).
2. "T3 leaves `document_test.go` broken." (`operationNames` calls the two deleted functions; lines 427, 512-518,
   768-769, and 791-837 use the job or catalog ops; `adminOperations` and `submissionOperations` lose their reader.)
   Folded: T3 steps 2 and 4 rebuild `operationNames` from `h.routes`, delete the two lists, retarget
   `TheChecksCanFail` to `add_member`, and list each test; mechanisms 9 and 25; T3 rises from 55 to 72 minutes with the
   `cmd` retargets moved out.
3. "Mechanism 6 is wrong, and T3 contradicts it. Under T2 the job handlers (`opshttp.go:151-187`) still decode before
   `authorized`, so rows 44-54 stay at zero authorizer calls." Folded: mechanism 6 corrected; T2 step 3 moves only rows
   58 and 59; T3 step 5 moves each retargeted body row into T2's one-call table.
4. "T4 misses `cmd/dubplate-server/member_test.go:337`." Folded: mechanism 14 and T4 step 1 change it to
   `actor=local:uid:%d`, and the file joins T4's list.
5. "T5 step 3 would log requests that aren't logged today." (`methodGetHead` fronts the stylesheet and asset handlers;
   `newExactPathRoutes`' outside caller is `exactPath` at `serve.go:548-549`.) Folded: mechanisms 16 and 17 corrected;
   T5 step 3 logs only the document path; T5 step 1 adds the no-record rows and mutation (b) proves them.
6. "OWNER FORK, the path. The plan's 275 minutes sits 5 under the park line ... 80 to 90 minutes is realistic [for
   T3]." Folded under option (a), re-estimated before launch: T3 is re-figured, and T4 with the `cmd` probe retargets
   (now T3a) moves to subagent B beside T2 (D13), which brings the path to 274, under the line. No `decide` document
   is needed; the thin margin is flagged under "Path".
7. "T1's handler wording and expected failure. ... The handler has to call the test's own `cancel` for the context
   passed to the seam. ... Record it as 'an error wrapping context.Canceled'." Folded: T1 step 1 and mechanism 27.
8. "T3 file list. It omits `internal/ops/ops_test.go:298-302`." Folded: T3 step 2 and its file list.
9. "T5 step 2 contradicts the Layering rule." (`resolveBranding` and `resolveOperatorCSS` call `config.ParseString`;
   lines 1-226 alone are 226.) Folded: mechanism 17 and T5 step 2 keep the parsing in `cmd` and move `webRoutes`,
   `webDocuments`, and the theme constants; `webapp` takes a resolved-value config struct and exports the router for
   `exactPath`.
10. "T5 test-first. A test written in `webapp` can't pass at the dispatch commit." Folded: T5 step 1 writes it in
    `serve`, then moves it.
11. "T6 file list. It names only `webapp/document.go`, but the citations at `web.go:762,789,832,850` also move into
    `webapp`." Folded: T6 names them and takes `internal/webapp/**`.
12. "The dead-code comparisons go by path." Folded: T0 step 3, T3, T5, and T7 compare by function name.
13. "T7 step 6. It runs `git mv` on `docs/plans/2026-11-15-cleanup-ops-prompt.txt`, which is untracked." Folded: the
    prompt is committed at hand-over, so the `git mv` works.

The dispatch shape changed after the read: subagent B (T4 and T3a beside T2), A9's hunk lists, A11, D13, and T2's
`contract_test.go:561-569` site (mechanism 26, found while checking the disjointness). The reviewer did not see that
shape. Its evidence is mechanism 28, the planner's run of both probe retargets at base, plus the hunk-level
disjointness recorded under A9, each checked against the code at `16e97c4`.
