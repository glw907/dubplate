# Rung 14a spec review, lens B: mechanics and feasibility

Target: `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md` at `9d730d7`. Every claim below was
checked against the source at `9d730d7`, the module cache (`github.com/getkin/kin-openapi@v0.149.0`), `go doc`, or a
probe module under the session scratchpad (`scratchpad/kinprobe`). Correctness gaps come first, ranked by
consequence. Over-ceremony follows. Owner rulings 1 to 3, the 14a/14b split, and the `DUBPLATE_ADMIN_EMAIL`
ruling are not reopened.

## Verified as stated

- `Shell.Register` refuses a name carrying `/` (`internal/serve/serve.go:120-121`), and it mounts both
  `<prefix>/<name>` and `<prefix>/<name>/` without stripping the prefix (`:96-98`).
- The kin-openapi doc quote is accurate. `openapi3filter` turns on the 2020-12 validator by itself for a 3.1 document
  (`openapi3filter/validate_request.go:247-248`, `:359-360`; `validate_response.go:76-77`).
- A refresh keeps the anchor (`internal/auth/session.go:138-141`). `OpenWrite` refuses an older store
  (`internal/engine/store/open.go:195-196`). `Bump` runs each subscriber under a 5s deadline (`epoch.go:98-101`,
  `:122`). The write pool begins `IMMEDIATE` (`conn.go:146`).
- `EpochStore.Bump` and `ops.NetworkActor` have no production caller (`git grep` at `9d730d7`).
- Lookup refuses a tombstoned row (`roster.go:257`). A re-add revives the old id (`roster.go:395-417`).
- Every family has an `Add*` entry point that already panics on a malformed declaration (`internal/config/registry.go:132-247`),
  so a registration panic for hook 18 is feasible. `NewRegistry` registers the static `ownedWall` rows through
  `AddWallRow` (`:74-76`), so those rows are checked as well.

## Correctness findings

### B1 (major): kin-openapi silently drops 3.1-only keywords in any schema with a nested `$ref`

Location: spec:41-47, spec:351-353 (AC2).

The 2020-12 path marshals the resolved schema and compiles it as a standalone resource at
`https://example.com/schema.json` (`openapi3/schema_jsonschema_validator.go:20-56`). A nested
`#/components/schemas/...` reference cannot resolve there, so the compile fails. On a failed compile the validator
silently falls back to the built-in 3.0 validator (`:290-297`, "Fall back to built-in validator if compilation fails").
The probe confirms it, with `VisitJSON(v, EnableJSONSchema2020())` on a 3.1 document:

| Schema | Violating value | Result |
|---|---|---|
| `dependentRequired` and no nested ref | `{"a":"x"}` | refused (2020-12 message) |
| `dependentRequired` and a nested `$ref` property | same | **nil** |
| `unevaluatedProperties:false` and no nested ref | an extra field | refused |
| `unevaluatedProperties:false` and a nested `$ref` | same | **nil** |
| `if`/`then` and a nested `$ref` | violates `then` | **nil** |
| type array, `const`, `prefixItems`, `additionalProperties:false`, nested `$ref` | violating | refused (built-in messages) |

A real admin API document nests references, such as a job inside a status answer or the error envelope inside every
refusal. Any 3.1-only keyword in such a schema then passes everything. "Supports OpenAPI 3.1" is true of parsing and
of the simple case only.

**Fold:** add this behavior to the spec's quoted-behaviors list (spec:303-305) and close it with one of these:

- (a) Keep 3.1 and restrict the document to keywords the built-in validator enforces: `type`, type arrays, `const`,
  `enum`, `required`, `properties`, `additionalProperties`, `items`, `prefixItems`, `oneOf`, `anyOf`, `allOf`, and
  `format`. Add a contract test that walks the document and fails on any other schema keyword.
- (b) Author the document as 3.0.3, the only path the repo has exercised.

Recommend (a). It costs one test and keeps 3.1's type arrays.

### B2 (major): `ValidateResponse` passes an undeclared status and a schemaless answer by default, so AC1 and AC2 can pass vacuously

Location: spec:44-45, spec:348-353 (AC1, AC2).

`ValidateResponse` returns nil for a status the operation does not declare, unless `IncludeResponseStatus` is set
(`openapi3filter/validate_response.go:51-56`, "By default, status that is not documented is allowed"). It also returns
nil for HEAD, for 301/304/307/308, for a response with no content, and for content with no schema (`:22-35`,
`:99-117`). The probe confirms that an undeclared 403 carrying `{"whatever":1}` validates as nil under default options.
AC1 says every wire code is documented, and AC2 says errors validate. Under default options, a refusal the document
forgot still passes. The repo already solved this once: `internal/opensubsonic/conformance/conformance.go:1-9` and
`:96-136` classify each silent pass and set `IncludeResponseStatus: true`.

The renamed-field proof in AC2 has a related precondition. A rename fails validation only when the response schema
is closed (`additionalProperties: false`) or the field is `required`. The probe shows that a closed schema reports the
rename ("property "applyMode" is unsupported").

**Fold:** the spec states three things:
- Response validation runs with `IncludeResponseStatus: true` and treats each kin-openapi silent pass as a failure, on
  the `conformance` precedent.
- Every response schema is closed, and every field that is always present is required.
- The document declares no `security` requirement, or the tests pass an `AuthenticationFunc`. Otherwise
  `ValidateRequest` returns `ErrAuthenticationServiceMissing` (`validate_request.go:19`, `:438`).

### B3 (major): the re-authentication gate has no seat in the `opshttp` mechanism as built

Location: spec:135-148, spec:374-376 (AC12), spec:75-79.

The spec says freshness "comes from the verified session alone". It does not say where the check runs, and none of
the existing seats can hold it:

- The authorizer receives only `(ctx, op)` (`internal/ops/opshttp/authorize.go:19-21`). It never sees the request body,
  so it cannot know whether a `set_setting` targets a sensitive key.
- `admit` wraps every authorizer error under `NewForbiddenError` first (`authorize.go:100-103`). A stale session
  refused there answers `forbidden`, which breaks AC12's "distinct code".
- `ops.Actor` carries only a wire string and a channel (`internal/ops/ops.go:24-27`). Nothing carries the session
  anchor past the authorizer into the seam implementation, which is where the key and its mark are known.
- A new wire code must join `ops.Codes()` and get an HTTP status, or `newHandler` fails (`opshttp.go:24-44`).

**Fold:** the spec names the seat. The authorizer builds the actor from the verified session and records whether the
session is fresh. That could be an unexported field on `ops.Actor`, set by a `NetworkActor` variant that takes the
anchor and the window, while `LocalActor` is always fresh. The seam implementation checks freshness after it decodes
the key: for `set_setting` and `reset_setting` when the key is sensitive, and always for `remove_member` and
`demote_member`. It refuses with a new code such as `reauthentication_required`, which gets a status in
`opshttp.statuses` and a place in the document as an additive enumeration value. This touches `internal/ops`'s public
surface, so the plan's privileged-fields table (spec:312-320) gains a "freshness" row: set by the authorizer alone,
never by a request field.

### B4 (major): an older store is refused at every write, not only at open

Location: spec:163-167, spec:379-381 (AC14).

The spec leaves "the open mechanism" to the plan. The write path has a second gate. `Handle.Update` reads
`user_version` as its first statement and refuses with `errVersionMismatch` whenever it differs from `h.version`
(`internal/engine/store/write.go:62-63`, `:78-84`). `OpenWrite` sets `h.version = len(files)`, the binary's head
(`open.go:179`). So a handle on an older store refuses every `Update`, whatever opened it. The verb needs a new store
entry point that pins `h.version` to the store's own version, plus SQL that is valid at every version from `0004` up to
head. That is new surface in the store package: a write path into an unmigrated store, which the start-order contract
(R-6d2) exists to protect.

**Fold:** if B5 keeps the allowance, the spec says so plainly. It names a new store open mode that accepts versions 4
through head and pins the handle's version. It names a test that runs the verb against a fixture at every version from
`0004` to head, so a later migration that alters `settings` or `settings_audit` fails the gate. It also says how the
mode takes the lock. Shared is enough, because `settle` needs the lock exclusive (`open.go:263`), so a starting
`serve` waits instead of migrating under the verb.

### B6 (minor): AC17's refusal contradicts `opshttp`'s decode contract and AC5

Location: spec:387 (AC17), spec:358-359 (AC5), spec:84-85.

`decode` ignores unknown keys by design (`opshttp.go:141-144`: "Unknown keys are ignored, so a body naming an actor or
a privilege changes nothing"). AC5 relies on that. AC17 says a body carrying an admin field "is refused". The
retiring roster API did refuse it (`internal/auth/admin.go:291-293`, `admin_not_grantable`), but the new mount will
not unless a per-operation strict decode is built.

**Fold:** AC17 reads "a body carrying an admin field adds a member without the admin flag", which matches AC5 and the
package contract. It needs no new decode path.

### B7 (minor): AC2 cannot hold for the negative tests the other criteria require

Location: spec:351-352 (AC2).

"Every request the tests send validates against the document" conflicts with tests that must send off-contract
requests: a wrong method or content type, a query string, malformed or oversized JSON (`opshttp.go:88-101`, `:145-171`),
`api/v2/...` (AC4), and an actor field (AC5) if request schemas are closed. Some of these have no route in the
document at all. In that case `FindRoute` fails, so the answer cannot be validated.

**Fold:** AC2 says that every request sent by a positive-path test validates, and every answer to a routed request
validates, refusals included. An unrouted path is asserted by its status and JSON body alone.

### B8 (minor): the mount's wrapper, its ordering, and its no-session answer are unstated

Location: spec:68-79, spec:355-359 (AC4, AC5).

The handler sees the full path, prefix included (`serve.go:96-98`). `opshttp` routes only `/<name>` (`opshttp.go:82`).
`http.StripPrefix` answers a non-matching path with a plain-text 404 (`go doc net/http.StripPrefix`), so `api` and
`api/v2/x` would answer plain text, not AC4's JSON. The only existing session middleware, `RequireAdmin`, answers
non-admins with plain `http.NotFound` (`admin.go:125-128`). It would refuse a member's `catalog` call (AC5), and its
answer is not JSON.

**Fold:** the spec names three things:
- A mount wrapper that strips `<prefix>/api/v1` and answers `not_found` JSON for every other path under `api`.
- The order: routing to `not_found` comes before session verification. That order is safe because the document is
  public.
- The answer for a request with no valid session, in JSON and declared in the document, from a middleware that admits
  any verified member and puts the full verified session in the context for the authorizer and B3's freshness.

### B9 (minor): `ErrEpochBackward` must not roll a removal back

Location: spec:203-205.

`Bump` returns `ErrEpochBackward` when the stored epoch is later than now, as after a clock step back. It then
leaves the later epoch in place (`epoch.go:84-86`, `:152-153`). The revocation already holds, so "a bump failure rolls
the removal back" would refuse a removal for no reason.

**Fold:** treat `errors.Is(err, auth.ErrEpochBackward)` as success inside the callback, and add that state to the
removal's state list (spec:308).

### B10 (minor): subscribers would run while the store's single writer is held, and AC15 needs a one-second step

Location: spec:206-209, spec:383-384 (AC15).

`Bump` writes the epoch and then notifies subscribers synchronously, up to 5s each (`epoch.go:103-117`, `:122`). Inside
the write callback, a future subscriber therefore holds every writer on the box, in every process, behind it. None
subscribes today, which leaves a trap for slice 6. Separately, `CheckSession` and `CheckLink` refuse at an equal
second (`!AnchoredAt.After(epoch)`, `epoch.go:237`, `:265`). A sign-in in the same second as the removal's bump is
refused, so AC15's "a sign-in after the re-add verifies" needs the fixture clock at least 1s past the bump.

**Fold:** record a one-line contract on `Subscribe` that a subscriber must not block, because a bump may run under the
store's write lock. Alternatively, notify after the commit. Also name the 1s step in AC15's fixture.

### B11 (minor): the BACKLOG #23 rider targets an answer shape that retires in this pass

Location: spec:247-249, spec:391-392 (AC19).

The only JSON 403 any `postForm` caller receives is the roster API's `refused` (`admin.go:318-319`), and 14a retires
that API. `/api/v1` answers `{"error":{"code","reason"}}` with no `outcome` field (`opshttp.go:230-243`) and refuses any
body that is not `application/json` (`:97-101`), so the members page cannot keep `postForm`.

**Fold:** the rider becomes a JSON client for `/api/v1` that tells `opshttp`'s JSON 403 (by its `Content-Type`) from
the origin check's plain-text 403. BACKLOG #23 then closes with the roster API's retirement, and `post.ts` stays as it
is for its three remaining callers.

### B12 (minor): a reset of a live-apply key puts the default in force until the next start

Location: spec:129-133.

`reset_setting` deletes the `admin` row. The accessor returns `s.Default` for a key with no row (`accessor.go:113-116`),
and the environment seed runs only at start. When the environment sets the key, a live-apply key therefore runs on the
default between the reset and the next restart, not on the environment's value.

**Fold:** the reset's answer names the value now in force and the value the next start will seed. The page shows both.
This needs no new mechanism.

### B13 (minor): the OpenSubsonic precedent is misstated, and "`_test.go` only" is stricter than the no-link test

Location: spec:45-47.

The OpenSubsonic document is the upstream OpenAPI 3.0.0 file, fetched and pinned by sha256
(`internal/harness/subsonic/data/openapi/PROVENANCE.md`), so it is not hand-written. The repo also uses kin-openapi in
a non-test package, `internal/opensubsonic/conformance`. The no-link test forbids `github.com/getkin` only in
`cmd/dubplate-server`'s dependency closure (`arch_test.go:213-217`). It does not forbid it in non-test files.

**Fold:** correct the sentence. Allow a non-test contract-checking package that the server never imports, so the
silent-pass classification from B2 can be written once instead of in each test file.

### B14 (minor): the sensitive mark has no stated meaning for keys that cannot change live

Location: spec:101-110.

The spec's meaning for the mark is "dangerous to change live". Wall rows, thresholds, and integration keys are
environment-only and apply at restart, so that meaning is empty for them. The other named consumer, `dubplate init`
(SS:766-767), needs a different fact: whether a value is a secret it must mint or hide. R-6d6 refused that meaning for
settings.

**Fold:** state the mark's value for environment-only families, for example "not applicable, apply at restart",
and leave secret-ness to `init`'s own declaration when it lands. Otherwise the plan has to guess and the page shows a
meaningless mark.

### B15 (minor): doctor's admin-address check needs a new fact from `Inspect`

Location: spec:242-243, spec:396-397 (AC21).

`store.Facts` carries only counts of members and admins (`inspect.go:166-171`, `:357-362`). It does not say whether a
given address is a live admin. This is a plan-level note. Either `InspectOptions` takes the address, or `Facts` lists
the admin addresses.

## Over-ceremony

### B5 (OWNER FORK, major cost): the older-store allowance for `settings reset`

Location: spec:163-168, spec:307, spec:323-324, spec:379-381.

**Cost:** B4's new store mode, a cross-version fixture for every version from `0004` to head, three AC14 states, and a
second write path into an unmigrated store in the auth-data class, which brings extra security review. The case arises
only after launch, when an upgrade's bounds refuse a row an admin set. An environment contradiction does not need it,
because the operator can fix that through the environment file.

**Options:**
- (a) Keep the allowance as specified, with B4's fold.
- (b) Drop it. The seed rule's bounds refusal on an older store names the recovery: run `settings reset <key>` with
  the binary that matches the store, then start the new binary. That works because every store after launch was
  created by a binary that already has the verb (14a lands before 0.8.0). The verb keeps `OpenWrite`'s refusals
  unchanged.

**Recommendation:** (b). It removes the riskiest new surface in the pass at a rare cost to the operator. Revisit at
1.0 with `dubplate upgrade`, which owns the swap between binaries.

## Summary

Lens B returns 15 findings: 4 major (B1 to B4), 10 minor (B6 to B15), and one owner fork about over-ceremony (B5,
major cost). The top three:
- **B1:** kin-openapi v0.149.0 silently falls back to its 3.0 validator whenever a schema nests a `$ref`, so 3.1-only
  keywords such as `unevaluatedProperties`, `dependentRequired`, and `if`/`then` validate nothing in a real document.
  The probe confirmed it.
- **B2:** `ValidateResponse` passes undeclared statuses and schemaless answers by default, so AC1 and AC2 can pass
  vacuously without `IncludeResponseStatus` and the `conformance` package's silent-pass classification.
- **B3:** the re-authentication gate has no seat. The authorizer sees only the operation name and folds every error
  into `forbidden`, and `ops.Actor` carries no session anchor. The spec must name where freshness travels and the
  distinct wire code it answers.

B4, which shows that `Handle.Update` refuses an older store at every write, makes B5's recommendation to drop the
older-store allowance the cheaper path.
