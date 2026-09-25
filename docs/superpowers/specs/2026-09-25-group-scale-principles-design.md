# What dubplate is for, and what it isn't

Design spec. Revision 2, 2026-09-25. Status: **awaiting Geoff's read.** Revision 1 (`1d6bc6a`)
went through four review lenses (contract, mechanics, misuse, consistency) and an early prose
read. The fold record is `docs/superpowers/research/2026-09-25-group-scale-principles-fold.md`.

Owner intent (Geoff, 2026-09-25): "I would like dubplate to be a tool for groups of family and
friends to share purchased music. I don't want it to be an easy way for stolen music to be
mass-distributed." Owner framing, the same day: "I realize that I can't totally stop piracy. My
goal is simply for dubplate to be easy to use for its intended purpose, and hard to use for
nefarious ones. And it's a bonus if it helps induce people to buy digital and physical music
that supports artists."

That framing is the test every mechanism below passes. A mechanism that adds friction for an
honest member and none for an abuser fails it, and this revision drops the one that did. The
mechanisms are this spec's decisions, except the questions listed under "Rulings for Geoff".

dubplate is a music server for a small group of people who know each other, such as a family
or a few close friends. Each member adds music they bought or own. dubplate tags it, organizes
it into one shared library, and streams that library to every member's phone and computer. An
admin adds each member, and nothing plays without a login. dubplate can't get music from the
internet, so the library holds only what members bring. dubplate is self-hosted, so these
choices shape the product the project ships. They aren't an enforcement system, and the
operator of a server can change the code.

## Versions this spec speaks to

- **The launch (0.x).** The greenfield ladder ends in a 0.x pre-release that the group uses.
  Navidrome still serves the audio in this period.
- **0.8.** dubplate's own streamer replaces Navidrome. Most mechanisms below land here.
- **0.9.** A local client paired with the server. Its scope isn't specified yet.
- **1.0.** Polish and more audio formats.

## dubplate is, and dubplate isn't

This section is written so the website can use it nearly as it stands. Every line is true of
the product as decided. Lines about later releases say so.

**dubplate is:**

- A shared music library for a group of people who know each other, such as a family or a
  few close friends.
- A place to pool music you bought or own: lossless (FLAC) downloads from stores such as
  Bandcamp or Qobuz, rips of your own CDs and records, and music given to you as a gift, such
  as a CD or a gifted download.
- A private server. An admin adds each member, and everything on it needs a login.
- A tidy library. dubplate looks up every album in MusicBrainz, the open music encyclopedia,
  and files it in a consistent layout.
- A record of who added each album and where it came from.
- Built to point you toward buying music. Album pages will link to places to buy each album.

**dubplate isn't:**

- A way to get music from the internet. It has no search-and-download feature, and it
  connects to no torrent, Usenet, or Soulseek network.
- A public music site. It has no public links, no listening without a login, and no way to let
  anyone outside the group in.
- A network of servers. One dubplate server doesn't connect to another.
- A way to hand a library to strangers. From 0.8, dubplate has no download of an artist's
  catalog, a playlist, or the whole library in one piece. Members' devices can keep music for
  offline listening, and the admin can see how much each member's devices take.
- Legal advice. Whether you may share music with your household or friends depends on your
  country and on the terms of the store you bought it from.

## What these principles can and cannot do

dubplate runs on hardware its operator controls, and its source carries an MIT license.
Anyone with the source can remove any limit this spec describes. The limits do two narrower
things. They set the shape of the server a stranger installs, so the default serves a small
group and nothing larger. They also record what the project refuses to build, so a feature
request that would turn dubplate into a distribution tool has a written answer.

Each limit makes the intended use the easy use. Where the intended use and misuse look the
same on the wire, as a phone filling its offline cache and a script copying the library do,
dubplate doesn't refuse either. It counts both and shows the admin.

## The group stays small and closed

**Membership is invite-only.** An admin adds each member. There is no self-signup. Rung 5
built this: the `dubplate-server member` command-line verb and an admin-only add-member page
write the roster, and the web never grants the admin flag
(`2026-09-08-rung-5-identity-auth.md`, "Fork 1" and "The roster"). From rung 6d the roster
lives in the store, and the roster file becomes a one-time import source
(`2026-09-23-rung-6-ops-seam-store.md`, ruling 1). The login page answers every address the
same way and names the admin contact. An address that isn't on the roster gets no mail and no
account (rung 5, fork 4).

**A member cap bounds the roster.** The cap is a sensitive setting in the store's settings
table, so a web change needs fresh re-authentication and notifies the admin address
(`2026-09-11-configuration-homes-design.md`, "Two surfaces, one API"). The operator can raise
it to any number. It exists so that growing past a group is a deliberate act, never a default.
The default number is ruling 4.

- The count is the roster's active members, admins included. An offboarded member doesn't
  count. Rung 5 offboards by bumping a valid-after epoch and keeps the roster row, so the rung
  that enforces the cap adds an offboarded state to the row if the store has none.
- One function enforces the cap: the roster's add path, which every write channel calls. The
  channels are the `member add` verb, the add-member page, `member import`, and
  `dubplate import navidrome`. A channel that can't reach the store fails closed.
- The first-admin bootstrap (`member add --admin` on an empty roster) is never refused.
- The Nth member is admitted and the next is refused. The refusal names both ways out:
  offboard someone who has left, or raise the setting.
- A roster already over the cap loads unchanged and raises a `doctor` row. Lowering the cap
  refuses new adds only.
- Raising the cap from the command line is audited under the command-line channel and
  notifies the admin, with no re-authentication, because shell access is the stronger
  credential.
- `dubplate import navidrome` counts the members it would add before it writes anything. If
  the total passes the cap, it stops, names the setting, and writes nothing. The operator
  raises the cap or passes an explicit list of members to bring. A partial import would orphan
  the skipped members' stars, playlists, and history.

**Nothing is reachable without a login.** dubplate has no public share link, no anonymous
stream, and no public collection. A collection is either visible to the group or private, and
the configuration homes spec declines sharing outside the roster (Geoff, 2026-09-11), because
the roster is the trust boundary. The unauthenticated surface is a checked-in table, not
prose: a route-inventory test builds the real serve shell, compares every registered route
against the table's auth class (public, session, or admin), and fails on any route the table
doesn't name. Today the public rows are the login flow, the web app's static assets, and the
branding, font, and operator CSS routes. `dubplate init`'s first-run page on an unconfigured
server joins at 0.8. At 0.8 the table extends to every OpenSubsonic endpoint, and only
`getOpenSubsonicExtensions` answers without credentials. None of the public rows serves audio,
art, or catalog data.

The OpenSubsonic API calls a playlist "public" when every user on the server can see it. On
dubplate every user is a member, so a public playlist is visible to the group and no further.

**Servers don't federate.** A dubplate server exchanges no catalog, audio, or membership with
another dubplate server. The configuration homes spec's attach verbs connect a member's own
laptop to a server, under a device credential bound to that member. A member who holds
profiles for two servers can still copy albums out of one and contribute them to the other by
hand. That relay is bounded by the checkout rule (ruling 2), by intake's per-album checks on
the receiving server, and by that server's admin.

**Every credential belongs to a member.** The 0.8 streamer has each member mint a named app
password for each device from a device page, shown once and revocable alone
(`2026-09-07-streamer-2-0-design.md`, "Identity: app passwords minted from the roster
account"). The configuration homes spec adds device credentials for the local client, one per
connection profile, already listed on the member's own page with their last-seen address and
a one-step revoke ("Profiles in the TUI"). This spec adds three things.

- **A credential cap per member.** The count covers app passwords of both kinds (recoverable
  and hash-only) and device credentials. Web sessions don't count. The cap counts credentials,
  not physical devices, because one credential can be pasted into several devices. At the cap,
  the device page refuses a new credential. The refusal lists the member's credentials by last
  use, marks any unused for 90 days first, and offers a one-step revoke on each. A member
  already over a lowered cap keeps every credential and can't mint another. The refusal is a
  web page and never reaches a Subsonic client. The default cap is ruling 4.
- **An admin credential list** across all members. Each row shows the credential's name,
  owner, kind, creation time, last use, and the number of distinct addresses it was used from
  in the last seven days. A credential pasted into several people's phones shows up there. The
  number is a prompt for the admin to ask, never a verdict, because a phone's address changes
  often.
- **A one-step admin revoke** of any one credential, beside the roster-level revoke that the
  valid-after epoch already provides. A revoked credential's next request answers error 40.

## dubplate never fetches music

Music enters dubplate only from files a member or the operator already holds. dubplate has no
search-and-download feature, no integration with torrent, Usenet, or Soulseek, and no
automation that fetches missing albums, and the project doesn't accept a contribution that
adds one. This is a project policy. The documented API and the lifecycle hooks can carry files
from any script that holds a credential, and a determined operator can point one at a
download client. The project ships no hook recipe, example, or integration that fetches music.

A server that can fill its own library distributes music whatever its member count. A server
that holds only what its members bring can't.

Every channel that adds albums re-validates at the server, and none of them fetches anything.

| Channel | Who uses it | Per-album intake verdict | `added_by` | Source answer |
|---|---|---|---|---|
| Uploader (rung 7) | A member, signed in | Yes | The session's member | The member, per upload, overridable per album |
| TUI contribute | A member, by device credential | Yes | The credential's member | The member, per push |
| Mirror | A member, by device credential | No; the admin approves each member's first mirror, and quota bounds the rest | The credential's member, set by the server | The member, when the mirrored collection is created |
| Timer path and operator bulk intake | The operator | Yes | Null, shown as "unknown" | Unrecorded unless the operator supplies one |
| Copy between collections | A member or admin | No; the quota check runs in the copy transaction | Carried from the source row | Carried from the source row |
| Hand-over and seize | The member or admin | No | Preserved, never reset to the admin | Preserved |
| Rung 16 migration | The operator | Through rung 16's tooling | Null, shown as "unknown" | Unrecorded |
| Restore | The operator, offline | No | As backed up | As backed up |

No client field, API parameter, or admin form writes `added_by`. After admission, only an
admin edits the source answer.

## Downloads and offline listening

**No single request returns more than an album.** From 0.8, `download` accepts a song id or an
album id. A folder-view directory that holds exactly one album counts as that album. Every
other id kind answers error 70: an artist, a playlist, a folder-view directory above the album
level, and a music folder. The streamer spec currently lets `download` return artist and
playlist zips. This spec narrows it, as listed under "Conflicts with existing specs". An
artist zip can span most of a library, and a playlist can hold any selection.

**Offline listening is intended use.** The group listens on phones, and phones keep music for
offline play. Stock clients do this one track at a time. Tempus fetches each track through
`download` or `stream`, Symfonium through `download`, and DSub through `stream`. Symfonium's
automatic offline cache copies every track in the library after each sync. A per-track fetch for an offline cache can't be
told apart from a script copying the library. So dubplate refuses neither, and it counts both.
Whether a member's own devices holding the whole library is intended use is ruling 1.

**A meter counts everything served.** Every `stream` and `download` response is counted
against the member and the credential that authenticated it, in bytes sent and distinct tracks,
over a rolling window. Revision 1's per-member album download limit is dropped. It counted
album zips, which no offline-sync client uses, so it would have stopped almost no one and
landed only on an honest member.

- The meter is store-backed, so a restart doesn't reset its window.
- Bytes are the bytes written per response, including a partial (206) response and an aborted
  stream.
- A stream start is counted once per credential, track, and window. A Range continuation adds
  bytes only.
- The credential id comes from authentication, never from the client's `c` parameter, which a
  client asserts and two phones running one app share.

**A notice tells the admin.** When a member's distinct tracks served in a rolling 24 hours
pass a threshold, the admin gets one email naming the member and the credentials that crossed
it, through the same notify path the configuration homes spec uses for sensitive changes.
Nothing is refused. The threshold is a setting whose default sits well above a heavy honest
day. The default is ruling 4. The notice fires once per crossing, not once per request.

**The admin sees volume.** The admin's volume view shows bytes and distinct tracks served per
member and per credential. Whether it names members, and which measures it shows, is ruling 5.
The view shows counts and names no track or album. Play history, which the streamer keeps for
recently played and most played, does record songs per member. An admin with shell or backup
access can read it. Each member's own page says so in plain words.

**Backup and restore stay with the operator.** Restore is a command-line verb that runs
offline under the store's exclusive lock, and no HTTP route reaches it
(`2026-09-23-rung-6-ops-seam-store.md`, the field-channel table). A backup runs from the
command line, or over HTTP only behind rung 14's admin authorizer, and it writes to the remote
the operator configured. No response returns a backup's contents to the caller.

**Checkout follows ruling 1.** The configuration homes spec's checkout verb takes a read-only
offline copy of a server collection onto a member's laptop, and the group library is one such
collection. Checkout is served bytes under a device credential, so the meter counts it like any
other. Which collections a member may check out is ruling 2. The local client that carries
checkout isn't scheduled yet.

## Before 0.8

From launch until 0.8 replaces it, and afterward whenever Navidrome is the escape hatch,
Navidrome serves the audio. dubplate's sign-on gates the web app, and `/rest/*` authenticates
through Navidrome's own passwords. The principles hold in this period as follows.

- **Sharing is pinned off.** Navidrome's public share links are off by default, and nothing
  keeps them off today. The Navidrome configuration sets `EnableSharing = false` explicitly,
  with a gate leg that fails if the key is absent or true. Rung 12's driver carries the pin in
  its provisioning and proves it with a `doctor` row. The build spec's `/share/*` bypass is
  inert while the pin holds.
- **Downloads are ruling 3.** Navidrome's `download` returns artist and playlist zips while
  downloads are enabled, which is its default.
- **Credentials and volume aren't covered.** Navidrome has one Subsonic password per user, no
  credential cap, and no meter. These arrive at 0.8.

## Money points toward buying

**Album pages carry buy links.** MusicBrainz records links from a release to the places that
sell it. dubplate captures them for each admitted release and shows them on the album's page
in the web app. The web app has album pages from rung 7
(`2026-09-24-visual-identity-design.md`, "Open for rung 7"), so the release that first renders
buy links is ruling 7.

- **Where the links come from, in order.** The admitted release's own links come first. Most
  physical releases carry none, because stores link the digital release. A sampled 3 of 28
  releases carried a buy or stream link, while every credited artist carried one. So the next
  source is sibling releases in the same release group, through one browse call under the
  shared MusicBrainz rate limiter. The last source is the artist's own store and Bandcamp
  links, labeled as the artist's rather than this album's.
- **Which links render.** Only purchase types render: "purchase for download", "purchase for
  mail-order", and the artist's own store or Bandcamp page. Streaming-service links don't
  render, because the section's purpose is buying. Every link must be `https`.
- **What order they appear in.** The artist's own store and Bandcamp come first, then physical
  copies by mail order, then other download stores. The artist-direct store pays the artist
  the most.
- **Staleness.** Links are captured at admission and can go stale. The pass that builds the
  feature sets the refresh cadence.

**The source shows beside who added the album.** The album page's inscription names who added
the album and when (visual identity spec). It also names the source, such as "from Ana's
vinyl". From 0.8, `getAlbumInfo2`'s `notes` field carries the same line under the member's
note. It shows that the group owns the music, without preaching.

**The client era gets a MusicBrainz link.** From 0.8, `getAlbum` and `getAlbumInfo2` fill
`musicBrainzId`. Several clients render it as a link to the MusicBrainz release page, which
lists that release's stores. dubplate already knows every admitted album's release id.

**Scrobbling is planned.** ListenBrainz and Last.fm scrobbling are on the streamer spec's list
of work after 0.8 (its "2.1 list", not yet re-keyed to the new version numbers).

**A wishlist is an idea, not a feature.** A group wishlist of albums someone should buy would
merge with the `Requests` idea in `IDEAS.md`, where a member asks whether anyone has an album
and intake tells the asker when a match arrives. It isn't scheduled. The merged entry belongs
in `IDEAS.md`, with a purchase claim ("I'm buying this") and the album's buy links. One
constraint travels with it: the wishlist is a web-app feature, absent from the documented
client API, and it emits no lifecycle-hook event. A machine-readable want list is the input a
fetcher consumes. A local record store link, an optional group setting shown beside buy links,
goes to `IDEAS.md` the same way.

## The source question

Rung 7's uploader asks where the music came from. The options are:

- Bought as a download (with an optional store name or link)
- Ripped from my CD
- Ripped from my vinyl or cassette
- Free or name-your-price download
- My own recording
- Gift (a CD, a record, or a gifted download)
- Other (with an optional note)

The uploader asks once per upload, and each album in the batch can override the answer on the
review list the uploader already shows. A mirror records the member's answer when each mirrored
collection is created, as a default a later answer for one album can replace. dubplate doesn't
verify the answer. The question sets a norm. It tells every member, at the moment they add
music, that the group expects music someone paid for or owns.

The answer is an intake API field, not a page-only control, so every intake client carries it.
If ruling 6 makes it required, intake refuses a submission without a known value, in the
reject-wording contract, and the uploader page only renders the question. The submission's
answer is a forward hook beside hook 17, captured at rung 7 and carried to the album row at
rung 9.

## Declined mechanisms

**A hard download limit.** Stock clients fill offline caches one track at a time, so a limit on
album downloads binds almost no one. A limit on per-track fetches breaks the offline listening
the product exists for, and Symfonium's documented response is to retry or to switch endpoints.
The meter and the admin notice make a bulk copy visible within a day at no cost to an honest
member.

**Per-member audio watermarking.** Marking each member's stream with an inaudible identifier
moves every stream onto the transcode path. A further transcode can weaken or remove the mark.
It also tells a family or a group of friends that the server distrusts them. The `added_by`
record and the meter answer the same question at less cost.

**An ethical-use license.** A license that forbids distribution uses isn't an open source
license under the Open Source Definition. Self-hosters read an unusual license as legal risk.
It can't be enforced against a fork that ignores it, and that fork is the only party it would
target. The repository keeps its MIT license.

**A gate leg against downloader subprocesses.** A test could fail the gate when shipped code
runs a program outside a fixed list. The refusal targets a deliberate feature, which review and
the project's written refusals already catch. The test would add upkeep for every honest
subprocess, and a fork deletes it in one line.

## A legal note

Whether members of a household or a group of friends may share purchased music depends on the
country they live in and on each store's terms of sale, and those terms differ from store to
store. The operator of a dubplate server is responsible for the music it hosts and for who can
reach it. Nothing in this spec, the README, or the website is legal advice.

## Where each principle lands

Status values:

- **proven**: landed code with a named test.
- **specified**: a ratified spec holds it, and no code does yet.
- **planned**: a ratified spec schedules it to a rung or release.
- **designed, unscheduled**: a ratified spec designs it, and nothing schedules it.
- **new**: this spec introduces it.

| Mechanism | Lands in | Status |
|---|---|---|
| Invite-only, admin-added membership; no self-signup | Rung 5 roster (`TestRosterWrites_AdminInvariantAndGrant`, `TestAddMember_ActorAndChannelRefusals`, `TestRequireAdmin_RefusalsMatchNotFound`) | proven; the absence of a signup route waits on the route inventory |
| Roster in the store | Rung 6d `member import` | planned |
| Member cap on every add channel; preflight Navidrome import | Rung 14 settings; the roster's add path | new |
| Route inventory with an auth class per route | The next rung that registers a route | new |
| Login required for audio, art, and catalog data | Rung 5 gates the web app; 0.8 API surface | specified for 0.8 |
| No public share links; share endpoints refused at 0.8 | 0.8 slice 6 | new (amends the streamer spec) |
| Navidrome sharing pinned off, with a gate leg | Navidrome configuration now; rung 12 driver | new |
| Group-visible as the visibility ceiling | Configuration homes spec; rung 14 API | specified |
| No federation between servers | Build spec "Out of scope"; contributor guide | new (records a refusal) |
| App password per device, revocable alone | 0.8 slice 6 | planned |
| Credential cap per member, with the revoke list at the refusal | 0.8 slice 6 | new |
| Admin credential list with distinct addresses and one-step revoke | 0.8 slice 6; rung 14 admin API | new |
| Device credentials for the local client, counted by the cap | Configuration homes spec | designed, unscheduled |
| No fetching, no fetch integrations or hook recipes | Build spec "Out of scope"; CLAUDE.md; contributor guide | new (records a refusal) |
| `download` limited to a song or an album | 0.8 API surface | new (amends the streamer spec) |
| Served-volume meter per member and credential | 0.8 slice 5 (streaming); store-backed | new |
| Admin volume notice by email | 0.8; rung 5 SMTP notify path | new |
| Admin volume view | 0.8; rung 14 admin API | new |
| Credential id and bytes on serve events | 0.8 per-user state (a new forward hook) | new |
| Restore offline and command-line only | Rung 6b-ii | specified |
| Backup contents never returned over HTTP | Rung 6; rung 14 admin authorizer | specified |
| Source question, per upload with per-album override | Rung 7 uploader; intake API field | new |
| Source answer per mirrored collection | Local client mirror | new (designed, unscheduled carrier) |
| `added_by` per channel, including the mirror | Rung 7 submission, rung 9 admission (hook 17, amended) | planned |
| Album notes, by the `added_by` member or an admin | 0.8 group features | planned |
| Source shown beside `added_by` | Rung 7 album page; 0.8 `notes` | new |
| `musicBrainzId` filled in album responses | 0.8 API surface | new |
| Buy links with fallback order, purchase types, `https` only | The first web album page to render them (ruling 7) | new |
| Wishlist, merged with Requests; web only | `IDEAS.md` | new (idea) |
| ListenBrainz and Last.fm scrobbling | After 0.8, the streamer spec's "2.1 list" | planned |
| "dubplate is, dubplate isn't" and the legal note | README; website | new |
| Refused features, so contributors know before they build | CLAUDE.md now; contributor guide with the packaging initiative | new |

## Acceptance

Each row names the fixture, what passes, and the change that makes the test fail, so no test
can pass on an empty fixture. A rung plan cites rows by number.

| # | Mechanism | Fixture | Passes when | Fails when |
|---|---|---|---|---|
| A1 | Member cap | Roster at N-1 with cap N, on each of `member add`, the add-member page, `member import`, and `dubplate import navidrome`; an empty roster for the bootstrap; an over-cap roster at load | The Nth add lands; the next is refused naming the setting and offboarding; the bootstrap lands; the over-cap roster loads with a `doctor` row; an import over the cap writes nothing | The check is removed from any one channel |
| A2 | Route inventory | The real serve shell, every registered route | Every route matches its table row and auth class; an unauthenticated request is refused on every non-public row | A route is added without a row, or an auth wrapper is dropped |
| A3 | OpenSubsonic auth | Every implemented endpoint, no credentials | Only `getOpenSubsonicExtensions` answers; every other endpoint answers error 10 or 40 | Any other endpoint returns data |
| A4 | Shares absent | The four share endpoints; `getUser` | `createShare`, `updateShare`, and `deleteShare` answer error 50; `getShares` answers error 50 or, if the study rung requires it, an empty list; `shareRole` is false for every user; no share route is mounted | Any returns success |
| A5 | Download scope | One id of every kind the catalog mints | A song and an album answer; a one-album folder directory answers as its album; every other kind answers error 70 | Any other kind returns a body |
| A6 | Meter and notice | Two members, three credentials (two on one client name), one aborted Range stream, and one member at threshold minus one then plus one; a restart between reads | Counts and bytes are exact per credential; the notice fires once, naming the member and credentials; nothing is refused; the restart changes nothing | Two credentials merge, the notice fires twice or not at all, or the window resets |
| A7 | Credential cap and revoke | Mixed kinds at the cap, including an app password minted before the revoke | Minting is refused at the cap with the revoke list; the revoked credential's next request answers error 40 | A kind escapes the count, or the revoke lags |
| A8 | Volume view | The A6 fixture | The response decodes under `DisallowUnknownFields` into a struct with no entity-id field, and shows the measures ruling 5 names | A track or album id appears |
| A9 | Source answer | A submission without a source, and one per channel in the channel table | If ruling 6 says required, the missing answer is refused in reject wording; each channel stores the table's value | A missing answer is admitted under "required", or a channel stores another value |
| A10 | `added_by` per channel | One album per channel in the channel table | Each stores the table's value; a client-supplied value is ignored | Any channel stores a client value or resets on hand-over |
| A11 | Restore not over HTTP | The route inventory and the ops HTTP projection | No route or ops verb reaches restore | One does |
| A12 | Backup contents never returned | An HTTP backup under the admin authorizer | The response carries a handle and status only | Any snapshot bytes are in the body |
| A13 | Navidrome sharing pinned | The Navidrome configuration | `EnableSharing = false` is present | The key is absent or true |
| A14 | Buy links | One link of each allowed type, one streaming link, one `http` link, one `javascript:` link, and a release with no links whose release group and artist have them | Only the allowed `https` links render, in the stated order, with the fallback labeled | Another link renders, or the fallback is missing or unlabeled |

## Conflicts with existing specs

Each passage below conflicts with or needs an addition from this spec. Each needs an erratum in
its own document once Geoff approves this spec. This spec edits none of them.

1. **Shares.** The streamer spec's "Shares" section ships four share endpoints at launch, with
   a public token path that serves audio. The configuration homes spec (ratified
   2026-09-23) declines sharing outside the roster and reserves share rows only "if sharing
   ever ships". The streamer spec's revisions 2.2 (2026-09-12) through 2.4 (2026-09-24) left
   the section standing. The erratum drops shares from 0.8, per row A4, and touches every passage
   that names them:
   - the streamer spec's "Shares" section, the slice 6 name ("Identity, app passwords, and
     shares"), the "Web surface" section's share management, revision 2's item 12, the API
     surface's launch list, and hook 15's reserved share rows;
   - the configuration homes spec's reserved share rows, its rung 6 routing list, and
     `createShare` and "shares re-checked at fetch time" in the projector's test list;
   - the rung 6 spec's line that reserves share rows.
2. **Download scope.** The streamer spec's "API surface" section has `download` return album,
   artist, or playlist zips, the playlist zip with an `.m3u8`. The erratum limits it per row
   A5. The "Digital audio players" section names `download` as the server's contribution to
   device sync. The erratum says a player loads per album, or through the local binary's
   export profile.
3. **Credentials and the meter.** The streamer spec's "Identity and app passwords" section
   rate-limits failed sign-ins and sets no cap and no admin list. The erratum adds the
   credential cap and the admin list to slice 6, and the meter and notice to slice 5.
4. **Serve events.** The streamer spec's play history carries "user, song, client, timestamp,
   event kind, and how reported". The erratum adds a forward hook: every serve event carries
   the authenticating credential id and the bytes sent, and download is an event kind. The
   views over play history (play counts, most played, smart playlists) filter by event kind.
5. **`added_by` channels.** Hook 17 and the "Added by" paragraph say only intake sets
   `added_by`. The erratum replaces that with the channel table above: the server sets it from
   the authenticated principal of intake or of a mirror push, and copy, hand-over, and seize
   preserve it. The hook list also gains the source answer beside hook 17.
6. **Album responses.** The streamer spec's API surface gains `musicBrainzId` in `getAlbum`
   and `getAlbumInfo2`, and the source line in `notes`.
7. **Checkout.** The configuration homes spec's "Check out" verb lets a member take an offline
   copy of any server collection they can see, and the organizer spec's 2026-09-12 erratum and
   `ROADMAP.md` schedule it with the TUI. The erratum records ruling 2's answer.
8. **Navidrome sharing.** The build spec's trust-boundary requirements name a `/share/*`
   bypass. The erratum records it as inert while the driver pins sharing off, and routes the
   pin and its `doctor` row to rung 12.
9. **Refusals in the build spec.** The build spec's "Out of scope" section gains the refused
   features: fetching music, fetch integrations, federation, and public sharing.
10. **CLAUDE.md.** The project instructions gain a short "Refused features" block, because
    implementer dispatches and plan reviewers read that file and never read this spec.
11. **`IDEAS.md`.** The Requests entry absorbs the wishlist, with its purchase claim and its
    web-only constraint. A local record store link joins as its own idea.

No spec proposes federation, self-signup, public collections, or any way to fetch music. The
organizer spec's refusal of scrapers rests on API fragility and concerns metadata sources such
as lyrics. This spec's refusal of music acquisition rests on the product's purpose.

## Rulings for Geoff

1. **Is a member's own devices holding the whole library intended use?** Recommendation: yes.
   A phone holding the group's library is ordinary use, and Symfonium's automatic offline
   cache is how the leading client provides it.
   - Yes: the admin notice words a whole-library sync as expected ("Ana's new phone synced the
     library") and asks the admin to confirm the credential is hers. The website copy says
     members' devices can keep the library for offline listening.
   - No: nothing more is refused, because a cache fetch can't be told from playback. The
     notice words the same event as unusual, and the copy says offline listening covers what
     a member plays.
2. **Which collections may a member check out?** Recommendation: follow ruling 1.
   - Any collection they can see, counted by the meter. This matches a yes on ruling 1,
     because a checkout is the laptop's version of a phone's offline cache.
   - Only collections they own, plus the copies that hand-over and seize already leave behind.
     This matches a no on ruling 1.
3. **Navidrome downloads before 0.8.** Recommendation: leave them on and state the residual.
   The launch group is the founding group, the window closes at 0.8, and some clients' offline
   sync uses `download`.
   - On: nothing changes, and the "Before 0.8" section names artist and playlist zips as the
     residual.
   - Off: the Navidrome configuration sets `EnableDownloads = false` with a gate leg. Members
     keep streaming. Clients that sync through `download` lose offline sync until 0.8.
4. **The defaults.** Recommendation: a member cap of 20, a credential cap of 10 per member,
   and a volume notice when a member's distinct tracks served in a rolling 24 hours pass half
   the library's tracks. Twenty covers an extended family or a wide circle of friends and
   stays below any number that reads as an audience. Ten covers a member's phone, tablet,
   laptops, and a car or speaker. Half the library in a day is well above any listening day,
   and a first full sync still reaches it, which ruling 1 words. Each is a setting the admin
   can change, per member where it applies.
5. **What the volume view names.** Recommendation: name members and credentials on bytes and
   distinct tracks served, show stream counts only as group totals, and show no time-of-day
   breakdown. Each member's page says: "The admin can see how much data your devices download,
   not what you play." Only bytes and distinct tracks answer the view's one question, which
   is unusual copying. Per-person stream counts over time read as a listening log inside a
   family.
   - The recommendation builds per-member and per-credential byte and track columns, plus
     group-total stream counts.
   - Exceptions only: group totals by default, naming a member only when the notice fires.
   - Everything named: stream counts, bytes, and tracks per member and credential.
6. **Is the source question required?** Recommendation: required, with "Other" accepting an
   optional note. A skippable question lets the norm fade exactly when it matters, and the
   per-album override keeps a mixed batch honest.
   - Required: intake refuses a submission without an answer (row A9).
   - Optional: intake stores "unrecorded" for a missing answer.
7. **When buy links ship.** The web app has album pages from rung 7, before launch.
   Recommendation: rung 7, if a rung 7 plan task can absorb it without growing the pass.
   - Rung 7: the album page renders buy links, and the uploader captures the optional store
     link from the source question, which becomes the album's first buy link. Rung 9 captures
     the MusicBrainz links at admission as a new forward hook.
   - A later release: that release renders the links and captures them for albums already
     admitted. The uploader still captures the store link at rung 7, because upload is the
     only moment it exists.
