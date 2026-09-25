# What dubplate is for, and what it isn't

Design spec. Draft, 2026-09-25. Status: **awaiting Geoff's read.** Owner intent (Geoff,
2026-09-25): "I would like dubplate to be a tool for groups of family and friends to share
purchased music. I don't want it to be an easy way for stolen music to be mass-distributed."
The four principles below are the conductor's proposal, which Geoff accepted for write-up on
the same date. The mechanisms each principle names are this spec's decisions, except the
forks listed under "Rulings for Geoff".

dubplate is a music server for a small group of people who know each other, such as a family
or a circle of friends. Each member adds music they bought or own, dubplate tags and organizes
it into one shared library, and every member streams that library to their own phone and
computer. The group is closed: an admin adds each member, and nothing plays without a login.
dubplate has no way to find or download music from the internet, so the library holds only what
members bring. It records who added each album and where the music came from, and it points
members toward buying the music they like. dubplate is self-hosted, so these choices shape the
default product and what the project builds. They are not an enforcement system, and a
determined operator can change the code.

## dubplate is, and dubplate isn't

This section is written so a website page can use it nearly as it stands.

**dubplate is:**

- A shared library for a group of people who know each other. A family is the typical group;
  a few close friends is another.
- A place to pool music you bought or own: downloads from stores such as Bandcamp or Qobuz,
  rips of your CDs and records, and music someone gave you.
- A private server. An admin invites each member, and everything on it needs a login.
- A tool that keeps the library well organized. It tags every album against MusicBrainz and
  files it in one consistent layout.
- A way to support the artists you listen to. Album pages link to places to buy the music,
  and the group keeps a list of albums someone should buy.

**dubplate isn't:**

- A way to download music. It has no search-and-download feature and connects to no torrent,
  Usenet, or Soulseek network.
- A public music site. It has no public links, no anonymous listening, and no way to open a
  collection to people outside the group.
- A network of servers. One dubplate server doesn't connect to another.
- A way to hand a library to strangers. Members stream one track at a time and download one
  album at a time, and whole-library copies are backups that stay with the person who runs
  the server.
- Legal advice. Whether you may share music with your household or friends depends on your
  country and on the terms of the store you bought it from.

## What these principles can and cannot do

dubplate runs on hardware its operator controls. The repository carries an MIT license, and
the project may publish it. Anyone with the source can remove any limit this spec describes,
and nothing in dubplate can stop them.

These principles do two narrower things. They set the shape of the product that dubplate
ships, so the default server a stranger installs serves a small group and nothing larger.
They also record what the project refuses to build, so a feature request that would turn
dubplate into a distribution tool has a written answer. A limit that a fork can remove still
does useful work. It makes the ordinary use the easy use, and it keeps the project's own
effort pointed at the group it was built for.

The limits below aren't built to stop anyone. They make the intended use obvious, and they make
unusual volume visible to the admin.

## The group stays small and closed

**Membership is invite-only.** An admin adds each member. Rung 5 already builds this: the
roster is a file-backed store written by the `dubplate-server member` command-line verb and an
admin-only add-member page, and the web never grants the admin flag
(`2026-09-08-rung-5-identity-auth.md`, "Fork 1" and "The roster"). There is no self-signup. An
address that isn't on the roster gets a kind answer on the login page that names the admin
contact, and no account (fork 4 of the same spec).

**A soft member cap bounds the roster.** The server refuses a new member once the roster
reaches the cap, and the refusal names the setting that raises it. The cap is a setting in the
settings table, marked sensitive, so raising it requires fresh re-authentication and notifies
the admin address, as every sensitive write does
(`2026-09-11-configuration-homes-design.md`, "The server's own knobs"). The cap is soft
because the operator can raise it to any number. It exists so that growing past a group is a
deliberate act and never a default. `dubplate import navidrome` invites members through the
roster's normal add path (`2026-09-07-streamer-2-0-design.md`, "Hosting and setup"), so the
cap applies to it: the import stops adding members at the cap and reports each one it skipped.
The default number is ruling 1.

**Nothing is reachable without a login.** dubplate offers no public share link to audio, no
anonymous stream, and no public collection. The visibility ceiling for a collection is the
group. The configuration homes spec settles the two visibility levels, group-visible and
private, and declines sharing outside the roster (Geoff, 2026-09-11) because the roster is the
trust boundary. The unauthenticated surface is the login flow, the web app's static assets,
and `getOpenSubsonicExtensions`, the streamer spec's one unauthenticated API carve-out. None
of these serves audio, art, or catalog data.

The OpenSubsonic API calls a playlist "public" when every user on the server can see it. On
dubplate every user is a member, so a public playlist is visible to the group and no further.

**Servers don't federate.** A dubplate server exchanges no catalog, audio, or membership with
another dubplate server. The configuration homes spec's attach and mirror verbs connect a
member's own laptop to the server, under a device credential bound to that member. They don't
connect two servers.

**Devices belong to members.** The 0.8 streamer mints a named app password for each device a
member signs in on, shown once and revocable alone
(`2026-09-07-streamer-2-0-design.md`, "Identity: app passwords minted from the roster
account"). This spec adds three things to that design. Each member has a device cap, and the
device page refuses a new app password at the cap until the member revokes one. The admin has
a device list across all members, with each device's name, owner, creation time, and last-seen
address. The admin can revoke any one device in one step, beside the roster-level revoke that
the valid-after epoch already provides. The 0.9 local client pairs to a member in the same way,
through the server-minted device credential the configuration homes spec defines ("The writer
is a device credential"), and the same cap counts it. The default device cap is ruling 3.

## dubplate never fetches music

Music enters dubplate only when a member supplies files they hold. The entry paths are the
uploader (rung 7), the laptop mirror of a member's own collection, which is scheduled with the
local client, and the one-time migration of the founding library at rung 16. dubplate has no
search-and-download feature, no integration with torrent, Usenet, or Soulseek, and no
automation that fetches missing albums. The project doesn't accept a contribution that adds
one. The organizer spec already declines a plugin ABI (its extensibility fork), and the
documented API admits music only through intake, the same gate a member's upload passes.

This refusal is the core of the product's shape. A server that can fill its own library is a
distribution node whatever its member count. A server that holds only what members bring is a
shared shelf.

**No bulk export over the API.** Streaming is per track. The `download` endpoint returns one
track or one album, and album downloads are rate-limited per member. The streamer spec
currently lets `download` return artist and playlist zips as well; this spec narrows it, as
listed under "Conflicts with existing specs". The default album download limit is ruling 3.

Backup and restore stay admin operations on the operator's own storage. Restore is a
command-line verb that runs offline under the store's exclusive lock, and no HTTP route reaches
it (`2026-09-23-rung-6-ops-seam-store.md`, the field-channel table). A backup runs from the
command line, or over HTTP only behind rung 14's admin authorizer, and it writes to the remote
the operator configured. No path returns a backup's contents to the caller that asked for it.

## Every album has a source and a member

**The uploader asks where the music came from.** Rung 7's uploader asks one question per
upload: Bandcamp, Qobuz, CD rip, vinyl rip, gift, or other. dubplate doesn't verify the
answer. The question sets a norm: it tells every member, at the moment they add music, that
the group expects music someone paid for or owns. The answer travels with the submission and
lands on the album row beside `added_by`. A laptop mirror records one answer for the
collection when the admin approves its first mirror. Whether a member may skip the question is
ruling 2.

**The library shows who added each album.** The album row carries `added_by`, the uploading
member's surrogate id, which intake sets from the authenticated session and no other channel
writes (`2026-09-07-streamer-2-0-design.md`, forward hook 17 and "Group features"). The member
who added an album may attach a note, which reaches every client through `getAlbumInfo2`.
Albums migrated at rung 16 show "unknown".

**Admins can see volume.** The admin sees stream and download volume per member and per
device: counts of streams, album downloads, and bytes served over a window. Volume is counts
only. It names no track or album, so it answers "is something unusual happening" without
becoming a listening log. Stream starts are already an event kind in play history (forward
hook 4). This spec adds download events of the same shape. Whether the view names members or
shows only group totals is ruling 4.

## Money points back at artists

**Album pages carry buy and support links.** MusicBrainz records relationships from a release
to the places that sell or stream it. dubplate reads those relationships for each admitted
release and shows them as links on the album's page. The release in which they first appear
is ruling 5.

**The group keeps a wishlist.** A member adds an album that someone in the group should buy.
The `Requests` idea in `IDEAS.md` is the same list seen from the other side, a member asking
whether anyone has an album, so the two build as one feature: an entry closes when a member
uploads a matching album, and intake tells the member who added it. Neither is scheduled yet.

**Scrobbling is planned.** ListenBrainz and Last.fm scrobbling are on the streamer spec's list
of work after 0.8 ("The 2.1 list"). Both send listening data to services that credit artists
and help listeners find them.

## Declined mechanisms

**Per-member audio watermarking.** Marking each member's copy of a stream with an inaudible
identifier is heavy to build, because it moves every stream onto the transcode path. It is
brittle, since a further transcode can weaken or remove the mark. It also tells the members of
a family or friend group that the server distrusts them, which works against the product. The
added-by record and the volume view answer the same question with less cost.

**An ethical-use license.** A license that forbids distribution uses isn't an open source
license under the Open Source Definition. It deters the self-hosters dubplate is built for, who
read an unusual license as a legal risk. It can't be enforced against a fork that ignores it,
which is the only party it would target. The repository keeps its MIT license.

## A legal note

Whether members of a household or a group of friends may share purchased music depends on the
country they live in and on each store's terms of sale, and those terms differ from store to
store. The operator of a dubplate server is responsible for the music it hosts
and for who can reach it. Nothing in this spec, the README, or the website is legal advice.

## Where each principle lands

Status is "exists" when the property holds in a ratified spec or landed code, "planned" when a
ratified spec schedules it, and "new" when this spec introduces it.

| Mechanism | Lands in | Status |
|---|---|---|
| Invite-only, admin-added membership; no self-signup | Rung 5 roster, command-line verb and add-member page | exists |
| Soft member cap, a sensitive setting; Navidrome import stops at the cap | Rung 14 settings, enforced on the roster's add path | new |
| Login required for all audio, art, and catalog data | Rung 5 auth; 0.8 API surface | exists |
| No public share links (share endpoints dropped from 0.8) | 0.8 identity slice | new (amends the streamer spec) |
| Group-visible as the visibility ceiling; no sharing outside the roster | Configuration homes spec; rung 14 API | exists |
| No federation between servers | Build spec principles | new (records a refusal) |
| App password per device, revocable alone | 0.8 identity slice | planned |
| Per-member device cap | 0.8 identity slice | new |
| Admin device list with one-step revoke | 0.8 identity slice; rung 14 admin API | new |
| Local client pairs to a member through a device credential | 0.9 (configuration homes spec's device credential) | planned |
| No search-and-download, no torrent, Usenet, or Soulseek, no fetch automation | Build spec principles; contributor guide | new (records a refusal) |
| Downloads per track or per album, never artist or playlist zips | 0.8 API surface | new (amends the streamer spec) |
| Album download rate limit per member | 0.8 streaming slice | new |
| Restore offline and command-line only; backup to the operator's remote | Rung 6 | exists |
| Upload source question | Rung 7 uploader | new |
| Source recorded once per mirrored collection | The laptop mirror, with the local client | new |
| `added_by` on every admitted album | Rung 7 submission, rung 9 admission (hook 17) | planned |
| Album notes | 0.8 group features | planned |
| Per-member and per-device volume view for admins | 0.8; rung 14 admin API | new |
| Download events beside stream-start events | 0.8 per-user state | new |
| Buy and support links from MusicBrainz relationships | Ruling 5 | new |
| Group wishlist, merged with the Requests idea | After 0.8, unscheduled | new |
| ListenBrainz and Last.fm scrobbling | After 0.8 ("The 2.1 list") | planned |
| "dubplate is, dubplate isn't" and the legal note | README; website | new |
| Refused features, so contributors know before they build | Contributor guide (does not exist yet) | new |

## Conflicts with existing specs

Three passages in ratified specs conflict with these principles. Each needs an erratum in its
own spec once Geoff approves this one.

1. **Shares in the streamer spec.** `2026-09-07-streamer-2-0-design.md`, "Shares" (lines
   289-295), ships `createShare`, `updateShare`, `deleteShare`, and `getShares` at launch, with
   a token-gated public path that serves audio to anyone holding the link. The ladder's slice 6
   is "Identity, app passwords, and shares", and the "Web surface" section puts share
   management in the web app. This contradicts the configuration homes spec, ratified later,
   which declines sharing outside the roster (Geoff, 2026-09-11) and files share rows only as a
   reserved schema hook "if sharing ever ships". The streamer spec's revision 2.2 folded
   configuration homes amendments but left this section standing. The erratum drops shares
   from 0.8: the four endpoints answer error code 0 as unimplemented endpoints do, unless the
   study rung finds a client in the roster that requires an empty success. The configuration
   homes spec's reserved share rows and its "shares re-checked at fetch time" line in the
   projector hook retire with it.
2. **Download scope in the streamer spec.** The "API surface" section (lines 384-386) has
   `download` return album, artist, or playlist zips. An artist zip can span most of a
   library, and a playlist zip can hold any selection. The erratum limits `download` to one
   track or one album and adds the per-member album limit.
3. **Device limits in the streamer spec.** The "Identity and app passwords" section rate-limits
   failed authentication but sets no device cap and gives the admin no device list. This is an
   addition rather than a contradiction, and the erratum adds both to slice 6.

No spec proposes federation, self-signup, public collections, or any way to fetch music. The
organizer spec's refusal of scrapers (`2026-09-01-dubplate-organizer-design.md`, "Scrapers are
refused") rests on API fragility and maintenance grounds and concerns metadata sources such as
lyrics. This spec's refusal of music
acquisition rests on the product's purpose and stands on its own.

## Rulings for Geoff

1. **The default member cap.** Recommendation: 20. It covers an extended family or a wide
   circle of friends with room to spare, and it stays well below any number that reads as an
   audience. An operator with a larger real group raises it in one sensitive setting.
2. **Whether the source question is required.** Recommendation: required, asked once per
   upload, with "other" accepting an optional short note. A required question with an honest
   "other" costs one click, and a skippable one lets the norm fade exactly when it matters. The
   answer is never verified, so requiring it adds no friction beyond that click.
3. **The default album download limit and device cap.** Recommendation: 20 album downloads per
   member per rolling 24 hours, and 10 devices per member. Twenty albums covers loading a
   phone or a car's music for a trip. Ten devices covers a member's phone, tablet, laptops, and
   a car or speaker. The limit counts album zips only; single-track downloads fall under the
   volume view instead. Both are settings an admin can change per member.
4. **Whether the volume view names members.** Recommendation: name members, visible to admins
   only, and tell each member on their own page that the admin sees their stream and download
   counts. Group totals alone can't show which account or device is behaving unusually, which
   is the view's only purpose, and the counts carry no titles. Telling members keeps the view
   honest inside a group that trusts each other.
5. **When buy links ship.** Recommendation: with the first release that has an album page in
   the web app, which is the first-party web player after 0.8. In 0.8 the group listens through
   OpenSubsonic clients, and the only album text dubplate controls there is the `notes` field,
   which already carries the member's note. Mixing store links into it would crowd that note.
   Under the rule that a hook is built by its first consumer, the pass that renders the links
   also captures the MusicBrainz relationships.
