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
`standards:diff-reviewer` read on every task, since every task is `auth-data` (T1 to T4 by commit, T5 and T6 by the
subagent's merged range); at the close the whole-branch `standards:diff-reviewer` and one
`standards:web-auth-security-reviewer` read, each given `git diff 16e97c4...HEAD` (the base is `master`, whose diff
carries passes 2 to 4). Executor: `dubplate-cleanup-ops`, `sonnet` at `medium` with the Fable advisor; T5 and T6 go to
one `opus` subagent. Design path minutes: 140. This plan's path: 275 (see "Path"), under the 280-minute park line
(standing answer 14).

## Pre-flight

No step needs Geoff. The pass adds no secret, no sign-in, no dialog, no migration, no config key, and no deploy, and it
reaches no live account and never the box.

Permissions checked: read `~/.local/state/claude-denials.jsonl` on 2026-11-16 for the last 30 days. Fifteen entries
carry a dubplate cwd. Eleven are "Classifier unavailable" refusals, classifier outages and no rule. One "[Unauthorized
Access]" is a `curl` download of a libxml2 tarball. Three are `sed -i` edits refused in another agent's worktree
(`cmd/dubplate-server`, `internal/engine/store`), two of them by pair subagents on 2026-11-16. No task here downloads
anything, and every edit is made with the Edit or Write tool inside the session's or the subagent's own worktree,
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
   `"local:operator"`, `"203.0.113.7"`, `403`, `forbidden`), never a figure the code under test printed or a value
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
| 6 | `TestServeHTTPRefusesBeforeAuthorizing` asserts zero authorizer calls and zero seam calls on every row. Rows 31-43 and 55-57, 60 (method, path, query, content type) stay at zero. Rows 44-54 and 58-59 refuse a body shape (empty, over the cap, `null`, an array, a string, a number, two objects, garbage, a wrong type, a malformed or missing handle, an array on `remove_member`, a wrong type on `reset_setting`); under T2 each is admitted first, so the authorizer is called once and the seam never. Most body rows name `start_backup`, `job_status`, or `cancel_job` | `internal/ops/opshttp/opshttp_test.go:20-60`; `helpers_test.go:218` (`fakeAuthorizer.calls`) | T2 splits the table; T3 retargets the job rows to admin and submission ops and drops the two handle rows with `validHandle` |
| 7 | `api.go:35` always builds the mount over `ops.NewUnavailableJobs(noJobRunner)`. `ops.Seam` is implemented by `authorizingSeam`, `UnavailableJobs`, and `local.Runner` (`local/local.go:92`, an assertion only). `.Catalog(` has one production call, `authorize.go:157`; `AdminCatalogEntries` and `SubmissionCatalogEntries` have one each, in the same function. The CLI calls `runner.StartBackup`, `JobStatus`, and `CancelJob` directly (`cmd/dubplate-server/backup.go:257,322,343`) and `ops.KindBackup` (`store.go:540`, `local.go:182`) | `grep -rn` over `cmd internal`, non-test | T3: the job request types, `Handle`, `ParseHandle`, `KindBackup`, and `Runner`'s three job methods stay; `Seam`, the catalog types, `Runner.Catalog`, and the two `*CatalogEntries` go |
| 8 | `internal/ops/catalog.go` holds `Privilege` (`valid`, `String`, `MarshalText`, `UnmarshalText`), the op constants (`:65`), `KindBackup` (`:73`), `privileges` (`:79`), `PrivilegeOf` (`:103`), and `CatalogEntry`, `Catalog`, `BuildCatalog`, `Entries`, `catalogEntries`. `catalog_test.go` holds `TestPrivilegeOf` and `TestPrivilege_UnmarshalText`, which follow `Privilege`, and the catalog tests, which go | as cited | T3: the kept half moves to `internal/ops/privilege.go` with its two tests in `privilege_test.go` |
| 9 | `contract_test.go` (1,023 lines) drives a mount over a real `local.Runner` with a stand-in rclone (`TestMain`); none of its helpers (`withCaller`, `denyAll`, `bounded`, `syncBuffer`, `newContract`) is referenced from another file. `document_test.go` keeps `operationNames` (`:172`), `TestDocument_PathsMatchTheCatalogBothWays` (`:213`), `goConstants`, and `enumOf`; `TestDocument_JobValuesValidateAgainstTheSchema` (`:596`) tests only job values | `grep` over `internal/ops/opshttp/*_test.go` | T3: `contract_test.go` and the job-values test go; the paths test proves the document and the routes agree |
| 10 | `ops.Actor` is `{wire, channel, userID, anchor}`; `NetworkActor(id identity.Identity, anchor time.Time)` renders `member:<hex>`, `LocalActor(username)` renders `local:<username>`. `rosterActor` in the admin seam returns `auth.MemberActor(actor.UserID(), "")`, and its comment says the seam "is handed no request". `apiVerifier` builds the actor from `auth.CheckRequest` and logs `cfg.ClientIP(r)` beside it | `internal/ops/ops.go:25-80`; `internal/ops/admin/admin.go:97-102`; `cmd/dubplate-server/api.go:79-97` | T4 |
| 11 | `auth.Actor` is `{channel, id, name, clientIP}`. `MemberActor(id, clientIP)` (`:166`); `OperatorActor(username)` sets `clientIP: "local"` (`:173`); `String()` (`:210`) returns the bare `name` for the command line and `member:<hex>` otherwise. `logChange` writes attrs `actor` (`by.String()`), `channel`, `address`, and `client_ip` (`by.clientIP`) | `internal/auth/roster.go:152-215`, `:634` | T4: `String()` returns `"local:"+name` for an operator |
| 12 | The command line's settings audit is `settings.Author{Actor: auth.OperatorActor(resolveActor()).String(), Channel: ...}`, which `internal/settings/write.go` writes to the setting's `actor` column and its audit row. The jobs audit writes `ops.LocalActor(...)`'s `local:<user>` already | `cmd/dubplate-server/settings.go:53-56`; `internal/settings/write.go` | T4: the one `String()` change reaches both audits |
| 13 | `ClientIP` on `auth.AdminConfig` is "the client address a request is held accountable for", the shell's `serve.ClientIP` under its trusted-proxy policy; the gate's refusal records and the sign-in handlers already log it | `internal/auth/admin.go:17-19`; `handlers.go:121-125` | T4: the address the web audit carries is that one value |
| 14 | Test literals that read an operator's bare name: `internal/auth/roster_test.go:407,416,1576` and `cmd/dubplate-server/settings_test.go:436` (`actor != "operator"`). `cmd/dubplate-server/member_test.go:337` expects `actor=uid:%d`, and `internal/ops/admin/admin_test.go:278` already reads `local:test` | `grep` | T4: the executor reads each and changes the ones that read the operator's actor |
| 15 | `statusCapture` (`internal/serve/serve.go:260-318`, `WriteHeader` at `:275`, with `Unwrap`, `Flush`, `Hijack`, `ReadFrom`) has two users: `logRefusal` (`serve.go:357-371`), which logs `capture.status` only when a shell wrapper filed a refusal through `markRefusal`, and `LogRequests` (`document.go:261`). Every `markRefusal(r, status)` site writes the status it marks | as cited | T5: `logRefusal` logs `rec.Status` |
| 16 | `LogRequests` wraps three handlers only: login (`cmd/dubplate-server/web.go:197`), and members and server settings inside `auth.RequireAdmin` (`cmd/dubplate-server/serve.go:456-457`). Each is `newExactPathRoutes` (`web.go:851`) over `serve.NewDocument`, so the statuses that reach it are 200 (`NewDocument`), 405 with `Allow: GET, HEAD` (`methodGetHead`), and 404 (the exact-path router's not-found answer). No `http.ServeContent` and no 304 reach it; the asset handler is not wrapped | as cited | T5: the handler that picks each status logs it, and no wrapper remains |
| 17 | `cmd/dubplate-server/web.go` is 860 lines: the route names and limits (`:31-92`), `webDocuments`, `privateFiles`, the four `config.Threshold` vars and `registerWebConfig` (`:124-180`), `webRoutes` and `register` (`:182-226`), `newWebRoutes`/`newWebRoutesFromFS` (`:228-382`), and from `:384` the CSP check, stylesheet and static-file handlers, fonts, icons, branding resolution, theme colors, operator CSS, theme extraction, and `newExactPathRoutes`. It imports `config`, `serve`, and `web`, never `auth`. `newExactPathRoutes` is also called from `doctorprobe.go` | `grep -n '^func \|^type \|^var \|^const '`; `grep -rln` | T5 |
| 18 | `internal/serve/document.go` (440), `csp.go` (146), and `branding.go` (197) are referenced from serve's other non-test files only by a comment naming `DocumentCSP` (`doc.go`). `document.go` calls `boundedPath` and `maxLoggedPathBytes` (`serve.go:385,392`), which stay in `serve` for `logRefusal` | `grep` over `internal/serve` | T5: `webapp` needs its own bound or an export from `serve` (D6) |
| 19 | The citation leg greps the text from each line's first `//` onward in tracked non-test `.go` files under `cmd` and `internal` for `ruled, conductor\|Resolution R\|rung [0-9]\|Task [0-9]\|\(C[0-9]\|plans/.*\.md:[0-9]\|_salvage/.*:[0-9]\|Ruled, \|\bR-[0-9]+[a-z]+[0-9]*`, prints the first twenty, and warns unless `CHECK_CITATIONS=fail` | `scripts/check.sh:793-830` | T6: the count probe below reproduces it per directory |
| 20 | `go tool deadcode ./...` warns on each function no `main` reaches unless `CHECK_EXPORTED_UNUSED=fail` | `scripts/check.sh` (the leg after the citations) | T3, T5: the report for this pass's packages does not grow |
| 21 | `check.sh --scope-changed --dry-run` on this branch measures from the merge base with `master`, which spans passes 2 to 4 and holds `.gitleaks.toml`, so every fast lane runs the full gate, about 4 to 5 minutes on pavilion | `bash scripts/check.sh --scope-changed --dry-run` | D2; the path counts it |
| 22 | Outside `internal/ops`, tests use `catalog` as the generic signed-in probe: `cmd/dubplate-server/apimount_test.go:171,199,210,220,240,320,324,357,379,399,526` (the 401, refresh, remove-yourself, and member-privilege tests), `apisubmissions_test.go:117-136` (the member 403 table names the three job ops; the catalog read asserts the submission ops' privilege), `routes_test.go:59`, and `serve_test.go:509,1992,2305`. `internal/auth/request_test.go:20,41` only builds a request path that `auth.CheckRequest` never routes. `web/src` names none of the four | `grep -rln` over `cmd internal web/src` | T3: each `cmd` site is forced and retargets to a kept op (`list_submissions` for a member probe, `list_members` for an admin one); `request_test.go` stays as is |
| 23 | `httptest.NewRequest` sets `req.RemoteAddr = "192.0.2.1:1234"`, and the api harness's `callAt` builds its requests that way. `TestAPI_RecordsTheSessionsActor` reads the "api request admitted" record from the harness's `a.logs` and asserts only a non-empty `client_ip` | `go doc -src net/http/httptest.NewRequestWithContext`; `cmd/dubplate-server/apimount_test.go:89-150,277-309` | T4: the expected address is the literal `192.0.2.1` |

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

Every task's check includes the fast lane at exit 0 and, since every task is `auth-data`, a task-scope
`diff-reviewer` read (adapter "Per-task review"). The session runs T1 to T4 in order. One subagent runs T5 then T6 in
its own worktree beside them, on files disjoint from the chain's (listed under T5 and T6).

### T0: Setup and baselines (10 min, session)

1. The one-executor checks (adapter "Rules the pass carries"), `EnterWorktree` `path`, then `npm ci --prefix web`.
2. Write `<git dir>/pr-notes.md` with `launch-base=<HEAD>` (expected `16e97c4`).
3. Record the baselines in the PR notes: `go test -list . ./cmd/dubplate-server ./internal/auth/... ./internal/serve/...
   ./internal/ops/... ./internal/identity/... ./internal/notify/...` to `<git dir>/tests-base.txt` (expected counts:
   cmd 295, auth 148, serve 92, ops 136 across its packages, identity 10, notify 9; a different count is recorded, not
   a stop); the citation probe; the item 5 grep; `wc -l` on `web.go`; and `go tool deadcode ./...` to
   `<git dir>/deadcode-base.txt`.
4. Dispatch the T5+T6 subagent (below) now, after appending `dispatch T5+T6 sha=<HEAD>`, so it runs beside T1 to T4.

### T1: Adding a member answers the stored member (35 min + 5 review, `auth-data`, session)

O6, item 1. Test first.

1. Write the failing test in `internal/ops/admin/members_test.go`: an add through the seam over a populated roster
   (one admin, the actor), whose `Config.Logger` carries a `slog.Handler` that cancels the call's context when it
   receives the "roster member added" record, which `logChange` writes after the commit and before the old `Lookup`.
   Want: no error, and the answer's member equals the stored row read back afterwards through `ListMembers` on a fresh
   context (address `"erin@example.com"` from the input `"Erin@Example.com"`, its id the stored id, never the
   candidate's zero id). Run it at the task's start commit: it must fail with the 500 (`Lookup` on a cancelled
   context); record the output. If the record arrives on a path where cancelling cannot reach the `Lookup`, use a
   handler that closes the store's read path instead, and keep the stored-row assertion only where the fault leaves
   the row in place (a fault that deletes the row contradicts it).
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
lane; the task review. Push, open the draft PR (`--base master`), and start the CI watch.

### T2: Privilege before the body (45 min + 5 review, `auth-data`, session)

Item 2, O1's order. Test first. BACKLOG #77 (a log line before the method check) sits in `mount.go`'s verify step,
which T2 does not reorder; it stays open (D7).

1. Write the failing test in `internal/ops/opshttp/opshttp_test.go`, through `newHandler`: a caller the fake
   authorizer grants `ops.PrivilegeMember` posts each of `[`, `null`, `[]`, and an empty body, with
   `Content-Type: application/json`, to `add_member` and to `set_setting`. Want: 403, code `forbidden`, the denial log
   naming `errPrivilegeShort`'s text, one authorizer call, and zero seam calls. At the start commit each row answers
   400 `invalid_request`; record it.
2. Change `route` so it admits first: after `ServeHTTP`'s route, method, query, and content-type checks, `route` calls
   `h.seam.admit(r.Context(), op)`, answers a denial through `h.respond`, then decodes, then calls the seam beneath
   (`admin` or `subs`) with the admitted actor. The routes map binds the beneath seams' methods. The `authorizingSeam`
   keeps `admit`, `reaches`, the channel check, and the denial causes; it loses its seven admin and five submission
   wrapper methods and the `AdminSeam` and `SubmissionSeam` assertions. The job handlers and their four wrappers, with
   `authorized`, stay for T3 to delete. Rewrite the type's doc comment to say what it now is.
3. Split `TestServeHTTPRefusesBeforeAuthorizing` (mechanism 6): the transport rows keep zero authorizer calls; the
   body-shape rows for admin and submission ops move to a table that asserts one authorizer call and zero seam calls.
   The `authorize_test.go` tests that called the wrapper methods drive the handler instead, keeping their names and
   assertions (rule 11).
4. Mutations, each alone and reverted: (a) decode before `admit` again (the new test must fail); (b) `route` calls the
   seam with `ops.Actor{}` rather than the admitted actor; the test that asserts the actor the seam receives must fail
   (the executor names it in the notes, or adds one assertion to an existing test if none fails).

Files: `internal/ops/opshttp/opshttp.go`, `authorize.go`, `opshttp_test.go`, `authorize_test.go`, `helpers_test.go`
(only if a fake needs a field).

Check: the new test passes; the fast lane; the task review. Push.

### T3: The job and catalog operations go (50 min + 5 review, `auth-data`, session)

O2, item 4, call 8. One commit, so `document_test.go` proves the document and the server agree.

1. Delete `internal/ops/unavailable.go` and `unavailable_test.go`, and `internal/ops/catalog.go` and
   `catalog_test.go` after moving `Privilege` with its methods, the surviving op constants, `privileges`,
   `PrivilegeOf`, and `KindBackup` to `internal/ops/privilege.go`, and `TestPrivilegeOf` and
   `TestPrivilege_UnmarshalText` to `privilege_test.go` unchanged. `OpStartBackup`, `OpJobStatus`, `OpCancelJob`, and
   `OpCatalog` leave the constants and `privileges`.
2. In `internal/ops`, delete `Seam`, `CatalogRequest`, `CatalogResponse`, `AdminCatalogEntries`, and
   `SubmissionCatalogEntries`. Keep the job request and response types, `JobState`, `JobStatus`, `Handle`, and
   `ParseHandle`, which the command line calls. In `internal/ops/local`, delete `Runner.Catalog` and
   `var _ ops.Seam = (*Runner)(nil)`, and rewrite the two comments that name `ops.Seam` (`local.go:70`, `doc.go:1`).
3. In `opshttp`, delete the four job handlers, `validHandle`, the job wrappers, `authorized`, the `jobs` field, and the
   job seam parameter of `newAuthorizingSeam`, `newHandler`, and `NewMount`. Update the `handler` comment's count
   ("sixteen operations") to the routes that remain (twelve). In `cmd/dubplate-server/api.go`, drop `noJobRunner` and
   the `ops.NewUnavailableJobs` argument (a forced call site).
4. Remove `catalog`, `start_backup`, `job_status`, and `cancel_job` from `internal/ops/opshttp/openapi.yaml`, with every
   schema only they reference. Delete `contract_test.go` and `TestDocument_JobValuesValidateAgainstTheSchema`; retarget
   the job rows in `opshttp_test.go`, `mount_test.go`, `carrier_test.go`, and `authorize_test.go` to an admin or
   submission op, keeping each row's assertion; drop the two handle rows with `validHandle`. Where the `caller` type
   was defined in `contract_test.go`, move it to `helpers_test.go`. Retarget every `cmd` site in mechanism 22 to a
   kept operation with its assertion intact; the privilege check in `apisubmissions_test.go` reads
   `ops.PrivilegeOf` against the literal `ops.PrivilegeMember` in place of the catalog answer. A test whose name names
   the catalog (`TestAPI_MemberReachesOnlyTheCatalog`) is renamed to what it now proves, and the rename goes in the PR
   notes for item 8's test-list diff (D12). `serve_test.go`'s three lines are T3's; the subagent edits other hunks of
   that file, and a textual conflict on merge is a stop (A9).
5. Mutation, run alone and reverted: restore the `catalog` path in `openapi.yaml` only;
   `TestDocument_PathsMatchTheCatalogBothWays` must fail naming it.

Files: `internal/ops/{catalog.go,catalog_test.go,unavailable.go,unavailable_test.go,privilege.go,privilege_test.go,ops.go,admin.go,submissions.go,doc.go}`,
`internal/ops/local/{local.go,doc.go,local_test.go}`, `internal/ops/opshttp/**`, `cmd/dubplate-server/api.go`, and
the `cmd` tests in mechanism 22 (`apimount_test.go`, `apisubmissions_test.go`, `routes_test.go`, and the three
`serve_test.go` lines).

Check: `test ! -e internal/ops/unavailable.go && test ! -e internal/ops/catalog.go`; `grep -nE
'^\s+/(catalog|start_backup|job_status|cancel_job):' internal/ops/opshttp/openapi.yaml` prints nothing; `go test
./internal/ops/...` green, the document test included; `go tool deadcode ./...` shows no new line under
`internal/ops` against `deadcode-base.txt`; the fast lane; the task review. Push.

### T4: Audit actors name who acted (40 min + 5 review, `auth-data`, session)

O5's audit half, item 3, call 15. Test first.

1. Write the failing tests:
   - In `cmd/dubplate-server/apimount_test.go`, through `newAPIHarness` (the real mount and `apiVerifier`), an admin
     calls `add_member`; the one "roster member added" record in `a.logs` carries `client_ip` equal to the literal
     `"192.0.2.1"` (mechanism 23). The executor first confirms the harness's `ClientIP` answers the peer address with
     no trusted proxy, and records the line it read. At the start commit the field is empty.
   - Change the literals in mechanism 14 that read an operator's bare name to `"local:operator"` (the roster records
     in `internal/auth/roster_test.go`, the settings audit in `cmd/dubplate-server/settings_test.go:436`). At the start
     commit they fail.
   Record each failure.
2. `ops.NetworkActor` gains the client address as a third parameter, held in an unexported field and read by one
   method, `ClientAddress`; its production callers are `apiVerifier`, which passes `cfg.ClientIP(r)` (mechanism 13),
   and `rosterActor`, which passes `actor.ClientAddress()` to `auth.MemberActor`. Rewrite `rosterActor`'s comment.
   Every test call site of `NetworkActor` takes the extra argument (forced; D5).
3. `auth.Actor.String()` returns `"local:" + a.name` for the command line. Read `logChange` and every other reader of
   `String()` (`grep -rn '\.String()' internal/auth cmd/dubplate-server` against the actor) to confirm no record keys on
   the bare name; `cmd/dubplate-server/member_test.go:337`'s `uid:` form is read and left if it reads another field.
4. Sweep the two citation lines in files this task already edits: `internal/auth/roster.go:109` and
   `cmd/dubplate-server/member.go:240` (D8).
5. Mutations, each alone and reverted: (a) `rosterActor` passes `""` again (the cmd test must fail); (b)
   `String()` returns the bare name (the roster and settings literals must fail).

Files: `internal/ops/ops.go`, `internal/ops/ops_test.go`, `internal/ops/admin/admin.go`, `admin_test.go`,
`internal/ops/opshttp/*_test.go` (the `NetworkActor` sites), `internal/auth/roster.go`, `roster_test.go`,
`cmd/dubplate-server/api.go`, `apimount_test.go`, `settings_test.go`, `member.go`,
`internal/intake/seam_test.go` (forced, one line).

Check: the new tests pass; the fast lane; the task review. Push.

### T5: One web app package, no status wrapper (subagent, about 105 min, off the path)

Items 5 and 6, O7 and O10, call 16. Dispatched at T0 as one `general-purpose` subagent with `isolation: "worktree"`
and `model: opus`, under the adapter's "Dispatches". Its first act is `git merge-base --is-ancestor <dispatch sha>
HEAD`, then `npm ci --prefix web`.

The package is `internal/webapp`. It builds the HTTP surface of the embedded web app: its documents with their CSP,
its stylesheets with the theme and branding stamped in, its static files, fonts, and icons, and the exact-path router
they sit behind. `internal/serve` keeps the transport shell it is today: listeners, the prefix, origin and
client-address policy, normalization, the strip, headers, the probe, and the refusal record. That split is the reason
for the name: `serve` is how a request reaches the process, `webapp` is what the web app answers.

1. Test first (a cut): the request-log tests move to `webapp` and are rewritten to drive the composed handlers rather
   than `LogRequests`: a GET to the login path logs one "serve: document request" record with `method=GET`, the path,
   and `status=200`; a POST logs `status=405`; an unknown path under the exact-path router logs `status=404`. Each
   passes at the dispatch commit through today's `LogRequests`; record the run. Then write the import-graph test
   (Layering).
2. Move `internal/serve/document.go`, `csp.go`, and `branding.go` with their tests to `internal/webapp`, and move
   web.go's handler logic from `:384` on (mechanism 17), with `newWebRoutesFromFS`'s body, into `webapp`, behind one
   exported constructor whose production caller is `cmd/dubplate-server/web.go`. `cmd/dubplate-server` keeps the route
   names, `registerWebConfig` and the four thresholds, the resolution of each config value from the environment, and
   `register`, and hands `webapp` resolved values. `doctorprobe.go` calls the router through `webapp`. `web_test.go`'s
   tests of moved functions move with them, names and assertions kept (rule 11).
3. Remove `LogRequests` and its wrapper: the handler that chooses each status logs the record at the point it writes
   it (`NewDocument` 200, `methodGetHead` 405, the exact-path router 404), through one unexported helper in `webapp`.
   `cmd/dubplate-server/serve.go:456-457` and `web.go:197` register the handlers directly. `logRefusal` logs
   `rec.Status` and `statusCapture` goes, with its methods.
4. Mutations, each alone and reverted: (a) drop the 405 log call (the POST row must fail); (b) a `markRefusal` site in
   `serve` marks a status other than the one it writes (the refusal-record test must fail; name it, or add a status
   assertion to the existing one); (c) the blank import of `internal/auth` in `webapp` (the import-graph test must
   fail).

Files: `internal/serve/{document.go,document_test.go,csp.go,csp_test.go,branding.go,branding_test.go,serve.go,serve_test.go,doc.go}`,
`internal/webapp/**` (new), `cmd/dubplate-server/{web.go,web_test.go,serve.go,serve_test.go,doctorprobe.go,doctorprobe_test.go,main.go}`.
None is in T1 to T4's lists.

Check, in the subagent: `grep -rn 'func (.*) WriteHeader' internal/serve internal/webapp` prints nothing; `wc -l
cmd/dubplate-server/web.go` under 250; `go list -deps ./internal/serve | grep internal/webapp` prints nothing; the
moved tests pass under their old names; `go test ./internal/serve/... ./internal/webapp/... ./cmd/dubplate-server/`
green; `go tool deadcode ./...` shows no new line under `serve`, `webapp`, or `cmd/dubplate-server` against the base
report. The subagent runs no gate (D3) and commits T5.

### T6: Process citations go (subagent, about 30 min, off the path)

Item 7. After T5 in the same subagent. Rewrite or drop each comment the citation probe lists under
`cmd/dubplate-server`, `internal/auth`, `internal/serve`, `internal/webapp`, and `internal/identity` (D10), keeping
what the comment says about the code and dropping the process reference. `internal/auth/roster.go:109` and
`cmd/dubplate-server/member.go:240` are T4's. `auth/token.go`'s three lines (26, 44, 63) sit outside the "Leave alone"
range (236-335). `doctor.go:134,278` are comment edits only (D9).

Files: `cmd/dubplate-server/{auth.go,backup.go,doctor.go,main.go,serve.go,store.go,web.go}`,
`internal/auth/{epoch.go,nonce.go,session.go,store.go,token.go}`, `internal/serve/{freebind_linux.go,normalize.go,server.go}`,
`internal/webapp/document.go` (moved), `internal/identity/identity.go`. Comment lines only.

Check: the probe prints 0 for every directory it names; `git diff --stat` touches comment lines only (`git diff -U0 |
grep '^[-+][^-+]' | grep -v '^[-+]\s*//'` prints nothing). Commits T6 and returns its report: commit ids, the
test-first and mutation records, the probe and grep outputs, and any find.

### T5 and T6 merge (10 min, session, after T4)

`git merge --no-edit <subagent branch>`; a conflict is a stop. Run the fast lane on the merge, then the task-scope
`diff-reviewer` on the merged range (`<T4 sha>..HEAD`), with the `webapp` reason, the Layering section, and items 5 to
7. Copy the subagent's records into the PR notes. Push.

### T7: The close (50 min, session)

1. Re-run the T0 measures on the head: the citation probe, the item 5 grep, `wc -l web.go`, the test lists to
   `<git dir>/tests-head.txt`, and `go tool deadcode`. Diff the test lists by name across packages (a moved test counts
   as kept under its name); every missing name is one the review names (`contract_test.go`'s, the catalog and
   unavailable tests, the job-values test) or is a stop.
2. Add `internal/webapp/**` to the `auth-data` list at `.claude/skills/dubplate-pass/SKILL.md:215` (item 6; the one
   adapter line this pass changes).
3. Dispatch, in one message, the whole-branch `standards:diff-reviewer` and `standards:web-auth-security-reviewer`.
   Give each the design, this plan, and `git diff 16e97c4...HEAD`, and ask for correctness gaps ranked by consequence
   and for conformance to idiomatic, architecturally beautiful Go. Ask the security reviewer to confirm the admit
   order, the four removed operations, the audit address's source and trust, and that no kept sign-in, session, or
   roster test lost an assertion (item 8), with the test-list diff attached.
4. Fold once, quoting each finding verbatim. Each reviewer that did not return accept gets one delta read of the
   fold; a delta's Low findings are folded once or filed. A named reviewer short of accept after the delta is a stop.
5. `BACKLOG.md`: reword #121 to #125 to their residue (D11), keep #52 and #77 open, and file each find through
   `standards:log-issue`.
6. `git mv docs/plans/2026-11-15-cleanup-ops.md docs/plans/2026-11-15-cleanup-ops-prompt.txt docs/archive/plans/`.
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
| A9 | The subagent's branch conflicts with the chain on merge | A stop. The lists share one file, `cmd/dubplate-server/serve_test.go`, in disjoint hunks (T3's three probe lines); any textual conflict means a task left its list. |
| A10 | A finding or a cleanup reaches past this pass's acceptance | File it through `standards:log-issue` and list it in the PR; never do it (standing answer 15). |

## Decisions for the executor

- D1: The PR's base is `master`, as the planner prompt says. The branch sits on PR #15's head, so until #13 to #15
  land the PR's diff carries passes 2 to 4; every review is given `16e97c4...HEAD`. Under the stacking ruling, the
  close document lists the PRs in the designed order.
- D2: The fast lane names no `--base`. The merge base with `master` holds pass 2's machinery, so every fast lane runs
  the full gate (mechanism 21); the path counts it.
- D3: The subagent runs no gate. Every fast lane is a full gate, it would serialize under `cairn-run-gate` with the
  session's (no `defer` on `auth-data`), and the session's run on the merge covers both. The subagent runs `go test`,
  `go vet`, and the greps on its packages.
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
| T0 | 10 |
| T1, with its review | 40 |
| T2, with its review | 50 |
| T3, with its review | 55 |
| T4, with its review | 45 |
| Lane contention (the subagent's `go test` beside the session's gates) | 10 |
| T5 and T6 merge, fast lane, and review | 15 |
| T7 | 50 |
| Total | 275 `[estimate]` |

The session's chain, T1 to T4, is 190 minutes. The subagent's T5 and T6 (about 135 minutes) run beside it from T0 and
are off the path. Each task figure counts one fast lane at about 5 minutes (every lane is a full gate, D2), a
test-first run and its mutations under `go test -run` at under a minute each, and a task review at about 5 minutes. T7
counts the `--merge` gate, the two reviewers, one fold with a delta read, and a CI watch of about 7 minutes under the
PR body. Pass 4 planned 180 and took about 290, mostly in gate runs and review folds; this figure counts those, and it
stays under the 280-minute line. The design's figure is 140; the PR records the difference under `## Decisions` with
the actual.

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

(Filled after the `standards:plan-reviewer` read.)
