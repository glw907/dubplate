# Next pass: planning brief (2026-10-10)

This brief prepares the planning session that follows rung 12 slice 2b. It was written against master at `6d90dcd`.
Every claim about the tree cites a file and line read at that commit. The planning session should still re-read
any fact it builds on (CLAUDE.md, "Planning-miss prevention"). Abbreviations used below: BS is the build spec
(`specs/2026-09-02-greenfield-server-build.md`), RP is the re-plan (`specs/2026-10-05-rungs-12-16-replan-design.md`),
SS is the streamer spec (`specs/2026-09-07-streamer-2-0-design.md`), CH is the configuration homes spec
(`specs/2026-09-11-configuration-homes-design.md`), R6 is the rung 6 spec (`specs/2026-09-23-rung-6-ops-seam-store.md`),
and PA is the parallel build approach (`specs/2026-09-03-parallel-build-approach.md`). All paths sit under
`docs/superpowers/` unless they start at the repo root.

## 1. Which pass first

Plan rung 14 next. Slice 3 cannot start yet.

The specs gate slice 3 behind rung 9, and rung 9 has not run. RP:126 reads "Slices 3 onward start after rung 9 lands
the catalog", and SS:319-320 and BS's rung 12 erratum say the same. STATUS's "Slice 3 is ready to plan" is true of
2b's evidence, but the build order says otherwise. The schema agrees with the build order. The catalog tables
in `internal/engine/store/migrations/0003_catalog.sql:36-64` carry ids, MusicBrainz ids, `rel_path`, and timestamps
only. The projector SS:448-449 describes reads "the tags as applied, gain values" and the folded sort key, and those
columns arrive with rungs 7 and 9 (the erratum at SS:990-995: "the audio key, file hash, and sort key migrate in
with rungs 7 and 9"). A slice 3 projector built now would have nothing to project.

Rung 14 is the spine's next rung (CLAUDE.md, "Build process": spine 6, 14, 7, 9), and its inbound edges are mostly
soft. The edge analysis classes (14,7), (14,12), and (14,13) as soft and (14,9) as **hard**
(`research/2026-09-04-reverse-edge-completion.md:36-39`). The hard edge's blocking artifact is rung 9's
concurrency-model ruling, not rung 9's code: "Rung 14's wire types are a published projection of the ops job model,
which rung 9's concurrency ruling shapes" (`:37`, argued at `:227-235`). PA:118 schedules that ruling "before lane V
drafts its plan", and CLAUDE.md lists it among the expected owner touchpoints. It is question 1 below.

Nothing slice 3 builds is an input to rung 14, with one exception. RP:146-148 keeps private visibility
unsettable "until slice 3's projector enforces visibility on every endpoint", so rung 14's collection write refuses
`private` and slice 3 later lifts that refusal. Rung 14 builds nothing slice 3 consumes either. Slice 3 is a
Subsonic surface on its own route, and its bearer-bypass use arrives at slice 5 at the earliest (RP:228-230).

The two cannot run as lanes now, because slice 3 is blocked. The useful second lane is rung 7, for three reasons.
PA:116 pairs lane V (rung 14) with lane I (rung 7). Rung 7 needs no concurrency ruling. And 7, then 9, then slices
3 to 8, is the critical path to 0.8; rung 14 is not on it. Lane I's territory is `internal/intake`,
`internal/reject`, and `web/` (PA:113). The contested surfaces between the two lanes, each to be written into the
lane registry before launch:

- Store migrations: both rungs add a migration, so the number `0006` and the older-store helper's map
  (`cmd/dubplate-server/olderstore_test.go:28-33`) are contested. Assign 0006 to whichever lane merges first and
  have the other rebase onto it.
- `internal/serve` registration and `cmd/dubplate-server/serve.go` wiring (`:345-367`). Each lane registers
  from its own file through `Shell.Register` (`internal/serve/serve.go:84`), which PA's contested-surface
  table requires (PA:205).
- `web/`: rung 14's admin pages (question 3) against rung 7's uploader, plus the shared `web/src/lib/api/post.ts`.
- The journey vocabulary's wire projection: rung 7 owns the vocabulary, and rung 14's contract document rules
  whether enum extension is additive (`research/2026-09-04-reverse-edge-completion.md:223`).

Launching rung 7 as a second lane sits inside the standing grant of up to three code lanes. It is listed under
decisions (D1), with its cost.

## 2. Scope

### Rung 14 (recommended)

The ladder entry, BS:983-988: "The contract document, `/api/v1` under the owned prefix, the additive-within-version
policy, the authorization model (AdminEmail, header-only admin verbs enforced through provenance, service tokens for
unattended machine verbs), and non-browser token acquisition on the session vocabulary, so native clients and the
post-launch TUI have a credential path." RP:107-108 leaves the rung unchanged except its private gate (RP:146-148).

CH adds the admin spine's content at CH:729-732: "two mounts; settings, per-member overrides, and collection
endpoints the web app and the TUI share, including retire, purge, reassign, adopt, and seize; the re-authentication
gate and outbound-URL validation on sensitive settings; the audit on every write; member removal as the checklist
above; version negotiation on connect." The two mounts are CH:147-151 (cookie for the web app, bearer for the TUI).
The sensitive-write rules are CH:167-173. Member removal's checklist is CH:576-586: "epoch bump, app passwords,
device credentials, writer epoch freeze, in-flight push abort, collection reassignment, streamer deprovision."
`if_version` on collection shape edits is CH:296-298, and R6:183-184 names rung 14 its first writer. R6:197-198 adds
"reconciling `DUBPLATE_ADMIN_EMAIL` is rung 14's".

Several of these have no caller at 0.8. The bearer mount's caller is the TUI (CH:147-151), and the TUI ships with
the 0.9 local client (CLAUDE.md, "Versions"). Version negotiation is "connect" from a local store (CH:205-208).
Adopt and seize act on laptop-written collections, which also wait for 0.9. Geoff's 2026-09-12 refinement in CLAUDE.md
("nothing is exported ahead of a caller") bears on all of these. Question 2 asks where they go.

Carried items from STATUS "Open items" that land here, each with its origin:

1. The last-admin guard and criterion 6's concurrent self-demotion clause, with the removal verb. R6:205-207 routes
   it: "rung 14's removal verb ... the first verb that can remove or demote." HISTORY 6d, "Routed items".
2. The removal verb bumps the address's revocation epoch as it tombstones (6d segment B security read, Low 3:
   "sessions key on the address, so a removed and re-added member revives old sessions"). This makes the removal
   verb the first production caller of `EpochStore.Bump`. R-6d20 then applies (6d plan
   `plans/2026-09-26-rung-6d-identity-serve-store.md:833-848`): "the first of the two to write an epoch weighs
   whether it needs the address key at all", and criteria 7 and 8 route to that writer.
3. The network actor keyed on `UserID`, "since a renamed principal breaks a job's correlation" (HISTORY 6d,
   "Routed items").
4. The admin write path, with a resolving channel for an `admin` row that refuses `serve`'s start. HISTORY 6c,
   "To rung 14": "source `admin`, `seed_value` recording the seed it replaced, its channel, its audit kind; and a
   resolving channel that works while `serve` is down ... such as a CLI verb."
5. From "First migration after `0005`", if rung 14 writes a migration: R-6d4's keep limit (a pre-lock `Inspect`
   that sees the store busy or under an intent yields no facts, so a migrating start prunes to the default) and the
   older-store helper's table map, which the new migration's tables must join.

Item 5 lands only if rung 14 adds a table or column. Candidates are an integer row version on `collections` for
`if_version`, a bearer-token table if question 2 keeps it, and an audit table for member writes.

Routed BACKLOG items: none names rung 14. These sit on surfaces rung 14 edits and are cheap to take with it: #23
(`web/src/lib/api/post.ts:65` maps every 403 to the cross-site copy, which a removal refusal would hit), #25
(`internal/settings/evaluate.go:265` wraps a nil with `%w`; the admin write path touches this file), #12
(`member add --help` lacks "(required)"; a `member remove` verb lands beside it), and R-6d17's note plus #17 to #20
if rung 14 edits the members page (HISTORY 6d, "Routed items").

Forward hooks (SS:945-1040) against rung 14's surface:

| Hook | Verdict | Why |
|---|---|---|
| 15, configuration homes (SS:996-1009) | lands here | "The rung 14 admin API exposes settings and collections on two mounts"; private stays unsettable until slice 3 |
| 18, description and sensitive mark on every key (SS:1017-1019) | lands here | names rung 14; `config.Setting` has `Sensitive` but no description, and wall rows, integration classes, and thresholds carry neither (section 4) |
| 5, stable identity (SS:963-965) | lands here in part | the actor and, under R-6d20, possibly the epoch move onto `UserID` |
| 4 and 20, event-shaped tables | neutral unless an audit table is added | an added audit table carries a timestamp and a kind; it is not a serve event, so hook 20 stays neutral |
| 22, integer wire id per collection (SS:1033-1040) | neutral | "a rung that creates collections before slice 3 is neutral, because slice 3's migration backfills" |
| 7, 8, 9 (migrations, writer, transaction model) | neutral | rung 6 landed them; rung 14 uses them |
| 1 to 3, 11 to 14, 16, 17, 19, 21 | neutral | each names rung 7, 8, 9, 13, or slice 8 (RP:111) |

### Rung 12 slice 3 (for when rung 9 lands)

The ladder entry is SS:903-906. Its routed items are hook 22, both method-name forms, the CORS policy, the base-URL
decision with a BS erratum, and the fixed-width timestamp layout. The full list is the study record's "Routed"
section (`research/2026-10-06-slice-1-study-record.md`). STATUS adds UNI-H2 with the root-route set, the `;`-segment
and case-variant prefix gaps, the percent-encoding ban revisit (RP:426-428), and ledger rows L-20, L-21, and L-51.
HISTORY 2b's "What slice 3 would be wrong to rediscover" says each row "closes when its slice builds the test its
`test` field names" and leaves through the accounting test's `retiredRows` map. Fork C is ruled: no folder browsing
at 0.8 (SS:15-28), so `getIndexes` answers empty, `getMusicDirectory` answers error 70, and slice 3 decides `parent`
(SS:476-480). Slice 3 also lifts rung 14's private refusal once visibility holds on every endpoint (RP:146-148).

## 3. Prior art and dependencies

Rung 14 adopts no new library on current evidence. These entries need a recorded consult in the plan:

- `golang.org/x/time/rate` (SS prior-art row): already adopted for the auth limits. R6:838 calls a live apply of
  the four rate limits "rung 14's option"; the plan rules it in or out.
- `crypto/hmac` and `crypto/sha256` (stdlib): the session and link tokens already sign through
  `internal/auth/token.go:255-332`. Any bearer credential (question 2) reuses that signing shape or stores a hash.
  Neither needs a library.
- `github.com/getkin/kin-openapi` v0.149.0, the newest on the module proxy (2026-08-28). It is already a module
  dependency for the harness, with evidence in `research/2026-09-25-streamer-prior-art-survey.md:166` (MIT; 16
  releases in 12 months; 481 importers of `openapi3filter`). If the contract document is an OpenAPI file, a test can
  validate `/api/v1` answers against it. The no-link test (`internal/harness/conform/arch_test.go`) forbids it in
  the server's dependency set, so it may appear only in `_test.go` files.
- `modernc.org/sqlite` v1.60.1 (newest, 2026-09-29) and its standing rule that `modernc.org/libc` moves only with it.
- `github.com/wneessen/go-mail` v0.8.1 (newest): the sensitive-change notice (CH:171-172) mails the admin address
  through the existing mail path.
- The OpenSubsonic specification, Navidrome, gonic, taglib, mewkiz/flac, Picard's tables, and chromaprint: neutral
  for rung 14. The plan says so in one line.

The pass opens with the `dependency-upgrade` sweep over `go.mod`, `web/package.json`,
`tools/venv/requirements.txt`, and the pinned binaries (rclone, golangci-lint, Playwright, and libxml2 under
`brew pin`). A proxy read on 2026-10-10 shows two movers in `go.mod`: `golang.org/x/sys` v0.48.0 to v0.49.0 and
the indirect `golang.org/x/crypto` v0.57.0 to v0.58.0. `cobra` v1.10.2, `go-mail` v0.8.1, `x/time` v0.16.0,
`kin-openapi` v0.149.0, and `sqlite` v1.60.1 are current. The workstation runs go1.27.2, golangci-lint 2.14.0,
rclone v1.74.3, and libxml 21504. The sweep checks each binary for a newer release; this brief did not.

## 4. Facts the plan will rely on, verified at `6d90dcd`

Store and migrations:

- The highest migration is `0005_routed.sql`; `internal/engine/store/migrations/` holds `0001` to `0005`. `0005`
  creates `member_overrides`, `member_preferences`, `member_quota`, `group_quota`, `device_credentials`, and
  `move_intents`.
- `users` (`0003_catalog.sql:1-10`): `id BLOB` primary key, `address TEXT NOT NULL UNIQUE`, `username TEXT NOT NULL
  UNIQUE COLLATE NOCASE`, `admin INTEGER ... CHECK (admin IN (0, 1))`, `tombstoned_at INTEGER`.
- `collections` (`0003_catalog.sql:12-34`): `owner BLOB NOT NULL REFERENCES users (id)`, `visibility TEXT NOT NULL
  CHECK (visibility IN ('group', 'private'))`, `writer`, `writer_epoch`, and no row-version column. Nothing in
  `internal/` or `cmd/` inserts a collection row (grep at HEAD); R6:186-187 gives the default collection to "rung 7
  or rung 10".
- `settings` (`0004_settings.sql:1-8`): `source TEXT NOT NULL CHECK (source IN ('env-seed', 'admin'))`,
  `seed_value TEXT`, `actor TEXT NOT NULL`. `settings_audit` (`:10-19`) has `at`, `kind`, `actor`, `channel`,
  `key`, `old_value`, and `new_value`.
- `jobs` (`0002_jobs.sql:1-15`, index at `:17`): `state` in the six `ops.JobState` values, `actor TEXT NOT NULL`. A kind lease
  admits one job of a kind across processes (`internal/engine/store/lease.go:13-16`).
- The older-store helper: `migrationTables` at `cmd/dubplate-server/olderstore_test.go:28-33` maps versions 2 to 5;
  "a migration whose tables this list does not name fails the re-migration with 'table already exists'"
  (`:20-23`).

Identity and actors:

- `identity.Channel` (`internal/identity/identity.go:162-172`) has two values, `ChannelCookie` and
  `ChannelCommandLine`. No bearer channel exists.
- `identity.UserID` is `[16]byte` (`:199`). `Identity` (`:203-224`) carries `Address`, `UserID`, `Principal`,
  `DisplayName`, `Source`, and `Channel`. `Valid` refuses a zero `UserID` and `ChannelCommandLine` (`:228-231`).
- `ops.NetworkActor` (`internal/ops/ops.go:47-53`) builds the wire form from the principal:
  `Actor{wire: networkActorPrefix + ":" + id.Principal, channel: id.Channel}`. It has no non-test caller.

The ops seam and its HTTP projection:

- `ops.Seam` (`ops.go:169-189`) has four methods: `StartBackup`, `JobStatus`, `CancelJob`, and `Catalog`. Their
  privileges are in `internal/ops/catalog.go:76-81`, with `catalog` at member level and the rest at admin level.
- `opshttp` is unmounted. `internal/ops/opshttp/doc.go:17-19`: "Nothing here is exported: the package has no caller
  until a server mounts it behind the session middleware that puts a verified identity in that context."
  `newHandler(next ops.Seam, auth authorizer, channel identity.Channel, log *slog.Logger)` is at `opshttp.go:60`.
  The unexported `authorizer` interface is `authorize(ctx context.Context, op string) (ops.Actor, ops.Privilege,
  error)` (`authorize.go:19-21`). `newAuthorizingSeam` refuses the command-line channel (`authorize.go:36-39`).

Auth:

- `EpochStore.Bump(address string, now time.Time) error` (`internal/auth/epoch.go:103`) has no caller outside tests:
  `grep -rn '\.Bump(' --include=*.go cmd internal | grep -v _test` returns nothing. The record is keyed by
  normalized address (`epoch.go:48-57`), and the comment names the move to the surrogate id as pending (R34).
- `Roster.AddMember(ctx, candidate Member, actor string, channel identity.Channel, clientIP string, logger
  *slog.Logger) error` (`internal/auth/roster.go:313`). A tombstoned address is "reactivated under its old id"
  (`:298-301`, written at `:395-419`). The only admin check is in the add transaction: `SELECT EXISTS (SELECT 1 FROM
  users WHERE admin = 1 AND tombstoned_at IS NULL)`, refused with `errNoAdmin` (`:426-432`). No remove or demote
  path exists.
- `RequireAdmin(cfg AdminConfig, next http.Handler) http.Handler` (`internal/auth/admin.go:110`) answers a non-admin
  exactly as `http.NotFound` does. `NewRosterAPI` (`admin.go:78`) serves GET and HEAD of the list and POST of one
  member.

Serve:

- `Shell.Register(name string, h http.Handler)` (`internal/serve/serve.go:84`) mounts `prefix/name` and its subtree.
  It panics on a duplicate, a late registration, or a widening name.
- `originCheck.addBearerBypass(pattern string)` (`internal/serve/origin.go:92`) is unexported. `Shell` exposes no
  method for it, and its only caller is `origin_test.go:127`. The doc at `origin.go:37-50` requires the first
  consumer to "add a shell-level test that a bypassed route never reads the session cookie".
- Wiring in `cmd/dubplate-server/serve.go:345-367`: the auth routes, the probe, `roster` (`:359`), the web routes,
  and `members` behind `RequireAdmin` (`:365`). The route names are constants at `:35` and `:39`.

Config and settings:

- `config.Setting` (`internal/config/settings.go:59-80`): `Key`, `Kind`, `Default`, `Doc` ("the accepted form a
  refusal reports on a parse failure", `:67`), `Apply`, `Floor`, `Ceiling`, `Values`, and `Sensitive` (`:76`). It
  has no description field.
- `WallRow` has `Name` only (`internal/config/keys.go:7-9`). `IntegrationClass` has `Name`, `Rung`, `Flag`, and
  `Required` (`:155-160`). `Threshold` has `Key`, `Kind`, `Default`, and `Doc` (`internal/config/thresholds.go:45-50`).
- The settings package writes only `env-seed` rows: `Apply`'s doc says "Every row it writes carries the source
  'env-seed'" (`internal/settings/apply.go:28`), with SQL at `:80-88`. The two source values are constants at
  `internal/settings/evaluate.go:20-21`. No admin writer exists.
- `DUBPLATE_ADMIN_EMAIL` is a wall row (`keys.go:19`), parsed into `Config.AdminEmail`
  (`internal/config/config.go:97`, `:402`).

CLI and web:

- The cobra verbs are `check`, `backup`, `version`, `member` with `add` and `list` (`cmd/dubplate-server/member.go:40`,
  `:76`, `:119`), `restore`, `doctor`, `serve`, and `unidecode install`. No `member remove` and no settings verb exist.
- The `web/src/routes/` pages are `login` (with `link`) and `members`.

## 5. Questions for Geoff

1. **Rung 9's concurrency model: a per-album lease or an explicit queue?** PA:118 takes this before rung 14's plan,
   because `/api/v1` publishes the job types. Recommendation: a per-album lease. The store already admits jobs
   through flock leases (`lease.go:13-16`), and the album is the verdict unit. A lease per album has no global lock
   for an interactive session to hold, so timer imports cannot starve behind it (the binding constraint, organizer
   spec `specs/2026-09-01-dubplate-organizer-design.md:612-616`). The MusicBrainz limiter's caller priority, which
   rung 9 re-verifies, arbitrates the one shared budget. Alternatives: an explicit queue, which gives a visible order
   at the cost of a written starvation policy; or no ruling now. Under the third, rung 14 publishes today's job types
   on the reading that either model only adds fields, which sets aside PA:118's order.
2. **Does rung 14 at 0.8 build only what has a 0.8 caller?** Recommendation: yes. The bearer mount, non-browser token
   acquisition, service tokens, connect version negotiation, and the adopt and seize verbs move to 0.9 with the
   local client, their first caller. The cookie mount, settings, member removal, and the server-written collection
   verbs stay. This follows the 2026-09-12 refinement in CLAUDE.md. Alternative: build the whole of BS:983-988 and
   CH:729-732 now. That adds about four tasks and a second pass (section 6), and the bearer half would ship with no
   caller until 0.9.
3. **Does rung 14 carry the web pages that call its cookie mount?** Recommendation: yes for a removal control on the
   existing members page and a settings page. The collection page waits for the collection verbs (D3). Every
   endpoint then ships with its caller and is rendered and read before Geoff sees it, and the pages follow the visual
   identity spec. Alternative: API only, with the pages in lane 1's next `web/` pass. The endpoints would then sit
   uncalled, against the same refinement.

## Decisions the plan will take

These are method and architecture calls, each with its evidence.

- **D1. Plan rung 7 as a second lane.** PA:116 pairs lanes V and I, the standing grant allows three code lanes, and
  7 is on the critical path to slice 3 (section 1). Rung 7 needs its own spec (BS:683-686, "Each rung gets its own
  spec and plan") and carries Geoff's uploader before-and-after touchpoint, so it starts as a planning lane. Its code
  lane launches on its own approval, with the contested surfaces in section 1 written to the registry first.
- **D2. Write a short rung 14 spec before the plan.** BS:683-686 requires one per rung. CH and R6 already settle most
  of the design, so the spec records the contract-document form, the authorization model, and the answers to
  questions 1 to 3.
- **D3. Split rung 14 into two passes.** 14a lands now: the contract document, the `/api/v1` cookie mount, the
  actor on `UserID`, hook 18, the admin settings write path with the re-authentication gate and the resolving verb,
  member removal, and the pages. 14b follows rung 7, which writes the first collection row (R6:186-187): collection
  endpoints with `if_version`, the private refusal, and the collection page. BS:684-685 lets an M rung that outgrows
  one pass split at its own planning.
- **D4. The contract document is a hand-written OpenAPI file validated in tests by kin-openapi.** The repo already
  holds that dependency at its newest version. The OpenSubsonic layer took the hand-written route with evidence
  (ROADMAP, "Give back"), and the no-link test keeps it out of the binary.
- **D5. Key the network actor on `UserID`.** The wire becomes `member:<hex user id>`. The principal stays in log
  fields only, since `identity.go:209-212` calls `UserID` the durable key and `Principal` "never the durable key".
- **D6. The removal verb is the first epoch writer, so it settles R-6d20.** Evidence for the plan's reading: a
  re-added member keeps the old id (`roster.go:298-301`), so the address-to-id resolution never changes. The
  standalone revoke verb stays routed with slice 6, where app passwords give it a second consumer (STATUS).
- **D7. The last-admin guard is one statement inside the removal and demotion transaction,** re-read under the
  store's single writer, extending the add-path check at `roster.go:426-432`. A concurrent self-demotion then
  serializes on the writer.
- **D8. The outbound-URL validator waits for its first URL-valued key.** No registered setting holds a URL today
  (R6:193-194 records none at rung 6). Hook 18's declarations name the key that will need it.
- **D9. `DUBPLATE_ADMIN_EMAIL` reconciliation (R6:197-198).** After 6d, `users.admin` is authoritative. The plan
  rules whether the wall row becomes a seed for the first admin or a `doctor` comparison, and records the ruling as
  architecture, since both keep the box's one admin.

## 6. Proposed segmentation and ceiling (14a, under the recommended answers)

| Task | Outcome | Pass class |
|---|---|---|
| 0 | Dependency sweep recorded; x/sys and x/crypto taken; binaries checked | `tool` |
| 1 | Contract document and the `/api/v1` cookie mount over `opshttp`, with unknown API paths ending in JSON 404s | `auth-data` |
| 2 | Network actor on `UserID`; job records and logs agree | `auth-data` |
| 3 | Hook 18: every registered key carries a description and a sensitive mark, and registration refuses one without | `tool` |
| 4 | Admin settings write path (source `admin`, `seed_value`, channel, audit kind) and its `/api/v1` operations | `auth-data` |
| 5 | Re-authentication gate on sensitive writes and the admin-address notice | `auth-data` |
| 6 | CLI resolving verb for an `admin` row that refuses `serve`'s start | `tool` |
| 7 | Member removal: tombstone, epoch bump (R-6d20), last-admin guard, checklist test over what exists | `auth-data` |
| 8 | Removal and demotion over the CLI and `/api/v1`; `DUBPLATE_ADMIN_EMAIL` ruling applied | `auth-data` |
| 9 | Members-page removal control and settings page, rendered and read | `engine-logic` (web) |

The segments are A (0 to 3), B (4 to 6), and C (7 to 9), each ending on a green full gate on the heavy lane, since
`web/` runs Chromium. The close runs the simplifier, one `go-architecture-reader` per touched package (`ops`,
`opshttp`, `auth`, `settings`, `config`, `identity`, `serve`), a `web-auth-security-reviewer` read after segments B
and C (as 6d ran two), and `svelte-reviewer` plus `daisyui-a11y-reviewer` against the visual identity spec.

The token derivation uses recent spend. 2b spent about 7.4M over twelve tasks, close included, about 0.62M each.
2a spent about 6.0M at its park over nine tasks, plus its merge and close (HISTORY 2a, "Budgets"). 6d is the closest
match, since it was also an `auth-data` pass with two security reads. Its runner tallies for segments A to C, which
covered Task 0 and the eight build tasks, sum to about 5.2M, about 0.58M a task inside the chain. Its overhead
outside the chain was about 4.3M of its 9.5M: the architecture fold at about 2.59M, the C fold at about 0.26M, seven
architecture reads at about 0.63M, two security reads at about 0.29M, pre-flights at about 0.22M, and the simplifier,
Task 0's standalone agent, and the live smoke at about 0.13M each (HISTORY 6d, "Budgets"). For 14a, ten tasks at
0.6M come to about 6.0M. An overhead of about 3.0M assumes an architecture fold near half of 6d's, because 6d's
fold carried seven packages' worth of first reads. That gives about 9.0M. **Proposed ceiling 9.5M, flag 7.6M
(80%)**, with a checkpoint every four tasks. Planning spend (the spec, the plan, and `spec-plan-review`) sits
outside it, as 6d's did. If question 2 is answered "build it all", the bearer half adds about four `auth-data`
tasks: the bearer channel, token minting and acquisition, the bearer mount with its shell-level cookie test, and
service tokens. At 0.6M each plus its reads, that half costs about 3.0M and becomes its own pass under its own
ceiling.

The clock estimate comes from 2b's overnight run, which took about 2.5 hours for Tasks 9 to 12 and the close.
Segment D ran 38 minutes for two tasks, and the boundary and merge full gates ran 7.5 and 9.4 minutes (HISTORY 2b,
"Budgets"). That is about 20 minutes per task inside the chain and about 9 minutes per boundary gate. For 14a: ten
tasks at about 20 minutes, three boundary gates at about 9 minutes, and a close of about 1.5 hours with eleven
reviewer reads. **Proposed clock estimate: about 5.5 hours of execution,** excluding owner pauses and the planning
session. The `auth-data` fix rounds 6d saw would push it toward 7 hours.
