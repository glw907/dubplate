# What dubplate is for, and what it isn't

Design spec. Revision 3, 2026-09-25. Status: **ratified (Geoff, 2026-09-25); all twelve rulings ruled, 4 to 12 as recommended.**
Revision 1 (`1d6bc6a`) went through four review lenses (contract, mechanics, misuse,
consistency) and an early prose read. Revision 2 (`08c812f`) went through a fold verification
read, and revision 3 folds that read and Geoff's rulings of 2026-09-25. The fold record,
including its "Second fold" section, is
`docs/superpowers/research/2026-09-25-group-scale-principles-fold.md`.

Revision 3 changelog: bounded shares filled in from research
(`docs/superpowers/research/2026-09-25-bounded-shares-research.md`), adding rulings 8 to 12.

Owner intent (Geoff, 2026-09-25): "I would like dubplate to be a tool for groups of family and
friends to share purchased music. I don't want it to be an easy way for stolen music to be
mass-distributed." Owner framing, the same day: "I realize that I can't totally stop piracy. My
goal is simply for dubplate to be easy to use for its intended purpose, and hard to use for
nefarious ones. And it's a bonus if it helps induce people to buy digital and physical music
that supports artists."

That framing is the test every mechanism below passes. A mechanism that adds friction for an
honest member and none for an abuser fails it, and revision 2 dropped the one that did,
revision 1's per-member album download limit. The mechanisms are this spec's decisions, except
the questions listed under "Rulings for Geoff".

dubplate is a music server for a small group of people who know each other, such as a family
or a few close friends. Each member adds music they bought or own. dubplate tags it, organizes
it into one shared library, and streams that library to every member's phone and computer. An
admin adds each member, and nothing plays without a login except a bounded share link a member
sends. dubplate can't get music from the internet, so the library holds only what members
bring. These choices shape the product the project ships. dubplate is self-hosted, so they
aren't an enforcement system. The operator of a server can change the code.

## Versions this spec speaks to

- **The launch (0.x).** The greenfield ladder ends in a 0.x pre-release that the group uses.
  Navidrome still serves the audio in this period, with the residuals that "Before 0.8" lists.
- **0.8.** dubplate's own streamer replaces Navidrome. Most mechanisms below land here.
- **0.9.** A local client paired with the server. Its scope isn't specified yet.
- **1.0.** Polish and more audio formats.

## dubplate is, and dubplate isn't

This section is website copy for dubplate from 0.8, the first release that serves audio
itself. Every line is true of the product as decided from that release. The website publishes
it when 0.8 ships. Until then the server runs Navidrome, and "Before 0.8" lists where the
running server differs. A line about a feature that isn't built yet says so. The share line
follows rulings 8 and 11, and the buy-link line follows ruling 7. Each line changes with its
ruling.

**dubplate is:**

- A shared music library for a group of people who know each other, such as a family or a
  few close friends.
- A place to pool music you bought or own: lossless (FLAC) downloads from stores such as
  Bandcamp or Qobuz, rips of your own CDs and records, and music given to you, such as a CD or
  a download someone bought for you.
- A private server. An admin adds each member, and the library needs a login.
- A way to send a friend one album. A link to one album or song plays in the browser for a
  week, and you can extend it to a month. The friend needs no account. The link plays the
  music but never hands over the files.
- A tidy library. dubplate looks up every album in MusicBrainz, the open music encyclopedia,
  and files it in a consistent layout.
- A record of who added each album and what they said about where it came from.
- Built to point you toward buying music. Album pages will link to places to buy the music,
  where MusicBrainz knows one.

**dubplate isn't:**

- A way to get music from the internet. It has no search-and-download feature, and it
  connects to no torrent, Usenet, or Soulseek network.
- A public music site. It has no public library, no link that opens the whole collection, and
  no way to let anyone outside the group browse it.
- A network of servers. One dubplate server doesn't connect to another.
- A way to hand a library to strangers. dubplate has no download of an artist's catalog, a
  playlist, or the whole library in one piece. Each member's own phone or computer can still
  keep the whole library for offline listening, and the admin can see how much music members
  download.
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

**Nothing is reachable without a login except a bounded share.** dubplate has no anonymous
stream of the library and no public collection. A collection is either visible to the group
or private, and the roster is the trust boundary. The one exception is a share link a member
creates, bounded as "Bounded shares" describes. A checked-in table lists the unauthenticated
surface. A route-inventory test builds the real serve shell, compares every registered route
against the table's auth class, and fails on any route the table doesn't name. The classes
are public, session, admin, and delegated. A delegated route hands authentication to another
server. Its one row is the shell's fallthrough (`Shell.SetFallthrough`), which from rung 12
forwards `/rest/*` to Navidrome under Navidrome's own passwords, and "Before 0.8" covers what
that route serves. Today's public routes are the login flow, the web app's static assets, and
the branding, font, and operator CSS routes. `dubplate init`'s first-run page on an
unconfigured server joins at 0.8. At 0.8 the table extends to every OpenSubsonic endpoint, and
only `getOpenSubsonicExtensions` answers without credentials. The share link's route joins
the public rows at 0.8, and it is the only public row that serves audio or art. No public row
serves catalog data.

The OpenSubsonic API calls a playlist "public" when every user on the server can see it. On
dubplate every user is a member, so a public playlist is visible to the group and no further.

**Servers don't federate.** A dubplate server exchanges no catalog, audio, or membership with
another dubplate server. The configuration homes spec's attach verbs connect a member's own
laptop to a server, under a device credential bound to that member. A member who holds
profiles for two servers can still copy albums out of one and contribute them to the other by
hand. The sending server's meter counts the copy. Intake's per-album checks on the receiving
server and that server's admin bound the rest.

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
  often. The address comes from `serve.ClientIP`, which returns the real client address only
  in trusted-proxy mode. In the default mode it returns the connecting peer, which behind the
  tunnel is the tunnel itself. So the list shows the count only in trusted-proxy mode, and a
  `doctor` row names the gap otherwise.
- **A one-step admin revoke** of any one credential, beside the roster-level revoke that the
  valid-after epoch already provides. A revoked credential's next request answers error 40.

## Bounded shares

Casual one-off sharing is an important feature. Geoff, 2026-09-25: "I want to be as easy and
permissive as I can without making dubplate attractive to pirates." A member who sends a friend
a link to one album is using dubplate as intended. A permanent public link to the library
isn't. So 0.8 keeps OpenSubsonic shares, with six bounds.

The bounds follow `docs/superpowers/research/2026-09-25-bounded-shares-research.md` (cited
below as "the research"). A share link fails as a piracy tool when it offers nothing a pirate
can't get more easily elsewhere. Released music is already free to stream as lossy audio, and
stream ripping is the most common form of consumer piracy (the research, "How share links get
abused"). What a pirate would want from a dubplate link is lossless files, many albums behind
one link, and a link that lasts. The six bounds remove those things and leave the rest alone.

The threat these bounds answer is a link that escapes its intended friend. A member acting in
bad faith already holds downloads of the whole library under ruling 1, so no share bound stops
that member. The meter is the tool for that case.

Each bound below gives its default, its evidence, and its cost to honest use. Rulings 8 to 12
set the numbers the evidence doesn't settle.

**1. One item per share.** A share names one song or one album. A folder-view directory that
holds exactly one album counts as that album, as it does for `download`. A playlist is ruling
11. A request that names more than one id, or an artist, a music folder, or a directory above
the album level, answers error 50 with a message that names the one-item rule. The web app
offers only the allowed kinds.

- Evidence: every documented abuse case is about a whole library, not an album. Navidrome
  allows artists and several ids per share, and DSub sends every selected id (the research,
  "How share links get abused" and "How casual sharing works today").
- Cost to honest use: a member sends two links for two albums.

**2. A lossy stream, never a file.** A share serves audio only as a lossy stream, at 128 kbps
Opus VBR by default (ruling 12). A browser that can't play Opus gets a codec it plays at an
equivalent rate. A share has no download route, never serves the original format, and has no
M3U list of stream URLs. The encode runs under the streamer's transcode budget.

- Evidence: the OpenSubsonic API has no downloadable flag, so the server decides. Communities
  that value quality reject lossy sources. Navidrome's share dialog defaults to the original
  format, which serves FLAC from a FLAC library (the research, "How casual sharing works
  today" and "How share links get abused").
- Cost to honest use: near zero, since Xiph calls Opus at 128 kbps "pretty much transparent"
  for stereo music (the research, "What bounds cost honest listeners"). A friend who wants the
  files follows the buy links.

**3. Expiry clamped, not refused.** A share expires 7 days after creation by default, and
never later than 30 days after creation (ruling 8). The member can extend a share up to that
maximum. The server clamps `expires` instead of refusing it, so stock clients work.

- An absent, null, or zero `expires`, or a time already past, becomes the default.
- A time past the maximum becomes the maximum. `updateShare` clamps the same way, measured from
  the share's creation.
- Every stream and art request checks the share's expiry and deletion, not only the page.
- Evidence: Navidrome's one-year default, and its bug that kept expired and deleted shares
  streaming until v0.62.0. The Jellyfin ShareLinks plugin defaults to 24 hours and caps at 30
  days, and Nextcloud lets an admin enforce a maximum. Secret links leak through mail scanners
  and search indexes (the research, "How casual sharing works today", "How comparable products
  bound private sharing", and "How share links get abused").
- Cost to honest use: a friend who opens the link after day 7 asks for a new one. In one 2011
  Bitly measurement, messenger links got half their clicks within about 3.4 hours of their peak
  (the research, "What bounds cost honest listeners").

**4. A pause when a link draws a crowd.** A share pauses when 50 distinct listeners have played
it over its life (ruling 9). A paused share's page says the link is paused, and its streams
answer no audio.

- A listener counts at the first audio byte served, never at a page view, so a chat app's
  link-preview fetch doesn't count.
- The share page sets a random per-browser cookie, and that cookie is the listener key. An
  address isn't, so one person moving between Wi-Fi and cellular counts once. A stream request
  without the cookie counts as a new listener, so clearing the cookie can only raise the count.
- At the pause, the member and the admin each get one notice through the notify path the
  volume notice uses.
- The member can resume a paused share once. The resume doubles the share's threshold, and the
  second pause is final.
- Share traffic counts in the sharing member's meter, under the share's id where a credential
  id would sit.
- Evidence: Bandcamp alerts an artist at more than 100 unique visits in 48 hours on a private
  link. In one study of WhatsApp admins, family groups averaged 30 members and friends groups
  20 (the research, "How comparable products bound private sharing" and "What bounds cost
  honest listeners").
- Cost to honest use: none below the threshold.

**5. An unguessable, unindexed link.** The share token carries 128 bits from `crypto/rand` and
is compared in constant time, as the streamer spec already requires. The share page and every
share response send `X-Robots-Tag: noindex` and `Referrer-Policy: no-referrer`. No listing
names a share, except the member's own shares and the admin's view. `robots.txt` never names
the share path, because a crawler that robots.txt blocks can't read the page's noindex.

- Evidence: OWASP asks for at least 64 bits in a session identifier, and Navidrome's share id
  has about 59.5 bits. Mail scanners and chat previews fetch every link they see, so a token
  is a bearer secret that leaks (the research, "How casual sharing works today", "How share
  links get abused", and bound 5 of its recommended set).
- Cost to honest use: none.

**6. Server-side clamping, owner from authentication.** One share service enforces bounds 1 to
5, and the web app and `createShare` both call it. The share's owner is the authenticated
member. Any owner, user, or username field in the request is ignored.

- Evidence: DSub sends `expires=0` and every selected id, and Tempo sends a null `expires`.
  Navidrome's High-severity advisory GHSA-82gh-4ggp-gfg5 came from trusting a client-supplied
  `userId` (the research, "How casual sharing works today").
- Cost to honest use: none.

**Buy links are a feature, not a bound.** The share page shows the album's buy links under the
rules in "Money points toward buying", the same as the album page. The evidence that they sell
music is weak. No public data ties share pages to purchases, and studies of free listening
find mixed effects on sales (the research, "The buy-link funnel"). They stay because they cost
little and point a friend who likes the album toward the artist.

**Considered and dropped.** Each of these costs honest listeners more than it costs an abuser,
or it duplicates a bound above (the research, "Bounds considered and dropped").

- A play cap. A 25-play cap stops a 12-track album in a family chat on the third listener, and
  the listener pause bounds a crowd without counting repeat listens.
- A cap on live links per member. It can't stop a member acting in bad faith, who already holds
  the library's downloads.
- Preview clips. A clip defeats sending a friend an album, and a full lossy stream offers a
  pirate nothing a free stream lacks.
- Attribution as a deterrent. No evidence shows that naming the sharer deters a leak. Whether
  the page names the member is ruling 10, decided on the friend's experience.
- A password on the link. It is a second secret that travels with the first.
- Per-link bandwidth caps. A lossy stream is small, and the pause and the transcode budget
  already bound load.
- A robots.txt block of the share path. It hides the noindex from the crawler it targets.

## dubplate never fetches music

Music enters dubplate only from files a member or the operator already holds. dubplate has no
search-and-download feature, no integration with torrent, Usenet, or Soulseek, and no
automation that fetches missing albums. The project doesn't accept a contribution that adds
one. This is a project policy. The documented API and the lifecycle hooks can carry files
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
other id kind answers error 50, which the OpenSubsonic error list defines as "User is not
authorized for the given operation.", with a message that names the song-or-album rule. The
refused kinds are an artist, a playlist, a folder-view directory above the album level, and a
music folder. Error 70 ("The requested data was not found.") would be untrue for an id that
exists, and error 50 is what Navidrome answers when downloads are off. The streamer spec
currently lets `download` return artist and playlist zips. This spec narrows it, as listed
under "Conflicts with existing specs". An artist zip can span most of a library, and a
playlist can hold any selection.

**Offline listening is intended use.** The group listens on phones, and phones keep music for
offline play. Stock clients do this one track at a time. Tempus fetches each track through
`download` or `stream`, Symfonium through `download`, and DSub through `stream`. Symfonium's
automatic offline cache copies every track in the library after each sync. A per-track fetch
for an offline cache can't be told apart from a script copying the library. So dubplate
refuses neither, and it counts both. A member's own devices holding the whole library is
intended use (ruling 1, ruled).

**A meter counts everything served.** Every `stream` and `download` response is counted
against the member and the credential that authenticated it, in bytes sent and distinct tracks,
over the trailing 24 hours. Revision 1's per-member album download limit is dropped. It counted
album zips, which no offline-sync client uses, so it would have stopped almost no one and
landed only on an honest member.

- The meter is store-backed, so a restart doesn't reset its window.
- Bytes are the bytes written per response, including a partial (206) response and an aborted
  stream.
- A credential's distinct tracks are the tracks with any byte served to it in the trailing 24
  hours. A member's distinct tracks are the union over the member's credentials, so two devices
  syncing the same track count it once. A request with a nonzero Range offset adds bytes to
  the credential's byte count and records no new track.
- The credential id comes from authentication, never from the client's `c` parameter, which a
  client asserts and two phones running one app share.

**A notice tells the admin.** When a member's distinct tracks served in a rolling 24 hours
pass a threshold, the admin gets one email naming the member and the credentials that crossed
it, through the same notify path the configuration homes spec uses for sensitive changes.
Nothing is refused. The threshold is a setting whose default sits well above a heavy honest
day. The default is ruling 4. The notice fires once per crossing, not once per request.
Because ruling 1 makes a whole-library sync intended use, the notice words it as expected
("Ana's devices synced most of the library today") and asks the admin to confirm the
credentials are hers.

**The admin sees volume.** The admin's volume view shows bytes and distinct tracks served.
Ruling 5 decides whether it breaks them out by member and credential, and which measures it
shows. The view shows counts and names no track or album. Play history, which the streamer
keeps for recently played and most played, does record songs per member. An admin with shell
or backup access can read it. Each member's own page says so in plain words, in the sentence
ruling 5 gives.

**Backup and restore stay with the operator.** Restore is a command-line verb that runs
offline under the store's exclusive lock, and no HTTP route reaches it
(`2026-09-23-rung-6-ops-seam-store.md`, the field-channel table). A backup runs from the
command line, or over HTTP only behind rung 14's admin authorizer, and it writes to the remote
the operator configured. No response returns a backup's contents to the caller.

**A member may check out any collection they can see.** The configuration homes spec's
checkout verb takes a read-only offline copy of a server collection onto a member's laptop,
and the group library is one such collection. A checkout is the laptop's version of a phone's
offline cache, so it follows ruling 1 (ruling 2, ruled). Checkout is served bytes under a
device credential, so the meter counts it like any other. Checkout ships with the TUI, which
the ladder doesn't schedule yet. The 0.9 local client's scope isn't specified, so whether it
carries checkout is open.

## Before 0.8

From launch until 0.8 replaces it, and afterward whenever Navidrome is the escape hatch,
Navidrome serves the audio. dubplate's sign-on gates the web app, and `/rest/*` authenticates
through Navidrome's own passwords. In this period the running server differs from the
is/isn't copy in three ways.

- **Sharing is on, at Navidrome's default.** Navidrome 0.63.2, the version `compose.yaml`
  pins, defaults `EnableSharing` to `true` (`conf/configuration.go:797`). The repo sets
  neither the key in `config/navidrome.toml` nor `ND_ENABLESHARING` in the container's
  environment. The tunnel forwards every path, `/share/*` included, to Navidrome. So a member
  can create a public share link that serves audio without a login. A read-only check on
  2026-09-25 found zero shares in production. Navidrome's share links there keep its defaults:
  a one-year expiry, several ids of one kind, and the original format. The pinned 0.63.2 also
  predates the fix for advisory GHSA-82gh-4ggp-gfg5, in which share creation trusted a
  client-supplied `userId` (the bounded-shares research, "How casual sharing works today").
- **Downloads are on, at Navidrome's default.** Navidrome 0.63.2 defaults `EnableDownloads` to
  `true` (`conf/configuration.go:758`), and its `download` returns artist and playlist zips.
- **Credentials and volume aren't covered.** Navidrome has one Subsonic password per user, no
  credential cap, and no meter.

The owner accepted the first two as residuals. The third closes at 0.8 with the credential cap
and the meter. Geoff, 2026-09-25: "I'm not worried about Navidrome now. We'll fix it when we
switch it over to dubplate." No pin, gate leg, or deploy is owed for this period. The website
publishes the is/isn't copy at 0.8, so it never describes this period.

At 0.8 dubplate's streamer serves shares only under the bounds that "Bounded shares" sets
(row A4). One rule governs any check on sharing from then on, including a check on Navidrome
while it stays the escape hatch. The check reads the effective value, never a configuration
file alone. Navidrome's configuration docs give the precedence as "environment variables,
command line arguments, configuration file", so a check that reads `navidrome.toml` passes
while `ND_ENABLESHARING` turns sharing back on. The escape hatch's `doctor` row reads `shareRole`
from Navidrome's `getUser`, which Navidrome sets from `EnableSharing` (row A13).

## Money points toward buying

**Album pages carry buy links.** MusicBrainz records links from a release to the places that
sell it. dubplate captures them for each admitted release and shows them on the album's page
in the web app. The web app has album pages from rung 7
(`2026-09-24-visual-identity-design.md`, "Open for rung 7"), so the release that first renders
buy links is ruling 7.

- **Which links count.** Only purchase types count. From a release, those are `purchase for
  download` and `purchase for mail-order`. From an artist, they are `bandcamp`, `purchase for
  download`, and `purchase for mail-order`. Streaming-service links don't count, because the
  section's purpose is buying. Every link must be `https`.
- **Where the links come from.** dubplate takes the first source that yields at least one
  link that counts, in this order: the admitted release, its sibling releases in the same
  release group, then the artist. Artist links render only when the release and its siblings
  yield none. Most physical
  releases carry no links, because stores link the digital release. A sampled 3 of 28 releases
  carried a buy or stream link. Every credited artist carried a buy or stream link, and how
  many carry a purchase link is unmeasured. Sibling releases come through one browse call
  under the shared MusicBrainz rate limiter. Links from the artist carry the label "the
  artist's", since they aren't this album's.
- **What order they appear in.** MusicBrainz has no type for the artist's own store, so the
  order uses the link's host. A link whose host is `bandcamp.com` or one of its subdomains, or
  equals the host of the artist's `official homepage` link when MusicBrainz records one,
  comes first. dubplate reads that homepage link whichever source yields. Then come
  `purchase for mail-order` links, then the other `purchase for download` links. Links in the
  same group sort by URL. A store the artist runs pays the artist the most.
- **Staleness.** Links are captured at admission and can go stale. The pass that builds the
  feature sets the refresh cadence.

**The source shows beside who added the album.** The album page's inscription names who added
the album and when (visual identity spec). It also names the source, such as "from Ana's
vinyl". From 0.8, `getAlbumInfo2`'s `notes` field carries the same line under the member's
note. It reminds every listener where the group's music came from.

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
the product exists for. When the server refuses a fetch, Symfonium retries, and its docs tell
the user to switch off the download endpoint.
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
- **residual**: a gap the owner accepted for a stated period.

| Mechanism | Lands in | Status |
|---|---|---|
| Invite-only, admin-added membership; no self-signup | Rung 5 roster (`TestRosterWrites_AdminInvariantAndGrant`, `TestAddMember_ActorAndChannelRefusals`, `TestRequireAdmin_RefusalsMatchNotFound`) | proven; the absence of a signup route waits on the route inventory |
| Roster in the store | Rung 6d `member import` | planned |
| Member cap on every add channel; preflight Navidrome import | Rung 14 settings; the roster's add path | new |
| Route inventory with an auth class per route, the fallthrough classed as delegated | The next rung that registers a route | new |
| Login required for audio, art, and catalog data | Rung 5 gates the web app; 0.8 API surface | specified for 0.8 |
| Shares kept at 0.8 under six bounds: one item, lossy stream, clamped expiry, listener pause, unindexed token, server-side clamping | 0.8 slice 6 | new (amends the streamer spec) |
| Navidrome sharing and downloads left at their defaults before 0.8 | Residual the owner accepted, 2026-09-25 | residual |
| Any sharing check reads the effective value; escape-hatch `doctor` row on `shareRole` | 0.8 | new |
| Group-visible as the visibility ceiling | Configuration homes spec; rung 14 API | specified |
| No federation between servers | Build spec "Out of scope"; contributor guide | new (records a refusal) |
| App password per device, revocable alone | 0.8 slice 6 | planned |
| Credential cap per member, with the revoke list at the refusal | 0.8 slice 6 | new |
| Admin credential list with distinct addresses (trusted-proxy mode) and one-step revoke | 0.8 slice 6; rung 14 admin API | new |
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
| Buy links: first source that yields, host-based order, purchase types, `https` only | The first web album page to render them (ruling 7) | new |
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
| A2 | Route inventory | The real serve shell, every registered route, the fallthrough set | Every route matches its table row and auth class (public, session, admin, or delegated); an unauthenticated request is refused on every session and admin row; the fallthrough is the only delegated row | A route is added without a row, or an auth wrapper is dropped |
| A3 | OpenSubsonic auth | Every implemented endpoint, no credentials | Only `getOpenSubsonicExtensions` answers; every other endpoint answers error 10 or 40 | Any other endpoint returns data |
| A4 | Shares bounded | Each case below twice, once through the web app's share service call and once through a client `createShare`: a song, an album, and a one-album folder directory; an artist id; two album ids, as DSub sends; an absent `expires`; Tempo's null `expires`; DSub's `expires=0`; an `expires` already past; an `expires` past the maximum; an `updateShare` past the maximum from creation; a request whose owner field names another member; an expired share and a deleted share, each fetched on the stream route; one share at threshold minus one listener, then a page view, a preview fetch of the page, and the next first audio byte; one cookie from two addresses; one stream request with no cookie; a resumed share at twice the threshold; under ruling 11, a playlist at and one track over the limit | Song, album, and one-album directory shares land; the artist, the two ids, and any other kind answer error 50 with the one-item message; absent, null, zero, and past times clamp to the default, and a later time clamps to the maximum; the owner is the authenticated member; the stream is lossy at ruling 12's rate, and no download, original-format, or M3U request returns audio; expired and deleted shares serve no bytes; the count moves only at the first audio byte, counts one cookie once across addresses, and counts a cookieless request as new; the pause lands at the threshold with one notice each to the member and the admin, and the second pause refuses a resume; the page and every share response carry `noindex` and `no-referrer`, the token has 128 bits, and `robots.txt` names no share path; share bytes appear in the member's meter | Any bound is missing on either path, or a client field sets the owner |
| A5 | Download scope | One id of every kind the catalog mints | A song and an album answer; a one-album folder directory answers as its album; every other kind answers error 50 with the song-or-album message | Any other kind returns a body |
| A6 | Meter and notice | Two members, three credentials (two on one client name), one track served to two of one member's credentials, one aborted stream and its nonzero-offset Range continuation, and one member at threshold minus one then plus one; a restart between reads | Per credential, distinct tracks and bytes match the trailing-24-hour rule; the continuation adds bytes and no track; the shared track counts once for the member; the notice fires once, naming the member and credentials; nothing is refused; the restart changes nothing | Two credentials merge, the continuation counts a track, the notice fires twice or not at all, or the window resets |
| A7 | Credential cap, admin list, and revoke | Mixed kinds at the cap, including an app password minted before the revoke; one credential used from three client addresses in trusted-proxy mode, and the same requests in the default mode | Minting is refused at the cap with the revoke list; the admin list shows three addresses in trusted-proxy mode, no count and a `doctor` row in the default mode; the revoked credential's next request answers error 40 | A kind escapes the count, the default mode shows a count, or the revoke lags |
| A8 | Volume view | The A6 fixture | The response decodes under `DisallowUnknownFields` into a struct with no entity-id field, and shows the measures ruling 5 names | A track or album id appears |
| A9 | Source answer | A submission without a source, and one per channel in the channel table | If ruling 6 says required, the missing answer is refused in reject wording; each channel stores the table's value | A missing answer is admitted under "required", or a channel stores another value |
| A10 | `added_by` per channel | One album per channel in the channel table | Each stores the table's value; a client-supplied value is ignored | Any channel stores a client value or resets on hand-over |
| A11 | Restore not over HTTP | The route inventory and the ops HTTP projection | No route or ops verb reaches restore | One does |
| A12 | Backup contents never returned | An HTTP backup under the admin authorizer | The response carries a handle and status only | Any snapshot bytes are in the body |
| A13 | Escape-hatch sharing check (0.8) | A Navidrome fake whose configuration file sets sharing off while its environment sets `ND_ENABLESHARING=true`, so `getUser` reports `shareRole` true; and one reporting false | The `doctor` row flags the first and passes the second | The check reads the configuration file, or the first passes |
| A14 | Buy links | A release with one link of each release purchase type, a `purchase for download` link on a Bandcamp host, one on the artist's `official homepage` host, one streaming link, one `http` link, and one `javascript:` link; a release with no links whose siblings have them; a release whose siblings have none and whose artist has them | Only counted `https` links render; the Bandcamp-host and homepage-host links sort first, then mail order, then other download stores; the first source that yields is the only one shown; artist links carry the artist's label | Another link renders, a host-matched link sorts after another, two sources mix, or artist links are unlabeled |
| A15 | `musicBrainzId` in album responses | An admitted album with a known release id | `getAlbum` and `getAlbumInfo2` carry that id in `musicBrainzId` | The field is empty or carries another id |

## Conflicts with existing specs

Each passage below conflicts with or needs an addition from this spec. Each needs an erratum in
its own document once Geoff approves this spec. This spec edits none of them.

1. **Shares.** The streamer spec's "Shares" section ships four share endpoints at launch, with
   a public token path that serves audio. The configuration homes spec (ratified
   2026-09-23) declines sharing outside the roster and reserves share rows only "if sharing
   ever ships". The streamer spec's erratum makes these changes, per "Bounded shares".
   - Scope: "one album or one playlist" becomes one song or one album, with a playlist only as
     ruling 11 decides. Every other id kind, and more than one id, answers error 50.
   - Expiry: "An expiry is required" becomes an expiry clamped to the default and the maximum
     (ruling 8), never refused. Every stream and art request checks expiry and deletion.
   - Audio: "it serves the direct path, or a configured per-share encode budget" becomes a
     lossy stream only, at ruling 12's rate, under the transcode budget. The direct path, the
     original format, a download route, and an M3U list are gone from shares.
   - Crowd: the listener pause, its cookie key, its notices, its one resume, and share bytes in
     the member's meter join the section.
   - Headers: the page and every share response send `noindex` and `no-referrer`, and
     `robots.txt` never names the share path.
   - Owner: the owner comes from authentication, and one share service serves the web app and
     `createShare`.
   - The 128-bit token, the constant-time compare, the per-address rate limit, and the revoke
     stay as written. Slice 6's name, share management in "Web surface", and the launch list's
     shares stay.
   - Hook 15's reserved share rows gain the fields the bounds need: the item, creation and
     expiry times, the listener count and threshold, the paused and resumed state, and the
     listener keys.

   The configuration homes spec's decline of sharing outside the roster gains the bounded-share
   exception. Its "shares re-checked at fetch time" entry stands and matches bound 3. Its
   reserved share rows and the rung 6 spec's reserved rows stay.
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
   `ROADMAP.md` schedule it with the TUI. Ruling 2 confirms that scope, so none of the three
   changes. The configuration homes spec gains one line: the meter counts checkout.
8. **Navidrome sharing.** The build spec's trust-boundary requirements name a `/share/*`
   bypass. The erratum records it as live until 0.8, a residual the owner accepted, and
   records that any later sharing check reads the effective value (row A13).
9. **Refusals in the build spec.** The build spec's "Out of scope" section gains the refused
   features: fetching music, fetch integrations, federation, and a public library.
10. **CLAUDE.md.** The project instructions gain a short "Refused features" block, because
    implementer dispatches and plan reviewers read that file and never read this spec.
11. **`IDEAS.md`.** The Requests entry absorbs the wishlist, with its purchase claim and its
    web-only constraint. A local record store link joins as its own idea.

No spec proposes federation, self-signup, public collections, or any way to fetch music. The
organizer spec's refusal of scrapers rests on API fragility and concerns metadata sources such
as lyrics. This spec's refusal of music acquisition rests on the product's purpose.

## Rulings for Geoff

Rulings 4 to 12 are ruled as recommended (Geoff, 2026-09-25). Each recommendation below is the
spec's decision; the alternatives stay as a record of what was weighed.

Rulings 1 to 3 are ruled (Geoff, 2026-09-25), and the body follows them. Rulings 4 to 12 are
open. Rulings 8 to 12 come from the bounded-shares research. The numbers stay fixed, because
the body and the fold record cite them.

### Ruled

1. **A member's own devices holding the whole library is intended use.** Ruled yes. The meter
   counts it and the admin sees it, and nothing refuses it. The admin notice words a
   whole-library sync as expected, and the website copy says members' devices can keep the
   whole library for offline listening.
2. **A member may check out any collection they can see.** Ruled as a consequence of ruling
   1, since a checkout is the laptop's version of a phone's offline cache. The meter counts
   it. The configuration homes spec already says this.
3. **Navidrome sharing and downloads before 0.8 stay at Navidrome's defaults.** Ruled by
   Geoff's answer on sharing: "I'm not worried about Navidrome now. We'll fix it when we switch
   it over to dubplate." Both stay on as residuals the owner accepted, and "Before 0.8" states
   them. No pin or gate leg is owed. At 0.8 the streamer bounds shares (row A4), and any
   sharing check reads the effective value (row A13).

### Open

4. **The defaults.** Recommendation: a member cap of 20, a credential cap of 10 per member,
   and a volume notice when a member's distinct tracks in a rolling 24 hours pass half the
   library's tracks or 1,000 tracks, whichever is larger. Twenty covers an extended family or
   a wide circle of friends and stays below any number that reads as an audience. Ten covers a
   member's phone, tablet, laptops, and a car or speaker. For the notice, a member playing
   music nonstop for 24 hours at about four minutes a track hears about 360 tracks, so 1,000
   distinct tracks in a day means devices fetching the library, not a person listening. The
   floor keeps a new server with a few hundred tracks from sending a notice on an ordinary
   day. On a library under 2,000 tracks, a whole-library sync can pass unnoticed by email, and
   the volume view still shows it. Each default is a setting the admin can change, per member
   where it applies.
5. **What the volume view names.** Recommendation: name members and credentials on bytes and
   distinct tracks served, show stream counts only as group totals, and show no time-of-day
   breakdown. Only bytes and distinct tracks answer the view's one question, which is unusual
   copying. Per-person stream counts over time read as a listening log inside a family. Each
   member's page says: "In the app, the admin sees how much data your devices download, not
   what you play. The server keeps a history of what you play for your recently played and
   most played lists, and whoever runs the server can read it."
   - The recommendation builds per-member and per-credential byte and track columns, plus
     group-total stream counts.
   - Exceptions only: group totals by default, naming a member only when the notice fires.
   - Everything named: stream counts, bytes, and tracks per member and credential. The
     member-page sentence then drops "not what you play".
6. **Is the source question required?** Recommendation: required, with "Other" accepting an
   optional note. A skippable question lets the norm fade exactly when it matters, and the
   per-album override keeps a mixed batch honest.
   - Required: intake refuses a submission without an answer (row A9).
   - Optional: intake stores "unrecorded" for a missing answer.
7. **When buy links ship.** The web app has album pages from rung 7, before launch.
   Recommendation: rung 7, if a rung 7 plan task can absorb it without growing the pass.
   - Rung 7: the album page renders buy links. The uploader captures the optional store link
     from the source question, and that link renders first, ahead of the MusicBrainz source.
     Rung 9 captures the MusicBrainz links at admission as a new forward hook.
   - A later release: that release renders the links and captures them for albums already
     admitted. The uploader still captures the store link at rung 7, because upload is the
     only moment it exists.
8. **Share expiry default and maximum.** The evidence spans 24 hours (the Jellyfin plugin) to
   one year (Navidrome). Recommendation: a 7-day default and a 30-day maximum. Clicks on a
   messenger link cluster in its first hours, and a member can extend a link or send a new one.
   - 7 and 30: the clamp and the is/isn't copy say "a week" and "a month".
   - 14 and 30: the clamp and the copy say "two weeks" and "a month". This suits a group with
     slow openers.
9. **Listener-pause threshold.** Recommendation: 50 distinct listeners over the share's life,
   with one member resume that doubles the threshold. The operator carries the hosting and
   complaint risk, so a pause matters more for dubplate than for Bandcamp.
   - 50 over the life: the store keeps a lifetime listener count per share.
   - Bandcamp's 100 in 48 hours: the store keeps timestamped listener keys and a rolling
     window.
   - An alert with no pause: the notices fire at the threshold, and nothing stops the stream.
10. **Whether the share page names the member.** Recommendation: show the member's display
    name, as "Shared by Maya", since the friend already knows the sender. The cost is that an
    escaped link shows that name to strangers.
    - Named: the share page renders the display name, and the share response carries it.
    - Unnamed: the page says "Shared from a dubplate library", and no share response carries
      a member's name.
11. **Playlists.** A playlist can hold any selection, which is why `download` refuses playlist
    zips. Recommendation: allow a playlist share of up to 20 tracks, because a mixtape for a
    friend is casual sharing at its most typical.
    - Up to 20 tracks: the share snapshots the playlist's tracks at creation, so a later edit
      doesn't grow it. A longer playlist answers error 50 with a message that names the limit.
    - Left out at 0.8: a playlist id answers error 50, and the web app offers no playlist
      share.
12. **Share stream bitrate.** Recommendation: 128 kbps Opus VBR. The research didn't check
    whether Safari plays Opus in the share page's audio element, so the slice that builds the
    page confirms the fallback codec.
    - 128 kbps: near transparent for stereo music.
    - 96 kbps: within Xiph's streaming range, and saves bandwidth.
    - 192 kbps: adds nothing a listener hears, and adds value to a rip.
