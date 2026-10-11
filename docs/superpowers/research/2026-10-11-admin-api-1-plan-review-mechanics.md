# admin-api-1 plan review: mechanics and feasibility

Target: `docs/superpowers/plans/2026-10-11-admin-api-1.md` at `9c58bcc`. Reviewer: the `plan-reviewer` agent (Opus,
medium), fresh context. Worktree read: `/var/home/glw907/Projects/dubplate/.claude/worktrees/admin-api-1-plan`; every
pointer below was checked against the files there, and probes ran in a scratch directory. The planning session saved
this file from the agent's reply.

## Verified sound (no finding)

- **`--base` on the fast lane works with the shipped scripts.** `gate.sh` accepts only a 40-character lowercase hex
  value. `check.sh:268-294` uses a supplied base only when it resolves to itself and is an ancestor of `HEAD`;
  otherwise it falls back to the full gate. `go.mod`, `go.sum`, `web/package*.json`, `web/scripts/gates/*`, and
  `web/scripts/rendered/*` are gate machinery (`check.sh:364-370`, the `go.mod` row at `:368`), and machinery promotes
  a scoped run to the full form (`:458-459`). After T0 the plan's rationale holds, and T9 to T11 running the full form
  is correct.
- **Pair halves are file-disjoint**, generated outputs included: T1/T2, T3/T4, and T7/T8. T8's `InspectOptions`
  change is additive, so T7's callers of `store.Inspect` are unaffected.
- **Subagent worktree base.** The worktrees docs: "Inside a worktree, `"head"` resolves to that worktree's `HEAD`, not
  the main checkout's." The `merge-base --is-ancestor` guard backs it up.
- **Per-dispatch model and effort.** The model order is "1. The per-invocation `model` parameter ... 3.
  `CLAUDE_CODE_SUBAGENT_MODEL`", so `CLAUDE_CODE_SUBAGENT_MODEL=sonnet` does not override `opus`. Per-invocation
  effort needs CLI 2.1.292 or later; the workstation has 2.1.296. `claude --bg`, `--effort`, and `--name` exist.
- **The clock hook** reads only `task=`, `start=`, and `estimate=` with `sed -n`, so `base=` is inert.
- **Quoted behaviors match the source:** `hex.DecodeString`, `strings.CutPrefix`, `context.WithValue`, the kin-openapi
  `IncludeResponseStatus` comment, `session.go:76-86`, `conn.go:34-36`, `accessor.go:20-21`, `auth/store.go:49-58`,
  `notify.go:80-89`, `epoch.go:84-91`, and `durationAtLeast` (refuses at `auth.go:282`).
- **Most code pointers hold**, give or take a line: `admin.go:110-165`, `serve.go:120-121`, `serve.go:359`,
  `opshttp.go:37-44`, `opshttp.go:141-144`, `catalog.go:78-83`, `catalog.go:96-101`, `roster.go:321`,
  `roster.go:379-388`, `keys.go:155-160`, `evaluate.go:243`, `:265`, `:292-294`, `:304-307`, `inspect.go:166-171`,
  `deck-coverage.ts:29`, `:43`, `:69-72`, `:155-167`, `rendered_test.go:386`.
- **golangci-lint `unused` is safe.** With `.golangci.yml`'s `exported-fields-are-used: false`, golangci-lint 2.14.0
  reports 0 issues on exported request and response fields never read, so T1's types will not redden lint before T5.
- **The new window rule keeps existing tests green.** Fixtures set the absolute cap to 1h or 3h (`doctor_test.go:546`,
  `:554`, `:1140`; `serve_test.go:2054`, `:2186`) and assert substrings or already expect a refusal.
- **No other gate leg is at risk.** `TestAuthCorrelatedRules` uses 4320h. The citation and exported-unused legs only
  warn. The no-link test (`arch_test.go:213-217`) checks `go list -deps` of the binary, so a kin-openapi import in a
  `_test.go` stays invisible to it.
- **Spec-fold carries are placed:** V4 in T6, V5 in P7 and T5, V6 and V7 in T3, over-ceremony item 2 in P8.

## Findings, ranked by consequence

### M1. major, correctness: T5 cannot classify roster refusals without editing `internal/auth`

- **Location:** plan:441-465 (T5), plan:371-393 (T3).
- **Defect:** every roster sentinel the admin seam must map is unexported in `internal/auth/roster.go:63-104`:
  `errDuplicateAddress`, `errInvalidMemberAddress`, `errNoAdmin` (for `last_admin`), `errActorNotAdmin` (for
  `forbidden`), and `errMemberNotFound` (for `not_found`). `internal/ops/admin` cannot `errors.Is` them, and T5's
  files exclude `internal/auth`. The likely result is a stop or an unplanned edit to an `auth-data` file.
- **Proposed fold:** T3 exports the sentinels the admin seam maps (or one classifier, e.g. `auth.RefusalDetail(err)
  (detail string, ok bool)`), naming T5's package as the caller, with an acceptance row that each maps through
  `errors.Is`.

### M2. major, correctness: the route inventory has no way to enumerate routes

- **Location:** plan:506-508 (T6).
- **Defect:** PS A2 (`group-scale-principles-design.md:620`) requires the test to fail when "a route is added without
  a row". `serve.Shell` exposes no route list (`Register`, `Handler`, `Logger`, `EnableHSTS`), and `newServeShell`
  (`cmd/dubplate-server/serve.go:286-370`) builds the shell internally. The spec says 14a writes nothing in
  `internal/serve`, and a test-only `Shell.Routes()` breaks "no export ahead of a caller". PS A2 also requires "an
  unauthenticated request is refused on every session and admin row", which T6 omits.
- **Proposed fold:** `newServeShell` registers through a cmd-local `routeRegistrar interface{ Register(string,
  http.Handler) }`, which `*serve.Shell` satisfies, and the test wraps it with a recorder; add the unauthenticated
  probe per session and admin row.

### M3. major, correctness: hook 18 and the AC9 sentinel scan miss the identity-source keys

- **Location:** plan:346-363 (T2), plan:444-446 (T5), plan:521 (T6 fixture).
- **Defect:** `magicLinkSource` (`cmd/dubplate-server/auth.go:66-83`) registers ten keys through `AddIdentitySource`
  (`registry.go:352`), among them `DUBPLATE_SESSION_SECRET`, `DUBPLATE_MAGIC_LINK_SECRET`, and
  `DUBPLATE_SMTP_PASSWORD`. T2 names only `Setting`, `Threshold`, `WallRow`, and integration-class keys, so these get
  no description or mark (AC8 says "every registered key"). T6's leak proof seeds sentinels only on wall and
  integration rows, so the secrets and thresholds are never scanned. The `registry.go:132-247` precedent stops before
  `AddIdentitySource` (`:352`) and `AddSetting` (`:416`).
- **Proposed fold:** T2 extends the declarations to `IdentitySource.Required` and `.Optional` (and `Selector` via its
  wall row), with a panic proof; T5 states that no environment-only value (wall, threshold, integration, identity
  source) crosses the wire; T6's fixture puts a distinct sentinel on every non-store-homed key, the four secrets
  named.

### M4. major, correctness: P7's detail enumeration contradicts the spec and T9

- **Location:** plan:200-205 (P7), plan:593-595 and :611 (T9).
- **Defect:** P7 takes the details from `+page.svelte:46-54` plus three. That set includes `duplicate_principal`, which
  no Go code returns and T9 deletes (#53), and omits `invalid_address` (`handlers.go:57`, returned by
  `additionOutcome`, `admin.go:324-325`), which T9's #17 row needs. The spec: "Each existing add outcome keeps its name
  as a `detail`" (spec:95-96).
- **Proposed fold:** P7's set is `additionOutcome`'s refusals (`duplicate_address`, `duplicate_username`,
  `invalid_address`, `invalid_display_name`, `blank_display_name`, `invalid_username`, `missing_username`) plus
  `last_admin`, `environment_sets`, and `busy`; `duplicate_principal` drops.

### M5. major, optional: T5, T6, T9, and T10 estimates are too low

- **Location:** plan:44-49, :487, :534, :622, :649.
- **Defect:** gates run 10 to 19 minutes plus up to 11 minutes of lock wait, and every Go task selects the rendered
  leg. T6 (50) packs the mount, carrier, session check, 401/500 split, an 11-operation OpenAPI document, the
  kin-openapi contract tests with three proofs, V4, the route inventory, 8 mutations, a gate, and the security read:
  closer to 90. T9 and T10 (35 each) are test-first on rendered rows with a full gate; T9 rewrites about six
  `**/roster` intercepts in a 5,437-line `run.ts` (`:3098`, `:3151`, `:4144`, `:4196`, `:4467`, `:5013`) and the
  `AUTH_ERROR_STATUS_RE` path list (`run.ts:866-867`). T5 (45) is an opus dispatch with about 10 fixtures, 9 mutations,
  `npm ci`, and a gate.
- **Proposed fold:** T5 60, T6 90, T9 55, T10 55; task-minutes about 630, under the 660 line.

### M6. minor, correctness: T5's busy fixture leans on a deadline the store says does not bound the wait

- **Location:** plan:480; quote at plan:246-249.
- **Defect:** `go doc ./internal/engine/store Handle.Update` continues past the plan's quote: "a deadline shorter than
  the busy timeout does not shorten it." A short deadline against a second handle's lock answers `ErrBusy` after the
  busy timeout, so the `context.DeadlineExceeded` arm is never exercised.
- **Proposed fold:** complete the quote; split the fixture (same handle holding the write pool's connection under a
  short ctx deadline proves `DeadlineExceeded`; a second handle holding `BEGIN IMMEDIATE` under a short busy timeout
  proves `ErrBusy`); add the mutation "map `DeadlineExceeded` to `internal`".

### M7. minor, correctness: T0's gate goes red unless `go.mod` and `go.sum` are staged

- **Location:** plan:295-313.
- **Defect:** the Makefile's `tidy-check` runs `go mod tidy` then `git diff --exit-code -- go.mod go.sum`, which diffs
  against the index; `check.sh`'s header says to "stage go.mod and go.sum before a dependency sweep's gate".
- **Proposed fold:** T0 stages `go.mod`, `go.sum` (and the web lockfile) before the gate.

### M8. minor, correctness: a Playwright bump in T0 conflicts with "never install outside the worktree"

- **Location:** plan:297-302.
- **Defect:** `playwright` is pinned in `web/package.json`; a minor bump lands under T0's rule but needs its matching
  Chromium in the user cache, which the plan forbids installing. The rendered leg goes red.
- **Proposed fold:** if Playwright moves, hold it to the PR body, like libxml2.

### M9. minor, correctness: T7 misses tests that assert the seed-rule messages, and one message covers env-seed rows

- **Location:** plan:546-550.
- **Defect:** `cmd/dubplate-server/store_test.go:731` and `:741` assert the exact admin-row refusal text.
  `evaluate.go:265`'s parse refusal also covers `env-seed` rows (`store_test.go:770`), where `settings reset` cannot
  help.
- **Proposed fold:** add `store_test.go` and the evaluate tests to T7's files; name the verb only when the stored
  source is `admin`.

### M10. minor, correctness: T2 misses `internal/config/config.go`

- **Location:** plan:356-358.
- **Defect:** turning `IntegrationClass.Required` into declarations breaks readers at `config.go:139`, `:429-431`, and
  `integrationClassOf` (`keys.go:170`).
- **Proposed fold:** add `config.go` and `config_test.go` to T2's files.

### M11. minor, correctness: T10 misses the web document fixture

- **Location:** plan:630-635.
- **Defect:** `cmd/dubplate-server/web_test.go:61-65` builds a fixture filesystem holding every document, and
  `newWebRoutes` refuses a missing one (`web_test.go:114-120`).
- **Proposed fold:** add `web_test.go` to T10's files.

### M12. minor, correctness: T11 removes the deck gate's only parser sanity check

- **Location:** plan:657-658.
- **Defect:** dropping the must-have list and floor (`deck-coverage.ts:155-167`) leaves the `handlers.go` parse
  unchecked, and T9's `Detail<Name>` parse never gets one; a broken regex passes vacuously.
- **Proposed fold:** keep a one-name must-have per source (`signed_out` for handlers, `last_admin` for details) and a
  non-empty check; this also closes BACKLOG #21.

### M13. minor, correctness: Resume cannot find an in-flight dispatch after a crash

- **Location:** plan:154-161.
- **Defect:** the dispatch sha and branch reach the PR notes only when the subagent returns.
- **Proposed fold:** write `dispatch <task> <sha>` to the PR notes before each dispatch; Resume lists dispatch branches
  with `git worktree list`.

### M14. minor, correctness: launch-block gaps

- **Location:** plan:62-79.
- **Defect:** the adapter's one-executor rule requires a bracket-class `pgrep -f`, which the block omits; the `jq`
  filter's `startswith(".../admin-api-1")` matches the planning session; "idle or gone" is ambiguous, since `claude
  agents --json` reports `status` and `state` (the finished `dubplate-alignment` session reads `"status": "busy",
  "state": "done"`).
- **Proposed fold:** match `.cwd | test("/worktrees/admin-api-1$")` or `.name == "admin-api-1"`; check the planning
  session on `.state == "done"`; add `pgrep -f '[P]rojects/dubplate/.claude/worktrees/admin-api-1$'`.

### M15. minor, correctness: dispatches do not tell the subagent to read the spec

- **Location:** plan:146-147; the prompt's dispatch paragraph.
- **Defect:** T3 and T5 cite spec sections and ACs without their text.
- **Proposed fold:** each dispatch names the spec path and tells the subagent to read the sections and ACs its task
  cites before writing tests.

### M16. minor, correctness: the smoke's 5m window constrains every gated step

- **Location:** plan:684-690.
- **Defect:** with the window seeded at 5m for the whole run, every gated step must land within 5 minutes of its own
  session's sign-in.
- **Proposed fold:** sign in session A first and let it age; do the gated steps on a fresh session B, re-signing past 5
  minutes; use A for the refused write.

### M17. minor, correctness: the notice's deadline relies on an unquoted behavior

- **Location:** plan:260-261, :450-452.
- **Defect:** the notice depends on `Send` honoring `ctx` mid-dial (`notify.go:140`,
  `client.DialAndSendWithContext(ctx, msg)`); only the retry-policy sentence is quoted.
- **Proposed fold:** quote go-mail's `DialAndSendWithContext` doc.

### M18. minor, correctness: T6 should name kin-openapi's legacy router

- **Location:** plan:499-502.
- **Defect:** `ValidateRequest` needs a router; `routers/gorillamux` would add `github.com/gorilla/mux` to `go.mod`
  (only in `go.sum` today), a machinery change outside T6's files.
- **Proposed fold:** name `routers/legacy`, as `conformance.go:27-30` does.

### M19. minor, correctness: T9's non-admin row has two readings

- **Location:** plan:601-602, :610.
- **Defect:** a member's `list_members` answers a JSON 403 `forbidden`; T9 keeps the non-admin row's not-found state
  (`run.ts:3691`) and adds a "forbidden" copy. Which one a list-time `forbidden` renders is open.
- **Proposed fold:** a list-time `forbidden` renders the not-found state; the forbidden copy is for writes.

### M20. minor, optional: P9's guard shape is unstated while T3 and T4 build the two sides in parallel

- **Location:** plan:213-215.
- **Proposed fold:** name T4's parameter as `func(context.Context, *store.Tx) error`.

### M21. minor, optional: T5's single dispatch gates the same tree twice

- **Location:** plan:146-148, :486.
- **Defect:** the subagent runs the fast lane, then the session fast-forward merges the identical tree and runs it
  again (10 to 20 minutes). `gate.sh` takes no `--receipt`.
- **Proposed fold:** for a single dispatch, the subagent runs targeted `go test` and the mutations, and the session's
  post-merge fast lane is the one gate.

### M22. minor, optional: T0's risk class does not follow the adapter's map

- **Location:** plan:308.
- **Defect:** the adapter maps `runner` to `scripts/` and `.claude/` only; `go.mod` and `web/package.json` are
  `ordinary`, and only Geoff raises a class.
- **Proposed fold:** class T0 `ordinary`, noting its files trigger the full gate.

### M23. minor, optional: Close has an estimate but no stated clock

- **Location:** plan:136-139, :705-706.
- **Proposed fold:** one sentence either way.

### M24. minor, optional: P4 overstates its citation

- **Location:** plan:185-188.
- **Defect:** `visual-identity-design.md:145-146` reads only "Returning to 'follow the system' belongs to a later
  settings page"; it reserves no route.
- **Proposed fold:** cite AC20.

### M25. minor, optional: exports land ahead of their production callers

- **Location:** plan:315-345 (T1), :441-465 (T5).
- **Defect:** T1's `AdminSeam`, details, and unavailable job seam, and T5's constructor, have production callers only
  in T5 and T6; the convention says each task names its caller.
- **Proposed fold:** name the later task as the caller in each outcome.

### M26. minor: OWNER FORK, where the "Server settings" entry lives

- **Location:** plan:629, :632-633 (T10, `SiteHeader.svelte`, "one appended navigation entry").
- **Defect:** `SiteHeader.svelte` has no navigation list (`:102-108`: the wordmark link only), and a static page
  cannot tell an admin from a member. The non-admin row asserts no admin navigation link in the header
  (`run.ts:3706-3710`, `membersLinkPresent`). A header entry shown to every member exposes an admin-only link.
- **Options:** (1) a header link for every signed-in member, non-admins meeting the not-found page; (2) a header link
  rendered only after an admin-revealing API answer such as `catalog`, which is extra machinery; (3) a link from the
  members page only, which is already admin-only.
- **Recommendation:** option 3; it adds no role detection.

## Counts

0 blockers, 5 majors (4 `correctness`, 1 `optional`), 21 minors, 1 marked owner fork.
