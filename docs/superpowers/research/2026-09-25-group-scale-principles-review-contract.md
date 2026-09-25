# Group-scale principles spec: contract and criteria review

Target: `docs/superpowers/specs/2026-09-25-group-scale-principles-design.md` at `1d6bc6a`
(revision 1, pre-review). Lens: is every promise testable by the rung that builds it, does
each mechanism name a fixture and a failure reason, can any criterion pass vacuously, are the
"exists" rows proven, and is each refusal durable. Findings are ranked by consequence. Polish
is out of scope. Line numbers are the spec's.

Counts: 0 blocker, 10 major, 7 minor. Two majors and one minor carry an OWNER FORK part.

The spec is sound in shape. Its principles are coherent and match the owner framing. Every
finding below is about the gap between a stated promise and what a rung plan can prove. No
finding asks for more enforcement than the framing ("easy for the intended use, hard for
nefarious ones") supports. Several folds make a promise smaller so that it becomes true.

## Evidence checked

- Rung 5 roster code and tests: `internal/auth/roster_test.go` (`TestRosterWrites_AdminInvariantAndGrant`,
  `TestAddMember_ActorAndChannelRefusals`), `internal/auth/admin_test.go` (`TestRequireAdmin_RefusalsMatchNotFound`),
  `internal/auth/handlers.go:256-263` (unknown address gets `admin_contact`), and
  `cmd/dubplate-server/member_test.go`.
- Route registration: `cmd/dubplate-server/serve.go:289-311` and `cmd/dubplate-server/web.go:165-170`.
  `TestNewServeShell_MountsTheThreeAuthRoutes` (`serve_test.go:391`) proves three routes are
  present. No test enumerates the full route set or its auth class.
- No `restore` verb exists in `cmd/`. Rung 6b-ii-a (ops seam) has landed. The restore verb is
  6b-ii (rung 6 spec, line 153).
- Streamer spec: shares at lines 289-295, `download` at 384, play history shape at 301
  ("user, song, client, timestamp, event kind, and how reported"), pre-existing library
  intake at 79, Navidrome import at 350-355, album notes at 325-326.
- `config/navidrome.toml` sets neither `EnableSharing` nor `EnableDownloads`.
- Organizer spec lines 344-348: lifecycle hooks are configured external commands at
  post-import, on-reject, on-review, and post-scan.

## Major findings

### M1. No mechanism carries an acceptance criterion

**Location:** lines 65-170 (every mechanism paragraph) and the table at lines 201-227.

**Defect.** The spec names about twenty mechanisms, and it states a boundary, fixture, or
failure reason for none of them. The repo's planning-miss rule (CLAUDE.md, 2026-09-23) needs
four things closed before a rung plan's approval: the states each check can meet, every
channel that can set a privileged field, and the fixture state each proof needs. A plan can
close those only if the spec states the contract. Right now the spec states the intent. The
limits are the sharpest case. "Refuses a new member once the roster reaches the cap" (line 74)
does not say whether the Nth add or the (N+1)th add is refused. A test written only below the
limit passes whichever reading the code takes.

**Fold.** Add an "Acceptance" section, one row per new or amended mechanism. Each row names
the boundary fixture, the observable refusal, and the mutation that makes the test fail. The
section below, "Proposed acceptance rows", is a draft of it. Rung plans then cite rows by
number rather than re-deriving them.

### M2. Three "exists" rows are not proven by any landed test or gate

**Location:** line 198 (status definition), rows at lines 205, 207, and 216.

**Defect.** The definition "exists when the property holds in a ratified spec or landed code"
merges two different states. Only row 203 is backed by landed code with tests.

- Row 205 ("Login required for all audio, art, and catalog data | Rung 5 auth; 0.8 API
  surface | exists"). dubplate serves no audio, art, or catalog today. The 0.8 surface is
  unbuilt. Through rung 12 and the escape-hatch months, Navidrome serves audio (see M10).
- Row 207 (group-visible ceiling, "rung 14 API"). Only the configuration homes spec holds it.
  The rung 14 API does not exist.
- Row 216 ("Restore offline and command-line only | Rung 6 | exists"). No `restore` verb
  exists in `cmd/dubplate-server/`. It lands at 6b-ii.

A rung plan that reads "exists" will not author the proof. So the property ships untested
at exactly the rung that builds it.

**Fold.** Split the status column four ways:

- "proven": landed code plus a named test.
- "specified": a ratified spec holds it, and no code does yet.
- "planned": a ratified spec schedules it to a rung.
- "new": this spec introduces it.

Row 203 becomes "proven" and cites `TestRosterWrites_AdminInvariantAndGrant`,
`TestRequireAdmin_RefusalsMatchNotFound`, and the unknown-address row of
`internal/auth/handlers_test.go`. It still needs the route inventory from M3 for the "no
self-signup" half, because no test proves that a signup route is absent. Rows 205, 207, and
216 become "specified" and name the rung whose acceptance proves them (0.8 slice 6, rung 14,
6b-ii).

### M3. "Nothing is reachable without a login" has no route enumeration, and the stated unauthenticated set is incomplete

**Location:** lines 85-91 and row 205.

**Defect.** Line 89 lists the unauthenticated surface as three items: the login flow, the web
app's static assets, and `getOpenSubsonicExtensions`. The serve shell today also mounts the
branding, fonts, and operator CSS routes without a login (`web.go:166-170`). It also forwards
unowned paths to a fallthrough, which is Navidrome from rung 12. Arguably those three routes
count as "static assets", but the claim that none of them serves art depends on reading them
that way. No test enumerates the registered routes. So a later rung can add a route that
bypasses the auth wrapper, such as a share path, a signup page, or a debug handler, and every
existing test stays green. This is the vacuous-pass case the lens asks about. The property is
checked only by the absence of a failing test.

**Fold.** This one has a single clearly correct answer. Add a route-inventory test in
`cmd/dubplate-server`. The test builds the real shell, walks every registered subtree name,
and compares each route against a checked-in table of (name, auth class). The auth classes
are public, session, and admin. The test sends an unauthenticated request to each route and
asserts the refusal for every non-public row. A new route missing from the table fails the
test. At 0.8, the table extends to the OpenSubsonic endpoint set: every endpoint except
`getOpenSubsonicExtensions` answers error 10 or 40 without credentials. The spec then states
the unauthenticated set as that table, not as prose. Land the test at the next rung that
registers a route, and name that rung in the table.

### M4. The download erratum leaves folder-view ids open, which reopens the artist-zip path

**Location:** lines 126-129 and erratum 2 (lines 246-249).

**Defect.** The erratum says `download` returns "one track or one album". The streamer
implements folder view (`getIndexes`, `getMusicDirectory`, `getMusicFolders`, with gonic as
the oracle, streamer spec line 538). In Subsonic semantics, `download` on a directory id
returns a zip of that directory. A folder-view directory id at the artist level, or a music
folder id, is an artist or library zip under another name. A test that refuses only an ID3
artist id and a playlist id passes while the bypass stays open.

**Fold.** Enumerate the id kinds the catalog mints. `download` accepts a song id and an ID3
album id. It answers error 70 for every other kind: ID3 artist, playlist, folder-view
directory at any depth, music folder, and share. The acceptance fixture is a table test over
one id of each kind the catalog mints. The test fails if any kind outside the two allowed
kinds returns a body. If the owner wants an album-level folder directory downloadable, that
row maps it to its album explicitly.

### M5. "No bulk export" and "download one album at a time" promise more than the mechanism delivers

**Location:** lines 43-45 (is/isn't list, written for website reuse), line 126 (heading), and
ruling 3 (lines 272-273).

**Defect.** The album limit counts album zips only. Single-track `download` and `stream` are
uncapped. Offline-sync clients such as Symfonium routinely pull a whole library track by
track. So "Members stream one track at a time and download one album at a time, and
whole-library copies are backups" is false as a statement about what a member can do. Any
test of the heading "No bulk export over the API" passes vacuously: refusing an artist zip
proves nothing about bulk export. The line is written for a public website, so the false
claim would be published.

**Fold.** The wording part has one correct answer. Reword the promise to the mechanism: "The
API has no one-request export larger than an album. Album downloads are rate-limited, and the
admin sees each member's volume." Drop "whole-library copies are backups" as a capability
claim, or scope it to the operator's backup.

**OWNER FORK: whether per-track volume gets a bound.**

- (a) Visibility only. Keep the volume view, with no track cap.
- (b) Visibility plus a notice. The admin gets a notice when a member's bytes over a window
  pass a setting.
- (c) A hard per-member daily byte or track ceiling on `download` and untranscoded `stream`.

Recommendation: (b). A hard cap (c) breaks the offline caches a group member legitimately
fills, which works against the intended use. A notice fits "make unusual volume visible to the
admin" (line 63). The notice is testable: a fixture above and below the threshold, asserting
the notice fires once. Without it, the view is visible only if someone looks.

### M6. The member cap has no single enforcement point across its channels and states

**Location:** lines 74-83 and row 204.

**Defect.** The cap is a settings-table value in the rung 14 store. The roster is a file
written by `dubplate-server member add`, which runs without `serve`. Rung 5's tests cover a
hand-edited roster (`TestMemberAddCmd_NoAdminRefusedOnHandEditedRoster`), so the file is also
an accepted channel. The spec says "the server refuses a new member", which leaves three
things open:

- Channels. Does the cap bind the command-line verb, the web add page, `dubplate import
  navidrome`, and a hand edit? Sensitive re-authentication (line 76) is a magic-link step
  that exists only on the web channel. How a cap is raised from the command line is unstated.
- States. The spec names none of these:
  - a roster already at or over the cap at load, after a hand edit or a lowered cap;
  - a cap set below the current count;
  - whether admins count;
  - whether tombstoned members count.
- Import. The spec does not say in what order import admits users up to the cap, or what
  happens to a skipped user's stars, playlists, and play history.

**Fold.** Name one enforcement point: the roster's add function that every write channel
already calls (the rung 5 `AddMember` path). Channels that cannot reach the store must fail
closed, not skip the check. State the following:

- A roster over the cap at load is accepted and raises a doctor row. It is never truncated.
- Lowering the cap refuses new adds only.
- The count is active roster entries, admins included.
- The import admits in Navidrome creation order, admin first. It reports each skipped user
  along with that user's unimported stars, playlists, and history.
- Raising the cap from the command line is audited under the command-line channel and
  notifies the admin, with no re-authentication, because shell access is the stronger
  credential.

Acceptance: at cap N, the Nth add succeeds and the (N+1)th is refused on each of the command
line, the web, and the import, and the refusal names the key. An over-cap roster loads.

### M7. The volume view promises measures that no event records

**Location:** lines 153-158 and rows 221-222.

**Defect.** The view shows streams, album downloads, and bytes, per member and per device.
Play history's shape is "user, song, client, timestamp, event kind, and how reported"
(streamer spec line 301). `client` is the client-name parameter, not the device credential.
Two devices running the same app are indistinguishable, and a credential shared across
devices looks like one device. No event carries bytes. "Download events of the same shape"
cannot represent an album zip, because the shape is keyed by song. As written, a per-device
bytes test has no data to read. It would either fail or be written against a fake that
proves nothing.

**Fold.** State the following:

- The event carries the authenticating credential id (app password or device credential),
  beside `client`.
- Download events carry the entity kind (song or album) and id.
- Bytes are the bytes written per response, including 206 Range responses and aborted
  streams. Otherwise drop "bytes" from the view.

Add the counts-only guarantee as a checkable contract. The admin volume API response decodes
into a struct under `DisallowUnknownFields` that has no entity-id field. A test with a
populated fixture (at least two members, two devices, one aborted Range stream) asserts the
per-device split and byte totals. The fixture must be populated so the test cannot pass on
zero rows. Record the credential id in play history as a forward hook in the streamer spec's
list, because the rung 6 or rung 9 schema is where it is cheap.

### M8. The source field names no write channels, no server-side validation, and no forward hook

**Location:** lines 139-145 and rows 217-218.

**Defect.** The source answer is a privileged field on the album row. Under the repo rule,
each privileged field names every channel that may set it. The spec leaves these open:

- Who answers for a mirrored collection: "records one answer for the collection when the
  admin approves its first mirror" does not say whether the admin or the member answers.
- Whether the answer can be edited later, and by whom.
- What rung 16 migrated albums carry. `added_by` is covered ("unknown"). `source` is not.
- What a command-line bulk intake of a stranger's pre-existing library carries (streamer spec
  line 79: "a pre-existing library enters through the pipeline").
- Whether a required answer (ruling 2) is enforced by the server. A tus client can omit
  upload metadata, so a UI-only requirement can be bypassed.

The source column is also not in the streamer spec's forward-hooks list. Hook 17 carries only
`added_by`. So the rung 7 and rung 9 plans, which cite that list, have no pointer to it.

**Fold.** Add a channel table for `source` with these rows: the uploader (member, from the
submission), the mirror (the member at first-mirror request, confirmed at admin approval),
command-line intake (operator-supplied, default `other`), and rung 16 (`unknown`). State that
only an admin edits the value after admission. If ruling 2 rules "required", the submission
endpoint refuses a missing or unknown value with the reject-wording contract, and a test
posts a submission without it. Add a forward hook beside 17: "the submission's source answer,
rung 7, carried to the album row at rung 9."

### M9. The refusals are prose only, and the contributor guide has no landing rung

**Location:** lines 117-120, row 208 (federation), row 213 (fetching), and row 227
("Contributor guide (does not exist yet) | new").

**Defect.** The refusals are no search-and-download, no torrent, Usenet, or Soulseek, no fetch
automation, no federation, and no public share. Each lands only in "build spec principles" and
an unscheduled contributor guide. Most of this repo's code is written by agents. Under the
workstation rule "a rule lives where it executes", a refusal in a spec that no implementer
dispatch reads reaches no implementer. Nothing makes the refusal durable against an agent or a
contributor who never opened this spec. The build spec edit is also not listed among the
errata (lines 229-252), even though the table routes the refusals there.

**Fold.** This part has one correct answer:

- Land a short "Refused features" block in the project CLAUDE.md in the same fold. Implementer
  dispatches and plan reviewers read that file.
- Add the build spec principles edit to the errata list.
- Give the contributor guide a landing point: the packaging initiative, or rung 17 if the repo
  is published at launch.
- The route inventory (M3) makes the public-path refusal mechanical.

**OWNER FORK: whether the acquisition refusal also gets a gate leg.**

- (a) Prose in CLAUDE.md and the contributor guide only.
- (b) A subprocess allowlist test: every `exec.CommandContext` target in shipped packages is
  in a fixed set (ffmpeg, rclone, fpcalc, and similar), so adding `yt-dlp`, `aria2c`, or an
  `slskd` client fails the gate.
- (c) Option (b) plus a `go.mod` denylist of torrent and Soulseek modules.

Recommendation: (b). It is cheap because the subprocess set is small (`internal/backup`, the
harness oracle is test-only). It also catches the most likely acquisition route, which is
shelling out to a downloader. A module denylist (c) is a list that never ends, and it misses
the next library.

### M10. The principles do not bind the Navidrome-backed period

**Location:** lines 85-91 and 126-129, and rows 205, 206, and 214.

**Defect.** From rung 12 to 0.8, and through "0.8's first months" as the escape hatch (ROADMAP
line 97), members reach Navidrome behind dubplate's sign-on. Navidrome's `download` returns
artist and playlist zips when downloads are enabled, which is Navidrome's default. Its share
feature, with a public `/share/` path, is off by default but not pinned: `config/navidrome.toml`
sets neither key. So every "no public path" and "no artist zip" row is true of the 0.8 server
and silent about the server the group actually uses at launch.

**Fold.** The sharing half has one correct answer. Pin `EnableSharing = false` in
`config/navidrome.toml`. Add a gate leg in `scripts/check.sh` that asserts the key is present
and false. Add a row to the table naming it.

**OWNER FORK: Navidrome downloads before 0.8.**

- (a) Accept Navidrome's download scope until 0.8, and say so in the spec.
- (b) Set `EnableDownloads = false`. Members keep streaming and client offline caching, and
  lose zips entirely.

Recommendation: (a). The window is bounded, the group is the trusted founding group, and (b)
removes album downloads that the principles explicitly allow. State the accepted window so the
table stays honest.

## Minor findings

### m1. The share erratum's "unless" leaves create and update undefined

**Location:** lines 241-245.

**Defect.** "Answer error code 0 ... unless the study rung finds a client that requires an
empty success" cannot apply to `createShare`, which has no meaningful empty success. The
erratum is also silent on three things:

- `getUser`'s `shareRole`. Clients hide share UI based on it, and the streamer derives role
  flags from the roster role (streamer spec line 287).
- The extensions list.
- The web app's share pages (streamer spec lines 42 and 395).

**Fold.** `createShare`, `updateShare`, and `deleteShare` always answer error 0. Only
`getShares` may answer an empty list if the study rung requires it. `shareRole` is false for
every user. No share route or page is mounted, which the M3 inventory proves.

### m2. The album download limit has no refusal shape, window source, or boundary fixture

**Location:** lines 126-129 and ruling 3.

**Defect.** The spec leaves four things unstated:

- The refusal an OpenSubsonic client sees. Subsonic has no rate-limit code, and clients
  expect HTTP 200 with an error element.
- Whether the rolling 24 hours survives a restart. An in-memory `x/time/rate` bucket, which
  the streamer already adopts for auth, resets on restart.
- Whether a per-member override is a sensitive write.
- The at-limit behavior.

**Fold.** Count the member's album download events in the store over the trailing 24 hours,
so a restart does not reset the window. The refusal is error 0 with a message naming the
limit and the time the next slot frees. The per-member override is sensitive. Acceptance: 20
events at 23h59m ago refuses the 21st. The same 20 events at 24h01m ago admit it. A restart
between the two does not change the answer.

### m3. The device cap's counted set and edge states are unstated

**Location:** lines 101-110, rows 210-211, and ruling 3.

**Defect.** Three things are open:

- Whether modern-only (hashed) app passwords, 0.9 device credentials, and web sessions share
  one count. The spec says 0.9 credentials do. It says nothing about web sessions or the
  modern-only kind.
- What happens to a member already over a lowered cap.
- What "revoke in one step" must prove, since a revoke that does not end the next request
  proves nothing.

**Fold.** The count is app passwords of both kinds plus device credentials, and web sessions
are excluded. A member over a lowered cap keeps existing devices and cannot mint. Acceptance
uses a mixed-kind fixture at the cap, and the revoked credential's next request answers error
40. The fixture must include an app password minted before the revoke, so the test cannot
pass on a never-valid credential.

### m4. The album note's writers contradict the streamer spec

**Location:** lines 149-150.

**Defect.** This spec lets "the member who added an album" write the note. The streamer spec
(line 325) says the `added_by` member or an admin.

**Fold.** Use the streamer spec's channel list verbatim.

### m5. Buy links render third-party URLs with no type allowlist or URL check

**Location:** lines 162-165 and ruling 5.

**Defect.** MusicBrainz URL relationships include purchase, streaming, free-download, and
many non-commerce types. The spec does not say which types render, and it does not validate
the URL scheme of data that arrives from outside.

**Fold.** Name an allowlist of relationship type ids: purchase for download, purchase for
mail-order, streaming, and free streaming, plus the artist's own site if wanted. Render only
https URLs. The acceptance fixture carries one of each allowed type, one disallowed type, and
one `javascript:` or http URL. It asserts that exactly the allowed https links render.

### m6. The entry-path list omits a stranger's pre-existing library

**Location:** lines 114-117.

**Defect.** The streamer spec routes a pre-existing library through intake (line 79). That
path is an operator-run bulk intake, and it is how every stranger's first library arrives. It
is neither the uploader, the mirror, nor rung 16, so the list's "only" is incomplete. It also
leaves the `added_by` value for that path undefined.

**Fold.** Add the path to the list. Name its `added_by` value: the admin who ran it, or
"unknown". Pick one in the fold. The admin who ran it matches "intake sets it from the actor".

### m7. OWNER FORK: the wishlist is a ready-made feed for fetch automation

**Location:** lines 167-170 and row 224.

**Defect.** A machine-readable list of wanted albums is the input a Lidarr-style fetcher
consumes. The pieces already exist: an external script reads the wishlist over the documented
API, fetches the album, and uploads it with a member's credential. That composes the refused
automation from shipped parts. An on-reject or on-review lifecycle hook (organizer spec lines
344-348) can serve as the trigger. The feature is unscheduled, so this is minor now.

**Options:**

- (a) The wishlist is a web-page feature only. It is absent from the documented API and emits
  no lifecycle hook event.
- (b) Expose it on the API like any other group feature.
- (c) Drop the wishlist half and keep only Requests.

Recommendation: (a). It keeps the social purpose and removes the documented integration
point. A determined operator can still scrape the page, and that is within the spec's stated
limits. Record the choice under "Declined mechanisms", or as a constraint on the feature's
future spec.

## Proposed acceptance rows (the M1 fold, draft)

Each row states the boundary and the failure the test must be able to show.

| # | Mechanism | Fixture | Passes when | Fails (proves non-vacuous) when |
|---|---|---|---|---|
| A1 | Member cap | Roster at N-1, cap N, on each of the command line, web, and import | Nth add lands; (N+1)th refused naming the key; over-cap roster loads with a doctor row | The check is removed from any one channel |
| A2 | Route inventory | Real shell, every registered subtree | Every route matches its table row and auth class; unauthenticated requests refused on non-public rows | A route is added or an auth wrapper dropped |
| A3 | OpenSubsonic auth | Every implemented endpoint, no credentials | Only `getOpenSubsonicExtensions` answers; all others error 10 or 40 | Any endpoint answers data |
| A4 | Shares absent | The four share endpoints; `getUser` | create, update, and delete error 0; `shareRole` false; no share route | Any returns success |
| A5 | Download scope | One id of every kind the catalog mints | Song and album answer; the rest error 70 | Any other kind returns a body |
| A6 | Album download limit | 20 events at 23h59m and at 24h01m, restart between | 21st refused and admitted respectively | The window resets on restart or counts wrong |
| A7 | Device cap and revoke | Mixed-kind credentials at the cap; one live credential revoked | Mint refused at the cap; revoked credential gets error 40 next request | A kind escapes the count, or the revoke lags |
| A8 | Volume view | Two members, two devices, one aborted Range stream | Per-device counts and bytes exact; response has no entity id | Shared-client devices merge, or an entity id appears |
| A9 | Source answer | Submission without source (if ruling 2 is required) | Refused with reject wording; stored value per channel table | Missing source admitted |
| A10 | Restore not over HTTP | Route inventory plus the ops HTTP projection | No route or ops verb reaches restore | One does |
| A11 | Backup contents never returned | HTTP backup under the admin authorizer | Response carries a handle and status only | Any snapshot bytes are in the body |
| A12 | Navidrome sharing pinned | `config/navidrome.toml` | `EnableSharing = false` present | Key absent or true |
| A13 | Buy links | One allowed, one disallowed, one unsafe URL | Only the allowed https link renders | The others render |
