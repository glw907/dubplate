# Rung 14a: the admin API's first half

Status: revision 2, folded from the four-lens review, 2026-10-10. The fold record is
`research/2026-10-10-rung-14a-spec-fold.md`. Design approved in session the same day; this document records it.

This spec covers the first of rung 14's two passes. It records the contract-document form, the authorization
model at 0.8, and the decisions the planning brief (`research/2026-10-10-next-pass-planning-brief.md`) left open.
Most of the design is already settled in CH and R6, so this spec cites them rather than restating them.

Abbreviations follow the brief: BS is the build spec (`specs/2026-09-02-greenfield-server-build.md`), RP the re-plan
(`specs/2026-10-05-rungs-12-16-replan-design.md`), SS the streamer spec (`specs/2026-09-07-streamer-2-0-design.md`),
CH the configuration homes spec (`specs/2026-09-11-configuration-homes-design.md`), R6 the rung 6 spec
(`specs/2026-09-23-rung-6-ops-seam-store.md`), and PA the parallel build approach
(`specs/2026-09-03-parallel-build-approach.md`). Paths sit under `docs/superpowers/` unless they start at the repo
root. Code citations were read at `e6fdf2b`; no code changed through `045aadc`.

## Owner rulings (Geoff, 2026-10-10)

1. **Rung 9's concurrency model is a per-album lease.** Recorded as ruling 10 in
   `research/2026-09-02-greenfield-server-rulings.md`. This discharges the hard (14,9) edge
   (`research/2026-09-04-reverse-edge-completion.md:37`) before this plan, as PA:118 requires.
2. **Rung 14 at 0.8 builds only what has a 0.8 caller.** The bearer mount, non-browser token acquisition, service
   tokens, connect version negotiation, and the adopt and seize verbs move to 0.9 with the local client.
3. **Rung 14a carries its pages.** A remove and demote control on the existing members page, and a new admin
   settings page. The collection page goes to 14b.

## The split

Rung 14 is banded M (BS:983). BS:682-684 says "M is one pass under one plan" and "L splits into multiple passes at
its own planning", so this split re-bands rung 14 as L. The brief's D3 reads BS:684-685 as letting an M rung split;
the text gives that latitude to L only.

**14a, now:** the contract document; `/api/v1` on the cookie mount over `opshttp`; the network actor on `UserID`;
hook 18; the admin settings write path with its re-authentication gate, sensitive-change notice, and CLI resolving
verb; member removal and demotion over the CLI and `/api/v1`; the members and admin settings pages.

**14b, after the first rung that writes a collection row** (rung 7 or rung 10, R6:185-186): the collection
endpoints with `if_version`, the private-visibility refusal, the per-member override endpoints, the collection page,
and the adoption of any JSON subtree rung 7 mounted outside `/api/v1` (see "Contested surfaces").

## The contract document (D4)

The contract is one hand-written OpenAPI 3.1 file. `go doc github.com/getkin/kin-openapi/openapi3` at v0.149.0 reads
"Supports OpenAPI 3.0, OpenAPI 3.1, and OpenAPI 3.2". The plan names the file's path. Nothing serves it at 14a, and
it declares no `security` requirement, so `ValidateRequest` needs no `AuthenticationFunc`
(`openapi3filter/validate_request.go:19`, `:438`).

Tests validate `/api/v1` requests and answers through `openapi3filter.ValidateRequest` and `ValidateResponse`. Four
kin-openapi behaviors decide whether those tests can fail, and the contract tests close each:

- `ValidateResponse` passes a status the operation does not declare unless `Options.IncludeResponseStatus` is set
  ("By default, status that is not documented is allowed", `validate_response.go:51-56`). It also passes HEAD,
  redirects, an answer with no content, and content with no schema (`:22-35`, `:99-117`). The tests set
  `IncludeResponseStatus: true` and treat each other silent pass as a failure, on the `conformance` package's
  precedent (`internal/opensubsonic/conformance/conformance.go:96-136`).
- `openapi3filter` turns on the JSON Schema 2020-12 validator itself for a 3.1 document
  (`validate_request.go:247-248`, `validate_response.go:76-77`). A direct `Schema.VisitJSON` call passes
  `openapi3.EnableJSONSchema2020()` explicitly.
- The 2020-12 path compiles each schema as a standalone resource, so a nested `$ref` fails the compile, and the
  validator then falls back silently to the built-in 3.0 validator (`openapi3/schema_jsonschema_validator.go:20-56`,
  `:290-297`, "Fall back to built-in validator if compilation fails"). A 3.1-only keyword such as
  `unevaluatedProperties`, `dependentRequired`, or `if`/`then` then validates nothing. The document uses only
  keywords the built-in validator enforces: `type` and type arrays, `const`, `enum`, `required`, `properties`,
  `additionalProperties`, `items`, `prefixItems`, `oneOf`, `anyOf`, `allOf`, and `format`. A contract test walks the
  document and fails on any other schema keyword.
- A renamed field fails validation only when it is `required` or its schema is closed. Every response schema sets
  `additionalProperties: false` and lists every always-present field under `required`.

kin-openapi appears in `_test.go` files only, in the mount's own package, so the silent-pass classification is
written once. The no-link test (`internal/harness/conform/arch_test.go:213-217`) forbids any `github.com/getkin` path
in `cmd/dubplate-server`'s dependency set. The OpenSubsonic precedent differs in kind: its document is the upstream
OpenAPI 3.0.0 file, pinned by sha256 (`internal/harness/subsonic/data/openapi/PROVENANCE.md`), not a hand-written one.

The document states the version policy, and the policy is additive within v1. These changes are additive: a new
operation, an optional request field, a new response field, and a new value in any enumeration. The enumerations
are job states, job kinds, wire codes, error details, and the journey vocabulary's wire projection. A client must
ignore an unknown field and treat an unknown enumeration value as opaque. Removing or renaming any of these, or making
an optional field required, is a v2 change. This rules the journey question the reverse-edge record left to rung 14
(`research/2026-09-04-reverse-edge-completion.md:223`): rung 7 owns the vocabulary, and adding a state to it is
additive. It also answers the (14,13) edge (`:39`): rung 13's queue verbs that have a page caller join `/api/v1` as
catalog operations, written by rung 13 in its own files, under this policy.

### The error body and its two new codes

The opshttp error body is `{"error":{"code","reason"}}` (`internal/ops/opshttp/opshttp.go:230-243`). The members page
today keys field-specific copy and focus moves on its refusal outcomes (`web/src/routes/members/+page.svelte:46-54`),
and `reason` is free text, so the body gains one optional field: `error.detail`, an enumerated machine name finer than
the code. Each existing add outcome keeps its name as a `detail` (`duplicate_address`, `invalid_username`, and the
rest), and 14a adds `last_admin` and `environment_sets`.

R6:589-590's error vocabulary gains two codes, each additive under the policy above:

- `unauthenticated`, HTTP 401: a request under `api` with no valid session. A 404 would make a lapsed session look
  like an unknown operation, and the pages need to send a signed-out admin to sign-in.
- `reauthentication_required`, HTTP 403: a gated write from a session anchored outside the window.

Each joins `ops.Codes()`, and `checkStatuses` (`opshttp.go:37-44`) forces its status.

### Job types under the per-album lease

The document publishes the job types `internal/ops/ops.go:86-161` already defines: the opaque handle, the six states,
phase, failure code and message, and four timestamps. Ruling 1 shapes them in one way. A per-album lease admits one
job per album, as the kind lease admits one job per kind (`internal/engine/store/lease.go:166-168`). The document
gives every job-starting operation a `conflict` answer naming the running job, the same answer a second backup gets
(R6:600-602). No queue position or order field enters the wire, and `queued` keeps its present meaning. Rung 9's
import job then adds a job kind, which the policy above makes additive.

## The mount

`Shell.Register` refuses a subtree name that carries a separator (`internal/serve/serve.go:120-121`), so the mount
registers the subtree `api`. The shell hands it the full path, prefix included (`:96-98`). The mount's own handler cuts
`<prefix>/api/v1/` and routes the rest to `opshttp`. It does not use `http.StripPrefix`, whose mismatch answer is plain
text. Every other path under `<prefix>/api`, `api/v2` included, answers 404 with a JSON `not_found` body. That
terminates the owned prefix, which the (14,12) edge requires so that rung 12's later fallthrough cannot capture an
unknown API path (`research/2026-09-04-reverse-edge-completion.md:38`). Because the subtree claims every `api` path,
14a writes nothing in `internal/serve`. Routing to `not_found` runs before session verification; the document is
public, so the order reveals nothing.

A session middleware in a 14a file admits any verified member through the auth seam
(`internal/auth/seam.go:107`). A request with no valid session answers `unauthenticated` in the JSON error shape.
`RequireAdmin` does not serve here: it answers a non-admin with a plain-text 404 (`internal/auth/admin.go:125-128`),
which would refuse a member's `catalog` call. The middleware builds the actor (below) and places the caller on the
context. The authorizer lives in `opshttp`, beside its unexported interface (HISTORY, 6b-ii-a, "To rung 14 (the
mount)"), and reads that caller; `opshttp` imports no `internal/auth`. The plan names the context carrier. The mount
never reads a bearer header, and the origin check covers it like every other path. One log record per admitted
operation carries the actor, the channel, and `client_ip` (security L5 to L7). The shared-origin M2 question stays with
streamer slice 6, where STATUS routes it.

Each new operation joins `ops.Seam`, the privileges table (`internal/ops/catalog.go:78-83`), the catalog, and
`opshttp`'s routes, so the HTTP projection and the CLI verb bind to one vocabulary (R6:580-591). Each CLI verb and its
operation share one write path.

**The job operations answer `unavailable` on this mount.** `opshttp` routes all four existing operations
(`opshttp.go:72-77`), but `serve` builds no job runner (R6:943-944, erratum 6d-R6, which routed job recovery to rung
14), and its unit grants only the `auth` and `store` leaves (`systemd/dubplate-server-serve.service:43`). Ruling 2
decides it: no 0.8 page starts a backup, and rung 16's timer uses the CLI verb. `serve`'s seam answers
`start_backup`, `job_status`, and `cancel_job` with `unavailable`, and its catalog reports each with `available:
false` and a reason (`catalog.go:96-101`). R6:627-636 requires this asymmetry to be named, and this paragraph names
it. `serve`'s runner, its instance lease, and job recovery go to the first rung whose page starts a job over HTTP
(rung 9's import or rung 13).

The roster API at `roster` (`cmd/dubplate-server/serve.go:359`) retires in this pass. Its list and add join `/api/v1`
as operations, and the members page calls them there. A member endpoint outside the contract document would be
undocumented surface, and the 0.9 TUI shares the admin surface (CH:139-142), so leaving add outside the contract only
moves the same work to 0.9. The add operation carries no admin field: the admin grant stays the CLI's alone
(`internal/auth/roster.go:304-305`). This spec names that asymmetry, as R6:627-636 requires. A body naming `admin` is
an unknown key, which `opshttp`'s decoder ignores (`opshttp.go:141-144`), as it ignores an actor field (R6:585-586):
the member is added without the flag.

The proposed operations are `list_members`, `add_member`, `remove_member`, `demote_member`, `list_settings`,
`set_setting`, and `reset_setting`, all at admin privilege. A member is named on the wire by the lowercase hex form of
its `UserID`.

## The network actor (D5)

`ops.NetworkActor` renders `member:<principal>` (`internal/ops/ops.go:42-53`). It becomes `member:<hex UserID>`,
because `identity.go:209-212` names `UserID` the durable key and `Principal` "never the durable key". A renamed
member's jobs and audit rows then keep one actor. Log lines carry the principal as a separate field, for reading.
`NetworkActor` has no production caller today, so no stored job carries the old form. This supersedes "and the network
actor" in R6's 6d-R2 erratum (R6:913-915); the plan edits `identity.go:209-211` and `ops.go:42-46` to match.
`settings_audit.actor` then holds the hex form, which needs a join to read; no 14a surface reads the audit.

`NetworkActor` takes the verified identity and the session's anchor. `ops.Actor` keeps the `UserID` and the anchor in
unexported fields behind read-only accessors, so the write path re-reads the actor's admin flag by id and applies the
re-authentication window to the anchor without parsing its own wire string. `LocalActor` carries neither: the
command line has no session.

## Hook 18: description and sensitive mark

Every registered key carries a description and a sensitive mark: settings, thresholds, wall rows, and
integration-class keys, as SS:1017-1019 asks. Registration panics naming the key when either is missing, on the
registry's existing precedent: every family's `Add*` entry point already panics on a malformed declaration
(`internal/config/registry.go:132-247`). Today `config.Setting` has `Sensitive` (`internal/config/settings.go:76`)
and no description; `Doc` is the accepted-form text a refusal prints (`:67`). `WallRow` has `Name` only
(`internal/config/keys.go:7-9`), `IntegrationClass` names its keys as strings (`:155-160`), and `Threshold` has
neither field (`internal/config/thresholds.go:45-50`).

A Go `bool` cannot tell "not sensitive" from "never declared", so the mark is a closed type with no valid zero
value, on the precedent of `ops.Privilege` (R6:587-588). `sensitive` keeps CH's meaning, "dangerous to change"
(CH:167-169; HISTORY 6d, R-6d6). On a store-homed setting it gates the live write. On a key only the environment
sets, which changes only at a restart, it marks a change that needs care (a path, an address, a credential), and the
configuration reference and `dubplate init` show it. Whether a value is a secret is not the mark (R-6d6); `init`
declares that at its own rung.

The settings read operation is the 14a consumer. It answers every registered key with its name, home, description,
apply mode, and mark. Only a store-homed key also carries its stored value, its source, and its seed value; for a
`restart` key the answer and the page name the stored value as the one the next start applies. A wall or integration
value never crosses the wire, because those rows hold site paths, an address, and credentials. `dubplate init`
(SS:1017-1019) and the configuration-reference generator (STATUS, "Owed") read the same declarations later.

## The settings write path

R6:815-816 left the `admin` source unexported until rung 14. HISTORY 6c, "To rung 14", states what the write
carries: source `admin`, `seed_value` recording the seed it replaced, its channel, and its audit kind.

**One validator.** `set_setting` and `reset_setting` each run the seed rule's own evaluation
(`internal/settings/evaluate.go`) over the post-write rows and the serving process's environment, and refuse when it
refuses. That one check covers parse, bounds, every correlated rule, and the environment contradiction, so no write
from the web can make the next start refuse.

- `set_setting` writes the row with source `admin` and bumps the settings generation in one write transaction
  (R6:770-778). Over an `env-seed` row, `seed_value` records the seed replaced. Over an `admin` row it keeps the earlier
  `seed_value`. Over no row it is empty. One `settings_audit` row records the kind, the actor's wire form, the channel,
  the key, and the old and new values (`internal/engine/store/migrations/0004_settings.sql:10-19`). A wall key or an
  unregistered key is refused.
- **A write against the environment is refused.** R6's seed table refuses `serve`'s start for an `admin` row whose
  environment value differs (R6:809; `evaluate.go:292-294`), so a durable admin value against an environment value is
  already impossible. Committing it with a warning would arm that refusal for the next deploy or reboot, and only box
  access clears it. The evaluation refuses it at the write instead, with `conflict` and detail `environment_sets`. The
  settings page shows such a key as set in the server's environment file, with its value, and offers no edit. Writing
  the environment's own value back commits. The serving environment can lag an edited environment file; a false
  refusal then clears at the next restart.
- `reset_setting` deletes an `admin` row and audits the same way. Over no row or an `env-seed` row it writes nothing
  (no row, audit, generation bump, or notice) and answers success naming the source in force. With the row gone, the
  accessor resolves the default (`internal/settings/accessor.go:113-116`) until the next start seeds the environment's
  value; R6:1002 keeps the `env-seed` source with the seed rule alone. When the serving environment sets the key, the
  answer and the page name both values.
- The actor's admin flag is re-read by `UserID` inside each write transaction, as removal does, so a demotion is final
  for writes already in flight.
- The answer to a write names the key's apply mode.

### The re-authentication gate

A write to a sensitive key on the cookie mount needs a fresh sign-in, "not the cookie a shared laptop holds"
(CH:169-170). Fresh means the session's anchor, the time its magic link was consumed (`internal/auth/session.go:80-86`),
falls within a re-authentication window. A refresh moves the issue time and leaves the anchor (`:138-141`), so
staying signed in never renews freshness. The window is a new store-homed duration setting, itself sensitive,
registered from rung 14's own file with a researched default. A correlated rule holds it below the session absolute
cap, which would otherwise make the gate a no-op, and its floor sits above a magic link's round trip, so the web
cannot lock itself out of resetting it.

The seam implementation decides the gate after it decodes the request, from the actor's anchor and the key's
registered mark. The authorizer cannot hold it: it receives only the operation name (`opshttp/authorize.go:19-21`),
never the key, and `admit` wraps every authorizer denial as `forbidden` (`:100-103`). The gate applies to
`set_setting` and `reset_setting` when the key is sensitive, and always to `add_member`, `remove_member`, and
`demote_member` (D10). A stale session's call answers `reauthentication_required` and writes nothing; the page then
offers a fresh sign-in link. A window setting that cannot be read refuses the write with `internal`, never a default,
on the seam's own rule (`internal/auth/seam.go:104-105`). The plan reads where a consumed link lands today and rules
how the admin returns to the page.

The command-line channel has no session, so the gate does not apply to it: owning the process is its proof
(`internal/identity/identity.go:155-161`).

### The sensitive-change notice

Every committed `admin`-source write or reset of a sensitive key on the cookie mount mails every live admin through
`internal/notify` (`notify.go:89`). After D9, the admin addresses are the roster's live admins. The send runs after
`Update` returns nil, outside the write callback, under its own deadline, so no writer waits on the mail server. A mail
failure is logged and leaves the write committed. A crash between the commit and the send loses the notice; the
`settings_audit` row is the durable record. Neither the seed rule nor the command line sends a notice: both rest on an
operator with box access, and the audit row records each.

### The resolving verb

`dubplate-server settings reset <key>` deletes an `admin` row from the command line, with `serve` stopped or
running. It opens the store through `store.OpenWrite`, at the binary's version only, and keeps its refusals of an
older store, a newer store, and a restore intent (R6:319-323). It writes only the settings row, its audit row, and the
generation, under the same validator.

The verb resolves the three refusals the seed rule raises over an `admin` row on a current store: a stored value the
binary cannot parse (`evaluate.go:264-266`), an environment that contradicts it (`:292-294`), and a value outside the
binary's bounds (`:304-307`). Each message names the verb. An environment contradiction also clears through the
environment, as its message says.

An older store gets no write mode. `Handle.Update` refuses any write whose `user_version` differs from the binary's
(`internal/engine/store/write.go:62-63`, `:78-84`), and R-6d2 (`docs/HISTORY.md:411-416`) keeps an older store
unmigrated so that the rollback binary still runs it. 14a adds no migration, so no store before 14a holds an `admin`
row, and every store that can meet this refusal was written by a binary that has the verb. When `serve`'s pre-open
check refuses an older store over an `admin` row, its message names that recovery: redeploy the previous release, run
its `settings reset`, then upgrade. A later `dubplate upgrade` (1.0) may revisit this.

### Rate limits stay restart-applied

R6:838 left a live apply of the four rate limits as rung 14's option. This spec rules it out. Each limiter is a
keyed map of per-client and per-address buckets (`internal/auth/limit.go:58-72`). A live apply would either walk
every bucket under its lock or rebuild the maps, which refills every bucket and hands an enumerator a fresh burst.
No 0.8 caller needs it. The four keys keep apply mode `restart` (`cmd/dubplate-server/auth.go:166-173`), and the
write answer and the settings page say so.

### Outbound-URL validation waits (D8)

CH:172-173 validates any URL-valued setting against loopback, link-local, and private ranges. No registered setting
holds a URL (R6:193-194), so the validator lands with the first URL-valued key, at that key's write-path validator.

## Member removal and demotion

`remove_member` and `member remove <address>` tombstone the member's row, delete their `member_preferences` rows
(CH:578), and bump the address's revocation epoch. The member's `member_overrides` and `member_quota` rows
(`0005_routed.sql`) survive: overrides are the admin's policy and return with a re-add, and quota reflects audio that
stays. Demotion clears the admin flag. The grant stays CLI-only (R6:1007). The actor's admin flag is re-read inside
the write, as the add path does (`roster.go:380-386`). Every removal and demotion logs one audit record, the log
record the add path writes (`roster.go:321`). CH routes no member-audit table (R6:909-910). A removal or demotion
over the command line therefore leaves its record on the operator's terminal (`member.go:96`), not in the journal;
the tombstone records when, and nothing records who. This spec accepts that.

### The epoch bump and R-6d20 (D6)

Lookup already refuses a tombstoned member (`roster.go:257`), so the tombstone alone ends the member's sessions. The
bump matters for a re-add. Sessions key on the address, and a re-added address revives under its old id
(`roster.go:298-301`, written at `:415-417`), so without a bump a pre-removal session would verify again (HISTORY
6d, segment B Low 3). Removal becomes `EpochStore.Bump`'s first production caller.

The reactivation branch of `addUser` (`roster.go:415-417`) bumps the address's epoch too, in the same callback and
order as removal. The removal's record alone can be lost or outrun. The epoch file lives under the auth leaf, which no
backup carries (`internal/backup/doc.go:1-9`), so a restore to a rebuilt box keeps the tombstone and loses the epoch.
A magic link consumed between the removal's bump and its commit writes no store row, so the write lock does not stop
it. The re-add bump revokes every credential anchored before it, whether or not the removal's record survived. The
removal bump stays, because it also refuses an outstanding magic link for its whole expiry.

That settles R-6d20 (6d plan `plans/2026-09-26-rung-6d-identity-serve-store.md:833-848`): the epoch stays keyed on
the address. A re-added member keeps the old id, so the address-to-id resolution never changes, and a second key
would add a record nothing could disagree with. R6 criterion 7's dual-key clauses and erratum row 10 ("bumps to both
keys") retire. The revoke verb and the rung 5 review's minor finding 14 stay with slice 6.

The epoch is a file record, not a store row (R6:929-930), so "one transaction" holds only by ordering. The bump runs
inside the store's write callback, after the tombstone and the guard and before the commit. A bump failure rolls the
removal back, except `ErrEpochBackward` (`internal/auth/epoch.go:84-87`): after a clock step back the stored epoch is
already later than now and revokes everything the removal must, so the removal logs it and commits. A commit failure
after a bump leaves a live member signed out, which is the safe direction. Since every bump holds the store's
`IMMEDIATE` write lock (R6:341-342), two in two processes serialize. That meets R6 criterion 8 for the only epoch
writers 14a builds. `Bump` notifies subscribers synchronously under a deadline (`epoch.go:98-101`), so a subscriber
would run while every writer on the box waits. No production caller subscribes today; `Subscribe`'s doc comment
gains the contract that a subscriber must not block, because a bump may run under the store's write lock.

### The last-admin guard (D7)

The guard is one statement inside the removal and demotion transaction. It extends the add path's check
(`roster.go:426-432`): after the write, at least one live admin remains, or the transaction rolls back with
`errNoAdmin`. The store has one writer, so two concurrent self-demotions of the last two admins serialize, and the
second is refused. That is 6d criterion 6's clause (R6:1169-1172), routed here by R6:204-207.

### The checklist (CH:576-586)

One test asserts the items that exist at 0.8. The rest belong to the rungs that build their subjects.

| Checklist item | Owner |
|---|---|
| epoch bump | 14a |
| tombstone with a removal time; preferences deleted | 14a |
| collection reassignment (server-written collections pass to the admin) | 14b |
| app passwords; streamer deprovision | rung 12 slice 6 (SS:915; STATUS "Slice 6") |
| playlist transfer | rung 12 slice 7 (SS:921-923) |
| device credentials; writer epoch freeze; in-flight push abort | 0.9, with the local client |

## `DUBPLATE_ADMIN_EMAIL` (D9)

After 6d, `users.admin` is the one source of truth for who is an admin (R6:196-197). The brief frames D9 as a choice
of what the wall row becomes. At HEAD it already grants nothing. `Config.AdminEmail` (`internal/config/config.go:97`)
has no admin-deciding reader. The first admin comes from `member add --admin` (`roster.go:304-305`). The login page's
admin contact is the roster's first admin (`internal/auth/handlers.go:398-407`). The address's one planned reader is
the MusicBrainz contact (`internal/engine/musicbrainz/client.go:85-90`).

The architecture is recorded as this: the wall row seeds the first admin at bootstrap and nothing more. No code path
grants or authorizes admin from the environment. The reason is one source of truth: an environment value that granted
admin could silently re-grant a member an admin removed.

14a edits the runbook to match. Install step 10 (`docs/server-runbook.md:164-170`) bootstraps with a placeholder
address; it becomes `member add` with the `DUBPLATE_ADMIN_EMAIL` address and `--admin`, run only when `member list`
shows no admin, because on a tombstoned address `addUser` reactivates the row with the flag. The key's description
(`:59-60`) gains the seed role. A new line after the restore steps says to re-run any removal made since the snapshot,
because a restore brings back a member removed after it.

`doctor`'s existing `DUBPLATE_ADMIN_EMAIL` row, whose "ok" today means the value parses (`doctor_test.go:1023`), stays
ok when the address names a live admin. It warns, without failing, when the address names a live non-admin, a removed
member, or no member. Before bootstrap the `members` row's empty-roster failure already reports, so the row adds no
second warning. `store.Facts` carries only counts (`internal/engine/store/inspect.go:166-171`), so the plan gives
`Inspect` the fact it needs.

## Riders

- BACKLOG #23: `web/src/lib/api/post.ts:65` maps every 403 to the cross-site copy. The only JSON 403 any `postForm`
  caller receives is the roster API's (`internal/auth/admin.go:318-319`), which retires here, and `opshttp` refuses a
  body that is not JSON (`opshttp.go:97-101`). The members and admin settings pages call `/api/v1` through a new JSON
  client in a 14a file, which reads `error.code` on a JSON 403 and treats only the origin check's plain-text 403 as
  cross-site. `post.ts` stays as it is for its three remaining callers, and #23 closes with the roster API's
  retirement.
- BACKLOG #25: `internal/settings/evaluate.go:265` wraps `s.Validate(stored.Value)` after `parse` failed; it wraps
  the parse error instead. The write path edits this package.
- BACKLOG #12: `member add --help` does not mark `--name` and `--username` required (`member.go:100-107`). Each
  flag's usage says "(required)", and `member remove`, `member demote`, and `settings reset` follow the same form.

## Decisions

D1 to D9 are the brief's, taken as approved above. One more:

- **D10. Adding, removing, and demoting a member on the cookie mount need the re-authentication window.** Each
  changes who holds access, and CH's gate exists because a shared laptop holds the cookie. GitHub applies the same
  rule to "actions that can affect access to accounts and resources"
  (`https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/sudo-mode`, read 2026-10-10). An
  add from a borrowed cookie is the worst of the three, because the access it grants outlives the cookie, and a re-add
  clears a tombstone an admin set. A demotion is also one the web cannot undo, since the grant is CLI-only.

## The migration

14a adds no migration. Re-authentication reads the session anchor, the settings tables exist (`0004`), the member
audit stays a log record, and the `collections` row version is 14b's. The first migration after `0005` therefore
belongs to 14b or rung 7, whichever merges first, and takes `0006`. That migration's plan carries R-6d4's keep limit
and adds its tables to the older-store helper's map (`cmd/dubplate-server/olderstore_test.go:28-33`), per STATUS
"First migration after `0005`". The brief listed a member-audit table as a migration candidate; R6:909-910 rules it
out.

## Forward hooks (SS:945-1040)

| Hook | Verdict |
|---|---|
| 5, stable identity | lands in part: the network actor keys on `UserID`; the epoch stays on the address (D6) |
| 15, configuration homes | lands in part: settings on the cookie mount; collections at 14b; the bearer mount at 0.9 |
| 18, description and sensitive mark | lands: every registered key, with the settings read as first consumer |
| 4 and 20, event-shaped tables | neutral: 14a adds no table, and `settings_audit` is not a serve event |
| 9, transaction model | neutral: removal and the settings write ride the one-writer seam |
| 22, integer collection wire id | neutral: 14a creates no collection |
| 6 | retired (SS:966) |
| 1 to 3, 7, 8, 10 to 14, 16, 17, 19, 21 | neutral: each hook's own text in SS:945-1040 names another rung or a streamer slice |

## Prior art

| Entry | Verdict | Reason |
|---|---|---|
| `github.com/getkin/kin-openapi` v0.149.0 | adopt, tests only | already a module dependency at its newest version; MIT, 16 releases in 12 months, 481 importers of `openapi3filter` (`research/2026-09-25-streamer-prior-art-survey.md:166`); the no-link test keeps it out of the binary; its silent passes and its 2020-12 fallback are closed above |
| `github.com/wneessen/go-mail` v0.8.1 | adopt, through `internal/notify` | the notice reuses the existing send seam; no new dependency |
| `golang.org/x/time/rate` v0.16.0 | keep as is | live apply ruled out above |
| `crypto/hmac`, `crypto/sha256` | neutral | no new credential at 14a; the bearer credential is 0.9's |
| `modernc.org/sqlite` v1.60.1 | neutral | no migration; `modernc.org/libc` moves only with it |
| goose | neutral | no migration at 14a |
| GitHub sudo mode | adopt the rule, decline the timer reset | it gates access-changing actions behind a fresh sign-in; its "two-hour session timeout" resets on each sensitive action, which would need state the anchor already provides |
| OpenSubsonic specification, Navidrome, gonic, taglib, mewkiz/flac, Picard's tables, chromaprint | neutral | none bears on an admin API |

The pass opens with the `dependency-upgrade` sweep (CLAUDE.md, "Current dependencies").

## Planning-miss prevention

- Behaviors relied on are quoted above: `Shell.Register`'s separator refusal, kin-openapi's 3.1 support, its silent
  passes and its 2020-12 fallback, the session anchor across a refresh, `Update`'s version guard, `Bump`'s subscriber
  deadline and its backward-clock error, and the decoder's unknown-key rule. The plan quotes any it adds.
- Privileged fields and every channel that may set them:

| Field | Channels after 14a |
|---|---|
| settings source `admin` | `set_setting` on the cookie mount, inside the window for a sensitive key; none on the command line |
| removal of an `admin` row | `reset_setting` on the cookie mount, inside the window for a sensitive key; `settings reset` on the command line |
| a user's admin flag, set | the CLI grant alone (unchanged) |
| a user's admin flag, cleared | `member demote`; `demote_member` on the cookie mount, inside the window |
| a user's tombstone, set | `member remove`; `remove_member` on the cookie mount, inside the window |
| a user's tombstone, cleared | a re-add: `member add`; `add_member` on the cookie mount, inside the window |
| a revocation epoch | the removal's and the re-add's write callbacks alone |
| a write's freshness | the mount's session middleware, from the verified session's anchor, carried on the actor; never a request field; absent on the command line, which the gate skips |
| a job's or an audit row's actor | the implementation, from the authorizer's actor; never a request field |

- States each check meets, and the report each names:

| Call | State | Answer |
|---|---|---|
| any `/api/v1` call | no valid session | 401 `unauthenticated` |
| any admin operation | the actor is no longer a live admin at the in-write re-read | `forbidden`, nothing written |
| any gated call | session anchored outside the window | `reauthentication_required`, nothing written |
| any gated call | the window setting cannot be read | `internal`, nothing written |
| remove or demote | unknown id, or a removed member | `not_found` |
| any member call | malformed hex id (wrong length, uppercase) | `invalid_request` |
| remove or demote | no live admin would remain | `conflict`, detail `last_admin` |
| remove | self, with another live admin | commits; the bump ends the caller's session, whose next request answers `unauthenticated` |
| demote | target is not an admin | writes nothing, answers success |
| remove | `Bump` returns `ErrEpochBackward` | commits, logged |
| remove | any other bump error | rolls back, `internal` |
| add | each existing refusal | its outcome name as `detail`, under `conflict` for a duplicate and `invalid_request` otherwise |
| `set_setting` | unparseable, out of bounds, a broken correlated rule, a wall key, an unregistered key | `invalid_request` |
| `set_setting` | the serving environment sets the key to another value | `conflict`, detail `environment_sets` |
| `reset_setting` or `settings reset` | no row, or an `env-seed` row | writes nothing, answers success naming the source in force |
| `reset_setting` | the post-reset rows break a correlated rule | `invalid_request` |
| `settings reset` | older store, newer store, restore intent | `OpenWrite`'s own refusals |
| hook 18 registration | missing description, zero mark | panic naming the key |

- Proofs name their fixtures, in the criteria below.

## Contested surfaces with the rung 7 lane

PA:173: "A contested surface without a written owner blocks the second lane." This table is the architecture
decision for the 14a and rung 7a lanes, written to the lane registry before either launches. It amends PA's
`internal/auth` row ("no other lane writes it") for these two lanes. The lane that merges second rebases under the
merge ritual; a conflict between functions two lanes own is mechanical and keeps both sides.

| Surface | Owner | Consumers |
|---|---|---|
| migration `0006`, the older-store map, R-6d4's keep limit | 14a takes none; whichever of 14b and rung 7 merges first takes `0006` and the STATUS item | the other renumbers on rebase and adds its map row |
| the `internal/config` declaration types (hook 18: `Setting`, `Threshold`, `WallRow`, `IntegrationClass`, and `registry.go`'s `Add*` panics) | 14a, landed in its first segment, which also gives every existing key and wall row its description and mark | rung 7 declares its keys in its own file; if 14a merges first, rung 7's rebase adds a description and mark to each; if rung 7 merges first, 14a's hook-18 task marks rung 7's keys |
| `internal/config/keys.go` | 14a for the wall rows' declarations; rung 7 for the `intake` reserved class (`:50-57`), which it retires in the commit that registers its keys | both; the regions are disjoint |
| `internal/ops`, `internal/ops/opshttp`, and the contract document | 14a | rung 7 writes none of them while 14a's lane is open; its JSON endpoints go under `/api/v1` only after 14a merges, and otherwise mount as their own subtree from a rung 7 file, which 14b adopts |
| `internal/auth/roster.go` (remove, demote, the re-add bump, the last-admin guard) and `internal/auth/epoch.go` (`Subscribe`'s contract) | 14a | rung 7 reads the roster through its interface |
| `internal/auth/admin.go` | rung 7 for `RequireAdmin` and the member-level page gate beside it, and the styled 404's hook in their refusals; 14a for removing the roster API's handlers | both; 14a's session middleware for `api` lives in a new 14a file |
| `internal/serve` | rung 7, for the styled 404's edit of `Shell.terminated` (`internal/serve/serve.go:248-258`), under PA's rung 12 row; 14a writes nothing here, because the `api` subtree terminates its own prefix | 14a registers through `Shell.Register` |
| `cmd/dubplate-server/serve.go` registration | each lane registers from its own file through `Shell.Register` (PA:205); 14a also removes the `roster` registration at `:359` | both |
| `cmd/dubplate-server/doctor.go` rows | each lane adds rows from its own file; 14a edits the `DUBPLATE_ADMIN_EMAIL` row in place | both |
| `webDocuments` and `webRoutes.register` (`cmd/dubplate-server/web.go:104-108`, `:185-195`) | each lane appends its own rows | both |
| `web/` | rung 7 (PA:210); 14a owns `routes/members`, the admin settings route, and its `/api/v1` client module, and leaves `post.ts` unchanged | rung 7 owns `SiteHeader.svelte`'s navigation, the deck gates, and the rendered harness; 14a appends its rows and its navigation entry |
| the route inventory (PS:581, row A2) | the first lane to register a route lands the inventory test | the other adds its rows |
| `go.mod` and `web/package.json` | the first lane to launch runs the pass-opening sweep and merges it as its own commit | the second takes it by rebase; rung 7's new dependencies are rung 7's |
| the journey vocabulary's wire projection | rung 7 | 14a's document makes a new value additive |

## Process

`auth-data` is the pass class for the auth, settings, and member tasks. Review follows the brief's section 6:
`web-auth-security-reviewer` after the segment that lands the write path and the re-authentication gate, and again
after removal and the pages; `svelte-reviewer` and `daisyui-a11y-reviewer` against the visual identity spec; one
`go-architecture-reader` per touched package at the merge. Every user-path page is rendered and read before Geoff
sees it. The plan derives tasks, segments, ceiling, and clock estimate from the brief's section 6, less its migration
candidates and the older-store write mode revision 1 carried.

## Acceptance criteria

1. One hand-written OpenAPI 3.1 document describes every `/api/v1` operation, its request and response bodies, and
   every wire code and detail it can answer. A test fails when a catalog operation has no path in the document or a
   path names no catalog operation. A test walks the document and fails on a schema keyword outside the enforced set.
2. A request a test means to be well formed passes `ValidateRequest`. A request a test means to be malformed at the
   schema level fails it. Every answer to a routed request, refusals included, passes `ValidateResponse` with
   `IncludeResponseStatus: true`, and each other kin-openapi silent pass fails the test. A transport refusal and an
   unrouted path validate their body against the error schema directly. Three proofs fail as they should: a renamed
   `required` response field, a Go response field the document omits, and an undocumented status. kin-openapi appears
   only in `_test.go` files, and the no-link test stays green.
3. The document states the additive-within-v1 policy, enumeration and detail extension and the journey vocabulary
   included.
4. `api` is registered through `Shell.Register` from rung 14's own file. Every path under `<prefix>/api` that names
   no v1 operation, `api` and `api/v2/...` included, answers 404 with a JSON `not_found` body.
5. The mount is built on `identity.ChannelCookie` and never reads a bearer header. A request with no valid session
   answers 401 `unauthenticated` and reaches no operation. A member session reaches `catalog` and is refused every
   admin operation with `forbidden`. A body carrying an actor field is ignored. Each admitted operation logs actor,
   channel, and `client_ip`.
6. A cookie-mount `set_setting` records `settings_audit.actor` as `member:<hex UserID>`. After a direct store write
   renames the member, a later write records the same actor. The log line carries the principal as its own field.
7. The document gives `start_backup` a `conflict` answer for a held lease, and no job wire type carries a queue
   position or order field. On the cookie mount the three job operations answer `unavailable`, and the catalog reports
   each with `available: false` and a reason.
8. Every registered setting, threshold, wall row, and integration-class key carries a description and an explicit
   sensitive mark. A test proves registration panics naming the key for a missing description and for a zero mark,
   and the existing construction test builds the binary's whole registry.
9. `list_settings` answers every registered key's name, home, description, apply mode, and mark, plus stored value,
   source, and seed value for store-homed keys only. With every wall and integration row set to a distinct sentinel,
   no sentinel appears in the raw body of any `/api/v1` answer, refusals included.
10. `set_setting` writes source `admin`, the replaced seed in `seed_value`, the generation bump, and one audit row
    with kind, actor, channel, key, and old and new values. An unparseable value, an out-of-bounds value, a broken
    correlated rule, a wall key, and an unregistered key each answer `invalid_request` and change nothing. A value the
    serving environment contradicts answers `conflict` with detail `environment_sets` and changes nothing; the
    environment's own value commits.
11. A write's answer names the apply mode. A reset whose post-reset rows break a correlated rule answers
    `invalid_request` and changes nothing. A reset over no row or an `env-seed` row writes nothing. After a reset of a
    key the serving environment sets, the answer names the default now in force and the value the next start seeds.
12. On the cookie mount, from a session anchored outside the re-authentication window, a sensitive write, a sensitive
    reset, an add, a removal, and a demotion each answer `reauthentication_required` and write nothing. The same
    calls from a session anchored inside it commit. A stale session's non-sensitive write and non-sensitive reset
    commit. A refreshed session keeps its anchor and its staleness. An unreadable window refuses. The window's
    correlated rule refuses a window at or above the absolute cap.
13. Each committed cookie-mount write or reset of a sensitive key mails every live admin, with a fixture of two live
    admins and one removed admin, who receives nothing. A non-sensitive write and a command-line reset mail no one.
    A mail failure is logged and the write stays committed. A seed-rule change mails no one.
14. With `serve` stopped, `settings reset <key>` on a current store resolves each refusal: an `admin` row the binary
    cannot parse, one the environment contradicts, and one outside the binary's bounds. `serve` then starts. The verb
    refuses an older store, a newer store, and a restore intent with `OpenWrite`'s messages. All three seed-rule
    messages name the verb, and `serve`'s pre-open refusal of an older store names the previous-release recovery.
15. `remove_member` and `member remove` tombstone the row, delete the member's preferences, keep their overrides and
    quota, and bump the address's epoch; the fixture writes a preference, an override, and a quota row first. A
    session minted before the removal is refused, and stays refused after the address is re-added under its old id,
    also when the removal's epoch record is deleted before the re-add. With the injected clock at least one second
    past the re-add's bump, a sign-in verifies. An injected bump failure leaves the member live, and
    `ErrEpochBackward` commits the removal. This test is the checklist test for the 14a rows of the checklist table.
16. A removal or demotion that would leave no live admin answers `conflict` with detail `last_admin` and changes
    nothing. The concurrency proof holds the store's write lock from a second connection with one admin's
    demotion uncommitted, starts the other admin's demotion, then commits the first: exactly one admin remains, and
    the same test fails against a guard moved outside the transaction. An actor demoted before a removal, demotion, or
    settings write is refused.
17. `add_member` over `/api/v1` cannot grant admin: a body carrying an admin field adds the member without the flag.
    The `roster` subtree answers 404.
18. Every state in the planning-miss table answers as the table says.
19. The members page lists, adds, removes, and demotes members over `/api/v1`, confirms before a removal or
    demotion, and shows each refusal's own copy (last admin, re-authentication needed, forbidden, signed out). Each
    add refusal still moves focus to its field. A JSON 403 never shows the cross-site copy.
20. The admin settings page, at a route other than `settings` (which the visual identity spec reserves for the member
    preferences page, `specs/2026-09-24-visual-identity-design.md:145-146`), lists every key with its description
    and mark, edits and resets store-homed keys, shows an environment-set key as set in the environment file with no
    edit, shows each write's apply mode, and leads a stale session to a fresh sign-in. Both pages follow the visual identity spec, are rendered and read, and pass the
    `web/` leg's static and rendered gates.
21. No code path reads `DUBPLATE_ADMIN_EMAIL` to grant or authorize admin. `doctor`'s `DUBPLATE_ADMIN_EMAIL` row
    warns, without failing, when the address names a live non-admin, a removed member, or no member, reports ok when
    it names a live admin, and adds nothing before bootstrap. The runbook's step 10, key description, and restore
    line read as "`DUBPLATE_ADMIN_EMAIL`" above.
22. BACKLOG #12 and #25 are closed by their riders, and #23 by the roster API's retirement.

## Deferred to 14b

- Collection endpoints (retire, purge, reassign, and shape edits) with `if_version` (CH:296-298; R6:183-184).
- The private-visibility refusal (RP:146-148), which slice 3 later lifts.
- Per-member override endpoints (CH:162-165). Their first reader is rung 7's intake.
- Collection reassignment on removal.
- The collection page.
- Adopting any JSON subtree rung 7 mounted outside `/api/v1`.

## Deferred past 14a

- `serve`'s job runner, its instance lease, and job recovery, to the first rung whose page starts a job over HTTP.
- To 0.9: the bearer mount. The shell-level test that a bypassed route never reads the session cookie goes with the
  first bypass consumer (`internal/serve/origin.go:37-50`), which may now be a streamer slice rather than this mount.
- To 0.9: non-browser token acquisition and service tokens (BS:985-988), version negotiation on connect
  (CH:205-207), and the adopt and seize verbs.
- To 0.9: device-credential revocation, the writer epoch freeze, and the in-flight push abort on removal.

## Rulings for Geoff

None. The review raised three candidate forks (A7 with C2, A10, and B5), and each has a dominant answer recorded in
the fold record: a write
against the environment is refused, because R6's ratified seed table already forbids the state it would create;
the older-store write mode is dropped, because R-6d2's rollback path already recovers it without bypassing the
store's version guard; and add joins D10's gate, because D10's own reason covers it.
