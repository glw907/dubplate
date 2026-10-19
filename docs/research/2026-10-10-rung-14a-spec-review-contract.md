# Rung 14a spec review, lens A: contract and criteria

Target: `docs/superpowers/specs/2026-10-10-rung-14a-admin-api-design.md` at `9d730d7`. Reviewer: fresh context,
lens A (is every promise testable; does each criterion name its fixture and its failure; can any pass vacuously; does
every check state name its report; does every privileged field name every channel that sets it). Code read at
`9d730d7`. Spec line numbers below are `spec:N`. The rulings the review prompt lists are not reopened.

Ranked by consequence. Correctness gaps come first, then over-ceremony.

## Correctness gaps

### A1 (major). The freshness input has no named channel into the code that decides it

- **Location:** spec:137-142, spec:75-76, spec:310-320 (the privileged-field table).
- **Defect:** The gate needs two facts together: the session anchor and whether the target key is sensitive. Neither
  existing carrier holds both. The opshttp authorizer "receives no request, so a verified identity reaches it only
  through ctx" (`internal/ops/opshttp/authorize.go:12-21`). That means it sees `op` but never the key in a
  `set_setting` body, so it cannot tell a sensitive write from a plain one. The seam implementation does see the key.
  But it receives only an `ops.Actor`, which holds a wire string and a channel and nothing else
  (`internal/ops/ops.go:24-27`). `identity.Identity` holds no anchor either (`internal/identity/identity.go:203-224`).
  The anchor lives only on `auth.Session`, inside the `VerifiedIdentity` the gate puts on the context
  (`internal/auth/seam.go:43-61`). Line 142 says "freshness comes from the verified session alone, never from a
  request field", but it does not say how that fact reaches the decision point. The spec has the same blind spot
  that opshttp's design closes for the actor. The privileged-field table has no row for "session freshness".
- **Failure scenario:** An implementer reads the anchor straight off the context inside the seam implementation. That
  brings back the pattern opshttp's doc rules out ("calls through only with the actor the authorizer answered",
  `opshttp/doc.go:6-15`). Another implementer adds a freshness flag to a request type, or gates every `set_setting`
  in the authorizer whatever the key's mark (see A9). Each choice passes criterion 12 as written.
- **Fold:** Name the carrier. The recommended answer is that the authorizer stamps the anchor (or a computed
  `FreshAt`) onto the `ops.Actor` it answers. It is an unexported field set only by a constructor that takes a
  verified session, the way `NetworkActor` takes a valid identity, and the command-line actor carries none. The seam
  implementation decides the gate from the actor plus the key's registered mark. Add a privileged-field row: "session
  freshness: the authorizer, from the verified session's `AnchoredAt`; never a request field; absent on the command
  line, which the gate skips."

### A2 (major). Criterion 2 cannot be met as written: the negative-path requests cannot validate

- **Location:** spec:351-353 (criterion 2), spec:44-45.
- **Defect:** Criterion 2 says every request the `/api/v1` tests send validates against the document, with errors
  included. The same criteria require tests that send requests the document cannot describe:
  - criterion 4's `api/v2/...` and unknown paths (kin-openapi's router finds no route, so neither `ValidateRequest`
    nor `ValidateResponse` has an operation to run against);
  - criterion 5's body carrying an actor field and criterion 17's body carrying an admin field (if the request schemas
    set `additionalProperties: false`, these fail request validation; if they do not, A3 applies);
  - criterion 10's out-of-bounds and wall-key bodies;
  - opshttp's transport refusals: a non-POST method, a query string, a wrong content type, an empty or oversized body
    (`opshttp.go:96-114`, `:145-175`).
  An implementer has to carve out exceptions the spec never names, and an unnamed carve-out is where the vacuity
  hides.
- **Fold:** Split criterion 2 into three rules:
  1. A request the test means to be well formed must pass `ValidateRequest`.
  2. A request the test means to be malformed at the schema level must fail `ValidateRequest`. This proves the
     document encodes the refusal.
  3. Transport-level refusals, and paths with no operation, validate their response body against the shared error
     schema directly (`openapi3.Schema.VisitJSON`), since no route exists.
  Every answer then validates, either through its route or against the error schema.

### A3 (major). The validator options and schema strictness are unnamed, so criteria 1 and 2 can pass vacuously

- **Location:** spec:41-47, spec:348-353 (criteria 1 and 2).
- **Defect:** Three kin-openapi behaviors decide whether the contract tests can fail at all. The spec quotes none of
  them.
  1. **Undeclared status codes pass by default.** `openapi3filter.Options.IncludeResponseStatus`'s documented default
     is off: "Set IncludeResponseStatus so ValidateResponse fails on response status not defined in OpenAPI spec"
     (`go doc .../openapi3filter.Options`). With the default, the promise in criterion 1 that the document lists
     "every wire code it can answer" is never checked. A 409 the document omits still validates.
  2. **3.1 semantics need an explicit option.** The package doc reads "For OpenAPI 3.1 validation, use the JSON Schema
     2020-12 validator option: `openapi3.EnableJSONSchema2020()`". Without it, a 3.1 document's type arrays such as
     `["string","null"]` are not checked under 3.1 rules.
  3. **The rename proof can pass vacuously.** Criterion 2's "renamed response field fails validation" fails only when
     the renamed field is `required`, or when the response schema sets `additionalProperties: false`. Renaming an
     optional field validates cleanly. A Go response field the document never describes is caught only by
     `additionalProperties: false`.
  The precedent the spec cites does not cover this. Line 47 says "The OpenSubsonic layer took the same hand-written
  route", but that layer validates against a vendored OpenAPI **3.0** document. It sets `IncludeResponseStatus: true`
  (`internal/opensubsonic/conformance/conformance.go:133`) and never enables the 2020-12 validator.
- **Fold:** Have criterion 2 name the options: `IncludeResponseStatus: true` and `EnableJSONSchema2020()` through
  `SchemaValidationOptions`. Rule that response schemas list every always-present field under `required` and set
  `additionalProperties: false`, and that the rename proof targets a `required` field. Add a second proof: a Go
  response field missing from the document fails validation. Correct line 47's precedent claim. The additive policy
  survives, because the server and the document change together, and the client-side ignore rule lives in the policy
  text.

### A4 (major). The wire gives the pages no way to tell their refusals apart

- **Location:** spec:144, spec:247-249 (rider #23), spec:374-375 (criterion 12), spec:390-392 (criterion 19).
- **Defect:** The opshttp error body is `{"error":{"code","reason"}}` (`opshttp.go:239-243`), drawn from seven closed
  codes (`internal/ops/errors.go:15-21`). Three things in the spec depend on a discriminator that does not exist on
  that wire:
  1. **Criterion 19's three refusal copies have no code to key on.** The criterion asks for distinct copy for "last
     admin, re-authentication needed, forbidden". The spec names no code for re-authentication (line 144 says only
     "a distinct refusal") and no code for the last-admin guard (`errNoAdmin` has no stated mapping). If
     re-authentication rides `forbidden` and last-admin rides `conflict`, the page has to parse the free-text
     `reason` to choose copy.
  2. **The members page's existing per-field refusals lose their discriminator.** Today the page maps
     `duplicate_address`, `duplicate_username`, `invalid_username`, `missing_username`, `invalid_display_name`,
     `blank_display_name`, `admin_not_grantable`, and `busy` to field-specific copy and focus moves
     (`web/src/routes/members/+page.svelte:46-55`, `:95-99`, W5). Once add moves to `/api/v1`, these collapse into
     `invalid_request` or `conflict` with free text. That is an accessibility regression, and no criterion would
     catch it.
  3. **Rider #23 is written against the old body shape.** It says "the client reads a JSON outcome on a 403", but
     the field is `error.code`, not `outcome`. `postForm` also posts urlencoded bodies, which opshttp refuses
     (`opshttp.go:110-113`).
- **Fold:**
  - Name the codes. Re-authentication gets a new wire code, `reauth_required`, at HTTP 403. The version policy at
    spec:49-55 already makes this additive. Last admin gets `conflict` plus the discriminator below.
  - Add an optional, enumerated `error.detail` field to the error body. It is additive under the same policy and
    carries the per-field outcomes and `last_admin`.
  - Have criterion 19 assert that each add refusal still moves focus to its field.
  - Rewrite the rider: the client reads `error.code` on a JSON 403 and falls back to `cross_origin` only for the
    origin check's plain-text body.

### A5 (major). Criteria 5 and 17 disagree on unknown request keys

- **Location:** spec:359 (criterion 5: "A body carrying an actor field is ignored"), spec:387 (criterion 17: "a body
  carrying an admin field is refused and changes nothing").
- **Defect:** opshttp's decoder ignores unknown keys by design: "Unknown keys are ignored, so a body naming an actor
  or a privilege changes nothing" (`opshttp.go:143-144`). `add_member` "carries no admin field" (spec:84), so an
  `admin` key is just an unknown key and would be ignored. Criterion 17 says it is refused. The existing roster API
  refuses it explicitly with `admin_not_grantable` (`internal/auth/admin.go:42`, `:292`). The spec does not say which
  rule `/api/v1` follows. If the key is silently ignored, a 0.9 TUI that sends `"admin": true` gets a non-admin
  member and no signal.
- **Fold:** Take the existing behavior. `add_member`'s handler refuses a body that names `admin`, with
  `invalid_request` and `detail: admin_not_grantable`, and the document encodes the refusal (the `admin` property
  declared `not: {}` or similar). Other unknown keys are still ignored, per opshttp. State both rules in the spec, so
  that criterion 5 and criterion 17 each test one of them.

### A6 (major). The planning-miss section lists the states each check meets but names no report for them

- **Location:** spec:306-309, plus the states that list never reaches.
- **Defect:** CLAUDE.md's planning-miss rule requires that "each state a check can meet (absent, empty, malformed,
  many-bad) names its report". Line 306 lists states and stops there. The following have no named answer:
  - **Removal:** already removed, unknown id, a malformed hex id (wrong length or uppercase; `ParseHandle` at
    `ops.go:~73` is the precedent), and self-removal by a non-last admin. For self-removal, is it allowed, and is the
    caller's own session revoked by the bump?
  - **Demotion** (missing from the list entirely): a target that is already a member, a tombstoned target, an unknown
    id, a malformed id, and self-demotion.
  - **The resolving verb:** no row; an `env-seed` row; an older store without the settings table.
  - **`reset_setting` on the cookie mount:** no row, or an `env-seed` row.
  - **`set_setting`:** an unparseable value, as distinct from an out-of-bounds one.
  - **`doctor`'s admin-address check:** an address naming a live non-admin, a tombstoned member, or no member; and an
    empty roster before bootstrap, which already has its own report per R6's 6d erratum.
- **Fold:** Add one table to the spec mapping each state to its wire code (plus `detail`, per A4) or CLI exit and
  message. Recommended answers:
  - removing an already-removed or unknown member: `not_found`;
  - a malformed id: `invalid_request`;
  - demoting a member who is not an admin: a no-op success, or `conflict`; pick one;
  - resetting when no row or only an `env-seed` row exists: a no-op success naming the source in force;
  - the verb on a store without the settings table: refused, telling the operator to fix the environment;
  - `doctor`: warn on every non-admin case, and stay silent before bootstrap, where the empty-roster report already
    covers it.

### A7 (major, OWNER FORK). `set_setting` knowingly commits a value that makes the next start refuse

- **Location:** spec:131-133, spec:372-373 (criterion 11).
- **Defect:** Under the ratified seed rule, an `admin` row whose value differs from the environment makes `serve`
  fail to start (R6 seed table; `internal/settings/evaluate.go:293-294`). The drafted write path commits such a value
  and only warns. A kernel-update reboot or a deploy before the operator acts leaves the box down. The fix then needs
  box access: `settings reset` or an environment edit. The web admin who made the write may not be the operator. The
  serving process's environment is also not the next start's: the environment file may already have been edited
  since `serve` started. So the check is approximate in both directions.
- **Options:**
  - (a) Refuse the write with `conflict` (and `detail: env_contradicts`) when the serving environment sets the key to
    another value. A false refusal is possible when the environment file has already changed; it clears at the next
    restart.
  - (b) Commit and warn, as drafted.
  - (c) Commit and warn, and also re-check at shutdown or in `doctor`.
- **Recommendation:** (a). A refused write is visible and reversible, while (b)'s failure mode is an outage found
  only at the next restart. Criterion 11 then becomes "the write is refused naming the environment value and both
  resolutions".

### A8 (major). The concurrency proof cannot fail on the defect it exists to catch

- **Location:** spec:385-386 (criterion 16), spec:322.
- **Defect:** "Two concurrent self-demotions of the last two admins leave exactly one admin" also passes when the two
  calls run one after the other. It passes even with a guard placed outside the transaction: the second caller's
  check then sees one admin, itself, and refuses. So the proof cannot tell an in-transaction guard from a
  pre-transaction one, which is the property 6d criterion 6's clause names. The fixture line ("two live admins and
  two demotions") names no interleaving.
- **Fold:** The proof forces the interleaving. Both callers pass every pre-write step before either enters the write
  callback, using a barrier in a test seam before `Update`. Run the test once against a deliberately misplaced guard
  to show it can fail, per CLAUDE.md's "prove tests can fail".

### A9 (major). The re-authentication gate's scope is under-specified and its criterion cannot detect over-gating

- **Location:** spec:137, spec:315, spec:374-376 (criterion 12).
- **Defect:**
  1. **`reset_setting` of a sensitive key is never stated as gated.** The section text says "a write to a sensitive
     key". The privileged-field row for "removal of an `admin` row" omits "inside the window", though the demote and
     remove rows include it. Criterion 12 lists "a sensitive write, a removal, and a demotion" but not a reset.
     Resetting the session lifetime or the re-authentication window itself is as dangerous as setting it.
  2. **Nothing proves the gate keys on the mark.** Criterion 12 has no counter-case showing that a non-sensitive
     `set_setting` from a stale session commits. A gate on every write, or a gate in the authorizer by operation
     name, passes criterion 12.
- **Fold:** Gate `reset_setting` on a sensitive key, and add "inside the window" to its privileged-field row. Add two
  clauses to criterion 12:
  - a stale session's non-sensitive write and non-sensitive reset commit;
  - a stale session's sensitive reset is refused with the re-authentication code.

### A10 (minor, OWNER FORK). D10's reason also covers add and re-add, which stay ungated

- **Location:** spec:259-263 (D10), spec:318 ("cleared only by a re-add").
- **Defect:** D10 gates removal and demotion because "each changes who holds access". `add_member` also changes who
  holds access, and so does a re-add that clears a tombstone. A shared-laptop holder cannot remove a member but can
  add their own address, or restore a member an admin just removed, all without a fresh sign-in. The privileged-field
  row "a user's tombstone ... cleared only by a re-add" names no channels: `member add` on the CLI and `add_member` on
  the cookie mount, the second ungated as drafted.
- **Options:**
  - (a) Gate `add_member` too.
  - (b) Narrow D10's stated reason to "removes access, or cannot be undone from the web", and record add as
    deliberately ungated.
- **Recommendation:** (a). It costs one more gated operation and keeps D10's reasoning whole. Either way, name both
  channels in the tombstone row.

### A11 (minor). `ErrEpochBackward` would roll back a removal whose revocation already holds

- **Location:** spec:203-206.
- **Defect:** `Bump` returns `ErrEpochBackward` when `now` is earlier than the stored epoch
  (`internal/auth/epoch.go:84-87`). The spec says "a bump failure rolls the removal back", so a backward clock step
  refuses the removal. In that case the stored epoch is already later than `now`, so every session the removal must
  revoke is already revoked.
- **Fold:** Treat `ErrEpochBackward` as satisfied: log it and commit. Name this state in A6's table.

### A12 (minor). Epoch comparison at second granularity makes criterion 15's re-add sign-in time-dependent

- **Location:** spec:382-384 (criterion 15), spec:322.
- **Defect:** The epoch is truncated to the second (`epoch.go:105`), and `CheckSession` and `CheckLink` refuse any
  anchor or mint time "at or before" it (`epoch.go:237`). On a real clock, "a sign-in after the re-add verifies"
  fails whenever the bump, the link mint, and the consume fall in the same second. On a frozen clock it always
  fails.
- **Fold:** The fixture advances the injected clock at least one second past the bump before minting the link. State
  the same-second refusal as the intended behavior.

### A13 (minor). Several criteria name no fixture that makes them able to fail

- **Location:** criteria 6, 9, 13, and 18 (spec:360-361, 367-368, 377-378, 389).
- **Defects and folds:**
  - **Criterion 9** ("no wall or integration value appears in any answer") passes trivially when the test's wall
    values are empty. Fold: set every wall and integration row to a distinctive sentinel string, and search the raw
    body of every `/api/v1` answer for each sentinel, refusals included.
  - **Criterion 13** names no negative case. Fold: a non-sensitive committed write mails no one; the fixture holds
    two or more live admins and one tombstoned admin, and the tombstoned admin receives nothing.
  - **Criterion 18**'s preference-deletion assertion passes when the member had no `member_preferences` row. Fold:
    the fixture writes one first.
  - **Criterion 6** needs a rename, and no rename path exists at 14a. Fold: say the fixture renames through a direct
    store write.

### A14 (minor). The value in force between a cookie-mount reset and the next start is unstated

- **Location:** spec:129-130.
- **Defect:** Deleting an `admin` row leaves no row. Until the next start the accessor then resolves the registry
  default, even when the serving environment sets the key (R6 seed table: "no row; reads resolve the default"). For
  an immediately-applied key such as `DUBPLATE_SESSION_ABSOLUTE_CAP`, the live value jumps to the default and then
  to the environment value at restart. Neither the answer nor the page says so.
- **Fold:** On the cookie mount, when the serving environment sets the key, have reset re-seed it in the same
  transaction: an `env-seed` row, audited with the `env-seed` actor per R6. Otherwise the answer names the default
  as in force. The CLI verb with `serve` stopped keeps the delete-only behavior.

### A15 (minor). The resolving verb's criterion omits the third refusal on an `admin` row

- **Location:** spec:160-161, spec:167-168, spec:379-381 (criterion 14).
- **Defect:** `outcome` refuses an `admin` row in three places, not two: when the stored value is unparseable for
  the key's kind (`evaluate.go:264-266`, for example a new binary that changes a key's kind), when the environment
  contradicts it (`:293-294`), and when it is out of bounds (`:304-307`). Criterion 14 and the "two refusal
  messages" line cover only the last two.
- **Fold:** Add the unparseable `admin` row to criterion 14's fixtures. Have that message name the verb too. It is
  the line rider #25 already edits.

### A16 (minor). The sensitive mark has no testable meaning on wall and integration rows

- **Location:** spec:101-110, spec:364-366 (criterion 8).
- **Defect:** Per R-6d6, `sensitive` means "dangerous to change live". No channel changes a wall or integration row
  live, so for those rows the mark has no defined meaning, and nothing can check the value an author picks. SS:766
  plans for `dubplate init` to read the mark when it asks for SMTP and secret values. That use needs "secret, do not
  echo", which R-6d6 says the mark is not. Requiring an arbitrary mark on every wall row now invites a later reader
  to misread it.
- **Fold:** Give wall and integration rows a description only. Keep the mark on store-homed keys, the only ones a
  live write reaches. Note that `init`'s secret-ness question belongs to `init`'s own rung. The alternative is to
  define the mark's meaning for wall rows in the spec before criterion 8 can be graded.

### A17 (minor). Removal says nothing about the member's `member_overrides` and `member_quota` rows

- **Location:** spec:185-186.
- **Defect:** Migration 0005 created `member_overrides` and `member_quota`, both keyed on the user id. A re-add under
  the old id (spec:194-195) revives whatever removal left behind. The spec deletes preferences and says nothing about
  these two tables.
- **Fold:** State that both survive removal. Overrides are admin policy and quota reflects audio that stays.
  Criterion 18's fixture then asserts they survive.

### A18 (minor). The re-authentication window's bounds and its relation to the absolute cap are unstated

- **Location:** spec:140-141.
- **Defect:** A window at or above `DUBPLATE_SESSION_ABSOLUTE_CAP` makes the gate a no-op. A window shorter than a
  realistic link round trip locks the web out of every sensitive write, including resetting the window itself, until
  someone runs the CLI verb.
- **Fold:** Name a correlated rule (window shorter than the absolute cap) and a floor above a link's consume
  latency. The plan researches the default, as drafted.

## Over-ceremony

### A19 (minor). Criterion 8's whole-registry test duplicates the registration panic

- **Location:** spec:364-366.
- **Cost:** One test, plus review tokens to grade it. A registration that panics at construction already fails every
  test binary and every start that registers the key. A sweep finding "none missing" can only pass once the binary
  constructs at all.
- **Fold:** Keep the panic, plus one test proving the panic fires for a missing description and for a zero mark.
  Drop the whole-registry sweep, or reduce it to one line in an existing construction test.

### A20 (minor). Criterion 18 repeats criterion 15

- **Location:** spec:382-384 and spec:389.
- **Cost:** A second test over the same fixture and assertions (tombstone, bump, preferences), and a second grading
  pass.
- **Fold:** Criterion 15's test is the checklist test. Fold criterion 18 into it, and keep the table at spec:222-229
  as the owner map.

## Summary

Twenty findings: no blockers, nine major (two of them owner forks, A7 major and A10 minor), and eleven minor (two of
them over-ceremony). The top three are:

1. **A1:** the spec never names how session freshness reaches the code that decides the gate. The authorizer cannot
   see the key, and the actor the seam receives holds no anchor. The fold is an anchor stamped on the actor by the
   authorizer, with a privileged-field row for it.
2. **A3 with A2:** criteria 1 and 2 can pass vacuously or not at all. kin-openapi's `IncludeResponseStatus` defaults
   off, 3.1 validation needs `EnableJSONSchema2020`, the rename proof needs a `required` or closed schema, and the
   negative-path requests cannot validate against the document as criterion 2 demands.
3. **A4:** the opshttp error body has no discriminator for the refusals the pages must tell apart. These are
   re-authentication, last admin, and the members page's eight per-field add outcomes with their W5 focus moves.
   Rider #23 is written against the old `outcome` shape.
