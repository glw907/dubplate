# Draft HISTORY entry: rung 6, pass 6b-ii-a (the ops seam's HTTP projection)

Paste-ready for the conductor's ledger fold, once the merge ritual and the
architecture reads land. The bracketed fields need the merge's own commit
ids; everything else is final. Model: the top entries of `docs/HISTORY.md`
(passes 5c-ii and 6b-i) are the format this draft follows. Scope: Task 0
and Tasks 1-3 only. Restore-side outcomes (the step 3 measurements,
`copyto`, the combination table, the runbook lines) are 6b-ii-b's, not
repeated here.

---

## Rung 6, pass 6b-ii-a (the ops seam's HTTP projection): closed 2026-09-25, lane 2

Executed against
`docs/plans/2026-09-25-rung-6b-ii-restore-http.md` (revision 4,
approved by Geoff 2026-09-25; Ruling 1 split the pass into 6b-ii-a, this
entry, and 6b-ii-b) in the lane 2 worktree
(`.claude/worktrees/lane-2-6b-ii-a`, branch `lane-2-6b-ii-a`), from
`f6c8edb`. Task 0's staleness read (`opus`, fresh context) found no
drift against the plan's code pointers; one note routed forward (below).

Task 0 ran as a report only, no commit. Segment A1 (Tasks 1-3) ran four
commits (`git log --oneline f6c8edb..a8cd736 | wc -l` reports 4), and
the close side (simplifier round, security fold) added two more before
Task 8's own (`git log --oneline f6c8edb..629f3df | wc -l` reports 6
total):

- **Task 1** (`6caccf1`, `sonnet`; reviewer fix `ab58067`) added
  `internal/ops`'s last two sentinels (`ErrForbidden`, `ErrInternal`),
  one classifier (`CodeOf`) and the full code set (`Codes`), the network
  kind of `Actor` (built from an `identity.Identity`, refused unless
  `Valid`, rendering `member:<principal>`), and the one privilege table
  (`PrivilegeOf`) `ops/local`'s `Catalog` now reads instead of
  restating. `ab58067` fixed `wrapCanceled`: `NewCanceledError` now
  takes an empty reason, so the store's own aborted-call text (which
  could carry a path) stays out of the wire `Reason` and reaches only
  the log line.
- **Task 2** (`a0afb06`, `claude-opus-5-5`) added
  `internal/ops/opshttp`: the unexported authorizing seam (looks up
  each operation's privilege, calls the authorizer, admits only as the
  authorizer's own actor, discarding the caller's) and the HTTP layer
  (exact-path `POST /<name>` dispatch, six pre-authorizer request
  refusals, the JSON error shape with a coded reason only).
- **Task 3** (`a8cd736`, `claude-opus-5-5`) added the contract suite
  (`contract_test.go`, 955 lines): every wire code driven in-process
  through the authorizing layer over `ops/local` (deviation D1) and over
  HTTP through the whole projection, both legs asserted to classify to
  the same code; the job path over HTTP; a second process's job
  canceled through the projection; a coverage guard iterating
  `ops.Codes`.
- The `code-simplifier` round (`3d6dd92`, `sonnet`) dropped
  `PrivilegeOf`'s redundant named-return reassignment; accepted by
  review with no other finding.
- The `web-auth-security-reviewer` fold (`629f3df`, `claude-opus-5-5`)
  reshaped the authorizer's shape (see R-6bii2's amendment below),
  closed-set denial logs, an unmapped code's 500 fallback, `CutPrefix`
  for the path match, and `Cache-Control: no-store` on every answer.

### The merge

6b-ii-a merges to master at `<merge-sha, filled at the fold>`: `<rebase
outcome>`, a green worktree gate, `git merge --ff-only`, a green gate on
master, `<merged-range diff-reviewer verdict, if the rebase is not a
fast-forward>`.

### The architecture reads

Three `go-architecture-reader` grades run at the merge, each with no plan
in context: `internal/ops`, `internal/ops/local`, `internal/ops/opshttp`.
`<findings and routing, filled at the fold>`.

### The gate evidence

`go test -count=1 ./internal/ops/...`: **5.671s** wall
(`internal/ops` 0.008s, `internal/ops/local` 2.721s, `internal/ops/opshttp`
4.654s; 62 `--- PASS` lines from `-v`). `go test ./... -count=1` at HEAD
(`8059732`): **2m17.388s** wall (`internal/harness/conform` at 134.373s
dominates, unchanged from prior passes; `cmd/dubplate-server` at 52.320s
is next largest; every touched package finishes under 5s).

`CAIRN_GATE_LANE=light cairn-run-gate 'bash scripts/check.sh'` at HEAD:
**exit 0**, 227-line log. gitleaks scanned about 113.59 MB in 59.5s and
found no leaks.

Two staged legs stayed at warning: **326** process citations in Go
comments (unchanged; this pass's one comment edit, Task 8 Part 1,
carries no citation) and **125** exported-unused `internal/` identifiers
(`ops.Codes` and `ops.CodeOf` gained their production caller in
`opshttp`'s non-test code this pass, but the projection itself is
unmounted, so no binary reaches them yet — the gate does not claim
otherwise).

### The exit gate items this half owns

Per the plan's "The exit gate" (Ruling 1's no: 6b-ii-a owns items 1, 10,
12, and 13, plus 9 and 14 over its own tasks):

- **Item 10** (`internal/ops/opshttp` imports only `internal/ops` and
  `internal/identity`; `internal/ops` imports only `internal/identity`):
  confirmed by `go list -f '{{join .Imports "\n"}}'` over both packages,
  filtered to the module.
- **Item 13** (no HTTP route mounted): `git grep -n opshttp --
  ':!internal/ops/opshttp' ':!docs'` finds nothing.
- **Item 12** (no store created or migrated outside a test): not this
  half's surface; Tasks 1-3 touch no store code.
- **Item 9** (state rows each have a named test): R-6bii1's four literal
  privilege values and R-6bii5's seven-sentinel code set are each
  asserted, per Task 1's and Task 3's own state tables.
- **Item 14** (every exported symbol lists its production caller):
  `ops.PrivilegeOf`, `ops.NewForbiddenError` called from
  `opshttp/authorize.go`; `ops.Codes`, `ops.CodeOf` called from
  `opshttp/opshttp.go`, noted (per item 14's own text) as reachable by
  no binary until rung 14 mounts the projection. `ops.NetworkActor` is
  the network actor's constructor, the plan's named exception routed to
  rung 14.

### R-6bii1 to R-6bii5, where each landed

| Resolution | What it decides | Landed |
| --- | --- | --- |
| R-6bii1 | One privilege table in `internal/ops`, four literal values | `internal/ops/catalog.go` (`PrivilegeOf`), Task 1 (`6caccf1`) |
| R-6bii2 | The network actor, the authorizer, the channel | `internal/ops/ops.go` (`NetworkActor`), Task 1 (`6caccf1`); the authorizer shape amended at the security fold (`629f3df`, see below) |
| R-6bii3 | The authorizing layer is a seam decorator; contract suite's in-process leg runs `ops/local` wrapped in it (D1) | `internal/ops/opshttp/authorize.go`, Task 2 (`a0afb06`); D1 in `contract_test.go`, Task 3 (`a8cd736`) |
| R-6bii4 | The wire mapping: exact-path `POST /<name>`, six pre-authorizer refusals, the JSON error shape | `internal/ops/opshttp/opshttp.go`, Task 2 (`a0afb06`) |
| R-6bii5 | `ErrForbidden`/`ErrInternal`, the classifier, `Codes` | `internal/ops/errors.go`, Task 1 (`6caccf1`) |

**R-6bii2's amendment (conductor, 2026-09-25, at the security fold
`629f3df`).** The plan as approved gave the authorizer the shape
`authorize(ctx, op string, p ops.Privilege) (ops.Actor, error)`. The
security review found this let an authorizer's own privilege argument
be echoed back as its grant. The fold changed the shape to
`authorize(ctx context.Context, op string) (ops.Actor, ops.Privilege,
error)`: the authorizer answers the caller's granted privilege, and the
authorizing seam checks it against `ops.PrivilegeOf(op)` on an
allowlist (admin reaches admin and member, member reaches member, any
other grant is denied). The master copy of the plan was amended at
`f3d407c` to match. This changes the plan's stated shape; the amendment
is the record.

### Task 0's routed note

`IntegrationValue` answers `""` for a disabled backup-integration class
(`cmd/dubplate-server/backup.go`). This routes to 6b-ii-b: R-6bii9's
disabled-class row must key on `IntegrationEnabled`, not on
`IntegrationValue`'s emptiness, since a disabled class and an enabled
class with an empty value would otherwise read the same.

### Mutations run, per task

- **Task 1**: four named mutations. Three fired (each on its own named
  test). The fourth, "key the canceled wrap on the error chain rather
  than `ctx.Err()`", did not fire. Recorded as reviewer-read, which the
  plan allows: no production path today separates `ctx.Err() ==
  Canceled` from an error chain holding `context.Canceled`, because
  every seam context derives from the caller's own and `storeContext`
  uses `context.WithoutCancel`.
- **Task 2**: 8 named mutations, all fired.
- **Task 3**: 4 named mutations, all fired.

### The pass-end review

`web-auth-security-reviewer` ran over `internal/ops/opshttp` and the
`internal/ops` actor and classifier, briefed that the projection is
unmounted and that rung 5's admin gate is the comparison surface. It
found no critical or high issues. Of its findings: M1 and L1, L2, L4,
and L8 were folded at `629f3df` (the grant allowlist above, closed-set
denial logs, a 500 on an unmapped code, `CutPrefix` for the path match,
`Cache-Control: no-store`), and the fold's own `diff-reviewer` read
accepted it. L3 (job `FailureMessage` text) was refused: that text
reaches only admins, so the risk the fix would remove is smaller than
the fix's cost. L5 to L7 route to rung 14 (below).

### Both budgets

Tokens: **about 1.8M** against the **3.90M execution ceiling** with the
80% flag at **3.12M** (Ruling 1). Breakdown: Task 0 about 0.09M,
segment A1 (Tasks 1-3) about 1.20M (run `wf_0f376bd8-ee8`), the
canceled-mutation probe about 0.11M, the simplifier round about 0.12M,
the security review about 0.11M, its fold about 0.11M, the fold's
review about 0.07M, and this task. `<Task 8's own remaining spend, the
merge ritual's, and the architecture reads', filled at the fold>`.
Attended time: **zero attended interactions after approval**. Fix
rounds: Task 1 one, Task 2 none, Task 3 none.

---

## The errata this pass owes

- **D6 (R-6bii2).** `internal/ops/opshttp` imports `internal/identity`
  beside `internal/ops`, for the mount's channel type. Confirmed by
  `go list`, above.
- **R-6bii2's amendment** (above) is itself an erratum: the plan's
  authorizer signature as approved is superseded by the fold's shape,
  already reflected on the master plan copy at `f3d407c`.

No other deviation (D1-D5, D7-D10) touches Task 0-3's surface; D1
(the contract suite's in-process leg) landed as stated, without change.

---

## The 6b-ii-a to 6b-ii-b, rung 14, and rung 12 handoff

**To 6b-ii-b** (fresh Task 0, Tasks 4A-7):
- `ServeOptions.Reconcile` and `backup.Reconcile` remain 6d's wiring, not
  this half's.
- The store identity that would let restore refuse another store's
  snapshot: still unminted.
- The network actor's `member:<principal>` form against
  `Identity.UserID`, for restore's own actor needs if any arise.
- The routed note above: key R-6bii9's disabled-class row on
  `IntegrationEnabled`, not `IntegrationValue`'s emptiness.

**To rung 14** (the mount):
- The projection's mount point, `http.StripPrefix` **without** a
  trailing slash (`CutPrefix("/")` would 404 every op).
- The production authorizer: build a session authorizer answering
  `(Actor, Privilege, error)`, and never grant a zero or unchecked
  privilege.
- The bearer channel (CH, spec `:107-110`), mounted alongside the
  cookie channel; `ChannelCookie` is the only network channel
  `internal/identity` declares today.
- **Ruling 3's answer**: a verified member lacking an operation's
  privilege gets HTTP 403 with code `forbidden`; a caller with no
  identity is rung 14's own denial kind (401 or 404), since `opshttp`
  is unexported and can add a second kind at no cost to any caller.
- Security L5-L7, carried from the fold: audit logging of admitted
  operations with actor and `client_ip`; `CancelJob` discards the actor
  it receives, so a durable id belongs in `jobs.actor` rather than
  principal reuse; the unauthenticated caller's response shape (folds
  into Ruling 3's answer above).
- Before the cookie mount lands: the 5c-ii security M2 question applies
  to the projection too — the roster API shares an origin with the
  Navidrome fallthrough, and `originCheck` admits
  `Sec-Fetch-Site: same-origin` (`internal/serve/origin.go:13-24`); the
  bearer mount should go behind `addBearerBypass` only.

**To rung 12**: the streamer database's place in a restore (2.0 note:
Geoff re-keyed versions 2026-09-25; the streamer formerly "2.0" is now
0.8, so this is a 0.8 forward-hook reference).

---

## What a later pass would be wrong to rediscover

- **R-6bii2's authorizer shape changed mid-pass** (above): a reader of
  the plan's Task 2/R-6bii2 prose alone would build the three-argument
  `authorize(ctx, op, p)` signature the fold replaced. The amended
  master plan (`f3d407c`) and this entry are the record; the shipped
  code is the two-return-value-plus-error form.
- **Mutation 4 on Task 1 is architecturally unfireable today**, not a
  gap in the test: no seam context in this codebase separates
  `ctx.Err() == context.Canceled` from an error chain already holding
  `context.Canceled`, since every seam context derives from its
  caller's and `storeContext` strips cancellation with
  `context.WithoutCancel`. A future seam that builds its own
  independent context would need this row re-checked.
- **`Codes` and `CodeOf` have a real caller today** (`opshttp`'s
  constructor and its error encoder) but no binary reaches either until
  rung 14 mounts the projection; the exported-unused gate's 125 count
  already reflects this and should not be read as "unused" in the
  ordinary sense.
