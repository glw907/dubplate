# Rung 14a spec fold

Target: `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md`, revision 1 at `045aadc`, folded to revision 2.
Reviews: `research/2026-10-10-rung-14a-spec-review-{contract,mechanics,integrity,consistency}.md` (lenses A, B, C, D).
Code read at `045aadc`; no file under `internal/`, `cmd/`, or `web/` changed since `e6fdf2b`, the spec's citation
commit.

Each finding was checked against the tree before it was folded. Duplicates across lenses share one row, which lists
every ID. "Refused in part" means the defect is folded and the finding's proposed fix is declined for another; the
reason is given. Owner rulings 1 to 3, the 14a/14b split, the seed-only `DUBPLATE_ADMIN_EMAIL` ruling, and the
TypeScript 7 sweep ruling were not reopened.

## Dispositions

| IDs | Severity | Verified against | Disposition |
|---|---|---|---|
| D1 | blocker | `opshttp.go:72-77` routes all four operations; `serve.go` holds no `ops.` reference; R6:943-944 (6d-R6) routes job recovery to rung 14; the serve unit grants `auth` and `store` only (`:43`) | **Folded**, "The mount": the three job operations answer `unavailable` on the cookie mount and the catalog reports them unavailable with a reason (`CatalogEntry.Available` and `Reason` already exist, `catalog.go:96-101`); the runner routes to the first rung whose page starts a job; AC6 proves the actor on `settings_audit`; AC7 keeps the document clause and adds the `unavailable` answer. Ruling 2 decides it, so no fork. |
| A1, B3, D2 | major | `authorize.go:19-21` (no request reaches the authorizer), `:100-103` (every denial wrapped as `forbidden`); `ops.Actor` holds wire and channel only (`ops.go:24-27`) | **Folded**, "The network actor" and "The re-authentication gate": `NetworkActor` takes the verified identity and the anchor, `Actor` keeps `UserID` and anchor unexported behind accessors, the seam implementation decides the gate from the anchor and the key's mark, and the privileged-field table gains the freshness row. |
| A2, B7 | major, minor | `opshttp.go:88-114` transport refusals; an unrouted path has no operation to validate | **Folded**, AC2 split into well-formed, schema-malformed, and unrouted or transport cases, the last validated against the error schema directly (A2's form, stronger than B7's status-only check). |
| A3, B2, D9 | major, major, minor | `validate_response.go:51-56` and `:22-35`, `:99-117` (silent passes); `conformance.go:96-136` precedent | **Folded**, "The contract document" and AC2: `IncludeResponseStatus: true`, every other silent pass a failure, closed response schemas with required fields, rename proof on a `required` field, a missing-Go-field proof, an undocumented-status proof, and no `security` requirement (B2's third point). **Refused in part:** A3 item 2 and D9's 2020-12 option clause, because `openapi3filter` enables the 2020-12 validator itself for a 3.1 document (`validate_request.go:247-248`, `validate_response.go:76-77`); only a direct `VisitJSON` passes it, and the spec says so. |
| B1 | major | `schema_jsonschema_validator.go:20-56` compiles a standalone resource and `:290-297` falls back silently on a failed compile; lens B's probe measured three 3.1-only keywords validating nil beside a nested `$ref`. Authoring 3.0.3 does not escape it: the 3.0 `Schema` struct parses the same 3.1 fields (`schema.go:650-667`) | **Folded**, option (a): the document uses only keywords the built-in validator enforces, and a contract test walks it and fails on any other (AC1). |
| A4, B11 | major, minor | members page keys copy and focus on outcome names (`+page.svelte:46-54`); opshttp body is `{"error":{"code","reason"}}` (`opshttp.go:230-243`); `postForm` posts urlencoded and opshttp refuses non-JSON (`:97-101`); the only JSON 403 `postForm` receives is the retiring roster API's | **Folded**, "The error body and its two new codes", the planning-miss table, AC19, and the #23 rider: optional enumerated `error.detail`, `reauthentication_required` (403), each add outcome kept as a `detail`, focus moves asserted, and a new JSON client for `/api/v1` that leaves `post.ts` unchanged, so #23 closes by retirement (B11's form). |
| B8, D3 | minor, major | `serve.go:96-98` passes the full path; `http.StripPrefix` mismatch is plain text; `RequireAdmin` answers plain 404 (`admin.go:125-128`); HISTORY:2821-2839 lists the mount items | **Folded**, "The mount": the handler cuts `<prefix>/api/v1/` itself, routing to `not_found` precedes session verification, a member-admitting middleware answers 401 `unauthenticated`, the authorizer lives in `opshttp` reading a context carrier the plan names, one log record per admitted operation (L5 to L7), and M2 stays with slice 6. |
| A5, B6, D5 | major, minor, major | `opshttp.go:141-144`: "Unknown keys are ignored"; R6:585-586 | **Folded** as the ignore rule (B6, D5): a body naming `admin` adds the member without the flag; AC17 rewritten. **Refused in part:** A5's per-operation refusal, because it needs a strict decode path the package contract rules out, and a client following the document never sends the key. |
| A6 | major | the states listed in revision 1 had no answers | **Folded**, the planning-miss states table and AC18. Demoting a non-admin writes nothing and answers success; self-removal commits and the bump ends the caller's session. |
| A7, C2 | major (both marked owner fork) | R6:809 and `evaluate.go:292-294`: an `admin` row against a different environment value refuses the start | **Folded, not forked:** the write is refused with `conflict` and detail `environment_sets`. R6's ratified seed table already makes a durable admin value against an environment value impossible, so refusing at the write keeps the ratified rule and moves the failure from an unattended restart to the write. No answer leaves the table intact while committing. |
| A8 | major | a guard outside the transaction also passes a sequential run | **Folded** in AC16: the proof holds the write lock from a second connection with one demotion uncommitted, starts the other, then commits, and runs once against a misplaced guard. **Refused in part:** A8's barrier hook before `Update`, a production test seam; the store's own lock forces the interleaving without one. |
| A9, C7 | major, minor | revision 1's privileged row for an `admin` row's removal omitted the window; `seam.go:104-105` "never a default" | **Folded**, the gate's scope and AC12: a sensitive reset is gated, counter-cases prove a stale non-sensitive write and reset commit, and an unreadable window refuses with `internal`. |
| A10 | minor (marked owner fork) | revision 1's D10 reason, "changes who holds access", covers an add | **Folded, not forked:** `add_member` joins the gate (D10). An add from a borrowed cookie grants access that outlives the cookie, and a re-add clears a tombstone an admin set, so it is the worst case D10's reason names; the cost is one sign-in per window on an occasional action. The tombstone rows now name both channels. |
| A11, A12, B9, B10, C6, D14 | minor | `epoch.go:82-88` (`ErrEpochBackward` writes nothing and the later epoch stands); `:237`, `:265` refuse at an equal second; `:98-101` notifies synchronously | **Folded**, "The epoch bump" and AC15: `ErrEpochBackward` commits the removal, the fixture clock steps one second past the bump, and `Subscribe`'s doc comment gains the must-not-block contract (B10's first option; notifying after the commit was not needed with no subscriber). |
| A13 | minor | criteria 6, 9, 13, 18 could pass on empty fixtures | **Folded**: sentinel wall values (AC9), a removed admin who receives no notice and a non-sensitive write that mails no one (AC13), a written preference row (AC15), and a rename by direct store write (AC6). |
| A14, B12, C1, D8 | minor, minor, major, minor | `accessor.go:113-116` returns the default for a missing row; `evaluate.go:217-238` checks correlated rules at start; R6:1002 gives `env-seed` to the seed rule alone | B12 and D8 **folded**: the reset answer and page name the default now in force and the value the next start seeds. C1 **folded in part**: both writes run the seed rule's own evaluation over the post-write rows and the serving environment, which closes the correlated-rule gap. **A14 refused**, with C1's re-seed half: writing an `env-seed` row from a reset adds a channel R6:1002 reserves for the seed rule, and after A7/C2 an `admin` row can sit beside an environment value only when the two agree, so the transient default is narrow and named. |
| A15 | minor | `evaluate.go:264-266` refuses an unparseable `admin` row | **Folded**, "The resolving verb" and AC14: three refusals, each naming the verb. |
| A16, B14 | minor, minor | SS:1017-1019 asks for a description and a mark on every registered key; R-6d6 | **Folded** with B14's definition: on a key only the environment sets, the mark records a change that needs care at a restart; secret-ness stays `init`'s. **Refused in part:** A16's description-only option for wall and integration rows, which would contradict SS hook 18. |
| A17, C13 | minor | `0005_routed.sql` keys both tables on the user id | **Folded**: overrides and quota survive removal and return with a re-add; AC15 asserts it. |
| A18 | minor | a window at or above the absolute cap makes the gate a no-op | **Folded**: a correlated rule below the cap and a floor above a link round trip. |
| A19 | minor (over-ceremony) | `registry.go:132-247` already panics in every `Add*` | **Folded**: AC8 keeps a panic proof and leans on the existing construction test for the whole registry. |
| A20 | minor (over-ceremony) | criterion 18 repeated criterion 15's fixture | **Folded**: AC15 is the checklist test; AC18 now asserts the states table. |
| B4, B5, C4, D13 | major, major (marked owner fork), major, minor | `write.go:62-63`, `:78-84`: `Update` refuses any version but `h.version`, which `OpenWrite` sets to head (`open.go:179`); R-6d2 (`HISTORY.md:411-416`) keeps an older store unmigrated for the rollback binary; no pre-14a store can hold an `admin` row | **Folded at the root, not forked:** the older-store allowance is dropped. The verb opens through `OpenWrite` at the binary's version, and `serve`'s pre-open refusal names the rollback recovery. An allowance needs a write path that bypasses the store's most protective check, against a recovery R-6d2 already designed, for a case only an upgrade that narrows a bound can reach. C4 also corrects revision 1's claim that an environment contradiction cannot be fixed through the environment; the message says how (`evaluate.go:293`). |
| B13 | minor | `PROVENANCE.md`: upstream 3.0.0 file pinned by sha256; `arch_test.go:213-217` | **Folded in part**: the precedent sentence is corrected. **Refused in part:** the non-test checker package, because the contract tests live in the mount's one package and write the classification once; a new package would widen the no-link test's list for no measured need. |
| B15 | minor | `inspect.go:166-171` carries counts only | **Folded** as a plan note in "`DUBPLATE_ADMIN_EMAIL`". |
| C3 | major | `internal/backup/doc.go:1-9`: a backup carries the library and the store, not the auth leaf; `roster.go:415-417` revives the old id | **Folded**, "The epoch bump": the reactivation branch of `addUser` bumps too, in the same callback; AC15 proves it with the removal's record deleted. Uses the existing `Bump`, no new state. |
| C5 | minor | `seam.go:132-141` reads `Admin` once at verification | **Folded**: each settings write re-reads the actor's admin flag by `UserID` inside its transaction; AC16 extends to settings writes. |
| C8 | minor | revision 1 implied but did not state the send's position | **Folded**: the notice sends after `Update` returns, under its own deadline; the audit row is the durable record. |
| C9 | minor (over-ceremony) | the seed-rule exemption rests on box access, as the command line does (`identity.go:155-161`) | **Folded**: the command line sends no notice. Owed erratum 5. |
| C10 | minor | `member.go:96` logs to the verb's stderr | **Folded**: the spec states that a CLI removal's record is a terminal line, and accepts it. |
| C11, D7 | minor, minor | runbook step 10 uses a placeholder address (`server-runbook.md:164-170`); key description at `:59-60`; `doctor_test.go:1023`; `addUser` reactivation keeps the candidate's flag | **Folded**, "`DUBPLATE_ADMIN_EMAIL`" and AC21: the runbook edits are in 14a's scope, step 10 runs only when `member list` shows no admin, the existing doctor row carries the warning, AC21 reads "grant or authorize", and the 5d plan citation is replaced by the runbook's. |
| C12 | minor | revision 1 named the states and no outcome | **Folded** with A6's answer: no row or an `env-seed` row writes nothing and answers success naming the source in force. **Refused in part:** C12's `invalid_request`, because a reset is idempotent and the page never offers one on a non-admin row. |
| C14 | minor | a same-box restore replaces the store and keeps the auth leaf | **Folded in part**: the runbook gains the re-run-removals line. The rung 16 check C14 proposes is routed to the conductor for BACKLOG; it is not 14a's. |
| C15 | minor | the limiter is built once at start | **Folded**: a `restart` key's stored value is named as the one the next start applies. |
| D4 | major | rung 7 brief section 2 rows for `internal/auth`, `internal/serve`, config registration; PA:173 form | **Folded**, "Contested surfaces with the rung 7 lane", as an architecture decision: 14a owns `internal/ops`, `opshttp`, the document, the hook-18 declaration types, and `roster.go` and `epoch.go`; rung 7 owns `RequireAdmin`, its member page gate, and the styled 404's hooks in `admin.go` and `Shell.terminated`; 14a writes nothing in `internal/serve`; the second merger rebases. Owed erratum 6. |
| D6 | minor | R6:913-915 names the username "the network actor" | **Folded**: one sentence in D5 names the superseded clause and the two comments the plan edits. Owed erratum 1. |
| D10 | minor | each of the seven re-read | **Folded**: `catalog.go:78-83`, R6:185-186 with "rung 7 or rung 10", CH:139-142, BS:985-988, the hook row cites SS's own text, the ruling 8 clause is dropped, and `roster.go:304-305`. |
| D11 | minor | CLAUDE.md's prior-art list names goose | **Folded**: goose row, neutral. |
| D12 | minor | RE:39 left the (14,13) edge open | **Folded**: one sentence under the version policy. |
| D15 | minor | `visual-identity-design.md:145-146` reserves a later member settings page | **Folded**: AC20 names the admin settings page and keeps it off the `settings` route; the plan names the route. |

## Owner forks

None. Lenses A, B, and C marked four findings as owner forks: A7 and C2 (one question), A10, and B5. Each has a
dominant answer, recorded in its row: the ratified seed table settles A7 and C2, D10's own reason settles A10, and
R-6d2's rollback path with the store's version guard settles B5. The spec's "Rulings for Geoff" says so.

## Why the fold refuses little

One finding is refused in full (A14) and eight in part. Every major was a verified gap between the spec and the
code, and most minors were states or fixtures with a one-line conventional fix. The four over-ceremony findings
(A19, A20, C9, and the B5 and C4 pair) were folded because each removes scope. The refusals sit where a fix adds a
channel, a seam, or a package that the conventional fix does without.

## Owed errata (ratified documents this revision changes in meaning; not edited here)

1. **R6:913-915** (erratum 6d-R2 and 6d-R14): "the username is the readable name in logs and the network actor"
   loses "and the network actor"; the network actor keys on `UserID`.
2. **R6:589-590**: the error vocabulary gains `unauthenticated` (401) and `reauthentication_required` (403), and the
   error body gains an optional enumerated `detail`.
3. **R6:943-944** (erratum 6d-R6): job recovery and `serve`'s runner route from rung 14 to the first rung whose page
   starts a job over HTTP.
4. **R6 criterion 7 and erratum row 10**: the dual-key epoch clauses retire; the epoch stays on the address
   (R-6d20). Carried from revision 1, which stated it without recording it.
5. **CH:171-172**: "a sensitive change notifies the admin address" reads as notifying every live admin, for a
   cookie-mount change only; the seed rule and the command line send none.
6. **PA's contested-surface table** (`specs/2026-09-03-parallel-build-approach.md:173-215`): the `internal/auth` row
   ("no other lane writes it") gains the 14a and rung 7a split; new rows for `internal/ops` with `opshttp` and the
   contract document, and for the hook-18 declaration types. Appended as conductor-ruling rows, PA's own form.
7. **BS:983**: rung 14 re-banded from M to L by the split. Carried from revision 1.

## Conductor-owed

- STATUS: carry the routing of `serve`'s runner, instance lease, and job recovery (erratum 3).
- BACKLOG: C14's rung 16 check, which lists each live member whose epoch record is later than the restored
  snapshot's stamp.
- Lane registry: the spec's contested-surfaces rows before either lane launches. The `internal/serve` row assigns
  rung 7's `Shell.terminated` edit under PA's rung 12 row; whether rung 7 and a running rung 12 slice may both
  write `internal/serve` is outside 14a and still needs its ruling.

## Measures

- New-mechanism findings folded: 3.
  - B1, the schema-keyword walk. Source: kin-openapi's own fallback (`schema_jsonschema_validator.go:290-297`) and
    this repo's `conformance` package, which already classifies each kin-openapi silent pass and fails on it
    (`conformance.go:96-136`). Measured defect: lens B's probe, with three 3.1-only keywords validating nil beside a
    nested `$ref`.
  - A4 with D3, the `error.detail` field and two wire codes. Source: Google's AIP-193, whose `ErrorInfo.reason` is a
    machine-readable identifier finer than the status code, and RFC 9110's 401 for a missing credential. Measured
    defect: the members page keys copy and focus on outcome names (`+page.svelte:46-54`) that the opshttp body cannot
    carry, and `admit` folds every denial into `forbidden` (`authorize.go:100-103`).
  - A1 with B3 and D2, the anchor and `UserID` on `ops.Actor`. Source: R6:583-584, which makes the verified identity
    the network caller's method parameter, and opshttp's own actor-from-authorizer design. Measured defect: no
    carrier reaches the decision point (`authorize.go:19-21`, `ops.go:24-27`).
- New-mechanism findings refused: 4: A8's barrier hook (the store's lock suffices), B4 and D13's older-store write
  mode (dropped at the root), A14 and C1's reseed channel (R6:1002), and B13's non-test checker package.
- Findings: 65 IDs in 40 rows. 64 folded (8 of them refused in part), 1 refused, 0 forked.
- Spec length: 419 lines before, 609 after. The growth is the states table, the contested-surfaces table, and the
  quoted kin-openapi behaviors.
- Ceiling: no change. The spec sets none and derives it from the brief's section 6. The fold moves scope both ways:
  it drops the older-store write mode and keeps the runner out, and adds the error detail, two codes, the re-add bump,
  and the keyword walk.

## Second fold

Input: the verification read (`research/2026-10-10-rung-14a-spec-fold-verification.md`), findings V1 to V8 and its two
over-ceremony items. Each was checked against the tree at `483057c` before it was folded. The spec moves to revision
2.1.

| ID | Severity | Verified against | Disposition |
|---|---|---|---|
| V1 | major | `local.go:92` asserts `var _ ops.Seam = (*Runner)(nil)`; `local.New` takes an instance lease and recovers jobs (`:94-97`); `ops.Seam` holds the four job and catalog methods (`ops.go:164-189`); `internal/ops` imports no `internal/auth` and nothing under `internal/auth` or `internal/settings` imports `internal/ops`, so a sibling package can implement over the roster without a cycle; `member.go:97` calls `auth.NewRoster(h).AddMember` directly; `roster.go:379-388` skips the in-write admin re-read on the command line | **Folded**, "The mount", with the verifier's shape: a second interface in `internal/ops`, the admin seam, holds the seven operations; the authorizing decorator and `opshttp` take both seams; the `catalog` answer joins the job seam's entries with one per admin operation through `BuildCatalog`; a new 14a package under `internal/ops` implements the admin seam for `serve`; a small type answers the three job methods `unavailable`; `local.Runner` gains nothing. The CLI verbs keep calling the roster and the settings write function directly, which are the shared write path. The contested-surfaces `internal/ops` row names the new pieces. Owed erratum 8. |
| V2 | major | `evaluate.go:243` sets `Evaluation.Err` to `errors.Join` over every refusal; each key's `Outcome.Err` is per key (`:76-108`), and a correlated rule's refusal is written onto every key it names (`:231-240`) | **Folded**, "One validator": `set_setting` checks the written key's parse, bounds, and environment first; then every settings write evaluates the pre-write and post-write rows and refuses only when a key refuses after that did not before. The written-key check keeps a still-bad value from riding an existing refusal. AC14 gains the two out-of-bounds `admin` rows fixture, reset one at a time, after which `serve` starts; AC11 gains a valid write beside an existing refusal; the states table gains the row. |
| Over-ceremony 1 | n/a | revision 2's spec:211 let the web write the environment's own value as an `admin` row, the only web path to the two-value reset state | **Folded**: a `set_setting` equal to the serving environment's value writes nothing and answers success naming the environment. The two-value reset answer drops. A `reset_setting` of an `admin` row for a key the serving environment sets answers `conflict`, `environment_sets`, because deleting it would flip the value to the default and back at the next start; only an operator's environment edit can create that row, and `settings reset` still clears it. The page offers neither edit nor reset on such a key. AC10, AC11, AC20, and the states table follow. |
| Over-ceremony 2 | n/a | `doctor`'s three-way warning | **Not taken**: outside this fold's scope. Owed to the plan, which keeps the split only if the three messages name different actions. |
| V3 | minor | `schema_formats.go:56-62` registers only `byte`, `date`, `date-time`, `int32`, `int64`; the request body enables 2020-12 at `validate_request.go:359-360` (`:247-248` is `ValidateParameter`) | **Folded**: the validation keywords add `not`, `minimum`, `maximum`, `minLength`, `maxLength`, `pattern`, `minItems`, `maxItems`, `uniqueItems`; `format` is limited to the five registered; `$ref`, `description`, `title`, `example`, and `deprecated` are allowed and validate nothing; the walk refuses any other keyword or format (AC1). The citation is corrected. |
| V4 | minor | `ops.go:86-161` job types reach no routed answer on the only mount | **Owed to the plan**: one AC7 test marshals job response values (every state, a failure, every timestamp) and validates each against the document's schema directly. |
| V5 | minor | `admin.go:44`, `:310-311` answer `busy` on lock contention | **Owed to the plan**: a write whose store write lock times out answers `unavailable` with detail `busy`, on every write; `admin_not_grantable` leaves the detail enumeration. |
| V6 | minor | `roster.go:382` re-reads the actor by address; the folded `Actor` carries `UserID` and the anchor | **Owed to the plan**: `AddMember`'s actor becomes the `UserID`, re-read by id like removal and the settings writes. |
| V7 | minor | `epoch.go:84-87` writes nothing on a backward bump | **Owed to the plan**: restate the `ErrEpochBackward` reason (the stored epoch is later than now, so committing loses no revocation a write could make) and name the accepted residual, a session anchored between that epoch and the clock step. |
| V8 | minor | the version policy tells a client to ignore unknown fields, which a published `additionalProperties: false` contradicts | **Folded**: the published document leaves response objects open, and the contract test closes every response object schema in its loaded copy before validating (spec "The contract document", AC2). |

### Owed errata, continued

8. **R6:582**: "`internal/ops` defines one interface" becomes two, the job seam and the admin seam.

### Measures

- Folded: V1, V2, V3, V8, and over-ceremony 1. Owed to the plan: V4 to V7 and over-ceremony 2. Refused: none.
- Spec length: 609 lines before, 651 after. The growth is V1's paragraph and list, V2's validator rule, and three
  states rows (the environment's own value, a reset of an environment-set key, and the per-key comparison); the
  over-ceremony fold removed the two-value answer and its page state.
