# Rung 14a spec fold verification

Target: `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md`, revision 2 at `42374b4`, against revision 1
at `9d730d7`. Inputs: the fold record (`research/2026-10-10-rung-14a-spec-fold.md`) and the four lens reviews. The
reader had no part in the fold. Code was read at `42374b4`. kin-openapi claims were read in the module cache
(`github.com/getkin/kin-openapi@v0.149.0`) and probed in a scratch module (`scratchpad/kinprobe`, `go run` against
the cached module). Spec line numbers are revision 2's (`spec:N`).

Findings are ranked by consequence and limited to correctness and the stated requirements.

## 1. Did each blocker and major close where the fold says?

Yes, every one closed at its cited location. Two folds left a residual, filed below as V1 and V2.

| Finding | Closed at | Verified against |
|---|---|---|
| D1 (blocker) | spec:133-140, AC6 (:526-527), AC7 (:528-530), "Deferred past 14a" (:596) | `opshttp.go:72-77` routes all four operations; `CatalogEntry.Available` and `Reason` exist (`catalog.go:96-101`); the serve unit grants `auth` and `store` (`dubplate-server-serve.service:43`); R6:942-944 routes the runner to rung 14. Residual: V1, V4. |
| A1, B3, D2 | spec:163-166, :232-236, privileged row :446 | `authorize.go:19-21` takes `(ctx, op)` only; `admit` wraps every denial as forbidden (`:97-103`); `Actor` holds wire and channel only (`ops.go:24-27`). Residual: V6. |
| A2 | AC2 (:512-515) | `opshttp.go:88-114` transport refusals carry the JSON error body. |
| A3, B2 | spec:51-55, :66-67, AC2 | `validate_response.go:51-56` quoted exactly; `:22-35` and `:99-117` are the HEAD, redirect, no-content, and no-schema passes; `Options.IncludeResponseStatus` exists (`openapi3filter/options.go:22-24`). Empty security passes: `ValidateSecurityRequirements` returns nil for zero requirements (`validate_request.go:402-405`). |
| A3 item 2 (refused) | spec:56-58 | Correctly refused. `ValidateResponse` appends `EnableJSONSchema2020()` for a 3.1 document (`validate_response.go:76-77`), and so does the request side. The spec cites `validate_request.go:247-248`, which is `ValidateParameter`; the request body's site is `:359-360`. |
| B1 | spec:59-65, AC1 | Reproduced. A 3.1 schema with `dependentRequired` and a nested `$ref` validates a violating value as nil; without the `$ref` it refuses. The fallback is `schema_jsonschema_validator.go:290-297`. Residual: V3. |
| A4 | spec:83-97, AC19, rider #23 | `opshttp.go:230-243` body shape; `+page.svelte:46-54` keys focus on outcome names. Residual: V5. |
| A5, D5 | spec:146-148, AC17 | `opshttp.go:140-144`: "Unknown keys are ignored". |
| A6 | states table spec:449-470 | Residual: V2 (many-bad), V5 (busy). |
| A7, C2 | spec:206-212, AC10 | `evaluate.go:292-294` refuses an `admin` row the environment contradicts; R6:809. |
| A8 | AC16 (:563-567) | Feasible as written. A raw second connection holding `BEGIN IMMEDIATE` with one demotion uncommitted makes the production `Update` wait on the busy timeout. A guard read before `Update` sees the committed state and lets both demotions through, so the proof can fail. |
| A9 | spec:235-238, AC12, rows :439-440 | `seam.go:104-105` "never a default". |
| B4, B5, C4 | spec:255-270, AC14 | `write.go:78-84` refuses any `user_version` but `h.version`; `open.go:179` sets it to `len(files)`; R-6d2 at `HISTORY.md:411-416`. |
| C1 | spec:196-199, AC11 | `accessor.go:113-116` returns the default for a missing row. Residual: V2. |
| C3 | spec:303-308, AC15 | `roster.go:415-417` reactivates under the old id; `backup/doc.go:1-9` carries no auth leaf. |
| D3 | spec:91-97, :110-127 | `serve.go:96-98` mounts both patterns without stripping; `admin.go:125-128` answers `http.NotFound`. |
| D4 | spec:474-496 | Rows cover `internal/config`, `internal/ops`, and `internal/auth`. |

## 2. Contradictions and build order

The D1 fold does not contradict AC1 or AC7. The job operations stay catalog entries, so AC1's bijection still
requires their paths. AC7's `conflict` answer is a declared status the mount never emits. Under
`IncludeResponseStatus`, a declared but unemitted status is legal. Two build-order gaps follow from the folds (V1, V2),
and the D1 fold leaves the job types unexercised (V4).

## 3. Mechanisms stated from memory

- kin-openapi: every quoted behavior checks out against the source and the probe, with one exception. The fold's
  enforced-keyword list includes `format`, but the built-in validator enforces only the formats registered in
  `schema_formats.go:56-62` (`byte`, `date`, `date-time`, `int32`, `int64`). The probe validates `"not-an-email"`
  against `format: email` as nil, both with and without a nested `$ref` (V3).
- `Handle.Update`'s version guard is as stated (`write.go:62-63`, `:78-84`).
- `ErrEpochBackward`: the spec's reason is overstated (V7).
- `Bump`'s synchronous notification under a 5s per-subscriber deadline is as stated (`epoch.go:103-138`).

## 4. Can the owner rule the forks?

The fold rightly names none. None of the three it ruled is a genuine product fork.

- **A7 and C2.** The ratified seed table (R6:809) already makes an `admin` row beside a different environment value a
  start failure. Refusing at the write is the only option that keeps the table and removes the outage. Option (c)
  would reopen a ratified ruling. The product cost is narrow: the env template sets no store-homed setting
  (`env/dubplate-server.env.template` carries only wall and integration rows), so on a default install every setting
  stays editable from the web.
- **A10.** D10's own reason ("changes who holds access") covers an add. The cost is one sign-in per window, which
  the window's researched default bounds. "Rulings for Geoff" reports it, which is the no-ceremony form.
- **B5.** R-6d2's rollback path recovers the state without a write mode that bypasses `Update`'s guard. The
  allowance's only gain is operator convenience in a post-launch, narrowed-bound case.

## Findings

### V1. major. The spec names no `ops.Seam` implementation for `serve`, and `local.Runner` cannot take the seven new methods as it stands

- **Location:** spec:129-131 ("Each new operation joins `ops.Seam`"), spec:137 ("`serve`'s seam answers ... with
  `unavailable`").
- **Defect:** `ops.Seam` is "the four operations" interface (`internal/ops/ops.go:164-189`). Its only implementation
  is `local.Runner`, which asserts `var _ ops.Seam = (*Runner)(nil)` (`internal/ops/local/local.go:92`). `local.New`
  takes an instance lease and runs job recovery (`local.go:94-97`), which is exactly what D1's fold keeps out of
  `serve`. So `serve` needs a new implementation that the spec never names: no type, no home, no owner row. Adding
  seven methods to the interface also breaks `local.Runner`'s build. The spec does not say what the Runner answers
  for them. The CLI member verbs call `auth.NewRoster(h)` directly, not through the Runner
  (`cmd/dubplate-server/member.go:97`). R6:627-636 requires every asymmetry the catalog reports to be named, and a
  Runner answering `unavailable` for member and settings operations would be an unnamed one. The first plan task
  meets this. It is an architecture call the plan reviewers would otherwise grade mid-pass.
- **Proposed fold:** add one paragraph to "The mount" that names the shape. Recommended: the member and settings
  operations form a second interface beside the job seam, and the authorizing decorator and `opshttp`'s routes take
  both. A new 14a type, owned by `internal/ops` under the contested-surfaces row, implements the admin interface for
  `serve` over the roster and the settings write path. A small type answers the three job methods with `unavailable`
  and the reason the catalog carries. `local.Runner` stays a job runner with no stub methods. If the plan keeps one
  interface instead, the spec names the Runner's answer to each new method and that asymmetry.

### V2. major. The one validator makes `settings reset` unable to repair a store with two refusing rows

- **Location:** spec:196-199 ("refuse when it refuses"), spec:257-258 ("under the same validator"), AC14
  (:553-556), the states table.
- **Defect:** the seed rule's evaluation returns one `Evaluation.Err` that joins every key's refusal
  (`internal/settings/evaluate.go:243`). Take a store that the next start refuses for two `admin` rows, for example
  two keys whose bounds a new binary narrowed with no migration. Resetting either key leaves the other refusing, so
  the post-write evaluation refuses and the verb writes nothing. The verb exists to resolve a start refusal, and in
  this state it resolves none. Its messages still name it. The rollback binary is the only way out. The C1 fold
  introduced this: revision 1's verb ran no validator. AC14 tests each refusal alone, so the defect passes it. This
  is the "many-bad" state that CLAUDE.md's planning-miss rule names.
- **Proposed fold:** refuse a write only on a refusal the write introduces. The evaluation already carries a
  per-key `Outcome.Err` (`evaluate.go:76-108`), so the check compares the pre-write and post-write evaluations and
  refuses on any refusal the post-write one adds. This rule also covers the web path. AC14 gains a fixture with two
  out-of-bounds `admin` rows, reset one at a time, after which `serve` starts.

### V3. minor. The enforced-keyword allowlist omits `$ref` and the annotations, and lists `format`, which is mostly unenforced

- **Location:** spec:62-65, AC1.
- **Defect:** read literally, the walk "fails on any other schema keyword", so it forbids `$ref`. The B1 finding
  presumes `$ref` throughout the document. The walk also forbids `description`, `title`, `minimum`, `maximum`,
  `minLength`, `maxLength`, `pattern`, and `minItems`. The built-in validator enforces each of these, and a readable
  contract needs the annotations. Meanwhile `format` is listed as enforced, yet only `byte`, `date`, `date-time`,
  `int32`, and `int64` are registered (`schema_formats.go:56-62`). The probe validates `format: email` against
  `"not-an-email"` as nil. The implementer must improvise the carve-out, and a `format: email` or `format: uuid`
  passes silently.
- **Proposed fold:** split the list into the enforced validation keywords (add `minimum`, `maximum`, `minLength`,
  `maxLength`, `pattern`, `minItems`, `maxItems`, `uniqueItems`, `not`) and the annotation keywords the walk allows
  (`$ref`, `description`, `title`, `example`, `deprecated`). The walk also refuses a `format` outside the five
  registered ones. Correct the request-side citation to `validate_request.go:359-360`.

### V4. minor. The job types are published at 14a, but no answer exercises them

- **Location:** spec:101-106, AC7.
- **Defect:** after D1, the three job operations answer only 401, `forbidden`, `invalid_request`, or `unavailable`
  on the only mount. The handle, the six states, the phase, the failure fields, and the four timestamps are written
  into the document, but no routed answer ever validates against them. Those schemas can drift from `ops.go:86-161`
  unseen until the runner's rung. This undercuts AC2's purpose.
- **Proposed fold:** one AC7 clause: a test marshals `ops` job response values (every state, a failure, every
  timestamp set) and validates each against the document's schema directly.

### V5. minor. The add path's `busy` outcome, and a store-busy answer on every write, map to `invalid_request`

- **Location:** states table spec:464 ("`invalid_request` otherwise").
- **Defect:** the retiring roster API answers `busy` when the write lock or the busy timeout gives way
  (`internal/auth/admin.go:44`, `:310-311`), and the page tells the member to retry. The table sends every non-duplicate
  add refusal to `invalid_request` (400), which tells a client its request is wrong. No row names lock contention for
  remove, demote, set, or reset. `admin_not_grantable` also becomes unreachable once the `admin` key is ignored.
- **Proposed fold:** one row: any write whose store write lock times out answers `unavailable` with detail `busy`.
  `admin_not_grantable` leaves the detail enumeration.

### V6. minor. The add path's in-write admin re-read keys on the actor's address, but the actor no longer carries one

- **Location:** spec:163-166, :218-219, :290-291.
- **Defect:** `addUser` re-reads the actor's flag with `WHERE address = ?` on the actor string
  (`internal/auth/roster.go:382`). The folded `Actor` exposes `UserID` and the anchor, not an address. `add_member`
  over `/api/v1` cannot pass the current argument without a wire-string parse or a second identity read.
- **Proposed fold:** one sentence: `AddMember`'s actor becomes the `UserID`, re-read by id like removal and the
  settings writes.

### V7. minor. The `ErrEpochBackward` reason overstates what the stored epoch revokes

- **Location:** spec:316-318.
- **Defect:** a backward bump writes nothing (`epoch.go:84-87`). The stored epoch E revokes only credentials
  anchored at or before E. A session anchored after E but before the clock stepped back survives the epoch. The
  tombstone still refuses it while the member is removed (`roster.go:257`). A re-add while the clock is still behind
  that anchor hits `ErrEpochBackward` again, and the session verifies. The window is narrow, and committing is still
  right, but the sentence "revokes everything the removal must" is false.
- **Proposed fold:** restate the reason. The stored epoch is later than now, so committing loses no revocation that
  a write could have made. The residual (a session anchored between that epoch and the step) is named and accepted.

### V8. minor. Closed response schemas in the published document contradict its own additive policy

- **Location:** spec:66-67 against :74-77.
- **Defect:** the policy says a new response field is additive and that a client must ignore unknown fields. The
  document, by publishing `additionalProperties: false`, tells any client generated from it that an unknown field is
  invalid. The strictness exists for the tests.
- **Proposed fold:** the contract test closes every response object schema when it loads the document, before
  validating. The published file stays open. The rename and missing-field proofs keep working.

## 5. Over-ceremony in the growth

The fold grew the spec from 419 to 609 lines. Most of that growth is what CLAUDE.md's planning-miss rule requires: the
states table, the quoted library behaviors, and the privileged-field rows. The contested-surfaces table has no
execution cost, and it removes merge-time discovery. Two items cost execution work out of proportion to the state
they cover:

1. **The two-value reset answer** (spec:216-217, AC11's last sentence, AC20's matching page state). It needs an answer
   field, page copy, a fixture, and a render-and-read state. The state it serves is an `admin` row equal to the
   serving environment's value. The only web path that creates one is "writing the environment's own value back
   commits" (spec:211). That path is useless, because the value is already in force. It also arms the start refusal
   at the operator's next environment edit. Fold: a `set_setting` equal to the serving environment's value writes
   nothing and answers success naming the source in force. The two-value answer then drops. Spec:210 ("offers no
   edit") then also covers reset on such a key, which resolves the ambiguity between spec:210 and spec:216-217.
2. **`doctor`'s three-way warning** (spec:363-367, AC21). Revision 1's single "names no live admin" warning needed the
   same new `Inspect` fact. Splitting it into non-admin, removed, and no member adds three fixtures for one operator
   message. This is low cost. Keep it only if the messages differ in the action they name.

## Counts

0 blockers, 2 major (V1, V2), 6 minor (V3 to V8), and 2 over-ceremony items. Every blocker and major from the four
lenses closed at its cited location. No fork needs the owner.
