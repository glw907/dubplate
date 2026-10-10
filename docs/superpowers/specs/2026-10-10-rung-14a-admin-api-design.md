# Rung 14a: the admin API's first half

Status: draft for Geoff's read, 2026-10-10. Design approved in session the same day; this document records it.

This spec covers the first of rung 14's two passes. It records the contract-document form, the authorization
model at 0.8, and the decisions the planning brief (`research/2026-10-10-next-pass-planning-brief.md`) left open.
Most of the design is already settled in CH and R6, so this spec cites them rather than restating them.

Abbreviations follow the brief: BS is the build spec (`specs/2026-09-02-greenfield-server-build.md`), RP the re-plan
(`specs/2026-10-05-rungs-12-16-replan-design.md`), SS the streamer spec (`specs/2026-09-07-streamer-2-0-design.md`),
CH the configuration homes spec (`specs/2026-09-11-configuration-homes-design.md`), R6 the rung 6 spec
(`specs/2026-09-23-rung-6-ops-seam-store.md`), and PA the parallel build approach
(`specs/2026-09-03-parallel-build-approach.md`). Paths sit under `docs/superpowers/` unless they start at the repo
root. Code citations were read at `e6fdf2b`.

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
verb; member removal and demotion over the CLI and `/api/v1`; the members and settings pages.

**14b, after rung 7 writes the first collection row** (R6:185-187): the collection endpoints with `if_version`,
the private-visibility refusal, the per-member override endpoints, and the collection page.

## The contract document (D4)

The contract is one hand-written OpenAPI 3.1 file. `go doc github.com/getkin/kin-openapi/openapi3` at v0.149.0 reads
"Supports OpenAPI 3.0, OpenAPI 3.1, and OpenAPI 3.2". The plan names the file's path. Nothing serves it at 14a.

Tests validate every `/api/v1` request they send and every answer they receive against the document, through
`openapi3filter.ValidateRequest` and `ValidateResponse`. kin-openapi appears in `_test.go` files only, because the
no-link test (`internal/harness/conform/arch_test.go:216`) forbids any `github.com/getkin` path in
`cmd/dubplate-server`'s dependency set. The OpenSubsonic layer took the same hand-written route.

The document states the version policy, and the policy is additive within v1. These changes are additive: a new
operation, an optional request field, a new response field, and a new value in any enumeration. The enumerations
are job states, job kinds, wire codes, and the journey vocabulary's wire projection. A client must ignore an unknown
field and treat an unknown enumeration value as opaque. Removing or renaming any of these, or making an optional
field required, is a v2 change. This rules the journey question the reverse-edge record left to rung 14
(`research/2026-09-04-reverse-edge-completion.md:223`): rung 7 owns the vocabulary, and adding a state to it is
additive.

### Job types under the per-album lease

`/api/v1` publishes the job types `internal/ops/ops.go:86-161` already defines: the opaque handle, the six states,
phase, failure code and message, and four timestamps. Ruling 1 shapes them in one way. A per-album lease admits one
job per album, as the kind lease admits one job per kind (`internal/engine/store/lease.go:166-168`). A start that
finds the lease held answers `conflict` naming the running job, the same answer a second backup gets (R6:600-602).
No queue position or order field enters the wire, and `queued` keeps its present meaning. Rung 9's import job then
adds a job kind, which the policy above makes additive.

## The mount

`Shell.Register` refuses a subtree name that carries a separator (`internal/serve/serve.go:120-121`), so the mount
registers the subtree `api`, and its handler routes `v1`. The brief does not record this constraint. Every path under
`<prefix>/api` that names no v1 operation, `api/v2` included, answers 404 with a JSON body carrying `not_found`. That
terminates the owned prefix, which the (14,12) edge requires so that rung 12's later fallthrough cannot capture an
unknown API path (`research/2026-09-04-reverse-edge-completion.md:38`). `opshttp` already answers an unknown
operation this way (`internal/ops/opshttp/opshttp.go:85`).

The mount is `opshttp`'s handler over the ops seam, built with `identity.ChannelCookie` and an authorizer that reads
the verified session from the request context (`internal/ops/opshttp/doc.go:6-19`). It never reads a bearer header,
and the origin check covers it like every other path. Each new operation joins `ops.Seam`, the privileges table
(`internal/ops/catalog.go:76-81`), and the catalog, so the HTTP projection and the CLI verb bind to one vocabulary
(R6:580-591). Each CLI verb and its operation share one write path.

The roster API at `roster` (`cmd/dubplate-server/serve.go:359`) retires in this pass. Its list and add join `/api/v1`
as operations, and the members page calls them there. Two reasons decide it. A member endpoint outside the contract
document is undocumented surface, against the greenfield rulings' ruling 8 ("the documented API"). And the 0.9 TUI shares the admin surface
(CH:147-150), so leaving add outside the contract only moves the same work to 0.9. The add operation carries no admin
field: the admin grant stays the CLI's alone (`internal/auth/roster.go:304-305`). This spec names that asymmetry,
as R6:627-636 requires.

The proposed operations are `list_members`, `add_member`, `remove_member`, `demote_member`, `list_settings`,
`set_setting`, and `reset_setting`, all at admin privilege. A member is named on the wire by the hex form of its
`UserID`.

## The network actor (D5)

`ops.NetworkActor` renders `member:<principal>` (`internal/ops/ops.go:42-53`). It becomes `member:<hex UserID>`,
because `identity.go:209-212` names `UserID` the durable key and `Principal` "never the durable key". A renamed
member's jobs then keep one actor. Log lines carry the principal as a separate field, for reading. `NetworkActor` has
no production caller today, so no stored job carries the old form.

## Hook 18: description and sensitive mark

Every registered key carries a description and a sensitive mark: settings, thresholds, wall rows, and
integration-class keys. Registration panics naming the key when either is missing, on the registry's existing
collision-gate precedent. Today `config.Setting` has `Sensitive` (`internal/config/settings.go:76`) and no
description; `Doc` is the accepted-form text a refusal prints (`:67`). `WallRow` has `Name` only
(`internal/config/keys.go:7-9`), `IntegrationClass` names its keys as strings (`:155-160`), and `Threshold` has
neither field (`internal/config/thresholds.go:45-50`).

A Go `bool` cannot tell "not sensitive" from "never declared", so the mark is a closed type with no valid zero
value, on the precedent of `ops.Privilege` (R6:587-588). `sensitive` keeps CH's meaning, "dangerous to change live"
(CH:167-169; HISTORY 6d, R-6d6).

The settings read operation is the 14a consumer. It answers every registered key with its name, home, description,
apply mode, and mark. Only a store-homed key also carries its value in force, its source, and its seed value. A wall
or integration value never crosses the wire, because those rows hold site paths, an address, and credentials.
`dubplate init` (SS:1017-1019) and the configuration-reference generator (STATUS, "Owed") read the same declarations
later.

## The settings write path

R6:815-816 left the `admin` source unexported until rung 14. HISTORY 6c, "To rung 14", states what the write
carries: source `admin`, `seed_value` recording the seed it replaced, its channel, and its audit kind.

- `set_setting` validates the value against the key's bounds and every correlated rule over the post-write set, then
  writes the row with source `admin` and bumps the settings generation in one write transaction (R6:770-778). Over
  an `env-seed` row, `seed_value` records the seed replaced. Over an `admin` row it keeps the earlier `seed_value`.
  Over no row it is empty. One `settings_audit` row records the kind, the actor's wire form, the channel, the key,
  and the old and new values (`internal/engine/store/migrations/0004_settings.sql:10-19`). A wall key or an
  unregistered key is refused.
- `reset_setting` deletes an `admin` row, so the next start seeds the key from the environment or resolves its
  default. It audits the same way.
- The answer to a write names the key's apply mode. When the serving process's environment sets the key to another
  value, the answer also says that the next start will refuse (R6:809) and names both resolutions. The seed rule
  stays as ratified.

### The re-authentication gate

A write to a sensitive key on the cookie mount needs a fresh sign-in, "not the cookie a shared laptop holds"
(CH:169-170). Fresh means the session's anchor, the time its magic link was consumed (`internal/auth/session.go:80-86`),
falls within a re-authentication window. A refresh moves the issue time and leaves the anchor (`:123-125`), so
staying signed in never renews freshness. The window is a new store-homed duration setting, itself sensitive,
registered from rung 14's own file with a researched default. Freshness comes from the verified session alone,
never from a request field.

A stale session's write answers a distinct refusal and writes nothing. The page then offers a fresh sign-in link.
The plan reads where a consumed link lands today and rules how the admin returns to the page.

Removal and demotion on the cookie mount take the same gate (D10). The command-line channel has no session, so the
gate does not apply to it: owning the process is its proof (`internal/identity/identity.go:155-161`).

### The sensitive-change notice

Every committed `admin`-source write or reset of a sensitive key, on any channel, mails every live admin through
`internal/notify` (`notify.go:89`). After D9, the admin addresses are the roster's live admins. A mail failure is
logged and leaves the write committed. The seed rule's own changes send no notice: they come from the box's
environment file, which only an operator with box access edits.

### The resolving verb

`dubplate-server settings reset <key>` deletes an `admin` row from the command line, with `serve` stopped or
running. It resolves the two refusals the operator cannot fix through the environment alone: an `admin` row the
environment contradicts, and an `admin` row outside a new binary's bounds.

The verb must also open a store older than the binary. The start order makes an older store whose seed evaluation
refuses stay unmigrated (HISTORY 6d, R-6d2, `docs/HISTORY.md:411-416`). `store.OpenWrite` refuses any version but
the binary's (R6:319-323). So the verb accepts a store at the binary's version or older, provided it holds the
settings table (version 4 or later). It refuses a newer store and a restore intent, with the store's own messages.
It writes only the settings row, its audit row, and the generation. The plan picks the open mechanism. The seed
rule's two refusal messages (`internal/settings/evaluate.go:293-294`, `:306`) name the verb.

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
(CH:578), and bump the address's revocation epoch. Demotion clears the admin flag. The grant stays CLI-only
(R6:1007). The actor's admin flag is re-read inside the write, as the add path does (`roster.go:380-386`). Every
removal and demotion logs one audit record, the log record the add path writes (`roster.go:321`). CH routes no
member-audit table (R6:909-910).

### The epoch bump and R-6d20 (D6)

Lookup already refuses a tombstoned member (`roster.go:257`), so the tombstone alone ends the member's sessions. The
bump matters for a re-add. Sessions key on the address, and a re-added address revives under its old id
(`roster.go:298-301`, written at `:415-417`), so without a bump a pre-removal session would verify again (HISTORY
6d, segment B Low 3). Removal becomes `EpochStore.Bump`'s first production caller.

That settles R-6d20 (6d plan `plans/2026-09-26-rung-6d-identity-serve-store.md:833-848`): the epoch stays keyed on
the address. A re-added member keeps the old id, so the address-to-id resolution never changes, and a second key
would add a record nothing could disagree with. R6 criterion 7's dual-key clauses and erratum row 10 ("bumps to both
keys") retire. The revoke verb and the rung 5 review's minor finding 14 stay with slice 6.

The epoch is a file record, not a store row (R6:929-930), so "one transaction" holds only by ordering. The bump runs
inside the store's write callback, after the tombstone and the guard and before the commit. A bump failure rolls the
removal back. A commit failure after a bump leaves a live member signed out, which is the safe direction. Since every
removal holds the store's `IMMEDIATE` write lock (R6:341-342) across its bump, two removals in two processes
serialize. That meets R6 criterion 8 for the only epoch writer 14a builds. `Bump` notifies subscribers under a
deadline (`internal/auth/epoch.go:98-101`). No production caller subscribes today, and the plan confirms that
nothing holds the write lock past that deadline.

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
has no admin-deciding reader. The first admin comes from `member add --admin` (`roster.go:431`). The login page's
admin contact is the roster's first admin (`internal/auth/handlers.go:398-407`). The address's one planned reader is
the MusicBrainz contact (`internal/engine/musicbrainz/client.go:85-90`).

The architecture is recorded as this: the wall row seeds the first admin at bootstrap and nothing more. The runbook's
bootstrap step adds `member add <DUBPLATE_ADMIN_EMAIL> --admin` (5d plan
`plans/2026-09-25-rung-5d-boundary-deployment-close.md:1399`). No code path grants or checks admin from the
environment. After bootstrap, `doctor` warns, without failing, when the address names no live admin. The reason is
one source of truth: an environment value that granted admin could silently re-grant a member an admin removed.

## Riders

- BACKLOG #23: `web/src/lib/api/post.ts:65` maps every 403 to the cross-site copy. The members page's new refusals
  would land on it, so the client reads a JSON outcome on a 403 first and falls back to `cross_origin` only for the
  origin check's plain-text body.
- BACKLOG #25: `internal/settings/evaluate.go:265` wraps `s.Validate(stored.Value)` after `parse` failed; it wraps
  the parse error instead. The write path edits this package.
- BACKLOG #12: `member add --help` does not mark `--name` and `--username` required (`member.go:100-107`). Each
  flag's usage says "(required)", and `member remove`, `member demote`, and `settings reset` follow the same form.

## Decisions

D1 to D9 are the brief's, taken as approved above. One more:

- **D10. Removal and demotion on the cookie mount need the re-authentication window.** Each changes who holds
  access, and CH's gate exists because a shared laptop holds the cookie. GitHub applies the same rule to "actions
  that can affect access to accounts and resources"
  (`https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/sudo-mode`, read 2026-10-10). A
  demotion is also one the web cannot undo, since the grant is CLI-only.

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
| 1 to 3, 7, 8, 10 to 14, 16, 17, 19, 21 | neutral: each names rung 6, 7, 8, 9, 13, or a streamer slice (RP:110-111) |

## Prior art

| Entry | Verdict | Reason |
|---|---|---|
| `github.com/getkin/kin-openapi` v0.149.0 | adopt, tests only | already a module dependency at its newest version; MIT, 16 releases in 12 months, 481 importers of `openapi3filter` (`research/2026-09-25-streamer-prior-art-survey.md:166`); the no-link test keeps it out of the binary |
| `github.com/wneessen/go-mail` v0.8.1 | adopt, through `internal/notify` | the notice reuses the existing send seam; no new dependency |
| `golang.org/x/time/rate` v0.16.0 | keep as is | live apply ruled out above |
| `crypto/hmac`, `crypto/sha256` | neutral | no new credential at 14a; the bearer credential is 0.9's |
| `modernc.org/sqlite` v1.60.1 | neutral | no migration; `modernc.org/libc` moves only with it |
| GitHub sudo mode | adopt the rule, decline the timer reset | it gates access-changing actions behind a fresh sign-in; its "two-hour session timeout" resets on each sensitive action, which would need state the anchor already provides |
| OpenSubsonic specification, Navidrome, gonic, taglib, mewkiz/flac, Picard's tables, chromaprint | neutral | none bears on an admin API |

The pass opens with the `dependency-upgrade` sweep (CLAUDE.md, "Current dependencies").

## Planning-miss prevention

- Behaviors relied on are quoted above: `Shell.Register`'s separator refusal, kin-openapi's 3.1 support, the session
  anchor across a refresh, `OpenWrite`'s version refusal, and `Bump`'s subscriber deadline. The plan quotes any it
  adds.
- States each check meets: the re-authentication gate (no session, stale anchor, fresh anchor, command line); the
  resolving verb (current, older with the settings table, older without it, newer, restore intent, no row, an
  `env-seed` row); removal (live member, already removed, unknown id, last admin, self); hook 18 registration (missing
  description, zero mark).
- Privileged fields and every channel that may set them:

| Field | Channels after 14a |
|---|---|
| settings source `admin` | `set_setting` on the cookie mount; none on the command line |
| removal of an `admin` row | `reset_setting` on the cookie mount; `settings reset` on the command line |
| a user's admin flag, set | the CLI grant alone (unchanged) |
| a user's admin flag, cleared | `member demote`; `demote_member` on the cookie mount, inside the re-authentication window |
| a user's tombstone | `member remove`; `remove_member` on the cookie mount, inside the window; cleared only by a re-add |
| a revocation epoch | the removal's write callback alone |
| a job's actor | the implementation, from the authorizer's actor; never a request field |

- Proofs name their fixtures: the concurrency proof needs two live admins and two demotions; the re-add proof needs
  a session minted before the removal; the older-store proof needs a store at a version below the binary's with an
  `admin` row the binary refuses.

## Contested surfaces with the rung 7 lane

Written to the lane registry before either lane launches.

| Surface | Rule |
|---|---|
| migration `0006` and the older-store map | 14a takes none; whichever of 14b and rung 7 merges first takes `0006` |
| `internal/serve` registration and `cmd/dubplate-server/serve.go` | each lane registers from its own file through `Shell.Register` (PA:205); 14a also removes the `roster` registration at `serve.go:359` |
| `web/` | 14a owns `routes/members` and a new settings route; rung 7 owns the uploader; `web/src/lib/api/post.ts` and the shared layout's navigation are contested, landed by whichever lane merges first |
| the journey vocabulary's wire projection | rung 7 owns the vocabulary; enumeration extension is additive (above) |

## Process

`auth-data` is the pass class for the auth, settings, and member tasks. Review follows the brief's section 6:
`web-auth-security-reviewer` after the segment that lands the write path and the re-authentication gate, and again
after removal and the pages; `svelte-reviewer` and `daisyui-a11y-reviewer` against the visual identity spec; one
`go-architecture-reader` per touched package at the merge. Every user-path page is rendered and read before Geoff
sees it. The plan derives tasks, segments, ceiling, and clock estimate from the brief's section 6, less its migration
candidates.

## Acceptance criteria

1. One hand-written OpenAPI 3.1 document describes every `/api/v1` operation, its request and response bodies, and
   every wire code it can answer. A test fails when a catalog operation has no path in the document or a path names
   no catalog operation.
2. Every request the `/api/v1` tests send and every answer they receive, errors included, validates against the
   document through kin-openapi. A proof shows a renamed response field fails validation. kin-openapi appears only in
   `_test.go` files, and the no-link test stays green.
3. The document states the additive-within-v1 policy, enumeration extension and the journey vocabulary included.
4. `api` is registered through `Shell.Register` from rung 14's own file. Every path under `<prefix>/api` that names
   no v1 operation, `api/v2/...` included, answers 404 with a JSON `not_found` body.
5. The mount is built on `identity.ChannelCookie` and never reads a bearer header. A request with no valid session
   reaches no operation. A member session reaches `catalog` and is refused every admin operation. A body carrying an
   actor field is ignored.
6. A job started over `/api/v1` records the actor `member:<hex UserID>`. After the member's username changes, a
   later job records the same actor. The log line carries the principal as its own field.
7. The document gives every job-starting operation a `conflict` answer for a held lease, and no job wire type
   carries a queue position or order field.
8. Every registered setting, threshold, wall row, and integration-class key carries a description and an explicit
   sensitive mark. Registration panics naming the key when either is missing. A test over the binary's whole registry
   finds none missing.
9. `list_settings` answers every registered key's name, home, description, apply mode, and mark, plus value,
   source, and seed value for store-homed keys only. No wall or integration value appears in any answer.
10. `set_setting` writes source `admin`, the replaced seed in `seed_value`, the generation bump, and one audit row
    with kind, actor, channel, key, and old and new values. An out-of-bounds value, a broken correlated rule, a wall
    key, and an unregistered key each answer `invalid_request` and change nothing.
11. A write's answer names the apply mode. When the serving environment sets the key to another value, it names the
    next start's refusal and both resolutions.
12. On the cookie mount, a sensitive write, a removal, and a demotion from a session anchored outside the
    re-authentication window are refused with a distinct code and write nothing. The same calls from a session
    anchored inside it commit. A refreshed session keeps its anchor and its staleness.
13. Each committed `admin`-source write or reset of a sensitive key mails every live admin. A mail failure is logged
    and the write stays committed. A seed-rule change mails no one.
14. `settings reset <key>` with `serve` stopped resolves each refusal: an `admin` row the environment contradicts,
    an `admin` row outside the binary's bounds, and an older store that `serve`'s pre-open check refused. `serve`
    then starts. The verb refuses a newer store and a restore intent. Both seed-rule refusal messages name it.
15. `remove_member` and `member remove` tombstone the row, delete the member's preferences, and bump the address's
    epoch. A session minted before the removal is refused, and stays refused after the address is re-added under its
    old id. A sign-in after the re-add verifies. An injected bump failure leaves the member live.
16. A removal or demotion that would leave no live admin is refused and changes nothing. Two concurrent
    self-demotions of the last two admins leave exactly one admin. An actor demoted before the write is refused.
17. `add_member` over `/api/v1` cannot grant admin: a body carrying an admin field is refused and changes nothing.
    The `roster` subtree answers 404.
18. One checklist test asserts the epoch bump, the tombstone with its removal time, and the preference deletion.
19. The members page lists, adds, removes, and demotes members over `/api/v1`, confirms before a removal or
    demotion, and shows each refusal's own copy (last admin, re-authentication needed, forbidden). A JSON 403 never
    shows the cross-site copy.
20. The settings page lists every key with its description and mark, edits and resets store-homed keys, shows each
    write's apply mode, and leads a stale session to a fresh sign-in. Both pages follow the visual identity spec, are
    rendered and read, and pass the `web/` leg's static and rendered gates.
21. No code path reads `DUBPLATE_ADMIN_EMAIL` to grant or check admin. `doctor` warns, without failing, when the
    address names no live admin, and reports ok when it names one.
22. BACKLOG #12, #23, and #25 are closed by their riders.

## Deferred to 14b

- Collection endpoints (retire, purge, reassign, and shape edits) with `if_version` (CH:296-298; R6:183-184).
- The private-visibility refusal (RP:146-148), which slice 3 later lifts.
- Per-member override endpoints (CH:162-165). Their first reader is rung 7's intake.
- Collection reassignment on removal.
- The collection page.

## Deferred to 0.9

- The bearer mount. The shell-level test that a bypassed route never reads the session cookie goes with the first
  bypass consumer (`internal/serve/origin.go:37-50`), which may now be a streamer slice rather than this mount.
- Non-browser token acquisition and service tokens (BS:986-989).
- Version negotiation on connect (CH:205-207).
- The adopt and seize verbs.
- Device-credential revocation, the writer epoch freeze, and the in-flight push abort on removal.

## Rulings for Geoff

None. The approved design and rulings 1 to 3 settle every product fork this spec meets.
